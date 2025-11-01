package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyEnvelope(t *testing.T) {
	// Generate a test keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	// Create valid test data
	validData := json.RawMessage(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80,"protocol":"tcp"}]}]}`)
	validTimestamp := time.Now().Unix()

	// Sign the data
	message := append([]byte(fmt.Sprintf("%d", validTimestamp)), validData...)
	signature := ed25519.Sign(privKey, message)

	tests := []struct {
		name    string
		env     ScanEnvelope
		wantErr error
	}{
		{
			name: "valid envelope",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: validTimestamp,
			},
			wantErr: nil,
		},
		{
			name: "empty data",
			env: ScanEnvelope{
				Data:      nil,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: validTimestamp,
			},
			wantErr: ErrMissingData,
		},
		{
			name: "empty public key",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: "",
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: validTimestamp,
			},
			wantErr: ErrMissingData,
		},
		{
			name: "empty signature",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: "",
				Timestamp: validTimestamp,
			},
			wantErr: ErrMissingData,
		},
		{
			name: "zero timestamp",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: 0,
			},
			wantErr: ErrMissingData,
		},
		{
			name: "expired timestamp - too old",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: time.Now().Add(-10 * time.Minute).Unix(),
			},
			wantErr: ErrExpiredTimestamp,
		},
		{
			name: "expired timestamp - too far in future",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: time.Now().Add(10 * time.Minute).Unix(),
			},
			wantErr: ErrExpiredTimestamp,
		},
		{
			name: "invalid public key - not base64",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: "not-valid-base64!@#$",
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: validTimestamp,
			},
			wantErr: ErrInvalidPublicKey,
		},
		{
			name: "invalid public key - wrong length",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString([]byte("short")),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: validTimestamp,
			},
			wantErr: ErrInvalidPublicKey,
		},
		{
			name: "invalid signature - not base64",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: "not-valid-base64!@#$",
				Timestamp: validTimestamp,
			},
			wantErr: ErrInvalidSignature,
		},
		{
			name: "invalid signature - wrong length",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString([]byte("short")),
				Timestamp: validTimestamp,
			},
			wantErr: ErrInvalidSignature,
		},
		{
			name: "tampered data",
			env: ScanEnvelope{
				Data:      json.RawMessage(`{"hosts":[{"ip":"5.6.7.8"}]}`), // Different data
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: validTimestamp,
			},
			wantErr: ErrInvalidSignature,
		},
		{
			name: "wrong signature",
			env: ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(make([]byte, ed25519.SignatureSize)),
				Timestamp: validTimestamp,
			},
			wantErr: ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyEnvelope(tt.env)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyEnvelope_TimestampBoundary(t *testing.T) {
	// Test timestamps at the boundary of the acceptable window
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	validData := json.RawMessage(`{"test":"data"}`)

	tests := []struct {
		name      string
		timestamp time.Time
		wantErr   bool
	}{
		{
			name:      "4 minutes old - should pass",
			timestamp: time.Now().Add(-4 * time.Minute),
			wantErr:   false,
		},
		{
			name:      "4 minutes in future - should pass",
			timestamp: time.Now().Add(4 * time.Minute),
			wantErr:   false,
		},
		{
			name:      "6 minutes old - should fail",
			timestamp: time.Now().Add(-6 * time.Minute),
			wantErr:   true,
		},
		{
			name:      "6 minutes in future - should fail",
			timestamp: time.Now().Add(6 * time.Minute),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp := tt.timestamp.Unix()
			message := append([]byte(fmt.Sprintf("%d", timestamp)), validData...)
			signature := ed25519.Sign(privKey, message)

			env := ScanEnvelope{
				Data:      validData,
				PublicKey: base64.StdEncoding.EncodeToString(pubKey),
				Signature: base64.StdEncoding.EncodeToString(signature),
				Timestamp: timestamp,
			}

			err := VerifyEnvelope(env)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifySignature(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	message := []byte("test message")
	signature := ed25519.Sign(privKey, message)

	tests := []struct {
		name      string
		publicKey []byte
		message   []byte
		signature []byte
		wantErr   error
	}{
		{
			name:      "valid signature",
			publicKey: pubKey,
			message:   message,
			signature: signature,
			wantErr:   nil,
		},
		{
			name:      "invalid public key length",
			publicKey: []byte("short"),
			message:   message,
			signature: signature,
			wantErr:   ErrInvalidPublicKey,
		},
		{
			name:      "invalid signature length",
			publicKey: pubKey,
			message:   message,
			signature: []byte("short"),
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "wrong signature",
			publicKey: pubKey,
			message:   message,
			signature: make([]byte, ed25519.SignatureSize),
			wantErr:   ErrInvalidSignature,
		},
		{
			name:      "tampered message",
			publicKey: pubKey,
			message:   []byte("different message"),
			signature: signature,
			wantErr:   ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifySignature(tt.publicKey, tt.message, tt.signature)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateTestKey(t *testing.T) {
	pubKey, privKey, err := GenerateTestKey()
	require.NoError(t, err)

	assert.Len(t, pubKey, ed25519.PublicKeySize)
	assert.Len(t, privKey, ed25519.PrivateKeySize)

	// Test that keys can be used to sign and verify
	message := []byte("test message")
	signature := ed25519.Sign(privKey, message)
	assert.True(t, ed25519.Verify(pubKey, message, signature))
}

func TestVerifyEnvelope_RealWorldScenario(t *testing.T) {
	// Simulate a real-world scan submission
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	// Create scan data
	scanData := map[string]interface{}{
		"scanner_id": "test-scanner-001",
		"hosts": []map[string]interface{}{
			{
				"ip": "192.168.1.1",
				"ports": []map[string]interface{}{
					{"number": 22, "protocol": "tcp", "state": "open"},
					{"number": 80, "protocol": "tcp", "state": "open"},
					{"number": 443, "protocol": "tcp", "state": "open"},
				},
			},
			{
				"ip": "192.168.1.2",
				"ports": []map[string]interface{}{
					{"number": 3306, "protocol": "tcp", "state": "open"},
				},
			},
		},
	}

	data, err := json.Marshal(scanData)
	require.NoError(t, err)

	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), data...)
	signature := ed25519.Sign(privKey, message)

	env := ScanEnvelope{
		Data:      data,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	err = VerifyEnvelope(env)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkVerifyEnvelope(b *testing.B) {
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	require.NoError(b, err)

	data := json.RawMessage(`{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80}]}]}`)
	timestamp := time.Now().Unix()
	message := append([]byte(fmt.Sprintf("%d", timestamp)), data...)
	signature := ed25519.Sign(privKey, message)

	env := ScanEnvelope{
		Data:      data,
		PublicKey: base64.StdEncoding.EncodeToString(pubKey),
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: timestamp,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyEnvelope(env)
	}
}
