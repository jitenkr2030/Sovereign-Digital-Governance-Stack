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
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// PostgresWalletRepository handles wallet data access
type PostgresWalletRepository struct {
	db *sql.DB
}

// NewPostgresWalletRepository creates a new wallet repository
func NewPostgresWalletRepository(cfg config.DatabaseConfig) (*PostgresWalletRepository, error) {
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

	return &PostgresWalletRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresWalletRepository) Close() error {
	return r.db.Close()
}

// Create creates a new wallet
func (r *PostgresWalletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (
			id, wallet_id, type, status, blockchain, address, address_checksum,
			label, description, exchange_id, exchange_name, owner_entity_id,
			owner_entity_name, signers_required, signers_total, threshold,
			total_balance, balance_currency, compliance_score, is_whitelisted,
			is_blacklisted, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24
		)
	`

	wallet.ID = uuid.New()
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(wallet.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = r.db.ExecContext(ctx, query,
		wallet.ID, wallet.WalletID, wallet.Type, wallet.Status, wallet.Blockchain,
		wallet.Address, wallet.AddressChecksum, wallet.Label, wallet.Description,
		wallet.ExchangeID, wallet.ExchangeName, wallet.OwnerEntityID, wallet.OwnerEntityName,
		wallet.SignersRequired, wallet.SignersTotal, wallet.Threshold, wallet.TotalBalance,
		wallet.BalanceCurrency, wallet.ComplianceScore, wallet.IsWhitelisted, wallet.IsBlacklisted,
		metadataJSON, wallet.CreatedAt, wallet.UpdatedAt,
	)

	return err
}

// GetByID retrieves a wallet by ID
func (r *PostgresWalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	query := `
		SELECT id, wallet_id, type, status, blockchain, address, address_checksum,
			label, description, exchange_id, exchange_name, owner_entity_id, owner_entity_name,
			signers_required, signers_total, threshold, total_balance, balance_currency,
			last_activity_at, compliance_score, is_whitelisted, is_blacklisted, metadata,
			created_at, updated_at, revoked_at
		FROM wallets WHERE id = $1
	`

	var wallet models.Wallet
	var metadata []byte
	var revokedAt sql.NullTime
	var exchangeID, lastActivityAt interface{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wallet.ID, &wallet.WalletID, &wallet.Type, &wallet.Status, &wallet.Blockchain,
		&wallet.Address, &wallet.AddressChecksum, &wallet.Label, &wallet.Description,
		&exchangeID, &wallet.ExchangeName, &wallet.OwnerEntityID, &wallet.OwnerEntityName,
		&wallet.SignersRequired, &wallet.SignersTotal, &wallet.Threshold, &wallet.TotalBalance,
		&wallet.BalanceCurrency, &lastActivityAt, &wallet.ComplianceScore,
		&wallet.IsWhitelisted, &wallet.IsBlacklisted, &metadata, &wallet.CreatedAt,
		&wallet.UpdatedAt, &revokedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if exchangeID != nil {
		id := exchangeID.(uuid.UUID)
		wallet.ExchangeID = &id
	}
	if lastActivityAt != nil {
		t := lastActivityAt.(time.Time)
		wallet.LastActivityAt = &t
	}
	if revokedAt.Valid {
		wallet.RevokedAt = &revokedAt.Time
	}

	json.Unmarshal(metadata, &wallet.Metadata)

	return &wallet, nil
}

// GetByAddress retrieves a wallet by blockchain address
func (r *PostgresWalletRepository) GetByAddress(ctx context.Context, address string, blockchain models.BlockchainType) (*models.Wallet, error) {
	query := `
		SELECT id, wallet_id, type, status, blockchain, address, address_checksum,
			label, description, exchange_id, exchange_name, owner_entity_id, owner_entity_name,
			signers_required, signers_total, threshold, total_balance, balance_currency,
			last_activity_at, compliance_score, is_whitelisted, is_blacklisted, metadata,
			created_at, updated_at, revoked_at
		FROM wallets WHERE address = $1 AND blockchain = $2
	`

	var wallet models.Wallet
	var metadata []byte
	var revokedAt sql.NullTime
	var exchangeID, lastActivityAt interface{}

	err := r.db.QueryRowContext(ctx, query, address, blockchain).Scan(
		&wallet.ID, &wallet.WalletID, &wallet.Type, &wallet.Status, &wallet.Blockchain,
		&wallet.Address, &wallet.AddressChecksum, &wallet.Label, &wallet.Description,
		&exchangeID, &wallet.ExchangeName, &wallet.OwnerEntityID, &wallet.OwnerEntityName,
		&wallet.SignersRequired, &wallet.SignersTotal, &wallet.Threshold, &wallet.TotalBalance,
		&wallet.BalanceCurrency, &lastActivityAt, &wallet.ComplianceScore,
		&wallet.IsWhitelisted, &wallet.IsBlacklisted, &metadata, &wallet.CreatedAt,
		&wallet.UpdatedAt, &revokedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if exchangeID != nil {
		id := exchangeID.(uuid.UUID)
		wallet.ExchangeID = &id
	}
	if lastActivityAt != nil {
		t := lastActivityAt.(time.Time)
		wallet.LastActivityAt = &t
	}
	if revokedAt.Valid {
		wallet.RevokedAt = &revokedAt.Time
	}

	json.Unmarshal(metadata, &wallet.Metadata)

	return &wallet, nil
}

// Update updates a wallet
func (r *PostgresWalletRepository) Update(ctx context.Context, wallet *models.Wallet) error {
	query := `
		UPDATE wallets SET
			status = $1, label = $2, description = $3, signers_required = $4,
			signers_total = $5, threshold = $6, total_balance = $7, balance_currency = $8,
			last_activity_at = $9, compliance_score = $10, is_whitelisted = $11,
			is_blacklisted = $12, metadata = $13, updated_at = $14
		WHERE id = $15
	`

	wallet.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(wallet.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = r.db.ExecContext(ctx, query,
		wallet.Status, wallet.Label, wallet.Description, wallet.SignersRequired,
		wallet.SignersTotal, wallet.Threshold, wallet.TotalBalance, wallet.BalanceCurrency,
		wallet.LastActivityAt, wallet.ComplianceScore, wallet.IsWhitelisted, wallet.IsBlacklisted,
		metadataJSON, wallet.UpdatedAt, wallet.ID,
	)

	return err
}

// Revoke revokes a wallet
func (r *PostgresWalletRepository) Revoke(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE wallets SET
			status = $1, revoked_at = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, models.WalletStatusRevoked, now, now, id)

	return err
}

// List retrieves wallets with filters
func (r *PostgresWalletRepository) List(ctx context.Context, filter *models.WalletFilter, limit, offset int) ([]*models.Wallet, error) {
	query := `
		SELECT id, wallet_id, type, status, blockchain, address, address_checksum,
			label, description, exchange_id, exchange_name, owner_entity_id, owner_entity_name,
			signers_required, signers_total, threshold, total_balance, balance_currency,
			last_activity_at, compliance_score, is_whitelisted, is_blacklisted, metadata,
			created_at, updated_at, revoked_at
		FROM wallets WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if len(filter.Types) > 0 {
		query += fmt.Sprintf(" AND type = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Types))
		argIndex++
	}

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Statuses))
		argIndex++
	}

	if len(filter.Blockchains) > 0 {
		query += fmt.Sprintf(" AND blockchain = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Blockchains))
		argIndex++
	}

	if filter.IsBlacklisted != nil {
		query += fmt.Sprintf(" AND is_blacklisted = $%d", argIndex)
		args = append(args, *filter.IsBlacklisted)
		argIndex++
	}

	if filter.IsWhitelisted != nil {
		query += fmt.Sprintf(" AND is_whitelisted = $%d", argIndex)
		args = append(args, *filter.IsWhitelisted)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []*models.Wallet
	for rows.Next() {
		var wallet models.Wallet
		var metadata []byte
		var revokedAt sql.NullTime
		var exchangeID, lastActivityAt interface{}

		err := rows.Scan(
			&wallet.ID, &wallet.WalletID, &wallet.Type, &wallet.Status, &wallet.Blockchain,
			&wallet.Address, &wallet.AddressChecksum, &wallet.Label, &wallet.Description,
			&exchangeID, &wallet.ExchangeName, &wallet.OwnerEntityID, &wallet.OwnerEntityName,
			&wallet.SignersRequired, &wallet.SignersTotal, &wallet.Threshold, &wallet.TotalBalance,
			&wallet.BalanceCurrency, &lastActivityAt, &wallet.ComplianceScore,
			&wallet.IsWhitelisted, &wallet.IsBlacklisted, &metadata, &wallet.CreatedAt,
			&wallet.UpdatedAt, &revokedAt,
		)
		if err != nil {
			return nil, err
		}

		if exchangeID != nil {
			id := exchangeID.(uuid.UUID)
			wallet.ExchangeID = &id
		}
		if lastActivityAt != nil {
			t := lastActivityAt.(time.Time)
			wallet.LastActivityAt = &t
		}
		if revokedAt.Valid {
			wallet.RevokedAt = &revokedAt.Time
		}

		json.Unmarshal(metadata, &wallet.Metadata)
		wallets = append(wallets, &wallet)
	}

	return wallets, rows.Err()
}

// Count counts wallets with filters
func (r *PostgresWalletRepository) Count(ctx context.Context, filter *models.WalletFilter) (int, error) {
	query := `SELECT COUNT(*) FROM wallets WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if len(filter.Types) > 0 {
		query += fmt.Sprintf(" AND type = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Types))
		argIndex++
	}

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Statuses))
		argIndex++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)

	return count, err
}

// GetSummary retrieves wallet statistics summary
func (r *PostgresWalletRepository) GetSummary(ctx context.Context) (*models.WalletSummary, error) {
	summary := &models.WalletSummary{
		ByType:       make(map[string]int),
		ByBlockchain: make(map[string]int),
	}

	// Total and status counts
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'ACTIVE') as active,
			COUNT(*) FILTER (WHERE status = 'FROZEN') as frozen,
			COUNT(*) FILTER (WHERE status = 'SUSPENDED') as suspended
		FROM wallets
	`

	err := r.db.QueryRowContext(ctx, query).Scan(
		&summary.TotalWallets, &summary.ActiveWallets, &summary.FrozenWallets, &summary.SuspendedWallets,
	)
	if err != nil {
		return nil, err
	}

	// By type
	query = `SELECT type, COUNT(*) FROM wallets GROUP BY type`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var walletType string
		var count int
		if err := rows.Scan(&walletType, &count); err != nil {
			return nil, err
		}
		summary.ByType[walletType] = count
	}

	// By blockchain
	query = `SELECT blockchain, COUNT(*) FROM wallets GROUP BY blockchain`
	rows, err = r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var blockchain string
		var count int
		if err := rows.Scan(&blockchain, &count); err != nil {
			return nil, err
		}
		summary.ByBlockchain[blockchain] = count
	}

	// Blacklisted and whitelisted counts
	query = `
		SELECT
			COUNT(*) FILTER (WHERE is_blacklisted) as blacklisted,
			COUNT(*) FILTER (WHERE is_whitelisted) as whitelisted
		FROM wallets
	`
	err = r.db.QueryRowContext(ctx, query).Scan(&summary.BlacklistedAddresses, &summary.WhitelistedAddresses)
	if err != nil {
		return nil, err
	}

	summary.UpdatedAt = time.Now()
	return summary, nil
}
