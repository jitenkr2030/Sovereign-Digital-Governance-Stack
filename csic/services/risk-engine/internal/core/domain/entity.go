package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RiskLevel represents the severity level of a risk
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// RiskRule defines a rule for evaluating risk
type RiskRule struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Metric      string          `json:"metric" db:"metric"`
	Operator    RuleOperator    `json:"operator" db:"operator"`
	Threshold   decimal.Decimal `json:"threshold" db:"threshold"`
	Severity    RiskLevel       `json:"severity" db:"severity"`
	Enabled     bool            `json:"enabled" db:"enabled"`
	SourceFilter string         `json:"source_filter,omitempty" db:"source_filter"`
	SymbolFilter string         `json:"symbol_filter,omitempty" db:"symbol_filter"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// RuleOperator defines the comparison operator for a rule
type RuleOperator string

const (
	RuleOperatorGreaterThan    RuleOperator = "GT"
	RuleOperatorLessThan       RuleOperator = "LT"
	RuleOperatorEqual          RuleOperator = "EQ"
	RuleOperatorGreaterOrEqual RuleOperator = "GTE"
	RuleOperatorLessOrEqual    RuleOperator = "LTE"
	RuleOperatorNotEqual       RuleOperator = "NEQ"
	RuleOperatorIn             RuleOperator = "IN"
	RuleOperatorBetween        RuleOperator = "BETWEEN"
)

// RiskAlert represents an alert triggered by a risk rule
type RiskAlert struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	RuleID      uuid.UUID       `json:"rule_id" db:"rule_id"`
	RuleName    string          `json:"rule_name" db:"rule_name"`
	SourceID    string          `json:"source_id" db:"source_id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	Metric      string          `json:"metric" db:"metric"`
	Value       decimal.Decimal `json:"value" db:"value"`
	Threshold   decimal.Decimal `json:"threshold" db:"threshold"`
	Severity    RiskLevel       `json:"severity" db:"severity"`
	Message     string          `json:"message" db:"message"`
	Status      AlertStatus     `json:"status" db:"status"`
	Acknowledged bool           `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy string       `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt *time.Time   `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "ACTIVE"
	AlertStatusAcknowledged AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved  AlertStatus = "RESOLVED"
	AlertStatusDismissed AlertStatus = "DISMISSED"
)

// RiskProfile represents the risk profile for an entity
type RiskProfile struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	EntityID      string          `json:"entity_id" db:"entity_id"`
	EntityType    string          `json:"entity_type" db:"entity_type"`
	TotalExposure decimal.Decimal `json:"total_exposure" db:"total_exposure"`
	MaxExposure   decimal.Decimal `json:"max_exposure" db:"max_exposure"`
	RiskScore     decimal.Decimal `json:"risk_score" db:"risk_score"`
	RiskLevel     RiskLevel       `json:"risk_level" db:"risk_level"`
	AlertsCount   int             `json:"alerts_count" db:"alerts_count"`
	LastUpdated   time.Time       `json:"last_updated" db:"last_updated"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// Exposure represents a position or exposure
type Exposure struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	EntityID    string          `json:"entity_id" db:"entity_id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	Position    decimal.Decimal `json:"position" db:"position"`
	Notional    decimal.Decimal `json:"notional" db:"notional"`
	EntryPrice  decimal.Decimal `json:"entry_price" db:"entry_price"`
	MarkPrice   decimal.Decimal `json:"mark_price" db:"mark_price"`
	PnL         decimal.Decimal `json:"pnl" db:"pnl"`
	Side        string          `json:"side" db:"side"`
	Leverage    decimal.Decimal `json:"leverage" db:"leverage"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// RiskAssessment represents a complete risk assessment result
type RiskAssessment struct {
	Timestamp    time.Time          `json:"timestamp"`
	EntityID     string             `json:"entity_id"`
	EntityType   string             `json:"entity_type"`
	OverallScore decimal.Decimal    `json:"overall_score"`
	RiskLevel    RiskLevel          `json:"risk_level"`
	Metrics      []RiskMetric       `json:"metrics"`
	Alerts       []RiskAlert        `json:"alerts"`
	Exposures    []Exposure         `json:"exposures"`
}

// RiskMetric represents a single risk metric
type RiskMetric struct {
	Name        string          `json:"name"`
	Value       decimal.Decimal `json:"value"`
	Threshold   decimal.Decimal `json:"threshold"`
	Status      string          `json:"status"`
	Description string          `json:"description"`
}

// RiskThresholdConfig defines global risk thresholds
type RiskThresholdConfig struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	MaxPositionSize decimal.Decimal `json:"max_position_size" db:"max_position_size"`
	MaxExposure     decimal.Decimal `json:"max_exposure" db:"max_exposure"`
	MaxLeverage     decimal.Decimal `json:"max_leverage" db:"max_leverage"`
	MinMargin       decimal.Decimal `json:"min_margin" db:"min_margin"`
	MaxVolatility   decimal.Decimal `json:"max_volatility" db:"max_volatility"`
	MaxDrawdown     decimal.Decimal `json:"max_drawdown" db:"max_drawdown"`
	Enabled         bool            `json:"enabled" db:"enabled"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// RiskStats contains aggregated risk statistics
type RiskStats struct {
	TotalAlerts         int64           `json:"total_alerts"`
	ActiveAlerts        int64           `json:"active_alerts"`
	CriticalAlerts      int64           `json:"critical_alerts"`
	HighAlerts          int64           `json:"high_alerts"`
	MediumAlerts        int64           `json:"medium_alerts"`
	LowAlerts           int64           `json:"low_alerts"`
	ResolvedToday       int64           `json:"resolved_today"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
	LastAssessment      *time.Time      `json:"last_assessment,omitempty"`
}
