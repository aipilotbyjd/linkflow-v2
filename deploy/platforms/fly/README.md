# Fly.io Deployment

## Setup

### 1. Install Fly CLI
```bash
curl -L https://fly.io/install.sh | sh
fly auth login
```

### 2. Create Apps (one per environment)
```bash
fly apps create linkflow-dev
fly apps create linkflow-staging
fly apps create linkflow-production
```

### 3. Create Databases
```bash
# Dev
fly postgres create --name linkflow-dev-db --region sjc
fly postgres attach linkflow-dev-db --app linkflow-dev

# Staging
fly postgres create --name linkflow-staging-db --region sjc
fly postgres attach linkflow-staging-db --app linkflow-staging

# Production (with replicas)
fly postgres create --name linkflow-prod-db --region sjc --initial-cluster-size 2
fly postgres attach linkflow-prod-db --app linkflow-production
```

### 4. Create Redis
```bash
fly redis create linkflow-dev-redis --region sjc
fly redis create linkflow-staging-redis --region sjc
fly redis create linkflow-prod-redis --region sjc --plan 100
```

### 5. Set Secrets
```bash
fly secrets set \
  JWT_SECRET=$(openssl rand -hex 32) \
  ENCRYPTION_KEY=$(openssl rand -base64 24) \
  --app linkflow-production
```

## Deploy

```bash
# Deploy API
fly deploy -c deploy/platforms/fly/fly.toml --app linkflow-production

# Deploy Worker
fly deploy -c deploy/platforms/fly/fly.toml --app linkflow-production-worker --build-arg SERVICE=worker

# Deploy Scheduler
fly deploy -c deploy/platforms/fly/fly.toml --app linkflow-production-scheduler --build-arg SERVICE=scheduler
```

## Scale

```bash
# Scale API
fly scale count 3 --app linkflow-production

# Scale Workers
fly scale count 2 --app linkflow-production-worker
```

## Logs

```bash
fly logs --app linkflow-production
```

## Rollback

```bash
fly releases --app linkflow-production
fly deploy --image <previous-image> --app linkflow-production
```
