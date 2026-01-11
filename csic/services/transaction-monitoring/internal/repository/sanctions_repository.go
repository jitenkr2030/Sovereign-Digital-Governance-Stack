package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/csic/monitoring/internal/core/domain"
	"go.uber.org/zap"
)

// SanctionsRepository implements ports.SanctionsRepository for PostgreSQL
type SanctionsRepository struct {
	db     *sql.DB
	logger *zap.Logger
	table  string
}

// NewSanctionsRepository creates a new sanctions repository
func NewSanctionsRepository(db *sql.DB, logger *zap.Logger) *SanctionsRepository {
	return &SanctionsRepository{
		db:     db,
		logger: logger,
		table:  "monitoring_sanctioned_addresses",
	}
}

// Create inserts a new sanctioned address
func (r *SanctionsRepository) Create(ctx context.Context, sanction *domain.SanctionedAddress) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (address, chain, source_list, reason, entity_name, entity_type, program, added_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (address, chain, source_list) DO NOTHING
	`, r.table)

	_, err := r.db.ExecContext(ctx, query,
		sanction.Address, sanction.Chain, sanction.SourceList, sanction.Reason,
		sanction.EntityName, sanction.EntityType, sanction.Program,
		sanction.AddedAt, sanction.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert sanctioned address: %w", err)
	}

	return nil
}

// CreateBatch inserts multiple sanctioned addresses
func (r *SanctionsRepository) CreateBatch(ctx context.Context, sanctions []*domain.SanctionedAddress) error {
	if len(sanctions) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (address, chain, source_list, reason, entity_name, entity_type, program, added_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (address, chain, source_list) DO NOTHING
	`, r.table))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, sanction := range sanctions {
		_, err := stmt.ExecContext(ctx,
			sanction.Address, sanction.Chain, sanction.SourceList, sanction.Reason,
			sanction.EntityName, sanction.EntityType, sanction.Program,
			sanction.AddedAt, sanction.ExpiresAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert sanctioned address: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a sanctioned address by ID
func (r *SanctionsRepository) GetByID(ctx context.Context, id string) (*domain.SanctionedAddress, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, r.table)
	return r.scanSanction(r.db.QueryRowContext(ctx, query, id))
}

// GetByAddress retrieves all sanctions for a specific address
func (r *SanctionsRepository) GetByAddress(ctx context.Context, address, chain string) ([]*domain.SanctionedAddress, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE address = $1 AND chain = $2
		AND (expires_at IS NULL OR expires_at > NOW())
	`, r.table)

	rows, err := r.db.QueryContext(ctx, query, address, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get sanctions by address: %w", err)
	}
	defer rows.Close()

	sanctions := make([]*domain.SanctionedAddress, 0)
	for rows.Next() {
		sanction, err := r.scanSanctionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sanction: %w", err)
		}
		sanctions = append(sanctions, sanction)
	}

	return sanctions, nil
}

// Exists checks if an address is in the sanctions list
func (r *SanctionsRepository) Exists(ctx context.Context, address, chain string) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM %s
			WHERE address = $1 AND chain = $2
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`, r.table)

	var exists bool
	err := r.db.QueryRowContext(ctx, query, address, chain).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check sanctions: %w", err)
	}

	return exists, nil
}

// Delete removes a sanctioned address
func (r *SanctionsRepository) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.table)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List retrieves all sanctioned addresses with pagination
func (r *SanctionsRepository) List(ctx context.Context, page, pageSize int) ([]*domain.SanctionedAddress, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE expires_at IS NULL OR expires_at > NOW()`, r.table)
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count sanctions: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE expires_at IS NULL OR expires_at > NOW()
		ORDER BY added_at DESC
		LIMIT $1 OFFSET $2
	`, r.table)

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sanctions: %w", err)
	}
	defer rows.Close()

	sanctions := make([]*domain.SanctionedAddress, 0)
	for rows.Next() {
		sanction, err := r.scanSanctionRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan sanction: %w", err)
		}
		sanctions = append(sanctions, sanction)
	}

	return sanctions, total, nil
}

// Search searches for sanctions by address or entity name
func (r *SanctionsRepository) Search(ctx context.Context, query string) ([]*domain.SanctionedAddress, error) {
	searchQuery := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE (address ILIKE $1 OR entity_name ILIKE $1 OR reason ILIKE $1)
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY added_at DESC
		LIMIT 100
	`, r.table)

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search sanctions: %w", err)
	}
	defer rows.Close()

	sanctions := make([]*domain.SanctionedAddress, 0)
	for rows.Next() {
		sanction, err := r.scanSanctionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sanction: %w", err)
		}
		sanctions = append(sanctions, sanction)
	}

	return sanctions, nil
}

// Helper function to scan a single sanction row
func (r *SanctionsRepository) scanSanctionRow(rows *sql.Rows) (*domain.SanctionedAddress, error) {
	var sanction domain.SanctionedAddress
	var expiresAt sql.NullTime

	err := rows.Scan(
		&sanction.ID, &sanction.Address, &sanction.Chain, &sanction.SourceList,
		&sanction.Reason, &sanction.EntityName, &sanction.EntityType, &sanction.Program,
		&sanction.AddedAt, &expiresAt,
	)

	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		sanction.ExpiresAt = &expiresAt.Time
	}

	return &sanction, nil
}

// Helper function to scan a single sanction from a row
func (r *SanctionsRepository) scanSanction(row *sql.Row) (*domain.SanctionedAddress, error) {
	rows, err := row.(*sql.Row).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return r.scanSanctionRow(rows)
}
