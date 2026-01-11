# Local Registry Operations Guide

This guide provides detailed procedures for operating and managing the local OCI container registry used in NEAM platform air-gapped deployments.

## Registry Overview

The local registry is a Docker Registry v2 compatible OCI registry that serves as the central artifact repository for air-gapped deployments. It provides:

- Container image storage and distribution
- Helm chart repository capabilities
- Configuration management
- Access control and authentication
- Monitoring and metrics

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Air-Gapped Environment                    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Local Registry (port 5000)              │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │   Images    │  │    Charts   │  │  Configs    │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │                                                     │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │           Authentication & TLS              │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                          │                                  │
│          ┌───────────────┼───────────────┐                 │
│          ▼               ▼               ▼                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   K8s Pods  │  │   CI/CD     │  │  Operators  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Initial Deployment

### Prerequisites

Before deploying the local registry, ensure:

- Kubernetes cluster with storage class available
- cert-manager for TLS certificate management
- Sufficient storage (minimum 50GB recommended)
- DNS resolution for `local-registry.neam.internal`

### Deployment Steps

#### 1. Create Registry Namespace

```bash
# Create the registry namespace
kubectl create namespace registry

# Apply label for network policies
kubectl label namespace registry \
  kubernetes.io/metadata.name=registry \
  app=local-registry tier=infrastructure
```

#### 2. Deploy Using Helm

```bash
# Add Helm repository (if charts are packaged)
helm repo add neam-registry https://charts.neam.internal

# Deploy with default configuration
helm install local-registry neam-registry/local-registry \
  --namespace registry \
  --create-namespace

# Deploy with custom values
helm install local-registry ./charts/local-registry \
  --namespace registry \
  --create-namespace \
  --values values-production.yaml
```

#### 3. Verify Deployment

```bash
# Check registry pod status
kubectl get pods -n registry -l app=local-registry

# Check registry service
kubectl get svc -n registry -l app=local-registry

# Test registry connectivity
curl -s http://local-registry:5000/v2/ | jq .

# Expected output: {}
```

### Configuration Options

#### Production Values File Example

```yaml
# values-production.yaml
replicaCount: 1

image:
  repository: docker.io/library/registry
  tag: "2.8"

persistence:
  enabled: true
  storageClass: "fast-regional"
  size: 100Gi

service:
  type: ClusterIP
  port: 5000
  annotations:
    prometheus.io/scrape: "true"

resources:
  requests:
    cpu: 200m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 2Gi

auth:
  enabled: true
  htpasswd: |
    admin:$2y$05$LvGpY8VF5Uq5dZ8v.8Hiv.OHKoIXdL6q0gHKp0d3E7jNKp0f8jNKp

networkPolicy:
  enabled: true
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: default
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 15s
```

## Registry Operations

### Managing Images

#### Pulling Images from External Registries

```bash
# Configure authentication for external registry (if needed)
docker login external-registry.example.com

# Pull image
docker pull external-registry.example.com/image:tag

# Tag for local registry
docker tag external-registry.example.com/image:tag \
  local-registry:5000/image:tag

# Push to local registry
docker push local-registry:5000/image:tag
```

#### Using Skopeo for Image Operations

```bash
# Copy image from external to local registry
skopeo copy \
  docker://external-registry.example.com/image:tag \
  docker://local-registry:5000/image:tag

# Copy with specific platform
skopeo copy \
  --dest-platform linux/amd64 \
  docker://external-registry.example.com/image:multiarch \
  docker://local-registry:5000/image:amd64

# Inspect remote image
skopeo inspect docker://local-registry:5000/image:tag
```

#### Using ORAS for Image Operations

```bash
# Login to registry
oras login local-registry:5000 -u admin -p password

# Push single image
oras push local-registry:5000/myapp:v1.0 \
  myapp:v1.0

# Push image with artifacts
oras push local-registry:5000/myapp:v1.0 \
  --artifact-type application/vnd.oci.image.manifest.v1+json \
  myapp:v1.0 sbom.tar.gz

# Pull image
oras pull local-registry:5000/myapp:v1.0

# List repository tags
oras repo tags local-registry:5000/myapp
```

### Managing Helm Charts

#### Adding Charts to Local Registry

```bash
# Download chart
helm pull https://charts.example.com/chart-1.0.0.tgz

# Add to local registry (using chartmuseum or similar)
helm plugin install https://github.com/chartmuseum/helm-push

# Push chart to local registry
cm-push chart-1.0.0.tgz local-registry:5000

# Alternative: Use ChartMuseum
CHARTMUSEUM_URL=http://local-registry:5000/chartrepo
helm push chart-1.0.0.tgz $CHARTMUSEUM_URL
```

#### Using Local Charts in Deployments

```bash
# Add local repository
helm repo add local-registry http://local-registry:5000/chartrepo

# Update repositories
helm repo update

# Install chart from local registry
helm install myapp local-registry/myapp \
  --version 1.0.0 \
  -f values.yaml
```

### Managing Access Control

#### htpasswd User Management

```bash
# Create new user
htpasswd -Bbn newuser 'secure-password' >> htpasswd

# Update existing user
htpasswd -Bbn admin 'new-password' htpasswd

# Remove user
htpasswd -D htpasswd username

# Create Kubernetes secret
kubectl create secret generic registry-htpasswd \
  --from-file=htpasswd=htpasswd \
  --namespace registry

# Restart registry to apply changes
kubectl rollout restart deployment local-registry -n registry
```

#### TLS Certificate Management

```bash
# Check certificate status
kubectl get certificate -n registry

# View certificate details
kubectl describe certificate registry-tls -n registry

# Check certificate expiration
kubectl get secret registry-tls-secret -n registry \
  -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -enddate

# Manual certificate renewal (if needed)
kubectl delete secret registry-tls-secret -n registry
kubectl delete certificate registry-tls -n registry
# Certificate will be recreated by cert-manager
```

## Replication and Synchronization

### Configuring Image Mirroring

The registry can be configured as a pull-through cache for external registries:

```yaml
# registry-config ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: registry-config
  namespace: registry
data:
  config.yml: |
    version: 0.1
    proxy:
      remoteurl: https://registry-1.docker.io
```

### Automated Image Sync CronJobs

#### Image Sync Job Configuration

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: registry-image-sync
  namespace: registry
spec:
  schedule: "0 */4 * * *"  # Every 4 hours
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: local-registry
          containers:
          - name: sync
            image: oras:1.1
            command:
              - /bin/sh
              - -c
              - |
                # Sync critical images
                IMAGES=(
                  "docker.io/library/nginx:1.25-alpine"
                  "docker.io/library/redis:7-alpine"
                )
                for IMG in "${IMAGES[@]}"; do
                  oras pull "docker://${IMG}"
                  oras push "docker://local-registry:5000/${IMG}"
                done
```

#### Manual Sync Execution

```bash
# Run sync job immediately
kubectl create job --from=cronjob/registry-image-sync sync-now -n registry

# View sync logs
kubectl logs job/sync-now -n registry

# Check sync status
kubectl describe job sync-now -n registry
```

### Managing Sync Manifest

The sync manifest defines which images should be synchronized:

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
          schedule: "0 2 * * *"
        
        - source: "gcr.io/spiffe-io/spire-server:1.8.0"
          target: "local-registry:5000/spiffe-io/spire-server:1.8.0"
          platform: "linux/amd64"
          schedule: "0 3 * * *"
      
      policies:
        keep_last_versions: 3
        max_image_age_days: 30
        cleanup_unused: true
```

## Monitoring and Observability

### Accessing Metrics

The registry exposes metrics on port 5001:

```bash
# Direct metrics access
curl http://local-registry:5001/metrics

# Key metrics include:
# - http_requests_total - Total HTTP requests
# - http_request_duration_seconds - Request latency
# - http_response_size_bytes - Response sizes
# - storage_bytes - Storage used
# - repository_count - Number of repositories
# - manifest_count - Number of manifests
# - blob_count - Number of blobs
```

### Grafana Dashboard

Access the registry dashboard in Grafana:

1. Navigate to Grafana: `http://grafana.neam.internal`
2. Search for "Local Registry Overview" dashboard
3. View metrics for:
   - Registry availability
   - Request rate and latency
   - Storage usage
   - Error rates
   - Image counts

### Prometheus Alerts

Key alerts configured for registry monitoring:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: local-registry-alerts
  namespace: monitoring
spec:
  groups:
  - name: local-registry
    rules:
    - alert: RegistryDown
      expr: up{job="local-registry"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Registry is down"
    
    - alert: RegistryHighErrorRate
      expr: |
        rate(http_requests_total{job="local-registry",status=~"5.."}[5m]) >
        rate(http_requests_total{job="local-registry"}[5m]) * 0.1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High error rate on registry"
    
    - alert: RegistryStorageFull
      expr: |
        (pv_capacity_bytes - pv_usage_bytes) / pv_capacity_bytes < 0.1
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Registry storage is low"
```

## Maintenance Procedures

### Regular Maintenance Tasks

#### Daily Tasks

```bash
# Check registry health
kubectl exec -n registry deploy/local-registry -- wget -q -O - http://localhost:5000/v2/

# Review recent logs
kubectl logs -n registry deploy/local-registry --tail=100 --since=24h

# Check storage usage
kubectl exec -n registry deploy/local-registry -- df -h /var/lib/registry
```

#### Weekly Tasks

```bash
# Review sync job status
kubectl get cronjob -n registry

# Check for unused images
oras repo ls local-registry:5000

# Review error rates
curl -s http://local-registry:5001/metrics | grep http_requests_total | grep 5
```

#### Monthly Tasks

```bash
# Review and rotate credentials
htpasswd -c /tmp/htpasswd admin
# Update htpasswd secret

# Certificate review
kubectl get certificates -n registry

# Storage analysis
kubectl exec -n registry deploy/local-registry -- \
  du -sh /var/lib/registry/* | sort -h

# Performance review
# Check Grafana dashboard for latency trends
```

### Storage Management

#### Monitoring Storage Usage

```bash
# Check PVC usage
kubectl get pvc -n registry

# View detailed storage metrics
kubectl exec -n registry deploy/local-registry -- \
  cat /proc/mounts | grep registry

# Check blob statistics
curl -s http://local-registry:5001/metrics | grep storage
```

#### Expanding Storage

```bash
# Patch PVC to increase size
kubectl patch pvc registry-pvc -n registry \
  -p '{"spec":{"resources":{"requests":{"storage":"150Gi"}}}}'

# Verify expansion
kubectl get pvc registry-pvc -n registry
```

#### Garbage Collection

```bash
# Trigger garbage collection
curl -X POST http://local-registry:5000/v2/_oci/artifacts/delete

# Schedule automatic GC
# See CronJob: registry-gc in replication.yaml
```

### Backup and Recovery

#### Creating Backups

```bash
# Create backup script
cat > /tmp/backup-registry.sh << 'EOF'
#!/bin/bash
BACKUP_DIR=/backups/registry-$(date +%Y%m%d)
mkdir -p $BACKUP_DIR

# Backup registry data
kubectl exec -n registry deploy/local-registry -- \
  tar -czf /tmp/registry-data.tar.gz /var/lib/registry

# Copy backup locally
kubectl cp registry/deploy-local-registry:/tmp/registry-data.tar.gz \
  $BACKUP_DIR/

# Backup configuration
kubectl get configmap -n registry -o yaml > $BACKUP_DIR/configmaps.yaml

# Backup secrets (without sensitive data)
kubectl get secret -n registry -o yaml > $BACKUP_DIR/secrets.yaml

echo "Backup complete: $BACKUP_DIR"
EOF

chmod +x /tmp/backup-registry.sh
./backup-registry.sh
```

#### Restoring from Backup

```bash
# Restore registry data
kubectl exec -n registry deploy/local-registry -- \
  rm -rf /var/lib/registry/*

kubectl cp $BACKUP_DIR/registry-data.tar.gz \
  registry/deploy-local-registry:/tmp/

kubectl exec -n registry deploy/local-registry -- \
  tar -xzf /tmp/registry-data.tar.gz -C /

# Restart registry
kubectl rollout restart deployment local-registry -n registry
```

### Upgrade Procedures

#### Upgrading Registry Version

```bash
# Check current version
kubectl get deployment local-registry -n registry \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# Update image version
kubectl set image deployment local-registry \
  registry=docker.io/library/registry:2.9 \
  -n registry

# Monitor rollout
kubectl rollout status deployment local-registry -n registry

# Verify deployment
kubectl exec -n registry deploy/local-registry -- \
  wget -q -O - http://localhost:5000/v2/
```

#### Helm Upgrade

```bash
# Update Helm repository
helm repo update

# Upgrade release
helm upgrade local-registry neam-registry/local-registry \
  --namespace registry \
  --values values-production.yaml

# Verify upgrade
helm status local-registry -n registry
```

## Troubleshooting

### Common Issues and Resolutions

#### Registry Returns 401 Unauthorized

```bash
# Check authentication configuration
kubectl get configmap registry-auth-config -n registry -o yaml

# Verify htpasswd secret
kubectl get secret registry-htpasswd -n registry

# Test authentication
curl -u admin:password http://local-registry:5000/v2/

# Restart registry if configuration changed
kubectl rollout restart deployment local-registry -n registry
```

#### TLS Certificate Errors

```bash
# Check certificate status
kubectl get certificate registry-tls -n registry

# View certificate events
kubectl describe certificate registry-tls -n registry

# Verify secret
kubectl get secret registry-tls-secret -n registry -o yaml

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager
```

#### Storage Issues

```bash
# Check PVC status
kubectl get pvc registry-pvc -n registry

# Check storage class
kubectl get storageclass fast-regional

# View storage events
kubectl get events -n registry --field-selector involvedObject.name=registry-pvc

# Check node storage capacity
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}: {.status.allocatable.storage}'{"\n"}{end}'
```

#### Image Push/Pull Failures

```bash
# Check network policies
kubectl get networkpolicy -n registry

# Test network connectivity
kubectl exec -n registry deploy/local-registry -- \
  nc -zv kubernetes.default.svc 443

# Check DNS resolution
kubectl exec -n registry deploy/local-registry -- \
  nslookup local-registry

# Verify service endpoint
kubectl endpoint local-registry -n registry
```

### Diagnostic Commands

```bash
# Registry health check
kubectl exec -n registry deploy/local-registry -- \
  wget -q -O - http://localhost:5000/v2/

# Registry configuration
kubectl exec -n registry deploy/local-registry -- \
  cat /etc/docker/registry/config.yml

# Storage statistics
kubectl exec -n registry deploy/local-registry -- \
  du -sh /var/lib/registry/*

# Active connections
kubectl exec -n registry deploy/local-registry -- \
  netstat -ant | grep 5000

# Recent errors
kubectl logs -n registry deploy/local-registry --since=1h | grep -i error
```

## Security Best Practices

### Access Control

1. **Use Strong Authentication**
   - Implement htpasswd with bcrypt
   - Rotate passwords quarterly
   - Use service accounts for automated operations

2. **Implement Network Policies**
   - Restrict ingress to authorized namespaces
   - Limit egress to necessary destinations
   - Deny all by default

3. **Enable TLS**
   - Use valid TLS certificates
   - Disable insecure operations
   - Configure strong cipher suites

### Data Protection

1. **Storage Security**
   - Enable encryption at rest
   - Use encrypted storage classes
   - Implement regular backups

2. **Image Signing**
   - Use notation or sigstore for image signing
   - Verify signatures before deployment
   - Maintain trust records

3. **Vulnerability Scanning**
   - Scan images before pushing to registry
   - Block vulnerable images
   - Regular vulnerability assessments

### Audit and Compliance

1. **Enable Logging**
   - Enable registry access logging
   - Export logs to central system
   - Retain logs per compliance requirements

2. **Audit Trail**
   - Log all image push/pull operations
   - Track user access
   - Document configuration changes

3. **Compliance Reporting**
   - Generate compliance reports
   - Track security exceptions
   - Maintain policy documentation

## API Reference

### Registry REST API

#### Authentication

```bash
# Token-based authentication
curl -H "Authorization: Bearer <token>" http://local-registry:5000/v2/

# Basic authentication
curl -u username:password http://local-registry:5000/v2/
```

#### Catalog Operations

```bash
# List repositories
curl -u admin:password http://local-registry:5000/v2/_catalog

# List tags for repository
curl -u admin:password http://local-registry:5000/v2/<repo>/tags/list

# Get manifest
curl -u admin:password \
  http://local-registry:5000/v2/<repo>/manifests/<tag>

# Get blob
curl -u admin:password \
  http://local-registry:5000/v2/<repo>/blobs/<digest>
```

#### Health Check

```bash
# Basic health check
curl http://local-registry:5000/v2/

# Detailed health (if enabled)
curl http://localhost:5000/debug/health
```

## Support Resources

### Useful Links

- Docker Registry v2 Documentation: https://docs.docker.com/registry/
- ORAS Project: https://oras.land/
- Skopeo Documentation: https://github.com/containers/skopeo
- Helm Chart Management: https://helm.sh/docs/

### Getting Help

For issues with local registry operations:

1. Review this documentation
2. Check kubectl logs and events
3. Consult Grafana dashboards
4. Contact platform operations team
5. Submit issue to project repository
