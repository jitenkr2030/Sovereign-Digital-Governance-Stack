package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/reporting-service/reporting/internal/core/domain"
	"github.com/reporting-service/reporting/internal/core/ports"
)

// Common errors for reporting operations.
var (
	ErrReportNotFound        = errors.New("report not found")
	ErrReportAlreadyExists   = errors.New("report already exists")
	ErrInvalidReportStatus   = errors.New("invalid report status transition")
	ErrAlertNotFound         = errors.New("alert not found")
	ErrRuleNotFound          = errors.New("compliance rule not found")
	ErrFilingNotFound        = errors.New("filing record not found")
	ErrFilingFailed          = errors.New("filing submission failed")
	ErrSubjectNot Screened   = errors.New("subject has not been screened")
)

// ReportingService provides the core business logic for regulatory reporting.
type ReportingService struct {
	sarRepo       ports.SARRepository
	ctrRepo       ports.CTRRepository
	ruleRepo      ports.ComplianceRuleRepository
	alertRepo     ports.AlertRepository
	checkRepo     ports.ComplianceCheckRepository
	reportRepo    ports.ComplianceReportRepository
	filingRepo    ports.FilingRecordRepository
	transactionRepo ports.TransactionRepository
	screeningRepo ports.ScreeningRepository
}

// NewReportingService creates a new ReportingService with the required dependencies.
func NewReportingService(
	sarRepo ports.SARRepository,
	ctrRepo ports.CTRRepository,
	ruleRepo ports.ComplianceRuleRepository,
	alertRepo ports.AlertRepository,
	checkRepo ports.ComplianceCheckRepository,
	reportRepo ports.ComplianceReportRepository,
	filingRepo ports.FilingRecordRepository,
	transactionRepo ports.TransactionRepository,
	screeningRepo ports.ScreeningRepository,
) *ReportingService {
	return &ReportingService{
		sarRepo:        sarRepo,
		ctrRepo:        ctrRepo,
		ruleRepo:       ruleRepo,
		alertRepo:      alertRepo,
		checkRepo:      checkRepo,
		reportRepo:     reportRepo,
		filingRepo:     filingRepo,
		transactionRepo: transactionRepo,
		screeningRepo:  screeningRepo,
	}
}

// ==================== SAR Operations ====================

// CreateSAR creates a new Suspicious Activity Report.
func (s *ReportingService) CreateSAR(ctx context.Context, sar *domain.SAR) error {
	sar.ReportNumber = s.generateReportNumber("SAR")
	sar.Status = domain.ReportStatusDraft
	sar.CreatedAt = time.Now().UTC()
	sar.UpdatedAt = time.Now().UTC()

	if err := s.sarRepo.Create(ctx, sar); err != nil {
		return fmt.Errorf("failed to create SAR: %w", err)
	}

	return nil
}

// GetSAR retrieves a SAR by ID.
func (s *ReportingService) GetSAR(ctx context.Context, id string) (*domain.SAR, error) {
	sar, err := s.sarRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get SAR: %w", err)
	}
	if sar == nil {
		return nil, ErrReportNotFound
	}
	return sar, nil
}

// ListSARs lists SARs with optional filtering.
func (s *ReportingService) ListSARs(ctx context.Context, filter ports.SARFilter) ([]*domain.SAR, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return s.sarRepo.List(ctx, filter)
}

// UpdateSARStatus updates the status of a SAR.
func (s *ReportingService) UpdateSARStatus(ctx context.Context, id string, newStatus domain.ReportStatus, reviewerID string) error {
	sar, err := s.GetSAR(ctx, id)
	if err != nil {
		return err
	}

	// Validate status transition
	if err := s.validateStatusTransition(sar.Status, newStatus); err != nil {
		return err
	}

	sar.Status = newStatus
	sar.ReviewerID = reviewerID
	sar.UpdatedAt = time.Now().UTC()

	if err := s.sarRepo.Update(ctx, sar); err != nil {
		return fmt.Errorf("failed to update SAR: %w", err)
	}

	return nil
}

// SubmitSAR submits a SAR to the appropriate agency.
func (s *ReportingService) SubmitSAR(ctx context.Context, id string, filingMethod, filingAgency string) error {
	sar, err := s.GetSAR(ctx, id)
	if err != nil {
		return err
	}

	if sar.Status != domain.ReportStatusApproved {
		return errors.New("SAR must be approved before submission")
	}

	// Create filing record
	filing := &domain.FilingRecord{
		ID:           domain.NewEntityID(),
		ReportID:     sar.ID,
		ReportType:   domain.ReportTypeSAR,
		ReportNumber: sar.ReportNumber,
		FilingMethod: filingMethod,
		FilingAgency: filingAgency,
		FilingStatus: "submitted",
		SubmittedAt:  time.Now().UTC(),
	}

	if err := s.filingRepo.Create(ctx, filing); err != nil {
		return fmt.Errorf("failed to create filing record: %w", err)
	}

	now := time.Now()
	sar.Status = domain.ReportStatusSubmitted
	sar.SubmittedAt = &now
	sar.UpdatedAt = time.Now().UTC()

	if err := s.sarRepo.Update(ctx, sar); err != nil {
		return fmt.Errorf("failed to update SAR: %w", err)
	}

	return nil
}

// ==================== CTR Operations ====================

// CreateCTR creates a new Currency Transaction Report.
func (s *ReportingService) CreateCTR(ctx context.Context, ctr *domain.CTR) error {
	ctr.ReportNumber = s.generateReportNumber("CTR")
	ctr.Status = domain.ReportStatusDraft
	ctr.CreatedAt = time.Now().UTC()
	ctr.UpdatedAt = time.Now().UTC()

	if err := s.ctrRepo.Create(ctx, ctr); err != nil {
		return fmt.Errorf("failed to create CTR: %w", err)
	}

	return nil
}

// GetCTR retrieves a CTR by ID.
func (s *ReportingService) GetCTR(ctx context.Context, id string) (*domain.CTR, error) {
	ctr, err := s.ctrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get CTR: %w", err)
	}
	if ctr == nil {
		return nil, ErrReportNotFound
	}
	return ctr, nil
}

// ListCTRs lists CTRs with optional filtering.
func (s *ReportingService) ListCTRs(ctx context.Context, filter ports.CTRFilter) ([]*domain.CTR, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return s.ctrRepo.List(ctx, filter)
}

// ==================== Compliance Rule Operations ====================

// CreateComplianceRule creates a new compliance rule.
func (s *ReportingService) CreateComplianceRule(ctx context.Context, rule *domain.ComplianceRule) error {
	// Check for existing rule with same code
	existing, err := s.ruleRepo.GetByCode(ctx, rule.RuleCode)
	if err != nil {
		return fmt.Errorf("failed to check existing rule: %w", err)
	}
	if existing != nil {
		return errors.New("rule with this code already exists")
	}

	rule.CreatedAt = time.Now().UTC()
	rule.UpdatedAt = time.Now().UTC()

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return fmt.Errorf("failed to create compliance rule: %w", err)
	}

	return nil
}

// GetComplianceRule retrieves a compliance rule by ID.
func (s *ReportingService) GetComplianceRule(ctx context.Context, id string) (*domain.ComplianceRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance rule: %w", err)
	}
	if rule == nil {
		return nil, ErrRuleNotFound
	}
	return rule, nil
}

// ListComplianceRules lists compliance rules.
func (s *ReportingService) ListComplianceRules(ctx context.Context, filter ports.ComplianceRuleFilter) ([]*domain.ComplianceRule, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	return s.ruleRepo.List(ctx, filter)
}

// ==================== Alert Operations ====================

// CreateAlert creates a new compliance alert.
func (s *ReportingService) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	alert.AlertNumber = s.generateAlertNumber()
	alert.Status = "open"
	alert.CreatedAt = time.Now().UTC()
	alert.UpdatedAt = time.Now().UTC()

	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetAlert retrieves an alert by ID.
func (s *ReportingService) GetAlert(ctx context.Context, id string) (*domain.Alert, error) {
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	if alert == nil {
		return nil, ErrAlertNotFound
	}
	return alert, nil
}

// ListAlerts lists alerts with optional filtering.
func (s *ReportingService) ListAlerts(ctx context.Context, filter ports.AlertFilter) ([]*domain.Alert, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return s.alertRepo.List(ctx, filter)
}

// ResolveAlert resolves an alert.
func (s *ReportingService) ResolveAlert(ctx context.Context, id, resolution, resolvedBy string) error {
	alert, err := s.GetAlert(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	alert.Status = "resolved"
	alert.Resolution = resolution
	alert.ResolvedAt = &now
	alert.ResolvedBy = resolvedBy
	alert.UpdatedAt = time.Now().UTC()

	if err := s.alertRepo.Update(ctx, alert); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	return nil
}

// EscalateAlert escalates an alert.
func (s *ReportingService) EscalateAlert(ctx context.Context, id, assignedTo string) error {
	alert, err := s.GetAlert(ctx, id)
	if err != nil {
		return err
	}

	alert.Status = "escalated"
	alert.AssignedTo = assignedTo
	alert.UpdatedAt = time.Now().UTC()

	if err := s.alertRepo.Update(ctx, alert); err != nil {
		return fmt.Errorf("failed to escalate alert: %w", err)
	}

	return nil
}

// ==================== Compliance Check Operations ====================

// PerformComplianceCheck performs a compliance check on a subject.
func (s *ReportingService) PerformComplianceCheck(
	ctx context.Context,
	subjectID, subjectType, checkType string,
) (*domain.ComplianceCheck, error) {
	// Perform screening
	matches, err := s.screeningRepo.Search(ctx, subjectID, []string{"ofac", "eu_sanctions", "pep"})
	if err != nil {
		return nil, fmt.Errorf("screening failed: %w", err)
	}

	// Evaluate check result
	result := "pass"
	riskLevel := "low"
	matchScore := 0.0

	if len(matches) > 0 {
		result = "review_required"
		riskLevel = "high"
		matchScore = 0.95
	}

	check := &domain.ComplianceCheck{
		ID:           domain.NewEntityID(),
		SubjectID:    subjectID,
		SubjectType:  subjectType,
		CheckType:    checkType,
		CheckResult:  result,
		RiskLevel:    riskLevel,
		MatchScore:   matchScore,
		MatchedEntity: "",
		ScreeningList: "ofac",
		CheckedAt:    time.Now().UTC(),
		ExpiresAt:    func() *time.Time { t := time.Now().AddDate(1, 0, 0); return &t }(),
	}

	if err := s.checkRepo.Create(ctx, check); err != nil {
		return nil, fmt.Errorf("failed to save compliance check: %w", err)
	}

	return check, nil
}

// ==================== Report Generation ====================

// GenerateComplianceReport generates a compliance report for a period.
func (s *ReportingService) GenerateComplianceReport(
	ctx context.Context,
	reportType domain.ReportType,
	periodStart, periodEnd time.Time,
	generatedBy string,
) (*domain.ComplianceReport, error) {
	// Gather data based on report type
	var summary domain.ReportSummary
	var details interface{}

	switch reportType {
	case domain.ReportTypeSAR:
		sars, _ := s.sarRepo.List(ctx, ports.SARFilter{
			StartDate: &periodStart,
			EndDate:   &periodEnd,
			Limit:     10000,
		})
		summary.SuspiciousActivity = int64(len(sars))
		details = sars

	case domain.ReportTypeCTR:
		ctrs, _ := s.ctrRepo.List(ctx, ports.CTRFilter{
			StartDate: &periodStart,
			EndDate:   &periodEnd,
			Limit:     10000,
		})
		summary.TotalTransactions = int64(len(ctrs))
		for _, ctr := range ctrs {
			summary.TotalVolume += ctr.TotalAmount
		}
		details = ctrs

	case domain.ReportTypeInternalAlert:
		alerts, _ := s.alertRepo.List(ctx, ports.AlertFilter{
			StartDate: &periodStart,
			EndDate:   &periodEnd,
			Limit:     10000,
		})
		summary.AlertsGenerated = int64(len(alerts))
		for _, alert := range alerts {
			if alert.Status == "resolved" {
				summary.AlertsResolved++
			}
			switch alert.Severity {
			case "high":
				summary.HighRiskCount++
			case "medium":
				summary.MediumRiskCount++
			case "low":
				summary.LowRiskCount++
			}
		}
		details = alerts
	}

	report := &domain.ComplianceReport{
		ID:          domain.NewEntityID(),
		ReportType:  reportType,
		ReportName:  fmt.Sprintf("%s Report - %s to %s", reportType, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02")),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Summary:     summary,
		Details:     details,
		GeneratedBy: generatedBy,
		GeneratedAt: time.Now().UTC(),
		Status:      domain.ReportStatusDraft,
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create compliance report: %w", err)
	}

	return report, nil
}

// ==================== Helper Methods ====================

// validateStatusTransition validates that a status transition is allowed.
func (s *ReportingService) validateStatusTransition(current, new domain.ReportStatus) error {
	allowedTransitions := map[domain.ReportStatus][]domain.ReportStatus{
		domain.ReportStatusDraft:     {domain.ReportStatusPending},
		domain.ReportStatusPending:   {domain.ReportStatusApproved, domain.ReportStatusRejected},
		domain.ReportStatusApproved:  {domain.ReportStatusSubmitted, domain.ReportStatusRejected},
		domain.ReportStatusRejected:  {domain.ReportStatusDraft, domain.ReportStatusClosed},
		domain.ReportStatusSubmitted: {domain.ReportStatusClosed},
		domain.ReportStatusClosed:    {},
	}

	allowed, ok := allowedTransitions[current]
	if !ok {
		return errors.New("invalid current status")
	}

	for _, status := range allowed {
		if status == new {
			return nil
		}
	}

	return ErrInvalidReportStatus
}

// generateReportNumber generates a unique report number.
func (s *ReportingService) generateReportNumber(reportType string) string {
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	timestamp := time.Now().Format("150405")
	return fmt.Sprintf("%s-%s%s%s-%s", reportType, year, month, day, timestamp)
}

// generateAlertNumber generates a unique alert number.
func (s *ReportingService) generateAlertNumber() string {
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	timestamp := time.Now().Format("150405")
	return fmt.Sprintf("ALT-%s%s%s-%s", year, month, day, timestamp)
}

// EvaluateComplianceRules evaluates all active compliance rules against a transaction.
func (s *ReportingService) EvaluateComplianceRules(ctx context.Context, transaction interface{}) ([]string, error) {
	rules, err := s.ruleRepo.List(ctx, ports.ComplianceRuleFilter{
		IsActive: func() *bool { b := true; return &b }(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list compliance rules: %w", err)
	}

	var triggeredRules []string
	for _, rule := range rules {
		// Evaluate rule logic (simplified - would implement full logic evaluation)
		if s.evaluateRule(rule, transaction) {
			triggeredRules = append(triggeredRules, rule.RuleCode)
		}
	}

	return triggeredRules, nil
}

// evaluateRule evaluates a single compliance rule against a transaction.
func (s *ReportingService) evaluateRule(rule *domain.ComplianceRule, transaction interface{}) bool {
	// Simplified rule evaluation
	// In production, this would parse and evaluate JSON logic
	return false
}
