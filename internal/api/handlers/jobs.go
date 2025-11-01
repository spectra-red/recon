package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spectra-red/recon/internal/db"
	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// GetJobHandler creates an HTTP handler for GET /v1/jobs/{job_id}
// Returns the current status and metadata for a job
func GetJobHandler(dbClient *surrealdb.DB, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Extract job_id from URL path
		jobID := chi.URLParam(r, "job_id")
		if jobID == "" {
			logger.Warn("missing job_id parameter")
			jobErrorResponse(w, "missing_parameter", "job_id is required", http.StatusBadRequest)
			return
		}

		// Query the job from database
		job, err := db.GetJob(ctx, dbClient, logger, jobID)
		if err != nil {
			logger.Error("failed to get job",
				zap.Error(err),
				zap.String("job_id", jobID))
			jobErrorResponse(w, "internal_error", "Failed to retrieve job", http.StatusInternalServerError)
			return
		}

		// Return 404 if job not found
		if job == nil {
			logger.Debug("job not found",
				zap.String("job_id", jobID))
			jobErrorResponse(w, "not_found", "Job not found", http.StatusNotFound)
			return
		}

		// Return job data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(job); err != nil {
			logger.Error("failed to encode job response",
				zap.Error(err),
				zap.String("job_id", jobID))
		}

		logger.Debug("job retrieved successfully",
			zap.String("job_id", jobID),
			zap.String("state", job.State.String()))
	}
}

// ListJobsHandler creates an HTTP handler for GET /v1/jobs
// Returns a paginated list of jobs with optional filters
func ListJobsHandler(dbClient *surrealdb.DB, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Parse query parameters
		req := models.JobListRequest{
			Limit:     50,
			Offset:    0,
			OrderBy:   "created_at",
			OrderDesc: true,
		}

		// Parse limit
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				logger.Warn("invalid limit parameter",
					zap.String("limit", limitStr))
				jobErrorResponse(w, "invalid_parameter", "limit must be an integer", http.StatusBadRequest)
				return
			}
			req.Limit = limit
		}

		// Parse offset
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				logger.Warn("invalid offset parameter",
					zap.String("offset", offsetStr))
				jobErrorResponse(w, "invalid_parameter", "offset must be an integer", http.StatusBadRequest)
				return
			}
			req.Offset = offset
		}

		// Parse scanner_key filter
		if scannerKey := r.URL.Query().Get("scanner_key"); scannerKey != "" {
			req.ScannerKey = &scannerKey
		}

		// Parse state filter
		if stateStr := r.URL.Query().Get("state"); stateStr != "" {
			state := models.JobState(stateStr)
			req.State = &state
		}

		// Parse order_by
		if orderBy := r.URL.Query().Get("order_by"); orderBy != "" {
			req.OrderBy = orderBy
		}

		// Parse order direction
		if orderDesc := r.URL.Query().Get("order_desc"); orderDesc != "" {
			req.OrderDesc = orderDesc != "false"
		}

		// Validate request
		if err := req.Validate(); err != nil {
			logger.Warn("invalid list request",
				zap.Error(err))
			jobErrorResponse(w, "invalid_parameter", err.Error(), http.StatusBadRequest)
			return
		}

		// Query jobs from database
		response, err := db.ListJobs(ctx, dbClient, logger, req)
		if err != nil {
			logger.Error("failed to list jobs",
				zap.Error(err))
			jobErrorResponse(w, "internal_error", "Failed to list jobs", http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("failed to encode list response",
				zap.Error(err))
		}

		logger.Debug("jobs listed successfully",
			zap.Int("count", len(response.Jobs)),
			zap.Int("total", response.Total))
	}
}

// jobErrorResponse writes a consistent error response for job endpoints
func jobErrorResponse(w http.ResponseWriter, errorCode, message string, statusCode int) {
	response := struct {
		Error     string `json:"error"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}{
		Error:     errorCode,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Best effort encoding - ignore errors at this point
	_ = json.NewEncoder(w).Encode(response)
}
