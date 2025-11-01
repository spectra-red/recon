-- ============================================================================
-- Spectra-Red Intel Mesh - SurrealDB Schema Definition
-- ============================================================================
-- This schema defines the complete data model for the threat intelligence mesh
-- including hosts, ports, services, vulnerabilities, and geographic relationships.
--
-- Schema Version: 1.0.0
-- SurrealDB Version: >=1.3.0 (requires vector index support)
-- ============================================================================

-- ============================================================================
-- CORE ASSET TABLES
-- ============================================================================

-- Host: IP addresses with geo/network metadata
DEFINE TABLE host SCHEMAFULL;
DEFINE FIELD ip ON TABLE host TYPE string ASSERT $value != NONE;
DEFINE FIELD asn ON TABLE host TYPE int;
DEFINE FIELD city ON TABLE host TYPE string;
DEFINE FIELD region ON TABLE host TYPE string;
DEFINE FIELD country ON TABLE host TYPE string;
DEFINE FIELD cloud_region ON TABLE host TYPE string;
DEFINE FIELD first_seen ON TABLE host TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE host TYPE datetime DEFAULT time::now();
DEFINE FIELD last_scanned_at ON TABLE host TYPE datetime;
DEFINE INDEX idx_host_ip ON TABLE host COLUMNS ip UNIQUE;
DEFINE INDEX idx_host_asn ON TABLE host COLUMNS asn;
DEFINE INDEX idx_host_country ON TABLE host COLUMNS country;
DEFINE INDEX idx_host_last_scanned ON TABLE host COLUMNS last_scanned_at;

-- Port: Port numbers with protocol and transport info
DEFINE TABLE port SCHEMAFULL;
DEFINE FIELD number ON TABLE port TYPE int ASSERT $value > 0 AND $value < 65536;
DEFINE FIELD protocol ON TABLE port TYPE string ASSERT $value IN ['tcp', 'udp'];
DEFINE FIELD transport ON TABLE port TYPE string; -- e.g., 'tls', 'plain'
DEFINE FIELD common ON TABLE port TYPE bool DEFAULT false;
DEFINE FIELD first_seen ON TABLE port TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE port TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_port_number ON TABLE port COLUMNS number;
DEFINE INDEX idx_port_protocol ON TABLE port COLUMNS protocol;

-- Service: Service identification (name, product, version, CPE)
DEFINE TABLE service SCHEMAFULL;
DEFINE FIELD name ON TABLE service TYPE string; -- e.g., 'http', 'ssh'
DEFINE FIELD product ON TABLE service TYPE string; -- e.g., 'nginx', 'openssh'
DEFINE FIELD version ON TABLE service TYPE string; -- e.g., '1.25.1'
DEFINE FIELD cpe ON TABLE service TYPE array<string>; -- CPE 2.3 identifiers
DEFINE FIELD fingerprint ON TABLE service TYPE string; -- SHA256 hash for dedup
DEFINE FIELD first_seen ON TABLE service TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE service TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_service_fp ON TABLE service COLUMNS fingerprint;
DEFINE INDEX idx_service_name ON TABLE service COLUMNS name;
DEFINE INDEX idx_service_product ON TABLE service COLUMNS product;

-- Banner: Service banners (hashed for deduplication)
DEFINE TABLE banner SCHEMAFULL;
DEFINE FIELD hash ON TABLE banner TYPE string ASSERT $value != NONE;
DEFINE FIELD sample ON TABLE banner TYPE string; -- max 2KB sample
DEFINE FIELD first_seen ON TABLE banner TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_banner_hash ON TABLE banner COLUMNS hash UNIQUE;

-- TLS Certificate: TLS/SSL certificate metadata
DEFINE TABLE tls_cert SCHEMAFULL;
DEFINE FIELD sha256 ON TABLE tls_cert TYPE string ASSERT $value != NONE;
DEFINE FIELD cn ON TABLE tls_cert TYPE string; -- common name
DEFINE FIELD sans ON TABLE tls_cert TYPE array<string>; -- subject alt names
DEFINE FIELD not_before ON TABLE tls_cert TYPE datetime;
DEFINE FIELD not_after ON TABLE tls_cert TYPE datetime;
DEFINE FIELD first_seen ON TABLE tls_cert TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_tls_sha256 ON TABLE tls_cert COLUMNS sha256 UNIQUE;
DEFINE INDEX idx_tls_cn ON TABLE tls_cert COLUMNS cn;
DEFINE INDEX idx_tls_expiry ON TABLE tls_cert COLUMNS not_after;

-- ============================================================================
-- VULNERABILITY TABLES (PRO TIER)
-- ============================================================================

-- Vuln: Core vulnerability metadata (CVE, CVSS, severity)
DEFINE TABLE vuln SCHEMAFULL;
DEFINE FIELD cve_id ON TABLE vuln TYPE string ASSERT $value != NONE;
DEFINE FIELD cvss ON TABLE vuln TYPE float;
DEFINE FIELD severity ON TABLE vuln TYPE string; -- 'critical', 'high', 'medium', 'low'
DEFINE FIELD kev_flag ON TABLE vuln TYPE bool DEFAULT false; -- CISA known exploited
DEFINE FIELD first_seen ON TABLE vuln TYPE datetime DEFAULT time::now();
DEFINE FIELD last_updated ON TABLE vuln TYPE datetime DEFAULT time::now();
DEFINE INDEX idx_vuln_cve ON TABLE vuln COLUMNS cve_id UNIQUE;
DEFINE INDEX idx_vuln_severity ON TABLE vuln COLUMNS severity;
DEFINE INDEX idx_vuln_cvss ON TABLE vuln COLUMNS cvss;
DEFINE INDEX idx_vuln_kev ON TABLE vuln COLUMNS kev_flag;

-- Vuln Doc: Extended vulnerability info for RAG (vector search)
DEFINE TABLE vuln_doc SCHEMAFULL;
DEFINE FIELD cve_id ON TABLE vuln_doc TYPE string ASSERT $value != NONE;
DEFINE FIELD title ON TABLE vuln_doc TYPE string;
DEFINE FIELD summary ON TABLE vuln_doc TYPE string;
DEFINE FIELD cvss ON TABLE vuln_doc TYPE float;
DEFINE FIELD epss ON TABLE vuln_doc TYPE float; -- exploit prediction score
DEFINE FIELD cpe ON TABLE vuln_doc TYPE array<string>;
DEFINE FIELD exploit_refs ON TABLE vuln_doc TYPE array<string>; -- URLs
DEFINE FIELD embedding ON TABLE vuln_doc TYPE array<float>; -- 1536 dims for OpenAI
DEFINE FIELD published_date ON TABLE vuln_doc TYPE datetime;
DEFINE FIELD last_modified ON TABLE vuln_doc TYPE datetime;
DEFINE INDEX idx_vuln_doc_cve ON TABLE vuln_doc COLUMNS cve_id UNIQUE;
DEFINE INDEX idx_vuln_doc_cvss ON TABLE vuln_doc COLUMNS cvss;
DEFINE INDEX idx_vuln_doc_epss ON TABLE vuln_doc COLUMNS epss;
-- Vector index for semantic search (cosine similarity)
DEFINE INDEX idx_vuln_doc_embedding ON TABLE vuln_doc COLUMNS embedding MTREE DIMENSION 1536 DIST COSINE;

-- ============================================================================
-- GEOGRAPHY AND TAXONOMY TABLES
-- ============================================================================

-- City: City-level geographic data
DEFINE TABLE city SCHEMAFULL;
DEFINE FIELD name ON TABLE city TYPE string;
DEFINE FIELD cc ON TABLE city TYPE string; -- country code (ISO 3166-1 alpha-2)
DEFINE FIELD lat ON TABLE city TYPE float;
DEFINE FIELD lon ON TABLE city TYPE float;
DEFINE INDEX idx_city_name ON TABLE city COLUMNS name;
DEFINE INDEX idx_city_cc ON TABLE city COLUMNS cc;

-- Region: State/province-level geographic data
DEFINE TABLE region SCHEMAFULL;
DEFINE FIELD name ON TABLE region TYPE string;
DEFINE FIELD cc ON TABLE region TYPE string; -- country code
DEFINE FIELD code ON TABLE region TYPE string; -- region code (e.g., 'CA' for California)
DEFINE INDEX idx_region_name ON TABLE region COLUMNS name;
DEFINE INDEX idx_region_cc ON TABLE region COLUMNS cc;

-- Country: Country-level geographic data
DEFINE TABLE country SCHEMAFULL;
DEFINE FIELD cc ON TABLE country TYPE string ASSERT $value != NONE; -- ISO 3166-1 alpha-2
DEFINE FIELD name ON TABLE country TYPE string;
DEFINE INDEX idx_country_cc ON TABLE country COLUMNS cc UNIQUE;

-- ASN: Autonomous System Number data
DEFINE TABLE asn SCHEMAFULL;
DEFINE FIELD number ON TABLE asn TYPE int ASSERT $value > 0;
DEFINE FIELD org ON TABLE asn TYPE string;
DEFINE FIELD country ON TABLE asn TYPE string;
DEFINE FIELD type ON TABLE asn TYPE string; -- 'hosting', 'isp', 'enterprise', 'cloud'
DEFINE INDEX idx_asn_number ON TABLE asn COLUMNS number UNIQUE;
DEFINE INDEX idx_asn_org ON TABLE asn COLUMNS org;

-- Cloud Region: Cloud provider region metadata
DEFINE TABLE cloud_region SCHEMAFULL;
DEFINE FIELD provider ON TABLE cloud_region TYPE string; -- 'aws', 'gcp', 'azure', 'digitalocean', 'linode'
DEFINE FIELD code ON TABLE cloud_region TYPE string; -- 'us-east-1', 'europe-west1'
DEFINE FIELD name ON TABLE cloud_region TYPE string; -- Human-readable name
DEFINE FIELD city ON TABLE cloud_region TYPE string;
DEFINE FIELD country ON TABLE cloud_region TYPE string;
DEFINE INDEX idx_cloud_region_code ON TABLE cloud_region COLUMNS provider, code UNIQUE;
DEFINE INDEX idx_cloud_region_provider ON TABLE cloud_region COLUMNS provider;

-- Common Port: Common port taxonomy (well-known services)
DEFINE TABLE common_port SCHEMAFULL;
DEFINE FIELD number ON TABLE common_port TYPE int;
DEFINE FIELD label ON TABLE common_port TYPE string; -- 'http', 'https', 'ssh', 'mysql'
DEFINE FIELD description ON TABLE common_port TYPE string;
DEFINE INDEX idx_common_port_number ON TABLE common_port COLUMNS number UNIQUE;

-- ============================================================================
-- RELATIONSHIP EDGES
-- ============================================================================

-- HAS: host → port (host has open port)
DEFINE TABLE HAS SCHEMAFULL TYPE RELATION FROM host TO port;
DEFINE FIELD first_seen ON TABLE HAS TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE HAS TYPE datetime DEFAULT time::now();

-- RUNS: port → service (port runs service)
DEFINE TABLE RUNS SCHEMAFULL TYPE RELATION FROM port TO service;
DEFINE FIELD confidence ON TABLE RUNS TYPE float DEFAULT 1.0;
DEFINE FIELD first_seen ON TABLE RUNS TYPE datetime DEFAULT time::now();
DEFINE FIELD last_seen ON TABLE RUNS TYPE datetime DEFAULT time::now();

-- EVIDENCED_BY: service → banner | tls_cert (service evidenced by banner/cert)
DEFINE TABLE EVIDENCED_BY SCHEMAFULL TYPE RELATION FROM service TO banner | tls_cert;
DEFINE FIELD evidence_type ON TABLE EVIDENCED_BY TYPE string; -- 'banner', 'tls_cert'
DEFINE FIELD first_seen ON TABLE EVIDENCED_BY TYPE datetime DEFAULT time::now();

-- AFFECTED_BY: service → vuln (service affected by vulnerability)
DEFINE TABLE AFFECTED_BY SCHEMAFULL TYPE RELATION FROM service TO vuln;
DEFINE FIELD confidence ON TABLE AFFECTED_BY TYPE float DEFAULT 1.0;
DEFINE FIELD first_detected ON TABLE AFFECTED_BY TYPE datetime DEFAULT time::now();
DEFINE FIELD last_confirmed ON TABLE AFFECTED_BY TYPE datetime DEFAULT time::now();

-- IN_CITY: host → city (host located in city)
DEFINE TABLE IN_CITY SCHEMAFULL TYPE RELATION FROM host TO city;

-- IN_REGION: city → region (city in region)
DEFINE TABLE IN_REGION SCHEMAFULL TYPE RELATION FROM city TO region;

-- IN_COUNTRY: region → country (region in country)
DEFINE TABLE IN_COUNTRY SCHEMAFULL TYPE RELATION FROM region TO country;

-- IN_ASN: host → asn (host belongs to ASN)
DEFINE TABLE IN_ASN SCHEMAFULL TYPE RELATION FROM host TO asn;

-- IN_CLOUD_REGION: host → cloud_region (host in cloud region)
DEFINE TABLE IN_CLOUD_REGION SCHEMAFULL TYPE RELATION FROM host TO cloud_region;

-- IS_COMMON: port → common_port (port is common/well-known)
DEFINE TABLE IS_COMMON SCHEMAFULL TYPE RELATION FROM port TO common_port;

-- OBSERVED_AT: service → ANY (observation metadata with contributor info)
DEFINE TABLE OBSERVED_AT SCHEMAFULL TYPE RELATION FROM service TO ANY;
DEFINE FIELD scan_id ON TABLE OBSERVED_AT TYPE string;
DEFINE FIELD contributor_id ON TABLE OBSERVED_AT TYPE string;
DEFINE FIELD ts ON TABLE OBSERVED_AT TYPE datetime DEFAULT time::now();
DEFINE FIELD trust ON TABLE OBSERVED_AT TYPE float DEFAULT 1.0; -- trust score 0.0-1.0

-- ============================================================================
-- JOB TRACKING TABLES
-- ============================================================================

-- Job: Ingest job tracking with state machine
DEFINE TABLE job SCHEMAFULL;
DEFINE FIELD id ON TABLE job TYPE string ASSERT $value != NONE;
DEFINE FIELD scanner_key ON TABLE job TYPE string ASSERT $value != NONE;
DEFINE FIELD state ON TABLE job TYPE string ASSERT $value IN ['pending', 'processing', 'completed', 'failed'];
DEFINE FIELD created_at ON TABLE job TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_at ON TABLE job TYPE datetime DEFAULT time::now();
DEFINE FIELD completed_at ON TABLE job TYPE option<datetime>;
DEFINE FIELD error_msg ON TABLE job TYPE option<string>;
DEFINE INDEX idx_job_id ON TABLE job COLUMNS id UNIQUE;
DEFINE INDEX idx_job_scanner ON TABLE job COLUMNS scanner_key;
DEFINE INDEX idx_job_state ON TABLE job COLUMNS state;
DEFINE INDEX idx_job_created ON TABLE job COLUMNS created_at;

-- ============================================================================
-- FULL-TEXT SEARCH ANALYZERS
-- ============================================================================

-- Analyzer for vulnerability text search
DEFINE ANALYZER vuln_analyzer TOKENIZERS blank,class FILTERS lowercase,snowball(english);

-- ============================================================================
-- SCHEMA VALIDATION RULES
-- ============================================================================

-- Ensure CVSS scores are valid (0.0-10.0)
-- Note: This is enforced at application level; SurrealDB doesn't support CHECK constraints yet

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
