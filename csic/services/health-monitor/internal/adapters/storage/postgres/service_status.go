package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
)

// PostgresServiceStatusRepository implements ServiceStatusRepository using PostgreSQL
type PostgresServiceStatusRepository struct {
	db          *sql.DB
	tablePrefix string
}

// NewPostgresServiceStatusRepository creates a new PostgreSQL service status repository
func NewPostgresServiceStatusRepository(databaseURL string) (ports.ServiceStatusRepository, error) {
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

	return &PostgresServiceStatusRepository{
		db:          db,
		tablePrefix: "health_monitor_",
	}, nil
}

// Close closes the database connection
func (r *PostgresServiceStatusRepository) Close() error {
	return r.db.Close()
}

// tableName returns the prefixed table name
func (r *PostgresServiceStatusRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// GetServiceStatus retrieves service status by name
func (r *PostgresServiceStatusRepository) GetServiceStatus(ctx context.Context, serviceName string) (*domain.ServiceStatus, error) {
	query := fmt.Sprintf(`
		SELECT id, name, status, last_heartbeat, last_check, uptime, 
		       response_time, error_rate, cpu_usage, memory_usage, disk_usage,
		       message, created_at, updated_at
		FROM %s
		WHERE name = $1
	`, r.tableName("service_status"))

	var status domain.ServiceStatus
	err := r.db.QueryRowContext(ctx, query, serviceName).Scan(
		&status.ID,
		&status.Name,
		&status.Status,
		&status.LastHeartbeat,
		&status.LastCheck,
		&status.Uptime,
		&status.ResponseTime,
		&status.ErrorRate,
		&status.CPUUsage,
		&status.MemoryUsage,
		&status.DiskUsage,
		&status.Message,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service status: %w", err)
	}

	return &status, nil
}

// GetAllServiceStatuses retrieves all service statuses
func (r *PostgresServiceStatusRepository) GetAllServiceStatuses(ctx context.Context) ([]*domain.ServiceStatus, error) {
	query := fmt.Sprintf(`
		SELECT id, name, status, last_heartbeat, last_check, uptime, 
		       response_time, error_rate, cpu_usage, memory_usage, disk_usage,
		       message, created_at, updated_at
		FROM %s
		ORDER BY name ASC
	`, r.tableName("service_status"))

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query service statuses: %w", err)
	}
	defer rows.Close()

	var statuses []*domain.ServiceStatus

	for rows.Next() {
		var status domain.ServiceStatus
		if err := rows.Scan(
			&status.ID,
			&status.Name,
			&status.Status,
			&status.LastHeartbeat,
			&status.LastCheck,
			&status.Uptime,
			&status.ResponseTime,
			&status.ErrorRate,
			&status.CPUUsage,
			&status.MemoryUsage,
			&status.DiskUsage,
			&status.Message,
			&status.CreatedAt,
			&status.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan service status: %w", err)
		}
		statuses = append(statuses, &status)
	}

	return statuses, nil
}

// UpsertServiceStatus creates or updates a service status
func (r *PostgresServiceStatusRepository) UpsertServiceStatus(ctx context.Context, status *domain.ServiceStatus) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, status, last_heartbeat, last_check, uptime, 
		                response_time, error_rate, cpu_usage, memory_usage, disk_usage,
		                message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (name) DO UPDATE SET
			status = EXCLUDED.status,
			last_heartbeat = EXCLUDED.last_heartbeat,
			last_check = EXCLUDED.last_check,
			uptime = EXCLUDED.uptime,
			response_time = EXCLUDED.response_time,
			error_rate = EXCLUDED.error_rate,
			cpu_usage = EXCLUDED.cpu_usage,
			memory_usage = EXCLUDED.memory_usage,
			disk_usage = EXCLUDED.disk_usage,
			message = EXCLUDED.message,
			updated_at = EXCLUDED.updated_at
	`, r.tableName("service_status"))

	_, err := r.db.ExecContext(ctx, query,
		status.ID,
		status.Name,
		status.Status,
		status.LastHeartbeat,
		status.LastCheck,
		status.Uptime,
		status.ResponseTime,
		status.ErrorRate,
		status.CPUUsage,
		status.MemoryUsage,
		status.DiskUsage,
		status.Message,
		status.CreatedAt,
		status.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert service status: %w", err)
	}

	return nil
}

// UpdateServiceStatus updates a service status
func (r *PostgresServiceStatusRepository) UpdateServiceStatus(ctx context.Context, serviceName string, status domain.ServiceStatus) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, last_check = $2, uptime = $3, response_time = $4,
		    error_rate = $5, cpu_usage = $6, memory_usage = $7, disk_usage = $8,
		    message = $9, updated_at = $10
		WHERE name = $11
	`, r.tableName("service_status"))

	_, err := r.db.ExecContext(ctx, query,
		status.Status,
		status.LastCheck,
		status.Uptime,
		status.ResponseTime,
		status.ErrorRate,
		status.CPUUsage,
		status.MemoryUsage,
		status.DiskUsage,
		status.Message,
		time.Now(),
		serviceName,
	)

	if err != nil {
		return fmt.Errorf("failed to update service status: %w", err)
	}

	return nil
}

// DeleteServiceStatus deletes a service status
func (r *PostgresServiceStatusRepository) DeleteServiceStatus(ctx context.Context, serviceName string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE name = $1`, r.tableName("service_status"))

	_, err := r.db.ExecContext(ctx, query, serviceName)
	if err != nil {
		return fmt.Errorf("failed to delete service status: %w", err)
	}

	return nil
}

// GetServicesByStatus retrieves services by status
func (r *PostgresServiceStatusRepository) GetServicesByStatus(ctx context.Context, status string) ([]*domain.ServiceStatus, error) {
	query := fmt.Sprintf(`
		SELECT id, name, status, last_heartbeat, last_check, uptime, 
		       response_time, error_rate, cpu_usage, memory_usage, disk_usage,
		       message, created_at, updated_at
		FROM %s
		WHERE status = $1
		ORDER BY name ASC
	`, r.tableName("service_status"))

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query service statuses: %w", err)
	}
	defer rows.Close()

	var statuses []*domain.ServiceStatus

	for rows.Next() {
		var status domain.ServiceStatus
		if err := rows.Scan(
			&status.ID,
			&status.Name,
			&status.Status,
			&status.LastHeartbeat,
			&status.LastCheck,
			&status.Uptime,
			&status.ResponseTime,
			&status.ErrorRate,
			&status.CPUUsage,
			&status.MemoryUsage,
			&status.DiskUsage,
			&status.Message,
			&status.CreatedAt,
			&status.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan service status: %w", err)
		}
		statuses = append(statuses, &status)
	}

	return statuses, nil
}

// GetServiceUptime calculates service uptime percentage
func (r *PostgresServiceStatusRepository) GetServiceUptime(ctx context.Context, serviceName string, since time.Time) (float64, error) {
	query := fmt.Sprintf(`
		SELECT 
			COUNT(CASE WHEN status = 'healthy' THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0)
		FROM %s
		WHERE name = $1 AND updated_at >= $2
	`, r.tableName("service_status"))

	var uptime float64
	err := r.db.QueryRowContext(ctx, query, serviceName, since).Scan(&uptime)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate uptime: %w", err)
	}

	return uptime, nil
}
