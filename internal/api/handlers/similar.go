package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/spectra-red/recon/internal/db"
	"github.com/spectra-red/recon/internal/embeddings"
	"github.com/spectra-red/recon/internal/models"
	"go.uber.org/zap"
)

// SimilarHandler handles similarity search requests for vulnerability documents
type SimilarHandler struct {
	embeddingClient *embeddings.Client
	vectorClient    *db.VectorSearchClient
	logger          *zap.Logger
}

// NewSimilarHandler creates a new similarity search handler
func NewSimilarHandler(embeddingClient *embeddings.Client, vectorClient *db.VectorSearchClient, logger *zap.Logger) *SimilarHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &SimilarHandler{
		embeddingClient: embeddingClient,
		vectorClient:    vectorClient,
		logger:          logger,
	}
}

// ServeHTTP handles POST /v1/query/similar requests
func (h *SimilarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.writeError(w, "method not allowed", http.StatusMethodNotAllowed, "")
		return
	}

	ctx := r.Context()

	// Parse request body
	var req models.SimilarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request",
			zap.Error(err))
		h.writeError(w, "invalid request body", http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.logger.Warn("request validation failed",
			zap.Error(err),
			zap.String("query", req.Query))
		h.writeError(w, "validation error", http.StatusBadRequest, err.Error())
		return
	}

	// Log the request
	h.logger.Info("processing similarity search",
		zap.String("query", req.Query),
		zap.Int("k", req.GetK()))

	// Execute similarity search
	results, err := h.executeSimilaritySearch(ctx, req)
	if err != nil {
		h.handleSearchError(w, err, req.Query)
		return
	}

	// Build response
	response := models.SimilarResponse{
		Query:     req.Query,
		Results:   results,
		Count:     len(results),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response",
			zap.Error(err))
	}

	h.logger.Info("similarity search completed",
		zap.String("query", req.Query),
		zap.Int("results", len(results)))
}

// executeSimilaritySearch performs the complete similarity search workflow
func (h *SimilarHandler) executeSimilaritySearch(ctx context.Context, req models.SimilarRequest) ([]models.VulnResult, error) {
	// Step 1: Generate embedding from query text
	embedding, err := h.embeddingClient.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		h.logger.Error("failed to generate embedding",
			zap.Error(err),
			zap.String("query", req.Query))
		return nil, err
	}

	h.logger.Debug("embedding generated",
		zap.Int("dimension", len(embedding)),
		zap.String("query", req.Query))

	// Step 2: Perform vector similarity search
	results, err := h.vectorClient.VectorSearch(ctx, db.VectorSearchParams{
		QueryEmbedding: embedding,
		K:              req.GetK(),
		MinScore:       0.0, // No minimum score filter for now
	})

	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			// No results is not an error - return empty list
			h.logger.Info("no similar vulnerabilities found",
				zap.String("query", req.Query))
			return []models.VulnResult{}, nil
		}
		return nil, err
	}

	return results, nil
}

// handleSearchError handles errors from the search operation with graceful fallback
func (h *SimilarHandler) handleSearchError(w http.ResponseWriter, err error, query string) {
	// Check error type and provide appropriate response
	switch {
	case errors.Is(err, embeddings.ErrServiceUnavailable):
		// Embedding service is unavailable - return 503 with helpful message
		h.logger.Error("embedding service unavailable",
			zap.Error(err),
			zap.String("query", query))
		h.writeError(w,
			"embedding service is temporarily unavailable",
			http.StatusServiceUnavailable,
			"Please ensure the OpenAI API key is configured and the service is accessible. Check the OPENAI_API_KEY environment variable.")

	case errors.Is(err, embeddings.ErrInvalidAPIKey):
		// API key issue - return 500 (this is a configuration error)
		h.logger.Error("embedding service configuration error",
			zap.Error(err))
		h.writeError(w,
			"embedding service configuration error",
			http.StatusInternalServerError,
			"The embedding service is not properly configured. Please contact the administrator.")

	case errors.Is(err, db.ErrDatabaseUnavailable):
		// Database unavailable - return 503
		h.logger.Error("database unavailable",
			zap.Error(err),
			zap.String("query", query))
		h.writeError(w,
			"database service is temporarily unavailable",
			http.StatusServiceUnavailable,
			"The vector search database is currently unavailable. Please try again later.")

	case errors.Is(err, db.ErrNoResults):
		// No results found - return empty result set (already handled in executeSimilaritySearch)
		// This case shouldn't happen here, but handle it anyway
		h.logger.Info("no results found",
			zap.String("query", query))
		response := models.SimilarResponse{
			Query:     query,
			Results:   []models.VulnResult{},
			Count:     0,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	default:
		// Unknown error - return 500
		h.logger.Error("similarity search failed",
			zap.Error(err),
			zap.String("query", query))
		h.writeError(w,
			"internal server error",
			http.StatusInternalServerError,
			"An unexpected error occurred during the similarity search.")
	}
}

// writeError writes an error response
func (h *SimilarHandler) writeError(w http.ResponseWriter, message string, statusCode int, details string) {
	response := models.ErrorResponse{
		Error:     message,
		Details:   details,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Add error code based on status
	switch statusCode {
	case http.StatusBadRequest:
		response.Code = "BAD_REQUEST"
	case http.StatusServiceUnavailable:
		response.Code = "SERVICE_UNAVAILABLE"
	case http.StatusInternalServerError:
		response.Code = "INTERNAL_ERROR"
	case http.StatusMethodNotAllowed:
		response.Code = "METHOD_NOT_ALLOWED"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode error response",
			zap.Error(err))
	}
}

// SimilarHandlerFunc creates a handler function for similarity search
// This is a convenience function for route registration
func SimilarHandlerFunc(embeddingClient *embeddings.Client, vectorClient *db.VectorSearchClient, logger *zap.Logger) http.HandlerFunc {
	handler := NewSimilarHandler(embeddingClient, vectorClient, logger)
	return handler.ServeHTTP
}
