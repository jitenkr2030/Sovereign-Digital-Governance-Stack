// Compliance Management Module - Violation Service
// Business logic for violation tracking and enforcement

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/compliance/internal/domain"
	"github.com/csic-platform/compliance/internal/port"
)

// ViolationService handles compliance violation operations
type ViolationService struct {
	violationRepo port.ViolationRepository
	penaltyRepo   port.PenaltyRepository
	entityRepo    port.EntityRepository
	audit         port.AuditLogPort
}

// NewViolationService creates a new violation service
func NewViolationService(
	violationRepo port.ViolationRepository,
	penaltyRepo port.PenaltyRepository,
	entityRepo port.EntityRepository,
	audit port.AuditLogPort,
) *ViolationService {
	return &ViolationService{
		violationRepo: violationRepo,
		penaltyRepo:   penaltyRepo,
		entityRepo:    entityRepo,
		audit:         audit,
	}
}

// CreateViolation creates a new violation record
func (s *ViolationService) CreateViolation(ctx context.Context, violation *domain.ComplianceViolation, actorID string) error {
	if err := violation.Validate(); err != nil {
		return err
	}

	violation.Status = domain.ViolationStatusDetected
	violation.ViolationNumber = violation.GenerateViolationNumber()
	violation.DetectionDate = time.Now()
	violation.CreatedAt = time.Now()
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Create(ctx, violation); err != nil {
		return fmt.Errorf("failed to create violation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "VIOLATION_CREATED",
		ResourceType:  "VIOLATION",
		ResourceID:    violation.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Created violation %s: %s (Severity: %s)", violation.ViolationNumber, violation.Title, violation.Severity),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"violation_number": violation.ViolationNumber,
			"severity":         violation.Severity,
			"type":             violation.Type,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetViolation retrieves a violation by ID
func (s *ViolationService) GetViolation(ctx context.Context, violationID string) (*domain.ComplianceViolation, error) {
	return s.violationRepo.GetByID(ctx, violationID)
}

// ListViolations lists violations with filters
func (s *ViolationService) ListViolations(ctx context.Context, filter port.ViolationFilter) ([]*domain.ComplianceViolation, error) {
	return s.violationRepo.List(ctx, filter)
}

// UpdateViolation updates a violation
func (s *ViolationService) UpdateViolation(ctx context.Context, violation *domain.ComplianceViolation, actorID string) error {
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to update violation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "VIOLATION_UPDATED",
		ResourceType:  "VIOLATION",
		ResourceID:    violation.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Updated violation: %s", violation.ViolationNumber),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// AssignInvestigator assigns an investigator to a violation
func (s *ViolationService) AssignInvestigator(ctx context.Context, violationID, investigatorID, actorID string) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return err
	}

	violation.AssignedInvestigator = investigatorID
	violation.Status = domain.ViolationStatusInvestigating
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to assign investigator: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "VIOLATION_ASSIGNED",
		ResourceType:  "VIOLATION",
		ResourceID:    violation.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Assigned investigator to violation: %s", violation.ViolationNumber),
		Metadata: map[string]interface{}{
			"investigator_id": investigatorID,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// ConfirmViolation confirms a violation after investigation
func (s *ViolationService) ConfirmViolation(ctx context.Context, violationID, actorID string) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return err
	}

	if err := violation.Confirm(); err != nil {
		return err
	}

	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to confirm violation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "VIOLATION_CONFIRMED",
		ResourceType:  "VIOLATION",
		ResourceID:    violation.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Confirmed violation: %s", violation.ViolationNumber),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// IssuePenalty issues a penalty for a violation
func (s *ViolationService) IssuePenalty(ctx context.Context, violationID string, penalty *domain.Penalty, actorID string) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return err
	}

	// Create penalty
	penalty.PenaltyNumber = penalty.GeneratePenaltyNumber()
	penalty.IssuedDate = time.Now()
	penalty.Status = domain.PenaltyStatusIssued

	if err := s.penaltyRepo.Create(ctx, penalty); err != nil {
		return fmt.Errorf("failed to create penalty: %w", err)
	}

	// Add penalty to violation
	violation.AddPenalty(penalty.ID)
	violation.Status = domain.ViolationStatusPendingAction
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to update violation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "PENALTY_ISSUED",
		ResourceType:  "PENALTY",
		ResourceID:    penalty.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Issued penalty %s for violation %s (Amount: %s %.2f)", penalty.PenaltyNumber, violation.ViolationNumber, penalty.Currency, penalty.Amount),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"penalty_number": penalty.PenaltyNumber,
			"amount":         penalty.Amount,
			"type":           penalty.Type,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// ResolveViolation resolves a violation
func (s *ViolationService) ResolveViolation(ctx context.Context, violationID, correctiveAction, preventiveAction, actorID string) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return err
	}

	violation.CorrectiveAction = correctiveAction
	violation.PreventiveAction = preventiveAction

	if err := violation.Resolve(); err != nil {
		return err
	}

	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to resolve violation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "VIOLATION_RESOLVED",
		ResourceType:  "VIOLATION",
		ResourceID:    violation.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Resolved violation: %s", violation.ViolationNumber),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// CloseViolation closes a violation
func (s *ViolationService) CloseViolation(ctx context.Context, violationID, actorID string) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return err
	}

	if err := violation.Close(); err != nil {
		return err
	}

	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return fmt.Errorf("failed to close violation: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "VIOLATION_CLOSED",
		ResourceType:  "VIOLATION",
		ResourceID:    violation.ID,
		EntityID:      violation.EntityID,
		Description:   fmt.Sprintf("Closed violation: %s", violation.ViolationNumber),
		Result:        "SUCCESS",
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetOpenViolations retrieves all open violations for an entity
func (s *ViolationService) GetOpenViolations(ctx context.Context, entityID string) ([]*domain.ComplianceViolation, error) {
	return s.violationRepo.GetOpenByEntityID(ctx, entityID)
}

// GetViolationStatistics returns violation statistics
func (s *ViolationService) GetViolationStatistics(ctx context.Context, entityID string) (*port.ViolationStatistics, error) {
	return s.violationRepo.GetStatistics(ctx, entityID)
}

// GetPenalty retrieves a penalty by ID
func (s *ViolationService) GetPenalty(ctx context.Context, penaltyID string) (*domain.Penalty, error) {
	return s.penaltyRepo.GetByID(ctx, penaltyID)
}

// ListPenalty lists penalties with filters
func (s *ViolationService) ListPenalty(ctx context.Context, filter port.PenaltyFilter) ([]*domain.Penalty, error) {
	return s.penaltyRepo.List(ctx, filter)
}

// MarkPenaltyPaid marks a penalty as paid
func (s *ViolationService) MarkPenaltyPaid(ctx context.Context, penaltyID, paymentRef, actorID string) error {
	penalty, err := s.penaltyRepo.GetByID(ctx, penaltyID)
	if err != nil {
		return err
	}

	if err := penalty.MarkPaid(paymentRef); err != nil {
		return err
	}

	if err := s.penaltyRepo.Update(ctx, penalty); err != nil {
		return fmt.Errorf("failed to update penalty: %w", err)
	}

	// Audit log
	if err := s.audit.Log(ctx, &port.AuditEntry{
		Timestamp:     time.Now(),
		ActorID:       actorID,
		Action:        "PENALTY_PAID",
		ResourceType:  "PENALTY",
		ResourceID:    penalty.ID,
		EntityID:      penalty.ViolationID,
		Description:   fmt.Sprintf("Penalty paid: %s (Ref: %s)", penalty.PenaltyNumber, paymentRef),
		Result:        "SUCCESS",
		Metadata: map[string]interface{}{
			"payment_ref": paymentRef,
		},
	}); err != nil {
		return fmt.Errorf("failed to audit log: %w", err)
	}

	return nil
}

// GetOverduePenalties retrieves all overdue penalties
func (s *ViolationService) GetOverduePenalties(ctx context.Context) ([]*domain.Penalty, error) {
	return s.penaltyRepo.GetOverdue(ctx)
}
