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

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// PostgresMiningPoolRepository implements MiningPoolRepository for PostgreSQL
type PostgresMiningPoolRepository struct {
	db *sql.DB
}

// NewPostgresMiningPoolRepository creates a new PostgreSQL mining pool repository
func NewPostgresMiningPoolRepository(config PostgresConfig) (*PostgresMiningPoolRepository, error) {
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

	return &PostgresMiningPoolRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresMiningPoolRepository) Close() error {
	return r.db.Close()
}

// Create creates a new mining pool
func (r *PostgresMiningPoolRepository) Create(ctx context.Context, pool *domain.MiningPool) error {
	energySourcesJSON, _ := json.Marshal(pool.AllowedEnergySources)

	query := `INSERT INTO mining_pools (
		id, license_number, name, owner_entity_id, owner_entity_name, status, region_code,
		max_allowed_energy_kw, current_energy_usage_kw, allowed_energy_sources,
		total_hash_rate_th, active_machine_count, total_machine_count, carbon_footprint_kg,
		license_issued_at, license_expires_at, contact_email, contact_phone, facility_address,
		gps_coordinates, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`

	_, err := r.db.ExecContext(ctx, query,
		pool.ID, pool.LicenseNumber, pool.Name, pool.OwnerEntityID, pool.OwnerEntityName,
		pool.Status, pool.RegionCode, pool.MaxAllowedEnergyKW, pool.CurrentEnergyUsageKW,
		energySourcesJSON, pool.TotalHashRateTH, pool.ActiveMachineCount, pool.TotalMachineCount,
		pool.CarbonFootprintKG, pool.LicenseIssuedAt, pool.LicenseExpiresAt,
		pool.ContactEmail, pool.ContactPhone, pool.FacilityAddress, pool.GPSCoordinates,
		pool.CreatedAt, pool.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create mining pool: %w", err)
	}

	return nil
}

// GetByID retrieves a mining pool by ID
func (r *PostgresMiningPoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MiningPool, error) {
	query := `SELECT id, license_number, name, owner_entity_id, owner_entity_name, status, region_code,
		max_allowed_energy_kw, current_energy_usage_kw, allowed_energy_sources,
		total_hash_rate_th, active_machine_count, total_machine_count, carbon_footprint_kg,
		license_issued_at, license_expires_at, contact_email, contact_phone, facility_address,
		gps_coordinates, created_at, updated_at
		FROM mining_pools WHERE id = $1`

	pool := &domain.MiningPool{}
	var energySourcesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pool.ID, &pool.LicenseNumber, &pool.Name, &pool.OwnerEntityID, &pool.OwnerEntityName,
		&pool.Status, &pool.RegionCode, &pool.MaxAllowedEnergyKW, &pool.CurrentEnergyUsageKW,
		&energySourcesJSON, &pool.TotalHashRateTH, &pool.ActiveMachineCount, &pool.TotalMachineCount,
		&pool.CarbonFootprintKG, &pool.LicenseIssuedAt, &pool.LicenseExpiresAt,
		&pool.ContactEmail, &pool.ContactPhone, &pool.FacilityAddress, &pool.GPSCoordinates,
		&pool.CreatedAt, &pool.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mining pool: %w", err)
	}

	if err := json.Unmarshal(energySourcesJSON, &pool.AllowedEnergySources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal energy sources: %w", err)
	}

	return pool, nil
}

// GetByLicenseNumber retrieves a mining pool by license number
func (r *PostgresMiningPoolRepository) GetByLicenseNumber(ctx context.Context, licenseNumber string) (*domain.MiningPool, error) {
	query := `SELECT id, license_number, name, owner_entity_id, owner_entity_name, status, region_code,
		max_allowed_energy_kw, current_energy_usage_kw, allowed_energy_sources,
		total_hash_rate_th, active_machine_count, total_machine_count, carbon_footprint_kg,
		license_issued_at, license_expires_at, contact_email, contact_phone, facility_address,
		gps_coordinates, created_at, updated_at
		FROM mining_pools WHERE license_number = $1`

	pool := &domain.MiningPool{}
	var energySourcesJSON []byte

	err := r.db.QueryRowContext(ctx, query, licenseNumber).Scan(
		&pool.ID, &pool.LicenseNumber, &pool.Name, &pool.OwnerEntityID, &pool.OwnerEntityName,
		&pool.Status, &pool.RegionCode, &pool.MaxAllowedEnergyKW, &pool.CurrentEnergyUsageKW,
		&energySourcesJSON, &pool.TotalHashRateTH, &pool.ActiveMachineCount, &pool.TotalMachineCount,
		&pool.CarbonFootprintKG, &pool.LicenseIssuedAt, &pool.LicenseExpiresAt,
		&pool.ContactEmail, &pool.ContactPhone, &pool.FacilityAddress, &pool.GPSCoordinates,
		&pool.CreatedAt, &pool.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mining pool: %w", err)
	}

	if err := json.Unmarshal(energySourcesJSON, &pool.AllowedEnergySources); err != nil {
		return nil, fmt.Errorf("failed to unmarshal energy sources: %w", err)
	}

	return pool, nil
}

// Update updates an existing mining pool
func (r *PostgresMiningPoolRepository) Update(ctx context.Context, pool *domain.MiningPool) error {
	energySourcesJSON, _ := json.Marshal(pool.AllowedEnergySources)

	query := `UPDATE mining_pools SET
		name = $1, owner_entity_name = $2, status = $3, region_code = $4,
		max_allowed_energy_kw = $5, current_energy_usage_kw = $6, allowed_energy_sources = $7,
		total_hash_rate_th = $8, active_machine_count = $9, total_machine_count = $10,
		carbon_footprint_kg = $11, license_issued_at = $12, license_expires_at = $13,
		contact_email = $14, contact_phone = $15, facility_address = $16,
		gps_coordinates = $17, updated_at = $18 WHERE id = $19`

	_, err := r.db.ExecContext(ctx, query,
		pool.Name, pool.OwnerEntityName, pool.Status, pool.RegionCode,
		pool.MaxAllowedEnergyKW, pool.CurrentEnergyUsageKW, energySourcesJSON,
		pool.TotalHashRateTH, pool.ActiveMachineCount, pool.TotalMachineCount,
		pool.CarbonFootprintKG, pool.LicenseIssuedAt, pool.LicenseExpiresAt,
		pool.ContactEmail, pool.ContactPhone, pool.FacilityAddress, pool.GPSCoordinates,
		pool.UpdatedAt, pool.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update mining pool: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a mining pool
func (r *PostgresMiningPoolRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PoolStatus) error {
	query := `UPDATE mining_pools SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update pool status: %w", err)
	}
	return nil
}

// Delete deletes a mining pool
func (r *PostgresMiningPoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mining_pools WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete mining pool: %w", err)
	}
	return nil
}

// List retrieves mining pools based on filter criteria
func (r *PostgresMiningPoolRepository) List(ctx context.Context, filter domain.PoolFilter, limit, offset int) ([]domain.MiningPool, error) {
	query := `SELECT id, license_number, name, owner_entity_id, owner_entity_name, status, region_code,
		max_allowed_energy_kw, current_energy_usage_kw, allowed_energy_sources,
		total_hash_rate_th, active_machine_count, total_machine_count, carbon_footprint_kg,
		license_issued_at, license_expires_at, contact_email, contact_phone, facility_address,
		gps_coordinates, created_at, updated_at
		FROM mining_pools WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argNum)
		args = append(args, filter.Statuses)
		argNum++
	}

	if len(filter.RegionCodes) > 0 {
		query += fmt.Sprintf(" AND region_code = ANY($%d)", argNum)
		args = append(args, filter.RegionCodes)
		argNum++
	}

	if len(filter.EntityIDs) > 0 {
		query += fmt.Sprintf(" AND owner_entity_id = ANY($%d)", argNum)
		args = append(args, filter.EntityIDs)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list mining pools: %w", err)
	}
	defer rows.Close()

	var pools []domain.MiningPool
	for rows.Next() {
		pool := &domain.MiningPool{}
		var energySourcesJSON []byte

		err := rows.Scan(
			&pool.ID, &pool.LicenseNumber, &pool.Name, &pool.OwnerEntityID, &pool.OwnerEntityName,
			&pool.Status, &pool.RegionCode, &pool.MaxAllowedEnergyKW, &pool.CurrentEnergyUsageKW,
			&energySourcesJSON, &pool.TotalHashRateTH, &pool.ActiveMachineCount, &pool.TotalMachineCount,
			&pool.CarbonFootprintKG, &pool.LicenseIssuedAt, &pool.LicenseExpiresAt,
			&pool.ContactEmail, &pool.ContactPhone, &pool.FacilityAddress, &pool.GPSCoordinates,
			&pool.CreatedAt, &pool.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mining pool: %w", err)
		}

		json.Unmarshal(energySourcesJSON, &pool.AllowedEnergySources)
		pools = append(pools, *pool)
	}

	return pools, nil
}

// Count returns the count of mining pools matching the filter
func (r *PostgresMiningPoolRepository) Count(ctx context.Context, filter domain.PoolFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM mining_pools WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argNum)
		args = append(args, filter.Statuses)
		argNum++
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count mining pools: %w", err)
	}

	return count, nil
}

// IncrementMachineCount increments the machine count for a pool
func (r *PostgresMiningPoolRepository) IncrementMachineCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE mining_pools SET total_machine_count = total_machine_count + 1, active_machine_count = active_machine_count + 1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment machine count: %w", err)
	}
	return nil
}

// DecrementMachineCount decrements the machine count for a pool
func (r *PostgresMiningPoolRepository) DecrementMachineCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE mining_pools SET total_machine_count = total_machine_count - 1, active_machine_count = active_machine_count - 1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to decrement machine count: %w", err)
	}
	return nil
}

// UpdateEnergyUsage updates the energy usage for a pool
func (r *PostgresMiningPoolRepository) UpdateEnergyUsage(ctx context.Context, id uuid.UUID, usageKW decimal.Decimal) error {
	query := `UPDATE mining_pools SET current_energy_usage_kw = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, usageKW, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update energy usage: %w", err)
	}
	return nil
}

// UpdateHashRate updates the hash rate for a pool
func (r *PostgresMiningPoolRepository) UpdateHashRate(ctx context.Context, id uuid.UUID, hashRateTH decimal.Decimal) error {
	query := `UPDATE mining_pools SET total_hash_rate_th = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, hashRateTH, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update hash rate: %w", err)
	}
	return nil
}
