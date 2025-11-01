package api

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/spectra-red/recon/internal/api/handlers"
	"github.com/spectra-red/recon/internal/api/middleware"
	"github.com/spectra-red/recon/internal/db"
	"github.com/spectra-red/recon/internal/embeddings"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// SetupRoutes configures all routes and middleware for the API server
func SetupRoutes(logger *zap.Logger, dbClient *surrealdb.DB) *chi.Mux {
	r := chi.NewRouter()

	// Middleware chain - order matters!
	// 1. Request ID - must be first to ensure all logs have request IDs
	r.Use(middleware.RequestID())

	// 2. Logger - logs all requests with request IDs
	r.Use(middleware.Logger(logger))

	// 3. Recoverer - recovers from panics
	r.Use(chimiddleware.Recoverer)

	// Health check endpoint (no authentication required)
	r.Get("/health", handlers.HealthHandler(logger))

	// Initialize rate limiter for ingest endpoint (60 requests per minute per scanner)
	ingestRateLimiter := middleware.NewRateLimiter(60, logger)
	// Start background cleanup of stale rate limit buckets (every 10 minutes, remove buckets older than 1 hour)
	ingestRateLimiter.StartCleanupRoutine(10*time.Minute, 1*time.Hour)

	// Initialize rate limiter for query endpoints (30 requests per minute per user)
	queryRateLimiter := middleware.NewRateLimiter(30, logger)
	queryRateLimiter.StartCleanupRoutine(10*time.Minute, 1*time.Hour)

	// Get Restate URL from environment (for workflow triggering)
	restateURL := getEnv("RESTATE_URL", "http://localhost:8080")

	// API routes under /v1 prefix
	r.Route("/v1", func(r chi.Router) {
		// Mesh ingest endpoint with rate limiting
		r.Route("/mesh", func(r chi.Router) {
			r.With(middleware.RateLimitMiddleware(ingestRateLimiter)).
				Post("/ingest", handlers.IngestHandler(logger, dbClient, restateURL))
		})

		// Job tracking endpoints
		r.Route("/jobs", func(r chi.Router) {
			// Apply rate limiting to job endpoints
			r.Use(middleware.RateLimitMiddleware(queryRateLimiter))

			// GET /v1/jobs - List jobs with optional filters
			// Query params: ?limit=50&offset=0&state=pending&scanner_key=xyz&order_by=created_at&order_desc=true
			r.Get("/", handlers.ListJobsHandler(dbClient, logger))

			// GET /v1/jobs/{job_id} - Get job status by ID
			r.Get("/{job_id}", handlers.GetJobHandler(dbClient, logger))
		})

		// Query endpoints
		r.Route("/query", func(r chi.Router) {
			// Apply rate limiting to all query endpoints
			r.Use(middleware.RateLimitMiddleware(queryRateLimiter))

			// GET /v1/query/host/{ip} - Query host by IP with optional depth parameter
			// Query params: ?depth=0-5 (default: 2)
			r.Get("/host/{ip}", handlers.QueryHandler(logger))

			// POST /v1/query/graph - Advanced graph traversal queries
			// Supports: by_asn, by_location, by_vuln, by_service
			r.Post("/graph", handlers.GraphQueryHandlerFunc(logger))

			// POST /v1/query/similar - Vector similarity search for vulnerabilities
			// Accepts natural language query, returns top K similar vulnerability documents
			r.Post("/similar", setupSimilarityHandler(logger))
		})
	})

	// API routes under /v0 prefix (legacy, for future use)
	r.Route("/v0", func(r chi.Router) {
		// Future API endpoints will be defined here
		// Example structure:
		// r.Route("/targets", func(r chi.Router) {
		//     r.Get("/", listTargetsHandler)
		//     r.Post("/", createTargetHandler)
		// })
	})

	return r
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupSimilarityHandler initializes and returns the similarity search handler
// This function handles the initialization of dependencies (embedding client, vector search client)
// and returns a configured handler function with graceful degradation if services are unavailable
func setupSimilarityHandler(logger *zap.Logger) http.HandlerFunc {
	// Initialize embedding client from environment
	embeddingClient, err := embeddings.NewClientFromEnv(logger)
	if err != nil {
		logger.Warn("failed to initialize embedding client",
			zap.Error(err),
			zap.String("hint", "similarity search will return errors until OPENAI_API_KEY is configured"))

		// Return a handler that always returns an error about missing configuration
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"embedding service not configured","code":"SERVICE_UNAVAILABLE","details":"The OpenAI API key is not configured. Please set the OPENAI_API_KEY environment variable.","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
		}
	}

	// Initialize vector search client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	vectorClient, err := db.CreateVectorSearchClient(ctx, logger)
	if err != nil {
		logger.Warn("failed to initialize vector search client",
			zap.Error(err),
			zap.String("hint", "similarity search will return errors until database is available"))

		// Return a handler that always returns an error about database unavailability
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"database service not available","code":"SERVICE_UNAVAILABLE","details":"The vector search database is not available. Please ensure SurrealDB is running and accessible.","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
		}
	}

	logger.Info("similarity search endpoint initialized successfully")

	// Return the configured handler
	return handlers.SimilarHandlerFunc(embeddingClient, vectorClient, logger)
}
