-- Regulatory Reporting Service Database Migrations
-- This file contains the initial schema for SARs, CTRs, compliance rules, alerts, and filing records

-- Create SARs table
CREATE TABLE IF NOT EXISTS sars (
    id UUID PRIMARY KEY,
    report_number VARCHAR(32) NOT NULL UNIQUE,
    subject_id VARCHAR(128) NOT NULL,
    subject_type VARCHAR(32) NOT NULL,
    subject_name VARCHAR(256) NOT NULL,
    suspicious_activity TEXT NOT NULL,
    activity_date TIMESTAMP WITH TIME ZONE NOT NULL,
    dollar_amount DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(8) DEFAULT 'USD',
    transaction_count INTEGER DEFAULT 1,
    narrative TEXT NOT NULL,
    risk_indicators JSONB DEFAULT '[]'::jsonb,
    filing_institution VARCHAR(256) NOT NULL,
    reporter_id VARCHAR(128) NOT NULL,
    reviewer_id VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    field_office VARCHAR(128),
    originating_agency VARCHAR(128),
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sars_report_number ON sars(report_number);
CREATE INDEX IF NOT EXISTS idx_sars_subject_id ON sars(subject_id);
CREATE INDEX IF NOT EXISTS idx_sars_status ON sars(status);
CREATE INDEX IF NOT EXISTS idx_sars_activity_date ON sars(activity_date);
CREATE INDEX IF NOT EXISTS idx_sars_reporter_id ON sars(reporter_id);

-- Create CTRs table
CREATE TABLE IF NOT EXISTS ctrs (
    id UUID PRIMARY KEY,
    report_number VARCHAR(32) NOT NULL UNIQUE,
    person_name VARCHAR(256) NOT NULL,
    person_id_type VARCHAR(32) NOT NULL,
    person_id_number VARCHAR(64) NOT NULL,
    person_address JSONB NOT NULL,
    account_number VARCHAR(64) NOT NULL,
    transaction_type VARCHAR(32) NOT NULL,
    cash_in_amount DECIMAL(20, 2) DEFAULT 0,
    cash_out_amount DECIMAL(20, 2) DEFAULT 0,
    total_amount DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(8) DEFAULT 'USD',
    transaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    method_received VARCHAR(32) NOT NULL,
    institution_id VARCHAR(64) NOT NULL,
    branch_id VARCHAR(64),
    teller_id VARCHAR(64),
    reviewer_id VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ctrs_report_number ON ctrs(report_number);
CREATE INDEX IF NOT EXISTS idx_ctrs_person_id ON ctrs(person_id_number);
CREATE INDEX IF NOT EXISTS idx_ctrs_account_number ON ctrs(account_number);
CREATE INDEX IF NOT EXISTS idx_ctrs_status ON ctrs(status);
CREATE INDEX IF NOT EXISTS idx_ctrs_transaction_date ON ctrs(transaction_date);

-- Create compliance_rules table
CREATE TABLE IF NOT EXISTS compliance_rules (
    id UUID PRIMARY KEY,
    rule_code VARCHAR(32) NOT NULL UNIQUE,
    rule_name VARCHAR(128) NOT NULL,
    category VARCHAR(64) NOT NULL,
    regulation VARCHAR(64) NOT NULL,
    description TEXT,
    threshold JSONB,
    logic TEXT,
    risk_score INTEGER DEFAULT 5,
    severity VARCHAR(16) DEFAULT 'medium',
    is_active BOOLEAN NOT NULL DEFAULT true,
    effective_date TIMESTAMP WITH TIME ZONE NOT NULL,
    expiry_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_compliance_rules_code ON compliance_rules(rule_code);
CREATE INDEX IF NOT EXISTS idx_compliance_rules_category ON compliance_rules(category);
CREATE INDEX IF NOT EXISTS idx_compliance_rules_regulation ON compliance_rules(regulation);
CREATE INDEX IF NOT EXISTS idx_compliance_rules_is_active ON compliance_rules(is_active);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    alert_number VARCHAR(32) NOT NULL UNIQUE,
    alert_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    subject_id VARCHAR(128) NOT NULL,
    subject_type VARCHAR(32) NOT NULL,
    subject_name VARCHAR(256) NOT NULL,
    transaction_ids JSONB DEFAULT '[]'::jsonb,
    description TEXT NOT NULL,
    triggered_rules JSONB DEFAULT '[]'::jsonb,
    risk_score INTEGER DEFAULT 5,
    assigned_to VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    resolution TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(128),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_alert_number ON alerts(alert_number);
CREATE INDEX IF NOT EXISTS idx_alerts_subject_id ON alerts(subject_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_risk_score ON alerts(risk_score);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

-- Create compliance_checks table
CREATE TABLE IF NOT EXISTS compliance_checks (
    id UUID PRIMARY KEY,
    subject_id VARCHAR(128) NOT NULL,
    subject_type VARCHAR(32) NOT NULL,
    check_type VARCHAR(64) NOT NULL,
    check_result VARCHAR(32) NOT NULL,
    risk_level VARCHAR(16) NOT NULL,
    match_score DECIMAL(5, 4) DEFAULT 0,
    matched_entity VARCHAR(256),
    match_details TEXT,
    screening_list VARCHAR(64),
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_compliance_checks_subject ON compliance_checks(subject_id, check_type);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_result ON compliance_checks(check_result);
CREATE INDEX IF NOT EXISTS idx_compliance_checks_checked_at ON compliance_checks(checked_at);

-- Create compliance_reports table
CREATE TABLE IF NOT EXISTS compliance_reports (
    id UUID PRIMARY KEY,
    report_type VARCHAR(32) NOT NULL,
    report_name VARCHAR(256) NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    summary JSONB NOT NULL,
    details JSONB,
    generated_by VARCHAR(128) NOT NULL,
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    file_path VARCHAR(512)
);

CREATE INDEX IF NOT EXISTS idx_compliance_reports_type ON compliance_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_status ON compliance_reports(status);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_generated_at ON compliance_reports(generated_at);

-- Create filing_records table
CREATE TABLE IF NOT EXISTS filing_records (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL,
    report_type VARCHAR(32) NOT NULL,
    report_number VARCHAR(32) NOT NULL,
    filing_method VARCHAR(32) NOT NULL,
    filing_agency VARCHAR(128) NOT NULL,
    filing_status VARCHAR(32) NOT NULL DEFAULT 'submitted',
    confirmation VARCHAR(128),
    submitted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMP WITH TIME ZONE,
    response_due TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_filing_records_report_id ON filing_records(report_id);
CREATE INDEX IF NOT EXISTS idx_filing_records_report_number ON filing_records(report_number);
CREATE INDEX IF NOT EXISTS idx_filing_records_filing_agency ON filing_records(filing_agency);
CREATE INDEX IF NOT EXISTS idx_filing_records_submitted_at ON filing_records(submitted_at);

-- Create supporting_documents table
CREATE TABLE IF NOT EXISTS supporting_documents (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL,
    report_type VARCHAR(32) NOT NULL,
    document_type VARCHAR(64) NOT NULL,
    file_name VARCHAR(256) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(64) NOT NULL,
    file_path VARCHAR(512) NOT NULL,
    description TEXT,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_supporting_docs_report_id ON supporting_documents(report_id);

-- Create triggers for automatic updated_at updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sars_updated_at
    BEFORE UPDATE ON sars
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ctrs_updated_at
    BEFORE UPDATE ON ctrs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compliance_rules_updated_at
    BEFORE UPDATE ON compliance_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alerts_updated_at
    BEFORE UPDATE ON alerts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
