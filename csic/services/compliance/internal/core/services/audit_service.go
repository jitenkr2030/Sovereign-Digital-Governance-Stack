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

// AuditService implements the AuditService interface
type AuditService struct {
	repo ports.AuditRepository
	log  *zap.Logger
}

// NewAuditService creates a new AuditService instance
func NewAuditService(repo ports.AuditRepository, log *zap.Logger) *AuditService {
	return &AuditService{
		repo: repo,
		log:  log,
	}
}

// CreateAuditRecord creates a new audit record
func (s *AuditService) CreateAuditRecord(ctx context.Context, req ports.CreateAuditRequest) (*domain.AuditRecord, error) {
	s.log.Debug("Creating audit record",
		zap.String("entity_id", req.EntityID.String()),
		zap.String("action", req.ActionType),
		zap.String("resource", req.ResourceType),
	)

	record := &domain.AuditRecord{
		ID:           uuid.New(),
		EntityID:     req.EntityID,
		ActionType:   req.ActionType,
		ActorID:      req.ActorID,
		ActorType:    req.ActorType,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
		Timestamp:    time.Now().UTC(),
		OldValue:     req.OldValue,
		NewValue:     req.NewValue,
		Changes:      req.Changes,
		Metadata:     req.Metadata,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
	}

	if err := s.repo.CreateAuditRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create audit record: %w", err)
	}

	s.log.Debug("Audit record created", zap.String("record_id", record.ID.String()))
	return record, nil
}

// GetAuditRecord retrieves an audit record by ID
func (s *AuditService) GetAuditRecord(ctx context.Context, recordID uuid.UUID) (*domain.AuditRecord, error) {
	return s.repo.GetAuditRecord(ctx, recordID)
}

// GetEntityAuditTrail retrieves the audit trail for an entity
func (s *AuditService) GetEntityAuditTrail(ctx context.Context, entityID uuid.UUID, limit, offset int) ([]domain.AuditRecord, error) {
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	return s.repo.GetAuditRecordsByEntity(ctx, entityID, limit, offset)
}

// GetResourceAuditTrail retrieves the audit trail for a resource
func (s *AuditService) GetResourceAuditTrail(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]domain.AuditRecord, error) {
	return s.repo.GetAuditRecordsByResource(ctx, resourceType, resourceID)
}

// GetAuditLogs retrieves audit logs with filters
func (s *AuditService) GetAuditLogs(ctx context.Context, filter ports.AuditLogFilter) ([]domain.AuditRecord, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 1000 {
		filter.PageSize = 100
	}

	offset := (filter.Page - 1) * filter.PageSize

	auditFilter := ports.AuditFilter{
		EntityID:     filter.EntityID,
		ActorID:      filter.ActorID,
		ResourceType: filter.ResourceType,
		ResourceID:   filter.ResourceID,
		ActionType:   filter.ActionType,
		From:         filter.From,
		To:           filter.To,
		Limit:        filter.PageSize,
		Offset:       offset,
	}

	return s.repo.GetAuditRecords(ctx, auditFilter)
}

// CountAuditLogs counts audit logs matching the filter
func (s *AuditService) CountAuditLogs(ctx context.Context, filter ports.AuditLogFilter) (int64, error) {
	auditFilter := ports.AuditFilter{
		EntityID:     filter.EntityID,
		ActorID:      filter.ActorID,
		ResourceType: filter.ResourceType,
		ResourceID:   filter.ResourceID,
		ActionType:   filter.ActionType,
		From:         filter.From,
		To:           filter.To,
	}
	return s.repo.CountAuditRecords(ctx, auditFilter)
}

// Helper function to create audit record from business operation
func (s *AuditService) LogAction(ctx context.Context, entityID, actorID uuid.UUID, actorType, actionType, resourceType string, resourceID uuid.UUID, oldVal, newVal interface{}) error {
	oldJSON := ""
	newJSON := ""
	changes := ""

	if oldVal != nil {
		if bytes, err := json.Marshal(oldVal); err == nil {
			oldJSON = string(bytes)
		}
	}
	if newVal != nil {
		if bytes, err := json.Marshal(newVal); err == nil {
			newJSON = string(bytes)
		}
	}

	// Generate changes summary
	if oldJSON != newJSON {
		changes = fmt.Sprintf("Changed from %s to %s", oldJSON, newJSON)
	}

	req := ports.CreateAuditRequest{
		EntityID:     entityID,
		ActionType:   actionType,
		ActorID:      actorID,
		ActorType:    actorType,
		ResourceID:   resourceID,
		ResourceType: resourceType,
		OldValue:     oldJSON,
		NewValue:     newJSON,
		Changes:      changes,
	}

	_, err := s.CreateAuditRecord(ctx, req)
	return err
}
