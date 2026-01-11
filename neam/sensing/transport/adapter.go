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
	"github.com/neamplatform/sensing/transport/config"
	"github.com/neamplatform/sensing/transport/kafka"
	"github.com/neamplatform/sensing/transport/mqtt"
	"github.com/neamplatform/sensing/transport/session"
	"github.com/neamplatform/sensing/transport/signalproc"
	"go.uber.org/zap"
)

// TransportAdapter represents the main transport adapter service
type TransportAdapter struct {
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
	fallbacks        *resilience.TransportFallbacks
}

// ProtocolClient defines the interface for transport protocol drivers
type ProtocolClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Subscribe(topicPatterns []string) (<-chan TelemetryPoint, error)
	HealthCheck() bool
	ID() string
	GetProtocolType() string
}

// TelemetryPoint represents a single data point from transport systems
type TelemetryPoint struct {
	ID         string          `json:"id"`
	SourceID   string          `json:"source_id"`
	PointID    string          `json:"point_id"`
	Value      interface{}     `json:"value"`
	Quality    string          `json:"quality"`
	Protocol   string          `json:"protocol"`
	Timestamp  time.Time       `json:"timestamp"`
	Criticality string         `json:"criticality"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

// NewTransportAdapter creates a new transport adapter instance
func NewTransportAdapter(cfg *config.Config, logger *zap.Logger) (*TransportAdapter, error) {
	adapter := &TransportAdapter{
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
func (t *TransportAdapter) initializeComponents() error {
	var err error

	// Initialize resilience components
	if err := t.initializeResilience(); err != nil {
		return fmt.Errorf("failed to initialize resilience: %w", err)
	}

	// Initialize Kafka producer
	t.kafkaProducer, err = kafka.NewProducer(t.cfg.Kafka, t.logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Initialize session manager
	t.sessionMgr = session.NewManager(t.cfg.Session, t.logger)

	// Initialize signal processor
	t.signalProc = signalproc.NewProcessor(t.cfg.Filtering, t.logger)

	// Initialize protocol drivers
	t.initializeProtocols()

	return nil
}

// initializeResilience sets up circuit breaker, DLQ, and retry policies
func (t *TransportAdapter) initializeResilience() error {
	// Create metrics collector
	metrics := resilience.NewMetricsCollector(t.logger)

	// Initialize circuit breaker
	circuitConfig := resilience.CircuitBreakerConfig{
		Enabled:              t.cfg.Resilience.CircuitBreaker.Enabled,
		Timeout:              time.Duration(t.cfg.Resilience.CircuitBreaker.Timeout),
		MaxRequests:          t.cfg.Resilience.CircuitBreaker.MaxRequests,
		Interval:             time.Duration(t.cfg.Resilience.CircuitBreaker.Interval),
		ReadyToTripThreshold: t.cfg.Resilience.CircuitBreaker.ReadyToTripThreshold,
		SuccessThreshold:     t.cfg.Resilience.CircuitBreaker.SuccessThreshold,
		RequestTimeout:       time.Duration(t.cfg.Resilience.CircuitBreaker.RequestTimeout),
	}

	t.circuitBreaker = resilience.NewCircuitBreakerManager(circuitConfig, metrics, t.logger)

	// Create fallback functions
	t.fallbacks = resilience.NewTransportFallbacks(t.logger)

	// Register circuit breakers for critical operations
	if err := t.circuitBreaker.RegisterCircuitBreaker(
		"mqtt-connection",
		t.fallbacks.CachedDataFallback,
	); err != nil {
		return fmt.Errorf("failed to register MQTT circuit breaker: %w", err)
	}

	if err := t.circuitBreaker.RegisterCircuitBreaker(
		"kafka-publish",
		t.fallbacks.DegradedModeFallback,
	); err != nil {
		return fmt.Errorf("failed to register Kafka circuit breaker: %w", err)
	}

	// Initialize DLQ producer
	dlqConfig := resilience.DLQConfig{
		Enabled:                t.cfg.Resilience.DLQ.Enabled,
		TopicSuffix:            t.cfg.Resilience.DLQ.TopicSuffix,
		RetryCount:             t.cfg.Resilience.DLQ.RetryCount,
		RetryDelay:             time.Duration(t.cfg.Resilience.DLQ.RetryDelay),
		MaxMessageSize:         t.cfg.Resilience.DLQ.MaxMessageSize,
		EnableSchemaValidation: t.cfg.Resilience.DLQ.EnableSchemaValidation,
		KafkaBrokers:           []string{t.cfg.Kafka.Brokers},
		ConsumerGroup:          "transport-dlq-consumer",
	}

	t.dlqProducer = resilience.NewDLQProducer(dlqConfig, metrics, t.logger)

	// Initialize retry policy
	retryConfig := resilience.RetryConfig{
		MaxAttempts:    t.cfg.Resilience.Retry.MaxAttempts,
		InitialDelay:   time.Duration(t.cfg.Resilience.Retry.InitialDelay),
		MaxDelay:       time.Duration(t.cfg.Resilience.Retry.MaxDelay),
		Multiplier:     t.cfg.Resilience.Retry.Multiplier,
		Jitter:         t.cfg.Resilience.Retry.Jitter,
		RetryableCodes: t.cfg.Resilience.Retry.RetryableCodes,
	}

	t.retryPolicy = resilience.NewRetryPolicy(retryConfig, metrics, t.logger)

	// Initialize schema validator
	t.schemaValidator = resilience.NewSchemaValidator(t.logger)

	return nil
}

// initializeProtocols creates and configures all protocol drivers
func (t *TransportAdapter) initializeProtocols() {
	// MQTT Protocol Drivers
	for _, mqttConfig := range t.cfg.MQTT.Sources {
		driver := mqtt.NewDriver(mqtt.DriverConfig{
			BrokerURL:       mqttConfig.BrokerURL,
			ClientID:        mqttConfig.ClientID,
			Username:        mqttConfig.Username,
			Password:        mqttConfig.Password,
			CleanStart:      mqttConfig.CleanStart,
			KeepAlive:       time.Duration(t.cfg.MQTT.KeepAlive),
			SessionExpiry:   time.Duration(t.cfg.MQTT.SessionExpiry),
			QoSByCriticality: t.cfg.MQTT.QoSByCriticality,
		}, t.logger)
		t.protocols = append(t.protocols, driver)
		t.logger.Info("Initialized MQTT driver",
			zap.String("broker", mqttConfig.BrokerURL),
			zap.String("client_id", mqttConfig.ClientID))
	}
}

// Start begins processing telemetry data
func (t *TransportAdapter) Start(ctx context.Context) error {
	t.logger.Info("Starting transport adapter service",
		zap.Int("protocol_count", len(t.protocols)))

	// Collect all topics of interest from configuration
	allTopics := make([]string, 0)
	for _, mqttConfig := range t.cfg.MQTT.Sources {
		allTopics = append(allTopics, mqttConfig.Topics...)
	}

	// Create merged telemetry stream from all protocols
	telemetryStream := t.sessionMgr.Start(ctx, t.protocols, allTopics)

	// Process and filter the stream
	processedStream := t.signalProc.Process(telemetryStream)

	// Publish to Kafka with resilience
	t.publishToKafka(ctx, processedStream)

	t.logger.Info("Transport adapter service started")
	return nil
}

// publishToKafka sends processed telemetry to Kafka with resilience
func (t *TransportAdapter) publishToKafka(ctx context.Context, stream <-chan TelemetryPoint) {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		for point := range stream {
			// Apply schema validation
			if t.cfg.Resilience.DLQ.EnableSchemaValidation {
				if err := t.validateTelemetry(point); err != nil {
					t.logger.Warn("Telemetry validation failed, sending to DLQ",
						zap.Error(err),
						zap.String("point_id", point.PointID))

					// Send to DLQ
					payload, _ := json.Marshal(point)
					t.dlqProducer.Publish(
						ctx,
						"transport",
						t.cfg.Kafka.OutputTopic,
						payload,
						resilience.ErrorReasonSchemaValidation,
						err,
						nil,
					)
					continue
				}
			}

			// Publish with retry and circuit breaker protection
			result := t.circuitBreaker.Execute(ctx, "kafka-publish", func(ctx context.Context) (interface{}, error) {
				return nil, t.kafkaProducer.Produce(ctx, t.cfg.Kafka.OutputTopic, point)
			})

			if result.Error != nil {
				t.logger.Error("Failed to publish telemetry",
					zap.Error(result.Error),
					zap.String("point_id", point.PointID))

				// Send failed message to DLQ
				payload, _ := json.Marshal(point)
				t.dlqProducer.Publish(
					ctx,
					"transport",
					t.cfg.Kafka.OutputTopic,
					payload,
					resilience.ErrorReasonProcessing,
					result.Error,
					nil,
				)
			}
		}
	}()
}

// validateTelemetry validates transport telemetry against schema
func (t *TransportAdapter) validateTelemetry(point TelemetryPoint) error {
	telemetry := &resilience.TransportTelemetry{
		Timestamp: point.Timestamp,
	}

	// Extract data from point
	if data, ok := point.Value.(map[string]interface{}); ok {
		if lat, ok := data["latitude"].(float64); ok {
			telemetry.Latitude = lat
		}
		if lng, ok := data["longitude"].(float64); ok {
			telemetry.Longitude = lng
		}
		if speed, ok := data["speed"].(float64); ok {
			telemetry.Speed = speed
		}
		if fuel, ok := data["fuel_level"].(float64); ok {
			telemetry.FuelLevel = fuel
		}
	}

	result := t.schemaValidator.Validate(context.Background(), telemetry)
	if !result.Valid {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// Stop gracefully shuts down the adapter
func (t *TransportAdapter) Stop() error {
	t.logger.Info("Stopping transport adapter service")

	close(t.shutdownChan)

	// Close resilience components
	if t.dlqProducer != nil {
		t.dlqProducer.Close()
	}
	if t.circuitBreaker != nil {
		t.circuitBreaker.Close()
	}

	// Disconnect all protocols
	for _, protocol := range t.protocols {
		if err := protocol.Disconnect(); err != nil {
			t.logger.Error("Error disconnecting protocol",
				zap.Error(err),
				zap.String("id", protocol.ID()))
		}
	}

	// Wait for goroutines
	t.wg.Wait()

	// Close connections
	if err := t.kafkaProducer.Close(); err != nil {
		t.logger.Error("Error closing Kafka producer", zap.Error(err))
	}

	t.logger.Info("Transport adapter service stopped")
	return nil
}

// Health returns the health status of the adapter
func (t *TransportAdapter) Health() map[string]interface{} {
	health := make(map[string]interface{})

	// Check protocol connections
	protocolHealth := make(map[string]interface{})
	for _, protocol := range t.protocols {
		protocolHealth[protocol.ID()] = map[string]interface{}{
			"connected": protocol.HealthCheck(),
			"protocol":  protocol.GetProtocolType(),
		}
	}
	health["protocols"] = protocolHealth

	// Check Kafka connection
	health["kafka"] = map[string]interface{}{
		"connected": t.kafkaProducer.HealthCheck(),
	}

	// Circuit breaker states
	if t.circuitBreaker != nil {
		health["circuit_breakers"] = t.circuitBreaker.GetAllStates()
	}

	// Session manager stats
	health["sessions"] = t.sessionMgr.Stats()

	return health
}

// DLQAdmin returns the DLQ admin interface for operational tasks
func (t *TransportAdapter) DLQAdmin() *resilience.DLQAdmin {
	if t.dlqProducer == nil {
		return nil
	}

	dlqConfig := resilience.DLQConfig{
		Enabled:        t.cfg.Resilience.DLQ.Enabled,
		TopicSuffix:    t.cfg.Resilience.DLQ.TopicSuffix,
		KafkaBrokers:   []string{t.cfg.Kafka.Brokers},
		ConsumerGroup:  "transport-dlq-consumer",
	}

	consumer := resilience.NewDLQConsumer(dlqConfig, nil, t.logger)
	metrics := resilience.NewMetricsCollector(t.logger)

	return resilience.NewDLQAdmin(t.dlqProducer, consumer, metrics, t.logger)
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
	adapter, err := NewTransportAdapter(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create transport adapter: %w", err)
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
		fmt.Fprintf(os.Stderr, "Transport adapter failed: %v\n", err)
		os.Exit(1)
	}
}
