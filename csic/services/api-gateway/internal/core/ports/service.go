package ports

import (
	"context"

	"github.com/csic-platform/services/api-gateway/internal/core/domain"
)

// GatewayService defines the interface for gateway business logic
type GatewayService interface {
	// Dashboard operations
	GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error)

	// Exchange operations
	GetExchanges(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)
	GetExchangeByID(ctx context.Context, id string) (*domain.Exchange, error)
	SuspendExchange(ctx context.Context, id, reason, userID string) error

	// Wallet operations
	GetWallets(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)
	GetWalletByID(ctx context.Context, id string) (*domain.Wallet, error)
	FreezeWallet(ctx context.Context, id, reason, userID string) error

	// Miner operations
	GetMiners(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)
	GetMinerByID(ctx context.Context, id string) (*domain.Miner, error)

	// Alert operations
	GetAlerts(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)
	GetAlertByID(ctx context.Context, id string) (*domain.Alert, error)
	AcknowledgeAlert(ctx context.Context, id, userID string) error

	// Compliance operations
	GetComplianceReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)
	GenerateComplianceReport(ctx context.Context, entityType, entityID, period string) (*domain.ComplianceReport, error)

	// Audit operations
	GetAuditLogs(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)
	CreateAuditLog(ctx context.Context, log *domain.AuditLog) error

	// Blockchain operations
	GetBlockchainStatus(ctx context.Context) (map[string]*domain.BlockchainStatus, error)

	// User operations
	GetCurrentUser(ctx context.Context, userID string) (*domain.User, error)
	GetUsers(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Token operations
	GenerateToken(userID, username, role string, permissions []string) (string, error)
	ValidateToken(token string) (*TokenClaims, error)
	RefreshToken(token string) (string, error)

	// User operations
	Authenticate(username, password string) (*domain.User, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) bool
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	Role         string   `json:"role"`
	Permissions  []string `json:"permissions"`
	TokenID      string   `json:"token_id"`
	IssueAt      int64    `json:"iat"`
	ExpiresAt    int64    `json:"exp"`
}

// MessageProducer defines the interface for message publishing operations
type MessageProducer interface {
	// Generic publish
	Publish(ctx context.Context, topic string, message []byte) error

	// Specific event publishing
	PublishAlert(ctx context.Context, alert *domain.Alert) error
	PublishTransaction(ctx context.Context, tx interface{}) error
	PublishAuditLog(ctx context.Context, log *domain.AuditLog) error
	PublishComplianceReport(ctx context.Context, report *domain.ComplianceReport) error
}

// HealthChecker defines the interface for health checking operations
type HealthChecker interface {
	Check(ctx context.Context) HealthStatus
	GetDetails() map[string]HealthComponent
}

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthComponent represents the health of a single component
type HealthComponent struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	LatencyMs int64        `json:"latency_ms"`
}
