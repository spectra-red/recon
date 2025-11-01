package handlers

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spectra-red/recon/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestIngestHandler_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	// Generate test keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	// Create valid scan data
	scanData := json.RawMessage(`{
		"scanner_id": "test-001",
		"hosts": [
			{"ip": "192.168.1.1", "ports": [{"number": 80, "protocol": "tcp"}]}
		]
	}`)

	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	body, err := json.Marshal(envelope)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response IngestResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.JobID)
	assert.Equal(t, "accepted", response.Status)
	assert.NotEmpty(t, response.Timestamp)

	// Validate job ID is a valid UUID
	_, err = uuid.Parse(response.JobID)
	assert.NoError(t, err, "job_id should be a valid UUID")
}

func TestIngestHandler_InvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_json", response.Error)
	assert.Contains(t, response.Message, "Invalid JSON")
}

func TestIngestHandler_InvalidSignature(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	// Create envelope with invalid signature
	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	scanData := json.RawMessage(`{"test": "data"}`)
	invalidSignature := make([]byte, ed25519.SignatureSize)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(invalidSignature),
		Timestamp: time.Now().Unix(),
	}

	body, err := json.Marshal(envelope)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_signature", response.Error)
}

func TestIngestHandler_ExpiredTimestamp(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	scanData := json.RawMessage(`{"test": "data"}`)
	oldTimestamp := time.Now().Add(-10 * time.Minute).Unix()
	message := append([]byte(fmt.Sprintf("%d", oldTimestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: oldTimestamp,
	}

	body, err := json.Marshal(envelope)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "invalid_signature", response.Error)
}

func TestIngestHandler_MissingData(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	tests := []struct {
		name     string
		envelope auth.ScanEnvelope
	}{
		{
			name: "missing data",
			envelope: auth.ScanEnvelope{
				Data:      nil,
				PublicKey: "dGVzdA==",
				Signature: "dGVzdA==",
				Timestamp: time.Now().Unix(),
			},
		},
		{
			name: "missing public key",
			envelope: auth.ScanEnvelope{
				Data:      json.RawMessage(`{"test": "data"}`),
				PublicKey: "",
				Signature: "dGVzdA==",
				Timestamp: time.Now().Unix(),
			},
		},
		{
			name: "missing signature",
			envelope: auth.ScanEnvelope{
				Data:      json.RawMessage(`{"test": "data"}`),
				PublicKey: "dGVzdA==",
				Signature: "",
				Timestamp: time.Now().Unix(),
			},
		},
		{
			name: "missing timestamp",
			envelope: auth.ScanEnvelope{
				Data:      json.RawMessage(`{"test": "data"}`),
				PublicKey: "dGVzdA==",
				Signature: "dGVzdA==",
				Timestamp: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.envelope)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestIngestHandler_RequestBodyTooLarge(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	// Create a 15MB payload (exceeds 10MB limit)
	largeData := make([]byte, 15*1024*1024)
	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(largeData))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should fail to parse JSON since we hit the limit
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIngestHandler_ContentTypeHandling(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	scanData := json.RawMessage(`{"test": "data"}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	body, err := json.Marshal(envelope)
	require.NoError(t, err)

	// Test without Content-Type header (should still work)
	req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestIngestHandler_MultipleRequests(t *testing.T) {
	// Test that handler can process multiple requests (idempotency check)
	logger := zaptest.NewLogger(t)
	handler := IngestHandler(logger)

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	scanData := json.RawMessage(`{"test": "data"}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	body, err := json.Marshal(envelope)
	require.NoError(t, err)

	jobIDs := make(map[string]bool)

	// Submit same scan 5 times
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response IngestResponse
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Each request should get a unique job ID
		assert.NotContains(t, jobIDs, response.JobID)
		jobIDs[response.JobID] = true
	}

	assert.Len(t, jobIDs, 5, "should have 5 unique job IDs")
}

func TestGenerateJobID(t *testing.T) {
	// Test job ID generation
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := generateJobID()

		// Should be valid UUID
		_, err := uuid.Parse(id)
		assert.NoError(t, err)

		// Should be unique
		assert.NotContains(t, ids, id)
		ids[id] = true
	}

	assert.Len(t, ids, 100)
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		errorCode  string
		message    string
		statusCode int
	}{
		{
			name:       "bad request",
			errorCode:  "invalid_json",
			message:    "Invalid JSON format",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "unauthorized",
			errorCode:  "invalid_signature",
			message:    "Signature verification failed",
			statusCode: http.StatusUnauthorized,
		},
		{
			name:       "internal error",
			errorCode:  "internal_error",
			message:    "Something went wrong",
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeErrorResponse(w, tt.errorCode, tt.message, tt.statusCode)

			assert.Equal(t, tt.statusCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, tt.errorCode, response.Error)
			assert.Equal(t, tt.message, response.Message)
			assert.NotEmpty(t, response.Timestamp)

			// Validate timestamp format
			_, err = time.Parse(time.RFC3339, response.Timestamp)
			assert.NoError(t, err)
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
			key:      "12345678",
			expected: "12345678",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskPublicKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkIngestHandler(b *testing.B) {
	logger := zaptest.NewLogger(b)
	handler := IngestHandler(logger)

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	scanData := json.RawMessage(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80}]}]}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	body, err := json.Marshal(envelope)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkIngestHandler_Parallel(b *testing.B) {
	logger := zaptest.NewLogger(b)
	handler := IngestHandler(logger)

	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	scanData := json.RawMessage(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80}]}]}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), scanData...)
	signature := ed25519.Sign(privKey, message)

	envelope := auth.ScanEnvelope{
		Data:      scanData,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	body, err := json.Marshal(envelope)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/v1/mesh/ingest", bytes.NewReader(body))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	})
}

// Test helper to read full response body
func readBody(t *testing.T, r io.Reader) []byte {
	body, err := io.ReadAll(r)
	require.NoError(t, err)
	return body
}
