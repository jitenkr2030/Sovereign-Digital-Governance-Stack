-- Migration V2: Create TimescaleDB Hypertables for Time-Series Data
-- This migration converts relevant tables to TimescaleDB hypertables
-- for optimized time-series data storage and querying.
-- Direction: UP

-- Check if TimescaleDB extension is installed
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        RAISE EXCEPTION 'TimescaleDB extension is not installed. Please install TimescaleDB before running this migration.';
    END IF;
END;
$$;

-- Convert energy_consumption_logs to hypertable
-- This enables efficient storage and querying of time-series energy data
SELECT create_hypertable(
    'energy_consumption_logs',
    'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    migrate_data => TRUE
);

-- Create additional indexes specifically optimized for hypertables
-- These indexes are automatically partitioned with the hypertable chunks
CREATE INDEX IF NOT EXISTS idx_energy_logs_pool_timestamp_power 
    ON energy_consumption_logs(pool_id, timestamp DESC, power_consumption_watts DESC);

CREATE INDEX IF NOT EXISTS idx_energy_logs_machine_timestamp_power 
    ON energy_consumption_logs(machine_id, timestamp DESC, power_consumption_watts DESC);

-- Create continuous aggregate for hourly energy consumption statistics
-- This materializes aggregated data for faster queries
CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_energy_stats
WITH (timescaledb.continuous) AS
SELECT 
    pool_id,
    time_bucket(INTERVAL '1 hour', timestamp) AS hour,
    COUNT(*) AS sample_count,
    SUM(energy_used_wh) AS total_energy_wh,
    AVG(power_consumption_watts) AS avg_power_watts,
    MAX(power_consumption_watts) AS max_power_watts,
    MIN(power_consumption_watts) AS min_power_watts,
    AVG(voltage) AS avg_voltage,
    AVG(current) AS avg_current,
    AVG(power_factor) AS avg_power_factor,
    AVG(temperature) AS avg_temperature
FROM energy_consumption_logs
GROUP BY pool_id, hour;

-- Create indexes for the continuous aggregate
CREATE INDEX IF NOT EXISTS idx_hourly_energy_stats_pool_hour 
    ON hourly_energy_stats(pool_id, hour DESC);

-- Create continuous aggregate for daily energy consumption statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_energy_stats
WITH (timescaledb.continuous) AS
SELECT 
    pool_id,
    time_bucket(INTERVAL '1 day', timestamp) AS day,
    COUNT(*) AS sample_count,
    SUM(energy_used_wh) AS total_energy_wh,
    AVG(power_consumption_watts) AS avg_power_watts,
    MAX(power_consumption_watts) AS peak_power_watts,
    MIN(power_consumption_watts) AS min_power_watts,
    AVG(power_factor) AS avg_power_factor,
    MAX(temperature) AS max_temperature
FROM energy_consumption_logs
GROUP BY pool_id, day;

-- Create index for daily aggregate
CREATE INDEX IF NOT EXISTS idx_daily_energy_stats_pool_day 
    ON daily_energy_stats(pool_id, day DESC);

-- Create continuous aggregate for machine-level hourly statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_machine_stats
WITH (timescaledb.continuous) AS
SELECT 
    machine_id,
    pool_id,
    time_bucket(INTERVAL '1 hour', timestamp) AS hour,
    COUNT(*) AS sample_count,
    SUM(energy_used_wh) AS total_energy_wh,
    AVG(power_consumption_watts) AS avg_power_watts,
    MAX(power_consumption_watts) AS max_power_watts,
    MIN(power_consumption_watts) AS min_power_watts,
    AVG(temperature) AS avg_temperature,
    MAX(temperature) AS max_temperature
FROM energy_consumption_logs
WHERE machine_id IS NOT NULL
GROUP BY machine_id, pool_id, hour;

-- Create index for machine-level aggregate
CREATE INDEX IF NOT EXISTS idx_hourly_machine_stats_machine_hour 
    ON hourly_machine_stats(machine_id, hour DESC);

-- Create function to refresh continuous aggregates
-- This function can be called by a scheduled job to keep aggregates up to date
CREATE OR REPLACE FUNCTION refresh_continuous_aggregates()
RETURNS VOID AS $$
BEGIN
    -- Refresh hourly aggregates
    CALL refresh_continuous_aggregate('hourly_energy_stats');
    CALL refresh_continuous_aggregate('hourly_machine_stats');
    
    -- Refresh daily aggregates
    CALL refresh_continuous_aggregate('daily_energy_stats');
END;
$$ LANGUAGE plpgsql;

-- Create view for real-time energy consumption monitoring
-- Combines raw data with aggregated statistics
CREATE OR REPLACE VIEW v_current_energy_status AS
SELECT 
    mp.id AS pool_id,
    mp.name AS pool_name,
    mp.total_power_consumption AS current_power_watts,
    mp.energy_limit_watts AS limit_watts,
    mp.status AS pool_status,
    COALESCE(hourly.avg_power_watts, 0) AS hourly_avg_power_watts,
    COALESCE(hourly.max_power_watts, 0) AS hourly_max_power_watts,
    COALESCE(hourly.total_energy_wh, 0) AS hourly_energy_wh,
    COALESCE(daily.total_energy_wh, 0) AS daily_energy_wh,
    CASE 
        WHEN mp.energy_limit_watts > 0 
        THEN ROUND((mp.total_power_consumption / mp.energy_limit_watts * 100)::numeric, 2)
        ELSE 0 
    END AS utilization_percentage,
    mp.geographic_zone,
    mp.last_inspection_date
FROM mining_pools mp
LEFT JOIN LATERAL (
    SELECT * FROM hourly_energy_stats
    WHERE pool_id = mp.id
    ORDER BY hour DESC
    LIMIT 1
) hourly ON TRUE
LEFT JOIN LATERAL (
    SELECT * FROM daily_energy_stats
    WHERE pool_id = mp.id
    ORDER BY day DESC
    LIMIT 1
) daily ON TRUE;

-- Create function to calculate pool efficiency score
CREATE OR REPLACE FUNCTION calculate_pool_efficiency(pool_uuid UUID)
RETURNS DECIMAL(5, 2) AS $$
DECLARE
    total_hashrate DECIMAL(20, 2);
    total_energy_wh DECIMAL(15, 4);
    efficiency_score DECIMAL(5, 2);
BEGIN
    -- Get total hashrate for the pool
    SELECT COALESCE(SUM(hashrate), 0)
    INTO total_hashrate
    FROM mining_machines
    WHERE pool_id = pool_uuid AND status IN ('online', 'mining');

    -- Get total energy consumption for last 24 hours
    SELECT COALESCE(SUM(energy_used_wh), 0)
    INTO total_energy_wh
    FROM energy_consumption_logs
    WHERE pool_id = pool_uuid
    AND timestamp >= NOW() - INTERVAL '24 hours';

    -- Calculate efficiency: TH/s per Wh consumed
    -- Higher is better (more hash power per unit of energy)
    IF total_energy_wh > 0 THEN
        efficiency_score := ROUND((total_hashrate / (total_energy_wh / 1000))::numeric, 2);
    ELSE
        efficiency_score := 0;
    END IF;

    RETURN efficiency_score;
END;
$$ LANGUAGE plpgsql;

-- Create function to detect energy anomalies
CREATE OR REPLACE FUNCTION detect_energy_anomalies(
    p_pool_id UUID,
    p_std_deviations DECIMAL DEFAULT 3.0
)
RETURNS TABLE (
    timestamp TIMESTAMP WITH TIME ZONE,
    actual_value DECIMAL(15, 2),
    expected_value DECIMAL(15, 2),
    deviation DECIMAL(15, 2),
    is_anomaly BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    WITH stats AS (
        SELECT 
            timestamp,
            power_consumption_watts,
            AVG(power_consumption_watts) OVER (
                PARTITION BY pool_id 
                ORDER BY timestamp 
                ROWS BETWEEN 100 PRECEDING AND 100 FOLLOWING
            ) AS moving_avg,
            STDDEV(power_consumption_watts) OVER (
                PARTITION BY pool_id 
                ORDER BY timestamp 
                ROWS BETWEEN 100 PRECEDING AND 100 FOLLOWING
            ) AS moving_stddev
        FROM energy_consumption_logs
        WHERE pool_id = p_pool_id
    )
    SELECT 
        e.timestamp,
        e.power_consumption_watts AS actual_value,
        s.moving_avg AS expected_value,
        ABS(e.power_consumption_watts - s.moving_avg) AS deviation,
        CASE 
            WHEN s.moving_stddev > 0 
            THEN ABS(e.power_consumption_watts - s.moving_avg) > (p_std_deviations * s.moving_stddev)
            ELSE FALSE
        END AS is_anomaly
    FROM energy_consumption_logs e
    JOIN stats s ON e.timestamp = s.timestamp
    WHERE e.pool_id = p_pool_id
    AND e.timestamp >= NOW() - INTERVAL '24 hours'
    ORDER BY e.timestamp DESC;
END;
$$ LANGUAGE plpgsql;

-- Set up default compression policy
-- Compress data older than 7 days to save storage
SELECT add_compression_policy('energy_consumption_logs', INTERVAL '7 days');

-- Set up default retention policy
-- Delete data older than 2 years (730 days)
SELECT add_retention_policy('energy_consumption_logs', INTERVAL '730 days');

-- Create index for efficient querying of recent data
CREATE INDEX IF NOT EXISTS idx_energy_logs_pool_timestamp_recent 
    ON energy_consumption_logs(pool_id, timestamp DESC)
    WHERE timestamp > NOW() - INTERVAL '30 days';

-- Direction: DOWN
-- Drop continuous aggregates, policies, and revert hypertable
-- Note: In production, you would want to be more careful with data destruction

-- DROP MATERIALIZED VIEW IF EXISTS hourly_machine_stats CASCADE;
-- DROP MATERIALIZED VIEW IF EXISTS daily_energy_stats CASCADE;
-- DROP MATERIALIZED VIEW IF EXISTS hourly_energy_stats CASCADE;

-- DROP FUNCTION IF EXISTS refresh_continuous_aggregates() CASCADE;
-- DROP FUNCTION IF EXISTS calculate_pool_efficiency(UUID) CASCADE;
-- DROP FUNCTION IF EXISTS detect_energy_anomalies(UUID, DECIMAL) CASCADE;

-- DROP VIEW IF EXISTS v_current_energy_status CASCADE;

-- To remove hypertable and revert to regular table:
-- SELECT drop_hypertable('energy_consumption_logs', 'CASCADE');
