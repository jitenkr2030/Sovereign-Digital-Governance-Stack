package service

import (
	"context"
	"log"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReportingService handles report generation
type ReportingService struct {
	poolRepo      MiningPoolRepository
	energyRepo    EnergyRepository
	hashRepo      HashRateRepository
	violationRepo ViolationRepository
}

// NewReportingService creates a new reporting service
func NewReportingService(poolRepo MiningPoolRepository, energyRepo EnergyRepository, hashRepo HashRateRepository, violationRepo ViolationRepository) *ReportingService {
	return &ReportingService{
		poolRepo:      poolRepo,
		energyRepo:    energyRepo,
		hashRepo:      hashRepo,
		violationRepo: violationRepo,
	}
}

// GetRegionalEnergyLoad retrieves energy load statistics by region
func (s *ReportingService) GetRegionalEnergyLoad(ctx context.Context, startTime, endTime time.Time) ([]domain.RegionalEnergyStats, error) {
	// Get all active pools
	filter := domain.PoolFilter{}
	pools, err := s.poolRepo.List(ctx, filter, 1000, 0)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pools",
			Err:     err,
		}
	}

	// Aggregate by region
	regionStats := make(map[string]*domain.RegionalEnergyStats)
	now := time.Now()

	for _, pool := range pools {
		stats, exists := regionStats[pool.RegionCode]
		if !exists {
			stats = &domain.RegionalEnergyStats{
				RegionCode:   pool.RegionCode,
				PeriodStart:  startTime,
				PeriodEnd:    endTime,
				TotalPools:   0,
				ActivePools:  0,
				BySource:     make(map[string]decimal.Decimal),
			}
			regionStats[pool.RegionCode] = stats
		}

		stats.TotalPools++
		if pool.Status == domain.PoolStatusActive {
			stats.ActivePools++
		}

		// Get energy data for the pool
		energyData, err := s.energyRepo.GetTimeSeries(ctx, pool.ID, startTime, endTime)
		if err != nil {
			log.Printf("Failed to get energy data for pool %s: %v", pool.ID, err)
			continue
		}

		for _, e := range energyData {
			stats.TotalEnergyKWh = stats.TotalEnergyKWh.Add(e.PowerUsageKWh)
			if e.PeakPowerKW.GreaterThan(stats.PeakPowerKW) {
				stats.PeakPowerKW = e.PeakPowerKW
			}
			stats.TotalCarbonKG = stats.TotalCarbonKG.Add(e.CarbonEmissionsKG)
		}

		// Get open violations
		violations, err := s.violationRepo.GetOpenViolations(ctx, pool.ID)
		if err != nil {
			log.Printf("Failed to get violations for pool %s: %v", pool.ID, err)
			continue
		}
		stats.OpenViolations += len(violations)
	}

	// Convert map to slice
	result := make([]domain.RegionalEnergyStats, 0, len(regionStats))
	for _, stats := range regionStats {
		if stats.TotalPools > 0 {
			stats.AveragePowerKW = stats.TotalEnergyKWh.Div(decimal.NewFromInt(int64(stats.ActivePools)))
			stats.UpdatedAt = now
			result = append(result, *stats)
		}
	}

	return result, nil
}

// GetCarbonFootprintReport generates a carbon footprint report for a pool
func (s *ReportingService) GetCarbonFootprintReport(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) (*domain.CarbonFootprintReport, error) {
	pool, err := s.poolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	// Get energy data
	energyData, err := s.energyRepo.GetTimeSeries(ctx, poolID, startTime, endTime)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve energy data",
			Err:     err,
		}
	}

	totalKWh := decimal.Zero
	totalCarbon := decimal.Zero
	renewableKWh := decimal.Zero
	bySource := make(map[string]decimal.Decimal)

	for _, e := range energyData {
		totalKWh = totalKWh.Add(e.PowerUsageKWh)
		totalCarbon = totalCarbon.Add(e.CarbonEmissionsKG)

		source := string(e.EnergySource)
		bySource[source] = bySource[source].Add(e.PowerUsageKWh)

		// Calculate renewable energy
		switch e.EnergySource {
		case domain.EnergySourceSolar, domain.EnergySourceHydro, domain.EnergySourceWind:
			renewableKWh = renewableKWh.Add(e.PowerUsageKWh)
		case domain.EnergySourceMixed:
			// Estimate 50% renewable for mixed
			renewableKWh = renewableKWh.Add(e.PowerUsageKWh.Mul(decimal.NewFromFloat(0.5)))
		}
	}

	var carbonPerKWh, renewablePercent decimal.Decimal
	if totalKWh.GreaterThan(decimal.Zero) {
		carbonPerKWh = totalCarbon.Div(totalKWh)
		renewablePercent = renewableKWh.Div(totalKWh).Mul(decimal.NewFromInt(100))
	}

	// Calculate rating
	rating := s.calculateCarbonRating(carbonPerKWh)

	report := &domain.CarbonFootprintReport{
		PoolID:              poolID,
		PoolName:            pool.Name,
		ReportPeriodStart:   startTime,
		ReportPeriodEnd:     endTime,
		TotalEnergyKWh:      totalKWh,
		TotalCarbonKG:       totalCarbon,
		CarbonPerKWh:        carbonPerKWh,
		RenewableEnergyKWh:  renewableKWh,
		RenewablePercent:    renewablePercent,
		BySource:            bySource,
		ComparedToAverage:   decimal.Zero, // Would require comparison data
		Rating:              rating,
		GeneratedAt:         time.Now(),
	}

	return report, nil
}

// GetMiningSummaryReport generates a comprehensive mining summary report
func (s *ReportingService) GetMiningSummaryReport(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) (*domain.MiningSummaryReport, error) {
	pool, err := s.poolRepo.GetByID(ctx, poolID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return nil, &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	// Get machine stats
	machineStats, err := s.machineRepo.GetPoolMachinesStats(ctx, poolID)
	if err != nil {
		log.Printf("Failed to get machine stats for pool %s: %v", poolID, err)
		machineStats = &MachineStats{}
	}

	// Get energy data
	energyData, err := s.energyRepo.GetTimeSeries(ctx, poolID, startTime, endTime)
	if err != nil {
		log.Printf("Failed to get energy data for pool %s: %v", poolID, err)
		energyData = []domain.EnergyConsumptionLog{}
	}

	// Get hash rate data
	hashData, err := s.hashRepo.GetTimeSeries(ctx, poolID, startTime, endTime)
	if err != nil {
		log.Printf("Failed to get hash rate data for pool %s: %v", poolID, err)
		hashData = []domain.HashRateMetric{}
	}

	// Get violations
	violations, err := s.violationRepo.GetOpenViolations(ctx, poolID)
	if err != nil {
		log.Printf("Failed to get violations for pool %s: %v", poolID, err)
		violations = []domain.ComplianceViolation{}
	}

	// Calculate totals
	totalEnergyKWh := decimal.Zero
	peakPower := decimal.Zero
	totalCarbon := decimal.Decimal
	var avgPower decimal.Decimal

	if len(energyData) > 0 {
		for _, e := range energyData {
			totalEnergyKWh = totalEnergyKWh.Add(e.PowerUsageKWh)
			if e.PeakPowerKW.GreaterThan(peakPower) {
				peakPower = e.PeakPowerKW
			}
			totalCarbon = totalCarbon.Add(e.CarbonEmissionsKG)
		}
		avgPower = totalEnergyKWh.Div(decimal.NewFromInt(int64(len(energyData))))
	}

	// Calculate total hash rate
	var totalHashRate decimal.Decimal
	for _, h := range hashData {
		if h.ReportedHashRate.GreaterThan(totalHashRate) {
			totalHashRate = h.ReportedHashRate
		}
	}

	// Calculate energy limit utilization
	var energyUtilization decimal.Decimal
	if pool.MaxAllowedEnergyKW.GreaterThan(decimal.Zero) {
		energyUtilization = avgPower.Div(pool.MaxAllowedEnergyKW).Mul(decimal.NewFromInt(100))
	}

	// Calculate compliance score
	complianceScore := s.calculateComplianceScore(violations)

	// Count resolved violations
	resolvedFilter := domain.ViolationFilter{
		Statuses: []domain.ViolationStatus{domain.ViolationStatusResolved, domain.ViolationStatusDismissed},
	}
	resolvedViolations, _ := s.violationRepo.List(ctx, resolvedFilter, 0, 0)

	report := &domain.MiningSummaryReport{
		PoolID:               poolID,
		ReportPeriodStart:    startTime,
		ReportPeriodEnd:      endTime,
		PoolName:             pool.Name,
		LicenseNumber:        pool.LicenseNumber,
		RegionCode:           pool.RegionCode,
		Status:               pool.Status,
		TotalMachines:        machineStats.TotalMachines,
		ActiveMachines:       machineStats.ActiveMachines,
		OfflineMachines:      machineStats.OfflineMachines + machineStats.MaintenanceCount,
		TotalHashRateTH:      totalHashRate,
		AverageEfficiency:    decimal.Zero, // Would require calculation
		TotalEnergyKWh:       totalEnergyKWh,
		PeakPowerKW:          peakPower,
		AveragePowerKW:       avgPower,
		EnergyLimitUtilization: energyUtilization,
		TotalCarbonKG:        totalCarbon,
		CarbonPerTH:          decimal.Zero, // Would require calculation
		RenewablePercent:     decimal.Zero, // Would require calculation
		OpenViolations:       len(violations),
		ResolvedViolations:   len(resolvedViolations),
		ComplianceScore:      complianceScore,
		CertificateStatus:    domain.CertificateStatusValid,
		GeneratedAt:          time.Now(),
	}

	return report, nil
}

// GetDashboardStats retrieves dashboard statistics
func (s *ReportingService) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	// Get all pools
	filter := domain.PoolFilter{}
	pools, err := s.poolRepo.List(ctx, filter, 10000, 0)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pools",
			Err:     err,
		}
	}

	now := time.Now()
	startTime := now.AddDate(0, 0, -1) // Last 24 hours

	stats := &domain.DashboardStats{
		TotalPools:     len(pools),
		TotalMachines:  0,
		ActiveMachines: 0,
		ByRegion:       []domain.RegionalEnergyStats{},
		ViolationTrend: []domain.TrendPoint{},
		EnergyTrend:    []domain.TrendPoint{},
		UpdatedAt:      now,
	}

	var totalEnergy, totalCarbon, totalHashRate decimal.Decimal
	regionMap := make(map[string]*domain.RegionalEnergyStats)

	for _, pool := range pools {
		// Count by status
		switch pool.Status {
		case domain.PoolStatusActive:
			stats.ActivePools++
		case domain.PoolStatusSuspended:
			stats.SuspendedPools++
		}

		stats.TotalMachines += pool.TotalMachineCount
		stats.ActiveMachines += pool.ActiveMachineCount

		// Get energy data for the pool
		energyData, _ := s.energyRepo.GetTimeSeries(ctx, pool.ID, startTime, now)
		for _, e := range energyData {
			totalEnergy = totalEnergy.Add(e.PowerUsageKWh)
			totalCarbon = totalCarbon.Add(e.CarbonEmissionsKG)
			if e.PeakPowerKW.GreaterThan(stats.PeakPowerKW) {
				stats.PeakPowerKW = e.PeakPowerKW
			}
		}

		// Get hash rate
		hashData, _ := s.hashRepo.GetTimeSeries(ctx, pool.ID, startTime, now)
		for _, h := range hashData {
			totalHashRate = totalHashRate.Add(h.ReportedHashRate)
		}

		// Aggregate by region
		if region, exists := regionMap[pool.RegionCode]; exists {
			region.TotalPools++
			if pool.Status == domain.PoolStatusActive {
				region.ActivePools++
			}
		} else {
			regionMap[pool.RegionCode] = &domain.RegionalEnergyStats{
				RegionCode: pool.RegionCode,
				TotalPools: 1,
				ActivePools: 1,
			}
		}

		// Get open violations
		violations, _ := s.violationRepo.GetOpenViolations(ctx, pool.ID)
		stats.OpenViolations += len(violations)
		for _, v := range violations {
			if v.Severity == domain.ViolationSeverityCritical {
				stats.CriticalViolations++
			}
		}
	}

	stats.TotalEnergyKWh = totalEnergy
	stats.TotalCarbonKG = totalCarbon
	stats.TotalHashRateTH = totalHashRate

	// Convert region map to slice
	for _, region := range regionMap {
		stats.ByRegion = append(stats.ByRegion, *region)
	}

	return stats, nil
}

// calculateCarbonRating calculates a carbon rating based on emissions
func (s *ReportingService) calculateCarbonRating(carbonPerKWh decimal.Decimal) string {
	carbonFloat := carbonPerKWh.InexactFloat64()

	if carbonFloat < 0.05 {
		return "A+"
	} else if carbonFloat < 0.1 {
		return "A"
	} else if carbonFloat < 0.2 {
		return "B"
	} else if carbonFloat < 0.4 {
		return "C"
	} else if carbonFloat < 0.6 {
		return "D"
	} else {
		return "E"
	}
}

// calculateComplianceScore calculates a compliance score based on violations
func (s *ReportingService) calculateComplianceScore(violations []domain.ComplianceViolation) decimal.Decimal {
	if len(violations) == 0 {
		return decimal.NewFromInt(100)
	}

	score := decimal.NewFromInt(100)
	for _, v := range violations {
		switch v.Severity {
		case domain.ViolationSeverityLow:
			score = score.Sub(decimal.NewFromInt(1))
		case domain.ViolationSeverityMedium:
			score = score.Sub(decimal.NewFromInt(5))
		case domain.ViolationSeverityHigh:
			score = score.Sub(decimal.NewFromInt(10))
		case domain.ViolationSeverityCritical:
			score = score.Sub(decimal.NewFromInt(25))
		}
	}

	if score.LessThan(decimal.Zero) {
		return decimal.Zero
	}

	return score
}
