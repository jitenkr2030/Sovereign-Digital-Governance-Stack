package ports

import (
	"context"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/google/uuid"
)

// LicenseService defines the input port for licensing operations
type LicenseService interface {
	// Application workflow
	SubmitApplication(ctx context.Context, req SubmitApplicationRequest) (*domain.LicenseApplication, error)
	ReviewApplication(ctx context.Context, appID uuid.UUID, decision ApplicationReviewRequest) (*domain.License, error)
	WithdrawApplication(ctx context.Context, appID uuid.UUID) error
	GetApplication(ctx context.Context, appID uuid.UUID) (*domain.LicenseApplication, error)
	ListApplications(ctx context.Context, filter ApplicationFilter) ([]domain.LicenseApplication, error)

	// License management
	IssueLicense(ctx context.Context, req IssueLicenseRequest) (*domain.License, error)
	SuspendLicense(ctx context.Context, licenseID uuid.UUID, reason string) error
	RevokeLicense(ctx context.Context, licenseID uuid.UUID, reason string) error
	RenewLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error)
	GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error)
	GetEntityLicenses(ctx context.Context, entityID uuid.UUID) ([]domain.License, error)
	GetExpiringLicenses(ctx context.Context, days int) ([]domain.License, error)

	// Entity management
	RegisterEntity(ctx context.Context, req RegisterEntityRequest) (*domain.Entity, error)
	GetEntity(ctx context.Context, entityID uuid.UUID) (*domain.Entity, error)
	UpdateEntityRiskLevel(ctx context.Context, entityID uuid.UUID, riskLevel string) error

	// Regulation management
	CreateRegulation(ctx context.Context, req CreateRegulationRequest) (*domain.Regulation, error)
	GetRegulation(ctx context.Context, regID uuid.UUID) (*domain.Regulation, error)
}

// ComplianceService defines the input port for compliance scoring operations
type ComplianceService interface {
	CalculateScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error)
	GetCurrentScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error)
	GetScoreHistory(ctx context.Context, entityID uuid.UUID, limit int) ([]domain.ComplianceScore, error)
	RecalculateAllScores(ctx context.Context) error
	GetComplianceStats(ctx context.Context) (*domain.ComplianceStats, error)
}

// ObligationService defines the input port for obligation management operations
type ObligationService interface {
	CreateObligation(ctx context.Context, req CreateObligationRequest) (*domain.Obligation, error)
	GetObligation(ctx context.Context, obligationID uuid.UUID) (*domain.Obligation, error)
	GetEntityObligations(ctx context.Context, entityID uuid.UUID) ([]domain.Obligation, error)
	GetObligationsByStatus(ctx context.Context, status domain.ObligationStatus) ([]domain.Obligation, error)
	FulfillObligation(ctx context.Context, obligationID uuid.UUID, evidence string) error
	WaiveObligation(ctx context.Context, obligationID uuid.UUID, reason string) error
	GetOverdueObligations(ctx context.Context) ([]domain.Obligation, error)
	GetUpcomingObligations(ctx context.Context, days int) ([]domain.Obligation, error)
	CheckAndUpdateOverdueObligations(ctx context.Context) error
}

// AuditService defines the input port for audit operations
type AuditService interface {
	CreateAuditRecord(ctx context.Context, req CreateAuditRequest) (*domain.AuditRecord, error)
	GetAuditRecord(ctx context.Context, recordID uuid.UUID) (*domain.AuditRecord, error)
	GetEntityAuditTrail(ctx context.Context, entityID uuid.UUID, limit, offset int) ([]domain.AuditRecord, error)
	GetResourceAuditTrail(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]domain.AuditRecord, error)
	GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]domain.AuditRecord, error)
	CountAuditLogs(ctx context.Context, filter AuditLogFilter) (int64, error)
}

// DTOs for service operations

// SubmitApplicationRequest represents a license application submission
type SubmitApplicationRequest struct {
	EntityID       uuid.UUID        `json:"entity_id" binding:"required"`
	Type           domain.ApplicationType `json:"type" binding:"required"`
	LicenseType    domain.LicenseType    `json:"license_type" binding:"required"`
	RequestedTerms string           `json:"requested_terms" binding:"required"`
	Documents      []string         `json:"documents"`
}

// ApplicationReviewRequest represents a license application review decision
type ApplicationReviewRequest struct {
	Approved     bool     `json:"approved" binding:"required"`
	ReviewerID   uuid.UUID `json:"reviewer_id" binding:"required"`
	Notes        string   `json:"notes"`
	GrantedTerms string   `json:"granted_terms,omitempty"`
	Conditions   string   `json:"conditions,omitempty"`
}

// ApplicationFilter represents filter criteria for applications
type ApplicationFilter struct {
	EntityID   *uuid.UUID
	Status     *domain.ApplicationStatus
	Type       *domain.ApplicationType
	From       *time.Time
	To         *time.Time
	Page       int
	PageSize   int
}

// IssueLicenseRequest represents a license issuance request
type IssueLicenseRequest struct {
	EntityID       uuid.UUID        `json:"entity_id" binding:"required"`
	LicenseType    domain.LicenseType  `json:"license_type" binding:"required"`
	LicenseNumber  string           `json:"license_number" binding:"required"`
	ExpiryDays     int              `json:"expiry_days"`
	Conditions     string           `json:"conditions"`
	Restrictions   string           `json:"restrictions"`
	Jurisdiction   string           `json:"jurisdiction" binding:"required"`
	IssuedBy       string           `json:"issued_by" binding:"required"`
}

// RegisterEntityRequest represents an entity registration request
type RegisterEntityRequest struct {
	Name            string `json:"name" binding:"required"`
	LegalName       string `json:"legal_name" binding:"required"`
	RegistrationNum string `json:"registration_num" binding:"required"`
	Jurisdiction    string `json:"jurisdiction" binding:"required"`
	EntityType      string `json:"entity_type" binding:"required"`
	Address         string `json:"address" binding:"required"`
	ContactEmail    string `json:"contact_email" binding:"required,email"`
	RiskLevel       string `json:"risk_level"`
}

// CreateRegulationRequest represents a regulation creation request
type CreateRegulationRequest struct {
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description" binding:"required"`
	Category       string `json:"category" binding:"required"`
	Jurisdiction   string `json:"jurisdiction" binding:"required"`
	EffectiveDate  string `json:"effective_date"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Requirements   string `json:"requirements" binding:"required"`
	PenaltyDetails string `json:"penalty_details"`
}

// CreateObligationRequest represents an obligation creation request
type CreateObligationRequest struct {
	EntityID     uuid.UUID `json:"entity_id" binding:"required"`
	RegulationID uuid.UUID `json:"regulation_id" binding:"required"`
	Description  string   `json:"description" binding:"required"`
	DueDate      string   `json:"due_date" binding:"required"`
	Priority     int      `json:"priority"`
	EvidenceRefs string   `json:"evidence_refs"`
}

// CreateAuditRequest represents an audit record creation request
type CreateAuditRequest struct {
	EntityID     uuid.UUID `json:"entity_id" binding:"required"`
	ActionType   string   `json:"action_type" binding:"required"`
	ActorID      uuid.UUID `json:"actor_id" binding:"required"`
	ActorType    string   `json:"actor_type" binding:"required"`
	ResourceID   uuid.UUID `json:"resource_id" binding:"required"`
	ResourceType string   `json:"resource_type" binding:"required"`
	OldValue     string   `json:"old_value,omitempty"`
	NewValue     string   `json:"new_value,omitempty"`
	Changes      string   `json:"changes,omitempty"`
	Metadata     string   `json:"metadata,omitempty"`
	IPAddress    string   `json:"ip_address,omitempty"`
	UserAgent    string   `json:"user_agent,omitempty"`
}

// AuditLogFilter represents filter criteria for audit logs
type AuditLogFilter struct {
	EntityID     *uuid.UUID
	ActorID      *uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	ActionType   string
	From         *time.Time
	To           *time.Time
	Page         int
	PageSize     int
}

// Time is needed for DTOs
import "time"
