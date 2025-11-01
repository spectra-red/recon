# Quick Start: Submitting Scans with `spectra ingest`

This guide shows you how to submit scan results to the Spectra-Red Intel Mesh using the `spectra ingest` command.

---

## Prerequisites

1. **API Server Running**
   ```bash
   go run cmd/api/main.go
   ```

2. **Ed25519 Keys Generated**
   ```bash
   spectra keys generate
   ```

   This creates:
   - `~/.spectra/.spectra.yaml` with your keys
   - Private key for signing requests
   - Public key for verification

---

## Basic Usage

### 1. Submit from File

Create a test scan file:
```bash
cat > scan-results.json << 'EOF'
{
  "hosts": [
    {
      "ip": "192.168.1.1",
      "ports": [
        {"number": 22, "protocol": "tcp", "state": "open"},
        {"number": 80, "protocol": "tcp", "state": "open"}
      ]
    }
  ]
}
EOF
```

Submit the scan:
```bash
spectra ingest scan-results.json
```

Expected output:
```
✓ Scan submitted successfully

  Job ID:    job_abc123xyz
  Status:    accepted
  Message:   Scan submitted successfully, processing asynchronously
  Timestamp: 2025-11-01T12:00:00Z

Track job status with: spectra jobs get job_abc123xyz
```

### 2. Submit from stdin

```bash
cat scan-results.json | spectra ingest -
```

### 3. Direct from Naabu

```bash
naabu -host example.com -json | spectra ingest -
```

---

## Output Formats

### JSON Output

```bash
spectra ingest scan.json --output json
```

Output:
```json
{
  "job_id": "job_abc123xyz",
  "status": "accepted",
  "message": "Scan submitted successfully, processing asynchronously",
  "timestamp": "2025-11-01T12:00:00Z"
}
```

### YAML Output

```bash
spectra ingest scan.json --output yaml
```

Output:
```yaml
---
job_id: job_abc123xyz
status: accepted
message: Scan submitted successfully, processing asynchronously
timestamp: 2025-11-01T12:00:00Z
```

### Table Output (Default)

```bash
spectra ingest scan.json --output table
```

Shows a nicely formatted table with the job details.

---

## Configuration

Edit `~/.spectra/.spectra.yaml`:

```yaml
api:
  url: http://localhost:3000    # API endpoint
  timeout: 30s                  # Request timeout

scanner:
  private_key: "..."            # Auto-generated
  public_key: "..."             # Auto-generated

output:
  format: table                 # json, yaml, or table
  color: true                   # Enable color output
```

---

## Environment Variables

Override config with environment variables:

```bash
# Override API URL
export SPECTRA_API_URL=https://api.spectra-red.com
spectra ingest scan.json

# Override timeout
export SPECTRA_API_TIMEOUT=60s
spectra ingest scan.json

# Override output format
export SPECTRA_OUTPUT_FORMAT=json
spectra ingest scan.json
```

---

## Real-World Examples

### Example 1: Scan and Submit Single Host

```bash
naabu -host example.com -json | spectra ingest -
```

### Example 2: Scan and Submit Multiple Hosts

```bash
# Create target list
cat > targets.txt << 'EOF'
example.com
test.com
demo.com
EOF

# Scan and submit
naabu -list targets.txt -json | spectra ingest -
```

### Example 3: Batch Process Scan Files

```bash
# Scan multiple targets
for domain in $(cat targets.txt); do
  naabu -host $domain -json > "scans/${domain}.json"
done

# Submit all scans
for scan in scans/*.json; do
  spectra ingest "$scan"
done
```

### Example 4: Continuous Scanning

```bash
# Monitor and submit in real-time
tail -f live-scan.json | spectra ingest -
```

---

## Troubleshooting

### Error: "no private key configured"

**Solution**: Generate keys first
```bash
spectra keys generate
```

### Error: "failed to send request: connection refused"

**Solution**: Ensure API server is running
```bash
go run cmd/api/main.go
```

### Error: "invalid JSON in scan data"

**Solution**: Validate your JSON
```bash
cat scan.json | jq .  # Validate with jq
```

### Error: "Signature verification failed"

**Solutions**:
1. Ensure your public key is registered with the API
2. Check that your private key hasn't been corrupted
3. Regenerate keys if needed: `spectra keys generate --force`

---

## Testing Your Setup

### 1. Create Test Data

```bash
echo '{"hosts":[{"ip":"1.2.3.4","ports":[{"number":80,"protocol":"tcp"}]}]}' > test.json
```

### 2. Submit Test Data

```bash
spectra ingest test.json
```

### 3. Verify Submission

```bash
# Get the job ID from the output above
spectra jobs get job_abc123xyz
```

### 4. Check API Logs

```bash
# In the API server terminal
# You should see:
# INFO scan received, job created
# INFO workflow triggered successfully
```

---

## Performance Tips

### Large Files

For files >1MB, you'll see a progress indicator:
```bash
spectra ingest large-scan.json
# Output: Submitting 2048576 bytes of scan data...
```

### Parallel Submissions

Submit multiple scans in parallel:
```bash
for scan in scans/*.json; do
  spectra ingest "$scan" &
done
wait
```

### Compression (Future)

Large scans can be compressed before submission:
```bash
gzip < scan.json | spectra ingest -  # Future feature
```

---

## Next Steps

1. **Monitor Jobs**: Use `spectra jobs list` to see all submissions
2. **Query Results**: Use `spectra query` to search the mesh
3. **Automate Scans**: Set up cron jobs for continuous scanning

---

## Help & Documentation

```bash
# Get help
spectra ingest --help

# View configuration
cat ~/.spectra/.spectra.yaml

# Check version
spectra version
```

---

## Example Workflow

Complete workflow from scan to submission:

```bash
# 1. Generate keys (one-time setup)
spectra keys generate

# 2. Start API server (in separate terminal)
go run cmd/api/main.go

# 3. Create target list
echo "example.com" > targets.txt

# 4. Scan targets
naabu -list targets.txt -json > scan-results.json

# 5. Submit to mesh
spectra ingest scan-results.json

# 6. Get job status
spectra jobs get job_abc123xyz

# 7. Query results (once processed)
spectra query --ip 93.184.216.34
```

---

That's it! You're now ready to submit scans to the Spectra-Red Intel Mesh.

For more information, see:
- `M4-T2_COMPLETION_REPORT.md` - Full implementation details
- `DETAILED_IMPLEMENTATION_PLAN.md` - Architecture and design
- API documentation at http://localhost:3000/docs
