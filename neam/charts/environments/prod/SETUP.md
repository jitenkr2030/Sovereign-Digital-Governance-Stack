# NEAM Platform Production Environment Setup Guide

This guide provides comprehensive instructions for setting up the NEAM Platform production environment.

## ⚠️ Important Warnings

- **Production deployment requires approval from security and operations teams**
- **All credentials must be obtained from the official secret management system**
- **Never use development or staging credentials in production**
- **All changes must go through the change management process**
- **Ensure you have read and understood the runbook before proceeding**

## Prerequisites

### Required Access

- [ ] Production Kubernetes cluster access (approved)
- [ ] Vault access with production secrets role
- [ ] Container registry access (production)
- [ ] Cloud provider console access
- [ ] Monitoring and alerting access
- [ ] On-call rotation access (if applicable)

### Required Tools

- [ ] kubectl configured with production context
- [ ] Helm 3.0+
- [ ] Vault CLI installed and configured
- [ ] Access to CI/CD pipeline
- [ ] PagerDuty access for incident response

### Pre-Deployment Checklist

- [ ] Security review completed
- [ ] Performance testing completed
- [ ] Integration testing completed
- [ ] User acceptance testing completed
- [ ] Backup and recovery tested
- [ ] Rollback procedure documented
- [ ] Monitoring and alerting configured
- [ ] Runbook created and reviewed
- [ ] Team notified of deployment

## Production Infrastructure

### Architecture Overview

```
                          ┌─────────────────────┐
                          │    CDN / WAF        │
                          └──────────┬──────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
          ┌─────────▼────────┐ ┌────▼────┐ ┌─────────▼────────┐
          │   Ingress NGINX   │ │  ALB   │ │   Ingress Istio   │
          └─────────┬────────┘ └───┬────┘ └─────────┬────────┘
                    │              │                │
                    └──────────────┼────────────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
    ┌─────────▼────────┐ ┌────────▼────────┐ ┌─────────▼────────┐
    │  Sensing Service │ │ Intelligence    │ │  Macro Service   │
    │  (3 replicas)    │ │ Service (3 repl)│ │  (3 replicas)    │
    └─────────┬────────┘ └────────┬────────┘ └─────────┬────────┘
              │                   │                    │
              └───────────────────┼────────────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
┌─────────▼─────────┐ ┌───────────▼───────────┐ ┌─────────▼─────────┐
│   PostgreSQL HA   │ │    Redis Cluster      │ │   Kafka Cluster   │
│   (Primary+2 rep) │ │     (6 nodes)         │ │   (6 brokers)     │
└───────────────────┘ └───────────────────────┘ └───────────────────┘
```

### Production Services

| Service | Replicas | CPU Limit | Memory Limit | Host |
|---------|----------|-----------|--------------|------|
| Sensing | 3-20 | 4 | 4Gi | sensing.neam.io |
| Black Economy | 3-10 | 4 | 4Gi | - |
| Feature Store | 3-6 | 2 | 4Gi | - |
| Macro | 3-5 | 2 | 2Gi | macro.neam.io |
| Intelligence | 3-6 | 8 | 16Gi | - |
| Policy | 2-5 | 1 | 1Gi | - |
| Reporting | 2-5 | 2 | 2Gi | reports.neam.io |
| Intervention | 1-3 | 1 | 1Gi | - |

### Database Services

| Service | Host | Port | SSL Mode | HA |
|---------|------|------|----------|-----|
| PostgreSQL | neam-prod-postgresql-primary.production | 5432 | verify-full | 3 nodes |
| Redis | neam-prod-redis-master.production | 6379 | TLS | 6 nodes |
| MongoDB | neam-prod-mongodb-0.production | 27017 | TLS | 3 nodes |
| Kafka | neam-prod-kafka-*.production | 9092 | SASL/SSL | 6 brokers |

## Deployment Steps

### 1. Pre-Deployment Verification

```bash
# Verify cluster access
kubectl config use-context production-neam
kubectl cluster-info
kubectl get nodes

# Verify all namespaces exist
kubectl get namespaces | grep neam

# Check current resource usage
kubectl top nodes
kubectl top pods -n neam-production

# Verify Vault access
vault login -method=kubernetes role=neam-production
vault status
```

### 2. Create Production Namespace

```bash
# Create namespace with resource quotas
kubectl apply -f namespaces/production.yaml

# Verify namespace
kubectl get namespace neam-production
kubectl describe namespace neam-production
```

### 3. Configure Secrets

```bash
# Get secrets from Vault
vault kv get -field=connection_string secret/production/database/postgresql

# Create registry secret
kubectl create secret docker-registry prod-registry-credentials \
  --docker-server=prod.neam.io \
  --docker-username=prod-ci \
  --docker-password=$(vault kv get -field=registry_password secret/production/registry) \
  --namespace=neam-production

# Apply SecretProviderClass
kubectl apply -f vault/secret-provider-class.yaml

# Verify secrets
kubectl get secrets -n neam-production | grep -E "(registry|vault)"
```

### 4. Install Prerequisites

```bash
# Install CSI Secret Provider
helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm upgrade --install -n kube-system csi-secrets-store \
  secrets-store-csi-driver/secrets-store-csi-driver \
  --set syncSecret.enabled=true \
  --set enableSecretRotation=true

# Install Vault Secrets Provider
helm repo add hashicorp https://helm.releases.hashicorp.com
helm upgrade --install -n kube-system vault-secrets-provider \
  hashicorp/vault-secrets-provider \
  --set secrets-store-csi-driver.enable=true
```

### 5. Deploy Services

```bash
# Deploy using the production deployment script
./charts/deploy-prod.sh

# Or deploy individual services with confirmation
helm upgrade --install sensing ./charts/sensing \
  --namespace neam-production \
  --values ./charts/sensing/values.yaml \
  --values ./charts/environments/prod/values.yaml \
  --wait \
  --timeout 10m \
  --atomic
```

### 6. Post-Deployment Verification

```bash
# Check all pods are running
kubectl get pods -n neam-production

# Check services
kubectl get svc -n neam-production

# Check ingress
kubectl get ingress -n neam-production

# Verify external access
curl -s https://sensing.neam.io/health | jq .
curl -s https://macro.neam.io/health | jq .

# Check logs for errors
kubectl logs -n neam-production -l app.kubernetes.io/part-of=neam-platform --tail=1000 | grep -i error | head -50
```

### 7. Run Smoke Tests

```bash
# Run smoke tests
./tests/smoke-test.sh --environment production

# Verify all endpoints
./tests/endpoint-test.sh
```

## Production Configuration

### Resource Limits

```yaml
resources:
  limits:
    cpu: "4"
    memory: 4Gi
  requests:
    cpu: "2"
    memory: 2Gi
```

### Autoscaling

```yaml
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

### Feature Flags

```yaml
featureFlags:
  advancedAnalytics: true
  aiPrediction: true
  realTimeProcessing: true
  customRules: true
  auditLogging: true
  debugMode: false
  performanceMonitoring: true
  betaFeatures: false
```

### Logging

```yaml
env:
  - name: LOG_LEVEL
    value: "info"
  - name: LOG_FORMAT
    value: "json"
```

## Monitoring and Alerting

### Key Metrics

| Metric | Warning | Critical | Action |
|--------|---------|----------|--------|
| CPU Usage | > 70% | > 90% | Scale up |
| Memory Usage | > 80% | > 95% | Scale up |
| Error Rate | > 1% | > 5% | Page on-call |
| Latency P99 | > 500ms | > 1s | Investigate |
| Pod Restarts | > 2/hour | > 10/hour | Investigate |

### Dashboards

- **Production Overview**: Grafana dashboard (ID: 10000)
- **Service Metrics**: Grafana dashboard (ID: 10001)
- **Infrastructure**: Grafana dashboard (ID: 10002)
- **Business Metrics**: Grafana dashboard (ID: 10003)

### Alert Channels

| Alert Type | Channel | Priority |
|------------|---------|----------|
| Critical | PagerDuty + Slack | P1 |
| Warning | Slack | P2 |
| Info | Slack | P3 |

## Backup and Recovery

### Backup Schedule

| Service | Frequency | Retention | Location |
|---------|-----------|-----------|----------|
| PostgreSQL | Continuous (WAL) | 30 days | S3 |
| PostgreSQL | Daily snapshot | 90 days | S3 |
| Redis | Every 15 min | 7 days | S3 |
| MongoDB | Every 6 hours | 30 days | S3 |
| Elasticsearch | Daily snapshot | 14 days | S3 |

### Recovery Procedures

```bash
# Restore PostgreSQL from backup
./scripts/restore-postgresql.sh --backup-id=<backup-id>

# Restore Redis from backup
./scripts/restore-redis.sh --backup-id=<backup-id>

# Restore MongoDB from backup
./scripts/restore-mongodb.sh --backup-id=<backup-id>
```

### Disaster Recovery

- **RTO (Recovery Time Objective)**: 4 hours
- **RPO (Recovery Point Objective)**: 15 minutes
- **DR Region**: us-west-2 (failover procedure in runbook)

## Rollback Procedure

### Automatic Rollback

The deployment script automatically rolls back on failure:

```bash
# If deployment fails, the --atomic flag will:
# 1. Rollback to previous release
# 2. Keep failed release for debugging
# 3. Exit with error code
```

### Manual Rollback

```bash
# List releases
helm list -n neam-production

# Rollback to previous version
helm rollback sensing 1 -n neam-production

# Verify rollback
kubectl get pods -n neam-production -l app.kubernetes.io/name=sensing
```

### Database Rollback

```bash
# Only if schema changes were deployed
# Restore from most recent backup
./scripts/restore-database.sh --type=postgresql --point-in-time
```

## Security Checklist

- [ ] All secrets from Vault
- [ ] TLS certificates valid
- [ ] Network policies applied
- [ ] RBAC configured
- [ ] Pod security policies enforced
- [ ] Audit logging enabled
- [ ] Vulnerability scan passed
- [ ] Penetration test passed

## On-Call Information

### Rotation

- Primary: Week 1
- Secondary: Week 2
- Escalation: Engineering Manager

### Escalation Path

1. On-call engineer
2. Engineering Manager
3. Director of Engineering
4. CTO

### Runbooks

- [ ] High CPU/Memory
- [ ] Pod crash loop
- [ ] Database connectivity issues
- [ ] Kafka lag
- [ ] SSL certificate expiry
- [ ] Security incident

## Post-Deployment Checklist

- [ ] Smoke tests passed
- [ ] Integration tests passed
- [ ] Performance baselines met
- [ ] Monitoring dashboards verified
- [ ] Alerting verified
- [ ] Team notified
- [ ] Documentation updated
- [ ] Runbook reviewed

## Support

### Incident Response

- **P1 (Critical)**: Immediate response, all hands on deck
- **P2 (High)**: Response within 30 minutes
- **P3 (Medium)**: Response within 4 hours
- **P4 (Low)**: Response within 24 hours

### Contact

- Slack: #neam-platform-oncall
- PagerDuty: https://neam.pagerduty.com
- Emergency: Call through PagerDuty
