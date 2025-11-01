package cli

import (
	"github.com/spf13/cobra"
)

// NewJobsCommand creates the jobs command with subcommands
func NewJobsCommand() *cobra.Command {
	jobsCmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage scan ingestion jobs",
		Long: `Manage and monitor scan ingestion jobs.

When you submit scan results via 'spectra ingest', they are processed asynchronously
as background jobs. Use these commands to check job status, view progress, and retrieve
results.

Jobs have the following states:
  - pending: Job is queued and waiting to be processed
  - processing: Job is currently being processed
  - completed: Job has completed successfully
  - failed: Job encountered an error during processing`,
		Example: `  # List all jobs
  spectra jobs list

  # List only failed jobs
  spectra jobs list --state failed

  # Get details for a specific job
  spectra jobs get <job-id>

  # Watch a job until completion
  spectra jobs get <job-id> --watch`,
	}

	// Add subcommands
	jobsCmd.AddCommand(NewJobsListCommand())
	jobsCmd.AddCommand(NewJobsGetCommand())

	return jobsCmd
}
