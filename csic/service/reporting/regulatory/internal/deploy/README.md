# Regulatory Reporting Service - Deployment Guide

This document provides comprehensive instructions for deploying and operating the CSIC Regulatory Reporting Service.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Deployment Options](#deployment-options)
- [Monitoring](#monitoring)
- [Maintenance](#maintenance)
- [Troubleshooting](#troubleshooting)

## Overview

The Regulatory Reporting Service is a critical component of the CSIC platform that handles:

- **Report Generation**: PDF and CSV exports for regulatory compliance
- **Scheduled Reporting**: Automated report scheduling with BullMQ
- **WORM Storage**: Write Once Read Many compliant storage for audit trails
- **Event Streaming**: Kafka-based event processing for real-time updates

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Regulatory Reporting Service              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   REST API  │  │  Job Queue  │  │   Event Processor   │  │
│  │  (Express)  │  │  (BullMQ)   │  │     (Kafka)         │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
│         │                │                     │              │
│         ▼                ▼                     ▼              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   Data Layer                         │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐  │    │
│  │  │PostgreSQL│ │  Redis  │ │  Kafka  │ │  MinIO   │  │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └──────────┘  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

### System Requirements

- **CPU**: 2 cores minimum (4 cores recommended)
- **Memory**: 4GB minimum (8GB recommended)
- **Storage**: 50GB minimum (100GB recommended)
- **Docker**: Version 20.10+ 
- **Docker Compose**: Version 2.0+

### Required Services

| Service | Default Port | Purpose |
|---------|-------------|---------|
| PostgreSQL | 5433 | Report metadata storage |
| Redis | 6379 | Job queue backend |
| Kafka | 9092 | Event streaming |
| MinIO | 9000 | S3-compatible storage |
| Prometheus | 9090 | Metrics collection |
| Grafana | 3002 | Visualization |
| Reporting API | 3001 | REST API |
| Metrics | 9091 | Prometheus metrics |

## Quick Start

### 1. Clone and Navigate

```bash
cd csic-platform/service/reporting/regulatory
```

### 2. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

### 3. Start Services

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f reporting-api
```

### 4. Verify Installation

```bash
# Check service health
curl http://localhost:3001/health

# Check metrics
curl http://localhost:9091/metrics
```

## Configuration

### Environment Variables

#### Application Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `NODE_ENV` | production | Environment mode |
| `PORT` | 3001 | API server port |
| `LOG_LEVEL` | info | Logging level (debug, info, warn, error) |

#### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | reporting-db | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_NAME` | csic_reporting | Database name |
| `DB_USER` | reporting_admin | Database user |
| `DB_PASS` | (required) | Database password |

#### Redis Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | reporting-redis | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `REDIS_PASSWORD` | "" | Redis password (optional) |

#### Kafka Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `KAFKA_BROKERS` | kafka:9092 | Kafka broker list |
| `KAFKA_CLIENT_ID` | csic-reporting-service | Client identifier |
| `KAFKA_GROUP_ID` | csic-reporting-group | Consumer group ID |

#### Storage Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `S3_ENDPOINT` | http://minio:9000 | S3 endpoint URL |
| `S3_ACCESS_KEY` | minioadmin | S3 access key |
| `S3_SECRET_KEY` | (required) | S3 secret key |
| `S3_BUCKET` | csic-reports | S3 bucket name |
| `WORM_ENABLED` | true | Enable WORM storage |
| `WORM_RETENTION_DAYS` | 2555 | WORM retention period |

### Custom Configuration File

Create a `config.yaml` file for advanced configuration:

```yaml
app:
  name: "Regulatory Reporting Service"
  host: "0.0.0.0"
  port: 3001

database:
  host: "reporting-db"
  port: 5432
  name: "csic_reporting"
  pool:
    min: 2
    max: 10

queue:
  redis:
    host: "reporting-redis"
    port: 6379
  jobs:
    concurrency: 5
    retryAttempts: 3
    retryDelay: 5000

storage:
  type: "s3"
  s3:
    endpoint: "http://minio:9000"
    bucket: "csic-reports"
  worm:
    enabled: true
    retentionDays: 2555

monitoring:
  enabled: true
  port: 9091
  path: "/metrics"
```

## Deployment Options

### Development Deployment

```bash
# Start with development settings
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

### Production Deployment

```bash
# Start with production settings
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### High Availability Deployment

For production environments, deploy each service on separate hosts:

```bash
# Database server
docker-compose -f docker-compose.db.yml up -d

# Message queue server
docker-compose -f docker-compose.queue.yml up -d

# Application servers (multiple replicas)
docker-compose -f docker-compose.app.yml up -d --scale reporting-api=3
```

### Kubernetes Deployment

For Kubernetes deployment, use the provided Helm charts:

```bash
helm install csic-reporting ./helm/reporting/
```

## Monitoring

### Accessing Grafana Dashboards

1. Navigate to `http://localhost:3002`
2. Login with credentials from `.env` (default: admin/grafana_secure_2024)
3. Navigate to "CSIC / Reporting" dashboard

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `up` | Service availability | < 1 |
| `http_request_duration_seconds` | API latency | p95 > 2s |
| `report_generation_duration_seconds` | Report generation time | p95 > 300s |
| `bullmq_queue_jobs_total{status="waiting"}` | Job queue backlog | > 1000 |
| `bullmq_queue_jobs_total{status="failed"}` | Failed jobs | > 10 |

### Prometheus Alerts

The following alerts are configured:

- **ReportingAPIDown**: Service is unreachable
- **ReportingAPIHighErrorRate**: > 5% error rate
- **ReportingDatabaseDown**: Database is unreachable
- **ReportingJobQueueBacklogHigh**: > 1000 pending jobs
- **ReportingGenerationFailureRateHigh**: > 10% failure rate

### Accessing Prometheus

Navigate to `http://localhost:9090` for:

- Query browser
- Status endpoints
- Alertmanager configuration

## Maintenance

### Backup Procedures

#### Database Backup

```bash
# Create backup
docker exec csic-reporting-db pg_dump -U reporting_admin csic_reporting > backup.sql

# Automated backup (add to crontab)
0 2 * * * docker exec csic-reporting-db pg_dump -U reporting_admin csic_reporting | gzip > /backup/report_$(date +\%Y\%m\%d).sql.gz
```

#### MinIO Bucket Backup

```bash
# Sync to local directory
mc alias set local http://localhost:9000 minioadmin minioadmin_secure_2024
mc mirror local/csic-reports /backup/minio/
```

### Scaling

#### Horizontal Scaling

```bash
# Scale API instances
docker-compose up -d --scale reporting-api=4

# Configure load balancer (nginx.conf)
upstream reporting_api {
    server reporting-api-1:3001;
    server reporting-api-2:3001;
    server reporting-api-3:3001;
    server reporting-api-4:3001;
}
```

#### Vertical Scaling

Edit resource limits in `docker-compose.yml`:

```yaml
deploy:
  resources:
    limits:
      cpus: '4'
      memory: 4G
    reservations:
      cpus: '1'
      memory: 1G
```

### Updates and Rollbacks

#### Update Procedure

```bash
# Pull latest images
docker-compose pull

# Backup database
docker exec csic-reporting-db pg_dump -U reporting_admin csic_reporting > backup.sql

# Rollout update
docker-compose up -d

# Verify
curl http://localhost:3001/health
```

#### Rollback Procedure

```bash
# Previous version
docker-compose down

# Restore database
docker exec -i csic-reporting-db psql -U reporting_admin csic_reporting < backup.sql

# Deploy previous version
git checkout v1.0.0
docker-compose up -d
```

## Troubleshooting

### Common Issues

#### Service Won't Start

```bash
# Check logs
docker-compose logs reporting-api

# Verify ports
netstat -tlnp | grep -E '(3001|5433|6379|9090)'

# Check disk space
df -h
```

#### Database Connection Issues

```bash
# Test database connectivity
docker exec -it csic-reporting-db psql -U reporting_admin -d csic_reporting

# Check database health
docker exec csic-reporting-db pg_isready -U reporting_admin
```

#### Job Queue Backlog

```bash
# Check queue status
curl http://localhost:3001/api/v1/queue/stats

# Clear stuck jobs
curl -X POST http://localhost:3001/api/v1/queue/clean-stalled
```

#### Storage Issues

```bash
# Check MinIO health
docker exec csic-reporting-minio mc admin health local/

# Verify bucket
docker exec csic-reporting-minio mc ls local/csic-reports
```

### Log Analysis

```bash
# Application logs
docker-compose logs reporting-api | grep ERROR

# Search for specific patterns
docker-compose logs reporting-api | grep -i "report.*generation"

# Real-time log monitoring
docker-compose logs -f --tail=100 reporting-api
```

### Performance Tuning

#### Database Optimization

```sql
-- Add indexes for common queries
CREATE INDEX idx_reports_type_date ON reports(report_type, created_at);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_schedules_next_run ON schedules(next_run_at);
```

#### Redis Optimization

```bash
# Monitor memory usage
docker exec csic-reporting-redis redis-cli info memory

# Clear expired keys
docker exec csic-reporting-redis redis-cli FLUSHDB
```

#### Kafka Optimization

```bash
# Check topic lag
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group csic-reporting-group

# Increase partition count
kafka-topics.sh --bootstrap-server localhost:9092 --alter --topic reports --partitions 6
```

## Security Considerations

### Network Security

- Use TLS for all external connections
- Restrict database access to application network
- Implement firewall rules for production deployments

### Data Protection

- Enable WORM storage for compliance
- Implement encryption at rest
- Regular backup verification
- Access logging and auditing

### Secret Management

- Use secrets management (Vault, AWS Secrets Manager)
- Rotate credentials regularly
- Never commit secrets to version control

## Support

For issues and feature requests:

1. Check the troubleshooting section above
2. Review application logs
3. Contact the platform team

## License

This service is part of the CSIC Platform and is licensed under the same terms.
