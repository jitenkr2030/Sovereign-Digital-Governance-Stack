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

// EnforcementHandler handles enforcement actions
type EnforcementHandler interface {
	ListEnforcements(ctx context.Context) ([]*domain.Enforcement, error)
	GetEnforcement(ctx context.Context, id string) (*domain.Enforcement, error)
	CreateEnforcement(ctx context.Context, req *domain.CreateEnforcementRequest) (*domain.Enforcement, error)
	UpdateStatus(ctx context.Context, id string, status domain.EnforcementStatus) error
	GetStats(ctx context.Context, since time.Time) (*domain.EnforcementStats, error)
}

// EnforcementHandlerService implements the EnforcementHandler interface
type EnforcementHandlerService struct {
	repositories  ports.Repositories
	messagingPort ports.MessagingPort
	logger        *zap.Logger
	metrics       *metrics.MetricsCollector
}

// NewEnforcementHandler creates a new enforcement handler
func NewEnforcementHandler(
	repositories ports.Repositories,
	messagingPort ports.MessagingPort,
	logger *zap.Logger,
	metricsCollector *metrics.MetricsCollector,
) EnforcementHandler {
	return &EnforcementHandlerService{
		repositories:  repositories,
		messagingPort: messagingPort,
		logger:        logger,
		metrics:       metricsCollector,
	}
}

// ListEnforcements lists all enforcements
func (h *EnforcementHandlerService) ListEnforcements(ctx context.Context) ([]*domain.Enforcement, error) {
	return h.repositories.EnforcementRepository.GetRecentEnforcements(ctx, time.Now().Add(-24*time.Hour))
}

// GetEnforcement gets an enforcement by ID
func (h *EnforcementHandlerService) GetEnforcement(ctx context.Context, id string) (*domain.Enforcement, error) {
	enforcementUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid enforcement ID: %w", err)
	}

	return h.repositories.EnforcementRepository.GetEnforcementByID(ctx, enforcementUUID)
}

// CreateEnforcement creates a new enforcement
func (h *EnforcementHandlerService) CreateEnforcement(ctx context.Context, req *domain.CreateEnforcementRequest) (*domain.Enforcement, error) {
	start := time.Now()

	now := time.Now()
	enforcement := &domain.Enforcement{
		ID:            uuid.New(),
		PolicyID:      req.PolicyID,
		TargetService: req.TargetService,
		ActionType:    req.ActionType,
		Severity:      req.Severity,
		Status:        domain.EnforcementStatusPending,
		Message:       req.Message,
		Metadata:      req.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.repositories.EnforcementRepository.CreateEnforcement(ctx, enforcement); err != nil {
		return nil, fmt.Errorf("failed to create enforcement: %w", err)
	}

	// Publish to Kafka
	if err := h.messagingPort.Producer.PublishEnforcement(enforcement); err != nil {
		h.logger.Warn("Failed to publish enforcement to Kafka", logger.Error(err))
	}

	h.metrics.RecordEnforcementAction(
		enforcement.PolicyID.String(),
		enforcement.ActionType,
		enforcement.TargetService,
		float64(time.Since(start).Milliseconds()),
	)

	h.logger.Info("Created enforcement",
		logger.String("enforcement_id", enforcement.ID.String()),
		logger.String("policy_id", enforcement.PolicyID.String()),
		logger.String("action", enforcement.ActionType),
		logger.String("target", enforcement.TargetService),
	)

	return enforcement, nil
}

// UpdateStatus updates the status of an enforcement
func (h *EnforcementHandlerService) UpdateStatus(ctx context.Context, id string, status domain.EnforcementStatus) error {
	enforcementUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid enforcement ID: %w", err)
	}

	if err := h.repositories.EnforcementRepository.UpdateEnforcementStatus(ctx, enforcementUUID, status); err != nil {
		return fmt.Errorf("failed to update enforcement status: %w", err)
	}

	h.logger.Info("Updated enforcement status",
		logger.String("enforcement_id", id),
		logger.String("status", string(status)),
	)

	return nil
}

// GetStats gets enforcement statistics
func (h *EnforcementHandlerService) GetStats(ctx context.Context, since time.Time) (*domain.EnforcementStats, error) {
	return h.repositories.EnforcementRepository.GetEnforcementStats(ctx, since)
}

// EnforcementAction performs an enforcement action
func (h *EnforcementHandlerService) EnforcementAction(ctx context.Context, policyID uuid.UUID, target, actionType, severity, message string, metadata map[string]interface{}) (*domain.Enforcement, error) {
	req := &domain.CreateEnforcementRequest{
		PolicyID:      policyID,
		TargetService: target,
		ActionType:    actionType,
		Severity:      severity,
		Message:       message,
		Metadata:      metadata,
	}

	return h.CreateEnforcement(ctx, req)
}
