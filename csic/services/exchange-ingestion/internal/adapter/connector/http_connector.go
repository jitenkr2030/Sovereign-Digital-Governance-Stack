package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// HTTPConnector implements ExchangeConnector for REST API-based exchanges
type HTTPConnector struct {
	config     *domain.DataSourceConfig
	logger     *zap.Logger
	httpClient *http.Client
	connected  bool
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// NewHTTPConnector creates a new HTTPConnector
func NewHTTPConnector(config *domain.DataSourceConfig, logger *zap.Logger) *HTTPConnector {
	return &HTTPConnector{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		stopChan: make(chan struct{}),
	}
}

// Connect establishes a connection to the exchange REST API
func (c *HTTPConnector) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	c.logger.Info("Connecting to exchange REST API",
		zap.String("source_id", c.config.ID.String()),
		zap.String("endpoint", c.config.Endpoint))

	// Test connection by making a simple request
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers if configured
	if c.config.AuthType == domain.AuthTypeAPIKey {
		req.Header.Set("X-MBX-APIKEY", c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	c.connected = true
	c.logger.Info("Successfully connected to exchange REST API",
		zap.String("source_id", c.config.ID.String()))

	return nil
}

// Disconnect closes the connection to the exchange
func (c *HTTPConnector) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.logger.Info("Disconnecting from exchange REST API",
		zap.String("source_id", c.config.ID.String()))

	close(c.stopChan)
	c.connected = false

	return nil
}

// IsConnected returns whether the connector is currently connected
func (c *HTTPConnector) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetSourceID returns the unique identifier for this data source
func (c *HTTPConnector) GetSourceID() string {
	return c.config.ID.String()
}

// GetExchangeType returns the type of exchange this connector supports
func (c *HTTPConnector) GetExchangeType() domain.ExchangeType {
	return c.config.ExchangeType
}

// StreamData starts polling for market data and sends it through the provided channel
func (c *HTTPConnector) StreamData(ctx context.Context, dataChan chan<- domain.MarketData) error {
	c.logger.Info("Starting HTTP polling for market data",
		zap.String("source_id", c.config.ID.String()),
		zap.Duration("polling_rate", c.config.PollingRate),
		zap.Strings("symbols", c.config.Symbols))

	ticker := time.NewTicker(c.config.PollingRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping HTTP polling",
				zap.String("source_id", c.config.ID.String()))
			return nil
		case <-c.stopChan:
			c.logger.Info("Stop signal received, stopping HTTP polling",
				zap.String("source_id", c.config.ID.String()))
			return nil
		case <-ticker.C:
			for _, symbol := range c.config.Symbols {
				data, err := c.fetchTickerData(ctx, symbol)
				if err != nil {
					c.logger.Error("Failed to fetch ticker data",
						zap.String("symbol", symbol),
						zap.Error(err))
					continue
				}

				select {
				case dataChan <- *data:
				default:
					c.logger.Warn("Data channel full, dropping message",
						zap.String("symbol", symbol))
				}
			}
		}
	}
}

// fetchTickerData fetches ticker data for a single symbol
func (c *HTTPConnector) fetchTickerData(ctx context.Context, symbol string) (*domain.MarketData, error) {
	// Construct API endpoint based on exchange type
	url := c.buildTickerURL(symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.config.AuthType == domain.AuthTypeAPIKey {
		req.Header.Set("X-MBX-APIKEY", c.config.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, domain.ErrRateLimited
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	// Parse response based on exchange type
	var tickerData TickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&tickerData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertToMarketData(&tickerData, symbol), nil
}

// buildTickerURL constructs the ticker API URL based on exchange type
func (c *HTTPConnector) buildTickerURL(symbol string) string {
	switch c.config.ExchangeType {
	case domain.ExchangeTypeBinance:
		return fmt.Sprintf("%s/api/v3/ticker/24hr?symbol=%s", c.config.Endpoint, symbol)
	case domain.ExchangeTypeCoinbase:
		return fmt.Sprintf("%s/products/%s/ticker", c.config.Endpoint, strings.ReplaceAll(symbol, "-", "-"))
	case domain.ExchangeTypeKraken:
		return fmt.Sprintf("%s/0/public/Ticker?pair=%s", c.config.Endpoint, symbol)
	default:
		return fmt.Sprintf("%s/ticker?symbol=%s", c.config.Endpoint, symbol)
	}
}

// convertToMarketData converts API response to domain MarketData
func (c *HTTPConnector) convertToMarketData(ticker *TickerResponse, symbol string) *domain.MarketData {
	price, _ := decimal.NewFromString(ticker.Price)
	volume, _ := decimal.NewFromString(ticker.Volume)
	high24h, _ := decimal.NewFromString(ticker.HighPrice)
	low24h, _ := decimal.NewFromString(ticker.LowPrice)

	var change24h decimal.Decimal
	if price.IsPositive() && high24h.IsPositive() {
		change24h = price.Sub(low24h).Div(high24h).Mul(decimal.NewFromInt(100))
	}

	return &domain.MarketData{
		ID:          uuid.New(),
		SourceID:    c.config.ID.String(),
		Symbol:      symbol,
		BaseSymbol:  extractBaseSymbol(symbol),
		QuoteSymbol: extractQuoteSymbol(symbol),
		Price:       price,
		Volume:      volume,
		High24h:     high24h,
		Low24h:      low24h,
		Change24h:   change24h,
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}
}

// FetchSnapshot fetches a snapshot of current market data for specific symbols
func (c *HTTPConnector) FetchSnapshot(ctx context.Context, symbols []string) ([]*domain.MarketData, error) {
	if symbols == nil {
		symbols = c.config.Symbols
	}

	c.logger.Debug("Fetching market data snapshot",
		zap.String("source_id", c.config.ID.String()),
		zap.Strings("symbols", symbols))

	dataList := make([]*domain.MarketData, 0, len(symbols))
	for _, symbol := range symbols {
		data, err := c.fetchTickerData(ctx, symbol)
		if err != nil {
			c.logger.Warn("Failed to fetch data for symbol",
				zap.String("symbol", symbol),
				zap.Error(err))
			continue
		}
		dataList = append(dataList, data)
	}

	return dataList, nil
}

// SubscribeTrades is not implemented for HTTP connector
func (c *HTTPConnector) SubscribeTrades(ctx context.Context, symbols []string, tradeChan chan<- domain.Trade) error {
	c.logger.Info("Trade subscription not supported for HTTP connector, use polling instead")
	return nil
}

// SubscribeOrderBook is not implemented for HTTP connector
func (c *HTTPConnector) SubscribeOrderBook(ctx context.Context, symbols []string, orderbookChan chan<- domain.OrderBook) error {
	c.logger.Info("Order book subscription not supported for HTTP connector")
	return nil
}

// GetSupportedSymbols returns the list of symbols supported by this connector
func (c *HTTPConnector) GetSupportedSymbols() []string {
	return []string{
		"BTCUSD", "ETHUSD", "BTCETH", "ETHEUR",
		"LINKUSD", "SOLUSD", "ADAUSD", "DOTUSD",
	}
}

// ValidateConfig validates the configuration for this connector
func (c *HTTPConnector) ValidateConfig() error {
	if c.config.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if len(c.config.Symbols) == 0 {
		return fmt.Errorf("at least one symbol is required")
	}
	if c.config.Timeout == 0 {
		c.config.Timeout = 30 * time.Second
	}
	if c.config.PollingRate == 0 {
		c.config.PollingRate = 5 * time.Second
	}
	return nil
}

// TickerResponse represents the common ticker response structure
type TickerResponse struct {
	Symbol      string `json:"symbol"`
	Price       string `json:"price"`
	Volume      string `json:"volume"`
	HighPrice   string `json:"highPrice"`
	LowPrice    string `json:"lowPrice"`
	Change24h   string `json:"priceChangePercent"`
	LastPrice   string `json:"lastPrice"`
	BidPrice    string `json:"bidPrice"`
	AskPrice    string `json:"askPrice"`
}

// Ensure HTTPConnector implements ExchangeConnector
var _ ports.ExchangeConnector = (*HTTPConnector)(nil)
