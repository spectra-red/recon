//go:build integration
// +build integration

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// TestQueryHandlerIntegration tests the full query handler with a real database
// Run with: go test -tags=integration -v ./internal/api/handlers
func TestQueryHandlerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create database connection
	ctx := context.Background()
	db, err := setupTestDatabase(ctx, t)
	require.NoError(t, err)
	defer db.Close(ctx)

	// Insert test data
	testIP := "192.168.1.100"
	err = insertTestHost(ctx, db, testIP)
	require.NoError(t, err)

	// Test cases
	tests := []struct {
		name           string
		ip             string
		depth          string
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "query existing host - default depth",
			ip:             testIP,
			depth:          "",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, testIP, response["ip"])
				assert.NotEmpty(t, response["last_seen"])
			},
		},
		{
			name:           "query existing host - depth 0",
			ip:             testIP,
			depth:          "0",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, testIP, response["ip"])
				// At depth 0, should not have ports
				_, hasports := response["ports"]
				assert.False(t, hasports || response["ports"] == nil)
			},
		},
		{
			name:           "query existing host - depth 1",
			ip:             testIP,
			depth:          "1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, testIP, response["ip"])
			},
		},
		{
			name:           "query non-existent host",
			ip:             "10.0.0.1",
			depth:          "",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response["message"], "not found")
			},
		},
		{
			name:           "invalid depth parameter",
			ip:             testIP,
			depth:          "invalid",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response["message"], "invalid depth")
			},
		},
		{
			name:           "depth out of range",
			ip:             testIP,
			depth:          "10",
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response["message"], "between 0 and 5")
			},
		},
	}

	logger := zap.NewNop()
	handler := QueryHandler(logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			url := fmt.Sprintf("/v1/query/host/%s", tt.ip)
			if tt.depth != "" {
				url += "?depth=" + tt.depth
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("ip", tt.ip)
			req = req.WithContext(chi.NewRouteContext().WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute request
			handler(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and check response
			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

// setupTestDatabase creates a connection to a test SurrealDB instance
func setupTestDatabase(ctx context.Context, t *testing.T) (*surrealdb.DB, error) {
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Sign in
	if _, err := db.SignIn(ctx, map[string]interface{}{
		"user": "root",
		"pass": "root",
	}); err != nil {
		db.Close(ctx)
		return nil, fmt.Errorf("failed to sign in: %w", err)
	}

	// Use test namespace and database
	testNS := "test_spectra"
	testDB := "test_intel"
	if err := db.Use(ctx, testNS, testDB); err != nil {
		db.Close(ctx)
		return nil, fmt.Errorf("failed to use database: %w", err)
	}

	return db, nil
}

// insertTestHost inserts a test host record into the database
func insertTestHost(ctx context.Context, db *surrealdb.DB, ip string) error {
	query := `
		CREATE host:$id CONTENT {
			ip: $ip,
			asn: 15169,
			city: "San Francisco",
			region: "California",
			country: "US",
			first_seen: $now,
			last_seen: $now
		}
	`

	_, err := surrealdb.Query[map[string]interface{}](ctx, db, query, map[string]interface{}{
		"id":  ip,
		"ip":  ip,
		"now": time.Now(),
	})

	return err
}

// BenchmarkQueryHandler benchmarks the query handler performance
func BenchmarkQueryHandler(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	ctx := context.Background()
	db, err := setupTestDatabase(ctx, &testing.T{})
	if err != nil {
		b.Fatalf("failed to setup database: %v", err)
	}
	defer db.Close(ctx)

	testIP := "192.168.1.200"
	err = insertTestHost(ctx, db, testIP)
	if err != nil {
		b.Fatalf("failed to insert test host: %v", err)
	}

	logger := zap.NewNop()
	handler := QueryHandler(logger)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/v1/query/host/"+testIP, nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("ip", testIP)
		req = req.WithContext(chi.NewRouteContext().WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

// TestQueryResponse_DepthLevels tests response structure at different depth levels
func TestQueryResponse_DepthLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	db, err := setupTestDatabase(ctx, t)
	require.NoError(t, err)
	defer db.Close(ctx)

	testIP := "192.168.1.101"
	err = insertTestHost(ctx, db, testIP)
	require.NoError(t, err)

	depths := []int{0, 1, 2, 3}

	logger := zap.NewNop()
	handler := QueryHandler(logger)

	for _, depth := range depths {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			url := fmt.Sprintf("/v1/query/host/%s?depth=%d", testIP, depth)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("ip", testIP)
			req = req.WithContext(chi.NewRouteContext().WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.HostQueryResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, testIP, response.IP)
			assert.NotZero(t, response.LastSeen)

			// Verify depth-specific fields
			if depth >= 1 {
				// Ports should be included (may be empty)
				assert.NotNil(t, response.Ports)
			}
			if depth >= 2 {
				// Services should be included (may be empty)
				assert.NotNil(t, response.Services)
			}
			if depth >= 3 {
				// Vulnerabilities should be included (may be empty)
				assert.NotNil(t, response.Vulns)
			}
		})
	}
}
