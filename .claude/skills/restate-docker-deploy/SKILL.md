---
name: restate-docker-deploy
description: Guide for deploying Restate server with Docker including single-node and multi-node clusters. Use when deploying Restate infrastructure, configuring Docker deployments, or setting up production clusters.
---

# Restate Docker Deployment

Deploy Restate server using Docker for local development, testing, and production.

## Recommended Use Cases

**Docker is suggested for**:
- Local development
- Testing environments
- Single-node production (with durable storage)

**Production deployments should consider**:
- Kubernetes with StatefulSets
- Persistent volumes (EBS, etc.)
- Multi-node clusters for fault tolerance

## Single-Node Server

Basic Docker deployment for development and single-node production.

### Basic Command

```bash
docker run --name restate \
  --rm \
  -p 8080:8080 \
  -p 9070:9070 \
  -p 5122:5122 \
  docker.io/restatedev/restate:latest
```

### Port Explanation

- **8080**: Ingress (client requests)
- **9070**: Admin interface (management API)
- **5122**: Node-to-node communication

### With Data Persistence

**Critical for production**: Mount a host volume for data durability.

```bash
docker run --name restate \
  --rm \
  -p 8080:8080 \
  -p 9070:9070 \
  -p 5122:5122 \
  -v $(pwd)/restate-data:/restate-data \
  docker.io/restatedev/restate:latest \
  --node-name restate-node-1
```

**Important**:
- Volume mount at `/restate-data` ensures data persistence
- Consistent `--node-name` required for proper data restoration
- Node name mismatch after restart causes data loss

### With Configuration File

```bash
docker run --name restate \
  --rm \
  -p 8080:8080 \
  -p 9070:9070 \
  -p 5122:5122 \
  -v $(pwd)/restate-data:/restate-data \
  -v $(pwd)/restate.toml:/restate.toml \
  docker.io/restatedev/restate:latest \
  --config-path /restate.toml \
  --node-name restate-node-1
```

### Docker Compose (Single Node)

```yaml
version: '3.8'

services:
  restate:
    image: docker.io/restatedev/restate:latest
    container_name: restate
    ports:
      - "8080:8080"   # Ingress
      - "9070:9070"   # Admin
      - "5122:5122"   # Node-to-node
    volumes:
      - ./restate-data:/restate-data
      - ./restate.toml:/restate.toml
    command:
      - "--config-path=/restate.toml"
      - "--node-name=restate-node-1"
    restart: unless-stopped
```

Run with:
```bash
docker compose up -d
```

## Multi-Node Cluster

3-node cluster setup for fault tolerance (survives one node failure).

### Docker Compose Configuration

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - ./minio-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 5s
      timeout: 3s
      retries: 3

  restate-1:
    image: docker.io/restatedev/restate:latest
    ports:
      - "8080:8080"
      - "9070:9070"
      - "5122:5122"
    environment:
      RESTATE_CLUSTER_NAME: my-restate-cluster
      RESTATE_DEFAULT_REPLICATION: 2
      RESTATE_METADATA_CLIENT__ADDRESSES: "http://restate-1:5122,http://restate-2:25122,http://restate-3:35122"
      RESTATE_LOG_BIFROST__PROVIDER: replicated
      RESTATE_LOG_BIFROST__REPLICATED__LOGLET_PROVIDER: local_loglet
      AWS_ACCESS_KEY_ID: minioadmin
      AWS_SECRET_ACCESS_KEY: minioadmin
      AWS_ENDPOINT: http://minio:9000
      AWS_REGION: us-east-1
    volumes:
      - ./restate-data-1:/restate-data
    command:
      - "--node-name=restate-1"
    depends_on:
      minio:
        condition: service_healthy

  restate-2:
    image: docker.io/restatedev/restate:latest
    ports:
      - "28080:8080"
      - "29070:9070"
      - "25122:5122"
    environment:
      RESTATE_CLUSTER_NAME: my-restate-cluster
      RESTATE_DEFAULT_REPLICATION: 2
      RESTATE_METADATA_CLIENT__ADDRESSES: "http://restate-1:5122,http://restate-2:25122,http://restate-3:35122"
      RESTATE_LOG_BIFROST__PROVIDER: replicated
      RESTATE_LOG_BIFROST__REPLICATED__LOGLET_PROVIDER: local_loglet
      AWS_ACCESS_KEY_ID: minioadmin
      AWS_SECRET_ACCESS_KEY: minioadmin
      AWS_ENDPOINT: http://minio:9000
      AWS_REGION: us-east-1
    volumes:
      - ./restate-data-2:/restate-data
    command:
      - "--node-name=restate-2"
    depends_on:
      minio:
        condition: service_healthy

  restate-3:
    image: docker.io/restatedev/restate:latest
    ports:
      - "38080:8080"
      - "39070:9070"
      - "35122:5122"
    environment:
      RESTATE_CLUSTER_NAME: my-restate-cluster
      RESTATE_DEFAULT_REPLICATION: 2
      RESTATE_METADATA_CLIENT__ADDRESSES: "http://restate-1:5122,http://restate-2:25122,http://restate-3:35122"
      RESTATE_LOG_BIFROST__PROVIDER: replicated
      RESTATE_LOG_BIFROST__REPLICATED__LOGLET_PROVIDER: local_loglet
      AWS_ACCESS_KEY_ID: minioadmin
      AWS_SECRET_ACCESS_KEY: minioadmin
      AWS_ENDPOINT: http://minio:9000
      AWS_REGION: us-east-1
    volumes:
      - ./restate-data-3:/restate-data
    command:
      - "--node-name=restate-3"
    depends_on:
      minio:
        condition: service_healthy
```

### Multi-Node Cluster Components

**MinIO Server**: Stores snapshots for fast recovery
- Port 9000: API
- Port 9001: Console
- Strongly recommended for all clusters

**Replicated Bifrost Provider**: Handles log writes across nodes
- Writes accepted when quorum acknowledges
- Minimum nodes required: `RESTATE_DEFAULT_REPLICATION`

**Partition State Replication**: Replicates to all workers
- One leader per partition
- Followers enable rapid failover

**Metadata Cluster**: Quorum-based consensus
- Manages cluster configuration
- Handles partition placement

### Port Mapping Strategy

Each node needs distinct port mappings to avoid conflicts:

**Node 1**:
- 8080:8080 (Ingress)
- 9070:9070 (Admin)
- 5122:5122 (Node-to-node)

**Node 2**:
- 28080:8080
- 29070:9070
- 25122:5122

**Node 3**:
- 38080:8080
- 39070:9070
- 35122:5122

### Environment Variables

**RESTATE_CLUSTER_NAME**: Cluster identifier
```yaml
RESTATE_CLUSTER_NAME: my-restate-cluster
```

**RESTATE_DEFAULT_REPLICATION**: Minimum nodes for write acceptance
```yaml
RESTATE_DEFAULT_REPLICATION: 2
```

**RESTATE_METADATA_CLIENT__ADDRESSES**: Internal communication paths
```yaml
RESTATE_METADATA_CLIENT__ADDRESSES: "http://restate-1:5122,http://restate-2:25122,http://restate-3:35122"
```

**Bifrost Configuration**: Log provider settings
```yaml
RESTATE_LOG_BIFROST__PROVIDER: replicated
RESTATE_LOG_BIFROST__REPLICATED__LOGLET_PROVIDER: local_loglet
```

**AWS/MinIO Credentials**: For snapshot storage
```yaml
AWS_ACCESS_KEY_ID: minioadmin
AWS_SECRET_ACCESS_KEY: minioadmin
AWS_ENDPOINT: http://minio:9000
AWS_REGION: us-east-1
```

## Production Deployment Patterns

### With Load Balancer

```yaml
version: '3.8'

services:
  nginx:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - restate-1
      - restate-2
      - restate-3

  restate-1:
    image: docker.io/restatedev/restate:latest
    # ... configuration ...

  restate-2:
    image: docker.io/restatedev/restate:latest
    # ... configuration ...

  restate-3:
    image: docker.io/restatedev/restate:latest
    # ... configuration ...
```

**nginx.conf**:
```nginx
upstream restate_ingress {
    server restate-1:8080;
    server restate-2:8080;
    server restate-3:8080;
}

server {
    listen 80;

    location / {
        proxy_pass http://restate_ingress;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Health Checks

```yaml
services:
  restate:
    image: docker.io/restatedev/restate:latest
    # ... other configuration ...
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9070/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### Resource Limits

```yaml
services:
  restate:
    image: docker.io/restatedev/restate:latest
    # ... other configuration ...
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
```

### Logging Configuration

```yaml
services:
  restate:
    image: docker.io/restatedev/restate:latest
    # ... other configuration ...
    environment:
      RESTATE_LOG_FORMAT: json
      RESTATE_LOG_FILTER: "warn,restate=info"
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Network Configuration

### Custom Network

```yaml
version: '3.8'

networks:
  restate-network:
    driver: bridge

services:
  restate:
    image: docker.io/restatedev/restate:latest
    networks:
      - restate-network
    # ... other configuration ...
```

### Host Network Mode

```yaml
services:
  restate:
    image: docker.io/restatedev/restate:latest
    network_mode: host
    # Ports are automatically exposed
```

## Service Registration

After starting Restate, register your services:

```bash
# Register service running on host
docker exec restate \
  restate deployments register http://host.docker.internal:9080

# Register service in same Docker network
docker exec restate \
  restate deployments register http://my-service:9080
```

## Common Operations

### View Logs

```bash
# Single container
docker logs restate

# Follow logs
docker logs -f restate

# With Docker Compose
docker compose logs restate-1
docker compose logs -f
```

### Access CLI

```bash
# Execute CLI commands
docker exec restate restate deployments list
docker exec restate restate invocations list

# Interactive shell
docker exec -it restate /bin/bash
```

### Backup Data

```bash
# Stop container
docker stop restate

# Backup data directory
tar -czf restate-backup.tar.gz ./restate-data

# Restart container
docker start restate
```

### Restore Data

```bash
# Stop container
docker stop restate

# Restore data
tar -xzf restate-backup.tar.gz

# Restart with same node name
docker start restate
```

## Troubleshooting

### Data Loss on Restart

**Problem**: Data lost after container restart

**Solution**: Ensure volume mount and consistent node name
```bash
-v $(pwd)/restate-data:/restate-data
--node-name restate-node-1  # Must be same on restart
```

### Port Conflicts

**Problem**: Port already in use

**Solution**: Change host port mapping
```bash
-p 9080:8080  # Use 9080 on host instead of 8080
```

### Node Communication Failures

**Problem**: Nodes can't communicate in cluster

**Solution**: Verify network configuration and internal addresses
```yaml
RESTATE_METADATA_CLIENT__ADDRESSES: "http://restate-1:5122,..."
```

### MinIO Connection Issues

**Problem**: Cannot connect to MinIO

**Solution**: Check MinIO health and credentials
```bash
docker logs minio
curl http://localhost:9000/minio/health/live
```

## Best Practices

1. **Always Use Volumes**: Mount `/restate-data` for persistence
2. **Consistent Node Names**: Use same `--node-name` across restarts
3. **Health Checks**: Implement health checks for production
4. **Resource Limits**: Set appropriate CPU and memory limits
5. **Logging**: Configure structured logging for production
6. **MinIO for Clusters**: Always use object storage for multi-node setups
7. **Monitoring**: Set up monitoring and alerting
8. **Backups**: Regular data backups for disaster recovery

## Kubernetes Alternative

For production, consider Kubernetes with StatefulSets:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: restate
spec:
  serviceName: restate
  replicas: 3
  selector:
    matchLabels:
      app: restate
  template:
    metadata:
      labels:
        app: restate
    spec:
      containers:
      - name: restate
        image: docker.io/restatedev/restate:latest
        ports:
        - containerPort: 8080
        - containerPort: 9070
        - containerPort: 5122
        volumeMounts:
        - name: data
          mountPath: /restate-data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 100Gi
```

## References

- Official Docs: https://docs.restate.dev/server/deploy/docker
- Server Configuration: See restate-server-config skill
- Restate Architecture: See restate-architecture skill
- Docker Hub: https://hub.docker.com/r/restatedev/restate
