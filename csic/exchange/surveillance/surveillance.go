// CSIC Platform - Exchange Surveillance Service
// Market manipulation detection and trading surveillance

package surveillance

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/csic-platform/exchange/ingestion"
)

// DetectionType represents the type of market manipulation
type DetectionType string

const (
    WashTrading      DetectionType = "WASH_TRADING"
    Spoofing         DetectionType = "SPOOFING"
    Layering         DetectionType = "LAYERING"
    PumpAndDump      DetectionType = "PUMP_AND_DUMP"
    FrontRunning     DetectionType = "FRONT_RUNNING"
    InsiderTrading   DetectionType = "INSIDER_TRADING"
    MarketManipulation DetectionType = "MARKET_MANIPULATION"
)

// Alert represents a market surveillance alert
type Alert struct {
    ID              string         `json:"id"`
    ExchangeID      string         `json:"exchange_id"`
    TradingPair     string         `json:"trading_pair"`
    DetectionType   DetectionType  `json:"detection_type"`
    Severity        string         `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
    Confidence      float64        `json:"confidence"` // 0-100
    Description     string         `json:"description"`
    Evidence        []Evidence     `json:"evidence"`
    Transactions    []string       `json:"transaction_ids"`
    DetectedAt      time.Time      `json:"detected_at"`
    Status          string         `json:"status"` // NEW, INVESTIGATING, RESOLVED, FALSE_POSITIVE
    AssignedTo      string         `json:"assigned_to,omitempty"`
    ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
    Resolution      string         `json:"resolution,omitempty"`
}

// Evidence represents supporting evidence for an alert
type Evidence struct {
    Type        string  `json:"type"`
    Description string  `json:"description"`
    Value       float64 `json:"value"`
    Threshold   float64 `json:"threshold"`
    Timestamp   time.Time `json:"timestamp"`
}

// TradeData represents incoming trade data
type TradeData struct {
    ExchangeID     string    `json:"exchange_id"`
    TradingPair    string    `json:"trading_pair"`
    TradeID        string    `json:"trade_id"`
    Price          float64   `json:"price"`
    Amount         float64   `json:"amount"`
    Side           string    `json:"side"` // BUY, SELL
    Timestamp      time.Time `json:"timestamp"`
    MakerOrderID   string    `json:"maker_order_id"`
    TakerOrderID   string    `json:"taker_order_id"`
}

// OrderBookData represents order book data
type OrderBookData struct {
    ExchangeID   string    `json:"exchange_id"`
    TradingPair  string    `json:"trading_pair"`
    Bids         []OrderLevel `json:"bids"`
    Asks         []OrderLevel `json:"asks"`
    Timestamp    time.Time `json:"timestamp"`
}

// OrderLevel represents an order book level
type OrderLevel struct {
    Price    float64 `json:"price"`
    Amount   float64 `json:"amount"`
    OrderID  string  `json:"order_id"`
    OrderCount int   `json:"order_count"`
}

// SurveillanceConfig contains surveillance configuration
type SurveillanceConfig struct {
    WashTradingThreshold    float64 `yaml:"wash_trading_threshold"`
    SpoofingThreshold       float64 `yaml:"spoofing_threshold"`
    LayeringThreshold       float64 `yaml:"layering_threshold"`
    PumpAndDumpThreshold    float64 `yaml:"pump_and_dump_threshold"`
    FrontRunningThreshold   float64 `yaml:"front_running_threshold"`
    AlertCooldown           int     `yaml:"alert_cooldown_minutes"`
    MinConfidence           float64 `yaml:"min_confidence_threshold"`
    CheckInterval           int     `yaml:"check_interval_seconds"`
}

// Service provides market surveillance functionality
type Service struct {
    config        *SurveillanceConfig
    tradeConsumer *ingestion.TradeConsumer
    orderConsumer *ingestion.OrderBookConsumer
    alertStore    AlertStore
    mu            sync.RWMutex
    activeAlerts  map[string]*Alert
    patternCache  *PatternCache
}

// AlertStore interface for alert persistence
type AlertStore interface {
    CreateAlert(ctx context.Context, alert *Alert) error
    UpdateAlert(ctx context.Context, alert *Alert) error
    GetAlert(ctx context.Context, id string) (*Alert, error)
    GetAlerts(ctx context.Context, filter AlertFilter) ([]*Alert, error)
}

// AlertFilter defines filters for querying alerts
type AlertFilter struct {
    ExchangeID    string
    TradingPair   string
    DetectionType DetectionType
    Severity      string
    Status        string
    StartTime     time.Time
    EndTime       time.Time
    Limit         int
    Offset        int
}

// PatternCache caches detected patterns
type PatternCache struct {
    mu               sync.RWMutex
    washTradingCache map[string]*WashTradingPattern
    spoofingCache    map[string]*SpoofingPattern
}

// WashTradingPattern represents a wash trading pattern
type WashTradingPattern struct {
    ExchangeID    string
    TradingPair   string
    TradingWindow time.Duration
    Volume        float64
    SelfTrades    int
    AccountPairs  []string
    FirstSeen     time.Time
    LastSeen      time.Time
}

// SpoofingPattern represents a spoofing pattern
type SpoofingPattern struct {
    ExchangeID    string
    TradingPair   string
    OrderID       string
    FakeVolume    float64
    RealVolume    float64
    CancellationRate float64
    FirstSeen     time.Time
    LastSeen      time.Time
}

// NewService creates a new surveillance service
func NewService(config *SurveillanceConfig, tradeConsumer *ingestion.TradeConsumer, alertStore AlertStore) *Service {
    return &Service{
        config:       config,
        tradeConsumer: tradeConsumer,
        alertStore:   alertStore,
        activeAlerts: make(map[string]*Alert),
        patternCache: &PatternCache{
            washTradingCache: make(map[string]*WashTradingPattern),
            spoofingCache:    make(map[string]*SpoofingPattern),
        },
    }
}

// Start starts the surveillance service
func (s *Service) Start(ctx context.Context) error {
    log.Println("Starting market surveillance service...")

    // Start trade monitoring
    go s.monitorTrades(ctx)

    // Start order book monitoring
    go s.monitorOrderBooks(ctx)

    // Start pattern analysis
    go s.analyzePatterns(ctx)

    log.Println("Market surveillance service started")
    return nil
}

// monitorTrades monitors incoming trades for suspicious patterns
func (s *Service) monitorTrades(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(s.config.CheckInterval) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            trades := s.tradeConsumer.GetRecentTrades(time.Minute * 5)
            for _, trade := range trades {
                s.analyzeTrade(trade)
            }
        }
    }
}

// analyzeTrade analyzes a single trade for suspicious patterns
func (s *Service) analyzeTrade(trade *TradeData) {
    // Check for wash trading
    if s.detectWashTrading(trade) {
        s.createWashTradingAlert(trade)
    }
}

// detectWashTrading detects wash trading patterns
func (s *Service) detectWashTrading(trade *TradeData) bool {
    s.patternCache.mu.RLock()
    defer s.patternCache.mu.RUnlock()

    cacheKey := fmt.Sprintf("%s:%s", trade.ExchangeID, trade.TradingPair)
    pattern, exists := s.patternCache.washTradingCache[cacheKey]

    if !exists {
        return false
    }

    // Check if self-trading detected
    if pattern.SelfTrades > 0 {
        // Calculate wash trading ratio
        washRatio := float64(pattern.SelfTrades) / float64(pattern.TradingWindow.Seconds() / 60)
        return washRatio >= s.config.WashTradingThreshold
    }

    return false
}

// createWashTradingAlert creates an alert for wash trading
func (s *Service) createWashTradingAlert(trade *TradeData) {
    cacheKey := fmt.Sprintf("%s:%s", trade.ExchangeID, trade.TradingPair)
    pattern := s.patternCache.washTradingCache[cacheKey]

    evidence := []Evidence{
        {
            Type:        "SELF_TRADE_COUNT",
            Description: "自交易次数",
            Value:       float64(pattern.SelfTrades),
            Threshold:   float64(s.config.WashTradingThreshold * 100),
            Timestamp:   time.Now(),
        },
    }

    alert := &Alert{
        ID:            generateAlertID(),
        ExchangeID:    trade.ExchangeID,
        TradingPair:   trade.TradingPair,
        DetectionType: WashTrading,
        Severity:      "HIGH",
        Confidence:    85.0,
        Description:   "检测到可疑的自交易行为",
        Evidence:      evidence,
        Transactions:  []string{trade.TradeID},
        DetectedAt:    time.Now(),
        Status:        "NEW",
    }

    s.mu.Lock()
    s.activeAlerts[alert.ID] = alert
    s.mu.Unlock()

    // Persist alert
    if err := s.alertStore.CreateAlert(context.Background(), alert); err != nil {
        log.Printf("Failed to create wash trading alert: %v", err)
    }
}

// monitorOrderBooks monitors order books for manipulation
func (s *Service) monitorOrderBooks(ctx context.Context) {
    ticker := time.NewTicker(time.Duration(s.config.CheckInterval) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            orderBooks := s.orderConsumer.GetRecentOrderBooks(time.Minute * 1)
            for _, ob := range orderBooks {
                s.analyzeOrderBook(ob)
            }
        }
    }
}

// analyzeOrderBook analyzes an order book for manipulation
func (s *Service) analyzeOrderBook(ob *OrderBookData) {
    // Check for spoofing
    if s.detectSpoofing(ob) {
        s.createSpoofingAlert(ob)
    }

    // Check for layering
    if s.detectLayering(ob) {
        s.createLayeringAlert(ob)
    }
}

// detectSpoofing detects spoofing patterns
func (s *Service) detectSpoofing(ob *OrderBookData) bool {
    // Check for large orders near market price that are likely fake
    for _, ask := range ob.Asks {
        // Skip orders that are too far from best bid
        if len(ob.Bids) == 0 {
            continue
        }
        bestBid := ob.Bids[0].Price
        spread := (ask.Price - bestBid) / bestBid

        // Check if this could be a spoofing order
        if spread < 0.01 && ask.Amount > 100 { // Large order close to market
            cacheKey := fmt.Sprintf("%s:%s:%s", ob.ExchangeID, ob.TradingPair, ask.OrderID)
            pattern, exists := s.patternCache.spoofingCache[cacheKey]

            if exists {
                pattern.FakeVolume += ask.Amount
                return pattern.FakeVolume > s.config.SpoofingThreshold
            }
        }
    }
    return false
}

// detectLayering detects layering patterns
func (s *Service) detectLayering(ob *OrderBookData) bool {
    // Check for multiple orders at similar price levels
    priceBuckets := make(map[float64]float64)

    for _, ask := range ob.Asks {
        priceBucket := ask.Price - (ask.Price % 0.001)
        priceBuckets[priceBucket] += ask.Amount
    }

    for _, amount := range priceBuckets {
        if amount > s.config.LayeringThreshold {
            return true
        }
    }

    return false
}

// createSpoofingAlert creates an alert for spoofing
func (s *Service) createSpoofingAlert(ob *OrderBookData) {
    alert := &Alert{
        ID:            generateAlertID(),
        ExchangeID:    ob.ExchangeID,
        TradingPair:   ob.TradingPair,
        DetectionType: Spoofing,
        Severity:      "HIGH",
        Confidence:    80.0,
        Description:   "检测到欺骗交易模式",
        DetectedAt:    time.Now(),
        Status:        "NEW",
    }

    s.mu.Lock()
    s.activeAlerts[alert.ID] = alert
    s.mu.Unlock()

    if err := s.alertStore.CreateAlert(context.Background(), alert); err != nil {
        log.Printf("Failed to create spoofing alert: %v", err)
    }
}

// createLayeringAlert creates an alert for layering
func (s *Service) createLayeringAlert(ob *OrderBookData) {
    alert := &Alert{
        ID:            generateAlertID(),
        ExchangeID:    ob.ExchangeID,
        TradingPair:   ob.TradingPair,
        DetectionType: Layering,
        Severity:      "MEDIUM",
        Confidence:    75.0,
        Description:   "检测到层次交易模式",
        DetectedAt:    time.Now(),
        Status:        "NEW",
    }

    s.mu.Lock()
    s.activeAlerts[alert.ID] = alert
    s.mu.Unlock()

    if err := s.alertStore.CreateAlert(context.Background(), alert); err != nil {
        log.Printf("Failed to create layering alert: %v", err)
    }
}

// analyzePatterns runs periodic pattern analysis
func (s *Service) analyzePatterns(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.runPatternAnalysis()
        }
    }
}

// runPatternAnalysis runs comprehensive pattern analysis
func (s *Service) runPatternAnalysis() {
    // Analyze wash trading patterns
    s.analyzeWashTradingPatterns()

    // Analyze pump and dump patterns
    s.analyzePumpAndDumpPatterns()

    // Analyze front running patterns
    s.analyzeFrontRunningPatterns()
}

// analyzeWashTradingPatterns analyzes for wash trading
func (s *Service) analyzeWashTradingPatterns() {
    // Implementation for wash trading pattern analysis
}

// analyzePumpAndDumpPatterns analyzes for pump and dump
func (s *Service) analyzePumpAndDumpPatterns() {
    // Implementation for pump and dump pattern analysis
}

// analyzeFrontRunningPatterns analyzes for front running
func (s *Service) analyzeFrontRunningPatterns() {
    // Implementation for front running pattern analysis
}

// GetActiveAlerts returns all active alerts
func (s *Service) GetActiveAlerts() []*Alert {
    s.mu.RLock()
    defer s.mu.RUnlock()

    alerts := make([]*Alert, 0, len(s.activeAlerts))
    for _, alert := range s.activeAlerts {
        alerts = append(alerts, alert)
    }
    return alerts
}

// GetAlerts returns alerts matching the filter
func (s *Service) GetAlerts(ctx context.Context, filter AlertFilter) ([]*Alert, error) {
    return s.alertStore.GetAlerts(ctx, filter)
}

// ResolveAlert marks an alert as resolved
func (s *Service) ResolveAlert(ctx context.Context, id, resolution string) error {
    alert, err := s.alertStore.GetAlert(ctx, id)
    if err != nil {
        return err
    }

    now := time.Now()
    alert.Status = "RESOLVED"
    alert.ResolvedAt = &now
    alert.Resolution = resolution

    s.mu.Lock()
    delete(s.activeAlerts, id)
    s.mu.Unlock()

    return s.alertStore.UpdateAlert(ctx, alert)
}

// Helper function to generate unique alert ID
func generateAlertID() string {
    return fmt.Sprintf("salert_%d", time.Now().UnixNano())
}
