package client

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIngestClient(t *testing.T) {
	client := NewIngestClient("http://localhost:3000", 30)

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:3000", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestIngestClient_Submit_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/mesh/ingest", r.URL.Path)

		// Verify headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("User-Agent"), "spectra-cli")

		// Parse request body
		var req IngestRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify request structure
		assert.NotEmpty(t, req.Data)
		assert.NotEmpty(t, req.PublicKey)
		assert.NotEmpty(t, req.Signature)
		assert.NotZero(t, req.Timestamp)

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(IngestResponse{
			JobID:     "job_abc123",
			Status:    "accepted",
			Message:   "Scan submitted successfully",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Generate test data
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	scanData := []byte(`{"hosts":[{"ip":"1.2.3.4"}]}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	// Submit request
	req := IngestRequest{
		Data:      json.RawMessage(scanData),
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	resp, err := client.Submit(req)
	require.NoError(t, err)

	// Verify response
	assert.Equal(t, "job_abc123", resp.JobID)
	assert.Equal(t, "accepted", resp.Status)
	assert.Equal(t, "Scan submitted successfully", resp.Message)
	assert.NotEmpty(t, resp.Timestamp)
}

func TestIngestClient_Submit_InvalidSignature(t *testing.T) {
	// Create a mock server that returns 401 Unauthorized
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(IngestErrorResponse{
			Error:     "invalid_signature",
			Message:   "Signature verification failed",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Create request with invalid signature
	req := IngestRequest{
		Data:      json.RawMessage(`{"test":"data"}`),
		PublicKey: "invalid",
		Signature: "invalid",
		Timestamp: time.Now().Unix(),
	}

	resp, err := client.Submit(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid_signature")
}

func TestIngestClient_Submit_ServerError(t *testing.T) {
	// Create a mock server that returns 500 Internal Server Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(IngestErrorResponse{
			Error:     "internal_error",
			Message:   "Failed to create job",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Create request
	req := IngestRequest{
		Data:      json.RawMessage(`{"test":"data"}`),
		PublicKey: "test",
		Signature: "test",
		Timestamp: time.Now().Unix(),
	}

	resp, err := client.Submit(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "internal_error")
}

func TestIngestClient_Submit_MalformedResponse(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Create request
	req := IngestRequest{
		Data:      json.RawMessage(`{"test":"data"}`),
		PublicKey: "test",
		Signature: "test",
		Timestamp: time.Now().Unix(),
	}

	resp, err := client.Submit(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestIngestClient_Submit_NetworkError(t *testing.T) {
	// Create client with invalid URL
	client := NewIngestClient("http://localhost:99999", 1)

	// Create request
	req := IngestRequest{
		Data:      json.RawMessage(`{"test":"data"}`),
		PublicKey: "test",
		Signature: "test",
		Timestamp: time.Now().Unix(),
	}

	resp, err := client.Submit(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestIngestClient_SubmitWithRetry_Success(t *testing.T) {
	// Create a server that fails once then succeeds
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount == 1 {
			// First attempt fails
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Second attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(IngestResponse{
			JobID:     "job_abc123",
			Status:    "accepted",
			Message:   "Scan submitted successfully",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Create request
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	scanData := []byte(`{"hosts":[{"ip":"1.2.3.4"}]}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	req := IngestRequest{
		Data:      json.RawMessage(scanData),
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	resp, err := client.SubmitWithRetry(req, 2)
	require.NoError(t, err)

	// Verify we retried and eventually succeeded
	assert.Equal(t, 2, attemptCount)
	assert.Equal(t, "job_abc123", resp.JobID)
}

func TestIngestClient_SubmitWithRetry_NoRetryOn4xx(t *testing.T) {
	// Create a server that always returns 400 Bad Request
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(IngestErrorResponse{
			Error:     "invalid_request",
			Message:   "Bad request",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Create request
	req := IngestRequest{
		Data:      json.RawMessage(`{"test":"data"}`),
		PublicKey: "test",
		Signature: "test",
		Timestamp: time.Now().Unix(),
	}

	resp, err := client.SubmitWithRetry(req, 3)
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Should not retry on 4xx errors
	assert.Equal(t, 1, attemptCount)
}

func TestIngestClient_SubmitWithRetry_MaxRetriesExceeded(t *testing.T) {
	// Create a server that always fails with 503
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 1)

	// Create request
	req := IngestRequest{
		Data:      json.RawMessage(`{"test":"data"}`),
		PublicKey: "test",
		Signature: "test",
		Timestamp: time.Now().Unix(),
	}

	resp, err := client.SubmitWithRetry(req, 2)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed after")

	// Should attempt initial + 2 retries = 3 total
	assert.Equal(t, 3, attemptCount)
}

// Benchmark tests
func BenchmarkIngestClient_Submit(b *testing.B) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(IngestResponse{
			JobID:     "job_abc123",
			Status:    "accepted",
			Message:   "Scan submitted successfully",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Create client
	client := NewIngestClient(server.URL, 10)

	// Generate test data
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	scanData := []byte(`{"hosts":[{"ip":"1.2.3.4"}]}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	req := IngestRequest{
		Data:      json.RawMessage(scanData),
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Submit(req)
	}
}
