package ports

import (
	"context"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/google/uuid"
)

// LicenseRepository defines the output port for licensing operations
type LicenseRepository interface {
	// License operations
	CreateLicense(ctx context.Context, license *domain.License) error
	GetLicense(ctx context.Context, id uuid.UUID) (*domain.License, error)
	GetLicenseByNumber(ctx context.Context, number string) (*domain.License, error)
	GetLicensesByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.License, error)
	UpdateLicense(ctx context.Context, license *domain.License) error
	UpdateLicenseStatus(ctx context.Context, id uuid.UUID, status domain.LicenseStatus) error
	GetLicensesExpiringSoon(ctx context.Context, days int) ([]domain.License, error)
	GetActiveLicenses(ctx context.Context) ([]domain.License, error)

	// Application operations
	CreateApplication(ctx context.Context, app *domain.LicenseApplication) error
	GetApplication(ctx context.Context, id uuid.UUID) (*domain.LicenseApplication, error)
	GetApplicationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.LicenseApplication, error)
	GetApplicationsByStatus(ctx context.Context, status domain.ApplicationStatus) ([]domain.LicenseApplication, error)
	UpdateApplication(ctx context.Context, app *domain.LicenseApplication) error
	UpdateApplicationStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, reviewerID *uuid.UUID, notes string) error

	// Entity operations
	CreateEntity(ctx context.Context, entity *domain.Entity) error
	GetEntity(ctx context.Context, id uuid.UUID) (*domain.Entity, error)
	GetEntityByRegistration(ctx context.Context, regNum string) (*domain.Entity, error)
	UpdateEntity(ctx context.Context, entity *domain.Entity) error

	// Regulation operations
	CreateRegulation(ctx context.Context, reg *domain.Regulation) error
	GetRegulation(ctx context.Context, id uuid.UUID) (*domain.Regulation, error)
	GetRegulationsByCategory(ctx context.Context, category string) ([]domain.Regulation, error)
}

// ComplianceRepository defines the output port for compliance scoring operations
type ComplianceRepository interface {
	CreateScore(ctx context.Context, score *domain.ComplianceScore) error
	GetCurrentScore(ctx context.Context, entityID uuid.UUID) (*domain.ComplianceScore, error)
	GetScoreHistory(ctx context.Context, entityID uuid.UUID, limit int) ([]domain.ComplianceScore, error)
	GetScoresByTier(ctx context.Context, tier domain.ComplianceTier) ([]domain.ComplianceScore, error)
	CalculateAverageScore(ctx context.Context) (float64, error)
}

// ObligationRepository defines the output port for obligation management operations
type ObligationRepository interface {
	CreateObligation(ctx context.Context, obligation *domain.Obligation) error
	GetObligation(ctx context.Context, id uuid.UUID) (*domain.Obligation, error)
	GetObligationsByEntity(ctx context.Context, entityID uuid.UUID) ([]domain.Obligation, error)
	GetObligationsByStatus(ctx context.Context, status domain.ObligationStatus) ([]domain.Obligation, error)
	GetOverdueObligations(ctx context.Context) ([]domain.Obligation, error)
	UpdateObligation(ctx context.Context, obligation *domain.Obligation) error
	UpdateObligationStatus(ctx context.Context, id uuid.UUID, status domain.ObligationStatus) error
	MarkObligationFulfilled(ctx context.Context, id uuid.UUID, evidence string) error
	GetUpcomingObligations(ctx context.Context, days int) ([]domain.Obligation, error)
}

// AuditRepository defines the output port for audit operations
type AuditRepository interface {
	CreateAuditRecord(ctx context.Context, record *domain.AuditRecord) error
	GetAuditRecord(ctx context.Context, id uuid.UUID) (*domain.AuditRecord, error)
	GetAuditRecordsByEntity(ctx context.Context, entityID uuid.UUID, limit, offset int) ([]domain.AuditRecord, error)
	GetAuditRecordsByResource(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]domain.AuditRecord, error)
	GetAuditRecordsByActor(ctx context.Context, actorID uuid.UUID, limit int) ([]domain.AuditRecord, error)
	GetAuditRecordsByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]domain.AuditRecord, error)
	GetAuditRecords(ctx context.Context, filter AuditFilter) ([]domain.AuditRecord, error)
	CountAuditRecords(ctx context.Context, filter AuditFilter) (int64, error)
}

// AuditFilter defines filter criteria for audit records
type AuditFilter struct {
	EntityID     *uuid.UUID
	ActorID      *uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	ActionType   string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

// Time is needed for audit repository operations
import "time"
