---
name: restate-go-logging
description: Guide for Restate Go logging using slog, context-based logging, and replay suppression. Use when implementing logging, debugging handlers, or configuring log output.
---

# Restate Go Logging

Implement structured logging in Restate handlers with automatic replay suppression.

## Logging Foundation

Restate's Go SDK uses the standard library's `log/slog` package for all logging functionality.

### Key Features

- Standard Go `log/slog` interface
- Custom handler support
- Automatic replay suppression
- Context-aware logging
- Structured logging

## Custom Log Handlers

Provide custom slog handlers when initializing the Restate server.

### JSON Handler

```go
import (
    "log/slog"
    "os"

    "github.com/restatedev/sdk-go"
)

func main() {
    // Create JSON handler
    jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

    // Configure server with custom logger
    server := restate.NewServer(
        restate.WithPort(9080),
        restate.WithLogger(jsonHandler),
    )

    server.Bind(restate.Reflect(&MyService{}))
    server.Start()
}
```

### Text Handler (Pretty Logging)

```go
func main() {
    // Create text handler for human-readable output
    textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })

    server := restate.NewServer(
        restate.WithPort(9080),
        restate.WithLogger(textHandler),
    )

    server.Bind(restate.Reflect(&MyService{}))
    server.Start()
}
```

### Custom Handler Options

```go
func main() {
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:     slog.LevelInfo,           // Minimum log level
        AddSource: true,                      // Include source code location
        ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
            // Customize attribute formatting
            if a.Key == slog.TimeKey {
                // Custom time format
                return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
            }
            return a
        },
    })

    server := restate.NewServer(restate.WithLogger(handler))
    server.Bind(restate.Reflect(&MyService{}))
    server.Start()
}
```

## Context-Based Logging

Use `ctx.Log()` for context-aware logging with automatic replay suppression.

### Basic Usage

```go
type MyService struct{}

func (s *MyService) Process(ctx restate.Context, data Data) (Result, error) {
    // Get logger from context
    logger := ctx.Log()

    logger.Info("Processing data", "data_id", data.ID)

    result, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
        return processData(data)
    })

    if err != nil {
        logger.Error("Processing failed", "error", err)
        return Result{}, err
    }

    logger.Info("Processing completed", "result_id", result.ID)
    return result, nil
}
```

### Log Levels

```go
func (s *Service) Handler(ctx restate.Context, input Input) error {
    logger := ctx.Log()

    // Debug level
    logger.Debug("Detailed debugging info", "input", input)

    // Info level (default)
    logger.Info("Operation started", "operation", "process")

    // Warn level
    logger.Warn("Potential issue detected", "issue", "timeout_approaching")

    // Error level
    logger.Error("Operation failed", "error", err, "retry_count", retries)

    return nil
}
```

### Automatic Replay Suppression

**Key Feature**: Context logger automatically suppresses duplicate log statements during replay cycles.

```go
func (s *Service) Example(ctx restate.Context) error {
    // This log appears ONLY on first execution
    // NOT on replay after failures
    ctx.Log().Info("This will not be printed again during replays")

    // Do work
    _, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
        return externalCall()
    })

    // This also suppressed on replay
    ctx.Log().Info("Completed external call")

    return err
}
```

**Why it matters**: Prevents log spam during automatic retries and deterministic replay.

## Structured Logging

Add structured attributes to logs for better searchability and analysis.

### Key-Value Pairs

```go
func (s *Service) CreateOrder(ctx restate.Context, order Order) error {
    ctx.Log().Info("Creating order",
        "order_id", order.ID,
        "user_id", order.UserID,
        "amount", order.Amount,
        "currency", order.Currency,
    )

    return nil
}
```

### Nested Attributes

```go
func (s *Service) Process(ctx restate.Context, request Request) error {
    ctx.Log().Info("Processing request",
        "request_id", request.ID,
        "metadata", slog.Group("metadata",
            "source", request.Source,
            "timestamp", request.Timestamp,
            "version", request.Version,
        ),
        "user", slog.Group("user",
            "id", request.User.ID,
            "email", request.User.Email,
        ),
    )

    return nil
}
```

### Error Logging

```go
func (s *Service) Risky(ctx restate.Context) error {
    result, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
        return riskyOperation()
    })

    if err != nil {
        ctx.Log().Error("Operation failed",
            "error", err,
            "error_type", fmt.Sprintf("%T", err),
            "retry_count", getRetryCount(),
        )
        return err
    }

    return nil
}
```

## Advanced Handler Configuration

### Conditional Replay Suppression

Control replay suppression behavior manually.

```go
import "github.com/restatedev/sdk-go/rcontext"

func main() {
    handler := slog.NewJSONHandler(os.Stdout, nil)

    // Mode 1: Automatic suppression (default, pass true)
    server := restate.NewServer(restate.WithLogger(handler)) // Default: drops replays

    // Mode 2: Manual control (pass false)
    server := restate.NewServer(restate.WithLogger(handler)) // Custom handling
}
```

### Custom Replay Handling

```go
import (
    "log/slog"
    "github.com/restatedev/sdk-go/rcontext"
)

type ReplayAwareHandler struct {
    underlying slog.Handler
}

func (h *ReplayAwareHandler) Handle(ctx context.Context, r slog.Record) error {
    // Check if this would be suppressed during replay
    if rcontext.LogContextFrom(ctx) {
        // Custom logic for replay logs
        r.AddAttrs(slog.Bool("is_replay", true))
    }

    return h.underlying.Handle(ctx, r)
}

func (h *ReplayAwareHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &ReplayAwareHandler{underlying: h.underlying.WithAttrs(attrs)}
}

func (h *ReplayAwareHandler) WithGroup(name string) slog.Handler {
    return &ReplayAwareHandler{underlying: h.underlying.WithGroup(name)}
}

func (h *ReplayAwareHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.underlying.Enabled(ctx, level)
}
```

## Common Patterns

### Operation Tracing

```go
func (s *Service) ComplexOperation(ctx restate.Context, input Input) (Output, error) {
    logger := ctx.Log()

    logger.Info("Operation started",
        "operation", "complex_operation",
        "input_id", input.ID,
    )

    // Step 1
    logger.Debug("Step 1: Validation")
    if err := validate(input); err != nil {
        logger.Error("Validation failed", "error", err)
        return Output{}, restate.TerminalError(err, 400)
    }

    // Step 2
    logger.Debug("Step 2: External call")
    data, err := restate.Run(ctx, func(ctx restate.RunContext) (Data, error) {
        return externalAPI.Fetch(input.ID)
    })
    if err != nil {
        logger.Error("External call failed", "error", err)
        return Output{}, err
    }

    // Step 3
    logger.Debug("Step 3: Processing")
    result := process(data)

    logger.Info("Operation completed",
        "operation", "complex_operation",
        "output_id", result.ID,
    )

    return result, nil
}
```

### Performance Logging

```go
import "time"

func (s *Service) TimedOperation(ctx restate.Context, task Task) error {
    logger := ctx.Log()
    start := time.Now()

    logger.Info("Starting timed operation", "task_id", task.ID)

    err := performTask(ctx, task)

    duration := time.Since(start)
    logger.Info("Operation completed",
        "task_id", task.ID,
        "duration_ms", duration.Milliseconds(),
        "success", err == nil,
    )

    return err
}
```

### State Change Logging

```go
func (o *Object) UpdateState(ctx restate.ObjectContext, newValue Value) error {
    logger := ctx.Log()

    oldValue, _ := restate.Get[Value](ctx, "value")

    logger.Info("State update",
        "object_key", restate.Key(ctx),
        "old_value", oldValue,
        "new_value", newValue,
    )

    restate.Set(ctx, "value", newValue)

    return nil
}
```

### Workflow Progress Logging

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    logger := ctx.Log()
    workflowID := restate.Key(ctx)

    logger.Info("Workflow started",
        "workflow_id", workflowID,
        "order_id", order.ID,
    )

    // Payment
    logger.Info("Processing payment", "workflow_id", workflowID)
    payment, err := processPayment(ctx, order)
    if err != nil {
        logger.Error("Payment failed",
            "workflow_id", workflowID,
            "error", err,
        )
        return Receipt{}, err
    }

    // Fulfillment
    logger.Info("Processing fulfillment", "workflow_id", workflowID)
    fulfillment, err := processFulfillment(ctx, order)
    if err != nil {
        logger.Error("Fulfillment failed",
            "workflow_id", workflowID,
            "error", err,
        )
        return Receipt{}, err
    }

    logger.Info("Workflow completed",
        "workflow_id", workflowID,
        "payment_id", payment.ID,
        "fulfillment_id", fulfillment.ID,
    )

    return createReceipt(payment, fulfillment), nil
}
```

## Integration with Third-Party Loggers

### Zerolog Example

```go
import (
    "log/slog"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// Zerolog to slog adapter
type ZerologHandler struct {
    logger zerolog.Logger
}

func (h *ZerologHandler) Handle(ctx context.Context, r slog.Record) error {
    event := h.logger.WithLevel(zerologLevel(r.Level))

    r.Attrs(func(a slog.Attr) bool {
        event = event.Interface(a.Key, a.Value.Any())
        return true
    })

    event.Msg(r.Message)
    return nil
}

// Implement other slog.Handler methods...

func main() {
    zerologger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    handler := &ZerologHandler{logger: zerologger}

    server := restate.NewServer(restate.WithLogger(handler))
    server.Bind(restate.Reflect(&MyService{}))
    server.Start()
}
```

## Best Practices

1. **Use Context Logger**: Always use `ctx.Log()` for replay suppression
2. **Structured Logging**: Use key-value pairs instead of string formatting
3. **Appropriate Levels**: Use correct log levels (Debug/Info/Warn/Error)
4. **Add Context**: Include relevant IDs and metadata
5. **Error Details**: Log errors with full context
6. **Performance**: Log durations for time-sensitive operations
7. **State Changes**: Log important state transitions
8. **Avoid Secrets**: Never log sensitive data (passwords, tokens, keys)

## Logging vs Println

**Don't use**:
```go
fmt.Println("Processing data")  // ❌ Not structured, appears on replay
log.Println("Error occurred")    // ❌ Not replay-aware
```

**Use instead**:
```go
ctx.Log().Info("Processing data")           // ✅ Structured, replay-suppressed
ctx.Log().Error("Error occurred", "err", err) // ✅ Proper error logging
```

## References

- Official Docs: https://docs.restate.dev/develop/go/logging
- Go slog: https://pkg.go.dev/log/slog
- Serving: See restate-go-serving skill
- Services: See restate-go-services skill
