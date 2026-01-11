package domain

import (
	"encoding/json"
	"time"
)

// SystemState represents the state of an entity in the system
type SystemState struct {
	EntityID          string           `json:"entity_id" db:"entity_id"`
	EntityType        EntityType       `json:"entity_type" db:"entity_type"`
	EntityName        string           `json:"entity_name" db:"entity_name"`
	ComplianceScore   float64          `json:"compliance_score" db:"compliance_score"`
	OperationalStatus OperationalStatus `json:"operational_status" db:"operational_status"`
	LicenseStatus     LicenseStatus    `json:"license_status" db:"license_status"`
	RiskLevel         RiskLevel        `json:"risk_level" db:"risk_level"`
	LastAuditDate     *time.Time       `json:"last_audit_date" db:"last_audit_date"`
	NextAuditDate     *time.Time       `json:"next_audit_date" db:"next_audit_date"`
	ActiveEnforcements int             `json:"active_enforcements" db:"active_enforcements"`
	Flags             []string         `json:"flags" db:"flags"`
	Tags              []string         `json:"tags" db:"tags"`
	Metadata          json.RawMessage  `json:"metadata" db:"metadata"`
	Version           int              `json:"version" db:"version"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

// OperationalStatus represents the operational status of an entity
type OperationalStatus string

const (
	OperationalStatusActive      OperationalStatus = "active"
	OperationalStatusInactive    OperationalStatus = "inactive"
	OperationalStatusSuspended   OperationalStatus = "suspended"
	OperationalStatusRevoked     OperationalStatus = "revoked"
	OperationalStatusPending     OperationalStatus = "pending_approval"
	OperationalStatusUnderReview OperationalStatus = "under_review"
	OperationalStatusRestricted  OperationalStatus = "restricted"
	OperationalStatusOffline     OperationalStatus = "offline"
)

// LicenseStatus represents the license status of an entity
type LicenseStatus string

const (
	LicenseStatusActive     LicenseStatus = "active"
	LicenseStatusExpired    LicenseStatus = "expired"
	LicenseStatusSuspended  LicenseStatus = "suspended"
	LicenseStatusRevoked    LicenseStatus = "revoked"
	LicenseStatusPending    LicenseStatus = "pending"
	LicenseStatusDenied     LicenseStatus = "denied"
	LicenseStatusExpiredSoon LicenseStatus = "expiring_soon"
)

// RiskLevel represents the risk level of an entity
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
	RiskLevelUnknown  RiskLevel = "unknown"
)

// StateTransition represents a state transition in the system
type StateTransition struct {
	ID            string           `json:"id" db:"id"`
	EntityID      string           `json:"entity_id" db:"entity_id"`
	FromStatus    OperationalStatus `json:"from_status" db:"from_status"`
	ToStatus      OperationalStatus `json:"to_status" db:"to_status"`
	Reason        string           `json:"reason" db:"reason"`
	TriggeredBy   string           `json:"triggered_by" db:"triggered_by"`
	TriggerType   TriggerType      `json:"trigger_type" db:"trigger_type"`
	PolicyID      string           `json:"policy_id" db:"policy_id"`
	Duration      time.Duration    `json:"duration" db:"duration"`
	ApprovedBy    string           `json:"approved_by" db:"approved_by"`
	ApprovedAt    *time.Time       `json:"approved_at" db:"approved_at"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
}

// TriggerType represents what triggered a state transition
type TriggerType string

const (
	TriggerManual       TriggerType = "manual"
	TriggerPolicy       TriggerType = "policy"
	TriggerScheduled    TriggerType = "scheduled"
	TriggerEnforcement  TriggerType = "enforcement"
	TriggerAudit        TriggerType = "audit"
	TriggerSystem       TriggerType = "system"
	TriggerExternal     TriggerType = "external"
)

// StateQuery represents a query for system states
type StateQuery struct {
	EntityIDs       []string         `json:"entity_ids,omitempty"`
	EntityTypes     []EntityType     `json:"entity_types,omitempty"`
	Statuses        []OperationalStatus `json:"statuses,omitempty"`
	LicenseStatuses []LicenseStatus  `json:"license_statuses,omitempty"`
	RiskLevels      []RiskLevel      `json:"risk_levels,omitempty"`
	MinComplianceScore *float64      `json:"min_compliance_score,omitempty"`
	MaxComplianceScore *float64      `json:"max_compliance_score,omitempty"`
	HasActiveEnforcements *bool      `json:"has_active_enforcements,omitempty"`
	Flags           []string         `json:"flags,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	Offset          int              `json:"offset"`
	Limit           int              `json:"limit"`
	SortBy          string           `json:"sort_by"`
	SortOrder       string           `json:"sort_order"`
}

// StateSnapshot represents a snapshot of system state at a point in time
type StateSnapshot struct {
	SnapshotID     string         `json:"snapshot_id"`
	EntityID       string         `json:"entity_id"`
	EntityType     EntityType     `json:"entity_type"`
	State          SystemState    `json:"state"`
	Enforcements   []EnforcementAction `json:"enforcements"`
	Policies       []Policy       `json:"policies"`
	Timestamp      time.Time      `json:"timestamp"`
	Checksum       string         `json:"checksum"`
}

// StateRegistryConfig represents configuration for the state registry
type StateRegistryConfig struct {
	CacheTTL           time.Duration `json:"cache_ttl"`
	MaxCacheSize       int           `json:"max_cache_size"`
	SyncInterval       time.Duration `json:"sync_interval"`
	EnableCompression  bool          `json:"enable_compression"`
	EnableEncryption   bool          `json:"enable_encryption"`
	SnapshotInterval   time.Duration `json:"snapshot_interval"`
	RetentionPeriod    time.Duration `json:"retention_period"`
}

// StateStatistics represents statistics about the state registry
type StateStatistics struct {
	TotalEntities    int                    `json:"total_entities"`
	ByType           map[EntityType]int     `json:"by_type"`
	ByStatus         map[OperationalStatus]int `json:"by_status"`
	ByLicenseStatus  map[LicenseStatus]int  `json:"by_license_status"`
	ByRiskLevel      map[RiskLevel]int      `json:"by_risk_level"`
	AverageCompliance float64               `json:"average_compliance_score"`
	ActiveEnforcements int                  `json:"active_enforcements"`
	RecentTransitions int                   `json:"recent_transitions"`
	LastUpdated      time.Time              `json:"last_updated"`
}
