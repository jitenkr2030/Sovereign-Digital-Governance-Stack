-- +goose Up
-- +goose StatementBegin

-- Create market events hypertable for time-series data
CREATE TABLE IF NOT EXISTS market_events (
    id UUID PRIMARY KEY,
    exchange_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    event_type VARCHAR(20) NOT NULL,
    order_id UUID,
    user_id VARCHAR(100),
    price DECIMAL(24, 12) NOT NULL,
    quantity DECIMAL(24, 12) NOT NULL,
    filled_quantity DECIMAL(24, 12) DEFAULT 0,
    direction VARCHAR(10),
    status VARCHAR(20),
    "timestamp" TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    sequence_num BIGINT,
    raw_data TEXT,
    checksum VARCHAR(128)
);

-- Create index on exchange_id and symbol for filtering
CREATE INDEX IF NOT EXISTS idx_market_events_exchange_symbol 
ON market_events(exchange_id, symbol);

-- Create index on timestamp for time-range queries
CREATE INDEX IF NOT EXISTS idx_market_events_timestamp 
ON market_events("timestamp" DESC);

-- Create composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_market_events_exchange_symbol_timestamp 
ON market_events(exchange_id, symbol, "timestamp" DESC);

-- Create index on user_id for account-based queries
CREATE INDEX IF NOT EXISTS idx_market_events_user_id 
ON market_events(user_id);

-- Create index on order_id for order-based queries
CREATE INDEX IF NOT EXISTS idx_market_events_order_id 
ON market_events(order_id);

-- Create index on event_type for filtering by event type
CREATE INDEX IF NOT EXISTS idx_market_events_type 
ON market_events(event_type);

-- Create hypertable for time-series optimization (TimescaleDB)
SELECT create_hypertable('market_events', 'timestamp');

-- Enable compression on the hypertable
ALTER TABLE market_events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'exchange_id, symbol'
);

-- Add compression policy (compress data older than 7 days)
SELECT add_compression_policy('market_events', INTERVAL '7 days');

-- Add retention policy (delete data older than 1 year)
-- Uncomment the following line to enable retention
-- SELECT add_retention_policy('market_events', INTERVAL '1 year');

-- Create continuous aggregate for 1-minute price candles
CREATE MATERIALIZED VIEW IF NOT EXISTS candle_1m
WITH (timescaledb.continuous) AS
SELECT
    exchange_id,
    symbol,
    time_bucket('1 minute', "timestamp") AS bucket,
    FIRST(price, "timestamp") AS open_price,
    MAX(price) AS high_price,
    MIN(price) AS low_price,
    LAST(price, "timestamp") AS close_price,
    SUM(quantity) AS volume,
    COUNT(*) AS trade_count
FROM market_events
WHERE event_type = 'TRADE'
GROUP BY exchange_id, symbol, time_bucket('1 minute', "timestamp");

-- Create refresh policy for continuous aggregate
CREATE OR REPLACE FUNCTION refresh_candle_1m_policy()
RETURNS VOID AS $$
BEGIN
    CALL refresh_continuous_aggregate('candle_1m', NULL, NOW());
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop continuous aggregate
DROP MATERIALIZED VIEW IF EXISTS candle_1m CASCADE;

-- Drop retention policy (if exists)
-- SELECT drop_retention_policy('market_events');

-- Drop compression policy (if exists)
-- SELECT drop_compression_policy('market_events');

-- Drop hypertable
DROP TABLE IF EXISTS market_events CASCADE;

-- Drop indexes
DROP INDEX IF EXISTS idx_market_events_type;
DROP INDEX IF EXISTS idx_market_events_order_id;
DROP INDEX IF EXISTS idx_market_events_user_id;
DROP INDEX IF EXISTS idx_market_events_exchange_symbol_timestamp;
DROP INDEX IF EXISTS idx_market_events_timestamp;
DROP INDEX IF EXISTS idx_market_events_exchange_symbol;

-- +goose StatementEnd
