package enrichment

import (
	"context"
	"testing"
	"time"
)

func TestNewNVDClient(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "without API key",
			apiKey: "",
		},
		{
			name:   "with API key",
			apiKey: "test-api-key-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewNVDClient(tt.apiKey)

			if client == nil {
				t.Fatal("NewNVDClient() returned nil")
			}

			if client.httpClient == nil {
				t.Error("httpClient is nil")
			}

			if client.limiter == nil {
				t.Error("limiter is nil")
			}

			if client.cache == nil {
				t.Error("cache is nil")
			}

			if client.apiKey != tt.apiKey {
				t.Errorf("apiKey = %v, want %v", client.apiKey, tt.apiKey)
			}
		})
	}
}

func TestNVDCache(t *testing.T) {
	cache := &NVDCache{
		entries: make(map[string]*CacheEntry),
	}

	testData := []CVEItem{
		{
			CVEID:    "CVE-2023-1234",
			CVSS:     9.8,
			Severity: "CRITICAL",
		},
	}

	t.Run("Set and Get", func(t *testing.T) {
		// Set cache entry
		cache.Set("test-key", testData, 1*time.Hour)

		// Get cache entry
		data, ok := cache.Get("test-key")
		if !ok {
			t.Fatal("Get() returned false for existing key")
		}

		if len(data) != 1 {
			t.Errorf("cached data length = %d, want 1", len(data))
		}

		if data[0].CVEID != "CVE-2023-1234" {
			t.Errorf("cached CVE ID = %v, want CVE-2023-1234", data[0].CVEID)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		_, ok := cache.Get("non-existent")
		if ok {
			t.Error("Get() returned true for non-existent key")
		}
	})

	t.Run("Expired entry", func(t *testing.T) {
		// Set entry with very short TTL
		cache.Set("expire-test", testData, 1*time.Nanosecond)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Try to get expired entry
		_, ok := cache.Get("expire-test")
		if ok {
			t.Error("Get() returned true for expired entry")
		}
	})

	t.Run("Clear cache", func(t *testing.T) {
		cache.Set("key1", testData, 1*time.Hour)
		cache.Set("key2", testData, 1*time.Hour)

		cache.Clear()

		_, ok := cache.Get("key1")
		if ok {
			t.Error("Get() returned true after Clear()")
		}
	})
}

func TestMatchServicesToCVEs(t *testing.T) {
	serviceCPEs := map[string][]CPEIdentifier{
		"service1": {
			{
				Vendor:  "nginx",
				Product: "nginx",
				Version: "1.24.0",
				CPE:     "cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*",
			},
		},
		"service2": {
			{
				Vendor:  "apache",
				Product: "http_server",
				Version: "2.4.57",
				CPE:     "cpe:2.3:a:apache:http_server:2.4.57:*:*:*:*:*:*:*",
			},
		},
	}

	cvesByCPE := map[string][]CVEItem{
		"cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*": {
			{
				CVEID:    "CVE-2023-1001",
				CVSS:     7.5,
				Severity: "HIGH",
			},
			{
				CVEID:    "CVE-2023-1002",
				CVSS:     5.3,
				Severity: "MEDIUM",
			},
		},
		"cpe:2.3:a:apache:http_server:2.4.57:*:*:*:*:*:*:*": {
			{
				CVEID:    "CVE-2023-2001",
				CVSS:     9.8,
				Severity: "CRITICAL",
			},
		},
	}

	matches := MatchServicesToCVEs(serviceCPEs, cvesByCPE)

	// Should have 3 matches total (2 for nginx, 1 for apache)
	if len(matches) != 3 {
		t.Errorf("MatchServicesToCVEs() returned %d matches, want 3", len(matches))
	}

	// Verify service1 matches
	service1Matches := 0
	for _, match := range matches {
		if match.ServiceID == "service1" {
			service1Matches++
		}
	}
	if service1Matches != 2 {
		t.Errorf("service1 has %d matches, want 2", service1Matches)
	}

	// Verify service2 matches
	service2Matches := 0
	for _, match := range matches {
		if match.ServiceID == "service2" {
			service2Matches++
			if match.CVE != "CVE-2023-2001" {
				t.Errorf("service2 CVE = %v, want CVE-2023-2001", match.CVE)
			}
		}
	}
	if service2Matches != 1 {
		t.Errorf("service2 has %d matches, want 1", service2Matches)
	}
}

func TestFilterHighSeverity(t *testing.T) {
	matches := []VulnMatch{
		{ServiceID: "s1", CVE: "CVE-1", CVSS: 9.8, Severity: "CRITICAL"},
		{ServiceID: "s2", CVE: "CVE-2", CVSS: 7.5, Severity: "HIGH"},
		{ServiceID: "s3", CVE: "CVE-3", CVSS: 5.3, Severity: "MEDIUM"},
		{ServiceID: "s4", CVE: "CVE-4", CVSS: 3.1, Severity: "LOW"},
	}

	filtered := FilterHighSeverity(matches)

	// Should have 2 matches (CRITICAL + HIGH)
	if len(filtered) != 2 {
		t.Errorf("FilterHighSeverity() returned %d matches, want 2", len(filtered))
	}

	// Verify only HIGH and CRITICAL
	for _, match := range filtered {
		if match.Severity != "HIGH" && match.Severity != "CRITICAL" {
			t.Errorf("FilterHighSeverity() included %s severity", match.Severity)
		}
	}
}

func TestDeduplicateMatches(t *testing.T) {
	matches := []VulnMatch{
		{ServiceID: "s1", CVE: "CVE-1", CVSS: 9.8, Severity: "CRITICAL"},
		{ServiceID: "s1", CVE: "CVE-1", CVSS: 9.8, Severity: "CRITICAL"}, // Duplicate
		{ServiceID: "s1", CVE: "CVE-2", CVSS: 7.5, Severity: "HIGH"},
		{ServiceID: "s2", CVE: "CVE-1", CVSS: 9.8, Severity: "CRITICAL"}, // Different service, not duplicate
	}

	deduplicated := DeduplicateMatches(matches)

	// Should have 3 unique matches
	if len(deduplicated) != 3 {
		t.Errorf("DeduplicateMatches() returned %d matches, want 3", len(deduplicated))
	}

	// Verify uniqueness
	seen := make(map[string]bool)
	for _, match := range deduplicated {
		key := match.ServiceID + ":" + match.CVE
		if seen[key] {
			t.Errorf("DeduplicateMatches() has duplicate: %s", key)
		}
		seen[key] = true
	}
}

func TestConvertResponse(t *testing.T) {
	client := NewNVDClient("")

	// Mock NVD API response
	mockResp := NVDResponse{
		ResultsPerPage: 1,
		StartIndex:     0,
		TotalResults:   1,
		Vulnerabilities: []struct {
			CVE struct {
				ID           string `json:"id"`
				Published    string `json:"published"`
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
							Criteria   string `json:"criteria"`
							Vulnerable bool   `json:"vulnerable"`
						} `json:"cpeMatch"`
					} `json:"nodes"`
				} `json:"configurations"`
			} `json:"cve"`
		}{
			{
				CVE: struct {
					ID           string `json:"id"`
					Published    string `json:"published"`
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
								Criteria   string `json:"criteria"`
								Vulnerable bool   `json:"vulnerable"`
							} `json:"cpeMatch"`
						} `json:"nodes"`
					} `json:"configurations"`
				}{
					ID:           "CVE-2023-1234",
					Published:    "2023-03-15T10:00:00.000Z",
					LastModified: "2023-03-20T12:00:00.000Z",
					Descriptions: []struct {
						Lang  string `json:"lang"`
						Value string `json:"value"`
					}{
						{Lang: "en", Value: "Test vulnerability description"},
					},
					Metrics: struct {
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
					}{
						CVSSMetricV31: []struct {
							CVSSData struct {
								BaseScore    float64 `json:"baseScore"`
								BaseSeverity string  `json:"baseSeverity"`
							} `json:"cvssData"`
						}{
							{
								CVSSData: struct {
									BaseScore    float64 `json:"baseScore"`
									BaseSeverity string  `json:"baseSeverity"`
								}{
									BaseScore:    9.8,
									BaseSeverity: "CRITICAL",
								},
							},
						},
					},
					References: []struct {
						URL string `json:"url"`
					}{
						{URL: "https://example.com/vuln1"},
					},
					Configurations: []struct {
						Nodes []struct {
							CPEMatch []struct {
								Criteria   string `json:"criteria"`
								Vulnerable bool   `json:"vulnerable"`
							} `json:"cpeMatch"`
						} `json:"nodes"`
					}{
						{
							Nodes: []struct {
								CPEMatch []struct {
									Criteria   string `json:"criteria"`
									Vulnerable bool   `json:"vulnerable"`
								} `json:"cpeMatch"`
							}{
								{
									CPEMatch: []struct {
										Criteria   string `json:"criteria"`
										Vulnerable bool   `json:"vulnerable"`
									}{
										{
											Criteria:   "cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*",
											Vulnerable: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	items := client.convertResponse(mockResp)

	if len(items) != 1 {
		t.Fatalf("convertResponse() returned %d items, want 1", len(items))
	}

	item := items[0]

	if item.CVEID != "CVE-2023-1234" {
		t.Errorf("CVE ID = %v, want CVE-2023-1234", item.CVEID)
	}

	if item.CVSS != 9.8 {
		t.Errorf("CVSS = %v, want 9.8", item.CVSS)
	}

	if item.Severity != "CRITICAL" {
		t.Errorf("Severity = %v, want CRITICAL", item.Severity)
	}

	if item.Description != "Test vulnerability description" {
		t.Errorf("Description = %v, want 'Test vulnerability description'", item.Description)
	}

	if len(item.CPEs) != 1 {
		t.Errorf("CPEs length = %d, want 1", len(item.CPEs))
	}

	if len(item.References) != 1 {
		t.Errorf("References length = %d, want 1", len(item.References))
	}
}

// Note: We skip testing actual NVD API calls to avoid rate limiting and external dependencies
// In production, you would use mocks or record/replay HTTP interactions
func TestQueryByCPE_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := NewNVDClient("")
	ctx := context.Background()

	// Test with a known CPE (this will hit the real NVD API)
	// Only run this manually when needed
	t.Skip("Skipping real NVD API call to avoid rate limits")

	cpe := "cpe:2.3:a:nginx:nginx:1.18.0:*:*:*:*:*:*:*"
	items, err := client.QueryByCPE(ctx, cpe)

	if err != nil {
		t.Fatalf("QueryByCPE() error = %v", err)
	}

	t.Logf("Found %d vulnerabilities for %s", len(items), cpe)

	// Verify cache works
	cachedItems, err := client.QueryByCPE(ctx, cpe)
	if err != nil {
		t.Fatalf("QueryByCPE() (cached) error = %v", err)
	}

	if len(cachedItems) != len(items) {
		t.Errorf("Cached result length = %d, want %d", len(cachedItems), len(items))
	}
}
