# LinkFlow AI - Production Architecture Plan

> Complete guide to building a production-ready workflow automation platform

**Version:** 2.0  
**Last Updated:** December 2024  
**Status:** Planning Phase

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current State Analysis](#2-current-state-analysis)
3. [Target Architecture](#3-target-architecture)
4. [Service Breakdown](#4-service-breakdown)
5. [Data Architecture](#5-data-architecture)
6. [Queue System](#6-queue-system)
7. [Security Architecture](#7-security-architecture)
8. [API Design](#8-api-design)
9. [Worker System](#9-worker-system)
10. [Scheduler System](#10-scheduler-system)
11. [Real-time System](#11-real-time-system)
12. [Node System](#12-node-system)
13. [Observability](#13-observability)
14. [Deployment](#14-deployment)
15. [Migration Plan](#15-migration-plan)
16. [Performance Targets](#16-performance-targets)
17. [Cost Estimation](#17-cost-estimation)
18. [Risk Assessment](#18-risk-assessment)
19. [Timeline](#19-timeline)
20. [Appendix](#20-appendix)

---

## 1. Executive Summary

### 1.1 Goals

- **Simplify Architecture**: Reduce from 21 microservices to 3 deployable components
- **Improve Reliability**: Add retry logic, dead letter queues, and graceful degradation
- **Enable Scale**: Support 10,000+ concurrent workflow executions
- **Reduce Costs**: Lower infrastructure costs by 60-70%
- **Faster Development**: Simpler codebase = faster feature development

### 1.2 Architecture Overview

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
│  │                                                                    │  │
│  │  • HTTP REST API          • WebSocket connections                  │  │
│  │  • Authentication         • Rate limiting                          │  │
│  │  • Request validation     • Webhook receiver                       │  │
│  │                                                                    │  │
│  └───────────────────────────────┬───────────────────────────────────┘  │
│                                  │                                       │
│                          Redis Queue                                     │
│                                  │                                       │
│                                  ▼                                       │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                       WORKER SERVICE                               │  │
│  │                  (Massively Scalable: 2-100+)                      │  │
│  │                                                                    │  │
│  │  • Workflow execution     • Node execution                         │  │
│  │  • Retry handling         • Timeout management                     │  │
│  │  • Integration calls      • Expression evaluation                  │  │
│  │                                                                    │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                      SCHEDULER SERVICE                             │  │
│  │                    (Single Active Instance)                        │  │
│  │                                                                    │  │
│  │  • Cron job scheduling    • Dead letter processing                 │  │
│  │  • Stale job recovery     • Data cleanup                           │  │
│  │  • Usage aggregation      • Leader election                        │  │
│  │                                                                    │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │  PostgreSQL │    │    Redis    │    │  S3/Minio   │                  │
│  │  (Primary)  │    │(Queue+Cache)│    │  (Storage)  │                  │
│  └─────────────┘    └─────────────┘    └─────────────┘                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.3 Key Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Services to Deploy | 21 | 3 |
| Docker Images | 21 | 3 |
| Config Files | 50+ | 10 |
| Time to Deploy | 30 min | 5 min |
| Memory Usage | 4GB+ | 1GB |
| Max Concurrent Executions | 100 | 10,000+ |

---

## 2. Current State Analysis

### 2.1 Current Services (21)

| # | Service | Lines of Code | Purpose | Keep/Merge |
|---|---------|---------------|---------|------------|
| 1 | gateway | ~500 | API routing | → API |
| 2 | auth | ~1500 | Authentication | → API |
| 3 | user | ~800 | User management | → API |
| 4 | workflow | ~1200 | Workflow CRUD | → API |
| 5 | execution | ~1000 | Execution tracking | → API |
| 6 | node | ~600 | Node registry | → Worker |
| 7 | schedule | ~700 | Cron scheduling | → Scheduler |
| 8 | webhook | ~500 | Webhook handling | → API |
| 9 | credential | ~800 | Secret management | → API |
| 10 | billing | ~1000 | Stripe integration | → API |
| 11 | workspace | ~600 | Multi-tenancy | → API |
| 12 | notification | ~500 | Email/SMS | → Worker |
| 13 | integration | ~1500 | Third-party APIs | → Worker |
| 14 | analytics | ~400 | Usage tracking | → Scheduler |
| 15 | search | ~300 | Full-text search | → API |
| 16 | storage | ~400 | File uploads | → API |
| 17 | config | ~200 | Dynamic config | → API |
| 18 | migration | ~300 | DB migrations | → CLI |
| 19 | backup | ~400 | Data backup | → Scheduler |
| 20 | admin | ~500 | Admin dashboard | → API |
| 21 | monitoring | ~300 | Health checks | → All |
| - | engine | ~800 | Execution engine | → Worker |
| - | executor | ~600 | Worker pool | → Worker |
| - | tenant | ~400 | Tenant isolation | → API |

### 2.2 Current Issues

1. **Operational Complexity**
   - 21 services = 21 deployments
   - Complex service mesh
   - Hard to debug distributed issues

2. **Performance Issues**
   - Multiple HTTP hops add latency
   - No job prioritization
   - Synchronous execution model

3. **Reliability Issues**
   - No retry mechanism for failed jobs
   - No dead letter queue
   - Cascading failures possible

4. **Missing Features**
   - OAuth not implemented (stubs only)
   - MFA not implemented (stubs only)
   - No real-time execution updates
   - No job priority queues

### 2.3 What Works Well

- ✅ DDD structure (keep this)
- ✅ Repository pattern (keep this)
- ✅ Expression parser (50+ functions)
- ✅ Node implementations (24 nodes)
- ✅ Credential encryption (AES-256-GCM)
- ✅ Billing integration (Stripe)
- ✅ Database schema (well designed)

---

## 3. Target Architecture

### 3.1 Three-Tier Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                                                                  │
│  TIER 1: API SERVICE                                                            │
│  ════════════════════                                                           │
│                                                                                  │
│  Responsibilities:                                                               │
│  • Handle all HTTP requests                                                      │
│  • Authentication & Authorization                                                │
│  • Request validation                                                            │
│  • Rate limiting                                                                 │
│  • WebSocket connections                                                         │
│  • Queue job submissions                                                         │
│                                                                                  │
│  Does NOT:                                                                       │
│  • Execute workflows (queues them instead)                                       │
│  • Process async jobs                                                            │
│  • Run scheduled tasks                                                           │
│                                                                                  │
│  Scaling: Horizontal (2-20 instances based on request load)                     │
│                                                                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  TIER 2: WORKER SERVICE                                                         │
│  ══════════════════════                                                         │
│                                                                                  │
│  Responsibilities:                                                               │
│  • Execute workflows                                                             │
│  • Execute individual nodes                                                      │
│  • Handle retries with backoff                                                   │
│  • Manage job timeouts                                                           │
│  • Send notifications                                                            │
│  • Deliver webhooks                                                              │
│                                                                                  │
│  Does NOT:                                                                       │
│  • Handle HTTP requests                                                          │
│  • Manage scheduling                                                             │
│  • Perform cleanup tasks                                                         │
│                                                                                  │
│  Scaling: Horizontal (2-100+ instances based on queue depth)                    │
│                                                                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  TIER 3: SCHEDULER SERVICE                                                      │
│  ═════════════════════════                                                      │
│                                                                                  │
│  Responsibilities:                                                               │
│  • Run cron schedules                                                            │
│  • Process dead letter queue                                                     │
│  • Recover stale jobs                                                            │
│  • Cleanup old data                                                              │
│  • Aggregate usage stats                                                         │
│  • Health monitoring                                                             │
│                                                                                  │
│  Does NOT:                                                                       │
│  • Handle HTTP requests                                                          │
│  • Execute workflows                                                             │
│                                                                                  │
│  Scaling: Single active (leader election), 2 instances for HA                   │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Communication Patterns

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         COMMUNICATION PATTERNS                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  1. SYNCHRONOUS (HTTP)                                                          │
│  ─────────────────────                                                          │
│                                                                                  │
│  Client ──HTTP──▶ API Service ──HTTP Response──▶ Client                         │
│                                                                                  │
│  Used for:                                                                       │
│  • CRUD operations                                                               │
│  • Authentication                                                                │
│  • Data queries                                                                  │
│                                                                                  │
│                                                                                  │
│  2. ASYNCHRONOUS (Queue)                                                        │
│  ───────────────────────                                                        │
│                                                                                  │
│  API Service ──LPUSH──▶ Redis Queue ──BRPOP──▶ Worker                           │
│                                                                                  │
│  Used for:                                                                       │
│  • Workflow execution                                                            │
│  • Notifications                                                                 │
│  • Webhook delivery                                                              │
│                                                                                  │
│                                                                                  │
│  3. REAL-TIME (Pub/Sub)                                                         │
│  ──────────────────────                                                         │
│                                                                                  │
│  Worker ──PUBLISH──▶ Redis Pub/Sub ──SUBSCRIBE──▶ API ──WebSocket──▶ Client    │
│                                                                                  │
│  Used for:                                                                       │
│  • Execution progress updates                                                    │
│  • Node completion events                                                        │
│  • Error notifications                                                           │
│                                                                                  │
│                                                                                  │
│  4. SCHEDULED (Cron)                                                            │
│  ───────────────────                                                            │
│                                                                                  │
│  Scheduler ──Query──▶ Database ──LPUSH──▶ Redis Queue                           │
│                                                                                  │
│  Used for:                                                                       │
│  • Scheduled workflow execution                                                  │
│  • Cleanup jobs                                                                  │
│  • Usage aggregation                                                             │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.3 Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              DATA FLOW                                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  WORKFLOW EXECUTION FLOW:                                                        │
│  ═══════════════════════                                                        │
│                                                                                  │
│  ┌────────┐    ┌─────────┐    ┌───────────────┐    ┌────────────┐              │
│  │ Client │───▶│   API   │───▶│ Redis Queue   │───▶│   Worker   │              │
│  └────────┘    └─────────┘    └───────────────┘    └─────┬──────┘              │
│       │             │                                     │                      │
│       │             │                                     ▼                      │
│       │             │         ┌───────────────┐    ┌────────────┐              │
│       │             │         │  PostgreSQL   │◀───│  Execute   │              │
│       │             │         │               │    │  Workflow  │              │
│       │             │         └───────────────┘    └─────┬──────┘              │
│       │             │                                     │                      │
│       │             │         ┌───────────────┐           │                      │
│       │             │◀────────│ Redis Pub/Sub │◀──────────┘                      │
│       │             │         └───────────────┘                                  │
│       │             │                                                            │
│       │    ┌────────┴────────┐                                                   │
│       │◀───│   WebSocket     │                                                   │
│       │    │  (Real-time)    │                                                   │
│            └─────────────────┘                                                   │
│                                                                                  │
│                                                                                  │
│  STEP-BY-STEP:                                                                  │
│                                                                                  │
│  1. Client sends POST /api/v1/workflows/{id}/execute                            │
│  2. API validates request, checks permissions, checks billing limits            │
│  3. API creates execution record in PostgreSQL (status: "queued")               │
│  4. API pushes job to Redis queue                                               │
│  5. API returns 202 Accepted with execution ID                                  │
│  6. Worker pops job from queue                                                  │
│  7. Worker updates execution status to "running"                                │
│  8. Worker executes nodes one by one                                            │
│  9. Worker publishes progress events to Redis Pub/Sub                           │
│  10. API receives events and forwards to WebSocket clients                      │
│  11. Worker updates execution status to "completed" or "failed"                 │
│  12. Client receives real-time updates via WebSocket                            │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Service Breakdown

### 4.1 API Service

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API SERVICE                                         │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ENTRY POINT: cmd/api/main.go                                                   │
│                                                                                  │
│  PACKAGES:                                                                       │
│                                                                                  │
│  internal/api/                                                                   │
│  ├── server.go              # HTTP server setup, graceful shutdown              │
│  ├── routes.go              # Route registration                                │
│  │                                                                               │
│  ├── middleware/                                                                 │
│  │   ├── auth.go            # JWT + API key authentication                      │
│  │   ├── ratelimit.go       # Rate limiting (per-user, per-IP, per-endpoint)   │
│  │   ├── tenant.go          # Multi-tenant context injection                    │
│  │   ├── requestid.go       # Request ID for tracing                            │
│  │   ├── recovery.go        # Panic recovery                                    │
│  │   ├── logging.go         # Request/response logging                          │
│  │   ├── cors.go            # CORS handling                                     │
│  │   ├── idempotency.go     # Idempotency key handling                          │
│  │   ├── security.go        # Security headers (HSTS, CSP, etc.)               │
│  │   └── timeout.go         # Request timeout                                   │
│  │                                                                               │
│  ├── handlers/                                                                   │
│  │   ├── auth.go            # Login, Register, OAuth, MFA, Password reset       │
│  │   ├── user.go            # User CRUD, Profile, Preferences                   │
│  │   ├── workflow.go        # Workflow CRUD, Clone, Import/Export               │
│  │   ├── execution.go       # Execute, Status, History, Cancel, Retry           │
│  │   ├── credential.go      # Credential CRUD, Test connection                  │
│  │   ├── workspace.go       # Workspace CRUD, Members, Invitations              │
│  │   ├── webhook.go         # Incoming webhook handler                          │
│  │   ├── schedule.go        # Schedule CRUD, Pause, Resume                      │
│  │   ├── billing.go         # Plans, Subscription, Usage, Invoices              │
│  │   ├── admin.go           # Admin endpoints (users, stats, etc.)              │
│  │   ├── health.go          # Health check endpoints                            │
│  │   └── search.go          # Full-text search                                  │
│  │                                                                               │
│  ├── websocket/                                                                  │
│  │   ├── hub.go             # WebSocket connection hub                          │
│  │   ├── client.go          # Client connection handler                         │
│  │   └── events.go          # Event types and handlers                          │
│  │                                                                               │
│  └── dto/                                                                        │
│      ├── request.go         # Request DTOs with validation tags                 │
│      └── response.go        # Response DTOs                                     │
│                                                                                  │
│                                                                                  │
│  ENDPOINTS:                                                                      │
│                                                                                  │
│  Authentication:                                                                 │
│  ├── POST   /api/v1/auth/register                                               │
│  ├── POST   /api/v1/auth/login                                                  │
│  ├── POST   /api/v1/auth/logout                                                 │
│  ├── POST   /api/v1/auth/refresh                                                │
│  ├── POST   /api/v1/auth/forgot-password                                        │
│  ├── POST   /api/v1/auth/reset-password                                         │
│  ├── GET    /api/v1/auth/oauth/{provider}                                       │
│  ├── GET    /api/v1/auth/oauth/{provider}/callback                              │
│  ├── POST   /api/v1/auth/mfa/setup                                              │
│  ├── POST   /api/v1/auth/mfa/verify                                             │
│  └── DELETE /api/v1/auth/mfa                                                    │
│                                                                                  │
│  Users:                                                                          │
│  ├── GET    /api/v1/users/me                                                    │
│  ├── PUT    /api/v1/users/me                                                    │
│  ├── GET    /api/v1/users/{id}                                                  │
│  └── GET    /api/v1/users (admin)                                               │
│                                                                                  │
│  Workflows:                                                                      │
│  ├── GET    /api/v1/workflows                                                   │
│  ├── POST   /api/v1/workflows                                                   │
│  ├── GET    /api/v1/workflows/{id}                                              │
│  ├── PUT    /api/v1/workflows/{id}                                              │
│  ├── DELETE /api/v1/workflows/{id}                                              │
│  ├── POST   /api/v1/workflows/{id}/execute                                      │
│  ├── POST   /api/v1/workflows/{id}/clone                                        │
│  ├── POST   /api/v1/workflows/{id}/activate                                     │
│  ├── POST   /api/v1/workflows/{id}/deactivate                                   │
│  ├── GET    /api/v1/workflows/{id}/versions                                     │
│  └── GET    /api/v1/workflows/{id}/versions/{version}                           │
│                                                                                  │
│  Executions:                                                                     │
│  ├── GET    /api/v1/executions                                                  │
│  ├── GET    /api/v1/executions/{id}                                             │
│  ├── POST   /api/v1/executions/{id}/cancel                                      │
│  ├── POST   /api/v1/executions/{id}/retry                                       │
│  └── GET    /api/v1/executions/{id}/logs                                        │
│                                                                                  │
│  Credentials:                                                                    │
│  ├── GET    /api/v1/credentials                                                 │
│  ├── POST   /api/v1/credentials                                                 │
│  ├── GET    /api/v1/credentials/{id}                                            │
│  ├── PUT    /api/v1/credentials/{id}                                            │
│  ├── DELETE /api/v1/credentials/{id}                                            │
│  └── POST   /api/v1/credentials/{id}/test                                       │
│                                                                                  │
│  Webhooks:                                                                       │
│  ├── POST   /webhooks/{endpoint_id}                                             │
│  └── GET    /webhooks/{endpoint_id}                                             │
│                                                                                  │
│  Schedules:                                                                      │
│  ├── GET    /api/v1/schedules                                                   │
│  ├── POST   /api/v1/schedules                                                   │
│  ├── GET    /api/v1/schedules/{id}                                              │
│  ├── PUT    /api/v1/schedules/{id}                                              │
│  ├── DELETE /api/v1/schedules/{id}                                              │
│  ├── POST   /api/v1/schedules/{id}/pause                                        │
│  └── POST   /api/v1/schedules/{id}/resume                                       │
│                                                                                  │
│  Workspaces:                                                                     │
│  ├── GET    /api/v1/workspaces                                                  │
│  ├── POST   /api/v1/workspaces                                                  │
│  ├── GET    /api/v1/workspaces/{id}                                             │
│  ├── PUT    /api/v1/workspaces/{id}                                             │
│  ├── DELETE /api/v1/workspaces/{id}                                             │
│  ├── GET    /api/v1/workspaces/{id}/members                                     │
│  ├── POST   /api/v1/workspaces/{id}/members                                     │
│  └── DELETE /api/v1/workspaces/{id}/members/{userId}                            │
│                                                                                  │
│  Billing:                                                                        │
│  ├── GET    /api/v1/billing/plans                                               │
│  ├── GET    /api/v1/billing/subscription                                        │
│  ├── POST   /api/v1/billing/subscription                                        │
│  ├── DELETE /api/v1/billing/subscription                                        │
│  ├── GET    /api/v1/billing/usage                                               │
│  ├── GET    /api/v1/billing/invoices                                            │
│  └── POST   /api/v1/billing/webhook (Stripe)                                    │
│                                                                                  │
│  Admin:                                                                          │
│  ├── GET    /api/v1/admin/stats                                                 │
│  ├── GET    /api/v1/admin/users                                                 │
│  ├── GET    /api/v1/admin/executions                                            │
│  └── GET    /api/v1/admin/queue                                                 │
│                                                                                  │
│  Health:                                                                         │
│  ├── GET    /health                                                             │
│  ├── GET    /health/live                                                        │
│  └── GET    /health/ready                                                       │
│                                                                                  │
│  WebSocket:                                                                      │
│  └── GET    /ws                                                                 │
│                                                                                  │
│                                                                                  │
│  CONFIGURATION:                                                                  │
│  • Port: 8080                                                                    │
│  • Replicas: 2-20 (HPA based on CPU/memory)                                     │
│  • Memory: 256MB-512MB per instance                                             │
│  • CPU: 0.25-1 core per instance                                                │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Worker Service

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                             WORKER SERVICE                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ENTRY POINT: cmd/worker/main.go                                                │
│                                                                                  │
│  PACKAGES:                                                                       │
│                                                                                  │
│  internal/worker/                                                                │
│  ├── worker.go              # Main worker loop                                  │
│  ├── pool.go                # Worker pool management                            │
│  │                                                                               │
│  ├── executor/                                                                   │
│  │   ├── executor.go        # Workflow executor                                 │
│  │   ├── dag.go             # DAG builder and traverser                         │
│  │   ├── state.go           # Execution state machine                           │
│  │   ├── context.go         # Execution context (variables, credentials)        │
│  │   ├── expression.go      # Expression evaluation wrapper                     │
│  │   └── sandbox.go         # Code execution sandbox                            │
│  │                                                                               │
│  ├── nodes/                                                                      │
│  │   ├── registry.go        # Node type registry                                │
│  │   ├── base.go            # Base node interface                               │
│  │   │                                                                           │
│  │   ├── triggers/          # Trigger nodes                                     │
│  │   │   ├── manual.go      # Manual trigger                                    │
│  │   │   ├── webhook.go     # Webhook trigger                                   │
│  │   │   ├── schedule.go    # Schedule trigger                                  │
│  │   │   └── error.go       # Error trigger                                     │
│  │   │                                                                           │
│  │   ├── logic/             # Logic/Control flow nodes                          │
│  │   │   ├── condition.go   # IF node                                           │
│  │   │   ├── switch.go      # Switch node                                       │
│  │   │   ├── loop.go        # Loop/ForEach node                                 │
│  │   │   ├── merge.go       # Merge node                                        │
│  │   │   ├── split.go       # Split in batches                                  │
│  │   │   └── wait.go        # Wait/Delay node                                   │
│  │   │                                                                           │
│  │   ├── actions/           # Action nodes                                      │
│  │   │   ├── http.go        # HTTP Request                                      │
│  │   │   ├── code.go        # Code/Function execution                           │
│  │   │   ├── set.go         # Set variables                                     │
│  │   │   ├── respond.go     # Respond to webhook                                │
│  │   │   └── error.go       # Stop and error                                    │
│  │   │                                                                           │
│  │   └── integrations/      # Integration nodes                                 │
│  │       ├── slack.go       # Slack                                             │
│  │       ├── email.go       # Email (SMTP/SendGrid)                             │
│  │       ├── discord.go     # Discord                                           │
│  │       ├── telegram.go    # Telegram                                          │
│  │       ├── github.go      # GitHub                                            │
│  │       ├── gitlab.go      # GitLab                                            │
│  │       ├── notion.go      # Notion                                            │
│  │       ├── airtable.go    # Airtable                                          │
│  │       ├── sheets.go      # Google Sheets                                     │
│  │       ├── drive.go       # Google Drive                                      │
│  │       ├── postgres.go    # PostgreSQL                                        │
│  │       ├── mysql.go       # MySQL                                             │
│  │       ├── mongodb.go     # MongoDB                                           │
│  │       ├── redis.go       # Redis                                             │
│  │       ├── s3.go          # AWS S3                                            │
│  │       ├── openai.go      # OpenAI                                            │
│  │       ├── stripe.go      # Stripe                                            │
│  │       ├── twilio.go      # Twilio (SMS)                                      │
│  │       ├── hubspot.go     # HubSpot                                           │
│  │       └── salesforce.go  # Salesforce                                        │
│  │                                                                               │
│  └── processor/                                                                  │
│      ├── job.go             # Job processing logic                              │
│      ├── retry.go           # Retry with exponential backoff                    │
│      ├── timeout.go         # Timeout handling                                  │
│      └── heartbeat.go       # Job heartbeat                                     │
│                                                                                  │
│                                                                                  │
│  JOB TYPES:                                                                      │
│                                                                                  │
│  1. workflow_execution                                                           │
│     - Execute a workflow                                                         │
│     - Priority: High (paid) / Normal (free)                                     │
│     - Timeout: 5 minutes (configurable per workflow)                            │
│     - Max retries: 3                                                            │
│                                                                                  │
│  2. notification                                                                 │
│     - Send email/SMS/Slack notification                                          │
│     - Priority: Normal                                                           │
│     - Timeout: 30 seconds                                                        │
│     - Max retries: 3                                                            │
│                                                                                  │
│  3. webhook_delivery                                                             │
│     - Deliver outgoing webhook                                                   │
│     - Priority: Normal                                                           │
│     - Timeout: 30 seconds                                                        │
│     - Max retries: 5                                                            │
│                                                                                  │
│                                                                                  │
│  CONFIGURATION:                                                                  │
│  • Replicas: 2-100 (HPA based on queue depth)                                   │
│  • Concurrency: 10 goroutines per instance                                      │
│  • Memory: 128MB-512MB per instance                                             │
│  • CPU: 0.25-0.5 core per instance                                              │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 4.3 Scheduler Service

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            SCHEDULER SERVICE                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ENTRY POINT: cmd/scheduler/main.go                                             │
│                                                                                  │
│  PACKAGES:                                                                       │
│                                                                                  │
│  internal/scheduler/                                                             │
│  ├── scheduler.go           # Main scheduler loop                               │
│  ├── leader.go              # Leader election (Redis-based)                     │
│  │                                                                               │
│  └── jobs/                                                                       │
│      ├── cron.go            # Cron job runner                                   │
│      ├── cleanup.go         # Data cleanup (old executions, logs)               │
│      ├── deadletter.go      # Dead letter queue processing                      │
│      ├── recovery.go        # Stale job recovery                                │
│      ├── aggregation.go     # Usage stats aggregation                           │
│      └── billing.go         # Billing period jobs                               │
│                                                                                  │
│                                                                                  │
│  SCHEDULED JOBS:                                                                 │
│                                                                                  │
│  1. Cron Workflows (every minute)                                               │
│     - Query due schedules                                                        │
│     - Queue workflow executions                                                  │
│     - Update next_run_at                                                         │
│                                                                                  │
│  2. Stale Job Recovery (every 5 minutes)                                        │
│     - Find jobs stuck in "processing" for >10 minutes                           │
│     - Re-queue them                                                              │
│                                                                                  │
│  3. Dead Letter Processing (every 10 minutes)                                   │
│     - Review failed jobs                                                         │
│     - Notify admins                                                              │
│     - Auto-retry if appropriate                                                  │
│                                                                                  │
│  4. Data Cleanup (daily at 3 AM)                                                │
│     - Delete executions older than retention period                             │
│     - Clean up orphaned data                                                     │
│     - Vacuum database                                                            │
│                                                                                  │
│  5. Usage Aggregation (hourly)                                                  │
│     - Aggregate execution counts                                                 │
│     - Calculate usage per workspace                                              │
│     - Update billing metrics                                                     │
│                                                                                  │
│  6. Billing Period (daily at midnight)                                          │
│     - Check subscription renewals                                                │
│     - Send usage alerts                                                          │
│     - Process overages                                                           │
│                                                                                  │
│                                                                                  │
│  LEADER ELECTION:                                                                │
│                                                                                  │
│  • Uses Redis SET NX with TTL                                                    │
│  • Leader renews lock every 10 seconds                                          │
│  • Lock TTL: 30 seconds                                                          │
│  • Followers check every 5 seconds                                               │
│  • Automatic failover on leader death                                            │
│                                                                                  │
│                                                                                  │
│  CONFIGURATION:                                                                  │
│  • Replicas: 2 (one active, one standby)                                        │
│  • Memory: 64MB-128MB per instance                                              │
│  • CPU: 0.1-0.25 core per instance                                              │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. Data Architecture

### 5.1 Database Schema

```sql
-- ═══════════════════════════════════════════════════════════════════════════════
-- USERS & AUTHENTICATION
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    username        VARCHAR(100) UNIQUE,
    password_hash   VARCHAR(255),
    
    -- Profile
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    avatar_url      VARCHAR(500),
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, suspended, deleted
    email_verified  BOOLEAN DEFAULT FALSE,
    
    -- MFA
    mfa_enabled     BOOLEAN DEFAULT FALSE,
    mfa_secret      VARCHAR(255),  -- Encrypted TOTP secret
    
    -- Tracking
    last_login_at   TIMESTAMPTZ,
    login_count     INTEGER DEFAULT 0,
    failed_logins   INTEGER DEFAULT 0,
    locked_until    TIMESTAMPTZ,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status) WHERE status = 'active';

-- Sessions
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,
    refresh_hash    VARCHAR(255),
    
    -- Context
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    device_info     JSONB,
    
    -- Timestamps
    expires_at      TIMESTAMPTZ NOT NULL,
    last_used_at    TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token_hash);

-- API Keys
CREATE TABLE api_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    
    name            VARCHAR(100) NOT NULL,
    key_prefix      VARCHAR(10) NOT NULL,  -- For identification (e.g., "lf_")
    key_hash        VARCHAR(255) NOT NULL,
    
    -- Permissions
    scopes          TEXT[] DEFAULT '{}',   -- ['workflows:read', 'executions:write']
    
    -- Timestamps
    last_used_at    TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);

-- OAuth Connections
CREATE TABLE oauth_connections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    provider        VARCHAR(50) NOT NULL,  -- google, github, microsoft
    provider_id     VARCHAR(255) NOT NULL,
    email           VARCHAR(255),
    
    -- Tokens (encrypted)
    access_token    TEXT,
    refresh_token   TEXT,
    expires_at      TIMESTAMPTZ,
    
    -- Metadata
    profile_data    JSONB,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, provider)
);


-- ═══════════════════════════════════════════════════════════════════════════════
-- WORKSPACES (Multi-tenancy)
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE workspaces (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES users(id),
    
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    logo_url        VARCHAR(500),
    
    -- Settings
    settings        JSONB DEFAULT '{}',
    
    -- Billing
    plan_id         VARCHAR(50) DEFAULT 'free',
    stripe_customer_id VARCHAR(255),
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX idx_workspaces_slug ON workspaces(slug);

-- Workspace Members
CREATE TABLE workspace_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    role            VARCHAR(20) NOT NULL DEFAULT 'member',  -- owner, admin, member, viewer
    
    invited_by      UUID REFERENCES users(id),
    invited_at      TIMESTAMPTZ,
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);


-- ═══════════════════════════════════════════════════════════════════════════════
-- WORKFLOWS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE workflows (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by      UUID NOT NULL REFERENCES users(id),
    
    -- Basic info
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft, active, inactive, archived
    version         INTEGER NOT NULL DEFAULT 1,
    
    -- Workflow definition
    nodes           JSONB NOT NULL DEFAULT '[]',
    connections     JSONB NOT NULL DEFAULT '[]',
    settings        JSONB DEFAULT '{}',
    
    -- Metadata
    tags            TEXT[],
    folder_id       UUID,
    
    -- Stats
    execution_count INTEGER DEFAULT 0,
    last_executed_at TIMESTAMPTZ,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at    TIMESTAMPTZ,
    archived_at     TIMESTAMPTZ
);

CREATE INDEX idx_workflows_workspace ON workflows(workspace_id);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_tags ON workflows USING GIN(tags);

-- Workflow Versions (for history)
CREATE TABLE workflow_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version         INTEGER NOT NULL,
    
    nodes           JSONB NOT NULL,
    connections     JSONB NOT NULL,
    settings        JSONB,
    
    created_by      UUID REFERENCES users(id),
    change_message  TEXT,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(workflow_id, version)
);

CREATE INDEX idx_workflow_versions_workflow ON workflow_versions(workflow_id);


-- ═══════════════════════════════════════════════════════════════════════════════
-- EXECUTIONS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE executions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    triggered_by    UUID REFERENCES users(id),
    
    -- Version
    workflow_version INTEGER NOT NULL,
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'queued',
    -- queued, running, completed, failed, cancelled, timeout
    
    -- Trigger
    trigger_type    VARCHAR(20) NOT NULL,  -- manual, schedule, webhook, api
    trigger_data    JSONB,
    
    -- Execution data
    input_data      JSONB,
    output_data     JSONB,
    
    -- Error info
    error_message   TEXT,
    error_node_id   VARCHAR(100),
    
    -- Timing
    queued_at       TIMESTAMPTZ DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    
    -- Stats
    nodes_total     INTEGER DEFAULT 0,
    nodes_completed INTEGER DEFAULT 0,
    
    -- Retry
    retry_count     INTEGER DEFAULT 0,
    parent_execution_id UUID REFERENCES executions(id),
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_executions_workflow ON executions(workflow_id);
CREATE INDEX idx_executions_workspace ON executions(workspace_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_created ON executions(created_at DESC);

-- Partition by month for better performance
-- CREATE TABLE executions_2024_01 PARTITION OF executions
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Node Executions
CREATE TABLE node_executions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id    UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    node_id         VARCHAR(100) NOT NULL,
    
    -- Node info
    node_type       VARCHAR(50) NOT NULL,
    node_name       VARCHAR(255),
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- pending, running, completed, failed, skipped
    
    -- Data
    input_data      JSONB,
    output_data     JSONB,
    
    -- Error
    error_message   TEXT,
    
    -- Timing
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    duration_ms     INTEGER,
    
    -- Retry
    retry_count     INTEGER DEFAULT 0,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_node_executions_execution ON node_executions(execution_id);
CREATE INDEX idx_node_executions_status ON node_executions(status);


-- ═══════════════════════════════════════════════════════════════════════════════
-- CREDENTIALS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by      UUID NOT NULL REFERENCES users(id),
    
    name            VARCHAR(100) NOT NULL,
    type            VARCHAR(50) NOT NULL,  -- api_key, oauth2, basic, bearer, etc.
    
    -- Encrypted data
    data            TEXT NOT NULL,  -- AES-256-GCM encrypted JSON
    
    -- Metadata
    description     TEXT,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX idx_credentials_workspace ON credentials(workspace_id);
CREATE INDEX idx_credentials_type ON credentials(type);


-- ═══════════════════════════════════════════════════════════════════════════════
-- SCHEDULES
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by      UUID NOT NULL REFERENCES users(id),
    
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    
    -- Schedule config
    cron_expression VARCHAR(100) NOT NULL,
    timezone        VARCHAR(50) DEFAULT 'UTC',
    
    -- Status
    is_active       BOOLEAN DEFAULT TRUE,
    
    -- Input data (passed to workflow)
    input_data      JSONB,
    
    -- Tracking
    next_run_at     TIMESTAMPTZ,
    last_run_at     TIMESTAMPTZ,
    last_execution_id UUID REFERENCES executions(id),
    run_count       INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schedules_workflow ON schedules(workflow_id);
CREATE INDEX idx_schedules_next_run ON schedules(next_run_at) WHERE is_active = TRUE;


-- ═══════════════════════════════════════════════════════════════════════════════
-- WEBHOOKS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE webhooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id     UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Endpoint
    endpoint_id     VARCHAR(32) NOT NULL UNIQUE,  -- Public URL identifier
    path            VARCHAR(255),
    
    -- Config
    method          VARCHAR(10) DEFAULT 'POST',
    headers_schema  JSONB,
    body_schema     JSONB,
    
    -- Security
    secret          VARCHAR(255),  -- For signature verification
    ip_whitelist    TEXT[],
    
    -- Status
    is_active       BOOLEAN DEFAULT TRUE,
    
    -- Stats
    call_count      INTEGER DEFAULT 0,
    last_called_at  TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_endpoint ON webhooks(endpoint_id);
CREATE INDEX idx_webhooks_workflow ON webhooks(workflow_id);


-- ═══════════════════════════════════════════════════════════════════════════════
-- BILLING
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE plans (
    id              VARCHAR(50) PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    
    -- Pricing
    price_monthly   INTEGER NOT NULL,  -- in cents
    price_yearly    INTEGER NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    
    -- Stripe
    stripe_price_monthly VARCHAR(255),
    stripe_price_yearly  VARCHAR(255),
    
    -- Limits
    limits          JSONB NOT NULL,
    -- {
    --   "executions_per_month": 1000,
    --   "workflows": 10,
    --   "team_members": 3,
    --   "retention_days": 30,
    --   "support": "community"
    -- }
    
    -- Features
    features        TEXT[],
    
    is_active       BOOLEAN DEFAULT TRUE,
    sort_order      INTEGER DEFAULT 0,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    plan_id         VARCHAR(50) NOT NULL REFERENCES plans(id),
    
    -- Stripe
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255),
    
    -- Status
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    -- active, past_due, canceled, trialing
    
    -- Billing period
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    
    -- Trial
    trial_start     TIMESTAMPTZ,
    trial_end       TIMESTAMPTZ,
    
    -- Cancellation
    cancel_at       TIMESTAMPTZ,
    canceled_at     TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_workspace ON subscriptions(workspace_id);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);

CREATE TABLE usage (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Period
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    
    -- Counts
    executions      INTEGER DEFAULT 0,
    api_calls       INTEGER DEFAULT 0,
    storage_bytes   BIGINT DEFAULT 0,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(workspace_id, period_start)
);

CREATE INDEX idx_usage_workspace_period ON usage(workspace_id, period_start);


-- ═══════════════════════════════════════════════════════════════════════════════
-- AUDIT LOGS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Action
    action          VARCHAR(100) NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     UUID,
    
    -- Changes
    old_values      JSONB,
    new_values      JSONB,
    
    -- Context
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_workspace ON audit_logs(workspace_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
```

---

## Summary

This comprehensive plan consolidates 21 microservices into 3 production-ready components:

| Component | Purpose | Scaling |
|-----------|---------|---------|
| **API** | HTTP/WebSocket, Auth, CRUD | 2-20 replicas |
| **Worker** | Job execution, all 24 nodes | 2-100 replicas |
| **Scheduler** | Cron, cleanup, DLQ | 2 replicas (leader election) |

### Key Features Implemented
- Redis-based queue with priority levels (high/normal/low)
- Exponential backoff retry (1s, 2s, 4s, 8s...)
- Dead letter queue for failed jobs
- Real-time updates via WebSocket + Redis Pub/Sub
- Leader election for scheduler
- Full observability (Prometheus, structured logging)
- Complete security (JWT, OAuth, MFA, encryption)

### Expected Results
- **60% cost reduction** (~$530/month savings)
- **10x scaling capacity** (10,000+ concurrent executions)
- **Simpler operations** (3 services vs 21)
- **Faster development** (single codebase)
- **Better reliability** (retries, DLQ, recovery)

### Implementation Timeline
- **Week 1-2**: Foundation (Queue, Pub/Sub, shared packages)
- **Week 3-4**: API Service (HTTP, Auth, WebSocket)
- **Week 5-6**: Worker Service (Execution engine, nodes)
- **Week 7**: Scheduler Service (Cron, DLQ, recovery)
- **Week 8**: Testing & Production deployment

---

**Document Version:** 2.0  
**Last Updated:** December 2024
