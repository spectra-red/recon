---
name: restate-go-services
description: Guide for creating Restate Go services including basic services, virtual objects, and workflows. Use when building Restate services, implementing handlers, or working with service types in Go.
---

# Restate Go Services

Create and manage Restate services in Go with proper service types, handlers, and context usage.

## Getting Started

Add the Restate Go SDK (requires Go 1.21.0+):

```bash
go get github.com/restatedev/sdk-go
```

## Three Service Types

### Basic Services

Group related handlers as callable endpoints. Handlers use `Context` for durability through Restate's journaling system.

```go
type MyService struct{}

func (s *MyService) MyHandler(ctx restate.Context, input string) (string, error) {
    // Handler logic here
    return "response", nil
}
```

**Service URL**: `MyService/MyHandler` at the Restate ingress URL

### Virtual Objects

Provide stateful, key-addressable services where each instance maintains persistent state.

```go
type Counter struct{}

// Exclusive handler - can read and write state
func (c *Counter) Increment(ctx restate.ObjectContext, delta int) (int, error) {
    key := restate.Key(ctx) // Get the object key
    count, _ := restate.Get[int](ctx, "count")
    newCount := count + delta
    restate.Set(ctx, "count", newCount)
    return newCount, nil
}

// Shared handler - read-only access
func (c *Counter) Get(ctx restate.ObjectSharedContext) (int, error) {
    count, _ := restate.Get[int](ctx, "count")
    return count, nil
}
```

**Key Features**:
- Retrieve object key via `restate.Key(ctx)`
- Exclusive handlers use `ObjectContext` (read/write)
- Shared handlers use `ObjectSharedContext` (read-only)

### Workflows

Orchestrate long-lived, multi-step operations with exactly-once execution.

```go
type OrderWorkflow struct{}

// Required: Run handler executes exactly once
func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    // Main workflow logic
    // Executes exactly once per workflow
    return receipt, nil
}

// Additional handlers must use WorkflowSharedContext
func (w *OrderWorkflow) GetStatus(ctx restate.WorkflowSharedContext) (string, error) {
    status, _ := restate.Get[string](ctx, "status")
    return status, nil
}
```

**Key Features**:
- `Run` handler required as main entry point
- Executes exactly once per workflow
- Additional handlers use `WorkflowSharedContext`
- Can signal or query workflow state

## Handler Fundamentals

### Handler Signature

```go
func (s *MyService) HandlerName(ctx restate.Context, input InputType) (OutputType, error)
```

**Components**:
- Context parameter (first argument)
- Input arguments (optional)
- Return result (optional)
- Return error (optional)

### Serialization

**Default**: JSON serialization
**Alternative**: Protocol Buffers for code generation (see code generation skill)

### Service Registration

Use `restate.Reflect` to convert struct methods into handlers:

```go
server := restate.NewServer(restate.WithPort(9080))
server.Bind(restate.Reflect(&MyService{}))
server.Start()
```

The SDK automatically skips:
- Unexported methods
- Methods with incorrect signatures

## Service Patterns

### Input/Output Flexibility

```go
// No input, no output
func (s *Service) DoSomething(ctx restate.Context) error

// Input only
func (s *Service) Process(ctx restate.Context, data string) error

// Output only
func (s *Service) Generate(ctx restate.Context) (string, error)

// Both input and output
func (s *Service) Transform(ctx restate.Context, input Data) (Result, error)
```

### Void Returns

For handlers with no meaningful return value:

```go
func (s *Service) Fire(ctx restate.Context, event Event) (restate.Void, error) {
    // Process event
    return restate.Void{}, nil
}
```

## Workflow Resubmission

Workflows track request headers to prevent duplicate execution on resubmission. The framework ensures exactly-once semantics even if the same request is submitted multiple times.

## Best Practices

1. **Service Naming**: Use descriptive struct names - they become part of the service URL
2. **Handler Granularity**: Keep handlers focused on single responsibilities
3. **Context Usage**: Always use the provided context, never create your own
4. **Error Handling**: Return errors for failures that should trigger retries
5. **Virtual Object Keys**: Design meaningful key schemes for partitioning

## Common Use Cases

**Basic Services**:
- Stateless operations
- External API integrations
- Data transformations

**Virtual Objects**:
- User sessions
- Shopping carts
- Account management
- Entity-specific state

**Workflows**:
- Order processing
- Approval flows
- Multi-step orchestrations
- Saga patterns

## Additional Configuration

For advanced options:
- Timeouts and retention policies: See service configuration docs
- Code generation: See Protocol Buffers skill
- Deployment: See serving skill

## Example: Complete Service

```go
package main

import (
    "github.com/restatedev/sdk-go"
)

type GreeterService struct{}

func (g *GreeterService) Greet(ctx restate.Context, name string) (string, error) {
    // Use durable execution
    greeting, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
        return "Hello, " + name + "!", nil
    })
    if err != nil {
        return "", err
    }
    return greeting, nil
}

func main() {
    server := restate.NewServer(restate.WithPort(9080))
    server.Bind(restate.Reflect(&GreeterService{}))
    server.Start()
}
```

## References

- Official Docs: https://docs.restate.dev/develop/go/services
- SDK Repository: https://github.com/restatedev/sdk-go
