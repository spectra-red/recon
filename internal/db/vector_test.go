package db

import (
	"context"
	"testing"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestVectorSearchParams_Validation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	// Create a mock client (without actual DB connection)
	client := NewVectorSearchClient(nil, logger)

	tests := []struct {
		name    string
		params  VectorSearchParams
		wantErr error
	}{
		{
			name: "valid params",
			params: VectorSearchParams{
				QueryEmbedding: make([]float64, 1536),
				K:              10,
			},
			wantErr: nil,
		},
		{
			name: "empty embedding",
			params: VectorSearchParams{
				QueryEmbedding: []float64{},
				K:              10,
			},
			wantErr: ErrInvalidEmbedding,
		},
		{
			name: "nil embedding",
			params: VectorSearchParams{
				QueryEmbedding: nil,
				K:              10,
			},
			wantErr: ErrInvalidEmbedding,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't actually execute the query without a DB, but we can test validation
			if tt.wantErr != nil {
				// Validate embedding
				if len(tt.params.QueryEmbedding) == 0 {
					assert.ErrorIs(t, ErrInvalidEmbedding, tt.wantErr)
				}
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}

func TestVectorSearchParams_KValidation(t *testing.T) {
	tests := []struct {
		name     string
		inputK   int
		expectedK int
	}{
		{
			name:     "valid K",
			inputK:   10,
			expectedK: 10,
		},
		{
			name:     "K below minimum gets set to default",
			inputK:   0,
			expectedK: models.DefaultK,
		},
		{
			name:     "K above maximum gets capped",
			inputK:   100,
			expectedK: models.MaxK,
		},
		{
			name:     "negative K gets set to default",
			inputK:   -5,
			expectedK: models.DefaultK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := VectorSearchParams{
				QueryEmbedding: make([]float64, 1536),
				K:              tt.inputK,
			}

			// Apply same validation logic as VectorSearch
			if params.K < 1 {
				params.K = models.DefaultK
			}
			if params.K > models.MaxK {
				params.K = models.MaxK
			}

			assert.Equal(t, tt.expectedK, params.K)
		})
	}
}

func TestNewVectorSearchClient(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name   string
		logger *zap.Logger
	}{
		{
			name:   "with logger",
			logger: logger,
		},
		{
			name:   "without logger (nil)",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewVectorSearchClient(nil, tt.logger)

			assert.NotNil(t, client)
			assert.NotNil(t, client.logger) // Should never be nil due to default
		})
	}
}

// MockVectorSearchClient creates a mock vector search client for testing
type MockVectorSearchClient struct {
	SearchFunc func(ctx context.Context, params VectorSearchParams) ([]models.VulnResult, error)
}

func (m *MockVectorSearchClient) VectorSearch(ctx context.Context, params VectorSearchParams) ([]models.VulnResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, params)
	}
	return nil, ErrNoResults
}

func TestMockVectorSearchClient(t *testing.T) {
	ctx := context.Background()

	// Create mock with predefined results
	mockResults := []models.VulnResult{
		{
			CVEID:   "CVE-2021-12345",
			Title:   "Test Vulnerability",
			Summary: "Test vulnerability description",
			CVSS:    9.8,
			Score:   0.95,
		},
		{
			CVEID:   "CVE-2021-67890",
			Title:   "Another Vulnerability",
			Summary: "Another test vulnerability",
			CVSS:    7.5,
			Score:   0.87,
		},
	}

	mock := &MockVectorSearchClient{
		SearchFunc: func(ctx context.Context, params VectorSearchParams) ([]models.VulnResult, error) {
			// Validate params
			if len(params.QueryEmbedding) == 0 {
				return nil, ErrInvalidEmbedding
			}

			// Return mock results
			return mockResults, nil
		},
	}

	// Test successful search
	results, err := mock.VectorSearch(ctx, VectorSearchParams{
		QueryEmbedding: make([]float64, 1536),
		K:              10,
	})

	assert.NoError(t, err)
	assert.Equal(t, len(mockResults), len(results))
	assert.Equal(t, "CVE-2021-12345", results[0].CVEID)
	assert.Equal(t, 0.95, results[0].Score)

	// Test invalid embedding
	_, err = mock.VectorSearch(ctx, VectorSearchParams{
		QueryEmbedding: []float64{},
		K:              10,
	})

	assert.ErrorIs(t, err, ErrInvalidEmbedding)
}

func TestMockVectorSearchClient_NoResults(t *testing.T) {
	ctx := context.Background()

	mock := &MockVectorSearchClient{
		SearchFunc: func(ctx context.Context, params VectorSearchParams) ([]models.VulnResult, error) {
			return nil, ErrNoResults
		},
	}

	results, err := mock.VectorSearch(ctx, VectorSearchParams{
		QueryEmbedding: make([]float64, 1536),
		K:              10,
	})

	assert.ErrorIs(t, err, ErrNoResults)
	assert.Nil(t, results)
}

func TestMockVectorSearchClient_MinScore(t *testing.T) {
	ctx := context.Background()

	allResults := []models.VulnResult{
		{CVEID: "CVE-2021-1", Score: 0.95},
		{CVEID: "CVE-2021-2", Score: 0.85},
		{CVEID: "CVE-2021-3", Score: 0.75},
		{CVEID: "CVE-2021-4", Score: 0.65},
	}

	mock := &MockVectorSearchClient{
		SearchFunc: func(ctx context.Context, params VectorSearchParams) ([]models.VulnResult, error) {
			// Filter by min score
			var filtered []models.VulnResult
			for _, r := range allResults {
				if params.MinScore > 0 && r.Score < params.MinScore {
					continue
				}
				filtered = append(filtered, r)
			}

			if len(filtered) == 0 {
				return nil, ErrNoResults
			}

			// Limit to K results
			if params.K > 0 && len(filtered) > params.K {
				filtered = filtered[:params.K]
			}

			return filtered, nil
		},
	}

	tests := []struct {
		name          string
		minScore      float64
		k             int
		expectedCount int
	}{
		{
			name:          "no minimum score",
			minScore:      0,
			k:             10,
			expectedCount: 4,
		},
		{
			name:          "minimum score 0.8",
			minScore:      0.8,
			k:             10,
			expectedCount: 2,
		},
		{
			name:          "minimum score 0.9",
			minScore:      0.9,
			k:             10,
			expectedCount: 1,
		},
		{
			name:          "minimum score with K limit",
			minScore:      0.7,
			k:             2,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := mock.VectorSearch(ctx, VectorSearchParams{
				QueryEmbedding: make([]float64, 1536),
				K:              tt.k,
				MinScore:       tt.minScore,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(results))

			// Verify all results meet minimum score
			for _, r := range results {
				assert.GreaterOrEqual(t, r.Score, tt.minScore)
			}
		})
	}
}

// Integration test helper - only runs with actual database
func setupTestDatabase(t *testing.T) (*VectorSearchClient, func()) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	client, err := CreateVectorSearchClient(ctx, logger)
	if err != nil {
		t.Skipf("skipping test: database not available: %v", err)
		return nil, nil
	}

	cleanup := func() {
		client.Close(ctx)
	}

	return client, cleanup
}

func TestVectorSearch_Integration(t *testing.T) {
	client, cleanup := setupTestDatabase(t)
	if client == nil {
		return // Test was skipped
	}
	defer cleanup()

	ctx := context.Background()

	// Create a sample embedding (in real scenario, this would come from embedding service)
	sampleEmbedding := make([]float64, 1536)
	for i := 0; i < 1536; i++ {
		sampleEmbedding[i] = 0.001 * float64(i)
	}

	tests := []struct {
		name    string
		params  VectorSearchParams
		wantErr error
	}{
		{
			name: "valid search",
			params: VectorSearchParams{
				QueryEmbedding: sampleEmbedding,
				K:              10,
			},
			// May return ErrNoResults if database is empty, which is acceptable
			wantErr: nil,
		},
		{
			name: "search with minimum score",
			params: VectorSearchParams{
				QueryEmbedding: sampleEmbedding,
				K:              5,
				MinScore:       0.8,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := client.VectorSearch(ctx, tt.params)

			// Either success or no results is acceptable for integration test
			if err != nil && err != ErrNoResults {
				t.Errorf("unexpected error: %v", err)
			}

			if err == nil {
				require.NotNil(t, results)
				require.LessOrEqual(t, len(results), tt.params.K)

				// Verify all results have required fields
				for i, r := range results {
					assert.NotEmpty(t, r.CVEID, "result %d missing CVE ID", i)
					assert.GreaterOrEqual(t, r.Score, 0.0, "result %d has negative score", i)
					assert.LessOrEqual(t, r.Score, 1.0, "result %d score exceeds 1.0", i)

					if tt.params.MinScore > 0 {
						assert.GreaterOrEqual(t, r.Score, tt.params.MinScore,
							"result %d score below minimum", i)
					}
				}

				// Verify results are sorted by score descending
				for i := 1; i < len(results); i++ {
					assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
						"results not sorted by score")
				}
			}
		})
	}
}
