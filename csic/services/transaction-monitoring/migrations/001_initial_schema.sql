-- Transaction Monitoring Service Database Schema
-- Migration: 001_initial_schema

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tx_hash VARCHAR(128) NOT NULL UNIQUE,
    chain VARCHAR(64) NOT NULL,
    block_number BIGINT,
    from_address VARCHAR(128) NOT NULL,
    to_address VARCHAR(128),
    token_address VARCHAR(128),
    amount DECIMAL(32, 8) NOT NULL,
    amount_usd DECIMAL(32, 8),
    gas_used BIGINT,
    gas_price DECIMAL(32, 8),
    gas_fee_usd DECIMAL(32, 8),
    nonce BIGINT,
    tx_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    risk_score INTEGER DEFAULT 0,
    flagged BOOLEAN DEFAULT FALSE,
    flag_reason TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by VARCHAR(128),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_from_address ON transactions(from_address);
CREATE INDEX IF NOT EXISTS idx_transactions_to_address ON transactions(to_address);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(tx_timestamp);
CREATE INDEX IF NOT EXISTS idx_transactions_risk_score ON transactions(risk_score);
CREATE INDEX IF NOT EXISTS idx_transactions_flagged ON transactions(flagged);

-- Wallet profiles table
CREATE TABLE IF NOT EXISTS wallet_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(128) NOT NULL UNIQUE,
    chain VARCHAR(64) NOT NULL DEFAULT 'ethereum',
    first_seen TIMESTAMP WITH TIME ZONE,
    last_seen TIMESTAMP WITH TIME ZONE,
    tx_count INTEGER DEFAULT 0,
    total_volume_usd DECIMAL(32, 8) DEFAULT 0,
    avg_tx_value_usd DECIMAL(32, 8) DEFAULT 0,
    risk_score INTEGER DEFAULT 0,
    risk_level VARCHAR(32) DEFAULT 'LOW',
    is_sanctioned BOOLEAN DEFAULT FALSE,
    flag_count INTEGER DEFAULT 0,
    labels TEXT[],
    known_entity_type VARCHAR(64),
    known_entity_name VARCHAR(256),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_profiles_address ON wallet_profiles(address);
CREATE INDEX IF NOT EXISTS idx_wallet_profiles_risk_score ON wallet_profiles(risk_score);
CREATE INDEX IF NOT EXISTS idx_wallet_profiles_sanctioned ON wallet_profiles(is_sanctioned);

-- Sanctioned addresses table
CREATE TABLE IF NOT EXISTS sanctioned_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(128) NOT NULL,
    chain VARCHAR(64) NOT NULL,
    source_list VARCHAR(128) NOT NULL,
    reason TEXT,
    entity_name VARCHAR(256),
    entity_type VARCHAR(128),
    program VARCHAR(256),
    federal_register VARCHAR(128),
    remarks TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_sanctions_address ON sanctioned_addresses(address);
CREATE INDEX IF NOT EXISTS idx_sanctions_chain ON sanctioned_addresses(chain);
CREATE INDEX IF NOT EXISTS idx_sanctions_active ON sanctioned_addresses(is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sanctions_unique ON sanctioned_addresses(address, chain, source_list);

-- Monitoring rules table
CREATE TABLE IF NOT EXISTS monitoring_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(256) NOT NULL,
    description TEXT,
    rule_type VARCHAR(64) NOT NULL,
    condition JSONB NOT NULL,
    parameters JSONB,
    risk_weight DECIMAL(8, 2) DEFAULT 0,
    severity VARCHAR(32) DEFAULT 'WARNING',
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rules_active ON monitoring_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_rules_type ON monitoring_rules(rule_type);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_type VARCHAR(64) NOT NULL,
    transaction_id UUID REFERENCES transactions(id),
    wallet_address VARCHAR(128),
    severity VARCHAR(32) NOT NULL,
    risk_score DECIMAL(8, 2),
    status VARCHAR(32) DEFAULT 'OPEN',
    title VARCHAR(512) NOT NULL,
    description TEXT,
    triggered_rules UUID[],
    assigned_to UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_wallet ON alerts(wallet_address);
CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_alerts_transaction ON alerts(transaction_id);

-- Risk assessments table (audit trail)
CREATE TABLE IF NOT EXISTS risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID REFERENCES transactions(id),
    wallet_address VARCHAR(128),
    overall_score INTEGER NOT NULL,
    risk_level VARCHAR(32) NOT NULL,
    factors JSONB NOT NULL,
    recommendations TEXT[],
    requires_review BOOLEAN DEFAULT FALSE,
    flagged BOOLEAN DEFAULT FALSE,
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_risk_assessments_tx ON risk_assessments(transaction_id);
CREATE INDEX IF NOT EXISTS idx_risk_assessments_wallet ON risk_assessments(wallet_address);

-- Velocity tracking table
CREATE TABLE IF NOT EXISTS velocity_tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_address VARCHAR(128) NOT NULL,
    tx_count_24h INTEGER DEFAULT 0,
    volume_24h DECIMAL(32, 8) DEFAULT 0,
    tx_count_7d INTEGER DEFAULT 0,
    volume_7d DECIMAL(32, 8) DEFAULT 0,
    tx_count_30d INTEGER DEFAULT 0,
    volume_30d DECIMAL(32, 8) DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_velocity_wallet ON velocity_tracking(wallet_address);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_wallet_profiles_updated_at BEFORE UPDATE ON wallet_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default monitoring rules
INSERT INTO monitoring_rules (id, name, description, rule_type, condition, risk_weight, severity, is_active, priority) VALUES
(uuid_generate_v4(), 'Large Transaction Alert', 'Flags transactions exceeding $100,000', 'THRESHOLD',
 '{"max_amount": 100000, "currency": "USD"}', 25, 'ALERT', true, 100),
(uuid_generate_v4(), 'Very Large Transaction Alert', 'Flags transactions exceeding $1,000,000', 'THRESHOLD',
 '{"max_amount": 1000000, "currency": "USD"}', 50, 'CRITICAL', true, 200),
(uuid_generate_v4(), 'Structuring Detection', 'Detects potential structuring (multiple just-below-threshold transactions)', 'PATTERN',
 '{"pattern": "round_numbers", "threshold": 9000}', 40, 'ALERT', true, 150),
(uuid_generate_v4(), 'New Wallet High Value', 'Flags high-value transactions from new wallets (<24h)', 'VELOCITY',
 '{"max_age_hours": 24, "min_amount": 10000}', 30, 'WARNING', true, 120),
(uuid_generate_v4(), 'High Velocity Alert', 'Flags wallets with >50 transactions in 24h', 'VELOCITY',
 '{"max_transactions": 50, "window_hours": 24}', 35, 'ALERT', true, 130);
