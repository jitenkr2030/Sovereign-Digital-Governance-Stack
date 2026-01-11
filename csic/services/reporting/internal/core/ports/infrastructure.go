package ports

import (
	"context"

	"github.com/csic-platform/services/reporting/internal/core/domain"
)

// Scheduler defines the interface for job scheduling
type Scheduler interface {
	// Schedule schedules a job
	Schedule(ctx context.Context, job *ScheduledJob) error

	// Cancel cancels a scheduled job
	Cancel(ctx context.Context, id string) error

	// GetNextExecution returns the next execution time
	GetNextExecution(ctx context.Context, id string) (string, error)

	// ListJobs lists all scheduled jobs
	ListJobs(ctx context.Context) ([]*ScheduledJob, error)
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Schedule    string            `json:"schedule"`
	Handler     string            `json:"handler"`
	Payload     map[string]string `json:"payload"`
	Enabled     bool              `json:"enabled"`
	NextRunAt   string            `json:"next_run_at,omitempty"`
	LastRunAt   string            `json:"last_run_at,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// NotificationClient defines the interface for sending notifications
type NotificationClient interface {
	// SendEmail sends an email notification
	SendEmail(ctx context.Context, req *EmailNotificationRequest) error

	// SendWebhook sends a webhook notification
	SendWebhook(ctx context.Context, req *WebhookNotificationRequest) error

	// Close closes the notification client
	Close() error
}

// EmailNotificationRequest represents an email notification
type EmailNotificationRequest struct {
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	ReportID    string   `json:"report_id,omitempty"`
	ReportName  string   `json:"report_name,omitempty"`
	Attachment  []byte   `json:"attachment,omitempty"`
}

// WebhookNotificationRequest represents a webhook notification
type WebhookNotificationRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Payload map[string]string `json:"payload"`
}

// DataProvider defines the interface for fetching report data
type DataProvider interface {
	// FetchComplianceData fetches compliance data
	FetchComplianceData(ctx context.Context, req *ComplianceReportRequest) (*domain.ComplianceReport, error)

	// FetchRiskData fetches risk data
	FetchRiskData(ctx context.Context, req *RiskReportRequest) (*domain.RiskReport, error)

	// FetchAuditData fetches audit data
	FetchAuditData(ctx context.Context, req *AuditReportRequest) (*AuditReportData, error)

	// FetchExposureData fetches exposure data
	FetchExposureData(ctx context.Context, req *ExposureReportRequest) (*ExposureReportData, error)
}

// StorageClient defines the interface for object storage
type StorageClient interface {
	// Upload uploads a file to storage
	Upload(ctx context.Context, key string, content []byte, contentType string) error

	// Download downloads a file from storage
	Download(ctx context.Context, key string) ([]byte, error)

	// Delete deletes a file from storage
	Delete(ctx context.Context, key string) error

	// GetURL returns the URL for a file
	GetURL(ctx context.Context, key string) (string, error)

	// List lists files in a prefix
	List(ctx context.Context, prefix string) ([]string, error)
}

// MetricsClient defines the interface for metrics collection
type MetricsClient interface {
	// IncrementReportGenerated increments the report generated counter
	IncrementReportGenerated(reportType domain.ReportType, format domain.OutputFormat)

	// IncrementReportDownloaded increments the report downloaded counter
	IncrementReportDownloaded(reportType domain.ReportType)

	// RecordGenerationTime records the report generation time
	RecordGenerationTime(reportType string, duration string)

	// SetStorageUsed sets the storage used gauge
	SetStorageUsed(bytes int64)

	// Close closes the metrics client
	Close() error
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	// Check checks the health of the service
	Check(ctx context.Context) (*HealthStatus, error)
}

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status      string            `json:"status"`
	Components  map[string]string `json:"components"`
	Timestamp   string            `json:"timestamp"`
}
