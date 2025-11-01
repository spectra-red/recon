package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IngestClient handles API requests to the /v1/mesh/ingest endpoint
type IngestClient struct {
	baseURL    string
	httpClient *http.Client
}

// IngestRequest represents the request body for submitting scans
type IngestRequest struct {
	Data      json.RawMessage `json:"data"`
	PublicKey string          `json:"public_key"`
	Signature string          `json:"signature"`
	Timestamp int64           `json:"timestamp"`
}

// IngestResponse represents the response from the ingest endpoint
type IngestResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// IngestErrorResponse represents an error response from the API
type IngestErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NewIngestClient creates a new ingest API client
func NewIngestClient(baseURL string, timeoutSeconds int) *IngestClient {
	return &IngestClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// Submit submits scan results to the mesh
func (c *IngestClient) Submit(req IngestRequest) (*IngestResponse, error) {
	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.baseURL + "/v1/mesh/ingest"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "spectra-cli/0.1.0")

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		var errResp IngestErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			// If we can't parse the error response, return the raw body
			return nil, &HTTPError{StatusCode: httpResp.StatusCode, Body: string(respBody)}
		}
		return nil, &APIError{StatusCode: httpResp.StatusCode, ErrorCode: errResp.Error, Message: errResp.Message}
	}

	// Parse success response
	var resp IngestResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// APIError represents a structured API error response
type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%s): %s", e.ErrorCode, e.Message)
}

// IsClientError returns true if the error is a 4xx client error
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// SubmitWithRetry submits scan results with automatic retry on transient failures
func (c *IngestClient) SubmitWithRetry(req IngestRequest, maxRetries int) (*IngestResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.Submit(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if isClientError(err) {
			return nil, err
		}

		// Wait before retry (exponential backoff)
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// isClientError checks if an error is a client error (4xx)
func isClientError(err error) bool {
	// Check if it's an APIError with 4xx status code
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsClientError()
	}

	// Check if it's an HTTPError with 4xx status code
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
	}

	return false
}
