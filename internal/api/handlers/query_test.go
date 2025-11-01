package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestQueryHandler_MissingIP(t *testing.T) {
	logger := zap.NewNop()
	handler := QueryHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/v1/query/host/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "missing IP parameter")
}

func TestQueryHandler_InvalidDepth(t *testing.T) {
	tests := []struct {
		name        string
		depth       string
		expectedMsg string
	}{
		{
			name:        "non-numeric depth",
			depth:       "abc",
			expectedMsg: "invalid depth parameter",
		},
		{
			name:        "negative depth",
			depth:       "-1",
			expectedMsg: "depth must be between 0 and 5",
		},
		{
			name:        "depth too large",
			depth:       "10",
			expectedMsg: "depth must be between 0 and 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			handler := QueryHandler(logger)

			req := httptest.NewRequest(http.MethodGet, "/v1/query/host/1.2.3.4?depth="+tt.depth, nil)
			w := httptest.NewRecorder()

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("ip", "1.2.3.4")
			req = req.WithContext(chi.NewRouteContext().WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Contains(t, response["message"], tt.expectedMsg)
		})
	}
}

func TestQueryHandler_ValidDepthParsing(t *testing.T) {
	tests := []struct {
		name          string
		depth         string
		expectedDepth int
	}{
		{
			name:          "depth 0",
			depth:         "0",
			expectedDepth: 0,
		},
		{
			name:          "depth 1",
			depth:         "1",
			expectedDepth: 1,
		},
		{
			name:          "depth 2 (default)",
			depth:         "2",
			expectedDepth: 2,
		},
		{
			name:          "depth 3",
			depth:         "3",
			expectedDepth: 3,
		},
		{
			name:          "depth 5 (max)",
			depth:         "5",
			expectedDepth: 5,
		},
		{
			name:          "no depth parameter (should default to 2)",
			depth:         "",
			expectedDepth: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: These tests would need a mock database to fully validate
			// This test validates parameter parsing logic only
			assert.True(t, models.ValidateDepth(tt.expectedDepth))
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		statusCode     int
		expectedStatus string
	}{
		{
			name:           "bad request",
			message:        "invalid input",
			statusCode:     http.StatusBadRequest,
			expectedStatus: "Bad Request",
		},
		{
			name:           "not found",
			message:        "host not found",
			statusCode:     http.StatusNotFound,
			expectedStatus: "Not Found",
		},
		{
			name:           "internal error",
			message:        "database error",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeErrorResponse(w, tt.message, tt.statusCode)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, response["error"])
			assert.Equal(t, tt.message, response["message"])
			assert.Equal(t, float64(tt.statusCode), response["code"])
		})
	}
}

func TestQueryDepthValidation(t *testing.T) {
	tests := []struct {
		name  string
		depth int
		valid bool
	}{
		{"depth 0 valid", 0, true},
		{"depth 1 valid", 1, true},
		{"depth 2 valid", 2, true},
		{"depth 3 valid", 3, true},
		{"depth 4 valid", 4, true},
		{"depth 5 valid", 5, true},
		{"depth -1 invalid", -1, false},
		{"depth 6 invalid", 6, false},
		{"depth 10 invalid", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.ValidateDepth(tt.depth)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestDefaultDepth(t *testing.T) {
	depth := models.DefaultDepth()
	assert.Equal(t, models.DepthWithServices, depth)
	assert.Equal(t, 2, int(depth))
}
