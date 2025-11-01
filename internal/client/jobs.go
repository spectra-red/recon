package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spectra-red/recon/internal/models"
)

// GetJob retrieves a job by its ID
func (c *Client) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v1/jobs/"+jobID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var job models.Job
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}

	return &job, nil
}

// ListJobsOptions contains options for listing jobs
type ListJobsOptions struct {
	ScannerKey *string
	State      *models.JobState
	Limit      int
	Offset     int
	OrderBy    string
	OrderDesc  bool
}

// ListJobs retrieves a paginated list of jobs
func (c *Client) ListJobs(ctx context.Context, opts ListJobsOptions) (*models.JobListResponse, error) {
	// Build query parameters
	params := url.Values{}

	if opts.ScannerKey != nil {
		params.Set("scanner_key", *opts.ScannerKey)
	}

	if opts.State != nil {
		params.Set("state", opts.State.String())
	}

	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}

	if opts.Offset > 0 {
		params.Set("offset", strconv.Itoa(opts.Offset))
	}

	if opts.OrderBy != "" {
		params.Set("order_by", opts.OrderBy)
	}

	if !opts.OrderDesc {
		params.Set("order_desc", "false")
	}

	path := "/v1/jobs"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var listResp models.JobListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse list response: %w", err)
	}

	return &listResp, nil
}
