---
name: restate-database-patterns
description: Guide for database integration patterns with Restate including durable reads, idempotency, Virtual Object updates, and two-phase commit. Use when integrating databases, ensuring consistency, or implementing exactly-once semantics.
---

# Restate Database Patterns

Integrate databases with Restate for reliable, consistent data operations.

## When to Use What

### Restate State (Built-in K/V)

✅ **Use for**:
- Transactional state machines (order status, payment flow)
- Session data (shopping carts, user sessions)
- Workflow coordination
- Agent/digital twin memory
- Event processing aggregates

### External Databases

✅ **Use for**:
- Complex queries (JOINs, aggregations)
- Core business data accessed by multiple services
- Text search, time-series analysis
- Large datasets requiring specialized storage

## Pattern 1: Simple Database Access

Standard read/write with no special handling:

```go
func (s *UserService) GetUser(ctx restate.Context, userID string) (User, error) {
    // Direct database access
    user, err := restate.Run(ctx, func(ctx restate.RunContext) (User, error) {
        return database.QueryUser(userID)
    })
    return user, err
}
```

## Pattern 2: Durable Reads

Ensure consistent values across retries:

```go
func (s *OrderService) ProcessOrder(ctx restate.Context, orderID string) error {
    // Durable read - same value on retry
    order, err := restate.Run(ctx, func(ctx restate.RunContext) (Order, error) {
        return database.GetOrder(orderID)
    })
    if err != nil {
        return err
    }

    // Business logic based on order
    if order.Status == "pending" {
        // Process...
    }
    return nil
}
```

## Pattern 3: Virtual Object Updates

Serialize access per entity to prevent conflicts:

```go
type AccountService struct{}

func (a *AccountService) UpdateBalance(
    ctx restate.ObjectContext,
    delta int,
) (int, error) {
    accountID := restate.Key(ctx)

    // Read current balance
    currentBalance, err := restate.Run(ctx, func(ctx restate.RunContext) (int, error) {
        return database.GetBalance(accountID)
    })
    if err != nil {
        return 0, err
    }

    newBalance := currentBalance + delta

    // Update with version check
    success, err := restate.Run(ctx, func(ctx restate.RunContext) (bool, error) {
        return database.UpdateBalanceIfUnchanged(accountID, currentBalance, newBalance)
    })
    if err != nil {
        return 0, err
    }

    if !success {
        return 0, restate.TerminalError(
            fmt.Errorf("concurrent modification detected"),
            409,
        )
    }

    return newBalance, nil
}
```

## Pattern 4: Idempotency Key Pattern

Prevent duplicate operations:

```go
func (s *PaymentService) ProcessPayment(
    ctx restate.Context,
    payment Payment,
) (Receipt, error) {
    // Generate deterministic idempotency key
    idempotencyKey := restate.UUID(ctx).String()

    // Check if already processed
    exists, err := restate.Run(ctx, func(ctx restate.RunContext) (bool, error) {
        return database.IdempotencyKeyExists(idempotencyKey)
    })
    if err != nil {
        return Receipt{}, err
    }

    if exists {
        // Already processed, return cached result
        receipt, _ := restate.Run(ctx, func(ctx restate.RunContext) (Receipt, error) {
            return database.GetReceiptByIdempotencyKey(idempotencyKey)
        })
        return receipt, nil
    }

    // Process payment
    receipt, err := restate.Run(ctx, func(ctx restate.RunContext) (Receipt, error) {
        // Atomic: insert idempotency key + process payment
        tx := database.BeginTransaction()
        defer tx.Rollback()

        tx.InsertIdempotencyKey(idempotencyKey)
        receipt := tx.ProcessPayment(payment)

        tx.Commit()
        return receipt, nil
    })

    if err != nil {
        return Receipt{}, err
    }

    // Schedule key expiration
    restate.ServiceSend(ctx, "IdempotencyService", "ExpireKey").
        Send(idempotencyKey, restate.WithDelay(24*time.Hour))

    return receipt, nil
}
```

## Pattern 5: Two-Phase Commit (PostgreSQL)

Exactly-once updates using PREPARE TRANSACTION:

```go
func (s *OrderService) CreateOrder(ctx restate.Context, order Order) error {
    // Generate deterministic transaction ID
    txID := restate.UUID(ctx).String()

    // Phase 1: Prepare transaction (first time only)
    _, err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        db := connectDB()

        // Start transaction
        tx := db.Begin()

        // Insert order
        tx.Exec("INSERT INTO orders (...) VALUES (...)")

        // Prepare (don't commit)
        tx.Exec(fmt.Sprintf("PREPARE TRANSACTION '%s'", txID))

        return restate.Void{}, nil
    })
    if err != nil {
        return err
    }

    // Phase 2: Commit prepared transaction (idempotent)
    _, err = restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        db := connectDB()

        // Commit prepared transaction (safe to retry)
        db.Exec(fmt.Sprintf("COMMIT PREPARED '%s'", txID))

        return restate.Void{}, nil
    })

    return err
}
```

**PostgreSQL Configuration**:
```sql
-- Enable prepared transactions
ALTER SYSTEM SET max_prepared_transactions = 100;
SELECT pg_reload_conf();
```

## Best Practices

1. **Wrap All Database Calls**: Use `restate.Run()` for durability
2. **Use Deterministic Keys**: For idempotency and transactions
3. **Version Checks**: Prevent lost updates with optimistic locking
4. **Virtual Objects**: Serialize access per entity
5. **Connection Pooling**: Manage connections efficiently
6. **Timeouts**: Set appropriate database timeouts
7. **Retry Logic**: Let Restate handle retries automatically

## References

- Official Docs: https://docs.restate.dev/guides/databases
- Durable Steps: See restate-go-durable-steps skill
- Virtual Objects: See restate-go-services skill
- Error Handling: See restate-go-error-handling skill
