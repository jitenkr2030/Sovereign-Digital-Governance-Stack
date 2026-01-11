-- +goose Up
-- +goose StatementBegin

-- Create regulatory reporting tables for CSIC Platform
-- This migration establishes the schema for report templates, generated reports,
-- schedules, regulators, and secure export audit logs.

-- Enum types
CREATE TYPE report_type AS ENUM (
    'periodic',
    'suspicious_activity',
    'compliance',
    'audit',
    'risk_assessment',
    'enforcement_action',
    'regulatory_filing'
);

CREATE TYPE report_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'archived'
);

CREATE TYPE report_format AS ENUM (
    'pdf',
    'csv',
    'xlsx',
    'json',
    'xbrl',
    'xml'
);

CREATE TYPE report_frequency AS ENUM (
    'daily',
    'weekly',
    'monthly',
    'quarterly',
    'annually',
    'on_demand'
);

CREATE TYPE regulator_type AS ENUM (
    'central_bank',
    'securities_commission',
    'financial_conduct_authority',
    'aml_supervisor',
    'tax_authority',
    'data_protection',
    'other'
);

CREATE TYPE export_type AS ENUM (
    'report',
    'data',
    'audit_trail',
    'compliance_data',
    'custom'
);

CREATE TYPE export_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE encryption_type AS ENUM (
    'none',
    'aes256',
    'pgp'
);

-- Regulators table
CREATE TABLE regulators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    official_name VARCHAR(255),
    type regulator_type NOT NULL,
    jurisdiction VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    website VARCHAR(500),
    submission_email VARCHAR(255),
    address_line_1 VARCHAR(255),
    address_line_2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    api_endpoint VARCHAR(500),
    api_auth_type VARCHAR(50),
    api_config JSONB,
    supported_types report_type[],
    supported_formats report_format[],
    xbrl_taxonomy VARCHAR(255),
    filing_requirements TEXT,
    validation_rules TEXT,
    time_zone VARCHAR(50) DEFAULT 'UTC',
    business_hours_start VARCHAR(5) DEFAULT '09:00',
    business_hours_end VARCHAR(5) DEFAULT '18:00',
    working_days INTEGER[],
    holidays TEXT[],
    view_config JSONB,
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    verification_date TIMESTAMP WITH TIME ZONE,
    verification_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

CREATE INDEX idx_regulators_name ON regulators(name);
CREATE INDEX idx_regulators_type ON regulators(type);
CREATE INDEX idx_regulators_jurisdiction ON regulators(jurisdiction);
CREATE INDEX idx_regulators_active ON regulators(is_active);

-- Report templates table
CREATE TABLE report_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    report_type report_type NOT NULL,
    regulator_id UUID REFERENCES regulators(id),
    version VARCHAR(50) DEFAULT '1.0.0',
    template_schema JSONB NOT NULL,
    generation_rules JSONB NOT NULL,
    output_formats report_format[],
    frequency report_frequency DEFAULT 'on_demand',
    cron_schedule VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_templates_regulator ON report_templates(regulator_id);
CREATE INDEX idx_templates_type ON report_templates(report_type);
CREATE INDEX idx_templates_active ON report_templates(is_active);

-- Generated reports table
CREATE TABLE generated_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES report_templates(id),
    template_name VARCHAR(255) NOT NULL,
    report_type report_type NOT NULL,
    regulator_id UUID REFERENCES regulators(id),
    regulator_name VARCHAR(255),
    status report_status DEFAULT 'pending',
    version VARCHAR(50) DEFAULT '1.0.0',
    title VARCHAR(500) NOT NULL,
    description TEXT,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    report_date DATE NOT NULL,
    generated_at TIMESTAMP WITH TIME ZONE,
    generated_by VARCHAR(255),
    generation_method VARCHAR(50) DEFAULT 'manual',
    triggered_by VARCHAR(255),
    output_formats report_format[],
    files JSONB,
    total_size BIGINT DEFAULT 0,
    checksum VARCHAR(64),
    data_snapshot BYTEA,
    processing_time BIGINT,
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    is_validated BOOLEAN DEFAULT false,
    validated_at TIMESTAMP WITH TIME ZONE,
    validated_by VARCHAR(255),
    is_approved BOOLEAN DEFAULT false,
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by VARCHAR(255),
    submission_status VARCHAR(50),
    submitted_at TIMESTAMP WITH TIME ZONE,
    submitted_by VARCHAR(255),
    submission_ref VARCHAR(255),
    acknowledgment_ref VARCHAR(255),
    is_archived BOOLEAN DEFAULT false,
    archived_at TIMESTAMP WITH TIME ZONE,
    retention_period INTEGER DEFAULT 365,
    custom_fields JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

CREATE INDEX idx_reports_template ON generated_reports(template_id);
CREATE INDEX idx_reports_regulator ON generated_reports(regulator_id);
CREATE INDEX idx_reports_type ON generated_reports(report_type);
CREATE INDEX idx_reports_status ON generated_reports(status);
CREATE INDEX idx_reports_period ON generated_reports(period_start, period_end);
CREATE INDEX idx_reports_date ON generated_reports(report_date);
CREATE INDEX idx_reports_archived ON generated_reports(is_archived);

-- Report schedules table
CREATE TABLE report_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID REFERENCES report_templates(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    frequency report_frequency NOT NULL,
    cron_expression VARCHAR(100) NOT NULL,
    time_zone VARCHAR(50) DEFAULT 'UTC',
    start_date DATE,
    end_date DATE,
    skip_weekends BOOLEAN DEFAULT false,
    skip_holidays BOOLEAN DEFAULT false,
    priority INTEGER DEFAULT 5,
    output_formats report_format[],
    notify_on_success BOOLEAN DEFAULT true,
    notify_on_failure BOOLEAN DEFAULT true,
    recipients TEXT[],
    is_active BOOLEAN DEFAULT true,
    last_run_at TIMESTAMP WITH TIME ZONE,
    next_run_at TIMESTAMP WITH TIME ZONE,
    run_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_schedules_template ON report_schedules(template_id);
CREATE INDEX idx_schedules_active ON report_schedules(is_active);
CREATE INDEX idx_schedules_next_run ON report_schedules(next_run_at);

-- Report queue table
CREATE TABLE report_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES generated_reports(id),
    template_id UUID REFERENCES report_templates(id),
    priority INTEGER DEFAULT 5,
    status VARCHAR(20) DEFAULT 'queued',
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    worker_id VARCHAR(100),
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_queue_status ON report_queue(status);
CREATE INDEX idx_queue_scheduled ON report_queue(scheduled_at);
CREATE INDEX idx_queue_priority ON report_queue(priority);

-- Report history table
CREATE TABLE report_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES generated_reports(id),
    action VARCHAR(50) NOT NULL,
    previous_value TEXT,
    new_value TEXT,
    changed_by VARCHAR(255),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reason TEXT,
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX idx_history_report ON report_history(report_id);
CREATE INDEX idx_history_action ON report_history(action);
CREATE INDEX idx_history_date ON report_history(changed_at);

-- Export logs table
CREATE TABLE export_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES generated_reports(id),
    export_type export_type NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    user_name VARCHAR(255),
    user_email VARCHAR(255),
    user_role VARCHAR(100),
    user_organization VARCHAR(255),
    regulator_id UUID REFERENCES regulators(id),
    regulator_name VARCHAR(255),
    export_format report_format NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    file_path VARCHAR(1000),
    file_size BIGINT,
    checksum VARCHAR(64),
    checksum_type VARCHAR(20) DEFAULT 'sha256',
    encryption_type encryption_type DEFAULT 'none',
    encrypted_key_id VARCHAR(255),
    encrypted_at TIMESTAMP WITH TIME ZONE,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    downloaded_at TIMESTAMP WITH TIME ZONE,
    download_count INTEGER DEFAULT 0,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    is_authorized BOOLEAN DEFAULT true,
    authorized_by VARCHAR(255),
    authorization_time TIMESTAMP WITH TIME ZONE,
    authorization_ref VARCHAR(255),
    status export_status DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    failure_reason TEXT,
    error_code VARCHAR(50),
    retry_count INTEGER DEFAULT 0,
    data_scope JSONB,
    filter_criteria TEXT,
    record_count INTEGER DEFAULT 0,
    contains_sensitive BOOLEAN DEFAULT false,
    contains_pii BOOLEAN DEFAULT false,
    redaction_applied BOOLEAN DEFAULT false,
    audit_trail_id UUID,
    metadata JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_exports_user ON export_logs(user_id);
CREATE INDEX idx_exports_report ON export_logs(report_id);
CREATE INDEX idx_exports_regulator ON export_logs(regulator_id);
CREATE INDEX idx_exports_status ON export_logs(status);
CREATE INDEX idx_exports_requested ON export_logs(requested_at);
CREATE INDEX idx_exports_downloaded ON export_logs(downloaded_at);

-- Regulator users table
CREATE TABLE regulator_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    regulator_id UUID REFERENCES regulators(id),
    user_id VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    permissions TEXT[],
    last_login_at TIMESTAMP WITH TIME ZONE,
    login_count INTEGER DEFAULT 0,
    last_activity_at TIMESTAMP WITH TIME ZONE,
    preferences JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255)
);

CREATE INDEX idx_regulator_users_user ON regulator_users(user_id);
CREATE INDEX idx_regulator_users_regulator ON regulator_users(regulator_id);
CREATE INDEX idx_regulator_users_email ON regulator_users(email);

-- Regulator API keys table
CREATE TABLE regulator_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    regulator_id UUID REFERENCES regulators(id),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    permissions TEXT[],
    rate_limit INTEGER DEFAULT 1000,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    usage_count BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by VARCHAR(255)
);

CREATE INDEX idx_api_keys_regulator ON regulator_api_keys(regulator_id);
CREATE INDEX idx_api_keys_hash ON regulator_api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON regulator_api_keys(is_active);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name 
        FROM information_schema.columns 
        WHERE column_name = 'updated_at' 
        AND table_schema = 'public'
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS update_%I_updated_at ON %I', t, t);
        EXECUTE format('CREATE TRIGGER update_%I_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()', t, t);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop tables in reverse order
DROP TABLE IF EXISTS regulator_api_keys CASCADE;
DROP TABLE IF EXISTS regulator_users CASCADE;
DROP TABLE IF EXISTS export_logs CASCADE;
DROP TABLE IF EXISTS report_history CASCADE;
DROP TABLE IF EXISTS report_queue CASCADE;
DROP TABLE IF EXISTS report_schedules CASCADE;
DROP TABLE IF EXISTS generated_reports CASCADE;
DROP TABLE IF EXISTS report_templates CASCADE;
DROP TABLE IF EXISTS regulators CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS report_type CASCADE;
DROP TYPE IF EXISTS report_status CASCADE;
DROP TYPE IF EXISTS report_format CASCADE;
DROP TYPE IF EXISTS report_frequency CASCADE;
DROP TYPE IF EXISTS regulator_type CASCADE;
DROP TYPE IF EXISTS export_type CASCADE;
DROP TYPE IF EXISTS export_status CASCADE;
DROP TYPE IF EXISTS encryption_type CASCADE;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd
