# NEAM Platform Helm Library Chart

A shared Helm library chart providing reusable Kubernetes resource templates for all NEAM Platform microservices.

## Overview

This library chart defines common patterns for Deployments, Services, Ingresses, ConfigMaps, Secrets, and other Kubernetes resources used across all NEAM Platform services. By using this library, all services maintain consistent configurations for labels, annotations, security contexts, probes, and resource limits.

## Usage

To use this library chart in your service chart, add the following dependency to your `Chart.yaml`:

```yaml
dependencies:
  - name: library
    version: "1.0.0"
    repository: "file://../library"
```

## Templates

The library provides the following templates that can be included in your service's templates:

### Core Templates

| Template | Description |
|----------|-------------|
| `library.deployment` | Main deployment resource with all container configuration |
| `library.service` | ClusterIP service for internal communication |
| `library.serviceHeadless` | Headless service for stateful applications |
| `library.ingress` | Ingress resource with TLS and host configuration |
| `library.serviceAccount` | Service account with optional annotations |

### Data Management Templates

| Template | Description |
|----------|-------------|
| `library.configmap` | Standard ConfigMap for configuration files |
| `library.configmapFromFile` | ConfigMap with binary data support |
| `library.secret` | Generic secret for sensitive data |
| `library.tlsSecret` | TLS certificate secret |
| `library.persistentVolumeClaim` | PVC for persistent storage |

### Operational Templates

| Template | Description |
|----------|-------------|
| `library.podDisruptionBudget` | PDB for controlled disruptions |
| `library.horizontalPodAutoscaler` | HPA for automatic scaling |
| `library.helpers` | Helper functions for labels and naming |

## Helper Functions

The library provides the following helper functions:

| Function | Description |
|----------|-------------|
| `library.name` | Chart name (respects nameOverride) |
| `library.fullname` | Full qualified name (release-chart) |
| `library.chart` | Chart name and version |
| `library.labels` | Common labels for all resources |
| `library.selectorLabels` | Labels for resource matching |
| `library.serviceAccountName` | Service account name |
| `library.podMetadata` | Pod metadata with annotations |
| `library.imageTag` | Image tag (respects overrides) |
| `library.imagePullSecrets` | Image pull secret configuration |

## Values Schema

### Global Values

```yaml
global:
  imagePullSecrets: []    # Global image pull secrets
  labels: {}              # Labels applied to all resources
```

### Core Configuration

```yaml
# Override chart name
nameOverride: ""
fullnameOverride: ""

# Replica configuration
replicaCount: 1

# Image configuration
image:
  repository: neam-platform/service
  pullPolicy: IfNotPresent
  tag: ""

# Service account
serviceAccount:
  create: true
  name: ""
  annotations: {}

# Deployment configuration
deployment:
  annotations: {}
  strategy: {}

# Pod configuration
podAnnotations: {}
podSecurityContext:
  enabled: false
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000

# Container configuration
containerPorts: {}
containerPortProtocol: TCP
env: []
envFrom: []
command: []
args: []
volumeMounts: []
volumes: []

# Security context
containerSecurityContext:
  enabled: false
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false

# Health probes
livenessProbe:
  enabled: true
  path: /health
  port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
  successThreshold: 1

readinessProbe:
  enabled: true
  path: /ready
  port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
  successThreshold: 1

startupProbe:
  enabled: false
  path: /health
  port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 30

# Resources
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# Networking
service:
  enabled: true
  type: ClusterIP
  port: 8080
  annotations: {}
  portMappings: {}

serviceHeadless:
  enabled: false
  annotations: {}

ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts: []
  tls: []

# Configuration
configmap:
  create: false
  data: {}

configmapFromFile:
  create: false
  data: {}
  binaryData: {}

secret:
  create: false
  type: Opaque
  data: {}
  stringData: {}

tlsSecret:
  create: false
  crt: ""
  key: ""

# Storage
persistentVolumeClaim:
  create: false
  size: 1Gi
  accessModes:
    - ReadWriteOnce
  storageClassName: ""
  annotations: {}

# Advanced configuration
topologySpreadConstraints: []
nodeSelector: {}
affinity: {}
tolerations: []
terminationGracePeriodSeconds: 30
priorityClassName: ""
schedulerName: ""
dnsPolicy: ""
dnsConfig: {}
hostAliases: []
initContainers: []
lifecycle: {}

# Autoscaling
autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 10
  metrics: []
  behavior: {}
  annotations: {}

# Pod disruption budget
podDisruptionBudget:
  enabled: false
  minAvailable: 1
  annotations: {}

# Image pull secrets
imagePullSecrets: []
```

## Example Usage

### Basic Deployment

```yaml
# templates/deployment.yaml
{{- include "library.deployment" . -}}
```

### Custom Environment Variables

```yaml
# values.yaml
env:
  - name: LOG_LEVEL
    value: "info"
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: my-secret
        key: database-url
```

### Custom Ports

```yaml
# values.yaml
containerPorts:
  http: 8080
  grpc: 9090
  metrics: 2112
```

### With Ingress

```yaml
# values.yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - hosts:
        - api.example.com
      secretName: api-tls
```

### With Autoscaling

```yaml
# values.yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
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

### With Pod Disruption Budget

```yaml
# values.yaml
podDisruptionBudget:
  enabled: true
  minAvailable: 1
```

## Best Practices

1. **Use Standard Labels**: All resources should use `library.labels` for consistency
2. **Health Checks**: Always enable liveness and readiness probes
3. **Resource Limits**: Set appropriate resource requests and limits
4. **Security Context**: Run containers as non-root with read-only filesystem
5. **Image Tag**: Avoid using `latest` tag in production
6. **Secrets**: Use external secrets management in production

## Versioning

This chart follows [Semantic Versioning](https://semver.org/). Major version changes indicate breaking changes that may require updates to dependent charts.

## License

NEAM Platform Internal
