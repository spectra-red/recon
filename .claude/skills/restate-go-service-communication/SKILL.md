---
name: restate-go-service-communication
description: Guide for Restate Go service-to-service communication including request-response calls, one-way messages, delayed invocations, and idempotency. Use when implementing handler interactions, async messaging, or workflow coordination.
---

# Restate Go Service Communication

Enable handlers to communicate through request-response calls, one-way messages, and delayed invocations.

## Three Communication Patterns

### Request-Response Calls
Synchronous calls that wait for results

### One-Way Messages
Asynchronous fire-and-forget calls

### Delayed Messages
Messages sent after a specified delay

## Request-Response Calls

Invoke a handler and await its response.

### Calling Services

```go
result, err := restate.Service[string](ctx, "GreeterService", "Greet").
    Request("Alice")
if err != nil {
    return "", err
}
```

### Calling Virtual Objects

```go
count, err := restate.Object[int](ctx, "Counter", "user-123", "Increment").
    Request(1)
if err != nil {
    return 0, err
}
```

### Calling Workflows

```go
receipt, err := restate.Workflow[Receipt](ctx, "OrderWorkflow", "order-456", "Run").
    Request(orderData)
if err != nil {
    return Receipt{}, err
}
```

### Deadlock Warning

**Request-response calls between exclusive handlers of Virtual Objects may lead to deadlocks** through:
- Cross-references between objects
- Circular dependencies
- Cycles in call chains

**Resolution**: Use cancellation or refactor to avoid cycles.

## One-Way Messages

Send messages without blocking on response.

### Service Messages

```go
err := restate.ServiceSend(ctx, "NotificationService", "SendEmail").
    Send(emailData)
if err != nil {
    return err
}
```

### Virtual Object Messages

```go
err := restate.ObjectSend(ctx, "Analytics", "user-123", "TrackEvent").
    Send(eventData)
if err != nil {
    return err
}
```

### Workflow Messages

```go
err := restate.WorkflowSend(ctx, "OrderWorkflow", "order-456", "Cancel").
    Send(reason)
if err != nil {
    return err
}
```

### Ordering Guarantee

**Key guarantee**: "Calls to a Virtual Object execute in order of arrival, serially"

This ensures:
- Messages to the same object key process in order
- No concurrent execution for the same key
- Predictable state updates

## Delayed Messages

Schedule messages for future execution.

### Syntax

Use `restate.WithDelay()` option with any `Send()` call:

```go
// Delay for 5 days
err := restate.ServiceSend(ctx, "ReminderService", "SendReminder").
    Send(reminder, restate.WithDelay(5*24*time.Hour))
```

### Use Cases

```go
// Delayed notification
err := restate.ObjectSend(ctx, "UserAccount", userID, "SendTrialEnding").
    Send(notification, restate.WithDelay(7*24*time.Hour))

// Scheduled workflow step
err := restate.WorkflowSend(ctx, "SubscriptionWorkflow", subID, "ProcessRenewal").
    Send(renewalData, restate.WithDelay(30*24*time.Hour))
```

### Advantages Over Sleep

Prefer delayed messages over `restate.Sleep()` + `Send()` because they:
- Allow the calling handler to complete immediately
- Prevent Virtual Object blocking during delays
- Simplify service versioning
- Avoid long-running invocations

## Idempotency Keys

Prevent duplicate executions across independent handler calls.

### Basic Usage

```go
result, err := restate.Service[string](ctx, "PaymentService", "Charge").
    Request(chargeData, restate.WithIdempotencyKey("charge-order-123"))
```

### Guarantees

- Same idempotency key returns cached response
- Responses persist for 24 hours
- Prevents duplicate side effects
- Works across all communication patterns

### Common Patterns

```go
// Use order ID as idempotency key
orderID := "order-" + uuid.String()
receipt, err := restate.Service[Receipt](ctx, "PaymentService", "ProcessPayment").
    Request(payment, restate.WithIdempotencyKey(orderID))

// Use deterministic UUID for idempotency
idempotencyKey := restate.UUID(ctx).String()
_, err := restate.ServiceSend(ctx, "EmailService", "Send").
    Send(email, restate.WithIdempotencyKey(idempotencyKey))
```

## Invocation Management

### Attach to Previous Invocations

Retrieve results from previously sent messages:

```go
// Using invocation ID
result, err := restate.AttachInvocation[string](ctx, invocationID).Response()

// Using idempotency key for services
result, err := restate.ServiceInvocationByIdempotencyKey[string](
    ctx,
    "PaymentService",
    "processPayment/idempotency-key-123",
).Response()
```

### Cancel Invocations

```go
err := restate.CancelInvocation(ctx, invocationID)
if err != nil {
    return err
}
```

## Complete Examples

### Request-Response with Error Handling

```go
func (o *OrderService) PlaceOrder(ctx restate.Context, order Order) (Receipt, error) {
    // Validate payment
    paymentOK, err := restate.Service[bool](ctx, "PaymentService", "Validate").
        Request(order.PaymentInfo)
    if err != nil {
        return Receipt{}, err
    }
    if !paymentOK {
        return Receipt{}, restate.TerminalError(fmt.Errorf("payment failed"), 400)
    }

    // Process payment with idempotency
    receipt, err := restate.Service[Receipt](ctx, "PaymentService", "Charge").
        Request(order.Amount, restate.WithIdempotencyKey("order-"+order.ID))
    if err != nil {
        return Receipt{}, err
    }

    return receipt, nil
}
```

### One-Way Message Pattern

```go
func (o *OrderService) CompleteOrder(ctx restate.Context, orderID string) error {
    // Update order status
    restate.Set(ctx, "status", "completed")

    // Send async notifications (fire and forget)
    restate.ServiceSend(ctx, "EmailService", "SendConfirmation").
        Send(emailData)

    restate.ObjectSend(ctx, "Analytics", orderID, "TrackCompletion").
        Send(analyticsData)

    // Schedule follow-up
    restate.ServiceSend(ctx, "FeedbackService", "RequestReview").
        Send(reviewRequest, restate.WithDelay(7*24*time.Hour))

    return nil
}
```

### Virtual Object Communication

```go
type ShoppingCart struct{}

func (c *ShoppingCart) Checkout(ctx restate.ObjectContext, payment PaymentInfo) (Receipt, error) {
    cartKey := restate.Key(ctx)
    items, _ := restate.Get[[]Item](ctx, "items")

    // Call payment service
    receipt, err := restate.Service[Receipt](ctx, "PaymentService", "Process").
        Request(payment, restate.WithIdempotencyKey("cart-"+cartKey))
    if err != nil {
        return Receipt{}, err
    }

    // Notify inventory service (async)
    for _, item := range items {
        restate.ObjectSend(ctx, "Inventory", item.ProductID, "Reserve").
            Send(item.Quantity)
    }

    // Clear cart
    restate.ClearAll(ctx)

    return receipt, nil
}
```

### Workflow Coordination

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (OrderResult, error) {
    // Call payment service
    paymentResult, err := restate.Service[PaymentResult](ctx, "PaymentService", "Charge").
        Request(order.Payment)
    if err != nil {
        return OrderResult{}, err
    }

    // Trigger fulfillment workflow
    fulfillmentResult, err := restate.Workflow[FulfillmentResult](
        ctx,
        "FulfillmentWorkflow",
        order.ID,
        "Run",
    ).Request(order.Items)
    if err != nil {
        // Compensate payment
        restate.Service[void](ctx, "PaymentService", "Refund").
            Request(paymentResult.TransactionID)
        return OrderResult{}, err
    }

    return OrderResult{
        Payment:     paymentResult,
        Fulfillment: fulfillmentResult,
    }, nil
}
```

## Best Practices

1. **Use Idempotency Keys**: For critical operations to prevent duplicates
2. **Prefer Delayed Messages**: Over sleep + send for scheduled operations
3. **Avoid Deadlocks**: Don't create circular dependencies between Virtual Objects
4. **Handle Errors**: Always check and handle call errors appropriately
5. **Order Guarantees**: Leverage Virtual Object ordering for state consistency
6. **Fire-and-Forget**: Use one-way messages for non-critical notifications
7. **Attach When Needed**: Reattach to long-running invocations to get results

## Type Safety

All communication methods are type-safe using Go generics:

```go
// Type parameter specifies expected return type
result, err := restate.Service[ReturnType](ctx, service, handler).Request(input)

// Compiler enforces type matching
var receipt Receipt
receipt, err = restate.Service[Receipt](ctx, "PaymentService", "Charge").Request(data)
```

## References

- Official Docs: https://docs.restate.dev/develop/go/service-communication
- Durable Steps: See restate-go-durable-steps skill
- Invocation Management: See restate-invocation-management skill
- Workflows: See restate-go-services skill
