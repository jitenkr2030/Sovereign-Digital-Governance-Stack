-- Control Layer Service Database Schema
-- PostgreSQL migrations for policies, enforcements, and interventions

-- Create policies table
CREATE TABLE IF NOT EXISTS control_layer_policies (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rule_json JSONB NOT NULL,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index on policies for faster lookups
CREATE INDEX IF NOT EXISTS idx_control_layer_policies_active 
ON control_layer_policies(is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_control_layer_policies_priority 
ON control_layer_policies(priority DESC);

CREATE INDEX IF NOT EXISTS idx_control_layer_policies_created_at 
ON control_layer_policies(created_at DESC);

-- Create enforcements table
CREATE TABLE IF NOT EXISTS control_layer_enforcements (
    id UUID PRIMARY KEY,
    policy_id UUID REFERENCES control_layer_policies(id),
    target_service VARCHAR(255) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes on enforcements
CREATE INDEX IF NOT EXISTS idx_control_layer_enforcements_policy_id 
ON control_layer_enforcements(policy_id);

CREATE INDEX IF NOT EXISTS idx_control_layer_enforcements_target_service 
ON control_layer_enforcements(target_service);

CREATE INDEX IF NOT EXISTS idx_control_layer_enforcements_status 
ON control_layer_enforcements(status);

CREATE INDEX IF NOT EXISTS idx_control_layer_enforcements_created_at 
ON control_layer_enforcements(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_control_layer_enforcements_policy_status 
ON control_layer_enforcements(policy_id, status);

-- Create interventions table
CREATE TABLE IF NOT EXISTS control_layer_interventions (
    id UUID PRIMARY KEY,
    policy_id UUID REFERENCES control_layer_policies(id),
    enforcement_id UUID,
    target_service VARCHAR(255) NOT NULL,
    intervention_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reason TEXT,
    resolution TEXT,
    metadata JSONB,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes on interventions
CREATE INDEX IF NOT EXISTS idx_control_layer_interventions_policy_id 
ON control_layer_interventions(policy_id);

CREATE INDEX IF NOT EXISTS idx_control_layer_interventions_enforcement_id 
ON control_layer_interventions(enforcement_id);

CREATE INDEX IF NOT EXISTS idx_control_layer_interventions_target_service 
ON control_layer_interventions(target_service);

CREATE INDEX IF NOT EXISTS idx_control_layer_interventions_status 
ON control_layer_interventions(status);

CREATE INDEX IF NOT EXISTS idx_control_layer_interventions_severity 
ON control_layer_interventions(severity);

CREATE INDEX IF NOT EXISTS idx_control_layer_interventions_created_at 
ON control_layer_interventions(created_at DESC);

-- Create state registry table (for tracking system state)
CREATE TABLE IF NOT EXISTS control_layer_state_registry (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    ttl TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create audit log table
CREATE TABLE IF NOT EXISTS control_layer_audit_log (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor VARCHAR(255),
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_control_layer_audit_log_entity 
ON control_layer_audit_log(entity_type, entity_id);

CREATE INDEX IF NOT EXISTS idx_control_layer_audit_log_created_at 
ON control_layer_audit_log(created_at DESC);

-- Insert default policies
INSERT INTO control_layer_policies (id, name, description, rule_json, priority, is_active) VALUES
('11111111-1111-1111-1111-111111111111', 'CPU Usage Policy', 'Alert when CPU usage exceeds threshold', 
 '{"condition": "cpu_usage > 80", "target": "metrics.cpu_percent", "threshold": 80, "operator": ">", "action": "alert", "alert_severity": "high"}', 
 100, true),
('22222222-2222-2222-2222-222222222222', 'Memory Usage Policy', 'Alert when memory usage exceeds threshold', 
 '{"condition": "memory_usage > 85", "target": "metrics.memory_percent", "threshold": 85, "operator": ">", "action": "alert", "alert_severity": "critical"}', 
 100, true),
('33333333-3333-3333-3333-333333333333', 'Disk Usage Policy', 'Alert when disk usage exceeds threshold', 
 '{"condition": "disk_usage > 90", "target": "metrics.disk_percent", "threshold": 90, "operator": ">", "action": "alert", "alert_severity": "warning"}', 
 50, true),
('44444444-4444-4444-4444-444444444444', 'Error Rate Policy', 'Alert when error rate exceeds threshold', 
 '{"condition": "error_rate > 5", "target": "metrics.error_rate_percent", "threshold": 5, "operator": ">", "action": "enforce", "alert_severity": "high"}', 
 150, true),
('55555555-5555-5555-5555-555555555555', 'Latency Policy', 'Alert when latency exceeds threshold', 
 '{"condition": "latency > 500", "target": "metrics.latency_ms", "threshold": 500, "operator": ">", "action": "alert", "alert_severity": "medium"}', 
 75, true)
ON CONFLICT (id) DO NOTHING;
