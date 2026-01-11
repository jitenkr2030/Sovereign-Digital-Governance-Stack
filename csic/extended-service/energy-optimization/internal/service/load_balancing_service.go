package service

import (
	"context"
	"fmt"
	"time"

	"energy-optimization/internal/domain"
	"energy-optimization/internal/repository"

	"github.com/shopspring/decimal"
)

// LoadBalancingService handles load balancing and optimization decisions
type LoadBalancingService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisRepository
	config *Config
}

// NewLoadBalancingService creates a new load balancing service
func NewLoadBalancingService(repo *repository.PostgresRepository, redis *repository.RedisRepository, config *Config) *LoadBalancingService {
	return &LoadBalancingService{
		repo:   repo,
		redis:  redis,
		config: config,
	}
}

// RecordLoadMetric records a new load metric for a facility or zone
func (s *LoadBalancingService) RecordLoadMetric(ctx context.Context, metric *domain.LoadMetric) error {
	if metric.ID == "" {
		metric.ID = fmt.Sprintf("lm-%d", time.Now().UnixNano())
	}

	// Calculate utilization
	if metric.TotalConsumption.IsPositive() && metric.AvgConsumption.IsPositive() {
		// Utilization is already calculated
	}

	// Save to database
	if err := s.repo.CreateLoadMetric(ctx, metric); err != nil {
		return fmt.Errorf("failed to save load metric: %w", err)
	}

	// Cache for quick access
	ttl := s.config.LoadBalance.CheckInterval * 2
	return s.redis.CacheLoadMetrics(ctx, metric.FacilityID, metric, ttl)
}

// GetLatestLoadMetrics retrieves the latest load metrics for a facility
func (s *LoadBalancingService) GetLatestLoadMetrics(ctx context.Context, facilityID string) ([]*domain.LoadMetric, error) {
	return s.repo.GetLatestLoadMetrics(ctx, facilityID)
}

// EvaluateLoad evaluates the current load and determines if balancing is needed
func (s *LoadBalancingService) EvaluateLoad(ctx context.Context, facilityID string) (*LoadEvaluation, error) {
	metrics, err := s.repo.GetLatestLoadMetrics(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get load metrics: %w", err)
	}

	if len(metrics) == 0 {
		return &LoadEvaluation{
			FacilityID:   facilityID,
			NeedsAction:  false,
			Reason:       "no_metrics_available",
		}, nil
	}

	// Get the most recent metric
	latestMetric := metrics[0]

	evaluation := &LoadEvaluation{
		FacilityID:       facilityID,
		CurrentLoad:      latestMetric.AvgConsumption,
		PeakLoad:         latestMetric.PeakConsumption,
		Utilization:      latestMetric.AvgUtilization,
		NeedsAction:      false,
		Recommendations:  []string{},
	}

	// Check if load is too high
	if latestMetric.AvgUtilization.GreaterThan(decimal.NewFromFloat(s.config.LoadBalance.ScaleUp)) {
		evaluation.NeedsAction = true
		evaluation.Recommendations = append(evaluation.Recommendations,
			fmt.Sprintf("Load is at %.1f%%, consider scaling up or shifting load", latestMetric.AvgUtilization))
	}

	// Check if load is too low
	if latestMetric.AvgUtilization.LessThan(decimal.NewFromFloat(s.config.LoadBalance.ScaleDown)) {
		evaluation.NeedsAction = true
		evaluation.Recommendations = append(evaluation.Recommendations,
			fmt.Sprintf("Load is at %.1f%, consider scaling down to save energy", latestMetric.AvgUtilization))
	}

	// Check if load is optimal
	targetMin := decimal.NewFromFloat(s.config.LoadBalance.TargetMin)
	targetMax := decimal.NewFromFloat(s.config.LoadBalance.TargetMax)

	if latestMetric.AvgUtilization.GreaterThanOrEqual(targetMin) &&
		latestMetric.AvgUtilization.LessThanOrEqual(targetMax) {
		evaluation.Status = "optimal"
	} else if latestMetric.AvgUtilization.LessThan(targetMin) {
		evaluation.Status = "underutilized"
	} else {
		evaluation.Status = "overutilized"
	}

	return evaluation, nil
}

// LoadEvaluation represents the result of a load evaluation
type LoadEvaluation struct {
	FacilityID      string            `json:"facility_id"`
	CurrentLoad     decimal.Decimal   `json:"current_load"`
	PeakLoad        decimal.Decimal   `json:"peak_load"`
	Utilization     decimal.Decimal   `json:"utilization"`
	Status          string            `json:"status"`
	NeedsAction     bool              `json:"needs_action"`
	Recommendations []string          `json:"recommendations"`
}

// CreateLoadBalancingDecision creates a new load balancing decision
func (s *LoadBalancingService) CreateLoadBalancingDecision(ctx context.Context, decision *domain.LoadBalancingDecision) error {
	if decision.ID == "" {
		decision.ID = fmt.Sprintf("lbd-%d", time.Now().UnixNano())
	}

	decision.CreatedAt = time.Now()
	decision.UpdatedAt = time.Now()
	decision.Status = domain.DecisionStatusPending

	if decision.ScheduledAt.IsZero() {
		decision.ScheduledAt = time.Now()
	}

	if err := s.repo.CreateLoadBalancingDecision(ctx, decision); err != nil {
		return fmt.Errorf("failed to create decision: %w", err)
	}

	// Cache the decision
	ttl := 1 * time.Hour
	return s.redis.CacheLoadBalancingDecision(ctx, decision.ID, decision, ttl)
}

// GetPendingDecisions retrieves pending load balancing decisions for a facility
func (s *LoadBalancingService) GetPendingDecisions(ctx context.Context, facilityID string) ([]*domain.LoadBalancingDecision, error) {
	return s.repo.GetPendingDecisions(ctx, facilityID)
}

// ExecuteDecision executes a load balancing decision
func (s *LoadBalancingService) ExecuteDecision(ctx context.Context, decisionID string) error {
	decision, err := s.getCachedDecision(ctx, decisionID)
	if err != nil {
		return err
	}

	if decision == nil {
		// Try to get from database
		decisions, err := s.GetPendingDecisions(ctx, "")
		if err != nil {
			return err
		}
		for _, d := range decisions {
			if d.ID == decisionID {
				decision = d
				break
			}
		}
	}

	if decision == nil {
		return fmt.Errorf("decision not found: %s", decisionID)
	}

	if decision.Status != domain.DecisionStatusPending {
		return fmt.Errorf("decision is not pending: %s", decision.Status)
	}

	// Execute the decision (in a real system, this would trigger actual load balancing)
	now := time.Now()
	decision.Status = domain.DecisionStatusExecuting
	decision.UpdatedAt = now

	// Simulate execution
	time.Sleep(100 * time.Millisecond)

	decision.Status = domain.DecisionStatusCompleted
	decision.ExecutedAt = &now
	decision.Result = fmt.Sprintf("Successfully executed %s: shifted %.2f kW",
		decision.DecisionType, decision.LoadShiftAmount)
	decision.UpdatedAt = time.Now()

	// Update in database
	if err := s.repo.UpdateLoadBalancingDecision(ctx, decision.ID, decision.Status, decision.Result); err != nil {
		return fmt.Errorf("failed to update decision: %w", err)
	}

	// Update cache
	ttl := 5 * time.Minute
	return s.redis.CacheLoadBalancingDecision(ctx, decision.ID, decision, ttl)
}

// CancelDecision cancels a pending load balancing decision
func (s *LoadBalancingService) CancelDecision(ctx context.Context, decisionID string, reason string) error {
	now := time.Now()

	if err := s.repo.UpdateLoadBalancingDecision(ctx, decisionID, domain.DecisionStatusCancelled, reason); err != nil {
		return fmt.Errorf("failed to cancel decision: %w", err)
	}

	// Update cache
	key := fmt.Sprintf("lb_decision:%s", decisionID)
	_ = s.redis.ReleaseLock(ctx, key) // Cleanup lock if any

	return nil
}

// getCachedDecision retrieves a decision from cache
func (s *LoadBalancingService) getCachedDecision(ctx context.Context, decisionID string) (*domain.LoadBalancingDecision, error) {
	data, err := s.redis.GetCachedLoadBalancingDecision(ctx, decisionID)
	if err == nil && data != nil {
		var decision domain.LoadBalancingDecision
		if err := json.Unmarshal(data, &decision); err == nil {
			return &decision, nil
		}
	}
	return nil, nil
}

// json unmarshal helper
func json.Unmarshal(data []byte, v interface{}) error {
	return fmt.Errorf("json.Unmarshal not imported")
}

// AnalyzeAndRecommend performs load analysis and creates recommendations
func (s *LoadBalancingService) AnalyzeAndRecommend(ctx context.Context, facilityID string) ([]*Recommendation, error) {
	// Get latest metrics
	metrics, err := s.repo.GetLatestLoadMetrics(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	var recommendations []*Recommendation

	if len(metrics) == 0 {
		return recommendations, nil
	}

	latestMetric := metrics[0]

	// Analyze power factor
	if latestMetric.PowerFactor.IsPositive() && latestMetric.PowerFactor.LessThan(decimal.NewFromFloat(0.85)) {
		recommendations = append(recommendations, &Recommendation{
			ID:              fmt.Sprintf("rec-pf-%d", time.Now().UnixNano()),
			Category:        "power_quality",
			Title:           "Improve Power Factor",
			Description:     "Power factor is below 0.85. Consider adding power factor correction equipment to reduce reactive power consumption.",
			EstimatedSavings: decimal.NewFromFloat(150.00),
			Priority:        2,
			Implementation:  "Install power factor correction capacitors or synchronous condensers",
		})
	}

	// Analyze harmonic distortion
	if latestMetric.HarmonicDistortion.IsPositive() && latestMetric.HarmonicDistortion.GreaterThan(decimal.NewFromFloat(5.0)) {
		recommendations = append(recommendations, &Recommendation{
			ID:              fmt.Sprintf("rec-harm-%d", time.Now().UnixNano()),
			Category:        "power_quality",
			Title:           "Reduce Harmonic Distortion",
			Description:     "Total harmonic distortion (THD) exceeds 5%. Consider installing harmonic filters.",
			EstimatedSavings: decimal.NewFromFloat(200.00),
			Priority:        2,
			Implementation:  "Install active or passive harmonic filters at the main panel",
		})
	}

	// Analyze peak consumption
	peakThreshold := decimal.NewFromFloat(s.config.LoadBalance.ScaleUp)
	if latestMetric.PeakUtilization.GreaterThan(peakThreshold) {
		recommendations = append(recommendations, &Recommendation{
			ID:              fmt.Sprintf("rec-peak-%d", time.Now().UnixNano()),
			Category:        "demand_management",
			Title:           "Implement Peak Shaving",
			Description:     fmt.Sprintf("Peak utilization at %.1f%% exceeds threshold. Consider peak shaving with battery storage or load shifting.", latestMetric.PeakUtilization),
			EstimatedSavings: decimal.NewFromFloat(500.00),
			Priority:        1,
			Implementation:  "Deploy battery energy storage system (BESS) for peak shaving",
		})
	}

	// Analyze off-peak opportunity
	if latestMetric.AvgUtilization.LessThan(decimal.NewFromFloat(s.config.LoadBalance.TargetMin)) {
		recommendations = append(recommendations, &Recommendation{
			ID:              fmt.Sprintf("rec-offpeak-%d", time.Now().UnixNano()),
			Category:        "load_shifting",
			Title:           "Optimize Off-Peak Usage",
			Description:     "Facility is underutilized. Consider shifting flexible loads to off-peak hours for lower energy costs.",
			EstimatedSavings: decimal.NewFromFloat(300.00),
			Priority:        3,
			Implementation:  "Schedule non-critical loads (HVAC pre-cooling, EV charging, etc.) during off-peak hours",
		})
	}

	return recommendations, nil
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
	ID              string          `json:"id"`
	Category        string          `json:"category"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	EstimatedSavings decimal.Decimal `json:"estimated_savings"`
	Priority        int             `json:"priority"`
	Implementation  string          `json:"implementation"`
}

// AutoBalance performs automated load balancing based on current metrics
func (s *LoadBalancingService) AutoBalance(ctx context.Context, facilityID string) ([]*domain.LoadBalancingDecision, error) {
	// Get zones for this facility
	zones, err := s.repo.GetEnergyZones(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("no zones found for facility: %s", facilityID)
	}

	// Get latest metrics
	metrics, err := s.repo.GetLatestLoadMetrics(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	var decisions []*domain.LoadBalancingDecision

	// Analyze each zone for load balancing opportunities
	for _, zone := range zones {
		if zone.CurrentLoad.GreaterThan(zone.Capacity) {
			// Zone is overloaded, create a scale-down decision
			excessLoad := zone.CurrentLoad.Sub(zone.Capacity)

			decision := &domain.LoadBalancingDecision{
				FacilityID:          facilityID,
				DecisionType:        domain.DecisionTypeScaleDown,
				SourceZoneID:        zone.ID,
				LoadShiftAmount:     excessLoad,
				Reason:              fmt.Sprintf("Zone %s is overloaded (%.2f kW / %.2f kW)", zone.Name, zone.CurrentLoad, zone.Capacity),
				TriggeringCondition: "over_capacity",
				Priority:            1,
				Status:              domain.DecisionStatusPending,
				ScheduledAt:         time.Now().Add(5 * time.Minute),
			}

			if err := s.CreateLoadBalancingDecision(ctx, decision); err != nil {
				return decisions, err
			}

			decisions = append(decisions, decision)
		}

		// Check if zone can shed load
		if zone.CanShedLoad && zone.CurrentLoad.GreaterThan(zone.MinLoad) {
			utilization := zone.CurrentLoad.Div(zone.Capacity).Mul(decimal.NewFromInt(100))

			if utilization.GreaterThan(decimal.NewFromFloat(s.config.LoadBalance.ScaleUp)) {
				// Create a load shedding decision
				sheddableLoad := zone.CurrentLoad.Sub(zone.MinLoad)

				decision := &domain.LoadBalancingDecision{
					FacilityID:          facilityID,
					DecisionType:        domain.DecisionTypeShiftLoad,
					SourceZoneID:        zone.ID,
					LoadShiftAmount:     sheddableLoad,
					Reason:              fmt.Sprintf("Zone %s can shed load during peak period", zone.Name),
					TriggeringCondition: "peak_shaving",
					Priority:            2,
					Status:              domain.DecisionStatusPending,
					ScheduledAt:         time.Now().Add(1 * time.Minute),
				}

				if err := s.CreateLoadBalancingDecision(ctx, decision); err != nil {
					return decisions, err
				}

				decisions = append(decisions, decision)
			}
		}
	}

	return decisions, nil
}

// GetZoneLoads retrieves current loads for all zones in a facility
func (s *LoadBalancingService) GetZoneLoads(ctx context.Context, facilityID string) (map[string]ZoneLoadInfo, error) {
	zones, err := s.repo.GetEnergyZones(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	result := make(map[string]ZoneLoadInfo)

	for _, zone := range zones {
		utilization := decimal.Zero
		if zone.Capacity.IsPositive() {
			utilization = zone.CurrentLoad.Div(zone.Capacity).Mul(decimal.NewFromInt(100))
		}

		var status string
		if utilization.GreaterThan(decimal.NewFromFloat(s.config.LoadBalance.ScaleUp)) {
			status = "overloaded"
		} else if utilization.GreaterThanOrEqual(decimal.NewFromFloat(s.config.LoadBalance.TargetMax)) {
			status = "high"
		} else if utilization.GreaterThanOrEqual(decimal.NewFromFloat(s.config.LoadBalance.TargetMin)) {
			status = "optimal"
		} else {
			status = "low"
		}

		result[zone.ID] = ZoneLoadInfo{
			ZoneID:       zone.ID,
			ZoneName:     zone.Name,
			ZoneType:     string(zone.ZoneType),
			Capacity:     zone.Capacity,
			CurrentLoad:  zone.CurrentLoad,
			Utilization:  utilization,
			Status:       status,
			CanShedLoad:  zone.CanShedLoad,
			Sheddable:    zone.CanShedLoad,
		}
	}

	return result, nil
}

// ZoneLoadInfo represents the load status of a zone
type ZoneLoadInfo struct {
	ZoneID      string          `json:"zone_id"`
	ZoneName    string          `json:"zone_name"`
	ZoneType    string          `json:"zone_type"`
	Capacity    decimal.Decimal `json:"capacity"`
	CurrentLoad decimal.Decimal `json:"current_load"`
	Utilization decimal.Decimal `json:"utilization"`
	Status      string          `json:"status"`
	CanShedLoad bool            `json:"can_shed_load"`
	Sheddable   bool            `json:"sheddable"`
}

// GetOptimizationReport generates an optimization report for a facility
func (s *LoadBalancingService) GetOptimizationReport(ctx context.Context, facilityID string, periodStart, periodEnd time.Time) (*domain.EnergyOptimizationReport, error) {
	// Get carbon emissions
	carbonEmissions, err := s.repo.GetCarbonEmissionsByFacility(ctx, facilityID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get carbon emissions: %w", err)
	}

	// Get load metrics
	loadMetrics, err := s.repo.GetLatestLoadMetrics(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get load metrics: %w", err)
	}

	// Calculate totals
	var totalCarbonEmission decimal.Decimal
	var peakDemand decimal.Decimal

	for _, ce := range carbonEmissions {
		totalCarbonEmission = totalCarbonEmission.Add(ce.TotalEmission)
	}

	for _, lm := range loadMetrics {
		if lm.PeakConsumption.GreaterThan(peakDemand) {
			peakDemand = lm.PeakConsumption
		}
	}

	// Get recommendations
	recommendations, err := s.AnalyzeAndRecommend(ctx, facilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Estimate savings based on recommendations
	var totalSavings decimal.Decimal
	for _, rec := range recommendations {
		if rec.Priority == 1 {
			totalSavings = totalSavings.Add(rec.EstimatedSavings)
		}
	}

	report := &domain.EnergyOptimizationReport{
		ID:                  fmt.Sprintf("report-%d", time.Now().UnixNano()),
		FacilityID:          facilityID,
		ReportType:          domain.ReportTypeCustom,
		PeriodStart:         periodStart,
		PeriodEnd:           periodEnd,
		TotalConsumption:    decimal.Zero, // Would be calculated from readings
		PeakDemand:          peakDemand,
		TotalCarbonEmission: totalCarbonEmission,
		CarbonSavings:       totalSavings.Div(decimal.NewFromFloat(10)), // Rough estimate
		EnergySavings:       totalSavings.Div(decimal.NewFromFloat(0.15)), // Assuming $0.15/kWh
		CostSavings:         totalSavings,
		OptimizationsApplied: len(recommendations),
		Recommendations:     convertToDomainRecommendations(recommendations),
		GeneratedAt:         time.Now(),
	}

	return report, nil
}

// convertToDomainRecommendations converts service recommendations to domain recommendations
func convertToDomainRecommendations(recs []*Recommendation) []domain.Recommendation {
	result := make([]domain.Recommendation, len(recs))
	for i, r := range recs {
		result[i] = domain.Recommendation{
			ID:                r.ID,
			Category:          r.Category,
			Title:             r.Title,
			Description:       r.Description,
			EstimatedSavings:  r.EstimatedSavings,
			Priority:          r.Priority,
			Implementation:    r.Implementation,
			Status:            domain.RecStatusPending,
			CreatedAt:         time.Now(),
		}
	}
	return result
}
