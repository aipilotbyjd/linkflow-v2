# Horizontal Scaling Guide

This guide explains how to scale Linkflow for high availability and performance.

## Architecture Overview

```
                              ┌─────────────────┐
                              │  Load Balancer  │
                              │   (nginx/ALB)   │
                              └────────┬────────┘
                                       │
              ┌────────────────────────┼────────────────────────┐
              │                        │                        │
       ┌──────▼──────┐          ┌──────▼──────┐          ┌──────▼──────┐
       │   API #1    │          │   API #2    │          │   API #3    │
       │  (Stateless)│          │  (Stateless)│          │  (Stateless)│
       └──────┬──────┘          └──────┬──────┘          └──────┬──────┘
              │                        │                        │
              └────────────────────────┼────────────────────────┘
                                       │
       ┌───────────────────────────────┼───────────────────────────────┐
       │                               │                               │
┌──────▼──────┐                 ┌──────▼──────┐                 ┌──────▼──────┐
│   Redis     │                 │  PostgreSQL │                 │    Redis    │
│  (Primary)  │                 │  (Primary)  │                 │   Cluster   │
└──────┬──────┘                 └──────┬──────┘                 └─────────────┘
       │                               │
┌──────▼──────┐                 ┌──────▼──────┐
│   Redis     │                 │  PostgreSQL │
│  (Replica)  │                 │  (Replicas) │
└─────────────┘                 └─────────────┘

       ┌────────────────────────────────────────────────────────┐
       │                      Worker Pool                        │
       │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐   │
       │  │Worker #1│  │Worker #2│  │Worker #3│  │Worker #N│   │
       │  └─────────┘  └─────────┘  └─────────┘  └─────────┘   │
       └────────────────────────────────────────────────────────┘
```

## Components

### 1. API Servers (Stateless)

API servers are completely stateless and can be scaled horizontally.

```yaml
# docker-compose.scale.yml
services:
  api:
    image: linkflow/api
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 1G
    environment:
      - DATABASE_URL=postgres://...
      - REDIS_URL=redis://...
```

**Scaling Triggers:**
- CPU > 70% for 5 minutes
- Memory > 80%
- Response time > 500ms p95
- Request rate > 1000 req/s per instance

### 2. Workers

Workers process background jobs from Redis queues.

```yaml
services:
  worker:
    image: linkflow/worker
    deploy:
      replicas: 5
      resources:
        limits:
          cpus: '2'
          memory: 2G
    environment:
      - WORKER_CONCURRENCY=10
      - QUEUE_PRIORITIES=critical:6,default:3,low:1
```

**Scaling Strategy:**
- Scale based on queue depth
- Target: queue wait time < 5 seconds
- Each worker handles 10 concurrent jobs by default

### 3. Scheduler (Single Leader)

Only one scheduler instance runs at a time using leader election.

```yaml
services:
  scheduler:
    image: linkflow/scheduler
    deploy:
      replicas: 2  # Hot standby
    environment:
      - LEADER_ELECTION_KEY=scheduler-leader
      - LEADER_TTL=30s
```

The scheduler uses Redis-based leader election to ensure only one instance processes schedules.

## Database Scaling

### PostgreSQL

**Read Replicas:**
```yaml
# Primary
postgresql_primary:
  image: postgres:15
  environment:
    - POSTGRES_REPLICATION_MODE=master

# Replicas
postgresql_replica:
  image: postgres:15
  environment:
    - POSTGRES_REPLICATION_MODE=slave
    - POSTGRES_MASTER_HOST=postgresql_primary
  deploy:
    replicas: 2
```

**Connection Pooling with PgBouncer:**
```yaml
pgbouncer:
  image: pgbouncer/pgbouncer
  environment:
    - DATABASES_HOST=postgresql_primary
    - POOL_MODE=transaction
    - MAX_CLIENT_CONN=1000
    - DEFAULT_POOL_SIZE=20
```

### Redis

**Redis Cluster for High Availability:**
```yaml
redis:
  image: redis:7
  command: redis-server --cluster-enabled yes
  deploy:
    replicas: 6  # 3 masters, 3 replicas
```

**Sentinel for Automatic Failover:**
```yaml
redis-sentinel:
  image: redis:7
  command: redis-sentinel /etc/sentinel.conf
  deploy:
    replicas: 3
```

## Kubernetes Deployment

### API Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkflow-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: linkflow-api
  template:
    metadata:
      labels:
        app: linkflow-api
    spec:
      containers:
      - name: api
        image: linkflow/api:latest
        ports:
        - containerPort: 8090
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8090
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8090
          initialDelaySeconds: 15
          periodSeconds: 20
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: linkflow-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: linkflow-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Worker Deployment with KEDA

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: linkflow-worker
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: worker
        image: linkflow/worker:latest
        env:
        - name: WORKER_CONCURRENCY
          value: "10"
---
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: linkflow-worker-scaler
spec:
  scaleTargetRef:
    name: linkflow-worker
  minReplicaCount: 2
  maxReplicaCount: 20
  triggers:
  - type: redis
    metadata:
      address: redis:6379
      listName: asynq:{default}
      listLength: "100"  # Scale up when > 100 jobs pending
```

## Load Balancing

### Nginx Configuration

```nginx
upstream linkflow_api {
    least_conn;
    server api1:8090 weight=5;
    server api2:8090 weight=5;
    server api3:8090 weight=5;
    keepalive 32;
}

server {
    listen 80;
    server_name api.linkflow.ai;

    location / {
        proxy_pass http://linkflow_api;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 5s;
        proxy_read_timeout 60s;
    }

    location /ws {
        proxy_pass http://linkflow_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
    }
}
```

### WebSocket Sticky Sessions

For WebSocket connections, use sticky sessions:

```nginx
upstream linkflow_ws {
    ip_hash;  # Sticky sessions based on client IP
    server api1:8090;
    server api2:8090;
    server api3:8090;
}
```

## Queue Partitioning

For high-volume workloads, partition queues by workspace or workflow:

```go
// Partition key based on workspace
func getQueueName(workspaceID uuid.UUID) string {
    partition := workspaceID[0] % 10  // 10 partitions
    return fmt.Sprintf("workflow:partition:%d", partition)
}
```

Deploy dedicated workers per partition:

```yaml
worker-partition-0:
  environment:
    - QUEUE_NAME=workflow:partition:0
worker-partition-1:
  environment:
    - QUEUE_NAME=workflow:partition:1
# ... etc
```

## Monitoring & Alerts

### Key Metrics to Monitor

| Metric | Warning | Critical |
|--------|---------|----------|
| API Response Time (p95) | > 500ms | > 2s |
| Queue Depth | > 1000 | > 5000 |
| Queue Wait Time | > 30s | > 2min |
| Worker CPU | > 70% | > 90% |
| Database Connections | > 80% pool | > 95% pool |
| Redis Memory | > 70% | > 90% |
| Error Rate | > 1% | > 5% |

### Prometheus Alerts

```yaml
groups:
- name: linkflow
  rules:
  - alert: HighQueueDepth
    expr: asynq_queue_size{queue="default"} > 5000
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: Queue depth critical
      
  - alert: WorkerDown
    expr: up{job="linkflow-worker"} == 0
    for: 1m
    labels:
      severity: critical
      
  - alert: HighResponseTime
    expr: histogram_quantile(0.95, http_request_duration_seconds_bucket) > 2
    for: 5m
    labels:
      severity: warning
```

## Best Practices

### 1. Graceful Shutdown

All services implement graceful shutdown:

```go
// Wait for in-flight requests
server.Shutdown(ctx)

// Wait for workers to finish current jobs
worker.Shutdown()
```

### 2. Health Checks

```
GET /health       - Overall health
GET /health/live  - Liveness (process is running)
GET /health/ready - Readiness (can accept traffic)
```

### 3. Connection Pooling

- Database: Use connection pooling (PgBouncer)
- Redis: Use connection pooling
- HTTP: Reuse connections with keep-alive

### 4. Caching Strategy

```
Level 1: In-memory (per-instance, short TTL)
Level 2: Redis (shared, medium TTL)
Level 3: Database (persistent)
```

### 5. Circuit Breakers

Protect against cascading failures:

```go
breaker := circuitbreaker.New(circuitbreaker.Config{
    FailureThreshold: 5,
    Timeout:          30 * time.Second,
})
```

## Capacity Planning

### Baseline Requirements

| Component | Small | Medium | Large |
|-----------|-------|--------|-------|
| API Instances | 2 | 4 | 8+ |
| Workers | 2 | 5 | 20+ |
| PostgreSQL | 2 vCPU, 4GB | 4 vCPU, 16GB | 8 vCPU, 32GB |
| Redis | 1 vCPU, 2GB | 2 vCPU, 8GB | 4 vCPU, 16GB |

### Scaling Formulas

**Workers needed:**
```
workers = (executions_per_hour * avg_execution_time_seconds) / (3600 * concurrency_per_worker)
```

**API instances needed:**
```
instances = requests_per_second / (max_rps_per_instance * 0.7)
```

## Disaster Recovery

### Backup Strategy

- PostgreSQL: Daily full backup, continuous WAL archiving
- Redis: RDB snapshots every hour, AOF for point-in-time recovery
- Configuration: Version controlled in Git

### Recovery Time Objectives

| Component | RTO | RPO |
|-----------|-----|-----|
| API | < 5 min | N/A (stateless) |
| Workers | < 5 min | N/A (idempotent) |
| Database | < 30 min | < 5 min |
| Redis | < 15 min | < 1 hour |

### Multi-Region Setup

For global availability:

```
Region A (Primary)          Region B (DR)
├── API Cluster            ├── API Cluster (standby)
├── Worker Pool            ├── Worker Pool (standby)
├── PostgreSQL Primary ──▶ ├── PostgreSQL Replica
├── Redis Primary ────────▶├── Redis Replica
└── Scheduler (Leader)     └── Scheduler (Standby)
```

Use DNS-based failover (Route53, Cloudflare) for automatic region switching.
