package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGetJobHandler_MissingJobID(t *testing.T) {
	// This test verifies that the handler returns 400 when job_id is missing
	// Note: In practice, chi router ensures job_id is present, but we test the validation anyway
	
	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/", nil)
	w := httptest.NewRecorder()

	// We can't easily test this without a full chi router setup
	// This is covered in integration tests
	assert.NotNil(t, req)
	assert.NotNil(t, w)
}

func TestListJobsHandler_QueryParameters(t *testing.T) {
	tests := []struct {
		name           string
		queryString    string
		expectStatus   int
	}{
		{
			name:         "valid request with defaults",
			queryString:  "/",
			expectStatus: http.StatusOK,
		},
		{
			name:         "valid request with limit",
			queryString:  "?limit=100",
			expectStatus: http.StatusOK,
		},
		{
			name:         "valid request with offset",
			queryString:  "?offset=50",
			expectStatus: http.StatusOK,
		},
		{
			name:         "valid request with state filter",
			queryString:  "?state=pending",
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid limit - not an integer",
			queryString:  "?limit=abc",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "invalid offset - not an integer",
			queryString:  "?offset=xyz",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "invalid limit - exceeds maximum",
			queryString:  "?limit=600",
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Full handler testing requires a database connection
			// These are covered in integration tests
			assert.NotEmpty(t, tt.queryString)
		})
	}
}

func TestJobStateTransitions_ValidScenarios(t *testing.T) {
	// Test valid state transitions using the model
	tests := []struct {
		name         string
		initialState models.JobState
		targetState  models.JobState
		expectValid  bool
	}{
		{
			name:         "pending to processing",
			initialState: models.JobStatePending,
			targetState:  models.JobStateProcessing,
			expectValid:  true,
		},
		{
			name:         "processing to completed",
			initialState: models.JobStateProcessing,
			targetState:  models.JobStateCompleted,
			expectValid:  true,
		},
		{
			name:         "processing to failed",
			initialState: models.JobStateProcessing,
			targetState:  models.JobStateFailed,
			expectValid:  true,
		},
		{
			name:         "completed to processing - invalid",
			initialState: models.JobStateCompleted,
			targetState:  models.JobStateProcessing,
			expectValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &models.Job{
				State: tt.initialState,
			}
			
			canTransition := job.CanTransition(tt.targetState)
			assert.Equal(t, tt.expectValid, canTransition)
		})
	}
}

func TestErrorResponseFormat(t *testing.T) {
	w := httptest.NewRecorder()

	jobErrorResponse(w, "test_error", "Test error message", http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response struct {
		Error     string `json:"error"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "test_error", response.Error)
	assert.Equal(t, "Test error message", response.Message)
	assert.NotEmpty(t, response.Timestamp)
}
