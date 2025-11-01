---
name: restate-go-state
description: Guide for Restate Go state management including storing, retrieving, and clearing state in Virtual Objects and Workflows. Use when implementing stateful services, managing persistent data, or working with key-value storage.
---

# Restate Go State Management

Built-in key-value store for persisting state in Virtual Objects and Workflows.

## Storage & Scope

### Virtual Objects

**Scope**: State is scoped per object key
**Retention**: Retained indefinitely until manually cleared
**Access**: Each object key has its own isolated state

```go
type UserAccount struct{}

func (u *UserAccount) UpdateProfile(ctx restate.ObjectContext, profile Profile) error {
    key := restate.Key(ctx) // e.g., "user-123"
    // State for "user-123" is isolated from "user-456"
    restate.Set(ctx, "profile", profile)
    return nil
}
```

### Workflows

**Scope**: State is scoped per workflow instance
**Retention**: Persists only during configured retention period
**Access**: Available to the Run handler and shared handlers

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    restate.Set(ctx, "status", "processing")
    // State persists during workflow execution + retention period
    return receipt, nil
}
```

## Access Controls

### Exclusive Handlers (Read/Write)

Handlers with write access:
- Virtual Object: `ObjectContext`
- Workflow: `WorkflowContext` (Run handler)

```go
// Virtual Object - exclusive handler
func (c *Counter) Increment(ctx restate.ObjectContext, delta int) error {
    count, _ := restate.Get[int](ctx, "count")
    restate.Set(ctx, "count", count+delta) // Can write
    return nil
}

// Workflow - Run handler
func (w *Workflow) Run(ctx restate.WorkflowContext, input Input) error {
    restate.Set(ctx, "data", input) // Can write
    return nil
}
```

### Shared Handlers (Read-Only)

Handlers with read-only access:
- Virtual Object: `ObjectSharedContext`
- Workflow: `WorkflowSharedContext`

```go
// Virtual Object - shared handler
func (c *Counter) Get(ctx restate.ObjectSharedContext) (int, error) {
    count, _ := restate.Get[int](ctx, "count")
    // Cannot call restate.Set() here
    return count, nil
}

// Workflow - shared handler
func (w *Workflow) GetStatus(ctx restate.WorkflowSharedContext) (string, error) {
    status, _ := restate.Get[string](ctx, "status")
    // Cannot call restate.Set() here
    return status, nil
}
```

## Core Operations

### List All Keys

Retrieve all stored keys for the current object:

```go
keys, err := restate.Keys(ctx)
if err != nil {
    return err
}

for _, key := range keys {
    ctx.Log().Info("Found key", "key", key)
}
```

### Read Values

Get typed values from state:

```go
// String value
name, err := restate.Get[string](ctx, "name")
if err != nil {
    return err
}

// Pointer for nullable values
email, err := restate.Get[*string](ctx, "email")
if err != nil {
    return err
}
if email == nil {
    // Key doesn't exist or value is null
}

// Struct value
type UserProfile struct {
    Name  string
    Email string
    Age   int
}
profile, err := restate.Get[UserProfile](ctx, "profile")

// Slice value
tags, err := restate.Get[[]string](ctx, "tags")
```

### Write Values

Store or update values:

```go
// Simple value
restate.Set(ctx, "username", "alice")

// Struct
profile := UserProfile{Name: "Alice", Email: "alice@example.com"}
restate.Set(ctx, "profile", profile)

// Slice
restate.Set(ctx, "tags", []string{"premium", "verified"})

// Nested structures
settings := map[string]interface{}{
    "theme": "dark",
    "notifications": true,
}
restate.Set(ctx, "settings", settings)
```

### Delete Values

#### Clear Specific Key

```go
err := restate.Clear(ctx, "temporary-data")
if err != nil {
    return err
}
```

#### Clear All State

```go
err := restate.ClearAll(ctx)
if err != nil {
    return err
}
```

## Performance Modes

### Eager Mode (Default)

State loads automatically with the request:

**Advantages**:
- Immediate availability
- No latency on first access
- Predictable performance

**Best for**: Objects with small to medium state size

```go
// State already loaded when handler starts
func (o *Object) Handler(ctx restate.ObjectContext) error {
    // No loading delay
    value, _ := restate.Get[string](ctx, "key")
    return nil
}
```

### Lazy Mode

State fetches on-demand during `Get` calls:

**Advantages**:
- Reduced initial request size
- Faster handler startup for large objects
- Only loads needed keys

**Best for**: Objects with large state or many unused keys

**Configuration**: See service configuration documentation

## Common Patterns

### Counter Pattern

```go
type Counter struct{}

func (c *Counter) Increment(ctx restate.ObjectContext, delta int) (int, error) {
    current, err := restate.Get[int](ctx, "count")
    if err != nil {
        current = 0 // Default if not exists
    }

    newValue := current + delta
    restate.Set(ctx, "count", newValue)

    return newValue, nil
}

func (c *Counter) Reset(ctx restate.ObjectContext) error {
    return restate.Clear(ctx, "count")
}
```

### Shopping Cart

```go
type ShoppingCart struct{}

type CartItem struct {
    ProductID string
    Quantity  int
    Price     float64
}

func (s *ShoppingCart) AddItem(ctx restate.ObjectContext, item CartItem) error {
    items, err := restate.Get[[]CartItem](ctx, "items")
    if err != nil {
        items = []CartItem{}
    }

    items = append(items, item)
    restate.Set(ctx, "items", items)

    return nil
}

func (s *ShoppingCart) GetTotal(ctx restate.ObjectSharedContext) (float64, error) {
    items, err := restate.Get[[]CartItem](ctx, "items")
    if err != nil {
        return 0.0, nil
    }

    var total float64
    for _, item := range items {
        total += item.Price * float64(item.Quantity)
    }

    return total, nil
}

func (s *ShoppingCart) Clear(ctx restate.ObjectContext) error {
    return restate.ClearAll(ctx)
}
```

### Workflow State Tracking

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    // Track progress
    restate.Set(ctx, "status", "payment_processing")

    payment, err := processPayment(ctx, order)
    if err != nil {
        restate.Set(ctx, "status", "payment_failed")
        return Receipt{}, err
    }

    restate.Set(ctx, "status", "fulfillment_processing")
    restate.Set(ctx, "payment_id", payment.ID)

    fulfillment, err := processFulfillment(ctx, order)
    if err != nil {
        restate.Set(ctx, "status", "fulfillment_failed")
        return Receipt{}, err
    }

    restate.Set(ctx, "status", "completed")
    restate.Set(ctx, "fulfillment_id", fulfillment.ID)

    return receipt, nil
}

func (w *OrderWorkflow) GetStatus(ctx restate.WorkflowSharedContext) (string, error) {
    status, err := restate.Get[string](ctx, "status")
    if err != nil {
        return "unknown", err
    }
    return status, nil
}
```

### State Migration

```go
func (o *Object) MigrateData(ctx restate.ObjectContext) error {
    // Read old format
    oldData, err := restate.Get[OldFormat](ctx, "data")
    if err != nil {
        return err
    }

    // Transform to new format
    newData := transformToNewFormat(oldData)

    // Write new format
    restate.Set(ctx, "data_v2", newData)

    // Clean up old format
    restate.Clear(ctx, "data")

    return nil
}
```

### Conditional Updates

```go
func (o *Object) UpdateIfExists(ctx restate.ObjectContext, key string, value string) (bool, error) {
    existing, err := restate.Get[*string](ctx, key)
    if err != nil {
        return false, err
    }

    if existing == nil {
        return false, nil // Key doesn't exist
    }

    restate.Set(ctx, key, value)
    return true, nil
}
```

## Best Practices

1. **Use Appropriate Types**: Use pointers for nullable values
2. **Handle Missing Keys**: Check for nil when reading optional state
3. **Keep State Size Reasonable**: Large state impacts performance
4. **Clear Unused State**: Remove obsolete keys to reduce storage
5. **Use Eager/Lazy Wisely**: Choose based on state size and access patterns
6. **Version State Schema**: Plan for data migration when schema changes
7. **Shared Handlers for Reads**: Use shared handlers when only reading state
8. **Idempotent Updates**: Design state updates to be safely retryable

## Type Safety

State operations are fully type-safe with Go generics:

```go
// Compiler enforces type matching
var count int
count, err = restate.Get[int](ctx, "count")

var profile UserProfile
profile, err = restate.Get[UserProfile](ctx, "profile")

// Type mismatch causes compile error
var wrong string
wrong, err = restate.Get[string](ctx, "count") // Compiles, but may fail at runtime
```

## Serialization

Default serialization uses JSON. State values must be:
- JSON-serializable
- Compatible with Go's `encoding/json` package
- Consistently typed across reads and writes

## References

- Official Docs: https://docs.restate.dev/develop/go/state
- Virtual Objects: See restate-go-services skill
- Workflows: See restate-go-services skill
- Performance Tuning: See Restate server configuration
