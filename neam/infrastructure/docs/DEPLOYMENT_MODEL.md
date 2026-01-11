# NEAM Platform - Deployment Model Documentation
## Government-Friendly Deployment Guide

This document provides comprehensive guidance for deploying the NEAM (National Economy Administration Platform) across various government infrastructure environments.

---

## 1. Supported Deployment Models

### 1.1 On-Premises Data Centers

**Description**: Full deployment within government-owned and operated data centers.

**Requirements**:
- **Hardware Specifications**:
  - Control Plane Nodes: 3x servers with 4x vCPU, 16GB RAM, 100GB SSD
  - Database Nodes: 6x servers with 8x vCPU, 32GB RAM, 500GB NVMe SSD
  - Analytics Nodes: 4x servers with 16x vCPU, 64GB RAM, 500GB NVMe SSD
  - Storage Nodes: 4x servers with 8x vCPU, 32GB RAM, 4TB HDD (RAID 6)

- **Network Requirements**:
  - Internal network: 10Gbps connectivity between nodes
  - Management network: Isolated management VLAN
  - Storage network: Dedicated high-bandwidth network

- **Storage**:
  - Local RAID 10 for databases
  - RAID 6 for bulk storage
  - Network Attached Storage (NAS) for backups

**Deployment Process**:
```bash
# Step 1: Prepare infrastructure
./scripts/prepare-infra.sh --on-prem --nodes=10

# Step 2: Install Kubernetes (RKE2 recommended)
./scripts/install-k8s.sh --distro=rke2 --high-availability

# Step 3: Deploy storage layer
./scripts/deploy-storage.sh --storage-class=local-path

# Step 4: Deploy databases
./scripts/deploy-databases.sh --postgres --timescaledb --clickhouse

# Step 5: Deploy application services
./scripts/deploy-services.sh --environment=production

# Step 6: Configure monitoring
./scripts/deploy-monitoring.sh --retention=30d

# Step 7: Verify deployment
./scripts/verify-deployment.sh --comprehensive
```

### 1.2 State Private Clouds

**Description**: Deployment within government-approved private cloud infrastructure (e.g., AWS GovCloud, Azure Government, OCI Government).

**Requirements**:
- **Cloud Account**: Dedicated government cloud account with proper IAM
- **VPC Configuration**: Private subnets with NAT gateway
- **Security Compliance**: FedRAMP Moderate or equivalent certification

**Key Configuration**:
```hcl
# Terraform configuration for private cloud
module "neam_private_cloud" {
  source = "./modules/cloud"

  deployment_model = "private-cloud"
  cloud_provider   = "aws"

  # AWS GovCloud configuration
  aws_config = {
    region              = "us-gov-west-1"
    vpc_cidr            = "10.0.0.0/16"
    private_subnets     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
    public_subnets      = ["10.0.101.0/24"]
    availability_zones  = ["us-gov-west-1a", "us-gov-west-1b", "us-gov-west-1c"]
  }

  # EKS configuration
  eks_config = {
    kubernetes_version = "1.28"
    instance_types     = ["m5.xlarge", "m5.2xlarge", "m5.4xlarge"]
    node_pools = {
      database = { min_size = 3, max_size = 9, instance_type = "m5.2xlarge" }
      general  = { min_size = 3, max_size = 15, instance_type = "m5.xlarge" }
      analytics = { min_size = 2, max_size = 6, instance_type = "m5.4xlarge" }
    }
  }
}
```

### 1.3 Air-Gapped Installations

**Description**: Deployment in completely isolated environments with no internet connectivity.

**Requirements**:
- **Bastion Server**: Air-gapped build server for preparing artifacts
- **Physical Media**: Secure transfer via encrypted USB or optical media
- **Local Registry**: Harbor or Docker Registry for container images
- **Offline Helm**: Local Helm repository for charts

**Air-Gap Preparation Process**:
```bash
# Step 1: Download all artifacts on connected machine
./scripts/download-artifacts.sh \
  --images=all \
  --charts=all \
  --output=/media/secure-usb/neam-artifacts

# Step 2: Create offline package
./scripts/create-offline-package.sh \
  --input=/media/secure-usb/neam-artifacts \
  --output=/media/secure-usb/neam-offline-$(date +%Y%m%d).tar.gz

# Step 3: Transfer to air-gapped environment
# Use secure physical transfer with proper chain of custody

# Step 4: Import artifacts in air-gapped environment
./scripts/import-artifacts.sh \
  --input=/media/secure-usb/neam-offline-*.tar.gz \
  --registry=harbor.neam.internal
```

**Air-Gap Manifest Configuration**:
```yaml
# Values override for air-gapped deployment
global:
  imageRegistry: "harbor.neam.internal"
  imagePullSecrets:
    - name: harbor-secret

  # Disable external service calls
  prometheus:
    externalLabels: "cluster=air-gapped"

  # Local Helm repository
  helm:
    repository: "http://harbor.neam.internal/chartrepo/neam"
```

### 1.4 Hybrid Central-State Deployments

**Description**: Distributed deployment across central government data centers and state/regional facilities.

**Architecture**:
```
Central Government Data Center
├── Primary Kubernetes Cluster
│   ├── Core Services (Authentication, Authorization)
│   ├── Primary Databases (PostgreSQL, TimescaleDB)
│   ├── Central Analytics (ClickHouse)
│   └── Compliance & Audit Storage
│
└── State/Regional Data Centers
    ├── State Cluster 1 (Cairo)
    │   ├── Regional Services
    │   ├── Local Cache (Redis)
    │   └── Local Search (OpenSearch)
    │
    ├── State Cluster 2 (Alexandria)
    │   └── ...
    │
    └── State Cluster N (Other Regions)
        └── ...
```

**Data Synchronization**:
```yaml
# Cross-cluster synchronization configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: sync-config
  namespace: neam
data:
  sync.yaml: |
    sync_mode: "distributed"
    clusters:
      - name: central
        endpoint: https://central.neam.gov.eg
        role: "primary"
        sync_interval: "1m"

      - name: cairo
        endpoint: https://cairo.neam.gov.eg
        role: "replica"
        sync_interval: "5m"

      - name: alexandria
        endpoint: https://alexandria.neam.gov.eg
        role: "replica"
        sync_interval: "5m"

    data_categories:
      transactional:
        replication: "sync"
        consistency: "strong"

      analytical:
        replication: "async"
        consistency: "eventual"

      archival:
        replication: "async"
        retention_central: "7y"
        retention_local: "1y"
```

---

## 2. Technology Stack

### 2.1 Container Runtime

**Docker** with hardened configuration:
```yaml
# Container security configuration
containerRuntime:
  type: "containerd"  # or CRI-O for FIPS compliance

  security:
    seccompProfile: "/etc/kubernetes/seccomp.json"
    apparmorProfile: "/etc/kubernetes/apparmor.json"
    rootless: false

  registry:
    mirror:
      - registry.neam.local:5000
```

### 2.2 Orchestration

**Kubernetes** (RKE2 for on-prem, managed service for cloud):
- **RKE2**: Kubernetes distribution with embedded etcd for high availability
- **K3s**: Lightweight option for edge/remote locations
- **AKS/EKS/GKE**: Managed Kubernetes for private cloud

**Configuration**:
```yaml
# Kubernetes configuration (config.yaml)
kubernetes:
  version: "1.28.0"
  network:
    plugin: "cilium"  # or "calico" for FIPS compliance
    cidr: "10.244.0.0/16"
  etcd:
    snapshot_interval: "6h"
    snapshot_retention: 5
  kubelet:
    eviction_hard: "memory.available<100Mi,nodefs.available<5%"
    feature_gates: "TTLAfterFinished=true,NodeLease=true"
```

### 2.3 Infrastructure as Code

**Terraform** for infrastructure provisioning:
- **State Management**: S3 backend with DynamoDB locking
- **Module Structure**: Modular design for flexibility
- **Secret Management**: Integration with HashiCorp Vault

### 2.4 CI/CD

**GitOps with ArgoCD**:
```yaml
# ArgoCD Application definition
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: neam-platform
  namespace: argocd
spec:
  project: neam-production
  source:
    repoURL: https://github.com/neam-platform/deployment
    targetRevision: main
    path: environments/production
  destination:
    server: https://kubernetes.default.svc
    namespace: neam
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

---

## 3. Security & Compliance

### 3.1 Security Controls

**Network Security**:
- Zero-trust network architecture
- mTLS for all service-to-service communication
- Network policies restricting traffic by namespace

**Data Security**:
- Encryption at rest (LUKS, TDE)
- Encryption in transit (TLS 1.3)
- HSM integration for key management

**Access Control**:
- RBAC with least privilege principle
- Service accounts with limited permissions
- Regular access reviews and certifications

### 3.2 Compliance Standards

**Supported Compliance Frameworks**:
- **Government Standards**: NIST 800-53, ISO 27001, SOC 2
- **Data Residency**: All data remains within national boundaries
- **Audit Requirements**: Complete audit trail for all operations

---

## 4. Disaster Recovery

### 4.1 Backup Strategy

**Database Backups**:
- PostgreSQL: pgBackRest with incremental backups
- TimescaleDB: Continuous backup with point-in-time recovery
- ClickHouse: Replicated tables with regular snapshots
- Redis: AOF persistence with scheduled snapshots

**Storage Backups**:
- MinIO: Object versioning with WORM compliance
- OpenSearch: Snapshot to MinIO with ISM policies

### 4.2 Recovery Procedures

**RTO (Recovery Time Objective)**: 4 hours
**RPO (Recovery Point Objective)**: 15 minutes

**Recovery Steps**:
```bash
# Restore PostgreSQL from backup
./scripts/restore-postgres.sh \
  --backup=s3://neam-backups/postgres/2024-01-15 \
  --target=neam-postgres-restore

# Restore ClickHouse
./scripts/restore-clickhouse.sh \
  --backup=minio://neam-backups/clickhouse/2024-01-15 \
  --target=neam-clickhouse-restore

# Verify data integrity
./scripts/verify-restore.sh --all
```

---

## 5. Monitoring & Maintenance

### 5.1 Monitoring Stack

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboards
- **OpenSearch**: Log aggregation and analysis
- **Jaeger**: Distributed tracing

### 5.2 Maintenance Procedures

**Certificate Rotation**:
```bash
# Rotate TLS certificates
./scripts/rotate-certificates.sh \
  --component=all \
  --validity=90d

# Verify certificate expiration
./scripts/check-certificates.sh --expiring=30d
```

**Database Maintenance**:
```bash
# PostgreSQL vacuum and reindex
./scripts/database-maintenance.sh \
  --postgres \
  --vacuum \
  --reindex \
  --analyze

# ClickHouse merge parts
./scripts/clickhouse-maintenance.sh \
  --optimize-tables \
  --cleanup-old-parts
```

---

## 6. Troubleshooting Guide

### 6.1 Common Issues

**Pod Not Starting**:
```bash
# Check pod events
kubectl describe pod <pod-name> -n <namespace>

# Check resource quotas
kubectl describe quota -n <namespace>

# Check node resources
kubectl describe nodes | grep -A5 "Allocated resources"
```

**Database Connection Issues**:
```bash
# Check PostgreSQL connection
kubectl exec -n neam-database neam-postgres-0 -- psql -c "\conninfo"

# Check connection pool
kubectl exec -n neam-database neam-postgres-pooler-0 -- psql -c "SELECT * FROM pg_stat_activity;"
```

### 6.2 Log Collection

**Gather Diagnostic Information**:
```bash
# Collect all logs
./scripts/gather-diagnostics.sh \
  --namespace=all \
  --since=24h \
  --output=diagnostics-$(date +%Y%m%d).tar.gz

# Export cluster state
kubectl cluster-info dump --all-namespaces > cluster-state.txt
```

---

This deployment model provides a comprehensive framework for deploying the NEAM platform across all supported government infrastructure environments while maintaining security, compliance, and operational excellence.
