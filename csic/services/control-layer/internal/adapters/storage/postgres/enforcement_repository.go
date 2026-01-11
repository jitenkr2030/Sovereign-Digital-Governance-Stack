package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
)

// PostgresEnforcementRepository implements EnforcementRepository using PostgreSQL
type PostgresEnforcementRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresEnforcementRepository creates a new PostgreSQL enforcement repository
func NewPostgresEnforcementRepository(databaseURL string) (ports.EnforcementRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresEnforcementRepository{
		db:          db,
		tablePrefix: "control_layer_",
	}, nil
}

// Close closes the database connection
func (r *PostgresEnforcementRepository) Close() error {
	return r.db.Close()
}

// tableName returns the prefixed table name
func (r *PostgresEnforcementRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// CreateEnforcement creates a new enforcement record
func (r *PostgresEnforcementRepository) CreateEnforcement(ctx context.Context, enforcement *domain.Enforcement) error {
	metadataJSON, err := json.Marshal(enforcement.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal enforcement metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, policy_id, target_service, action_type, severity, status, 
		                message, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.tableName("enforcements"))

	_, err = r.db.ExecContext(ctx, query,
		enforcement.ID,
		enforcement.PolicyID,
		enforcement.TargetService,
		enforcement.ActionType,
		enforcement.Severity,
		enforcement.Status,
		enforcement.Message,
		metadataJSON,
		enforcement.CreatedAt,
		enforcement.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create enforcement: %w", err)
	}

	return nil
}

// GetEnforcementByID retrieves an enforcement by ID
func (r *PostgresEnforcementRepository) GetEnforcementByID(ctx context.Context, id uuid.UUID) (*domain.Enforcement, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, target_service, action_type, severity, status, 
		       message, metadata, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.tableName("enforcements"))

	var enforcement domain.Enforcement
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&enforcement.ID,
		&enforcement.PolicyID,
		&enforcement.TargetService,
		&enforcement.ActionType,
		&enforcement.Severity,
		&enforcement.Status,
		&enforcement.Message,
		&metadataJSON,
		&enforcement.CreatedAt,
		&enforcement.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get enforcement: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &enforcement.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &enforcement, nil
}

// GetEnforcementsByPolicy retrieves enforcements for a specific policy
func (r *PostgresEnforcementRepository) GetEnforcementsByPolicy(ctx context.Context, policyID uuid.UUID, limit int) ([]*domain.Enforcement, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, target_service, action_type, severity, status, 
		       message, metadata, created_at, updated_at
		FROM %s
		WHERE policy_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, r.tableName("enforcements"))

	rows, err := r.db.QueryContext(ctx, query, policyID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query enforcements: %w", err)
	}
	defer rows.Close()

	var enforcements []*domain.Enforcement

	for rows.Next() {
		var enforcement domain.Enforcement
		var metadataJSON []byte

		if err := rows.Scan(
			&enforcement.ID,
			&enforcement.PolicyID,
			&enforcement.TargetService,
			&enforcement.ActionType,
			&enforcement.Severity,
			&enforcement.Status,
			&enforcement.Message,
			&metadataJSON,
			&enforcement.CreatedAt,
			&enforcement.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan enforcement: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &enforcement.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		enforcements = append(enforcements, &enforcement)
	}

	return enforcements, nil
}

// GetEnforcementsByTarget retrieves enforcements targeting a specific service
func (r *PostgresEnforcementRepository) GetEnforcementsByTarget(ctx context.Context, target string, limit int) ([]*domain.Enforcement, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, target_service, action_type, severity, status, 
		       message, metadata, created_at, updated_at
		FROM %s
		WHERE target_service = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, r.tableName("enforcements"))

	rows, err := r.db.QueryContext(ctx, query, target, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query enforcements: %w", err)
	}
	defer rows.Close()

	var enforcements []*domain.Enforcement

	for rows.Next() {
		var enforcement domain.Enforcement
		var metadataJSON []byte

		if err := rows.Scan(
			&enforcement.ID,
			&enforcement.PolicyID,
			&enforcement.TargetService,
			&enforcement.ActionType,
			&enforcement.Severity,
			&enforcement.Status,
			&enforcement.Message,
			&metadataJSON,
			&enforcement.CreatedAt,
			&enforcement.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan enforcement: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &enforcement.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		enforcements = append(enforcements, &enforcement)
	}

	return enforcements, nil
}

// GetRecentEnforcements retrieves recent enforcements
func (r *PostgresEnforcementRepository) GetRecentEnforcements(ctx context.Context, since time.Time) ([]*domain.Enforcement, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, target_service, action_type, severity, status, 
		       message, metadata, created_at, updated_at
		FROM %s
		WHERE created_at >= $1
		ORDER BY created_at DESC
	`, r.tableName("enforcements"))

	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query enforcements: %w", err)
	}
	defer rows.Close()

	var enforcements []*domain.Enforcement

	for rows.Next() {
		var enforcement domain.Enforcement
		var metadataJSON []byte

		if err := rows.Scan(
			&enforcement.ID,
			&enforcement.PolicyID,
			&enforcement.TargetService,
			&enforcement.ActionType,
			&enforcement.Severity,
			&enforcement.Status,
			&enforcement.Message,
			&metadataJSON,
			&enforcement.CreatedAt,
			&enforcement.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan enforcement: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &enforcement.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		enforcements = append(enforcements, &enforcement)
	}

	return enforcements, nil
}

// UpdateEnforcementStatus updates the status of an enforcement
func (r *PostgresEnforcementRepository) UpdateEnforcementStatus(ctx context.Context, id uuid.UUID, status domain.EnforcementStatus) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, r.tableName("enforcements"))

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update enforcement status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("enforcement not found: %s", id)
	}

	return nil
}

// GetEnforcementStats retrieves enforcement statistics
func (r *PostgresEnforcementRepository) GetEnforcementStats(ctx context.Context, since time.Time) (*domain.EnforcementStats, error) {
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed
		FROM %s
		WHERE created_at >= $1
	`, r.tableName("enforcements"))

	var stats domain.EnforcementStats
	err := r.db.QueryRowContext(ctx, query, since).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.Completed,
		&stats.Failed,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get enforcement stats: %w", err)
	}

	return &stats, nil
}
