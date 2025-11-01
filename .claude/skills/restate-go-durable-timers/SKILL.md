---
name: restate-go-durable-timers
description: Guide for Restate Go durable timers including sleep, delayed execution, and timeouts. Use when implementing delays, scheduling future work, or creating operation timeouts in handlers.
---

# Restate Go Durable Timers

Fault-tolerant timing features for pausing handlers, scheduling future invocations, and setting operation timeouts.

## Core Capabilities

Restate provides durable timing that:
- Tracks and manages timers across failures and restarts
- Survives crashes and service redeployments
- Persists in the execution log for replay safety
- Supports long-duration delays (months/years)

## Durable Sleep

Pause handler execution for a specified duration.

### Basic Usage

```go
import "time"

func (s *Service) ProcessWithDelay(ctx restate.Context, data Data) error {
    // Do some work
    processData(data)

    // Sleep for 10 seconds
    if err := restate.Sleep(ctx, 10*time.Second); err != nil {
        return err
    }

    // Continue after sleep
    finalizeProcessing(data)
    return nil
}
```

### Duration Examples

```go
// Short delays
restate.Sleep(ctx, 5*time.Second)
restate.Sleep(ctx, 30*time.Minute)

// Long delays
restate.Sleep(ctx, 24*time.Hour)       // 1 day
restate.Sleep(ctx, 7*24*time.Hour)     // 1 week
restate.Sleep(ctx, 30*24*time.Hour)    // 30 days
restate.Sleep(ctx, 365*24*time.Hour)   // 1 year
```

### Impact on Virtual Objects

**Important**: Sleep blocks the Virtual Object, preventing other calls from processing.

```go
type RateLimiter struct{}

func (r *RateLimiter) WaitAndProcess(ctx restate.ObjectContext, request Request) error {
    // This blocks ALL calls to this object key until sleep completes
    if err := restate.Sleep(ctx, 1*time.Minute); err != nil {
        return err
    }

    return processRequest(request)
}
```

**Recommendation**: Use delayed messages instead of sleep for Virtual Objects.

### Version Compatibility

Long sleeps require maintaining deployment versions until completion. The same handler version must be available when the sleep completes.

## Delayed Messages (Recommended)

Instead of sleep + send, use delayed messages for scheduling future work.

### Why Delayed Messages?

**Advantages**:
- Calling handler finishes immediately
- No blocking of Virtual Objects
- Simpler service versioning
- Better resource utilization

```go
// BAD: Sleep + Send
func (s *Service) ScheduleReminder(ctx restate.Context, user User) error {
    restate.Sleep(ctx, 7*24*time.Hour) // Blocks for 7 days!
    restate.ServiceSend(ctx, "NotificationService", "Send").Send(notification)
    return nil
}

// GOOD: Delayed message
func (s *Service) ScheduleReminder(ctx restate.Context, user User) error {
    // Handler completes immediately
    restate.ServiceSend(ctx, "NotificationService", "Send").
        Send(notification, restate.WithDelay(7*24*time.Hour))
    return nil
}
```

### Delayed Message Examples

```go
// Schedule notification for 24 hours from now
restate.ServiceSend(ctx, "EmailService", "SendReminder").
    Send(emailData, restate.WithDelay(24*time.Hour))

// Schedule Virtual Object call
restate.ObjectSend(ctx, "Subscription", userID, "ProcessRenewal").
    Send(renewalData, restate.WithDelay(30*24*time.Hour))

// Schedule workflow step
restate.WorkflowSend(ctx, "OrderWorkflow", orderID, "CancelIfNotPaid").
    Send(cancellationData, restate.WithDelay(15*time.Minute))
```

## Timeouts with Combinators

Race async operations against timeouts using Restate's combinators.

### Basic Timeout Pattern

```go
import "github.com/restatedev/sdk-go"

func (s *Service) CallWithTimeout(ctx restate.Context, request Request) (Response, error) {
    // Create timeout future
    sleepFuture := restate.After(ctx, 30*time.Second)

    // Create operation future
    callFuture := restate.Service[Response](ctx, "SlowService", "Process").
        RequestFuture(request)

    // Race them
    selected, err := restate.WaitFirst(ctx, sleepFuture, callFuture)
    if err != nil {
        return Response{}, err
    }

    switch selected.Index() {
    case 0:
        // Timeout won
        return Response{}, fmt.Errorf("operation timed out after 30 seconds")
    case 1:
        // Operation completed
        result, _ := callFuture.Done()
        return result, nil
    }

    return Response{}, nil
}
```

### Multiple Operations with Timeout

```go
func (s *Service) FetchWithTimeout(ctx restate.Context) ([]Data, error) {
    timeout := restate.After(ctx, 1*time.Minute)

    // Start multiple operations
    future1 := restate.Service[Data](ctx, "DataService", "Fetch1").RequestFuture("source1")
    future2 := restate.Service[Data](ctx, "DataService", "Fetch2").RequestFuture("source2")
    future3 := restate.Service[Data](ctx, "DataService", "Fetch3").RequestFuture("source3")

    // Wait for all with timeout
    allFutures := restate.After(ctx, 0) // Dummy future to start
    selected, err := restate.WaitFirst(ctx, timeout, future1, future2, future3)
    if err != nil {
        return nil, err
    }

    if selected.Index() == 0 {
        return nil, fmt.Errorf("timeout waiting for data")
    }

    // Collect results
    results := []Data{}
    if d, err := future1.Done(); err == nil {
        results = append(results, d)
    }
    if d, err := future2.Done(); err == nil {
        results = append(results, d)
    }
    if d, err := future3.Done(); err == nil {
        results = append(results, d)
    }

    return results, nil
}
```

## Common Patterns

### Retry with Exponential Backoff

```go
func (s *Service) RetryWithBackoff(ctx restate.Context, operation string) error {
    backoff := 1 * time.Second
    maxBackoff := 1 * time.Minute
    maxRetries := 5

    for i := 0; i < maxRetries; i++ {
        result, err := restate.Run(ctx, func(ctx restate.RunContext) (bool, error) {
            return tryOperation(operation)
        })

        if err == nil && result {
            return nil // Success
        }

        if i < maxRetries-1 {
            ctx.Log().Info("Retrying after backoff", "attempt", i+1, "backoff", backoff)
            if err := restate.Sleep(ctx, backoff); err != nil {
                return err
            }

            // Exponential backoff
            backoff *= 2
            if backoff > maxBackoff {
                backoff = maxBackoff
            }
        }
    }

    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

### Scheduled Recurring Task

```go
type SchedulerService struct{}

func (s *SchedulerService) StartDailyTask(ctx restate.Context, taskID string) error {
    // Trigger first execution
    restate.ObjectSend(ctx, "TaskRunner", taskID, "Execute").Send(restate.Void{})

    // Schedule next execution for 24 hours
    restate.ServiceSend(ctx, "SchedulerService", "StartDailyTask").
        Send(taskID, restate.WithDelay(24*time.Hour))

    return nil
}
```

### Rate Limiting with Sleep

```go
type APIClient struct{}

func (a *APIClient) BatchProcess(ctx restate.Context, items []Item) error {
    for i, item := range items {
        // Process item
        err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            callExternalAPI(item)
            return restate.Void{}, nil
        })
        if err != nil {
            return err
        }

        // Rate limit: 1 request per second (except last item)
        if i < len(items)-1 {
            if err := restate.Sleep(ctx, 1*time.Second); err != nil {
                return err
            }
        }
    }

    return nil
}
```

### Delayed Workflow Cancellation

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    // Schedule auto-cancellation if not paid within 15 minutes
    workflowID := restate.Key(ctx)
    restate.WorkflowSend(ctx, "OrderWorkflow", workflowID, "AutoCancel").
        Send(restate.Void{}, restate.WithDelay(15*time.Minute))

    // Wait for payment
    promise := restate.Promise[Payment](ctx, "payment-received")
    payment, err := promise.Result()
    if err != nil {
        return Receipt{}, err
    }

    // Payment received, mark to prevent auto-cancel
    restate.Set(ctx, "paid", true)

    return processOrder(order, payment), nil
}

func (w *OrderWorkflow) AutoCancel(ctx restate.WorkflowSharedContext) error {
    paid, _ := restate.Get[bool](ctx, "paid")
    if !paid {
        // Cancel unpaid order
        restate.Set(ctx, "status", "cancelled")
    }
    return nil
}
```

### Polling with Sleep

```go
func (s *Service) PollUntilComplete(ctx restate.Context, jobID string) (Result, error) {
    maxAttempts := 60 // 5 minutes with 5-second intervals
    pollInterval := 5 * time.Second

    for i := 0; i < maxAttempts; i++ {
        // Check job status
        status, err := restate.Run(ctx, func(ctx restate.RunContext) (JobStatus, error) {
            return checkJobStatus(jobID)
        })
        if err != nil {
            return Result{}, err
        }

        if status.Complete {
            return status.Result, nil
        }

        // Wait before next poll (except on last attempt)
        if i < maxAttempts-1 {
            if err := restate.Sleep(ctx, pollInterval); err != nil {
                return Result{}, err
            }
        }
    }

    return Result{}, fmt.Errorf("job did not complete within timeout")
}
```

## Clock Synchronization

**Important**: SDK and Server must maintain synchronized system clocks.

Clock mismatches can cause timers to:
- Fire earlier than expected (server ahead)
- Fire later than expected (server behind)

**Recommendation**: Use NTP or similar time synchronization in production.

## Best Practices

1. **Prefer Delayed Messages**: Over sleep for scheduling future work
2. **Avoid Long Sleeps in Virtual Objects**: They block all calls to that key
3. **Use Timeouts**: Race operations against timeouts for bounded execution
4. **Synchronize Clocks**: Ensure server and SDK have accurate time
5. **Plan for Long Delays**: Maintain service versions for long-running sleeps
6. **Log Sleep Events**: Add logging before sleeps for observability
7. **Handle Sleep Errors**: Always check error returns from Sleep
8. **Consider Alternatives**: Use combinators for complex timing logic

## Comparison: Sleep vs Delayed Messages

| Feature | Sleep | Delayed Messages |
|---------|-------|------------------|
| **Handler State** | Blocked, active | Completes immediately |
| **Virtual Objects** | Blocks other calls | Doesn't block |
| **Use Case** | Rate limiting, polling | Scheduling future work |
| **Versioning** | Must maintain version | More flexible |
| **Resource Usage** | Holds resources | Frees resources |

## References

- Official Docs: https://docs.restate.dev/develop/go/durable-timers
- Service Communication: See restate-go-service-communication skill
- Concurrent Tasks: See restate-go-concurrent-tasks skill
- External Events: See restate-go-external-events skill
