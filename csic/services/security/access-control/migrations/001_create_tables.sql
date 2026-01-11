-- Access Control Service Database Schema
-- PostgreSQL database for the CSIC Platform

-- Create database if not exists
-- CREATE DATABASE csic_platform;

-- Create schema for access control tables
CREATE SCHEMA IF NOT EXISTS access_control;

-- Policies table
CREATE TABLE IF NOT EXISTS access_control.policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    effect VARCHAR(10) NOT NULL CHECK (effect IN ('ALLOW', 'DENY')),
    priority INTEGER NOT NULL DEFAULT 0,
    conditions JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for policies
CREATE INDEX IF NOT EXISTS idx_policies_enabled ON access_control.policies(enabled);
CREATE INDEX IF NOT EXISTS idx_policies_priority ON access_control.policies(priority DESC);
CREATE INDEX IF NOT EXISTS idx_policies_name ON access_control.policies(name);

-- Audit logs table
CREATE TABLE IF NOT EXISTS access_control.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request JSONB NOT NULL,
    response JSONB NOT NULL,
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('ALLOW', 'DENY', 'NOT_APPLICABLE', 'INDETERMINATE')),
    user_id VARCHAR(255),
    resource_id VARCHAR(255),
    action VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for audit logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON access_control.audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON access_control.audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON access_control.audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_decision ON access_control.audit_logs(decision);

-- Resource ownership table
CREATE TABLE IF NOT EXISTS access_control.resource_ownership (
    resource_id VARCHAR(255) PRIMARY KEY,
    owner_id VARCHAR(255) NOT NULL,
    owner_type VARCHAR(50) NOT NULL CHECK (owner_type IN ('user', 'group', 'service')),
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for resource ownership
CREATE INDEX IF NOT EXISTS idx_resource_ownership_owner_id ON access_control.resource_ownership(owner_id);

-- User roles table
CREATE TABLE IF NOT EXISTS access_control.user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('global', 'organization', 'project', 'resource')),
    scope_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, role_id, scope, scope_id)
);

-- Create indexes for user roles
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON access_control.user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON access_control.user_roles(role_id);

-- Roles table
CREATE TABLE IF NOT EXISTS access_control.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    policy_ids JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for roles
CREATE INDEX IF NOT EXISTS idx_roles_name ON access_control.roles(name);

-- Add foreign key constraints
ALTER TABLE access_control.user_roles 
    ADD CONSTRAINT fk_user_roles_role 
    FOREIGN KEY (role_id) 
    REFERENCES access_control.roles(id)
    ON DELETE CASCADE;

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION access_control.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_policies_updated_at
    BEFORE UPDATE ON access_control.policies
    FOR EACH ROW
    EXECUTE FUNCTION access_control.update_updated_at_column();

CREATE TRIGGER update_resource_ownership_updated_at
    BEFORE UPDATE ON access_control.resource_ownership
    FOR EACH ROW
    EXECUTE FUNCTION access_control.update_updated_at_column();

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON access_control.roles
    FOR EACH ROW
    EXECUTE FUNCTION access_control.update_updated_at_column();

-- Insert default policies
INSERT INTO access_control.policies (id, name, description, effect, priority, conditions, enabled)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'Deny All', 'Default deny policy for unmatched requests', 'DENY', 0, '{}', true),
    ('00000000-0000-0000-0000-000000000002', 'Admin Full Access', 'Administrators have full access', 'ALLOW', 100, 
     '{"subjects": [{"roles": ["admin"]}], "resources": [{"types": ["*"]}], "actions": [{"operations": ["*"]}]}', true)
ON CONFLICT (id) DO NOTHING;

-- Insert default roles
INSERT INTO access_control.roles (id, name, description, permissions, policy_ids)
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'admin', 'Administrator with full access', '["*"]', '["00000000-0000-0000-0000-000000000002"]'),
    ('00000000-0000-0000-0000-000000000002', 'user', 'Regular user with basic access', '["read"]', '[]')
ON CONFLICT (id) DO NOTHING;
