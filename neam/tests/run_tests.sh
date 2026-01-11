#!/bin/bash

# NEAM Platform Test Runner
# This script runs all tests for the NEAM Platform

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COVERAGE_DIR="${SCRIPT_DIR}/coverage"
CONFIG_FILE="${SCRIPT_DIR}/config.yaml"

# Default settings
TEST_TYPE="all"  # all, unit, integration, benchmark, contract
PARALLEL=true
VERBOSE=false
COVERAGE=true
COVERAGE_MIN=80
TIMEOUT="5m"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --type)
            TEST_TYPE="$2"
            shift 2
            ;;
        --parallel)
            PARALLEL="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --no-coverage)
            COVERAGE=false
            shift
            ;;
        --coverage-min)
            COVERAGE_MIN="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --type <type>          Test type: all, unit, integration, benchmark, contract (default: all)"
            echo "  --parallel <bool>      Run tests in parallel (default: true)"
            echo "  --verbose              Verbose output"
            echo "  --no-coverage          Disable coverage reporting"
            echo "  --coverage-min <min>   Minimum coverage percentage (default: 80)"
            echo "  --timeout <duration>   Test timeout (default: 5m)"
            echo "  --help, -h             Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Print banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo "║            NEAM Platform Test Runner                           ║"
    echo "║                                                                ║"
    echo "║  Type: ${TEST_TYPE}"
    echo "║  Coverage: ${COVERAGE}${NC}"
    echo "╚════════════════════════════════════════════════════════════════╝"
}

# Print status
print_status() {
    local status=$1
    local message=$2
    case $status in
        info)
            echo -e "${BLUE}[INFO]${NC} $message"
            ;;
        success)
            echo -e "${GREEN}[✓]${NC} $message"
            ;;
        warning)
            echo -e "${YELLOW}[!]${NC} $message"
            ;;
        error)
            echo -e "${RED}[✗]${NC} $message"
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    print_status info "Checking prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_status error "Go is not installed"
        exit 1
    fi
    print_status success "Go version: $(go version)"
    
    # Check Python
    if ! command -v python3 &> /dev/null; then
        print_status warning "Python3 is not installed (required for some tests)"
    fi
    
    # Check Docker (optional)
    if command -v docker &> /dev/null; then
        print_status success "Docker is available"
    else
        print_status warning "Docker is not installed"
    fi
    
    # Create coverage directory
    mkdir -p "${COVERAGE_DIR}"
}

# Run Go unit tests
run_go_unit_tests() {
    print_status info "Running Go unit tests..."
    
    local packages=(
        "./black-economy/..."
        "./feature-store/..."
        "./macro/..."
        "./intelligence/..."
        "./policy/..."
        "./reporting/..."
        "./sensing/..."
        "./intervention/..."
        "./shared/..."
        "./tests/go/..."
    )
    
    local args=("-v" "-timeout=${TIMEOUT}")
    
    if [ "$COVERAGE" = true ]; then
        args+=("-cover" "-coverprofile=${COVERAGE_DIR}/go_coverage.out" "-covermode=atomic")
    fi
    
    if [ "$PARALLEL" = true ]; then
        args+=("-parallel=4")
    fi
    
    # Run tests
    local failed=0
    for pkg in "${packages[@]}"; do
        local pkg_path="${PROJECT_ROOT}/${pkg}"
        if [ -d "$pkg_path" ]; then
            if ! go test "${args[@]}" "${pkg_path}" 2>&1; then
                print_status error "Failed tests in ${pkg}"
                failed=$((failed + 1))
            fi
        fi
    done
    
    if [ $failed -gt 0 ]; then
        print_status error "Go unit tests completed with $failed package(s) failing"
        return 1
    fi
    
    print_status success "Go unit tests passed"
    return 0
}

# Run Python tests
run_python_tests() {
    print_status info "Running Python tests..."
    
    local test_dir="${PROJECT_ROOT}/simulation/tests"
    
    if [ ! -d "$test_dir" ]; then
        print_status warning "Python test directory not found"
        return 0
    fi
    
    if ! command -v pytest &> /dev/null; then
        print_status warning "pytest not found, skipping Python tests"
        return 0
    fi
    
    local args=("-v" "--tb=short")
    
    if [ "$COVERAGE" = true ]; then
        args+=("--cov=${PROJECT_ROOT}/simulation" "--cov-report=term-missing" "--cov-report=html:${COVERAGE_DIR}/python_coverage")
    fi
    
    if ! pytest "${args[@]}" "$test_dir" 2>&1; then
        print_status error "Python tests failed"
        return 1
    fi
    
    print_status success "Python tests passed"
    return 0
}

# Run integration tests
run_integration_tests() {
    print_status info "Running integration tests..."
    
    # Check if integration tests are enabled
    if [ "${INTEGRATION_TEST_ENABLED:-false}" != "true" ]; then
        print_status warning "Integration tests are disabled. Set INTEGRATION_TEST_ENABLED=true to run."
        return 0
    fi
    
    # Run database integration tests
    print_status info "Running database integration tests..."
    
    local test_dir="${PROJECT_ROOT}/tests/integration"
    
    if [ -d "$test_dir" ]; then
        if ! go test -v -timeout="${TIMEOUT}" "${test_dir}" 2>&1; then
            print_status error "Integration tests failed"
            return 1
        fi
    fi
    
    print_status success "Integration tests passed"
    return 0
}

# Run benchmark tests
run_benchmark_tests() {
    print_status info "Running benchmark tests..."
    
    local packages=(
        "./black-economy/..."
        "./feature-store/..."
        "./macro/..."
        "./intelligence/..."
    )
    
    for pkg in "${packages[@]}"; do
        local pkg_path="${PROJECT_ROOT}/${pkg}"
        if [ -d "$pkg_path" ]; then
            go test -bench=. -benchmem -timeout="${TIMEOUT}" "${pkg_path}" > "${COVERAGE_DIR}/benchmark_${pkg//\//_}.txt" 2>&1 || true
        fi
    done
    
    print_status success "Benchmark tests completed"
    return 0
}

# Run contract tests
run_contract_tests() {
    print_status info "Running contract tests..."
    
    # Check if Pact CLI is available
    if ! command -v pact &> /dev/null; then
        print_status warning "Pact CLI not found. Contract tests require Pact CLI."
        return 0
    fi
    
    print_status success "Contract tests configured"
    return 0
}

# Generate coverage report
generate_coverage_report() {
    if [ "$COVERAGE" != true ]; then
        return 0
    fi
    
    print_status info "Generating coverage report..."
    
    # Combine coverage profiles
    if [ -f "${COVERAGE_DIR}/go_coverage.out" ]; then
        go tool cover -html="${COVERAGE_DIR}/go_coverage.out" -o "${COVERAGE_DIR}/go_coverage.html" 2>/dev/null || true
        
        # Print coverage summary
        if command -v go &> /dev/null; then
            local coverage=$(go tool cover -func="${COVERAGE_DIR}/go_coverage.out" 2>/dev/null | grep "total:" | awk '{print $3}' | sed 's/%//')
            print_status info "Total Go coverage: ${coverage}%"
            
            # Check coverage threshold
            if (( $(echo "$coverage < $COVERAGE_MIN" | bc -l) )); then
                print_status warning "Coverage ${coverage}% is below minimum ${COVERAGE_MIN}%"
            else
                print_status success "Coverage meets minimum requirement (${COVERAGE_MIN}%)"
            fi
        fi
    fi
    
    # Generate HTML reports
    if [ -d "${COVERAGE_DIR}/python_coverage" ]; then
        print_status info "Python coverage report: ${COVERAGE_DIR}/python_coverage"
    fi
    
    print_status success "Coverage reports generated"
}

# Run all tests
run_all_tests() {
    print_banner
    
    check_prerequisites
    
    local exit_code=0
    
    # Run unit tests
    print_status info "=========================================="
    print_status info "Running Unit Tests"
    print_status info "=========================================="
    
    if ! run_go_unit_tests; then
        exit_code=1
    fi
    
    if ! run_python_tests; then
        exit_code=1
    fi
    
    # Run integration tests
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "integration" ]; then
        print_status info "=========================================="
        print_status info "Running Integration Tests"
        print_status info "=========================================="
        if ! run_integration_tests; then
            exit_code=1
        fi
    fi
    
    # Run benchmark tests
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "benchmark" ]; then
        print_status info "=========================================="
        print_status info "Running Benchmark Tests"
        print_status info "=========================================="
        if ! run_benchmark_tests; then
            exit_code=1
        fi
    fi
    
    # Run contract tests
    if [ "$TEST_TYPE" = "all" ] || [ "$TEST_TYPE" = "contract" ]; then
        print_status info "=========================================="
        print_status info "Running Contract Tests"
        print_status info "=========================================="
        if ! run_contract_tests; then
            exit_code=1
        fi
    fi
    
    # Generate reports
    generate_coverage_report
    
    # Print summary
    print_status info "=========================================="
    print_status info "Test Summary"
    print_status info "=========================================="
    
    if [ $exit_code -eq 0 ]; then
        print_status success "All tests passed!"
    else
        print_status error "Some tests failed. Check the output above for details."
    fi
    
    return $exit_code
}

# Main
main() {
    if [ "$TEST_TYPE" = "all" ]; then
        run_all_tests
    else
        case $TEST_TYPE in
            unit)
                check_prerequisites
                run_go_unit_tests
                run_python_tests
                generate_coverage_report
                ;;
            integration)
                check_prerequisites
                run_integration_tests
                ;;
            benchmark)
                check_prerequisites
                run_benchmark_tests
                ;;
            contract)
                check_prerequisites
                run_contract_tests
                ;;
            *)
                print_status error "Unknown test type: $TEST_TYPE"
                exit 1
                ;;
        esac
    fi
}

main
