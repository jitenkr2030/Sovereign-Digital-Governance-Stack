package domain

import (
	"encoding/json"
	"time"
)

// Intervention represents an intervention request or action
type Intervention struct {
	ID               string             `json:"id" db:"id"`
	Title            string             `json:"title" db:"title"`
	Description      string             `json:"description" db:"description"`
	Type             InterventionType   `json:"type" db:"type"`
	TriggerEvent     string             `json:"trigger_event" db:"trigger_event"`
	TargetID         string             `json:"target_id" db:"target_id"`
	TargetType       EntityType         `json:"target_type" db:"target_type"`
	Severity         PolicySeverity     `json:"severity" db:"severity"`
	Status           InterventionStatus `json:"status" db:"status"`
	Priority         int                `json:"priority" db:"priority"`
	RequiresApproval bool               `json:"requires_approval" db:"requires_approval"`
	ApproverRole     string             `json:"approver_role" db:"approver_role"`
	ApprovedBy       string             `json:"approved_by" db:"approved_by"`
	ApprovedAt       *time.Time         `json:"approved_at" db:"approved_at"`
	RejectedBy       string             `json:"rejected_by" db:"rejected_by"`
	RejectedAt       *time.Time         `json:"rejected_at" db:"rejected_at"`
	RejectionReason  string             `json:"rejection_reason" db:"rejection_reason"`
	ExecutedAt       *time.Time         `json:"executed_at" db:"executed_at"`
	CompletedAt      *time.Time         `json:"completed_at" db:"completed_at"`
	CancelledAt      *time.Time         `json:"cancelled_at" db:"cancelled_at"`
	CancellationReason string           `json:"cancellation_reason" db:"cancellation_reason"`
	PolicyID         string             `json:"policy_id" db:"policy_id"`
	Actions          []InterventionAction `json:"actions" db:"actions"`
	Result           json.RawMessage    `json:"result" db:"result"`
	Metadata         json.RawMessage    `json:"metadata" db:"metadata"`
	CreatedBy        string             `json:"created_by" db:"created_by"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
}

// InterventionType represents the type of intervention
type InterventionType string

const (
	InterventionManual   InterventionType = "manual"
	InterventionAutomatic InterventionType = "automatic"
	InterventionEmergency InterventionType = "emergency"
	InterventionScheduled InterventionType = "scheduled"
	InterventionRemediation InterventionType = "remediation"
)

// InterventionStatus represents the status of an intervention
type InterventionStatus string

const (
	InterventionPending    InterventionStatus = "pending"
	InterventionAwaitingApproval InterventionStatus = "awaiting_approval"
	InterventionApproved   InterventionStatus = "approved"
	InterventionRejected   InterventionStatus = "rejected"
	InterventionExecuting  InterventionStatus = "executing"
	InterventionCompleted  InterventionStatus = "completed"
	InterventionFailed     InterventionStatus = "failed"
	InterventionCancelled  InterventionStatus = "cancelled"
	InterventionOnHold     InterventionStatus = "on_hold"
)

// InterventionAction represents an action within an intervention
type InterventionAction struct {
	ActionID     string          `json:"action_id"`
	Sequence     int             `json:"sequence"`
	Type         string          `json:"type"`
	Parameters   json.RawMessage `json:"parameters"`
	Status       ActionStatus    `json:"status"`
	ExecutedAt   *time.Time      `json:"executed_at"`
	CompletedAt  *time.Time      `json:"completed_at"`
	Result       json.RawMessage `json:"result"`
	Error        string          `json:"error"`
}

// ActionStatus represents the status of an intervention action
type ActionStatus string

const (
	ActionPending   ActionStatus = "pending"
	ActionReady     ActionStatus = "ready"
	ActionExecuting ActionStatus = "executing"
	ActionCompleted ActionStatus = "completed"
	ActionFailed    ActionStatus = "failed"
	ActionSkipped   ActionStatus = "skipped"
	ActionCancelled ActionStatus = "cancelled"
)

// InterventionTemplate represents a template for common interventions
type InterventionTemplate struct {
	ID              string               `json:"id" db:"id"`
	Name            string               `json:"name" db:"name"`
	Description     string               `json:"description" db:"description"`
	Type            InterventionType     `json:"type" db:"type"`
	Severity        PolicySeverity       `json:"severity" db:"severity"`
	TriggerEvent    string               `json:"trigger_event" db:"trigger_event"`
	RequiresApproval bool                `json:"requires_approval" db:"requires_approval"`
	ApproverRole    string               `json:"approver_role" db:"approver_role"`
	Actions         []InterventionAction `json:"actions" db:"actions"`
	PreConditions   []string             `json:"pre_conditions" db:"pre_conditions"`
	PostConditions  []string             `json:"post_conditions" db:"post_conditions"`
	EscalationRules []EscalationRule     `json:"escalation_rules" db:"escalation_rules"`
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" db:"updated_at"`
}

// EscalationRule defines rules for escalating interventions
type EscalationRule struct {
	RuleID         string          `json:"rule_id"`
	Condition      string          `json:"condition"`
	EscalateTo     string          `json:"escalate_to"`
	EscalateAfter  time.Duration   `json:"escalate_after"`
	MaxEscalations int             `json:"max_escalations"`
	NotifyOriginal bool            `json:"notify_original"`
}

// InterventionEvent represents an event related to an intervention
type InterventionEvent struct {
	EventID        string             `json:"event_id"`
	InterventionID string             `json:"intervention_id"`
	EventType      InterventionEventType `json:"event_type"`
	PreviousStatus InterventionStatus `json:"previous_status"`
	NewStatus      InterventionStatus `json:"new_status"`
	Description    string             `json:"description"`
	PerformedBy    string             `json:"performed_by"`
	Timestamp      time.Time          `json:"timestamp"`
	Metadata       json.RawMessage    `json:"metadata"`
}

// InterventionEventType represents the type of intervention event
type InterventionEventType string

const (
	InterventionEventCreated      InterventionEventType = "created"
	InterventionEventApproved     InterventionEventType = "approved"
	InterventionEventRejected     InterventionEventType = "rejected"
	InterventionEventExecuted     InterventionEventType = "executed"
	InterventionEventCompleted    InterventionEventType = "completed"
	InterventionEventFailed       InterventionEventType = "failed"
	InterventionEventCancelled    InterventionEventType = "cancelled"
	InterventionEventEscalated    InterventionEventType = "escalated"
	InterventionEventOnHold       InterventionEventType = "on_hold"
	InterventionEventComment      InterventionEventType = "comment"
)

// InterventionStatistics represents statistics about interventions
type InterventionStatistics struct {
	PeriodStart       time.Time               `json:"period_start"`
	PeriodEnd         time.Time               `json:"period_end"`
	TotalInterventions int                    `json:"total_interventions"`
	ByType            map[InterventionType]int `json:"by_type"`
	ByStatus          map[InterventionStatus]int `json:"by_status"`
	BySeverity        map[PolicySeverity]int  `json:"by_severity"`
	AverageResolution time.Duration           `json:"average_resolution"`
	ApprovalRate      float64                 `json:"approval_rate"`
	CompletionRate    float64                 `json:"completion_rate"`
	EscalationRate    float64                 `json:"escalation_rate"`
}
