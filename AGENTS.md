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

---

## CODING RULES (MUST FOLLOW)

### Rule 1: One Handler Per File

**ALWAYS create one handler per file.** Never consolidate multiple handlers into a single file.

```
handlers/
├── billing/
│   ├── types.go              # Shared types, interfaces, constants
│   ├── get_plans.go          # GetPlansHandler
│   ├── get_subscription.go   # GetSubscriptionHandler
│   ├── create_subscription.go
│   ├── cancel_subscription.go
│   ├── get_usage.go
│   ├── get_invoices.go
│   └── stripe_webhook.go
```

**Handler File Structure:**
```go
package billing

import (
    // imports
)

// Request/Response types specific to this handler (if not in types.go)
type GetPlansRequest struct { ... }

// Handler struct with dependencies
type GetPlansHandler struct {
    service BillingService
}

// Constructor
func NewGetPlansHandler(service BillingService) *GetPlansHandler {
    return &GetPlansHandler{service: service}
}

// Handle method
func (h *GetPlansHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

### Rule 2: Shared Types in types.go

Each handler package should have a `types.go` file containing:
- Shared structs (DTOs, entities)
- Interfaces
- Constants
- Helper functions used across multiple handlers

```go
// handlers/billing/types.go
package billing

type Plan struct { ... }
type Subscription struct { ... }
type BillingService interface { ... }
```

### Rule 3: Handler Naming Conventions

| Action | Handler Name | File Name |
|--------|--------------|-----------|
| List items | `ListHandler` | `list.go` |
| Get single | `GetHandler` | `get.go` |
| Create | `CreateHandler` | `create.go` |
| Update | `UpdateHandler` | `update.go` |
| Delete | `DeleteHandler` | `delete.go` |
| Custom action | `{Action}Handler` | `{action}.go` |

Examples:
- `GetPlansHandler` → `get_plans.go`
- `CreateSubscriptionHandler` → `create_subscription.go`
- `StripeWebhookHandler` → `stripe_webhook.go`

### Rule 4: Use Middleware Helpers

Use the provided middleware helper functions:
```go
// Get user ID from context
userID := middleware.GetUserID(r.Context())

// Get workspace ID from context
workspaceID := middleware.GetWorkspaceID(r.Context())

// Get full workspace context
wsCtx := middleware.GetWorkspaceFromContext(r.Context())

// Get user claims
claims := middleware.GetUserFromContext(r.Context())
```

### Rule 5: Error Handling

Use the common response helpers:
```go
common.Success(w, data)           // 200 OK
common.Created(w, data)           // 201 Created
common.BadRequest(w, "message")   // 400 Bad Request
common.Unauthorized(w, "message") // 401 Unauthorized
common.Forbidden(w, "message")    // 403 Forbidden
common.NotFound(w, "resource")    // 404 Not Found
common.HandleError(w, err)        // Maps domain errors to HTTP
```

### Rule 6: Always Check Error Returns

**NEVER ignore error returns.** If a function returns an error, handle it:

```go
// WRONG - ignored error
result, _ := someFunction()

// CORRECT
result, err := someFunction()
if err != nil {
    // handle error
}
```

### Rule 7: IPv6-Safe Network Code

Use `net.JoinHostPort()` instead of `fmt.Sprintf()` for addresses:

```go
// WRONG - breaks with IPv6
addr := fmt.Sprintf("%s:%d", host, port)

// CORRECT - works with IPv4 and IPv6
addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
```

### Rule 8: Repository Method Naming

Follow these naming conventions for repository methods:
- `Create(ctx, entity)` - Create new
- `Update(ctx, entity)` - Update existing
- `Delete(ctx, id)` - Delete by ID
- `FindByID(ctx, id)` - Find single by ID
- `FindByWorkspaceID(ctx, wsID, opts)` - Find by workspace
- `FindBy{Field}(ctx, value)` - Find by specific field
- `CountBy{Field}(ctx, value)` - Count by field
- `ExistsBy{Field}(ctx, value)` - Check existence

### Rule 9: Domain Model Field Access

Access domain entity fields correctly:
```go
// If field is a pointer, check for nil
if entity.CompletedAt != nil {
    duration := entity.CompletedAt.Sub(*entity.StartedAt)
}

// Use getter methods when available
workspaceID := entity.GetWorkspaceID()
```

### Rule 10: Run Verification Before Committing

Always run these checks before committing:
```bash
go build ./...    # Must pass
go vet ./...      # Must pass
go test ./...     # Fix any failures
```

---

## Project Structure (v2 Clean Architecture)

```
linkflow-v2/
│
├── cmd/                                    # Application Entrypoints
│   ├── api/main.go                         # API server entry
│   ├── worker/main.go                      # Background worker entry
│   ├── scheduler/main.go                   # Cron scheduler
│   └── migrate/main.go                     # Database migration CLI
│
├── internal/                               # Private Application Code
│   │
│   ├── core/                               # CORE BUSINESS LOGIC
│   │   ├── domain/                         # Domain Models & Business Rules
│   │   │   ├── user/
│   │   │   ├── workspace/
│   │   │   ├── workflow/
│   │   │   ├── execution/
│   │   │   ├── credential/
│   │   │   ├── schedule/
│   │   │   ├── webhook/
│   │   │   ├── billing/
│   │   │   ├── template/
│   │   │   └── folder/
│   │   │
│   │   └── application/                    # Application Services (CQRS)
│   │       ├── command/                    # Write Operations
│   │       └── query/                      # Read Operations
│   │
│   ├── adapters/                           # ADAPTERS LAYER
│   │   ├── http/                           # HTTP Adapters (REST API)
│   │   │   ├── routes/routes.go            # All route registration
│   │   │   ├── handlers/                   # HTTP handlers (ONE FILE PER HANDLER)
│   │   │   │   ├── admin/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── metrics.go
│   │   │   │   │   ├── stream_stats.go
│   │   │   │   │   ├── replay_dlq.go
│   │   │   │   │   └── trim_stream.go
│   │   │   │   ├── auth/
│   │   │   │   │   ├── register.go
│   │   │   │   │   ├── login.go
│   │   │   │   │   ├── refresh.go
│   │   │   │   │   ├── logout.go
│   │   │   │   │   ├── forgot_password.go
│   │   │   │   │   ├── reset_password.go
│   │   │   │   │   ├── setup_mfa.go
│   │   │   │   │   ├── verify_mfa.go
│   │   │   │   │   ├── disable_mfa.go
│   │   │   │   │   ├── oauth_redirect.go
│   │   │   │   │   └── oauth_callback.go
│   │   │   │   ├── billing/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── get_plans.go
│   │   │   │   │   ├── get_subscription.go
│   │   │   │   │   ├── create_subscription.go
│   │   │   │   │   ├── cancel_subscription.go
│   │   │   │   │   ├── get_usage.go
│   │   │   │   │   ├── get_invoices.go
│   │   │   │   │   └── stripe_webhook.go
│   │   │   │   ├── binarydata/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── upload.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── get_info.go
│   │   │   │   │   ├── download.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── stats.go
│   │   │   │   │   └── cleanup.go
│   │   │   │   ├── credential/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── test.go
│   │   │   │   │   └── refresh.go
│   │   │   │   ├── dashboard/
│   │   │   │   │   ├── dashboard.go
│   │   │   │   │   └── quick_stats.go
│   │   │   │   ├── execution/
│   │   │   │   │   ├── start.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── cancel.go
│   │   │   │   │   ├── retry.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   ├── bulk_delete.go
│   │   │   │   │   ├── replay.go
│   │   │   │   │   ├── nodes.go
│   │   │   │   │   ├── get_node.go
│   │   │   │   │   ├── stats.go
│   │   │   │   │   ├── waiting.go
│   │   │   │   │   └── resume.go
│   │   │   │   ├── folder/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── tree.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   └── delete.go
│   │   │   │   ├── health/
│   │   │   │   │   └── health.go
│   │   │   │   ├── marketplace/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── browse.go
│   │   │   │   │   ├── featured.go
│   │   │   │   │   ├── categories.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── use.go
│   │   │   │   │   ├── publish.go
│   │   │   │   │   ├── my_published.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── sync.go
│   │   │   │   │   ├── unpublish.go
│   │   │   │   │   ├── rate.go
│   │   │   │   │   ├── get_my_rating.go
│   │   │   │   │   ├── list_ratings.go
│   │   │   │   │   ├── rating_stats.go
│   │   │   │   │   └── delete_rating.go
│   │   │   │   ├── nodetypes/
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── categories.go
│   │   │   │   │   └── get.go
│   │   │   │   ├── oauth/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── list_providers.go
│   │   │   │   │   ├── authorize.go
│   │   │   │   │   └── callback.go
│   │   │   │   ├── pinneddata/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── get_all.go
│   │   │   │   │   ├── get_by_node.go
│   │   │   │   │   ├── set.go
│   │   │   │   │   └── delete.go
│   │   │   │   ├── schedule/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── pause.go
│   │   │   │   │   └── resume.go
│   │   │   │   ├── share/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── shared_by_me.go
│   │   │   │   │   ├── shared_with_me.go
│   │   │   │   │   ├── pending.go
│   │   │   │   │   ├── accept.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   └── revoke.go
│   │   │   │   ├── template/
│   │   │   │   │   ├── types.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── featured.go
│   │   │   │   │   ├── categories.go
│   │   │   │   │   ├── by_category.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   └── use.go
│   │   │   │   ├── user/
│   │   │   │   │   ├── get.go
│   │   │   │   │   └── update.go
│   │   │   │   ├── webhook/
│   │   │   │   │   ├── trigger.go
│   │   │   │   │   ├── create_endpoint.go
│   │   │   │   │   ├── list_endpoints.go
│   │   │   │   │   ├── regenerate_secret.go
│   │   │   │   │   ├── activate.go
│   │   │   │   │   └── deactivate.go
│   │   │   │   ├── workflow/
│   │   │   │   │   ├── create.go
│   │   │   │   │   ├── get.go
│   │   │   │   │   ├── list.go
│   │   │   │   │   ├── search.go
│   │   │   │   │   ├── update.go
│   │   │   │   │   ├── delete.go
│   │   │   │   │   ├── activate.go
│   │   │   │   │   ├── deactivate.go
│   │   │   │   │   ├── clone.go
│   │   │   │   │   ├── duplicate.go
│   │   │   │   │   ├── export.go
│   │   │   │   │   ├── import.go
│   │   │   │   │   ├── validate.go
│   │   │   │   │   ├── test_node.go
│   │   │   │   │   ├── versions.go
│   │   │   │   │   ├── get_version.go
│   │   │   │   │   ├── rollback.go
│   │   │   │   │   └── compare_versions.go
│   │   │   │   └── workspace/
│   │   │   │       ├── create.go
│   │   │   │       ├── get.go
│   │   │   │       ├── list.go
│   │   │   │       ├── update.go
│   │   │   │       ├── delete.go
│   │   │   │       ├── list_members.go
│   │   │   │       ├── invite_member.go
│   │   │   │       ├── update_member.go
│   │   │   │       └── remove_member.go
│   │   │   │
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go                 # JWT authentication + GetUserID()
│   │   │   │   ├── apikey.go               # API key authentication
│   │   │   │   ├── tenant.go               # Workspace context + GetWorkspaceID()
│   │   │   │   ├── ratelimit.go
│   │   │   │   ├── logging.go
│   │   │   │   ├── recovery.go
│   │   │   │   ├── cors.go
│   │   │   │   ├── metrics.go
│   │   │   │   ├── rbac.go
│   │   │   │   └── idempotency.go
│   │   │   │
│   │   │   └── dto/common/response.go
│   │   │
│   │   ├── persistence/                    # Database Adapters
│   │   │   ├── postgres/
│   │   │   │   ├── client.go
│   │   │   │   ├── models/
│   │   │   │   ├── repositories/
│   │   │   │   └── mappers/
│   │   │   └── redis/
│   │   │
│   │   ├── messaging/asynq/               # Message Queue
│   │   ├── websocket/                     # WebSocket Adapter
│   │   ├── worker/                        # Background Worker
│   │   │   ├── executor/
│   │   │   └── nodes/                     # 27 Node Types
│   │   └── scheduler/                     # Cron Scheduler
│   │
│   ├── infrastructure/                    # SHARED INFRASTRUCTURE
│   │   ├── auth/jwt/
│   │   ├── auth/oauth/
│   │   ├── crypto/
│   │   ├── email/
│   │   ├── storage/
│   │   ├── cache/
│   │   ├── observability/
│   │   ├── resilience/
│   │   ├── validation/
│   │   └── config/
│   │
│   └── shared/                            # SHARED KERNEL
│       ├── types/
│       ├── errors/
│       └── events/
│
├── configs/
├── deploy/
├── docs/
├── migrations/
├── pkg/
├── scripts/
├── go.mod
├── go.sum
├── Makefile
└── AGENTS.md
```

---

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

---

## API Endpoints Reference

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Invalidate session |
| POST | `/api/v1/auth/forgot-password` | Request password reset |
| POST | `/api/v1/auth/reset-password` | Reset password |
| POST | `/api/v1/auth/mfa/setup` | Setup MFA |
| POST | `/api/v1/auth/mfa/verify` | Verify MFA code |
| DELETE | `/api/v1/auth/mfa` | Disable MFA |
| GET | `/api/v1/auth/oauth/{provider}` | OAuth redirect |
| GET | `/api/v1/auth/oauth/{provider}/callback` | OAuth callback |

### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/me` | Get current user |
| PUT | `/api/v1/users/me` | Update current user |

### Workspaces
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces` | Create workspace |
| GET | `/api/v1/workspaces` | List workspaces |
| GET | `/api/v1/workspaces/{id}` | Get workspace |
| PUT | `/api/v1/workspaces/{id}` | Update workspace |
| DELETE | `/api/v1/workspaces/{id}` | Delete workspace |
| GET | `/api/v1/workspaces/{id}/members` | List members |
| POST | `/api/v1/workspaces/{id}/members` | Invite member |
| PUT | `/api/v1/workspaces/{id}/members/{mid}` | Update member |
| DELETE | `/api/v1/workspaces/{id}/members/{mid}` | Remove member |

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
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/clone` | Clone workflow |
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/duplicate` | Duplicate |
| GET | `/api/v1/workspaces/{id}/workflows/{wfId}/export` | Export JSON |
| POST | `/api/v1/workspaces/{id}/workflows/import` | Import JSON |
| POST | `/api/v1/workspaces/{id}/workflows/validate` | Validate |
| POST | `/api/v1/workspaces/{id}/workflows/test-node` | Test node |
| GET | `/api/v1/workspaces/{id}/workflows/{wfId}/versions` | List versions |
| GET | `/api/v1/workspaces/{id}/workflows/{wfId}/versions/{v}` | Get version |
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/versions/{v}/rollback` | Rollback |
| GET | `/api/v1/workspaces/{id}/workflows/{wfId}/compare-versions` | Compare |

### Executions
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces/{id}/workflows/{wfId}/execute` | Start execution |
| GET | `/api/v1/workspaces/{id}/executions` | List executions |
| GET | `/api/v1/workspaces/{id}/executions/search` | Search executions |
| GET | `/api/v1/workspaces/{id}/executions/stats` | Get stats |
| DELETE | `/api/v1/workspaces/{id}/executions/bulk` | Bulk delete |
| GET | `/api/v1/workspaces/{id}/executions/{exId}` | Get execution |
| POST | `/api/v1/workspaces/{id}/executions/{exId}/cancel` | Cancel |
| POST | `/api/v1/workspaces/{id}/executions/{exId}/retry` | Retry |
| POST | `/api/v1/workspaces/{id}/executions/{exId}/replay` | Replay |
| GET | `/api/v1/workspaces/{id}/executions/{exId}/nodes` | Get nodes |
| GET | `/api/v1/workspaces/{id}/waiting-executions` | List waiting |

### Credentials
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/workspaces/{id}/credentials` | Create credential |
| GET | `/api/v1/workspaces/{id}/credentials` | List credentials |
| GET | `/api/v1/workspaces/{id}/credentials/{credId}` | Get credential |
| PUT | `/api/v1/workspaces/{id}/credentials/{credId}` | Update credential |
| DELETE | `/api/v1/workspaces/{id}/credentials/{credId}` | Delete credential |
| POST | `/api/v1/workspaces/{id}/credentials/{credId}/test` | Test connection |
| POST | `/api/v1/workspaces/{id}/credentials/{credId}/refresh` | Refresh token |

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

### Billing
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/billing/plans` | Get available plans |
| GET | `/api/v1/workspaces/{id}/billing/subscription` | Get subscription |
| POST | `/api/v1/workspaces/{id}/billing/subscription` | Create subscription |
| DELETE | `/api/v1/workspaces/{id}/billing/subscription` | Cancel subscription |
| GET | `/api/v1/workspaces/{id}/billing/usage` | Get usage |
| GET | `/api/v1/workspaces/{id}/billing/invoices` | Get invoices |

### Templates & Marketplace
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/templates` | List templates |
| GET | `/api/v1/templates/featured` | Featured templates |
| GET | `/api/v1/templates/categories` | Categories |
| GET | `/api/v1/marketplace` | Browse marketplace |
| GET | `/api/v1/marketplace/featured` | Featured items |
| POST | `/api/v1/workspaces/{id}/marketplace` | Publish to marketplace |

### Health & Monitoring
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/health/live` | Liveness probe |
| GET | `/api/v1/health/ready` | Readiness probe |
| GET | `/metrics` | Prometheus metrics |

---

## Code Patterns

### HTTP Handler Pattern (REQUIRED STRUCTURE)
```go
// internal/adapters/http/handlers/billing/get_plans.go
package billing

import (
    "net/http"
    "github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// GetPlansHandler handles get billing plans request
type GetPlansHandler struct {
    service BillingService
}

// NewGetPlansHandler creates a new handler
func NewGetPlansHandler(service BillingService) *GetPlansHandler {
    return &GetPlansHandler{service: service}
}

// Handle handles the request
func (h *GetPlansHandler) Handle(w http.ResponseWriter, r *http.Request) {
    plans, err := h.service.GetPlans()
    if err != nil {
        common.HandleError(w, err)
        return
    }
    common.Success(w, map[string]interface{}{
        "plans": plans,
    })
}
```

### Domain Entity Pattern
```go
// internal/core/domain/workflow/workflow.go
package workflow

func NewWorkflow(workspaceID, createdBy uuid.UUID, name string) *Workflow {
    return &Workflow{
        ID:          uuid.New(),
        WorkspaceID: workspaceID,
        CreatedBy:   createdBy,
        Name:        name,
        Status:      StatusDraft,
        Version:     1,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}
```

### Repository Interface Pattern
```go
// internal/core/domain/workflow/repository.go
package workflow

type Repository interface {
    Create(ctx context.Context, workflow *Workflow) error
    Update(ctx context.Context, workflow *Workflow) error
    Delete(ctx context.Context, id uuid.UUID) error
    FindByID(ctx context.Context, id uuid.UUID) (*Workflow, error)
    FindByWorkspaceID(ctx context.Context, wsID uuid.UUID, opts *ListOptions) ([]Workflow, int64, error)
}
```

---

## Development Commands

```bash
# Build
make build              # Build all binaries
make build-api          # Build API only

# Development
make dev-infra          # Start Postgres + Redis in Docker
make dev-api            # Start infra + run API locally
make dev-worker         # Start infra + run Worker locally

# Docker
make docker-dev-up      # Start full stack in Docker
make docker-dev-down    # Stop Docker stack

# Testing
make test               # Run tests
make test-coverage      # Run with coverage
make lint               # Run golangci-lint

# Verification (run before commits)
go build ./...          # Must pass
go vet ./...            # Must pass
go test ./...           # Fix failures
```

---

## Security Considerations

1. **JWT Tokens**: 15-minute expiry, refresh tokens rotated
2. **Passwords**: bcrypt hashed (cost 10)
3. **Credentials**: AES-256-GCM encrypted at rest
4. **Rate Limiting**: Per-user, per-workspace, per-endpoint
5. **RBAC**: Role-based access control middleware
6. **Input Validation**: All requests validated
7. **Error Returns**: Always check and handle errors
8. **Secrets**: Never log sensitive data

---

## Deployment

### Production Checklist
- [ ] Set `APP_ENVIRONMENT=production`
- [ ] Use strong `JWT_SECRET` (32+ characters)
- [ ] Enable SSL for database
- [ ] Configure CORS origins
- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure log aggregation
- [ ] Set up database backups
- [ ] Deploy multiple scheduler instances (leader election)
- [ ] Scale workers based on load

</coding_guidelines>
