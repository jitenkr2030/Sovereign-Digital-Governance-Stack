// Compliance Management Module - Entity Service
// Business logic for regulated entity management

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
	"github.com/csic-platform/compliance/internal/port"
)

// EntityService handles regulated entity operations
type EntityService struct {
	repo      port.EntityRepository
	audit     port.AuditLogPort
}

// NewEntityService creates a new entity service
func NewEntityService(repo port.EntityRepository, audit port.AuditLogPort) *EntityService {
	return &EntityService{
		repo:  repo,
		audit: audit,
	}
}

// CreateEntity creates a new regulated entity
func (s *EntityService) CreateEntity(ctx context.Context, entity *domain.RegulatedEntity, actorID string) error {
	// Validate entity data
	if err := entity.Validate(); err != nil {
		return err
	}

	// Check for duplicate registration number
	existing, err := s.repo.GetByRegistrationNumber(ctx, entity.RegistrationNumber)
	if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
		return err
	}
	if existing != nil {
		return domain.ErrEntityAlreadyExists
	}

	// Set defaults
	entity.Status = domain.EntityStatusPending
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	// Create entity
	if err := s.repo.Create(ctx, entity); err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "ENTITY_CREATED",
		ResourceType:  "ENTITY",
		ResourceID:    entity.ID,
		EntityID:      entity.ID,
		Description:   fmt.Sprintf("Created new entity: %s (%s)", entity.Name, entity.Type),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetEntity retrieves an entity by ID
func (s *EntityService) GetEntity(ctx context.Context, entityID string) (*domain.RegulatedEntity, error) {
	entity, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// GetEntityByRegistration retrieves an entity by registration number
func (s *EntityService) GetEntityByRegistration(ctx context.Context, regNumber string) (*domain.RegulatedEntity, error) {
	return s.repo.GetByRegistrationNumber(ctx, regNumber)
}

// ListEntities lists entities with filters
func (s *EntityService) ListEntities(ctx context.Context, filter port.EntityFilter) ([]*domain.RegulatedEntity, error) {
	return s.repo.List(ctx, filter)
}

// UpdateEntity updates an entity
func (s *EntityService) UpdateEntity(ctx context.Context, entity *domain.RegulatedEntity, actorID string) error {
	// Validate entity data
	if err := entity.Validate(); err != nil {
		return err
	}

	entity.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "ENTITY_UPDATED",
		ResourceType:  "ENTITY",
		ResourceID:    entity.ID,
		EntityID:      entity.ID,
		Description:   fmt.Sprintf("Updated entity: %s", entity.Name),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// ActivateEntity activates an entity
func (s *EntityService) ActivateEntity(ctx context.Context, entityID string, actorID string) error {
	entity, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return err
	}

	if err := entity.Activate(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to activate entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "ENTITY_ACTIVATED",
		ResourceType:  "ENTITY",
		ResourceID:    entity.ID,
		EntityID:      entity.ID,
		Description:   fmt.Sprintf("Activated entity: %s", entity.Name),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// SuspendEntity suspends an entity
func (s *EntityService) SuspendEntity(ctx context.Context, entityID string, reason string, actorID string) error {
	entity, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return err
	}

	if err := entity.Suspend(reason); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to suspend entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "ENTITY_SUSPENDED",
		ResourceType:  "ENTITY",
		ResourceID:    entity.ID,
		EntityID:      entity.ID,
		Description:   fmt.Sprintf("Suspended entity: %s - Reason: %s", entity.Name, reason),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// RevokeEntity revokes an entity
func (s *EntityService) RevokeEntity(ctx context.Context, entityID string, reason string, actorID string) error {
	entity, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return err
	}

	if err := entity.Revoke(reason); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to revoke entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "ENTITY_REVOKED",
		ResourceType:  "ENTITY",
		ResourceID:    entity.ID,
		EntityID:      entity.ID,
		Description:   fmt.Sprintf("Revoked entity: %s - Reason: %s", entity.Name, reason),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"reason": reason,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// AddLicenseToEntity adds a license ID to an entity
func (s *EntityService) AddLicenseToEntity(ctx context.Context, entityID, licenseID string, actorID string) error {
	entity, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return err
	}

	entity.AddLicense(licenseID)
	entity.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, entity); err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "LICENSE_ADDED_TO_ENTITY",
		ResourceType:  "ENTITY",
		ResourceID:    entity.ID,
		EntityID:      entity.ID,
		Description:   fmt.Sprintf("Added license %s to entity: %s", licenseID, entity.Name),
		Metadata: map[string]interface{}{
			"license_id": licenseID,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetEntityComplianceReport generates a compliance report for an entity
func (s *EntityService) GetEntityComplianceReport(ctx context.Context, entityID string, startDate, endDate time.Time) (*port.ComplianceReport, error) {
	entity, err := s.repo.GetByID(ctx, entityID)
	if err != nil {
		return nil, err
	}

	report := &port.ComplianceReport{
		EntityID:      entityID,
		EntityName:    entity.Name,
		ReportPeriod:  port.ReportPeriod{Start: startDate, End: endDate},
		GeneratedAt:   time.Now(),
	}

	return report, nil
}
