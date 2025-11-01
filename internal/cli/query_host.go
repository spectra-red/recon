package cli

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/client"
	"github.com/spectra-red/recon/internal/models"
)

var (
	hostDepth int
)

var hostQueryCmd = &cobra.Command{
	Use:   "host <ip>",
	Short: "Query host information by IP address",
	Long: `Query detailed information about a host by its IP address.

The depth parameter controls how much related data is retrieved:
  0 - Host information only
  1 - Host + ports
  2 - Host + ports + services (default)
  3 - Host + ports + services + vulnerabilities
  4-5 - Extended relationships

Examples:
  # Query host with default depth (2)
  spectra query host 1.2.3.4

  # Query with full vulnerability information
  spectra query host 1.2.3.4 --depth 3

  # Output as JSON
  spectra query host 1.2.3.4 --output json

  # Output as YAML without colors
  spectra query host 1.2.3.4 --output yaml --no-color`,
	Args: cobra.ExactArgs(1),
	Run:  runHostQuery,
}

func init() {
	hostQueryCmd.Flags().IntVarP(&hostDepth, "depth", "d", int(models.DefaultDepth()), "Query depth (0-5)")
}

func runHostQuery(cmd *cobra.Command, args []string) {
	ip := args[0]

	// Validate IP address
	if net.ParseIP(ip) == nil {
		handleError(fmt.Errorf("invalid IP address: %s", ip), "")
	}

	// Validate depth
	if !models.ValidateDepth(hostDepth) {
		handleError(fmt.Errorf("depth must be between 0 and 5, got %d", hostDepth), "")
	}

	// Get API URL
	baseURL := getAPIURL()

	// Create client
	queryClient := client.NewQueryClient(baseURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute query
	result, err := queryClient.QueryHost(ctx, ip, hostDepth)
	if err != nil {
		handleError(err, "failed to query host")
	}

	// Format and output result
	opts := getOutputOptions()
	formatter := NewFormatter()

	if err := formatter.FormatHostQuery(opts, result); err != nil {
		handleError(err, "failed to format output")
	}
}
