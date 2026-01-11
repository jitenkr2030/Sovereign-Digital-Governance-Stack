package service

import (
	"context"
	"fmt"
	"time"

	"defense-telemetry/internal/domain"
	"defense-telemetry/internal/repository"

	"github.com/shopspring/decimal"
)

// MaintenanceService handles maintenance operations and predictions
type MaintenanceService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisRepository
	config *Config
}

// NewMaintenanceService creates a new maintenance service
func NewMaintenanceService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *MaintenanceService {
	return &MaintenanceService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// CreateMaintenanceLog creates a new maintenance log
func (s *MaintenanceService) CreateMaintenanceLog(ctx context.Context, log *domain.MaintenanceLog) error {
	// Validate
	if log.UnitID == "" {
		return fmt.Errorf("unit_id is required")
	}
	if log.Component == "" {
		return fmt.Errorf("component is required")
	}

	// Set defaults
	if log.ID == "" {
		log.ID = fmt.Sprintf("maint-%d", time.Now().UnixNano())
	}
	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()

	if log.Status == "" {
		log.Status = domain.MaintenanceStatusScheduled
	}

	if err := s.repo.CreateMaintenanceLog(ctx, log); err != nil {
		return fmt.Errorf("failed to create maintenance log: %w", err)
	}

	// Invalidate cache
	_ = s.redis.CacheMaintenancePredictions(ctx, log.UnitID, nil, 0)

	return nil
}

// GetMaintenanceHistory retrieves maintenance history for a unit
func (s *MaintenanceService) GetMaintenanceHistory(ctx context.Context, unitID string, limit int) ([]*domain.MaintenanceLog, error) {
	// This would query the repository with limit
	// For now, return a placeholder
	return []*domain.MaintenanceLog{}, nil
}

// GetMaintenancePredictions retrieves maintenance predictions for a unit
func (s *MaintenanceService) GetMaintenancePredictions(ctx context.Context, unitID string) ([]*domain.MaintenancePrediction, error) {
	// Try cache first
	cached, err := s.redis.GetCachedMaintenancePredictions(ctx, unitID)
	if err == nil && cached != nil {
		var predictions []*domain.MaintenancePrediction
		if err := json.Unmarshal(cached, &predictions); err == nil {
			return predictions, nil
		}
	}

	// Fall back to database
	predictions, err := s.repo.GetMaintenancePredictions(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get predictions: %w", err)
	}

	// Cache the results
	ttl := 1 * time.Hour
	if err := s.redis.CacheMaintenancePredictions(ctx, unitID, predictions, ttl); err != nil {
		fmt.Printf("Warning: failed to cache predictions: %v\n", err)
	}

	return predictions, nil
}

// PredictMaintenanceNeeds analyzes telemetry to predict maintenance needs
func (s *MaintenanceService) PredictMaintenanceNeeds(ctx context.Context, unitID string) ([]*domain.MaintenancePrediction, error) {
	// Get latest telemetry
	telemetry, err := s.repo.GetLatestTelemetry(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get telemetry: %w", err)
	}
	if telemetry == nil {
		return nil, fmt.Errorf("no telemetry available for unit: %s", unitID)
	}

	var predictions []*domain.MaintenancePrediction

	// Predict engine maintenance
	if telemetry.EngineTemp.GreaterThan(decimal.NewFromInt(95)) {
		pred := s.createPrediction(unitID, "engine", telemetry.EngineTemp, "Engine temperature trending high")
		predictions = append(predictions, pred)
	}

	// Predict hull maintenance
	if telemetry.HullIntegrity.LessThan(decimal.NewFromFloat(s.config.Maintenance.CriticalThreshold)) {
		pred := s.createPrediction(unitID, "hull", telemetry.HullIntegrity, "Hull integrity critically low")
		predictions = append(predictions, pred)
	}

	// Predict based on component health
	if telemetry.ComponentHealth.Engine.GreaterThan(decimal.Zero) &&
		telemetry.ComponentHealth.Engine.LessThan(decimal.NewFromFloat(s.config.Maintenance.WarningThreshold)) {
		pred := s.createPrediction(unitID, "engine", telemetry.ComponentHealth.Engine, "Engine health degrading")
		predictions = append(predictions, pred)
	}

	return predictions, nil
}

// createPrediction creates a maintenance prediction
func (s *MaintenanceService) createPrediction(unitID, component string, health decimal.Decimal, reason string) *domain.MaintenancePrediction {
	// Calculate predicted failure date based on health trend
	daysUntilFailure := int(health.InexactFloat64())
	if daysUntilFailure < 1 {
		daysUntilFailure = 7
	}

	predictedDate := time.Now().AddDate(0, 0, daysUntilFailure)

	// Determine risk level
	var riskLevel string
	if health.LessThan(decimal.NewFromFloat(30)) {
		riskLevel = "critical"
	} else if health.LessThan(decimal.NewFromFloat(50)) {
		riskLevel = "high"
	} else if health.LessThan(decimal.NewFromFloat(70)) {
		riskLevel = "medium"
	} else {
		riskLevel = "low"
	}

	// Calculate confidence based on data quality
	confidence := health
	if confidence.GreaterThan(decimal.NewFromInt(100)) {
		confidence = decimal.NewFromInt(100)
	}

	return &domain.MaintenancePrediction{
		ID:                   fmt.Sprintf("pred-%s-%s-%d", unitID, component, time.Now().UnixNano()),
		UnitID:               unitID,
		Component:            component,
		CurrentHealth:        health,
		PredictedFailureDate: predictedDate,
		ConfidenceLevel:      confidence,
		RiskLevel:            riskLevel,
		Recommendations:      []string{reason},
		ModelVersion:         "v1.0",
		PredictionDate:       time.Now(),
	}
}

// ScheduleMaintenance schedules a maintenance task
func (s *MaintenanceService) ScheduleMaintenance(ctx context.Context, log *domain.MaintenanceLog) error {
	log.Status = domain.MaintenanceStatusScheduled
	log.ScheduledAt = time.Now()
	log.NextServiceDue = log.ScheduledAt.Add(30 * 24 * time.Hour) // Default to 30 days

	return s.CreateMaintenanceLog(ctx, log)
}

// CompleteMaintenance marks maintenance as completed
func (s *MaintenanceService) CompleteMaintenance(ctx context.Context, logID string, conditionScore decimal.Decimal) error {
	now := time.Now()
	log := &domain.MaintenanceLog{
		ID:             logID,
		Status:         domain.MaintenanceStatusCompleted,
		CompletedAt:    &now,
		ConditionScore: conditionScore,
	}

	// Update in database
	// In a real implementation, you'd fetch the existing log and update it

	return nil
}

// GetUpcomingMaintenance retrieves scheduled maintenance for the next period
func (s *MaintenanceService) GetUpcomingMaintenance(ctx context.Context, days int) ([]*domain.MaintenanceLog, error) {
	// This would query for scheduled maintenance in the next N days
	return []*domain.MaintenanceLog{}, nil
}

// AnalyzeEquipmentHealth performs a comprehensive health analysis
func (s *MaintenanceService) AnalyzeEquipmentHealth(ctx context.Context, unitID string) (*HealthAnalysis, error) {
	// Get cached health data
	healthData, err := s.redis.GetCachedEquipmentHealth(ctx, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health data: %w", err)
	}

	var health map[string]decimal.Decimal
	if healthData != nil {
		json.Unmarshal(healthData, &health)
	}

	analysis := &HealthAnalysis{
		UnitID:        unitID,
		AnalysisTime:  time.Now(),
		OverallHealth: decimal.NewFromInt(100),
		Components:    make(map[string]ComponentAnalysis),
	}

	if len(health) == 0 {
		analysis.OverallHealth = decimal.NewFromInt(0)
		return analysis, nil
	}

	var totalHealth decimal.Decimal
	var componentCount int

	for component, score := range health {
		compAnalysis := s.analyzeComponent(component, score)
		analysis.Components[component] = compAnalysis
		totalHealth = totalHealth.Add(score)
		componentCount++
	}

	if componentCount > 0 {
		analysis.OverallHealth = totalHealth.Div(decimal.NewFromInt(int64(componentCount)))
	}

	return analysis, nil
}

// HealthAnalysis represents a comprehensive health analysis
type HealthAnalysis struct {
	UnitID        string                    `json:"unit_id"`
	AnalysisTime  time.Time                 `json:"analysis_time"`
	OverallHealth decimal.Decimal           `json:"overall_health"`
	Components    map[string]ComponentAnalysis `json:"components"`
}

// ComponentAnalysis represents analysis of a single component
type ComponentAnalysis struct {
	HealthScore     decimal.Decimal `json:"health_score"`
	Status          string          `json:"status"`
	Recommendations []string        `json:"recommendations"`
}

// analyzeComponent analyzes a single component
func (s *MaintenanceService) analyzeComponent(component string, health decimal.Decimal) ComponentAnalysis {
	var status string
	var recommendations []string

	if health.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Maintenance.HealthyThreshold)) {
		status = "healthy"
	} else if health.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Maintenance.WarningThreshold)) {
		status = "warning"
		recommendations = append(recommendations, fmt.Sprintf("Schedule inspection for %s", component))
	} else if health.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Maintenance.CriticalThreshold)) {
		status = "critical"
		recommendations = append(recommendations, fmt.Sprintf("Immediate maintenance required for %s", component))
	} else {
		status = "failed"
		recommendations = append(recommendations, fmt.Sprintf("URGENT: %s needs replacement", component))
	}

	return ComponentAnalysis{
		HealthScore:     health,
		Status:          status,
		Recommendations: recommendations,
	}
}

// json unmarshal helper
func json.Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("json.Unmarshal not imported")
}
