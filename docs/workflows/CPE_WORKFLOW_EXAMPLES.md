# CPE Matching Workflow - Usage Examples

This document provides practical examples for using the CPE matching workflow (M5-T3) in the Spectra-Red Intel Mesh project.

---

## Table of Contents

1. [Environment Setup](#environment-setup)
2. [Service CPE Generation](#service-cpe-generation)
3. [NVD API Queries](#nvd-api-queries)
4. [Workflow Invocation](#workflow-invocation)
5. [Database Queries](#database-queries)
6. [Batch Processing](#batch-processing)

---

## Environment Setup

### Required Environment Variables

```bash
# SurrealDB Connection
export SURREALDB_URL="ws://localhost:8000/rpc"
export SURREALDB_USER="root"
export SURREALDB_PASS="root"
export SURREALDB_NAMESPACE="spectra"
export SURREALDB_DATABASE="intel_mesh"

# NVD API (Optional - for higher rate limits)
export NVD_API_KEY="your-nvd-api-key-here"

# Workflow Service Port
export PORT="9080"
```

### Start the Workflow Service

```bash
cd cmd/workflows
go run main.go
```

Output:
```
{"level":"info","msg":"initializing Spectra-Red workflow service","port":"9080"}
{"level":"info","msg":"connected to SurrealDB successfully"}
{"level":"warn","msg":"NVD_API_KEY not set, using public rate limit (5 req/30s)"}
{"level":"info","msg":"workflows initialized","nvd_api_key_configured":false}
{"level":"info","msg":"workflow service starting","address":":9080"}
```

---

## Service CPE Generation

### Example 1: Generate CPE from Product/Version

```go
package main

import (
    "fmt"
    "github.com/spectra-red/recon/internal/enrichment"
)

func main() {
    service := enrichment.ServiceInfo{
        ID:      "service:web-001",
        Name:    "http",
        Product: "nginx",
        Version: "1.24.0",
    }

    cpes := enrichment.GenerateCPE(service)

    for _, cpe := range cpes {
        fmt.Printf("CPE: %s\n", cpe.CPE)
        fmt.Printf("  Vendor: %s\n", cpe.Vendor)
        fmt.Printf("  Product: %s\n", cpe.Product)
        fmt.Printf("  Version: %s\n", cpe.Version)
    }
}
```

Output:
```
CPE: cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*
  Vendor: nginx
  Product: nginx
  Version: 1.24.0
```

### Example 2: Parse Banner to Generate CPE

```go
service := enrichment.ServiceInfo{
    ID:     "service:ssh-001",
    Name:   "ssh",
    Banner: "SSH-2.0-OpenSSH_9.0p1 Ubuntu-1ubuntu7.3",
}

cpes := enrichment.GenerateCPE(service)
// Result: cpe:2.3:a:openbsd:openssh:9.0p1:*:*:*:*:*:*:*
```

### Example 3: Batch CPE Generation

```go
services := []enrichment.ServiceInfo{
    {ID: "s1", Product: "nginx", Version: "1.24.0"},
    {ID: "s2", Product: "apache", Version: "2.4.57"},
    {ID: "s3", Banner: "Redis server v=7.0.12"},
}

serviceCPEs := enrichment.GenerateCPEBatch(services)

for serviceID, cpes := range serviceCPEs {
    fmt.Printf("Service %s:\n", serviceID)
    for _, cpe := range cpes {
        fmt.Printf("  - %s\n", cpe.CPE)
    }
}
```

Output:
```
Service s1:
  - cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*
Service s2:
  - cpe:2.3:a:apache:apache:2.4.57:*:*:*:*:*:*:*
Service s3:
  - cpe:2.3:a:redis:redis:7.0.12:*:*:*:*:*:*:*
```

---

## NVD API Queries

### Example 4: Query NVD for Vulnerabilities

```go
package main

import (
    "context"
    "fmt"
    "github.com/spectra-red/recon/internal/enrichment"
)

func main() {
    // Create NVD client (with or without API key)
    client := enrichment.NewNVDClient("")  // Use "" for public API

    ctx := context.Background()
    cpe := "cpe:2.3:a:nginx:nginx:1.18.0:*:*:*:*:*:*:*"

    cves, err := client.QueryByCPE(ctx, cpe)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d vulnerabilities for nginx 1.18.0\n", len(cves))
    for _, cve := range cves {
        fmt.Printf("\n%s (CVSS: %.1f, Severity: %s)\n",
            cve.CVEID, cve.CVSS, cve.Severity)
        fmt.Printf("  %s\n", cve.Description[:100])
    }
}
```

### Example 5: Batch NVD Queries

```go
client := enrichment.NewNVDClient(os.Getenv("NVD_API_KEY"))

cpes := []string{
    "cpe:2.3:a:nginx:nginx:1.24.0:*:*:*:*:*:*:*",
    "cpe:2.3:a:apache:http_server:2.4.57:*:*:*:*:*:*:*",
}

cvesByCPE, err := client.QueryByCPEBatch(context.Background(), cpes)

for cpe, cves := range cvesByCPE {
    fmt.Printf("%s: %d vulnerabilities\n", cpe, len(cves))
}
```

### Example 6: Filter High-Severity Vulnerabilities

```go
matches := []enrichment.VulnMatch{
    {ServiceID: "s1", CVE: "CVE-2023-1001", CVSS: 9.8, Severity: "CRITICAL"},
    {ServiceID: "s2", CVE: "CVE-2023-1002", CVSS: 5.3, Severity: "MEDIUM"},
    {ServiceID: "s3", CVE: "CVE-2023-1003", CVSS: 7.5, Severity: "HIGH"},
}

// Filter to only HIGH and CRITICAL
highSeverity := enrichment.FilterHighSeverity(matches)
fmt.Printf("High-severity vulnerabilities: %d\n", len(highSeverity))
```

---

## Workflow Invocation

### Example 7: Invoke CPE Enrichment Workflow via Restate CLI

```bash
# Install Restate CLI if not already installed
curl -sSf https://get.restate.dev | sh

# Invoke the workflow
restate invocations invoke \
  --name EnrichCPEWorkflow/Run \
  --json '{
    "services": [
      {
        "id": "service:web-prod-001",
        "name": "http",
        "product": "nginx",
        "version": "1.24.0"
      },
      {
        "id": "service:ssh-prod-001",
        "name": "ssh",
        "banner": "SSH-2.0-OpenSSH_8.9p1"
      },
      {
        "id": "service:db-prod-001",
        "product": "postgresql",
        "version": "15.4"
      }
    ],
    "batch_id": "scan-2025-11-01-001"
  }'
```

Response:
```json
{
  "batch_id": "scan-2025-11-01-001",
  "services_processed": 3,
  "cpes_generated": 3,
  "vulns_found": 12,
  "relationships_created": 8
}
```

### Example 8: Invoke Workflow via HTTP (curl)

```bash
curl -X POST http://localhost:9080/EnrichCPEWorkflow/Run \
  -H "Content-Type: application/json" \
  -d '{
    "services": [
      {
        "id": "service:123",
        "product": "nginx",
        "version": "1.24.0"
      }
    ],
    "batch_id": "manual-test-001"
  }'
```

---

## Database Queries

### Example 9: Query Services with Vulnerabilities

```sql
-- Find all services with CRITICAL vulnerabilities
SELECT
    service.id,
    service.product,
    service.version,
    ->AFFECTED_BY->vuln.cve_id AS vulnerabilities,
    ->AFFECTED_BY->vuln.cvss AS cvss_scores
FROM service
WHERE ->AFFECTED_BY->vuln.severity = 'CRITICAL';
```

### Example 10: Get Vulnerability Details

```sql
-- Get detailed vulnerability information
SELECT
    vuln.cve_id,
    vuln.cvss,
    vuln.severity,
    vuln_doc.summary,
    vuln_doc.exploit_refs
FROM vuln
FETCH vuln_doc
WHERE vuln.severity IN ['CRITICAL', 'HIGH']
ORDER BY vuln.cvss DESC
LIMIT 20;
```

### Example 11: Find Services Without CPE

```sql
-- Find services that need CPE enrichment
SELECT * FROM service
WHERE cpe IS NONE OR array::len(cpe) = 0
LIMIT 100;
```

### Example 12: Get Vulnerability Count by Service

```sql
-- Count vulnerabilities per service
SELECT
    id,
    product,
    version,
    count(->AFFECTED_BY) AS vuln_count,
    max(->AFFECTED_BY->vuln.cvss) AS max_cvss
FROM service
GROUP BY id, product, version
HAVING vuln_count > 0
ORDER BY max_cvss DESC;
```

---

## Batch Processing

### Example 13: Process Services in Batches

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/spectra-red/recon/internal/workflows"
)

func main() {
    // Get services that need CPE enrichment
    filter := workflows.ServiceFilter{
        OnlyMissingCPE: true,
        Limit:          50,  // Process 50 at a time
        MinLastSeen:    &time.Time{},  // Optional: only recent services
    }

    services, err := workflows.GetServicesByFilter(db, filter)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Processing %d services\n", len(services))

    // Invoke workflow (would use Restate client in production)
    // This is pseudocode - actual invocation via Restate SDK
}
```

### Example 14: Scheduled Batch Processing Script

```bash
#!/bin/bash
# cron job: 0 2 * * * /path/to/cpe_batch_enrich.sh

# Process services without CPE identifiers
restate invocations invoke \
  --name EnrichCPEWorkflow/Run \
  --json "$(cat <<EOF
{
  "services": $(curl -s http://localhost:3000/v0/services?missing_cpe=true&limit=50),
  "batch_id": "scheduled-$(date +%Y%m%d-%H%M%S)"
}
EOF
)"
```

---

## Banner Parsing Examples

### Supported Banner Formats

```go
// SSH
"SSH-2.0-OpenSSH_9.0" → openssh 9.0
"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.1" → openssh 8.9p1

// HTTP Servers
"nginx/1.24.0" → nginx 1.24.0
"Apache/2.4.57 (Unix)" → apache 2.4.57
"Microsoft-IIS/10.0" → iis 10.0

// Databases
"MySQL/8.0.35" → mysql 8.0.35
"PostgreSQL 15.4" → postgresql 15.4
"Redis server v=7.0.12" → redis 7.0.12
"MongoDB 6.0.11" → mongodb 6.0.11

// Mail Servers
"Postfix 3.7.2" → postfix 3.7.2
"Exim 4.96" → exim 4.96

// FTP
"ProFTPD 1.3.8" → proftpd 1.3.8
"vsftpd 3.0.5" → vsftpd 3.0.5
```

---

## Monitoring and Debugging

### Example 15: Check NVD Cache Stats

```go
client := enrichment.NewNVDClient("")

// Check cache size (not directly exposed, but you can monitor via logging)
// In production, implement cache metrics
```

### Example 16: Monitor Rate Limiting

```bash
# Watch workflow logs for rate limiting
tail -f /var/log/spectra/workflows.log | grep "rate"

# Expected output:
# {"level":"info","msg":"waiting for rate limiter"}
# {"level":"info","msg":"rate limit available, proceeding"}
```

---

## Error Handling

### Example 17: Handle NVD API Errors

```go
cves, err := client.QueryByCPE(ctx, cpe)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "status 403") {
        fmt.Println("Rate limit exceeded or API key invalid")
    } else if strings.Contains(err.Error(), "timeout") {
        fmt.Println("NVD API timeout - will retry")
    } else {
        fmt.Printf("NVD query failed: %v\n", err)
    }
    return
}
```

---

## Performance Tuning

### Example 18: Optimal Batch Size

```go
// For public API (5 req/30s):
// Batch size: 5-10 services
// Expected duration: 30-60 seconds

// For API key (50 req/30s):
// Batch size: 20-50 services
// Expected duration: 10-30 seconds

filter := workflows.ServiceFilter{
    Limit: 20,  // Optimal for API key
}
```

---

## Testing

### Example 19: Unit Test a CPE Generation

```go
func TestCustomBannerParsing(t *testing.T) {
    service := enrichment.ServiceInfo{
        Banner: "MyCustomServer/3.2.1",
    }

    cpes := enrichment.GenerateCPE(service)

    if len(cpes) == 0 {
        t.Log("Banner not recognized - may need pattern")
    }
}
```

### Example 20: Integration Test with Mock NVD

```go
// Use httptest to mock NVD API
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Return mock CVE data
    json.NewEncoder(w).Encode(mockNVDResponse)
}))
defer server.Close()

// Point client to mock server (would need to modify NVDClient for testing)
```

---

## Troubleshooting

### Common Issues

**Issue**: No CPEs generated for service
```
Solution: Check banner format matches patterns, or add custom pattern
```

**Issue**: NVD API rate limit exceeded
```
Solution: Use API key or increase batch delay
```

**Issue**: Workflow timeout
```
Solution: Reduce batch size or increase workflow timeout
```

---

## Additional Resources

- **NVD API Documentation**: https://nvd.nist.gov/developers/vulnerabilities
- **CPE Specification**: https://nvd.nist.gov/products/cpe
- **Restate Documentation**: https://docs.restate.dev
- **SurrealDB Queries**: https://surrealdb.com/docs/surrealql

---

**Last Updated**: November 1, 2025
**Version**: 1.0.0
**Maintained by**: Spectra-Red Team
