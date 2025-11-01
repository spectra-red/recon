package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/client"
	"github.com/spectra-red/recon/internal/models"
)

var (
	graphType    string
	graphValue   string
	graphLimit   int
	graphOffset  int
	graphCity    string
	graphRegion  string
	graphCountry string
	graphProduct string
	graphService string
)

var graphQueryCmd = &cobra.Command{
	Use:   "graph",
	Short: "Execute advanced graph traversal queries",
	Long: `Execute advanced graph traversal queries across the intelligence mesh.

Query Types:
  by_asn      - Find hosts by Autonomous System Number
  by_location - Find hosts by geographic location
  by_vuln     - Find hosts affected by a specific CVE
  by_service  - Find hosts running a specific service

Examples:
  # Query by ASN
  spectra query graph --type by_asn --value 16509 --limit 100

  # Query by location (city)
  spectra query graph --type by_location --city "San Francisco"

  # Query by location (country)
  spectra query graph --type by_location --country "United States"

  # Query by vulnerability
  spectra query graph --type by_vuln --value CVE-2024-1234

  # Query by service product
  spectra query graph --type by_service --product nginx

  # With pagination
  spectra query graph --type by_asn --value 16509 --limit 50 --offset 50

  # Output as JSON
  spectra query graph --type by_vuln --value CVE-2024-1234 --output json`,
	Run: runGraphQuery,
}

func init() {
	graphQueryCmd.Flags().StringVar(&graphType, "type", "", "Query type (by_asn, by_location, by_vuln, by_service)")
	graphQueryCmd.Flags().StringVar(&graphValue, "value", "", "Query value (ASN number or CVE ID)")
	graphQueryCmd.Flags().IntVar(&graphLimit, "limit", 100, "Maximum number of results (1-1000)")
	graphQueryCmd.Flags().IntVar(&graphOffset, "offset", 0, "Offset for pagination")

	// Location-specific flags
	graphQueryCmd.Flags().StringVar(&graphCity, "city", "", "City name for location queries")
	graphQueryCmd.Flags().StringVar(&graphRegion, "region", "", "Region name for location queries")
	graphQueryCmd.Flags().StringVar(&graphCountry, "country", "", "Country name for location queries")

	// Service-specific flags
	graphQueryCmd.Flags().StringVar(&graphProduct, "product", "", "Product name for service queries (e.g., 'nginx')")
	graphQueryCmd.Flags().StringVar(&graphService, "service", "", "Service name for service queries (e.g., 'http')")

	graphQueryCmd.MarkFlagRequired("type")
}

func runGraphQuery(cmd *cobra.Command, args []string) {
	// Validate query type
	var queryType models.GraphQueryType
	switch graphType {
	case "by_asn":
		queryType = models.QueryByASN
	case "by_location":
		queryType = models.QueryByLocation
	case "by_vuln":
		queryType = models.QueryByVuln
	case "by_service":
		queryType = models.QueryByService
	default:
		handleError(fmt.Errorf("invalid query type: %s", graphType), "must be one of: by_asn, by_location, by_vuln, by_service")
	}

	// Validate limit
	if graphLimit < 1 || graphLimit > 1000 {
		handleError(fmt.Errorf("limit must be between 1 and 1000, got %d", graphLimit), "")
	}

	// Build request based on query type
	var req *models.GraphQueryRequest

	switch queryType {
	case models.QueryByASN:
		if graphValue == "" {
			handleError(fmt.Errorf("--value is required for by_asn queries"), "")
		}
		asn, err := strconv.Atoi(graphValue)
		if err != nil {
			handleError(fmt.Errorf("invalid ASN: %s", graphValue), "ASN must be a number")
		}
		req = client.GraphQueryByASN(asn, graphLimit, graphOffset)

	case models.QueryByLocation:
		if graphCity == "" && graphRegion == "" && graphCountry == "" {
			handleError(fmt.Errorf("at least one of --city, --region, or --country is required for by_location queries"), "")
		}
		req = client.GraphQueryByLocation(graphCity, graphRegion, graphCountry, graphLimit, graphOffset)

	case models.QueryByVuln:
		if graphValue == "" {
			handleError(fmt.Errorf("--value is required for by_vuln queries"), "CVE ID required")
		}
		req = client.GraphQueryByVuln(graphValue, graphLimit, graphOffset)

	case models.QueryByService:
		if graphProduct == "" && graphService == "" {
			handleError(fmt.Errorf("at least one of --product or --service is required for by_service queries"), "")
		}
		req = client.GraphQueryByService(graphProduct, graphService, graphLimit, graphOffset)
	}

	// Get API URL
	baseURL := getAPIURL()

	// Create client
	queryClient := client.NewQueryClient(baseURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute query
	result, err := queryClient.GraphQuery(ctx, req)
	if err != nil {
		handleError(err, "failed to execute graph query")
	}

	// Format and output result
	opts := getOutputOptions()
	formatter := NewFormatter()

	if err := formatter.FormatGraphQuery(opts, result); err != nil {
		handleError(err, "failed to format output")
	}
}
