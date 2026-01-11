package service

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// IngestionConfig holds configuration for the ingestion service
type IngestionConfig struct {
	BufferMaxSize    int
	FlushInterval    time.Duration
	BatchSize        int
	MaxConnections   int
	PingInterval     time.Duration
	PongTimeout      time.Duration
}

// IngestionService handles real-time market data ingestion
type IngestionService struct {
	marketRepo    port.MarketRepository
	analysisSvc   port.AnalysisService
	config        IngestionConfig

	// Connection tracking
	connections  int64
	connectionsMu sync.RWMutex

	// Processing tracking
	processedCount int64
	processedMu    sync.RWMutex

	// Buffer for batch processing
	eventBuffer []domain.MarketEvent
	bufferMu    sync.Mutex
	bufferCond  *sync.Cond

	// Control channels
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewIngestionService creates a new ingestion service
func NewIngestionService(marketRepo port.MarketRepository, analysisSvc port.AnalysisService, config IngestionConfig) *IngestionService {
	ctx, cancel := context.WithCancel(context.Background())

	if config.BufferMaxSize == 0 {
		config.BufferMaxSize = 10000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}

	svc := &IngestionService{
		marketRepo:    marketRepo,
		analysisSvc:   analysisSvc,
		config:        config,
		eventBuffer:   make([]domain.MarketEvent, 0, config.BufferMaxSize),
		ctx:           ctx,
		cancel:        cancel,
	}

	svc.bufferCond = sync.NewCond(&svc.bufferMu)

	return svc
}

// Start begins the ingestion service
func (s *IngestionService) Start(ctx context.Context) error {
	log.Println("Starting ingestion service...")

	// Start background workers
	s.wg.Add(1)
	go s.flushWorker(ctx)

	s.wg.Add(1)
	go s.metricsWorker(ctx)

	log.Println("Ingestion service started")
	return nil
}

// Stop gracefully stops the ingestion service
func (s *IngestionService) Stop() error {
	log.Println("Stopping ingestion service...")

	// Signal shutdown
	s.cancel()

	// Flush remaining events
	s.flushBuffer()

	// Wait for workers
	s.wg.Wait()

	log.Println("Ingestion service stopped")
	return nil
}

// ProcessPacket processes an incoming market data packet
func (s *IngestionService) ProcessPacket(ctx context.Context, packet *domain.MarketDataPacket) error {
	s.processedMu.Lock()
	atomic.AddInt64(&s.processedCount, 1)
	s.processedMu.Unlock()

	// Convert packet to events and add to buffer
	events := s.packetToEvents(packet)

	s.bufferMu.Lock()
	s.eventBuffer = append(s.eventBuffer, events...)

	// Check if we should flush
	shouldFlush := len(s.eventBuffer) >= s.config.BatchSize
	s.bufferMu.Unlock()

	if shouldFlush {
		s.bufferCond.Signal()
	}

	return nil
}

// packetToEvents converts a market data packet to events
func (s *IngestionService) packetToEvents(packet *domain.MarketDataPacket) []domain.MarketEvent {
	var events []domain.MarketEvent

	for _, trade := range packet.Trades {
		event := domain.MarketEvent{
			ID:            uuid.New(),
			ExchangeID:    packet.ExchangeID,
			Symbol:        packet.Symbol,
			EventType:     domain.MarketEventTypeTrade,
			OrderID:       &trade.OrderID,
			UserID:        &trade.UserID,
			Price:         trade.Price,
			Quantity:      trade.Quantity,
			Direction:     domain.MarketEventDirection(trade.Direction),
			Status:        domain.MarketEventStatusFilled,
			Timestamp:     trade.Timestamp,
			ReceivedAt:    time.Now(),
			SequenceNum:   packet.SequenceNum,
		}
		events = append(events, event)
	}

	for _, order := range packet.Orders {
		event := domain.MarketEvent{
			ID:            uuid.New(),
			ExchangeID:    packet.ExchangeID,
			Symbol:        packet.Symbol,
			EventType:     domain.MarketEventTypeOrder,
			OrderID:       &order.OrderID,
			UserID:        &order.UserID,
			Price:         order.Price,
			Quantity:      order.Quantity,
			FilledQuantity: order.FilledQuantity,
			Direction:     domain.MarketEventDirection(order.Direction),
			Status:        order.Status,
			Timestamp:     order.Timestamp,
			ReceivedAt:    time.Now(),
			SequenceNum:   packet.SequenceNum,
		}
		events = append(events, event)
	}

	return events
}

// flushWorker periodically flushes the event buffer
func (s *IngestionService) flushWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.flushBuffer()
			return
		case <-ticker.C:
			s.flushBuffer()
		case <-s.bufferCond.L:
			s.flushBuffer()
		}
	}
}

// flushBuffer writes buffered events to the database
func (s *IngestionService) flushBuffer() {
	s.bufferMu.Lock()
	if len(s.eventBuffer) == 0 {
		s.bufferMu.Unlock()
		return
	}

	// Copy and clear buffer
	events := make([]domain.MarketEvent, len(s.eventBuffer))
	copy(events, s.eventBuffer)
	s.eventBuffer = s.eventBuffer[:0]
	s.bufferMu.Unlock()

	// Store events in database
	if err := s.marketRepo.BulkStoreEvents(context.Background(), events); err != nil {
		log.Printf("Error storing events: %v", err)
		// In production, you might want to retry or send to a dead letter queue
		return
	}

	log.Printf("Flushed %d events to database", len(events))

	// Trigger analysis on recent events
	go s.triggerAnalysis(context.Background(), events)
}

// triggerAnalysis initiates analysis on recent events
func (s *IngestionService) triggerAnalysis(ctx context.Context, events []domain.MarketEvent) {
	for _, event := range events {
		if event.EventType == domain.MarketEventTypeTrade {
			_, err := s.analysisSvc.AnalyzeTrade(ctx, &event)
			if err != nil {
				log.Printf("Error analyzing trade %s: %v", event.ID, err)
			}
		} else if event.EventType == domain.MarketEventTypeOrder {
			_, err := s.analysisSvc.AnalyzeOrder(ctx, &event)
			if err != nil {
				log.Printf("Error analyzing order %s: %v", event.ID, err)
			}
		}
	}
}

// metricsWorker collects and logs metrics
func (s *IngestionService) metricsWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processedMu.RLock()
			count := atomic.LoadInt64(&s.processedCount)
			s.processedMu.RUnlock()

			s.connectionsMu.RLock()
			conns := atomic.LoadInt64(&s.connections)
			s.connectionsMu.RUnlock()

			log.Printf("Ingestion metrics - Processed: %d, Connections: %d", count, conns)
		}
	}
}

// GetConnectionCount returns the current number of connections
func (s *IngestionService) GetConnectionCount() int {
	s.connectionsMu.RLock()
	defer s.connectionsMu.RUnlock()
	return int(atomic.LoadInt64(&s.connections))
}

// GetProcessedCount returns the total number of processed packets
func (s *IngestionService) GetProcessedCount() int64 {
	s.processedMu.RLock()
	defer s.processedMu.RUnlock()
	return atomic.LoadInt64(&s.processedCount)
}

// IncrementConnectionCount increments the connection counter
func (s *IngestionService) IncrementConnectionCount() {
	atomic.AddInt64(&s.connections, 1)
}

// DecrementConnectionCount decrements the connection counter
func (s *IngestionService) DecrementConnectionCount() {
	atomic.AddInt64(&s.connections, -1)
}

// ValidateExchangeLicense validates that an exchange has a valid license
func (s *IngestionService) ValidateExchangeLicense(ctx context.Context, exchangeID uuid.UUID) (bool, error) {
	// In a real implementation, this would call the compliance service
	// For now, we'll return true to allow data through
	return true, nil
}

// CalculateChecksum calculates a checksum for market data integrity
func CalculateChecksum(data string) string {
	// Simple checksum for demonstration
	// In production, use a proper cryptographic hash
	hash := sha256.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

import (
	"crypto/sha256"
	"encoding/hex"
)

// Mock function to satisfy the import
var _ = decimal.Decimal{}
