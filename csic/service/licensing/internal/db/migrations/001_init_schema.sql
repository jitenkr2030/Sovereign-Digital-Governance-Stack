-- +goose Up
-- +goose StatementBegin

-- Create licensing and compliance management tables for CSIC Platform
-- This migration establishes the schema for exchange licensing, compliance monitoring,
-- automated reporting, and regulatory dashboard functionality.

-- Enum types for license and compliance status
CREATE TYPE license_type AS ENUM (
    'exchange_license',
    'custodial_license',
    'broker_license',
    'derivatives_license',
    'stablescoin_license',
    'payment_services_license'
);

CREATE TYPE license_status AS ENUM (
    'pending',
    'under_review',
    'approved',
    'active',
    'suspended',
    'revoked',
    'expired',
    'renewed'
);

CREATE TYPE compliance_status AS ENUM (
    'compliant',
    'non_compliant',
    'warning',
    'under_investigation',
    'resolved'
);

CREATE TYPE violation_severity AS ENUM (
    'low',
    'medium',
    'high',
    'critical'
);

CREATE TYPE report_type AS ENUM (
    'periodic',
    'suspicious_activity',
    'regulatory',
    'audit',
    'risk_assessment',
    'enforcement_action'
);

CREATE TYPE report_status AS ENUM (
    'draft',
    'submitted',
    'under_review',
    'approved',
    'rejected',
    'archived'
);

-- Licensing authority registry table
CREATE TABLE licensing_authorities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    jurisdiction VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    website VARCHAR(500),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    regulatory_framework TEXT,
    established_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_authorities_jurisdiction ON licensing_authorities(jurisdiction);
CREATE INDEX idx_authorities_active ON licensing_authorities(is_active);

-- License types catalog table
CREATE TABLE license_type_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_type license_type NOT NULL UNIQUE,
    description TEXT NOT NULL,
    regulatory_requirements TEXT[],
    validity_period_months INTEGER,
    renewal_requirements TEXT[],
    fee_structure DECIMAL(15, 2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Exchange licensing table
CREATE TABLE exchange_licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_number VARCHAR(100) NOT NULL UNIQUE,
    entity_name VARCHAR(255) NOT NULL,
    entity_registration_number VARCHAR(100),
    entity_type VARCHAR(100) NOT NULL,
    license_type license_type NOT NULL,
    license_status license_status DEFAULT 'pending',
    issuing_authority_id UUID REFERENCES licensing_authorities(id),
    jurisdiction VARCHAR(100) NOT NULL,
    registration_address TEXT,
    operational_address TEXT,
    website_url VARCHAR(500),
    business_description TEXT,
    supported_cryptocurrencies TEXT[],
    trading_pairs TEXT[],
    daily_volume_estimate DECIMAL(20, 2),
    monthly_volume_estimate DECIMAL(20, 2),
    user_base_estimate INTEGER,
    fiat_currencies_supported TEXT[],
    payment_methods TEXT[],
    license_issue_date DATE,
    license_expiry_date DATE,
    renewal_application_date DATE,
    last_audit_date DATE,
    next_audit_due_date DATE,
    conditions_and_restrictions TEXT[],
    special_permissions TEXT[],
    regulatory_filing_frequency VARCHAR(50),
    compliance_contact_name VARCHAR(255),
    compliance_contact_email VARCHAR(255),
    compliance_contact_phone VARCHAR(50),
    chief_compliance_officer VARCHAR(255),
    executive_contact_email VARCHAR(255),
    risk_management_framework TEXT,
    anti_money_laundering_policy TEXT,
    know_your_customer_policy TEXT,
    suspicious_activity_reporting_policy TEXT,
    sanctions_compliance_policy TEXT,
    transaction_monitoring_policy TEXT,
    security_incident_response_plan TEXT,
    data_protection_policy TEXT,
    audit_trail_policy TEXT,
    regulatory_reporting_capability BOOLEAN DEFAULT false,
    segregation_of_funds_policy BOOLEAN DEFAULT false,
    cybersecurity_framework TEXT,
    insurance_coverage DECIMAL(15, 2),
    insurance_provider VARCHAR(255),
    insurance_expiry_date DATE,
    capital_requirements_met BOOLEAN DEFAULT false,
    capital_requirements_amount DECIMAL(20, 2),
    previous_violations INTEGER DEFAULT 0,
    previous_sanctions INTEGER DEFAULT 0,
    public_notice_url VARCHAR(500),
    verification_documents TEXT[],
    notes TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_exchange_licenses_number ON exchange_licenses(license_number);
CREATE INDEX idx_exchange_licenses_entity ON exchange_licenses(entity_name);
CREATE INDEX idx_exchange_licenses_status ON exchange_licenses(license_status);
CREATE INDEX idx_exchange_licenses_type ON exchange_licenses(license_type);
CREATE INDEX idx_exchange_licenses_authority ON exchange_licenses(issuing_authority_id);
CREATE INDEX idx_exchange_licenses_expiry ON exchange_licenses(license_expiry_date);
CREATE INDEX idx_exchange_licenses_jurisdiction ON exchange_licenses(jurisdiction);

-- License status history for audit trail
CREATE TABLE license_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    previous_status license_status,
    new_status license_status NOT NULL,
    change_reason TEXT,
    changed_by VARCHAR(255),
    supporting_documents TEXT[],
    effective_date DATE NOT NULL,
    expiration_date DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_license_history_license ON license_status_history(license_id);
CREATE INDEX idx_license_history_date ON license_status_history(effective_date);

-- License amendments and modifications
CREATE TABLE license_amendments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    amendment_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    requested_changes JSONB NOT NULL,
    approval_status VARCHAR(50) DEFAULT 'pending',
    requested_by VARCHAR(255),
    requested_date DATE NOT NULL,
    reviewed_by VARCHAR(255),
    review_date DATE,
    approval_date DATE,
    effective_date DATE,
    rejection_reason TEXT,
    documents TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_amendments_license ON license_amendments(license_id);
CREATE INDEX idx_amendments_status ON license_amendments(approval_status);

-- Compliance requirements catalog
CREATE TABLE compliance_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requirement_code VARCHAR(50) NOT NULL UNIQUE,
    requirement_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    regulatory_source VARCHAR(255),
    applicable_license_types license_type[],
    jurisdiction VARCHAR(100),
    compliance_category VARCHAR(100),
    risk_level violation_severity,
    frequency_requirement VARCHAR(50),
    documentation_required TEXT[],
    verification_method VARCHAR(100),
    penalty_guidelines TEXT,
    effective_date DATE,
    expiry_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_compliance_requirements_code ON compliance_requirements(requirement_code);
CREATE INDEX idx_compliance_requirements_category ON compliance_requirements(compliance_category);
CREATE INDEX idx_compliance_requirements_active ON compliance_requirements(is_active);

-- Compliance monitoring records
CREATE TABLE compliance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    requirement_id UUID REFERENCES compliance_requirements(id),
    compliance_status compliance_status NOT NULL,
    assessment_period_start DATE NOT NULL,
    assessment_period_end DATE NOT NULL,
    assessment_date DATE NOT NULL,
    assessed_by VARCHAR(255) NOT NULL,
    methodology_used TEXT,
    findings_summary TEXT,
    risk_score DECIMAL(5, 2),
    evidence_documents TEXT[],
    remediation_plan TEXT,
    remediation_deadline DATE,
    follow_up_required BOOLEAN DEFAULT false,
    follow_up_date DATE,
    regulatory_notifications_sent TEXT[],
    internal_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_compliance_records_license ON compliance_records(license_id);
CREATE INDEX idx_compliance_records_status ON compliance_records(compliance_status);
CREATE INDEX idx_compliance_records_period ON compliance_records(assessment_period_start, assessment_period_end);
CREATE INDEX idx_compliance_records_assessor ON compliance_records(assessed_by);

-- Violations tracking table
CREATE TABLE compliance_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    compliance_record_id UUID REFERENCES compliance_records(id),
    violation_type VARCHAR(100) NOT NULL,
    violation_description TEXT NOT NULL,
    severity violation_severity NOT NULL,
    detection_date DATE NOT NULL,
    violation_status VARCHAR(50) DEFAULT 'open',
    root_cause_analysis TEXT,
    corrective_actions TEXT,
    preventive_actions TEXT,
    remediation_deadline DATE,
    remediation_completed_date DATE,
    penalty_imposed BOOLEAN DEFAULT false,
    penalty_type VARCHAR(100),
    penalty_amount DECIMAL(15, 2),
    penalty_currency VARCHAR(10),
    penalty_reason TEXT,
    legal_action_initiated BOOLEAN DEFAULT false,
    legal_action_details TEXT,
    regulatory_notice_received BOOLEAN DEFAULT false,
    regulatory_notice_date DATE,
    public_disclosure_required BOOLEAN DEFAULT false,
    impact_assessment TEXT,
    resolution_notes TEXT,
    closed_by VARCHAR(255),
    closure_date DATE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_violations_license ON compliance_violations(license_id);
CREATE INDEX idx_violations_status ON compliance_violations(violation_status);
CREATE INDEX idx_violations_severity ON compliance_violations(severity);
CREATE INDEX idx_violations_detection_date ON compliance_violations(detection_date);

-- Enforcement actions table
CREATE TABLE enforcement_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    violation_id UUID REFERENCES compliance_violations(id),
    action_type VARCHAR(100) NOT NULL,
    action_description TEXT NOT NULL,
    legal_reference VARCHAR(255),
    initiation_date DATE NOT NULL,
    resolution_date DATE,
    action_status VARCHAR(50) DEFAULT 'pending',
    regulatory_authority VARCHAR(255),
    penalties_imposed TEXT[],
    compliance_directives TEXT[],
    supervised_release_conditions TEXT[],
    monitor_appointment BOOLEAN DEFAULT false,
    monitor_details TEXT,
    revocation_order_date DATE,
    fine_amount DECIMAL(15, 2),
    fine_currency VARCHAR(10),
    restitution_required BOOLEAN DEFAULT false,
    restitution_amount DECIMAL(15, 2),
    public_statement_required BOOLEAN DEFAULT false,
    appeal_filed BOOLEAN DEFAULT false,
    appeal_date DATE,
    appeal_outcome VARCHAR(100),
    appeal_decision_date DATE,
    case_documents TEXT[],
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_enforcement_license ON enforcement_actions(license_id);
CREATE INDEX idx_enforcement_status ON enforcement_actions(action_status);
CREATE INDEX idx_enforcement_type ON enforcement_actions(action_type);
CREATE INDEX idx_enforcement_initiation ON enforcement_actions(initiation_date);

-- Regulatory reports table
CREATE TABLE regulatory_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_number VARCHAR(100) NOT NULL UNIQUE,
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    report_type report_type NOT NULL,
    report_status report_status DEFAULT 'draft',
    report_period_start DATE NOT NULL,
    report_period_end DATE NOT NULL,
    submission_deadline DATE NOT NULL,
    submitted_date DATE,
    reporting_authority VARCHAR(255),
    report_title VARCHAR(500) NOT NULL,
    executive_summary TEXT,
    key_findings TEXT,
    statistical_data JSONB,
    compliance_metrics JSONB,
    risk_indicators JSONB,
    suspicious_activity_summary TEXT,
    transaction_statistics JSONB,
    user_statistics JSONB,
    financial_data JSONB,
    recommended_actions TEXT,
    attachments TEXT[],
    prepared_by VARCHAR(255),
    reviewed_by VARCHAR(255),
    approval_date DATE,
    submission_channel VARCHAR(100),
    acknowledgment_received BOOLEAN DEFAULT false,
    acknowledgment_date DATE,
    follow_up_required BOOLEAN DEFAULT false,
    follow_up_deadline DATE,
    regulatory_response TEXT,
    notes TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reports_license ON regulatory_reports(license_id);
CREATE INDEX idx_reports_type ON regulatory_reports(report_type);
CREATE INDEX idx_reports_status ON regulatory_reports(report_status);
CREATE INDEX idx_reports_period ON regulatory_reports(report_period_start, report_period_end);
CREATE INDEX idx_reports_number ON regulatory_reports(report_number);
CREATE INDEX idx_reports_deadline ON regulatory_reports(submission_deadline);

-- Suspicious activity reports (SARs)
CREATE TABLE suspicious_activity_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sar_number VARCHAR(100) NOT NULL UNIQUE,
    license_id UUID NOT NULL REFERENCES exchange_licenses(id),
    suspicious_activity_type VARCHAR(100) NOT NULL,
    detection_date DATE NOT NULL,
    suspicious_period_start DATE NOT NULL,
    suspicious_period_end DATE NOT NULL,
    subject_type VARCHAR(50), -- individual, entity, wallet
    subject_identifier VARCHAR(500),
    subject_name VARCHAR(255),
    subject_accounts TEXT[],
    involved_wallets TEXT[],
    transaction_count INTEGER,
    total_volume DECIMAL(20, 2),
    currency VARCHAR(10),
    suspicious_pattern_description TEXT,
    red_flags_identified TEXT[],
    preliminary_assessment TEXT,
    supporting_evidence TEXT[],
    btc_amount DECIMAL(20, 8),
    usd_value DECIMAL(20, 2),
    fiat_currency_involved TEXT[],
    geographic_origin TEXT[],
    geographic_destination TEXT[],
    correspondent_bank_involved TEXT[],
    legal_review_required BOOLEAN DEFAULT false,
    legal_review_completed BOOLEAN DEFAULT false,
    law_enforcement_referred BOOLEAN DEFAULT false,
    law_enforcement_referral_date DATE,
    filing_deadline DATE,
    filed_date DATE,
    filed_by VARCHAR(255),
    regulatory acknowledgment TEXT,
    sar_status VARCHAR(50) DEFAULT 'under_review',
    closure_date DATE,
    closure_reason TEXT,
    lessons_learned TEXT,
    pattern_detected TEXT[],
    counter_measures_implemented TEXT,
    notes TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sars_license ON suspicious_activity_reports(license_id);
CREATE INDEX idx_sars_number ON suspicious_activity_reports(sar_number);
CREATE INDEX idx_sars_status ON suspicious_activity_reports(sar_status);
CREATE INDEX idx_sars_detection_date ON suspicious_activity_reports(detection_date);
CREATE INDEX idx_sars_subject ON suspicious_activity_reports(subject_identifier);

-- Compliance dashboard metrics table
CREATE TABLE dashboard_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_date DATE NOT NULL,
    metric_type VARCHAR(100) NOT NULL,
    metric_category VARCHAR(100),
    metric_value DECIMAL(20, 4) NOT NULL,
    metric_unit VARCHAR(50),
    previous_value DECIMAL(20, 4),
    change_percentage DECIMAL(10, 2),
    benchmark_value DECIMAL(20, 4),
    trend_direction VARCHAR(20), -- up, down, stable
    calculation_method TEXT,
    data_sources TEXT[],
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_dashboard_metrics_date ON dashboard_metrics(metric_date);
CREATE INDEX idx_dashboard_metrics_type ON dashboard_metrics(metric_type);
CREATE INDEX idx_dashboard_metrics_category ON dashboard_metrics(metric_category);

-- Dashboard data snapshots for historical tracking
CREATE TABLE dashboard_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date TIMESTAMP WITH TIME ZONE NOT NULL,
    reporting_period_start DATE,
    reporting_period_end DATE,
    total_licensed_entities INTEGER,
    active_licenses INTEGER,
    pending_applications INTEGER,
    suspended_licenses INTEGER,
    revoked_licenses INTEGER,
    compliance_rate DECIMAL(5, 2),
    average_risk_score DECIMAL(5, 2),
    total_violations INTEGER,
    critical_violations INTEGER,
    high_violations INTEGER,
    medium_violations INTEGER,
    low_violations INTEGER,
    open_investigations INTEGER,
    pending_reports INTEGER,
    overdue_reports INTEGER,
    sar_filings_count INTEGER,
    enforcement_actions_active INTEGER,
    new_licenses_issued INTEGER,
    licenses_expired INTEGER,
    licenses_renewed INTEGER,
    top_jurisdictions JSONB,
    license_type_distribution JSONB,
    compliance_trend JSONB,
    risk_distribution JSONB,
    snapshot_data JSONB NOT NULL,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_dashboard_snapshots_date ON dashboard_snapshots(snapshot_date);

-- Audit log for compliance activities
CREATE TABLE compliance_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(100),
    license_id UUID REFERENCES exchange_licenses(id),
    record_type VARCHAR(100),
    record_id UUID,
    action_taken VARCHAR(100) NOT NULL,
    actor_type VARCHAR(50), -- system, user, automated_process
    actor_id VARCHAR(255),
    actor_name VARCHAR(255),
    previous_state JSONB,
    new_state JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    reason_for_change TEXT,
    supporting_documentation TEXT[],
    risk_assessment TEXT,
    approval_chain JSONB,
    event_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_log_date ON compliance_audit_log(event_date);
CREATE INDEX idx_audit_log_type ON compliance_audit_log(event_type);
CREATE INDEX idx_audit_log_license ON compliance_audit_log(license_id);
CREATE INDEX idx_audit_log_actor ON compliance_audit_log(actor_id);
CREATE INDEX idx_audit_log_session ON compliance_audit_log(session_id);

-- Notifications and alerts table
CREATE TABLE compliance_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_type VARCHAR(100) NOT NULL,
    notification_priority VARCHAR(50) DEFAULT 'normal',
    license_id UUID REFERENCES exchange_licenses(id),
    recipient_type VARCHAR(50), -- internal, external, regulatory
    recipient_email VARCHAR(255),
    recipient_name VARCHAR(255),
    subject VARCHAR(500) NOT NULL,
    message_body TEXT NOT NULL,
    action_required TEXT,
    due_date DATE,
    escalation_rules JSONB,
    related_record_type VARCHAR(100),
    related_record_id UUID,
    status VARCHAR(50) DEFAULT 'pending',
    sent_date TIMESTAMP WITH TIME ZONE,
    read_date TIMESTAMP WITH TIME ZONE,
    acknowledged_by VARCHAR(255),
    response_received BOOLEAN DEFAULT false,
    response_details TEXT,
    closure_date TIMESTAMP WITH TIME ZONE,
    closure_notes TEXT,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifications_type ON compliance_notifications(notification_type);
CREATE INDEX idx_notifications_status ON compliance_notifications(status);
CREATE INDEX idx_notifications_license ON compliance_notifications(license_id);
CREATE INDEX idx_notifications_due_date ON compliance_notifications(due_date);
CREATE INDEX idx_notifications_priority ON compliance_notifications(notification_priority);

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

-- Drop all tables in reverse order to handle foreign key constraints
DROP TABLE IF EXISTS compliance_notifications CASCADE;
DROP TABLE IF EXISTS compliance_audit_log CASCADE;
DROP TABLE IF EXISTS dashboard_snapshots CASCADE;
DROP TABLE IF EXISTS dashboard_metrics CASCADE;
DROP TABLE IF EXISTS suspicious_activity_reports CASCADE;
DROP TABLE IF EXISTS regulatory_reports CASCADE;
DROP TABLE IF EXISTS enforcement_actions CASCADE;
DROP TABLE IF EXISTS compliance_violations CASCADE;
DROP TABLE IF EXISTS compliance_records CASCADE;
DROP TABLE IF EXISTS license_amendments CASCADE;
DROP TABLE IF EXISTS license_status_history CASCADE;
DROP TABLE IF EXISTS exchange_licenses CASCADE;
DROP TABLE IF EXISTS compliance_requirements CASCADE;
DROP TABLE IF EXISTS license_type_catalog CASCADE;
DROP TABLE IF EXISTS licensing_authorities CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS license_type CASCADE;
DROP TYPE IF EXISTS license_status CASCADE;
DROP TYPE IF EXISTS compliance_status CASCADE;
DROP TYPE IF EXISTS violation_severity CASCADE;
DROP TYPE IF EXISTS report_type CASCADE;
DROP TYPE IF EXISTS report_status CASCADE;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd
