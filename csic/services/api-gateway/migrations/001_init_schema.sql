-- CSIC Platform - API Gateway Database Schema
-- PostgreSQL migration for api-gateway service

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'VIEWER',
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    password VARCHAR(255) NOT NULL,
    last_login TIMESTAMP,
    mfa_enabled BOOLEAN NOT NULL DEFAULT false,
    permissions TEXT[] DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Exchanges table
CREATE TABLE IF NOT EXISTS exchanges (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    jurisdiction VARCHAR(100),
    website VARCHAR(500),
    contact_email VARCHAR(255),
    compliance_score INTEGER DEFAULT 0,
    risk_level VARCHAR(20) DEFAULT 'LOW',
    registration_date DATE,
    last_audit DATE,
    next_audit DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_exchanges_status ON exchanges(status);
CREATE INDEX IF NOT EXISTS idx_exchanges_license ON exchanges(license_number);
CREATE INDEX IF NOT EXISTS idx_exchanges_jurisdiction ON exchanges(jurisdiction);

-- Wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(500) NOT NULL,
    label VARCHAR(255),
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    risk_score INTEGER DEFAULT 0,
    first_seen TIMESTAMP,
    last_activity TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE INDEX IF NOT EXISTS idx_wallets_status ON wallets(status);
CREATE INDEX IF NOT EXISTS idx_wallets_type ON wallets(type);

-- Miners table
CREATE TABLE IF NOT EXISTS miners (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    jurisdiction VARCHAR(100),
    hash_rate DECIMAL(20, 2),
    energy_consumption DECIMAL(20, 2),
    energy_source VARCHAR(100),
    compliance_status VARCHAR(50) DEFAULT 'UNKNOWN',
    registration_date DATE,
    last_inspection DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_miners_status ON miners(status);
CREATE INDEX IF NOT EXISTS idx_miners_license ON miners(license_number);
CREATE INDEX IF NOT EXISTS idx_miners_jurisdiction ON miners(jurisdiction);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    category VARCHAR(100),
    source VARCHAR(100),
    evidence JSONB DEFAULT '[]',
    acknowledged_by VARCHAR(36),
    acknowledged_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_category ON alerts(category);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);

-- Compliance reports table
CREATE TABLE IF NOT EXISTS compliance_reports (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(36) NOT NULL,
    entity_name VARCHAR(255) NOT NULL,
    period VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    score INTEGER DEFAULT 0,
    generated_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_compliance_entity ON compliance_reports(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_compliance_status ON compliance_reports(status);
CREATE INDEX IF NOT EXISTS idx_compliance_period ON compliance_reports(period);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(36),
    username VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    details TEXT,
    ip_address VARCHAR(45),
    status VARCHAR(20) DEFAULT 'SUCCESS',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Insert default admin user (password: admin123)
INSERT INTO users (id, username, email, role, status, password, mfa_enabled, permissions)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'admin@csic.gov',
    'ADMIN',
    'ACTIVE',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.PGblOYNBlxNXkD4e.G', -- admin123
    false,
    ARRAY['*']
) ON CONFLICT (username) DO NOTHING;

-- Insert sample exchanges
INSERT INTO exchanges (id, name, license_number, status, jurisdiction, compliance_score, risk_level, registration_date)
VALUES
    ('ex_001', 'CryptoExchange Pro', 'CSIC-2024-001', 'ACTIVE', 'Singapore', 92, 'LOW', '2024-01-15'),
    ('ex_002', 'Digital Asset Hub', 'CSIC-2024-002', 'ACTIVE', 'Switzerland', 88, 'LOW', '2024-02-20'),
    ('ex_003', 'BlockTrade Global', 'CSIC-2024-003', 'SUSPENDED', 'British Virgin Islands', 45, 'HIGH', '2024-03-10')
ON CONFLICT (license_number) DO NOTHING;

-- Insert sample wallets
INSERT INTO wallets (id, address, label, type, status, risk_score, first_seen)
VALUES
    ('w_001', '1A2B3C4D5E6F7890ABCDEF1234567890', '可疑地址 #4521', 'MIXER', 'BLACKLISTED', 95, '2023-06-15'),
    ('w_002', '0x1234567890ABCDEF1234567890ABCDEF12345678', '交易所热钱包 B', 'EXCHANGE', 'ACTIVE', 15, '2024-01-20')
ON CONFLICT DO NOTHING;

-- Insert sample miners
INSERT INTO miners (id, name, license_number, status, jurisdiction, hash_rate, energy_consumption, energy_source, compliance_status)
VALUES
    ('m_001', 'Northern Mining Pool', 'CSIC-MIN-2024-001', 'ACTIVE', 'Canada', 450.00, 520, 'Hydroelectric', 'COMPLIANT'),
    ('m_002', 'GreenHash Energy', 'CSIC-MIN-2024-002', 'ACTIVE', 'Iceland', 320.00, 380, 'Geothermal', 'COMPLIANT')
ON CONFLICT (license_number) DO NOTHING;

-- Insert sample alerts
INSERT INTO alerts (id, title, description, severity, status, category, source)
VALUES
    ('alert_001', '高风险交易模式检测', '检测到与已知混币服务相关的异常交易模式', 'CRITICAL', 'ACTIVE', '交易监控', '风险引擎'),
    ('alert_002', '交易所合规分数下降', 'CryptoExchange Pro 的合规分数从 92 下降到 78', 'WARNING', 'ACTIVE', '合规监控', '合规系统')
ON CONFLICT DO NOTHING;
