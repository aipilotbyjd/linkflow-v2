<coding_guidelines>
# LinkFlow - Workflow Automation Platform (v2 Clean Architecture)

LinkFlow is a production-ready n8n-like workflow automation platform built in Go using Clean Architecture principles. It enables users to create, execute, and monitor automated workflows through a visual interface with support for webhooks, schedules, and manual triggers.

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.23+ | Backend services |
| Web Framework | Chi v5 | HTTP routing |
| ORM | GORM | Database operations |
| Database | PostgreSQL 15+ | Persistent storage |
| Cache | Redis 7+ | Caching, sessions, leader election |
| Job Queue | Asynq | Background job processing |
| Auth | JWT + OAuth2 | Authentication |
| Metrics | Prometheus | Observability |
| Container | Docker | Deployment |

## Architecture Overview

### Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                        │
│                  (HTTP, WebSocket, CLI)                      │
│              internal/adapters/http/handlers/                │
└────────────────────────┬────────────────────────────────────┘
                         │ DTOs
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   APPLICATION LAYER                          │
│                   (Use Cases / CQRS)                         │
│         internal/core/application/command/                   │
│         internal/core/application/query/                     │
└────────────────────────┬────────────────────────────────────┘
                         │ Commands/Queries
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      DOMAIN LAYER                            │
│                (Business Logic & Rules)                      │
│         internal/core/domain/{aggregate}/                    │
│         - Entities, Value Objects, Domain Events             │
│         - Repository Interfaces (Ports)                      │
│         - Domain Services                                    │
│         - NO external dependencies                           │
└────────────────────────┬────────────────────────────────────┘
                         │ Interfaces
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  INFRASTRUCTURE LAYER                        │
│              (Framework & External Services)                 │
│         internal/adapters/persistence/                       │
│         internal/adapters/messaging/                         │
│         internal/infrastructure/                             │
└─────────────────────────────────────────────────────────────┘
```

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLIENTS                                     │
│                    (Web UI, Mobile, External Systems)                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
            ┌───────────┐   ┌───────────┐   ┌───────────┐
            │  REST API │   │ WebSocket │   │ Webhooks  │
            │  :8080    │   │    /ws    │   │ /webhooks │
            └─────┬─────┘   └─────┬─────┘   └─────┬─────┘
                  │               │               │
                  └───────────────┼───────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │        API SERVER         │
                    │     (cmd/api/main.go)     │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │     │     Redis       │     │   Asynq Queue   │
└─────────────────┘     └─────────────────┘     └────────┬────────┘
                                                         │
                        ┌────────────────────────────────┘
                        │
          ┌─────────────▼─────────────┐     ┌─────────────────────────┐
          │         WORKER            │     │       SCHEDULER         │
          │   (cmd/worker/main.go)    │     │ (cmd/scheduler/main.go) │
          │                           │     │                         │
          │  • 27 Node Types          │     │  • Leader Election      │
          │  • Workflow Execution     │     │  • Health Checks        │
          │  • Credential Support     │     │  • Metrics              │
          └───────────────────────────┘     └─────────────────────────┘
```

## Project Structure (v2 Clean Architecture)

```
linkflow-v2/
│
├── cmd/                                    # Application Entrypoints
│   ├── api/
│   │   ├── main.go                         # API server entry
│   │   └── wire.go                         # Dependency injection
│   ├── worker/
│   │   ├── main.go                         # Background worker entry
│   │   └── wire.go
│   ├── scheduler/
│   │   ├── main.go                         # Cron scheduler with leader election
│   │   └── wire.go
│   └── migrate/
│       └── main.go                         # Database migration CLI
│
├── internal/                               # Private Application Code
│   │
│   ├── core/                               # CORE BUSINESS LOGIC
│   │   │                                   # (Framework-agnostic, pure Go)
│   │   │
│   │   ├── domain/                         # Domain Models & Business Rules
│   │   │   ├── user/                       # User Aggregate
│   │   │   │   ├── user.go                 # User entity (aggregate root)
│   │   │   │   ├── session.go              # Session entity
│   │   │   │   ├── status.go               # Status value object
│   │   │   │   ├── repository.go           # Repository interface (port)
│   │   │   │   └── errors.go               # Domain errors
│   │   │   │
│   │   │   ├── workspace/                  # Workspace Aggregate
│   │   │   │   ├── workspace.go            # Workspace entity
│   │   │   │   ├── member.go               # Member entity
│   │   │   │   ├── role.go                 # Role value object
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── workflow/                   # Workflow Aggregate
│   │   │   │   ├── workflow.go             # Workflow entity
│   │   │   │   ├── version.go              # Version entity
│   │   │   │   ├── status.go               # Status value object
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── execution/                  # Execution Aggregate
│   │   │   │   ├── execution.go            # Execution entity
│   │   │   │   ├── node_execution.go       # NodeExecution entity
│   │   │   │   ├── status.go               # Status value object
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── credential/                 # Credential Aggregate
│   │   │   │   ├── credential.go           # Credential entity
│   │   │   │   ├── share.go                # Share entity
│   │   │   │   ├── types.go                # Credential types
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── schedule/                   # Schedule Aggregate
│   │   │   │   ├── schedule.go             # Schedule entity
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── webhook/                    # Webhook Aggregate
│   │   │   │   ├── endpoint.go             # WebhookEndpoint entity
│   │   │   │   ├── event.go                # WebhookEvent entity
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── billing/                    # Billing Aggregate
│   │   │   │   ├── subscription.go         # Subscription entity
│   │   │   │   ├── plan.go                 # Plan entity
│   │   │   │   ├── usage.go                # Usage entity
│   │   │   │   ├── invoice.go              # Invoice entity
│   │   │   │   ├── repository.go           # Repository interface
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   └── template/                   # Template Aggregate
│   │   │       ├── template.go
│   │   │       ├── category.go
│   │   │       ├── repository.go
│   │   │       └── errors.go
│   │   │
│   │   └── application/                    # Application Services (Use Cases)
│   │       │
│   │       ├── command/                    # Write Operations (CQRS)
│   │       │   ├── user/
│   │       │   │   ├── register_user.go
│   │       │   │   └── login_user.go
│   │       │   ├── workspace/
│   │       │   │   └── create_workspace.go
│   │       │   ├── workflow/
│   │       │   │   ├── create_workflow.go
│   │       │   │   ├── update_workflow.go
│   │       │   │   └── activate_workflow.go
│   │       │   ├── execution/
│   │       │   │   └── start_execution.go
│   │       │   ├── credential/
│   │       │   │   ├── create_credential.go
│   │       │   │   ├── update_credential.go
│   │       │   │   ├── delete_credential.go
│   │       │   │   ├── share_credential.go
│   │       │   │   └── test_credential.go
│   │       │   ├── schedule/
│   │       │   │   ├── create_schedule.go
│   │       │   │   ├── update_schedule.go
│   │       │   │   ├── pause_schedule.go
│   │       │   │   └── resume_schedule.go
│   │       │   └── webhook/
│   │       │       ├── create_endpoint.go
│   │       │       ├── regenerate_secret.go
│   │       │       └── trigger_webhook.go
│   │       │
│   │       └── query/                      # Read Operations (CQRS)
│   │           ├── user/
│   │           │   └── get_user.go
│   │           ├── workspace/
│   │           │   └── get_workspace.go
│   │           ├── workflow/
│   │           │   ├── get_workflow.go
│   │           │   └── get_versions.go
│   │           ├── execution/
│   │           │   └── get_execution.go
│   │           ├── credential/
│   │           │   ├── get_credential.go
│   │           │   ├── list_credentials.go
│   │           │   └── get_credential_shares.go
│   │           ├── schedule/
│   │           │   ├── get_schedule.go
│   │           │   └── list_schedules.go
│   │           └── analytics/
│   │               ├── get_workspace_analytics.go
│   │               ├── get_workflow_analytics.go
│   │               └── get_execution_metrics.go
│   │
│   ├── adapters/                           # ADAPTERS LAYER
│   │   │
│   │   ├── http/                           # HTTP Adapters (REST API)
│   │   │   ├── server.go                   # HTTP server setup
│   │   │   │
│   │   │   ├── routes/
│   │   │   │   └── routes.go               # All route registration
│   │   │   │
│   │   │   ├── handlers/                   # HTTP handlers (thin layer)
│   │   │   │   ├── auth/
│   │   │   │   │   ├── register.go
│   │   │   │   │   ├── login.go
│   │   │   │   │   ├── refresh.go
│   │   │   │   │   └── logout.go
│   │   │   │   ├── workspace/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   └── members.go
│   │   │   │   ├── workflow/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── clone.go
│   │   │   │   │   ├── activate.go
│   │   │   │   │   ├── deactivate.go
│   │   │   │   │   ├── versions.go
│   │   │   │   │   ├── rollback.go
│   │   │   │   │   ├── export.go
│   │   │   │   │   └── import.go
│   │   │   │   ├── execution/
│   │   │   │   │   ├── start.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── cancel.go
│   │   │   │   │   └── retry.go
│   │   │   │   ├── credential/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   └── delete.go
│   │   │   │   ├── schedule/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── pause.go
│   │   │   │   │   └── resume.go
│   │   │   │   ├── webhook/
│   │   │   │   │   └── trigger.go
│   │   │   │   ├── billing/
│   │   │   │   │   ├── get_plans.go
│   │   │   │   │   ├── get_subscription.go
│   │   │   │   │   ├── get_usage.go
│   │   │   │   │   ├── get_invoices.go
│   │   │   │   │   └── stripe_webhook.go
│   │   │   │   ├── analytics/
│   │   │   │   │   ├── workspace_analytics.go
│   │   │   │   │   └── workflow_analytics.go
│   │   │   │   ├── health/
│   │   │   │   │   └── health.go
│   │   │   │   └── admin/
│   │   │   │       ├── stream_stats.go
│   │   │   │       └── metrics.go
│   │   │   │
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go                 # JWT authentication
│   │   │   │   ├── apikey.go               # API key authentication
│   │   │   │   ├── tenant.go               # Workspace context
│   │   │   │   ├── ratelimit.go            # Rate limiting
│   │   │   │   ├── logging.go              # Request logging
│   │   │   │   ├── recovery.go             # Panic recovery
│   │   │   │   ├── cors.go                 # CORS handling
│   │   │   │   ├── metrics.go              # Prometheus metrics
│   │   │   │   ├── rbac.go                 # Role-based access control
│   │   │   │   └── idempotency.go          # Idempotency handling
│   │   │   │
│   │   │   └── dto/
│   │   │       └── common/
│   │   │           └── response.go         # Common response types
│   │   │
│   │   ├── persistence/                    # Database Adapters
│   │   │   ├── postgres/
│   │   │   │   ├── client.go               # Database client setup
│   │   │   │   ├── transaction.go          # Transaction wrapper
│   │   │   │   ├── models/                 # Database models
│   │   │   │   │   ├── user.go
│   │   │   │   │   ├── workspace.go
│   │   │   │   │   ├── workflow.go
│   │   │   │   │   └── execution.go
│   │   │   │   ├── repositories/           # Repository implementations
│   │   │   │   │   ├── user_repository.go
│   │   │   │   │   ├── workspace_repository.go
│   │   │   │   │   ├── member_repository.go
│   │   │   │   │   ├── workflow_repository.go
│   │   │   │   │   ├── version_repository.go
│   │   │   │   │   ├── execution_repository.go
│   │   │   │   │   ├── node_execution_repository.go
│   │   │   │   │   ├── credential_repository.go
│   │   │   │   │   ├── schedule_repository.go
│   │   │   │   │   ├── webhook_repository.go
│   │   │   │   │   └── session_repository.go
│   │   │   │   └── mappers/                # Domain ↔ DB mapping
│   │   │   │       ├── user_mapper.go
│   │   │   │       ├── workspace_mapper.go
│   │   │   │       ├── workflow_mapper.go
│   │   │   │       └── execution_mapper.go
│   │   │   │
│   │   │   └── redis/
│   │   │       └── client.go               # Redis client
│   │   │
│   │   ├── messaging/                      # Message Queue Adapters
│   │   │   └── asynq/
│   │   │       └── client.go               # Queue client
│   │   │
│   │   ├── websocket/                      # WebSocket Adapter
│   │   │   ├── hub.go                      # Connection hub
│   │   │   ├── client.go                   # WebSocket client
│   │   │   ├── subscriber.go               # Redis subscriber
│   │   │   ├── events.go                   # Event types
│   │   │   └── handler.go                  # Connection handler
│   │   │
│   │   ├── worker/                         # Background Worker Adapter
│   │   │   ├── executor/
│   │   │   │   ├── executor.go             # Main execution orchestrator
│   │   │   │   ├── processor.go            # Node processor
│   │   │   │   └── runtime.go              # Runtime context
│   │   │   │
│   │   │   ├── nodes/                      # Node Implementations (27 total)
│   │   │   │   ├── registry.go             # Node registry
│   │   │   │   ├── interface.go            # Node interface
│   │   │   │   ├── loader.go               # Auto-registration
│   │   │   │   │
│   │   │   │   ├── triggers/               # 3 Trigger Nodes
│   │   │   │   │   ├── manual.go
│   │   │   │   │   ├── webhook.go
│   │   │   │   │   └── schedule.go
│   │   │   │   │
│   │   │   │   ├── actions/                # 4 Action Nodes
│   │   │   │   │   ├── http/
│   │   │   │   │   │   └── http_request.go
│   │   │   │   │   ├── email/
│   │   │   │   │   │   └── send_email.go
│   │   │   │   │   ├── code/
│   │   │   │   │   │   └── javascript.go
│   │   │   │   │   └── transform/
│   │   │   │   │       └── set.go
│   │   │   │   │
│   │   │   │   ├── logic/                  # 10 Logic Nodes
│   │   │   │   │   ├── if.go
│   │   │   │   │   ├── switch.go
│   │   │   │   │   ├── merge.go
│   │   │   │   │   ├── filter.go
│   │   │   │   │   ├── sort.go
│   │   │   │   │   ├── limit.go
│   │   │   │   │   ├── aggregate.go
│   │   │   │   │   ├── loop.go
│   │   │   │   │   ├── wait.go
│   │   │   │   │   └── noop.go
│   │   │   │   │
│   │   │   │   └── integrations/           # 10 Integration Nodes
│   │   │   │       ├── ai/
│   │   │   │       │   ├── openai.go
│   │   │   │       │   └── anthropic.go
│   │   │   │       ├── cloud/
│   │   │   │       │   ├── aws_s3.go
│   │   │   │       │   └── google_drive.go
│   │   │   │       ├── communication/
│   │   │   │       │   ├── slack.go
│   │   │   │       │   ├── discord.go
│   │   │   │       │   ├── telegram.go
│   │   │   │       │   └── twilio.go
│   │   │   │       ├── database/
│   │   │   │       │   ├── postgres.go
│   │   │   │       │   ├── mysql.go
│   │   │   │       │   ├── mongodb.go
│   │   │   │       │   └── redis.go
│   │   │   │       ├── crm/
│   │   │   │       │   ├── salesforce.go
│   │   │   │       │   ├── hubspot.go
│   │   │   │       │   ├── airtable.go
│   │   │   │       │   └── notion.go
│   │   │   │       └── payment/
│   │   │   │           └── stripe.go
│   │   │   │
│   │   │   ├── middleware/                 # Worker middleware
│   │   │   │   ├── logging.go
│   │   │   │   ├── tracing.go
│   │   │   │   ├── recovery.go
│   │   │   │   ├── retry.go
│   │   │   │   └── metrics.go
│   │   │   │
│   │   │   ├── cache/
│   │   │   │   └── cache.go                # Worker caching
│   │   │   │
│   │   │   └── types/
│   │   │       └── node.go                 # Node metadata types
│   │   │
│   │   └── scheduler/                      # Scheduler Adapter (Production-ready)
│   │       ├── server.go                   # Scheduler server
│   │       ├── poller.go                   # Schedule poller
│   │       ├── dispatcher.go               # Job dispatcher
│   │       ├── leader_election.go          # Redis-based leader election (HA)
│   │       ├── cron.go                     # Cron parser
│   │       └── metrics.go                  # Scheduler metrics
│   │
│   ├── infrastructure/                     # SHARED INFRASTRUCTURE
│   │   │
│   │   ├── auth/
│   │   │   ├── jwt/
│   │   │   │   ├── manager.go              # JWT token management
│   │   │   │   └── blacklist.go            # Token blacklist
│   │   │   └── oauth/
│   │   │       ├── manager.go              # OAuth flow management
│   │   │       ├── provider.go             # Provider interface
│   │   │       └── providers/
│   │   │           ├── google.go
│   │   │           └── github.go
│   │   │
│   │   ├── crypto/
│   │   │   ├── encryption.go               # AES-256-GCM encryption
│   │   │   ├── hashing.go                  # bcrypt hashing
│   │   │   ├── otp.go                      # TOTP for MFA
│   │   │   └── signing.go                  # Webhook signatures
│   │   │
│   │   ├── email/
│   │   │   ├── service.go                  # Email service
│   │   │   ├── template.go                 # Template engine
│   │   │   ├── sendgrid.go                 # SendGrid HTTP API
│   │   │   ├── smtp.go                     # SMTP provider
│   │   │   └── templates/
│   │   │       ├── welcome.html
│   │   │       ├── reset_password.html
│   │   │       ├── invitation.html
│   │   │       └── execution_failed.html
│   │   │
│   │   ├── storage/
│   │   │   ├── storage.go                  # Storage interface
│   │   │   ├── s3/
│   │   │   │   └── client.go               # S3 implementation
│   │   │   └── local/
│   │   │       └── filesystem.go           # Local filesystem
│   │   │
│   │   ├── cache/
│   │   │   ├── cache.go                    # Cache interface
│   │   │   ├── redis_cache.go              # Redis implementation
│   │   │   ├── memory_cache.go             # In-memory implementation
│   │   │   └── noop_cache.go               # No-op implementation
│   │   │
│   │   ├── observability/
│   │   │   ├── logger/
│   │   │   │   └── logger.go               # Logger interface + zerolog
│   │   │   ├── metrics/
│   │   │   │   └── metrics.go              # Prometheus metrics
│   │   │   └── tracing/
│   │   │       └── tracing.go              # Tracing interface
│   │   │
│   │   ├── resilience/
│   │   │   ├── circuitbreaker.go           # Circuit breaker pattern
│   │   │   ├── retry.go                    # Retry with backoff
│   │   │   └── ratelimiter.go              # Rate limiting
│   │   │
│   │   ├── validation/
│   │   │   └── validator.go                # Input validation
│   │   │
│   │   └── config/
│   │       └── config.go                   # Configuration loader
│   │
│   └── shared/                             # SHARED KERNEL
│       ├── types/
│       │   ├── id.go                       # UUID wrapper
│       │   ├── pagination.go               # Pagination types
│       │   ├── filter.go                   # Filter types
│       │   └── json.go                     # JSON types
│       │
│       ├── errors/
│       │   ├── errors.go                   # Error types
│       │   └── codes.go                    # Error codes
│       │
│       └── events/
│           ├── event.go                    # Base event interface
│           └── bus.go                      # Event bus interface
│
├── configs/                                # Configuration
│   ├── config.yaml                         # Default config
│   ├── config.test.yaml                    # Test config
│   ├── config.staging.yaml                 # Staging config
│   └── config.production.yaml              # Production config
│
├── deploy/                                 # Deployment
│   └── ...
│
├── docs/                                   # Documentation
│   └── architecture/
│       └── STRUCTURE_V2_PLAN.md            # Architecture plan
│
├── go.mod
├── go.sum
├── Makefile
└── AGENTS.md                               # This file
```

## Dependency Rules

### The Clean Architecture Dependency Rule

**Layers can only depend on layers below them:**

```
Adapters → Application → Domain
Infrastructure → (Any)
Shared → (Nothing - base layer)
```

### Import Rules

**Allowed:**
```go
// Handler imports application
import "internal/core/application/command/workflow"

// Application imports domain
import "internal/core/domain/workflow"

// Adapter imports domain (for interfaces)
import "internal/core/domain/workflow"
```

**NOT Allowed:**
```go
// Domain imports adapter - WRONG!
import "internal/adapters/persistence/postgres"

// Domain imports infrastructure - WRONG!
import "internal/infrastructure/crypto"
```

## API Endpoints Reference

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Invalidate session |
| GET | `/api/v1/me` | Get current user |

### Workspaces
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces` | Create workspace |
| GET | `/api/v1/workspaces` | List workspaces |
| GET | `/api/v1/workspaces/{id}` | Get workspace |
| PUT | `/api/v1/workspaces/{id}` | Update workspace |
| DELETE | `/api/v1/workspaces/{id}` | Delete workspace |
| GET | `/api/v1/workspaces/{id}/members` | List members |

### Workflows (workspace-scoped)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces/{id}/workflows` | Create workflow |
| GET | `/api/v1/workspaces/{id}/workflows` | List workflows |
| GET | `/api/v1/workspaces/{id}/workflows/{wfId}` | Get workflow |
| PUT | `/api/v1/workspaces/{id}/workflows/{wfId}` | Update workflow |
| DELETE | `/api/v1/workspaces/{id}/workflows/{wfId}` | Delete workflow |
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/activate` | Activate |
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/deactivate` | Deactivate |
| GET | `/api/v1/workspaces/{id}/workflows/{wfId}/versions` | List versions |

### Executions
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/execute` | Start execution |
| GET | `/api/v1/workspaces/{id}/executions` | List executions |
| GET | `/api/v1/workspaces/{id}/executions/{exId}` | Get execution |
| POST | `/api/v1/workspaces/{id}/executions/{exId}/cancel` | Cancel |
| POST | `/api/v1/workspaces/{id}/executions/{exId}/retry` | Retry |

### Credentials
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces/{id}/credentials` | Create credential |
| GET | `/api/v1/workspaces/{id}/credentials` | List credentials |
| GET | `/api/v1/workspaces/{id}/credentials/{credId}` | Get credential |
| PUT | `/api/v1/workspaces/{id}/credentials/{credId}` | Update credential |
| DELETE | `/api/v1/workspaces/{id}/credentials/{credId}` | Delete credential |

### Schedules
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces/{id}/schedules` | Create schedule |
| GET | `/api/v1/workspaces/{id}/schedules` | List schedules |
| GET | `/api/v1/workspaces/{id}/schedules/{schId}` | Get schedule |
| PUT | `/api/v1/workspaces/{id}/schedules/{schId}` | Update schedule |
| DELETE | `/api/v1/workspaces/{id}/schedules/{schId}` | Delete schedule |
| POST | `/api/v1/workspaces/{id}/schedules/{schId}/pause` | Pause |
| POST | `/api/v1/workspaces/{id}/schedules/{schId}/resume` | Resume |

### Monitoring
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

### Scheduler Health (port 8091)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health/live` | Liveness probe |
| GET | `/health/ready` | Readiness probe (leader status) |
| GET | `/metrics` | Scheduler metrics |

## Node Types Reference (27 Total)

### Triggers (3)
| Type | Description |
|------|-------------|
| `trigger.manual` | Manual execution start |
| `trigger.webhook` | HTTP webhook trigger |
| `trigger.schedule` | Cron-based scheduling |

### Actions (4)
| Type | Description |
|------|-------------|
| `action.http_request` | HTTP request (GET, POST, etc.) |
| `action.send_email` | Send email |
| `action.javascript` | Execute JavaScript |
| `action.set` | Set variable values |

### Logic (10)
| Type | Description |
|------|-------------|
| `logic.if` | Conditional branching |
| `logic.switch` | Multi-way branching |
| `logic.merge` | Merge multiple inputs |
| `logic.filter` | Filter array data |
| `logic.sort` | Sort array data |
| `logic.limit` | Limit/paginate data |
| `logic.aggregate` | Sum, count, avg, etc. |
| `logic.loop` | Iterate over items |
| `logic.wait` | Pause execution |
| `logic.noop` | No operation (pass-through) |

### Integrations (10)
| Type | Description |
|------|-------------|
| `integration.openai` | OpenAI API |
| `integration.anthropic` | Anthropic Claude API |
| `integration.aws_s3` | AWS S3 operations |
| `integration.google_drive` | Google Drive operations |
| `integration.slack` | Slack messages |
| `integration.discord` | Discord messages |
| `integration.telegram` | Telegram messages |
| `integration.twilio` | SMS/calls |
| `integration.postgres` | PostgreSQL queries |
| `integration.mysql` | MySQL queries |
| `integration.mongodb` | MongoDB operations |
| `integration.redis` | Redis commands |
| `integration.salesforce` | Salesforce CRM |
| `integration.hubspot` | HubSpot CRM |
| `integration.airtable` | Airtable bases |
| `integration.notion` | Notion pages |
| `integration.stripe` | Stripe payments |

## Code Patterns

### Domain Entity Pattern
```go
// internal/core/domain/workflow/workflow.go
package workflow

type Workflow struct {
    id          WorkflowID
    workspaceID WorkspaceID
    name        string
    status      Status
    nodes       []Node
    version     int
}

func NewWorkflow(id WorkflowID, wsID WorkspaceID, name string) *Workflow {
    return &Workflow{
        id:          id,
        workspaceID: wsID,
        name:        name,
        status:      StatusDraft,
        version:     1,
    }
}

func (w *Workflow) Activate() error {
    if len(w.nodes) == 0 {
        return ErrEmptyWorkflow
    }
    w.status = StatusActive
    return nil
}
```

### Repository Interface Pattern
```go
// internal/core/domain/workflow/repository.go
package workflow

type Repository interface {
    Save(ctx context.Context, workflow *Workflow) error
    FindByID(ctx context.Context, id WorkflowID) (*Workflow, error)
    FindByWorkspace(ctx context.Context, wsID WorkspaceID, opts ListOptions) ([]*Workflow, int64, error)
    Delete(ctx context.Context, id WorkflowID) error
}
```

### Command Handler Pattern (CQRS)
```go
// internal/core/application/command/workflow/create_workflow.go
package workflow

type CreateWorkflowCommand struct {
    WorkspaceID uuid.UUID
    Name        string
    Nodes       []workflow.Node
}

type CreateWorkflowHandler struct {
    workflowRepo workflow.Repository
}

func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd CreateWorkflowCommand) (*workflow.Workflow, error) {
    wf := workflow.NewWorkflow(
        workflow.WorkflowID(uuid.New()),
        workflow.WorkspaceID(cmd.WorkspaceID),
        cmd.Name,
    )
    
    if err := h.workflowRepo.Save(ctx, wf); err != nil {
        return nil, err
    }
    
    return wf, nil
}
```

### HTTP Handler Pattern
```go
// internal/adapters/http/handlers/workflow/create.go
package workflow

func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    // 2. Create command
    cmd := command.CreateWorkflowCommand{
        WorkspaceID: getWorkspaceID(r),
        Name:        req.Name,
    }
    
    // 3. Execute
    result, err := h.handler.Handle(r.Context(), cmd)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 4. Return response
    json.NewEncoder(w).Encode(result)
}
```

### Node Implementation Pattern
```go
// internal/adapters/worker/nodes/logic/if.go
package logic

type IfNode struct{}

func (n *IfNode) Execute(ctx context.Context, input NodeInput) (*NodeOutput, error) {
    condition := input.Parameters["condition"].(string)
    
    result, err := evaluateCondition(condition, input.Data)
    if err != nil {
        return nil, err
    }
    
    return &NodeOutput{
        Data:   input.Data,
        Branch: result,
    }, nil
}

func (n *IfNode) Metadata() NodeMetadata {
    return NodeMetadata{
        Type:        "logic.if",
        Name:        "If",
        Description: "Conditional branching",
        Category:    "logic",
    }
}
```

## Development Setup

### Prerequisites
```bash
go version   # 1.23+
docker -v    # 20+
docker compose version  # 2.0+
```

### Quick Start
```bash
# 1. Start infrastructure
docker compose up -d postgres redis

# 2. Configure
cp configs/config.yaml.example configs/config.yaml

# 3. Run services (3 terminals)
go run cmd/api/main.go        # API on :8080
go run cmd/worker/main.go     # Worker
go run cmd/scheduler/main.go  # Scheduler on :8091 (health)
```

### Build Commands
```bash
go build ./...                    # Build all
go build -o bin/api cmd/api       # Build API
go build -o bin/worker cmd/worker # Build worker
go build -o bin/scheduler cmd/scheduler # Build scheduler

go test ./...                     # Run tests
```

## Common Tasks

### Adding a New Domain Aggregate

1. Create directory: `internal/core/domain/{aggregate}/`
2. Add files: `entity.go`, `repository.go`, `errors.go`
3. Define entity with business methods
4. Define repository interface

### Adding a New Use Case

1. Create command: `internal/core/application/command/{aggregate}/`
2. Or query: `internal/core/application/query/{aggregate}/`
3. Define Command/Query struct
4. Implement Handler with Handle method

### Adding a New API Endpoint

1. Create handler: `internal/adapters/http/handlers/{resource}/`
2. Add route in `internal/adapters/http/routes/routes.go`
3. Wire dependencies in `cmd/api/wire.go`

### Adding a New Node Type

1. Create node: `internal/adapters/worker/nodes/{category}/`
2. Implement `Node` interface (Execute, Metadata)
3. Register in `internal/adapters/worker/nodes/loader.go`

### Adding a New Repository Implementation

1. Create in `internal/adapters/persistence/postgres/repositories/`
2. Implement domain repository interface
3. Add mappers in `mappers/` if needed

## Security Considerations

1. **JWT Tokens**: 15-minute expiry, refresh tokens rotated
2. **Passwords**: bcrypt hashed (cost 10)
3. **Credentials**: AES-256-GCM encrypted at rest
4. **Rate Limiting**: Per-user, per-workspace, per-endpoint
5. **RBAC**: Role-based access control middleware
6. **Input Validation**: All requests validated

## Deployment

### Production Checklist
- [ ] Set `APP_ENVIRONMENT=production`
- [ ] Use strong `JWT_SECRET` (32+ characters)
- [ ] Enable SSL for database
- [ ] Configure CORS origins
- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure log aggregation
- [ ] Set up database backups
- [ ] Deploy multiple scheduler instances (leader election enabled)
- [ ] Scale workers based on load

### Scheduler High Availability
The scheduler uses Redis-based leader election. Deploy multiple instances:
```yaml
# Only the leader polls and dispatches
# Other instances remain on standby
# Health endpoints show leader status
```
</coding_guidelines>
