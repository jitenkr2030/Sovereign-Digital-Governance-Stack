package ports

import (
	"context"

	"github.com/csic-platform/services/services/mining/internal/core/domain"
	"github.com/google/uuid"
)

// MiningRepository defines the output port for mining operations
type MiningRepository interface {
	// Registry operations
	CreateOperation(ctx context.Context, op *domain.MiningOperation) error
	GetOperation(ctx context.Context, id uuid.UUID) (*domain.MiningOperation, error)
	GetOperationByWallet(ctx context.Context, walletAddress string) (*domain.MiningOperation, error)
	UpdateOperationStatus(ctx context.Context, id uuid.UUID, status domain.OperationStatus) error
	UpdateOperationHashrate(ctx context.Context, id uuid.UUID, hashrate float64, timestamp time.Time) error
	ListOperations(ctx context.Context, status *domain.OperationStatus, limit, offset int) ([]domain.MiningOperation, error)
	CountOperations(ctx context.Context, status *domain.OperationStatus) (int64, error)

	// Hashrate monitoring
	RecordHashrate(ctx context.Context, record *domain.HashrateRecord) error
	GetHashrateHistory(ctx context.Context, opID uuid.UUID, limit int) ([]domain.HashrateRecord, error)
	GetLatestHashrate(ctx context.Context, opID uuid.UUID) (*domain.HashrateRecord, error)
	CountHashrateRecords(ctx context.Context, opID uuid.UUID) (int64, error)

	// Quota management
	SetQuota(ctx context.Context, quota *domain.MiningQuota) error
	GetCurrentQuota(ctx context.Context, opID uuid.UUID) (*domain.MiningQuota, error)
	GetQuotas(ctx context.Context, opID uuid.UUID) ([]domain.MiningQuota, error)
	DeleteQuota(ctx context.Context, id uuid.UUID) error

	// Shutdown commands
	CreateShutdownCommand(ctx context.Context, cmd *domain.ShutdownCommand) error
	GetShutdownCommand(ctx context.Context, id uuid.UUID) (*domain.ShutdownCommand, error)
	GetPendingCommands(ctx context.Context, opID uuid.UUID) ([]domain.ShutdownCommand, error)
	UpdateCommandStatus(ctx context.Context, id uuid.UUID, status domain.CommandStatus, executedAt *time.Time) error
	GetAllPendingCommands(ctx context.Context) ([]domain.ShutdownCommand, error)

	// Violations
	RecordViolation(ctx context.Context, violation *domain.QuotaViolation) error
	GetViolations(ctx context.Context, opID uuid.UUID, limit int) ([]domain.QuotaViolation, error)

	// Statistics
	GetRegistryStats(ctx context.Context) (*domain.RegistryStats, error)
}

// Time is needed for repository operations
import "time"
