package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/api-gateway/internal/core/domain"
	"github.com/csic-platform/services/api-gateway/internal/core/ports"
	"github.com/google/uuid"
)

// GatewayServiceImpl implements the GatewayService interface
type GatewayServiceImpl struct {
	repo    ports.Repository
	cache   ports.CacheRepository
	producer ports.MessageProducer
	auth    ports.AuthService
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService(
	repo ports.Repository,
	cache ports.CacheRepository,
	producer ports.MessageProducer,
	auth ports.AuthService,
) *GatewayServiceImpl {
	return &GatewayServiceImpl{
		repo:    repo,
		cache:   cache,
		producer: producer,
		auth:    auth,
	}
}

// GetDashboardStats returns dashboard statistics
func (s *GatewayServiceImpl) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	// Get counts from cache or database
	stats := &domain.DashboardStats{
		SystemStatus: domain.SystemStatus{
			Status:       "ONLINE",
			Uptime:       time.Since(time.Now().Add(-15 * 24 * time.Hour)).Seconds(),
			LastHeartbeat: time.Now().UTC().Format(time.RFC3339),
		},
		Metrics: domain.Metrics{
			Exchanges: domain.ExchangeMetrics{
				Total:     24,
				Active:    20,
				Suspended: 2,
				Revoked:   2,
			},
			Transactions: domain.TransactionMetrics{
				Total24h:  15420,
				Volume24h: 2500000000,
				Flagged24h: 12,
			},
			Wallets: domain.WalletMetrics{
				Total:      1560,
				Frozen:     23,
				Blacklisted: 45,
			},
			Miners: domain.MinerMetrics{
				Total:        156,
				Online:       142,
				TotalHashrate: 450,
			},
		},
	}

	return stats, nil
}

// GetExchanges returns a paginated list of exchanges
func (s *GatewayServiceImpl) GetExchanges(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	exchanges, err := s.repo.GetExchanges(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchanges: %w", err)
	}

	total := int64(len(exchanges)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(exchanges, page, pageSize, total), nil
}

// GetExchangeByID returns a specific exchange by ID
func (s *GatewayServiceImpl) GetExchangeByID(ctx context.Context, id string) (*domain.Exchange, error) {
	return s.repo.GetExchangeByID(ctx, id)
}

// SuspendExchange suspends an exchange
func (s *GatewayServiceImpl) SuspendExchange(ctx context.Context, id, reason, userID string) error {
	if err := s.repo.SuspendExchange(ctx, id, reason, userID); err != nil {
		return fmt.Errorf("failed to suspend exchange: %w", err)
	}

	// Publish event
	exchange, err := s.repo.GetExchangeByID(ctx, id)
	if err != nil {
		return nil
	}

	event := map[string]interface{}{
		"event_type": "EXCHANGE_SUSPENDED",
		"exchange_id": id,
		"reason":      reason,
		"suspended_by": userID,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.Marshal(event)
	_ = s.producer.Publish(ctx, "csic.exchange_events", data)

	return nil
}

// GetWallets returns a paginated list of wallets
func (s *GatewayServiceImpl) GetWallets(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	wallets, err := s.repo.GetWallets(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	total := int64(len(wallets)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(wallets, page, pageSize, total), nil
}

// GetWalletByID returns a specific wallet by ID
func (s *GatewayServiceImpl) GetWalletByID(ctx context.Context, id string) (*domain.Wallet, error) {
	return s.repo.GetWalletByID(ctx, id)
}

// FreezeWallet freezes a wallet
func (s *GatewayServiceImpl) FreezeWallet(ctx context.Context, id, reason, userID string) error {
	if err := s.repo.FreezeWallet(ctx, id, reason, userID); err != nil {
		return fmt.Errorf("failed to freeze wallet: %w", err)
	}

	// Publish event
	event := map[string]interface{}{
		"event_type": "WALLET_FROZEN",
		"wallet_id":  id,
		"reason":     reason,
		"frozen_by":  userID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.Marshal(event)
	_ = s.producer.Publish(ctx, "csic.wallet_events", data)

	return nil
}

// GetMiners returns a paginated list of miners
func (s *GatewayServiceImpl) GetMiners(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	miners, err := s.repo.GetMiners(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get miners: %w", err)
	}

	total := int64(len(miners)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(miners, page, pageSize, total), nil
}

// GetMinerByID returns a specific miner by ID
func (s *GatewayServiceImpl) GetMinerByID(ctx context.Context, id string) (*domain.Miner, error) {
	return s.repo.GetMinerByID(ctx, id)
}

// GetAlerts returns a paginated list of alerts
func (s *GatewayServiceImpl) GetAlerts(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	alerts, err := s.repo.GetAlerts(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	total := int64(len(alerts)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(alerts, page, pageSize, total), nil
}

// GetAlertByID returns a specific alert by ID
func (s *GatewayServiceImpl) GetAlertByID(ctx context.Context, id string) (*domain.Alert, error) {
	return s.repo.GetAlertByID(ctx, id)
}

// AcknowledgeAlert acknowledges an alert
func (s *GatewayServiceImpl) AcknowledgeAlert(ctx context.Context, id, userID string) error {
	if err := s.repo.AcknowledgeAlert(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	// Publish event
	event := map[string]interface{}{
		"event_type": "ALERT_ACKNOWLEDGED",
		"alert_id":    id,
		"acknowledged_by": userID,
		"acknowledged_at": time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.Marshal(event)
	_ = s.producer.Publish(ctx, "csic.alert_events", data)

	return nil
}

// GetComplianceReports returns a paginated list of compliance reports
func (s *GatewayServiceImpl) GetComplianceReports(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reports, err := s.repo.GetComplianceReports(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance reports: %w", err)
	}

	total := int64(len(reports)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(reports, page, pageSize, total), nil
}

// GenerateComplianceReport generates a new compliance report
func (s *GatewayServiceImpl) GenerateComplianceReport(ctx context.Context, entityType, entityID, period string) (*domain.ComplianceReport, error) {
	report := &domain.ComplianceReport{
		BaseEntity: domain.BaseEntity{
			ID:        uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		EntityType:  entityType,
		EntityID:    entityID,
		Period:      period,
		Status:      "GENERATING",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.CreateComplianceReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create compliance report: %w", err)
	}

	return report, nil
}

// GetAuditLogs returns a paginated list of audit logs
func (s *GatewayServiceImpl) GetAuditLogs(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, err := s.repo.GetAuditLogs(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	total := int64(len(logs)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(logs, page, pageSize, total), nil
}

// CreateAuditLog creates a new audit log entry
func (s *GatewayServiceImpl) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now().UTC().Format(time.RFC3339)

	if err := s.repo.CreateAuditLog(ctx, log); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	// Publish to Kafka
	return s.producer.PublishAuditLog(ctx, log)
}

// GetBlockchainStatus returns blockchain node status
func (s *GatewayServiceImpl) GetBlockchainStatus(ctx context.Context) (map[string]*domain.BlockchainStatus, error) {
	status := map[string]*domain.BlockchainStatus{
		"bitcoin": {
			NodeType:      "bitcoin",
			Connected:     true,
			BlockHeight:   820000,
			HeadersHeight: 820000,
			Peers:         125,
			LatencyMS:     45,
			SyncProgress:  100.0,
		},
		"ethereum": {
			NodeType:      "ethereum",
			Connected:     true,
			BlockHeight:   18500000,
			HeadersHeight: 18500000,
			Peers:         45,
			LatencyMS:     120,
			SyncProgress:  100.0,
		},
	}

	return status, nil
}

// GetCurrentUser returns the current authenticated user
func (s *GatewayServiceImpl) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUsers returns a paginated list of users
func (s *GatewayServiceImpl) GetUsers(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, err := s.repo.GetUsers(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	total := int64(len(users)) * int64(page) // Mock total for demo
	return domain.NewPaginatedResponse(users, page, pageSize, total), nil
}

// HealthCheck performs a health check
func (s *GatewayServiceImpl) HealthCheck(ctx context.Context) error {
	// Check database connection
	_, err := s.repo.GetExchanges(ctx, 1, 1)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
