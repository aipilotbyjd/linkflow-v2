# Data Model

This document describes the database schema and data relationships in LinkFlow.

## Entity Relationship Diagram

```
┌─────────────┐       ┌─────────────────┐       ┌─────────────┐
│   users     │───────│workspace_members│───────│  workspaces │
└──────┬──────┘       └─────────────────┘       └──────┬──────┘
       │                                               │
       │              ┌─────────────────┐              │
       └──────────────│   workflows     │──────────────┘
                      └────────┬────────┘
                               │
       ┌───────────────────────┼───────────────────────┐
       │                       │                       │
┌──────▼──────┐        ┌───────▼───────┐       ┌───────▼───────┐
│  schedules  │        │  executions   │       │ workflow_     │
└─────────────┘        └───────┬───────┘       │ versions      │
                               │               └───────────────┘
                       ┌───────▼───────┐
                       │node_executions│
                       └───────────────┘

┌─────────────┐       ┌─────────────────┐       ┌─────────────┐
│ credentials │       │   api_keys      │       │  sessions   │
└─────────────┘       └─────────────────┘       └─────────────┘
```

## Core Tables

### users

User accounts and authentication.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) | Unique email |
| password_hash | VARCHAR(255) | bcrypt hash |
| first_name | VARCHAR(100) | First name |
| last_name | VARCHAR(100) | Last name |
| status | VARCHAR(20) | active/suspended/deleted |
| email_verified | BOOLEAN | Email verified flag |
| mfa_enabled | BOOLEAN | MFA enabled flag |
| mfa_secret | VARCHAR(255) | Encrypted TOTP secret |
| created_at | TIMESTAMPTZ | Creation timestamp |
| updated_at | TIMESTAMPTZ | Last update |

### workspaces

Multi-tenant workspace containers.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| owner_id | UUID | FK to users |
| name | VARCHAR(100) | Workspace name |
| slug | VARCHAR(100) | Unique URL slug |
| plan_id | VARCHAR(50) | Billing plan |
| stripe_customer_id | VARCHAR(255) | Stripe customer |
| settings | JSONB | Workspace settings |
| created_at | TIMESTAMPTZ | Creation timestamp |

### workspace_members

Workspace membership and roles.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workspace_id | UUID | FK to workspaces |
| user_id | UUID | FK to users |
| role | VARCHAR(20) | owner/admin/member/viewer |
| joined_at | TIMESTAMPTZ | Join timestamp |

### workflows

Workflow definitions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workspace_id | UUID | FK to workspaces |
| created_by | UUID | FK to users |
| name | VARCHAR(255) | Workflow name |
| description | TEXT | Description |
| status | VARCHAR(20) | draft/active/archived |
| version | INTEGER | Current version |
| nodes | JSONB | Node definitions |
| connections | JSONB | Node connections |
| settings | JSONB | Workflow settings |
| tags | TEXT[] | Tags array |
| execution_count | INTEGER | Total executions |
| last_executed_at | TIMESTAMPTZ | Last execution |
| created_at | TIMESTAMPTZ | Creation timestamp |
| updated_at | TIMESTAMPTZ | Last update |

### workflow_versions

Workflow version history.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workflow_id | UUID | FK to workflows |
| version | INTEGER | Version number |
| nodes | JSONB | Node definitions |
| connections | JSONB | Connections |
| created_by | UUID | FK to users |
| change_message | TEXT | Version message |
| created_at | TIMESTAMPTZ | Creation timestamp |

### executions

Workflow execution records.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workflow_id | UUID | FK to workflows |
| workspace_id | UUID | FK to workspaces |
| triggered_by | UUID | FK to users |
| workflow_version | INTEGER | Version executed |
| status | VARCHAR(20) | queued/running/completed/failed/cancelled |
| trigger_type | VARCHAR(20) | manual/schedule/webhook |
| trigger_data | JSONB | Trigger context |
| input_data | JSONB | Input data |
| output_data | JSONB | Output data |
| error_message | TEXT | Error message |
| error_node_id | VARCHAR(100) | Failed node ID |
| queued_at | TIMESTAMPTZ | Queue timestamp |
| started_at | TIMESTAMPTZ | Start timestamp |
| completed_at | TIMESTAMPTZ | Completion timestamp |
| nodes_total | INTEGER | Total nodes |
| nodes_completed | INTEGER | Completed nodes |
| retry_count | INTEGER | Retry attempts |
| parent_execution_id | UUID | Parent execution (retry) |

### node_executions

Individual node execution records.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| execution_id | UUID | FK to executions |
| node_id | VARCHAR(100) | Node ID in workflow |
| node_type | VARCHAR(50) | Node type |
| node_name | VARCHAR(255) | Node display name |
| status | VARCHAR(20) | pending/running/completed/failed/skipped |
| input_data | JSONB | Node input |
| output_data | JSONB | Node output |
| error_message | TEXT | Error message |
| started_at | TIMESTAMPTZ | Start timestamp |
| completed_at | TIMESTAMPTZ | Completion timestamp |
| duration_ms | INTEGER | Execution duration |
| retry_count | INTEGER | Retry attempts |

### credentials

Encrypted credential storage.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workspace_id | UUID | FK to workspaces |
| created_by | UUID | FK to users |
| name | VARCHAR(100) | Credential name |
| type | VARCHAR(50) | api_key/oauth2/basic |
| data | TEXT | AES-256-GCM encrypted |
| description | TEXT | Description |
| created_at | TIMESTAMPTZ | Creation timestamp |
| last_used_at | TIMESTAMPTZ | Last usage |

### schedules

Cron schedules for workflows.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workflow_id | UUID | FK to workflows |
| workspace_id | UUID | FK to workspaces |
| created_by | UUID | FK to users |
| name | VARCHAR(100) | Schedule name |
| cron_expression | VARCHAR(100) | Cron expression |
| timezone | VARCHAR(50) | Timezone |
| is_active | BOOLEAN | Active flag |
| input_data | JSONB | Default input |
| next_run_at | TIMESTAMPTZ | Next execution |
| last_run_at | TIMESTAMPTZ | Last execution |
| run_count | INTEGER | Total runs |

### webhook_endpoints

Webhook trigger endpoints.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| workflow_id | UUID | FK to workflows |
| workspace_id | UUID | FK to workspaces |
| endpoint_id | VARCHAR(100) | Public endpoint ID |
| path | VARCHAR(255) | Webhook path |
| method | VARCHAR(10) | HTTP method |
| is_active | BOOLEAN | Active flag |
| secret | VARCHAR(255) | Verification secret |
| created_at | TIMESTAMPTZ | Creation timestamp |

### sessions

User sessions.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | FK to users |
| token_hash | VARCHAR(255) | Token hash |
| refresh_hash | VARCHAR(255) | Refresh token hash |
| ip_address | VARCHAR(45) | Client IP |
| user_agent | TEXT | User agent |
| expires_at | TIMESTAMPTZ | Expiration |
| created_at | TIMESTAMPTZ | Creation timestamp |
| revoked_at | TIMESTAMPTZ | Revocation timestamp |

### api_keys

API key authentication.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| user_id | UUID | FK to users |
| workspace_id | UUID | FK to workspaces |
| name | VARCHAR(100) | Key name |
| key_prefix | VARCHAR(10) | Key prefix (lf_) |
| key_hash | VARCHAR(255) | Key hash |
| scopes | TEXT[] | Permission scopes |
| last_used_at | TIMESTAMPTZ | Last usage |
| expires_at | TIMESTAMPTZ | Expiration |
| created_at | TIMESTAMPTZ | Creation timestamp |
| revoked_at | TIMESTAMPTZ | Revocation timestamp |

## Indexes

```sql
-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status) WHERE status = 'active';

-- Workflows
CREATE INDEX idx_workflows_workspace ON workflows(workspace_id);
CREATE INDEX idx_workflows_status ON workflows(status);
CREATE INDEX idx_workflows_tags ON workflows USING GIN(tags);

-- Executions
CREATE INDEX idx_executions_workflow ON executions(workflow_id);
CREATE INDEX idx_executions_workspace ON executions(workspace_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_created ON executions(created_at DESC);

-- Node Executions
CREATE INDEX idx_node_executions_execution ON node_executions(execution_id);

-- Schedules
CREATE INDEX idx_schedules_next_run ON schedules(next_run_at) WHERE is_active = true;
```

## JSON Schema

### Workflow Nodes

```json
{
  "id": "node-uuid",
  "type": "action.http",
  "name": "HTTP Request",
  "position": {
    "x": 100,
    "y": 200
  },
  "config": {
    "method": "GET",
    "url": "https://api.example.com",
    "headers": {},
    "body": {}
  },
  "credentials": "credential-uuid"
}
```

### Workflow Connections

```json
{
  "source": "trigger-node-id",
  "target": "action-node-id",
  "sourceHandle": "main",
  "targetHandle": "main"
}
```

### Execution Context

```json
{
  "workflow_id": "uuid",
  "execution_id": "uuid",
  "trigger_type": "manual",
  "input_data": {},
  "variables": {},
  "node_outputs": {
    "trigger": {},
    "http_1": {}
  }
}
```

## Data Lifecycle

### Execution Retention

- Active executions: Indefinite
- Completed executions: 30 days (configurable)
- Failed executions: 90 days
- Audit logs: 1 year

### Cleanup Jobs

```sql
-- Delete old executions (runs daily)
DELETE FROM executions 
WHERE status = 'completed' 
  AND completed_at < NOW() - INTERVAL '30 days';

-- Delete old node executions
DELETE FROM node_executions
WHERE execution_id NOT IN (SELECT id FROM executions);
```

## Partitioning

For high-volume deployments, partition executions by month:

```sql
CREATE TABLE executions (
    id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    ...
) PARTITION BY RANGE (created_at);

CREATE TABLE executions_2024_01 PARTITION OF executions
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## Migration Strategy

Migrations are embedded in the application and run on startup:

```go
//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(db *sql.DB) error {
    // Apply migrations in order
}
```
