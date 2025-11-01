-- ============================================================================
-- Spectra-Red Intel Mesh - Seed Data
-- ============================================================================
-- This file contains seed data for the threat intelligence mesh including:
-- - Common ports (well-known services)
-- - Geographic data (countries, regions, ASNs)
-- - Cloud provider regions
-- - Sample vulnerability data
--
-- Total Records: 100+
-- ============================================================================

-- ============================================================================
-- COMMON PORTS (Well-Known Services)
-- ============================================================================

INSERT INTO common_port [
    { number: 20, label: 'ftp-data', description: 'FTP Data Transfer' },
    { number: 21, label: 'ftp', description: 'FTP Control' },
    { number: 22, label: 'ssh', description: 'Secure Shell (SSH)' },
    { number: 23, label: 'telnet', description: 'Telnet' },
    { number: 25, label: 'smtp', description: 'Simple Mail Transfer Protocol' },
    { number: 53, label: 'dns', description: 'Domain Name System' },
    { number: 80, label: 'http', description: 'Hypertext Transfer Protocol' },
    { number: 110, label: 'pop3', description: 'Post Office Protocol v3' },
    { number: 143, label: 'imap', description: 'Internet Message Access Protocol' },
    { number: 443, label: 'https', description: 'HTTP over TLS/SSL' },
    { number: 445, label: 'smb', description: 'Server Message Block' },
    { number: 587, label: 'smtp-tls', description: 'SMTP with STARTTLS' },
    { number: 993, label: 'imaps', description: 'IMAP over TLS/SSL' },
    { number: 995, label: 'pop3s', description: 'POP3 over TLS/SSL' },
    { number: 1433, label: 'mssql', description: 'Microsoft SQL Server' },
    { number: 1521, label: 'oracle', description: 'Oracle Database' },
    { number: 3306, label: 'mysql', description: 'MySQL Database' },
    { number: 3389, label: 'rdp', description: 'Remote Desktop Protocol' },
    { number: 5432, label: 'postgres', description: 'PostgreSQL Database' },
    { number: 5672, label: 'amqp', description: 'Advanced Message Queuing Protocol' },
    { number: 6379, label: 'redis', description: 'Redis Key-Value Store' },
    { number: 8080, label: 'http-alt', description: 'HTTP Alternate' },
    { number: 8443, label: 'https-alt', description: 'HTTPS Alternate' },
    { number: 9200, label: 'elasticsearch', description: 'Elasticsearch HTTP' },
    { number: 9300, label: 'elasticsearch-transport', description: 'Elasticsearch Transport' },
    { number: 11211, label: 'memcached', description: 'Memcached' },
    { number: 27017, label: 'mongodb', description: 'MongoDB Database' },
    { number: 27018, label: 'mongodb-shard', description: 'MongoDB Shard Server' }
];

-- ============================================================================
-- COUNTRIES (Top 20 + Major Tech Hubs)
-- ============================================================================

INSERT INTO country [
    { cc: 'US', name: 'United States' },
    { cc: 'GB', name: 'United Kingdom' },
    { cc: 'CA', name: 'Canada' },
    { cc: 'FR', name: 'France' },
    { cc: 'DE', name: 'Germany' },
    { cc: 'NL', name: 'Netherlands' },
    { cc: 'SE', name: 'Sweden' },
    { cc: 'FI', name: 'Finland' },
    { cc: 'NO', name: 'Norway' },
    { cc: 'DK', name: 'Denmark' },
    { cc: 'JP', name: 'Japan' },
    { cc: 'CN', name: 'China' },
    { cc: 'IN', name: 'India' },
    { cc: 'SG', name: 'Singapore' },
    { cc: 'HK', name: 'Hong Kong' },
    { cc: 'KR', name: 'South Korea' },
    { cc: 'AU', name: 'Australia' },
    { cc: 'NZ', name: 'New Zealand' },
    { cc: 'BR', name: 'Brazil' },
    { cc: 'MX', name: 'Mexico' },
    { cc: 'AR', name: 'Argentina' },
    { cc: 'ZA', name: 'South Africa' },
    { cc: 'RU', name: 'Russia' },
    { cc: 'UA', name: 'Ukraine' },
    { cc: 'PL', name: 'Poland' }
];

-- ============================================================================
-- ASNS (Top Cloud Providers + Major ISPs)
-- ============================================================================

INSERT INTO asn [
    -- AWS
    { number: 16509, org: 'Amazon.com, Inc.', country: 'US', type: 'cloud' },
    { number: 14618, org: 'Amazon.com, Inc.', country: 'US', type: 'cloud' },

    -- Google Cloud
    { number: 15169, org: 'Google LLC', country: 'US', type: 'cloud' },
    { number: 19527, org: 'Google LLC', country: 'US', type: 'cloud' },

    -- Microsoft Azure
    { number: 8075, org: 'Microsoft Corporation', country: 'US', type: 'cloud' },
    { number: 8068, org: 'Microsoft Corporation', country: 'US', type: 'cloud' },

    -- DigitalOcean
    { number: 14061, org: 'DigitalOcean, LLC', country: 'US', type: 'cloud' },

    -- Linode/Akamai
    { number: 63949, org: 'Linode, LLC', country: 'US', type: 'cloud' },

    -- Cloudflare
    { number: 13335, org: 'Cloudflare, Inc.', country: 'US', type: 'cloud' },

    -- OVH
    { number: 16276, org: 'OVH SAS', country: 'FR', type: 'hosting' },

    -- Hetzner
    { number: 24940, org: 'Hetzner Online GmbH', country: 'DE', type: 'hosting' },

    -- Alibaba Cloud
    { number: 45102, org: 'Alibaba (US) Technology Co., Ltd.', country: 'CN', type: 'cloud' },

    -- Tencent Cloud
    { number: 132203, org: 'Tencent Building, Kejizhongyi Avenue', country: 'CN', type: 'cloud' },

    -- Oracle Cloud
    { number: 31898, org: 'Oracle Corporation', country: 'US', type: 'cloud' },

    -- IBM Cloud
    { number: 36351, org: 'SoftLayer Technologies Inc.', country: 'US', type: 'cloud' },

    -- Major ISPs
    { number: 7922, org: 'Comcast Cable Communications, LLC', country: 'US', type: 'isp' },
    { number: 209, org: 'CenturyLink Communications, LLC', country: 'US', type: 'isp' },
    { number: 3356, org: 'Level 3 Parent, LLC', country: 'US', type: 'isp' },
    { number: 174, org: 'Cogent Communications', country: 'US', type: 'isp' },
    { number: 1299, org: 'Telia Company AB', country: 'SE', type: 'isp' }
];

-- ============================================================================
-- CLOUD REGIONS (AWS, GCP, Azure)
-- ============================================================================

INSERT INTO cloud_region [
    -- AWS Regions
    { provider: 'aws', code: 'us-east-1', name: 'US East (N. Virginia)', city: 'Ashburn', country: 'US' },
    { provider: 'aws', code: 'us-east-2', name: 'US East (Ohio)', city: 'Columbus', country: 'US' },
    { provider: 'aws', code: 'us-west-1', name: 'US West (N. California)', city: 'San Francisco', country: 'US' },
    { provider: 'aws', code: 'us-west-2', name: 'US West (Oregon)', city: 'Portland', country: 'US' },
    { provider: 'aws', code: 'eu-west-1', name: 'Europe (Ireland)', city: 'Dublin', country: 'IE' },
    { provider: 'aws', code: 'eu-west-2', name: 'Europe (London)', city: 'London', country: 'GB' },
    { provider: 'aws', code: 'eu-central-1', name: 'Europe (Frankfurt)', city: 'Frankfurt', country: 'DE' },
    { provider: 'aws', code: 'ap-southeast-1', name: 'Asia Pacific (Singapore)', city: 'Singapore', country: 'SG' },
    { provider: 'aws', code: 'ap-northeast-1', name: 'Asia Pacific (Tokyo)', city: 'Tokyo', country: 'JP' },

    -- GCP Regions
    { provider: 'gcp', code: 'us-central1', name: 'Iowa', city: 'Council Bluffs', country: 'US' },
    { provider: 'gcp', code: 'us-east1', name: 'South Carolina', city: 'Moncks Corner', country: 'US' },
    { provider: 'gcp', code: 'us-west1', name: 'Oregon', city: 'The Dalles', country: 'US' },
    { provider: 'gcp', code: 'europe-west1', name: 'Belgium', city: 'St. Ghislain', country: 'BE' },
    { provider: 'gcp', code: 'europe-west2', name: 'London', city: 'London', country: 'GB' },
    { provider: 'gcp', code: 'asia-southeast1', name: 'Singapore', city: 'Singapore', country: 'SG' },
    { provider: 'gcp', code: 'asia-northeast1', name: 'Tokyo', city: 'Tokyo', country: 'JP' },

    -- Azure Regions
    { provider: 'azure', code: 'eastus', name: 'East US', city: 'Virginia', country: 'US' },
    { provider: 'azure', code: 'eastus2', name: 'East US 2', city: 'Virginia', country: 'US' },
    { provider: 'azure', code: 'westus', name: 'West US', city: 'California', country: 'US' },
    { provider: 'azure', code: 'westus2', name: 'West US 2', city: 'Washington', country: 'US' },
    { provider: 'azure', code: 'northeurope', name: 'North Europe', city: 'Dublin', country: 'IE' },
    { provider: 'azure', code: 'westeurope', name: 'West Europe', city: 'Amsterdam', country: 'NL' },
    { provider: 'azure', code: 'uksouth', name: 'UK South', city: 'London', country: 'GB' },
    { provider: 'azure', code: 'southeastasia', name: 'Southeast Asia', city: 'Singapore', country: 'SG' },
    { provider: 'azure', code: 'japaneast', name: 'Japan East', city: 'Tokyo', country: 'JP' }
];

-- ============================================================================
-- SAMPLE REGIONS (US States + Major International)
-- ============================================================================

INSERT INTO region [
    -- US States
    { name: 'California', cc: 'US', code: 'CA' },
    { name: 'New York', cc: 'US', code: 'NY' },
    { name: 'Texas', cc: 'US', code: 'TX' },
    { name: 'Florida', cc: 'US', code: 'FL' },
    { name: 'Illinois', cc: 'US', code: 'IL' },
    { name: 'Virginia', cc: 'US', code: 'VA' },
    { name: 'Washington', cc: 'US', code: 'WA' },
    { name: 'Massachusetts', cc: 'US', code: 'MA' },

    -- International
    { name: 'England', cc: 'GB', code: 'ENG' },
    { name: 'Ontario', cc: 'CA', code: 'ON' },
    { name: 'Île-de-France', cc: 'FR', code: 'IDF' },
    { name: 'Bavaria', cc: 'DE', code: 'BY' },
    { name: 'Tokyo', cc: 'JP', code: '13' },
    { name: 'Singapore', cc: 'SG', code: 'SG' }
];

-- ============================================================================
-- SAMPLE CITIES (Major Tech Hubs)
-- ============================================================================

INSERT INTO city [
    -- US Cities
    { name: 'San Francisco', cc: 'US', lat: 37.7749, lon: -122.4194 },
    { name: 'New York', cc: 'US', lat: 40.7128, lon: -74.0060 },
    { name: 'Seattle', cc: 'US', lat: 47.6062, lon: -122.3321 },
    { name: 'Austin', cc: 'US', lat: 30.2672, lon: -97.7431 },
    { name: 'Boston', cc: 'US', lat: 42.3601, lon: -71.0589 },
    { name: 'Los Angeles', cc: 'US', lat: 34.0522, lon: -118.2437 },
    { name: 'Chicago', cc: 'US', lat: 41.8781, lon: -87.6298 },
    { name: 'Miami', cc: 'US', lat: 25.7617, lon: -80.1918 },
    { name: 'Denver', cc: 'US', lat: 39.7392, lon: -104.9903 },
    { name: 'Ashburn', cc: 'US', lat: 39.0438, lon: -77.4874 },

    -- International Cities
    { name: 'London', cc: 'GB', lat: 51.5074, lon: -0.1278 },
    { name: 'Paris', cc: 'FR', lat: 48.8566, lon: 2.3522 },
    { name: 'Frankfurt', cc: 'DE', lat: 50.1109, lon: 8.6821 },
    { name: 'Amsterdam', cc: 'NL', lat: 52.3676, lon: 4.9041 },
    { name: 'Stockholm', cc: 'SE', lat: 59.3293, lon: 18.0686 },
    { name: 'Tokyo', cc: 'JP', lat: 35.6762, lon: 139.6503 },
    { name: 'Singapore', cc: 'SG', lat: 1.3521, lon: 103.8198 },
    { name: 'Hong Kong', cc: 'HK', lat: 22.3193, lon: 114.1694 },
    { name: 'Sydney', cc: 'AU', lat: -33.8688, lon: 151.2093 },
    { name: 'Toronto', cc: 'CA', lat: 43.6532, lon: -79.3832 },
    { name: 'Dublin', cc: 'IE', lat: 53.3498, lon: -6.2603 },
    { name: 'Bangalore', cc: 'IN', lat: 12.9716, lon: 77.5946 },
    { name: 'Mumbai', cc: 'IN', lat: 19.0760, lon: 72.8777 },
    { name: 'São Paulo', cc: 'BR', lat: -23.5505, lon: -46.6333 },
    { name: 'Mexico City', cc: 'MX', lat: 19.4326, lon: -99.1332 }
];

-- ============================================================================
-- SAMPLE VULNERABILITIES (Recent High-Impact CVEs)
-- ============================================================================

INSERT INTO vuln [
    {
        cve_id: 'CVE-2024-3094',
        cvss: 10.0,
        severity: 'critical',
        kev_flag: true,
        first_seen: time::now(),
        last_updated: time::now()
    },
    {
        cve_id: 'CVE-2024-21762',
        cvss: 9.8,
        severity: 'critical',
        kev_flag: true,
        first_seen: time::now(),
        last_updated: time::now()
    },
    {
        cve_id: 'CVE-2023-46604',
        cvss: 9.8,
        severity: 'critical',
        kev_flag: true,
        first_seen: time::now(),
        last_updated: time::now()
    },
    {
        cve_id: 'CVE-2023-4966',
        cvss: 9.4,
        severity: 'critical',
        kev_flag: true,
        first_seen: time::now(),
        last_updated: time::now()
    },
    {
        cve_id: 'CVE-2023-0669',
        cvss: 8.8,
        severity: 'high',
        kev_flag: false,
        first_seen: time::now(),
        last_updated: time::now()
    }
];

-- ============================================================================
-- SAMPLE VULN_DOC ENTRIES (For RAG/Vector Search Testing)
-- ============================================================================
-- Note: Embeddings are placeholder zeros - real embeddings would come from OpenAI API

INSERT INTO vuln_doc [
    {
        cve_id: 'CVE-2024-3094',
        title: 'XZ Utils Backdoor',
        summary: 'Malicious code in xz/liblzma leading to SSH server compromise',
        cvss: 10.0,
        epss: 0.89,
        cpe: ['cpe:2.3:a:tukaani:xz:5.6.0:*:*:*:*:*:*:*', 'cpe:2.3:a:tukaani:xz:5.6.1:*:*:*:*:*:*:*'],
        exploit_refs: ['https://www.openwall.com/lists/oss-security/2024/03/29/4'],
        embedding: array::new(1536, 0.0),
        published_date: <datetime>'2024-03-29T00:00:00Z',
        last_modified: time::now()
    },
    {
        cve_id: 'CVE-2024-21762',
        title: 'Fortinet FortiOS Out-of-Bounds Write',
        summary: 'Out-of-bounds write in FortiOS administrative interface allows remote code execution',
        cvss: 9.8,
        epss: 0.75,
        cpe: ['cpe:2.3:o:fortinet:fortios:7.0.0:*:*:*:*:*:*:*', 'cpe:2.3:o:fortinet:fortios:7.2.0:*:*:*:*:*:*:*'],
        exploit_refs: ['https://www.fortiguard.com/psirt/FG-IR-24-029'],
        embedding: array::new(1536, 0.0),
        published_date: <datetime>'2024-02-08T00:00:00Z',
        last_modified: time::now()
    },
    {
        cve_id: 'CVE-2023-46604',
        title: 'Apache ActiveMQ Remote Code Execution',
        summary: 'Remote code execution via OpenWire protocol deserialization',
        cvss: 9.8,
        epss: 0.92,
        cpe: ['cpe:2.3:a:apache:activemq:5.18.0:*:*:*:*:*:*:*', 'cpe:2.3:a:apache:activemq:5.17.0:*:*:*:*:*:*:*'],
        exploit_refs: ['https://nvd.nist.gov/vuln/detail/CVE-2023-46604'],
        embedding: array::new(1536, 0.0),
        published_date: <datetime>'2023-10-27T00:00:00Z',
        last_modified: time::now()
    }
];

-- ============================================================================
-- SUMMARY
-- ============================================================================
-- Records Created:
-- - Common Ports: 28
-- - Countries: 25
-- - ASNs: 20
-- - Cloud Regions: 26
-- - Regions: 14
-- - Cities: 25
-- - Vulnerabilities: 5
-- - Vuln Docs: 3
-- Total: 146 records
-- ============================================================================
