# NEAM Platform Staging Environment Setup Guide

This guide provides instructions for setting up the NEAM Platform staging environment.

## Overview

The staging environment mirrors production configuration with reduced resources. It is used for:
- Pre-production testing
- Integration testing
- Performance testing
- User acceptance testing

## Prerequisites

- Access to staging Kubernetes cluster
- kubectl configured with staging context
- Access to staging container registry
- Access to staging Vault instance
- Helm 3.0+

## Quick Start

### 1. Configure kubectl

```bash
# Set context to staging cluster
kubectl config use-context staging-neam

# Verify cluster access
kubectl cluster-info
kubectl get nodes
```

### 2. Create Namespace

```bash
# Create staging namespace
kubectl create namespace neam-staging

# Verify namespace
kubectl get namespaces | grep neam-staging
```

### 3. Install Prerequisites

```bash
# Install CSI Secret Provider (if not already installed)
helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm install -n kube-system csi-secrets-store \
  secrets-store-csi-driver/secrets-store-csi-driver

# Install Vault Secrets Provider
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install -n kube-system vault-secrets-provider \
  hashicorp/vault-secrets-provider
```

### 4. Configure Secrets

```bash
# Create Kubernetes secrets for registry access
kubectl create secret docker-registry staging-registry-credentials \
  --docker-server=staging.neam.io \
  --docker-username=staging-ci \
  --docker-password=${STAGING_REGISTRY_PASSWORD} \
  --namespace=neam-staging

# Create namespace-specific secrets
kubectl apply -f vault/secret-provider-class.yaml
```

### 5. Deploy Services

```bash
# Deploy all services to staging
./charts/deploy-staging.sh

# Or deploy individual services
helm upgrade --install sensing ./charts/sensing \
  --namespace neam-staging \
  --values ./charts/sensing/values.yaml \
  --values ./charts/environments/staging/values.yaml
```

### 6. Verify Deployment

```bash
# Check pod status
kubectl get pods -n neam-staging

# Check services
kubectl get svc -n neam-staging

# Check ingress
kubectl get ingress -n neam-staging

# View logs
kubectl logs -n neam-staging -l app.kubernetes.io/part-of=neam-platform --tail=100
```

## Staging Infrastructure

### Database Services

| Service | Host | Port | SSL |
|---------|------|------|-----|
| PostgreSQL | neam-staging-postgresql.staging | 5432 | Required |
| Redis | neam-staging-redis.staging | 6379 | TLS |
| MongoDB | neam-staging-mongodb.staging | 27017 | TLS |
| Kafka | neam-staging-kafka-0.staging:9092 | 9092 | SASL |

### External Services

| Service | URL | Purpose |
|---------|-----|---------|
| Schema Registry | http://neam-staging-schema-registry.staging:8081 | Schema validation |
| Data Quality | http://neam-staging-data-quality.staging:8082 | Data quality checks |
| Notification | http://neam-staging-notification.staging:8083 | Alert delivery |
| Vault | https://vault-staging.neam.io | Secret management |

### Domain Configuration

| Service | External URL | Internal Endpoint |
|---------|--------------|-------------------|
| Sensing | https://sensing.staging.neam.io | neam-sensing.neam-staging:8080 |
| Macro | https://macro.staging.neam.io | neam-macro.neam-staging:8080 |
| Reporting | https://reports.staging.neam.io | neam-reporting.neam-staging:8080 |

## Configuration

### Resource Limits

```yaml
resources:
  limits:
    cpu: "2"
    memory: 2Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

### Replica Configuration

```yaml
replicaCount: 2

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
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

### Logging Level

```yaml
env:
  - name: LOG_LEVEL
    value: "info"
  - name: LOG_FORMAT
    value: "json"
```

## Vault Configuration

### Enable Vault Integration

```bash
# Configure Kubernetes authentication
vault auth enable kubernetes

# Create staging role
vault write auth/kubernetes/role/neam-staging \
  bound_service_account_names="neam-staging-sa" \
  bound_service_account_namespaces="neam-staging" \
  policies="neam-staging-readonly" \
  ttl=1h

# Create staging secrets
vault kv put secret/staging/database/postgresql \
  username="neam_staging" \
  password="staging_password" \
  connection_string="postgresql://neam_staging:staging_password@neam-staging-postgresql.staging:5432/neam_staging?sslmode=require"
```

### Secret Rotation

```bash
# Rotate database credentials weekly
kubectl apply -f vault/secret-rotation.yaml
```

## Monitoring

### Prometheus

```bash
# Access Prometheus
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# View NEAM metrics
# Navigate to: http://localhost:9090/graph
# Query: neam_*
```

### Grafana

```bash
# Access Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Default credentials
# Username: admin
# Password: admin
```

### Alerts

- Warning threshold: CPU > 70% for 5 minutes
- Critical threshold: CPU > 90% for 3 minutes
- Warning threshold: Memory > 80%
- Critical threshold: Memory > 95%

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -n neam-staging

# Describe pod for events
kubectl describe pod <pod-name> -n neam-staging

# Check logs
kubectl logs <pod-name> -n neam-staging

# Check resource quotas
kubectl describe resourcequota -n neam-staging
```

### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -n neam-staging <pod-name> -- nc -zv neam-staging-postgresql.staging 5432

# Check secrets
kubectl get secrets -n neam-staging | grep postgresql

# Verify Vault secrets
vault kv get secret/staging/database/postgresql
```

### Vault Secret Issues

```bash
# Check SecretProviderClass
kubectl get secretproviderclass -n neam-staging

# Check CSI driver
kubectl logs -n kube-system -l app=csi-secrets-store-driver

# Test secret retrieval
kubectl exec -n neam-staging <pod-name> -- cat /etc/secrets/postgresql-username
```

### Performance Issues

```bash
# Check resource usage
kubectl top pods -n neam-staging

# Check node resources
kubectl describe nodes | grep -A 5 "Allocated resources"

# Check for throttling
kubectl logs -n neam-staging <pod-name> | grep throttl
```

## Rollback Procedure

### Rollback Deployment

```bash
# List previous revisions
helm history sensing -n neam-staging

# Rollback to previous version
helm rollback sensing 1 -n neam-staging

# Verify rollback
kubectl get pods -n neam-staging -l app.kubernetes.io/name=sensing
```

### Rollback Database

```bash
# Restore from backup
# (Use your database backup and restore procedure)
```

## Data Management

### Backup Schedule

- PostgreSQL: Daily at 2 AM UTC
- Redis: Every 15 minutes
- MongoDB: Every 6 hours

### Data Reset

```bash
# Reset staging database (WARNING: deletes all data)
kubectl exec -n neam-staging neam-staging-postgresql-0 -- psql -U neam_staging -c "DROP DATABASE neam_staging; CREATE DATABASE neam_staging;"
```

## Security

### Access Control

- Production access: Engineering team leads
- Staging access: All developers
- Read-only access: QA team

### Secrets

- All secrets managed through Vault
- No hardcoded secrets in values.yaml
- Secrets rotated every 30 days

### Network Policy

```yaml
# Default deny all ingress
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
  namespace: neam-staging
spec:
  podSelector: {}
  policyTypes:
  - Ingress
```

## Next Steps

1. Complete [Integration Testing](../tests/integration/README.md)
2. Review [Performance Guidelines](../docs/performance.md)
3. Prepare [Release Checklist](../RELEASE_CHECKLIST.md)
