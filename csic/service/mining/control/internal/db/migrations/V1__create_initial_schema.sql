-- Migration V1: Create Initial Schema for Mining Control Module
-- This migration creates all core tables for managing mining operations,
-- energy consumption tracking, and compliance monitoring.
-- Direction: UP

-- Create ENUM types for status fields
CREATE TYPE pool_status AS ENUM ('active', 'inactive', 'suspended', 'maintenance');
CREATE TYPE machine_status AS ENUM ('online', 'offline', 'mining', 'idle', 'error', 'banned');
CREATE TYPE violation_severity AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE violation_type AS ENUM ('energy_limit_exceeded', 'hashrate_mismatch', 'unauthorized_operation', 'geofence_violation', 'power_factor_violation', 'peak_hour_violation');

-- Create mining pools table
-- Stores information about registered mining pools/operations
CREATE TABLE IF NOT EXISTS mining_pools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    operator_id UUID NOT NULL,
    location VARCHAR(500) NOT NULL,
    geographic_zone VARCHAR(100),
    total_hashrate DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_power_consumption DECIMAL(15, 2) NOT NULL DEFAULT 0,
    max_power_allocation DECIMAL(15, 2) NOT NULL,
    energy_limit_watts DECIMAL(15, 2) NOT NULL,
    status pool_status NOT NULL DEFAULT 'active',
    registration_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_inspection_date TIMESTAMP WITH TIME ZONE,
    compliance_score DECIMAL(5, 2) NOT NULL DEFAULT 100.00,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for mining pools
CREATE INDEX IF NOT EXISTS idx_mining_pools_operator_id ON mining_pools(operator_id);
CREATE INDEX IF NOT EXISTS idx_mining_pools_status ON mining_pools(status);
CREATE INDEX IF NOT EXISTS idx_mining_pools_geographic_zone ON mining_pools(geographic_zone);
CREATE INDEX IF NOT EXISTS idx_mining_pools_location ON mining_pools USING gin(to_tsvector('english', location));

-- Create mining machines table
-- Stores information about individual mining machines within pools
CREATE TABLE IF NOT EXISTS mining_machines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES mining_pools(id) ON DELETE CASCADE,
    machine_identifier VARCHAR(100) NOT NULL,
    machine_type VARCHAR(100) NOT NULL,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    serial_number VARCHAR(100),
    hashrate DECIMAL(20, 2) NOT NULL,
    power_consumption DECIMAL(15, 2) NOT NULL,
    status machine_status NOT NULL DEFAULT 'online',
    installation_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_maintenance_date TIMESTAMP WITH TIME ZONE,
    firmware_version VARCHAR(50),
    ip_address INET,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_machine_identifier UNIQUE (machine_identifier)
);

-- Create indexes for mining machines
CREATE INDEX IF NOT EXISTS idx_mining_machines_pool_id ON mining_machines(pool_id);
CREATE INDEX IF NOT EXISTS idx_mining_machines_status ON mining_machines(status);
CREATE INDEX IF NOT EXISTS idx_mining_machines_machine_type ON mining_machines(machine_type);
CREATE INDEX IF NOT EXISTS idx_mining_machines_hashrate ON mining_machines(hashrate DESC);

-- Create energy consumption logs table
-- Stores time-series data for energy consumption tracking (will be converted to hypertable)
CREATE TABLE IF NOT EXISTS energy_consumption_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES mining_pools(id) ON DELETE CASCADE,
    machine_id UUID REFERENCES mining_machines(id) ON DELETE SET NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    power_consumption_watts DECIMAL(15, 2) NOT NULL,
    energy_used_wh DECIMAL(15, 4) NOT NULL,
    voltage DECIMAL(10, 2),
    current DECIMAL(10, 2),
    power_factor DECIMAL(5, 3),
    frequency DECIMAL(8, 4),
    temperature DECIMAL(6, 2),
    source VARCHAR(50) NOT NULL DEFAULT 'meter',
    raw_data JSONB
);

-- Create indexes for energy consumption logs
CREATE INDEX IF NOT EXISTS idx_energy_logs_pool_id ON energy_consumption_logs(pool_id);
CREATE INDEX IF NOT EXISTS idx_energy_logs_machine_id ON energy_consumption_logs(machine_id);
CREATE INDEX IF NOT EXISTS idx_energy_logs_timestamp ON energy_consumption_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_energy_logs_pool_timestamp ON energy_consumption_logs(pool_id, timestamp DESC);

-- Create compliance violations table
-- Stores records of energy policy and regulation violations
CREATE TABLE IF NOT EXISTS compliance_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES mining_pools(id) ON DELETE CASCADE,
    machine_id UUID REFERENCES mining_machines(id) ON DELETE SET NULL,
    violation_type violation_type NOT NULL,
    severity violation_severity NOT NULL,
    description TEXT NOT NULL,
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    assigned_inspector UUID,
    evidence_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for compliance violations
CREATE INDEX IF NOT EXISTS idx_violations_pool_id ON compliance_violations(pool_id);
CREATE INDEX IF NOT EXISTS idx_violations_status ON compliance_violations(status);
CREATE INDEX IF NOT EXISTS idx_violations_severity ON compliance_violations(severity);
CREATE INDEX IF NOT EXISTS idx_violations_type ON compliance_violations(violation_type);
CREATE INDEX IF NOT EXISTS idx_violations_detected_at ON compliance_violations(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_violations_pool_status ON compliance_violations(pool_id, status);

-- Create compliance records table
-- Stores periodic compliance check results for mining operations
CREATE TABLE IF NOT EXISTS compliance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES mining_pools(id) ON DELETE CASCADE,
    check_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    compliance_score DECIMAL(5, 2) NOT NULL,
    energy_efficiency_rating VARCHAR(20),
    power_factor_rating VARCHAR(20),
    peak_usage_rating VARCHAR(20),
    geofence_compliance BOOLEAN NOT NULL DEFAULT TRUE,
    inspection_notes TEXT,
    inspector_id UUID NOT NULL,
    check_type VARCHAR(50) NOT NULL DEFAULT 'routine',
    findings JSONB,
    recommendations JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for compliance records
CREATE INDEX IF NOT EXISTS idx_compliance_records_pool_id ON compliance_records(pool_id);
CREATE INDEX IF NOT EXISTS idx_compliance_records_check_date ON compliance_records(check_date DESC);
CREATE INDEX IF NOT EXISTS idx_compliance_records_inspector ON compliance_records(inspector_id);
CREATE INDEX IF NOT EXISTS idx_compliance_records_pool_date ON compliance_records(pool_id, check_date DESC);

-- Create energy alerts table
-- Stores real-time alerts for energy threshold violations
CREATE TABLE IF NOT EXISTS energy_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES mining_pools(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    threshold_type VARCHAR(50) NOT NULL,
    threshold_value DECIMAL(15, 2) NOT NULL,
    actual_value DECIMAL(15, 2) NOT NULL,
    message TEXT NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for energy alerts
CREATE INDEX IF NOT EXISTS idx_energy_alerts_pool_id ON energy_alerts(pool_id);
CREATE INDEX IF NOT EXISTS idx_energy_alerts_status ON energy_alerts(status);
CREATE INDEX IF NOT EXISTS idx_energy_alerts_severity ON energy_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_energy_alerts_created_at ON energy_alerts(created_at DESC);

-- Create power schedule table
-- Stores scheduled power on/off times for mining operations
CREATE TABLE IF NOT EXISTS power_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pool_id UUID NOT NULL REFERENCES mining_pools(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    power_limit_watts DECIMAL(15, 2),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_time_range CHECK (start_time < end_time)
);

-- Create indexes for power schedules
CREATE INDEX IF NOT EXISTS idx_power_schedules_pool_id ON power_schedules(pool_id);
CREATE INDEX IF NOT EXISTS idx_power_schedules_active ON power_schedules(pool_id, is_active);

-- Create geographic zones table
-- Stores defined geographic zones with specific energy regulations
CREATE TABLE IF NOT EXISTS geographic_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    max_power_density_watts_per_sqm DECIMAL(10, 2),
    peak_hour_restrictions BOOLEAN NOT NULL DEFAULT FALSE,
    peak_hour_start TIME,
    peak_hour_end TIME,
    max_peak_power_watts DECIMAL(15, 2),
    noise_restrictions BOOLEAN NOT NULL DEFAULT FALSE,
    max_noise_level_db INTEGER,
    special_regulations JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for geographic zones
CREATE INDEX IF NOT EXISTS idx_geo_zones_name ON geographic_zones(name);

-- Create audit log table
-- Stores audit trail for all significant actions in the system
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id UUID,
    actor_type VARCHAR(50),
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for audit logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to relevant tables
DO $$
DECLARE
    table_name TEXT;
BEGIN
    FOR table_name IN 
        SELECT table_name 
        FROM information_schema.columns 
        WHERE column_name = 'updated_at' 
        AND table_schema = 'public'
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS update_%I_updated_at ON %I', table_name, table_name);
        EXECUTE format('CREATE TRIGGER update_%I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', table_name, table_name);
    END LOOP;
END;
$$ language 'plpgsql';

-- Direction: DOWN
-- Drop all created tables and types in reverse order
-- DROP TABLE IF EXISTS audit_logs CASCADE;
-- DROP TABLE IF EXISTS power_schedules CASCADE;
-- DROP TABLE IF EXISTS energy_alerts CASCADE;
-- DROP TABLE IF EXISTS compliance_records CASCADE;
-- DROP TABLE IF EXISTS compliance_violations CASCADE;
-- DROP TABLE IF EXISTS energy_consumption_logs CASCADE;
-- DROP TABLE IF EXISTS mining_machines CASCADE;
-- DROP TABLE IF EXISTS mining_pools CASCADE;
-- DROP TABLE IF EXISTS geographic_zones CASCADE;
-- DROP TYPE IF EXISTS pool_status CASCADE;
-- DROP TYPE IF EXISTS machine_status CASCADE;
-- DROP TYPE IF EXISTS violation_severity CASCADE;
-- DROP TYPE IF EXISTS violation_type CASCADE;
