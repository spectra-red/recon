# Spectra-Red Intel Mesh MVP - API Context Index

**Status**: COMPLETE  
**Date**: 2025-11-01  
**Agent**: API Context Gathering Agent  

## Quick Navigation

### Primary Documentation
- **[API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)** - Complete 2,194-line API specification (main deliverable)

### Reference Documents
- **Source PRD**: `/Users/seanknowles/Library/Application Support/com.conductor.app/uploads/originals/1e408e9c-b3f7-4149-9238-c3013f9b2d8b.txt`
- **Restate Architecture**: `/Users/seanknowles/Projects/recon/docs/learning/restate-deployment-architecture.md`

---

## API Surface at a Glance

### REST Endpoints (9 total)

#### Ingest (2)
- `POST /v0/mesh/ingest` - Direct cache write (60 req/min, 2s SLO)
- `POST /v0/ingest` - Workflow path (shared 60 req/min bucket)

#### Planning (1)
- `POST /v0/plan` - Stale target generation (30 req/min, 700ms SLO)

#### Coverage (1)
- `GET /v0/coverage` - Freshness statistics (60 req/min, 600ms SLO)

#### Query (2)
- `GET /v0/host/{ip}` - Single host graph (60 req/min, 600ms SLO)
- `GET /v0/search` - Selector-based search (60 req/min, 600ms SLO)

#### AI - Pro Only (1)
- `POST /v0/ai/summary` - Hybrid RAG + LLM (20 req/min, 4s SLO, 503 fallback)

#### Admin (2)
- `POST /v0/keys/rotate` - Key rotation (5 req/min, 30-day grace)
- `POST /v0/keys/revoke` - Key revocation (5 req/min, immediate)

### Restate Services & Workflows (8 + 1 separate app)

#### Main App (1 service + 4 workflows)
1. **svc.mesh_ingest** (Service) - Synchronous envelope validation + DB write
2. **wf.plan_scan** (Workflow) - Stale target computation
3. **wf.scan** (Workflow) - Probe execution + normalization
4. **wf.enrich** (Workflow) - Geo/ASN/CPE/vuln enrichment
5. **wf.graph** (Workflow) - SurrealDB node/edge upsert

#### Separate App: spectra-vuln-ingest
6. **wf.vuln_sync** (Service) - NVD/OSV/KEV mirror
7. **wf.vuln_vectorize** (Workflow) - Vector embedding

#### Pro-Only
8. **wf.ai_engine** (Workflow) - Hybrid retrieval + LLM summary

---

## Key Features

### Authentication
- OAuth2 PKCE flow (CLI-friendly)
- JWT tokens (RS256, 1-hour lifetime)
- Ed25519 envelope signing (per-ingest)
- Scope-based access control (mesh.read/write, ai.use, admin.manage)

### Authorization
- Tier-based gating (OSS/Pro/Enterprise)
- 402 Payment Required for Pro features with OSS key
- Do-Not-Scan list enforcement
- Contributor key management (rotate/revoke)

### Rate Limiting
- Per-endpoint, per-contributor limits
- X-RateLimit-* response headers
- 429 with Retry-After when exceeded
- Strict limits on key management (5 req/min)

### Error Handling
- Standard error schema with error_code + message
- Multi-status (207) for partial ingest failures
- Validation errors with per-field details
- 503 graceful fallback for AI unavailability

### Data Model
- 13 core node types (host, port, service, banner, tls_cert, vuln, vuln_doc, etc.)
- 6 geography/taxonomy types (city, region, country, asn, cloud_region, common_port)
- 12+ edge types (topology, geography, taxonomy, history)
- Vector index on vuln_doc.embedding (cosine ANN, Pro)
- Service-specific TTLs (HTTP 6h, SSH 24h, RDP 12h)

---

## SLOs & Performance Targets

| Operation | P95 Target | Notes |
|-----------|-----------|-------|
| Ingest→visible | 2 seconds | Critical user-facing SLA |
| Plan (city-level) | 700ms | Rescan planning efficiency |
| Query search | 600ms | With ≤5k results; paginate for more |
| Vector k-NN (k=20) | 250ms | Pro: requires ANN index |
| AI summary | 4 seconds | Pro: 503 fallback if timeout |
| Coverage aggregation | 600ms | Computed on-the-fly (v1: materialize) |
| Overall availability | 99.5% | Single Restate node |

---

## Architecture Highlights

### Design Principles
1. **Single Restate Node** - Centralized orchestration, simplified operations
2. **Separate Deployables** - Handler apps scale independently (Lambda/Cloud Run/K8s)
3. **Central SurrealDB** - Single source of truth for mesh graph
4. **Fire-and-Forget Async** - Loose coupling (enrich happens in background)
5. **Idempotent Workflows** - Safe retries on failures

### Data Flow
```
Community Runner (with CLI + private key)
    │
    ├─ Sign envelope (Ed25519)
    ├─ OAuth2 PKCE login
    │
    └─ POST /v0/mesh/ingest + JWT
         │
         v
    svc.mesh_ingest (Restate Service)
         │
         ├─ Verify JWT scope (mesh.write)
         ├─ Verify Ed25519 signature
         ├─ Normalize/validate
         │
         └─ Upsert SurrealDB (SYNC)
            ├─ host, port, service, banner, tls_cert
            └─ OBSERVED_AT metadata
         │
         ├─ Return 201 Created
         │
         └─ Queue async workflows (fire-and-forget)
            ├─ wf.enrich → geo/ASN/CPE/vulns
            └─ wf.graph → upsert edges + timestamps
         │
         └─ SLA: visible in ≤2s P95
         │
    User queries:
    GET /v0/search?city=Paris&service=http
         │
         v
    SurrealDB graph query
         │
         └─ Returns: (IP, port, service, product, version, geo, ...)
```

---

## Implementation Map

### Phase 1: Core REST API
- [ ] Implement 9 REST endpoints (Go HTTP handlers)
- [ ] Integrate OAuth2 PKCE + JWT validation
- [ ] Ed25519 signature verification
- [ ] Rate limiting middleware
- [ ] Error response formatting

### Phase 2: Restate Integration
- [ ] Implement svc.mesh_ingest (Service)
- [ ] Implement wf.plan_scan (Workflow)
- [ ] Implement wf.scan (Workflow)
- [ ] Implement wf.enrich (Workflow)
- [ ] Implement wf.graph (Workflow)
- [ ] Register with Restate node via HTTP/2

### Phase 3: Database
- [ ] Create SurrealDB schema (all 13 node types)
- [ ] Define edge relationships (12+ types)
- [ ] Create indices (host.ip, port.number, service.name, etc.)
- [ ] Setup vector index (vuln_doc.embedding, cosine ANN)
- [ ] Configure TTLs per service

### Phase 4: Pro Features
- [ ] Implement wf.ai_engine (Workflow)
- [ ] Implement wf.vuln_sync (separate app)
- [ ] Implement wf.vuln_vectorize (Workflow)
- [ ] Setup vector embedding pipeline
- [ ] LLM integration (GLM/Haiku)
- [ ] 503 fallback to graph-only

### Phase 5: Testing & SLO Validation
- [ ] E2E: Ingest→visible ≤2s (P95)
- [ ] E2E: Plan ≤700ms (P95, city-level)
- [ ] E2E: Query ≤600ms (P95, ≤5k results)
- [ ] Load: Vector k-NN ≤250ms (P95)
- [ ] Load: AI summary ≤4s (P95)
- [ ] Integration: Pro feature 402 gating
- [ ] Security: Ed25519 signature validation
- [ ] Rate limiting: 429 responses + headers

---

## API Request/Response Examples

### Ingest Example
```bash
curl -X POST https://api.spectra.red/v0/mesh/ingest \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "envelope": {
      "contributor_id": "550e8400-e29b-41d4-a716-446655440000",
      "scan_id": "650e8400-e29b-41d4-a716-446655440001",
      "signature": "cafe...",
      "timestamp": "2025-11-01T12:00:00Z",
      "privacy_mode": "full-banner"
    },
    "assets": [{
      "host": {"ip": "192.0.2.1", "asn": 65001, "geo": {"city": "Paris", "country": "FR"}},
      "ports": [
        {
          "number": 80,
          "protocol": "tcp",
          "state": "open",
          "service": {"name": "http", "product": "Apache", "version": "2.4.41"},
          "banner": {"hash": "abc123...", "sample": "Apache/2.4.41"}
        }
      ]
    }]
  }'

# Response (201 Created)
{
  "status": "created",
  "scan_id": "650e8400-e29b-41d4-a716-446655440001",
  "assets_ingested": 1,
  "assets_ignored": 0,
  "messages": [],
  "visible_in_secs": 1
}
```

### Plan Example
```bash
curl -X POST https://api.spectra.red/v0/plan \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "selectors": {"city": ["Paris"], "service": ["redis"]},
    "min_age": "5m",
    "pagination": {"limit": 10000}
  }'

# Response (200 OK)
{
  "status": "success",
  "plan_id": "550e8400-e29b-41d4-a716-446655440000",
  "targets": [
    {"ip": "192.0.2.1", "port": 6379, "service": "redis", "last_seen": "2025-11-01T11:55:00Z", "age_secs": 300},
    {"ip": "192.0.2.2", "port": 6379, "service": "redis", "last_seen": "2025-11-01T11:54:00Z", "age_secs": 360}
  ],
  "target_count": 2,
  "stats": {"total_matching": 10, "stale_only": 2, "estimated_probe_time_minutes": 1}
}
```

### AI Summary Example (Pro)
```bash
curl -X POST https://api.spectra.red/v0/ai/summary \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "selector",
    "selector": {"city": ["Paris"], "service": ["http"], "min_severity": "high"},
    "k": 20,
    "style": "bullets"
  }'

# Response (200 OK)
{
  "status": "success",
  "summary": {
    "title": "Critical Vulnerabilities in Paris HTTP Services",
    "overview": "3 critical RCE CVEs affecting 45% of hosts running Apache",
    "critical_vulns": [
      {"cve_id": "CVE-2024-1234", "cvss": 9.8, "kev_listed": true}
    ],
    "top_risks": [
      "Unpatched Apache with RCE (CVE-2024-1234)",
      "Exposed HTTP admin panels",
      "Weak SSL/TLS configurations"
    ]
  }
}
```

---

## Authentication Flow Diagram

```
┌─────────────────────────┐
│  User/CLI Login Request │
└───────────┬─────────────┘
            │
            v
    ┌───────────────────────────────────┐
    │  CLI generates code_verifier      │
    │  code_challenge = base64url(sha256│
    │  Opens browser to /authorize      │
    └───────────┬───────────────────────┘
                │
                v
        ┌──────────────────────┐
        │ User approves scope  │
        │ (mesh.read/write/ai) │
        └──────────┬───────────┘
                   │
                   v
        ┌────────────────────────────┐
        │  Authorization code → CLI  │
        │  Redirects to localhost:99 │
        └──────────┬─────────────────┘
                   │
                   v
    ┌──────────────────────────────────┐
    │  CLI exchanges code + code_verifier
    │  POST /token (PKCE flow)          │
    └──────────┬────────────────────────┘
               │
               v
        ┌──────────────────────┐
        │ Returns JWT + Refresh │
        │ CLI stores in keyring │
        └──────────┬───────────┘
                   │
                   v
      ┌────────────────────────────┐
      │  JWT included in API calls │
      │  Authorization: Bearer ... │
      └────────────────────────────┘
```

---

## Error Response Reference

### 400 Bad Request (Validation)
```json
{
  "status": "error",
  "error_code": "invalid_ip",
  "message": "Invalid IP address",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "details": {"field": "assets[0].host.ip", "reason": "192.0.2.999 is not valid"}
}
```

### 401 Unauthorized (Auth)
```json
{
  "status": "error",
  "error_code": "unauthorized",
  "message": "Missing or invalid token",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 402 Payment Required (Pro Gate)
```json
{
  "status": "error",
  "error_code": "payment_required",
  "message": "Pro feature requires upgrade",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 429 Rate Limit Exceeded
```json
{
  "status": "error",
  "error_code": "rate_limit_exceeded",
  "message": "60 requests per minute limit exceeded",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 207 Multi-Status (Partial Ingest)
```json
{
  "status": "multi-status",
  "scan_id": "650e8400-e29b-41d4-a716-446655440001",
  "assets_ingested": 8,
  "assets_ignored": 2,
  "errors": [
    {"asset_index": 0, "error_code": "invalid_ip", "message": "Invalid IP"},
    {"asset_index": 5, "error_code": "invalid_port", "message": "Port out of range"}
  ]
}
```

---

## Deployment Checklist

### Infrastructure
- [ ] Restate server (high availability, 99.5% target)
- [ ] SurrealDB (with vector index support)
- [ ] Auth server (OAuth2 PKCE issuer)
- [ ] Handler app deployment (Lambda/Cloud Run/K8s)
- [ ] AI model (GLM/Haiku or equivalent)
- [ ] Key management (KMS for encryption at rest)

### Operations
- [ ] Monitoring: P95 latency per endpoint
- [ ] Monitoring: Rate limit bucket usage
- [ ] Logging: All auth events (signed/unsigned)
- [ ] Logging: All Pro feature access (402 responses)
- [ ] Audit trail: Key rotation/revocation events
- [ ] Backup: SurrealDB (durable graph data)

### Security
- [ ] OAuth2 PKCE enforcement
- [ ] Ed25519 signature verification
- [ ] JWT expiration validation
- [ ] Scope-based access control
- [ ] IP/ASN blocklist enforcement
- [ ] Do-Not-Scan list checking
- [ ] KMS encryption for keys at rest

---

## Testing Strategy

### Unit Tests
- Ed25519 signature verification (valid/invalid)
- JWT parsing and scope validation
- CPE format validation
- IP/port range checks
- Rate limit counter logic

### Integration Tests
- Full ingest→visible flow (2s SLO)
- Plan generation with selectors
- Graph traversal queries
- Key rotation with grace period
- Key revocation (immediate)

### Load Tests
- P95 latency targets (all endpoints)
- Concurrent ingest volume
- Vector k-NN performance (250ms)
- Rate limiting under load

### E2E Tests
- OAuth2 PKCE flow (CLI login)
- Ingest with signed envelope
- Plan→rescan pipeline
- Pro features with 402 gating
- Relay rotation on ISP block

---

## Documentation Structure

```
API_CONTEXT_CATALOG.md (2,194 lines)
├── Executive Summary
├── 1. REST API Endpoints Catalog
│   ├── 1.1 Ingest APIs
│   ├── 1.2 Planning APIs
│   ├── 1.3 Coverage APIs
│   ├── 1.4 Query APIs
│   ├── 1.5 AI APIs (Pro)
│   └── 1.6 Admin APIs
├── 2. Restate Service & Workflow Interfaces
│   ├── 2.1 svc.mesh_ingest
│   ├── 2.2 wf.plan_scan
│   ├── 2.3 wf.scan
│   ├── 2.4 wf.enrich
│   ├── 2.5 wf.graph
│   ├── 2.6 wf.ai_engine
│   ├── 2.7 wf.vuln_sync
│   └── 2.8 wf.vuln_vectorize
├── 3. Authentication & Authorization
│   ├── 3.1 OAuth2 PKCE Flow
│   ├── 3.2 JWT Token Structure
│   ├── 3.3 Envelope Signing (Ed25519)
│   └── 3.4 Scoped Access Control
├── 4. Rate Limiting Policies
├── 5. Error Response Formats
├── 6. Data Model (SurrealDB)
├── 7. Example API Flows
├── 8. Implementation Checklist
└── 9. References & Resources
```

---

## Quick Links

### API Reference
- All endpoints: See Section 1 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)
- Authentication: See Section 3 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)
- Rate limiting: See Section 4 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)
- Errors: See Section 5 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)

### Implementation
- Services: Section 2.1 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)
- Workflows: Sections 2.2-2.8 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)
- Data model: Section 6 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)
- Checklist: Section 8 of [API_CONTEXT_CATALOG.md](./API_CONTEXT_CATALOG.md)

---

## Status Summary

COMPLETE - All API context documented:
- 9 REST endpoints specified
- 8 Restate services/workflows defined
- Authentication flows documented
- Rate limiting policies defined
- Error formats standardized
- SurrealDB data model complete
- Example flows provided

**Ready for**: Implementation Planning → Design Review → Development Kickoff

---

**Document**: API_CONTEXT_INDEX.md  
**Version**: 1.0  
**Generated**: 2025-11-01  
**Agent**: API Context Gathering Agent
