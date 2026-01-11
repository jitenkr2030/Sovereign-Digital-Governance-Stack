package services

import (
	"context"
	"sync"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"go.uber.org/zap"
)

// IngestionService handles the processing of incoming trade events
// It coordinates the flow of trade data through the abuse detection
// and health scoring pipelines
type IngestionService struct {
	abuseDetector  ports.OversightService
	healthScorer   ports.OversightService
	tradeAnalytics ports.TradeAnalyticsEngine
	eventBus       ports.EventBus
	logger         *zap.Logger
	metrics        *IngestionMetrics
	mu             sync.RWMutex
}

// IngestionMetrics tracks ingestion performance
type IngestionMetrics struct {
	TotalProcessed   int64
	FailedProcessing int64
	ProcessingTime   time.Duration
	LastProcessedAt  time.Time
	ProcessedByExchange map[string]int64
	ProcessedByPair   map[string]int64
}

// NewIngestionService creates a new IngestionService
func NewIngestionService(
	abuseDetector ports.OversightService,
	healthScorer ports.OversightService,
	tradeAnalytics ports.TradeAnalyticsEngine,
	eventBus ports.EventBus,
	logger *zap.Logger,
) *IngestionService {
	return &IngestionService{
		abuseDetector:  abuseDetector,
		healthScorer:   healthScorer,
		tradeAnalytics: tradeAnalytics,
		eventBus:       eventBus,
		logger:         logger,
		metrics: &IngestionMetrics{
			ProcessedByExchange: make(map[string]int64),
			ProcessedByPair:     make(map[string]int64),
		},
	}
}

// ProcessTradeStream processes a single trade event
// This is the main entry point for trade data from Kafka
func (s *IngestionService) ProcessTradeStream(ctx context.Context, trade domain.TradeEvent) error {
	startTime := time.Now()

	// Validate trade event
	if err := s.validateTradeEvent(trade); err != nil {
		s.logger.Warn("Invalid trade event received",
			zap.String("trade_id", trade.TradeID),
			zap.String("exchange_id", trade.ExchangeID),
			zap.Error(err),
		)
		s.incrementFailedProcessing()
		return err
	}

	// Enrich trade with additional metadata
	s.enrichTradeEvent(&trade)

	// Process through abuse detector
	if err := s.abuseDetector.ProcessTradeStream(ctx, trade); err != nil {
		s.logger.Error("Failed to process trade through abuse detector",
			zap.String("trade_id", trade.TradeID),
			zap.Error(err),
		)
		// Continue processing even if abuse detection fails
	}

	// Update health score
	if err := s.healthScorer.ProcessTradeStream(ctx, trade); err != nil {
		s.logger.Error("Failed to update health score",
			zap.String("trade_id", trade.TradeID),
			zap.Error(err),
		)
	}

	// Index trade for analytics (async with error handling)
	go func() {
		if err := s.tradeAnalytics.IndexTrade(context.Background(), trade); err != nil {
			s.logger.Error("Failed to index trade for analytics",
				zap.String("trade_id", trade.TradeID),
				zap.Error(err),
			)
		}
	}()

	// Update metrics
	s.updateMetrics(trade)

	// Log processing completion
	processingTime := time.Since(startTime)
	s.logger.Debug("Trade event processed",
		zap.String("trade_id", trade.TradeID),
		zap.String("exchange_id", trade.ExchangeID),
		zap.String("trading_pair", trade.TradingPair),
		zap.Float64("price", trade.Price),
		zap.Float64("volume", trade.Volume),
		zap.Duration("processing_time", processingTime),
	)

	return nil
}

// ProcessTradeBatch processes multiple trade events efficiently
func (s *IngestionService) ProcessTradeBatch(ctx context.Context, trades []domain.TradeEvent) error {
	if len(trades) == 0 {
		return nil
	}

	startTime := time.Now()

	// Validate all trades first
	validTrades := make([]domain.TradeEvent, 0, len(trades))
	for _, trade := range trades {
		if err := s.validateTradeEvent(trade); err != nil {
			s.logger.Warn("Invalid trade in batch, skipping",
				zap.String("trade_id", trade.TradeID),
				zap.Error(err),
			)
			s.incrementFailedProcessing()
			continue
		}
		s.enrichTradeEvent(&trade)
		validTrades = append(validTrades, trade)
	}

	if len(validTrades) == 0 {
		return nil
	}

	// Process in parallel for better throughput
	var wg sync.WaitGroup

	// Process through abuse detector
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, trade := range validTrades {
			if err := s.abuseDetector.ProcessTradeStream(ctx, trade); err != nil {
				s.logger.Error("Failed to process trade through abuse detector",
					zap.String("trade_id", trade.TradeID),
					zap.Error(err),
				)
			}
		}
	}()

	// Update health scores
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, trade := range validTrades {
			if err := s.healthScorer.ProcessTradeStream(ctx, trade); err != nil {
				s.logger.Error("Failed to update health score",
					zap.String("trade_id", trade.TradeID),
					zap.Error(err),
				)
			}
		}
	}()

	// Batch index to analytics
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.tradeAnalytics.IndexTradeBatch(context.Background(), validTrades); err != nil {
			s.logger.Error("Failed to batch index trades for analytics",
				zap.Int("count", len(validTrades)),
				zap.Error(err),
			)
		}
	}()

	wg.Wait()

	// Update batch metrics
	s.updateBatchMetrics(validTrades)

	processingTime := time.Since(startTime)
	s.logger.Info("Trade batch processed",
		zap.Int("total_received", len(trades)),
		zap.Int("valid_trades", len(validTrades)),
		zap.Duration("processing_time", processingTime),
	)

	return nil
}

// validateTradeEvent validates a trade event for required fields
func (s *IngestionService) validateTradeEvent(trade domain.TradeEvent) error {
	if trade.TradeID == "" {
		return ErrMissingTradeID
	}
	if trade.ExchangeID == "" {
		return ErrMissingExchangeID
	}
	if trade.TradingPair == "" {
		return ErrMissingTradingPair
	}
	if trade.Price < 0 {
		return ErrInvalidPrice
	}
	if trade.Volume < 0 {
		return ErrInvalidVolume
	}
	if trade.Timestamp.IsZero() {
		return ErrMissingTimestamp
	}
	return nil
}

// enrichTradeEvent adds additional metadata to the trade event
func (s *IngestionService) enrichTradeEvent(trade *domain.TradeEvent) {
	// Set received timestamp if not set
	if trade.ReceivedAt.IsZero() {
		trade.ReceivedAt = time.Now().UTC()
	}

	// Calculate quote volume if not set
	if trade.QuoteVolume == 0 && trade.Price > 0 {
		trade.QuoteVolume = trade.Price * trade.Volume
	}
}

// updateMetrics updates ingestion metrics
func (s *IngestionService) updateMetrics(trade domain.TradeEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalProcessed++
	s.metrics.LastProcessedAt = time.Now()
	s.metrics.ProcessedByExchange[trade.ExchangeID]++
	s.metrics.ProcessedByPair[trade.TradingPair]++
}

// updateBatchMetrics updates metrics for a batch of trades
func (s *IngestionService) updateBatchMetrics(trades []domain.TradeEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, trade := range trades {
		s.metrics.TotalProcessed++
		s.metrics.ProcessedByExchange[trade.ExchangeID]++
		s.metrics.ProcessedByPair[trade.TradingPair]++
	}
	s.metrics.LastProcessedAt = time.Now()
}

// incrementFailedProcessing increments the failed processing counter
func (s *IngestionService) incrementFailedProcessing() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics.FailedProcessing++
}

// GetMetrics returns current ingestion metrics
func (s *IngestionService) GetMetrics() IngestionMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return IngestionMetrics{
		TotalProcessed:   s.metrics.TotalProcessed,
		FailedProcessing: s.metrics.FailedProcessing,
		ProcessingTime:   s.metrics.ProcessingTime,
		LastProcessedAt:  s.metrics.LastProcessedAt,
		ProcessedByExchange: s.copyMap(s.metrics.ProcessedByExchange),
		ProcessedByPair:     s.copyMap(s.metrics.ProcessedByPair),
	}
}

// copyMap creates a copy of the metrics map
func (s *IngestionService) copyMap(m map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// Processing errors
var (
	ErrMissingTradeID    = &ValidationError{Field: "TradeID", Message: "trade ID is required"}
	ErrMissingExchangeID = &ValidationError{Field: "ExchangeID", Message: "exchange ID is required"}
	ErrMissingTradingPair = &ValidationError{Field: "TradingPair", Message: "trading pair is required"}
	ErrInvalidPrice      = &ValidationError{Field: "Price", Message: "price must be non-negative"}
	ErrInvalidVolume     = &ValidationError{Field: "Volume", Message: "volume must be non-negative"}
	ErrMissingTimestamp  = &ValidationError{Field: "Timestamp", Message: "timestamp is required"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
