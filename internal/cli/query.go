package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// QueryCmd represents the query command group
var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query the Spectra-Red intelligence mesh",
	Long: `Query the Spectra-Red intelligence mesh for threat intelligence data.

Available subcommands:
  host    - Query host information by IP address
  graph   - Execute advanced graph traversal queries
  similar - Search for similar vulnerabilities using vector similarity

Examples:
  spectra query host 1.2.3.4
  spectra query graph --type by_asn --value 16509
  spectra query similar "nginx remote code execution"`,
}

var (
	// Global flags for all query commands
	outputFormat string
	noColor      bool
	queryAPIURL  string
)

func init() {
	// Add global flags
	QueryCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json, yaml, table)")
	QueryCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	QueryCmd.PersistentFlags().StringVar(&queryAPIURL, "api-url", "", "API base URL (overrides config)")

	// Bind flags to viper
	viper.BindPFlag("output", QueryCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("no-color", QueryCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("api.url", QueryCmd.PersistentFlags().Lookup("api-url"))

	// Add subcommands
	QueryCmd.AddCommand(hostQueryCmd)
	QueryCmd.AddCommand(graphQueryCmd)
	QueryCmd.AddCommand(similarQueryCmd)
}

// NewQueryCommand creates the query command with subcommands (for compatibility)
func NewQueryCommand() *cobra.Command {
	return QueryCmd
}

// getAPIURL returns the API URL from config or flag
func getAPIURL() string {
	if queryAPIURL != "" {
		return queryAPIURL
	}

	// Try viper config
	if url := viper.GetString("api.url"); url != "" {
		return url
	}

	// Try environment variable
	if url := os.Getenv("SPECTRA_API_URL"); url != "" {
		return url
	}

	// Default
	return "http://localhost:3000"
}

// getOutputOptions returns output options based on flags
func getOutputOptions() *OutputOptions {
	format := outputFormat
	if format == "" {
		format = viper.GetString("output")
		if format == "" {
			format = "table"
		}
	}

	nc := noColor
	if !nc {
		nc = viper.GetBool("no-color")
	}

	return NewOutputOptions(format, nc)
}

// handleError prints an error message and exits
func handleError(err error, message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", message, err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(1)
}
