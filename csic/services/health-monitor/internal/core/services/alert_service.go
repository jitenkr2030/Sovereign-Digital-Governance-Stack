package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
	"csic-platform/health-monitor/pkg/metrics"
)

// AlertService manages alerts
type AlertService interface {
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) (*domain.AlertRule, error)
	GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error)
	GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error)
	UpdateAlertRule(ctx context.Context, id string, rule *domain.AlertRule) (*domain.AlertRule, error)
	DeleteAlertRule(ctx context.Context, id string) error
	GetAlerts(ctx context.Context, filter ports.AlertFilter) ([]*domain.Alert, error)
	GetFiringAlerts(ctx context.Context) ([]*domain.Alert, error)
	ResolveAlert(ctx context.Context, id string) error
	EvaluateAlerts(ctx context.Context, serviceName string, metrics map[string]float64) ([]*domain.Alert, error)
	StartAlertEvaluator(logger *zap.Logger)
	GetActiveOutages(ctx context.Context) ([]*domain.Outage, error)
	GetOutages(ctx context.Context, serviceName string, limit int) ([]*domain.Outage, error)
	CreateOutage(ctx context.Context, outage *domain.Outage) (*domain.Outage, error)
	ResolveOutage(ctx context.Context, id string, rootCause string) error
}

// AlertServiceService implements AlertService
type AlertServiceService struct {
	repositories  ports.Repositories
	cachePort     ports.CachePort
	messagingPort ports.MessagingPort
	logger        *zap.Logger
	metrics       *metrics.MetricsCollector
}

// NewAlertService creates a new alert service
func NewAlertService(
	repositories ports.Repositories,
	cachePort ports.CachePort,
	messagingPort ports.MessagingPort,
	logger *zap.Logger,
	metricsCollector *metrics.MetricsCollector,
) AlertService {
	return &AlertServiceService{
		repositories:  repositories,
		cachePort:     cachePort,
		messagingPort: messagingPort,
		logger:        logger,
		metrics:       metricsCollector,
	}
}

// CreateAlertRule creates a new alert rule
func (s *AlertServiceService) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) (*domain.AlertRule, error) {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	rule.Enabled = true

	if err := s.repositories.AlertRuleRepository.CreateAlertRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	s.logger.Info("Created alert rule",
		logger.String("rule_id", rule.ID),
		logger.String("name", rule.Name),
		logger.String("service", rule.ServiceName),
	)

	return rule, nil
}

// GetAlertRules gets all alert rules
func (s *AlertServiceService) GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error) {
	return s.repositories.AlertRuleRepository.GetAllAlertRules(ctx)
}

// GetAlertRule gets an alert rule by ID
func (s *AlertServiceService) GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error) {
	return s.repositories.AlertRuleRepository.GetAlertRule(ctx, id)
}

// UpdateAlertRule updates an alert rule
func (s *AlertServiceService) UpdateAlertRule(ctx context.Context, id string, rule *domain.AlertRule) (*domain.AlertRule, error) {
	rule.UpdatedAt = time.Now()

	if err := s.repositories.AlertRuleRepository.UpdateAlertRule(ctx, id, rule); err != nil {
		return nil, fmt.Errorf("failed to update alert rule: %w", err)
	}

	return rule, nil
}

// DeleteAlertRule deletes an alert rule
func (s *AlertServiceService) DeleteAlertRule(ctx context.Context, id string) error {
	if err := s.repositories.AlertRuleRepository.DeleteAlertRule(ctx, id); err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	s.logger.Info("Deleted alert rule", logger.String("rule_id", id))

	return nil
}

// GetAlerts gets alerts based on filter
func (s *AlertServiceService) GetAlerts(ctx context.Context, filter ports.AlertFilter) ([]*domain.Alert, error) {
	return s.repositories.AlertRepository.GetAlerts(ctx, filter)
}

// GetFiringAlerts gets all currently firing alerts
func (s *AlertServiceService) GetFiringAlerts(ctx context.Context) ([]*domain.Alert, error) {
	return s.repositories.AlertRepository.GetFiringAlerts(ctx)
}

// ResolveAlert resolves an alert
func (s *AlertServiceService) ResolveAlert(ctx context.Context, id string) error {
	if err := s.repositories.AlertRepository.ResolveAlert(ctx, id); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	s.logger.Info("Resolved alert", logger.String("alert_id", id))

	return nil
}

// EvaluateAlerts evaluates metrics against alert rules
func (s *AlertServiceService) EvaluateAlerts(ctx context.Context, serviceName string, metrics map[string]float64) ([]*domain.Alert, error) {
	// Get enabled alert rules for this service
	rules, err := s.repositories.AlertRuleRepository.GetAlertRulesByService(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rules: %w", err)
	}

	var alerts []*domain.Alert

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check if alert is in cooldown
		inCooldown, err := s.cachePort.IsAlertInCooldown(ctx, rule.ID)
		if err == nil && inCooldown {
			continue
		}

		// Evaluate the rule
		alert := s.evaluateRule(rule, metrics)
		if alert != nil {
			alerts = append(alerts, alert)

			// Save the alert
			if err := s.repositories.AlertRepository.CreateAlert(ctx, alert); err != nil {
				s.logger.Warn("Failed to create alert", logger.Error(err))
				continue
			}

			// Set cooldown
			if rule.Cooldown > 0 {
				until := time.Now().Add(time.Duration(rule.Cooldown) * time.Second)
				if err := s.cachePort.SetAlertCooldown(ctx, rule.ID, until); err != nil {
					s.logger.Warn("Failed to set alert cooldown", logger.Error(err))
				}
			}

			// Publish alert
			if err := s.messagingPort.PublishAlert(ctx, alert); err != nil {
				s.logger.Warn("Failed to publish alert", logger.Error(err))
			}

			s.metrics.RecordAlertFired(alert.Severity, alert.ServiceName)

			s.logger.Warn("Alert fired",
				logger.String("rule_id", rule.ID),
				logger.String("rule_name", rule.Name),
				logger.String("service", serviceName),
				logger.String("severity", rule.Severity),
			)
		}
	}

	return alerts, nil
}

// evaluateRule evaluates a single alert rule
func (s *AlertServiceService) evaluateRule(rule *domain.AlertRule, metrics map[string]float64) *domain.Alert {
	// Parse the condition (e.g., "cpu > 80")
	parts := strings.Split(rule.Condition, " ")
	if len(parts) != 3 {
		return nil
	}

	metricName := parts[0]
	operator := parts[1]
	threshold := rule.Threshold

	// Get the metric value
	value, ok := metrics[metricName]
	if !ok {
		return nil
	}

	// Evaluate the condition
	triggered := false
	switch operator {
	case ">":
		triggered = value > threshold
	case ">=":
		triggered = value >= threshold
	case "<":
		triggered = value < threshold
	case "<=":
		triggered = value <= threshold
	case "==":
		triggered = value == threshold
	case "!=":
		triggered = value != threshold
	}

	if !triggered {
		return nil
	}

	// Create the alert
	alert := &domain.Alert{
		ID:           uuid.New().String(),
		RuleID:       rule.ID,
		ServiceName:  rule.ServiceName,
		Severity:     rule.Severity,
		Condition:    rule.Condition,
		CurrentValue: value,
		Threshold:    threshold,
		Status:       domain.AlertStatusFiring,
		Message:      fmt.Sprintf("%s is %s (threshold: %s %s)", metricName, strconv.FormatFloat(value, 'f', 2, 64), operator, strconv.FormatFloat(threshold, 'f', 2, 64)),
		FiredAt:      time.Now(),
		Metadata:     domain.Metadata{"rule_name": rule.Name},
	}

	return alert
}

// StartAlertEvaluator starts the background alert evaluator
func (s *AlertServiceService) StartAlertEvaluator(logger *zap.Logger) {
	go func() {
		// Evaluate alerts every 30 seconds
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runAlertEvaluation()
			}
		}
	}()

	logger.Info("Alert evaluator started")
}

// runAlertEvaluation runs alert evaluation on all services
func (s *AlertServiceService) runAlertEvaluation() {
	ctx := context.Background()

	// Get all service statuses
	statuses, err := s.repositories.ServiceStatusRepository.GetAllServiceStatuses(ctx)
	if err != nil {
		s.logger.Error("Failed to get service statuses for alert evaluation", logger.Error(err))
		return
	}

	for _, status := range statuses {
		// Collect metrics
		metrics := map[string]float64{
			"cpu_percent":    status.CPUUsage,
			"memory_percent": status.MemoryUsage,
			"disk_percent":   status.DiskUsage,
			"latency_ms":     status.ResponseTime,
			"error_rate":     status.ErrorRate,
		}

		// Evaluate alerts
		if _, err := s.EvaluateAlerts(ctx, status.Name, metrics); err != nil {
			s.logger.Warn("Failed to evaluate alerts for service",
				logger.String("service", status.Name),
				logger.Error(err),
			)
		}
	}
}

// GetActiveOutages gets all active outages
func (s *AlertServiceService) GetActiveOutages(ctx context.Context) ([]*domain.Outage, error) {
	return s.repositories.OutageRepository.GetActiveOutages(ctx)
}

// GetOutages gets outages for a specific service
func (s *AlertServiceService) GetOutages(ctx context.Context, serviceName string, limit int) ([]*domain.Outage, error) {
	return s.repositories.OutageRepository.GetOutagesByService(ctx, serviceName, limit)
}

// CreateOutage creates a new outage
func (s *AlertServiceService) CreateOutage(ctx context.Context, outage *domain.Outage) (*domain.Outage, error) {
	outage.ID = uuid.New().String()
	outage.CreatedAt = time.Now()
	outage.UpdatedAt = time.Now()

	if err := s.repositories.OutageRepository.CreateOutage(ctx, outage); err != nil {
		return nil, fmt.Errorf("failed to create outage: %w", err)
	}

	s.logger.Warn("Outage detected",
		logger.String("outage_id", outage.ID),
		logger.String("service", outage.ServiceName),
		logger.String("severity", outage.Severity),
	)

	return outage, nil
}

// ResolveOutage resolves an outage
func (s *AlertServiceService) ResolveOutage(ctx context.Context, id string, rootCause string) error {
	if err := s.repositories.OutageRepository.ResolveOutage(ctx, id, rootCause); err != nil {
		return fmt.Errorf("failed to resolve outage: %w", err)
	}

	s.logger.Info("Outage resolved",
		logger.String("outage_id", id),
		logger.String("root_cause", rootCause),
	)

	return nil
}
