package connector

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MockConnector implements ExchangeConnector for testing and development
type MockConnector struct {
	config     *domain.DataSourceConfig
	logger     *zap.Logger
	connected  bool
	mu         sync.RWMutex
	dataChan   chan domain.MarketData
	stopChan   chan struct{}
}

// NewMockConnector creates a new MockConnector
func NewMockConnector(config *domain.DataSourceConfig, logger *zap.Logger) *MockConnector {
	return &MockConnector{
		config:   config,
		logger:   logger,
		dataChan: make(chan domain.MarketData, 1000),
		stopChan: make(chan struct{}),
	}
}

// Connect establishes a connection to the mock exchange
func (c *MockConnector) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	c.logger.Info("Connecting to mock exchange",
		zap.String("source_id", c.config.ID.String()),
		zap.Strings("symbols", c.config.Symbols))

	// Simulate connection delay
	time.Sleep(100 * time.Millisecond)

	c.connected = true
	c.logger.Info("Successfully connected to mock exchange",
		zap.String("source_id", c.config.ID.String()))

	return nil
}

// Disconnect closes the connection to the mock exchange
func (c *MockConnector) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.logger.Info("Disconnecting from mock exchange",
		zap.String("source_id", c.config.ID.String()))

	close(c.stopChan)
	c.connected = false

	return nil
}

// IsConnected returns whether the connector is currently connected
func (c *MockConnector) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetSourceID returns the unique identifier for this data source
func (c *MockConnector) GetSourceID() string {
	return c.config.ID.String()
}

// GetExchangeType returns the type of exchange this connector supports
func (c *MockConnector) GetExchangeType() domain.ExchangeType {
	return domain.ExchangeTypeMock
}

// StreamData starts streaming market data and sends it through the provided channel
func (c *MockConnector) StreamData(ctx context.Context, dataChan chan<- domain.MarketData) error {
	c.logger.Info("Starting mock data stream",
		zap.String("source_id", c.config.ID.String()),
		zap.Strings("symbols", c.config.Symbols))

	// Base prices for common trading pairs
	basePrices := map[string]decimal.Decimal{
		"BTC-USD":  decimal.NewFromFloat(45000.00),
		"ETH-USD":  decimal.NewFromFloat(2800.00),
		"BTC-ETH":  decimal.NewFromFloat(16.07),
		"ETH-EUR":  decimal.NewFromFloat(2600.00),
		"LINK-USD": decimal.NewFromFloat(15.50),
		"SOL-USD":  decimal.NewFromFloat(98.00),
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping mock data stream",
				zap.String("source_id", c.config.ID.String()))
			return nil
		case <-c.stopChan:
			c.logger.Info("Stop signal received, stopping mock data stream",
				zap.String("source_id", c.config.ID.String()))
			return nil
		case <-ticker.C:
			for _, symbol := range c.config.Symbols {
				basePrice, exists := basePrices[symbol]
				if !exists {
					basePrice = decimal.NewFromFloat(100.00)
				}

				// Add some random price movement
				volatility := decimal.NewFromFloat(0.002) // 0.2% volatility
				changePercent := decimal.NewFromFloat(rand.Float64()-0.5).Mul(volatility.Mul(decimal.NewFromInt(2)))
				price := basePrice.Add(basePrice.Mul(changePercent))

				volume := decimal.NewFromFloat(rand.Float64() * 10)

				data := domain.MarketData{
					ID:          uuid.New(),
					SourceID:    c.config.ID.String(),
					Symbol:      symbol,
					BaseSymbol:  extractBaseSymbol(symbol),
					QuoteSymbol: extractQuoteSymbol(symbol),
					Price:       price.Round(2),
					Volume:      volume.Round(4),
					High24h:     price.Mul(decimal.NewFromFloat(1.05)).Round(2),
					Low24h:      price.Mul(decimal.NewFromFloat(0.95)).Round(2),
					Change24h:   changePercent.Mul(decimal.NewFromInt(100)).Round(2),
					Timestamp:   time.Now(),
					CreatedAt:   time.Now(),
				}

				select {
				case dataChan <- data:
				default:
					c.logger.Warn("Data channel full, dropping message",
						zap.String("symbol", symbol))
				}
			}
		}
	}
}

// FetchSnapshot fetches a snapshot of current market data for specific symbols
func (c *MockConnector) FetchSnapshot(ctx context.Context, symbols []string) ([]*domain.MarketData, error) {
	if symbols == nil {
		symbols = c.config.Symbols
	}

	c.logger.Debug("Fetching mock snapshot",
		zap.String("source_id", c.config.ID.String()),
		zap.Strings("symbols", symbols))

	now := time.Now()
	dataList := make([]*domain.MarketData, 0, len(symbols))

	for _, symbol := range symbols {
		price := decimal.NewFromFloat(100.0 + rand.Float64()*100)
		data := &domain.MarketData{
			ID:          uuid.New(),
			SourceID:    c.config.ID.String(),
			Symbol:      symbol,
			BaseSymbol:  extractBaseSymbol(symbol),
			QuoteSymbol: extractQuoteSymbol(symbol),
			Price:       price.Round(2),
			Volume:      decimal.NewFromFloat(rand.Float64() * 100),
			Timestamp:   now,
			CreatedAt:   now,
		}
		dataList = append(dataList, data)
	}

	return dataList, nil
}

// SubscribeTrades is not implemented for mock connector
func (c *MockConnector) SubscribeTrades(ctx context.Context, symbols []string, tradeChan chan<- domain.Trade) error {
	c.logger.Info("Trade subscription not supported for mock connector")
	return nil
}

// SubscribeOrderBook is not implemented for mock connector
func (c *MockConnector) SubscribeOrderBook(ctx context.Context, symbols []string, orderbookChan chan<- domain.OrderBook) error {
	c.logger.Info("Order book subscription not supported for mock connector")
	return nil
}

// GetSupportedSymbols returns the list of symbols supported by this connector
func (c *MockConnector) GetSupportedSymbols() []string {
	return []string{
		"BTC-USD", "ETH-USD", "BTC-ETH", "ETH-EUR",
		"LINK-USD", "SOL-USD", "ADA-USD", "DOT-USD",
	}
}

// ValidateConfig validates the configuration for this connector
func (c *MockConnector) ValidateConfig() error {
	if len(c.config.Symbols) == 0 {
		return domain.ErrInvalidDataSourceConfig
	}
	return nil
}

// extractBaseSymbol extracts the base symbol from a trading pair
func extractBaseSymbol(pair string) string {
	for i := 0; i < len(pair)-1; i++ {
		if pair[i] == '-' {
			return pair[:i]
		}
	}
	return pair[:3]
}

// extractQuoteSymbol extracts the quote symbol from a trading pair
func extractQuoteSymbol(pair string) string {
	for i := 0; i < len(pair)-1; i++ {
		if pair[i] == '-' {
			return pair[i+1:]
		}
	}
	return "USD"
}

// Ensure MockConnector implements ExchangeConnector
var _ ports.ExchangeConnector = (*MockConnector)(nil)
