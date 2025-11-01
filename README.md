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
│   ├── api/          # HTTP API server
│   ├── cli/          # spectra CLI tool
│   └── workflows/    # Restate workflow service
├── internal/
│   ├── api/          # HTTP handlers and middleware
│   ├── db/           # Database layer (SurrealDB)
│   ├── workflows/    # Restate workflows
│   ├── scanner/      # Scan execution and parsing
│   ├── auth/         # Authentication (Ed25519, JWT)
│   └── models/       # Domain models
├── pkg/
│   ├── client/       # Go SDK for API
│   └── types/        # Shared types
├── tests/
│   ├── integration/  # E2E tests
│   └── load/         # Performance tests
└── docs/             # Documentation

```

## Getting Started

### Prerequisites

- Go 1.23 or later
- Docker Desktop
- Naabu (optional, for scanning)

### Installation

```bash
# Clone the repository
git clone https://github.com/spectra-red/recon.git
cd recon/.conductor/melbourne

# Install dependencies
go mod download

# Build all binaries
go build ./...

# Run tests
go test ./...
```

### Development

```bash
# Start infrastructure services (SurrealDB, Restate)
docker-compose up -d

# Run API server
go run cmd/api/main.go

# Run workflow service
go run cmd/workflows/main.go

# Use the CLI
go run cmd/cli/main.go --help
```

## Development Status

**Current Phase:** Foundation (M1-T1 Complete)

- [x] Go workspace initialized
- [x] Module structure created
- [x] Dependencies added
- [x] Directory structure established
- [x] Skeleton files created
- [ ] Docker Compose setup
- [ ] SurrealDB schema
- [ ] Basic HTTP server
- [ ] Ingest endpoints
- [ ] Query API
- [ ] CLI tool
- [ ] Restate workflows
- [ ] Enrichment pipeline
- [ ] Vulnerability correlation
- [ ] AI summarization

See [DETAILED_IMPLEMENTATION_PLAN.md](DETAILED_IMPLEMENTATION_PLAN.md) for full roadmap.

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

## Support

- Documentation: See `/docs` directory
- Issues: GitHub Issues
- Community: (Coming soon)

## Roadmap

**Phase 1 (Weeks 1-8):** Core mesh ingest + graph storage + basic queries
**Phase 2 (Weeks 9-14):** Enrichment pipeline + vulnerability correlation
**Phase 3 (Weeks 15-20):** AI summarization + Pro tier features

Target: Ship usable MVP in 5 months with 100+ community beta testers