package domain

import (
	"time"

	"github.com/google/uuid"
)

// Policy represents an access control policy
type Policy struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Effect      PolicyEffect `json:"effect" db:"effect"`
	Priority    int        `json:"priority" db:"priority"`
	Enabled     bool       `json:"enabled" db:"enabled"`
	Conditions  PolicyConditions `json:"conditions" db:"-"`
	ConditionsJSON string   `json:"-" db:"conditions"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// PolicyEffect represents the effect of a policy
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "ALLOW"
	PolicyEffectDeny  PolicyEffect = "DENY"
)

// PolicyConditions defines the conditions for a policy
type PolicyConditions struct {
	Subjects    []SubjectCondition  `json:"subjects,omitempty"`
	Resources   []ResourceCondition `json:"resources,omitempty"`
	Actions     []ActionCondition   `json:"actions,omitempty"`
	Environment []EnvironmentCondition `json:"environment,omitempty"`
}

// SubjectCondition defines conditions for the subject (who)
type SubjectCondition struct {
	Roles       []string `json:"roles,omitempty"`
	Users       []string `json:"users,omitempty"`
	Groups      []string `json:"groups,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// ResourceCondition defines conditions for the resource (what)
type ResourceCondition struct {
	Types       []string `json:"types,omitempty"`
	IDs         []string `json:"ids,omitempty"`
	Patterns    []string `json:"patterns,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// ActionCondition defines conditions for the action (how)
type ActionCondition struct {
	Operations  []string `json:"operations,omitempty"`
	Methods     []string `json:"methods,omitempty"`
}

// EnvironmentCondition defines conditions for the environment
type EnvironmentCondition struct {
	IPRanges    []string `json:"ip_ranges,omitempty"`
	TimeRanges  []TimeRangeCondition `json:"time_ranges,omitempty"`
	Devices     []string `json:"devices,omitempty"`
	Locations   []string `json:"locations,omitempty"`
}

// TimeRangeCondition defines a time range condition
type TimeRangeCondition struct {
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	DaysOfWeek []int  `json:"days_of_week"`
}

// AccessDecision represents the result of an access check
type AccessDecision string

const (
	AccessDecisionAllow    AccessDecision = "ALLOW"
	AccessDecisionDeny     AccessDecision = "DENY"
	AccessDecisionNotApplicable AccessDecision = "NOT_APPLICABLE"
	AccessDecisionIndeterminate AccessDecision = "INDETERMINATE"
)

// AccessRequest represents a request for access control decision
type AccessRequest struct {
	Subject   Subject   `json:"subject"`
	Resource  Resource  `json:"resource"`
	Action    Action    `json:"action"`
	Context   Context   `json:"context"`
}

// Subject represents the entity requesting access
type Subject struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Roles      []string          `json:"roles"`
	Groups     []string          `json:"groups"`
	Attributes map[string]string `json:"attributes"`
}

// Resource represents the resource being accessed
type Resource struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Path       string            `json:"path"`
	Attributes map[string]string `json:"attributes"`
}

// Action represents the action being performed
type Action struct {
	Name   string `json:"name"`
	Method string `json:"method"`
}

// Context represents the context of the access request
type Context struct {
	IPAddress  string            `json:"ip_address"`
	UserAgent  string            `json:"user_agent"`
	Time       time.Time         `json:"time"`
	Location   string            `json:"location"`
	Device     string            `json:"device"`
	Custom     map[string]string `json:"custom,omitempty"`
}

// AccessResponse represents the response from an access check
type AccessResponse struct {
	Decision   AccessDecision `json:"decision"`
	PolicyID   *uuid.UUID     `json:"policy_id,omitempty"`
	Reason     string         `json:"reason"`
	Obligations []Obligation  `json:"obligations,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

// Obligation represents a required action after access is granted
type Obligation struct {
	Name       string `json:"name"`
	Parameters map[string]string `json:"parameters"`
}

// ResourceOwnership represents resource ownership information
type ResourceOwnership struct {
	ResourceID  string    `json:"resource_id"`
	OwnerID     string    `json:"owner_id"`
	OwnerType   string    `json:"owner_type"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuditLog represents an access control audit log entry
type AuditLog struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	Request     AccessRequest `json:"request" db:"-"`
	RequestJSON string        `json:"-" db:"request"`
	Response    AccessResponse `json:"response" db:"-"`
	ResponseJSON string       `json:"-" db:"response"`
	Decision    AccessDecision `json:"decision" db:"decision"`
	UserID      string         `json:"user_id" db:"user_id"`
	ResourceID  string         `json:"resource_id" db:"resource_id"`
	Action      string         `json:"action" db:"action"`
	IPAddress   string         `json:"ip_address" db:"ip_address"`
	UserAgent   string         `json:"user_agent" db:"user_agent"`
	Timestamp   time.Time      `json:"timestamp" db:"timestamp"`
}
