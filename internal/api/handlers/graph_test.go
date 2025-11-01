package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap/zaptest"
)

// setupTestGraphDB creates a test database connection and seeds data
func setupTestGraphDB(t *testing.T) *surrealdb.DB {
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
	err = db.Use(ctx, "test", "graph_handler_test")
	require.NoError(t, err, "failed to use test database")

	// Seed test data
	seedTestGraphData(t, db)

	return db
}

// cleanupTestGraphDB removes test data
func cleanupTestGraphDB(t *testing.T, db *surrealdb.DB) {
	ctx := context.Background()

	// Delete all test data
	_, err := db.Query(ctx, "DELETE host; DELETE port; DELETE service; DELETE vuln;", nil)
	if err != nil {
		t.Logf("cleanup error (non-fatal): %v", err)
	}

	db.Close(ctx)
}

// seedTestGraphData creates test data in the database
func seedTestGraphData(t *testing.T, db *surrealdb.DB) {
	ctx := context.Background()

	queries := []string{
		`CREATE host:test1 SET ip = "192.168.1.1", asn = 15169, city = "Paris", region = "Ile-de-France", country = "France", last_seen = time::now(), first_seen = time::now() - 7d;`,
		`CREATE host:test2 SET ip = "192.168.1.2", asn = 15169, city = "Paris", region = "Ile-de-France", country = "France", last_seen = time::now(), first_seen = time::now() - 14d;`,
		`CREATE host:test3 SET ip = "10.0.0.1", asn = 8075, city = "London", region = "England", country = "UK", last_seen = time::now(), first_seen = time::now() - 1d;`,
		`CREATE port:test1_80 SET number = 80, protocol = "tcp", state = "open";`,
		`CREATE port:test2_22 SET number = 22, protocol = "tcp", state = "open";`,
		`CREATE port:test3_6379 SET number = 6379, protocol = "tcp", state = "open";`,
		`CREATE service:nginx SET name = "http", product = "nginx", version = "1.25.1";`,
		`CREATE service:openssh SET name = "ssh", product = "openssh", version = "8.2";`,
		`CREATE service:redis SET name = "redis", product = "redis", version = "7.0.0";`,
		`CREATE vuln:cve_2023_1234 SET cve = "CVE-2023-1234", title = "Test Vulnerability", cvss = 9.8;`,
		`RELATE host:test1->HAS->port:test1_80;`,
		`RELATE host:test2->HAS->port:test2_22;`,
		`RELATE host:test3->HAS->port:test3_6379;`,
		`RELATE port:test1_80->RUNS->service:nginx;`,
		`RELATE port:test2_22->RUNS->service:openssh;`,
		`RELATE port:test3_6379->RUNS->service:redis;`,
		`RELATE service:nginx->AFFECTED_BY->vuln:cve_2023_1234;`,
	}

	for _, query := range queries {
		_, err := db.Query(ctx, query, nil)
		require.NoError(t, err, "failed to seed test data: %s", query)
	}
}

func TestGraphQueryHandler_HandleGraphQuery_ByASN(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	// Prepare request
	asn := 15169
	reqBody := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     10,
		Offset:    0,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute
	handler.HandleGraphQuery(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.GraphQueryResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(resp.Results), 1)
	assert.GreaterOrEqual(t, resp.Pagination.Total, 1)
	assert.Greater(t, resp.QueryTime, 0.0)
}

func TestGraphQueryHandler_HandleGraphQuery_ByLocation(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		reqBody   models.GraphQueryRequest
		wantCount int
	}{
		{
			name: "query by city",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByLocation,
				City:      "Paris",
				Limit:     10,
			},
			wantCount: 2,
		},
		{
			name: "query by region",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByLocation,
				Region:    "England",
				Limit:     10,
			},
			wantCount: 1,
		},
		{
			name: "query by country",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByLocation,
				Country:   "France",
				Limit:     10,
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.HandleGraphQuery(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp models.GraphQueryResponse
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.GreaterOrEqual(t, len(resp.Results), tt.wantCount)
		})
	}
}

func TestGraphQueryHandler_HandleGraphQuery_ByVuln(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	// Prepare request
	reqBody := models.GraphQueryRequest{
		QueryType: models.QueryByVuln,
		CVE:       "CVE-2023-1234",
		Limit:     10,
		Offset:    0,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute
	handler.HandleGraphQuery(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.GraphQueryResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Should find at least the host with nginx
	assert.GreaterOrEqual(t, len(resp.Results), 1)
}

func TestGraphQueryHandler_HandleGraphQuery_ByService(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		product   string
		service   string
		wantCount int
	}{
		{
			name:      "query by product nginx",
			product:   "nginx",
			wantCount: 1,
		},
		{
			name:      "query by product redis",
			product:   "redis",
			wantCount: 1,
		},
		{
			name:      "query by service name",
			service:   "ssh",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.GraphQueryRequest{
				QueryType: models.QueryByService,
				Product:   tt.product,
				Service:   tt.service,
				Limit:     10,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.HandleGraphQuery(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp models.GraphQueryResponse
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.GreaterOrEqual(t, len(resp.Results), tt.wantCount)
		})
	}
}

func TestGraphQueryHandler_HandleGraphQuery_Pagination(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	// First page
	asn := 15169
	reqBody1 := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     1,
		Offset:    0,
	}

	body1, err := json.Marshal(reqBody1)
	require.NoError(t, err)

	req1 := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	handler.HandleGraphQuery(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	var resp1 models.GraphQueryResponse
	err = json.NewDecoder(w1.Body).Decode(&resp1)
	require.NoError(t, err)

	assert.Equal(t, 1, len(resp1.Results))
	assert.True(t, resp1.Pagination.HasMore)
	assert.Equal(t, 1, resp1.Pagination.NextOffset)

	// Second page
	reqBody2 := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     1,
		Offset:    resp1.Pagination.NextOffset,
	}

	body2, err := json.Marshal(reqBody2)
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	handler.HandleGraphQuery(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var resp2 models.GraphQueryResponse
	err = json.NewDecoder(w2.Body).Decode(&resp2)
	require.NoError(t, err)

	assert.Equal(t, 1, len(resp2.Results))
}

func TestGraphQueryHandler_HandleGraphQuery_ValidationErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	tests := []struct {
		name       string
		reqBody    models.GraphQueryRequest
		wantStatus int
	}{
		{
			name: "missing ASN for by_asn query",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByASN,
				Limit:     10,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing location for by_location query",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByLocation,
				Limit:     10,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing CVE for by_vuln query",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByVuln,
				Limit:     10,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing service for by_service query",
			reqBody: models.GraphQueryRequest{
				QueryType: models.QueryByService,
				Limit:     10,
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.HandleGraphQuery(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var errResp ErrorResponse
			err = json.NewDecoder(w.Body).Decode(&errResp)
			require.NoError(t, err)
			assert.NotEmpty(t, errResp.Message)
		})
	}
}

func TestGraphQueryHandler_HandleGraphQuery_InvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleGraphQuery(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Message, "invalid request body")
}

func TestGraphQueryHandler_HandleGraphQuery_Timeout(t *testing.T) {
	// This test is challenging to implement without mocking
	// In a real scenario, we would mock the executor to simulate a slow query
	t.Skip("Timeout testing requires mocking or extremely slow queries")
}

func TestGraphQueryHandler_HandleGraphQuery_DefaultsAndLimits(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

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
			asn := 15169
			reqBody := models.GraphQueryRequest{
				QueryType: models.QueryByASN,
				ASN:       &asn,
				Limit:     tt.inputLimit,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.HandleGraphQuery(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp models.GraphQueryResponse
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tt.wantLimit, resp.Pagination.Limit)
		})
	}
}

func TestGraphQueryHandler_HandleGraphQuery_QueryTimeReported(t *testing.T) {
	// Setup
	db := setupTestGraphDB(t)
	defer cleanupTestGraphDB(t, db)

	logger := zaptest.NewLogger(t)
	handler, err := NewGraphQueryHandler(logger)
	require.NoError(t, err)

	asn := 15169
	reqBody := models.GraphQueryRequest{
		QueryType: models.QueryByASN,
		ASN:       &asn,
		Limit:     10,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/query/graph", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	start := time.Now()
	handler.HandleGraphQuery(w, req)
	elapsed := time.Since(start).Milliseconds()

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.GraphQueryResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Query time should be reasonable
	assert.Greater(t, resp.QueryTime, 0.0)
	assert.Less(t, resp.QueryTime, float64(elapsed)+100) // Within 100ms of wall clock time
}
