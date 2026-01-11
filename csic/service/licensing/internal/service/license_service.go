package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-licensing/internal/domain/errors"
	"github.com/csic-licensing/internal/domain/models"
	"github.com/csic-licensing/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LicenseService handles license management operations
type LicenseService struct {
	licenseRepo    repository.LicenseRepository
	violationRepo  repository.ViolationRepository
	reportRepo     repository.ReportRepository
	complianceRepo repository.ComplianceRepository
	eventPublisher EventPublisher
	logger         *zap.Logger
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	PublishLicenseEvent(ctx context.Context, event *models.LicenseEvent) error
	PublishComplianceEvent(ctx context.Context, eventType string, data map[string]interface{}) error
}

// NewLicenseService creates a new license service
func NewLicenseService(
	licenseRepo repository.LicenseRepository,
	violationRepo repository.ViolationRepository,
	reportRepo repository.ReportRepository,
	complianceRepo repository.ComplianceRepository,
	eventPublisher EventPublisher,
	logger *zap.Logger,
) *LicenseService {
	return &LicenseService{
		licenseRepo:    licenseRepo,
		violationRepo:  violationRepo,
		reportRepo:     reportRepo,
		complianceRepo: complianceRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateLicense creates a new license for an exchange
func (s *LicenseService) CreateLicense(ctx context.Context, license *models.License) error {
	// Validate license
	if license.ExchangeID == "" {
		return errors.ErrMissingRequiredField
	}
	if license.LicenseNumber == "" {
		return errors.ErrMissingRequiredField
	}

	// Check for duplicate license number
	existing, err := s.licenseRepo.GetByNumber(ctx, license.LicenseNumber)
	if err != nil {
		return fmt.Errorf("failed to check existing license: %w", err)
	}
	if existing != nil {
		return errors.ErrLicenseAlreadyExists
	}

	// Set default values
	if license.ID == "" {
		license.ID = uuid.New().String()
	}
	license.CreatedAt = time.Now()
	license.UpdatedAt = time.Now()

	// Create license
	if err := s.licenseRepo.Create(ctx, license); err != nil {
		return fmt.Errorf("failed to create license: %w", err)
	}

	// Publish event
	event := &models.LicenseEvent{
		ID:        uuid.New().String(),
		LicenseID: license.ID,
		EventType: "created",
		NewStatus: string(license.Status),
		ActorID:   license.ApprovedBy,
		CreatedAt: time.Now(),
	}
	s.eventPublisher.PublishLicenseEvent(ctx, event)

	s.logger.Info("License created",
		zap.String("license_id", license.ID),
		zap.String("exchange_id", license.ExchangeID),
		zap.String("license_number", license.LicenseNumber))

	return nil
}

// GetLicense retrieves a license by ID
func (s *LicenseService) GetLicense(ctx context.Context, id string) (*models.License, error) {
	license, err := s.licenseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get license: %w", err)
	}
	if license == nil {
		return nil, errors.ErrLicenseNotFound
	}
	return license, nil
}

// GetLicenseByNumber retrieves a license by license number
func (s *LicenseService) GetLicenseByNumber(ctx context.Context, number string) (*models.License, error) {
	license, err := s.licenseRepo.GetByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get license: %w", err)
	}
	if license == nil {
		return nil, errors.ErrLicenseNotFound
	}
	return license, nil
}

// GetLicensesByExchange retrieves all licenses for an exchange
func (s *LicenseService) GetLicensesByExchange(ctx context.Context, exchangeID string) ([]*models.License, error) {
	return s.licenseRepo.GetByExchange(ctx, exchangeID)
}

// GetActiveLicenses retrieves all active licenses
func (s *LicenseService) GetActiveLicenses(ctx context.Context) ([]*models.License, error) {
	return s.licenseRepo.GetByStatus(ctx, models.LicenseStatusActive)
}

// GetExpiringLicenses retrieves licenses expiring within specified days
func (s *LicenseService) GetExpiringLicenses(ctx context.Context, days int) ([]*models.License, error) {
	return s.licenseRepo.GetExpiringSoon(ctx, days)
}

// UpdateLicenseStatus updates the status of a license
func (s *LicenseService) UpdateLicenseStatus(ctx context.Context, id string, newStatus models.LicenseStatus, actorID, reason string) error {
	license, err := s.GetLicense(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := license.Status

	// Validate status transition
	if err := license.TransitionStatus(newStatus); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update license
	license.UpdatedAt = time.Now()
	if err := s.licenseRepo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to update license: %w", err)
	}

	// Publish event
	event := &models.LicenseEvent{
		ID:        uuid.New().String(),
		LicenseID: license.ID,
		EventType: "status_changed",
		OldStatus: string(oldStatus),
		NewStatus: string(newStatus),
		Details:   reason,
		ActorID:   actorID,
		CreatedAt: time.Now(),
	}
	s.eventPublisher.PublishLicenseEvent(ctx, event)

	// If license is suspended or revoked, create compliance alerts
	if newStatus == models.LicenseStatusSuspended || newStatus == models.LicenseStatusRevoked {
		s.eventPublisher.PublishComplianceEvent(ctx, "license_suspended", map[string]interface{}{
			"license_id":  license.ID,
			"exchange_id": license.ExchangeID,
			"reason":      reason,
		})
	}

	s.logger.Info("License status updated",
		zap.String("license_id", license.ID),
		zap.String("old_status", oldStatus),
		zap.String("new_status", newStatus),
		zap.String("actor_id", actorID))

	return nil
}

// SuspendLicense suspends a license
func (s *LicenseService) SuspendLicense(ctx context.Context, id string, reason, actorID string) error {
	return s.UpdateLicenseStatus(ctx, id, models.LicenseStatusSuspended, actorID, reason)
}

// RevokeLicense revokes a license
func (s *LicenseService) RevokeLicense(ctx context.Context, id string, reason, actorID string) error {
	return s.UpdateLicenseStatus(ctx, id, models.LicenseStatusRevoked, actorID, reason)
}

// RenewLicense renews a license
func (s *LicenseService) RenewLicense(ctx context.Context, id string, newExpiryDate time.Time, actorID string) error {
	license, err := s.GetLicense(ctx, id)
	if err != nil {
		return err
	}

	oldExpiry := license.ExpiryDate
	license.ExpiryDate = newExpiryDate
	license.Status = models.LicenseStatusActive
	license.UpdatedAt = time.Now()

	if err := s.licenseRepo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to renew license: %w", err)
	}

	event := &models.LicenseEvent{
		ID:        uuid.New().String(),
		LicenseID: license.ID,
		EventType: "renewed",
		OldStatus: string(models.LicenseStatusPendingRenewal),
		NewStatus: string(models.LicenseStatusActive),
		Details:   fmt.Sprintf("Expiry extended from %s to %s", oldExpiry.Format(time.RFC3339), newExpiryDate.Format(time.RFC3339)),
		ActorID:   actorID,
		CreatedAt: time.Now(),
	}
	s.eventPublisher.PublishLicenseEvent(ctx, event)

	s.logger.Info("License renewed",
		zap.String("license_id", license.ID),
		zap.Time("old_expiry", oldExpiry),
		zap.Time("new_expiry", newExpiryDate))

	return nil
}

// ListLicenses retrieves licenses with filtering and pagination
func (s *LicenseService) ListLicenses(ctx context.Context, filter repository.LicenseFilter) ([]*models.License, int64, error) {
	licenses, err := s.licenseRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list licenses: %w", err)
	}

	count, err := s.licenseRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count licenses: %w", err)
	}

	return licenses, count, nil
}

// ValidateLicenseForOperation checks if a license permits a specific operation
func (s *LicenseService) ValidateLicenseForOperation(ctx context.Context, licenseID string, operation string) error {
	license, err := s.GetLicense(ctx, licenseID)
	if err != nil {
		return err
	}

	if !license.IsValid() {
		return errors.ErrLicenseExpired
	}

	switch operation {
	case "trading":
		if !license.CanTrade() {
			return errors.ErrLicenseSuspended
		}
	case "deposit":
		if !license.IsValid() {
			return errors.ErrLicenseExpired
		}
	case "withdrawal":
		if !license.IsValid() {
			return errors.ErrLicenseExpired
		}
	}

	return nil
}

// GetLicenseComplianceStatus retrieves the compliance status for a license
func (s *LicenseService) GetLicenseComplianceStatus(ctx context.Context, licenseID string) (*ComplianceStatus, error) {
	license, err := s.GetLicense(ctx, licenseID)
	if err != nil {
		return nil, err
	}

	// Get violation stats for this license
	stats, err := s.violationRepo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get violation stats: %w", err)
	}

	// Get compliance checks
	checks, err := s.complianceRepo.GetComplianceChecks(ctx, licenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance checks: %w", err)
	}

	status := &ComplianceStatus{
		LicenseStatus:     license.Status,
		LicenseValid:      license.IsValid(),
		ExpiryDate:        license.ExpiryDate,
		DaysUntilExpiry:   int(time.Until(license.ExpiryDate).Hours() / 24),
		TotalViolations:   stats.TotalViolations,
		OpenViolations:    stats.OpenViolations,
		CriticalViolations: stats.CriticalCount,
		ComplianceChecks:  len(checks),
		CompliantChecks:   0,
	}

	for _, check := range checks {
		if check.Status == "compliant" {
			status.CompliantChecks++
		}
	}

	return status, nil
}

// ComplianceStatus represents the compliance status of a license
type ComplianceStatus struct {
	LicenseStatus     models.LicenseStatus `json:"license_status"`
	LicenseValid      bool                 `json:"license_valid"`
	ExpiryDate        time.Time            `json:"expiry_date"`
	DaysUntilExpiry   int                  `json:"days_until_expiry"`
	TotalViolations   int64                `json:"total_violations"`
	OpenViolations    int64                `json:"open_violations"`
	CriticalViolations int64               `json:"critical_violations"`
	ComplianceChecks  int                  `json:"compliance_checks"`
	CompliantChecks   int                  `json:"compliant_checks"`
}
