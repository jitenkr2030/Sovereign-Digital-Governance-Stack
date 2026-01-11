package service

import (
	"context"
	"log"
	"math"
	"sort"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AnalysisConfig holds configuration for the analysis service
type AnalysisConfig struct {
	WashTrade       WashTradeAnalysisConfig
	Spoofing        SpoofingAnalysisConfig
	PriceDeviation  PriceDeviationConfig
	VolumeAnalysis  VolumeAnalysisConfig
}

// WashTradeAnalysisConfig holds wash trade detection settings
type WashTradeAnalysisConfig struct {
	Enabled                   bool
	TimeWindow                time.Duration
	PriceDeviationThreshold   float64
	QuantitySimilarityThreshold float64
	MinTradesForDetection     int
}

// SpoofingAnalysisConfig holds spoofing detection settings
type SpoofingAnalysisConfig struct {
	Enabled                bool
	OrderLifetimeWindow    time.Duration
	LargeOrderThreshold    float64
	CancellationRateThreshold float64
	BaitOrderThreshold     float64
}

// PriceDeviationConfig holds price deviation detection settings
type PriceDeviationConfig struct {
	Enabled              bool
	GlobalAverageWindow  time.Duration
	DeviationThreshold   float64
	MinDataPoints        int
}

// VolumeAnalysisConfig holds volume analysis settings
type VolumeAnalysisConfig struct {
	Enabled          bool
	AnomalyThreshold float64
	RollingWindow    time.Duration
}

// AnalysisService handles market analysis and pattern detection
type AnalysisService struct {
	marketRepo    port.MarketRepository
	alertRepo     port.AlertRepository
	config        AnalysisConfig
	thresholds    map[domain.AlertType]domain.AlertThreshold
	thresholdsMu  sync.RWMutex
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(marketRepo port.MarketRepository, alertRepo port.AlertRepository) *AnalysisService {
	svc := &AnalysisService{
		marketRepo: marketRepo,
		alertRepo:  alertRepo,
		config: AnalysisConfig{
			WashTrade: WashTradeAnalysisConfig{
				Enabled:                   true,
				TimeWindow:                60 * time.Second,
				PriceDeviationThreshold:   0.01,
				QuantitySimilarityThreshold: 0.9,
				MinTradesForDetection:     3,
			},
			Spoofing: SpoofingAnalysisConfig{
				Enabled:                  true,
				OrderLifetimeWindow:      300 * time.Second,
				LargeOrderThreshold:      0.1,
				CancellationRateThreshold: 0.8,
				BaitOrderThreshold:       0.05,
			},
			PriceDeviation: PriceDeviationConfig{
				Enabled:             true,
				GlobalAverageWindow: 3600 * time.Second,
				DeviationThreshold:  0.05,
				MinDataPoints:       100,
			},
			VolumeAnalysis: VolumeAnalysisConfig{
				Enabled:          true,
				AnomalyThreshold: 3.0,
				RollingWindow:    24 * time.Hour,
			},
		},
		thresholds: make(map[domain.AlertType]domain.AlertThreshold),
	}

	// Set default thresholds
	svc.setDefaultThresholds()

	return svc
}

// setDefaultThresholds initializes default thresholds
func (s *AnalysisService) setDefaultThresholds() {
	s.thresholdsMu.Lock()
	defer s.thresholdsMu.Unlock()

	s.thresholds[domain.AlertTypeWashTrading] = domain.AlertThreshold{
		Name:          "Wash Trade Detection",
		AlertType:     domain.AlertTypeWashTrading,
		Enabled:       true,
		Severity:      domain.AlertSeverityCritical,
		MinConfidence: 0.7,
		CooldownPeriod: 5 * time.Minute,
	}

	s.thresholds[domain.AlertTypeSpoofing] = domain.AlertThreshold{
		Name:          "Spoofing Detection",
		AlertType:     domain.AlertTypeSpoofing,
		Enabled:       true,
		Severity:      domain.AlertSeverityCritical,
		MinConfidence: 0.75,
		CooldownPeriod: 10 * time.Minute,
	}

	s.thresholds[domain.AlertTypePriceManipulation] = domain.AlertThreshold{
		Name:          "Price Manipulation Detection",
		AlertType:     domain.AlertTypePriceManipulation,
		Enabled:       true,
		Severity:      domain.AlertSeverityCritical,
		MinConfidence: 0.8,
		CooldownPeriod: 15 * time.Minute,
	}

	s.thresholds[domain.AlertTypeVolumeAnomaly] = domain.AlertThreshold{
		Name:          "Volume Anomaly Detection",
		AlertType:     domain.AlertTypeVolumeAnomaly,
		Enabled:       true,
		Severity:      domain.AlertSeverityWarning,
		MinConfidence: 0.6,
		CooldownPeriod: 5 * time.Minute,
	}
}

// DetectWashTrading identifies potential wash trading patterns
func (s *AnalysisService) DetectWashTrading(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.WashTradeCandidate, error) {
	if !s.config.WashTrade.Enabled {
		return nil, nil
	}

	// Get all trades in the time window
	trades, err := s.marketRepo.GetTradesByTimeWindow(ctx, exchangeID, symbol, startTime, endTime)
	if err != nil {
		return nil, err
	}

	if len(trades) < s.config.WashTrade.MinTradesForDetection {
		return nil, nil
	}

	// Group trades by account
	accountTrades := make(map[string][]domain.MarketEvent)
	for _, trade := range trades {
		if trade.UserID != nil {
			accountTrades[*trade.UserID] = append(accountTrades[*trade.UserID], trade)
		}
	}

	// Detect wash trading patterns
	var candidates []domain.WashTradeCandidate

	for accountID, accountTradeList := range accountTrades {
		if len(accountTradeList) < 2 {
			continue
		}

		// Sort by timestamp
		sort.Slice(accountTradeList, func(i, j int) bool {
			return accountTradeList[i].Timestamp.Before(accountTradeList[j].Timestamp)
		})

		// Check for self-trading (same account buying and selling)
		candidates = append(candidates, s.analyzeSelfTrading(exchangeID, symbol, accountID, accountTradeList)...)
	}

	// Check for coordinated trading (different accounts trading in patterns)
	candidates = append(candidates, s.analyzeCoordinatedTrading(exchangeID, symbol, trades, accountTrades)...)

	return candidates, nil
}

// analyzeSelfTrading analyzes trades from a single account for wash trading
func (s *AnalysisService) analyzeSelfTrading(exchangeID uuid.UUID, symbol, accountID string, trades []domain.MarketEvent) []domain.WashTradeCandidate {
	var candidates []domain.WashTradeCandidate

	windowStart := time.Now().Add(-s.config.WashTrade.TimeWindow)

	for i := 0; i < len(trades)-1; i++ {
		for j := i + 1; j < len(trades) && trades[j].Timestamp.Sub(trades[i].Timestamp) < s.config.WashTrade.TimeWindow; j++ {
			trade1 := trades[i]
			trade2 := trades[j]

			// Check if trades are in opposite directions
			if trade1.Direction == trade2.Direction {
				continue
			}

			// Check for price similarity
			priceDiff := trade1.Price.Sub(trade2.Price)
			priceDeviation := priceDiff.Abs().Div(trade1.Price.Add(trade2.Price).Div(decimal.NewFromInt(2)))

			if priceDeviation.GreaterThan(decimal.NewFromFloat(s.config.WashTrade.PriceDeviationThreshold)) {
				continue
			}

			// Check for quantity similarity
			qtyDiff := trade1.Quantity.Sub(trade2.Quantity)
			qtySimilarity := decimal.NewFromInt(1).Sub(qtyDiff.Abs().Div(trade1.Quantity.Add(trade2.Quantity).Div(decimal.NewFromInt(2))))

			if qtySimilarity.LessThan(decimal.NewFromFloat(s.config.WashTrade.QuantitySimilarityThreshold)) {
				continue
			}

			// Found potential wash trade
			candidate := domain.WashTradeCandidate{
				ID:             uuid.New(),
				ExchangeID:     exchangeID,
				Symbol:         symbol,
				AccountIDs:     []string{accountID},
				TradeIDs:       []uuid.UUID{*trade1.ID, *trade2.ID},
				TimeWindow:     trade2.Timestamp.Sub(trade1.Timestamp),
				TotalVolume:    trade1.Quantity.Add(trade2.Quantity),
				PriceDeviation: priceDeviation,
				Confidence:     s.calculateWashTradeConfidence(trade1, trade2, priceDeviation, qtySimilarity),
				DetectedAt:     time.Now(),
				Evidence:       []domain.MarketEvent{trade1, trade2},
			}

			candidates = append(candidates, candidate)
		}
	}

	return candidates
}

// analyzeCoordinatedTrading analyzes coordinated trading patterns across accounts
func (s *AnalysisService) analyzeCoordinatedTrading(exchangeID uuid.UUID, symbol string, allTrades []domain.MarketEvent, accountTrades map[string][]domain.MarketEvent) []domain.WashTradeCandidate {
	var candidates []domain.WashTradeCandidate

	// Look for accounts with very similar trade patterns
	var suspiciousAccountPairs [][]string

	accountIDs := make([]string, 0, len(accountTrades))
	for accountID := range accountTrades {
		accountIDs = append(accountIDs, accountID)
	}

	// Compare account trade patterns
	for i := 0; i < len(accountIDs); i++ {
		for j := i + 1; j < len(accountIDs); j++ {
			similarity := s.compareTradePatterns(accountTrades[accountIDs[i]], accountTrades[accountIDs[j]])
			if similarity > 0.8 {
				suspiciousAccountPairs = append(suspiciousAccountPairs, []string{accountIDs[i], accountIDs[j]})
			}
		}
	}

	// Create candidates for suspicious pairs
	for _, pair := range suspiciousAccountPairs {
		var tradeIDs []uuid.UUID
		var evidence []domain.MarketEvent

		for _, accountID := range pair {
			for _, trade := range accountTrades[accountID] {
				tradeIDs = append(tradeIDs, *trade.ID)
				evidence = append(evidence, trade)
			}
		}

		totalVolume := decimal.Zero
		for _, trade := range evidence {
			totalVolume = totalVolume.Add(trade.Quantity)
		}

		candidate := domain.WashTradeCandidate{
			ID:            uuid.New(),
			ExchangeID:    exchangeID,
			Symbol:        symbol,
			AccountIDs:    pair,
			TradeIDs:      tradeIDs,
			TimeWindow:    s.config.WashTrade.TimeWindow,
			TotalVolume:   totalVolume,
			Confidence:    0.75,
			DetectedAt:    time.Now(),
			Evidence:      evidence,
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// compareTradePatterns compares trade patterns between two accounts
func (s *AnalysisService) compareTradePatterns(trades1, trades2 []domain.MarketEvent) float64 {
	if len(trades1) == 0 || len(trades2) == 0 {
		return 0
	}

	// Compare price distribution
	prices1 := make([]decimal.Decimal, len(trades1))
	prices2 := make([]decimal.Decimal, len(trades2))

	for i, trade := range trades1 {
		prices1[i] = trade.Price
	}
	for i, trade := range trades2 {
		prices2[i] = trade.Price
	}

	avg1 := s.calculateAverage(prices1)
	avg2 := s.calculateAverage(prices2)

	if avg1.IsZero() || avg2.IsZero() {
		return 0
	}

	priceSimilarity := decimal.NewFromInt(1).Sub(avg1.Sub(avg2).Abs().Div(avg1.Add(avg2).Div(decimal.NewFromInt(2))))

	// Compare timing patterns (simplified)
	timingScore := 1.0
	if len(trades1) == len(trades2) {
		timingScore = 0.9 // High similarity if counts match
	}

	// Combine scores
	similarity := float64(priceSimilarity) * 0.7 + timingScore * 0.3

	return math.Min(1.0, similarity)
}

// calculateAverage calculates the average of a slice of decimals
func (s *AnalysisService) calculateAverage(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	return sum.Div(decimal.NewFromInt(int64(len(values))))
}

// calculateWashTradeConfidence calculates confidence score for wash trade detection
func (s *AnalysisService) calculateWashTradeConfidence(trade1, trade2 domain.MarketEvent, priceDeviation, qtySimilarity decimal.Decimal) float64 {
	// Higher confidence for:
	// - Lower price deviation
	// - Higher quantity similarity
	// - Shorter time window
	// - Same account (self-trading)

	priceScore := 1.0 - float64(priceDeviation)
	qtyScore := float64(qtySimilarity)

	// Time decay factor (more recent = higher confidence)
	timeDiff := trade2.Timestamp.Sub(trade1.Timestamp)
	timeScore := 1.0
	if timeDiff > 0 {
		timeScore = math.Exp(-float64(timeDiff) / float64(s.config.WashTrade.TimeWindow))
	}

	// Combined confidence
	confidence := priceScore*0.4 + qtyScore*0.4 + timeScore*0.2

	return math.Min(1.0, confidence)
}

// DetectSpoofing identifies potential spoofing patterns
func (s *AnalysisService) DetectSpoofing(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.SpoofingIndicator, error) {
	if !s.config.Spoofing.Enabled {
		return nil, nil
	}

	// Get cancelled orders in the time window
	cancelledOrders, err := s.marketRepo.GetCancelledOrders(ctx, exchangeID, symbol, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Get order lifetimes
	orderLifetimes, err := s.marketRepo.GetOrderLifetimes(ctx, exchangeID, symbol, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var indicators []domain.SpoofingIndicator

	// Detect bait-and-switch pattern
	for _, order := range cancelledOrders {
		if order.OrderID == nil {
			continue
		}

		// Find the order in lifetimes
		var lifetime domain.OrderLifetime
		for _, ol := range orderLifetimes {
			if ol.OrderID == *order.OrderID {
				lifetime = ol
				break
			}
		}

		if lifetime.OrderID == uuid.Nil {
			continue
		}

		// Check cancellation rate
		cancellationRate := lifetime.CancelledQty.Div(lifetime.InitialQty)
		if cancellationRate.LessThan(decimal.NewFromFloat(s.config.Spoofing.CancellationRateThreshold)) {
			continue
		}

		// Check lifetime
		if lifetime.Lifetime > s.config.Spoofing.OrderLifetimeWindow {
			continue
		}

		// Detect pattern type
		patternType := s.detectSpoofingPattern(order, lifetime, cancellationRate)

		indicator := domain.SpoofingIndicator{
			ID:               uuid.New(),
			ExchangeID:       exchangeID,
			Symbol:           symbol,
			OrderID:          *order.OrderID,
			AccountID:        *order.UserID,
			OrderSize:        order.Quantity,
			OrderPrice:       order.Price,
			OrderDirection:   order.Direction,
			Lifetime:         lifetime.Lifetime,
			CancelledAmount:  lifetime.CancelledQty,
			ExecutedAmount:   lifetime.FilledQty,
			CancellationRate: float64(cancellationRate),
			Confidence:       s.calculateSpoofingConfidence(lifetime, cancellationRate),
			PatternType:      patternType,
			DetectedAt:       time.Now(),
			Evidence:         []domain.MarketEvent{order},
		}

		indicators = append(indicators, indicator)
	}

	return indicators, nil
}

// detectSpoofingPattern determines the type of spoofing pattern
func (s *AnalysisService) detectSpoofingPattern(order domain.MarketEvent, lifetime domain.OrderLifetime, cancellationRate decimal.Decimal) string {
	// Bait laying: Large order that gets mostly cancelled
	if lifetime.CancelledQty.GreaterThan(lifetime.FilledQty.Mul(decimal.NewFromInt(3))) {
		return "BAIT_LAYING"
	}

	// Layer adding: Multiple orders at different price levels
	// This would require analyzing order book history, simplified here
	return "LAYER_ADDING"
}

// calculateSpoofingConfidence calculates confidence score for spoofing detection
func (s *AnalysisService) calculateSpoofingConfidence(lifetime domain.OrderLifetime, cancellationRate decimal.Decimal) float64 {
	// Higher confidence for:
	// - Higher cancellation rate
	// - Shorter lifetime
	// - Less execution (shows intent to manipulate, not legitimate trading)

	cancelScore := float64(cancellationRate)
	lifetimeScore := math.Exp(-float64(lifetime.Lifetime) / float64(s.config.Spoofing.OrderLifetimeWindow))
	executionScore := 1.0 - float64(lifetime.FilledQty.Div(lifetime.InitialQty))

	confidence := cancelScore*0.5 + lifetimeScore*0.3 + executionScore*0.2

	return math.Min(1.0, confidence)
}

// DetectPriceAnomalies detects price deviations from global average
func (s *AnalysisService) DetectPriceAnomalies(ctx context.Context, exchangeID uuid.UUID, symbol string, currentPrice decimal.Decimal) (*domain.PriceAnomaly, error) {
	if !s.config.PriceDeviation.Enabled {
		return nil, nil
	}

	// Get global average price
	globalAvg, err := s.marketRepo.GetGlobalAveragePrice(ctx, symbol, time.Now().Add(-s.config.PriceDeviation.GlobalAverageWindow), time.Now())
	if err != nil {
		return nil, err
	}

	if globalAvg.IsZero() {
		return nil, nil
	}

	// Calculate deviation
	deviation := currentPrice.Sub(globalAvg).Abs().Div(globalAvg)

	if deviation.LessThan(decimal.NewFromFloat(s.config.PriceDeviation.DeviationThreshold)) {
		return nil, nil
	}

	anomaly := &domain.PriceAnomaly{
		ID:             uuid.New(),
		ExchangeID:     exchangeID,
		Symbol:         symbol,
		CurrentPrice:   currentPrice,
		GlobalAverage:  globalAvg,
		DeviationPct:   deviation.Mul(decimal.NewFromInt(100)),
		ThresholdPct:   decimal.NewFromFloat(s.config.PriceDeviation.DeviationThreshold).Mul(decimal.NewFromInt(100)),
		Direction:      s.getPriceDirection(currentPrice, globalAvg),
		Confidence:     s.calculatePriceAnomalyConfidence(deviation),
		DetectedAt:     time.Now(),
	}

	return anomaly, nil
}

// getPriceDirection determines if price is above or below average
func (s *AnalysisService) getPriceDirection(current, average decimal.Decimal) string {
	if current.GreaterThan(average) {
		return "ABOVE"
	}
	return "BELOW"
}

// calculatePriceAnomalyConfidence calculates confidence for price anomaly
func (s *AnalysisService) calculatePriceAnomalyConfidence(deviation decimal.Decimal) float64 {
	// Confidence increases with deviation magnitude
	baseDeviation := decimal.NewFromFloat(s.config.PriceDeviation.DeviationThreshold)
	confidence := float64(deviation.Div(baseDeviation))

	return math.Min(1.0, confidence)
}

// DetectVolumeAnomalies detects volume spikes or drops
func (s *AnalysisService) DetectVolumeAnomalies(ctx context.Context, exchangeID uuid.UUID, symbol string, currentVolume decimal.Decimal) (*domain.VolumeAnomaly, error) {
	if !s.config.VolumeAnalysis.Enabled {
		return nil, nil
	}

	// Get market stats for historical context
	stats, err := s.marketRepo.GetMarketStats(ctx, exchangeID, symbol, s.config.VolumeAnalysis.RollingWindow)
	if err != nil {
		return nil, err
	}

	if stats.VolumeMA.IsZero() || stats.VolumeStdDev.IsZero() {
		return nil, nil
	}

	// Calculate Z-score
	zScore := currentVolume.Sub(stats.VolumeMA).Div(stats.VolumeStdDev)

	if zScore.Abs().LessThan(decimal.NewFromFloat(s.config.VolumeAnalysis.AnomalyThreshold)) {
		return nil, nil
	}

	anomaly := &domain.VolumeAnomaly{
		ID:              uuid.New(),
		ExchangeID:      exchangeID,
		Symbol:          symbol,
		CurrentVolume:   currentVolume,
		AverageVolume:   stats.VolumeMA,
		StdDev:          stats.VolumeStdDev,
		ZScore:          zScore,
		Threshold:       s.config.VolumeAnalysis.AnomalyThreshold,
		Direction:       s.getVolumeDirection(zScore),
		Confidence:      s.calculateVolumeAnomalyConfidence(zScore),
		DetectedAt:      time.Now(),
	}

	return anomaly, nil
}

// getVolumeDirection determines if volume is spike or drop
func (s *AnalysisService) getVolumeDirection(zScore decimal.Decimal) string {
	if zScore.GreaterThan(decimal.Zero) {
		return "SPIKE"
	}
	return "DROP"
}

// calculateVolumeAnomalyConfidence calculates confidence for volume anomaly
func (s *AnalysisService) calculateVolumeAnomalyConfidence(zScore decimal.Decimal) float64 {
	// Confidence based on Z-score magnitude
	absZScore := float64(zScore.Abs())
	confidence := math.Tanh(absZScore / s.config.VolumeAnalysis.AnomalyThreshold)

	return math.Min(1.0, confidence)
}

// AnalyzeEvent performs analysis on a single market event
func (s *AnalysisService) AnalyzeEvent(ctx context.Context, event *domain.MarketEvent) ([]domain.Alert, error) {
	switch event.EventType {
	case domain.MarketEventTypeTrade:
		return s.AnalyzeTrade(ctx, event)
	case domain.MarketEventTypeOrder:
		return s.AnalyzeOrder(ctx, event)
	default:
		return nil, nil
	}
}

// AnalyzeTrade analyzes a trade for suspicious patterns
func (s *AnalysisService) AnalyzeTrade(ctx context.Context, trade *domain.MarketEvent) ([]domain.Alert, error) {
	var alerts []domain.Alert

	// Get recent trades for pattern detection
	recentTrades, err := s.marketRepo.GetRecentEvents(ctx, trade.ExchangeID, trade.Symbol, s.config.WashTrade.TimeWindow)
	if err != nil {
		log.Printf("Error getting recent trades: %v", err)
		return nil, err
	}

	// Check for wash trading
	washTradeCandidates, err := s.DetectWashTrading(ctx, trade.ExchangeID, trade.Symbol, time.Now().Add(-s.config.WashTrade.TimeWindow), time.Now())
	if err != nil {
		log.Printf("Error detecting wash trading: %v", err)
		return nil, err
	}

	for _, candidate := range washTradeCandidates {
		alert := s.createWashTradeAlert(trade, candidate)
		if alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	_ = recentTrades // Use for additional pattern detection

	return alerts, nil
}

// AnalyzeOrder analyzes an order for suspicious patterns
func (s *AnalysisService) AnalyzeOrder(ctx context.Context, order *domain.MarketEvent) ([]domain.Alert, error) {
	var alerts []domain.Alert

	// Check for spoofing when orders are cancelled
	if order.Status == domain.MarketEventStatusCancelled {
		spoofingIndicators, err := s.DetectSpoofing(ctx, order.ExchangeID, order.Symbol, time.Now().Add(-s.config.Spoofing.OrderLifetimeWindow), time.Now())
		if err != nil {
			log.Printf("Error detecting spoofing: %v", err)
			return nil, err
		}

		for _, indicator := range spoofingIndicators {
			alert := s.createSpoofingAlert(order, indicator)
			if alert != nil {
				alerts = append(alerts, *alert)
			}
		}
	}

	return alerts, nil
}

// createWashTradeAlert creates an alert from a wash trade candidate
func (s *AnalysisService) createWashTradeAlert(trade *domain.MarketEvent, candidate domain.WashTradeCandidate) *domain.Alert {
	s.thresholdsMu.RLock()
	threshold, exists := s.thresholds[domain.AlertTypeWashTrading]
	s.thresholdsMu.RUnlock()

	if !exists || !threshold.Enabled {
		return nil
	}

	if candidate.Confidence < threshold.MinConfidence {
		return nil
	}

	alert := &domain.Alert{
		ID:               uuid.New(),
		Type:             domain.AlertTypeWashTrading,
		Severity:         threshold.Severity,
		Status:           domain.AlertStatusOpen,
		ExchangeID:       candidate.ExchangeID,
		Symbol:           candidate.Symbol,
		AccountID:        &candidate.AccountIDs[0],
		TradeID:          &candidate.TradeIDs[0],
		Title:            "Potential Wash Trading Detected",
		Description:      s.generateWashTradeDescription(candidate),
		Evidence: domain.AlertEvidence{
			TradeIDs:   candidate.TradeIDs,
			AccountIDs: candidate.AccountIDs,
			TimeWindow: domain.TimeWindow{
				Start: time.Now().Add(-candidate.TimeWindow),
				End:   time.Now(),
			},
		},
		RiskScore:         candidate.Confidence * 100,
		PatternConfidence: candidate.Confidence,
		DetectedAt:        time.Now(),
		UpdatedAt:         time.Now(),
		Tags:              []string{"wash_trading", "market_manipulation"},
	}

	return alert
}

// createSpoofingAlert creates an alert from a spoofing indicator
func (s *AnalysisService) createSpoofingAlert(order *domain.MarketEvent, indicator domain.SpoofingIndicator) *domain.Alert {
	s.thresholdsMu.RLock()
	threshold, exists := s.thresholds[domain.AlertTypeSpoofing]
	s.thresholdsMu.RUnlock()

	if !exists || !threshold.Enabled {
		return nil
	}

	if indicator.Confidence < threshold.MinConfidence {
		return nil
	}

	alert := &domain.Alert{
		ID:              uuid.New(),
		Type:            domain.AlertTypeSpoofing,
		Severity:        threshold.Severity,
		Status:          domain.AlertStatusOpen,
		ExchangeID:      indicator.ExchangeID,
		Symbol:          indicator.Symbol,
		AccountID:       &indicator.AccountID,
		OrderID:         &indicator.OrderID,
		Title:           "Potential Spoofing Detected",
		Description:     s.generateSpoofingDescription(indicator),
		Evidence: domain.AlertEvidence{
			OrderIDs: []uuid.UUID{indicator.OrderID},
			TimeWindow: domain.TimeWindow{
				Start: time.Now().Add(-indicator.Lifetime),
				End:   time.Now(),
			},
		},
		RiskScore:         indicator.Confidence * 100,
		PatternConfidence: indicator.Confidence,
		DetectedAt:        time.Now(),
		UpdatedAt:         time.Now(),
		Tags:              []string{"spoofing", "order_manipulation"},
	}

	return alert
}

// generateWashTradeDescription generates a human-readable description for wash trade alerts
func (s *AnalysisService) generateWashTradeDescription(candidate domain.WashTradeCandidate) string {
	return "Potential wash trading pattern detected involving account(s): " +
		joinStrings(candidate.AccountIDs, ", ") +
		". " +
		"Volume: " + candidate.TotalVolume.String() +
		", Time window: " + candidate.TimeWindow.String() +
		", Confidence: " + formatFloat(candidate.Confidence*100) + "%"
}

// generateSpoofingDescription generates a human-readable description for spoofing alerts
func (s *AnalysisService) generateSpoofingDescription(indicator domain.SpoofingIndicator) string {
	return "Potential " + indicator.PatternType + " pattern detected. " +
		"Order: " + indicator.OrderID.String() +
		", Cancellation rate: " + formatFloat(indicator.CancellationRate*100) + "%" +
		", Lifetime: " + indicator.Lifetime.String() +
		", Confidence: " + formatFloat(indicator.Confidence*100) + "%"
}

// ConfigureThresholds updates analysis thresholds
func (s *AnalysisService) ConfigureThresholds(ctx context.Context, thresholds map[domain.AlertType]map[string]any) error {
	s.thresholdsMu.Lock()
	defer s.thresholdsMu.Unlock()

	for alertType, params := range thresholds {
		if threshold, exists := s.thresholds[alertType]; exists {
			for key, value := range params {
				switch key {
				case "enabled":
					if v, ok := value.(bool); ok {
						threshold.Enabled = v
					}
				case "min_confidence":
					if v, ok := value.(float64); ok {
						threshold.MinConfidence = v
					}
				case "severity":
					if v, ok := value.(string); ok {
						threshold.Severity = domain.AlertSeverity(v)
					}
				}
			}
			threshold.UpdatedAt = time.Now()
			s.thresholds[alertType] = threshold
		}
	}

	return nil
}

// GetThresholds returns the current analysis thresholds
func (s *AnalysisService) GetThresholds(ctx context.Context) (map[domain.AlertType]domain.AlertThreshold, error) {
	s.thresholdsMu.RLock()
	defer s.thresholdsMu.RUnlock()

	result := make(map[domain.AlertType]domain.AlertThreshold)
	for k, v := range s.thresholds {
		result[k] = v
	}

	return result, nil
}

// Helper functions
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

import "fmt"
