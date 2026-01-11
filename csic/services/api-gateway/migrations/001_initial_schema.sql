-- API Gateway Database Migrations
-- This file contains the initial schema for routes, consumers, services, and analytics

-- Create routes table
CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    path VARCHAR(256) NOT NULL,
    methods JSONB NOT NULL,
    upstream_url VARCHAR(512) NOT NULL,
    upstream_path VARCHAR(256),
    timeout INTEGER DEFAULT 30,
    retry_count INTEGER DEFAULT 3,
    plugins JSONB DEFAULT '[]'::jsonb,
    rate_limit JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(path, (methods->>0))
);

CREATE INDEX IF NOT EXISTS idx_routes_path ON routes(path);
CREATE INDEX IF NOT EXISTS idx_routes_is_active ON routes(is_active);

-- Create consumers table
CREATE TABLE IF NOT EXISTS consumers (
    id UUID PRIMARY KEY,
    username VARCHAR(128) NOT NULL UNIQUE,
    groups JSONB DEFAULT '[]'::jsonb,
    quota JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consumers_username ON consumers(username);
CREATE INDEX IF NOT EXISTS idx_consumers_is_active ON consumers(is_active);

-- Create api_keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY,
    consumer_id UUID NOT NULL REFERENCES consumers(id) ON DELETE CASCADE,
    key VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    scopes JSONB DEFAULT '[]'::jsonb,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
CREATE INDEX IF NOT EXISTS idx_api_keys_consumer_id ON api_keys(consumer_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);

-- Create services table
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    host VARCHAR(256) NOT NULL,
    port INTEGER NOT NULL,
    protocol VARCHAR(16) NOT NULL DEFAULT 'http',
    health_check JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
CREATE INDEX IF NOT EXISTS idx_services_is_active ON services(is_active);

-- Create rate_limits table
CREATE TABLE IF NOT EXISTS rate_limits (
    id UUID PRIMARY KEY,
    identifier VARCHAR(256) NOT NULL,
    route_id UUID REFERENCES routes(id) ON DELETE CASCADE,
    requests INTEGER NOT NULL DEFAULT 0,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rate_limits_identifier ON rate_limits(identifier);
CREATE INDEX IF NOT EXISTS idx_rate_limits_route_id ON rate_limits(route_id);
CREATE INDEX IF NOT EXISTS idx_rate_limits_expires_at ON rate_limits(expires_at);

-- Create transform_rules table
CREATE TABLE IF NOT EXISTS transform_rules (
    id UUID PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    type VARCHAR(32) NOT NULL,
    action VARCHAR(32) NOT NULL,
    target VARCHAR(256) NOT NULL,
    value TEXT,
    pattern VARCHAR(256),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transform_rules_route_id ON transform_rules(route_id);

-- Create analytics_events table
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    event_type VARCHAR(32) NOT NULL,
    route_id UUID REFERENCES routes(id) ON DELETE SET NULL,
    consumer_id UUID REFERENCES consumers(id) ON DELETE SET NULL,
    request_id VARCHAR(64) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent VARCHAR(512),
    method VARCHAR(8) NOT NULL,
    path VARCHAR(512) NOT NULL,
    status_code INTEGER NOT NULL,
    latency_ms BIGINT NOT NULL,
    body_bytes BIGINT DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_analytics_timestamp ON analytics_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_route_id ON analytics_events(route_id);
CREATE INDEX IF NOT EXISTS idx_analytics_consumer_id ON analytics_events(consumer_id);
CREATE INDEX IF NOT EXISTS idx_analytics_request_id ON analytics_events(request_id);

-- Create certificates table
CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY,
    domain VARCHAR(256) NOT NULL UNIQUE,
    cert_data TEXT NOT NULL,
    key_data TEXT NOT NULL,
    issuer VARCHAR(256),
    not_before TIMESTAMP WITH TIME ZONE NOT NULL,
    not_after TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_certificates_domain ON certificates(domain);
CREATE INDEX IF NOT EXISTS idx_certificates_not_after ON certificates(not_after);

-- Create health_check_results table
CREATE TABLE IF NOT EXISTS health_check_results (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    healthy BOOLEAN NOT NULL,
    latency_ms BIGINT NOT NULL,
    status_code INTEGER,
    error_message TEXT,
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_check_service_id ON health_check_results(service_id);
CREATE INDEX IF NOT EXISTS idx_health_check_checked_at ON health_check_results(checked_at DESC);

-- Create triggers for automatic updated_at updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_routes_updated_at
    BEFORE UPDATE ON routes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_consumers_updated_at
    BEFORE UPDATE ON consumers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_services_updated_at
    BEFORE UPDATE ON services
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
