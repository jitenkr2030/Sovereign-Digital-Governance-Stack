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

// PostgresInterventionRepository implements InterventionRepository using PostgreSQL
type PostgresInterventionRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresInterventionRepository creates a new PostgreSQL intervention repository
func NewPostgresInterventionRepository(databaseURL string) (ports.InterventionRepository, error) {
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

	return &PostgresInterventionRepository{
		db:          db,
		tablePrefix: "control_layer_",
	}, nil
}

// Close closes the database connection
func (r *PostgresInterventionRepository) Close() error {
	return r.db.Close()
}

// tableName returns the prefixed table name
func (r *PostgresInterventionRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// CreateIntervention creates a new intervention record
func (r *PostgresInterventionRepository) CreateIntervention(ctx context.Context, intervention *domain.Intervention) error {
	metadataJSON, err := json.Marshal(intervention.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal intervention metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, policy_id, enforcement_id, target_service, intervention_type, 
		                severity, status, reason, resolution, metadata, 
		                started_at, ended_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, r.tableName("interventions"))

	_, err = r.db.ExecContext(ctx, query,
		intervention.ID,
		intervention.PolicyID,
		intervention.EnforcementID,
		intervention.TargetService,
		intervention.InterventionType,
		intervention.Severity,
		intervention.Status,
		intervention.Reason,
		intervention.Resolution,
		metadataJSON,
		intervention.StartedAt,
		intervention.EndedAt,
		intervention.CreatedAt,
		intervention.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create intervention: %w", err)
	}

	return nil
}

// GetInterventionByID retrieves an intervention by ID
func (r *PostgresInterventionRepository) GetInterventionByID(ctx context.Context, id uuid.UUID) (*domain.Intervention, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, enforcement_id, target_service, intervention_type, 
		       severity, status, reason, resolution, metadata,
		       started_at, ended_at, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.tableName("interventions"))

	var intervention domain.Intervention
	var metadataJSON []byte
	var enforcementID, reason, resolution sql.NullString
	var startedAt, endedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&intervention.ID,
		&intervention.PolicyID,
		&enforcementID,
		&intervention.TargetService,
		&intervention.InterventionType,
		&intervention.Severity,
		&intervention.Status,
		&reason,
		&resolution,
		&metadataJSON,
		&startedAt,
		&endedAt,
		&intervention.CreatedAt,
		&intervention.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get intervention: %w", err)
	}

	if enforcementID.Valid {
		intervention.EnforcementID = enforcementID.String
	}
	if reason.Valid {
		intervention.Reason = reason.String
	}
	if resolution.Valid {
		intervention.Resolution = resolution.String
	}
	if startedAt.Valid {
		intervention.StartedAt = startedAt.Time
	}
	if endedAt.Valid {
		intervention.EndedAt = endedAt.Time
	}

	if err := json.Unmarshal(metadataJSON, &intervention.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &intervention, nil
}

// GetActiveInterventions retrieves all active interventions
func (r *PostgresInterventionRepository) GetActiveInterventions(ctx context.Context) ([]*domain.Intervention, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, enforcement_id, target_service, intervention_type, 
		       severity, status, reason, resolution, metadata,
		       started_at, ended_at, created_at, updated_at
		FROM %s
		WHERE status IN ('pending', 'active')
		ORDER BY severity DESC, created_at DESC
	`, r.tableName("interventions"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query interventions: %w", err)
	}
	defer rows.Close()

	var interventions []*domain.Intervention

	for rows.Next() {
		var intervention domain.Intervention
		var metadataJSON []byte
		var enforcementID, reason, resolution sql.NullString
		var startedAt, endedAt sql.NullTime

		if err := rows.Scan(
			&intervention.ID,
			&intervention.PolicyID,
			&enforcementID,
			&intervention.TargetService,
			&intervention.InterventionType,
			&intervention.Severity,
			&intervention.Status,
			&reason,
			&resolution,
			&metadataJSON,
			&startedAt,
			&endedAt,
			&intervention.CreatedAt,
			&intervention.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan intervention: %w", err)
		}

		if enforcementID.Valid {
			intervention.EnforcementID = enforcementID.String
		}
		if reason.Valid {
			intervention.Reason = reason.String
		}
		if resolution.Valid {
			intervention.Resolution = resolution.String
		}
		if startedAt.Valid {
			intervention.StartedAt = startedAt.Time
		}
		if endedAt.Valid {
			intervention.EndedAt = endedAt.Time
		}

		if err := json.Unmarshal(metadataJSON, &intervention.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		interventions = append(interventions, &intervention)
	}

	return interventions, nil
}

// GetInterventionsByPolicy retrieves interventions for a specific policy
func (r *PostgresInterventionRepository) GetInterventionsByPolicy(ctx context.Context, policyID uuid.UUID, limit int) ([]*domain.Intervention, error) {
	query := fmt.Sprintf(`
		SELECT id, policy_id, enforcement_id, target_service, intervention_type, 
		       severity, status, reason, resolution, metadata,
		       started_at, ended_at, created_at, updated_at
		FROM %s
		WHERE policy_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, r.tableName("interventions"))

	rows, err := r.db.QueryContext(ctx, query, policyID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query interventions: %w", err)
	}
	defer rows.Close()

	var interventions []*domain.Intervention

	for rows.Next() {
		var intervention domain.Intervention
		var metadataJSON []byte
		var enforcementID, reason, resolution sql.NullString
		var startedAt, endedAt sql.NullTime

		if err := rows.Scan(
			&intervention.ID,
			&intervention.PolicyID,
			&enforcementID,
			&intervention.TargetService,
			&intervention.InterventionType,
			&intervention.Severity,
			&intervention.Status,
			&reason,
			&resolution,
			&metadataJSON,
			&startedAt,
			&endedAt,
			&intervention.CreatedAt,
			&intervention.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan intervention: %w", err)
		}

		if enforcementID.Valid {
			intervention.EnforcementID = enforcementID.String
		}
		if reason.Valid {
			intervention.Reason = reason.String
		}
		if resolution.Valid {
			intervention.Resolution = resolution.String
		}
		if startedAt.Valid {
			intervention.StartedAt = startedAt.Time
		}
		if endedAt.Valid {
			intervention.EndedAt = endedAt.Time
		}

		if err := json.Unmarshal(metadataJSON, &intervention.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		interventions = append(interventions, &intervention)
	}

	return interventions, nil
}

// UpdateInterventionStatus updates the status of an intervention
func (r *PostgresInterventionRepository) UpdateInterventionStatus(ctx context.Context, id uuid.UUID, status domain.InterventionStatus) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, r.tableName("interventions"))

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update intervention status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("intervention not found: %s", id)
	}

	return nil
}

// ResolveIntervention resolves an intervention
func (r *PostgresInterventionRepository) ResolveIntervention(ctx context.Context, id uuid.UUID, resolution string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, resolution = $2, ended_at = $3, updated_at = $4
		WHERE id = $5
	`, r.tableName("interventions"))

	result, err := r.db.ExecContext(ctx, query, domain.InterventionStatusResolved, resolution, time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to resolve intervention: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("intervention not found: %s", id)
	}

	return nil
}

// GetInterventionStats retrieves intervention statistics
func (r *PostgresInterventionRepository) GetInterventionStats(ctx context.Context, since time.Time) (*domain.InterventionStats, error) {
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'resolved' THEN 1 END) as resolved,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled
		FROM %s
		WHERE created_at >= $1
	`, r.tableName("interventions"))

	var stats domain.InterventionStats
	err := r.db.QueryRowContext(ctx, query, since).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.Active,
		&stats.Resolved,
		&stats.Cancelled,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get intervention stats: %w", err)
	}

	return &stats, nil
}
