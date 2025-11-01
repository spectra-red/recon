---
name: restate-invocation-management
description: Guide for managing Restate invocations including cancellation, kill, resume, purge, and status checking. Use when managing running invocations, handling stuck handlers, or cleaning up completed invocations.
---

# Managing Restate Invocations

Manage the lifecycle of Restate invocations through cancellation, resumption, and cleanup operations.

## Core Concepts

### Invocation ID

Each invocation receives a unique identifier:
- Format: `inv_<unique_id>`
- Visible in: Restate UI, logs, CLI commands
- Used for: Tracking, cancellation, attachment

### Invocation Lifecycle

Invocations progress through well-defined states:
- **Pending**: Queued for execution
- **Running**: Currently executing
- **Suspended**: Waiting for external events or timers
- **Completed**: Successfully finished
- **Failed**: Terminated with error

## Management Operations

### Cancellation

Cancel an invocation at any point in its lifecycle.

#### CLI

```bash
restate invocations cancel inv_<ID>
```

#### HTTP API

```bash
curl -X PATCH http://localhost:9070/invocations/<ID>/cancel
```

#### Go SDK

```go
import "github.com/restatedev/sdk-go/ingress"

client := restateingress.NewClient("http://localhost:8080")
err := restateingress.CancelInvocation(client, invocationID)
```

#### From Within Handler

```go
func (s *Service) CancelRelatedTask(ctx restate.Context, taskID string) error {
    // Cancel another invocation
    err := restate.CancelInvocation(ctx, invocationID)
    return err
}
```

### How Cancellation Works

1. **Non-blocking**: API returns immediately, cancellation may complete later
2. **Cooperative**: Runtime forwards cancellation to SDK
3. **At Await Points**: Exception surfaces at next await (e.g., `ctx.run()`, service calls)
4. **Compensation**: Handlers can catch cancellation to execute cleanup logic

### Handling Cancellation in Handlers

```go
func (s *Service) CancellableOperation(ctx restate.Context, data Data) error {
    // Step 1: Reserve resource
    resourceID, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
        return reserveResource()
    })
    if err != nil {
        return err
    }

    // Step 2: Long-running operation
    // Cancellation will surface here as error
    result, err := restate.Service[Result](ctx, "ProcessorService", "Process").
        Request(data)

    if err != nil {
        // Check if cancellation occurred
        ctx.Log().Warn("Operation cancelled or failed, releasing resource")

        // Compensate: release resource
        restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            releaseResource(resourceID)
            return restate.Void{}, nil
        })

        return err
    }

    return nil
}
```

### Kill

Immediately stop an invocation without compensation.

**Use when**: Cancellation fails (e.g., endpoints permanently unavailable)

**Warning**: May leave services in inconsistent states. Use as last resort only.

#### CLI

```bash
restate invocations kill inv_<ID>
```

#### HTTP API

```bash
curl -X DELETE http://localhost:9070/invocations/<ID>
```

### Resume

Manually resume paused invocations after excessive retries.

#### Basic Resume

```bash
restate invocations resume inv_<ID>
```

#### HTTP API

```bash
curl -X PATCH http://localhost:9070/invocations/<ID>/resume
```

#### Resume at Different Deployment

```bash
# Override deployment for resumption
restate invocations resume inv_<ID> --deployment <DEPLOYMENT_ID>
```

**Warning**: Resuming at different deployment version risks non-determinism errors if business logic differs.

### Purge

Remove completed invocations to free disk space.

**When to use**: After retention period, for cleanup

#### CLI

```bash
restate invocations purge inv_<ID>
```

#### HTTP API

```bash
curl -X DELETE http://localhost:9070/invocations/<ID>/purge
```

### Restart as New

Restart a completed invocation with a new ID, preserving original inputs.

**Note**: Not available for workflows

#### CLI

```bash
restate invocations restart inv_<ID>
```

**Behavior**:
- Creates new invocation ID
- Re-executes all operations
- Uses original input parameters

## Invocation Introspection

### Get Invocation Status

#### CLI

```bash
restate invocations describe inv_<ID>
```

#### HTTP API

```bash
curl http://localhost:9070/invocations/<ID>
```

### List Invocations

#### CLI

```bash
# List all invocations
restate invocations list

# Filter by service
restate invocations list --service MyService

# Filter by status
restate invocations list --status running
```

#### HTTP API

```bash
# List invocations
curl http://localhost:9070/invocations

# With filters
curl "http://localhost:9070/invocations?service=MyService&status=running"
```

## Common Patterns

### Timeout and Cancel Pattern

```go
func (s *Service) TimeBoundOperation(ctx restate.Context, data Data) error {
    // Start async operation
    future := restate.Service[Result](ctx, "SlowService", "Process").
        RequestFuture(data)

    // Create timeout
    timeout := restate.After(ctx, 30*time.Second)

    // Race operation against timeout
    selected, err := restate.WaitFirst(ctx, timeout, future)
    if err != nil {
        return err
    }

    if selected.Index() == 0 {
        // Timeout occurred - could cancel the operation here
        ctx.Log().Warn("Operation timed out")
        return fmt.Errorf("operation exceeded timeout")
    }

    // Operation completed
    result, err := future.Done()
    return err
}
```

### Workflow Cancellation with Cleanup

```go
type OrderWorkflow struct{}

func (w *OrderWorkflow) Run(ctx restate.WorkflowContext, order Order) (Receipt, error) {
    // Reserve inventory
    reservationID, err := restate.Service[string](ctx, "Inventory", "Reserve").
        Request(order.Items)
    if err != nil {
        return Receipt{}, err
    }

    // Process payment (cancellable)
    payment, err := restate.Service[Payment](ctx, "PaymentService", "Charge").
        Request(order.PaymentInfo)

    if err != nil {
        // Cancelled or failed - release inventory
        ctx.Log().Info("Payment failed, releasing inventory")
        restate.Service[void](ctx, "Inventory", "Release").
            Request(reservationID)

        return Receipt{}, err
    }

    return createReceipt(payment), nil
}

// External cancellation handler
func (w *OrderWorkflow) Cancel(ctx restate.WorkflowSharedContext, reason string) error {
    // Mark workflow as cancelled
    restate.Set(ctx, "cancelled", true)
    restate.Set(ctx, "cancel_reason", reason)

    // Note: Main Run handler will handle cleanup when it receives cancellation
    return nil
}
```

### Monitoring and Auto-Recovery

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"
)

type InvocationStatus struct {
    ID     string
    Status string
    Retry  int
}

func monitorInvocations() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        // Fetch stuck invocations
        resp, err := http.Get("http://localhost:9070/invocations?status=suspended")
        if err != nil {
            log.Printf("Failed to fetch invocations: %v", err)
            continue
        }

        var invocations []InvocationStatus
        json.NewDecoder(resp.Body).Decode(&invocations)
        resp.Body.Close()

        for _, inv := range invocations {
            // Check if invocation has been stuck too long
            if inv.Retry > 10 {
                log.Printf("Resuming stuck invocation: %s", inv.ID)

                // Resume invocation
                req, _ := http.NewRequest(
                    "PATCH",
                    "http://localhost:9070/invocations/"+inv.ID+"/resume",
                    nil,
                )
                http.DefaultClient.Do(req)
            }
        }
    }
}
```

### Bulk Cancellation

```bash
#!/bin/bash

# Cancel all invocations for a specific service
SERVICE_NAME="MyService"

restate invocations list --service $SERVICE_NAME --status running | \
  jq -r '.[] | .id' | \
  xargs -I {} restate invocations cancel {}
```

### Invocation Cleanup Script

```bash
#!/bin/bash

# Purge completed invocations older than 7 days
CUTOFF_DATE=$(date -d '7 days ago' -u +%Y-%m-%dT%H:%M:%SZ)

restate invocations list --status completed | \
  jq -r --arg cutoff "$CUTOFF_DATE" \
    '.[] | select(.completed_at < $cutoff) | .id' | \
  xargs -I {} restate invocations purge {}
```

## Idempotent Deduplication

Restate retains completed invocations for idempotent deduplication.

### How It Works

```go
// First call
response1, _ := restateingress.Service[Input, Output](
    client,
    "PaymentService",
    "Charge",
).Request(ctx, input, restate.WithIdempotencyKey("payment-123"))

// Duplicate call - returns cached response
response2, _ := restateingress.Service[Input, Output](
    client,
    "PaymentService",
    "Charge",
).Request(ctx, input, restate.WithIdempotencyKey("payment-123"))

// response1 == response2 (from cache, not re-executed)
```

### Retention Period

- Default: 24 hours
- Configurable via server settings
- Invocations can be purged manually after retention

## Best Practices

1. **Use Cancellation First**: Try cancellation before kill
2. **Implement Compensation**: Handle cancellation gracefully in handlers
3. **Monitor Stuck Invocations**: Set up alerting for long-running invocations
4. **Clean Up Regularly**: Purge old completed invocations to free disk space
5. **Idempotency Keys**: Use for critical operations to enable deduplication
6. **Log Cancellations**: Add logging when handling cancellation
7. **Test Cancellation**: Verify your handlers handle cancellation correctly
8. **Avoid Kill**: Only use kill as absolute last resort

## Admin API Endpoints

### Base URL

```
http://localhost:9070
```

### Endpoints

```bash
# Get invocation details
GET /invocations/<ID>

# List invocations
GET /invocations?service=<SERVICE>&status=<STATUS>

# Cancel invocation
PATCH /invocations/<ID>/cancel

# Kill invocation
DELETE /invocations/<ID>

# Resume invocation
PATCH /invocations/<ID>/resume

# Purge invocation
DELETE /invocations/<ID>/purge
```

## CLI Commands Reference

```bash
# List invocations
restate invocations list [--service SERVICE] [--status STATUS]

# Describe invocation
restate invocations describe <ID>

# Cancel invocation
restate invocations cancel <ID>

# Kill invocation
restate invocations kill <ID>

# Resume invocation
restate invocations resume <ID> [--deployment DEPLOYMENT]

# Purge invocation
restate invocations purge <ID>

# Restart invocation
restate invocations restart <ID>
```

## Troubleshooting

### Invocation Stuck in Retry Loop

**Problem**: Invocation keeps retrying and failing

**Solutions**:
1. Check logs for error details
2. Fix underlying issue (service down, bad data)
3. Resume after fix, or
4. Cancel if unrecoverable

### Cancellation Not Working

**Problem**: Invocation doesn't cancel

**Possible causes**:
- Service endpoint unavailable
- Invocation in tight loop without await points

**Solutions**:
1. Wait for service to become available
2. Use kill if endpoint permanently down
3. Ensure handlers have await points for cancellation

### Workflow Won't Complete

**Problem**: Workflow stuck waiting for external event

**Solutions**:
1. Check if promise/awakeable was resolved
2. Verify external system integration
3. Cancel workflow if unrecoverable
4. Fix external system and resume

## References

- Official Docs: https://docs.restate.dev/services/invocation/managing-invocations
- Go SDK Client: See restate-go-client skill
- Service Communication: See restate-go-service-communication skill
- External Events: See restate-go-external-events skill
