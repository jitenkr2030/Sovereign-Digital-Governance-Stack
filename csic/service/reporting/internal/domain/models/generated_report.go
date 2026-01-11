package domain

import (
	"time"

	"github.com/google/uuid"
)

// GeneratedReport represents a generated regulatory report instance
type GeneratedReport struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	TemplateID        uuid.UUID       `json:"template_id" db:"template_id"`
	TemplateName      string          `json:"template_name" db:"template_name"`
	ReportType        ReportType      `json:"report_type" db:"report_type"`
	RegulatorID       uuid.UUID       `json:"regulator_id" db:"regulator_id"`
	RegulatorName     string          `json:"regulator_name" db:"regulator_name"`
	Status            ReportStatus    `json:"status" db:"status"`
	Version           string          `json:"version" db:"version"`
	
	// Report metadata
	Title             string          `json:"title" db:"title"`
	Description       string          `json:"description" db:"description"`
	PeriodStart       time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time       `json:"period_end" db:"period_end"`
	ReportDate        time.Time       `json:"report_date" db:"report_date"`
	
	// Generation details
	GeneratedAt       *time.Time      `json:"generated_at" db:"generated_at"`
	GeneratedBy       string          `json:"generated_by" db:"generated_by"`
	GenerationMethod  string          `json:"generation_method" db:"generation_method"` // scheduled, manual, triggered
	TriggeredBy       string          `json:"triggered_by" db:"triggered_by"`
	
	// Output file information
	OutputFormats     []ReportFormat  `json:"output_formats" db:"output_formats"`
	Files             []ReportFile    `json:"files" db:"files"`
	TotalSize         int64           `json:"total_size" db:"total_size"`
	
	// Data and checksum
	DataSnapshot      []byte          `json:"-" db:"data_snapshot"` // Compressed data
	Checksum          string          `json:"checksum" db:"checksum"`
	
	// Processing metadata
	ProcessingTime    *int64          `json:"processing_time" db:"processing_time"` // milliseconds
	RetryCount        int             `json:"retry_count" db:"retry_count"`
	LastRetryAt       *time.Time      `json:"last_retry_at" db:"last_retry_at"`
	FailureReason     string          `json:"failure_reason" db:"failure_reason"`
	
	// Validation and approval
	IsValidated       bool            `json:"is_validated" db:"is_validated"`
	ValidatedAt       *time.Time      `json:"validated_at" db:"validated_at"`
	ValidatedBy       string          `json:"validated_by" db:"validated_by"`
	IsApproved        bool            `json:"is_approved" db:"is_approved"`
	ApprovedAt        *time.Time      `json:"approved_at" db:"approved_at"`
	ApprovedBy        string          `json:"approved_by" db:"approved_by"`
	
	// Submission tracking
	SubmissionStatus  string          `json:"submission_status" db:"submission_status"`
	SubmittedAt       *time.Time      `json:"submitted_at" db:"submitted_at"`
	SubmittedBy       string          `json:"submitted_by" db:"submitted_by"`
	SubmissionRef     string          `json:"submission_ref" db:"submission_ref"`
	AcknowledgmentRef string          `json:"acknowledgment_ref" db:"acknowledgment_ref"`
	
	// Archival
	IsArchived        bool            `json:"is_archived" db:"is_archived"`
	ArchivedAt        *time.Time      `json:"archived_at" db:"archived_at"`
	RetentionPeriod   int             `json:"retention_period" db:"retention_period"` // days until deletion
	
	// Audit trail
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
	CreatedBy         string          `json:"created_by" db:"created_by"`
	UpdatedBy         string          `json:"updated_by" db:"updated_by"`
	
	// Custom fields for regulator-specific data
	CustomFields      map[string]interface{} `json:"custom_fields" db:"custom_fields"`
}

// ReportFile represents an output file for a generated report
type ReportFile struct {
	ID           uuid.UUID   `json:"id"`
	ReportID     uuid.UUID   `json:"report_id"`
	Format       ReportFormat `json:"format"`
	Filename     string      `json:"filename"`
	FilePath     string      `json:"file_path"`
	FileSize     int64       `json:"file_size"`
	Checksum     string      `json:"checksum"`
	ChecksumType string      `json:"checksum_type"` // sha256, md5
	Encryption   string      `json:"encryption"` // none, aes256
	EncryptedAt  *time.Time  `json:"encrypted_at"`
	CreatedAt    time.Time   `json:"created_at"`
}

// ReportQueueItem represents a report in the generation queue
type ReportQueueItem struct {
	ID              uuid.UUID       `json:"id"`
	ReportID        uuid.UUID       `json:"report_id"`
	TemplateID      uuid.UUID       `json:"template_id"`
	Priority        int             `json:"priority"` // 1-10, higher = more urgent
	Status          string          `json:"status"` // queued, processing, completed, failed
	ScheduledAt     time.Time       `json:"scheduled_at"`
	StartedAt       *time.Time      `json:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at"`
	WorkerID        string          `json:"worker_id"`
	Attempts        int             `json:"attempts"`
	MaxAttempts     int             `json:"max_attempts"`
	LastError       string          `json:"last_error"`
	CreatedAt       time.Time       `json:"created_at"`
}

// ReportHistory represents the history of report changes
type ReportHistory struct {
	ID            uuid.UUID   `json:"id"`
	ReportID      uuid.UUID   `json:"report_id"`
	Action        string      `json:"action"` // created, regenerated, validated, approved, submitted, archived
	PreviousValue string      `json:"previous_value"`
	NewValue      string      `json:"new_value"`
	ChangedBy     string      `json:"changed_by"`
	ChangedAt     time.Time   `json:"changed_at"`
	Reason        string      `json:"reason"`
	IPAddress     string      `json:"ip_address"`
	UserAgent     string      `json:"user_agent"`
}

// ReportSchedule represents a scheduled report generation
type ReportSchedule struct {
	ID              uuid.UUID       `json:"id"`
	TemplateID      uuid.UUID       `json:"template_id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Frequency       ReportFrequency `json:"frequency"`
	CronExpression  string          `json:"cron_expression"`
	TimeZone        string          `json:"time_zone"` // e.g., "America/New_York"
	
	// Scheduling options
	StartDate       *time.Time      `json:"start_date"`
	EndDate         *time.Time      `json:"end_date"`
	SkipWeekends    bool            `json:"skip_weekends"`
	SkipHolidays    bool            `json:"skip_holidays"`
	
	// Generation settings
	Priority        int             `json:"priority"`
	OutputFormats   []ReportFormat  `json:"output_formats"`
	
	// Notification settings
	NotifyOnSuccess bool            `json:"notify_on_success"`
	NotifyOnFailure bool            `json:"notify_on_failure"`
	Recipients      []string        `json:"recipients"`
	
	// State
	IsActive        bool            `json:"is_active"`
	LastRunAt       *time.Time      `json:"last_run_at"`
	NextRunAt       *time.Time      `json:"next_run_at"`
	RunCount        int             `json:"run_count"`
	FailureCount    int             `json:"failure_count"`
	
	CreatedBy       string          `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ReportFilter represents filter criteria for querying reports
type ReportFilter struct {
	RegulatorIDs    []uuid.UUID     `json:"regulator_ids"`
	ReportTypes     []ReportType    `json:"report_types"`
	Statuses        []ReportStatus  `json:"statuses"`
	PeriodStart     *time.Time      `json:"period_start"`
	PeriodEnd       *time.Time      `json:"period_end"`
	GeneratedAfter  *time.Time      `json:"generated_after"`
	GeneratedBefore *time.Time      `json:"generated_before"`
	GeneratedBy     string          `json:"generated_by"`
	TemplateID      *uuid.UUID      `json:"template_id"`
	SearchQuery     string          `json:"search_query"`
	Tags            []string        `json:"tags"`
	
	// Pagination
	Page            int             `json:"page"`
	PageSize        int             `json:"page_size"`
	
	// Sorting
	SortBy          string          `json:"sort_by"` // created_at, report_date, title
	SortOrder       string          `json:"sort_order"` // asc, desc
}
