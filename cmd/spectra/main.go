package main

import (
	"fmt"
	"os"

	"github.com/spectra-red/recon/internal/cli"
)

// Version information (set via ldflags at build time)
var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	// Set version info for the CLI package
	cli.Version = version
	cli.GitCommit = gitCommit
	cli.BuildDate = buildDate

	// Execute the root command
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
