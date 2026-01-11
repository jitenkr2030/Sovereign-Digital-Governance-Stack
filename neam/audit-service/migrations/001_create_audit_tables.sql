-- Audit Service Database Schema
-- PostgreSQL DDL for Audit, Security & Sovereignty Layer

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- Core Audit Tables
-- =====================================================

-- Audit Events Table
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL DEFAULT 'USER',
    actor_name VARCHAR(255),
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(512) NOT NULL,
    resource_id VARCHAR(255),
    resource_type VARCHAR(100),
    outcome VARCHAR(50) NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    location JSONB,
    session_id VARCHAR(255),
    request_id UUID,
    trace_id UUID,
    integrity_hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64),
    chain_index BIGINT NOT NULL DEFAULT 0,
    sealed BOOLEAN NOT NULL DEFAULT FALSE,
    sealed_at TIMESTAMPTZ,
    merkle_proof JSONB,
    compliance_tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for audit_events
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_events_actor_id ON audit_events(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_resource ON audit_events(resource);
CREATE INDEX IF NOT EXISTS idx_audit_events_outcome ON audit_events(outcome);
CREATE INDEX IF NOT EXISTS idx_audit_events_sealed ON audit_events(sealed) WHERE sealed = FALSE;
CREATE INDEX IF NOT EXISTS idx_audit_events_chain_index ON audit_events(chain_index);
CREATE INDEX IF NOT EXISTS idx_audit_events_ip_address ON audit_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_events_session_id ON audit_events(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp_range ON audit_events(timestamp) 
    WITH (autosummarize = true);

-- Full-text search index for details
CREATE INDEX IF NOT EXISTS idx_audit_events_details_gin ON audit_events USING gin(details);
CREATE INDEX IF NOT EXISTS idx_audit_events_location_gin ON audit_events USING gin(location);

-- =====================================================
-- Seal Records Table
-- =====================================================

CREATE TABLE IF NOT EXISTS audit_seals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    root_hash VARCHAR(64) NOT NULL,
    first_event_id UUID NOT NULL,
    last_event_id UUID NOT NULL,
    event_count BIGINT NOT NULL,
    signature TEXT NOT NULL,
    signer_id VARCHAR(255) NOT NULL,
    signer_type VARCHAR(50) NOT NULL DEFAULT 'SYSTEM',
    algorithm VARCHAR(50) NOT NULL DEFAULT 'RSA-SHA256',
    witness_type VARCHAR(50),
    witness_id VARCHAR(255),
    witness_signature TEXT,
    previous_seal_id UUID,
    previous_root_hash VARCHAR(64),
    metadata JSONB,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for audit_seals
CREATE INDEX IF NOT EXISTS idx_audit_seals_timestamp ON audit_seals(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_seals_root_hash ON audit_seals(root_hash);
CREATE INDEX IF NOT EXISTS idx_audit_seals_first_event ON audit_seals(first_event_id);
CREATE INDEX IF NOT EXISTS idx_audit_seals_last_event ON audit_seals(last_event_id);
CREATE INDEX IF NOT EXISTS idx_audit_seals_verified ON audit_seals(verified) WHERE verified = FALSE;

-- =====================================================
-- User Management Tables
-- =====================================================

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    permissions TEXT[],
    groups TEXT[],
    active BOOLEAN NOT NULL DEFAULT TRUE,
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret TEXT,
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    locked_until TIMESTAMPTZ,
    failed_logins INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    password_changed TIMESTAMPTZ
);

-- Indexes for users
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active) WHERE active = TRUE;

-- =====================================================
-- Session Management Tables
-- =====================================================

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    location JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    last_activity TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMPTZ,
    revoke_reason TEXT
);

-- Indexes for sessions
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_revoked ON sessions(revoked) WHERE revoked = FALSE;

-- =====================================================
-- Policy Management Tables
-- =====================================================

CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    effect VARCHAR(10) NOT NULL CHECK (effect IN ('allow', 'deny')),
    priority INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    conditions JSONB,
    resources TEXT[] NOT NULL,
    actions TEXT[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for policies
CREATE INDEX IF NOT EXISTS idx_policies_active ON policies(active) WHERE active = TRUE;
CREATE INDEX IF NOT EXISTS idx_policies_priority ON policies(priority DESC);

-- =====================================================
-- Forensic Investigation Tables
-- =====================================================

CREATE TABLE IF NOT EXISTS forensic_investigations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id UUID NOT NULL REFERENCES users(id),
    requester_role VARCHAR(50) NOT NULL,
    query JSONB NOT NULL,
    justification TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Indexes for forensic_investigations
CREATE INDEX IF NOT EXISTS idx_forensic_investigations_status ON forensic_investigations(status);
CREATE INDEX IF NOT EXISTS idx_forensic_investigations_requester ON forensic_investigations(requester_id);
CREATE INDEX IF NOT EXISTS idx_forensic_investigations_created ON forensic_investigations(created_at DESC);

-- =====================================================
-- Chain of Custody Tables
-- =====================================================

CREATE TABLE IF NOT EXISTS chain_of_custody (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    record_type VARCHAR(50) NOT NULL,
    record_id UUID NOT NULL,
    event_timestamp TIMESTAMPTZ NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id UUID REFERENCES users(id),
    actor_type VARCHAR(50) NOT NULL DEFAULT 'USER',
    details JSONB,
    integrity_hash VARCHAR(64) NOT NULL,
    signature TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for chain_of_custody
CREATE INDEX IF NOT EXISTS idx_chain_of_custody_record ON chain_of_custody(record_type, record_id);
CREATE INDEX IF NOT EXISTS idx_chain_of_custody_timestamp ON chain_of_custody(event_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_chain_of_custody_actor ON chain_of_custody(actor_id);

-- =====================================================
-- Compliance Evidence Tables
-- =====================================================

CREATE TABLE IF NOT EXISTS compliance_evidence (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    standard VARCHAR(50) NOT NULL,
    requirement_id VARCHAR(100) NOT NULL,
    evidence_type VARCHAR(50) NOT NULL,
    record_ids UUID[],
    time_range JSONB NOT NULL,
    generated_by UUID REFERENCES users(id),
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMPTZ,
    metadata JSONB,
    signature TEXT NOT NULL
);

-- Indexes for compliance_evidence
CREATE INDEX IF NOT EXISTS idx_compliance_evidence_standard ON compliance_evidence(standard);
CREATE INDEX IF NOT EXISTS idx_compliance_evidence_requirement ON compliance_evidence(requirement_id);
CREATE INDEX IF NOT EXISTS idx_compliance_evidence_generated ON compliance_evidence(generated_at DESC);

-- =====================================================
-- Functions and Triggers
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-update updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at
    BEFORE UPDATE ON policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to calculate chain index
CREATE OR REPLACE FUNCTION calculate_chain_index()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.chain_index IS NULL THEN
        SELECT COALESCE(MAX(chain_index), -1) + 1 INTO NEW.chain_index
        FROM audit_events;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-set chain index
CREATE TRIGGER set_chain_index
    BEFORE INSERT ON audit_events
    FOR EACH ROW
    EXECUTE FUNCTION calculate_chain_index();

-- Function to verify chain integrity
CREATE OR REPLACE FUNCTION verify_chain_integrity()
RETURNS TABLE (
    is_valid BOOLEAN,
    broken_links INTEGER,
    first_broken_id UUID
) AS $$
DECLARE
    broken_count INTEGER := 0;
    first_broken UUID;
BEGIN
    -- Check for chain continuity
    SELECT COUNT(*) INTO broken_count
    FROM (
        SELECT 
            id,
            LAG(integrity_hash) OVER (ORDER BY chain_index) as prev_hash,
            previous_hash
        FROM audit_events
        WHERE sealed = FALSE
    ) t
    WHERE prev_hash IS NOT NULL AND prev_hash != previous_hash;

    -- Get first broken link
    SELECT id INTO first_broken
    FROM (
        SELECT 
            id,
            LAG(integrity_hash) OVER (ORDER BY chain_index) as prev_hash,
            previous_hash,
            ROW_NUMBER() OVER (ORDER BY chain_index) as rn
        FROM audit_events
        WHERE sealed = FALSE
    ) t
    WHERE prev_hash IS NOT NULL AND prev_hash != previous_hash
    ORDER BY rn
    LIMIT 1;

    RETURN QUERY SELECT broken_count = 0, broken_count, first_broken;
END;
$$ language 'plpgsql';

-- Function to generate audit report
CREATE OR REPLACE FUNCTION generate_audit_report(
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    event_types_filter TEXT[] DEFAULT NULL
) RETURNS TABLE (
    total_events BIGINT,
    events_by_type JSONB,
    events_by_outcome JSONB,
    top_actors JSONB,
    top_resources JSONB,
    time_series JSONB
) AS $$
BEGIN
    RETURN QUERY
    WITH event_stats AS (
        SELECT 
            COUNT(*) as total_events,
            jsonb_object_agg(event_type, count) as by_type,
            jsonb_object_agg(outcome, count) as by_outcome
        FROM audit_events
        WHERE timestamp BETWEEN start_time AND end_time
        AND (event_types_filter IS NULL OR event_type = ANY(event_types_filter))
    ),
    top_actor_stats AS (
        SELECT jsonb_agg(
            jsonb_build_object('actor_id', actor_id, 'count', cnt)
            ORDER BY cnt DESC
            LIMIT 10
        ) as top_actors
        FROM (
            SELECT actor_id, COUNT(*) as cnt
            FROM audit_events
            WHERE timestamp BETWEEN start_time AND end_time
            AND (event_types_filter IS NULL OR event_type = ANY(event_types_filter))
            GROUP BY actor_id
            ORDER BY cnt DESC
            LIMIT 10
        ) t
    ),
    top_resource_stats AS (
        SELECT jsonb_agg(
            jsonb_build_object('resource', resource, 'count', cnt)
            ORDER BY cnt DESC
            LIMIT 10
        ) as top_resources
        FROM (
            SELECT resource, COUNT(*) as cnt
            FROM audit_events
            WHERE timestamp BETWEEN start_time AND end_time
            AND (event_types_filter IS NULL OR event_type = ANY(event_types_filter))
            GROUP BY resource
            ORDER BY cnt DESC
            LIMIT 10
        ) t
    ),
    time_series_stats AS (
        SELECT jsonb_agg(
            jsonb_build_object('hour', hour, 'count', cnt)
            ORDER BY hour
        ) as time_series
        FROM (
            SELECT 
                DATE_TRUNC('hour', timestamp) as hour,
                COUNT(*) as cnt
            FROM audit_events
            WHERE timestamp BETWEEN start_time AND end_time
            AND (event_types_filter IS NULL OR event_type = ANY(event_types_filter))
            GROUP BY hour
            ORDER BY hour
        ) t
    )
    SELECT 
        es.total_events,
        es.by_type,
        es.by_outcome,
        ta.top_actors,
        tr.top_resources,
        ts.time_series
    FROM event_stats es
    CROSS JOIN top_actor_stats ta
    CROSS JOIN top_resource_stats tr
    CROSS JOIN time_series_stats ts;
END;
$$ language 'plpgsql';

-- =====================================================
-- Views
-- =====================================================

-- View for current unsealed events
CREATE OR REPLACE VIEW v_unsealed_events AS
SELECT * FROM audit_events
WHERE sealed = FALSE
ORDER BY chain_index ASC;

-- View for seal chain
CREATE OR REPLACE VIEW v_seal_chain AS
SELECT 
    s.id,
    s.timestamp,
    s.root_hash,
    s.first_event_id,
    s.last_event_id,
    s.event_count,
    s.verified,
    ae.timestamp as first_event_timestamp,
    ae.action as first_event_action,
    ae.resource as first_event_resource
FROM audit_seals s
LEFT JOIN audit_events ae ON s.first_event_id = ae.id
ORDER BY s.timestamp DESC;

-- View for compliance summary
CREATE OR REPLACE VIEW v_compliance_summary AS
SELECT 
    event_type,
    COUNT(*) as event_count,
    COUNT(CASE WHEN outcome = 'SUCCESS' THEN 1 END) as successful,
    COUNT(CASE WHEN outcome IN ('FAILURE', 'DENIED') THEN 1 END) as failed,
    COUNT(DISTINCT actor_id) as unique_actors,
    MIN(timestamp) as first_event,
    MAX(timestamp) as last_event
FROM audit_events
WHERE timestamp >= NOW() - INTERVAL '30 days'
GROUP BY event_type;

-- =====================================================
-- Row Level Security (RLS) Policies
-- =====================================================

-- Enable RLS on tables
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE forensic_investigations ENABLE ROW LEVEL SECURITY;

-- RLS policies for audit_events
CREATE POLICY "Users can view all audit events" ON audit_events
    FOR SELECT
    USING (true);

CREATE POLICY "System can insert audit events" ON audit_events
    FOR INSERT
    WITH CHECK (true);

-- RLS policies for users
CREATE POLICY "Users can view own data" ON users
    FOR SELECT
    USING (auth.uid() = id);

CREATE POLICY "Admins can view all users" ON users
    FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM users
            WHERE id = auth.uid() AND role = 'admin'
        )
    );

-- RLS policies for sessions
CREATE POLICY "Users can view own sessions" ON sessions
    FOR SELECT
    USING (user_id IN (SELECT id FROM users WHERE auth.uid() = id));

CREATE POLICY "Users can manage own sessions" ON sessions
    FOR ALL
    USING (user_id IN (SELECT id FROM users WHERE auth.uid() = id));

-- =====================================================
-- Constraints and Checks
-- =====================================================

-- Constraint for event types
ALTER TABLE audit_events
    ADD CONSTRAINT chk_event_type
    CHECK (event_type IN (
        'AUTHENTICATION', 'AUTHORIZATION', 'DATA_ACCESS', 'DATA_MODIFICATION',
        'DATA_DELETION', 'CONFIGURATION', 'SYSTEM', 'NETWORK', 'SECURITY',
        'POLICY', 'EXPORT', 'IMPORT', 'INTEGRATION', 'FORENSIC'
    ));

-- Constraint for event outcomes
ALTER TABLE audit_events
    ADD CONSTRAINT chk_event_outcome
    CHECK (outcome IN ('SUCCESS', 'FAILURE', 'DENIED', 'PARTIAL', 'PENDING', 'UNKNOWN'));

-- Constraint for actor types
ALTER TABLE audit_events
    ADD CONSTRAINT chk_actor_type
    CHECK (actor_type IN ('USER', 'SERVICE', 'SYSTEM', 'API', 'SCHEDULER', 'EXTERNAL'));

-- Constraint for policy effects
ALTER TABLE policies
    ADD CONSTRAINT chk_policy_effect
    CHECK (effect IN ('allow', 'deny'));

-- =====================================================
-- Data Retention Policies
-- =====================================================

-- Function to archive old sealed events
CREATE OR REPLACE FUNCTION archive_old_events(retention_days INTEGER DEFAULT 2555)
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER := 0;
BEGIN
    -- Move old sealed events to archive table
    INSERT INTO audit_events_archive
    SELECT * FROM audit_events
    WHERE sealed = TRUE
    AND sealed_at < NOW() - INTERVAL '1 day'
    AND timestamp < NOW() - INTERVAL CAST(retention_days AS TEXT) || ' days';

    GET DIAGNOSTICS archived_count = ROW_COUNT;

    -- Delete archived events
    DELETE FROM audit_events
    WHERE sealed = TRUE
    AND sealed_at < NOW() - INTERVAL '1 day'
    AND timestamp < NOW() - INTERVAL CAST(retention_days AS TEXT) || ' days';

    RETURN archived_count;
END;
$$ language 'plpgsql';

-- Archive table (for long-term retention)
CREATE TABLE IF NOT EXISTS audit_events_archive (
    LIKE audit_events INCLUDING ALL
);

-- Index on archive table
CREATE INDEX IF NOT EXISTS idx_audit_events_archive_timestamp 
    ON audit_events_archive(timestamp DESC);

-- =====================================================
-- Grant Permissions
-- =====================================================

-- Grant SELECT on all tables to public role (for authenticated users)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO authenticated;
GRANT INSERT, UPDATE, DELETE ON audit_events, audit_seals TO service_role;
GRANT ALL ON ALL TABLES IN SCHEMA public TO service_role;

-- Grant usage on sequences
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO authenticated;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO service_role;
