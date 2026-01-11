package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"csic-platform/service/reporting/internal/domain"
	"csic-platform/service/reporting/internal/repository"
)

// ExportService handles secure export operations
type ExportService struct {
	reportRepo    repository.GeneratedReportRepository
	exportLogRepo repository.ExportLogRepository
	cacheRepo     *repository.CacheRepository
	storage       FileStorage
	config        SecureExportConfig
}

// SecureExportConfig represents configuration for secure exports
type SecureExportConfig struct {
	EncryptionEnabled   bool
	DefaultAlgorithm    string
	MaxDownloadCount    int
	DefaultExpiryHours  int
	RequireAuth         bool
	AuditAllExports     bool
	WatermarkEnabled    bool
	WatermarkText       string
	AllowEmailDelivery  bool
	AllowSFTPDelivery   bool
	MaxFileSizeMB       int
}

// NewExportService creates a new export service
func NewExportService(
	reportRepo repository.GeneratedReportRepository,
	exportLogRepo repository.ExportLogRepository,
	cacheRepo *repository.CacheRepository,
	storage FileStorage,
	config SecureExportConfig,
) *ExportService {
	return &ExportService{
		reportRepo:    reportRepo,
		exportLogRepo: exportLogRepo,
		cacheRepo:     cacheRepo,
		storage:       storage,
		config:        config,
	}
}

// ExportRequest represents a request to export a report
type ExportRequest struct {
	ReportID        uuid.UUID
	UserID          string
	UserName        string
	UserEmail       string
	UserRole        string
	UserOrganization string
	Format          domain.ReportFormat
	Encryption      domain.EncryptionType
	Password        string // For password-protected exports
	Watermark       bool
	WatermarkText   string
	IPAddress       string
	UserAgent       string
	SessionID       string
	DeliveryMethod  string
	DeliveryAddress string
}

// ExportResult represents the result of an export operation
type ExportResult struct {
	ExportLog     *domain.ExportLog
	FileReader    io.ReadCloser
	FileName      string
	FileSize      int64
	ContentType   string
}

// ExportReport exports a report securely
func (s *ExportService) ExportReport(ctx context.Context, req *ExportRequest) (*ExportResult, error) {
	// Get the report
	report, err := s.reportRepo.GetByID(ctx, req.ReportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	if report == nil {
		return nil, fmt.Errorf("report not found: %s", req.ReportID)
	}

	if report.Status != domain.ReportStatusCompleted {
		return nil, fmt.Errorf("report is not ready for export: status=%s", report.Status)
	}

	// Find the file in the requested format
	var targetFile *domain.ReportFile
	for i := range report.Files {
		if report.Files[i].Format == req.Format {
			targetFile = &report.Files[i]
			break
		}
	}

	if targetFile == nil {
		return nil, fmt.Errorf("report file not found in format: %s", req.Format)
	}

	// Create export log entry
	exportLog := s.createExportLog(ctx, report, req, targetFile)

	// Save the export log
	if err := s.exportLogRepo.Create(ctx, exportLog); err != nil {
		return nil, fmt.Errorf("failed to create export log: %w", err)
	}

	// Get the file from storage
	fileReader, err := s.storage.GetReport(ctx, targetFile.FilePath)
	if err != nil {
		s.updateExportStatus(ctx, exportLog.ID, domain.ExportStatusFailed, err.Error())
		return nil, fmt.Errorf("failed to get file from storage: %w", err)
	}

	// Apply encryption if requested
	if req.Encryption == domain.EncryptionAES256 {
		fileReader, err = s.encryptReader(ctx, fileReader, req.Password)
		if err != nil {
			fileReader.Close()
			s.updateExportStatus(ctx, exportLog.ID, domain.ExportStatusFailed, err.Error())
			return nil, fmt.Errorf("failed to encrypt file: %w", err)
		}

		// Update export log with encryption info
		s.exportLogRepo.RecordEncryption(ctx, exportLog.ID, "aes256")
	}

	// Apply watermark if requested
	if req.Watermark {
		fileReader, err = s.applyWatermark(ctx, fileReader, req.WatermarkText)
		if err != nil {
			fileReader.Close()
			s.updateExportStatus(ctx, exportLog.ID, domain.ExportStatusFailed, err.Error())
			return nil, fmt.Errorf("failed to apply watermark: %w", err)
		}
	}

	// Update export log status
	s.updateExportStatus(ctx, exportLog.ID, domain.ExportStatusCompleted, "")

	return &ExportResult{
		ExportLog:   exportLog,
		FileReader:  fileReader,
		FileName:    targetFile.Filename,
		FileSize:    targetFile.FileSize,
		ContentType: s.getContentType(req.Format),
	}, nil
}

// createExportLog creates an export log entry
func (s *ExportService) createExportLog(ctx context.Context, report *domain.GeneratedReport, req *ExportRequest, file *domain.ReportFile) *domain.ExportLog {
	// Determine expiration
	var expiresAt *time.Time
	if s.config.DefaultExpiryHours > 0 {
		t := time.Now().Add(time.Duration(s.config.DefaultExpiryHours) * time.Hour)
		expiresAt = &t
	}

	return &domain.ExportLog{
		ID:               uuid.New(),
		ReportID:         &report.ID,
		ExportType:       domain.ExportTypeReport,
		UserID:           req.UserID,
		UserName:         req.UserName,
		UserEmail:        req.UserEmail,
		UserRole:         req.UserRole,
		UserOrganization: req.UserOrganization,
		ExportFormat:     req.Format,
		FileName:         file.Filename,
		FilePath:         file.FilePath,
		FileSize:         file.FileSize,
		Checksum:         file.Checksum,
		ChecksumType:     file.ChecksumType,
		EncryptionType:   req.Encryption,
		RequestedAt:      time.Now(),
		ExpiresAt:        expiresAt,
		IPAddress:        req.IPAddress,
		UserAgent:        req.UserAgent,
		SessionID:        req.SessionID,
		Status:           domain.ExportStatusProcessing,
		Progress:         100,
		DataScope: domain.ExportDataScope{
			DateRangeStart: &report.PeriodStart,
			DateRangeEnd:   &report.PeriodEnd,
		},
		ContainsSensitive: false,
		ContainsPII:       false,
		CreatedAt:         time.Now(),
	}
}

// encryptReader encrypts the content of a reader
func (s *ExportService) encryptReader(ctx context.Context, reader io.Reader, password string) (io.ReadCloser, error) {
	// Derive key from password
	key := sha256.Sum256([]byte(password))

	// Create cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Encrypt content
	encrypted := gcm.Seal(nonce, nonce, content, nil)

	return io.NopCloser(io.Reader(bytes.NewReader(encrypted))), nil
}

// applyWatermark applies a watermark to the content
func (s *ExportService) applyWatermark(ctx context.Context, reader io.Reader, watermarkText string) (io.ReadCloser, error) {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Detect the file type (PDF, image, etc.)
	2. Apply the watermark to the content
	// 3. Return the watermarked content

	// For now, just return the original reader
	return reader, nil
}

// updateExportStatus updates the status of an export
func (s *ExportService) updateExportStatus(ctx context.Context, exportLogID uuid.UUID, status domain.ExportStatus, reason string) {
	err := s.exportLogRepo.UpdateStatus(ctx, exportLogID, status)
	if err != nil {
		log.Printf("Failed to update export status: %v", err)
	}
}

// getContentType returns the content type for a format
func (s *ExportService) getContentType(format domain.ReportFormat) string {
	contentTypes := map[domain.ReportFormat]string{
		domain.ReportFormatPDF:  "application/pdf",
		domain.ReportFormatCSV:  "text/csv",
		domain.ReportFormatXLSX: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		domain.ReportFormatJSON: "application/json",
		domain.ReportFormatXBRL: "application/xbrl+xml",
		domain.ReportFormatXML:  "application/xml",
	}

	if ct, ok := contentTypes[format]; ok {
		return ct
	}
	return "application/octet-stream"
}

// GetExportLogs retrieves export logs with filters
func (s *ExportService) GetExportLogs(ctx context.Context, filter domain.ExportFilter) ([]*domain.ExportLog, int, error) {
	return s.exportLogRepo.List(ctx, filter)
}

// GetExportLogsByUser retrieves export logs for a specific user
func (s *ExportService) GetExportLogsByUser(ctx context.Context, userID string, page, pageSize int) ([]*domain.ExportLog, int, error) {
	return s.exportLogRepo.GetByUserID(ctx, userID, page, pageSize)
}

// RecordDownload records a download event
func (s *ExportService) RecordDownload(ctx context.Context, exportLogID uuid.UUID) error {
	return s.exportLogRepo.RecordDownload(ctx, exportLogID)
}

// GetExportStats retrieves export statistics
func (s *ExportService) GetExportStats(ctx context.Context, start, end time.Time) (ExportStats, error) {
	total, successful, failed, totalSize, err := s.exportLogRepo.GetExportStats(ctx, start, end)
	if err != nil {
		return ExportStats{}, err
	}

	return ExportStats{
		TotalExports:  total,
		Successful:    successful,
		Failed:        failed,
		TotalSize:     totalSize,
		SuccessRate:   float64(successful) / float64(total) * 100 if total > 0 else 0,
	}, nil
}

// ExportStats represents export statistics
type ExportStats struct {
	TotalExports  int
	Successful    int
	Failed        int
	TotalSize     int64
	SuccessRate   float64
}

// VerifyChecksum verifies the checksum of a downloaded file
func (s *ExportService) VerifyChecksum(ctx context.Context, exportLogID uuid.UUID, fileContent []byte) (bool, error) {
	// Get the export log
	exportLog, err := s.exportLogRepo.GetByID(ctx, exportLogID)
	if err != nil {
		return false, err
	}

	if exportLog == nil {
		return false, fmt.Errorf("export log not found: %s", exportLogID)
	}

	// Calculate checksum
	var calculatedChecksum string
	switch exportLog.ChecksumType {
	case "sha256":
		hash := sha256.Sum256(fileContent)
		calculatedChecksum = hex.EncodeToString(hash[:])
	case "md5":
		// Implement MD5 if needed
		calculatedChecksum = ""
	default:
		calculatedChecksum = ""
	}

	return calculatedChecksum == exportLog.Checksum, nil
}

// GetTopExporters retrieves the top exporters
func (s *ExportService) GetTopExporters(ctx context.Context, limit int) ([]*repository.ExporterStats, error) {
	return s.exportLogRepo.GetTopExporters(ctx, limit)
}

// CancelExport cancels an export
func (s *ExportService) CancelExport(ctx context.Context, exportLogID uuid.UUID) error {
	return s.exportLogRepo.UpdateStatus(ctx, exportLogID, domain.ExportStatusCancelled)
}

// CleanupExpiredExports removes expired exports
func (s *ExportService) CleanupExpiredExports(ctx context.Context) (int, error) {
	// Get all pending exports
	exports, err := s.exportLogRepo.GetPendingExports(ctx, 1000)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	cleaned := 0

	for _, export := range exports {
		if export.ExpiresAt != nil && export.ExpiresAt.Before(now) {
			// Cancel the export
			if err := s.CancelExport(ctx, export.ID); err != nil {
				log.Printf("Failed to cancel expired export %s: %v", export.ID, err)
				continue
			}
			cleaned++
		}
	}

	return cleaned, nil
}

// Import necessary packages
import (
	"bytes"
	"log"
)
