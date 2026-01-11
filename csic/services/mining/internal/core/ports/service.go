package ports

import (
	"context"

	"github.com/csic-platform/services/services/mining/internal/core/domain"
	"github.com/google/uuid"
)

// MiningService defines the input port for mining control operations
type MiningService interface {
	// Registry Management
	RegisterOperation(ctx context.Context, req RegisterOperationRequest) (*domain.MiningOperation, error)
	GetOperation(ctx context.Context, id uuid.UUID) (*domain.MiningOperation, error)
	GetOperationByWallet(ctx context.Context, walletAddress string) (*domain.MiningOperation, error)
	ListOperations(ctx context.Context, status *domain.OperationStatus, page, pageSize int) ([]domain.MiningOperation, error)
	UpdateOperationStatus(ctx context.Context, id uuid.UUID, status domain.OperationStatus) error

	// Hashrate Monitoring
	ReportHashrate(ctx context.Context, req TelemetryRequest) error
	GetHashrateHistory(ctx context.Context, opID uuid.UUID, limit int) ([]domain.HashrateRecord, error)
	GetCurrentHashrate(ctx context.Context, opID uuid.UUID) (*domain.HashrateRecord, error)

	// Quota Management
	AssignQuota(ctx context.Context, req QuotaRequest) (*domain.MiningQuota, error)
	GetCurrentQuota(ctx context.Context, opID uuid.UUID) (*domain.MiningQuota, error)
	RevokeQuota(ctx context.Context, quotaID uuid.UUID) error

	// Remote Shutdown
	IssueShutdownCommand(ctx context.Context, req ShutdownRequest) (*domain.ShutdownCommand, error)
	AcknowledgeCommand(ctx context.Context, commandID uuid.UUID) error
	ConfirmShutdown(ctx context.Context, commandID uuid.UUID, result string) error
	GetPendingCommands(ctx context.Context, opID uuid.UUID) ([]domain.ShutdownCommand, error)

	// Compliance and Statistics
	GetRegistryStats(ctx context.Context) (*domain.RegistryStats, error)
}

// DTOs for service operations

// RegisterOperationRequest represents a request to register a new mining operation
type RegisterOperationRequest struct {
	OperatorName  string  `json:"operator_name" binding:"required"`
	WalletAddress string  `json:"wallet_address" binding:"required,eth_addr"`
	Location      string  `json:"location" binding:"required"`
	Region        string  `json:"region" binding:"required"`
	MachineType   string  `json:"machine_type" binding:"required"`
	InitialHashrate float64 `json:"initial_hashrate"`
	Metadata      string  `json:"metadata,omitempty"`
}

// TelemetryRequest represents incoming hashrate telemetry
type TelemetryRequest struct {
	OperationID uuid.UUID `json:"operation_id" binding:"required"`
	Hashrate    float64   `json:"hashrate" binding:"required,gt=0"`
	Unit        string    `json:"unit" binding:"required,oneof=TH/s PH/s GH/s"`
	BlockHeight uint64    `json:"block_height"`
	Timestamp   string    `json:"timestamp"`
}

// QuotaRequest represents a request to assign or update a quota
type QuotaRequest struct {
	OperationID uuid.UUID       `json:"operation_id" binding:"required"`
	MaxHashrate float64         `json:"max_hashrate" binding:"required,gt=0"`
	QuotaType   domain.QuotaType `json:"quota_type" binding:"required,oneof=FIXED DYNAMIC_GRID"`
	ValidFrom   string          `json:"valid_from"`
	ValidTo     string          `json:"valid_to,omitempty"`
	Region      string          `json:"region"`
	Priority    int             `json:"priority"`
}

// ShutdownRequest represents a request to issue a shutdown command
type ShutdownRequest struct {
	OperationID uuid.UUID       `json:"operation_id" binding:"required"`
	CommandType domain.CommandType `json:"command_type" binding:"required,oneof=GRACEFUL IMMEDIATE FORCE_KILL"`
	Reason      string          `json:"reason" binding:"required"`
	IssuedBy    string          `json:"issued_by" binding:"required"`
	ExpiresIn   int             `json:"expires_in,omitempty"` // Duration in seconds
}

// CommandConfirmation represents a confirmation of command execution
type CommandConfirmation struct {
	CommandID uuid.UUID `json:"command_id" binding:"required"`
	Success   bool      `json:"success"`
	Result    string    `json:"result,omitempty"`
}
