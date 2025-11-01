package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJob(t *testing.T) {
	tests := []struct {
		name           string
		jobID          string
		serverResponse models.Job
		serverStatus   int
		wantErr        bool
		errContains    string
	}{
		{
			name:  "successful get",
			jobID: "job-123",
			serverResponse: models.Job{
				ID:         "job-123",
				State:      models.JobStateCompleted,
				ScannerKey: "scanner-key-abc",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				HostCount:  10,
				PortCount:  50,
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "job not found",
			jobID:        "nonexistent",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
			errContains:  "job not found",
		},
		{
			name:         "server error",
			jobID:        "job-123",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
			errContains:  "API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/jobs/"+tt.jobID, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)

				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				} else if tt.serverStatus == http.StatusNotFound {
					json.NewEncoder(w).Encode(ErrorResponse{
						Error:   "not_found",
						Message: "Job not found",
					})
				} else {
					json.NewEncoder(w).Encode(ErrorResponse{
						Error:   "internal_error",
						Message: "Internal server error",
					})
				}
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL)

			// Execute
			ctx := context.Background()
			job, err := client.GetJob(ctx, tt.jobID)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, job)
			} else {
				require.NoError(t, err)
				require.NotNil(t, job)
				assert.Equal(t, tt.serverResponse.ID, job.ID)
				assert.Equal(t, tt.serverResponse.State, job.State)
				assert.Equal(t, tt.serverResponse.ScannerKey, job.ScannerKey)
			}
		})
	}
}

func TestListJobs(t *testing.T) {
	tests := []struct {
		name           string
		opts           ListJobsOptions
		serverResponse models.JobListResponse
		wantErr        bool
		wantQueryCheck func(t *testing.T, url string)
	}{
		{
			name: "list all jobs",
			opts: ListJobsOptions{
				Limit:  50,
				Offset: 0,
			},
			serverResponse: models.JobListResponse{
				Jobs: []models.Job{
					{
						ID:         "job-1",
						State:      models.JobStateCompleted,
						ScannerKey: "key-1",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					},
					{
						ID:         "job-2",
						State:      models.JobStateProcessing,
						ScannerKey: "key-2",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					},
				},
				Total:      2,
				Limit:      50,
				Offset:     0,
				HasMore:    false,
				NextOffset: 2,
			},
			wantErr: false,
		},
		{
			name: "filter by state",
			opts: ListJobsOptions{
				State:  ptrJobState(models.JobStateCompleted),
				Limit:  10,
				Offset: 0,
			},
			serverResponse: models.JobListResponse{
				Jobs: []models.Job{
					{
						ID:         "job-1",
						State:      models.JobStateCompleted,
						ScannerKey: "key-1",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					},
				},
				Total:      1,
				Limit:      10,
				Offset:     0,
				HasMore:    false,
				NextOffset: 1,
			},
			wantErr: false,
			wantQueryCheck: func(t *testing.T, url string) {
				assert.Contains(t, url, "state=completed")
			},
		},
		{
			name: "filter by scanner key",
			opts: ListJobsOptions{
				ScannerKey: ptrString("scanner-abc"),
				Limit:      20,
				Offset:     0,
			},
			serverResponse: models.JobListResponse{
				Jobs:       []models.Job{},
				Total:      0,
				Limit:      20,
				Offset:     0,
				HasMore:    false,
				NextOffset: 0,
			},
			wantErr: false,
			wantQueryCheck: func(t *testing.T, url string) {
				assert.Contains(t, url, "scanner_key=scanner-abc")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/jobs", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				// Check query parameters if needed
				if tt.wantQueryCheck != nil {
					tt.wantQueryCheck(t, r.URL.RawQuery)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL)

			// Execute
			ctx := context.Background()
			resp, err := client.ListJobs(ctx, tt.opts)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, len(tt.serverResponse.Jobs), len(resp.Jobs))
				assert.Equal(t, tt.serverResponse.Total, resp.Total)
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with short timeout
	client := NewClient(server.URL).WithTimeout(100 * time.Millisecond)

	// Execute
	ctx := context.Background()
	_, err := client.GetJob(ctx, "job-123")

	// Should timeout
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// Helper functions

func ptrString(s string) *string {
	return &s
}

func ptrJobState(s models.JobState) *models.JobState {
	return &s
}
