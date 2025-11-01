package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/cli"
)

var rootCmd = &cobra.Command{
	Use:   "spectra",
	Short: "Spectra-Red security intelligence CLI",
	Long: `Spectra-Red is a community-driven security intelligence mesh.

Use this CLI to scan targets, query the mesh, and contribute to the community.`,
	Version: "0.1.0",
}

func init() {
	// Create jobs command group
	jobsCmd := cli.NewJobsCommand()

	// Add subcommands
	jobsCmd.AddCommand(cli.NewJobsListCommand())
	jobsCmd.AddCommand(cli.NewJobsGetCommand())

	// Register commands with root command
	rootCmd.AddCommand(jobsCmd)
	rootCmd.AddCommand(cli.NewIngestCommand())
	rootCmd.AddCommand(cli.QueryCmd)

	// Future commands will be added here
	// rootCmd.AddCommand(scanCmd)
	// rootCmd.AddCommand(meshCmd)
	// rootCmd.AddCommand(authCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
