package service

import (
	"context"
	"fmt"
	"time"

	"energy-optimization/internal/domain"
	"energy-optimization/internal/repository"

	"github.com/shopspring/decimal"
)

// CarbonService handles carbon emission tracking and budget management
type CarbonService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisRepository
	config *Config
}

// NewCarbonService creates a new carbon service
func NewCarbonService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *CarbonService {
	return &CarbonService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// CalculateEmission calculates carbon emissions for an energy reading
func (s *CarbonService) CalculateEmission(reading *domain.EnergyReading) *domain.CarbonEmission {
	// Get emission factor for the energy source
	emissionFactor := s.getEmissionFactor(reading.EnergySource)

	// Calculate total emission (kg CO2 = kWh * kg CO2/kWh)
	totalEmission := reading.Consumption.Mul(emissionFactor)

	// Carbon intensity in g CO2 per kWh
	carbonIntensity := emissionFactor.Mul(decimal.NewFromInt(1000))

	return &domain.CarbonEmission{
		ID:               "",
		ReadingID:        reading.ID,
		FacilityID:       reading.FacilityID,
		EnergySource:     reading.EnergySource,
		EmissionFactor:   emissionFactor,
		Consumption:      reading.Consumption,
		TotalEmission:    totalEmission,
		CarbonIntensity:  carbonIntensity,
		Timestamp:        reading.Timestamp,
		CalculatedAt:     time.Now(),
	}
}

// getEmissionFactor returns the emission factor for an energy source
func (s *CarbonService) getEmissionFactor(source domain.EnergySource) decimal.Decimal {
	factors := map[domain.EnergySource]decimal.Decimal{
		domain.EnergySourceGrid:      decimal.NewFromFloat(0.42),
		domain.EnergySourceSolar:     decimal.NewFromFloat(0.041),
		domain.EnergySourceWind:      decimal.NewFromFloat(0.011),
		domain.EnergySourceNaturalGas: decimal.NewFromFloat(0.45),
		domain.EnergySourceCoal:      decimal.NewFromFloat(0.95),
		domain.EnergySourceNuclear:   decimal.NewFromFloat(0.012),
		domain.EnergySourceHydro:     decimal.NewFromFloat(0.024),
		domain.EnergySourceBattery:   decimal.NewFromFloat(0.0), // Batteries are considered carbon-neutral for charging
	}

	if factor, ok := factors[source]; ok {
		return factor
	}

	// Default to grid average if unknown
	return decimal.NewFromFloat(0.42)
}

// ProcessReading processes an energy reading and calculates carbon emissions
func (s *CarbonService) ProcessReading(ctx context.Context, reading *domain.EnergyReading) (*domain.CarbonEmission, error) {
	// Calculate emission
	emission := s.CalculateEmission(reading)

	// Save to database
	if err := s.repo.CreateCarbonEmission(ctx, emission); err != nil {
		return nil, fmt.Errorf("failed to save emission: %w", err)
	}

	// Update budget usage
	if err := s.updateBudgetUsage(ctx, reading.FacilityID, emission.TotalEmission); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to update budget usage: %v\n", err)
	}

	// Check if we're approaching budget limits
	status, err := s.CheckBudgetStatus(ctx, reading.FacilityID)
	if err != nil {
		fmt.Printf("Warning: failed to check budget status: %v\n", err)
	} else if status == domain.BudgetStatusWarning {
		fmt.Printf("Warning: Facility %s is approaching carbon budget limit\n", reading.FacilityID)
	} else if status == domain.BudgetStatusCritical {
		fmt.Printf("CRITICAL: Facility %s has exceeded carbon budget limit\n", reading.FacilityID)
	}

	return emission, nil
}

// updateBudgetUsage updates the carbon budget usage for a facility
func (s *CarbonService) updateBudgetUsage(ctx context.Context, facilityID string, additionalEmission decimal.Decimal) error {
	// Update Redis counter
	newUsage, err := s.redis.UpdateCarbonBudgetUsage(ctx, facilityID, additionalEmission)
	if err != nil {
		return fmt.Errorf("failed to update Redis counter: %w", err)
	}

	// Get budget to check thresholds
	budget, err := s.repo.GetCarbonBudget(ctx, facilityID)
	if err != nil {
		return fmt.Errorf("failed to get budget: %w", err)
	}

	if budget == nil {
		// No budget configured, skip
		return nil
	}

	// Update budget status based on new usage
	percentage := newUsage.Div(budget.BudgetAmount).Mul(decimal.NewFromInt(100))

	var status domain.BudgetStatus
	if percentage.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Carbon.BudgetCritical)) {
		status = domain.BudgetStatusCritical
	} else if percentage.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Carbon.BudgetWarning)) {
		status = domain.BudgetStatusWarning
	} else {
		status = domain.BudgetStatusNormal
	}

	// Update budget in database
	budget.UsedAmount = newUsage
	budget.RemainingAmount = budget.BudgetAmount.Sub(newUsage)
	budget.UsagePercentage = percentage
	budget.Status = status
	budget.UpdatedAt = time.Now()

	if err := s.repo.CreateOrUpdateCarbonBudget(ctx, budget); err != nil {
		return fmt.Errorf("failed to update budget: %w", err)
	}

	// Update cache
	ttl := 5 * time.Minute
	return s.redis.CacheCarbonBudget(ctx, facilityID, budget, ttl)
}

// GetCarbonEmissions retrieves carbon emissions for a facility
func (s *CarbonService) GetCarbonEmissions(ctx context.Context, facilityID string, startTime, endTime time.Time) ([]*domain.CarbonEmission, error) {
	return s.repo.GetCarbonEmissionsByFacility(ctx, facilityID, startTime, endTime)
}

// GetTotalEmissions calculates total emissions for a facility over a period
func (s *CarbonService) GetTotalEmissions(ctx context.Context, facilityID string, startTime, endTime time.Time) (decimal.Decimal, error) {
	emissions, err := s.repo.GetCarbonEmissionsByFacility(ctx, facilityID, startTime, endTime)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get emissions: %w", err)
	}

	var total decimal.Decimal
	for _, e := range emissions {
		total = total.Add(e.TotalEmission)
	}

	return total, nil
}

// CheckBudgetStatus checks the current budget status for a facility
func (s *CarbonService) CheckBudgetStatus(ctx context.Context, facilityID string) (domain.BudgetStatus, error) {
	// Get usage from Redis (faster)
	usage, err := s.redis.GetCarbonBudgetUsage(ctx, facilityID)
	if err != nil {
		return domain.BudgetStatusNormal, fmt.Errorf("failed to get usage: %w", err)
	}

	// Get budget from database
	budget, err := s.repo.GetCarbonBudget(ctx, facilityID)
	if err != nil {
		return domain.BudgetStatusNormal, fmt.Errorf("failed to get budget: %w", err)
	}

	if budget == nil {
		return domain.BudgetStatusNormal, nil
	}

	if usage.IsZero() {
		return budget.Status, nil
	}

	percentage := usage.Div(budget.BudgetAmount).Mul(decimal.NewFromInt(100))

	if percentage.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Carbon.BudgetCritical)) {
		return domain.BudgetStatusCritical, nil
	}
	if percentage.GreaterThanOrEqual(decimal.NewFromFloat(s.config.Carbon.BudgetWarning)) {
		return domain.BudgetStatusWarning, nil
	}

	return domain.BudgetStatusNormal, nil
}

// GetBudget retrieves the current carbon budget for a facility
func (s *CarbonService) GetBudget(ctx context.Context, facilityID string) (*domain.CarbonBudget, error) {
	// Try cache first
	data, err := s.redis.GetCachedCarbonBudget(ctx, facilityID)
	if err == nil && data != nil {
		var budget domain.CarbonBudget
		if err := json.Unmarshal(data, &budget); err == nil {
			return &budget, nil
		}
	}

	// Fall back to database
	budget, err := s.repo.GetCarbonBudget(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	if budget == nil {
		return nil, nil
	}

	// Update usage from Redis
	usage, err := s.redis.GetCarbonBudgetUsage(ctx, facilityID)
	if err == nil && !usage.IsZero() {
		budget.UsedAmount = usage
		budget.RemainingAmount = budget.BudgetAmount.Sub(usage)
		budget.UsagePercentage = usage.Div(budget.BudgetAmount).Mul(decimal.NewFromInt(100))
	}

	return budget, nil
}

// CreateBudget creates a new carbon budget for a facility
func (s *CarbonService) CreateBudget(ctx context.Context, budget *domain.CarbonBudget) error {
	if budget.ID == "" {
		budget.ID = fmt.Sprintf("budget-%s-%s", budget.FacilityID, time.Now().Format("2006-01-02"))
	}

	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()
	budget.UsedAmount = decimal.Zero
	budget.RemainingAmount = budget.BudgetAmount
	budget.UsagePercentage = decimal.Zero
	budget.Status = domain.BudgetStatusNormal

	if budget.WarningThreshold.IsZero() {
		budget.WarningThreshold = decimal.NewFromFloat(s.config.Carbon.BudgetWarning)
	}
	if budget.CriticalThreshold.IsZero() {
		budget.CriticalThreshold = decimal.NewFromFloat(s.config.Carbon.BudgetCritical)
	}

	if err := s.repo.CreateOrUpdateCarbonBudget(ctx, budget); err != nil {
		return fmt.Errorf("failed to create budget: %w", err)
	}

	// Reset usage counter for new budget period
	if err := s.redis.ResetCarbonBudgetUsage(ctx, budget.FacilityID); err != nil {
		fmt.Printf("Warning: failed to reset budget usage: %v\n", err)
	}

	// Cache the budget
	ttl := 5 * time.Minute
	return s.redis.CacheCarbonBudget(ctx, budget.FacilityID, budget, ttl)
}

// ResetBudgetUsage resets the budget usage for a new period
func (s *CarbonService) ResetBudgetUsage(ctx context.Context, facilityID string) error {
	if err := s.redis.ResetCarbonBudgetUsage(ctx, facilityID); err != nil {
		return fmt.Errorf("failed to reset usage: %w", err)
	}

	budget, err := s.repo.GetCarbonBudget(ctx, facilityID)
	if err != nil {
		return fmt.Errorf("failed to get budget: %w", err)
	}

	if budget != nil {
		budget.UsedAmount = decimal.Zero
		budget.RemainingAmount = budget.BudgetAmount
		budget.UsagePercentage = decimal.Zero
		budget.Status = domain.BudgetStatusNormal
		budget.UpdatedAt = time.Now()

		if err := s.repo.CreateOrUpdateCarbonBudget(ctx, budget); err != nil {
			return fmt.Errorf("failed to update budget: %w", err)
		}
	}

	return nil
}

// GetEmissionsBySource returns emissions breakdown by energy source
func (s *CarbonService) GetEmissionsBySource(ctx context.Context, facilityID string, startTime, endTime time.Time) (map[string]decimal.Decimal, error) {
	emissions, err := s.repo.GetCarbonEmissionsByFacility(ctx, facilityID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get emissions: %w", err)
	}

	result := make(map[string]decimal.Decimal)
	for _, e := range emissions {
		source := string(e.EnergySource)
		result[source] = result[source].Add(e.TotalEmission)
	}

	return result, nil
}

// GetCarbonIntensity returns the average carbon intensity for a facility
func (s *CarbonService) GetCarbonIntensity(ctx context.Context, facilityID string, startTime, endTime time.Time) (decimal.Decimal, error) {
	emissions, err := s.repo.GetCarbonEmissionsByFacility(ctx, facilityID, startTime, endTime)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get emissions: %w", err)
	}

	if len(emissions) == 0 {
		return decimal.Zero, nil
	}

	var totalEmission decimal.Decimal
	var totalConsumption decimal.Decimal

	for _, e := range emissions {
		totalEmission = totalEmission.Add(e.TotalEmission)
		totalConsumption = totalConsumption.Add(e.Consumption)
	}

	if totalConsumption.IsZero() {
		return decimal.Zero, nil
	}

	// Carbon intensity in g CO2 per kWh
	return totalEmission.Div(totalConsumption).Mul(decimal.NewFromInt(1000)), nil
}

// json unmarshal helper
func json.Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("json.Unmarshal not imported")
}
