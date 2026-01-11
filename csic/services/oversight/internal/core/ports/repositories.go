package ports

import (
	"context"
	"time"

	"github.com/csic/oversight/internal/core/domain"
)

// OversightRepository defines the database operations for oversight
type OversightRepository interface {
	// Exchange operations
	CreateExchange(ctx context.Context, exchange *domain.Exchange) error
	GetExchange(ctx context.Context, id string) (*domain.Exchange, error)
	UpdateExchangeStatus(ctx context.Context, id string, status domain.ExchangeStatus) error
	ListExchanges(ctx context.Context) ([]*domain.Exchange, error)

	// Health metrics operations
	RecordHealthMetrics(ctx context.Context, metrics *domain.ExchangeHealthMetrics) error
	GetHealthMetrics(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.ExchangeHealthMetrics, error)
	CalculateAverageLatency(ctx context.Context, exchangeID string, window time.Duration) (int, error)

	// Anomaly operations
	RecordAnomaly(ctx context.Context, anomaly *domain.TradeAnomaly) error
	GetAnomalies(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.TradeAnomaly, error)
	UpdateAnomalyStatus(ctx context.Context, id string, status string) error
	GetPendingAnomalies(ctx context.Context) ([]*domain.TradeAnomaly, error)

	// Market data operations
	RecordTrade(ctx context.Context, trade *domain.Trade) error
	GetRecentTrades(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.Trade, error)
	RecordMarketDepth(ctx context.Context, depth *domain.MarketDepth) error
	GetMarketDepth(ctx context.Context, exchangeID, symbol string, timestamp time.Time) (*domain.MarketDepth, error)
}

// OversightEventPort defines the event publishing operations
type OversightEventPort interface {
	PublishAlertEvent(ctx context.Context, alertType string, payload map[string]interface{}) error
	PublishAnomalyDetected(ctx context.Context, anomaly *domain.TradeAnomaly) error
	PublishHealthAlert(ctx context.Context, exchangeID string, healthScore float64) error
}

// OversightDomainService defines the business logic operations for oversight
type OversightDomainService interface {
	// Exchange health monitoring
	RegisterExchange(ctx context.Context, exchange *domain.Exchange) error
	GetExchangeHealth(ctx context.Context, id string) (*domain.Exchange, error)
	CalculateHealthScore(ctx context.Context, exchangeID string) (float64, error)
	UpdateExchangeStatus(ctx context.Context, id string, status domain.ExchangeStatus) error

	// Anomaly detection
	DetectWashTrading(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.TradeAnomaly, error)
	DetectSpoofing(ctx context.Context, exchangeID, symbol string) ([]*domain.TradeAnomaly, error)
	DetectPumpAndDump(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.TradeAnomaly, error)
	GetAnomalies(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.TradeAnomaly, error)
	ResolveAnomaly(ctx context.Context, id string, resolution string) error

	// Market data processing
	ProcessTrade(ctx context.Context, trade *domain.Trade) error
	ProcessMarketDepth(ctx context.Context, depth *domain.MarketDepth) error
	GetMarketOverview(ctx context.Context) (*MarketOverview, error)
}

// MarketOverview provides a summary of all exchange activities
type MarketOverview struct {
	TotalExchanges      int                    `json:"total_exchanges"`
	OnlineExchanges     int                    `json:"online_exchanges"`
	AverageHealthScore  float64                `json:"average_health_score"`
	PendingAnomalies    int                    `json:"pending_anomalies"`
	RecentAnomalies     []*domain.TradeAnomaly `json:"recent_anomalies"`
	TopRiskSymbols      []RiskSymbol           `json:"top_risk_symbols"`
	LastUpdated         time.Time              `json:"last_updated"`
}

// RiskSymbol represents a symbol with elevated risk
type RiskSymbol struct {
	Symbol          string  `json:"symbol"`
	ExchangeID      string  `json:"exchange_id"`
	AnomalyCount    int     `json:"anomaly_count"`
	RiskScore       float64 `json:"risk_score"`
}
