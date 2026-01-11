package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-licensing/internal/domain/models"
	"github.com/csic-licensing/internal/repository"
	"go.uber.org/zap"
)

// ComplianceService handles compliance monitoring and dashboard operations
type ComplianceService struct {
	licenseRepo    repository.LicenseRepository
	violationRepo  repository.ViolationRepository
	complianceRepo repository.ComplianceRepository
	reportRepo     repository.ReportRepository
	logger         *zap.Logger
}

// NewComplianceService creates a new compliance service
func NewComplianceService(
	licenseRepo repository.LicenseRepository,
	violationRepo repository.ViolationRepository,
	complianceRepo repository.ComplianceRepository,
	reportRepo repository.ReportRepository,
	logger *zap.Logger,
) *ComplianceService {
	return &ComplianceService{
		licenseRepo:    licenseRepo,
		violationRepo:  violationRepo,
		complianceRepo: complianceRepo,
		reportRepo:     reportRepo,
		logger:         logger,
	}
}

// GetDashboardMetrics retrieves comprehensive dashboard metrics
func (s *ComplianceService) GetDashboardMetrics(ctx context.Context, exchangeID string) (*models.DashboardMetrics, error) {
	metrics := models.NewDashboardMetrics()

	// Get license metrics
	licenses, err := s.licenseRepo.GetByExchange(ctx, exchangeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get licenses: %w", err)
	}

	for _, lic := range licenses {
		metrics.TotalLicenses++
		switch lic.Status {
		case models.LicenseStatusActive:
			metrics.ActiveLicenses++
		case models.LicenseStatusSuspended:
			metrics.SuspendedLicenses++
		}
		if lic.IsExpiringSoon(30) {
			metrics.ExpiringSoon++
		}
		if lic.Status == models.LicenseStatusPendingRenewal {
			metrics.PendingRenewals++
		}
	}

	// Get violation metrics
	violationStats, err := s.violationRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get violation stats: %w", err)
	}

	metrics.TotalViolations = violationStats.TotalViolations
	metrics.OpenViolations = violationStats.OpenViolations
	metrics.CriticalViolations = violationStats.CriticalCount
	metrics.HighViolations = violationStats.HighCount

	// Get report metrics
	reportStats, err := s.reportRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get report stats: %w", err)
	}

	metrics.TotalReports = reportStats.TotalReports
	metrics.PendingReports = reportStats.PendingReports
	metrics.SubmittedReports = reportStats.SubmittedReports
	metrics.OverdueReports = reportStats.OverdueReports

	// Get compliance score
	complianceScore, err := s.complianceRepo.CalculateComplianceScore(ctx, exchangeID)
	if err != nil {
		s.logger.Warn("Failed to calculate compliance score", zap.Error(err))
	} else {
		metrics.AverageComplianceScore = complianceScore.OverallScore
	}

	// Get recent activity
	activity, err := s.complianceRepo.GetRecentActivity(ctx, 20)
	if err != nil {
		s.logger.Warn("Failed to get recent activity", zap.Error(err))
	} else {
		metrics.RecentActivity = activity
	}

	// Get violation trends
	trends, err := s.violationRepo.GetTrends(ctx, time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		s.logger.Warn("Failed to get violation trends", zap.Error(err))
	} else {
		metrics.ViolationTrends = trends
	}

	// Get license expiry timeline
	expiryTimeline, err := s.complianceRepo.GetLicenseExpiryTimeline(ctx, 90)
	if err != nil {
		s.logger.Warn("Failed to get expiry timeline", zap.Error(err))
	} else {
		metrics.LicenseExpiryTimeline = expiryTimeline
	}

	metrics.LastUpdated = time.Now()
	return metrics, nil
}

// GetViolationStats retrieves detailed violation statistics
func (s *ComplianceService) GetViolationStats(ctx context.Context) (*models.ViolationStats, error) {
	return s.violationRepo.GetStats(ctx)
}

// GetViolationTrends retrieves violation trends over time
func (s *ComplianceService) GetViolationTrends(ctx context.Context, startDate, endDate time.Time) ([]models.ViolationTrend, error) {
	return s.violationRepo.GetTrends(ctx, startDate, endDate)
}

// GetOpenViolations retrieves all open violations
func (s *ComplianceService) GetOpenViolations(ctx context.Context) ([]*models.ComplianceViolation, error) {
	return s.violationRepo.GetOpenViolations(ctx)
}

// GetViolationsByFilter retrieves violations matching the filter
func (s *ComplianceService) GetViolationsByFilter(ctx context.Context, filter repository.ViolationFilter) ([]*models.ComplianceViolation, int64, error) {
	violations, err := s.violationRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list violations: %w", err)
	}

	count, err := s.violationRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count violations: %w", err)
	}

	return violations, count, nil
}

// CreateViolation creates a new violation record
func (s *ComplianceService) CreateViolation(ctx context.Context, violation *models.ComplianceViolation) error {
	// Validate license exists
	license, err := s.licenseRepo.GetByID(ctx, violation.LicenseID)
	if err != nil {
		return fmt.Errorf("failed to validate license: %w", err)
	}
	if license == nil {
		return fmt.Errorf("license not found")
	}

	violation.ExchangeID = license.ExchangeID
	violation.CreatedAt = time.Now()
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Create(ctx, violation); err != nil {
		return fmt.Errorf("failed to create violation: %w", err)
	}

	// Log activity
	activity := models.NewActivityItem(
		"violation_detected",
		fmt.Sprintf("New %s violation detected", violation.Severity),
		violation.Description,
		violation.ID,
		"violation",
		"system",
		"System",
	)
	_ = activity

	s.logger.Info("Violation created",
		zap.String("violation_id", violation.ID),
		zap.String("license_id", violation.LicenseID),
		zap.String("severity", string(violation.Severity)))

	return nil
}

// UpdateViolationStatus updates the status of a violation
func (s *ComplianceService) UpdateViolationStatus(ctx context.Context, id string, status models.ViolationStatus, updatedBy, reason string) error {
	violation, err := s.violationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get violation: %w", err)
	}
	if violation == nil {
		return fmt.Errorf("violation not found")
	}

	oldStatus := violation.Status
	violation.Status = status
	violation.UpdatedAt = time.Now()

	if status == models.ViolationStatusResolved {
		now := time.Now()
		violation.ResolvedAt = &now
		violation.ResolvedBy = updatedBy
		violation.Resolution = reason
	}

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to update violation: %w", err)
	}

	// Create event
	event := &models.ViolationEvent{
		ID:          id,
		ViolationID: id,
		EventType:   "status_changed",
		OldStatus:   string(oldStatus),
		NewStatus:   string(status),
		ChangedBy:   updatedBy,
		ChangeReason: reason,
		CreatedAt:   time.Now(),
	}
	_ = event

	s.logger.Info("Violation status updated",
		zap.String("violation_id", id),
		zap.String("old_status", oldStatus),
		zap.String("new_status", status))

	return nil
}

// AssignViolation assigns a violation to an investigator
func (s *ComplianceService) AssignViolation(ctx context.Context, id string, assignee, team string) error {
	violation, err := s.violationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get violation: %w", err)
	}
	if violation == nil {
		return fmt.Errorf("violation not found")
	}

	violation.Assign(assignee, team)
	violation.UpdatedAt = time.Now()

	return s.violationRepo.Update(ctx, violation)
}

// EscalateViolation escalates a violation to a higher authority
func (s *ComplianceService) EscalateViolation(ctx context.Context, id string, escalateTo string) error {
	violation, err := s.violationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get violation: %w", err)
	}
	if violation == nil {
		return fmt.Errorf("violation not found")
	}

	violation.Escalate(escalateTo)
	violation.UpdatedAt = time.Now()

	return s.violationRepo.Update(ctx, violation)
}

// ResolveViolation resolves a violation with full details
func (s *ComplianceService) ResolveViolation(ctx context.Context, id string, resolvedBy, resolution, rootCause, correctiveAction, preventiveAction string) error {
	violation, err := s.violationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get violation: %w", err)
	}
	if violation == nil {
		return fmt.Errorf("violation not found")
	}

	violation.Resolve(resolvedBy, resolution, rootCause, correctiveAction, preventiveAction)

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to resolve violation: %w", err)
	}

	s.logger.Info("Violation resolved",
		zap.String("violation_id", id),
		zap.String("resolved_by", resolvedBy))

	return nil
}

// GetComplianceScore calculates the overall compliance score for an exchange
func (s *ComplianceService) GetComplianceScore(ctx context.Context, exchangeID string) (*models.ComplianceScore, error) {
	return s.complianceRepo.CalculateComplianceScore(ctx, exchangeID)
}

// GetKPIMetrics retrieves key performance indicators
func (s *ComplianceService) GetKPIMetrics(ctx context.Context, exchangeID string) ([]*models.KPIMetric, error) {
	return s.complianceRepo.GetKPIMetrics(ctx, exchangeID)
}

// GetComplianceChecks retrieves compliance checks for a license
func (s *ComplianceService) GetComplianceChecks(ctx context.Context, licenseID string) ([]*models.ComplianceCheck, error) {
	return s.complianceRepo.GetComplianceChecks(ctx, licenseID)
}

// UpdateComplianceCheck updates a compliance check
func (s *ComplianceService) UpdateComplianceCheck(ctx context.Context, check *models.ComplianceCheck) error {
	return s.complianceRepo.UpdateComplianceCheck(ctx, check)
}

// CreateComplianceCheck creates a new compliance check
func (s *ComplianceService) CreateComplianceCheck(ctx context.Context, check *models.ComplianceCheck) error {
	return s.complianceRepo.CreateComplianceCheck(ctx, check)
}

// GetRecentActivity retrieves recent compliance activity
func (s *ComplianceService) GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityItem, error) {
	return s.complianceRepo.GetRecentActivity(ctx, limit)
}
