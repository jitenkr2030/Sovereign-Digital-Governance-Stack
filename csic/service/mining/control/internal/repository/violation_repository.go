package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresViolationRepository implements ViolationRepository for PostgreSQL
type PostgresViolationRepository struct {
	db *sql.DB
}

// NewPostgresViolationRepository creates a new PostgreSQL violation repository
func NewPostgresViolationRepository(config PostgresConfig) (*PostgresViolationRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresViolationRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresViolationRepository) Close() error {
	return r.db.Close()
}

// Create creates a new compliance violation
func (r *PostgresViolationRepository) Create(ctx context.Context, violation *domain.ComplianceViolation) error {
	detailsJSON, _ := json.Marshal(violation.Details)
	evidenceJSON, _ := json.Marshal(violation.Evidence)
	notesJSON, _ := json.Marshal(violation.Notes)

	query := `INSERT INTO compliance_violations (
		id, pool_id, pool_name, violation_type, severity, status, title, description,
		details, evidence, detected_at, resolved_at, resolved_by, resolution,
		assigned_to, alerts_sent, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err := r.db.ExecContext(ctx, query,
		violation.ID, violation.PoolID, violation.PoolName, violation.ViolationType,
		violation.Severity, violation.Status, violation.Title, violation.Description,
		detailsJSON, evidenceJSON, violation.DetectedAt, violation.ResolvedAt,
		violation.ResolvedBy, violation.Resolution, violation.AssignedTo,
		violation.AlertsSent, violation.CreatedAt, violation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create violation: %w", err)
	}

	return nil
}

// GetByID retrieves a violation by ID
func (r *PostgresViolationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ComplianceViolation, error) {
	query := `SELECT id, pool_id, pool_name, violation_type, severity, status, title, description,
		details, evidence, detected_at, resolved_at, resolved_by, resolution,
		assigned_to, alerts_sent, created_at, updated_at
		FROM compliance_violations WHERE id = $1`

	violation := &domain.ComplianceViolation{}
	var detailsJSON, evidenceJSON, notesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&violation.ID, &violation.PoolID, &violation.PoolName, &violation.ViolationType,
		&violation.Severity, &violation.Status, &violation.Title, &violation.Description,
		&detailsJSON, &evidenceJSON, &violation.DetectedAt, &violation.ResolvedAt,
		&violation.ResolvedBy, &violation.Resolution, &violation.AssignedTo,
		&violation.AlertsSent, &violation.CreatedAt, &violation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get violation: %w", err)
	}

	json.Unmarshal(detailsJSON, &violation.Details)
	json.Unmarshal(evidenceJSON, &violation.Evidence)
	json.Unmarshal(notesJSON, &violation.Notes)

	return violation, nil
}

// Update updates an existing violation
func (r *PostgresViolationRepository) Update(ctx context.Context, violation *domain.ComplianceViolation) error {
	detailsJSON, _ := json.Marshal(violation.Details)
	evidenceJSON, _ := json.Marshal(violation.Evidence)
	notesJSON, _ := json.Marshal(violation.Notes)

	query := `UPDATE compliance_violations SET
		status = $1, details = $2, evidence = $3, resolved_at = $4, resolved_by = $5,
		resolution = $6, assigned_to = $7, alerts_sent = $8, updated_at = $9,
		notes = $10 WHERE id = $11`

	_, err := r.db.ExecContext(ctx, query,
		violation.Status, detailsJSON, evidenceJSON, violation.ResolvedAt,
		violation.ResolvedBy, violation.Resolution, violation.AssignedTo,
		violation.AlertsSent, violation.UpdatedAt, notesJSON, violation.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update violation: %w", err)
	}

	return nil
}

// GetOpenViolations retrieves all open violations for a pool
func (r *PostgresViolationRepository) GetOpenViolations(ctx context.Context, poolID uuid.UUID) ([]domain.ComplianceViolation, error) {
	query := `SELECT id, pool_id, pool_name, violation_type, severity, status, title, description,
		details, evidence, detected_at, resolved_at, resolved_by, resolution,
		assigned_to, alerts_sent, created_at, updated_at
		FROM compliance_violations WHERE pool_id = $1 AND status IN ('OPEN', 'INVESTIGATING')
		ORDER BY detected_at DESC`

	rows, err := r.db.QueryContext(ctx, query, poolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get open violations: %w", err)
	}
	defer rows.Close()

	var violations []domain.ComplianceViolation
	for rows.Next() {
		violation := &domain.ComplianceViolation{}
		var detailsJSON, evidenceJSON, notesJSON []byte

		err := rows.Scan(
			&violation.ID, &violation.PoolID, &violation.PoolName, &violation.ViolationType,
			&violation.Severity, &violation.Status, &violation.Title, &violation.Description,
			&detailsJSON, &evidenceJSON, &violation.DetectedAt, &violation.ResolvedAt,
			&violation.ResolvedBy, &violation.Resolution, &violation.AssignedTo,
			&violation.AlertsSent, &violation.CreatedAt, &violation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", err)
		}

		json.Unmarshal(detailsJSON, &violation.Details)
		json.Unmarshal(evidenceJSON, &violation.Evidence)
		json.Unmarshal(notesJSON, &violation.Notes)
		violations = append(violations, *violation)
	}

	return violations, nil
}

// List retrieves violations based on filter criteria
func (r *PostgresViolationRepository) List(ctx context.Context, filter domain.ViolationFilter, limit, offset int) ([]domain.ComplianceViolation, error) {
	query := `SELECT id, pool_id, pool_name, violation_type, severity, status, title, description,
		details, evidence, detected_at, resolved_at, resolved_by, resolution,
		assigned_to, alerts_sent, created_at, updated_at
		FROM compliance_violations WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	if len(filter.Types) > 0 {
		query += fmt.Sprintf(" AND violation_type = ANY($%d)", argNum)
		args = append(args, filter.Types)
		argNum++
	}

	if len(filter.Severities) > 0 {
		query += fmt.Sprintf(" AND severity = ANY($%d)", argNum)
		args = append(args, filter.Severities)
		argNum++
	}

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argNum)
		args = append(args, filter.Statuses)
		argNum++
	}

	if len(filter.PoolIDs) > 0 {
		query += fmt.Sprintf(" AND pool_id = ANY($%d)", argNum)
		args = append(args, filter.PoolIDs)
		argNum++
	}

	if filter.DetectedAfter != nil {
		query += fmt.Sprintf(" AND detected_at >= $%d", argNum)
		args = append(args, *filter.DetectedAfter)
		argNum++
	}

	if filter.DetectedBefore != nil {
		query += fmt.Sprintf(" AND detected_at <= $%d", argNum)
		args = append(args, *filter.DetectedBefore)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY detected_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list violations: %w", err)
	}
	defer rows.Close()

	var violations []domain.ComplianceViolation
	for rows.Next() {
		violation := &domain.ComplianceViolation{}
		var detailsJSON, evidenceJSON, notesJSON []byte

		err := rows.Scan(
			&violation.ID, &violation.PoolID, &violation.PoolName, &violation.ViolationType,
			&violation.Severity, &violation.Status, &violation.Title, &violation.Description,
			&detailsJSON, &evidenceJSON, &violation.DetectedAt, &violation.ResolvedAt,
			&violation.ResolvedBy, &violation.Resolution, &violation.AssignedTo,
			&violation.AlertsSent, &violation.CreatedAt, &violation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", err)
		}

		json.Unmarshal(detailsJSON, &violation.Details)
		json.Unmarshal(evidenceJSON, &violation.Evidence)
		json.Unmarshal(notesJSON, &violation.Notes)
		violations = append(violations, *violation)
	}

	return violations, nil
}

// PostgresComplianceRepository implements compliance repository for PostgreSQL
type PostgresComplianceRepository struct {
	db *sql.DB
}

// NewPostgresComplianceRepository creates a new PostgreSQL compliance repository
func NewPostgresComplianceRepository(config PostgresConfig) (*PostgresComplianceRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresComplianceRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresComplianceRepository) Close() error {
	return r.db.Close()
}

// This is a placeholder implementation - in a real system, this would integrate
// with the Compliance Management module via gRPC or HTTP
