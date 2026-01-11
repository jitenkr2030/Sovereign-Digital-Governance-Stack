package domain

import (
	"encoding/json"
	"time"
)

// Policy represents a regulatory policy rule
type Policy struct {
	ID               string            `json:"id" db:"id"`
	Name             string            `json:"name" db:"name"`
	Description      string            `json:"description" db:"description"`
	RuleSet          json.RawMessage   `json:"rule_set" db:"rule_set"`
	Version          int               `json:"version" db:"version"`
	Severity         PolicySeverity    `json:"severity" db:"severity"`
	TargetEntityType EntityType        `json:"target_entity_type" db:"target_entity_type"`
	TargetAction     ActionType        `json:"target_action" db:"target_action"`
	Scope            PolicyScope       `json:"scope" db:"scope"`
	Conditions       PolicyConditions  `json:"conditions" db:"conditions"`
	Effects          []PolicyEffect    `json:"effects" db:"effects"`
	Status           PolicyStatus      `json:"status" db:"status"`
	Priority         int               `json:"priority" db:"priority"`
	Tags             []string          `json:"tags" db:"tags"`
	Metadata         json.RawMessage   `json:"metadata" db:"metadata"`
	CreatedBy        string            `json:"created_by" db:"created_by"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" db:"updated_at"`
	ActivatedAt      *time.Time        `json:"activated_at" db:"activated_at"`
	DeactivatedAt    *time.Time        `json:"deactivated_at" db:"deactivated_at"`
}

// PolicySeverity represents the severity level of a policy violation
type PolicySeverity string

const (
	SeverityCritical PolicySeverity = "critical"
	SeverityHigh     PolicySeverity = "high"
	SeverityMedium   PolicySeverity = "medium"
	SeverityLow      PolicySeverity = "low"
	SeverityInfo     PolicySeverity = "info"
)

// EntityType represents the type of entity a policy applies to
type EntityType string

const (
	EntityTypeMiner   EntityType = "miner"
	EntityTypeExchange EntityType = "exchange"
	EntityTypeWallet  EntityType = "wallet"
	EntityTypeVASP    EntityType = "vasp"
	EntityTypeCASP    EntityType = "casp"
	EntityTypeAll     EntityType = "all"
)

// ActionType represents the type of action that can be regulated
type ActionType string

const (
	ActionTypeTransaction     ActionType = "transaction"
	ActionTypeWithdrawal      ActionType = "withdrawal"
	ActionTypeDeposit         ActionType = "deposit"
	ActionTypeTrading         ActionType = "trading"
	ActionTypeMining          ActionType = "mining"
	ActionTypeEnergyConsume   ActionType = "energy_consume"
	ActionTypeLicenseRenew    ActionType = "license_renew"
	ActionTypeReportSubmit    ActionType = "report_submit"
	ActionTypeComplianceCheck ActionType = "compliance_check"
)

// PolicyScope represents the scope of a policy
type PolicyScope string

const (
	ScopeGlobal    PolicyScope = "global"
	ScopeRegional  PolicyScope = "regional"
	ScopeEntity    PolicyScope = "entity"
	ScopeJurisdiction PolicyScope = "jurisdiction"
)

// PolicyStatus represents the status of a policy
type PolicyStatus string

const (
	PolicyStatusDraft     PolicyStatus = "draft"
	PolicyStatusActive    PolicyStatus = "active"
	PolicyStatusSuspended PolicyStatus = "suspended"
	PolicyStatusArchived  PolicyStatus = "archived"
)

// PolicyConditions represents the conditions under which a policy applies
type PolicyConditions struct {
	TimeRanges      []TimeRange      `json:"time_ranges,omitempty"`
	GeographicAreas []string         `json:"geographic_areas,omitempty"`
	EntityCategories []string        `json:"entity_categories,omitempty"`
	TransactionTypes []string        `json:"transaction_types,omitempty"`
	ValueThresholds ValueThresholds  `json:"value_thresholds,omitempty"`
	CustomConditions json.RawMessage `json:"custom_conditions,omitempty"`
}

// TimeRange represents a time range for policy conditions
type TimeRange struct {
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	Timezone  string `json:"timezone"`
	DaysOfWeek []int `json:"days_of_week"` // 1-7 (Monday-Sunday)
}

// ValueThresholds represents value thresholds for policy conditions
type ValueThresholds struct {
	MinValue      *float64 `json:"min_value,omitempty"`
	MaxValue      *float64 `json:"max_value,omitempty"`
	MaxDailyValue *float64 `json:"max_daily_value,omitempty"`
	MaxMonthlyValue *float64 `json:"max_monthly_value,omitempty"`
	Currency      string   `json:"currency,omitempty"`
}

// PolicyEffect represents the effect of a policy when conditions are met
type PolicyEffect struct {
	Type          EffectType       `json:"type"`
	Parameters    json.RawMessage `json:"parameters,omitempty"`
	Duration      *time.Duration  `json:"duration,omitempty"`
	Escalation    *Escalation     `json:"escalation,omitempty"`
	Notification  Notification    `json:"notification,omitempty"`
}

// EffectType represents the type of effect a policy can have
type EffectType string

const (
	EffectAllow    EffectType = "allow"
	EffectDeny     EffectType = "deny"
	EffectThrottle EffectType = "throttle"
	EffectAlert    EffectType = "alert"
	EffectLog      EffectType = "log"
	EffectFreeze   EffectType = "freeze"
	EffectRevoke   EffectType = "revoke"
	EffectFine     EffectType = "fine"
	EffectRequire  EffectType = "require_approval"
	EffectRateLimit EffectType = "rate_limit"
)

// Escalation defines escalation behavior for policy effects
type Escalation struct {
	MaxAttempts     int           `json:"max_attempts"`
	EscalateTo      string        `json:"escalate_to"`
	EscalationDelay time.Duration `json:"escalation_delay"`
	NotifyOnEscalate bool         `json:"notify_on_escalate"`
}

// Notification defines notification behavior for policy effects
type Notification struct {
	Channels     []string `json:"channels"` // email, sms, webhook, etc.
	Recipients   []string `json:"recipients"`
	TemplateID   string   `json:"template_id"`
	Priority     string   `json:"priority"` // low, normal, high, urgent
}

// PolicyDecision represents the result of policy evaluation
type PolicyDecision struct {
	PolicyID     string        `json:"policy_id"`
	PolicyName   string        `json:"policy_name"`
	Decision     DecisionResult `json:"decision"`
	Reason       string        `json:"reason"`
	AppliedEffects []PolicyEffect `json:"applied_effects"`
	EnforcementAction *EnforcementAction `json:"enforcement_action,omitempty"`
	Trace        []DecisionStep `json:"trace,omitempty"`
	EvaluatedAt  time.Time     `json:"evaluated_at"`
}

// DecisionResult represents the result of a policy decision
type DecisionResult string

const (
	DecisionAllowed  DecisionResult = "allowed"
	DecisionDenied   DecisionResult = "denied"
	DecisionDeferred DecisionResult = "deferred"
	DecisionError    DecisionResult = "error"
)

// DecisionStep represents a step in the policy evaluation trace
type DecisionStep struct {
	Step       string                 `json:"step"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
	DurationMs float64                `json:"duration_ms"`
}

// PolicyEvaluationContext represents the context for policy evaluation
type PolicyEvaluationContext struct {
	EntityID       string                 `json:"entity_id"`
	EntityType     EntityType             `json:"entity_type"`
	EntityCategory string                 `json:"entity_category"`
	Action         ActionType             `json:"action"`
	Transaction    *TransactionContext    `json:"transaction,omitempty"`
	EnergyContext  *EnergyContext         `json:"energy_context,omitempty"`
	Environment    map[string]interface{} `json:"environment"`
	Timestamp      time.Time              `json:"timestamp"`
	RequestID      string                 `json:"request_id"`
}

// TransactionContext represents transaction details for policy evaluation
type TransactionContext struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	Amount          float64  `json:"amount"`
	Currency        string   `json:"currency"`
	SourceWallet    string   `json:"source_wallet"`
	DestWallet      string   `json:"dest_wallet"`
	SourceExchange  string   `json:"source_exchange,omitempty"`
	DestExchange    string   `json:"dest_exchange,omitempty"`
	AssetType       string   `json:"asset_type"`
	Blockchain      string   `json:"blockchain"`
	Timestamp       time.Time `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// EnergyContext represents energy consumption details for policy evaluation
type EnergyContext struct {
	OperationID     string    `json:"operation_id"`
	OperationType   string    `json:"operation_type"`
	EnergyConsumed  float64   `json:"energy_consumed"` // in kWh
	PowerUsage      float64   `json:"power_usage"`     // in kW
	Duration        float64   `json:"duration"`        // in minutes
	Location        string    `json:"location"`
	GridRegion      string    `json:"grid_region"`
	Timestamp       time.Time `json:"timestamp"`
}
