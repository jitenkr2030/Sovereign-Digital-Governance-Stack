-- +migrate Up
-- +migrate StatementBegin

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Exchange Oversight Schema
CREATE TABLE IF NOT EXISTS oversight_exchanges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    exchange_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    api_endpoint VARCHAR(512),
    status VARCHAR(32) DEFAULT 'ACTIVE',
    health_score DECIMAL(5,2) DEFAULT 100.00,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS oversight_trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trade_id VARCHAR(128) UNIQUE NOT NULL,
    exchange_id VARCHAR(64) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    price DECIMAL(24,12) NOT NULL,
    quantity DECIMAL(24,12) NOT NULL,
    quote_volume DECIMAL(24,12),
    buyer_order_id VARCHAR(128),
    seller_order_id VARCHAR(128),
    buyer_user_id VARCHAR(128),
    seller_user_id VARCHAR(128),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_maker BOOLEAN DEFAULT FALSE,
    fee_currency VARCHAR(16),
    fee_amount DECIMAL(24,12),
    raw_message JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS oversight_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_type VARCHAR(64) NOT NULL,
    severity VARCHAR(32) NOT NULL,
    exchange_id VARCHAR(64),
    symbol VARCHAR(32),
    user_id VARCHAR(128),
    trade_id UUID,
    details JSONB NOT NULL,
    evidence JSONB,
    status VARCHAR(32) DEFAULT 'OPEN',
    assigned_to VARCHAR(128),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_oversight_trades_exchange_symbol ON oversight_trades(exchange_id, symbol);
CREATE INDEX IF NOT EXISTS idx_oversight_trades_timestamp ON oversight_trades(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_oversight_trades_buyer_seller ON oversight_trades(buyer_user_id, seller_user_id);
CREATE INDEX IF NOT EXISTS idx_oversight_alerts_status ON oversight_alerts(status);
CREATE INDEX IF NOT EXISTS idx_oversight_alerts_created ON oversight_alerts(created_at DESC);

-- Custodial Wallet Governance Schema
CREATE TABLE IF NOT EXISTS governance_custodians (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    custodian_id VARCHAR(64) UNIQUE NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    license_number VARCHAR(128),
    jurisdiction VARCHAR(64) NOT NULL,
    regulatory_body VARCHAR(128),
    kyc_level VARCHAR(32) DEFAULT 'BASIC',
    status VARCHAR(32) DEFAULT 'ACTIVE',
    license_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS governance_wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(128) NOT NULL,
    chain VARCHAR(32) NOT NULL,
    custodian_id UUID REFERENCES governance_custodians(id),
    wallet_type VARCHAR(32) DEFAULT 'CUSTODIAL',
    asset_type VARCHAR(32) DEFAULT 'MULTI',
    status VARCHAR(32) DEFAULT 'PENDING',
    kyc_level VARCHAR(32),
    risk_level VARCHAR(32) DEFAULT 'LOW',
    daily_limit DECIMAL(24,2),
    balance DECIMAL(24,8),
    last_activity TIMESTAMP WITH TIME ZONE,
    blacklisted_at TIMESTAMP WITH TIME ZONE,
    blacklist_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(address, chain)
);

CREATE TABLE IF NOT EXISTS governance_whitelist (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID REFERENCES governance_wallets(id),
    whitelist_type VARCHAR(32) NOT NULL,
    reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    approved_by VARCHAR(128),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS governance_blacklist (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID REFERENCES governance_wallets(id),
    address VARCHAR(128) NOT NULL,
    chain VARCHAR(32) NOT NULL,
    reason TEXT NOT NULL,
    severity VARCHAR(32) DEFAULT 'HIGH',
    source VARCHAR(64),
    expires_at TIMESTAMP WITH TIME ZONE,
    blocked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(address, chain)
);

CREATE INDEX IF NOT EXISTS idx_governance_wallets_status ON governance_wallets(status);
CREATE INDEX IF NOT EXISTS idx_governance_wallets_custodian ON governance_wallets(custodian_id);
CREATE INDEX IF NOT EXISTS idx_governance_blacklist_address ON governance_blacklist(address);

-- Transaction Monitoring Schema
CREATE TABLE IF NOT EXISTS monitoring_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tx_hash VARCHAR(256) UNIQUE NOT NULL,
    chain VARCHAR(32) NOT NULL,
    block_number BIGINT,
    from_address VARCHAR(128) NOT NULL,
    to_address VARCHAR(128),
    token_address VARCHAR(128),
    amount DECIMAL(24,8) NOT NULL,
    amount_usd DECIMAL(24,2),
    gas_used BIGINT,
    gas_price DECIMAL(24,8),
    gas_fee_usd DECIMAL(24,2),
    nonce INTEGER,
    tx_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    risk_score INTEGER DEFAULT 0,
    risk_factors JSONB,
    flagged BOOLEAN DEFAULT FALSE,
    flag_reason TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by VARCHAR(128),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitoring_sanctioned_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(128) NOT NULL,
    chain VARCHAR(32) NOT NULL,
    source_list VARCHAR(64) NOT NULL,
    reason TEXT,
    entity_name VARCHAR(256),
    entity_type VARCHAR(64),
    program VARCHAR(128),
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(address, chain, source_list)
);

CREATE TABLE IF NOT EXISTS monitoring_wallet_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(128) NOT NULL,
    chain VARCHAR(32) NOT NULL,
    first_seen TIMESTAMP WITH TIME ZONE,
    last_seen TIMESTAMP WITH TIME ZONE,
    tx_count INTEGER DEFAULT 0,
    total_volume_usd DECIMAL(24,2) DEFAULT 0,
    avg_tx_value_usd DECIMAL(24,2) DEFAULT 0,
    connected_tags JSONB,
    risk_indicators JSONB,
    wallet_age_hours INTEGER,
    is_contract BOOLEAN DEFAULT FALSE,
    contract_type VARCHAR(32),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(address, chain)
);

CREATE INDEX IF NOT EXISTS idx_monitoring_transactions_chain ON monitoring_transactions(chain);
CREATE INDEX IF NOT EXISTS idx_monitoring_transactions_from ON monitoring_transactions(from_address);
CREATE INDEX IF NOT EXISTS idx_monitoring_transactions_to ON monitoring_transactions(to_address);
CREATE INDEX IF NOT EXISTS idx_monitoring_transactions_timestamp ON monitoring_transactions(tx_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_monitoring_transactions_flagged ON monitoring_transactions(flagged);
CREATE INDEX IF NOT EXISTS idx_monitoring_sanctioned ON monitoring_sanctioned_addresses(address);

-- Energy Monitoring Schema
CREATE TABLE IF NOT EXISTS energy_facilities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    facility_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    owner_id VARCHAR(128),
    location VARCHAR(256),
    country VARCHAR(64),
    power_source_type VARCHAR(64) NOT NULL,
    renewable_source VARCHAR(64),
    total_capacity_kw DECIMAL(24,2),
    status VARCHAR(32) DEFAULT 'ACTIVE',
    grid_operator VARCHAR(128),
    co2_per_kwh DECIMAL(10,6) DEFAULT 0,
    certifications JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS energy_telemetry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    facility_id UUID REFERENCES energy_facilities(id),
    hashrate_th DECIMAL(24,2),
    power_consumption_kw DECIMAL(24,2),
    efficiency_th_per_kw DECIMAL(24,2),
    ambient_temperature DECIMAL(8,2),
    humidity_percent DECIMAL(5,2),
    uptime_percent DECIMAL(5,2),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS energy_mining_rewards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    facility_id UUID REFERENCES energy_facilities(id),
    reward_amount DECIMAL(24,8) NOT NULL,
    reward_currency VARCHAR(16) NOT NULL,
    usd_value DECIMAL(24,2),
    block_height BIGINT,
    pool_address VARCHAR(128),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS energy_impact_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    facility_id UUID REFERENCES energy_facilities(id),
    report_type VARCHAR(32) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    total_energy_kwh DECIMAL(24,2),
    renewable_energy_kwh DECIMAL(24,2),
    total_co2_kg DECIMAL(24,2),
    carbon_intensity DECIMAL(10,4),
    efficiency_score DECIMAL(5,2),
    recommendations JSONB,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_energy_telemetry_facility ON energy_telemetry(facility_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_energy_rewards_facility ON energy_mining_rewards(facility_id, timestamp DESC);

-- Regulatory Reporting Schema
CREATE TABLE IF NOT EXISTS reporting_report_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_type VARCHAR(64) NOT NULL,
    format VARCHAR(16) NOT NULL,
    module_source VARCHAR(64),
    requester_id VARCHAR(128) NOT NULL,
    requester_email VARCHAR(256),
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    filters JSONB,
    status VARCHAR(32) DEFAULT 'PENDING',
    file_path VARCHAR(512),
    file_size_bytes BIGINT,
    download_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reporting_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(128) NOT NULL,
    report_type VARCHAR(64) NOT NULL,
    description TEXT,
    template_content JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    version INTEGER DEFAULT 1,
    created_by VARCHAR(128),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reporting_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_request_id UUID REFERENCES reporting_report_requests(id),
    action VARCHAR(64) NOT NULL,
    actor_id VARCHAR(128),
    actor_ip VARCHAR(45),
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reporting_requests_status ON reporting_report_requests(status);
CREATE INDEX IF NOT EXISTS idx_reporting_requests_requester ON reporting_report_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_reporting_requests_created ON reporting_report_requests(created_at DESC);

-- +migrate StatementEnd
