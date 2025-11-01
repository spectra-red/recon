# ASN Enrichment - Quick Start Guide

## 5-Minute Setup

### 1. Start Services

```bash
# Start SurrealDB and Restate
docker-compose up -d

# Start workflow service
go run cmd/workflows/main.go
```

### 2. Trigger ASN Enrichment

```bash
# Enrich IP addresses
curl -X POST http://localhost:9080/EnrichASNWorkflow/enrich/send \
  -H "Content-Type: application/json" \
  -d '{
    "ips": ["8.8.8.8", "1.1.1.1", "9.9.9.9"]
  }'
```

### 3. View Results

**Restate UI**: http://localhost:9070

**Query enriched host**:
```sql
-- In SurrealDB console
SELECT *, ->IN_ASN->asn.* AS asn_info
FROM host
WHERE ip = '8.8.8.8';
```

## Common Operations

### Enrich Single IP
```bash
curl -X POST http://localhost:9080/EnrichASNWorkflow/enrich/send \
  -d '{"ips": ["8.8.8.8"]}'
```

### Enrich Batch (50 IPs)
```bash
curl -X POST http://localhost:9080/EnrichASNWorkflow/enrich/send \
  -d '{
    "ips": [
      "8.8.8.8", "8.8.4.4", "1.1.1.1", "1.0.0.1",
      "9.9.9.9", "149.112.112.112", "208.67.222.222"
    ]
  }'
```

### Force Refresh (bypass cache)
```bash
curl -X POST http://localhost:9080/EnrichASNWorkflow/enrich/send \
  -d '{"ips": ["8.8.8.8"], "force_refresh": true}'
```

## Response Examples

### Success
```json
{
  "total_ips": 3,
  "enriched_ips": 3,
  "cached_ips": 0,
  "failed_ips": 0,
  "asn_data": {
    "8.8.8.8": {"asn": 15169, "org": "GOOGLE, US", "country": "US"},
    "1.1.1.1": {"asn": 13335, "org": "CLOUDFLARENET, US", "country": "US"}
  }
}
```

### Partial Failure
```json
{
  "total_ips": 3,
  "enriched_ips": 2,
  "failed_ips": 1,
  "failed_ips_list": ["invalid.ip"],
  "asn_data": {...}
}
```

## Configuration

Edit `cmd/workflows/main.go`:

```go
// Adjust rate limit (requests per minute)
asnRateLimit := 50  // Default: 100

// Adjust cache TTL
asnCacheTTL := 12 * time.Hour  // Default: 24h
```

## Testing

```bash
# Unit tests
go test ./internal/enrichment/asn_test.go -v -short

# Integration tests (requires network)
go test ./internal/enrichment/asn_integration_test.go -v -tags=integration

# All tests
go test ./... -short
```

## Troubleshooting

### Workflow not starting
```bash
# Check Restate is running
curl http://localhost:9070/health

# Check workflow service logs
go run cmd/workflows/main.go
```

### Rate limit errors
```go
// Increase rate limit in cmd/workflows/main.go
asnRateLimit := 200  // Double the limit
```

### Cache not working
```go
// Check cache stats (add to workflow)
size, oldestEntry := asnClient.GetCacheStats()
log.Printf("Cache: %d entries, oldest: %v", size, oldestEntry)
```

## Performance Tips

1. **Use batches**: Process multiple IPs in single request (max 100)
2. **Leverage cache**: Default 24h TTL reduces API calls
3. **Don't force refresh**: Only use when ASN data changed
4. **Monitor rate limits**: Stay under 100 req/min (or configured limit)

## Integration with Ingest

Automatically enrich after scan ingestion:

```go
// In your ingest handler
ips := extractIPsFromScan(scanData)
enrichResponse := triggerASNEnrichment(ips)
log.Printf("Enriched %d/%d IPs", enrichResponse.EnrichedIPs, enrichResponse.TotalIPs)
```

## Query Enriched Data

```sql
-- Find all Google-owned hosts
SELECT * FROM host
WHERE ->IN_ASN->asn.number = 15169;

-- Hosts by country
SELECT * FROM host
WHERE country = 'US';

-- Complex query: All US Cloudflare hosts
SELECT * FROM host
WHERE country = 'US'
  AND ->IN_ASN->asn.number = 13335;
```

## Monitoring

### Key Metrics
- **Enrichment rate**: `enriched_ips / total_ips`
- **Cache hit ratio**: `cached_ips / total_ips`
- **Failure rate**: `failed_ips / total_ips`

### Logging
```bash
# View workflow logs
docker logs -f workflows-service
```

### Restate UI
- View running workflows: http://localhost:9070/workflows
- Inspect workflow state: Click on workflow ID
- Retry failed workflows: Click "Retry" button

## Next Steps

1. ✅ Read full documentation: `docs/ASN_ENRICHMENT.md`
2. ✅ Run integration tests: Verify network connectivity
3. ✅ Configure rate limits: Adjust for your needs
4. ✅ Monitor performance: Track enrichment metrics
5. ✅ Integrate with ingest: Auto-enrich new hosts

## Support

- Documentation: `/docs/ASN_ENRICHMENT.md`
- Tests: `/internal/enrichment/asn_test.go`
- Issues: Check workflow logs and Restate UI
