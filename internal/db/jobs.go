package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spectra-red/recon/internal/models"
	"github.com/surrealdb/surrealdb.go"
	"go.uber.org/zap"
)

// CreateJob creates a new job record in the database with UUID v7 ID
// Returns the created job with all fields populated
func CreateJob(ctx context.Context, db *surrealdb.DB, logger *zap.Logger, scannerKey string) (*models.Job, error) {
	// Generate UUID v7 (time-ordered) for the job ID
	jobID, err := uuid.NewV7()
	if err != nil {
		logger.Error("failed to generate UUID v7",
			zap.Error(err))
		// Fallback to UUID v4 if v7 fails
		jobID = uuid.New()
	}

	now := time.Now().UTC()

	job := &models.Job{
		ID:         jobID.String(),
		ScannerKey: scannerKey,
		State:      models.JobStatePending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Create job record in SurrealDB
	// Using CREATE with explicit ID to ensure idempotency
	query := `CREATE job CONTENT {
		id: $id,
		scanner_key: $scanner_key,
		state: $state,
		created_at: $created_at,
		updated_at: $updated_at,
		completed_at: NONE,
		error_msg: NONE
	}`

	result, err := surrealdb.Query[map[string]interface{}](ctx, db, query, map[string]interface{}{
		"id":          job.ID,
		"scanner_key": job.ScannerKey,
		"state":       job.State.String(),
		"created_at":  job.CreatedAt,
		"updated_at":  job.UpdatedAt,
	})

	if err != nil {
		logger.Error("failed to create job",
			zap.Error(err),
			zap.String("job_id", job.ID))
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Check for query errors
	if result != nil && len(*result) > 0 && (*result)[0].Error != nil {
		logger.Error("query returned error",
			zap.Error((*result)[0].Error),
			zap.String("job_id", job.ID))
		return nil, fmt.Errorf("query error: %w", (*result)[0].Error)
	}

	logger.Info("job created",
		zap.String("job_id", job.ID),
		zap.String("scanner_key", maskPublicKey(scannerKey)),
		zap.String("state", job.State.String()))

	return job, nil
}

// GetJob retrieves a job by its ID
// Returns nil if the job is not found
func GetJob(ctx context.Context, db *surrealdb.DB, logger *zap.Logger, jobID string) (*models.Job, error) {
	query := `SELECT * FROM job WHERE id = $id LIMIT 1`

	result, err := surrealdb.Query[map[string]interface{}](ctx, db, query, map[string]interface{}{
		"id": jobID,
	})

	if err != nil {
		logger.Error("failed to query job",
			zap.Error(err),
			zap.String("job_id", jobID))
		return nil, fmt.Errorf("failed to query job: %w", err)
	}

	// Check if job was found
	if result == nil || len(*result) == 0 {
		logger.Debug("job not found",
			zap.String("job_id", jobID))
		return nil, nil
	}

	// Get the first query result
	queryResult := (*result)[0]
	if queryResult.Error != nil {
		logger.Error("query returned error",
			zap.Error(queryResult.Error),
			zap.String("job_id", jobID))
		return nil, fmt.Errorf("query error: %w", queryResult.Error)
	}

	if queryResult.Result == nil {
		logger.Debug("job not found",
			zap.String("job_id", jobID))
		return nil, nil
	}

	// Parse the result into a Job struct
	job, err := parseJobResult(queryResult.Result)
	if err != nil {
		logger.Error("failed to parse job result",
			zap.Error(err),
			zap.String("job_id", jobID))
		return nil, fmt.Errorf("failed to parse job: %w", err)
	}

	return job, nil
}

// UpdateJobState updates the state of a job atomically
// This enforces the state machine transitions defined in models.Job
func UpdateJobState(ctx context.Context, db *surrealdb.DB, logger *zap.Logger, jobID string, newState models.JobState, errorMsg *string) error {
	// First, get the current job to validate the transition
	job, err := GetJob(ctx, db, logger, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job for state update: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}

	// Validate the state transition
	if !job.CanTransition(newState) {
		logger.Warn("invalid state transition attempted",
			zap.String("job_id", jobID),
			zap.String("current_state", job.State.String()),
			zap.String("new_state", newState.String()))
		return fmt.Errorf("invalid state transition from %s to %s", job.State, newState)
	}

	now := time.Now().UTC()

	// Build the update query
	query := `UPDATE job SET
		state = $state,
		updated_at = $updated_at`

	params := map[string]interface{}{
		"state":      newState.String(),
		"updated_at": now,
	}

	// Add completed_at for terminal states
	if newState == models.JobStateCompleted || newState == models.JobStateFailed {
		query += `, completed_at = $completed_at`
		params["completed_at"] = now
	}

	// Add error message if provided
	if errorMsg != nil {
		query += `, error_msg = $error_msg`
		params["error_msg"] = *errorMsg
	}

	query += ` WHERE id = $id`
	params["id"] = jobID

	result, err := surrealdb.Query[map[string]interface{}](ctx, db, query, params)
	if err != nil {
		logger.Error("failed to update job state",
			zap.Error(err),
			zap.String("job_id", jobID),
			zap.String("new_state", newState.String()))
		return fmt.Errorf("failed to update job state: %w", err)
	}

	// Check for query errors
	if result != nil && len(*result) > 0 && (*result)[0].Error != nil {
		logger.Error("query returned error",
			zap.Error((*result)[0].Error),
			zap.String("job_id", jobID))
		return fmt.Errorf("query error: %w", (*result)[0].Error)
	}

	logger.Info("job state updated",
		zap.String("job_id", jobID),
		zap.String("old_state", job.State.String()),
		zap.String("new_state", newState.String()))

	return nil
}

// ListJobs retrieves a paginated list of jobs based on filters
func ListJobs(ctx context.Context, db *surrealdb.DB, logger *zap.Logger, req models.JobListRequest) (*models.JobListResponse, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid list request: %w", err)
	}

	// Build the query with filters
	query := `SELECT * FROM job`
	params := make(map[string]interface{})

	var whereClauses []string

	// Add scanner_key filter if provided
	if req.ScannerKey != nil {
		whereClauses = append(whereClauses, "scanner_key = $scanner_key")
		params["scanner_key"] = *req.ScannerKey
	}

	// Add state filter if provided
	if req.State != nil {
		whereClauses = append(whereClauses, "state = $state")
		params["state"] = req.State.String()
	}

	// Add WHERE clause if there are filters
	if len(whereClauses) > 0 {
		query += ` WHERE `
		for i, clause := range whereClauses {
			if i > 0 {
				query += ` AND `
			}
			query += clause
		}
	}

	// Add ORDER BY
	orderDir := "DESC"
	if !req.OrderDesc {
		orderDir = "ASC"
	}
	query += fmt.Sprintf(` ORDER BY %s %s`, req.OrderBy, orderDir)

	// Add LIMIT and OFFSET
	query += ` LIMIT $limit START $offset`
	params["limit"] = req.Limit
	params["offset"] = req.Offset

	logger.Debug("listing jobs",
		zap.String("query", query),
		zap.Any("params", params))

	// Execute query
	results, err := surrealdb.Query[[]map[string]interface{}](ctx, db, query, params)
	if err != nil {
		logger.Error("failed to list jobs",
			zap.Error(err))
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Parse results
	jobs := make([]models.Job, 0)
	if results != nil && len(*results) > 0 {
		queryResult := (*results)[0]
		if queryResult.Error != nil {
			logger.Error("query returned error",
				zap.Error(queryResult.Error))
			return nil, fmt.Errorf("query error: %w", queryResult.Error)
		}

		if queryResult.Result != nil {
			for _, jobData := range queryResult.Result {
				job, err := parseJobResult(jobData)
				if err != nil {
					logger.Warn("failed to parse job in list",
						zap.Error(err))
					continue
				}
				jobs = append(jobs, *job)
			}
		}
	}

	// Get total count for pagination
	countQuery := `SELECT count() FROM job`
	if len(whereClauses) > 0 {
		countQuery += ` WHERE `
		for i, clause := range whereClauses {
			if i > 0 {
				countQuery += ` AND `
			}
			countQuery += clause
		}
	}
	countQuery += ` GROUP ALL`

	// Note: Count query simplified - in production you'd want proper count handling
	total := len(jobs) // Simplified for now

	// Build response
	response := &models.JobListResponse{
		Jobs:       jobs,
		Total:      total,
		Limit:      req.Limit,
		Offset:     req.Offset,
		HasMore:    len(jobs) == req.Limit,
		NextOffset: req.Offset + len(jobs),
	}

	logger.Debug("jobs listed",
		zap.Int("count", len(jobs)),
		zap.Int("total", total))

	return response, nil
}

// parseJobResult parses a SurrealDB result into a Job struct
func parseJobResult(data map[string]interface{}) (*models.Job, error) {
	job := &models.Job{}

	// Required fields
	if id, ok := data["id"].(string); ok {
		job.ID = id
	} else {
		return nil, fmt.Errorf("missing or invalid id field")
	}

	if scannerKey, ok := data["scanner_key"].(string); ok {
		job.ScannerKey = scannerKey
	} else {
		return nil, fmt.Errorf("missing or invalid scanner_key field")
	}

	if state, ok := data["state"].(string); ok {
		job.State = models.JobState(state)
	} else {
		return nil, fmt.Errorf("missing or invalid state field")
	}

	// Parse timestamps
	if createdAt, err := parseTimeField(data, "created_at"); err == nil {
		job.CreatedAt = createdAt
	}
	if updatedAt, err := parseTimeField(data, "updated_at"); err == nil {
		job.UpdatedAt = updatedAt
	}

	// Optional fields
	if completedAt, err := parseTimeField(data, "completed_at"); err == nil {
		job.CompletedAt = &completedAt
	}

	if errorMsg, ok := data["error_msg"].(string); ok && errorMsg != "" {
		job.ErrorMessage = &errorMsg
	}

	// Parse optional host/port counts
	if hostCount, ok := getIntField(data, "host_count"); ok {
		job.HostCount = hostCount
	}
	if portCount, ok := getIntField(data, "port_count"); ok {
		job.PortCount = portCount
	}

	return job, nil
}

// maskPublicKey masks a public key for logging (shows first 8 chars only)
func maskPublicKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:8] + "..."
}
