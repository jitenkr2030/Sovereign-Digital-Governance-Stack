package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/csic-platform/services/reporting/internal/core/domain"
	"github.com/csic-platform/services/reporting/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReportServiceImpl implements the ReportService interface
type ReportServiceImpl struct {
	repo          ports.ReportRepository
	generator     ports.ReportGenerator
	formatter     ports.ReportFormatter
	storage       ports.ReportStorage
	statsRepo     ports.ReportStatisticsRepository
	metrics       ports.MetricsClient
	logger        *zap.Logger
}

// NewReportService creates a new ReportServiceImpl
func NewReportService(
	repo ports.ReportRepository,
	generator ports.ReportGenerator,
	formatter ports.ReportFormatter,
	storage ports.ReportStorage,
	statsRepo ports.ReportStatisticsRepository,
	metrics ports.MetricsClient,
	logger *zap.Logger,
) *ReportServiceImpl {
	return &ReportServiceImpl{
		repo:      repo,
		generator: generator,
		formatter: formatter,
		storage:   storage,
		statsRepo: statsRepo,
		metrics:   metrics,
		logger:    logger,
	}
}

// GenerateReport generates a new report
func (s *ReportServiceImpl) GenerateReport(
	ctx context.Context,
	req *ports.GenerateReportRequest,
) (*ports.GenerateReportResponse, error) {
	s.logger.Info("Generating report",
		zap.String("type", string(req.Type)),
		zap.String("format", string(req.Format)),
		zap.String("name", req.Name))

	// Parse date range
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	if startDate.After(endDate) {
		return nil, domain.ErrInvalidDateRange
	}

	// Create report record
	report := &domain.Report{
		ID:          uuid.New(),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Status:      domain.ReportStatusPending,
		Format:      req.Format,
		StartDate:   startDate,
		EndDate:     endDate,
		EntityID:    req.EntityID,
		EntityType:  req.EntityType,
		Filters:     req.Filters,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save initial report
	if err := s.repo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Generate report content asynchronously
	go s.generateReportAsync(context.Background(), report, req)

	return &ports.GenerateReportResponse{
		ReportID: report.ID.String(),
		Status:   string(domain.ReportStatusProcessing),
		Message:  "Report generation started",
	}, nil
}

// generateReportAsync generates the report content asynchronously
func (s *ReportServiceImpl) generateReportAsync(ctx context.Context, report *domain.Report, req *ports.GenerateReportRequest) {
	start := time.Now()
	s.logger.Info("Starting async report generation",
		zap.String("report_id", report.ID.String()))

	// Update status to processing
	s.repo.UpdateStatus(ctx, report.ID.String(), domain.ReportStatusProcessing)

	var reportData interface{}
	var err error

	// Generate data based on report type
	switch report.Type {
	case domain.ReportTypeCompliance:
		complianceReq := &ports.ComplianceReportRequest{
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			EntityID:   req.EntityID,
			EntityType: req.EntityType,
		}
		reportData, err = s.generator.GenerateComplianceReport(ctx, complianceReq)
	case domain.ReportTypeRisk:
		riskReq := &ports.RiskReportRequest{
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			EntityID:   req.EntityID,
			EntityType: req.EntityType,
		}
		reportData, err = s.generator.GenerateRiskReport(ctx, riskReq)
	case domain.ReportTypeAudit:
		auditReq := &ports.AuditReportRequest{
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			EntityID:   req.EntityID,
			EntityType: req.EntityType,
		}
		reportData, err = s.generator.GenerateAuditReport(ctx, auditReq)
	case domain.ReportTypeExposure:
		exposureReq := &ports.ExposureReportRequest{
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			EntityID:   req.EntityID,
			EntityType: req.EntityType,
		}
		reportData, err = s.generator.GenerateExposureReport(ctx, exposureReq)
	default:
		err = fmt.Errorf("unsupported report type: %s", report.Type)
	}

	if err != nil {
		s.logger.Error("Failed to generate report data",
			zap.String("report_id", report.ID.String()),
			zap.Error(err))
		s.repo.UpdateStatus(ctx, report.ID.String(), domain.ReportStatusFailed)
		report.ErrorMessage = err.Error()
		s.repo.Update(ctx, report)
		return
	}

	// Format report to the requested format
	var content []byte
	switch report.Format {
	case domain.OutputFormatPDF:
		content, err = s.formatter.FormatToPDF(reportData)
	case domain.OutputFormatCSV:
		content, err = s.formatter.FormatToCSV(reportData)
	case domain.OutputFormatJSON:
		content, err = s.formatter.FormatToJSON(reportData)
	case domain.OutputFormatHTML:
		content, err = s.formatter.FormatToHTML(reportData)
	default:
		err = domain.ErrFormatUnsupported
	}

	if err != nil {
		s.logger.Error("Failed to format report",
			zap.String("report_id", report.ID.String()),
			zap.Error(err))
		s.repo.UpdateStatus(ctx, report.ID.String(), domain.ReportStatusFailed)
		report.ErrorMessage = err.Error()
		s.repo.Update(ctx, report)
		return
	}

	// Store the report file
	if err := s.storage.Store(ctx, report.ID.String(), content, report.Format); err != nil {
		s.logger.Error("Failed to store report",
			zap.String("report_id", report.ID.String()),
			zap.Error(err))
		s.repo.UpdateStatus(ctx, report.ID.String(), domain.ReportStatusFailed)
		report.ErrorMessage = "Failed to store report file"
		s.repo.Update(ctx, report)
		return
	}

	// Update report with file info
	now := time.Now()
	report.Status = domain.ReportStatusCompleted
	report.FilePath = s.storage.GetPath(report.ID.String(), report.Format)
	report.FileSize = int64(len(content))
	report.CompletedAt = &now
	s.repo.Update(ctx, report)

	// Record metrics
	if s.metrics != nil {
		s.metrics.IncrementReportGenerated(report.Type, report.Format)
		s.metrics.RecordGenerationTime(string(report.Type), time.Since(start).String())
	}

	s.logger.Info("Report generation completed",
		zap.String("report_id", report.ID.String()),
		zap.Duration("duration", time.Since(start)))
}

// GetReport retrieves a report by ID
func (s *ReportServiceImpl) GetReport(ctx context.Context, id string) (*domain.Report, error) {
	report, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}
	return report, nil
}

// ListReports lists reports with filtering
func (s *ReportServiceImpl) ListReports(ctx context.Context, req *ports.ListReportsRequest) (*ports.ListReportsResponse, error) {
	return s.repo.FindAll(ctx, req)
}

// DownloadReport downloads a report file
func (s *ReportServiceImpl) DownloadReport(ctx context.Context, id string) (io.ReadCloser, error) {
	report, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}

	if report.Status != domain.ReportStatusCompleted {
		return nil, domain.ErrReportNotReady
	}

	reader, err := s.storage.Retrieve(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve report: %w", err)
	}

	// Increment download count
	s.repo.IncrementDownload(ctx, id)

	if s.metrics != nil {
		s.metrics.IncrementReportDownloaded(report.Type)
	}

	return reader, nil
}

// DeleteReport deletes a report
func (s *ReportServiceImpl) DeleteReport(ctx context.Context, id string) error {
	// Delete from storage
	if err := s.storage.Delete(ctx, id); err != nil {
		s.logger.Warn("Failed to delete report file", zap.Error(err))
	}

	// Delete from repository
	return s.repo.Delete(ctx, id)
}

// ArchiveReport archives a report
func (s *ReportServiceImpl) ArchiveReport(ctx context.Context, id string) error {
	return s.repo.Archive(ctx, id)
}

// GetStatistics returns report statistics
func (s *ReportServiceImpl) GetStatistics(ctx context.Context) (*domain.ReportStatistics, error) {
	return s.statsRepo.GetStatistics(ctx)
}

// Ensure ReportServiceImpl implements ReportService
var _ ports.ReportService = (*ReportServiceImpl)(nil)
