package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MarketData represents normalized market data from any exchange
type MarketData struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	SourceID    string          `json:"source_id" db:"source_id"`
	Symbol      string          `json:"symbol" db:"symbol"`
	BaseSymbol  string          `json:"base_symbol" db:"base_symbol"`
	QuoteSymbol string          `json:"quote_symbol" db:"quote_symbol"`
	Price       decimal.Decimal `json:"price" db:"price"`
	Volume      decimal.Decimal `json:"volume" db:"volume"`
	High24h     decimal.Decimal `json:"high_24h" db:"high_24h"`
	Low24h      decimal.Decimal `json:"low_24h" db:"low_24h"`
	Change24h   decimal.Decimal `json:"change_24h" db:"change_24h"`
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// Trade represents a single trade execution from an exchange
type Trade struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	SourceID      string          `json:"source_id" db:"source_id"`
	Symbol        string          `json:"symbol" db:"symbol"`
	TradeID       string          `json:"trade_id" db:"trade_id"`
	Price         decimal.Decimal `json:"price" db:"price"`
	Quantity      decimal.Decimal `json:"quantity" db:"quantity"`
	QuoteQuantity decimal.Decimal `json:"quote_quantity" db:"quote_quantity"`
	Side          TradeSide       `json:"side" db:"side"`
	Timestamp     time.Time       `json:"timestamp" db:"timestamp"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// TradeSide represents the direction of a trade
type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

// OrderBook represents the order book snapshot for a symbol
type OrderBook struct {
	ID              uuid.UUID            `json:"id" db:"id"`
	SourceID        string               `json:"source_id" db:"source_id"`
	Symbol          string               `json:"symbol" db:"symbol"`
	Bids            []OrderBookEntry     `json:"bids" db:"-"`
	Asks            []OrderBookEntry     `json:"asks" db:"-"`
	BidDepth        decimal.Decimal      `json:"bid_depth" db:"bid_depth"`
	AskDepth        decimal.Decimal      `json:"ask_depth" db:"ask_depth"`
	Spread          decimal.Decimal      `json:"spread" db:"spread"`
	Timestamp       time.Time            `json:"timestamp" db:"timestamp"`
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
}

// OrderBookEntry represents a single order book entry
type OrderBookEntry struct {
	Price       decimal.Decimal `json:"price"`
	Quantity    decimal.Decimal `json:"quantity"`
	OrderCount  int             `json:"order_count"`
}

// DataSourceConfig represents configuration for an exchange connection
type DataSourceConfig struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	ExchangeType  ExchangeType    `json:"exchange_type" db:"exchange_type"`
	Endpoint      string          `json:"endpoint" db:"endpoint"`
	APIKey        string          `json:"-" db:"api_key"`
	APISecret     string          `json:"-" db:"api_secret"`
	Symbols       []string        `json:"symbols" db:"symbols"`
	Enabled       bool            `json:"enabled" db:"enabled"`
	PollingRate   time.Duration   `json:"polling_rate" db:"polling_rate"`
	WSEnabled     bool            `json:"ws_enabled" db:"ws_enabled"`
	AuthType      AuthType        `json:"auth_type" db:"auth_type"`
	Timeout       time.Duration   `json:"timeout" db:"timeout"`
	RetryAttempts int             `json:"retry_attempts" db:"retry_attempts"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// ExchangeType represents the type of exchange
type ExchangeType string

const (
	ExchangeTypeBinance   ExchangeType = "BINANCE"
	ExchangeTypeCoinbase  ExchangeType = "COINBASE"
	ExchangeTypeKraken    ExchangeType = "KRAKEN"
	ExchangeTypeCME       ExchangeType = "CME"
	ExchangeTypeGeneric   ExchangeType = "GENERIC"
	ExchangeTypeMock      ExchangeType = "MOCK"
)

// AuthType represents the authentication type required
type AuthType string

const (
	AuthTypeNone     AuthType = "NONE"
	AuthTypeAPIKey   AuthType = "API_KEY"
	AuthTypeSigned   AuthType = "SIGNED"
)

// IngestionStats tracks the statistics of data ingestion
type IngestionStats struct {
	SourceID          string          `json:"source_id"`
	TotalMessages     int64           `json:"total_messages"`
	ProcessedMessages int64           `json:"processed_messages"`
	FailedMessages    int64           `json:"failed_messages"`
	LastMessageTime   *time.Time      `json:"last_message_time"`
	AverageLatency    time.Duration   `json:"average_latency"`
	LastError         string          `json:"last_error"`
	Status            ConnectionStatus `json:"status"`
}

// ConnectionStatus represents the status of a connection
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "CONNECTED"
	ConnectionStatusDisconnected ConnectionStatus = "DISCONNECTED"
	ConnectionStatusConnecting   ConnectionStatus = "CONNECTING"
	ConnectionStatusError        ConnectionStatus = "ERROR"
)
