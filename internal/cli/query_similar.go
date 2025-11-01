package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/client"
	"github.com/spectra-red/recon/internal/models"
)

var (
	similarK int
)

var similarQueryCmd = &cobra.Command{
	Use:   "similar <text>",
	Short: "Search for similar vulnerabilities using vector similarity",
	Long: `Search for similar vulnerabilities using natural language and vector similarity.

This command uses AI embeddings to find vulnerabilities that are semantically
similar to your query text. It's useful for:
  - Finding related CVEs based on description
  - Discovering similar attack vectors
  - Identifying comparable vulnerabilities

The similarity score ranges from 0.0 to 1.0, with higher scores indicating
greater similarity.

Examples:
  # Search for nginx vulnerabilities
  spectra query similar "nginx remote code execution"

  # Search for buffer overflow issues
  spectra query similar "buffer overflow privilege escalation"

  # Get more results
  spectra query similar "SQL injection" --k 20

  # Output as JSON
  spectra query similar "XSS vulnerability" --output json

  # Disable colored output
  spectra query similar "authentication bypass" --no-color`,
	Args: cobra.MinimumNArgs(1),
	Run:  runSimilarQuery,
}

func init() {
	similarQueryCmd.Flags().IntVarP(&similarK, "k", "k", models.DefaultK, fmt.Sprintf("Number of results to return (1-%d)", models.MaxK))
}

func runSimilarQuery(cmd *cobra.Command, args []string) {
	// Join all arguments as the query text
	queryText := ""
	for i, arg := range args {
		if i > 0 {
			queryText += " "
		}
		queryText += arg
	}

	// Validate K
	if similarK < 1 || similarK > models.MaxK {
		handleError(fmt.Errorf("k must be between 1 and %d, got %d", models.MaxK, similarK), "")
	}

	// Create request
	req := client.NewSimilarRequest(queryText, similarK)

	// Validate request
	if err := req.Validate(); err != nil {
		handleError(err, "invalid request")
	}

	// Get API URL
	baseURL := getAPIURL()

	// Create client
	queryClient := client.NewQueryClient(baseURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute query
	result, err := queryClient.SimilarQuery(ctx, req)
	if err != nil {
		handleError(err, "failed to execute similarity search")
	}

	// Format and output result
	opts := getOutputOptions()
	formatter := NewFormatter()

	if err := formatter.FormatSimilarQuery(opts, result); err != nil {
		handleError(err, "failed to format output")
	}
}
