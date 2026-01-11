# Payment Adapter Kubernetes Manifests

This directory contains Kubernetes manifests for deploying the NEAM Payment Adapter
to a production Kubernetes cluster.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.x (optional, for templating)
- kubectl configured with cluster access
- Configured secrets for Redis and Kafka credentials

## Directory Structure

```
payment-adapter/
├── 01-configmap.yaml          # Application configuration
├── 02-secrets.yaml            # Sensitive credentials
├── 03-deployment.yaml         # Deployment and HPA
├── 04-service.yaml            # Service, Ingress, PDB
├── 05-pvc.yaml                # Persistent volume claims
└── 06-prometheus-rules.yaml   # Alerting rules
```

## Quick Deployment

### 1. Create namespace (if not exists)

```bash
kubectl create namespace neam-production
```

### 2. Apply manifests

```bash
# Apply in order (configmaps and secrets must exist before deployment)
kubectl apply -f 01-configmap.yaml
kubectl apply -f 02-secrets.yaml
kubectl apply -f 03-deployment.yaml
kubectl apply -f 04-service.yaml
kubectl apply -f 05-pvc.yaml
kubectl apply -f 06-prometheus-rules.yaml
```

### 3. Verify deployment

```bash
# Check pod status
kubectl get pods -n neam-production -l app=payment-adapter

# Check deployment status
kubectl get deployment payment-adapter -n neam-production

# View logs
kubectl logs -n neam-production -l app=payment-adapter --tail=100
```

## Configuration

### ConfigMap

The ConfigMap (`01-configmap.yaml`) contains all non-sensitive configuration:

- Kafka broker addresses
- Redis connection details
- Processing parameters (worker count, batch size)
- Retry policies
- Logging levels

Update the ConfigMap to match your environment:

```yaml
data:
  KAFKA_BROKERS: "your-kafka-cluster:9092"
  REDIS_ADDR: "your-redis-master:6379"
  WORKER_COUNT: "10"
```

### Secrets

The Secrets manifest (`02-secrets.yaml`) contains sensitive data:

- Redis password
- Kafka SASL credentials
- TLS certificates
- Encryption keys

Create secrets before applying:

```bash
# Using kubectl
kubectl create secret generic payment-adapter-secrets \
  --from-literal=REDIS_PASSWORD='your-password' \
  --from-literal=KAFKA_SASL_USERNAME='your-user' \
  --from-literal=KAFKA_SASL_PASSWORD='your-password' \
  -n neam-production
```

## Scaling

### Manual scaling

```bash
kubectl scale deployment payment-adapter --replicas=6 -n neam-production
```

### Horizontal Pod Autoscaler

The deployment includes an HPA that scales based on CPU and memory utilization:

```yaml
metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Vertical Pod Autoscaler (optional)

For VPA support, add:

```yaml
resources:
  requests:
    cpu: "500m"
    memory: "1Gi"
```

## Monitoring

### Prometheus Rules

The `06-prometheus-rules.yaml` file contains:

- Recording rules for throughput, latency, and lag metrics
- Alert rules for:
  - High consumer lag (>10,000 messages)
  - Low throughput (<1,000 msg/sec)
  - High error rate (>5%)
  - High latency (>500ms p99)
  - Growing DLQ
  - Pod issues

### Grafana Dashboard

Import the dashboard from:
`monitoring/grafana/dashboards/payment-adapter.json`

## High Availability

### Pod Disruption Budget

The PDB ensures at least 2 pods are available during updates:

```yaml
spec:
  minAvailable: 2
```

### Anti-Affinity

Pods are distributed across nodes:

```yaml
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app
              operator: In
              values:
                - payment-adapter
        topologyKey: kubernetes.io/hostname
```

## Troubleshooting

### Check pod status

```bash
kubectl get pods -n neam-production -l app=payment-adapter -o wide
```

### View logs

```bash
kubectl logs -n neam-production -l app=payment-adapter --tail=1000 -f
```

### Check events

```bash
kubectl describe pod -n neam-production -l app=payment-adapter
```

### Port forward for local testing

```bash
kubectl port-forward -n neam-production svc/payment-adapter 8084:8084 9090:9090
```

### Restart deployment

```bash
kubectl rollout restart deployment/payment-adapter -n neam-production
```

## Resource Recommendations

### Development

```yaml
resources:
  requests:
    cpu: "500m"
    memory: "1Gi"
  limits:
    cpu: "1000m"
    memory: "2Gi"
replicas: 1
```

### Staging

```yaml
resources:
  requests:
    cpu: "1000m"
    memory: "2Gi"
  limits:
    cpu: "2000m"
    memory: "4Gi"
replicas: 2
```

### Production

```yaml
resources:
  requests:
    cpu: "1000m"
    memory: "2Gi"
  limits:
    cpu: "2000m"
    memory: "4Gi"
replicas: 3-10 (autoscaling)
```

## Performance Tuning

### For 5,000+ msg/sec throughput

```yaml
env:
  - name: WORKER_COUNT
    value: "10"
  - name: BATCH_SIZE
    value: "100"
  - name: GOMAXPROCS
    value: "4"
  - name: GOMEMLIMIT
    value: "4GiB"

resources:
  requests:
    cpu: "1500m"
    memory: "3Gi"
  limits:
    cpu: "3000m"
    memory: "6Gi"
```

### Kafka tuning

```yaml
env:
  - name: KAFKA_MAX_OPEN_REQUESTS
    value: "100"
  - name: KAFKA_SESSION_TIMEOUT
    value: "30s"
```

## Upgrade Procedure

### Rolling update

```bash
kubectl set image deployment/payment-adapter \
  payment-adapter=registry.neam.io/payment-adapter:v1.2.1 \
  -n neam-production

kubectl rollout status deployment/payment-adapter -n neam-production
```

### Blue-green deployment (optional)

For zero-downtime updates, use a blue-green deployment strategy.

## Rollback Procedure

### Rollback to previous version

```bash
kubectl rollout undo deployment/payment-adapter -n neam-production
```

### View rollout history

```bash
kubectl rollout history deployment/payment-adapter -n neam-production
```

## Security Considerations

- Run as non-root user (uid 1000)
- Use read-only filesystem where possible
- Restrict network policies
- Enable TLS for Kafka and Redis
- Use secrets management (Vault, AWS Secrets Manager)
- Regular vulnerability scanning of container image
