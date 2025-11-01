package cli

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutputOptions(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		noColor  bool
		expected OutputFormat
	}{
		{
			name:     "json format",
			format:   "json",
			noColor:  false,
			expected: FormatJSON,
		},
		{
			name:     "yaml format",
			format:   "yaml",
			noColor:  false,
			expected: FormatYAML,
		},
		{
			name:     "yml format",
			format:   "yml",
			noColor:  false,
			expected: FormatYAML,
		},
		{
			name:     "table format",
			format:   "table",
			noColor:  false,
			expected: FormatTable,
		},
		{
			name:     "default to table",
			format:   "",
			noColor:  false,
			expected: FormatTable,
		},
		{
			name:     "invalid defaults to table",
			format:   "invalid",
			noColor:  false,
			expected: FormatTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewOutputOptions(tt.format, tt.noColor)
			assert.Equal(t, tt.expected, opts.Format)
			assert.Equal(t, tt.noColor, opts.NoColor)
		})
	}
}

func TestFormatJSON(t *testing.T) {
	data := map[string]interface{}{
		"test":  "value",
		"count": 42,
	}

	var buf bytes.Buffer
	err := formatJSON(&buf, data)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, `"test"`)
	assert.Contains(t, output, `"value"`)
	assert.Contains(t, output, `"count"`)
	assert.Contains(t, output, `42`)
}

func TestFormatYAML(t *testing.T) {
	data := map[string]interface{}{
		"test":  "value",
		"count": 42,
	}

	var buf bytes.Buffer
	err := formatYAML(&buf, data)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "test: value")
	assert.Contains(t, output, "count: 42")
}

func TestFormatHostTable(t *testing.T) {
	result := &models.HostQueryResponse{
		IP:        "1.2.3.4",
		ASN:       15169,
		City:      "Mountain View",
		Country:   "United States",
		FirstSeen: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		LastSeen:  time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Ports: []models.PortDetail{
			{
				Number:   80,
				Protocol: "tcp",
				Services: []models.ServiceDetail{
					{
						Name:    "http",
						Product: "nginx",
						Version: "1.25.1",
					},
				},
			},
			{
				Number:   443,
				Protocol: "tcp",
				Services: []models.ServiceDetail{
					{
						Name:    "https",
						Product: "nginx",
						Version: "1.25.1",
					},
				},
			},
		},
		Vulns: []models.VulnDetail{
			{
				CVEID:     "CVE-2024-1234",
				CVSS:      9.8,
				Severity:  "Critical",
				KEVFlag:   true,
				FirstSeen: time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	var buf bytes.Buffer
	opts := &OutputOptions{
		Format:     FormatTable,
		NoColor:    true,
		Writer:     &buf,
		IsTerminal: false,
	}

	err := formatHostTable(opts, result)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "1.2.3.4")
	assert.Contains(t, output, "15169")
	assert.Contains(t, output, "Mountain View")
	assert.Contains(t, output, "80")
	assert.Contains(t, output, "443")
	assert.Contains(t, output, "nginx")
	assert.Contains(t, output, "CVE-2024-1234")
	assert.Contains(t, output, "9.8")
	assert.Contains(t, output, "Critical")
}

func TestFormatGraphTable(t *testing.T) {
	result := &models.GraphQueryResponse{
		Results: []models.HostResult{
			{
				IP:       "1.2.3.4",
				ASN:      15169,
				City:     "Paris",
				Country:  "France",
				LastSeen: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
				Ports: []models.Port{
					{Number: 80, Protocol: "tcp"},
					{Number: 443, Protocol: "tcp"},
				},
				Services: []models.Service{
					{Name: "http", Product: "nginx"},
					{Name: "https", Product: "nginx"},
				},
			},
			{
				IP:       "5.6.7.8",
				ASN:      15169,
				City:     "London",
				Country:  "United Kingdom",
				LastSeen: time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC),
				Ports: []models.Port{
					{Number: 22, Protocol: "tcp"},
				},
				Services: []models.Service{
					{Name: "ssh", Product: "openssh"},
				},
			},
		},
		Pagination: models.PaginationMetadata{
			Limit:      100,
			Offset:     0,
			Total:      2,
			HasMore:    false,
			NextOffset: 0,
		},
		QueryTime: 123.45,
	}

	var buf bytes.Buffer
	opts := &OutputOptions{
		Format:     FormatTable,
		NoColor:    true,
		Writer:     &buf,
		IsTerminal: false,
	}

	err := formatGraphTable(opts, result)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "1.2.3.4")
	assert.Contains(t, output, "5.6.7.8")
	assert.Contains(t, output, "Paris")
	assert.Contains(t, output, "London")
	assert.Contains(t, output, "15169")
	assert.Contains(t, output, "123.45 ms")
}

func TestFormatGraphTable_Empty(t *testing.T) {
	result := &models.GraphQueryResponse{
		Results:    []models.HostResult{},
		Pagination: models.PaginationMetadata{},
		QueryTime:  0.0,
	}

	var buf bytes.Buffer
	opts := &OutputOptions{
		Format:     FormatTable,
		NoColor:    true,
		Writer:     &buf,
		IsTerminal: false,
	}

	err := formatGraphTable(opts, result)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No results found")
}

func TestFormatGraphTable_WithPagination(t *testing.T) {
	result := &models.GraphQueryResponse{
		Results: []models.HostResult{
			{
				IP:       "1.2.3.4",
				ASN:      15169,
				LastSeen: time.Now(),
			},
		},
		Pagination: models.PaginationMetadata{
			Limit:      50,
			Offset:     0,
			Total:      150,
			HasMore:    true,
			NextOffset: 50,
		},
		QueryTime: 100.0,
	}

	var buf bytes.Buffer
	opts := &OutputOptions{
		Format:     FormatTable,
		NoColor:    true,
		Writer:     &buf,
		IsTerminal: false,
	}

	err := formatGraphTable(opts, result)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "More results available")
	assert.Contains(t, output, "offset 50")
}

func TestFormatSimilarTable(t *testing.T) {
	result := &models.SimilarResponse{
		Query: "nginx remote code execution",
		Results: []models.VulnResult{
			{
				CVEID:   "CVE-2024-1234",
				Title:   "Nginx Buffer Overflow Vulnerability",
				Summary: "A buffer overflow in nginx allows remote code execution",
				CVSS:    9.8,
				Score:   0.95,
			},
			{
				CVEID:   "CVE-2024-5678",
				Title:   "Nginx Authentication Bypass",
				Summary: "Authentication bypass leading to RCE",
				CVSS:    8.1,
				Score:   0.87,
			},
		},
		Count:     2,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	opts := &OutputOptions{
		Format:     FormatTable,
		NoColor:    true,
		Writer:     &buf,
		IsTerminal: false,
	}

	err := formatSimilarTable(opts, result)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "nginx remote code execution")
	assert.Contains(t, output, "CVE-2024-1234")
	assert.Contains(t, output, "CVE-2024-5678")
	assert.Contains(t, output, "0.95")
	assert.Contains(t, output, "0.87")
	assert.Contains(t, output, "9.8")
	assert.Contains(t, output, "8.1")
}

func TestFormatSimilarTable_Empty(t *testing.T) {
	result := &models.SimilarResponse{
		Query:     "test query",
		Results:   []models.VulnResult{},
		Count:     0,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	var buf bytes.Buffer
	opts := &OutputOptions{
		Format:     FormatTable,
		NoColor:    true,
		Writer:     &buf,
		IsTerminal: false,
	}

	err := formatSimilarTable(opts, result)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No similar vulnerabilities found")
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "valid time",
			input:    time.Date(2024, 12, 1, 15, 30, 0, 0, time.UTC),
			expected: "2024-12-01 15:30",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string",
			input:    "this is a very long string that should be truncated",
			maxLen:   20,
			expected: "this is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAPIURL(t *testing.T) {
	// Save and restore environment
	originalEnv := os.Getenv("SPECTRA_API_URL")
	defer func() {
		if originalEnv != "" {
			os.Setenv("SPECTRA_API_URL", originalEnv)
		} else {
			os.Unsetenv("SPECTRA_API_URL")
		}
	}()

	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "flag set",
			setup: func() {
				apiURL = "http://flag:3000"
			},
			expected: "http://flag:3000",
		},
		{
			name: "environment variable",
			setup: func() {
				apiURL = ""
				os.Setenv("SPECTRA_API_URL", "http://env:3000")
			},
			expected: "http://env:3000",
		},
		{
			name: "default",
			setup: func() {
				apiURL = ""
				os.Unsetenv("SPECTRA_API_URL")
			},
			expected: "http://localhost:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result := getAPIURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultFormatter(t *testing.T) {
	formatter := NewFormatter()
	assert.NotNil(t, formatter)

	t.Run("FormatHostQuery JSON", func(t *testing.T) {
		result := &models.HostQueryResponse{
			IP:  "1.2.3.4",
			ASN: 15169,
		}
		var buf bytes.Buffer
		opts := &OutputOptions{
			Format: FormatJSON,
			Writer: &buf,
		}

		err := formatter.FormatHostQuery(opts, result)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), `"ip"`)
		assert.Contains(t, buf.String(), `"1.2.3.4"`)
	})

	t.Run("FormatGraphQuery YAML", func(t *testing.T) {
		result := &models.GraphQueryResponse{
			Results:   []models.HostResult{},
			QueryTime: 100.0,
		}
		var buf bytes.Buffer
		opts := &OutputOptions{
			Format: FormatYAML,
			Writer: &buf,
		}

		err := formatter.FormatGraphQuery(opts, result)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "query_time_ms")
	})

	t.Run("FormatSimilarQuery Table", func(t *testing.T) {
		result := &models.SimilarResponse{
			Query:   "test",
			Results: []models.VulnResult{},
			Count:   0,
		}
		var buf bytes.Buffer
		opts := &OutputOptions{
			Format:  FormatTable,
			NoColor: true,
			Writer:  &buf,
		}

		err := formatter.FormatSimilarQuery(opts, result)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "test")
	})
}
