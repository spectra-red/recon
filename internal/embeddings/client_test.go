package embeddings

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewClient(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "valid config with API key",
			cfg: Config{
				APIKey: "test-api-key",
				Logger: logger,
			},
			wantErr: nil,
		},
		{
			name: "missing API key and no env var",
			cfg: Config{
				Logger: logger,
			},
			wantErr: ErrInvalidAPIKey,
		},
		{
			name: "uses default model when not specified",
			cfg: Config{
				APIKey: "test-api-key",
				Logger: logger,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env var for consistent testing
			oldKey := os.Getenv("OPENAI_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")
			defer func() {
				if oldKey != "" {
					os.Setenv("OPENAI_API_KEY", oldKey)
				}
			}()

			client, err := NewClient(tt.cfg)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, DefaultModel, client.model)
				assert.Equal(t, DefaultTimeout, client.timeout)
			}
		})
	}
}

func TestNewClientFromEnv(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name    string
		envKey  string
		wantErr error
	}{
		{
			name:    "valid API key from environment",
			envKey:  "test-env-api-key",
			wantErr: nil,
		},
		{
			name:    "missing environment variable",
			envKey:  "",
			wantErr: ErrInvalidAPIKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			oldKey := os.Getenv("OPENAI_API_KEY")
			if tt.envKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.envKey)
			} else {
				os.Unsetenv("OPENAI_API_KEY")
			}
			defer func() {
				if oldKey != "" {
					os.Setenv("OPENAI_API_KEY", oldKey)
				} else {
					os.Unsetenv("OPENAI_API_KEY")
				}
			}()

			client, err := NewClientFromEnv(logger)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestGenerateEmbedding_Validation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client, err := NewClient(Config{
		APIKey: "test-api-key",
		Logger: logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr error
	}{
		{
			name:    "empty query",
			query:   "",
			wantErr: ErrEmptyQuery,
		},
		{
			name:    "query too long",
			query:   strings.Repeat("a", MaxQueryLength+1),
			wantErr: ErrQueryTooLong,
		},
		{
			name:    "valid query length (at max)",
			query:   strings.Repeat("a", MaxQueryLength),
			wantErr: ErrServiceUnavailable, // Will fail at API call, but validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GenerateEmbedding(ctx, tt.query)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestGenerateEmbedding_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	logger := zaptest.NewLogger(t)
	client, err := NewClient(Config{
		APIKey:  "test-api-key",
		Logger:  logger,
		Timeout: 1 * time.Nanosecond, // Very short timeout to trigger timeout
	})
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.GenerateEmbedding(ctx, "test query")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrServiceUnavailable)
}

func TestGenerateEmbeddingBatch_Validation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client, err := NewClient(Config{
		APIKey: "test-api-key",
		Logger: logger,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		queries []string
		wantErr error
	}{
		{
			name:    "empty queries",
			queries: []string{},
			wantErr: ErrEmptyQuery,
		},
		{
			name:    "contains empty query",
			queries: []string{"valid", "", "also valid"},
			wantErr: nil, // Will show specific error about index
		},
		{
			name:    "contains query too long",
			queries: []string{"valid", strings.Repeat("a", MaxQueryLength+1)},
			wantErr: ErrQueryTooLong,
		},
		{
			name:    "all valid queries",
			queries: []string{"query 1", "query 2", "query 3"},
			wantErr: ErrServiceUnavailable, // Will fail at API call, but validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GenerateEmbeddingBatch(ctx, tt.queries)

			assert.Error(t, err)
			if tt.wantErr != nil {
				// Check if error contains the expected error
				if tt.wantErr == ErrEmptyQuery || tt.wantErr == ErrQueryTooLong {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			}
		})
	}
}

// MockEmbeddingGenerator creates a mock embedding generator for testing
// Returns a static vector for predictable test results
func MockEmbeddingGenerator(dimension int) func(ctx context.Context, query string) ([]float64, error) {
	return func(ctx context.Context, query string) ([]float64, error) {
		// Return a deterministic embedding based on query
		embedding := make([]float64, dimension)
		for i := 0; i < dimension; i++ {
			// Use query length to generate different but consistent embeddings
			embedding[i] = float64(len(query)+i) / float64(dimension)
		}
		return embedding, nil
	}
}

// MockEmbeddingBatchGenerator creates a mock batch embedding generator
func MockEmbeddingBatchGenerator(dimension int) func(ctx context.Context, queries []string) ([][]float64, error) {
	return func(ctx context.Context, queries []string) ([][]float64, error) {
		gen := MockEmbeddingGenerator(dimension)
		embeddings := make([][]float64, len(queries))
		for i, query := range queries {
			emb, err := gen(ctx, query)
			if err != nil {
				return nil, err
			}
			embeddings[i] = emb
		}
		return embeddings, nil
	}
}

func TestMockEmbeddingGenerator(t *testing.T) {
	mockGen := MockEmbeddingGenerator(ExpectedDimension)
	ctx := context.Background()

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "short query",
			query: "test",
		},
		{
			name:  "longer query",
			query: "this is a longer test query",
		},
		{
			name:  "vulnerability query",
			query: "nginx remote code execution CVE-2021-23017",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedding, err := mockGen(ctx, tt.query)

			assert.NoError(t, err)
			assert.NotNil(t, embedding)
			assert.Equal(t, ExpectedDimension, len(embedding))

			// Verify embeddings are deterministic
			embedding2, err := mockGen(ctx, tt.query)
			assert.NoError(t, err)
			assert.Equal(t, embedding, embedding2)

			// Verify different queries produce different embeddings
			if tt.query != "test" {
				testEmb, _ := mockGen(ctx, "test")
				assert.NotEqual(t, testEmb, embedding)
			}
		})
	}
}

func TestMockEmbeddingBatchGenerator(t *testing.T) {
	mockBatchGen := MockEmbeddingBatchGenerator(ExpectedDimension)
	ctx := context.Background()

	queries := []string{
		"short",
		"medium length query",
		"this is a much longer query string",
	}

	embeddings, err := mockBatchGen(ctx, queries)

	assert.NoError(t, err)
	assert.NotNil(t, embeddings)
	assert.Equal(t, len(queries), len(embeddings))

	// Verify each embedding has correct dimension
	for i, emb := range embeddings {
		assert.Equal(t, ExpectedDimension, len(emb), "embedding %d has wrong dimension", i)
	}

	// Verify embeddings are deterministic for the same query
	embeddings2, _ := mockBatchGen(ctx, queries)
	assert.Equal(t, embeddings[0], embeddings2[0])

	// Verify different query lengths produce different embeddings
	// Check that the first few elements differ (based on query length)
	assert.NotEqual(t, embeddings[0][0], embeddings[1][0], "different length queries should produce different embeddings")
	assert.NotEqual(t, embeddings[1][0], embeddings[2][0], "different length queries should produce different embeddings")
}
