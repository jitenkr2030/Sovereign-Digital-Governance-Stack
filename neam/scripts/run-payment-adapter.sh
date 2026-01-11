#!/usr/bin/env bash
#
# NEAM Payment Adapter - Quick Start and Validation Script
#
# This script provides quick commands to start, test, and validate
# the Payment Adapter service in various environments.
#
# Usage:
#   ./scripts/run-payment-adapter.sh [command] [options]
#
# Commands:
#   start       Start the Payment Adapter service
#   stop        Stop the Payment Adapter service
#   status      Check service status
#   logs        View service logs
#   test        Run unit tests
#   benchmark   Run performance benchmark
#   integration Run integration tests
#   health      Check service health
#

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PAYMENT_ADAPTER_DIR="$PROJECT_ROOT/sensing/payments"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yml"

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

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing=()

    # Check Docker
    if ! command -v docker &> /dev/null; then
        missing+=("docker")
    fi

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        missing+=("docker-compose")
    fi

    # Check Go
    if ! command -v go &> /dev/null; then
        missing+=("go")
    fi

    if [ ${#missing[@]} -ne 0 ]; then
        log_error "Missing prerequisites: ${missing[*]}"
        exit 1
    fi

    log_success "All prerequisites satisfied"
}

# Start the Payment Adapter service
cmd_start() {
    log_info "Starting NEAM Payment Adapter..."

    # Check if service is already running
    if docker ps --format '{{.Names}}' | grep -q "neam-payment-adapter"; then
        log_warn "Payment Adapter is already running"
        return 0
    fi

    # Start via docker-compose
    cd "$PROJECT_ROOT"
    docker-compose up -d neam-payment-adapter

    # Wait for service to be healthy
    log_info "Waiting for service to be healthy..."
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -sf http://localhost:8084/health > /dev/null 2>&1; then
            log_success "Payment Adapter is healthy and ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done

    log_error "Payment Adapter failed to become healthy within ${max_attempts}s"
    exit 1
}

# Stop the Payment Adapter service
cmd_stop() {
    log_info "Stopping NEAM Payment Adapter..."

    cd "$PROJECT_ROOT"
    docker-compose stop neam-payment-adapter 2>/dev/null || true

    log_success "Payment Adapter stopped"
}

# Check service status
cmd_status() {
    log_info "Checking Payment Adapter status..."

    # Check Docker status
    local container_status=$(docker inspect -f '{{.State.Status}}' neam-payment-adapter 2>/dev/null || echo "not_found")

    case $container_status in
        running)
            log_success "Container is running"
            ;;
        exited)
            log_warn "Container has exited"
            ;;
        not_found)
            log_info "Container not found"
            ;;
        *)
            log_warn "Unknown status: $container_status"
            ;;
    esac

    # Check health endpoint
    if curl -sf http://localhost:8084/health > /dev/null 2>&1; then
        log_success "Health endpoint responding"
    else
        log_warn "Health endpoint not responding"
    fi

    # Check metrics endpoint
    if curl -sf http://localhost:9090/metrics > /dev/null 2>&1; then
        log_success "Metrics endpoint responding"
    else
        log_warn "Metrics endpoint not responding"
    fi
}

# View service logs
cmd_logs() {
    local lines=${1:-100}

    log_info "Showing last $lines lines of logs..."

    if docker ps --format '{{.Names}}' | grep -q "neam-payment-adapter"; then
        docker logs --tail $lines -f neam-payment-adapter
    else
        log_warn "Payment Adapter container not running"
    fi
}

# Run unit tests
cmd_test() {
    log_info "Running Payment Adapter unit tests..."

    cd "$PAYMENT_ADAPTER_DIR"

    # Run tests with coverage
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

    # Calculate coverage
    if [ -f coverage.out ]; then
        local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        log_info "Test coverage: $coverage"

        if [ -f "$PROJECT_ROOT/monitoring/grafana/dashboards/payment-adapter.json" ]; then
            echo "Coverage report saved to coverage.out"
        fi
    fi
}

# Run performance benchmark
cmd_benchmark() {
    log_info "Running performance benchmark..."

    # Check if required tools are available
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is required for benchmarking"
        exit 1
    fi

    # Install dependencies if needed
    pip3 install -q kafka-python prometheus-api-client tqdm 2>/dev/null || true

    # Run benchmark
    cd "$PROJECT_ROOT/tests/benchmark"
    python3 benchmark_throughput.py \
        --brokers "${KAFKA_BROKERS:-localhost:9092}" \
        --num-messages 50000 \
        --num-producers 5 \
        --target-throughput 5000
}

# Run integration tests
cmd_integration() {
    log_info "Running integration tests..."

    # Check if required tools are available
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 is required for integration tests"
        exit 1
    fi

    # Install dependencies if needed
    pip3 install -q kafka-python redis 2>/dev/null || true

    # Start infrastructure if not running
    if ! docker ps --format '{{.Names}}' | grep -q "neam-kafka"; then
        log_info "Starting Kafka and Redis..."
        cd "$PROJECT_ROOT"
        docker-compose up -d kafka redis
        sleep 10  # Wait for services to be ready
    fi

    # Run integration tests
    cd "$PROJECT_ROOT/tests/integration"
    python3 test_kafka_redis.py \
        --brokers "${KAFKA_BROKERS:-localhost:9092}" \
        --redis "${REDIS_HOST:-localhost}"
}

# Check service health
cmd_health() {
    log_info "Performing health check..."

    local checks=0
    local passed=0

    # Check 1: Process running
    ((checks++))
    if docker ps --format '{{.Names}}' | grep -q "neam-payment-adapter"; then
        log_success "[1/4] Container is running"
        ((passed++))
    else
        log_error "[1/4] Container is not running"
    fi

    # Check 2: Health endpoint
    ((checks++))
    if curl -sf http://localhost:8084/health > /dev/null 2>&1; then
        log_success "[2/4] Health endpoint responding"
        ((passed++))
    else
        log_error "[2/4] Health endpoint not responding"
    fi

    # Check 3: Metrics endpoint
    ((checks++))
    if curl -sf http://localhost:9090/metrics > /dev/null 2>&1; then
        log_success "[3/4] Metrics endpoint responding"
        ((passed++))
    else
        log_error "[3/4] Metrics endpoint not responding"
    fi

    # Check 4: Redis connection
    ((checks++))
    if command -v redis-cli &> /dev/null; then
        if redis-cli -h "${REDIS_HOST:-localhost}" ping > /dev/null 2>&1; then
            log_success "[4/4] Redis connection successful"
            ((passed++))
        else
            log_error "[4/4] Redis connection failed"
        fi
    else
        log_warn "[4/4] redis-cli not available, skipping Redis check"
        ((passed++))
    fi

    log_info "Health check: $passed/$checks checks passed"

    if [ $passed -eq $checks ]; then
        log_success "All health checks passed"
        return 0
    else
        log_error "Some health checks failed"
        return 1
    fi
}

# Build the Payment Adapter
cmd_build() {
    log_info "Building Payment Adapter..."

    cd "$PAYMENT_ADAPTER_DIR"

    # Build binary
    go build -o payment-adapter main.go

    # Build Docker image
    cd "$PROJECT_ROOT"
    docker build -t neam/payment-adapter:latest -f sensing/payments/Dockerfile .

    log_success "Payment Adapter built successfully"
}

# Deploy to Kubernetes
cmd_deploy_k8s() {
    log_info "Deploying to Kubernetes..."

    local namespace="${NAMESPACE:-neam-production}"

    # Apply manifests
    kubectl apply -f "$PROJECT_ROOT/deployments/k8s/payment-adapter/"

    # Wait for deployment
    log_info "Waiting for deployment to be ready..."
    kubectl rollout status deployment/payment-adapter -n "$namespace"

    log_success "Payment Adapter deployed to Kubernetes"
}

# Show help
cmd_help() {
    echo "NEAM Payment Adapter - Quick Start and Validation Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  start        Start the Payment Adapter service"
    echo "  stop         Stop the Payment Adapter service"
    echo "  status       Check service status"
    echo "  logs         View service logs (use 'tail -f' style)"
    echo "  test         Run unit tests with coverage"
    echo "  benchmark    Run performance benchmark (5000+ msg/sec target)"
    echo "  integration  Run Kafka and Redis integration tests"
    echo "  health       Perform comprehensive health check"
    echo "  build        Build Payment Adapter binary and Docker image"
    echo "  deploy_k8s   Deploy to Kubernetes"
    echo "  help         Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  KAFKA_BROKERS   Kafka broker addresses (default: localhost:9092)"
    echo "  REDIS_HOST      Redis host (default: localhost)"
    echo "  NAMESPACE       Kubernetes namespace (default: neam-production)"
}

# Main entry point
main() {
    local command=${1:-help}

    case $command in
        start|stop|status|logs|test|benchmark|integration|health|build|deploy_k8s|help)
            cmd_"$command" "${@:2}"
            ;;
        *)
            log_error "Unknown command: $command"
            cmd_help
            exit 1
            ;;
    esac
}

main "$@"
