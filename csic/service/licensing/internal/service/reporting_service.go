package service

import (
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/csic-licensing/internal/domain/models"
	"github.com/csic-licensing/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ReportingService handles report generation and management
type ReportingService struct {
	reportRepo     repository.ReportRepository
	licenseRepo    repository.LicenseRepository
	violationRepo  repository.ViolationRepository
	complianceRepo repository.ComplianceRepository
	templateStore  TemplateStore
	fileStorage    FileStorage
	logger         *zap.Logger
}

// TemplateStore defines the interface for report templates
type TemplateStore interface {
	GetTemplate(templateID string) (*models.ReportTemplate, error)
	GetTemplatesByType(reportType models.ReportType) ([]*models.ReportTemplate, error)
	SaveTemplate(template *models.ReportTemplate) error
}

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	UploadFile(ctx context.Context, fileName string, content []byte, mimeType string) (string, error)
	DownloadFile(ctx context.Context, filePath string) ([]byte, error)
	GetDownloadURL(ctx context.Context, filePath string) (string, error)
	DeleteFile(ctx context.Context, filePath string) error
}

// NewReportingService creates a new reporting service
func NewReportingService(
	reportRepo repository.ReportRepository,
	licenseRepo repository.LicenseRepository,
	violationRepo repository.ViolationRepository,
	complianceRepo repository.ComplianceRepository,
	templateStore TemplateStore,
	fileStorage FileStorage,
	logger *zap.Logger,
) *ReportingService {
	return &ReportingService{
		reportRepo:     reportRepo,
		licenseRepo:    licenseRepo,
		violationRepo:  violationRepo,
		complianceRepo: complianceRepo,
		templateStore:  templateStore,
		fileStorage:    fileStorage,
		logger:         logger,
	}
}

// GenerateReport generates a new report
func (s *ReportingService) GenerateReport(ctx context.Context, reportType models.ReportType, periodStart, periodEnd time.Time, licenseID, generatedBy string, parameters map[string]interface{}) (*models.Report, error) {
	// Get license
	license, err := s.licenseRepo.GetByID(ctx, licenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get license: %w", err)
	}
	if license == nil {
		return nil, fmt.Errorf("license not found")
	}

	// Create report record
	report := models.NewReport(reportType, periodStart, periodEnd, licenseID, license.ExchangeID)
	report.GeneratedBy = generatedBy
	report.Status = models.ReportStatusGenerating
	report.Parameters = parameters

	// Set title based on report type
	report.Title = s.getReportTitle(reportType, periodStart, periodEnd)

	// Save initial report
	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Generate report content asynchronously (in production, use worker queue)
	go s.generateReportContent(report.ID)

	s.logger.Info("Report generation started",
		zap.String("report_id", report.ID),
		zap.String("report_type", string(reportType)),
		zap.String("license_id", licenseID))

	return report, nil
}

// generateReportContent generates the actual report content
func (s *ReportingService) generateReportContent(reportID string) {
	ctx := context.Background()

	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		s.logger.Error("Failed to get report for generation", zap.String("report_id", reportID), zap.Error(err))
		return
	}

	// Collect report data
	reportData := s.collectReportData(ctx, report)

	// Generate content using template
	content, err := s.renderReport(report, reportData)
	if err != nil {
		s.logger.Error("Failed to render report", zap.String("report_id", reportID), zap.Error(err))
		s.updateReportStatus(ctx, reportID, models.ReportStatusPending, err.Error())
		return
	}

	// Upload file
	filePath, err := s.fileStorage.UploadFile(ctx, fmt.Sprintf("reports/%s/%s.pdf", report.ID, report.ReportType), content, "application/pdf")
	if err != nil {
		s.logger.Error("Failed to upload report", zap.String("report_id", reportID), zap.Error(err))
		s.updateReportStatus(ctx, reportID, models.ReportStatusPending, err.Error())
		return
	}

	// Update report with file info
	report.FilePath = filePath
	report.FileSize = int64(len(content))
	report.MimeType = "application/pdf"
	report.Status = models.ReportStatusReady

	if err := s.reportRepo.Update(ctx, report); err != nil {
		s.logger.Error("Failed to update report", zap.String("report_id", reportID), zap.Error(err))
	}

	s.logger.Info("Report generated successfully", zap.String("report_id", reportID))
}

// collectReportData collects data for the report
func (s *ReportingService) collectReportData(ctx context.Context, report *models.Report) map[string]interface{} {
	data := make(map[string]interface{})

	// Basic info
	data["report_id"] = report.ID
	data["report_type"] = report.ReportType
	data["period_start"] = report.PeriodStart
	data["period_end"] = report.PeriodEnd
	data["generated_at"] = time.Now()

	// License info
	license, _ := s.licenseRepo.GetByID(ctx, report.LicenseID)
	if license != nil {
		data["license"] = license
	}

	// Violation summary
	violationStats, _ := s.violationRepo.GetStats(ctx)
	data["violation_stats"] = violationStats

	// Compliance score
	complianceScore, _ := s.complianceRepo.CalculateComplianceScore(ctx, report.ExchangeID)
	data["compliance_score"] = complianceScore

	// Open violations
	openViolations, _ := s.violationRepo.GetOpenViolations(ctx)
	data["open_violations"] = openViolations

	return data
}

// renderReport renders the report using a template
func (s *ReportingService) renderReport(report *models.Report, data map[string]interface{}) ([]byte, error) {
	// Default template
	templateContent := `
# Compliance Report
## {{.ReportType}}
**Period:** {{.PeriodStart}} to {{.PeriodEnd}}
**Generated:** {{.GeneratedAt}}

## Summary
This report covers the compliance status for the specified period.

## Violation Summary
{{if .ViolationStats}}
- Total Violations: {{.ViolationStats.TotalViolations}}
- Open Violations: {{.ViolationStats.OpenViolations}}
- Critical: {{.ViolationStats.CriticalCount}}
- High: {{.ViolationStats.HighCount}}
{{else}}
No violations recorded.
{{end}}

## Compliance Score
{{if .ComplianceScore}}
Overall Score: {{.ComplianceScore.OverallScore}}
Risk Level: {{.ComplianceScore.RiskLevel}}
{{else}}
Compliance score not available.
{{end}}

## Open Violations
{{if .OpenViolations}}
{{range .OpenViolations}}
- [{{.Severity}}] {{.Title}}: {{.Description}}
{{end}}
{{else}}
No open violations.
{{end}}
`

	tmpl, err := template.New("report").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// In production, would render to PDF using a library like chromedp or unioffice
	// For now, return plain text
	content := fmt.Sprintf("Report: %s\nPeriod: %s to %s\nGenerated: %s\n\n",
		report.Title,
		report.PeriodStart.Format("2006-01-02"),
		report.PeriodEnd.Format("2006-01-02"),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return []byte(content), nil
}

// updateReportStatus updates the report status
func (s *ReportingService) updateReportStatus(ctx context.Context, reportID string, status models.ReportStatus, errorMsg string) {
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return
	}

	report.Status = status
	if errorMsg != "" {
		report.Metadata["error"] = errorMsg
	}

	s.reportRepo.Update(ctx, report)
}

// getReportTitle returns a human-readable title for the report type
func (s *ReportingService) getReportTitle(reportType models.ReportType, start, end time.Time) string {
	period := start.Format("Jan 2006")
	switch reportType {
	case models.ReportTypeDailyLiquidity:
		return fmt.Sprintf("Daily Liquidity Report - %s", start.Format("2006-01-02"))
	case models.ReportTypeMonthlyVolume:
		return fmt.Sprintf("Monthly Volume Report - %s", period)
	case models.ReportTypeQuarterlyFinancial:
		return fmt.Sprintf("Quarterly Financial Report - Q%d %d", int(start.Month()/3)+1, start.Year())
	case models.ReportTypeSuspiciousActivity:
		return fmt.Sprintf("Suspicious Activity Report - %s", start.Format("2006-01-02"))
	case models.ReportTypeRiskAssessment:
		return fmt.Sprintf("Risk Assessment Report - %s", period)
	default:
		return fmt.Sprintf("Compliance Report - %s", period)
	}
}

// GetReport retrieves a report by ID
func (s *ReportingService) GetReport(ctx context.Context, id string) (*models.Report, error) {
	report, err := s.reportRepo.GetReportByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, fmt.Errorf("report not found")
	}
	return report, nil
}

// GetReportsByLicense retrieves all reports for a license
func (s *ReportingService) GetReportsByLicense(ctx context.Context, licenseID string) ([]*models.Report, error) {
	return s.reportRepo.GetByLicense(ctx, licenseID)
}

// GetReportsByExchange retrieves all reports for an exchange
func (s *ReportingService) GetReportsByExchange(ctx context.Context, exchangeID string) ([]*models.Report, error) {
	return s.reportRepo.GetByExchange(ctx, exchangeID)
}

// GetReportsByFilter retrieves reports matching the filter
func (s *ReportingService) GetReportsByFilter(ctx context.Context, filter repository.ReportFilter) ([]*models.Report, int64, error) {
	reports, err := s.reportRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	count, err := s.reportRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	return reports, count, nil
}

// DownloadReport gets the download URL for a report
func (s *ReportingService) DownloadReport(ctx context.Context, reportID string) (string, error) {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return "", err
	}

	if report.Status != models.ReportStatusReady && report.Status != models.ReportStatusSubmitted {
		return "", fmt.Errorf("report is not ready for download")
	}

	return s.fileStorage.GetDownloadURL(ctx, report.FilePath)
}

// SubmitReport marks a report as submitted to regulators
func (s *ReportingService) SubmitReport(ctx context.Context, reportID, submittedTo, submittedBy string) error {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return err
	}

	if !report.IsSubmittable() {
		return fmt.Errorf("report cannot be submitted in current status: %s", report.Status)
	}

	now := time.Now()
	report.Status = models.ReportStatusSubmitted
	report.SubmittedAt = &now
	report.SubmittedTo = submittedTo
	report.ReviewedBy = submittedBy

	if err := s.reportRepo.Update(ctx, report); err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	s.logger.Info("Report submitted",
		zap.String("report_id", reportID),
		zap.String("submitted_to", submittedTo),
		zap.String("submitted_by", submittedBy))

	return nil
}

// RejectReport rejects a report
func (s *ReportingService) RejectReport(ctx context.Context, reportID, reason, rejectedBy string) error {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return err
	}

	report.Status = models.ReportStatusRejected
	report.Metadata["rejection_reason"] = reason
	report.ReviewedBy = rejectedBy

	if err := s.reportRepo.Update(ctx, report); err != nil {
		return fmt.Errorf("failed to reject report: %w", err)
	}

	s.logger.Info("Report rejected",
		zap.String("report_id", reportID),
		zap.String("reason", reason))

	return nil
}

// GetReportStats retrieves report statistics
func (s *ReportingService) GetReportStats(ctx context.Context) (*models.ReportStats, error) {
	return s.reportRepo.GetStats(ctx)
}

// GetOverdueReports retrieves reports past their due date
func (s *ReportingService) GetOverdueReports(ctx context.Context) ([]*models.Report, error) {
	return s.reportRepo.GetOverdueReports(ctx)
}

// GenerateComplianceReport generates a comprehensive compliance report
func (s *ReportingService) GenerateComplianceReport(ctx context.Context, licenseID, generatedBy string) (*models.Report, error) {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	periodEnd := time.Date(now.Year(), now.Month(), 0, 23, 59, 59, 0, now.Location())

	return s.GenerateReport(ctx, models.ReportTypeComplianceCertificate, periodStart, periodEnd, licenseID, generatedBy, nil)
}

// ScheduleReport schedules automated report generation
func (s *ReportingService) ScheduleReport(ctx context.Context, schedule *models.ReportSchedule) error {
	// In production, this would integrate with a cron scheduler
	s.logger.Info("Report scheduled",
		zap.String("template_id", schedule.TemplateID),
		zap.String("license_id", schedule.LicenseID),
		zap.String("schedule", schedule.CronSchedule))

	return nil
}
