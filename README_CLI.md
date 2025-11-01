# Spectra CLI Documentation

The Spectra command-line interface (CLI) provides a unified tool for interacting with the Spectra-Red Intel Mesh.

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/spectra-red/recon
cd recon

# Build the CLI
go build -o spectra ./cmd/spectra

# Optional: Install to $GOPATH/bin
go install ./cmd/spectra

# Build with version information
go build -ldflags="-X main.version=1.0.0 -X main.gitCommit=$(git rev-parse HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o spectra ./cmd/spectra
```

### Using Pre-built Binaries

Download the latest release from the [releases page](https://github.com/spectra-red/recon/releases).

## Quick Start

```bash
# View help
spectra --help

# Check version
spectra version

# View configuration
spectra version -v

# Query a host
spectra query host 1.2.3.4

# Submit scan results
naabu -host example.com -json | spectra ingest
```

## Configuration

### Configuration File

The CLI looks for configuration files in the following locations (in order of precedence):

1. `./.spectra.yaml` (current directory)
2. `~/.spectra/.spectra.yaml` (home directory)
3. `/etc/spectra/config.yaml` (system-wide)

Example configuration file:

```yaml
api:
  url: http://localhost:3000
  timeout: 30s

scanner:
  public_key: <your-ed25519-public-key>
  private_key: <your-ed25519-private-key>

output:
  format: json
  color: true
```

### Environment Variables

All configuration options can be set via environment variables with the `SPECTRA_` prefix:

```bash
export SPECTRA_API_URL=http://api.example.com:3000
export SPECTRA_OUTPUT_FORMAT=yaml
export SPECTRA_OUTPUT_COLOR=false
export SPECTRA_SCANNER_PUBLIC_KEY=<your-public-key>
export SPECTRA_SCANNER_PRIVATE_KEY=<your-private-key>
```

### Configuration Precedence

Configuration is loaded in the following order (later sources override earlier ones):

1. Default values
2. Configuration file
3. Environment variables
4. Command-line flags

## Commands

### Global Flags

These flags are available for all commands:

- `--config <file>` - Specify a custom config file
- `--api-url <url>` - Override the API endpoint URL
- `--verbose, -v` - Enable verbose output

### `spectra version`

Display version information.

```bash
# Default text output
spectra version

# JSON output
spectra version --output json

# YAML output
spectra version --output yaml
```

Example output:
```
Spectra CLI
  Version:    1.0.0
  Git Commit: abc123def
  Build Date: 2025-11-01T12:00:00Z
  Go Version: go1.25.3
```

### `spectra ingest`

Submit scan results to the mesh.

**Note:** Full implementation coming in M4-T2.

```bash
# Ingest from stdin (Naabu JSON output)
naabu -host example.com -json | spectra ingest

# Ingest from file
spectra ingest --file scan-results.json

# Dry run (validate without submitting)
spectra ingest --file scan-results.json --dry-run

# Specify input format
spectra ingest --file scan.xml --format nmap
```

**Flags:**
- `--file, -f <path>` - Input file containing scan results
- `--dry-run` - Validate input without submitting to the mesh
- `--format <format>` - Input format (naabu, nmap) (default: naabu)

### `spectra query`

Query threat intelligence from the mesh.

#### `spectra query host <ip>`

Query detailed information about a specific host.

**Note:** Implementation coming in future milestone.

```bash
# Query host information
spectra query host 1.2.3.4

# JSON output
spectra query host 1.2.3.4 --output json
```

Returns:
- Open ports and services
- Geographic location
- ASN information
- Known vulnerabilities
- Last seen timestamp

#### `spectra query graph`

Query using graph-based selectors.

**Note:** Implementation coming in future milestone.

```bash
# Find all Redis servers in Paris
spectra query graph --city Paris --service redis

# Find all hosts in a specific ASN
spectra query graph --asn 15169

# Complex query
spectra query graph --city "New York" --port 22 --service openssh
```

**Flags:**
- `--city <name>` - Filter by city
- `--country <name>` - Filter by country
- `--asn <number>` - Filter by ASN
- `--port <number>` - Filter by port number
- `--service <name>` - Filter by service name
- `--limit <number>` - Maximum number of results (default: 100)

#### `spectra query similar <ip>`

Find similar hosts using vector similarity search.

**Note:** Implementation coming in M8 (AI/Vector Search milestone).

```bash
# Find similar hosts
spectra query similar 1.2.3.4

# Adjust similarity threshold
spectra query similar 1.2.3.4 --threshold 0.9 --limit 20
```

**Flags:**
- `--limit <number>` - Number of similar hosts to return (default: 10)
- `--threshold <float>` - Similarity threshold 0.0-1.0 (default: 0.8)

### `spectra jobs`

Manage background ingest jobs.

**Note:** Implementation coming in future milestone.

#### `spectra jobs list`

List recent ingest jobs.

```bash
# List all recent jobs
spectra jobs list

# Filter by status
spectra jobs list --status failed

# Custom limit
spectra jobs list --limit 50

# Jobs from last 24 hours
spectra jobs list --since 24h
```

**Flags:**
- `--status <status>` - Filter by status (pending, processing, completed, failed)
- `--limit <number>` - Maximum number of jobs (default: 20)
- `--since <duration>` - Show jobs since duration (e.g., 24h, 7d)

#### `spectra jobs get <job-id>`

Get details of a specific job.

```bash
# Get job details
spectra jobs get job_abc123

# JSON output
spectra jobs get job_abc123 --output json
```

Returns:
- Job status
- Submission timestamp
- Processing time
- Error messages (if failed)
- Result summary

## Output Formats

All query commands support multiple output formats:

### JSON (default)

```bash
spectra query host 1.2.3.4 --output json
```

```json
{
  "ip": "1.2.3.4",
  "asn": 15169,
  "city": "Paris",
  "ports": [...]
}
```

### YAML

```bash
spectra query host 1.2.3.4 --output yaml
```

```yaml
ip: 1.2.3.4
asn: 15169
city: Paris
ports: [...]
```

### Table

```bash
spectra query graph --city Paris --output table
```

```
IP          ASN    CITY   PORTS
1.2.3.4    15169  Paris  80,443
5.6.7.8    15169  Paris  22,80,443
```

## Authentication

The CLI uses Ed25519 key pairs for authentication when submitting scan results.

### Generate Keys

**Note:** Key management commands coming in M4-T3.

```bash
# Generate new key pair
spectra keys generate

# Rotate existing keys
spectra keys rotate
```

Keys are stored in `~/.spectra/keys/`:
- `contributor.key` - Private key (permissions: 0600)
- `contributor.pub` - Public key (permissions: 0644)

### Security Best Practices

1. Never share your private key
2. Never commit keys to version control
3. Use environment variables in CI/CD pipelines
4. Rotate keys periodically
5. Store keys securely (e.g., password manager, key vault)

## Examples

### Scanning and Submitting Results

```bash
# Scan a single host and submit
naabu -host example.com -json | spectra ingest

# Scan multiple hosts
cat targets.txt | naabu -json | spectra ingest

# Scan with custom config
spectra --config ~/.spectra-prod.yaml ingest --file results.json
```

### Querying the Mesh

```bash
# Find all MongoDB servers in San Francisco
spectra query graph --city "San Francisco" --service mongodb

# Get details about a specific host
spectra query host 203.0.113.42

# Find similar vulnerable hosts
spectra query similar 203.0.113.42
```

### Job Management

```bash
# Check status of recent jobs
spectra jobs list

# Get details of a failed job
spectra jobs get job_abc123
```

## Troubleshooting

### Configuration Issues

```bash
# Check which config file is being used
spectra version -v

# Test with explicit config file
spectra --config /path/to/.spectra.yaml version
```

### API Connection Issues

```bash
# Test connectivity with verbose output
spectra --verbose query host 1.2.3.4

# Override API URL
spectra --api-url http://api.example.com:3000 version
```

### Authentication Issues

```bash
# Verify keys are configured
spectra version -v

# Generate new keys if needed
spectra keys generate
```

## Development

### Building

```bash
# Development build
go build -o spectra ./cmd/spectra

# Production build with version info
make build
```

### Testing

```bash
# Run all tests
go test ./internal/cli/...

# Run with verbose output
go test -v ./internal/cli/...

# Run specific test
go test -v -run TestInitConfig ./internal/cli
```

### Adding New Commands

1. Create a new file in `internal/cli/` (e.g., `newcommand.go`)
2. Implement a `NewCommandCommand()` function returning `*cobra.Command`
3. Add the command to `root.go` in the `NewRootCommand()` function
4. Add tests in `newcommand_test.go`
5. Update this documentation

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

See [LICENSE](LICENSE) for details.

## Support

- GitHub Issues: https://github.com/spectra-red/recon/issues
- Documentation: https://docs.spectra-red.com
- Community Forum: https://community.spectra-red.com
