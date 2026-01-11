package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/core/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Connection manages PostgreSQL connection
type Connection struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewConnection creates a new database connection
func NewConnection(logger *zap.Logger) (*Connection, error) {
	// In production, connection string would come from config
	connStr := "postgres://postgres:postgres@localhost:5432/csic_platform?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")
	return &Connection{
		pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.logger.Info("Database connection closed")
	}
}

// GetPool returns the underlying connection pool
func (c *Connection) GetPool() *pgxpool.Pool {
	return c.pool
}

// TransactionRepository implements ports.TransactionRepository
type TransactionRepository struct {
	conn   *Connection
	logger *zap.Logger
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(conn *Connection, logger *zap.Logger) *TransactionRepository {
	return &TransactionRepository{
		conn:   conn,
		logger: logger,
	}
}

// CreateTransaction creates a new transaction record
func (r *TransactionRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, tx_hash, chain, block_number, from_address, to_address,
			token_address, amount, amount_usd, gas_used, gas_price, gas_fee_usd,
			nonce, tx_timestamp, risk_score, flagged, flag_reason,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err := r.conn.pool.Exec(ctx, query,
		tx.ID, tx.TxHash, tx.Chain, tx.BlockNumber, tx.FromAddress, tx.ToAddress,
		tx.TokenAddress, tx.Amount, tx.AmountUSD, tx.GasUsed, tx.GasPrice, tx.GasFeeUSD,
		tx.Nonce, tx.TxTimestamp, tx.RiskScore, tx.Flagged, tx.FlagReason,
		time.Now(), time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetTransaction retrieves a transaction by ID
func (r *TransactionRepository) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `SELECT * FROM transactions WHERE id = $1`
	row := r.conn.pool.QueryRow(ctx, query, id)

	var tx domain.Transaction
	err := row.Scan(
		&tx.ID, &tx.TxHash, &tx.Chain, &tx.BlockNumber, &tx.FromAddress, &tx.ToAddress,
		&tx.TokenAddress, &tx.Amount, &tx.AmountUSD, &tx.GasUsed, &tx.GasPrice, &tx.GasFeeUSD,
		&tx.Nonce, &tx.TxTimestamp, &tx.RiskScore, &tx.Flagged, &tx.FlagReason,
		&tx.ReviewedAt, &tx.ReviewedBy, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	return &tx, nil
}

// GetTransactionByHash retrieves a transaction by hash
func (r *TransactionRepository) GetTransactionByHash(ctx context.Context, txHash string) (*domain.Transaction, error) {
	query := `SELECT * FROM transactions WHERE tx_hash = $1`
	row := r.conn.pool.QueryRow(ctx, query, txHash)

	var tx domain.Transaction
	err := row.Scan(
		&tx.ID, &tx.TxHash, &tx.Chain, &tx.BlockNumber, &tx.FromAddress, &tx.ToAddress,
		&tx.TokenAddress, &tx.Amount, &tx.AmountUSD, &tx.GasUsed, &tx.GasPrice, &tx.GasFeeUSD,
		&tx.Nonce, &tx.TxTimestamp, &tx.RiskScore, &tx.Flagged, &tx.FlagReason,
		&tx.ReviewedAt, &tx.ReviewedBy, &tx.CreatedAt, &tx.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	return &tx, nil
}

// UpdateTransaction updates an existing transaction
func (r *TransactionRepository) UpdateTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		UPDATE transactions SET
			risk_score = $1, flagged = $2, flag_reason = $3, reviewed_at = $4,
			reviewed_by = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.conn.pool.Exec(ctx, query,
		tx.RiskScore, tx.Flagged, tx.FlagReason, tx.ReviewedAt, tx.ReviewedBy,
		time.Now(), tx.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// ListTransactions retrieves transactions with filtering
func (r *TransactionRepository) ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]*domain.Transaction, int64, error) {
	// Implementation would include filtering logic
	return []*domain.Transaction{}, 0, nil
}

// GetTransactionsByAddress retrieves transactions for a wallet address
func (r *TransactionRepository) GetTransactionsByAddress(ctx context.Context, address string, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT * FROM transactions
		WHERE from_address = $1 OR to_address = $1
		ORDER BY tx_timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.conn.pool.Query(ctx, query, address, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		err := rows.Scan(
			&tx.ID, &tx.TxHash, &tx.Chain, &tx.BlockNumber, &tx.FromAddress, &tx.ToAddress,
			&tx.TokenAddress, &tx.Amount, &tx.AmountUSD, &tx.GasUsed, &tx.GasPrice, &tx.GasFeeUSD,
			&tx.Nonce, &tx.TxTimestamp, &tx.RiskScore, &tx.Flagged, &tx.FlagReason,
			&tx.ReviewedAt, &tx.ReviewedBy, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// GetTransactionsByTimeRange retrieves transactions within a time range
func (r *TransactionRepository) GetTransactionsByTimeRange(ctx context.Context, start, end time.Time) ([]*domain.Transaction, error) {
	query := `
		SELECT * FROM transactions
		WHERE tx_timestamp BETWEEN $1 AND $2
		ORDER BY tx_timestamp DESC
	`

	rows, err := r.conn.pool.Query(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		err := rows.Scan(
			&tx.ID, &tx.TxHash, &tx.Chain, &tx.BlockNumber, &tx.FromAddress, &tx.ToAddress,
			&tx.TokenAddress, &tx.Amount, &tx.AmountUSD, &tx.GasUsed, &tx.GasPrice, &tx.GasFeeUSD,
			&tx.Nonce, &tx.TxTimestamp, &tx.RiskScore, &tx.Flagged, &tx.FlagReason,
			&tx.ReviewedAt, &tx.ReviewedBy, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

// CountTransactions counts transactions matching filter
func (r *TransactionRepository) CountTransactions(ctx context.Context, filter domain.TransactionFilter) (int64, error) {
	return 0, nil
}
