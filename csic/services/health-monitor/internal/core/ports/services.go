package ports

import (
	"context"
	"time"

	"csic-platform/health-monitor/internal/core/domain"
)

// MonitorService defines the interface for monitoring services
type MonitorService interface {
	// Heartbeat management
	ProcessHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error
	GetServiceStatus(ctx context.Context, serviceName string) (*domain.ServiceStatus, error)
	GetAllServiceStatuses(ctx context.Context) ([]*domain.ServiceStatus, error)
	GetHealthSummary(ctx context.Context) (*domain.HealthSummary, error)

	// Health checks
	PerformHealthCheck(ctx context.Context, serviceName string, endpoint string, timeout time.Duration) (*domain.HealthCheck, error)
	StartHealthCheckMonitor(logger interface{})
	StartHeartbeatConsumer(logger interface{})

	// Service registration
	RegisterService(ctx context.Context, req *domain.RegisterServiceRequest) (*domain.RegisterServiceResponse, error)
	GetRegisteredServices(ctx context.Context) ([]string, error)

	// Metrics
	GetServiceMetrics(ctx context.Context, serviceName string, since time.Time) (map[string]interface{}, error)
}

// AlertService defines the interface for alert services
type AlertService interface {
	// Alert rules
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) (*domain.AlertRule, error)
	GetAlertRules(ctx context.Context) ([]*domain.AlertRule, error)
	GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error)
	UpdateAlertRule(ctx context.Context, id string, rule *domain.AlertRule) (*domain.AlertRule, error)
	DeleteAlertRule(ctx context.Context, id string) error

	// Alerts
	GetAlerts(ctx context.Context, filter AlertFilter) ([]*domain.Alert, error)
	GetFiringAlerts(ctx context.Context) ([]*domain.Alert, error)
	ResolveAlert(ctx context.Context, id string) error

	// Alert evaluation
	EvaluateAlerts(ctx context.Context, serviceName string, metrics map[string]float64) ([]*domain.Alert, error)
	StartAlertEvaluator(logger interface{})

	// Outages
	GetActiveOutages(ctx context.Context) ([]*domain.Outage, error)
	GetOutages(ctx context.Context, serviceName string, limit int) ([]*domain.Outage, error)
	CreateOutage(ctx context.Context, outage *domain.Outage) (*domain.Outage, error)
	ResolveOutage(ctx context.Context, id string, rootCause string) error
}

// HealthSummaryService generates health summaries
type HealthSummaryService interface {
	GenerateSummary(ctx context.Context) (*domain.HealthSummary, error)
	GetServiceHealth(serviceName string) (*domain.ServiceHealth, error)
}
