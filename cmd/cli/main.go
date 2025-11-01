package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spectra",
	Short: "Spectra-Red security intelligence CLI",
	Long: `Spectra-Red is a community-driven security intelligence mesh.

Use this CLI to scan targets, query the mesh, and contribute to the community.`,
	Version: "0.1.0",
}

func init() {
	// Subcommands will be added here
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
