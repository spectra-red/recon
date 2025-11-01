package db

import (
	"context"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap/zaptest"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *surrealdb.DB {
	ctx := context.Background()

	db, err := surrealdb.New("ws://localhost:8000/rpc")
	require.NoError(t, err, "failed to connect to SurrealDB")

	// Sign in
	_, err = db.SignIn(ctx, map[string]interface{}{
		"user": "root",
		"pass": "root",
	})
	require.NoError(t, err, "failed to sign in")

	// Use test namespace and database
	err = db.Use(ctx, "test", "graph_test")
	require.NoError(t, err, "failed to use test database")

	return db
}

// cleanupTestDB removes test data
func cleanupTestDB(t *testing.T, db *surrealdb.DB) {
	ctx := context.Background()

	// Delete all test data
	_, err := db.Query(ctx, "DELETE host; DELETE port; DELETE service; DELETE vuln;", nil)
	if err != nil {
		t.Logf("cleanup error (non-fatal): %v", err)
	}

	db.Close(ctx)
}

// seedTestData creates test data in the database
func seedTestData(t *testing.T, db *surrealdb.DB) {
	ctx := context.Background()

	// Create test hosts
	queries := []string{
		`CREATE host:test1 SET ip = "192.168.1.1", asn = 15169, city = "Paris", region = "Ile-de-France", country = "France", last_seen = time::now(), first_seen = time::now() - 7d;`,
		`CREATE host:test2 SET ip = "192.168.1.2", asn = 15169, city = "Paris", region = "Ile-de-France", country = "France", last_seen = time::now(), first_seen = time::now() - 14d;`,
		`CREATE host:test3 SET ip = "10.0.0.1", asn = 8075, city = "London", region = "England", country = "UK", last_seen = time::now(), first_seen = time::now() - 1d;`,

		// Create test ports
		`CREATE port:test1_80 SET number = 80, protocol = "tcp", state = "open";`,
		`CREATE port:test1_443 SET number = 443, protocol = "tcp", state = "open";`,
		`CREATE port:test2_22 SET number = 22, protocol = "tcp", state = "open";`,
		`CREATE port:test3_6379 SET number = 6379, protocol = "tcp", state = "open";`,

		// Create test services
		`CREATE service:nginx SET name = "http", product = "nginx", version = "1.25.1", cpe = "cpe:2.3:a:nginx:nginx:1.25.1";`,
		`CREATE service:openssh SET name = "ssh", product = "openssh", version = "8.2", cpe = "cpe:2.3:a:openbsd:openssh:8.2";`,
		`CREATE service:redis SET name = "redis", product = "redis", version = "7.0.0", cpe = "cpe:2.3:a:redis:redis:7.0.0";`,

		// Create test vulnerabilities
		`CREATE vuln:cve_2023_1234 SET cve = "CVE-2023-1234", title = "Test Vulnerability", cvss = 9.8;`,
		`CREATE vuln:cve_2023_5678 SET cve = "CVE-2023-5678", title = "Redis Vulnerability", cvss = 7.5;`,

		// Create HAS edges (host -> port)
		`RELATE host:test1->HAS->port:test1_80;`,
		`RELATE host:test1->HAS->port:test1_443;`,
		`RELATE host:test2->HAS->port:test2_22;`,
		`RELATE host:test3->HAS->port:test3_6379;`,

		// Create RUNS edges (port -> service)
		`RELATE port:test1_80->RUNS->service:nginx;`,
		`RELATE port:test1_443->RUNS->service:nginx;`,
		`RELATE port:test2_22->RUNS->service:openssh;`,
		`RELATE port:test3_6379->RUNS->service:redis;`,

		// Create AFFECTED_BY edges (service -> vuln)
		`RELATE service:nginx->AFFECTED_BY->vuln:cve_2023_1234;`,
		`RELATE service:redis->AFFECTED_BY->vuln:cve_2023_5678;`,
	}

	for _, query := range queries {
		_, err := db.Query(ctx, query, nil)
		require.NoError(t, err, "failed to seed test data: %s", query)
	}
}

func TestGraphQueryExecutor_QueryByASN(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	tests := []struct {
		name        string
		asn         int
		limit       int
		offset      int
		wantCount   int
		wantTotal   int
		wantHasMore bool
	}{
		{
			name:        "query ASN with multiple hosts",
			asn:         15169,
			limit:       10,
			offset:      0,
			wantCount:   2,
			wantTotal:   2,
			wantHasMore: false,
		},
		{
			name:        "query ASN with single host",
			asn:         8075,
			limit:       10,
			offset:      0,
			wantCount:   1,
			wantTotal:   1,
			wantHasMore: false,
		},
		{
			name:        "query ASN with pagination",
			asn:         15169,
			limit:       1,
			offset:      0,
			wantCount:   1,
			wantTotal:   2,
			wantHasMore: true,
		},
		{
			name:        "query non-existent ASN",
			asn:         99999,
			limit:       10,
			offset:      0,
			wantCount:   0,
			wantTotal:   0,
			wantHasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := models.GraphQueryRequest{
				QueryType: models.QueryByASN,
				ASN:       &tt.asn,
				Limit:     tt.limit,
				Offset:    tt.offset,
			}

			resp, err := executor.ExecuteGraphQuery(ctx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCount, len(resp.Results))
			assert.Equal(t, tt.wantTotal, resp.Pagination.Total)
			assert.Equal(t, tt.wantHasMore, resp.Pagination.HasMore)

			// Verify all results have the correct ASN
			for _, host := range resp.Results {
				assert.Equal(t, tt.asn, host.ASN)
			}
		})
	}
}

func TestGraphQueryExecutor_QueryByLocation(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	tests := []struct {
		name        string
		city        string
		region      string
		country     string
		limit       int
		offset      int
		wantCount   int
		wantTotal   int
		wantHasMore bool
	}{
		{
			name:        "query by city",
			city:        "Paris",
			limit:       10,
			offset:      0,
			wantCount:   2,
			wantTotal:   2,
			wantHasMore: false,
		},
		{
			name:        "query by region",
			region:      "England",
			limit:       10,
			offset:      0,
			wantCount:   1,
			wantTotal:   1,
			wantHasMore: false,
		},
		{
			name:        "query by country",
			country:     "France",
			limit:       10,
			offset:      0,
			wantCount:   2,
			wantTotal:   2,
			wantHasMore: false,
		},
		{
			name:        "query with pagination",
			city:        "Paris",
			limit:       1,
			offset:      0,
			wantCount:   1,
			wantTotal:   2,
			wantHasMore: true,
		},
		{
			name:        "query non-existent city",
			city:        "Tokyo",
			limit:       10,
			offset:      0,
			wantCount:   0,
			wantTotal:   0,
			wantHasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := models.GraphQueryRequest{
				QueryType: models.QueryByLocation,
				City:      tt.city,
				Region:    tt.region,
				Country:   tt.country,
				Limit:     tt.limit,
				Offset:    tt.offset,
			}

			resp, err := executor.ExecuteGraphQuery(ctx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCount, len(resp.Results))
			assert.Equal(t, tt.wantTotal, resp.Pagination.Total)
			assert.Equal(t, tt.wantHasMore, resp.Pagination.HasMore)

			// Verify all results have the correct location
			for _, host := range resp.Results {
				if tt.city != "" {
					assert.Equal(t, tt.city, host.City)
				}
				if tt.region != "" {
					assert.Equal(t, tt.region, host.Region)
				}
				if tt.country != "" {
					assert.Equal(t, tt.country, host.Country)
				}
			}
		})
	}
}

func TestGraphQueryExecutor_QueryByVuln(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	tests := []struct {
		name        string
		cve         string
		limit       int
		offset      int
		wantMinimum int // Minimum expected count
		wantHasMore bool
	}{
		{
			name:        "query CVE affecting nginx",
			cve:         "CVE-2023-1234",
			limit:       10,
			offset:      0,
			wantMinimum: 1, // At least host:test1 has nginx
			wantHasMore: false,
		},
		{
			name:        "query CVE affecting redis",
			cve:         "CVE-2023-5678",
			limit:       10,
			offset:      0,
			wantMinimum: 1, // At least host:test3 has redis
			wantHasMore: false,
		},
		{
			name:        "query non-existent CVE",
			cve:         "CVE-9999-9999",
			limit:       10,
			offset:      0,
			wantMinimum: 0,
			wantHasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := models.GraphQueryRequest{
				QueryType: models.QueryByVuln,
				CVE:       tt.cve,
				Limit:     tt.limit,
				Offset:    tt.offset,
			}

			resp, err := executor.ExecuteGraphQuery(ctx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.GreaterOrEqual(t, len(resp.Results), tt.wantMinimum)
			assert.Equal(t, tt.wantHasMore, resp.Pagination.HasMore)

			// Verify results have IP addresses
			for _, host := range resp.Results {
				assert.NotEmpty(t, host.IP)
			}
		})
	}
}

func TestGraphQueryExecutor_QueryByService(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	tests := []struct {
		name        string
		product     string
		service     string
		limit       int
		offset      int
		wantMinimum int
		wantHasMore bool
	}{
		{
			name:        "query by product nginx",
			product:     "nginx",
			limit:       10,
			offset:      0,
			wantMinimum: 1, // At least host:test1 has nginx
			wantHasMore: false,
		},
		{
			name:        "query by product redis",
			product:     "redis",
			limit:       10,
			offset:      0,
			wantMinimum: 1, // At least host:test3 has redis
			wantHasMore: false,
		},
		{
			name:        "query by service name",
			service:     "ssh",
			limit:       10,
			offset:      0,
			wantMinimum: 1, // At least host:test2 has ssh
			wantHasMore: false,
		},
		{
			name:        "query non-existent product",
			product:     "apache",
			limit:       10,
			offset:      0,
			wantMinimum: 0,
			wantHasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := models.GraphQueryRequest{
				QueryType: models.QueryByService,
				Product:   tt.product,
				Service:   tt.service,
				Limit:     tt.limit,
				Offset:    tt.offset,
			}

			resp, err := executor.ExecuteGraphQuery(ctx, req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.GreaterOrEqual(t, len(resp.Results), tt.wantMinimum)
			assert.Equal(t, tt.wantHasMore, resp.Pagination.HasMore)

			// Verify results have IP addresses
			for _, host := range resp.Results {
				assert.NotEmpty(t, host.IP)
			}
		})
	}
}

func TestGraphQueryExecutor_QueryTimeout(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	asn := 15169
	req := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     10,
		Offset:    0,
	}

	// Should fail due to timeout
	_, err := executor.ExecuteGraphQuery(ctx, req)
	assert.Error(t, err)
}

func TestGraphQueryExecutor_Pagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	ctx := context.Background()
	asn := 15169

	// First page
	req1 := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     1,
		Offset:    0,
	}

	resp1, err := executor.ExecuteGraphQuery(ctx, req1)
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp1.Results))
	assert.True(t, resp1.Pagination.HasMore)
	assert.Equal(t, 1, resp1.Pagination.NextOffset)

	// Second page
	req2 := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     1,
		Offset:    resp1.Pagination.NextOffset,
	}

	resp2, err := executor.ExecuteGraphQuery(ctx, req2)
	require.NoError(t, err)
	assert.Equal(t, 1, len(resp2.Results))
	assert.False(t, resp2.Pagination.HasMore)

	// Verify we got different hosts
	assert.NotEqual(t, resp1.Results[0].IP, resp2.Results[0].IP)
}

func TestGraphQueryExecutor_ValidationErrors(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	ctx := context.Background()

	tests := []struct {
		name    string
		req     models.GraphQueryRequest
		wantErr error
	}{
		{
			name: "missing ASN for by_asn query",
			req: models.GraphQueryRequest{
				QueryType: models.QueryByASN,
				Limit:     10,
			},
			wantErr: models.ErrMissingASN,
		},
		{
			name: "missing location for by_location query",
			req: models.GraphQueryRequest{
				QueryType: models.QueryByLocation,
				Limit:     10,
			},
			wantErr: models.ErrMissingLocation,
		},
		{
			name: "missing CVE for by_vuln query",
			req: models.GraphQueryRequest{
				QueryType: models.QueryByVuln,
				Limit:     10,
			},
			wantErr: models.ErrMissingCVE,
		},
		{
			name: "missing service for by_service query",
			req: models.GraphQueryRequest{
				QueryType: models.QueryByService,
				Limit:     10,
			},
			wantErr: models.ErrMissingService,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.ExecuteGraphQuery(ctx, tt.req)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestGraphQueryExecutor_DefaultsAndLimits(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	ctx := context.Background()
	asn := 15169

	tests := []struct {
		name       string
		inputLimit int
		wantLimit  int
	}{
		{
			name:       "default limit applied",
			inputLimit: 0,
			wantLimit:  models.DefaultLimit,
		},
		{
			name:       "max limit enforced",
			inputLimit: 5000,
			wantLimit:  models.MaxLimit,
		},
		{
			name:       "valid limit preserved",
			inputLimit: 50,
			wantLimit:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.GraphQueryRequest{
				QueryType: models.QueryByASN,
				ASN:       &asn,
				Limit:     tt.inputLimit,
			}

			resp, err := executor.ExecuteGraphQuery(ctx, req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, resp.Pagination.Limit)
		})
	}
}

func TestGraphQueryExecutor_QueryTime(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	seedTestData(t, db)

	logger := zaptest.NewLogger(t)
	executor := NewGraphQueryExecutor(db, logger)

	ctx := context.Background()
	asn := 15169
	req := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     10,
	}

	resp, err := executor.ExecuteGraphQuery(ctx, req)
	require.NoError(t, err)
	assert.Greater(t, resp.QueryTime, 0.0)
	assert.Less(t, resp.QueryTime, 5000.0) // Should be less than 5 seconds
}
