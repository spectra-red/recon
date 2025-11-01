package models

import (
	"fmt"
	"time"
)

// JobState represents the state of a scan ingestion job
type JobState string

const (
	JobStatePending    JobState = "pending"
	JobStateProcessing JobState = "processing"
	JobStateCompleted  JobState = "completed"
	JobStateFailed     JobState = "failed"
)

// IsValid checks if the job state is one of the allowed values
func (s JobState) IsValid() bool {
	switch s {
	case JobStatePending, JobStateProcessing, JobStateCompleted, JobStateFailed:
		return true
	default:
		return false
	}
}

// String returns the string representation of the JobState
func (s JobState) String() string {
	return string(s)
}

// Job represents a scan ingestion job in the workflow system
type Job struct {
	ID           string     `json:"id"`
	State        JobState   `json:"state"`
	ScannerKey   string     `json:"scanner_key"`           // Public key of the contributor
	ErrorMessage *string    `json:"error_message,omitempty"` // Error message if state is failed
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	HostCount    int        `json:"host_count"`    // Number of hosts processed
	PortCount    int        `json:"port_count"`    // Number of ports processed
}

// JobStateTransition defines allowed state transitions
type JobStateTransition struct {
	From JobState
	To   JobState
}

// AllowedTransitions defines the valid state machine transitions for jobs
var AllowedTransitions = map[JobStateTransition]bool{
	// From pending
	{JobStatePending, JobStateProcessing}: true,
	{JobStatePending, JobStateFailed}:     true,

	// From processing
	{JobStateProcessing, JobStateCompleted}: true,
	{JobStateProcessing, JobStateFailed}:    true,

	// Terminal states (completed/failed) cannot transition further
	// This is enforced by the absence of transitions from these states
}

// CanTransition checks if a state transition is allowed
func (j *Job) CanTransition(newState JobState) bool {
	// Check if the new state is valid
	if !newState.IsValid() {
		return false
	}

	// If already in the target state, that's a no-op (allowed)
	if j.State == newState {
		return true
	}

	// Check if this transition is allowed
	transition := JobStateTransition{From: j.State, To: newState}
	return AllowedTransitions[transition]
}

// TransitionTo attempts to transition the job to a new state
// Returns an error if the transition is not allowed
func (j *Job) TransitionTo(newState JobState) error {
	if !j.CanTransition(newState) {
		return fmt.Errorf("invalid state transition from %s to %s", j.State, newState)
	}

	j.State = newState
	j.UpdatedAt = time.Now().UTC()

	// Set completed_at for terminal states
	if newState == JobStateCompleted || newState == JobStateFailed {
		now := time.Now().UTC()
		j.CompletedAt = &now
	}

	return nil
}

// SetError sets the error message and transitions to failed state
func (j *Job) SetError(errMsg string) error {
	j.ErrorMessage = &errMsg
	return j.TransitionTo(JobStateFailed)
}

// IngestWorkflowRequest represents the request to the ingest workflow
type IngestWorkflowRequest struct {
	JobID      string `json:"job_id"`
	ScannerKey string `json:"scanner_key"`
	ScanData   []byte `json:"scan_data"` // Raw JSON scan data
}

// IngestWorkflowResponse represents the response from the ingest workflow
type IngestWorkflowResponse struct {
	JobID     string `json:"job_id"`
	State     JobState `json:"state"`
	HostCount int    `json:"host_count"`
	PortCount int    `json:"port_count"`
}

// ScanData represents the parsed scan data structure (Naabu format)
type ScanData struct {
	Hosts []ScanHost `json:"hosts"`
}

// ScanHost represents a scanned host with its ports
type ScanHost struct {
	IP    string     `json:"ip"`
	Ports []ScanPort `json:"ports"`
}

// ScanPort represents a scanned port
type ScanPort struct {
	Number   int    `json:"number"`
	Protocol string `json:"protocol"` // tcp, udp
	State    string `json:"state"`    // open, closed, filtered
}

// JobListRequest represents the parameters for listing jobs
type JobListRequest struct {
	ScannerKey *string   // Filter by scanner_key (optional)
	State      *JobState // Filter by state (optional)
	Limit      int       // Maximum number of results (default: 50, max: 500)
	Offset     int       // Offset for pagination (default: 0)
	OrderBy    string    // Field to order by (default: "created_at", options: "created_at", "updated_at")
	OrderDesc  bool      // Order descending (default: true)
}

// Validate validates the JobListRequest parameters
func (r *JobListRequest) Validate() error {
	// Validate limit
	if r.Limit < 1 {
		r.Limit = 50 // default
	}
	if r.Limit > 500 {
		return fmt.Errorf("limit cannot exceed 500 (got %d)", r.Limit)
	}

	// Validate offset
	if r.Offset < 0 {
		return fmt.Errorf("offset cannot be negative (got %d)", r.Offset)
	}

	// Validate state if provided
	if r.State != nil && !r.State.IsValid() {
		return fmt.Errorf("invalid state: %s", *r.State)
	}

	// Validate order_by
	if r.OrderBy == "" {
		r.OrderBy = "created_at" // default
	}
	validOrderFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
	}
	if !validOrderFields[r.OrderBy] {
		return fmt.Errorf("invalid order_by field: %s (must be 'created_at' or 'updated_at')", r.OrderBy)
	}

	return nil
}

// JobListResponse represents the response for listing jobs
type JobListResponse struct {
	Jobs       []Job `json:"jobs"`        // List of jobs
	Total      int   `json:"total"`       // Total number of jobs matching the filter
	Limit      int   `json:"limit"`       // Limit applied
	Offset     int   `json:"offset"`      // Offset applied
	HasMore    bool  `json:"has_more"`    // Whether there are more results
	NextOffset int   `json:"next_offset"` // Offset for the next page
}
