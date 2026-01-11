-- Blockchain Indexer Database Migration
-- Creates all necessary tables for blockchain data indexing

-- Create blocks table
CREATE TABLE IF NOT EXISTS blocks (
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
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
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
);

-- Create logs table
CREATE TABLE IF NOT EXISTS logs (
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
);

-- Create token transfers table
CREATE TABLE IF NOT EXISTS token_transfers (
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
);

-- Create address metadata table
CREATE TABLE IF NOT EXISTS address_meta (
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
);

-- Create contracts table
CREATE TABLE IF NOT EXISTS contracts (
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
);

-- Create indexes for blocks
CREATE INDEX IF NOT EXISTS idx_blocks_number ON blocks(number DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_hash ON blocks(hash);
CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blocks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_blocks_miner ON blocks(miner);

-- Create indexes for transactions
CREATE INDEX IF NOT EXISTS idx_transactions_block_number ON transactions(block_number DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS idx_transactions_from_address ON transactions(from_address);
CREATE INDEX IF NOT EXISTS idx_transactions_to_address ON transactions(to_address);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC);

-- Create indexes for logs
CREATE INDEX IF NOT EXISTS idx_logs_transaction_hash ON logs(transaction_hash);
CREATE INDEX IF NOT EXISTS idx_logs_block_number ON logs(block_number);
CREATE INDEX IF NOT EXISTS idx_logs_address ON logs(address);

-- Create indexes for token transfers
CREATE INDEX IF NOT EXISTS idx_token_transfers_token_address ON token_transfers(token_address);
CREATE INDEX IF NOT EXISTS idx_token_transfers_from_address ON token_transfers(from_address);
CREATE INDEX IF NOT EXISTS idx_token_transfers_to_address ON token_transfers(to_address);
CREATE INDEX IF NOT EXISTS idx_token_transfers_block_number ON token_transfers(block_number DESC);

-- Create indexes for address metadata
CREATE INDEX IF NOT EXISTS idx_address_meta_type ON address_meta(type);
CREATE INDEX IF NOT EXISTS idx_address_meta_label ON address_meta(label);
CREATE INDEX IF NOT EXISTS idx_address_meta_tx_count ON address_meta(tx_count DESC);

-- Insert known contract addresses
INSERT INTO address_meta (address, label, type, first_seen, last_seen, tx_count, is_contract) VALUES
('0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D', 'Uniswap V2 Router', 'PROTOCOL', NOW(), NOW(), 0, TRUE),
('0xE592427A0AEce92De3Edee1F18E0157C05861564', 'Uniswap V3 Router', 'PROTOCOL', NOW(), NOW(), 0, TRUE),
('0x3fC91A3afd70395Cd496C647d5a6CC9D4b2b7FAD', 'Universal Router', 'PROTOCOL', NOW(), NOW(), 0, TRUE);
