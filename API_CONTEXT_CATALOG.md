# Spectra-Red Intel Mesh MVP - API Context & Service Documentation

**Generated**: 2025-11-01  
**Project**: Spectra-Red Intel Mesh MVP PRD  
**Source PRD**: /Users/seanknowles/Library/Application Support/com.conductor.app/uploads/originals/1e408e9c-b3f7-4149-9238-c3013f9b2d8b.txt

---

## Executive Summary

This document catalogs the complete API surface for the Spectra-Red Intel Mesh MVP, including:
- **REST API endpoints** for ingest, planning, querying, and AI operations
- **Restate service and workflow interfaces** for durable orchestration
- **Authentication & authorization** patterns (OAuth2 PKCE + JWT scopes)
- **Rate limiting policies** per endpoint tier
- **Error response formats** and SLO targets
- **Data models** underlying all API operations

**Key Architecture**: Single Restate node orchestrates separately deployable Go services, with central SurrealDB for graph storage. Community runners submit scan data via TLS, visible in near-real-time.

---

## 1. REST API Endpoints Catalog

### 1.1 Ingest APIs

#### POST /v0/mesh/ingest (Direct Cache Path)

**Purpose**: Fast remote-cache write path with minimal validation. Directly upserts normalized host/port/service/banner/TLS data.

**Authentication**:
- Required: Yes
- Method: OAuth2 (PKCE) JWT + Ed25519 signed envelope
- Scopes: `mesh.write`

**Rate Limiting**:
- Limit: 60 req/min
- Headers: `X-RateLimit-Remaining`, `X-RateLimit-Reset`
- Abuse control: IP/ASN blocklist enforced

**Request Schema**:
```json
{
  "envelope": {
    "contributor_id": "string (UUID)",
    "scan_id": "string (UUID)",
    "signature": "string (Ed25519 hex)",
    "timestamp": "string (ISO8601)",
    "privacy_mode": "meta-only|full-banner"
  },
  "assets": [
    {
      "host": {
        "ip": "string (IPv4/IPv6)",
        "asn": "number (optional)",
        "geo": {
          "city": "string (optional)",
          "region": "string (optional)",
          "country": "string (ISO 3166-1 alpha-2)",
          "cloud_region": "string (optional, e.g., 'us-east-1')"
        }
      },
      "ports": [
        {
          "number": "number (1-65535)",
          "protocol": "tcp|udp",
          "state": "open|closed|filtered",
          "service": {
            "name": "string (http, ssh, rdp, ...)",
            "product": "string (optional, e.g., 'Apache')",
            "version": "string (optional)",
            "cpe": ["string (optional, CPE v2.3 URIs)"]
          },
          "banner": {
            "hash": "string (SHA256 hex, optional)",
            "sample": "string (≤2KB, optional)",
            "fingerprint": "string (optional, normalized)"
          },
          "tls_cert": {
            "sha256": "string (hex, optional)",
            "cn": "string (common name, optional)",
            "sans": ["string (SANs, optional)"],
            "not_before": "string (ISO8601, optional)",
            "not_after": "string (ISO8601, optional)"
          }
        }
      ]
    }
  ]
}
```

**Response (201 Created)**:
```json
{
  "status": "created",
  "scan_id": "string (UUID)",
  "assets_ingested": "number",
  "assets_ignored": "number",
  "messages": ["string (diagnostics)"],
  "visible_in_secs": "number"
}
```

**Response (207 Multi-Status)** - Partial failures:
```json
{
  "status": "multi-status",
  "scan_id": "string (UUID)",
  "assets_ingested": "number",
  "assets_ignored": "number",
  "errors": [
    {
      "asset_index": "number",
      "error_code": "invalid_ip|invalid_port|invalid_cpe|...",
      "message": "string"
    }
  ]
}
```

**Status Codes**:
- `201` - Ingest successful, async enrich + graph queued
- `207` - Partial success, see errors array
- `400` - Invalid envelope signature or malformed data
- `401` - Unauthorized (invalid JWT or missing scope)
- `402` - Payment Required (Pro features with OSS key)
- `429` - Rate limit exceeded (60 req/min)
- `500` - Server error

**Headers**:
- `Authorization: Bearer <jwt_token>` (Required)
- `Content-Type: application/json` (Required)
- `X-Signature-Algorithm: ed25519` (Optional, defaults to ed25519)

**SLO**: P95 ingest→visible ≤ **2 seconds** for normalized records

**Side Effects**:
1. Synchronous: Normalize & validate, upsert to SurrealDB `host`, `port`, `service`, `banner`, `tls_cert` with `OBSERVED_AT`
2. Asynchronous (fire-and-forget):
   - `wf.enrich(asset_set_ref)` - Add geo/ASN/CPE/TLS parsing
   - `wf.graph(enriched_set_ref)` - Upsert nodes/edges, update `last_seen`
   - `wf.ai_engine` - (Pro only) Vector RAG on vulns

**Example cURL**:
```bash
curl -X POST https://api.spectra.red/v0/mesh/ingest \
  -H "Authorization: Bearer eyJ..." \
  -H "Content-Type: application/json" \
  -d @- << 'PAYLOAD'
{
  "envelope": {
    "contributor_id": "550e8400-e29b-41d4-a716-446655440000",
    "scan_id": "650e8400-e29b-41d4-a716-446655440001",
    "signature": "cafe...",
    "timestamp": "2025-11-01T12:00:00Z",
    "privacy_mode": "full-banner"
  },
  "assets": [
    {
      "host": {"ip": "192.0.2.1", "asn": 65001, "geo": {"city": "Paris", "country": "FR"}},
      "ports": [
        {
          "number": 80,
          "protocol": "tcp",
          "state": "open",
          "service": {"name": "http", "product": "Apache", "version": "2.4.41", "cpe": ["cpe:2.3:a:apache:http_server:2.4.41:*:*:*:*:*:*:*"]},
          "banner": {"hash": "abc123...", "sample": "Apache/2.4.41"},
          "tls_cert": null
        },
        {
          "number": 22,
          "protocol": "tcp",
          "state": "open",
          "service": {"name": "ssh", "product": "OpenSSH", "version": "7.4p1"},
          "banner": null,
          "tls_cert": null
        }
      ]
    }
  ]
}
PAYLOAD
```

---

#### POST /v0/ingest (Workflow Path)

**Purpose**: Submit raw scan envelope to full enrichment workflow (plan → scan → enrich → graph).

**Authentication**:
- Required: Yes (same as `/v0/mesh/ingest`)
- Scopes: `mesh.write`

**Rate Limiting**:
- Limit: 60 req/min (same bucket as `/v0/mesh/ingest`)

**Request Schema**: Same as `/v0/mesh/ingest`

**Response (202 Accepted)**:
```json
{
  "status": "accepted",
  "workflow_id": "string (UUID for tracking)",
  "estimated_completion_secs": "number"
}
```

**Status Codes**:
- `202` - Accepted, workflow starting
- `400` - Invalid request
- `401` - Unauthorized
- `429` - Rate limit exceeded

**SLO**: Same as direct path (2s P95 for visible)

**Side Effects**:
- Queues: `wf.scan()` → `wf.enrich()` → `wf.graph()` → (Pro) `wf.ai_engine()`
- Returns workflow ID for polling with `GET /v0/workflow/{workflow_id}/status` (future)

---

### 1.2 Planning APIs

#### POST /v0/plan (Scan Planning)

**Purpose**: Generate stale-only target list for rescan based on age threshold and selectors.

**Authentication**:
- Required: Yes
- Scopes: `mesh.read`

**Rate Limiting**:
- Limit: 30 req/min
- Headers: `X-RateLimit-Remaining`

**Request Schema**:
```json
{
  "selectors": {
    "city": ["Paris", "London"],              // Optional
    "region": ["Île-de-France"],              // Optional (ISO 3166-2 code)
    "country": ["FR", "GB"],                  // Optional (ISO 3166-1 alpha-2)
    "asn": [65001, 65002],                    // Optional
    "cidr": ["192.0.2.0/24"],                 // Optional
    "cloud_region": ["us-east-1", "eu-west-1"], // Optional
    "service": ["http", "redis", "mongodb"],  // Optional
    "common_port": [80, 443, 6379],           // Optional
    "since": "string (ISO8601 or duration, optional, e.g., '7d')"
  },
  "min_age": "string (duration, e.g., '5m', '1h', '7d')",
  "exclude": {
    "ip_ranges": ["string (CIDR)"],           // Optional: Do-Not-Scan list
    "ports": ["number"]                       // Optional
  },
  "pagination": {
    "limit": "number (default 10000, max 50000)",
    "cursor": "string (optional, for continuation)"
  }
}
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "plan_id": "string (UUID, references this plan execution)",
  "query_params": {
    "selector_count": "number",
    "filter_description": "string"
  },
  "targets": [
    {
      "ip": "string (IPv4/IPv6)",
      "port": "number",
      "service": "string (optional)",
      "last_seen": "string (ISO8601)",
      "age_secs": "number",
      "last_scanned_at": "string (ISO8601, optional)"
    }
  ],
  "target_count": "number",
  "pagination": {
    "has_more": "boolean",
    "next_cursor": "string (if has_more)"
  },
  "stats": {
    "total_matching": "number",
    "stale_only": "number",
    "estimated_probe_time_minutes": "number"
  }
}
```

**Status Codes**:
- `200` - Planning succeeded
- `400` - Invalid selectors or min_age format
- `401` - Unauthorized
- `429` - Rate limit exceeded

**SLO**: P95 ≤ **700 ms** for city-level selectors

**Business Logic**:
- Computes `age = now - last_seen` for each `(host, port)` pair
- Returns only targets where `age >= min_age`
- Applies exclusions from Do-Not-Scan list (stored on contributor profile)
- Paginates results if >10k targets

**Example cURL**:
```bash
curl -X POST https://api.spectra.red/v0/plan \
  -H "Authorization: Bearer eyJ..." \
  -H "Content-Type: application/json" \
  -d '{
    "selectors": {
      "city": ["Paris"],
      "service": ["redis"]
    },
    "min_age": "5m",
    "pagination": {"limit": 10000}
  }'
```

---

### 1.3 Coverage APIs

#### GET /v0/coverage (Freshness Statistics)

**Purpose**: Query aggregate freshness/coverage statistics without returning individual targets.

**Authentication**:
- Required: Yes
- Scopes: `mesh.read`

**Rate Limiting**:
- Limit: 60 req/min
- Headers: `X-RateLimit-Remaining`

**Query Parameters**:
```
?city=Paris&region=Île-de-France&country=FR&asn=65001&service=http&common_port=80&since=7d
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "selectors": {
    "city": "string (description)",
    "count": "number"
  },
  "coverage": {
    "total_assets": "number (unique host/port combos)",
    "scanned_in_last_24h": "number",
    "scanned_in_last_7d": "number",
    "scanned_in_last_30d": "number"
  },
  "freshness": {
    "p50_age_secs": "number (median age)",
    "p95_age_secs": "number",
    "p99_age_secs": "number",
    "oldest_age_secs": "number",
    "newest_age_secs": "number"
  },
  "by_service": [
    {
      "service": "string (http, ssh, ...)",
      "count": "number",
      "p50_age_secs": "number",
      "p95_age_secs": "number"
    }
  ],
  "by_region": [
    {
      "country": "string (ISO 3166-1 alpha-2)",
      "region": "string (optional)",
      "count": "number",
      "p50_age_secs": "number"
    }
  ]
}
```

**Status Codes**:
- `200` - Success
- `400` - Invalid selector format
- `401` - Unauthorized
- `429` - Rate limit exceeded

**SLO**: P95 ≤ **600 ms** (computed on-the-fly via Surreal aggregations)

**Implementation Note**: MVP computes coverage on-demand; v1 may materialize tables if P95 drifts.

---

### 1.4 Query APIs

#### GET /v0/host/{ip} (Single Host Graph)

**Purpose**: Retrieve complete host graph (ports, services, banners, TLS certs, and Pro: vulnerabilities).

**Authentication**:
- Required: Yes
- Scopes: `mesh.read`

**Rate Limiting**:
- Limit: 60 req/min

**URL Parameters**:
```
GET /v0/host/192.0.2.1
GET /v0/host/2001:db8::1
```

**Query Parameters**:
```
?include=ports,services,banners,tls_certs,vulns&since=7d
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "host": {
    "ip": "string",
    "asn": "number",
    "asn_org": "string",
    "geo": {
      "city": "string (optional)",
      "region": "string (optional)",
      "country": "string (ISO 3166-1 alpha-2)",
      "cloud_region": "string (optional)"
    },
    "first_seen": "string (ISO8601)",
    "last_seen": "string (ISO8601)",
    "last_scanned_at": "string (ISO8601, optional)"
  },
  "ports": [
    {
      "port": "number",
      "protocol": "tcp|udp",
      "is_common": "boolean",
      "services": [
        {
          "name": "string",
          "product": "string (optional)",
          "version": "string (optional)",
          "cpe": ["string (optional)"],
          "first_seen": "string (ISO8601)",
          "last_seen": "string (ISO8601)",
          "observations_count": "number"
        }
      ],
      "banners": [
        {
          "hash": "string (SHA256)",
          "sample": "string (≤2KB, optional)",
          "observed_at": "string (ISO8601)",
          "contributors": ["string (contributor_id, optional)"]
        }
      ],
      "tls_certs": [
        {
          "sha256": "string (hex)",
          "cn": "string",
          "sans": ["string"],
          "not_before": "string (ISO8601)",
          "not_after": "string (ISO8601)",
          "issuer": "string (optional)",
          "observed_at": "string (ISO8601)"
        }
      ],
      "vulns": [  // Pro only
        {
          "cve_id": "string",
          "cvss": "number (0-10)",
          "cvss_vector": "string (optional)",
          "severity": "critical|high|medium|low",
          "kev_status": "boolean (in CISA KEV catalog)",
          "exploits": ["string (URLs, optional)"]
        }
      ]
    }
  ],
  "summary": {
    "total_ports_observed": "number",
    "open_ports": "number",
    "services_count": "number",
    "vuln_count": "number (Pro only)",
    "critical_vulns": "number (Pro only)"
  }
}
```

**Status Codes**:
- `200` - Success
- `400` - Invalid IP format
- `401` - Unauthorized
- `402` - Payment Required (if vulns requested with OSS key)
- `404` - Host not found in mesh
- `429` - Rate limit exceeded

**SLO**: P95 ≤ **600 ms**

**Business Logic**:
- Returns merged view of all observations for the host
- `since` parameter filters observations to last N days
- Vulnerabilities only included with Pro tier or higher
- TTLs per service: HTTP 6h, SSH 24h, RDP 12h

---

#### GET /v0/search (Selector Query)

**Purpose**: Complex selector-based search across the mesh (similar to `/v0/plan` but returns full graph data).

**Authentication**:
- Required: Yes
- Scopes: `mesh.read`

**Rate Limiting**:
- Limit: 60 req/min

**Query Parameters**:
```
?city=Paris&region=Île-de-France&country=FR&asn=65001&cidr=192.0.2.0/24&cloud_region=us-east-1&service=http&common_port=80&since=7d&limit=5000&cursor=...
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "query": {
    "filters": ["string (description of each filter)"],
    "limit": "number"
  },
  "results": [
    {
      "ip": "string",
      "port": "number",
      "service": "string",
      "product": "string (optional)",
      "version": "string (optional)",
      "banner_hash": "string (optional)",
      "tls_cn": "string (optional)",
      "geo": {
        "city": "string (optional)",
        "country": "string"
      },
      "last_seen": "string (ISO8601)",
      "observation_count": "number"
    }
  ],
  "result_count": "number",
  "pagination": {
    "has_more": "boolean",
    "next_cursor": "string (if has_more)",
    "total_matching": "number (estimated)"
  }
}
```

**Status Codes**:
- `200` - Success (may be empty)
- `400` - Invalid selector format
- `401` - Unauthorized
- `429` - Rate limit exceeded

**SLO**: P95 ≤ **600 ms** with ≤5k nodes/edges; larger responses require pagination

**Business Logic**:
- Returns distinct `(ip, port, service)` tuples matching all selectors
- Default limit 5000; can request up to 50k (with pagination)
- Paginated via cursor-based pagination
- Optimized via indices on `host.ip`, `port.number`, `service.name`, `city.name`, `region.name`, etc.

---

### 1.5 AI APIs (Pro-only)

#### POST /v0/ai/summary (Vulnerability Summarization)

**Purpose**: (Pro tier) Hybrid retrieval + LLM summary for vulnerabilities matching selectors or graph filters.

**Authentication**:
- Required: Yes
- Scopes: `mesh.read`, `ai.use` (Pro-gated)

**Rate Limiting**:
- Limit: 20 req/min (Pro only)
- Headers: `X-RateLimit-Remaining`

**Request Schema**:
```json
{
  "mode": "selector|graph|host",
  "selector": {
    "city": ["Paris"],               // If mode=selector
    "service": ["http"],
    "min_severity": "high",          // Optional: critical, high, medium, low
    "kev_only": "boolean (optional, default false)"
  },
  "host_ip": "string (if mode=host)",  // e.g., "192.0.2.1"
  "k": "number (default 20, top-k vulns)",
  "style": "bullets|detailed|brief"  // Optional response format
}
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "mode": "selector|graph|host",
  "summary": {
    "title": "string (e.g., 'Critical Vulnerabilities in Paris HTTP Services')",
    "overview": "string (1-2 sentence summary)",
    "critical_vulns": [
      {
        "cve_id": "string",
        "title": "string",
        "cvss": "number",
        "severity": "critical|high|...",
        "affected_count": "number (hosts with this vuln)",
        "kev_listed": "boolean"
      }
    ],
    "top_risks": [
      "string (1-3 bullet points)"
    ],
    "recommendations": [
      "string (action items)"
    ]
  },
  "retrieval_stats": {
    "vector_hits": "number",
    "graph_hits": "number",
    "processing_time_ms": "number"
  }
}
```

**Response (503 Service Unavailable)** - Fallback:
```json
{
  "status": "timeout",
  "fallback": "graph-only-summary",
  "message": "AI model unavailable; returning graph-only data",
  "results": {
    "matching_hosts": "number",
    "matching_vulns": "number"
  }
}
```

**Status Codes**:
- `200` - Success
- `400` - Invalid request format
- `401` - Unauthorized (missing `ai.use` scope)
- `402` - Payment Required (OSS key calling Pro endpoint)
- `429` - Rate limit exceeded
- `503` - AI model timeout/unavailable (returns graph-only fallback)

**SLO**: P95 ≤ **4 seconds**; fallback to graph-only if AI unavailable

**Implementation Details**:
1. **Hybrid Retrieval**:
   - Graph filter: Apply selectors → matching hosts/ports/services
   - Vector k-NN: Search `vuln_doc.embedding` (k=20 by default)
   - Combine: Filter vector results by CPE match + recency + KEV status
2. **LLM Processing**: GLM/Haiku model summarizes top results
3. **Fallback**: If AI unavailable, return graph-filtered vulnerability counts

**Example cURL**:
```bash
curl -X POST https://api.spectra.red/v0/ai/summary \
  -H "Authorization: Bearer eyJ..." \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "selector",
    "selector": {
      "city": ["Paris"],
      "service": ["http"],
      "min_severity": "high"
    },
    "k": 20,
    "style": "bullets"
  }'
```

---

### 1.6 Admin APIs

#### POST /v0/keys/rotate (Contributor Key Rotation)

**Purpose**: Rotate Ed25519 signing key for envelope signatures. Generates new key, marks old as deprecated.

**Authentication**:
- Required: Yes
- Scopes: `admin.manage` or self (contributor can rotate own key)

**Rate Limiting**:
- Limit: 5 req/min (strict limit)

**Request Schema**:
```json
{
  "contributor_id": "string (UUID, optional if self)",
  "new_public_key": "string (hex-encoded Ed25519 public key, optional - server can generate)"
}
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "contributor_id": "string",
  "old_key": {
    "public_key": "string (hex)",
    "status": "deprecated",
    "deprecated_at": "string (ISO8601)"
  },
  "new_key": {
    "public_key": "string (hex)",
    "private_key": "string (hex, only returned once - client must store)",
    "created_at": "string (ISO8601)"
  },
  "migration_period_days": "number (both keys valid for 30 days)"
}
```

**Status Codes**:
- `200` - Key rotated successfully
- `400` - Invalid key format
- `401` - Unauthorized
- `403` - Forbidden (cannot rotate other contributor's key)
- `404` - Contributor not found
- `429` - Rate limit exceeded

**Business Logic**:
- Both old and new keys valid for 30 days (migration period)
- After 30 days, old key rejected
- Server can generate new key if not provided
- Key storage: KMS-encrypted at rest

---

#### POST /v0/keys/revoke (Contributor Key Revocation)

**Purpose**: Immediately revoke signing key (e.g., leaked/compromised).

**Authentication**:
- Required: Yes
- Scopes: `admin.manage` or self

**Rate Limiting**:
- Limit: 5 req/min

**Request Schema**:
```json
{
  "contributor_id": "string (UUID, optional if self)",
  "public_key_to_revoke": "string (hex-encoded Ed25519 public key)",
  "reason": "string (optional, logged for audit)"
}
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "contributor_id": "string",
  "revoked_key": "string (hex, first 16 chars shown)",
  "revoked_at": "string (ISO8601)",
  "remaining_keys": "number"
}
```

**Status Codes**:
- `200` - Key revoked
- `400` - Invalid key format
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Key not found
- `429` - Rate limit exceeded

**Business Logic**:
- Immediate revocation; no grace period
- Ingest attempts with revoked key rejected with 401
- Audit log: who revoked, when, reason

---

## 2. Restate Service & Workflow Interfaces

### Architecture Context

**Single Restate Node**: One orchestration runtime per environment (dev/staging/prod). All handlers (services, objects, workflows) register with this node via HTTP/2.

**Deployment Model**: Handlers run in separate deployable applications (Lambda, Cloud Run, K8s) and Restate calls them over HTTP with context (state, execution log).

---

### 2.1 Service: svc.mesh_ingest

**Type**: Restate Service (stateless)

**Purpose**: Lightweight synchronous handler for direct mesh ingest path. Validates envelope signature, normalizes data, upserts to SurrealDB.

**Location** (projected): `apps/restate/mesh-ingest-service/`

**Handler Signature** (Go pseudocode):
```go
type MeshIngestService struct{}

func (s *MeshIngestService) Process(
  ctx restate.Context,
  envelope IngestEnvelope,
) (IngestResponse, error) {
  // 1. Verify Ed25519 signature
  // 2. Normalize host/port/service/banner/tls_cert
  // 3. Upsert to SurrealDB (synchronous)
  // 4. Trigger async wf.enrich + wf.graph (via Restate async invocation)
  // 5. Return response with ingested_count + visible_in_secs estimate
}
```

**Input**:
```go
type IngestEnvelope struct {
  ContributorID string    // UUID
  ScanID        string    // UUID
  Signature     string    // Ed25519 hex
  Timestamp     time.Time // RFC3339
  PrivacyMode   string    // "meta-only" or "full-banner"
  Assets        []Asset
}

type Asset struct {
  Host  HostData
  Ports []PortData
}

type HostData struct {
  IP        string  // IPv4 or IPv6
  ASN       *int    // Optional
  Geo       GeoData // City, Region, Country, CloudRegion
}

type PortData struct {
  Number      int
  Protocol    string // "tcp" or "udp"
  State       string // "open", "closed", "filtered"
  Service     ServiceData
  Banner      BannerData
  TLSCert     TLSCertData
}

type ServiceData struct {
  Name    string
  Product string
  Version string
  CPE     []string
}

type BannerData struct {
  Hash        string // SHA256
  Sample      string // ≤2KB
  Fingerprint string
}

type TLSCertData struct {
  SHA256    string
  CN        string
  SANs      []string
  NotBefore time.Time
  NotAfter  time.Time
}
```

**Output**:
```go
type IngestResponse struct {
  Status              string        // "created" or "multi-status"
  ScanID              string
  AssetsIngested      int
  AssetsIgnored       int
  Messages            []string
  VisibleInSecs       int           // Estimate when visible in queries
  Errors              []IngestError // If multi-status
}

type IngestError struct {
  AssetIndex int
  ErrorCode  string
  Message    string
}
```

**Concurrency**: Stateless; multiple invocations in parallel.

**Dependencies**:
- SurrealDB client (connection pool)
- Ed25519 signature verification library
- Restate client (for async invocations)

**Async Invocations** (fire-and-forget):
```go
// After successful ingest, queue:
client.SendAsyncInvocation("wf.enrich", assetSetRef)
client.SendAsyncInvocation("wf.graph", enrichedSetRef)
```

**Error Handling**:
- Invalid signature → 400 Bad Request
- Malformed data (IP, port, CPE) → 207 Multi-Status with per-asset errors
- Database connection error → 500 Internal Server Error (will be retried by Restate)

**SLO**: P95 ≤ **2 seconds** (synchronous path only)

---

### 2.2 Workflow: wf.plan_scan

**Type**: Restate Workflow

**Purpose**: Durable scan planning. Computes stale-only targets based on age threshold and selectors. Idempotent and resumable.

**Location** (projected): `apps/restate/mesh-workflows/wf_plan_scan.go`

**Workflow Signature** (Go pseudocode):
```go
type PlanScanWorkflow struct{}

// Run handler - executes exactly once per plan_id
func (w *PlanScanWorkflow) Run(
  ctx restate.WorkflowContext,
  planID string,
  selectors Selectors,
  minAge time.Duration,
) (PlanResponse, error) {
  // 1. Query SurrealDB for matching hosts/ports
  // 2. Compute age = now - last_seen
  // 3. Filter to age >= minAge
  // 4. Apply exclusions (Do-Not-Scan list)
  // 5. Sort and paginate
  // 6. Store plan metadata in workflow state
  // 7. Return targets + plan_id for later scan submission
}

// Query handler - can run while workflow executing (read-only)
func (w *PlanScanWorkflow) GetStatus(
  ctx restate.WorkflowSharedContext,
  planID string,
) (PlanStatus, error) {
  var status PlanStatus
  ctx.Get("status", &status)
  return status, nil
}
```

**Input**:
```go
type Selectors struct {
  City         []string
  Region       []string
  Country      []string
  ASN          []int
  CIDR         []string
  CloudRegion  []string
  Service      []string
  CommonPort   []int
  Since        string         // ISO8601 or duration
}
```

**Output**:
```go
type PlanResponse struct {
  PlanID          string
  Targets         []Target
  TargetCount     int
  Pagination      PaginationInfo
  Stats           PlanStats
}

type Target struct {
  IP             string
  Port           int
  Service        string
  LastSeen       time.Time
  AgeSecs        int
  LastScannedAt  *time.Time
}

type PlanStats struct {
  TotalMatching              int
  StaleOnly                  int
  EstimatedProbeTimeMinutes  int
}
```

**Workflow State**:
```go
type PlanState struct {
  Status                 string    // "pending", "completed"
  Selectors              Selectors
  MinAge                 time.Duration
  TargetCount            int
  CreatedAt              time.Time
  CompletedAt            *time.Time
}
```

**Concurrency**: Multiple plans can execute in parallel; each plan_id has isolated state.

**Idempotency**: If same plan_id invoked twice, returns cached result.

**Dependencies**:
- SurrealDB client (for querying)
- Pagination cursor generation

**Error Handling**:
- Invalid selectors → Return error via Restate
- Database timeout → Restate retries from last checkpoint
- Empty results → Return empty target list (OK)

**SLO**: P95 ≤ **700 ms** for city-level selectors

---

### 2.3 Workflow: wf.scan

**Type**: Restate Workflow

**Purpose**: Execute scan plan or raw envelope. Normalizes and fingerprints probe results. Outputs `asset_set_ref` for downstream enrichment.

**Location** (projected): `apps/restate/mesh-workflows/wf_scan.go`

**Workflow Signature** (Go pseudocode):
```go
type ScanWorkflow struct{}

func (w *ScanWorkflow) Run(
  ctx restate.WorkflowContext,
  scanID string,
  planID *string,      // Either planID or raw envelope
  envelope *ScanEnvelope,
) (ScanResult, error) {
  // 1. Load plan targets (if planID provided)
  // 2. Execute probes (SYN on relays, CONNECT local)
  // 3. Filter stale hits (re-check against mesh cache)
  // 4. Normalize fingerprints (nmap/naabu output)
  // 5. Generate CPE candidates
  // 6. Persist asset_set_ref to state
  // 7. Return asset_set_ref for next stage
}

func (w *ScanWorkflow) GetProgress(
  ctx restate.WorkflowSharedContext,
  scanID string,
) (ScanProgress, error) {
  // Query-only; can run while Run is executing
}
```

**Input**:
```go
type ScanEnvelope struct {
  PlanID     *string  // Reference to plan
  Envelope   *IngestEnvelope  // Or raw envelope
  RelayID    string   // Which relay executed
  ProbeStats ProbeStats
}
```

**Output**:
```go
type ScanResult struct {
  ScanID       string
  AssetSetRef  string   // UUID/pointer to enriched asset set
  AssetsFound  int
  AssetsSaved  int
  Errors       []string
}
```

**Dependencies**:
- SurrealDB (persist asset_set)
- Restate async invocation (for wf.enrich)

**ISP Block Handling**:
- Relay backoff + rotation on port-ban detection
- Confirmed via e2e test

---

### 2.4 Workflow: wf.enrich

**Type**: Restate Workflow

**Purpose**: Add geo/ASN enrichment, parse TLS certs, map CPEs, (Pro) fetch NVD/OSV/KEV vulns. Outputs `enriched_set_ref`.

**Location** (projected): `apps/restate/mesh-workflows/wf_enrich.go`

**Workflow Signature** (Go pseudocode):
```go
type EnrichWorkflow struct{}

func (w *EnrichWorkflow) Run(
  ctx restate.WorkflowContext,
  assetSetRef string,
) (EnrichResult, error) {
  // 1. Load assets by assetSetRef
  // 2. For each host IP:
  //    a. Fetch GeoIP (city, region, country, lat/lon)
  //    b. Fetch ASN (lookup service)
  //    c. Detect cloud region
  // 3. For each service:
  //    a. Parse banner fingerprint
  //    b. Identify product/version
  //    c. Tag as common_port (if applicable)
  // 4. For each TLS cert:
  //    a. Parse CN, SANs, validity dates
  // 5. (Pro-only):
  //    a. Map CPEs to NVD
  //    b. Fetch OSV records
  //    c. Check CISA KEV
  // 6. Persist enrichedSetRef
  // 7. Queue wf.graph async
}
```

**Input**:
```go
type AssetSetRef struct {
  ID string  // Reference to saved assets
}
```

**Output**:
```go
type EnrichResult struct {
  EnrichedSetRef string  // UUID for next stage
  AssetsEnriched int
  NewVulnsFound  int     // Pro-only
  Errors         []string
}
```

**Enrichment Data Sources**:
- **Geo**: MaxMind GeoLite or similar (seeded)
- **ASN**: WHOIS or BGP feeds
- **CPE Mapping**: NVD CPE dictionary
- **Vulns**: NVD (required), OSV (optional), KEV (optional)
- **Common Ports**: Seed list (22→ssh, 80→http, 443→https, 6379→redis, etc.)

**Pro-gating**:
- Vulnerability joins only added if contributor has Pro subscription
- OSS contributors get enrichment but no vuln data

---

### 2.5 Workflow: wf.graph

**Type**: Restate Workflow

**Purpose**: Upsert normalized nodes/edges to SurrealDB. Update `last_seen`, `last_scanned_at`. Apply TTLs per service.

**Location** (projected): `apps/restate/mesh-workflows/wf_graph.go`

**Workflow Signature** (Go pseudocode):
```go
type GraphWorkflow struct{}

func (w *GraphWorkflow) Run(
  ctx restate.WorkflowContext,
  enrichedSetRef string,
) (GraphResult, error) {
  // 1. Load enriched assets
  // 2. For each asset:
  //    a. Upsert or fetch host node
  //    b. Upsert port node
  //    c. Upsert service node
  //    d. Create edges: host-HAS->port, port-RUNS->service
  //    e. Create EVIDENCED_BY edges to banner/tls_cert
  //    f. (Pro) Create AFFECTED_BY edges to vulns
  //    g. Create geo edges: host-IN_CITY->city, etc.
  //    h. Create network edges: host-IN_ASN->asn
  // 3. Update last_seen + last_scanned_at timestamps
  // 4. Apply TTLs (HTTP 6h, SSH 24h, RDP 12h)
  // 5. Log OBSERVED_AT metadata (scan_id, contributor_id, trust)
  // 6. Persist graph
  // 7. (Pro) Queue wf.ai_engine async for pro subscribers
}
```

**Input**:
```go
type EnrichedSetRef struct {
  ID string
}
```

**Output**:
```go
type GraphResult struct {
  NodesUpserted int
  EdgesUpserted int
  LastSeenUpdated int
  Errors []string
}
```

**Graph Schema** (SurrealDB DEFINE RELATION/THING):
```surql
-- Nodes
DEFINE TABLE host SCHEMAFULL;
DEFINE TABLE port SCHEMAFULL;
DEFINE TABLE service SCHEMAFULL;
DEFINE TABLE banner SCHEMAFULL;
DEFINE TABLE tls_cert SCHEMAFULL;
DEFINE TABLE city SCHEMAFULL;
DEFINE TABLE region SCHEMAFULL;
DEFINE TABLE country SCHEMAFULL;
DEFINE TABLE asn SCHEMAFULL;
DEFINE TABLE cloud_region SCHEMAFULL;
DEFINE TABLE common_port SCHEMAFULL;
DEFINE TABLE vuln SCHEMAFULL;        -- Pro only
DEFINE TABLE vuln_doc SCHEMAFULL;    -- Pro only

-- Edges (topology)
DEFINE TABLE has SCHEMAFULL;           -- host -> port
DEFINE TABLE runs SCHEMAFULL;          -- port -> service
DEFINE TABLE evidenced_by SCHEMAFULL;  -- service -> banner|tls_cert
DEFINE TABLE affected_by SCHEMAFULL;   -- service -> vuln (Pro only)

-- Edges (geography)
DEFINE TABLE in_city SCHEMAFULL;       -- host -> city
DEFINE TABLE in_region SCHEMAFULL;     -- city -> region
DEFINE TABLE in_country SCHEMAFULL;    -- region -> country
DEFINE TABLE in_asn SCHEMAFULL;        -- host -> asn
DEFINE TABLE in_cloud_region SCHEMAFULL; -- host -> cloud_region

-- Edges (taxonomy)
DEFINE TABLE is_common SCHEMAFULL;     -- port -> common_port

-- Edges (history)
DEFINE TABLE observed_at SCHEMAFULL;   -- service -> observed_at (metadata)
```

---

### 2.6 Workflow: wf.ai_engine (Pro)

**Type**: Restate Workflow

**Purpose**: (Pro tier only) Hybrid retrieval over vulnerability graph + vector k-NN, then LLM summarization.

**Location** (projected): `apps/restate/mesh-workflows/wf_ai_engine.go`

**Workflow Signature** (Go pseudocode):
```go
type AIEngineWorkflow struct{}

func (w *AIEngineWorkflow) Run(
  ctx restate.WorkflowContext,
  selectors Selectors,
) (AIResult, error) {
  // 1. Graph filter: Apply selectors to host/port/service nodes
  // 2. Vector search: k-NN over vuln_doc.embedding (k=20)
  // 3. Hybrid merge: Filter vector results by CPE + recency + KEV
  // 4. LLM call: GLM/Haiku model with retrieved vulns
  // 5. Generate 3-bullet summary (Summary/Risks/Next)
  // 6. Store result in workflow state
  // 7. Return summary or graph-only fallback if AI unavailable
}
```

**Input**:
```go
type AIRequest struct {
  Mode        string      // "selector", "graph", "host"
  Selectors   *Selectors  // If mode=selector
  HostIP      *string     // If mode=host
  K           int         // Top-K vulns (default 20)
  MinSeverity string      // "critical", "high", "medium", "low"
  KEVOnly     bool        // Only KEV-listed exploitable vulns
  Style       string      // "bullets", "detailed", "brief"
}
```

**Output**:
```go
type AIResult struct {
  Title           string
  Overview        string
  CriticalVulns   []VulnSummary
  TopRisks        []string   // 3 bullet points
  Recommendations []string   // Action items
  RetrievalStats  RetrievalStats
  Fallback        string     // "none" or "graph-only-summary"
}
```

**Hybrid Retrieval**:
1. **Graph Phase**: Query `(host, port, service)` matching selectors
2. **Vector Phase**: Embed query → k-NN search over `vuln_doc.embedding`
3. **Merge Phase**: Intersect vector results with CPE of matched services
4. **Ranking**: Sort by CVSS + KEV status + recency

**LLM Prompt Template**:
```
You are a security analyst. Summarize vulnerabilities affecting these hosts:
{graph_results}

Top vulnerabilities from vector search:
{vector_results}

Provide:
1. One-sentence overview
2. Top 3 critical vulnerabilities
3. Top 3 risk categories
4. Recommended mitigations

Keep each bullet under 15 words.
```

**Fallback** (if AI unavailable):
- Return graph-only vulnerability counts + severity breakdown
- HTTP 503 Service Unavailable with `fallback: "graph-only-summary"`

**SLO**: P95 ≤ **4 seconds**; fallback if timeout > 4s

---

### 2.7 Service: wf.vuln_sync (Separate App: spectra-vuln-ingest)

**Type**: Restate Service (stateless)

**Purpose**: Ingest and mirror vulnerability feeds (NVD, OSV, KEV) to Surreal `vuln_doc` table.

**Location** (projected): Separate Restate app `spectra-vuln-ingest`

**Handler Signature** (pseudocode):
```go
type VulnSyncService struct{}

func (s *VulnSyncService) SyncNVD(
  ctx restate.Context,
  since time.Time,
) (SyncResult, error) {
  // 1. Fetch NVD feed (required)
  // 2. For each CVE:
  //    a. Parse: cve_id, title, summary, cvss, cpelist
  //    b. Upsert to vuln_doc
  // 3. Return count synced
}

func (s *VulnSyncService) SyncOSV(
  ctx restate.Context,
) (SyncResult, error) {
  // Optional OSV sync
}

func (s *VulnSyncService) SyncKEV(
  ctx restate.Context,
) (SyncResult, error) {
  // Optional CISA KEV sync
}
```

**Data Flow**:
```
NVD/OSV/KEV Feeds
      │
      v
SyncNVD/OSV/KEV (stateless handler)
      │
      v
SurrealDB vuln_doc table
      │
      v
wf.vuln_vectorize (separate workflow)
```

---

### 2.8 Workflow: wf.vuln_vectorize (Pro)

**Type**: Restate Workflow

**Purpose**: Embed vulnerability documents into vector space and index in SurrealDB for k-NN searches.

**Handler Signature** (pseudocode):
```go
type VulnVectorizeWorkflow struct{}

func (w *VulnVectorizeWorkflow) Run(
  ctx restate.WorkflowContext,
  vulnDocID string,
) (VectorizeResult, error) {
  // 1. Load vuln_doc by ID
  // 2. Extract text: title + summary + cpe list
  // 3. Call embedding model (OpenAI/OSS)
  // 4. Persist vector to vuln_doc.embedding field
  // 5. Index added automatically by SurrealDB (cosine ANN)
}
```

**Embedding Model**: TBD (options: OpenAI, OSS like `sentence-transformers`)

**Vector Dimensions**: TBD (typically 1536 for OpenAI)

**Index Type**: SurrealDB cosine similarity ANN

---

## 3. Authentication & Authorization

### 3.1 OAuth2 PKCE Flow

**Grant Type**: Authorization Code with PKCE

**Authorization Server**: Hosted by Spectra-Red

**Flow**:
```
1. User clicks "Login" in CLI/UI
2. CLI generates:
   - code_challenge = base64url(sha256(code_verifier))
   - state = random nonce
3. Redirects to: https://auth.spectra.red/authorize
   ?client_id=...
   &redirect_uri=http://localhost:9999  (loopback for CLI)
   &scope=mesh.read%20mesh.write%20ai.use
   &code_challenge=...
   &code_challenge_method=S256
   &state=...
4. User approves; redirected to:
   http://localhost:9999/?code=...&state=...
5. CLI exchanges code for token:
   POST https://auth.spectra.red/token
   {
     "grant_type": "authorization_code",
     "code": "...",
     "code_verifier": "...",
     "client_id": "..."
   }
6. Returns:
   {
     "access_token": "eyJ...",
     "token_type": "Bearer",
     "expires_in": 3600,
     "refresh_token": "..."
   }
7. CLI stores in keyring (platform-specific)
```

**Scope Mapping**:
- `mesh.read` - Read `/v0/search`, `/v0/host/{ip}`, `/v0/plan`, `/v0/coverage`
- `mesh.write` - Write `/v0/mesh/ingest`, `/v0/ingest`
- `ai.use` - Write `/v0/ai/summary` (Pro tier)
- `admin.manage` - Key rotation/revocation (admin)

**Token Lifetime**:
- Access token: 1 hour
- Refresh token: 30 days
- Refresh flow: `POST /token` with `grant_type=refresh_token`

---

### 3.2 JWT Token Structure

**Header**:
```json
{
  "alg": "RS256",
  "kid": "spectra-key-20250101",
  "typ": "JWT"
}
```

**Payload**:
```json
{
  "iss": "https://auth.spectra.red",
  "sub": "contributor-550e8400-e29b-41d4-a716-446655440000",
  "aud": "https://api.spectra.red",
  "iat": 1667900400,
  "exp": 1667904000,
  "scope": "mesh.read mesh.write ai.use",
  "tier": "pro|oss|enterprise",
  "consent": {
    "accepted_at": "2025-01-01T00:00:00Z",
    "version": "1.0"
  }
}
```

**Claims**:
- `sub` - Contributor ID (unique per user/org)
- `scope` - Space-separated scopes granted
- `tier` - Subscription tier (determines feature access)
- `consent` - Legal/compliance consent metadata

---

### 3.3 Envelope Signing (Ed25519)

**Purpose**: Verify ingest submission authenticity; prevent spoofing of source IP/contributor.

**Key Management**:
- Each contributor has primary + deprecated keys
- Keys stored in Spectra KMS (at-rest encrypted)
- Rotation via `/v0/keys/rotate` (30-day grace period)
- Revocation via `/v0/keys/revoke` (immediate)

**Signing Process** (runner side):
```
1. Contributor stores private key locally (encrypted keyring)
2. For each ingest envelope:
   a. Serialize: JSON canonical form (RFC 7159)
   b. Hash: SHA512(canonical_json)
   c. Sign: Ed25519_sign(hash, private_key)
   d. Include signature in envelope.signature field
3. Submit via HTTPS + OAuth JWT
```

**Verification** (server side):
```
1. Extract public_key from contributor profile
2. Extract signature from envelope
3. Serialize envelope (excluding signature field)
4. Hash: SHA512(canonical_json)
5. Verify: Ed25519_verify(signature, hash, public_key)
6. If invalid → 400 Bad Request
7. Log signature_verified metadata in OBSERVED_AT
```

---

### 3.4 Scoped Access Control

**Tier-Based Gating**:

| Endpoint/Feature | OSS | Pro | Enterprise |
|------------------|-----|-----|-----------|
| `/v0/mesh/ingest` | ✓ | ✓ | ✓ |
| `/v0/ingest` | ✓ | ✓ | ✓ |
| `/v0/plan` | ✓ | ✓ (higher quota) | ✓ |
| `/v0/coverage` | ✓ | ✓ | ✓ |
| `/v0/search` | ✓ | ✓ (higher quota) | ✓ |
| `/v0/host/{ip}` | ✓ | ✓ | ✓ |
| Vuln edges | ✗ | ✓ | ✓ |
| Vector RAG | ✗ | ✓ | ✓ |
| `/v0/ai/summary` | ✗ | ✓ | ✓ |
| Read-only Surreal mirror | ✗ | ✗ | ✓ |

**Pro-gated Responses**:
- Any response containing vuln/vector/AI content returns `402 Payment Required` if called with OSS key
- Example: `GET /v0/host/{ip}?include=vulns` with OSS token → 402

---

## 4. Rate Limiting Policies

### 4.1 Per-Endpoint Limits (MVP)

| Endpoint | Limit | Unit | Scope |
|----------|-------|------|-------|
| `/v0/mesh/ingest` | 60 | req/min | Per contributor |
| `/v0/ingest` | 60 | req/min | Shared with mesh/ingest |
| `/v0/plan` | 30 | req/min | Per contributor |
| `/v0/coverage` | 60 | req/min | Per contributor |
| `/v0/search` | 60 | req/min | Per contributor |
| `/v0/host/{ip}` | 60 | req/min | Per contributor |
| `/v0/ai/summary` | 20 | req/min | Per contributor (Pro only) |
| `/v0/keys/rotate` | 5 | req/min | Per contributor (strict) |
| `/v0/keys/revoke` | 5 | req/min | Per contributor (strict) |

### 4.2 Rate Limit Response Headers

**On each response**:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1667900460 (Unix timestamp)
```

**When limit exceeded (429)**:
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1667900460
Retry-After: 30

{
  "status": "rate_limit_exceeded",
  "message": "60 requests per minute limit exceeded",
  "reset_in_seconds": 30
}
```

### 4.3 Quota Tiers

**OSS**:
- Ingest: 60 req/min
- Plan: 30 req/min
- Search: 60 req/min (≤5k results)

**Pro**:
- Ingest: 120 req/min
- Plan: 60 req/min
- Search: 120 req/min (≤50k results)
- AI: 20 req/min

**Enterprise**:
- All Pro limits + 2x multiplier
- Custom negotiable

---

## 5. Error Response Formats

### 5.1 Standard Error Response

**Format** (all error status codes):
```json
{
  "status": "error",
  "error_code": "string (machine-readable)",
  "message": "string (human-readable)",
  "request_id": "string (UUID for logging)",
  "details": {
    "field": "string (optional, for validation errors)",
    "reason": "string (optional)"
  }
}
```

### 5.2 Common Error Codes

| HTTP | Code | Message | When |
|------|------|---------|------|
| 400 | `bad_request` | Invalid request format | Malformed JSON, missing fields |
| 400 | `invalid_ip` | Invalid IP address | IP outside valid ranges |
| 400 | `invalid_port` | Port out of range | Port < 1 or > 65535 |
| 400 | `invalid_cpe` | Malformed CPE URI | CPE doesn't match v2.3 format |
| 400 | `invalid_selector` | Unknown selector field | Selector not in schema |
| 400 | `invalid_signature` | Envelope signature verification failed | Ed25519 validation error |
| 401 | `unauthorized` | Missing or invalid token | No JWT, expired JWT, wrong key |
| 401 | `invalid_scope` | Token missing required scope | `mesh.write` needed but token has only `mesh.read` |
| 402 | `payment_required` | Pro feature requires upgrade | Vuln/vector/AI endpoint with OSS key |
| 403 | `forbidden` | Not permitted | Trying to revoke another's key |
| 404 | `not_found` | Resource not found | Host not in mesh, plan doesn't exist |
| 429 | `rate_limit_exceeded` | Rate limit exceeded | Too many requests |
| 500 | `internal_error` | Internal server error | Database error, Restate unavailable |
| 503 | `service_unavailable` | Service temporarily unavailable | AI model timeout (returns fallback) |

### 5.3 Validation Error (400)

```json
{
  "status": "error",
  "error_code": "validation_error",
  "message": "Request validation failed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "details": {
    "errors": [
      {
        "field": "assets[0].host.ip",
        "code": "invalid_ip",
        "message": "192.0.2.999 is not a valid IP address"
      },
      {
        "field": "assets[0].ports[1].number",
        "code": "invalid_port",
        "message": "Port must be between 1 and 65535"
      }
    ]
  }
}
```

### 5.4 Multi-Status Response (207)

```json
{
  "status": "multi-status",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "summary": {
    "total": 10,
    "success": 8,
    "failed": 2
  },
  "results": [
    {
      "index": 0,
      "status": "success",
      "asset_id": "550e8400-e29b-41d4-a716-446655440001"
    },
    {
      "index": 5,
      "status": "error",
      "error_code": "invalid_ip",
      "message": "Invalid IP address"
    }
  ]
}
```

---

## 6. Data Model Overview

### 6.1 Core Node Types

```surql
-- Hosts
DEFINE TABLE host SCHEMAFULL {
  FIELDS {
    ip: string,                      -- PK: IPv4 or IPv6
    asn: int,                        -- Optional ASN
    geo_cc: string,                  -- ISO 3166-1 alpha-2
    city: string,                    -- City name (optional)
    region: string,                  -- Region/state (optional)
    country: string,                 -- Country name
    cloud_region: string,            -- Cloud provider region (optional)
    first_seen: datetime,            -- When first observed
    last_seen: datetime,             -- When last observed
    last_scanned_at: datetime        -- When last scanned
  },
  INDEXES {
    idx_ip: UNIQUE,
    idx_asn,
    idx_city,
    idx_cloud_region
  }
};

-- Ports
DEFINE TABLE port SCHEMAFULL {
  FIELDS {
    number: int,                     -- PK: 1-65535
    protocol: string,                -- "tcp" or "udp"
    transport: string,               -- Transport type (optional)
    common: bool,                    -- Is in common_port list
    first_seen: datetime,
    last_seen: datetime
  },
  INDEXES {
    idx_number: UNIQUE,
    idx_common
  }
};

-- Services
DEFINE TABLE service SCHEMAFULL {
  FIELDS {
    name: string,                    -- e.g., "http", "ssh"
    product: string,                 -- e.g., "Apache" (optional)
    version: string,                 -- Version string (optional)
    cpe: [string],                   -- CPE v2.3 URIs
    fp: string,                      -- Fingerprint hash
    first_seen: datetime,
    last_seen: datetime
  },
  INDEXES {
    idx_name,
    idx_product,
    idx_cpe,
    idx_fp: UNIQUE
  }
};

-- Banners
DEFINE TABLE banner SCHEMAFULL {
  FIELDS {
    hash: string,                    -- PK: SHA256
    sample: string,                  -- Actual banner (≤2KB)
    first_seen: datetime,
    last_seen: datetime
  },
  INDEXES {
    idx_hash: UNIQUE
  }
};

-- TLS Certificates
DEFINE TABLE tls_cert SCHEMAFULL {
  FIELDS {
    sha256: string,                  -- PK: Certificate hash
    cn: string,                      -- Common name
    sans: [string],                  -- Subject alternate names
    not_before: datetime,            -- Validity start
    not_after: datetime,             -- Validity end
    issuer: string,                  -- Issuer DN (optional)
    first_seen: datetime,
    last_seen: datetime
  },
  INDEXES {
    idx_sha256: UNIQUE,
    idx_cn,
    idx_sans
  }
};

-- Vulnerabilities (Pro)
DEFINE TABLE vuln SCHEMAFULL {
  FIELDS {
    cve_id: string,                  -- PK: e.g., CVE-2024-1234
    cvss: float,                     -- CVSS v3.1 score
    severity: string,                -- critical|high|medium|low
    published: datetime,
    kev_listed: bool                 -- CISA KEV catalog
  },
  INDEXES {
    idx_cve_id: UNIQUE,
    idx_severity
  }
};

-- Vulnerability Documents (Pro, for vector RAG)
DEFINE TABLE vuln_doc SCHEMAFULL {
  FIELDS {
    cve_id: string,                  -- PK/FK to vuln
    title: string,
    summary: string,                 -- Full description
    cvss: float,
    epss: float,                     -- Exploit Prediction Scoring System (optional)
    cpe: [string],                   -- Affected CPEs
    exploit_refs: [string],          -- URLs to exploit PoCs
    embedding: [float],              -- Vector embedding (dimension TBD)
    published: datetime,
    updated: datetime
  },
  INDEXES {
    idx_cve_id: UNIQUE,
    idx_cpe,
    idx_embedding: MTREE (cosine)     -- Vector ANN index
  }
};
```

### 6.2 Geography & Taxonomy Nodes

```surql
DEFINE TABLE city {
  FIELDS {
    id: string,                      -- PK: "Paris"
    name: string,
    cc: string,                      -- Country code
    lat: float,
    lon: float
  }
};

DEFINE TABLE region {
  FIELDS {
    id: string,                      -- PK: "FR-75" (ISO 3166-2)
    name: string,
    cc: string
  }
};

DEFINE TABLE country {
  FIELDS {
    cc: string,                      -- PK: ISO 3166-1 alpha-2
    name: string
  }
};

DEFINE TABLE asn {
  FIELDS {
    number: int,                     -- PK: AS number
    org: string                      -- Organization name
  }
};

DEFINE TABLE cloud_region {
  FIELDS {
    id: string,                      -- PK: "us-east-1"
    provider: string,                -- "aws", "gcp", "azure"
    code: string,
    name: string
  }
};

DEFINE TABLE common_port {
  FIELDS {
    number: int,                     -- PK
    label: string                    -- e.g., "ssh", "http"
  }
};
```

### 6.3 Edge Types

```surql
-- Topology
DEFINE TABLE has {
  FIELDS {
    in: record<host>,                -- Source host
    out: record<port>,               -- Target port
    first_seen: datetime,
    last_seen: datetime
  }
};

DEFINE TABLE runs {
  FIELDS {
    in: record<port>,                -- Source port
    out: record<service>,            -- Target service
    fingerprint: string,             -- Service fingerprint
    first_seen: datetime,
    last_seen: datetime
  }
};

DEFINE TABLE evidenced_by {
  FIELDS {
    in: record<service>,             -- Source service
    out: record<banner|tls_cert>,    -- Evidence
    scan_id: string                  -- Which scan observed this
  }
};

DEFINE TABLE affected_by {
  FIELDS {
    in: record<service>,             -- Vulnerable service
    out: record<vuln>,               -- Vulnerability
    score: float                     -- CVSS for this match
  }
};

-- History
DEFINE TABLE observed_at {
  FIELDS {
    in: record<service>,             -- Service observed
    out: record<>,                   -- Observation metadata
    scan_id: string,                 -- Scan that observed it
    contributor_id: string,          -- Who submitted
    ts: datetime,                    -- When observed
    trust: bool                      -- Signature verified
  }
};

-- Geography
DEFINE TABLE in_city { in: record<host>, out: record<city> };
DEFINE TABLE in_region { in: record<city>, out: record<region> };
DEFINE TABLE in_country { in: record<region>, out: record<country> };
DEFINE TABLE in_asn { in: record<host>, out: record<asn> };
DEFINE TABLE in_cloud_region { in: record<host>, out: record<cloud_region> };

-- Taxonomy
DEFINE TABLE is_common { in: record<port>, out: record<common_port> };
```

---

## 7. Example API Flows

### 7.1 Complete Ingest → Enrich → Query Flow

```
User CLI with private key
      │
      ├─ Sign envelope (Ed25519)
      │
      ├─ OAuth2 login (PKCE)
      │
      └─ POST /v0/mesh/ingest + JWT
           │
           v
      API Server (svc.mesh_ingest)
           │
           ├─ Verify JWT (scope=mesh.write)
           ├─ Verify Ed25519 signature
           ├─ Normalize/validate data
           │
           └─ Upsert to SurrealDB (SYNC):
              ├─ host (192.0.2.1)
              ├─ port (80)
              ├─ service (http, Apache 2.4.41)
              └─ banner hash + OBSERVED_AT
           │
           ├─ Respond 201 Created
           │
           └─ Queue async Restate workflows:
              ├─ wf.enrich(asset_set_ref)
              │   ├─ Add geo (Paris, FR)
              │   ├─ Add ASN (65001)
              │   ├─ Parse CPE → NVD lookup (Pro)
              │   └─ Return enriched_set_ref
              │
              └─ wf.graph(enriched_set_ref)
                  ├─ Upsert host → city → region → country
                  ├─ Upsert port → is_common (if 80)
                  ├─ Upsert service with CPE edges
                  ├─ Update last_seen timestamps
                  └─ (Pro) Create affected_by edges to vulns
           │
           └─ SLA: visible in 2s P95
           │
           └─ User queries:
              GET /v0/search?city=Paris&service=http
                  │
                  └─ Returns: (192.0.2.1, 80, Apache 2.4.41)
                             + last_seen, observation_count
```

### 7.2 Planning & Rescan Flow

```
User: "Plan rescan of Paris, 5m age min"
      │
      └─ POST /v0/plan
           {
             "selectors": {"city": ["Paris"]},
             "min_age": "5m"
           }
           │
           v
      wf.plan_scan(plan_id)
           │
           ├─ Query: host-IN_CITY->city{name="Paris"}
           ├─ Compute: age = now - host.last_seen
           ├─ Filter: age >= 5m
           ├─ Exclude: Do-Not-Scan list
           │
           └─ Return plan_id + [targets]
                {
                  "plan_id": "550e8400...",
                  "targets": [
                    {"ip": "192.0.2.1", "port": 80, ...},
                    {"ip": "192.0.2.2", "port": 22, ...}
                  ]
                }
           │
User: "Execute plan"
      │
      └─ wf.scan(plan_id)
           │
           ├─ Load targets from plan_id
           ├─ Execute probes (relay SYN, local CONNECT)
           ├─ Filter hits vs. mesh cache (skip stale)
           ├─ Normalize fingerprints
           │
           └─ POST /v0/mesh/ingest (self-loop)
                ├─ Sign envelope
                ├─ Ingest normalized results
                ├─ Trigger wf.enrich + wf.graph
                │
                └─ Results visible in 2s
```

### 7.3 Pro: AI Summary Flow

```
User: "Summarize critical vulns in Paris"
      │
      └─ POST /v0/ai/summary
           {
             "mode": "selector",
             "selector": {
               "city": ["Paris"],
               "min_severity": "critical"
             },
             "k": 20
           }
           │
           v
      wf.ai_engine
           │
           ├─ Graph filter: host-IN_CITY->city{Paris}
           │   │
           │   └─ Results: {192.0.2.1:80, 192.0.2.2:443, ...}
           │
           ├─ Vector search: k-NN over vuln_doc.embedding
           │   │
           │   └─ Top-20 vulns by cosine similarity
           │
           ├─ Hybrid merge: Intersect by CPE + KEV status
           │   │
           │   └─ Ranked list of exploitable vulns
           │
           ├─ LLM: GLM/Haiku summarization
           │   │
           │   └─ Prompt: "Summarize these vulns for Paris hosts"
           │
           └─ Return:
              {
                "title": "Critical Vulns in Paris",
                "overview": "3 critical RCE CVEs affecting 45% of hosts",
                "critical_vulns": [
                  {"cve_id": "CVE-2024-1234", "cvss": 9.8}
                ],
                "top_risks": [
                  "Unpatched Apache with RCE",
                  "Exposed Redis (6379)",
                  "Outdated SSH with weak ciphers"
                ],
                "recommendations": [...]
              }
```

---

## 8. Implementation Checklist

### API Endpoints (Go Services)

- [ ] `POST /v0/mesh/ingest` - svc.mesh_ingest (Restate Service)
- [ ] `POST /v0/ingest` - Router to wf.scan (async workflow)
- [ ] `POST /v0/plan` - Router to wf.plan_scan
- [ ] `GET /v0/coverage` - SurrealDB aggregation query
- [ ] `GET /v0/host/{ip}` - SurrealDB graph traversal
- [ ] `GET /v0/search` - SurrealDB selector query with pagination
- [ ] `POST /v0/ai/summary` - Router to wf.ai_engine (Pro-gated)
- [ ] `POST /v0/keys/rotate` - Key rotation handler
- [ ] `POST /v0/keys/revoke` - Key revocation handler

### Restate Workflows

- [ ] `wf.plan_scan` - Scan planning (selector → stale targets)
- [ ] `wf.scan` - Probe execution + normalization
- [ ] `wf.enrich` - Geo/ASN/CPE/vuln enrichment
- [ ] `wf.graph` - SurrealDB upsert (nodes + edges)
- [ ] `wf.ai_engine` (Pro) - Hybrid RAG + LLM summary
- [ ] `wf.vuln_sync` (separate app) - NVD/OSV/KEV mirror
- [ ] `wf.vuln_vectorize` (Pro) - Vector embedding

### Authentication

- [ ] OAuth2 PKCE flow + JWT issuance
- [ ] JWT validation middleware (scope + tier checks)
- [ ] Ed25519 envelope signature verification
- [ ] Tier-based response filtering (402 Payment Required)

### Database (SurrealDB)

- [ ] Schema: host, port, service, banner, tls_cert
- [ ] Schema: city, region, country, asn, cloud_region, common_port
- [ ] Schema: vuln, vuln_doc (Pro)
- [ ] Edges: has, runs, evidenced_by, affected_by (Pro)
- [ ] Edges: geography (in_city, in_region, etc.)
- [ ] Indices: ip, port.number, service.name, asn.number, etc.
- [ ] Vector index: vuln_doc.embedding (cosine ANN)

### Tests

- [ ] E2E: Ingest → visible in query (2s SLO)
- [ ] E2E: Plan → rescan → graph update
- [ ] E2E: Pro features blocked with OSS key (402)
- [ ] E2E: Relay rotation on port-ban
- [ ] Unit: Signature verification (valid/invalid)
- [ ] Unit: Rate limiting headers + 429 responses
- [ ] Unit: Pagination cursor consistency
- [ ] Load: P95 plan SLA (700ms city-level)
- [ ] Load: P95 query SLA (600ms ≤5k)
- [ ] Load: P95 ingest SLA (2s visible)

---

## 9. References & Resources

**PRD Source**:
- `/Users/seanknowles/Library/Application Support/com.conductor.app/uploads/originals/1e408e9c-b3f7-4149-9238-c3013f9b2d8b.txt`

**Restate Documentation**:
- [Restate Go SDK](https://docs.restate.dev/develop/go/overview/)
- [Durable Execution](https://docs.restate.dev/develop/concepts)
- [State Management](https://docs.restate.dev/develop/go/state/)
- [Serving Strategies](https://docs.restate.dev/develop/go/serving/)

**Standards**:
- [OAuth 2.0 PKCE (RFC 7636)](https://tools.ietf.org/html/rfc7636)
- [JWT (RFC 7519)](https://tools.ietf.org/html/rfc7519)
- [CPE v2.3 (NIST)](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-188.pdf)
- [CVSS v3.1 (FIRST)](https://www.first.org/cvss/v3.1/specification-document)

**Related Protocols**:
- [SurrealDB Documentation](https://surrealdb.com/docs/)
- [Ed25519 Signature Scheme (RFC 8032)](https://tools.ietf.org/html/rfc8032)

---

**Document Version**: 1.0  
**Last Updated**: 2025-11-01  
**Status**: Complete (MVP API Context)
