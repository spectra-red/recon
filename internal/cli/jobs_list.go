package cli

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/client"
	"github.com/spectra-red/recon/internal/models"
)

var (
	listScannerKey string
	listState      string
	listLimit      int
	listOffset     int
	listOrderBy    string
	listNoColor    bool
)

// NewJobsListCommand creates the jobs list subcommand
func NewJobsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List scan ingestion jobs",
		Long: `List scan ingestion jobs with optional filters and pagination.

The list command displays all jobs in the system, with support for filtering
by scanner key, job state, and pagination. Output can be formatted as JSON,
YAML, or a table (default).`,
		Example: `  # List all jobs (default: table format)
  spectra jobs list

  # List jobs in JSON format
  spectra jobs list --output json

  # List only completed jobs
  spectra jobs list --state completed

  # List jobs from a specific scanner
  spectra jobs list --scanner <public-key>

  # List with pagination
  spectra jobs list --limit 100 --offset 50

  # Combine filters
  spectra jobs list --state processing --limit 20`,
		RunE: runJobsList,
	}

	// Add flags
	cmd.Flags().StringVar(&listScannerKey, "scanner", "", "Filter by scanner public key")
	cmd.Flags().StringVar(&listState, "state", "", "Filter by job state (pending, processing, completed, failed)")
	cmd.Flags().IntVar(&listLimit, "limit", 50, "Maximum number of results (max: 500)")
	cmd.Flags().IntVar(&listOffset, "offset", 0, "Offset for pagination")
	cmd.Flags().StringVar(&listOrderBy, "order-by", "created_at", "Order by field (created_at, updated_at)")
	cmd.Flags().BoolVar(&listNoColor, "no-color", false, "Disable colored output")

	return cmd
}

func runJobsList(cmd *cobra.Command, args []string) error {
	// Get output format from viper/config
	format := GetOutputFormat()

	// Build options
	opts := client.ListJobsOptions{
		Limit:     listLimit,
		Offset:    listOffset,
		OrderBy:   listOrderBy,
		OrderDesc: true,
	}

	// Add scanner key filter if provided
	if listScannerKey != "" {
		opts.ScannerKey = &listScannerKey
	}

	// Add state filter if provided
	if listState != "" {
		state := models.JobState(listState)
		if !state.IsValid() {
			return fmt.Errorf("invalid state: %s (must be one of: pending, processing, completed, failed)", listState)
		}
		opts.State = &state
	}

	// Create client and fetch jobs
	ctx, cancel := context.WithTimeout(context.Background(), GetAPITimeout())
	defer cancel()

	apiClient := client.NewClient(GetAPIURL()).WithTimeout(GetAPITimeout())
	resp, err := apiClient.ListJobs(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	// Format and output results
	outputOpts := NewOutputOptions(format, listNoColor)

	switch outputOpts.Format {
	case FormatJSON:
		return formatJSON(outputOpts.Writer, resp)
	case FormatYAML:
		return formatYAML(outputOpts.Writer, resp)
	case FormatTable:
		return formatJobsListTable(outputOpts, resp)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func formatJobsListTable(opts *OutputOptions, resp *models.JobListResponse) error {
	if len(resp.Jobs) == 0 {
		fmt.Fprintln(opts.Writer, "No jobs found")
		return nil
	}

	headerColor := color.New(color.FgCyan, color.Bold)

	// Print header
	if !opts.NoColor && opts.IsTerminal {
		headerColor.Fprintf(opts.Writer, "\nScan Ingestion Jobs\n\n")
	} else {
		fmt.Fprintf(opts.Writer, "\nScan Ingestion Jobs\n\n")
	}

	// Create table
	table := tablewriter.NewWriter(opts.Writer)
	table.SetHeader([]string{"Job ID", "State", "Scanner", "Created", "Updated", "Hosts", "Ports"})
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	// Add rows
	for _, job := range resp.Jobs {
		state := job.State.String()
		if !opts.NoColor && opts.IsTerminal {
			state = colorizeJobState(job.State)
		}

		table.Append([]string{
			truncate(job.ID, 36),
			state,
			maskScannerKey(job.ScannerKey),
			formatTime(job.CreatedAt),
			formatTime(job.UpdatedAt),
			formatCount(job.HostCount),
			formatCount(job.PortCount),
		})
	}

	table.Render()

	// Print pagination info
	fmt.Fprintf(opts.Writer, "\nShowing %d-%d of %d jobs\n",
		resp.Offset+1,
		resp.Offset+len(resp.Jobs),
		resp.Total)

	if resp.HasMore {
		fmt.Fprintf(opts.Writer, "Use --offset %d to see more results\n", resp.NextOffset)
	}

	return nil
}

// colorizeJobState adds color to job state display
func colorizeJobState(state models.JobState) string {
	switch state {
	case models.JobStateCompleted:
		return color.GreenString(state.String())
	case models.JobStateFailed:
		return color.RedString(state.String())
	case models.JobStateProcessing:
		return color.YellowString(state.String())
	case models.JobStatePending:
		return color.CyanString(state.String())
	default:
		return state.String()
	}
}

// maskScannerKey masks the scanner key for display
func maskScannerKey(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:8] + "..." + key[len(key)-4:]
}

// formatCount formats an integer count for display
func formatCount(count int) string {
	if count == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", count)
}
