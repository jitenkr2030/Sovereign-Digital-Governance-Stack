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

	"neam-platform/sensing/common/resilience"
	"github.com/neamplatform/sensing/energy/config"
	"github.com/neamplatform/sensing/energy/iccpc"
	"github.com/neamplatform/sensing/energy/iec61850"
	"github.com/neamplatform/sensing/energy/kafka"
	"github.com/neamplatform/sensing/energy/session"
	"github.com/neamplatform/sensing/energy/signalproc"
	"go.uber.org/zap"
)

// EnergyAdapter represents the main energy adapter service
type EnergyAdapter struct {
	cfg              *config.Config
	logger           *zap.Logger
	kafkaProducer    kafka.Producer
	sessionMgr       *session.Manager
	signalProc       *signalproc.Processor
	protocols        []ProtocolClient
	shutdownChan     chan struct{}
	wg               sync.WaitGroup
	// Resilience components
	circuitBreaker   *resilience.CircuitBreakerManager
	dlqProducer      *resilience.DLQProducer
	schemaValidator  *resilience.SchemaValidator
	retryPolicy      *resilience.RetryPolicy
	fallbacks        *resilience.EnergyFallbacks
}

// ProtocolClient defines the interface for industrial protocol drivers
type ProtocolClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Subscribe(pointIDs []string) (<-chan TelemetryPoint, error)
	HealthCheck() bool
	ID() string
	GetProtocolType() string
}

// TelemetryPoint represents a single data point from the grid
type TelemetryPoint struct {
	ID         string          `json:"id"`
	SourceID   string          `json:"source_id"`
	PointID    string          `json:"point_id"`
	Value      float64         `json:"value"`
	Quality    string          `json:"quality"`
	Protocol   string          `json:"protocol"`
	Timestamp  time.Time       `json:"timestamp"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

// NewEnergyAdapter creates a new energy adapter instance
func NewEnergyAdapter(cfg *config.Config, logger *zap.Logger) (*EnergyAdapter, error) {
	adapter := &EnergyAdapter{
		cfg:          cfg,
		logger:       logger,
		shutdownChan: make(chan struct{}),
		protocols:    make([]ProtocolClient, 0),
	}

	if err := adapter.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return adapter, nil
}

// initializeComponents sets up all required components
func (e *EnergyAdapter) initializeComponents() error {
	var err error

	// Initialize resilience components
	if err := e.initializeResilience(); err != nil {
		return fmt.Errorf("failed to initialize resilience: %w", err)
	}

	// Initialize Kafka producer
	e.kafkaProducer, err = kafka.NewProducer(e.cfg.Kafka, e.logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Initialize session manager
	e.sessionMgr = session.NewManager(e.cfg.Session, e.logger)

	// Initialize signal processor
	e.signalProc = signalproc.NewProcessor(e.cfg.Filtering, e.logger)

	// Initialize protocol drivers
	e.initializeProtocols()

	return nil
}

// initializeResilience sets up circuit breaker, DLQ, and retry policies
func (e *EnergyAdapter) initializeResilience() error {
	// Create metrics collector
	metrics := resilience.NewMetricsCollector(e.logger)

	// Initialize circuit breaker
	circuitConfig := resilience.CircuitBreakerConfig{
		Enabled:              e.cfg.Resilience.CircuitBreaker.Enabled,
		Timeout:              time.Duration(e.cfg.Resilience.CircuitBreaker.Timeout),
		MaxRequests:          e.cfg.Resilience.CircuitBreaker.MaxRequests,
		Interval:             time.Duration(e.cfg.Resilience.CircuitBreaker.Interval),
		ReadyToTripThreshold: e.cfg.Resilience.CircuitBreaker.ReadyToTripThreshold,
		SuccessThreshold:     e.cfg.Resilience.CircuitBreaker.SuccessThreshold,
		RequestTimeout:       time.Duration(e.cfg.Resilience.CircuitBreaker.RequestTimeout),
	}

	e.circuitBreaker = resilience.NewCircuitBreakerManager(circuitConfig, metrics, e.logger)

	// Create fallback functions
	e.fallbacks = resilience.NewEnergyFallbacks(e.logger)

	// Register circuit breakers for critical operations
	if err := e.circuitBreaker.RegisterCircuitBreaker(
		"grid-connection",
		e.fallbacks.CachedDataFallback,
	); err != nil {
		return fmt.Errorf("failed to register grid protocol circuit breaker: %w", err)
	}

	if err := e.circuitBreaker.RegisterCircuitBreaker(
		"kafka-publish",
		e.fallbacks.DegradedModeFallback,
	); err != nil {
		return fmt.Errorf("failed to register Kafka circuit breaker: %w", err)
	}

	// Initialize DLQ producer
	dlqConfig := resilience.DLQConfig{
		Enabled:                e.cfg.Resilience.DLQ.Enabled,
		TopicSuffix:            e.cfg.Resilience.DLQ.TopicSuffix,
		RetryCount:             e.cfg.Resilience.DLQ.RetryCount,
		RetryDelay:             time.Duration(e.cfg.Resilience.DLQ.RetryDelay),
		MaxMessageSize:         e.cfg.Resilience.DLQ.MaxMessageSize,
		EnableSchemaValidation: e.cfg.Resilience.DLQ.EnableSchemaValidation,
		KafkaBrokers:           []string{e.cfg.Kafka.Brokers},
		ConsumerGroup:          "energy-dlq-consumer",
	}

	e.dlqProducer = resilience.NewDLQProducer(dlqConfig, metrics, e.logger)

	// Initialize retry policy
	retryConfig := resilience.RetryConfig{
		MaxAttempts:    e.cfg.Resilience.Retry.MaxAttempts,
		InitialDelay:   time.Duration(e.cfg.Resilience.Retry.InitialDelay),
		MaxDelay:       time.Duration(e.cfg.Resilience.Retry.MaxDelay),
		Multiplier:     e.cfg.Resilience.Retry.Multiplier,
		Jitter:         e.cfg.Resilience.Retry.Jitter,
		RetryableCodes: e.cfg.Resilience.Retry.RetryableCodes,
	}

	e.retryPolicy = resilience.NewRetryPolicy(retryConfig, metrics, e.logger)

	// Initialize schema validator
	e.schemaValidator = resilience.NewSchemaValidator(e.logger)

	return nil
}

// initializeProtocols creates and configures all protocol drivers
func (e *EnergyAdapter) initializeProtocols() {
	// ICCP Protocol Drivers
	for _, iccpConfig := range e.cfg.ICCP.Sources {
		driver := iccpc.NewDriver(iccpc.DriverConfig{
			RemoteEndPoint:   iccpConfig.EndPoint,
			LocalTSAP:        iccpConfig.LocalTSAP,
			RemoteTSAP:       iccpConfig.RemoteTSAP,
			BilateralTable:  iccpConfig.BilateralTable,
			Timeout:          time.Duration(e.cfg.ICCP.Timeout),
			RetryInterval:    time.Duration(e.cfg.ICCP.RetryInterval),
		}, e.logger)
		e.protocols = append(e.protocols, driver)
		e.logger.Info("Initialized ICCP driver",
			zap.String("endpoint", iccpConfig.EndPoint))
	}

	// IEC 61850 Protocol Drivers
	for _, iecConfig := range e.cfg.IEC61850.Sources {
		driver := iec61850.NewDriver(iec61850.DriverConfig{
			EndPoint:     iecConfig.EndPoint,
			Port:         iecConfig.Port,
			IEDName:      iecConfig.IEDName,
			APTitle:      iecConfig.APTitle,
			AEQualifier:  iecConfig.AEQualifier,
			Timeout:      time.Duration(e.cfg.IEC61850.Timeout),
			RetryInterval: time.Duration(e.cfg.IEC61850.RetryInterval),
		}, e.logger)
		e.protocols = append(e.protocols, driver)
		e.logger.Info("Initialized IEC 61850 driver",
			zap.String("endpoint", iecConfig.EndPoint),
			zap.String("ied", iecConfig.IEDName))
	}
}

// Start begins processing telemetry data
func (e *EnergyAdapter) Start(ctx context.Context) error {
	e.logger.Info("Starting energy adapter service",
		zap.Int("protocol_count", len(e.protocols)))

	// Collect all points of interest from configuration
	allPoints := make([]string, 0)
	allPoints = append(allPoints, e.cfg.ICCP.PointsOfInterest...)
	allPoints = append(allPoints, e.cfg.IEC61850.PointsOfInterest...)

	// Create merged telemetry stream from all protocols
	telemetryStream := e.sessionMgr.Start(ctx, e.protocols, allPoints)

	// Process and filter the stream
	processedStream := e.signalProc.Process(telemetryStream)

	// Publish to Kafka
	e.publishToKafka(ctx, processedStream)

	e.logger.Info("Energy adapter service started")
	return nil
}

// publishToKafka sends processed telemetry to Kafka with resilience
func (e *EnergyAdapter) publishToKafka(ctx context.Context, stream <-chan TelemetryPoint) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		for point := range stream {
			// Apply schema validation
			if e.cfg.Resilience.DLQ.EnableSchemaValidation {
				if err := e.validateTelemetry(point); err != nil {
					e.logger.Warn("Telemetry validation failed, sending to DLQ",
						zap.Error(err),
						zap.String("point_id", point.PointID))

					// Send to DLQ
					payload, _ := json.Marshal(point)
					e.dlqProducer.Publish(
						ctx,
						"energy",
						e.cfg.Kafka.OutputTopic,
						payload,
						resilience.ErrorReasonSchemaValidation,
						err,
						nil,
					)
					continue
				}
			}

			// Publish with retry and circuit breaker protection
			result := e.circuitBreaker.Execute(ctx, "kafka-publish", func(ctx context.Context) (interface{}, error) {
				return nil, e.kafkaProducer.Produce(ctx, e.cfg.Kafka.OutputTopic, point)
			})

			if result.Error != nil {
				e.logger.Error("Failed to publish telemetry",
					zap.Error(result.Error),
					zap.String("point_id", point.PointID))

				// Send failed message to DLQ
				payload, _ := json.Marshal(point)
				e.dlqProducer.Publish(
					ctx,
					"energy",
					e.cfg.Kafka.OutputTopic,
					payload,
					resilience.ErrorReasonProcessing,
					result.Error,
					nil,
				)
			}
		}
	}()
}

// validateTelemetry validates energy telemetry against schema
func (e *EnergyAdapter) validateTelemetry(point TelemetryPoint) error {
	telemetry := &resilience.EnergyTelemetry{
		Timestamp: point.Timestamp,
		PointID:   point.PointID,
		Value:     point.Value,
	}

	result := e.schemaValidator.Validate(context.Background(), telemetry)
	if !result.Valid {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// Stop gracefully shuts down the adapter
func (e *EnergyAdapter) Stop() error {
	e.logger.Info("Stopping energy adapter service")

	close(e.shutdownChan)

	// Close resilience components
	if e.dlqProducer != nil {
		e.dlqProducer.Close()
	}
	if e.circuitBreaker != nil {
		e.circuitBreaker.Close()
	}

	// Disconnect all protocols
	for _, protocol := range e.protocols {
		if err := protocol.Disconnect(); err != nil {
			e.logger.Error("Error disconnecting protocol",
				zap.Error(err),
				zap.String("id", protocol.ID()))
		}
	}

	// Wait for goroutines
	e.wg.Wait()

	// Close connections
	if err := e.kafkaProducer.Close(); err != nil {
		e.logger.Error("Error closing Kafka producer", zap.Error(err))
	}

	e.logger.Info("Energy adapter service stopped")
	return nil
}

// Health returns the health status of the adapter
func (e *EnergyAdapter) Health() map[string]interface{} {
	health := make(map[string]interface{})

	// Check protocol connections
	protocolHealth := make(map[string]interface{})
	for _, protocol := range e.protocols {
		protocolHealth[protocol.ID()] = map[string]interface{}{
			"connected": protocol.HealthCheck(),
			"protocol":  protocol.GetProtocolType(),
		}
	}
	health["protocols"] = protocolHealth

	// Check Kafka connection
	health["kafka"] = map[string]interface{}{
		"connected": e.kafkaProducer.HealthCheck(),
	}

	// Circuit breaker states
	if e.circuitBreaker != nil {
		health["circuit_breakers"] = e.circuitBreaker.GetAllStates()
	}

	// Session manager stats
	health["sessions"] = e.sessionMgr.Stats()

	return health
}

// DLQAdmin returns the DLQ admin interface for operational tasks
func (e *EnergyAdapter) DLQAdmin() *resilience.DLQAdmin {
	if e.dlqProducer == nil {
		return nil
	}

	dlqConfig := resilience.DLQConfig{
		Enabled:        e.cfg.Resilience.DLQ.Enabled,
		TopicSuffix:    e.cfg.Resilience.DLQ.TopicSuffix,
		KafkaBrokers:   []string{e.cfg.Kafka.Brokers},
		ConsumerGroup:  "energy-dlq-consumer",
	}

	consumer := resilience.NewDLQConsumer(dlqConfig, nil, e.logger)
	metrics := resilience.NewMetricsCollector(e.logger)

	return resilience.NewDLQAdmin(e.dlqProducer, consumer, metrics, e.logger)
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
	adapter, err := NewEnergyAdapter(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create energy adapter: %w", err)
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

	// Wait for shutdown
	<-ctx.Done()
	return adapter.Stop()
}

func main() {
	// Determine config path
	cfgPath := "config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	if err := Run(cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "Energy adapter failed: %v\n", err)
		os.Exit(1)
	}
}
