// +build integration

package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestASNClient_BatchProcessing tests batch processing of multiple IPs
func TestASNClient_BatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create client with reasonable limits
	client := NewTeamCymruClient(20, 1*time.Hour)
	ctx := context.Background()

	// Test with a batch of 20 IPs (within rate limit)
	ips := []string{
		"8.8.8.8",      // Google
		"1.1.1.1",      // Cloudflare
		"9.9.9.9",      // Quad9
		"208.67.222.222", // OpenDNS
		"76.76.2.0",    // Comcast
		"8.8.4.4",      // Google Secondary
		"1.0.0.1",      // Cloudflare Secondary
		"149.112.112.112", // Quad9 Secondary
		"208.67.220.220", // OpenDNS Secondary
		"76.76.19.19",  // Comcast Secondary
	}

	// Measure time for batch lookup
	start := time.Now()
	results, err := client.LookupBatch(ctx, ips)
	elapsed := time.Since(start)

	require.NoError(t, err, "batch lookup should succeed")
	assert.GreaterOrEqual(t, len(results), 8, "should get results for most IPs")

	t.Logf("Batch lookup of %d IPs completed in %v", len(ips), elapsed)
	t.Logf("Successfully enriched %d IPs", len(results))

	// Verify specific known ASNs
	if info, ok := results["8.8.8.8"]; ok {
		assert.Equal(t, 15169, info.Number, "Google ASN")
		assert.Contains(t, info.Org, "GOOGLE")
	}

	if info, ok := results["1.1.1.1"]; ok {
		assert.Equal(t, 13335, info.Number, "Cloudflare ASN")
		assert.Contains(t, info.Org, "CLOUDFLARE")
	}
}

// TestASNClient_Caching tests cache hit and miss scenarios
func TestASNClient_Caching(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create client with 5-second cache TTL for testing
	client := NewTeamCymruClient(20, 5*time.Second)
	ctx := context.Background()

	ip := "8.8.8.8"

	// First lookup - cache miss (will hit external API)
	start := time.Now()
	info1, err := client.LookupASN(ctx, ip)
	firstLookupTime := time.Since(start)
	require.NoError(t, err)
	require.NotNil(t, info1)

	t.Logf("First lookup (cache miss) took: %v", firstLookupTime)

	// Second lookup - cache hit (should be instant)
	start = time.Now()
	info2, err := client.LookupASN(ctx, ip)
	secondLookupTime := time.Since(start)
	require.NoError(t, err)
	require.NotNil(t, info2)

	t.Logf("Second lookup (cache hit) took: %v", secondLookupTime)

	// Cache hit should be much faster
	assert.Less(t, secondLookupTime, firstLookupTime/10, "cached lookup should be at least 10x faster")

	// Verify same data returned
	assert.Equal(t, info1.Number, info2.Number)
	assert.Equal(t, info1.Org, info2.Org)
	assert.Equal(t, info1.Country, info2.Country)

	// Wait for cache to expire
	t.Log("Waiting for cache to expire...")
	time.Sleep(6 * time.Second)

	// Third lookup - cache miss again (should hit API)
	start = time.Now()
	info3, err := client.LookupASN(ctx, ip)
	thirdLookupTime := time.Since(start)
	require.NoError(t, err)
	require.NotNil(t, info3)

	t.Logf("Third lookup (cache expired) took: %v", thirdLookupTime)

	// Should be slower than cached lookup
	assert.Greater(t, thirdLookupTime, secondLookupTime*5, "expired cache lookup should be slower")

	// Verify cache stats
	size, oldestEntry := client.GetCacheStats()
	assert.Equal(t, 1, size, "cache should have 1 entry")
	assert.True(t, time.Since(oldestEntry) < 2*time.Second, "entry should be fresh")
}

// TestASNClient_RateLimiting tests rate limiting behavior
func TestASNClient_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create client with strict rate limit: 5 req/min
	client := NewTeamCymruClient(5, 1*time.Hour)
	ctx := context.Background()

	// Try to make 10 lookups (should be rate limited)
	ips := []string{
		"8.8.8.8",
		"8.8.4.4",
		"1.1.1.1",
		"1.0.0.1",
		"9.9.9.9",
		"149.112.112.112",
		"208.67.222.222",
		"208.67.220.220",
		"76.76.2.0",
		"76.76.19.19",
	}

	start := time.Now()

	// Make lookups sequentially (rate limiter should slow them down)
	successCount := 0
	for _, ip := range ips {
		_, err := client.LookupASN(ctx, ip)
		if err == nil {
			successCount++
		}
	}

	elapsed := time.Since(start)

	t.Logf("Completed %d lookups in %v", successCount, elapsed)

	// With 5 req/min rate limit, 10 requests should take at least 1 minute
	// (5 requests immediate, then 5 more at 12 seconds each = 60 seconds total)
	expectedMinTime := 60 * time.Second
	assert.GreaterOrEqual(t, elapsed, expectedMinTime, "rate limiting should enforce delay")

	assert.Equal(t, 10, successCount, "all requests should eventually succeed")
}

// TestASNClient_CacheCleanup tests cache cleanup functionality
func TestASNClient_CacheCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create client with 100ms cache TTL
	client := NewTeamCymruClient(20, 100*time.Millisecond)
	ctx := context.Background()

	// Add several entries
	ips := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}

	for _, ip := range ips {
		_, err := client.LookupASN(ctx, ip)
		require.NoError(t, err)
	}

	// Verify all entries are cached
	size, _ := client.GetCacheStats()
	assert.Equal(t, len(ips), size)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Clean expired entries
	removed := client.ClearExpiredCache()
	assert.Equal(t, len(ips), removed, "should remove all expired entries")

	// Verify cache is empty
	size, _ = client.GetCacheStats()
	assert.Equal(t, 0, size, "cache should be empty")
}

// TestASNClient_ConcurrentAccess tests thread-safe concurrent access
func TestASNClient_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewTeamCymruClient(50, 1*time.Hour)
	ctx := context.Background()

	// Test concurrent lookups
	ips := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9", "208.67.222.222", "76.76.2.0"}

	// Make concurrent lookups (should be thread-safe)
	done := make(chan bool, len(ips))
	errors := make(chan error, len(ips))

	for _, ip := range ips {
		go func(ip string) {
			_, err := client.LookupASN(ctx, ip)
			if err != nil {
				errors <- err
			}
			done <- true
		}(ip)
	}

	// Wait for all to complete
	for i := 0; i < len(ips); i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errList []error
	for err := range errors {
		errList = append(errList, err)
	}

	assert.Empty(t, errList, "concurrent access should not cause errors")

	// Verify cache is populated
	size, _ := client.GetCacheStats()
	assert.GreaterOrEqual(t, size, 3, "cache should have multiple entries")
}

// TestASNClient_BatchVsSingleLookup compares batch vs individual lookups
func TestASNClient_BatchVsSingleLookup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	ips := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9", "208.67.222.222", "76.76.2.0"}

	// Test individual lookups
	client1 := NewTeamCymruClient(50, 1*time.Hour)
	start := time.Now()
	for _, ip := range ips {
		_, _ = client1.LookupASN(ctx, ip)
	}
	individualTime := time.Since(start)

	t.Logf("Individual lookups took: %v", individualTime)

	// Test batch lookup
	client2 := NewTeamCymruClient(50, 1*time.Hour)
	start = time.Now()
	results, err := client2.LookupBatch(ctx, ips)
	batchTime := time.Since(start)

	require.NoError(t, err)
	t.Logf("Batch lookup took: %v", batchTime)
	t.Logf("Enriched %d/%d IPs", len(results), len(ips))

	// Batch should be faster or comparable
	// (Note: Team Cymru batch queries are not necessarily faster per-IP,
	//  but they're more efficient for network round-trips)
	t.Logf("Speedup: %.2fx", float64(individualTime)/float64(batchTime))
}
