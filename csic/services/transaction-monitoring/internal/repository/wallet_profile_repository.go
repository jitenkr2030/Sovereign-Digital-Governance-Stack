package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/csic/monitoring/internal/core/domain"
	"go.uber.org/zap"
)

// WalletProfileRepository implements ports.WalletProfileRepository for PostgreSQL
type WalletProfileRepository struct {
	db     *sql.DB
	logger *zap.Logger
	table  string
}

// NewWalletProfileRepository creates a new wallet profile repository
func NewWalletProfileRepository(db *sql.DB, logger *zap.Logger) *WalletProfileRepository {
	return &WalletProfileRepository{
		db:     db,
		logger: logger,
		table:  "monitoring_wallet_profiles",
	}
}

// Create inserts a new wallet profile
func (r *WalletProfileRepository) Create(ctx context.Context, profile *domain.WalletProfile) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			address, chain, first_seen, last_seen, tx_count, total_volume_usd,
			avg_tx_value_usd, connected_tags, risk_indicators, wallet_age_hours,
			is_contract, contract_type, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, r.table)

	connectedTagsJSON, _ := json.Marshal(profile.ConnectedTags)
	riskIndicatorsJSON, _ := json.Marshal(profile.RiskIndicators)

	_, err := r.db.ExecContext(ctx, query,
		profile.Address, profile.Chain, profile.FirstSeen, profile.LastSeen,
		profile.TxCount, profile.TotalVolumeUSD, profile.AvgTxValueUSD,
		connectedTagsJSON, riskIndicatorsJSON, profile.WalletAgeHours,
		profile.IsContract, profile.ContractType, profile.CreatedAt, profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert wallet profile: %w", err)
	}

	return nil
}

// Update updates an existing wallet profile
func (r *WalletProfileRepository) Update(ctx context.Context, profile *domain.WalletProfile) error {
	query := fmt.Sprintf(`
		UPDATE %s SET
			last_seen = $1, tx_count = $2, total_volume_usd = $3, avg_tx_value_usd = $4,
			connected_tags = $5, risk_indicators = $6, wallet_age_hours = $7,
			is_contract = $8, contract_type = $9, updated_at = $10
		WHERE id = $11
	`, r.table)

	connectedTagsJSON, _ := json.Marshal(profile.ConnectedTags)
	riskIndicatorsJSON, _ := json.Marshal(profile.RiskIndicators)

	_, err := r.db.ExecContext(ctx, query,
		profile.LastSeen, profile.TxCount, profile.TotalVolumeUSD, profile.AvgTxValueUSD,
		connectedTagsJSON, riskIndicatorsJSON, profile.WalletAgeHours,
		profile.IsContract, profile.ContractType, profile.UpdatedAt, profile.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update wallet profile: %w", err)
	}

	return nil
}

// GetByID retrieves a wallet profile by ID
func (r *WalletProfileRepository) GetByID(ctx context.Context, id string) (*domain.WalletProfile, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, r.table)
	return r.scanWalletProfile(r.db.QueryRowContext(ctx, query, id))
}

// GetByAddress retrieves a wallet profile by address and chain
func (r *WalletProfileRepository) GetByAddress(ctx context.Context, address, chain string) (*domain.WalletProfile, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE address = $1 AND chain = $2`, r.table)
	return r.scanWalletProfile(r.db.QueryRowContext(ctx, query, address, chain))
}

// Upsert inserts or updates a wallet profile
func (r *WalletProfileRepository) Upsert(ctx context.Context, profile *domain.WalletProfile) error {
	// First try to get existing profile
	existing, err := r.GetByAddress(ctx, profile.Address, profile.Chain)
	if err != nil {
		return fmt.Errorf("failed to check existing profile: %w", err)
	}

	connectedTagsJSON, _ := json.Marshal(profile.ConnectedTags)
	riskIndicatorsJSON, _ := json.Marshal(profile.RiskIndicators)

	if existing != nil {
		// Update existing
		profile.ID = existing.ID
		profile.CreatedAt = existing.CreatedAt
		profile.UpdatedAt = profile.UpdatedAt

		query := fmt.Sprintf(`
			UPDATE %s SET
				last_seen = $1, tx_count = $2, total_volume_usd = $3, avg_tx_value_usd = $4,
				connected_tags = $5, risk_indicators = $6, wallet_age_hours = $7,
				is_contract = $8, contract_type = $9, updated_at = $10
			WHERE id = $11
		`, r.table)

		_, err := r.db.ExecContext(ctx, query,
			profile.LastSeen, profile.TxCount, profile.TotalVolumeUSD, profile.AvgTxValueUSD,
			connectedTagsJSON, riskIndicatorsJSON, profile.WalletAgeHours,
			profile.IsContract, profile.ContractType, profile.UpdatedAt, profile.ID,
		)

		return err
	}

	// Insert new
	profile.CreatedAt = profile.UpdatedAt
	return r.Create(ctx, profile)
}

// GetHighRisk retrieves wallet profiles with high risk indicators
func (r *WalletProfileRepository) GetHighRisk(ctx context.Context, limit int) ([]*domain.WalletProfile, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE jsonb_array_length(risk_indicators) > 0
		ORDER BY created_at DESC
		LIMIT $1
	`, r.table)

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get high risk wallets: %w", err)
	}
	defer rows.Close()

	profiles := make([]*domain.WalletProfile, 0)
	for rows.Next() {
		profile, err := r.scanWalletProfileRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// Search searches for wallet profiles by address
func (r *WalletProfileRepository) Search(ctx context.Context, query string, page, pageSize int) ([]*domain.WalletProfile, int64, error) {
	offset := (page - 1) * pageSize
	searchPattern := "%" + strings.ToLower(query) + "%"

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s
		WHERE LOWER(address) LIKE $1
	`, r.table)
	if err := r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count wallets: %w", err)
	}

	// Get paginated results
	searchQuery := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE LOWER(address) LIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, r.table)

	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search wallets: %w", err)
	}
	defer rows.Close()

	profiles := make([]*domain.WalletProfile, 0)
	for rows.Next() {
		profile, err := r.scanWalletProfileRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan wallet profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	return profiles, total, nil
}

// Helper function to scan a single wallet profile row
func (r *WalletProfileRepository) scanWalletProfileRow(rows *sql.Rows) (*domain.WalletProfile, error) {
	var profile domain.WalletProfile
	var connectedTagsJSON, riskIndicatorsJSON []byte
	var firstSeen, lastSeen sql.NullTime
	var contractType sql.NullString

	err := rows.Scan(
		&profile.ID, &profile.Address, &profile.Chain, &firstSeen, &lastSeen,
		&profile.TxCount, &profile.TotalVolumeUSD, &profile.AvgTxValueUSD,
		&connectedTagsJSON, &riskIndicatorsJSON, &profile.WalletAgeHours,
		&profile.IsContract, &contractType, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if firstSeen.Valid {
		profile.FirstSeen = &firstSeen.Time
	}
	if lastSeen.Valid {
		profile.LastSeen = &lastSeen.Time
	}
	if contractType.Valid {
		profile.ContractType = &contractType.String
	}

	// Parse JSON fields
	if len(connectedTagsJSON) > 0 {
		json.Unmarshal(connectedTagsJSON, &profile.ConnectedTags)
	}
	if len(riskIndicatorsJSON) > 0 {
		json.Unmarshal(riskIndicatorsJSON, &profile.RiskIndicators)
	}

	return &profile, nil
}

// Helper function to scan a single wallet profile from a row
func (r *WalletProfileRepository) scanWalletProfile(row *sql.Row) (*domain.WalletProfile, error) {
	rows, err := row.(*sql.Row).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return r.scanWalletProfileRow(rows)
}
