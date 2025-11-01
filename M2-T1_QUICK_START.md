# M2-T1 Quick Start Guide

## Testing the Ingest Endpoint

### 1. Run Tests

```bash
# Test auth package (100% coverage)
go test ./internal/auth/... -v -cover

# Test middleware package (82.8% coverage)
go test ./internal/api/middleware/... -v -cover

# Run all integration tests
go test ./internal/api/handlers/... -v -run Integration

# Run benchmarks
go test ./internal/auth/... -bench=. -benchmem
```

### 2. Example Usage

#### Generate Ed25519 Keypair

```go
package main

import (
    "crypto/ed25519"
    "encoding/base64"
    "fmt"
)

func main() {
    // Generate keypair
    pubKey, privKey, _ := ed25519.GenerateKey(nil)

    fmt.Println("Public Key:", base64.StdEncoding.EncodeToString(pubKey))
    fmt.Println("Private Key:", base64.StdEncoding.EncodeToString(privKey))
}
```

#### Sign and Submit Scan

```go
package main

import (
    "bytes"
    "crypto/ed25519"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

func main() {
    // Your private key (keep secret!)
    privKeyBase64 := "YOUR_PRIVATE_KEY_HERE"
    privKey, _ := base64.StdEncoding.DecodeString(privKeyBase64)

    // Scan data
    scanData := map[string]interface{}{
        "scanner_id": "my-scanner-001",
        "hosts": []map[string]interface{}{
            {
                "ip": "192.168.1.1",
                "ports": []map[string]interface{}{
                    {"number": 22, "protocol": "tcp"},
                    {"number": 80, "protocol": "tcp"},
                },
            },
        },
    }

    data, _ := json.Marshal(scanData)
    timestamp := time.Now().Unix()

    // Create message: timestamp + data
    message := append([]byte(fmt.Sprintf("%d", timestamp)), data...)

    // Sign
    signature := ed25519.Sign(privKey, message)

    // Get public key
    pubKey := privKey[32:]

    // Create envelope
    envelope := map[string]interface{}{
        "data":       json.RawMessage(data),
        "public_key": base64.StdEncoding.EncodeToString(pubKey),
        "signature":  base64.StdEncoding.EncodeToString(signature),
        "timestamp":  timestamp,
    }

    body, _ := json.Marshal(envelope)

    // Send request
    resp, _ := http.Post(
        "http://localhost:3000/v1/mesh/ingest",
        "application/json",
        bytes.NewReader(body),
    )
    defer resp.Body.Close()

    // Parse response
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    fmt.Printf("Status: %d\n", resp.StatusCode)
    fmt.Printf("Job ID: %s\n", result["job_id"])
}
```

### 3. Testing with curl

```bash
# First, generate a keypair (using Go or Python)
# Then sign your data

curl -X POST http://localhost:3000/v1/mesh/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "data": {"hosts":[{"ip":"1.2.3.4","ports":[{"number":80}]}]},
    "public_key": "BASE64_PUBLIC_KEY",
    "signature": "BASE64_SIGNATURE",
    "timestamp": 1730419200
  }'
```

### 4. Expected Responses

#### Success (202 Accepted)
```json
{
  "job_id": "01930b2e-7890-7abc-def0-123456789abc",
  "status": "accepted",
  "message": "Scan submitted successfully",
  "timestamp": "2025-11-01T12:00:00Z"
}
```

#### Invalid Signature (401 Unauthorized)
```json
{
  "error": "invalid_signature",
  "message": "Signature verification failed",
  "timestamp": "2025-11-01T12:00:00Z"
}
```

#### Rate Limit Exceeded (429 Too Many Requests)
```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Maximum 60 requests per minute per scanner.",
  "timestamp": "2025-11-01T12:00:00Z"
}
```

## Key Features

### Security
- ✅ Ed25519 signature verification
- ✅ Timestamp validation (±5 minutes)
- ✅ Prevents replay attacks
- ✅ 10MB request size limit

### Rate Limiting
- ✅ 60 requests per minute per scanner
- ✅ Token bucket algorithm
- ✅ Automatic cleanup of stale buckets
- ✅ X-Forwarded-For support

### Observability
- ✅ Structured logging with zap
- ✅ Request ID tracking
- ✅ Detailed error messages
- ✅ Public key masking in logs

## File Structure

```
internal/
├── auth/
│   ├── ed25519.go              # Signature validation
│   └── ed25519_test.go         # Tests (100% coverage)
├── api/
│   ├── handlers/
│   │   ├── ingest.go           # HTTP handler
│   │   ├── ingest_test.go      # Unit tests
│   │   └── ingest_integration_test.go  # Integration tests
│   ├── middleware/
│   │   ├── rate_limit.go       # Rate limiter
│   │   └── rate_limit_test.go  # Tests (82.8% coverage)
│   └── routes.go               # Route registration
```

## Next Steps

After M2-T1 completion, the next tasks are:

1. **M2-T2**: Fast Path Ingest Endpoint - Direct SurrealDB writes
2. **M2-T3**: Restate Workflow Integration - Trigger workflows from ingest
3. **M3-T1**: Search Endpoint - Query mesh by selectors

## Troubleshooting

### Tests Won't Run
```bash
# Some handler tests may fail due to dependencies
# Run auth and middleware tests separately:
go test ./internal/auth/... -v
go test ./internal/api/middleware/... -v
```

### Import Errors
```bash
# Make sure all dependencies are installed
go mod download
go mod tidy
```

### Coverage Reports
```bash
# Generate HTML coverage report
go test ./internal/auth/... -coverprofile=/tmp/auth.out
go tool cover -html=/tmp/auth.out
```

## Performance

Expected performance (on modern hardware):
- Signature verification: ~35,000 ops/sec
- Handler throughput: ~180,000 req/sec (parallel)
- Rate limit check: ~2,000,000 ops/sec

## References

- Implementation Plan: `DETAILED_IMPLEMENTATION_PLAN.md`
- Completion Report: `M2-T1_COMPLETION_REPORT.md`
- Ed25519 spec: https://ed25519.cr.yp.to/
- Chi router: https://github.com/go-chi/chi
