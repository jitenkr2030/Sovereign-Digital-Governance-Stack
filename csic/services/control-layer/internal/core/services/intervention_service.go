package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/pkg/metrics"
)

// InterventionService manages interventions
type InterventionService interface {
	ListInterventions(ctx context.Context) ([]*domain.Intervention, error)
	GetIntervention(ctx context.Context, id string) (*domain.Intervention, error)
	CreateIntervention(ctx context.Context, req *domain.CreateInterventionRequest) (*domain.Intervention, error)
	UpdateStatus(ctx context.Context, id string, status domain.InterventionStatus) error
	Resolve(ctx context.Context, id string, resolution string) error
	StartInterventionMonitor(logger *zap.Logger)
	TriggerIntervention(ctx context.Context, policyID uuid.UUID, target, reason string, severity string, data map[string]interface{}) (*domain.Intervention, error)
}

// InterventionServiceService implements the InterventionService interface
type InterventionServiceService struct {
	repositories  ports.Repositories
	messagingPort ports.MessagingPort
	policyEngine  PolicyEngine
	logger        *zap.Logger
	metrics       *metrics.MetricsCollector
	stopCh        chan struct{}
}

// NewInterventionService creates a new intervention service
func NewInterventionService(
	repositories ports.Repositories,
	messagingPort ports.MessagingPort,
	logger *zap.Logger,
	metricsCollector *metrics.MetricsCollector,
	policyEngine PolicyEngine,
) InterventionService {
	return &InterventionServiceService{
		repositories:  repositories,
		messagingPort: messagingPort,
		policyEngine:  policyEngine,
		logger:        logger,
		metrics:       metricsCollector,
		stopCh:        make(chan struct{}),
	}
}

// ListInterventions lists all interventions
func (s *InterventionServiceService) ListInterventions(ctx context.Context) ([]*domain.Intervention, error) {
	return s.repositories.InterventionRepository.GetActiveInterventions(ctx)
}

// GetIntervention gets an intervention by ID
func (s *InterventionServiceService) GetIntervention(ctx context.Context, id string) (*domain.Intervention, error) {
	interventionUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid intervention ID: %w", err)
	}

	return s.repositories.InterventionRepository.GetInterventionByID(ctx, interventionUUID)
}

// CreateIntervention creates a new intervention
func (s *InterventionServiceService) CreateIntervention(ctx context.Context, req *domain.CreateInterventionRequest) (*domain.Intervention, error) {
	start := time.Now()

	now := time.Now()
	intervention := &domain.Intervention{
		ID:              uuid.New(),
		PolicyID:        req.PolicyID,
		EnforcementID:   req.EnforcementID,
		TargetService:   req.TargetService,
		InterventionType: req.InterventionType,
		Severity:        req.Severity,
		Status:          domain.InterventionStatusPending,
		Reason:          req.Reason,
		Metadata:        req.Metadata,
		StartedAt:       now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repositories.InterventionRepository.CreateIntervention(ctx, intervention); err != nil {
		return nil, fmt.Errorf("failed to create intervention: %w", err)
	}

	// Publish to Kafka
	if err := s.messagingPort.Producer.PublishIntervention(intervention); err != nil {
		s.logger.Warn("Failed to publish intervention to Kafka", logger.Error(err))
	}

	s.metrics.RecordIntervention(
		intervention.PolicyID.String(),
		intervention.Severity,
		string(intervention.Status),
		float64(time.Since(start).Milliseconds()),
	)

	s.metrics.SetActiveInterventions(float64(s.getActiveInterventionCount(ctx)))

	s.logger.Info("Created intervention",
		logger.String("intervention_id", intervention.ID.String()),
		logger.String("policy_id", intervention.PolicyID.String()),
		logger.String("target", intervention.TargetService),
		logger.String("severity", intervention.Severity),
	)

	return intervention, nil
}

// UpdateStatus updates the status of an intervention
func (s *InterventionServiceService) UpdateStatus(ctx context.Context, id string, status domain.InterventionStatus) error {
	interventionUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid intervention ID: %w", err)
	}

	if err := s.repositories.InterventionRepository.UpdateInterventionStatus(ctx, interventionUUID, status); err != nil {
		return fmt.Errorf("failed to update intervention status: %w", err)
	}

	s.metrics.SetActiveInterventions(float64(s.getActiveInterventionCount(ctx)))

	s.logger.Info("Updated intervention status",
		logger.String("intervention_id", id),
		logger.String("status", string(status)),
	)

	return nil
}

// Resolve resolves an intervention
func (s *InterventionServiceService) Resolve(ctx context.Context, id string, resolution string) error {
	interventionUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid intervention ID: %w", err)
	}

	if err := s.repositories.InterventionRepository.ResolveIntervention(ctx, interventionUUID, resolution); err != nil {
		return fmt.Errorf("failed to resolve intervention: %w", err)
	}

	s.metrics.SetActiveInterventions(float64(s.getActiveInterventionCount(ctx)))

	s.logger.Info("Resolved intervention",
		logger.String("intervention_id", id),
		logger.String("resolution", resolution),
	)

	return nil
}

// StartInterventionMonitor starts the background intervention monitor
func (s *InterventionServiceService) StartInterventionMonitor(logger *zap.Logger) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkActiveInterventions()
			case <-s.stopCh:
				logger.Info("Intervention monitor stopped")
				return
			}
		}
	}()

	logger.Info("Intervention monitor started")
}

// TriggerIntervention triggers an intervention based on policy violation
func (s *InterventionServiceService) TriggerIntervention(ctx context.Context, policyID uuid.UUID, target, reason string, severity string, data map[string]interface{}) (*domain.Intervention, error) {
	req := &domain.CreateInterventionRequest{
		PolicyID:        policyID,
		TargetService:   target,
		InterventionType: "automated",
		Severity:        severity,
		Reason:          reason,
		Metadata:        data,
	}

	return s.CreateIntervention(ctx, req)
}

// checkActiveInterventions checks and updates active interventions
func (s *InterventionServiceService) checkActiveInterventions() {
	ctx := context.Background()

	interventions, err := s.repositories.InterventionRepository.GetActiveInterventions(ctx)
	if err != nil {
		s.logger.Error("Failed to get active interventions", logger.Error(err))
		return
	}

	for _, intervention := range interventions {
		// Check if intervention has exceeded max duration
		maxDuration := s.getMaxDurationForSeverity(intervention.Severity)
		if time.Since(intervention.StartedAt) > maxDuration {
			// Auto-escalate or take further action
			s.logger.Warn("Intervention exceeded max duration",
				logger.String("intervention_id", intervention.ID.String()),
				logger.String("severity", intervention.Severity),
			)
		}
	}
}

// getMaxDurationForSeverity returns the max duration for an intervention based on severity
func (s *InterventionServiceService) getMaxDurationForSeverity(severity string) time.Duration {
	switch severity {
	case "critical":
		return 5 * time.Minute
	case "high":
		return 15 * time.Minute
	case "medium":
		return 30 * time.Minute
	case "low":
		return 1 * time.Hour
	default:
		return 30 * time.Minute
	}
}

// getActiveInterventionCount returns the count of active interventions
func (s *InterventionServiceService) getActiveInterventionCount(ctx context.Context) int {
	interventions, err := s.repositories.InterventionRepository.GetActiveInterventions(ctx)
	if err != nil {
		return 0
	}
	return len(interventions)
}
