package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Severity levels for market alerts
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "LOW"
	SeverityMedium   AlertSeverity = "MEDIUM"
	SeverityHigh     AlertSeverity = "HIGH"
	SeverityCritical AlertSeverity = "CRITICAL"
)

// AlertType represents the type of market abuse detected
type AlertType string

const (
	AlertTypeWashTrading   AlertType = "WASH_TRADING"
	AlertTypeSpoofing      AlertType = "SPOOFING"
	AlertTypeLayering      AlertType = "LAYERING"
	AlertTypeManipulation  AlertType = "MANIPULATION"
	AlertTypePumpAndDump   AlertType = "PUMP_AND_DUMP"
	AlertTypeVolumeSpike   AlertType = "VOLUME_SPIKE"
	AlertTypePriceDeviation AlertType = "PRICE_DEVIATION"
	AlertTypeLatencyIssue  AlertType = "LATENCY_ISSUE"
	AlertTypeConnectivity  AlertType = "CONNECTIVITY_ISSUE"
)

// ExchangeStatus represents the operational status of an exchange
type ExchangeStatus string

const (
	ExchangeStatusActive     ExchangeStatus = "ACTIVE"
	ExchangeStatusDegraded   ExchangeStatus = "DEGRADED"
	ExchangeStatusThrottled  ExchangeStatus = "THROTTLED"
	ExchangeStatusSuspended  ExchangeStatus = "SUSPENDED"
	ExchangeStatusOffline    ExchangeStatus = "OFFLINE"
)

// TradeEvent represents a normalized trade execution from an exchange
type TradeEvent struct {
	ID            string    `json:"id" db:"id"`
	TradeID       string    `json:"trade_id" db:"trade_id"`
	ExchangeID    string    `json:"exchange_id" db:"exchange_id"`
	TradingPair   string    `json:"trading_pair" db:"trading_pair"`
	Price         float64   `json:"price" db:"price"`
	Volume        float64   `json:"volume" db:"volume"`
	QuoteVolume   float64   `json:"quote_volume" db:"quote_volume"`
	BuyerOrderID  string    `json:"buyer_order_id" db:"buyer_order_id"`
	SellerOrderID string    `json:"seller_order_id" db:"seller_order_id"`
	BuyerUserID   string    `json:"buyer_user_id" db:"buyer_user_id"`
	SellerUserID  string    `json:"seller_user_id" db:"seller_user_id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	ReceivedAt    time.Time `json:"received_at" db:"received_at"`
	IsMaker       bool      `json:"is_maker" db:"is_maker"`
	FeeCurrency   string    `json:"fee_currency" db:"fee_currency"`
	FeeAmount     float64   `json:"fee_amount" db:"fee_amount"`
	RawMessage    []byte    `json:"raw_message,omitempty" db:"raw_message"`
}

// MarketAlert represents a detected anomaly or violation
type MarketAlert struct {
	ID            string        `json:"id" db:"id"`
	Severity      AlertSeverity `json:"severity" db:"severity"`
	AlertType     AlertType     `json:"alert_type" db:"alert_type"`
	ExchangeID    string        `json:"exchange_id" db:"exchange_id"`
	TradingPair   string        `json:"trading_pair,omitempty" db:"trading_pair"`
	UserID        string        `json:"user_id,omitempty" db:"user_id"`
	Details       AlertDetails  `json:"details" db:"details"`
	Evidence      []string      `json:"evidence,omitempty" db:"evidence"`
	Status        AlertStatus   `json:"status" db:"status"`
	AssignedTo    string        `json:"assigned_to,omitempty" db:"assigned_to"`
	ResolvedAt    *time.Time    `json:"resolved_at,omitempty" db:"resolved_at"`
	Resolution    string        `json:"resolution,omitempty" db:"resolution"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// AlertDetails contains detailed information about an alert
type AlertDetails struct {
	RuleID        string  `json:"rule_id,omitempty"`
	RuleName      string  `json:"rule_name,omitempty"`
	Threshold     float64 `json:"threshold,omitempty"`
	ObservedValue float64 `json:"observed_value,omitempty"`
	TimeWindow    string  `json:"time_window,omitempty"`
	Description   string  `json:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	AlertStatusOpen       AlertStatus = "OPEN"
	AlertStatusInvestigating AlertStatus = "INVESTIGATING"
	AlertStatusResolved   AlertStatus = "RESOLVED"
	AlertStatusDismissed  AlertStatus = "DISMISSED"
	AlertStatusEscalated  AlertStatus = "ESCALATED"
)

// ExchangeHealth represents the health score and status of an exchange
type ExchangeHealth struct {
	ID              string        `json:"id" db:"id"`
	ExchangeID      string        `json:"exchange_id" db:"exchange_id"`
	HealthScore     float64       `json:"health_score" db:"health_score"`
	LatencyMs       float64       `json:"latency_ms" db:"latency_ms"`
	ErrorRate       float64       `json:"error_rate" db:"error_rate"`
	TradeVolume24h  float64       `json:"trade_volume_24h" db:"trade_volume_24h"`
	TradeCount24h   int64         `json:"trade_count_24h" db:"trade_count_24h"`
	UptimePercent   float64       `json:"uptime_percent" db:"uptime_percent"`
	LastTradeAt     *time.Time    `json:"last_trade_at,omitempty" db:"last_trade_at"`
	LastHeartbeatAt time.Time     `json:"last_heartbeat_at" db:"last_heartbeat_at"`
	Status          ExchangeStatus `json:"status" db:"status"`
	Metrics         HealthMetrics `json:"metrics,omitempty" db:"metrics"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// HealthMetrics contains detailed health metrics for an exchange
type HealthMetrics struct {
	OrderBookDepth      float64 `json:"order_book_depth"`
	SpreadBps           float64 `json:"spread_bps"`
	FillRate            float64 `json:"fill_rate"`
	AverageSlippage     float64 `json:"average_slippage"`
	DataFreshnessMs     int64   `json:"data_freshness_ms"`
	APIResponseTimeMs   int64   `json:"api_response_time_ms"`
	WebSocketLatencyMs  int64   `json:"websocket_latency_ms"`
	ConnectionStability float64 `json:"connection_stability"`
}

// DetectionRule represents a configurable rule for abuse detection
type DetectionRule struct {
	ID            string          `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	AlertType     AlertType       `json:"alert_type" db:"alert_type"`
	Severity      AlertSeverity   `json:"severity" db:"severity"`
	Logic         RuleLogic       `json:"logic" db:"logic"`
	TimeWindow    int             `json:"time_window_ms" db:"time_window_ms"`
	Threshold     float64         `json:"threshold" db:"threshold"`
	CooldownSecs  int             `json:"cooldown_secs" db:"cooldown_secs"`
	Enabled       bool            `json:"enabled" db:"is_active"`
	Exchanges     []string        `json:"exchanges,omitempty" db:"exchanges"`
	TradingPairs  []string        `json:"trading_pairs,omitempty" db:"trading_pairs"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// RuleLogic contains the specific detection logic for a rule
type RuleLogic struct {
	Type              string                 `json:"type"`
	Comparison        string                 `json:"comparison"`
	TargetField       string                 `json:"target_field"`
	GroupByFields     []string               `json:"group_by_fields"`
	Conditions        []RuleCondition        `json:"conditions"`
	CustomLogic       map[string]interface{} `json:"custom_logic,omitempty"`
}

// RuleCondition represents a single condition in a detection rule
type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// ExchangeProfile contains static information about an exchange
type ExchangeProfile struct {
	ID                string    `json:"id" db:"id"`
	ExchangeID        string    `json:"exchange_id" db:"exchange_id"`
	Name              string    `json:"name" db:"name"`
	LicenseNumber     string    `json:"license_number,omitempty" db:"license_number"`
	Jurisdiction      string    `json:"jurisdiction" db:"jurisdiction"`
	APIEndpoint       string    `json:"api_endpoint" db:"api_endpoint"`
	WebSocketEndpoint string    `json:"websocket_endpoint,omitempty" db:"websocket_endpoint"`
	RateLimitRps      int       `json:"rate_limit_rps" db:"rate_limit_rps"`
	ThrottleEnabled   bool      `json:"throttle_enabled" db:"throttle_enabled"`
	MaxLatencyMs      int       `json:"max_latency_ms" db:"max_latency_ms"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// ThrottleCommand represents a command to throttle or suspend an exchange
type ThrottleCommand struct {
	ID            string          `json:"id"`
	ExchangeID    string          `json:"exchange_id"`
	Command       ThrottleAction  `json:"command"`
	Reason        string          `json:"reason"`
	TargetRatePct float64         `json:"target_rate_percent"`
	DurationSecs  int             `json:"duration_secs"`
	IssuedBy      string          `json:"issued_by"`
	IssuedAt      time.Time       `json:"issued_at"`
	ExpiresAt     *time.Time      `json:"expires_at,omitempty"`
}

// ThrottleAction represents the type of throttle action
type ThrottleAction string

const (
	ThrottleActionLimit    ThrottleAction = "LIMIT"
	ThrottleActionSuspend  ThrottleAction = "SUSPEND"
	ThrottleActionResume   ThrottleAction = "RESUME"
	ThrottleActionBlock    ThrottleAction = "BLOCK"
)

// RegulatoryReport represents a report for central bank or regulatory body
type RegulatoryReport struct {
	ID            string          `json:"id"`
	ReportType    string          `json:"report_type"`
	PeriodStart   time.Time       `json:"period_start"`
	PeriodEnd     time.Time       `json:"period_end"`
	GeneratedAt   time.Time       `json:"generated_at"`
	ExchangeID    string          `json:"exchange_id,omitempty"`
	Summary       ReportSummary   `json:"summary"`
	Alerts        []MarketAlert   `json:"alerts"`
	Metrics       ReportMetrics   `json:"metrics"`
	Format        string          `json:"format"`
	Version       string          `json:"version"`
}

// ReportSummary contains summary information for a regulatory report
type ReportSummary struct {
	TotalTrades      int64   `json:"total_trades"`
	TotalVolume      float64 `json:"total_volume"`
	TotalAlerts      int     `json:"total_alerts"`
	CriticalAlerts   int     `json:"critical_alerts"`
	HighRiskExchanges int    `json:"high_risk_exchanges"`
	MarketHealthScore float64 `json:"market_health_score"`
}

// ReportMetrics contains detailed metrics for a regulatory report
type ReportMetrics struct {
	WashTradingCount      int                    `json:"wash_trading_count"`
	ManipulationAttempts  int                    `json:"manipulation_attempts"`
	ExchangeUptimeAvg     float64                `json:"exchange_uptime_avg"`
	TradeAnomalyRate      float64                `json:"trade_anomaly_rate"`
	VolumeByExchange      map[string]float64     `json:"volume_by_exchange"`
	AlertByType           map[string]int         `json:"alert_by_type"`
	TopViolatingUsers     []UserViolationSummary `json:"top_violating_users"`
}

// UserViolationSummary contains violation summary for a user
type UserViolationSummary struct {
	UserID         string  `json:"user_id"`
	ViolationCount int     `json:"violation_count"`
	AlertTypes     []string `json:"alert_types"`
	TotalPenalty   float64 `json:"total_penalty"`
}

// NewTradeEvent creates a new TradeEvent with generated ID
func NewTradeEvent(exchangeID, tradingPair string, price, volume float64) *TradeEvent {
	return &TradeEvent{
		ID:          uuid.New().String(),
		TradeID:     uuid.New().String(),
		ExchangeID:  exchangeID,
		TradingPair: tradingPair,
		Price:       price,
		Volume:      volume,
		QuoteVolume: price * volume,
		ReceivedAt:  time.Now().UTC(),
	}
}

// NewMarketAlert creates a new MarketAlert with generated ID
func NewMarketAlert(severity AlertSeverity, alertType AlertType, exchangeID string, details AlertDetails) *MarketAlert {
	now := time.Now().UTC()
	return &MarketAlert{
		ID:         uuid.New().String(),
		Severity:   severity,
		AlertType:  alertType,
		ExchangeID: exchangeID,
		Details:    details,
		Status:     AlertStatusOpen,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewDetectionRule creates a new DetectionRule with generated ID
func NewDetectionRule(name string, alertType AlertType, severity AlertSeverity, timeWindowMs int, threshold float64) *DetectionRule {
	now := time.Now().UTC()
	return &DetectionRule{
		ID:         uuid.New().String(),
		Name:       name,
		AlertType:  alertType,
		Severity:   severity,
		TimeWindow: timeWindowMs,
		Threshold:  threshold,
		Enabled:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewThrottleCommand creates a new ThrottleCommand with generated ID
func NewThrottleCommand(exchangeID string, action ThrottleAction, reason string, targetRatePct float64, durationSecs int) *ThrottleCommand {
	now := time.Now().UTC()
	var expiresAt *time.Time
	if durationSecs > 0 {
		expires := now.Add(time.Duration(durationSecs) * time.Second)
		expiresAt = &expires
	}
	return &ThrottleCommand{
		ID:           uuid.New().String(),
		ExchangeID:   exchangeID,
		Command:      action,
		Reason:       reason,
		TargetRatePct: targetRatePct,
		DurationSecs:  durationSecs,
		IssuedBy:     "system",
		IssuedAt:     now,
		ExpiresAt:    expiresAt,
	}
}

// ToJSON converts the alert details to JSON bytes
func (ad *AlertDetails) ToJSON() ([]byte, error) {
	return json.Marshal(ad)
}

// FromJSON parses JSON bytes into alert details
func (ad *AlertDetails) FromJSON(data []byte) error {
	return json.Unmarshal(data, ad)
}

// CalculateHealthScore calculates the overall health score based on metrics
func (eh *ExchangeHealth) CalculateHealthScore() {
	// Base score starts at 100
	score := 100.0

	// Deduct for latency issues
	if eh.LatencyMs > 1000 {
		score -= 20
	} else if eh.LatencyMs > 500 {
		score -= 10
	} else if eh.LatencyMs > 200 {
		score -= 5
	}

	// Deduct for error rate
	if eh.ErrorRate > 0.05 {
		score -= 25
	} else if eh.ErrorRate > 0.01 {
		score -= 15
	} else if eh.ErrorRate > 0.001 {
		score -= 5
	}

	// Deduct for uptime issues
	if eh.UptimePercent < 95 {
		score -= 20
	} else if eh.UptimePercent < 99 {
		score -= 10
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	eh.HealthScore = score

	// Update status based on score
	if score >= 80 {
		eh.Status = ExchangeStatusActive
	} else if score >= 60 {
		eh.Status = ExchangeStatusDegraded
	} else if score >= 40 {
		eh.Status = ExchangeStatusThrottled
	} else {
		eh.Status = ExchangeStatusSuspended
	}
}

// IsActive returns true if the exchange is considered active
func (eh *ExchangeHealth) IsActive() bool {
	return eh.Status == ExchangeStatusActive || eh.Status == ExchangeStatusDegraded
}
