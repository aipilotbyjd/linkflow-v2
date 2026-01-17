# Database Down Runbook

## Quick Fix
```bash
# 1. Check status
./scripts/status production postgres

# 2. Restart
docker compose restart postgres

# 3. If still down, restore
./scripts/restore production latest
```

## Investigation
```bash
./scripts/logs production postgres
./scripts/shell production postgres
df -h  # Check disk space
```
