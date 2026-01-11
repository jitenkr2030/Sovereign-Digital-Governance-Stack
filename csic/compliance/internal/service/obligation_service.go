// Compliance Management Module - Obligation Service
// Business logic for compliance obligation tracking

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
	"github.com/csic-platform/compliance/internal/port"
)

// ObligationService handles compliance obligation operations
type ObligationService struct {
	repo   port.ObligationRepository
	audit  port.AuditLogPort
}

// NewObligationService creates a new obligation service
func NewObligationService(repo port.ObligationRepository, audit port.AuditLogPort) *ObligationService {
	return &ObligationService{
		repo:  repo,
		audit: audit,
	}
}

// CreateObligation creates a new compliance obligation
func (s *ObligationService) CreateObligation(ctx context.Context, obligation *domain.ComplianceObligation, actorID string) error {
	if err := obligation.Validate(); err != nil {
		return err
	}

	obligation.Status = domain.ObligationStatusPending
	obligation.CreatedAt = time.Now()

	if err := s.repo.Create(ctx, obligation); err != nil {
		return fmt.Errorf("failed to create obligation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "OBLIGATION_CREATED",
		ResourceType:  "OBLIGATION",
		ResourceID:    obligation.ID,
		EntityID:      obligation.EntityID,
		Description:   fmt.Sprintf("Created obligation: %s (Due: %s)", obligation.Title, obligation.DueDate.Format("2006-01-02")),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetObligation retrieves an obligation by ID
func (s *ObligationService) GetObligation(ctx context.Context, obligationID string) (*domain.ComplianceObligation, error) {
	return s.repo.GetByID(ctx, obligationID)
}

// ListObligations lists obligations with filters
func (s *ObligationService) ListObligations(ctx context.Context, filter port.ObligationFilter) ([]*domain.ComplianceObligation, error) {
	return s.repo.List(ctx, filter)
}

// UpdateObligation updates an obligation
func (s *ObligationService) UpdateObligation(ctx context.Context, obligation *domain.ComplianceObligation, actorID string) error {
	if err := s.repo.Update(ctx, obligation); err != nil {
		return fmt.Errorf("failed to update obligation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "OBLIGATION_UPDATED",
		ResourceType:  "OBLIGATION",
		ResourceID:    obligation.ID,
		EntityID:      obligation.EntityID,
		Description:   fmt.Sprintf("Updated obligation: %s", obligation.Title),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// CompleteObligation marks an obligation as completed
func (s *ObligationService) CompleteObligation(ctx context.Context, obligationID string, actorID string) error {
	obligation, err := s.repo.GetByID(ctx, obligationID)
	if err != nil {
		return err
	}

	if err := obligation.Complete(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, obligation); err != nil {
		return fmt.Errorf("failed to complete obligation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "OBLIGATION_COMPLETED",
		ResourceType:  "OBLIGATION",
		ResourceID:    obligation.ID,
		EntityID:      obligation.EntityID,
		Description:   fmt.Sprintf("Completed obligation: %s", obligation.Title),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// VerifyObligation verifies a submitted obligation
func (s *ObligationService) VerifyObligation(ctx context.Context, obligationID string, notes string, actorID string) error {
	obligation, err := s.repo.GetByID(ctx, obligationID)
	if err != nil {
		return err
	}

	if err := obligation.Verify(notes); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, obligation); err != nil {
		return fmt.Errorf("failed to verify obligation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "OBLIGATION_VERIFIED",
		ResourceType:  "OBLIGATION",
		ResourceID:    obligation.ID,
		EntityID:      obligation.EntityID,
		Description:   fmt.Sprintf("Verified obligation: %s", obligation.Title),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"verification_notes": notes,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// ApproveObligation approves a verified obligation
func (s *ObligationService) ApproveObligation(ctx context.Context, obligationID string, notes string, actorID string) error {
	obligation, err := s.repo.GetByID(ctx, obligationID)
	if err != nil {
		return err
	}

	if err := obligation.Approve(notes); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, obligation); err != nil {
		return fmt.Errorf("failed to approve obligation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "OBLIGATION_APPROVED",
		ResourceType:  "OBLIGATION",
		ResourceID:    obligation.ID,
		EntityID:      obligation.EntityID,
		Description:   fmt.Sprintf("Approved obligation: %s", obligation.Title),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// RejectObligation rejects a submitted obligation
func (s *ObligationService) RejectObligation(ctx context.Context, obligationID string, notes string, actorID string) error {
	obligation, err := s.repo.GetByID(ctx, obligationID)
	if err != nil {
		return err
	}

	if err := obligation.Reject(notes); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, obligation); err != nil {
		return fmt.Errorf("failed to reject obligation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "OBLIGATION_REJECTED",
		ResourceType:  "OBLIGATION",
		ResourceID:    obligation.ID,
		EntityID:      obligation.EntityID,
		Description:   fmt.Sprintf("Rejected obligation: %s - Notes: %s", obligation.Title, notes),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// CheckAndMarkOverdueObligations finds and marks overdue obligations
func (s *ObligationService) CheckAndMarkOverdueObligations(ctx context.Context) ([]*domain.ComplianceObligation, error) {
	overdue, err := s.repo.GetOverdue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue obligations: %w", err)
	}

	for _, obligation := range overdue {
		if obligation.Status != domain.ObligationStatusOverdue {
			if err := obligation.MarkOverdue(); err != nil {
				continue
			}
			if err := s.repo.Update(ctx, obligation); err != nil {
				continue
			}
		}
	}

	return overdue, nil
}

// GetOverdueObligations retrieves all overdue obligations
func (s *ObligationService) GetOverdueObligations(ctx context.Context) ([]*domain.ComplianceObligation, error) {
	return s.repo.GetOverdue(ctx)
}

// GetPendingObligations retrieves pending obligations for an entity
func (s *ObligationService) GetPendingObligations(ctx context.Context, entityID string) ([]*domain.ComplianceObligation, error) {
	return s.repo.GetByEntityID(ctx, entityID, domain.ObligationStatusPending)
}

// GetUpcomingObligations retrieves obligations due within given days
func (s *ObligationService) GetUpcomingObligations(ctx context.Context, entityID string, withinDays int) ([]*domain.ComplianceObligation, error) {
	return s.repo.GetDueWithin(ctx, entityID, withinDays)
}

// GetObligationStatistics returns obligation statistics
func (s *ObligationService) GetObligationStatistics(ctx context.Context, entityID string) (*port.ObligationStatistics, error) {
	return s.repo.GetStatistics(ctx, entityID)
}

// CreateObligationFromTemplate creates an obligation from a template
func (s *ObligationService) CreateObligationFromTemplate(ctx context.Context, template domain.ObligationTemplate, entityID, licenseID string, actorID string) (*domain.CompligationObligation, error) {
	obligation := &domain.ComplianceObligation{
		LicenseID:       licenseID,
		EntityID:        entityID,
		Type:            template.Type,
		Category:        template.Category,
		Title:           template.Title,
		Description:     template.Description,
		ComplianceRef:   template.ComplianceRef,
		RegulatoryBody:  template.RegulatoryBody,
		Priority:        template.Priority,
		Status:          domain.ObligationStatusPending,
		AssignedTo:      actorID,
		DueDate:         time.Now().AddDate(0, 0, template.DueInDays),
		Frequency:       template.Frequency,
		Recurring:       template.Recurring,
	}

	if template.Recurring {
		obligation.SetNextOccurrence(template.Frequency)
	}

	if err := s.CreateObligation(ctx, obligation, actorID); err != nil {
		return nil, err
	}

	return obligation, nil
}
