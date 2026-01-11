// Compliance Management Module - Port Interfaces
// Interface definitions for repository and external service contracts

package port

import (
	"context"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
)

// AuditLogPort defines the interface for audit logging
type AuditLogPort interface {
	Log(ctx context.Context, entry *AuditEntry) error
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp     time.Time
	ActorID       string
	Action        string
	ResourceType  string
	ResourceID    string
	EntityID      string
	SessionID     string
	IPAddress     string
	Description   string
	Result        string
	Error         string
	Metadata      map[string]interface{}
}

// EntityRepository defines the interface for entity storage
type EntityRepository interface {
	Create(ctx context.Context, entity *domain.RegulatedEntity) error
	GetByID(ctx context.Context, id string) (*domain.RegulatedEntity, error)
	GetByRegistrationNumber(ctx context.Context, regNumber string) (*domain.RegulatedEntity, error)
	Update(ctx context.Context, entity *domain.RegulatedEntity) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter EntityFilter) ([]*domain.RegulatedEntity, error)
}

// LicenseRepository defines the interface for license storage
type LicenseRepository interface {
	Create(ctx context.Context, license *domain.License) error
	GetByID(ctx context.Context, id string) (*domain.License, error)
	GetByLicenseNumber(ctx context.Context, number string) (*domain.License, error)
	GetActiveByEntityAndType(ctx context.Context, entityID string, licenseType domain.LicenseType) (*domain.License, error)
	Update(ctx context.Context, license *domain.License) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter LicenseFilter) ([]*domain.License, error)
	GetExpiring(ctx context.Context, withinDays int) ([]*domain.License, error)
	GetStatistics(ctx context.Context) (*LicenseStatistics, error)
}

// ObligationRepository defines the interface for obligation storage
type ObligationRepository interface {
	Create(ctx context.Context, obligation *domain.ComplianceObligation) error
	GetByID(ctx context.Context, id string) (*domain.ComplianceObligation, error)
	Update(ctx context.Context, obligation *domain.ComplianceObligation) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ObligationFilter) ([]*domain.ComplianceObligation, error)
	GetByEntityID(ctx context.Context, entityID string, status ...domain.ObligationStatus) ([]*domain.ComplianceObligation, error)
	GetOverdue(ctx context.Context) ([]*domain.ComplianceObligation, error)
	GetDueWithin(ctx context.Context, entityID string, withinDays int) ([]*domain.ComplianceObligation, error)
	GetStatistics(ctx context.Context, entityID string) (*ObligationStatistics, error)
}

// ViolationRepository defines the interface for violation storage
type ViolationRepository interface {
	Create(ctx context.Context, violation *domain.ComplianceViolation) error
	GetByID(ctx context.Context, id string) (*domain.ComplianceViolation, error)
	Update(ctx context.Context, violation *domain.ComplianceViolation) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ViolationFilter) ([]*domain.ComplianceViolation, error)
	GetOpenByEntityID(ctx context.Context, entityID string) ([]*domain.ComplianceViolation, error)
	GetStatistics(ctx context.Context, entityID string) (*ViolationStatistics, error)
}

// PenaltyRepository defines the interface for penalty storage
type PenaltyRepository interface {
	Create(ctx context.Context, penalty *domain.Penalty) error
	GetByID(ctx context.Context, id string) (*domain.Penalty, error)
	Update(ctx context.Context, penalty *domain.Penalty) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter PenaltyFilter) ([]*domain.Penalty, error)
	GetByViolationID(ctx context.Context, violationID string) ([]*domain.Penalty, error)
	GetOverdue(ctx context.Context) ([]*domain.Penalty, error)
}

// Filter types

// EntityFilter defines filters for entity queries
type EntityFilter struct {
	Status     []domain.EntityStatus
	Type       []domain.EntityType
	Jurisdiction string
	Search     string
	Limit      int
	Offset     int
}

// LicenseFilter defines filters for license queries
type LicenseFilter struct {
	EntityID    string
	Status      []domain.LicenseStatus
	Type        []domain.LicenseType
	Jurisdiction string
	ExpiresBefore *time.Time
	ExpiresAfter  *time.Time
	Limit       int
	Offset      int
}

// ObligationFilter defines filters for obligation queries
type ObligationFilter struct {
	EntityID    string
	LicenseID   string
	Status      []domain.ObligationStatus
	Type        []domain.ObligationType
	Priority    []domain.ObligationPriority
	DueBefore   *time.Time
	DueAfter    *time.Time
	Limit       int
	Offset      int
}

// ViolationFilter defines filters for violation queries
type ViolationFilter struct {
	EntityID    string
	Status      []domain.ViolationStatus
	Type        []domain.ViolationType
	Severity    []domain.ViolationSeverity
	DetectedAfter *time.Time
	DetectedBefore *time.Time
	Limit       int
	Offset      int
}

// PenaltyFilter defines filters for penalty queries
type PenaltyFilter struct {
	ViolationID string
	Status      []domain.PenaltyStatus
	Type        []domain.PenaltyType
	DueBefore   *time.Time
	Limit       int
	Offset      int
}

// Statistics types

// LicenseStatistics holds license statistics
type LicenseStatistics struct {
	Total        int
	Active       int
	Expired      int
	Suspended    int
	Revoked      int
	PendingReview int
	ExpiringSoon int
}

// ObligationStatistics holds obligation statistics
type ObligationStatistics struct {
	Total       int
	Pending     int
	Submitted   int
	Verified    int
	Approved    int
	Overdue     int
	NonCompliant int
}

// ViolationStatistics holds violation statistics
type ViolationStatistics struct {
	Total       int
	Open        int
	Resolved    int
	Closed      int
	BySeverity  map[domain.ViolationSeverity]int
	ByType      map[domain.ViolationType]int
	TotalPenalties float64
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	EntityID       string
	EntityName     string
	ReportPeriod   ReportPeriod
	GeneratedAt    time.Time
	Licenses       []*domain.License
	Obligations    []*domain.ComplianceObligation
	Violations     []*domain.ComplianceViolation
	Statistics     interface{}
}

// ReportPeriod represents a time period for reports
type ReportPeriod struct {
	Start time.Time
	End   time.Time
}
