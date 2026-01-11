# NEAM Platform Secret Management Documentation
# This document describes the secret management strategy using HashiCorp Vault

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Secret Types](#secret-types)
4. [CSI Driver Integration](#csi-driver-integration)
5. [Environment-Specific Configuration](#environment-specific-configuration)
6. [Secret Rotation](#secret-rotation)
7. [Access Control](#access-control)
8. [Security Best Practices](#security-best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

The NEAM Platform uses HashiCorp Vault for centralized secret management. Secrets are securely stored, accessed, and rotated using Vault, with Kubernetes CSI (Container Storage Interface) driver for seamless injection into pods.

### Key Features

- **Centralized Secret Storage**: All secrets stored in HashiCorp Vault
- **CSI Driver Integration**: Secrets mounted as files in pods
- **Dynamic Credentials**: Database credentials generated on-demand
- **Automatic Rotation**: Scheduled and manual secret rotation
- **Audit Logging**: All secret access logged in Vault
- **Encryption at Rest**: All secrets encrypted using KMS

## Architecture

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
│  │           (Mounts secrets as files in pods)              │    │
│  └─────────────────────────────────────────────────────────┘    │
│                            │                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Vault Agent Sidecar                   │    │
│  │           (Handles authentication and refresh)           │    │
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

## Secret Types

### Static Secrets

These secrets are stored in Vault and manually or automatically updated:

| Secret Type | Description | Rotation Frequency |
|-------------|-------------|-------------------|
| Database Credentials | PostgreSQL, Redis, MongoDB usernames and passwords | 30 days |
| API Keys | External service API keys | 90 days |
| TLS Certificates | Server certificates | 30 days |
| Application Secrets | JWT, encryption, session secrets | 90 days |
| Kafka Credentials | SASL/SSL credentials | 90 days |

### Dynamic Secrets

These secrets are generated on-demand by Vault:

| Secret Type | Description | TTL |
|-------------|-------------|-----|
| Database Credentials | Temporary database users with limited permissions | 1 hour |
| AWS Credentials | Temporary AWS keys for S3, RDS access | 1 hour |

### Certificate Secrets

| Secret Type | Description | Rotation Frequency |
|-------------|-------------|-------------------|
| TLS Certificates | Server certificates for HTTPS | 30 days |
| CA Certificates | Certificate authority certificates | 90 days |
| Client Certificates | Mutual TLS certificates | 30 days |

## CSI Driver Integration

### Prerequisites

1. Install Vault CSI driver:
```bash
helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
helm install -n kube-system csi-secrets-store \
  secrets-store-csi-driver/secrets-store-csi-driver \
  --set syncSecret.enabled=true
```

2. Install Vault secrets provider:
```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install -n kube-system vault-secrets-provider \
  hashicorp/vault-secrets-provider \
  --set secrets-store-csi-driver.enable=true
```

### SecretProviderClass Configuration

Create a SecretProviderClass for each environment:

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: neam-vault-secrets
  namespace: neam-platform
spec:
  provider: vault
  parameters:
    roleName: "neam-production"
    vaultAddress: "https://vault.neam.io:8200"
    objects: |
      - objectName: "postgresql-username"
        secretPath: "secret/production/database/postgresql"
        secretKey: "username"
```

### Pod Configuration

Reference the SecretProviderClass in pod specs:

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        volumeMounts:
        - name: vault-secrets
          mountPath: "/etc/secrets"
          readOnly: true
      volumes:
      - name: vault-secrets
        csi:
          driver: secrets-store.csi.k8s.io
          readOnly: true
          volumeAttributes:
            secretProviderClass: "neam-vault-secrets"
```

### Accessing Secrets in Application

Secrets are mounted as files:

```python
# Example: Reading PostgreSQL credentials
with open('/etc/secrets/postgresql-username', 'r') as f:
    username = f.read().strip()

with open('/etc/secrets/postgresql-password', 'r') as f:
    password = f.read().strip()

connection_string = f"postgresql://{username}:{password}@host:5432/db"
```

## Environment-Specific Configuration

### Development Environment

```yaml
vault:
  enabled: false  # Use local secrets for development
csiSecretProvider:
  enabled: false
```

### Staging Environment

```yaml
vault:
  enabled: true
  address: "https://vault-staging.neam.io"
  role: "neam-staging"
csiSecretProvider:
  enabled: true
  providerName: "vault"
  parameters:
    roleName: "neam-staging"
    vaultAddress: "https://vault-staging.neam.io"
```

### Production Environment

```yaml
vault:
  enabled: true
  address: "https://vault.neam.io"
  role: "neam-production"
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
```

## Secret Rotation

### Automatic Rotation

Vault automatically rotates dynamic secrets:

```yaml
# Database role with automatic rotation
vault write database/roles/app-db-role \
  default_ttl="1h" \
  max_ttl="24h"
```

### Scheduled Rotation

CronJobs handle static secret rotation:

```yaml
# Rotate database credentials weekly
apiVersion: batch/v1
kind: CronJob
metadata:
  name: neam-secret-rotation
spec:
  schedule: "0 2 * * 0"  # Sunday at 2 AM
```

### Manual Rotation

```bash
# Rotate PostgreSQL credentials
vault kv put secret/production/database/postgresql \
  username="neam_production" \
  password="$(openssl rand -base64 32)" \
  connection_string="postgresql://neam_production:NEW_PASSWORD@host:5432/db"

# Trigger application restart
kubectl rollout restart deployment/neam-sensing -n neam-production
```

### Emergency Rotation

```bash
# Emergency rotation script
./emergency_rotation.sh production
```

## Access Control

### Vault Policies

#### Production Policy

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

path "sys/leases/renew" {
  capabilities = ["update"]
}
```

#### Kubernetes Roles

```bash
# Create production role
vault write auth/kubernetes/role/neam-production \
  bound_service_account_names="neam-production-sa" \
  bound_service_account_namespaces="neam-production" \
  policies="neam-production-read" \
  ttl=1h
```

### RBAC in Kubernetes

```yaml
# Service account with secret access
apiVersion: v1
kind: ServiceAccount
metadata:
  name: neam-production-sa
  namespace: neam-production
```

## Security Best Practices

### 1. Secret Encryption

- Use KMS for encryption at rest
- Enable TLS for all Vault communication
- Use mTLS between Vault and Kubernetes

### 2. Access Limitation

- Use least privilege principle
- Separate secrets by environment
- Regular audit of secret access

### 3. Rotation Strategy

- Rotate credentials regularly
- Use dynamic credentials when possible
- Have rollback plan for failed rotations

### 4. Monitoring

- Enable Vault audit logging
- Monitor failed authentication attempts
- Alert on unusual secret access patterns

### 5. Backup and Recovery

- Enable Vault replication
- Regular backup of Vault data
- Test recovery procedures

## Troubleshooting

### Secret Not Mounted

```bash
# Check SecretProviderClass status
kubectl get secretproviderclass -n neam-platform

# Check pod events
kubectl describe pod <pod-name> -n neam-platform

# Check CSI driver logs
kubectl logs -n kube-system -l app=csi-secrets-store-driver
```

### Authentication Failures

```bash
# Verify Vault login
vault login -method=kubernetes role=neam-production

# Check Kubernetes auth configuration
vault read auth/kubernetes/config

# Verify service account
kubectl get serviceaccount neam-production-sa -n neam-production -o yaml
```

### Secret Rotation Issues

```bash
# Check Vault logs
vault audit list

# Verify lease status
vault list sys/leases/lookup/secret/production/database/postgresql

# Manually refresh secrets
kubectl delete pod <pod-name> -n neam-platform
```

### Performance Issues

```bash
# Check Vault performance
vault status

# Monitor CSI driver performance
kubectl top pods -n kube-system -l app=csi-secrets-store-driver

# Check network latency
kubectl exec -it <pod-name> -n neam-platform -- curl -s -o /dev/null -w "%{time_total}" https://vault.neam.io/v1/sys/health
```

## Additional Resources

- [Vault Documentation](https://www.vaultproject.io/docs)
- [CSI Driver Documentation](https://secrets-store-csi-driver.sigs.k8s.io)
- [Vault Kubernetes Auth](https://www.vaultproject.io/docs/auth/kubernetes)
- [Dynamic Secrets](https://www.vaultproject.io/docs/secrets/databases)
