package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
)

// PostgresOutageRepository implements OutageRepository using PostgreSQL
type PostgresOutageRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresOutageRepository creates a new PostgreSQL outage repository
func NewPostgresOutageRepository(databaseURL string) (ports.OutageRepository, error) {
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

	return &PostgresOutageRepository{
		db:          db,
		tablePrefix: "health_monitor_",
	}, nil
}

// Close closes the database connection
func (r *PostgresOutageRepository) Close() error {
	return r.db.Close()
}

// tableName returns the prefixed table name
func (r *PostgresOutageRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// GetOutage retrieves an outage by ID
func (r *PostgresOutageRepository) GetOutage(ctx context.Context, id string) (*domain.Outage, error) {
	query := fmt.Sprintf(`
		SELECT id, service_name, instance_id, severity, status, started_at, detected_at,
		       resolved_at, duration, impact, description, root_cause, affected_users,
		       metadata, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.tableName("outages"))

	var outage domain.Outage
	var instanceID, rootCause sql.NullString
	var resolvedAt sql.NullTime
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&outage.ID,
		&outage.ServiceName,
		&instanceID,
		&outage.Severity,
		&outage.Status,
		&outage.StartedAt,
		&outage.DetectedAt,
		&resolvedAt,
		&outage.Duration,
		&outage.Impact,
		&outage.Description,
		&rootCause,
		&outage.AffectedUsers,
		&metadata,
		&outage.CreatedAt,
		&outage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get outage: %w", err)
	}

	if instanceID.Valid {
		outage.InstanceID = instanceID.String
	}
	if rootCause.Valid {
		outage.RootCause = rootCause.String
	}
	if resolvedAt.Valid {
		outage.ResolvedAt = resolvedAt.Time
	}

	return &outage, nil
}

// GetActiveOutages retrieves all active outages
func (r *PostgresOutageRepository) GetActiveOutages(ctx context.Context) ([]*domain.Outage, error) {
	query := fmt.Sprintf(`
		SELECT id, service_name, instance_id, severity, status, started_at, detected_at,
		       resolved_at, duration, impact, description, root_cause, affected_users,
		       metadata, created_at, updated_at
		FROM %s
		WHERE status != 'resolved'
		ORDER BY severity DESC, started_at DESC
	`, r.tableName("outages"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query outages: %w", err)
	}
	defer rows.Close()

	var outages []*domain.Outage

	for rows.Next() {
		var outage domain.Outage
		var instanceID, rootCause sql.NullString
		var resolvedAt sql.NullTime
		var metadata []byte

		if err := rows.Scan(
			&outage.ID,
			&outage.ServiceName,
			&instanceID,
			&outage.Severity,
			&outage.Status,
			&outage.StartedAt,
			&outage.DetectedAt,
			&resolvedAt,
			&outage.Duration,
			&outage.Impact,
			&outage.Description,
			&rootCause,
			&outage.AffectedUsers,
			&metadata,
			&outage.CreatedAt,
			&outage.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", err)
		}

		if instanceID.Valid {
			outage.InstanceID = instanceID.String
		}
		if rootCause.Valid {
			outage.RootCause = rootCause.String
		}
		if resolvedAt.Valid {
			outage.ResolvedAt = resolvedAt.Time
		}

		outages = append(outages, &outage)
	}

	return outages, nil
}

// GetOutagesByService retrieves outages for a specific service
func (r *PostgresOutageRepository) GetOutagesByService(ctx context.Context, serviceName string, limit int) ([]*domain.Outage, error) {
	query := fmt.Sprintf(`
		SELECT id, service_name, instance_id, severity, status, started_at, detected_at,
		       resolved_at, duration, impact, description, root_cause, affected_users,
		       metadata, created_at, updated_at
		FROM %s
		WHERE service_name = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, r.tableName("outages"))

	rows, err := r.db.QueryContext(ctx, query, serviceName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outages: %w", err)
	}
	defer rows.Close()

	var outages []*domain.Outage

	for rows.Next() {
		var outage domain.Outage
		var instanceID, rootCause sql.NullString
		var resolvedAt sql.NullTime
		var metadata []byte

		if err := rows.Scan(
			&outage.ID,
			&outage.ServiceName,
			&instanceID,
			&outage.Severity,
			&outage.Status,
			&outage.StartedAt,
			&outage.DetectedAt,
			&resolvedAt,
			&outage.Duration,
			&outage.Impact,
			&outage.Description,
			&rootCause,
			&outage.AffectedUsers,
			&metadata,
			&outage.CreatedAt,
			&outage.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", err)
		}

		if instanceID.Valid {
			outage.InstanceID = instanceID.String
		}
		if rootCause.Valid {
			outage.RootCause = rootCause.String
		}
		if resolvedAt.Valid {
			outage.ResolvedAt = resolvedAt.Time
		}

		outages = append(outages, &outage)
	}

	return outages, nil
}

// GetRecentOutages retrieves recent outages
func (r *PostgresOutageRepository) GetRecentOutages(ctx context.Context, since time.Time) ([]*domain.Outage, error) {
	query := fmt.Sprintf(`
		SELECT id, service_name, instance_id, severity, status, started_at, detected_at,
		       resolved_at, duration, impact, description, root_cause, affected_users,
		       metadata, created_at, updated_at
		FROM %s
		WHERE detected_at >= $1
		ORDER BY detected_at DESC
	`, r.tableName("outages"))

	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query outages: %w", err)
	}
	defer rows.Close()

	var outages []*domain.Outage

	for rows.Next() {
		var outage domain.Outage
		var instanceID, rootCause sql.NullString
		var resolvedAt sql.NullTime
		var metadata []byte

		if err := rows.Scan(
			&outage.ID,
			&outage.ServiceName,
			&instanceID,
			&outage.Severity,
			&outage.Status,
			&outage.StartedAt,
			&outage.DetectedAt,
			&resolvedAt,
			&outage.Duration,
			&outage.Impact,
			&outage.Description,
			&rootCause,
			&outage.AffectedUsers,
			&metadata,
			&outage.CreatedAt,
			&outage.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", err)
		}

		if instanceID.Valid {
			outage.InstanceID = instanceID.String
		}
		if rootCause.Valid {
			outage.RootCause = rootCause.String
		}
		if resolvedAt.Valid {
			outage.ResolvedAt = resolvedAt.Time
		}

		outages = append(outages, &outage)
	}

	return outages, nil
}

// CreateOutage creates a new outage
func (r *PostgresOutageRepository) CreateOutage(ctx context.Context, outage *domain.Outage) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, service_name, instance_id, severity, status, started_at, detected_at,
		                resolved_at, duration, impact, description, root_cause, affected_users,
		                metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, r.tableName("outages"))

	_, err := r.db.ExecContext(ctx, query,
		outage.ID,
		outage.ServiceName,
		outage.InstanceID,
		outage.Severity,
		outage.Status,
		outage.StartedAt,
		outage.DetectedAt,
		outage.ResolvedAt,
		outage.Duration,
		outage.Impact,
		outage.Description,
		outage.RootCause,
		outage.AffectedUsers,
		[]byte("{}"),
		outage.CreatedAt,
		outage.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create outage: %w", err)
	}

	return nil
}

// UpdateOutageStatus updates the status of an outage
func (r *PostgresOutageRepository) UpdateOutageStatus(ctx context.Context, id string, status string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, r.tableName("outages"))

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update outage status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("outage not found: %s", id)
	}

	return nil
}

// ResolveOutage resolves an outage
func (r *PostgresOutageRepository) ResolveOutage(ctx context.Context, id string, rootCause string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, root_cause = $2, resolved_at = $3, duration = $4, updated_at = $5
		WHERE id = $6
	`, r.tableName("outages"))

	now := time.Now()
	var duration int

	// Calculate duration from started_at
	var startedAt time.Time
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT started_at FROM %s WHERE id = $1", r.tableName("outages")), id).Scan(&startedAt)
	if err == nil {
		duration = int(now.Sub(startedAt).Seconds())
	}

	result, err := r.db.ExecContext(ctx, query,
		domain.OutageStatusResolved,
		rootCause,
		now,
		duration,
		now,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to resolve outage: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("outage not found: %s", id)
	}

	return nil
}

// GetOutage retrieves an outage by ID
func (r *PostgresOutageRepository) GetAlert(ctx context.Context, id string) (*domain.Alert, error) {
	return nil, nil
}

// GetAlerts retrieves alerts based on filter
func (r *PostgresOutageRepository) GetAlerts(ctx context.Context, filter ports.AlertFilter) ([]*domain.Alert, error) {
	return nil, nil
}

// GetFiringAlerts retrieves all currently firing alerts
func (r *PostgresOutageRepository) GetFiringAlerts(ctx context.Context) ([]*domain.Alert, error) {
	return nil, nil
}

// GetAlertsByService retrieves alerts for a specific service
func (r *PostgresOutageRepository) GetAlertsByService(ctx context.Context, serviceName string, limit int) ([]*domain.Alert, error) {
	return nil, nil
}

// CreateAlert creates a new alert
func (r *PostgresOutageRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	return nil
}

// UpdateAlertStatus updates the status of an alert
func (r *PostgresOutageRepository) UpdateAlertStatus(ctx context.Context, id string, status string) error {
	return nil
}

// ResolveAlert resolves an alert
func (r *PostgresOutageRepository) ResolveAlert(ctx context.Context, id string) error {
	return nil
}

// DeleteResolvedAlerts deletes resolved alerts older than specified time
func (r *PostgresOutageRepository) DeleteResolvedAlerts(ctx context.Context, olderThan time.Time) error {
	return nil
}
