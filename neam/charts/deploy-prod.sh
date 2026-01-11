#!/bin/bash
#
# NEAM Platform Production Environment Deployment Script
# This script deploys all NEAM Platform services to production environment
# WARNING: This deploys to production - ensure you have reviewed all changes
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHARTS_DIR="${SCRIPT_DIR}/.."
ENV_DIR="${CHARTS_DIR}/environments"
NAMESPACE="neam-prod"

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

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Confirmation prompt
confirm_deployment() {
    print_header "Production Deployment Confirmation"
    
    echo -e "${RED}WARNING: You are about to deploy to PRODUCTION environment${NC}"
    echo ""
    echo "Please confirm the following:"
    echo ""
    echo "1. You have reviewed all changes"
    echo "2. You have tested in staging environment"
    echo "3. You have backup of current deployment"
    echo "4. You have notified relevant teams"
    echo ""
    
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "${confirm}" != "yes" ]; then
        echo "Deployment cancelled."
        exit 0
    fi
    
    read -p "Type 'deploy' to confirm: " confirm2
    
    if [ "${confirm2}" != "deploy" ]; then
        echo "Deployment cancelled."
        exit 0
    fi
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
    
    # Check context
    print_info "Current context: $(kubectl config current-context)"
    
    # Check if namespace exists
    if ! kubectl get namespace "${NAMESPACE}" &> /dev/null; then
        print_error "Namespace ${NAMESPACE} does not exist"
        print_info "Create it with: kubectl create namespace ${NAMESPACE}"
        exit 1
    fi
    print_success "Namespace ${NAMESPACE} exists"
}

# Backup current deployment
backup_current_deployment() {
    print_header "Backing Up Current Deployment"
    
    print_info "Creating backup of current deployment..."
    kubectl get all -n "${NAMESPACE}" -o yaml > "${CHARTS_DIR}/backup-${NAMESPACE}-$(date +%Y%m%d-%H%M%S).yaml"
    print_success "Backup created"
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

# Dry run deployment
dry_run_deployment() {
    print_header "Performing Dry Run Deployment"
    
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
        print_info "Dry running ${service}..."
        cd "${CHARTS_DIR}/${service}"
        helm template "${service}" . --values "${ENV_DIR}/prod/values.yaml" --dry-run --debug 2>&1 | head -50
    done
}

# Deploy a service with rollout status
deploy_service() {
    local service_name="$1"
    local chart_path="${CHARTS_DIR}/${service_name}"
    local env_values="${ENV_DIR}/prod/values.yaml"
    local service_values="${chart_path}/values.yaml"
    
    print_info "Deploying ${service_name}..."
    
    # Combine environment values with service values
    # Production values take precedence
    helm upgrade --install "${service_name}" "${chart_path}" \
        --namespace "${NAMESPACE}" \
        --values "${service_values}" \
        --values "${env_values}" \
        --wait \
        --timeout 10m \
        --atomic \
        --cleanup-on-fail
        
    if [ $? -eq 0 ]; then
        print_success "${service_name} deployed successfully"
        
        # Wait for rollout
        print_info "Waiting for ${service_name} rollout..."
        kubectl rollout status deployment/"${service_name}" -n "${NAMESPACE}" --timeout=10m
    else
        print_error "Failed to deploy ${service_name}"
        return 1
    fi
}

# Deploy with canary strategy
deploy_canary() {
    local service_name="$1"
    local chart_path="${CHARTS_DIR}/${service_name}"
    local env_values="${ENV_DIR}/prod/values.yaml"
    
    print_info "Deploying ${service_name} with canary strategy..."
    
    # Deploy with lower replicas first
    helm upgrade --install "${service_name}-canary" "${chart_path}" \
        --namespace "${NAMESPACE}" \
        --set "replicaCount=1" \
        --set "autoscaling.enabled=false" \
        --values "${chart_path}/values.yaml" \
        --values "${env_values}" \
        --wait \
        --timeout 5m \
        --debug
        
    if [ $? -ne 0 ]; then
        print_error "Canary deployment failed for ${service_name}"
        return 1
    fi
    
    print_success "Canary deployed for ${service_name}"
    print_info "Manual verification required before full rollout"
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
    
    print_info "Checking horizontal pod autoscalers..."
    kubectl get hpa -n "${NAMESPACE}" -l "app.kubernetes.io/part-of=neam-platform"
    
    print_info "Checking pod disruption budgets..."
    kubectl get pdb -n "${NAMESPACE}" -l "app.kubernetes.io/part-of=neam-platform"
}

# Health check
health_check() {
    print_header "Performing Health Check"
    
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
        print_info "Checking ${service} health..."
        local pod=$(kubectl get pods -n "${NAMESPACE}" -l "app.kubernetes.io/name=${service}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
        
        if [ -n "${pod}" ]; then
            local status=$(kubectl get pod "${pod}" -n "${NAMESPACE}" -o jsonpath='{.status.phase}' 2>/dev/null)
            if [ "${status}" == "Running" ]; then
                print_success "${service} is running"
            else
                print_error "${service} is not running (status: ${status})"
            fi
        else
            print_error "${service} pod not found"
        fi
    done
}

# Print access information
print_access_info() {
    print_header "Production Deployment Complete"
    
    echo ""
    echo "Production environment deployed successfully!"
    echo ""
    echo "External access:"
    echo "- Sensing: https://sensing.neam.example.com"
    echo "- Macro: https://macro.neam.example.com"
    echo "- Reporting: https://reports.neam.example.com"
    echo ""
    echo "To monitor deployment:"
    echo "kubectl get all -n ${NAMESPACE}"
    echo "kubectl get pods -n ${NAMESPACE} -w"
    echo ""
    echo "To check logs:"
    echo "kubectl logs -n ${NAMESPACE} deployment/neam-sensing -f"
    echo ""
    echo "To rollback if needed:"
    echo "helm rollback neam-sensing -n ${NAMESPACE}"
}

# Main function
main() {
    echo "NEAM Platform Production Deployment"
    echo "====================================="
    echo ""
    echo "This will deploy to: ${NAMESPACE}"
    echo "Date: $(date)"
    echo ""
    
    confirm_deployment
    check_prerequisites
    backup_current_deployment
    update_dependencies
    dry_run_deployment
    deploy_all
    verify_deployment
    health_check
    print_access_info
}

main "$@"
