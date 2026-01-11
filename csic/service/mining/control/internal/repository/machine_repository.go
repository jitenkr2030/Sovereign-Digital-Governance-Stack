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

// PostgresMachineRepository implements MachineRepository for PostgreSQL
type PostgresMachineRepository struct {
	db *sql.DB
}

// NewPostgresMachineRepository creates a new PostgreSQL machine repository
func NewPostgresMachineRepository(config PostgresConfig) (*PostgresMachineRepository, error) {
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

	return &PostgresMachineRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresMachineRepository) Close() error {
	return r.db.Close()
}

// Create creates a new mining machine
func (r *PostgresMachineRepository) Create(ctx context.Context, machine *domain.MiningMachine) error {
	query := `INSERT INTO mining_machines (
		id, serial_number, pool_id, model_type, manufacturer, model_name,
		hash_rate_spec_th, power_spec_watts, location_id, location_description,
		gps_coordinates, installation_date, last_maintenance_date, status, is_active,
		current_hash_rate_th, current_power_watts, total_energy_kwh, uptime_hours,
		created_at, updated_at, decommissioned_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`

	_, err := r.db.ExecContext(ctx, query,
		machine.ID, machine.SerialNumber, machine.PoolID, machine.ModelType, machine.Manufacturer, machine.ModelName,
		machine.HashRateSpecTH, machine.PowerSpecWatts, machine.LocationID, machine.LocationDescription,
		machine.GPSCoordinates, machine.InstallationDate, machine.LastMaintenanceDate, machine.Status, machine.IsActive,
		machine.CurrentHashRateTH, machine.CurrentPowerWatts, machine.TotalEnergyKWh, machine.UptimeHours,
		machine.CreatedAt, machine.UpdatedAt, machine.DecommissionedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create mining machine: %w", err)
	}

	return nil
}

// GetByID retrieves a mining machine by ID
func (r *PostgresMachineRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MiningMachine, error) {
	query := `SELECT id, serial_number, pool_id, model_type, manufacturer, model_name,
		hash_rate_spec_th, power_spec_watts, location_id, location_description,
		gps_coordinates, installation_date, last_maintenance_date, status, is_active,
		current_hash_rate_th, current_power_watts, total_energy_kwh, uptime_hours,
		created_at, updated_at, decommissioned_at
		FROM mining_machines WHERE id = $1`

	machine := &domain.MiningMachine{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&machine.ID, &machine.SerialNumber, &machine.PoolID, &machine.ModelType, &machine.Manufacturer, &machine.ModelName,
		&machine.HashRateSpecTH, &machine.PowerSpecWatts, &machine.LocationID, &machine.LocationDescription,
		&machine.GPSCoordinates, &machine.InstallationDate, &machine.LastMaintenanceDate, &machine.Status, &machine.IsActive,
		&machine.CurrentHashRateTH, &machine.CurrentPowerWatts, &machine.TotalEnergyKWh, &machine.UptimeHours,
		&machine.CreatedAt, &machine.UpdatedAt, &machine.DecommissionedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mining machine: %w", err)
	}

	return machine, nil
}

// GetBySerialNumber retrieves a mining machine by serial number
func (r *PostgresMachineRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (*domain.MiningMachine, error) {
	query := `SELECT id, serial_number, pool_id, model_type, manufacturer, model_name,
		hash_rate_spec_th, power_spec_watts, location_id, location_description,
		gps_coordinates, installation_date, last_maintenance_date, status, is_active,
		current_hash_rate_th, current_power_watts, total_energy_kwh, uptime_hours,
		created_at, updated_at, decommissioned_at
		FROM mining_machines WHERE serial_number = $1`

	machine := &domain.MiningMachine{}
	err := r.db.QueryRowContext(ctx, query, serialNumber).Scan(
		&machine.ID, &machine.SerialNumber, &machine.PoolID, &machine.ModelType, &machine.Manufacturer, &machine.ModelName,
		&machine.HashRateSpecTH, &machine.PowerSpecWatts, &machine.LocationID, &machine.LocationDescription,
		&machine.GPSCoordinates, &machine.InstallationDate, &machine.LastMaintenanceDate, &machine.Status, &machine.IsActive,
		&machine.CurrentHashRateTH, &machine.CurrentPowerWatts, &machine.TotalEnergyKWh, &machine.UptimeHours,
		&machine.CreatedAt, &machine.UpdatedAt, &machine.DecommissionedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mining machine: %w", err)
	}

	return machine, nil
}

// Update updates an existing mining machine
func (r *PostgresMachineRepository) Update(ctx context.Context, machine *domain.MiningMachine) error {
	query := `UPDATE mining_machines SET
		location_description = $1, last_maintenance_date = $2, status = $3, is_active = $4,
		current_hash_rate_th = $5, current_power_watts = $6, total_energy_kwh = $7,
		uptime_hours = $8, updated_at = $9, decommissioned_at = $10
		WHERE id = $11`

	_, err := r.db.ExecContext(ctx, query,
		machine.LocationDescription, machine.LastMaintenanceDate, machine.Status, machine.IsActive,
		machine.CurrentHashRateTH, machine.CurrentPowerWatts, machine.TotalEnergyKWh,
		machine.UptimeHours, machine.UpdatedAt, machine.DecommissionedAt, machine.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update mining machine: %w", err)
	}

	return nil
}

// Delete deletes a mining machine
func (r *PostgresMachineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mining_machines WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete mining machine: %w", err)
	}
	return nil
}

// ListByPool retrieves all machines for a pool
func (r *PostgresMachineRepository) ListByPool(ctx context.Context, poolID uuid.UUID, activeOnly bool, limit, offset int) ([]domain.MiningMachine, error) {
	query := `SELECT id, serial_number, pool_id, model_type, manufacturer, model_name,
		hash_rate_spec_th, power_spec_watts, location_id, location_description,
		gps_coordinates, installation_date, last_maintenance_date, status, is_active,
		current_hash_rate_th, current_power_watts, total_energy_kwh, uptime_hours,
		created_at, updated_at, decommissioned_at
		FROM mining_machines WHERE pool_id = $1`

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"

	rows, err := r.db.QueryContext(ctx, query, poolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list mining machines: %w", err)
	}
	defer rows.Close()

	var machines []domain.MiningMachine
	for rows.Next() {
		machine := &domain.MiningMachine{}
		err := rows.Scan(
			&machine.ID, &machine.SerialNumber, &machine.PoolID, &machine.ModelType, &machine.Manufacturer, &machine.ModelName,
			&machine.HashRateSpecTH, &machine.PowerSpecWatts, &machine.LocationID, &machine.LocationDescription,
			&machine.GPSCoordinates, &machine.InstallationDate, &machine.LastMaintenanceDate, &machine.Status, &machine.IsActive,
			&machine.CurrentHashRateTH, &machine.CurrentPowerWatts, &machine.TotalEnergyKWh, &machine.UptimeHours,
			&machine.CreatedAt, &machine.UpdatedAt, &machine.DecommissionedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mining machine: %w", err)
		}
		machines = append(machines, *machine)
	}

	return machines, nil
}

// CountByPool returns the count of machines for a pool
func (r *PostgresMachineRepository) CountByPool(ctx context.Context, poolID uuid.UUID, activeOnly bool) (int64, error) {
	query := `SELECT COUNT(*) FROM mining_machines WHERE pool_id = $1`
	if activeOnly {
		query += " AND is_active = true"
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, poolID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count mining machines: %w", err)
	}

	return count, nil
}

// BatchCreate creates multiple mining machines
func (r *PostgresMachineRepository) BatchCreate(ctx context.Context, machines []domain.MiningMachine) error {
	if len(machines) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO mining_machines (
		id, serial_number, pool_id, model_type, manufacturer, model_name,
		hash_rate_spec_th, power_spec_watts, location_id, location_description,
		gps_coordinates, installation_date, status, is_active,
		current_hash_rate_th, current_power_watts, total_energy_kwh, uptime_hours,
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, machine := range machines {
		_, err := stmt.ExecContext(ctx,
			machine.ID, machine.SerialNumber, machine.PoolID, machine.ModelType, machine.Manufacturer, machine.ModelName,
			machine.HashRateSpecTH, machine.PowerSpecWatts, machine.LocationID, machine.LocationDescription,
			machine.GPSCoordinates, machine.InstallationDate, machine.Status, machine.IsActive,
			machine.CurrentHashRateTH, machine.CurrentPowerWatts, machine.TotalEnergyKWh, machine.UptimeHours,
			machine.CreatedAt, machine.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert machine: %w", err)
		}
	}

	return tx.Commit()
}

// GetPoolMachinesStats returns statistics for machines in a pool
func (r *PostgresMachineRepository) GetPoolMachinesStats(ctx context.Context, poolID uuid.UUID) (*MachineStats, error) {
	query := `SELECT
		COUNT(*) as total,
		COUNT(*) FILTER (WHERE status = 'ACTIVE' AND is_active = true) as active,
		COUNT(*) FILTER (WHERE status = 'OFFLINE' OR is_active = false) as offline,
		COUNT(*) FILTER (WHERE status = 'MAINTENANCE') as maintenance
		FROM mining_machines WHERE pool_id = $1`

	stats := &MachineStats{}
	err := r.db.QueryRowContext(ctx, query, poolID).Scan(
		&stats.TotalMachines, &stats.ActiveMachines, &stats.OfflineMachines, &stats.MaintenanceCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine stats: %w", err)
	}

	return stats, nil
}
