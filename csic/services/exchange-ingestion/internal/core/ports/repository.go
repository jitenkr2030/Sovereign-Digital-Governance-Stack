package ports

import (
	"context"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
)

// DataSourceRepository defines the interface for managing data source configurations
type DataSourceRepository interface {
	// FindByID retrieves a data source configuration by its unique identifier
	FindByID(ctx context.Context, id string) (*domain.DataSourceConfig, error)

	// FindByName retrieves a data source configuration by its name
	FindByName(ctx context.Context, name string) (*domain.DataSourceConfig, error)

	// FindAll retrieves all data source configurations
	FindAll(ctx context.Context) ([]*domain.DataSourceConfig, error)

	// FindEnabled retrieves all enabled data source configurations
	FindEnabled(ctx context.Context) ([]*domain.DataSourceConfig, error)

	// Save creates or updates a data source configuration
	Save(ctx context.Context, config *domain.DataSourceConfig) error

	// Delete removes a data source configuration
	Delete(ctx context.Context, id string) error

	// UpdateStatus updates the connection status of a data source
	UpdateStatus(ctx context.Context, id string, status domain.ConnectionStatus) error
}

// MarketDataRepository defines the interface for storing and retrieving market data
type MarketDataRepository interface {
	// Store saves market data to the repository
	Store(ctx context.Context, data *domain.MarketData) error

	// StoreBatch saves multiple market data records in a single operation
	StoreBatch(ctx context.Context, data []*domain.MarketData) error

	// FindLatest retrieves the latest market data for a specific symbol and source
	FindLatest(ctx context.Context, sourceID, symbol string) (*domain.MarketData, error)

	// FindByTimeRange retrieves market data within a specified time range
	FindByTimeRange(ctx context.Context, sourceID, symbol string, start, end time.Time) ([]*domain.MarketData, error)

	// FindLatestBySymbol retrieves the latest market data for a symbol across all sources
	FindLatestBySymbol(ctx context.Context, symbol string) ([]*domain.MarketData, error)
}

// TradeRepository defines the interface for storing and retrieving trade data
type TradeRepository interface {
	// Store saves trade data to the repository
	Store(ctx context.Context, trade *domain.Trade) error

	// StoreBatch saves multiple trade records in a single operation
	StoreBatch(ctx context.Context, trades []*domain.Trade) error

	// FindByTimeRange retrieves trade data within a specified time range
	FindByTimeRange(ctx context.Context, sourceID, symbol string, start, end time.Time) ([]*domain.Trade, error)

	// FindRecent retrieves the most recent trades for a symbol
	FindRecent(ctx context.Context, sourceID, symbol string, limit int) ([]*domain.Trade, error)
}

// OrderBookRepository defines the interface for storing and retrieving order book data
type OrderBookRepository interface {
	// Store saves order book data to the repository
	Store(ctx context.Context, orderbook *domain.OrderBook) error

	// FindLatest retrieves the latest order book for a specific symbol and source
	FindLatest(ctx context.Context, sourceID, symbol string) (*domain.OrderBook, error)
}

// IngestionStatsRepository defines the interface for managing ingestion statistics
type IngestionStatsRepository interface {
	// GetStats retrieves current statistics for a data source
	GetStats(ctx context.Context, sourceID string) (*domain.IngestionStats, error)

	// UpdateStats updates statistics for a data source
	UpdateStats(ctx context.Context, stats *domain.IngestionStats) error

	// IncrementMessages increments message counters
	IncrementMessages(ctx context.Context, sourceID string, processed, failed int64) error

	// RecordLatency records the processing latency
	RecordLatency(ctx context.Context, sourceID string, latency time.Duration) error
}
