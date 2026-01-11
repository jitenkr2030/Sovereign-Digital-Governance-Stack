package ports

import (
	"context"
	"time"

	"github.com/csic/oversight/internal/core/domain"
)

// OversightService is the primary input port for the oversight service
// This interface is implemented by the core application services
type OversightService interface {
	// ProcessTradeStream processes incoming trade events from the message stream
	ProcessTradeStream(ctx context.Context, trade domain.TradeEvent) error

	// ProcessTradeBatch processes a batch of trade events
	ProcessTradeBatch(ctx context.Context, trades []domain.TradeEvent) error

	// GetExchangeHealth retrieves the current health status of an exchange
	GetExchangeHealth(ctx context.Context, exchangeID string) (*domain.ExchangeHealth, error)

	// GetAllExchangeHealth retrieves health status for all monitored exchanges
	GetAllExchangeHealth(ctx context.Context) ([]domain.ExchangeHealth, error)

	// RegisterExchange registers a new exchange
	RegisterExchange(ctx context.Context, exchange *domain.Exchange) error

	// CalculateHealthScore calculates the health score for an exchange
	CalculateHealthScore(ctx context.Context, exchangeID string) (float64, error)

	// UpdateExchangeStatus updates the status of an exchange
	UpdateExchangeStatus(ctx context.Context, id string, status domain.ExchangeStatus) error

	// ProcessTrade processes a trade
	ProcessTrade(ctx context.Context, trade *domain.Trade) error

	// ProcessMarketDepth processes market depth data
	ProcessMarketDepth(ctx context.Context, depth *domain.MarketDepth) error

	// GetMarketOverview provides a summary of all exchange activities
	GetMarketOverview(ctx context.Context) (*MarketOverview, error)

	// GetAnomalies retrieves anomalies for an exchange
	GetAnomalies(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.TradeAnomaly, error)

	// ResolveAnomaly resolves an anomaly
	ResolveAnomaly(ctx context.Context, id string, resolution string) error

	// DetectWashTrading detects wash trading patterns
	DetectWashTrading(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.TradeAnomaly, error)

	// DetectSpoofing detects spoofing patterns
	DetectSpoofing(ctx context.Context, exchangeID, symbol string) ([]*domain.TradeAnomaly, error)

	// UpdateDetectionRules updates the detection rules configuration
	UpdateDetectionRules(ctx context.Context, rules []domain.DetectionRule) error

	// GetDetectionRules retrieves all active detection rules
	GetDetectionRules(ctx context.Context) ([]domain.DetectionRule, error)

	// ExecuteThrottleCommand executes a throttle command for an exchange
	ExecuteThrottleCommand(ctx context.Context, cmd domain.ThrottleCommand) error

	// GenerateRegulatoryReport generates a regulatory report for the specified period
	GenerateRegulatoryReport(ctx context.Context, start, end time.Time, exchangeID string) (*domain.RegulatoryReport, error)
}

// AlertRepository is the output port for persisting alerts
type AlertRepository interface {
	// SaveAlert saves a market alert to the database
	SaveAlert(ctx context.Context, alert domain.MarketAlert) error

	// SaveAlertBatch saves multiple alerts in a single transaction
	SaveAlertBatch(ctx context.Context, alerts []domain.MarketAlert) error

	// GetAlerts retrieves alerts based on filter criteria
	GetAlerts(ctx context.Context, filter AlertFilter) ([]domain.MarketAlert, error)

	// GetAlertByID retrieves a specific alert by ID
	GetAlertByID(ctx context.Context, alertID string) (*domain.MarketAlert, error)

	// UpdateAlertStatus updates the status of an alert
	UpdateAlertStatus(ctx context.Context, alertID string, status domain.AlertStatus, resolution string) error

	// GetAlertStats retrieves alert statistics for a time period
	GetAlertStats(ctx context.Context, start, end time.Time) (AlertStats, error)

	// GetAlertsByExchange retrieves all alerts for a specific exchange
	GetAlertsByExchange(ctx context.Context, exchangeID string, limit int) ([]domain.MarketAlert, error)
}

// AlertFilter contains filter criteria for querying alerts
type AlertFilter struct {
	ExchangeID   string
	AlertType    domain.AlertType
	Severity     domain.AlertSeverity
	Status       domain.AlertStatus
	StartTime    *time.Time
	EndTime      *time.Time
	TradingPair  string
	UserID       string
	Limit        int
	Offset       int
}

// AlertStats contains aggregated alert statistics
type AlertStats struct {
	TotalAlerts     int                    `json:"total_alerts"`
	BySeverity      map[string]int         `json:"by_severity"`
	ByType          map[string]int         `json:"by_type"`
	ByExchange      map[string]int         `json:"by_exchange"`
	CriticalAlerts  []domain.MarketAlert   `json:"critical_alerts"`
}

// RuleRepository is the output port for persisting detection rules
type RuleRepository interface {
	// SaveRule saves a detection rule to the database
	SaveRule(ctx context.Context, rule domain.DetectionRule) error

	// SaveRuleBatch saves multiple detection rules
	SaveRuleBatch(ctx context.Context, rules []domain.DetectionRule) error

	// GetRuleByID retrieves a specific rule by ID
	GetRuleByID(ctx context.Context, ruleID string) (*domain.DetectionRule, error)

	// GetAllRules retrieves all detection rules
	GetAllRules(ctx context.Context) ([]domain.DetectionRule, error)

	// GetActiveRules retrieves all active detection rules
	GetActiveRules(ctx context.Context) ([]domain.DetectionRule, error)

	// GetRulesByType retrieves rules by alert type
	GetRulesByType(ctx context.Context, alertType domain.AlertType) ([]domain.DetectionRule, error)

	// UpdateRule updates an existing detection rule
	UpdateRule(ctx context.Context, rule domain.DetectionRule) error

	// DeleteRule soft deletes a detection rule (sets enabled to false)
	DeleteRule(ctx context.Context, ruleID string) error

	// HardDeleteRule permanently removes a detection rule
	HardDeleteRule(ctx context.Context, ruleID string) error
}

// ExchangeRepository is the output port for persisting exchange information
type ExchangeRepository interface {
	// SaveExchange saves an exchange profile
	SaveExchange(ctx context.Context, exchange domain.ExchangeProfile) error

	// GetExchangeByID retrieves an exchange by ID
	GetExchangeByID(ctx context.Context, exchangeID string) (*domain.ExchangeProfile, error)

	// GetAllExchanges retrieves all registered exchanges
	GetAllExchanges(ctx context.Context) ([]domain.ExchangeProfile, error)

	// GetActiveExchanges retrieves all active exchanges
	GetActiveExchanges(ctx context.Context) ([]domain.ExchangeProfile, error)

	// UpdateExchange updates an existing exchange profile
	UpdateExchange(ctx context.Context, exchange domain.ExchangeProfile) error

	// DeleteExchange soft deletes an exchange
	DeleteExchange(ctx context.Context, exchangeID string) error
}

// HealthRepository is the output port for persisting health metrics
type HealthRepository interface {
	// SaveHealthRecord saves an exchange health record
	SaveHealthRecord(ctx context.Context, health domain.ExchangeHealth) error

	// SaveHealthRecordBatch saves multiple health records
	SaveHealthRecordBatch(ctx context.Context, healthRecords []domain.ExchangeHealth) error

	// GetLatestHealth retrieves the latest health record for an exchange
	GetLatestHealth(ctx context.Context, exchangeID string) (*domain.ExchangeHealth, error)

	// GetHealthHistory retrieves health history for an exchange
	GetHealthHistory(ctx context.Context, exchangeID string, start, end time.Time) ([]domain.ExchangeHealth, error)

	// GetHealthStats retrieves aggregated health statistics
	GetHealthStats(ctx context.Context, exchangeID string, period time.Duration) (HealthStats, error)
}

// HealthStats contains aggregated health statistics
type HealthStats struct {
	ExchangeID           string    `json:"exchange_id"`
	AvgHealthScore       float64   `json:"avg_health_score"`
	MinHealthScore       float64   `json:"min_health_score"`
	MaxHealthScore       float64   `json:"max_health_score"`
	AvgLatencyMs         float64   `json:"avg_latency_ms"`
	AvgErrorRate         float64   `json:"avg_error_rate"`
	AvgUptimePercent     float64   `json:"avg_uptime_percent"`
	TotalDowntimeSeconds float64   `json:"total_downtime_seconds"`
	PeriodStart          time.Time `json:"period_start"`
	PeriodEnd            time.Time `json:"period_end"`
}

// TradeAnalyticsEngine is the output port for trade analytics (OpenSearch)
type TradeAnalyticsEngine interface {
	// IndexTrade indexes a trade event for analytics
	IndexTrade(ctx context.Context, trade domain.TradeEvent) error

	// IndexTradeBatch indexes multiple trade events
	IndexTradeBatch(ctx context.Context, trades []domain.TradeEvent) error

	// FindWashTradingPatterns detects potential wash trading patterns
	FindWashTradingPatterns(ctx context.Context, query PatternQuery) ([]domain.TradeEvent, error)

	// FindVolumeAnomalies detects unusual volume patterns
	FindVolumeAnomalies(ctx context.Context, query PatternQuery) ([]VolumeAnomaly, error)

	// CalculateVolumeStats calculates volume statistics for a time period
	CalculateVolumeStats(ctx context.Context, exchangeID, tradingPair string, start, end time.Time) (*VolumeStats, error)

	// GetTradeCount retrieves trade count for a query
	GetTradeCount(ctx context.Context, query TradeQuery) (int64, error)

	// GetAggregatedVolume retrieves aggregated volume data
	GetAggregatedVolume(ctx context.Context, query AggregationQuery) ([]AggregationResult, error)

	// DetectPriceManipulation detects potential price manipulation
	DetectPriceManipulation(ctx context.Context, query PatternQuery) ([]PriceManipulationEvent, error)
}

// PatternQuery contains query parameters for pattern detection
type PatternQuery struct {
	ExchangeID   string
	TradingPair  string
	UserID       string
	StartTime    time.Time
	EndTime      time.Time
	TimeWindow   time.Duration
	Threshold    float64
	Limit        int
}

// TradeQuery contains query parameters for trade searches
type TradeQuery struct {
	ExchangeID   string
	TradingPair  string
	UserID       string
	StartTime    *time.Time
	EndTime      *time.Time
	MinPrice     *float64
	MaxPrice     *float64
	MinVolume    *float64
	MaxVolume    *float64
	Limit        int
}

// AggregationQuery contains parameters for aggregated queries
type AggregationQuery struct {
	ExchangeID   string
	TradingPair  string
	StartTime    time.Time
	EndTime      time.Time
	Interval     time.Duration
	GroupBy      []string
	Metric       string
}

// AggregationResult contains aggregated query results
type AggregationResult struct {
	Key         string  `json:"key"`
	Interval    string  `json:"interval"`
	Count       int64   `json:"count"`
	SumVolume   float64 `json:"sum_volume"`
	AvgPrice    float64 `json:"avg_price"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	SumFees     float64 `json:"sum_fees"`
}

// VolumeAnomaly represents a detected volume anomaly
type VolumeAnomaly struct {
	ExchangeID      string    `json:"exchange_id"`
	TradingPair     string    `json:"trading_pair"`
	Timestamp       time.Time `json:"timestamp"`
	ExpectedVolume  float64   `json:"expected_volume"`
	ActualVolume    float64   `json:"actual_volume"`
	AnomalyScore    float64   `json:"anomaly_score"`
	DeviationPct    float64   `json:"deviation_percent"`
}

// VolumeStats contains volume statistics
type VolumeStats struct {
	ExchangeID       string    `json:"exchange_id"`
	TradingPair      string    `json:"trading_pair"`
	TotalVolume      float64   `json:"total_volume"`
	TradeCount       int64     `json:"trade_count"`
	AvgVolume        float64   `json:"avg_volume"`
	MaxVolume        float64   `json:"max_volume"`
	MinVolume        float64   `json:"min_volume"`
	StdDevVolume     float64   `json:"std_dev_volume"`
	MovingAvgVolume  float64   `json:"moving_avg_volume"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
}

// PriceManipulationEvent represents a potential price manipulation event
type PriceManipulationEvent struct {
	ExchangeID      string    `json:"exchange_id"`
	TradingPair     string    `json:"trading_pair"`
	Timestamp       time.Time `json:"timestamp"`
	ManipulationType string   `json:"manipulation_type"`
	PriceDeviation  float64   `json:"price_deviation"`
	VolumeInvolved  float64   `json:"volume_involved"`
	AffectedTrades  int       `json:"affected_trades"`
	Severity        string    `json:"severity"`
}

// EventBus is the output port for publishing events to Kafka
type EventBus interface {
	// PublishAlert publishes a market alert to the alert topic
	PublishAlert(ctx context.Context, alert domain.MarketAlert) error

	// PublishBatch publishes multiple alerts
	PublishBatch(ctx context.Context, alerts []domain.MarketAlert) error

	// StreamToRegulator streams regulatory reports to the reporting topic
	StreamToRegulator(ctx context.Context, report domain.RegulatoryReport) error
}

// ThrottleCommandPublisher is the output port for publishing throttle commands
type ThrottleCommandPublisher interface {
	// PublishThrottleCommand publishes a throttle command to the control topic
	PublishThrottleCommand(ctx context.Context, cmd domain.ThrottleCommand) error

	// PublishBatch publishes multiple throttle commands
	PublishBatch(ctx context.Context, cmds []domain.ThrottleCommand) error
}

// CachePort is the output port for caching operations
type CachePort interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with expiration
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, keys ...string) error

	// HGet retrieves a hash field
	HGet(ctx context.Context, key, field string) (string, error)

	// HSet sets a hash field
	HSet(ctx context.Context, key string, values ...interface{}) error

	// Incr increments a counter
	Incr(ctx context.Context, key string) (int64, error)

	// SetNX sets a value if it doesn't exist
	SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error)

	// GetMulti retrieves multiple keys
	GetMulti(ctx context.Context, keys []string) (map[string]string, error)
}

// HealthCheckPort is used for service health checks
type HealthCheckPort interface {
	// CheckDatabase verifies database connectivity
	CheckDatabase(ctx context.Context) error

	// CheckCache verifies cache connectivity
	CheckCache(ctx context.Context) error

	// CheckMessageBus verifies message bus connectivity
	CheckMessageBus(ctx context.Context) error

	// CheckAnalytics verifies analytics engine connectivity
	CheckAnalytics(ctx context.Context) error
}
