-- +goose Up
-- +goose StatementBegin

-- Create alerts table for storing regulatory alerts
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    exchange_id UUID NOT NULL,
    symbol VARCHAR(20),
    account_id VARCHAR(100),
    order_id UUID,
    trade_id UUID,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    evidence JSONB NOT NULL DEFAULT '{}',
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    pattern_confidence DECIMAL(5,2) NOT NULL DEFAULT 0,
    detected_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    resolution TEXT,
    assigned_to UUID,
    tags JSONB NOT NULL DEFAULT '[]',
    related_alert_ids UUID[] NOT NULL DEFAULT '{}',
    audit_hash VARCHAR(128)
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(type);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_exchange_id ON alerts(exchange_id);
CREATE INDEX IF NOT EXISTS idx_alerts_symbol ON alerts(symbol);
CREATE INDEX IF NOT EXISTS idx_alerts_detected_at ON alerts(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_status_detected ON alerts(status, detected_at DESC);

-- Composite index for filtered queries
CREATE INDEX IF NOT EXISTS idx_alerts_exchange_type_status ON alerts(exchange_id, type, status);

-- Full-text search index on title and description
CREATE INDEX IF NOT EXISTS idx_alerts_search ON alerts USING GIN (
    to_tsvector('english', title || ' ' || description)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_alerts_search;
DROP INDEX IF EXISTS idx_alerts_exchange_type_status;
DROP INDEX IF EXISTS idx_alerts_status_detected;
DROP INDEX IF EXISTS idx_alerts_symbol;
DROP INDEX IF EXISTS idx_alerts_exchange_id;
DROP INDEX IF EXISTS idx_alerts_severity;
DROP INDEX IF EXISTS idx_alerts_type;
DROP INDEX IF EXISTS idx_alerts_status;
DROP TABLE IF EXISTS alerts;

-- +goose StatementEnd
