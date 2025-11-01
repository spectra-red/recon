package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/spectra-red/recon/internal/db"
	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// GraphQueryHandler handles graph traversal queries
type GraphQueryHandler struct {
	executor *db.GraphQueryExecutor
	logger   *zap.Logger
}

// NewGraphQueryHandler creates a new graph query handler
func NewGraphQueryHandler(logger *zap.Logger) (*GraphQueryHandler, error) {
	// Create database connection
	dbConn, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Sign in with credentials
	if _, err := dbConn.SignIn(ctx, map[string]interface{}{
		"user": "root",
		"pass": "root",
	}); err != nil {
		dbConn.Close(ctx)
		return nil, err
	}

	// Use namespace and database
	if err := dbConn.Use(ctx, "spectra", "intel"); err != nil {
		dbConn.Close(ctx)
		return nil, err
	}

	executor := db.NewGraphQueryExecutor(dbConn, logger)

	return &GraphQueryHandler{
		executor: executor,
		logger:   logger,
	}, nil
}

// HandleGraphQuery handles POST /v1/query/graph requests
func (h *GraphQueryHandler) HandleGraphQuery(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout protection (5 seconds max)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parse request body
	var req models.GraphQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode graph query request",
			zap.Error(err),
			zap.String("remote_addr", r.RemoteAddr))
		h.respondWithError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Log query request
	h.logger.Info("executing graph query",
		zap.String("query_type", string(req.QueryType)),
		zap.Any("asn", req.ASN),
		zap.String("city", req.City),
		zap.String("cve", req.CVE),
		zap.String("product", req.Product),
		zap.Int("limit", req.Limit),
		zap.Int("offset", req.Offset))

	// Execute query with timeout protection
	resp, err := h.executor.ExecuteGraphQuery(ctx, req)
	if err != nil {
		// Check if error was due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			h.logger.Warn("graph query timeout",
				zap.String("query_type", string(req.QueryType)),
				zap.Duration("timeout", 5*time.Second))
			h.respondWithError(w, http.StatusRequestTimeout, "query timeout exceeded", err)
			return
		}

		// Check for validation errors
		if validationErr, ok := err.(*models.ValidationError); ok {
			h.logger.Warn("graph query validation error",
				zap.String("field", validationErr.Field),
				zap.String("message", validationErr.Message))
			h.respondWithError(w, http.StatusBadRequest, validationErr.Message, err)
			return
		}

		// Other errors
		h.logger.Error("graph query execution failed",
			zap.Error(err),
			zap.String("query_type", string(req.QueryType)))
		h.respondWithError(w, http.StatusInternalServerError, "query execution failed", err)
		return
	}

	// Log query success
	h.logger.Info("graph query completed",
		zap.String("query_type", string(req.QueryType)),
		zap.Int("result_count", len(resp.Results)),
		zap.Float64("query_time_ms", resp.QueryTime))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode graph query response",
			zap.Error(err))
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// respondWithError sends an error response
func (h *GraphQueryHandler) respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	if err != nil {
		errResp.Details = err.Error()
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		h.logger.Error("failed to encode error response",
			zap.Error(err))
	}
}

// GraphQueryHandlerFunc returns a handler function that can be used with chi router
func GraphQueryHandlerFunc(logger *zap.Logger) http.HandlerFunc {
	handler, err := NewGraphQueryHandler(logger)
	if err != nil {
		logger.Error("failed to create graph query handler",
			zap.Error(err))
		// Return a handler that always returns 503
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "Service Unavailable",
				Message: "database connection unavailable",
			})
		}
	}

	return handler.HandleGraphQuery
}
