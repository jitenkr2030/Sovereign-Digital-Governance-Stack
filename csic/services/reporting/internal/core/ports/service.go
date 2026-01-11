package ports

import (
	"context"
	"io"

	"github.com/csic-platform/services/reporting/internal/core/domain"
)

// ReportService defines the business logic for report management
type ReportService interface {
	// GenerateReport generates a new report
	GenerateReport(ctx context.Context, req *GenerateReportRequest) (*GenerateReportResponse, error)

	// GetReport retrieves a report by ID
	GetReport(ctx context.Context, id string) (*domain.Report, error)

	// ListReports lists reports with filtering
	ListReports(ctx context.Context, req *ListReportsRequest) (*ListReportsResponse, error)

	// DownloadReport downloads a report file
	DownloadReport(ctx context.Context, id string) (io.ReadCloser, error)

	// DeleteReport deletes a report
	DeleteReport(ctx context.Context, id string) error

	// ArchiveReport archives a report
	ArchiveReport(ctx context.Context, id string) error

	// GetStatistics returns report statistics
	GetStatistics(ctx context.Context) (*domain.ReportStatistics, error)
}

// ReportGenerator defines the interface for generating report content
type ReportGenerator interface {
	// GenerateComplianceReport generates a compliance report
	GenerateComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*domain.ComplianceReport, error)

	// GenerateRiskReport generates a risk report
	GenerateRiskReport(ctx context.Context, req *RiskReportRequest) (*domain.RiskReport, error)

	// GenerateAuditReport generates an audit report
	GenerateAuditReport(ctx context.Context, req *AuditReportRequest) (*AuditReportData, error)

	// GenerateExposureReport generates an exposure report
	GenerateExposureReport(ctx context.Context, req *ExposureReportRequest) (*ExposureReportData, error)
}

// ReportFormatter defines the interface for formatting reports
type ReportFormatter interface {
	// FormatToPDF formats report data to PDF
	FormatToPDF(data interface{}) ([]byte, error)

	// FormatToCSV formats report data to CSV
	FormatToCSV(data interface{}) ([]byte, error)

	// FormatToJSON formats report data to JSON
	FormatToJSON(data interface{}) ([]byte, error)

	// FormatToHTML formats report data to HTML
	FormatToHTML(data interface{}) ([]byte, error)
}

// ReportStorage defines the interface for storing report files
type ReportStorage interface {
	// Store stores a report file
	Store(ctx context.Context, id string, content []byte, format domain.OutputFormat) error

	// Retrieve retrieves a report file
	Retrieve(ctx context.Context, id string) (io.ReadCloser, error)

	// Delete deletes a report file
	Delete(ctx context.Context, id string) error

	// GetPath returns the file path for a report
	GetPath(id string, format domain.OutputFormat) string
}

// TemplateService defines the business logic for report templates
type TemplateService interface {
	// CreateTemplate creates a new report template
	CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*CreateTemplateResponse, error)

	// GetTemplate retrieves a template by ID
	GetTemplate(ctx context.Context, id string) (*domain.ReportTemplate, error)

	// ListTemplates lists all templates
	ListTemplates(ctx context.Context, req *ListTemplatesRequest) (*ListTemplatesResponse, error)

	// UpdateTemplate updates a template
	UpdateTemplate(ctx context.Context, req *UpdateTemplateRequest) error

	// DeleteTemplate deletes a template
	DeleteTemplate(ctx context.Context, id string) error
}

// SchedulerService defines the business logic for scheduled reports
type SchedulerService interface {
	// CreateScheduledReport creates a new scheduled report
	CreateScheduledReport(ctx context.Context, req *CreateScheduledRequest) (*CreateScheduledResponse, error)

	// GetScheduledReport retrieves a scheduled report
	GetScheduledReport(ctx context.Context, id string) (*domain.ScheduledReport, error)

	// ListScheduledReports lists all scheduled reports
	ListScheduledReports(ctx context.Context) ([]*domain.ScheduledReport, error)

	// UpdateScheduledReport updates a scheduled report
	UpdateScheduledReport(ctx context.Context, req *UpdateScheduledRequest) error

	// DeleteScheduledReport deletes a scheduled report
	DeleteScheduledReport(ctx context.Context, id string) error

	// EnableScheduledReport enables a scheduled report
	EnableScheduledReport(ctx context.Context, id string) error

	// DisableScheduledReport disables a scheduled report
	DisableScheduledReport(ctx context.Context, id string) error

	// ExecuteScheduledReport executes a scheduled report immediately
	ExecuteScheduledReport(ctx context.Context, id string) error
}

// Request/Response types for ReportService

type GenerateReportRequest struct {
	Type        domain.ReportType    `json:"type" validate:"required"`
	Format      domain.OutputFormat  `json:"format" validate:"required"`
	Name        string               `json:"name" validate:"required"`
	Description string               `json:"description,omitempty"`
	StartDate   string               `json:"start_date" validate:"required"`
	EndDate     string               `json:"end_date" validate:"required"`
	EntityID    string               `json:"entity_id,omitempty"`
	EntityType  string               `json:"entity_type,omitempty"`
	Filters     domain.ReportFilters `json:"filters,omitempty"`
}

type GenerateReportResponse struct {
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}

// Request/Response types for ReportGenerator

type ComplianceReportRequest struct {
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	EntityID   string   `json:"entity_id,omitempty"`
	EntityType string   `json:"entity_type,omitempty"`
	Filters    []string `json:"filters,omitempty"`
}

type RiskReportRequest struct {
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	EntityID   string   `json:"entity_id,omitempty"`
	EntityType string   `json:"entity_type,omitempty"`
	Filters    []string `json:"filters,omitempty"`
}

type AuditReportRequest struct {
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	EntityID   string   `json:"entity_id,omitempty"`
	EntityType string   `json:"entity_type,omitempty"`
	Actions    []string `json:"actions,omitempty"`
	Users      []string `json:"users,omitempty"`
}

type ExposureReportRequest struct {
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	EntityID   string `json:"entity_id,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
}

// AuditReportData represents audit report content
type AuditReportData struct {
	Period        domain.TimeRange  `json:"period"`
	TotalEvents   int               `json:"total_events"`
	EventsByType  map[string]int    `json:"events_by_type"`
	EventsByUser  map[string]int    `json:"events_by_user"`
	CriticalEvents []AuditEvent     `json:"critical_events"`
}

// AuditEvent represents a single audit event
type AuditEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	UserID      string    `json:"user_id"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	Status      string    `json:"status"`
	IPAddress   string    `json:"ip_address"`
	Details     string    `json:"details"`
}

// ExposureReportData represents exposure report content
type ExposureReportData struct {
	Period         domain.TimeRange      `json:"period"`
	TotalExposure  decimal.Decimal       `json:"total_exposure"`
	ExposureBySymbol map[string]decimal.Decimal `json:"exposure_by_symbol"`
	TopPositions   []ExposurePosition    `json:"top_positions"`
	Concentration  ConcentrationData     `json:"concentration"`
}

// ExposurePosition represents a single position
type ExposurePosition struct {
	Symbol      string          `json:"symbol"`
	Side        string          `json:"side"`
	Quantity    decimal.Decimal `json:"quantity"`
	Notional    decimal.Decimal `json:"notional"`
	PnL         decimal.Decimal `json:"pnl"`
	Leverage    decimal.Decimal `json:"leverage"`
}

// ConcentrationData represents concentration risk data
type ConcentrationData struct {
	LargestPosition    decimal.Decimal `json:"largest_position"`
	ConcentrationRatio decimal.Decimal `json:"concentration_ratio"`
	RiskLevel          string          `json:"risk_level"`
}

// Request/Response types for TemplateService

type CreateTemplateRequest struct {
	Name        string                `json:"name" validate:"required"`
	Description string                `json:"description"`
	Type        domain.ReportType     `json:"type" validate:"required"`
	Format      domain.OutputFormat   `json:"format" validate:"required"`
	Content     string                `json:"content" validate:"required"`
	Parameters  []TemplateParameter   `json:"parameters"`
}

type CreateTemplateResponse struct {
	TemplateID string `json:"template_id"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
}

type ListTemplatesRequest struct {
	Type *domain.ReportType `json:"type,omitempty"`
	Page int                `json:"page"`
	PageSize int            `json:"page_size"`
}

type ListTemplatesResponse struct {
	Templates []*domain.ReportTemplate `json:"templates"`
	Total     int                      `json:"total"`
	Page      int                      `json:"page"`
	PageSize  int                      `json:"page_size"`
}

type UpdateTemplateRequest struct {
	ID          string  `json:"id" validate:"required"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Content     *string `json:"content,omitempty"`
	Parameters  []TemplateParameter `json:"parameters,omitempty"`
}

// Request/Response types for SchedulerService

type CreateScheduledRequest struct {
	Name         string               `json:"name" validate:"required"`
	Description  string               `json:"description"`
	TemplateID   string               `json:"template_id" validate:"required"`
	Type         domain.ReportType    `json:"type" validate:"required"`
	Format       domain.OutputFormat  `json:"format" validate:"required"`
	CronSchedule string               `json:"cron_schedule" validate:"required"`
	TimeZone     string               `json:"time_zone"`
	EntityID     string               `json:"entity_id,omitempty"`
	Recipients   []string             `json:"recipients"`
}

type CreateScheduledResponse struct {
	ScheduledID string `json:"scheduled_id"`
	NextRunAt   string `json:"next_run_at"`
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
}

type UpdateScheduledRequest struct {
	ID           string    `json:"id" validate:"required"`
	Name         *string   `json:"name,omitempty"`
	Description  *string   `json:"description,omitempty"`
	CronSchedule *string   `json:"cron_schedule,omitempty"`
	TimeZone     *string   `json:"time_zone,omitempty"`
	Recipients   []string  `json:"recipients,omitempty"`
}
