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

// ObligationService implements the ObligationService interface
type ObligationService struct {
	repo ports.ObligationRepository
	log  *zap.Logger
}

// NewObligationService creates a new ObligationService instance
func NewObligationService(repo ports.ObligationRepository, log *zap.Logger) *ObligationService {
	return &ObligationService{
		repo: repo,
		log:  log,
	}
}

// CreateObligation creates a new regulatory obligation
func (s *ObligationService) CreateObligation(ctx context.Context, req ports.CreateObligationRequest) (*domain.Obligation, error) {
	s.log.Info("Creating obligation",
		zap.String("entity_id", req.EntityID.String()),
		zap.String("description", req.Description),
	)

	dueDate := time.Now().AddDate(0, 1, 0) // Default 1 month
	if req.DueDate != "" {
		if parsed, err := time.Parse(time.RFC3339, req.DueDate); err == nil {
			dueDate = parsed
		}
	}

	priority := req.Priority
	if priority < 1 {
		priority = 1 // Default priority
	}

	obligation := &domain.Obligation{
		ID:           uuid.New(),
		EntityID:     req.EntityID,
		RegulationID: req.RegulationID,
		Description:  req.Description,
		DueDate:      dueDate,
		Status:       domain.ObligationPending,
		Priority:     priority,
		EvidenceRefs: req.EvidenceRefs,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateObligation(ctx, obligation); err != nil {
		return nil, fmt.Errorf("failed to create obligation: %w", err)
	}

	s.log.Info("Obligation created", zap.String("obligation_id", obligation.ID.String()))
	return obligation, nil
}

// GetObligation retrieves an obligation by ID
func (s *ObligationService) GetObligation(ctx context.Context, obligationID uuid.UUID) (*domain.Obligation, error) {
	return s.repo.GetObligation(ctx, obligationID)
}

// GetEntityObligations retrieves all obligations for an entity
func (s *ObligationService) GetEntityObligations(ctx context.Context, entityID uuid.UUID) ([]domain.Obligation, error) {
	return s.repo.GetObligationsByEntity(ctx, entityID)
}

// GetObligationsByStatus retrieves obligations by status
func (s *ObligationService) GetObligationsByStatus(ctx context.Context, status domain.ObligationStatus) ([]domain.Obligation, error) {
	return s.repo.GetObligationsByStatus(ctx, status)
}

// FulfillObligation marks an obligation as fulfilled
func (s *ObligationService) FulfillObligation(ctx context.Context, obligationID uuid.UUID, evidence string) error {
	s.log.Info("Fulfilling obligation", zap.String("obligation_id", obligationID.String()))

	obligation, err := s.repo.GetObligation(ctx, obligationID)
	if err != nil {
		return fmt.Errorf("failed to get obligation: %w", err)
	}
	if obligation == nil {
		return fmt.Errorf("obligation not found: %s", obligationID.String())
	}

	if obligation.Status == domain.ObligationFulfilled {
		return fmt.Errorf("obligation is already fulfilled")
	}

	now := time.Now().UTC()

	// Check for early fulfillment bonus (if due date is more than 7 days away)
	var bonusMessage string
	if obligation.DueDate.After(now.AddDate(0, 0, 7)) {
		bonusMessage = fmt.Sprintf("Early fulfillment bonus applied (+%.1f points)", 5.0)
	}

	return s.repo.MarkObligationFulfilled(ctx, obligationID, evidence+"; "+bonusMessage)
}

// WaiveObligation waives an obligation
func (s *ObligationService) WaiveObligation(ctx context.Context, obligationID uuid.UUID, reason string) error {
	s.log.Info("Waiving obligation", zap.String("obligation_id", obligationID.String()))

	obligation, err := s.repo.GetObligation(ctx, obligationID)
	if err != nil {
		return fmt.Errorf("failed to get obligation: %w", err)
	}
	if obligation == nil {
		return fmt.Errorf("obligation not found: %s", obligationID.String())
	}

	return s.repo.UpdateObligationStatus(ctx, obligationID, domain.ObligationWaived)
}

// GetOverdueObligations retrieves all overdue obligations
func (s *ObligationService) GetOverdueObligations(ctx context.Context) ([]domain.Obligation, error) {
	return s.repo.GetOverdueObligations(ctx)
}

// GetUpcomingObligations retrieves obligations due within specified days
func (s *ObligationService) GetUpcomingObligations(ctx context.Context, days int) ([]domain.Obligation, error) {
	return s.repo.GetUpcomingObligations(ctx, days)
}

// CheckAndUpdateOverdueObligations checks for and updates overdue obligations
func (s *ObligationService) CheckAndUpdateOverdueObligations(ctx context.Context) error {
	s.log.Info("Checking for overdue obligations")

	overdue, err := s.repo.GetOverdueObligations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get overdue obligations: %w", err)
	}

	now := time.Now().UTC()
	updatedCount := 0

	for _, obs := range overdue {
		if obs.Status == domain.ObligationPending && obs.DueDate.Before(now) {
			if err := s.repo.UpdateObligationStatus(ctx, obs.ID, domain.ObligationOverdue); err != nil {
				s.log.Error("Failed to update obligation status",
					zap.String("obligation_id", obs.ID.String()),
					zap.Error(err),
				)
			} else {
				updatedCount++
			}
		}
	}

	s.log.Info("Overdue obligation check completed", zap.Int("updated", updatedCount))
	return nil
}
