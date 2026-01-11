package service

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// EnergyRepository defines the interface for energy data persistence
type EnergyRepository interface {
	StoreLog(ctx context.Context, log *domain.EnergyConsumptionLog) error
	StoreBatch(ctx context.Context, logs []domain.EnergyConsumptionLog) error
	GetLatest(ctx context.Context, poolID uuid.UUID) (*domain.EnergyConsumptionLog, error)
	GetTimeSeries(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) ([]domain.EnergyConsumptionLog, error)
	GetAggregated(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time, interval string) (*EnergyAggregation, error)
}

// HashRateRepository defines the interface for hash rate data persistence
type HashRateRepository interface {
	StoreMetric(ctx context.Context, metric *domain.HashRateMetric) error
	StoreBatch(ctx context.Context, metrics []domain.HashRateMetric) error
	GetLatest(ctx context.Context, poolID uuid.UUID) (*domain.HashRateMetric, error)
	GetTimeSeries(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) ([]domain.HashRateMetric, error)
	GetAggregated(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time, interval string) (*HashRateAggregation, error)
}

// ViolationRepository defines the interface for violation persistence
type ViolationRepository interface {
	Create(ctx context.Context, violation *domain.ComplianceViolation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ComplianceViolation, error)
	Update(ctx context.Context, violation *domain.ComplianceViolation) error
	GetOpenViolations(ctx context.Context, poolID uuid.UUID) ([]domain.ComplianceViolation, error)
	List(ctx context.Context, filter domain.ViolationFilter, limit, offset int) ([]domain.ComplianceViolation, error)
}

// EnergyAggregation represents aggregated energy data
type EnergyAggregation struct {
	PeriodStart       time.Time
	PeriodEnd         time.Time
	TotalKWh          decimal.Decimal
	PeakKW            decimal.Decimal
	AverageKW         decimal.Decimal
	TotalCarbonKG     decimal.Decimal
	BySource          map[string]decimal.Decimal
	RecordCount       int64
}

// HashRateAggregation represents aggregated hash rate data
type HashRateAggregation struct {
	PeriodStart        time.Time
	PeriodEnd          time.Time
	AverageHashRate    decimal.Decimal
	MaxHashRate        decimal.Decimal
	MinHashRate        decimal.Decimal
	TotalWorkers       int64
	AverageEfficiency  decimal.Decimal
	RecordCount        int64
}

// MonitoringConfig holds configuration for the monitoring service
type MonitoringConfig struct {
	EnergyThresholds       EnergyThresholdConfig
	HashRateThresholds     HashRateThresholdConfig
	ComplianceCheckInterval time.Duration
	AnomalyCheckEnabled    bool
}

// EnergyThresholdConfig holds energy monitoring thresholds
type EnergyThresholdConfig struct {
	WarningLimitPercent    float64
	CriticalLimitPercent   float64
	OverageGracePeriod     time.Duration
	CarbonFactorDefault    float64
	CarbonFactorsByRegion  map[string]float64
}

// HashRateThresholdConfig holds hash rate monitoring thresholds
type HashRateThresholdConfig struct {
	AnomalyStdDev        float64
	MinReportedHashRate  float64
	MaxReportedHashRate  float64
	PowerHashRateMin     float64
	PowerHashRateMax     float64
}

// MonitoringService handles energy and hash rate monitoring
type MonitoringService struct {
	energyRepo   EnergyRepository
	hashRepo     HashRateRepository
	poolRepo     MiningPoolRepository
	machineRepo  MachineRepository
	violationRepo ViolationRepository
	config       MonitoringConfig
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(energyRepo EnergyRepository, hashRepo HashRateRepository, poolRepo MiningPoolRepository, machineRepo MachineRepository, violationRepo ViolationRepository) *MonitoringService {
	return &MonitoringService{
		energyRepo:   energyRepo,
		hashRepo:     hashRepo,
		poolRepo:     poolRepo,
		machineRepo:  machineRepo,
		violationRepo: violationRepo,
		config: MonitoringConfig{
			EnergyThresholds: EnergyThresholdConfig{
				WarningLimitPercent:  0.8,
				CriticalLimitPercent: 1.0,
				OverageGracePeriod:   15 * time.Minute,
				CarbonFactorDefault:  0.5,
				CarbonFactorsByRegion: map[string]float64{
					"US-CA": 0.3, "US-WA": 0.1, "IS": 0.02, "NO": 0.02,
				},
			},
			HashRateThresholds: HashRateThresholdConfig{
				AnomalyStdDev:       3.0,
				MinReportedHashRate: 0.001,
				MaxReportedHashRate: 1000000,
				PowerHashRateMin:    0.001,
				PowerHashRateMax:    0.1,
			},
			ComplianceCheckInterval: 1 * time.Hour,
			AnomalyCheckEnabled:     true,
		},
		stopChan: make(chan struct{}),
	}
}

// StartComplianceChecker starts the background compliance checker
func (s *MonitoringService) StartComplianceChecker() {
	s.wg.Add(1)
	go s.complianceChecker()
	log.Println("Compliance checker started")
}

// StopComplianceChecker stops the background compliance checker
func (s *MonitoringService) StopComplianceChecker() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Compliance checker stopped")
}

// complianceChecker runs periodic compliance checks
func (s *MonitoringService) complianceChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.ComplianceCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.runComplianceChecks()
		}
	}
}

// runComplianceChecks runs compliance checks for all active pools
func (s *MonitoringService) runComplianceChecks() {
	ctx := context.Background()

	// Get all active pools
	filter := domain.PoolFilter{
		Statuses: []domain.PoolStatus{domain.PoolStatusActive},
	}
	pools, err := s.poolRepo.List(ctx, filter, 1000, 0)
	if err != nil {
		log.Printf("Failed to list pools for compliance check: %v", err)
		return
	}

	for _, pool := range pools {
		s.checkPoolCompliance(ctx, &pool)
	}
}

// checkPoolCompliance performs compliance checks for a single pool
func (s *MonitoringService) checkPoolCompliance(ctx context.Context, pool *domain.MiningPool) {
	// Check energy usage against limit
	usagePercent := pool.CurrentEnergyUsageKW.Div(pool.MaxAllowedEnergyKW)

	if usagePercent.GreaterThan(decimal.NewFromFloat(s.config.EnergyThresholds.CriticalLimitPercent)) {
		violation := &domain.ComplianceViolation{
			ID:            uuid.New(),
			PoolID:        pool.ID,
			PoolName:      pool.Name,
			ViolationType: domain.ViolationTypeExcessEnergy,
			Severity:      domain.ViolationSeverityCritical,
			Status:        domain.ViolationStatusOpen,
			Title:         "Energy usage exceeds allowed limit",
			Description:   "Current energy usage exceeds the maximum allowed energy limit for this mining pool",
			Details: map[string]interface{}{
				"current_usage_kw":  pool.CurrentEnergyUsageKW.String(),
				"max_allowed_kw":    pool.MaxAllowedEnergyKW.String(),
				"usage_percent":     usagePercent.Mul(decimal.NewFromInt(100)).String(),
			},
			DetectedAt: time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.violationRepo.Create(ctx, violation); err != nil {
			log.Printf("Failed to create energy violation for pool %s: %v", pool.ID, err)
		} else {
			log.Printf("Created CRITICAL energy violation for pool %s", pool.ID)
		}
	} else if usagePercent.GreaterThan(decimal.NewFromFloat(s.config.EnergyThresholds.WarningLimitPercent)) {
		violation := &domain.ComplianceViolation{
			ID:            uuid.New(),
			PoolID:        pool.ID,
			PoolName:      pool.Name,
			ViolationType: domain.ViolationTypeExcessEnergy,
			Severity:      domain.ViolationSeverityMedium,
			Status:        domain.ViolationStatusOpen,
			Title:         "Energy usage approaching limit",
			Description:   "Current energy usage is approaching the maximum allowed energy limit",
			Details: map[string]interface{}{
				"current_usage_kw":  pool.CurrentEnergyUsageKW.String(),
				"max_allowed_kw":    pool.MaxAllowedEnergyKW.String(),
				"usage_percent":     usagePercent.Mul(decimal.NewFromInt(100)).String(),
			},
			DetectedAt: time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.violationRepo.Create(ctx, violation); err != nil {
			log.Printf("Failed to create energy warning for pool %s: %v", pool.ID, err)
		}
	}
}

// ReportEnergyConsumption reports energy consumption for a pool
func (s *MonitoringService) ReportEnergyConsumption(ctx context.Context, packet *domain.EnergyTelemetryPacket) (*domain.EnergyConsumptionLog, error) {
	// Verify pool exists
	pool, err := s.poolRepo.GetByID(ctx, packet.PoolID)
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

	// Calculate carbon emissions
	carbonEmissions := s.calculateCarbonEmissions(packet)

	// Create the log
	logEntry := &domain.EnergyConsumptionLog{
		ID:                  uuid.New(),
		PoolID:              packet.PoolID,
		Timestamp:           packet.Timestamp,
		DurationMinutes:     packet.DurationMinutes,
		PowerUsageKWh:       packet.PowerUsageKWh,
		PeakPowerKW:         packet.PeakPowerKW,
		AveragePowerKW:      packet.AveragePowerKW,
		EnergySource:        packet.EnergySource,
		SourceMixPercentage: s.convertSourceMix(packet.SourceMixPercentage),
		CarbonEmissionsKG:   carbonEmissions,
		GridRegionCode:      packet.GridRegionCode,
		TemperatureC:        packet.TemperatureC,
		HumidityPercent:     packet.HumidityPercent,
		DataQuality:         domain.DataQualityGood,
		SubmittedAt:         time.Now(),
		Checksum:            "",
	}

	// Store the log
	if err := s.energyRepo.StoreLog(ctx, logEntry); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to store energy consumption log",
			Err:     err,
		}
	}

	// Update pool energy usage
	s.poolRepo.UpdateEnergyUsage(ctx, packet.PoolID, packet.AveragePowerKW)

	log.Printf("Recorded energy consumption for pool %s: %.2f kWh", packet.PoolID, packet.PowerUsageKWh)

	return logEntry, nil
}

// BatchReportEnergyConsumption reports batch energy consumption
func (s *MonitoringService) BatchReportEnergyConsumption(ctx context.Context, packets []domain.EnergyTelemetryPacket) ([]domain.EnergyConsumptionLog, error) {
	if len(packets) == 0 {
		return nil, nil
	}

	var logs []domain.EnergyConsumptionLog

	for _, packet := range packets {
		logEntry, err := s.ReportEnergyConsumption(ctx, &packet)
		if err != nil {
			return logs, err
		}
		logs = append(logs, *logEntry)
	}

	log.Printf("Batch recorded %d energy consumption logs", len(logs))

	return logs, nil
}

// ReportHashRate reports hash rate metrics for a pool
func (s *MonitoringService) ReportHashRate(ctx context.Context, packet *domain.HashRateTelemetryPacket) (*domain.HashRateMetric, error) {
	// Verify pool exists
	pool, err := s.poolRepo.GetByID(ctx, packet.PoolID)
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

	// Validate hash rate
	if err := s.validateHashRate(packet, pool); err != nil {
		return nil, err
	}

	// Create the metric
	metric := &domain.HashRateMetric{
		ID:                uuid.New(),
		PoolID:            packet.PoolID,
		Timestamp:         packet.Timestamp,
		ReportedHashRate:  packet.ReportedHashRate,
		CalculatedHashRate: packet.ReportedHashRate,
		ActiveWorkerCount: packet.ActiveWorkerCount,
		StaleWorkerCount:  packet.StaleWorkerCount,
		TotalShares:       packet.TotalShares,
		AcceptedShares:    packet.AcceptedShares,
		RejectedShares:    packet.RejectedShares,
		Difficulty:        packet.Difficulty,
		PoolFeePercent:    packet.PoolFeePercent,
		DataQuality:       domain.DataQualityGood,
		SubmittedAt:       time.Now(),
		Checksum:          "",
	}

	// Check for anomalies
	if s.config.AnomalyCheckEnabled {
		anomalyCheck := s.checkHashRateAnomaly(metric, pool)
		if anomalyCheck != nil {
			metric.DataQuality = domain.DataQualitySuspect
			// Create anomaly violation
			violation := &domain.ComplianceViolation{
				ID:            uuid.New(),
				PoolID:        pool.ID,
				PoolName:      pool.Name,
				ViolationType: domain.ViolationTypeHashRateAnomaly,
				Severity:      domain.ViolationSeverityMedium,
				Status:        domain.ViolationStatusOpen,
				Title:         "Hash rate anomaly detected",
				Description:   "Reported hash rate differs significantly from expected values",
				Details:       anomalyCheck,
				DetectedAt:    time.Now(),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			s.violationRepo.Create(ctx, violation)
		}
	}

	// Store the metric
	if err := s.hashRepo.StoreMetric(ctx, metric); err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to store hash rate metric",
			Err:     err,
		}
	}

	// Update pool hash rate
	s.poolRepo.UpdateHashRate(ctx, packet.PoolID, packet.ReportedHashRate)

	log.Printf("Recorded hash rate for pool %s: %.2f TH/s", packet.PoolID, packet.ReportedHashRate)

	return metric, nil
}

// BatchReportHashRate reports batch hash rate metrics
func (s *MonitoringService) BatchReportHashRate(ctx context.Context, packets []domain.HashRateTelemetryPacket) ([]domain.HashRateMetric, error) {
	if len(packets) == 0 {
		return nil, nil
	}

	var metrics []domain.HashRateMetric

	for _, packet := range packets {
		metric, err := s.ReportHashRate(ctx, &packet)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, *metric)
	}

	log.Printf("Batch recorded %d hash rate metrics", len(metrics))

	return metrics, nil
}

// GetEnergyUsage retrieves energy usage for a pool
func (s *MonitoringService) GetEnergyUsage(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) ([]domain.EnergyConsumptionLog, error) {
	logs, err := s.energyRepo.GetTimeSeries(ctx, poolID, startTime, endTime)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve energy usage",
			Err:     err,
		}
	}

	return logs, nil
}

// GetHashRateMetrics retrieves hash rate metrics for a pool
func (s *MonitoringService) GetHashRateMetrics(ctx context.Context, poolID uuid.UUID, startTime, endTime time.Time) ([]domain.HashRateMetric, error) {
	metrics, err := s.hashRepo.GetTimeSeries(ctx, poolID, startTime, endTime)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve hash rate metrics",
			Err:     err,
		}
	}

	return metrics, nil
}

// GetComplianceStatus retrieves compliance status for a pool
func (s *MonitoringService) GetComplianceStatus(ctx context.Context, poolID uuid.UUID) (*ComplianceStatus, error) {
	violations, err := s.violationRepo.GetOpenViolations(ctx, poolID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violations",
			Err:     err,
		}
	}

	var criticalCount, highCount, mediumCount, lowCount int
	for _, v := range violations {
		switch v.Severity {
		case domain.ViolationSeverityCritical:
			criticalCount++
		case domain.ViolationSeverityHigh:
			highCount++
		case domain.ViolationSeverityMedium:
			mediumCount++
		case domain.ViolationSeverityLow:
			lowCount++
		}
	}

	status := &ComplianceStatus{
		PoolID:          poolID,
		OpenViolations:  len(violations),
		CriticalCount:   criticalCount,
		HighCount:       highCount,
		MediumCount:     mediumCount,
		LowCount:        lowCount,
		IsCompliant:     criticalCount == 0 && highCount == 0,
		LastCheckAt:     time.Now(),
	}

	return status, nil
}

// ComplianceStatus represents the compliance status of a pool
type ComplianceStatus struct {
	PoolID         uuid.UUID
	OpenViolations int
	CriticalCount  int
	HighCount      int
	MediumCount    int
	LowCount       int
	IsCompliant    bool
	LastCheckAt    time.Time
}

// calculateCarbonEmissions calculates carbon emissions for energy consumption
func (s *MonitoringService) calculateCarbonEmissions(packet *domain.EnergyTelemetryPacket) decimal.Decimal {
	factor := s.config.EnergyThresholds.CarbonFactorDefault

	if regionFactor, ok := s.config.EnergyThresholds.CarbonFactorsByRegion[packet.GridRegionCode]; ok {
		factor = regionFactor
	}

	// Adjust for energy source
	switch packet.EnergySource {
	case domain.EnergySourceSolar, domain.EnergySourceHydro, domain.EnergySourceWind, domain.EnergySourceNuclear:
		factor *= 0.1 // Renewable sources have much lower carbon footprint
	case domain.EnergySourceCoal:
		factor *= 1.5 // Coal has higher carbon footprint
	case domain.EnergySourceNaturalGas:
		factor *= 0.8 // Natural gas is cleaner than coal
	case domain.EnergySourceMixed:
		factor *= 0.9 // Mixed sources
	}

	return packet.PowerUsageKWh.Mul(decimal.NewFromFloat(factor))
}

// convertSourceMix converts string map to domain map
func (s *MonitoringService) convertSourceMix(sourceMix map[string]int) map[string]int {
	result := make(map[string]int)
	for k, v := range sourceMix {
		result[k] = v
	}
	return result
}

// validateHashRate validates hash rate data
func (s *MonitoringService) validateHashRate(packet *domain.HashRateTelemetryPacket, pool *domain.MiningPool) error {
	hashRateFloat := packet.ReportedHashRate.InexactFloat64()

	if hashRateFloat < s.config.HashRateThresholds.MinReportedHashRate {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Hash rate is below minimum threshold",
		}
	}

	if hashRateFloat > s.config.HashRateThresholds.MaxReportedHashRate {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Hash rate exceeds maximum threshold",
		}
	}

	// Check power to hash rate ratio
	if pool.CurrentEnergyUsageKW.GreaterThan(decimal.Zero) {
		powerKW := pool.CurrentEnergyUsageKW
		hashRateTH := packet.ReportedHashRate
		ratio := powerKW.Div(hashRateTH)

		minRatio := decimal.NewFromFloat(s.config.HashRateThresholds.PowerHashRateMin)
		maxRatio := decimal.NewFromFloat(s.config.HashRateThresholds.PowerHashRateMax)

		if ratio.LessThan(minRatio) {
			return &ServiceError{
				Code:    "VALIDATION_ERROR",
				Message: "Power to hash rate ratio is suspiciously low - possible reporting error",
			}
		}

		if ratio.GreaterThan(maxRatio) {
			return &ServiceError{
				Code:    "VALIDATION_ERROR",
				Message: "Power to hash rate ratio exceeds expected range",
			}
		}
	}

	return nil
}

// checkHashRateAnomaly checks for hash rate anomalies
func (s *MonitoringService) checkHashRateAnomaly(metric *domain.HashRateMetric, pool *domain.MiningPool) map[string]interface{} {
	// Get recent metrics for comparison
	recentMetrics, err := s.hashRepo.GetTimeSeries(context.Background(), pool.ID, time.Now().Add(-1*time.Hour), time.Now())
	if err != nil || len(recentMetrics) < 5 {
		return nil // Not enough data for comparison
	}

	// Calculate mean and standard deviation
	var sum decimal.Decimal
	for _, m := range recentMetrics {
		sum = sum.Add(m.ReportedHashRate)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(recentMetrics))))

	var sumSq decimal.Decimal
	for _, m := range recentMetrics {
		diff := m.ReportedHashRate.Sub(mean)
		sumSq = sumSq.Add(diff.Mul(diff))
	}
	stdDev := sumSq.Div(decimal.NewFromInt(int64(len(recentments))))

	// Check if current metric is an outlier
	zScore := metric.ReportedHashRate.Sub(mean).Abs().Div(stdDev)

	if zScore.GreaterThan(decimal.NewFromFloat(s.config.HashRateThresholds.AnomalyStdDev)) {
		return map[string]interface{}{
			"current_hash_rate":   metric.ReportedHashRate.String(),
			"mean_hash_rate":      mean.String(),
			"standard_deviation":  stdDev.String(),
			"z_score":             zScore.String(),
			"threshold":           s.config.HashRateThresholds.AnomalyStdDev,
		}
	}

	return nil
}
