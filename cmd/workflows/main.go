package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	restate "github.com/restatedev/sdk-go"
	"github.com/restatedev/sdk-go/server"
	"github.com/spectra-red/recon/internal/workflows"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Get configuration from environment
	surrealURL := getEnv("SURREALDB_URL", "ws://localhost:8000/rpc")
	surrealUser := getEnv("SURREALDB_USER", "root")
	surrealPass := getEnv("SURREALDB_PASS", "root")
	surrealNS := getEnv("SURREALDB_NAMESPACE", "spectra")
	surrealDB := getEnv("SURREALDB_DATABASE", "intel_mesh")
	port := getEnv("PORT", "9080")

	logger.Info("initializing Spectra-Red workflow service",
		zap.String("port", port),
		zap.String("surrealdb_url", surrealURL))

	// Connect to SurrealDB
	db, err := surrealdb.New(surrealURL)
	if err != nil {
		logger.Fatal("failed to connect to SurrealDB",
			zap.Error(err),
			zap.String("url", surrealURL))
	}
	defer db.Close(context.Background())

	// Authenticate and use namespace/database
	if _, err := db.SignIn(context.Background(), surrealdb.Auth{
		Username: surrealUser,
		Password: surrealPass,
	}); err != nil {
		logger.Fatal("failed to authenticate with SurrealDB",
			zap.Error(err))
	}

	if err := db.Use(context.Background(), surrealNS, surrealDB); err != nil {
		logger.Fatal("failed to use namespace/database",
			zap.Error(err),
			zap.String("namespace", surrealNS),
			zap.String("database", surrealDB))
	}

	logger.Info("connected to SurrealDB successfully",
		zap.String("namespace", surrealNS),
		zap.String("database", surrealDB))

	// Initialize workflows
	ingestWorkflow := workflows.NewIngestWorkflow(db)

	// Create Restate server and register workflows
	restateServer := server.NewRestate().Bind(
		restate.Reflect(ingestWorkflow),
	)

	// Get HTTP handler
	handler, err := restateServer.Handler()
	if err != nil {
		logger.Fatal("failed to create Restate handler",
			zap.Error(err))
	}

	// Setup HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("workflow service starting",
			zap.String("address", httpServer.Addr))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start",
				zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down workflow service...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown",
			zap.Error(err))
	}

	logger.Info("workflow service stopped")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
