---
name: restate-architecture
description: Guide to Restate architecture including durable log, partition processors, ingress layer, and distributed cluster design. Use when understanding Restate internals, designing systems, or troubleshooting distributed behavior.
---

# Restate Architecture

Understanding Restate's design principles, components, and distributed execution model.

## Core Design Philosophy

Restate is engineered as a **partitioned, log-centric runtime** that:
- Prioritizes simplicity for single-node deployments
- Scales to distributed clusters
- Treats the replicated log as ground truth
- Uses S3-backed snapshots for bounded recovery

**Key principle**: "The replicated log as ground truth"

## Primary Components

### Ingress Layer

Routes client and internal calls to appropriate partitions.

**Routing Strategy**: Hash-based
- Workflow IDs → Partition
- Virtual Object keys → Partition
- Idempotency keys → Partition

**Dynamic Updates**: Ingress automatically updates routing when leadership changes occur.

**Capabilities**:
- HTTP/2 endpoint for client requests
- Automatic partition discovery
- Load distribution across partitions
- Failover handling

### Durable Log (Bifrost)

Each partition maintains a single leader that:
- Orders all events
- Replicates across peer nodes
- Commits when quorum acknowledges

**Segmented Virtual Logs**:
- Enable clean and fast leadership changes
- Support placement updates
- Allow reconfiguration without data copying

**Write Path**:
1. Partition leader appends record to log
2. Log replicates to peer nodes
3. Write commits when quorum acknowledges
4. Commit point establishes durable ordering

**Guarantees**:
- Invocations ordered durably
- State updates persisted
- Timers recorded
- All operations survive failures

### Partition Processor

Maintains invocation state and materialized data.

**Storage**: Embedded RocksDB
- Invocation state
- Virtual Object state
- Workflow state
- Materialized views of log data

**Processing Model**:
- Tails the durable log
- Invokes handler code via bidirectional streams
- Optional followers for rapid failover

**Key Characteristic**: RocksDB cache is **derivative**
- Can always be rebuilt from log
- Not authoritative source of truth
- Accelerates processing, not durability

### Control Plane (Metadata Layer)

Built on Raft consensus for cluster coordination.

**Responsibilities**:
- Cluster configuration management
- Partition placement decisions
- Failover coordination
- Leadership elections

**Consensus Interface**:
- Leadership changes
- Rebalancing operations
- Configuration updates

## Durability Model

### Write Path Sequence

1. **Client Request** → Ingress
2. **Ingress** → Routes to partition leader
3. **Leader** → Appends to durable log
4. **Log** → Replicates to peers
5. **Quorum Ack** → Write commits
6. **Processor** → Updates RocksDB cache
7. **Response** → Returns to client

**Commit Point**: When quorum acknowledges log append

This establishes durable ordering for:
- Invocations
- State updates
- Timers
- Service calls

### Recovery Model

**Snapshot-Based Recovery**:
1. Partition processor periodically creates snapshots
2. Snapshots upload to S3 (or compatible storage)
3. On recovery: Download latest snapshot
4. Replay log suffix since snapshot sequence number

**Benefits**:
- Bounded recovery time
- No need to replay entire log
- Predictable startup performance

## Scalability Mechanisms

### Keyed Routing

Deterministic hashing assigns keys to partitions:
- Hot paths remain partition-local
- Each partition owns orchestration for its keys
- No cross-partition coordination needed for normal operations

**Advantages**:
- Scalability through partitioning
- Isolation of failure domains
- Predictable performance

### Exactly-Once Delivery

Cross-partition message delivery:

1. **Origin**: Message recorded in origin partition log
2. **Delivery**: Internal shuffler delivers to target partition
3. **Deduplication**: Sequence numbers prevent duplicates
4. **Epoch Fencing**: Prevents split-brain during failover

**Failover Safety**: Epoch fencing rejects superseded attempts

### Partition Distribution

Partitions distributed across worker nodes:
- Each partition has one leader
- Optional followers for fast failover
- Leaders process invocations
- Followers tail log, ready to promote

## Node Roles in Distributed Clusters

### Metadata Servers

- **Raft-based consensus** for cluster state
- Manage configuration
- Coordinate failovers
- Track partition placement

### HTTP Ingress Nodes

- Handle external client requests
- Stateless (no persistent data)
- Route to partition leaders
- Can scale independently

### Log Servers

- Persist replicated log segments
- Provide durability guarantees
- Support quorum-based replication
- Store until snapshots allow truncation

### Worker Nodes

- Host partition processors
- Run in leader or follower mode
- Execute handler code
- Maintain RocksDB caches

**Role Specialization**: Allows efficient resource allocation as clusters scale horizontally.

## Data Flow

### Invocation Lifecycle

```
Client Request
    ↓
Ingress (route by key hash)
    ↓
Partition Leader
    ↓
Append to Log (replicate to peers)
    ↓
Wait for Quorum Ack
    ↓
Commit Point Established
    ↓
Partition Processor Processes
    ↓
Invoke Handler via Bidirectional Stream
    ↓
Handler Executes (with durable steps)
    ↓
Results Append to Log
    ↓
Response to Client
```

### State Update Flow

```
Handler: restate.Set(ctx, "key", value)
    ↓
SDK: Record in execution journal
    ↓
Processor: Append to partition log
    ↓
Log: Replicate to quorum
    ↓
Commit: State update durable
    ↓
RocksDB: Update materialized cache
    ↓
Future Reads: Served from cache
```

### Cross-Partition Communication

```
Partition A: Call service in Partition B
    ↓
Append call intent to Partition A log
    ↓
Internal Shuffler: Deliver to Partition B
    ↓
Partition B: Append incoming call to log
    ↓
Execute handler in Partition B
    ↓
Response recorded in Partition B log
    ↓
Shuffler: Deliver response to Partition A
    ↓
Partition A: Complete awaiting invocation
```

## Fault Tolerance

### Single Node Failure

**Leader Failure**:
1. Metadata cluster detects failure
2. Promotes follower to leader (if available)
3. Or: Assigns partition to different worker
4. New leader recovers from snapshot + log
5. Ingress updates routing

**Worker Failure**:
1. All partitions on worker become unavailable
2. Metadata cluster reassigns partitions
3. New workers recover from snapshots
4. Processing resumes

### Network Partition

**Split Brain Prevention**:
- Epoch fencing rejects stale leaders
- Raft ensures single source of truth
- Quorum requirements prevent inconsistency

### Data Corruption

**Log Corruption**:
- Replicated log provides redundancy
- Quorum ensures at least N copies exist
- Can recover from peer copies

**RocksDB Corruption**:
- Delete corrupted cache
- Rebuild from snapshot + log replay
- No data loss (log is source of truth)

## Performance Characteristics

### Latency Sources

1. **Network RTT**: Client to ingress to partition
2. **Log Replication**: Wait for quorum acknowledgment
3. **Processing**: Handler execution time
4. **External Calls**: Durable step execution

### Optimization Strategies

**Reduce Cross-Partition Calls**:
- Co-locate related entities
- Use keyed routing effectively
- Minimize distributed coordination

**Batch Operations**:
- Group state updates
- Batch log appends
- Reduce round trips

**Followers for Read-Heavy**:
- Shared handlers can read from followers
- Reduces leader load
- Increases read throughput

## Consistency Guarantees

### Per-Partition

- **Sequential Consistency**: All operations on a key are totally ordered
- **Linearizability**: Writes committed in log order
- **Durability**: Committed operations survive failures

### Cross-Partition

- **Causal Consistency**: Within single invocation chain
- **Exactly-Once Delivery**: Messages delivered once despite failures
- **No Global Ordering**: Operations across partitions may interleave

## Replication Strategies

### Log Replication

**Quorum-Based**:
- Configurable replication factor (default: 1)
- Writes commit when quorum acknowledges
- Example: RF=3, quorum=2 (tolerates 1 failure)

### State Replication

**Leader-Based**:
- Single leader per partition
- Followers tail log asynchronously
- Can be promoted to leader on failure

## Deployment Topologies

### Single Node

- All roles on one node
- No replication (RF=1)
- Suitable for development
- Data persists to local disk

### Multi-Node Cluster

- Specialized roles across nodes
- Log replication (RF ≥ 2)
- Fault tolerance
- Object storage for snapshots

### Kubernetes

- StatefulSets for workers
- Persistent volumes for data
- Service load balancing
- Horizontal scaling

## Monitoring Points

### Metrics to Watch

**Partition Health**:
- Leader elections
- Follower lag
- Replication delay

**Log Performance**:
- Append latency
- Commit latency
- Segment growth rate

**Processor Performance**:
- Invocation throughput
- Queue depth
- Processing latency

**Storage**:
- RocksDB size
- Log segment count
- Snapshot frequency

## Design Patterns

### Partition-Local Operations

**Good**: Single Virtual Object operations
```go
// All state local to partition
func (c *Counter) Increment(ctx restate.ObjectContext, delta int) (int, error) {
    count, _ := restate.Get[int](ctx, "count")
    newCount := count + delta
    restate.Set(ctx, "count", newCount)
    return newCount, nil
}
```

### Cross-Partition Coordination

**Acceptable**: Workflows coordinating multiple objects
```go
// Workflow can coordinate across partitions
func (w *Workflow) Run(ctx restate.WorkflowContext, input Input) (Output, error) {
    // Calls to different partitions
    result1 := restate.Object[R1](ctx, "Obj1", "key1", "Method").Request(data1)
    result2 := restate.Object[R2](ctx, "Obj2", "key2", "Method").Request(data2)
    return combine(result1, result2), nil
}
```

### Avoid: Excessive Cross-Partition

**Anti-pattern**: Chatty cross-partition communication
```go
// BAD: Too many cross-partition calls
for _, item := range manyItems {
    restate.Object[void](ctx, "Processor", item.ID, "Process").Request(item)
}
```

**Better**: Batch or co-locate
```go
// GOOD: Batch processing
restate.Service[void](ctx, "BatchProcessor", "ProcessBatch").Request(manyItems)
```

## References

- Official Docs: https://docs.restate.dev/references/architecture
- Docker Deployment: See restate-docker-deploy skill
- Server Configuration: See restate-server-config skill
- Partition design impacts service performance
