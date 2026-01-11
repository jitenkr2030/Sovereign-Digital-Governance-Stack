package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReportType represents the type of report
type ReportType string

const (
	ReportTypeCompliance   ReportType = "COMPLIANCE"
	ReportTypeRisk         ReportType = "RISK"
	ReportTypeAudit        ReportType = "AUDIT"
	ReportTypeExposure     ReportType = "EXPOSURE"
	ReportTypeTransaction  ReportType = "TRANSACTION"
	ReportTypePerformance  ReportType = "PERFORMANCE"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "PENDING"
	ReportStatusProcessing ReportStatus = "PROCESSING"
	ReportStatusCompleted  ReportStatus = "COMPLETED"
	ReportStatusFailed     ReportStatus = "FAILED"
	ReportStatusArchived   ReportStatus = "ARCHIVED"
)

// OutputFormat represents the output format of a report
type OutputFormat string

const (
	OutputFormatPDF  OutputFormat = "PDF"
	OutputFormatCSV  OutputFormat = "CSV"
	OutputFormatJSON OutputFormat = "JSON"
	OutputFormatXLSX OutputFormat = "XLSX"
	OutputFormatHTML OutputFormat = "HTML"
)

// Report represents a generated report
type Report struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	Name          string        `json:"name" db:"name"`
	Type          ReportType    `json:"type" db:"type"`
	Description   string        `json:"description" db:"description"`
	Status        ReportStatus  `json:"status" db:"status"`
	Format        OutputFormat  `json:"format" db:"format"`
	GeneratedBy   string        `json:"generated_by" db:"generated_by"`
	EntityID      string        `json:"entity_id,omitempty" db:"entity_id"`
	EntityType    string        `json:"entity_type,omitempty" db:"entity_type"`
	DateRange     TimeRange     `json:"date_range" db:"-"`
	StartDate     time.Time     `json:"start_date" db:"start_date"`
	EndDate       time.Time     `json:"end_date" db:"end_date"`
	Filters       ReportFilters `json:"filters,omitempty" db:"-"`
	FiltersJSON   string        `json:"-" db:"filters"`
	FilePath      string        `json:"file_path,omitempty" db:"file_path"`
	FileSize      int64         `json:"file_size,omitempty" db:"file_size"`
	DownloadCount int           `json:"download_count" db:"download_count"`
	ErrorMessage  string        `json:"error_message,omitempty" db:"error_message"`
	ScheduledID   *uuid.UUID    `json:"scheduled_id,omitempty" db:"scheduled_id"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
}

// TimeRange represents a time range for reports
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ReportFilters represents filters for report data
type ReportFilters struct {
	Symbols      []string `json:"symbols,omitempty"`
	Severities   []string `json:"severities,omitempty"`
	Statuses     []string `json:"statuses,omitempty"`
	Sources      []string `json:"sources,omitempty"`
	EntityTypes  []string `json:"entity_types,omitempty"`
	SearchQuery  string   `json:"search_query,omitempty"`
}

// ReportTemplate defines a reusable report template
type ReportTemplate struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Type        ReportType   `json:"type" db:"type"`
	Format      OutputFormat `json:"format" db:"format"`
	Content     string       `json:"content" db:"content"`
	Parameters  []TemplateParameter `json:"parameters" db:"-"`
	ParametersJSON string     `json:"-" db:"parameters"`
	IsDefault   bool         `json:"is_default" db:"is_default"`
	CreatedBy   string       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// TemplateParameter defines a parameter for a report template
type TemplateParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

// ScheduledReport defines an automated scheduled report
type ScheduledReport struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	Name         string       `json:"name" db:"name"`
	Description  string       `json:"description" db:"description"`
	TemplateID   uuid.UUID    `json:"template_id" db:"template_id"`
	Type         ReportType   `json:"type" db:"type"`
	Format       OutputFormat `json:"format" db:"format"`
	CronSchedule string       `json:"cron_schedule" db:"cron_schedule"`
	TimeZone     string       `json:"time_zone" db:"time_zone"`
	Enabled      bool         `json:"enabled" db:"enabled"`
	EntityID     string       `json:"entity_id,omitempty" db:"entity_id"`
	Recipients   []string     `json:"recipients" db:"-"`
	RecipientsJSON string     `json:"-" db:"recipients"`
	LastRunAt    *time.Time   `json:"last_run_at,omitempty" db:"last_run_at"`
	NextRunAt    *time.Time   `json:"next_run_at,omitempty" db:"next_run_at"`
	CreatedBy    string       `json:"created_by" db:"created_by"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

// ComplianceReport represents a compliance-specific report
type ComplianceReport struct {
	ReportID         uuid.UUID              `json:"report_id"`
	GeneratedAt      time.Time              `json:"generated_at"`
	Period           TimeRange              `json:"period"`
	TotalChecks      int                    `json:"total_checks"`
	PassedChecks     int                    `json:"passed_checks"`
	FailedChecks     int                    `json:"failed_checks"`
	ComplianceScore  decimal.Decimal        `json:"compliance_score"`
	Violations       []ComplianceViolation  `json:"violations"`
	Obligations      []ComplianceObligation `json:"obligations"`
	Recommendations  []string               `json:"recommendations"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID            uuid.UUID   `json:"id"`
	Type          string      `json:"type"`
	Severity      string      `json:"severity"`
	Description   string      `json:"description"`
	DetectedAt    time.Time   `json:"detected_at"`
	ResolvedAt    *time.Time  `json:"resolved_at,omitempty"`
	Status        string      `json:"status"`
	EntityID      string      `json:"entity_id"`
	EntityType    string      `json:"entity_type"`
}

// ComplianceObligation represents a compliance obligation
type ComplianceObligation struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	DueDate     time.Time   `json:"due_date"`
	Status      string      `json:"status"`
	AssignedTo  string      `json:"assigned_to"`
}

// RiskReport represents a risk-specific report
type RiskReport struct {
	ReportID           uuid.UUID           `json:"report_id"`
	GeneratedAt        time.Time           `json:"generated_at"`
	Period             TimeRange           `json:"period"`
	OverallRiskScore   decimal.Decimal     `json:"overall_risk_score"`
	RiskLevel          string              `json:"risk_level"`
	TotalAlerts        int                 `json:"total_alerts"`
	ActiveAlerts       int                 `json:"active_alerts"`
	ResolvedAlerts     int                 `json:"resolved_alerts"`
	AlertsBySeverity   map[string]int      `json:"alerts_by_severity"`
	TopRisks           []RiskItem          `json:"top_risks"`
	TrendData          []RiskTrendPoint    `json:"trend_data"`
	Recommendations    []string            `json:"recommendations"`
}

// RiskItem represents a significant risk item
type RiskItem struct {
	Symbol       string          `json:"symbol"`
	RiskScore    decimal.Decimal `json:"risk_score"`
	RiskLevel    string          `json:"risk_level"`
	Exposure     decimal.Decimal `json:"exposure"`
	AlertsCount  int             `json:"alerts_count"`
}

// RiskTrendPoint represents a data point for risk trends
type RiskTrendPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	RiskScore   decimal.Decimal `json:"risk_score"`
	AlertsCount int       `json:"alerts_count"`
}

// ReportStatistics contains aggregated report statistics
type ReportStatistics struct {
	TotalReports      int64            `json:"total_reports"`
	ReportsToday      int64            `json:"reports_today"`
	ReportsThisWeek   int64            `json:"reports_this_week"`
	ReportsThisMonth  int64            `json:"reports_this_month"`
	DownloadsToday    int64            `json:"downloads_today"`
	StorageUsed       int64            `json:"storage_used"`
	ByType            map[string]int64 `json:"by_type"`
	ByFormat          map[string]int64 `json:"by_format"`
}
