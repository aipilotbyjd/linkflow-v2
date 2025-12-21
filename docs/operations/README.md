# Operations Guide

This guide covers operating LinkFlow in production environments.

## Overview

Operating LinkFlow involves:
- [Monitoring](#monitoring) - Track metrics and logs
- [Scaling](#scaling) - Handle increased load
- [Maintenance](#maintenance) - Regular upkeep tasks
- [Troubleshooting](#troubleshooting) - Debug issues

## Health Checks

All services expose health endpoints:

```bash
# Overall health
curl http://localhost:8090/health

# Liveness probe (is the process running?)
curl http://localhost:8090/health/live

# Readiness probe (can it accept traffic?)
curl http://localhost:8090/health/ready
```

### Response Format

```json
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "queue": "ok"
  },
  "version": "2.0.0",
  "uptime": "72h15m"
}
```

## Monitoring

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `http_request_duration_seconds` | API response time | p95 > 2s |
| `http_requests_total` | Request count by status | Error rate > 1% |
| `queue_depth` | Jobs waiting in queue | > 1000 |
| `queue_processing_time` | Job processing time | p95 > 60s |
| `active_workers` | Worker count | < min replicas |
| `db_connections_active` | Database connections | > 80% pool |
| `executions_total` | Execution count | - |
| `executions_failed_total` | Failed executions | > 5% |

### Prometheus Metrics

Metrics are exposed at `/metrics`:

```bash
curl http://localhost:8090/metrics
```

### Grafana Dashboards

Import dashboards from `docs/operations/grafana/`:
- `api-dashboard.json` - API metrics
- `worker-dashboard.json` - Worker metrics
- `queue-dashboard.json` - Queue metrics

### Log Aggregation

Logs are structured JSON:

```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "service": "api",
  "request_id": "abc123",
  "user_id": "user-uuid",
  "method": "POST",
  "path": "/api/v1/workflows",
  "status": 201,
  "duration_ms": 45
}
```

Query logs in your aggregator:
```
# High latency requests
level:info AND duration_ms:>1000

# Errors
level:error

# Specific user
user_id:"user-uuid"
```

## Scaling

### API Service

Scale based on CPU utilization:

```yaml
# Kubernetes HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70
```

### Worker Service

Scale based on queue depth:

```yaml
# KEDA ScaledObject
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
spec:
  minReplicaCount: 2
  maxReplicaCount: 50
  triggers:
  - type: redis
    metadata:
      listName: asynq:{default}
      listLength: "100"
```

### Scheduler Service

Only 2 replicas needed (leader election):
- Active: Processes scheduled jobs
- Standby: Takes over if active fails

See [Scaling Guide](scaling.md) for detailed strategies.

## Maintenance

### Database Maintenance

```bash
# Vacuum analyze (weekly)
VACUUM ANALYZE;

# Reindex (monthly, during low traffic)
REINDEX DATABASE linkflow;

# Check table bloat
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Redis Maintenance

```bash
# Check memory usage
redis-cli INFO memory

# Check slow queries
redis-cli SLOWLOG GET 10

# Background save
redis-cli BGSAVE
```

### Log Rotation

Configure log rotation:

```
/var/log/linkflow/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 linkflow linkflow
}
```

### Cleanup Jobs

The scheduler automatically runs cleanup jobs:

| Job | Frequency | Description |
|-----|-----------|-------------|
| Old executions | Daily 3 AM | Delete executions > retention |
| Dead letter queue | Every 10 min | Process failed jobs |
| Stale jobs | Every 5 min | Re-queue stuck jobs |
| Usage aggregation | Hourly | Calculate usage stats |

## Troubleshooting

### Common Issues

#### High API Latency

1. Check database connections:
```sql
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';
```

2. Check slow queries:
```sql
SELECT query, calls, mean_time, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC LIMIT 10;
```

3. Check Redis latency:
```bash
redis-cli --latency
```

#### Queue Backlog

1. Check queue depth:
```bash
redis-cli LLEN asynq:{default}
```

2. Scale workers:
```bash
kubectl scale deployment linkflow-worker --replicas=10
```

3. Check for failed jobs:
```bash
redis-cli LLEN asynq:{default}:dead
```

#### Worker Crashes

1. Check logs:
```bash
kubectl logs -f deployment/linkflow-worker
```

2. Check resource usage:
```bash
kubectl top pods -l app=linkflow-worker
```

3. Look for OOM kills:
```bash
kubectl describe pod <pod-name> | grep -A5 "Last State"
```

### Debug Commands

```bash
# API health
curl -s localhost:8090/health | jq .

# Queue stats
redis-cli KEYS "asynq:*" | head -20

# Database connections
docker exec linkflow-postgres psql -U postgres -c "SELECT * FROM pg_stat_activity"

# Worker processes
ps aux | grep worker
```

See [Troubleshooting Guide](troubleshooting.md) for more scenarios.

## Backups

### Database Backups

```bash
# Full backup
pg_dump -Fc linkflow > backup-$(date +%Y%m%d).dump

# Restore
pg_restore -d linkflow backup.dump
```

### Automated Backups

Configure automated backups:
- Daily full backups
- Continuous WAL archiving
- Retain for 30 days
- Test restoration monthly

### Redis Backups

```bash
# Manual snapshot
redis-cli BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb /backup/
```

## Security

### Rotate Credentials

1. Generate new credential
2. Update secret/environment
3. Rolling restart services
4. Revoke old credential

### Audit Logs

Enable audit logging for compliance:

```yaml
audit:
  enabled: true
  events:
    - user.login
    - workflow.execute
    - credential.access
```

### Security Updates

- Subscribe to security advisories
- Update dependencies monthly
- Run security scans in CI

## Incident Response

### Severity Levels

| Level | Description | Response Time |
|-------|-------------|---------------|
| P1 | Service down | 15 minutes |
| P2 | Degraded service | 1 hour |
| P3 | Minor issue | 4 hours |
| P4 | Low priority | Next business day |

### On-Call Playbook

1. Acknowledge alert
2. Assess severity
3. Start incident channel
4. Investigate and mitigate
5. Communicate status
6. Document and post-mortem

## Further Reading

- [Monitoring Guide](monitoring.md)
- [Scaling Guide](scaling.md)
- [Troubleshooting Guide](troubleshooting.md)
