# Quick Deployment Guide for NEAM Platform Helm Charts
# This guide covers common deployment scenarios

## Prerequisites

- Helm 3.0+
- kubectl configured with cluster access
- Access to container registry
- AWS credentials (if using IAM roles)

## Directory Structure

```
charts/
├── library/                 # Shared library chart
├── sensing/                 # Sensing service
├── black-economy/          # Black economy detection
├── feature-store/          # Feature store
├── macro/                  # Macro analysis
├── intelligence/           # AI/ML serving
├── policy/                 # Policy management
├── reporting/              # Reporting service
├── intervention/           # Automated responses
├── environments/           # Environment values
│   ├── dev/
│   ├── staging/
│   └── prod/
├── deploy-dev.sh          # Development deployment
├── deploy-staging.sh      # Staging deployment
├── deploy-prod.sh         # Production deployment
├── validate-charts.sh      # Validation script
└── kustomization.yaml      # Kustomize overlay
```

## Quick Start

### 1. Validate Charts

```bash
# Make validation script executable
chmod +x validate-charts.sh

# Run validation
./validate-charts.sh
```

### 2. Development Deployment

```bash
# Deploy to development
chmod +x deploy-dev.sh
./deploy-dev.sh

# Access services via port-forward
kubectl port-forward -n neam-dev svc/neam-sensing 8080:8080
```

### 3. Staging Deployment

```bash
# Deploy to staging
chmod +x deploy-staging.sh
./deploy-staging.sh
```

### 4. Production Deployment

```bash
# Deploy to production (requires confirmation)
chmod +x deploy-prod.sh
./deploy-prod.sh
```

## Environment-Specific Configuration

### Development
- Single replica per service
- Resource limits relaxed
- No ingress (use port-forward)
- Dev image tags

### Staging
- 2 replicas per service
- Medium resource limits
- Ingress enabled with staging certificates
- Staging image tags

### Production
- 3+ replicas per service
- High resource limits
- Full ingress with production certificates
- Version-tagged images
- Pod disruption budgets
- Horizontal pod autoscaling

## Customizing Deployments

### Using Custom Values

```bash
# Deploy with custom values
helm upgrade --install my-release ./sensing \
  --values sensing/values.yaml \
  --values my-custom-values.yaml \
  --namespace neam-platform
```

### Using Environment Values

```bash
# Deploy to specific environment
helm upgrade --install sensing ./sensing \
  --values ./environments/dev/values.yaml \
  --namespace neam-dev
```

### Overriding Individual Values

```bash
# Override specific values
helm upgrade --install sensing ./sensing \
  --set replicaCount=5 \
  --set image.tag="v1.2.0" \
  --set resources.limits.cpu="4" \
  --namespace neam-platform
```

## Common Operations

### View Rendered Templates

```bash
# Render without installing
helm template my-release ./sensing
```

### Dry-Run Install

```bash
# Test installation without applying
helm upgrade --install my-release ./sensing \
  --dry-run \
  --debug
```

### Upgrade Deployment

```bash
# Upgrade with new values
helm upgrade my-release ./sensing \
  --values ./environments/prod/values.yaml \
  --set image.tag="v1.1.0"
```

### Rollback

```bash
# List revisions
helm history my-release

# Rollback to previous revision
helm rollback my-release 1
```

### Uninstall

```bash
# Uninstall release
helm uninstall my-release
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n neam-platform

# Check pod logs
kubectl logs <pod-name> -n neam-platform
```

### Image Pull Errors

```bash
# Verify image pull secrets
kubectl get secrets -n neam-platform

# Check image exists
kubectl describe deployment <deployment-name> -n neam-platform | grep -i image
```

### Resource Issues

```bash
# Check resource usage
kubectl top pods -n neam-platform

# Check resource requests/limits
kubectl get deployment <deployment-name> -n neam-platform -o yaml | grep -A 5 resources
```

### Network Issues

```bash
# Check service endpoints
kubectl get endpoints <service-name> -n neam-platform

# Test service connectivity
kubectl exec -it <pod-name> -n neam-platform -- curl <service>:8080/health
```

## Best Practices

1. **Always validate before deployment**: Run `./validate-charts.sh` first
2. **Use specific image tags**: Avoid `latest` in production
3. **Set resource limits**: Prevent resource contention
4. **Enable health checks**: Use liveness and readiness probes
5. **Configure PDBs**: Ensure high availability
6. **Use autoscaling**: Handle variable load
7. **Monitor deployments**: Use Prometheus and Grafana
8. **Backup before upgrades**: Keep rollback capability

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy to Staging
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Validate Charts
        run: ./validate-charts.sh
        
      - name: Deploy to Staging
        run: ./deploy-staging.sh
        env:
          KUBECONFIG: ${{ secrets.KUBECONFIG }}
```

## Support

For issues or questions:
- Check the library chart README
- Review the LIBRARY_INFRASTRUCTURE.md documentation
- Contact the Platform team
