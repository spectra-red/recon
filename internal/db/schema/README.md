# Spectra-Red SurrealDB Schema

This directory contains the complete database schema and seed data for the Spectra-Red Intel Mesh.

## Files

- **schema.surql**: Complete database schema definition with all tables, fields, indices, and relationships
- **seed.surql**: Seed data including common ports, geographic data, ASNs, and sample vulnerabilities
- **README.md**: This file

## Schema Overview

### Core Asset Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `host` | IP addresses with geo/network metadata | ip, asn, city, country, cloud_region |
| `port` | Port numbers with protocol info | number, protocol, transport |
| `service` | Service identification | name, product, version, cpe, fingerprint |
| `banner` | Service banners (deduplicated by hash) | hash, sample |
| `tls_cert` | TLS/SSL certificates | sha256, cn, sans, not_before, not_after |

### Vulnerability Tables (Pro Tier)

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `vuln` | Core vulnerability metadata | cve_id, cvss, severity, kev_flag |
| `vuln_doc` | Extended vulnerability info for RAG | cve_id, title, summary, embedding (1536-dim vector) |

### Geography Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `city` | City-level geographic data | name, cc, lat, lon |
| `region` | State/province data | name, cc, code |
| `country` | Country data | cc (ISO 3166-1), name |
| `asn` | Autonomous System Numbers | number, org, type |
| `cloud_region` | Cloud provider regions | provider, code, name |
| `common_port` | Common port taxonomy | number, label, description |

### Relationship Edges

| Edge | From | To | Purpose |
|------|------|-----|---------|
| `HAS` | host | port | Host has open port |
| `RUNS` | port | service | Port runs service |
| `EVIDENCED_BY` | service | banner \| tls_cert | Service evidenced by banner/cert |
| `AFFECTED_BY` | service | vuln | Service affected by vulnerability |
| `IN_CITY` | host | city | Host located in city |
| `IN_REGION` | city | region | City in region |
| `IN_COUNTRY` | region | country | Region in country |
| `IN_ASN` | host | asn | Host belongs to ASN |
| `IN_CLOUD_REGION` | host | cloud_region | Host in cloud region |
| `IS_COMMON` | port | common_port | Port is well-known |
| `OBSERVED_AT` | service | ANY | Observation metadata with contributor info |

## Key Indices

### Performance-Critical Indices

- `idx_host_ip` (UNIQUE) - Fast host lookup by IP
- `idx_port_number` - Fast port queries
- `idx_service_fp` - Service deduplication by fingerprint
- `idx_vuln_cve` (UNIQUE) - Vulnerability lookup by CVE ID
- `idx_vuln_doc_embedding` (VECTOR COSINE) - Semantic vulnerability search

### Geographic Indices

- `idx_country_cc` (UNIQUE) - Country lookup by code
- `idx_asn_number` (UNIQUE) - ASN lookup
- `idx_cloud_region_code` (UNIQUE) - Cloud region lookup

## Seed Data Summary

The seed data includes:

- **28 Common Ports**: SSH (22), HTTP (80), HTTPS (443), MySQL (3306), Redis (6379), PostgreSQL (5432), MongoDB (27017), etc.
- **25 Countries**: Major tech hubs and cloud regions (US, GB, FR, DE, JP, CN, IN, SG, etc.)
- **20 ASNs**: Top cloud providers (AWS, GCP, Azure, DigitalOcean, Linode) and major ISPs
- **26 Cloud Regions**: AWS, GCP, and Azure regions worldwide
- **14 Regions**: US states and international regions
- **25 Cities**: Major tech hubs (San Francisco, New York, London, Tokyo, Singapore, etc.)
- **5 Vulnerabilities**: Recent high-impact CVEs (CVE-2024-3094, CVE-2024-21762, etc.)
- **3 Vuln Docs**: Extended vulnerability info with placeholder embeddings

**Total Seed Records**: 146

## Usage

### 1. Apply Schema and Seed Data

Using the provided setup script:

```bash
cd /Users/seanknowles/Projects/recon/.conductor/melbourne
./scripts/setup-db.sh
```

### 2. Manual Application

If you prefer manual control:

```bash
# Apply schema
surreal import \
  --conn http://localhost:8000 \
  --user root \
  --pass root \
  --ns spectra \
  --db intel \
  internal/db/schema/schema.surql

# Load seed data
surreal import \
  --conn http://localhost:8000 \
  --user root \
  --pass root \
  --ns spectra \
  --db intel \
  internal/db/schema/seed.surql
```

### 3. Using HTTP API

```bash
# Apply schema
curl -X POST http://localhost:8000/sql \
  -H "Accept: application/json" \
  -H "NS: spectra" \
  -H "DB: intel" \
  -u root:root \
  --data-binary @internal/db/schema/schema.surql

# Load seed data
curl -X POST http://localhost:8000/sql \
  -H "Accept: application/json" \
  -H "NS: spectra" \
  -H "DB: intel" \
  -u root:root \
  --data-binary @internal/db/schema/seed.surql
```

### 4. Verify Setup

Run the verification script to ensure everything is correctly configured:

```bash
./scripts/verify-schema.sh
```

This will:
- Check all tables exist
- Verify record counts
- Test key indices
- Validate specific seed data records

## Example Queries

### Find All Hosts in AWS us-east-1

```sql
SELECT host.* FROM host
WHERE id IN (
  SELECT in FROM IN_CLOUD_REGION
  WHERE out IN (
    SELECT id FROM cloud_region
    WHERE provider = 'aws' AND code = 'us-east-1'
  )
);
```

### Get All Services Running on Port 443

```sql
SELECT service.* FROM service
WHERE id IN (
  SELECT out FROM RUNS
  WHERE in IN (
    SELECT id FROM port WHERE number = 443
  )
);
```

### Find Critical Vulnerabilities

```sql
SELECT * FROM vuln
WHERE severity = 'critical'
ORDER BY cvss DESC;
```

### Semantic Vulnerability Search (Vector)

```sql
SELECT * FROM vuln_doc
WHERE vector::similarity(embedding, $query_embedding) > 0.85
ORDER BY vector::distance(embedding, $query_embedding) ASC
LIMIT 20;
```

### Get Host Graph (Ports, Services, Vulns)

```sql
SELECT *,
  ->HAS->port AS ports,
  ->HAS->port->RUNS->service AS services,
  ->HAS->port->RUNS->service->AFFECTED_BY->vuln AS vulnerabilities
FROM host
WHERE ip = '192.168.1.1';
```

## Schema Design Principles

### 1. SCHEMAFULL for Production

All tables use `SCHEMAFULL` mode for data integrity and validation.

### 2. Temporal Tracking

All entities track:
- `first_seen`: When first observed
- `last_seen`: Most recent observation
- Defaults to `time::now()`

### 3. Deduplication

- Services use SHA256 fingerprint for dedup
- Banners hashed to avoid duplicates
- TLS certs identified by SHA256

### 4. Vector Search Ready

- `vuln_doc.embedding`: 1536-dimensional vectors (OpenAI embeddings)
- COSINE distance for semantic similarity
- MTREE index for efficient vector search

### 5. Graph Relationships

- Rich relationship model with metadata
- Confidence scores on edges (e.g., service detection confidence)
- Observation tracking with contributor trust scores

## Performance Considerations

### Index Strategy

- UNIQUE indices on natural keys (IP, CVE ID, ASN number)
- Composite indices on frequently queried fields
- Vector index on embeddings for ML workloads

### Query Optimization

- Use indices for filtering
- Leverage graph traversal syntax (`->`, `<-`)
- Batch operations for bulk inserts
- Consider denormalization for hot paths

## Migration Notes

### Adding New Fields

```sql
DEFINE FIELD new_field ON TABLE host TYPE string;
```

### Adding New Indices

```sql
DEFINE INDEX idx_new_field ON TABLE host COLUMNS new_field;
```

### Schema Version Control

Track schema changes in version control. For major changes:
1. Create migration script (e.g., `migrations/001_add_field.surql`)
2. Test on staging environment
3. Apply to production with backup

## Troubleshooting

### Schema Application Fails

Check SurrealDB logs for syntax errors. Common issues:
- Field type mismatches
- Invalid ASSERT constraints
- Missing table dependencies

### Vector Index Issues

Ensure SurrealDB version >= 1.3.0 for vector index support.

### Seed Data Conflicts

If re-running seed data, unique constraints may fail. Either:
- Drop and recreate database
- Use `UPDATE` instead of `INSERT` for idempotent loading

## References

- [SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md](../../../SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md) - Full PRD with schema rationale
- [SURREALDB_SCHEMA_GUIDE.md](../../../SURREALDB_SCHEMA_GUIDE.md) - SurrealDB best practices
- [SurrealDB Documentation](https://surrealdb.com/docs) - Official docs

## Maintenance

### Backup

```bash
surreal export --conn http://localhost:8000 --user root --pass root --ns spectra --db intel > backup.sql
```

### Restore

```bash
surreal import --conn http://localhost:8000 --user root --pass root --ns spectra --db intel backup.sql
```

### Health Check

```bash
curl -X POST http://localhost:8000/sql \
  -H "NS: spectra" \
  -H "DB: intel" \
  -u root:root \
  -d "INFO FOR DB;"
```

---

**Schema Version**: 1.0.0
**Last Updated**: 2025-11-01
**Maintainer**: Spectra-Red Engineering Team
