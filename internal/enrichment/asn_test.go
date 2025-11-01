package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTeamCymruClient(t *testing.T) {
	tests := []struct {
		name      string
		rateLimit int
		cacheTTL  time.Duration
		wantRate  int
		wantTTL   time.Duration
	}{
		{
			name:      "default values",
			rateLimit: 0,
			cacheTTL:  0,
			wantRate:  100,
			wantTTL:   24 * time.Hour,
		},
		{
			name:      "custom values",
			rateLimit: 50,
			cacheTTL:  1 * time.Hour,
			wantRate:  50,
			wantTTL:   1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewTeamCymruClient(tt.rateLimit, tt.cacheTTL)
			assert.NotNil(t, client)
			assert.Equal(t, tt.wantTTL, client.cacheTTL)
			assert.Equal(t, tt.wantRate, client.rateLimit.maxTokens)
		})
	}
}

func TestTeamCymruClient_parseTeamCymruResponse(t *testing.T) {
	client := NewTeamCymruClient(100, 24*time.Hour)

	tests := []struct {
		name    string
		line    string
		want    *ASNInfo
		wantErr bool
	}{
		{
			name: "valid Google DNS response",
			line: "15169 | 8.8.8.8 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US",
			want: &ASNInfo{
				Number:  15169,
				Org:     "GOOGLE, US",
				Country: "US",
			},
			wantErr: false,
		},
		{
			name: "valid Cloudflare response",
			line: "13335 | 1.1.1.1 | 1.1.1.0/24 | US | arin | 2011-04-15 | CLOUDFLARENET, US",
			want: &ASNInfo{
				Number:  13335,
				Org:     "CLOUDFLARENET, US",
				Country: "US",
			},
			wantErr: false,
		},
		{
			name:    "invalid format - too few fields",
			line:    "15169 | 8.8.8.8",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid ASN number",
			line:    "invalid | 8.8.8.8 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.parseTeamCymruResponse(tt.line)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Number, got.Number)
			assert.Equal(t, tt.want.Org, got.Org)
			assert.Equal(t, tt.want.Country, got.Country)
		})
	}
}

func TestTeamCymruClient_Cache(t *testing.T) {
	client := NewTeamCymruClient(100, 100*time.Millisecond)

	// Set cache entry
	ip := "8.8.8.8"
	info := &ASNInfo{
		Number:  15169,
		Org:     "GOOGLE, US",
		Country: "US",
	}

	client.setCache(ip, info)

	// Check cache hit
	cached := client.checkCache(ip)
	require.NotNil(t, cached)
	assert.Equal(t, info.Number, cached.Number)
	assert.Equal(t, info.Org, cached.Org)
	assert.Equal(t, info.Country, cached.Country)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Check cache miss (expired)
	cached = client.checkCache(ip)
	assert.Nil(t, cached)
}

func TestTeamCymruClient_GetCacheStats(t *testing.T) {
	client := NewTeamCymruClient(100, 1*time.Hour)

	// Empty cache
	size, _ := client.GetCacheStats()
	assert.Equal(t, 0, size)

	// Add some entries
	client.setCache("8.8.8.8", &ASNInfo{Number: 15169, Org: "GOOGLE", Country: "US"})
	client.setCache("1.1.1.1", &ASNInfo{Number: 13335, Org: "CLOUDFLARE", Country: "US"})

	size, oldestEntry := client.GetCacheStats()
	assert.Equal(t, 2, size)
	assert.True(t, time.Since(oldestEntry) < 1*time.Second)
}

func TestTeamCymruClient_ClearExpiredCache(t *testing.T) {
	client := NewTeamCymruClient(100, 50*time.Millisecond)

	// Add entries
	client.setCache("8.8.8.8", &ASNInfo{Number: 15169, Org: "GOOGLE", Country: "US"})
	time.Sleep(30 * time.Millisecond)
	client.setCache("1.1.1.1", &ASNInfo{Number: 13335, Org: "CLOUDFLARE", Country: "US"})

	// Wait for first entry to expire
	time.Sleep(40 * time.Millisecond)

	// Clear expired entries
	removed := client.ClearExpiredCache()
	assert.Equal(t, 1, removed)

	// Check that only one entry remains
	size, _ := client.GetCacheStats()
	assert.Equal(t, 1, size)
}

func TestRateLimiter_Wait(t *testing.T) {
	// Create a rate limiter with 2 tokens, refilling every 100ms
	rl := &rateLimiter{
		tokens:     2,
		maxTokens:  2,
		refillRate: 100 * time.Millisecond,
		lastRefill: time.Now(),
	}

	ctx := context.Background()

	// Should succeed immediately (token available)
	start := time.Now()
	err := rl.wait(ctx)
	assert.NoError(t, err)
	assert.Less(t, time.Since(start), 10*time.Millisecond)

	// Second call should also succeed (1 token left)
	err = rl.wait(ctx)
	assert.NoError(t, err)

	// Third call should wait for refill
	start = time.Now()
	err = rl.wait(ctx)
	assert.NoError(t, err)
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond, "should wait for token refill")
	assert.Less(t, elapsed, 150*time.Millisecond, "should not wait too long")
}

func TestRateLimiter_WaitWithContext(t *testing.T) {
	// Create a rate limiter with 0 tokens
	rl := &rateLimiter{
		tokens:     0,
		maxTokens:  1,
		refillRate: 1 * time.Second, // Long refill time
		lastRefill: time.Now(),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should fail with context deadline exceeded
	start := time.Now()
	err := rl.wait(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	assert.Less(t, elapsed, 200*time.Millisecond)
}

// Integration test - only runs if INTEGRATION_TEST env var is set
func TestTeamCymruClient_LookupASN_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewTeamCymruClient(10, 1*time.Hour)
	ctx := context.Background()

	tests := []struct {
		name    string
		ip      string
		wantASN int
	}{
		{
			name:    "Google DNS",
			ip:      "8.8.8.8",
			wantASN: 15169,
		},
		{
			name:    "Cloudflare DNS",
			ip:      "1.1.1.1",
			wantASN: 13335,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.LookupASN(ctx, tt.ip)
			if err != nil {
				t.Skipf("skipping test due to network error: %v", err)
			}

			require.NotNil(t, info)
			assert.Equal(t, tt.wantASN, info.Number)
			assert.NotEmpty(t, info.Org)
			assert.NotEmpty(t, info.Country)
		})
	}
}

// Integration test for batch lookup
func TestTeamCymruClient_LookupBatch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewTeamCymruClient(10, 1*time.Hour)
	ctx := context.Background()

	ips := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}

	results, err := client.LookupBatch(ctx, ips)
	if err != nil {
		t.Skipf("skipping test due to network error: %v", err)
	}

	require.NotNil(t, results)
	assert.GreaterOrEqual(t, len(results), 2, "should have results for at least 2 IPs")

	// Check specific results
	if info, ok := results["8.8.8.8"]; ok {
		assert.Equal(t, 15169, info.Number)
	}

	if info, ok := results["1.1.1.1"]; ok {
		assert.Equal(t, 13335, info.Number)
	}
}

// Benchmark tests
func BenchmarkTeamCymruClient_parseTeamCymruResponse(b *testing.B) {
	client := NewTeamCymruClient(100, 24*time.Hour)
	line := "15169 | 8.8.8.8 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.parseTeamCymruResponse(line)
	}
}

func BenchmarkTeamCymruClient_checkCache(b *testing.B) {
	client := NewTeamCymruClient(100, 24*time.Hour)
	client.setCache("8.8.8.8", &ASNInfo{Number: 15169, Org: "GOOGLE", Country: "US"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.checkCache("8.8.8.8")
	}
}
