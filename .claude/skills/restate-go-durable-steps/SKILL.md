---
name: restate-go-durable-steps
description: Guide for using Restate Go durable steps to wrap non-deterministic operations like HTTP calls, database queries, and random generation. Use when implementing durable execution, handling side effects, or ensuring deterministic replay.
---

# Restate Go Durable Steps

Persist operation results using durable execution logs that replay safely after failures and suspensions.

## What Are Durable Steps?

Restate uses an execution log to persist operation results. The framework requires wrapping non-deterministic operations to ensure deterministic replay during retries.

**Non-deterministic operations include**:
- HTTP requests
- Database calls
- UUID generation
- Random number generation
- Current time/date retrieval
- External API calls

## The Run Function

Safely wrap non-deterministic operations with `restate.Run`:

```go
result, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
    // Non-deterministic operation
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    return string(body), nil
})
```

### Key Characteristics

**Persistence**: Results are stored in the execution log
**Replay Safety**: On retry, logged results are returned without re-execution
**Isolation**: Cannot access standard Restate context methods inside `Run`

### Critical Constraint

Inside `Run`, you work exclusively with the `RunContext` provided to your function. You **cannot**:
- Access the parent Restate context
- Make nested Restate calls
- Use state operations
- Call other handlers

This ensures proper journaling and deterministic replay.

## Return Values

### Serializable Payloads

The function accepts any serializable payload using `JSONCodec` by default:

```go
// String result
message, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
    return "Hello, World!", nil
})

// Struct result
type User struct {
    ID   string
    Name string
}

user, err := restate.Run(ctx, func(ctx restate.RunContext) (User, error) {
    return User{ID: "123", Name: "Alice"}, nil
})

// Slice result
items, err := restate.Run(ctx, func(ctx restate.RunContext) ([]string, error) {
    return []string{"item1", "item2"}, nil
})
```

### Void Returns

When no return value is needed:

```go
_, err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
    // Perform side effect
    http.Post("https://api.example.com/webhook", "application/json", bytes.NewBuffer(data))
    return restate.Void{}, nil
})
```

## Error Handling & Retries

### Automatic Retries

Failures in `Run` trigger automatic retries following standard handler error treatment:

```go
data, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
    resp, err := fetchExternalData()
    if err != nil {
        // This will trigger retry
        return "", err
    }
    return resp, nil
})
```

### Terminal Errors

Use `TerminalError` to stop retries immediately:

```go
data, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
    resp, err := fetchExternalData()
    if err != nil {
        if isUnrecoverable(err) {
            return "", restate.TerminalError(err, 400)
        }
        return "", err // Will retry
    }
    return resp, nil
})
```

### Custom Retry Policies

Configure retry behavior with options:

```go
result, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
    return performOperation()
},
    restate.WithName("my-operation"),              // For observability
    restate.WithMaxRetryDuration(5*time.Minute),   // Maximum retry duration
    restate.WithRetryInterval(1*time.Second),      // Initial retry interval
    restate.WithRetryIntervalMultiplier(2.0),      // Exponential backoff factor
)
```

When the retry policy is exhausted, a `TerminalError` is thrown.

## Deterministic Random Operations

For values that must remain consistent across retries, use Restate's seeded helpers.

### UUIDs

Generate stable UUIDs using invocation ID as seed:

```go
uuid := restate.UUID(ctx)
// Same UUID on every replay of this step
```

**Use case**: Idempotency keys, unique identifiers that must persist across retries

### Random Numbers

```go
rng := restate.Rand(ctx)

// Random uint64
randomInt := rng.Uint64()

// Random float64 [0.0, 1.0)
randomFloat := rng.Float64()
```

### Custom Random Generation

Seed the standard library's random generator:

```go
import "math/rand/v2"

rng := restate.Rand(ctx)
customRand := rand.New(rand.NewPCG(rng.Uint64(), rng.Uint64()))
```

**Important**: These are seeded by the invocation ID, ensuring identical results on replay.

## Common Patterns

### Database Query

```go
user, err := restate.Run(ctx, func(ctx restate.RunContext) (User, error) {
    var user User
    err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID).Scan(&user)
    if err != nil {
        return User{}, err
    }
    return user, nil
})
```

### HTTP API Call

```go
apiResponse, err := restate.Run(ctx, func(ctx restate.RunContext) (APIResponse, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get("https://api.example.com/resource")
    if err != nil {
        return APIResponse{}, err
    }
    defer resp.Body.Close()

    var result APIResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
})
```

### External Service with Retry Policy

```go
receipt, err := restate.Run(ctx, func(ctx restate.RunContext) (Receipt, error) {
    return paymentService.Charge(amount, cardToken)
},
    restate.WithName("payment-charge"),
    restate.WithMaxRetryDuration(3*time.Minute),
    restate.WithRetryInterval(500*time.Millisecond),
)
```

### Idempotent External Call

```go
// Generate stable idempotency key
idempotencyKey := restate.UUID(ctx).String()

result, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
    return externalAPI.CallWithIdempotency(idempotencyKey, payload)
})
```

## Best Practices

1. **Always Use Run**: Wrap all non-deterministic operations, no exceptions
2. **Keep Run Functions Small**: Single responsibility per Run block
3. **Name Operations**: Use `WithName()` for observability and debugging
4. **Handle Errors Properly**: Distinguish between retriable and terminal errors
5. **Use Deterministic Helpers**: Never use `uuid.New()` or `rand.Intn()` directly
6. **Avoid Nested Context**: Don't pass the parent context into Run blocks
7. **Set Appropriate Timeouts**: Configure retry policies based on operation criticality

## What NOT to Do

```go
// BAD: Using non-deterministic UUID
uuid := uuid.New() // Different on each replay!

// GOOD: Using deterministic UUID
uuid := restate.UUID(ctx) // Same on each replay

// BAD: Using standard random
random := rand.Intn(100) // Different on each replay!

// GOOD: Using deterministic random
rng := restate.Rand(ctx)
random := rng.Uint64() // Same on each replay

// BAD: Accessing parent context
restate.Run(ctx, func(runCtx restate.RunContext) (string, error) {
    restate.Get[string](ctx, "key") // ERROR! Wrong context
})

// GOOD: Only use RunContext inside Run
restate.Run(ctx, func(runCtx restate.RunContext) (string, error) {
    // Use only runCtx or external resources
    return fetchData(), nil
})
```

## Observability

Name your operations for better tracing and debugging:

```go
restate.Run(ctx, func(ctx restate.RunContext) (Data, error) {
    return fetchUserData(userID)
}, restate.WithName("fetch-user-data"))
```

This appears in logs, traces, and the Restate UI for easier troubleshooting.

## References

- Official Docs: https://docs.restate.dev/develop/go/durable-steps
- Error Handling: See restate-go-error-handling skill
- Concurrent Operations: See restate-go-concurrent-tasks skill
