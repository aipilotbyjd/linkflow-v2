# LinkFlow - Workflow Automation Platform

LinkFlow is a production-ready n8n-like workflow automation platform built in Go. It enables users to create, execute, and monitor automated workflows through a visual interface with support for webhooks, schedules, and manual triggers.

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.23+ | Backend services |
| Web Framework | Chi v5 | HTTP routing |
| ORM | GORM | Database operations |
| Database | PostgreSQL 15+ | Persistent storage |
| Cache | Redis 7+ | Caching, sessions |
| Job Queue | Asynq | Background job processing |
| Streaming | Redis Streams | Durable webhook buffering |
| Auth | JWT + OAuth2 | Authentication |
| Metrics | Prometheus | Observability |
| Container | Docker | Deployment |

## Architecture Overview

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
            │  :8090    │   │    /ws    │   │ /webhooks │
            └─────┬─────┘   └─────┬─────┘   └─────┬─────┘
                  │               │               │
                  └───────────────┼───────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │        API SERVER         │
                    │   (cmd/api/main.go)       │
                    │                           │
                    │  • Authentication         │
                    │  • Rate Limiting          │
                    │  • Request Validation     │
                    │  • Response Formatting    │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │     │     Redis       │     │   Asynq Queue   │
│                 │     │                 │     │                 │
│  • Users        │     │  • Sessions     │     │  • Executions   │
│  • Workflows    │     │  • Rate Limits  │     │  • Emails       │
│  • Executions   │     │  • Cache        │     │  • Cleanup      │
│  • Credentials  │     │  • Pub/Sub      │     │                 │
└─────────────────┘     │  • Streams      │     └────────┬────────┘
                        └─────────────────┘              │
                                                         │
                        ┌────────────────────────────────┘
                        │
          ┌─────────────▼─────────────┐     ┌─────────────────────────┐
          │         WORKER            │     │       SCHEDULER         │
          │   (cmd/worker/main.go)    │     │ (cmd/scheduler/main.go) │
          │                           │     │                         │
          │  • Execute Workflows      │     │  • Cron Job Dispatch    │
          │  • Process Nodes          │     │  • Schedule Management  │
          │  • Handle Retries         │     │  • Trigger Executions   │
          │  • Stream Consumers       │     │                         │
          └───────────────────────────┘     └─────────────────────────┘
```

## Project Structure

```
linkflow-v2/
├── cmd/                          # Application entrypoints
│   ├── api/main.go               # API server (port 8090)
│   ├── worker/main.go            # Background job processor
│   └── scheduler/main.go         # Cron scheduler
│
├── internal/                     # Private application code
│   ├── api/
│   │   ├── handlers/             # HTTP request handlers
│   │   │   ├── auth.go           # Login, register, OAuth
│   │   │   ├── workflow.go       # Workflow CRUD, execute
│   │   │   ├── execution.go      # Execution monitoring
│   │   │   ├── credential.go     # Credential management
│   │   │   ├── webhook.go        # Webhook receiver
│   │   │   └── ...
│   │   ├── middleware/
│   │   │   ├── auth.go           # JWT validation
│   │   │   ├── ratelimit.go      # Rate limiting
│   │   │   ├── tenant.go         # Workspace context
│   │   │   └── logger.go         # Request logging
│   │   ├── dto/                  # Data transfer objects
│   │   ├── websocket/            # Real-time events
│   │   └── server.go             # Route definitions
│   │
│   ├── domain/
│   │   ├── models/               # Database entities
│   │   │   ├── user.go           # User, UserSession
│   │   │   ├── workspace.go      # Workspace, WorkspaceMember
│   │   │   ├── workflow.go       # Workflow, WorkflowVersion
│   │   │   ├── execution.go      # Execution, NodeExecution
│   │   │   ├── credential.go     # Credential (encrypted)
│   │   │   ├── schedule.go       # Schedule (cron)
│   │   │   └── webhook.go        # WebhookEndpoint, WebhookLog
│   │   ├── repositories/         # Data access layer
│   │   └── services/             # Business logic
│   │
│   ├── worker/
│   │   ├── worker.go             # Asynq worker setup
│   │   ├── executor.go           # Workflow execution engine
│   │   ├── registry.go           # Node type registry
│   │   └── nodes/                # Node implementations
│   │       ├── triggers/         # manual, webhook, schedule
│   │       ├── actions/          # http, email, code
│   │       ├── logic/            # if, switch, merge, filter
│   │       └── integrations/     # slack, github, aws, etc.
│   │
│   ├── scheduler/                # Cron job management
│   │
│   └── pkg/                      # Shared utilities
│       ├── config/               # Viper configuration
│       ├── crypto/               # JWT, AES encryption
│       ├── database/             # GORM setup
│       ├── redis/                # Redis client
│       ├── queue/                # Asynq client
│       ├── streams/              # Redis Streams
│       ├── metrics/              # Prometheus metrics
│       ├── email/                # Email service
│       └── validator/            # Request validation
│
├── configs/
│   └── config.yaml               # Application configuration
│
├── docs/
│   └── api/postman/              # Postman collection
│
├── Dockerfile                    # Container build
├── docker-compose.yaml           # Local development
├── Makefile                      # Build commands
├── go.mod                        # Go modules
└── AGENTS.md                     # This file
```

## Database Schema

### Core Tables

```sql
-- Users and Authentication
users (id, email, password_hash, first_name, last_name, mfa_enabled, ...)
user_sessions (id, user_id, refresh_token, expires_at, ...)

-- Multi-tenancy
workspaces (id, name, slug, owner_id, plan, ...)
workspace_members (id, workspace_id, user_id, role, ...)

-- Workflows
workflows (id, workspace_id, name, status, version, nodes, connections, settings, ...)
workflow_versions (id, workflow_id, version, nodes, connections, ...)

-- Executions
executions (id, workflow_id, workspace_id, status, trigger_type, input_data, output_data, ...)
node_executions (id, execution_id, node_id, status, input_data, output_data, duration_ms, ...)
execution_logs (id, execution_id, node_id, level, message, ...)

-- Credentials
credentials (id, workspace_id, name, type, encrypted_data, ...)

-- Scheduling
schedules (id, workflow_id, cron_expression, timezone, enabled, ...)

-- Webhooks
webhook_endpoints (id, workflow_id, path, method, secret, is_active, ...)
webhook_logs (id, endpoint_id, method, headers, body, response_code, ...)
```

### Model Relationships

```
User 1──N Workspace (owner)
Workspace 1──N WorkspaceMember
Workspace 1──N Workflow
Workspace 1──N Credential
Workflow 1──N WorkflowVersion
Workflow 1──N Execution
Workflow 1──N Schedule
Workflow 1──N WebhookEndpoint
Execution 1──N NodeExecution
```

## Development Setup

### Prerequisites

```bash
# Required
go version   # 1.23+
docker -v    # 20+
docker compose version  # 2.0+

# Recommended
golangci-lint --version  # For linting
```

### Quick Start

```bash
# 1. Clone and enter directory
cd linkflow-v2

# 2. Start infrastructure
docker compose up -d postgres redis

# 3. Copy and configure
cp configs/config.yaml.example configs/config.yaml
# Edit config.yaml with your settings

# 4. Run services (3 terminals)
go run cmd/api/main.go        # API on :8090
go run cmd/worker/main.go     # Worker
go run cmd/scheduler/main.go  # Scheduler
```

### Docker Development

```bash
# Build and start everything
docker compose up -d --build

# View logs
docker compose logs -f api worker scheduler

# Rebuild single service
docker compose up -d --build api

# Stop all
docker compose down

# Reset (including data)
docker compose down -v
```

## Configuration Reference

```yaml
# configs/config.yaml

app:
  name: linkflow
  environment: development  # development, staging, production
  debug: true
  url: http://localhost:8090
  frontend_url: http://localhost:3000

server:
  host: 0.0.0.0
  port: 8090
  read_timeout: 15s
  write_timeout: 15s

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: linkflow
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-32-character-secret-key-here"
  access_expiry: 15m
  refresh_expiry: 168h  # 7 days

features:
  webhook_stream:
    enabled: true
    max_len: 100000
    batch_size: 10
    max_retries: 3
    consumer_count: 2
```

### Environment Variables

All config values can be overridden with environment variables:
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`
- `REDIS_HOST`, `REDIS_PORT`
- `JWT_SECRET`
- `APP_ENVIRONMENT`

## Build & Test Commands

```bash
# Build
go build ./...                    # Build all packages
go build -o bin/api cmd/api       # Build API binary
go build -o bin/worker cmd/worker # Build worker binary

# Lint
golangci-lint run                 # Run all linters
golangci-lint run --fix           # Auto-fix issues

# Test
go test ./...                     # Run all tests
go test -v ./internal/...         # Verbose test output
go test -cover ./...              # With coverage
go test -race ./...               # Race detection

# Generate
go generate ./...                 # Run go:generate directives
```

## API Endpoints Reference

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login, get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Invalidate session |
| POST | `/api/v1/auth/forgot-password` | Request reset |
| POST | `/api/v1/auth/reset-password` | Reset with token |

### Workflows (workspace-scoped)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/{id}/workflows` | List workflows |
| POST | `/workspaces/{id}/workflows` | Create workflow |
| GET | `/workspaces/{id}/workflows/{wfId}` | Get workflow |
| PUT | `/workspaces/{id}/workflows/{wfId}` | Update workflow |
| DELETE | `/workspaces/{id}/workflows/{wfId}` | Delete workflow |
| POST | `/workspaces/{id}/workflows/{wfId}/execute` | Execute workflow |
| POST | `/workspaces/{id}/workflows/{wfId}/activate` | Activate |
| POST | `/workspaces/{id}/workflows/{wfId}/deactivate` | Deactivate |
| GET | `/workspaces/{id}/workflows/{wfId}/versions` | List versions |
| POST | `/workspaces/{id}/workflows/{wfId}/versions/{v}/rollback` | Rollback |

### Executions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/workspaces/{id}/executions` | List executions |
| GET | `/workspaces/{id}/executions/search` | Search with filters |
| GET | `/workspaces/{id}/executions/stats` | Get statistics |
| DELETE | `/workspaces/{id}/executions/bulk` | Bulk delete |
| GET | `/workspaces/{id}/executions/{exId}` | Get execution |
| POST | `/workspaces/{id}/executions/{exId}/cancel` | Cancel |
| POST | `/workspaces/{id}/executions/{exId}/retry` | Retry |
| GET | `/workspaces/{id}/executions/{exId}/nodes` | Get node data |

### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/webhooks/{endpointId}` | Trigger webhook |
| GET | `/webhooks/{endpointId}` | Trigger (GET) |

### Monitoring
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Node Types Reference

### Triggers
| Type | Description |
|------|-------------|
| `trigger.manual` | Manual execution start |
| `trigger.webhook` | HTTP webhook trigger |
| `trigger.schedule` | Cron-based scheduling |

### Actions
| Type | Description |
|------|-------------|
| `action.http` | HTTP request (GET, POST, etc.) |
| `action.email` | Send email via SMTP |
| `action.code` | Execute JavaScript/Python |
| `action.set` | Set variable values |
| `action.respond` | Send webhook response |

### Logic
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

### Integrations
| Type | Description |
|------|-------------|
| `integration.slack` | Slack messages |
| `integration.github` | GitHub API |
| `integration.postgres` | PostgreSQL queries |
| `integration.mysql` | MySQL queries |
| `integration.mongodb` | MongoDB operations |
| `integration.redis` | Redis commands |
| `integration.aws_s3` | S3 operations |
| `integration.google_drive` | Drive operations |
| `integration.twilio` | SMS/calls |
| `integration.sendgrid` | Email via SendGrid |
| `integration.stripe` | Stripe payments |
| `integration.jira` | Jira issues |
| `integration.salesforce` | Salesforce CRM |
| `integration.airtable` | Airtable bases |
| `integration.notion` | Notion pages |
| `integration.graphql` | GraphQL queries |
| `integration.ftp` | FTP operations |

## Code Patterns

### Handler Pattern
```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Get context (user, workspace)
    claims := middleware.GetUserFromContext(r.Context())
    wsCtx := middleware.GetWorkspaceFromContext(r.Context())
    
    // 2. Parse and validate request
    var req dto.CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        dto.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
        return
    }
    if err := validator.Validate(&req); err != nil {
        dto.ValidationErrorResponse(w, err)
        return
    }
    
    // 3. Call service
    result, err := h.service.Create(r.Context(), input)
    if err != nil {
        dto.ErrorResponse(w, http.StatusInternalServerError, "failed to create")
        return
    }
    
    // 4. Return response
    dto.Created(w, result)
}
```

### Service Pattern
```go
func (s *Service) Create(ctx context.Context, input CreateInput) (*Model, error) {
    // 1. Business validation
    if err := s.validateInput(input); err != nil {
        return nil, err
    }
    
    // 2. Create entity
    entity := &models.Entity{
        Field: input.Field,
    }
    
    // 3. Save to repository
    if err := s.repo.Create(ctx, entity); err != nil {
        return nil, fmt.Errorf("failed to create: %w", err)
    }
    
    return entity, nil
}
```

### Repository Pattern
```go
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*models.Entity, error) {
    var entity models.Entity
    err := r.DB().WithContext(ctx).
        Where("id = ?", id).
        First(&entity).Error
    if err != nil {
        return nil, err
    }
    return &entity, nil
}
```

### Node Implementation Pattern
```go
type MyNode struct{}

func (n *MyNode) Execute(ctx context.Context, input NodeInput) (*NodeOutput, error) {
    // 1. Get parameters
    param := input.Parameters["param"].(string)
    
    // 2. Get input data from previous node
    data := input.Data
    
    // 3. Process
    result, err := n.process(param, data)
    if err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }
    
    // 4. Return output
    return &NodeOutput{
        Data: result,
    }, nil
}

func (n *MyNode) Metadata() NodeMetadata {
    return NodeMetadata{
        Type:        "action.mynode",
        Name:        "My Node",
        Description: "Does something useful",
        Category:    "action",
        Version:     "1.0.0",
        Inputs: []NodePort{
            {Name: "main", Type: "any"},
        },
        Outputs: []NodePort{
            {Name: "main", Type: "any"},
        },
        Parameters: []NodeParameter{
            {
                Name:        "param",
                Type:        "string",
                Required:    true,
                Description: "Parameter description",
            },
        },
    }
}
```

## Error Handling

### Custom Errors
```go
// Define in services
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrInvalidInput  = errors.New("invalid input")
)

// Use with wrapping
return nil, fmt.Errorf("failed to get workflow: %w", ErrNotFound)

// Check in handler
if errors.Is(err, services.ErrNotFound) {
    dto.ErrorResponse(w, http.StatusNotFound, "workflow not found")
    return
}
```

### Logging
```go
import "github.com/rs/zerolog/log"

// Info
log.Info().
    Str("workflow_id", id.String()).
    Int("version", version).
    Msg("Workflow activated")

// Error
log.Error().
    Err(err).
    Str("execution_id", execID.String()).
    Msg("Execution failed")

// Debug (only in development)
log.Debug().
    Interface("data", data).
    Msg("Processing data")
```

## Testing

### Unit Test Pattern
```go
func TestService_Create(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    repo := repositories.NewRepository(db)
    svc := services.NewService(repo)
    
    // Test
    result, err := svc.Create(context.Background(), input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result.Field)
}
```

### Integration Test Pattern
```go
func TestAPI_CreateWorkflow(t *testing.T) {
    // Setup server
    srv := setupTestServer(t)
    token := getTestToken(t, srv)
    
    // Make request
    req := httptest.NewRequest("POST", "/api/v1/workspaces/"+wsID+"/workflows", body)
    req.Header.Set("Authorization", "Bearer "+token)
    
    rec := httptest.NewRecorder()
    srv.Router().ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

## Common Tasks

### Adding a New API Endpoint

1. **Add DTO** in `internal/api/dto/`
2. **Add handler method** in `internal/api/handlers/`
3. **Add route** in `internal/api/server.go`
4. **Add to Postman** in `docs/api/postman/`

### Adding a New Node Type

1. **Create node file** in `internal/worker/nodes/{category}/`
2. **Implement `Node` interface** (Execute, Metadata)
3. **Register in registry** in `internal/worker/registry.go`
4. **Add tests** in same directory

### Adding a New Database Model

1. **Create model** in `internal/domain/models/`
2. **Create repository** in `internal/domain/repositories/`
3. **Add to auto-migrate** in `internal/pkg/database/`
4. **Create service** if needed in `internal/domain/services/`

### Adding a New Background Job

1. **Define task type** in `internal/pkg/queue/`
2. **Create handler** in `internal/worker/`
3. **Register handler** in worker setup
4. **Enqueue** from services using queue client

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `connection refused` to DB | Check PostgreSQL is running, verify config |
| `connection refused` to Redis | Check Redis is running, verify config |
| `token expired` | Refresh token or re-login |
| `workspace context required` | Ensure workspace ID in URL path |
| `rate limit exceeded` | Wait or increase limits in config |
| Node execution timeout | Increase `timeout_seconds` in workflow settings |

### Debug Mode

```yaml
# In config.yaml
app:
  debug: true
```

Enables:
- Detailed error messages
- SQL query logging
- Debug-level logs

### Checking Logs

```bash
# Docker
docker compose logs -f api
docker compose logs -f worker

# Local
# Logs print to stdout with zerolog
```

## Security Considerations

1. **JWT Tokens**: 15-minute expiry, refresh tokens rotated
2. **Passwords**: bcrypt hashed (cost 10)
3. **Credentials**: AES-256-GCM encrypted at rest
4. **Rate Limiting**: Per-user, per-workspace, per-endpoint
5. **Token Blacklist**: Redis-based for logout/revocation
6. **CORS**: Configured for frontend URL only
7. **Input Validation**: All requests validated before processing

## Performance Tips

1. **Database**: Use indexes, avoid N+1 queries
2. **Redis**: Use pipelining for batch operations
3. **Webhooks**: Enable Redis Streams for high throughput
4. **Workers**: Scale horizontally with multiple instances
5. **Executions**: Use pagination for large result sets

## Deployment Checklist

- [ ] Set `APP_ENVIRONMENT=production`
- [ ] Use strong `JWT_SECRET` (32+ characters)
- [ ] Enable SSL for database connections
- [ ] Configure proper CORS origins
- [ ] Set up monitoring (Prometheus + Grafana)
- [ ] Configure log aggregation
- [ ] Set up database backups
- [ ] Configure rate limits for production load
- [ ] Enable Redis Streams for webhook buffering
- [ ] Scale workers based on expected load
