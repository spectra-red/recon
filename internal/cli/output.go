package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
	"github.com/spectra-red/recon/internal/models"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatTable OutputFormat = "table"
)

// OutputOptions controls output formatting behavior
type OutputOptions struct {
	Format     OutputFormat
	NoColor    bool
	Writer     io.Writer
	IsTerminal bool
}

// NewOutputOptions creates output options with sensible defaults
func NewOutputOptions(format string, noColor bool) *OutputOptions {
	opts := &OutputOptions{
		Format:  FormatTable, // Default to table
		NoColor: noColor,
		Writer:  os.Stdout,
	}

	// Check if output is a terminal
	if f, ok := opts.Writer.(*os.File); ok {
		opts.IsTerminal = isatty.IsTerminal(f.Fd())
	} else {
		opts.IsTerminal = false
	}

	// Parse format
	switch strings.ToLower(format) {
	case "json":
		opts.Format = FormatJSON
	case "yaml", "yml":
		opts.Format = FormatYAML
	case "table":
		opts.Format = FormatTable
	default:
		opts.Format = FormatTable
	}

	// Disable color if not a terminal or explicitly disabled
	if !opts.IsTerminal || noColor {
		color.NoColor = true
	}

	return opts
}

// OutputFormatter is the interface for formatting different data types
type OutputFormatter interface {
	FormatHostQuery(opts *OutputOptions, result *models.HostQueryResponse) error
	FormatGraphQuery(opts *OutputOptions, result *models.GraphQueryResponse) error
	FormatSimilarQuery(opts *OutputOptions, result *models.SimilarResponse) error
}

// DefaultFormatter implements OutputFormatter
type DefaultFormatter struct{}

// NewFormatter creates a new output formatter
func NewFormatter() OutputFormatter {
	return &DefaultFormatter{}
}

// FormatHostQuery formats a host query response
func (f *DefaultFormatter) FormatHostQuery(opts *OutputOptions, result *models.HostQueryResponse) error {
	switch opts.Format {
	case FormatJSON:
		return formatJSON(opts.Writer, result)
	case FormatYAML:
		return formatYAML(opts.Writer, result)
	case FormatTable:
		return formatHostTable(opts, result)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// FormatGraphQuery formats a graph query response
func (f *DefaultFormatter) FormatGraphQuery(opts *OutputOptions, result *models.GraphQueryResponse) error {
	switch opts.Format {
	case FormatJSON:
		return formatJSON(opts.Writer, result)
	case FormatYAML:
		return formatYAML(opts.Writer, result)
	case FormatTable:
		return formatGraphTable(opts, result)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// FormatSimilarQuery formats a similarity search response
func (f *DefaultFormatter) FormatSimilarQuery(opts *OutputOptions, result *models.SimilarResponse) error {
	switch opts.Format {
	case FormatJSON:
		return formatJSON(opts.Writer, result)
	case FormatYAML:
		return formatYAML(opts.Writer, result)
	case FormatTable:
		return formatSimilarTable(opts, result)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// formatJSON outputs data as JSON
func formatJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatYAML outputs data as YAML
func formatYAML(w io.Writer, data interface{}) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}

// formatHostTable formats host query results as a table
func formatHostTable(opts *OutputOptions, result *models.HostQueryResponse) error {
	// Header information
	headerColor := color.New(color.FgCyan, color.Bold)
	if !opts.NoColor && opts.IsTerminal {
		headerColor.Fprintf(opts.Writer, "\nHost: %s\n", result.IP)
		fmt.Fprintf(opts.Writer, "ASN: %d | City: %s | Country: %s\n",
			result.ASN, result.City, result.Country)
		fmt.Fprintf(opts.Writer, "First Seen: %s | Last Seen: %s\n\n",
			formatTime(result.FirstSeen), formatTime(result.LastSeen))
	} else {
		fmt.Fprintf(opts.Writer, "\nHost: %s\n", result.IP)
		fmt.Fprintf(opts.Writer, "ASN: %d | City: %s | Country: %s\n",
			result.ASN, result.City, result.Country)
		fmt.Fprintf(opts.Writer, "First Seen: %s | Last Seen: %s\n\n",
			formatTime(result.FirstSeen), formatTime(result.LastSeen))
	}

	// Ports table
	if len(result.Ports) > 0 {
		table := tablewriter.NewWriter(opts.Writer)
		table.SetHeader([]string{"Port", "Protocol", "Service", "Product", "Version"})
		table.SetBorder(true)

		for _, port := range result.Ports {
			serviceName := ""
			product := ""
			version := ""

			if len(port.Services) > 0 {
				serviceName = port.Services[0].Name
				product = port.Services[0].Product
				version = port.Services[0].Version
			}

			table.Append([]string{
				fmt.Sprintf("%d", port.Number),
				port.Protocol,
				serviceName,
				product,
				version,
			})
		}

		table.Render()
		fmt.Fprintln(opts.Writer)
	}

	// Vulnerabilities table
	if len(result.Vulns) > 0 {
		if !opts.NoColor && opts.IsTerminal {
			headerColor.Fprintln(opts.Writer, "Vulnerabilities:")
		} else {
			fmt.Fprintln(opts.Writer, "Vulnerabilities:")
		}

		table := tablewriter.NewWriter(opts.Writer)
		table.SetHeader([]string{"CVE ID", "CVSS", "Severity", "KEV", "First Detected"})
		table.SetBorder(true)

		for _, vuln := range result.Vulns {
			kevFlag := "No"
			if vuln.KEVFlag {
				kevFlag = "Yes"
			}

			severity := vuln.Severity
			if !opts.NoColor && opts.IsTerminal {
				severity = colorSeverity(vuln.Severity)
			}

			table.Append([]string{
				vuln.CVEID,
				fmt.Sprintf("%.1f", vuln.CVSS),
				severity,
				kevFlag,
				formatTime(vuln.FirstSeen),
			})
		}

		table.Render()
		fmt.Fprintln(opts.Writer)
	}

	return nil
}

// formatGraphTable formats graph query results as a table
func formatGraphTable(opts *OutputOptions, result *models.GraphQueryResponse) error {
	headerColor := color.New(color.FgCyan, color.Bold)

	// Summary
	if !opts.NoColor && opts.IsTerminal {
		headerColor.Fprintf(opts.Writer, "\nGraph Query Results\n")
	} else {
		fmt.Fprintf(opts.Writer, "\nGraph Query Results\n")
	}

	fmt.Fprintf(opts.Writer, "Results: %d | Query Time: %.2f ms\n\n",
		len(result.Results), result.QueryTime)

	if len(result.Results) == 0 {
		fmt.Fprintln(opts.Writer, "No results found.")
		return nil
	}

	// Results table
	table := tablewriter.NewWriter(opts.Writer)
	table.SetHeader([]string{"IP", "ASN", "City", "Country", "Ports", "Services", "Last Seen"})
	table.SetBorder(true)

	for _, host := range result.Results {
		portCount := len(host.Ports)
		serviceCount := len(host.Services)

		table.Append([]string{
			host.IP,
			fmt.Sprintf("%d", host.ASN),
			host.City,
			host.Country,
			fmt.Sprintf("%d", portCount),
			fmt.Sprintf("%d", serviceCount),
			formatTime(host.LastSeen),
		})
	}

	table.Render()

	// Pagination info
	if result.Pagination.HasMore {
		fmt.Fprintf(opts.Writer, "\nMore results available. Use --offset %d to continue.\n",
			result.Pagination.NextOffset)
	}

	return nil
}

// formatSimilarTable formats similarity search results as a table
func formatSimilarTable(opts *OutputOptions, result *models.SimilarResponse) error {
	headerColor := color.New(color.FgCyan, color.Bold)

	// Header
	if !opts.NoColor && opts.IsTerminal {
		headerColor.Fprintf(opts.Writer, "\nSimilarity Search: %s\n", result.Query)
	} else {
		fmt.Fprintf(opts.Writer, "\nSimilarity Search: %s\n", result.Query)
	}

	fmt.Fprintf(opts.Writer, "Results: %d | Time: %s\n\n", result.Count, result.Timestamp)

	if result.Count == 0 {
		fmt.Fprintln(opts.Writer, "No similar vulnerabilities found.")
		return nil
	}

	// Results table
	table := tablewriter.NewWriter(opts.Writer)
	table.SetHeader([]string{"Score", "CVE ID", "CVSS", "Title"})
	table.SetBorder(true)
	table.SetAutoWrapText(true)
	table.SetColWidth(60)

	for _, vuln := range result.Results {
		score := fmt.Sprintf("%.3f", vuln.Score)
		if !opts.NoColor && opts.IsTerminal {
			score = colorScore(vuln.Score)
		}

		table.Append([]string{
			score,
			vuln.CVEID,
			fmt.Sprintf("%.1f", vuln.CVSS),
			truncate(vuln.Title, 60),
		})
	}

	table.Render()

	return nil
}

// Helper functions

// formatTime formats a time.Time for display
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04")
}

// colorSeverity returns colored severity text
func colorSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return color.RedString(severity)
	case "high":
		return color.New(color.FgRed).Sprint(severity)
	case "medium":
		return color.YellowString(severity)
	case "low":
		return color.GreenString(severity)
	default:
		return severity
	}
}

// colorScore returns colored similarity score
func colorScore(score float64) string {
	scoreStr := fmt.Sprintf("%.3f", score)
	switch {
	case score >= 0.9:
		return color.GreenString(scoreStr)
	case score >= 0.8:
		return color.YellowString(scoreStr)
	case score >= 0.7:
		return color.New(color.FgYellow).Sprint(scoreStr)
	default:
		return scoreStr
	}
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
