package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidSignature is returned when signature verification fails
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidPublicKey is returned when the public key is malformed
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrExpiredTimestamp is returned when the timestamp is outside acceptable window
	ErrExpiredTimestamp = errors.New("timestamp outside acceptable window")
	// ErrMissingData is returned when required envelope data is missing
	ErrMissingData = errors.New("missing required envelope data")
)

// TimestampWindow defines the acceptable time window for request timestamps (±5 minutes)
const TimestampWindow = 5 * time.Minute

// ScanEnvelope represents a signed scan submission
type ScanEnvelope struct {
	Data      json.RawMessage `json:"data"`
	PublicKey string          `json:"public_key"`
	Signature string          `json:"signature"`
	Timestamp int64           `json:"timestamp"`
}

// VerifyEnvelope validates the Ed25519 signature on a scan envelope
// It performs the following checks:
// 1. Timestamp freshness (±5 minutes from current time)
// 2. Public key format validation
// 3. Signature format validation
// 4. Cryptographic signature verification
func VerifyEnvelope(env ScanEnvelope) error {
	// Validate required fields
	if len(env.Data) == 0 {
		return fmt.Errorf("%w: data is empty", ErrMissingData)
	}
	if env.PublicKey == "" {
		return fmt.Errorf("%w: public_key is empty", ErrMissingData)
	}
	if env.Signature == "" {
		return fmt.Errorf("%w: signature is empty", ErrMissingData)
	}
	if env.Timestamp == 0 {
		return fmt.Errorf("%w: timestamp is zero", ErrMissingData)
	}

	// Check timestamp freshness
	requestTime := time.Unix(env.Timestamp, 0)
	now := time.Now()
	timeDiff := now.Sub(requestTime).Abs()

	if timeDiff > TimestampWindow {
		return fmt.Errorf("%w: timestamp %v is %v from current time (max %v)",
			ErrExpiredTimestamp, requestTime, timeDiff, TimestampWindow)
	}

	// Decode base64-encoded public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(env.PublicKey)
	if err != nil {
		return fmt.Errorf("%w: failed to decode public key: %v", ErrInvalidPublicKey, err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: public key must be %d bytes, got %d",
			ErrInvalidPublicKey, ed25519.PublicKeySize, len(pubKeyBytes))
	}

	// Decode base64-encoded signature
	sigBytes, err := base64.StdEncoding.DecodeString(env.Signature)
	if err != nil {
		return fmt.Errorf("%w: failed to decode signature: %v", ErrInvalidSignature, err)
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return fmt.Errorf("%w: signature must be %d bytes, got %d",
			ErrInvalidSignature, ed25519.SignatureSize, len(sigBytes))
	}

	// Construct the message that was signed
	// Format: timestamp + data (this ensures timestamp binding)
	message := append([]byte(fmt.Sprintf("%d", env.Timestamp)), env.Data...)

	// Verify the signature
	if !ed25519.Verify(pubKeyBytes, message, sigBytes) {
		return ErrInvalidSignature
	}

	return nil
}

// VerifySignature is a lower-level function that verifies an Ed25519 signature
// This is useful for custom signing schemes or testing
func VerifySignature(publicKey, message, signature []byte) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: public key must be %d bytes, got %d",
			ErrInvalidPublicKey, ed25519.PublicKeySize, len(publicKey))
	}

	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("%w: signature must be %d bytes, got %d",
			ErrInvalidSignature, ed25519.SignatureSize, len(signature))
	}

	if !ed25519.Verify(publicKey, message, signature) {
		return ErrInvalidSignature
	}

	return nil
}

// GenerateTestKey generates an Ed25519 keypair for testing purposes
// Returns (publicKey, privateKey, error)
func GenerateTestKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}
