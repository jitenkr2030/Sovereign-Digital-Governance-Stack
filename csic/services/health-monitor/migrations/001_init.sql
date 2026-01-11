-- Health Monitor Service Database Schema
-- PostgreSQL migrations for service status, alert rules, alerts, and outages

-- Create service status table
CREATE TABLE IF NOT EXISTS health_monitor_service_status (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'unknown',
    last_heartbeat TIMESTAMPTZ,
    last_check TIMESTAMPTZ,
    uptime FLOAT DEFAULT 0,
    response_time FLOAT DEFAULT 0,
    error_rate FLOAT DEFAULT 0,
    cpu_usage FLOAT DEFAULT 0,
    memory_usage FLOAT DEFAULT 0,
    disk_usage FLOAT DEFAULT 0,
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for service status
CREATE INDEX IF NOT EXISTS idx_health_monitor_service_status_status 
ON health_monitor_service_status(status);

CREATE INDEX IF NOT EXISTS idx_health_monitor_service_status_last_heartbeat 
ON health_monitor_service_status(last_heartbeat DESC);

-- Create alert rules table
CREATE TABLE IF NOT EXISTS health_monitor_alert_rules (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    condition VARCHAR(255) NOT NULL,
    threshold FLOAT NOT NULL,
    duration INT DEFAULT 0,
    severity VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    cooldown INT DEFAULT 300,
    notification_channels JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for alert rules
CREATE INDEX IF NOT EXISTS idx_health_monitor_alert_rules_service_name 
ON health_monitor_alert_rules(service_name);

CREATE INDEX IF NOT EXISTS idx_health_monitor_alert_rules_enabled 
ON health_monitor_alert_rules(enabled) WHERE enabled = true;

CREATE INDEX IF NOT EXISTS idx_health_monitor_alert_rules_severity 
ON health_monitor_alert_rules(severity);

-- Create alerts table
CREATE TABLE IF NOT EXISTS health_monitor_alerts (
    id VARCHAR(255) PRIMARY KEY,
    rule_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    condition VARCHAR(255) NOT NULL,
    current_value FLOAT NOT NULL,
    threshold FLOAT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'firing',
    message TEXT,
    fired_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for alerts
CREATE INDEX IF NOT EXISTS idx_health_monitor_alerts_service_name 
ON health_monitor_alerts(service_name);

CREATE INDEX IF NOT EXISTS idx_health_monitor_alerts_status 
ON health_monitor_alerts(status);

CREATE INDEX IF NOT EXISTS idx_health_monitor_alerts_fired_at 
ON health_monitor_alerts(fired_at DESC);

CREATE INDEX IF NOT EXISTS idx_health_monitor_alerts_rule_status 
ON health_monitor_alerts(rule_id, status);

-- Create outages table
CREATE TABLE IF NOT EXISTS health_monitor_outages (
    id VARCHAR(255) PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    instance_id VARCHAR(255),
    severity VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'detected',
    started_at TIMESTAMPTZ NOT NULL,
    detected_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    duration INT DEFAULT 0,
    impact TEXT,
    description TEXT,
    root_cause TEXT,
    affected_users INT DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for outages
CREATE INDEX IF NOT EXISTS idx_health_monitor_outages_service_name 
ON health_monitor_outages(service_name);

CREATE INDEX IF NOT EXISTS idx_health_monitor_outages_status 
ON health_monitor_outages(status) WHERE status != 'resolved';

CREATE INDEX IF NOT EXISTS idx_health_monitor_outages_started_at 
ON health_monitor_outages(started_at DESC);

-- Create heartbeat cache table (for Redis fallback)
CREATE TABLE IF NOT EXISTS health_monitor_heartbeats (
    service_name VARCHAR(255) PRIMARY KEY,
    heartbeat_data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_health_monitor_heartbeats_expires_at 
ON health_monitor_heartbeats(expires_at);

-- Insert default alert rules
INSERT INTO health_monitor_alert_rules (id, name, service_name, condition, threshold, duration, severity, enabled, cooldown) VALUES
('alert-cpu-high-001', 'High CPU Usage', '%', 'cpu > 80', 80, 60, 'warning', true, 300),
('alert-memory-high-001', 'High Memory Usage', '%', 'memory > 85', 85, 60, 'warning', true, 300),
('alert-disk-high-001', 'High Disk Usage', '%', 'disk > 90', 90, 300, 'critical', true, 600),
('alert-latency-high-001', 'High Latency', 'ms', 'latency > 500', 500, 120, 'warning', true, 300),
('alert-error-rate-high-001', 'High Error Rate', '%', 'error_rate > 5', 5, 60, 'critical', true, 600)
ON CONFLICT (id) DO NOTHING;

-- Insert default services
INSERT INTO health_monitor_service_status (id, name, status) VALUES
('service-api-gateway', 'api-gateway', 'unknown'),
('service-control-layer', 'control-layer', 'unknown'),
('service-audit-log', 'audit-log', 'unknown'),
('service-compliance', 'compliance', 'unknown'),
('service-health-monitor', 'health-monitor', 'healthy')
ON CONFLICT (id) DO NOTHING;
