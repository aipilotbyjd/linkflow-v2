# Monitoring Guide

This guide covers setting up monitoring for LinkFlow.

## Metrics

### Prometheus Setup

LinkFlow exposes Prometheus metrics at `/metrics`.

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'linkflow-api'
    static_configs:
      - targets: ['api:8090']
    metrics_path: /metrics
    
  - job_name: 'linkflow-worker'
    static_configs:
      - targets: ['worker:8091']
    metrics_path: /metrics
```

### Key Metrics

#### API Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | Request duration |
| `http_requests_in_flight` | Gauge | Active requests |
| `websocket_connections` | Gauge | Active WebSocket connections |

#### Worker Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `worker_jobs_processed_total` | Counter | Jobs processed |
| `worker_job_duration_seconds` | Histogram | Job processing time |
| `worker_jobs_failed_total` | Counter | Failed jobs |
| `worker_active_jobs` | Gauge | Currently processing |

#### Queue Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `queue_depth` | Gauge | Jobs waiting |
| `queue_latency_seconds` | Histogram | Time in queue |
| `dead_letter_queue_size` | Gauge | Failed jobs |

#### Database Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `db_connections_open` | Gauge | Open connections |
| `db_connections_in_use` | Gauge | Active connections |
| `db_query_duration_seconds` | Histogram | Query duration |

### Grafana Dashboards

#### API Dashboard

```json
{
  "panels": [
    {
      "title": "Request Rate",
      "type": "graph",
      "targets": [{
        "expr": "rate(http_requests_total[5m])"
      }]
    },
    {
      "title": "Response Time (p95)",
      "type": "graph",
      "targets": [{
        "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
      }]
    },
    {
      "title": "Error Rate",
      "type": "graph",
      "targets": [{
        "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])"
      }]
    }
  ]
}
```

#### Worker Dashboard

```json
{
  "panels": [
    {
      "title": "Jobs per Second",
      "targets": [{
        "expr": "rate(worker_jobs_processed_total[5m])"
      }]
    },
    {
      "title": "Queue Depth",
      "targets": [{
        "expr": "queue_depth"
      }]
    },
    {
      "title": "Processing Time (p95)",
      "targets": [{
        "expr": "histogram_quantile(0.95, rate(worker_job_duration_seconds_bucket[5m]))"
      }]
    }
  ]
}
```

## Alerting

### Alert Rules

```yaml
# alerts.yml
groups:
- name: linkflow
  rules:
  # High error rate
  - alert: HighErrorRate
    expr: |
      rate(http_requests_total{status=~"5.."}[5m]) 
      / rate(http_requests_total[5m]) > 0.01
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High error rate ({{ $value | humanizePercentage }})
      
  # High latency
  - alert: HighLatency
    expr: |
      histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: High p95 latency ({{ $value | humanizeDuration }})
      
  # Queue backlog
  - alert: QueueBacklog
    expr: queue_depth > 1000
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: Queue depth is {{ $value }}
      
  # Worker down
  - alert: WorkerDown
    expr: up{job="linkflow-worker"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: Worker instance down
      
  # Database connections
  - alert: HighDBConnections
    expr: |
      db_connections_in_use / db_connections_open > 0.8
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: Database connection pool > 80%
```

### PagerDuty Integration

```yaml
# alertmanager.yml
receivers:
  - name: pagerduty
    pagerduty_configs:
      - service_key: <your-service-key>
        severity: '{{ .GroupLabels.severity }}'
        description: '{{ .CommonAnnotations.summary }}'

route:
  receiver: pagerduty
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  routes:
    - match:
        severity: critical
      receiver: pagerduty
```

## Logging

### Log Format

All services use structured JSON logging:

```json
{
  "level": "info",
  "time": "2024-01-15T10:30:00Z",
  "service": "api",
  "version": "2.0.0",
  "request_id": "req-abc123",
  "trace_id": "trace-xyz789",
  "user_id": "user-uuid",
  "workspace_id": "workspace-uuid",
  "method": "POST",
  "path": "/api/v1/workflows",
  "status": 201,
  "duration_ms": 45,
  "bytes": 1234
}
```

### Log Levels

| Level | Usage |
|-------|-------|
| `debug` | Detailed debugging information |
| `info` | Normal operations |
| `warn` | Unexpected but handled situations |
| `error` | Errors requiring attention |
| `fatal` | Unrecoverable errors |

### ELK Stack Setup

```yaml
# filebeat.yml
filebeat.inputs:
  - type: container
    paths:
      - '/var/lib/docker/containers/*/*.log'
    processors:
      - add_kubernetes_metadata: ~
      - decode_json_fields:
          fields: ["message"]
          target: ""

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "linkflow-%{+yyyy.MM.dd}"
```

### Useful Log Queries

```
# Errors in last hour
level:error AND @timestamp:[now-1h TO now]

# Slow requests
duration_ms:>1000 AND level:info

# Specific workflow executions
workflow_id:"uuid" AND level:*

# Failed job processing
service:worker AND level:error

# Authentication failures
path:"/api/v1/auth/login" AND status:401
```

## Distributed Tracing

### OpenTelemetry Setup

```go
// Initialize tracer
tp := trace.NewTracerProvider(
    trace.WithBatcher(exporter),
    trace.WithResource(resource.NewWithAttributes(
        semconv.ServiceNameKey.String("linkflow-api"),
    )),
)
otel.SetTracerProvider(tp)
```

### Trace Context

Traces flow through:
1. API receives request (trace starts)
2. API enqueues job (span)
3. Worker processes job (child span)
4. Worker executes nodes (child spans)
5. Worker completes (span ends)

### Jaeger Setup

```yaml
# docker-compose.yml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"  # UI
    - "6831:6831/udp"  # Thrift
```

## Health Monitoring

### Uptime Monitoring

Configure external uptime monitoring:

```yaml
# Better Uptime / Pingdom config
checks:
  - name: LinkFlow API
    url: https://api.linkflow.ai/health
    interval: 60
    regions: [us-east, eu-west, ap-south]
    
  - name: LinkFlow Health
    url: https://api.linkflow.ai/health/ready
    interval: 30
    expected_status: 200
```

### Synthetic Monitoring

Create synthetic tests for critical flows:

```javascript
// Example: Workflow execution synthetic test
const response = await fetch('/api/v1/workflows/test-id/execute', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
});

if (response.status !== 202) {
  throw new Error('Workflow execution failed');
}
```

## Dashboards

### SLO Dashboard

Track Service Level Objectives:

| SLO | Target | Current |
|-----|--------|---------|
| Availability | 99.9% | 99.95% |
| API Latency (p95) | < 500ms | 245ms |
| Error Rate | < 0.1% | 0.02% |
| Queue Wait Time | < 30s | 5s |

### Business Metrics

| Metric | Description |
|--------|-------------|
| Active users (DAU) | Daily active users |
| Workflows created | New workflows per day |
| Executions | Total executions per day |
| Execution success rate | Completed / Total |
| Revenue | MRR, ARR tracking |

## Runbooks

Link alerts to runbooks:

| Alert | Runbook |
|-------|---------|
| HighErrorRate | [Investigating High Error Rates](troubleshooting.md#high-error-rate) |
| QueueBacklog | [Queue Backlog Resolution](troubleshooting.md#queue-backlog) |
| WorkerDown | [Worker Recovery](troubleshooting.md#worker-down) |
| HighLatency | [Performance Investigation](troubleshooting.md#high-latency) |
