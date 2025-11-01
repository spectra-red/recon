package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

var (
	// ErrNoResults indicates no results were found
	ErrNoResults = errors.New("no results found")

	// ErrDatabaseUnavailable indicates the database is unavailable
	ErrDatabaseUnavailable = errors.New("database is unavailable")

	// ErrInvalidEmbedding indicates the embedding vector is invalid
	ErrInvalidEmbedding = errors.New("invalid embedding vector")
)

// VectorSearchClient handles vector similarity searches in SurrealDB
type VectorSearchClient struct {
	db     *surrealdb.DB
	logger *zap.Logger
}

// NewVectorSearchClient creates a new vector search client
func NewVectorSearchClient(db *surrealdb.DB, logger *zap.Logger) *VectorSearchClient {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &VectorSearchClient{
		db:     db,
		logger: logger,
	}
}

// VectorSearchParams holds parameters for vector similarity search
type VectorSearchParams struct {
	// QueryEmbedding is the embedding vector to search for
	QueryEmbedding []float64

	// K is the number of results to return
	K int

	// MinScore is the minimum similarity score (optional, 0.0 to 1.0)
	MinScore float64
}

// VulnDocResult represents a vulnerability document from the database
type VulnDocResult struct {
	ID            string    `json:"id"`
	CVEID         string    `json:"cve_id"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary"`
	CVSS          float64   `json:"cvss"`
	CPE           []string  `json:"cpe"`
	PublishedDate time.Time `json:"published_date"`
	Score         float64   `json:"score"` // Similarity score
}

// VectorSearch performs a cosine similarity search on vulnerability documents
func (c *VectorSearchClient) VectorSearch(ctx context.Context, params VectorSearchParams) ([]models.VulnResult, error) {
	// Validate embedding
	if len(params.QueryEmbedding) == 0 {
		return nil, ErrInvalidEmbedding
	}

	// Validate K
	if params.K < 1 {
		params.K = models.DefaultK
	}
	if params.K > models.MaxK {
		params.K = models.MaxK
	}

	startTime := time.Now()

	c.logger.Debug("executing vector search",
		zap.Int("embedding_dim", len(params.QueryEmbedding)),
		zap.Int("k", params.K),
		zap.Float64("min_score", params.MinScore))

	// Construct SurrealDB query
	// Uses vector::similarity::cosine for cosine similarity
	// The <|> operator performs vector similarity search using the index
	query := `
		SELECT
			meta::id(id) AS id,
			cve_id,
			title,
			summary,
			cvss,
			cpe,
			published_date,
			vector::similarity::cosine(embedding, $query_embedding) AS score
		FROM vuln_doc
		WHERE embedding <|> $query_embedding
		ORDER BY score DESC
		LIMIT $k
	`

	// Execute query using the new v1.0.0 API
	// Query returns []VulnDocResult wrapped in QueryResult
	result, err := surrealdb.Query[[]VulnDocResult](ctx, c.db, query, map[string]interface{}{
		"query_embedding": params.QueryEmbedding,
		"k":               params.K,
	})
	if err != nil {
		c.logger.Error("vector search query failed",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)))
		return nil, fmt.Errorf("%w: %v", ErrDatabaseUnavailable, err)
	}

	// Extract results from query response
	if result == nil || len(*result) == 0 {
		c.logger.Debug("empty query result")
		return nil, ErrNoResults
	}

	// Get the first (and only) query result
	queryResult := (*result)[0]

	// Check for query errors
	if queryResult.Error != nil {
		c.logger.Error("query returned error",
			zap.String("error", queryResult.Error.Error()))
		return nil, fmt.Errorf("%w: %v", ErrDatabaseUnavailable, queryResult.Error)
	}

	dbResults := queryResult.Result

	// Filter by minimum score if specified
	var filtered []VulnDocResult
	for _, r := range dbResults {
		if params.MinScore > 0 && r.Score < params.MinScore {
			continue
		}
		filtered = append(filtered, r)
	}

	// Convert to model results
	results := make([]models.VulnResult, 0, len(filtered))
	for _, r := range filtered {
		results = append(results, models.VulnResult{
			CVEID:         r.CVEID,
			Title:         r.Title,
			Summary:       r.Summary,
			CVSS:          r.CVSS,
			CPE:           r.CPE,
			PublishedDate: r.PublishedDate.Format(time.RFC3339),
			Score:         r.Score,
		})
	}

	c.logger.Info("vector search completed",
		zap.Int("results", len(results)),
		zap.Int("filtered", len(dbResults)-len(filtered)),
		zap.Duration("elapsed", time.Since(startTime)))

	if len(results) == 0 {
		return nil, ErrNoResults
	}

	return results, nil
}

// VectorSearchWithMinScore is a convenience method that searches with a minimum score
func (c *VectorSearchClient) VectorSearchWithMinScore(ctx context.Context, embedding []float64, k int, minScore float64) ([]models.VulnResult, error) {
	return c.VectorSearch(ctx, VectorSearchParams{
		QueryEmbedding: embedding,
		K:              k,
		MinScore:       minScore,
	})
}

// CreateVectorSearchClient creates and initializes a vector search client with database connection
func CreateVectorSearchClient(ctx context.Context, logger *zap.Logger) (*VectorSearchClient, error) {
	// Create database connection
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseUnavailable, err)
	}

	// Sign in
	if _, err := db.SignIn(ctx, map[string]interface{}{
		"user": "root",
		"pass": "root",
	}); err != nil {
		db.Close(ctx)
		return nil, fmt.Errorf("%w: authentication failed", ErrDatabaseUnavailable)
	}

	// Use namespace and database
	if err := db.Use(ctx, "spectra", "intel"); err != nil {
		db.Close(ctx)
		return nil, fmt.Errorf("%w: failed to use database", ErrDatabaseUnavailable)
	}

	return NewVectorSearchClient(db, logger), nil
}

// Close closes the database connection
func (c *VectorSearchClient) Close(ctx context.Context) error {
	if c.db != nil {
		return c.db.Close(ctx)
	}
	return nil
}
