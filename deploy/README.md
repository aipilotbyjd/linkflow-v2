# Deployment

## Coolify (Production)

1. Add Docker Compose application in Coolify
2. Docker Compose path: `deploy/docker-compose.yml`
3. Set environment variables:

   | Variable | Generate |
   |----------|----------|
   | `POSTGRES_PASSWORD` | `openssl rand -base64 32` |
   | `JWT_SECRET` | `openssl rand -hex 32` |
   | `APP_URL` | `https://api.yourdomain.com` |
   | `APP_FRONTEND_URL` | `https://app.yourdomain.com` |

4. Configure domain for `api` service on port `8090`
5. Deploy

## Local Development

```bash
# From project root
docker compose -f deploy/docker-compose.dev.yml up -d
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| api | 8090 | HTTP API |
| worker | - | Background jobs |
| scheduler | - | Cron jobs |
| postgres | 5432 | Database |
| redis | 6379 | Cache & Queue |
