package models

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents a policy rule for economic intervention
type Policy struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	Name             string          `json:"name" db:"name"`
	Description      string          `json:"description" db:"description"`
	Type             PolicyType      `json:"type" db:"type"`
	Category         string          `json:"category" db:"category"`
	Rules            PolicyRules     `json:"rules" db:"rules"`
	Conditions       string          `json:"conditions" db:"conditions"` // OPA Rego policy
	EnforcementLevel EnforcementLevel `json:"enforcement_level" db:"enforcement_level"`
	Scope            PolicyScope     `json:"scope" db:"scope"`
	Priority         int             `json:"priority" db:"priority"`
	Status           PolicyStatus    `json:"status" db:"status"`
	Version          string          `json:"version" db:"version"`
	CreatedBy        uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
	ActivatedAt      *time.Time      `json:"activated_at,omitempty" db:"activated_at"`
}

// PolicyType defines the type of policy
type PolicyType string

const (
	PolicyTypeInflationControl     PolicyType = "inflation_control"
	PolicyTypeSubsidyManagement    PolicyType = "subsidy_management"
	PolicyTypeSectorIntervention   PolicyType = "sector_intervention"
	PolicyTypeRegionalRestriction  PolicyType = "regional_restriction"
	PolicyTypeEmergencyLock        PolicyType = "emergency_lock"
	PolicyTypeTradeControl         PolicyType = "trade_control"
	PolicyTypePriceControl         PolicyType = "price_control"
)

// EnforcementLevel defines how strictly the policy is enforced
type EnforcementLevel string

const (
	EnforcementLevelAdvisory EnforcementLevel = "ADVISORY"
	EnforcementLevelMandatory EnforcementLevel = "MANDATORY"
	EnforcementLevelBlocking EnforcementLevel  = "BLOCKING"
)

// PolicyStatus defines the current status of a policy
type PolicyStatus string

const (
	PolicyStatusDraft     PolicyStatus = "draft"
	PolicyStatusActive    PolicyStatus = "active"
	PolicyStatusSuspended PolicyStatus = "suspended"
	PolicyStatusArchived  PolicyStatus = "archived"
)

// PolicyScope defines where the policy applies
type PolicyScope struct {
	Geographic   []string          `json:"geographic,omitempty"`
	Sector       []string          `json:"sector,omitempty"`
	EntityType   []string          `json:"entity_type,omitempty"`
	AmountRange  *AmountRange      `json:"amount_range,omitempty"`
	TimeRange    *TimeRange        `json:"time_range,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

// AmountRange defines a monetary range
type AmountRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Unit string `json:"unit"`
}

// TimeRange defines a time period
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PolicyRules contains the actual policy rules
type PolicyRules struct {
	Threshold       *ThresholdRule      `json:"threshold,omitempty"`
	Limit           *LimitRule          `json:"limit,omitempty"`
	RateLimit       *RateLimitRule      `json:"rate_limit,omitempty"`
	Condition       *ConditionalRule    `json:"condition,omitempty"`
	CustomRules     []interface{}       `json:"custom_rules,omitempty"`
}

// ThresholdRule defines a threshold-based rule
type ThresholdRule struct {
	Metric      string        `json:"metric"`
	Operator    string        `json:"operator"` // gt, lt, gte, lte, eq
	Value       float64       `json:"value"`
	Duration    time.Duration `json:"duration"`
	Action      string        `json:"action"`
	Cooldown    time.Duration `json:"cooldown"`
}

// LimitRule defines a limit-based rule
type LimitRule struct {
	Resource     string  `json:"resource"`
	LimitType    string  `json:"limit_type"` // daily, weekly, monthly, per_transaction
	MaxValue     float64 `json:"max_value"`
	ResetPeriod  string  `json:"reset_period"`
	GracePeriod  int     `json:"grace_period"` // grace period in minutes
}

// RateLimitRule defines a rate-based rule
type RateLimitRule struct {
	Endpoint     string        `json:"endpoint"`
	MaxRequests  int           `json:"max_requests"`
	WindowSize   time.Duration `json:"window_size"`
	Strategy     string        `json:"strategy"` // reject, queue, throttle
}

// ConditionalRule defines a conditional rule
type ConditionalRule struct {
	Condition string `json:"condition"` // OPA-compatible condition
	Then      string `json:"then"`
	Else      string `json:"else,omitempty"`
}

// Intervention represents an intervention action
type Intervention struct {
	ID             uuid.UUID          `json:"id" db:"id"`
	PolicyID       uuid.UUID          `json:"policy_id" db:"policy_id"`
	Type           InterventionType   `json:"type" db:"type"`
	Name           string             `json:"name" db:"name"`
	Description    string             `json:"description" db:"description"`
	Parameters     InterventionParams `json:"parameters" db:"parameters"`
	TargetScope    PolicyScope        `json:"target_scope" db:"target_scope"`
	Trigger        TriggerCondition   `json:"trigger" db:"trigger"`
	Status         InterventionStatus `json:"status" db:"status"`
	Priority       int                `json:"priority" db:"priority"`
	WorkflowID     *uuid.UUID         `json:"workflow_id,omitempty" db:"workflow_id"`
	ExecutedBy     uuid.UUID          `json:"executed_by,omitempty" db:"executed_by"`
	ExecutedAt     *time.Time         `json:"executed_at,omitempty" db:"executed_at"`
	CompletedAt    *time.Time         `json:"completed_at,omitempty" db:"completed_at"`
	RolledBackAt   *time.Time         `json:"rolled_back_at,omitempty" db:"rolled_back_at"`
	Result         *InterventionResult `json:"result,omitempty" db:"result"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" db:"updated_at"`
}

// InterventionType defines the type of intervention
type InterventionType string

const (
	InterventionTypeLock        InterventionType = "lock"
	InterventionTypeUnlock      InterventionType = "unlock"
	InterventionTypeCap         InterventionType = "cap"
	InterventionTypeRedirect    InterventionType = "redirect"
	InterventionTypeSubsidize   InterventionType = "subsidize"
	InterventionTypeRestrict    InterventionType = "restrict"
	InterventionTypeRelease     InterventionType = "release"
	InterventionTypeOverride    InterventionType = "override"
)

// InterventionStatus defines the status of an intervention
type InterventionStatus string

const (
	InterventionStatusPending    InterventionStatus = "pending"
	InterventionStatusApproved   InterventionStatus = "approved"
	InterventionStatusExecuting  InterventionStatus = "executing"
	InterventionStatusActive     InterventionStatus = "active"
	InterventionStatusRollingBack InterventionStatus = "rolling_back"
	InterventionStatusRolledBack InterventionStatus = "rolled_back"
	InterventionStatusFailed     InterventionStatus = "failed"
	InterventionStatusCancelled  InterventionStatus = "cancelled"
)

// InterventionParams contains parameters for an intervention
type InterventionParams struct {
	TargetEntity   string                 `json:"target_entity" db:"target_entity"`
	TargetValue    float64                `json:"target_value" db:"target_value"`
	Unit           string                 `json:"unit" db:"unit"`
	Duration       *time.Duration         `json:"duration,omitempty" db:"duration"`
	Reason         string                 `json:"reason" db:"reason"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
}

// TriggerCondition defines when an intervention should trigger
type TriggerCondition struct {
	Type          string        `json:"type" db:"type"` // immediate, scheduled, threshold, manual
	Threshold     *ThresholdRule `json:"threshold,omitempty" db:"threshold"`
	ScheduledTime *time.Time    `json:"scheduled_time,omitempty" db:"scheduled_time"`
	ManualApproval bool         `json:"manual_approval" db:"manual_approval"`
	ApprovalRoles []string      `json:"approval_roles" db:"approval_roles"`
}

// InterventionResult contains the result of an intervention
type InterventionResult struct {
	Success          bool        `json:"success" db:"success"`
	AffectedEntities int         `json:"affected_entities" db:"affected_entities"`
	ImpactMetrics    interface{} `json:"impact_metrics" db:"impact_metrics"`
	ErrorMessage     string      `json:"error_message,omitempty" db:"error_message"`
	RollbackRequired bool        `json:"rollback_required" db:"rollback_required"`
}

// EmergencyLock represents an emergency economic lock
type EmergencyLock struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	LockType       EmergencyLockType `json:"lock_type" db:"lock_type"`
	ResourceID     string          `json:"resource_id" db:"resource_id"`
	ResourceType   string          `json:"resource_type" db:"resource_type"`
	Scope          PolicyScope     `json:"scope" db:"scope"`
	Reason         string          `json:"reason" db:"reason"`
	AuthorizedBy   uuid.UUID       `json:"authorized_by" db:"authorized_by"`
	Status         LockStatus      `json:"status" db:"status"`
	Priority       int             `json:"priority" db:"priority"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty" db:"expires_at"`
	ReleasedAt     *time.Time      `json:"released_at,omitempty" db:"released_at"`
	ReleasedBy     *uuid.UUID      `json:"released_by,omitempty" db:"released_by"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// EmergencyLockType defines types of emergency locks
type EmergencyLockType string

const (
	EmergencyLockFreeze    EmergencyLockType = "FREEZE"
	EmergencyLockCap       EmergencyLockType = "CAP"
	EmergencyLockRedirect  EmergencyLockType = "REDIRECT"
	EmergencyLockBlock     EmergencyLockType = "BLOCK"
)

// LockStatus defines the status of a lock
type LockStatus string

const (
	LockStatusActive   LockStatus = "active"
	LockStatusExpired  LockStatus = "expired"
	LockStatusReleased LockStatus = "released"
)

// Override represents an override authorization
type Override struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	Type         string        `json:"type" db:"type"`
	TargetID     string        `json:"target_id" db:"target_id"`
	OverrideValue interface{}  `json:"override_value" db:"override_value"`
	Reason       string        `json:"reason" db:"reason"`
	AuthorizedBy uuid.UUID     `json:"authorized_by" db:"authorized_by"`
	SignerID     uuid.UUID     `json:"signer_id" db:"signer_id"`
	Signature    string        `json:"signature" db:"signature"`
	Status       string        `json:"status" db:"status"`
	ExpiresAt    time.Time     `json:"expires_at" db:"expires_at"`
	UsedAt       *time.Time    `json:"used_at,omitempty" db:"used_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
}

// Scenario represents a simulation scenario
type Scenario struct {
	ID          uuid.UUID          `json:"id" db:"id"`
	Name        string             `json:"name" db:"name"`
	Description string             `json:"description" db:"description"`
	Type        ScenarioType       `json:"type" db:"type"`
	Parameters  ScenarioParameters `json:"parameters" db:"parameters"`
	Constraints []ScenarioConstraint `json:"constraints" db:"constraints"`
	Status      ScenarioStatus     `json:"status" db:"status"`
	Results     *ScenarioResults   `json:"results,omitempty" db:"results"`
	CreatedBy   uuid.UUID          `json:"created_by" db:"created_by"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// ScenarioType defines the type of scenario
type ScenarioType string

const (
	ScenarioTypeWhatIf      ScenarioType = "what_if"
	ScenarioTypeMonteCarlo  ScenarioType = "monte_carlo"
	ScenarioTypeSensitivity ScenarioType = "sensitivity"
	ScenarioTypeImpact      ScenarioType = "impact"
)

// ScenarioParameters defines parameters for a scenario
type ScenarioParameters struct {
	Variables      []ScenarioVariable `json:"variables"`
	BaselineDate   time.Time          `json:"baseline_date"`
	ForecastPeriod int                `json:"forecast_period"` // days
	Iterations     int                `json:"iterations"`
	ConfidenceLevel float64           `json:"confidence_level"`
	Assumptions    map[string]float64 `json:"assumptions"`
}

// ScenarioVariable defines a variable in a scenario
type ScenarioVariable struct {
	Name        string   `json:"name"`
	CurrentValue float64  `json:"current_value"`
	MinValue    float64  `json:"min_value"`
	MaxValue    float64  `json:"max_value"`
	Distribution string  `json:"distribution"` // normal, uniform, triangular, lognormal
	Parameters  map[string]float64 `json:"parameters"`
}

// ScenarioConstraint defines constraints for a scenario
type ScenarioConstraint struct {
	Type      string  `json:"type"`
	Variable  string  `json:"variable"`
	Operator  string  `json:"operator"`
	Value     float64 `json:"value"`
	Priority  int     `json:"priority"`
}

// ScenarioStatus defines the status of a scenario
type ScenarioStatus string

const (
	ScenarioStatusDraft      ScenarioStatus = "draft"
	ScenarioStatusRunning    ScenarioStatus = "running"
	ScenarioStatusCompleted  ScenarioStatus = "completed"
	ScenarioStatusFailed     ScenarioStatus = "failed"
	ScenarioStatusCancelled  ScenarioStatus = "cancelled"
)

// ScenarioResults contains the results of a scenario simulation
type ScenarioResults struct {
	RunID          uuid.UUID              `json:"run_id"`
	Status         string                 `json:"status"`
	Summary        ScenarioSummary        `json:"summary"`
	Distribution   map[string]interface{} `json:"distribution"`
	Sensitivity    map[string]interface{} `json:"sensitivity"`
	Recommendations []string              `json:"recommendations"`
	CompletedAt    time.Time              `json:"completed_at"`
	Duration       time.Duration          `json:"duration"`
}

// ScenarioSummary contains summary statistics
type ScenarioSummary struct {
	Mean        float64 `json:"mean"`
	Median      float64 `json:"median"`
	StdDev      float64 `json:"std_dev"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Percentiles map[int]float64 `json:"percentiles"`
}

// ScenarioTemplate represents a reusable scenario template
type ScenarioTemplate struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Type        ScenarioType `json:"type" db:"type"`
	Template    interface{} `json:"template" db:"template"`
	CreatedBy   uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}
