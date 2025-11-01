---
name: restate-rate-limiting
description: Guide for implementing rate limiting with Restate using token bucket algorithm and Virtual Objects. Use when implementing API rate limits, throttling, or request quotas.
---

# Restate Rate Limiting

Implement reliable rate limiting using Virtual Objects and the token bucket algorithm.

## Overview

Restate enables rate limiting through:
- **Virtual Objects**: Isolated rate limiters per key
- **Durable State**: Reliable token tracking
- **Durable Timers**: Token refill scheduling

## Token Bucket Algorithm

**Parameters**:
- **Limit**: Tokens per second
- **Burst**: Maximum tokens available
- **Tokens**: Current available tokens

## Go Implementation

### Rate Limiter Virtual Object

```go
type RateLimiter struct{}

type LimiterConfig struct {
    Limit float64  // tokens per second
    Burst int      // max tokens
}

type LimiterState struct {
    Tokens      float64
    LastRefill  time.Time
}

func (r *RateLimiter) SetRate(ctx restate.ObjectContext, config LimiterConfig) error {
    restate.Set(ctx, "config", config)
    restate.Set(ctx, "state", LimiterState{
        Tokens:     float64(config.Burst),
        LastRefill: time.Now(),
    })
    return nil
}

func (r *RateLimiter) Wait(ctx restate.ObjectContext, tokens int) error {
    config, _ := restate.Get[LimiterConfig](ctx, "config")

    if tokens > config.Burst {
        return restate.TerminalError(
            fmt.Errorf("requested %d exceeds burst %d", tokens, config.Burst),
            429,
        )
    }

    for {
        state := r.advanceTokens(ctx)

        if state.Tokens >= float64(tokens) {
            // Consume tokens
            state.Tokens -= float64(tokens)
            restate.Set(ctx, "state", state)
            return nil
        }

        // Calculate wait time
        needed := float64(tokens) - state.Tokens
        waitTime := time.Duration(needed/config.Limit*1000) * time.Millisecond

        // Durable sleep
        restate.Sleep(ctx, waitTime)
    }
}

func (r *RateLimiter) Allow(ctx restate.ObjectContext, tokens int) (bool, error) {
    state := r.advanceTokens(ctx)

    if state.Tokens >= float64(tokens) {
        state.Tokens -= float64(tokens)
        restate.Set(ctx, "state", state)
        return true, nil
    }

    return false, nil
}

func (r *RateLimiter) advanceTokens(ctx restate.ObjectContext) LimiterState {
    config, _ := restate.Get[LimiterConfig](ctx, "config")
    state, _ := restate.Get[LimiterState](ctx, "state")

    now := time.Now()
    elapsed := now.Sub(state.LastRefill).Seconds()

    // Add tokens based on elapsed time
    newTokens := state.Tokens + elapsed*config.Limit
    if newTokens > float64(config.Burst) {
        newTokens = float64(config.Burst)
    }

    state.Tokens = newTokens
    state.LastRefill = now

    restate.Set(ctx, "state", state)
    return state
}
```

### Client Usage

```go
// Initialize rate limiter
restateClient.Object[void]("RateLimiter", "api-user-123", "SetRate").
    Request(ctx, LimiterConfig{
        Limit: 10.0,  // 10 requests per second
        Burst: 20,    // Allow bursts up to 20
    })

// Wait for permission
err := restateClient.Object[void]("RateLimiter", "api-user-123", "Wait").
    Request(ctx, 1)

if err != nil {
    // Rate limit exceeded
    return http.StatusTooManyRequests
}

// Allow without waiting
allowed, _ := restateClient.Object[bool]("RateLimiter", "api-user-123", "Allow").
    Request(ctx, 1)

if !allowed {
    return http.StatusTooManyRequests
}
```

## Use Cases

**Per-User API Rate Limiting**: Key by user ID
**Per-Endpoint Throttling**: Key by endpoint path
**Global Service Limits**: Single key for service-wide limits
**Tiered Rate Limits**: Different limits based on subscription tier

## References

- Official Docs: https://docs.restate.dev/guides/rate-limiting
- Virtual Objects: See restate-go-services skill
- Durable Timers: See restate-go-durable-timers skill
