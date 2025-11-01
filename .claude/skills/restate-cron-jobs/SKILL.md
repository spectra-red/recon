---
name: restate-cron-jobs
description: Guide for implementing cron jobs and scheduled tasks with Restate using durable timers and Virtual Objects. Use when implementing scheduled tasks, recurring jobs, or time-based automation.
---

# Restate Cron Jobs

Implement reliable scheduled tasks with automatic retries and durable execution.

## Overview

Restate doesn't have built-in cron functionality but provides building blocks for reliable scheduling:
- **Durable timers**: Survive crashes and restarts
- **Virtual Objects**: Manage individual job state
- **Automatic retries**: Ensure jobs complete
- **Horizontal scalability**: Run many jobs in parallel

## Architecture

### Two Components

**1. CronJobInitiator Service**: Creates and assigns job IDs
**2. CronJob Virtual Object**: Manages individual job execution

### How It Works

```
User → CronJobInitiator.create() →Generate unique ID
          ↓
    CronJob VirtualObject initialized
          ↓
    Calculate next execution time
          ↓
    Schedule delayed invocation
          ↓
    Execute job → Reschedule itself
```

## Go Implementation

### Job Request Structure

```go
type CronJobRequest struct {
    CronExpression string      `json:"cronExpression"`  // "0 0 * * *"
    Service        string      `json:"service"`          // Target service
    Method         string      `json:"method"`           // Handler to call
    Key            string      `json:"key,omitempty"`    // For Virtual Objects
    Payload        interface{} `json:"payload,omitempty"` // Job parameters
}

type CronJobInfo struct {
    ID             string
    CronExpression string
    NextExecution  time.Time
    LastExecution  time.Time
    Status         string
    ExecutionCount int
}
```

### CronJobInitiator Service

```go
type CronJobInitiator struct{}

func (c *CronJobInitiator) Create(
    ctx restate.Context,
    request CronJobRequest,
) (string, error) {
    // Generate unique job ID
    jobID := restate.UUID(ctx).String()

    // Initialize the CronJob Virtual Object
    _, err := restate.Object[CronJobInfo](
        ctx,
        "CronJob",
        jobID,
        "Initialize",
    ).Request(request)

    if err != nil {
        return "", err
    }

    return jobID, nil
}
```

### CronJob Virtual Object

```go
type CronJob struct{}

func (c *CronJob) Initialize(
    ctx restate.ObjectContext,
    request CronJobRequest,
) (CronJobInfo, error) {
    jobID := restate.Key(ctx)

    // Store job configuration
    restate.Set(ctx, "config", request)
    restate.Set(ctx, "status", "active")
    restate.Set(ctx, "execution_count", 0)

    // Parse cron expression
    schedule, err := restate.Run(ctx, func(ctx restate.RunContext) (*cron.Schedule, error) {
        return cron.ParseStandard(request.CronExpression)
    })
    if err != nil {
        return CronJobInfo{}, restate.TerminalError(err, 400)
    }

    // Calculate next execution
    now := time.Now()
    nextExec := schedule.Next(now)

    restate.Set(ctx, "next_execution", nextExec)

    // Schedule first execution
    delay := nextExec.Sub(now)
    restate.ObjectSend(ctx, "CronJob", jobID, "Execute").
        Send(restate.Void{}, restate.WithDelay(delay))

    return CronJobInfo{
        ID:             jobID,
        CronExpression: request.CronExpression,
        NextExecution:  nextExec,
        Status:         "active",
    }, nil
}

func (c *CronJob) Execute(ctx restate.ObjectContext) error {
    jobID := restate.Key(ctx)

    // Get job config
    config, err := restate.Get[CronJobRequest](ctx, "config")
    if err != nil {
        return err
    }

    status, _ := restate.Get[string](ctx, "status")
    if status != "active" {
        ctx.Log().Info("Job cancelled or inactive", "job_id", jobID)
        return nil
    }

    // Execute the job
    ctx.Log().Info("Executing cron job",
        "job_id", jobID,
        "service", config.Service,
        "method", config.Method)

    if config.Key != "" {
        // Call Virtual Object
        _, err = restate.Object[interface{}](
            ctx,
            config.Service,
            config.Key,
            config.Method,
        ).Request(config.Payload)
    } else {
        // Call Service
        _, err = restate.Service[interface{}](
            ctx,
            config.Service,
            config.Method,
        ).Request(config.Payload)
    }

    if err != nil {
        ctx.Log().Error("Job execution failed",
            "job_id", jobID,
            "error", err)
        // Don't return error - will retry and reschedule
    }

    // Update execution count and last execution time
    count, _ := restate.Get[int](ctx, "execution_count")
    restate.Set(ctx, "execution_count", count+1)
    restate.Set(ctx, "last_execution", time.Now())

    // Schedule next execution
    schedule, _ := restate.Run(ctx, func(ctx restate.RunContext) (*cron.Schedule, error) {
        return cron.ParseStandard(config.CronExpression)
    })

    now := time.Now()
    nextExec := schedule.Next(now)
    delay := nextExec.Sub(now)

    restate.Set(ctx, "next_execution", nextExec)

    // Schedule self for next execution
    restate.ObjectSend(ctx, "CronJob", jobID, "Execute").
        Send(restate.Void{}, restate.WithDelay(delay))

    return nil
}

func (c *CronJob) GetInfo(ctx restate.ObjectSharedContext) (CronJobInfo, error) {
    config, _ := restate.Get[CronJobRequest](ctx, "config")
    status, _ := restate.Get[string](ctx, "status")
    nextExec, _ := restate.Get[time.Time](ctx, "next_execution")
    lastExec, _ := restate.Get[time.Time](ctx, "last_execution")
    count, _ := restate.Get[int](ctx, "execution_count")

    return CronJobInfo{
        ID:             restate.Key(ctx),
        CronExpression: config.CronExpression,
        NextExecution:  nextExec,
        LastExecution:  lastExec,
        Status:         status,
        ExecutionCount: count,
    }, nil
}

func (c *CronJob) Cancel(ctx restate.ObjectContext) error {
    restate.Set(ctx, "status", "cancelled")
    ctx.Log().Info("Cron job cancelled", "job_id", restate.Key(ctx))
    return nil
}
```

## Common Cron Patterns

### Daily Task (Midnight)

```go
cronJobID, err := restateClient.Service[string](
    "CronJobInitiator",
    "Create",
).Request(ctx, CronJobRequest{
    CronExpression: "0 0 * * *",        // Daily at midnight
    Service:        "ReportService",
    Method:         "GenerateDailyReport",
})
```

### Hourly Task

```go
cronJobID, err := restateClient.Service[string](
    "CronJobInitiator",
    "Create",
).Request(ctx, CronJobRequest{
    CronExpression: "0 * * * *",        // Every hour
    Service:        "DataService",
    Method:         "SyncData",
    Payload:        SyncConfig{Source: "external-api"},
})
```

### Every 15 Minutes

```go
cronJobID, err := restateClient.Service[string](
    "CronJobInitiator",
    "Create",
).Request(ctx, CronJobRequest{
    CronExpression: "*/15 * * * *",     // Every 15 minutes
    Service:        "MonitorService",
    Method:         "CheckHealth",
})
```

### Weekly Task (Monday 9 AM)

```go
cronJobID, err := restateClient.Service[string](
    "CronJobInitiator",
    "Create",
).Request(ctx, CronJobRequest{
    CronExpression: "0 9 * * 1",        // Monday at 9 AM
    Service:        "ReportService",
    Method:         "GenerateWeeklyReport",
})
```

## Job Management

### Query Job Status

```go
info, err := restateClient.Object[CronJobInfo](
    "CronJob",
    jobID,
    "GetInfo",
).Request(ctx, nil)

fmt.Printf("Job: %s\n", info.ID)
fmt.Printf("Next execution: %s\n", info.NextExecution)
fmt.Printf("Executions: %d\n", info.ExecutionCount)
```

### Cancel Job

```go
err := restateClient.Object[void](
    "CronJob",
    jobID,
    "Cancel",
).Request(ctx, nil)
```

### Update Job (Cancel and Recreate)

```go
// Cancel old job
restateClient.Object[void]("CronJob", oldJobID, "Cancel").Request(ctx, nil)

// Create new job
newJobID, _ := restateClient.Service[string](
    "CronJobInitiator",
    "Create",
).Request(ctx, newCronJobRequest)
```

## Advanced Patterns

### Job with Timeout

```go
func (c *CronJob) Execute(ctx restate.ObjectContext) error {
    config, _ := restate.Get[CronJobRequest](ctx, "config")

    // Execute with timeout
    timeout := restate.After(ctx, 5*time.Minute)
    jobFuture := restate.Service[Result](
        ctx,
        config.Service,
        config.Method,
    ).RequestFuture(config.Payload)

    selected, err := restate.WaitFirst(ctx, timeout, jobFuture)
    if selected.Index() == 0 {
        ctx.Log().Error("Job execution timeout", "job_id", restate.Key(ctx))
        // Still reschedule for next run
    }

    // Reschedule...
    return nil
}
```

### Job with Retry Logic

```go
func (c *CronJob) Execute(ctx restate.ObjectContext) error {
    config, _ := restate.Get[CronJobRequest](ctx, "config")

    maxRetries := 3
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        if config.Key != "" {
            _, lastErr = restate.Object[interface{}](
                ctx,
                config.Service,
                config.Key,
                config.Method,
            ).Request(config.Payload)
        } else {
            _, lastErr = restate.Service[interface{}](
                ctx,
                config.Service,
                config.Method,
            ).Request(config.Payload)
        }

        if lastErr == nil {
            break
        }

        if attempt < maxRetries-1 {
            // Exponential backoff
            restate.Sleep(ctx, time.Duration(1<<uint(attempt))*time.Second)
        }
    }

    if lastErr != nil {
        ctx.Log().Error("Job failed after retries",
            "job_id", restate.Key(ctx),
            "attempts", maxRetries,
            "error", lastErr)
    }

    // Reschedule next execution regardless of success/failure
    // ...
    return nil
}
```

### Conditional Execution

```go
func (c *CronJob) Execute(ctx restate.ObjectContext) error {
    // Check if job should run
    shouldRun, _ := restate.Service[bool](
        ctx,
        "ConfigService",
        "IsJobEnabled",
    ).Request(restate.Key(ctx))

    if !shouldRun {
        ctx.Log().Info("Job skipped (disabled)", "job_id", restate.Key(ctx))
        // Still reschedule
        c.scheduleNext(ctx)
        return nil
    }

    // Execute job...
    return nil
}
```

## Best Practices

1. **Handle Failures Gracefully**: Always reschedule even if execution fails
2. **Use Idempotency Keys**: For job executions that shouldn't duplicate
3. **Monitor Execution**: Track execution count and last execution time
4. **Set Timeouts**: Prevent jobs from running indefinitely
5. **Log Everything**: Aid debugging and monitoring
6. **Use Shared Handlers**: For querying job status without blocking
7. **Plan for Holidays**: Consider business calendar integration
8. **Timezone Awareness**: Use UTC or handle timezone conversions

## References

- Official Docs: https://docs.restate.dev/guides/cron
- Durable Timers: See restate-go-durable-timers skill
- Virtual Objects: See restate-go-services skill
- Service Communication: See restate-go-service-communication skill
