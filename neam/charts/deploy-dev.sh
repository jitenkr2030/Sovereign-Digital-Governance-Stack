#!/bin/bash
#
# NEAM Platform Development Environment Deployment Script
# This script deploys all NEAM Platform services to development environment
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHARTS_DIR="${SCRIPT_DIR}/.."
ENV_DIR="${CHARTS_DIR}/environments"
NAMESPACE="neam-dev"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_header() {
    echo ""
    echo "=============================================="
    echo "$1"
    echo "=============================================="
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "  ℹ $1"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    if ! command -v helm &> /dev/null; then
        print_error "Helm is not installed"
        exit 1
    fi
    print_success "Helm is installed"
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed"
        exit 1
    fi
    print_success "kubectl is installed"
    
    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    print_success "Connected to Kubernetes cluster"
    
    # Check if namespace exists
    if ! kubectl get namespace "${NAMESPACE}" &> /dev/null; then
        print_info "Creating namespace: ${NAMESPACE}"
        kubectl create namespace "${NAMESPACE}"
    else
        print_success "Namespace ${NAMESPACE} exists"
    fi
}

# Update Helm dependencies
update_dependencies() {
    print_header "Updating Helm Dependencies"
    
    # Update library chart
    print_info "Updating library chart dependencies..."
    cd "${CHARTS_DIR}/library"
    helm dependency update
    print_success "Library chart dependencies updated"
    
    # Update all service charts
    local services=(
        "sensing"
        "black-economy"
        "feature-store"
        "macro"
        "intelligence"
        "policy"
        "reporting"
        "intervention"
    )
    
    for service in "${services[@]}"; do
        print_info "Updating ${service} chart dependencies..."
        cd "${CHARTS_DIR}/${service}"
        helm dependency update
        print_success "${service} chart dependencies updated"
    done
}

# Deploy a service
deploy_service() {
    local service_name="$1"
    local chart_path="${CHARTS_DIR}/${service_name}"
    local env_values="${ENV_DIR}/dev/values.yaml"
    local service_values="${chart_path}/values.yaml"
    
    print_info "Deploying ${service_name}..."
    
    # Combine environment values with service values
    # Environment values take precedence
    helm upgrade --install "${service_name}" "${chart_path}" \
        --namespace "${NAMESPACE}" \
        --values "${service_values}" \
        --values "${env_values}" \
        --wait \
        --timeout 5m \
        --debug
        
    if [ $? -eq 0 ]; then
        print_success "${service_name} deployed successfully"
    else
        print_error "Failed to deploy ${service_name}"
        return 1
    fi
}

# Deploy all services
deploy_all() {
    print_header "Deploying All Services"
    
    local services=(
        "sensing"
        "black-economy"
        "feature-store"
        "macro"
        "intelligence"
        "policy"
        "reporting"
        "intervention"
    )
    
    for service in "${services[@]}"; do
        deploy_service "${service}" || true
    done
}

# Verify deployment
verify_deployment() {
    print_header "Verifying Deployment"
    
    print_info "Checking pod status..."
    kubectl get pods -n "${NAMESPACE}" -l "app.kubernetes.io/part-of=neam-platform"
    
    print_info "Checking services..."
    kubectl get svc -n "${NAMESPACE}" -l "app.kubernetes.io/part-of=neam-platform"
    
    print_info "Checking deployments..."
    kubectl get deployment -n "${NAMESPACE}" -l "app.kubernetes.io/part-of=neam-platform"
}

# Print access information
print_access_info() {
    print_header "Access Information"
    
    echo ""
    echo "Development environment deployed successfully!"
    echo ""
    echo "To access services, use kubectl port-forward:"
    echo ""
    echo "# Example: Access sensing service"
    echo "kubectl port-forward -n ${NAMESPACE} svc/neam-sensing 8080:8080"
    echo ""
    echo "# Then open in browser"
    echo "http://localhost:8080"
    echo ""
    echo "To check logs:"
    echo "kubectl logs -n ${NAMESPACE} deployment/neam-sensing -f"
    echo ""
    echo "To check status:"
    echo "kubectl get all -n ${NAMESPACE}"
}

# Main function
main() {
    echo "NEAM Platform Development Deployment"
    echo "====================================="
    
    check_prerequisites
    update_dependencies
    deploy_all
    verify_deployment
    print_access_info
}

main "$@"
