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

// PostgresBlacklistRepository handles blacklist data access
type PostgresBlacklistRepository struct {
	db *sql.DB
}

// NewPostgresBlacklistRepository creates a new blacklist repository
func NewPostgresBlacklistRepository(cfg config.DatabaseConfig) (*PostgresBlacklistRepository, error) {
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

	return &PostgresBlacklistRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresBlacklistRepository) Close() error {
	return r.db.Close()
}

// Create creates a new blacklist entry
func (r *PostgresBlacklistRepository) Create(ctx context.Context, entry *models.BlacklistEntry) error {
	query := `
		INSERT INTO blacklist (
			id, address, address_hash, blockchain, reason, reason_details,
			source, risk_level, status, expires_at, added_by, added_by_name,
			approved_by, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = r.db.ExecContext(ctx, query,
		entry.ID, entry.Address, entry.AddressHash, entry.Blockchain, entry.Reason,
		entry.ReasonDetails, entry.Source, entry.RiskLevel, entry.Status, entry.ExpiresAt,
		entry.AddedBy, entry.AddedByName, entry.ApprovedBy, metadataJSON, entry.CreatedAt, entry.UpdatedAt,
	)

	return err
}

// GetByAddress retrieves a blacklist entry by address
func (r *PostgresBlacklistRepository) GetByAddress(ctx context.Context, address string, blockchain models.BlockchainType) (*models.BlacklistEntry, error) {
	query := `
		SELECT id, address, address_hash, blockchain, reason, reason_details,
			source, risk_level, status, expires_at, added_by, added_by_name,
			approved_by, removed_by, removal_reason, metadata, created_at, updated_at
		FROM blacklist WHERE address = $1 AND blockchain = $2 AND status = 'ACTIVE'
	`

	var entry models.BlacklistEntry
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, address, blockchain).Scan(
		&entry.ID, &entry.Address, &entry.AddressHash, &entry.Blockchain, &entry.Reason,
		&entry.ReasonDetails, &entry.Source, &entry.RiskLevel, &entry.Status, &entry.ExpiresAt,
		&entry.AddedBy, &entry.AddedByName, &entry.ApprovedBy, &entry.RemovedBy,
		&entry.RemovalReason, &metadata, &entry.CreatedAt, &entry.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadata, &entry.Metadata)
	return &entry, nil
}

// GetByID retrieves a blacklist entry by ID
func (r *PostgresBlacklistRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.BlacklistEntry, error) {
	query := `
		SELECT id, address, address_hash, blockchain, reason, reason_details,
			source, risk_level, status, expires_at, added_by, added_by_name,
			approved_by, removed_by, removal_reason, metadata, created_at, updated_at
		FROM blacklist WHERE id = $1
	`

	var entry models.BlacklistEntry
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID, &entry.Address, &entry.AddressHash, &entry.Blockchain, &entry.Reason,
		&entry.ReasonDetails, &entry.Source, &entry.RiskLevel, &entry.Status, &entry.ExpiresAt,
		&entry.AddedBy, &entry.AddedByName, &entry.ApprovedBy, &entry.RemovedBy,
		&entry.RemovalReason, &metadata, &entry.CreatedAt, &entry.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadata, &entry.Metadata)
	return &entry, nil
}

// Update updates a blacklist entry
func (r *PostgresBlacklistRepository) Update(ctx context.Context, entry *models.BlacklistEntry) error {
	query := `
		UPDATE blacklist SET
			status = $1, removed_by = $2, removal_reason = $3, updated_at = $4
		WHERE id = $5
	`

	entry.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		entry.Status, entry.RemovedBy, entry.RemovalReason, entry.UpdatedAt, entry.ID,
	)

	return err
}

// Remove removes an address from the blacklist (soft delete)
func (r *PostgresBlacklistRepository) Remove(ctx context.Context, id uuid.UUID, removedBy uuid.UUID, reason string) error {
	query := `
		UPDATE blacklist SET
			status = $1, removed_by = $2, removal_reason = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, models.BlacklistStatusRemoved, removedBy, reason, time.Now(), id)

	return err
}

// List retrieves blacklist entries with filters
func (r *PostgresBlacklistRepository) List(ctx context.Context, filter *models.BlacklistFilter, limit, offset int) ([]*models.BlacklistEntry, error) {
	query := `
		SELECT id, address, address_hash, blockchain, reason, reason_details,
			source, risk_level, status, expires_at, added_by, added_by_name,
			approved_by, removed_by, removal_reason, metadata, created_at, updated_at
		FROM blacklist WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if len(filter.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Statuses))
		argIndex++
	}

	if len(filter.RiskLevels) > 0 {
		query += fmt.Sprintf(" AND risk_level = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.RiskLevels))
		argIndex++
	}

	if len(filter.Sources) > 0 {
		query += fmt.Sprintf(" AND source = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Sources))
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.BlacklistEntry
	for rows.Next() {
		var entry models.BlacklistEntry
		var metadata []byte

		err := rows.Scan(
			&entry.ID, &entry.Address, &entry.AddressHash, &entry.Blockchain, &entry.Reason,
			&entry.ReasonDetails, &entry.Source, &entry.RiskLevel, &entry.Status, &entry.ExpiresAt,
			&entry.AddedBy, &entry.AddedByName, &entry.ApprovedBy, &entry.RemovedBy,
			&entry.RemovalReason, &metadata, &entry.CreatedAt, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(metadata, &entry.Metadata)
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// Count counts blacklist entries
func (r *PostgresBlacklistRepository) Count(ctx context.Context, filter *models.BlacklistFilter) (int, error) {
	query := `SELECT COUNT(*) FROM blacklist WHERE status = 'ACTIVE'`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}

// IsBlacklisted checks if an address is blacklisted
func (r *PostgresBlacklistRepository) IsBlacklisted(ctx context.Context, address string, blockchain models.BlockchainType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM blacklist
			WHERE address = $1 AND blockchain = $2 AND status = 'ACTIVE'
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, address, blockchain).Scan(&exists)

	return exists, err
}

// BatchAdd adds multiple addresses to the blacklist
func (r *PostgresBlacklistRepository) BatchAdd(ctx context.Context, entries []*models.BlacklistEntry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, entry := range entries {
		if err := r.Create(ctx, entry); err != nil {
			return fmt.Errorf("failed to add entry: %w", err)
		}
	}

	return tx.Commit()
}
