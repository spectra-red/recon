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

// QueryHandler creates a handler for querying host information by IP
func QueryHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Extract IP from URL parameter
		ip := chi.URLParam(r, "ip")
		if ip == "" {
			logger.Warn("missing IP parameter in request")
			writeErrorResponse(w, "missing IP parameter", http.StatusBadRequest)
			return
		}

		// Parse optional depth parameter (default: 2)
		depth := int(models.DefaultDepth())
		if depthParam := r.URL.Query().Get("depth"); depthParam != "" {
			parsedDepth, err := strconv.Atoi(depthParam)
			if err != nil {
				logger.Warn("invalid depth parameter",
					zap.String("depth", depthParam),
					zap.Error(err))
				writeErrorResponse(w, "invalid depth parameter: must be an integer", http.StatusBadRequest)
				return
			}

			if !models.ValidateDepth(parsedDepth) {
				logger.Warn("depth out of range",
					zap.Int("depth", parsedDepth))
				writeErrorResponse(w, "depth must be between 0 and 5", http.StatusBadRequest)
				return
			}

			depth = parsedDepth
		}

		logger.Info("querying host",
			zap.String("ip", ip),
			zap.Int("depth", depth))

		// Create database connection
		dbConn, err := createDBConnection(ctx, logger)
		if err != nil {
			logger.Error("database connection failed",
				zap.Error(err),
				zap.String("ip", ip))
			writeErrorResponse(w, "database connection error", http.StatusInternalServerError)
			return
		}
		defer dbConn.Close(ctx)

		// Query the host
		result, err := db.QueryHost(ctx, dbConn, logger, ip, depth)
		if err != nil {
			logger.Error("host query failed",
				zap.Error(err),
				zap.String("ip", ip),
				zap.Int("depth", depth))
			writeErrorResponse(w, "failed to query host", http.StatusInternalServerError)
			return
		}

		// Check if host was found
		if result == nil {
			logger.Info("host not found",
				zap.String("ip", ip))
			writeErrorResponse(w, "host not found", http.StatusNotFound)
			return
		}

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.Error("failed to encode response",
				zap.Error(err),
				zap.String("ip", ip))
			// Response already started, can't change status code
			return
		}

		logger.Info("host query successful",
			zap.String("ip", ip),
			zap.Int("depth", depth),
			zap.Int("port_count", len(result.Ports)),
			zap.Int("service_count", len(result.Services)),
			zap.Int("vuln_count", len(result.Vulns)))
	}
}

// createDBConnection establishes a connection to SurrealDB
func createDBConnection(ctx context.Context, logger *zap.Logger) (*surrealdb.DB, error) {
	// Create database connection
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		return nil, err
	}

	// Sign in with credentials
	if _, err := db.SignIn(ctx, map[string]interface{}{
		"user": "root",
		"pass": "root",
	}); err != nil {
		db.Close(ctx)
		return nil, err
	}

	// Use namespace and database
	if err := db.Use(ctx, "spectra", "intel"); err != nil {
		db.Close(ctx)
		return nil, err
	}

	return db, nil
}

// writeErrorResponse writes a standard JSON error response
func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
		"code":    statusCode,
	}

	json.NewEncoder(w).Encode(errorResp)
}
