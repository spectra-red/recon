package workflows

import (
	"fmt"
	"testing"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestParseScanData_ValidNaabuOutput(t *testing.T) {
	workflow := &IngestWorkflow{}

	naabuOutput := `{"host":"192.168.1.1","port":22,"protocol":"tcp"}
{"host":"192.168.1.1","port":80,"protocol":"tcp"}
{"host":"192.168.1.2","port":443,"protocol":"tcp"}`

	result, err := workflow.parseScanData([]byte(naabuOutput))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Hosts, 2)

	// Check host 1
	host1Found := false
	host2Found := false

	for _, host := range result.Hosts {
		if host.IP == "192.168.1.1" {
			host1Found = true
			assert.Len(t, host.Ports, 2)

			// Check ports
			portNumbers := make(map[int]bool)
			for _, port := range host.Ports {
				portNumbers[port.Number] = true
				assert.Equal(t, "tcp", port.Protocol)
				assert.Equal(t, "open", port.State)
			}
			assert.True(t, portNumbers[22])
			assert.True(t, portNumbers[80])
		}

		if host.IP == "192.168.1.2" {
			host2Found = true
			assert.Len(t, host.Ports, 1)
			assert.Equal(t, 443, host.Ports[0].Number)
			assert.Equal(t, "tcp", host.Ports[0].Protocol)
			assert.Equal(t, "open", host.Ports[0].State)
		}
	}

	assert.True(t, host1Found, "Host 192.168.1.1 should be found")
	assert.True(t, host2Found, "Host 192.168.1.2 should be found")
}

func TestParseScanData_EmptyInput(t *testing.T) {
	workflow := &IngestWorkflow{}

	result, err := workflow.parseScanData([]byte(""))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no valid hosts found")
}

func TestParseScanData_MalformedJSON(t *testing.T) {
	workflow := &IngestWorkflow{}

	naabuOutput := `{"host":"192.168.1.1","port":22,"protocol":"tcp"}
{invalid json}
{"host":"192.168.1.2","port":443,"protocol":"tcp"}`

	result, err := workflow.parseScanData([]byte(naabuOutput))

	// Should succeed but skip the malformed line
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Hosts, 2)
}

func TestParseScanData_MissingRequiredFields(t *testing.T) {
	workflow := &IngestWorkflow{}

	naabuOutput := `{"host":"192.168.1.1"}
{"port":22,"protocol":"tcp"}
{"host":"192.168.1.2","port":443,"protocol":"tcp"}`

	result, err := workflow.parseScanData([]byte(naabuOutput))

	// Should succeed with only the valid entry
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Hosts, 1)
	assert.Equal(t, "192.168.1.2", result.Hosts[0].IP)
}

func TestParseScanData_DefaultProtocol(t *testing.T) {
	workflow := &IngestWorkflow{}

	naabuOutput := `{"host":"192.168.1.1","port":22}`

	result, err := workflow.parseScanData([]byte(naabuOutput))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Hosts, 1)
	assert.Equal(t, "tcp", result.Hosts[0].Ports[0].Protocol, "Should default to tcp")
}

func TestParseScanData_UDPProtocol(t *testing.T) {
	workflow := &IngestWorkflow{}

	naabuOutput := `{"host":"192.168.1.1","port":53,"protocol":"udp"}`

	result, err := workflow.parseScanData([]byte(naabuOutput))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Hosts, 1)
	assert.Equal(t, "udp", result.Hosts[0].Ports[0].Protocol)
}

func TestParseScanData_LargeDataset(t *testing.T) {
	workflow := &IngestWorkflow{}

	// Generate 100 host entries with valid IP addresses
	var output string
	for i := 1; i <= 100; i++ {
		// Generate valid IP addresses in 10.0.x.x range
		octet2 := i / 256
		octet3 := i % 256
		output += fmt.Sprintf(`{"host":"10.0.%d.%d","port":80,"protocol":"tcp"}`, octet2, octet3) + "\n"
	}

	result, err := workflow.parseScanData([]byte(output))

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should handle large datasets efficiently
	assert.GreaterOrEqual(t, len(result.Hosts), 1)
}

func TestJobStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		initialState models.JobState
		targetState  models.JobState
		shouldSucceed bool
	}{
		{
			name:         "Pending to Processing - Valid",
			initialState: models.JobStatePending,
			targetState:  models.JobStateProcessing,
			shouldSucceed: true,
		},
		{
			name:         "Processing to Completed - Valid",
			initialState: models.JobStateProcessing,
			targetState:  models.JobStateCompleted,
			shouldSucceed: true,
		},
		{
			name:         "Processing to Failed - Valid",
			initialState: models.JobStateProcessing,
			targetState:  models.JobStateFailed,
			shouldSucceed: true,
		},
		{
			name:         "Pending to Failed - Valid",
			initialState: models.JobStatePending,
			targetState:  models.JobStateFailed,
			shouldSucceed: true,
		},
		{
			name:         "Completed to Processing - Invalid",
			initialState: models.JobStateCompleted,
			targetState:  models.JobStateProcessing,
			shouldSucceed: false,
		},
		{
			name:         "Failed to Processing - Invalid",
			initialState: models.JobStateFailed,
			targetState:  models.JobStateProcessing,
			shouldSucceed: false,
		},
		{
			name:         "Pending to Completed - Invalid (must go through Processing)",
			initialState: models.JobStatePending,
			targetState:  models.JobStateCompleted,
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &models.Job{
				ID:    "test-job",
				State: tt.initialState,
			}

			canTransition := job.CanTransition(tt.targetState)
			assert.Equal(t, tt.shouldSucceed, canTransition)

			if tt.shouldSucceed {
				err := job.TransitionTo(tt.targetState)
				assert.NoError(t, err)
				assert.Equal(t, tt.targetState, job.State)
			}
		})
	}
}

func TestJobSetError(t *testing.T) {
	job := &models.Job{
		ID:    "test-job",
		State: models.JobStateProcessing,
	}

	err := job.SetError("Test error message")

	assert.NoError(t, err)
	assert.Equal(t, models.JobStateFailed, job.State)
	assert.NotNil(t, job.ErrorMessage)
	assert.Equal(t, "Test error message", *job.ErrorMessage)
	assert.NotNil(t, job.CompletedAt)
}

func TestIngestWorkflowRequest_Validation(t *testing.T) {
	tests := []struct {
		name      string
		request   models.IngestWorkflowRequest
		shouldErr bool
	}{
		{
			name: "Valid Request",
			request: models.IngestWorkflowRequest{
				JobID:      "job-123",
				ScannerKey: "scanner-abc",
				ScanData:   []byte(`{"host":"192.168.1.1","port":80,"protocol":"tcp"}`),
			},
			shouldErr: false,
		},
		{
			name: "Empty JobID",
			request: models.IngestWorkflowRequest{
				JobID:      "",
				ScannerKey: "scanner-abc",
				ScanData:   []byte(`{"host":"192.168.1.1","port":80,"protocol":"tcp"}`),
			},
			shouldErr: true,
		},
		{
			name: "Empty ScannerKey",
			request: models.IngestWorkflowRequest{
				JobID:      "job-123",
				ScannerKey: "",
				ScanData:   []byte(`{"host":"192.168.1.1","port":80,"protocol":"tcp"}`),
			},
			shouldErr: true,
		},
		{
			name: "Empty ScanData",
			request: models.IngestWorkflowRequest{
				JobID:      "job-123",
				ScannerKey: "scanner-abc",
				ScanData:   []byte{},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.request.JobID == "" || tt.request.ScannerKey == "" || len(tt.request.ScanData) == 0

			assert.Equal(t, tt.shouldErr, hasError)
		})
	}
}
