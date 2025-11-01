# Spectra-Red Intel Mesh Documentation

Welcome to the Spectra-Red Intel Mesh documentation. This guide will help you understand, deploy, and use the platform.

## Quick Start

- **[Main README](../README.md)** - Project overview and quick start
- **[Docker Deployment](deployment/README_DOCKER_SETUP.md)** - Deploy with Docker Compose
- **[CLI Guide](cli/README_CLI.md)** - Command-line interface documentation

## Documentation Structure

### 📡 API Documentation
API endpoints, request/response formats, and usage examples.

- **[Query API](api/API_QUERY_ENDPOINT.md)** - Host, graph, and vector similarity queries
- **[Quick Start Guide](api/QUICK_START_QUERY_API.md)** - Get started with query APIs

### 💻 CLI Documentation
Command-line tool for interacting with Spectra-Red.

- **[CLI Reference](cli/README_CLI.md)** - Complete CLI documentation
- **[Ingest Quick Start](cli/QUICK_START_INGEST.md)** - Submit scan results via CLI

### 🔄 Workflow Documentation
Automated enrichment workflows powered by Restate.

- **[ASN Enrichment](workflows/ASN_ENRICHMENT.md)** - Autonomous System Number lookup
- **[ASN Quick Start](workflows/ASN_QUICK_START.md)** - Get started with ASN enrichment
- **[GeoIP Quick Start](workflows/GEOIP_QUICK_START.md)** - Geographic enrichment
- **[CPE Workflow Examples](workflows/CPE_WORKFLOW_EXAMPLES.md)** - Vulnerability correlation

### 🚀 Deployment
Production deployment guides and infrastructure setup.

- **[Docker Deployment](deployment/README_DOCKER_SETUP.md)** - Deploy with Docker Compose
- **[Service Architecture](#)** - Multi-service architecture overview

### 📋 Planning & Architecture
Technical specifications and implementation plans.

- **[PRD](planning/SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md)** - Product requirements document
- **[Implementation Plan](planning/DETAILED_IMPLEMENTATION_PLAN.md)** - Detailed technical plan
- **[Roadmap](planning/IMPLEMENTATION_ROADMAP.md)** - Development milestones
- **[Planning Integration](planning/PLANNING_INTEGRATION_SUMMARY.md)** - Build system enhancement

### 📦 Archive
Historical documentation, completion reports, and research artifacts.

- **[Archived Documentation](archive/)** - Milestone completion reports and research

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Spectra-Red Intel Mesh                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────────────────┐ │
│  │   CLI    │───▶│ API      │───▶│  Restate Workflows   │ │
│  │ spectra  │    │ Server   │    │  ┌────────────────┐  │ │
│  └──────────┘    │ (Chi)    │    │  │ Ingest         │  │ │
│                  └──────────┘    │  │ ASN Enrich     │  │ │
│                       │          │  │ GeoIP Enrich   │  │ │
│                       │          │  │ CPE Match      │  │ │
│                       │          │  └────────────────┘  │ │
│                       ▼          └──────────────────────┘ │
│                  ┌──────────┐              │              │
│                  │SurrealDB │◀─────────────┘              │
│                  │  Graph   │                             │
│                  │  Vector  │                             │
│                  └──────────┘                             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. API Server (`cmd/api`)
- **Port**: 3000
- **Tech**: Go + Chi router
- **Features**: 12 RESTful endpoints, rate limiting, Ed25519 auth

### 2. Workflow Service (`cmd/workflows`)
- **Port**: 9080
- **Tech**: Restate SDK for Go
- **Features**: Durable workflows, automatic retries, state management

### 3. CLI Tool (`cmd/spectra`)
- **Tech**: Cobra + Viper
- **Features**: 8 commands, Ed25519 signing, multiple output formats

### 4. Database (SurrealDB)
- **Port**: 8000
- **Features**: Graph database, vector search, 24 tables, 26 indices

## Getting Started

### 1. Deploy Infrastructure

```bash
# Start all services
cd deployments
docker-compose up -d

# Initialize database schema
./scripts/setup-db.sh
```

### 2. Configure CLI

```bash
# Generate scanner keys
ssh-keygen -t ed25519 -f ~/.spectra/scanner_key

# Create config
cat > ~/.spectra/.spectra.yaml <<EOF
api:
  url: http://localhost:3000
scanner:
  private_key: $(cat ~/.spectra/scanner_key | base64)
  public_key: $(cat ~/.spectra/scanner_key.pub | base64)
EOF
```

### 3. Submit Scan Data

```bash
# From file
spectra ingest scan-results.json

# From stdin
naabu -host example.com -json | spectra ingest -
```

### 4. Query Intelligence

```bash
# Query host
spectra query host 1.2.3.4 --depth 3

# Vector similarity search
spectra query similar "nginx remote code execution"

# Graph queries
spectra query graph --type by_vuln --value CVE-2024-1234
```

## API Endpoints

### Ingest
- `POST /v1/mesh/ingest` - Submit scan results

### Query
- `GET /v1/query/host/{ip}` - Host details with graph traversal
- `POST /v1/query/graph` - Advanced graph queries
- `POST /v1/query/similar` - Vector similarity search

### Jobs
- `GET /v1/jobs` - List all jobs
- `GET /v1/jobs/{job_id}` - Get job status

### Health
- `GET /health` - Service health check

## Technologies

- **Language**: Go 1.23+
- **Web Framework**: Chi v5
- **Database**: SurrealDB 1.0+
- **Workflows**: Restate SDK
- **CLI**: Cobra + Viper
- **Vector Embeddings**: OpenAI API
- **GeoIP**: MaxMind GeoLite2

## Support

- **GitHub Issues**: [Report bugs](https://github.com/spectra-red/recon/issues)
- **Documentation**: This folder
- **License**: See [LICENSE](../LICENCE.md)

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.
