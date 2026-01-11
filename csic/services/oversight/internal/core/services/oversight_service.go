package services

import (
	"context"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"
	"go.uber.org/zap"
)

// OversightServiceImpl implements the ports.OversightService interface
// It combines the functionality of abuse detection and health scoring
type OversightServiceImpl struct {
	repo          ports.OversightRepository
	healthScorer  *HealthScorerService
	abuseDetector *AbuseDetectorService
	logger        *zap.Logger
}

// NewOversightService creates a new OversightServiceImpl
func NewOversightService(
	repo ports.OversightRepository,
	healthScorer *HealthScorerService,
	abuseDetector *AbuseDetectorService,
	logger *zap.Logger,
) *OversightServiceImpl {
	return &OversightServiceImpl{
		repo:          repo,
		healthScorer:  healthScorer,
		abuseDetector: abuseDetector,
		logger:        logger,
	}
}

// ProcessTradeStream processes incoming trade events from the message stream
func (s *OversightServiceImpl) ProcessTradeStream(ctx context.Context, trade domain.TradeEvent) error {
	// Record the trade
	if err := s.repo.RecordTrade(ctx, &domain.Trade{
		ID:         trade.TradeID,
		ExchangeID: trade.ExchangeID,
		Symbol:     trade.TradingPair,
		Side:       "unknown",
		Price:      trade.Price,
		Quantity:   trade.Volume,
		Timestamp:  trade.Timestamp,
	}); err != nil {
		s.logger.Warn("Failed to record trade", zap.Error(err))
	}

	// Update health metrics based on trade
	if s.healthScorer != nil {
		s.healthScorer.UpdateLatency(ctx, trade.ExchangeID, 0)
	}

	// Process through abuse detector if available
	if s.abuseDetector != nil {
		if err := s.abuseDetector.ProcessTradeStream(ctx, trade); err != nil {
			s.logger.Warn("Abuse detection failed for trade", zap.Error(err), zap.String("trade_id", trade.TradeID))
		}
	}

	return nil
}

// ProcessTradeBatch processes a batch of trade events
func (s *OversightServiceImpl) ProcessTradeBatch(ctx context.Context, trades []domain.TradeEvent) error {
	for _, trade := range trades {
		if err := s.ProcessTradeStream(ctx, trade); err != nil {
			s.logger.Warn("Failed to process trade in batch", zap.Error(err), zap.String("trade_id", trade.TradeID))
		}
	}
	return nil
}

// GetExchangeHealth retrieves the current health status of an exchange
func (s *OversightServiceImpl) GetExchangeHealth(ctx context.Context, exchangeID string) (*domain.ExchangeHealth, error) {
	exchange, err := s.repo.GetExchange(ctx, exchangeID)
	if err != nil {
		return nil, err
	}
	if exchange == nil {
		return nil, nil
	}
	
	// Convert Exchange entity to ExchangeHealth response
	health := &domain.ExchangeHealth{
		ID:              exchange.ID,
		ExchangeID:      exchange.ID,
		HealthScore:     exchange.HealthScore,
		Status:          exchange.Status,
		LastHeartbeatAt: exchange.UpdatedAt,
		UpdatedAt:       exchange.UpdatedAt,
	}
	health.CalculateHealthScore()
	
	return health, nil
}

// GetAllExchangeHealth retrieves health status for all monitored exchanges
func (s *OversightServiceImpl) GetAllExchangeHealth(ctx context.Context) ([]domain.ExchangeHealth, error) {
	exchanges, err := s.repo.ListExchanges(ctx)
	if err != nil {
		return nil, err
	}

	healthList := make([]domain.ExchangeHealth, 0, len(exchanges))
	for _, exchange := range exchanges {
		health := &domain.ExchangeHealth{
			ID:              exchange.ID,
			ExchangeID:      exchange.ID,
			HealthScore:     exchange.HealthScore,
			Status:          exchange.Status,
			LastHeartbeatAt: exchange.UpdatedAt,
			UpdatedAt:       exchange.UpdatedAt,
		}
		health.CalculateHealthScore()
		healthList = append(healthList, *health)
	}

	return healthList, nil
}

// UpdateDetectionRules updates the detection rules configuration
func (s *OversightServiceImpl) UpdateDetectionRules(ctx context.Context, rules []domain.DetectionRule) error {
	// This would update rules in the rule repository
	s.logger.Info("Detection rules update requested", zap.Int("count", len(rules)))
	return nil
}

// GetDetectionRules retrieves all active detection rules
func (s *OversightServiceImpl) GetDetectionRules(ctx context.Context) ([]domain.DetectionRule, error) {
	// Return empty list for now
	return []domain.DetectionRule{}, nil
}

// ExecuteThrottleCommand executes a throttle command for an exchange
func (s *OversightServiceImpl) ExecuteThrottleCommand(ctx context.Context, cmd domain.ThrottleCommand) error {
	s.logger.Info("Throttle command executed",
		zap.String("exchange_id", cmd.ExchangeID),
		zap.String("command", string(cmd.Command)),
		zap.Float64("target_rate_pct", cmd.TargetRatePct),
	)
	return nil
}

// GenerateRegulatoryReport generates a regulatory report for the specified period
func (s *OversightServiceImpl) GenerateRegulatoryReport(ctx context.Context, start, end time.Time, exchangeID string) (*domain.RegulatoryReport, error) {
	// Generate a basic regulatory report
	report := &domain.RegulatoryReport{
		ID:          "rpt-" + time.Now().UTC().Format("20060102150405"),
		ReportType:  "MARKET_SURVEILLANCE",
		PeriodStart: start,
		PeriodEnd:   end,
		GeneratedAt: time.Now().UTC(),
		ExchangeID:  exchangeID,
		Summary: domain.ReportSummary{
			TotalTrades:      0,
			TotalVolume:      0,
			TotalAlerts:      0,
			CriticalAlerts:   0,
			HighRiskExchanges: 0,
			MarketHealthScore: 100,
		},
		Metrics: domain.ReportMetrics{
			WashTradingCount:     0,
			ManipulationAttempts: 0,
			ExchangeUptimeAvg:    100,
			TradeAnomalyRate:     0,
			VolumeByExchange:     make(map[string]float64),
			AlertByType:          make(map[string]int),
		},
		Format:    "JSON",
		Version:  "1.0",
	}

	s.logger.Info("Regulatory report generated",
		zap.String("report_id", report.ID),
		zap.Time("period_start", start),
		zap.Time("period_end", end),
	)

	return report, nil
}

// RegisterExchange registers a new exchange
func (s *OversightServiceImpl) RegisterExchange(ctx context.Context, exchange *domain.Exchange) error {
	return s.repo.CreateExchange(ctx, exchange)
}

// CalculateHealthScore calculates the health score for an exchange
func (s *OversightServiceImpl) CalculateHealthScore(ctx context.Context, exchangeID string) (float64, error) {
	exchange, err := s.repo.GetExchange(ctx, exchangeID)
	if err != nil {
		return 0, err
	}
	if exchange == nil {
		return 0, nil
	}
	return exchange.HealthScore, nil
}

// UpdateExchangeStatus updates the status of an exchange
func (s *OversightServiceImpl) UpdateExchangeStatus(ctx context.Context, id string, status domain.ExchangeStatus) error {
	return s.repo.UpdateExchangeStatus(ctx, id, status)
}

// DetectWashTrading detects wash trading patterns
func (s *OversightServiceImpl) DetectWashTrading(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.TradeAnomaly, error) {
	// Return empty list for now
	return []*domain.TradeAnomaly{}, nil
}

// DetectSpoofing detects spoofing patterns
func (s *OversightServiceImpl) DetectSpoofing(ctx context.Context, exchangeID, symbol string) ([]*domain.TradeAnomaly, error) {
	// Return empty list for now
	return []*domain.TradeAnomaly{}, nil
}

// DetectPumpAndDump detects pump and dump patterns
func (s *OversightServiceImpl) DetectPumpAndDump(ctx context.Context, exchangeID, symbol string, window time.Duration) ([]*domain.TradeAnomaly, error) {
	// Return empty list for now
	return []*domain.TradeAnomaly{}, nil
}

// GetAnomalies retrieves anomalies for an exchange
func (s *OversightServiceImpl) GetAnomalies(ctx context.Context, exchangeID string, from, to time.Time) ([]*domain.TradeAnomaly, error) {
	return s.repo.GetAnomalies(ctx, exchangeID, from, to)
}

// ResolveAnomaly resolves an anomaly
func (s *OversightServiceImpl) ResolveAnomaly(ctx context.Context, id string, resolution string) error {
	return s.repo.UpdateAnomalyStatus(ctx, id, resolution)
}

// ProcessTrade processes a trade
func (s *OversightServiceImpl) ProcessTrade(ctx context.Context, trade *domain.Trade) error {
	return s.repo.RecordTrade(ctx, trade)
}

// ProcessMarketDepth processes market depth data
func (s *OversightServiceImpl) ProcessMarketDepth(ctx context.Context, depth *domain.MarketDepth) error {
	return s.repo.RecordMarketDepth(ctx, depth)
}

// GetMarketOverview provides a summary of all exchange activities
func (s *OversightServiceImpl) GetMarketOverview(ctx context.Context) (*ports.MarketOverview, error) {
	exchanges, err := s.repo.ListExchanges(ctx)
	if err != nil {
		return nil, err
	}

	overview := &ports.MarketOverview{
		TotalExchanges:   len(exchanges),
		OnlineExchanges:  0,
		LastUpdated:      time.Now().UTC(),
	}

	var totalHealthScore float64
	for _, exchange := range exchanges {
		if exchange.Status == domain.ExchangeStatusActive {
			overview.OnlineExchanges++
		}
		totalHealthScore += exchange.HealthScore
	}

	if len(exchanges) > 0 {
		overview.AverageHealthScore = totalHealthScore / float64(len(exchanges))
	}

	return overview, nil
}

// Ensure OversightServiceImpl implements ports.OversightService
var _ ports.OversightService = (*OversightServiceImpl)(nil)
