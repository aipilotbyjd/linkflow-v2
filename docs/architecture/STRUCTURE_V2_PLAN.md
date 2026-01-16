# LinkFlow v2 - Project Structure Plan

> **Version:** 2.0  
> **Date:** 2026-01-16  
> **Status:** Proposed  
> **Author:** Architecture Team

## 📋 Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Design Philosophy](#design-philosophy)
4. [Complete Directory Structure](#complete-directory-structure)
5. [Layer-by-Layer Breakdown](#layer-by-layer-breakdown)
6. [Code Organization Patterns](#code-organization-patterns)
7. [Dependency Rules](#dependency-rules)
8. [Naming Conventions](#naming-conventions)
9. [Migration Strategy](#migration-strategy)
10. [Code Examples](#code-examples)
11. [Testing Strategy](#testing-strategy)
12. [Benefits & Trade-offs](#benefits--trade-offs)
13. [FAQ](#faq)

---

## Executive Summary

### Problem Statement

LinkFlow's current structure has grown organically and now faces several challenges:

- **God file**: `server.go` contains 678 lines
- **Flat directories**: 38+ files in single directories
- **Scattered domain logic**: Models, repositories, and services in separate locations
- **Tight coupling**: Difficult to test and modify
- **Unclear boundaries**: No clear separation of concerns
- **Hard to scale**: Cannot easily extract microservices

### Solution Overview

Adopt **Clean Architecture** with **Domain-Driven Design** principles:

1. **Core Layer** - Business logic (domain + application services)
2. **Adapters Layer** - Infrastructure implementations
3. **Infrastructure Layer** - Shared utilities
4. **Shared Layer** - Common types and interfaces

### Key Metrics

| Metric | Current | Target | Improvement |
|--------|---------|--------|-------------|
| Average file size | 250 lines | 150 lines | 40% reduction |
| Files per directory | 38 | 10 | 74% reduction |
| Import depth | 5-6 levels | 3-4 levels | Cleaner |
| Test coverage | 40% | 80% | 100% increase |
| Onboarding time | 2 weeks | 3 days | 5x faster |

---

## Current State Analysis

### Codebase Statistics

```
Total Go files: 289
Total size: 2.3 MB

By module:
- Worker:     780 KB  (34%)
- Domain:     564 KB  (24%)
- API:        540 KB  (23%)
- PKG:        364 KB  (16%)
- Scheduler:   72 KB  (3%)
```

### Current Issues

#### 1. God Objects
- `internal/api/server.go` - 678 lines
- Handles all route definitions
- 40+ handler initializations
- Violates SRP

#### 2. Flat Structures
```
internal/api/handlers/     (38 files)
internal/domain/models/    (26 files)
internal/domain/services/  (34 files)
```

#### 3. Ambiguous Naming
```
internal/pkg/  ← Inside internal but named pkg?
```

#### 4. Scattered Domain Logic
```
User logic is in 3 places:
- internal/domain/models/user.go
- internal/domain/repositories/user.go
- internal/domain/services/user.go
```

#### 5. Duplicate Concerns
```
internal/pkg/cache/execution.go
internal/worker/cache/credentials.go  ← Duplicate cache concept
```

---

## Design Philosophy

### Guiding Principles

#### 1. Clean Architecture (Uncle Bob)

**Layers flow inward:**
```
External → Adapters → Application → Domain
        (Dependency Inversion)
```

**Rules:**
- Inner layers know nothing about outer layers
- Domain layer has ZERO framework dependencies
- All dependencies point inward

#### 2. Domain-Driven Design (Eric Evans)

**Strategic Design:**
- Organize by business domain (User, Workflow, Execution)
- Each domain has clear boundaries
- Ubiquitous language in code

**Tactical Design:**
- Aggregates, Entities, Value Objects
- Domain Events
- Repository pattern

#### 3. Hexagonal Architecture (Ports & Adapters)

**Ports** (interfaces) defined in domain:
```go
type WorkflowRepository interface {
    Save(ctx context.Context, wf *Workflow) error
}
```

**Adapters** (implementations) in infrastructure:
```go
type PostgresWorkflowRepository struct {
    db *gorm.DB
}
```

#### 4. CQRS (Command Query Responsibility Segregation)

**Separate reads and writes:**
- Commands (write): `CreateWorkflow`, `UpdateWorkflow`
- Queries (read): `GetWorkflow`, `ListWorkflows`

Benefits:
- Optimized for different access patterns
- Easier to scale reads vs writes
- Clear intent in code

#### 5. SOLID Principles

- **S**ingle Responsibility
- **O**pen/Closed
- **L**iskov Substitution
- **I**nterface Segregation
- **D**ependency Inversion

### Architecture Diagram

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
│                   (Use Cases / Services)                     │
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

---

## Complete Directory Structure

```
linkflow-v2/
│
├── cmd/                                    # 🚀 Application Entrypoints
│   ├── api/
│   │   ├── main.go                         # API server entry
│   │   └── wire.go                         # Dependency injection config
│   ├── worker/
│   │   ├── main.go                         # Background worker entry
│   │   └── wire.go
│   ├── scheduler/
│   │   ├── main.go                         # Cron scheduler entry
│   │   └── wire.go
│   └── migrate/
│       └── main.go                         # Database migration CLI
│
├── internal/                               # 🔒 Private Application Code
│   │
│   ├── core/                               # 💎 CORE BUSINESS LOGIC
│   │   │                                   # (Framework-agnostic, pure Go)
│   │   │
│   │   ├── domain/                         # Domain Models & Business Rules
│   │   │   │
│   │   │   ├── user/                       # User Aggregate
│   │   │   │   ├── user.go                 # User entity (aggregate root)
│   │   │   │   ├── session.go             # Session entity
│   │   │   │   ├── email.go               # Email value object
│   │   │   │   ├── password.go            # Password value object
│   │   │   │   ├── repository.go          # Repository interface (port)
│   │   │   │   ├── events.go              # Domain events (UserRegistered, etc.)
│   │   │   │   └── errors.go              # Domain errors (ErrUserNotFound)
│   │   │   │
│   │   │   ├── workspace/                  # Workspace Aggregate
│   │   │   │   ├── workspace.go           # Workspace entity (aggregate root)
│   │   │   │   ├── member.go              # Member entity
│   │   │   │   ├── invitation.go          # Invitation entity
│   │   │   │   ├── role.go                # Role value object
│   │   │   │   ├── repository.go          # Repository interface
│   │   │   │   ├── member_repository.go   # Member repository interface
│   │   │   │   ├── events.go              # WorkspaceCreated, MemberInvited
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── workflow/                   # Workflow Aggregate
│   │   │   │   ├── workflow.go            # Workflow entity (aggregate root)
│   │   │   │   ├── version.go             # Version entity
│   │   │   │   ├── node.go                # Node value object
│   │   │   │   ├── connection.go          # Connection value object
│   │   │   │   ├── status.go              # Status value object
│   │   │   │   ├── settings.go            # Settings value object
│   │   │   │   ├── repository.go          # Workflow repository interface
│   │   │   │   ├── version_repository.go  # Version repository interface
│   │   │   │   ├── service.go             # Domain service (if needed)
│   │   │   │   ├── events.go              # WorkflowCreated, Activated
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── execution/                  # Execution Aggregate
│   │   │   │   ├── execution.go           # Execution entity (aggregate root)
│   │   │   │   ├── node_execution.go      # NodeExecution entity
│   │   │   │   ├── status.go              # Execution status value object
│   │   │   │   ├── result.go              # Result value object
│   │   │   │   ├── repository.go          # Repository interface
│   │   │   │   ├── executor.go            # Executor service interface (port)
│   │   │   │   ├── events.go              # ExecutionStarted, Completed
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── credential/                 # Credential Aggregate
│   │   │   │   ├── credential.go          # Credential entity
│   │   │   │   ├── encrypted_data.go      # EncryptedData value object
│   │   │   │   ├── credential_type.go     # Type value object
│   │   │   │   ├── repository.go          # Repository interface
│   │   │   │   ├── encryption_service.go  # Encryption service interface (port)
│   │   │   │   ├── events.go              # CredentialCreated, Rotated
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── schedule/                   # Schedule Aggregate
│   │   │   │   ├── schedule.go            # Schedule entity
│   │   │   │   ├── cron_expression.go     # Cron value object
│   │   │   │   ├── repository.go          # Repository interface
│   │   │   │   ├── events.go              # ScheduleCreated, Triggered
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── webhook/                    # Webhook Aggregate
│   │   │   │   ├── endpoint.go            # WebhookEndpoint entity
│   │   │   │   ├── event.go               # WebhookEvent entity
│   │   │   │   ├── secret.go              # Secret value object
│   │   │   │   ├── repository.go          # Repository interface
│   │   │   │   ├── events.go              # WebhookTriggered
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   ├── billing/                    # Billing Aggregate
│   │   │   │   ├── subscription.go        # Subscription entity
│   │   │   │   ├── plan.go                # Plan value object
│   │   │   │   ├── usage.go               # Usage entity
│   │   │   │   ├── invoice.go             # Invoice entity
│   │   │   │   ├── repository.go          # Repository interface
│   │   │   │   ├── events.go              # SubscriptionCreated, PaymentFailed
│   │   │   │   └── errors.go
│   │   │   │
│   │   │   └── template/                   # Template Aggregate
│   │   │       ├── template.go
│   │   │       ├── category.go
│   │   │       ├── repository.go
│   │   │       ├── events.go
│   │   │       └── errors.go
│   │   │
│   │   └── application/                    # Application Services (Use Cases)
│   │       │                               # Orchestrate domain objects
│   │       │
│   │       ├── command/                    # Write Operations (CQRS)
│   │       │   │
│   │       │   ├── user/
│   │       │   │   ├── register_user.go
│   │       │   │   ├── update_profile.go
│   │       │   │   ├── change_password.go
│   │       │   │   ├── setup_mfa.go
│   │       │   │   ├── verify_mfa.go
│   │       │   │   └── disable_mfa.go
│   │       │   │
│   │       │   ├── workspace/
│   │       │   │   ├── create_workspace.go
│   │       │   │   ├── update_workspace.go
│   │       │   │   ├── delete_workspace.go
│   │       │   │   ├── invite_member.go
│   │       │   │   ├── remove_member.go
│   │       │   │   └── update_member_role.go
│   │       │   │
│   │       │   ├── workflow/
│   │       │   │   ├── create_workflow.go
│   │       │   │   ├── update_workflow.go
│   │       │   │   ├── delete_workflow.go
│   │       │   │   ├── clone_workflow.go
│   │       │   │   ├── activate_workflow.go
│   │       │   │   ├── deactivate_workflow.go
│   │       │   │   └── rollback_version.go
│   │       │   │
│   │       │   ├── execution/
│   │       │   │   ├── start_execution.go
│   │       │   │   ├── cancel_execution.go
│   │       │   │   ├── retry_execution.go
│   │       │   │   └── replay_execution.go
│   │       │   │
│   │       │   ├── credential/
│   │       │   │   ├── create_credential.go
│   │       │   │   ├── update_credential.go
│   │       │   │   ├── delete_credential.go
│   │       │   │   ├── share_credential.go
│   │       │   │   └── test_credential.go
│   │       │   │
│   │       │   ├── schedule/
│   │       │   │   ├── create_schedule.go
│   │       │   │   ├── update_schedule.go
│   │       │   │   ├── pause_schedule.go
│   │       │   │   └── resume_schedule.go
│   │       │   │
│   │       │   └── webhook/
│   │       │       ├── create_endpoint.go
│   │       │       ├── regenerate_secret.go
│   │       │       └── trigger_webhook.go
│   │       │
│   │       └── query/                      # Read Operations (CQRS)
│   │           │
│   │           ├── user/
│   │           │   ├── get_user.go
│   │           │   ├── get_current_user.go
│   │           │   └── list_users.go
│   │           │
│   │           ├── workspace/
│   │           │   ├── get_workspace.go
│   │           │   ├── list_workspaces.go
│   │           │   ├── list_members.go
│   │           │   └── get_workspace_stats.go
│   │           │
│   │           ├── workflow/
│   │           │   ├── get_workflow.go
│   │           │   ├── list_workflows.go
│   │           │   ├── search_workflows.go
│   │           │   ├── get_workflow_versions.go
│   │           │   └── compare_versions.go
│   │           │
│   │           ├── execution/
│   │           │   ├── get_execution.go
│   │           │   ├── list_executions.go
│   │           │   ├── search_executions.go
│   │           │   ├── get_execution_stats.go
│   │           │   └── get_node_executions.go
│   │           │
│   │           ├── credential/
│   │           │   ├── get_credential.go
│   │           │   ├── list_credentials.go
│   │           │   └── get_credential_shares.go
│   │           │
│   │           ├── schedule/
│   │           │   ├── get_schedule.go
│   │           │   └── list_schedules.go
│   │           │
│   │           └── analytics/
│   │               ├── get_workspace_analytics.go
│   │               ├── get_workflow_analytics.go
│   │               └── get_execution_metrics.go
│   │
│   ├── adapters/                           # 🔌 ADAPTERS LAYER
│   │   │                                   # Infrastructure Implementations
│   │   │
│   │   ├── persistence/                    # Database Adapters
│   │   │   │
│   │   │   ├── postgres/
│   │   │   │   ├── client.go              # Database client setup
│   │   │   │   ├── transaction.go         # Transaction wrapper
│   │   │   │   │
│   │   │   │   ├── migrations/            # SQL migrations
│   │   │   │   │   └── .gitkeep
│   │   │   │   │
│   │   │   │   ├── models/                # Database models (DB schema)
│   │   │   │   │   ├── user.go
│   │   │   │   │   ├── workspace.go
│   │   │   │   │   ├── workflow.go
│   │   │   │   │   └── execution.go
│   │   │   │   │
│   │   │   │   ├── repositories/          # Repository implementations
│   │   │   │   │   ├── user_repository.go
│   │   │   │   │   ├── workspace_repository.go
│   │   │   │   │   ├── workflow_repository.go
│   │   │   │   │   ├── execution_repository.go
│   │   │   │   │   ├── credential_repository.go
│   │   │   │   │   ├── schedule_repository.go
│   │   │   │   │   └── webhook_repository.go
│   │   │   │   │
│   │   │   │   └── mappers/               # Domain ↔ DB mapping
│   │   │   │       ├── user_mapper.go
│   │   │   │       ├── workspace_mapper.go
│   │   │   │       ├── workflow_mapper.go
│   │   │   │       └── execution_mapper.go
│   │   │   │
│   │   │   └── redis/
│   │   │       ├── client.go
│   │   │       ├── session_store.go
│   │   │       ├── cache_store.go
│   │   │       └── lock.go                # Distributed lock
│   │   │
│   │   ├── messaging/                      # Message Queue Adapters
│   │   │   │
│   │   │   ├── asynq/
│   │   │   │   ├── client.go              # Queue client
│   │   │   │   ├── server.go              # Queue server
│   │   │   │   └── tasks/                 # Task handlers
│   │   │   │       ├── workflow_execution.go
│   │   │   │       ├── send_email.go
│   │   │   │       ├── token_refresh.go
│   │   │   │       └── cleanup.go
│   │   │   │
│   │   │   └── streams/
│   │   │       ├── webhook_stream.go
│   │   │       ├── webhook_consumer.go
│   │   │       └── event_stream.go
│   │   │
│   │   ├── http/                           # HTTP Adapters (REST API)
│   │   │   │
│   │   │   ├── server.go                   # HTTP server setup (~100 lines)
│   │   │   ├── router.go                   # Main router
│   │   │   │
│   │   │   ├── routes/                     # Route definitions
│   │   │   │   ├── routes.go              # Main route registration
│   │   │   │   ├── auth_routes.go
│   │   │   │   ├── workspace_routes.go
│   │   │   │   ├── workflow_routes.go
│   │   │   │   ├── execution_routes.go
│   │   │   │   ├── credential_routes.go
│   │   │   │   ├── schedule_routes.go
│   │   │   │   ├── webhook_routes.go
│   │   │   │   ├── billing_routes.go
│   │   │   │   └── admin_routes.go
│   │   │   │
│   │   │   ├── handlers/                   # HTTP handlers (thin layer)
│   │   │   │   │                           # Call application layer use cases
│   │   │   │   ├── auth/
│   │   │   │   │   ├── register.go
│   │   │   │   │   ├── login.go
│   │   │   │   │   ├── refresh_token.go
│   │   │   │   │   ├── logout.go
│   │   │   │   │   ├── forgot_password.go
│   │   │   │   │   ├── reset_password.go
│   │   │   │   │   ├── oauth_redirect.go
│   │   │   │   │   ├── oauth_callback.go
│   │   │   │   │   ├── setup_mfa.go
│   │   │   │   │   ├── verify_mfa.go
│   │   │   │   │   └── disable_mfa.go
│   │   │   │   │
│   │   │   │   ├── workspace/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── list_members.go
│   │   │   │   │   ├── invite_member.go
│   │   │   │   │   ├── update_member.go
│   │   │   │   │   └── remove_member.go
│   │   │   │   │
│   │   │   │   ├── workflow/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── execute.go
│   │   │   │   │   ├── clone.go
│   │   │   │   │   ├── activate.go
│   │   │   │   │   ├── deactivate.go
│   │   │   │   │   ├── list_versions.go
│   │   │   │   │   ├── get_version.go
│   │   │   │   │   ├── rollback_version.go
│   │   │   │   │   ├── compare_versions.go
│   │   │   │   │   ├── export.go
│   │   │   │   │   └── import.go
│   │   │   │   │
│   │   │   │   ├── execution/
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   ├── stats.go
│   │   │   │   │   ├── cancel.go
│   │   │   │   │   ├── retry.go
│   │   │   │   │   ├── get_nodes.go
│   │   │   │   │   └── bulk_delete.go
│   │   │   │   │
│   │   │   │   ├── credential/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── test.go
│   │   │   │   │   ├── share.go
│   │   │   │   │   └── list_shares.go
│   │   │   │   │
│   │   │   │   ├── schedule/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── pause.go
│   │   │   │   │   └── resume.go
│   │   │   │   │
│   │   │   │   ├── webhook/
│   │   │   │   │   ├── trigger.go
│   │   │   │   │   ├── create_endpoint.go
│   │   │   │   │   ├── list_endpoints.go
│   │   │   │   │   ├── regenerate_secret.go
│   │   │   │   │   ├── activate.go
│   │   │   │   │   └── deactivate.go
│   │   │   │   │
│   │   │   │   ├── billing/
│   │   │   │   │   ├── get_plans.go
│   │   │   │   │   ├── get_subscription.go
│   │   │   │   │   ├── create_subscription.go
│   │   │   │   │   ├── cancel_subscription.go
│   │   │   │   │   ├── get_usage.go
│   │   │   │   │   ├── get_invoices.go
│   │   │   │   │   └── stripe_webhook.go
│   │   │   │   │
│   │   │   │   ├── analytics/
│   │   │   │   │   ├── workspace_analytics.go
│   │   │   │   │   └── workflow_analytics.go
│   │   │   │   │
│   │   │   │   ├── health/
│   │   │   │   │   ├── health.go
│   │   │   │   │   ├── live.go
│   │   │   │   │   └── ready.go
│   │   │   │   │
│   │   │   │   └── admin/
│   │   │   │       ├── stream_stats.go
│   │   │   │       └── metrics.go
│   │   │   │
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go                # JWT authentication
│   │   │   │   ├── apikey.go              # API key authentication
│   │   │   │   ├── tenant.go              # Workspace context
│   │   │   │   ├── ratelimit.go           # Rate limiting
│   │   │   │   ├── logger.go              # Request logging
│   │   │   │   ├── recovery.go            # Panic recovery
│   │   │   │   ├── cors.go                # CORS handling
│   │   │   │   ├── metrics.go             # Prometheus metrics
│   │   │   │   ├── rbac.go                # Role-based access control
│   │   │   │   └── idempotency.go         # Idempotency handling
│   │   │   │
│   │   │   └── dto/                        # Request/Response DTOs
│   │   │       ├── auth/
│   │   │       │   ├── register_request.go
│   │   │       │   ├── login_request.go
│   │   │       │   ├── login_response.go
│   │   │       │   └── token_response.go
│   │   │       ├── workspace/
│   │   │       │   ├── create_request.go
│   │   │       │   ├── workspace_response.go
│   │   │       │   └── member_response.go
│   │   │       ├── workflow/
│   │   │       │   ├── create_request.go
│   │   │       │   ├── update_request.go
│   │   │       │   └── workflow_response.go
│   │   │       ├── execution/
│   │   │       │   └── execution_response.go
│   │   │       └── common/
│   │   │           ├── response.go
│   │   │           ├── error_response.go
│   │   │           ├── pagination.go
│   │   │           └── validation.go
│   │   │
│   │   ├── websocket/                      # WebSocket Adapter
│   │   │   ├── hub.go                      # Connection hub
│   │   │   ├── client.go                   # WebSocket client
│   │   │   ├── subscriber.go               # Redis subscriber
│   │   │   ├── events.go                   # Event types
│   │   │   └── handler.go                  # Connection handler
│   │   │
│   │   ├── worker/                         # Background Worker Adapter
│   │   │   │
│   │   │   ├── server.go                   # Worker server
│   │   │   ├── worker.go                   # Worker implementation
│   │   │   │
│   │   │   ├── executor/                   # Execution Engine
│   │   │   │   ├── executor.go            # Main execution orchestrator
│   │   │   │   ├── processor.go           # Node processor
│   │   │   │   ├── runtime.go             # Runtime context
│   │   │   │   └── error_handler.go       # Error workflow handler
│   │   │   │
│   │   │   ├── nodes/                     # Node Implementations
│   │   │   │   │
│   │   │   │   ├── registry.go            # Node registry
│   │   │   │   ├── interface.go           # Node interface
│   │   │   │   ├── loader.go              # Auto-registration
│   │   │   │   │
│   │   │   │   ├── triggers/
│   │   │   │   │   ├── manual.go
│   │   │   │   │   ├── webhook.go
│   │   │   │   │   └── schedule.go
│   │   │   │   │
│   │   │   │   ├── actions/
│   │   │   │   │   ├── http/
│   │   │   │   │   │   └── http_request.go
│   │   │   │   │   ├── email/
│   │   │   │   │   │   └── send_email.go
│   │   │   │   │   ├── code/
│   │   │   │   │   │   ├── javascript.go
│   │   │   │   │   │   └── python.go
│   │   │   │   │   └── transform/
│   │   │   │   │       ├── set.go
│   │   │   │   │       └── respond.go
│   │   │   │   │
│   │   │   │   ├── logic/
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
│   │   │   │   └── integrations/
│   │   │   │       │
│   │   │   │       ├── cloud/
│   │   │   │       │   ├── aws_s3.go
│   │   │   │       │   └── google_drive.go
│   │   │   │       │
│   │   │   │       ├── communication/
│   │   │   │       │   ├── slack.go
│   │   │   │       │   ├── discord.go
│   │   │   │       │   ├── telegram.go
│   │   │   │       │   └── twilio.go
│   │   │   │       │
│   │   │   │       ├── database/
│   │   │   │       │   ├── postgres.go
│   │   │   │       │   ├── mysql.go
│   │   │   │       │   ├── mongodb.go
│   │   │   │       │   └── redis.go
│   │   │   │       │
│   │   │   │       ├── ai/
│   │   │   │       │   ├── openai.go
│   │   │   │       │   └── anthropic.go
│   │   │   │       │
│   │   │   │       ├── crm/
│   │   │   │       │   ├── salesforce.go
│   │   │   │       │   ├── hubspot.go
│   │   │   │       │   ├── airtable.go
│   │   │   │       │   └── notion.go
│   │   │   │       │
│   │   │   │       ├── payment/
│   │   │   │       │   └── stripe.go
│   │   │   │       │
│   │   │   │       ├── productivity/
│   │   │   │       │   ├── jira.go
│   │   │   │       │   ├── github.go
│   │   │   │       │   └── google_sheets.go
│   │   │   │       │
│   │   │   │       └── email/
│   │   │   │           ├── sendgrid.go
│   │   │   │           └── smtp.go
│   │   │   │
│   │   │   ├── middleware/                # Worker middleware
│   │   │   │   ├── logging.go
│   │   │   │   ├── tracing.go
│   │   │   │   ├── recovery.go
│   │   │   │   └── retry.go
│   │   │   │
│   │   │   └── cache/                     # Worker-specific caching
│   │   │       ├── credential_cache.go
│   │   │       └── result_cache.go
│   │   │
│   │   └── scheduler/                      # Scheduler Adapter
│   │       ├── server.go                   # Scheduler server
│   │       ├── poller.go                   # Schedule poller
│   │       ├── dispatcher.go               # Job dispatcher
│   │       ├── leader_election.go          # Leader election (HA)
│   │       ├── cron.go                     # Cron parser
│   │       └── metrics.go                  # Scheduler metrics
│   │
│   ├── infrastructure/                     # 🛠️ SHARED INFRASTRUCTURE
│   │   │                                   # Framework-agnostic utilities
│   │   │
│   │   ├── auth/
│   │   │   ├── jwt/
│   │   │   │   ├── manager.go
│   │   │   │   ├── claims.go
│   │   │   │   └── blacklist.go
│   │   │   └── oauth/
│   │   │       ├── manager.go
│   │   │       ├── provider.go
│   │   │       └── providers/
│   │   │           ├── google.go
│   │   │           ├── github.go
│   │   │           └── microsoft.go
│   │   │
│   │   ├── crypto/
│   │   │   ├── encryption.go               # AES-256-GCM
│   │   │   ├── hashing.go                  # bcrypt
│   │   │   ├── otp.go                      # TOTP for MFA
│   │   │   └── signing.go                  # Digital signatures
│   │   │
│   │   ├── email/
│   │   │   ├── service.go
│   │   │   ├── template.go
│   │   │   ├── templates/
│   │   │   │   ├── welcome.html
│   │   │   │   ├── reset_password.html
│   │   │   │   ├── invitation.html
│   │   │   │   └── execution_failed.html
│   │   │   └── providers/
│   │   │       ├── smtp.go
│   │   │       └── sendgrid.go
│   │   │
│   │   ├── storage/
│   │   │   ├── storage.go                  # Storage interface
│   │   │   ├── s3/
│   │   │   │   └── client.go
│   │   │   └── local/
│   │   │       └── filesystem.go
│   │   │
│   │   ├── cache/
│   │   │   ├── cache.go                    # Cache interface
│   │   │   ├── redis_cache.go
│   │   │   ├── memory_cache.go
│   │   │   └── noop_cache.go
│   │   │
│   │   ├── observability/
│   │   │   │
│   │   │   ├── logger/
│   │   │   │   ├── logger.go              # Logger interface
│   │   │   │   ├── zerolog.go             # Zerolog implementation
│   │   │   │   └── noop.go
│   │   │   │
│   │   │   ├── metrics/
│   │   │   │   ├── metrics.go             # Metrics interface
│   │   │   │   ├── prometheus.go
│   │   │   │   ├── collector.go
│   │   │   │   └── noop.go
│   │   │   │
│   │   │   └── tracing/
│   │   │       ├── tracer.go              # Tracer interface
│   │   │       ├── opentelemetry.go
│   │   │       ├── jaeger.go
│   │   │       └── noop.go
│   │   │
│   │   ├── resilience/
│   │   │   ├── circuitbreaker/
│   │   │   │   ├── breaker.go
│   │   │   │   └── manager.go
│   │   │   ├── retry/
│   │   │   │   ├── retry.go
│   │   │   │   └── backoff.go
│   │   │   └── ratelimit/
│   │   │       ├── limiter.go
│   │   │       └── redis_limiter.go
│   │   │
│   │   ├── validation/
│   │   │   ├── validator.go
│   │   │   └── rules/
│   │   │       ├── email.go
│   │   │       ├── password.go
│   │   │       └── cron.go
│   │   │
│   │   └── config/
│   │       ├── config.go
│   │       ├── loader.go
│   │       ├── validation.go
│   │       └── types.go
│   │
│   └── shared/                             # 📦 SHARED KERNEL
│       │                                   # Types/interfaces used across layers
│       │
│       ├── types/
│       │   ├── id.go                       # UUID wrapper
│       │   ├── pagination.go               # Pagination types
│       │   ├── filter.go                   # Filter types
│       │   └── json.go                     # JSON types
│       │
│       ├── errors/
│       │   ├── errors.go                   # Error types
│       │   ├── codes.go                    # Error codes
│       │   └── chain.go                    # Error wrapping
│       │
│       └── events/
│           ├── event.go                    # Base event interface
│           ├── bus.go                      # Event bus interface
│           └── publisher.go                # Event publisher
│
├── pkg/                                    # 📤 PUBLIC PACKAGES
│   │                                       # Can be imported by other projects
│   ├── httpclient/
│   │   ├── client.go
│   │   └── pool.go
│   ├── timeutil/
│   │   └── time.go
│   ├── sliceutil/
│   │   └── slice.go
│   └── stringutil/
│       └── string.go
│
├── test/                                   # 🧪 TEST INFRASTRUCTURE
│   │
│   ├── integration/                        # Integration tests
│   │   ├── api/
│   │   │   ├── auth_test.go
│   │   │   ├── workflow_test.go
│   │   │   └── execution_test.go
│   │   ├── worker/
│   │   │   └── execution_test.go
│   │   └── scheduler/
│   │       └── schedule_test.go
│   │
│   ├── e2e/                                # End-to-end tests
│   │   ├── workflow_execution_test.go
│   │   ├── webhook_trigger_test.go
│   │   └── scheduled_execution_test.go
│   │
│   ├── fixtures/                           # Test data
│   │   ├── users.json
│   │   ├── workspaces.json
│   │   ├── workflows.json
│   │   ├── credentials.json
│   │   └── executions.json
│   │
│   ├── mocks/                              # Generated mocks (mockery)
│   │   ├── domain/
│   │   │   ├── WorkflowRepository.go
│   │   │   └── ExecutionRepository.go
│   │   └── infrastructure/
│   │       └── Cache.go
│   │
│   └── testutil/                           # Test utilities
│       ├── database.go                     # Test DB setup
│       ├── redis.go                        # Test Redis setup
│       ├── server.go                       # Test HTTP server
│       ├── assertions.go                   # Custom assertions
│       └── factories.go                    # Test data factories
│
├── migrations/                             # 🗄️ DATABASE MIGRATIONS
│   └── postgres/
│       ├── 000001_initial_schema.up.sql
│       ├── 000001_initial_schema.down.sql
│       ├── 000002_add_webhooks.up.sql
│       ├── 000002_add_webhooks.down.sql
│       ├── 000003_add_billing.up.sql
│       └── 000003_add_billing.down.sql
│
├── scripts/                                # 📜 AUTOMATION SCRIPTS
│   ├── build.sh                            # Build all services
│   ├── build-api.sh                        # Build API only
│   ├── build-worker.sh                     # Build worker only
│   ├── deploy.sh                           # Deployment script
│   ├── seed-dev.sh                         # Seed dev data
│   ├── seed-prod.sh                        # Seed prod data
│   ├── test.sh                             # Run all tests
│   ├── test-integration.sh                 # Run integration tests
│   ├── test-e2e.sh                         # Run e2e tests
│   ├── generate-mocks.sh                   # Generate mocks with mockery
│   ├── migrate.sh                          # Run migrations
│   ├── rollback.sh                         # Rollback migrations
│   └── lint.sh                             # Run linters
│
├── configs/                                # ⚙️ CONFIGURATION
│   ├── config.yaml                         # Default config
│   ├── config.dev.yaml                     # Development config
│   ├── config.staging.yaml                 # Staging config
│   ├── config.prod.yaml                    # Production config
│   └── .golangci.yml                       # Linter config
│
├── docs/                                   # 📚 DOCUMENTATION
│   ├── README.md                           # Docs overview
│   │
│   ├── getting-started/
│   │   ├── quickstart.md
│   │   ├── installation.md
│   │   └── first-workflow.md
│   │
│   ├── architecture/
│   │   ├── overview.md
│   │   ├── clean-architecture.md
│   │   ├── domain-driven-design.md
│   │   ├── data-flow.md
│   │   └── STRUCTURE_V2_PLAN.md          # This document
│   │
│   ├── api/
│   │   ├── README.md
│   │   ├── authentication.md
│   │   ├── workspaces.md
│   │   ├── workflows.md
│   │   ├── executions.md
│   │   ├── credentials.md
│   │   ├── schedules.md
│   │   ├── webhooks.md
│   │   └── billing.md
│   │
│   ├── development/
│   │   ├── setup.md
│   │   ├── project-structure.md
│   │   ├── coding-standards.md
│   │   ├── testing.md
│   │   ├── contributing.md
│   │   └── code-review.md
│   │
│   ├── deployment/
│   │   ├── docker.md
│   │   ├── kubernetes.md
│   │   ├── aws.md
│   │   ├── production.md
│   │   └── scaling.md
│   │
│   └── operations/
│       ├── monitoring.md
│       ├── alerting.md
│       ├── troubleshooting.md
│       ├── runbooks.md
│       └── disaster-recovery.md
│
├── deploy/                                 # 🚀 DEPLOYMENT
│   ├── docker/
│   │   ├── Dockerfile.api
│   │   ├── Dockerfile.worker
│   │   ├── Dockerfile.scheduler
│   │   └── .dockerignore
│   │
│   ├── kubernetes/
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secrets.yaml
│   │   ├── api/
│   │   │   ├── deployment.yaml
│   │   │   ├── service.yaml
│   │   │   ├── hpa.yaml
│   │   │   └── ingress.yaml
│   │   ├── worker/
│   │   │   ├── deployment.yaml
│   │   │   └── hpa.yaml
│   │   ├── scheduler/
│   │   │   └── deployment.yaml
│   │   ├── postgres/
│   │   │   ├── statefulset.yaml
│   │   │   ├── service.yaml
│   │   │   └── pvc.yaml
│   │   └── redis/
│   │       ├── statefulset.yaml
│   │       ├── service.yaml
│   │       └── pvc.yaml
│   │
│   ├── docker-compose.yml                  # Production compose
│   ├── docker-compose.dev.yml              # Development compose
│   └── README.md
│
├── .github/                                # CI/CD
│   ├── workflows/
│   │   ├── ci.yml                          # Continuous Integration
│   │   ├── test.yml                        # Test workflow
│   │   ├── build.yml                       # Build workflow
│   │   ├── deploy-staging.yml              # Deploy to staging
│   │   └── deploy-production.yml           # Deploy to production
│   │
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   │
│   └── PULL_REQUEST_TEMPLATE.md
│
├── .air.toml                               # Hot reload config
├── .editorconfig                           # Editor config
├── .gitignore
├── .env.example                            # Environment variables template
├── go.mod
├── go.sum
├── Makefile                                # Build commands
├── README.md                               # Project README
└── LICENSE
```

---

## Layer-by-Layer Breakdown

### 1. Core Layer (`internal/core/`)

**Purpose**: Pure business logic, framework-agnostic

#### Domain Layer (`internal/core/domain/`)

**What goes here:**
- Entities (aggregate roots)
- Value objects
- Repository interfaces (ports)
- Domain services
- Domain events
- Business validation rules

**What DOESN'T go here:**
- HTTP handling
- Database code
- External API calls
- Framework-specific code

**Example structure for User aggregate:**
```
internal/core/domain/user/
├── user.go              # Aggregate root
├── session.go           # Entity
├── email.go             # Value object
├── password.go          # Value object
├── repository.go        # Repository interface (port)
├── events.go            # Domain events
└── errors.go            # Domain-specific errors
```

#### Application Layer (`internal/core/application/`)

**Purpose**: Use cases / Application services

**CQRS Pattern:**
- `command/` - Write operations
- `query/` - Read operations

**What goes here:**
- Orchestration of domain objects
- Transaction boundaries
- Application-specific validation
- Event publishing

**What DOESN'T go here:**
- Domain logic (goes in domain)
- HTTP handling (goes in adapters)
- Direct database queries (use repositories)

---

### 2. Adapters Layer (`internal/adapters/`)

**Purpose**: Implementations of interfaces defined in domain

#### Persistence Adapters (`internal/adapters/persistence/`)

**What goes here:**
- Repository implementations
- Database models (separate from domain models)
- Mappers (DB ↔ Domain)
- Database-specific logic

**Key principle**: Domain models ≠ Database models

```go
// Database model (adapters/persistence/postgres/models/user.go)
type UserDB struct {
    ID           string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
}

// Domain model (core/domain/user/user.go)
type User struct {
    id       UserID
    email    Email
    password Password
}
```

#### HTTP Adapters (`internal/adapters/http/`)

**What goes here:**
- HTTP handlers (thin layer)
- Route definitions
- DTOs (Request/Response objects)
- Middleware

**Handler pattern:**
```go
func (h *CreateWorkflowHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request → DTO
    // 2. Convert DTO → Command
    // 3. Call use case
    // 4. Convert result → Response DTO
    // 5. Write response
}
```

---

### 3. Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Shared utilities (no business logic)

**What goes here:**
- Email service
- Crypto utilities
- Logger
- Metrics
- Configuration
- Cache
- Storage

**Key principle**: No business rules

---

### 4. Shared Layer (`internal/shared/`)

**Purpose**: Types/interfaces used across all layers

**What goes here:**
- Common types (ID, Pagination)
- Base error types
- Event interfaces

---

## Code Organization Patterns

### 1. Aggregate Organization

Each aggregate is self-contained:

```
workflow/
├── workflow.go           # Aggregate root + methods
├── version.go           # Entity
├── node.go              # Value object
├── connection.go        # Value object
├── repository.go        # Repository interface
├── events.go            # Domain events
└── errors.go            # Domain errors
```

**Why?**
- High cohesion
- Clear boundaries
- Easy to understand
- Easy to test
- Easy to extract to microservice

### 2. CQRS Organization

**Commands** (writes):
```
application/command/workflow/
├── create_workflow.go
├── update_workflow.go
├── delete_workflow.go
└── activate_workflow.go
```

Each file has:
```go
// Command (input)
type CreateWorkflowCommand struct {
    WorkspaceID uuid.UUID
    Name        string
    Nodes       []Node
}

// Handler
type CreateWorkflowHandler struct {
    workflowRepo domain.WorkflowRepository
    eventBus     events.Bus
}

// Handle method
func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd CreateWorkflowCommand) (*Workflow, error) {
    // Business logic here
}
```

**Queries** (reads):
```
application/query/workflow/
├── get_workflow.go
├── list_workflows.go
└── search_workflows.go
```

**Benefits:**
- Clear separation of reads/writes
- Easy to optimize separately
- Easy to test

### 3. Handler Organization

**Old way (bad):**
```
handlers/workflow.go  (1000+ lines)
```

**New way (good):**
```
handlers/workflow/
├── create.go        (~50 lines each)
├── update.go
├── get.go
├── list.go
└── execute.go
```

### 4. Integration Nodes Organization

**Grouped by category:**
```
nodes/integrations/
├── cloud/           # AWS S3, Google Drive
├── communication/   # Slack, Discord, Email
├── database/        # Postgres, MySQL, MongoDB
├── ai/              # OpenAI, Anthropic
├── crm/             # Salesforce, HubSpot
└── payment/         # Stripe
```

**Why?**
- Easy to find
- Clear categorization
- Easy to add new integrations

---

## Dependency Rules

### The Dependency Rule

**Layers can only depend on layers below them:**

```
Adapters → Application → Domain
Infrastructure → (Any)
```

**NEVER:**
- Domain → Application
- Domain → Adapters
- Application → Adapters (except interfaces)

### Import Rules

**✅ Allowed:**
```go
// Handler imports application
import "internal/core/application/command/workflow"

// Application imports domain
import "internal/core/domain/workflow"

// Adapter imports domain (for interfaces)
import "internal/core/domain/workflow"
```

**❌ NOT Allowed:**
```go
// Domain imports adapter
import "internal/adapters/postgres"  // WRONG!

// Domain imports HTTP
import "internal/adapters/http"  // WRONG!

// Domain imports infrastructure
import "internal/infrastructure/crypto"  // WRONG!
```

**How to solve:**
```go
// Define interface in domain
type EncryptionService interface {
    Encrypt(data []byte) ([]byte, error)
}

// Implement in infrastructure
type AESEncryption struct { ... }

// Inject via dependency injection
```

### Package Import Guidelines

**By layer:**

| Layer | Can Import |
|-------|------------|
| **Domain** | Only other domain packages, shared |
| **Application** | Domain, shared |
| **Adapters** | Application, domain, infrastructure, shared |
| **Infrastructure** | Shared only |
| **Shared** | Nothing (base layer) |

---

## Naming Conventions

### File Naming

**Use snake_case for files:**
```
create_workflow.go        ✅
createWorkflow.go         ❌
CreateWorkflow.go         ❌
```

**Name files by their primary type/function:**
```
user.go                  # Contains User type
user_repository.go       # Contains UserRepository interface
create_workflow.go       # Contains CreateWorkflow command
```

### Package Naming

**Use singular, lowercase:**
```
user/         ✅
users/        ❌
User/         ❌
```

**Exceptions:**
- `handlers` (plural okay)
- `mappers` (plural okay)
- `types` (plural okay)

### Type Naming

**Aggregates:** Noun
```go
type Workflow struct { ... }
type User struct { ... }
```

**Commands:** Verb + Noun
```go
type CreateWorkflowCommand struct { ... }
type UpdateUserCommand struct { ... }
```

**Queries:** Get/List + Noun + Query
```go
type GetWorkflowQuery struct { ... }
type ListExecutionsQuery struct { ... }
```

**Handlers:** Command/Query + Handler
```go
type CreateWorkflowHandler struct { ... }
type GetWorkflowHandler struct { ... }
```

**Repositories:** Noun + Repository
```go
type WorkflowRepository interface { ... }
```

**Events:** Past tense
```go
type WorkflowCreated struct { ... }
type ExecutionCompleted struct { ... }
```

### Method Naming

**Commands:**
```go
func (h *Handler) Handle(ctx context.Context, cmd Command) error
```

**Queries:**
```go
func (h *Handler) Handle(ctx context.Context, q Query) (Result, error)
```

**Repository:**
```go
func (r *Repo) Save(ctx context.Context, entity *Entity) error
func (r *Repo) FindByID(ctx context.Context, id ID) (*Entity, error)
func (r *Repo) Delete(ctx context.Context, id ID) error
```

---

## Migration Strategy

### Phase-by-Phase Approach

**Don't do a big-bang rewrite!** Migrate incrementally.

### Phase 1: Foundation (Week 1)

**Goal**: Create new structure, move utilities

**Steps:**
1. Create directory structure
2. Move shared utilities
3. Set up test infrastructure
4. Update documentation

**Commands:**
```bash
# Create core directories
mkdir -p internal/core/{domain,application/{command,query}}

# Create adapters directories
mkdir -p internal/adapters/{http,persistence,messaging,worker,websocket,scheduler}

# Create infrastructure
mkdir -p internal/infrastructure/{auth,crypto,email,storage,cache,observability,resilience,validation,config}

# Create shared
mkdir -p internal/shared/{types,errors,events}

# Create test infrastructure
mkdir -p test/{integration,e2e,fixtures,mocks,testutil}

# Create other directories
mkdir -p migrations/postgres scripts
```

**Move utilities:**
```bash
# Move config
mv internal/pkg/config internal/infrastructure/config

# Move crypto
mv internal/pkg/crypto internal/infrastructure/crypto

# Move logger
mv internal/pkg/logger internal/infrastructure/observability/logger

# Move metrics
mv internal/pkg/metrics internal/infrastructure/observability/metrics

# Move validation
mv internal/pkg/validator internal/infrastructure/validation

# Move email
mv internal/pkg/email internal/infrastructure/email
```

### Phase 2: Domain Layer (Week 2)

**Goal**: Extract domain models

**Start with simplest aggregate:**

**Step 1: User aggregate**
```bash
mkdir -p internal/core/domain/user

# Create files
touch internal/core/domain/user/user.go
touch internal/core/domain/user/session.go
touch internal/core/domain/user/email.go
touch internal/core/domain/user/password.go
touch internal/core/domain/user/repository.go
touch internal/core/domain/user/events.go
touch internal/core/domain/user/errors.go
```

**Extract from existing:**
```go
// From: internal/domain/models/user.go
// To: internal/core/domain/user/user.go

package user

type User struct {
    id        UserID
    email     Email
    password  Password
    firstName string
    lastName  string
    createdAt time.Time
}

// Business methods
func (u *User) ChangePassword(old, new string) error {
    // Business logic here
}
```

**Repository interface:**
```go
// internal/core/domain/user/repository.go
package user

type Repository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id UserID) (*User, error)
    FindByEmail(ctx context.Context, email Email) (*User, error)
    Delete(ctx context.Context, id UserID) error
}
```

**Repeat for:** Workspace, Workflow, Execution, Credential

### Phase 3: Application Layer (Week 3)

**Goal**: Extract use cases from services

**Example: Create Workflow**

```go
// internal/core/application/command/workflow/create_workflow.go
package workflow

type CreateWorkflowCommand struct {
    WorkspaceID uuid.UUID
    Name        string
    Description string
    Nodes       []Node
    Connections []Connection
    CreatedBy   uuid.UUID
}

type CreateWorkflowHandler struct {
    workflowRepo domain.WorkflowRepository
    eventBus     events.Bus
}

func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd CreateWorkflowCommand) (*domain.Workflow, error) {
    // 1. Create workflow aggregate
    workflow := domain.NewWorkflow(
        domain.WorkflowID(uuid.New()),
        domain.WorkspaceID(cmd.WorkspaceID),
        cmd.Name,
        cmd.Description,
    )
    
    // 2. Add nodes
    for _, node := range cmd.Nodes {
        workflow.AddNode(node)
    }
    
    // 3. Save
    if err := h.workflowRepo.Save(ctx, workflow); err != nil {
        return nil, err
    }
    
    // 4. Publish event
    h.eventBus.Publish(workflow.Events()...)
    
    return workflow, nil
}
```

**Migrate services one by one:**
- `WorkflowService.Create()` → `command/workflow/create_workflow.go`
- `WorkflowService.GetByID()` → `query/workflow/get_workflow.go`
- etc.

### Phase 4: Adapters Layer (Week 4)

**Goal**: Move infrastructure implementations

**Step 1: Split server.go**

```bash
mkdir -p internal/adapters/http/routes

# Create route files
touch internal/adapters/http/routes/routes.go
touch internal/adapters/http/routes/auth_routes.go
touch internal/adapters/http/routes/workspace_routes.go
touch internal/adapters/http/routes/workflow_routes.go
```

**Example:**
```go
// internal/adapters/http/routes/workflow_routes.go
package routes

func RegisterWorkflowRoutes(r chi.Router, handlers *WorkflowHandlers) {
    r.Post("/workflows", handlers.Create)
    r.Get("/workflows", handlers.List)
    r.Get("/workflows/{id}", handlers.Get)
    r.Put("/workflows/{id}", handlers.Update)
    r.Delete("/workflows/{id}", handlers.Delete)
    r.Post("/workflows/{id}/execute", handlers.Execute)
}
```

**Step 2: Organize handlers**

```bash
mkdir -p internal/adapters/http/handlers/workflow

# Move and split
# From: internal/api/handlers/workflow.go
# To: internal/adapters/http/handlers/workflow/create.go
#     internal/adapters/http/handlers/workflow/get.go
#     internal/adapters/http/handlers/workflow/list.go
#     etc.
```

**Step 3: Move repositories**

```bash
mkdir -p internal/adapters/persistence/postgres/repositories

# Implement repository interfaces
```

```go
// internal/adapters/persistence/postgres/repositories/workflow_repository.go
package repositories

type PostgresWorkflowRepository struct {
    db *gorm.DB
}

// Implement domain.WorkflowRepository interface
func (r *PostgresWorkflowRepository) Save(ctx context.Context, wf *domain.Workflow) error {
    // Map domain → DB
    dbModel := mappers.WorkflowToDB(wf)
    
    // Save
    return r.db.WithContext(ctx).Save(dbModel).Error
}
```

### Phase 5: Testing & Documentation (Week 5)

**Goal**: Add tests and update docs

**Steps:**
1. Write unit tests for domain
2. Write integration tests
3. Update documentation
4. Add migration guide
5. Team training

---

## Code Examples

### Example 1: Domain Aggregate

```go
// internal/core/domain/workflow/workflow.go
package workflow

import (
    "time"
    "github.com/google/uuid"
)

// Workflow is an aggregate root
type Workflow struct {
    id          WorkflowID
    workspaceID WorkspaceID
    name        string
    description string
    status      Status
    nodes       []Node
    connections []Connection
    version     int
    createdAt   time.Time
    updatedAt   time.Time
    events      []Event
}

// NewWorkflow creates a new workflow
func NewWorkflow(id WorkflowID, workspaceID WorkspaceID, name, description string) *Workflow {
    wf := &Workflow{
        id:          id,
        workspaceID: workspaceID,
        name:        name,
        description: description,
        status:      StatusDraft,
        nodes:       make([]Node, 0),
        connections: make([]Connection, 0),
        version:     1,
        createdAt:   time.Now(),
        updatedAt:   time.Now(),
        events:      make([]Event, 0),
    }
    
    // Record domain event
    wf.recordEvent(WorkflowCreated{
        WorkflowID:  wf.id,
        WorkspaceID: wf.workspaceID,
        Name:        wf.name,
        CreatedAt:   wf.createdAt,
    })
    
    return wf
}

// Activate activates the workflow
func (w *Workflow) Activate() error {
    // Business rule: Cannot activate empty workflow
    if len(w.nodes) == 0 {
        return ErrEmptyWorkflow
    }
    
    // Business rule: Must have trigger node
    if !w.hasTriggerNode() {
        return ErrNoTriggerNode
    }
    
    w.status = StatusActive
    w.updatedAt = time.Now()
    
    // Record event
    w.recordEvent(WorkflowActivated{
        WorkflowID: w.id,
        ActivatedAt: time.Now(),
    })
    
    return nil
}

// AddNode adds a node to the workflow
func (w *Workflow) AddNode(node Node) error {
    // Business rule: Node ID must be unique
    if w.hasNode(node.ID()) {
        return ErrDuplicateNodeID
    }
    
    w.nodes = append(w.nodes, node)
    w.updatedAt = time.Now()
    w.version++
    
    return nil
}

// Private methods
func (w *Workflow) hasTriggerNode() bool {
    for _, node := range w.nodes {
        if node.Type().IsTrigger() {
            return true
        }
    }
    return false
}

func (w *Workflow) hasNode(id string) bool {
    for _, node := range w.nodes {
        if node.ID() == id {
            return true
        }
    }
    return false
}

func (w *Workflow) recordEvent(event Event) {
    w.events = append(w.events, event)
}

// Getters
func (w *Workflow) ID() WorkflowID { return w.id }
func (w *Workflow) WorkspaceID() WorkspaceID { return w.workspaceID }
func (w *Workflow) Name() string { return w.name }
func (w *Workflow) Status() Status { return w.status }
func (w *Workflow) Nodes() []Node { return w.nodes }
func (w *Workflow) Events() []Event { return w.events }
```

### Example 2: Repository Interface

```go
// internal/core/domain/workflow/repository.go
package workflow

import (
    "context"
)

// Repository defines the interface for workflow persistence
type Repository interface {
    // Save persists a workflow
    Save(ctx context.Context, workflow *Workflow) error
    
    // FindByID retrieves a workflow by ID
    FindByID(ctx context.Context, id WorkflowID) (*Workflow, error)
    
    // FindByWorkspace retrieves workflows for a workspace
    FindByWorkspace(ctx context.Context, workspaceID WorkspaceID, opts ListOptions) ([]*Workflow, int64, error)
    
    // Delete removes a workflow
    Delete(ctx context.Context, id WorkflowID) error
    
    // ExistsByName checks if a workflow with given name exists
    ExistsByName(ctx context.Context, workspaceID WorkspaceID, name string) (bool, error)
}

// ListOptions defines options for listing workflows
type ListOptions struct {
    Page     int
    PageSize int
    Status   *Status
    Search   string
    SortBy   string
    SortDesc bool
}
```

### Example 3: Application Command Handler

```go
// internal/core/application/command/workflow/create_workflow.go
package workflow

import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
    "github.com/linkflow/internal/core/domain/workflow"
    "github.com/linkflow/internal/shared/events"
)

// CreateWorkflowCommand represents the command to create a workflow
type CreateWorkflowCommand struct {
    WorkspaceID uuid.UUID
    Name        string
    Description string
    Nodes       []workflow.Node
    Connections []workflow.Connection
    CreatedBy   uuid.UUID
}

// CreateWorkflowHandler handles workflow creation
type CreateWorkflowHandler struct {
    workflowRepo workflow.Repository
    eventBus     events.Bus
    logger       Logger
}

// NewCreateWorkflowHandler creates a new handler
func NewCreateWorkflowHandler(
    workflowRepo workflow.Repository,
    eventBus events.Bus,
    logger Logger,
) *CreateWorkflowHandler {
    return &CreateWorkflowHandler{
        workflowRepo: workflowRepo,
        eventBus:     eventBus,
        logger:       logger,
    }
}

// Handle executes the command
func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd CreateWorkflowCommand) (*workflow.Workflow, error) {
    // Validation
    if cmd.Name == "" {
        return nil, workflow.ErrNameRequired
    }
    
    // Check uniqueness
    exists, err := h.workflowRepo.ExistsByName(ctx, workflow.WorkspaceID(cmd.WorkspaceID), cmd.Name)
    if err != nil {
        return nil, fmt.Errorf("failed to check workflow existence: %w", err)
    }
    if exists {
        return nil, workflow.ErrDuplicateName
    }
    
    // Create workflow aggregate
    wf := workflow.NewWorkflow(
        workflow.WorkflowID(uuid.New()),
        workflow.WorkspaceID(cmd.WorkspaceID),
        cmd.Name,
        cmd.Description,
    )
    
    // Add nodes
    for _, node := range cmd.Nodes {
        if err := wf.AddNode(node); err != nil {
            return nil, fmt.Errorf("failed to add node: %w", err)
        }
    }
    
    // Add connections
    for _, conn := range cmd.Connections {
        if err := wf.AddConnection(conn); err != nil {
            return nil, fmt.Errorf("failed to add connection: %w", err)
        }
    }
    
    // Save
    if err := h.workflowRepo.Save(ctx, wf); err != nil {
        h.logger.Error("failed to save workflow", "error", err)
        return nil, fmt.Errorf("failed to save workflow: %w", err)
    }
    
    // Publish domain events
    for _, event := range wf.Events() {
        if err := h.eventBus.Publish(ctx, event); err != nil {
            h.logger.Warn("failed to publish event", "event", event, "error", err)
        }
    }
    
    h.logger.Info("workflow created", "workflow_id", wf.ID())
    
    return wf, nil
}
```

### Example 4: Repository Implementation

```go
// internal/adapters/persistence/postgres/repositories/workflow_repository.go
package repositories

import (
    "context"
    "fmt"
    
    "gorm.io/gorm"
    "github.com/linkflow/internal/core/domain/workflow"
    "github.com/linkflow/internal/adapters/persistence/postgres/models"
    "github.com/linkflow/internal/adapters/persistence/postgres/mappers"
)

type PostgresWorkflowRepository struct {
    db *gorm.DB
}

func NewPostgresWorkflowRepository(db *gorm.DB) *PostgresWorkflowRepository {
    return &PostgresWorkflowRepository{db: db}
}

// Save implements workflow.Repository
func (r *PostgresWorkflowRepository) Save(ctx context.Context, wf *workflow.Workflow) error {
    // Start transaction
    tx := r.db.WithContext(ctx).Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // Map domain → DB
    dbModel := mappers.WorkflowToDB(wf)
    
    // Save workflow
    if err := tx.Save(&dbModel).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to save workflow: %w", err)
    }
    
    // Save version
    versionModel := mappers.WorkflowVersionToD B(wf)
    if err := tx.Create(&versionModel).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to save workflow version: %w", err)
    }
    
    // Commit
    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return nil
}

// FindByID implements workflow.Repository
func (r *PostgresWorkflowRepository) FindByID(ctx context.Context, id workflow.WorkflowID) (*workflow.Workflow, error) {
    var dbModel models.WorkflowDB
    
    err := r.db.WithContext(ctx).
        Where("id = ?", id).
        First(&dbModel).Error
    
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, workflow.ErrNotFound
        }
        return nil, fmt.Errorf("failed to find workflow: %w", err)
    }
    
    // Map DB → Domain
    return mappers.WorkflowToDomain(&dbModel)
}
```

### Example 5: HTTP Handler

```go
// internal/adapters/http/handlers/workflow/create.go
package workflow

import (
    "encoding/json"
    "net/http"
    
    "github.com/linkflow/internal/core/application/command/workflow"
    "github.com/linkflow/internal/adapters/http/dto/workflow"
    "github.com/linkflow/internal/adapters/http/dto/common"
    "github.com/linkflow/internal/adapters/http/middleware"
)

type CreateHandler struct {
    createWorkflow *workflow.CreateWorkflowHandler
}

func NewCreateHandler(createWorkflow *workflow.CreateWorkflowHandler) *CreateHandler {
    return &CreateHandler{
        createWorkflow: createWorkflow,
    }
}

func (h *CreateHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Get context (user, workspace)
    claims := middleware.GetUserFromContext(r.Context())
    wsCtx := middleware.GetWorkspaceFromContext(r.Context())
    
    // 2. Parse request
    var req dto.CreateWorkflowRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        common.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    // 3. Validate request
    if err := req.Validate(); err != nil {
        common.ValidationErrorResponse(w, err)
        return
    }
    
    // 4. Convert DTO → Command
    cmd := workflow.CreateWorkflowCommand{
        WorkspaceID: wsCtx.WorkspaceID,
        Name:        req.Name,
        Description: req.Description,
        Nodes:       req.Nodes,
        Connections: req.Connections,
        CreatedBy:   claims.UserID,
    }
    
    // 5. Execute command
    wf, err := h.createWorkflow.Handle(r.Context(), cmd)
    if err != nil {
        // Handle domain errors
        switch err {
        case workflow.ErrNameRequired:
            common.ErrorResponse(w, http.StatusBadRequest, "name is required")
        case workflow.ErrDuplicateName:
            common.ErrorResponse(w, http.StatusConflict, "workflow name already exists")
        default:
            common.ErrorResponse(w, http.StatusInternalServerError, "failed to create workflow")
        }
        return
    }
    
    // 6. Convert Domain → Response DTO
    response := dto.ToWorkflowResponse(wf)
    
    // 7. Write response
    common.Created(w, response)
}
```

---

## Testing Strategy

### 1. Unit Tests (Domain Layer)

**Test pure business logic:**

```go
// internal/core/domain/workflow/workflow_test.go
package workflow_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/linkflow/internal/core/domain/workflow"
)

func TestWorkflow_Activate(t *testing.T) {
    t.Run("activates workflow with valid nodes", func(t *testing.T) {
        // Arrange
        wf := workflow.NewWorkflow(
            workflow.WorkflowID("wf-1"),
            workflow.WorkspaceID("ws-1"),
            "Test Workflow",
            "Description",
        )
        wf.AddNode(workflow.NewManualTriggerNode("node-1"))
        
        // Act
        err := wf.Activate()
        
        // Assert
        assert.NoError(t, err)
        assert.Equal(t, workflow.StatusActive, wf.Status())
    })
    
    t.Run("fails to activate empty workflow", func(t *testing.T) {
        // Arrange
        wf := workflow.NewWorkflow(
            workflow.WorkflowID("wf-1"),
            workflow.WorkspaceID("ws-1"),
            "Test Workflow",
            "Description",
        )
        
        // Act
        err := wf.Activate()
        
        // Assert
        assert.ErrorIs(t, err, workflow.ErrEmptyWorkflow)
    })
}
```

### 2. Application Tests (Use Cases)

**Test use case logic with mocks:**

```go
// internal/core/application/command/workflow/create_workflow_test.go
package workflow_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "github.com/linkflow/internal/core/application/command/workflow"
    "github.com/linkflow/test/mocks"
)

func TestCreateWorkflowHandler_Handle(t *testing.T) {
    t.Run("creates workflow successfully", func(t *testing.T) {
        // Arrange
        mockRepo := new(mocks.WorkflowRepository)
        mockEventBus := new(mocks.EventBus)
        mockLogger := new(mocks.Logger)
        
        handler := workflow.NewCreateWorkflowHandler(mockRepo, mockEventBus, mockLogger)
        
        cmd := workflow.CreateWorkflowCommand{
            WorkspaceID: uuid.New(),
            Name:        "Test Workflow",
            Description: "Test Description",
        }
        
        // Mock expectations
        mockRepo.On("ExistsByName", mock.Anything, mock.Anything, "Test Workflow").Return(false, nil)
        mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
        mockEventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)
        mockLogger.On("Info", mock.Anything, mock.Anything).Return()
        
        // Act
        result, err := handler.Handle(context.Background(), cmd)
        
        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, "Test Workflow", result.Name())
        
        mockRepo.AssertExpectations(t)
        mockEventBus.AssertExpectations(t)
    })
}
```

### 3. Integration Tests

**Test adapters with real infrastructure:**

```go
// test/integration/api/workflow_test.go
package api_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/linkflow/test/testutil"
)

func TestCreateWorkflow_Integration(t *testing.T) {
    // Setup test server with real dependencies
    srv := testutil.NewTestServer(t)
    defer srv.Close()
    
    // Get auth token
    token := srv.Login(t, "test@example.com", "password")
    
    // Prepare request
    reqBody := map[string]interface{}{
        "name":        "Integration Test Workflow",
        "description": "Test",
        "nodes":       []interface{}{},
    }
    body, _ := json.Marshal(reqBody)
    
    // Make request
    req := httptest.NewRequest("POST", "/api/v1/workspaces/"+srv.WorkspaceID+"/workflows", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    
    rec := httptest.NewRecorder()
    srv.Router.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
    
    var response map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.Equal(t, "Integration Test Workflow", response["name"])
}
```

### 4. E2E Tests

**Test full workflows:**

```go
// test/e2e/workflow_execution_test.go
package e2e_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/linkflow/test/testutil"
)

func TestWorkflowExecution_E2E(t *testing.T) {
    // Setup
    client := testutil.NewAPIClient(t)
    
    // 1. Create workflow
    workflow := client.CreateWorkflow(t, map[string]interface{}{
        "name": "E2E Test Workflow",
        "nodes": []interface{}{
            map[string]interface{}{
                "type": "trigger.manual",
                "id":   "trigger-1",
            },
            map[string]interface{}{
                "type": "action.http",
                "id":   "action-1",
                "parameters": map[string]interface{}{
                    "url":    "https://api.example.com/data",
                    "method": "GET",
                },
            },
        },
        "connections": []interface{}{
            map[string]interface{}{
                "source": "trigger-1",
                "target": "action-1",
            },
        },
    })
    
    // 2. Activate workflow
    client.ActivateWorkflow(t, workflow.ID)
    
    // 3. Execute workflow
    execution := client.ExecuteWorkflow(t, workflow.ID, map[string]interface{}{})
    
    // 4. Wait for completion
    client.WaitForExecution(t, execution.ID, 30*time.Second)
    
    // 5. Assert results
    result := client.GetExecution(t, execution.ID)
    assert.Equal(t, "completed", result.Status)
    assert.NotEmpty(t, result.OutputData)
}
```

---

## Benefits & Trade-offs

### Benefits

#### 1. Maintainability
- **Smaller files**: Average 150 lines vs 250 lines
- **Clear organization**: Know where everything is
- **Easy navigation**: Related code together

#### 2. Testability
- **Unit test domain**: No dependencies
- **Mock interfaces**: Easy to test in isolation
- **Integration tests**: Test adapters separately

#### 3. Scalability
- **Microservice extraction**: Take domain + adapters
- **Team scalability**: Clear boundaries prevent conflicts
- **Performance**: Optimize layers independently

#### 4. Flexibility
- **Swap implementations**: Change DB without touching domain
- **Multiple adapters**: REST + GraphQL + gRPC
- **Framework independence**: Migrate frameworks easily

#### 5. Understanding
- **New developers**: Clear structure = faster onboarding
- **Documentation**: Structure = documentation
- **Code review**: Easier to review smaller files

### Trade-offs

#### 1. More Files
- **Current**: 289 files
- **New**: ~400-450 files
- **Mitigation**: Better organized, easier to find

#### 2. More Indirection
- **More layers**: Handler → Command → Domain → Repository
- **Mitigation**: Clear flow, better separation

#### 3. More Boilerplate
- **More interfaces**: Repository, Service interfaces
- **More mappers**: Domain ↔ DB ↔ DTO
- **Mitigation**: Code generation, templates

#### 4. Learning Curve
- **Clean Architecture**: Team needs training
- **DDD**: New concepts to learn
- **Mitigation**: Documentation, pair programming

### When NOT to Use This Structure

- **Very small projects** (< 10k LOC)
- **Prototypes** or POCs
- **Single-developer projects**
- **CRUD-only applications**

**For LinkFlow:** With 289 files and growing, this structure is appropriate.

---

## FAQ

### Q: Do we need to rewrite everything?

**A:** No! Migrate incrementally:
1. New features use new structure
2. Old code stays until touched
3. Refactor when modifying existing code

### Q: What about existing tests?

**A:** Keep them running:
1. Keep old tests passing
2. Add new tests for new structure
3. Migrate tests gradually

### Q: How do we handle the transition period?

**A:** Both structures coexist:
```
internal/
├── api/              ← Old (keep for now)
├── domain/           ← Old (keep for now)
├── worker/           ← Old (keep for now)
└── core/             ← New (add incrementally)
    └── adapters/     ← New (add incrementally)
```

### Q: What about dependency injection?

**A:** Use Wire:
```go
// cmd/api/wire.go
func InitializeServer() (*Server, error) {
    wire.Build(
        // Repositories
        postgres.NewWorkflowRepository,
        
        // Use cases
        workflow.NewCreateWorkflowHandler,
        
        // Handlers
        handlers.NewCreateWorkflowHandler,
        
        // Server
        NewServer,
    )
    return nil, nil
}
```

### Q: How do we handle database transactions?

**A:** At application layer:
```go
// Application layer handles transactions
func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd Command) error {
    return h.txManager.InTransaction(ctx, func(ctx context.Context) error {
        // Create workflow
        wf := domain.NewWorkflow(...)
        
        // Save
        if err := h.workflowRepo.Save(ctx, wf); err != nil {
            return err  // Rollback
        }
        
        // Save version
        if err := h.versionRepo.Save(ctx, version); err != nil {
            return err  // Rollback
        }
        
        return nil  // Commit
    })
}
```

### Q: What about shared types?

**A:** In `internal/shared/`:
- UUID wrappers
- Pagination
- Common errors
- Event interfaces

### Q: How do we version the API?

**A:**
```
internal/adapters/http/
├── v1/
│   ├── handlers/
│   └── routes/
└── v2/
    ├── handlers/
    └── routes/
```

### Q: What about feature flags?

**A:** In application layer:
```go
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    if h.features.IsEnabled("new-execution-engine") {
        return h.newExecutor.Execute(ctx, cmd)
    }
    return h.oldExecutor.Execute(ctx, cmd)
}
```

---

## Summary

### Key Takeaways

1. **Clean Architecture** - Clear layer boundaries
2. **Domain-Driven Design** - Organize by business domain
3. **CQRS** - Separate reads and writes
4. **Incremental migration** - Don't rewrite everything
5. **Test-driven** - Write tests as you migrate

### Next Steps

1. **Review this document** with the team
2. **Start Phase 1** (foundation)
3. **Migrate one aggregate** (e.g., User)
4. **Get feedback**
5. **Continue incrementally**

### Success Metrics

| Metric | Current | Target | Timeline |
|--------|---------|--------|----------|
| Average file size | 250 lines | 150 lines | 3 months |
| Files per directory | 38 | 10 | 3 months |
| Test coverage | 40% | 80% | 6 months |
| Onboarding time | 2 weeks | 3 days | 6 months |
| Build time | 2 min | 1 min | 3 months |

---

## Resources

### Books
- "Clean Architecture" by Robert C. Martin
- "Domain-Driven Design" by Eric Evans
- "Implementing Domain-Driven Design" by Vaughn Vernon

### Articles
- [The Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html)

### Go-Specific
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Clean Architecture Example](https://github.com/bxcodec/go-clean-arch)

---

**Document Version:** 2.0  
**Last Updated:** 2026-01-16  
**Next Review:** 2026-02-16
