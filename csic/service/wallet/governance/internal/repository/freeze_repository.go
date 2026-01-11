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

// PostgresWalletFreezeRepository handles wallet freeze data access
type PostgresWalletFreezeRepository struct {
	db *sql.DB
}

// NewPostgresWalletFreezeRepository creates a new wallet freeze repository
func NewPostgresWalletFreezeRepository(cfg config.DatabaseConfig) (*PostgresWalletFreezeRepository, error) {
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

	return &PostgresWalletFreezeRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresWalletFreezeRepository) Close() error {
	return r.db.Close()
}

// Create creates a new wallet freeze record
func (r *PostgresWalletFreezeRepository) Create(ctx context.Context, freeze *models.WalletFreeze) error {
	query := `
		INSERT INTO wallet_freezes (
			id, wallet_id, wallet_address, blockchain, reason, reason_details,
			status, freeze_level, legal_order_id, issued_by, issued_by_name,
			approved_by, expires_at, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	freeze.ID = uuid.New()
	freeze.CreatedAt = time.Now()
	freeze.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(freeze.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = r.db.ExecContext(ctx, query,
		freeze.ID, freeze.WalletID, freeze.WalletAddress, freeze.Blockchain, freeze.Reason,
		freeze.ReasonDetails, freeze.Status, freeze.FreezeLevel, freeze.LegalOrderID,
		freeze.IssuedBy, freeze.IssuedByName, freeze.ApprovedBy, freeze.ExpiresAt,
		metadataJSON, freeze.CreatedAt, freeze.UpdatedAt,
	)

	return err
}

// GetByID retrieves a freeze record by ID
func (r *PostgresWalletFreezeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WalletFreeze, error) {
	query := `
		SELECT id, wallet_id, wallet_address, blockchain, reason, reason_details,
			status, freeze_level, legal_order_id, issued_by, issued_by_name,
			approved_by, expires_at, released_at, release_reason, metadata,
			created_at, updated_at
		FROM wallet_freezes WHERE id = $1
	`

	var freeze models.WalletFreeze
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&freeze.ID, &freeze.WalletID, &freeze.WalletAddress, &freeze.Blockchain, &freeze.Reason,
		&freeze.ReasonDetails, &freeze.Status, &freeze.FreezeLevel, &freeze.LegalOrderID,
		&freeze.IssuedBy, &freeze.IssuedByName, &freeze.ApprovedBy, &freeze.ExpiresAt,
		&freeze.ReleasedAt, &freeze.ReleaseReason, &metadata, &freeze.CreatedAt, &freeze.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadata, &freeze.Metadata)
	return &freeze, nil
}

// GetActiveByWallet retrieves the active freeze for a wallet
func (r *PostgresWalletFreezeRepository) GetActiveByWallet(ctx context.Context, walletID uuid.UUID) (*models.WalletFreeze, error) {
	query := `
		SELECT id, wallet_id, wallet_address, blockchain, reason, reason_details,
			status, freeze_level, legal_order_id, issued_by, issued_by_name,
			approved_by, expires_at, released_at, release_reason, metadata,
			created_at, updated_at
		FROM wallet_freezes
		WHERE wallet_id = $1 AND status IN ('ACTIVE', 'PARTIAL')
		ORDER BY created_at DESC LIMIT 1
	`

	var freeze models.WalletFreeze
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, walletID).Scan(
		&freeze.ID, &freeze.WalletID, &freeze.WalletAddress, &freeze.Blockchain, &freeze.Reason,
		&freeze.ReasonDetails, &freeze.Status, &freeze.FreezeLevel, &freeze.LegalOrderID,
		&freeze.IssuedBy, &freeze.IssuedByName, &freeze.ApprovedBy, &freeze.ExpiresAt,
		&freeze.ReleasedAt, &freeze.ReleaseReason, &metadata, &freeze.CreatedAt, &freeze.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadata, &freeze.Metadata)
	return &freeze, nil
}

// Update updates a freeze record
func (r *PostgresWalletFreezeRepository) Update(ctx context.Context, freeze *models.WalletFreeze) error {
	query := `
		UPDATE wallet_freezes SET
			status = $1, released_at = $2, release_reason = $3, updated_at = $4
		WHERE id = $5
	`

	freeze.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		freeze.Status, freeze.ReleasedAt, freeze.ReleaseReason, freeze.UpdatedAt, freeze.ID,
	)

	return err
}

// Release releases a wallet freeze
func (r *PostgresWalletFreezeRepository) Release(ctx context.Context, id uuid.UUID, releasedBy uuid.UUID, reason string) error {
	query := `
		UPDATE wallet_freezes SET
			status = $1, released_at = $2, release_reason = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, models.FreezeStatusReleased, time.Now(), reason, time.Now(), id)

	return err
}

// List retrieves freeze records with filters
func (r *PostgresWalletFreezeRepository) List(ctx context.Context, filter *models.FreezeFilter, limit, offset int) ([]*models.WalletFreeze, error) {
	query := `
		SELECT id, wallet_id, wallet_address, blockchain, reason, reason_details,
			status, freeze_level, legal_order_id, issued_by, issued_by_name,
			approved_by, expires_at, released_at, release_reason, metadata,
			created_at, updated_at
		FROM wallet_freezes WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Statuses))
		argIndex++
	}

	if len(filter.Reasons) > 0 {
		query += fmt.Sprintf(" AND reason = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Reasons))
		argIndex++
	}

	if len(filter.WalletIDs) > 0 {
		query += fmt.Sprintf(" AND wallet_id = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.WalletIDs))
		argIndex++
	}

	if filter.ActiveOnly {
		query += " AND status IN ('ACTIVE', 'PARTIAL')"
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var freezes []*models.WalletFreeze
	for rows.Next() {
		var freeze models.WalletFreeze
		var metadata []byte

		err := rows.Scan(
			&freeze.ID, &freeze.WalletID, &freeze.WalletAddress, &freeze.Blockchain, &freeze.Reason,
			&freeze.ReasonDetails, &freeze.Status, &freeze.FreezeLevel, &freeze.LegalOrderID,
			&freeze.IssuedBy, &freeze.IssuedByName, &freeze.ApprovedBy, &freeze.ExpiresAt,
			&freeze.ReleasedAt, &freeze.ReleaseReason, &metadata, &freeze.CreatedAt, &freeze.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(metadata, &freeze.Metadata)
		freezes = append(frees, &freeze)
	}

	return freezes, rows.Err()
}

// GetActiveFreezes retrieves all active freezes
func (r *PostgresWalletFreezeRepository) GetActiveFreezes(ctx context.Context) ([]*models.WalletFreeze, error) {
	query := `
		SELECT id, wallet_id, wallet_address, blockchain, reason, reason_details,
			status, freeze_level, legal_order_id, issued_by, issued_by_name,
			approved_by, expires_at, released_at, release_reason, metadata,
			created_at, updated_at
		FROM wallet_freezes
		WHERE status IN ('ACTIVE', 'PARTIAL')
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var freezes []*models.WalletFreeze
	for rows.Next() {
		var freeze models.WalletFreeze
		var metadata []byte

		err := rows.Scan(
			&freeze.ID, &freeze.WalletID, &freeze.WalletAddress, &freeze.Blockchain, &freeze.Reason,
			&freeze.ReasonDetails, &freeze.Status, &freeze.FreezeLevel, &freeze.LegalOrderID,
			&freeze.IssuedBy, &freeze.IssuedByName, &freeze.ApprovedBy, &freeze.ExpiresAt,
			&freeze.ReleasedAt, &freeze.ReleaseReason, &metadata, &freeze.CreatedAt, &freeze.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(metadata, &freeze.Metadata)
		freezes = append(freezes, &freeze)
	}

	return freezes, rows.Err()
}

// Count counts active freezes
func (r *PostgresWalletFreezeRepository) Count(ctx context.Context, activeOnly bool) (int, error) {
	query := `SELECT COUNT(*) FROM wallet_freezes`
	if activeOnly {
		query += " WHERE status IN ('ACTIVE', 'PARTIAL')"
	}

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}

// GetExpiredFreezes retrieves freezes that have expired
func (r *PostgresWalletFreezeRepository) GetExpiredFreezes(ctx context.Context) ([]*models.WalletFreeze, error) {
	query := `
		SELECT id, wallet_id, wallet_address, blockchain, reason, reason_details,
			status, freeze_level, legal_order_id, issued_by, issued_by_name,
			approved_by, expires_at, released_at, release_reason, metadata,
			created_at, updated_at
		FROM wallet_freezes
		WHERE status IN ('ACTIVE', 'PARTIAL')
		AND expires_at IS NOT NULL
		AND expires_at <= NOW()
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var freezes []*models.WalletFreeze
	for rows.Next() {
		var freeze models.WalletFreeze
		var metadata []byte

		err := rows.Scan(
			&freeze.ID, &freeze.WalletID, &freeze.WalletAddress, &freeze.Blockchain, &freeze.Reason,
			&freeze.ReasonDetails, &freeze.Status, &freeze.FreezeLevel, &freeze.LegalOrderID,
			&freeze.IssuedBy, &freeze.IssuedByName, &freeze.ApprovedBy, &freeze.ExpiresAt,
			&freeze.ReleasedAt, &freeze.ReleaseReason, &metadata, &freeze.CreatedAt, &freeze.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(metadata, &freeze.Metadata)
		freezes = append(freezes, &freeze)
	}

	return freezes, rows.Err()
}
