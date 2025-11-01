package db

import (
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBuildHostQuery(t *testing.T) {
	tests := []struct {
		name          string
		ip            string
		depth         int
		expectedQuery string
	}{
		{
			name:          "depth 0 - host only",
			ip:            "1.2.3.4",
			depth:         0,
			expectedQuery: `SELECT * FROM host WHERE ip = $ip LIMIT 1;`,
		},
		{
			name:  "depth 1 - host and ports",
			ip:    "1.2.3.4",
			depth: 1,
			expectedQuery: `SELECT *,
			->HAS->port.* AS ports
		FROM host WHERE ip = $ip LIMIT 1;`,
		},
		{
			name:  "depth 2 - host, ports, and services",
			ip:    "5.6.7.8",
			depth: 2,
			expectedQuery: `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services
		FROM host WHERE ip = $ip LIMIT 1;`,
		},
		{
			name:  "depth 3 - host, ports, services, and vulns",
			ip:    "10.0.0.1",
			depth: 3,
			expectedQuery: `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services,
			->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns
		FROM host WHERE ip = $ip LIMIT 1;`,
		},
		{
			name:  "depth 4 - extended relationships",
			ip:    "192.168.1.1",
			depth: 4,
			expectedQuery: `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services,
			->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns,
			->IN_CITY->city.* AS city_detail,
			->IN_ASN->asn.* AS asn_detail
		FROM host WHERE ip = $ip LIMIT 1;`,
		},
		{
			name:  "depth 5 - maximum depth",
			ip:    "172.16.0.1",
			depth: 5,
			expectedQuery: `SELECT *,
			->HAS->port.* AS ports,
			->HAS->port->RUNS->service.* AS services,
			->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns,
			->IN_CITY->city.* AS city_detail,
			->IN_ASN->asn.* AS asn_detail
		FROM host WHERE ip = $ip LIMIT 1;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := buildHostQuery(tt.ip, tt.depth)
			assert.Contains(t, query, "FROM host WHERE ip = $ip")
			assert.Contains(t, query, "LIMIT 1;")

			// Check depth-specific clauses
			if tt.depth >= 1 {
				assert.Contains(t, query, "->HAS->port.* AS ports")
			}
			if tt.depth >= 2 {
				assert.Contains(t, query, "->HAS->port->RUNS->service.* AS services")
			}
			if tt.depth >= 3 {
				assert.Contains(t, query, "->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns")
			}
			if tt.depth >= 4 {
				assert.Contains(t, query, "->IN_CITY->city.* AS city_detail")
				assert.Contains(t, query, "->IN_ASN->asn.* AS asn_detail")
			}
		})
	}
}

func TestGetStringField(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "existing string field",
			data:     map[string]interface{}{"ip": "1.2.3.4"},
			key:      "ip",
			expected: "1.2.3.4",
		},
		{
			name:     "missing field",
			data:     map[string]interface{}{"ip": "1.2.3.4"},
			key:      "asn",
			expected: "",
		},
		{
			name:     "non-string field",
			data:     map[string]interface{}{"port": 80},
			key:      "port",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringField(tt.data, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntField(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected int
		expectOk bool
	}{
		{
			name:     "existing int field",
			data:     map[string]interface{}{"port": 80},
			key:      "port",
			expected: 80,
			expectOk: true,
		},
		{
			name:     "existing int64 field",
			data:     map[string]interface{}{"asn": int64(15169)},
			key:      "asn",
			expected: 15169,
			expectOk: true,
		},
		{
			name:     "existing float64 field",
			data:     map[string]interface{}{"count": float64(42)},
			key:      "count",
			expected: 42,
			expectOk: true,
		},
		{
			name:     "missing field",
			data:     map[string]interface{}{"port": 80},
			key:      "asn",
			expected: 0,
			expectOk: false,
		},
		{
			name:     "non-numeric field",
			data:     map[string]interface{}{"ip": "1.2.3.4"},
			key:      "ip",
			expected: 0,
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := getIntField(tt.data, tt.key)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectOk, ok)
		})
	}
}

func TestGetFloatField(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected float64
		expectOk bool
	}{
		{
			name:     "existing float64 field",
			data:     map[string]interface{}{"cvss": 9.8},
			key:      "cvss",
			expected: 9.8,
			expectOk: true,
		},
		{
			name:     "existing float32 field",
			data:     map[string]interface{}{"score": float32(7.5)},
			key:      "score",
			expected: 7.5,
			expectOk: true,
		},
		{
			name:     "existing int field",
			data:     map[string]interface{}{"count": 10},
			key:      "count",
			expected: 10.0,
			expectOk: true,
		},
		{
			name:     "missing field",
			data:     map[string]interface{}{"cvss": 9.8},
			key:      "epss",
			expected: 0,
			expectOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := getFloatField(tt.data, tt.key)
			assert.InDelta(t, tt.expected, result, 0.01)
			assert.Equal(t, tt.expectOk, ok)
		})
	}
}

func TestParseTimeField(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		data      map[string]interface{}
		key       string
		shouldErr bool
	}{
		{
			name:      "valid time.Time",
			data:      map[string]interface{}{"timestamp": now},
			key:       "timestamp",
			shouldErr: false,
		},
		{
			name:      "valid RFC3339 string",
			data:      map[string]interface{}{"timestamp": "2025-11-01T12:00:00Z"},
			key:       "timestamp",
			shouldErr: false,
		},
		{
			name:      "missing field",
			data:      map[string]interface{}{"other": "value"},
			key:       "timestamp",
			shouldErr: true,
		},
		{
			name:      "invalid time string",
			data:      map[string]interface{}{"timestamp": "not-a-time"},
			key:       "timestamp",
			shouldErr: true,
		},
		{
			name:      "invalid type",
			data:      map[string]interface{}{"timestamp": 123},
			key:       "timestamp",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeField(tt.data, tt.key)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}

func TestParsePorts(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		portsData []interface{}
		depth     int
		expected  int
	}{
		{
			name: "valid ports",
			portsData: []interface{}{
				map[string]interface{}{
					"number":     80,
					"protocol":   "tcp",
					"transport":  "plain",
					"first_seen": "2025-11-01T12:00:00Z",
					"last_seen":  "2025-11-01T13:00:00Z",
				},
				map[string]interface{}{
					"number":     443,
					"protocol":   "tcp",
					"transport":  "tls",
					"first_seen": "2025-11-01T12:00:00Z",
					"last_seen":  "2025-11-01T13:00:00Z",
				},
			},
			depth:    2,
			expected: 2,
		},
		{
			name:      "empty ports",
			portsData: []interface{}{},
			depth:     2,
			expected:  0,
		},
		{
			name: "invalid port data type",
			portsData: []interface{}{
				"not-a-map",
			},
			depth:    2,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := parsePorts(tt.portsData, tt.depth, logger)
			assert.Len(t, ports, tt.expected)
			for _, port := range ports {
				assert.NotZero(t, port.Number)
				assert.NotEmpty(t, port.Protocol)
			}
		})
	}
}

func TestParseServices(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name         string
		servicesData []interface{}
		depth        int
		expected     int
	}{
		{
			name: "valid services",
			servicesData: []interface{}{
				map[string]interface{}{
					"name":       "http",
					"product":    "nginx",
					"version":    "1.25.1",
					"cpe":        []interface{}{"cpe:2.3:a:nginx:nginx:1.25.1:*:*:*:*:*:*:*"},
					"first_seen": "2025-11-01T12:00:00Z",
					"last_seen":  "2025-11-01T13:00:00Z",
				},
			},
			depth:    3,
			expected: 1,
		},
		{
			name:         "empty services",
			servicesData: []interface{}{},
			depth:        3,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := parseServices(tt.servicesData, tt.depth, logger)
			assert.Len(t, services, tt.expected)
		})
	}
}

func TestParseVulns(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		vulnsData []interface{}
		expected  int
	}{
		{
			name: "valid vulnerabilities",
			vulnsData: []interface{}{
				map[string]interface{}{
					"cve_id":     "CVE-2023-12345",
					"cvss":       9.8,
					"severity":   "critical",
					"kev_flag":   true,
					"first_seen": "2025-11-01T12:00:00Z",
				},
			},
			expected: 1,
		},
		{
			name:      "empty vulnerabilities",
			vulnsData: []interface{}{},
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vulns := parseVulns(tt.vulnsData, logger)
			assert.Len(t, vulns, tt.expected)
			for _, vuln := range vulns {
				assert.NotEmpty(t, vuln.CVEID)
			}
		})
	}
}

func TestQueryDepthConstants(t *testing.T) {
	assert.Equal(t, 0, int(models.DepthHostOnly))
	assert.Equal(t, 1, int(models.DepthWithPorts))
	assert.Equal(t, 2, int(models.DepthWithServices))
	assert.Equal(t, 3, int(models.DepthWithVulns))
	assert.Equal(t, 5, int(models.DepthMaximum))
}
