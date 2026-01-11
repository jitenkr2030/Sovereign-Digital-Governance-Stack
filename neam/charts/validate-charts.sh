#!/bin/bash
#
# NEAM Platform Helm Chart Validation Script
# This script validates all Helm charts in the neam-platform repository
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHARTS_DIR="${SCRIPT_DIR}/.."
LIBRARY_CHARTS_DIR="${CHARTS_DIR}/library"
TEMP_DIR=$(mktemp -d)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
SKIPPED=0

print_header() {
    echo "=============================================="
    echo "$1"
    echo "=============================================="
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

print_failure() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((SKIPPED++))
}

print_info() {
    echo -e "  ℹ $1"
}

cleanup() {
    rm -rf "${TEMP_DIR}"
    echo ""
    echo "Cleanup complete."
}

trap cleanup EXIT

echo ""
print_header "NEAM Platform Helm Chart Validation"
echo ""
echo "Charts Directory: ${CHARTS_DIR}"
echo "Date: $(date)"
echo ""

# Check if Helm is available
check_helm() {
    print_header "Checking Helm Installation"
    
    if command -v helm &> /dev/null; then
        HELM_VERSION=$(helm version --short 2>/dev/null || helm version 2>/dev/null)
        print_success "Helm is installed: ${HELM_VERSION}"
    else
        print_failure "Helm is not installed"
        print_info "Install Helm from: https://helm.sh/docs/intro/install/"
        echo ""
        print_info "Skipping Helm-specific validations..."
        return 1
    fi
}

# Validate library chart
validate_library_chart() {
    print_header "Validating Library Chart"
    
    if [ ! -d "${LIBRARY_CHARTS_DIR}" ]; then
        print_failure "Library chart directory not found: ${LIBRARY_CHARTS_DIR}"
        return 1
    fi
    
    # Check Chart.yaml
    if [ -f "${LIBRARY_CHARTS_DIR}/Chart.yaml" ]; then
        print_success "Chart.yaml exists"
        
        # Validate Chart.yaml content
        if grep -q "type: library" "${LIBRARY_CHARTS_DIR}/Chart.yaml"; then
            print_success "Chart type is library"
        else
            print_failure "Chart type is not set to library"
        fi
    else
        print_failure "Chart.yaml not found"
    fi
    
    # Check values.yaml
    if [ -f "${LIBRARY_CHARTS_DIR}/values.yaml" ]; then
        print_success "values.yaml exists"
    else
        print_failure "values.yaml not found"
    fi
    
    # Check templates
    local templates=(
        "_helpers.tpl"
        "_serviceaccount.tpl"
        "_deployment.tpl"
        "_service.tpl"
        "_ingress.tpl"
        "_configmap.tpl"
        "_secret.tpl"
        "_pvc.tpl"
        "_pdb.tpl"
        "_hpa.tpl"
    )
    
    for template in "${templates[@]}"; do
        if [ -f "${LIBRARY_CHARTS_DIR}/templates/${template}" ]; then
            print_success "Template exists: ${template}"
        else
            print_failure "Template missing: ${template}"
        fi
    done
    
    # Check README
    if [ -f "${LIBRARY_CHARTS_DIR}/README.md" ]; then
        print_success "README.md exists"
    else
        print_failure "README.md not found"
    fi
}

# Validate a service chart
validate_service_chart() {
    local chart_name="$1"
    local chart_dir="${CHARTS_DIR}/${chart_name}"
    
    echo ""
    print_info "Validating ${chart_name} chart..."
    
    if [ ! -d "${chart_dir}" ]; then
        print_failure "Chart directory not found: ${chart_dir}"
        return 1
    fi
    
    # Check Chart.yaml
    if [ -f "${chart_dir}/Chart.yaml" ]; then
        print_success "Chart.yaml exists for ${chart_name}"
        
        # Validate library dependency
        if grep -q "dependencies:" "${chart_dir}/Chart.yaml"; then
            if grep -q "name: library" "${chart_dir}/Chart.yaml"; then
                print_success "${chart_name} depends on library chart"
            else
                print_failure "${chart_name} does not depend on library chart"
            fi
        else
            print_failure "${chart_name} is missing dependencies section"
        fi
    else
        print_failure "Chart.yaml not found for ${chart_name}"
    fi
    
    # Check values.yaml
    if [ -f "${chart_dir}/values.yaml" ]; then
        print_success "values.yaml exists for ${chart_name}"
    else
        print_failure "values.yaml not found for ${chart_name}"
    fi
    
    # Check templates
    local templates=(
        "deployment.yaml"
        "service.yaml"
    )
    
    for template in "${templates[@]}"; do
        if [ -f "${chart_dir}/templates/${template}" ]; then
            print_success "Template exists: ${template}"
            
            # Verify template uses library
            if grep -q "include.*library\." "${chart_dir}/templates/${template}"; then
                print_success "Template ${template} uses library"
            else
                print_warning "Template ${template} may not use library"
            fi
        else
            print_failure "Template missing: ${template}"
        fi
    done
    
    # Check optional templates
    if [ -f "${chart_dir}/templates/ingress.yaml" ]; then
        print_success "ingress.yaml exists"
    fi
    
    if [ -f "${chart_dir}/templates/hpa.yaml" ]; then
        print_success "hpa.yaml exists"
    fi
    
    if [ -f "${chart_dir}/templates/pdb.yaml" ]; then
        print_success "pdb.yaml exists"
    fi
    
    if [ -f "${chart_dir}/templates/pvc.yaml" ]; then
        print_success "pvc.yaml exists"
    fi
    
    if [ -f "${chart_dir}/templates/configmap.yaml" ]; then
        print_success "configmap.yaml exists"
    fi
    
    if [ -f "${chart_dir}/templates/serviceaccount.yaml" ]; then
        print_success "serviceaccount.yaml exists"
    fi
}

# Validate all YAML files are valid YAML
validate_yaml_syntax() {
    print_header "Validating YAML Syntax"
    
    local yaml_files=()
    while IFS= read -r -d '' file; do
        yaml_files+=("$file")
    done < <(find "${CHARTS_DIR}" -name "*.yaml" -print0 2>/dev/null)
    
    for yaml_file in "${yaml_files[@]}"; do
        # Basic syntax check using Python if available
        if command -v python3 &> /dev/null; then
            if python3 -c "import yaml; yaml.safe_load(open('${yaml_file}'))" 2>/dev/null; then
                print_success "Valid YAML: $(basename ${yaml_file})"
            else
                print_failure "Invalid YAML: $(basename ${yaml_file})"
            fi
        else
            # Fallback to basic check
            if head -1 "${yaml_file}" | grep -q "apiVersion\|kind"; then
                print_success "Appears valid: $(basename ${yaml_file})"
            else
                print_failure "Possible invalid YAML: $(basename ${yaml_file})"
            fi
        fi
    done
}

# Template rendering test (requires Helm)
test_template_rendering() {
    if ! command -v helm &> /dev/null; then
        print_warning "Helm not available, skipping template rendering tests"
        return 1
    fi
    
    print_header "Testing Template Rendering"
    
    # Update dependencies
    echo ""
    print_info "Updating Helm dependencies..."
    cd "${CHARTS_DIR}/library"
    helm dependency update 2>/dev/null || print_warning "Could not update library dependencies"
    
    # Test each service chart
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
        echo ""
        print_info "Testing ${service} chart rendering..."
        
        cd "${CHARTS_DIR}/${service}"
        
        if [ -f "values.yaml" ]; then
            # Dry-run template
            if helm template "${service}" . --dry-run --debug 2>&1 | grep -q "Error"; then
                print_failure "Template rendering failed for ${service}"
            else
                print_success "Template rendering successful for ${service}"
            fi
        else
            print_failure "values.yaml not found for ${service}"
        fi
    done
}

# Check for common issues
check_common_issues() {
    print_header "Checking for Common Issues"
    
    # Check for hardcoded values that should be templated
    local hardcoded_patterns=(
        "replicas: [0-9]"
        "image: .*:latest"
        "memory: [0-9]*Mi"
        "cpu: [0-9]*m"
    )
    
    for pattern in "${hardcoded_patterns[@]}"; do
        if grep -r "$pattern" "${CHARTS_DIR}/templates/" --include="*.tpl" --include="*.yaml" 2>/dev/null | grep -v "^#" > /dev/null; then
            print_warning "Possible hardcoded value found matching: ${pattern}"
        fi
    done
    
    # Check for missing security contexts
    if ! grep -r "containerSecurityContext\|podSecurityContext" "${CHARTS_DIR}/sensing/values.yaml" > /dev/null; then
        print_failure "Security context may be missing in sensing values"
    else
        print_success "Security context configured in sensing"
    fi
    
    # Check for missing health probes
    if ! grep -r "livenessProbe\|readinessProbe" "${CHARTS_DIR}/sensing/values.yaml" > /dev/null; then
        print_failure "Health probes may be missing in sensing values"
    else
        print_success "Health probes configured in sensing"
    fi
}

# Generate validation report
generate_report() {
    print_header "Validation Summary"
    
    echo ""
    echo "Total Passed: ${PASSED}"
    echo "Total Failed: ${FAILED}"
    echo "Total Skipped: ${SKIPPED}"
    echo ""
    
    if [ ${FAILED} -eq 0 ]; then
        echo -e "${GREEN}All validations passed!${NC}"
        return 0
    else
        echo -e "${RED}Some validations failed. Please review the output above.${NC}"
        return 1
    fi
}

# Main execution
main() {
    check_helm || true
    validate_library_chart
    validate_yaml_syntax
    
    # Validate each service chart
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
        validate_service_chart "${service}"
    done
    
    check_common_issues
    test_template_rendering || true
    generate_report
}

# Run main function
main "$@"
