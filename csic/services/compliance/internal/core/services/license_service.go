package services

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/services/compliance/internal/core/domain"
	"github.com/csic-platform/services/services/compliance/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LicenseService implements the LicenseService interface
type LicenseService struct {
	repo ports.LicenseRepository
	log  *zap.Logger
}

// NewLicenseService creates a new LicenseService instance
func NewLicenseService(repo ports.LicenseRepository, log *zap.Logger) *LicenseService {
	return &LicenseService{
		repo: repo,
		log:  log,
	}
}

// SubmitApplication submits a new license application
func (s *LicenseService) SubmitApplication(ctx context.Context, req ports.SubmitApplicationRequest) (*domain.LicenseApplication, error) {
	s.log.Info("Submitting license application",
		zap.String("entity_id", req.EntityID.String()),
		zap.String("type", string(req.Type)),
	)

	// Verify entity exists
	entity, err := s.repo.GetEntity(ctx, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify entity: %w", err)
	}
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", req.EntityID.String())
	}

	now := time.Now().UTC()
	app := &domain.LicenseApplication{
		ID:             uuid.New(),
		EntityID:       req.EntityID,
		Type:           req.Type,
		LicenseType:    req.LicenseType,
		Status:         domain.AppStatusSubmitted,
		SubmittedAt:    now,
		RequestedTerms: req.RequestedTerms,
		Documents:      req.Documents,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	s.log.Info("License application submitted", zap.String("app_id", app.ID.String()))
	return app, nil
}

// ReviewApplication reviews and approves/rejects a license application
func (s *LicenseService) ReviewApplication(ctx context.Context, appID uuid.UUID, decision ports.ApplicationReviewRequest) (*domain.License, error) {
	s.log.Info("Reviewing license application",
		zap.String("app_id", appID.String()),
		zap.Bool("approved", decision.Approved),
	)

	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("application not found: %s", appID.String())
	}

	if app.Status != domain.AppStatusSubmitted && app.Status != domain.AppStatusUnderReview {
		return nil, fmt.Errorf("application cannot be reviewed in current status: %s", app.Status)
	}

	now := time.Now().UTC()
	app.Status = domain.AppStatusUnderReview
	app.ReviewedAt = &now
	app.ReviewerID = &decision.ReviewerID
	app.ReviewerNotes = decision.Notes

	var license *domain.License

	if decision.Approved {
		app.Status = domain.AppStatusApproved
		s.log.Info("Application approved", zap.String("app_id", appID.String()))

		// Issue the license
		expiryDate := now.AddDate(1, 0, 0) // Default 1 year
		license = &domain.License{
			ID:            uuid.New(),
			EntityID:      app.EntityID,
			Type:          app.LicenseType,
			Status:        domain.LicenseStatusActive,
			LicenseNumber: generateLicenseNumber(),
			IssuedDate:    now,
			ExpiryDate:    expiryDate,
			Conditions:    decision.Conditions,
			Jurisdiction:  "",
			IssuedBy:      decision.ReviewerID.String(),
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.repo.CreateLicense(ctx, license); err != nil {
			return nil, fmt.Errorf("failed to create license: %w", err)
		}

		app.GrantedTerms = decision.GrantedTerms
		s.log.Info("License issued", zap.String("license_id", license.ID.String()))
	} else {
		app.Status = domain.AppStatusRejected
		s.log.Info("Application rejected", zap.String("app_id", appID.String()))
	}

	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return license, nil
}

// WithdrawApplication withdraws a license application
func (s *LicenseService) WithdrawApplication(ctx context.Context, appID uuid.UUID) error {
	s.log.Info("Withdrawing application", zap.String("app_id", appID.String()))

	app, err := s.repo.GetApplication(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return fmt.Errorf("application not found: %s", appID.String())
	}

	if app.Status != domain.AppStatusSubmitted {
		return fmt.Errorf("application cannot be withdrawn in current status: %s", app.Status)
	}

	app.Status = domain.AppStatusWithdrawn
	return s.repo.UpdateApplication(ctx, app)
}

// GetApplication retrieves a license application
func (s *LicenseService) GetApplication(ctx context.Context, appID uuid.UUID) (*domain.LicenseApplication, error) {
	return s.repo.GetApplication(ctx, appID)
}

// ListApplications lists license applications with filters
func (s *LicenseService) ListApplications(ctx context.Context, filter ports.ApplicationFilter) ([]domain.LicenseApplication, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	offset := (filter.Page - 1) * filter.PageSize

	var apps []domain.LicenseApplication

	if filter.Status != nil {
		apps, err := s.repo.GetApplicationsByStatus(ctx, *filter.Status)
		if err != nil {
			return nil, err
		}
		return filterAndPaginateApps(apps, filter, offset), nil
	}

	if filter.EntityID != nil {
		return s.repo.GetApplicationsByEntity(ctx, *filter.EntityID)
	}

	// Return empty for unhandled filters
	return []domain.LicenseApplication{}, nil
}

// IssueLicense issues a new license directly (bypassing application)
func (s *LicenseService) IssueLicense(ctx context.Context, req ports.IssueLicenseRequest) (*domain.License, error) {
	s.log.Info("Issuing license",
		zap.String("entity_id", req.EntityID.String()),
		zap.String("type", string(req.LicenseType)),
	)

	entity, err := s.repo.GetEntity(ctx, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify entity: %w", err)
	}
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", req.EntityID.String())
	}

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, req.ExpiryDays)
	if req.ExpiryDays <= 0 {
		expiryDate = now.AddDate(1, 0, 0) // Default 1 year
	}

	license := &domain.License{
		ID:            uuid.New(),
		EntityID:      req.EntityID,
		Type:          req.LicenseType,
		Status:        domain.LicenseStatusActive,
		LicenseNumber: req.LicenseNumber,
		IssuedDate:    now,
		ExpiryDate:    expiryDate,
		Conditions:    req.Conditions,
		Restrictions:  req.Restrictions,
		Jurisdiction:  req.Jurisdiction,
		IssuedBy:      req.IssuedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateLicense(ctx, license); err != nil {
		return nil, fmt.Errorf("failed to create license: %w", err)
	}

	s.log.Info("License issued", zap.String("license_id", license.ID.String()))
	return license, nil
}

// SuspendLicense suspends an active license
func (s *LicenseService) SuspendLicense(ctx context.Context, licenseID uuid.UUID, reason string) error {
	s.log.Info("Suspending license", zap.String("license_id", licenseID.String()))

	license, err := s.repo.GetLicense(ctx, licenseID)
	if err != nil {
		return fmt.Errorf("failed to get license: %w", err)
	}
	if license == nil {
		return fmt.Errorf("license not found: %s", licenseID.String())
	}

	if license.Status != domain.LicenseStatusActive {
		return fmt.Errorf("license cannot be suspended in current status: %s", license.Status)
	}

	return s.repo.UpdateLicenseStatus(ctx, licenseID, domain.LicenseStatusSuspended)
}

// RevokeLicense revokes a license
func (s *LicenseService) RevokeLicense(ctx context.Context, licenseID uuid.UUID, reason string) error {
	s.log.Info("Revoking license", zap.String("license_id", licenseID.String()))

	license, err := s.repo.GetLicense(ctx, licenseID)
	if err != nil {
		return fmt.Errorf("failed to get license: %w", err)
	}
	if license == nil {
		return fmt.Errorf("license not found: %s", licenseID.String())
	}

	now := time.Now().UTC()
	license.Status = domain.LicenseStatusRevoked
	license.RevokedAt = &now
	license.RevocationReason = reason

	return s.repo.UpdateLicense(ctx, license)
}

// RenewLicense renews an existing license
func (s *LicenseService) RenewLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error) {
	s.log.Info("Renewing license", zap.String("license_id", licenseID.String()))

	license, err := s.repo.GetLicense(ctx, licenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get license: %w", err)
	}
	if license == nil {
		return nil, fmt.Errorf("license not found: %s", licenseID.String())
	}

	now := time.Now().UTC()
	newExpiry := now.AddDate(1, 0, 0) // Default 1 year renewal

	// Create a new license record for the renewal
	newLicense := &domain.License{
		ID:            uuid.New(),
		EntityID:      license.EntityID,
		Type:          license.Type,
		Status:        domain.LicenseStatusActive,
		LicenseNumber: license.LicenseNumber + "-R" + now.Format("2006"),
		IssuedDate:    now,
		ExpiryDate:    newExpiry,
		Conditions:    license.Conditions,
		Restrictions:  license.Restrictions,
		Jurisdiction:  license.Jurisdiction,
		IssuedBy:      "SYSTEM",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateLicense(ctx, newLicense); err != nil {
		return nil, fmt.Errorf("failed to create renewal license: %w", err)
	}

	// Update old license to expired
	oldExpiry := now.AddDate(0, 0, -1) // Yesterday
	license.Status = domain.LicenseStatusExpired
	license.ExpiryDate = oldExpiry
	if err := s.repo.UpdateLicense(ctx, license); err != nil {
		s.log.Error("Failed to update old license status", zap.Error(err))
	}

	s.log.Info("License renewed", zap.String("old_id", licenseID.String()), zap.String("new_id", newLicense.ID.String()))
	return newLicense, nil
}

// GetLicense retrieves a license by ID
func (s *LicenseService) GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error) {
	return s.repo.GetLicense(ctx, licenseID)
}

// GetEntityLicenses retrieves all licenses for an entity
func (s *LicenseService) GetEntityLicenses(ctx context.Context, entityID uuid.UUID) ([]domain.License, error) {
	return s.repo.GetLicensesByEntity(ctx, entityID)
}

// GetExpiringLicenses retrieves licenses expiring within specified days
func (s *LicenseService) GetExpiringLicenses(ctx context.Context, days int) ([]domain.License, error) {
	return s.repo.GetLicensesExpiringSoon(ctx, days)
}

// RegisterEntity registers a new entity
func (s *LicenseService) RegisterEntity(ctx context.Context, req ports.RegisterEntityRequest) (*domain.Entity, error) {
	s.log.Info("Registering entity",
		zap.String("name", req.Name),
		zap.String("reg_num", req.RegistrationNum),
	)

	// Check for duplicate registration number
	existing, err := s.repo.GetEntityByRegistration(ctx, req.RegistrationNum)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing registration: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("entity with registration number already exists: %s", req.RegistrationNum)
	}

	now := time.Now().UTC()
	riskLevel := req.RiskLevel
	if riskLevel == "" {
		riskLevel = "MEDIUM"
	}

	entity := &domain.Entity{
		ID:              uuid.New(),
		Name:            req.Name,
		LegalName:       req.LegalName,
		RegistrationNum: req.RegistrationNum,
		Jurisdiction:    req.Jurisdiction,
		EntityType:      req.EntityType,
		Address:         req.Address,
		ContactEmail:    req.ContactEmail,
		Status:          "ACTIVE",
		RiskLevel:       riskLevel,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateEntity(ctx, entity); err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	s.log.Info("Entity registered", zap.String("entity_id", entity.ID.String()))
	return entity, nil
}

// GetEntity retrieves an entity by ID
func (s *LicenseService) GetEntity(ctx context.Context, entityID uuid.UUID) (*domain.Entity, error) {
	return s.repo.GetEntity(ctx, entityID)
}

// UpdateEntityRiskLevel updates an entity's risk level
func (s *LicenseService) UpdateEntityRiskLevel(ctx context.Context, entityID uuid.UUID, riskLevel string) error {
	entity, err := s.repo.GetEntity(ctx, entityID)
	if err != nil {
		return fmt.Errorf("failed to get entity: %w", err)
	}
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityID.String())
	}

	entity.RiskLevel = riskLevel
	entity.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateEntity(ctx, entity)
}

// CreateRegulation creates a new regulation
func (s *LicenseService) CreateRegulation(ctx context.Context, req ports.CreateRegulationRequest) (*domain.Regulation, error) {
	s.log.Info("Creating regulation", zap.String("title", req.Title))

	now := time.Now().UTC()
	effectiveDate := now
	if req.EffectiveDate != "" {
		if parsed, err := time.Parse(time.RFC3339, req.EffectiveDate); err == nil {
			effectiveDate = parsed
		}
	}

	reg := &domain.Regulation{
		ID:             uuid.New(),
		Title:          req.Title,
		Description:    req.Description,
		Category:       req.Category,
		Jurisdiction:   req.Jurisdiction,
		EffectiveDate:  effectiveDate,
		ParentID:       req.ParentID,
		Requirements:   req.Requirements,
		PenaltyDetails: req.PenaltyDetails,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateRegulation(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to create regulation: %w", err)
	}

	s.log.Info("Regulation created", zap.String("reg_id", reg.ID.String()))
	return reg, nil
}

// GetRegulation retrieves a regulation by ID
func (s *LicenseService) GetRegulation(ctx context.Context, regID uuid.UUID) (*domain.Regulation, error) {
	return s.repo.GetRegulation(ctx, regID)
}

// Helper functions

func generateLicenseNumber() string {
	// Generate a unique license number: LCC-YYYY-NNNNNN
	return fmt.Sprintf("LCC-%d-%06d", time.Now().Year(), uuid.New().ID()%1000000)
}

func filterAndPaginateApps(apps []domain.LicenseApplication, filter ports.ApplicationFilter, offset int) []domain.LicenseApplication {
	start := offset
	end := offset + filter.PageSize
	if start > len(apps) {
		return []domain.LicenseApplication{}
	}
	if end > len(apps) {
		end = len(apps)
	}
	return apps[start:end]
}
