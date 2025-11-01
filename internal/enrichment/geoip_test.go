package enrichment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeoIPClient_NewClient tests client creation
func TestGeoIPClient_NewClient(t *testing.T) {
	tests := []struct {
		name      string
		config    GeoIPConfig
		wantError bool
	}{
		{
			name: "valid config with MMDB path",
			config: GeoIPConfig{
				MMDBPath: "/tmp/test.mmdb", // Will fail to open but client should still be created
			},
			wantError: false, // Client creation succeeds even if MMDB fails to open
		},
		{
			name: "config without MMDB (API only)",
			config: GeoIPConfig{
				APIKey: "test_key",
				APIURL: "https://ipinfo.io",
			},
			wantError: false,
		},
		{
			name:      "empty config",
			config:    GeoIPConfig{},
			wantError: false, // Client can be created, but lookups will fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeoIPClient(tt.config)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				// Client creation should succeed even if MMDB doesn't exist
				assert.NotNil(t, client)
				if client != nil {
					client.Close()
				}
			}
		})
	}
}

// TestGeoIPClient_Lookup tests single IP lookup
func TestGeoIPClient_Lookup(t *testing.T) {
	// Skip if MMDB file not available
	mmdbPath := getTestMMDBPath()
	if mmdbPath == "" {
		t.Skip("No GeoIP MMDB file available for testing (set GEOIP_MMDB_PATH environment variable)")
	}

	client, err := NewGeoIPClient(GeoIPConfig{
		MMDBPath: mmdbPath,
	})
	require.NoError(t, err)
	defer client.Close()

	tests := []struct {
		name      string
		ip        string
		wantError bool
		validate  func(t *testing.T, info *GeoIPInfo)
	}{
		{
			name:      "valid public IP (Google DNS)",
			ip:        "8.8.8.8",
			wantError: false,
			validate: func(t *testing.T, info *GeoIPInfo) {
				assert.Equal(t, "8.8.8.8", info.IP)
				assert.NotEmpty(t, info.Country)
				assert.NotEmpty(t, info.CountryCC)
				// Google DNS is in US
				assert.Equal(t, "US", info.CountryCC)
			},
		},
		{
			name:      "valid public IP (Cloudflare DNS)",
			ip:        "1.1.1.1",
			wantError: false,
			validate: func(t *testing.T, info *GeoIPInfo) {
				assert.Equal(t, "1.1.1.1", info.IP)
				assert.NotEmpty(t, info.Country)
				assert.NotEmpty(t, info.CountryCC)
			},
		},
		{
			name:      "invalid IP address",
			ip:        "invalid",
			wantError: true,
		},
		{
			name:      "empty IP address",
			ip:        "",
			wantError: true,
		},
		{
			name:      "private IP (RFC 1918)",
			ip:        "192.168.1.1",
			wantError: true, // MaxMind MMDB typically doesn't have private IP data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.Lookup(tt.ip)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				if tt.validate != nil {
					tt.validate(t, info)
				}
			}
		})
	}
}

// TestGeoIPClient_LookupBatch tests batch IP lookup
func TestGeoIPClient_LookupBatch(t *testing.T) {
	// Skip if MMDB file not available
	mmdbPath := getTestMMDBPath()
	if mmdbPath == "" {
		t.Skip("No GeoIP MMDB file available for testing (set GEOIP_MMDB_PATH environment variable)")
	}

	client, err := NewGeoIPClient(GeoIPConfig{
		MMDBPath: mmdbPath,
	})
	require.NoError(t, err)
	defer client.Close()

	t.Run("batch of valid IPs", func(t *testing.T) {
		ips := []string{
			"8.8.8.8",      // Google DNS
			"1.1.1.1",      // Cloudflare DNS
			"208.67.222.222", // OpenDNS
		}

		results, err := client.LookupBatch(ips)
		require.NoError(t, err)
		assert.NotEmpty(t, results)

		// Should have results for all valid IPs
		for _, ip := range ips {
			info, ok := results[ip]
			assert.True(t, ok, "Expected result for IP %s", ip)
			if ok {
				assert.Equal(t, ip, info.IP)
				assert.NotEmpty(t, info.Country)
			}
		}
	})

	t.Run("batch with mixed valid and invalid IPs", func(t *testing.T) {
		ips := []string{
			"8.8.8.8",       // Valid
			"invalid",       // Invalid
			"1.1.1.1",       // Valid
			"192.168.1.1",   // Private (likely to fail)
		}

		results, err := client.LookupBatch(ips)
		// Should succeed even if some IPs fail
		require.NoError(t, err)
		assert.NotEmpty(t, results)

		// Should have at least the valid public IPs
		assert.Contains(t, results, "8.8.8.8")
		assert.Contains(t, results, "1.1.1.1")
	})

	t.Run("empty batch", func(t *testing.T) {
		results, err := client.LookupBatch([]string{})
		assert.Error(t, err)
		assert.Empty(t, results)
	})

	t.Run("batch of all invalid IPs", func(t *testing.T) {
		ips := []string{
			"invalid1",
			"invalid2",
			"not-an-ip",
		}

		results, err := client.LookupBatch(ips)
		assert.Error(t, err) // Should error when no successful lookups
		assert.Empty(t, results)
	})
}

// TestValidateMMDB tests MMDB file validation
func TestValidateMMDB(t *testing.T) {
	mmdbPath := getTestMMDBPath()
	if mmdbPath == "" {
		t.Skip("No GeoIP MMDB file available for testing (set GEOIP_MMDB_PATH environment variable)")
	}

	t.Run("valid MMDB file", func(t *testing.T) {
		err := ValidateMMDB(mmdbPath)
		assert.NoError(t, err)
	})

	t.Run("non-existent file", func(t *testing.T) {
		err := ValidateMMDB("/tmp/nonexistent.mmdb")
		assert.Error(t, err)
	})

	t.Run("invalid file", func(t *testing.T) {
		// Create a temporary invalid file
		tmpFile := filepath.Join(t.TempDir(), "invalid.mmdb")
		err := os.WriteFile(tmpFile, []byte("not a valid MMDB"), 0644)
		require.NoError(t, err)

		err = ValidateMMDB(tmpFile)
		assert.Error(t, err)
	})
}

// TestGeoIPClient_Close tests client cleanup
func TestGeoIPClient_Close(t *testing.T) {
	mmdbPath := getTestMMDBPath()
	if mmdbPath == "" {
		t.Skip("No GeoIP MMDB file available for testing (set GEOIP_MMDB_PATH environment variable)")
	}

	client, err := NewGeoIPClient(GeoIPConfig{
		MMDBPath: mmdbPath,
	})
	require.NoError(t, err)

	// Close should not error
	err = client.Close()
	assert.NoError(t, err)

	// Second close should not error
	err = client.Close()
	assert.NoError(t, err)
}

// TestGeoIPInfo_Struct tests the GeoIPInfo structure
func TestGeoIPInfo_Struct(t *testing.T) {
	info := &GeoIPInfo{
		IP:        "8.8.8.8",
		City:      "Mountain View",
		Region:    "California",
		Country:   "United States",
		CountryCC: "US",
		Latitude:  37.4056,
		Longitude: -122.0775,
	}

	assert.Equal(t, "8.8.8.8", info.IP)
	assert.Equal(t, "Mountain View", info.City)
	assert.Equal(t, "California", info.Region)
	assert.Equal(t, "United States", info.Country)
	assert.Equal(t, "US", info.CountryCC)
	assert.InDelta(t, 37.4056, info.Latitude, 0.0001)
	assert.InDelta(t, -122.0775, info.Longitude, 0.0001)
}

// getTestMMDBPath returns the path to a test MMDB file
// Checks environment variable GEOIP_MMDB_PATH or common locations
func getTestMMDBPath() string {
	// Check environment variable first
	if path := os.Getenv("GEOIP_MMDB_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check common locations
	commonPaths := []string{
		"/usr/share/GeoIP/GeoLite2-City.mmdb",
		"/var/lib/GeoIP/GeoLite2-City.mmdb",
		"/opt/GeoIP/GeoLite2-City.mmdb",
		"./GeoLite2-City.mmdb",
		"../GeoLite2-City.mmdb",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
