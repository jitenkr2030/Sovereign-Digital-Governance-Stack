-- Mining Control Platform Database Schema
-- Migration: 001_initial_schema

-- Create UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Mining Operations Registry Table
CREATE TABLE IF NOT EXISTS mining_operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operator_name VARCHAR(255) NOT NULL,
    wallet_address VARCHAR(255) NOT NULL UNIQUE,
    current_hashrate DECIMAL(20, 4) DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING_REGISTRATION',
    quota_id UUID,
    location VARCHAR(500) NOT NULL,
    region VARCHAR(100) NOT NULL,
    machine_type VARCHAR(255) NOT NULL,
    registered_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_report_at TIMESTAMPTZ,
    violation_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    
    CONSTRAINT fk_quota FOREIGN KEY (quota_id) REFERENCES mining_quotas(id) ON DELETE SET NULL
);

-- Create index on wallet_address for fast lookups
CREATE INDEX IF NOT EXISTS idx_mining_operations_wallet ON mining_operations(wallet_address);
CREATE INDEX IF NOT EXISTS idx_mining_operations_status ON mining_operations(status);
CREATE INDEX IF NOT EXISTS idx_mining_operations_region ON mining_operations(region);
CREATE INDEX IF NOT EXISTS idx_mining_operations_registered_at ON mining_operations(registered_at DESC);

-- Hashrate Records Table (Time-series data)
CREATE TABLE IF NOT EXISTS hashrate_records (
    id BIGSERIAL PRIMARY KEY,
    operation_id UUID NOT NULL,
    hashrate DECIMAL(20, 4) NOT NULL,
    unit VARCHAR(20) NOT NULL DEFAULT 'TH/s',
    block_height BIGINT,
    timestamp TIMESTAMPTZ NOT NULL,
    submitted_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_operation FOREIGN KEY (operation_id) REFERENCES mining_operations(id) ON DELETE CASCADE
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_hashrate_records_operation_id ON hashrate_records(operation_id);
CREATE INDEX IF NOT EXISTS idx_hashrate_records_timestamp ON hashrate_records(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_hashrate_records_operation_timestamp ON hashrate_records(operation_id, timestamp DESC);

-- Note: For production, consider using TimescaleDB hypertable:
-- SELECT create_hypertable('hashrate_records', 'timestamp');

-- Mining Quotas Table
CREATE TABLE IF NOT EXISTS mining_quotas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation_id UUID NOT NULL,
    max_hashrate DECIMAL(20, 4) NOT NULL,
    quota_type VARCHAR(50) NOT NULL DEFAULT 'FIXED',
    valid_from TIMESTAMPTZ DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    region VARCHAR(100),
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_quota_operation FOREIGN KEY (operation_id) REFERENCES mining_operations(id) ON DELETE CASCADE,
    CONSTRAINT chk_max_hashrate_positive CHECK (max_hashrate > 0)
);

CREATE INDEX IF NOT EXISTS idx_mining_quotas_operation_id ON mining_quotas(operation_id);
CREATE INDEX IF NOT EXISTS idx_mining_quotas_validity ON mining_quotas(valid_from, valid_to);

-- Shutdown Commands Table
CREATE TABLE IF NOT EXISTS shutdown_commands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation_id UUID NOT NULL,
    command_type VARCHAR(50) NOT NULL,
    reason TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ISSUED',
    issued_at TIMESTAMPTZ DEFAULT NOW(),
    executed_at TIMESTAMPTZ,
    acked_at TIMESTAMPTZ,
    issued_by VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ,
    result TEXT,
    
    CONSTRAINT fk_shutdown_operation FOREIGN KEY (operation_id) REFERENCES mining_operations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_shutdown_commands_operation_id ON shutdown_commands(operation_id);
CREATE INDEX IF NOT EXISTS idx_shutdown_commands_status ON shutdown_commands(status);
CREATE INDEX IF NOT EXISTS idx_shutdown_commands_issued_at ON shutdown_commands(issued_at DESC);

-- Quota Violations Table
CREATE TABLE IF NOT EXISTS quota_violations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation_id UUID NOT NULL,
    quota_id UUID NOT NULL,
    reported_hashrate DECIMAL(20, 4) NOT NULL,
    max_hashrate DECIMAL(20, 4) NOT NULL,
    recorded_at TIMESTAMPTZ DEFAULT NOW(),
    consecutive BOOLEAN DEFAULT TRUE,
    automatic_shutdown BOOLEAN DEFAULT FALSE,
    
    CONSTRAINT fk_violation_operation FOREIGN KEY (operation_id) REFERENCES mining_operations(id) ON DELETE CASCADE,
    CONSTRAINT fk_violation_quota FOREIGN KEY (quota_id) REFERENCES mining_quotas(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_quota_violations_operation_id ON quota_violations(operation_id);
CREATE INDEX IF NOT EXISTS idx_quota_violations_recorded_at ON quota_violations(recorded_at DESC);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at column
CREATE TRIGGER update_mining_operations_updated_at
    BEFORE UPDATE ON mining_operations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mining_quotas_updated_at
    BEFORE UPDATE ON mining_quotas
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create enum type for operation status (optional, for clarity)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'operation_status') THEN
        CREATE TYPE operation_status AS ENUM (
            'ACTIVE',
            'SUSPENDED',
            'SHUTDOWN_ORDERED',
            'NON_COMPLIANT',
            'SHUTDOWN_EXECUTED',
            'PENDING_REGISTRATION'
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'command_status') THEN
        CREATE TYPE command_status AS ENUM (
            'ISSUED',
            'ACKNOWLEDGED',
            'EXECUTED',
            'FAILED'
        );
    END IF;
END $$;

-- Insert some sample data for testing
INSERT INTO mining_operations (operator_name, wallet_address, current_hashrate, status, location, region, machine_type)
VALUES 
    ('Test Mining Corp', '0x742d35Cc6634C0532925a3b844Bc9e7595f7547b', 150.5, 'ACTIVE', 'Data Center A, Nevada', 'US-WEST', 'Antminer S19 Pro'),
    ('Hash Power LLC', '0xdD2FD4581271e230360230F9337D5c0430Bf44C0', 75.25, 'ACTIVE', 'Facility B, Texas', 'US-CENTRAL', 'MicroBT M30S+'),
    ('Crypto Hash Inc', '0x2546BcD3c84621e976D8185a91A922aE77ECEc30', 200.0, 'SUSPENDED', 'Plant C, Georgia', 'US-EAST', 'Canaan Avalon 1246')
ON CONFLICT (wallet_address) DO NOTHING;

-- Add sample quota for first operation
INSERT INTO mining_quotas (operation_id, max_hashrate, quota_type, valid_from, region, priority)
SELECT id, 200.0, 'FIXED', NOW(), region, 1
FROM mining_operations WHERE wallet_address = '0x742d35Cc6634C0532925a3b844Bc9e7595f7547b'
ON CONFLICT DO NOTHING;
