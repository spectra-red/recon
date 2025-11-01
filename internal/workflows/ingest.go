package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	restate "github.com/restatedev/sdk-go"
	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
)

// IngestWorkflow handles the durable scan ingestion workflow
type IngestWorkflow struct {
	db *surrealdb.DB
}

// NewIngestWorkflow creates a new IngestWorkflow instance
func NewIngestWorkflow(db *surrealdb.DB) *IngestWorkflow {
	return &IngestWorkflow{
		db: db,
	}
}

// ServiceName returns the Restate service name
func (w *IngestWorkflow) ServiceName() string {
	return "IngestWorkflow"
}

// PersistResult holds the result of persisting scan data
type PersistResult struct {
	Hosts int
	Ports int
}

// Run executes the ingest workflow with durable steps
// This workflow is idempotent and can be safely retried
func (w *IngestWorkflow) Run(ctx restate.Context, req models.IngestWorkflowRequest) (models.IngestWorkflowResponse, error) {
	// Step 1: Update job state to "processing"
	_, err := restate.Run[string](ctx, func(ctx restate.RunContext) (string, error) {
		return "", w.updateJobState(req.JobID, models.JobStateProcessing, "", req.ScannerKey)
	})
	if err != nil {
		// If we can't update to processing, fail the job immediately
		_ = w.updateJobState(req.JobID, models.JobStateFailed, fmt.Sprintf("Failed to update job to processing: %v", err), req.ScannerKey)
		return models.IngestWorkflowResponse{
			JobID: req.JobID,
			State: models.JobStateFailed,
		}, fmt.Errorf("failed to update job to processing: %w", err)
	}

	// Step 2: Parse and validate scan data
	scanData, err := restate.Run[*models.ScanData](ctx, func(ctx restate.RunContext) (*models.ScanData, error) {
		return w.parseScanData(req.ScanData)
	})
	if err != nil {
		_ = w.updateJobState(req.JobID, models.JobStateFailed, fmt.Sprintf("Failed to parse scan data: %v", err), req.ScannerKey)
		return models.IngestWorkflowResponse{
			JobID: req.JobID,
			State: models.JobStateFailed,
		}, fmt.Errorf("failed to parse scan data: %w", err)
	}

	// Step 3: Persist scan results to SurrealDB
	persistResult, err := restate.Run[PersistResult](ctx, func(ctx restate.RunContext) (PersistResult, error) {
		hosts, ports, err := w.persistScanData(req.JobID, scanData, req.ScannerKey)
		return PersistResult{Hosts: hosts, Ports: ports}, err
	})
	if err != nil {
		_ = w.updateJobState(req.JobID, models.JobStateFailed, fmt.Sprintf("Failed to persist scan data: %v", err), req.ScannerKey)
		return models.IngestWorkflowResponse{
			JobID: req.JobID,
			State: models.JobStateFailed,
		}, fmt.Errorf("failed to persist scan data: %w", err)
	}

	// Step 4: Update job state to "completed"
	_, err = restate.Run[string](ctx, func(ctx restate.RunContext) (string, error) {
		return "", w.updateJobStateWithCounts(req.JobID, models.JobStateCompleted, "", req.ScannerKey, persistResult.Hosts, persistResult.Ports)
	})
	if err != nil {
		// Even if we fail to update to completed, the data is persisted
		// This is a non-critical error, so we log it but don't fail the workflow
		return models.IngestWorkflowResponse{
			JobID:     req.JobID,
			State:     models.JobStateCompleted, // Data was persisted successfully
			HostCount: persistResult.Hosts,
			PortCount: persistResult.Ports,
		}, nil
	}

	return models.IngestWorkflowResponse{
		JobID:     req.JobID,
		State:     models.JobStateCompleted,
		HostCount: persistResult.Hosts,
		PortCount: persistResult.Ports,
	}, nil
}

// updateJobState updates the job state in SurrealDB
func (w *IngestWorkflow) updateJobState(jobID string, state models.JobState, errorMsg string, scannerKey string) error {
	ctx := context.Background()
	now := time.Now().UTC()

	var errorPtr *string
	if errorMsg != "" {
		errorPtr = &errorMsg
	}

	// Try to get existing job first
	query := `SELECT * FROM type::thing('job', $job_id) LIMIT 1;`
	existingJobsResult, err := surrealdb.Query[[]models.Job](ctx, w.db, query, map[string]interface{}{
		"job_id": jobID,
	})

	// Extract jobs from result
	var existingJobs []models.Job
	if existingJobsResult != nil && len(*existingJobsResult) > 0 {
		existingJobs = (*existingJobsResult)[0].Result
	}

	// If job doesn't exist, create it
	if err != nil || len(existingJobs) == 0 {
		createQuery := `
			CREATE type::thing('job', $job_id) CONTENT {
				id: $job_id,
				state: $state,
				scanner_key: $scanner_key,
				error_message: $error_message,
				created_at: $now,
				updated_at: $now,
				host_count: 0,
				port_count: 0
			};
		`
		_, err = surrealdb.Query[interface{}](ctx, w.db, createQuery, map[string]interface{}{
			"job_id":        jobID,
			"state":         string(state),
			"scanner_key":   scannerKey,
			"error_message": errorPtr,
			"now":           now,
		})
		return err
	}

	// Update existing job
	updateData := map[string]interface{}{
		"state":      string(state),
		"updated_at": now,
	}
	if errorMsg != "" {
		updateData["error_message"] = errorPtr
	}
	if state == models.JobStateCompleted || state == models.JobStateFailed {
		updateData["completed_at"] = now
	}

	updateQuery := `UPDATE type::thing('job', $job_id) MERGE $data;`
	_, err = surrealdb.Query[interface{}](ctx, w.db, updateQuery, map[string]interface{}{
		"job_id": jobID,
		"data":   updateData,
	})

	return err
}

// updateJobStateWithCounts updates the job state with host and port counts
func (w *IngestWorkflow) updateJobStateWithCounts(jobID string, state models.JobState, errorMsg string, scannerKey string, hostCount, portCount int) error {
	ctx := context.Background()
	now := time.Now().UTC()

	var errorPtr *string
	if errorMsg != "" {
		errorPtr = &errorMsg
	}

	updateData := map[string]interface{}{
		"state":      string(state),
		"updated_at": now,
		"host_count": hostCount,
		"port_count": portCount,
	}
	if errorMsg != "" {
		updateData["error_message"] = errorPtr
	}
	if state == models.JobStateCompleted || state == models.JobStateFailed {
		updateData["completed_at"] = now
	}

	updateQuery := `UPDATE type::thing('job', $job_id) MERGE $data;`
	_, err := surrealdb.Query[interface{}](ctx, w.db, updateQuery, map[string]interface{}{
		"job_id": jobID,
		"data":   updateData,
	})

	return err
}

// parseScanData parses and validates scan data from Naabu JSON format
func (w *IngestWorkflow) parseScanData(rawData []byte) (*models.ScanData, error) {
	// Naabu outputs JSON lines format (one JSON object per line)
	// Example:
	// {"host":"1.2.3.4","port":80,"protocol":"tcp"}
	// {"host":"1.2.3.4","port":443,"protocol":"tcp"}

	lines := strings.Split(string(rawData), "\n")
	hostMap := make(map[string]*models.ScanHost)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var naabuEntry struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Protocol string `json:"protocol"`
		}

		if err := json.Unmarshal([]byte(line), &naabuEntry); err != nil {
			// Skip malformed lines but don't fail the entire parse
			continue
		}

		// Validate required fields
		if naabuEntry.Host == "" || naabuEntry.Port == 0 {
			continue
		}

		// Default protocol to tcp if not specified
		if naabuEntry.Protocol == "" {
			naabuEntry.Protocol = "tcp"
		}

		// Add to host map (group ports by host)
		host, exists := hostMap[naabuEntry.Host]
		if !exists {
			host = &models.ScanHost{
				IP:    naabuEntry.Host,
				Ports: []models.ScanPort{},
			}
			hostMap[naabuEntry.Host] = host
		}

		host.Ports = append(host.Ports, models.ScanPort{
			Number:   naabuEntry.Port,
			Protocol: naabuEntry.Protocol,
			State:    "open", // Naabu only reports open ports
		})
	}

	// Convert map to slice
	hosts := make([]models.ScanHost, 0, len(hostMap))
	for _, host := range hostMap {
		hosts = append(hosts, *host)
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("no valid hosts found in scan data")
	}

	return &models.ScanData{
		Hosts: hosts,
	}, nil
}

// persistScanData persists scan data to SurrealDB
// Returns (hostCount, portCount, error)
func (w *IngestWorkflow) persistScanData(jobID string, scanData *models.ScanData, scannerKey string) (int, int, error) {
	ctx := context.Background()
	hostCount := 0
	portCount := 0
	now := time.Now().UTC()

	for _, host := range scanData.Hosts {
		// Upsert host node
		upsertHostQuery := `
			LET $host_id = type::thing('host', $ip_encoded);
			CREATE $host_id CONTENT {
				ip: $ip,
				last_seen: $now,
				last_scanned_at: $now,
				first_seen: $now
			} ON DUPLICATE KEY UPDATE {
				last_seen: $now,
				last_scanned_at: $now
			};
		`
		_, err := surrealdb.Query[interface{}](ctx, w.db, upsertHostQuery, map[string]interface{}{
			"ip_encoded": strings.ReplaceAll(host.IP, ".", "_"),
			"ip":         host.IP,
			"now":        now,
		})

		if err != nil {
			return hostCount, portCount, fmt.Errorf("failed to upsert host %s: %w", host.IP, err)
		}
		hostCount++

		// Upsert ports and create HAS edges
		for _, port := range host.Ports {
			portID := fmt.Sprintf("port_%d_%s", port.Number, port.Protocol)

			// Upsert port
			upsertPortQuery := `
				LET $port_id = type::thing('port', $port_encoded);
				CREATE $port_id CONTENT {
					number: $number,
					protocol: $protocol,
					last_seen: $now,
					first_seen: $now
				} ON DUPLICATE KEY UPDATE {
					last_seen: $now
				};
			`
			_, err := surrealdb.Query[interface{}](ctx, w.db, upsertPortQuery, map[string]interface{}{
				"port_encoded": portID,
				"number":       port.Number,
				"protocol":     port.Protocol,
				"now":          now,
			})

			if err != nil {
				return hostCount, portCount, fmt.Errorf("failed to upsert port %d: %w", port.Number, err)
			}

			// Create HAS edge (host -> port)
			relateQuery := `
				LET $host_id = type::thing('host', $host_encoded);
				LET $port_id = type::thing('port', $port_encoded);
				RELATE $host_id->HAS->$port_id CONTENT {
					first_seen: $now,
					last_seen: $now
				} ON DUPLICATE KEY UPDATE {
					last_seen: $now
				};
			`
			_, err = surrealdb.Query[interface{}](ctx, w.db, relateQuery, map[string]interface{}{
				"host_encoded": strings.ReplaceAll(host.IP, ".", "_"),
				"port_encoded": portID,
				"now":          now,
			})

			if err != nil {
				return hostCount, portCount, fmt.Errorf("failed to create HAS edge: %w", err)
			}

			portCount++
		}
	}

	return hostCount, portCount, nil
}
