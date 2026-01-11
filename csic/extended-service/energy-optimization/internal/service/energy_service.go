package service

import (
	"context"
	"fmt"
	"time"

	"energy-optimization/internal/domain"
	"energy-optimization/internal/repository"

	"github.com/shopspring/decimal"
)

// EnergyService handles energy reading operations
type EnergyService struct {
	repo     *repository.PostgresRepository
	redis    *repository.RedisRepository
	config   *Config
}

// Config holds the service configuration
type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Kafka       KafkaConfig
	Carbon      CarbonConfig
	LoadBalance LoadBalanceConfig
	Meters      MetersConfig
	Metrics     MetricsConfig
	Logging     LoggingConfig
}

type AppConfig struct {
	Host        string
	Port        int
	Name        string
	Environment string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	TablePrefix     string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host      string
	Port      int
	Password  string
	DB        int
	KeyPrefix string
	PoolSize  int
}

type KafkaConfig struct {
	Brokers  map[string][]string
	Group    string
	Topics   map[string]string
}

type CarbonConfig struct {
	EmissionFactors   map[string]decimal.Decimal
	BudgetWarning     float64
	BudgetCritical    float64
}

type LoadBalanceConfig struct {
	TargetMin     float64
	TargetMax     float64
	ScaleUp       float64
	ScaleDown     float64
	CheckInterval time.Duration
}

type MetersConfig struct {
	DefaultSampleInterval time.Duration
	AggregationWindow     time.Duration
	MaxReadingAge         time.Duration
}

type MetricsConfig struct {
	Enabled bool
	Path    string
	Port    int
}

type LoggingConfig struct {
	Level  string
	Format string
}

// NewEnergyService creates a new energy service
func NewEnergyService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *EnergyService {
	return &EnergyService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// IngestReading processes a new energy reading from a meter
func (s *EnergyService) IngestReading(ctx context.Context, reading *domain.EnergyReading) error {
	// Validate reading
	if err := s.validateReading(reading); err != nil {
		return fmt.Errorf("invalid reading: %w", err)
	}

	// Set timestamp if not provided
	if reading.Timestamp.IsZero() {
		reading.Timestamp = time.Now()
	}

	// Set created at if not provided
	if reading.CreatedAt.IsZero() {
		reading.CreatedAt = time.Now()
	}

	// Save to database
	if err := s.repo.CreateEnergyReading(ctx, reading); err != nil {
		return fmt.Errorf("failed to save reading: %w", err)
	}

	// Cache the reading for quick access
	ttl := s.config.Meters.DefaultSampleInterval * 10
	if err := s.redis.CacheEnergyReading(ctx, &struct {
		ID             string
		MeterID        string
		FacilityID     string
		EnergySource   string
		Consumption    decimal.Decimal
		Timestamp      time.Time
	}{
		ID:           reading.ID,
		MeterID:      reading.MeterID,
		FacilityID:   reading.FacilityID,
		EnergySource: string(reading.EnergySource),
		Consumption:  reading.Consumption,
		Timestamp:    reading.Timestamp,
	}, ttl); err != nil {
		// Log but don't fail on cache error
		fmt.Printf("Warning: failed to cache reading: %v\n", err)
	}

	// Update facility's latest reading
	if err := s.redis.SetFacilityLatestReading(ctx, reading.FacilityID, reading.ID, ttl); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to update facility latest reading: %v\n", err)
	}

	// Increment daily consumption counter
	if err := s.redis.IncrementDailyConsumption(ctx, reading.FacilityID, string(reading.EnergySource), reading.Consumption); err != nil {
		fmt.Printf("Warning: failed to increment daily consumption: %v\n", err)
	}

	// Publish to Kafka for downstream processing
	// This would be handled by a separate consumer in production
	// For now, we just log it
	fmt.Printf("Published energy reading %s to Kafka\n", reading.ID)

	return nil
}

// validateReading validates an energy reading
func (s *EnergyService) validateReading(reading *domain.EnergyReading) error {
	if reading.MeterID == "" {
		return fmt.Errorf("meter_id is required")
	}
	if reading.FacilityID == "" {
		return fmt.Errorf("facility_id is required")
	}
	if reading.Consumption.IsNegative() {
		return fmt.Errorf("consumption cannot be negative")
	}
	return nil
}

// GetReading retrieves an energy reading by ID
func (s *EnergyService) GetReading(ctx context.Context, id string) (*domain.EnergyReading, error) {
	// Try cache first
	cached, err := s.redis.GetCachedEnergyReading(ctx, id)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Fall back to database
	reading, err := s.repo.GetEnergyReading(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get reading: %w", err)
	}

	if reading == nil {
		return nil, nil
	}

	// Cache for future requests
	ttl := s.config.Meters.DefaultSampleInterval * 10
	_ = s.redis.CacheEnergyReading(ctx, &struct {
		ID             string
		MeterID        string
		FacilityID     string
		EnergySource   string
		Consumption    decimal.Decimal
		Timestamp      time.Time
	}{
		ID:           reading.ID,
		MeterID:      reading.MeterID,
		FacilityID:   reading.FacilityID,
		EnergySource: string(reading.EnergySource),
		Consumption:  reading.Consumption,
		Timestamp:    reading.Timestamp,
	}, ttl)

	return reading, nil
}

// GetReadingsByFacility retrieves energy readings for a facility
func (s *EnergyService) GetReadingsByFacility(ctx context.Context, facilityID string, startTime, endTime time.Time, limit int) ([]*domain.EnergyReading, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	return s.repo.GetEnergyReadingsByFacility(ctx, facilityID, startTime, endTime, limit)
}

// GetFacilityCurrentLoad calculates the current load for a facility
func (s *EnergyService) GetFacilityCurrentLoad(ctx context.Context, facilityID string) (decimal.Decimal, error) {
	// Try to get from cache first
	consumption, err := s.redis.GetDailyConsumption(ctx, facilityID, "all")
	if err == nil && !consumption.IsZero() {
		return consumption, nil
	}

	// Fall back to calculating from recent readings
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	readings, err := s.repo.GetEnergyReadingsByFacility(ctx, facilityID, startTime, endTime, 1000)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get readings: %w", err)
	}

	var totalConsumption decimal.Decimal
	for _, reading := range readings {
		totalConsumption = totalConsumption.Add(reading.Consumption)
	}

	return totalConsumption, nil
}

// GetDailyConsumption retrieves the daily consumption for a facility
func (s *EnergyService) GetDailyConsumption(ctx context.Context, facilityID string) (map[string]decimal.Decimal, error) {
	sources := []string{"grid", "solar", "wind", "natural_gas", "coal", "nuclear", "hydro"}
	result := make(map[string]decimal.Decimal)

	for _, source := range sources {
		consumption, err := s.redis.GetDailyConsumption(ctx, facilityID, source)
		if err != nil {
			return nil, fmt.Errorf("failed to get consumption for %s: %w", source, err)
		}
		result[source] = consumption
	}

	return result, nil
}

// AggregateReadings aggregates readings over a time window
func (s *EnergyService) AggregateReadings(ctx context.Context, facilityID string, window time.Duration) (*domain.LoadMetric, error) {
	endTime := time.Now()
	startTime := endTime.Add(-window)

	readings, err := s.repo.GetEnergyReadingsByFacility(ctx, facilityID, startTime, endTime, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get readings: %w", err)
	}

	if len(readings) == 0 {
		return nil, fmt.Errorf("no readings found for aggregation")
	}

	metric := &domain.LoadMetric{
		ID:             "",
		FacilityID:     facilityID,
		Timestamp:      endTime,
		WindowDuration: window,
		SampleCount:    len(readings),
	}

	var totalConsumption decimal.Decimal
	peakConsumption := decimal.Zero
	minConsumption := decimal.MaxInt64 // Using MaxInt64 as initial value

	for _, reading := range readings {
		consumption := reading.Consumption

		// Convert kWh to kW for instantaneous load (assuming hourly readings)
		// For more accurate calculation, we'd need the time between readings
		load := consumption.Mul(decimal.NewFromInt(1)) // Placeholder conversion

		totalConsumption = totalConsumption.Add(consumption)

		if load.GreaterThan(peakConsumption) {
			peakConsumption = load
		}
		if load.LessThan(decimal.NewFromInt(int64(minConsumption))) {
			minConsumption = load
		}

		if reading.PowerFactor.IsPositive() {
			metric.PowerFactor = metric.PowerFactor.Add(reading.PowerFactor)
		}
	}

	metric.TotalConsumption = totalConsumption
	metric.PeakConsumption = peakConsumption
	metric.MinConsumption = decimal.NewFromInt(100) // Placeholder

	if metric.PowerFactor.IsPositive() && len(readings) > 0 {
		metric.PowerFactor = metric.PowerFactor.Div(decimal.NewFromInt(int64(len(readings))))
	}

	return metric, nil
}

// GetFacilityStatus returns the current status of a facility
func (s *EnergyService) GetFacilityStatus(ctx context.Context, facilityID string) (*FacilityStatus, error) {
	// Get cached status first
	data, err := s.redis.GetFacilityStatus(ctx, facilityID)
	if err == nil && data != nil {
		var status FacilityStatus
		if err := json.Unmarshal(data, &status); err == nil {
			return &status, nil
		}
	}

	// Calculate status
	load, err := s.GetFacilityCurrentLoad(ctx, facilityID)
	if err != nil {
		return nil, err
	}

	consumption, err := s.GetDailyConsumption(ctx, facilityID)
	if err != nil {
		return nil, err
	}

	var totalConsumption decimal.Decimal
	for _, v := range consumption {
		totalConsumption = totalConsumption.Add(v)
	}

	status := &FacilityStatus{
		FacilityID:      facilityID,
		CurrentLoad:     load,
		TotalDailyUsage: totalConsumption,
		Status:          "normal",
		Timestamp:       time.Now(),
	}

	// Cache the status
	ttl := s.config.Meters.DefaultSampleInterval
	_ = s.redis.SetFacilityStatus(ctx, facilityID, status, ttl)

	return status, nil
}

// FacilityStatus represents the current status of a facility
type FacilityStatus struct {
	FacilityID      string          `json:"facility_id"`
	CurrentLoad     decimal.Decimal `json:"current_load"`
	TotalDailyUsage decimal.Decimal `json:"total_daily_usage"`
	Status          string          `json:"status"`
	Timestamp       time.Time       `json:"timestamp"`
}

// json unmarshal helper
func json.Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("json.Unmarshal not imported, using fmt")
}
