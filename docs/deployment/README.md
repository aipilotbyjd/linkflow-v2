# Deployment Guide

This guide covers deploying LinkFlow to production environments.

## Deployment Options

| Method | Best For |
|--------|----------|
| [Docker Compose](docker.md) | Single server, small teams |
| [Kubernetes](kubernetes.md) | Large scale, high availability |
| [Manual](production.md) | Custom infrastructure |

## Quick Start - Docker Compose

```bash
# Clone repository
git clone https://github.com/linkflow-ai/linkflow.git
cd linkflow

# Configure environment
cp .env.example .env
# Edit .env with production values

# Start services
docker-compose -f docker-compose.prod.yaml up -d
```

## Architecture Requirements

### Minimum Requirements (Small)

| Component | Specification |
|-----------|--------------|
| API | 1 instance, 1 vCPU, 512MB |
| Worker | 2 instances, 1 vCPU, 512MB each |
| Scheduler | 1 instance, 0.5 vCPU, 256MB |
| PostgreSQL | 2 vCPU, 4GB RAM, 50GB SSD |
| Redis | 1 vCPU, 2GB RAM |

### Recommended (Production)

| Component | Specification |
|-----------|--------------|
| API | 2-4 instances, 2 vCPU, 1GB each |
| Worker | 5-20 instances, 2 vCPU, 1GB each |
| Scheduler | 2 instances (HA), 1 vCPU, 512MB |
| PostgreSQL | 4 vCPU, 16GB RAM, 200GB SSD |
| Redis | 2 vCPU, 8GB RAM |
| Load Balancer | nginx, AWS ALB, or similar |

## Infrastructure Components

### PostgreSQL

- Version 15+ recommended
- Enable connection pooling (PgBouncer)
- Configure automatic backups
- Set up read replicas for scaling

### Redis

- Version 7+ recommended
- Enable persistence (AOF)
- Configure memory limits
- Set up Sentinel for HA

### Load Balancer

- SSL/TLS termination
- Health check endpoints
- WebSocket support
- Rate limiting

### Object Storage

- S3 or MinIO
- For file uploads and exports
- Configure lifecycle policies

## Environment Configuration

### Required Variables

```bash
# Security
JWT_SECRET=<32-character-random-string>

# Database
DATABASE_HOST=postgres.example.com
DATABASE_PORT=5432
DATABASE_USER=linkflow
DATABASE_PASSWORD=<secure-password>
DATABASE_NAME=linkflow
DATABASE_SSLMODE=require

# Redis
REDIS_HOST=redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=<redis-password>

# Application
APP_ENVIRONMENT=production
APP_DEBUG=false
APP_URL=https://api.linkflow.ai
```

### Optional Variables

```bash
# Stripe (for billing)
STRIPE_SECRET_KEY=sk_live_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx

# OAuth
OAUTH_GOOGLE_CLIENT_ID=xxx
OAUTH_GOOGLE_CLIENT_SECRET=xxx

# Email
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=<sendgrid-api-key>
```

## Deployment Checklist

### Pre-Deployment

- [ ] Configure production environment variables
- [ ] Set strong JWT secret (32+ characters)
- [ ] Enable database SSL
- [ ] Configure backup strategy
- [ ] Set up monitoring and alerting
- [ ] Configure rate limiting
- [ ] Review security headers

### Deployment

- [ ] Build Docker images
- [ ] Run database migrations
- [ ] Deploy API service
- [ ] Deploy Worker service
- [ ] Deploy Scheduler service
- [ ] Verify health endpoints
- [ ] Test critical workflows

### Post-Deployment

- [ ] Verify logs are flowing
- [ ] Check metrics dashboards
- [ ] Test alerting
- [ ] Document rollback procedure

## Health Checks

All services expose health endpoints:

```bash
# Overall health
curl https://api.linkflow.ai/health

# Liveness probe (is the process running?)
curl https://api.linkflow.ai/health/live

# Readiness probe (can it accept traffic?)
curl https://api.linkflow.ai/health/ready
```

## Scaling

### Horizontal Scaling

- **API**: Scale based on CPU (target 70%)
- **Worker**: Scale based on queue depth
- **Scheduler**: 2 instances max (leader election)

### Vertical Scaling

Increase resources when:
- Response times > 500ms p95
- Queue wait time > 30 seconds
- Memory usage > 80%

See [Scaling Guide](../operations/scaling.md) for details.

## Database Migrations

Migrations run automatically on API startup. For manual control:

```bash
# Run migrations
./bin/api migrate up

# Rollback
./bin/api migrate down 1

# Check status
./bin/api migrate status
```

## Rollback Procedure

1. Stop new deployments
2. Scale down new version
3. Deploy previous version
4. Verify health
5. Investigate issue

```bash
# Docker Compose
docker-compose down
docker-compose -f docker-compose.previous.yaml up -d

# Kubernetes
kubectl rollout undo deployment/linkflow-api
```

## SSL/TLS Configuration

### Using Let's Encrypt with nginx

```nginx
server {
    listen 443 ssl http2;
    server_name api.linkflow.ai;
    
    ssl_certificate /etc/letsencrypt/live/api.linkflow.ai/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.linkflow.ai/privkey.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    
    location / {
        proxy_pass http://linkflow-api:8090;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /ws {
        proxy_pass http://linkflow-api:8090;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## Backup Strategy

### PostgreSQL

```bash
# Daily full backup
pg_dump -Fc linkflow > backup-$(date +%Y%m%d).dump

# Point-in-time recovery with WAL
archive_mode = on
archive_command = 'aws s3 cp %p s3://backups/wal/%f'
```

### Redis

```bash
# RDB snapshots
save 900 1
save 300 10
save 60 10000

# AOF persistence
appendonly yes
appendfsync everysec
```

## Security Hardening

- [ ] Use non-root containers
- [ ] Enable security headers (HSTS, CSP)
- [ ] Configure network policies
- [ ] Enable audit logging
- [ ] Rotate credentials regularly
- [ ] Use secrets management (Vault, AWS SM)

## Further Reading

- [Docker Deployment](docker.md)
- [Kubernetes Deployment](kubernetes.md)
- [Production Checklist](production.md)
- [Monitoring Guide](../operations/monitoring.md)
