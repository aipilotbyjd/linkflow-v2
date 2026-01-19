# Koyeb Deployment (Free Tier - No Credit Card)

## Overview

Koyeb offers a free tier with:
- 2 nano services (512MB RAM each)
- Auto-deploy from GitHub
- Free SSL
- No credit card required

**Use with:**
- **Neon** - Free PostgreSQL (neon.tech)
- **Upstash** - Free Redis (upstash.com)

## Quick Setup

### 1. Create External Services (Free, No Credit Card)

**Neon (PostgreSQL):**
1. Go to [neon.tech](https://neon.tech)
2. Sign up with GitHub
3. Create project → Copy connection string
```
postgresql://user:pass@ep-xxx.us-east-1.aws.neon.tech/neondb?sslmode=require
```

**Upstash (Redis):**
1. Go to [upstash.com](https://upstash.com)
2. Sign up with GitHub
3. Create database → Copy Redis URL
```
rediss://default:xxx@xxx.upstash.io:6379
```

### 2. Deploy to Koyeb

**Option A: Web Dashboard (Recommended)**

1. Go to [koyeb.com](https://koyeb.com)
2. Sign up with GitHub (no credit card)
3. Click **Create App**
4. Select **GitHub** → Choose `linkflow-v2` repo
5. Configure:
   - **Builder:** Dockerfile
   - **Dockerfile path:** `deploy/docker/Dockerfile`
   - **Build args:** `SERVICE=api`
   - **Port:** 8090
   - **Instance:** Free
   - **Region:** Frankfurt (fra)
6. Add environment variables (see below)
7. Click **Deploy**

**Option B: CLI**

```bash
# Install Koyeb CLI
curl -fsSL https://raw.githubusercontent.com/koyeb/koyeb-cli/master/install.sh | sh

# Login
koyeb login

# Deploy
koyeb app create linkflow-api \
  --docker deploy/docker/Dockerfile \
  --docker-args SERVICE=api \
  --ports 8090:http \
  --routes /:8090 \
  --instance-type free \
  --regions fra \
  --env DATABASE_URL="postgresql://..." \
  --env REDIS_URL="rediss://..." \
  --env JWT_SECRET="your-secret"
```

### 3. Environment Variables

Set these in Koyeb dashboard → App → Settings → Environment:

```bash
# App
APP_ENVIRONMENT=production
APP_URL=https://linkflow-api-<your-org>.koyeb.app
PORT=8090

# Database (from Neon)
DATABASE_URL=postgresql://user:pass@ep-xxx.neon.tech/neondb?sslmode=require

# Redis (from Upstash)  
REDIS_URL=rediss://default:xxx@xxx.upstash.io:6379

# Security (generate secure random strings)
JWT_SECRET=<32+ character random string>
ENCRYPTION_KEY=<exactly 32 characters>

# OAuth (optional - from Google Cloud Console)
OAUTH_GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
OAUTH_GOOGLE_CLIENT_SECRET=GOCSPX-xxx
OAUTH_GOOGLE_REDIRECT_URL=https://linkflow-api-<your-org>.koyeb.app/api/v1/auth/oauth/google/callback
```

### 4. Google OAuth Setup (Optional)

Add redirect URI in [Google Cloud Console](https://console.cloud.google.com/apis/credentials):
```
https://linkflow-api-<your-org>.koyeb.app/api/v1/auth/oauth/google/callback
```

## Costs

| Service | Free Tier | Limits |
|---------|-----------|--------|
| Koyeb | Free | 2 nano instances |
| Neon | Free | 0.5GB storage, always free |
| Upstash | Free | 10K commands/day |

**Total: $0/month**

## Limitations

- Nano instances have 512MB RAM
- App may sleep after inactivity (cold starts)
- Limited to 2 services on free tier

## Scaling (When Needed)

Upgrade to paid tiers:
- Koyeb: Starter ($5.50/month)
- Neon: Pro ($19/month)
- Upstash: Pay-as-you-go

## Troubleshooting

**Build fails:**
```bash
# Check build logs in Koyeb dashboard
# Ensure Dockerfile path is correct: deploy/docker/Dockerfile
```

**Health check fails:**
```bash
# Verify the app starts on port 8090
# Check DATABASE_URL and REDIS_URL are correct
```

**Cold starts:**
- Free tier may sleep after inactivity
- First request after sleep takes ~10-30s
- Upgrade to paid tier for always-on

## Alternative Free Stacks

If Koyeb doesn't work:

| Stack | API | Database | Redis |
|-------|-----|----------|-------|
| Render + Neon + Upstash | Render (free) | Neon (free) | Upstash (free) |
| Koyeb + Supabase + Upstash | Koyeb (free) | Supabase (free) | Upstash (free) |
