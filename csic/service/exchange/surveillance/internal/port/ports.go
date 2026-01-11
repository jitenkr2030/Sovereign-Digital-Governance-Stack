package port

import (
	"context"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/google/uuid"
)

// AlertRepository defines the interface for alert persistence
type AlertRepository interface {
	// CRUD operations
	Create(ctx context.Context, alert *domain.Alert) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Alert, error)
	Update(ctx context.Context, alert *domain.Alert) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter domain.AlertFilter, limit, offset int) ([]domain.Alert, error)
	Count(ctx context.Context, filter domain.AlertFilter) (int64, error)
	GetByStatus(ctx context.Context, status domain.AlertStatus) ([]domain.Alert, error)
	GetByExchange(ctx context.Context, exchangeID uuid.UUID, limit, offset int) ([]domain.Alert, error)
	GetBySymbol(ctx context.Context, symbol string, limit, offset int) ([]domain.Alert, error)
	GetOpenAlerts(ctx context.Context) ([]domain.Alert, error)
	GetRecentAlerts(ctx context.Context, since time.Time) ([]domain.Alert, error)

	// Statistics
	GetStatistics(ctx context.Context, startTime, endTime time.Time) (*domain.AlertStatistics, error)
	GetAlertsByType(ctx context.Context, alertType domain.AlertType, startTime, endTime time.Time) ([]domain.Alert, error)
	GetAlertTrend(ctx context.Context, interval string, startTime, endTime time.Time) ([]domain.AlertTrendPoint, error)

	// Bulk operations
	BulkCreate(ctx context.Context, alerts []domain.Alert) error
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.AlertStatus) error
}

// MarketRepository defines the interface for market data persistence
type MarketRepository interface {
	// Market event operations
	StoreEvent(ctx context.Context, event *domain.MarketEvent) error
	StoreEvents(ctx context.Context, events []domain.MarketEvent) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*domain.MarketEvent, error)
	GetEventsBySymbol(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error)
	GetEventsByExchange(ctx context.Context, exchangeID uuid.UUID, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error)
	GetRecentEvents(ctx context.Context, exchangeID uuid.UUID, symbol string, duration time.Duration) ([]domain.MarketEvent, error)

	// Trade operations
	GetTradesByAccount(ctx context.Context, accountID string, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error)
	GetTradesByTimeWindow(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.MarketEvent, error)
	GetRelatedTrades(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, accountIDs []string) ([]domain.MarketEvent, error)

	// Order operations
	GetOrdersByAccount(ctx context.Context, accountID string, startTime, endTime time.Time) ([]domain.MarketEvent, error)
	GetCancelledOrders(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.MarketEvent, error)
	GetOrderLifetimes(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.OrderLifetime, error)

	// Statistics and aggregation
	GetMarketSummary(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) (*domain.MarketSummary, error)
	GetMarketStats(ctx context.Context, exchangeID uuid.UUID, symbol string, duration time.Duration) (*domain.MarketStats, error)
	GetPriceHistory(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, interval string) ([]domain.PricePoint, error)
	GetVolumeHistory(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, interval string) ([]domain.VolumePoint, error)
	GetGlobalAveragePrice(ctx context.Context, symbol string, startTime, endTime time.Time) (decimal.Decimal, error)

	// Order book operations
	StoreOrderBook(ctx context.Context, snapshot *domain.OrderBookSnapshot) error
	GetOrderBookSnapshot(ctx context.Context, id uuid.UUID) (*domain.OrderBookSnapshot, error)

	// Bulk operations
	BulkStoreEvents(ctx context.Context, events []domain.MarketEvent) error
}

// OrderLifetime represents the lifetime of an order
type OrderLifetime struct {
	OrderID       uuid.UUID       `json:"order_id"`
	AccountID     string          `json:"account_id"`
	InitialQty    decimal.Decimal `json:"initial_qty"`
	FilledQty     decimal.Decimal `json:"filled_qty"`
	CancelledQty  decimal.Decimal `json:"cancelled_qty"`
	Lifetime      time.Duration   `json:"lifetime"`
	CreatedAt     time.Time       `json:"created_at"`
	CancelledAt   *time.Time      `json:"cancelled_at,omitempty"`
}

// IngestionService defines the interface for market data ingestion
type IngestionService interface {
	Start(ctx context.Context) error
	Stop() error
	ProcessPacket(ctx context.Context, packet *domain.MarketDataPacket) error
	GetConnectionCount() int
	GetProcessedCount() int64
}

// AnalysisService defines the interface for market analysis
type AnalysisService interface {
	// Pattern detection
	DetectWashTrading(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.WashTradeCandidate, error)
	DetectSpoofing(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.SpoofingIndicator, error)
	DetectPriceAnomalies(ctx context.Context, exchangeID uuid.UUID, symbol string, currentPrice decimal.Decimal) (*domain.PriceAnomaly, error)
	DetectVolumeAnomalies(ctx context.Context, exchangeID uuid.UUID, symbol string, currentVolume decimal.Decimal) (*domain.VolumeAnomaly, error)

	// Real-time analysis
	AnalyzeEvent(ctx context.Context, event *domain.MarketEvent) ([]domain.Alert, error)
	AnalyzeTrade(ctx context.Context, trade *domain.MarketEvent) ([]domain.Alert, error)
	AnalyzeOrder(ctx context.Context, order *domain.MarketEvent) ([]domain.Alert, error)

	// Analysis control
	ConfigureThresholds(ctx context.Context, thresholds map[domain.AlertType]map[string]any) error
	GetThresholds(ctx context.Context) (map[domain.AlertType]domain.AlertThreshold, error)
}

// AlertService defines the interface for alert management
type AlertService interface {
	// Alert lifecycle
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error)
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	ResolveAlert(ctx context.Context, id uuid.UUID, resolution string, resolvedBy uuid.UUID) error
	DismissAlert(ctx context.Context, id uuid.UUID, reason string, dismissedBy uuid.UUID) error
	EscalateAlert(ctx context.Context, id uuid.UUID, reason string, escalatedTo uuid.UUID) error

	// Alert assignment
	AssignAlert(ctx context.Context, id uuid.UUID, assigneeID uuid.UUID) error
	AddNote(ctx context.Context, alertID uuid.UUID, note *domain.AlertNote) error

	// Alert queries
	ListAlerts(ctx context.Context, filter domain.AlertFilter, limit, offset int) ([]domain.Alert, error)
	GetOpenAlerts(ctx context.Context) ([]domain.Alert, error)
	GetAlertsByExchange(ctx context.Context, exchangeID uuid.UUID, limit, offset int) ([]domain.Alert, error)
	GetAlertsBySymbol(ctx context.Context, symbol string, limit, offset int) ([]domain.Alert, error)

	// Statistics
	GetStatistics(ctx context.Context, startTime, endTime time.Time) (*domain.AlertStatistics, error)
	GetTrend(ctx context.Context, interval string, startTime, endTime time.Time) ([]domain.AlertTrendPoint, error)

	// Threshold management
	GetThresholds(ctx context.Context) (map[domain.AlertType]domain.AlertThreshold, error)
	UpdateThreshold(ctx context.Context, threshold *domain.AlertThreshold) error
}

// ComplianceService defines the interface for compliance verification
type ComplianceService interface {
	// Entity verification
	IsExchangeLicensed(ctx context.Context, exchangeID uuid.UUID) (bool, error)
	GetExchangeLicense(ctx context.Context, exchangeID uuid.UUID) (*ExchangeLicense, error)
	GetAllLicensedExchanges(ctx context.Context) ([]ExchangeLicense, error)

	// Account verification
	IsAccountVerified(ctx context.Context, accountID string) (bool, error)
	GetAccountVerification(ctx context.Context, accountID string) (*AccountVerification, error)

	// Caching
	InvalidateCache(ctx context.Context) error
	RefreshExchangeCache(ctx context.Context, exchangeID uuid.UUID) error
}

// ExchangeLicense represents a license for a cryptocurrency exchange
type ExchangeLicense struct {
	ExchangeID      uuid.UUID     `json:"exchange_id"`
	ExchangeName    string        `json:"exchange_name"`
	LicenseNumber   string        `json:"license_number"`
	LicenseType     string        `json:"license_type"`
	Status          LicenseStatus `json:"status"`
	IssuedAt        time.Time     `json:"issued_at"`
	ExpiresAt       time.Time     `json:"expires_at"`
	Jurisdiction    string        `json:"jurisdiction"`
	AllowedSymbols  []string      `json:"allowed_symbols"`
	Restrictions    []string      `json:"restrictions"`
	Conditions      []string      `json:"conditions"`
}

// LicenseStatus represents the status of a license
type LicenseStatus string

const (
	LicenseStatusActive    LicenseStatus = "ACTIVE"
	LicenseStatusSuspended LicenseStatus = "SUSPENDED"
	LicenseStatusRevoked   LicenseStatus = "REVOKED"
	LicenseStatusExpired   LicenseStatus = "EXPIRED"
	LicenseStatusPending   LicenseStatus = "PENDING"
)

// AccountVerification represents the verification status of an account
type AccountVerification struct {
	AccountID      uuid.UUID `json:"account_id"`
	Verified       bool      `json:"verified"`
	VerificationID string    `json:"verification_id"`
	Tier           int       `json:"tier"`
	KYCStatus      string    `json:"kyc_status"`
	AMLStatus      string    `json:"aml_status"`
	RiskScore      float64   `json:"risk_score"`
	LastVerifiedAt time.Time `json:"last_verified_at"`
}

// Import decimal for the interface
import "github.com/shopspring/decimal"
