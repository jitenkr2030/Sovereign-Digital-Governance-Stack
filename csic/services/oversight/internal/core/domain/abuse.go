package domain

import "time"

// Exchange represents a registered cryptocurrency exchange
type Exchange struct {
	ID          string         `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	APIEndpoint string         `json:"api_endpoint" db:"api_endpoint"`
	Status      ExchangeStatus `json:"status" db:"status"`
	HealthScore float64        `json:"health_score" db:"health_score"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// ExchangeHealthMetrics holds health monitoring data
type ExchangeHealthMetrics struct {
	ID              int64     `json:"id" db:"id"`
	ExchangeID      string    `json:"exchange_id" db:"exchange_id"`
	LatencyMs       int       `json:"latency_ms" db:"latency_ms"`
	UptimePercent   float64   `json:"uptime_percent" db:"uptime_percent"`
	LiquidityDepth  float64   `json:"liquidity_depth" db:"liquidity_depth"`
	ApiStatus       int       `json:"api_status" db:"api_status"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
}

// Trade represents a single trade event from an exchange
type Trade struct {
	ID          string    `json:"id"`
	ExchangeID  string    `json:"exchange_id"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"` // buy or sell
	Price       float64   `json:"price"`
	Quantity    float64   `json:"quantity"`
	Timestamp   time.Time `json:"timestamp"`
}

// AbuseType represents different types of market abuse
type AbuseType string

const (
	AbuseWashTrading   AbuseType = "wash_trading"
	AbuseSpoofing      AbuseType = "spoofing"
	AbusePumpDump      AbuseType = "pump_and_dump"
	AbuseFrontRunning  AbuseType = "front_running"
	AbuseLayering      AbuseType = "layering"
)

// TradeAnomaly represents a detected market abuse anomaly
type TradeAnomaly struct {
	ID             string    `json:"id" db:"id"`
	ExchangeID     string    `json:"exchange_id" db:"exchange_id"`
	Symbol         string    `json:"symbol" db:"symbol"`
	AbuseType      AbuseType `json:"abuse_type" db:"abuse_type"`
	ConfidenceScore float64  `json:"confidence_score" db:"confidence_score"`
	TradeIDs       []string  `json:"trade_ids"`
	Description    string    `json:"description" db:"description"`
	DetectedAt     time.Time `json:"detected_at" db:"detected_at"`
	Status         string    `json:"status" db:"status"` // pending, confirmed, dismissed
}

// MarketDepth represents order book liquidity
type MarketDepth struct {
	ExchangeID string             `json:"exchange_id"`
	Symbol     string             `json:"symbol"`
	Bids       []OrderBookLevel   `json:"bids"`
	Asks       []OrderBookLevel   `json:"asks"`
	Timestamp  time.Time          `json:"timestamp"`
}

// OrderBookLevel represents a single price level in the order book
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// HealthScoreWeights defines the weighting factors for health calculation
type HealthScoreWeights struct {
	UptimeWeight       float64 `json:"uptime_weight"`
	LatencyWeight      float64 `json:"latency_weight"`
	LiquidityWeight    float64 `json:"liquidity_weight"`
}

// DefaultHealthWeights returns the default health scoring weights
func DefaultHealthWeights() HealthScoreWeights {
	return HealthScoreWeights{
		UptimeWeight:    0.40,
		LatencyWeight:   0.30,
		LiquidityWeight: 0.30,
	}
}

// AnomalyDetectionConfig holds configuration for abuse detection algorithms
type AnomalyDetectionConfig struct {
	WashTradingWindowSecs int     `json:"wash_trading_window_secs"`
	SpoofingOrderThreshold float64 `json:"spoofing_order_threshold"`
	PumpDumpVolumeMultiplier float64 `json:"pump_dump_volume_multiplier"`
	ConfidenceThreshold float64    `json:"confidence_threshold"`
}

// DefaultAnomalyConfig returns default anomaly detection configuration
func DefaultAnomalyConfig() AnomalyDetectionConfig {
	return AnomalyDetectionConfig{
		WashTradingWindowSecs:   1,
		SpoofingOrderThreshold:  0.5,
		PumpDumpVolumeMultiplier: 5.0,
		ConfidenceThreshold:     0.75,
	}
}
