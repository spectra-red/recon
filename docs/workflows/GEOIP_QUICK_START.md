# GeoIP Enrichment - Quick Start Guide

## Overview

The GeoIP enrichment workflow provides geographic data enrichment for IP addresses in the Spectra-Red Intel Mesh. It uses MaxMind GeoLite2 MMDB files for fast, local lookups without rate limits.

## Prerequisites

### 1. MaxMind GeoLite2 Database

Download the free GeoLite2-City database:

1. **Register for free account:** https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
2. **Download GeoLite2-City.mmdb**
3. **Place in default location:**
   ```bash
   sudo mkdir -p /var/lib/GeoIP
   sudo mv GeoLite2-City.mmdb /var/lib/GeoIP/
   ```

**Alternative locations:**
```bash
# Custom path via environment variable
export GEOIP_MMDB_PATH=/path/to/GeoLite2-City.mmdb
```

### 2. Services Running

```bash
# Start SurrealDB
docker run -d -p 8000:8000 \
  surrealdb/surrealdb:latest \
  start --user root --pass root memory

# Start Restate
docker run -d -p 8080:8080 -p 9070:9070 \
  restatedev/restate:latest
```

## Quick Start

### 1. Start Workflow Service

```bash
# Set environment variables
export SURREALDB_URL=ws://localhost:8000/rpc
export GEOIP_MMDB_PATH=/var/lib/GeoIP/GeoLite2-City.mmdb

# Start workflow service
go run cmd/workflows/main.go
```

### 2. Register Workflows with Restate

```bash
# Using Restate CLI
restate deployments register http://localhost:9080

# Or using curl
curl -X POST http://localhost:9070/deployments \
  -H 'Content-Type: application/json' \
  -d '{"uri": "http://localhost:9080"}'
```

### 3. Prepare Test Data

```bash
# Create test hosts in SurrealDB
surreal sql --endpoint ws://localhost:8000 \
  --namespace spectra --database intel_mesh \
  --username root --password root << EOF

-- Create some test hosts
CREATE host:8_8_8_8 SET ip = '8.8.8.8', last_seen = time::now();
CREATE host:1_1_1_1 SET ip = '1.1.1.1', last_seen = time::now();
CREATE host:208_67_222_222 SET ip = '208.67.222.222', last_seen = time::now();

EOF
```

### 4. Invoke GeoIP Enrichment

```bash
# Using curl
curl -X POST http://localhost:8080/EnrichGeoWorkflow/Run \
  -H 'Content-Type: application/json' \
  -d '{
    "ips": [
      "8.8.8.8",
      "1.1.1.1",
      "208.67.222.222"
    ]
  }'
```

**Expected Response:**
```json
{
  "enriched": 3,
  "failed": 0,
  "errors": []
}
```

### 5. Verify Enrichment

```bash
surreal sql --endpoint ws://localhost:8000 \
  --namespace spectra --database intel_mesh \
  --username root --password root << EOF

-- Check host enrichment
SELECT * FROM host WHERE ip = '8.8.8.8';

-- Check geographic nodes created
SELECT * FROM city LIMIT 5;
SELECT * FROM region LIMIT 5;
SELECT * FROM country LIMIT 5;

-- Check relationships
SELECT
  *,
  ->IN_CITY->city.* AS city_info
FROM host
WHERE ip = '8.8.8.8'
FETCH city_info;

EOF
```

## Usage Examples

### Example 1: Enrich Single Host After Scan

```bash
# After ingesting scan data, enrich with geo data
curl -X POST http://localhost:8080/EnrichGeoWorkflow/Run \
  -H 'Content-Type: application/json' \
  -d '{"ips": ["203.0.113.42"]}'
```

### Example 2: Batch Enrichment

```bash
# Enrich multiple hosts at once (up to 100 per batch)
curl -X POST http://localhost:8080/EnrichGeoWorkflow/Run \
  -H 'Content-Type: application/json' \
  -d '{
    "ips": [
      "8.8.8.8",
      "8.8.4.4",
      "1.1.1.1",
      "1.0.0.1",
      "208.67.222.222",
      "208.67.220.220"
    ]
  }'
```

### Example 3: Query Enriched Data

```sql
-- Find all hosts in a specific city
SELECT * FROM host
WHERE city = 'Mountain View';

-- Find hosts by country
SELECT * FROM host
WHERE country = 'United States';

-- Get full geographic hierarchy
SELECT
    *,
    ->IN_CITY->city.* AS city,
    ->IN_CITY->city->IN_REGION->region.* AS region,
    ->IN_CITY->city->IN_REGION->region->IN_COUNTRY->country.* AS country
FROM host
WHERE ip = '8.8.8.8'
FETCH city, region, country;

-- Count hosts by country
SELECT
    country,
    count() as host_count
FROM host
GROUP BY country
ORDER BY host_count DESC;
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GEOIP_MMDB_PATH` | `/var/lib/GeoIP/GeoLite2-City.mmdb` | Path to MaxMind MMDB file |
| `GEOIP_API_KEY` | (empty) | Optional ipinfo.io API key for fallback |
| `SURREALDB_URL` | `ws://localhost:8000/rpc` | SurrealDB connection URL |
| `SURREALDB_NAMESPACE` | `spectra` | SurrealDB namespace |
| `SURREALDB_DATABASE` | `intel_mesh` | SurrealDB database name |

## Performance Tips

### 1. Use Local MMDB for Best Performance

- **MMDB lookup:** <10ms per IP
- **API lookup:** 100-500ms per IP
- **Batch of 100 IPs:** ~2-5 seconds total (MMDB)

### 2. Batch Processing

Process IPs in batches of 100 for optimal performance:

```bash
# Good: Batch of 100
curl -X POST http://localhost:8080/EnrichGeoWorkflow/Run \
  -d '{"ips": [...100 IPs...]}'

# Avoid: Single IP at a time
for ip in $ips; do
  curl -X POST http://localhost:8080/EnrichGeoWorkflow/Run \
    -d "{\"ips\": [\"$ip\"]}"
done
```

### 3. Filter Private IPs

MaxMind doesn't provide data for private IPs. Filter them before enrichment:

```bash
# Filter out RFC 1918 addresses
ips=$(echo "$all_ips" | grep -v '^10\.' | grep -v '^192\.168\.' | grep -v '^172\.(1[6-9]|2[0-9]|3[01])\.')
```

## Troubleshooting

### Issue: "MMDB file not found"

**Solution:**
```bash
# Check if file exists
ls -lh /var/lib/GeoIP/GeoLite2-City.mmdb

# If missing, download from MaxMind
# Set custom path if needed
export GEOIP_MMDB_PATH=/custom/path/GeoLite2-City.mmdb
```

### Issue: "No successful GeoIP lookups"

**Causes:**
1. All IPs are private (RFC 1918)
2. MMDB file corrupted
3. Invalid IP addresses

**Solution:**
```bash
# Validate MMDB file
go run cmd/workflows/main.go 2>&1 | grep "GeoIP"

# Check logs for specific errors
# Test with known public IP
curl -X POST http://localhost:8080/EnrichGeoWorkflow/Run \
  -d '{"ips": ["8.8.8.8"]}'
```

### Issue: Workflow fails with database error

**Solution:**
```bash
# Verify SurrealDB connection
surreal sql --endpoint ws://localhost:8000 \
  --namespace spectra --database intel_mesh \
  --username root --password root << EOF
INFO FOR DB;
EOF

# Check schema is applied
# Re-apply schema if needed
surreal import --endpoint ws://localhost:8000 \
  --namespace spectra --database intel_mesh \
  --username root --password root \
  internal/db/schema/schema.surql
```

## Data Model

### GeoIPInfo Structure
```go
type GeoIPInfo struct {
    IP        string  `json:"ip"`
    City      string  `json:"city"`
    Region    string  `json:"region"`
    Country   string  `json:"country"`
    CountryCC string  `json:"country_cc"` // ISO 3166-1 alpha-2
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}
```

### SurrealDB Schema

**City:**
```sql
CREATE city:$city_id SET
  name = $name,
  cc = $country_code,
  lat = $latitude,
  lon = $longitude;
```

**Region:**
```sql
CREATE region:$region_id SET
  name = $name,
  cc = $country_code,
  code = $region_code;
```

**Country:**
```sql
CREATE country:$cc SET
  cc = $country_code,
  name = $country_name;
```

**Relationships:**
```sql
RELATE host->IN_CITY->city;
RELATE city->IN_REGION->region;
RELATE region->IN_COUNTRY->country;
```

## Next Steps

1. **Combine with ASN Enrichment:** Run M5-T1 ASN workflow for complete network context
2. **Query API Integration:** Use geographic filters in search queries
3. **Scheduled Re-enrichment:** Periodically update stale geographic data
4. **Analytics:** Aggregate threats by geographic region

## Resources

- **MaxMind GeoLite2:** https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
- **Implementation Details:** M5-T2_GEOIP_ENRICHMENT_COMPLETION.md
- **Architecture:** DETAILED_IMPLEMENTATION_PLAN.md (Milestone 5, Task 2)
- **Restate Documentation:** https://docs.restate.dev
