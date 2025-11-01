package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// HealthHandler creates a health check handler with database connectivity check
func HealthHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		services := make(map[string]string)
		services["api"] = "ok"

		// Check SurrealDB connectivity
		dbStatus := checkDatabaseConnection(ctx, logger)
		services["database"] = dbStatus

		// Determine overall health status
		overallStatus := "healthy"
		if dbStatus != "ok" {
			overallStatus = "degraded"
			logger.Warn("database connectivity issue",
				zap.String("db_status", dbStatus))
		}

		response := HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Services:  services,
		}

		w.Header().Set("Content-Type", "application/json")

		// Return 200 even if degraded (API is still functional)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("failed to encode health response",
				zap.Error(err))
		}
	}
}

// checkDatabaseConnection attempts to connect to SurrealDB and returns status
func checkDatabaseConnection(ctx context.Context, logger *zap.Logger) string {
	// Create database connection
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		logger.Debug("database connection failed",
			zap.Error(err),
			zap.String("reason", "connection_error"))
		return "unavailable"
	}
	defer db.Close(ctx)

	// Sign in with credentials
	if _, err := db.SignIn(ctx, map[string]interface{}{
		"user": "root",
		"pass": "root",
	}); err != nil {
		logger.Debug("database authentication failed",
			zap.Error(err),
			zap.String("reason", "auth_error"))
		return "unavailable"
	}

	// Use namespace and database
	if err := db.Use(ctx, "spectra", "intel"); err != nil {
		logger.Debug("database use failed",
			zap.Error(err),
			zap.String("reason", "use_error"))
		return "unavailable"
	}

	// Verify connection with version check
	_, err = db.Version(ctx)
	if err != nil {
		logger.Debug("database version check failed",
			zap.Error(err),
			zap.String("reason", "version_error"))
		return "unavailable"
	}

	return "ok"
}
