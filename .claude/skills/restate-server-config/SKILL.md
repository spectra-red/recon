---
name: restate-server-config
description: Guide for Restate server configuration including roles, cluster settings, storage, networking, and observability. Use when configuring Restate server, tuning performance, or setting up production deployments.
---

# Restate Server Configuration

Configure Restate server for development and production deployments.

## Configuration Methods

Settings can be provided via:
1. **TOML configuration file**: `--config-path=<PATH>`
2. **Environment variables**: Override file settings
3. **Command-line arguments**: Override both file and env vars

**Priority**: CLI args > Environment variables > Config file

## Key Configuration Sections

### Roles

Define server functionality. Available roles:
- `worker`: Process invocations
- `admin`: Admin API endpoint
- `metadata-server`: Raft consensus node
- `log-server`: Persist log segments
- `http-ingress`: Handle client requests

**Default**: All roles enabled (single-node mode)

#### Configuration

```toml
[roles]
worker = true
admin = true
metadata-server = true
log-server = true
http-ingress = true
```

#### Environment Variables

```bash
RESTATE_ROLES__WORKER=true
RESTATE_ROLES__ADMIN=true
RESTATE_ROLES__METADATA_SERVER=true
RESTATE_ROLES__LOG_SERVER=true
RESTATE_ROLES__HTTP_INGRESS=true
```

### Cluster Settings

Configure cluster identity and behavior.

#### Configuration

```toml
[cluster]
cluster-name = "my-restate-cluster"
auto-provision = true
advertised-address = "http://127.0.0.1:5122/"
```

#### Key Settings

**cluster-name** (default: "localcluster")
- Identifies the cluster
- Must match across all nodes

**auto-provision** (default: true)
- Automatically provision cluster on startup
- Disable for manual cluster setup

**advertised-address** (default: "http://127.0.0.1:5122/")
- Address advertised to other nodes
- Critical for multi-node clusters

#### Environment Variables

```bash
RESTATE_CLUSTER__CLUSTER_NAME=my-restate-cluster
RESTATE_CLUSTER__AUTO_PROVISION=true
RESTATE_CLUSTER__ADVERTISED_ADDRESS=http://10.0.1.5:5122/
```

### Partition Configuration

Control partition count and replication.

#### Configuration

```toml
[partitions]
default-num-partitions = 24
default-replication = 1
```

#### Key Settings

**default-num-partitions** (default: 24)
- Number of partitions
- Maximum: 65,535
- Cannot change after cluster creation

**default-replication** (default: 1)
- Replication factor for log
- Only applies during initial provisioning
- Higher values = more fault tolerance

#### Environment Variables

```bash
RESTATE_PARTITIONS__DEFAULT_NUM_PARTITIONS=24
RESTATE_PARTITIONS__DEFAULT_REPLICATION=2
```

## Storage & Performance

### RocksDB Memory Management

Control memory allocation for RocksDB (partition processor cache).

#### Configuration

```toml
[rocksdb]
rocksdb-total-memory-size = "6.0 GiB"
rocksdb-total-memtables-ratio = 0.5
rocksdb-high-priority-bg-threads = 2
```

#### Key Settings

**rocksdb-total-memory-size** (default: 6.0 GiB)
- Total memory budget for RocksDB
- Shared across all partitions on worker

**rocksdb-total-memtables-ratio** (default: 0.5)
- Ratio of memory for memtables (0.0-1.0)
- Remainder used for block cache

**rocksdb-high-priority-bg-threads** (default: 2)
- Background threads for compaction

#### Environment Variables

```bash
RESTATE_ROCKSDB__ROCKSDB_TOTAL_MEMORY_SIZE="8.0 GiB"
RESTATE_ROCKSDB__ROCKSDB_TOTAL_MEMTABLES_RATIO=0.5
RESTATE_ROCKSDB__ROCKSDB_HIGH_PRIORITY_BG_THREADS=4
```

### Worker Configuration

Configure invocation processing.

#### Configuration

```toml
[worker]
internal-queue-length = 1000
invoker-concurrency = 1000
message-size-warning = "10 MiB"
```

#### Key Settings

**internal-queue-length** (default: 1000)
- Queue depth for internal operations

**invoker-concurrency** (default: 1000)
- Maximum concurrent invocations per worker

**message-size-warning** (default: 10 MiB)
- Log warning for messages exceeding size

#### Environment Variables

```bash
RESTATE_WORKER__INTERNAL_QUEUE_LENGTH=2000
RESTATE_WORKER__INVOKER_CONCURRENCY=2000
RESTATE_WORKER__MESSAGE_SIZE_WARNING="20 MiB"
```

## Network & Ingress

### HTTP Configuration

Configure HTTP endpoints and behavior.

#### Admin API

```toml
[admin]
bind-address = "0.0.0.0:9070"
```

**Default**: `0.0.0.0:9070`

```bash
RESTATE_ADMIN__BIND_ADDRESS=0.0.0.0:9070
```

#### Ingress

```toml
[ingress]
bind-address = "0.0.0.0:8080"
request-compression-threshold = "4 MiB"
http2-keep-alive-interval = "40s"
```

**Key Settings**:
- **bind-address** (default: `0.0.0.0:8080`): Ingress endpoint
- **request-compression-threshold** (default: 4 MiB): Compress responses above size
- **http2-keep-alive-interval** (default: 40s): HTTP/2 keepalive

```bash
RESTATE_INGRESS__BIND_ADDRESS=0.0.0.0:8080
RESTATE_INGRESS__REQUEST_COMPRESSION_THRESHOLD="8 MiB"
RESTATE_INGRESS__HTTP2_KEEP_ALIVE_INTERVAL=60s
```

### Metadata Client

Configure metadata cluster connection.

#### Configuration

```toml
[metadata-client]
addresses = ["http://node1:5122", "http://node2:5122", "http://node3:5122"]
```

**Supports**:
- `replicated`: Default, replicated mode
- `etcd`: External etcd cluster
- `object-store`: S3-compatible storage

#### Environment Variables

```bash
RESTATE_METADATA_CLIENT__ADDRESSES="http://node1:5122,http://node2:5122,http://node3:5122"
```

## Logging & Observability

### Logging Configuration

Control log output format and level.

#### Configuration

```toml
[log]
log-filter = "warn,restate=info"
log-format = "pretty"
```

#### Key Settings

**log-filter** (default: "warn,restate=info")
- Log level filter using tracing directives
- Format: `target=level,target2=level2`
- Levels: `error`, `warn`, `info`, `debug`, `trace`

**log-format** (default: "pretty")
- Options: `pretty`, `compact`, `json`
- `json` recommended for production

#### Examples

```toml
# Verbose debugging
log-filter = "debug,restate=trace"

# Production: structured JSON
log-format = "json"
log-filter = "warn,restate=info"

# Specific component debugging
log-filter = "warn,restate::worker=debug,restate::bifrost=trace"
```

#### Environment Variables

```bash
RESTATE_LOG__LOG_FILTER="warn,restate=info"
RESTATE_LOG__LOG_FORMAT=json
```

### Distributed Tracing

Configure OpenTelemetry tracing.

#### Configuration

```toml
[tracing]
endpoint = "http://localhost:4317"
json-export = "/var/log/restate/traces.json"
```

**Key Settings**:
- **endpoint**: OTLP gRPC endpoint
- **json-export**: Optional Jaeger JSON export path

#### Environment Variables

```bash
RESTATE_TRACING__ENDPOINT=http://jaeger:4317
RESTATE_TRACING__JSON_EXPORT=/var/log/restate/traces.json
```

## Complete Configuration Examples

### Development (Single Node)

```toml
[cluster]
cluster-name = "dev-cluster"
auto-provision = true

[log]
log-format = "pretty"
log-filter = "info,restate=debug"

[admin]
bind-address = "127.0.0.1:9070"

[ingress]
bind-address = "127.0.0.1:8080"

[rocksdb]
rocksdb-total-memory-size = "2.0 GiB"
```

### Production (Multi-Node Worker)

```toml
[roles]
worker = true
admin = true
http-ingress = true
metadata-server = false
log-server = false

[cluster]
cluster-name = "prod-restate"
auto-provision = false
advertised-address = "http://10.0.1.5:5122/"

[metadata-client]
addresses = [
  "http://metadata-1:5122",
  "http://metadata-2:5122",
  "http://metadata-3:5122"
]

[log]
log-format = "json"
log-filter = "warn,restate=info"

[admin]
bind-address = "0.0.0.0:9070"

[ingress]
bind-address = "0.0.0.0:8080"
request-compression-threshold = "4 MiB"

[rocksdb]
rocksdb-total-memory-size = "16.0 GiB"
rocksdb-total-memtables-ratio = 0.5
rocksdb-high-priority-bg-threads = 4

[worker]
invoker-concurrency = 5000

[tracing]
endpoint = "http://jaeger:4317"
```

### Production (Metadata Server)

```toml
[roles]
worker = false
admin = true
http-ingress = false
metadata-server = true
log-server = true

[cluster]
cluster-name = "prod-restate"
auto-provision = false
advertised-address = "http://10.0.2.1:5122/"

[metadata-client]
addresses = [
  "http://metadata-1:5122",
  "http://metadata-2:5122",
  "http://metadata-3:5122"
]

[log]
log-format = "json"
log-filter = "warn,restate=info"

[admin]
bind-address = "0.0.0.0:9070"
```

## Environment Variable Examples

### Docker

```bash
docker run --name restate \
  -e RESTATE_CLUSTER__CLUSTER_NAME=my-cluster \
  -e RESTATE_LOG__LOG_FORMAT=json \
  -e RESTATE_ROCKSDB__ROCKSDB_TOTAL_MEMORY_SIZE="8.0 GiB" \
  -e RESTATE_WORKER__INVOKER_CONCURRENCY=2000 \
  -p 8080:8080 \
  -p 9070:9070 \
  -v ./restate-data:/restate-data \
  docker.io/restatedev/restate:latest
```

### Docker Compose

```yaml
services:
  restate:
    image: docker.io/restatedev/restate:latest
    environment:
      RESTATE_CLUSTER__CLUSTER_NAME: production-cluster
      RESTATE_LOG__LOG_FORMAT: json
      RESTATE_LOG__LOG_FILTER: "warn,restate=info"
      RESTATE_ROCKSDB__ROCKSDB_TOTAL_MEMORY_SIZE: "16.0 GiB"
      RESTATE_WORKER__INVOKER_CONCURRENCY: 5000
      RESTATE_INGRESS__REQUEST_COMPRESSION_THRESHOLD: "8 MiB"
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: restate-config
data:
  RESTATE_CLUSTER__CLUSTER_NAME: "k8s-cluster"
  RESTATE_LOG__LOG_FORMAT: "json"
  RESTATE_LOG__LOG_FILTER: "warn,restate=info"
  RESTATE_ROCKSDB__ROCKSDB_TOTAL_MEMORY_SIZE: "16.0 GiB"
```

## Performance Tuning

### Memory-Constrained Environments

```toml
[rocksdb]
rocksdb-total-memory-size = "1.0 GiB"
rocksdb-total-memtables-ratio = 0.3

[worker]
invoker-concurrency = 500
```

### High-Throughput Workloads

```toml
[rocksdb]
rocksdb-total-memory-size = "32.0 GiB"
rocksdb-total-memtables-ratio = 0.5
rocksdb-high-priority-bg-threads = 8

[worker]
invoker-concurrency = 10000
internal-queue-length = 5000

[ingress]
request-compression-threshold = "8 MiB"
```

### Large Message Workloads

```toml
[worker]
message-size-warning = "50 MiB"

[ingress]
request-compression-threshold = "1 MiB"
```

## Best Practices

1. **Use Config Files**: For complex configurations
2. **Environment Variables**: For deployment-specific overrides
3. **JSON Logging**: Always use in production
4. **Memory Allocation**: Size RocksDB based on state size
5. **Concurrency Limits**: Set based on workload and resources
6. **Monitoring**: Enable tracing in production
7. **Cluster Names**: Use descriptive, environment-specific names
8. **Documentation**: Document custom configurations

## Configuration Validation

Start server with `--validate-config` to check configuration:

```bash
restate-server --config-path config.toml --validate-config
```

## Common Configuration Patterns

### Local Development

```bash
# Minimal configuration
restate-server \
  --cluster-name dev \
  --log-format pretty \
  --log-filter debug
```

### Docker Production

```yaml
environment:
  RESTATE_CLUSTER__CLUSTER_NAME: ${CLUSTER_NAME}
  RESTATE_LOG__LOG_FORMAT: json
  RESTATE_ROCKSDB__ROCKSDB_TOTAL_MEMORY_SIZE: ${MEMORY_SIZE}
  RESTATE_METADATA_CLIENT__ADDRESSES: ${METADATA_SERVERS}
```

### Kubernetes Production

Use ConfigMaps for configuration, Secrets for credentials, and StatefulSets for workers.

## Troubleshooting

### High Memory Usage

**Solution**: Reduce RocksDB memory allocation
```toml
rocksdb-total-memory-size = "4.0 GiB"
```

### Low Throughput

**Solution**: Increase concurrency
```toml
invoker-concurrency = 5000
internal-queue-length = 2000
```

### Cluster Connection Issues

**Solution**: Verify metadata addresses
```bash
RESTATE_METADATA_CLIENT__ADDRESSES="http://node1:5122,http://node2:5122"
```

## References

- Official Docs: https://docs.restate.dev/references/server-config
- Docker Deployment: See restate-docker-deploy skill
- Architecture: See restate-architecture skill
- All configuration options in official documentation
