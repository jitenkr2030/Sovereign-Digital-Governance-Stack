package services

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"go.uber.org/zap"
)

// AbuseDetectorService detects market abuse patterns in trade data
// It implements sliding window algorithms and pattern matching to identify
// wash trading, spoofing, layering, and other manipulation behaviors
type AbuseDetectorService struct {
	ruleRepo     ports.RuleRepository
	alertRepo    ports.AlertRepository
	eventBus     ports.EventBus
	logger       *zap.Logger
	windowStore  *TradeWindowStore
	mu           sync.RWMutex
}

// TradeWindowStore maintains a sliding window of recent trades for pattern detection
type TradeWindowStore struct {
	trades      []domain.TradeEvent
	capacity    int
	windowSize  time.Duration
	mu          sync.RWMutex
}

// NewAbuseDetectorService creates a new AbuseDetectorService
func NewAbuseDetectorService(
	ruleRepo ports.RuleRepository,
	alertRepo ports.AlertRepository,
	eventBus ports.EventBus,
	logger *zap.Logger,
) *AbuseDetectorService {
	return &AbuseDetectorService{
		ruleRepo:    ruleRepo,
		alertRepo:   alertRepo,
		eventBus:    eventBus,
		logger:      logger,
		windowStore: NewTradeWindowStore(10000, 1*time.Minute),
	}
}

// ProcessTradeStream processes a trade event for abuse detection
func (s *AbuseDetectorService) ProcessTradeStream(ctx context.Context, trade domain.TradeEvent) error {
	// Get active detection rules
	rules, err := s.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		s.logger.Error("Failed to get active rules", zap.Error(err))
		return err
	}

	// Add trade to sliding window
	s.windowStore.Add(trade)

	// Run all applicable detection checks
	for _, rule := range rules {
		if s.ruleAppliesToTrade(rule, trade) {
			switch rule.AlertType {
			case domain.AlertTypeWashTrading:
				s.checkWashTrading(ctx, trade, rule)
			case domain.AlertTypeVolumeSpike:
				s.checkVolumeSpike(ctx, trade, rule)
			case domain.AlertTypeManipulation:
				s.checkManipulation(ctx, trade, rule)
			case domain.AlertTypePriceDeviation:
				s.checkPriceDeviation(ctx, trade, rule)
			}
		}
	}

	return nil
}

// ruleAppliesToTrade checks if a rule applies to a specific trade
func (s *AbuseDetectorService) ruleAppliesToTrade(rule domain.DetectionRule, trade domain.TradeEvent) bool {
	// Check exchange filter
	if len(rule.Exchanges) > 0 {
		applies := false
		for _, exchange := range rule.Exchanges {
			if exchange == trade.ExchangeID {
				applies = true
				break
			}
		}
		if !applies {
			return false
		}
	}

	// Check trading pair filter
	if len(rule.TradingPairs) > 0 {
		applies := false
		for _, pair := range rule.TradingPairs {
			if pair == trade.TradingPair {
				applies = true
				break
			}
		}
		if !applies {
			return false
		}
	}

	return true
}

// checkWashTrading detects wash trading patterns
// Wash trading occurs when the same entities trade with themselves
// or coordinate to trade back and forth at similar prices
func (s *AbuseDetectorService) checkWashTrading(ctx context.Context, trade domain.TradeEvent, rule domain.DetectionRule) {
	s.windowStore.mu.RLock()
	defer s.windowStore.mu.RUnlock()

	// Check for self-trading (same buyer and seller)
	if trade.BuyerUserID != "" && trade.SellerUserID != "" && trade.BuyerUserID == trade.SellerUserID {
		s.createAlert(ctx, trade, domain.AlertTypeWashTrading, domain.SeverityCritical, rule,
			"Self-trading detected: buyer and seller are the same user",
			map[string]interface{}{
				"user_id":     trade.BuyerUserID,
				"trade_id":    trade.TradeID,
				"price":       trade.Price,
				"volume":      trade.Volume,
			})
		return
	}

	// Check for circular trading patterns
	// Look for trades between same user pairs within time window
	recentTrades := s.windowStore.GetRecentTrades(rule.TimeWindow)
	userPairTrades := make([]domain.TradeEvent, 0)

	for _, t := range recentTrades {
		if (t.BuyerUserID == trade.BuyerUserID && t.SellerUserID == trade.SellerUserID) ||
			(t.BuyerUserID == trade.SellerUserID && t.SellerUserID == trade.BuyerUserID) {
			userPairTrades = append(userPairTrades, t)
		}
	}

	// If more trades than threshold between same user pair, flag as wash trading
	if len(userPairTrades) >= int(rule.Threshold) {
		// Calculate price variance
		priceVariance := s.calculatePriceVariance(userPairTrades)

		// If price variance is very low, it's likely wash trading
		if priceVariance < 0.001 { // 0.1% variance threshold
			s.createAlert(ctx, trade, domain.AlertTypeWashTrading, domain.SeverityHigh, rule,
				"Potential wash trading: circular trades between same users with minimal price movement",
				map[string]interface{}{
					"user_id_1":      trade.BuyerUserID,
					"user_id_2":      trade.SellerUserID,
					"trade_count":    len(userPairTrades) + 1,
					"price_variance": priceVariance,
					"time_window_ms": rule.TimeWindow,
				})
		}
	}
}

// checkVolumeSpike detects unusual volume spikes
func (s *AbuseDetectorService) checkVolumeSpike(ctx context.Context, trade domain.TradeEvent, rule domain.DetectionRule) {
	s.windowStore.mu.RLock()
	defer s.windowStore.mu.RUnlock()

	// Get recent trades for the same trading pair
	recentTrades := s.windowStore.GetRecentTrades(rule.TimeWindow)
	pairTrades := make([]domain.TradeEvent, 0)

	for _, t := range recentTrades {
		if t.TradingPair == trade.TradingPair && t.ExchangeID == trade.ExchangeID {
			pairTrades = append(pairTrades, t)
		}
	}

	if len(pairTrades) == 0 {
		return
	}

	// Calculate average volume
	totalVolume := trade.Volume
	for _, t := range pairTrades {
		totalVolume += t.Volume
	}
	avgVolume := totalVolume / float64(len(pairTrades)+1)

	// Check for spike
	if avgVolume > 0 && trade.Volume > avgVolume*rule.Threshold {
		spikeMultiple := trade.Volume / avgVolume
		s.createAlert(ctx, trade, domain.AlertTypeVolumeSpike, domain.SeverityMedium, rule,
			"Unusual volume spike detected",
			map[string]interface{}{
				"trade_volume":   trade.Volume,
				"avg_volume":     avgVolume,
				"spike_multiple": spikeMultiple,
				"threshold":      rule.Threshold,
				"trading_pair":   trade.TradingPair,
			})
	}
}

// checkManipulation detects potential market manipulation
func (s *AbuseDetectorService) checkManipulation(ctx context.Context, trade domain.TradeEvent, rule domain.DetectionRule) {
	s.windowStore.mu.RLock()
	defer s.windowStore.mu.RUnlock()

	// Check for pump and dump patterns
	// Rapid price increase followed by high volume sell-off
	recentTrades := s.windowStore.GetRecentTrades(rule.TimeWindow)

	var priceChanges []float64
	var volumes []float64
	var relevantTrades []domain.TradeEvent

	for _, t := range recentTrades {
		if t.TradingPair == trade.TradingPair && t.ExchangeID == trade.ExchangeID {
			relevantTrades = append(relevantTrades, t)
			if len(relevantTrades) > 1 {
				priceChanges = append(priceChanges, (t.Price-relevantTrades[len(relevantTrades)-2].Price)/relevantTrades[len(relevantTrades)-2].Price)
			}
			volumes = append(volumes, t.Volume)
		}
	}

	if len(relevantTrades) < 3 {
		return
	}

	// Check for large price movement
	for i, change := range priceChanges {
		if math.Abs(change) > rule.Threshold {
			// Check if followed by high volume
			if i < len(volumes) && volumes[i] > s.calculateAverage(volumes)*rule.Threshold {
				s.createAlert(ctx, trade, domain.AlertTypeManipulation, domain.SeverityHigh, rule,
					"Potential market manipulation: rapid price movement with high volume",
					map[string]interface{}{
						"price_change":     change * 100,
						"volume":           volumes[i],
						"avg_volume":       s.calculateAverage(volumes),
						"trading_pair":     trade.TradingPair,
						"price_direction":  func() string { if change > 0 { return "up" } else { return "down" } }(),
					})
			}
		}
	}
}

// checkPriceDeviation detects price deviation from market average
func (s *AbuseDetectorService) checkPriceDeviation(ctx context.Context, trade domain.TradeEvent, rule domain.DetectionRule) {
	s.windowStore.mu.RLock()
	defer s.windowStore.mu.RUnlock()

	// Get reference price from recent trades
	recentTrades := s.windowStore.GetRecentTrades(rule.TimeWindow)
	var relevantTrades []domain.TradeEvent

	for _, t := range recentTrades {
		if t.TradingPair == trade.TradingPair {
			relevantTrades = append(relevantTrades, t)
		}
	}

	if len(relevantTrades) == 0 {
		return
	}

	// Calculate average price
	avgPrice := s.calculateAveragePrice(relevantTrades)

	// Calculate deviation
	if avgPrice > 0 {
		deviation := math.Abs(trade.Price - avgPrice) / avgPrice

		if deviation > rule.Threshold {
			s.createAlert(ctx, trade, domain.AlertTypePriceDeviation, domain.SeverityMedium, rule,
				"Price deviation from market average detected",
				map[string]interface{}{
					"trade_price":     trade.Price,
					"avg_price":       avgPrice,
					"deviation_pct":   deviation * 100,
					"threshold_pct":   rule.Threshold * 100,
					"trading_pair":    trade.TradingPair,
				})
		}
	}
}

// createAlert creates and persists a market alert
func (s *AbuseDetectorService) createAlert(
	ctx context.Context,
	trade domain.TradeEvent,
	alertType domain.AlertType,
	severity domain.AlertSeverity,
	rule domain.DetectionRule,
	description string,
	metadata map[string]interface{},
) {
	details := domain.AlertDetails{
		RuleID:        rule.ID,
		RuleName:      rule.Name,
		Threshold:     rule.Threshold,
		ObservedValue: 0, // Calculated based on alert type
		TimeWindow:    formatDuration(time.Duration(rule.TimeWindow) * time.Millisecond),
		Description:   description,
		Metadata:      metadata,
	}

	alert := domain.NewMarketAlert(severity, alertType, trade.ExchangeID, details)
	alert.TradingPair = trade.TradingPair
	alert.UserID = trade.BuyerUserID
	if trade.SellerUserID != "" {
		alert.Evidence = append(alert.Evidence, trade.TradeID)
	}

	// Persist alert
	if err := s.alertRepo.SaveAlert(ctx, *alert); err != nil {
		s.logger.Error("Failed to save alert", zap.Error(err), zap.String("alert_id", alert.ID))
	}

	// Publish to event bus
	if err := s.eventBus.PublishAlert(ctx, *alert); err != nil {
		s.logger.Error("Failed to publish alert", zap.Error(err), zap.String("alert_id", alert.ID))
	}

	s.logger.Info("Market abuse alert created",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", string(alertType)),
		zap.String("severity", string(severity)),
		zap.String("exchange_id", trade.ExchangeID),
		zap.String("trading_pair", trade.TradingPair),
		zap.String("description", description),
	)
}

// calculatePriceVariance calculates the price variance for a set of trades
func (s *AbuseDetectorService) calculatePriceVariance(trades []domain.TradeEvent) float64 {
	if len(trades) == 0 {
		return 0
	}

	avgPrice := s.calculateAveragePrice(trades)
	if avgPrice == 0 {
		return 0
	}

	var sumSqDiff float64
	for _, t := range trades {
		diff := (t.Price - avgPrice) / avgPrice
		sumSqDiff += diff * diff
	}

	return math.Sqrt(sumSqDiff / float64(len(trades)))
}

// calculateAverage calculates the average of a slice of floats
func (s *AbuseDetectorService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateAveragePrice calculates the average price from trades
func (s *AbuseDetectorService) calculateAveragePrice(trades []domain.TradeEvent) float64 {
	if len(trades) == 0 {
		return 0
	}

	var sum float64
	for _, t := range trades {
		sum += t.Price
	}
	return sum / float64(len(trades))
}

// NewTradeWindowStore creates a new TradeWindowStore
func NewTradeWindowStore(capacity int, windowSize time.Duration) *TradeWindowStore {
	return &TradeWindowStore{
		capacity:   capacity,
		windowSize: windowSize,
		trades:     make([]domain.TradeEvent, 0, capacity),
	}
}

// Add adds a trade to the window store
func (t *TradeWindowStore) Add(trade domain.TradeEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.trades = append(t.trades, trade)

	// Remove old trades outside the window
	cutoff := time.Now().UTC().Add(-t.windowSize)
	for len(t.trades) > 0 && t.trades[0].Timestamp.Before(cutoff) {
		t.trades = t.trades[1:]
	}

	// Ensure capacity limit
	if len(t.trades) > t.capacity {
		t.trades = t.trades[len(t.trades)-t.capacity:]
	}
}

// GetRecentTrades returns trades within the specified time window
func (t *TradeWindowStore) GetRecentTrades(windowMs int) []domain.TradeEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	windowDuration := time.Duration(windowMs) * time.Millisecond
	cutoff := time.Now().UTC().Add(-windowDuration)

	result := make([]domain.TradeEvent, 0)
	for _, trade := range t.trades {
		if trade.Timestamp.After(cutoff) {
			result = append(result, trade)
		}
	}

	return result
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return d.Truncate(time.Hour).String()
	}
	if d >= time.Minute {
		return d.Truncate(time.Minute).String()
	}
	return d.Truncate(time.Second).String()
}
