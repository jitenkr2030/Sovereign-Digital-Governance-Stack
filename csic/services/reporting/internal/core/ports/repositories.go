package ports

import (
	"context"

	"github.com/reporting-service/reporting/internal/core/domain"
)

// SARRepository defines the interface for SAR persistence.
type SARRepository interface {
	Create(ctx context.Context, sar *domain.SAR) error
	GetByID(ctx context.Context, id string) (*domain.SAR, error)
	GetByReportNumber(ctx context.Context, number string) (*domain.SAR, error)
	Update(ctx context.Context, sar *domain.SAR) error
	List(ctx context.Context, filter SARFilter) ([]*domain.SAR, error)
	Count(ctx context.Context, filter SARFilter) (int64, error)
	Delete(ctx context.Context, id string) error
}

// SARFilter represents filtering criteria for SAR queries.
type SARFilter struct {
	Status      []domain.ReportStatus
	SubjectID   string
	ReporterID  string
	StartDate   *time.Time
	EndDate     *time.Time
	MinAmount   float64
	MaxAmount   float64
	FieldOffice string
	Limit       int
	Offset      int
}

// CTRRepository defines the interface for CTR persistence.
type CTRRepository interface {
	Create(ctx context.Context, ctr *domain.CTR) error
	GetByID(ctx context.Context, id string) (*domain.CTR, error)
	GetByReportNumber(ctx context.Context, number string) (*domain.CTR, error)
	Update(ctx context.Context, ctr *domain.CTR) error
	List(ctx context.Context, filter CTRFilter) ([]*domain.CTR, error)
	Count(ctx context.Context, filter CTRFilter) (int64, error)
	Delete(ctx context.Context, id string) error
}

// CTRFilter represents filtering criteria for CTR queries.
type CTRFilter struct {
	Status           []domain.ReportStatus
	PersonID         string
	AccountNumber    string
	TransactionType  string
	MinAmount        float64
	MaxAmount        float64
	StartDate        *time.Time
	EndDate          *time.Time
	InstitutionID    string
	Limit            int
	Offset           int
}

// ComplianceRuleRepository defines the interface for compliance rule persistence.
type ComplianceRuleRepository interface {
	Create(ctx context.Context, rule *domain.ComplianceRule) error
	GetByID(ctx context.Context, id string) (*domain.ComplianceRule, error)
	GetByCode(ctx context.Context, code string) (*domain.ComplianceRule, error)
	Update(ctx context.Context, rule *domain.ComplianceRule) error
	List(ctx context.Context, filter ComplianceRuleFilter) ([]*domain.ComplianceRule, error)
	Delete(ctx context.Context, id string) error
}

// ComplianceRuleFilter represents filtering criteria for compliance rule queries.
type ComplianceRuleFilter struct {
	Category     string
	Regulation   string
	Severity     string
	IsActive     *bool
	EffectiveDate *time.Time
	Limit        int
	Offset       int
}

// AlertRepository defines the interface for alert persistence.
type AlertRepository interface {
	Create(ctx context.Context, alert *domain.Alert) error
	GetByID(ctx context.Context, id string) (*domain.Alert, error)
	GetByAlertNumber(ctx context.Context, number string) (*domain.Alert, error)
	Update(ctx context.Context, alert *domain.Alert) error
	List(ctx context.Context, filter AlertFilter) ([]*domain.Alert, error)
	Count(ctx context.Context, filter AlertFilter) (int64, error)
	Delete(ctx context.Context, id string) error
}

// AlertFilter represents filtering criteria for alert queries.
type AlertFilter struct {
	Status       []string
	Severity     []string
	AlertType    string
	SubjectID    string
	AssignedTo   string
	StartDate    *time.Time
	EndDate      *time.Time
	MinRiskScore int
	MaxRiskScore int
	Limit        int
	Offset       int
}

// ComplianceCheckRepository defines the interface for compliance check persistence.
type ComplianceCheckRepository interface {
	Create(ctx context.Context, check *domain.ComplianceCheck) error
	GetByID(ctx context.Context, id string) (*domain.ComplianceCheck, error)
	GetLatestBySubject(ctx context.Context, subjectID, checkType string) (*domain.ComplianceCheck, error)
	List(ctx context.Context, filter ComplianceCheckFilter) ([]*domain.ComplianceCheck, error)
	Delete(ctx context.Context, id string) error
}

// ComplianceCheckFilter represents filtering criteria for compliance check queries.
type ComplianceCheckFilter struct {
	SubjectID   string
	CheckType   string
	CheckResult []string
	RiskLevel   []string
	ScreeningList string
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Offset      int
}

// ComplianceReportRepository defines the interface for compliance report persistence.
type ComplianceReportRepository interface {
	Create(ctx context.Context, report *domain.ComplianceReport) error
	GetByID(ctx context.Context, id string) (*domain.ComplianceReport, error)
	Update(ctx context.Context, report *domain.ComplianceReport) error
	List(ctx context.Context, filter ComplianceReportFilter) ([]*domain.ComplianceReport, error)
	Delete(ctx context.Context, id string) error
}

// ComplianceReportFilter represents filtering criteria for compliance report queries.
type ComplianceReportFilter struct {
	ReportType  domain.ReportType
	Status      domain.ReportStatus
	StartDate   *time.Time
	EndDate     *time.Time
	GeneratedBy string
	Limit       int
	Offset      int
}

// FilingRecordRepository defines the interface for filing record persistence.
type FilingRecordRepository interface {
	Create(ctx context.Context, filing *domain.FilingRecord) error
	GetByID(ctx context.Context, id string) (*domain.FilingRecord, error)
	GetByReportID(ctx context.Context, reportID string) ([]*domain.FilingRecord, error)
	Update(ctx context.Context, filing *domain.FilingRecord) error
	List(ctx context.Context, filter FilingRecordFilter) ([]*domain.FilingRecord, error)
	Delete(ctx context.Context, id string) error
}

// FilingRecordFilter represents filtering criteria for filing record queries.
type FilingRecordFilter struct {
	ReportType   domain.ReportType
	FilingAgency string
	FilingStatus []string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// TransactionRepository defines the interface for transaction data access.
type TransactionRepository interface {
	GetByID(ctx context.Context, id string) (interface{}, error)
	GetByAccount(ctx context.Context, accountID string, startDate, endDate time.Time) ([]interface{}, error)
	GetByUser(ctx context.Context, userID string, startDate, endDate time.Time) ([]interface{}, error)
	GetHighRiskTransactions(ctx context.Context, startDate, endDate time.Time, limit int) ([]interface{}, error)
	GetTransactionPatterns(ctx context.Context, subjectID string) (interface{}, error)
}

// ScreeningRepository defines the interface for screening list operations.
type ScreeningRepository interface {
	Search(ctx context.Context, query string, lists []string) ([]interface{}, error)
	GetByList(ctx context.Context, listName string) ([]interface{}, error)
	AddToList(ctx context.Context, listName string, entity interface{}) error
	RemoveFromList(ctx context.Context, listName string, entityID string) error
}
