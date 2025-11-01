# Go Patterns Reference for Spectra-Red

This document provides battle-tested Go patterns for building high-throughput security scanning systems.

## 1. Worker Pool Pattern

Perfect for parallel scanning of thousands of targets.

```go
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
    wg      sync.WaitGroup
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    pool := &WorkerPool{
        workers: numWorkers,
        jobs:    make(chan Job, numWorkers*2),
        results: make(chan Result, numWorkers*2),
    }
    
    for i := 0; i < numWorkers; i++ {
        pool.wg.Add(1)
        go pool.worker()
    }
    
    return pool
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for job := range p.jobs {
        result := processJob(job)
        p.results <- result
    }
}

func (p *WorkerPool) Submit(job Job) {
    p.jobs <- job
}

func (p *WorkerPool) Close() {
    close(p.jobs)
    p.wg.Wait()
}

// Usage
pool := NewWorkerPool(16)
for _, target := range targets {
    pool.Submit(ScanJob{Target: target})
}

results := make([]Result, 0)
pool.Close()
for result := range pool.results {
    results = append(results, result)
}
```

## 2. Rate Limiter (Token Bucket)

Prevent overwhelming targets or hitting API rate limits.

```go
type RateLimiter struct {
    capacity      float64
    current       float64
    refillPerSec  float64
    lastRefillTime time.Time
    mu            sync.Mutex
}

func NewRateLimiter(capacity, refillPerSec float64) *RateLimiter {
    return &RateLimiter{
        capacity:       capacity,
        current:        capacity,
        refillPerSec:   refillPerSec,
        lastRefillTime: time.Now(),
    }
}

func (rl *RateLimiter) WaitAndTake(tokens float64) {
    rl.mu.Lock()
    
    // Refill based on elapsed time
    now := time.Now()
    elapsed := now.Sub(rl.lastRefillTime).Seconds()
    refilled := elapsed * rl.refillPerSec
    rl.current = math.Min(rl.current+refilled, rl.capacity)
    rl.lastRefillTime = now
    
    // Wait if not enough tokens
    for rl.current < tokens {
        rl.mu.Unlock()
        
        waitTime := time.Duration((tokens-rl.current)/rl.refillPerSec*1000) * time.Millisecond
        time.Sleep(waitTime)
        
        rl.mu.Lock()
        
        // Refill again
        now := time.Now()
        elapsed := now.Sub(rl.lastRefillTime).Seconds()
        refilled := elapsed * rl.refillPerSec
        rl.current = math.Min(rl.current+refilled, rl.capacity)
        rl.lastRefillTime = now
    }
    
    // Take tokens
    rl.current -= tokens
    rl.mu.Unlock()
}

// Usage
limiter := NewRateLimiter(100, 10)  // 100 token capacity, 10 tokens/sec refill
limiter.WaitAndTake(1)              // Wait for 1 token
makeRequest()                        // Now safe to make request
```

## 3. Circuit Breaker Pattern

Handle failing external services gracefully.

```go
type CircuitBreaker struct {
    name              string
    maxFailures       int
    resetTimeout      time.Duration
    failureCount      int
    lastFailureTime   time.Time
    state             CircuitState
    mu                sync.RWMutex
}

type CircuitState string

const (
    CLOSED     CircuitState = "CLOSED"
    OPEN       CircuitState = "OPEN"
    HALF_OPEN CircuitState = "HALF_OPEN"
)

func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        name:         name,
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        CLOSED,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Try to transition from OPEN to HALF_OPEN
    if cb.state == OPEN {
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.state = HALF_OPEN
            cb.failureCount = 0
        } else {
            return fmt.Errorf("%s circuit breaker is open", cb.name)
        }
    }
    
    cb.mu.Unlock()
    
    // Execute the function
    err := fn()
    
    cb.mu.Lock()
    
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.maxFailures {
            cb.state = OPEN
        } else if cb.state == HALF_OPEN {
            cb.state = OPEN
        }
        
        return err
    }
    
    // Success
    if cb.state == HALF_OPEN {
        cb.state = CLOSED
        cb.failureCount = 0
    }
    
    return nil
}

// Usage
breaker := NewCircuitBreaker("externalAPI", 5, 30*time.Second)
err := breaker.Call(func() error {
    return externalAPI.Call()
})
```

## 4. Batch Processing

Improve throughput by batching operations.

```go
type BatchProcessor struct {
    batchSize     int
    flushInterval time.Duration
    items         []interface{}
    mu            sync.Mutex
    done          chan struct{}
}

func NewBatchProcessor(batchSize int, flushInterval time.Duration) *BatchProcessor {
    bp := &BatchProcessor{
        batchSize:     batchSize,
        flushInterval: flushInterval,
        items:         make([]interface{}, 0, batchSize),
        done:          make(chan struct{}),
    }
    
    go bp.periodicFlush()
    return bp
}

func (bp *BatchProcessor) Add(item interface{}) {
    bp.mu.Lock()
    bp.items = append(bp.items, item)
    
    if len(bp.items) >= bp.batchSize {
        batch := make([]interface{}, len(bp.items))
        copy(batch, bp.items)
        bp.items = bp.items[:0]
        bp.mu.Unlock()
        
        go bp.processBatch(batch)
    } else {
        bp.mu.Unlock()
    }
}

func (bp *BatchProcessor) periodicFlush() {
    ticker := time.NewTicker(bp.flushInterval)
    for range ticker.C {
        bp.mu.Lock()
        if len(bp.items) > 0 {
            batch := make([]interface{}, len(bp.items))
            copy(batch, bp.items)
            bp.items = bp.items[:0]
            bp.mu.Unlock()
            
            go bp.processBatch(batch)
        } else {
            bp.mu.Unlock()
        }
    }
}

func (bp *BatchProcessor) processBatch(batch []interface{}) {
    // Batch insert into database
    db.BulkInsert(batch)
}

// Usage
processor := NewBatchProcessor(1000, 5*time.Second)
for _, item := range items {
    processor.Add(item)
}
```

## 5. Retry with Exponential Backoff

Handle transient failures in external calls.

```go
type RetryConfig struct {
    MaxRetries       int
    InitialBackoff   time.Duration
    MaxBackoff       time.Duration
    BackoffMultiplier float64
}

func (r *RetryConfig) Do(ctx context.Context, operation func() error) error {
    backoff := r.InitialBackoff
    var lastErr error
    
    for attempt := 0; attempt <= r.MaxRetries; attempt++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        lastErr = operation()
        if lastErr == nil {
            return nil
        }
        
        if attempt < r.MaxRetries {
            select {
            case <-time.After(backoff):
                backoff = time.Duration(float64(backoff) * r.BackoffMultiplier)
                if backoff > r.MaxBackoff {
                    backoff = r.MaxBackoff
                }
            case <-ctx.Done():
                return ctx.Err()
            }
        }
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", r.MaxRetries+1, lastErr)
}

// Usage
config := &RetryConfig{
    MaxRetries:        3,
    InitialBackoff:    100 * time.Millisecond,
    MaxBackoff:        10 * time.Second,
    BackoffMultiplier: 2.0,
}

err := config.Do(ctx, func() error {
    return externalAPI.Call()
})
```

## 6. Concurrent Map (Thread-Safe Cache)

Store results safely across goroutines.

```go
type ConcurrentMap struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func NewConcurrentMap() *ConcurrentMap {
    return &ConcurrentMap{
        data: make(map[string]interface{}),
    }
}

func (cm *ConcurrentMap) Set(key string, value interface{}) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    cm.data[key] = value
}

func (cm *ConcurrentMap) Get(key string) (interface{}, bool) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    val, ok := cm.data[key]
    return val, ok
}

func (cm *ConcurrentMap) Delete(key string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    delete(cm.data, key)
}

func (cm *ConcurrentMap) Range(f func(string, interface{}) bool) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    for k, v := range cm.data {
        if !f(k, v) {
            break
        }
    }
}

// Or use sync.Map (built-in)
var cache sync.Map

cache.Store("key", "value")
value, ok := cache.Load("key")
cache.Delete("key")
```

## 7. Fan-Out/Fan-In Pattern

Distribute work and collect results.

```go
func FanOutFanIn(ctx context.Context, inputs []string, numWorkers int) ([]Result, error) {
    // Create channels
    jobs := make(chan string, len(inputs))
    results := make(chan Result, len(inputs))
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                result := processJob(ctx, job)
                results <- result
            }
        }()
    }
    
    // Send jobs
    go func() {
        for _, input := range inputs {
            select {
            case jobs <- input:
            case <-ctx.Done():
                break
            }
        }
        close(jobs)
    }()
    
    // Close results when workers finish
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var collected []Result
    for result := range results {
        collected = append(collected, result)
    }
    
    return collected, nil
}
```

## 8. Structured Logging

Use structured logging for production debugging.

```go
type Logger struct {
    logger *slog.Logger
}

func (l *Logger) Info(msg string, attrs map[string]interface{}) {
    var logAttrs []slog.Attr
    for k, v := range attrs {
        logAttrs = append(logAttrs, slog.Any(k, v))
    }
    l.logger.InfoContext(context.Background(), msg, logAttrs...)
}

func (l *Logger) Error(msg string, err error, attrs map[string]interface{}) {
    attrs["error"] = err.Error()
    var logAttrs []slog.Attr
    for k, v := range attrs {
        logAttrs = append(logAttrs, slog.Any(k, v))
    }
    l.logger.Error(msg, logAttrs...)
}

// Usage
logger.Info("scan_started", map[string]interface{}{
    "target": target,
    "timestamp": time.Now(),
})
```

## 9. Context Timeouts

Prevent hanging requests.

```go
// In HTTP handler
ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
defer cancel()

// Pass context to processing
result, err := processor.Process(ctx, input)
if err == context.DeadlineExceeded {
    http.Error(w, "request timeout", http.StatusRequestTimeout)
    return
}
```

## 10. Health Checks

Make services observable.

```go
type HealthChecker struct {
    checks map[string]func() error
}

func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        checks: make(map[string]func() error),
    }
}

func (hc *HealthChecker) Register(name string, check func() error) {
    hc.checks[name] = check
}

func (hc *HealthChecker) Check() map[string]string {
    results := make(map[string]string)
    
    for name, check := range hc.checks {
        if err := check(); err != nil {
            results[name] = "unhealthy: " + err.Error()
        } else {
            results[name] = "healthy"
        }
    }
    
    return results
}

// Usage
hc := NewHealthChecker()
hc.Register("database", func() error {
    return db.Ping()
})
hc.Register("restate", func() error {
    return restateClient.Health()
})

http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    results := hc.Check()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
})
```

## Best Practices

### 1. Always Use Context
```go
// GOOD
func Process(ctx context.Context, input string) error {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Pass context to sub-operations
    return subProcess(ctx, input)
}

// BAD
func Process(input string) error {
    // No timeout, no cancellation
    return subProcess(input)
}
```

### 2. Defer Cleanup
```go
// GOOD
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

// Process file
```

### 3. Use Sync Primitives Correctly
```go
// Use RWMutex for read-heavy workloads
var mu sync.RWMutex
var data map[string]string

// Many readers
mu.RLock()
val := data[key]
mu.RUnlock()

// Few writers
mu.Lock()
data[key] = value
mu.Unlock()
```

### 4. Avoid Goroutine Leaks
```go
// GOOD: Using context to cancel goroutines
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    select {
    case <-ctx.Done():
        return
    case <-time.After(10 * time.Second):
        doWork()
    }
}()

// BAD: Goroutine leaks because it never exits
go func() {
    for {
        doWork()
    }
}()
```

### 5. Table-Driven Tests
```go
func TestNormalizeThreat(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Threat
        wantErr bool
    }{
        {"valid", "CVE-2025-1234", Threat{CVE: "CVE-2025-1234"}, false},
        {"invalid", "garbage", Threat{}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Normalize(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, want %v", err != nil, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

**Key Principle**: Write concurrent, resilient code using proven Go patterns. Always consider context, timeouts, and error handling.

