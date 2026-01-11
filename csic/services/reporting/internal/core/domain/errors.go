package domain

import "errors"

// Common error definitions for the reporting domain

var (
	// Report errors
	ErrReportNotFound       = errors.New("report not found")
	ErrReportInvalid        = errors.New("invalid report configuration")
	ErrReportGenerationFailed = errors.New("report generation failed")
	ErrReportNotReady       = errors.New("report is not ready")
	ErrReportDownloadFailed = errors.New("report download failed")
	ErrReportAlreadyExists  = errors.New("report already exists")

	// Template errors
	ErrTemplateNotFound     = errors.New("report template not found")
	ErrTemplateInvalid      = errors.New("invalid template configuration")
	ErrTemplateAlreadyExists = errors.New("template already exists")
	ErrTemplateInUse        = errors.New("template is in use by scheduled reports")

	// Scheduled report errors
	ErrScheduledNotFound    = errors.New("scheduled report not found")
	ErrScheduledInvalid     = errors.New("invalid scheduled report configuration")
	ErrScheduledAlreadyExists = errors.New("scheduled report already exists")
	ErrCronInvalid          = errors.New("invalid cron schedule")
	ErrSchedulerNotRunning  = errors.New("scheduler is not running")

	// Generation errors
	ErrGenerationTimeout    = errors.New("report generation timed out")
	ErrGenerationCancelled  = errors.New("report generation was cancelled")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrDataSourceUnavailable = errors.New("data source unavailable")

	// Formatting errors
	ErrFormatUnsupported    = errors.New("unsupported output format")
	ErrFormattingFailed     = errors.New("formatting failed")

	// Storage errors
	ErrStorageFull          = errors.New("storage is full")
	ErrStorageUnavailable   = errors.New("storage unavailable")
	ErrFileNotFound         = errors.New("file not found")
	ErrFileCorrupted        = errors.New("file is corrupted")

	// Repository errors
	ErrRepositoryNotFound   = errors.New("record not found in repository")
	ErrRepositoryConflict   = errors.New("record conflict in repository")
	ErrRepositoryUnavailable = errors.New("repository unavailable")
)
