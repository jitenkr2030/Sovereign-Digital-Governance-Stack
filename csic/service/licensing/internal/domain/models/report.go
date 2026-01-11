package models

import (
	"time"

	"github.com/google/uuid"
)

// ReportType represents different types of regulatory reports
type ReportType string

const (
	ReportTypeDailyLiquidity      ReportType = "daily_liquidity"
	ReportTypeMonthlyVolume       ReportType = "monthly_volume"
	ReportTypeQuarterlyFinancial  ReportType = "quarterly_financial"
	ReportTypeAnnualAudit         ReportType = "annual_audit"
	ReportTypeSuspiciousActivity  ReportType = "sar" // Suspicious Activity Report
	ReportTypeForm10K             ReportType = "10-k"
	ReportTypeForm10Q             ReportType = "10-q"
	ReportTypeFormADV             ReportType = "adv"
	ReportTypeAMLReview           ReportType = "aml_review"
	ReportTypeRiskAssessment      ReportType = "risk_assessment"
	ReportTypeComplianceCertificate ReportType = "compliance_certificate"
	ReportTypeLicenseRenewal      ReportType = "license_renewal"
	ReportTypeRegulatoryUpdate    ReportType = "regulatory_update"
	ReportTypeCustom              ReportType = "custom"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending     ReportStatus = "pending"
	ReportStatusGenerating  ReportStatus = "generating"
	ReportStatusReady       ReportStatus = "ready"
	ReportStatusReviewing   ReportStatus = "reviewing"
	ReportStatusSubmitted   ReportStatus = "submitted"
	ReportStatusRejected    ReportStatus = "rejected"
	ReportStatusArchived    ReportStatus = "archived"
)

// Report represents a regulatory compliance report
type Report struct {
	ID           string                 `json:"id" db:"id"`
	ReportType   ReportType             `json:"report_type" db:"report_type"`
	Title        string                 `json:"title" db:"title"`
	Description  string                 `json:"description" db:"description"`
	PeriodStart  time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd    time.Time              `json:"period_end" db:"period_end"`
	Status       ReportStatus           `json:"status" db:"status"`
	LicenseID    string                 `json:"license_id" db:"license_id"`
	ExchangeID   string                 `json:"exchange_id" db:"exchange_id"`
	GeneratedBy  string                 `json:"generated_by" db:"generated_by"`
	ReviewedBy   string                 `json:"reviewed_by" db:"reviewed_by"`
	SubmittedAt  *time.Time             `json:"submitted_at" db:"submitted_at"`
	SubmittedTo  string                 `json:"submitted_to" db:"submitted_to"`
	FilePath     string                 `json:"file_path" db:"file_path"`
	FileSize     int64                  `json:"file_size" db:"file_size"`
	MimeType     string                 `json:"mime_type" db:"mime_type"`
	Checksum     string                 `json:"checksum" db:"checksum"`
	TemplateID   string                 `json:"template_id" db:"template_id"`
	Parameters   map[string]interface{} `json:"parameters" db:"parameters"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	Sections     []ReportSection        `json:"sections" db:"-"`
	Signatures   []ReportSignature      `json:"signatures" db:"-"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportSection represents a section within a report
type ReportSection struct {
	ID          string                 `json:"id" db:"id"`
	ReportID    string                 `json:"report_id" db:"report_id"`
	SectionName string                 `json:"section_name" db:"section_name"`
	Order       int                    `json:"order" db:"order"`
	Content     string                 `json:"content" db:"content"`
	Data        map[string]interface{} `json:"data" db:"data"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// ReportSignature represents a signature on a report
type ReportSignature struct {
	ID          string    `json:"id" db:"id"`
	ReportID    string    `json:"report_id" db:"report_id"`
	SignerID    string    `json:"signer_id" db:"signer_id"`
	SignerName  string    `json:"signer_name" db:"signer_name"`
	SignerRole  string    `json:"signer_role" db:"signer_role"`
	SignedAt    time.Time `json:"signed_at" db:"signed_at"`
	Signature   string    `json:"signature" db:"signature"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
}

// ReportTemplate defines a template for generating reports
type ReportTemplate struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	ReportType  ReportType             `json:"report_type" db:"report_type"`
	Description string                 `json:"description" db:"description"`
	Template    string                 `json:"template" db:"template"` // Go template string
	Parameters  []TemplateParameter    `json:"parameters" db:"-"`
	Schedule    string                 `json:"schedule" db:"schedule"` // Cron expression
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Version     int                    `json:"version" db:"version"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// TemplateParameter defines a parameter for a report template
type TemplateParameter struct {
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"` // string, number, date, boolean, select
	Required    bool   `json:"required" db:"required"`
	Default     string `json:"default" db:"default"`
	Options     []string `json:"options" db:"-"` // For select type
	Description string `json:"description" db:"description"`
}

// ReportSchedule defines an automated report schedule
type ReportSchedule struct {
	ID           string                 `json:"id" db:"id"`
	TemplateID   string                 `json:"template_id" db:"template_id"`
	LicenseID    string                 `json:"license_id" db:"license_id"`
	CronSchedule string                 `json:"cron_schedule" db:"cron_schedule"`
	TimeZone     string                 `json:"time_zone" db:"time_zone"`
	Recipients   []string               `json:"recipients" db:"-"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	LastRunAt    *time.Time             `json:"last_run_at" db:"last_run_at"`
	NextRunAt    *time.Time             `json:"next_run_at" db:"next_run_at"`
	Parameters   map[string]interface{} `json:"parameters" db:"parameters"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportHistory represents the generation history of reports
type ReportHistory struct {
	ID           string    `json:"id" db:"id"`
	ReportID     string    `json:"report_id" db:"report_id"`
	Action       string    `json:"action" db:"action"` // generated, reviewed, submitted, rejected
	PerformedBy  string    `json:"performed_by" db:"performed_by"`
	Details      string    `json:"details" db:"details"`
	Duration     int64     `json:"duration_ms" db:"duration_ms"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// NewReport creates a new report record
func NewReport(reportType ReportType, periodStart, periodEnd time.Time, licenseID, exchangeID string) *Report {
	return &Report{
		ID:          uuid.New().String(),
		ReportType:  reportType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      ReportStatusPending,
		LicenseID:   licenseID,
		ExchangeID:  exchangeID,
		Sections:    []ReportSection{},
		Signatures:  []ReportSignature{},
		Metadata:    make(map[string]interface{}),
		Parameters:  make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// IsGenerated checks if the report has been generated
func (r *Report) IsGenerated() bool {
	return r.Status == ReportStatusReady ||
		r.Status == ReportStatusReviewing ||
		r.Status == ReportStatusSubmitted
}

// IsSubmittable checks if the report can be submitted to regulators
func (r *Report) IsSubmittable() bool {
	return r.Status == ReportStatusReady ||
		r.Status == ReportStatusReviewing
}

// ReportStats represents aggregated report statistics
type ReportStats struct {
	TotalReports        int64                  `json:"total_reports"`
	PendingReports      int64                  `json:"pending_reports"`
	SubmittedReports    int64                  `json:"submitted_reports"`
	RejectedReports     int64                  `json:"rejected_reports"`
	ByType              map[ReportType]int64   `json:"by_type"`
	ByLicense           map[string]int64       `json:"by_license"`
	AverageGenerationTime float64              `json:"average_generation_time_seconds"`
	OverdueReports      int64                  `json:"overdue_reports"`
}

// ComplianceReport represents a comprehensive compliance status report
type ComplianceReport struct {
	ID               string                 `json:"id" db:"id"`
	LicenseID        string                 `json:"license_id" db:"license_id"`
	ExchangeID       string                 `json:"exchange_id" db:"exchange_id"`
	ReportDate       time.Time              `json:"report_date" db:"report_date"`
	PeriodStart      time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd        time.Time              `json:"period_end" db:"period_end"`
	ComplianceScore  float64                `json:"compliance_score" db:"compliance_score"`
	RiskLevel        string                 `json:"risk_level" db:"risk_level"`
	TotalViolations  int64                  `json:"total_violations" db:"total_violations"`
	OpenViolations   int64                  `json:"open_violations" db:"open_violations"`
	ResolvedViolations int64                `json:"resolved_violations" db:"resolved_violations"`
	PendingReports   int64                  `json:"pending_reports" db:"pending_reports"`
	UpcomingRenewals int64                  `json:"upcoming_renewals" db:"upcoming_renewals"`
	KeyFindings      []string               `json:"key_findings" db:"-"`
	Recommendations  []string               `json:"recommendations" db:"-"`
	GeneratedBy      string                 `json:"generated_by" db:"generated_by"`
	ReviewedBy       string                 `json:"reviewed_by" db:"reviewed_by"`
	Status           string                 `json:"status" db:"status"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}
