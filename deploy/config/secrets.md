# Secrets Management Guide

## Overview

**NEVER** commit secrets to git. Use platform secret management.

## By Platform

### Coolify
1. Go to your project → Environment Variables
2. Add each secret
3. Mark sensitive ones as "Secret" (hidden in UI)

### Fly.io
```bash
fly secrets set DB_PASSWORD=your-password --app linkflow-production
fly secrets set JWT_SECRET=your-jwt-secret --app linkflow-production
```

### Railway
1. Go to your project → Variables
2. Add secrets in the UI
3. They're automatically encrypted

### Render
1. Go to your service → Environment
2. Add secret environment variables
3. Mark as "Secret" to hide values

### AWS
Use AWS Secrets Manager or Parameter Store:
```bash
aws secretsmanager create-secret --name linkflow/production/db-password --secret-string "your-password"
```

### Kubernetes
```bash
kubectl create secret generic linkflow-secrets \
  --from-literal=DB_PASSWORD=your-password \
  --from-literal=JWT_SECRET=your-jwt-secret \
  -n linkflow-production
```

## Required Secrets

| Secret | Description | Rotate |
|--------|-------------|--------|
| `DB_PASSWORD` | Database password | 90 days |
| `REDIS_PASSWORD` | Redis password | 90 days |
| `JWT_SECRET` | JWT signing key (32+ chars) | 90 days |
| `ENCRYPTION_KEY` | Credential encryption (32 chars) | Never* |
| `STRIPE_SECRET_KEY` | Stripe API key | On compromise |
| `SMTP_PASSWORD` | Email password | 90 days |
| `S3_SECRET_KEY` | Storage access | 90 days |

*Rotating encryption key requires re-encrypting all credentials

## Generating Secrets

```bash
# JWT Secret (64 chars hex)
openssl rand -hex 32

# Encryption Key (32 chars)
openssl rand -base64 24

# Database Password
openssl rand -base64 32

# Random Password
openssl rand -base64 24 | tr -d '/+=' | head -c 32
```

## Rotation Procedure

See `security/secrets/rotate-*.sh` scripts.
