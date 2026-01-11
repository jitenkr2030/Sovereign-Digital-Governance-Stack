-- Mining Control Service Database Migrations
-- Creates PostgreSQL schema with TimescaleDB for time-series data

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create enum types
DO $$ BEGIN
    CREATE TYPE pool_status AS ENUM ('PENDING', 'ACTIVE', 'SUSPENDED', 'REVOKED', 'EXPIRED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE machine_status AS ENUM ('PENDING', 'ACTIVE', 'MAINTENANCE', 'OFFLINE', 'DECOMMISSIONED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE machine_type AS ENUM ('ASIC', 'GPU', 'FPGA', 'CPU');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE violation_severity AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE violation_type AS ENUM (
        'EXCESS_ENERGY', 'UNREGISTERED_HARDWARE', 'BANNED_REGION',
        'ENERGY_SOURCE_VIOLATION', 'HASH_RATE_ANOMALY', 'LICENSE_EXPIRED',
        'CARBON_LIMIT_EXCEEDED', 'REPORTING_VIOLATION'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE violation_status AS ENUM ('OPEN', 'INVESTIGATING', 'RESOLVED', 'DISMISSED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE energy_source_type AS ENUM (
        'GRID', 'SOLAR', 'HYDRO', 'WIND', 'NUCLEAR', 'COAL', 'NATURAL_GAS', 'MIXED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create mining pools table
CREATE TABLE IF NOT EXISTS mining_pools (
    id UUID PRIMARY KEY,
    license_number VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    owner_entity_id UUID NOT NULL,
    owner_entity_name VARCHAR(200) NOT NULL,
    status pool_status NOT NULL DEFAULT 'PENDING',
    region_code VARCHAR(20) NOT NULL,
    max_allowed_energy_kw DECIMAL(15, 2) NOT NULL,
    current_energy_usage_kw DECIMAL(15, 2) NOT NULL DEFAULT 0,
    allowed_energy_sources energy_source_type[] NOT NULL,
    total_hash_rate_th DECIMAL(20, 2) NOT NULL DEFAULT 0,
    active_machine_count INTEGER NOT NULL DEFAULT 0,
    total_machine_count INTEGER NOT NULL DEFAULT 0,
    carbon_footprint_kg DECIMAL(20, 2) NOT NULL DEFAULT 0,
    license_issued_at TIMESTAMPTZ NOT NULL,
    license_expires_at TIMESTAMPTZ NOT NULL,
    last_inspection_at TIMESTAMPTZ,
    contact_email VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50) NOT NULL,
    facility_address TEXT NOT NULL,
    gps_coordinates JSONB,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create mining machines table
CREATE TABLE IF NOT EXISTS mining_machines (
    id UUID PRIMARY KEY,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    pool_id UUID NOT NULL REFERENCES mining_pools(id),
    model_type machine_type NOT NULL,
    manufacturer VARCHAR(100) NOT NULL,
    model_name VARCHAR(200) NOT NULL,
    hash_rate_spec_th DECIMAL(15, 2) NOT NULL,
    power_spec_watts DECIMAL(15, 2) NOT NULL,
    location_id VARCHAR(100) NOT NULL,
    location_description TEXT,
    gps_coordinates JSONB,
    installation_date TIMESTAMPTZ NOT NULL,
    last_maintenance_date TIMESTAMPTZ,
    status machine_status NOT NULL DEFAULT 'PENDING',
    is_active BOOLEAN NOT NULL DEFAULT false,
    current_hash_rate_th DECIMAL(15, 2) NOT NULL DEFAULT 0,
    current_power_watts DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_energy_kwh DECIMAL(20, 2) NOT NULL DEFAULT 0,
    uptime_hours DECIMAL(15, 2) NOT NULL DEFAULT 0,
    decommissioned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create compliance certificates table
CREATE TABLE IF NOT EXISTS compliance_certificates (
    id UUID PRIMARY KEY,
    pool_id UUID NOT NULL REFERENCES mining_pools(id),
    certificate_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'VALID',
    issue_date TIMESTAMPTZ NOT NULL,
    expiry_date TIMESTAMPTZ NOT NULL,
    inspection_date TIMESTAMPTZ,
    inspector_id UUID,
    energy_rating VARCHAR(10),
    carbon_rating VARCHAR(10),
    compliance_score DECIMAL(5, 2) NOT NULL,
    terms_conditions TEXT[],
    restrictions TEXT[],
    digital_signature TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create energy consumption hypertable (TimescaleDB)
CREATE TABLE IF NOT EXISTS energy_consumption_logs (
    time TIMESTAMPTZ NOT NULL,
    pool_id UUID NOT NULL,
    duration_minutes INTEGER NOT NULL,
    power_usage_kwh DECIMAL(15, 4) NOT NULL,
    peak_power_kw DECIMAL(15, 4) NOT NULL,
    average_power_kw DECIMAL(15, 4) NOT NULL,
    energy_source energy_source_type NOT NULL,
    source_mix_percentage JSONB,
    carbon_emissions_kg DECIMAL(15, 4) NOT NULL,
    grid_region_code VARCHAR(20) NOT NULL,
    temperature_c DECIMAL(5, 2),
    humidity_percent DECIMAL(5, 2),
    data_quality VARCHAR(20) NOT NULL DEFAULT 'GOOD',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checksum VARCHAR(64)
);

SELECT create_hypertable('energy_consumption_logs', 'time', 
    chunk_time_interval => INTERVAL '7 days');

-- Create index on pool_id for energy queries
CREATE INDEX IF NOT EXISTS idx_energy_pool_id 
ON energy_consumption_logs (pool_id, time DESC);

-- Create hash rate hypertable (TimescaleDB)
CREATE TABLE IF NOT EXISTS hashrate_metrics (
    time TIMESTAMPTZ NOT NULL,
    pool_id UUID NOT NULL,
    reported_hash_rate_th DECIMAL(20, 4) NOT NULL,
    calculated_hash_rate_th DECIMAL(20, 4) NOT NULL,
    active_worker_count INTEGER NOT NULL DEFAULT 0,
    stale_worker_count INTEGER NOT NULL DEFAULT 0,
    total_shares DECIMAL(20, 4) NOT NULL DEFAULT 0,
    accepted_shares DECIMAL(20, 4) NOT NULL DEFAULT 0,
    rejected_shares DECIMAL(20, 4) NOT NULL DEFAULT 0,
    difficulty DECIMAL(20, 4) NOT NULL DEFAULT 0,
    pool_fee_percent DECIMAL(5, 2) NOT NULL DEFAULT 1.5,
    data_quality VARCHAR(20) NOT NULL DEFAULT 'GOOD',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    checksum VARCHAR(64)
);

SELECT create_hypertable('hashrate_metrics', 'time',
    chunk_time_interval => INTERVAL '7 days');

-- Create index on pool_id for hashrate queries
CREATE INDEX IF NOT EXISTS idx_hashrate_pool_id 
ON hashrate_metrics (pool_id, time DESC);

-- Create compliance violations table
CREATE TABLE IF NOT EXISTS compliance_violations (
    id UUID PRIMARY KEY,
    pool_id UUID NOT NULL REFERENCES mining_pools(id),
    violation_type violation_type NOT NULL,
    severity violation_severity NOT NULL,
    status violation_status NOT NULL DEFAULT 'OPEN',
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    details JSONB,
    evidence TEXT[],
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    resolution TEXT,
    assigned_to UUID,
    alerts_sent INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create violation notes table
CREATE TABLE IF NOT EXISTS violation_notes (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL REFERENCES compliance_violations(id),
    author_id UUID NOT NULL,
    author_name VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_pools_status ON mining_pools(status);
CREATE INDEX IF NOT EXISTS idx_pools_region ON mining_pools(region_code);
CREATE INDEX IF NOT EXISTS idx_pools_owner ON mining_pools(owner_entity_id);
CREATE INDEX IF NOT EXISTS idx_machines_pool_id ON mining_machines(pool_id);
CREATE INDEX IF NOT EXISTS idx_machines_status ON mining_machines(status);
CREATE INDEX IF NOT EXISTS idx_machines_serial ON mining_machines(serial_number);
CREATE INDEX IF NOT EXISTS idx_violations_pool_id ON compliance_violations(pool_id);
CREATE INDEX IF NOT EXISTS idx_violations_status ON compliance_violations(status);
CREATE INDEX IF NOT EXISTS idx_violations_type ON compliance_violations(violation_type);
CREATE INDEX IF NOT EXISTS idx_violations_severity ON compliance_violations(severity);
CREATE INDEX IF NOT EXISTS idx_violations_detected ON compliance_violations(detected_at DESC);

-- Create view for dashboard statistics
CREATE OR REPLACE VIEW dashboard_stats AS
SELECT 
    COUNT(*) FILTER (WHERE status = 'ACTIVE') as active_pools,
    COUNT(*) FILTER (WHERE status = 'SUSPENDED') as suspended_pools,
    COUNT(*) FILTER (WHERE status = 'PENDING') as pending_pools,
    COUNT(*) FILTER (WHERE status = 'EXPIRED' OR status = 'REVOKED') as inactive_pools,
    COUNT(*) as total_pools,
    SUM(total_machine_count) as total_machines,
    SUM(active_machine_count) as active_machines,
    SUM(total_hash_rate_th) as total_hash_rate_th,
    SUM(current_energy_usage_kw) as total_energy_kw,
    SUM(carbon_footprint_kg) as total_carbon_kg,
    COUNT(*) FILTER (WHERE id IN (
        SELECT pool_id FROM compliance_violations WHERE status = 'OPEN'
    )) as pools_with_violations,
    COUNT(*) FILTER (WHERE severity = 'CRITICAL' AND status = 'OPEN') as critical_violations
FROM mining_pools;

-- Create function to update timestamp on row update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to auto-update timestamps
CREATE TRIGGER update_pools_updated_at BEFORE UPDATE ON mining_pools
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_machines_updated_at BEFORE UPDATE ON mining_machines
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_certificates_updated_at BEFORE UPDATE ON compliance_certificates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_violations_updated_at BEFORE UPDATE ON compliance_violations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions
GRANT SELECT ON ALL TABLES IN SCHEMA public TO csic_reader;
GRANT ALL ON ALL TABLES IN SCHEMA public TO csic_mining;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO csic_mining;
