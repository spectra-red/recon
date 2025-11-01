# Spectra-Red Deployment

Docker Compose environment for local development and testing.

## Services

### SurrealDB (Port 8000)
- **Purpose**: Graph + Vector database for intel mesh
- **Credentials**: root / root
- **Health**: http://localhost:8000/health
- **Data**: In-memory (use volume for persistence)

### Restate (Ports 8080, 9070)
- **Purpose**: Durable workflow orchestration
- **Ingress**: http://localhost:8080
- **Admin UI**: http://localhost:9070
- **Data**: Persisted in `restate-data` volume

### API Server (Port 3000)
- **Purpose**: HTTP API for scan submission and queries
- **Health**: http://localhost:3000/health
- **Endpoints**: See API documentation

## Quick Start

### 1. Start All Services

```bash
docker-compose up -d
```

### 2. Check Service Health

```bash
# Check all services
docker-compose ps

# View logs
docker-compose logs -f

# Check specific service
docker-compose logs -f api
```

### 3. Initialize Database

```bash
# Wait for services to be healthy
docker-compose ps

# Run setup script
./scripts/setup-db.sh
```

### 4. Access Services

- **API**: http://localhost:3000
- **SurrealDB**: http://localhost:8000
- **Restate UI**: http://localhost:9070

## Development Workflow

### Build and Run

```bash
# Build all services
docker-compose build

# Start in foreground (see logs)
docker-compose up

# Start in background
docker-compose up -d

# Rebuild specific service
docker-compose build api
docker-compose up -d api
```

### Logs and Debugging

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f surrealdb
docker-compose logs -f restate
docker-compose logs -f api

# View last 100 lines
docker-compose logs --tail=100 api
```

### Database Management

```bash
# Connect to SurrealDB
docker-compose exec surrealdb surreal sql \
  --conn http://localhost:8000 \
  --user root --pass root \
  --ns spectra --db intel

# Run setup script
./scripts/setup-db.sh

# Apply schema manually
curl -X POST http://localhost:8000/sql \
  -H "NS: spectra" \
  -H "DB: intel" \
  -u root:root \
  --data-binary @internal/db/schema/schema.surql
```

### Restate Management

```bash
# View registered services
curl http://localhost:9070/services

# View invocations
curl http://localhost:9070/invocations

# Open Admin UI
open http://localhost:9070
```

## Cleanup

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v

# Remove all containers, networks, and volumes
docker-compose down -v --remove-orphans
```

## Configuration

### Environment Variables

Create a `.env` file in the `deployments/` directory:

```bash
# SurrealDB
SURREALDB_URL=http://surrealdb:8000
SURREALDB_USER=root
SURREALDB_PASS=root
SURREALDB_NS=spectra
SURREALDB_DB=intel

# Restate
RESTATE_URL=http://restate:8080

# API Server
PORT=3000
LOG_LEVEL=info

# Optional: External Services
# OPENAI_API_KEY=sk-...
# MAXMIND_LICENSE_KEY=...
```

### Persistent Storage

By default, SurrealDB runs in-memory mode. For persistence:

```yaml
# In docker-compose.yml, modify surrealdb service:
surrealdb:
  image: surrealdb/surrealdb:latest
  command: start --log trace --user root --pass root file:///data/spectra.db
  volumes:
    - surrealdb-data:/data
```

## Health Checks

All services include health checks:

```bash
# Check service health
docker-compose ps

# Manual health check
curl http://localhost:8000/health  # SurrealDB
curl http://localhost:9070/health  # Restate
curl http://localhost:3000/health  # API
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs [service-name]

# Restart specific service
docker-compose restart [service-name]

# Rebuild and restart
docker-compose up -d --build [service-name]
```

### Port Conflicts

If ports are already in use, modify `docker-compose.yml`:

```yaml
ports:
  - "8001:8000"  # Use 8001 externally instead of 8000
```

### Database Connection Issues

```bash
# Verify SurrealDB is running
curl http://localhost:8000/health

# Check API environment variables
docker-compose exec api env | grep SURREALDB

# Restart with fresh database
docker-compose down -v
docker-compose up -d
./scripts/setup-db.sh
```

### Build Failures

```bash
# Clean build cache
docker-compose build --no-cache api

# Check Go module issues
docker-compose run --rm api go mod verify
docker-compose run --rm api go mod tidy
```

## Network Architecture

Services communicate via the `spectra-network` bridge:

```
┌─────────────┐
│   API:3000  │
└──────┬──────┘
       │
   ┌───┴────┬────────────┐
   │        │            │
┌──▼───┐ ┌─▼──────┐ ┌───▼────┐
│ Sur  │ │ Restate│ │ (ext)  │
│ real │ │ :8080  │ │ clients│
│ DB   │ │ :9070  │ │        │
└──────┘ └────────┘ └────────┘
```

## Production Deployment

For production, see:
- `deployments/k8s/` - Kubernetes manifests
- `docs/deployment.md` - Production deployment guide

**DO NOT use default credentials in production!**

## Next Steps

After starting services:

1. ✅ Verify health checks pass
2. ✅ Run database setup: `./scripts/setup-db.sh`
3. ✅ Test API endpoint: `curl http://localhost:3000/health`
4. ✅ Open Restate UI: http://localhost:9070
5. 📝 Continue with M1-T3: SurrealDB Schema Definition

## Support

- **Issues**: See docker-compose logs
- **Documentation**: See `docs/` directory
- **Schema**: See `internal/db/schema/`
