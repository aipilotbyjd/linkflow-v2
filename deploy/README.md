# LinkFlow Deployment

## CI/CD Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           AUTOMATED PIPELINE                             │
└─────────────────────────────────────────────────────────────────────────┘

  git push                    git push                   git tag v1.0.0
  to dev branch               to main branch             (or manual)
       │                           │                          │
       ▼                           ▼                          ▼
┌─────────────┐            ┌─────────────┐            ┌─────────────┐
│    CI       │            │    CI       │            │    CI       │
│  ─────────  │            │  ─────────  │            │  ─────────  │
│  • Lint     │            │  • Lint     │            │  • Lint     │
│  • Test     │            │  • Test     │            │  • Test     │
│  • Build    │            │  • Build    │            │  • Build    │
│  • Security │            │  • Security │            │  • Security │
└──────┬──────┘            └──────┬──────┘            └──────┬──────┘
       │ ✅                       │ ✅                       │ ✅
       ▼                          ▼                          ▼
┌─────────────┐            ┌─────────────┐            ┌─────────────┐
│  DEV        │            │  STAGING    │            │  PRODUCTION │
│  Auto-deploy│            │  Auto-deploy│            │  + Backup   │
│             │            │  + Health   │            │  + Health   │
│             │            │    check    │            │  + Notify   │
└─────────────┘            └─────────────┘            └─────────────┘
```

### GitHub Secrets Required

| Secret | Description |
|--------|-------------|
| `FLY_API_TOKEN` | Fly.io API token |
| `SLACK_WEBHOOK_URL` | Slack notifications |
| `R2_ACCESS_KEY` | Cloudflare R2 for backups |
| `R2_SECRET_KEY` | Cloudflare R2 secret |
| `R2_ENDPOINT` | Cloudflare R2 endpoint |

## Quick Start

```bash
# Local development
./scripts/deploy local

# Deploy to production
./scripts/deploy production fly
```

## Structure

```
deploy/
├── docker/                 # Dockerfiles
├── config/                 # Environment configs
├── compose/                # Docker Compose files
├── platforms/              # Platform-specific configs
│   ├── coolify/           # Self-hosted ($7/mo)
│   ├── fly/               # Fly.io ($30/mo)
│   ├── railway/           # Railway
│   ├── render/            # Render
│   └── kubernetes/        # K8s (scale phase)
├── monitoring/             # Prometheus, Grafana, Alerts
├── scripts/                # Management scripts
├── backup/                 # Backup scripts
└── runbooks/               # Emergency procedures
```

## Environments

| Environment | Purpose | URL |
|-------------|---------|-----|
| local | Development | localhost:8090 |
| dev | Team testing | dev.linkflow.io |
| staging | Pre-production | staging.linkflow.io |
| production | Live | api.linkflow.io |

## Commands

```bash
# Deploy
./scripts/deploy <env> [platform]     # Deploy
./scripts/rollback <env>              # Rollback
./scripts/scale <env> <service> <n>   # Scale

# Database
./scripts/db-migrate <env>            # Run migrations
./scripts/backup <env>                # Backup
./scripts/restore <env> <file>        # Restore

# Monitoring
./scripts/status <env>                # Health check
./scripts/logs <env> [service]        # View logs
./scripts/shell <env> [service]       # Shell access
```

## Platforms

### Phase 1: Bootstrap ($7-15/mo)
Use **Coolify + Hetzner**
```bash
./scripts/deploy production coolify
```

### Phase 2: Growth ($50-100/mo)
Use **Fly.io**
```bash
./scripts/deploy production fly
```

### Phase 3: Scale ($200+/mo)
Use **Kubernetes**
```bash
./scripts/deploy production k8s
```

## Setup New Environment

1. Copy config:
   ```bash
   cp deploy/config/production.env.example deploy/config/production.env
   ```

2. Fill in secrets

3. Deploy:
   ```bash
   ./scripts/deploy production
   ```

## Backup & Recovery

```bash
# Backup
./scripts/backup production           # Full backup
./scripts/backup production db        # Database only

# Restore
./scripts/restore production <file>
```

## Monitoring

Start monitoring stack:
```bash
docker compose -f deploy/compose/docker-compose.yml \
               -f deploy/compose/docker-compose.monitoring.yml up -d
```

Access:
- Grafana: http://localhost:3001
- Prometheus: http://localhost:9090
- Alertmanager: http://localhost:9093

## Runbooks

See `deploy/runbooks/` for emergency procedures:
- [Incident Response](runbooks/incident-response.md)
- [Database Down](runbooks/database-down.md)
- [High CPU](runbooks/high-cpu.md)
- [Deployment Failed](runbooks/deployment-failed.md)
