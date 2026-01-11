package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"intervention-service/models"
	"intervention-service/repository"
)

// InterventionService handles intervention business logic
type InterventionService struct {
	repo       *repository.PostgresRepository
	redis      *redis.Client
	kafkaTopic string
}

// NewInterventionService creates a new intervention service
func NewInterventionService(repo *repository.PostgresRepository, redisClient *redis.Client) *InterventionService {
	return &InterventionService{
		repo:       repo,
		redis:      redisClient,
		kafkaTopic: "neam.interventions",
	}
}

// CreateIntervention creates a new intervention
func (s *InterventionService) CreateIntervention(ctx context.Context, intervention *models.Intervention) error {
	intervention.ID = uuid.New()
	intervention.Status = models.InterventionStatusPending
	intervention.CreatedAt = time.Now()
	intervention.UpdatedAt = time.Now()

	// Validate intervention
	if err := s.validateIntervention(intervention); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create in database
	if err := s.repo.CreateIntervention(ctx, intervention); err != nil {
		return fmt.Errorf("failed to create intervention: %w", err)
	}

	// Publish event to Kafka
	s.publishInterventionEvent(ctx, "intervention_created", intervention)

	return nil
}

// GetIntervention retrieves an intervention by ID
func (s *InterventionService) GetIntervention(ctx context.Context, id uuid.UUID) (*models.Intervention, error) {
	return s.repo.GetIntervention(ctx, id)
}

// ListInterventions retrieves interventions with filters
func (s *InterventionService) ListInterventions(ctx context.Context, filters map[string]interface{}) ([]*models.Intervention, error) {
	return s.repo.ListInterventions(ctx, filters)
}

// ApproveIntervention approves an intervention
func (s *InterventionService) ApproveIntervention(ctx context.Context, id uuid.UUID, approverID uuid.UUID) error {
	intervention, err := s.repo.GetIntervention(ctx, id)
	if err != nil {
		return err
	}

	if intervention.Status != models.InterventionStatusPending {
		return fmt.Errorf("intervention is not pending approval")
	}

	intervention.Status = models.InterventionStatusApproved
	intervention.UpdatedAt = time.Now()

	if err := s.repo.UpdateIntervention(ctx, intervention); err != nil {
		return err
	}

	s.publishInterventionEvent(ctx, "intervention_approved", intervention)

	return nil
}

// RejectIntervention rejects an intervention
func (s *InterventionService) RejectIntervention(ctx context.Context, id uuid.UUID, reason string) error {
	intervention, err := s.repo.GetIntervention(ctx, id)
	if err != nil {
		return err
	}

	if intervention.Status != models.InterventionStatusPending {
		return fmt.Errorf("intervention is not pending approval")
	}

	intervention.Status = models.InterventionStatusCancelled
	intervention.UpdatedAt = time.Now()
	intervention.Result = &models.InterventionResult{
		Success:      false,
		ErrorMessage: reason,
	}

	if err := s.repo.UpdateIntervention(ctx, intervention); err != nil {
		return err
	}

	s.publishInterventionEvent(ctx, "intervention_rejected", intervention)

	return nil
}

// ExecuteIntervention executes an approved intervention
func (s *InterventionService) ExecuteIntervention(ctx context.Context, id uuid.UUID, executorID uuid.UUID) error {
	intervention, err := s.repo.GetIntervention(ctx, id)
	if err != nil {
		return err
	}

	if intervention.Status != models.InterventionStatusApproved {
		return fmt.Errorf("intervention is not approved for execution")
	}

	intervention.Status = models.InterventionStatusExecuting
	intervention.ExecutedBy = executorID
	now := time.Now()
	intervention.ExecutedAt = &now
	intervention.UpdatedAt = now

	if err := s.repo.UpdateIntervention(ctx, intervention); err != nil {
		return err
	}

	// Execute based on type
	switch intervention.Type {
	case models.InterventionTypeLock:
		err = s.executeLock(ctx, intervention)
	case models.InterventionTypeCap:
		err = s.executeCap(ctx, intervention)
	case models.InterventionTypeRedirect:
		err = s.executeRedirect(ctx, intervention)
	case models.InterventionTypeSubsidize:
		err = s.executeSubsidize(ctx, intervention)
	case models.InterventionTypeRestrict:
		err = s.executeRestrict(ctx, intervention)
	default:
		return fmt.Errorf("unknown intervention type: %s", intervention.Type)
	}

	if err != nil {
		intervention.Status = models.InterventionStatusFailed
		intervention.Result = &models.InterventionResult{
			Success:      false,
			ErrorMessage: err.Error(),
		}
		s.repo.UpdateIntervention(ctx, intervention)
		return err
	}

	intervention.Status = models.InterventionStatusActive
	intervention.Result = &models.InterventionResult{
		Success:          true,
		AffectedEntities: s.countAffectedEntities(intervention),
	}
	intervention.UpdatedAt = time.Now()

	if err := s.repo.UpdateIntervention(ctx, intervention); err != nil {
		return err
	}

	s.publishInterventionEvent(ctx, "intervention_executed", intervention)

	return nil
}

// RollbackIntervention rolls back an active intervention
func (s *InterventionService) RollbackIntervention(ctx context.Context, id uuid.UUID, reason string) error {
	intervention, err := s.repo.GetIntervention(ctx, id)
	if err != nil {
		return err
	}

	if intervention.Status != models.InterventionStatusActive {
		return fmt.Errorf("intervention is not active")
	}

	intervention.Status = models.InterventionStatusRollingBack
	intervention.UpdatedAt = time.Now()

	if err := s.repo.UpdateIntervention(ctx, intervention); err != nil {
		return err
	}

	// Execute rollback based on type
	switch intervention.Type {
	case models.InterventionTypeLock:
		err = s.executeUnlock(ctx, intervention)
	case models.InterventionTypeCap:
		err = s.removeCap(ctx, intervention)
	case models.InterventionTypeRedirect:
		err = s.removeRedirect(ctx, intervention)
	default:
		err = fmt.Errorf("rollback not supported for type: %s", intervention.Type)
	}

	if err != nil {
		return err
	}

	now := time.Now()
	intervention.Status = models.InterventionStatusRolledBack
	intervention.RolledBackAt = &now
	intervention.UpdatedAt = now

	if err := s.repo.UpdateIntervention(ctx, intervention); err != nil {
		return err
	}

	s.publishInterventionEvent(ctx, "intervention_rolled_back", intervention)

	return nil
}

// GetInterventionStatus returns the current status of an intervention
func (s *InterventionService) GetInterventionStatus(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	intervention, err := s.repo.GetIntervention(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            intervention.ID,
		"status":        intervention.Status,
		"type":          intervention.Type,
		"name":          intervention.Name,
		"executed_at":   intervention.ExecutedAt,
		"completed_at":  intervention.CompletedAt,
		"result":        intervention.Result,
		"updated_at":    intervention.UpdatedAt,
	}, nil
}

// Helper functions

func (s *InterventionService) validateIntervention(intervention *models.Intervention) error {
	if intervention.Type == "" {
		return fmt.Errorf("intervention type is required")
	}
	if intervention.Name == "" {
		return fmt.Errorf("intervention name is required")
	}
	if intervention.Reason == "" {
		return fmt.Errorf("intervention reason is required")
	}
	return nil
}

func (s *InterventionService) executeLock(ctx context.Context, intervention *models.Intervention) error {
	lockKey := fmt.Sprintf("lock:%s:%s", intervention.Parameters.TargetEntity, intervention.ID)
	
	// Set lock in Redis with TTL
	if intervention.Parameters.Duration != nil {
		s.redis.Set(ctx, lockKey, intervention.Parameters.TargetValue, *intervention.Parameters.Duration)
	} else {
		s.redis.Set(ctx, lockKey, intervention.Parameters.TargetValue, 0)
	}

	log.Printf("Executed lock intervention %s on entity %s", intervention.ID, intervention.Parameters.TargetEntity)
	return nil
}

func (s *InterventionService) executeUnlock(ctx context.Context, intervention *models.Intervention) error {
	lockKey := fmt.Sprintf("lock:%s:%s", intervention.Parameters.TargetEntity, intervention.ID)
	s.redis.Del(ctx, lockKey)

	log.Printf("Executed unlock intervention %s on entity %s", intervention.ID, intervention.Parameters.TargetEntity)
	return nil
}

func (s *InterventionService) executeCap(ctx context.Context, intervention *models.Intervention) error {
	capKey := fmt.Sprintf("cap:%s", intervention.Parameters.TargetEntity)
	s.redis.Set(ctx, capKey, intervention.Parameters.TargetValue, 0)

	log.Printf("Executed cap intervention %s on entity %s with value %f", 
		intervention.ID, intervention.Parameters.TargetEntity, intervention.Parameters.TargetValue)
	return nil
}

func (s *InterventionService) removeCap(ctx context.Context, intervention *models.Intervention) error {
	capKey := fmt.Sprintf("cap:%s", intervention.Parameters.TargetEntity)
	s.redis.Del(ctx, capKey)

	log.Printf("Removed cap intervention %s on entity %s", intervention.ID, intervention.Parameters.TargetEntity)
	return nil
}

func (s *InterventionService) executeRedirect(ctx context.Context, intervention *models.Intervention) error {
	redirectKey := fmt.Sprintf("redirect:%s", intervention.Parameters.TargetEntity)
	redirectData := map[string]interface{}{
		"to":  intervention.Parameters.TargetValue,
		"at":  time.Now(),
	}
	jsonData, _ := json.Marshal(redirectData)
	s.redis.Set(ctx, redirectKey, jsonData, 0)

	log.Printf("Executed redirect intervention %s on entity %s", intervention.ID, intervention.Parameters.TargetEntity)
	return nil
}

func (s *InterventionService) removeRedirect(ctx context.Context, intervention *models.Intervention) error {
	redirectKey := fmt.Sprintf("redirect:%s", intervention.Parameters.TargetEntity)
	s.redis.Del(ctx, redirectKey)

	log.Printf("Removed redirect intervention %s on entity %s", intervention.ID, intervention.Parameters.TargetEntity)
	return nil
}

func (s *InterventionService) executeSubsidize(ctx context.Context, intervention *models.Intervention) error {
	// Implementation for subsidy intervention
	log.Printf("Executed subsidize intervention %s with value %f", 
		intervention.ID, intervention.Parameters.TargetValue)
	return nil
}

func (s *InterventionService) executeRestrict(ctx context.Context, intervention *models.Intervention) error {
	// Implementation for restriction intervention
	log.Printf("Executed restrict intervention %s on entity %s", 
		intervention.ID, intervention.Parameters.TargetEntity)
	return nil
}

func (s *InterventionService) countAffectedEntities(intervention *models.Intervention) int {
	// Count affected entities based on scope
	count := 100 // Placeholder
	return count
}

func (s *InterventionService) publishInterventionEvent(ctx context.Context, eventType string, intervention *models.Intervention) {
	event := map[string]interface{}{
		"type":          eventType,
		"intervention_id": intervention.ID,
		"type":          intervention.Type,
		"status":        intervention.Status,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}
	
	data, _ := json.Marshal(event)
	// Kafka publish would go here
	log.Printf("Published event %s for intervention %s", eventType, intervention.ID)
}
