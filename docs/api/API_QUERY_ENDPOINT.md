# Host Query API Endpoint

## Overview

The Host Query API endpoint provides a way to retrieve detailed information about hosts in the Spectra-Red Intel Mesh, including their ports, services, and vulnerabilities.

**Endpoint:** `GET /v1/query/host/{ip}`

**Implemented in:** Milestone 3, Task 1 (M3-T1)

## Features

- ✅ Query host by IP address
- ✅ Configurable graph traversal depth (0-5)
- ✅ Returns related ports, services, and vulnerabilities
- ✅ Includes last_seen timestamp
- ✅ Returns 404 for unknown hosts
- ✅ Full error handling and validation
- ✅ Structured logging with request IDs
- ✅ Comprehensive test coverage

## API Specification

### Request

```
GET /v1/query/host/{ip}?depth={0-5}
```

**Path Parameters:**
- `ip` (required): IP address to query (e.g., `1.2.3.4`)

**Query Parameters:**
- `depth` (optional): Graph traversal depth (default: 2)
  - `0`: Host only
  - `1`: Host + Ports
  - `2`: Host + Ports + Services (default)
  - `3`: Host + Ports + Services + Vulnerabilities
  - `4-5`: Extended relationships (geographic, ASN)

### Response

#### Success Response (200 OK)

```json
{
  "ip": "1.2.3.4",
  "asn": 15169,
  "city": "San Francisco",
  "region": "California",
  "country": "US",
  "cloud_region": "us-west1",
  "first_seen": "2025-10-15T10:30:00Z",
  "last_seen": "2025-11-01T14:22:33Z",
  "ports": [
    {
      "number": 80,
      "protocol": "tcp",
      "transport": "plain",
      "first_seen": "2025-10-15T10:30:00Z",
      "last_seen": "2025-11-01T14:22:33Z",
      "services": []
    },
    {
      "number": 443,
      "protocol": "tcp",
      "transport": "tls",
      "first_seen": "2025-10-15T10:30:00Z",
      "last_seen": "2025-11-01T14:22:33Z",
      "services": []
    }
  ],
  "services": [
    {
      "name": "http",
      "product": "nginx",
      "version": "1.25.1",
      "cpe": ["cpe:2.3:a:nginx:nginx:1.25.1:*:*:*:*:*:*:*"],
      "confidence": 1.0,
      "first_seen": "2025-10-15T10:30:00Z",
      "last_seen": "2025-11-01T14:22:33Z",
      "vulnerabilities": []
    }
  ],
  "vulnerabilities": [
    {
      "cve_id": "CVE-2023-12345",
      "cvss": 9.8,
      "severity": "critical",
      "kev_flag": true,
      "confidence": 0.95,
      "first_detected": "2025-10-16T08:00:00Z"
    }
  ]
}
```

#### Error Responses

**404 Not Found** - Host does not exist:
```json
{
  "error": "Not Found",
  "message": "host not found",
  "code": 404
}
```

**400 Bad Request** - Invalid parameters:
```json
{
  "error": "Bad Request",
  "message": "depth must be between 0 and 5",
  "code": 400
}
```

**500 Internal Server Error** - Database or server error:
```json
{
  "error": "Internal Server Error",
  "message": "database connection error",
  "code": 500
}
```

## Usage Examples

### cURL

```bash
# Query host with default depth (2)
curl http://localhost:3000/v1/query/host/1.2.3.4

# Query host with depth 0 (host only)
curl http://localhost:3000/v1/query/host/1.2.3.4?depth=0

# Query host with depth 3 (include vulnerabilities)
curl http://localhost:3000/v1/query/host/1.2.3.4?depth=3

# Query host with maximum depth
curl http://localhost:3000/v1/query/host/1.2.3.4?depth=5
```

### Go Client Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type HostQueryResponse struct {
    IP        string    `json:"ip"`
    ASN       int       `json:"asn,omitempty"`
    City      string    `json:"city,omitempty"`
    LastSeen  string    `json:"last_seen"`
    Ports     []Port    `json:"ports,omitempty"`
}

func queryHost(ip string, depth int) (*HostQueryResponse, error) {
    url := fmt.Sprintf("http://localhost:3000/v1/query/host/%s?depth=%d", ip, depth)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("query failed: %s", resp.Status)
    }

    var result HostQueryResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    host, err := queryHost("1.2.3.4", 2)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Host: %s, Last Seen: %s\n", host.IP, host.LastSeen)
    fmt.Printf("Ports: %d\n", len(host.Ports))
}
```

### JavaScript/TypeScript Example

```typescript
interface HostQueryResponse {
  ip: string;
  asn?: number;
  city?: string;
  last_seen: string;
  ports?: Port[];
  services?: Service[];
  vulnerabilities?: Vulnerability[];
}

async function queryHost(ip: string, depth: number = 2): Promise<HostQueryResponse> {
  const response = await fetch(
    `http://localhost:3000/v1/query/host/${ip}?depth=${depth}`
  );

  if (!response.ok) {
    throw new Error(`Query failed: ${response.statusText}`);
  }

  return await response.json();
}

// Usage
queryHost('1.2.3.4', 3)
  .then(host => {
    console.log(`Host: ${host.ip}`);
    console.log(`Last Seen: ${host.last_seen}`);
    console.log(`Vulnerabilities: ${host.vulnerabilities?.length || 0}`);
  })
  .catch(console.error);
```

## Performance Characteristics

### Query Performance

- **Depth 0-1:** < 100ms (P95)
- **Depth 2:** < 600ms (P95) - Default, optimized for common use case
- **Depth 3-5:** < 1000ms (P95) - More expensive graph traversal

### Optimization Strategies

1. **Use appropriate depth:** Only request the data you need
2. **Caching:** Consider caching responses for frequently queried hosts
3. **Batch queries:** If querying multiple hosts, use pagination or batch endpoints (future)

## SurrealDB Query Pattern

The endpoint uses SurrealDB's graph traversal syntax to efficiently query related data:

```sql
-- Depth 2 query (default)
SELECT *,
  ->HAS->port.* AS ports,
  ->HAS->port->RUNS->service.* AS services
FROM host WHERE ip = $ip
LIMIT 1;

-- Depth 3 query (with vulnerabilities)
SELECT *,
  ->HAS->port.* AS ports,
  ->HAS->port->RUNS->service.* AS services,
  ->HAS->port->RUNS->service->AFFECTED_BY->vuln.* AS vulns
FROM host WHERE ip = $ip
LIMIT 1;
```

**Graph Traversal Path:**
```
host → HAS → port → RUNS → service → AFFECTED_BY → vuln
```

## Testing

### Unit Tests

Run unit tests:
```bash
go test -v ./internal/api/handlers -run TestQueryHandler
go test -v ./internal/db -run TestBuildHostQuery
```

### Integration Tests

Run integration tests (requires SurrealDB):
```bash
# Start SurrealDB
docker run -p 8000:8000 surrealdb/surrealdb:latest \
  start --log trace --user root --pass root memory

# Run integration tests
go test -tags=integration -v ./internal/api/handlers
```

### Performance Testing

Run benchmarks:
```bash
go test -tags=integration -bench=BenchmarkQueryHandler -benchmem \
  ./internal/api/handlers
```

## Files Created

### Core Implementation
- `internal/models/query.go` - Query request/response models
- `internal/db/queries.go` - SurrealDB query functions
- `internal/api/handlers/query.go` - HTTP handler implementation
- `internal/api/routes.go` - Route registration (updated)

### Tests
- `internal/api/handlers/query_test.go` - Unit tests
- `internal/db/queries_test.go` - Database query tests
- `internal/api/handlers/query_integration_test.go` - Integration tests

### Documentation
- `docs/API_QUERY_ENDPOINT.md` - This file

## Architecture Integration

### Components

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP GET
       ▼
┌─────────────────────────────┐
│   QueryHandler              │
│   - Parse IP & depth        │
│   - Validate parameters     │
│   - Create DB connection    │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│   db.QueryHost()            │
│   - Build SurrealDB query   │
│   - Execute graph traversal │
│   - Parse result            │
└──────┬──────────────────────┘
       │
       ▼
┌─────────────────────────────┐
│   SurrealDB                 │
│   - Graph query execution   │
│   - Return host + relations │
└─────────────────────────────┘
```

### Error Handling

The endpoint implements comprehensive error handling:

1. **Input Validation**: Missing IP, invalid depth values
2. **Database Errors**: Connection failures, query errors
3. **Not Found**: Host doesn't exist in database
4. **Server Errors**: Unexpected failures with proper logging

All errors are logged with structured logging (zap) including:
- Request ID
- IP address
- Depth parameter
- Error details

## Security Considerations

### Current Implementation
- No authentication required (MVP phase)
- No rate limiting on query endpoint
- IP validation through URL routing only

### Future Enhancements
- Add JWT authentication
- Implement rate limiting (30 req/min recommended)
- Add API key validation
- IP allowlist/blocklist support

## Future Enhancements

1. **Pagination:** Support for large result sets
2. **Field Selection:** Allow clients to specify which fields to return
3. **Aggregations:** Support for counting hosts, grouping by attributes
4. **Time Range Queries:** Query hosts seen within a time range
5. **Bulk Queries:** Query multiple IPs in a single request
6. **WebSocket Support:** Real-time updates for monitored hosts

## Acceptance Criteria Status

- ✅ GET /v1/query/host/{ip} returns host details
- ✅ Queries SurrealDB for host record
- ✅ Returns related ports, services, vulnerabilities
- ✅ Supports optional depth parameter for graph traversal
- ✅ Returns 404 for unknown hosts
- ✅ Response includes last_seen timestamp
- ✅ Unit and integration tests

All acceptance criteria from M3-T1 have been successfully implemented and tested.

## References

- [SurrealDB Documentation](https://surrealdb.com/docs)
- [Chi Router Documentation](https://github.com/go-chi/chi)
- [DETAILED_IMPLEMENTATION_PLAN.md](../DETAILED_IMPLEMENTATION_PLAN.md) - Milestone 3, Task 1
