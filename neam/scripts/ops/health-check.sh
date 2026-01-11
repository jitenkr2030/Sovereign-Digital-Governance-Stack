#!/bin/bash
# NEAM Health Check Script
# Verifies all platform components are operational

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NEAM_NAMESPACE=${NEAM_NAMESPACE:-neam-system}
TIMEOUT=${TIMEOUT:-30}
API_ENDPOINT=${API_ENDPOINT:-http://localhost:8080}

# Counters
PASSED=0
FAILED=0
WARNINGS=0

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

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check Kubernetes namespace
check_namespace() {
    log_info "Checking Kubernetes namespace: $NEAM_NAMESPACE"

    if command_exists kubectl; then
        if kubectl get namespace "$NEAM_NAMESPACE" >/dev/null 2>&1; then
            log_success "Namespace $NEAM_NAMESPACE exists"
        else
            log_fail "Namespace $NEAM_NAMESPACE not found"
            return 1
        fi
    else
        log_warning "kubectl not available, skipping namespace check"
    fi
}

# Check pod status
check_pods() {
    log_info "Checking NEAM pods status"

    if command_exists kubectl; then
        local not_ready=$(kubectl get pods -n "$NEAM_NAMESPACE" --no-headers 2>/dev/null | grep -v "Running" | grep -v "Completed" | wc -l)

        if [ "$not_ready" -eq 0 ]; then
            log_success "All pods are running or completed"
        else
            log_fail "$not_ready pods are not in running state"
            kubectl get pods -n "$NEAM_NAMESPACE" 2>/dev/null | grep -v "Running" | grep -v "Completed"
        fi

        # Show pod count
        local total=$(kubectl get pods -n "$NEAM_NAMESPACE" --no-headers 2>/dev/null | wc -l)
        log_info "Total pods: $total"
    else
        log_warning "kubectl not available, skipping pod check"
    fi
}

# Check service endpoints
check_services() {
    log_info "Checking service endpoints"

    local services=("neam-api-gateway" "neam-sensing" "neam-intelligence" "neam-intervention" "neam-reporting" "neam-security" "neam-frontend")

    for svc in "${services[@]}"; do
        if command_exists kubectl; then
            local endpoints=$(kubectl get endpoints "$svc" -n "$NEAM_NAMESPACE" -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null)
            if [ -n "$endpoints" ]; then
                log_success "Service $svc has endpoints"
            else
                log_fail "Service $svc has no endpoints"
            fi
        fi
    done
}

# Check API health endpoint
check_api_health() {
    log_info "Checking API health endpoint: $API_ENDPOINT/health"

    if command_exists curl; then
        local response=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$API_ENDPOINT/health" 2>/dev/null || echo "000")

        if [ "$response" = "200" ]; then
            log_success "API health endpoint returns 200"
        else
            log_fail "API health endpoint returns $response"
        fi
    else
        log_warning "curl not available, skipping API health check"
    fi
}

# Check database connectivity
check_databases() {
    log_info "Checking database connectivity"

    # PostgreSQL check
    if command_exists kubectl; then
        local pg_ready=$(kubectl exec -n "$NEAM_NAMESPACE" neam-postgresql-0 -- pg_isready -U neam_user -q 2>/dev/null && echo "ok" || echo "fail")
        if [ "$pg_ready" = "ok" ]; then
            log_success "PostgreSQL is ready"
        else
            log_fail "PostgreSQL is not ready"
        fi

        # ClickHouse check
        local ch_ready=$(kubectl exec -n "$NEAM_NAMESPACE" neam-clickhouse-0 -- clickhouse-client -u neam_user -q "SELECT 1" 2>/dev/null && echo "ok" || echo "fail")
        if [ "$ch_ready" = "ok" ]; then
            log_success "ClickHouse is ready"
        else
            log_fail "ClickHouse is not ready"
        fi

        # Redis check
        local redis_ready=$(kubectl exec -n "$NEAM_NAMESPACE" neam-redis-0 -- redis-cli ping 2>/dev/null | grep -q "PONG" && echo "ok" || echo "fail")
        if [ "$redis_ready" = "ok" ]; then
            log_success "Redis is ready"
        else
            log_fail "Redis is not ready"
        fi
    else
        log_warning "kubectl not available, skipping database checks"
    fi
}

# Check Kafka connectivity
check_kafka() {
    log_info "Checking Kafka connectivity"

    if command_exists kubectl; then
        local kafka_ready=$(kubectl exec -n "$NEAM_NAMESPACE" kafka-0 -- kafka-broker-api-versions --bootstrap-server localhost:9092 2>/dev/null && echo "ok" || echo "fail")
        if [ "$kafka_ready" = "ok" ]; then
            log_success "Kafka is ready"
        else
            log_fail "Kafka is not ready"
        fi
    else
        log_warning "kubectl not available, skipping Kafka check"
    fi
}

# Check disk usage
check_disk_usage() {
    log_info "Checking disk usage"

    if command_exists df; then
        local usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
        if [ "$usage" -lt 80 ]; then
            log_success "Disk usage at ${usage}%"
        elif [ "$usage" -lt 90 ]; then
            log_warning "Disk usage at ${usage}%"
        else
            log_fail "Disk usage critical at ${usage}%"
        fi
    fi
}

# Check memory usage
check_memory_usage() {
    log_info "Checking memory usage"

    if command_exists free; then
        local usage=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}')
        if [ "$usage" -lt 80 ]; then
            log_success "Memory usage at ${usage}%"
        elif [ "$usage" -lt 90 ]; then
            log_warning "Memory usage at ${usage}%"
        else
            log_fail "Memory usage critical at ${usage}%"
        fi
    fi
}

# Print summary
print_summary() {
    echo ""
    echo "========================================"
    echo "Health Check Summary"
    echo "========================================"
    echo -e "Passed:   ${GREEN}$PASSED${NC}"
    echo -e "Failed:   ${RED}$FAILED${NC}"
    echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
    echo "========================================"

    if [ "$FAILED" -gt 0 ]; then
        echo -e "${RED}Health check FAILED${NC}"
        exit 1
    elif [ "$WARNINGS" -gt 0 ]; then
        echo -e "${YELLOW}Health check completed with warnings${NC}"
        exit 0
    else
        echo -e "${GREEN}Health check PASSED${NC}"
        exit 0
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "NEAM Platform Health Check"
    echo "========================================"
    echo "Namespace: $NEAM_NAMESPACE"
    echo "API Endpoint: $API_ENDPOINT"
    echo "Timeout: ${TIMEOUT}s"
    echo "========================================"
    echo ""

    check_namespace
    check_pods
    check_services
    check_api_health
    check_databases
    check_kafka
    check_disk_usage
    check_memory_usage

    print_summary
}

# Run main function
main "$@"
