# Incident Response Runbook

## Quick Commands
```bash
./scripts/status production      # Check status
./scripts/logs production api    # View logs
./scripts/rollback production    # Rollback
./scripts/scale production api 5 # Scale up
./scripts/backup production      # Backup before changes
```

## Severity Levels
- **P1**: System down (15 min response)
- **P2**: Major feature broken (1 hour)
- **P3**: Minor issue (4 hours)
- **P4**: Cosmetic (next day)

## Response Steps
1. Acknowledge alert
2. Assess severity
3. Communicate to team
4. Investigate
5. Mitigate (rollback/scale/restart)
6. Resolve
7. Post-incident report (within 48h)
