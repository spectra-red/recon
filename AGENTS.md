# Spectra-Red Intel Mesh - AI Agent Development Guide

This document provides comprehensive guidance for AI agents (including Factory Droids, Claude Code, and other AI assistants) working with the Spectra-Red security intelligence mesh repository.

## Project Overview

Spectra-Red is a distributed security intelligence mesh for researchers, red teams, and security professionals to share real-time network scan data, query a global graph of internet-facing assets, and get AI-powered vulnerability summaries.

**Key Architecture Components:**
- **Durable workflows**: Restate-based async processing engine
- **Graph database**: SurrealDB with graph and vector storage
- **API Gateway**: Go + Chi with Ed25519 authentication
- **CLI Tool**: Cobra-based client for community runners

## Development Environment & Commands

### Build & Test Commands
```bash
# Build
go build ./...

# Run services locally
make run-api                    # Start API server (port 3000)
go run cmd/workflows/main.go   # Start workflow service (port 9080)
go run cmd/spectra/main.go     # Run CLI commands

# Testing
make test                       # Run all tests
make test-integration           # Run integration tests (requires services)
make test-coverage              # Generate coverage report
```

### Docker Development
```bash
make up                         # Start all services
make dev                        # Start development environment with DB setup
make down                       # Stop all services
make health                     # Check service health
make logs                       # View all logs
```

## Repository Structure

```
├─ cmd/                   → Executable services
│   ├── api/             → HTTP API server (Go + Chi, port 3000)
│   ├── spectra/         → CLI tool (Cobra-based)
│   └── workflows/       → Restate workflow service (port 9080)
├─ internal/             → Internal packages
│   ├── api/             → HTTP handlers, middleware, routes
│   ├── cli/             → CLI command implementations
│   ├── db/              → SurrealDB database layer
│   ├── workflows/       → Restate durable workflows
│   ├── enrichment/      → ASN, GeoIP, CPE enrichment
│   ├── auth/            → Ed25519 authentication
│   └── models/          → Domain models and data structures
├─ deployments/          → Docker Compose configurations
├─ scripts/             → Setup and utility scripts
├─ tests/                → Integration tests (requires full stack)
└─ docs/                 → Documentation
```

## Workflow Architecture

### Key Durable Workflows
The system uses Restate for durable, asynchronous processing:

1. **`wf.scan`** - Parse and normalize incoming scan data
2. **`wf.enrich`** - ASN, GeoIP, CPE enrichment pipelines
3. **`wf.graph`** - Upsert nodes/edges to graph database
4. **`wf.ai`** - Vector RAG + AI analysis with summaries

### Workflow Integration Rules
- **No direct HTTP calls** in API request handlers - always route through workflows
- **All database operations** use SurrealDB client with proper error handling
- **Rate limiting**: 60/min ingest, 30/min query enforced
- **Authentication**: Ed25519 signatures required for all API submissions

## AI Agent Development Patterns

### Recommended AI Agent Workflow

When working on this codebase, follow these patterns:

1. **Understand the codebase structure first** - Use Glob/Grep tools to explore
2. **Check dependencies** - Confirm existing libraries before adding new ones
3. **Follow Go conventions** - Use `make fmt && make lint && make vet && make test`
4. **Database first** - Ensure services are running before integration testing
5. **Workflow mindset** - Prefer async durable workflows over synchronous processing

### Code Style Requirements
- Go 1.23+ required
- Follow standard Go formatting (`make fmt`)
- Target 85%+ test coverage for new features
- Keep functions focused (<50 lines when possible)
- Use structured logging with zap
- Add comprehensive tests for all new features

### Before Making Changes
```bash
# Always run this command sequence first
make fmt && make lint && make vet && make test

# For integration work, start services
make dev
make health
```

## Testing Strategy

### Unit Tests (`*_test.go`)
- Co-located with source files
- Mock external dependencies (database, external APIs)
- Focus on business logic and validation

### Integration Tests
- Located in `tests/integration/`
- Require full stack running (`make dev`)
- Test API endpoints end-to-end
- Verify workflow execution and async operations

### Performance Requirements
- **Ingest operations**: P95 < 2 seconds
- **Query operations**: P95 < 600ms
- Use Go benchmarking for performance testing

## Development Gotchas

### Common Issues
- **Database connection**: Services must be up before running tests
- **Workflow timing**: Async workflows take time to complete (account for this)
- **Docker networking**: Services communicate via Docker network, not localhost
- **Environment config**: Use `.spectra.yaml` or environment variables for configuration
- **Test namespace**: Integration tests use separate database namespace

### External Dependencies
- **SurrealDB**: Database (port 8000)
- **Restate**: Workflow engine (port 9070)
- **MaxMind GeoIP**: ASN/location enrichment
- **Team Cymru**: ASN lookup service
- **OpenAI API**: GPT-4 analysis with vector embeddings

## Git Workflow

1. **Branch naming**: Use `feature/...` or `bugfix/...` patterns
2. **Pre-commit checks**: Run `make fmt && make lint && make vet && make test`
3. **Atomic commits**: Keep commits focused and well-described
4. **Documentation**: Update docs for API changes
5. **PR requirements**: Must pass tests and code review

## Troubleshooting Common Issues

### Service Startup Problems
```bash
# Check if services are running
make health

# View specific service logs
make logs-api       # API service logs
make logs-db        # SurrealDB logs
make logs-restate   # Restate workflow logs
```

### Database Issues
```bash
# Reset database schema
make db-reset

# Check database connection
make logs-db
```

### Workflow Issues
- Check that all services are healthy (`make health`)
- Verify workflow execution in Restate logs
- Account for async timing in integration tests

## Configuration

- **CLI config**: `~/.spectra/.spectra.yaml`
- **Environment variables**: Override config file settings
- **Database connections**: Configured via environment
- **API keys**: Loaded from secure storage, not committed to repo

## Integration with AI Agents

### This Project and Factory Droids
The Spectra-Red project is designed with AI agent collaboration in mind:

1. **Durable workflow architecture** allows agents to submit long-running tasks
2. **Graph database** provides queryable asset intelligence
3. **Rate limiting and security** protects against automated abuse
4. **Comprehensive testing** ensures AI-generated code is reliable

### AI Agent Best Practices
- **Batch operations**: Use workflows for bulk data processing
- **Error handling**: Implement proper retry logic for external service calls
- **Security**: Never commit API keys or sensitive configuration
- **Testing**: Generate comprehensive tests alongside feature code
- **Documentation**: Update this guide when adding new patterns

### Common AI Agent Tasks
1. **Adding new enrichment sources** - Create new workflow handlers
2. **API endpoint development** - Follow existing handler patterns
3. **CLI command enhancement** - Extend Cobra-based commands
4. **Database queries** - Use existing SurrealDB client patterns
5. **Workflow optimization** - Improve async processing paths

Remember: This system handles potentially sensitive security data. Always follow security best practices and test thoroughly before production deployment.

## Project Layout

```
├─ cmd/              → Executable services
│   ├── api/         → HTTP API server (port 3000)
│   ├── spectra/     → CLI tool
│   └── workflows/   → Restate workflow service (port 9080)
├─ internal/         → Internal packages
│   ├── api/         → HTTP handlers and middleware
│   ├── cli/         → CLI commands
│   ├── db/          → Database layer (SurrealDB)
│   ├── workflows/   → Restate durable workflows
│   ├── enrichment/  → ASN, GeoIP, CPE enrichment
│   ├── auth/        → Authentication (Ed25519)
│   └── models/      → Domain models
├─ deployments/      → Docker Compose configs
├─ scripts/          → Setup and utility scripts
└─ docs/             → Documentation
```

- **API Code**: Live in `internal/api/` and `cmd/api/`
- **CLI Code**: Live in `internal/cli/` and `cmd/spectra/`
- **Workflow Code**: Live in `internal/workflows/` and `cmd/workflows/`
- **Database**: SurrealDB with graph and vector storage
- **Tests**: Co-located with source code (`*_test.go`)
- **Integration Tests**: In `tests/integration/`

## Architecture Overview

The system is a distributed security intelligence mesh:
- **API Gateway**: Go + Chi router with Ed25519 auth (port 3000)
- **Workflow Engine**: Restate durable workflows for async processing (port 9080)
- **Database**: SurrealDB cluster with graph + vector storage (port 8000)
- **CLI Tool**: Cobra-based CLI for interacting with the system

Key workflows:
- `wf.scan` - Parse & normalize scan data
- `wf.enrich` - ASN, GeoIP, CPE enrichment
- `wf.graph` - Upsert nodes/edges to graph database
- `wf.ai` - Vector RAG + GPT-4 analysis

## Development Patterns & Constraints

### Code Style
- Follow Go standard formatting: `make fmt`
- Use `make lint` and `make vet` before commits
- Add comprehensive tests for new features (target 85%+ coverage)
- Keep functions focused and small (<50 lines when possible)
- Use structured logging with zap

### Dependencies
- Go 1.23+ required
- Use Go modules (go.mod) for dependency management
- Prefer standard library over external dependencies
- New dependencies must be justified in PR description
- API packages use chi/router, SurrealDB client, OpenAI client

### Database & Workflows
- All database operations use SurrealDB client
- Durable workflows use Restate SDK
- Async operations go through workflow engine
- Never do direct HTTP calls in request handlers - use workflows

### Security
- Ed25519 signatures for all API submissions
- JWT tokens for authentication
- API key validation required
- Rate limiting enforced (60/min ingest, 30/min query)

## Environment Setup

### Local Development
```bash
# Start services
make up

# Initialize database
make db-setup

# Run API locally
make run-api

# Run CLI commands
go run cmd/spectra/main.go version
```

### Docker Development
```bash
# Start all services
make dev

# Check health
make health

# View logs
make logs
```

## Commands Reference

### Make Commands (Recommended)
- `make test` - Run all tests
- `make test-integration` - Run integration tests (requires services)
- `make test-coverage` - Generate coverage report
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint
- `make vet` - Run go vet
- `make build-local` - Build binaries in bin/
- `make health` - Check all service health

### Go Commands
- `go run cmd/api/main.go` - Start API server
- `go run cmd/workflows/main.go` - Start workflow service
- `go run cmd/spectra/main.go` - Run CLI
- `go test ./...` - Run tests
- `go build ./...` - Build all packages

## Testing Strategy

### Unit Tests
- Co-located with source code (`*_test.go`)
- Mock external dependencies (database, APIs)
- Focus on business logic and validation

### Integration Tests
- Located in `tests/integration/`
- Require full stack running (make dev)
- Test API endpoints end-to-end
- Verify workflow execution

### Performance Tests
- Target: P95 < 2s ingest, <600ms queries
- Use built-in Go benchmarking
- Test with realistic data volumes

## Gotchas

- **Database connection**: Services need to be up before running tests
- **Workflow timing**: Async workflows take time to complete
- **API rate limits**: Development env has stricter limits
- **Docker networking**: Services communicate via Docker network
- **Environment variables**: Configuration via `.spectra.yaml` or env vars
- **Test database**: Integration tests use separate test namespace

## Git Workflow

1. Branch from `main` with descriptive names (`feature/...`, `bugfix/...`)
2. Run `make fmt && make lint && make vet && make test` before committing
3. Keep commits focused and atomic
4. PRs require passing tests and code review
5. Documentation updates required for API changes

## External Dependencies

- **SurrealDB**: Database (runs on port 8000)
- **Restate**: Workflow engine (port 9070)
- **MaxMind GeoIP**: ASN/location enrichment
- **Team Cymru**: ASN lookup service
- **OpenAI API**: GPT-4 analysis with vector embeddings

## Configuration

- CLI config: `~/.spectra/.spectra.yaml`
- Environment variables override config file
- Database connection strings in environment
- API keys loaded from secure storage
