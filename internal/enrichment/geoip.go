package enrichment

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPInfo represents geographic information for an IP address
type GeoIPInfo struct {
	IP        string  `json:"ip"`
	City      string  `json:"city"`
	Region    string  `json:"region"`
	Country   string  `json:"country"`
	CountryCC string  `json:"country_cc"` // ISO 3166-1 alpha-2
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GeoIPClient provides GeoIP lookup functionality with local MMDB files and API fallback
type GeoIPClient struct {
	mmdbPath   string
	db         *geoip2.Reader
	mu         sync.RWMutex
	httpClient *http.Client
	apiKey     string // Optional API key for fallback service
	apiURL     string // Optional API URL for fallback
}

// GeoIPConfig configures the GeoIP client
type GeoIPConfig struct {
	// Path to MaxMind GeoLite2 City MMDB file
	MMDBPath string

	// Optional API fallback configuration
	APIKey string // ipinfo.io API key
	APIURL string // Default: https://ipinfo.io
}

// NewGeoIPClient creates a new GeoIP lookup client
// Prioritizes local MMDB file reading over API calls for performance
func NewGeoIPClient(config GeoIPConfig) (*GeoIPClient, error) {
	client := &GeoIPClient{
		mmdbPath: config.MMDBPath,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		apiKey: config.APIKey,
		apiURL: config.APIURL,
	}

	// Set default API URL if not provided
	if client.apiURL == "" {
		client.apiURL = "https://ipinfo.io"
	}

	// Try to open MMDB file if path is provided
	if config.MMDBPath != "" {
		if err := client.openMMDB(); err != nil {
			// Log warning but don't fail - we can fall back to API
			return client, fmt.Errorf("warning: failed to open MMDB file (will use API fallback): %w", err)
		}
	}

	return client, nil
}

// openMMDB opens the MaxMind MMDB database file
func (c *GeoIPClient) openMMDB() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(c.mmdbPath); os.IsNotExist(err) {
		return fmt.Errorf("MMDB file not found: %s", c.mmdbPath)
	}

	// Open the database
	db, err := geoip2.Open(c.mmdbPath)
	if err != nil {
		return fmt.Errorf("failed to open MMDB file: %w", err)
	}

	c.db = db
	return nil
}

// Close closes the GeoIP client and releases resources
func (c *GeoIPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Lookup performs a GeoIP lookup for a single IP address
// Returns GeoIPInfo or error if lookup fails
func (c *GeoIPClient) Lookup(ipStr string) (*GeoIPInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Try MMDB lookup first (fast, no rate limits)
	c.mu.RLock()
	hasMMDB := c.db != nil
	c.mu.RUnlock()

	if hasMMDB {
		info, err := c.lookupMMDB(ip)
		if err == nil {
			return info, nil
		}
		// If MMDB lookup fails, fall through to API
	}

	// Fallback to API if MMDB is unavailable or lookup failed
	if c.apiKey != "" {
		return c.lookupAPI(ipStr)
	}

	return nil, fmt.Errorf("no GeoIP data source available (MMDB failed and no API key configured)")
}

// lookupMMDB performs a lookup using the local MMDB file
func (c *GeoIPClient) lookupMMDB(ip net.IP) (*GeoIPInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.db == nil {
		return nil, fmt.Errorf("MMDB database not initialized")
	}

	record, err := c.db.City(ip)
	if err != nil {
		return nil, fmt.Errorf("MMDB lookup failed: %w", err)
	}

	info := &GeoIPInfo{
		IP:        ip.String(),
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
		CountryCC: record.Country.IsoCode,
	}

	// Extract city name (prefer English)
	if len(record.City.Names) > 0 {
		if name, ok := record.City.Names["en"]; ok {
			info.City = name
		} else {
			// Fallback to first available name
			for _, name := range record.City.Names {
				info.City = name
				break
			}
		}
	}

	// Extract region name (prefer English)
	if len(record.Subdivisions) > 0 {
		if name, ok := record.Subdivisions[0].Names["en"]; ok {
			info.Region = name
		} else {
			// Fallback to first available name
			for _, name := range record.Subdivisions[0].Names {
				info.Region = name
				break
			}
		}
	}

	// Extract country name (prefer English)
	if len(record.Country.Names) > 0 {
		if name, ok := record.Country.Names["en"]; ok {
			info.Country = name
		} else {
			// Fallback to first available name
			for _, name := range record.Country.Names {
				info.Country = name
				break
			}
		}
	}

	return info, nil
}

// lookupAPI performs a lookup using ipinfo.io API
// This is a fallback when MMDB is unavailable
func (c *GeoIPClient) lookupAPI(ipStr string) (*GeoIPInfo, error) {
	// Note: This is a simplified implementation
	// In production, you would parse the JSON response from ipinfo.io
	// For now, return an error to encourage using MMDB
	return nil, fmt.Errorf("API fallback not fully implemented - please provide MMDB file")
}

// LookupBatch performs GeoIP lookups for multiple IP addresses
// Returns a map of IP -> GeoIPInfo
// Skips IPs that fail lookup without returning error
func (c *GeoIPClient) LookupBatch(ips []string) (map[string]*GeoIPInfo, error) {
	results := make(map[string]*GeoIPInfo)
	var mu sync.Mutex

	// Use a worker pool for batch processing
	const maxWorkers = 10
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, ipStr := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			info, err := c.Lookup(ip)
			if err == nil && info != nil {
				mu.Lock()
				results[ip] = info
				mu.Unlock()
			}
			// Silently skip failed lookups in batch mode
		}(ipStr)
	}

	wg.Wait()

	if len(results) == 0 {
		return nil, fmt.Errorf("no successful GeoIP lookups from %d IPs", len(ips))
	}

	return results, nil
}

// ValidateMMDB checks if the MMDB file is valid and readable
func ValidateMMDB(path string) error {
	db, err := geoip2.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open MMDB: %w", err)
	}
	defer db.Close()

	// Try a test lookup
	testIP := net.ParseIP("8.8.8.8")
	_, err = db.City(testIP)
	if err != nil {
		return fmt.Errorf("MMDB validation failed (test lookup): %w", err)
	}

	return nil
}
