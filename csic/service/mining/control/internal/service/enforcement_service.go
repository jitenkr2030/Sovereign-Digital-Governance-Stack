package service

import (
	"context"
	"log"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
)

// EnforcementService handles compliance enforcement actions
type EnforcementService struct {
	poolRepo       MiningPoolRepository
	violationRepo  ViolationRepository
	complianceRepo ComplianceRepository
}

// NewEnforcementService creates a new enforcement service
func NewEnforcementService(poolRepo MiningPoolRepository, violationRepo ViolationRepository, complianceRepo ComplianceRepository) *EnforcementService {
	return &EnforcementService{
		poolRepo:       poolRepo,
		violationRepo:  violationRepo,
		complianceRepo: complianceRepo,
	}
}

// ResolveViolation resolves a compliance violation
func (s *EnforcementService) ResolveViolation(ctx context.Context, violationID uuid.UUID, resolution string, resolvedBy uuid.UUID) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violation",
			Err:     err,
		}
	}

	if violation == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Violation not found",
		}
	}

	now := time.Now()
	violation.Status = domain.ViolationStatusResolved
	violation.ResolvedAt = &now
	violation.ResolvedBy = &resolvedBy
	violation.Resolution = &resolution
	violation.UpdatedAt = now

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to update violation",
			Err:     err,
		}
	}

	log.Printf("Resolved violation %s: %s", violationID, resolution)

	return nil
}

// DismissViolation dismisses a violation
func (s *EnforcementService) DismissViolation(ctx context.Context, violationID uuid.UUID, reason string, dismissedBy uuid.UUID) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violation",
			Err:     err,
		}
	}

	if violation == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Violation not found",
		}
	}

	now := time.Now()
	violation.Status = domain.ViolationStatusDismissed
	violation.ResolvedAt = &now
	violation.ResolvedBy = &dismissedBy
	violation.Resolution = &reason
	violation.UpdatedAt = now

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to dismiss violation",
			Err:     err,
		}
	}

	log.Printf("Dismissed violation %s: %s", violationID, reason)

	return nil
}

// EscalateViolation escalates a violation to higher severity
func (s *EnforcementService) EscalateViolation(ctx context.Context, violationID uuid.UUID, reason string, escalatedTo uuid.UUID) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violation",
			Err:     err,
		}
	}

	if violation == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Violation not found",
		}
	}

	// Escalate severity
	switch violation.Severity {
	case domain.ViolationSeverityLow:
		violation.Severity = domain.ViolationSeverityMedium
	case domain.ViolationSeverityMedium:
		violation.Severity = domain.ViolationSeverityHigh
	case domain.ViolationSeverityHigh:
		violation.Severity = domain.ViolationSeverityCritical
	}

	violation.AssignedTo = &escalatedTo
	violation.Status = domain.ViolationStatusInvestigating
	violation.UpdatedAt = time.Now()

	// Add escalation note
	note := domain.ViolationNote{
		ID:         uuid.New(),
		ViolationID: violationID,
		AuthorID:   escalatedTo,
		Content:    "Violation escalated: " + reason,
		CreatedAt:  time.Now(),
		IsInternal: true,
	}
	violation.Notes = append(violation.Notes, note)

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to escalate violation",
			Err:     err,
		}
	}

	log.Printf("Escalated violation %s to %s", violationID, violation.Severity)

	return nil
}

// AssignViolation assigns a violation to an investigator
func (s *EnforcementService) AssignViolation(ctx context.Context, violationID uuid.UUID, assigneeID uuid.UUID) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violation",
			Err:     err,
		}
	}

	if violation == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Violation not found",
		}
	}

	violation.AssignedTo = &assigneeID
	violation.Status = domain.ViolationStatusInvestigating
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to assign violation",
			Err:     err,
		}
	}

	log.Printf("Assigned violation %s to %s", violationID, assigneeID)

	return nil
}

// AddViolationNote adds a note to a violation
func (s *EnforcementService) AddViolationNote(ctx context.Context, violationID uuid.UUID, note *domain.ViolationNote) error {
	violation, err := s.violationRepo.GetByID(ctx, violationID)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violation",
			Err:     err,
		}
	}

	if violation == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Violation not found",
		}
	}

	if note.ID == uuid.Nil {
		note.ID = uuid.New()
	}
	note.ViolationID = violationID
	note.CreatedAt = time.Now()

	violation.Notes = append(violation.Notes, *note)
	violation.UpdatedAt = time.Now()

	if err := s.violationRepo.Update(ctx, violation); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to add note to violation",
			Err:     err,
		}
	}

	return nil
}

// AutoSuspendPool automatically suspends a pool for critical violations
func (s *EnforcementService) AutoSuspendPool(ctx context.Context, poolID uuid.UUID, violationID uuid.UUID) error {
	pool, err := s.poolRepo.GetByID(ctx, poolID)
	if err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve mining pool",
			Err:     err,
		}
	}

	if pool == nil {
		return &ServiceError{
			Code:    "NOT_FOUND",
			Message: "Mining pool not found",
		}
	}

	if pool.Status == domain.PoolStatusSuspended {
		return nil // Already suspended
	}

	// Suspend the pool
	if err := s.poolRepo.UpdateStatus(ctx, poolID, domain.PoolStatusSuspended); err != nil {
		return &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to suspend mining pool",
			Err:     err,
		}
	}

	log.Printf("Auto-suspended mining pool %s (%s) due to violation %s", pool.Name, pool.LicenseNumber, violationID)

	return nil
}

// CheckAndEnforceCompliance checks compliance and takes enforcement action
func (s *EnforcementService) CheckAndEnforceCompliance(ctx context.Context, poolID uuid.UUID) ([]domain.ComplianceViolation, error) {
	violations, err := s.violationRepo.GetOpenViolations(ctx, poolID)
	if err != nil {
		return nil, &ServiceError{
			Code:    "DATABASE_ERROR",
			Message: "Failed to retrieve violations",
			Err:     err,
		}
	}

	// Check for critical violations requiring immediate action
	for _, v := range violations {
		if v.Severity == domain.ViolationSeverityCritical {
			// Auto-suspend pool
			if err := s.AutoSuspendPool(ctx, poolID, v.ID); err != nil {
				log.Printf("Failed to auto-suspend pool %s: %v", poolID, err)
			}
		}
	}

	return violations, nil
}

// ServiceError represents a service error
type ServiceError struct {
	Code    string
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}
