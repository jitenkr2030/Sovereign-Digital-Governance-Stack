package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TimescaleEnergyRepository implements EnergyRepository for TimescaleDB
type TimescaleEnergyRepository struct {
	db        *sql.DB
	chunkDays int
}

// NewTimescaleEnergyRepository creates a new TimescaleDB energy repository
func NewTimescaleEnergyRepository(config PostgresConfig) (*TimescaleEnergyRepository, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TimescaleEnergyRepository{
		db:        db,
		chunkDays: 7,
	}, nil
}

// Close closes the database connection
func (r *TimescaleEnergyRepository) Close() error {
	return r.db.Close()
}

// StoreLog stores an energy consumption log
func (r *TimescaleEnergyRepository) StoreLog(ctx context.Context, logEntry *domain.EnergyConsumptionLog) error {
	sourceMixJSON, _ := json.Marshal(logEntry.SourceMixPercentage)

	query := `INSERT INTO energy_consumption (
		time, pool_id, duration_minutes, power_usage_kwh, peak_power_kw, average_power_kw,
		energy_source, source_mix_percentage, carbon_emissions_kg, grid_region_code,
		temperature_c, humidity_percent, data_quality, submitted_at, checksum
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, query,
		logEntry.Timestamp, logEntry.PoolID, logEntry.DurationMinutes, logEntry.PowerUsageKWh,
		logEntry.PeakPowerKW, logEntry.AveragePowerKW, logEntry.EnergySource, sourceMixJSON,
		logEntry.CarbonEmissionsKG, logEntry.GridRegionCode, logEntry.TemperatureC,
		logEntry.HumidityPercent, logEntry.DataQuality, logEntry.SubmittedAt, logEntry.Checksum,
	)

	if err != nil {
		return fmt.Errorf("failed to store energy log: %w", err)
	}

	return nil
}

// StoreBatch stores multiple energy consumption logs
func (r *TimescaleEnergyRepository) StoreBatch(ctx context.Context, logs []domain.EnergyConsumptionLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO energy_consumption (
		time, pool_id, duration_minutes, power_usage_kwh, peak_power_kw, average_power_kw,
		energy_source, source_mix_percentage, carbon_emissions_kg, grid_region_code,
		temperature_c, humidity_percent, data_quality, submitted_at, checksum
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, logEntry := range logs {
		sourceMixJSON, _ := json.Marshal(logEntry.SourceMixPercentage)

		_, err := stmt.ExecContext(ctx,
			logEntry.Timestamp, logEntry.PoolID, logEntry.DurationMinutes, logEntry.PowerUsageKWh,
			logEntry.PeakPowerKW, logEntry.AveragePowerKW, logEntry.EnergySource, sourceMixJSON,
			logEntry.CarbonEmissionsKG, logEntry.GridRegionCode, logEntry.TemperatureC,
			logEntry.HumidityPercent, logEntry.DataQuality, logEntry.SubmittedAt, logEntry.Checksum,
		)
		if err != nil {
			return fmt.Errorf("failed to insert energy log: %w", err)
		}
	}

	return tx.Commit()
}

// GetLatest retrieves the latest energy log for a pool
func (r *TimescaleEnergyRepository) GetLatest(ctx context.Context, poolID uuid.UUID) (*domain.EnergyConsumptionLog, error) {
	query := `SELECT time, pool_id, duration_minutes, power_usage_kwh, peak_power_kw, average_power_kw,
		energy_source, source_mix_percentage, carbon_emissions_kg, grid_region_code,
		temperature_c, humidity_percent, data_quality, submitted_at, checksum
		FROM energy_consumption WHERE pool_id = $1 ORDER BY time DESC LIMIT 1`

	logEntry := &domain.EnergyConsumptionLog{}
	var sourceMixJSON []byte

	err := r.db.QueryRowContext(ctx, query, poolID).Scan(
		&logEntry.Timestamp, &logEntry.PoolID, &logEntry.DurationMinutes, &logEntry.PowerUsageKWh,
		&logEntry.PeakPowerKW, &logEntry.AveragePowerKW, &logEntry.EnergySource, &sourceMixJSON,
		&logEntry.CarbonEmissionsKG, &logEntry.GridRegionCode, &logEntry.TemperatureC,
		&logEntry.HumidityPercent, &logEntry.DataQuality, &logEntry.SubmittedAt, &logEntry.Checksum,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest energy log: %w", err)
	}

	json.Unmarshal(sourceMixJSON, &logEntry.SourceMixPercentage)
	return logEntry, nil
}

// GetTimeSeries retrieves energy consumption logs for a time range
func (r *TimescaleEnergyRepository) GetTimeSeries(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) ([]domain.EnergyConsumptionLog, error) {
	query := `SELECT time, pool_id, duration_minutes, power_usage_kwh, peak_power_kw, average_power_kw,
		energy_source, source_mix_percentage, carbon_emissions_kg, grid_region_code,
		temperature_c, humidity_percent, data_quality, submitted_at, checksum
		FROM energy_consumption WHERE pool_id = $1 AND time >= $2 AND time <= $3 ORDER BY time`

	rows, err := r.db.QueryContext(ctx, query, poolID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get energy time series: %w", err)
	}
	defer rows.Close()

	var logs []domain.EnergyConsumptionLog
	for rows.Next() {
		logEntry := &domain.EnergyConsumptionLog{}
		var sourceMixJSON []byte

		err := rows.Scan(
			&logEntry.Timestamp, &logEntry.PoolID, &logEntry.DurationMinutes, &logEntry.PowerUsageKWh,
			&logEntry.PeakPowerKW, &logEntry.AveragePowerKW, &logEntry.EnergySource, &sourceMixJSON,
			&logEntry.CarbonEmissionsKG, &logEntry.GridRegionCode, &logEntry.TemperatureC,
			&logEntry.HumidityPercent, &logEntry.DataQuality, &logEntry.SubmittedAt, &logEntry.Checksum,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan energy log: %w", err)
		}

		json.Unmarshal(sourceMixJSON, &logEntry.SourceMixPercentage)
		logs = append(logs, *logEntry)
	}

	return logs, nil
}

// GetAggregated retrieves aggregated energy data
func (r *TimescaleEnergyRepository) GetAggregated(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time, interval string) (*EnergyAggregation, error) {
	// Simplified aggregation query
	query := `SELECT
		MIN(time) as period_start,
		MAX(time) as period_end,
		COALESCE(SUM(power_usage_kwh), 0) as total_kwh,
		COALESCE(MAX(peak_power_kw), 0) as peak_kw,
		COALESCE(AVG(average_power_kw), 0) as avg_kw,
		COALESCE(SUM(carbon_emissions_kg), 0) as total_carbon,
		COUNT(*) as record_count
		FROM energy_consumption WHERE pool_id = $1 AND time >= $2 AND time <= $3`

	agg := &EnergyAggregation{}
	err := r.db.QueryRowContext(ctx, query, poolID, startTime, endTime).Scan(
		&agg.PeriodStart, &agg.PeriodEnd, &agg.TotalKWh, &agg.PeakKW, &agg.AverageKW,
		&agg.TotalCarbonKG, &agg.RecordCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated energy data: %w", err)
	}

	agg.BySource = make(map[string]decimal.Decimal)
	return agg, nil
}

// Import for decimal
import "github.com/shopspring/decimal"
import "encoding/json"
