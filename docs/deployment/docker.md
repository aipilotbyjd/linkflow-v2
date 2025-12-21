# Docker Deployment

Deploy LinkFlow using Docker and Docker Compose.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB+ RAM available

## Quick Start

```bash
# Clone repository
git clone https://github.com/linkflow-ai/linkflow.git
cd linkflow

# Configure environment
cp .env.example .env
nano .env  # Edit with production values

# Build and start
docker-compose up -d
```

## Production Docker Compose

Create `docker-compose.prod.yaml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_USER: ${DATABASE_USER}
      POSTGRES_PASSWORD: ${DATABASE_PASSWORD}
      POSTGRES_DB: ${DATABASE_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DATABASE_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 2G

  minio:
    image: minio/minio:latest
    restart: always
    environment:
      MINIO_ROOT_USER: ${S3_ACCESS_KEY_ID}
      MINIO_ROOT_PASSWORD: ${S3_SECRET_ACCESS_KEY}
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

  api:
    image: linkflow/api:${VERSION:-latest}
    build:
      context: .
      dockerfile: Dockerfile
      target: api
    restart: always
    ports:
      - "8090:8090"
    environment:
      - APP_ENVIRONMENT=production
      - APP_DEBUG=false
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=${DATABASE_USER}
      - DATABASE_PASSWORD=${DATABASE_PASSWORD}
      - DATABASE_NAME=${DATABASE_NAME}
      - DATABASE_SSLMODE=disable
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - S3_ENDPOINT=http://minio:9000
      - S3_ACCESS_KEY_ID=${S3_ACCESS_KEY_ID}
      - S3_SECRET_ACCESS_KEY=${S3_SECRET_ACCESS_KEY}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1'
          memory: 1G

  worker:
    image: linkflow/worker:${VERSION:-latest}
    build:
      context: .
      dockerfile: Dockerfile
      target: worker
    restart: always
    environment:
      - APP_ENVIRONMENT=production
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=${DATABASE_USER}
      - DATABASE_PASSWORD=${DATABASE_PASSWORD}
      - DATABASE_NAME=${DATABASE_NAME}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - WORKER_CONCURRENCY=10
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '2'
          memory: 1G

  scheduler:
    image: linkflow/scheduler:${VERSION:-latest}
    build:
      context: .
      dockerfile: Dockerfile
      target: scheduler
    restart: always
    environment:
      - APP_ENVIRONMENT=production
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=${DATABASE_USER}
      - DATABASE_PASSWORD=${DATABASE_PASSWORD}
      - DATABASE_NAME=${DATABASE_NAME}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.5'
          memory: 512M

  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certs:/etc/nginx/certs:ro
    depends_on:
      - api

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

## Nginx Configuration

Create `nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream api {
        least_conn;
        server api:8090;
    }

    server {
        listen 80;
        server_name api.linkflow.ai;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name api.linkflow.ai;

        ssl_certificate /etc/nginx/certs/fullchain.pem;
        ssl_certificate_key /etc/nginx/certs/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;

        # Security headers
        add_header Strict-Transport-Security "max-age=31536000" always;
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;

        location / {
            proxy_pass http://api;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeouts
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        location /ws {
            proxy_pass http://api;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_read_timeout 3600s;
        }

        location /health {
            proxy_pass http://api;
            access_log off;
        }
    }
}
```

## Commands

```bash
# Start services
docker-compose -f docker-compose.prod.yaml up -d

# View logs
docker-compose -f docker-compose.prod.yaml logs -f

# View specific service logs
docker-compose -f docker-compose.prod.yaml logs -f api

# Scale workers
docker-compose -f docker-compose.prod.yaml up -d --scale worker=5

# Stop services
docker-compose -f docker-compose.prod.yaml down

# Rebuild and restart
docker-compose -f docker-compose.prod.yaml up -d --build
```

## Building Images

```bash
# Build all images
docker-compose build

# Build specific image
docker-compose build api

# Build with specific tag
docker build -t linkflow/api:v1.0.0 --target api .

# Push to registry
docker push linkflow/api:v1.0.0
```

## Environment Variables

Create `.env` file:

```bash
# Version
VERSION=latest

# Database
DATABASE_USER=linkflow
DATABASE_PASSWORD=<secure-password>
DATABASE_NAME=linkflow

# Redis
REDIS_PASSWORD=<secure-password>

# JWT
JWT_SECRET=<32-character-secret>

# S3/MinIO
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=<secure-password>
```

## Monitoring

### View container stats

```bash
docker stats
```

### Health checks

```bash
# Check API health
curl http://localhost:8090/health

# Check all containers
docker-compose -f docker-compose.prod.yaml ps
```

## Backup and Restore

### PostgreSQL

```bash
# Backup
docker exec linkflow-postgres pg_dump -U linkflow linkflow > backup.sql

# Restore
cat backup.sql | docker exec -i linkflow-postgres psql -U linkflow linkflow
```

### Redis

```bash
# Backup (triggers RDB save)
docker exec linkflow-redis redis-cli -a $REDIS_PASSWORD BGSAVE

# Copy RDB file
docker cp linkflow-redis:/data/dump.rdb ./backup/
```

## Updating

```bash
# Pull latest images
docker-compose -f docker-compose.prod.yaml pull

# Restart with new images
docker-compose -f docker-compose.prod.yaml up -d

# Or rebuild from source
docker-compose -f docker-compose.prod.yaml up -d --build
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker-compose logs api

# Check container status
docker inspect linkflow-api
```

### Database connection issues

```bash
# Test database connection
docker exec -it linkflow-postgres psql -U linkflow -d linkflow -c "SELECT 1"
```

### Out of memory

```bash
# Check memory usage
docker stats --no-stream

# Increase limits in docker-compose.yaml
```
