---
name: restate-go-external-events
description: Guide for Restate Go external events using awakeables and durable promises for human-in-the-loop workflows, manual approvals, and external system integration. Use when implementing waiting for external signals, webhooks, or async event handling.
---

# Restate Go External Events

Enable handlers to pause and wait for external processes through durable waiting primitives.

## Core Purpose

Restate provides mechanisms for handlers to suspend execution and resume when external events occur. This supports:

- Human-in-the-loop workflows (approvals, reviews, manual steps)
- External API responses
- Webhook callbacks
- Manual interventions
- Third-party system integration

## Two Primary Primitives

### Awakeables

**Best for**: Services and Virtual Objects
**Mechanism**: Unique ID-based completion
**Use case**: External systems that need to signal completion

### Durable Promises

**Best for**: Workflows only
**Mechanism**: Logical naming instead of ID management
**Use case**: Workflow internal coordination and external signaling
**Retention**: Up to 24 hours post-workflow completion

## Awakeables

### Implementation Pattern

Awakeables follow a three-step process:

```go
// 1. Create awakeable and get unique ID
awakeable := restate.Awakeable[ApprovalResult](ctx)
awakeableID := awakeable.Id()

// 2. Send ID to external system
err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
    sendToExternalSystem(awakeableID, requestData)
    return restate.Void{}, nil
})

// 3. Wait for external system to respond
result, err := awakeable.Result()
if err != nil {
    return err
}
```

### Complete Example: Approval Workflow

```go
type ApprovalService struct{}

type ApprovalRequest struct {
    Document string
    Approver string
}

type ApprovalResult struct {
    Approved bool
    Comments string
}

func (a *ApprovalService) RequestApproval(
    ctx restate.Context,
    request ApprovalRequest,
) (ApprovalResult, error) {
    // Create awakeable
    awakeable := restate.Awakeable[ApprovalResult](ctx)

    // Send notification to approver with awakeable ID
    _, err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        emailService.SendApprovalRequest(
            request.Approver,
            request.Document,
            awakeable.Id(),
        )
        return restate.Void{}, nil
    })
    if err != nil {
        return ApprovalResult{}, err
    }

    // Wait for approval (handler suspends here)
    result, err := awakeable.Result()
    if err != nil {
        return ApprovalResult{}, err
    }

    return result, nil
}
```

### Completing Awakeables

External systems can complete awakeables using:

#### Go SDK

```go
import "github.com/restatedev/sdk-go/ingress"

client := restateingress.NewClient("http://localhost:8080")

// Resolve with success
err := restateingress.ResolveAwakeable(
    client,
    awakeableID,
    ApprovalResult{Approved: true, Comments: "LGTM"},
)

// Reject with error
err := restateingress.RejectAwakeable(
    client,
    awakeableID,
    "Approval denied",
)
```

#### HTTP API

```bash
# Resolve
curl -X POST http://localhost:8080/restate/awakeable/{awakeableID}/resolve \
  -H "Content-Type: application/json" \
  -d '{"approved": true, "comments": "LGTM"}'

# Reject
curl -X POST http://localhost:8080/restate/awakeable/{awakeableID}/reject \
  -H "Content-Type: text/plain" \
  -d "Approval denied"
```

## Durable Promises

Workflows-only feature for bidirectional communication using logical names.

### External to Workflow

External handlers signal the Run handler:

```go
type PaymentWorkflow struct{}

func (w *PaymentWorkflow) Run(ctx restate.WorkflowContext, payment Payment) (Receipt, error) {
    // Send payment request to external system
    _, err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        externalPaymentSystem.Process(payment, ctx.WorkflowID())
        return restate.Void{}, nil
    })

    // Wait for external payment confirmation via promise
    promise := restate.Promise[PaymentResult](ctx, "payment-result")
    result, err := promise.Result()
    if err != nil {
        return Receipt{}, err
    }

    return createReceipt(result), nil
}

// External webhook endpoint completes the promise
func (w *PaymentWorkflow) WebhookCallback(
    ctx restate.WorkflowSharedContext,
    result PaymentResult,
) error {
    // Resolve the promise, waking up the Run handler
    promise := restate.Promise[PaymentResult](ctx, "payment-result")
    promise.Resolve(result)
    return nil
}
```

### Workflow to External

The Run handler broadcasts to external handlers:

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    // Process order steps...
    restate.Set(ctx, "status", "processing")

    // Signal completion to waiting handlers
    promise := restate.Promise[string](ctx, "order-completed")
    promise.Resolve("Order completed successfully")

    return receipt, nil
}

// External handler waits for workflow completion
func (w *OrderWorkflow) WaitForCompletion(ctx restate.WorkflowSharedContext) (string, error) {
    promise := restate.Promise[string](ctx, "order-completed")
    message, err := promise.Result()
    return message, err
}
```

### Promise Operations

```go
// Create/get promise
promise := restate.Promise[ResultType](ctx, "promise-name")

// Wait for result (blocks until resolved/rejected)
result, err := promise.Result()

// Resolve promise (success)
promise.Resolve(successData)

// Reject promise (error)
promise.Reject("Error message")

// Peek at result without waiting
result, err := promise.Peek()
```

## Comparison: Awakeables vs Promises

| Feature | Awakeables | Durable Promises |
|---------|-----------|------------------|
| **Scope** | Services, Virtual Objects, Workflows | Workflows only |
| **Identifier** | Auto-generated UUID | Logical string name |
| **Lifetime** | Until resolved/rejected | Workflow retention period |
| **Use Case** | External system callbacks | Workflow coordination |
| **Completion** | Via SDK or HTTP API | Via workflow handlers |

## Common Patterns

### Human Approval with Timeout

```go
func (s *Service) ApproveWithTimeout(
    ctx restate.Context,
    request Request,
) (ApprovalResult, error) {
    awakeable := restate.Awakeable[ApprovalResult](ctx)

    // Send approval request
    restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        notificationService.Send(request.Approver, awakeable.Id())
        return restate.Void{}, nil
    })

    // Race approval against timeout
    timeoutFuture := restate.After(ctx, 24*time.Hour)
    approvalFuture := awakeable.Future()

    selected, err := restate.WaitFirst(ctx, timeoutFuture, approvalFuture)
    if err != nil {
        return ApprovalResult{}, err
    }

    switch selected.Index() {
    case 0: // Timeout
        return ApprovalResult{Approved: false, Comments: "Timed out"}, nil
    case 1: // Approval received
        result, _ := approvalFuture.Done()
        return result, nil
    }

    return ApprovalResult{}, nil
}
```

### External API Callback

```go
type WebhookService struct{}

func (w *WebhookService) WaitForWebhook(
    ctx restate.Context,
    callbackURL string,
) (WebhookData, error) {
    awakeable := restate.Awakeable[WebhookData](ctx)

    // Register webhook with awakeable ID
    _, err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        externalAPI.RegisterWebhook(callbackURL + "?id=" + awakeable.Id())
        return restate.Void{}, nil
    })

    // Wait for webhook to be called
    data, err := awakeable.Result()
    return data, err
}

// Webhook handler (separate endpoint)
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
    awakeableID := r.URL.Query().Get("id")
    var data WebhookData
    json.NewDecoder(r.Body).Decode(&data)

    client := restateingress.NewClient("http://localhost:8080")
    restateingress.ResolveAwakeable(client, awakeableID, data)

    w.WriteHeader(http.StatusOK)
}
```

### Multi-Step Workflow with Promises

```go
type OnboardingWorkflow struct{}

func (w *OnboardingWorkflow) Run(
    ctx restate.WorkflowContext,
    user User,
) (OnboardingResult, error) {
    // Step 1: Email verification
    emailPromise := restate.Promise[bool](ctx, "email-verified")
    sendVerificationEmail(ctx, user.Email)

    emailVerified, err := emailPromise.Result()
    if err != nil || !emailVerified {
        return OnboardingResult{}, fmt.Errorf("email verification failed")
    }

    // Step 2: Document upload
    docPromise := restate.Promise[Document](ctx, "document-uploaded")
    notifyUserToUploadDocuments(ctx, user.ID)

    document, err := docPromise.Result()
    if err != nil {
        return OnboardingResult{}, err
    }

    // Step 3: Manual review
    reviewPromise := restate.Promise[ReviewResult](ctx, "review-completed")
    assignToReviewer(ctx, document)

    review, err := reviewPromise.Result()
    if err != nil {
        return OnboardingResult{}, err
    }

    return OnboardingResult{Approved: review.Approved}, nil
}

// Called when user verifies email
func (w *OnboardingWorkflow) EmailVerified(
    ctx restate.WorkflowSharedContext,
) error {
    promise := restate.Promise[bool](ctx, "email-verified")
    promise.Resolve(true)
    return nil
}

// Called when user uploads document
func (w *OnboardingWorkflow) DocumentUploaded(
    ctx restate.WorkflowSharedContext,
    doc Document,
) error {
    promise := restate.Promise[Document](ctx, "document-uploaded")
    promise.Resolve(doc)
    return nil
}

// Called when reviewer completes review
func (w *OnboardingWorkflow) ReviewCompleted(
    ctx restate.WorkflowSharedContext,
    result ReviewResult,
) error {
    promise := restate.Promise[ReviewResult](ctx, "review-completed")
    promise.Resolve(result)
    return nil
}
```

## Best Practices

1. **Use Descriptive Names**: For promises, use meaningful names that describe what's being awaited
2. **Handle Rejections**: Always check errors from `Result()` calls
3. **Set Timeouts**: Race awakeables/promises against timeouts for bounded waiting
4. **Store IDs Durably**: Use `restate.Run()` to persist awakeable IDs to external systems
5. **Idempotent Completion**: Design external systems to handle duplicate resolve/reject calls
6. **Choose Right Primitive**: Use promises for workflows, awakeables for everything else
7. **Document Contracts**: Clearly document what data external systems should send

## Architectural Benefits

Restate promises and awakeables are:
- **Durable**: Survive crashes and restarts
- **Distributed**: Can be resolved from any handler
- **Suspendable**: Handlers suspend on serverless, resume when signaled
- **Type-safe**: Full Go generics support

## References

- Official Docs: https://docs.restate.dev/develop/go/external-events
- Durable Timers: See restate-go-durable-timers skill
- Workflows: See restate-go-services skill
- Concurrent Tasks: See restate-go-concurrent-tasks skill
