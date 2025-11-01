# Quick Start: Host Query API

## Fast Track Testing (5 minutes)

### 1. Start SurrealDB
```bash
docker run -d --name surrealdb -p 8000:8000 \
  surrealdb/surrealdb:latest \
  start --log info --user root --pass root memory
```

### 2. Load Test Data
```bash
# Connect to SurrealDB
docker exec -it surrealdb /surreal sql \
  --conn http://localhost:8000 --user root --pass root --ns spectra --db intel

# Create a test host
CREATE host:test_host CONTENT {
    ip: "192.168.1.1",
    asn: 15169,
    city: "San Francisco",
    country: "US",
    first_seen: time::now(),
    last_seen: time::now()
};
```

### 3. Start API Server
```bash
cd cmd/api
go run main.go
```

### 4. Test the Endpoint
```bash
# Query the test host
curl http://localhost:3000/v1/query/host/192.168.1.1

# Expected response:
# {
#   "ip": "192.168.1.1",
#   "asn": 15169,
#   "city": "San Francisco",
#   "country": "US",
#   "first_seen": "2025-11-01T...",
#   "last_seen": "2025-11-01T...",
#   "ports": [],
#   "services": []
# }
```

---

## API Quick Reference

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/query/host/{ip}` | Query host by IP |

### Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| ip | string (path) | Yes | - | IP address to query |
| depth | int (query) | No | 2 | Graph traversal depth (0-5) |

### Depth Levels

```
0: Host only
1: Host + Ports
2: Host + Ports + Services [DEFAULT]
3: Host + Ports + Services + Vulnerabilities
4-5: Extended (geo, ASN)
```

### Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Invalid parameters |
| 404 | Host not found |
| 500 | Server error |

---

## Test Examples

### All Depth Levels
```bash
# Depth 0: Host only
curl "http://localhost:3000/v1/query/host/192.168.1.1?depth=0"

# Depth 1: + Ports
curl "http://localhost:3000/v1/query/host/192.168.1.1?depth=1"

# Depth 2: + Services (default)
curl "http://localhost:3000/v1/query/host/192.168.1.1"
curl "http://localhost:3000/v1/query/host/192.168.1.1?depth=2"

# Depth 3: + Vulnerabilities
curl "http://localhost:3000/v1/query/host/192.168.1.1?depth=3"
```

### Error Cases
```bash
# 404: Host not found
curl "http://localhost:3000/v1/query/host/1.2.3.4"

# 400: Invalid depth
curl "http://localhost:3000/v1/query/host/192.168.1.1?depth=10"

# 400: Invalid depth format
curl "http://localhost:3000/v1/query/host/192.168.1.1?depth=abc"
```

---

## Development Workflow

### 1. Make Changes
Edit files in:
- `internal/models/query.go` - Data models
- `internal/db/queries.go` - Database queries
- `internal/api/handlers/query.go` - HTTP handlers

### 2. Run Tests
```bash
# Unit tests
go test -v ./internal/api/handlers -run TestQueryHandler
go test -v ./internal/db -run "Test.*Query"

# Integration tests (requires SurrealDB)
go test -tags=integration -v ./internal/api/handlers

# All tests
go test ./...
```

### 3. Format Code
```bash
gofmt -w internal/models/query.go \
  internal/db/queries.go \
  internal/api/handlers/query.go
```

### 4. Build
```bash
go build ./cmd/api
```

---

## Debugging

### Enable Debug Logging
Change in `cmd/api/main.go`:
```go
// Replace
logger, err := zap.NewProduction()

// With
logger, err := zap.NewDevelopment()
```

### Check SurrealDB Connection
```bash
# Test connection
curl http://localhost:8000/health

# Check data
docker exec -it surrealdb /surreal sql \
  --conn http://localhost:8000 --user root --pass root \
  --ns spectra --db intel << EOF
SELECT * FROM host;
EOF
```

### Common Issues

**Error: "database connection error"**
- Check SurrealDB is running: `docker ps | grep surrealdb`
- Check port 8000 is accessible: `curl http://localhost:8000/health`
- Verify credentials in `handlers/query.go`

**Error: "host not found"**
- Verify host exists: `SELECT * FROM host WHERE ip = "YOUR_IP"`
- Check namespace/database: `spectra.intel`
- Ensure test data was inserted

---

## Performance Tips

1. **Use appropriate depth**
   - Depth 0-1: Fastest (< 100ms)
   - Depth 2: Standard (< 600ms)
   - Depth 3+: Slower (< 1000ms)

2. **Cache responses**
   - Consider Redis for frequently queried hosts
   - Set reasonable TTL (5-10 minutes)

3. **Monitor query latency**
   - Check logs for slow queries
   - Watch for depth=3+ queries

---

## Files Reference

```
Spectra-Red Project Structure
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── query.go                    ← HTTP handler
│   │   │   ├── query_test.go               ← Unit tests
│   │   │   └── query_integration_test.go   ← Integration tests
│   │   └── routes.go                       ← Route registration
│   ├── db/
│   │   ├── queries.go                      ← Database queries
│   │   └── queries_test.go                 ← DB tests
│   └── models/
│       └── query.go                        ← Data models
├── docs/
│   └── API_QUERY_ENDPOINT.md              ← Full documentation
└── M3-T1_IMPLEMENTATION_SUMMARY.md         ← Implementation details
```

---

## Next Steps

After testing the query endpoint:

1. **Integrate with CLI** (M4)
   - Add `spectra query host <ip>` command
   - Support depth flag: `--depth 3`

2. **Add more query endpoints** (M3-T2)
   - Query by city
   - Query by service
   - Query by vulnerability

3. **Add authentication** (M7)
   - JWT tokens
   - API keys

---

## Support

- **Documentation:** `docs/API_QUERY_ENDPOINT.md`
- **Implementation Details:** `M3-T1_IMPLEMENTATION_SUMMARY.md`
- **Tests:** Run with `go test -v ./internal/api/handlers`

---

**Quick Start Version:** 1.0
**Last Updated:** November 1, 2025
