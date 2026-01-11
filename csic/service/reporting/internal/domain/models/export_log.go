package domain

import (
	"time"

	"github.com/google/uuid"
)

// ExportStatus represents the status of an export operation
type ExportStatus string

const (
	ExportStatusPending    ExportStatus = "pending"
	ExportStatusProcessing ExportStatus = "processing"
	ExportStatusCompleted  ExportStatus = "completed"
	ExportStatusFailed     ExportStatus = "failed"
	ExportStatusCancelled  ExportStatus = "cancelled"
)

// ExportType represents the type of export
type ExportType string

const (
	ExportTypeReport         ExportType = "report"
	ExportTypeData           ExportType = "data"
	ExportTypeAuditTrail     ExportType = "audit_trail"
	ExportTypeComplianceData ExportType = "compliance_data"
	ExportTypeCustom         ExportType = "custom"
)

// EncryptionType represents the encryption method for exports
type EncryptionType string

const (
	EncryptionNone   EncryptionType = "none"
	EncryptionAES256 EncryptionType = "aes256"
	EncryptionPGP    EncryptionType = "pgp"
)

// ExportLog represents a secure export audit log entry
type ExportLog struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	ReportID          *uuid.UUID      `json:"report_id" db:"report_id"`
	ExportType        ExportType      `json:"export_type" db:"export_type"`
	
	// User information
	UserID            string          `json:"user_id" db:"user_id"`
	UserName          string          `json:"user_name" db:"user_name"`
	UserEmail         string          `json:"user_email" db:"user_email"`
	UserRole          string          `json:"user_role" db:"user_role"`
	UserOrganization  string          `json:"user_organization" db:"user_organization"`
	
	// Regulator context (if applicable)
	RegulatorID       *uuid.UUID      `json:"regulator_id" db:"regulator_id"`
	RegulatorName     string          `json:"regulator_name" db:"regulator_name"`
	
	// Export details
	ExportFormat      ReportFormat    `json:"export_format" db:"export_format"`
	FileName          string          `json:"file_name" db:"file_name"`
	FilePath          string          `json:"file_path" db:"file_path"`
	FileSize          int64           `json:"file_size" db:"file_size"`
	
	// Checksum for integrity verification
	Checksum          string          `json:"checksum" db:"checksum"`
	ChecksumType      string          `json:"checksum_type" db:"checksum_type"` // sha256, md5
	
	// Encryption details
	EncryptionType    EncryptionType  `json:"encryption_type" db:"encryption_type"`
	EncryptedKeyID    string          `json:"encrypted_key_id" db:"encrypted_key_id"`
	EncryptedAt       *time.Time      `json:"encrypted_at" db:"encrypted_at"`
	
	// Access tracking
	RequestedAt       time.Time       `json:"requested_at" db:"requested_at"`
	StartedAt         *time.Time      `json:"started_at" db:"started_at"`
	CompletedAt       *time.Time      `json:"completed_at" db:"completed_at"`
	ExpiresAt         *time.Time      `json:"expires_at" db:"expires_at"`
	DownloadedAt      *time.Time      `json:"downloaded_at" db:"downloaded_at"`
	DownloadCount     int             `json:"download_count" db:"download_count"`
	
	// Access control
	IPAddress         string          `json:"ip_address" db:"ip_address"`
	UserAgent         string          `json:"user_agent" db:"user_agent"`
	SessionID         string          `json:"session_id" db:"session_id"`
	AccessTokenID     string          `json:"access_token_id" db:"access_token_id"`
	
	// Authorization
	IsAuthorized      bool            `json:"is_authorized" db:"is_authorized"`
	AuthorizedBy      string          `json:"authorized_by" db:"authorized_by"`
	AuthorizationTime *time.Time      `json:"authorization_time" db:"authorization_time"`
	AuthorizationRef  string          `json:"authorization_ref" db:"authorization_ref"`
	
	// Status and error handling
	Status            ExportStatus    `json:"status" db:"status"`
	Progress          int             `json:"progress" db:"progress"` // percentage
	FailureReason     string          `json:"failure_reason" db:"failure_reason"`
	ErrorCode         string          `json:"error_code" db:"error_code"`
	RetryCount        int             `json:"retry_count" db:"retry_count"`
	
	// Data scope
	DataScope         ExportDataScope `json:"data_scope" db:"data_scope"`
	FilterCriteria    string          `json:"filter_criteria" db:"filter_criteria"`
	RecordCount       int             `json:"record_count" db:"record_count"`
	
	// Security flags
	ContainsSensitive bool            `json:"contains_sensitive" db:"contains_sensitive"`
	ContainsPII       bool            `json:"contains_pii" db:"contains_pii"`
	RedactionApplied  bool            `json:"redaction_applied" db:"redaction_applied"`
	
	// Audit trail
	AuditTrailID      uuid.UUID       `json:"audit_trail_id" db:"audit_trail_id"`
	
	// Additional metadata
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	Notes             string          `json:"notes" db:"notes"`
	
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// ExportDataScope defines the scope of exported data
type ExportDataScope struct {
	DateRangeStart   *time.Time  `json:"date_range_start"`
	DateRangeEnd     *time.Time  `json:"date_range_end"`
	EntityTypes      []string    `json:"entity_types"` // transactions, users, accounts
	IncludedFields   []string    `json:"included_fields"`
	ExcludedFields   []string    `json:"excluded_fields"`
	GeographicScope  []string    `json:"geographic_scope"` // countries, regions
	Filters          []string    `json:"filters"`
}

// ExportRequest represents a request to export data
type ExportRequest struct {
	ID              uuid.UUID       `json:"id"`
	ReportID        *uuid.UUID      `json:"report_id"`
	ExportType      ExportType      `json:"export_type"`
	
	// Requester information
	UserID          string          `json:"user_id"`
	UserEmail       string          `json:"user_email"`
	UserName        string          `json:"user_name"`
	
	// Export parameters
	Format          ReportFormat    `json:"format"`
	Filters         map[string]interface{} `json:"filters"`
	DateRange       DateRange       `json:"date_range"`
	
	// Security options
	Encryption      EncryptionType  `json:"encryption"`
	Password        string          `json:"password"` // For password-protected exports
	Watermark       bool            `json:"watermark"`
	WatermarkText   string          `json:"watermark_text"`
	
	// Delivery options
	DeliveryMethod  string          `json:"delivery_method"` // download, email, sftp, api
	DeliveryAddress string          `json:"delivery_address"`
	
	// Access control
	AccessExpiresIn int             `json:"access_expires_in"` // hours
	MaxDownloads    int             `json:"max_downloads"`
	
	// Notification
	NotifyOnComplete bool           `json:"notify_on_complete"`
	NotifyOnError    bool           `json:"notify_on_error"`
	
	// Audit requirements
	RequireApproval  bool           `json:"require_approval"`
	ApprovalNotes    string         `json:"approval_notes"`
	
	CreatedAt       time.Time       `json:"created_at"`
}

// DateRange represents a date range for exports
type DateRange struct {
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	RelativeType string     `json:"relative_type"` // last_7d, last_30d, last_quarter, etc.
	RelativeValue int       `json:"relative_value"`
}

// ExportSession represents an active export session for streaming
type ExportSession struct {
	ID              uuid.UUID       `json:"id"`
	ExportLogID     uuid.UUID       `json:"export_log_id"`
	UserID          string          `json:"user_id"`
	SessionToken    string          `json:"session_token"`
	
	Status          ExportStatus    `json:"status"`
	Progress        int             `json:"progress"`
	RecordsProcessed int            `json:"records_processed"`
	TotalRecords    int             `json:"total_records"`
	
	// Stream information
	StreamID        string          `json:"stream_id"`
	ChunkSize       int             `json:"chunk_size"`
	CurrentOffset   int64           `json:"current_offset"`
	FileSize        int64           `json:"file_size"`
	
	// Security
	IPAddress       string          `json:"ip_address"`
	UserAgent       string          `json:"user_agent"`
	
	// Timing
	CreatedAt       time.Time       `json:"created_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
	LastActivityAt  time.Time       `json:"last_activity_at"`
}

// SecureExportConfig represents configuration for secure exports
type SecureExportConfig struct {
	// Encryption settings
	EncryptionEnabled  bool            `json:"encryption_enabled"`
	DefaultAlgorithm   EncryptionType `json:"default_algorithm"`
	PGPKeyID           string          `json:"pgp_key_id"`
	AESKeySize         int             `json:"aes_key_size"` // 256, 512
	
	// Access control
	MaxDownloadCount   int             `json:"max_download_count"`
	DefaultExpiryHours int             `json:"default_expiry_hours"`
	RequireAuth        bool            `json:"require_auth"`
	
	// Audit requirements
	AuditAllExports    bool            `json:"audit_all_exports"`
	AuditRetentionDays int             `json:"audit_retention_days"`
	
	// Watermark settings
	WatermarkEnabled   bool            `json:"watermark_enabled"`
	WatermarkText      string          `json:"watermark_text"`
	WatermarkFontSize  int             `json:"watermark_font_size"`
	WatermarkOpacity   int             `json:"watermark_opacity"`
	
	// Delivery settings
	AllowEmailDelivery bool            `json:"allow_email_delivery"`
	AllowSFTPDelivery  bool            `json:"allow_sftp_delivery"`
	AllowAPIDelivery   bool            `json:"allow_api_delivery"`
	
	// Rate limiting
	RateLimitPerHour   int             `json:"rate_limit_per_hour"`
	RateLimitPerDay    int             `json:"rate_limit_per_day"`
	
	// File settings
	MaxFileSizeMB      int             `json:"max_file_size_mb"`
	AllowedFormats     []ReportFormat  `json:"allowed_formats"`
}

// ExportFilter represents filter criteria for querying export logs
type ExportFilter struct {
	UserIDs          []string        `json:"user_ids"`
	RegulatorIDs     []uuid.UUID     `json:"regulator_ids"`
	ReportIDs        []uuid.UUID     `json:"report_ids"`
	ExportTypes      []ExportType    `json:"export_types"`
	Statuses         []ExportStatus  `json:"statuses"`
	
	// Date filters
	RequestedAfter   *time.Time      `json:"requested_after"`
	RequestedBefore  *time.Time      `json:"requested_before"`
	DownloadedAfter  *time.Time      `json:"downloaded_after"`
	DownloadedBefore *time.Time      `json:"downloaded_before"`
	
	// Size filters
	MinFileSize      int64           `json:"min_file_size"`
	MaxFileSize      int64           `json:"max_file_size"`
	
	// Security filters
	ContainsSensitive *bool          `json:"contains_sensitive"`
	ContainsPII       *bool          `json:"contains_pii"`
	
	// Search
	SearchQuery      string          `json:"search_query"`
	
	// Pagination
	Page             int             `json:"page"`
	PageSize         int             `json:"page_size"`
	
	// Sorting
	SortBy           string          `json:"sort_by"`
	SortOrder        string          `json:"sort_order"`
}
