package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/config"
	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresAuditRepository handles audit log data access
type PostgresAuditRepository struct {
	db *sql.DB
}

// NewPostgresAuditRepository creates a new audit repository
func NewPostgresAuditRepository(cfg config.DatabaseConfig) (*PostgresAuditRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresAuditRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresAuditRepository) Close() error {
	return r.db.Close()
}

// Create creates a new audit log entry
func (r *PostgresAuditRepository) Create(ctx context.Context, log *models.WalletAuditLog) error {
	query := `
		INSERT INTO wallet_audit_logs (
			id, entity_type, entity_id, action, actor_id, actor_name, actor_type,
			old_value, new_value, ip_address, user_agent, request_id, success,
			error_message, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	log.ID = uuid.New()
	log.CreatedAt = time.Now()

	oldValueJSON, _ := json.Marshal(log.OldValue)
	newValueJSON, _ := json.Marshal(log.NewValue)

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.Action, log.ActorID, log.ActorName,
		log.ActorType, oldValueJSON, newValueJSON, log.IPAddress, log.UserAgent,
		log.RequestID, log.Success, log.ErrorMessage, log.CreatedAt,
	)

	return err
}

// GetByEntity retrieves audit logs for a specific entity
func (r *PostgresAuditRepository) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*models.WalletAuditLog, error) {
	query := `
		SELECT id, entity_type, entity_id, action, actor_id, actor_name, actor_type,
			old_value, new_value, ip_address, user_agent, request_id, success,
			error_message, created_at
		FROM wallet_audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.WalletAuditLog
	for rows.Next() {
		var log models.WalletAuditLog
		var oldValue, newValue []byte

		err := rows.Scan(
			&log.ID, &log.EntityType, &log.EntityID, &log.Action, &log.ActorID,
			&log.ActorName, &log.ActorType, &oldValue, &newValue, &log.IPAddress,
			&log.UserAgent, &log.RequestID, &log.Success, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(oldValue, &log.OldValue)
		json.Unmarshal(newValue, &log.NewValue)
		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// GetByActor retrieves audit logs for a specific actor
func (r *PostgresAuditRepository) GetByActor(ctx context.Context, actorID uuid.UUID, limit, offset int) ([]*models.WalletAuditLog, error) {
	query := `
		SELECT id, entity_type, entity_id, action, actor_id, actor_name, actor_type,
			old_value, new_value, ip_address, user_agent, request_id, success,
			error_message, created_at
		FROM wallet_audit_logs
		WHERE actor_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, actorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.WalletAuditLog
	for rows.Next() {
		var log models.WalletAuditLog
		var oldValue, newValue []byte

		err := rows.Scan(
			&log.ID, &log.EntityType, &log.EntityID, &log.Action, &log.ActorID,
			&log.ActorName, &log.ActorType, &oldValue, &newValue, &log.IPAddress,
			&log.UserAgent, &log.RequestID, &log.Success, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(oldValue, &log.OldValue)
		json.Unmarshal(newValue, &log.NewValue)
		logs = append(logs, &log)
	}

	return logs, rows.Err()
}
