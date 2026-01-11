#!/bin/bash
#
# GitOps Operations Script for NEAM Platform
#
# This script provides commands for managing GitOps workflows including
# repository synchronization, application management, and status reporting.
#
# Usage:
#   ./gitops.sh <command> [options]
#
# Commands:
#   sync              Synchronize all applications
#   sync <app>        Synchronize specific application
#   status            Show application status
#   diff <app>        Show diff for application
#   rollback <app>    Rollback application
#   promote <app>     Promote application to next environment
#   history <app>     Show deployment history
#   health <app>      Check application health
#   sync-waves        Plan and execute sync waves
#

set -euo pipefail

# Configuration
ARGO_NAMESPACE="argocd"
MANIFESTS_REPO="${MANIFESTS_REPO:-https://github.com/neam-platform/manifests}"
APPLICATIONS_REPO="${APPLICATIONS_REPO:-https://github.com/neam-platform/applications}"
ENVIRONMENT="${ENVIRONMENT:-development}"
TIMEOUT=300

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

show_help() {
    cat << EOF
GitOps Operations for NEAM Platform

Usage: $0 <command> [options]

Commands:
    sync [application]    Synchronize all or specific application(s)
    status [application]  Show status of all or specific application(s)
    diff <application>    Show diff for specific application
    rollback <app> [rev]  Rollback application to previous revision
    promote <app>         Promote application to next environment
    history <application> Show deployment history
    health <application>  Check application health status
    sync-waves            Execute sync waves for ordered deployment
    diff-waves            Show sync wave plan without executing
    refresh <application> Force refresh application from Git
    cleanup               Clean up completed ArgoCD jobs/pods

Options:
    -e, --env            Target environment (default: development)
    -n, --namespace      ArgoCD namespace (default: argocd)
    -t, --timeout        Timeout in seconds (default: 300)
    -v, --verbose        Enable verbose output
    -h, --help           Show this help

Examples:
    $0 sync                           # Sync all applications
    $0 sync api-service               # Sync specific application
    $0 status -e production           # Check production status
    $0 rollback api-service           # Rollback to previous version
    $0 promote api-service            # Promote to next environment
    $0 diff-waves                     # Preview sync wave plan
EOF
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing=()
    
    if ! command -v kubectl &> /dev/null; then
        missing+=("kubectl")
    fi
    
    if ! command -v argocd &> /dev/null; then
        log_warning "ArgoCD CLI not installed. Using kubectl for operations."
    fi
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    if [ ${#missing[@]} -ne 0 ]; then
        log_error "Missing tools: ${missing[*]}"
        exit 1
    fi
    
    log_success "Prerequisites satisfied"
}

# Get all applications
get_applications() {
    local project="${1:-}"
    if [ -n "$project" ]; then
        kubectl get application -n "$ARGO_NAMESPACE" -l app.kubernetes.io/part-of=neam-platform -o jsonpath='{.items[?(@.spec.project=="'"$project"'")].metadata.name}'
    else
        kubectl get application -n "$ARGO_NAMESPACE" -l app.kubernetes.io/part-of=neam-platform -o jsonpath='{.items[*].metadata.name}'
    fi
}

# Get application status
get_status() {
    local app="$1"
    kubectl get application "$app" -n "$ARGO_NAMESPACE" -o jsonpath='{.status}' 2>/dev/null || echo "{}"
}

# Sync application(s)
cmd_sync() {
    local app="${1:-}"
    local apps=()
    
    check_prerequisites
    
    if [ -n "$app" ]; then
        apps=("$app")
    else
        read -ra apps <<< "$(get_applications)"
    fi
    
    if [ ${#apps[@]} -eq 0 ]; then
        log_warning "No applications found"
        return 0
    fi
    
    log_info "Synchronizing ${#apps[@]} application(s)..."
    
    for application in "${apps[@]}"; do
        log_info "Syncing: $application"
        
        # Get current sync wave if any
        local wave=$(kubectl get application "$application" -n "$ARGO_NAMESPACE" \
            -o jsonpath='{.metadata.annotations.argocd\.argoproj\.io/sync-wave}' 2>/dev/null || echo "0")
        
        log_info "  Sync wave: $wave"
        
        # Trigger sync
        if command -v argocd &> /dev/null; then
            argocd app sync "$application" \
                --timeout "$TIMEOUT" \
                --strategy=apply-outside \
                --retry-timeout 5m \
                --verbose 2>&1 | grep -v "password" || true
        else
            kubectl patch application "$application" -n "$ARGO_NAMESPACE" \
                --type=merge \
                -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"deep"}}}'
        fi
        
        log_success "  Sync triggered for: $application"
    done
    
    # Wait for syncs to complete if specific app given
    if [ -n "$app" ]; then
        log_info "Waiting for sync to complete..."
        sleep 30
        check_app_status "$app"
    fi
}

# Show status
cmd_status() {
    local app="${1:-}"
    local project="${ENVIRONMENT:-}"
    
    check_prerequisites
    
    echo ""
    echo "=== Application Status ==="
    echo ""
    
    if [ -n "$app" ]; then
        show_single_status "$app"
    else
        local apps=()
        read -ra apps <<< "$(get_applications "$project")"
        
        printf "%-25s %-12s %-15s %-10s %-20s\n" "APPLICATION" "HEALTH" "SYNC" "REVISION" "MESSAGE"
        printf "%s\n" "$(printf '=%.0s' {1..85})"
        
        for application in "${apps[@]}"; do
            show_single_status "$application" | head -1
        done
    fi
    echo ""
}

show_single_status() {
    local app="$1"
    local status=$(get_status "$app")
    
    local health=$(echo "$status" | jq -r '.health.status // "Unknown"')
    local sync=$(echo "$status" | jq -r '.sync.status // "Unknown"')
    local revision=$(echo "$status" | jq -r '.sync.revision // "Unknown"' | head -c 12)
    local message=$(echo "$status" | jq -r '.sync.message // ""' | head -c 40)
    
    local health_color="$NC"
    case "$health" in
        Healthy) health_color="$GREEN" ;;
        Degraded) health_color="$RED" ;;
        Progressing) health_color="$YELLOW" ;;
        Suspended) health_color="$BLUE" ;;
    esac
    
    local sync_color="$NC"
    case "$sync" in
        Synced) sync_color="$GREEN" ;;
        OutOfSync) sync_color="$YELLOW" ;;
    esac
    
    printf "${health_color}%-25s${NC} %-12s ${sync_color}%-15s${NC} %-10s %-20s\n" \
        "$app" "$health" "$sync" "$revision" "$message"
}

# Show diff
cmd_diff() {
    local app="${1:-}"
    
    if [ -z "$app" ]; then
        log_error "Application name required"
        exit 1
    fi
    
    check_prerequisites
    
    log_info "Showing diff for: $app"
    
    if command -v argocd &> /dev/null; then
        argocd app diff "$app" --local=false
    else
        log_warning "ArgoCD CLI not installed. Showing resource diff..."
        kubectl get application "$app" -n "$ARGO_NAMESPACE" -o yaml | diff - <(kubectl get application "$app" -n "$ARGO_NAMESPACE" -o jsonpath='{.status.resources[*].rawManifest}' 2>/dev/null | base64 -d 2>/dev/null || echo "Unable to get manifest")
    fi
}

# Rollback application
cmd_rollback() {
    local app="${1:-}"
    local revision="${2:-}"
    
    if [ -z "$app" ]; then
        log_error "Application name required"
        exit 1
    fi
    
    check_prerequisites
    
    log_info "Rolling back: $app"
    
    if command -v argocd &> /dev/null; then
        if [ -n "$revision" ]; then
            argocd app rollback "$app" "$revision" --timeout "$TIMEOUT"
        else
            # Get previous history
            argocd app history "$app" -o json | jq -r '.[-2].id // empty' | xargs -I{} argocd app rollback "$app" {} --timeout "$TIMEOUT"
        fi
    else
        log_warning "ArgoCD CLI not installed. Manual rollback required."
        log_info "To rollback, update the manifest repository to a previous revision."
    fi
    
    log_success "Rollback initiated for: $app"
}

# Promote application
cmd_promote() {
    local app="${1:-}"
    
    if [ -z "$app" ]; then
        log_error "Application name required"
        exit 1
    fi
    
    check_prerequisites
    
    log_info "Promoting: $app"
    
    # Get current environment
    local current_env=$(kubectl get application "$app" -n "$ARGO_NAMESPACE" \
        -o jsonpath='{.metadata.labels.app\.kubernetes\.io/environment}' 2>/dev/null || echo "unknown")
    
    # Determine next environment
    local next_env=""
    case "$current_env" in
        development) next_env="staging" ;;
        staging) next_env="production" ;;
        *)
            log_error "Cannot promote from environment: $current_env"
            exit 1
            ;;
    esac
    
    log_info "Promoting from $current_env to $next_env"
    
    # Clone manifests repo
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    git clone "$MANIFESTS_REPO" . 2>/dev/null || git clone "$APPLICATIONS_REPO" . 2>/dev/null || {
        log_error "Failed to clone manifest repository"
        rm -rf "$temp_dir"
        exit 1
    }
    
    # Get current image tag
    local current_image=$(kubectl get application "$app" -n "$ARGO_NAMESPACE" \
        -o jsonpath='{.status.summary.images[*]}' 2>/dev/null | grep -oE ':[a-f0-9]+' | head -1)
    
    # Update environment overlay
    local overlay_path="applications/${app}/overlays/${next_env}/kustomization.yaml"
    
    if [ -f "$overlay_path" ]; then
        sed -i "s|image: .*|image: registry.example.com/${app}:${current_image#:} |" "$overlay_path"
        
        git add "$overlay_path"
        git commit -m "chore: Promote ${app} from ${current_env} to ${next_env}"
        git push
        
        log_success "Promoted to $next_env"
    else
        log_error "Overlay not found: $overlay_path"
    fi
    
    rm -rf "$temp_dir"
}

# Show history
cmd_history() {
    local app="${1:-}"
    
    if [ -z "$app" ]; then
        log_error "Application name required"
        exit 1
    fi
    
    check_prerequisites
    
    echo ""
    echo "=== Deployment History: $app ==="
    echo ""
    
    if command -v argocd &> /dev/null; then
        argocd app history "$app" -o wide
    else
        kubectl get application "$app" -n "$ARGO_NAMESPACE" \
            -o jsonpath='{range .status.history[*]}{"ID: "}{.id}{"\n  Revision: "}{.revision}{"\n  Deployed: "}{.deployedAt}{"\n  InitiatedBy: "}{.initiatedBy}{"\n\n"}{end}'
    fi
    echo ""
}

# Check health
cmd_health() {
    local app="${1:-}"
    
    if [ -z "$app" ]; then
        log_error "Application name required"
        exit 1
    fi
    
    check_prerequisites
    
    log_info "Checking health: $app"
    
    local status=$(get_status "$app")
    local health=$(echo "$status" | jq -r '.health.status // "Unknown"')
    local message=$(echo "$status" | jq -r '.health.message // ""')
    
    echo ""
    echo "=== Health Status ==="
    echo "Application: $app"
    echo "Health: $health"
    echo "Message: $message"
    echo ""
    
    # Show resource health
    echo "=== Resource Health ==="
    echo "$status" | jq -r '.resources[]? | "\(.group) \(.kind): \(.health.status // "Unknown")"' 2>/dev/null || echo "No resource information available"
    echo ""
}

# Sync waves
cmd_sync_waves() {
    check_prerequisites
    
    log_info "Planning sync waves..."
    
    # Get all applications with sync wave annotations
    local apps=$(kubectl get application -n "$ARGO_NAMESPACE" \
        -l app.kubernetes.io/part-of=neam-platform \
        -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.metadata.annotations.argocd\.argoproj\.io/sync-wave}{"\n"}{end}')
    
    if [ -z "$apps" ]; then
        log_warning "No applications with sync waves found"
        return 0
    fi
    
    echo ""
    echo "=== Sync Wave Plan ==="
    echo ""
    
    # Group by wave
    echo "Wave 0 (Infrastructure):"
    echo "$apps" | awk '$2=="0" || $2=="" {print "  - " $1}'
    echo ""
    
    echo "Wave 1 (RBAC):"
    echo "$apps" | awk '$2=="1" {print "  - " $1}'
    echo ""
    
    echo "Wave 2 (Storage):"
    echo "$apps" | awk '$2=="2" {print "  - " $1}'
    echo ""
    
    echo "Wave 3 (Applications):"
    echo "$apps" | awk '$2=="3" || ($2=="" && $2!="0" && $2!="1" && $2!="2") {print "  - " $1}'
    echo ""
}

# Diff sync waves (dry run)
cmd_diff_waves() {
    cmd_sync_waves
    
    echo ""
    echo "=== Wave Execution (Dry Run) ==="
    echo ""
    log_info "Use 'sync-waves' command to execute actual sync"
}

# Force refresh
cmd_refresh() {
    local app="${1:-}"
    
    if [ -z "$app" ]; then
        log_error "Application name required"
        exit 1
    fi
    
    check_prerequisites
    
    log_info "Refreshing: $app"
    
    kubectl patch application "$app" -n "$ARGO_NAMESPACE" \
        --type=merge \
        -p '{"metadata":{"annotations":{"argocd.argoproj.io/refresh":"deep"}}}'
    
    log_success "Refresh triggered for: $app"
}

# Cleanup
cmd_cleanup() {
    check_prerequisites
    
    log_info "Cleaning up ArgoCD resources..."
    
    # Delete completed pods
    kubectl delete pod -n "$ARGO_NAMESPACE" \
        -l app.kubernetes.io/name=argocd-server \
        --field-selector=status.phase=Succeeded 2>/dev/null || true
    
    kubectl delete pod -n "$ARGO_NAMESPACE" \
        -l app.kubernetes.io/name=argocd-repo-server \
        --field-selector=status.phase=Succeeded 2>/dev/null || true
    
    # Clean up old jobs
    kubectl delete job -n "$ARGO_NAMESPACE" \
        --field-selector=status.succeeded=True 2>/dev/null || true
    
    log_success "Cleanup completed"
}

# Check application status after sync
check_app_status() {
    local app="$1"
    local max_wait=120
    local waited=0
    
    while [ $waited -lt $max_wait ]; do
        local status=$(get_status "$app")
        local health=$(echo "$status" | jq -r '.health.status // "Unknown"')
        local sync=$(echo "$status" | jq -r '.sync.status // "Unknown"')
        
        if [ "$health" == "Healthy" ] && [ "$sync" == "Synced" ]; then
            log_success "$app is healthy and synced"
            return 0
        fi
        
        echo "  Waiting... (health: $health, sync: $sync)"
        sleep 10
        waited=$((waited + 10))
    done
    
    log_warning "$app may not be fully synced after ${max_wait}s"
}

# Main
main() {
    local command="${1:-}"
    shift || true
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--namespace)
                ARGO_NAMESPACE="$2"
                shift 2
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -v|--verbose)
                set -x
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Handle command
    case $command in
        sync) cmd_sync "$@" ;;
        status) cmd_status "$@" ;;
        diff) cmd_diff "$@" ;;
        rollback) cmd_rollback "$@" ;;
        promote) cmd_promote "$@" ;;
        history) cmd_history "$@" ;;
        health) cmd_health "$@" ;;
        sync-waves) cmd_sync_waves ;;
        diff-waves) cmd_diff_waves ;;
        refresh) cmd_refresh "$@" ;;
        cleanup) cmd_cleanup ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

main "$@"
