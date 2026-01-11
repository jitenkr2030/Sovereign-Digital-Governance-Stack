-- Transaction Monitoring Service Database Schema
-- Initial migration

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(64) PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL UNIQUE,
    network VARCHAR(20) NOT NULL,
    block_number BIGINT,
    block_hash VARCHAR(66),
    timestamp TIMESTAMP NOT NULL,
    sender VARCHAR(64),
    receiver VARCHAR(64),
    amount DECIMAL(36, 18) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    gas_used DECIMAL(36, 18),
    gas_price DECIMAL(36, 18),
    fee DECIMAL(36, 18),
    status VARCHAR(20) DEFAULT 'confirmed',
    input_count INTEGER DEFAULT 0,
    output_count INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender);
CREATE INDEX IF NOT EXISTS idx_transactions_receiver ON transactions(receiver);
CREATE INDEX IF NOT EXISTS idx_transactions_network ON transactions(network);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_block ON transactions(network, block_number);

-- Wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id VARCHAR(64) PRIMARY KEY,
    address VARCHAR(64) NOT NULL,
    network VARCHAR(20) NOT NULL,
    wallet_type VARCHAR(20) DEFAULT 'EOA',
    label VARCHAR(255),
    tags TEXT[] DEFAULT '{}',
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    tx_count BIGINT DEFAULT 0,
    total_received DECIMAL(36, 18) DEFAULT 0,
    total_sent DECIMAL(36, 18) DEFAULT 0,
    current_balance DECIMAL(36, 18) DEFAULT 0,
    risk_score DECIMAL(5, 2) DEFAULT 0,
    risk_level VARCHAR(20) DEFAULT 'low',
    is_sanctioned BOOLEAN DEFAULT FALSE,
    is_blacklisted BOOLEAN DEFAULT FALSE,
    is_whitelisted BOOLEAN DEFAULT FALSE,
    cluster_id VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(address, network)
);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address);
CREATE INDEX IF NOT EXISTS idx_wallets_network ON wallets(network);
CREATE INDEX IF NOT EXISTS idx_wallets_risk_score ON wallets(risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_wallets_cluster ON wallets(cluster_id);
CREATE INDEX IF NOT EXISTS idx_wallets_sanctioned ON wallets(is_sanctioned);
CREATE INDEX IF NOT EXISTS idx_wallets_blacklisted ON wallets(is_blacklisted);

-- Entity clusters table
CREATE TABLE IF NOT EXISTS entity_clusters (
    id VARCHAR(64) PRIMARY KEY,
    cluster_type VARCHAR(50) NOT NULL,
    primary_address VARCHAR(64) NOT NULL,
    label VARCHAR(255),
    description TEXT,
    wallet_count INTEGER DEFAULT 1,
    total_volume DECIMAL(36, 18) DEFAULT 0,
    total_tx_count BIGINT DEFAULT 0,
    risk_score DECIMAL(5, 2) DEFAULT 0,
    risk_level VARCHAR(20) DEFAULT 'low',
    confidence_score DECIMAL(3, 2) DEFAULT 1.0,
    tags TEXT[] DEFAULT '{}',
    is_verified BOOLEAN DEFAULT FALSE,
    verified_by VARCHAR(64),
    suspected_entity VARCHAR(255),
    discovery_method VARCHAR(50),
    first_seen TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clusters_risk_score ON entity_clusters(risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_clusters_type ON entity_clusters(cluster_type);
CREATE INDEX IF NOT EXISTS idx_clusters_verified ON entity_clusters(is_verified);

-- Cluster members table
CREATE TABLE IF NOT EXISTS cluster_members (
    id VARCHAR(64) PRIMARY KEY,
    cluster_id VARCHAR(64) NOT NULL REFERENCES entity_clusters(id),
    wallet_address VARCHAR(64) NOT NULL,
    network VARCHAR(20) NOT NULL,
    member_type VARCHAR(20) DEFAULT 'linked',
    link_type VARCHAR(50) NOT NULL,
    link_strength DECIMAL(3, 2) DEFAULT 1.0,
    tx_count BIGINT DEFAULT 0,
    volume DECIMAL(36, 18) DEFAULT 0,
    first_linked_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(cluster_id, wallet_address)
);

CREATE INDEX IF NOT EXISTS idx_cluster_members_cluster ON cluster_members(cluster_id);
CREATE INDEX IF NOT EXISTS idx_cluster_members_address ON cluster_members(wallet_address);

-- Cluster relationships table
CREATE TABLE IF NOT EXISTS cluster_relationships (
    id VARCHAR(64) PRIMARY KEY,
    source_cluster_id VARCHAR(64) NOT NULL REFERENCES entity_clusters(id),
    target_cluster_id VARCHAR(64) NOT NULL REFERENCES entity_clusters(id),
    relationship_type VARCHAR(50) NOT NULL,
    tx_count BIGINT DEFAULT 0,
    total_volume DECIMAL(36, 18) DEFAULT 0,
    first_seen TIMESTAMP NOT NULL,
    last_seen TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(source_cluster_id, target_cluster_id, relationship_type)
);

CREATE INDEX IF NOT EXISTS idx_cluster_rel_source ON cluster_relationships(source_cluster_id);
CREATE INDEX IF NOT EXISTS idx_cluster_rel_target ON cluster_relationships(target_cluster_id);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id VARCHAR(64) PRIMARY KEY,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    target_type VARCHAR(20) NOT NULL,
    target_id VARCHAR(64) NOT NULL,
    target_value VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    rule_id VARCHAR(64),
    rule_name VARCHAR(100),
    triggered_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    assigned_to VARCHAR(64),
    assigned_team VARCHAR(100),
    case_id VARCHAR(64),
    score DECIMAL(5, 2) DEFAULT 0,
    evidence JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_alerts_target ON alerts(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered ON alerts(triggered_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_assigned ON alerts(assigned_to);

-- Alert notes table
CREATE TABLE IF NOT EXISTS alert_notes (
    id VARCHAR(64) PRIMARY KEY,
    alert_id VARCHAR(64) NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    author VARCHAR(64) NOT NULL,
    author_role VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_notes_alert ON alert_notes(alert_id);

-- Alert evidence table
CREATE TABLE IF NOT EXISTS alert_evidence (
    id VARCHAR(64) PRIMARY KEY,
    alert_id VARCHAR(64) NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    evidence_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(64),
    description TEXT,
    data TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_evidence_alert ON alert_evidence(alert_id);

-- Alert rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50) NOT NULL,
    description TEXT,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    conditions JSONB NOT NULL,
    parameters JSONB DEFAULT '{}',
    risk_weight DECIMAL(5, 2) DEFAULT 0,
    cooldown_secs INTEGER DEFAULT 0,
    last_triggered TIMESTAMP,
    trigger_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_category ON alert_rules(category);

-- Wallet history table (for historical metrics)
CREATE TABLE IF NOT EXISTS wallet_history (
    id VARCHAR(64) PRIMARY KEY,
    wallet_id VARCHAR(64) NOT NULL REFERENCES wallets(id),
    date DATE NOT NULL,
    tx_count BIGINT DEFAULT 0,
    volume DECIMAL(36, 18) DEFAULT 0,
    risk_score DECIMAL(5, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(wallet_id, date)
);

CREATE INDEX IF NOT EXISTS idx_wallet_history_wallet ON wallet_history(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_history_date ON wallet_history(date DESC);

-- Wallet activity table (for monitoring)
CREATE TABLE IF NOT EXISTS wallet_activity (
    id VARCHAR(64) PRIMARY KEY,
    wallet_id VARCHAR(64) NOT NULL REFERENCES wallets(id),
    activity_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(64),
    amount DECIMAL(36, 18),
    network VARCHAR(20),
    details TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_activity_wallet ON wallet_activity(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_activity_type ON wallet_activity(activity_type);
CREATE INDEX IF NOT EXISTS idx_wallet_activity_created ON wallet_activity(created_at DESC);

-- Sanctions screening results table
CREATE TABLE IF NOT EXISTS sanctions_screening (
    id VARCHAR(64) PRIMARY KEY,
    wallet_address VARCHAR(64) NOT NULL,
    network VARCHAR(20) NOT NULL,
    is_sanctioned BOOLEAN DEFAULT FALSE,
    matched_list VARCHAR(20),
    matched_entity VARCHAR(255),
    match_type VARCHAR(20),
    match_score DECIMAL(3, 2),
    screened_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sanctions_address ON sanctions_screening(wallet_address);
CREATE INDEX IF NOT EXISTS idx_sanctions_sanctioned ON sanctions_screening(is_sanctioned);
CREATE INDEX IF NOT EXISTS idx_sanctions_screened ON sanctions_screening(screened_at DESC);

-- Sanctions entities table (cached from OFAC, UN, EU lists)
CREATE TABLE IF NOT EXISTS sanctions_entities (
    id VARCHAR(64) PRIMARY KEY,
    list_source VARCHAR(20) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50),
    program VARCHAR(100),
    remarks TEXT,
    addresses TEXT[] DEFAULT '{}',
    aliases TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(list_source, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_sanctions_entities_name ON sanctions_entities(name);
CREATE INDEX IF NOT EXISTS idx_sanctions_entities_list ON sanctions_entities(list_source);

-- Risk score history table
CREATE TABLE IF NOT EXISTS risk_score_history (
    id VARCHAR(64) PRIMARY KEY,
    wallet_id VARCHAR(64) NOT NULL REFERENCES wallets(id),
    score DECIMAL(5, 2) NOT NULL,
    factors JSONB DEFAULT '{}',
    triggered_rules TEXT[] DEFAULT '{}',
    calculated_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(wallet_id, calculated_at)
);

CREATE INDEX IF NOT EXISTS idx_risk_history_wallet ON risk_score_history(wallet_id);
CREATE INDEX IF NOT EXISTS idx_risk_history_calculated ON risk_score_history(calculated_at DESC);

-- Insert default alert rules
INSERT INTO alert_rules (id, name, category, description, alert_type, severity, enabled, conditions, parameters, risk_weight)
VALUES
    ('rule-001', 'high_risk_score', 'risk', 'Triggered when wallet risk score exceeds threshold', 'high_risk_score', 'high', TRUE,
     '{"operator": "gte", "field": "risk_score", "value": 75}', '{"threshold": 75}', 25),
    ('rule-002', 'sanctions_match', 'compliance', 'Triggered when address matches sanctions list', 'sanctions_match', 'critical', TRUE,
     '{"operator": "eq", "field": "is_sanctioned", "value": true}', '{}', 100),
    ('rule-003', 'rapid_movement', 'velocity', 'Triggered when high velocity detected', 'rapid_movement', 'high', TRUE,
     '{"operator": "gt", "field": "velocity_24h", "value": 100000}', '{"threshold": 100000}', 30),
    ('rule-004', 'structuring', 'pattern', 'Triggered when structuring patterns detected', 'structuring', 'high', TRUE,
     '{"operator": "contains", "field": "pattern", "value": "structuring"}', '{}', 40),
    ('rule-005', 'new_high_risk', 'behavioral', 'Triggered when new wallet shows high risk', 'new_high_risk_entity', 'medium', TRUE,
     '{"operator": "and", "conditions": [{"operator": "lt", "field": "wallet_age_days", "value": 7}, {"operator": "gte", "field": "risk_score", "value": 60}]}',
     '{"max_age_days": 7, "risk_threshold": 60}', 20)
ON CONFLICT (name) DO NOTHING;
