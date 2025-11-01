package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// VersionInfo holds version information
type VersionInfo struct {
	Version   string `json:"version" yaml:"version"`
	GitCommit string `json:"git_commit" yaml:"git_commit"`
	BuildDate string `json:"build_date" yaml:"build_date"`
	GoVersion string `json:"go_version" yaml:"go_version"`
}

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	var outputFormat string

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display version, git commit, and build information for the Spectra CLI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := VersionInfo{
				Version:   Version,
				GitCommit: GitCommit,
				BuildDate: BuildDate,
				GoVersion: "go1.25.3", // From go.mod
			}

			switch outputFormat {
			case "json":
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(info); err != nil {
					return fmt.Errorf("failed to encode JSON: %w", err)
				}
			case "yaml":
				encoder := yaml.NewEncoder(cmd.OutOrStdout())
				if err := encoder.Encode(info); err != nil {
					return fmt.Errorf("failed to encode YAML: %w", err)
				}
			default:
				// Default text format
				fmt.Fprintf(cmd.OutOrStdout(), "Spectra CLI\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  Version:    %s\n", info.Version)
				fmt.Fprintf(cmd.OutOrStdout(), "  Git Commit: %s\n", info.GitCommit)
				fmt.Fprintf(cmd.OutOrStdout(), "  Build Date: %s\n", info.BuildDate)
				fmt.Fprintf(cmd.OutOrStdout(), "  Go Version: %s\n", info.GoVersion)
			}

			return nil
		},
	}

	versionCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, yaml)")

	return versionCmd
}
