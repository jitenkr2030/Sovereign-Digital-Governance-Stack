package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/risk-engine/internal/core/domain"
	"github.com/csic-platform/services/risk-engine/internal/core/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RiskEvaluationServiceImpl implements the RiskEvaluationService interface
type RiskEvaluationServiceImpl struct {
	ruleRepo     ports.RiskRuleRepository
	alertRepo    ports.RiskAlertRepository
	exposureRepo ports.ExposureRepository
	profileRepo  ports.RiskProfileRepository
	alertPub     ports.AlertPublisher
	metrics      ports.MetricsCollector
	compliance   ports.ComplianceClient
	logger       *zap.Logger
}

// NewRiskEvaluationService creates a new RiskEvaluationServiceImpl
func NewRiskEvaluationService(
	ruleRepo ports.RiskRuleRepository,
	alertRepo ports.RiskAlertRepository,
	exposureRepo ports.ExposureRepository,
	profileRepo ports.RiskProfileRepository,
	alertPub ports.AlertPublisher,
	metrics ports.MetricsCollector,
	compliance ports.ComplianceClient,
	logger *zap.Logger,
) *RiskEvaluationServiceImpl {
	return &RiskEvaluationServiceImpl{
		ruleRepo:     ruleRepo,
		alertRepo:    alertRepo,
		exposureRepo: exposureRepo,
		profileRepo:  profileRepo,
		alertPub:     alertPub,
		metrics:      metrics,
		compliance:   compliance,
		logger:       logger,
	}
}

// EvaluateMarketData evaluates market data against enabled risk rules
func (s *RiskEvaluationServiceImpl) EvaluateMarketData(
	ctx context.Context,
	data *domain.MarketData,
) ([]*domain.RiskAlert, error) {
	start := time.Now()
	s.logger.Debug("Evaluating market data",
		zap.String("symbol", data.Symbol),
		zap.String("source_id", data.SourceID))

	// Get enabled rules
	rules, err := s.ruleRepo.FindEnabled(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch enabled rules", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch rules: %w", err)
	}

	var alerts []*domain.RiskAlert

	for _, rule := range rules {
		// Check symbol filter
		if rule.SymbolFilter != "" && rule.SymbolFilter != data.Symbol {
			continue
		}

		// Check source filter
		if rule.SourceFilter != "" && rule.SourceFilter != data.SourceID {
			continue
		}

		// Evaluate the rule
		alert, err := s.evaluateRule(ctx, rule, data)
		if err != nil {
			s.logger.Warn("Failed to evaluate rule",
				zap.String("rule_id", rule.ID.String()),
				zap.Error(err))
			continue
		}

		if alert != nil {
			alerts = append(alerts, alert)
		}
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordLatency("risk_evaluation", time.Since(start).String())
		for _, alert := range alerts {
			s.metrics.IncrementAlerts(alert.Severity)
		}
	}

	return alerts, nil
}

// evaluateRule evaluates a single risk rule against market data
func (s *RiskEvaluationServiceImpl) evaluateRule(
	ctx context.Context,
	rule *domain.RiskRule,
	data *domain.MarketData,
) (*domain.RiskAlert, error) {
	var metricValue decimal.Decimal

	// Determine the metric value based on the rule metric
	switch rule.Metric {
	case "price":
		metricValue = data.Price
	case "volume":
		metricValue = data.Volume
	case "change_24h":
		metricValue = data.Change24h
	case "volatility":
		// Calculate simple volatility (high - low as percentage of price)
		if data.Price.IsPositive() {
			metricValue = data.High24h.Sub(data.Low24h).Div(data.Price).Mul(decimal.NewFromInt(100))
		}
	case "spread":
		// Would need order book data for this
		return nil, nil
	default:
		s.logger.Warn("Unknown metric type", zap.String("metric", rule.Metric))
		return nil, nil
	}

	// Evaluate the condition
	triggered := s.evaluateCondition(metricValue, rule.Operator, rule.Threshold)

	if !triggered {
		return nil, nil
	}

	// Create alert
	alert := &domain.RiskAlert{
		ID:        uuid.New(),
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		SourceID:  data.SourceID,
		Symbol:    data.Symbol,
		Metric:    rule.Metric,
		Value:     metricValue,
		Threshold: rule.Threshold,
		Severity:  rule.Severity,
		Message:   fmt.Sprintf("Risk rule '%s' triggered: %s %s %s (current: %s)",
			rule.Name, rule.Metric, rule.Operator, rule.Threshold, metricValue),
		Status:      domain.AlertStatusActive,
		Acknowledged: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save alert
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	// Publish alert
	if err := s.alertPub.PublishAlert(ctx, alert); err != nil {
		s.logger.Warn("Failed to publish alert", zap.Error(err))
	}

	s.logger.Info("Risk alert created",
		zap.String("rule_name", rule.Name),
		zap.String("symbol", data.Symbol),
		zap.String("severity", string(rule.Severity)))

	return alert, nil
}

// evaluateCondition evaluates a condition based on the operator
func (s *RiskEvaluationServiceImpl) evaluateCondition(
	value decimal.Decimal,
	operator domain.RuleOperator,
	threshold decimal.Decimal,
) bool {
	switch operator {
	case domain.RuleOperatorGreaterThan:
		return value.GreaterThan(threshold)
	case domain.RuleOperatorLessThan:
		return value.LessThan(threshold)
	case domain.RuleOperatorEqual:
		return value.Equal(threshold)
	case domain.RuleOperatorGreaterOrEqual:
		return value.GreaterThanOrEqual(threshold)
	case domain.RuleOperatorLessOrEqual:
		return value.LessThanOrEqual(threshold)
	case domain.RuleOperatorNotEqual:
		return !value.Equal(threshold)
	default:
		return false
	}
}

// EvaluateExposure evaluates an exposure against risk thresholds
func (s *RiskEvaluationServiceImpl) EvaluateExposure(
	ctx context.Context,
	exposure *domain.Exposure,
) ([]*domain.RiskAlert, error) {
	// Implementation would check exposure against thresholds
	// and create alerts if limits are exceeded
	return nil, nil
}

// AssessRisk performs a comprehensive risk assessment
func (s *RiskEvaluationServiceImpl) AssessRisk(
	ctx context.Context,
	req *ports.RiskAssessmentRequest,
) (*domain.RiskAssessment, error) {
	s.logger.Info("Performing risk assessment",
		zap.String("entity_id", req.EntityID),
		zap.String("entity_type", req.EntityType))

	// Get exposures for the entity
	exposures, err := s.exposureRepo.FindByEntity(ctx, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exposures: %w", err)
	}

	// Calculate total exposure
	totalExposure, err := s.exposureRepo.CalculateTotalExposure(ctx, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate exposure: %w", err)
	}

	// Get risk profile
	profile, err := s.profileRepo.FindByEntity(ctx, req.EntityID)
	if err != nil {
		profile = &domain.RiskProfile{
			EntityID:      req.EntityID,
			EntityType:    req.EntityType,
			TotalExposure: totalExposure,
			RiskScore:     decimal.Zero,
			RiskLevel:     domain.RiskLevelLow,
		}
	}

	// Calculate risk metrics
	metrics := s.calculateRiskMetrics(exposures, totalExposure, profile)

	// Get active alerts
	alerts, err := s.alertRepo.FindActive(ctx)
	if err != nil {
		s.logger.Warn("Failed to fetch active alerts", zap.Error(err))
		alerts = []*domain.RiskAlert{}
	}

	// Calculate overall risk score
	overallScore := s.calculateOverallScore(metrics, alerts)

	// Determine risk level
	riskLevel := s.determineRiskLevel(overallScore)

	// Update profile
	profile.RiskScore = overallScore
	profile.RiskLevel = riskLevel
	profile.AlertsCount = len(alerts)
	profile.LastUpdated = time.Now()
	s.profileRepo.Save(ctx, profile)

	assessment := &domain.RiskAssessment{
		Timestamp:    time.Now(),
		EntityID:     req.EntityID,
		EntityType:   req.EntityType,
		OverallScore: overallScore,
		RiskLevel:    riskLevel,
		Metrics:      metrics,
		Alerts:       alerts,
		Exposures:    exposures,
	}

	// Record metrics
	if s.metrics != nil {
		s.metrics.RecordAssessment(overallScore, riskLevel)
	}

	return assessment, nil
}

// calculateRiskMetrics calculates risk metrics from exposures
func (s *RiskEvaluationServiceImpl) calculateRiskMetrics(
	exposures []*domain.Exposure,
	totalExposure decimal.Decimal,
	profile *domain.RiskProfile,
) []domain.RiskMetric {
	metrics := []domain.RiskMetric{}

	// Total exposure metric
	metrics = append(metrics, domain.RiskMetric{
		Name:        "total_exposure",
		Value:       totalExposure,
		Threshold:   profile.MaxExposure,
		Status:      "normal",
		Description: "Total exposure across all positions",
	})

	// Position count metric
	metrics = append(metrics, domain.RiskMetric{
		Name:        "position_count",
		Value:       decimal.NewFromInt(int64(len(exposures))),
		Threshold:   decimal.NewFromInt(100),
		Status:      "normal",
		Description: "Number of open positions",
	})

	// Calculate concentration risk (largest position as percentage)
	if len(exposures) > 0 {
		maxNotional := decimal.Zero
		for _, e := range exposures {
			if e.Notional.GreaterThan(maxNotional) {
				maxNotional = e.Notional
			}
		}
		if totalExposure.IsPositive() {
			concentration := maxNotional.Div(totalExposure).Mul(decimal.NewFromInt(100))
			metrics = append(metrics, domain.RiskMetric{
				Name:        "concentration",
				Value:       concentration,
				Threshold:   decimal.NewFromInt(50), // 50% concentration limit
				Status:      "warning",
				Description: "Largest position as percentage of total exposure",
			})
		}
	}

	return metrics
}

// calculateOverallScore calculates the overall risk score
func (s *RiskEvaluationServiceImpl) calculateOverallScore(
	metrics []domain.RiskMetric,
	alerts []*domain.RiskAlert,
) decimal.Decimal {
	score := decimal.Zero

	// Factor in active alerts by severity
	for _, alert := range alerts {
		switch alert.Severity {
		case domain.RiskLevelCritical:
			score = score.Add(decimal.NewFromInt(25))
		case domain.RiskLevelHigh:
			score = score.Add(decimal.NewFromInt(15))
		case domain.RiskLevelMedium:
			score = score.Add(decimal.NewFromInt(10))
		case domain.RiskLevelLow:
			score = score.Add(decimal.NewFromInt(5))
		}
	}

	// Cap score at 100
	if score.GreaterThan(decimal.NewFromInt(100)) {
		score = decimal.NewFromInt(100)
	}

	return score
}

// determineRiskLevel determines the risk level based on the score
func (s *RiskEvaluationServiceImpl) determineRiskLevel(score decimal.Decimal) domain.RiskLevel {
	if score.GreaterThanOrEqual(decimal.NewFromInt(80)) {
		return domain.RiskLevelCritical
	}
	if score.GreaterThanOrEqual(decimal.NewFromInt(60)) {
		return domain.RiskLevelHigh
	}
	if score.GreaterThanOrEqual(decimal.NewFromInt(40)) {
		return domain.RiskLevelMedium
	}
	return domain.RiskLevelLow
}

// CalculateExposure calculates total exposure for an entity
func (s *RiskEvaluationServiceImpl) CalculateExposure(
	ctx context.Context,
	entityID string,
) (*domain.Exposure, error) {
	total, err := s.exposureRepo.CalculateTotalExposure(ctx, entityID)
	if err != nil {
		return nil, err
	}

	return &domain.Exposure{
		ID:          uuid.New(),
		EntityID:    entityID,
		Notional:    total,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// Ensure RiskEvaluationServiceImpl implements RiskEvaluationService
var _ ports.RiskEvaluationService = (*RiskEvaluationServiceImpl)(nil)
