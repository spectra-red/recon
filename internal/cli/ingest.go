package cli

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spectra-red/recon/internal/client"
)

// NewIngestCommand creates the ingest command
func NewIngestCommand() *cobra.Command {
	var filePath string

	ingestCmd := &cobra.Command{
		Use:   "ingest [file]",
		Short: "Submit scan results to the mesh",
		Long: `Submit scan results to the Spectra-Red Intel Mesh.

The ingest command accepts scan data in Naabu JSON format,
signs it with your private key, and submits it to the mesh for processing.

Examples:
  # Ingest from stdin (Naabu JSON)
  naabu -host example.com -json | spectra ingest -

  # Ingest from file
  spectra ingest scan-results.json

  # Ingest with explicit file flag
  spectra ingest --file scan-results.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine input source: flag, positional arg, or stdin
			inputPath := filePath
			if inputPath == "" && len(args) > 0 {
				inputPath = args[0]
			}
			if inputPath == "" {
				inputPath = "-" // default to stdin
			}

			return runIngest(inputPath)
		},
	}

	ingestCmd.Flags().StringVarP(&filePath, "file", "f", "", "Input file containing scan results (use '-' for stdin)")

	return ingestCmd
}

// runIngest executes the ingest command
func runIngest(filePath string) error {
	// Get private key from config
	privKey, err := GetPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get private key: %w\n\nHint: Run 'spectra keys generate' to create a keypair", err)
	}

	// Derive public key from private key
	pubKey := privKey.Public().(ed25519.PublicKey)

	// Read scan data
	scanData, err := readScanData(filePath)
	if err != nil {
		return fmt.Errorf("failed to read scan data: %w", err)
	}

	// Validate it's valid JSON
	if !json.Valid(scanData) {
		return fmt.Errorf("invalid JSON in scan data")
	}

	// Show progress for large files (>1MB)
	if len(scanData) > 1024*1024 {
		fmt.Fprintf(os.Stderr, "Submitting %d bytes of scan data...\n", len(scanData))
	}

	// Sign the scan data
	timestamp := time.Now().Unix()
	signature, err := signScanData(scanData, timestamp, privKey)
	if err != nil {
		return fmt.Errorf("failed to sign scan data: %w", err)
	}

	// Get config values
	apiURL := GetAPIURL()
	timeout := GetAPITimeout()
	outputFormat := GetOutputFormat()

	// Create ingest client
	ingestClient := client.NewIngestClient(apiURL, int(timeout.Seconds()))

	// Submit to API
	req := client.IngestRequest{
		Data:      json.RawMessage(scanData),
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	resp, err := ingestClient.Submit(req)
	if err != nil {
		return fmt.Errorf("failed to submit scan: %w", err)
	}

	// Display response
	return displayIngestResponse(resp, outputFormat)
}

// readScanData reads scan data from a file or stdin
func readScanData(filePath string) ([]byte, error) {
	var reader io.Reader

	if filePath == "-" {
		// Read from stdin
		reader = os.Stdin
	} else {
		// Read from file
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	// Read all data (with a reasonable limit of 100MB)
	data, err := io.ReadAll(io.LimitReader(reader, 100*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data to submit")
	}

	return data, nil
}

// signScanData creates an Ed25519 signature for the scan data
// The signature covers: timestamp + scan_data
func signScanData(scanData []byte, timestamp int64, privKey ed25519.PrivateKey) ([]byte, error) {
	// Construct the message: timestamp + data
	// This binds the timestamp to the data, preventing replay attacks
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)

	// Sign the message
	signature := ed25519.Sign(privKey, message)

	return signature, nil
}

// displayIngestResponse formats and displays the ingest response
func displayIngestResponse(resp *client.IngestResponse, format string) error {
	switch format {
	case "json":
		return displayJSON(resp)
	case "yaml":
		return displayYAML(resp)
	case "table", "":
		return displayTable(resp)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// displayJSON outputs the response as JSON
func displayJSON(resp *client.IngestResponse) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resp)
}

// displayYAML outputs the response as YAML
func displayYAML(resp *client.IngestResponse) error {
	// Convert to YAML-friendly structure
	data := map[string]interface{}{
		"job_id":    resp.JobID,
		"status":    resp.Status,
		"message":   resp.Message,
		"timestamp": resp.Timestamp,
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	for k, v := range data {
		buf.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}

	fmt.Print(buf.String())
	return nil
}

// displayTable outputs the response in a human-readable table format
func displayTable(resp *client.IngestResponse) error {
	fmt.Println()
	fmt.Println("✓ Scan submitted successfully")
	fmt.Println()
	fmt.Printf("  Job ID:    %s\n", resp.JobID)
	fmt.Printf("  Status:    %s\n", resp.Status)
	fmt.Printf("  Message:   %s\n", resp.Message)
	fmt.Printf("  Timestamp: %s\n", resp.Timestamp)
	fmt.Println()
	fmt.Printf("Track job status with: spectra jobs get %s\n", resp.JobID)
	fmt.Println()
	return nil
}
