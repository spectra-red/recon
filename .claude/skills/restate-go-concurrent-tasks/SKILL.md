---
name: restate-go-concurrent-tasks
description: Guide for Restate Go concurrent task execution including parallel operations, racing futures, and combinators. Use when implementing parallel processing, racing operations, or handling multiple async tasks.
---

# Restate Go Concurrent Tasks

Execute multiple durable operations in parallel while maintaining deterministic replay behavior.

## Core Purpose

Restate enables concurrent execution of durable operations with:
- **Deterministic replay**: Completion order logged for consistency during failures
- **Fault tolerance**: Completed tasks replay with cached results; pending tasks retry
- **Parallel execution**: Multiple operations run simultaneously
- **Type safety**: Full Go generics support

## Key Use Cases

- Simultaneous calls to multiple external services
- Racing operations to use the first successful result
- Timeout implementation via operation racing
- Parallel batch processing
- Fan-out/fan-in patterns

## Execution Patterns

### Parallel Execution with RunAsync

Start durable operations without blocking:

```go
func (s *Service) ParallelFetch(ctx restate.Context) ([]Data, error) {
    // Start three operations in parallel
    future1 := restate.RunAsync(ctx, func(ctx restate.RunContext) (Data, error) {
        return fetchFromSource1()
    })

    future2 := restate.RunAsync(ctx, func(ctx restate.RunContext) (Data, error) {
        return fetchFromSource2()
    })

    future3 := restate.RunAsync(ctx, func(ctx restate.RunContext) (Data, error) {
        return fetchFromSource3()
    })

    // Wait for all to complete
    data1, err1 := future1.Done()
    data2, err2 := future2.Done()
    data3, err3 := future3.Done()

    if err1 != nil || err2 != nil || err3 != nil {
        return nil, fmt.Errorf("one or more fetches failed")
    }

    return []Data{data1, data2, data3}, nil
}
```

### Parallel Service Calls

```go
func (s *Service) AggregateData(ctx restate.Context, userID string) (Report, error) {
    // Start multiple service calls concurrently
    profileFuture := restate.Service[Profile](ctx, "UserService", "GetProfile").
        RequestFuture(userID)

    ordersFuture := restate.Service[[]Order](ctx, "OrderService", "GetUserOrders").
        RequestFuture(userID)

    preferencesFuture := restate.Service[Preferences](ctx, "PreferenceService", "Get").
        RequestFuture(userID)

    // Wait for all results
    profile, err1 := profileFuture.Done()
    orders, err2 := ordersFuture.Done()
    preferences, err3 := preferencesFuture.Done()

    if err1 != nil || err2 != nil || err3 != nil {
        return Report{}, fmt.Errorf("failed to fetch user data")
    }

    return createReport(profile, orders, preferences), nil
}
```

## Race Pattern: WaitFirst

Select the first successful completion from multiple operations.

### Basic Race

```go
func (s *Service) FetchFromFastest(ctx restate.Context) (Data, error) {
    // Start multiple redundant operations
    future1 := restate.Service[Data](ctx, "DataService1", "Fetch").RequestFuture(query)
    future2 := restate.Service[Data](ctx, "DataService2", "Fetch").RequestFuture(query)
    future3 := restate.Service[Data](ctx, "DataService3", "Fetch").RequestFuture(query)

    // Wait for first to complete
    selected, err := restate.WaitFirst(ctx, future1, future2, future3)
    if err != nil {
        return Data{}, err
    }

    switch selected.Index() {
    case 0:
        return future1.Done()
    case 1:
        return future2.Done()
    case 2:
        return future3.Done()
    }

    return Data{}, fmt.Errorf("no future completed")
}
```

### Race with Timeout

```go
func (s *Service) CallWithTimeout(ctx restate.Context, request Request) (Response, error) {
    // Create timeout
    timeoutFuture := restate.After(ctx, 30*time.Second)

    // Create operation
    callFuture := restate.Service[Response](ctx, "SlowService", "Process").
        RequestFuture(request)

    // Race operation against timeout
    selected, err := restate.WaitFirst(ctx, timeoutFuture, callFuture)
    if err != nil {
        return Response{}, err
    }

    switch selected.Index() {
    case 0:
        // Timeout won
        return Response{}, fmt.Errorf("operation timed out")
    case 1:
        // Operation completed
        return callFuture.Done()
    }

    return Response{}, nil
}
```

### Fallback Pattern

```go
func (s *Service) FetchWithFallback(ctx restate.Context, id string) (Data, error) {
    // Try primary source
    primaryFuture := restate.Service[Data](ctx, "PrimaryDB", "Get").RequestFuture(id)

    // Wait for primary
    primary, err := primaryFuture.Done()
    if err == nil {
        return primary, nil
    }

    ctx.Log().Warn("Primary fetch failed, trying fallback")

    // Try fallback source
    fallbackFuture := restate.Service[Data](ctx, "CacheService", "Get").RequestFuture(id)
    return fallbackFuture.Done()
}
```

## Wait for All Completions

Block until all futures complete.

### Using Done() on Each Future

```go
func (s *Service) WaitAll(ctx restate.Context, ids []string) ([]Data, error) {
    // Start all operations
    futures := make([]restate.Future[Data], len(ids))
    for i, id := range ids {
        futures[i] = restate.Service[Data](ctx, "DataService", "Get").RequestFuture(id)
    }

    // Collect all results
    results := make([]Data, len(futures))
    for i, future := range futures {
        data, err := future.Done()
        if err != nil {
            return nil, fmt.Errorf("failed to fetch data for id %s: %w", ids[i], err)
        }
        results[i] = data
    }

    return results, nil
}
```

### Using Wait Combinator

```go
func (s *Service) ParallelProcessing(ctx restate.Context, items []Item) error {
    // Start all processing operations
    futures := make([]restate.Future[restate.Void], len(items))
    for i, item := range items {
        futures[i] = restate.Service[restate.Void](ctx, "ProcessorService", "Process").
            RequestFuture(item)
    }

    // Wait for all to complete
    for _, future := range futures {
        _, err := future.Done()
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Common Patterns

### Fan-Out / Fan-In

```go
func (s *Service) ProcessBatch(ctx restate.Context, batch []Task) (Summary, error) {
    // Fan-out: Start all tasks in parallel
    futures := make([]restate.Future[Result], len(batch))
    for i, task := range batch {
        futures[i] = restate.Service[Result](ctx, "TaskProcessor", "Process").
            RequestFuture(task)
    }

    // Fan-in: Collect all results
    results := make([]Result, len(futures))
    var errors []error

    for i, future := range futures {
        result, err := future.Done()
        if err != nil {
            errors = append(errors, err)
            continue
        }
        results[i] = result
    }

    if len(errors) > 0 {
        return Summary{}, fmt.Errorf("some tasks failed: %v", errors)
    }

    return createSummary(results), nil
}
```

### Parallel with Partial Failures

```go
func (s *Service) BestEffortFetch(ctx restate.Context, sources []string) ([]Data, error) {
    futures := make([]restate.Future[Data], len(sources))
    for i, source := range sources {
        futures[i] = restate.Service[Data](ctx, "FetchService", "Get").
            RequestFuture(source)
    }

    // Collect successful results, ignore failures
    var results []Data
    for i, future := range futures {
        data, err := future.Done()
        if err != nil {
            ctx.Log().Warn("Failed to fetch from source",
                "source", sources[i],
                "error", err)
            continue
        }
        results = append(results, data)
    }

    if len(results) == 0 {
        return nil, fmt.Errorf("all fetches failed")
    }

    return results, nil
}
```

### Map-Reduce Pattern

```go
func (s *Service) MapReduce(ctx restate.Context, items []Item) (Result, error) {
    // Map phase: Process items in parallel
    futures := make([]restate.Future[Intermediate], len(items))
    for i, item := range items {
        futures[i] = restate.RunAsync(ctx, func(ctx restate.RunContext) (Intermediate, error) {
            return mapFunction(item)
        })
    }

    // Collect intermediate results
    intermediates := make([]Intermediate, len(futures))
    for i, future := range futures {
        result, err := future.Done()
        if err != nil {
            return Result{}, err
        }
        intermediates[i] = result
    }

    // Reduce phase
    finalResult, err := restate.Run(ctx, func(ctx restate.RunContext) (Result, error) {
        return reduceFunction(intermediates)
    })

    return finalResult, err
}
```

### Concurrent Writes to Different Objects

```go
func (s *Service) UpdateMultipleAccounts(
    ctx restate.Context,
    updates map[string]int,
) error {
    // Start updates in parallel
    futures := []restate.Future[restate.Void]{}

    for accountID, amount := range updates {
        future := restate.Object[restate.Void](
            ctx,
            "Account",
            accountID,
            "UpdateBalance",
        ).RequestFuture(amount)

        futures = append(futures, future)
    }

    // Wait for all updates
    for _, future := range futures {
        _, err := future.Done()
        if err != nil {
            return err
        }
    }

    return nil
}
```

### Racing Multiple Strategies

```go
func (s *Service) OptimizeQuery(ctx restate.Context, query Query) (Results, error) {
    // Try multiple query strategies simultaneously
    exactMatch := restate.Service[Results](ctx, "SearchService", "ExactMatch").
        RequestFuture(query)

    fuzzyMatch := restate.Service[Results](ctx, "SearchService", "FuzzyMatch").
        RequestFuture(query)

    semanticSearch := restate.Service[Results](ctx, "AIService", "SemanticSearch").
        RequestFuture(query)

    // Take first result that returns
    selected, err := restate.WaitFirst(ctx, exactMatch, fuzzyMatch, semanticSearch)
    if err != nil {
        return Results{}, err
    }

    switch selected.Index() {
    case 0:
        ctx.Log().Info("Using exact match results")
        return exactMatch.Done()
    case 1:
        ctx.Log().Info("Using fuzzy match results")
        return fuzzyMatch.Done()
    case 2:
        ctx.Log().Info("Using semantic search results")
        return semanticSearch.Done()
    }

    return Results{}, nil
}
```

### Concurrent Durable Steps

```go
func (s *Service) ParallelExternalCalls(ctx restate.Context) (Combined, error) {
    // Multiple HTTP calls in parallel
    future1 := restate.RunAsync(ctx, func(ctx restate.RunContext) (Response1, error) {
        return callExternalAPI1()
    })

    future2 := restate.RunAsync(ctx, func(ctx restate.RunContext) (Response2, error) {
        return callExternalAPI2()
    })

    future3 := restate.RunAsync(ctx, func(ctx restate.RunContext) (Response3, error) {
        return callDatabase()
    })

    // Wait for all
    resp1, err1 := future1.Done()
    resp2, err2 := future2.Done()
    resp3, err3 := future3.Done()

    if err1 != nil || err2 != nil || err3 != nil {
        return Combined{}, fmt.Errorf("one or more calls failed")
    }

    return combine(resp1, resp2, resp3), nil
}
```

## Critical Constraints

### DO NOT Use Goroutines

**Wrong**:
```go
// DON'T DO THIS - Not deterministic!
func (s *Service) BadParallel(ctx restate.Context) error {
    var wg sync.WaitGroup
    ch := make(chan Data)

    go func() { // ❌ Non-deterministic
        data := fetchData()
        ch <- data
    }()

    result := <-ch // ❌ Cannot replay deterministically
    return nil
}
```

**Correct**:
```go
// DO THIS - Deterministic with futures
func (s *Service) GoodParallel(ctx restate.Context) error {
    future := restate.RunAsync(ctx, func(ctx restate.RunContext) (Data, error) {
        return fetchData()
    })

    result, err := future.Done() // ✅ Deterministic
    return err
}
```

### Goroutines Only Inside Run

Goroutines are allowed ONLY inside `restate.Run()` blocks:

```go
func (s *Service) CorrectUsage(ctx restate.Context) error {
    result, err := restate.Run(ctx, func(ctx restate.RunContext) (Data, error) {
        // Inside Run, goroutines are OK
        var wg sync.WaitGroup
        ch := make(chan Item, 10)

        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(id int) {
                defer wg.Done()
                ch <- fetchItem(id)
            }(i)
        }

        wg.Wait()
        close(ch)

        return aggregateFromChannel(ch)
    })

    return err
}
```

## Best Practices

1. **Use Futures for Parallelism**: Never use goroutines outside `Run` blocks
2. **Handle Partial Failures**: Decide whether to fail-fast or continue
3. **Race with Timeouts**: Prevent unbounded waiting
4. **Log Completions**: Track which operations complete first for debugging
5. **Consider Ordering**: Futures complete in non-deterministic order (logged for replay)
6. **Type Safety**: Leverage generics for compile-time safety
7. **Error Aggregation**: Collect and report all errors, not just the first
8. **Limit Concurrency**: Don't start unlimited parallel operations

## Deterministic Replay Guarantee

When failures occur:
1. Previously completed operations replay from the log (instant)
2. Pending operations retry (actual execution)
3. Completion order is preserved across retries
4. Results remain consistent

## Performance Considerations

- Parallel operations reduce total latency
- Each concurrent operation uses resources
- Virtual Object calls to same key are serialized
- Consider backend capacity when parallelizing

## References

- Official Docs: https://docs.restate.dev/develop/go/concurrent-tasks
- Durable Steps: See restate-go-durable-steps skill
- Durable Timers: See restate-go-durable-timers skill
- Service Communication: See restate-go-service-communication skill
