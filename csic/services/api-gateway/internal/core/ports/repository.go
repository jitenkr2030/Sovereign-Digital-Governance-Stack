package ports

import (
	"context"

	"github.com/csic-platform/services/api-gateway/internal/core/domain"
)

// Repository defines the interface for data access operations
type Repository interface {
	// Exchange operations
	GetExchanges(ctx context.Context, page, pageSize int) ([]*domain.Exchange, error)
	GetExchangeByID(ctx context.Context, id string) (*domain.Exchange, error)
	CreateExchange(ctx context.Context, exchange *domain.Exchange) error
	UpdateExchange(ctx context.Context, exchange *domain.Exchange) error
	SuspendExchange(ctx context.Context, id, reason, userID string) error

	// Wallet operations
	GetWallets(ctx context.Context, page, pageSize int) ([]*domain.Wallet, error)
	GetWalletByID(ctx context.Context, id string) (*domain.Wallet, error)
	GetWalletByAddress(ctx context.Context, address string) (*domain.Wallet, error)
	CreateWallet(ctx context.Context, wallet *domain.Wallet) error
	UpdateWallet(ctx context.Context, wallet *domain.Wallet) error
	FreezeWallet(ctx context.Context, id, reason, userID string) error

	// Miner operations
	GetMiners(ctx context.Context, page, pageSize int) ([]*domain.Miner, error)
	GetMinerByID(ctx context.Context, id string) (*domain.Miner, error)
	CreateMiner(ctx context.Context, miner *domain.Miner) error
	UpdateMiner(ctx context.Context, miner *domain.Miner) error

	// Alert operations
	GetAlerts(ctx context.Context, page, pageSize int) ([]*domain.Alert, error)
	GetAlertByID(ctx context.Context, id string) (*domain.Alert, error)
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	AcknowledgeAlert(ctx context.Context, id, userID string) error

	// Compliance report operations
	GetComplianceReports(ctx context.Context, page, pageSize int) ([]*domain.ComplianceReport, error)
	GetComplianceReportByID(ctx context.Context, id string) (*domain.ComplianceReport, error)
	CreateComplianceReport(ctx context.Context, report *domain.ComplianceReport) error

	// Audit log operations
	GetAuditLogs(ctx context.Context, page, pageSize int) ([]*domain.AuditLog, error)
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error

	// User operations
	GetUsers(ctx context.Context, page, pageSize int) ([]*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	// Generic cache operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Rate limiting
	IncrementRateLimit(ctx context.Context, key string, window int) (int, error)
	GetRateLimit(ctx context.Context, key string, window int) (int, error)

	// Session management
	SetSession(ctx context.Context, sessionID string, userID string, ttl int) error
	GetSession(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
