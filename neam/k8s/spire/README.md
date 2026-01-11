# Zero Trust Security Implementation

This directory contains the complete implementation of Zero Trust security architecture for the NEAM platform, including SPIFFE/SPIRE identity infrastructure, Istio service mesh mTLS configuration, and Calico network policies.

## Overview

The Zero Trust implementation consists of three main pillars:

1. **Identity (SPIFFE/SPIRE)**: Workload identity through SPIFFE IDs and automatic certificate management
2. **Encryption (Istio mTLS)**: Mutual TLS authentication for all service-to-service communication
3. **Segmentation (Calico)**: Network policies implementing default-deny and namespace isolation

## Directory Structure

```
k8s/
├── spire/
│   ├── server/
│   │   ├── configmap.yaml          # Server configuration and bootstrap entries
│   │   └── statefulset.yaml        # SPIRE Server deployment (3 replicas)
│   └── agent/
│       └── daemonset.yaml          # SPIRE Agent on all worker nodes
├── istio/
│   ├── mtls/
│   │   └── peerauthentication.yaml # STRICT mTLS policies by namespace
│   └── authorization/
│       └── authorization-policies.yaml # Service-level authorization rules
└── policies/
    └── calico/
        ├── global-policies.yaml    # Cluster-wide network policies
        └── namespace-policies.yaml # Namespace-specific policies

monitoring/
└── dashboards/
    └── zero-trust-overview.yaml    # Grafana dashboard for Zero Trust metrics

scripts/
└── spire/
    ├── deploy-zero-trust.sh        # Automated deployment script
    └── spire-manage.sh             # SPIRE management utility
```

## Prerequisites

- Kubernetes cluster (v1.24+)
- Helm 3.x
- kubectl configured with cluster access
- Istio service mesh installed
- Calico CNI installed
- SPIRE images available (gcr.io/spiffe-io/spire-server, spire-agent)

## Quick Start

### 1. Deploy Complete Zero Trust Infrastructure

```bash
cd scripts/spire
chmod +x deploy-zero-trust.sh
./deploy-zero-trust.sh deploy
```

### 2. Verify Deployment

```bash
./deploy-zero-trust.sh verify
```

### 3. Register Workload Entries

```bash
# Register a new workload
./spire-manage.sh create-entry spiffe://neam.internal/myapp/api my-namespace my-service-account 24h

# List all entries
./spire-manage.sh list-entries
```

## SPIRE Server Configuration

### Trust Domain
- **Trust Domain**: `neam.internal`
- **Cluster Name**: `neam-cluster`
- **SVID TTL**: 24 hours (workload), 72 hours (agents)

### Server Architecture
- 3 replicas in StatefulSet
- Anti-affinity for high availability
- SQLite database with persistent storage
- Kubernetes PSAT node attestor

### Registration Entry Format

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: spire-entries
  namespace: spire
data:
  entries.yaml: |
    - spiffe_id: "spiffe://neam.internal/<namespace>/<service-type>/<service-name>"
      parent_id: "spiffe://neam.internal/spire-server"
      selectors:
        - k8s:ns:<namespace>
        - k8s:sa:<service-account>
      ttl: "24h"
```

## mTLS Configuration

### PeerAuthentication Modes

All namespaces use `STRICT` mTLS mode:

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default-mtls
  namespace: <namespace>
spec:
  mtls:
    mode: STRICT
```

### Authorization Policy Pattern

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-<service>
  namespace: <namespace>
spec:
  selector:
    matchLabels:
      app: <service-name>
  action: ALLOW
  rules:
    - from:
        - source:
            principals: ["spiffe://neam.internal/<source-namespace>/api/<source-service>"]
      to:
        - operation:
            methods: ["GET", "POST"]
            paths: ["/api/v1/*"]
```

## Network Policies

### Default Deny Architecture

All namespaces implement default-deny:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: <namespace>
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
  ingress: []
  egress: []
```

### Allowed Traffic Patterns

1. **DNS Resolution**: All namespaces allow DNS (port 53 UDP/TCP)
2. **SPIRE Access**: All namespaces allow access to SPIRE agents
3. **Monitoring**: Monitoring namespace can scrape all workloads
4. **Istio Control Plane**: All workloads communicate with Istiod
5. **Namespace Internal**: Communication within the same namespace

## Workload Registration

### Kubernetes-Based Registration

Workloads are registered using Kubernetes service account selectors:

```bash
# Register a workload
kubectl exec -n spire spire-server-0 -- spire-server entry create \
  -spiffeID spiffe://neam.internal/finance/api/payment-processor \
  -parentID spiffe://neam.internal/spire-server \
  -selector k8s:ns:finance \
  -selector k8s:sa:payment-processor \
  -ttl 24h
```

### Supported Selectors

- `k8s:ns:<namespace>` - Kubernetes namespace
- `k8s:sa:<service-account>` - Kubernetes service account
- `k8s:pod-label:<key>=<value>` - Pod labels
- `docker:image-id:<digest>` - Container image digest

## Monitoring and Observability

### Zero Trust Dashboard

Grafana dashboard providing visibility into:

- SPIRE server and agent status
- SVID issuance metrics
- mTLS connection statistics
- Network policy enforcement
- Security event tracking

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| spiffe_agent_up | SPIRE agent availability | < 100% |
| spiffe_svid_expiry | Certificate expiration time | < 6 hours |
| istio_tcp_mcx_denied_connections | mTLS denials | > 0 |
| calico_v3_policy_dropped_packets | Policy drops | > 100/min |

### Security Alerts

1. **Critical**: mTLS denials detected
2. **High**: Attestation failures, certificate expiration
3. **Medium**: Policy violations, unusual traffic patterns

## Operational Procedures

### Certificate Rotation

Certificates are automatically rotated every 24 hours. Manual rotation:

```bash
# Force rotate certificates for a namespace
./spire-manage.sh rotate-entries <namespace>
```

### Entry Management

```bash
# List all entries
./spire-manage.sh list-entries

# Delete an entry
./spire-manage.sh delete-entry <entry-id>

# Export entries for backup
./spire-manage.sh export-entries
```

### Troubleshooting

```bash
# Check SPIRE server health
./spire-manage.sh health-check

# View agent status
./spire-manage.sh agents-status

# View trust bundle
./spire-manage.sh view-bundle
```

## Security Considerations

### Defense in Depth

1. **SPIRE**: Workload identity through attestation
2. **Istio**: mTLS encryption and authorization
3. **Calico**: Network-level isolation
4. **Kubernetes**: Pod security policies

### Default Security Posture

- All traffic denied by default
- Explicit allow required for all communication
- Short certificate lifetimes (24 hours)
- No internet egress without explicit policy
- No privileged containers by default

### Compliance

This implementation supports compliance with:

- PCI DSS (requirement 1.3.6, 4.1, 7.1, 8.3)
- SOC 2 (CC6.1, CC6.3, CC6.6, CC6.7)
- NIST 800-53 (AC-3, AC-4, SC-8, SC-13)
- HIPAA (164.312(a)(1), 164.312(e)(1))

## Troubleshooting Guide

### Common Issues

#### SPIRE Server Not Ready
```bash
# Check server logs
kubectl logs -n spire -l app=spire-server --tail=100

# Check resource usage
kubectl top pods -n spire -l app=spire-server
```

#### Agent Attestation Failures
```bash
# Check agent logs
kubectl logs -n spire -l app=spire-agent --tail=100

# Verify node selectors
kubectl get nodes --show-labels | grep -i spire
```

#### mTLS Connection Failures
```bash
# Check AuthorizationPolicy
kubectl get authorizationpolicies -A

# Check Istio proxy status
istioctl proxy-status

# Check mTLS configuration
istioctl x authz check <pod-name>.<namespace>
```

#### Network Policy Issues
```bash
# Check Calico policies
calicoctl get globalnetworkpolicy

# Check namespace policies
kubectl get networkpolicies -A

# View policy hits
kubectl logs -n calico-system -l k8s-app=calico-node | grep Policy
```

## Performance Considerations

### Resource Requirements

| Component | CPU | Memory |
|-----------|-----|--------|
| SPIRE Server (per replica) | 500m - 2000m | 512Mi - 1Gi |
| SPIRE Agent (per node) | 100m - 500m | 128Mi - 512Mi |

### Scaling Recommendations

- SPIRE Server: 3 replicas for high availability
- SPIRE Agent: DaemonSet on all worker nodes
- Typha: 3 replicas for clusters > 100 nodes

### Performance Tuning

1. Increase agent worker threads for high-workload clusters
2. Use connection pooling for SQLite database
3. Monitor certificate issuance latency
4. Tune Calico FELIX parameters for throughput

## Upgrades

### SPIRE Upgrade Procedure

1. Create backup of registration entries:
   ```bash
   ./spire-manage.sh export-entries
   ```

2. Update SPIRE server StatefulSet with new image
   ```bash
   kubectl set image statefulset spire-server -n spire \
     spire-server=gcr.io/spiffe-io/spire-server:<new-version>
   ```

3. Wait for rollout to complete:
   ```bash
   kubectl rollout status statefulset spire-server -n spire
   ```

4. Update SPIRE agent DaemonSet:
   ```bash
   kubectl set image daemonset spire-agent -n spire \
     spire-agent=gcr.io/spiffe-io/spire-agent:<new-version>
   ```

5. Verify deployment:
   ```bash
   ./deploy-zero-trust.sh verify
   ```

## Contributing

1. Follow the security-by-default principle
2. Document all new policies
3. Include monitoring for new controls
4. Test in staging before production
5. Review for least-privilege compliance

## License

This implementation is part of the NEAM platform and follows the project's licensing terms.
