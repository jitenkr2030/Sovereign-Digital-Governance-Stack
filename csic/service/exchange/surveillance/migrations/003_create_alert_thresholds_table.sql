-- +goose Up
-- +goose StatementBegin

-- Create alert thresholds configuration table
CREATE TABLE IF NOT EXISTS alert_thresholds (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    severity VARCHAR(20) NOT NULL DEFAULT 'WARNING',
    parameters JSONB NOT NULL DEFAULT '{}',
    min_confidence DECIMAL(5,4) NOT NULL DEFAULT 0.5,
    cooldown_period INTERVAL NOT NULL DEFAULT '5 minutes',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID,
    description TEXT
);

-- Create unique index on alert_type
CREATE UNIQUE INDEX IF NOT EXISTS idx_alert_thresholds_type 
ON alert_thresholds(alert_type);

-- Create indexes for querying
CREATE INDEX IF NOT EXISTS idx_alert_thresholds_enabled 
ON alert_thresholds(enabled);

-- Insert default threshold configurations
INSERT INTO alert_thresholds (id, name, alert_type, enabled, severity, parameters, min_confidence, cooldown_period, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Wash Trade Detection', 'WASH_TRADING', true, 'CRITICAL', 
     '{"time_window": 60, "price_deviation_threshold": 0.01, "quantity_similarity_threshold": 0.9, "min_trades_for_detection": 3}', 
     0.70, '5 minutes', 
     'Detects wash trading patterns where the same accounts buy and sell to themselves or coordinate trades'),
    ('22222222-2222-2222-2222-222222222222', 'Spoofing Detection', 'SPOOFING', true, 'CRITICAL',
     '{"order_lifetime_window": 300, "large_order_threshold": 0.1, "cancellation_rate_threshold": 0.8, "bait_order_threshold": 0.05}',
     0.75, '10 minutes',
     'Detects spoofing patterns where large orders are placed to manipulate price then cancelled'),
    ('33333333-3333-3333-3333-333333333333', 'Price Manipulation Detection', 'PRICE_MANIPULATION', true, 'CRITICAL',
     '{"global_average_window": 3600, "deviation_threshold": 0.05, "min_data_points": 100}',
     0.80, '15 minutes',
     'Detects price deviations from global average that may indicate manipulation'),
    ('44444444-4444-4444-4444-444444444444', 'Volume Anomaly Detection', 'VOLUME_ANOMALY', true, 'WARNING',
     '{"anomaly_threshold": 3.0, "rolling_window": 86400}',
     0.60, '5 minutes',
     'Detects unusual volume spikes or drops that may indicate market manipulation'),
    ('55555555-5555-5555-5555-555555555555', 'Layering Detection', 'LAYERING', true, 'CRITICAL',
     '{"layer_count_threshold": 5, "layer_spacing": 0.02, "layer_lifetime": 60}',
     0.70, '10 minutes',
     'Detects layering patterns where multiple orders are placed at different price levels'),
    ('66666666-6666-6666-6666-666666666666', 'Front Running Detection', 'FRONT_RUNNING', true, 'CRITICAL',
     '{"time_threshold": 1000, "price_deviation": 0.001}',
     0.75, '15 minutes',
     'Detects front running where trades are executed ahead of large orders'),
    ('77777777-7777-7777-7777-777777777777', 'Pump and Dump Detection', 'PUMP_AND_DUMP', true, 'CRITICAL',
     '{"price_increase_threshold": 0.5, "volume_spike_threshold": 5.0, "time_window": 3600}',
     0.80, '30 minutes',
     'Detects pump and dump schemes where price is artificially inflated then sold')
ON CONFLICT (alert_type) DO NOTHING;

-- Create audit log for threshold changes
CREATE TABLE IF NOT EXISTS alert_thresholds_audit (
    id UUID PRIMARY KEY,
    threshold_id UUID NOT NULL,
    changed_by UUID NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    old_values JSONB NOT NULL,
    new_values JSONB NOT NULL,
    change_reason TEXT
);

-- Create index on threshold_id for audit trail
CREATE INDEX IF NOT EXISTS idx_alert_thresholds_audit_threshold 
ON alert_thresholds_audit(threshold_id);

-- Create index on changed_at for temporal queries
CREATE INDEX IF NOT EXISTS idx_alert_thresholds_audit_changed 
ON alert_thresholds_audit(changed_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop audit table
DROP TABLE IF EXISTS alert_thresholds_audit;

-- Drop thresholds table
DROP TABLE IF EXISTS alert_thresholds;

-- Drop indexes
DROP INDEX IF EXISTS idx_alert_thresholds_enabled;
DROP INDEX IF EXISTS idx_alert_thresholds_type;

-- +goose StatementEnd
