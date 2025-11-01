package db

import (
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestJobStateIsValid(t *testing.T) {
	tests := []struct {
		name     string
		state    models.JobState
		expected bool
	}{
		{
			name:     "pending is valid",
			state:    models.JobStatePending,
			expected: true,
		},
		{
			name:     "processing is valid",
			state:    models.JobStateProcessing,
			expected: true,
		},
		{
			name:     "completed is valid",
			state:    models.JobStateCompleted,
			expected: true,
		},
		{
			name:     "failed is valid",
			state:    models.JobStateFailed,
			expected: true,
		},
		{
			name:     "invalid state",
			state:    models.JobState("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJobCanTransition(t *testing.T) {
	tests := []struct {
		name        string
		currentState models.JobState
		newState     models.JobState
		expected     bool
	}{
		{
			name:         "pending to processing - valid",
			currentState: models.JobStatePending,
			newState:     models.JobStateProcessing,
			expected:     true,
		},
		{
			name:         "pending to failed - valid",
			currentState: models.JobStatePending,
			newState:     models.JobStateFailed,
			expected:     true,
		},
		{
			name:         "processing to completed - valid",
			currentState: models.JobStateProcessing,
			newState:     models.JobStateCompleted,
			expected:     true,
		},
		{
			name:         "processing to failed - valid",
			currentState: models.JobStateProcessing,
			newState:     models.JobStateFailed,
			expected:     true,
		},
		{
			name:         "completed to processing - invalid",
			currentState: models.JobStateCompleted,
			newState:     models.JobStateProcessing,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &models.Job{
				ID:         "test-job-id",
				State:      tt.currentState,
				ScannerKey: "test-key",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			result := job.CanTransition(tt.newState)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskPublicKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "long key",
			key:      "abcdefghijklmnopqrstuvwxyz",
			expected: "abcdefgh...",
		},
		{
			name:     "short key",
			key:      "abc",
			expected: "abc",
		},
		{
			name:     "exactly 8 chars",
			key:      "abcdefgh",
			expected: "abcdefgh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskPublicKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
