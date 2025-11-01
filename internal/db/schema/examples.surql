-- ============================================================================
-- Spectra-Red Intel Mesh - Example Queries
-- ============================================================================
-- This file contains example queries for common use cases
-- ============================================================================

-- ============================================================================
-- BASIC QUERIES
-- ============================================================================

-- Get all common ports
SELECT * FROM common_port ORDER BY number;

-- Find specific port by label
SELECT * FROM common_port WHERE label = 'ssh';

-- Count total countries
SELECT count() FROM country GROUP ALL;

-- List all cloud providers
SELECT DISTINCT provider FROM cloud_region ORDER BY provider;

-- ============================================================================
-- GEOGRAPHIC QUERIES
-- ============================================================================

-- Find all AWS regions
SELECT * FROM cloud_region WHERE provider = 'aws' ORDER BY code;

-- Get ASNs by organization
SELECT * FROM asn WHERE org LIKE '%Amazon%';

-- List cities by country
SELECT * FROM city WHERE cc = 'US' ORDER BY name;

-- Find hosts in specific country
SELECT host.* FROM host
WHERE country = 'US'
LIMIT 10;

-- ============================================================================
-- VULNERABILITY QUERIES
-- ============================================================================

-- Find critical vulnerabilities
SELECT * FROM vuln
WHERE severity = 'critical'
ORDER BY cvss DESC;

-- Get vulnerabilities with CISA KEV flag
SELECT * FROM vuln
WHERE kev_flag = true
ORDER BY cvss DESC;

-- Find vulnerabilities by CVSS range
SELECT * FROM vuln
WHERE cvss >= 9.0 AND cvss <= 10.0
ORDER BY cvss DESC;

-- Get vulnerability details with extended info
SELECT
    v.cve_id,
    v.cvss,
    v.severity,
    vd.title,
    vd.summary,
    vd.epss
FROM vuln AS v
LEFT JOIN vuln_doc AS vd ON v.cve_id = vd.cve_id
WHERE v.severity = 'critical';

-- ============================================================================
-- GRAPH TRAVERSAL QUERIES
-- ============================================================================

-- Find all ports for a specific host
-- (Assumes host record exists with IP 192.168.1.1)
SELECT ->HAS->port AS open_ports
FROM host
WHERE ip = '192.168.1.1';

-- Get services running on a host
-- (This is a 2-hop graph traversal: host -> port -> service)
SELECT ->HAS->port->RUNS->service AS running_services
FROM host
WHERE ip = '192.168.1.1';

-- Find vulnerabilities affecting a host
-- (3-hop: host -> port -> service -> vuln)
SELECT ->HAS->port->RUNS->service->AFFECTED_BY->vuln AS vulnerabilities
FROM host
WHERE ip = '192.168.1.1';

-- Get full host context (ports, services, vulns in one query)
SELECT *,
    ->HAS->port AS ports,
    ->HAS->port->RUNS->service AS services,
    ->HAS->port->RUNS->service->AFFECTED_BY->vuln AS vulnerabilities
FROM host
WHERE ip = '192.168.1.1';

-- Find all hosts running a specific service
SELECT <-RUNS<-port<-HAS<-host AS affected_hosts
FROM service
WHERE name = 'nginx' AND version LIKE '1.18%';

-- ============================================================================
-- SERVICE IDENTIFICATION QUERIES
-- ============================================================================

-- Find services by name
SELECT * FROM service
WHERE name = 'http'
LIMIT 10;

-- Find services by product and version
SELECT * FROM service
WHERE product = 'nginx' AND version = '1.25.1';

-- Find services by CPE
SELECT * FROM service
WHERE 'cpe:2.3:a:nginx:nginx:1.25.1:*:*:*:*:*:*:*' IN cpe;

-- Get service with evidence (banners/certs)
SELECT *,
    ->EVIDENCED_BY->banner AS banners,
    ->EVIDENCED_BY->tls_cert AS tls_certs
FROM service
WHERE product = 'nginx'
LIMIT 10;

-- ============================================================================
-- PORT ANALYSIS QUERIES
-- ============================================================================

-- Find all instances of common port (e.g., SSH)
SELECT port.* FROM port
WHERE id IN (
    SELECT in FROM IS_COMMON
    WHERE out IN (
        SELECT id FROM common_port WHERE label = 'ssh'
    )
);

-- Count ports by protocol
SELECT protocol, count() AS total
FROM port
GROUP BY protocol;

-- Find non-standard ports (high port numbers)
SELECT * FROM port
WHERE number > 10000
ORDER BY number;

-- ============================================================================
-- NETWORK TOPOLOGY QUERIES
-- ============================================================================

-- Find all hosts in a specific ASN
SELECT host.* FROM host
WHERE id IN (
    SELECT in FROM IN_ASN
    WHERE out IN (
        SELECT id FROM asn WHERE number = 16509
    )
);

-- Get hosts by cloud provider
SELECT host.* FROM host
WHERE id IN (
    SELECT in FROM IN_CLOUD_REGION
    WHERE out IN (
        SELECT id FROM cloud_region WHERE provider = 'aws'
    )
)
LIMIT 10;

-- Find hosts in specific city
SELECT host.* FROM host
WHERE id IN (
    SELECT in FROM IN_CITY
    WHERE out IN (
        SELECT id FROM city WHERE name = 'San Francisco'
    )
);

-- ============================================================================
-- TEMPORAL QUERIES
-- ============================================================================

-- Find hosts scanned in last 24 hours
SELECT * FROM host
WHERE last_scanned_at > time::now() - 24h
ORDER BY last_scanned_at DESC;

-- Find stale hosts (not scanned in 7 days)
SELECT * FROM host
WHERE last_scanned_at < time::now() - 7d
ORDER BY last_scanned_at ASC
LIMIT 100;

-- Get recently discovered services
SELECT * FROM service
WHERE first_seen > time::now() - 1h
ORDER BY first_seen DESC;

-- ============================================================================
-- STATISTICS AND AGGREGATIONS
-- ============================================================================

-- Count hosts by country
SELECT country, count() AS total
FROM host
GROUP BY country
ORDER BY total DESC;

-- Count services by name
SELECT name, count() AS total
FROM service
GROUP BY name
ORDER BY total DESC
LIMIT 20;

-- Count vulnerabilities by severity
SELECT severity, count() AS total
FROM vuln
GROUP BY severity
ORDER BY
    CASE severity
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
    END;

-- Average CVSS score
SELECT math::mean(cvss) AS avg_cvss,
       math::max(cvss) AS max_cvss,
       math::min(cvss) AS min_cvss
FROM vuln;

-- ============================================================================
-- BANNER AND CERTIFICATE QUERIES
-- ============================================================================

-- Find banners by hash
SELECT * FROM banner
WHERE hash = 'abc123...';

-- Get TLS certificates expiring soon (within 30 days)
SELECT * FROM tls_cert
WHERE not_after < time::now() + 30d
ORDER BY not_after ASC;

-- Find expired certificates
SELECT * FROM tls_cert
WHERE not_after < time::now()
ORDER BY not_after DESC;

-- Find certificates by common name
SELECT * FROM tls_cert
WHERE cn LIKE '%.example.com'
LIMIT 10;

-- ============================================================================
-- OBSERVATION METADATA QUERIES
-- ============================================================================

-- Find observations by contributor
SELECT * FROM OBSERVED_AT
WHERE contributor_id = 'contributor:abc123'
ORDER BY ts DESC
LIMIT 100;

-- Get high-trust observations
SELECT * FROM OBSERVED_AT
WHERE trust >= 0.8
ORDER BY ts DESC;

-- Find recent observations (last hour)
SELECT * FROM OBSERVED_AT
WHERE ts > time::now() - 1h
ORDER BY ts DESC;

-- ============================================================================
-- ADVANCED GRAPH QUERIES
-- ============================================================================

-- Multi-hop traversal: Find all services affected by specific CVE
SELECT service.* FROM service
WHERE id IN (
    SELECT in FROM AFFECTED_BY
    WHERE out IN (
        SELECT id FROM vuln WHERE cve_id = 'CVE-2024-3094'
    )
);

-- Reverse lookup: Which hosts run vulnerable services?
SELECT DISTINCT host.* FROM host
WHERE id IN (
    SELECT in FROM HAS
    WHERE out IN (
        SELECT in FROM RUNS
        WHERE out IN (
            SELECT in FROM AFFECTED_BY
            WHERE out IN (
                SELECT id FROM vuln WHERE severity = 'critical'
            )
        )
    )
);

-- ============================================================================
-- VECTOR SEARCH QUERIES (Requires embeddings)
-- ============================================================================

-- Semantic vulnerability search by embedding similarity
-- Note: $query_embedding would be a 1536-dim vector from your application
-- SELECT * FROM vuln_doc
-- WHERE vector::similarity(embedding, $query_embedding) > 0.85
-- ORDER BY vector::distance(embedding, $query_embedding) ASC
-- LIMIT 20;

-- Hybrid search: Vector + CVSS filtering
-- SELECT * FROM vuln_doc
-- WHERE vector::similarity(embedding, $query_embedding) > 0.80
--   AND cvss >= 7.0
-- ORDER BY vector::distance(embedding, $query_embedding) ASC
-- LIMIT 10;

-- ============================================================================
-- DATA MODIFICATION EXAMPLES
-- ============================================================================

-- Insert a new host
-- INSERT INTO host {
--     ip: '192.168.1.100',
--     asn: 16509,
--     country: 'US',
--     first_seen: time::now(),
--     last_seen: time::now()
-- };

-- Update host last_scanned_at
-- UPDATE host SET last_scanned_at = time::now()
-- WHERE ip = '192.168.1.100';

-- Create relationship: host has port
-- RELATE (SELECT id FROM host WHERE ip = '192.168.1.100')->HAS->(SELECT id FROM port WHERE number = 443)
-- SET first_seen = time::now(), last_seen = time::now();

-- Delete old observations (cleanup)
-- DELETE FROM OBSERVED_AT WHERE ts < time::now() - 90d;

-- ============================================================================
-- SCHEMA INSPECTION
-- ============================================================================

-- Get database information
INFO FOR DB;

-- Get table schema
INFO FOR TABLE host;

-- List all tables
SELECT name FROM $tables;

-- List all indices on a table
SELECT * FROM $indexes WHERE table = 'host';

-- ============================================================================
-- PERFORMANCE TESTING
-- ============================================================================

-- Explain query plan (if supported)
-- EXPLAIN SELECT * FROM host WHERE ip = '192.168.1.1';

-- Count total records
SELECT
    (SELECT count() FROM host) AS hosts,
    (SELECT count() FROM port) AS ports,
    (SELECT count() FROM service) AS services,
    (SELECT count() FROM vuln) AS vulns;

-- ============================================================================
-- END OF EXAMPLES
-- ============================================================================
