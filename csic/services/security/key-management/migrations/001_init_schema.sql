-- CSIC Platform - Key Management Service Database Schema
-- PostgreSQL migration for key-management service

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Key vault table (stores key metadata)
CREATE TABLE IF NOT EXISTS key_vault (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    alias VARCHAR(255) UNIQUE NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    current_version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    usage TEXT[] NOT NULL DEFAULT '{}',
    description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rotated_at TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_key_vault_alias ON key_vault(alias);
CREATE INDEX IF NOT EXISTS idx_key_vault_status ON key_vault(status);
CREATE INDEX IF NOT EXISTS idx_key_vault_algorithm ON key_vault(algorithm);
CREATE INDEX IF NOT EXISTS idx_key_vault_created_at ON key_vault(created_at DESC);

-- Key versions table (stores encrypted key material)
CREATE TABLE IF NOT EXISTS key_versions (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id VARCHAR(36) NOT NULL REFERENCES key_vault(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    encrypted_material BYTEA NOT NULL,
    iv BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(key_id, version)
);

CREATE INDEX IF NOT EXISTS idx_key_versions_key_id ON key_versions(key_id);
CREATE INDEX IF NOT EXISTS idx_key_versions_version ON key_versions(key_id, version DESC);

-- Key policies table (access control for keys)
CREATE TABLE IF NOT EXISTS key_policies (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id VARCHAR(36) NOT NULL UNIQUE REFERENCES key_vault(id) ON DELETE CASCADE,
    allowed_roles TEXT[] NOT NULL DEFAULT '{}',
    allowed_services TEXT[] NOT NULL DEFAULT '{}',
    max_operations INTEGER DEFAULT 0,
    rate_limit INTEGER DEFAULT 1000  -- requests per minute
);

CREATE INDEX IF NOT EXISTS idx_key_policies_key_id ON key_policies(key_id);

-- Key audit log table (complete audit trail)
CREATE TABLE IF NOT EXISTS key_audit_log (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_id VARCHAR(36) NOT NULL REFERENCES key_vault(id) ON DELETE CASCADE,
    operation VARCHAR(50) NOT NULL,
    actor_id VARCHAR(36),
    service_id VARCHAR(36),
    success BOOLEAN NOT NULL DEFAULT true,
    error TEXT,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_key_audit_log_key_id ON key_audit_log(key_id);
CREATE INDEX IF NOT EXISTS idx_key_audit_log_timestamp ON key_audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_key_audit_log_operation ON key_audit_log(operation);
CREATE INDEX IF NOT EXISTS idx_key_audit_log_actor_id ON key_audit_log(actor_id);

-- Insert sample keys for testing
INSERT INTO key_vault (id, alias, algorithm, current_version, status, usage, description)
VALUES
    ('kms-key-001', 'master-encryption-key', 'AES-256-GCM', 1, 'ACTIVE', ARRAY['encrypt', 'decrypt'], 'Master key for data encryption'),
    ('kms-key-002', 'data-signing-key', 'ECDSA-P384', 1, 'ACTIVE', ARRAY['sign', 'verify'], 'Key for digital signatures'),
    ('kms-key-003', 'api-auth-key', 'Ed25519', 1, 'ACTIVE', ARRAY['sign'], 'Key for API authentication')
ON CONFLICT (alias) DO NOTHING;

-- Insert sample key versions (encrypted with a test master key)
-- Note: In production, these would be encrypted with the actual master key
INSERT INTO key_versions (key_id, version, encrypted_material, iv)
SELECT 
    id,
    1,
    decode('a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456', 'hex'),
    decode('0123456789ab', 'hex')
FROM key_vault WHERE id = 'kms-key-001'
ON CONFLICT DO NOTHING;

INSERT INTO key_versions (key_id, version, encrypted_material, iv)
SELECT 
    id,
    1,
    decode('c1d2e3f4a5b6789012345678901234567890abcdef1234567890abcdef123456', 'hex'),
    decode('fedcba987654', 'hex')
FROM key_vault WHERE id = 'kms-key-002'
ON CONFLICT DO NOTHING;

INSERT INTO key_versions (key_id, version, encrypted_material, iv)
SELECT 
    id,
    1,
    decode('d1e2f3a4b5c6789012345678901234567890abcdef1234567890abcdef123456', 'hex'),
    decode('abcdef123456', 'hex')
FROM key_vault WHERE id = 'kms-key-003'
ON CONFLICT DO NOTHING;

-- Insert sample audit log entries
INSERT INTO key_audit_log (key_id, operation, actor_id, success, metadata)
VALUES
    ('kms-key-001', 'CREATE', 'system', true, '{"description": "Initial key creation"}'),
    ('kms-key-002', 'CREATE', 'system', true, '{"description": "Initial key creation"}'),
    ('kms-key-003', 'CREATE', 'system', true, '{"description": "Initial key creation"}')
ON CONFLICT DO NOTHING;
