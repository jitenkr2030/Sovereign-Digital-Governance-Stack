package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"csic-platform/service/reporting/internal/domain"
)

// ReportTemplateRepository defines the interface for report template data access
type ReportTemplateRepository interface {
	Create(ctx context.Context, template *domain.ReportTemplate) error
	Update(ctx context.Context, template *domain.ReportTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportTemplate, error)
	GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID) ([]*domain.ReportTemplate, error)
	GetActiveTemplates(ctx context.Context) ([]*domain.ReportTemplate, error)
	GetByReportType(ctx context.Context, reportType domain.ReportType) ([]*domain.ReportTemplate, error)
	List(ctx context.Context, page, pageSize int) ([]*domain.ReportTemplate, int, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]*domain.ReportTemplate, int, error)
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	IncrementVersion(ctx context.Context, id uuid.UUID) error
}

// GeneratedReportRepository defines the interface for generated report data access
type GeneratedReportRepository interface {
	Create(ctx context.Context, report *domain.GeneratedReport) error
	Update(ctx context.Context, report *domain.GeneratedReport) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GeneratedReport, error)
	GetByTemplateID(ctx context.Context, templateID uuid.UUID, page, pageSize int) ([]*domain.GeneratedReport, int, error)
	GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID, filter domain.ReportFilter) ([]*domain.GeneratedReport, int, error)
	GetByStatus(ctx context.Context, status domain.ReportStatus, limit int) ([]*domain.GeneratedReport, error)
	GetPendingReports(ctx context.Context, limit int) ([]*domain.GeneratedReport, error)
	GetOverdueReports(ctx context.Context, deadline time.Time) ([]*domain.GeneratedReport, error)
	GetByDateRange(ctx context.Context, start, end time.Time, filter domain.ReportFilter) ([]*domain.GeneratedReport, int, error)
	List(ctx context.Context, filter domain.ReportFilter) ([]*domain.GeneratedReport, int, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]*domain.GeneratedReport, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ReportStatus, reason string) error
	UpdateSubmissionStatus(ctx context.Context, id uuid.UUID, status string, ref string) error
	Archive(ctx context.Context, id uuid.UUID) error
	Validate(ctx context.Context, id uuid.UUID, validatedBy string) error
	Approve(ctx context.Context, id uuid.UUID, approvedBy string) error
	RecordGeneration(ctx context.Context, id uuid.UUID, processingTime int64) error
	RecordFailure(ctx context.Context, id uuid.UUID, reason string) error
}

// ReportScheduleRepository defines the interface for report schedule data access
type ReportScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.ReportSchedule) error
	Update(ctx context.Context, schedule *domain.ReportSchedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportSchedule, error)
	GetByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*domain.ReportSchedule, error)
	GetActiveSchedules(ctx context.Context) ([]*domain.ReportSchedule, error)
	GetDueSchedules(ctx context.Context, now time.Time) ([]*domain.ReportSchedule, error)
	GetNextRunSchedules(ctx context.Context, now time.Time, limit int) ([]*domain.ReportSchedule, error)
	List(ctx context.Context, page, pageSize int) ([]*domain.ReportSchedule, int, error)
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	UpdateLastRun(ctx context.Context, id uuid.UUID, lastRun time.Time, nextRun time.Time) error
	IncrementRunCount(ctx context.Context, id uuid.UUID, success bool) error
}

// ReportHistoryRepository defines the interface for report history data access
type ReportHistoryRepository interface {
	Create(ctx context.Context, history *domain.ReportHistory) error
	GetByReportID(ctx context.Context, reportID uuid.UUID) ([]*domain.ReportHistory, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]*domain.ReportHistory, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*domain.ReportHistory, error)
	List(ctx context.Context, page, pageSize int) ([]*domain.ReportHistory, int, error)
}

// ReportQueueRepository defines the interface for report queue data access
type ReportQueueRepository interface {
	Create(ctx context.Context, item *domain.ReportQueueItem) error
	Update(ctx context.Context, item *domain.ReportQueueItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportQueueItem, error)
	GetByReportID(ctx context.Context, reportID uuid.UUID) (*domain.ReportQueueItem, error)
	GetPendingItems(ctx context.Context, limit int) ([]*domain.ReportQueueItem, error)
	GetProcessingItems(ctx context.Context, workerID string) ([]*domain.ReportQueueItem, error)
	ClaimItem(ctx context.Context, id uuid.UUID, workerID string) error
	CompleteItem(ctx context.Context, id uuid.UUID) error
	FailItem(ctx context.Context, id uuid.UUID, error string) error
	RequeueFailedItems(ctx context.Context, maxAttempts int) error
	GetQueueStats(ctx context.Context) (pending, processing, failed int, err error)
}

// RegulatorRepository defines the interface for regulator data access
type RegulatorRepository interface {
	Create(ctx context.Context, regulator *domain.Regulator) error
	Update(ctx context.Context, regulator *domain.Regulator) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Regulator, error)
	GetByName(ctx context.Context, name string) (*domain.Regulator, error)
	GetByType(ctx context.Context, regulatorType domain.RegulatorType) ([]*domain.Regulator, error)
	GetByJurisdiction(ctx context.Context, jurisdiction string) ([]*domain.Regulator, error)
	GetActiveRegulators(ctx context.Context) ([]*domain.Regulator, error)
	GetVerifiedRegulators(ctx context.Context) ([]*domain.Regulator, error)
	List(ctx context.Context, page, pageSize int) ([]*domain.Regulator, int, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]*domain.Regulator, int, error)
	Verify(ctx context.Context, id uuid.UUID, verifiedBy string) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	UpdateViewConfig(ctx context.Context, id uuid.UUID, config domain.RegulatorViewConfig) error
}

// RegulatorUserRepository defines the interface for regulator user data access
type RegulatorUserRepository interface {
	Create(ctx context.Context, user *domain.RegulatorUser) error
	Update(ctx context.Context, user *domain.RegulatorUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RegulatorUser, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.RegulatorUser, error)
	GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID) ([]*domain.RegulatorUser, error)
	GetByEmail(ctx context.Context, email string) (*domain.RegulatorUser, error)
	GetActiveByRegulatorID(ctx context.Context, regulatorID uuid.UUID) ([]*domain.RegulatorUser, error)
	List(ctx context.Context, regulatorID uuid.UUID, page, pageSize int) ([]*domain.RegulatorUser, int, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	RecordLogin(ctx context.Context, id uuid.UUID) error
}

// RegulatorAPIKeyRepository defines the interface for regulator API key data access
type RegulatorAPIKeyRepository interface {
	Create(ctx context.Context, apiKey *domain.RegulatorAPIKey) error
	Update(ctx context.Context, apiKey *domain.RegulatorAPIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RegulatorAPIKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*domain.RegulatorAPIKey, error)
	GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID) ([]*domain.RegulatorAPIKey, error)
	GetActiveByRegulatorID(ctx context.Context, regulatorID uuid.UUID) ([]*domain.RegulatorAPIKey, error)
	List(ctx context.Context, regulatorID uuid.UUID, page, pageSize int) ([]*domain.RegulatorAPIKey, int, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedBy string) error
	RecordUsage(ctx context.Context, id uuid.UUID) error
	IsValid(ctx context.Context, id uuid.UUID) (bool, error)
}

// ExportLogRepository defines the interface for export log data access
type ExportLogRepository interface {
	Create(ctx context.Context, log *domain.ExportLog) error
	Update(ctx context.Context, log *domain.ExportLog) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ExportLog, error)
	GetByReportID(ctx context.Context, reportID uuid.UUID) ([]*domain.ExportLog, error)
	GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]*domain.ExportLog, int, error)
	GetByRegulatorID(ctx context.Context, regulatorID uuid.UUID, filter domain.ExportFilter) ([]*domain.ExportLog, int, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*domain.ExportLog, error)
	GetPendingExports(ctx context.Context, limit int) ([]*domain.ExportLog, error)
	List(ctx context.Context, filter domain.ExportFilter) ([]*domain.ExportLog, int, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]*domain.ExportLog, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ExportStatus) error
	RecordDownload(ctx context.Context, id uuid.UUID) error
	RecordEncryption(ctx context.Context, id uuid.UUID, keyID string) error
	GetExportStats(ctx context.Context, start, end time.Time) (total, successful, failed int, totalSize int64, err error)
	GetTopExporters(ctx context.Context, limit int) ([]*ExporterStats, error)
}

// ExporterStats represents statistics for an exporter
type ExporterStats struct {
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	ExportCount    int    `json:"export_count"`
	TotalSize      int64  `json:"total_size"`
	LastExportAt   time.Time `json:"last_export_at"`
}
