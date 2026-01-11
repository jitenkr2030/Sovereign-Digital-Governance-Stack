package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/oversight/internal/core/domain"
)

// TradeAnalyticsAdapter provides analytics capabilities for trade data
type TradeAnalyticsAdapter struct {
	// Data sources
	tradeStore   TradeStorage
	priceStore   PriceStorage
	volumeStore  VolumeStorage
	exchangeRepo ExchangeRepository
	cache        CacheAdapter
}

// TradeStorage defines the interface for trade data access
type TradeStorage interface {
	GetTrades(ctx context.Context, query TradeQuery) ([]*domain.Trade, error)
	CountTrades(ctx context.Context, query TradeQuery) (int64, error)
	GetTradesByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*domain.Trade, error)
}

// PriceStorage defines the interface for price data access
type PriceStorage interface {
	GetPrice(ctx context.Context, symbol string, timestamp time.Time) (*domain.Price, error)
	GetPriceHistory(ctx context.Context, symbol string, start, end time.Time) ([]*domain.Price, error)
	GetCurrentPrice(ctx context.Context, symbol string) (*domain.Price, error)
}

// VolumeStorage defines the interface for volume data access
type VolumeStorage interface {
	GetVolume(ctx context.Context, symbol string, period string) (*domain.Volume, error)
	GetVolumeHistory(ctx context.Context, symbol string, start, end time.Time, interval string) ([]*domain.Volume, error)
}

// ExchangeRepository defines the interface for exchange data access
type ExchangeRepository interface {
	GetExchange(ctx context.Context, id string) (*domain.Exchange, error)
	ListExchanges(ctx context.Context) ([]*domain.Exchange, error)
	GetExchangeByAPIKey(ctx context.Context, apiKey string) (*domain.Exchange, error)
}

// CacheAdapter defines the interface for caching
type CacheAdapter interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)
}

// TradeQuery represents a query for trades
type TradeQuery struct {
	Symbol      string
	ExchangeID  string
	BuyerID     string
	SellerID    string
	StartTime   time.Time
	EndTime     time.Time
	MinPrice    float64
	MaxPrice    float64
	MinQuantity float64
	MaxQuantity float64
	Limit       int
	Offset      int
}

// NewTradeAnalyticsAdapter creates a new trade analytics adapter
func NewTradeAnalyticsAdapter(
	tradeStore TradeStorage,
	priceStore PriceStorage,
	volumeStore VolumeStorage,
	exchangeRepo ExchangeRepository,
	cache CacheAdapter,
) *TradeAnalyticsAdapter {
	return &TradeAnalyticsAdapter{
		tradeStore:   tradeStore,
		priceStore:   priceStore,
		volumeStore:  volumeStore,
		exchangeRepo: exchangeRepo,
		cache:        cache,
	}
}

// CalculateVolumeProfile calculates volume profile for a symbol
func (a *TradeAnalyticsAdapter) CalculateVolumeProfile(ctx context.Context, symbol string, period time.Duration) (*domain.VolumeProfile, error) {
	cacheKey := fmt.Sprintf("volume_profile:%s:%d", symbol, period)
	data, err := a.cache.Get(ctx, cacheKey)
	if err == nil && data != nil {
		var profile domain.VolumeProfile
		if err := profile.UnmarshalBinary(data); err == nil {
			return &profile, nil
		}
	}

	start := time.Now().Add(-period)
	trades, err := a.tradeStore.GetTradesByTimeRange(ctx, start, time.Now(), 100000)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	profile := a.buildVolumeProfile(symbol, trades, period)

	// Cache the result
	if data, err := profile.MarshalBinary(); err == nil {
		a.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return profile, nil
}

// buildVolumeProfile constructs a volume profile from trades
func (a *TradeAnalyticsAdapter) buildVolumeProfile(symbol string, trades []*domain.Trade, period time.Duration) *domain.VolumeProfile {
	profile := &domain.VolumeProfile{
		Symbol:     symbol,
		Period:     period.String(),
		Buckets:    make([]*domain.VolumeBucket, 0),
		TotalVolume: 0,
		TradeCount:  len(trades),
	}

	if len(trades) == 0 {
		return profile
	}

	// Create price buckets
	bucketCount := 50
	minPrice := trades[0].Price
	maxPrice := trades[0].Price

	for _, trade := range trades {
		if trade.Price < minPrice {
			minPrice = trade.Price
		}
		if trade.Price > maxPrice {
			maxPrice = trade.Price
		}
		profile.TotalVolume += trade.Quantity
	}

	if maxPrice == minPrice {
		maxPrice = minPrice * 1.01
	}

	bucketSize := (maxPrice - minPrice) / float64(bucketCount)
	buckets := make([]*domain.VolumeBucket, bucketCount)
	for i := range buckets {
		priceLow := minPrice + float64(i)*bucketSize
		priceHigh := priceLow + bucketSize
		buckets[i] = &domain.VolumeBucket{
			PriceLow:  priceLow,
			PriceHigh: priceHigh,
			Volume:    0,
			TradeCount: 0,
		}
	}

	// Distribute trades into buckets
	for _, trade := range trades {
		bucketIdx := int((trade.Price - minPrice) / bucketSize)
		if bucketIdx >= 0 && bucketIdx < bucketCount {
			buckets[bucketIdx].Volume += trade.Quantity
			buckets[bucketIdx].TradeCount++
		}
	}

	// Find point of control
	maxVolume := float64(0)
	var pocBucket *domain.VolumeBucket
	for _, bucket := range buckets {
		if bucket.Volume > maxVolume {
			maxVolume = bucket.Volume
			pocBucket = bucket
		}
	}

	if pocBucket != nil {
		profile.PointOfControl = (pocBucket.PriceLow + pocBucket.PriceHigh) / 2
	}

	profile.Buckets = buckets
	return profile
}

// CalculateOrderBookMetrics calculates order book metrics for a symbol
func (a *TradeAnalyticsAdapter) CalculateOrderBookMetrics(ctx context.Context, symbol string, exchangeID string) (*domain.OrderBookMetrics, error) {
	metrics := &domain.OrderBookMetrics{
		Symbol:     symbol,
		ExchangeID: exchangeID,
		CalculatedAt: time.Now(),
	}

	// Get recent trades to calculate spread and depth
	query := TradeQuery{
		Symbol:     symbol,
		ExchangeID: exchangeID,
		StartTime:  time.Now().Add(-1 * time.Hour),
		Limit:      1000,
	}

	trades, err := a.tradeStore.GetTrades(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	if len(trades) > 0 {
		var totalBuyVolume, totalSellVolume float64
		var buyCount, sellCount int

		for _, trade := range trades {
			if trade.BuyerOrderID != "" {
				totalBuyVolume += trade.Quantity
				buyCount++
			}
			if trade.SellerOrderID != "" {
				totalSellVolume += trade.Quantity
				sellCount++
			}
		}

		metrics.BuyVolume1H = totalBuyVolume
		metrics.SellVolume1H = totalSellVolume
		metrics.BuyTradeCount1H = buyCount
		metrics.SellTradeCount1H = sellCount
		metrics.BuySellRatio = float64(buyCount) / float64(sellCount+1)
	}

	return metrics, nil
}

// CalculateTradeMetrics calculates trade-related metrics
func (a *TradeAnalyticsAdapter) CalculateTradeMetrics(ctx context.Context, query TradeQuery) (*domain.TradeMetrics, error) {
	trades, err := a.tradeStore.GetTrades(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	count, err := a.tradeStore.CountTrades(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to count trades: %w", err)
	}

	metrics := &domain.TradeMetrics{
		Symbol:      query.Symbol,
		ExchangeID:  query.ExchangeID,
		StartTime:   query.StartTime,
		EndTime:     query.EndTime,
		TradeCount:  count,
	}

	if len(trades) == 0 {
		return metrics, nil
	}

	var totalVolume, totalValue float64
	var minPrice, maxPrice float64

	for _, trade := range trades {
		totalVolume += trade.Quantity
		totalValue += trade.Price * trade.Quantity

		if trade.Price < minPrice || minPrice == 0 {
			minPrice = trade.Price
		}
		if trade.Price > maxPrice {
			maxPrice = trade.Price
		}
	}

	metrics.TotalVolume = totalVolume
	metrics.AveragePrice = totalValue / totalVolume
	metrics.MinPrice = minPrice
	metrics.MaxPrice = maxPrice
	metrics.AverageQuantity = totalVolume / float64(len(trades))

	return metrics, nil
}

// DetectMarketAnomalies detects market anomalies in trade data
func (a *TradeAnalyticsAdapter) DetectMarketAnomalies(ctx context.Context, symbol string, window time.Duration) ([]*domain.MarketAnomaly, error) {
	start := time.Now().Add(-window)
	trades, err := a.tradeStore.GetTradesByTimeRange(ctx, start, time.Now(), 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	var anomalies []*domain.MarketAnomaly

	// Detect wash trading
	washTrades := a.detectWashTrading(ctx, trades)
	anomalies = append(anomalies, washTrades...)

	// Detect price manipulation
	priceAnomalies := a.detectPriceManipulation(ctx, trades)
	anomalies = append(anomalies, priceAnomalies...)

	// Detect abnormal volume
	volumeAnomalies := a.detectVolumeAnomalies(ctx, trades)
	anomalies = append(anomalies, volumeAnomalies...)

	return anomalies, nil
}

// detectWashTrading identifies potential wash trades
func (a *TradeAnalyticsAdapter) detectWashTrading(ctx context.Context, trades []*domain.Trade) []*domain.MarketAnomaly {
	var anomalies []*domain.MarketAnomaly

	// Group trades by participant
	participantTrades := make(map[string][]*domain.Trade)
	for _, trade := range trades {
		participant := trade.BuyerID + ":" + trade.SellerID
		participantTrades[[2]string{trade.BuyerID, trade.SellerID}[0]] = append(participantTrades[trade.BuyerID], trade)
	}

	// Check for circular trading patterns
	for participant, pTrades := range participantTrades {
		if len(pTrades) > 10 {
			var samePriceCount int
			for i := 1; i < len(pTrades); i++ {
				if pTrades[i].Price == pTrades[i-1].Price {
					samePriceCount++
				}
			}

			if float64(samePriceCount)/float64(len(pTrades)-1) > 0.8 {
				anomalies = append(anomalies, &domain.MarketAnomaly{
					Type:        domain.AnomalyTypeWashTrading,
					Severity:    domain.SeverityHigh,
					Description: fmt.Sprintf("High frequency of trades at same price detected for participant %s", participant),
					Timestamp:   time.Now(),
					TradeCount:  samePriceCount,
				})
			}
		}
	}

	return anomalies
}

// detectPriceManipulation identifies potential price manipulation
func (a *TradeAnalyticsAdapter) detectPriceManipulation(ctx context.Context, trades []*domain.Trade) []*domain.MarketAnomaly {
	var anomalies []*domain.MarketAnomaly

	if len(trades) < 10 {
		return anomalies
	}

	// Check for sudden price spikes
	for i := 10; i < len(trades); i++ {
		recentAvg := a.calculateAveragePrice(trades[i-10 : i])
		currentPrice := trades[i].Price

		change := (currentPrice - recentAvg) / recentAvg
		if change > 0.5 { // 50% increase
			anomalies = append(anomalies, &domain.MarketAnomaly{
				Type:        domain.AnomalyTypePriceManipulation,
				Severity:    domain.SeverityCritical,
				Description: fmt.Sprintf("Sudden price spike detected: %.2f%% increase", change*100),
				Timestamp:   trades[i].Timestamp,
				TradeCount:  1,
			})
		}
	}

	return anomalies
}

// detectVolumeAnomalies identifies abnormal volume patterns
func (a *TradeAnalyticsAdapter) detectVolumeAnomalies(ctx context.Context, trades []*domain.Trade) []*domain.MarketAnomaly {
	var anomalies []*domain.MarketAnomaly

	if len(trades) < 10 {
		return anomalies
	}

	// Calculate hourly volume
	hourlyVolume := make(map[int]float64)
	for _, trade := range trades {
		hour := trade.Timestamp.Hour()
		hourlyVolume[hour] += trade.Quantity
	}

	// Find average hourly volume
	var totalVolume float64
	for _, v := range hourlyVolume {
		totalVolume += v
	}
	avgVolume := totalVolume / float64(len(hourlyVolume))

	// Check for volume spikes
	for hour, volume := range hourlyVolume {
		if volume > avgVolume*3 {
			anomalies = append(anomalies, &domain.MarketAnomaly{
				Type:        domain.AnomalyTypeVolumeSpike,
				Severity:    domain.SeverityMedium,
				Description: fmt.Sprintf("Unusual trading volume at hour %d: %.2fx average", hour, volume/avgVolume),
				Timestamp:   time.Now(),
				TradeCount:  int(volume),
			})
		}
	}

	return anomalies
}

// calculateAveragePrice calculates the average price from trades
func (a *TradeAnalyticsAdapter) calculateAveragePrice(trades []*domain.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}

	var total float64
	for _, trade := range trades {
		total += trade.Price
	}
	return total / float64(len(trades))
}

// GetExchangeMetrics retrieves aggregated metrics for an exchange
func (a *TradeAnalyticsAdapter) GetExchangeMetrics(ctx context.Context, exchangeID string, period time.Duration) (*domain.ExchangeMetrics, error) {
	exchange, err := a.exchangeRepo.GetExchange(ctx, exchangeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange: %w", err)
	}

	query := TradeQuery{
		ExchangeID: exchangeID,
		StartTime:  time.Now().Add(-period),
		Limit:      100000,
	}

	metrics, err := a.CalculateTradeMetrics(ctx, query)
	if err != nil {
		return nil, err
	}

	return &domain.ExchangeMetrics{
		Exchange:     exchange,
		TradeMetrics: metrics,
		Period:       period.String(),
	}, nil
}
