-- Audit Log Service Database Migrations
-- Creates PostgreSQL schema for immutable audit logging

-- Create audit_entries table with cryptographic integrity fields
CREATE TABLE IF NOT EXISTS audit_entries (
    id UUID PRIMARY KEY,
    entry_id VARCHAR(64) UNIQUE NOT NULL,
    sequence_num BIGINT UNIQUE NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    chain_id VARCHAR(64) NOT NULL,
    
    -- Actor information
    actor_id VARCHAR(64) NOT NULL,
    actor_type VARCHAR(32) NOT NULL,  -- user, system, service
    actor_role VARCHAR(64),
    session_id VARCHAR(64),
    ip_address INET,
    user_agent TEXT,
    
    -- Action details
    service VARCHAR(64) NOT NULL,
    operation VARCHAR(64) NOT NULL,
    action_type VARCHAR(32) NOT NULL,  -- create, read, update, delete, execute
    resource VARCHAR(128) NOT NULL,
    resource_id VARCHAR(64),
    description TEXT,
    
    -- Result and impact
    result VARCHAR(32) NOT NULL,  -- success, failure, partial
    error_code VARCHAR(32),
    error_msg TEXT,
    affected_ids TEXT[],  -- array of affected resource IDs
    
    -- Compliance and classification
    compliance_tags TEXT[],  -- array of compliance tags (ISO27001, GDPR, etc.)
    risk_level VARCHAR(16) NOT NULL,  -- low, medium, high, critical
    regulatory_ref VARCHAR(128),
    
    -- Cryptographic integrity
    previous_hash CHAR(64) NOT NULL,
    current_hash CHAR(64) UNIQUE NOT NULL,
    signature TEXT,  -- Ed25519 or RSA signature
    
    -- Additional context
    metadata JSONB,
    trace_id VARCHAR(64),
    
    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create audit_chains table for chain sealing
CREATE TABLE IF NOT EXISTS audit_chains (
    chain_id VARCHAR(64) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    sealed_at TIMESTAMPTZ NOT NULL,
    sequence_start BIGINT NOT NULL,
    sequence_end BIGINT NOT NULL,
    entry_count INTEGER NOT NULL,
    root_hash CHAR(64) NOT NULL,
    seal_signature TEXT NOT NULL,
    previous_chain VARCHAR(64),
    metadata JSONB
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_audit_entries_timestamp ON audit_entries(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_entries_actor_id ON audit_entries(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_entries_service ON audit_entries(service);
CREATE INDEX IF NOT EXISTS idx_audit_entries_operation ON audit_entries(operation);
CREATE INDEX IF NOT EXISTS idx_audit_entries_action_type ON audit_entries(action_type);
CREATE INDEX IF NOT EXISTS idx_audit_entries_resource ON audit_entries(resource);
CREATE INDEX IF NOT EXISTS idx_audit_entries_result ON audit_entries(result);
CREATE INDEX IF NOT EXISTS idx_audit_entries_risk_level ON audit_entries(risk_level);
CREATE INDEX IF NOT EXISTS idx_audit_entries_chain_id ON audit_entries(chain_id);
CREATE INDEX IF NOT EXISTS idx_audit_entries_compliance_tags ON audit_entries USING GIN(compliance_tags);
CREATE INDEX IF NOT EXISTS idx_audit_entries_metadata ON audit_entries USING GIN(metadata);

-- Create indexes for chains table
CREATE INDEX IF NOT EXISTS idx_audit_chains_sealed_at ON audit_chains(sealed_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_chains_sequence ON audit_chains(sequence_start, sequence_end);

-- Create function to validate hash chain integrity
CREATE OR REPLACE FUNCTION validate_hash_chain()
RETURNS TRIGGER AS $$
DECLARE
    prev_entry audit_entries%ROWTYPE;
BEGIN
    -- Get previous entry
    SELECT * INTO prev_entry
    FROM audit_entries
    WHERE sequence_num = NEW.sequence_num - 1
    AND chain_id = NEW.chain_id;
    
    -- For first entry, previous hash must be all zeros
    IF NEW.sequence_num = 1 THEN
        IF NEW.previous_hash != '0000000000000000000000000000000000000000000000000000000000000000' THEN
            RAISE EXCEPTION 'Invalid previous hash for first entry';
        END IF;
    ELSE
        -- Verify previous hash matches
        IF NOT FOUND THEN
            RAISE EXCEPTION 'Previous entry not found for sequence %', NEW.sequence_num;
        END IF;
        
        IF prev_entry.current_hash != NEW.previous_hash THEN
            RAISE EXCEPTION 'Hash chain broken at sequence %', NEW.sequence_num;
        END IF;
    END IF;
    
    -- Calculate and verify current hash
    IF NEW.current_hash != encode(sha256(
        NEW.entry_id::text || 
        NEW.sequence_num::text || 
        NEW.timestamp::text || 
        NEW.chain_id::text || 
        NEW.actor_id::text || 
        NEW.service::text || 
        NEW.operation::text || 
        NEW.action_type::text || 
        NEW.resource::text || 
        NEW.result::text || 
        NEW.previous_hash::text
    ), 'hex') THEN
        RAISE EXCEPTION 'Invalid current hash for entry %', NEW.entry_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create trigger for hash chain validation
DROP TRIGGER IF EXISTS trigger_hash_chain_validation ON audit_entries;
CREATE TRIGGER trigger_hash_chain_validation
    BEFORE INSERT ON audit_entries
    FOR EACH ROW
    EXECUTE FUNCTION validate_hash_chain();

-- Create function to prevent updates and deletes (WORM)
CREATE OR REPLACE FUNCTION prevent_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit log entries cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create triggers to prevent modifications
DROP TRIGGER IF EXISTS trigger_prevent_update ON audit_entries;
CREATE TRIGGER trigger_prevent_update
    BEFORE UPDATE ON audit_entries
    FOR EACH ROW
    EXECUTE FUNCTION prevent_modification();

DROP TRIGGER IF EXISTS trigger_prevent_delete ON audit_entries;
CREATE TRIGGER trigger_prevent_delete
    BEFORE DELETE ON audit_entries
    FOR EACH ROW
    EXECUTE FUNCTION prevent_modification();

-- Create function to get chain summary
CREATE OR REPLACE FUNCTION get_chain_summary(p_chain_id VARCHAR(64))
RETURNS TABLE (
    chain_id VARCHAR(64),
    sequence_start BIGINT,
    sequence_end BIGINT,
    entry_count BIGINT,
    root_hash CHAR(64),
    sealed_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ac.chain_id,
        ac.sequence_start,
        ac.sequence_end,
        ac.entry_count::BIGINT,
        ac.root_hash,
        ac.sealed_at
    FROM audit_chains ac
    WHERE ac.chain_id = p_chain_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create view for audit summary statistics
CREATE OR REPLACE VIEW audit_summary AS
SELECT 
    service,
    action_type,
    result,
    risk_level,
    COUNT(*) as entry_count,
    DATE_TRUNC('day', timestamp) as day
FROM audit_entries
GROUP BY service, action_type, result, risk_level, DATE_TRUNC('day', timestamp);

-- Grant read access to authenticated users
GRANT SELECT ON ALL TABLES IN SCHEMA public TO csic_reader;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO csic_reader;

-- Grant write access to audit service
GRANT INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO csic_audit;
