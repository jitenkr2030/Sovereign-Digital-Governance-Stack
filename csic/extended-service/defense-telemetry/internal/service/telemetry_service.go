package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"defense-telemetry/internal/domain"
	"defense-telemetry/internal/repository"

	"github.com/shopspring/decimal"
)

// Config holds the service configuration
type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Kafka       KafkaConfig
	Telemetry   TelemetryConfig
	Equipment   EquipmentConfig
	Threats     ThreatsConfig
	Maintenance MaintenanceConfig
	Deployment  DeploymentConfig
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
	Brokers       map[string][]string
	ConsumerGroup string
	Topics        map[string]string
}

type TelemetryConfig struct {
	BatchSize         int
	BatchTimeout      time.Duration
	DedupWindow       time.Duration
	DedupCacheTTL     time.Duration
	MaxProcessingTime time.Duration
	WorkerCount       int
}

type EquipmentConfig struct {
	HullIntegrityWarning   float64
	HullIntegrityCritical  float64
	FuelLevelWarning       float64
	FuelLevelCritical      float64
	AmmunitionWarning      float64
	AmmunitionCritical     float64
	MaxVelocity            float64
	MinVelocity            float64
}

type ThreatsConfig struct {
	HighSeverityThreshold    int
	CriticalSeverityThreshold int
	AlertCooldown            time.Duration
	AutoEscalate            bool
	EscalationDelay         time.Duration
}

type MaintenanceConfig struct {
	HealthyThreshold   float64
	WarningThreshold   float64
	CriticalThreshold  float64
	PredictionHorizon  time.Duration
	TrendWindow        time.Duration
	Components         []string
}

type DeploymentConfig struct {
	Zones                 []ZoneConfig
	MaxDistanceFromBase   float64
	OptimalCoverageRadius float64
	OverlapTolerance      float64
}

type ZoneConfig struct {
	ID       string
	Name     string
	Priority int
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

// TelemetryService handles telemetry data processing
type TelemetryService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisRepository
	config *Config
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *TelemetryService {
	return &TelemetryService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// IngestTelemetry processes incoming telemetry data
func (s *TelemetryService) IngestTelemetry(ctx context.Context, data *domain.TelemetryData) (*TelemetryResult, error) {
	// Validate data
	if err := s.validateTelemetry(data); err != nil {
		return nil, fmt.Errorf("invalid telemetry: %w", err)
	}

	// Check for duplicate
	hash := s.generateTelemetryHash(data)
	isDuplicate, err := s.redis.CheckAndSetDedupKey(ctx, hash, s.config.Telemetry.DedupWindow)
	if err != nil {
		return nil, fmt.Errorf("dedup check failed: %w", err)
	}
	if isDuplicate {
		return &TelemetryResult{
			Status:   "duplicate",
			Data:     data,
			Message:  "Duplicate telemetry ignored",
		}, nil
	}

	// Set timestamps
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}
	data.ProcessedAt = time.Now()

	// Save to database
	if err := s.repo.CreateTelemetryData(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to save telemetry: %w", err)
	}

	// Update unit location cache
	if err := s.redis.CacheUnitLocation(ctx, data.UnitID,
		data.Coordinates.Latitude.InexactFloat64(),
		data.Coordinates.Longitude.InexactFloat64()); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to cache unit location: %v\n", err)
	}

	// Cache latest telemetry
	ttl := s.config.Telemetry.BatchTimeout * 2
	if err := s.redis.CacheLatestTelemetry(ctx, data.UnitID, data, ttl); err != nil {
		fmt.Printf("Warning: failed to cache telemetry: %v\n", err)
	}

	// Check for alerts
	alerts := s.checkForAlerts(data)

	// Update equipment health
	s.updateEquipmentHealth(ctx, data)

	// Increment metrics counter
	_ = s.redis.IncrementCounter(ctx, "telemetry_ingested")

	return &TelemetryResult{
		Status:   "processed",
		Data:     data,
		Alerts:   alerts,
		Message:  "Telemetry processed successfully",
	}, nil
}

// IngestBatch processes a batch of telemetry data
func (s *TelemetryService) IngestBatch(ctx context.Context, batch *domain.TelemetryBatch) (*BatchResult, error) {
	var results []*TelemetryResult
	var processed int
	var duplicates int
	var errors int

	for i := range batch.Readings {
		result, err := s.IngestTelemetry(ctx, &batch.Readings[i])
		if err != nil {
			errors++
			continue
		}

		results = append(results, result)
		switch result.Status {
		case "processed":
			processed++
		case "duplicate":
			duplicates++
		}
	}

	return &BatchResult{
		Total:      len(batch.Readings),
		Processed:  processed,
		Duplicates: duplicates,
		Errors:     errors,
		Results:    results,
	}, nil
}

// TelemetryResult represents the result of telemetry processing
type TelemetryResult struct {
	Status   string
	Data     *domain.TelemetryData
	Alerts   []*domain.Alert
	Message  string
}

// BatchResult represents the result of batch processing
type BatchResult struct {
	Total      int
	Processed  int
	Duplicates int
	Errors     int
	Results    []*TelemetryResult
}

// validateTelemetry validates telemetry data
func (s *TelemetryService) validateTelemetry(data *domain.TelemetryData) error {
	if data.UnitID == "" {
		return fmt.Errorf("unit_id is required")
	}
	if data.Velocity.IsNegative() {
		return fmt.Errorf("velocity cannot be negative")
	}
	if data.FuelLevel.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("fuel_level cannot exceed 100%")
	}
	if data.HullIntegrity.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("hull_integrity cannot exceed 100%")
	}
	return nil
}

// generateTelemetryHash generates a hash for deduplication
func (s *TelemetryService) generateTelemetryHash(data *domain.TelemetryData) string {
	hashInput := fmt.Sprintf("%s:%s:%s",
		data.UnitID,
		data.Timestamp.Format(time.RFC3339Nano),
		data.Coordinates.Latitude.String(),
	)

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// checkForAlerts checks telemetry data for conditions that trigger alerts
func (s *TelemetryService) checkForAlerts(data *domain.TelemetryData) []*domain.Alert {
	var alerts []*domain.Alert

	// Check hull integrity
	if data.HullIntegrity.LessThan(decimal.NewFromFloat(s.config.Equipment.HullIntegrityCritical)) {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeMaintenance,
			Severity:  domain.AlertSeverityCritical,
			Title:     fmt.Sprintf("Critical Hull Integrity: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s has critical hull integrity at %.1f%%", data.UnitID, data.HullIntegrity),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	} else if data.HullIntegrity.LessThan(decimal.NewFromFloat(s.config.Equipment.HullIntegrityWarning)) {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeMaintenance,
			Severity:  domain.AlertSeverityWarning,
			Title:     fmt.Sprintf("Hull Integrity Warning: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s hull integrity at %.1f%%", data.UnitID, data.HullIntegrity),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	}

	// Check fuel level
	if data.FuelLevel.LessThan(decimal.NewFromFloat(s.config.Equipment.FuelLevelCritical)) {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeResource,
			Severity:  domain.AlertSeverityCritical,
			Title:     fmt.Sprintf("Critical Fuel Level: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s fuel at critical level: %.1f%%", data.UnitID, data.FuelLevel),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	} else if data.FuelLevel.LessThan(decimal.NewFromFloat(s.config.Equipment.FuelLevelWarning)) {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeResource,
			Severity:  domain.AlertSeverityWarning,
			Title:     fmt.Sprintf("Low Fuel Warning: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s fuel level at %.1f%%", data.UnitID, data.FuelLevel),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	}

	// Check ammunition
	if float64(data.AmmunitionCount) < s.config.Equipment.AmmunitionCritical {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeResource,
			Severity:  domain.AlertSeverityCritical,
			Title:     fmt.Sprintf("Critical Ammunition: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s ammunition critically low: %d", data.UnitID, data.AmmunitionCount),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	} else if float64(data.AmmunitionCount) < s.config.Equipment.AmmunitionWarning {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeResource,
			Severity:  domain.AlertSeverityWarning,
			Title:     fmt.Sprintf("Ammunition Warning: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s ammunition running low: %d", data.UnitID, data.AmmunitionCount),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	}

	// Check engine temperature
	if data.EngineTemp.GreaterThan(decimal.NewFromInt(100)) {
		alerts = append(alerts, &domain.Alert{
			ID:        "",
			AlertType: domain.AlertTypeMaintenance,
			Severity:  domain.AlertSeverityWarning,
			Title:     fmt.Sprintf("High Engine Temperature: %s", data.UnitID),
			Message:   fmt.Sprintf("Unit %s engine temperature high: %.1f°C", data.UnitID, data.EngineTemp),
			Source:    "telemetry_monitor",
			EntityID:  data.UnitID,
			EntityType: "unit",
		})
	}

	// Save alerts to database
	for _, alert := range alerts {
		if err := s.repo.CreateAlert(ctx, alert); err != nil {
			fmt.Printf("Warning: failed to save alert: %v\n", err)
		}
	}

	return alerts
}

// updateEquipmentHealth updates the cached equipment health scores
func (s *TelemetryService) updateEquipmentHealth(ctx context.Context, data *domain.TelemetryData) {
	health := map[string]decimal.Decimal{
		"hull":           data.HullIntegrity,
		"engine":         s.calculateEngineHealth(data.EngineTemp),
		"power_systems":  data.BatteryLevel,
		"communications": s.calculateSignalHealth(data.SignalStrength),
	}

	if err := s.redis.CacheEquipmentHealth(ctx, data.UnitID, health, 5*time.Minute); err != nil {
		fmt.Printf("Warning: failed to cache equipment health: %v\n", err)
	}
}

// calculateEngineHealth calculates engine health from temperature
func (s *TelemetryService) calculateEngineHealth(temp decimal.Decimal) decimal.Decimal {
	// Normal temperature range: 60-90°C
	optimal := decimal.NewFromInt(80)
	if temp.LessThanOrEqual(optimal) {
		return decimal.NewFromInt(100)
	}

	// Degrade health as temperature increases above optimal
	excess := temp.Sub(optimal)
	health := decimal.NewFromInt(100).Sub(excess.Mul(decimal.NewFromInt(2)))
	if health.LessThan(decimal.NewFromInt(0)) {
		return decimal.NewFromInt(0)
	}
	return health
}

// calculateSignalHealth calculates signal health from signal strength
func (s *TelemetryService) calculateSignalHealth(signal decimal.Decimal) decimal.Decimal {
	// Signal strength in dBm, typical range: -50 (good) to -120 (poor)
	if signal.GreaterThan(decimal.NewFromInt(-50)) {
		return decimal.NewFromInt(100)
	}
	if signal.LessThan(decimal.NewFromInt(-120)) {
		return decimal.NewFromInt(0)
	}

	// Map -50 to -120 range to 100-0
	rangeSize := decimal.NewFromInt(70)
	signalOffset := signal.Sub(decimal.NewFromInt(-50)).Abs()
	health := decimal.NewFromInt(100).Sub(signalOffset.Mul(decimal.NewFromInt(100)).Div(rangeSize))
	return health
}

// GetLatestTelemetry retrieves the latest telemetry for a unit
func (s *TelemetryService) GetLatestTelemetry(ctx context.Context, unitID string) (*domain.TelemetryData, error) {
	// Try cache first
	cached, err := s.redis.GetCachedLatestTelemetry(ctx, unitID)
	if err == nil && cached != nil {
		var data domain.TelemetryData
		if err := json.Unmarshal(cached, &data); err == nil {
			return &data, nil
		}
	}

	// Fall back to database
	data, err := s.repo.GetLatestTelemetry(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry: %w", err)
	}

	// Cache the result
	if data != nil {
		ttl := s.config.Telemetry.BatchTimeout * 2
		_ = s.redis.CacheLatestTelemetry(ctx, unitID, data, ttl)
	}

	return data, nil
}

// GetNearbyUnits returns units within a radius of a location
func (s *TelemetryService) GetNearbyUnits(ctx context.Context, lat, lng, radius float64) ([]string, error) {
	return s.redis.GetNearbyUnits(ctx, lat, lng, radius)
}

// GetUnitStatus retrieves the status of a unit
func (s *TelemetryService) GetUnitStatus(ctx context.Context, unitID string) (*UnitStatus, error) {
	// Get unit info
	unit, err := s.repo.GetDefenseUnit(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}
	if unit == nil {
		return nil, nil
	}

	// Get latest telemetry
	telemetry, err := s.GetLatestTelemetry(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry: %w", err)
	}

	// Get equipment health
	healthData, _ := s.redis.GetCachedEquipmentHealth(ctx, unitID)
	var health map[string]decimal.Decimal
	if healthData != nil {
		json.Unmarshal(healthData, &health)
	}

	status := &UnitStatus{
		UnitID:       unitID,
		UnitCode:     unit.UnitCode,
		UnitType:     string(unit.UnitType),
		UnitStatus:   string(unit.Status),
		LastContact:  unit.LastContactAt,
		HealthScore:  s.calculateOverallHealth(health),
		IsDeployed:   unit.IsDeployed,
		DeploymentZone: unit.DeploymentZone,
	}

	if telemetry != nil {
		status.Location = &GeoLocation{
			Latitude:  telemetry.Coordinates.Latitude.String(),
			Longitude: telemetry.Coordinates.Longitude.String(),
			Altitude:  telemetry.Coordinates.Altitude.String(),
		}
		status.FuelLevel = telemetry.FuelLevel.String()
		status.HullIntegrity = telemetry.HullIntegrity.String()
		status.Velocity = telemetry.Velocity.String()
	}

	return status, nil
}

// UnitStatus represents the current status of a unit
type UnitStatus struct {
	UnitID         string       `json:"unit_id"`
	UnitCode       string       `json:"unit_code"`
	UnitType       string       `json:"unit_type"`
	UnitStatus     string       `json:"unit_status"`
	LastContact    time.Time    `json:"last_contact"`
	Location       *GeoLocation `json:"location,omitempty"`
	FuelLevel      string       `json:"fuel_level"`
	HullIntegrity  string       `json:"hull_integrity"`
	Velocity       string       `json:"velocity"`
	HealthScore    decimal.Decimal `json:"health_score"`
	IsDeployed     bool         `json:"is_deployed"`
	DeploymentZone string       `json:"deployment_zone,omitempty"`
}

// GeoLocation represents a geographic location
type GeoLocation struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Altitude  string `json:"altitude,omitempty"`
}

// calculateOverallHealth calculates overall health score from component health
func (s *TelemetryService) calculateOverallHealth(health map[string]decimal.Decimal) decimal.Decimal {
	if len(health) == 0 {
		return decimal.NewFromInt(100)
	}

	var total decimal.Decimal
	var count int
	for _, score := range health {
		total = total.Add(score)
		count++
	}

	if count == 0 {
		return decimal.NewFromInt(100)
	}

	return total.Div(decimal.NewFromInt(int64(count)))
}

// GetAllUnits retrieves all defense units
func (s *TelemetryService) GetAllUnits(ctx context.Context) ([]*domain.DefenseUnit, error) {
	return s.repo.GetAllDefenseUnits(ctx)
}

// CreateDefenseUnit creates a new defense unit
func (s *TelemetryService) CreateDefenseUnit(ctx context.Context, unit *domain.DefenseUnit) error {
	unit.CreatedAt = time.Now()
	unit.UpdatedAt = time.Now()
	return s.repo.CreateDefenseUnit(ctx, unit)
}

// GetDefenseUnit retrieves a defense unit by ID
func (s *TelemetryService) GetDefenseUnit(ctx context.Context, id string) (*domain.DefenseUnit, error) {
	return s.repo.GetDefenseUnit(ctx, id)
}

// GetDefenseUnitByCode retrieves a defense unit by code
func (s *TelemetryService) GetDefenseUnitByCode(ctx context.Context, code string) (*domain.DefenseUnit, error) {
	return s.repo.GetDefenseUnitByCode(ctx, code)
}

// UpdateDefenseUnit updates a defense unit
func (s *TelemetryService) UpdateDefenseUnit(ctx context.Context, unit *domain.DefenseUnit) error {
	return s.repo.UpdateDefenseUnit(ctx, unit)
}
