# Air-Gapped Deployment Guide

This guide provides comprehensive instructions for deploying the NEAM platform in air-gapped (offline) environments where external network access is not available or restricted.

## Overview

Air-gapped deployments require pre-positioning all necessary artifacts within the secure environment. The NEAM platform air-gapped deployment support includes:

- **Offline Bundle Generator**: Creates self-contained deployment bundles containing container images, Helm charts, and configurations
- **Local Registry**: Deploys an internal OCI-compliant container registry for artifact distribution
- **Replication Policies**: Automates synchronization of required images from public registries
- **Comprehensive Monitoring**: Tracks registry health, storage, and synchronization status

## Prerequisites

### Environment Requirements

- Kubernetes cluster (v1.24 or later)
- kubectl configured with cluster access
- Helm 3.x installed
- At least 100GB of storage for registry and bundle artifacts
- Internal DNS configured for local registry access
- Images synchronized or bundled before deployment

### Required Tools

The following tools should be available in the connected environment for bundle generation:

- Docker or compatible container runtime
- skopeo (for container image operations)
- Helm 3.x
- kubectl
- oras (OCI Registry As Storage tool)
- Python 3.9+ with required packages

## Offline Bundle Generation

### Generating a Complete Deployment Bundle

The bundle generator creates a self-contained archive containing all required artifacts for air-gapped deployment:

```bash
# Navigate to scripts directory
cd /workspace/neam-platform/scripts/airgap

# Generate bundle with default settings
python3 bundle_generator.py generate \
    --output /path/to/output \
    --images images.json \
    --charts charts.json \
    --configs configs.json

# Specify platform version and environment
python3 bundle_generator.py generate \
    --output /path/to/output \
    --platform-version 1.0.0 \
    --environment production-airgapped \
    --images images.json \
    --charts charts.json
```

### Configuration Files

#### Container Images Configuration (images.json)

```json
[
  {
    "repository": "neam/core-api",
    "tag": "v1.0.0",
    "platform": "linux/amd64"
  },
  {
    "repository": "neam/auth-service",
    "tag": "v1.2.0",
    "platform": "linux/amd64"
  },
  {
    "repository": "gcr.io/spiffe-io/spire-server",
    "tag": "1.8.0",
    "platform": "linux/amd64"
  }
]
```

#### Helm Charts Configuration (charts.json)

```json
[
  {
    "name": "neam-core",
    "version": "1.0.0",
    "repository": "neam-charts",
    "chart_url": "https://charts.neam.internal/neam-core-1.0.0.tgz",
    "values_file": "values-production.yaml"
  },
  {
    "name": "local-registry",
    "version": "1.0.0",
    "repository": "neam-charts",
    "chart_url": "https://charts.neam.internal/local-registry-1.0.0.tgz"
  }
]
```

#### Configuration Bundles Configuration (configs.json)

```json
[
  {
    "name": "production-config",
    "source_path": "k8s/production",
    "config_type": "kubernetes",
    "transformers": ["namespace-isolation", "image-registry"]
  }
]
```

### Verifying Generated Bundles

Before transferring bundles to the air-gapped environment, verify their integrity:

```bash
# Verify bundle integrity
python3 bundle_generator.py verify --bundle /path/to/bundle.tar.gz

# List bundle contents
python3 bundle_generator.py list --bundle /path/to/bundle.tar.gz
```

Output example:
```
Bundle: complete
Version: 1.0.0
Created: 2024-01-15T10:30:00
Platform: 1.0.0
Environment: production-airgapped
Size: 5242880000 bytes
Checksum: abc123def456...

Container Images: 45
  - neam/core-api:v1.0.0
  - neam/auth-service:v1.2.0
  ...

Helm Charts: 12
  - neam-core-1.0.0.tgz
  - local-registry-1.0.0.tgz
  ...

Config Bundles: 3
  - production-config
  - staging-config
  - monitoring-config
```

## Local Registry Deployment

### Deploying the Local Registry

Deploy the local registry using Helm or direct YAML manifests:

```bash
# Using Helm
helm install local-registry ./charts/local-registry \
  --namespace registry \
  --create-namespace \
  --set persistence.size=100Gi

# Or using YAML manifests
kubectl apply -f charts/local-registry/templates/
```

### Configuring Registry Authentication

#### Basic Authentication Setup

```bash
# Generate htpasswd file
htpasswd -Bbn admin 'secure-password' | base64

# Create secret with credentials
kubectl create secret generic registry-htpasswd \
  --from-literal=htpasswd='<base64-encoded-content>' \
  --namespace registry

# Deploy with authentication
kubectl apply -f charts/local-registry/templates/auth.yaml
```

#### TLS Configuration

The registry automatically provisions TLS certificates using cert-manager:

```bash
# Install cert-manager (if not already installed)
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true

# Deploy registry with TLS
kubectl apply -f charts/local-registry/templates/deployment.yaml
```

### Registry Access Configuration

Configure Docker or container runtimes to trust the local registry:

```json
{
  "insecure-registries": ["local-registry.neam.internal:5000"],
  "registry-mirrors": ["https://local-registry.neam.internal:5000"]
}
```

For Kubernetes, configure image pull secrets:

```bash
# Create secret for registry access
kubectl create secret docker-registry registry-credentials \
  --docker-server=local-registry.neam.internal:5000 \
  --docker-username=admin \
  --docker-password='secure-password' \
  --namespace default

# Reference in pod spec
spec:
  imagePullSecrets:
    - name: registry-credentials
```

## Image Synchronization

### Automated Sync CronJobs

The registry deployment includes CronJobs for automated image synchronization:

```bash
# View sync schedule
kubectl get cronjob -n registry

# Manually trigger sync
kubectl create job --from=cronjob/registry-image-sync sync-manual -n registry
```

### Sync Manifest Configuration

Edit the sync manifest to customize which images are synchronized:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: registry-sync-manifest
  namespace: registry
data:
  sync-manifest.yaml: |
    sync:
      frequency: daily
      images:
        - source: "docker.io/library/nginx:1.25-alpine"
          target: "local-registry:5000/nginx:1.25-alpine"
          platform: "linux/amd64"
        
        # Add custom images here
        - source: "my-registry.example.com/custom-image:v1.0"
          target: "local-registry:5000/custom-image:v1.0"
```

### Manual Image Push

Push images directly to the local registry:

```bash
# Tag image for local registry
docker tag myimage:v1.0 local-registry:5000/myimage:v1.0

# Push to local registry
docker push local-registry:5000/myimage:v1.0

# Using ORAS
oras login local-registry:5000 -u admin -p password
oras push local-registry:5000/myimage:v1.0
```

## Bundle Deployment in Air-Gapped Environment

### Transferring Bundles

Transfer bundles to the air-gapped environment using approved media:

```bash
# Create compressed archive for transfer
tar -czvf neam-bundle-$(date +%Y%m%d).tar.gz -C /path/to/output .

# Calculate checksum for verification
sha256sum neam-bundle-*.tar.gz > checksums.txt

# Transfer via approved method (USB, encrypted media, etc.)
```

### Extracting and Loading Bundle

```bash
# Extract bundle
tar -xzvf neam-bundle-YYYYMMDD.tar.gz

# Load container images
cd neam-bundle/images
for img in *.tar; do
  docker load < "$img"
done

# Load Helm charts
helm repo add local file://$(pwd)/../charts

# Apply configurations
kubectl apply -f neam-bundle/configs/
```

### Deploying from Bundle

```bash
# Install Helm charts from bundled charts
helm install neam-core local/neam-core-*.tgz \
  --namespace neam \
  --create-namespace \
  -f neam-bundle/configs/production-config/values.yaml

# Deploy additional configurations
kubectl apply -f neam-bundle/configs/production-config/
```

## Monitoring and Operations

### Accessing Registry Metrics

Access registry metrics through Prometheus and Grafana:

```bash
# View registry metrics endpoint
curl http://local-registry:5000/metrics

# Access Grafana dashboard
# URL: http://grafana.neam.internal/d/local-registry-overview
```

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| up{job="local-registry"} | Registry availability | < 1 |
| http_request_duration_seconds | Request latency | p95 > 5s |
| http_requests_total{status=~"5.."} | Error rate | > 10% |
| pv_usage_bytes | Storage usage | > 80% |

### Troubleshooting Common Issues

#### Registry Not Accessible

```bash
# Check registry pods
kubectl get pods -n registry -l app=local-registry

# Check registry logs
kubectl logs -n registry -l app=local-registry --tail=100

# Verify service endpoint
kubectl endpoint registry -n registry
```

#### Image Push/Pull Failures

```bash
# Check authentication
curl -u admin:password http://local-registry:5000/v2/

# Verify TLS certificates
openssl s_client -connect local-registry:5000 -servername local-registry

# Check storage capacity
kubectl get pvc -n registry
```

#### Synchronization Failures

```bash
# Check sync job logs
kubectl logs job/registry-image-sync -n registry

# Verify network policies
kubectl get networkpolicy -n registry

# Check secret configuration
kubectl get secret registry-mirror-credentials -n registry
```

## Security Considerations

### Network Security

- Deploy registry in dedicated namespace with network policies
- Restrict egress to necessary destinations only
- Use TLS for all registry communications
- Implement rate limiting for registry access

### Access Control

- Use strong authentication (htpasswd with bcrypt)
- Implement role-based access for registry management
- Rotate credentials regularly
- Audit all registry access

### Data Protection

- Enable storage encryption at rest
- Implement regular backups of registry data
- Monitor for unusual blob access patterns
- Archive and purge unused images

## Maintenance Procedures

### Certificate Rotation

```bash
# Check certificate status
kubectl get certificate -n registry

# Rotate certificates
kubectl delete certificate registry-tls -n registry
kubectl create -f <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
...
EOF
```

### Storage Expansion

```bash
# Expand PVC size
kubectl patch pvc registry-pvc -n registry \
  -p '{"spec":{"resources":{"requests":{"storage":"150Gi"}}}}'
```

### Registry Backup

```bash
# Create backup of registry data
kubectl exec -n registry local-registry-xxx -- \
  tar -czf /tmp/registry-backup.tar.gz /var/lib/registry

# Copy backup locally
kubectl cp registry/local-registry-xxx:/tmp/registry-backup.tar.gz \
  ./registry-backup.tar.gz
```

## Checklist

Before deploying to air-gapped environment, verify:

- [ ] All required images bundled or synced
- [ ] Helm charts downloaded and validated
- [ ] Configuration files prepared and tested
- [ ] Registry deployed and accessible
- [ ] Authentication configured
- [ ] TLS certificates valid
- [ ] Storage capacity verified
- [ ] Monitoring configured
- [ ] Network policies applied
- [ ] Access credentials available
- [ ] Bundle checksum verified
- [ ] Deployment procedures documented
- [ ] Rollback plan prepared

## Support

For issues with air-gapped deployment:

1. Review logs: `kubectl logs -n registry -l app=local-registry`
2. Check events: `kubectl get events -n registry --sort-by='.lastTimestamp'`
3. Verify prerequisites: `python3 bundle_generator.py verify --bundle <bundle>`
4. Consult troubleshooting guide above
5. Contact platform operations team
