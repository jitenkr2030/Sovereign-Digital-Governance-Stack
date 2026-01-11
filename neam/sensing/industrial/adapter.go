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
	"github.com/neamplatform/sensing/industrial/config"
	"github.com/neamplatform/sensing/industrial/kafka"
	"github.com/neamplatform/sensing/industrial/opcua"
	"github.com/neamplatform/sensing/industrial/session"
	"github.com/neamplatform/sensing/industrial/signalproc"
	"go.uber.org/zap"
)

// IndustrialAdapter represents the main industrial adapter service
type IndustrialAdapter struct {
	cfg              *config.Config
	logger           *zap.Logger
	kafkaProducer    kafka.Producer
	sessionMgr       *session.Manager
	signalProc       *signalproc.Processor
	protocols        []ProtocolClient
	discovery        *AssetDiscovery
	shutdownChan     chan struct{}
	wg               sync.WaitGroup
	// Resilience components
	circuitBreaker   *resilience.CircuitBreakerManager
	dlqProducer      *resilience.DLQProducer
	schemaValidator  *resilience.SchemaValidator
	retryPolicy      *resilience.RetryPolicy
	fallbacks        *resilience.IndustrialFallbacks
}

// ProtocolClient defines the interface for industrial protocol drivers
type ProtocolClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Subscribe(nodeIDs []string) (<-chan TelemetryPoint, error)
	Browse(ctx context.Context, nodeID string) ([]NodeInfo, error)
	HealthCheck() bool
	ID() string
	GetProtocolType() string
}

// TelemetryPoint represents a single data point from industrial systems
type TelemetryPoint struct {
	ID          string          `json:"id"`
	SourceID    string          `json:"source_id"`
	NodeID      string          `json:"node_id"`
	AssetID     string          `json:"asset_id"`
	AssetName   string          `json:"asset_name"`
	Value       interface{}     `json:"value"`
	Quality     string          `json:"quality"`
	Protocol    string          `json:"protocol"`
	Timestamp   time.Time       `json:"timestamp"`
	DataType    string          `json:"data_type"`
	Unit        string          `json:"unit"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

// NodeInfo represents information about an OPC UA node
type NodeInfo struct {
	NodeID      string `json:"node_id"`
	BrowseName  string `json:"browse_name"`
	DisplayName string `json:"display_name"`
	DataType    string `json:"data_type"`
	AccessLevel string `json:"access_level"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id,omitempty"`
	HasChildren bool   `json:"has_children"`
}

// AssetInfo represents information about a discovered industrial asset
type AssetInfo struct {
	AssetID      string            `json:"asset_id"`
	AssetName    string            `json:"asset_name"`
	AssetType    string            `json:"asset_type"`
	Manufacturer string            `json:"manufacturer"`
	Model        string            `json:"model"`
	SerialNumber string            `json:"serial_number"`
	Nodes        []NodeInfo        `json:"nodes"`
	Metadata     map[string]string `json:"metadata"`
}

// NewIndustrialAdapter creates a new industrial adapter instance
func NewIndustrialAdapter(cfg *config.Config, logger *zap.Logger) (*IndustrialAdapter, error) {
	adapter := &IndustrialAdapter{
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
func (i *IndustrialAdapter) initializeComponents() error {
	var err error

	// Initialize resilience components
	if err := i.initializeResilience(); err != nil {
		return fmt.Errorf("failed to initialize resilience: %w", err)
	}

	// Initialize Kafka producer
	i.kafkaProducer, err = kafka.NewProducer(i.cfg.Kafka, i.logger)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Initialize session manager
	i.sessionMgr = session.NewManager(i.cfg.Session, i.logger)

	// Initialize signal processor
	i.signalProc = signalproc.NewProcessor(i.cfg.Filtering, i.logger)

	// Initialize protocol drivers
	i.initializeProtocols()

	// Initialize asset discovery
	i.discovery = NewAssetDiscovery(i.cfg.Discovery, i.logger)

	return nil
}

// initializeResilience sets up circuit breaker, DLQ, and retry policies
func (i *IndustrialAdapter) initializeResilience() error {
	// Create metrics collector
	metrics := resilience.NewMetricsCollector(i.logger)

	// Initialize circuit breaker
	circuitConfig := resilience.CircuitBreakerConfig{
		Enabled:              i.cfg.Resilience.CircuitBreaker.Enabled,
		Timeout:              time.Duration(i.cfg.Resilience.CircuitBreaker.Timeout),
		MaxRequests:          i.cfg.Resilience.CircuitBreaker.MaxRequests,
		Interval:             time.Duration(i.cfg.Resilience.CircuitBreaker.Interval),
		ReadyToTripThreshold: i.cfg.Resilience.CircuitBreaker.ReadyToTripThreshold,
		SuccessThreshold:     i.cfg.Resilience.CircuitBreaker.SuccessThreshold,
		RequestTimeout:       time.Duration(i.cfg.Resilience.CircuitBreaker.RequestTimeout),
	}

	i.circuitBreaker = resilience.NewCircuitBreakerManager(circuitConfig, metrics, i.logger)

	// Create fallback functions
	i.fallbacks = resilience.NewIndustrialFallbacks(i.logger)

	// Register circuit breakers for critical operations
	if err := i.circuitBreaker.RegisterCircuitBreaker(
		"opcua-connection",
		i.fallbacks.CachedDataFallback,
	); err != nil {
		return fmt.Errorf("failed to register OPC UA circuit breaker: %w", err)
	}

	if err := i.circuitBreaker.RegisterCircuitBreaker(
		"kafka-publish",
		i.fallbacks.DegradedModeFallback,
	); err != nil {
		return fmt.Errorf("failed to register Kafka circuit breaker: %w", err)
	}

	// Initialize DLQ producer
	dlqConfig := resilience.DLQConfig{
		Enabled:                i.cfg.Resilience.DLQ.Enabled,
		TopicSuffix:            i.cfg.Resilience.DLQ.TopicSuffix,
		RetryCount:             i.cfg.Resilience.DLQ.RetryCount,
		RetryDelay:             time.Duration(i.cfg.Resilience.DLQ.RetryDelay),
		MaxMessageSize:         i.cfg.Resilience.DLQ.MaxMessageSize,
		EnableSchemaValidation: i.cfg.Resilience.DLQ.EnableSchemaValidation,
		KafkaBrokers:           []string{i.cfg.Kafka.Brokers},
		ConsumerGroup:          "industrial-dlq-consumer",
	}

	i.dlqProducer = resilience.NewDLQProducer(dlqConfig, metrics, i.logger)

	// Initialize retry policy
	retryConfig := resilience.RetryConfig{
		MaxAttempts:    i.cfg.Resilience.Retry.MaxAttempts,
		InitialDelay:   time.Duration(i.cfg.Resilience.Retry.InitialDelay),
		MaxDelay:       time.Duration(i.cfg.Resilience.Retry.MaxDelay),
		Multiplier:     i.cfg.Resilience.Retry.Multiplier,
		Jitter:         i.cfg.Resilience.Retry.Jitter,
		RetryableCodes: i.cfg.Resilience.Retry.RetryableCodes,
	}

	i.retryPolicy = resilience.NewRetryPolicy(retryConfig, metrics, i.logger)

	// Initialize schema validator
	i.schemaValidator = resilience.NewSchemaValidator(i.logger)

	return nil
}

// initializeProtocols creates and configures all protocol drivers
func (i *IndustrialAdapter) initializeProtocols() {
	// OPC UA Protocol Drivers
	for _, opcConfig := range i.cfg.OPCUA.Sources {
		driver := opcua.NewDriver(opcua.DriverConfig{
			EndpointURL:      opcConfig.EndpointURL,
			ApplicationURI:   opcConfig.ApplicationURI,
			ProductURI:       opcConfig.ProductURI,
			ApplicationName:  opcConfig.ApplicationName,
			SecurityPolicy:   opcConfig.SecurityPolicy,
			SecurityMode:     opcConfig.SecurityMode,
			AuthMode:         opcConfig.AuthMode,
			Username:         opcConfig.Username,
			Password:         opcConfig.Password,
			CertificatePath:  opcConfig.CertificatePath,
			PrivateKeyPath:   opcConfig.PrivateKeyPath,
			Timeout:          time.Duration(i.cfg.OPCUA.Timeout),
			RetryInterval:    time.Duration(i.cfg.OPCUA.RetryInterval),
			ReconnectWait:    time.Duration(i.cfg.OPCUA.ReconnectWait),
			SubscriptionRate: time.Duration(i.cfg.OPCUA.SubscriptionRate),
		}, i.logger)
		i.protocols = append(i.protocols, driver)
		i.logger.Info("Initialized OPC UA driver",
			zap.String("endpoint", opcConfig.EndpointURL),
			zap.String("ied", opcConfig.EndpointURL))
	}
}

// Start begins processing telemetry data
func (i *IndustrialAdapter) Start(ctx context.Context) error {
	i.logger.Info("Starting industrial adapter service",
		zap.Int("protocol_count", len(i.protocols)))

	// Collect all nodes of interest from configuration
	allNodes := make([]string, 0)
	for _, opcConfig := range i.cfg.OPCUA.Sources {
		allNodes = append(allNodes, opcConfig.NodesOfInterest...)
	}

	// Discover assets first
	if i.cfg.Discovery.Enabled {
		assets, err := i.DiscoverAssets(ctx)
		if err != nil {
			i.logger.Warn("Asset discovery failed", zap.Error(err))
		} else {
			i.logger.Info("Discovered assets", zap.Int("count", len(assets)))
			// Add discovered nodes to monitoring list
			for _, asset := range assets {
				for _, node := range asset.Nodes {
					allNodes = append(allNodes, node.NodeID)
				}
			}
		}
	}

	// Create merged telemetry stream from all protocols
	telemetryStream := i.sessionMgr.Start(ctx, i.protocols, allNodes)

	// Process and filter the stream
	processedStream := i.signalProc.Process(telemetryStream)

	// Publish to Kafka
	i.publishToKafka(ctx, processedStream)

	i.logger.Info("Industrial adapter service started")
	return nil
}

// publishToKafka sends processed telemetry to Kafka with resilience
func (i *IndustrialAdapter) publishToKafka(ctx context.Context, stream <-chan TelemetryPoint) {
	i.wg.Add(1)
	go func() {
		defer i.wg.Done()

		for point := range stream {
			// Apply schema validation
			if i.cfg.Resilience.DLQ.EnableSchemaValidation {
				if err := i.validateTelemetry(point); err != nil {
					i.logger.Warn("Telemetry validation failed, sending to DLQ",
						zap.Error(err),
						zap.String("node_id", point.NodeID))

					// Send to DLQ
					payload, _ := json.Marshal(point)
					i.dlqProducer.Publish(
						ctx,
						"industrial",
						i.cfg.Kafka.OutputTopic,
						payload,
						resilience.ErrorReasonSchemaValidation,
						err,
						nil,
					)
					continue
				}
			}

			// Publish with retry and circuit breaker protection
			result := i.circuitBreaker.Execute(ctx, "kafka-publish", func(ctx context.Context) (interface{}, error) {
				return nil, i.kafkaProducer.Produce(ctx, i.cfg.Kafka.OutputTopic, point)
			})

			if result.Error != nil {
				i.logger.Error("Failed to publish telemetry",
					zap.Error(result.Error),
					zap.String("node_id", point.NodeID))

				// Send failed message to DLQ
				payload, _ := json.Marshal(point)
				i.dlqProducer.Publish(
					ctx,
					"industrial",
					i.cfg.Kafka.OutputTopic,
					payload,
					resilience.ErrorReasonProcessing,
					result.Error,
					nil,
				)
			}
		}
	}()
}

// validateTelemetry validates industrial telemetry against schema
func (i *IndustrialAdapter) validateTelemetry(point TelemetryPoint) error {
	telemetry := &resilience.IndustrialTelemetry{
		Timestamp: point.Timestamp,
	}

	// Extract data from point
	telemetry.NodeID = point.NodeID
	telemetry.AssetID = point.AssetID
	telemetry.DataType = point.DataType

	if val, ok := point.Value.(float64); ok {
		telemetry.NumericValue = val
	}

	result := i.schemaValidator.Validate(context.Background(), telemetry)
	if !result.Valid {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}

// DiscoverAssets performs industrial asset discovery
func (i *IndustrialAdapter) DiscoverAssets(ctx context.Context) ([]AssetInfo, error) {
	return i.discovery.Discover(ctx, i.protocols)
}

// Stop gracefully shuts down the adapter
func (i *IndustrialAdapter) Stop() error {
	i.logger.Info("Stopping industrial adapter service")

	close(i.shutdownChan)

	// Close resilience components
	if i.dlqProducer != nil {
		i.dlqProducer.Close()
	}
	if i.circuitBreaker != nil {
		i.circuitBreaker.Close()
	}

	// Disconnect all protocols
	for _, protocol := range i.protocols {
		if err := protocol.Disconnect(); err != nil {
			i.logger.Error("Error disconnecting protocol",
				zap.Error(err),
				zap.String("id", protocol.ID()))
		}
	}

	// Wait for goroutines
	i.wg.Wait()

	// Close connections
	if err := i.kafkaProducer.Close(); err != nil {
		i.logger.Error("Error closing Kafka producer", zap.Error(err))
	}

	i.logger.Info("Industrial adapter service stopped")
	return nil
}

// Health returns the health status of the adapter
func (i *IndustrialAdapter) Health() map[string]interface{} {
	health := make(map[string]interface{})

	// Check protocol connections
	protocolHealth := make(map[string]interface{})
	for _, protocol := range i.protocols {
		protocolHealth[protocol.ID()] = map[string]interface{}{
			"connected": protocol.HealthCheck(),
			"protocol":  protocol.GetProtocolType(),
		}
	}
	health["protocols"] = protocolHealth

	// Check Kafka connection
	health["kafka"] = map[string]interface{}{
		"connected": i.kafkaProducer.HealthCheck(),
	}

	// Circuit breaker states
	if i.circuitBreaker != nil {
		health["circuit_breakers"] = i.circuitBreaker.GetAllStates()
	}

	// Session manager stats
	health["sessions"] = i.sessionMgr.Stats()

	// Discovery status
	health["discovery"] = map[string]interface{}{
		"enabled": i.cfg.Discovery.Enabled,
	}

	return health
}

// DLQAdmin returns the DLQ admin interface for operational tasks
func (i *IndustrialAdapter) DLQAdmin() *resilience.DLQAdmin {
	if i.dlqProducer == nil {
		return nil
	}

	dlqConfig := resilience.DLQConfig{
		Enabled:        i.cfg.Resilience.DLQ.Enabled,
		TopicSuffix:    i.cfg.Resilience.DLQ.TopicSuffix,
		KafkaBrokers:   []string{i.cfg.Kafka.Brokers},
		ConsumerGroup:  "industrial-dlq-consumer",
	}

	consumer := resilience.NewDLQConsumer(dlqConfig, nil, i.logger)
	metrics := resilience.NewMetricsCollector(i.logger)

	return resilience.NewDLQAdmin(i.dlqProducer, consumer, metrics, i.logger)
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
	adapter, err := NewIndustrialAdapter(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create industrial adapter: %w", err)
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
		fmt.Fprintf(os.Stderr, "Industrial adapter failed: %v\n", err)
		os.Exit(1)
	}
}
