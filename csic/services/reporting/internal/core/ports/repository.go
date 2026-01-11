package ports

import (
	"context"

	"github.com/csic-platform/services/reporting/internal/core/domain"
)

// ReportRepository defines the interface for managing reports
type ReportRepository interface {
	// FindByID retrieves a report by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.Report, error)

	// FindAll retrieves all reports with optional filtering
	FindAll(ctx context.Context, req *ListReportsRequest) (*ListReportsResponse, error)

	// FindByType retrieves reports of a specific type
	FindByType(ctx context.Context, reportType domain.ReportType) ([]*domain.Report, error)

	// FindByEntity retrieves reports for a specific entity
	FindByEntity(ctx context.Context, entityID string) ([]*domain.Report, error)

	// FindRecent retrieves the most recent reports
	FindRecent(ctx context.Context, limit int) ([]*domain.Report, error)

	// Create creates a new report record
	Create(ctx context.Context, report *domain.Report) error

	// Update updates an existing report
	Update(ctx context.Context, report *domain.Report) error

	// UpdateStatus updates the status of a report
	UpdateStatus(ctx context.Context, id string, status domain.ReportStatus) error

	// IncrementDownload increments the download count
	IncrementDownload(ctx context.Context, id string) error

	// Delete removes a report
	Delete(ctx context.Context, id string) error

	// Archive archives a report
	Archive(ctx context.Context, id string) error
}

// ReportTemplateRepository defines the interface for managing report templates
type ReportTemplateRepository interface {
	// FindByID retrieves a template by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.ReportTemplate, error)

	// FindAll retrieves all templates
	FindAll(ctx context.Context) ([]*domain.ReportTemplate, error)

	// FindByType retrieves templates of a specific type
	FindByType(ctx context.Context, reportType domain.ReportType) ([]*domain.ReportTemplate, error)

	// FindDefault retrieves the default template for a type
	FindDefault(ctx context.Context, reportType domain.ReportType) (*domain.ReportTemplate, error)

	// Create creates a new template
	Create(ctx context.Context, template *domain.ReportTemplate) error

	// Update updates an existing template
	Update(ctx context.Context, template *domain.ReportTemplate) error

	// Delete removes a template
	Delete(ctx context.Context, id string) error
}

// ScheduledReportRepository defines the interface for managing scheduled reports
type ScheduledReportRepository interface {
	// FindByID retrieves a scheduled report by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.ScheduledReport, error)

	// FindAll retrieves all scheduled reports
	FindAll(ctx context.Context) ([]*domain.ScheduledReport, error)

	// FindEnabled retrieves all enabled scheduled reports
	FindEnabled(ctx context.Context) ([]*domain.ScheduledReport, error)

	// FindDueForExecution retrieves reports due for execution
	FindDueForExecution(ctx context.Context, before time.Time) ([]*domain.ScheduledReport, error)

	// Create creates a new scheduled report
	Create(ctx context.Context, scheduled *domain.ScheduledReport) error

	// Update updates an existing scheduled report
	Update(ctx context.Context, scheduled *domain.ScheduledReport) error

	// UpdateNextExecution updates the next execution time
	UpdateNextExecution(ctx context.Context, id string, nextRun time.Time) error

	// UpdateLastExecution updates the last execution time
	UpdateLastExecution(ctx context.Context, id string, lastRun time.Time) error

	// Delete removes a scheduled report
	Delete(ctx context.Context, id string) error
}

// Request/Response types

type ListReportsRequest struct {
	Type        *domain.ReportType `json:"type,omitempty"`
	Format      *domain.OutputFormat `json:"format,omitempty"`
	Status      *domain.ReportStatus `json:"status,omitempty"`
	EntityID    string              `json:"entity_id,omitempty"`
	StartDate   string              `json:"start_date,omitempty"`
	EndDate     string              `json:"end_date,omitempty"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
}

type ListReportsResponse struct {
	Reports   []*domain.Report `json:"reports"`
	Total     int              `json:"total"`
	Page      int              `json:"page"`
	PageSize  int              `json:"page_size"`
}
