package domain

import (
	"time"

	"github.com/google/uuid"
)

// License Status constants
type LicenseStatus string

const (
	LicenseStatusActive    LicenseStatus = "ACTIVE"
	LicenseStatusSuspended LicenseStatus = "SUSPENDED"
	LicenseStatusRevoked   LicenseStatus = "REVOKED"
	LicenseStatusExpired   LicenseStatus = "EXPIRED"
)

// Application Status constants
type ApplicationStatus string

const (
	AppStatusDraft      ApplicationStatus = "DRAFT"
	AppStatusSubmitted  ApplicationStatus = "SUBMITTED"
	AppStatusUnderReview ApplicationStatus = "UNDER_REVIEW"
	AppStatusApproved   ApplicationStatus = "APPROVED"
	AppStatusRejected   ApplicationStatus = "REJECTED"
	AppStatusWithdrawn  ApplicationStatus = "WITHDRAWN"
)

// Application Type constants
type ApplicationType string

const (
	AppTypeNew       ApplicationType = "NEW"
	AppTypeRenewal   ApplicationType = "RENEWAL"
	AppTypeAmendment ApplicationType = "AMENDMENT"
	AppTypeTransfer  ApplicationType = "TRANSFER"
)

// LicenseType constants
type LicenseType string

const (
	LicenseTypeExchange      LicenseType = "EXCHANGE"
	LicenseTypeCustody       LicenseType = "CUSTODY"
	LicenseTypeMining        LicenseType = "MINING"
	LicenseTypeTrading       LicenseType = "TRADING"
	LicenseTypeATM           LicenseType = "ATM"
	LicenseTypeBrokerDealer  LicenseType = "BROKER_DEALER"
	LicenseTypeMSB           LicenseType = "MSB"
)

// Obligation Status constants
type ObligationStatus string

const (
	ObligationPending    ObligationStatus = "PENDING"
	ObligationFulfilled  ObligationStatus = "FULFILLED"
	ObligationOverdue    ObligationStatus = "OVERDUE"
	ObligationWaived     ObligationStatus = "WAIVED"
	ObligationInProgress ObligationStatus = "IN_PROGRESS"
)

// Compliance Tier constants
type ComplianceTier string

const (
	TierGold     ComplianceTier = "GOLD"
	TierSilver   ComplianceTier = "SILVER"
	TierBronze   ComplianceTier = "BRONZE"
	TierAtRisk   ComplianceTier = "AT_RISK"
	TierCritical ComplianceTier = "CRITICAL"
)

// License represents a regulatory license granted to an entity
type License struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	EntityID      uuid.UUID     `json:"entity_id" db:"entity_id"`
	Type          LicenseType   `json:"type" db:"type"`
	Status        LicenseStatus `json:"status" db:"status"`
	LicenseNumber string        `json:"license_number" db:"license_number"`
	IssuedDate    time.Time     `json:"issued_date" db:"issued_date"`
	ExpiryDate    time.Time     `json:"expiry_date" db:"expiry_date"`
	Conditions    string        `json:"conditions" db:"conditions"`
	Restrictions  string        `json:"restrictions" db:"restrictions"`
	Jurisdiction  string        `json:"jurisdiction" db:"jurisdiction"`
	IssuedBy      string        `json:"issued_by" db:"issued_by"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	RevokedAt     *time.Time    `json:"revoked_at,omitempty" db:"revoked_at"`
	RevocationReason string     `json:"revocation_reason,omitempty" db:"revocation_reason"`
}

// LicenseApplication represents a license application
type LicenseApplication struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	EntityID        uuid.UUID        `json:"entity_id" db:"entity_id"`
	Type            ApplicationType  `json:"type" db:"type"`
	LicenseType     LicenseType      `json:"license_type" db:"license_type"`
	Status          ApplicationStatus `json:"status" db:"status"`
	SubmittedAt     time.Time        `json:"submitted_at" db:"submitted_at"`
	ReviewedAt      *time.Time       `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewedBy      *uuid.UUID       `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewerNotes   string           `json:"reviewer_notes" db:"reviewer_notes"`
	RequestedTerms  string           `json:"requested_terms" db:"requested_terms"`
	GrantedTerms    string           `json:"granted_terms,omitempty" db:"granted_terms"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
	Documents       []string         `json:"documents,omitempty"`
}

// Regulation represents a regulatory requirement
type Regulation struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Title          string    `json:"title" db:"title"`
	Description    string    `json:"description" db:"description"`
	Category       string    `json:"category" db:"category"`
	Jurisdiction   string    `json:"jurisdiction" db:"jurisdiction"`
	EffectiveDate  time.Time `json:"effective_date" db:"effective_date"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Requirements   string    `json:"requirements" db:"requirements"`
	PenaltyDetails string    `json:"penalty_details" db:"penalty_details"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ComplianceScore represents an entity's compliance score
type ComplianceScore struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	EntityID      uuid.UUID     `json:"entity_id" db:"entity_id"`
	TotalScore    float64       `json:"total_score" db:"total_score"`
	Tier          ComplianceTier `json:"tier" db:"tier"`
	Breakdown     string        `json:"breakdown" db:"breakdown"`
	PeriodStart   time.Time     `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time     `json:"period_end" db:"period_end"`
	CalculatedAt  time.Time     `json:"calculated_at" db:"calculated_at"`
	CalculationDetails string   `json:"calculation_details" db:"calculation_details"`
}

// Obligation represents a regulatory obligation for an entity
type Obligation struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	EntityID        uuid.UUID        `json:"entity_id" db:"entity_id"`
	RegulationID    uuid.UUID        `json:"regulation_id" db:"regulation_id"`
	Description     string           `json:"description" db:"description"`
	DueDate         time.Time        `json:"due_date" db:"due_date"`
	Status          ObligationStatus `json:"status" db:"status"`
	Priority        int              `json:"priority" db:"priority"`
	EvidenceRefs    string           `json:"evidence_refs" db:"evidence_refs"`
	FulfilledAt     *time.Time       `json:"fulfilled_at,omitempty" db:"fulfilled_at"`
	FulfilledEvidence string         `json:"fulfilled_evidence,omitempty" db:"fulfilled_evidence"`
	ReminderSentAt  *time.Time       `json:"reminder_sent_at,omitempty" db:"reminder_sent_at"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
	Regulation      *Regulation      `json:"regulation,omitempty"`
}

// AuditRecord represents an audit trail entry
type AuditRecord struct {
	ID            uuid.UUID `json:"id" db:"id"`
	EntityID      uuid.UUID `json:"entity_id" db:"entity_id"`
	ActionType    string    `json:"action_type" db:"action_type"`
	ActorID       uuid.UUID `json:"actor_id" db:"actor_id"`
	ActorType     string    `json:"actor_type" db:"actor_type"`
	ResourceID    uuid.UUID `json:"resource_id" db:"resource_id"`
	ResourceType  string    `json:"resource_type" db:"resource_type"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	OldValue      string    `json:"old_value,omitempty" db:"old_value"`
	NewValue      string    `json:"new_value,omitempty" db:"new_value"`
	Changes       string    `json:"changes,omitempty" db:"changes"`
	Metadata      string    `json:"metadata,omitempty" db:"metadata"`
	IPAddress     string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     string    `json:"user_agent,omitempty" db:"user_agent"`
}

// Entity represents a regulated entity
type Entity struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	LegalName        string    `json:"legal_name" db:"legal_name"`
	RegistrationNum  string    `json:"registration_num" db:"registration_num"`
	Jurisdiction     string    `json:"jurisdiction" db:"jurisdiction"`
	EntityType       string    `json:"entity_type" db:"entity_type"`
	Address          string    `json:"address" db:"address"`
	ContactEmail     string    `json:"contact_email" db:"contact_email"`
	Status           string    `json:"status" db:"status"`
	RiskLevel        string    `json:"risk_level" db:"risk_level"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// ComplianceStats represents compliance statistics for reporting
type ComplianceStats struct {
	TotalEntities       int64   `json:"total_entities"`
	ActiveLicenses      int64   `json:"active_licenses"`
	PendingApplications int64   `json:"pending_applications"`
	OverdueObligations  int64   `json:"overdue_obligations"`
	AverageScore        float64 `json:"average_score"`
	GoldTierCount       int64   `json:"gold_tier_count"`
	AtRiskCount         int64   `json:"at_risk_count"`
}
