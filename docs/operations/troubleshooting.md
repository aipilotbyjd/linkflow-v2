# Troubleshooting Guide

This guide helps diagnose and resolve common issues in LinkFlow.

## Quick Diagnostics

### Health Check

```bash
# Check all services
curl -s http://localhost:8090/health | jq .

# Expected output
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "queue": "ok"
  }
}
```

### Service Status

```bash
# Docker
docker-compose ps

# Kubernetes
kubectl get pods -n linkflow
```

## Common Issues

### API Issues

#### High Error Rate

**Symptoms:**
- 5xx responses increasing
- Alert: HighErrorRate

**Diagnosis:**
```bash
# Check error logs
grep "level\":\"error" /var/log/linkflow/api.log | tail -20

# Check recent errors by type
grep -o '"error":"[^"]*"' /var/log/linkflow/api.log | sort | uniq -c | sort -rn | head
```

**Common Causes:**
1. Database connection issues
2. Redis unavailable
3. Memory exhaustion
4. Downstream service failures

**Resolution:**
```bash
# Check database connectivity
psql -h $DATABASE_HOST -U $DATABASE_USER -c "SELECT 1"

# Check Redis connectivity
redis-cli -h $REDIS_HOST ping

# Check memory
docker stats --no-stream
```

#### High Latency

**Symptoms:**
- Slow API responses
- Alert: HighLatency (p95 > 2s)

**Diagnosis:**
```bash
# Check slow database queries
psql -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10"

# Check slow requests in logs
grep "duration_ms" /var/log/linkflow/api.log | awk -F'"duration_ms":' '{print $2}' | sort -n | tail -20
```

**Common Causes:**
1. Slow database queries
2. Missing indexes
3. Large payload processing
4. N+1 query problems

**Resolution:**
```bash
# Add missing indexes
psql -c "CREATE INDEX CONCURRENTLY idx_executions_workflow ON executions(workflow_id)"

# Analyze tables
psql -c "ANALYZE workflows, executions"
```

#### Connection Refused

**Symptoms:**
- `connection refused` errors
- Service unreachable

**Diagnosis:**
```bash
# Check if service is running
pgrep -f "linkflow-api"

# Check listening port
netstat -tlnp | grep 8090

# Check container status
docker logs linkflow-api --tail 50
```

**Resolution:**
```bash
# Restart service
docker-compose restart api

# Check for port conflicts
lsof -i :8090
```

### Worker Issues

#### Queue Backlog

**Symptoms:**
- Jobs accumulating in queue
- Alert: QueueBacklog

**Diagnosis:**
```bash
# Check queue depth
redis-cli LLEN "asynq:{default}"

# Check processing rate
redis-cli INFO | grep instantaneous_ops_per_sec

# Check worker count
docker-compose ps worker | wc -l
```

**Common Causes:**
1. Not enough workers
2. Slow job processing
3. External API rate limits
4. Worker crashes

**Resolution:**
```bash
# Scale workers
docker-compose up -d --scale worker=10

# Kubernetes
kubectl scale deployment linkflow-worker --replicas=10 -n linkflow

# Check for stuck jobs
redis-cli LRANGE "asynq:{default}:active" 0 -1
```

#### Worker Crashes

**Symptoms:**
- Workers restarting frequently
- OOMKilled status

**Diagnosis:**
```bash
# Check logs
docker logs linkflow-worker --tail 100

# Check memory usage
docker stats linkflow-worker

# Check for OOM
dmesg | grep -i "oom\|killed"
```

**Common Causes:**
1. Memory leaks
2. Large payloads
3. Infinite loops
4. Resource limits too low

**Resolution:**
```bash
# Increase memory limit
# docker-compose.yml
worker:
  deploy:
    resources:
      limits:
        memory: 2G

# Check for memory-heavy jobs
redis-cli LRANGE "asynq:{default}:dead" 0 10
```

#### Job Failures

**Symptoms:**
- Jobs in dead letter queue
- Execution failures

**Diagnosis:**
```bash
# Check dead letter queue
redis-cli LLEN "asynq:{default}:dead"

# Get failed job details
redis-cli LRANGE "asynq:{default}:dead" 0 5

# Check execution failures
psql -c "SELECT error_message, COUNT(*) FROM executions WHERE status='failed' GROUP BY error_message ORDER BY count DESC LIMIT 10"
```

**Resolution:**
```bash
# Retry failed jobs
# Via admin API
curl -X POST http://localhost:8090/admin/queue/retry-all

# Clear dead letter queue (careful!)
redis-cli DEL "asynq:{default}:dead"
```

### Database Issues

#### Connection Pool Exhaustion

**Symptoms:**
- `too many connections` errors
- Slow queries

**Diagnosis:**
```bash
# Check active connections
psql -c "SELECT count(*) FROM pg_stat_activity"

# Check connections by state
psql -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state"

# Check waiting queries
psql -c "SELECT * FROM pg_stat_activity WHERE wait_event IS NOT NULL"
```

**Resolution:**
```bash
# Kill idle connections
psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle' AND query_start < now() - interval '10 minutes'"

# Increase pool size
# config.yaml
database:
  max_open_conns: 50
  max_idle_conns: 10
```

#### Lock Contention

**Symptoms:**
- Queries hanging
- Deadlock errors

**Diagnosis:**
```bash
# Check locks
psql -c "SELECT * FROM pg_locks WHERE NOT granted"

# Check blocking queries
psql -c "SELECT blocked.pid, blocked.query, blocking.pid, blocking.query 
FROM pg_stat_activity AS blocked
JOIN pg_locks AS blocked_locks ON blocked.pid = blocked_locks.pid
JOIN pg_locks AS blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
JOIN pg_stat_activity AS blocking ON blocking.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted AND blocked.pid != blocking.pid"
```

**Resolution:**
```bash
# Kill blocking query
psql -c "SELECT pg_cancel_backend(<pid>)"

# Or force terminate
psql -c "SELECT pg_terminate_backend(<pid>)"
```

### Redis Issues

#### Memory Full

**Symptoms:**
- `OOM command not allowed`
- Jobs not being queued

**Diagnosis:**
```bash
# Check memory usage
redis-cli INFO memory

# Check keys by size
redis-cli --bigkeys
```

**Resolution:**
```bash
# Clear completed jobs (older than 24h)
redis-cli EVAL "local keys = redis.call('keys', 'asynq:{default}:completed:*') for i=1,#keys do redis.call('del', keys[i]) end" 0

# Set memory limit with eviction
redis-cli CONFIG SET maxmemory 2gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

#### Connection Issues

**Symptoms:**
- `READONLY` errors
- Connection timeouts

**Diagnosis:**
```bash
# Check Redis status
redis-cli PING
redis-cli INFO replication

# Check connection count
redis-cli CLIENT LIST | wc -l
```

**Resolution:**
```bash
# Restart Redis
docker-compose restart redis

# Check Sentinel status (if using)
redis-cli -p 26379 SENTINEL masters
```

### Scheduler Issues

#### Schedules Not Running

**Symptoms:**
- Cron jobs not executing
- `next_run_at` not updating

**Diagnosis:**
```bash
# Check scheduler leader
redis-cli GET "scheduler:leader"

# Check scheduler logs
docker logs linkflow-scheduler --tail 50

# Check schedule records
psql -c "SELECT id, cron_expression, next_run_at, is_active FROM schedules LIMIT 10"
```

**Resolution:**
```bash
# Force leader re-election
redis-cli DEL "scheduler:leader"

# Restart scheduler
docker-compose restart scheduler

# Update next_run_at manually
psql -c "UPDATE schedules SET next_run_at = NOW() WHERE is_active = true AND next_run_at < NOW()"
```

## Emergency Procedures

### Service Rollback

```bash
# Docker
docker-compose down
docker-compose -f docker-compose.previous.yaml up -d

# Kubernetes
kubectl rollout undo deployment/linkflow-api -n linkflow
```

### Database Recovery

```bash
# Point-in-time recovery
pg_restore -d linkflow_new backup.dump

# Verify data
psql -d linkflow_new -c "SELECT COUNT(*) FROM workflows"

# Swap databases
psql -c "ALTER DATABASE linkflow RENAME TO linkflow_old"
psql -c "ALTER DATABASE linkflow_new RENAME TO linkflow"
```

### Clear Queue

```bash
# WARNING: This deletes all pending jobs
redis-cli FLUSHDB

# Or clear specific queue
redis-cli DEL "asynq:{default}"
```

## Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
# Set environment
APP_DEBUG=true LOG_LEVEL=debug

# Restart services
docker-compose restart

# View debug logs
docker logs -f linkflow-api 2>&1 | jq .
```

## Getting Help

If issues persist:

1. Collect diagnostic info:
```bash
# Run diagnostic script
./scripts/collect-diagnostics.sh > diagnostics.txt
```

2. Check documentation:
   - [FAQ](../guides/faq.md)
   - [Known Issues](https://github.com/linkflow-ai/linkflow/issues)

3. Contact support:
   - Email: support@linkflow.ai
   - Slack: #linkflow-support
