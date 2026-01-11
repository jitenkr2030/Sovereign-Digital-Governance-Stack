-- Exchange Ingestion Service Database Schema
-- PostgreSQL migration script

-- Data Sources table
CREATE TABLE IF NOT EXISTS data_sources (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    exchange_type VARCHAR(50) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    api_key TEXT,
    api_secret TEXT,
    symbols TEXT[] NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    polling_rate INTERVAL NOT NULL DEFAULT '5 seconds',
    ws_enabled BOOLEAN NOT NULL DEFAULT false,
    auth_type VARCHAR(20) NOT NULL DEFAULT 'NONE',
    timeout INTERVAL NOT NULL DEFAULT '30 seconds',
    retry_attempts INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Market Data table
CREATE TABLE IF NOT EXISTS market_data (
    id UUID PRIMARY KEY,
    source_id VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    base_symbol VARCHAR(10) NOT NULL,
    quote_symbol VARCHAR(10) NOT NULL,
    price DECIMAL(24, 8) NOT NULL,
    volume DECIMAL(24, 8) NOT NULL,
    high_24h DECIMAL(24, 8),
    low_24h DECIMAL(24, 8),
    change_24h DECIMAL(10, 4),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Trades table
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY,
    source_id VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    trade_id VARCHAR(100) NOT NULL,
    price DECIMAL(24, 8) NOT NULL,
    quantity DECIMAL(24, 8) NOT NULL,
    quote_quantity DECIMAL(24, 8) NOT NULL,
    side VARCHAR(10) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Order Book table
CREATE TABLE IF NOT EXISTS orderbook (
    id UUID PRIMARY KEY,
    source_id VARCHAR(50) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    bids JSONB NOT NULL,
    asks JSONB NOT NULL,
    bid_depth DECIMAL(24, 8) NOT NULL,
    ask_depth DECIMAL(24, 8) NOT NULL,
    spread DECIMAL(24, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Ingestion Statistics table
CREATE TABLE IF NOT EXISTS ingestion_stats (
    source_id VARCHAR(50) PRIMARY KEY,
    total_messages BIGINT NOT NULL DEFAULT 0,
    processed_messages BIGINT NOT NULL DEFAULT 0,
    failed_messages BIGINT NOT NULL DEFAULT 0,
    last_message_time TIMESTAMP WITH TIME ZONE,
    average_latency INTERVAL,
    last_error TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'DISCONNECTED',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for market data queries
CREATE INDEX IF NOT EXISTS idx_market_data_source_symbol_time
    ON market_data(source_id, symbol, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_market_data_symbol_time
    ON market_data(symbol, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_market_data_timestamp
    ON market_data(timestamp DESC);

-- Indexes for trades queries
CREATE INDEX IF NOT EXISTS idx_trades_source_symbol_time
    ON trades(source_id, symbol, timestamp DESC);

-- Index for ingestion stats lookup
CREATE INDEX IF NOT EXISTS idx_ingestion_stats_source_id
    ON ingestion_stats(source_id);

-- Comments for documentation
COMMENT ON TABLE data_sources IS 'Configuration for exchange data sources';
COMMENT ON TABLE market_data IS 'Normalized market data from all exchanges';
COMMENT ON TABLE trades IS 'Individual trade executions from exchanges';
COMMENT ON TABLE orderbook IS 'Order book snapshots';
COMMENT ON TABLE ingestion_stats IS 'Statistics for monitoring ingestion performance';
