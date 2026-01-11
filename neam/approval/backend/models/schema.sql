-- Multi-Level Approval System Database Schema
-- This schema implements an immutable approval ledger with cryptographic hashing

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- Core Tables
-- ============================================

-- Approval Definitions (Policy Templates)
CREATE TABLE IF NOT EXISTS approval_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL, -- 'financial', 'operational', 'security'
    policy_config JSONB NOT NULL DEFAULT '{}',
    version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    CONSTRAINT chk_policy_config CHECK (
        (policy_config->>'quorum_required')::INT > 0 AND
        (policy_config->>'total_approvers')::INT >= (policy_config->>'quorum_required')::INT
    )
);

CREATE INDEX IF NOT EXISTS idx_definitions_category ON approval_definitions(category);
CREATE INDEX IF NOT EXISTS idx_definitions_active ON approval_definitions(is_active);

-- Approval Requests (Active Requests)
CREATE TABLE IF NOT EXISTS approval_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    definition_id UUID NOT NULL REFERENCES approval_definitions(id),
    requester_id UUID NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    context_data JSONB, -- Snapshot of request data
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    current_step INT NOT NULL DEFAULT 1,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    workflow_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_status CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED', 'CANCELLED'))
);

CREATE INDEX IF NOT EXISTS idx_requests_requester ON approval_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_requests_status ON approval_requests(status);
CREATE INDEX IF NOT EXISTS idx_requests_deadline ON approval_requests(deadline);
CREATE INDEX IF NOT EXISTS idx_requests_resource ON approval_requests(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_requests_workflow ON approval_requests(workflow_id);

-- ============================================
-- Immutable Ledger Tables
-- ============================================

-- Approval Actions (Immutable Action Log with Hash Chaining)
CREATE TABLE IF NOT EXISTS approval_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL,
    actor_type VARCHAR(20) NOT NULL DEFAULT 'user', -- 'user', 'system', 'service'
    action_type VARCHAR(20) NOT NULL,
    on_behalf_of UUID, -- For delegations
    step_number INT NOT NULL DEFAULT 1,
    comments TEXT,
    signature_hash VARCHAR(128) NOT NULL, -- SHA-512 hash
    previous_hash VARCHAR(128), -- Hash chaining for immutability
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_action_type CHECK (action_type IN ('APPROVE', 'REJECT', 'ABSTAIN'))
);

-- Index for hash chain verification
CREATE INDEX IF NOT EXISTS idx_actions_request ON approval_actions(request_id);
CREATE INDEX IF NOT EXISTS idx_actions_actor ON approval_actions(actor_id);
CREATE INDEX IF NOT EXISTS idx_actions_timestamp ON approval_actions(timestamp);

-- Create trigger to prevent updates/deletes on immutable ledger
CREATE OR REPLACE FUNCTION prevent_approval_action_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'approval_actions table is immutable - no modifications allowed';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_immutable_approval_actions ON approval_actions;
CREATE TRIGGER trg_immutable_approval_actions
    BEFORE UPDATE OR DELETE ON approval_actions
    FOR EACH ROW
    EXECUTE FUNCTION prevent_approval_action_modification();

-- ============================================
-- Support Tables
-- ============================================

-- Delegations
CREATE TABLE IF NOT EXISTS delegations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delegator_id UUID NOT NULL,
    delegatee_id UUID NOT NULL,
    definition_id UUID REFERENCES approval_definitions(id), -- NULL means all types
    scope JSONB DEFAULT '{}', -- Limitations on delegation
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_delegation_time CHECK (valid_until > valid_from)
);

CREATE INDEX IF NOT EXISTS idx_delegations_delegator ON delegations(delegator_id);
CREATE INDEX IF NOT EXISTS idx_delegations_delegatee ON delegations(delegatee_id);
CREATE INDEX IF NOT EXISTS idx_delegations_valid ON delegations(valid_from, valid_until);

-- Approver Assignments
CREATE TABLE IF NOT EXISTS approver_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL,
    role VARCHAR(100) NOT NULL,
    step_order INT NOT NULL DEFAULT 1,
    has_acted BOOLEAN NOT NULL DEFAULT false,
    acted_at TIMESTAMP WITH TIME ZONE,
    action_type VARCHAR(20),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_assignment_action CHECK (
        (has_acted = false AND action_type IS NULL) OR
        (has_acted = true AND action_type IN ('APPROVE', 'REJECT', 'ABSTAIN'))
    )
);

CREATE INDEX IF NOT EXISTS idx_assignments_request ON approver_assignments(request_id);
CREATE INDEX IF NOT EXISTS idx_assignments_approver ON approver_assignments(approver_id);

-- ============================================
-- View for Quorum Status
-- ============================================

CREATE OR REPLACE VIEW v_approval_quorum_status AS
SELECT 
    ar.id AS request_id,
    ar.status,
    ar.deadline,
    COUNT(aa.id) FILTER (WHERE aa.has_acted) AS acted_count,
    COUNT(aa.id) AS total_approvers,
    COUNT(aa.id) FILTER (WHERE aa.has_acted AND aa.action_type = 'APPROVE') AS approval_count,
    COUNT(aa.id) FILTER (WHERE aa.has_acted AND aa.action_type = 'REJECT') AS rejection_count,
    COUNT(aa.id) FILTER (WHERE NOT aa.has_acted) AS pending_count,
    (ad.policy_config->>'quorum_required')::INT AS quorum_required,
    (ad.policy_config->>'total_approvers')::INT AS total_required,
    CASE 
        WHEN ar.status = 'PENDING' THEN 
            ar.deadline > NOW()
        ELSE NULL 
    END AS is_within_deadline
FROM approval_requests ar
JOIN approval_definitions ad ON ar.definition_id = ad.id
LEFT JOIN approver_assignments aa ON ar.id = aa.request_id
GROUP BY ar.id, ar.status, ar.deadline, ad.policy_config;

-- ============================================
-- Function for Hash Chain Calculation
-- ============================================

CREATE OR REPLACE FUNCTION calculate_action_hash(
    p_request_id UUID,
    p_actor_id UUID,
    p_action_type VARCHAR,
    p_timestamp TIMESTAMP WITH TIME ZONE,
    p_previous_hash VARCHAR
) RETURNS VARCHAR AS $$
DECLARE
    v_payload TEXT;
    v_hash VARCHAR;
BEGIN
    -- Concatenate fields for hashing
    v_payload := p_request_id::TEXT || p_actor_id::TEXT || p_action_type || 
                p_timestamp::TEXT || COALESCE(p_previous_hash, '');
    
    -- Generate SHA-512 hash
    v_hash := ENCODE(SHA512(v_payload::BYTEA), 'hex');
    
    RETURN v_hash;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- Function to Verify Hash Chain Integrity
-- ============================================

CREATE OR REPLACE FUNCTION verify_hash_chain(p_request_id UUID)
RETURNS TABLE (
    is_valid BOOLEAN,
    broken_at_action UUID,
    expected_hash VARCHAR,
    actual_hash VARCHAR
) AS $$
DECLARE
    v_previous_hash VARCHAR := NULL;
    v_expected_hash VARCHAR;
    v_actual_hash VARCHAR;
    v_action_id UUID;
BEGIN
    FOR v_action_id, v_actual_hash IN 
        SELECT aa.id, aa.signature_hash 
        FROM approval_actions aa 
        WHERE aa.request_id = p_request_id 
        ORDER BY aa.timestamp ASC
    LOOP
        v_expected_hash := calculate_action_hash(
            p_request_id,
            (SELECT actor_id FROM approval_actions WHERE id = v_action_id),
            (SELECT action_type FROM approval_actions WHERE id = v_action_id),
            (SELECT timestamp FROM approval_actions WHERE id = v_action_id),
            v_previous_hash
        );
        
        IF v_actual_hash != v_expected_hash THEN
            is_valid := false;
            broken_at_action := v_action_id;
            expected_hash := v_expected_hash;
            actual_hash := v_actual_hash;
            RETURN NEXT;
            RETURN;
        END IF;
        
        v_previous_hash := v_actual_hash;
    END LOOP;
    
    is_valid := true;
    broken_at_action := NULL;
    expected_hash := NULL;
    actual_hash := NULL;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Notification Queue Table
-- ============================================

CREATE TABLE IF NOT EXISTS approval_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL,
    notification_type VARCHAR(50) NOT NULL, -- 'approval_required', 'approved', 'rejected', 'expired'
    channel VARCHAR(20) NOT NULL, -- 'email', 'slack', 'in_app'
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    sent_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_notification_status CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'CANCELLED'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_request ON approval_notifications(request_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON approval_notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON approval_notifications(recipient_id);
