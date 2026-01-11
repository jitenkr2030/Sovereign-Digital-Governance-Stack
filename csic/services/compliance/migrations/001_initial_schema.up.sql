-- Compliance Module Database Schema
-- Migration: 001_initial_schema

-- Create UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Entities Table (Regulated entities)
CREATE TABLE IF NOT EXISTS compliance_entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    registration_num VARCHAR(100) NOT NULL UNIQUE,
    jurisdiction VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    risk_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entities_registration ON compliance_entities(registration_num);
CREATE INDEX IF NOT EXISTS idx_entities_jurisdiction ON compliance_entities(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_entities_status ON compliance_entities(status);

-- Regulations Table (Regulatory requirements)
CREATE TABLE IF NOT EXISTS compliance_regulations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    jurisdiction VARCHAR(100) NOT NULL,
    effective_date TIMESTAMPTZ NOT NULL,
    parent_id UUID,
    requirements TEXT NOT NULL,
    penalty_details TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_parent_regulation FOREIGN KEY (parent_id) REFERENCES compliance_regulations(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_regulations_category ON compliance_regulations(category);
CREATE INDEX IF NOT EXISTS idx_regulations_jurisdiction ON compliance_regulations(jurisdiction);
CREATE INDEX IF NOT EXISTS idx_regulations_effective_date ON compliance_regulations(effective_date);

-- Licenses Table
CREATE TABLE IF NOT EXISTS compliance_licenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    license_number VARCHAR(100) NOT NULL UNIQUE,
    issued_date TIMESTAMPTZ NOT NULL,
    expiry_date TIMESTAMPTZ NOT NULL,
    conditions TEXT,
    restrictions TEXT,
    jurisdiction VARCHAR(100),
    issued_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT,
    
    CONSTRAINT fk_license_entity FOREIGN KEY (entity_id) REFERENCES compliance_entities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_licenses_entity ON compliance_licenses(entity_id);
CREATE INDEX IF NOT EXISTS idx_licenses_status ON compliance_licenses(status);
CREATE INDEX IF NOT EXISTS idx_licenses_number ON compliance_licenses(license_number);
CREATE INDEX IF NOT EXISTS idx_licenses_expiry ON compliance_licenses(expiry_date);

-- License Applications Table
CREATE TABLE IF NOT EXISTS compliance_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    license_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    submitted_at TIMESTAMPTZ,
    reviewed_at TIMESTAMPTZ,
    reviewer_id UUID,
    reviewer_notes TEXT,
    requested_terms TEXT NOT NULL,
    granted_terms TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_application_entity FOREIGN KEY (entity_id) REFERENCES compliance_entities(id) ON DELETE CASCADE,
    CONSTRAINT fk_reviewer FOREIGN KEY (reviewer_id) REFERENCES compliance_entities(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_applications_entity ON compliance_applications(entity_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON compliance_applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_submitted ON compliance_applications(submitted_at);

-- Compliance Scores Table
CREATE TABLE IF NOT EXISTS compliance_scores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    total_score DECIMAL(5,2) NOT NULL,
    tier VARCHAR(50) NOT NULL,
    breakdown JSONB NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    calculated_at TIMESTAMPTZ DEFAULT NOW(),
    calculation_details TEXT,
    
    CONSTRAINT fk_score_entity FOREIGN KEY (entity_id) REFERENCES compliance_entities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scores_entity ON compliance_scores(entity_id);
CREATE INDEX IF NOT EXISTS idx_scores_tier ON compliance_scores(tier);
CREATE INDEX IF NOT EXISTS idx_scores_calculated ON compliance_scores(calculated_at DESC);

-- Obligations Table
CREATE TABLE IF NOT EXISTS compliance_obligations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    regulation_id UUID NOT NULL,
    description TEXT NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    priority INTEGER DEFAULT 1,
    evidence_refs TEXT,
    fulfilled_at TIMESTAMPTZ,
    fulfilled_evidence TEXT,
    reminder_sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT fk_obligation_entity FOREIGN KEY (entity_id) REFERENCES compliance_entities(id) ON DELETE CASCADE,
    CONSTRAINT fk_obligation_regulation FOREIGN KEY (regulation_id) REFERENCES compliance_regulations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_obligations_entity ON compliance_obligations(entity_id);
CREATE INDEX IF NOT EXISTS idx_obligations_status ON compliance_obligations(status);
CREATE INDEX IF NOT EXISTS idx_obligations_due_date ON compliance_obligations(due_date);
CREATE INDEX IF NOT EXISTS idx_obligations_regulation ON compliance_obligations(regulation_id);

-- Audit Records Table
CREATE TABLE IF NOT EXISTS compliance_audit_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    action_type VARCHAR(100) NOT NULL,
    actor_id UUID NOT NULL,
    actor_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    old_value JSONB,
    new_value JSONB,
    changes TEXT,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    
    CONSTRAINT fk_audit_entity FOREIGN KEY (entity_id) REFERENCES compliance_entities(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_entity ON compliance_audit_records(entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON compliance_audit_records(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON compliance_audit_records(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON compliance_audit_records(action_type);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON compliance_audit_records(timestamp DESC);

-- Trigger functions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_entities_updated_at
    BEFORE UPDATE ON compliance_entities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_regulations_updated_at
    BEFORE UPDATE ON compliance_regulations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_licenses_updated_at
    BEFORE UPDATE ON compliance_licenses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_applications_updated_at
    BEFORE UPDATE ON compliance_applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_obligations_updated_at
    BEFORE UPDATE ON compliance_obligations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO compliance_entities (name, legal_name, registration_num, jurisdiction, entity_type, address, contact_email, status, risk_level)
VALUES 
    ('Crypto Exchange Pro', 'Crypto Exchange Pro LLC', 'CRYPTO-2024-001', 'US', 'EXCHANGE', '123 Blockchain Ave, NY', 'compliance@cryptoexchangepro.com', 'ACTIVE', 'MEDIUM'),
    ('Digital Assets Inc', 'Digital Assets Incorporated', 'DIGITAL-2024-002', 'US', 'CUSTODY', '456 Ledger Lane, CA', 'legal@digitalassets.com', 'ACTIVE', 'LOW'),
    ('Mining Operations Corp', 'Mining Operations Corporation', 'MINING-2024-003', 'CANADA', 'MINING', '789 Hash Road, ON', 'ops@miningcorp.com', 'ACTIVE', 'HIGH')
ON CONFLICT (registration_num) DO NOTHING;

-- Insert sample regulations
INSERT INTO compliance_regulations (title, description, category, jurisdiction, effective_date, requirements, penalty_details)
VALUES 
    ('KYC/AML Requirements', 'Know Your Customer and Anti-Money Laundering requirements', 'AML', 'US', NOW(), 'Verify customer identity, monitor transactions, report suspicious activity', 'Fines up to $1M per violation'),
    ('Data Protection Act', 'Consumer data protection and privacy requirements', 'PRIVACY', 'EU', NOW(), 'Implement data protection measures, consent management, breach notification', 'Fines up to 4% of annual revenue'),
    ('Operational Resilience', 'Requirements for operational resilience and continuity', 'OPERATIONAL', 'UK', NOW(), 'Business continuity plans, disaster recovery, incident management', 'Regulatory sanctions')
ON CONFLICT DO NOTHING;
