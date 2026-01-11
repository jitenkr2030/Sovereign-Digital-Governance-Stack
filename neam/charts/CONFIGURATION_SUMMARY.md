# NEAM Platform Environment-Specific Configurations and Secret Management

This document summarizes the comprehensive environment-specific configurations and secret management integration implemented for the NEAM Platform.

## Overview

The NEAM Platform now includes fully configured environment-specific settings for development, staging, and production environments, with integrated HashiCorp Vault secret management using Kubernetes CSI driver for secure secret injection.

## Environment Configurations Summary

### Development Environment

The development environment is optimized for local development with minimal resource usage and full feature access.

| Configuration | Value | Purpose |
|---------------|-------|---------|
| **Replica Count** | 1 | Single instance for local development |
| **Resource Limits** | 1 CPU, 1Gi memory | Minimal resource usage |
| **Logging Level** | debug | Full debugging capabilities |
| **Autoscaling** | Disabled | No scaling needed for dev |
| **Ingress** | Disabled | Use port-forward for access |
| **Vault Integration** | Disabled | Use local secrets |
| **Feature Flags** | All enabled | Full feature testing |

**Key Database Configuration:**
- PostgreSQL: `localhost:5432` (no SSL)
- Redis: `localhost:6379` (no password)
- MongoDB: `localhost:27017` (no SSL)
- Kafka: `localhost:9092` (no auth)
- Elasticsearch: `localhost:9200` (no auth)

### Staging Environment

The staging environment mirrors production configuration with reduced resources for pre-production testing.

| Configuration | Value | Purpose |
|---------------|-------|---------|
| **Replica Count** | 2 | Medium availability |
| **Resource Limits** | 2 CPU, 2Gi memory | Moderate resource usage |
| **Logging Level** | info | Standard logging |
| **Autoscaling** | Enabled (2-5 replicas) | Load-based scaling |
| **Ingress** | Enabled (staging certs) | External access testing |
| **Vault Integration** | Enabled | Staging vault access |
| **Feature Flags** | Production-like | Stable features only |

**Key Database Configuration:**
- PostgreSQL: `neam-staging-postgresql.staging:5432` (SSL required)
- Redis: `neam-staging-redis.staging:6379` (password protected)
- MongoDB: `neam-staging-mongodb.staging:27017` (auth required)
- Kafka: 3-broker cluster (SASL authentication)
- Elasticsearch: `neam-staging-elasticsearch.staging:9200`

### Production Environment

The production environment is optimized for high availability, security, and performance.

| Configuration | Value | Purpose |
|---------------|-------|---------|
| **Replica Count** | 3-20 | High availability |
| **Resource Limits** | 4 CPU, 4Gi memory | High performance |
| **Logging Level** | info | Standard logging with monitoring |
| **Autoscaling** | Enabled (3-20 replicas) | Aggressive scaling |
| **Ingress** | Enabled (production certs) | External access |
| **Vault Integration** | Full TLS and rotation | Production vault |
| **Feature Flags** | Stable only | No beta features |

**Key Database Configuration:**
- PostgreSQL: HA cluster (3 nodes, synchronous commit, SSL)
- Redis: Cluster mode (6 nodes, TLS)
- MongoDB: Replica set (3 nodes, SSL, certificate auth)
- Kafka: 6-broker cluster (SASL/SSL authentication)
- Elasticsearch: Cluster with API key auth

## Feature Flags Configuration

### Development

```yaml
featureFlags:
  advancedAnalytics: true
  aiPrediction: true
  realTimeProcessing: true
  customRules: true
  auditLogging: true
  debugMode: true
  performanceMonitoring: true
  betaFeatures: true
```

### Staging

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

### Production

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

## Secret Management Integration

### Architecture

The NEAM Platform uses HashiCorp Vault with Kubernetes CSI driver for secret management:

```
┌─────────────────────────────────────────────────────────────────┐
│                        NEAM Platform                             │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │  Sensing │  │  Black   │  │ Feature  │  │  Macro   │        │
│  │          │  │ Economy  │  │  Store   │  │          │        │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘        │
│       │             │             │             │               │
│       └─────────────┴─────────────┴─────────────┘               │
│                            │                                     │
│                   CSI Secret Provider                            │
│                            │                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │            Kubernetes Secrets Store CSI Driver           │    │
│  └─────────────────────────────────────────────────────────┘    │
│                            │                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    HashiCorp Vault                       │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐  │    │
│  │  │  KV Store │  │   PKI    │  │  Database │  │  AWS   │  │    │
│  │  │  Engine   │  │  Engine  │  │  Engine  │  │ Engine │  │    │
│  │  └──────────┘  └──────────┘  └──────────┘  └────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Secret Structure

The secret hierarchy follows a clear organization pattern:

```
secret/
├── dev/
│   ├── database/
│   │   ├── postgresql/
│   │   ├── redis/
│   │   ├── mongodb/
│   │   └── elasticsearch/
│   ├── kafka/
│   ├── external/
│   └── application/
├── staging/
│   ├── database/
│   │   ├── postgresql/
│   │   ├── redis/
│   │   ├── mongodb/
│   │   └── elasticsearch/
│   ├── kafka/
│   ├── external/
│   └── application/
└── production/
    ├── database/
    │   ├── postgresql/
    │   ├── redis/
    │   ├── mongodb/
    │   └── elasticsearch/
    ├── kafka/
    ├── external/
    ├── application/
    └── certs/
```

### Secret Types

#### Static Secrets

| Secret Type | Description | Rotation Frequency |
|-------------|-------------|-------------------|
| Database Credentials | PostgreSQL, Redis, MongoDB usernames and passwords | 30 days |
| API Keys | External service API keys | 90 days |
| TLS Certificates | Server certificates | 30 days |
| Application Secrets | JWT, encryption, session secrets | 90 days |
| Kafka Credentials | SASL/SSL credentials | 90 days |

#### Dynamic Secrets

| Secret Type | Description | TTL |
|-------------|-------------|-----|
| Database Credentials | Temporary database users with limited permissions | 1 hour |
| AWS Credentials | Temporary AWS keys for S3, RDS access | 1 hour |

### CSI Driver Integration

#### SecretProviderClass Configuration

The SecretProviderClass is configured for each environment:

**Development:**
```yaml
csiSecretProvider:
  enabled: false
```

**Staging:**
```yaml
csiSecretProvider:
  enabled: true
  providerName: "vault"
  parameters:
    roleName: "neam-staging"
    vaultAddress: "https://vault-staging.neam.io"
    secrets: |
      - secretPath: "secret/staging/database/postgresql"
        data:
          - objectName: "username"
            key: "username"
          - objectName: "password"
            key: "password"
```

**Production:**
```yaml
csiSecretProvider:
  enabled: true
  providerName: "vault"
  parameters:
    roleName: "neam-production"
    vaultAddress: "https://vault.neam.io"
    vaultTLSClientCert: "/certs/vault/client.crt"
    vaultTLSClientKey: "/certs/vault/client.key"
    vaultCACert: "/certs/vault/ca.crt"
    secretRefreshRate: "120s"
    secretGracePeriod: "30s"
```

### Secret Rotation

#### Automatic Rotation

Vault automatically rotates dynamic secrets with configurable TTL:

```yaml
# Database role with automatic rotation
vault write database/roles/app-db-role \
  default_ttl="1h" \
  max_ttl="24h"
```

#### Scheduled Rotation

CronJobs handle static secret rotation:

| Secret Type | Schedule | Purpose |
|-------------|----------|---------|
| Database Credentials | Weekly (Sunday 2 AM) | Regular rotation |
| Application Secrets | Monthly (1st day 4 AM) | Monthly rotation |
| Kafka Credentials | Quarterly | Quarterly rotation |
| Emergency Rotation | On-demand | Security incidents |

### Vault Access Control

#### Kubernetes Authentication

```bash
# Enable Kubernetes authentication
vault auth enable kubernetes

# Configure Kubernetes auth
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc" \
  kubernetes_ca_cert="@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt" \
  token_reviewer_jwt="@/var/run/secrets/token" \
  issuer="kubernetes/serviceaccount"

# Create roles for each environment
vault write auth/kubernetes/role/neam-production \
  bound_service_account_names="neam-production-sa" \
  bound_service_account_namespaces="neam-production" \
  policies="neam-production-read" \
  ttl=1h
```

#### Vault Policies

**Production Policy:**
```hcl
path "secret/production/*" {
  capabilities = ["read"]
}

path "secret/production/database/*" {
  capabilities = ["read", "update"]
}

path "secret/production/kafka/*" {
  capabilities = ["read", "update"]
}
```

## Files Created

### Environment Configuration Files

| File | Purpose |
|------|---------|
| `environments/dev/values.yaml` | Development environment configuration |
| `environments/dev/secrets.yaml` | Development secrets template |
| `environments/dev/SETUP.md` | Development setup guide |
| `environments/staging/values.yaml` | Staging environment configuration |
| `environments/staging/SETUP.md` | Staging setup guide |
| `environments/prod/values.yaml` | Production environment configuration |
| `environments/prod/SETUP.md` | Production setup guide |

### Vault Integration Files

| File | Purpose |
|------|---------|
| `vault/secret-provider-class.yaml` | CSI SecretProviderClass configuration |
| `vault/vault-auth-config.yaml` | Vault authentication and policy configuration |
| `vault/secret-structure.yaml` | Secret hierarchy and structure documentation |
| `vault/secret-rotation.yaml` | Secret rotation policies and scripts |
| `vault/SECRET_MANAGEMENT.md` | Comprehensive secret management documentation |

## Deployment Instructions

### Development Deployment

```bash
# Deploy to development
./charts/deploy-dev.sh

# Access services via port-forward
kubectl port-forward -n neam-dev svc/neam-sensing 8080:8080
```

### Staging Deployment

```bash
# Deploy to staging
./charts/deploy-staging.sh

# Access services
# https://sensing.staging.neam.io
```

### Production Deployment

```bash
# Deploy to production (requires confirmation)
./charts/deploy-prod.sh

# Access services
# https://sensing.neam.io
```

### Manual Secret Configuration

```bash
# For development, create secrets manually
kubectl apply -f environments/dev/secrets.yaml

# For staging, retrieve from Vault
vault kv get secret/staging/database/postgresql

# For production, secrets are auto-injected via CSI
```

## Security Best Practices

### Development

- Use local secrets file for development
- Never commit secrets to version control
- Use different credentials than production
- Enable debug logging for troubleshooting

### Staging

- Use staging Vault instance
- Separate from production secrets
- Weekly secret rotation
- Staging certificates (not production)

### Production

- Use production Vault with TLS
- Strict access control policies
- Automatic secret rotation
- Production-grade certificates
- Regular security audits
- Comprehensive monitoring

## Monitoring and Alerting

### Secret Access Monitoring

- Vault audit logging enabled
- Failed authentication alerts
- Unusual access pattern detection
- Secret rotation notifications

### Environment Monitoring

| Environment | Critical Alert | Warning Alert |
|-------------|----------------|---------------|
| Development | N/A | Resource usage |
| Staging | Service down | High latency |
| Production | PagerDuty | Slack notification |

## Troubleshooting

### Common Issues

#### Secret Not Mounted

```bash
# Check SecretProviderClass status
kubectl get secretproviderclass -n neam-production

# Check pod events
kubectl describe pod <pod-name> -n neam-production

# Check CSI driver logs
kubectl logs -n kube-system -l app=csi-secrets-store-driver
```

#### Authentication Failures

```bash
# Verify Vault login
vault login -method=kubernetes role=neam-production

# Check Kubernetes auth configuration
vault read auth/kubernetes/config

# Verify service account
kubectl get serviceaccount neam-production-sa -n neam-production
```

#### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -n neam-production <pod-name> -- nc -zv neam-prod-postgresql 5432

# Check secrets
kubectl get secrets -n neam-production | grep postgresql

# Verify Vault secrets
vault kv get secret/production/database/postgresql
```

## Next Steps

1. **Complete Integration Testing**: Test all services with the new configuration
2. **Security Audit**: Review access policies and secret paths
3. **Performance Testing**: Validate production resource configuration
4. **Documentation Updates**: Update team documentation with new patterns
5. **CI/CD Integration**: Add validation to deployment pipeline
