package ports

import (
	"context"

	"github.com/csic-platform/services/risk-engine/internal/core/domain"
)

// RiskRuleRepository defines the interface for managing risk rules
type RiskRuleRepository interface {
	// FindByID retrieves a risk rule by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.RiskRule, error)

	// FindAll retrieves all risk rules
	FindAll(ctx context.Context) ([]*domain.RiskRule, error)

	// FindEnabled retrieves all enabled risk rules
	FindEnabled(ctx context.Context) ([]*domain.RiskRule, error)

	// FindByMetric retrieves risk rules for a specific metric
	FindByMetric(ctx context.Context, metric string) ([]*domain.RiskRule, error)

	// Save creates or updates a risk rule
	Save(ctx context.Context, rule *domain.RiskRule) error

	// Delete removes a risk rule
	Delete(ctx context.Context, id string) error
}

// RiskAlertRepository defines the interface for managing risk alerts
type RiskAlertRepository interface {
	// FindByID retrieves a risk alert by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.RiskAlert, error)

	// FindAll retrieves all risk alerts with optional filtering
	FindAll(ctx context.Context, req *ListAlertsRequest) (*ListAlertsResponse, error)

	// FindActive retrieves all active (unresolved) alerts
	FindActive(ctx context.Context) ([]*domain.RiskAlert, error)

	// FindByRule retrieves alerts for a specific rule
	FindByRule(ctx context.Context, ruleID string) ([]*domain.RiskAlert, error)

	// FindBySource retrieves alerts for a specific data source
	FindBySource(ctx context.Context, sourceID string) ([]*domain.RiskAlert, error)

	// Create creates a new risk alert
	Create(ctx context.Context, alert *domain.RiskAlert) error

	// Update updates an existing risk alert
	Update(ctx context.Context, alert *domain.RiskAlert) error

	// Acknowledge marks an alert as acknowledged
	Acknowledge(ctx context.Context, id string, acknowledgedBy string) error

	// Resolve marks an alert as resolved
	Resolve(ctx context.Context, id string) error

	// Dismiss marks an alert as dismissed
	Dismiss(ctx context.Context, id string) error

	// CountByStatus counts alerts by status
	CountByStatus(ctx context.Context) (*domain.RiskStats, error)
}

// ExposureRepository defines the interface for managing exposures
type ExposureRepository interface {
	// FindByID retrieves an exposure by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.Exposure, error)

	// FindByEntity retrieves all exposures for an entity
	FindByEntity(ctx context.Context, entityID string) ([]*domain.Exposure, error)

	// FindBySymbol retrieves exposures for a specific symbol
	FindBySymbol(ctx context.Context, symbol string) ([]*domain.Exposure, error)

	// Save creates or updates an exposure
	Save(ctx context.Context, exposure *domain.Exposure) error

	// Delete removes an exposure
	Delete(ctx context.Context, id string) error

	// CalculateTotalExposure calculates total exposure for an entity
	CalculateTotalExposure(ctx context.Context, entityID string) (domain.Decimal, error)
}

// RiskProfileRepository defines the interface for managing risk profiles
type RiskProfileRepository interface {
	// FindByID retrieves a risk profile by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.RiskProfile, error)

	// FindByEntity retrieves a risk profile for an entity
	FindByEntity(ctx context.Context, entityID string) (*domain.RiskProfile, error)

	// Save creates or updates a risk profile
	Save(ctx context.Context, profile *domain.RiskProfile) error

	// UpdateScore updates the risk score for a profile
	UpdateScore(ctx context.Context, entityID string, score domain.Decimal, level domain.RiskLevel) error
}

// ThresholdConfigRepository defines the interface for managing threshold configurations
type ThresholdConfigRepository interface {
	// Get retrieves the current threshold configuration
	Get(ctx context.Context) (*domain.RiskThresholdConfig, error)

	// Save creates or updates the threshold configuration
	Save(ctx context.Context, config *domain.RiskThresholdConfig) error
}

// Request/Response types for AlertRepository

type ListAlertsRequest struct {
	Status      *domain.AlertStatus `json:"status,omitempty"`
	Severity    *domain.RiskLevel   `json:"severity,omitempty"`
	SourceID    string              `json:"source_id,omitempty"`
	RuleID      string              `json:"rule_id,omitempty"`
	StartDate   string              `json:"start_date,omitempty"`
	EndDate     string              `json:"end_date,omitempty"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}

type ListAlertsResponse struct {
	Alerts     []*domain.RiskAlert `json:"alerts"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}
