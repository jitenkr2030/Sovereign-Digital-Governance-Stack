package ports

import (
	"context"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
)

// ExchangeConnector defines the interface for connecting to exchanges
// This is the primary port for external system integration
type ExchangeConnector interface {
	// Connect establishes a connection to the exchange
	Connect(ctx context.Context) error

	// Disconnect closes the connection to the exchange
	Disconnect(ctx context.Context) error

	// IsConnected returns whether the connector is currently connected
	IsConnected() bool

	// GetSourceID returns the unique identifier for this data source
	GetSourceID() string

	// GetExchangeType returns the type of exchange this connector supports
	GetExchangeType() domain.ExchangeType

	// StreamData starts streaming market data and sends it through the provided channel
	// The connector should handle reconnection and error handling internally
	StreamData(ctx context.Context, dataChan chan<- domain.MarketData) error

	// FetchSnapshot fetches a snapshot of current market data for specific symbols
	FetchSnapshot(ctx context.Context, symbols []string) ([]*domain.MarketData, error)

	// SubscribeTrades subscribes to trade updates for specific symbols
	SubscribeTrades(ctx context.Context, symbols []string, tradeChan chan<- domain.Trade) error

	// SubscribeOrderBook subscribes to order book updates for specific symbols
	SubscribeOrderBook(ctx context.Context, symbols []string, orderbookChan chan<- domain.OrderBook) error

	// GetSupportedSymbols returns the list of symbols supported by this connector
	GetSupportedSymbols() []string

	// ValidateConfig validates the configuration for this connector
	ValidateConfig() error
}

// ExchangeConnectorFactory creates ExchangeConnector instances based on exchange type
type ExchangeConnectorFactory interface {
	// CreateConnector creates a new connector for the specified exchange type
	CreateConnector(config *domain.DataSourceConfig) (ExchangeConnector, error)

	// GetSupportedExchangeTypes returns the list of supported exchange types
	GetSupportedExchangeTypes() []domain.ExchangeType
}

// EventPublisher defines the interface for publishing events to message brokers
// This enables communication with other services like Risk Engine and Reporting
type EventPublisher interface {
	// PublishMarketData publishes market data to the message broker
	PublishMarketData(ctx context.Context, data *domain.MarketData) error

	// PublishTrade publishes trade data to the message broker
	PublishTrade(ctx context.Context, trade *domain.Trade) error

	// PublishOrderBook publishes order book data to the message broker
	PublishOrderBook(ctx context.Context, orderbook *domain.OrderBook) error

	// PublishBatch publishes multiple market data records in a batch
	PublishBatch(ctx context.Context, data []*domain.MarketData) error

	// Close closes the publisher connection
	Close() error
}

// EventConsumer defines the interface for consuming events from message brokers
type EventConsumer interface {
	// ConsumeMarketData starts consuming market data from the message broker
	ConsumeMarketData(ctx context.Context, handler func(data *domain.MarketData) error) error

	// Start starts the consumer
	Start(ctx context.Context) error

	// Stop stops the consumer
	Stop() error
}

// HealthChecker defines the interface for checking the health of external dependencies
type HealthChecker interface {
	// CheckEndpoint checks if the exchange endpoint is reachable
	CheckEndpoint(ctx context.Context, endpoint string) error

	// CheckAuth checks if the authentication credentials are valid
	CheckAuth(ctx context.Context, config *domain.DataSourceConfig) error

	// GetEndpointHealth returns health information about the endpoint
	GetEndpointHealth(ctx context.Context, config *domain.DataSourceConfig) (*EndpointHealth, error)
}

// EndpointHealth represents the health status of an endpoint
type EndpointHealth struct {
	Endpoint   string `json:"endpoint"`
	Status     string `json:"status"`
	Latency    string `json:"latency"`
	LastCheck  string `json:"last_check"`
	ErrorCount int    `json:"error_count"`
}
