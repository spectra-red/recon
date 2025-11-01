package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestTokenBucket_Allow(t *testing.T) {
	// Create a bucket with 5 tokens, refilling at 1 token/second
	bucket := NewTokenBucket(5, 1)

	// Should allow 5 requests immediately
	for i := 0; i < 5; i++ {
		assert.True(t, bucket.Allow(), "request %d should be allowed", i+1)
	}

	// 6th request should be denied
	assert.False(t, bucket.Allow(), "6th request should be denied")

	// Wait for 1 second to refill 1 token
	time.Sleep(1100 * time.Millisecond) // Add buffer for timing

	// Should allow 1 more request
	assert.True(t, bucket.Allow(), "should allow after refill")

	// Should deny again
	assert.False(t, bucket.Allow(), "should deny after consuming refilled token")
}

func TestTokenBucket_Refill(t *testing.T) {
	// Create a bucket with 10 tokens, refilling at 10 tokens/second
	bucket := NewTokenBucket(10, 10)

	// Consume all tokens
	for i := 0; i < 10; i++ {
		bucket.Allow()
	}

	// Should be empty
	assert.False(t, bucket.Allow())

	// Wait for 500ms (should refill ~5 tokens)
	time.Sleep(600 * time.Millisecond)

	// Should allow ~5 requests
	allowed := 0
	for i := 0; i < 10; i++ {
		if bucket.Allow() {
			allowed++
		}
	}

	assert.GreaterOrEqual(t, allowed, 4, "should allow at least 4 requests after refill")
	assert.LessOrEqual(t, allowed, 7, "should not allow more than 7 requests")
}

func TestTokenBucket_Capacity(t *testing.T) {
	// Create a bucket with capacity 3
	bucket := NewTokenBucket(3, 100)

	// Wait to ensure bucket is full
	time.Sleep(100 * time.Millisecond)

	// Should still only allow 3 requests (not more than capacity)
	for i := 0; i < 3; i++ {
		assert.True(t, bucket.Allow())
	}
	assert.False(t, bucket.Allow())
}

func TestRateLimiter_Allow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(60, logger) // 60 requests per minute

	key1 := "scanner-001"
	key2 := "scanner-002"

	// Different keys should have independent limits
	assert.True(t, limiter.Allow(key1))
	assert.True(t, limiter.Allow(key2))

	// Same key should share the bucket
	for i := 0; i < 59; i++ {
		assert.True(t, limiter.Allow(key1), "request %d for key1", i+2)
	}

	// 61st request for key1 should be denied
	assert.False(t, limiter.Allow(key1))

	// key2 should still have capacity
	assert.True(t, limiter.Allow(key2))
}

func TestRateLimiter_CleanupStale(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(60, logger)

	// Use several keys
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		limiter.Allow(key)
	}

	assert.Len(t, limiter.buckets, 3)

	// Manually set lastRefillTime to simulate staleness
	for _, bucket := range limiter.buckets {
		bucket.mu.Lock()
		bucket.lastRefillTime = time.Now().Add(-2 * time.Hour)
		bucket.mu.Unlock()
	}

	// Cleanup buckets older than 1 hour
	limiter.CleanupStale(1 * time.Hour)

	assert.Len(t, limiter.buckets, 0, "all stale buckets should be removed")
}

func TestRateLimitMiddleware_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(60, logger)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

func TestRateLimitMiddleware_Exceeded(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(5, logger) // Low limit for testing
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Make 5 requests from same IP
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "60", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "1m", w.Header().Get("X-RateLimit-Window"))

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "rate_limit_exceeded", response["error"])
	assert.Contains(t, response["message"], "Rate limit exceeded")
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(5, logger)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Different IPs should have independent limits
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}

	for _, ip := range ips {
		// Each IP should be able to make 5 requests
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
			req.RemoteAddr = ip + ":12345"
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "IP %s request %d", ip, i+1)
		}

		// 6th request should be denied for each IP
		req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
		req.RemoteAddr = ip + ":12345"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "IP %s should be rate limited", ip)
	}
}

func TestRateLimitMiddleware_XForwardedFor(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(5, logger)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Test with X-Forwarded-For header (simulates proxy)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 6th request with same X-Forwarded-For should be limited
	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestExtractScannerKey(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		expectedPrefix string
	}{
		{
			name:           "direct connection",
			remoteAddr:     "192.168.1.1:12345",
			xForwardedFor:  "",
			expectedPrefix: "192.168.1.1",
		},
		{
			name:           "behind proxy",
			remoteAddr:     "127.0.0.1:12345",
			xForwardedFor:  "203.0.113.42",
			expectedPrefix: "203.0.113.42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			key := extractScannerKey(req)
			assert.Contains(t, key, tt.expectedPrefix)
		})
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "long key",
			key:      "192.168.1.1:12345",
			expected: "192.168.",
		},
		{
			name:     "short key",
			key:      "short",
			expected: "short",
		},
		{
			name:     "exactly 8 chars",
			key:      "12345678",
			expected: "12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskKey(tt.key)
			if len(tt.key) > 8 {
				assert.Contains(t, result, "...")
				assert.Contains(t, result, tt.expected)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRateLimiter_StartCleanupRoutine(t *testing.T) {
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiter(60, logger)

	// Add some buckets
	for i := 0; i < 5; i++ {
		limiter.Allow(string(rune('A' + i)))
	}

	assert.Len(t, limiter.buckets, 5)

	// Set buckets to be stale
	for _, bucket := range limiter.buckets {
		bucket.mu.Lock()
		bucket.lastRefillTime = time.Now().Add(-2 * time.Hour)
		bucket.mu.Unlock()
	}

	// Start cleanup with short intervals for testing
	limiter.StartCleanupRoutine(100*time.Millisecond, 1*time.Hour)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	assert.Len(t, limiter.buckets, 0, "cleanup should remove stale buckets")
}

// Benchmark tests
func BenchmarkTokenBucket_Allow(b *testing.B) {
	bucket := NewTokenBucket(1000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	logger := zaptest.NewLogger(b)
	limiter := NewRateLimiter(60, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("test-key")
	}
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	logger := zaptest.NewLogger(b)
	limiter := NewRateLimiter(10000, logger)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
	}
}
