# Configuration Guide

LinkFlow uses a layered configuration system with the following priority (highest to lowest):

1. Environment variables
2. Configuration file (`configs/config.yaml`)
3. Default values

## Configuration File

The main configuration file is `configs/config.yaml`:

```yaml
app:
  name: linkflow
  environment: development  # development, staging, production
  debug: true
  url: http://localhost:8080
  frontend_url: http://localhost:3000

server:
  host: 0.0.0.0
  port: 8090
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: linkflow
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "change-this-to-a-32-character-key"
  access_expiry: 15m
  refresh_expiry: 168h  # 7 days
  issuer: linkflow

oauth:
  google:
    client_id: ""
    client_secret: ""
    redirect_url: http://localhost:8080/api/v1/auth/oauth/google/callback
  github:
    client_id: ""
    client_secret: ""
    redirect_url: http://localhost:8080/api/v1/auth/oauth/github/callback
  microsoft:
    client_id: ""
    client_secret: ""
    redirect_url: http://localhost:8080/api/v1/auth/oauth/microsoft/callback

s3:
  endpoint: ""
  region: us-east-1
  bucket: linkflow
  access_key_id: ""
  secret_access_key: ""
  use_ssl: true

stripe:
  secret_key: ""
  webhook_secret: ""
  publishable_key: ""

smtp:
  host: ""
  port: 587
  username: ""
  password: ""
  from: ""
  from_name: LinkFlow
```

## Environment Variables

All configuration options can be overridden using environment variables.

### Application

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Application name | `linkflow` |
| `APP_ENVIRONMENT` | Environment (development/staging/production) | `development` |
| `APP_DEBUG` | Enable debug mode | `true` |
| `APP_URL` | API base URL | `http://localhost:8080` |
| `APP_FRONTEND_URL` | Frontend URL (for CORS) | `http://localhost:3000` |

### Server

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8090` |
| `SERVER_READ_TIMEOUT` | Read timeout | `15s` |
| `SERVER_WRITE_TIMEOUT` | Write timeout | `15s` |

### Database

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_HOST` | PostgreSQL host | `localhost` |
| `DATABASE_PORT` | PostgreSQL port | `5432` |
| `DATABASE_USER` | Database user | `postgres` |
| `DATABASE_PASSWORD` | Database password | `postgres` |
| `DATABASE_NAME` | Database name | `linkflow` |
| `DATABASE_SSLMODE` | SSL mode | `disable` |
| `DATABASE_MAX_OPEN_CONNS` | Max open connections | `25` |
| `DATABASE_MAX_IDLE_CONNS` | Max idle connections | `5` |

### Redis

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | `` |
| `REDIS_DB` | Redis database number | `0` |

### JWT Authentication

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | JWT signing secret | **Required** |
| `JWT_ACCESS_EXPIRY` | Access token expiry | `15m` |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | `168h` |
| `JWT_ISSUER` | JWT issuer | `linkflow` |

### OAuth Providers

| Variable | Description |
|----------|-------------|
| `OAUTH_GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | Google OAuth client secret |
| `OAUTH_GITHUB_CLIENT_ID` | GitHub OAuth client ID |
| `OAUTH_GITHUB_CLIENT_SECRET` | GitHub OAuth client secret |
| `OAUTH_MICROSOFT_CLIENT_ID` | Microsoft OAuth client ID |
| `OAUTH_MICROSOFT_CLIENT_SECRET` | Microsoft OAuth client secret |

### S3/MinIO Storage

| Variable | Description | Default |
|----------|-------------|---------|
| `S3_ENDPOINT` | S3 endpoint (empty for AWS) | `` |
| `S3_REGION` | S3 region | `us-east-1` |
| `S3_BUCKET` | S3 bucket name | `linkflow` |
| `S3_ACCESS_KEY_ID` | S3 access key | `` |
| `S3_SECRET_ACCESS_KEY` | S3 secret key | `` |
| `S3_USE_SSL` | Use SSL | `true` |

### Stripe Billing

| Variable | Description |
|----------|-------------|
| `STRIPE_SECRET_KEY` | Stripe secret key |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret |
| `STRIPE_PUBLISHABLE_KEY` | Stripe publishable key |

### SMTP Email

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server host | `` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP username | `` |
| `SMTP_PASSWORD` | SMTP password | `` |
| `SMTP_FROM` | From email address | `` |
| `SMTP_FROM_NAME` | From name | `LinkFlow` |

## Environment-Specific Configuration

### Development

```bash
# .env
APP_ENVIRONMENT=development
APP_DEBUG=true
DATABASE_HOST=localhost
REDIS_HOST=localhost
```

### Production

```bash
# Production environment
APP_ENVIRONMENT=production
APP_DEBUG=false
DATABASE_SSLMODE=require
JWT_SECRET=<secure-random-string>
```

## Docker Compose Environment

In Docker Compose, services are configured via environment:

```yaml
services:
  api:
    environment:
      - APP_ENVIRONMENT=production
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
```

## Secrets Management

For production, use a secrets manager:

```bash
# AWS Secrets Manager
export JWT_SECRET=$(aws secretsmanager get-secret-value --secret-id linkflow/jwt --query SecretString --output text)

# HashiCorp Vault
export JWT_SECRET=$(vault kv get -field=jwt_secret secret/linkflow)
```

## Validating Configuration

The application validates required configuration on startup:

```bash
# Missing JWT_SECRET will fail
./bin/api
# Error: JWT_SECRET is required

# Set required values
export JWT_SECRET=your-32-character-secret-key
./bin/api
# Server started on :8090
```

## Feature Flags

Some features can be enabled/disabled via configuration:

```yaml
features:
  oauth_enabled: true
  mfa_enabled: true
  billing_enabled: true
  websocket_enabled: true
```

Or via environment:

```bash
FEATURES_OAUTH_ENABLED=true
FEATURES_MFA_ENABLED=true
```
