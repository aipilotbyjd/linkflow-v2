# High CPU Runbook

## Quick Fix
```bash
# Scale up immediately
./scripts/scale production api 5

# Or rollback recent deploy
./scripts/rollback production
```

## Investigation
```bash
./scripts/logs production api
./scripts/shell production api
top -H  # Find CPU-heavy process
```
