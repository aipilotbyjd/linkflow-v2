# Deployment Failed Runbook

## Quick Fix
```bash
# Rollback immediately
./scripts/rollback production

# Check logs
./scripts/logs production api
```

## Common Issues
- **ImagePullBackOff**: Check registry credentials
- **OOMKilled**: Increase memory limits
- **CrashLoopBackOff**: Check app logs
- **Readiness failed**: Check health endpoint
