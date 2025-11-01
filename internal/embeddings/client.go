package embeddings

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	// MaxQueryLength is the maximum allowed query string length
	MaxQueryLength = 500

	// DefaultModel is the default OpenAI embedding model
	DefaultModel = openai.SmallEmbedding3

	// ExpectedDimension is the expected embedding dimension for text-embedding-3-small
	ExpectedDimension = 1536

	// DefaultTimeout for embedding generation
	DefaultTimeout = 10 * time.Second
)

var (
	// ErrQueryTooLong indicates the query string exceeds the maximum allowed length
	ErrQueryTooLong = errors.New("query string exceeds maximum length")

	// ErrServiceUnavailable indicates the embedding service is unavailable
	ErrServiceUnavailable = errors.New("embedding service is unavailable")

	// ErrInvalidAPIKey indicates the OpenAI API key is missing or invalid
	ErrInvalidAPIKey = errors.New("OpenAI API key is missing or invalid")

	// ErrEmptyQuery indicates the query string is empty
	ErrEmptyQuery = errors.New("query string cannot be empty")
)

// Client handles embedding generation via OpenAI API
type Client struct {
	openaiClient *openai.Client
	logger       *zap.Logger
	model        openai.EmbeddingModel
	timeout      time.Duration
}

// Config holds configuration for the embedding client
type Config struct {
	APIKey  string
	Model   openai.EmbeddingModel
	Timeout time.Duration
	Logger  *zap.Logger
}

// NewClient creates a new embedding client
func NewClient(cfg Config) (*Client, error) {
	// Validate API key
	if cfg.APIKey == "" {
		// Try to get from environment
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
		if cfg.APIKey == "" {
			return nil, ErrInvalidAPIKey
		}
	}

	// Set defaults
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	client := openai.NewClient(cfg.APIKey)

	return &Client{
		openaiClient: client,
		logger:       cfg.Logger,
		model:        cfg.Model,
		timeout:      cfg.Timeout,
	}, nil
}

// NewClientFromEnv creates a client using environment variables
func NewClientFromEnv(logger *zap.Logger) (*Client, error) {
	return NewClient(Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Logger: logger,
	})
}

// GenerateEmbedding generates an embedding vector for the given query text
func (c *Client) GenerateEmbedding(ctx context.Context, query string) ([]float64, error) {
	// Validate query
	if query == "" {
		return nil, ErrEmptyQuery
	}
	if len(query) > MaxQueryLength {
		return nil, fmt.Errorf("%w: %d characters (max %d)", ErrQueryTooLong, len(query), MaxQueryLength)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Record start time for logging
	startTime := time.Now()

	// Call OpenAI API
	c.logger.Debug("generating embedding",
		zap.String("model", string(c.model)),
		zap.Int("query_length", len(query)))

	req := openai.EmbeddingRequest{
		Input: []string{query},
		Model: c.model,
	}

	resp, err := c.openaiClient.CreateEmbeddings(ctx, req)
	if err != nil {
		c.logger.Error("failed to generate embedding",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)))

		// Check if it's a context error
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("%w: request timeout", ErrServiceUnavailable)
		}

		return nil, fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}

	// Validate response
	if len(resp.Data) == 0 {
		c.logger.Error("empty embedding response")
		return nil, fmt.Errorf("%w: empty response", ErrServiceUnavailable)
	}

	embeddingFloat32 := resp.Data[0].Embedding

	// Validate embedding dimension
	if len(embeddingFloat32) != ExpectedDimension {
		c.logger.Warn("unexpected embedding dimension",
			zap.Int("expected", ExpectedDimension),
			zap.Int("actual", len(embeddingFloat32)))
	}

	// Convert from []float32 to []float64
	embedding := make([]float64, len(embeddingFloat32))
	for i, v := range embeddingFloat32 {
		embedding[i] = float64(v)
	}

	// Log successful generation with timing
	c.logger.Info("embedding generated successfully",
		zap.Duration("elapsed", time.Since(startTime)),
		zap.Int("dimension", len(embedding)),
		zap.Int("total_tokens", resp.Usage.TotalTokens))

	return embedding, nil
}

// GenerateEmbeddingBatch generates embeddings for multiple queries in a single API call
// This is more efficient for batch processing
func (c *Client) GenerateEmbeddingBatch(ctx context.Context, queries []string) ([][]float64, error) {
	if len(queries) == 0 {
		return nil, ErrEmptyQuery
	}

	// Validate all queries
	for i, query := range queries {
		if query == "" {
			return nil, fmt.Errorf("query at index %d is empty", i)
		}
		if len(query) > MaxQueryLength {
			return nil, fmt.Errorf("query at index %d: %w: %d characters (max %d)",
				i, ErrQueryTooLong, len(query), MaxQueryLength)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Record start time
	startTime := time.Now()

	c.logger.Debug("generating batch embeddings",
		zap.String("model", string(c.model)),
		zap.Int("query_count", len(queries)))

	req := openai.EmbeddingRequest{
		Input: queries,
		Model: c.model,
	}

	resp, err := c.openaiClient.CreateEmbeddings(ctx, req)
	if err != nil {
		c.logger.Error("failed to generate batch embeddings",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)))

		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("%w: request timeout", ErrServiceUnavailable)
		}

		return nil, fmt.Errorf("%w: %v", ErrServiceUnavailable, err)
	}

	// Validate response
	if len(resp.Data) != len(queries) {
		c.logger.Error("embedding count mismatch",
			zap.Int("expected", len(queries)),
			zap.Int("actual", len(resp.Data)))
		return nil, fmt.Errorf("%w: embedding count mismatch", ErrServiceUnavailable)
	}

	// Extract embeddings and convert from []float32 to []float64
	embeddings := make([][]float64, len(resp.Data))
	for i, data := range resp.Data {
		embeddingFloat32 := data.Embedding
		embedding := make([]float64, len(embeddingFloat32))
		for j, v := range embeddingFloat32 {
			embedding[j] = float64(v)
		}
		embeddings[i] = embedding
	}

	c.logger.Info("batch embeddings generated successfully",
		zap.Duration("elapsed", time.Since(startTime)),
		zap.Int("count", len(embeddings)),
		zap.Int("total_tokens", resp.Usage.TotalTokens))

	return embeddings, nil
}

// HealthCheck verifies that the embedding service is accessible
func (c *Client) HealthCheck(ctx context.Context) error {
	// Try to generate a simple embedding
	_, err := c.GenerateEmbedding(ctx, "test")
	return err
}
