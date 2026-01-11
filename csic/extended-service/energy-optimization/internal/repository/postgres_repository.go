package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"energy-optimization/internal/domain"

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

// initSchema creates the required database tables
func (r *PostgresRepository) initSchema() error {
	ctx := context.Background()

	schemas := []string{
		// Energy readings table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				meter_id VARCHAR(100) NOT NULL,
				facility_id VARCHAR(100) NOT NULL,
				energy_source VARCHAR(50) NOT NULL,
				consumption DECIMAL(20, 6) NOT NULL,
				voltage DECIMAL(10, 4),
				current DECIMAL(10, 4),
				power_factor DECIMAL(5, 4),
				frequency DECIMAL(8, 4),
				timestamp TIMESTAMP NOT NULL,
				quality VARCHAR(20) DEFAULT 'good',
				metadata JSONB,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("energy_readings")),

		// Carbon emissions table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				reading_id VARCHAR(36) NOT NULL,
				facility_id VARCHAR(100) NOT NULL,
				energy_source VARCHAR(50) NOT NULL,
				emission_factor DECIMAL(10, 6) NOT NULL,
				consumption DECIMAL(20, 6) NOT NULL,
				total_emission DECIMAL(20, 6) NOT NULL,
				carbon_intensity DECIMAL(10, 2) NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("carbon_emissions")),

		// Carbon budgets table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				facility_id VARCHAR(100) NOT NULL,
				budget_period VARCHAR(20) NOT NULL,
				budget_amount DECIMAL(20, 2) NOT NULL,
				used_amount DECIMAL(20, 2) DEFAULT 0,
				remaining_amount DECIMAL(20, 2) NOT NULL,
				usage_percentage DECIMAL(5, 2) DEFAULT 0,
				status VARCHAR(20) DEFAULT 'normal',
				warning_threshold DECIMAL(5, 2) DEFAULT 80,
				critical_threshold DECIMAL(5, 2) DEFAULT 95,
				period_start TIMESTAMP NOT NULL,
				period_end TIMESTAMP NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("carbon_budgets")),

		// Load metrics table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				facility_id VARCHAR(100) NOT NULL,
				zone_id VARCHAR(100),
				timestamp TIMESTAMP NOT NULL,
				window_duration INTERVAL NOT NULL,
				avg_consumption DECIMAL(20, 6) NOT NULL,
				peak_consumption DECIMAL(20, 6) NOT NULL,
				min_consumption DECIMAL(20, 6) NOT NULL,
				total_consumption DECIMAL(20, 6) NOT NULL,
				avg_utilization DECIMAL(5, 2) NOT NULL,
				peak_utilization DECIMAL(5, 2) NOT NULL,
				reactive_power DECIMAL(20, 6),
				harmonic_distortion DECIMAL(8, 4),
				power_factor DECIMAL(5, 4),
				sample_count INTEGER DEFAULT 0
			)
		`, r.tableName("load_metrics")),

		// Load balancing decisions table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				facility_id VARCHAR(100) NOT NULL,
				decision_type VARCHAR(50) NOT NULL,
				source_zone_id VARCHAR(100),
				target_zone_id VARCHAR(100),
				load_shift_amount DECIMAL(20, 6) NOT NULL,
				reason TEXT,
				triggering_condition VARCHAR(100),
				priority INTEGER DEFAULT 0,
				status VARCHAR(20) DEFAULT 'pending',
				scheduled_at TIMESTAMP NOT NULL,
				executed_at TIMESTAMP,
				cancelled_at TIMESTAMP,
				result TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("load_balancing_decisions")),

		// Energy facilities table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				name VARCHAR(200) NOT NULL,
				location VARCHAR(500),
				timezone VARCHAR(50) DEFAULT 'UTC',
				total_capacity DECIMAL(20, 2) NOT NULL,
				current_load DECIMAL(20, 2) DEFAULT 0,
				utilization_rate DECIMAL(5, 2) DEFAULT 0,
				carbon_budget_id VARCHAR(36),
				is_active BOOLEAN DEFAULT true,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("energy_facilities")),

		// Energy zones table
		fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				id VARCHAR(36) PRIMARY KEY,
				facility_id VARCHAR(36) NOT NULL,
				name VARCHAR(200) NOT NULL,
				zone_type VARCHAR(50) NOT NULL,
				capacity DECIMAL(20, 2) NOT NULL,
				current_load DECIMAL(20, 2) DEFAULT 0,
				priority INTEGER DEFAULT 0,
				can_shed_load BOOLEAN DEFAULT false,
				min_load DECIMAL(20, 2) DEFAULT 0,
				max_load DECIMAL(20, 2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, r.tableName("energy_zones")),

		// Create indexes
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp)`,
			r.tableName("energy_readings"), r.tableName("energy_readings")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_facility_timestamp ON %s(facility_id, timestamp)`,
			r.tableName("energy_readings"), r.tableName("energy_readings")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_meter_id ON %s(meter_id)`,
			r.tableName("energy_readings"), r.tableName("energy_readings")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_facility_id ON %s(facility_id)`,
			r.tableName("carbon_emissions"), r.tableName("carbon_emissions")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_facility_id ON %s(facility_id)`,
			r.tableName("load_metrics"), r.tableName("load_metrics")),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_status ON %s(status)`,
			r.tableName("load_balancing_decisions"), r.tableName("load_balancing_decisions")),
	}

	for _, schema := range schemas {
		if _, err := r.pool.Exec(ctx, schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

// CreateEnergyReading inserts a new energy reading
func (r *PostgresRepository) CreateEnergyReading(ctx context.Context, reading *domain.EnergyReading) error {
	if reading.ID == "" {
		reading.ID = uuid.New().String()
	}

	metadata, err := json.Marshal(reading.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, meter_id, facility_id, energy_source, consumption, voltage, current,
			power_factor, frequency, timestamp, quality, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, r.tableName("energy_readings"))

	_, err = r.pool.Exec(ctx, query,
		reading.ID,
		reading.MeterID,
		reading.FacilityID,
		reading.EnergySource,
		reading.Consumption,
		reading.Voltage,
		reading.Current,
		reading.PowerFactor,
		reading.Frequency,
		reading.Timestamp,
		reading.Quality,
		metadata,
		reading.CreatedAt,
	)

	return err
}

// GetEnergyReading retrieves an energy reading by ID
func (r *PostgresRepository) GetEnergyReading(ctx context.Context, id string) (*domain.EnergyReading, error) {
	query := fmt.Sprintf(`
		SELECT id, meter_id, facility_id, energy_source, consumption, voltage, current,
			power_factor, frequency, timestamp, quality, metadata, created_at
		FROM %s WHERE id = $1
	`, r.tableName("energy_readings"))

	var reading domain.EnergyReading
	var metadata []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&reading.ID,
		&reading.MeterID,
		&reading.FacilityID,
		&reading.EnergySource,
		&reading.Consumption,
		&reading.Voltage,
		&reading.Current,
		&reading.PowerFactor,
		&reading.Frequency,
		&reading.Timestamp,
		&reading.Quality,
		&metadata,
		&reading.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if metadata != nil {
		json.Unmarshal(metadata, &reading.Metadata)
	}

	return &reading, nil
}

// GetEnergyReadingsByFacility retrieves energy readings for a facility within a time range
func (r *PostgresRepository) GetEnergyReadingsByFacility(ctx context.Context, facilityID string, startTime, endTime time.Time, limit int) ([]*domain.EnergyReading, error) {
	query := fmt.Sprintf(`
		SELECT id, meter_id, facility_id, energy_source, consumption, voltage, current,
			power_factor, frequency, timestamp, quality, metadata, created_at
		FROM %s
		WHERE facility_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
		LIMIT $4
	`, r.tableName("energy_readings"))

	rows, err := r.pool.Query(ctx, query, facilityID, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []*domain.EnergyReading
	for rows.Next() {
		var reading domain.EnergyReading
		var metadata []byte

		err := rows.Scan(
			&reading.ID,
			&reading.MeterID,
			&reading.FacilityID,
			&reading.EnergySource,
			&reading.Consumption,
			&reading.Voltage,
			&reading.Current,
			&reading.PowerFactor,
			&reading.Frequency,
			&reading.Timestamp,
			&reading.Quality,
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

// CreateCarbonEmission inserts a new carbon emission record
func (r *PostgresRepository) CreateCarbonEmission(ctx context.Context, emission *domain.CarbonEmission) error {
	if emission.ID == "" {
		emission.ID = uuid.New().String()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, reading_id, facility_id, energy_source, emission_factor,
			consumption, total_emission, carbon_intensity, timestamp, calculated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.tableName("carbon_emissions"))

	_, err := r.pool.Exec(ctx, query,
		emission.ID,
		emission.ReadingID,
		emission.FacilityID,
		emission.EnergySource,
		emission.EmissionFactor,
		emission.Consumption,
		emission.TotalEmission,
		emission.CarbonIntensity,
		emission.Timestamp,
		emission.CalculatedAt,
	)

	return err
}

// GetCarbonEmissionsByFacility retrieves carbon emissions for a facility
func (r *PostgresRepository) GetCarbonEmissionsByFacility(ctx context.Context, facilityID string, startTime, endTime time.Time) ([]*domain.CarbonEmission, error) {
	query := fmt.Sprintf(`
		SELECT id, reading_id, facility_id, energy_source, emission_factor,
			consumption, total_emission, carbon_intensity, timestamp, calculated_at
		FROM %s
		WHERE facility_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`, r.tableName("carbon_emissions"))

	rows, err := r.pool.Query(ctx, query, facilityID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emissions []*domain.CarbonEmission
	for rows.Next() {
		var emission domain.CarbonEmission
		err := rows.Scan(
			&emission.ID,
			&emission.ReadingID,
			&emission.FacilityID,
			&emission.EnergySource,
			&emission.EmissionFactor,
			&emission.Consumption,
			&emission.TotalEmission,
			&emission.CarbonIntensity,
			&emission.Timestamp,
			&emission.CalculatedAt,
		)
		if err != nil {
			return nil, err
		}
		emissions = append(emissions, &emission)
	}

	return emissions, rows.Err()
}

// CreateOrUpdateCarbonBudget creates or updates a carbon budget
func (r *PostgresRepository) CreateOrUpdateCarbonBudget(ctx context.Context, budget *domain.CarbonBudget) error {
	if budget.ID == "" {
		budget.ID = uuid.New().String()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, facility_id, budget_period, budget_amount, used_amount,
			remaining_amount, usage_percentage, status, warning_threshold, critical_threshold,
			period_start, period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			used_amount = EXCLUDED.used_amount,
			remaining_amount = EXCLUDED.remaining_amount,
			usage_percentage = EXCLUDED.usage_percentage,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, r.tableName("carbon_budgets"))

	_, err := r.pool.Exec(ctx, query,
		budget.ID,
		budget.FacilityID,
		budget.BudgetPeriod,
		budget.BudgetAmount,
		budget.UsedAmount,
		budget.RemainingAmount,
		budget.UsagePercentage,
		budget.Status,
		budget.WarningThreshold,
		budget.CriticalThreshold,
		budget.PeriodStart,
		budget.PeriodEnd,
		budget.CreatedAt,
		budget.UpdatedAt,
	)

	return err
}

// GetCarbonBudget retrieves the current carbon budget for a facility
func (r *PostgresRepository) GetCarbonBudget(ctx context.Context, facilityID string) (*domain.CarbonBudget, error) {
	query := fmt.Sprintf(`
		SELECT id, facility_id, budget_period, budget_amount, used_amount,
			remaining_amount, usage_percentage, status, warning_threshold, critical_threshold,
			period_start, period_end, created_at, updated_at
		FROM %s
		WHERE facility_id = $1 AND period_start <= NOW() AND period_end >= NOW()
		LIMIT 1
	`, r.tableName("carbon_budgets"))

	var budget domain.CarbonBudget
	err := r.pool.QueryRow(ctx, query, facilityID).Scan(
		&budget.ID,
		&budget.FacilityID,
		&budget.BudgetPeriod,
		&budget.BudgetAmount,
		&budget.UsedAmount,
		&budget.RemainingAmount,
		&budget.UsagePercentage,
		&budget.Status,
		&budget.WarningThreshold,
		&budget.CriticalThreshold,
		&budget.PeriodStart,
		&budget.PeriodEnd,
		&budget.CreatedAt,
		&budget.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &budget, nil
}

// CreateLoadMetric inserts a new load metric
func (r *PostgresRepository) CreateLoadMetric(ctx context.Context, metric *domain.LoadMetric) error {
	if metric.ID == "" {
		metric.ID = uuid.New().String()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, facility_id, zone_id, timestamp, window_duration,
			avg_consumption, peak_consumption, min_consumption, total_consumption,
			avg_utilization, peak_utilization, reactive_power, harmonic_distortion,
			power_factor, sample_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, r.tableName("load_metrics"))

	_, err := r.pool.Exec(ctx, query,
		metric.ID,
		metric.FacilityID,
		metric.ZoneID,
		metric.Timestamp,
		metric.WindowDuration,
		metric.AvgConsumption,
		metric.PeakConsumption,
		metric.MinConsumption,
		metric.TotalConsumption,
		metric.AvgUtilization,
		metric.PeakUtilization,
		metric.ReactivePower,
		metric.HarmonicDistortion,
		metric.PowerFactor,
		metric.SampleCount,
	)

	return err
}

// GetLatestLoadMetrics retrieves the latest load metrics for a facility
func (r *PostgresRepository) GetLatestLoadMetrics(ctx context.Context, facilityID string) ([]*domain.LoadMetric, error) {
	query := fmt.Sprintf(`
		SELECT id, facility_id, zone_id, timestamp, window_duration,
			avg_consumption, peak_consumption, min_consumption, total_consumption,
			avg_utilization, peak_utilization, reactive_power, harmonic_distortion,
			power_factor, sample_count
		FROM %s
		WHERE facility_id = $1
		ORDER BY timestamp DESC
	`, r.tableName("load_metrics"))

	rows, err := r.pool.Query(ctx, query, facilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*domain.LoadMetric
	for rows.Next() {
		var metric domain.LoadMetric
		err := rows.Scan(
			&metric.ID,
			&metric.FacilityID,
			&metric.ZoneID,
			&metric.Timestamp,
			&metric.WindowDuration,
			&metric.AvgConsumption,
			&metric.PeakConsumption,
			&metric.MinConsumption,
			&metric.TotalConsumption,
			&metric.AvgUtilization,
			&metric.PeakUtilization,
			&metric.ReactivePower,
			&metric.HarmonicDistortion,
			&metric.PowerFactor,
			&metric.SampleCount,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, &metric)
	}

	return metrics, rows.Err()
}

// CreateLoadBalancingDecision inserts a new load balancing decision
func (r *PostgresRepository) CreateLoadBalancingDecision(ctx context.Context, decision *domain.LoadBalancingDecision) error {
	if decision.ID == "" {
		decision.ID = uuid.New().String()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, facility_id, decision_type, source_zone_id, target_zone_id,
			load_shift_amount, reason, triggering_condition, priority, status,
			scheduled_at, executed_at, cancelled_at, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, r.tableName("load_balancing_decisions"))

	_, err := r.pool.Exec(ctx, query,
		decision.ID,
		decision.FacilityID,
		decision.DecisionType,
		decision.SourceZoneID,
		decision.TargetZoneID,
		decision.LoadShiftAmount,
		decision.Reason,
		decision.TriggeringCondition,
		decision.Priority,
		decision.Status,
		decision.ScheduledAt,
		decision.ExecutedAt,
		decision.CancelledAt,
		decision.Result,
		decision.CreatedAt,
		decision.UpdatedAt,
	)

	return err
}

// UpdateLoadBalancingDecision updates the status of a load balancing decision
func (r *PostgresRepository) UpdateLoadBalancingDecision(ctx context.Context, id string, status domain.DecisionStatus, result string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, result = $2, updated_at = NOW()
		WHERE id = $3
	`, r.tableName("load_balancing_decisions"))

	_, err := r.pool.Exec(ctx, query, status, result, id)
	return err
}

// GetPendingDecisions retrieves pending load balancing decisions
func (r *PostgresRepository) GetPendingDecisions(ctx context.Context, facilityID string) ([]*domain.LoadBalancingDecision, error) {
	query := fmt.Sprintf(`
		SELECT id, facility_id, decision_type, source_zone_id, target_zone_id,
			load_shift_amount, reason, triggering_condition, priority, status,
			scheduled_at, executed_at, cancelled_at, result, created_at, updated_at
		FROM %s
		WHERE facility_id = $1 AND status = 'pending' AND scheduled_at <= NOW()
		ORDER BY priority DESC, scheduled_at ASC
	`, r.tableName("load_balancing_decisions"))

	rows, err := r.pool.Query(ctx, query, facilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*domain.LoadBalancingDecision
	for rows.Next() {
		var decision domain.LoadBalancingDecision
		err := rows.Scan(
			&decision.ID,
			&decision.FacilityID,
			&decision.DecisionType,
			&decision.SourceZoneID,
			&decision.TargetZoneID,
			&decision.LoadShiftAmount,
			&decision.Reason,
			&decision.TriggeringCondition,
			&decision.Priority,
			&decision.Status,
			&decision.ScheduledAt,
			&decision.ExecutedAt,
			&decision.CancelledAt,
			&decision.Result,
			&decision.CreatedAt,
			&decision.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		decisions = append(decisions, &decision)
	}

	return decisions, rows.Err()
}

// CreateEnergyFacility inserts a new energy facility
func (r *PostgresRepository) CreateEnergyFacility(ctx context.Context, facility *domain.EnergyFacility) error {
	if facility.ID == "" {
		facility.ID = uuid.New().String()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, location, timezone, total_capacity, current_load,
			utilization_rate, carbon_budget_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, r.tableName("energy_facilities"))

	_, err := r.pool.Exec(ctx, query,
		facility.ID,
		facility.Name,
		facility.Location,
		facility.Timezone,
		facility.TotalCapacity,
		facility.CurrentLoad,
		facility.UtilizationRate,
		facility.CarbonBudgetID,
		facility.IsActive,
		facility.CreatedAt,
		facility.UpdatedAt,
	)

	return err
}

// GetEnergyFacility retrieves an energy facility by ID
func (r *PostgresRepository) GetEnergyFacility(ctx context.Context, id string) (*domain.EnergyFacility, error) {
	query := fmt.Sprintf(`
		SELECT id, name, location, timezone, total_capacity, current_load,
			utilization_rate, carbon_budget_id, is_active, created_at, updated_at
		FROM %s WHERE id = $1
	`, r.tableName("energy_facilities"))

	var facility domain.EnergyFacility
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&facility.ID,
		&facility.Name,
		&facility.Location,
		&facility.Timezone,
		&facility.TotalCapacity,
		&facility.CurrentLoad,
		&facility.UtilizationRate,
		&facility.CarbonBudgetID,
		&facility.IsActive,
		&facility.CreatedAt,
		&facility.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &facility, nil
}

// CreateEnergyZone inserts a new energy zone
func (r *PostgresRepository) CreateEnergyZone(ctx context.Context, zone *domain.EnergyZone) error {
	if zone.ID == "" {
		zone.ID = uuid.New().String()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, facility_id, name, zone_type, capacity, current_load,
			priority, can_shed_load, min_load, max_load, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.tableName("energy_zones"))

	_, err := r.pool.Exec(ctx, query,
		zone.ID,
		zone.FacilityID,
		zone.Name,
		zone.ZoneType,
		zone.Capacity,
		zone.CurrentLoad,
		zone.Priority,
		zone.CanShedLoad,
		zone.MinLoad,
		zone.MaxLoad,
		zone.CreatedAt,
		zone.UpdatedAt,
	)

	return err
}

// GetEnergyZones retrieves all zones for a facility
func (r *PostgresRepository) GetEnergyZones(ctx context.Context, facilityID string) ([]*domain.EnergyZone, error) {
	query := fmt.Sprintf(`
		SELECT id, facility_id, name, zone_type, capacity, current_load,
			priority, can_shed_load, min_load, max_load, created_at, updated_at
		FROM %s WHERE facility_id = $1 ORDER BY priority DESC
	`, r.tableName("energy_zones"))

	rows, err := r.pool.Query(ctx, query, facilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []*domain.EnergyZone
	for rows.Next() {
		var zone domain.EnergyZone
		err := rows.Scan(
			&zone.ID,
			&zone.FacilityID,
			&zone.Name,
			&zone.ZoneType,
			&zone.Capacity,
			&zone.CurrentLoad,
			&zone.Priority,
			&zone.CanShedLoad,
			&zone.MinLoad,
			&zone.MaxLoad,
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

// Helper to work with decimal from database
type nullDecimal struct {
	decimal.Decimal
	Valid bool
}

func (n *nullDecimal) Scan(value interface{}) error {
	if value == nil {
		n.Decimal = decimal.Zero
		n.Valid = false
		return nil
	}

	switch v := value.(type) {
	case float64:
		n.Decimal = decimal.NewFromFloat(v)
		n.Valid = true
	case string:
		d, err := decimal.NewFromString(v)
		if err != nil {
			return err
		}
		n.Decimal = d
		n.Valid = true
	default:
		return fmt.Errorf("cannot scan type %T into nullDecimal", value)
	}
	return nil
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
