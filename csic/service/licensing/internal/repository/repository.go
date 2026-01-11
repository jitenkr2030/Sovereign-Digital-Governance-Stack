package repository

import (
	"context"
	"time"

	"github.com/csic-licensing/internal/domain/models"
)

// LicenseRepository defines the interface for license data operations
type LicenseRepository interface {
	Create(ctx context.Context, license *models.License) error
	GetByID(ctx context.Context, id string) (*models.License, error)
	GetByNumber(ctx context.Context, number string) (*models.License, error)
	GetByExchange(ctx context.Context, exchangeID string) ([]*models.License, error)
	GetByStatus(ctx context.Context, status models.LicenseStatus) ([]*models.License, error)
	GetExpiringSoon(ctx context.Context, days int) ([]*models.License, error)
	Update(ctx context.Context, license *models.License) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter LicenseFilter) ([]*models.License, error)
	Count(ctx context.Context, filter LicenseFilter) (int64, error)
}

// LicenseFilter defines filters for license queries
type LicenseFilter struct {
	ExchangeID   string
	Jurisdiction models.Jurisdiction
	Status       models.LicenseStatus
	Type         models.LicenseType
	FromDate     *time.Time
	ToDate       *time.Time
	SearchTerm   string
	Page         int
	PageSize     int
}

// ViolationRepository defines the interface for violation data operations
type ViolationRepository interface {
	Create(ctx context.Context, violation *models.ComplianceViolation) error
	GetByID(ctx context.Context, id string) (*models.ComplianceViolation, error)
	GetByLicense(ctx context.Context, licenseID string) ([]*models.ComplianceViolation, error)
	GetByExchange(ctx context.Context, exchangeID string) ([]*models.ComplianceViolation, error)
	GetOpenViolations(ctx context.Context) ([]*models.ComplianceViolation, error)
	GetByStatus(ctx context.Context, status models.ViolationStatus) ([]*models.ComplianceViolation, error)
	GetBySeverity(ctx context.Context, severity models.ViolationSeverity) ([]*models.ComplianceViolation, error)
	Update(ctx context.Context, violation *models.ComplianceViolation) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ViolationFilter) ([]*models.ComplianceViolation, error)
	Count(ctx context.Context, filter ViolationFilter) (int64, error)
	GetStats(ctx context.Context) (*models.ViolationStats, error)
	GetTrends(ctx context.Context, startDate, endDate time.Time) ([]models.ViolationTrend, error)
}

// ViolationFilter defines filters for violation queries
type ViolationFilter struct {
	LicenseID    string
	ExchangeID   string
	Status       models.ViolationStatus
	Severity     models.ViolationSeverity
	Type         models.ViolationType
	FromDate     *time.Time
	ToDate       *time.Time
	AssignedTo   string
	Page         int
	PageSize     int
}

// ReportRepository defines the interface for report data operations
type ReportRepository interface {
	Create(ctx context.Context, report *models.Report) error
	GetByID(ctx context.Context, id string) (*models.Report, error)
	GetByLicense(ctx context.Context, licenseID string) ([]*models.Report, error)
	GetByExchange(ctx context.Context, exchangeID string) ([]*models.Report, error)
	GetByStatus(ctx context.Context, status models.ReportStatus) ([]*models.Report, error)
	GetByType(ctx context.Context, reportType models.ReportType) ([]*models.Report, error)
	Update(ctx context.Context, report *models.Report) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ReportFilter) ([]*models.Report, error)
	Count(ctx context.Context, filter ReportFilter) (int64, error)
	GetStats(ctx context.Context) (*models.ReportStats, error)
	GetOverdueReports(ctx context.Context) ([]*models.Report, error)
}

// ReportFilter defines filters for report queries
type ReportFilter struct {
	LicenseID    string
	ExchangeID   string
	Status       models.ReportStatus
	ReportType   models.ReportType
	FromDate     *time.Time
	ToDate       *time.Time
	GeneratedBy  string
	Page         int
	PageSize     int
}

// ComplianceRepository defines the interface for compliance data operations
type ComplianceRepository interface {
	CalculateComplianceScore(ctx context.Context, exchangeID string) (*models.ComplianceScore, error)
	GetDashboardMetrics(ctx context.Context, exchangeID string) (*models.DashboardMetrics, error)
	GetKPIMetrics(ctx context.Context, exchangeID string) ([]*models.KPIMetric, error)
	GetComplianceChecks(ctx context.Context, licenseID string) ([]*models.ComplianceCheck, error)
	UpdateComplianceCheck(ctx context.Context, check *models.ComplianceCheck) error
	CreateComplianceCheck(ctx context.Context, check *models.ComplianceCheck) error
	GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityItem, error)
	GetLicenseExpiryTimeline(ctx context.Context, days int) ([]models.LicenseExpiry, error)
}

// ApplicationRepository defines the interface for license application operations
type ApplicationRepository interface {
	Create(ctx context.Context, application *models.LicenseApplication) error
	GetByID(ctx context.Context, id string) (*models.LicenseApplication, error)
	GetByExchange(ctx context.Context, exchangeID string) ([]*models.LicenseApplication, error)
	GetByStatus(ctx context.Context, status string) ([]*models.LicenseApplication, error)
	Update(ctx context.Context, application *models.LicenseApplication) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ApplicationFilter) ([]*models.LicenseApplication, error)
}

// ApplicationFilter defines filters for application queries
type ApplicationFilter struct {
	ExchangeID   string
	Jurisdiction models.Jurisdiction
	Status       string
	FromDate     *time.Time
	ToDate       *time.Time
	Page         int
	PageSize     int
}

// AuditRepository defines the interface for audit log operations
type AuditRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	GetByEntity(ctx context.Context, entityType, entityID string) ([]*models.AuditLog, error)
	GetByActor(ctx context.Context, actorID string, limit int) ([]*models.AuditLog, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.AuditLog, error)
	List(ctx context.Context, filter AuditFilter) ([]*models.AuditLog, error)
}

// AuditFilter defines filters for audit log queries
type AuditFilter struct {
	EntityType   string
	EntityID     string
	ActorID      string
	Action       string
	FromDate     *time.Time
	ToDate       *time.Time
	Page         int
	PageSize     int
}

// DocumentRepository defines the interface for document storage operations
type DocumentRepository interface {
	Upload(ctx context.Context, doc *models.DocumentReference) error
	GetByID(ctx context.Context, id string) (*models.DocumentReference, error)
	GetByLicense(ctx context.Context, licenseID string) ([]*models.DocumentReference, error)
	Delete(ctx context.Context, id string) error
	GetDownloadURL(ctx context.Context, id string) (string, error)
}

// AlertRepository defines the interface for compliance alert operations
type AlertRepository interface {
	Create(ctx context.Context, alert *models.ComplianceAlert) error
	GetByID(ctx context.Context, id string) (*models.ComplianceAlert, error)
	GetUnread(ctx context.Context) ([]*models.ComplianceAlert, error)
	GetByEntity(ctx context.Context, entityType, entityID string) ([]*models.ComplianceAlert, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAsResolved(ctx context.Context, id string, resolvedBy string) error
	List(ctx context.Context, filter AlertFilter) ([]*models.ComplianceAlert, error)
}

// AlertFilter defines filters for alert queries
type AlertFilter struct {
	EntityType   string
	EntityID     string
	Severity     string
	IsRead       *bool
	IsResolved   *bool
	AlertType    string
	Page         int
	PageSize     int
}

// RegulatorRepository defines the interface for regulator operations
type RegulatorRepository interface {
	Create(ctx context.Context, regulator *models.Regulator) error
	GetByID(ctx context.Context, id string) (*models.Regulator, error)
	GetByCode(ctx context.Context, code string) (*models.Regulator, error)
	GetByJurisdiction(ctx context.Context, jurisdiction models.Jurisdiction) ([]*models.Regulator, error)
	GetAll(ctx context.Context) ([]*models.Regulator, error)
	Update(ctx context.Context, regulator *models.Regulator) error
	Delete(ctx context.Context, id string) error
}
