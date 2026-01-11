-- CSIC Exchange Oversight Service Database Migrations
-- Migration 001: Initial schema creation

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN (
        'WASH_TRADING', 'SPOOFING', 'LAYERING', 'MANIPULATION',
        'PUMP_AND_DUMP', 'VOLUME_SPIKE', 'PRICE_DEVIATION',
        'LATENCY_ISSUE', 'CONNECTIVITY_ISSUE'
    )),
    exchange_id VARCHAR(100) NOT NULL,
    trading_pair VARCHAR(50),
    user_id VARCHAR(100),
    details JSONB NOT NULL DEFAULT '{}',
    evidence JSONB DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN (
        'OPEN', 'INVESTIGATING', 'RESOLVED', 'DISMISSED', 'ESCALATED'
    )),
    assigned_to VARCHAR(100),
    resolved_at TIMESTAMPTZ,
    resolution TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for alerts table
CREATE INDEX IF NOT EXISTS idx_alerts_exchange ON alerts(exchange_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_user ON alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_trading_pair ON alerts(trading_pair);
CREATE INDEX IF NOT EXISTS idx_alerts_exchange_status ON alerts(exchange_id, status);

-- Create detection_rules table
CREATE TABLE IF NOT EXISTS detection_rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN (
        'WASH_TRADING', 'SPOOFING', 'LAYERING', 'MANIPULATION',
        'PUMP_AND_DUMP', 'VOLUME_SPIKE', 'PRICE_DEVIATION',
        'LATENCY_ISSUE', 'CONNECTIVITY_ISSUE'
    )),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    logic JSONB NOT NULL DEFAULT '{}',
    time_window_ms INTEGER NOT NULL DEFAULT 60000,
    threshold FLOAT NOT NULL DEFAULT 1.0,
    cooldown_secs INTEGER NOT NULL DEFAULT 300,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    exchanges JSONB DEFAULT '[]',
    trading_pairs JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for detection_rules table
CREATE INDEX IF NOT EXISTS idx_rules_type ON detection_rules(alert_type);
CREATE INDEX IF NOT EXISTS idx_rules_active ON detection_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_rules_exchange ON detection_rules USING GIN (exchanges);
CREATE INDEX IF NOT EXISTS idx_rules_trading_pairs ON detection_rules USING GIN (trading_pairs);

-- Create exchange_profiles table
CREATE TABLE IF NOT EXISTS exchange_profiles (
    id UUID PRIMARY KEY,
    exchange_id VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    license_number VARCHAR(100),
    jurisdiction VARCHAR(100) NOT NULL,
    api_endpoint VARCHAR(500) NOT NULL,
    websocket_endpoint VARCHAR(500),
    rate_limit_rps INTEGER NOT NULL DEFAULT 100,
    throttle_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    max_latency_ms INTEGER NOT NULL DEFAULT 1000,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for exchange_profiles table
CREATE INDEX IF NOT EXISTS idx_exchange_jurisdiction ON exchange_profiles(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_exchange_active ON exchange_profiles(is_active);

-- Create exchange_health table
CREATE TABLE IF NOT EXISTS exchange_health (
    id UUID PRIMARY KEY,
    exchange_id VARCHAR(100) NOT NULL,
    health_score FLOAT NOT NULL DEFAULT 100.0 CHECK (health_score >= 0 AND health_score <= 100),
    latency_ms FLOAT NOT NULL DEFAULT 0,
    error_rate FLOAT NOT NULL DEFAULT 0 CHECK (error_rate >= 0 AND error_rate <= 1),
    trade_volume_24h FLOAT NOT NULL DEFAULT 0,
    trade_count_24h BIGINT NOT NULL DEFAULT 0,
    uptime_percent FLOAT NOT NULL DEFAULT 100.0 CHECK (uptime_percent >= 0 AND uptime_percent <= 100),
    last_trade_at TIMESTAMPTZ,
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN (
        'ACTIVE', 'DEGRADED', 'THROTTLED', 'SUSPENDED', 'OFFLINE'
    )),
    metrics JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for exchange_health table
CREATE INDEX IF NOT EXISTS idx_health_exchange ON exchange_health(exchange_id);
CREATE INDEX IF NOT EXISTS idx_health_score ON exchange_health(health_score);
CREATE INDEX IF NOT EXISTS idx_health_status ON exchange_health(status);
CREATE INDEX IF NOT EXISTS idx_health_updated ON exchange_health(updated_at DESC);

-- Create throttle_commands table for audit trail
CREATE TABLE IF NOT EXISTS throttle_commands (
    id UUID PRIMARY KEY,
    exchange_id VARCHAR(100) NOT NULL,
    command VARCHAR(20) NOT NULL CHECK (command IN ('LIMIT', 'SUSPEND', 'RESUME', 'BLOCK')),
    reason TEXT NOT NULL,
    target_rate_percent FLOAT NOT NULL DEFAULT 100.0,
    duration_secs INTEGER,
    issued_by VARCHAR(100) NOT NULL DEFAULT 'system',
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    execution_status VARCHAR(20) DEFAULT 'PENDING' CHECK (execution_status IN (
        'PENDING', 'EXECUTED', 'FAILED', 'CANCELLED'
    )),
    execution_details TEXT
);

-- Create indexes for throttle_commands table
CREATE INDEX IF NOT EXISTS idx_throttle_exchange ON throttle_commands(exchange_id);
CREATE INDEX IF NOT EXISTS idx_throttle_issued ON throttle_commands(issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_throttle_status ON throttle_commands(execution_status);

-- Create regulatory_reports table
CREATE TABLE IF NOT EXISTS regulatory_reports (
    id UUID PRIMARY KEY,
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN (
        'DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'ANNUAL', 'AD_HOC'
    )),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    exchange_id VARCHAR(100),
    summary JSONB NOT NULL DEFAULT '{}',
    alerts JSONB DEFAULT '[]',
    metrics JSONB DEFAULT '{}',
    format VARCHAR(20) NOT NULL DEFAULT 'JSON' CHECK (format IN ('JSON', 'CSV', 'PDF', 'XML')),
    version VARCHAR(20) NOT NULL DEFAULT '1.0',
    generated_by VARCHAR(100) NOT NULL DEFAULT 'system',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_by VARCHAR(100),
    approved_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    file_path VARCHAR(500)
);

-- Create indexes for regulatory_reports table
CREATE INDEX IF NOT EXISTS idx_reports_period ON regulatory_reports(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_reports_type ON regulatory_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_reports_exchange ON regulatory_reports(exchange_id);
CREATE INDEX IF NOT EXISTS idx_reports_generated ON regulatory_reports(generated_at DESC);

-- Insert default detection rules
INSERT INTO detection_rules (id, name, description, alert_type, severity, logic, time_window_ms, threshold, is_active)
VALUES
    -- Wash trading detection rule
    ('11111111-1111-1111-1111-111111111111', 'Self-Trading Detection', 'Detects trades where buyer and seller are the same user', 'WASH_TRADING', 'CRITICAL', 
     '{"type": "self_trading", "condition": "buyer_id == seller_id"}', 1000, 1, true),
    
    -- Volume spike detection rule
    ('22222222-2222-2222-2222-222222222222', 'Volume Spike Detection', 'Detects unusual volume spikes compared to moving average', 'VOLUME_SPIKE', 'HIGH',
     '{"type": "volume_spike", "comparison": "moving_average", "multiplier": 3.0}', 3600000, 3.0, true),
    
    -- Price deviation detection rule
    ('33333333-3333-3333-3333-333333333333', 'Price Deviation Detection', 'Detects prices deviating significantly from market average', 'PRICE_DEVIATION', 'MEDIUM',
     '{"type": "price_deviation", "comparison": "market_average", "tolerance": 0.05}', 300000, 0.05, true),
    
    -- Circular trading detection rule
    ('44444444-4444-4444-4444-444444444444', 'Circular Trading Detection', 'Detects back-and-forth trading between same user pairs', 'WASH_TRADING', 'HIGH',
     '{"type": "circular_trading", "min_trades": 5, "price_variance_tolerance": 0.001}', 60000, 5, true)
ON CONFLICT (id) DO NOTHING;

-- Insert sample exchange profile (example)
INSERT INTO exchange_profiles (id, exchange_id, name, license_number, jurisdiction, api_endpoint, max_latency_ms, throttle_enabled)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'binance', 'Binance', 'CRYPTO-2024-001', 'Cayman Islands', 'https://api.binance.com', 500, true),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'coinbase', 'Coinbase Prime', 'CRYPTO-2024-002', 'United States', 'https://api.coinbase.com', 300, true),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'kraken', 'Kraken', 'CRYPTO-2024-003', 'United States', 'https://api.kraken.com', 400, true)
ON CONFLICT (id) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_detection_rules_updated_at BEFORE UPDATE ON detection_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_exchange_profiles_updated_at BEFORE UPDATE ON exchange_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_exchange_health_updated_at BEFORE UPDATE ON exchange_health
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
