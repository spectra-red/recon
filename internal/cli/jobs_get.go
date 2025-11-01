package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/client"
	"github.com/spectra-red/recon/internal/models"
)

var (
	getWatch      bool
	getInterval   string
	getNoColor    bool
)

// NewJobsGetCommand creates the jobs get subcommand
func NewJobsGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <job-id>",
		Short: "Get details for a specific job",
		Long: `Get detailed information about a specific scan ingestion job.

The get command displays the current state, timestamps, error messages (if any),
and statistics for a job. With the --watch flag, it will continuously poll the
job status until it reaches a terminal state (completed or failed).`,
		Example: `  # Get job details
  spectra jobs get 01933e8a-7b2c-7890-9abc-def012345678

  # Get job details in JSON format
  spectra jobs get 01933e8a-7b2c-7890-9abc-def012345678 --output json

  # Watch job until completion (polls every 2 seconds)
  spectra jobs get 01933e8a-7b2c-7890-9abc-def012345678 --watch

  # Watch with custom interval
  spectra jobs get 01933e8a-7b2c-7890-9abc-def012345678 --watch --interval 5s`,
		Args: cobra.ExactArgs(1),
		RunE: runJobsGet,
	}

	// Add flags
	cmd.Flags().BoolVarP(&getWatch, "watch", "w", false, "Watch job until completion")
	cmd.Flags().StringVar(&getInterval, "interval", "2s", "Polling interval for watch mode")
	cmd.Flags().BoolVar(&getNoColor, "no-color", false, "Disable colored output")

	return cmd
}

func runJobsGet(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	// Get output format from config
	format := GetOutputFormat()

	// Parse watch interval
	interval, err := time.ParseDuration(getInterval)
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	if interval < time.Second {
		return fmt.Errorf("interval must be at least 1s")
	}

	// Create client
	apiClient := client.NewClient(GetAPIURL()).WithTimeout(GetAPITimeout())

	// If watch mode, continuously poll until terminal state
	if getWatch {
		return watchJob(apiClient, jobID, interval, format)
	}

	// Otherwise, fetch once and display
	ctx, cancel := context.WithTimeout(context.Background(), GetAPITimeout())
	defer cancel()

	job, err := apiClient.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	return formatJob(job, format)
}

func watchJob(apiClient *client.Client, jobID string, interval time.Duration, format string) error {
	outputOpts := NewOutputOptions(format, getNoColor)

	// Only show watch progress in terminal and table mode
	showProgress := outputOpts.IsTerminal && outputOpts.Format == FormatTable

	if showProgress {
		headerColor := color.New(color.FgCyan, color.Bold)
		headerColor.Printf("Watching job %s (polling every %s)...\n\n", jobID, interval)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastState models.JobState

	for {
		ctx, cancel := context.WithTimeout(context.Background(), GetAPITimeout())
		job, err := apiClient.GetJob(ctx, jobID)
		cancel()

		if err != nil {
			return fmt.Errorf("failed to get job: %w", err)
		}

		// Print update if state changed
		if showProgress && job.State != lastState {
			timestamp := time.Now().Format("15:04:05")
			stateStr := colorizeJobState(job.State)
			fmt.Printf("[%s] State: %s\n", timestamp, stateStr)
			lastState = job.State

			// Print additional info for terminal states
			if job.State == models.JobStateCompleted {
				successColor := color.New(color.FgGreen, color.Bold)
				successColor.Println("Job completed successfully!")
				fmt.Printf("  Hosts processed: %d\n", job.HostCount)
				fmt.Printf("  Ports processed: %d\n", job.PortCount)
				if job.CompletedAt != nil {
					duration := job.CompletedAt.Sub(job.CreatedAt)
					fmt.Printf("  Duration: %s\n", formatDuration(duration))
				}
				fmt.Println()
			} else if job.State == models.JobStateFailed {
				errorColor := color.New(color.FgRed, color.Bold)
				errorColor.Println("Job failed!")
				if job.ErrorMessage != nil {
					fmt.Printf("  Error: %s\n", *job.ErrorMessage)
				}
				fmt.Println()
			}
		}

		// Check if we've reached a terminal state
		if job.State == models.JobStateCompleted || job.State == models.JobStateFailed {
			return formatJob(job, format)
		}

		// Wait for next tick
		<-ticker.C
	}
}

func formatJob(job *models.Job, format string) error {
	outputOpts := NewOutputOptions(format, getNoColor)

	switch outputOpts.Format {
	case FormatJSON:
		return formatJSON(outputOpts.Writer, job)
	case FormatYAML:
		return formatYAML(outputOpts.Writer, job)
	case FormatTable:
		return formatJobDetail(outputOpts, job)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func formatJobDetail(opts *OutputOptions, job *models.Job) error {
	headerColor := color.New(color.FgCyan, color.Bold)
	labelColor := color.New(color.FgWhite, color.Bold)

	// Header
	if !opts.NoColor && opts.IsTerminal {
		headerColor.Fprintln(opts.Writer, "\nJob Details")
		fmt.Fprintln(opts.Writer, "===========")
	} else {
		fmt.Fprintln(opts.Writer, "\nJob Details")
		fmt.Fprintln(opts.Writer, "===========")
	}
	fmt.Fprintln(opts.Writer)

	// Basic information
	if !opts.NoColor && opts.IsTerminal {
		labelColor.Fprint(opts.Writer, "ID:           ")
		fmt.Fprintln(opts.Writer, job.ID)

		labelColor.Fprint(opts.Writer, "State:        ")
		fmt.Fprintln(opts.Writer, colorizeJobState(job.State))

		labelColor.Fprint(opts.Writer, "Scanner Key:  ")
		fmt.Fprintln(opts.Writer, maskScannerKey(job.ScannerKey))
	} else {
		fmt.Fprintf(opts.Writer, "ID:           %s\n", job.ID)
		fmt.Fprintf(opts.Writer, "State:        %s\n", job.State)
		fmt.Fprintf(opts.Writer, "Scanner Key:  %s\n", maskScannerKey(job.ScannerKey))
	}

	// Timestamps
	fmt.Fprintf(opts.Writer, "Created:      %s\n", formatTime(job.CreatedAt))
	fmt.Fprintf(opts.Writer, "Updated:      %s\n", formatTime(job.UpdatedAt))

	if job.CompletedAt != nil {
		fmt.Fprintf(opts.Writer, "Completed:    %s\n", formatTime(*job.CompletedAt))
		duration := job.CompletedAt.Sub(job.CreatedAt)
		fmt.Fprintf(opts.Writer, "Duration:     %s\n", formatDuration(duration))
	}

	// Statistics
	fmt.Fprintln(opts.Writer)
	if !opts.NoColor && opts.IsTerminal {
		headerColor.Fprintln(opts.Writer, "Statistics")
		fmt.Fprintln(opts.Writer, "----------")
	} else {
		fmt.Fprintln(opts.Writer, "Statistics")
		fmt.Fprintln(opts.Writer, "----------")
	}
	fmt.Fprintf(opts.Writer, "Hosts:        %d\n", job.HostCount)
	fmt.Fprintf(opts.Writer, "Ports:        %d\n", job.PortCount)

	// Error message if present
	if job.ErrorMessage != nil {
		fmt.Fprintln(opts.Writer)
		if !opts.NoColor && opts.IsTerminal {
			errorColor := color.New(color.FgRed, color.Bold)
			errorColor.Fprintln(opts.Writer, "Error")
			fmt.Fprintln(opts.Writer, "-----")
		} else {
			fmt.Fprintln(opts.Writer, "Error")
			fmt.Fprintln(opts.Writer, "-----")
		}
		fmt.Fprintf(opts.Writer, "%s\n", *job.ErrorMessage)
	}

	fmt.Fprintln(opts.Writer)
	return nil
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}

	// Round to milliseconds for short durations
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	// Round to seconds
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm %ds", minutes, seconds%60)
	}

	hours := minutes / 60
	return fmt.Sprintf("%dh %dm", hours, minutes%60)
}
