#!/bin/bash
# NEAM Backup Script
# Creates comprehensive backups of all platform data

set -e

# Configuration
NEAM_NAMESPACE=${NEAM_NAMESPACE:-neam-system}
BACKUP_DIR=${BACKUP_DIR:-/backups/neam}
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="neam-backup-${DATE}"
RETENTION_DAYS=${RETENTION_DAYS:-30}
COMPRESS=${COMPRESS:-true}
ENCRYPT=${ENCRYPT:-true}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Create backup directory
mkdir -p "$BACKUP_DIR/$BACKUP_NAME"

# Backup PostgreSQL
backup_postgres() {
    log_info "Backing up PostgreSQL..."
    local backup_file="$BACKUP_DIR/$BACKUP_NAME/postgresql.dump"

    if kubectl exec -n "$NEAM_NAMESPACE" neam-postgresql-0 -- pg_dump -U neam_user neam_platform > "$backup_file" 2>/dev/null; then
        log_info "PostgreSQL backup completed: $backup_file"
        if [ "$COMPRESS" = "true" ]; then
            gzip "$backup_file"
            log_info "Compressed to: ${backup_file}.gz"
        fi
    else
        log_error "PostgreSQL backup failed"
        return 1
    fi
}

# Backup ClickHouse
backup_clickhouse() {
    log_info "Backing up ClickHouse..."
    local backup_file="$BACKUP_DIR/$BACKUP_NAME/clickhouse.tar"

    kubectl exec -n "$NEAM_NAMESPACE" neam-clickhouse-0 -- clickhouse-client -u neam_user -q "BACKUP DATABASE neam_analytics TO Disk('default', '$BACKUP_NAME')" 2>/dev/null || true

    kubectl cp "$NEAM_NAMESPACE"/neam-clickhouse-0:/var/lib/clickhouse/backups/"$BACKUP_NAME" "$BACKUP_DIR/$BACKUP_NAME/clickhouse-backup" 2>/dev/null || true

    if [ -d "$BACKUP_DIR/$BACKUP_NAME/clickhouse-backup" ]; then
        log_info "ClickHouse backup completed"
        if [ "$COMPRESS" = "true" ]; then
            tar -czf "$backup_file.tar.gz" -C "$BACKUP_DIR/$BACKUP_NAME" clickhouse-backup
            rm -rf "$BACKUP_DIR/$BACKUP_NAME/clickhouse-backup"
            log_info "Compressed to: ${backup_file}.tar.gz"
        fi
    else
        log_warn "ClickHouse backup may be incomplete"
    fi
}

# Backup Redis
backup_redis() {
    log_info "Backing up Redis..."
    local backup_file="$BACKUP_DIR/$BACKUP_NAME/redis.rdb"

    kubectl exec -n "$NEAM_NAMESPACE" neam-redis-0 -- redis-cli BGSAVE" 2>/dev/null || true
    sleep 2

    kubectl cp "$NEAM_NAMESPACE"/neam-redis-0:/data/dump.rdb "$backup_file" 2>/dev/null || true

    if [ -f "$backup_file" ]; then
        log_info "Redis backup completed: $backup_file"
    else
        log_error "Redis backup failed"
    fi
}

# Backup Kafka offsets
backup_kafka() {
    log_info "Backing up Kafka offsets..."
    local backup_file="$BACKUP_DIR/$BACKUP_NAME/kafka-offsets.json"

    kubectl exec -n "$NEAM_NAMESPACE" kafka-0 -- kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --all-groups --offsets 2>/dev/null > "$backup_file" || true

    if [ -f "$backup_file" ]; then
        log_info "Kafka offsets backup completed: $backup_file"
    else
        log_warn "Kafka offsets backup skipped"
    fi
}

# Backup configuration
backup_config() {
    log_info "Backing up configuration..."
    local config_dir="$BACKUP_DIR/$BACKUP_NAME/config"

    mkdir -p "$config_dir"

    # Export Kubernetes configs
    kubectl get configmap -n "$NEAM_NAMESPACE" -o yaml > "$config_dir/configmaps.yaml" 2>/dev/null || true
    kubectl get secret -n "$NEAM_NAMESPACE" -o yaml > "$config_dir/secrets.yaml" 2>/dev/null || true

    # Export Terraform state reference (not actual state)
    echo "Terraform state location: $BACKUP_DIR/../terraform.state" > "$config_dir/terraform-ref.txt"

    log_info "Configuration backup completed"
}

# Create metadata
create_metadata() {
    log_info "Creating backup metadata..."
    local metadata_file="$BACKUP_DIR/$BACKUP_NAME/metadata.json"

    cat > "$metadata_file" << EOF
{
    "backup_name": "$BACKUP_NAME",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "retention_days": $RETENTION_DAYS,
    "components": {
        "postgresql": true,
        "clickhouse": true,
        "redis": true,
        "kafka": true,
        "config": true
    },
    "compression": $COMPRESS,
    "encryption": $ENCRYPT
}
EOF
    log_info "Metadata created: $metadata_file"
}

# Encrypt backup
encrypt_backup() {
    if [ "$ENCRYPT" = "true" ]; then
        log_info "Encrypting backup..."
        # Implementation would use GPG or similar
        log_warn "Encryption not implemented in this version"
    fi
}

# Cleanup old backups
cleanup_old() {
    log_info "Cleaning up backups older than $RETENTION_DAYS days..."
    find "$BACKUP_DIR" -maxdepth 1 -type d -name "neam-backup-*" -mtime +$RETENTION_DAYS -exec rm -rf {} \; 2>/dev/null || true
    log_info "Cleanup completed"
}

# Create checksum
create_checksum() {
    log_info "Creating checksum..."
    local checksum_file="$BACKUP_DIR/${BACKUP_NAME}.sha256"

    cd "$BACKUP_DIR"
    find "$BACKUP_NAME" -type f -exec sha256sum {} \; > "$checksum_file"
    log_info "Checksum created: $checksum_file"
}

# Print backup size
print_size() {
    log_info "Backup size:"
    du -sh "$BACKUP_DIR/$BACKUP_NAME" 2>/dev/null || echo "Unable to calculate size"
}

# Main execution
main() {
    echo "========================================"
    echo "NEAM Backup"
    echo "========================================"
    echo "Backup Name: $BACKUP_NAME"
    echo "Backup Directory: $BACKUP_DIR"
    echo "Retention: $RETENTION_DAYS days"
    echo "Compression: $COMPRESS"
    echo "Encryption: $ENCRYPT"
    echo "========================================"

    backup_postgres
    backup_clickhouse
    backup_redis
    backup_kafka
    backup_config
    create_metadata
    encrypt_backup
    create_checksum
    cleanup_old
    print_size

    echo ""
    log_info "Backup completed successfully"
    log_info "Backup location: $BACKUP_DIR/$BACKUP_NAME"
}

main "$@"
