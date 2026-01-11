package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// AuditEventType categorizes different types of audit events
type AuditEventType string

const (
	EventTypeAuthentication   AuditEventType = "AUTHENTICATION"
	EventTypeAuthorization    AuditEventType = "AUTHORIZATION"
	EventTypeDataAccess       AuditEventType = "DATA_ACCESS"
	EventTypeDataModification AuditEventType = "DATA_MODIFICATION"
	EventTypeDataDeletion     AuditEventType = "DATA_DELETION"
	EventTypeConfiguration    AuditEventType = "CONFIGURATION"
	EventTypeSystem           AuditEventType = "SYSTEM"
	EventTypeNetwork          AuditEventType = "NETWORK"
	EventTypeSecurity         AuditEventType = "SECURITY"
	EventTypePolicy           AuditEventType = "POLICY"
	EventTypeExport           AuditEventType = "EXPORT"
	EventTypeImport           AuditEventType = "IMPORT"
	EventTypeIntegration      AuditEventType = "INTEGRATION"
	EventTypeForensic         AuditEventType = "FORENSIC"
)

// EventOutcome represents the outcome of an audited event
type EventOutcome string

const (
	OutcomeSuccess   EventOutcome = "SUCCESS"
	OutcomeFailure   EventOutcome = "FAILURE"
	OutcomeDenied    EventOutcome = "DENIED"
	OutcomePartial   EventOutcome = "PARTIAL"
	OutcomePending   EventOutcome = "PENDING"
	OutcomeUnknown   EventOutcome = "UNKNOWN"
)

// ActorType represents the type of actor performing an action
type ActorType string

const (
	ActorTypeUser       ActorType = "USER"
	ActorTypeService    ActorType = "SERVICE"
	ActorTypeSystem     ActorType = "SYSTEM"
	ActorTypeAPI        ActorType = "API"
	ActorTypeScheduler  ActorType = "SCHEDULER"
	ActorTypeExternal   ActorType = "EXTERNAL"
)

// JSONMap is a custom type for storing JSON data in PostgreSQL
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONMap: not a byte slice")
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// AuditEvent represents a single audit event in the system
type AuditEvent struct {
	ID              string                 `json:"id" db:"id"`
	Timestamp       time.Time              `json:"timestamp" db:"timestamp"`
	EventType       AuditEventType         `json:"event_type" db:"event_type"`
	ActorID         string                 `json:"actor_id" db:"actor_id"`
	ActorType       ActorType              `json:"actor_type" db:"actor_type"`
	ActorName       string                 `json:"actor_name,omitempty" db:"actor_name"`
	Action          string                 `json:"action" db:"action"`
	Resource        string                 `json:"resource" db:"resource"`
	ResourceID      string                 `json:"resource_id" db:"resource_id"`
	ResourceType    string                 `json:"resource_type" db:"resource_type"`
	Outcome         EventOutcome           `json:"outcome" db:"outcome"`
	Details         JSONMap                `json:"details,omitempty" db:"details"`
	IPAddress       string                 `json:"ip_address" db:"ip_address"`
	UserAgent       string                 `json:"user_agent" db:"user_agent"`
	Location        string                 `json:"location" db:"location"`
	SessionID       string                 `json:"session_id" db:"session_id"`
	RequestID       string                 `json:"request_id" db:"request_id"`
	TraceID         string                 `json:"trace_id" db:"trace_id"`
	IntegrityHash   string                 `json:"integrity_hash" db:"integrity_hash"`
	PreviousHash    string                 `json:"previous_hash" db:"previous_hash"`
	ChainIndex      int64                  `json:"chain_index" db:"chain_index"`
	Sealed          bool                   `json:"sealed" db:"sealed"`
	SealedAt        *time.Time             `json:"sealed_at,omitempty" db:"sealed_at"`
	MerkleProof     JSONMap                `json:"merkle_proof,omitempty" db:"merkle_proof"`
	ComplianceTags  []string               `json:"compliance_tags,omitempty" db:"compliance_tags"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// SealRecord represents a cryptographic seal for a batch of events
type SealRecord struct {
	ID              string                 `json:"id" db:"id"`
	Timestamp       time.Time              `json:"timestamp" db:"timestamp"`
	RootHash        string                 `json:"root_hash" db:"root_hash"`
	FirstEventID    string                 `json:"first_event_id" db:"first_event_id"`
	LastEventID     string                 `json:"last_event_id" db:"last_event_id"`
	EventCount      int64                  `json:"event_count" db:"event_count"`
	Signature       string                 `json:"signature" db:"signature"`
	SignerID        string                 `json:"signer_id" db:"signer_id"`
	SignerType      string                 `json:"signer_type" db:"signer_type"`
	Algorithm       string                 `json:"algorithm" db:"algorithm"`
	WitnessType     string                 `json:"witness_type,omitempty" db:"witness_type"`
	WitnessID       string                 `json:"witness_id,omitempty" db:"witness_id"`
	WitnessSignature string                `json:"witness_signature,omitempty" db:"witness_signature"`
	PreviousSealID  string                 `json:"previous_seal_id,omitempty" db:"previous_seal_id"`
	PreviousRootHash string               `json:"previous_root_hash,omitempty" db:"previous_root_hash"`
	Metadata        JSONMap                `json:"metadata,omitempty" db:"metadata"`
	Verified        bool                   `json:"verified" db:"verified"`
	VerifiedAt      *time.Time             `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// AuditChain represents a linked chain of audit events
type AuditChain struct {
	ID              string                 `json:"id" db:"id"`
	ChainType       string                 `json:"chain_type" db:"chain_type"`
	FirstEventID    string                 `json:"first_event_id" db:"first_event_id"`
	LastEventID     string                 `json:"last_event_id" db:"last_event_id"`
	CurrentRootHash string                 `json:"current_root_hash" db:"current_root_hash"`
	Length          int64                  `json:"length" db:"length"`
	SealID          string                 `json:"seal_id,omitempty" db:"seal_id"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// ForensicQuery represents a query for forensic investigation
type ForensicQuery struct {
	QueryID        string                 `json:"query_id"`
	RequesterID    string                 `json:"requester_id"`
	RequesterRole  string                 `json:"requester_role"`
	Justification  string                 `json:"justification"`
	EventTypes     []AuditEventType       `json:"event_types,omitempty"`
	ActorIDs       []string               `json:"actor_ids,omitempty"`
	ActorTypes     []ActorType            `json:"actor_types,omitempty"`
	Resources      []string               `json:"resources,omitempty"`
	Actions        []string               `json:"actions,omitempty"`
	Outcomes       []EventOutcome         `json:"outcomes,omitempty"`
	IPAddresses    []string               `json:"ip_addresses,omitempty"`
	TimeRange      TimeRange              `json:"time_range"`
	SearchTerms    []string               `json:"search_terms,omitempty"`
	IncludeDetails bool                   `json:"include_details"`
	IncludeChain   bool                   `json:"include_chain"`
	IncludeHash    bool                   `json:"include_hash"`
	Limit          int                    `json:"limit"`
	Offset         int                    `json:"offset"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ForensicResult represents the result of a forensic query
type ForensicResult struct {
	QueryID        string                 `json:"query_id"`
	EventsFound    int64                  `json:"events_found"`
	EventsReturned int                    `json:"events_returned"`
	Events         []AuditEvent           `json:"events"`
	IntegrityChain []SealRecord           `json:"integrity_chain,omitempty"`
	Verification   *VerificationReport    `json:"verification,omitempty"`
	QueryTime      time.Duration          `json:"query_time"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// VerificationReport represents the result of integrity verification
type VerificationReport struct {
	Verified       bool                   `json:"verified"`
	Timestamp      time.Time              `json:"timestamp"`
	EventsVerified int64                  `json:"events_verified"`
	EventsFailed   int64                  `json:"events_failed"`
	FailedEvents   []FailedEvent          `json:"failed_events,omitempty"`
	ChainIntegrity ChainIntegrityResult   `json:"chain_integrity"`
}

// FailedEvent represents an event that failed verification
type FailedEvent struct {
	EventID       string    `json:"event_id"`
	EventIndex    int64     `json:"event_index"`
	ExpectedHash  string    `json:"expected_hash"`
	ActualHash    string    `json:"actual_hash"`
	EventTime     time.Time `json:"event_time"`
	FailureReason string    `json:"failure_reason"`
}

// ChainIntegrityResult represents the result of chain integrity verification
type ChainIntegrityResult struct {
	Continuous   bool     `json:"continuous"`
	HashMatch    bool     `json:"hash_match"`
	SealValid    bool     `json:"seal_valid"`
	BrokenLinks  []string `json:"broken_links,omitempty"`
}

// AuditQuery represents a query for audit events
type AuditQuery struct {
	EventTypes    []AuditEventType `json:"event_types,omitempty"`
	ActorIDs      []string         `json:"actor_ids,omitempty"`
	Resources     []string         `json:"resources,omitempty"`
	Outcomes      []EventOutcome   `json:"outcomes,omitempty"`
	TimeRange     *TimeRange       `json:"time_range,omitempty"`
	IncludeSealed bool             `json:"include_sealed"`
	Limit         int              `json:"limit"`
	Offset        int              `json:"offset"`
	SortBy        string           `json:"sort_by"`
	SortOrder     string           `json:"sort_order"`
}

// AuditStatistics represents statistics about audit events
type AuditStatistics struct {
	TotalEvents       int64             `json:"total_events"`
	EventsByType      map[string]int64  `json:"events_by_type"`
	EventsByOutcome   map[string]int64  `json:"events_by_outcome"`
	EventsByActor     map[string]int64  `json:"events_by_actor"`
	EventsByResource  map[string]int64  `json:"events_by_resource"`
	EventsLast24h     int64             `json:"events_last_24h"`
	EventsLastHour    int64             `json:"events_last_hour"`
	FailedEvents      int64             `json:"failed_events"`
	SealedEvents      int64             `json:"sealed_events"`
	UnsealedEvents    int64             `json:"unsealed_events"`
	AverageEventsPerHour float64        `json:"average_events_per_hour"`
	GeneratedAt       time.Time         `json:"generated_at"`
}
