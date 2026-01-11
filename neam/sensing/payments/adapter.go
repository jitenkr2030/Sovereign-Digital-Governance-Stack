package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/neamplatform/sensing/payments/config"
	"github.com/neamplatform/sensing/payments/idempotency"
	"github.com/neamplatform/sensing/payments/iso20022"
	"github.com/neamplatform/sensing/payments/kafka"
	"github.com/neamplatform/sensing/payments/swift"
	"go.uber.org/zap"
)

// PaymentAdapter represents the main payment adapter service
type PaymentAdapter struct {
	cfg            *config.Config
	logger         *zap.Logger
	kafkaConsumer  kafka.Consumer
	kafkaProducer  kafka.Producer
	idempotencyMgr *idempotency.Manager
	processors     map[string]MessageProcessor
	workerPool     *WorkerPool
	shutdownChan   chan struct{}
	wg             sync.WaitGroup
}

// MessageProcessor defines the interface for message format processors
type MessageProcessor interface {
	Parse(ctx context.Context, raw []byte) (*NormalizedPayment, error)
	Validate(msg interface{}) error
	GetMessageType() string
}

// NormalizedPayment represents the standardized internal payment structure
type NormalizedPayment struct {
	ID              string                 `json:"id"`
	MessageHash     string                 `json:"message_hash"`
	Format          string                 `json:"format"`
	MessageType     string                 `json:"message_type"`
	TransactionID   string                 `json:"transaction_id"`
	Amount          decimal.Decimal        `json:"amount"`
	Currency        string                 `json:"currency"`
	SenderBIC       string                 `json:"sender_bic"`
	ReceiverBIC     string                 `json:"receiver_bic"`
	SenderAccount   string                 `json:"sender_account"`
	ReceiverAccount string                 `json:"receiver_account"`
	ValueDate       time.Time              `json:"value_date"`
	ProcessingDate  time.Time              `json:"processing_date"`
	RawPayload      json.RawMessage        `json:"raw_payload"`
	Metadata        map[string]interface{} `json:"metadata"`
	Timestamp       time.Time              `json:"timestamp"`
}

// DLQMessage represents a message sent to the dead letter queue
type DLQMessage struct {
	OriginalPayload []byte                 `json:"original_payload"`
	Error           string                 `json:"error"`
	Timestamp       time.Time              `json:"timestamp"`
	RetryCount      int                    `json:"retry_count"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewPaymentAdapter creates a new payment adapter instance
func NewPaymentAdapter(cfg *config.Config, logger *zap.Logger) (*PaymentAdapter, error) {
	adapter := &PaymentAdapter{
		cfg:          cfg,
		logger:       logger,
		shutdownChan: make(chan struct{}),
		processors:   make(map[string]MessageProcessor),
	}

	if err := adapter.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return adapter, nil
}

// initializeComponents sets up all required components
func (p *PaymentAdapter) initializeComponents() error {
	var err error

	// Initialize Kafka consumer
	p.kafkaConsumer, err = kafka.NewConsumer(p.cfg.Kafka, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	// Initialize Kafka producer
	p.kafkaProducer, err = kafka.NewProducer(p.cfg.Kafka, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Initialize idempotency manager
	p.idempotencyMgr, err = idempotency.NewManager(p.cfg.Redis, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create idempotency manager: %w", err)
	}

	// Register message processors
	p.registerProcessors()

	// Initialize worker pool
	p.workerPool = NewWorkerPool(p.cfg.Processing.WorkerCount, p.logger)

	return nil
}

// registerProcessors registers all supported message format processors
func (p *PaymentAdapter) registerProcessors() {
	// ISO 20022 processors
	iso20022Processor := iso20022.NewProcessor(p.logger)
	p.processors["iso20022"] = iso20022Processor

	// SWIFT MT processor
	swiftMTProcessor := swift.NewMTProcessor(p.logger)
	p.processors["swift_mt"] = swiftMTProcessor

	// SWIFT MX processor (same as ISO 20022)
	p.processors["swift_mx"] = iso20022Processor

	p.logger.Info("Registered message processors",
		zap.Int("count", len(p.processors)))
}

// Start begins processing messages
func (p *PaymentAdapter) Start(ctx context.Context) error {
	p.logger.Info("Starting payment adapter service",
		zap.String("mode", p.cfg.Processing.Mode))

	// Start worker pool
	p.workerPool.Start(ctx)

	// Start message processing based on mode
	switch p.cfg.Processing.Mode {
	case "realtime":
		return p.startRealtimeProcessing(ctx)
	case "batch":
		return p.startBatchProcessing(ctx)
	case "hybrid":
		return p.startHybridProcessing(ctx)
	default:
		return fmt.Errorf("unknown processing mode: %s", p.cfg.Processing.Mode)
	}
}

// startRealtimeProcessing begins real-time Kafka message processing
func (p *PaymentAdapter) startRealtimeProcessing(ctx context.Context) error {
	handler := p.createMessageHandler()

	p.logger.Info("Starting real-time message processing",
		zap.String("topic", p.cfg.Kafka.InputTopic))

	return p.kafkaConsumer.Consume(ctx, p.cfg.Kafka.InputTopic, handler)
}

// startBatchProcessing begins batch file processing
func (p *PaymentAdapter) startBatchProcessing(ctx context.Context) error {
	p.logger.Info("Starting batch message processing",
		zap.String("batch_dir", p.cfg.Processing.BatchDirectory))

	batchProcessor := NewBatchProcessor(p.cfg, p.logger, p.idempotencyMgr, p.processors)
	return batchProcessor.Process(ctx)
}

// startHybridProcessing handles both real-time and batch processing
func (p *PaymentAdapter) startHybridProcessing(ctx context.Context) error {
	var wg sync.WaitGroup
	var err error

	// Start real-time processing in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		if rtErr := p.startRealtimeProcessing(ctx); rtErr != nil {
			p.logger.Error("Real-time processing error", zap.Error(rtErr))
		}
	}()

	// Start batch processing
	if batchErr := p.startBatchProcessing(ctx); batchErr != nil {
		err = batchErr
		p.logger.Error("Batch processing error", zap.Error(batchErr))
	}

	wg.Wait()
	return err
}

// createMessageHandler creates a handler for processing individual messages
func (p *PaymentAdapter) createMessageHandler() kafka.MessageHandler {
	return func(ctx context.Context, message []byte) error {
		return p.workerPool.Submit(func() {
			if err := p.processMessage(ctx, message); err != nil {
				p.logger.Error("Failed to process message",
					zap.Error(err),
					zap.ByteString("message", message))
			}
		})
	}
}

// processMessage handles a single message through the full processing pipeline
func (p *PaymentAdapter) processMessage(ctx context.Context, rawPayload []byte) error {
	startTime := time.Now()

	// Calculate message hash for idempotency
	messageHash := idempotency.CalculateHash(rawPayload)

	// Check idempotency
	result, err := p.idempotencyMgr.Check(ctx, messageHash)
	if err != nil {
		p.logger.Error("Idempotency check failed",
			zap.Error(err),
			zap.String("hash", messageHash))
		return fmt.Errorf("idempotency check error: %w", err)
	}

	// Handle based on idempotency result
	switch result.Status {
	case idempotency.StatusCompleted:
		p.logger.Debug("Duplicate message detected, skipping",
			zap.String("hash", messageHash))
		return nil // Idempotent: skip duplicate

	case idempotency.StatusProcessing:
		p.logger.Warn("Message currently being processed",
			zap.String("hash", messageHash))
		return nil // Already processing, skip to avoid duplicate

	case idempotency.StatusNotFound:
		// Proceed with processing
		break
	}

	// Mark as processing
	if err := p.idempotencyMgr.MarkProcessing(ctx, messageHash); err != nil {
		return fmt.Errorf("failed to mark message as processing: %w", err)
	}

	// Detect message format
	format, messageType := detectMessageFormat(rawPayload)
	processor, ok := p.processors[format]
	if !ok {
		p.sendToDLQ(ctx, rawPayload, fmt.Errorf("unsupported message format: %s", format), messageHash)
		return p.idempotencyMgr.MarkFailed(ctx, messageHash, "unsupported format")
	}

	// Parse the message
	normalized, err := processor.Parse(ctx, rawPayload)
	if err != nil {
		p.sendToDLQ(ctx, rawPayload, err, messageHash)
		return p.idempotencyMgr.MarkFailed(ctx, messageHash, err.Error())
	}

	normalized.MessageHash = messageHash
	normalized.Format = format
	normalized.MessageType = messageType
	normalized.Timestamp = startTime

	// Produce to output topic
	if err := p.kafkaProducer.Produce(ctx, p.cfg.Kafka.OutputTopic, normalized); err != nil {
		return fmt.Errorf("failed to produce normalized message: %w", err)
	}

	// Mark as completed
	if err := p.idempotencyMgr.MarkCompleted(ctx, messageHash); err != nil {
		p.logger.Error("Failed to mark message as completed",
			zap.Error(err),
			zap.String("hash", messageHash))
	}

	p.logger.Info("Message processed successfully",
		zap.String("hash", messageHash),
		zap.String("format", format),
		zap.String("message_type", messageType),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}

// sendToDLQ sends a message to the dead letter queue
func (p *PaymentAdapter) sendToDLQ(ctx context.Context, payload []byte, processingError error, messageHash string) {
	dlqMsg := DLQMessage{
		OriginalPayload: payload,
		Error:           processingError.Error(),
		Timestamp:       time.Now(),
		RetryCount:      0,
		Metadata: map[string]interface{}{
			"message_hash": messageHash,
		},
	}

	if err := p.kafkaProducer.Produce(ctx, p.cfg.Kafka.DLQTopic, dlqMsg); err != nil {
		p.logger.Error("Failed to send message to DLQ",
			zap.Error(err),
			zap.String("hash", messageHash))
	}
}

// detectMessageFormat identifies the message format from raw payload
func detectMessageFormat(rawPayload []byte) (format string, messageType string) {
	trimmed := bytes.TrimSpace(rawPayload)

	// Check for XML-based formats (ISO 20022 / SWIFT MX)
	if bytes.HasPrefix(trimmed, []byte("<?xml")) || bytes.HasPrefix(trimmed, []byte("<")) {
		// Check for SWIFT MX namespace
		if bytes.Contains(trimmed, []byte("xmlns=\"urn:iso:std:iso:20022:tech:xsd")) {
			return "swift_mx", extractMXMessageType(trimmed)
		}
		// ISO 20022
		if bytes.Contains(trimmed, []byte("Document")) {
			return "iso20022", extractISO20022MessageType(trimmed)
		}
		return "iso20022", "unknown"
	}

	// Check for SWIFT MT format (block structure)
	if isSWIFTMTFormat(trimmed) {
		return "swift_mt", extractMTMessageType(trimmed)
	}

	return "unknown", "unknown"
}

// extractMXMessageType extracts message type from SWIFT MX XML
func extractMXMessageType(payload []byte) string {
	// Implementation for extracting MX message type
	return "mx"
}

// extractISO20022MessageType extracts message type from ISO 20022 XML
func extractISO20022MessageType(payload []byte) string {
	// Implementation for extracting ISO 20022 message type
	return "iso20022"
}

// extractMTMessageType extracts message type from SWIFT MT
func extractMTMessageType(payload []byte) string {
	// Implementation for extracting MT message type
	return "mt"
}

// isSWIFTMTFormat checks if payload matches SWIFT MT format
func isSWIFTMTFormat(payload []byte) bool {
	// Check for SWIFT block structure
	return bytes.Contains(payload, []byte("{1:")) ||
		bytes.Contains(payload, []byte("{2:")) ||
		bytes.Contains(payload, []byte("{3:")) ||
		bytes.Contains(payload, []byte("{4:"))
}

// Stop gracefully shuts down the adapter
func (p *PaymentAdapter) Stop() error {
	p.logger.Info("Stopping payment adapter service")

	close(p.shutdownChan)

	// Wait for workers to complete
	p.wg.Wait()

	// Close connections
	if err := p.kafkaConsumer.Close(); err != nil {
		p.logger.Error("Error closing Kafka consumer", zap.Error(err))
	}

	if err := p.kafkaProducer.Close(); err != nil {
		p.logger.Error("Error closing Kafka producer", zap.Error(err))
	}

	p.logger.Info("Payment adapter service stopped")
	return nil
}

// Run starts the adapter and handles graceful shutdown
func Run(cfgPath string) error {
	// Load configuration
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger, err := config.InitializeLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	// Create adapter
	adapter, err := NewPaymentAdapter(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create payment adapter: %w", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Start adapter
	if err := adapter.Start(ctx); err != nil {
		return fmt.Errorf("adapter failed: %w", err)
	}

	return nil
}

// Import for bytes package
import "bytes"
