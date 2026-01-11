package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	StatusPending   ApprovalStatus = "PENDING"
	StatusApproved  ApprovalStatus = "APPROVED"
	StatusRejected  ApprovalStatus = "REJECTED"
	StatusExpired   ApprovalStatus = "EXPIRED"
	StatusCancelled ApprovalStatus = "CANCELLED"
)

// ActionType represents the type of action taken on an approval
type ActionType string

const (
	ActionApprove ActionType = "APPROVE"
	ActionReject  ActionType = "REJECT"
	ActionAbstain ActionType = "ABSTAIN"
)

// PolicyConfig defines the rules for an approval policy
type PolicyConfig struct {
	QuorumRequired  int             `json:"quorum_required"`  // M - minimum approvals needed
	TotalApprovers  int             `json:"total_approvers"`  // N - total approvers required
	TimeoutHours    int             `json:"timeout_hours"`    // Time limit for approval
	DelegateAllowed bool            `json:"delegate_allowed"` // Whether delegation is permitted
	RequiredRoles   []string        `json:"required_roles"`   // Roles required to approve
	AutoRejectAfter int             `json:"auto_reject_after"` // Auto-reject after N rejections
	Priority        string          `json:"priority"`         // Priority level
	Metadata        json.RawMessage `json:"metadata"`         // Additional policy metadata
}

// Value implements driver.Valuer for PolicyConfig
func (p PolicyConfig) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner for PolicyConfig
func (p *PolicyConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

// ApprovalDefinition represents an approval policy template
type ApprovalDefinition struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Category    string       `json:"category" db:"category"` // e.g., "financial", "operational", "security"
	PolicyConfig PolicyConfig `json:"policy_config" db:"policy_config"`
	Version     int          `json:"version" db:"version"`
	IsActive    bool         `json:"is_active" db:"is_active"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID    `json:"created_by" db:"created_by"`
}

// ApprovalRequest represents an active approval request
type ApprovalRequest struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	DefinitionID   uuid.UUID       `json:"definition_id" db:"definition_id"`
	RequesterID    uuid.UUID       `json:"requester_id" db:"requester_id"`
	ResourceType   string          `json:"resource_type" db:"resource_type"`     // e.g., "policy", "budget", "transaction"
	ResourceID     string          `json:"resource_id" db:"resource_id"`         // ID of the resource being approved
	ContextData    json.RawMessage `json:"context_data" db:"context_data"`       // Snapshot of request data
	Status         ApprovalStatus  `json:"status" db:"status"`
	CurrentStep    int             `json:"current_step" db:"current_step"`       // Current approval step
	Priority       string          `json:"priority" db:"priority"`
	Deadline       time.Time       `json:"deadline" db:"deadline"`
	WorkflowID     string          `json:"workflow_id" db:"workflow_id"`         // Temporal workflow ID
	Metadata       json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// ApprovalAction represents an immutable action taken on an approval request
type ApprovalAction struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	RequestID       uuid.UUID   `json:"request_id" db:"request_id"`
	ActorID         uuid.UUID   `json:"actor_id" db:"actor_id"`
	ActorType       string      `json:"actor_type" db:"actor_type"`         // "user", "system", "service"
	ActionType      ActionType  `json:"action_type" db:"action_type"`
	OnBehalfOf      *uuid.UUID  `json:"on_behalf_of,omitempty" db:"on_behalf_of"` // If delegation was used
	StepNumber      int         `json:"step_number" db:"step_number"`
	Comments        string      `json:"comments" db:"comments"`
	SignatureHash   string      `json:"signature_hash" db:"signature_hash"`   // Cryptographic signature
	PreviousHash    string      `json:"previous_hash" db:"previous_hash"`     // Hash chaining for immutability
	IPAddress       string      `json:"ip_address" db:"ip_address"`
	UserAgent       string      `json:"user_agent" db:"user_agent"`
	Timestamp       time.Time   `json:"timestamp" db:"timestamp"`
}

// Delegation represents a delegation relationship between users
type Delegation struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	DelegatorID  uuid.UUID       `json:"delegator_id" db:"delegator_id"`
	DelegateeID  uuid.UUID       `json:"delegatee_id" db:"delegatee_id"`
	DefinitionID uuid.UUID       `json:"definition_id,omitempty" db:"definition_id"` // Optional: limit to specific approval type
	Scope        json.RawMessage `json:"scope" db:"scope"`                           // Limitations on delegation
	ValidFrom    time.Time       `json:"valid_from" db:"valid_from"`
	ValidUntil   time.Time       `json:"valid_until" db:"valid_until"`
	IsRevoked    bool            `json:"is_revoked" db:"is_revoked"`
	Reason       string          `json:"reason" db:"reason"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	RevokedAt    *time.Time      `json:"revoked_at,omitempty" db:"revoked_at"`
}

// ApproverAssignment represents an assigned approver for a request
type ApproverAssignment struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	RequestID    uuid.UUID  `json:"request_id" db:"request_id"`
	ApproverID   uuid.UUID  `json:"approver_id" db:"approver_id"`
	Role         string     `json:"role" db:"role"`               // Role required for this slot
	StepOrder    int        `json:"step_order" db:"step_order"`   // Order in approval sequence
	HasActed     bool       `json:"has_acted" db:"has_acted"`
	ActedAt      *time.Time `json:"acted_at,omitempty" db:"acted_at"`
	ActionType   *ActionType `json:"action_type,omitempty" db:"action_type"`
	AssignedAt   time.Time  `json:"assigned_at" db:"assigned_at"`
}

// ApprovalSummary provides a summarized view of approval progress
type ApprovalSummary struct {
	RequestID       uuid.UUID     `json:"request_id"`
	TotalRequired   int           `json:"total_required"`
	ApprovalsCount  int           `json:"approvals_count"`
	RejectionsCount int           `json:"rejections_count"`
	PendingCount    int           `json:"pending_count"`
	QuorumMet       bool          `json:"quorum_met"`
	Status          ApprovalStatus `json:"status"`
	Deadline        time.Time     `json:"deadline"`
	TimeRemaining   time.Duration `json:"time_remaining"`
}

// AuditTrailEntry represents a single entry in the audit trail
type AuditTrailEntry struct {
	ActionID       uuid.UUID   `json:"action_id"`
	ActorID        uuid.UUID   `json:"actor_id"`
	ActorName      string      `json:"actor_name"`
	ActionType     ActionType  `json:"action_type"`
	Comments       string      `json:"comments"`
	Timestamp      time.Time   `json:"timestamp"`
	SignatureValid bool        `json:"signature_valid"`
	DelegatedFrom  *uuid.UUID  `json:"delegated_from,omitempty"`
	IPAddress      string      `json:"ip_address"`
}
