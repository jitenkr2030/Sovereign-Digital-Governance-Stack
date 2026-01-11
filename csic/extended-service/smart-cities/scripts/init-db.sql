-- Smart Cities Database Initialization Script
-- PostgreSQL database initialization for smart city infrastructure

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Traffic Management Tables
CREATE TABLE IF NOT EXISTS traffic_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    capacity INTEGER DEFAULT 100,
    current_vehicle_count INTEGER DEFAULT 0,
    green_time INTEGER DEFAULT 30,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS traffic_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID REFERENCES traffic_nodes(id),
    incident_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) DEFAULT 'MEDIUM',
    description TEXT,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    reported_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS traffic_routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    node_ids UUID[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Emergency Response Tables
CREATE TABLE IF NOT EXISTS emergency_units (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    unit_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'AVAILABLE',
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    current_incident_id UUID,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emergency_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(100) NOT NULL,
    priority INTEGER DEFAULT 2,
    severity VARCHAR(50) DEFAULT 'MEDIUM',
    description TEXT,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    location_type VARCHAR(100) DEFAULT 'STREET',
    injuries_reported INTEGER DEFAULT 0,
    traffic_level DECIMAL(5, 4) DEFAULT 0.1,
    status VARCHAR(50) DEFAULT 'PENDING',
    assigned_unit_id UUID REFERENCES emergency_units(id),
    estimated_arrival TIMESTAMP WITH TIME ZONE,
    dispatched_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    reported_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Waste Management Tables
CREATE TABLE IF NOT EXISTS waste_bins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location VARCHAR(255) NOT NULL,
    zone VARCHAR(100) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    capacity INTEGER DEFAULT 1000,
    current_fill INTEGER DEFAULT 0,
    waste_type VARCHAR(50) DEFAULT 'GENERAL',
    last_collection TIMESTAMP WITH TIME ZONE,
    sensor_battery INTEGER DEFAULT 100,
    alert_flags TEXT[],
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS waste_collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bin_id UUID REFERENCES waste_bins(id),
    vehicle_id VARCHAR(255) NOT NULL,
    weight INTEGER NOT NULL,
    waste_type VARCHAR(50) DEFAULT 'GENERAL',
    collected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS collection_vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'AVAILABLE',
    capacity INTEGER DEFAULT 5000,
    current_load INTEGER DEFAULT 0,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Smart Lighting Tables
CREATE TABLE IF NOT EXISTS lighting_fixtures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    zone VARCHAR(100) NOT NULL,
    location VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    fixture_type VARCHAR(100) DEFAULT 'STREET_LIGHT',
    wattage INTEGER DEFAULT 100,
    current_brightness INTEGER DEFAULT 70,
    has_motion_sensor BOOLEAN DEFAULT FALSE,
    last_motion_detected TIMESTAMP WITH TIME ZONE,
    environment VARCHAR(100) DEFAULT 'URBAN',
    operating_hours INTEGER DEFAULT 0,
    last_maintenance TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'OPERATIONAL',
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS lighting_zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    fixture_ids UUID[],
    default_mode VARCHAR(50) DEFAULT 'NIGHT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_traffic_nodes_location ON traffic_nodes(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_traffic_incidents_status ON traffic_incidents(status);
CREATE INDEX IF NOT EXISTS idx_emergency_units_status ON emergency_units(status);
CREATE INDEX IF NOT EXISTS idx_emergency_incidents_priority ON emergency_incidents(priority DESC);
CREATE INDEX IF NOT EXISTS idx_waste_bins_zone ON waste_bins(zone);
CREATE INDEX IF NOT EXISTS idx_waste_bins_fill_level ON waste_bins(current_fill DESC);
CREATE INDEX IF NOT EXISTS idx_lighting_fixtures_zone ON lighting_fixtures(zone);
CREATE INDEX IF NOT EXISTS idx_lighting_fixtures_status ON lighting_fixtures(status);

-- Insert sample data
INSERT INTO traffic_nodes (name, location, latitude, longitude, capacity, current_vehicle_count, green_time) VALUES
('Main St & 1st Ave', 'Downtown Intersection', 40.7128, -74.0060, 150, 85, 45),
('Broadway & 5th St', 'Midtown Crossing', 40.7580, -73.9855, 200, 120, 60),
('Park Ave & 34th St', 'East Side Junction', 40.7484, -73.9857, 180, 95, 40)
ON CONFLICT DO NOTHING;

INSERT INTO emergency_units (name, unit_type, latitude, longitude) VALUES
('Engine 7', 'FIRE_TRUCK', 40.7128, -74.0060),
('Ambulance 12', 'AMBULANCE', 40.7580, -73.9855),
('Patrol Unit 45', 'POLICE', 40.7484, -73.9857),
('HazMat Unit 1', 'HAZMAT_UNIT', 40.7300, -74.0000)
ON CONFLICT DO NOTHING;

INSERT INTO waste_bins (location, zone, latitude, longitude, capacity, current_fill, waste_type) VALUES
('Central Park North', 'Zone A', 40.7969, -73.9530, 1000, 650, 'GENERAL'),
('5th Avenue Mall', 'Zone B', 40.7580, -73.9780, 800, 520, 'RECYCLABLE'),
('Industrial District', 'Zone C', 40.7200, -74.0100, 1500, 980, 'GENERAL')
ON CONFLICT DO NOTHING;

INSERT INTO lighting_fixtures (name, zone, location, latitude, longitude, fixture_type, wattage, current_brightness, has_motion_sensor) VALUES
('Main St Light 1', 'Zone A', 'Main Street', 40.7128, -74.0060, 'STREET_LIGHT', 100, 70, TRUE),
('Park Avenue Light 12', 'Zone B', 'Park Avenue', 40.7484, -73.9857, 'STREET_LIGHT', 150, 80, FALSE),
('Central Park Path Light', 'Zone A', 'Central Park Entrance', 40.7829, -73.9654, 'PATHWAY_LIGHT', 60, 60, TRUE)
ON CONFLICT DO NOTHING;

-- Verify installation
SELECT 'Smart Cities Database Initialized Successfully' AS status;
