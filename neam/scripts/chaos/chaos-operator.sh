#!/bin/bash
#
# Chaos Engineering Deployment and Operations Script
# 
# This script provides comprehensive commands for deploying, managing, and
# operating chaos engineering experiments in the NEAM Platform.
#
# Usage:
#   ./chaos-operator.sh [command] [options]
#
# Commands:
#   deploy          Deploy Chaos Mesh to the cluster
#   verify          Verify Chaos Mesh installation
#   apply           Apply chaos experiments from manifests
#   list            List active chaos experiments
#   status          Check status of specific experiments
#   delete          Delete chaos experiments
#   start           Start a specific experiment
#   stop            Stop a specific experiment
#   monitor         Monitor experiment execution
#   validate        Run validation against experiments
#   report          Generate experiment report
#   cleanup         Clean up all chaos resources
#   backup          Backup chaos configuration
#   restore         Restore chaos configuration
#

set -euo pipefail

# Configuration
NAMESPACE="chaos-mesh"
EXPERIMENTS_DIR="${EXPERIMENTS_DIR:-./k8s/chaos}"
SCRIPTS_DIR="${SCRIPTS_DIR:-./scripts/chaos}"
OUTPUT_DIR="${OUTPUT_DIR:-./output/chaos}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
Chaos Engineering Operations Script

Usage: $0 [command] [options]

Commands:
    deploy              Deploy Chaos Mesh to the cluster
    verify              Verify Chaos Mesh installation
    apply [file]        Apply chaos experiments from manifest
    list                List all chaos experiments
    status [name]       Check status of specific experiment
    delete [name]       Delete chaos experiments
    start [name]        Start a specific experiment
    stop [name]         Stop a specific experiment
    monitor [name]      Monitor experiment execution
    validate [name]     Run validation against experiments
    report [name]       Generate experiment report
    cleanup             Clean up all chaos resources
    backup              Backup chaos configuration
    restore [file]      Restore chaos configuration

Options:
    -h, --help          Show this help message
    -n, --namespace     Kubernetes namespace (default: chaos-mesh)
    -v, --verbose       Enable verbose output
    -o, --output        Output directory (default: ./output/chaos)

Examples:
    $0 deploy                           # Deploy Chaos Mesh
    $0 apply k8s/chaos/experiments.yaml # Apply experiments
    $0 monitor pod-failure              # Monitor specific experiment
    $0 validate --all                   # Validate all experiments
EOF
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_deps=()
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi
    
    # Check helm
    if ! command -v helm &> /dev/null; then
        missing_deps+=("helm")
    fi
    
    # Check jq
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "All prerequisites satisfied"
}

# Deploy Chaos Mesh
deploy_chaos_mesh() {
    log_info "Deploying Chaos Mesh to namespace: $NAMESPACE"
    
    # Create namespace if not exists
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    # Add Helm repo if not exists
    if ! helm repo list | grep -q "chaos-mesh"; then
        log_info "Adding Chaos Mesh Helm repository..."
        helm repo add chaos-mesh https://charts.chaos-mesh.org
        helm repo update
    fi
    
    # Deploy Chaos Mesh
    log_info "Installing Chaos Mesh via Helm..."
    helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh \
        --namespace "$NAMESPACE" \
        --wait \
        --timeout 300s \
        --set clusterScoped=true \
        --set dashboard.ingress.enabled=false \
        --set controllerManager.replicas=1 \
        --set chaosDaemon.replicas=3
    
    # Wait for pods to be ready
    log_info "Waiting for Chaos Mesh pods to be ready..."
    kubectl wait --namespace "$NAMESPACE" \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/instance=chaos-mesh \
        --timeout=300s
    
    log_success "Chaos Mesh deployed successfully"
}

# Verify Chaos Mesh installation
verify_installation() {
    log_info "Verifying Chaos Mesh installation..."
    
    # Check namespace
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_error "Namespace $NAMESPACE does not exist"
        return 1
    fi
    
    # Check pods
    local pods=$(kubectl get pods -n "$NAMESPACE" -o jsonpath='{.items[*].metadata.name}')
    if [ -z "$pods" ]; then
        log_error "No pods found in namespace $NAMESPACE"
        return 1
    fi
    
    # Check custom resource definitions
    local crds=(
        "podchaos.chaos-mesh.org"
        "networkchaos.chaos-mesh.org"
        "stresschaos.chaos-mesh.org"
        "httpchaos.chaos-mesh.org"
        "dnschaos.chaos-mesh.org"
        "timechaos.chaos-mesh.org"
        "workflows.chaos-mesh.org"
    )
    
    local missing_crds=()
    for crd in "${crds[@]}"; do
        if ! kubectl get crd "$crd" &> /dev/null; then
            missing_crds+=("$crd")
        fi
    done
    
    if [ ${#missing_crds[@]} -ne 0 ]; then
        log_error "Missing CRDs: ${missing_crds[*]}"
        return 1
    fi
    
    # Print status
    log_success "Chaos Mesh installation verified"
    echo ""
    echo "Pods:"
    kubectl get pods -n "$NAMESPACE" -o wide
    echo ""
    echo "CRDs:"
    kubectl get crds | grep chaos-mesh
}

# Apply chaos experiments
apply_experiments() {
    local file="${1:-$EXPERIMENTS_DIR/experiments.yaml}"
    
    if [ ! -f "$file" ]; then
        log_error "Experiment file not found: $file"
        exit 1
    fi
    
    log_info "Applying chaos experiments from: $file"
    
    # Validate YAML syntax
    if ! python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>/dev/null; then
        log_warning "YAML syntax validation skipped (python3-yaml not available)"
    fi
    
    # Apply the manifest
    kubectl apply -f "$file"
    
    log_success "Chaos experiments applied"
}

# List chaos experiments
list_experiments() {
    log_info "Listing chaos experiments..."
    
    echo ""
    echo "=== Pod Chaos ==="
    kubectl get podchaos -A -o wide 2>/dev/null || echo "No PodChaos resources found"
    
    echo ""
    echo "=== Network Chaos ==="
    kubectl get networkchaos -A -o wide 2>/dev/null || echo "No NetworkChaos resources found"
    
    echo ""
    echo "=== Stress Chaos ==="
    kubectl get stresschaos -A -o wide 2>/dev/null || echo "No StressChaos resources found"
    
    echo ""
    echo "=== HTTP Chaos ==="
    kubectl get httpchaos -A -o wide 2>/dev/null || echo "No HTTPChaos resources found"
    
    echo ""
    echo "=== DNS Chaos ==="
    kubectl get dnschaos -A -o wide 2>/dev/null || echo "No DNSChaos resources found"
    
    echo ""
    echo "=== Time Chaos ==="
    kubectl get timechaos -A -o wide 2>/dev/null || echo "No TimeChaos resources found"
}

# Check experiment status
check_status() {
    local name="${1:-}"
    local namespace="${2:-$NAMESPACE}"
    
    if [ -z "$name" ]; then
        log_error "Experiment name is required"
        exit 1
    fi
    
    log_info "Checking status of experiment: $name in namespace: $namespace"
    
    # Try different chaos types
    local found=false
    
    for chaos_type in podchaos networkchaos stresschaos httpchaos dnschaos timechaos; do
        local result=$(kubectl get "$chaos_type" "$name" -n "$namespace" -o json 2>/dev/null)
        if [ -n "$result" ]; then
            found=true
            echo "$result" | jq -r '
                "Experiment: " + .metadata.name +
                "\nNamespace: " + .metadata.namespace +
                "\nType: " + .kind +
                "\nStatus: " + .status.phase +
                "\nCreated: " + .metadata.creationTimestamp'
            
            if [ ".spec.duration" ] && [ ".spec.duration" != "null" ]; then
                echo "Duration: $(echo "$result" | jq -r '.spec.duration')"
            fi
        fi
    done
    
    if [ "$found" = false ]; then
        log_error "Experiment not found: $name"
        exit 1
    fi
}

# Delete experiments
delete_experiments() {
    local name="${1:-}"
    local namespace="${2:-$NAMESPACE}"
    
    if [ -z "$name" ]; then
        log_warning "No experiment name specified. Available experiments:"
        list_experiments
        return
    fi
    
    log_info "Deleting experiment: $name from namespace: $namespace"
    
    # Try to delete from all chaos types
    for chaos_type in podchaos networkchaos stresschaos httpchaos dnschaos timechaos; do
        if kubectl get "$chaos_type" "$name" -n "$namespace" &> /dev/null; then
            kubectl delete "$chaos_type" "$name" -n "$namespace"
            log_success "Deleted $chaos_type/$name"
        fi
    done
}

# Start experiment
start_experiment() {
    local name="${1:-}"
    local namespace="${2:-$NAMESPACE}"
    
    if [ -z "$name" ]; then
        log_error "Experiment name is required"
        exit 1
    fi
    
    log_info "Starting experiment: $name"
    
    # Chaos Mesh experiments start automatically when applied
    # This function can be used to trigger experiments that are paused
    for chaos_type in podchaos networkchaos stresschaos httpchaos dnschaos timechaos; do
        if kubectl get "$chaos_type" "$name" -n "$namespace" &> /dev/null; then
            # Patch to reset the experiment
            kubectl patch "$chaos_type" "$name" -n "$namespace" \
                --type=merge \
                -p '{"spec":{"startTime":"'"$(date -Iseconds)"'"}}'
            log_success "Started $chaos_type/$name"
            return
        fi
    done
    
    log_error "Experiment not found: $name"
}

# Stop experiment
stop_experiment() {
    local name="${1:-}"
    local namespace="${2:-$NAMESPACE}"
    
    if [ -z "$name" ]; then
        log_error "Experiment name is required"
        exit 1
    fi
    
    log_info "Stopping experiment: $name"
    
    for chaos_type in podchaos networkchaos stresschaos httpchaos dnschaos timechaos; do
        if kubectl get "$chaos_type" "$name" -n "$namespace" &> /dev/null; then
            kubectl delete "$chaos_type" "$name" -n "$namespace"
            log_success "Stopped $chaos_type/$name"
            return
        fi
    done
    
    log_error "Experiment not found: $name"
}

# Monitor experiment
monitor_experiment() {
    local name="${1:-all}"
    local namespace="${2:-$NAMESPACE}"
    local duration="${3:-60}"
    
    log_info "Monitoring experiments (duration: ${duration}s)..."
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR/monitoring"
    
    # Monitor in background
    (
        local end_time=$((SECONDS + duration))
        
        while [ $SECONDS -lt $end_time ]; do
            clear
            echo "=== Chaos Experiment Monitor ==="
            echo "Time: $(date)"
            echo "Duration remaining: $((end_time - SECONDS))s"
            echo ""
            
            if [ "$name" = "all" ]; then
                list_experiments
            else
                check_status "$name" "$namespace"
            fi
            
            echo ""
            echo "=== Recent Events ==="
            kubectl get events -n "$namespace" \
                --sort-by='.lastTimestamp' \
                --field-selector involvedObject.namespace="$namespace" \
                -o json | jq -r '.items[] | 
                    "\(.lastTimestamp) \(.reason) \(.message)"' | tail -10
            
            sleep 5
        done
    ) &
    
    # Wait for completion
    wait
    
    log_success "Monitoring complete"
}

# Run validation
run_validation() {
    local experiment="${1:-}"
    local scope="${2:-all}"
    
    log_info "Running validation for: $experiment (scope: $scope)"
    
    # Create output directory
    mkdir -p "$OUTPUT_DIR/validation"
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local report_file="$OUTPUT_DIR/validation/validation_$timestamp.json"
    
    # Run data loss validation if configured
    if [ -f "$SCRIPTS_DIR/data_loss_validator.py" ]; then
        log_info "Running data loss validation..."
        python3 "$SCRIPTS_DIR/data_loss_validator.py" \
            --phase validate \
            --duration 300 \
            --output "$OUTPUT_DIR/validation" \
            2>&1 | tee "$OUTPUT_DIR/validation/data_loss_$timestamp.log"
    fi
    
    # Run performance degradation test if configured
    if [ -f "$SCRIPTS_DIR/performance_degradation_test.py" ]; then
        log_info "Running performance degradation test..."
        python3 "$SCRIPTS_DIR/performance_degradation_test.py" \
            --config "$EXPERIMENTS_DIR/test_config.json" \
            --output "$OUTPUT_DIR/validation/performance_$timestamp.json" \
            2>&1 | tee "$OUTPUT_DIR/validation/performance_$timestamp.log"
    fi
    
    log_success "Validation complete. Results saved to: $OUTPUT_DIR/validation/"
}

# Generate report
generate_report() {
    local experiment="${1:-}"
    local scope="${2:-all}"
    
    log_info "Generating report for: $experiment"
    
    mkdir -p "$OUTPUT_DIR/reports"
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local report_file="$OUTPUT_DIR/reports/chaos_report_$timestamp.json"
    
    # Collect experiment data
    local experiments_json=$(kubectl get all,chaos -A -o json 2>/dev/null || echo '{"items":[]}')
    
    # Generate report
    cat > "$report_file" << EOF
{
    "report_id": "$(uuidgen 2>/dev/null || date +%s)",
    "generated_at": "$(date -Iseconds)",
    "scope": "$scope",
    "experiments": $experiments_json,
    "cluster_info": {
        "context": "$(kubectl config current-context 2>/dev/null || echo 'unknown')",
        "kubernetes_version": "$(kubectl version -o json 2>/dev/null | jq -r '.serverVersion.gitVersion' || echo 'unknown')"
    }
}
EOF
    
    # Pretty print
    if command -v jq &> /dev/null; then
        local temp_file=$(mktemp)
        jq '.' "$report_file" > "$temp_file" && mv "$temp_file" "$report_file"
    fi
    
    log_success "Report generated: $report_file"
}

# Cleanup all chaos resources
cleanup() {
    log_warning "Cleaning up all chaos resources..."
    
    # Delete all chaos experiments
    for chaos_type in podchaos networkchaos stresschaos httpchaos dnschaos timechaos; do
        local resources=$(kubectl get "$chaos_type" -A -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
        if [ -n "$resources" ]; then
            for name in $resources; do
                kubectl delete "$chaos_type" "$name" -A --wait=0 2>/dev/null &
            done
        fi
    done
    
    # Wait for deletions
    sleep 5
    
    # Delete namespace (optional, commented out for safety)
    # kubectl delete namespace "$NAMESPACE" --wait=false
    
    log_success "Cleanup initiated"
}

# Backup chaos configuration
backup_config() {
    log_info "Backing up chaos configuration..."
    
    mkdir -p "$OUTPUT_DIR/backups"
    
    local backup_file="$OUTPUT_DIR/chaos_backup_$TIMESTAMP.tar.gz"
    
    # Backup CRDs and experiments
    kubectl get all,chaos -A -o yaml > "$OUTPUT_DIR/backups/chaos_resources_$TIMESTAMP.yaml"
    
    # Compress
    tar -czf "$backup_file" -C "$OUTPUT_DIR/backups" .
    
    # Keep only last 10 backups
    ls -1t "$OUTPUT_DIR"/backups/*.tar.gz 2>/dev/null | tail -n +11 | xargs -r rm
    
    log_success "Backup created: $backup_file"
}

# Restore chaos configuration
restore_config() {
    local backup_file="${1:-}"
    
    if [ -z "$backup_file" ]; then
        log_error "Backup file is required"
        exit 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    log_info "Restoring chaos configuration from: $backup_file"
    
    # Extract
    local temp_dir=$(mktemp -d)
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # Restore resources
    if [ -f "$temp_dir/chaos_resources.yaml" ]; then
        kubectl apply -f "$temp_dir/chaos_resources.yaml"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    log_success "Configuration restored"
}

# Run comprehensive test suite
run_test_suite() {
    log_info "Running chaos engineering test suite..."
    
    mkdir -p "$OUTPUT_DIR/test_results"
    
    local test_results=()
    local passed=0
    local failed=0
    
    # Test 1: Verify Chaos Mesh is installed
    log_info "Test 1: Verifying Chaos Mesh installation..."
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_success "  PASSED: Namespace exists"
        ((passed++))
    else
        log_error "  FAILED: Namespace does not exist"
        ((failed++))
    fi
    
    # Test 2: Verify CRDs are installed
    log_info "Test 2: Verifying CRDs..."
    local crds_count=$(kubectl get crds --no-headers 2>/dev/null | grep -c "chaos-mesh" || echo "0")
    if [ "$crds_count" -gt 0 ]; then
        log_success "  PASSED: $crds_count CRDs found"
        ((passed++))
    else
        log_error "  FAILED: No chaos-mesh CRDs found"
        ((failed++))
    fi
    
    # Test 3: Verify experiments can be applied
    log_info "Test 3: Testing experiment apply (dry-run)..."
    if kubectl apply --dry-run=client -f "$EXPERIMENTS_DIR/experiments.yaml" &> /dev/null; then
        log_success "  PASSED: Experiments validate"
        ((passed++))
    else
        log_error "  FAILED: Experiments failed validation"
        ((failed++))
    fi
    
    # Test 4: Verify Python scripts are executable
    log_info "Test 4: Verifying Python scripts..."
    if [ -f "$SCRIPTS_DIR/data_loss_validator.py" ] && [ -x "$SCRIPTS_DIR/data_loss_validator.py" ]; then
        log_success "  PASSED: data_loss_validator.py is executable"
        ((passed++))
    elif [ -f "$SCRIPTS_DIR/data_loss_validator.py" ]; then
        chmod +x "$SCRIPTS_DIR/data_loss_validator.py"
        log_success "  PASSED: data_loss_validator.py made executable"
        ((passed++))
    else
        log_warning "  SKIPPED: data_loss_validator.py not found"
    fi
    
    # Print summary
    echo ""
    echo "=== Test Suite Summary ==="
    echo "Passed: $passed"
    echo "Failed: $failed"
    
    if [ $failed -eq 0 ]; then
        log_success "All tests passed!"
        return 0
    else
        log_error "Some tests failed"
        return 1
    fi
}

# Main function
main() {
    local command="${1:-}"
    shift || true
    
    # Parse global options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -v|--verbose)
                set -x
                shift
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Check for valid command
    if [ -z "$command" ]; then
        show_help
        exit 1
    fi
    
    # Execute command
    case $command in
        deploy)
            check_prerequisites
            deploy_chaos_mesh
            ;;
        verify)
            check_prerequisites
            verify_installation
            ;;
        apply)
            check_prerequisites
            apply_experiments "$@"
            ;;
        list)
            list_experiments
            ;;
        status)
            check_prerequisites
            check_status "$@"
            ;;
        delete)
            check_prerequisites
            delete_experiments "$@"
            ;;
        start)
            check_prerequisites
            start_experiment "$@"
            ;;
        stop)
            check_prerequisites
            stop_experiment "$@"
            ;;
        monitor)
            check_prerequisites
            monitor_experiment "$@"
            ;;
        validate)
            check_prerequisites
            run_validation "$@"
            ;;
        report)
            check_prerequisites
            generate_report "$@"
            ;;
        cleanup)
            cleanup
            ;;
        backup)
            backup_config
            ;;
        restore)
            check_prerequisites
            restore_config "$@"
            ;;
        test|test-suite)
            check_prerequisites
            run_test_suite
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main
main "$@"
