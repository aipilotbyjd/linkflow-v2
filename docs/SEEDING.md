# Database Seeding Guide

This document describes how to seed your LinkFlow database with test data for development and testing.

## Quick Start

### Option 1: Via HTTP Endpoint (Recommended for Coolify/Deployment)

```bash
# Check seed status
curl http://localhost:8090/api/v1/seed/status

# Seed the database (development mode - no secret required)
curl http://localhost:8090/api/v1/seed

# Seed with custom admin credentials
curl "http://localhost:8090/api/v1/seed?admin_email=admin@mycompany.com&admin_password=MySecurePass123"

# Seed without cleaning existing data
curl "http://localhost:8090/api/v1/seed?clean=false"

# Production mode - requires SEED_SECRET
curl "http://localhost:8090/api/v1/seed?secret=your-seed-secret"
# Or via header
curl -H "X-Seed-Secret: your-seed-secret" http://localhost:8090/api/v1/seed
```

### Option 2: Via CLI

```bash
# Start infrastructure (PostgreSQL + Redis)
make dev-infra

# Run the seeder
make seed
```

## HTTP API Endpoints

### GET /api/v1/seed

Seeds the database with development data.

**Query Parameters:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `secret` | - | Required in production (or use `X-Seed-Secret` header) |
| `admin_email` | `admin@linkflow.dev` | Admin user email |
| `admin_password` | `Admin123!` | Admin user password |
| `clean` | `true` | Set to `false` to keep existing data |

**Security:**
- In `development` or `local` environment: No secret required
- In other environments: Requires `SEED_SECRET` env var to be set and matched

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Database seeded successfully",
    "config": {
      "admin_email": "admin@linkflow.dev",
      "clean_first": true,
      "environment": "development"
    },
    "data_created": {
      "users": 5,
      "workspaces": 3,
      "workflows": 30,
      ...
    },
    "test_accounts": [
      {"email": "admin@linkflow.dev", "password": "Admin123!"},
      ...
    ]
  }
}
```

### GET /api/v1/seed/status

Check current database status (no authentication required).

**Response:**
```json
{
  "success": true,
  "data": {
    "seeded": true,
    "counts": {
      "users": 5,
      "workspaces": 3,
      "workflows": 30,
      "executions": 60,
      "credentials": 12
    },
    "environment": "development",
    "seed_secret_set": false,
    "seed_endpoint": "/api/v1/seed"
  }
}
```

## Coolify Deployment

### Step 1: Set Environment Variable (Production)

In Coolify, add this environment variable:
```
SEED_SECRET=your-secure-random-string
```

### Step 2: Deploy Your App

Deploy normally through Coolify.

### Step 3: Seed the Database

After deployment, hit the seed endpoint:
```bash
curl "https://your-app.coolify.io/api/v1/seed?secret=your-secure-random-string"
```

Or use the Coolify terminal:
```bash
curl "http://localhost:8090/api/v1/seed"
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `make seed` | Seed development data (default: clean existing dev data first) |
| `make seed-clean` | Same as `make seed` with explicit clean flag |
| `make seed-fresh` | Start infrastructure + clean and seed |

## CLI Options

```bash
go run ./cmd/seed [options]

Options:
  -clean          Clean existing dev data before seeding (default: true)
  -admin-email    Admin user email (default: admin@linkflow.dev)
  -admin-password Admin user password (default: Admin123!)
```

## What Gets Seeded

The development seeder creates comprehensive test data:

### Users (5 total)

| Email | Name | Role |
|-------|------|------|
| `admin@linkflow.dev` | Admin User | Admin |
| `john@linkflow.dev` | John Developer | Developer |
| `jane@linkflow.dev` | Jane Designer | Designer |
| `bob@linkflow.dev` | Bob Manager | Manager |
| `alice@linkflow.dev` | Alice Analyst | Analyst |

**Default Password:** `Admin123!` (same for all users)

### Workspaces (3 total)

| Name | Slug | Plan | Owner |
|------|------|------|-------|
| Acme Corporation | acme-corp | Pro | Admin |
| Marketing Team | marketing-team | Starter | John |
| DevOps Squad | devops-squad | Business | Jane |

### Workspace Members (11 total)

Each workspace has members with different roles:
- **Acme Corp:** 5 members (owner, admin, 2 members, viewer)
- **Marketing Team:** 3 members (owner, admin, member)
- **DevOps Squad:** 3 members (owner, admin, member)

### Folders (13 total)

Organized hierarchically:

**Acme Corporation:**
- Email Automations
  - Welcome Emails (sub-folder)
  - Newsletter (sub-folder)
- CRM Integrations
- Data Pipelines
- Notifications

**Marketing Team:**
- Lead Generation
- Social Media
- Analytics

**DevOps Squad:**
- CI/CD Pipelines
  - GitHub Actions (sub-folder)
- Monitoring
- Infrastructure

### Workflows (30 total)

Distributed across workspaces and folders with various statuses:
- Active workflows with realistic execution counts
- Inactive workflows
- Draft workflows
- Archived workflows

Example workflows:
- Welcome Email Sequence
- Newsletter Sender
- Salesforce Lead Sync
- Daily Data Export
- Slack Alert on Error
- GitHub PR Notifier
- Uptime Monitor
- And many more...

### Credentials (12 total)

Various integration credentials:
- Slack Bot Token
- SendGrid API Key
- PostgreSQL Production
- AWS S3 Bucket
- Stripe API
- HubSpot API
- GitHub Token
- Docker Hub
- PagerDuty API
- Datadog API
- And more...

### Schedules (7 total)

Cron-based schedules:
- Daily Data Export (0 0 * * *)
- Nightly S3 Backup (0 2 * * *)
- Health Check Every 5 Min (*/5 * * * *)
- Weekly Report Monday 9AM (0 9 * * 1)
- Uptime Check Every Minute (* * * * *)
- Monthly Cost Report (0 6 1 * *)
- Daily SSL Check (0 0 * * *)

### Executions (~60 total)

3-5 executions per workflow for the first 15 workflows:
- Completed executions with output data
- Failed executions with error messages
- Running executions
- Queued executions

### Environment Variables (7 total)

Workspace-level variables:
- API_BASE_URL
- DEBUG_MODE
- SECRET_KEY (secret)
- MARKETING_API_KEY (secret)
- CAMPAIGN_PREFIX
- DEPLOY_ENV
- SLACK_WEBHOOK (secret)

## Data Counts Summary

| Entity | Count |
|--------|-------|
| Users | 5 |
| Workspaces | 3 |
| Workspace Members | 11 |
| Folders | 13 |
| Workflows | 30 |
| Credentials | 12 |
| Schedules | 7 |
| Executions | ~60 |
| Environment Variables | 7 |
| **Total Records** | **~148** |

## Test Accounts

After seeding, you can login with any of these accounts:

```
Email: admin@linkflow.dev
Password: Admin123!
```

All test users have the same password: `Admin123!`

## Programmatic Usage

You can also call the seeder programmatically:

```go
import "github.com/linkflow-ai/linkflow/internal/pkg/database"

// With default config
err := database.SeedDevelopment(db, database.DefaultDevSeedConfig())

// With custom config
cfg := database.DevSeedConfig{
    AdminEmail:    "custom@example.com",
    AdminPassword: "CustomPass123!",
    CleanFirst:    true,
}
err := database.SeedDevelopment(db, cfg)
```

## Base Seeders

In addition to development data, the following base seeders run automatically:

1. **SeedPlans** - Creates pricing plans (Free, Starter, Pro, Business, Enterprise)
2. **SeedPermissions** - Creates permission definitions
3. **SeedRoles** - Creates system roles

These run before development seeding via `SeedAll()`.

## Cleaning Data

The seeder can optionally clean existing development data before seeding. This deletes recent records from:
- node_executions
- execution_logs
- executions
- schedules
- webhook_endpoints
- workflow_versions
- workflows
- credentials
- environment_variables
- workspace_members
- workspaces

**Note:** Users are not deleted to preserve any real accounts.

## Troubleshooting

### "duplicate key value violates unique constraint"

This means some seed data already exists. Run with `-clean=true`:
```bash
go run ./cmd/seed -clean=true
```

### "connection refused"

Make sure PostgreSQL is running:
```bash
make dev-infra
```

### "relation does not exist"

Run migrations first:
```bash
go run ./cmd/api migrate
```
