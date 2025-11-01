package enrichment

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ASNInfo represents ASN information for an IP address
type ASNInfo struct {
	Number  int    `json:"asn"`
	Org     string `json:"org"`
	Country string `json:"country"`
}

// ASNClient provides ASN lookup capabilities
type ASNClient interface {
	LookupASN(ctx context.Context, ip string) (*ASNInfo, error)
	LookupBatch(ctx context.Context, ips []string) (map[string]*ASNInfo, error)
}

// TeamCymruClient implements ASN lookups via Team Cymru's whois service
// https://www.team-cymru.com/ip-asn-mapping
type TeamCymruClient struct {
	cache      map[string]*cacheEntry
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
	rateLimit  *rateLimiter
}

type cacheEntry struct {
	info      *ASNInfo
	timestamp time.Time
}

type rateLimiter struct {
	tokens    int
	maxTokens int
	refillRate time.Duration
	lastRefill time.Time
	mu        sync.Mutex
}

// NewTeamCymruClient creates a new ASN client using Team Cymru
// rateLimit: max requests per minute (default 100)
// cacheTTL: how long to cache results (default 24 hours)
func NewTeamCymruClient(rateLimit int, cacheTTL time.Duration) *TeamCymruClient {
	if rateLimit <= 0 {
		rateLimit = 100 // Default 100 req/min
	}
	if cacheTTL <= 0 {
		cacheTTL = 24 * time.Hour // Default 24h cache
	}

	return &TeamCymruClient{
		cache:    make(map[string]*cacheEntry),
		cacheTTL: cacheTTL,
		rateLimit: &rateLimiter{
			tokens:    rateLimit,
			maxTokens: rateLimit,
			refillRate: time.Minute / time.Duration(rateLimit), // Refill rate per token
			lastRefill: time.Now(),
		},
	}
}

// LookupASN performs an ASN lookup for a single IP address
func (c *TeamCymruClient) LookupASN(ctx context.Context, ip string) (*ASNInfo, error) {
	// Check cache first
	if info := c.checkCache(ip); info != nil {
		return info, nil
	}

	// Wait for rate limit token
	if err := c.rateLimit.wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Perform lookup
	info, err := c.lookupTeamCymru(ctx, ip)
	if err != nil {
		return nil, err
	}

	// Cache result
	c.setCache(ip, info)

	return info, nil
}

// LookupBatch performs ASN lookups for multiple IP addresses
// This is more efficient than calling LookupASN multiple times
func (c *TeamCymruClient) LookupBatch(ctx context.Context, ips []string) (map[string]*ASNInfo, error) {
	results := make(map[string]*ASNInfo)
	var missing []string

	// Check cache for all IPs
	for _, ip := range ips {
		if info := c.checkCache(ip); info != nil {
			results[ip] = info
		} else {
			missing = append(missing, ip)
		}
	}

	// If all IPs were cached, return early
	if len(missing) == 0 {
		return results, nil
	}

	// Process missing IPs in batches to respect rate limiting
	// Team Cymru supports bulk queries (up to 100 IPs per connection)
	batchSize := 50 // Conservative batch size
	for i := 0; i < len(missing); i += batchSize {
		end := i + batchSize
		if end > len(missing) {
			end = len(missing)
		}

		batch := missing[i:end]

		// Wait for rate limit token
		if err := c.rateLimit.wait(ctx); err != nil {
			return results, fmt.Errorf("rate limit wait failed: %w", err)
		}

		// Perform batch lookup
		batchResults, err := c.lookupTeamCymruBatch(ctx, batch)
		if err != nil {
			return results, fmt.Errorf("batch lookup failed: %w", err)
		}

		// Merge results and cache
		for ip, info := range batchResults {
			results[ip] = info
			c.setCache(ip, info)
		}
	}

	return results, nil
}

// lookupTeamCymru performs a single ASN lookup via Team Cymru whois
func (c *TeamCymruClient) lookupTeamCymru(ctx context.Context, ip string) (*ASNInfo, error) {
	// Connect to Team Cymru whois server
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", "whois.cymru.com:43")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Team Cymru: %w", err)
	}
	defer conn.Close()

	// Set read/write deadline
	deadline := time.Now().Add(10 * time.Second)
	conn.SetDeadline(deadline)

	// Send query: " -v <ip>" for verbose output
	query := fmt.Sprintf(" -v %s\n", ip)
	if _, err := conn.Write([]byte(query)); err != nil {
		return nil, fmt.Errorf("failed to write query: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	var responseLine string
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comment lines and headers
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "AS") || strings.TrimSpace(line) == "" {
			continue
		}
		responseLine = line
		break
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if responseLine == "" {
		return nil, fmt.Errorf("no ASN data found for IP %s", ip)
	}

	// Parse response
	// Format: ASN | IP | BGP Prefix | CC | Registry | Allocated | AS Name
	// Example: 15169 | 8.8.8.8 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US
	return c.parseTeamCymruResponse(responseLine)
}

// lookupTeamCymruBatch performs batch ASN lookup via Team Cymru
func (c *TeamCymruClient) lookupTeamCymruBatch(ctx context.Context, ips []string) (map[string]*ASNInfo, error) {
	results := make(map[string]*ASNInfo)

	// Connect to Team Cymru whois server
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", "whois.cymru.com:43")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Team Cymru: %w", err)
	}
	defer conn.Close()

	// Set deadline
	deadline := time.Now().Add(30 * time.Second)
	conn.SetDeadline(deadline)

	// Send "begin" marker
	if _, err := conn.Write([]byte("begin\n")); err != nil {
		return nil, fmt.Errorf("failed to write begin marker: %w", err)
	}

	// Send all IPs with -v flag for verbose
	for _, ip := range ips {
		if _, err := conn.Write([]byte(fmt.Sprintf(" -v %s\n", ip))); err != nil {
			return nil, fmt.Errorf("failed to write IP %s: %w", ip, err)
		}
	}

	// Send "end" marker
	if _, err := conn.Write([]byte("end\n")); err != nil {
		return nil, fmt.Errorf("failed to write end marker: %w", err)
	}

	// Read responses
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comment lines, headers, and empty lines
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "AS") || strings.TrimSpace(line) == "" {
			continue
		}

		// Parse response
		info, err := c.parseTeamCymruResponse(line)
		if err != nil {
			// Log error but continue processing other IPs
			continue
		}

		// Extract IP from the response line to map it to the result
		// Format: ASN | IP | ...
		fields := strings.Split(line, "|")
		if len(fields) >= 2 {
			ip := strings.TrimSpace(fields[1])
			results[ip] = info
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return results, nil
}

// parseTeamCymruResponse parses a Team Cymru response line
// Format: ASN | IP | BGP Prefix | CC | Registry | Allocated | AS Name
// Example: 15169 | 8.8.8.8 | 8.8.8.0/24 | US | arin | 1992-12-01 | GOOGLE, US
func (c *TeamCymruClient) parseTeamCymruResponse(line string) (*ASNInfo, error) {
	fields := strings.Split(line, "|")
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid response format: %s", line)
	}

	// Parse ASN
	asnStr := strings.TrimSpace(fields[0])
	asn, err := strconv.Atoi(asnStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ASN number: %s", asnStr)
	}

	// Extract country code
	country := strings.TrimSpace(fields[3])

	// Extract AS name (organization)
	org := strings.TrimSpace(fields[6])

	return &ASNInfo{
		Number:  asn,
		Org:     org,
		Country: country,
	}, nil
}

// checkCache checks if an IP is in the cache and not expired
func (c *TeamCymruClient) checkCache(ip string) *ASNInfo {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	entry, exists := c.cache[ip]
	if !exists {
		return nil
	}

	// Check if entry is expired
	if time.Since(entry.timestamp) > c.cacheTTL {
		return nil
	}

	return entry.info
}

// setCache stores an ASN info in the cache
func (c *TeamCymruClient) setCache(ip string, info *ASNInfo) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.cache[ip] = &cacheEntry{
		info:      info,
		timestamp: time.Now(),
	}
}

// GetCacheStats returns cache statistics
func (c *TeamCymruClient) GetCacheStats() (size int, oldestEntry time.Time) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	size = len(c.cache)
	oldestEntry = time.Now()

	for _, entry := range c.cache {
		if entry.timestamp.Before(oldestEntry) {
			oldestEntry = entry.timestamp
		}
	}

	return size, oldestEntry
}

// ClearExpiredCache removes expired entries from the cache
func (c *TeamCymruClient) ClearExpiredCache() int {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	now := time.Now()
	removed := 0

	for ip, entry := range c.cache {
		if now.Sub(entry.timestamp) > c.cacheTTL {
			delete(c.cache, ip)
			removed++
		}
	}

	return removed
}

// wait blocks until a rate limit token is available
func (rl *rateLimiter) wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for {
		// Refill tokens based on elapsed time
		now := time.Now()
		elapsed := now.Sub(rl.lastRefill)
		tokensToAdd := int(elapsed / rl.refillRate)

		if tokensToAdd > 0 {
			rl.tokens += tokensToAdd
			if rl.tokens > rl.maxTokens {
				rl.tokens = rl.maxTokens
			}
			rl.lastRefill = now
		}

		// If we have tokens, consume one and return
		if rl.tokens > 0 {
			rl.tokens--
			return nil
		}

		// Calculate wait time until next token
		waitTime := rl.refillRate - (now.Sub(rl.lastRefill) % rl.refillRate)

		// Release lock while waiting
		rl.mu.Unlock()

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			rl.mu.Lock()
			return ctx.Err()
		case <-time.After(waitTime):
			rl.mu.Lock()
			// Loop to refill and try again
		}
	}
}
