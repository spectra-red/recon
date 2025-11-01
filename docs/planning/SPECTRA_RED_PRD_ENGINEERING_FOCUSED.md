# Spectra-Red Intel Mesh: Engineering-Focused PRD
## Product Requirements Document for Solo Developer Implementation

**Version:** 1.0
**Last Updated:** November 1, 2025
**Author:** Product & Engineering
**Status:** Ready for Implementation

---

## Executive Summary (2-Minute Read)

### What We're Building
A **community-driven security intelligence mesh** that lets researchers, red teams, and security professionals:
- Share real-time network scan data (ports, services, vulnerabilities)
- Query a global graph of internet-facing assets
- Get AI-powered vulnerability summaries and threat assessments

### The Problem
Current tools (Shodan, Censys) are:
- **Stale:** Data is 2-4 weeks old
- **Expensive:** $500-2K/year for comprehensive access
- **Siloed:** No community contributions, everyone scans the same targets repeatedly
- **Time-consuming:** Manual correlation of vulnerabilities across data sources

### The Solution
A **distributed mesh** where:
- Community runners submit fresh scan data (gets you free Pro access)
- Central graph database indexes everything (SurrealDB)
- Durable workflows handle enrichment and AI analysis (Restate)
- Real-time queries return results in <2 seconds
- OSS tier is free; Pro adds AI + vuln correlation

### Why This Can Win
- **Novel approach:** No one else does community mesh contributions
- **Real-time data:** Sub-2-second freshness vs. weeks-old data
- **AI integration:** Only platform with hybrid graph+vector RAG for vuln analysis
- **Cost advantage:** Free tier viable because community provides data

### MVP Scope (Solo Developer, 20 Weeks)
**Phase 1 (Weeks 1-8):** Core mesh ingest + graph storage + basic queries
**Phase 2 (Weeks 9-14):** Enrichment pipeline + vulnerability correlation
**Phase 3 (Weeks 15-20):** AI summarization + Pro tier features

**Target:** Ship usable MVP in 5 months with 100+ community beta testers

---

## 1. User Needs & Pain Points

### Primary User: OSS Security Researcher / Red Team Operator

**Current Workflow (Reconnaissance Phase):**
```
1. Run Nmap/Masscan scans (2-4 hours for medium targets)
2. Check Shodan for historical data ($59-99/month subscription)
3. Cross-reference Censys for certificate data (separate account)
4. Look up CVEs manually in NVD database (30-60 min)
5. Correlate findings across 3-4 tools (90+ minutes)
6. Write report summarizing risks (60-120 minutes)

Total time: 6-10 hours per target/engagement
```

**Pain Points (Ranked by Impact):**
1. **Data Staleness:** Shodan data is 2-4 weeks old; misses recent exposures (Critical)
2. **Cost Barrier:** $500-2K/year for Shodan Pro + Censys + other tools (High)
3. **Tool Fragmentation:** No single source of truth; manual correlation (High)
4. **Time Waste:** Duplicate scanning by thousands of researchers (Medium)
5. **Geographic Gaps:** Shodan is US-weighted; poor coverage in EU/APAC (Medium)
6. **No AI Assistance:** Manual CVE research and risk prioritization (Medium)

**Desired Outcome:**
- Real-time asset data (minutes old, not weeks)
- Single API/interface for all reconnaissance needs
- Automated vulnerability correlation and prioritization
- Free tier for contributors; affordable Pro tier ($50-100/month)
- Cut reconnaissance time from 6-10 hours to 1-2 hours

**Jobs to Be Done:**
- "I need to quickly discover what ports/services a target is running **right now**"
- "I want to see all known vulnerabilities for detected services without manual CVE lookup"
- "I need AI to summarize risks and suggest next steps so I can focus on exploitation"
- "I want to contribute my scan data in exchange for free access to the mesh"

---

## 2. Product Vision & Scope

### Product Vision
**"The Wikipedia of security intelligence"** - community-driven, real-time, AI-enhanced threat intelligence accessible to everyone.

### Core Value Proposition
**For security researchers:** Free real-time intelligence in exchange for contributing scan data
**For red teams:** 50-70% time savings on reconnaissance with AI-powered analysis
**For enterprises:** Continuous external asset monitoring at 1/10th the cost of traditional platforms

### In Scope for MVP (Solo Developer, 20 Weeks)

#### Must Have (P0 - Weeks 1-8)
- ✅ **Mesh Ingest API:** Accept scan submissions from community runners
- ✅ **SurrealDB Graph:** Store hosts, ports, services, banners, TLS certs
- ✅ **Basic Enrichment:** ASN lookup, GeoIP (city/country), common port tagging
- ✅ **Planning API:** Return stale targets based on freshness thresholds
- ✅ **Query API:** Search by IP, city, ASN, service type, port number
- ✅ **CLI Tool:** `spectra scan`, `spectra mesh plan`, `spectra mesh query`
- ✅ **Authentication:** Ed25519 envelope signing + basic OAuth2

#### Should Have (P1 - Weeks 9-14)
- ✅ **NVD Integration:** Fetch CVEs, map to CPE, link to services
- ✅ **Vulnerability Graph:** Service→AFFECTED_BY→Vuln edges
- ✅ **Restate Workflows:** Durable scan/enrich/graph pipeline
- ✅ **Pro Tier Gating:** 402 Payment Required for vuln features
- ✅ **Relay Rotation:** Basic ISP block mitigation (3-5 relays)

#### Nice to Have (P2 - Weeks 15-20)
- ✅ **AI Summarization:** OpenAI GPT-4 Turbo for risk summaries
- ✅ **Vector RAG:** Hybrid graph+vector search for vuln discovery
- ✅ **CISA KEV Integration:** Flag known exploited vulnerabilities
- ✅ **Coverage API:** Freshness histograms and gap analysis

### Out of Scope (Post-MVP)
- ❌ Billing/payments system (manual Pro tier activation initially)
- ❌ Web UI dashboard (CLI + API only for MVP)
- ❌ SIEM integrations (Splunk, Elastic)
- ❌ Multi-node Restate cluster (single node MVP)
- ❌ SSO/SAML (OAuth2 only)
- ❌ Read-only database mirrors (Enterprise feature)
- ❌ UDP scanning (TCP only for MVP)
- ❌ Passive DNS/certificate transparency feeds

---

## 3. Technical Architecture

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    COMMUNITY SCAN RUNNERS                        │
│  (Hetzner, Cloud, On-Prem - community operated)                 │
└─────────────────────────────────────────────────────────────────┘
                             │
                    ┌────────┴────────┐
                    │                 │
        Naabu Scans │     Nmap Scans  │
        (SYN/CONNECT)    (Service ID)  │
                    │                 │
                    └────────┬────────┘
                             │ HTTPS (TLS)
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  Spectra CLI (Go)                                               │
│  - spectra scan --plan <id> --submit                           │
│  - spectra mesh plan city "Paris" --min-age 5m                 │
│  - spectra mesh query service redis --city Paris               │
│  - spectra explain ip 1.2.3.4 (Pro)                           │
│                                                                 │
│  Authentication: Ed25519 envelope signing + JWT                 │
└─────────────────────────────────────────────────────────────────┘
                             │
                    ┌────────┴────────┐
                    │                 │
              POST /v0/mesh/ingest  POST /v0/ingest
                    │                 │
                    ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│  svc.api (Go HTTP Server - Chi Router)                         │
│  - JWT validation (OAuth2)                                      │
│  - Rate limiting (60 req/min ingest, 30 req/min plan)         │
│  - Ed25519 signature verification                              │
│  - Request routing to Restate                                  │
└─────────────────────────────────────────────────────────────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
          Fast Path     Workflow     Query
          (2s SLO)      (Async)     (600ms SLO)
                │            │            │
                ▼            ▼            ▼
┌──────────────────┐  ┌──────────────────────────────────┐
│   SurrealDB      │  │  Restate Node (Single)           │
│   (Graph DB)     │◄─┤  - wf.plan_scan                  │
│                  │  │  - wf.scan                       │
│  Nodes:          │  │  - wf.enrich                     │
│  - host          │  │  - wf.graph                      │
│  - port          │  │  - wf.ai_engine (Pro)            │
│  - service       │  └──────────────────────────────────┘
│  - banner        │           │
│  - tls_cert      │           │ Enrichment Calls
│  - vuln_doc      │           ▼
│  - city/asn/etc  │  ┌──────────────────────────────────┐
│                  │  │  External Services               │
│  Edges:          │  │  - MaxMind GeoIP (ASN/City)     │
│  - HAS           │  │  - NVD API (CVE data)           │
│  - RUNS          │  │  - CISA KEV (exploited CVEs)    │
│  - AFFECTED_BY   │  │  - OpenAI API (AI summaries)    │
│  - IN_CITY       │  └──────────────────────────────────┘
│  - OBSERVED_AT   │
│                  │
│  Indices:        │
│  - host.ip       │
│  - port.number   │
│  - service.fp    │
│  - city.name     │
│  - vector ANN    │
└──────────────────┘
```

### Technology Stack (Solo Developer Friendly)

| Component | Technology | Why This Choice | Learning Curve |
|-----------|------------|-----------------|----------------|
| **Language** | Golang 1.23+ | Fast, concurrent, single binary, great for CLIs | Medium (2-3 days if new) |
| **Orchestration** | Restate | Durable execution, automatic retries, exactly-once semantics | Medium (1 week) |
| **Database** | SurrealDB | Graph + documents + vectors in one DB, no PostgreSQL complexity | Low-Medium (3-5 days) |
| **HTTP Framework** | Chi Router | Lightweight, idiomatic Go, simple middleware | Low (1 day) |
| **CLI Framework** | Cobra (built-in) | Standard Go CLI library, clean UX | Low (1-2 days) |
| **AI** | OpenAI SDK | Mature, well-documented, $0.01-0.03 per summary | Low (1 day) |
| **Auth** | golang-jwt/jwt | Standard JWT library, OAuth2 helpers available | Low (1 day) |
| **Testing** | Testify + GoMock | Industry standard, good docs | Low (1 day) |
| **Logging** | Zap | Structured logging, high performance | Low (2 hours) |
| **Metrics** | Prometheus | Standard, integrates everywhere | Low (4 hours) |

**Total Learning Time (if all new):** ~2-3 weeks part-time

### Data Model (SurrealDB Schema)

```sql
-- Core asset nodes
DEFINE TABLE host SCHEMAFULL;
DEFINE FIELD ip ON TABLE host TYPE string ASSERT $value != NONE;
DEFINE FIELD asn ON TABLE host TYPE int;
DEFINE FIELD city ON TABLE host TYPE string;
DEFINE FIELD region ON TABLE host TYPE string;
DEFINE FIELD country ON TABLE host TYPE string;
DEFINE FIELD cloud_region ON TABLE host TYPE string;
DEFINE FIELD first_seen ON TABLE host TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE host TYPE datetime DEFAULT time::now();
DEFINE FIELD last_scanned_at ON TABLE host TYPE datetime;
DEFINE INDEX idx_host_ip ON TABLE host COLUMNS ip UNIQUE;

DEFINE TABLE port SCHEMAFULL;
DEFINE FIELD number ON TABLE port TYPE int ASSERT $value > 0 AND $value < 65536;
DEFINE FIELD protocol ON TABLE port TYPE string ASSERT $value IN ['tcp', 'udp'];
DEFINE FIELD transport ON TABLE port TYPE string; -- e.g., 'tls', 'plain'
DEFINE FIELD common ON TABLE port TYPE bool DEFAULT false;
DEFINE FIELD first_seen ON TABLE port TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE port TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_port_number ON TABLE port COLUMNS number;

DEFINE TABLE service SCHEMAFULL;
DEFINE FIELD name ON TABLE service TYPE string; -- e.g., 'http', 'ssh'
DEFINE FIELD product ON TABLE service TYPE string; -- e.g., 'nginx', 'openssh'
DEFINE FIELD version ON TABLE service TYPE string; -- e.g., '1.25.1'
DEFINE FIELD cpe ON TABLE service TYPE array<string>; -- CPE 2.3 identifiers
DEFINE FIELD fingerprint ON TABLE service TYPE string; -- hash for dedup
DEFINE FIELD first_seen ON TABLE service TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE service TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_service_fp ON TABLE service COLUMNS fingerprint;
DEFINE INDEX idx_service_name ON TABLE service COLUMNS name;

DEFINE TABLE banner SCHEMAFULL;
DEFINE FIELD hash ON TABLE banner TYPE string ASSERT $value != NONE;
DEFINE FIELD sample ON TABLE banner TYPE string; -- max 2KB
DEFINE FIELD first_seen ON TABLE banner TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_banner_hash ON TABLE banner COLUMNS hash UNIQUE;

DEFINE TABLE tls_cert SCHEMAFULL;
DEFINE FIELD sha256 ON TABLE tls_cert TYPE string ASSERT $value != NONE;
DEFINE FIELD cn ON TABLE tls_cert TYPE string; -- common name
DEFINE FIELD sans ON TABLE tls_cert TYPE array<string>; -- subject alt names
DEFINE FIELD not_before ON TABLE tls_cert TYPE datetime;
DEFINE FIELD not_after ON TABLE tls_cert TYPE datetime;
DEFINE FIELD first_seen ON TABLE tls_cert TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_tls_sha256 ON TABLE tls_cert COLUMNS sha256 UNIQUE;

-- Vulnerability nodes (Pro tier)
DEFINE TABLE vuln SCHEMAFULL;
DEFINE FIELD cve_id ON TABLE vuln TYPE string ASSERT $value != NONE;
DEFINE FIELD cvss ON TABLE vuln TYPE float;
DEFINE FIELD severity ON TABLE vuln TYPE string; -- 'critical', 'high', 'medium', 'low'
DEFINE FIELD kev_flag ON TABLE vuln TYPE bool DEFAULT false; -- CISA known exploited
DEFINE INDEX idx_vuln_cve ON TABLE vuln COLUMNS cve_id UNIQUE;

DEFINE TABLE vuln_doc SCHEMAFULL; -- Extended vuln info for RAG
DEFINE FIELD cve_id ON TABLE vuln_doc TYPE string ASSERT $value != NONE;
DEFINE FIELD title ON TABLE vuln_doc TYPE string;
DEFINE FIELD summary ON TABLE vuln_doc TYPE string;
DEFINE FIELD cvss ON TABLE vuln_doc TYPE float;
DEFINE FIELD epss ON TABLE vuln_doc TYPE float; -- exploit prediction
DEFINE FIELD cpe ON TABLE vuln_doc TYPE array<string>;
DEFINE FIELD exploit_refs ON TABLE vuln_doc TYPE array<string>; -- URLs
DEFINE FIELD embedding ON TABLE vuln_doc TYPE array<float>; -- 1536 dims for OpenAI
DEFINE FIELD published_date ON TABLE vuln_doc TYPE datetime;
DEFINE INDEX idx_vuln_doc_cve ON TABLE vuln_doc COLUMNS cve_id UNIQUE;
DEFINE INDEX idx_vuln_doc_embedding ON TABLE vuln_doc COLUMNS embedding
  SEARCH embedding VECTOR COSINE;

-- Geography/taxonomy nodes
DEFINE TABLE city SCHEMAFULL;
DEFINE FIELD name ON TABLE city TYPE string;
DEFINE FIELD cc ON TABLE city TYPE string; -- country code
DEFINE FIELD lat ON TABLE city TYPE float;
DEFINE FIELD lon ON TABLE city TYPE float;
DEFINE INDEX idx_city_name ON TABLE city COLUMNS name;

DEFINE TABLE region SCHEMAFULL;
DEFINE FIELD name ON TABLE region TYPE string;
DEFINE FIELD cc ON TABLE region TYPE string;

DEFINE TABLE country SCHEMAFULL;
DEFINE FIELD cc ON TABLE country TYPE string ASSERT $value != NONE;
DEFINE FIELD name ON TABLE country TYPE string;
DEFINE INDEX idx_country_cc ON TABLE country COLUMNS cc UNIQUE;

DEFINE TABLE asn SCHEMAFULL;
DEFINE FIELD number ON TABLE asn TYPE int ASSERT $value > 0;
DEFINE FIELD org ON TABLE asn TYPE string;
DEFINE INDEX idx_asn_number ON TABLE asn COLUMNS number UNIQUE;

DEFINE TABLE cloud_region SCHEMAFULL;
DEFINE FIELD provider ON TABLE cloud_region TYPE string; -- 'aws', 'gcp', 'azure'
DEFINE FIELD code ON TABLE cloud_region TYPE string; -- 'us-east-1'
DEFINE FIELD name ON TABLE cloud_region TYPE string;
DEFINE INDEX idx_cloud_region_code ON TABLE cloud_region COLUMNS provider, code UNIQUE;

DEFINE TABLE common_port SCHEMAFULL;
DEFINE FIELD number ON TABLE common_port TYPE int;
DEFINE FIELD label ON TABLE common_port TYPE string; -- 'http', 'https', 'ssh'
DEFINE INDEX idx_common_port_number ON TABLE common_port COLUMNS number UNIQUE;

-- Edges (relationships)
DEFINE TABLE HAS SCHEMAFULL TYPE RELATION FROM host TO port;
DEFINE TABLE RUNS SCHEMAFULL TYPE RELATION FROM port TO service;
DEFINE TABLE EVIDENCED_BY SCHEMAFULL TYPE RELATION FROM service TO banner | tls_cert;
DEFINE TABLE AFFECTED_BY SCHEMAFULL TYPE RELATION FROM service TO vuln;
DEFINE TABLE IN_CITY SCHEMAFULL TYPE RELATION FROM host TO city;
DEFINE TABLE IN_REGION SCHEMAFULL TYPE RELATION FROM city TO region;
DEFINE TABLE IN_COUNTRY SCHEMAFULL TYPE RELATION FROM region TO country;
DEFINE TABLE IN_ASN SCHEMAFULL TYPE RELATION FROM host TO asn;
DEFINE TABLE IN_CLOUD_REGION SCHEMAFULL TYPE RELATION FROM host TO cloud_region;
DEFINE TABLE IS_COMMON SCHEMAFULL TYPE RELATION FROM port TO common_port;

DEFINE TABLE OBSERVED_AT SCHEMAFULL TYPE RELATION FROM service TO ANY;
DEFINE FIELD scan_id ON TABLE OBSERVED_AT TYPE string;
DEFINE FIELD contributor_id ON TABLE OBSERVED_AT TYPE string;
DEFINE FIELD ts ON TABLE OBSERVED_AT TYPE datetime DEFAULT time::now();
DEFINE FIELD trust ON TABLE OBSERVED_AT TYPE float DEFAULT 1.0;
```

### API Endpoints (REST)

| Endpoint | Method | Purpose | SLO | Auth | Tier |
|----------|--------|---------|-----|------|------|
| `/v0/mesh/ingest` | POST | Fast-path scan submission (direct DB write) | P95 ≤ 2s | JWT + Ed25519 | All |
| `/v0/ingest` | POST | Workflow-based scan submission (full enrichment) | Async | JWT + Ed25519 | All |
| `/v0/plan` | POST | Generate scan plan (stale targets) | P95 ≤ 700ms | JWT | All |
| `/v0/coverage` | GET | Freshness histograms and coverage stats | P95 ≤ 600ms | JWT | All |
| `/v0/host/{ip}` | GET | Get full host graph (ports, services, vulns) | P95 ≤ 600ms | JWT | All (vulns Pro) |
| `/v0/search` | GET | Query mesh (by city, ASN, service, port, etc.) | P95 ≤ 600ms | JWT | All |
| `/v0/ai/summary` | POST | AI-powered vulnerability summary (Pro) | P95 ≤ 4s | JWT + Pro | Pro only |
| `/v0/keys/rotate` | POST | Rotate Ed25519 contributor key | P95 ≤ 200ms | JWT | All |
| `/v0/keys/revoke` | POST | Revoke compromised key | P95 ≤ 200ms | JWT | All |

### Restate Workflows (Durable Execution)

**wf.plan_scan:**
```go
// Generate scan plan based on freshness
type PlanScanInput struct {
    Selectors map[string]string // city, asn, service, etc.
    MinAge    time.Duration     // e.g., 5 minutes
}

type PlanScanOutput struct {
    PlanID      string
    Targets     []Target
    TotalStale  int
    Pagination  PaginationInfo
}

// Workflow steps (all idempotent):
// 1. Query SurrealDB for stale hosts (age > MinAge)
// 2. Apply selectors (city, service, etc.)
// 3. Generate plan_id and store in Restate state
// 4. Return paginated target list (max 10k per page)
```

**wf.scan:**
```go
// Execute scan plan or raw envelope
type ScanInput struct {
    PlanID       string        // if resuming plan
    ScanEnvelope *ScanEnvelope // if raw submission
}

// Workflow steps:
// 1. Parse Naabu/Nmap output
// 2. Normalize to canonical format
// 3. Fingerprint assets
// 4. Return asset_set_ref for downstream workflows
```

**wf.enrich:**
```go
// Enrich scan results with external data
type EnrichInput struct {
    AssetSetRef string
}

// Workflow steps:
// 1. ASN lookup (MaxMind)
// 2. GeoIP lookup (city, region, country)
// 3. Cloud region detection (AWS/GCP/Azure ranges)
// 4. CPE mapping (service → CPE 2.3)
// 5. Tag common ports (80=http, 22=ssh, etc.)
// 6. (Pro) NVD/OSV/KEV vulnerability joins
// 7. Return enriched_set_ref
```

**wf.graph:**
```go
// Upsert nodes and edges into SurrealDB
type GraphInput struct {
    EnrichedSetRef string
}

// Workflow steps (all idempotent via observation_id):
// 1. Create/update HOST nodes
// 2. Create/update PORT nodes
// 3. Create/update SERVICE nodes
// 4. Create/update BANNER/TLS_CERT nodes
// 5. Create topology edges (HAS, RUNS, EVIDENCED_BY)
// 6. Create geo edges (IN_CITY, IN_REGION, IN_COUNTRY)
// 7. Create network edges (IN_ASN, IN_CLOUD_REGION)
// 8. Create OBSERVED_AT edges (with contributor metadata)
// 9. Update last_seen and last_scanned_at timestamps
// 10. Apply service-specific TTLs (HTTP 6h, SSH 24h, etc.)
```

**wf.ai_engine (Pro):**
```go
// AI-powered vulnerability summarization
type AIEngineInput struct {
    Selectors map[string]string
    Query     string
}

// Workflow steps:
// 1. Graph query to filter assets by selectors
// 2. Vector k-NN search (k=20) for relevant vuln_docs
// 3. Hybrid context assembly (graph + vector results)
// 4. Call OpenAI GPT-4 Turbo with timeout (3.5s)
// 5. Parse 3-bullet format (Summary, Risks, Next Steps)
// 6. Fallback to graph-only summary on timeout
// 7. Cache result in Restate state (24h TTL)
```

---

## 4. Implementation Roadmap (Solo Developer, 20 Weeks)

### Phase 1: Core Mesh Infrastructure (Weeks 1-8)

#### Week 1-2: Project Setup & Database
**Goal:** Get SurrealDB running with basic schema

- [ ] Set up Go workspace with modules (`go.work`, `go.mod`)
- [ ] Install SurrealDB locally (Docker or binary)
- [ ] Define complete schema (copy from section 3 above)
- [ ] Write seed data script (common ports, countries)
- [ ] Create basic CRUD operations (host, port, service)
- [ ] Write integration tests with Testcontainers

**Deliverable:** Working SurrealDB with schema, seed data, basic tests

---

#### Week 3-4: Mesh Ingest API (Fast Path)
**Goal:** Accept scan submissions and write to DB in <2s

- [ ] Create `svc.api` HTTP server with Chi router
- [ ] Implement POST `/v0/mesh/ingest` endpoint
- [ ] Parse Naabu JSON output format
- [ ] Implement Ed25519 signature verification
- [ ] Implement idempotent upserts (observation_id deduplication)
- [ ] Add rate limiting (60 req/min)
- [ ] Write unit tests + integration tests
- [ ] Benchmark: P95 latency < 2s for 100-host batch

**Deliverable:** Working ingest API, <2s P95 latency

---

#### Week 5-6: CLI Tool (Scan & Ingest)
**Goal:** Build `spectra` CLI for scanning and submission

- [ ] Create CLI with Cobra (`spectra` command)
- [ ] Implement `spectra scan` (calls Naabu, parses output)
- [ ] Implement Ed25519 key generation and envelope signing
- [ ] Implement `spectra scan --submit` (POST to `/v0/mesh/ingest`)
- [ ] Add local config file (`~/.spectra/config.yaml`)
- [ ] Store JWT tokens and contributor keys securely
- [ ] Write CLI integration tests
- [ ] Document installation and usage

**Deliverable:** Working CLI, can scan and submit to mesh

---

#### Week 7-8: Query API & Planning
**Goal:** Query mesh and generate scan plans

- [ ] Implement GET `/v0/search` with selectors
  - [ ] Filter by IP, city, ASN, service, port
  - [ ] Pagination (cursor-based)
  - [ ] Return up to 5k results per page
- [ ] Implement POST `/v0/plan` (stale target planning)
  - [ ] Query SurrealDB for age > min_age
  - [ ] Generate plan_id and store
  - [ ] Return paginated target list
- [ ] Implement GET `/v0/host/{ip}` (single host graph)
- [ ] Add CLI commands:
  - [ ] `spectra mesh query service redis --city Paris`
  - [ ] `spectra mesh plan city "Paris" --min-age 5m`
- [ ] Write query tests with mock data
- [ ] Benchmark: P95 < 600ms (search), P95 < 700ms (plan)

**Deliverable:** Working query API, CLI query commands, <700ms P95

---

### Phase 2: Enrichment & Vulnerability Correlation (Weeks 9-14)

#### Week 9-10: Restate Setup & Workflows
**Goal:** Get Restate running with basic workflows

- [ ] Install Restate server locally (Docker or binary)
- [ ] Configure Restate Go SDK in project
- [ ] Implement `wf.scan` workflow (parse scan envelope)
- [ ] Implement `wf.enrich` workflow skeleton (no external calls yet)
- [ ] Implement `wf.graph` workflow (upsert to SurrealDB)
- [ ] Connect `/v0/ingest` endpoint to trigger workflows
- [ ] Test workflow execution and resumability
- [ ] Add Restate state storage configuration (PostgreSQL or in-memory)

**Deliverable:** Working Restate workflows, async ingest path

---

#### Week 11-12: Enrichment Pipeline
**Goal:** Add ASN/GeoIP/CPE enrichment

- [ ] Download MaxMind GeoLite2 databases (ASN, City)
- [ ] Implement ASN lookup in `wf.enrich`
- [ ] Implement GeoIP lookup (city, region, country)
- [ ] Implement cloud region detection (AWS/GCP/Azure CIDR ranges)
- [ ] Implement CPE mapping (service version → CPE 2.3)
- [ ] Tag common ports (create common_port nodes)
- [ ] Create geo edges (IN_CITY, IN_REGION, etc.)
- [ ] Test enrichment with real scan data
- [ ] Benchmark: enrichment adds <5s to total pipeline

**Deliverable:** Full enrichment pipeline, geo/network context

---

#### Week 13-14: NVD Integration & Vulnerability Graph
**Goal:** Add Pro tier vulnerability correlation

- [ ] Create separate `spectra-vuln-ingest` Restate app
- [ ] Implement `wf.vuln_sync` (fetch NVD data via API)
- [ ] Parse NVD JSON and extract CVE, CVSS, CPE
- [ ] Create vuln_doc nodes in SurrealDB
- [ ] Implement CPE matching (service.cpe → vuln_doc.cpe)
- [ ] Create AFFECTED_BY edges (service → vuln)
- [ ] Add CISA KEV flag to vuln_doc
- [ ] Implement Pro tier gating (402 Payment Required)
- [ ] Test vulnerability correlation with known CVEs
- [ ] Document Pro tier activation process (manual for MVP)

**Deliverable:** NVD sync, vulnerability edges, Pro tier gating

---

### Phase 3: AI & Pro Features (Weeks 15-20)

#### Week 15-16: Vector Embeddings & RAG
**Goal:** Implement hybrid graph+vector search

- [ ] Create OpenAI account and get API key
- [ ] Implement `wf.vuln_vectorize` workflow
  - [ ] Generate text from vuln_doc (title + summary + CPE)
  - [ ] Call OpenAI embeddings API (text-embedding-3-small, 1536 dims)
  - [ ] Store in vuln_doc.embedding field
- [ ] Configure SurrealDB vector index (cosine similarity)
- [ ] Implement hybrid retrieval in `wf.ai_engine`:
  - [ ] Graph query for asset filtering
  - [ ] Vector k-NN search (k=20)
  - [ ] Merge results by relevance
- [ ] Test vector search with sample queries
- [ ] Benchmark: k-NN P95 < 250ms

**Deliverable:** Working vector RAG, hybrid search

---

#### Week 17-18: AI Summarization
**Goal:** GPT-4 powered vulnerability summaries

- [ ] Implement `wf.ai_engine` workflow
  - [ ] Assemble hybrid context (graph + vector, max 8KB)
  - [ ] Call OpenAI GPT-4 Turbo API
  - [ ] Parse 3-bullet format (Summary, Risks, Next Steps)
  - [ ] Implement timeout handling (3.5s)
  - [ ] Implement fallback to graph-only summary
- [ ] Implement POST `/v0/ai/summary` endpoint
- [ ] Add CLI command: `spectra explain ip 1.2.3.4`
- [ ] Implement result caching (24h TTL in Restate state)
- [ ] Test with various query types
- [ ] Benchmark: P95 < 4s (or fallback)

**Deliverable:** AI-powered summaries, CLI explain command

---

#### Week 19-20: Polish, Testing, Documentation
**Goal:** Production-ready MVP

- [ ] End-to-end testing:
  - [ ] Full scan → ingest → enrich → graph → query flow
  - [ ] Pro tier vulnerability correlation
  - [ ] AI summarization with real CVEs
- [ ] Performance optimization:
  - [ ] Index tuning in SurrealDB
  - [ ] Batch operation optimization
  - [ ] Connection pool configuration
- [ ] Monitoring setup:
  - [ ] Prometheus metrics for all APIs
  - [ ] Grafana dashboard (optional)
  - [ ] Structured logging with Zap
- [ ] Documentation:
  - [ ] Installation guide (README)
  - [ ] CLI usage examples
  - [ ] API documentation (OpenAPI spec)
  - [ ] Architecture diagrams
  - [ ] Contribution guide for community runners
- [ ] Security hardening:
  - [ ] Rate limiting tuning
  - [ ] JWT expiration policies
  - [ ] Ed25519 key rotation testing
  - [ ] Do-not-scan list enforcement
- [ ] Beta testing:
  - [ ] Recruit 10-20 community beta testers
  - [ ] Set up feedback channel (GitHub Discussions)
  - [ ] Create sample scan targets for testing

**Deliverable:** Production-ready MVP, documentation, beta program

---

## 5. Acceptance Criteria & Testing

### Functional Acceptance Criteria

**AC1: Scan Submission (P0)**
- GIVEN a Naabu scan of 100 hosts
- WHEN submitted via CLI `spectra scan --submit`
- THEN results are visible in query API within 2 seconds (P95)
- AND duplicate submissions are idempotent (same observation_id)

**AC2: Freshness Planning (P0)**
- GIVEN 10,000 hosts in database with varying last_seen timestamps
- WHEN planning with `spectra mesh plan city "Paris" --min-age 5m`
- THEN only hosts with age > 5 minutes are returned
- AND response time is < 700ms (P95)
- AND pagination works for >10k results

**AC3: Query API (P0)**
- GIVEN hosts in Paris running Redis on port 6379
- WHEN querying `spectra mesh query service redis --city Paris`
- THEN all matching hosts are returned
- AND response time is < 600ms (P95)
- AND results include freshness metadata

**AC4: Vulnerability Correlation (P1)**
- GIVEN a service with known CPE (e.g., nginx 1.25.0)
- WHEN enrichment runs with Pro tier
- THEN AFFECTED_BY edges are created to matching CVEs
- AND CVEs include CVSS scores and KEV flags

**AC5: AI Summarization (P2)**
- GIVEN a host with multiple vulnerable services
- WHEN requesting `spectra explain ip 1.2.3.4` with Pro tier
- THEN AI summary returns in < 4s (P95)
- AND summary includes 3 bullets (Summary, Risks, Next Steps)
- AND fallback works if AI timeout occurs

### Performance SLOs

| Metric | Target | Measurement |
|--------|--------|-------------|
| Ingest visibility (fast path) | P95 ≤ 2s | Time from POST to query result |
| Planning query (city-level) | P95 ≤ 700ms | POST /v0/plan response time |
| Search query (≤5k results) | P95 ≤ 600ms | GET /v0/search response time |
| Vector k-NN (k=20) | P95 ≤ 250ms | Embedding similarity search |
| AI summary | P95 ≤ 4s | Full workflow including fallback |
| Throughput (ingest) | 100+ req/min | Sustained ingest rate |

### Test Coverage Goals

- **Unit tests:** >80% coverage for business logic
- **Integration tests:** All API endpoints, all workflows
- **E2E tests:** Complete scan → query flow
- **Performance tests:** Load test to validate SLOs
- **Security tests:** JWT validation, signature verification, rate limiting

---

## 6. MVP Success Metrics (3-6 Months Post-Launch)

### User Adoption
- **Community runners:** 50-100 active contributors
- **Beta users:** 500-1,000 registered users
- **Daily queries:** 5,000-10,000 API calls
- **Data coverage:** 1M+ hosts in mesh

### Technical Health
- **Uptime:** >99% availability
- **SLO compliance:** >95% of requests meet latency targets
- **Data freshness:** P95 scan age < 24 hours
- **Error rate:** <1% 5xx responses

### Business Validation
- **Pro tier conversions:** 5-10 paying users @ $50-100/month
- **User retention:** >60% monthly active users (MAU) retained
- **Community engagement:** 20+ GitHub stars, 5+ contributors
- **Feedback quality:** 80%+ positive sentiment in beta feedback

---

## 7. Risks & Mitigation

### Technical Risks

**Risk 1: SurrealDB Production Maturity**
- **Likelihood:** Medium
- **Impact:** High (database is critical path)
- **Mitigation:**
  - Run extensive load testing early (Week 3-4)
  - Have PostgreSQL+PostGIS fallback plan
  - Engage with SurrealDB community for support
  - Regular backups with restore testing

**Risk 2: Restate Learning Curve**
- **Likelihood:** Medium
- **Impact:** Medium (can delay Phase 2)
- **Mitigation:**
  - Allocate 1 week for Restate learning (before Week 9)
  - Start with simple workflows, add complexity gradually
  - Use Restate Discord for support
  - Fallback: Direct DB writes without workflows (less resilient)

**Risk 3: OpenAI API Costs**
- **Likelihood:** Low-Medium
- **Impact:** Medium (could exceed budget)
- **Mitigation:**
  - Cache all AI responses (24h TTL)
  - Rate limit AI endpoint aggressively (20 req/min)
  - Monitor costs daily with alerts
  - Consider local LLM fallback (Llama 3, Mistral)

**Risk 4: Community Adoption**
- **Likelihood:** Medium
- **Impact:** High (mesh needs data contributions)
- **Mitigation:**
  - Clear value prop: "contribute scans, get Pro tier free"
  - Easy onboarding (single CLI command)
  - Engage security communities (Reddit, Twitter, conferences)
  - Provide sample targets for easy testing
  - Recognition system (contributor leaderboard)

### Operational Risks

**Risk 5: ISP Blocking of Scans**
- **Likelihood:** High
- **Impact:** Medium (expected, plan for it)
- **Mitigation:**
  - Implement relay rotation (3-5 providers)
  - Backoff and retry logic
  - Do-not-scan list enforcement
  - Rate limiting per target network
  - Clear legal/ethical guidelines

**Risk 6: Abuse/Malicious Scans**
- **Likelihood:** Medium
- **Impact:** High (legal/reputation risk)
- **Mitigation:**
  - Require Ed25519 signatures (contributor accountability)
  - IP/ASN blocklist
  - Do-not-scan list (IANA, ECAD, manual)
  - Manual review of Pro tier activations
  - Legal consent checkbox during signup
  - Abuse reporting mechanism

---

## 8. Open Questions & Decisions Needed

### Architecture Decisions

**Q1: Restate State Storage**
- **Question:** In-memory (fast, volatile) or PostgreSQL (durable, slower)?
- **Recommendation:** Start with in-memory for MVP, migrate to PostgreSQL when hitting limits
- **Timeline:** Decide by Week 9

**Q2: Vector Model Selection**
- **Question:** OpenAI embeddings ($0.13/1M tokens) or local Sentence Transformers (free, slower)?
- **Recommendation:** OpenAI for MVP (simple API), evaluate local models if costs exceed $100/month
- **Timeline:** Decide by Week 15

**Q3: JWT Issuer**
- **Question:** Self-hosted (simple, less secure) or Auth0/Clerk (complex, more secure)?
- **Recommendation:** Self-hosted JWT for MVP, revisit for post-MVP security hardening
- **Timeline:** Implement in Week 4

### Product Decisions

**Q4: Pro Tier Pricing**
- **Question:** $50/month or $100/month for individual users?
- **Recommendation:** Start at $50/month, survey beta users for willingness to pay
- **Timeline:** Decide by Week 14 (before Pro launch)

**Q5: Do-Not-Scan List Sources**
- **Question:** Which lists to enforce? (IANA reserved, ECAD, manual submissions)
- **Recommendation:** IANA reserved (must have) + manual submissions, defer ECAD for complexity
- **Timeline:** Implement in Week 7

**Q6: Relay Provider Selection**
- **Question:** Hetzner, DigitalOcean, Linode, AWS?
- **Recommendation:** Start with Hetzner (cost-effective, permissive ToS), add others if needed
- **Timeline:** Set up first relay by Week 5

---

## 9. Dependencies & Prerequisites

### External Dependencies (Before Week 1)

- [ ] **Naabu:** Install locally (`go install github.com/projectdiscovery/naabu/v2/cmd/naabu@latest`)
- [ ] **SurrealDB:** Download binary or Docker image (https://surrealdb.com/install)
- [ ] **Restate:** Download binary or Docker image (https://docs.restate.dev/deploy/overview)
- [ ] **MaxMind Account:** Sign up for free GeoLite2 account (https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [ ] **OpenAI Account:** Create account and get API key (https://platform.openai.com/)
- [ ] **NVD API Key:** Request free API key (https://nvd.nist.gov/developers/request-an-api-key)

### Development Environment Setup

- [ ] **Go 1.23+:** Install latest Go (https://go.dev/dl/)
- [ ] **Docker:** For running SurrealDB, Restate, Testcontainers
- [ ] **Git:** For version control
- [ ] **IDE/Editor:** VS Code with Go extension recommended
- [ ] **HTTP Client:** curl, Postman, or Insomnia for API testing

### Infrastructure (MVP can run on single machine)

**Minimum Specs:**
- **CPU:** 4 cores
- **RAM:** 16 GB
- **Disk:** 100 GB SSD
- **Network:** 100 Mbps+

**Recommended Cloud Provider Options:**
- **Hetzner CPX31:** €13.90/month (4 vCPU, 8 GB RAM) - for Restate + SurrealDB
- **Hetzner CPX21:** €6.90/month (3 vCPU, 4 GB RAM) - for API server
- **Total:** ~€20/month (~$22 USD/month)

**Or run locally for MVP** (sufficient for beta testing)

---

## 10. Post-MVP Roadmap (Months 6-12)

### Phase 4: Scale & Enterprise Features

**Months 6-8:**
- Multi-node Restate cluster (HA)
- SurrealDB clustering/replication
- Web dashboard (React + Tailwind)
- User management UI
- Billing integration (Stripe)
- Email notifications/alerts

**Months 9-10:**
- SIEM integrations (Splunk, Elastic)
- Continuous monitoring (scheduled scans)
- Slack/Discord webhooks
- Advanced filtering and saved searches
- Historical trend analysis

**Months 11-12:**
- Enterprise tier features:
  - Read-only database mirrors
  - SSO/SAML integration
  - Custom TTL policies
  - Priority relay access
  - Dedicated support
- SOC2 compliance preparation
- Scale to 10M+ hosts in mesh

---

## Appendix A: Go Dependency List

### Core Dependencies (13 total)

```go
// go.mod for main application
module github.com/spectra-red/recon

go 1.23

require (
    // Orchestration & workflows
    github.com/restatedev/sdk-go v0.3.0

    // Database
    github.com/surrealdb/surrealdb.go v1.0.0

    // HTTP & API
    github.com/go-chi/chi/v5 v5.0.11

    // CLI
    github.com/spf13/cobra v1.8.0

    // Authentication
    github.com/golang-jwt/jwt/v5 v5.0.0

    // AI & embeddings
    github.com/sashabaranov/go-openai v1.24.0

    // Networking
    github.com/seancfoley/ipaddress-go v1.6.0

    // Config & data
    github.com/go-yaml/yaml v3.0.1

    // Logging
    go.uber.org/zap v1.26.0

    // Metrics
    github.com/prometheus/client_golang v1.17.0

    // Testing
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/testcontainers/testcontainers-go v0.27.0
)
```

### External Tools (not Go packages)

- **Naabu:** Port scanner binary
- **Nmap:** Service detection (optional)
- **MaxMind GeoLite2:** Database files (not a package)

**Total estimated bundle size:** ~30 MB compiled binary

---

## Appendix B: CLI Command Reference

```bash
# Authentication & setup
spectra auth login                    # OAuth2 login flow
spectra auth logout                   # Clear credentials
spectra keys generate                 # Generate Ed25519 key pair
spectra keys rotate                   # Rotate contributor key

# Scanning
spectra scan <target>                 # Scan target (IP, CIDR, hostname)
spectra scan --plan <plan_id>         # Scan from plan
spectra scan --submit                 # Submit results to mesh
spectra scan --privacy meta-only      # Don't include full banners

# Mesh operations
spectra mesh plan <selectors>         # Generate scan plan
  --min-age 5m                        # Only stale targets
  --service redis                     # Filter by service
  --city "Paris"                      # Filter by city
  --common-port 6379                  # Filter by common port

spectra mesh query <selectors>        # Query mesh
  --service http                      # Find HTTP services
  --city "London"                     # In specific city
  --since 7d                          # Seen in last 7 days
  --output json                       # JSON output

spectra coverage <selectors>          # Coverage stats
  --city "Tokyo"                      # Freshness histogram

# Pro features
spectra explain ip <ip>               # AI vulnerability summary
spectra explain host <hostname>       # Same, by hostname

# Configuration
spectra config show                   # Show current config
spectra config set <key> <value>      # Update config
```

---

## Appendix C: Sample Queries (SurrealDB)

```sql
-- Query 1: Find all Redis instances in Paris seen in last 7 days
SELECT
  host.ip,
  port.number,
  service.product,
  service.version,
  host.last_seen
FROM host
WHERE host->IN_CITY->city.name = "Paris"
  AND host->HAS->port->RUNS->service.name = "redis"
  AND host.last_seen > time::now() - 7d
ORDER BY host.last_seen DESC;

-- Query 2: Find vulnerable services with critical CVEs
SELECT
  host.ip,
  service.product,
  service.version,
  vuln.cve_id,
  vuln.cvss,
  vuln.severity,
  vuln.kev_flag
FROM service
  ->AFFECTED_BY->vuln
WHERE vuln.severity = "critical"
  AND vuln.cvss >= 9.0
FETCH host, service, vuln;

-- Query 3: Freshness histogram for Paris (planning query)
SELECT
  count() AS stale_count,
  math::floor((time::now() - host.last_seen) / 1h) AS hours_old
FROM host
WHERE host->IN_CITY->city.name = "Paris"
GROUP BY hours_old
ORDER BY hours_old ASC;

-- Query 4: Vector similarity search for nginx vulnerabilities
SELECT
  cve_id,
  title,
  cvss,
  vector::similarity::cosine(
    embedding,
    $query_embedding
  ) AS similarity
FROM vuln_doc
WHERE similarity > 0.8
ORDER BY similarity DESC
LIMIT 20;

-- Query 5: Host graph with all relationships
SELECT
  *,
  ->HAS->port.* AS ports,
  ports->RUNS->service.* AS services,
  services->AFFECTED_BY->vuln.* AS vulnerabilities,
  ->IN_CITY->city.* AS city,
  ->IN_ASN->asn.* AS asn
FROM host
WHERE ip = $target_ip
FETCH ports, services, vulnerabilities, city, asn;
```

---

## Appendix D: Useful Resources

### Official Documentation
- **Restate:** https://docs.restate.dev/
- **SurrealDB:** https://docs.surrealdb.com/
- **Naabu:** https://github.com/projectdiscovery/naabu
- **OpenAI API:** https://platform.openai.com/docs/

### Learning Resources
- **Restate Go Examples:** https://github.com/restatedev/examples/tree/main/go
- **SurrealDB Go SDK Guide:** https://surrealdb.com/docs/sdk/go
- **Chi Router Tutorial:** https://go-chi.io/#/README
- **OAuth2 in Go:** https://pkg.go.dev/golang.org/x/oauth2

### Community
- **Restate Discord:** https://discord.gg/restate
- **SurrealDB Discord:** https://discord.gg/surrealdb
- **r/netsec:** https://reddit.com/r/netsec (user feedback)
- **ProjectDiscovery Discord:** https://discord.gg/projectdiscovery

### Tools & Libraries
- **Postman:** https://postman.com (API testing)
- **k6:** https://k6.io (load testing)
- **Grafana:** https://grafana.com (monitoring)
- **Prometheus:** https://prometheus.io (metrics)

---

**END OF ENGINEERING-FOCUSED PRD**

*This document is ready for implementation. Start with Phase 1, Week 1-2 (SurrealDB setup) and follow the roadmap sequentially. Good luck building!*
