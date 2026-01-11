package ports

import (
	"context"
	"time"

	"csic-platform/health-monitor/internal/core/domain"
)

// ServiceStatusRepository manages service status data
type ServiceStatusRepository interface {
	GetServiceStatus(ctx context.Context, serviceName string) (*domain.ServiceStatus, error)
	GetAllServiceStatuses(ctx context.Context) ([]*domain.ServiceStatus, error)
	UpsertServiceStatus(ctx context.Context, status *domain.ServiceStatus) error
	UpdateServiceStatus(ctx context.Context, serviceName string, status domain.ServiceStatus) error
	DeleteServiceStatus(ctx context.Context, serviceName string) error
	GetServicesByStatus(ctx context.Context, status string) ([]*domain.ServiceStatus, error)
	GetServiceUptime(ctx context.Context, serviceName string, since time.Time) (float64, error)
}

// AlertRuleRepository manages alert rules
type AlertRuleRepository interface {
	GetAlertRule(ctx context.Context, id string) (*domain.AlertRule, error)
	GetAllAlertRules(ctx context.Context) ([]*domain.AlertRule, error)
	GetAlertRulesByService(ctx context.Context, serviceName string) ([]*domain.AlertRule, error)
	GetEnabledAlertRules(ctx context.Context) ([]*domain.AlertRule, error)
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	UpdateAlertRule(ctx context.Context, id string, rule *domain.AlertRule) error
	DeleteAlertRule(ctx context.Context, id string) error
}

// AlertRepository manages alerts
type AlertRepository interface {
	GetAlert(ctx context.Context, id string) (*domain.Alert, error)
	GetAlerts(ctx context.Context, filter AlertFilter) ([]*domain.Alert, error)
	GetFiringAlerts(ctx context.Context) ([]*domain.Alert, error)
	GetAlertsByService(ctx context.Context, serviceName string, limit int) ([]*domain.Alert, error)
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	UpdateAlertStatus(ctx context.Context, id string, status string) error
	ResolveAlert(ctx context.Context, id string) error
	DeleteResolvedAlerts(ctx context.Context, olderThan time.Time) error
}

// OutageRepository manages outages
type OutageRepository interface {
	GetOutage(ctx context.Context, id string) (*domain.Outage, error)
	GetActiveOutages(ctx context.Context) ([]*domain.Outage, error)
	GetOutagesByService(ctx context.Context, serviceName string, limit int) ([]*domain.Outage, error)
	GetRecentOutages(ctx context.Context, since time.Time) ([]*domain.Outage, error)
	CreateOutage(ctx context.Context, outage *domain.Outage) error
	UpdateOutageStatus(ctx context.Context, id string, status string) error
	ResolveOutage(ctx context.Context, id string, rootCause string) error
}

// AlertFilter defines filters for querying alerts
type AlertFilter struct {
	ServiceName string
	Severity    string
	Status      string
	Since       time.Time
	Limit       int
}

// CachePort defines the interface for cache operations
type CachePort interface {
	// Heartbeat operations
	SetHeartbeat(ctx context.Context, serviceName string, heartbeat *domain.Heartbeat, ttl time.Duration) error
	GetHeartbeat(ctx context.Context, serviceName string) (*domain.Heartbeat, error)
	DeleteHeartbeat(ctx context.Context, serviceName string) error
	GetAllHeartbeats(ctx context.Context) ([]*domain.Heartbeat, error)

	// Service state operations
	SetServiceState(ctx context.Context, serviceName string, state string, ttl time.Duration) error
	GetServiceState(ctx context.Context, serviceName string) (string, error)

	// Alert cooldown operations
	SetAlertCooldown(ctx context.Context, ruleID string, until time.Time) error
	IsAlertInCooldown(ctx context.Context, ruleID string) (bool, error)

	// Health state operations
	GetHealthState(ctx context.Context, serviceName string) (*domain.ServiceStatus, error)
	SetHealthState(ctx context.Context, status *domain.ServiceStatus, ttl time.Duration) error
}

// MessagingPort defines the interface for messaging operations
type MessagingPort interface {
	// Consumer operations
	StartConsuming(ctx context.Context, topics []string) error
	SetMessageHandler(handler MessageHandler) error

	// Producer operations
	PublishHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error
	PublishAlert(ctx context.Context, alert *domain.Alert) error
	PublishOutage(ctx context.Context, outage *domain.Outage) error
	PublishHealthEvent(ctx context.Context, event *domain.HealthCheck) error
}

// MessageHandler defines the interface for handling messages
type MessageHandler interface {
	HandleHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error
	HandleHealthCheck(ctx context.Context, check *domain.HealthCheck) error
	HandleStateUpdate(ctx context.Context, update map[string]interface{}) error
}
