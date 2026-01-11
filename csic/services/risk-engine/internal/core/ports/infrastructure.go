package ports

import (
	"context"

	"github.com/csic-platform/services/risk-engine/internal/core/domain"
)

// MarketDataConsumer defines the interface for consuming market data from Kafka
type MarketDataConsumer interface {
	// Start begins consuming market data messages
	Start(ctx context.Context) error

	// Stop stops the consumer
	Stop() error

	// SetHandler sets the handler function for processing market data
	SetHandler(handler func(data *domain.MarketData) error)
}

// AlertPublisher defines the interface for publishing alerts
type AlertPublisher interface {
	// PublishAlert publishes a risk alert to the message broker
	PublishAlert(ctx context.Context, alert *domain.RiskAlert) error

	// PublishBatch publishes multiple alerts in a batch
	PublishBatch(ctx context.Context, alerts []*domain.RiskAlert) error

	// Close closes the publisher
	Close() error
}

// ComplianceClient defines the interface for communicating with the Compliance module
type ComplianceClient interface {
	// GetGlobalLimits retrieves global risk limits from compliance
	GetGlobalLimits(ctx context.Context) (*ComplianceLimits, error)

	// GetRestrictedAssets retrieves list of restricted assets
	GetRestrictedAssets(ctx context.Context) ([]string, error)

	// CheckCompliance checks if an entity is compliant
	CheckCompliance(ctx context.Context, entityID string) (*ComplianceStatus, error)
}

// ComplianceLimits represents risk limits from the compliance module
type ComplianceLimits struct {
	MaxPositionSize decimal.Decimal
	MaxExposure     decimal.Decimal
	MaxLeverage     decimal.Decimal
	MinMargin       decimal.Decimal
}

// ComplianceStatus represents the compliance status of an entity
type ComplianceStatus struct {
	EntityID   string `json:"entity_id"`
	IsCompliant bool  `json:"is_compliant"`
	Violations []string `json:"violations"`
	LastCheck  string `json:"last_check"`
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	// SendAlertNotification sends a notification for a risk alert
	SendAlertNotification(ctx context.Context, alert *domain.RiskAlert) error

	// SendBulkNotification sends notifications for multiple alerts
	SendBulkNotification(ctx context.Context, alerts []*domain.RiskAlert) error

	// Close closes the notification service
	Close() error
}

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	// IncrementAlerts increments the alert counter
	IncrementAlerts(severity domain.RiskLevel)

	// RecordAssessment records a risk assessment
	RecordAssessment(score decimal.Decimal, level domain.RiskLevel)

	// RecordLatency records processing latency
	RecordLatency(component string, latency string)

	// SetGauge sets a gauge value
	SetGauge(name string, value float64)

	// Close closes the metrics collector
	Close() error
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	// Check checks the health of the service
	Check(ctx context.Context) (*HealthStatus, error)
}

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp string            `json:"timestamp"`
}
