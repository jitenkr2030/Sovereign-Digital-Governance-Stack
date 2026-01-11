package domain

import (
	"encoding/json"
	"time"
)

// EnforcementAction represents an enforcement action taken by the control layer
type EnforcementAction struct {
	ID               string              `json:"id" db:"id"`
	Type             EnforcementType     `json:"type" db:"type"`
	TargetID         string              `json:"target_id" db:"target_id"`
	TargetType       EntityType          `json:"target_type" db:"target_type"`
	Reason           string              `json:"reason" db:"reason"`
	PolicyID         string              `json:"policy_id" db:"policy_id"`
	Severity         PolicySeverity      `json:"severity" db:"severity"`
	Status           EnforcementStatus   `json:"status" db:"status"`
	Parameters       json.RawMessage     `json:"parameters" db:"parameters"`
	Duration         *time.Duration      `json:"duration" db:"duration"`
	ExpiresAt        *time.Time          `json:"expires_at" db:"expires_at"`
	ExecutedAt       *time.Time          `json:"executed_at" db:"executed_at"`
	CompletedAt      *time.Time          `json:"completed_at" db:"completed_at"`
	CancelledAt      *time.Time          `json:"cancelled_at" db:"cancelled_at"`
	CancellationReason string            `json:"cancellation_reason" db:"cancellation_reason"`
	RetryCount       int                 `json:"retry_count" db:"retry_count"`
	LastError        string              `json:"last_error" db:"last_error"`
	CreatedBy        string              `json:"created_by" db:"created_by"`
	CreatedAt        time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at" db:"updated_at"`
	Metadata         json.RawMessage     `json:"metadata" db:"metadata"`
}

// EnforcementType represents the type of enforcement action
type EnforcementType string

const (
	EnforcementThrottle    EnforcementType = "throttle"
	EnforcementFreeze      EnforcementType = "freeze"
	EnforcementRevoke      EnforcementType = "revoke"
	EnforcementAlert       EnforcementType = "alert"
	EnforcementLog         EnforcementType = "log"
	EnforcementRequireKYC  EnforcementType = "require_kyc"
	EnforcementRateLimit   EnforcementType = "rate_limit"
	EnforcementSuspend     EnforcementType = "suspend"
	EnforcementBlacklist   EnforcementType = "blacklist"
	EnforcementWhitelist   EnforcementType = "whitelist"
	EnforcementNotify      EnforcementType = "notify"
	EnforcementFine        EnforcementType = "fine"
)

// EnforcementStatus represents the status of an enforcement action
type EnforcementStatus string

const (
	EnforcementPending   EnforcementStatus = "pending"
	EnforcementExecuting EnforcementStatus = "executing"
	EnforcementActive    EnforcementStatus = "active"
	EnforcementCompleted EnforcementStatus = "completed"
	EnforcementFailed    EnforcementStatus = "failed"
	EnforcementCancelled EnforcementStatus = "cancelled"
	EnforcementExpired   EnforcementStatus = "expired"
)

// EnforcementExecutionResult represents the result of enforcement execution
type EnforcementExecutionResult struct {
	ActionID   string            `json:"action_id"`
	Success    bool              `json:"success"`
	Error      string            `json:"error,omitempty"`
	Response   map[string]interface{} `json:"response,omitempty"`
	ExecutedAt time.Time         `json:"executed_at"`
}

// EnforcementEvent represents an event for enforcement actions
type EnforcementEvent struct {
	EventID        string          `json:"event_id"`
	ActionID       string          `json:"action_id"`
	Type           EnforcementType `json:"type"`
	TargetID       string          `json:"target_id"`
	TargetType     EntityType      `json:"target_type"`
	Reason         string          `json:"reason"`
	PolicyID       string          `json:"policy_id"`
	Severity       PolicySeverity  `json:"severity"`
	Parameters     json.RawMessage `json:"parameters,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
	CorrelationID  string          `json:"correlation_id"`
	Source         string          `json:"source"`
}

// EnforcementCommand represents a command to execute enforcement
type EnforcementCommand struct {
	CommandID   string          `json:"command_id"`
	Action      EnforcementType `json:"action"`
	TargetID    string          `json:"target_id"`
	TargetType  EntityType      `json:"target_type"`
	Reason      string          `json:"reason"`
	PolicyID    string          `json:"policy_id"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	Priority    int             `json:"priority"`
	IdempotencyKey string       `json:"idempotency_key"`
	Timeout     time.Duration   `json:"timeout"`
	CallbackURL string          `json:"callback_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// EnforcementHistory represents the history of enforcement actions for an entity
type EnforcementHistory struct {
	EntityID    string             `json:"entity_id"`
	EntityType  EntityType         `json:"entity_type"`
	Actions     []EnforcementAction `json:"actions"`
	TotalCount  int                `json:"total_count"`
	ActiveCount int                `json:"active_count"`
	CompletedCount int             `json:"completed_count"`
	FailedCount int                `json:"failed_count"`
}

// EnforcementTemplate represents a template for common enforcement actions
type EnforcementTemplate struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        EnforcementType  `json:"type"`
	Parameters  json.RawMessage  `json:"parameters"`
	ValidFor    *time.Duration   `json:"valid_for"`
	Preconditions []string       `json:"preconditions"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// EnforcementStatistics represents statistics about enforcement actions
type EnforcementStatistics struct {
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	TotalActions    int       `json:"total_actions"`
	ActiveActions   int       `json:"active_actions"`
	CompletedActions int      `json:"completed_actions"`
	FailedActions   int       `json:"failed_actions"`
	ByType          map[EnforcementType]int `json:"by_type"`
	BySeverity      map[PolicySeverity]int `json:"by_severity"`
	AverageDuration time.Duration `json:"average_duration"`
	SuccessRate     float64    `json:"success_rate"`
}
