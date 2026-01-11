# NEAM Platform Helm Library Infrastructure

## Overview

This document describes the complete Helm library chart infrastructure that has been implemented to standardize Kubernetes deployments across all NEAM Platform microservices. The infrastructure reduces configuration duplication, enforces security best practices, and provides consistent observability and operational capabilities.

## Architecture

### Library Chart Structure

The library chart is located at `/workspace/neam-platform/charts/library/` and provides reusable templates for all common Kubernetes resources.

```
library/
├── Chart.yaml              # Library chart definition
├── values.yaml             # Default configuration schema
├── README.md               # Comprehensive documentation
└── templates/
    ├── _helpers.tpl        # Helper functions for labels and naming
    ├── _serviceaccount.tpl # Service account configuration
    ├── _deployment.tpl     # Deployment with full container configuration
    ├── _service.tpl        # ClusterIP and headless services
    ├── _serviceheadless.tpl
    ├── _ingress.tpl        # Ingress with TLS and host routing
    ├── _configmap.tpl      # Standard ConfigMaps
    ├── _configmapfromfile.tpl # ConfigMaps with binary data
    ├── _secret.tpl         # Generic secrets
    ├── _tlssecret.tpl      # TLS certificate secrets
    ├── _pvc.tpl            # Persistent volume claims
    ├── _pdb.tpl            # Pod disruption budgets
    ├── _hpa.tpl            # Horizontal pod autoscalers
    └── NOTES.txt           # Installation notes
```

### Service Charts Migrated

All eight NEAM Platform microservices have been migrated to use the shared library chart:

1. **sensing** - Data ingestion and processing with full observability
2. **black-economy** - Black economy detection with ML workloads
3. **feature-store** - Feature engineering and storage
4. **macro** - Market trend analysis
5. **intelligence** - AI/ML model serving with GPU support
6. **policy** - Policy management and enforcement
7. **reporting** - Report generation and scheduling
8. **intervention** - Automated response actions

## Key Features

### Security Configuration

All service charts include comprehensive security contexts:

- Non-root container execution (UID 1000)
- Read-only root filesystem by default
- Dropped Linux capabilities
- Disabled privilege escalation
- Resource limits and requests

### Observability Configuration

Standardized observability across all services:

- Prometheus metrics endpoints
- Health probe endpoints (/health/live, /health/ready)
- Startup probes for slow-starting applications
- Structured logging with configurable log levels
- OpenTelemetry tracing configuration

### Resource Management

Consistent resource configuration:

- Configurable replica counts
- Horizontal pod autoscaling based on CPU/memory
- Pod disruption budgets for high availability
- Resource limits and requests with sensible defaults

### Advanced Scheduling

Node selection and scheduling controls:

- Node selectors for workload placement
- Pod anti-affinity for high availability
- Topology spread constraints for zone distribution
- Tolerations for special node pools (GPU, compute)

## Template Usage

### Basic Deployment

```yaml
# templates/deployment.yaml
{{- include "library.deployment" . -}}
```

### With Ingress and Autoscaling

```yaml
# templates/deployment.yaml
{{- include "library.deployment" . -}}

# templates/service.yaml
{{- include "library.service" . -}}
{{- include "library.serviceHeadless" . -}}

# templates/ingress.yaml
{{- include "library.ingress" . -}}

# templates/hpa.yaml
{{- include "library.horizontalPodAutoscaler" . -}}

# templates/pdb.yaml
{{- include "library.podDisruptionBudget" . -}}
```

### With Persistent Storage

```yaml
# templates/pvc.yaml
{{- include "library.persistentVolumeClaim" . -}}

# templates/deployment.yaml (include volume mounts)
{{- include "library.deployment" . -}}
```

## Validation Commands

To validate the Helm charts in your environment, ensure Helm is installed and run the following commands:

### Validate Library Chart

```bash
cd /workspace/neam-platform/charts/library
helm dependency update
helm lint .
```

### Validate Service Charts

```bash
cd /workspace/neam-platform/charts/sensing
helm dependency update
helm lint .
helm template sensing . --debug | head -200
```

### Render All Resources

```bash
helm template neam-platform /workspace/neam-platform/charts/sensing \
  --namespace neam \
  --values /workspace/neam-platform/charts/sensing/values.yaml \
  --debug
```

### Dry-Run Install

```bash
helm install neam-sensing /workspace/neam-platform/charts/sensing \
  --namespace neam \
  --dry-run \
  --debug
```

## Configuration Comparison

### Before (Individual Deployment Templates)

Each service had its own deployment template with duplicated configurations:

```yaml
# Old pattern - repeated in every service
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sensing
  labels:
    app: sensing
    version: v1
spec:
  replicas: 2
  selector:
    matchLabels:
      app: sensing
  template:
    metadata:
      labels:
        app: sensing
    spec:
      securityContext:
        runAsUser: 1000
      containers:
      - name: sensing
        image: sensing:1.0
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
```

### After (Library-Based Templates)

Services now use simple include statements:

```yaml
# New pattern - single line per resource
{{- include "library.deployment" . -}}
```

All configuration is centralized in values.yaml with consistent defaults.

## Service-Specific Configurations

### GPU Workloads (Intelligence)

```yaml
resources:
  limits:
    nvidia.com/gpu: 1
nodeSelector:
  node-type: gpu
tolerations:
  - key: "nvidia.com/gpu"
    operator: "Exists"
```

### ML Workloads (Black Economy)

```yaml
resources:
  limits:
    cpu: "4"
    memory: 4Gi
nodeSelector:
  node-type: ml-compute
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/name
              operator: In
              values:
                - black-economy
```

### Stateful Workloads (Feature Store, Reporting)

```yaml
persistentVolumeClaim:
  create: true
  size: 50Gi
  storageClassName: gp3
volumeMounts:
  - name: data
    mountPath: /data
```

## Migration Checklist

The following tasks have been completed for all eight service charts:

- [x] Created Chart.yaml with library dependency
- [x] Created comprehensive values.yaml with all library options
- [x] Created templates/ directory with library includes
- [x] Configured security contexts
- [x] Configured health probes
- [x] Configured resource limits and requests
- [x] Configured autoscaling where applicable
- [x] Configured pod disruption budgets
- [x] Configured service accounts with IAM annotations
- [x] Configured ingress where applicable
- [x] Configured persistent storage where applicable
- [x] Created README documentation

## Benefits

### Reduced Duplication

Before: ~150 lines of YAML per service template
After: ~1 line per resource type
Savings: ~90% reduction in template code

### Consistency

- Same labels across all resources
- Same security contexts
- Same probe configurations
- Same resource limits
- Same monitoring configuration

### Maintainability

- Single point of change for common patterns
- Centralized security updates
- Consistent upgrade path
- Clear documentation

### Operational Excellence

- Standardized health checks
- Consistent metrics endpoints
- Predictable scaling behavior
- Proper pod disruption budgets

## Next Steps

1. **Validate Charts**: Run Helm lint and template commands in your environment
2. **Test Deployments**: Perform dry-run installs to verify resource generation
3. **Integration Testing**: Deploy to staging environment with monitoring
4. **Documentation Updates**: Update team documentation with new patterns
5. **CI/CD Integration**: Add Helm chart validation to pipeline
