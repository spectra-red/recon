package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spectra-red/recon/internal/api"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

const (
	// ServerPort is the port the API server listens on
	ServerPort = "3000"
	// ServerVersion is the current API version
	ServerVersion = "0.1.0"
	// ShutdownTimeout is the maximum time to wait for graceful shutdown
	ShutdownTimeout = 10 * time.Second
)

func main() {
	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("initializing Spectra-Red API server",
		zap.String("version", ServerVersion),
		zap.String("port", ServerPort))

	// Get database configuration from environment
	surrealURL := getEnv("SURREALDB_URL", "ws://localhost:8000/rpc")
	surrealUser := getEnv("SURREALDB_USER", "root")
	surrealPass := getEnv("SURREALDB_PASS", "root")
	surrealNS := getEnv("SURREALDB_NAMESPACE", "spectra")
	surrealDB := getEnv("SURREALDB_DATABASE", "intel_mesh")

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

	// Setup routes with middleware
	router := api.SetupRoutes(logger, db)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         ":" + ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		logger.Info("server starting",
			zap.String("addr", srv.Addr),
			zap.String("version", ServerVersion))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Block until we receive a signal or error
	select {
	case err := <-serverErrors:
		logger.Fatal("server failed to start",
			zap.Error(err))

	case sig := <-stop:
		logger.Info("shutdown signal received",
			zap.String("signal", sig.String()))

		// Create a deadline for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		logger.Info("shutting down server gracefully",
			zap.Duration("timeout", ShutdownTimeout))

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failed",
				zap.Error(err))
			// Force close
			srv.Close()
		}

		logger.Info("server stopped")
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
