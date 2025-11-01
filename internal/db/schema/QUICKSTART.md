# SurrealDB Schema - Quick Start Guide

## 30-Second Setup

```bash
# From project root
cd /Users/seanknowles/Projects/recon/.conductor/melbourne

# Apply schema and seed data
./scripts/setup-db.sh

# Verify everything works
./scripts/verify-schema.sh
```

## What You Get

- **24 Tables**: Hosts, ports, services, vulnerabilities, geography
- **26 Indices**: Optimized for performance
- **146 Seed Records**: Common ports, countries, ASNs, cloud regions
- **Vector Search**: Ready for semantic vulnerability search
- **Graph Relationships**: Multi-hop traversal for threat intelligence

## Key Tables

| Table | Records | Purpose |
|-------|---------|---------|
| `host` | 0 | IP addresses with geo metadata |
| `port` | 0 | Port numbers and protocols |
| `service` | 0 | Service identification |
| `vuln` | 5 | Vulnerabilities (CVE data) |
| `common_port` | 28 | Well-known ports (SSH, HTTP, etc.) |
| `country` | 25 | Countries for geographic queries |
| `asn` | 20 | Cloud providers and ISPs |
| `cloud_region` | 26 | AWS, GCP, Azure regions |

## Sample Queries

### Find SSH Port
```sql
SELECT * FROM common_port WHERE label = 'ssh';
```

### List All AWS Regions
```sql
SELECT * FROM cloud_region WHERE provider = 'aws' ORDER BY code;
```

### Get Critical Vulnerabilities
```sql
SELECT * FROM vuln WHERE severity = 'critical' ORDER BY cvss DESC;
```

### Find Hosts in US
```sql
SELECT * FROM host WHERE country = 'US' LIMIT 10;
```

## Environment Variables

```bash
export SURREALDB_URL="http://localhost:8000"
export SURREALDB_USER="root"
export SURREALDB_PASS="root"
export SURREALDB_NS="spectra"
export SURREALDB_DB="intel"
```

## Files

- `schema.surql` - Complete schema (229 lines)
- `seed.surql` - Seed data (332 lines, 146 records)
- `examples.surql` - 50+ example queries
- `README.md` - Full documentation
- `VALIDATION.md` - Acceptance criteria checklist

## Scripts

- `scripts/setup-db.sh` - Apply schema and seed data
- `scripts/verify-schema.sh` - Run 40+ validation tests

## Next Steps

1. Start adding real data via API endpoints (M1-T4)
2. Implement graph queries for threat intelligence
3. Configure vector embeddings for RAG (Pro tier)

## Troubleshooting

**Schema won't apply?**
- Check SurrealDB is running: `curl http://localhost:8000/health`
- Verify credentials in environment variables

**Seed data conflicts?**
- Drop database: `curl -X POST http://localhost:8000/sql -u root:root -d "REMOVE DATABASE intel;"`
- Rerun setup: `./scripts/setup-db.sh`

**Need help?**
- See `README.md` for full documentation
- Check `examples.surql` for query patterns

## Schema Version

**Version**: 1.0.0
**SurrealDB**: >= 1.3.0 (for vector index support)
**Created**: 2025-11-01

---

For complete documentation, see [README.md](./README.md)
