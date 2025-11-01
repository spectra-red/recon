package cli

import (
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestColorizeJobState(t *testing.T) {
	tests := []struct {
		name  string
		state models.JobState
	}{
		{"completed", models.JobStateCompleted},
		{"failed", models.JobStateFailed},
		{"processing", models.JobStateProcessing},
		{"pending", models.JobStatePending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorizeJobState(tt.state)
			// Should not be empty
			assert.NotEmpty(t, result)
			// Should contain the state string (with or without color codes)
			assert.Contains(t, result, tt.state.String())
		})
	}
}

func TestMaskScannerKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "long key",
			key:      "abcdefghijklmnopqrstuvwxyz",
			expected: "abcdefgh...wxyz",
		},
		{
			name:     "short key",
			key:      "abc123",
			expected: "abc123",
		},
		{
			name:     "exactly 12 chars",
			key:      "abcdefghijkl",
			expected: "abcdefghijkl",
		},
		{
			name:     "13 chars - should mask",
			key:      "abcdefghijklm",
			expected: "abcdefgh...jklm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskScannerKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"zero", 0, "-"},
		{"positive", 42, "42"},
		{"large", 10000, "10000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCount(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "-"},
		{"milliseconds", 500 * time.Millisecond, "500ms"},
		{"seconds", 45 * time.Second, "45s"},
		{"minutes", 5 * time.Minute, "5m 0s"},
		{"minutes and seconds", 5*time.Minute + 30*time.Second, "5m 30s"},
		{"hours", 2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJobsCommand(t *testing.T) {
	cmd := NewJobsCommand()

	// Check command structure
	assert.Equal(t, "jobs", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check subcommands are registered
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 2, "should have 2 subcommands: list and get")

	// Find list and get commands
	var hasListCmd, hasGetCmd bool
	for _, subcmd := range subcommands {
		if subcmd.Use == "list" {
			hasListCmd = true
		}
		if subcmd.Use[:3] == "get" { // "get <job-id>"
			hasGetCmd = true
		}
	}

	assert.True(t, hasListCmd, "should have list subcommand")
	assert.True(t, hasGetCmd, "should have get subcommand")
}

func TestJobsListCommand(t *testing.T) {
	cmd := NewJobsListCommand()

	// Check command structure
	assert.Equal(t, "list", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags exist
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("scanner"))
	assert.NotNil(t, flags.Lookup("state"))
	assert.NotNil(t, flags.Lookup("limit"))
	assert.NotNil(t, flags.Lookup("offset"))
	assert.NotNil(t, flags.Lookup("order-by"))
	assert.NotNil(t, flags.Lookup("no-color"))
}

func TestJobsGetCommand(t *testing.T) {
	cmd := NewJobsGetCommand()

	// Check command structure
	assert.Contains(t, cmd.Use, "get")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags exist
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("watch"))
	assert.NotNil(t, flags.Lookup("interval"))
	assert.NotNil(t, flags.Lookup("no-color"))

	// Check args validation
	assert.NotNil(t, cmd.Args)
}
