package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

const (
	// NVD API endpoint
	nvdBaseURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"

	// Rate limits (requests per 30 seconds)
	nvdRateLimitPublic  = 5  // Without API key
	nvdRateLimitWithKey = 50 // With API key

	// Request timeout
	nvdRequestTimeout = 30 * time.Second

	// Cache TTL
	nvdCacheTTL = 24 * time.Hour
)

// NVDClient provides methods for querying the NVD API
type NVDClient struct {
	httpClient *http.Client
	apiKey     string
	limiter    *rate.Limiter
	cache      *NVDCache
}

// NVDCache stores cached NVD responses
type NVDCache struct {
	entries map[string]*CacheEntry
}

// CacheEntry represents a cached NVD response
type CacheEntry struct {
	Data      []CVEItem
	ExpiresAt time.Time
}

// CVEItem represents a CVE from the NVD API
type CVEItem struct {
	CVEID       string    `json:"cve_id"`
	Description string    `json:"description"`
	CVSS        float64   `json:"cvss"`
	Severity    string    `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	Published   time.Time `json:"published"`
	Modified    time.Time `json:"modified"`
	CPEs        []string  `json:"cpes"`
	References  []string  `json:"references"`
}

// VulnMatch represents a vulnerability matched to a service
type VulnMatch struct {
	ServiceID string  `json:"service_id"`
	CVE       string  `json:"cve"`
	CVSS      float64 `json:"cvss"`
	Severity  string  `json:"severity"`
}

// NVDResponse represents the NVD API response structure
type NVDResponse struct {
	ResultsPerPage int `json:"resultsPerPage"`
	StartIndex     int `json:"startIndex"`
	TotalResults   int `json:"totalResults"`
	Vulnerabilities []struct {
		CVE struct {
			ID          string `json:"id"`
			Published   string `json:"published"`
			LastModified string `json:"lastModified"`
			Descriptions []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"descriptions"`
			Metrics struct {
				CVSSMetricV31 []struct {
					CVSSData struct {
						BaseScore    float64 `json:"baseScore"`
						BaseSeverity string  `json:"baseSeverity"`
					} `json:"cvssData"`
				} `json:"cvssMetricV31"`
				CVSSMetricV30 []struct {
					CVSSData struct {
						BaseScore    float64 `json:"baseScore"`
						BaseSeverity string  `json:"baseSeverity"`
					} `json:"cvssData"`
				} `json:"cvssMetricV30"`
				CVSSMetricV2 []struct {
					CVSSData struct {
						BaseScore float64 `json:"baseScore"`
					} `json:"cvssData"`
					BaseSeverity string `json:"baseSeverity"`
				} `json:"cvssMetricV2"`
			} `json:"metrics"`
			References []struct {
				URL string `json:"url"`
			} `json:"references"`
			Configurations []struct {
				Nodes []struct {
					CPEMatch []struct {
						Criteria string `json:"criteria"`
						Vulnerable bool `json:"vulnerable"`
					} `json:"cpeMatch"`
				} `json:"nodes"`
			} `json:"configurations"`
		} `json:"cve"`
	} `json:"vulnerabilities"`
}

// NewNVDClient creates a new NVD API client
func NewNVDClient(apiKey string) *NVDClient {
	// Determine rate limit based on API key presence
	rateLimit := nvdRateLimitPublic
	if apiKey != "" {
		rateLimit = nvdRateLimitWithKey
	}

	// Create rate limiter (requests per 30 seconds)
	limiter := rate.NewLimiter(rate.Every(30*time.Second/time.Duration(rateLimit)), rateLimit)

	return &NVDClient{
		httpClient: &http.Client{
			Timeout: nvdRequestTimeout,
		},
		apiKey:  apiKey,
		limiter: limiter,
		cache: &NVDCache{
			entries: make(map[string]*CacheEntry),
		},
	}
}

// QueryByCPE queries the NVD API for vulnerabilities matching a CPE identifier
func (c *NVDClient) QueryByCPE(ctx context.Context, cpe string) ([]CVEItem, error) {
	// Check cache first
	if cached, ok := c.cache.Get(cpe); ok {
		return cached, nil
	}

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build request URL
	reqURL, err := url.Parse(nvdBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("cpeName", cpe)
	reqURL.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if available
	if c.apiKey != "" {
		req.Header.Set("apiKey", c.apiKey)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NVD API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var nvdResp NVDResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvdResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to CVEItems
	items := c.convertResponse(nvdResp)

	// Cache the result
	c.cache.Set(cpe, items, nvdCacheTTL)

	return items, nil
}

// QueryByCPEBatch queries NVD for multiple CPEs with rate limiting
func (c *NVDClient) QueryByCPEBatch(ctx context.Context, cpes []string) (map[string][]CVEItem, error) {
	results := make(map[string][]CVEItem)

	for _, cpe := range cpes {
		items, err := c.QueryByCPE(ctx, cpe)
		if err != nil {
			// Log error but continue with other CPEs
			// In production, you might want to return partial results + errors
			continue
		}
		results[cpe] = items
	}

	return results, nil
}

// convertResponse converts NVD API response to our CVEItem format
func (c *NVDClient) convertResponse(resp NVDResponse) []CVEItem {
	items := make([]CVEItem, 0, len(resp.Vulnerabilities))

	for _, vuln := range resp.Vulnerabilities {
		cve := vuln.CVE

		// Extract description (prefer English)
		description := ""
		for _, desc := range cve.Descriptions {
			if desc.Lang == "en" {
				description = desc.Value
				break
			}
		}

		// Extract CVSS score and severity (prefer v3.1, then v3.0, then v2)
		cvss := 0.0
		severity := "UNKNOWN"

		if len(cve.Metrics.CVSSMetricV31) > 0 {
			cvss = cve.Metrics.CVSSMetricV31[0].CVSSData.BaseScore
			severity = cve.Metrics.CVSSMetricV31[0].CVSSData.BaseSeverity
		} else if len(cve.Metrics.CVSSMetricV30) > 0 {
			cvss = cve.Metrics.CVSSMetricV30[0].CVSSData.BaseScore
			severity = cve.Metrics.CVSSMetricV30[0].CVSSData.BaseSeverity
		} else if len(cve.Metrics.CVSSMetricV2) > 0 {
			cvss = cve.Metrics.CVSSMetricV2[0].CVSSData.BaseScore
			severity = cve.Metrics.CVSSMetricV2[0].BaseSeverity
		}

		// Extract CPEs
		cpes := []string{}
		for _, config := range cve.Configurations {
			for _, node := range config.Nodes {
				for _, match := range node.CPEMatch {
					if match.Vulnerable {
						cpes = append(cpes, match.Criteria)
					}
				}
			}
		}

		// Extract references
		refs := []string{}
		for _, ref := range cve.References {
			refs = append(refs, ref.URL)
		}

		// Parse timestamps
		published, _ := time.Parse(time.RFC3339, cve.Published)
		modified, _ := time.Parse(time.RFC3339, cve.LastModified)

		items = append(items, CVEItem{
			CVEID:       cve.ID,
			Description: description,
			CVSS:        cvss,
			Severity:    severity,
			Published:   published,
			Modified:    modified,
			CPEs:        cpes,
			References:  refs,
		})
	}

	return items
}

// Get retrieves a cached entry if it exists and is not expired
func (c *NVDCache) Get(key string) ([]CVEItem, bool) {
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, key)
		return nil, false
	}

	return entry.Data, true
}

// Set stores a cache entry with TTL
func (c *NVDCache) Set(key string, data []CVEItem, ttl time.Duration) {
	c.entries[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Clear removes all cache entries
func (c *NVDCache) Clear() {
	c.entries = make(map[string]*CacheEntry)
}

// MatchServicesToCVEs matches services to vulnerabilities based on CPE
func MatchServicesToCVEs(serviceCPEs map[string][]CPEIdentifier, cvesByCPE map[string][]CVEItem) []VulnMatch {
	matches := []VulnMatch{}

	for serviceID, cpes := range serviceCPEs {
		for _, cpe := range cpes {
			if cves, exists := cvesByCPE[cpe.CPE]; exists {
				for _, cve := range cves {
					matches = append(matches, VulnMatch{
						ServiceID: serviceID,
						CVE:       cve.CVEID,
						CVSS:      cve.CVSS,
						Severity:  cve.Severity,
					})
				}
			}
		}
	}

	return matches
}

// FilterHighSeverity filters vulnerability matches to only include HIGH and CRITICAL
func FilterHighSeverity(matches []VulnMatch) []VulnMatch {
	filtered := []VulnMatch{}

	for _, match := range matches {
		if match.Severity == "HIGH" || match.Severity == "CRITICAL" {
			filtered = append(filtered, match)
		}
	}

	return filtered
}

// DeduplicateMatches removes duplicate vulnerability matches
func DeduplicateMatches(matches []VulnMatch) []VulnMatch {
	seen := make(map[string]bool)
	deduplicated := []VulnMatch{}

	for _, match := range matches {
		key := fmt.Sprintf("%s:%s", match.ServiceID, match.CVE)
		if !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, match)
		}
	}

	return deduplicated
}
