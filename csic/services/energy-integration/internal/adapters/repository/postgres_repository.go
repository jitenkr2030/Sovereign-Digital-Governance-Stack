package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/energy-integration/energy/internal/core/domain"
	"github.com/energy-integration/energy/internal/core/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresEnergyRepository implements all energy-related repositories using PostgreSQL.
type PostgresEnergyRepository struct {
	pool            *pgxpool.Pool
	profileTable    string
	footprintTable  string
	certificateTable string
}

// NewPostgresEnergyRepository creates a new PostgreSQL-backed energy repository.
func NewPostgresEnergyRepository(pool *pgxpool.Pool) *PostgresEnergyRepository {
	return &PostgresEnergyRepository{
		pool:             pool,
		profileTable:     "energy_profiles",
		footprintTable:   "transaction_footprints",
		certificateTable: "offset_certificates",
	}
}

// ==================== Energy Profile Repository ====================

// Create creates a new energy profile.
func (r *PostgresEnergyRepository) Create(ctx context.Context, profile *domain.EnergyProfile) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, chain_id, chain_name, consensus_type, avg_kwh_per_tx,
			base_carbon_grams, energy_source, carbon_intensity,
			node_locations, network_hash_rate, transactions_per_second,
			is_active, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`, r.profileTable)

	nodeLocationsJSON, _ := json.Marshal(profile.NodeLocations)
	metadataJSON, _ := json.Marshal(profile.Metadata)

	_, err := r.pool.Exec(ctx, query,
		profile.ID, profile.ChainID, profile.ChainName, string(profile.ConsensusType),
		profile.AvgKWhPerTx, profile.BaseCarbonGrams, string(profile.EnergySource),
		profile.CarbonIntensity, nodeLocationsJSON, profile.NetworkHashRate,
		profile.TransactionsPerSecond, profile.IsActive, metadataJSON,
		profile.CreatedAt, profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create energy profile: %w", err)
	}
	return nil
}

// GetByID retrieves an energy profile by ID.
func (r *PostgresEnergyRepository) GetByID(ctx context.Context, id string) (*domain.EnergyProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, chain_id, chain_name, consensus_type, avg_kwh_per_tx,
		       base_carbon_grams, energy_source, carbon_intensity,
		       node_locations, network_hash_rate, transactions_per_second,
		       is_active, metadata, created_at, updated_at
		FROM %s WHERE id = $1 AND is_active = true
	`, r.profileTable)

	profile := &domain.EnergyProfile{}
	var nodeLocationsJSON, metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&profile.ID, &profile.ChainID, &profile.ChainName, &profile.ConsensusType,
		&profile.AvgKWhPerTx, &profile.BaseCarbonGrams, &profile.EnergySource,
		&profile.CarbonIntensity, &nodeLocationsJSON, &profile.NetworkHashRate,
		&profile.TransactionsPerSecond, &profile.IsActive, &metadataJSON,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get energy profile: %w", err)
	}

	if len(nodeLocationsJSON) > 0 {
		json.Unmarshal(nodeLocationsJSON, &profile.NodeLocations)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &profile.Metadata)
	}

	return profile, nil
}

// GetByChainID retrieves an energy profile by blockchain chain ID.
func (r *PostgresEnergyRepository) GetByChainID(ctx context.Context, chainID string) (*domain.EnergyProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, chain_id, chain_name, consensus_type, avg_kwh_per_tx,
		       base_carbon_grams, energy_source, carbon_intensity,
		       node_locations, network_hash_rate, transactions_per_second,
		       is_active, metadata, created_at, updated_at
		FROM %s WHERE chain_id = $1 AND is_active = true
	`, r.profileTable)

	profile := &domain.EnergyProfile{}
	var nodeLocationsJSON, metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, chainID).Scan(
		&profile.ID, &profile.ChainID, &profile.ChainName, &profile.ConsensusType,
		&profile.AvgKWhPerTx, &profile.BaseCarbonGrams, &profile.EnergySource,
		&profile.CarbonIntensity, &nodeLocationsJSON, &profile.NetworkHashRate,
		&profile.TransactionsPerSecond, &profile.IsActive, &metadataJSON,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get energy profile: %w", err)
	}

	if len(nodeLocationsJSON) > 0 {
		json.Unmarshal(nodeLocationsJSON, &profile.NodeLocations)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &profile.Metadata)
	}

	return profile, nil
}

// Update updates an existing energy profile.
func (r *PostgresEnergyRepository) Update(ctx context.Context, profile *domain.EnergyProfile) error {
	query := fmt.Sprintf(`
		UPDATE %s SET
			chain_name = $1, consensus_type = $2, avg_kwh_per_tx = $3,
			base_carbon_grams = $4, energy_source = $5, carbon_intensity = $6,
			node_locations = $7, network_hash_rate = $8, transactions_per_second = $9,
			is_active = $10, metadata = $11, updated_at = $12
		WHERE id = $13
	`, r.profileTable)

	nodeLocationsJSON, _ := json.Marshal(profile.NodeLocations)
	metadataJSON, _ := json.Marshal(profile.Metadata)
	profile.UpdatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx, query,
		profile.ChainName, string(profile.ConsensusType), profile.AvgKWhPerTx,
		profile.BaseCarbonGrams, string(profile.EnergySource), profile.CarbonIntensity,
		nodeLocationsJSON, profile.NetworkHashRate, profile.TransactionsPerSecond,
		profile.IsActive, metadataJSON, profile.UpdatedAt, profile.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update energy profile: %w", err)
	}
	return nil
}

// List retrieves all active energy profiles.
func (r *PostgresEnergyRepository) List(ctx context.Context) ([]*domain.EnergyProfile, error) {
	query := fmt.Sprintf(`
		SELECT id, chain_id, chain_name, consensus_type, avg_kwh_per_tx,
		       base_carbon_grams, energy_source, carbon_intensity,
		       node_locations, network_hash_rate, transactions_per_second,
		       is_active, metadata, created_at, updated_at
		FROM %s WHERE is_active = true ORDER BY chain_name
	`, r.profileTable)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list energy profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*domain.EnergyProfile
	for rows.Next() {
		profile := &domain.EnergyProfile{}
		var nodeLocationsJSON, metadataJSON []byte

		err := rows.Scan(
			&profile.ID, &profile.ChainID, &profile.ChainName, &profile.ConsensusType,
			&profile.AvgKWhPerTx, &profile.BaseCarbonGrams, &profile.EnergySource,
			&profile.CarbonIntensity, &nodeLocationsJSON, &profile.NetworkHashRate,
			&profile.TransactionsPerSecond, &profile.IsActive, &metadataJSON,
			&profile.CreatedAt, &profile.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan energy profile: %w", err)
		}

		if len(nodeLocationsJSON) > 0 {
			json.Unmarshal(nodeLocationsJSON, &profile.NodeLocations)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &profile.Metadata)
		}

		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

// Delete soft-deletes an energy profile.
func (r *PostgresEnergyRepository) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`UPDATE %s SET is_active = false, updated_at = $1 WHERE id = $2`, r.profileTable)
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete energy profile: %w", err)
	}
	return nil
}

// ==================== Transaction Footprint Repository ====================

// Create creates a new transaction footprint.
func (r *PostgresEnergyRepository) Create(ctx context.Context, footprint *domain.TransactionFootprint) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, transaction_hash, chain_id, block_number,
			energy_value, energy_unit, energy_source,
			carbon_value, carbon_unit, factor_id,
			offset_status, certificate_id, timestamp, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`, r.footprintTable)

	metadataJSON, _ := json.Marshal(footprint.Metadata)

	_, err := r.pool.Exec(ctx, query,
		footprint.ID, footprint.TransactionHash, footprint.ChainID, footprint.BlockNumber,
		footprint.EnergyUsed.Value, string(footprint.EnergyUsed.Unit), string(footprint.EnergyUsed.Source),
		footprint.CarbonEmission.Value, string(footprint.CarbonEmission.Unit), footprint.CarbonEmission.FactorID,
		string(footprint.OffsetStatus), footprint.CertificateID, footprint.Timestamp, metadataJSON, footprint.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction footprint: %w", err)
	}
	return nil
}

// GetByID retrieves a transaction footprint by ID.
func (r *PostgresEnergyRepository) GetByID(ctx context.Context, id string) (*domain.TransactionFootprint, error) {
	query := fmt.Sprintf(`
		SELECT id, transaction_hash, chain_id, block_number,
		       energy_value, energy_unit, energy_source,
		       carbon_value, carbon_unit, factor_id,
		       offset_status, certificate_id, timestamp, metadata, created_at
		FROM %s WHERE id = $1
	`, r.footprintTable)

	return r.scanFootprint(r.pool.QueryRow(ctx, query, id))
}

// GetByTransactionHash retrieves footprint by transaction hash.
func (r *PostgresEnergyRepository) GetByTransactionHash(ctx context.Context, txHash string) (*domain.TransactionFootprint, error) {
	query := fmt.Sprintf(`
		SELECT id, transaction_hash, chain_id, block_number,
		       energy_value, energy_unit, energy_source,
		       carbon_value, carbon_unit, factor_id,
		       offset_status, certificate_id, timestamp, metadata, created_at
		FROM %s WHERE transaction_hash = $1
	`, r.footprintTable)

	return r.scanFootprint(r.pool.QueryRow(ctx, query, txHash))
}

// Update updates an existing transaction footprint.
func (r *PostgresEnergyRepository) Update(ctx context.Context, footprint *domain.TransactionFootprint) error {
	query := fmt.Sprintf(`
		UPDATE %s SET
			chain_id = $1, block_number = $2,
			energy_value = $3, energy_unit = $4, energy_source = $5,
			carbon_value = $6, carbon_unit = $7, factor_id = $8,
			offset_status = $9, certificate_id = $10, metadata = $11
		WHERE id = $12
	`, r.footprintTable)

	metadataJSON, _ := json.Marshal(footprint.Metadata)

	_, err := r.pool.Exec(ctx, query,
		footprint.ChainID, footprint.BlockNumber,
		footprint.EnergyUsed.Value, string(footprint.EnergyUsed.Unit), string(footprint.EnergyUsed.Source),
		footprint.CarbonEmission.Value, string(footprint.CarbonEmission.Unit), footprint.CarbonEmission.FactorID,
		string(footprint.OffsetStatus), footprint.CertificateID, metadataJSON, footprint.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction footprint: %w", err)
	}
	return nil
}

// UpdateOffsetStatus updates only the offset status.
func (r *PostgresEnergyRepository) UpdateOffsetStatus(ctx context.Context, id string, status domain.OffsetStatus, certificateID string) error {
	query := fmt.Sprintf(`
		UPDATE %s SET offset_status = $1, certificate_id = $2 WHERE id = $3
	`, r.footprintTable)

	_, err := r.pool.Exec(ctx, query, string(status), certificateID, id)
	if err != nil {
		return fmt.Errorf("failed to update offset status: %w", err)
	}
	return nil
}

// List retrieves footprints with optional filtering.
func (r *PostgresEnergyRepository) List(ctx context.Context, filter ports.FootprintFilter) ([]*domain.TransactionFootprint, error) {
	baseQuery := fmt.Sprintf(`
		SELECT id, transaction_hash, chain_id, block_number,
		       energy_value, energy_unit, energy_source,
		       carbon_value, carbon_unit, factor_id,
		       offset_status, certificate_id, timestamp, metadata, created_at
		FROM %s WHERE 1=1
	`, r.footprintTable)

	args := []interface{}{}
	argIndex := 1

	if filter.ChainID != "" {
		baseQuery += fmt.Sprintf(" AND chain_id = $%d", argIndex)
		args = append(args, filter.ChainID)
		argIndex++
	}

	if len(filter.OffsetStatus) > 0 {
		placeholders := make([]string, len(filter.OffsetStatus))
		for i, status := range filter.OffsetStatus {
			args = append(args, string(status))
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			argIndex++
		}
		baseQuery += fmt.Sprintf(" AND offset_status IN (%s)", join(placeholders, ","))
	}

	if filter.StartTime != nil {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	// Add pagination
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	} else {
		baseQuery += " LIMIT 100"
	}

	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list footprints: %w", err)
	}
	defer rows.Close()

	var footprints []*domain.TransactionFootprint
	for rows.Next() {
		footprint, err := r.scanFootprint(rows)
		if err != nil {
			return nil, err
		}
		footprints = append(footprints, footprint)
	}

	return footprints, rows.Err()
}

// GetSummary retrieves aggregated footprint summary.
func (r *PostgresEnergyRepository) GetSummary(ctx context.Context, startTime, endTime time.Time, chainID string) (*domain.CarbonFootprintSummary, error) {
	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_transactions,
			COALESCE(SUM(energy_value), 0) as total_energy_kwh,
			COALESCE(SUM(carbon_value), 0) as total_carbon_grams,
			COALESCE(SUM(CASE WHEN offset_status = 'full' THEN carbon_value ELSE 0 END), 0) as offset_carbon_grams
		FROM %s
		WHERE timestamp >= $1 AND timestamp <= $2
		%s
	`, r.footprintTable, getChainFilter(chainID))

	var summary domain.CarbonFootprintSummary
	summary.PeriodStart = startTime
	summary.PeriodEnd = endTime

	var totalTx, totalEnergy, totalCarbon, offsetCarbon float64

	err := r.pool.QueryRow(ctx, query, startTime, endTime).Scan(
		&totalTx, &totalEnergy, &totalCarbon, &offsetCarbon,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get footprint summary: %w", err)
	}

	summary.TotalTransactions = int64(totalTx)
	summary.TotalEnergyKWh = totalEnergy
	summary.TotalCarbonGrams = totalCarbon
	summary.OffsetCarbonGrams = offsetCarbon
	summary.NetCarbonGrams = totalCarbon - offsetCarbon
	if totalTx > 0 {
		summary.AverageCarbonPerTx = totalCarbon / float64(totalTx)
	}

	return &summary, nil
}

// Count returns the count of footprints matching the filter.
func (r *PostgresEnergyRepository) Count(ctx context.Context, filter ports.FootprintFilter) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE 1=1", r.footprintTable)

	args := []interface{}{}
	argIndex := 1

	if filter.ChainID != "" {
		query += fmt.Sprintf(" AND chain_id = $%d", argIndex)
		args = append(args, filter.ChainID)
	}

	var count int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count footprints: %w", err)
	}
	return count, nil
}

// Delete soft-deletes a transaction footprint.
func (r *PostgresEnergyRepository) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.footprintTable)
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete footprint: %w", err)
	}
	return nil
}

// scanFootprint scans a row into a TransactionFootprint.
func (r *PostgresEnergyRepository) scanFootprint(row pgx.Row) (*domain.TransactionFootprint, error) {
	footprint := &domain.TransactionFootprint{}
	var metadataJSON []byte
	var blockNumber *int64

	err := row.Scan(
		&footprint.ID, &footprint.TransactionHash, &footprint.ChainID, &blockNumber,
		&footprint.EnergyUsed.Value, &footprint.EnergyUsed.Unit, &footprint.EnergyUsed.Source,
		&footprint.CarbonEmission.Value, &footprint.CarbonEmission.Unit, &footprint.CarbonEmission.FactorID,
		&footprint.OffsetStatus, &footprint.CertificateID, &footprint.Timestamp, &metadataJSON, &footprint.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan footprint: %w", err)
	}

	if blockNumber != nil {
		footprint.BlockNumber = *blockNumber
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &footprint.Metadata)
	}

	return footprint, nil
}

// ==================== Offset Certificate Repository ====================

// Create creates a new offset certificate.
func (r *PostgresEnergyRepository) Create(ctx context.Context, cert *domain.OffsetCertificate) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, certificate_number, energy_source, energy_amount, energy_unit,
			carbon_value, carbon_unit, registry, issue_date, expiry_date,
			status, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`, r.certificateTable)

	metadataJSON, _ := json.Marshal(cert.Metadata)

	_, err := r.pool.Exec(ctx, query,
		cert.ID, cert.CertificateNumber, string(cert.EnergySource),
		cert.EnergyAmount, string(cert.EnergyUnit), cert.CarbonOffset.Value,
		string(cert.CarbonOffset.Unit), cert.Registry, cert.IssueDate,
		cert.ExpiryDate, cert.Status, metadataJSON, cert.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}
	return nil
}

// GetByID retrieves an offset certificate by ID.
func (r *PostgresEnergyRepository) GetByID(ctx context.Context, id string) (*domain.OffsetCertificate, error) {
	query := fmt.Sprintf(`
		SELECT id, certificate_number, energy_source, energy_amount, energy_unit,
		       carbon_value, carbon_unit, registry, issue_date, expiry_date,
		       status, metadata, created_at
		FROM %s WHERE id = $1
	`, r.certificateTable)

	return r.scanCertificate(r.pool.QueryRow(ctx, query, id))
}

// GetByCertificateNumber retrieves certificate by its number.
func (r *PostgresEnergyRepository) GetByCertificateNumber(ctx context.Context, number string) (*domain.OffsetCertificate, error) {
	query := fmt.Sprintf(`
		SELECT id, certificate_number, energy_source, energy_amount, energy_unit,
		       carbon_value, carbon_unit, registry, issue_date, expiry_date,
		       status, metadata, created_at
		FROM %s WHERE certificate_number = $1
	`, r.certificateTable)

	return r.scanCertificate(r.pool.QueryRow(ctx, query, number))
}

// Update updates an existing offset certificate.
func (r *PostgresEnergyRepository) Update(ctx context.Context, cert *domain.OffsetCertificate) error {
	query := fmt.Sprintf(`
		UPDATE %s SET
			status = $1, metadata = $2
		WHERE id = $3
	`, r.certificateTable)

	metadataJSON, _ := json.Marshal(cert.Metadata)

	_, err := r.pool.Exec(ctx, query, cert.Status, metadataJSON, cert.ID)
	if err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}
	return nil
}

// MarkAsRedeemed marks a certificate as redeemed.
func (r *PostgresEnergyRepository) MarkAsRedeemed(ctx context.Context, id string, footprintID string) error {
	query := fmt.Sprintf(`
		UPDATE %s SET status = 'redeemed', metadata = jsonb_set(
			COALESCE(metadata, '{}'::jsonb),
			'{redeemed_for}',
			$1::jsonb
		) WHERE id = $2 AND status = 'active'
	`, r.certificateTable)

	_, err := r.pool.Exec(ctx, query, footprintID, id)
	if err != nil {
		return fmt.Errorf("failed to mark certificate as redeemed: %w", err)
	}
	return nil
}

// List retrieves certificates with optional filtering.
func (r *PostgresEnergyRepository) List(ctx context.Context, filter ports.CertificateFilter) ([]*domain.OffsetCertificate, error) {
	baseQuery := fmt.Sprintf(`
		SELECT id, certificate_number, energy_source, energy_amount, energy_unit,
		       carbon_value, carbon_unit, registry, issue_date, expiry_date,
		       status, metadata, created_at
		FROM %s WHERE 1=1
	`, r.certificateTable)

	args := []interface{}{}
	argIndex := 1

	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			args = append(args, status)
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			argIndex++
		}
		baseQuery += fmt.Sprintf(" AND status IN (%s)", join(placeholders, ","))
	}

	if filter.Registry != "" {
		baseQuery += fmt.Sprintf(" AND registry = $%d", argIndex)
		args = append(args, filter.Registry)
	}

	// Add pagination
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates: %w", err)
	}
	defer rows.Close()

	var certificates []*domain.OffsetCertificate
	for rows.Next() {
		cert, err := r.scanCertificate(rows)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, cert)
	}

	return certificates, rows.Err()
}

// Count returns the count of certificates matching the filter.
func (r *PostgresEnergyRepository) Count(ctx context.Context, filter ports.CertificateFilter) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE 1=1", r.certificateTable)

	args := []interface{}{}
	argIndex := 1

	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			args = append(args, status)
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			argIndex++
		}
		query += fmt.Sprintf(" AND status IN (%s)", join(placeholders, ","))
	}

	if filter.Registry != "" {
		query += fmt.Sprintf(" AND registry = $%d", argIndex)
		args = append(args, filter.Registry)
	}

	var count int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count certificates: %w", err)
	}
	return count, nil
}

// scanCertificate scans a row into an OffsetCertificate.
func (r *PostgresEnergyRepository) scanCertificate(row pgx.Row) (*domain.OffsetCertificate, error) {
	cert := &domain.OffsetCertificate{}
	var metadataJSON []byte

	err := row.Scan(
		&cert.ID, &cert.CertificateNumber, &cert.EnergySource,
		&cert.EnergyAmount, &cert.EnergyUnit, &cert.CarbonOffset.Value,
		&cert.CarbonOffset.Unit, &cert.Registry, &cert.IssueDate,
		&cert.ExpiryDate, &cert.Status, &metadataJSON, &cert.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan certificate: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &cert.Metadata)
	}

	return cert, nil
}

// ==================== External Oracle Repository ====================

// GetCarbonIntensity retrieves current carbon intensity for a region (placeholder implementation).
func (r *PostgresEnergyRepository) GetCarbonIntensity(ctx context.Context, regionCode string) (float64, error) {
	// Default implementation returns global average
	return 0.5, nil
}

// VerifyCertificate verifies a certificate with the external registry (placeholder implementation).
func (r *PostgresEnergyRepository) VerifyCertificate(ctx context.Context, certificateNumber string) (*domain.OffsetCertificate, error) {
	// This would normally call an external API
	return nil, errors.New("external oracle not configured")
}

// GetEnergyMix retrieves the energy mix for a region (placeholder implementation).
func (r *PostgresEnergyRepository) GetEnergyMix(ctx context.Context, regionCode string) (map[domain.EnergySource]float64, error) {
	// Default implementation returns global average mix
	return map[domain.EnergySource]float64{
		domain.EnergySourceMixed: 1.0,
	}, nil
}

// ==================== Helper Functions ====================

func getChainFilter(chainID string) string {
	if chainID == "" {
		return ""
	}
	return fmt.Sprintf("AND chain_id = '%s'", chainID)
}

func join(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += sep + items[i]
	}
	return result
}
