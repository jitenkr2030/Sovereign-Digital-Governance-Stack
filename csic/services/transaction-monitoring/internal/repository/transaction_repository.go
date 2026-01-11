package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
	"github.com/csic/monitoring/internal/core/ports"
	"go.uber.org/zap"
)

// TransactionRepository implements ports.TransactionRepository for PostgreSQL
type TransactionRepository struct {
	db     *sql.DB
	logger *zap.Logger
	table  string
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB, logger *zap.Logger) *TransactionRepository {
	return &TransactionRepository{
		db:     db,
		logger: logger,
		table:  "monitoring_transactions",
	}
}

// Create inserts a new transaction into the database
func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, tx_hash, chain, block_number, from_address, to_address, token_address,
			amount, amount_usd, gas_used, gas_price, gas_fee_usd, nonce, tx_timestamp,
			risk_score, risk_factors, flagged, flag_reason, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`, r.table)

	riskFactorsJSON, _ := json.Marshal(tx.RiskFactors)
	metadataJSON, _ := json.Marshal(tx.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.TxHash, tx.Chain, tx.BlockNumber, tx.FromAddress, tx.ToAddress,
		tx.TokenAddress, tx.Amount, tx.AmountUSD, tx.GasUsed, tx.GasPrice, tx.GasFeeUSD,
		tx.Nonce, tx.TxTimestamp, tx.RiskScore, riskFactorsJSON, tx.Flagged, tx.FlagReason,
		metadataJSON, tx.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a transaction by its ID
func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, r.table)
	return r.scanTransaction(r.db.QueryRowContext(ctx, query, id))
}

// GetByHash retrieves a transaction by its hash
func (r *TransactionRepository) GetByHash(ctx context.Context, txHash string) (*domain.Transaction, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE tx_hash = $1`, r.table)
	return r.scanTransaction(r.db.QueryRowContext(ctx, query, txHash))
}

// Update updates an existing transaction
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	query := fmt.Sprintf(`
		UPDATE %s SET
			risk_score = $1, risk_factors = $2, flagged = $3, flag_reason = $4,
			reviewed_at = $5, reviewed_by = $6
		WHERE id = $7
	`, r.table)

	riskFactorsJSON, _ := json.Marshal(tx.RiskFactors)

	_, err := r.db.ExecContext(ctx, query,
		tx.RiskScore, riskFactorsJSON, tx.Flagged, tx.FlagReason,
		tx.ReviewedAt, tx.ReviewedBy, tx.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// Delete removes a transaction (soft delete in production)
func (r *TransactionRepository) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.table)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List retrieves transactions based on filter criteria
func (r *TransactionRepository) List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, int64, error) {
	baseQuery := fmt.Sprintf(`FROM %s WHERE 1=1`, r.table)
	args := make([]interface{}, 0)
	argIndex := 1

	// Build WHERE clause
	if filter.Chain != "" {
		baseQuery += fmt.Sprintf(` AND chain = $%d`, argIndex)
		args = append(args, filter.Chain)
		argIndex++
	}

	if filter.FromAddress != "" {
		baseQuery += fmt.Sprintf(` AND from_address = $%d`, argIndex)
		args = append(args, filter.FromAddress)
		argIndex++
	}

	if filter.ToAddress != "" {
		baseQuery += fmt.Sprintf(` AND to_address = $%d`, argIndex)
		args = append(args, filter.ToAddress)
		argIndex++
	}

	if filter.MinAmountUSD > 0 {
		baseQuery += fmt.Sprintf(` AND amount_usd >= $%d`, argIndex)
		args = append(args, filter.MinAmountUSD)
		argIndex++
	}

	if filter.MaxAmountUSD > 0 {
		baseQuery += fmt.Sprintf(` AND amount_usd <= $%d`, argIndex)
		args = append(args, filter.MaxAmountUSD)
		argIndex++
	}

	if filter.Flagged != nil {
		baseQuery += fmt.Sprintf(` AND flagged = $%d`, argIndex)
		args = append(args, *filter.Flagged)
		argIndex++
	}

	if filter.MinRiskScore > 0 {
		baseQuery += fmt.Sprintf(` AND risk_score >= $%d`, argIndex)
		args = append(args, filter.MinRiskScore)
		argIndex++
	}

	if filter.StartTime != nil {
		baseQuery += fmt.Sprintf(` AND tx_timestamp >= $%d`, argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		baseQuery += fmt.Sprintf(` AND tx_timestamp <= $%d`, argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) %s`, baseQuery)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get paginated results
	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT * %s
		ORDER BY tx_timestamp DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		tx, err := r.scanTransactionRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// GetByAddress retrieves transactions for a specific address
func (r *TransactionRepository) GetByAddress(ctx context.Context, address, chain string, limit int) ([]*domain.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE (from_address = $1 OR to_address = $1) AND chain = $2
		ORDER BY tx_timestamp DESC
		LIMIT $3
	`, r.table)

	rows, err := r.db.QueryContext(ctx, query, address, chain, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by address: %w", err)
	}
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		tx, err := r.scanTransactionRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetFlagged retrieves flagged transactions with pagination
func (r *TransactionRepository) GetFlagged(ctx context.Context, page, pageSize int) ([]*domain.Transaction, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE flagged = true`, r.table)
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count flagged transactions: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE flagged = true
		ORDER BY tx_timestamp DESC
		LIMIT $1 OFFSET $2
	`, r.table)

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get flagged transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		tx, err := r.scanTransactionRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// Helper function to scan a single transaction row
func (r *TransactionRepository) scanTransactionRow(rows *sql.Rows) (*domain.Transaction, error) {
	var tx domain.Transaction
	var riskFactorsJSON, metadataJSON []byte
	var toAddress, flagReason sql.NullString
	var blockNumber sql.NullInt64
	var gasUsed sql.NullInt64
	var gasPrice, gasFeeUSD sql.NullFloat64
	var nonce sql.NullInt64
	var reviewedAt sql.NullTime
	var reviewedBy sql.NullString

	err := rows.Scan(
		&tx.ID, &tx.TxHash, &tx.Chain, &blockNumber, &tx.FromAddress, &toAddress,
		&tx.TokenAddress, &tx.Amount, &tx.AmountUSD, &gasUsed, &gasPrice, &gasFeeUSD,
		&nonce, &tx.TxTimestamp, &tx.RiskScore, &riskFactorsJSON, &tx.Flagged, &flagReason,
		&metadataJSON, &reviewedAt, &reviewedBy, &tx.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse nullable fields
	if blockNumber.Valid {
		tx.BlockNumber = &blockNumber.Int64
	}
	if toAddress.Valid {
		tx.ToAddress = &toAddress.String
	}
	if gasUsed.Valid {
		tx.GasUsed = &gasUsed.Int64
	}
	if gasPrice.Valid {
		tx.GasPrice = &gasPrice.Float64
	}
	if gasFeeUSD.Valid {
		tx.GasFeeUSD = &gasFeeUSD.Float64
	}
	if nonce.Valid {
		tx.Nonce = &nonce.Int
	}
	if flagReason.Valid {
		tx.FlagReason = &flagReason.String
	}
	if reviewedAt.Valid {
		tx.ReviewedAt = &reviewedAt.Time
	}
	if reviewedBy.Valid {
		tx.ReviewedBy = &reviewedBy.String
	}

	// Parse JSON fields
	if len(riskFactorsJSON) > 0 {
		json.Unmarshal(riskFactorsJSON, &tx.RiskFactors)
	}
	if len(metadataJSON) > 0 {
		tx.Metadata = make(map[string]interface{})
		json.Unmarshal(metadataJSON, &tx.Metadata)
	}

	return &tx, nil
}

// Helper function to scan a single transaction from a row
func (r *TransactionRepository) scanTransaction(row *sql.Row) (*domain.Transaction, error) {
	rows, err := row.(*sql.Row).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	return r.scanTransactionRow(rows)
}

// Ensure Repository implements the interface
var _ ports.TransactionRepository = (*TransactionRepository)(nil)
