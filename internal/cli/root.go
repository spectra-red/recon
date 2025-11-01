package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version information (set via ldflags at build time)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// Global flags
	cfgFile string
	apiURL  string
	verbose bool
)

// NewRootCommand creates and returns the root command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "spectra",
		Short: "Spectra-Red security intelligence CLI",
		Long: `Spectra-Red Intel Mesh - Community-Driven Security Intelligence

The Spectra CLI allows you to:
  - Ingest scan results into the mesh
  - Query threat intelligence data
  - Manage background jobs
  - View and analyze security information

Configuration precedence: flags > environment variables > config file > defaults

Environment Variables:
  SPECTRA_API_URL      API endpoint URL
  SPECTRA_CONFIG       Path to config file
  SPECTRA_OUTPUT_FORMAT Output format (json, yaml, table)

For more information, visit: https://github.com/spectra-red/recon`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize configuration
			cfg, err := InitConfig(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Override with flags if provided
			if cmd.Flags().Changed("api-url") {
				viper.Set("api.url", apiURL)
			}

			// Validate configuration
			if err := ValidateConfig(cfg); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			// Set verbose mode
			if verbose {
				fmt.Fprintf(os.Stderr, "Config file: %s\n", viper.ConfigFileUsed())
				fmt.Fprintf(os.Stderr, "API URL: %s\n", GetAPIURL())
			}

			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.spectra.yaml, ~/.spectra/.spectra.yaml, or /etc/spectra/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "API endpoint URL (default: http://localhost:3000)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Bind flags to viper
	viper.BindPFlag("api.url", rootCmd.PersistentFlags().Lookup("api-url"))

	// Add subcommands
	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewIngestCommand())
	rootCmd.AddCommand(NewQueryCommand())
	rootCmd.AddCommand(NewJobsCommand())

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	rootCmd := NewRootCommand()
	return rootCmd.Execute()
}
