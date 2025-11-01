# SurrealDB Schema Design Guide for Spectra-Red

## Overview

SurrealDB provides graph + vector capabilities perfect for threat intelligence:
- **Graph relationships**: Model threat connections, asset topology
- **Vector search**: Find semantically similar threats
- **Temporal queries**: Track changes over time
- **Full-text search**: Search threat descriptions

## Core Data Model

### Entities (Nodes)

```
Threat      - Individual threat/CVE
Asset       - IP, hostname, service
Vulnerability  - Specific vulnerability
Contributor - User submitting data
Scan        - Historical scan record
```

### Relationships (Edges)

```
targets:        Threat → Asset (threat affects asset)
exploits:       Threat → Vulnerability (threat uses vuln)
affects:        Vulnerability → Asset (vuln on asset)
correlates_with: Threat → Threat (related threats)
contains:       Network → Asset (asset in network)
contributed_by: Data → Contributor (submitter)
```

## Complete Schema Definition

```sql
-- ============================
-- THREAT INTELLIGENCE TABLES
-- ============================

DEFINE TABLE threat SCHEMAFULL
    COMMENT "Security threats, exploits, campaigns";

DEFINE FIELD id ON TABLE threat TYPE string;
DEFINE FIELD name ON TABLE threat TYPE string;
DEFINE FIELD description ON TABLE threat TYPE string;
DEFINE FIELD severity ON TABLE threat TYPE string 
    ENUM 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO';
DEFINE FIELD threat_type ON TABLE threat TYPE string 
    ENUM 'CVE', 'ZERO_DAY', 'APT', 'MALWARE', 'EXPLOIT';
DEFINE FIELD cwe_ids ON TABLE threat TYPE array;
DEFINE FIELD first_seen ON TABLE threat TYPE datetime;
DEFINE FIELD last_seen ON TABLE threat TYPE datetime;
DEFINE FIELD embedding ON TABLE threat TYPE array;  -- For vector search
DEFINE FIELD embedding_model ON TABLE threat TYPE string;
DEFINE FIELD references ON TABLE threat TYPE array;
DEFINE FIELD mitigations ON TABLE threat TYPE array;
DEFINE FIELD created_at ON TABLE threat TYPE datetime DEFAULT now();
DEFINE FIELD updated_at ON TABLE threat TYPE datetime DEFAULT now();

-- Indices for performance
DEFINE INDEX threat_severity_idx ON threat COLUMNS severity;
DEFINE INDEX threat_date_idx ON TABLE threat COLUMNS (first_seen, last_seen);
DEFINE INDEX threat_embedding_idx ON threat COLUMNS embedding VECTORIZE;
DEFINE INDEX threat_fulltext_idx ON threat COLUMNS description SEARCH ANALYZER ascii;

-- ============================
-- VULNERABILITY TABLES
-- ============================

DEFINE TABLE vulnerability SCHEMAFULL
    COMMENT "CVEs and vulnerability information";

DEFINE FIELD cve_id ON TABLE vulnerability TYPE string;
DEFINE FIELD title ON TABLE vulnerability TYPE string;
DEFINE FIELD description ON TABLE vulnerability TYPE string;
DEFINE FIELD cvss_v3 ON TABLE vulnerability TYPE number;
DEFINE FIELD cvss_v4 ON TABLE vulnerability TYPE number;
DEFINE FIELD cwe ON TABLE vulnerability TYPE array;
DEFINE FIELD affected_products ON TABLE vulnerability TYPE array;
DEFINE FIELD published_date ON TABLE vulnerability TYPE datetime;
DEFINE FIELD updated_date ON TABLE vulnerability TYPE datetime;
DEFINE FIELD embedding ON TABLE vulnerability TYPE array;

DEFINE INDEX vuln_cvss_idx ON vulnerability COLUMNS cvss_v3;
DEFINE INDEX vuln_published_idx ON vulnerability COLUMNS published_date;
DEFINE INDEX vuln_embedding_idx ON vulnerability COLUMNS embedding VECTORIZE;

-- ============================
-- ASSET TABLES
-- ============================

DEFINE TABLE asset SCHEMAFULL
    COMMENT "Network assets (IPs, services, applications)";

DEFINE FIELD ip_address ON TABLE asset TYPE string;
DEFINE FIELD hostname ON TABLE asset TYPE string;
DEFINE FIELD port ON TABLE asset TYPE number;
DEFINE FIELD service ON TABLE asset TYPE string;
DEFINE FIELD version ON TABLE asset TYPE string;
DEFINE FIELD location ON TABLE asset TYPE object;
    DEFINE FIELD location.country ON TABLE asset TYPE string;
    DEFINE FIELD location.region ON TABLE asset TYPE string;
    DEFINE FIELD location.latitude ON TABLE asset TYPE number;
    DEFINE FIELD location.longitude ON TABLE asset TYPE number;
DEFINE FIELD owner ON TABLE asset TYPE string;
DEFINE FIELD owner_email ON TABLE asset TYPE string;
DEFINE FIELD last_scanned ON TABLE asset TYPE datetime;
DEFINE FIELD scan_frequency ON TABLE asset TYPE string
    ENUM 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW';

DEFINE INDEX asset_ip_idx ON asset COLUMNS ip_address;
DEFINE INDEX asset_hostname_idx ON asset COLUMNS hostname;
DEFINE INDEX asset_owner_idx ON asset COLUMNS owner;
DEFINE INDEX asset_location_idx ON asset COLUMNS location.region;

-- ============================
-- NETWORK TOPOLOGY TABLES
-- ============================

DEFINE TABLE network SCHEMAFULL
    COMMENT "Network ranges and topologies";

DEFINE FIELD cidr ON TABLE network TYPE string;
DEFINE FIELD region ON TABLE network TYPE string;
DEFINE FIELD owner ON TABLE network TYPE string;
DEFINE FIELD created_at ON TABLE network TYPE datetime;

-- Relationships
DEFINE TABLE network_contains AS RELATION IN network OUT asset;

-- ============================
-- CONTRIBUTOR & COMMUNITY TABLES
-- ============================

DEFINE TABLE contributor SCHEMAFULL
    COMMENT "Community contributors";

DEFINE FIELD public_key ON TABLE contributor TYPE string;
DEFINE FIELD name ON TABLE contributor TYPE string;
DEFINE FIELD email ON TABLE contributor TYPE string;
DEFINE FIELD reputation_score ON TABLE contributor TYPE number;
DEFINE FIELD reputation_level ON TABLE contributor TYPE string
    ENUM 'BANNED', 'NOVICE', 'TRUSTED', 'EXPERT', 'CURATOR';
DEFINE FIELD submissions_count ON TABLE contributor TYPE number DEFAULT 0;
DEFINE FIELD submissions_accepted ON TABLE contributor TYPE number DEFAULT 0;
DEFINE FIELD joined_at ON TABLE contributor TYPE datetime DEFAULT now();

DEFINE INDEX contributor_key_idx ON contributor COLUMNS public_key;
DEFINE INDEX contributor_reputation_idx ON contributor COLUMNS reputation_score;

DEFINE TABLE contribution SCHEMAFULL
    COMMENT "Submitted threat intelligence";

DEFINE FIELD data ON TABLE contribution TYPE object;  -- The actual threat data
DEFINE FIELD signature ON TABLE contribution TYPE string;
DEFINE FIELD status ON TABLE contribution TYPE string
    ENUM 'PENDING', 'VERIFIED', 'ACCEPTED', 'REJECTED';
DEFINE FIELD confidence_score ON TABLE contribution TYPE number;
DEFINE FIELD peer_votes ON TABLE contribution TYPE number DEFAULT 0;
DEFINE FIELD created_at ON TABLE contribution TYPE datetime DEFAULT now();
DEFINE FIELD reviewed_at ON TABLE contribution TYPE datetime;

-- Relationships
DEFINE TABLE contributed_by AS RELATION IN contribution OUT contributor;

-- ============================
-- SCAN HISTORY TABLES
-- ============================

DEFINE TABLE scan_session SCHEMAFULL VERSION
    COMMENT "Historical scan records";

DEFINE FIELD campaign_id ON TABLE scan_session TYPE string;
DEFINE FIELD target ON TABLE scan_session TYPE string;
DEFINE FIELD scanner_type ON TABLE scan_session TYPE string;
DEFINE FIELD start_time ON TABLE scan_session TYPE datetime;
DEFINE FIELD end_time ON TABLE scan_session TYPE datetime;
DEFINE FIELD status ON TABLE scan_session TYPE string
    ENUM 'RUNNING', 'COMPLETED', 'FAILED';
DEFINE FIELD results_count ON TABLE scan_session TYPE number;
DEFINE FIELD scan_results ON TABLE scan_session TYPE array;

DEFINE INDEX scan_session_campaign_idx ON scan_session COLUMNS campaign_id;
DEFINE INDEX scan_session_target_idx ON scan_session COLUMNS target;

-- ============================
-- RELATIONSHIP DEFINITIONS
-- ============================

-- Threat targets assets
DEFINE TABLE targets AS RELATION IN threat OUT asset;

-- Threat exploits vulnerabilities  
DEFINE TABLE exploits AS RELATION IN threat OUT vulnerability;

-- Vulnerability affects assets
DEFINE TABLE affects AS RELATION IN vulnerability OUT asset;

-- Threats correlate with each other
DEFINE TABLE correlates_with AS RELATION IN threat OUT threat;

-- Contributor submitted data
DEFINE TABLE submitted AS RELATION IN contributor OUT contribution;

-- ============================
-- VECTOR INDEX CONFIGURATION
-- ============================

DEFINE ANALYZER threat_analyzer TOKENIZERS blank,lowercase,snowball FILTERS ascii,lowercase;
DEFINE INDEX threat_search ON threat COLUMNS description SEARCH ANALYZER threat_analyzer;

-- For semantic search
-- Requires SurrealDB v1.3+
DEFINE INDEX threat_vector_index ON threat COLUMNS embedding VECTORIZE;
DEFINE INDEX vuln_vector_index ON vulnerability COLUMNS embedding VECTORIZE;
```

## Common Query Patterns

### 1. Find All Threats Affecting a Region

```sql
SELECT DISTINCT in FROM targets 
WHERE out IN (
    SELECT id FROM asset 
    WHERE location.region = 'us-east-1'
)
AND in.severity = 'CRITICAL'
ORDER BY in.last_seen DESC;
```

### 2. Hybrid Search: Vector Similarity + Threat Context

```sql
SELECT * FROM threat
WHERE vector::similarity(embedding, $query_embedding) > 0.85
AND severity IN ['CRITICAL', 'HIGH']
ORDER BY vector::distance(embedding, $query_embedding) ASC
LIMIT 20;
```

### 3. Graph Expansion: Find Related Threats

```sql
-- Find threats that correlate with the primary threat
SELECT <-correlates_with<- FROM threat WHERE id = 'threat:ransomware-2025'
LIMIT 50;
```

### 4. Multi-hop Lateral Movement Risk

```sql
-- Find assets reachable from compromised asset via network topology
SELECT -> network_contains -> * -> (targets <-) <- 
FROM 'asset:compromised-ip'
DEPTH 3;
```

### 5. Vulnerability on Your Assets

```sql
SELECT 
    id,
    cve_id,
    title,
    cvss_v3,
    affected_products,
    (SELECT out FROM affects WHERE in = id) AS vulnerable_assets
FROM vulnerability
WHERE 'nginx-1.18' IN affected_products
ORDER BY cvss_v3 DESC;
```

### 6. Community Contribution Quality

```sql
-- Reputation-weighted threat data
SELECT 
    id,
    name,
    severity,
    (SELECT reputation_score FROM (SELECT out FROM submitted WHERE in = id)) AS contributor_reputation
FROM threat
WHERE id IN (SELECT in FROM submitted)
ORDER BY contributor_reputation DESC;
```

### 7. Time-Travel Query (version history)

```sql
-- Query threat data as it was on January 15
SELECT * FROM threat AT '2025-01-15T12:00:00Z'
WHERE severity = 'CRITICAL';
```

### 8. Geo-Distance Query

```sql
-- Find threats within 100 miles of New York
SELECT * FROM threat
WHERE id IN (
    SELECT in FROM targets
    WHERE out IN (
        SELECT id FROM asset
        WHERE math::distance(location.latitude, location.longitude, 40.7128, -74.0060) < 100
    )
);
```

## Performance Optimization

### Index Strategy

```sql
-- Query patterns determine indices

-- Pattern 1: Frequently filter by severity
DEFINE INDEX threat_severity ON threat COLUMNS severity;

-- Pattern 2: Time-range queries
DEFINE INDEX threat_dates ON threat COLUMNS (first_seen, last_seen);

-- Pattern 3: Full-text search
DEFINE INDEX threat_description ON threat COLUMNS description SEARCH ANALYZER ascii;

-- Pattern 4: Vector similarity
DEFINE INDEX threat_vector ON threat COLUMNS embedding VECTORIZE;

-- Pattern 5: Composite queries
DEFINE INDEX threat_severity_date ON threat COLUMNS (severity, last_seen);
```

### Query Optimization

```sql
-- SLOW: Complex join
SELECT * FROM threat
WHERE id IN (
    SELECT in FROM targets
    WHERE out IN (
        SELECT id FROM asset WHERE owner = 'customer-1'
    )
);

-- FAST: Denormalize relationship
DEFINE TABLE threat_for_customer AS RELATION IN threat OUT asset
WHERE asset.owner = 'customer-1';

SELECT * FROM threat_for_customer;
```

### Batch Operations

```sql
-- Batch insert
INSERT INTO asset [
    { ip_address: '192.168.1.1', hostname: 'web1' },
    { ip_address: '192.168.1.2', hostname: 'web2' },
    ...
];

-- Batch update
UPDATE asset SET last_scanned = now()
WHERE id IN ['asset:1', 'asset:2', ...];
```

## Scaling Strategies

### Partitioning by Region

```sql
-- Create region-specific tables
DEFINE TABLE threat_us SCHEMAFULL;
DEFINE TABLE threat_eu SCHEMAFULL;

-- Route queries to appropriate table
SELECT * FROM threat_us WHERE severity = 'CRITICAL';
```

### Compression for Old Data

```sql
-- Archive old scans
INSERT INTO scan_archive
SELECT * FROM scan_session
WHERE start_time < now() - 90days;

DELETE FROM scan_session
WHERE start_time < now() - 90days;
```

### Read Replicas

```yaml
# SurrealDB configuration
primary: "primary.db:8000"
replicas:
  - "replica1.db:8000"
  - "replica2.db:8000"
```

## Backup & Recovery

### Backup Strategy

```bash
# Daily full backup
surreal export /backup/surreal-$(date +%Y%m%d).sql

# Incremental backup (transaction log)
tail -f /var/lib/surrealdb/transaction.log > /backup/incremental.log
```

### Point-in-Time Recovery

```sql
-- Restore from backup
surreal import /backup/surreal-20250101.sql

-- Replay transactions to specific point
surreal replay --until '2025-01-15T12:30:00Z' /backup/incremental.log
```

## Access Control

```sql
-- Define roles
DEFINE ROLE viewer;
DEFINE ROLE analyst;
DEFINE ROLE curator;

-- Grant permissions
GRANT SELECT ON threat TO viewer;
GRANT SELECT, UPDATE ON threat TO analyst;
GRANT SELECT, UPDATE, DELETE ON threat TO curator;

-- Row-level security example
DEFINE TABLE threat_personal SCHEMAFULL
    PERMISSIONS
        FOR select WHERE owner = $auth.id
        FOR update WHERE owner = $auth.id
        FOR delete WHERE owner = $auth.id;
```

## Migration from Existing Databases

### From Traditional Relational DB

```sql
-- Convert relational threat data to SurrealDB
INSERT INTO threat
SELECT 
    'threat:' + cast(id as string) as id,
    name,
    description,
    severity,
    first_seen,
    last_seen,
    null as embedding
FROM legacy_threats;

-- Recreate relationships
INSERT INTO targets
SELECT 
    'threat:' + cast(threat_id as string) as in,
    'asset:' + cast(asset_id as string) as out
FROM legacy_threat_assets;
```

### Data Migration Validation

```sql
-- Verify record counts match
SELECT count() as legacy_threats FROM legacy_threats;
SELECT count() as new_threats FROM threat;

-- Spot check records
SELECT * FROM threat LIMIT 10;
```

---

**Key Principle**: SurrealDB's graph + vector capabilities enable sophisticated threat intelligence queries. Proper schema design and indexing strategy are critical for performance at scale.

