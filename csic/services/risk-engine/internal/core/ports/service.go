package ports

import (
	"context"

	"github.com/csic-platform/services/risk-engine/internal/core/domain"
)

// RiskRuleService defines the business logic for managing risk rules
type RiskRuleService interface {
	// CreateRule creates a new risk rule
	CreateRule(ctx context.Context, req *CreateRuleRequest) (*CreateRuleResponse, error)

	// UpdateRule updates an existing risk rule
	UpdateRule(ctx context.Context, req *UpdateRuleRequest) error

	// DeleteRule deletes a risk rule
	DeleteRule(ctx context.Context, id string) error

	// GetRule retrieves a specific risk rule
	GetRule(ctx context.Context, id string) (*domain.RiskRule, error)

	// ListRules lists all risk rules with optional filtering
	ListRules(ctx context.Context, req *ListRulesRequest) (*ListRulesResponse, error)

	// EnableRule enables a risk rule
	EnableRule(ctx context.Context, id string) error

	// DisableRule disables a risk rule
	DisableRule(ctx context.Context, id string) error
}

// RiskEvaluationService defines the business logic for risk evaluation
type RiskEvaluationService interface {
	// EvaluateMarketData evaluates market data against risk rules
	EvaluateMarketData(ctx context.Context, data *domain.MarketData) ([]*domain.RiskAlert, error)

	// EvaluateExposure evaluates an exposure against risk thresholds
	EvaluateExposure(ctx context.Context, exposure *domain.Exposure) ([]*domain.RiskAlert, error)

	// AssessRisk performs a comprehensive risk assessment
	AssessRisk(ctx context.Context, req *RiskAssessmentRequest) (*domain.RiskAssessment, error)

	// CalculateExposure calculates total exposure for an entity
	CalculateExposure(ctx context.Context, entityID string) (*domain.Exposure, error)
}

// AlertService defines the business logic for managing alerts
type AlertService interface {
	// GetAlert retrieves a specific alert
	GetAlert(ctx context.Context, id string) (*domain.RiskAlert, error)

	// ListAlerts lists alerts with filtering
	ListAlerts(ctx context.Context, req *ListAlertsRequest) (*ListAlertsResponse, error)

	// GetActiveAlerts retrieves all active alerts
	GetActiveAlerts(ctx context.Context) ([]*domain.RiskAlert, error)

	// AcknowledgeAlert acknowledges an alert
	AcknowledgeAlert(ctx context.Context, id string, userID string) error

	// ResolveAlert resolves an alert
	ResolveAlert(ctx context.Context, id string) error

	// DismissAlert dismisses an alert
	DismissAlert(ctx context.Context, id string) error

	// GetStats returns risk statistics
	GetStats(ctx context.Context) (*domain.RiskStats, error)
}

// ThresholdService defines the business logic for managing thresholds
type ThresholdService interface {
	// GetConfig retrieves the current threshold configuration
	GetConfig(ctx context.Context) (*domain.RiskThresholdConfig, error)

	// UpdateConfig updates the threshold configuration
	UpdateConfig(ctx context.Context, req *UpdateThresholdRequest) error
}

// Request/Response types for RiskRuleService

type CreateRuleRequest struct {
	Name         string                `json:"name" validate:"required"`
	Description  string                `json:"description"`
	Metric       string                `json:"metric" validate:"required"`
	Operator     domain.RuleOperator   `json:"operator" validate:"required"`
	Threshold    string                `json:"threshold" validate:"required"`
	Severity     domain.RiskLevel      `json:"severity" validate:"required"`
	SourceFilter string                `json:"source_filter,omitempty"`
	SymbolFilter string                `json:"symbol_filter,omitempty"`
}

type CreateRuleResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type UpdateRuleRequest struct {
	ID          string  `json:"id" validate:"required"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Operator    *string `json:"operator,omitempty"`
	Threshold   *string `json:"threshold,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	SourceFilter *string `json:"source_filter,omitempty"`
	SymbolFilter *string `json:"symbol_filter,omitempty"`
}

type ListRulesRequest struct {
	EnabledOnly  bool   `json:"enabled_only,omitempty"`
	Metric       string `json:"metric,omitempty"`
	Severity     string `json:"severity,omitempty"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

type ListRulesResponse struct {
	Rules     []*domain.RiskRule `json:"rules"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

// Request/Response types for RiskEvaluationService

type RiskAssessmentRequest struct {
	EntityID   string `json:"entity_id" validate:"required"`
	EntityType string `json:"entity_type" validate:"required"`
}

// Request/Response types for ThresholdService

type UpdateThresholdRequest struct {
	MaxPositionSize *string `json:"max_position_size,omitempty"`
	MaxExposure     *string `json:"max_exposure,omitempty"`
	MaxLeverage     *string `json:"max_leverage,omitempty"`
	MinMargin       *string `json:"min_margin,omitempty"`
	MaxVolatility   *string `json:"max_volatility,omitempty"`
	MaxDrawdown     *string `json:"max_drawdown,omitempty"`
}
