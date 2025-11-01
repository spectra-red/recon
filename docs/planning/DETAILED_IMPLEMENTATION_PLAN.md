# Spectra-Red Implementation Plan
**From PRD to Production-Ready Code**

Version: 1.0
Date: November 1, 2025
Timeline: 20 weeks to production MVP

---

## Executive Summary

This document transforms the Spectra-Red PRD into actionable implementation tasks. Each task is:
- **Atomic**: 1-4 hours of work
- **Testable**: Clear verification criteria
- **Sequenced**: Dependencies managed, risk-first approach
- **Specific**: Exact files, functions, and patterns

**Key Architecture Decisions:**
- Monorepo with Go workspaces
- SurrealDB for graph+vector storage
- Restate for durable workflow orchestration
- Ed25519 + JWT for authentication
- Chi router for HTTP APIs

---

## 1. Technical Architecture

### 1.1 System Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Community Runners                         │
│              (Naabu/Nmap → CLI → HTTPS)                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  API Gateway (Go + Chi)                                     │
│  - JWT validation                                           │
│  - Ed25519 signature verification                          │
│  - Rate limiting (60/min ingest, 30/min query)            │
│  - Request routing                                          │
└──────────────┬──────────────────────────────────────────────┘
               │
    ┌──────────┼──────────┐
    │          │          │
    ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────────────┐
│ Fast   │ │Workflow│ │ Query Engine   │
│ Path   │ │ Async  │ │ (SurrealDB)    │
│ <2s    │ │ Path   │ │ <600ms         │
└────┬───┘ └───┬────┘ └────┬───────────┘
     │         │            │
     ▼         ▼            ▼
┌─────────────────────────────────────────┐
│          SurrealDB Cluster              │
│  - Graph: hosts, ports, services        │
│  - Vector: vuln embeddings (1536 dims)  │
│  - Temporal: observation history        │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│       Restate Workflows                 │
│  wf.scan    - Parse & normalize         │
│  wf.enrich  - ASN, GeoIP, CPE           │
│  wf.graph   - Upsert nodes/edges        │
│  wf.ai      - Vector RAG + GPT-4        │
└─────────────────────────────────────────┘
```

### 1.2 Repository Structure

```
spectra-red/
├── go.work                    # Go workspace definition
├── go.mod                     # Root module
├── cmd/
│   ├── api/                   # HTTP API server
│   │   └── main.go
│   ├── cli/                   # spectra CLI tool
│   │   └── main.go
│   └── workflows/             # Restate workflow service
│       └── main.go
├── internal/
│   ├── api/                   # HTTP handlers
│   │   ├── middleware/        # Auth, rate limiting
│   │   ├── handlers/          # Endpoint handlers
│   │   └── routes.go
│   ├── db/                    # Database layer
│   │   ├── surrealdb/         # SurrealDB client
│   │   ├── schema/            # Schema definitions
│   │   └── queries/           # Query builders
│   ├── workflows/             # Restate workflows
│   │   ├── scan.go
│   │   ├── enrich.go
│   │   ├── graph.go
│   │   └── ai.go
│   ├── scanner/               # Scan execution
│   │   ├── naabu.go
│   │   └── parser.go
│   ├── auth/                  # Authentication
│   │   ├── jwt.go
│   │   └── ed25519.go
│   └── models/                # Domain models
│       ├── scan.go
│       ├── host.go
│       └── vuln.go
├── pkg/                       # Public libraries
│   ├── client/                # Go SDK
│   └── types/                 # Shared types
├── scripts/
│   ├── setup-db.sh
│   └── seed-data.sh
├── deployments/
│   ├── docker-compose.yml
│   └── k8s/
└── docs/
    ├── architecture.md
    └── api.yaml            # OpenAPI spec
```

### 1.3 Data Flow Patterns

**Pattern 1: Fast Path Ingest (P95 < 2s)**
```
CLI scan → POST /v0/mesh/ingest → API validates
  → Direct SurrealDB write → 200 OK
```

**Pattern 2: Enrichment Path (Async)**
```
CLI scan → POST /v0/ingest → Trigger wf.scan
  → wf.enrich (ASN, GeoIP, CPE)
  → wf.graph (upsert nodes/edges)
  → Observation stored
```

**Pattern 3: Query Path (P95 < 600ms)**
```
GET /v0/search?city=Paris&service=redis
  → SurrealDB graph query
  → Filter by selectors
  → Return JSON
```

**Pattern 4: AI Analysis (Pro, P95 < 4s)**
```
POST /v0/ai/summary {ip: "1.2.3.4"}
  → wf.ai: Graph query (services, vulns)
  → Vector k-NN (k=20 similar vulns)
  → GPT-4 summarization
  → Return 3-bullet analysis
```

---

## 2. Milestone-Based Task Breakdown

### MILESTONE 1: Foundation (Weeks 1-2)
**Goal:** Development environment ready, basic project structure

#### M1-T1: Project Initialization (2 hours)
**Files:**
- `go.work`, `go.mod`
- `cmd/api/main.go` (skeleton)
- `internal/api/routes.go` (skeleton)

**Steps:**
1. Initialize Go workspace: `go work init`
2. Create module: `go mod init github.com/spectra-red/recon`
3. Add dependencies:
   ```
   github.com/restatedev/sdk-go v0.3.0
   github.com/surrealdb/surrealdb.go v1.0.0
   github.com/go-chi/chi/v5 v5.0.11
   github.com/spf13/cobra v1.8.0
   ```
4. Create directory structure
5. Add `.gitignore`, `README.md`, `LICENSE`

**Acceptance:**
- `go build ./...` succeeds
- Module graph validated
- Git repository initialized

---

#### M1-T2: Docker Compose Setup (3 hours)
**Files:**
- `deployments/docker-compose.yml`
- `deployments/Dockerfile.api`
- `scripts/setup-db.sh`

**Docker Services:**
```yaml
services:
  surrealdb:
    image: surrealdb/surrealdb:latest
    ports: ["8000:8000"]
    command: start --log trace --user root --pass root memory

  restate:
    image: restatedev/restate:latest
    ports: ["8080:8080", "9070:9070"]

  api:
    build: .
    ports: ["3000:3000"]
    depends_on: [surrealdb, restate]
```

**Acceptance:**
- `docker-compose up` starts all services
- SurrealDB accessible at `localhost:8000`
- Restate UI at `localhost:9070`

---

#### M1-T3: SurrealDB Schema Definition (4 hours)
**Files:**
- `internal/db/schema/schema.surql`
- `internal/db/schema/seed.surql`
- `scripts/setup-db.sh`

**Schema (from PRD section 3):**
```sql
-- Core tables
DEFINE TABLE host SCHEMAFULL;
DEFINE FIELD ip ON TABLE host TYPE string ASSERT $value != NONE;
DEFINE FIELD asn ON TABLE host TYPE int;
DEFINE FIELD city ON TABLE host TYPE string;
DEFINE FIELD last_seen ON TABLE host TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_host_ip ON TABLE host COLUMNS ip UNIQUE;

DEFINE TABLE port SCHEMAFULL;
DEFINE FIELD number ON TABLE port TYPE int ASSERT $value > 0 AND $value < 65536;
DEFINE FIELD protocol ON TABLE port TYPE string ASSERT $value IN ['tcp', 'udp'];
DEFINE INDEX idx_port_number ON TABLE port COLUMNS number;

-- See PRD for complete schema
```

**Seed Data:**
- Common ports (22, 80, 443, 3306, 6379)
- Countries (from MaxMind)
- ASN data (top 1000)

**Acceptance:**
- Schema applies without errors
- Seed data loads (100+ records)
- Indices created successfully

---

#### M1-T4: Basic HTTP Server (3 hours)
**Files:**
- `cmd/api/main.go`
- `internal/api/routes.go`
- `internal/api/middleware/logging.go`

**Code:**
```go
// cmd/api/main.go
func main() {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Get("/health", healthHandler)
    r.Route("/v0", func(r chi.Router) {
        // Routes defined in internal/api/routes.go
    })

    http.ListenAndServe(":3000", r)
}
```

**Acceptance:**
- Server starts on port 3000
- `GET /health` returns 200
- Structured logging works

---

### MILESTONE 2: Ingest Path (Weeks 3-4)
**Goal:** Accept scan submissions, write to DB in <2s

#### M2-T1: Ed25519 Signature Verification (4 hours)
**Files:**
- `internal/auth/ed25519.go`
- `internal/auth/ed25519_test.go`
- `internal/api/middleware/verify_signature.go`

**Implementation:**
```go
type ScanEnvelope struct {
    Data      json.RawMessage `json:"data"`
    PublicKey string          `json:"public_key"`
    Signature string          `json:"signature"`
    Timestamp int64           `json:"timestamp"`
}

func VerifyEnvelope(env ScanEnvelope) error {
    // 1. Check timestamp freshness (±5 min)
    // 2. Decode public key and signature
    // 3. Verify signature: ed25519.Verify(pubKey, data, sig)
    // 4. Return error if invalid
}
```

**Test Cases:**
- Valid signature → passes
- Expired timestamp → fails
- Invalid signature → fails
- Tampered data → fails

**Acceptance:**
- All tests pass
- Middleware rejects invalid signatures

---

#### M2-T2: Fast Path Ingest Endpoint (4 hours)
**Files:**
- `internal/api/handlers/ingest.go`
- `internal/db/surrealdb/client.go`
- `internal/db/surrealdb/upsert.go`

**Endpoint:**
```go
POST /v0/mesh/ingest
{
  "data": {...},      // Naabu scan results
  "public_key": "...",
  "signature": "...",
  "timestamp": 1730419200
}
```

**Handler Logic:**
1. Verify signature (middleware)
2. Parse Naabu JSON
3. Normalize to canonical format
4. Idempotent upsert (by observation_id)
5. Return 200 OK

**SurrealDB Operations:**
```go
// Upsert host
CREATE host:$ip CONTENT {
    ip: $ip,
    asn: $asn,
    last_seen: time::now()
} ON DUPLICATE KEY UPDATE last_seen = time::now();

// Create HAS edge
RELATE host:$ip->HAS->port:$port_id
```

**Performance Target:** P95 < 2s for 100-host batch

**Acceptance:**
- Submit 100 hosts → visible in DB within 2s
- Duplicate submissions are idempotent
- Rate limiting enforced (60 req/min)

---

#### M2-T3: Naabu Output Parser (3 hours)
**Files:**
- `internal/scanner/naabu_parser.go`
- `internal/scanner/naabu_parser_test.go`

**Input Format (Naabu JSON):**
```json
{"host":"1.2.3.4","port":80,"protocol":"tcp"}
{"host":"1.2.3.4","port":443,"protocol":"tcp"}
```

**Output Format (Canonical):**
```go
type ScanResult struct {
    Hosts []Host `json:"hosts"`
}

type Host struct {
    IP    string `json:"ip"`
    Ports []Port `json:"ports"`
}

type Port struct {
    Number   int    `json:"number"`
    Protocol string `json:"protocol"`
}
```

**Test Cases:**
- Valid Naabu output → parses correctly
- Malformed JSON → error
- Empty input → empty result

**Acceptance:**
- All tests pass
- Handles 10K+ line Naabu output

---

### MILESTONE 3: Query API (Weeks 5-6)
**Goal:** Query mesh by IP, city, service, port

#### M3-T1: Search Endpoint (4 hours)
**Files:**
- `internal/api/handlers/search.go`
- `internal/db/queries/search.go`

**Endpoint:**
```
GET /v0/search?city=Paris&service=redis&port=6379
```

**Query Builder:**
```go
func BuildSearchQuery(params SearchParams) string {
    q := "SELECT * FROM host"

    var filters []string
    if params.City != "" {
        filters = append(filters, "host->IN_CITY->city.name = $city")
    }
    if params.Service != "" {
        filters = append(filters, "host->HAS->port->RUNS->service.name = $service")
    }

    if len(filters) > 0 {
        q += " WHERE " + strings.Join(filters, " AND ")
    }

    return q + " LIMIT $limit"
}
```

**Acceptance:**
- Query by city → correct results
- Query by service → correct results
- Combined filters work
- P95 latency < 600ms

---

#### M3-T2: Host Detail Endpoint (2 hours)
**Files:**
- `internal/api/handlers/host.go`

**Endpoint:**
```
GET /v0/host/1.2.3.4
```

**Response:**
```json
{
  "ip": "1.2.3.4",
  "asn": 15169,
  "city": "Paris",
  "ports": [
    {
      "number": 80,
      "service": {
        "name": "http",
        "product": "nginx",
        "version": "1.25.1"
      }
    }
  ],
  "last_seen": "2025-11-01T12:00:00Z"
}
```

**Query:**
```sql
SELECT
    *,
    ->HAS->port.* AS ports,
    ports->RUNS->service.* AS services,
    ->IN_CITY->city.* AS city
FROM host
WHERE ip = $ip
FETCH ports, services, city;
```

**Acceptance:**
- Returns full host graph
- P95 < 600ms

---

#### M3-T3: Planning Endpoint (4 hours)
**Files:**
- `internal/api/handlers/plan.go`
- `internal/db/queries/stale.go`

**Endpoint:**
```
POST /v0/plan
{
  "selectors": {"city": "Paris"},
  "min_age": "5m"
}
```

**Response:**
```json
{
  "plan_id": "plan_abc123",
  "targets": ["1.2.3.4", "5.6.7.8"],
  "total_stale": 2,
  "pagination": {"next": "cursor_xyz"}
}
```

**Query (find stale hosts):**
```sql
SELECT ip FROM host
WHERE host->IN_CITY->city.name = $city
  AND time::now() - last_seen > $min_age
ORDER BY last_seen ASC
LIMIT 10000;
```

**Acceptance:**
- Only returns hosts older than min_age
- Pagination works for >10K results
- P95 < 700ms

---

### MILESTONE 4: CLI Tool (Weeks 7-8)
**Goal:** `spectra` CLI for scanning and querying

#### M4-T1: CLI Skeleton with Cobra (2 hours)
**Files:**
- `cmd/cli/main.go`
- `cmd/cli/cmd/root.go`
- `cmd/cli/cmd/scan.go`

**Commands:**
```bash
spectra scan <target>
spectra mesh plan <selectors>
spectra mesh query <selectors>
spectra auth login
```

**Structure:**
```go
// cmd/cli/cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "spectra",
    Short: "Spectra-Red security intelligence CLI",
}

// cmd/cli/cmd/scan.go
var scanCmd = &cobra.Command{
    Use:   "scan <target>",
    Short: "Scan target and optionally submit to mesh",
    Run:   runScan,
}
```

**Acceptance:**
- `spectra --help` shows commands
- Commands parse flags correctly

---

#### M4-T2: Scan Command Implementation (4 hours)
**Files:**
- `cmd/cli/cmd/scan.go`
- `internal/scanner/naabu.go`
- `internal/scanner/executor.go`

**Implementation:**
```go
func runScan(cmd *cobra.Command, args []string) {
    target := args[0]
    submit := cmd.Flags().GetBool("submit")

    // 1. Execute Naabu
    results := executeNaabu(target)

    // 2. Parse output
    parsed := parseNaabuOutput(results)

    // 3. If --submit, sign and send
    if submit {
        envelope := signEnvelope(parsed, privateKey)
        submitToMesh(envelope)
    } else {
        printResults(parsed)
    }
}

func executeNaabu(target string) ([]byte, error) {
    cmd := exec.Command("naabu", "-host", target, "-json")
    return cmd.Output()
}
```

**Acceptance:**
- `spectra scan 192.168.1.1` runs Naabu
- `spectra scan 192.168.1.1 --submit` posts to API
- Results displayed in clean format

---

#### M4-T3: Key Management (3 hours)
**Files:**
- `cmd/cli/cmd/keys.go`
- `internal/auth/keystore.go`

**Commands:**
```bash
spectra keys generate
spectra keys rotate
```

**Key Storage:**
```
~/.spectra/
  config.yaml          # API endpoint, settings
  keys/
    contributor.key    # Ed25519 private key
    contributor.pub    # Ed25519 public key
```

**Implementation:**
```go
func generateKeys() {
    pub, priv, _ := ed25519.GenerateKey(nil)

    // Store in ~/.spectra/keys/
    os.WriteFile("~/.spectra/keys/contributor.key", priv, 0600)
    os.WriteFile("~/.spectra/keys/contributor.pub", pub, 0644)
}
```

**Acceptance:**
- Keys stored with correct permissions (0600)
- Public key can be shared
- Private key never transmitted

---

### MILESTONE 5: Restate Workflows (Weeks 9-10)
**Goal:** Durable enrichment pipeline

#### M5-T1: Restate Server Setup (2 hours)
**Files:**
- `deployments/docker-compose.yml` (update)
- `cmd/workflows/main.go`

**Restate Service Registration:**
```go
func main() {
    server := restate.NewServer().
        Bind(restate.Reflect(&ScanWorkflow{})).
        Bind(restate.Reflect(&EnrichWorkflow{})).
        Bind(restate.Reflect(&GraphWorkflow{}))

    http.ListenAndServe(":9080", server)
}
```

**Acceptance:**
- Workflows register with Restate
- Restate UI shows services

---

#### M5-T2: wf.scan Implementation (4 hours)
**Files:**
- `internal/workflows/scan.go`
- `internal/workflows/scan_test.go`

**Workflow:**
```go
type ScanWorkflow struct{}

type ScanInput struct {
    ScanEnvelope ScanEnvelope
}

type ScanOutput struct {
    AssetSetRef string
}

func (w *ScanWorkflow) Run(
    ctx restate.WorkflowContext,
    input ScanInput,
) (ScanOutput, error) {
    // Step 1: Parse envelope
    parsed, _ := restate.Run(ctx, func(ctx restate.RunContext) (ScanResult, error) {
        return parseNaabuOutput(input.ScanEnvelope.Data)
    })

    // Step 2: Normalize
    normalized := normalize(parsed)

    // Step 3: Fingerprint assets
    assetSetRef := generateAssetSetRef(normalized)

    // Step 4: Store in Restate state
    restate.Set(ctx, "assets", normalized)

    return ScanOutput{AssetSetRef: assetSetRef}, nil
}
```

**Acceptance:**
- Workflow executes successfully
- Survives restarts (durability test)
- State persisted correctly

---

#### M5-T3: wf.enrich Implementation (4 hours)
**Files:**
- `internal/workflows/enrich.go`
- `internal/enrichment/asn.go`
- `internal/enrichment/geoip.go`

**Workflow:**
```go
func (w *EnrichWorkflow) Run(
    ctx restate.WorkflowContext,
    input EnrichInput,
) (EnrichOutput, error) {
    assets := getAssets(input.AssetSetRef)

    for _, host := range assets.Hosts {
        // Step 1: ASN lookup
        asn, _ := restate.Run(ctx, func(ctx restate.RunContext) (int, error) {
            return lookupASN(host.IP)
        })

        // Step 2: GeoIP lookup
        geo, _ := restate.Run(ctx, func(ctx restate.RunContext) (GeoData, error) {
            return lookupGeoIP(host.IP)
        })

        host.ASN = asn
        host.City = geo.City
        host.Country = geo.Country
    }

    enrichedRef := storeEnrichedAssets(assets)
    return EnrichOutput{EnrichedSetRef: enrichedRef}, nil
}
```

**External Dependencies:**
- MaxMind GeoLite2 DB
- ASN lookup service

**Acceptance:**
- IPs enriched with ASN
- IPs enriched with city/country
- Adds <5s to pipeline

---

#### M5-T4: wf.graph Implementation (4 hours)
**Files:**
- `internal/workflows/graph.go`
- `internal/db/surrealdb/graph_ops.go`

**Workflow:**
```go
func (w *GraphWorkflow) Run(
    ctx restate.WorkflowContext,
    input GraphInput,
) error {
    assets := getEnrichedAssets(input.EnrichedSetRef)

    for _, host := range assets.Hosts {
        // Step 1: Upsert HOST node
        restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            upsertHost(host)
            return restate.Void{}, nil
        })

        // Step 2: Upsert PORT nodes
        for _, port := range host.Ports {
            restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
                upsertPort(port)
                return restate.Void{}, nil
            })

            // Step 3: Create HAS edge
            restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
                createHasEdge(host.IP, port.Number)
                return restate.Void{}, nil
            })
        }

        // Step 4: Create geo edges
        restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
            createGeoEdges(host)
            return restate.Void{}, nil
        })
    }

    return nil
}
```

**Acceptance:**
- All nodes upserted
- All edges created
- Idempotent (can re-run safely)

---

### MILESTONE 6: Enrichment Pipeline (Weeks 11-12)
**Goal:** ASN, GeoIP, CPE mapping

#### M6-T1: MaxMind Integration (3 hours)
**Files:**
- `internal/enrichment/maxmind.go`
- `scripts/download-maxmind.sh`

**Setup:**
```bash
# Download GeoLite2 databases
curl -o GeoLite2-City.mmdb https://...
curl -o GeoLite2-ASN.mmdb https://...
```

**Code:**
```go
import "github.com/oschwald/geoip2-golang"

func lookupGeoIP(ip string) (GeoData, error) {
    db, _ := geoip2.Open("GeoLite2-City.mmdb")
    defer db.Close()

    record, _ := db.City(net.ParseIP(ip))

    return GeoData{
        City:    record.City.Names["en"],
        Region:  record.Subdivisions[0].Names["en"],
        Country: record.Country.Names["en"],
        Lat:     record.Location.Latitude,
        Lon:     record.Location.Longitude,
    }, nil
}
```

**Acceptance:**
- Lookups return correct data
- Handles missing data gracefully
- Performance: <10ms per lookup

---

#### M6-T2: CPE Mapping (4 hours)
**Files:**
- `internal/enrichment/cpe.go`
- `internal/enrichment/cpe_dict.json`

**CPE Mapping Logic:**
```go
func mapServiceToCPE(service Service) []string {
    // Example: nginx 1.25.1 → cpe:2.3:a:nginx:nginx:1.25.1:*:*:*:*:*:*:*

    if service.Product == "" {
        return nil
    }

    cpe := fmt.Sprintf("cpe:2.3:a:%s:%s:%s:*:*:*:*:*:*:*",
        normalizeVendor(service.Product),
        service.Product,
        service.Version,
    )

    return []string{cpe}
}
```

**CPE Dictionary:**
- Load from NVD CPE dictionary
- Map common products (nginx, apache, openssh, etc.)

**Acceptance:**
- Nginx → correct CPE
- Apache → correct CPE
- Unknown products → empty list

---

### MILESTONE 7: Vulnerability Correlation (Weeks 13-14)
**Goal:** NVD integration, AFFECTED_BY edges

#### M7-T1: NVD Data Sync (4 hours)
**Files:**
- `cmd/vuln-sync/main.go`
- `internal/nvd/client.go`
- `internal/nvd/parser.go`

**NVD API Client:**
```go
func fetchCVEs(startDate time.Time) ([]CVE, error) {
    url := fmt.Sprintf("https://services.nvd.nist.gov/rest/json/cves/2.0?pubStartDate=%s",
        startDate.Format("2006-01-02T15:04:05"))

    resp, _ := http.Get(url)
    defer resp.Body.Close()

    var result NVDResponse
    json.NewDecoder(resp.Body).Decode(&result)

    return result.Vulnerabilities, nil
}
```

**Storage:**
```go
func storeVulnDoc(cve CVE) {
    db.Query(`
        CREATE vuln_doc:$cve_id CONTENT {
            cve_id: $cve_id,
            title: $title,
            summary: $summary,
            cvss: $cvss,
            cpe: $cpe,
            published_date: $published
        }
    `, map[string]interface{}{
        "cve_id": cve.ID,
        // ...
    })
}
```

**Acceptance:**
- Fetches recent CVEs
- Stores in vuln_doc table
- Handles rate limiting (50 req/30s)

---

#### M7-T2: CPE Matching & Edge Creation (4 hours)
**Files:**
- `internal/workflows/vuln_correlation.go`

**Workflow:**
```go
func (w *VulnCorrelationWorkflow) Run(
    ctx restate.WorkflowContext,
    input VulnInput,
) error {
    services := getAllServices()

    for _, service := range services {
        if len(service.CPE) == 0 {
            continue
        }

        // Find matching vulnerabilities
        vulns := findVulnsByCPE(service.CPE)

        // Create AFFECTED_BY edges
        for _, vuln := range vulns {
            restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
                createAffectedByEdge(service.ID, vuln.CVE)
                return restate.Void{}, nil
            })
        }
    }

    return nil
}
```

**Acceptance:**
- Services linked to CVEs
- AFFECTED_BY edges created
- Only matches CPE correctly

---

#### M7-T3: Pro Tier Gating (2 hours)
**Files:**
- `internal/api/middleware/pro_tier.go`
- `internal/db/queries/user_tier.go`

**Middleware:**
```go
func RequireProTier(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := getUserFromContext(r.Context())

        if user.Tier != "pro" {
            http.Error(w, "402 Payment Required", 402)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

**Apply to Routes:**
```go
r.Route("/v0", func(r chi.Router) {
    r.With(RequireProTier).Get("/host/{ip}/vulns", getHostVulns)
    r.With(RequireProTier).Post("/ai/summary", aiSummary)
})
```

**Acceptance:**
- Free tier → 402 on Pro endpoints
- Pro tier → access granted

---

### MILESTONE 8: Vector Search & AI (Weeks 15-18)
**Goal:** Hybrid RAG, GPT-4 summaries

#### M8-T1: OpenAI Embeddings (3 hours)
**Files:**
- `internal/ai/embeddings.go`
- `cmd/embed-vulns/main.go`

**Embedding Generation:**
```go
func generateEmbedding(text string) ([]float64, error) {
    client := openai.NewClient(apiKey)

    resp, _ := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
        Model: openai.AdaEmbeddingV2,
        Input: []string{text},
    })

    return resp.Data[0].Embedding, nil
}

func embedVulnDoc(vuln VulnDoc) {
    text := fmt.Sprintf("%s %s %s",
        vuln.Title,
        vuln.Summary,
        strings.Join(vuln.CPE, " "),
    )

    embedding := generateEmbedding(text)

    db.Query(`
        UPDATE vuln_doc:$id SET embedding = $embedding
    `, map[string]interface{}{
        "id": vuln.ID,
        "embedding": embedding,
    })
}
```

**Acceptance:**
- Embeddings generated for all vulns
- Stored in SurrealDB
- Vector index created

---

#### M8-T2: Vector Similarity Search (3 hours)
**Files:**
- `internal/db/queries/vector_search.go`

**Query:**
```sql
SELECT
    cve_id,
    title,
    cvss,
    vector::similarity::cosine(embedding, $query_embedding) AS similarity
FROM vuln_doc
WHERE similarity > 0.8
ORDER BY similarity DESC
LIMIT 20;
```

**Go Wrapper:**
```go
func vectorSearch(queryText string, k int) ([]VulnDoc, error) {
    embedding := generateEmbedding(queryText)

    result := db.Query(`...`, map[string]interface{}{
        "query_embedding": embedding,
        "k": k,
    })

    return result, nil
}
```

**Acceptance:**
- Returns similar vulns
- P95 latency < 250ms
- Relevance is high (manual check)

---

#### M8-T3: wf.ai_engine Implementation (4 hours)
**Files:**
- `internal/workflows/ai.go`
- `internal/ai/rag.go`

**Workflow:**
```go
func (w *AIEngineWorkflow) Run(
    ctx restate.WorkflowContext,
    input AIInput,
) (AISummary, error) {
    // Step 1: Graph query (get services, vulns for IP)
    graphData, _ := restate.Run(ctx, func(ctx restate.RunContext) (GraphData, error) {
        return queryHostGraph(input.IP)
    })

    // Step 2: Vector search (k=20 similar vulns)
    vectorData, _ := restate.Run(ctx, func(ctx restate.RunContext) ([]VulnDoc, error) {
        queryText := buildQueryText(graphData)
        return vectorSearch(queryText, 20)
    })

    // Step 3: Assemble hybrid context (max 8KB)
    context := assembleContext(graphData, vectorData)

    // Step 4: Call GPT-4 with timeout (3.5s)
    summary, err := restate.Run(ctx, func(ctx restate.RunContext) (string, error) {
        return callGPT4(context, input.Query)
    })

    if err != nil {
        // Fallback: graph-only summary
        summary = generateFallbackSummary(graphData)
    }

    // Step 5: Cache result (24h TTL)
    restate.Set(ctx, "cache:"+input.IP, summary)

    return AISummary{Summary: summary}, nil
}
```

**GPT-4 Prompt:**
```
You are a security analyst. Analyze this host:

Host: {ip}
Services:
- Port 22: OpenSSH 7.4 (CVE-2018-1234: CVSS 9.8)
- Port 80: nginx 1.18.0 (CVE-2021-5678: CVSS 7.5)

Similar vulnerabilities (vector search):
1. CVE-2020-1111: SSH key exposure, CVSS 9.8
2. CVE-2019-2222: nginx buffer overflow, CVSS 8.1

Provide:
1. Summary (1 sentence)
2. Top 3 Risks (bullet points)
3. Next Steps (bullet points)
```

**Acceptance:**
- Returns 3-bullet format
- P95 latency < 4s
- Fallback works on timeout

---

### MILESTONE 9: Testing & Documentation (Weeks 19-20)
**Goal:** Production-ready quality

#### M9-T1: Integration Tests (6 hours)
**Files:**
- `tests/integration/e2e_test.go`
- `tests/integration/docker-compose.test.yml`

**E2E Test:**
```go
func TestFullScanWorkflow(t *testing.T) {
    // 1. Submit scan via API
    resp := submitScan(t, "192.168.1.1")
    assert.Equal(t, 200, resp.StatusCode)

    // 2. Wait for workflow completion
    time.Sleep(10 * time.Second)

    // 3. Query host
    host := queryHost(t, "192.168.1.1")
    assert.NotNil(t, host)

    // 4. Verify enrichment
    assert.NotEmpty(t, host.City)
    assert.NotZero(t, host.ASN)

    // 5. Verify graph relationships
    assert.NotEmpty(t, host.Ports)
}
```

**Acceptance:**
- Full workflow completes
- All assertions pass
- Runs in CI/CD

---

#### M9-T2: Performance Testing (4 hours)
**Files:**
- `tests/load/k6-script.js`

**Load Test (k6):**
```javascript
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 100 },  // Steady
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<600'],  // 95% < 600ms
  },
};

export default function() {
  let res = http.get('http://localhost:3000/v0/search?city=Paris');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 600ms': (r) => r.timings.duration < 600,
  });
}
```

**Acceptance:**
- P95 latency < 600ms
- No errors at 100 RPS
- SLOs met

---

#### M9-T3: API Documentation (4 hours)
**Files:**
- `docs/api.yaml` (OpenAPI 3.0)
- `docs/architecture.md`
- `docs/deployment.md`

**OpenAPI Spec:**
```yaml
openapi: 3.0.0
info:
  title: Spectra-Red API
  version: 1.0.0
paths:
  /v0/mesh/ingest:
    post:
      summary: Submit scan results
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ScanEnvelope'
      responses:
        '200':
          description: Scan accepted
```

**Acceptance:**
- All endpoints documented
- Examples provided
- Swagger UI works

---

## 3. Testing Strategy

### 3.1 Unit Tests
**Coverage Target:** >80%

**Key Test Files:**
- `internal/auth/ed25519_test.go`
- `internal/scanner/naabu_parser_test.go`
- `internal/db/queries/*_test.go`
- `internal/workflows/*_test.go`

**Pattern:**
```go
func TestVerifySignature(t *testing.T) {
    tests := []struct {
        name    string
        input   ScanEnvelope
        wantErr bool
    }{
        {"valid signature", validEnvelope(), false},
        {"invalid signature", tamperedEnvelope(), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := VerifyEnvelope(tt.input)
            assert.Equal(t, tt.wantErr, err != nil)
        })
    }
}
```

### 3.2 Integration Tests
**Test Scenarios:**
1. Full scan workflow (scan → enrich → graph → query)
2. Pro tier vulnerability correlation
3. AI summarization with fallback
4. Rate limiting enforcement
5. Idempotency of ingest

**Test Infrastructure:**
- Testcontainers for SurrealDB, Restate
- Docker Compose for full stack
- Real HTTP requests

### 3.3 Performance Tests
**Tools:** k6, vegeta

**Test Scenarios:**
1. Ingest throughput (target: 100 req/min sustained)
2. Query latency (target: P95 < 600ms)
3. Vector search (target: P95 < 250ms)
4. AI endpoint (target: P95 < 4s)

**Success Criteria:**
- All SLOs met under load
- No memory leaks
- Graceful degradation

---

## 4. Deployment Plan

### 4.1 Local Development
```bash
# Start services
docker-compose up -d

# Run migrations
./scripts/setup-db.sh

# Start API
go run cmd/api/main.go

# Start workflows
go run cmd/workflows/main.go
```

### 4.2 Production (Kubernetes)
**Files:**
- `deployments/k8s/api-deployment.yaml`
- `deployments/k8s/workflows-deployment.yaml`
- `deployments/k8s/surrealdb-statefulset.yaml`
- `deployments/k8s/restate-deployment.yaml`

**Components:**
- API: 3 replicas (horizontal scaling)
- Workflows: 2 replicas
- SurrealDB: StatefulSet with persistent volumes
- Restate: 1 replica (MVP)

### 4.3 CI/CD Pipeline
**GitHub Actions:**
```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: go test ./...
      - run: go build ./...

  integration:
    runs-on: ubuntu-latest
    steps:
      - run: docker-compose -f docker-compose.test.yml up -d
      - run: go test -tags=integration ./tests/integration
```

---

## 5. Risk Assessment & Mitigation

### 5.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| SurrealDB write bottleneck | Medium | High | Batch writes, async commits, benchmarking |
| Restate learning curve | Medium | Medium | 1 week dedicated learning, community support |
| OpenAI API costs | Low | Medium | Aggressive caching (24h TTL), rate limiting |
| Naabu scanning blocked | High | Medium | Relay rotation, rate limiting, do-not-scan list |

### 5.2 Operational Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Community spam | Medium | High | Ed25519 signatures, reputation system, manual review |
| Data staleness | Low | Medium | Monitoring, alerting on freshness metrics |
| Legal/compliance | Medium | Critical | Do-not-scan list, GDPR compliance, legal review |

---

## 6. Success Metrics

### 6.1 Performance (Week 8)
- ✓ Ingest P95 < 2s
- ✓ Query P95 < 600ms
- ✓ Planning P95 < 700ms
- ✓ Throughput: 100+ scans/min

### 6.2 Quality (Week 20)
- ✓ Unit test coverage >80%
- ✓ All integration tests pass
- ✓ No critical security vulnerabilities
- ✓ API documentation complete

### 6.3 Adoption (Month 3 post-launch)
- ✓ 50+ community contributors
- ✓ 1M+ hosts in mesh
- ✓ 5-10 Pro tier users

---

## 7. Dependencies & Prerequisites

### 7.1 External Services
- MaxMind GeoLite2 account (free)
- NVD API key (free)
- OpenAI API key ($20/month budget)

### 7.2 Development Tools
- Go 1.23+
- Docker Desktop
- Naabu (`go install github.com/projectdiscovery/naabu/v2/cmd/naabu@latest`)

### 7.3 Infrastructure (MVP)
**Minimum:**
- Single server: 4 vCPU, 16GB RAM, 100GB SSD
- Or: Hetzner CPX31 (~€14/month)

---

## 8. Next Steps

### Immediate Actions (Week 1)
1. Review this plan with team
2. Set up development environment
3. Begin M1-T1 (project initialization)
4. Schedule weekly sync meetings

### Key Decision Points
- **Week 4:** Review ingest performance, adjust if needed
- **Week 8:** MVP demo, gather feedback
- **Week 14:** Evaluate Pro tier pricing
- **Week 20:** Launch decision

---

## Appendix: Quick Reference

### Key Commands
```bash
# Development
docker-compose up -d
go run cmd/api/main.go
go test ./...

# CLI Usage
spectra scan 192.168.1.1 --submit
spectra mesh plan city "Paris" --min-age 5m
spectra mesh query service redis --city Paris

# Database
surreal start --user root --pass root memory
surreal import schema.surql
```

### Important Files
- `SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md` - Requirements
- `GO_PATTERNS_REFERENCE.md` - Code patterns
- `RESTATE_DEEP_DIVE.md` - Workflow guidance
- `SURREALDB_SCHEMA_GUIDE.md` - Database queries

### Support Resources
- Restate Discord: https://discord.gg/restate
- SurrealDB Discord: https://discord.gg/surrealdb
- Project GitHub: (TBD)

---

**Document Version:** 1.0
**Last Updated:** November 1, 2025
**Status:** Ready for Implementation
