// Compliance Management Module - Licensing Service
// Business logic for license lifecycle management

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
	"github.com/csic-platform/compliance/internal/port"
)

// LicensingService handles license operations
type LicensingService struct {
	repo      port.LicenseRepository
	entityRepo port.EntityRepository
	audit     port.AuditLogPort
}

// NewLicensingService creates a new licensing service
func NewLicensingService(repo port.LicenseRepository, entityRepo port.EntityRepository, audit port.AuditLogPort) *LicensingService {
	return &LicensingService{
		repo:      repo,
		entityRepo: entityRepo,
		audit:     audit,
	}
}

// CreateLicense creates a new license application
func (s *LicensingService) CreateLicense(ctx context.Context, license *domain.License, actorID string) error {
	// Validate license
	if err := license.Validate(); err != nil {
		return err
	}

	// Verify entity exists and can apply for license
	entity, err := s.entityRepo.GetByID(ctx, license.EntityID)
	if err != nil {
		return fmt.Errorf("entity not found: %w", err)
	}

	if !entity.CanApplyForLicense() {
		return domain.ErrEntityInactive
	}

	// Check for duplicate active license
	activeLicense, err := s.repo.GetActiveByEntityAndType(ctx, license.EntityID, license.Type)
	if err != nil && !errors.Is(err, domain.ErrLicenseNotFound) {
		return err
	}
	if activeLicense != nil {
		return domain.ErrDuplicateLicense
	}

	// Set defaults
	license.Status = domain.LicenseStatusDraft
	license.CreatedAt = time.Now()
	license.UpdatedAt = time.Now()
	license.LicenseNumber = license.GenerateLicenseNumber()

	// Create license
	if err := s.repo.Create(ctx, license); err != nil {
		return fmt.Errorf("failed to create license: %w", err)
	}

	// Update entity with license ID
	entity.AddLicense(license.ID)
	if err := s.entityRepo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_CREATED",
		ResourceType:  "LICENSE",
		ResourceID:    license.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Created license application %s for entity %s", license.LicenseNumber, entity.Name),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// SubmitLicense submits a license application for review
func (s *LicensingService) SubmitLicense(ctx context.Context, licenseID string, actorID string) error {
	license, err := s.repo.GetByID(ctx, licenseID)
	if err != nil {
		return err
	}

	if err := license.Submit(); err != nil {
		return err
	}

	license.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to submit license: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_SUBMITTED",
		ResourceType:  "LICENSE",
		ResourceID:    license.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Submitted license application %s for review", license.LicenseNumber),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// ApproveLicense approves a license
func (s *LicensingService) ApproveLicense(ctx context.Context, licenseID string, officerID string, conditions []domain.LicenseCondition) error {
	license, err := s.repo.GetByID(ctx, licenseID)
	if err != nil {
		return err
	}

	if err := license.Approve(officerID); err != nil {
		return err
	}

	// Add any conditions
	for _, condition := range conditions {
		license.AddCondition(condition)
	}

	license.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to approve license: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       officerID,
		Action:        "LICENSE_APPROVED",
		ResourceType:  "LICENSE",
		ResourceID:    license.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Approved license %s", license.LicenseNumber),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"license_number": license.LicenseNumber,
			"conditions":     len(conditions),
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// ActivateLicense activates an approved license
func (s *LicensingService) ActivateLicense(ctx context.Context, licenseID string, actorID string) error {
	license, err := s.repo.GetByID(ctx, licenseID)
	if err != nil {
		return err
	}

	if err := license.Activate(); err != nil {
		return err
	}

	license.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to activate license: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_ACTIVATED",
		ResourceType:  "LICENSE",
		ResourceID:    license.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Activated license %s", license.LicenseNumber),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// SuspendLicense suspends a license
func (s *LicensingService) SuspendLicense(ctx context.Context, licenseID string, reason string, actorID string) error {
	license, err := s.repo.GetByID(ctx, licenseID)
	if err != nil {
		return err
	}

	if err := license.Suspend(reason); err != nil {
		return err
	}

	license.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to suspend license: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_SUSPENDED",
		ResourceType:  "LICENSE",
		ResourceID:    license.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Suspended license %s - Reason: %s", license.LicenseNumber, reason),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// RevokeLicense revokes a license
func (s *LicensingService) RevokeLicense(ctx context.Context, licenseID string, reason string, actorID string) error {
	license, err := s.repo.GetByID(ctx, licenseID)
	if err != nil {
		return err
	}

	if err := license.Revoke(reason); err != nil {
		return err
	}

	license.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to revoke license: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_REVOKED",
		ResourceType:  "LICENSE",
		ResourceID:    license.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Revoked license %s - Reason: %s", license.LicenseNumber, reason),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// RenewLicense renews a license
func (s *LicensingService) RenewLicense(ctx context.Context, licenseID string, newExpiryDate time.Time, actorID string) error {
	license, err := s.repo.GetByID(ctx, licenseID)
	if err != nil {
		return err
	}

	if !license.IsActive() {
		return domain.ErrLicenseInactive
	}

	// Create new license record from old one
	newLicense := *license
	newLicense.ID = ""
	newLicense.Status = domain.LicenseStatusActive
	newLicense.IssuedAt = time.Now()
	newLicense.EffectiveDate = time.Now()
	newLicense.ExpiresAt = newExpiryDate
	newLicense.PreviousLicense = license.ID
	newLicense.LicenseNumber = newLicense.GenerateLicenseNumber()
	newLicense.CreatedAt = time.Now()
	newLicense.UpdatedAt = time.Now()

	// Create new license
	if err := s.repo.Create(ctx, &newLicense); err != nil {
		return fmt.Errorf("failed to create renewed license: %w", err)
	}

	// Mark old license as expired
	if err := license.Expire(); err != nil {
		return err
	}
	license.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, license); err != nil {
		return fmt.Errorf("failed to update old license: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_RENEWED",
		ResourceType:  "LICENSE",
		ResourceID:    newLicense.ID,
		EntityID:      license.EntityID,
		Description:   fmt.Sprintf("Renewed license %s to %s", license.LicenseNumber, newLicense.LicenseNumber),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"previous_license": license.LicenseNumber,
			"new_license":      newLicense.LicenseNumber,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetLicense retrieves a license by ID
func (s *LicensingService) GetLicense(ctx context.Context, licenseID string) (*domain.License, error) {
	return s.repo.GetByID(ctx, licenseID)
}

// ListLicenses lists licenses with filters
func (s *LicensingService) ListLicenses(ctx context.Context, filter port.LicenseFilter) ([]*domain.License, error) {
	return s.repo.List(ctx, filter)
}

// CheckExpiringLicenses finds licenses expiring within given days
func (s *LicensingService) CheckExpiringLicenses(ctx context.Context, withinDays int) ([]*domain.License, error) {
	return s.repo.GetExpiring(ctx, withinDays)
}

// GetLicenseStatistics returns license statistics
func (s *LicensingService) GetLicenseStatistics(ctx context.Context) (*port.LicenseStatistics, error) {
	return s.repo.GetStatistics(ctx)
}
