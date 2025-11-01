package cli

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spectra-red/recon/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadScanData_FromFile(t *testing.T) {
	// Create a temp file with test data
	testData := []byte(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80,"protocol":"tcp"}]}]}`)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "scan.json")

	err := os.WriteFile(tmpFile, testData, 0644)
	require.NoError(t, err)

	// Test reading from file
	data, err := readScanData(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestReadScanData_EmptyFile(t *testing.T) {
	// Create an empty temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.json")

	err := os.WriteFile(tmpFile, []byte{}, 0644)
	require.NoError(t, err)

	// Test reading from empty file
	_, err = readScanData(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data to submit")
}

func TestReadScanData_NonExistentFile(t *testing.T) {
	// Test reading from non-existent file
	_, err := readScanData("/tmp/non-existent-file-12345.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestSignScanData(t *testing.T) {
	// Generate a test keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	testData := []byte(`{"test":"data"}`)
	timestamp := time.Now().Unix()

	// Sign the data
	signature, err := signScanData(testData, timestamp, privKey)
	require.NoError(t, err)

	// Verify the signature matches expected format
	assert.Len(t, signature, ed25519.SignatureSize)

	// Verify the signature is valid
	message := append([]byte(fmt.Sprintf("%d", timestamp)), testData...)
	valid := ed25519.Verify(pubKey, message, signature)
	assert.True(t, valid, "Signature should be valid")
}

func TestSignScanData_DifferentTimestamps(t *testing.T) {
	// Generate a test keypair
	_, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	testData := []byte(`{"test":"data"}`)

	// Sign with two different timestamps
	timestamp1 := time.Now().Unix()
	signature1, err := signScanData(testData, timestamp1, privKey)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	timestamp2 := time.Now().Unix()
	signature2, err := signScanData(testData, timestamp2, privKey)
	require.NoError(t, err)

	// Signatures should be different because timestamps are different
	assert.NotEqual(t, signature1, signature2,
		"Signatures with different timestamps should differ")
}

func TestSignScanData_ConsistentSignatures(t *testing.T) {
	// Generate a test keypair
	_, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	testData := []byte(`{"test":"data"}`)
	timestamp := time.Now().Unix()

	// Sign the same data twice
	signature1, err := signScanData(testData, timestamp, privKey)
	require.NoError(t, err)

	signature2, err := signScanData(testData, timestamp, privKey)
	require.NoError(t, err)

	// Signatures should be identical for same data and timestamp
	assert.Equal(t, signature1, signature2,
		"Signatures should be consistent for same input")
}

func TestDisplayJSON(t *testing.T) {
	// Capture stdout for testing
	// This is a simple test - in production you'd redirect stdout
	// For now, just verify the function doesn't panic
	resp := &client.IngestResponse{
		JobID:     "job_123",
		Status:    "accepted",
		Message:   "Test message",
		Timestamp: "2025-11-01T12:00:00Z",
	}

	err := displayJSON(resp)
	// We can't easily test stdout in unit tests without mocking,
	// so just verify no error
	assert.NoError(t, err)
}

func TestDisplayYAML(t *testing.T) {
	resp := &client.IngestResponse{
		JobID:     "job_123",
		Status:    "accepted",
		Message:   "Test message",
		Timestamp: "2025-11-01T12:00:00Z",
	}

	err := displayYAML(resp)
	// We can't easily test stdout in unit tests without mocking,
	// so just verify no error
	assert.NoError(t, err)
}

func TestDisplayTable(t *testing.T) {
	resp := &client.IngestResponse{
		JobID:     "job_123",
		Status:    "accepted",
		Message:   "Test message",
		Timestamp: "2025-11-01T12:00:00Z",
	}

	err := displayTable(resp)
	// We can't easily test stdout in unit tests without mocking,
	// so just verify no error
	assert.NoError(t, err)
}

func TestDisplayIngestResponse_InvalidFormat(t *testing.T) {
	resp := &client.IngestResponse{
		JobID:     "job_123",
		Status:    "accepted",
		Message:   "Test message",
		Timestamp: "2025-11-01T12:00:00Z",
	}

	err := displayIngestResponse(resp, "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported output format")
}

// Benchmark tests
func BenchmarkSignScanData(b *testing.B) {
	_, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	testData := []byte(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80}]}]}`)
	timestamp := time.Now().Unix()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = signScanData(testData, timestamp, privKey)
	}
}

func BenchmarkReadScanData_SmallFile(b *testing.B) {
	testData := []byte(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80}]}]}`)
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "scan.json")

	err := os.WriteFile(tmpFile, testData, 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = readScanData(tmpFile)
	}
}

// TestSignatureCompatibility ensures our signing is compatible with the auth package
func TestSignatureCompatibility(t *testing.T) {
	// Generate a test keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	testData := []byte(`{"hosts":[{"ip":"1.2.3.4"}]}`)
	timestamp := time.Now().Unix()

	// Sign using our function
	signature, err := signScanData(testData, timestamp, privKey)
	require.NoError(t, err)

	// Manually verify using the same message construction
	message := append([]byte(fmt.Sprintf("%d", timestamp)), testData...)
	valid := ed25519.Verify(pubKey, message, signature)
	assert.True(t, valid, "Signature should be valid with manual verification")

	// Encode to base64 (as the API expects)
	pubKeyB64 := base64.StdEncoding.EncodeToString(pubKey)
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Verify these can be decoded back
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKeyB64)
	require.NoError(t, err)
	assert.Equal(t, pubKey, ed25519.PublicKey(decodedPubKey))

	decodedSignature, err := base64.StdEncoding.DecodeString(signatureB64)
	require.NoError(t, err)
	assert.Equal(t, signature, decodedSignature)
}
