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
	_ "github.com/lib/pq"
)

// PostgresWhitelistRepository handles whitelist data access
type PostgresWhitelistRepository struct {
	db *sql.DB
}

// NewPostgresWhitelistRepository creates a new whitelist repository
func NewPostgresWhitelistRepository(cfg config.DatabaseConfig) (*PostgresWhitelistRepository, error) {
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

	return &PostgresWhitelistRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresWhitelistRepository) Close() error {
	return r.db.Close()
}

// Create creates a new whitelist entry
func (r *PostgresWhitelistRepository) Create(ctx context.Context, entry *models.WhitelistEntry) error {
	query := `
		INSERT INTO whitelist (
			id, address, address_hash, blockchain, label, description,
			entity_type, entity_id, entity_name, status, expires_at,
			added_by, added_by_name, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(entry.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		entry.ID, entry.Address, entry.AddressHash, entry.Blockchain, entry.Label,
		entry.Description, entry.EntityType, entry.EntityID, entry.EntityName,
		entry.Status, entry.ExpiresAt, entry.AddedBy, entry.AddedByName,
		metadataJSON, entry.CreatedAt, entry.UpdatedAt,
	)

	return err
}

// GetByAddress retrieves a whitelist entry by address
func (r *PostgresWhitelistRepository) GetByAddress(ctx context.Context, address string, blockchain models.BlockchainType) (*models.WhitelistEntry, error) {
	query := `
		SELECT id, address, address_hash, blockchain, label, description,
			entity_type, entity_id, entity_name, status, expires_at,
			added_by, added_by_name, metadata, created_at, updated_at
		FROM whitelist WHERE address = $1 AND blockchain = $2 AND status = 'ACTIVE'
	`

	var entry models.WhitelistEntry
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, address, blockchain).Scan(
		&entry.ID, &entry.Address, &entry.AddressHash, &entry.Blockchain, &entry.Label,
		&entry.Description, &entry.EntityType, &entry.EntityID, &entry.EntityName,
		&entry.Status, &entry.ExpiresAt, &entry.AddedBy, &entry.AddedByName,
		&metadata, &entry.CreatedAt, &entry.UpdatedAt,
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

// IsWhitelisted checks if an address is whitelisted
func (r *PostgresWhitelistRepository) IsWhitelisted(ctx context.Context, address string, blockchain models.BlockchainType) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM whitelist
			WHERE address = $1 AND blockchain = $2 AND status = 'ACTIVE'
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, address, blockchain).Scan(&exists)

	return exists, err
}

// List retrieves whitelist entries
func (r *PostgresWhitelistRepository) List(ctx context.Context, limit, offset int) ([]*models.WhitelistEntry, error) {
	query := `
		SELECT id, address, address_hash, blockchain, label, description,
			entity_type, entity_id, entity_name, status, expires_at,
			added_by, added_by_name, metadata, created_at, updated_at
		FROM whitelist WHERE status = 'ACTIVE'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.WhitelistEntry
	for rows.Next() {
		var entry models.WhitelistEntry
		var metadata []byte

		err := rows.Scan(
			&entry.ID, &entry.Address, &entry.AddressHash, &entry.Blockchain, &entry.Label,
			&entry.Description, &entry.EntityType, &entry.EntityID, &entry.EntityName,
			&entry.Status, &entry.ExpiresAt, &entry.AddedBy, &entry.AddedByName,
			&metadata, &entry.CreatedAt, &entry.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(metadata, &entry.Metadata)
		entries = append(entries, &entry)
	}

	return entries, rows.Err()
}

// Count counts whitelist entries
func (r *PostgresWhitelistRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM whitelist WHERE status = 'ACTIVE'`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}
