#!/bin/bash
# Zero Trust Infrastructure Deployment Script
# This script deploys the complete Zero Trust security infrastructure

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="spire"
ISTIO_NAMESPACE="istio-system"
MONITORING_NAMESPACE="monitoring"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_tools=()

    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi

    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
    fi

    if ! command -v calicoctl &> /dev/null; then
        log_warn "calicoctl not found - some Calico operations may be limited"
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi

    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    log_info "Prerequisites check passed"
}

# Create namespaces
create_namespaces() {
    log_info "Creating namespaces..."

    kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
    kubectl label namespace "${NAMESPACE}" app=spire tier=infrastructure --overwrite

    kubectl create namespace "${ISTIO_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
    kubectl label namespace "${ISTIO_NAMESPACE}" app=istio tier=infrastructure --overwrite

    kubectl create namespace "${MONITORING_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
    kubectl label namespace "${MONITORING_NAMESPACE}" app=monitoring tier=infrastructure --overwrite

    log_info "Namespaces created"
}

# Deploy SPIRE Server
deploy_spire_server() {
    log_info "Deploying SPIRE Server..."

    local spire_dir="${SCRIPT_DIR}/../k8s/spire/server"

    # Apply ConfigMaps
    kubectl apply -f "${spire_dir}/configmap.yaml"

    # Apply RBAC and ServiceAccount
    kubectl apply -f "${spire_dir}/statefulset.yaml"

    # Wait for SPIRE server to be ready
    log_info "Waiting for SPIRE server to be ready..."
    kubectl wait --for=condition=ready pod -l app=spire-server -n "${NAMESPACE}" --timeout=300s

    log_info "SPIRE Server deployed successfully"
}

# Deploy SPIRE Agent
deploy_spire_agent() {
    log_info "Deploying SPIRE Agent..."

    local spire_dir="${SCRIPT_DIR}/../k8s/spire/agent"

    # Apply ConfigMaps
    kubectl apply -f "${spire_dir}/daemonset.yaml"

    # Wait for SPIRE agents to be ready
    log_info "Waiting for SPIRE agents to be ready..."
    kubectl rollout status daemonset spire-agent -n "${NAMESPACE}" --timeout=300s

    log_info "SPIRE Agent deployed successfully"
}

# Register SPIRE entries
register_spire_entries() {
    log_info "Registering SPIRE workload entries..."

    local spire_dir="${SCRIPT_DIR}/../k8s/spire/server"

    # Use spire-server CLI to register entries from bootstrap config
    kubectl exec -n "${NAMESPACE}" spire-server-0 -- \
        spire-server entry create \
        -spiffeID spiffe://neam.internal/infrastructure/monitoring/prometheus \
        -parentID spiffe://neam.internal/spire-server \
        -selector k8s:ns:monitoring \
        -selector k8s:sa:prometheus \
        -ttl 24h

    kubectl exec -n "${NAMESPACE}" spire-server-0 -- \
        spire-server entry create \
        -spiffeID spiffe://neam.internal/core/api-gateway \
        -parentID spiffe://neam.internal/spire-server \
        -selector k8s:ns:core \
        -selector k8s:sa:api-gateway \
        -ttl 24h

    kubectl exec -n "${NAMESPACE}" spire-server-0 -- \
        spire-server entry create \
        -spiffeID spiffe://neam.internal/core/auth-service \
        -parentID spiffe://neam.internal/spire-server \
        -selector k8s:ns:core \
        -selector k8s:sa:auth-service \
        -ttl 24h

    kubectl exec -n "${NAMESPACE}" spire-server-0 -- \
        spire-server entry create \
        -spiffeID spiffe://neam.internal/finance/api/payment-processor \
        -parentID spiffe://neam.internal/spire-server \
        -selector k8s:ns:finance \
        -selector k8s:sa:payment-processor \
        -ttl 24h

    log_info "SPIRE workload entries registered"
}

# Deploy Istio mTLS
deploy_istio_mtls() {
    log_info "Deploying Istio mTLS configuration..."

    local istio_dir="${SCRIPT_DIR}/../k8s/istio"

    # Apply PeerAuthentication
    kubectl apply -f "${istio_dir}/mtls/peerauthentication.yaml"

    # Apply Authorization Policies
    kubectl apply -f "${istio_dir}/authorization/authorization-policies.yaml"

    log_info "Istio mTLS configuration deployed"
}

# Deploy Calico Network Policies
deploy_calico_policies() {
    log_info "Deploying Calico Network Policies..."

    local calico_dir="${SCRIPT_DIR}/../k8s/policies/calico"

    # Apply global policies
    kubectl apply -f "${calico_dir}/global-policies.yaml"

    # Apply namespace policies
    kubectl apply -f "${calico_dir}/namespace-policies.yaml"

    log_info "Calico Network Policies deployed"
}

# Deploy monitoring dashboards
deploy_monitoring() {
    log_info "Deploying monitoring dashboards..."

    local monitoring_dir="${SCRIPT_DIR}/../monitoring"

    # Apply ConfigMap with dashboard
    kubectl apply -f "${monitoring_dir}/dashboards/zero-trust-overview.yaml"

    log_info "Monitoring dashboards deployed"
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."

    # Check SPIRE server
    log_info "Checking SPIRE server..."
    kubectl get pods -n "${NAMESPACE}" -l app=spire-server
    kubectl get svc -n "${NAMESPACE}" -l app=spire-server

    # Check SPIRE agents
    log_info "Checking SPIRE agents..."
    kubectl get pods -n "${NAMESPACE}" -l app=spire-agent
    kubectl get daemonset spire-agent -n "${NAMESPACE}"

    # Check Istio mTLS
    log_info "Checking Istio mTLS policies..."
    kubectl get peerauthentication -A
    kubectl get authorizationpolicies -A | head -20

    # Check Calico policies
    log_info "Checking Calico policies..."
    kubectl get globalnetworkpolicies
    kubectl get networkpolicies -A | head -20

    log_info "Deployment verification complete"
}

# Print status summary
print_summary() {
    log_info "========================================"
    log_info "Zero Trust Infrastructure Deployment Complete"
    log_info "========================================"
    log_info ""
    log_info "SPIRE Server:"
    log_info "  - Namespace: ${NAMESPACE}"
    log_info "  - Trust Domain: neam.internal"
    log_info "  - Replicas: 3"
    log_info ""
    log_info "SPIRE Agent:"
    log_info "  - Namespace: ${NAMESPACE}"
    log_info "  - DaemonSet: All worker nodes"
    log_info ""
    log_info "Istio mTLS:"
    log_info "  - Namespace: ${ISTIO_NAMESPACE}"
    log_info "  - Mode: STRICT (all namespaces)"
    log_info ""
    log_info "Calico Network Policies:"
    log_info "  - Default Deny: Enabled"
    log_info "  - Namespace Isolation: Configured"
    log_info ""
    log_info "Monitoring:"
    log_info "  - Namespace: ${MONITORING_NAMESPACE}"
    log_info "  - Dashboard: Zero Trust Security Overview"
    log_info ""
    log_info "Next steps:"
    log_info "  1. Register additional workload entries as needed"
    log_info "  2. Configure namespace-specific policies"
    log_info "  3. Review and tune authorization policies"
    log_info "  4. Monitor Zero Trust dashboard for anomalies"
}

# Main deployment function
main() {
    log_info "Starting Zero Trust Infrastructure Deployment"
    log_info "============================================"

    check_prerequisites
    create_namespaces
    deploy_spire_server
    deploy_spire_agent
    register_spire_entries
    deploy_istio_mtls
    deploy_calico_policies
    deploy_monitoring
    verify_deployment
    print_summary
}

# Parse arguments
case "${1:-deploy}" in
    deploy)
        main
        ;;
    spire-server)
        check_prerequisites
        create_namespaces
        deploy_spire_server
        ;;
    spire-agent)
        check_prerequisites
        deploy_spire_agent
        ;;
    istio-mtls)
        check_prerequisites
        deploy_istio_mtls
        ;;
    calico)
        check_prerequisites
        deploy_calico_policies
        ;;
    monitoring)
        check_prerequisites
        deploy_monitoring
        ;;
    verify)
        verify_deployment
        ;;
    *)
        echo "Usage: $0 {deploy|spire-server|spire-agent|istio-mtls|calico|monitoring|verify}"
        echo ""
        echo "Commands:"
        echo "  deploy        - Deploy complete Zero Trust infrastructure"
        echo "  spire-server  - Deploy only SPIRE Server"
        echo "  spire-agent   - Deploy only SPIRE Agent"
        echo "  istio-mtls    - Deploy only Istio mTLS configuration"
        echo "  calico        - Deploy only Calico network policies"
        echo "  monitoring    - Deploy only monitoring dashboards"
        echo "  verify        - Verify deployment status"
        exit 1
        ;;
esac
