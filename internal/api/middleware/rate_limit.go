package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	tokens         float64
	capacity       float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
	mu             sync.Mutex
}

// NewTokenBucket creates a new token bucket with the given capacity and refill rate
func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:         capacity,
		capacity:       capacity,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request can proceed and consumes a token if so
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens += elapsed * tb.refillRate

	// Cap at maximum capacity
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefillTime = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// RateLimiter manages rate limits per scanner key
type RateLimiter struct {
	buckets  map[string]*TokenBucket
	capacity float64
	rate     float64 // refill rate in tokens per second
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewRateLimiter creates a new rate limiter
// requestsPerMinute: maximum requests allowed per minute
func NewRateLimiter(requestsPerMinute int, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*TokenBucket),
		capacity: float64(requestsPerMinute),
		rate:     float64(requestsPerMinute) / 60.0, // convert to tokens per second
		logger:   logger,
	}
}

// Allow checks if a request from the given key can proceed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = NewTokenBucket(rl.capacity, rl.rate)
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	return bucket.Allow()
}

// CleanupStale removes buckets that haven't been used recently (memory optimization)
func (rl *RateLimiter) CleanupStale(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if now.Sub(bucket.lastRefillTime) > maxAge {
			delete(rl.buckets, key)
		}
		bucket.mu.Unlock()
	}
}

// RateLimitMiddleware creates a middleware that enforces rate limiting per scanner
func RateLimitMiddleware(limiter *RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract scanner key from request
			// For now, we use the public key from the request body
			// In a production system, this might come from a header after auth
			scannerKey := extractScannerKey(r)

			if scannerKey == "" {
				// If we can't extract a key, allow the request
				// The auth layer will reject it if invalid
				next.ServeHTTP(w, r)
				return
			}

			if !limiter.Allow(scannerKey) {
				limiter.logger.Warn("rate limit exceeded",
					zap.String("scanner_key", maskKey(scannerKey)),
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr))

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", "60")
				w.Header().Set("X-RateLimit-Window", "1m")
				w.WriteHeader(http.StatusTooManyRequests)

				response := map[string]interface{}{
					"error":     "rate_limit_exceeded",
					"message":   "Rate limit exceeded. Maximum 60 requests per minute per scanner.",
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				}
				_ = json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractScannerKey extracts a unique identifier for rate limiting
// For the ingest endpoint, we use the client IP as a basic identifier
// In production, this would be enhanced to use the authenticated scanner ID
func extractScannerKey(r *http.Request) string {
	// Use X-Forwarded-For if behind a proxy, otherwise use RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	return r.RemoteAddr
}

// maskKey masks a key for safe logging
func maskKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:8] + "..."
}

// StartCleanupRoutine starts a background goroutine to clean up stale buckets
func (rl *RateLimiter) StartCleanupRoutine(interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.CleanupStale(maxAge)
		}
	}()
}
