#!/bin/bash
#
# CSIC Platform - Full Database Backup Script
# Version: 1.0.0
# Purpose: Create complete backups of all CSIC databases
#

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/../config/backup.conf"

# Source configuration
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "ERROR: Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Logging
LOG_FILE="${LOG_DIR}/backup-$(date +%Y%m%d-%H%M%S).log"
BACKUP_STATUS=0

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

log_error() {
    log "ERROR" "$@"
    BACKUP_STATUS=1
}

log_info() {
    log "INFO" "$@"
}

log_warn() {
    log "WARN" "$@"
}

# Signal handlers
cleanup() {
    if [[ -n "${WORK_DIR:-}" && -d "$WORK_DIR" ]]; then
        rm -rf "$WORK_DIR"
    fi
}

trap cleanup EXIT

# Create working directory
WORK_DIR=$(mktemp -d)
BACKUP_START_TIME=$(date +%s)

log_info "Starting backup process..."

# Create backup directories
create_backup_dirs() {
    log_info "Creating backup directories..."
    mkdir -p "$BACKUP_DIR/postgresql"
    mkdir -p "$BACKUP_DIR/timescaledb"
    mkdir -p "$BACKUP_DIR/redis"
    mkdir -p "$BACKUP_DIR/worm"
    mkdir -p "$BACKUP_DIR/config"
    chmod 700 "$BACKUP_DIR"
}

# Backup PostgreSQL
backup_postgresql() {
    log_info "Starting PostgreSQL backup..."
    local backup_file="${BACKUP_DIR}/postgresql/postgresql-$(date +%Y%m%d-%H%M%S).sql.gz.enc"
    local start_time=$(date +%s)
    
    # Create backup with compression and encryption
    PGPASSWORD="$DB_PASSWORD" pg_dump \
        -h "$DB_HOST" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --format=custom \
        --compress=9 \
        --verbose \
        2>> "$LOG_FILE" | \
        openssl enc -aes-256-cbc \
        -salt \
        -pbkdf2 \
        -pass pass:"$ENCRYPTION_KEY" \
        > "$backup_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $? -eq 0 ]]; then
        log_info "PostgreSQL backup completed: $backup_file (${duration}s)"
        # Calculate checksum
        sha256sum "$backup_file" > "${backup_file}.sha256"
        # Update manifest
        echo "postgresql:$(date +%Y-%m-%d-%H:%M):${backup_file}:${duration}:$(filesize "$backup_file")" >> "$MANIFEST_FILE"
    else
        log_error "PostgreSQL backup failed"
        return 1
    fi
}

# Backup TimescaleDB
backup_timescaledb() {
    log_info "Starting TimescaleDB backup..."
    local backup_file="${BACKUP_DIR}/timescaledb/timescaledb-$(date +%Y%m%d-%H%M%S).sql.gz.enc"
    local start_time=$(date +%s)
    
    # Backup time-series data with retention
    PGPASSWORD="$TSDB_PASSWORD" pg_dump \
        -h "$TSDB_HOST" \
        -U "$TSDB_USER" \
        -d "$TSDB_NAME" \
        --format=custom \
        --compress=9 \
        --verbose \
        -t 'telemetry.*' \
        -t 'mining_metrics' \
        -t 'energy_data' \
        2>> "$LOG_FILE" | \
        openssl enc -aes-256-cbc \
        -salt \
        -pbkdf2 \
        -pass pass:"$ENCRYPTION_KEY" \
        > "$backup_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $? -eq 0 ]]; then
        log_info "TimescaleDB backup completed: $backup_file (${duration}s)"
        sha256sum "$backup_file" > "${backup_file}.sha256"
    else
        log_error "TimescaleDB backup failed"
        return 1
    fi
}

# Backup Redis
backup_redis() {
    log_info "Starting Redis backup..."
    local backup_file="${BACKUP_DIR}/redis/redis-$(date +%Y%m%d-%H%M%S).rdb.gz.enc"
    local start_time=$(date +%s)
    
    # Trigger BGSAVE and wait
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" BGSAVE 2>> "$LOG_FILE"
    
    # Wait for backup to complete
    while [[ $(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" LASTSAVE) == $(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" LASTSAVE) ]]; do
        sleep 1
    done
    
    # Copy and encrypt RDB file
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" BGREWRITEAOF 2>> "$LOG_FILE"
    
    # Copy the RDB file
    local redis_host="${REDIS_HOST%%:*}"
    scp "$REDIS_HOST:$REDIS_PORT/dump.rdb" - 2>/dev/null | \
        gzip -9 | \
        openssl enc -aes-256-cbc \
        -salt \
        -pbkdf2 \
        -pass pass:"$ENCRYPTION_KEY" \
        > "$backup_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $? -eq 0 ]]; then
        log_info "Redis backup completed: $backup_file (${duration}s)"
        sha256sum "$backup_file" > "${backup_file}.sha256"
    else
        log_error "Redis backup failed"
        return 1
    fi
}

# Backup WORM storage
backup_worm() {
    log_info "Starting WORM storage backup..."
    local backup_file="${BACKUP_DIR}/worm/worm-$(date +%Y%m%d-%H%M%S).tar.gz.enc"
    local start_time=$(date +%s)
    
    # Create tar archive of WORM storage
    tar -czf - -C "$(dirname "$WORM_STORAGE_PATH")" "$(basename "$WORM_STORAGE_PATH")" 2>> "$LOG_FILE" | \
        openssl enc -aes-256-cbc \
        -salt \
        -pbkdf2 \
        -pass pass:"$ENCRYPTION_KEY" \
        > "$backup_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $? -eq 0 ]]; then
        log_info "WORM storage backup completed: $backup_file (${duration}s)"
        sha256sum "$backup_file" > "${backup_file}.sha256"
    else
        log_error "WORM storage backup failed"
        return 1
    fi
}

# Backup configuration files
backup_config() {
    log_info "Starting configuration backup..."
    local backup_file="${BACKUP_DIR}/config/config-$(date +%Y%m%d-%H%M%S).tar.gz.enc"
    local start_time=$(date +%s)
    
    # Create tar archive of configuration
    tar -czf - \
        -C /etc/csic \
        . 2>> "$LOG_FILE" | \
        openssl enc -aes-256-cbc \
        -salt \
        -pbkdf2 \
        -pass pass:"$ENCRYPTION_KEY" \
        > "$backup_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [[ $? -eq 0 ]]; then
        log_info "Configuration backup completed: $backup_file (${duration}s)"
        sha256sum "$backup_file" > "${backup_file}.sha256"
    else
        log_error "Configuration backup failed"
        return 1
    fi
}

# Create manifest
create_manifest() {
    log_info "Creating backup manifest..."
    local manifest_file="${BACKUP_DIR}/manifest-$(date +%Y%m%d).txt"
    
    cat > "$manifest_file" << EOF
CSIC Platform Backup Manifest
=============================
Backup Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)
Backup Version: $BACKUP_VERSION
Generated By: backup-full.sh

Backup Components:
------------------
EOF
    
    # Add backup metadata
    echo "" >> "$manifest_file"
    log_info "Manifest created: $manifest_file"
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups..."
    
    # Find and remove backups older than RETENTION_DAYS
    find "$BACKUP_DIR" -type f -name "*.enc" -mtime +$((RETENTION_DAYS)) -delete 2>/dev/null || true
    
    # Keep only last 30 manifests
    find "$BACKUP_DIR" -type f -name "manifest-*.txt" -mtime +30 -delete 2>/dev/null || true
    
    log_info "Old backup cleanup completed"
}

# Verify backup integrity
verify_backup() {
    log_info "Verifying backup integrity..."
    local verify_failed=0
    
    # Check all SHA256 checksums
    while IFS= read -r backup_file; do
        if ! sha256sum -c "${backup_file}.sha256" > /dev/null 2>&1; then
            log_error "Checksum verification failed: $backup_file"
            verify_failed=1
        fi
    done < <(find "$BACKUP_DIR" -type f -name "*.enc")
    
    if [[ $verify_failed -eq 0 ]]; then
        log_info "All backup checksums verified successfully"
    else
        log_error "Backup verification failed"
        return 1
    fi
}

# Create replication to remote storage
replicate_to_remote() {
    log_info "Replicating backups to remote storage..."
    
    if [[ -n "$REMOTE_BACKUP_HOST" ]]; then
        # Create incremental backup directory on remote
        ssh "$REMOTE_BACKUP_HOST" "mkdir -p $REMOTE_BACKUP_PATH" 2>/dev/null || true
        
        # Sync backup files
        rsync -avz --progress \
            -e "ssh -i $REMOTE_SSH_KEY" \
            "$BACKUP_DIR/" \
            "$REMOTE_BACKUP_HOST:$REMOTE_BACKUP_PATH/" 2>> "$LOG_FILE"
        
        if [[ $? -eq 0 ]]; then
            log_info "Backup replication completed"
        else
            log_warn "Backup replication encountered errors"
        fi
    else
        log_info "Remote backup not configured, skipping replication"
    fi
}

# Send notification
send_notification() {
    local status="$1"
    local message="$2"
    
    if [[ "$ENABLE_NOTIFICATIONS" == "true" ]]; then
        # Send email notification
        if [[ -n "$NOTIFICATION_EMAIL" ]]; then
            echo "$message" | mail -s "[CSIC Backup] $status: Backup Report" "$NOTIFICATION_EMAIL" 2>/dev/null || true
        fi
        
        # Send Slack notification
        if [[ -n "$SLACK_WEBHOOK" ]]; then
            curl -s -X POST \
                -H 'Content-type: application/json' \
                --data "{\"text\":\"[CSIC Backup] $status: $message\"}" \
                "$SLACK_WEBHOOK" 2>/dev/null || true
        fi
    fi
}

# Main function
main() {
    log_info "========================================="
    log_info "CSIC Platform Full Backup Starting"
    log_info "========================================="
    
    # Initialize
    create_backup_dirs
    MANIFEST_FILE="${BACKUP_DIR}/manifest-$(date +%Y%m%d).txt"
    touch "$MANIFEST_FILE"
    
    # Execute backups
    local backup_start=$(date +%s)
    local failed_backups=0
    
    if ! backup_postgresql; then
        failed_backups=$((failed_backups + 1))
    fi
    
    if ! backup_timescaledb; then
        failed_backups=$((failed_backups + 1))
    fi
    
    if ! backup_redis; then
        failed_backups=$((failed_backups + 1))
    fi
    
    if ! backup_worm; then
        failed_backups=$((failed_backups + 1))
    fi
    
    if ! backup_config; then
        failed_backups=$((failed_backups + 1))
    fi
    
    local backup_end=$(date +%s)
    local backup_duration=$((backup_end - backup_start))
    
    # Post-backup tasks
    create_manifest
    cleanup_old_backups
    verify_backup
    replicate_to_remote
    
    # Finalize
    log_info "========================================="
    log_info "Backup Process Completed"
    log_info "Duration: ${backup_duration} seconds"
    log_info "Failed Backups: $failed_backups"
    log_info "========================================="
    
    # Send notification
    if [[ $failed_backups -eq 0 ]]; then
        send_notification "SUCCESS" "Full backup completed in ${backup_duration}s"
        exit 0
    else
        send_notification "FAILED" "$failed_backups backup(s) failed"
        exit 1
    fi
}

# Run main function
main "$@"
