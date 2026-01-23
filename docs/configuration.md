# LinkFlow Configuration Guide

This document explains how LinkFlow's configuration system works, your responsibilities as a developer/operator, and how to manage configurations across different environments.

## Table of Contents

1. [Overview](#overview)
2. [How Configuration Works](#how-configuration-works)
3. [Configuration Priority](#configuration-priority)
4. [Environment Variables Reference](#environment-variables-reference)
5. [Configuration Files](#configuration-files)
6. [Your Role & Responsibilities](#your-role--responsibilities)
7. [Deployment Guide](#deployment-guide)
8. [Security Best Practices](#security-best-practices)
9. [Troubleshooting](#troubleshooting)

---

## Overview

LinkFlow uses a **layered configuration system** that supports:

- YAML configuration files for non-sensitive defaults
- Environment variables for secrets and environment-specific overrides
- Connection URL parsing (DATABASE_URL, REDIS_URL) for cloud platforms
- Environment-specific overlays (production, staging)
- Comprehensive validation on startup

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     CONFIGURATION LOADING FLOW                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   1. Load .env files (.env, .env.local, .env.{environment})            │
│                           │                                             │
│                           ▼                                             │
│   2. Set code defaults (setDefaults function)                          │
│                           │                                             │
│                           ▼                                             │
│   3. Parse DATABASE_URL / REDIS_URL if present                         │
│                           │                                             │
│                           ▼                                             │
│   4. Read configs/config.yaml                                          │
│                           │                                             │
│                           ▼                                             │
│   5. Merge configs/config.{environment}.yaml                           │
│                           │                                             │
│                           ▼                                             │
│   6. Apply environment variable overrides                              │
│                           │                                             │
│                           ▼                                             │
│   7. Validate configuration                                            │
│                           │                                             │
│                           ▼                                             │
│   8. Return Config struct (immutable)                                  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## How Configuration Works

### Loading Order

The configuration is loaded in this order (later sources override earlier ones):

| Priority | Source | Example |
|----------|--------|---------|
| 1 (Lowest) | Code defaults | `viper.SetDefault("server.port", 8080)` |
| 2 | Base config file | `configs/config.yaml` |
| 3 | Environment overlay | `configs/config.production.yaml` |
| 4 | .env files | `.env`, `.env.local` |
| 5 | Connection URLs | `DATABASE_URL`, `REDIS_URL` |
| 6 (Highest) | Environment variables | `DATABASE_HOST=...` |

### Key Principle: Secrets NEVER in Config Files

```yaml
# configs/config.yaml - SAFE TO COMMIT
database:
  host: localhost
  port: 5432
  user: postgres
  # password: NEVER HERE - only via environment variable
  name: linkflow

# .env - NEVER COMMIT (gitignored)
DATABASE_PASSWORD=your-secret-password
```

### Singleton Pattern

The configuration is loaded once at startup and accessed via singleton:

```go
// In your code
cfg := config.Get()  // Thread-safe, loaded once

// Access values
port := cfg.Server.Port
dbHost := cfg.Database.Host
```

---

## Configuration Priority

### Understanding Override Behavior

```
Environment Variable: DATABASE_HOST=production-db.example.com
     │
     │  OVERRIDES
     ▼
config.production.yaml: database.host: staging-db
     │
     │  OVERRIDES
     ▼
config.yaml: database.host: localhost
     │
     │  OVERRIDES
     ▼
Code default: database.host = "localhost"
```

### Example Scenario

You have:
- `configs/config.yaml`: `server.port: 8090`
- `configs/config.production.yaml`: (no port override)
- Environment variable: `PORT=3000`

**Result**: Server runs on port 3000 (env var wins)

---

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | JWT signing key (min 32 chars) | `openssl rand -base64 64` |
| `DATABASE_PASSWORD` | Database password | (or use `DATABASE_URL`) |

### Database Configuration

```bash
# Option 1: Connection URL (recommended for cloud)
DATABASE_URL=postgresql://user:pass@host:5432/db?sslmode=require

# Option 2: Individual variables
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=secret
DATABASE_NAME=linkflow
DATABASE_SSLMODE=disable  # Use 'require' in production
```

### Redis Configuration

```bash
# Option 1: Connection URL (recommended for cloud)
REDIS_URL=redis://user:pass@host:6379/0
REDIS_URL=rediss://user:pass@host:6379/0  # 'rediss' = TLS

# Option 2: Individual variables
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=secret
REDIS_DB=0
REDIS_TLS=false  # Use true for cloud Redis
```

### Application Settings

```bash
APP_ENVIRONMENT=development  # development, staging, production, test
APP_DEBUG=true               # false in production
APP_URL=http://localhost:8080
APP_FRONTEND_URL=http://localhost:3000
PORT=8090                    # Server port (Railway/Render use PORT)
```

### Optional Services

```bash
# OAuth (Google)
OAUTH_GOOGLE_CLIENT_ID=...
OAUTH_GOOGLE_CLIENT_SECRET=...
OAUTH_GOOGLE_REDIRECT_URL=http://localhost:3000/auth/callback/google

# OAuth (GitHub)
OAUTH_GITHUB_CLIENT_ID=...
OAUTH_GITHUB_CLIENT_SECRET=...
OAUTH_GITHUB_REDIRECT_URL=http://localhost:3000/auth/callback/github

# OAuth (Microsoft)
OAUTH_MICROSOFT_CLIENT_ID=...
OAUTH_MICROSOFT_CLIENT_SECRET=...
OAUTH_MICROSOFT_REDIRECT_URL=http://localhost:3000/auth/callback/microsoft

# AI Providers (for AI Builder)
AI_OPENAI_API_KEY=sk-...
AI_ANTHROPIC_API_KEY=sk-ant-...

# S3/Storage
S3_ENDPOINT=              # Leave empty for AWS, set for MinIO/R2
S3_REGION=us-east-1
S3_BUCKET=linkflow
S3_ACCESS_KEY_ID=...
S3_SECRET_ACCESS_KEY=...

# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# SMTP
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=...
SMTP_PASSWORD=...
SMTP_FROM=noreply@example.com

# Encryption (for credential storage)
ENCRYPTION_KEY=...        # 32 chars for AES-256

# Tracing
TRACING_ENDPOINT=http://localhost:4318

# Logging
LOG_LEVEL=info            # trace, debug, info, warn, error
LOG_FORMAT=json           # json, console
```

---

## Configuration Files

### File Structure

```
linkflow-v2/
├── configs/
│   ├── config.yaml              # Base config (committed)
│   ├── config.production.yaml   # Production overlay (committed)
│   ├── config.staging.yaml      # Staging overlay (committed)
│   └── config.test.yaml         # Test overlay (committed)
├── .env.example                 # Template (committed)
├── .env                         # Local secrets (NEVER commit)
└── .env.local                   # Local overrides (NEVER commit)
```

### configs/config.yaml

Contains non-sensitive defaults. Safe to commit.

```yaml
app:
  name: linkflow
  environment: development
  debug: true

server:
  port: 8090
  read_timeout: 30s

database:
  host: localhost
  port: 5432
  user: postgres
  # NO PASSWORD HERE
  name: linkflow
  sslmode: disable
```

### configs/config.production.yaml

Production-specific overrides. Automatically loaded when `APP_ENVIRONMENT=production`.

```yaml
app:
  debug: false

database:
  sslmode: require
  max_open_conns: 50

redis:
  tls: true
  pool_size: 20

logging:
  level: info
  redact_secrets: true
```

### .env File

Local development secrets. **NEVER COMMIT**.

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/linkflow

# or individual variables
DATABASE_PASSWORD=local-dev-password

# Redis
REDIS_URL=redis://localhost:6379/0

# JWT
JWT_SECRET=your-32-character-development-key-here
```

---

## Your Role & Responsibilities

### As a Developer (Local Development)

1. **One-time setup**:
   ```bash
   cp .env.example .env
   # Edit .env with your local credentials
   ```

2. **Never commit**:
   - `.env` files
   - Real credentials
   - API keys

3. **Keep updated**:
   - When new env vars are added, update your `.env`
   - Check `.env.example` for changes

### As a DevOps/Operator (Deployment)

1. **Set environment variables** in your deployment platform:
   - Railway: Project Settings → Variables
   - Render: Dashboard → Environment
   - Kubernetes: ConfigMaps/Secrets
   - Docker: `-e` flags or `.env` file

2. **Required for production**:
   ```bash
   APP_ENVIRONMENT=production
   APP_DEBUG=false
   JWT_SECRET=<64+ char secret>
   DATABASE_URL=postgresql://...?sslmode=require
   REDIS_URL=rediss://...  # Note: rediss for TLS
   ENCRYPTION_KEY=<32 char key>
   ```

3. **Validate before deploy**:
   - All required secrets set
   - SSL modes enabled
   - Debug disabled
   - Proper CORS origins

---

## Deployment Guide

### Railway

1. **Connect your repo** to Railway

2. **Set environment variables**:
   ```bash
   APP_ENVIRONMENT=production
   APP_DEBUG=false
   JWT_SECRET=<generate with: openssl rand -base64 64>
   
   # Railway provides DATABASE_URL automatically if you add PostgreSQL
   # Railway provides REDIS_URL automatically if you add Redis
   ```

3. **Railway auto-detects**:
   - `PORT` environment variable (set automatically)
   - Go runtime

### Render

1. **Create Web Service** from your repo

2. **Set environment variables** in Dashboard:
   ```bash
   APP_ENVIRONMENT=production
   DATABASE_URL=<from Render PostgreSQL>
   REDIS_URL=<from Render Redis or Upstash>
   JWT_SECRET=<your secret>
   ```

3. **Health check**: Set to `/health`

### Docker

```bash
# Using .env file
docker run --env-file .env.production linkflow/api

# Using individual variables
docker run \
  -e APP_ENVIRONMENT=production \
  -e DATABASE_URL=postgresql://... \
  -e REDIS_URL=rediss://... \
  -e JWT_SECRET=... \
  linkflow/api
```

### Kubernetes

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: linkflow-secrets
type: Opaque
stringData:
  JWT_SECRET: "your-secret-key"
  DATABASE_URL: "postgresql://..."
  REDIS_URL: "rediss://..."
  ENCRYPTION_KEY: "your-32-char-key"

# deployment.yaml
spec:
  containers:
  - name: api
    envFrom:
    - secretRef:
        name: linkflow-secrets
    env:
    - name: APP_ENVIRONMENT
      value: "production"
```

---

## Security Best Practices

### DO ✅

1. **Use environment variables for all secrets**
2. **Use connection URLs** (`DATABASE_URL`, `REDIS_URL`) for cloud services
3. **Enable SSL/TLS**:
   - `DATABASE_SSLMODE=require`
   - `REDIS_TLS=true` (or `rediss://` URL)
4. **Use strong secrets**:
   ```bash
   # JWT Secret (64+ chars)
   openssl rand -base64 64
   
   # Encryption Key (32 chars)
   openssl rand -base64 32 | head -c 32
   ```
5. **Rotate secrets periodically**
6. **Use secret managers** in production (AWS Secrets Manager, HashiCorp Vault)

### DON'T ❌

1. **Never commit** `.env` files
2. **Never put passwords** in `config.yaml`
3. **Never use** `sslmode=disable` in production
4. **Never use** wildcard (`*`) CORS origins in production
5. **Never enable** debug mode in production
6. **Never log** sensitive data

### Production Checklist

```bash
# Before deploying, verify:
[ ] APP_ENVIRONMENT=production
[ ] APP_DEBUG=false
[ ] DATABASE_SSLMODE=require (or in DATABASE_URL)
[ ] REDIS_TLS=true (or rediss:// URL)
[ ] JWT_SECRET is 64+ characters
[ ] ENCRYPTION_KEY is set (32 chars)
[ ] CORS origins are specific (no wildcards)
[ ] Rate limiting is enabled
[ ] LOG_LEVEL=info (not debug)
```

---

## Troubleshooting

### Common Issues

#### "jwt.secret is required"
```bash
# Solution: Set JWT_SECRET environment variable
export JWT_SECRET=$(openssl rand -base64 64)
```

#### "database.host is required"
```bash
# Solution: Set database configuration
export DATABASE_URL=postgresql://user:pass@localhost:5432/linkflow
# or
export DATABASE_HOST=localhost
export DATABASE_PASSWORD=yourpassword
```

#### "config validation error: database.sslmode should not be 'disable' in production"
```bash
# Solution: Enable SSL for production
export DATABASE_SSLMODE=require
# or use DATABASE_URL with ?sslmode=require
```

#### Connection refused to Redis
```bash
# Check if using TLS for cloud Redis
export REDIS_TLS=true
# or use rediss:// URL scheme
export REDIS_URL=rediss://user:pass@host:6379/0
```

### Debugging Configuration

```go
// Print loaded configuration (secrets masked)
cfg := config.Get()
fmt.Printf("%+v\n", cfg.Redact())

// Check specific values
fmt.Println("Environment:", cfg.App.Environment)
fmt.Println("Database Host:", cfg.Database.Host)
fmt.Println("Redis TLS:", cfg.Redis.TLS)
```

### Validation Errors

The configuration validates on startup. If validation fails, you'll see errors like:

```
config validation errors (3):
  1. jwt.secret must be at least 32 characters
  2. database.sslmode should not be 'disable' in production
  3. app.debug should be false in production
```

Fix each error by setting the appropriate environment variable.

---

## Quick Reference

### Generate Secrets

```bash
# JWT Secret (64 chars recommended for production)
openssl rand -base64 64

# Encryption Key (exactly 32 chars for AES-256)
openssl rand -base64 32 | head -c 32

# Random password
openssl rand -base64 24
```

### Test Connections

```bash
# Test PostgreSQL
psql "$DATABASE_URL" -c "SELECT 1"

# Test Redis
redis-cli -u "$REDIS_URL" ping
```

### Environment Quick Setup

```bash
# Development
cp .env.example .env
# Edit .env with local values

# Production (set in your platform)
APP_ENVIRONMENT=production
APP_DEBUG=false
JWT_SECRET=<64+ chars>
DATABASE_URL=postgresql://...?sslmode=require
REDIS_URL=rediss://...
ENCRYPTION_KEY=<32 chars>
```

---

## Summary

| Environment | Config File | Secrets Location |
|-------------|-------------|------------------|
| Development | `config.yaml` | `.env` file |
| Staging | `config.yaml` + `config.staging.yaml` | Platform env vars |
| Production | `config.yaml` + `config.production.yaml` | Platform env vars |
| Test | `config.yaml` + `config.test.yaml` | Test code / CI env vars |

**Remember**: Configuration files contain defaults. Secrets always come from environment variables.
