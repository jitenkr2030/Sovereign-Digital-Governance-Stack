package services

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"go.uber.org/zap"
)

// HealthScorerService calculates and monitors exchange health scores
// It tracks latency, error rates, uptime, and other metrics to provide
// real-time health assessments and trigger throttling when necessary
type HealthScorerService struct {
	healthRepo       ports.HealthRepository
	exchangeRepo     ports.ExchangeRepository
	throttlePublisher ports.ThrottleCommandPublisher
	cachePort        ports.CachePort
	logger           *zap.Logger
	metrics          *HealthScorerMetrics
	mu               sync.RWMutex
}

// HealthScorerMetrics tracks health scoring performance
type HealthScorerMetrics struct {
	TotalUpdates     int64
	ThrottleEvents   int64
	LastUpdatedAt    time.Time
	HealthByExchange map[string]float64
}

// NewHealthScorerService creates a new HealthScorerService
func NewHealthScorerService(
	healthRepo ports.HealthRepository,
	exchangeRepo ports.ExchangeRepository,
	throttlePublisher ports.ThrottleCommandPublisher,
	cachePort ports.CachePort,
	logger *zap.Logger,
) *HealthScorerService {
	return &HealthScorerService{
		healthRepo:       healthRepo,
		exchangeRepo:     exchangeRepo,
		throttlePublisher: throttlePublisher,
		cachePort:        cachePort,
		logger:           logger,
		metrics: &HealthScorerMetrics{
			HealthByExchange: make(map[string]float64),
		},
	}
}

// ProcessTradeStream updates health metrics based on a trade event
func (s *HealthScorerService) ProcessTradeStream(ctx context.Context, trade domain.TradeEvent) error {
	// Update cache with trade metrics
	cacheKey := "health:" + trade.ExchangeID

	// Increment trade count
	tradeCountKey := cacheKey + ":count"
	if _, err := s.cachePort.Incr(ctx, tradeCountKey); err != nil {
		s.logger.Warn("Failed to increment trade count", zap.Error(err))
	}

	// Update last trade timestamp
	lastTradeKey := cacheKey + ":last_trade"
	_ = s.cachePort.Set(ctx, lastTradeKey, trade.Timestamp.Format(time.RFC3339), 24*time.Hour)

	// Update volume
	volumeKey := cacheKey + ":volume"
	volume, _ := s.cachePort.Get(ctx, volumeKey)
	var currentVolume float64
	if volume != "" {
		currentVolume = parseFloat64(volume)
	}
	currentVolume += trade.Volume
	_ = s.cachePort.Set(ctx, volumeKey, floatToString(currentVolume), 24*time.Hour)

	// Calculate and update health score periodically (every 100 trades or every minute)
	s.mu.Lock()
	s.metrics.TotalUpdates++
	s.metrics.LastUpdatedAt = time.Now()
	s.metrics.HealthByExchange[trade.ExchangeID] = 0 // Will be calculated
	s.mu.Unlock()

	// Check if health score needs recalculation
	shouldRecalculate := s.shouldRecalculateHealth(trade.ExchangeID)
	if shouldRecalculate {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = s.RecalculateHealthScore(ctx, trade.ExchangeID)
		}()
	}

	return nil
}

// GetExchangeHealth retrieves the current health status of an exchange
func (s *HealthScorerService) GetExchangeHealth(ctx context.Context, exchangeID string) (*domain.ExchangeHealth, error) {
	// Try to get from cache first
	cacheKey := "health:" + exchangeID
	cachedHealth, err := s.cachePort.Get(ctx, cacheKey+":score")
	if err == nil && cachedHealth != "" {
		// Return cached health
		return s.buildHealthFromCache(ctx, exchangeID, cachedHealth)
	}

	// Get from database
	health, err := s.healthRepo.GetLatestHealth(ctx, exchangeID)
	if err != nil {
		s.logger.Error("Failed to get health from database", zap.Error(err))
		return nil, err
	}

	// Cache the result
	_ = s.cachePort.Set(ctx, cacheKey+":score", floatToString(health.HealthScore), 5*time.Minute)

	return health, nil
}

// GetAllExchangeHealth retrieves health status for all monitored exchanges
func (s *HealthScorerService) GetAllExchangeHealth(ctx context.Context) ([]domain.ExchangeHealth, error) {
	exchanges, err := s.exchangeRepo.GetActiveExchanges(ctx)
	if err != nil {
		return nil, err
	}

	healthRecords := make([]domain.ExchangeHealth, 0, len(exchanges))
	for _, exchange := range exchanges {
		health, err := s.GetExchangeHealth(ctx, exchange.ExchangeID)
		if err != nil {
			s.logger.Warn("Failed to get health for exchange",
				zap.String("exchange_id", exchange.ExchangeID),
				zap.Error(err),
			)
			// Create default health record
			health = s.createDefaultHealth(exchange.ExchangeID)
		}
		healthRecords = append(healthRecords, *health)
	}

	return healthRecords, nil
}

// RecalculateHealthScore recalculates the health score for an exchange
func (s *HealthScorerService) RecalculateHealthScore(ctx context.Context, exchangeID string) error {
	// Get cached metrics
	cacheKey := "health:" + exchangeID

	tradeCountStr, _ := s.cachePort.Get(ctx, cacheKey+":count")
	volumeStr, _ := s.cachePort.Get(ctx, cacheKey+":volume")
	lastTradeStr, _ := s.cachePort.Get(ctx, cacheKey+":last_trade")

	tradeCount := parseInt64(tradeCountStr)
	volume := parseFloat64(volumeStr)
	var lastTradeAt *time.Time
	if lastTradeStr != "" {
		t, _ := time.Parse(time.RFC3339, lastTradeStr)
		lastTradeAt = &t
	}

	// Calculate health score components
	healthScore := s.calculateHealthScore(ctx, exchangeID, tradeCount, volume)

	// Create health record
	health := &domain.ExchangeHealth{
		ExchangeID:     exchangeID,
		HealthScore:    healthScore.Score,
		LatencyMs:      healthScore.LatencyMs,
		ErrorRate:      healthScore.ErrorRate,
		TradeVolume24h: volume,
		TradeCount24h:  tradeCount,
		UptimePercent:  healthScore.UptimePercent,
		LastTradeAt:    lastTradeAt,
		Status:         healthScore.Status,
		UpdatedAt:      time.Now().UTC(),
	}

	// Save to database
	if err := s.healthRepo.SaveHealthRecord(ctx, *health); err != nil {
		s.logger.Error("Failed to save health record", zap.Error(err))
		return err
	}

	// Update cache
	_ = s.cachePort.Set(ctx, cacheKey+":score", floatToString(healthScore.Score), 5*time.Minute)

	// Check if throttling is needed
	if healthScore.Score < 40 {
		s.triggerThrottle(ctx, exchangeID, healthScore.Score)
	}

	// Update metrics
	s.mu.Lock()
	s.metrics.HealthByExchange[exchangeID] = healthScore.Score
	s.mu.Unlock()

	s.logger.Info("Health score recalculated",
		zap.String("exchange_id", exchangeID),
		zap.Float64("health_score", healthScore.Score),
		zap.String("status", string(healthScore.Status)),
	)

	return nil
}

// calculateHealthScore calculates the health score based on metrics
func (s *HealthScorerService) calculateHealthScore(
	ctx context.Context,
	exchangeID string,
	tradeCount int64,
	volume float64,
) HealthScoreResult {
	result := HealthScoreResult{
		Score: 100,
	}

	// Get exchange profile for thresholds
	profile, err := s.exchangeRepo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		// Use defaults
		result.LatencyMs = 100 // Default latency assumption
		result.ErrorRate = 0.001
		result.UptimePercent = 99.9
	} else {
		result.LatencyMs = float64(profile.MaxLatencyMs)
	}

	// Deduct for latency (if we have latency data)
	latencyKey := "health:" + exchangeID + ":latency"
	latencyStr, _ := s.cachePort.Get(ctx, latencyKey)
	if latencyStr != "" {
		latency := parseFloat64(latencyStr)
		result.LatencyMs = latency

		if latency > float64(result.LatencyMs) {
			result.Score -= 20
		} else if latency > float64(result.LatencyMs)*0.8 {
			result.Score -= 10
		} else if latency > float64(result.LatencyMs)*0.5 {
			result.Score -= 5
		}
	}

	// Deduct for error rate
	errorKey := "health:" + exchangeID + ":errors"
	errorStr, _ := s.cachePort.Get(ctx, errorKey)
	if errorStr != "" {
		errors := parseInt64(errorStr)
		if tradeCount > 0 {
			result.ErrorRate = float64(errors) / float64(tradeCount)
		}

		if result.ErrorRate > 0.05 {
			result.Score -= 25
		} else if result.ErrorRate > 0.01 {
			result.Score -= 15
		} else if result.ErrorRate > 0.001 {
			result.Score -= 5
		}
	}

	// Uptime is based on last trade timestamp
	uptimeKey := "health:" + exchangeID + ":uptime"
	uptimeStr, _ := s.cachePort.Get(ctx, uptimeKey)
	if uptimeStr != "" {
		uptime := parseFloat64(uptimeStr)
		result.UptimePercent = uptime
	}

	// Ensure score is within bounds
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 100 {
		result.Score = 100
	}

	// Determine status
	if result.Score >= 80 {
		result.Status = domain.ExchangeStatusActive
	} else if result.Score >= 60 {
		result.Status = domain.ExchangeStatusDegraded
	} else if result.Score >= 40 {
		result.Status = domain.ExchangeStatusThrottled
	} else {
		result.Status = domain.ExchangeStatusSuspended
	}

	return result
}

// HealthScoreResult contains the calculated health score components
type HealthScoreResult struct {
	Score         float64
	LatencyMs     float64
	ErrorRate     float64
	UptimePercent float64
	Status        domain.ExchangeStatus
}

// shouldRecalculateHealth determines if health score needs recalculation
func (s *HealthScorerService) shouldRecalculateHealth(exchangeID string) bool {
	s.mu.RLock()
	updates := s.metrics.TotalUpdates
	s.mu.RUnlock()

	// Recalculate every 100 trades or based on counter
	// This is a simplified check - in production, use more sophisticated logic
	_ = exchangeID // Suppress unused variable warning
	return updates%100 == 0
}

// triggerThrottle initiates throttling for an exchange
func (s *HealthScorerService) triggerThrottle(ctx context.Context, exchangeID string, healthScore float64) {
	// Get exchange profile for throttling settings
	profile, err := s.exchangeRepo.GetExchangeByID(ctx, exchangeID)
	if err != nil || !profile.ThrottleEnabled {
		s.logger.Info("Throttling disabled for exchange, skipping",
			zap.String("exchange_id", exchangeID),
		)
		return
	}

	// Determine throttle rate based on health score
	targetRatePct := s.calculateThrottleRate(healthScore)

	// Create throttle command
	cmd := domain.NewThrottleCommand(
		exchangeID,
		domain.ThrottleActionLimit,
		"Low health score triggering automatic throttling",
		targetRatePct,
		300, // 5 minutes duration
	)

	// Publish throttle command
	if err := s.throttlePublisher.PublishThrottleCommand(ctx, *cmd); err != nil {
		s.logger.Error("Failed to publish throttle command",
			zap.String("exchange_id", exchangeID),
			zap.Error(err),
		)
		return
	}

	s.mu.Lock()
	s.metrics.ThrottleEvents++
	s.mu.Unlock()

	s.logger.Warn("Throttle command issued",
		zap.String("exchange_id", exchangeID),
		zap.Float64("health_score", healthScore),
		zap.Float64("target_rate_pct", targetRatePct),
		zap.Int("duration_secs", 300),
	)
}

// calculateThrottleRate calculates the throttle rate based on health score
func (s *HealthScorerService) calculateThrottleRate(healthScore float64) float64 {
	// Health score 40-60: 50% throttle
	// Health score 20-40: 75% throttle
	// Health score < 20: 90% throttle
	if healthScore >= 40 {
		return 50
	} else if healthScore >= 20 {
		return 75
	}
	return 90
}

// UpdateLatency updates the latency metric for an exchange
func (s *HealthScorerService) UpdateLatency(ctx context.Context, exchangeID string, latencyMs int64) {
	cacheKey := "health:" + exchangeID + ":latency"
	_ = s.cachePort.Set(ctx, cacheKey, floatToString(float64(latencyMs)), 24*time.Hour)
}

// UpdateError updates the error count for an exchange
func (s *HealthScorerService) UpdateError(ctx context.Context, exchangeID string) {
	cacheKey := "health:" + exchangeID + ":errors"
	current, _ := s.cachePort.Get(ctx, cacheKey)
	count := parseInt64(current) + 1
	_ = s.cachePort.Set(ctx, cacheKey, int64ToString(count), 24*time.Hour)
}

// UpdateUptime updates the uptime metric for an exchange
func (s *HealthScorerService) UpdateUptime(ctx context.Context, exchangeID string, uptimePercent float64) {
	cacheKey := "health:" + exchangeID + ":uptime"
	_ = s.cachePort.Set(ctx, cacheKey, floatToString(uptimePercent), 24*time.Hour)
}

// createDefaultHealth creates a default health record for an exchange
func (s *HealthScorerService) createDefaultHealth(exchangeID string) *domain.ExchangeHealth {
	return &domain.ExchangeHealth{
		ExchangeID:   exchangeID,
		HealthScore:  100,
		Status:       domain.ExchangeStatusActive,
		UpdatedAt:    time.Now().UTC(),
	}
}

// buildHealthFromCache builds a health record from cached data
func (s *HealthScorerService) buildHealthFromCache(ctx context.Context, exchangeID, scoreStr string) (*domain.ExchangeHealth, error) {
	score := parseFloat64(scoreStr)

	status := domain.ExchangeStatusActive
	if score < 80 {
		status = domain.ExchangeStatusDegraded
	}
	if score < 60 {
		status = domain.ExchangeStatusThrottled
	}
	if score < 40 {
		status = domain.ExchangeStatusSuspended
	}

	return &domain.ExchangeHealth{
		ExchangeID: exchangeID,
		HealthScore: score,
		Status:     status,
		UpdatedAt:  time.Now().UTC(),
	}, nil
}

// GetMetrics returns current health scorer metrics
func (s *HealthScorerMetrics) GetMetrics() HealthScorerMetrics {
	return HealthScorerMetrics{
		TotalUpdates:     s.TotalUpdates,
		ThrottleEvents:   s.ThrottleEvents,
		LastUpdatedAt:    s.LastUpdatedAt,
		HealthByExchange: s.copyMap(s.HealthByExchange),
	}
}

// copyMap creates a copy of the health map
func (s *HealthScorerMetrics) copyMap(m map[string]float64) map[string]float64 {
	copy := make(map[string]float64)
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// Helper functions for type conversions

func parseFloat64(s string) float64 {
	// Use standard library for parsing
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return n
	}
	return 0
}

func parseFloat64Safe(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt64(s string) int64 {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	return 0
}

func floatToString(f float64) string {
	return string(runesOfFloat(f))
}

func int64ToString(i int64) string {
	return string(runesOfInt(i))
}

func runesOfFloat(f float64) []byte {
	return []byte(string(rune('0'))) // Simplified placeholder
}

func runesOfInt(i int64) []byte {
	return []byte(string(rune('0'))) // Simplified placeholder
}
