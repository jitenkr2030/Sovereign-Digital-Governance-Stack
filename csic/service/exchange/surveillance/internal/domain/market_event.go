package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MarketEventType represents the type of market event
type MarketEventType string

const (
	MarketEventTypeTrade   MarketEventType = "TRADE"
	MarketEventTypeOrder   MarketEventType = "ORDER"
	MarketEventTypeQuote   MarketEventType = "QUOTE"
)

// MarketEventDirection represents the direction of an order
type MarketEventDirection string

const (
	MarketEventDirectionBuy  MarketEventDirection = "BUY"
	MarketEventDirectionSell MarketEventDirection = "SELL"
)

// MarketEventStatus represents the status of an order event
type MarketEventStatus string

const (
	MarketEventStatusNew      MarketEventStatus = "NEW"
	MarketEventStatusPartial  MarketEventStatus = "PARTIAL"
	MarketEventStatusFilled   MarketEventStatus = "FILLED"
	MarketEventStatusCancelled MarketEventStatus = "CANCELLED"
	MarketEventStatusRejected MarketEventStatus = "REJECTED"
)

// MarketEvent represents a single market data event from an exchange
type MarketEvent struct {
	ID            uuid.UUID            `json:"id" db:"id"`
	ExchangeID    uuid.UUID            `json:"exchange_id" db:"exchange_id"`
	Symbol        string               `json:"symbol" db:"symbol"`
	EventType     MarketEventType      `json:"event_type" db:"event_type"`
	OrderID       *uuid.UUID           `json:"order_id,omitempty" db:"order_id"`
	UserID        *string              `json:"user_id,omitempty" db:"user_id"`
	Price         decimal.Decimal      `json:"price" db:"price"`
	Quantity      decimal.Decimal      `json:"quantity" db:"quantity"`
	FilledQuantity decimal.Decimal     `json:"filled_quantity,omitempty" db:"filled_quantity"`
	Direction     MarketEventDirection `json:"direction" db:"direction"`
	Status        MarketEventStatus    `json:"status,omitempty" db:"status"`
	Timestamp     time.Time            `json:"timestamp" db:"timestamp"`
	ReceivedAt    time.Time            `json:"received_at" db:"received_at"`
	SequenceNum   int64                `json:"sequence_num" db:"sequence_num"`
	RawData       string               `json:"raw_data,omitempty" db:"raw_data"`
	Checksum      string               `json:"checksum" db:"checksum"`
}

// MarketSummary represents a summary of market activity for a symbol
type MarketSummary struct {
	ExchangeID       uuid.UUID       `json:"exchange_id"`
	Symbol           string          `json:"symbol"`
	PeriodStart      time.Time       `json:"period_start"`
	PeriodEnd        time.Time       `json:"period_end"`
	OpenPrice        decimal.Decimal `json:"open_price"`
	HighPrice        decimal.Decimal `json:"high_price"`
	LowPrice         decimal.Decimal `json:"low_price"`
	ClosePrice       decimal.Decimal `json:"close_price"`
	Volume           decimal.Decimal `json:"volume"`
	TradeCount       int64           `json:"trade_count"`
	AvgTradeSize     decimal.Decimal `json:"avg_trade_size"`
	VWAP             decimal.Decimal `json:"vwap"`
	PriceVolatility  decimal.Decimal `json:"price_volatility"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// MarketStats represents statistical metrics for market analysis
type MarketStats struct {
	Symbol            string          `json:"symbol"`
	ExchangeID        uuid.UUID       `json:"exchange_id"`
	MovingAverage     decimal.Decimal `json:"moving_average"`
	StandardDeviation decimal.Decimal `json:"standard_deviation"`
	VolumeMA          decimal.Decimal `json:"volume_ma"`
	VolumeStdDev      decimal.Decimal `json:"volume_std_dev"`
	TypicalPrice      decimal.Decimal `json:"typical_price"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// WashTradeCandidate represents a potential wash trade pattern
type WashTradeCandidate struct {
	ID              uuid.UUID         `json:"id"`
	ExchangeID      uuid.UUID         `json:"exchange_id"`
	Symbol          string            `json:"symbol"`
	AccountIDs      []string          `json:"account_ids"`
	TradeIDs        []uuid.UUID       `json:"trade_ids"`
	TimeWindow      time.Duration     `json:"time_window"`
	TotalVolume     decimal.Decimal   `json:"total_volume"`
	PriceDeviation  decimal.Decimal   `json:"price_deviation"`
	Confidence      float64           `json:"confidence"`
	DetectedAt      time.Time         `json:"detected_at"`
	Evidence        []MarketEvent     `json:"evidence"`
}

// SpoofingIndicator represents potential spoofing behavior
type SpoofingIndicator struct {
	ID              uuid.UUID       `json:"id"`
	ExchangeID      uuid.UUID       `json:"exchange_id"`
	Symbol          string          `json:"symbol"`
	OrderID         uuid.UUID       `json:"order_id"`
	AccountID       string          `json:"account_id"`
	OrderSize       decimal.Decimal `json:"order_size"`
	OrderPrice      decimal.Decimal `json:"order_price"`
	OrderDirection  MarketEventDirection `json:"order_direction"`
	Lifetime        time.Duration   `json:"lifetime"`
	CancelledAmount decimal.Decimal `json:"cancelled_amount"`
	ExecutedAmount  decimal.Decimal `json:"executed_amount"`
	CancellationRate float64        `json:"cancellation_rate"`
	Confidence      float64         `json:"confidence"`
	PatternType     string          `json:"pattern_type"` // "BAIT_LAYING", "LAYER_ADDING"
	DetectedAt      time.Time       `json:"detected_at"`
	Evidence        []MarketEvent   `json:"evidence"`
}

// PriceAnomaly represents a detected price anomaly
type PriceAnomaly struct {
	ID              uuid.UUID       `json:"id"`
	ExchangeID      uuid.UUID       `json:"exchange_id"`
	Symbol          string          `json:"symbol"`
	CurrentPrice    decimal.Decimal `json:"current_price"`
	GlobalAverage   decimal.Decimal `json:"global_average"`
	DeviationPct    decimal.Decimal `json:"deviation_pct"`
	ThresholdPct    decimal.Decimal `json:"threshold_pct"`
	Direction       string          `json:"direction"` // "ABOVE" or "BELOW"
	Confidence      float64         `json:"confidence"`
	DetectedAt      time.Time       `json:"detected_at"`
}

// VolumeAnomaly represents detected volume anomalies
type VolumeAnomaly struct {
	ID              uuid.UUID       `json:"id"`
	ExchangeID      uuid.UUID       `json:"exchange_id"`
	Symbol          string          `json:"symbol"`
	CurrentVolume   decimal.Decimal `json:"current_volume"`
	AverageVolume   decimal.Decimal `json:"average_volume"`
	StdDev          decimal.Decimal `json:"std_dev"`
	ZScore          decimal.Decimal `json:"z_score"`
	Threshold       float64         `json:"threshold"`
	Direction       string          `json:"direction"` // "SPIKE" or "DROP"
	Confidence      float64         `json:"confidence"`
	DetectedAt      time.Time       `json:"detected_at"`
}

// OrderBookSnapshot represents a snapshot of the order book
type OrderBookSnapshot struct {
	ID            uuid.UUID         `json:"id"`
	ExchangeID    uuid.UUID         `json:"exchange_id"`
	Symbol        string            `json:"symbol"`
	Timestamp     time.Time         `json:"timestamp"`
	Bids          []OrderBookLevel  `json:"bids"`
	Asks          []OrderBookLevel  `json:"asks"`
	Checksum      string            `json:"checksum"`
}

// OrderBookLevel represents a single level in the order book
type OrderBookLevel struct {
	Price     decimal.Decimal `json:"price"`
	Quantity  decimal.Decimal `json:"quantity"`
	OrderCount int           `json:"order_count"`
}

// MarketDataPacket represents an incoming market data packet
type MarketDataPacket struct {
	ExchangeID   uuid.UUID            `json:"exchange_id"`
	Symbol       string               `json:"symbol"`
	PacketType   string               `json:"packet_type"` // "trade", "order", "book"
	Trades       []TradeData          `json:"trades,omitempty"`
	Orders       []OrderData          `json:"orders,omitempty"`
	OrderBook    *OrderBookSnapshot   `json:"order_book,omitempty"`
	Timestamp    time.Time            `json:"timestamp"`
	SequenceNum  int64                `json:"sequence_num"`
}

// TradeData represents trade information in a packet
type TradeData struct {
	TradeID        uuid.UUID       `json:"trade_id"`
	OrderID        uuid.UUID       `json:"order_id"`
	UserID         string          `json:"user_id"`
	Price          decimal.Decimal `json:"price"`
	Quantity       decimal.Decimal `json:"quantity"`
	Direction      string          `json:"direction"`
	Timestamp      time.Time       `json:"timestamp"`
	MakerOrderID   *uuid.UUID      `json:"maker_order_id,omitempty"`
	TakerOrderID   *uuid.UUID      `json:"taker_order_id,omitempty"`
}

// OrderData represents order information in a packet
type OrderData struct {
	OrderID        uuid.UUID           `json:"order_id"`
	UserID         string              `json:"user_id"`
	Price          decimal.Decimal     `json:"price"`
	Quantity       decimal.Decimal     `json:"quantity"`
	FilledQuantity decimal.Decimal     `json:"filled_quantity"`
	Direction      string              `json:"direction"`
	Status         MarketEventStatus   `json:"status"`
	Timestamp      time.Time           `json:"timestamp"`
}
