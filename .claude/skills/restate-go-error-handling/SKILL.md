---
name: restate-go-error-handling
description: Guide for Restate Go error handling including retry behavior, terminal errors, and error patterns. Use when implementing error recovery, retry logic, or failure handling in handlers.
---

# Restate Go Error Handling

Manage errors, retries, and failure scenarios in Restate handlers.

## Automatic Retry Behavior

Restate automatically handles retries for failed invocations with:
- **Infinite retries by default**: Continues until success
- **Exponential backoff**: Progressively longer delays between retries
- **Durability**: Retry state persists across crashes
- **Deterministic replay**: Completed steps don't re-execute

### Default Behavior

```go
func (s *Service) ProcessData(ctx restate.Context, data Data) (Result, error) {
    // Any returned error triggers automatic retry
    result, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
        return externalAPI.Process(data)
    })

    if err != nil {
        // This error will cause the handler to retry
        return Result{}, err
    }

    return result, nil
}
```

### What Gets Retried

Errors returned from:
- Handler functions
- `restate.Run()` blocks
- `restate.Service().Request()` calls
- `restate.Object().Request()` calls
- `restate.Workflow().Request()` calls

## Terminal Errors

Stop retries immediately and return error to caller.

### Basic Terminal Error

```go
func (s *Service) Validate(ctx restate.Context, input Input) (Output, error) {
    if input.ID == "" {
        // Don't retry - this will never succeed
        return Output{}, restate.TerminalError(
            fmt.Errorf("invalid input: ID is required"),
            400, // HTTP status code
        )
    }

    return processInput(input), nil
}
```

### Terminal Error without Status Code

```go
func (s *Service) Process(ctx restate.Context, data Data) error {
    if !isValid(data) {
        // Status code is optional
        return restate.TerminalError(
            fmt.Errorf("data validation failed"),
        )
    }

    return nil
}
```

### Conditional Terminal Errors

```go
func (s *Service) CallExternal(ctx restate.Context, request Request) (Response, error) {
    response, err := restate.Run(ctx, func(ctx restate.RunContext) (Response, error) {
        return externalAPI.Call(request)
    })

    if err != nil {
        // Check if error is retriable
        if isClientError(err) {
            // Client errors (4xx) shouldn't retry
            return Response{}, restate.TerminalError(err, 400)
        }

        // Server errors (5xx) should retry
        return Response{}, err
    }

    return response, nil
}

func isClientError(err error) bool {
    // Check if error indicates client-side problem
    if httpErr, ok := err.(*HTTPError); ok {
        return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
    }
    return false
}
```

## Custom Retry Policies

Configure retry behavior for specific operations.

### Retry Policy Options

```go
result, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
    return riskyOperation()
},
    // Name for observability
    restate.WithName("risky-operation"),

    // Maximum total time spent retrying
    restate.WithMaxRetryDuration(5*time.Minute),

    // Initial delay before first retry
    restate.WithRetryInterval(1*time.Second),

    // Multiply delay by this factor after each retry
    restate.WithRetryIntervalMultiplier(2.0),
)

if err != nil {
    // When policy exhausted, a TerminalError is thrown
    return Result{}, err
}
```

### Example: Bounded Retries

```go
func (s *Service) BoundedRetry(ctx restate.Context) (Data, error) {
    data, err := restate.Run(ctx, func(ctx restate.RunContext) (Data, error) {
        return externalService.Fetch()
    },
        restate.WithName("fetch-external-data"),
        restate.WithMaxRetryDuration(3*time.Minute),
        restate.WithRetryInterval(500*time.Millisecond),
        restate.WithRetryIntervalMultiplier(2.0),
    )

    if err != nil {
        // After 3 minutes, this becomes a terminal error
        ctx.Log().Error("Failed to fetch data after retries", "error", err)
        return Data{}, err
    }

    return data, nil
}
```

## Error Patterns

### Validate Before Processing

```go
func (s *Service) CreateUser(ctx restate.Context, user User) (string, error) {
    // Validation errors are terminal
    if user.Email == "" {
        return "", restate.TerminalError(
            fmt.Errorf("email is required"),
            400,
        )
    }

    if !isValidEmail(user.Email) {
        return "", restate.TerminalError(
            fmt.Errorf("invalid email format"),
            400,
        )
    }

    // Processing errors are retriable
    userID, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
        return database.CreateUser(user)
    })

    if err != nil {
        // Database errors will retry
        return "", err
    }

    return userID, nil
}
```

### Graceful Degradation

```go
func (s *Service) EnrichedData(ctx restate.Context, id string) (Data, error) {
    // Fetch core data (must succeed)
    coreData, err := restate.Service[CoreData](ctx, "CoreService", "Get").
        Request(id)
    if err != nil {
        return Data{}, err // Will retry
    }

    // Try to enrich with optional data (can fail)
    enrichment, err := restate.Service[Enrichment](ctx, "EnrichmentService", "Get").
        Request(id)
    if err != nil {
        ctx.Log().Warn("Failed to fetch enrichment data", "error", err)
        // Continue without enrichment
        enrichment = Enrichment{}
    }

    return combine(coreData, enrichment), nil
}
```

### Compensating Transactions (Saga Pattern)

```go
func (s *Service) TransferFunds(
    ctx restate.Context,
    from, to string,
    amount int,
) error {
    // Step 1: Debit from account
    _, err := restate.Object[restate.Void](ctx, "Account", from, "Debit").
        Request(amount)
    if err != nil {
        return err // Will retry
    }

    // Step 2: Credit to account
    _, err = restate.Object[restate.Void](ctx, "Account", to, "Credit").
        Request(amount)
    if err != nil {
        // Compensate: refund the debit
        ctx.Log().Error("Credit failed, compensating", "error", err)

        _, compensateErr := restate.Object[restate.Void](ctx, "Account", from, "Credit").
            Request(amount)
        if compensateErr != nil {
            ctx.Log().Error("Compensation failed!", "error", compensateErr)
        }

        // Return terminal error - don't retry after compensation
        return restate.TerminalError(
            fmt.Errorf("transfer failed and was compensated: %w", err),
            500,
        )
    }

    return nil
}
```

### Idempotent Error Recovery

```go
func (s *Service) IdempotentProcess(ctx restate.Context, order Order) (Receipt, error) {
    // Use deterministic idempotency key
    idempotencyKey := restate.UUID(ctx).String()

    // Payment call is idempotent - safe to retry
    receipt, err := restate.Service[Receipt](ctx, "PaymentService", "Charge").
        Request(
            order.Payment,
            restate.WithIdempotencyKey(idempotencyKey),
        )

    if err != nil {
        // Safe to retry - idempotency prevents duplicate charges
        return Receipt{}, err
    }

    return receipt, nil
}
```

### Timeout and Fallback

```go
func (s *Service) FetchWithFallback(ctx restate.Context, id string) (Data, error) {
    // Try primary source with timeout
    timeout := restate.After(ctx, 5*time.Second)
    primaryFuture := restate.Service[Data](ctx, "PrimaryDB", "Get").
        RequestFuture(id)

    selected, err := restate.WaitFirst(ctx, timeout, primaryFuture)
    if err != nil {
        return Data{}, err
    }

    if selected.Index() == 0 {
        // Timeout - try fallback
        ctx.Log().Warn("Primary source timed out, trying fallback")

        fallback, err := restate.Service[Data](ctx, "CacheService", "Get").
            Request(id)
        if err != nil {
            // Fallback also failed - terminal error
            return Data{}, restate.TerminalError(
                fmt.Errorf("both primary and fallback failed"),
                503,
            )
        }

        return fallback, nil
    }

    // Primary succeeded
    return primaryFuture.Done()
}
```

### Error Aggregation

```go
func (s *Service) ProcessBatch(ctx restate.Context, items []Item) (Summary, error) {
    futures := make([]restate.Future[Result], len(items))
    for i, item := range items {
        futures[i] = restate.Service[Result](ctx, "Processor", "Process").
            RequestFuture(item)
    }

    var results []Result
    var errors []error

    for i, future := range futures {
        result, err := future.Done()
        if err != nil {
            errors = append(errors, fmt.Errorf("item %d failed: %w", i, err))
            continue
        }
        results = append(results, result)
    }

    if len(errors) > 0 {
        // Decide: fail if ANY failed, or if ALL failed?
        if len(errors) == len(items) {
            // All failed - terminal error
            return Summary{}, restate.TerminalError(
                fmt.Errorf("all items failed: %v", errors),
                500,
            )
        }

        // Partial success - log but continue
        ctx.Log().Warn("Some items failed",
            "failed_count", len(errors),
            "success_count", len(results))
    }

    return createSummary(results), nil
}
```

## State Consistency Considerations

When using terminal errors, consider undoing prior actions:

### Before Terminal Error

```go
func (s *Service) ComplexOperation(ctx restate.Context, input Input) error {
    // Step 1
    restate.Set(ctx, "step1_done", true)
    err := doStep1(ctx, input)
    if err != nil {
        return err // Will retry
    }

    // Step 2
    err = doStep2(ctx, input)
    if err != nil {
        if isUnrecoverable(err) {
            // Clear state before terminal error
            restate.Clear(ctx, "step1_done")

            return restate.TerminalError(err, 500)
        }
        return err // Will retry
    }

    return nil
}
```

## Catching Terminal Errors

```go
func (s *Service) HandleTerminalError(ctx restate.Context) error {
    err := someOperation(ctx)

    if terminalErr, ok := err.(*restate.TerminalErrorType); ok {
        // This is a terminal error
        ctx.Log().Error("Terminal error occurred",
            "message", terminalErr.Error(),
            "code", terminalErr.Code)

        // Custom handling for terminal errors
        notifyAdmin(terminalErr)

        return terminalErr
    }

    // Regular error - will retry
    return err
}
```

## Best Practices

1. **Validate Early**: Use terminal errors for validation failures
2. **Distinguish Error Types**: Client errors (4xx) terminal, server errors (5xx) retry
3. **Idempotency Keys**: Use for operations that shouldn't duplicate on retry
4. **Compensating Actions**: Undo prior steps before terminal errors
5. **Bounded Retries**: Set max retry duration for time-sensitive operations
6. **Log Appropriately**: Log errors with context for debugging
7. **Graceful Degradation**: Continue with partial data when possible
8. **Error Aggregation**: Decide between fail-fast and partial success

## HTTP Status Codes

Common patterns for terminal error status codes:

```go
// 400 Bad Request - Invalid input
return restate.TerminalError(fmt.Errorf("invalid input"), 400)

// 401 Unauthorized - Auth failure
return restate.TerminalError(fmt.Errorf("unauthorized"), 401)

// 403 Forbidden - Permission denied
return restate.TerminalError(fmt.Errorf("forbidden"), 403)

// 404 Not Found - Resource doesn't exist
return restate.TerminalError(fmt.Errorf("not found"), 404)

// 409 Conflict - State conflict
return restate.TerminalError(fmt.Errorf("conflict"), 409)

// 422 Unprocessable Entity - Semantic error
return restate.TerminalError(fmt.Errorf("unprocessable"), 422)

// 500 Internal Server Error - Unrecoverable error
return restate.TerminalError(fmt.Errorf("internal error"), 500)
```

## References

- Official Docs: https://docs.restate.dev/develop/go/error-handling
- Durable Steps: See restate-go-durable-steps skill
- Service Communication: See restate-go-service-communication skill
- Sagas Pattern: Search Restate documentation for saga guides
