-- =====================================================
-- Regulatory Reporting Service - Database Init Script
-- =====================================================
-- This script runs on container first start to initialize
-- the PostgreSQL database with required schemas and data
-- =====================================================

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create custom types
DO $$ BEGIN
    CREATE TYPE report_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE report_format AS ENUM ('pdf', 'csv', 'xlsx', 'json', 'xml');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE schedule_frequency AS ENUM ('daily', 'weekly', 'monthly', 'quarterly', 'annually', 'custom');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE template_type AS ENUM ('regulatory', 'compliance', 'audit', 'custom');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create enum types for job status if not exists
DO $$ BEGIN
    CREATE TYPE job_status AS ENUM ('waiting', 'active', 'completed', 'failed', 'delayed', 'stalled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create reports table
CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_type VARCHAR(100) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    format report_format NOT NULL DEFAULT 'pdf',
    status report_status NOT NULL DEFAULT 'pending',
    parameters JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    file_path VARCHAR(1000),
    file_size BIGINT,
    checksum VARCHAR(64),
    error_message TEXT,
    generated_by VARCHAR(255),
    entity_id UUID,
    entity_type VARCHAR(100),
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for reports
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_type ON reports(report_type);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reports_entity ON reports(entity_id, entity_type);
CREATE INDEX IF NOT EXISTS idx_reports_status_created ON reports(status, created_at);
CREATE INDEX IF NOT EXISTS idx_reports_parameters ON reports USING GIN (parameters);

-- Create schedules table
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    report_type VARCHAR(100) NOT NULL,
    format report_format NOT NULL DEFAULT 'pdf',
    frequency schedule_frequency NOT NULL,
    cron_expression VARCHAR(100),
    parameters JSONB DEFAULT '{}',
    template_id UUID,
    recipients JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_status VARCHAR(50),
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for schedules
CREATE INDEX IF NOT EXISTS idx_schedules_active ON schedules(is_active, next_run_at);
CREATE INDEX IF NOT EXISTS idx_schedules_frequency ON schedules(frequency);

-- Create schedule_runs table
CREATE TABLE IF NOT EXISTS schedule_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    report_id UUID REFERENCES reports(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for schedule_runs
CREATE INDEX IF NOT EXISTS idx_schedule_runs_schedule ON schedule_runs(schedule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_runs_status ON schedule_runs(status);
CREATE INDEX IF NOT EXISTS idx_schedule_runs_triggered ON schedule_runs(triggered_at DESC);

-- Create templates table
CREATE TABLE IF NOT EXISTS templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template_type template_type NOT NULL,
    version INTEGER DEFAULT 1,
    content TEXT NOT NULL,
    parameters_schema JSONB DEFAULT '{}',
    default_parameters JSONB DEFAULT '{}',
    output_format report_format DEFAULT 'pdf',
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for templates
CREATE INDEX IF NOT EXISTS idx_templates_type ON templates(template_type);
CREATE INDEX IF NOT EXISTS idx_templates_active ON templates(is_active, is_system);

-- Create report_history table for audit trail
CREATE TABLE IF NOT EXISTS report_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    performed_by UUID,
    performed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT
);

-- Create indexes for report_history
CREATE INDEX IF NOT EXISTS idx_report_history_report ON report_history(report_id);
CREATE INDEX IF NOT EXISTS idx_report_history_action ON report_history(action);
CREATE INDEX IF NOT EXISTS idx_report_history_performed ON report_history(performed_at DESC);

-- Create users table for basic authentication
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Create api_keys table for service authentication
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,
    permissions JSONB DEFAULT '[]',
    rate_limit INTEGER DEFAULT 1000,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for api_keys
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active, expires_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);

-- Create audit_logs table for compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    api_key_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    request_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit_logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);

-- Create view for report statistics
CREATE OR REPLACE VIEW report_statistics AS
SELECT
    report_type,
    status,
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as count,
    SUM(file_size) as total_size
FROM reports
GROUP BY report_type, status, DATE_TRUNC('day', created_at);

-- Create function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN SELECT table_name FROM information_schema.columns
             WHERE column_name = 'updated_at'
             AND table_schema = 'public'
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS update_%I_updated_at ON %I', t, t);
        EXECUTE format('CREATE TRIGGER update_%I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END;
$$;

-- Insert default system templates
INSERT INTO templates (id, name, description, template_type, content, is_system, is_active, created_at)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'Monthly Compliance Report', 'Standard monthly regulatory compliance report', 'regulatory', '<html>Monthly compliance report template</html>', true, true, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440002', 'Quarterly Audit Summary', 'Quarterly audit summary report', 'audit', '<html>Quarterly audit summary template</html>', true, true, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440003', 'Annual Regulatory Report', 'Comprehensive annual regulatory filing', 'regulatory', '<html>Annual regulatory report template</html>', true, true, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Insert default admin user (password: admin123 - should be changed in production)
INSERT INTO users (id, email, password_hash, name, role, is_active, created_at)
VALUES ('550e8400-e29b-41d4-a716-446655440010', 'admin@csic.gov', '$2b$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'System Administrator', 'admin', true, CURRENT_TIMESTAMP)
ON CONFLICT (email) DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO reporting_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO reporting_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO reporting_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO reporting_admin;
