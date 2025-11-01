# Spectra-Red Intel Mesh

A community-driven security intelligence mesh that lets researchers, red teams, and security professionals share real-time network scan data, query a global graph of internet-facing assets, and get AI-powered vulnerability summaries.

## The Problem

Current tools (Shodan, Censys) are:
- **Stale:** Data is 2-4 weeks old
- **Expensive:** $500-2K/year for comprehensive access
- **Siloed:** No community contributions, everyone scans the same targets repeatedly
- **Time-consuming:** Manual correlation of vulnerabilities across data sources

## The Solution

A **distributed mesh** where:
- Community runners submit fresh scan data (gets you free Pro access)
- Central graph database indexes everything (SurrealDB)
- Durable workflows handle enrichment and AI analysis (Restate)
- Real-time queries return results in <2 seconds
- OSS tier is free; Pro adds AI + vuln correlation

## Architecture

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

## Technology Stack

- **Language:** Go 1.23+
- **HTTP Router:** Chi v5
- **Database:** SurrealDB (graph + vector storage)
- **Workflows:** Restate (durable execution)
- **Authentication:** Ed25519 signatures + JWT
- **AI:** OpenAI GPT-4 with hybrid RAG
- **CLI:** Cobra framework
- **Deployment:** Docker Compose (dev), Kubernetes (prod)

## Project Structure

```
.
├── cmd/
│   ├── api/          # HTTP API server (port 3000)
│   ├── spectra/      # spectra CLI tool
│   └── workflows/    # Restate workflow service (port 9080)
├── internal/
│   ├── api/          # HTTP handlers and middleware
│   ├── cli/          # CLI commands and configuration
│   ├── client/       # HTTP client for API
│   ├── db/           # Database layer (SurrealDB)
│   ├── workflows/    # Restate durable workflows
│   ├── enrichment/   # ASN, GeoIP, CPE enrichment
│   ├── auth/         # Authentication (Ed25519)
│   └── models/       # Domain models
├── deployments/
│   └── docker-compose.yml  # Infrastructure services
├── scripts/          # Setup and utility scripts
└── docs/             # Complete documentation
    ├── api/          # API endpoint documentation
    ├── cli/          # CLI reference and guides
    ├── workflows/    # Workflow documentation
    ├── deployment/   # Deployment guides
    ├── planning/     # PRD and implementation plans
    └── archive/      # Historical documentation
```

## Getting Started

### Prerequisites

- Go 1.23 or later
- Docker Desktop
- Naabu (optional, for scanning)

### Quick Start

```bash
# 1. Clone and build
git clone https://github.com/spectra-red/recon.git
cd recon/.conductor/melbourne
go mod download && go build ./...

# 2. Start infrastructure
cd deployments && docker-compose up -d

# 3. Initialize database
cd .. && ./scripts/setup-db.sh

# 4. Start services
go run cmd/api/main.go &           # API server (port 3000)
go run cmd/workflows/main.go &     # Workflow service (port 9080)

# 5. Configure CLI
mkdir -p ~/.spectra
cat > ~/.spectra/.spectra.yaml <<EOF
api:
  url: http://localhost:3000
output:
  format: table
EOF

# 6. Test the system
curl http://localhost:3000/health
go run cmd/spectra/main.go version
```

### Usage Examples

```bash
# Submit scan results
naabu -host example.com -json | spectra ingest -

# Query a host
spectra query host 1.2.3.4 --depth 2

# Search for vulnerabilities
spectra query similar "nginx remote code execution"

# Check job status
spectra jobs list
spectra jobs get <job-id> --watch
```

## Development Status

**Current Phase:** Waves 1-3 Complete ✅

### Completed Features (17/47 tasks - 36% MVP)

**Wave 1: Foundation** ✅
- [x] Go project structure with 78 source files
- [x] Docker Compose infrastructure (SurrealDB, Restate)
- [x] Database schema (24 tables, 26 indices, 146 seed records)
- [x] HTTP API server with Chi router and middleware

**Wave 2: Ingest & Query** ✅
- [x] Ingest API with Ed25519 authentication
- [x] Job tracking system with UUID v7 IDs
- [x] Restate workflow for scan processing
- [x] Host query API with graph traversal (depth 0-5)
- [x] Advanced graph queries (ASN, location, vuln, service)
- [x] Vector similarity search with OpenAI embeddings

**Wave 3: CLI & Workflows** ✅
- [x] CLI tool with 8 commands (ingest, query, jobs, version)
- [x] Configuration management (Viper + environment variables)
- [x] ASN enrichment workflow (Team Cymru)
- [x] GeoIP enrichment workflow (MaxMind MMDB)
- [x] CPE matching workflow (NVD API integration)

### Statistics
- **19,583** lines of code
- **180+** tests (100% pass rate)
- **85-90%** test coverage
- **12** API endpoints
- **4** Restate workflows
- **4** executable services

See [docs/planning/DETAILED_IMPLEMENTATION_PLAN.md](docs/planning/DETAILED_IMPLEMENTATION_PLAN.md) for the complete roadmap.

## Performance Targets

- **Ingest:** P95 < 2s for 100-host batch
- **Query:** P95 < 600ms for complex graph queries
- **Planning:** P95 < 700ms for stale host discovery
- **AI Summary:** P95 < 4s with vector RAG
- **Throughput:** 100+ scans/min sustained

## Contributing

Contributions welcome! This is an open-source project under development.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

See [LICENCE.md](LICENCE.md) and [LICENCE-POLICY.md](LICENCE-POLICY.md)

## Documentation

📚 **[Complete Documentation](docs/README.md)**

- **[API Reference](docs/api/)** - REST API endpoints and examples
- **[CLI Guide](docs/cli/README_CLI.md)** - Command-line interface
- **[Workflows](docs/workflows/)** - Enrichment workflow guides
- **[Deployment](docs/deployment/README_DOCKER_SETUP.md)** - Production deployment
- **[Planning](docs/planning/)** - PRD and implementation plans

## Support

- **Documentation**: [docs/](docs/README.md)
- **Issues**: [GitHub Issues](https://github.com/spectra-red/recon/issues)
- **Community**: (Coming soon)

## Roadmap

**Phase 1 (Weeks 1-8):** Core mesh ingest + graph storage + basic queries
**Phase 2 (Weeks 9-14):** Enrichment pipeline + vulnerability correlation
**Phase 3 (Weeks 15-20):** AI summarization + Pro tier features

Target: Ship usable MVP in 5 months with 100+ community beta testers