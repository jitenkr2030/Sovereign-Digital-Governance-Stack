package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"csic-platform/service/reporting/internal/domain"
	"csic-platform/service/reporting/internal/repository"
)

// ReportGenerationService handles report generation operations
type ReportGenerationService struct {
	templateRepo   repository.ReportTemplateRepository
	reportRepo     repository.GeneratedReportRepository
	cacheRepo      *repository.CacheRepository
	kafkaProducer  KafkaProducer
	fileStorage    FileStorage
	metrics        MetricsRecorder
}

// NewReportGenerationService creates a new report generation service
func NewReportGenerationService(
	templateRepo repository.ReportTemplateRepository,
	reportRepo repository.GeneratedReportRepository,
	cacheRepo *repository.CacheRepository,
	kafkaProducer KafkaProducer,
	fileStorage FileStorage,
	metrics MetricsRecorder,
) *ReportGenerationService {
	return &ReportGenerationService{
		templateRepo:  templateRepo,
		reportRepo:    reportRepo,
		cacheRepo:     cacheRepo,
		kafkaProducer: kafkaProducer,
		fileStorage:   fileStorage,
		metrics:       metrics,
	}
}

// MetricsRecorder defines the interface for recording metrics
type MetricsRecorder interface {
	RecordReportGenerated(reportType string, duration time.Duration, success bool)
	RecordReportSize(reportType string, size int64)
}

// KafkaProducer defines the interface for publishing Kafka events
type KafkaProducer interface {
	PublishReportGenerated(ctx context.Context, report *domain.GeneratedReport) error
	PublishReportFailed(ctx context.Context, reportID uuid.UUID, reason string) error
	PublishReportScheduled(ctx context.Context, schedule *domain.ReportSchedule) error
}

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	SaveReport(ctx context.Context, reportID uuid.UUID, format domain.ReportFormat, content io.Reader) (string, int64, error)
	GetReport(ctx context.Context, filePath string) (io.ReadCloser, error)
	DeleteReport(ctx context.Context, filePath string) error
}

// GenerateReportRequest represents a request to generate a report
type GenerateReportRequest struct {
	TemplateID      uuid.UUID
	PeriodStart     time.Time
	PeriodEnd       time.Time
	GeneratedBy     string
	GenerationMethod string // "scheduled", "manual", "triggered"
	TriggeredBy     string
	OutputFormats   []domain.ReportFormat
	CustomFields    map[string]interface{}
}

// GenerateReport generates a new report
func (s *ReportGenerationService) GenerateReport(ctx context.Context, req *GenerateReportRequest) (*domain.GeneratedReport, error) {
	// Get the template
	template, err := s.getTemplateWithCache(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return nil, fmt.Errorf("template not found: %s", req.TemplateID)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template is not active: %s", req.TemplateID)
	}

	// Create the generated report record
	report := &domain.GeneratedReport{
		ID:               uuid.New(),
		TemplateID:       template.ID,
		TemplateName:     template.Name,
		ReportType:       template.ReportType,
		RegulatorID:      template.RegulatorID,
		Status:           domain.ReportStatusPending,
		Version:          template.Version,
		Title:            fmt.Sprintf("%s - %s", template.Name, req.PeriodEnd.Format("2006-01-02")),
		Description:      template.Description,
		PeriodStart:      req.PeriodStart,
		PeriodEnd:        req.PeriodEnd,
		ReportDate:       time.Now(),
		GeneratedBy:      req.GeneratedBy,
		GenerationMethod: req.GenerationMethod,
		TriggeredBy:      req.TriggeredBy,
		OutputFormats:    req.OutputFormats,
		CustomFields:     req.CustomFields,
		RetentionPeriod:  365, // Default 1 year
		CreatedBy:        req.GeneratedBy,
	}

	// Set output formats from template if not specified
	if len(req.OutputFormats) == 0 {
		report.OutputFormats = template.OutputFormats
	}

	// Save the report record
	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create report record: %w", err)
	}

	// Start generation asynchronously
	go s.processReportGeneration(context.Background(), report, template)

	return report, nil
}

// processReportGeneration processes the report generation
func (s *ReportGenerationService) processReportGeneration(ctx context.Context, report *domain.GeneratedReport, template *domain.ReportTemplate) {
	startTime := time.Now()
	var err error
	var totalSize int64

	defer func() {
		duration := time.Since(startTime)
		success := err == nil
		s.metrics.RecordReportGenerated(string(report.ReportType), duration, success)
		if err == nil && totalSize > 0 {
			s.metrics.RecordReportSize(string(report.ReportType), totalSize)
		}
	}()

	// Update status to processing
	report.Status = domain.ReportStatusProcessing
	if updateErr := s.reportRepo.Update(ctx, report); updateErr != nil {
		log.Printf("Failed to update report status: %v", updateErr)
	}

	// Generate report data
	data, err := s.collectReportData(ctx, report, template)
	if err != nil {
		s.handleGenerationFailure(ctx, report, err)
		return
	}

	// Generate files for each output format
	var files []domain.ReportFile
	for _, format := range report.OutputFormats {
		file, genErr := s.generateFile(ctx, report, template, format, data)
		if genErr != nil {
			log.Printf("Failed to generate %s file for report %s: %v", format, report.ID, genErr)
			continue
		}
		files = append(files, *file)
	}

	if len(files) == 0 {
		s.handleGenerationFailure(ctx, report, fmt.Errorf("failed to generate any output files"))
		return
	}

	// Calculate total size
	for _, file := range files {
		totalSize += file.FileSize
	}

	// Update report with file information
	report.Files = files
	report.TotalSize = totalSize
	report.Checksum = s.calculateChecksum(data)

	// Record generation completion
	if err := s.reportRepo.RecordGeneration(ctx, report.ID, time.Since(startTime).Milliseconds()); err != nil {
		log.Printf("Failed to record generation: %v", err)
	}

	// Update the report
	if err := s.reportRepo.Update(ctx, report); err != nil {
		log.Printf("Failed to update report: %v", err)
	}

	// Publish Kafka event
	if pubErr := s.kafkaProducer.PublishReportGenerated(ctx, report); pubErr != nil {
		log.Printf("Failed to publish Kafka event: %v", pubErr)
	}

	log.Printf("Successfully generated report %s with %d files", report.ID, len(files))
}

// getTemplateWithCache retrieves a template, using cache if available
func (s *ReportGenerationService) getTemplateWithCache(ctx context.Context, templateID uuid.UUID) (*domain.ReportTemplate, error) {
	if s.cacheRepo == nil {
		return s.templateRepo.GetByID(ctx, templateID)
	}

	// Try cache first
	template, err := s.cacheRepo.GetTemplate(ctx, templateID.String())
	if err != nil || template != nil {
		return template, err
	}

	// Cache miss - get from database
	template, err = s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if template == nil {
		return nil, nil
	}

	// Cache the template
	if cacheErr := s.cacheRepo.SetTemplate(ctx, template); cacheErr != nil {
		log.Printf("Failed to cache template: %v", cacheErr)
	}

	return template, nil
}

// collectReportData collects data for the report
func (s *ReportGenerationService) collectReportData(ctx context.Context, report *domain.GeneratedReport, template *domain.ReportTemplate) ([]byte, error) {
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Query data sources defined in template.TemplateSchema.DataSources
	// 2. Apply transformations defined in template.TemplateSchema.Transformations
	// 3. Validate data against template.TemplateSchema.ValidationRules
	// 4. Return the collected data as JSON

	// For now, return sample data
	sampleData := map[string]interface{}{
		"report_id":      report.ID.String(),
		"template_name":  template.Name,
		"report_type":    report.ReportType,
		"period_start":   report.PeriodStart,
		"period_end":     report.PeriodEnd,
		"generated_at":   time.Now(),
		"data_sources":   template.TemplateSchema.DataSources,
		"sample_metric":  1234.56,
		"record_count":   100,
	}

	return json.Marshal(sampleData)
}

// generateFile generates a file in the specified format
func (s *ReportGenerationService) generateFile(ctx context.Context, report *domain.GeneratedReport, template *domain.ReportTemplate, format domain.ReportFormat, data []byte) (*domain.ReportFile, error) {
	filename := fmt.Sprintf("%s_%s.%s", report.ID.String(), format, format)
	
	// Create a reader from the data
	reader := io.NopCloser(io.Reader(bytes.NewReader(data)))

	// Save the file
	filePath, size, err := s.fileStorage.SaveReport(ctx, report.ID, format, reader)
	if err != nil {
		return nil, err
	}

	// Calculate checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	return &domain.ReportFile{
		ID:           uuid.New(),
		ReportID:     report.ID,
		Format:       format,
		Filename:     filename,
		FilePath:     filePath,
		FileSize:     size,
		Checksum:     checksum,
		ChecksumType: "sha256",
		CreatedAt:    time.Now(),
	}, nil
}

// handleGenerationFailure handles a generation failure
func (s *ReportGenerationService) handleGenerationFailure(ctx context.Context, report *domain.GeneratedReport, err error) {
	if updateErr := s.reportRepo.RecordFailure(ctx, report.ID, err.Error()); updateErr != nil {
		log.Printf("Failed to record failure: %v", updateErr)
	}

	if pubErr := s.kafkaProducer.PublishReportFailed(ctx, report.ID, err.Error()); pubErr != nil {
		log.Printf("Failed to publish failure event: %v", pubErr)
	}

	log.Printf("Failed to generate report %s: %v", report.ID, err)
}

// calculateChecksum calculates a checksum for the data
func (s *ReportGenerationService) calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GenerateReportAsync generates a report asynchronously
func (s *ReportGenerationService) GenerateReportAsync(ctx context.Context, req *GenerateReportRequest) (<-chan *domain.GeneratedReport, <-chan error) {
	reportChan := make(chan *domain.GeneratedReport, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(reportChan)
		defer close(errChan)

		report, err := s.GenerateReport(ctx, req)
		if err != nil {
			errChan <- err
			return
		}

		reportChan <- report
	}()

	return reportChan, errChan
}

// BulkGenerate generates multiple reports
func (s *ReportGenerationService) BulkGenerate(ctx context.Context, requests []*GenerateReportRequest) ([]*domain.GeneratedReport, []error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var reports []*domain.GeneratedReport
	var errors []error

	for _, req := range requests {
		wg.Add(1)
		go func(r *GenerateReportRequest) {
			defer wg.Done()

			report, err := s.GenerateReport(ctx, r)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			reports = append(reports, report)
			mu.Unlock()
		}(req)
	}

	wg.Wait()

	return reports, errors
}

// GetReport retrieves a report by ID
func (s *ReportGenerationService) GetReport(ctx context.Context, reportID uuid.UUID) (*domain.GeneratedReport, error) {
	// Try cache first
	if s.cacheRepo != nil {
		report, err := s.cacheRepo.GetReport(ctx, reportID.String())
		if err == nil && report != nil {
			return report, nil
		}
	}

	// Get from database
	report, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	// Cache the report
	if s.cacheRepo != nil && report != nil {
		s.cacheRepo.SetReport(ctx, report)
	}

	return report, nil
}

// ListReports lists reports with filters
func (s *ReportGenerationService) ListReports(ctx context.Context, filter domain.ReportFilter) ([]*domain.GeneratedReport, int, error) {
	return s.reportRepo.List(ctx, filter)
}

// RegenerateReport regenerates an existing report
func (s *ReportGenerationService) RegenerateReport(ctx context.Context, reportID uuid.UUID, regeneratedBy string) (*domain.GeneratedReport, error) {
	// Get existing report
	existingReport, err := s.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	if existingReport == nil {
		return nil, fmt.Errorf("report not found: %s", reportID)
	}

	// Get template
	template, err := s.getTemplateWithCache(ctx, existingReport.TemplateID)
	if err != nil {
		return nil, err
	}

	// Create new generation request
	req := &GenerateReportRequest{
		TemplateID:       existingReport.TemplateID,
		PeriodStart:      existingReport.PeriodStart,
		PeriodEnd:        existingReport.PeriodEnd,
		GeneratedBy:      regeneratedBy,
		GenerationMethod: "manual",
		OutputFormats:    existingReport.OutputFormats,
		CustomFields:     existingReport.CustomFields,
	}

	// Generate new report
	return s.GenerateReport(ctx, req)
}

// ApproveReport approves a report
func (s *ReportGenerationService) ApproveReport(ctx context.Context, reportID uuid.UUID, approvedBy string) error {
	return s.reportRepo.Approve(ctx, reportID, approvedBy)
}

// ValidateReport validates a report
func (s *ReportGenerationService) ValidateReport(ctx context.Context, reportID uuid.UUID, validatedBy string) error {
	return s.reportRepo.Validate(ctx, reportID, validatedBy)
}

// ArchiveReport archives a report
func (s *ReportGenerationService) ArchiveReport(ctx context.Context, reportID uuid.UUID) error {
	return s.reportRepo.Archive(ctx, reportID)
}

// SubmitReport marks a report as submitted
func (s *ReportGenerationService) SubmitReport(ctx context.Context, reportID uuid.UUID, submittedBy, submissionRef string) error {
	return s.reportRepo.UpdateSubmissionStatus(ctx, reportID, "submitted", submissionRef)
}

// Import necessary packages
import (
	"bytes"
	"encoding/json"
)
