# Architecture Overview

This document provides an overview of LinkFlow's architecture and design decisions.

## Three-Tier Architecture

LinkFlow uses a three-tier architecture consisting of:

1. **API Service** - Handles HTTP requests, authentication, and WebSocket connections
2. **Worker Service** - Executes workflows and processes background jobs
3. **Scheduler Service** - Manages cron jobs and system maintenance tasks

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         PRODUCTION ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                           ┌─────────────┐                                │
│                           │ Load Balancer│                               │
│                           └──────┬──────┘                                │
│                                  │                                       │
│                                  ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                        API SERVICE                                 │  │
│  │                   (Horizontally Scalable)                          │  │
│  │  • HTTP REST API          • WebSocket connections                  │  │
│  │  • Authentication         • Rate limiting                          │  │
│  └───────────────────────────────┬───────────────────────────────────┘  │
│                                  │                                       │
│                          Redis Queue                                     │
│                                  │                                       │
│                                  ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                       WORKER SERVICE                               │  │
│  │                  (Massively Scalable: 2-100+)                      │  │
│  │  • Workflow execution     • Node execution                         │  │
│  │  • Retry handling         • Timeout management                     │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                      SCHEDULER SERVICE                             │  │
│  │                    (Single Active Instance)                        │  │
│  │  • Cron job scheduling    • Dead letter processing                 │  │
│  │  • Stale job recovery     • Data cleanup                           │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │  PostgreSQL │    │    Redis    │    │  S3/Minio   │                  │
│  │  (Primary)  │    │(Queue+Cache)│    │  (Storage)  │                  │
│  └─────────────┘    └─────────────┘    └─────────────┘                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Service Responsibilities

### API Service

The API service is stateless and handles all client-facing requests:

- **HTTP REST API** - CRUD operations for workflows, credentials, executions
- **Authentication** - JWT tokens, API keys, OAuth
- **WebSocket** - Real-time execution updates
- **Rate Limiting** - Per-user and per-endpoint limits
- **Request Validation** - Input validation and sanitization

**Scaling**: Horizontal (2-20 instances based on request load)

### Worker Service

The worker service processes background jobs from Redis queues:

- **Workflow Execution** - Execute workflow DAGs
- **Node Execution** - Execute individual nodes
- **Retry Handling** - Exponential backoff for failures
- **Integration Calls** - External API requests
- **Notifications** - Email, Slack, webhook delivery

**Scaling**: Horizontal (2-100+ instances based on queue depth)

### Scheduler Service

The scheduler manages time-based operations using leader election:

- **Cron Jobs** - Schedule workflow executions
- **Dead Letter Processing** - Handle failed jobs
- **Stale Job Recovery** - Re-queue stuck jobs
- **Data Cleanup** - Remove old executions
- **Usage Aggregation** - Calculate billing metrics

**Scaling**: 2 instances (one active, one standby)

## Communication Patterns

### Synchronous (HTTP)
```
Client ──HTTP──▶ API Service ──HTTP Response──▶ Client
```
Used for CRUD operations, authentication, and data queries.

### Asynchronous (Queue)
```
API Service ──LPUSH──▶ Redis Queue ──BRPOP──▶ Worker
```
Used for workflow execution, notifications, and webhook delivery.

### Real-time (Pub/Sub)
```
Worker ──PUBLISH──▶ Redis Pub/Sub ──SUBSCRIBE──▶ API ──WebSocket──▶ Client
```
Used for execution progress updates and live notifications.

## Data Flow

### Workflow Execution Flow

1. Client sends `POST /api/v1/workflows/{id}/execute`
2. API validates request, checks permissions and billing limits
3. API creates execution record in PostgreSQL (status: "queued")
4. API pushes job to Redis queue
5. API returns `202 Accepted` with execution ID
6. Worker pops job from queue
7. Worker updates execution status to "running"
8. Worker executes nodes one by one
9. Worker publishes progress events to Redis Pub/Sub
10. API receives events and forwards to WebSocket clients
11. Worker updates execution status to "completed" or "failed"

## Key Design Decisions

### Why Three Services?

- **Separation of Concerns** - Each service has a clear responsibility
- **Independent Scaling** - Scale workers without scaling API
- **Fault Isolation** - Worker failure doesn't affect API
- **Simpler Deployment** - Smaller, focused deployments

### Why Redis for Queues?

- **Persistence** - AOF ensures no job loss
- **Pub/Sub** - Built-in real-time messaging
- **Performance** - Sub-millisecond latency
- **Simplicity** - No additional infrastructure

### Why PostgreSQL?

- **ACID Compliance** - Strong consistency for workflows
- **JSON Support** - JSONB for flexible node configs
- **Partitioning** - Table partitioning for executions
- **Maturity** - Battle-tested, well-documented

## Further Reading

- [System Design](system-design.md) - Detailed design document
- [Data Model](data-model.md) - Database schema
- [Scaling Guide](../operations/scaling.md) - Horizontal scaling strategies
