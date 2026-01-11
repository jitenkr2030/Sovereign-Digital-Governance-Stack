-- Energy Integration Service Database Migrations
-- This file contains the initial schema for energy profiles, transaction footprints, and carbon offset certificates

-- Create energy_profiles table
-- Stores energy characteristics of blockchain networks
CREATE TABLE IF NOT EXISTS energy_profiles (
    id UUID PRIMARY KEY,
    chain_id VARCHAR(64) NOT NULL UNIQUE,
    chain_name VARCHAR(128) NOT NULL,
    consensus_type VARCHAR(32) NOT NULL,
    avg_kwh_per_tx DECIMAL(20, 10) NOT NULL,
    base_carbon_grams DECIMAL(20, 4) NOT NULL,
    energy_source VARCHAR(32) NOT NULL,
    carbon_intensity DECIMAL(10, 4) NOT NULL DEFAULT 0.5,
    node_locations JSONB DEFAULT '[]'::jsonb,
    network_hash_rate DECIMAL(20, 4) DEFAULT 0,
    transactions_per_second DECIMAL(20, 4) DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for energy_profiles
CREATE INDEX IF NOT EXISTS idx_energy_profiles_chain_id ON energy_profiles(chain_id);
CREATE INDEX IF NOT EXISTS idx_energy_profiles_consensus_type ON energy_profiles(consensus_type);
CREATE INDEX IF NOT EXISTS idx_energy_profiles_is_active ON energy_profiles(is_active);

-- Create transaction_footprints table
-- Stores calculated carbon footprints for blockchain transactions
CREATE TABLE IF NOT EXISTS transaction_footprints (
    id UUID PRIMARY KEY,
    transaction_hash VARCHAR(128) NOT NULL,
    chain_id VARCHAR(64) NOT NULL,
    block_number BIGINT,
    energy_value DECIMAL(20, 10) NOT NULL,
    energy_unit VARCHAR(8) NOT NULL DEFAULT 'kWh',
    energy_source VARCHAR(32) NOT NULL,
    carbon_value DECIMAL(20, 4) NOT NULL,
    carbon_unit VARCHAR(4) NOT NULL DEFAULT 'g',
    factor_id VARCHAR(64),
    offset_status VARCHAR(16) NOT NULL DEFAULT 'none',
    certificate_id UUID,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for transaction_footprints
CREATE INDEX IF NOT EXISTS idx_transaction_footprints_tx_hash ON transaction_footprints(transaction_hash);
CREATE INDEX IF NOT EXISTS idx_transaction_footprints_chain_id ON transaction_footprints(chain_id);
CREATE INDEX IF NOT EXISTS idx_transaction_footprints_timestamp ON transaction_footprints(timestamp);
CREATE INDEX IF NOT EXISTS idx_transaction_footprints_offset_status ON transaction_footprints(offset_status);
CREATE INDEX IF NOT EXISTS idx_transaction_footprints_carbon_value ON transaction_footprints(carbon_value);

-- Create offset_certificates table
-- Stores carbon offset certificates (RECs, carbon credits, etc.)
CREATE TABLE IF NOT EXISTS offset_certificates (
    id UUID PRIMARY KEY,
    certificate_number VARCHAR(64) NOT NULL UNIQUE,
    energy_source VARCHAR(32) NOT NULL,
    energy_amount DECIMAL(20, 6) NOT NULL,
    energy_unit VARCHAR(8) NOT NULL DEFAULT 'MWh',
    carbon_value DECIMAL(20, 4) NOT NULL,
    carbon_unit VARCHAR(4) NOT NULL DEFAULT 'g',
    registry VARCHAR(64) NOT NULL,
    issue_date TIMESTAMP WITH TIME ZONE NOT NULL,
    expiry_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for offset_certificates
CREATE INDEX IF NOT EXISTS idx_offset_certificates_number ON offset_certificates(certificate_number);
CREATE INDEX IF NOT EXISTS idx_offset_certificates_status ON offset_certificates(status);
CREATE INDEX IF NOT EXISTS idx_offset_certificates_registry ON offset_certificates(registry);
CREATE INDEX IF NOT EXISTS idx_offset_certificates_expiry ON offset_certificates(expiry_date);

-- Insert default energy profiles for major blockchain networks
INSERT INTO energy_profiles (id, chain_id, chain_name, consensus_type, avg_kwh_per_tx, base_carbon_grams, energy_source, carbon_intensity, node_locations, is_active)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'ethereum', 'Ethereum', 'proof_of_work', 0.038, 19.0, 'mixed', 0.5, '["US", "EU", "Asia"]', true),
    ('22222222-2222-2222-2222-222222222222', 'ethereum-pos', 'Ethereum Proof of Stake', 'proof_of_stake', 0.0006, 0.3, 'mixed', 0.5, '["US", "EU", "Asia"]', true),
    ('33333333-3333-3333-3333-333333333333', 'bitcoin', 'Bitcoin', 'proof_of_work', 850.0, 425000.0, 'fossil_fuel', 0.9, '["US", "China", "Russia"]', true),
    ('44444444-4444-4444-4444-444444444444', 'polygon', 'Polygon', 'proof_of_stake', 0.0009, 0.45, 'mixed', 0.5, '["US", "EU"]', true),
    ('55555555-5555-5555-5555-555555555555', 'solana', 'Solana', 'proof_of_history', 0.0007, 0.35, 'mixed', 0.5, '["US", "EU", "Asia"]', true)
ON CONFLICT (chain_id) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for automatic updated_at updates
DROP TRIGGER IF EXISTS update_energy_profiles_updated_at ON energy_profiles;
CREATE TRIGGER update_energy_profiles_updated_at
    BEFORE UPDATE ON energy_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
