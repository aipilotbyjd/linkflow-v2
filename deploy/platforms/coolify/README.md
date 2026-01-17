# Coolify Deployment (Self-Hosted)

## Setup Coolify on Hetzner

### 1. Create Hetzner Server
- Go to hetzner.com
- Create server: CX21 (2 vCPU, 4GB RAM) - €4.50/month
- Select Ubuntu 22.04
- Add SSH key

### 2. Install Coolify
```bash
ssh root@your-server-ip
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

### 3. Access Coolify
- Go to `http://your-server-ip:8000`
- Create admin account
- Add your server as destination

### 4. Add LinkFlow Project
1. New Project → Add Resource → Docker Compose
2. Point to: `deploy/compose/docker-compose.yml`
3. Set environment variables

## Environment Setup

### Dev
1. Create new environment "Development"
2. Set `ENV=dev` in environment variables
3. Add all secrets from `deploy/config/dev.env.example`

### Staging
1. Create new environment "Staging"
2. Set `ENV=staging`
3. Add all secrets

### Production
1. Create new environment "Production"
2. Set `ENV=production`
3. Add all secrets
4. Enable auto-deploy on main branch

## Zero Downtime

Coolify handles this automatically:
1. Builds new container
2. Health check passes
3. Switches traffic
4. Stops old container

## Backups

### Database Backup (Automated)
1. Go to your PostgreSQL service
2. Enable scheduled backups
3. Set retention period

### Manual Backup
```bash
ssh root@your-server
docker exec linkflow-postgres pg_dump -U linkflow linkflow > backup.sql
```

## Scaling

For scaling beyond single server:
1. Add more Hetzner servers
2. Add them to Coolify
3. Deploy services across servers

## SSL

Coolify auto-configures SSL via Let's Encrypt:
1. Point your domain to server IP
2. Add domain in Coolify
3. SSL auto-provisioned

## Monitoring

Add monitoring stack:
1. New Resource → Docker Compose
2. Use `deploy/compose/docker-compose.monitoring.yml`
