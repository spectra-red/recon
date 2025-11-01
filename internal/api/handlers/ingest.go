package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/spectra-red/recon/internal/auth"
	"github.com/spectra-red/recon/internal/db"
	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// IngestRequest represents the incoming scan submission request
type IngestRequest struct {
	auth.ScanEnvelope
}

// IngestResponse represents the response returned after accepting a scan
type IngestResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// IngestHandler creates an HTTP handler for the /v1/mesh/ingest endpoint
// It validates Ed25519 signatures, creates a job record, and triggers the Restate workflow
func IngestHandler(logger *zap.Logger, dbClient *surrealdb.DB, restateURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Parse request body
		body, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024)) // 10MB max
		if err != nil {
			logger.Warn("failed to read request body",
				zap.Error(err))
			ingestErrorResponse(w, "invalid_request", "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req IngestRequest
		if err := json.Unmarshal(body, &req); err != nil {
			logger.Warn("failed to parse request JSON",
				zap.Error(err))
			ingestErrorResponse(w, "invalid_json", "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// Validate Ed25519 signature
		if err := auth.VerifyEnvelope(req.ScanEnvelope); err != nil {
			logger.Warn("signature verification failed",
				zap.Error(err),
				zap.String("public_key", maskPublicKey(req.PublicKey)))
			ingestErrorResponse(w, "invalid_signature", "Signature verification failed", http.StatusUnauthorized)
			return
		}

		// Create job record in database
		job, err := db.CreateJob(ctx, dbClient, logger, req.PublicKey)
		if err != nil {
			logger.Error("failed to create job",
				zap.Error(err),
				zap.String("public_key", maskPublicKey(req.PublicKey)))
			ingestErrorResponse(w, "internal_error", "Failed to create job", http.StatusInternalServerError)
			return
		}

		logger.Info("scan received, job created",
			zap.String("job_id", job.ID),
			zap.String("public_key", maskPublicKey(req.PublicKey)),
			zap.Int64("timestamp", req.Timestamp),
			zap.Int("data_size", len(req.Data)))

		// Trigger Restate workflow asynchronously
		workflowReq := models.IngestWorkflowRequest{
			JobID:      job.ID,
			ScannerKey: req.PublicKey,
			ScanData:   req.Data,
		}

		// Send to Restate (fire-and-forget)
		go func() {
			if err := triggerRestateWorkflow(context.Background(), restateURL, job.ID, workflowReq, logger); err != nil {
				logger.Error("failed to trigger workflow",
					zap.Error(err),
					zap.String("job_id", job.ID))
			}
		}()

		response := IngestResponse{
			JobID:     job.ID,
			Status:    "accepted",
			Message:   "Scan submitted successfully, processing asynchronously",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted) // 202 Accepted

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("failed to encode response",
				zap.Error(err),
				zap.String("job_id", job.ID))
		}
	}
}

// triggerRestateWorkflow triggers the IngestWorkflow via Restate HTTP ingress
func triggerRestateWorkflow(ctx context.Context, restateURL string, jobID string, req models.IngestWorkflowRequest, logger *zap.Logger) error {
	// Restate ingress endpoint for workflows
	// POST /IngestWorkflow/{workflow-key}/run
	url := fmt.Sprintf("%s/IngestWorkflow/%s/run", restateURL, jobID)

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to trigger workflow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("workflow trigger failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	logger.Info("workflow triggered successfully",
		zap.String("job_id", jobID),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// generateJobID creates a time-ordered UUID v7 for job tracking
func generateJobID() string {
	// UUID v7 uses timestamp + random bits for time-ordered IDs
	// This ensures job IDs are sortable by creation time
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to UUID v4 if v7 fails
		return uuid.New().String()
	}
	return id.String()
}

// ingestErrorResponse writes a consistent error response for ingest endpoint
func ingestErrorResponse(w http.ResponseWriter, errorCode, message string, statusCode int) {
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

// maskPublicKey masks a public key for logging (shows first 8 chars only)
func maskPublicKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:8] + "..."
}
