package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/csic-platform/services/services/mining/internal/core/domain"
	"github.com/csic-platform/services/services/mining/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MiningService implements the MiningService interface
type MiningService struct {
	repo ports.MiningRepository
	log  *zap.Logger
}

// NewMiningService creates a new MiningService instance
func NewMiningService(repo ports.MiningRepository, log *zap.Logger) *MiningService {
	return &MiningService{
		repo: repo,
		log:  log,
	}
}

// RegisterOperation registers a new mining operation in the national registry
func (s *MiningService) RegisterOperation(ctx context.Context, req ports.RegisterOperationRequest) (*domain.MiningOperation, error) {
	s.log.Info("Registering new mining operation",
		zap.String("operator", req.OperatorName),
		zap.String("wallet", req.WalletAddress),
		zap.String("region", req.Region),
	)

	// Check if wallet address is already registered
	existing, err := s.repo.GetOperationByWallet(ctx, req.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing registration: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("wallet address already registered: %s", req.WalletAddress)
	}

	now := time.Now().UTC()
	op := &domain.MiningOperation{
		ID:             uuid.New(),
		OperatorName:   req.OperatorName,
		WalletAddress:  strings.ToLower(req.WalletAddress),
		CurrentHashrate: req.InitialHashrate,
		Status:         domain.StatusActive,
		Location:       req.Location,
		Region:         req.Region,
		MachineType:    req.MachineType,
		RegisteredAt:   now,
		UpdatedAt:      now,
		ViolationCount: 0,
		Metadata:       req.Metadata,
	}

	if err := s.repo.CreateOperation(ctx, op); err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}

	s.log.Info("Successfully registered mining operation",
		zap.String("operation_id", op.ID.String()),
		zap.String("operator", op.OperatorName),
	)

	return op, nil
}

// GetOperation retrieves a mining operation by ID
func (s *MiningService) GetOperation(ctx context.Context, id uuid.UUID) (*domain.MiningOperation, error) {
	return s.repo.GetOperation(ctx, id)
}

// GetOperationByWallet retrieves a mining operation by wallet address
func (s *MiningService) GetOperationByWallet(ctx context.Context, walletAddress string) (*domain.MiningOperation, error) {
	return s.repo.GetOperationByWallet(ctx, strings.ToLower(walletAddress))
}

// ListOperations retrieves all operations with optional status filter
func (s *MiningService) ListOperations(ctx context.Context, status *domain.OperationStatus, page, pageSize int) ([]domain.MiningOperation, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListOperations(ctx, status, pageSize, offset)
}

// UpdateOperationStatus updates the status of a mining operation
func (s *MiningService) UpdateOperationStatus(ctx context.Context, id uuid.UUID, status domain.OperationStatus) error {
	s.log.Info("Updating operation status",
		zap.String("operation_id", id.String()),
		zap.String("new_status", string(status)),
	)

	// Validate status transition
	existing, err := s.repo.GetOperation(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("operation not found: %s", id.String())
	}

	if !s.isValidStatusTransition(existing.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", existing.Status, status)
	}

	return s.repo.UpdateOperationStatus(ctx, id, status)
}

// ReportHashrate processes incoming hashrate telemetry and checks for quota violations
func (s *MiningService) ReportHashrate(ctx context.Context, req ports.TelemetryRequest) error {
	s.log.Debug("Processing hashrate telemetry",
		zap.String("operation_id", req.OperationID.String()),
		zap.Float64("hashrate", req.Hashrate),
		zap.String("unit", req.Unit),
	)

	// Get operation
	op, err := s.repo.GetOperation(ctx, req.OperationID)
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}
	if op == nil {
		return fmt.Errorf("operation not found: %s", req.OperationID.String())
	}

	// Check if operation is active
	if op.Status != domain.StatusActive {
		s.log.Warn("Received telemetry for non-active operation",
			zap.String("operation_id", req.OperationID.String()),
			zap.String("status", string(op.Status)),
		)
		// Still record the telemetry for audit purposes
	}

	// Convert hashrate to TH/s
	normalizedHashrate := s.normalizeHashrate(req.Hashrate, req.Unit)

	// Create hashrate record
	timestamp := time.Now().UTC()
	if req.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			timestamp = parsed
		}
	}

	record := &domain.HashrateRecord{
		ID:          0, // Auto-generated
		OperationID: req.OperationID,
		Hashrate:    normalizedHashrate,
		Unit:        "TH/s",
		BlockHeight: req.BlockHeight,
		Timestamp:   timestamp,
		SubmittedAt: time.Now().UTC(),
	}

	if err := s.repo.RecordHashrate(ctx, record); err != nil {
		return fmt.Errorf("failed to record hashrate: %w", err)
	}

	// Update operation's current hashrate
	if err := s.repo.UpdateOperationHashrate(ctx, req.OperationID, normalizedHashrate, timestamp); err != nil {
		s.log.Error("Failed to update operation hashrate", zap.Error(err))
	}

	// Check for quota violations
	quota, err := s.repo.GetCurrentQuota(ctx, req.OperationID)
	if err != nil {
		s.log.Error("Failed to get current quota", zap.Error(err))
	}

	if quota != nil && normalizedHashrate > quota.MaxHashrate {
		return s.handleQuotaViolation(ctx, req.OperationID, normalizedHashrate, *quota)
	}

	return nil
}

// GetHashrateHistory retrieves historical hashrate records for an operation
func (s *MiningService) GetHashrateHistory(ctx context.Context, opID uuid.UUID, limit int) ([]domain.HashrateRecord, error) {
	if limit < 1 || limit > 10000 {
		limit = 1000
	}
	return s.repo.GetHashrateHistory(ctx, opID, limit)
}

// GetCurrentQuota retrieves the current quota for an operation
func (s *MiningService) GetCurrentQuota(ctx context.Context, opID uuid.UUID) (*domain.MiningQuota, error) {
	return s.repo.GetCurrentQuota(ctx, opID)
}

// AssignQuota assigns a new quota to a mining operation
func (s *MiningService) AssignQuota(ctx context.Context, req ports.QuotaRequest) (*domain.MiningQuota, error) {
	s.log.Info("Assigning quota to operation",
		zap.String("operation_id", req.OperationID.String()),
		zap.Float64("max_hashrate", req.MaxHashrate),
		zap.String("quota_type", string(req.QuotaType)),
	)

	// Verify operation exists
	op, err := s.repo.GetOperation(ctx, req.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}
	if op == nil {
		return nil, fmt.Errorf("operation not found: %s", req.OperationID.String())
	}

	now := time.Now().UTC()
	validFrom := now
	if req.ValidFrom != "" {
		if parsed, err := time.Parse(time.RFC3339, req.ValidFrom); err == nil {
			validFrom = parsed
		}
	}

	var validTo *time.Time
	if req.ValidTo != "" {
		if parsed, err := time.Parse(time.RFC3339, req.ValidTo); err == nil {
			validTo = &parsed
		}
	}

	quota := &domain.MiningQuota{
		ID:          uuid.New(),
		OperationID: req.OperationID,
		MaxHashrate: req.MaxHashrate,
		QuotaType:   req.QuotaType,
		ValidFrom:   validFrom,
		ValidTo:     validTo,
		Region:      req.Region,
		Priority:    req.Priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.SetQuota(ctx, quota); err != nil {
		return nil, fmt.Errorf("failed to set quota: %w", err)
	}

	return quota, nil
}

// RevokeQuota revokes a quota by ID
func (s *MiningService) RevokeQuota(ctx context.Context, quotaID uuid.UUID) error {
	s.log.Info("Revoking quota", zap.String("quota_id", quotaID.String()))
	return s.repo.DeleteQuota(ctx, quotaID)
}

// IssueShutdownCommand issues a remote shutdown command to a mining operation
func (s *MiningService) IssueShutdownCommand(ctx context.Context, req ports.ShutdownRequest) (*domain.ShutdownCommand, error) {
	s.log.Info("Issuing shutdown command",
		zap.String("operation_id", req.OperationID.String()),
		zap.String("command_type", string(req.CommandType)),
		zap.String("reason", req.Reason),
		zap.String("issued_by", req.IssuedBy),
	)

	// Verify operation exists
	op, err := s.repo.GetOperation(ctx, req.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}
	if op == nil {
		return nil, fmt.Errorf("operation not found: %s", req.OperationID.String())
	}

	now := time.Now().UTC()
	cmd := &domain.ShutdownCommand{
		ID:          uuid.New(),
		OperationID: req.OperationID,
		CommandType: req.CommandType,
		Reason:      req.Reason,
		Status:      domain.CommandIssued,
		IssuedAt:    now,
		IssuedBy:    req.IssuedBy,
	}

	// Set expiration if specified
	if req.ExpiresIn > 0 {
		expiresAt := now.Add(time.Duration(req.ExpiresIn) * time.Second)
		cmd.ExpiresAt = &expiresAt
	}

	if err := s.repo.CreateShutdownCommand(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to create shutdown command: %w", err)
	}

	// Update operation status
	if err := s.repo.UpdateOperationStatus(ctx, req.OperationID, domain.StatusShutdownOrdered); err != nil {
		s.log.Error("Failed to update operation status", zap.Error(err))
	}

	return cmd, nil
}

// AcknowledgeCommand acknowledges receipt of a shutdown command
func (s *MiningService) AcknowledgeCommand(ctx context.Context, commandID uuid.UUID) error {
	s.log.Info("Acknowledging shutdown command", zap.String("command_id", commandID.String()))

	cmd, err := s.repo.GetShutdownCommand(ctx, commandID)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}
	if cmd == nil {
		return fmt.Errorf("command not found: %s", commandID.String())
	}

	if cmd.Status != domain.CommandIssued {
		return fmt.Errorf("command is not in ISSUED status: %s", cmd.Status)
	}

	now := time.Now().UTC()
	return s.repo.UpdateCommandStatus(ctx, commandID, domain.CommandAcknowledged, &now)
}

// ConfirmShutdown confirms execution of a shutdown command
func (s *MiningService) ConfirmShutdown(ctx context.Context, commandID uuid.UUID, result string) error {
	s.log.Info("Confirming shutdown command", zap.String("command_id", commandID.String()))

	cmd, err := s.repo.GetShutdownCommand(ctx, commandID)
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}
	if cmd == nil {
		return fmt.Errorf("command not found: %s", commandID.String())
	}

	now := time.Now().UTC()
	status := domain.CommandExecuted
	if !strings.Contains(strings.ToLower(result), "success") {
		status = domain.CommandFailed
	}

	if err := s.repo.UpdateCommandStatus(ctx, commandID, status, &now); err != nil {
		return err
	}

	// Update operation status based on command result
	if status == domain.CommandExecuted {
		return s.repo.UpdateOperationStatus(ctx, cmd.OperationID, domain.StatusShutdownExecuted)
	}

	return nil
}

// GetPendingCommands retrieves pending shutdown commands for an operation
func (s *MiningService) GetPendingCommands(ctx context.Context, opID uuid.UUID) ([]domain.ShutdownCommand, error) {
	return s.repo.GetPendingCommands(ctx, opID)
}

// GetRegistryStats retrieves statistics for the national mining registry
func (s *MiningService) GetRegistryStats(ctx context.Context) (*domain.RegistryStats, error) {
	return s.repo.GetRegistryStats(ctx)
}

// GetCurrentHashrate retrieves the latest hashrate record for an operation
func (s *MiningService) GetCurrentHashrate(ctx context.Context, opID uuid.UUID) (*domain.HashrateRecord, error) {
	return s.repo.GetLatestHashrate(ctx, opID)
}

// Helper functions

// normalizeHashrate converts hashrate to TH/s
func (s *MiningService) normalizeHashrate(hashrate float64, unit string) float64 {
	switch strings.ToUpper(unit) {
	case "PH/s":
		return hashrate * 1000
	case "GH/s":
		return hashrate / 1000
	case "TH/s":
		return hashrate
	default:
		return hashrate
	}
}

// isValidStatusTransition checks if a status transition is valid
func (s *MiningService) isValidStatusTransition(from, to domain.OperationStatus) bool {
	validTransitions := map[domain.OperationStatus][]domain.OperationStatus{
		StatusActive:              {StatusSuspended, StatusShutdownOrdered, StatusNonCompliant},
		StatusSuspended:           {StatusActive, StatusShutdownOrdered},
		StatusShutdownOrdered:     {StatusShutdownExecuted, StatusActive},
		StatusNonCompliant:        {StatusSuspended, StatusActive},
		StatusShutdownExecuted:    {StatusActive},
		StatusPendingRegistration: {StatusActive, StatusSuspended},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// handleQuotaViolation processes a quota violation
func (s *MiningService) handleQuotaViolation(ctx context.Context, opID uuid.UUID, reportedHashrate float64, quota domain.MiningQuota) error {
	s.log.Warn("Quota violation detected",
		zap.String("operation_id", opID.String()),
		zap.Float64("reported_hashrate", reportedHashrate),
		zap.Float64("max_hashrate", quota.MaxHashrate),
	)

	// Record violation
	violation := &domain.QuotaViolation{
		ID:                uuid.New(),
		OperationID:       opID,
		QuotaID:           quota.ID,
		ReportedHashrate:  reportedHashrate,
		MaxHashrate:       quota.MaxHashrate,
		RecordedAt:        time.Now().UTC(),
		Consecutive:       true,
		AutomaticShutdown: false,
	}

	if err := s.repo.RecordViolation(ctx, violation); err != nil {
		s.log.Error("Failed to record violation", zap.Error(err))
	}

	// Increment violation count on operation
	op, err := s.repo.GetOperation(ctx, opID)
	if err != nil {
		s.log.Error("Failed to get operation for violation count", zap.Error(err))
	} else {
		op.ViolationCount++
		if err := s.repo.UpdateOperationStatus(ctx, opID, op.Status); err != nil {
			s.log.Error("Failed to update violation count", zap.Error(err))
		}
	}

	// Check if automatic shutdown should be triggered
	// This would typically come from config, using a default for now
	violationThreshold := 3
	if op.ViolationCount >= violationThreshold {
		s.log.Warn("Threshold exceeded, triggering automatic shutdown",
			zap.String("operation_id", opID.String()),
			zap.Int("violation_count", op.ViolationCount),
		)

		// Create automatic shutdown command
		autoCmd := &domain.ShutdownCommand{
			ID:          uuid.New(),
			OperationID: opID,
			CommandType: domain.CommandGraceful,
			Reason:      fmt.Sprintf("Automatic shutdown due to quota violation (threshold: %d)", violationThreshold),
			Status:      domain.CommandIssued,
			IssuedAt:    time.Now().UTC(),
			IssuedBy:    "SYSTEM",
		}

		if err := s.repo.CreateShutdownCommand(ctx, autoCmd); err != nil {
			s.log.Error("Failed to create automatic shutdown command", zap.Error(err))
		}

		// Update operation status
		if err := s.repo.UpdateOperationStatus(ctx, opID, domain.StatusNonCompliant); err != nil {
			s.log.Error("Failed to update operation status", zap.Error(err))
		}

		violation.AutomaticShutdown = true
	}

	return nil
}
