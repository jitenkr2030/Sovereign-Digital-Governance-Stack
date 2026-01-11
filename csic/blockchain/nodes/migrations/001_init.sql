-- Blockchain Node Manager Database Migrations

-- Nodes table
CREATE TABLE IF NOT EXISTS nodes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    network VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'full',
    rpc_url VARCHAR(500) NOT NULL,
    ws_url VARCHAR(500),
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    version VARCHAR(100),
    chain_id BIGINT,
    block_number BIGINT DEFAULT 0,
    peer_count INT DEFAULT 0,
    last_sync_time TIMESTAMP,
    last_health_check TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Node metrics table
CREATE TABLE IF NOT EXISTS node_metrics (
    id VARCHAR(36) PRIMARY KEY,
    node_id VARCHAR(36) NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    block_height BIGINT,
    block_time DOUBLE PRECISION,
    peer_count INT,
    transaction_pool_size INT,
    gas_price DOUBLE PRECISION,
    memory_usage BIGINT,
    cpu_usage DOUBLE PRECISION,
    disk_usage BIGINT,
    network_in BIGINT,
    network_out BIGINT,
    rpc_latency DOUBLE PRECISION,
    sync_progress DOUBLE PRECISION
);

-- Node events table
CREATE TABLE IF NOT EXISTS node_events (
    id VARCHAR(36) PRIMARY KEY,
    node_id VARCHAR(36) NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    message TEXT,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_nodes_network ON nodes(network);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
CREATE INDEX IF NOT EXISTS idx_nodes_chain_id ON nodes(chain_id);
CREATE INDEX IF NOT EXISTS idx_node_metrics_node_id ON node_metrics(node_id);
CREATE INDEX IF NOT EXISTS idx_node_metrics_timestamp ON node_metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_node_events_node_id ON node_events(node_id);
CREATE INDEX IF NOT EXISTS idx_node_events_timestamp ON node_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_node_events_event_type ON node_events(event_type);
CREATE INDEX IF NOT EXISTS idx_node_events_severity ON node_events(severity);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for automatic updated_at update
DROP TRIGGER IF EXISTS update_nodes_updated_at ON nodes;
CREATE TRIGGER update_nodes_updated_at
    BEFORE UPDATE ON nodes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
