package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"smart-cities/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// PostgresRepository handles all PostgreSQL database operations
type PostgresRepository struct {
	pool        *pgxpool.Pool
	tablePrefix string
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(config *DatabaseConfig) (*PostgresRepository, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
		config.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = int32(config.MaxOpenConns)
	poolConfig.MinConns = int32(config.MaxIdleConns)
	poolConfig.MaxConnLifetime = config.ConnMaxLifetime

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &PostgresRepository{
		pool:        pool,
		tablePrefix: config.TablePrefix,
	}

	// Initialize schema
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	r.pool.Close()
}

// tableName returns the full table name with prefix
func (r *PostgresRepository) tableName(name string) string {
	return r.tablePrefix + name
}

// DatabaseConfig matches the config.Database structure
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	TablePrefix     string        `yaml:"table_prefix"`
	SSLMode         string        `yaml:"sslmode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// initSchema creates the required database tables
func (r *PostgresRepository) initSchema() error {
	ctx := context.Background()

	schemas := []string{
		// Infrastructure nodes table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				node_type VARCHAR(50) NOT NULL,
				name VARCHAR(200) NOT NULL,
				district_id VARCHAR(100) NOT NULL,
				location_lat DECIMAL(12, 8) NOT NULL,
				location_lng DECIMAL(12, 8) NOT NULL,
				location_alt DECIMAL(10, 2),
				is_operational BOOLEAN DEFAULT true,
				last_ping_at TIMESTAMP NOT NULL,
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("infrastructure_nodes")),

		// Sensor readings table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				node_id VARCHAR(36) NOT NULL,
				district_id VARCHAR(100) NOT NULL,
				metric_type VARCHAR(50) NOT NULL,
				value DECIMAL(20, 6) NOT NULL,
				unit VARCHAR(20) NOT NULL,
				quality VARCHAR(20) DEFAULT 'good',
				timestamp TIMESTAMP NOT NULL,
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("sensor_readings")),

		// Citizen incidents table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				incident_type VARCHAR(50) NOT NULL,
				category VARCHAR(100) NOT NULL,
				district_id VARCHAR(100) NOT NULL,
				location_lat DECIMAL(12, 8) NOT NULL,
				location_lng DECIMAL(12, 8) NOT NULL,
				description TEXT NOT NULL,
				status VARCHAR(50) NOT NULL DEFAULT 'reported',
				priority INTEGER NOT NULL DEFAULT 4,
				reporter_hash VARCHAR(64) NOT NULL,
				image_urls JSONB,
				assigned_to VARCHAR(100),
				resolved_at TIMESTAMP,
				resolution TEXT,
				response_time INTERVAL,
				upvotes INTEGER DEFAULT 0,
				downvotes INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("citizen_incidents")),

		// Districts table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				name VARCHAR(200) NOT NULL,
				code VARCHAR(50) NOT NULL UNIQUE,
				population INTEGER NOT NULL,
				area_km2 DECIMAL(10, 4) NOT NULL,
				center_lat DECIMAL(12, 8) NOT NULL,
				center_lng DECIMAL(12, 8) NOT NULL,
				priority INTEGER DEFAULT 0,
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("districts")),

		// Resource metrics table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				district_id VARCHAR(100) NOT NULL,
				period_type VARCHAR(20) NOT NULL,
				period_start TIMESTAMP NOT NULL,
				period_end TIMESTAMP NOT NULL,
				energy_used DECIMAL(20, 2) NOT NULL,
				water_consumed DECIMAL(20, 2) NOT NULL,
				waste_generated DECIMAL(20, 2) NOT NULL,
				peak_demand DECIMAL(20, 2) NOT NULL,
				avg_aqi DECIMAL(10, 2) NOT NULL,
				cost DECIMAL(20, 2) NOT NULL,
				carbon_emission DECIMAL(20, 2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("resource_metrics")),

		// Service requests table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				request_type VARCHAR(50) NOT NULL,
				district_id VARCHAR(100) NOT NULL,
				location_lat DECIMAL(12, 8) NOT NULL,
				location_lng DECIMAL(12, 8) NOT NULL,
				description TEXT NOT NULL,
				status VARCHAR(50) NOT NULL DEFAULT 'created',
				priority INTEGER NOT NULL DEFAULT 3,
				scheduled_date TIMESTAMP,
				completed_date TIMESTAMP,
				assigned_team VARCHAR(100),
				estimated_cost DECIMAL(20, 2),
				actual_cost DECIMAL(20, 2),
				notes TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("service_requests")),

		// Traffic data table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				location_id VARCHAR(100) NOT NULL,
				district_id VARCHAR(100) NOT NULL,
				period_start TIMESTAMP NOT NULL,
				period_end TIMESTAMP NOT NULL,
				vehicle_count INTEGER NOT NULL,
				avg_speed DECIMAL(10, 2) NOT NULL,
				max_speed DECIMAL(10, 2) NOT NULL,
				congestion_level VARCHAR(20) NOT NULL,
				incidents INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("traffic_data")),

		// Environmental metrics table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				district_id VARCHAR(100) NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				aqi INTEGER NOT NULL,
				aqi_category VARCHAR(30) NOT NULL,
				pm25 DECIMAL(10, 2) NOT NULL,
				pm10 DECIMAL(10, 2) NOT NULL,
				co2 DECIMAL(10, 2) NOT NULL,
				no2 DECIMAL(10, 2) NOT NULL,
				o3 DECIMAL(10, 2) NOT NULL,
				so2 DECIMAL(10, 2) NOT NULL,
				temperature DECIMAL(8, 2) NOT NULL,
				humidity DECIMAL(5, 2) NOT NULL,
				noise_level DECIMAL(8, 2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("environmental_metrics")),

		// Alerts table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				alert_type VARCHAR(50) NOT NULL,
				severity VARCHAR(20) NOT NULL,
				title VARCHAR(200) NOT NULL,
				message TEXT NOT NULL,
				source VARCHAR(100) NOT NULL,
				entity_id VARCHAR(36),
				entity_type VARCHAR(100),
				is_read BOOLEAN DEFAULT false,
				expires_at TIMESTAMP,
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("alerts")),

		// Sustainability reports table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				district_id VARCHAR(100) NOT NULL,
				period_start TIMESTAMP NOT NULL,
				period_end TIMESTAMP NOT NULL,
				total_energy_used DECIMAL(20, 2) NOT NULL,
				energy_savings DECIMAL(20, 2) NOT NULL,
				energy_savings_pct DECIMAL(5, 2) NOT NULL,
				total_water_used DECIMAL(20, 2) NOT NULL,
				water_saved DECIMAL(20, 2) NOT NULL,
				water_saved_pct DECIMAL(5, 2) NOT NULL,
				total_waste DECIMAL(20, 2) NOT NULL,
				waste_recycled DECIMAL(20, 2) NOT NULL,
				recycling_rate DECIMAL(5, 2) NOT NULL,
				carbon_emissions DECIMAL(20, 2) NOT NULL,
				carbon_reduced DECIMAL(20, 2) NOT NULL,
				carbon_reduced_pct DECIMAL(5, 2) NOT NULL,
				green_spaces_added DECIMAL(10, 2) NOT NULL,
				generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("sustainability_reports")),

		// Create indexes
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_node_id ON %s(node_id)`,
			r.tableName("sensor_readings"), r.tableName("sensor_readings")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp)`,
			r.tableName("sensor_readings"), r.tableName("sensor_readings")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_district_id ON %s(district_id)`,
			r.tableName("sensor_readings"), r.tableName("sensor_readings")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_status ON %s(status)`,
			r.tableName("citizen_incidents"), r.tableName("citizen_incidents")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_district_id ON %s(district_id)`,
			r.tableName("citizen_incidents"), r.tableName("citizen_incidents")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_district_id ON %s(resource_metrics)`,
			r.tableName("resource_metrics"), r.tableName("resource_metrics")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_period ON %s(resource_metrics)`,
			r.tableName("resource_metrics"), r.tableName("resource_metrics")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_district_id ON %s(environmental_metrics)`,
			r.tableName("environmental_metrics"), r.tableName("environmental_metrics")),
	}

	for _, schema := range schemas {
		if _, err := r.pool.Exec(ctx, schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

// CreateInfrastructureNode creates a new infrastructure node
func (r *PostgresRepository) CreateInfrastructureNode(ctx context.Context, node *domain.InfrastructureNode) error {
	if node.ID == "" {
		node.ID = uuid.New().String()
	}

	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	metadata, err := json.Marshal(node.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, node_type, name, district_id, location_lat, location_lng,
			location_alt, is_operational, last_ping_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.tableName("infrastructure_nodes"))

	_, err = r.pool.Exec(ctx, query,
		node.ID,
		node.NodeType,
		node.Name,
		node.DistrictID,
		node.Location.Latitude,
		node.Location.Longitude,
		node.Location.Altitude,
		node.IsOperational,
		node.LastPingAt,
		metadata,
		node.CreatedAt,
		node.UpdatedAt,
	)

	return err
}

// GetInfrastructureNodes retrieves all infrastructure nodes
func (r *PostgresRepository) GetInfrastructureNodes(ctx context.Context, districtID string) ([]*domain.InfrastructureNode, error) {
	query := fmt.Sprintf(`
		SELECT id, node_type, name, district_id, location_lat, location_lng,
			location_alt, is_operational, last_ping_at, metadata, created_at, updated_at
		FROM %s
		WHERE district_id = $1 OR $1 = ''
		ORDER BY node_type, name
	`, r.tableName("infrastructure_nodes"))

	rows, err := r.pool.Query(ctx, query, districtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*domain.InfrastructureNode
	for rows.Next() {
		var node domain.InfrastructureNode
		var metadata []byte

		err := rows.Scan(
			&node.ID,
			&node.NodeType,
			&node.Name,
			&node.DistrictID,
			&node.Location.Latitude,
			&node.Location.Longitude,
			&node.Location.Altitude,
			&node.IsOperational,
			&node.LastPingAt,
			&metadata,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadata != nil {
			json.Unmarshal(metadata, &node.Metadata)
		}

		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

// CreateSensorReading creates a new sensor reading
func (r *PostgresRepository) CreateSensorReading(ctx context.Context, reading *domain.SensorReading) error {
	if reading.ID == "" {
		reading.ID = uuid.New().String()
	}

	reading.CreatedAt = time.Now()

	metadata, err := json.Marshal(reading.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, node_id, district_id, metric_type, value, unit,
			quality, timestamp, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.tableName("sensor_readings"))

	_, err = r.pool.Exec(ctx, query,
		reading.ID,
		reading.NodeID,
		reading.DistrictID,
		reading.MetricType,
		reading.Value,
		reading.Unit,
		reading.Quality,
		reading.Timestamp,
		metadata,
		reading.CreatedAt,
	)

	return err
}

// GetSensorReadings retrieves sensor readings for a district
func (r *PostgresRepository) GetSensorReadings(ctx context.Context, districtID string, startTime, endTime time.Time, limit int) ([]*domain.SensorReading, error) {
	query := fmt.Sprintf(`
		SELECT id, node_id, district_id, metric_type, value, unit,
			quality, timestamp, metadata, created_at
		FROM %s
		WHERE district_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
		LIMIT $4
	`, r.tableName("sensor_readings"))

	rows, err := r.pool.Query(ctx, query, districtID, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []*domain.SensorReading
	for rows.Next() {
		var reading domain.SensorReading
		var metadata []byte

		err := rows.Scan(
			&reading.ID,
			&reading.NodeID,
			&reading.DistrictID,
			&reading.MetricType,
			&reading.Value,
			&reading.Unit,
			&reading.Quality,
			&reading.Timestamp,
			&metadata,
			&reading.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadata != nil {
			json.Unmarshal(metadata, &reading.Metadata)
		}

		readings = append(readings, &reading)
	}

	return readings, rows.Err()
}

// CreateCitizenIncident creates a new citizen incident
func (r *PostgresRepository) CreateCitizenIncident(ctx context.Context, incident *domain.CitizenIncident) error {
	if incident.ID == "" {
		incident.ID = uuid.New().String()
	}

	incident.CreatedAt = time.Now()
	incident.UpdatedAt = time.Now()

	if incident.Status == "" {
		incident.Status = domain.IncidentStatusReported
	}

	imageURLs, err := json.Marshal(incident.ImageURLs)
	if err != nil {
		imageURLs = []byte("[]")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, incident_type, category, district_id, location_lat,
			location_lng, description, status, priority, reporter_hash, image_urls,
			assigned_to, resolved_at, resolution, response_time, upvotes, downvotes,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`, r.tableName("citizen_incidents"))

	_, err = r.pool.Exec(ctx, query,
		incident.ID,
		incident.IncidentType,
		incident.Category,
		incident.DistrictID,
		incident.Location.Latitude,
		incident.Location.Longitude,
		incident.Description,
		incident.Status,
		incident.Priority,
		incident.ReporterHash,
		imageURLs,
		incident.AssignedTo,
		incident.ResolvedAt,
		incident.Resolution,
		incident.ResponseTime,
		incident.Upvotes,
		incident.Downvotes,
		incident.CreatedAt,
		incident.UpdatedAt,
	)

	return err
}

// GetIncidentsByStatus retrieves incidents by status
func (r *PostgresRepository) GetIncidentsByStatus(ctx context.Context, status domain.IncidentStatus, limit int) ([]*domain.CitizenIncident, error) {
	query := fmt.Sprintf(`
		SELECT id, incident_type, category, district_id, location_lat,
			location_lng, description, status, priority, reporter_hash, image_urls,
			assigned_to, resolved_at, resolution, response_time, upvotes, downvotes,
			created_at, updated_at
		FROM %s
		WHERE status = $1
		ORDER BY priority ASC, created_at DESC
		LIMIT $2
	`, r.tableName("citizen_incidents"))

	rows, err := r.pool.Query(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*domain.CitizenIncident
	for rows.Next() {
		var incident domain.CitizenIncident
		var imageURLs []byte
		var resolvedAt sql.NullTime
		var resolution sql.NullString
		var assignedTo sql.NullString
		var responseTime sql.NullDuration

		err := rows.Scan(
			&incident.ID,
			&incident.IncidentType,
			&incident.Category,
			&incident.DistrictID,
			&incident.Location.Latitude,
			&incident.Location.Longitude,
			&incident.Description,
			&incident.Status,
			&incident.Priority,
			&incident.ReporterHash,
			&imageURLs,
			&assignedTo,
			&resolvedAt,
			&resolution,
			&responseTime,
			&incident.Upvotes,
			&incident.Downvotes,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			incident.ResolvedAt = &resolvedAt.Time
		}
		if resolution.Valid {
			incident.Resolution = resolution.String
		}
		if assignedTo.Valid {
			incident.AssignedTo = assignedTo.String
		}
		if responseTime.Valid {
			incident.ResponseTime = &responseTime.Duration
		}
		if imageURLs != nil {
			json.Unmarshal(imageURLs, &incident.ImageURLs)
		}

		incidents = append(incidents, &incident)
	}

	return incidents, rows.Err()
}

// CreateDistrict creates a new district
func (r *PostgresRepository) CreateDistrict(ctx context.Context, district *domain.District) error {
	if district.ID == "" {
		district.ID = uuid.New().String()
	}

	district.CreatedAt = time.Now()
	district.UpdatedAt = time.Now()

	metadata, err := json.Marshal(district.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, code, population, area_km2, center_lat,
			center_lng, priority, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, r.tableName("districts"))

	_, err = r.pool.Exec(ctx, query,
		district.ID,
		district.Name,
		district.Code,
		district.Population,
		district.AreaKm2,
		district.Coordinates.Latitude,
		district.Coordinates.Longitude,
		district.Priority,
		metadata,
		district.CreatedAt,
		district.UpdatedAt,
	)

	return err
}

// GetDistricts retrieves all districts
func (r *PostgresRepository) GetDistricts(ctx context.Context) ([]*domain.District, error) {
	query := fmt.Sprintf(`
		SELECT id, name, code, population, area_km2, center_lat,
			center_lng, priority, metadata, created_at, updated_at
		FROM %s ORDER BY priority
	`, r.tableName("districts"))

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var districts []*domain.District
	for rows.Next() {
		var district domain.District
		var metadata []byte

		err := rows.Scan(
			&district.ID,
			&district.Name,
			&district.Code,
			&district.Population,
			&district.AreaKm2,
			&district.Coordinates.Latitude,
			&district.Coordinates.Longitude,
			&district.Priority,
			&metadata,
			&district.CreatedAt,
			&district.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if metadata != nil {
			json.Unmarshal(metadata, &district.Metadata)
		}

		districts = append(districts, &district)
	}

	return districts, rows.Err()
}

// CreateResourceMetrics creates resource metrics
func (r *PostgresRepository) CreateResourceMetrics(ctx context.Context, metrics *domain.ResourceMetrics) error {
	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}

	metrics.CreatedAt = time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, district_id, period_type, period_start, period_end,
			energy_used, water_consumed, waste_generated, peak_demand, avg_aqi,
			cost, carbon_emission, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, r.tableName("resource_metrics"))

	_, err := r.pool.Exec(ctx, query,
		metrics.ID,
		metrics.DistrictID,
		metrics.PeriodType,
		metrics.PeriodStart,
		metrics.PeriodEnd,
		metrics.EnergyUsed,
		metrics.WaterConsumed,
		metrics.WasteGenerated,
		metrics.PeakDemand,
		metrics.AvgAQI,
		metrics.Cost,
		metrics.CarbonEmission,
		metrics.CreatedAt,
	)

	return err
}

// GetResourceMetrics retrieves resource metrics for a district
func (r *PostgresRepository) GetResourceMetrics(ctx context.Context, districtID string, startTime, endTime time.Time) ([]*domain.ResourceMetrics, error) {
	query := fmt.Sprintf(`
		SELECT id, district_id, period_type, period_start, period_end,
			energy_used, water_consumed, waste_generated, peak_demand, avg_aqi,
			cost, carbon_emission, created_at
		FROM %s
		WHERE district_id = $1 AND period_start BETWEEN $2 AND $3
		ORDER BY period_start DESC
	`, r.tableName("resource_metrics"))

	rows, err := r.pool.Query(ctx, query, districtID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.ResourceMetrics
	for rows.Next() {
		var m domain.ResourceMetrics

		err := rows.Scan(
			&m.ID,
			&m.DistrictID,
			&m.PeriodType,
			&m.PeriodStart,
			&m.PeriodEnd,
			&m.EnergyUsed,
			&m.WaterConsumed,
			&m.WasteGenerated,
			&m.PeakDemand,
			&m.AvgAQI,
			&m.Cost,
			&m.CarbonEmission,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, &m)
	}

	return metrics, rows.Err()
}

// CreateEnvironmentalMetrics creates environmental metrics
func (r *PostgresRepository) CreateEnvironmentalMetrics(ctx context.Context, metrics *domain.EnvironmentalMetrics) error {
	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}

	metrics.CreatedAt = time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, district_id, timestamp, aqi, aqi_category,
			pm25, pm10, co2, no2, o3, so2, temperature, humidity, noise_level, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, r.tableName("environmental_metrics"))

	_, err := r.pool.Exec(ctx, query,
		metrics.ID,
		metrics.DistrictID,
		metrics.Timestamp,
		metrics.AQI,
		metrics.AQICategory,
		metrics.PM25,
		metrics.PM10,
		metrics.CO2,
		metrics.NO2,
		metrics.O3,
		metrics.SO2,
		metrics.Temperature,
		metrics.Humidity,
		metrics.NoiseLevel,
		metrics.CreatedAt,
	)

	return err
}

// CreateAlert creates a new alert
func (r *PostgresRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	alert.CreatedAt = time.Now()

	metadata, err := json.Marshal(alert.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, alert_type, severity, title, message, source,
			entity_id, entity_type, is_read, expires_at, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.tableName("alerts"))

	_, err = r.pool.Exec(ctx, query,
		alert.ID,
		alert.AlertType,
		alert.Severity,
		alert.Title,
		alert.Message,
		alert.Source,
		alert.EntityID,
		alert.EntityType,
		alert.IsRead,
		alert.ExpiresAt,
		metadata,
		alert.CreatedAt,
	)

	return err
}

// GetActiveAlerts retrieves active alerts
func (r *PostgresRepository) GetActiveAlerts(ctx context.Context, minSeverity string) ([]*domain.Alert, error) {
	query := fmt.Sprintf(`
		SELECT id, alert_type, severity, title, message, source,
			entity_id, entity_type, is_read, expires_at, metadata, created_at
		FROM %s
		WHERE severity >= $1 AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY severity DESC, created_at DESC
	`, r.tableName("alerts"))

	rows, err := r.pool.Query(ctx, query, minSeverity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		var alert domain.Alert
		var metadata []byte
		var entityID, entityType sql.NullString
		var expiresAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Title,
			&alert.Message,
			&alert.Source,
			&entityID,
			&entityType,
			&alert.IsRead,
			&expiresAt,
			&metadata,
			&alert.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if entityID.Valid {
			alert.EntityID = entityID.String
		}
		if entityType.Valid {
			alert.EntityType = entityType.String
		}
		if expiresAt.Valid {
			alert.ExpiresAt = &expiresAt.Time
		}
		if metadata != nil {
			json.Unmarshal(metadata, &alert.Metadata)
		}

		alerts = append(alerts, &alert)
	}

	return alerts, rows.Err()
}
