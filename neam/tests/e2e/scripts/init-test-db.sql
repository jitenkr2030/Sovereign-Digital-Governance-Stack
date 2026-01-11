-- Initialize test database with sample data for E2E testing
-- This script runs on container startup

-- Create test users with different roles
INSERT INTO users (id, email, password_hash, role, first_name, last_name, created_at, updated_at) VALUES
('test-admin-001', 'admin@test.gov', '$2b$10$dummy_hash_admin', 'admin', 'Test', 'Administrator', NOW(), NOW()),
('test-analyst-001', 'analyst@test.gov', '$2b$10$dummy_hash_analyst', 'analyst', 'Test', 'Analyst', NOW(), NOW()),
('test-supervisor-001', 'supervisor@test.gov', '$2b$10$dummy_hash_supervisor', 'supervisor', 'Test', 'Supervisor', NOW(), NOW());

-- Create test intelligence sources
INSERT INTO intelligence_sources (id, name, type, endpoint, auth_type, is_active, created_at) VALUES
('source-001', 'Test Police Database', 'database', 'postgresql://localhost/police', 'api_key', true, NOW()),
('source-002', 'Test Emergency Services API', 'api', 'https://api.test-emergency.gov/v1', 'oauth2', true, NOW()),
('source-003', 'Test Social Media Stream', 'stream', 'wss://stream.test-social.media', 'none', true, NOW());

-- Create test threat categories
INSERT INTO threat_categories (id, name, severity_level, description, color_code) VALUES
('threat-low-001', 'Minor Incident', 1, 'Low priority incidents requiring monitoring', '#22c55e'),
('threat-medium-001', 'Moderate Threat', 2, 'Threats requiring attention and coordination', '#f59e0b'),
('threat-high-001', 'High Priority Threat', 3, 'Critical threats requiring immediate response', '#ef4444'),
('threat-critical-001', 'Critical Emergency', 4, 'Emergency situations requiring immediate intervention', '#7f1d1d');

-- Create sample intelligence items for testing
INSERT INTO intelligence_items (id, title, description, source_id, category_id, severity, status, created_by, created_at, updated_at) VALUES
('intel-001', 'Suspicious Activity Report #1', 'Reports of unusual gathering in downtown area', 'source-001', 'threat-medium-001', 2, 'pending', 'test-analyst-001', NOW(), NOW()),
('intel-002', 'Emergency Alert - District 7', 'Multiple emergency calls from commercial district', 'source-002', 'threat-high-001', 3, 'under_review', 'test-analyst-001', NOW(), NOW()),
('intel-003', 'Social Media Trend Analysis', 'Increasing negative sentiment in local forums', 'source-003', 'threat-low-001', 1, 'new', 'test-analyst-001', NOW(), NOW());

-- Create sample interventions
INSERT INTO interventions (id, intelligence_item_id, title, description, type, status, priority, assigned_to, created_by, created_at, updated_at) VALUES
('intervention-001', 'intel-001', 'Patrol Deployment - Downtown', 'Deploy additional patrol units to downtown area', 'deployment', 'pending', 2, NULL, 'test-analyst-001', NOW(), NOW()),
('intervention-002', 'intel-002', 'Emergency Response Coordination', 'Coordinate emergency response with fire and medical services', 'coordination', 'in_progress', 3, 'test-admin-001', NOW(), NOW());

-- Create sample crisis alerts
INSERT INTO crisis_alerts (id, intelligence_item_id, title, description, severity, status, acknowledged_by, acknowledged_at, created_at) VALUES
('alert-001', 'intel-002', 'CRITICAL: District 7 Emergency', 'Multiple emergency reports requiring immediate attention', 4, 'acknowledged', 'test-supervisor-001', NOW(), NOW());

-- Create test reports
INSERT INTO reports (id, title, type, date_range_start, date_range_end, status, generated_by, created_at, updated_at) VALUES
('report-001', 'Weekly Intelligence Summary', 'weekly', NOW() - INTERVAL '7 days', NOW(), 'completed', 'test-analyst-001', NOW(), NOW()),
('report-002', 'Monthly Threat Assessment', 'monthly', NOW() - INTERVAL '30 days', NOW(), 'generating', 'test-admin-001', NOW(), NOW());

-- Index for performance optimization
CREATE INDEX IF NOT EXISTS idx_intelligence_status ON intelligence_items(status);
CREATE INDEX IF NOT EXISTS idx_intelligence_severity ON intelligence_items(severity);
CREATE INDEX IF NOT EXISTS idx_intelligence_source ON intelligence_items(source_id);
CREATE INDEX IF NOT EXISTS idx_interventions_intel ON interventions(intelligence_item_id);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON crisis_alerts(severity);
