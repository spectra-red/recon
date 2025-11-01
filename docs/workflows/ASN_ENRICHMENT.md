# ASN Enrichment Workflow

## Overview

The ASN (Autonomous System Number) enrichment workflow provides automatic enrichment of IP addresses with ASN data including:
- ASN number
- Organization name
- Country code

This workflow integrates with the Spectra-Red Intel Mesh to enhance host records with network ownership information.

## Architecture

### Components

1. **ASN Client** (`internal/enrichment/asn.go`)
   - Team Cymru whois-based ASN lookup
   - Built-in caching with configurable TTL (default: 24 hours)
   - Rate limiting (default: 100 req/min)
   - Batch processing support

2. **ASN Workflow** (`internal/workflows/enrich_asn.go`)
   - Restate durable workflow
   - Batch processing (max 100 IPs per invocation)
   - Automatic filtering of already-enriched hosts
   - Database updates with ASN information
   - Graph edge creation (host→IN_ASN→asn)

### Data Flow

```
┌─────────────────┐
│ Trigger Request │
│ (IPs to enrich) │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ Step 1: Filter IPs                  │
│ Check which hosts need enrichment   │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ Step 2: ASN Lookup (External API)   │
│ - Check cache first                 │
│ - Batch lookup from Team Cymru      │
│ - Rate limited                      │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ Step 3: Update Host Records         │
│ Update host.asn, host.country       │
└────────┬────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────┐
│ Step 4: Create ASN Nodes & Edges    │
│ - Upsert asn nodes                  │
│ - Create IN_ASN relationships       │
└─────────────────────────────────────┘
```

## Usage

### Triggering the Workflow

The ASN enrichment workflow can be triggered via Restate:

```bash
# Enrich a batch of IPs
curl -X POST http://localhost:8080/EnrichASNWorkflow/enrich/send \
  -H "Content-Type: application/json" \
  -d '{
    "ips": ["8.8.8.8", "1.1.1.1", "9.9.9.9"],
    "job_id": "asn-enrich-001",
    "force_refresh": false
  }'
```

### Request Parameters

- `ips` (required): Array of IP addresses to enrich (max 100)
- `job_id` (optional): Job identifier for tracking
- `force_refresh` (optional): Force re-lookup even if cached (default: false)

### Response Format

```json
{
  "total_ips": 10,
  "enriched_ips": 8,
  "cached_ips": 2,
  "failed_ips": 0,
  "failed_ips_list": [],
  "asn_data": {
    "8.8.8.8": {
      "asn": 15169,
      "org": "GOOGLE, US",
      "country": "US"
    },
    "1.1.1.1": {
      "asn": 13335,
      "org": "CLOUDFLARENET, US",
      "country": "US"
    }
  }
}
```

## Configuration

### Environment Variables

Configuration is set in `cmd/workflows/main.go`:

```go
asnRateLimit := 100                // Max requests per minute
asnCacheTTL := 24 * time.Hour      // Cache TTL duration
```

To customize, modify these values or read from environment variables:

```go
asnRateLimit := getEnvInt("ASN_RATE_LIMIT", 100)
asnCacheTTL := getEnvDuration("ASN_CACHE_TTL", 24*time.Hour)
```

### Rate Limiting

The ASN client implements token bucket rate limiting:
- Default: 100 requests per minute
- Configurable per client instance
- Automatic backpressure (blocks when limit reached)
- Context-aware (respects cancellation)

### Caching

Built-in cache with:
- TTL-based expiration (default: 24 hours)
- Thread-safe access
- Cache statistics via `GetCacheStats()`
- Manual cleanup via `ClearExpiredCache()`

## Database Schema

### Host Table Update

```sql
UPDATE host SET
  asn = 15169,
  country = 'US'
WHERE ip = '8.8.8.8';
```

### ASN Table

```sql
CREATE asn:15169 CONTENT {
  number: 15169,
  org: 'GOOGLE, US',
  country: 'US'
};
```

### IN_ASN Relationship

```sql
RELATE host:8_8_8_8->IN_ASN->asn:15169;
```

## Performance

### Benchmarks

- **Single lookup**: ~200-500ms (network latency)
- **Cached lookup**: <1ms
- **Batch lookup (50 IPs)**: ~2-5 seconds
- **Rate limit enforcement**: Transparent (automatic backpressure)

### Optimization Tips

1. **Use batch processing**: Always prefer `LookupBatch()` over multiple `LookupASN()` calls
2. **Leverage caching**: Default 24h TTL reduces external API calls
3. **Filter already-enriched hosts**: Workflow automatically filters hosts with existing ASN data
4. **Adjust rate limits**: Balance between speed and API provider constraints

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./internal/enrichment/... -v -short

# Run only ASN tests
go test ./internal/enrichment/asn_test.go -v -short

# Run workflow tests
go test ./internal/workflows/enrich_asn_test.go -v -short
```

### Integration Tests

Integration tests require network access to Team Cymru whois service:

```bash
# Run integration tests (requires network)
go test ./internal/enrichment/asn_integration_test.go -v -tags=integration

# Specific integration tests
go test -v -tags=integration -run TestASNClient_BatchProcessing
go test -v -tags=integration -run TestASNClient_Caching
go test -v -tags=integration -run TestASNClient_RateLimiting
```

### Manual Testing with Restate

1. Start services:
```bash
docker-compose up -d
go run cmd/workflows/main.go
```

2. Trigger workflow:
```bash
curl -X POST http://localhost:9080/EnrichASNWorkflow/enrich/send \
  -H "Content-Type: application/json" \
  -d '{"ips": ["8.8.8.8", "1.1.1.1"]}'
```

3. Check Restate UI: http://localhost:9070

## Error Handling

### Workflow Errors

The workflow handles errors gracefully:
- **Network errors**: Retried by Restate durable execution
- **Rate limit exceeded**: Automatic backpressure (waits for token)
- **Invalid IP**: Skipped, reported in `failed_ips_list`
- **Database errors**: Workflow fails but can be retried

### Error Response

```json
{
  "total_ips": 10,
  "enriched_ips": 7,
  "cached_ips": 0,
  "failed_ips": 3,
  "failed_ips_list": ["invalid.ip", "256.256.256.256", "incomplete.data"],
  "asn_data": {...}
}
```

## Best Practices

1. **Batch size**: Keep batches under 100 IPs for optimal performance
2. **Force refresh**: Only use `force_refresh: true` when absolutely necessary
3. **Job IDs**: Use meaningful job IDs for tracking and debugging
4. **Rate limiting**: Monitor rate limit usage to avoid throttling
5. **Cache management**: Run periodic cache cleanup for long-running services
6. **Error handling**: Check `failed_ips_list` and retry if necessary

## Integration with Ingest Workflow

To automatically enrich hosts after ingestion:

```go
// After M2-T3 ingest workflow completes
func (w *IngestWorkflow) Run(ctx restate.Context, req IngestWorkflowRequest) error {
    // ... ingest logic ...

    // Trigger ASN enrichment
    ips := extractIPsFromScanData(req.ScanData)
    restate.ServiceCall(
        restate.CallOptions{
            Service: "EnrichASNWorkflow",
            Method:  "Run",
        },
        EnrichASNRequest{
            IPs:   ips,
            JobID: req.JobID + "-asn",
        },
    )

    return nil
}
```

## Monitoring

### Metrics to Track

1. **Enrichment rate**: IPs enriched per minute
2. **Cache hit ratio**: `cached_ips / total_ips`
3. **Failure rate**: `failed_ips / total_ips`
4. **Lookup latency**: P50, P95, P99
5. **Rate limit saturation**: Token bucket usage

### Logging

The workflow logs key events:
- ASN lookup initiated (with batch size)
- Cache hits/misses
- Rate limit waits
- Database update results
- Error conditions

### Restate UI

Monitor workflow executions at: http://localhost:9070
- View running workflows
- Inspect workflow state
- Retry failed workflows
- View execution history

## Troubleshooting

### Common Issues

1. **Slow enrichment**
   - Check network latency to Team Cymru
   - Verify rate limit configuration
   - Monitor cache hit ratio

2. **High failure rate**
   - Verify IP address format
   - Check Team Cymru service availability
   - Review error logs for specific failures

3. **Rate limit errors**
   - Increase rate limit value (if appropriate)
   - Reduce batch sizes
   - Add delays between batches

4. **Cache not working**
   - Verify TTL configuration
   - Check cache statistics
   - Run manual cache cleanup

## Future Enhancements

1. **Multiple ASN sources**
   - MaxMind GeoLite2 ASN (MMDB)
   - IPtoASN.com API
   - Fallback mechanisms

2. **Enhanced caching**
   - Redis-backed cache
   - Distributed cache for multi-instance deployments

3. **ASN enrichment on-the-fly**
   - Enrich during ingest (single workflow)
   - Real-time enrichment for query results

4. **ASN analytics**
   - Top ASNs by host count
   - ASN change tracking
   - Cloud provider detection

## References

- [Team Cymru IP to ASN Mapping](https://www.team-cymru.com/ip-asn-mapping)
- [Restate Documentation](https://docs.restate.dev/)
- [SurrealDB Graph Relationships](https://surrealdb.com/docs/surrealql/statements/relate)
- [Go Rate Limiting Patterns](https://golang.org/x/time/rate)

## Support

For issues or questions:
- GitHub Issues: [project repository]
- Discord: Restate community
- Documentation: `/docs/` directory
