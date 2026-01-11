package service

import (
	"context"
	"sync"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"go.uber.org/zap"
)

// IngestionServiceImpl implements the IngestionService interface
type IngestionServiceImpl struct {
	dataSourceRepo  ports.DataSourceRepository
	marketDataRepo  ports.MarketDataRepository
	statsRepo       ports.IngestionStatsRepository
	publisher       ports.EventPublisher
	factory         ports.ExchangeConnectorFactory
	logger          *zap.Logger

	// Runtime state
	connectors      map[string]ports.ExchangeConnector
	connectorMu     sync.RWMutex
	running         bool
	runMu           sync.Mutex
}

// NewIngestionService creates a new IngestionServiceImpl
func NewIngestionService(
	dataSourceRepo ports.DataSourceRepository,
	marketDataRepo ports.MarketDataRepository,
	statsRepo ports.IngestionStatsRepository,
	publisher ports.EventPublisher,
	factory ports.ExchangeConnectorFactory,
	logger *zap.Logger,
) *IngestionServiceImpl {
	return &IngestionServiceImpl{
		dataSourceRepo: dataSourceRepo,
		marketDataRepo: marketDataRepo,
		statsRepo:      statsRepo,
		publisher:      publisher,
		factory:        factory,
		logger:         logger,
		connectors:     make(map[string]ports.ExchangeConnector),
	}
}

// StartIngestion starts the ingestion process for all enabled data sources
func (s *IngestionServiceImpl) StartIngestion(ctx context.Context) error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	if s.running {
		return domain.ErrIngestionAlreadyRunning
	}

	s.logger.Info("Starting data ingestion for all enabled sources")

	// Get all enabled data sources
	sources, err := s.dataSourceRepo.FindEnabled(ctx)
	if err != nil {
		return err
	}

	s.running = true

	// Start ingestion for each enabled source
	for _, config := range sources {
		if err := s.startSource(ctx, config); err != nil {
			s.logger.Error("Failed to start ingestion for source",
				zap.String("source_id", config.ID.String()),
				zap.Error(err))
		}
	}

	return nil
}

// StopIngestion stops all active ingestion processes
func (s *IngestionServiceImpl) StopIngestion(ctx context.Context) error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	if !s.running {
		return domain.ErrIngestionNotRunning
	}

	s.logger.Info("Stopping all data ingestion")

	s.connectorMu.Lock()
	defer s.connectorMu.Unlock()

	for id, connector := range s.connectors {
		if err := connector.Disconnect(ctx); err != nil {
			s.logger.Error("Error disconnecting connector",
				zap.String("source_id", id),
				zap.Error(err))
		}
		s.dataSourceRepo.UpdateStatus(ctx, id, domain.ConnectionStatusDisconnected)
	}

	s.connectors = make(map[string]ports.ExchangeConnector)
	s.running = false

	return nil
}

// StartSource starts ingestion for a specific data source
func (s *IngestionServiceImpl) StartSource(ctx context.Context, sourceID string) error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	config, err := s.dataSourceRepo.FindByID(ctx, sourceID)
	if err != nil {
		return err
	}

	if !config.Enabled {
		return domain.ErrDataSourceDisabled
	}

	return s.startSource(ctx, config)
}

// startSource is the internal implementation without locking
func (s *IngestionServiceImpl) startSource(ctx context.Context, config *domain.DataSourceConfig) error {
	s.connectorMu.Lock()
	defer s.connectorMu.Unlock()

	// Check if already running
	if _, exists := s.connectors[config.ID.String()]; exists {
		return domain.ErrIngestionAlreadyRunning
	}

	connector, err := s.factory.CreateConnector(config)
	if err != nil {
		s.logger.Error("Failed to create connector",
			zap.String("source_id", config.ID.String()),
			zap.Error(err))
		return err
	}

	// Initialize stats
	stats := &domain.IngestionStats{
		SourceID:  config.ID.String(),
		Status:    domain.ConnectionStatusConnecting,
		TotalMessages: 0,
	}
	s.statsRepo.UpdateStats(ctx, stats)

	// Connect to the exchange
	if err := connector.Connect(ctx); err != nil {
		s.logger.Error("Failed to connect to exchange",
			zap.String("source_id", config.ID.String()),
			zap.Error(err))
		s.dataSourceRepo.UpdateStatus(ctx, config.ID.String(), domain.ConnectionStatusError)
		stats.Status = domain.ConnectionStatusError
		stats.LastError = err.Error()
		s.statsRepo.UpdateStats(ctx, stats)
		return err
	}

	// Create channels for data streaming
	dataChan := make(chan domain.MarketData, 1000)
	tradeChan := make(chan domain.Trade, 1000)

	// Start streaming data in a goroutine
	go s.streamData(ctx, config, connector, dataChan, tradeChan)

	// Subscribe to trades if supported
	if len(config.Symbols) > 0 {
		connector.SubscribeTrades(ctx, config.Symbols, tradeChan)
	}

	s.connectors[config.ID.String()] = connector
	s.dataSourceRepo.UpdateStatus(ctx, config.ID.String(), domain.ConnectionStatusConnected)
	s.logger.Info("Successfully started ingestion for source",
		zap.String("source_id", config.ID.String()),
		zap.String("name", config.Name))

	return nil
}

// StopSource stops ingestion for a specific data source
func (s *IngestionServiceImpl) StopSource(ctx context.Context, sourceID string) error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	s.connectorMu.Lock()
	defer s.connectorMu.Unlock()

	connector, exists := s.connectors[sourceID]
	if !exists {
		return nil // Already stopped
	}

	if err := connector.Disconnect(ctx); err != nil {
		s.logger.Error("Error disconnecting connector",
			zap.String("source_id", sourceID),
			zap.Error(err))
	}

	delete(s.connectors, sourceID)
	s.dataSourceRepo.UpdateStatus(ctx, sourceID, domain.ConnectionStatusDisconnected)

	s.logger.Info("Stopped ingestion for source", zap.String("source_id", sourceID))
	return nil
}

// streamData handles the continuous streaming of market data
func (s *IngestionServiceImpl) streamData(
	ctx context.Context,
	config *domain.DataSourceConfig,
	connector ports.ExchangeConnector,
	dataChan <-chan domain.MarketData,
	tradeChan <-chan domain.Trade,
) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context cancelled, stopping data stream",
				zap.String("source_id", config.ID.String()))
			return
		case data, ok := <-dataChan:
			if !ok {
				s.logger.Warn("Data channel closed",
					zap.String("source_id", config.ID.String()))
				return
			}
			s.processMarketData(ctx, config, &data)
		case trade, ok := <-tradeChan:
			if !ok {
				s.logger.Warn("Trade channel closed",
					zap.String("source_id", config.ID.String()))
				return
			}
			s.processTrade(ctx, config, &trade)
		}
	}
}

// processMarketData processes and stores market data
func (s *IngestionServiceImpl) processMarketData(
	ctx context.Context,
	config *domain.DataSourceConfig,
	data *domain.MarketData,
) {
	start := time.Now()

	// Store in repository
	if err := s.marketDataRepo.Store(ctx, data); err != nil {
		s.logger.Error("Failed to store market data",
			zap.String("source_id", config.ID.String()),
			zap.Error(err))
		s.recordFailure(ctx, config.ID.String())
		return
	}

	// Publish to message broker for other services
	if err := s.publisher.PublishMarketData(ctx, data); err != nil {
		s.logger.Warn("Failed to publish market data",
			zap.String("source_id", config.ID.String()),
			zap.Error(err))
	}

	// Update statistics
	s.recordSuccess(ctx, config.ID.String(), start)
}

// processTrade processes and stores trade data
func (s *IngestionServiceImpl) processTrade(
	ctx context.Context,
	config *domain.DataSourceConfig,
	trade *domain.Trade,
) {
	start := time.Now()

	// Store in repository
	if err := s.marketDataRepo.Store(ctx, &domain.MarketData{
		ID:          trade.ID,
		SourceID:    trade.SourceID,
		Symbol:      trade.Symbol,
		Price:       trade.Price,
		Volume:      trade.Quantity,
		Timestamp:   trade.Timestamp,
		CreatedAt:   trade.CreatedAt,
	}); err != nil {
		s.logger.Error("Failed to store trade",
			zap.String("source_id", config.ID.String()),
			zap.Error(err))
		s.recordFailure(ctx, config.ID.String())
		return
	}

	// Publish to message broker
	if err := s.publisher.PublishTrade(ctx, trade); err != nil {
		s.logger.Warn("Failed to publish trade",
			zap.String("source_id", config.ID.String()),
			zap.Error(err))
	}

	s.recordSuccess(ctx, config.ID.String(), start)
}

// recordSuccess records a successful message processing
func (s *IngestionServiceImpl) recordSuccess(ctx context.Context, sourceID string, start time.Time) {
	latency := time.Since(start)
	s.statsRepo.IncrementMessages(ctx, sourceID, 1, 0)
	s.statsRepo.RecordLatency(ctx, sourceID, latency)
}

// recordFailure records a failed message processing
func (s *IngestionServiceImpl) recordFailure(ctx context.Context, sourceID string) {
	s.statsRepo.IncrementMessages(ctx, sourceID, 0, 1)
}

// GetIngestionStatus returns the current status of all ingestion processes
func (s *IngestionServiceImpl) GetIngestionStatus(ctx context.Context) (*ports.IngestionStatusResponse, error) {
	sources, err := s.dataSourceRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]*ports.SourceStatusResponse, 0, len(sources))
	activeCount := 0

	for _, config := range sources {
		stats, err := s.statsRepo.GetStats(ctx, config.ID.String())
		if err != nil {
			stats = &domain.IngestionStats{
				SourceID: config.ID.String(),
				Status:   domain.ConnectionStatusDisconnected,
			}
		}

		status := &ports.SourceStatusResponse{
			SourceID:         config.ID.String(),
			Name:             config.Name,
			Status:           stats.Status,
			MessagesReceived: stats.ProcessedMessages,
			MessagesFailed:   stats.FailedMessages,
			Error:            stats.LastError,
		}
		if stats.LastMessageTime != nil {
			status.LastMessageTime = stats.LastMessageTime.Format(time.RFC3339)
		}

		statuses = append(statuses, status)

		if config.Enabled && stats.Status == domain.ConnectionStatusConnected {
			activeCount++
		}
	}

	overallStatus := "stopped"
	if s.running {
		if activeCount > 0 {
			overallStatus = "running"
		} else {
			overallStatus = "starting"
		}
	}

	return &ports.IngestionStatusResponse{
		OverallStatus:  overallStatus,
		ActiveSources:  activeCount,
		TotalSources:   len(sources),
		SourceStatuses: statuses,
	}, nil
}

// GetSourceStatus returns the status of a specific data source
func (s *IngestionServiceImpl) GetSourceStatus(ctx context.Context, sourceID string) (*ports.SourceStatusResponse, error) {
	config, err := s.dataSourceRepo.FindByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	stats, err := s.statsRepo.GetStats(ctx, sourceID)
	if err != nil {
		stats = &domain.IngestionStats{
			SourceID: sourceID,
			Status:   domain.ConnectionStatusDisconnected,
		}
	}

	status := &ports.SourceStatusResponse{
		SourceID:         sourceID,
		Name:             config.Name,
		Status:           stats.Status,
		MessagesReceived: stats.ProcessedMessages,
		MessagesFailed:   stats.FailedMessages,
		Error:            stats.LastError,
	}
	if stats.LastMessageTime != nil {
		status.LastMessageTime = stats.LastMessageTime.Format(time.RFC3339)
	}

	return status, nil
}

// ForceSync triggers an immediate data sync for a data source
func (s *IngestionServiceImpl) ForceSync(ctx context.Context, sourceID string) error {
	s.connectorMu.RLock()
	connector, exists := s.connectors[sourceID]
	s.connectorMu.RUnlock()

	if !exists {
		return domain.ErrNotConnected
	}

	snapshots, err := connector.FetchSnapshot(ctx, nil)
	if err != nil {
		return err
	}

	for _, data := range snapshots {
		s.processMarketData(ctx, &domain.DataSourceConfig{ID: uuid.MustParse(sourceID)}, data)
	}

	return nil
}

// GetStats returns ingestion statistics
func (s *IngestionServiceImpl) GetStats(ctx context.Context, sourceID string) (*domain.IngestionStats, error) {
	return s.statsRepo.GetStats(ctx, sourceID)
}
