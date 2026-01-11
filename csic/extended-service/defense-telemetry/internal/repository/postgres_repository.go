package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"defense-telemetry/internal/domain"

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
		// Defense units table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				unit_code VARCHAR(100) NOT NULL UNIQUE,
				unit_name VARCHAR(200) NOT NULL,
				unit_type VARCHAR(50) NOT NULL,
				status VARCHAR(50) NOT NULL DEFAULT 'active',
				command_id VARCHAR(36),
				base_location_lat DECIMAL(12, 8),
				base_location_lng DECIMAL(12, 8),
				base_location_alt DECIMAL(10, 2),
				last_contact_at TIMESTAMP NOT NULL,
				is_deployed BOOLEAN DEFAULT false,
				deployment_zone VARCHAR(100),
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("units")),

		// Telemetry logs table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				unit_id VARCHAR(36) NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				location_lat DECIMAL(12, 8) NOT NULL,
				location_lng DECIMAL(12, 8) NOT NULL,
				location_alt DECIMAL(10, 2),
				location_heading DECIMAL(8, 4),
				velocity DECIMAL(10, 4) NOT NULL,
				fuel_level DECIMAL(5, 2) NOT NULL,
				fuel_consumption DECIMAL(10, 4),
				ammunition_count INTEGER NOT NULL,
				hull_integrity DECIMAL(5, 2) NOT NULL,
				engine_temp DECIMAL(8, 4),
				internal_temp DECIMAL(8, 4),
				humidity DECIMAL(5, 2),
				power_output DECIMAL(10, 4),
				power_consumption DECIMAL(10, 4),
				battery_level DECIMAL(5, 2),
				signal_strength DECIMAL(10, 4),
				encryption_level INTEGER DEFAULT 1,
				component_health JSONB,
				payload JSONB,
				processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("telemetry_logs")),

		// Threats table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				unit_id VARCHAR(36) NOT NULL,
				threat_type VARCHAR(50) NOT NULL,
				threat_source VARCHAR(200),
				severity INTEGER NOT NULL CHECK (severity >= 1 AND severity <= 5),
				status VARCHAR(50) NOT NULL DEFAULT 'detected',
				source_lat DECIMAL(12, 8),
				source_lng DECIMAL(12, 8),
				source_alt DECIMAL(10, 2),
				target_lat DECIMAL(12, 8),
				target_lng DECIMAL(12, 8),
				target_alt DECIMAL(10, 2),
				velocity DECIMAL(10, 4),
				estimated_impact DECIMAL(10, 4),
				detection_time TIMESTAMP NOT NULL,
				first_seen_at TIMESTAMP NOT NULL,
				last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				resolved_at TIMESTAMP,
				resolution TEXT,
				countermeasures JSONB,
				escalation_level INTEGER DEFAULT 0,
				assigned_to VARCHAR(100),
				notes TEXT,
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("threats")),

		// Maintenance logs table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				unit_id VARCHAR(36) NOT NULL,
				component VARCHAR(100) NOT NULL,
				maintenance_type VARCHAR(50) NOT NULL,
				condition_score DECIMAL(5, 2) NOT NULL,
				previous_condition DECIMAL(5, 2),
				work_performed TEXT NOT NULL,
				parts_replaced JSONB,
				technician_id VARCHAR(100),
				estimated_life INTERVAL,
				scheduled_at TIMESTAMP,
				started_at TIMESTAMP,
				completed_at TIMESTAMP,
				next_service_due TIMESTAMP,
				predicted_failure_date TIMESTAMP,
				status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
				cost DECIMAL(12, 2),
				notes TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("maintenance_logs")),

		// Maintenance predictions table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				unit_id VARCHAR(36) NOT NULL,
				component VARCHAR(100) NOT NULL,
				current_health DECIMAL(5, 2) NOT NULL,
				predicted_failure_date TIMESTAMP NOT NULL,
				confidence_level DECIMAL(5, 2) NOT NULL,
				risk_level VARCHAR(50) NOT NULL,
				recommendations JSONB,
				model_version VARCHAR(50),
				prediction_date TIMESTAMP NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("maintenance_predictions")),

		// Deployments table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				unit_id VARCHAR(36) NOT NULL,
				mission_id VARCHAR(36),
				zone_id VARCHAR(100) NOT NULL,
				start_time TIMESTAMP NOT NULL,
				end_time TIMESTAMP,
				status VARCHAR(50) NOT NULL DEFAULT 'planned',
				start_lat DECIMAL(12, 8) NOT NULL,
				start_lng DECIMAL(12, 8) NOT NULL,
				start_alt DECIMAL(10, 2),
				target_lat DECIMAL(12, 8) NOT NULL,
				target_lng DECIMAL(12, 8) NOT NULL,
				target_alt DECIMAL(10, 2),
				actual_lat DECIMAL(12, 8),
				actual_lng DECIMAL(12, 8),
				actual_alt DECIMAL(10, 2),
				route_data JSONB,
				resource_usage JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("deployments")),

		// Strategic zones table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				name VARCHAR(200) NOT NULL,
				description TEXT,
				zone_type VARCHAR(50) NOT NULL,
				priority INTEGER NOT NULL DEFAULT 0,
				center_lat DECIMAL(12, 8) NOT NULL,
				center_lng DECIMAL(12, 8) NOT NULL,
				coverage_radius DECIMAL(10, 2) NOT NULL,
				max_units INTEGER NOT NULL DEFAULT 1,
				current_units INTEGER DEFAULT 0,
				status VARCHAR(50) NOT NULL DEFAULT 'active',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("strategic_zones")),

		// Alerts table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				alert_type VARCHAR(50) NOT NULL,
				severity VARCHAR(50) NOT NULL,
				title VARCHAR(200) NOT NULL,
				message TEXT NOT NULL,
				source VARCHAR(100) NOT NULL,
				entity_id VARCHAR(36),
				entity_type VARCHAR(100),
				is_read BOOLEAN DEFAULT false,
				is_acked BOOLEAN DEFAULT false,
				acked_by VARCHAR(100),
				acked_at TIMESTAMP,
				expires_at TIMESTAMP,
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("alerts")),

		// Commands table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				name VARCHAR(200) NOT NULL,
				command_type VARCHAR(100) NOT NULL,
				parent_id VARCHAR(36),
				location_lat DECIMAL(12, 8) NOT NULL,
				location_lng DECIMAL(12, 8) NOT NULL,
				location_alt DECIMAL(10, 2),
				status VARCHAR(50) NOT NULL DEFAULT 'active',
				contact_info JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("commands")),

		// Create indexes
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_unit_id ON %s(unit_id)`,
			r.tableName("telemetry_logs"), r.tableName("telemetry_logs")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp)`,
			r.tableName("telemetry_logs"), r.tableName("telemetry_logs")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_status ON %s(status)`,
			r.tableName("threats"), r.tableName("threats")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_severity ON %s(severity)`,
			r.tableName("threats"), r.tableName("threats")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_unit_id ON %s(unit_id)`,
			r.tableName("maintenance_logs"), r.tableName("maintenance_logs")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_status ON %s(status)`,
			r.tableName("maintenance_logs"), r.tableName("maintenance_logs")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_unit_id ON %s(unit_id)`,
			r.tableName("deployments"), r.tableName("deployments")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_status ON %s(status)`,
			r.tableName("deployments"), r.tableName("deployments")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_severity ON %s(severity)`,
			r.tableName("alerts"), r.tableName("alerts")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_is_read ON %s(is_read)`,
			r.tableName("alerts"), r.tableName("alerts")),
	}

	for _, schema := range schemas {
		if _, err := r.pool.Exec(ctx, schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

// CreateDefenseUnit creates a new defense unit
func (r *PostgresRepository) CreateDefenseUnit(ctx context.Context, unit *domain.DefenseUnit) error {
	if unit.ID == "" {
		unit.ID = uuid.New().String()
	}

	unit.CreatedAt = time.Now()
	unit.UpdatedAt = time.Now()

	metadata, err := json.Marshal(unit.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, unit_code, unit_name, unit_type, status, command_id,
			base_location_lat, base_location_lng, base_location_alt, last_contact_at,
			is_deployed, deployment_zone, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, r.tableName("units"))

	_, err = r.pool.Exec(ctx, query,
		unit.ID,
		unit.UnitCode,
		unit.UnitName,
		unit.UnitType,
		unit.Status,
		unit.CommandID,
		unit.Coordinates.Latitude,
		unit.Coordinates.Longitude,
		unit.Coordinates.Altitude,
		unit.LastContactAt,
		unit.IsDeployed,
		unit.DeploymentZone,
		metadata,
		unit.CreatedAt,
		unit.UpdatedAt,
	)

	return err
}

// GetDefenseUnit retrieves a defense unit by ID
func (r *PostgresRepository) GetDefenseUnit(ctx context.Context, id string) (*domain.DefenseUnit, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_code, unit_name, unit_type, status, command_id,
			base_location_lat, base_location_lng, base_location_alt, last_contact_at,
			is_deployed, deployment_zone, metadata, created_at, updated_at
		FROM %s WHERE id = $1
	`, r.tableName("units"))

	var unit domain.DefenseUnit
	var metadata []byte
	var cmdID sql.NullString
	var deployZone sql.NullString

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&unit.ID,
		&unit.UnitCode,
		&unit.UnitName,
		&unit.UnitType,
		&unit.Status,
		&cmdID,
		&unit.Coordinates.Latitude,
		&unit.Coordinates.Longitude,
		&unit.Coordinates.Altitude,
		&unit.LastContactAt,
		&unit.IsDeployed,
		&deployZone,
		&metadata,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if cmdID.Valid {
		unit.CommandID = cmdID.String
	}
	if deployZone.Valid {
		unit.DeploymentZone = deployZone.String
	}
	if metadata != nil {
		json.Unmarshal(metadata, &unit.Metadata)
	}

	return &unit, nil
}

// GetDefenseUnitByCode retrieves a defense unit by unit code
func (r *PostgresRepository) GetDefenseUnitByCode(ctx context.Context, code string) (*domain.DefenseUnit, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_code, unit_name, unit_type, status, command_id,
			base_location_lat, base_location_lng, base_location_alt, last_contact_at,
			is_deployed, deployment_zone, metadata, created_at, updated_at
		FROM %s WHERE unit_code = $1
	`, r.tableName("units"))

	var unit domain.DefenseUnit
	var metadata []byte
	var cmdID sql.NullString
	var deployZone sql.NullString

	err := r.pool.QueryRow(ctx, query, code).Scan(
		&unit.ID,
		&unit.UnitCode,
		&unit.UnitName,
		&unit.UnitType,
		&unit.Status,
		&cmdID,
		&unit.Coordinates.Latitude,
		&unit.Coordinates.Longitude,
		&unit.Coordinates.Altitude,
		&unit.LastContactAt,
		&unit.IsDeployed,
		&deployZone,
		&metadata,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if cmdID.Valid {
		unit.CommandID = cmdID.String
	}
	if deployZone.Valid {
		unit.DeploymentZone = deployZone.String
	}
	if metadata != nil {
		json.Unmarshal(metadata, &unit.Metadata)
	}

	return &unit, nil
}

// GetAllDefenseUnits retrieves all defense units
func (r *PostgresRepository) GetAllDefenseUnits(ctx context.Context) ([]*domain.DefenseUnit, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_code, unit_name, unit_type, status, command_id,
			base_location_lat, base_location_lng, base_location_alt, last_contact_at,
			is_deployed, deployment_zone, metadata, created_at, updated_at
		FROM %s ORDER BY unit_code
	`, r.tableName("units"))

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []*domain.DefenseUnit
	for rows.Next() {
		var unit domain.DefenseUnit
		var metadata []byte
		var cmdID sql.NullString
		var deployZone sql.NullString

		err := rows.Scan(
			&unit.ID,
			&unit.UnitCode,
			&unit.UnitName,
			&unit.UnitType,
			&unit.Status,
			&cmdID,
			&unit.Coordinates.Latitude,
			&unit.Coordinates.Longitude,
			&unit.Coordinates.Altitude,
			&unit.LastContactAt,
			&unit.IsDeployed,
			&deployZone,
			&metadata,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if cmdID.Valid {
			unit.CommandID = cmdID.String
		}
		if deployZone.Valid {
			unit.DeploymentZone = deployZone.String
		}
		if metadata != nil {
			json.Unmarshal(metadata, &unit.Metadata)
		}

		units = append(units, &unit)
	}

	return units, rows.Err()
}

// UpdateDefenseUnit updates a defense unit
func (r *PostgresRepository) UpdateDefenseUnit(ctx context.Context, unit *domain.DefenseUnit) error {
	unit.UpdatedAt = time.Now()

	query := fmt.Sprintf(`
		UPDATE %s SET
			status = $1, command_id = $2, last_contact_at = $3,
			is_deployed = $4, deployment_zone = $5, updated_at = $6
		WHERE id = $7
	`, r.tableName("units"))

	_, err := r.pool.Exec(ctx, query,
		unit.Status,
		unit.CommandID,
		unit.LastContactAt,
		unit.IsDeployed,
		unit.DeploymentZone,
		unit.UpdatedAt,
		unit.ID,
	)

	return err
}

// CreateTelemetryData creates a new telemetry data record
func (r *PostgresRepository) CreateTelemetryData(ctx context.Context, data *domain.TelemetryData) error {
	if data.ID == "" {
		data.ID = uuid.New().String()
	}

	data.CreatedAt = time.Now()
	data.ProcessedAt = time.Now()

	componentHealth, err := json.Marshal(data.ComponentHealth)
	if err != nil {
		componentHealth = []byte("{}")
	}

	payload, err := json.Marshal(data.Payload)
	if err != nil {
		payload = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, unit_id, timestamp, location_lat, location_lng, location_alt,
			location_heading, velocity, fuel_level, fuel_consumption, ammunition_count,
			hull_integrity, engine_temp, internal_temp, humidity, power_output,
			power_consumption, battery_level, signal_strength, encryption_level,
			component_health, payload, processed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
	`, r.tableName("telemetry_logs"))

	_, err = r.pool.Exec(ctx, query,
		data.ID,
		data.UnitID,
		data.Timestamp,
		data.Coordinates.Latitude,
		data.Coordinates.Longitude,
		data.Coordinates.Altitude,
		data.Coordinates.Heading,
		data.Velocity,
		data.FuelLevel,
		data.FuelConsumption,
		data.AmmunitionCount,
		data.HullIntegrity,
		data.EngineTemp,
		data.InternalTemp,
		data.Humidity,
		data.PowerOutput,
		data.PowerConsumption,
		data.BatteryLevel,
		data.SignalStrength,
		data.EncryptionLevel,
		componentHealth,
		payload,
		data.ProcessedAt,
		data.CreatedAt,
	)

	return err
}

// GetLatestTelemetry retrieves the latest telemetry for a unit
func (r *PostgresRepository) GetLatestTelemetry(ctx context.Context, unitID string) (*domain.TelemetryData, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_id, timestamp, location_lat, location_lng, location_alt,
			location_heading, velocity, fuel_level, fuel_consumption, ammunition_count,
			hull_integrity, engine_temp, internal_temp, humidity, power_output,
			power_consumption, battery_level, signal_strength, encryption_level,
			component_health, payload, processed_at, created_at
		FROM %s WHERE unit_id = $1
		ORDER BY timestamp DESC LIMIT 1
	`, r.tableName("telemetry_logs"))

	var data domain.TelemetryData
	var componentHealth []byte
	var payload []byte

	err := r.pool.QueryRow(ctx, query, unitID).Scan(
		&data.ID,
		&data.UnitID,
		&data.Timestamp,
		&data.Coordinates.Latitude,
		&data.Coordinates.Longitude,
		&data.Coordinates.Altitude,
		&data.Coordinates.Heading,
		&data.Velocity,
		&data.FuelLevel,
		&data.FuelConsumption,
		&data.AmmunitionCount,
		&data.HullIntegrity,
		&data.EngineTemp,
		&data.InternalTemp,
		&data.Humidity,
		&data.PowerOutput,
		&data.PowerConsumption,
		&data.BatteryLevel,
		&data.SignalStrength,
		&data.EncryptionLevel,
		&componentHealth,
		&payload,
		&data.ProcessedAt,
		&data.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if componentHealth != nil {
		json.Unmarshal(componentHealth, &data.ComponentHealth)
	}
	if payload != nil {
		json.Unmarshal(payload, &data.Payload)
	}

	return &data, nil
}

// CreateThreatEvent creates a new threat event
func (r *PostgresRepository) CreateThreatEvent(ctx context.Context, threat *domain.ThreatEvent) error {
	if threat.ID == "" {
		threat.ID = uuid.New().String()
	}

	threat.CreatedAt = time.Now()
	threat.FirstSeenAt = time.Now()
	threat.DetectionTime = time.Now()
	threat.LastUpdatedAt = time.Now()

	countermeasures, err := json.Marshal(threat.Countermeasures)
	if err != nil {
		countermeasures = []byte("[]")
	}

	metadata, err := json.Marshal(threat.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, unit_id, threat_type, threat_source, severity, status,
			source_lat, source_lng, source_alt, target_lat, target_lng, target_alt,
			velocity, estimated_impact, detection_time, first_seen_at, last_updated_at,
			resolved_at, resolution, countermeasures, escalation_level, assigned_to,
			notes, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
	`, r.tableName("threats"))

	_, err = r.pool.Exec(ctx, query,
		threat.ID,
		threat.UnitID,
		threat.ThreatType,
		threat.ThreatSource,
		threat.Severity,
		threat.Status,
		threat.SourceLocation.Latitude,
		threat.SourceLocation.Longitude,
		threat.SourceLocation.Altitude,
		threat.TargetLocation.Latitude,
		threat.TargetLocation.Longitude,
		threat.TargetLocation.Altitude,
		threat.Velocity,
		threat.EstimatedImpact,
		threat.DetectionTime,
		threat.FirstSeenAt,
		threat.LastUpdatedAt,
		threat.ResolvedAt,
		threat.Resolution,
		countermeasures,
		threat.EscalationLevel,
		threat.AssignedTo,
		threat.Notes,
		metadata,
		threat.CreatedAt,
	)

	return err
}

// GetActiveThreats retrieves all active threats
func (r *PostgresRepository) GetActiveThreats(ctx context.Context, minSeverity int) ([]*domain.ThreatEvent, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_id, threat_type, threat_source, severity, status,
			source_lat, source_lng, source_alt, target_lat, target_lng, target_alt,
			velocity, estimated_impact, detection_time, first_seen_at, last_updated_at,
			resolved_at, resolution, countermeasures, escalation_level, assigned_to,
			notes, metadata, created_at
		FROM %s
		WHERE severity >= $1 AND status NOT IN ('resolved', 'false_alarm')
		ORDER BY severity DESC, detection_time DESC
	`, r.tableName("threats"))

	rows, err := r.pool.Query(ctx, query, minSeverity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threats []*domain.ThreatEvent
	for rows.Next() {
		var threat domain.ThreatEvent
		var countermeasures []byte
		var metadata []byte
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&threat.ID,
			&threat.UnitID,
			&threat.ThreatType,
			&threat.ThreatSource,
			&threat.Severity,
			&threat.Status,
			&threat.SourceLocation.Latitude,
			&threat.SourceLocation.Longitude,
			&threat.SourceLocation.Altitude,
			&threat.TargetLocation.Latitude,
			&threat.TargetLocation.Longitude,
			&threat.TargetLocation.Altitude,
			&threat.Velocity,
			&threat.EstimatedImpact,
			&threat.DetectionTime,
			&threat.FirstSeenAt,
			&threat.LastUpdatedAt,
			&resolvedAt,
			&threat.Resolution,
			&countermeasures,
			&threat.EscalationLevel,
			&threat.AssignedTo,
			&threat.Notes,
			&metadata,
			&threat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			threat.ResolvedAt = &resolvedAt.Time
		}
		if countermeasures != nil {
			json.Unmarshal(countermeasures, &threat.Countermeasures)
		}
		if metadata != nil {
			json.Unmarshal(metadata, &threat.Metadata)
		}

		threats = append(threats, &threat)
	}

	return threats, rows.Err()
}

// UpdateThreatEvent updates a threat event
func (r *PostgresRepository) UpdateThreatEvent(ctx context.Context, threat *domain.ThreatEvent) error {
	threat.LastUpdatedAt = time.Now()

	query := fmt.Sprintf(`
		UPDATE %s SET
			status = $1, escalation_level = $2, assigned_to = $3,
			resolved_at = $4, resolution = $5, last_updated_at = $6
		WHERE id = $7
	`, r.tableName("threats"))

	_, err := r.pool.Exec(ctx, query,
		threat.Status,
		threat.EscalationLevel,
		threat.AssignedTo,
		threat.ResolvedAt,
		threat.Resolution,
		threat.LastUpdatedAt,
		threat.ID,
	)

	return err
}

// CreateMaintenanceLog creates a new maintenance log
func (r *PostgresRepository) CreateMaintenanceLog(ctx context.Context, log *domain.MaintenanceLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()

	partsReplaced, err := json.Marshal(log.PartsReplaced)
	if err != nil {
		partsReplaced = []byte("[]")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, unit_id, component, maintenance_type, condition_score,
			previous_condition, work_performed, parts_replaced, technician_id,
			estimated_life, scheduled_at, started_at, completed_at, next_service_due,
			predicted_failure_date, status, cost, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`, r.tableName("maintenance_logs"))

	_, err = r.pool.Exec(ctx, query,
		log.ID,
		log.UnitID,
		log.Component,
		log.MaintenanceType,
		log.ConditionScore,
		log.PreviousCondition,
		log.WorkPerformed,
		partsReplaced,
		log.TechnicianID,
		log.EstimatedLife,
		log.ScheduledAt,
		log.StartedAt,
		log.CompletedAt,
		log.NextServiceDue,
		log.PredictedFailureDate,
		log.Status,
		log.Cost,
		log.Notes,
		log.CreatedAt,
		log.UpdatedAt,
	)

	return err
}

// GetMaintenancePredictions retrieves maintenance predictions for a unit
func (r *PostgresRepository) GetMaintenancePredictions(ctx context.Context, unitID string) ([]*domain.MaintenancePrediction, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_id, component, current_health, predicted_failure_date,
			confidence_level, risk_level, recommendations, model_version, prediction_date, created_at
		FROM %s WHERE unit_id = $1 AND predicted_failure_date > NOW()
		ORDER BY predicted_failure_date ASC
	`, r.tableName("maintenance_predictions"))

	rows, err := r.pool.Query(ctx, query, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predictions []*domain.MaintenancePrediction
	for rows.Next() {
		var pred domain.MaintenancePrediction
		var recommendations []byte

		err := rows.Scan(
			&pred.ID,
			&pred.UnitID,
			&pred.Component,
			&pred.CurrentHealth,
			&pred.PredictedFailureDate,
			&pred.ConfidenceLevel,
			&pred.RiskLevel,
			&recommendations,
			&pred.ModelVersion,
			&pred.PredictionDate,
			&pred.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if recommendations != nil {
			json.Unmarshal(recommendations, &pred.Recommendations)
		}

		predictions = append(predictions, &pred)
	}

	return predictions, rows.Err()
}

// CreateDeployment creates a new deployment
func (r *PostgresRepository) CreateDeployment(ctx context.Context, deployment *domain.Deployment) error {
	if deployment.ID == "" {
		deployment.ID = uuid.New().String()
	}

	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()

	routeData, err := json.Marshal(deployment.RoutePlanned)
	if err != nil {
		routeData = []byte("{}")
	}

	resourceUsage, err := json.Marshal(deployment.ResourceUsage)
	if err != nil {
		resourceUsage = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, unit_id, mission_id, zone_id, start_time, end_time, status,
			start_lat, start_lng, start_alt, target_lat, target_lng, target_alt,
			actual_lat, actual_lng, actual_alt, route_data, resource_usage, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`, r.tableName("deployments"))

	_, err = r.pool.Exec(ctx, query,
		deployment.ID,
		deployment.UnitID,
		deployment.MissionID,
		deployment.ZoneID,
		deployment.StartTime,
		deployment.EndTime,
		deployment.Status,
		deployment.StartLocation.Latitude,
		deployment.StartLocation.Longitude,
		deployment.StartLocation.Altitude,
		deployment.TargetLocation.Latitude,
		deployment.TargetLocation.Longitude,
		deployment.TargetLocation.Altitude,
		deployment.ActualLocation.Latitude,
		deployment.ActualLocation.Longitude,
		deployment.ActualLocation.Altitude,
		routeData,
		resourceUsage,
		deployment.CreatedAt,
		deployment.UpdatedAt,
	)

	return err
}

// GetUnitDeployments retrieves deployments for a unit
func (r *PostgresRepository) GetUnitDeployments(ctx context.Context, unitID string) ([]*domain.Deployment, error) {
	query := fmt.Sprintf(`
		SELECT id, unit_id, mission_id, zone_id, start_time, end_time, status,
			start_lat, start_lng, start_alt, target_lat, target_lng, target_alt,
			actual_lat, actual_lng, actual_alt, route_data, resource_usage, created_at, updated_at
		FROM %s WHERE unit_id = $1 ORDER BY start_time DESC
	`, r.tableName("deployments"))

	rows, err := r.pool.Query(ctx, query, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*domain.Deployment
	for rows.Next() {
		var d domain.Deployment
		var routeData []byte
		var resourceUsage []byte

		err := rows.Scan(
			&d.ID,
			&d.UnitID,
			&d.MissionID,
			&d.ZoneID,
			&d.StartTime,
			&d.EndTime,
			&d.Status,
			&d.StartLocation.Latitude,
			&d.StartLocation.Longitude,
			&d.StartLocation.Altitude,
			&d.TargetLocation.Latitude,
			&d.TargetLocation.Longitude,
			&d.TargetLocation.Altitude,
			&d.ActualLocation.Latitude,
			&d.ActualLocation.Longitude,
			&d.ActualLocation.Altitude,
			&routeData,
			&resourceUsage,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if routeData != nil {
			json.Unmarshal(routeData, &d.RoutePlanned)
		}
		if resourceUsage != nil {
			json.Unmarshal(resourceUsage, &d.ResourceUsage)
		}

		deployments = append(deployments, &d)
	}

	return deployments, rows.Err()
}

// CreateAlert creates a new alert
func (r *PostgresRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}

	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	metadata, err := json.Marshal(alert.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, alert_type, severity, title, message, source,
			entity_id, entity_type, is_read, is_acked, acked_by, acked_at,
			expires_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
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
		alert.IsAcked,
		alert.AckedBy,
		alert.AckedAt,
		alert.ExpiresAt,
		metadata,
		alert.CreatedAt,
		alert.UpdatedAt,
	)

	return err
}

// GetActiveAlerts retrieves active alerts
func (r *PostgresRepository) GetActiveAlerts(ctx context.Context, minSeverity string) ([]*domain.Alert, error) {
	query := fmt.Sprintf(`
		SELECT id, alert_type, severity, title, message, source,
			entity_id, entity_type, is_read, is_acked, acked_by, acked_at,
			expires_at, metadata, created_at, updated_at
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
		var entityID, entityType, ackedBy sql.NullString
		var ackedAt sql.NullTime
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
			&alert.IsAcked,
			&ackedBy,
			&ackedAt,
			&expiresAt,
			&metadata,
			&alert.CreatedAt,
			&alert.UpdatedAt,
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
		if ackedBy.Valid {
			alert.AckedBy = ackedBy.String
		}
		if ackedAt.Valid {
			alert.AckedAt = &ackedAt.Time
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

// CreateStrategicZone creates a new strategic zone
func (r *PostgresRepository) CreateStrategicZone(ctx context.Context, zone *domain.StrategicZone) error {
	if zone.ID == "" {
		zone.ID = uuid.New().String()
	}

	zone.CreatedAt = time.Now()
	zone.UpdatedAt = time.Now()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, description, zone_type, priority,
			center_lat, center_lng, coverage_radius, max_units, current_units,
			status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, r.tableName("strategic_zones"))

	_, err := r.pool.Exec(ctx, query,
		zone.ID,
		zone.Name,
		zone.Description,
		zone.ZoneType,
		zone.Priority,
		zone.CenterPoint.Latitude,
		zone.CenterPoint.Longitude,
		zone.CoverageRadius,
		zone.MaxUnits,
		zone.CurrentUnits,
		zone.Status,
		zone.CreatedAt,
		zone.UpdatedAt,
	)

	return err
}

// GetStrategicZones retrieves all strategic zones
func (r *PostgresRepository) GetStrategicZones(ctx context.Context) ([]*domain.StrategicZone, error) {
	query := fmt.Sprintf(`
		SELECT id, name, description, zone_type, priority,
			center_lat, center_lng, coverage_radius, max_units, current_units,
			status, created_at, updated_at
		FROM %s ORDER BY priority DESC
	`, r.tableName("strategic_zones"))

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*domain.StrategicZone
	for rows.Next() {
		var zone domain.StrategicZone

		err := rows.Scan(
			&zone.ID,
			&zone.Name,
			&zone.Description,
			&zone.ZoneType,
			&zone.Priority,
			&zone.CenterPoint.Latitude,
			&zone.CenterPoint.Longitude,
			&zone.CoverageRadius,
			&zone.MaxUnits,
			&zone.CurrentUnits,
			&zone.Status,
			&zone.CreatedAt,
			&zone.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		zones = append(zones, &zone)
	}

	return zones, rows.Err()
}
