# Railway Deployment

## Setup

### 1. Install Railway CLI
```bash
npm install -g @railway/cli
railway login
```

### 2. Create Project
```bash
railway init
```

### 3. Add Services
In Railway dashboard:
1. Add PostgreSQL
2. Add Redis
3. Add API service (link to repo)
4. Add Worker service
5. Add Scheduler service

### 4. Configure Environment Variables
Set in Railway dashboard → Variables

## Deploy

Railway auto-deploys on git push. Manual deploy:
```bash
railway up
```

## Logs

```bash
railway logs
```

## Scale

Scale via Railway dashboard → Settings → Replicas
