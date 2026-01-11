package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"csic-platform/service/security/internal/domain/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgreSQL audit repository with tamper-proof hash chain implementation.
type PostgresAuditRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresAuditRepository creates a new PostgreSQL audit repository.
func NewPostgresAuditRepository(db *sql.DB, tablePrefix string) *PostgresAuditRepository {
	return &PostgresAuditRepository{
		db:          db,
		tablePrefix: tablePrefix,
	}
}

// tableName returns the full table name with prefix.
func (r *PostgresAuditRepository) tableName() string {
	return r.tablePrefix + "_audit_logs"
}

// chainMetadataTable returns the chain metadata table name.
func (r *PostgresAuditRepository) chainMetadataTable() string {
	return r.tablePrefix + "_chain_metadata"
}

// Create inserts a new audit log entry with hash chain linkage.
func (r *PostgresAuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	log.Timestamp = time.Now().UTC()
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now().UTC()
	}

	// Get the current chain head to establish hash chain linkage.
	head, err := r.GetChainHead(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to get chain head: %w", err)
	}

	// Compute the hash chain values.
	// PreviousHash links to the previous entry, CurrentHash includes all entry data.
	prevHash := ""
	if head != nil {
		prevHash = head.CurrentHash
	}

	log.PrevHash = prevHash

	// Serialize the log entry for hashing (excluding the computed hashes).
	logData := r.serializeForHash(log)
	hash := sha256.Sum256(logData)
	log.CurrentHash = hex.EncodeToString(hash[:])

	// Insert the audit log entry.
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, timestamp, event_type, severity, entity_type, entity_id,
			user_id, user_email, action, resource_type, resource_id,
			old_value, new_value, metadata, ip_address, user_agent,
			correlation_id, session_id, prev_hash, current_hash,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`, r.tableName())

	now := time.Now().UTC()
	_, err = r.db.ExecContext(ctx, query,
		log.ID,
		log.Timestamp,
		log.EventType,
		log.Severity,
		log.EntityType,
		log.EntityID,
		log.UserID,
		log.UserEmail,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.OldValue,
		log.NewValue,
		log.Metadata,
		log.IPAddress,
		log.UserAgent,
		log.CorrelationID,
		log.SessionID,
		log.PrevHash,
		log.CurrentHash,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	// Update chain metadata.
	if err := r.updateChainMetadata(ctx, log.ID, log.CurrentHash); err != nil {
		// Log the error but don't fail the operation.
		fmt.Printf("Warning: failed to update chain metadata: %v\n", err)
	}

	return nil
}

// GetByID retrieves an audit log entry by its unique identifier.
func (r *PostgresAuditRepository) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	query := fmt.Sprintf(`
		SELECT id, timestamp, event_type, severity, entity_type, entity_id,
			   user_id, user_email, action, resource_type, resource_id,
			   old_value, new_value, metadata, ip_address, user_agent,
			   correlation_id, session_id, prev_hash, current_hash, created_at
		FROM %s WHERE id = $1
	`, r.tableName())

	log := &models.AuditLog{}
	var metadata, oldValue, newValue sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.Timestamp,
		&log.EventType,
		&log.Severity,
		&log.EntityType,
		&log.EntityID,
		&log.UserID,
		&log.UserEmail,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&oldValue,
		&newValue,
		&metadata,
		&log.IPAddress,
		&log.UserAgent,
		&log.CorrelationID,
		&log.SessionID,
		&log.PrevHash,
		&log.CurrentHash,
		&log.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve audit log: %w", err)
	}

	log.OldValue = oldValue.String
	log.NewValue = newValue.String
	log.Metadata = metadata.String

	return log, nil
}

// GetByTimestampRange retrieves audit logs within a specified time range.
func (r *PostgresAuditRepository) GetByTimestampRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT id, timestamp, event_type, severity, entity_type, entity_id,
			   user_id, user_email, action, resource_type, resource_id,
			   old_value, new_value, metadata, ip_address, user_agent,
			   correlation_id, session_id, prev_hash, current_hash, created_at
		FROM %s
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`, r.tableName())

	rows, err := r.db.QueryContext(ctx, query, start, end, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetByEntity retrieves audit logs for a specific entity.
func (r *PostgresAuditRepository) GetByEntity(ctx context.Context, entityID string, limit, offset int) ([]*models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT id, timestamp, event_type, severity, entity_type, entity_id,
			   user_id, user_email, action, resource_type, resource_id,
			   old_value, new_value, metadata, ip_address, user_agent,
			   correlation_id, session_id, prev_hash, current_hash, created_at
		FROM %s
		WHERE entity_type || ':' || entity_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`, r.tableName())

	rows, err := r.db.QueryContext(ctx, query, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by entity: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// VerifyChain performs cryptographic verification of the hash chain.
func (r *PostgresAuditRepository) VerifyChain(ctx context.Context, startID, endID string) (bool, error) {
	// Get the start and end entries.
	startLog, err := r.GetByID(ctx, startID)
	if err != nil {
		return false, fmt.Errorf("failed to get start entry: %w", err)
	}
	if startLog == nil {
		return false, fmt.Errorf("start entry not found: %s", startID)
	}

	endLog, err := r.GetByID(ctx, endID)
	if err != nil {
		return false, fmt.Errorf("failed to get end entry: %w", err)
	}
	if endLog == nil {
		return false, fmt.Errorf("end entry not found: %s", endID)
	}

	// Verify the chain forward from start to end.
	current := startLog
	for current.ID != endID {
		// Get the next entry (the one that has this entry as its PreviousHash).
		next, err := r.getNextEntry(ctx, current.CurrentHash)
		if err != nil {
			return false, fmt.Errorf("failed to get next entry from chain: %w", err)
		}
		if next == nil {
			return false, fmt.Errorf("chain broken at entry: %s", current.ID)
		}

		// Verify that the next entry's PreviousHash matches the current entry's CurrentHash.
		if next.PrevHash != current.CurrentHash {
			return false, fmt.Errorf("chain tampered: hash mismatch at %s", next.ID)
		}

		// Recompute and verify the next entry's CurrentHash.
		logData := r.serializeForHash(next)
		hash := sha256.Sum256(logData)
		expectedHash := hex.EncodeToString(hash[:])
		if next.CurrentHash != expectedHash {
			return false, fmt.Errorf("chain tampered: current hash mismatch at %s", next.ID)
		}

		current = next
	}

	return true, nil
}

// GetChainHead retrieves the most recent audit log entry.
func (r *PostgresAuditRepository) GetChainHead(ctx context.Context) (*models.AuditLog, error) {
	query := fmt.Sprintf(`
		SELECT id, timestamp, event_type, severity, entity_type, entity_id,
			   user_id, user_email, action, resource_type, resource_id,
			   old_value, new_value, metadata, ip_address, user_agent,
			   correlation_id, session_id, prev_hash, current_hash, created_at
		FROM %s
		ORDER BY created_at DESC
		LIMIT 1
	`, r.tableName())

	log := &models.AuditLog{}
	var metadata, oldValue, newValue sql.NullString

	err := r.db.QueryRowContext(ctx, query).Scan(
		&log.ID,
		&log.Timestamp,
		&log.EventType,
		&log.Severity,
		&log.EntityType,
		&log.EntityID,
		&log.UserID,
		&log.UserEmail,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&oldValue,
		&newValue,
		&metadata,
		&log.IPAddress,
		&log.UserAgent,
		&log.CorrelationID,
		&log.SessionID,
		&log.PrevHash,
		&log.CurrentHash,
		&log.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get chain head: %w", err)
	}

	log.OldValue = oldValue.String
	log.NewValue = newValue.String
	log.Metadata = metadata.String

	return log, nil
}

// Count returns the total number of audit log entries.
func (r *PostgresAuditRepository) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName())
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}
	return count, nil
}

// Search performs a full-text search on audit log metadata and details.
func (r *PostgresAuditRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	searchQuery := fmt.Sprintf(`
		SELECT id, timestamp, event_type, severity, entity_type, entity_id,
			   user_id, user_email, action, resource_type, resource_id,
			   old_value, new_value, metadata, ip_address, user_agent,
			   correlation_id, session_id, prev_hash, current_hash, created_at
		FROM %s
		WHERE (
			TO_TSVECTOR('english', COALESCE(metadata, '') || ' ' || COALESCE(action, '') || ' ' || COALESCE(entity_id, ''))
			@@ TO_TSQUERY('english', $1)
			OR ILIKE($2, '%%' || entity_id || '%%')
			OR ILIKE($2, '%%' || user_id || '%%')
			OR ILIKE($2, '%%' || action || '%%')
		)
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`, r.tableName())

	// Convert search query to PostgreSQL full-text search format.
	tsQuery := strings.Join(strings.Fields(query), " & ")

	rows, err := r.db.QueryContext(ctx, searchQuery, tsQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search audit logs: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// getNextEntry retrieves the next entry in the hash chain.
func (r *PostgresAuditRepository) getNextEntry(ctx context.Context, prevHash string) (*models.AuditLog, error) {
	query := fmt.Sprintf(`
		SELECT id, timestamp, event_type, severity, entity_type, entity_id,
			   user_id, user_email, action, resource_type, resource_id,
			   old_value, new_value, metadata, ip_address, user_agent,
			   correlation_id, session_id, prev_hash, current_hash, created_at
		FROM %s
		WHERE prev_hash = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, r.tableName())

	log := &models.AuditLog{}
	var metadata, oldValue, newValue sql.NullString

	err := r.db.QueryRowContext(ctx, query, prevHash).Scan(
		&log.ID,
		&log.Timestamp,
		&log.EventType,
		&log.Severity,
		&log.EntityType,
		&log.EntityID,
		&log.UserID,
		&log.UserEmail,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&oldValue,
		&newValue,
		&metadata,
		&log.IPAddress,
		&log.UserAgent,
		&log.CorrelationID,
		&log.SessionID,
		&log.PrevHash,
		&log.CurrentHash,
		&log.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get next entry: %w", err)
	}

	log.OldValue = oldValue.String
	log.NewValue = newValue.String
	log.Metadata = metadata.String

	return log, nil
}

// updateChainMetadata updates the chain metadata table.
func (r *PostgresAuditRepository) updateChainMetadata(ctx context.Context, entryID, currentHash string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (last_entry_id, last_hash, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			last_entry_id = EXCLUDED.last_entry_id,
			last_hash = EXCLUDED.last_hash,
			updated_at = EXCLUDED.updated_at
	`, r.chainMetadataTable())

	_, err := r.db.ExecContext(ctx, query, entryID, currentHash, time.Now().UTC())
	return err
}

// serializeForHash creates a deterministic byte representation of the log for hashing.
// This excludes the computed hash fields to ensure consistency.
func (r *PostgresAuditRepository) serializeForHash(log *models.AuditLog) []byte {
	// Create a struct for hashing that excludes the hash fields.
	type hashableLog struct {
		ID           string    `json:"id"`
		Timestamp    time.Time `json:"timestamp"`
		EventType    string    `json:"event_type"`
		Severity     string    `json:"severity"`
		EntityType   string    `json:"entity_type"`
		EntityID     string    `json:"entity_id"`
		UserID       string    `json:"user_id"`
		UserEmail    string    `json:"user_email"`
		Action       string    `json:"action"`
		ResourceType string    `json:"resource_type"`
		ResourceID   string    `json:"resource_id"`
		OldValue     string    `json:"old_value"`
		NewValue     string    `json:"new_value"`
		Metadata     string    `json:"metadata"`
		IPAddress    string    `json:"ip_address"`
		UserAgent    string    `json:"user_agent"`
		CorrelationID string   `json:"correlation_id"`
		SessionID    string    `json:"session_id"`
		PrevHash     string    `json:"prev_hash"`
	}

	hashable := hashableLog{
		ID:           log.ID,
		Timestamp:    log.Timestamp,
		EventType:    log.EventType,
		Severity:     log.Severity,
		EntityType:   log.EntityType,
		EntityID:     log.EntityID,
		UserID:       log.UserID,
		UserEmail:    log.UserEmail,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID,
		OldValue:     log.OldValue,
		NewValue:     log.NewValue,
		Metadata:     log.Metadata,
		IPAddress:    log.IPAddress,
		UserAgent:    log.UserAgent,
		CorrelationID: log.CorrelationID,
		SessionID:    log.SessionID,
		PrevHash:     log.PrevHash,
	}

	data, _ := json.Marshal(hashable)
	return data
}

// scanLogs scans database rows into AuditLog structs.
func (r *PostgresAuditRepository) scanLogs(rows *sql.Rows) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog

	for rows.Next() {
		log := &models.AuditLog{}
		var metadata, oldValue, newValue sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.EventType,
			&log.Severity,
			&log.EntityType,
			&log.EntityID,
			&log.UserID,
			&log.UserEmail,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&oldValue,
			&newValue,
			&metadata,
			&log.IPAddress,
			&log.UserAgent,
			&log.CorrelationID,
			&log.SessionID,
			&log.PrevHash,
			&log.CurrentHash,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		log.OldValue = oldValue.String
		log.NewValue = newValue.String
		log.Metadata = metadata.String
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
