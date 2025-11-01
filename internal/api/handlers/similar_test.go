package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spectra-red/recon/internal/db"
	"github.com/spectra-red/recon/internal/embeddings"
	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockEmbeddingClient mocks the embedding client for testing
type MockEmbeddingClient struct {
	GenerateFunc func(ctx context.Context, query string) ([]float64, error)
}

func (m *MockEmbeddingClient) GenerateEmbedding(ctx context.Context, query string) ([]float64, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, query)
	}
	// Default: return a static embedding
	embedding := make([]float64, 1536)
	for i := 0; i < 1536; i++ {
		embedding[i] = float64(len(query)+i) / 1536.0
	}
	return embedding, nil
}

// MockVectorClient mocks the vector search client for testing
type MockVectorClient struct {
	SearchFunc func(ctx context.Context, params db.VectorSearchParams) ([]models.VulnResult, error)
}

func (m *MockVectorClient) VectorSearch(ctx context.Context, params db.VectorSearchParams) ([]models.VulnResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, params)
	}
	// Default: return empty results
	return []models.VulnResult{}, nil
}

func TestNewSimilarHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockEmbed := &MockEmbeddingClient{}
	mockVector := &MockVectorClient{}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
	assert.Equal(t, mockEmbed, handler.embeddingClient)
	assert.Equal(t, mockVector, handler.vectorClient)
}

func TestSimilarHandler_MethodNotAllowed(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := NewSimilarHandler(&MockEmbeddingClient{}, &MockVectorClient{}, logger)

	// Test GET request (should fail)
	req := httptest.NewRequest(http.MethodGet, "/v1/query/similar", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "method not allowed", errResp.Error)
	assert.Equal(t, "METHOD_NOT_ALLOWED", errResp.Code)
}

func TestSimilarHandler_InvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := NewSimilarHandler(&MockEmbeddingClient{}, &MockVectorClient{}, logger)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBufferString("{invalid json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid request body", errResp.Error)
}

func TestSimilarHandler_ValidationErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := NewSimilarHandler(&MockEmbeddingClient{}, &MockVectorClient{}, logger)

	tests := []struct {
		name        string
		request     models.SimilarRequest
		expectedErr string
	}{
		{
			name: "empty query",
			request: models.SimilarRequest{
				Query: "",
			},
			expectedErr: "query cannot be empty",
		},
		{
			name: "query too long",
			request: models.SimilarRequest{
				Query: string(make([]byte, 501)),
			},
			expectedErr: "query exceeds maximum length",
		},
		{
			name: "invalid K (negative)",
			request: models.SimilarRequest{
				Query: "test query",
				K:     ptr(-1),
			},
			expectedErr: "k must be greater than 0",
		},
		{
			name: "K too large",
			request: models.SimilarRequest{
				Query: "test query",
				K:     ptr(100),
			},
			expectedErr: "k exceeds maximum allowed value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errResp models.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errResp)
			require.NoError(t, err)
			assert.Contains(t, errResp.Details, tt.expectedErr)
		})
	}
}

func TestSimilarHandler_SuccessfulSearch(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockResults := []models.VulnResult{
		{
			CVEID:         "CVE-2021-12345",
			Title:         "Test Vulnerability 1",
			Summary:       "Test vulnerability description",
			CVSS:          9.8,
			CPE:           []string{"cpe:2.3:a:vendor:product:1.0"},
			PublishedDate: "2021-01-01T00:00:00Z",
			Score:         0.95,
		},
		{
			CVEID:         "CVE-2021-67890",
			Title:         "Test Vulnerability 2",
			Summary:       "Another test vulnerability",
			CVSS:          7.5,
			CPE:           []string{"cpe:2.3:a:vendor:product:2.0"},
			PublishedDate: "2021-02-01T00:00:00Z",
			Score:         0.87,
		},
	}

	mockEmbed := &MockEmbeddingClient{
		GenerateFunc: func(ctx context.Context, query string) ([]float64, error) {
			// Return static embedding
			embedding := make([]float64, 1536)
			for i := 0; i < 1536; i++ {
				embedding[i] = 0.001 * float64(i)
			}
			return embedding, nil
		},
	}

	mockVector := &MockVectorClient{
		SearchFunc: func(ctx context.Context, params db.VectorSearchParams) ([]models.VulnResult, error) {
			// Verify parameters
			assert.Equal(t, 1536, len(params.QueryEmbedding))
			assert.Equal(t, 10, params.K)
			return mockResults, nil
		},
	}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	// Create request
	reqBody := models.SimilarRequest{
		Query: "nginx remote code execution",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SimilarResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "nginx remote code execution", resp.Query)
	assert.Equal(t, 2, resp.Count)
	assert.Equal(t, len(mockResults), len(resp.Results))
	assert.Equal(t, "CVE-2021-12345", resp.Results[0].CVEID)
	assert.Equal(t, 0.95, resp.Results[0].Score)
	assert.NotEmpty(t, resp.Timestamp)
}

func TestSimilarHandler_CustomK(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEmbed := &MockEmbeddingClient{}
	mockVector := &MockVectorClient{
		SearchFunc: func(ctx context.Context, params db.VectorSearchParams) ([]models.VulnResult, error) {
			// Verify K parameter
			assert.Equal(t, 20, params.K)
			return []models.VulnResult{}, nil
		},
	}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	reqBody := models.SimilarRequest{
		Query: "test query",
		K:     ptr(20),
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSimilarHandler_NoResults(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEmbed := &MockEmbeddingClient{}
	mockVector := &MockVectorClient{
		SearchFunc: func(ctx context.Context, params db.VectorSearchParams) ([]models.VulnResult, error) {
			return nil, db.ErrNoResults
		},
	}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	reqBody := models.SimilarRequest{
		Query: "nonexistent vulnerability",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 200 with empty results
	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.SimilarResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Count)
	assert.Empty(t, resp.Results)
}

func TestSimilarHandler_EmbeddingServiceUnavailable(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEmbed := &MockEmbeddingClient{
		GenerateFunc: func(ctx context.Context, query string) ([]float64, error) {
			return nil, embeddings.ErrServiceUnavailable
		},
	}
	mockVector := &MockVectorClient{}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	reqBody := models.SimilarRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "embedding service is temporarily unavailable", errResp.Error)
	assert.Equal(t, "SERVICE_UNAVAILABLE", errResp.Code)
	assert.Contains(t, errResp.Details, "OpenAI API key")
}

func TestSimilarHandler_InvalidAPIKey(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEmbed := &MockEmbeddingClient{
		GenerateFunc: func(ctx context.Context, query string) ([]float64, error) {
			return nil, embeddings.ErrInvalidAPIKey
		},
	}
	mockVector := &MockVectorClient{}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	reqBody := models.SimilarRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 500 (configuration error)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "embedding service configuration error", errResp.Error)
	assert.Contains(t, errResp.Details, "not properly configured")
}

func TestSimilarHandler_DatabaseUnavailable(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEmbed := &MockEmbeddingClient{}
	mockVector := &MockVectorClient{
		SearchFunc: func(ctx context.Context, params db.VectorSearchParams) ([]models.VulnResult, error) {
			return nil, db.ErrDatabaseUnavailable
		},
	}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	reqBody := models.SimilarRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "database service is temporarily unavailable", errResp.Error)
	assert.Contains(t, errResp.Details, "vector search database")
}

func TestSimilarHandler_UnknownError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockEmbed := &MockEmbeddingClient{
		GenerateFunc: func(ctx context.Context, query string) ([]float64, error) {
			return nil, errors.New("unknown error")
		},
	}
	mockVector := &MockVectorClient{}

	handler := NewSimilarHandler(mockEmbed, mockVector, logger)

	reqBody := models.SimilarRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "internal server error", errResp.Error)
}

func TestSimilarHandlerFunc(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockEmbed := &MockEmbeddingClient{}
	mockVector := &MockVectorClient{}

	handlerFunc := SimilarHandlerFunc(mockEmbed, mockVector, logger)

	assert.NotNil(t, handlerFunc)

	// Test that it works as an http.HandlerFunc
	reqBody := models.SimilarRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/query/similar", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handlerFunc(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Helper function to create int pointer
func ptr(i int) *int {
	return &i
}
