package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/platform/blockchain/indexer/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host      string
	Port      int
	Password  string
	DB        int
	KeyPrefix string
	PoolSize  int
}

// PostgresRepository handles database operations for the indexer
type PostgresRepository struct {
	db        *sql.DB
	logger    *zap.Logger
	keyPrefix string
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(cfg DatabaseConfig, logger *zap.Logger) (*PostgresRepository, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database",
		zap.String("database", cfg.Database),
		zap.String("host", cfg.Host))

	return &PostgresRepository{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// InitSchema initializes the database schema
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		// Blocks table
		`CREATE TABLE IF NOT EXISTS blocks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			hash VARCHAR(66) NOT NULL UNIQUE,
			parent_hash VARCHAR(66) NOT NULL,
			number BIGINT NOT NULL UNIQUE,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			miner VARCHAR(42) NOT NULL,
			difficulty VARCHAR(64) NOT NULL,
			gas_limit BIGINT NOT NULL,
			gas_used BIGINT NOT NULL,
			transaction_count INTEGER NOT NULL,
			base_fee_per_gas VARCHAR(32),
			extra_data TEXT,
			size BIGINT NOT NULL,
			uncle_count INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			indexed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Transactions table
		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			hash VARCHAR(66) NOT NULL UNIQUE,
			block_hash VARCHAR(66) NOT NULL,
			block_number BIGINT NOT NULL,
			transaction_index INTEGER NOT NULL,
			from_address VARCHAR(42) NOT NULL,
			to_address VARCHAR(42),
			value VARCHAR(64) NOT NULL,
			gas_price VARCHAR(32) NOT NULL,
			gas BIGINT NOT NULL,
			input_data TEXT,
			nonce BIGINT NOT NULL,
			v VARCHAR(10),
			r VARCHAR(66),
			s VARCHAR(66),
			status VARCHAR(20) NOT NULL,
			gas_used BIGINT,
			contract_address VARCHAR(42),
			cumulative_gas_used BIGINT,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Logs table
		`CREATE TABLE IF NOT EXISTS logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transaction_hash VARCHAR(66) NOT NULL,
			block_hash VARCHAR(66) NOT NULL,
			block_number BIGINT NOT NULL,
			log_index INTEGER NOT NULL,
			address VARCHAR(42) NOT NULL,
			topics TEXT[] NOT NULL,
			data TEXT NOT NULL,
			removed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Token transfers table
		`CREATE TABLE IF NOT EXISTS token_transfers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transaction_hash VARCHAR(66) NOT NULL,
			block_number BIGINT NOT NULL,
			token_address VARCHAR(42) NOT NULL,
			from_address VARCHAR(42) NOT NULL,
			to_address VARCHAR(42) NOT NULL,
			value VARCHAR(128),
			token_id VARCHAR(128),
			token_type VARCHAR(20) NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Address metadata table
		`CREATE TABLE IF NOT EXISTS address_meta (
			address VARCHAR(42) PRIMARY KEY,
			label VARCHAR(255),
			type VARCHAR(20) NOT NULL,
			first_seen TIMESTAMP WITH TIME ZONE NOT NULL,
			last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
			tx_count INTEGER DEFAULT 0,
			is_contract BOOLEAN DEFAULT FALSE,
			contract_name VARCHAR(255),
			contract_symbol VARCHAR(50),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Contracts table
		`CREATE TABLE IF NOT EXISTS contracts (
			address VARCHAR(42) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			symbol VARCHAR(50),
			decimals INTEGER,
			total_supply VARCHAR(128),
			compiler_version VARCHAR(100),
			abi TEXT NOT NULL,
			bytecode TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_blocks_number ON blocks(number DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash)`,
		`CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blocks(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_blocks_miner ON blocks(miner)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_block_number ON transactions(block_number DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_from_address ON transactions(from_address)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_to_address ON transactions(to_address)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_transaction_hash ON logs(transaction_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_block_number ON logs(block_number)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_address ON logs(address)`,
		`CREATE INDEX IF NOT EXISTS idx_token_transfers_token_address ON token_transfers(token_address)`,
		`CREATE INDEX IF NOT EXISTS idx_token_transfers_from_address ON token_transfers(from_address)`,
		`CREATE INDEX IF NOT EXISTS idx_token_transfers_to_address ON token_transfers(to_address)`,
		`CREATE INDEX IF NOT EXISTS idx_address_meta_type ON address_meta(type)`,
		`CREATE INDEX IF NOT EXISTS idx_address_meta_label ON address_meta(label)`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	r.logger.Info("Database schema initialized")
	return nil
}

// Block operations

// InsertBlock inserts a new block
func (r *PostgresRepository) InsertBlock(ctx context.Context, block *domain.Block) error {
	query := `
		INSERT INTO blocks (hash, parent_hash, number, timestamp, miner, difficulty, gas_limit, gas_used, transaction_count, base_fee_per_gas, extra_data, size, uncle_count, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		ON CONFLICT (number) DO UPDATE SET
			hash = EXCLUDED.hash,
			parent_hash = EXCLUDED.parent_hash,
			timestamp = EXCLUDED.timestamp,
			miner = EXCLUDED.miner,
			difficulty = EXCLUDED.difficulty,
			gas_limit = EXCLUDED.gas_limit,
			gas_used = EXCLUDED.gas_used,
			transaction_count = EXCLUDED.transaction_count,
			base_fee_per_gas = EXCLUDED.base_fee_per_gas,
			extra_data = EXCLUDED.extra_data,
			size = EXCLUDED.size,
			uncle_count = EXCLUDED.uncle_count,
			indexed_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		block.Hash, block.ParentHash, block.Number, block.Timestamp,
		block.Miner, block.Difficulty, block.GasLimit, block.GasUsed,
		block.TransactionCount, block.BaseFeePerGas, block.ExtraData,
		block.Size, block.UncleCount)

	if err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}

	return nil
}

// GetBlockByHash retrieves a block by hash
func (r *PostgresRepository) GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error) {
	query := `
		SELECT id, hash, parent_hash, number, timestamp, miner, difficulty, gas_limit, gas_used, transaction_count, base_fee_per_gas, extra_data, size, uncle_count, created_at, indexed_at
		FROM blocks WHERE hash = $1
	`

	block := &domain.Block{}
	var createdAt, indexedAt time.Time

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&block.ID, &block.Hash, &block.ParentHash, &block.Number,
		&block.Timestamp, &block.Miner, &block.Difficulty, &block.GasLimit,
		&block.GasUsed, &block.TransactionCount, &block.BaseFeePerGas,
		&block.ExtraData, &block.Size, &block.UncleCount, &createdAt, &indexedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	block.CreatedAt = createdAt
	block.IndexedAt = indexedAt

	return block, nil
}

// GetBlockByNumber retrieves a block by number
func (r *PostgresRepository) GetBlockByNumber(ctx context.Context, number int64) (*domain.Block, error) {
	query := `
		SELECT id, hash, parent_hash, number, timestamp, miner, difficulty, gas_limit, gas_used, transaction_count, base_fee_per_gas, extra_data, size, uncle_count, created_at, indexed_at
		FROM blocks WHERE number = $1
	`

	block := &domain.Block{}
	var createdAt, indexedAt time.Time

	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&block.ID, &block.Hash, &block.ParentHash, &block.Number,
		&block.Timestamp, &block.Miner, &block.Difficulty, &block.GasLimit,
		&block.GasUsed, &block.TransactionCount, &block.BaseFeePerGas,
		&block.ExtraData, &block.Size, &block.UncleCount, &createdAt, &indexedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	block.CreatedAt = createdAt
	block.IndexedAt = indexedAt

	return block, nil
}

// GetLatestBlockNumber retrieves the latest block number
func (r *PostgresRepository) GetLatestBlockNumber(ctx context.Context) (int64, error) {
	query := `SELECT COALESCE(MAX(number), 0) FROM blocks`

	var number int64
	err := r.db.QueryRowContext(ctx, query).Scan(&number)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return number, nil
}

// Transaction operations

// InsertTransaction inserts a new transaction
func (r *PostgresRepository) InsertTransaction(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (hash, block_hash, block_number, transaction_index, from_address, to_address, value, gas_price, gas, input_data, nonce, v, r, s, status, gas_used, contract_address, cumulative_gas_used, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		ON CONFLICT (hash) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		tx.Hash, tx.BlockHash, tx.BlockNumber, tx.TransactionIndex,
		tx.From, tx.To, tx.Value, tx.GasPrice, tx.Gas, tx.InputData,
		tx.Nonce, tx.V, tx.R, tx.S, tx.Status, tx.GasUsed,
		tx.ContractAddress, tx.CumulativeGasUsed, tx.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

// GetTransactionByHash retrieves a transaction by hash
func (r *PostgresRepository) GetTransactionByHash(ctx context.Context, hash string) (*domain.Transaction, error) {
	query := `
		SELECT id, hash, block_hash, block_number, transaction_index, from_address, to_address, value, gas_price, gas, input_data, nonce, v, r, s, status, gas_used, contract_address, cumulative_gas_used, timestamp, created_at
		FROM transactions WHERE hash = $1
	`

	tx := &domain.Transaction{}
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&tx.ID, &tx.Hash, &tx.BlockHash, &tx.BlockNumber, &tx.TransactionIndex,
		&tx.From, &tx.To, &tx.Value, &tx.GasPrice, &tx.Gas, &tx.InputData,
		&tx.Nonce, &tx.V, &tx.R, &tx.S, &tx.Status, &tx.GasUsed,
		&tx.ContractAddress, &tx.CumulativeGasUsed, &tx.Timestamp, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	tx.CreatedAt = createdAt

	return tx, nil
}

// GetTransactionsByAddress retrieves transactions for an address
func (r *PostgresRepository) GetTransactionsByAddress(ctx context.Context, address string, limit, offset int) ([]domain.Transaction, error) {
	query := `
		SELECT id, hash, block_hash, block_number, transaction_index, from_address, to_address, value, gas_price, gas, input_data, nonce, v, r, s, status, gas_used, contract_address, cumulative_gas_used, timestamp, created_at
		FROM transactions
		WHERE from_address = $1 OR to_address = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, address, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		var createdAt time.Time

		err := rows.Scan(
			&tx.ID, &tx.Hash, &tx.BlockHash, &tx.BlockNumber, &tx.TransactionIndex,
			&tx.From, &tx.To, &tx.Value, &tx.GasPrice, &tx.Gas, &tx.InputData,
			&tx.Nonce, &tx.V, &tx.R, &tx.S, &tx.Status, &tx.GasUsed,
			&tx.ContractAddress, &tx.CumulativeGasUsed, &tx.Timestamp, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		tx.CreatedAt = createdAt
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// Log operations

// InsertLog inserts a new log
func (r *PostgresRepository) InsertLog(ctx context.Context, log *domain.Log) error {
	topicsJSON, err := json.Marshal(log.Topics)
	if err != nil {
		return fmt.Errorf("failed to marshal topics: %w", err)
	}

	query := `
		INSERT INTO logs (transaction_hash, block_hash, block_number, log_index, address, topics, data, removed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING
	`

	_, err = r.db.ExecContext(ctx, query,
		log.TransactionHash, log.BlockHash, log.BlockNumber, log.LogIndex,
		log.Address, topicsJSON, log.Data, log.Removed)

	if err != nil {
		return fmt.Errorf("failed to insert log: %w", err)
	}

	return nil
}

// GetLogsByAddress retrieves logs by address
func (r *PostgresRepository) GetLogsByAddress(ctx context.Context, address string, limit, offset int) ([]domain.Log, error) {
	query := `
		SELECT id, transaction_hash, block_hash, block_number, log_index, address, topics, data, removed, created_at
		FROM logs
		WHERE address = $1
		ORDER BY block_number DESC, log_index ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, address, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.Log
	for rows.Next() {
		var log domain.Log
		var topicsJSON []byte
		var createdAt time.Time

		err := rows.Scan(
			&log.ID, &log.TransactionHash, &log.BlockHash, &log.BlockNumber,
			&log.LogIndex, &log.Address, &topicsJSON, &log.Data, &log.Removed, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}

		if err := json.Unmarshal(topicsJSON, &log.Topics); err != nil {
			return nil, fmt.Errorf("failed to unmarshal topics: %w", err)
		}

		log.CreatedAt = createdAt
		logs = append(logs, log)
	}

	return logs, nil
}

// Token transfer operations

// InsertTokenTransfer inserts a new token transfer
func (r *PostgresRepository) InsertTokenTransfer(ctx context.Context, transfer *domain.TokenTransfer) error {
	query := `
		INSERT INTO token_transfers (transaction_hash, block_number, token_address, from_address, to_address, value, token_id, token_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		transfer.TransactionHash, transfer.BlockNumber, transfer.TokenAddress,
		transfer.From, transfer.To, transfer.Value, transfer.TokenID,
		transfer.TokenType, transfer.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to insert token transfer: %w", err)
	}

	return nil
}

// Address metadata operations

// UpsertAddressMeta upserts address metadata
func (r *PostgresRepository) UpsertAddressMeta(ctx context.Context, meta *domain.AddressMeta) error {
	query := `
		INSERT INTO address_meta (address, label, type, first_seen, last_seen, tx_count, is_contract, contract_name, contract_symbol, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (address) DO UPDATE SET
			label = COALESCE(EXCLUDED.label, address_meta.label),
			type = EXCLUDED.type,
			last_seen = EXCLUDED.last_seen,
			tx_count = address_meta.tx_count + 1,
			is_contract = EXCLUDED.is_contract,
			contract_name = COALESCE(EXCLUDED.contract_name, address_meta.contract_name),
			contract_symbol = COALESCE(EXCLUDED.contract_symbol, address_meta.contract_symbol),
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		meta.Address, meta.Label, meta.Type, meta.FirstSeen,
		meta.LastSeen, meta.TxCount, meta.IsContract, meta.ContractName,
		meta.ContractSymbol)

	if err != nil {
		return fmt.Errorf("failed to upsert address meta: %w", err)
	}

	return nil
}

// GetAddressMeta retrieves address metadata
func (r *PostgresRepository) GetAddressMeta(ctx context.Context, address string) (*domain.AddressMeta, error) {
	query := `
		SELECT address, label, type, first_seen, last_seen, tx_count, is_contract, contract_name, contract_symbol, created_at, updated_at
		FROM address_meta WHERE address = $1
	`

	meta := &domain.AddressMeta{}
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, address).Scan(
		&meta.Address, &meta.Label, &meta.Type, &meta.FirstSeen,
		&meta.LastSeen, &meta.TxCount, &meta.IsContract, &meta.ContractName,
		&meta.ContractSymbol, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get address meta: %w", err)
	}

	meta.CreatedAt = createdAt
	meta.UpdatedAt = updatedAt

	return meta, nil
}
