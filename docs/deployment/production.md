# Production Checklist

Use this checklist before going live with LinkFlow.

## Security

### Authentication & Authorization
- [ ] Strong JWT secret (32+ random characters)
- [ ] Secure password hashing (bcrypt, cost 12+)
- [ ] Rate limiting enabled on auth endpoints
- [ ] Account lockout after failed attempts
- [ ] Session timeout configured
- [ ] API key rotation policy

### Network Security
- [ ] TLS 1.2+ enforced
- [ ] HTTP to HTTPS redirect
- [ ] Security headers configured (HSTS, CSP, X-Frame-Options)
- [ ] CORS properly configured
- [ ] Internal services not exposed publicly
- [ ] Database not accessible from internet

### Data Security
- [ ] Credential encryption (AES-256-GCM)
- [ ] Secrets in environment variables or secrets manager
- [ ] No secrets in logs
- [ ] PII handling compliant
- [ ] Backup encryption enabled

## Infrastructure

### Database (PostgreSQL)
- [ ] Production-grade hosting (RDS, Cloud SQL)
- [ ] SSL/TLS connections required
- [ ] Connection pooling (PgBouncer)
- [ ] Automated backups enabled
- [ ] Point-in-time recovery configured
- [ ] Read replicas for scaling
- [ ] Monitoring and alerting

### Redis
- [ ] Password authentication enabled
- [ ] Persistence enabled (AOF)
- [ ] Memory limits configured
- [ ] Sentinel or Cluster for HA
- [ ] TLS if over network

### Object Storage (S3/MinIO)
- [ ] Bucket policies configured
- [ ] Encryption at rest
- [ ] Lifecycle policies for cleanup
- [ ] Access logging enabled

## Application

### Configuration
- [ ] `APP_ENVIRONMENT=production`
- [ ] `APP_DEBUG=false`
- [ ] All required environment variables set
- [ ] Timeouts configured appropriately
- [ ] Worker concurrency tuned
- [ ] Queue priorities configured

### Logging
- [ ] Structured logging (JSON)
- [ ] Log aggregation (ELK, CloudWatch)
- [ ] Sensitive data not logged
- [ ] Log retention policy
- [ ] Error tracking (Sentry)

### Health Checks
- [ ] Liveness probe configured
- [ ] Readiness probe configured
- [ ] Health endpoints accessible
- [ ] Dependencies health checked

## Monitoring & Alerting

### Metrics
- [ ] Prometheus metrics exposed
- [ ] Grafana dashboards configured
- [ ] Key metrics tracked:
  - Request latency (p50, p95, p99)
  - Error rate
  - Queue depth
  - Active workers
  - Database connections
  - Memory/CPU usage

### Alerts
- [ ] High error rate (> 1%)
- [ ] High latency (> 2s p95)
- [ ] Queue backlog (> 1000 jobs)
- [ ] Service down
- [ ] Database connection issues
- [ ] Disk space low
- [ ] Memory usage high

### On-Call
- [ ] PagerDuty/OpsGenie configured
- [ ] Escalation policy defined
- [ ] Runbooks documented
- [ ] Incident response process

## Scalability

### Horizontal Scaling
- [ ] API: HPA based on CPU (70%)
- [ ] Worker: HPA based on queue depth
- [ ] Scheduler: Leader election working
- [ ] Load balancer configured
- [ ] Session affinity for WebSockets

### Performance
- [ ] Response time < 500ms p95
- [ ] Queue wait time < 30s
- [ ] Database query optimization
- [ ] Caching strategy implemented
- [ ] CDN for static assets

## Disaster Recovery

### Backups
- [ ] Database: Daily full, hourly incremental
- [ ] Redis: RDB snapshots, AOF
- [ ] Configuration backed up
- [ ] Backup restoration tested

### Recovery
- [ ] RTO documented (< 1 hour)
- [ ] RPO documented (< 5 minutes)
- [ ] Failover procedure documented
- [ ] Disaster recovery tested

### Multi-Region (if applicable)
- [ ] Read replicas in secondary region
- [ ] Failover DNS configured
- [ ] Data replication verified

## Compliance

### Documentation
- [ ] Architecture documented
- [ ] API documented (OpenAPI)
- [ ] Runbooks for common issues
- [ ] Change management process

### Audit
- [ ] Audit logging enabled
- [ ] Access logs retained
- [ ] Security audit performed
- [ ] Penetration testing done

## Deployment

### CI/CD
- [ ] Automated testing in pipeline
- [ ] Security scanning (SAST, dependencies)
- [ ] Staging environment for testing
- [ ] Blue-green or canary deployments
- [ ] Rollback procedure tested

### Release
- [ ] Version tagging
- [ ] Changelog maintained
- [ ] Feature flags for gradual rollout
- [ ] Communication plan for outages

## Pre-Launch

### Final Checks
- [ ] Load testing performed
- [ ] Security review completed
- [ ] Monitoring verified
- [ ] Alerts firing correctly
- [ ] Team trained on operations
- [ ] Support processes ready

### Launch Day
- [ ] Deployment window scheduled
- [ ] Team available for support
- [ ] Monitoring dashboards ready
- [ ] Rollback plan ready
- [ ] Communication channels open

## Post-Launch

### First 24 Hours
- [ ] Monitor error rates
- [ ] Monitor performance metrics
- [ ] Check all integrations working
- [ ] Verify scheduled jobs running
- [ ] Collect user feedback

### First Week
- [ ] Review incident reports
- [ ] Tune performance
- [ ] Address any issues
- [ ] Update documentation
- [ ] Post-mortem for any incidents
