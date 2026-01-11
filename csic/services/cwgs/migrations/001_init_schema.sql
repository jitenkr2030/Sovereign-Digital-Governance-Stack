-- CSIC Custodial Wallet Governance System Database Migrations
-- Migration 001: Initial schema creation

-- Create enum types
DO $$ BEGIN
    CREATE TYPE wallet_type AS ENUM ('HOT', 'COLD', 'OPERATIONAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE wallet_status AS ENUM (
        'ACTIVE', 
        'FROZEN', 
        'PENDING_RECOVERY', 
        'ARCHIVED', 
        'COMPROMISED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE freeze_reason AS ENUM (
        'REGULATORY_HOLD',
        'SUSPICIOUS_ACTIVITY',
        'LEGAL_ORDER',
        'USER_REQUEST',
        'SECURITY_MEASURE',
        'MAINTENANCE'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE chain_type AS ENUM (
        'BTC', 'ETH', 'MATIC', 'SOL', 'ARB', 'OPT'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE policy_severity AS ENUM ('BLOCKING', 'WARNING', 'ENHANCED_REVIEW');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE blacklist_source AS ENUM (
        'REGULATORY', 'LAW_ENFORCEMENT', 'INTERNAL_RISK',
        'EXCHANGE_REPORTED', 'COMMUNITY_REPORTED', 'AUTOMATED_DETECTION'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE audit_category AS ENUM (
        'WALLET', 'POLICY', 'SIGNING', 'KEY_MANAGEMENT',
        'EMERGENCY', 'AUTHENTICATION', 'ADMINISTRATION',
        'COMPLIANCE', 'SYSTEM'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE audit_severity AS ENUM ('INFO', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    address VARCHAR(100) NOT NULL,
    chain_type VARCHAR(10) NOT NULL,
    wallet_type VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    signing_policy_id UUID,
    hsm_key_handle VARCHAR(255) NOT NULL,
    public_key BYTEA NOT NULL,
    max_tx_limit DECIMAL(20, 8) NOT NULL DEFAULT 100.00000000,
    daily_limit DECIMAL(20, 8) NOT NULL DEFAULT 1000.00000000,
    daily_usage DECIMAL(20, 8) NOT NULL DEFAULT 0,
    last_activity_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT wallets_address_chain_unique UNIQUE (address, chain_type)
);

CREATE INDEX IF NOT EXISTS idx_wallets_status ON wallets(status);
CREATE INDEX IF NOT EXISTS idx_wallets_wallet_type ON wallets(wallet_type);
CREATE INDEX IF NOT EXISTS idx_wallets_chain ON wallets(chain_type);
CREATE INDEX IF NOT EXISTS idx_wallets_created_at ON wallets(created_at DESC);

-- Create policies table
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 0,
    rules_json JSONB NOT NULL DEFAULT '[]',
    applies_to_wallet_types JSONB NOT NULL DEFAULT '[]',
    applies_to_chains JSONB NOT NULL DEFAULT '[]',
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_until TIMESTAMPTZ,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policies_active ON policies(is_active, effective_from, effective_until);
CREATE INDEX IF NOT EXISTS idx_policies_priority ON policies(priority DESC);

-- Create blacklist table
CREATE TABLE IF NOT EXISTS blacklist (
    id UUID PRIMARY KEY,
    address VARCHAR(100) NOT NULL,
    chain_type VARCHAR(10),
    reason TEXT NOT NULL,
    source VARCHAR(30) NOT NULL,
    risk_level SMALLINT NOT NULL DEFAULT 3,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blacklist_address ON blacklist(address);
CREATE INDEX IF NOT EXISTS idx_blacklist_chain ON blacklist(chain_type);
CREATE INDEX IF NOT EXISTS idx_blacklist_active ON blacklist(is_active);
CREATE INDEX IF NOT EXISTS idx_blacklist_source ON blacklist(source);

-- Create whitelist table
CREATE TABLE IF NOT EXISTS whitelist (
    id UUID PRIMARY KEY,
    address VARCHAR(100) NOT NULL,
    chain_type VARCHAR(10),
    label VARCHAR(200) NOT NULL,
    entity_name VARCHAR(200),
    verification_level SMALLINT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_whitelist_address ON whitelist(address);
CREATE INDEX IF NOT EXISTS idx_whitelist_active ON whitelist(is_active);

-- Create key_metadata table
CREATE TABLE IF NOT EXISTS key_metadata (
    id UUID PRIMARY KEY,
    label VARCHAR(100) NOT NULL,
    algorithm VARCHAR(30) NOT NULL,
    usage VARCHAR(20) NOT NULL,
    hsm_slot_id VARCHAR(100) NOT NULL,
    hsm_key_handle VARCHAR(255) NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    wallet_id UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    rotation_policy_id UUID,
    usage_count BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_key_metadata_status ON key_metadata(status);
CREATE INDEX IF NOT EXISTS idx_key_metadata_wallet ON key_metadata(wallet_id);

-- Create multisig_configs table
CREATE TABLE IF NOT EXISTS multisig_configs (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    required_signatures SMALLINT NOT NULL,
    total_signers SMALLINT NOT NULL,
    is_rolling BOOLEAN NOT NULL DEFAULT FALSE,
    signature_timeout_seconds INTEGER NOT NULL DEFAULT 3600,
    allow_partial_with_escalation BOOLEAN NOT NULL DEFAULT FALSE,
    escalation_threshold SMALLINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create signers table
CREATE TABLE IF NOT EXISTS signers (
    id UUID PRIMARY KEY,
    multisig_config_id UUID NOT NULL,
    signer_type VARCHAR(30) NOT NULL,
    identifier VARCHAR(200) NOT NULL,
    label VARCHAR(100) NOT NULL,
    hsm_key_handle VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    signer_order SMALLINT
);

CREATE INDEX IF NOT EXISTS idx_signers_config ON signers(multisig_config_id);

-- Create signing_requests table
CREATE TABLE IF NOT EXISTS signing_requests (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL,
    multisig_config_id UUID NOT NULL,
    payload BYTEA NOT NULL,
    payload_hash BYTEA NOT NULL,
    required_signatures SMALLINT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_signing_requests_wallet ON signing_requests(wallet_id);
CREATE INDEX IF NOT EXISTS idx_signing_requests_status ON signing_requests(status);
CREATE INDEX IF NOT EXISTS idx_signing_requests_created ON signing_requests(created_at DESC);

-- Create signatures table
CREATE TABLE IF NOT EXISTS signatures (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL,
    signer_id UUID NOT NULL,
    signature_bytes BYTEA NOT NULL,
    signature_type VARCHAR(20) NOT NULL,
    signed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_valid BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_signatures_request ON signatures(request_id);

-- Create audit_logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    category VARCHAR(30) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'INFO',
    actor_id VARCHAR(100) NOT NULL,
    actor_type VARCHAR(20) NOT NULL,
    actor_label VARCHAR(200),
    actor_ip VARCHAR(45),
    resource_id UUID,
    resource_type VARCHAR(50),
    description TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    before_state JSONB,
    after_state JSONB,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    request_id UUID,
    correlation_id UUID,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_category ON audit_logs(category);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_id, resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_correlation ON audit_logs(correlation_id);

-- Create emergency_access_configs table
CREATE TABLE IF NOT EXISTS emergency_access_configs (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    required_approvals SMALLINT NOT NULL DEFAULT 2,
    admin_ids UUID[] NOT NULL DEFAULT '{}',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    operation_timeout_seconds INTEGER NOT NULL DEFAULT 300,
    allowed_actions TEXT[] NOT NULL DEFAULT '{}',
    cooldown_seconds INTEGER NOT NULL DEFAULT 3600,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers
CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_blacklist_updated_at BEFORE UPDATE ON blacklist
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_whitelist_updated_at BEFORE UPDATE ON whitelist
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_multisig_configs_updated_at BEFORE UPDATE ON multisig_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default policies
INSERT INTO policies (id, name, description, rules_json, created_by)
VALUES
    (
        'a0000000-0000-0000-0000-000000000001',
        'Default Security Policy',
        'Default security policy for all wallets',
        '[
            {"BlacklistAddress": {"address": "0x000000000000000000000000000000000000dEaD", "reason": "Burn address", "source": "INTERNAL_RISK"}},
            {"MaxTransactionLimit": {"max_amount": "1000.00000000", "chain_type": null}},
            {"DailyWithdrawalLimit": {"max_amount": "10000.00000000", "window_hours": 24}}
        ]',
        'system'
    ),
    (
        'a0000000-0000-0000-0000-000000000002',
        'Cold Storage Policy',
        'Strict policy for cold storage wallets',
        '[
            {"MultiSigRequirement": {"required_signatures": 3, "total_signers": 5, "description": "3-of-5 multisig required"}},
            {"VelocityLimit": {"min_interval_seconds": 86400, "per_wallet": true}}
        ]',
        'system'
    )
ON CONFLICT (id) DO NOTHING;

-- Insert default multisig configuration
INSERT INTO multisig_configs (id, name, description, required_signatures, total_signers)
VALUES
    (
        'b0000000-0000-0000-0000-000000000001',
        'Standard 2-of-3',
        'Standard 2-of-3 multisig configuration',
        2,
        3
    ),
    (
        'b0000000-0000-0000-0000-000000000002',
        'Cold Storage 3-of-5',
        '3-of-5 multisig for cold storage transactions',
        3,
        5
    )
ON CONFLICT (id) DO NOTHING;
