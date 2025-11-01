# Restate Deep Dive: Building Distributed Scanning Workflows

## Overview

Restate is a **durable execution engine** that makes building reliable distributed systems simple. For Spectra-Red, Restate handles the complex orchestration of security scans across multiple targets while guaranteeing:

- **Exactly-once execution** - No duplicate processing
- **Automatic recovery** - Resume from failures without replay
- **State persistence** - Track scan progress, threat correlation
- **Service coordination** - Orchestrate multiple scanning services

## Core Concepts for Security Scanning

### 1. Durable Execution Model

Every operation is journaled. On failure, Restate replays completed steps and resumes from where execution stopped.

```
┌──────────────────────────────────────┐
│ Scan 100 Targets with Restate        │
├──────────────────────────────────────┤
│                                      │
│ Step 1: Validate targets ✓           │  → Journal: "validation_complete"
│ Step 2: Split into batches ✓         │  → Journal: "batches_created"
│ Step 3: Scan target 1-10 (CRASH)    │  → Journal: "batch1_done"
│         ↓ Process dies               │
│         ↓ Restart                    │
│ (Restate replays steps 1-3)          │
│ (Resumes from step 4)                │
│ Step 4: Scan target 11-20 ✓         │  → Journal: "batch2_done"
│ Step 5: Correlate results ✓         │  → Journal: "correlation_done"
│                                      │
└──────────────────────────────────────┘
```

**Key Insight**: Application developers don't write retry/recovery logic. Restate handles it automatically.

### 2. Three Service Types

#### Type A: Basic Services (Stateless)

Use for: One-off operations, scanning individual targets, ETL

```go
type ScanService struct{}

func (s *ScanService) ScanTarget(
    ctx restate.Context,
    target string,
) (ScanResult, error) {
    // No state - purely functional
    result, _ := restate.Run(ctx, func(ctx restate.RunContext) (ScanResult, error) {
        return runNmap(target)
    })
    return result, nil
}
```

#### Type B: Virtual Objects (Stateful Per-Key)

Use for: Threat records, scan sessions, contributor reputation

```go
// Per-threat state and history
type ThreatIntelligence struct{}

func (t *ThreatIntelligence) RecordThreat(
    ctx restate.ObjectContext,
    data ThreatData,
) error {
    threatID := restate.Key(ctx)  // e.g., "threat:CVE-2025-1234"
    
    // Read existing threat
    existing, _ := restate.Get[Threat](ctx, "record")
    
    // Update with new data
    existing.Updates = append(existing.Updates, data)
    existing.LastSeen = time.Now()
    
    // Write back
    restate.Set(ctx, "record", existing)
    
    return nil
}

// Concurrent read with shared context
func (t *ThreatIntelligence) GetThreat(
    ctx restate.ObjectSharedContext,
) (Threat, error) {
    record, _ := restate.Get[Threat](ctx, "record")
    return record, nil
}
```

**Key Guarantee**: Single-writer per key. Multiple concurrent reads. No deadlocks.

#### Type C: Workflows (Multi-Step Orchestration)

Use for: Complex scanning campaigns, threat correlation, approval workflows

```go
type SecurityScanWorkflow struct{}

func (w *SecurityScanWorkflow) Run(
    ctx restate.WorkflowContext,
    campaign ScanCampaign,
) (CampaignReport, error) {
    // Step 1: Validate targets
    validated, _ := restate.Service[[]string](
        ctx, "Validator", "ValidateTargets",
    ).Request(campaign.Targets)
    
    // Step 2: Run scans in parallel
    scanFutures := make([]restate.Future[ScanResult], 0)
    for _, target := range validated {
        future := restate.Service[ScanResult](
            ctx, "Scanner", "Scan",
        ).RequestFuture(target)
        scanFutures = append(scanFutures, future)
    }
    
    // Wait for all scans
    var results []ScanResult
    for _, future := range scanFutures {
        result, _ := future.Done()
        results = append(results, result)
    }
    
    // Step 3: Correlate threats
    report, _ := restate.Service[CampaignReport](
        ctx, "Correlator", "CorrelateThreats",
    ).Request(results)
    
    return report, nil
}
```

## Best Practices for Security Scanning with Restate

### Practice 1: Always Wrap External Operations

```go
// GOOD - Wrapped for durability
_, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
    return callExternalAPI()  // Only executes once, result cached
})

// BAD - Unprotected - may execute multiple times
result := callExternalAPI()
```

### Practice 2: Use Deterministic Operations Only

```go
// GOOD - Deterministic
id := restate.UUID(ctx)
rand := restate.Rand(ctx).Intn(100)

// BAD - Non-deterministic, breaks execution model
id := uuid.New()
rand := rand.Intn(100)
```

### Practice 3: Idempotency Keys for Critical Operations

```go
// For operations that must not be duplicated
paymentID, _ := restate.Service[string](ctx, "PaymentService", "Charge").
    Request(payment, restate.WithIdempotencyKey("scan-"+scanID))

// If request fails and retries, same key = same result, never double-charged
```

### Practice 4: Proper Error Handling

```go
// Transient error - Restate auto-retries
result, err := restate.Run(ctx, func(ctx restate.RunContext) (Data, error) {
    return fetchFromAPI()  // Network timeout = transient
})
if err != nil {
    return err  // Restate retries automatically
}

// Terminal error - stop retrying
if invalidInput {
    return restate.TerminalError(fmt.Errorf("invalid"), 400)
}
```

### Practice 5: Leverage Parallelism

```go
// DON'T: Serial requests
r1 := service1(input)
r2 := service2(input)
r3 := service3(input)  // Waits for r2 to complete

// DO: Parallel with futures
f1 := restate.Service[R1](ctx, "S1", "Op").RequestFuture(input)
f2 := restate.Service[R2](ctx, "S2", "Op").RequestFuture(input)
f3 := restate.Service[R3](ctx, "S3", "Op").RequestFuture(input)

r1, _ := f1.Done()
r2, _ := f2.Done()
r3, _ := f3.Done()  // All run concurrently
```

## Common Scanning Workflow Patterns

### Pattern 1: Batch Scanning with Progress Tracking

```go
type ScanCampaignWorkflow struct{}

func (w *ScanCampaignWorkflow) Run(
    ctx restate.WorkflowContext,
    campaign ScanCampaign,
) (Report, error) {
    // Store status for monitoring
    restate.Set(ctx, "status", "initializing")
    restate.Set(ctx, "progress", Progress{Scanned: 0, Total: len(campaign.Targets)})
    
    // Split into batches for resource management
    batchSize := 50
    batches := chunkTargets(campaign.Targets, batchSize)
    
    var allResults []ScanResult
    
    for i, batch := range batches {
        // Update progress
        restate.Set(ctx, "status", fmt.Sprintf("scanning_batch_%d", i+1))
        
        // Execute batch in parallel
        var futures []restate.Future[ScanResult]
        for _, target := range batch {
            f := restate.Service[ScanResult](ctx, "Scanner", "Scan").
                RequestFuture(target)
            futures = append(futures, f)
        }
        
        // Collect results
        for _, f := range futures {
            result, _ := f.Done()
            allResults = append(allResults, result)
        }
        
        // Update progress
        progress := Progress{
            Scanned: (i + 1) * batchSize,
            Total:   len(campaign.Targets),
        }
        restate.Set(ctx, "progress", progress)
    }
    
    // Finalize
    restate.Set(ctx, "status", "correlating")
    report, _ := restate.Service[Report](ctx, "Correlator", "Correlate").
        Request(allResults)
    
    restate.Set(ctx, "status", "complete")
    return report, nil
}
```

### Pattern 2: Saga Pattern (Distributed Transaction)

```go
func (w *WorkflowType) Run(ctx restate.WorkflowContext, order Order) error {
    var compensations []func()
    
    // Step 1: Reserve inventory
    reservationID, err := restate.Service[string](
        ctx, "Inventory", "Reserve",
    ).Request(order.Items)
    if err == nil {
        compensations = append(compensations, func() {
            restate.Service[void](ctx, "Inventory", "Release").
                Request(reservationID)
        })
    }
    
    // Step 2: Charge payment
    paymentID, err := restate.Service[string](
        ctx, "Payment", "Charge",
    ).Request(order.Payment)
    if err != nil {
        // Execute compensations in reverse order
        for i := len(compensations) - 1; i >= 0; i-- {
            compensations[i]()
        }
        return err
    }
    
    compensations = append(compensations, func() {
        restate.Service[void](ctx, "Payment", "Refund").
            Request(paymentID)
    })
    
    // Step 3: Ship
    shipmentID, err := restate.Service[string](
        ctx, "Shipping", "Ship",
    ).Request(order)
    if err != nil {
        // Undo previous steps
        for i := len(compensations) - 1; i >= 0; i-- {
            compensations[i]()
        }
        return err
    }
    
    return nil
}
```

### Pattern 3: Human Approval with Timeout

```go
func (w *ApprovalWorkflow) Run(
    ctx restate.WorkflowContext,
    claim Claim,
) (bool, error) {
    // Create awaitable for external approval
    awakeable := restate.Awakeable[ApprovalResult](ctx)
    
    // Send approval request
    restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        sendApprovalEmail(claim.Approver, awakeable.Id())
        return restate.Void{}, nil
    })
    
    // Wait for approval with 24-hour timeout
    timeout := restate.After(ctx, 24*time.Hour)
    approvalFuture := awakeable.Future()
    
    selected, _ := restate.WaitFirst(ctx, timeout, approvalFuture)
    
    if selected.Index() == 0 {
        // Timeout occurred
        return false, nil
    }
    
    // Got approval
    approval, _ := approvalFuture.Done()
    return approval.Approved, nil
}
```

## Performance Optimization

### 1. Partition Large Workflows

```go
// DON'T: Scan 10,000 targets serially
targets := getTargets()
for _, t := range targets {
    scan(t)
}

// DO: Process in batches with parallelism
targets := getTargets()
batchSize := 100
batches := partition(targets, batchSize)

for _, batch := range batches {
    futures := make([]Future, len(batch))
    for i, target := range batch {
        futures[i] = restate.Service[Result](
            ctx, "Scanner", "Scan",
        ).RequestFuture(target)
    }
    for _, f := range futures {
        f.Done()
    }
}
```

### 2. Use Virtual Objects for Hot Paths

```go
// DON'T: Store all threat updates in workflow state
func (w *ScanWorkflow) Run(ctx restate.WorkflowContext, ...) {
    threats, _ := restate.Get[[]Threat](ctx, "threats")
    threats = append(threats, newThreat)  // Entire list rewritten
    restate.Set(ctx, "threats", threats)
}

// DO: Use Virtual Objects for individual threats
func (t *ThreatIntel) RecordThreat(ctx restate.ObjectContext, threat Threat) {
    threatID := restate.Key(ctx)
    existing, _ := restate.Get[Threat](ctx, "data")
    existing.Updates = append(existing.Updates, threat)
    restate.Set(ctx, "data", existing)
}
```

### 3. Cache Expensive Computations

```go
func (s *Service) ExpensiveOperation(
    ctx restate.Context,
    input string,
) (Result, error) {
    // Check cache first
    cached, _ := restate.Get[Result](ctx, "cache:"+input)
    if cached != nil {
        return cached, nil
    }
    
    // Compute if not cached
    result, _ := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
        return doExpensiveWork(input)
    })
    
    // Cache for future requests
    restate.Set(ctx, "cache:"+input, result)
    
    return result, nil
}
```

## Testing Restate Applications

### Unit Testing

```go
func TestScanWorkflow(t *testing.T) {
    ctx := context.Background()
    workflow := &ScanCampaignWorkflow{}
    
    campaign := ScanCampaign{
        Targets: []string{"192.168.1.1", "192.168.1.2"},
    }
    
    // Run workflow
    report, err := workflow.Run(ctx, campaign)
    
    assert.NoError(t, err)
    assert.NotNil(t, report)
    assert.Len(t, report.Results, 2)
}
```

### Integration Testing with Testcontainers

```go
func TestWithRestateServer(t *testing.T) {
    ctx := context.Background()
    
    // Start Restate server
    container, _ := testcontainers.GenericContainer(ctx, ...)
    defer container.Terminate(ctx)
    
    // Connect and test
    client := restate.NewClient(...)
    
    result, _ := client.Service("Scanner", "Scan").Request(target)
    assert.NotNil(t, result)
}
```

## Monitoring & Observability

### Key Metrics to Track

1. **Workflow execution time** - Time from start to completion
2. **Failure rate** - % of workflows that fail
3. **Retry frequency** - How often steps are retried
4. **Journal size** - Growth of execution journals
5. **State store usage** - Space used by Virtual Objects

### Logging Best Practices

```go
func (s *Service) ScanTarget(ctx restate.Context, target string) error {
    logger := getLogger()
    
    logger.Info("scan_started", map[string]string{
        "target":    target,
        "request_id": restate.RequestID(ctx),
    })
    
    result, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
        return scan(target)
    })
    
    if err != nil {
        logger.Error("scan_failed", map[string]string{
            "target": target,
            "error":  err.Error(),
        })
    } else {
        logger.Info("scan_completed", map[string]interface{}{
            "target":  target,
            "services_found": len(result.Services),
        })
    }
    
    return err
}
```

## Common Pitfalls & Solutions

### Pitfall 1: Non-deterministic Code

```go
// WRONG: time.Now() differs on replay
timestamp := time.Now()

// RIGHT: Wrap in restate.Run()
timestamp, _ := restate.Run(ctx, func(ctx restate.RunContext) (time.Time, error) {
    return time.Now(), nil
})
```

### Pitfall 2: Goroutines Escaping Execution Model

```go
// WRONG: Goroutine escapes Restate's tracking
go externalAPI.Call()

// RIGHT: Use RunAsync for background execution
future := restate.RunAsync(ctx, func(ctx restate.RunContext) (Result, error) {
    return process()
})
result, _ := future.Done()
```

### Pitfall 3: Virtual Object Blocking

```go
// WRONG: Long sleep blocks all operations on this key
func (v *VirtualObject) Process(ctx restate.ObjectContext) {
    restate.Sleep(ctx, 24*time.Hour)  // All other requests wait!
}

// RIGHT: Use delayed messages
restate.ObjectSend(ctx, "Service", key, "Process").
    Send(input, restate.WithDelay(24*time.Hour))
```

### Pitfall 4: Unbounded State Growth

```go
// WRONG: State keeps growing forever
func (v *VirtualObject) AddEvent(ctx restate.ObjectContext, event Event) {
    events, _ := restate.Get[[]Event](ctx, "events")
    events = append(events, event)  // Keeps growing
    restate.Set(ctx, "events", events)
}

// RIGHT: Archive old events
func (v *VirtualObject) AddEvent(ctx restate.ObjectContext, event Event) {
    events, _ := restate.Get[[]Event](ctx, "events")
    events = append(events, event)
    
    // Keep only last 1000 events
    if len(events) > 1000 {
        archive(events[:len(events)-1000])
        events = events[len(events)-1000:]
    }
    
    restate.Set(ctx, "events", events)
}
```

## Deployment & Operations

### Docker Deployment

```dockerfile
FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o scanner .

CMD ["./scanner", "--restate-server=restate:8080"]
```

### Docker Compose

```yaml
version: '3.8'
services:
  restate:
    image: restatedev/restate:latest
    ports:
      - "8080:8080"
  
  scanner:
    build: .
    depends_on:
      - restate
    environment:
      RESTATE_SERVER: http://restate:8080
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: scanner-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: scanner
  template:
    metadata:
      labels:
        app: scanner
    spec:
      containers:
      - name: scanner
        image: scanner:latest
        env:
        - name: RESTATE_SERVER
          value: "http://restate:8080"
```

---

**Key Takeaway**: Restate abstracts away distributed systems complexity. Developers write simple, synchronous code. Restate handles durability, retries, and recovery automatically.

