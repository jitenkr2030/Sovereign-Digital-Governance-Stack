#!/bin/bash
# SPIRE Management Utility Script
# Provides common operations for SPIRE administration

set -euo pipefail

NAMESPACE="spire"
SERVER_POD=""
TRUST_DOMAIN="neam.internal"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get SPIRE server pod
get_server_pod() {
    SERVER_POD=$(kubectl get pods -n "${NAMESPACE}" -l app=spire-server -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    if [ -z "${SERVER_POD}" ]; then
        echo -e "${RED}Error: No SPIRE server pod found${NC}"
        exit 1
    fi
}

# Check if SPIRE server is ready
check_server_ready() {
    get_server_pod
    if ! kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        wget -q -O - http://localhost:8080/healthz &> /dev/null; then
        echo -e "${RED}Error: SPIRE server is not ready${NC}"
        exit 1
    fi
    echo -e "${GREEN}SPIRE server is ready${NC}"
}

# List all registered entries
list_entries() {
    check_server_ready
    get_server_pod
    echo -e "${BLUE}Listing SPIFFE entries...${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server entry show -output json | jq '.[] | {id: .id, spiffe_id: .spiffe_id, parent_id: .parent_id}'
}

# Create a new registration entry
create_entry() {
    if [ $# -lt 3 ]; then
        echo "Usage: $0 create-entry <spiffe_id> <namespace> <service_account> [ttl]"
        echo "Example: $0 create-entry spiffe://neam.internal/myapp api myapp 24h"
        exit 1
    fi

    local spiffe_id="$1"
    local namespace="$2"
    local service_account="$3"
    local ttl="${4:-24h}"

    check_server_ready
    get_server_pod

    echo -e "${BLUE}Creating entry for ${spiffe_id}...${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server entry create \
        -spiffeID "${spiffe_id}" \
        -parentID spiffe://${TRUST_DOMAIN}/spire-server \
        -selector k8s:ns:"${namespace}" \
        -selector k8s:sa:"${service_account}" \
        -ttl "${ttl}"

    echo -e "${GREEN}Entry created successfully${NC}"
}

# Delete a registration entry
delete_entry() {
    if [ $# -lt 1 ]; then
        echo "Usage: $0 delete-entry <entry_id>"
        echo "Use 'list-entries' to find entry IDs"
        exit 1
    fi

    local entry_id="$1"

    check_server_ready
    get_server_pod

    echo -e "${BLUE}Deleting entry ${entry_id}...${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server entry delete -entryID "${entry_id}"

    echo -e "${GREEN}Entry deleted successfully${NC}"
}

# View agent status
agents_status() {
    check_server_ready
    get_server_pod

    echo -e "${BLUE}SPIRE Agents Status:${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server agent list -output json | jq '.[] | {spiffe_id: .spiffe_id, attestation_type: .attestation_type, selectors: .selectors}'
}

# View bundle
view_bundle() {
    check_server_ready
    get_server_pod

    echo -e "${BLUE}SVID Bundle (first 100 chars):${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server fetch bundle -output json | jq -r '.bundle' | head -c 100
    echo ""
}

# View server health
health_check() {
    check_server_ready
    get_server_pod

    echo -e "${BLUE}SPIRE Server Health:${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        wget -q -O - http://localhost:8080/healthz
    echo ""

    echo -e "${BLUE}SPIRE Server Metrics:${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        wget -q -O - http://localhost:2112/metrics 2>/dev/null | head -20 || \
        echo "Metrics endpoint not available"
}

# Rotate entries for a namespace
rotate_entries() {
    if [ $# -lt 1 ]; then
        echo "Usage: $0 rotate-entries <namespace>"
        exit 1
    fi

    local namespace="$1"

    check_server_ready
    get_server_pod

    echo -e "${BLUE}Rotating entries in namespace ${namespace}...${NC}"

    # Get all entry IDs for the namespace
    local entries=$(kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server entry show -selector "k8s:ns:${namespace}" -output json 2>/dev/null | \
        jq -r '.[].id' || echo "")

    if [ -z "${entries}" ]; then
        echo -e "${YELLOW}No entries found for namespace ${namespace}${NC}"
        return
    fi

    for entry_id in ${entries}; do
        echo "Rotating entry ${entry_id}..."
        kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
            spire-server entry update -entryID "${entry_id}" -ttl 24h
    done

    echo -e "${GREEN}Entries rotated successfully${NC}"
}

# Export entries for backup
export_entries() {
    check_server_ready
    get_server_pod

    local backup_file="spire-entries-$(date +%Y%m%d-%H%M%S).json"

    echo -e "${BLUE}Exporting entries to ${backup_file}...${NC}"
    kubectl exec -n "${NAMESPACE}" "${SERVER_POD}" -- \
        spire-server entry show -output json > "/tmp/${backup_file}"

    kubectl cp "${NAMESPACE}/${SERVER_POD}:/tmp/${backup_file}" "${backup_file}" 2>/dev/null || \
        echo "Saved to /tmp/${backup_file} in spire-server pod"

    echo -e "${GREEN}Entries exported to ${backup_file}${NC}"
}

# Show help
show_help() {
    echo "SPIRE Management Utility"
    echo "========================"
    echo ""
    echo "Usage: $0 <command> [arguments]"
    echo ""
    echo "Commands:"
    echo "  list-entries           - List all registered SPIFFE entries"
    echo "  create-entry <args>    - Create a new registration entry"
    echo "  delete-entry <id>      - Delete a registration entry"
    echo "  agents-status          - Show SPIRE agent status"
    echo "  view-bundle            - View SVID bundle"
    echo "  health-check           - Show server health and metrics"
    echo "  rotate-entries <ns>    - Rotate all entries in a namespace"
    echo "  export-entries         - Export entries for backup"
    echo ""
    echo "Examples:"
    echo "  $0 create-entry spiffe://neam.internal/myapp api myapp 24h"
    echo "  $0 rotate-entries finance"
    echo "  $0 list-entries"
}

# Main
case "${1:-help}" in
    list-entries)
        list_entries
        ;;
    create-entry)
        shift
        create_entry "$@"
        ;;
    delete-entry)
        shift
        delete_entry "$@"
        ;;
    agents-status)
        agents_status
        ;;
    view-bundle)
        view_bundle
        ;;
    health-check)
        health_check
        ;;
    rotate-entries)
        shift
        rotate_entries "$@"
        ;;
    export-entries)
        export_entries
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        show_help
        exit 1
        ;;
esac
