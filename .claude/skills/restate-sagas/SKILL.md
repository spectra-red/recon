---
name: restate-sagas
description: Guide for implementing saga patterns with Restate including compensating transactions, two-phase operations, and distributed rollback. Use when implementing distributed transactions, handling multi-service operations, or needing rollback capabilities.
---

# Restate Saga Pattern

Implement reliable distributed transactions across multiple services with automatic compensation.

## What is a Saga?

A saga is **a design pattern for handling transactions that span multiple services**. It breaks complex processes into:
- **Local operations**: Individual steps that modify one service
- **Compensating actions**: Undo operations for each step

This ensures system consistency when failures occur in distributed systems.

## Why Use Sagas?

### Traditional Distributed Transactions Don't Scale

Two-phase commit (2PC) across services creates:
- Long-held locks
- Tight coupling
- Availability issues
- Performance bottlenecks

### Sagas Provide

✅ **No distributed locks**: Each service manages its own transactions
✅ **High availability**: Services remain independent
✅ **Eventual consistency**: System converges to consistent state
✅ **Automatic recovery**: Restate ensures completion with retries

## Restate Saga Advantages

**Durable Execution**: Restate guarantees saga completion and automatically retries from failure points

**Code-First**: Write sagas in regular code without domain-specific languages

**Automatic Journaling**: No manual state tracking needed

## Saga Pattern Structure

### Three Phases

1. **Try Block**: Execute business logic steps sequentially
2. **Compensation Registration**: Record undo operations before each step
3. **Catch Block**: Reverse compensations on terminal failures

### Execution Flow

```
Step 1 → Register Compensation 1
Step 2 → Register Compensation 2
Step 3 → Register Compensation 3
  ↓
Error?
  ↓
Execute Compensation 3 (LIFO order)
Execute Compensation 2
Execute Compensation 1
```

## Go Implementation Pattern

### Basic Saga Structure

```go
func (s *OrderService) ProcessOrder(ctx restate.Context, order Order) (Receipt, error) {
    var compensations []func() (restate.Void, error)

    // Deferred compensation execution
    defer func() {
        if err != nil {
            // Execute compensations in reverse order
            for i := len(compensations) - 1; i >= 0; i-- {
                compensation := compensations[i]
                if _, compErr := compensation(); compErr != nil {
                    ctx.Log().Error("Compensation failed", "error", compErr)
                }
            }
        }
    }()

    // Step 1: Reserve inventory
    reservationID, err := restate.Service[string](ctx, "Inventory", "Reserve").
        Request(order.Items)
    if err != nil {
        return Receipt{}, err
    }

    // Register compensation 1
    compensations = append(compensations, func() (restate.Void, error) {
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            inventoryAPI.Release(reservationID)
            return restate.Void{}, nil
        })
    })

    // Step 2: Charge payment
    paymentID, err := restate.Service[string](ctx, "Payment", "Charge").
        Request(order.Payment)
    if err != nil {
        return Receipt{}, err  // Triggers defer -> compensation 1
    }

    // Register compensation 2
    compensations = append(compensations, func() (restate.Void, error) {
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            paymentAPI.Refund(paymentID)
            return restate.Void{}, nil
        })
    })

    // Step 3: Ship order
    shipmentID, err := restate.Service[string](ctx, "Shipping", "Ship").
        Request(order.ShippingInfo)
    if err != nil {
        return Receipt{}, err  // Triggers defer -> compensations 2, 1
    }

    // Success - no compensations executed
    return Receipt{
        OrderID:    order.ID,
        PaymentID:  paymentID,
        ShipmentID: shipmentID,
    }, nil
}
```

## Advanced Patterns

### Two-Phase API Pattern

For services with reserve/confirm semantics:

```go
func (s *TravelBooking) BookTrip(ctx restate.Context, trip TripRequest) (Booking, error) {
    var compensations []func() (restate.Void, error)

    defer func() {
        if err != nil {
            for i := len(compensations) - 1; i >= 0; i-- {
                compensations[i]()
            }
        }
    }()

    // Phase 1: Reserve flight (get reservation ID)
    flightReservation, err := restate.Service[Reservation](
        ctx,
        "FlightService",
        "Reserve",
    ).Request(trip.Flight)
    if err != nil {
        return Booking{}, restate.TerminalError(err, 409)
    }

    // Register cancellation
    compensations = append(compensations, func() (restate.Void, error) {
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            flightAPI.Cancel(flightReservation.ID)
            return restate.Void{}, nil
        })
    })

    // Phase 1: Reserve hotel
    hotelReservation, err := restate.Service[Reservation](
        ctx,
        "HotelService",
        "Reserve",
    ).Request(trip.Hotel)
    if err != nil {
        return Booking{}, restate.TerminalError(err, 409)
    }

    compensations = append(compensations, func() (restate.Void, error) {
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            hotelAPI.Cancel(hotelReservation.ID)
            return restate.Void{}, nil
        })
    })

    // Phase 2: Confirm all reservations
    _, err = restate.Service[void](ctx, "FlightService", "Confirm").
        Request(flightReservation.ID)
    if err != nil {
        return Booking{}, err
    }

    _, err = restate.Service[void](ctx, "HotelService", "Confirm").
        Request(hotelReservation.ID)
    if err != nil {
        return Booking{}, err
    }

    return Booking{
        FlightID: flightReservation.ID,
        HotelID:  hotelReservation.ID,
    }, nil
}
```

### One-Shot API with Idempotency

For APIs without two-phase support:

```go
func (s *PaymentSaga) ProcessSubscription(
    ctx restate.Context,
    subscription SubscriptionRequest,
) (Result, error) {
    var compensations []func() (restate.Void, error)

    defer func() {
        if err != nil {
            for i := len(compensations) - 1; i >= 0; i-- {
                compensations[i]()
            }
        }
    }()

    // Generate deterministic idempotency key
    paymentKey := restate.UUID(ctx).String()

    // Register compensation BEFORE the operation
    compensations = append(compensations, func() (restate.Void, error) {
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            // Use same idempotency key to cancel
            paymentAPI.Cancel(paymentKey)
            return restate.Void{}, nil
        })
    })

    // Execute payment with idempotency key
    paymentID, err := restate.Service[string](ctx, "Payment", "Charge").
        Request(
            subscription.Payment,
            restate.WithIdempotencyKey(paymentKey),
        )
    if err != nil {
        return Result{}, err
    }

    // Continue with more steps...
    return Result{PaymentID: paymentID}, nil
}
```

## When to Use Sagas

### Use Sagas For

✅ **Non-transient failures**: Business logic errors where retrying won't help
- "Hotel is fully booked"
- "Insufficient inventory"
- "Credit limit exceeded"

✅ **User cancellations**: Explicit cancellation requiring rollback

✅ **Multi-service operations**: Operations spanning multiple services that need atomicity

### Don't Use Sagas For

❌ **Transient failures**: Network issues, temporary unavailability
- Restate handles these automatically with retries

❌ **Single-service operations**: Use database transactions instead

❌ **Read-only operations**: No state changes = no compensation needed

## Error Handling in Sagas

### Terminal vs Transient Errors

```go
// Transient error - will retry automatically
if err != nil {
    return Result{}, err
}

// Terminal error - triggers compensations immediately
if businessValidationFailed {
    return Result{}, restate.TerminalError(
        fmt.Errorf("validation failed: %s", reason),
        400,
    )
}
```

### Compensation Error Handling

```go
defer func() {
    if err != nil {
        for i := len(compensations) - 1; i >= 0; i-- {
            if _, compErr := compensations[i](); compErr != nil {
                // Log but continue with remaining compensations
                ctx.Log().Error("Compensation failed",
                    "step", i,
                    "error", compErr)

                // Optional: Track failed compensations for manual intervention
                restate.ServiceSend(ctx, "AlertService", "CompensationFailed").
                    Send(CompensationAlert{
                        SagaID: ctx.Request().ID,
                        Step:   i,
                        Error:  compErr.Error(),
                    })
            }
        }
    }
}()
```

## Complete Example: E-Commerce Order

```go
package main

import (
    "fmt"
    "github.com/restatedev/sdk-go"
)

type OrderSaga struct{}

type Order struct {
    ID              string
    CustomerID      string
    Items           []Item
    PaymentMethod   PaymentMethod
    ShippingAddress Address
    TotalAmount     float64
}

type Receipt struct {
    OrderID        string
    PaymentID      string
    ReservationID  string
    ShipmentID     string
    Status         string
}

func (o *OrderSaga) PlaceOrder(ctx restate.Context, order Order) (Receipt, error) {
    var compensations []func() (restate.Void, error)
    var err error

    defer func() {
        if err != nil {
            ctx.Log().Info("Order failed, executing compensations",
                "order_id", order.ID,
                "compensation_count", len(compensations))

            for i := len(compensations) - 1; i >= 0; i-- {
                if _, compErr := compensations[i](); compErr != nil {
                    ctx.Log().Error("Compensation failed",
                        "order_id", order.ID,
                        "step", i,
                        "error", compErr)
                }
            }
        }
    }()

    // Step 1: Validate order
    valid, err := restate.Service[bool](ctx, "ValidationService", "ValidateOrder").
        Request(order)
    if err != nil {
        return Receipt{}, err
    }
    if !valid {
        return Receipt{}, restate.TerminalError(
            fmt.Errorf("order validation failed"),
            400,
        )
    }

    // Step 2: Check and reserve inventory
    reservationID, err := restate.Service[string](
        ctx,
        "InventoryService",
        "ReserveItems",
    ).Request(InventoryRequest{
        Items:      order.Items,
        CustomerID: order.CustomerID,
    })
    if err != nil {
        return Receipt{}, err
    }

    compensations = append(compensations, func() (restate.Void, error) {
        ctx.Log().Info("Releasing inventory", "reservation_id", reservationID)
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            inventoryAPI.ReleaseReservation(reservationID)
            return restate.Void{}, nil
        })
    })

    // Step 3: Authorize payment
    paymentKey := fmt.Sprintf("order-payment-%s", order.ID)
    paymentID, err := restate.Service[string](ctx, "PaymentService", "Authorize").
        Request(
            PaymentRequest{
                Amount:        order.TotalAmount,
                PaymentMethod: order.PaymentMethod,
                CustomerID:    order.CustomerID,
            },
            restate.WithIdempotencyKey(paymentKey),
        )
    if err != nil {
        return Receipt{}, err
    }

    compensations = append(compensations, func() (restate.Void, error) {
        ctx.Log().Info("Voiding payment", "payment_id", paymentID)
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            paymentAPI.Void(paymentID)
            return restate.Void{}, nil
        })
    })

    // Step 4: Capture payment
    captureID, err := restate.Service[string](ctx, "PaymentService", "Capture").
        Request(paymentID)
    if err != nil {
        return Receipt{}, err
    }

    compensations = append(compensations, func() (restate.Void, error) {
        ctx.Log().Info("Refunding payment", "capture_id", captureID)
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            paymentAPI.Refund(captureID)
            return restate.Void{}, nil
        })
    })

    // Step 5: Create shipment
    shipmentID, err := restate.Service[string](ctx, "ShippingService", "CreateShipment").
        Request(ShipmentRequest{
            OrderID:         order.ID,
            ReservationID:   reservationID,
            Address:         order.ShippingAddress,
            CustomerID:      order.CustomerID,
        })
    if err != nil {
        return Receipt{}, err
    }

    compensations = append(compensations, func() (restate.Void, error) {
        ctx.Log().Info("Canceling shipment", "shipment_id", shipmentID)
        return restate.Void{}, restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            shippingAPI.CancelShipment(shipmentID)
            return restate.Void{}, nil
        })
    })

    // Step 6: Send confirmation email (optional, no compensation needed)
    restate.ServiceSend(ctx, "EmailService", "SendOrderConfirmation").
        Send(EmailRequest{
            OrderID:    order.ID,
            CustomerID: order.CustomerID,
            Email:      order.CustomerEmail,
        })

    // Success!
    return Receipt{
        OrderID:       order.ID,
        PaymentID:     captureID,
        ReservationID: reservationID,
        ShipmentID:    shipmentID,
        Status:        "completed",
    }, nil
}
```

## Best Practices

### 1. Register Compensations Before Actions

```go
// GOOD: Register compensation first
compensations = append(compensations, undoFunc)
result, err := performAction()

// BAD: Register after (might miss if action succeeds but fails to register)
result, err := performAction()
compensations = append(compensations, undoFunc)  // ❌ May not execute
```

### 2. Make Compensations Idempotent

```go
func (c *Compensation) ReleaseInventory(reservationID string) error {
    // Idempotent: Safe to call multiple times
    return inventoryAPI.ReleaseReservation(reservationID)
    // API should handle: "reservation already released"
}
```

### 3. Log Compensation Execution

```go
compensations = append(compensations, func() (restate.Void, error) {
    ctx.Log().Info("Executing compensation: refund payment",
        "payment_id", paymentID,
        "amount", amount)

    err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        return restate.Void{}, paymentAPI.Refund(paymentID)
    })

    if err != nil {
        ctx.Log().Error("Compensation failed", "error", err)
    }
    return restate.Void{}, err
})
```

### 4. Use Deterministic IDs

```go
// GOOD: Deterministic UUID
idempotencyKey := restate.UUID(ctx).String()

// BAD: Non-deterministic UUID
idempotencyKey := uuid.New().String()  // ❌ Different on replay
```

### 5. Handle Partial Compensation Failures

```go
defer func() {
    if err != nil {
        failedCompensations := []int{}

        for i := len(compensations) - 1; i >= 0; i-- {
            if _, compErr := compensations[i](); compErr != nil {
                failedCompensations = append(failedCompensations, i)
            }
        }

        if len(failedCompensations) > 0 {
            // Alert operations team
            restate.ServiceSend(ctx, "AlertService", "ManualIntervention").
                Send(Alert{
                    Type:     "saga-compensation-failed",
                    SagaID:   ctx.Request().ID,
                    Steps:    failedCompensations,
                })
        }
    }
}()
```

## Testing Sagas

### Unit Tests

```go
func TestOrderSaga_CompensationOnPaymentFailure(t *testing.T) {
    // Simulate payment failure after inventory reservation
    // Verify inventory compensation is called
}

func TestOrderSaga_AllCompensationsExecute(t *testing.T) {
    // Simulate failure at last step
    // Verify all compensations execute in reverse order
}
```

### Integration Tests

```go
func TestOrderSaga_E2E(t *testing.T) {
    // Run full saga
    // Inject failures at different steps
    // Verify system returns to consistent state
}
```

## References

- Official Docs: https://docs.restate.dev/guides/sagas
- Error Handling: See restate-go-error-handling skill
- Service Communication: See restate-go-service-communication skill
- Workflows: See restate-go-services skill
