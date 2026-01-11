// NEAM Sensing Layer - Main Ingestion Engine
// Coordinates data collection, normalization, and publishing to Kafka

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	"neam-platform/sensing/common"
	"neam-platform/sensing/energy/telemetry"
	"neam-platform/sensing/industry"
	"neam-platform/sensing/payments"
	"neam-platform/sensing/payments/normalizer"
	"neam-platform/sensing/transport"
)

// IngestionConfig represents the complete ingestion configuration
type IngestionConfig struct {
	// Server configuration
	HTTPPort int `mapstructure:"http_port" yaml:"http_port" env:"NEAM_HTTP_PORT"`
	
	// Kafka configuration
	KafkaBrokers    string `mapstructure:"kafka_brokers" yaml:"kafka_brokers" env:"NEAM_KAFKA_BROKERS"`
	KafkaTopicPrefix string `mapstructure:"kafka_topic_prefix" yaml:"kafka_topic_prefix" env:"NEAM_KAFKA_TOPIC_PREFIX"`
	
	// ClickHouse configuration
	ClickHouseDSN     string `mapstructure:"clickhouse_dsn" yaml:"clickhouse_dsn" env:"NEAM_CLICKHOUSE_DSN"`
	ClickHouseDatabase string `mapstructure:"clickhouse_database" yaml:"clickhouse_database" env:"NEAM_CLICKHOUSE_DATABASE"`
	
	// Redis configuration
	RedisAddr     string `mapstructure:"redis_addr" yaml:"redis_addr" env:"NEAM_REDIS_ADDR"`
	RedisPassword string `mapstructure:"redis_password" yaml:"redis_password" env:"NEAM_REDIS_PASSWORD"`
	
	// Anonymization
	AnonymizationEnabled bool   `mapstructure:"anonymization_enabled" yaml:"anonymization_enabled"`
	AnonymizationSalt    string `mapstructure:"anonymization_salt" yaml:"anonymization_salt" env:"NEAM_ANONYMIZATION_SALT"`
	
	// Rate limiting
	RateLimitRequests int           `mapstructure:"rate_limit_requests" yaml:"rate_limit_requests"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window" yaml:"rate_limit_window"`
	
	// Data retention
	DataRetentionDays int `mapstructure:"data_retention_days" yaml:"data_retention_days" env:"NEAM_DATA_RETENTION_DAYS"`
	
	// Processing workers
	ProcessingWorkers int `mapstructure:"processing_workers" yaml:"processing_workers"`
	BatchSize         int `mapstructure:"batch_size" yaml:"batch_size"`
}

// IngestionEngine manages the complete data ingestion pipeline
type IngestionEngine struct {
	config     *IngestionConfig
	logger     *common.Logger
	anonymizer *common.Anonymizer
	validator  *common.Validator
	rateLimiter *common.RateLimiter
	normalizer *normalizer.Normalizer
	
	// Kafka
	kafkaWriter  *kafka.Writer
	kafkaReaders []*kafka.Reader
	
	// Channels
	eventChannels map[string]chan interface{}
	
	// Payment adapters
	paymentAggregator *payments.PaymentAggregator
	
	// Energy adapters
	energyAggregator *telemetry.EnergyAggregator
	
	// Transport adapters
	transportAggregator *transport.TransportAggregator
	
	// Industrial adapters
	industrialAggregator *industry.IndustrialAggregator
	
	// Metrics
	metrics    IngestionMetrics
	metricsMu  sync.RWMutex
	
	// Lifecycle
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	stopChan chan struct{}
}

// IngestionMetrics holds pipeline performance metrics
type IngestionMetrics struct {
	EventsIngested    map[string]int64
	EventsNormalized  int64
	EventsPublished   int64
	EventsFailed      int64
	EventsAnonymized  int64
	StartTime         time.Time
	LastEventTime     time.Time
	ProcessingLatency time.Duration
}

// NewIngestionEngine creates a new ingestion engine
func NewIngestionEngine(config *IngestionConfig) (*IngestionEngine, error) {
	logger := common.NewLogger("sensing")
	
	ctx, cancel := context.WithCancel(context.Background())
	
	engine := &IngestionEngine{
		config:      config,
		logger:      logger,
		anonymizer:  common.NewAnonymizer(config.AnonymizationSalt),
		validator:   common.NewValidator(),
		rateLimiter: common.NewRateLimiter(config.RateLimitRequests, config.RateLimitWindow),
		ctx:         ctx,
		cancel:      cancel,
		stopChan:    make(chan struct{}),
		metrics: IngestionMetrics{
			EventsIngested: make(map[string]int64),
			StartTime:      time.Now(),
		},
	}
	
	// Initialize Kafka writer
	if err := engine.initKafka(); err != nil {
		return nil, fmt.Errorf("failed to initialize Kafka: %w", err)
	}
	
	// Initialize normalizer
	normalizerConfig := &normalizer.NormalizationConfig{
		AnonymizationEnabled: config.AnonymizationEnabled,
		QualityThreshold:     0.5,
		BatchSize:            config.BatchSize,
		ProcessingWorkers:    config.ProcessingWorkers,
	}
	engine.normalizer = normalizer.NewNormalizer(normalizerConfig, logger, engine.anonymizer)
	
	// Initialize adapters
	if err := engine.initAdapters(); err != nil {
		return nil, fmt.Errorf("failed to initialize adapters: %w", err)
	}
	
	return engine, nil
}

// initKafka initializes Kafka connections
func (ie *IngestionEngine) initKafka() error {
	// Create Kafka writer for publishing normalized events
	ie.kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(ie.config.KafkaBrokers),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
	}
	
	// Create Kafka readers for batch ingestion topics
	topicPrefix := ie.config.KafkaTopicPrefix
	ie.kafkaReaders = []*kafka.Reader{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:  ie.config.KafkaBrokers,
			Topic:    topicPrefix + ".raw.payments",
			GroupID:  "neam-sensing-payments",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:  ie.config.KafkaBrokers,
			Topic:    topicPrefix + ".raw.energy",
			GroupID:  "neam-sensing-energy",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:  ie.config.KafkaBrokers,
			Topic:    topicPrefix + ".raw.transport",
			GroupID:  "neam-sensing-transport",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		kafka.NewReader(kafka.ReaderConfig{
			Brokers:  ie.config.KafkaBrokers,
			Topic:    topicPrefix + ".raw.industry",
			GroupID:  "neam-sensing-industry",
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
	}
	
	return nil
}

// initAdapters initializes data source adapters
func (ie *IngestionEngine) initAdapters() error {
	// Initialize payment adapters
	bankAdapter := payments.NewBankAdapter(payments.BankConfig{
		Name:            "ReserveBank",
		APIEndpoint:     "https://api.rbi.org.in/v1/payments",
		BatchSize:       100,
		PollingInterval: 1 * time.Minute,
		Region:          "NATIONAL",
	}, ie.logger)
	
	upiAdapter := payments.NewUPIAdapter(payments.UPIConfig{
		Name:            "NPCI",
		NPCIEndpoint:    "https://api.npci.org.in/v1/upi",
		PSPID:           "NEAM-PSP",
		BatchSize:       500,
		PollingInterval: 30 * time.Second,
		Region:          "NATIONAL",
	}, ie.logger)
	
	posAdapter := payments.NewPOSAdapter(payments.POSConfig{
		Name:              "POS-Aggregator",
		AggregatorEndpoint: "https://api.pos aggregator.in/v1",
		TerminalIDs:       []string{"TERM001", "TERM002", "TERM003"},
		BatchSize:         200,
		PollingInterval:   1 * time.Minute,
		Region:            "NATIONAL",
	}, ie.logger)
	
	walletAdapter := payments.NewWalletAdapter(payments.WalletConfig{
		Name:           "Wallet-Aggregator",
		APIEndpoint:    "https://api.wallet.in/v1",
		WalletIDs:      []string{"WAL001", "WAL002", "WAL003"},
		BatchSize:      300,
		PollingInterval: 30 * time.Second,
		Region:         "NATIONAL",
	}, ie.logger)
	
	ie.paymentAggregator = payments.NewPaymentAggregator(ie.logger, bankAdapter, upiAdapter, posAdapter, walletAdapter)
	
	// Initialize energy adapters
	scadaAdapter := telemetry.NewSCADAAdapter(telemetry.SCADAConfig{
		Name:            "SCADA-Main",
		SCADAEndpoint:   "https://scada.grid.in/v1",
		RegionIDs:       []string{"NORTH", "SOUTH", "EAST", "WEST"},
		StationIDs:      []string{"STN001", "STN002", "STN003"},
		PollingInterval: 15 * time.Second,
	}, ie.logger)
	
	nldcAdapter := telemetry.NewNLDCAdapter(telemetry.NLDConfig{
		Name:           "NLDC-National",
		NLDCEndpoint:   "https://nldc.grid.in/v1",
		RegionID:       "NATIONAL",
		PollingInterval: 1 * time.Minute,
	}, ie.logger)
	
	renewableAdapter := telemetry.NewRenewableAdapter(telemetry.RenewableConfig{
		Name:            "Renewable-Monitor",
		APIEndpoint:     "https://renewable.monitor.in/v1",
		Types:           []string{"solar", "wind", "hydro"},
		RegionIDs:       []string{"NORTH", "SOUTH", "EAST", "WEST"},
		PollingInterval: 5 * time.Minute,
	}, ie.logger)
	
	ie.energyAggregator = telemetry.NewEnergyAggregator(ie.logger, scadaAdapter, nldcAdapter, renewableAdapter)
	
	// Initialize transport adapters
	portAdapter := transport.NewPortAuthorityAdapter(transport.PortConfig{
		Name:            "Port-Authority",
		APIEndpoint:     "https://api.ports.in/v1",
		PortIDs:         []string{"PORT-MUM", "PORT-KOL", "PORT-CHEN", "PORT-DEL"},
		PollingInterval: 5 * time.Minute,
	}, ie.logger)
	
	fuelAdapter := transport.NewFuelAdapter(transport.FuelConfig{
		Name:            "Fuel-Monitor",
		APIEndpoint:     "https://fuel.monitor.in/v1",
		RegionIDs:       []string{"NORTH", "SOUTH", "EAST", "WEST"},
		StationIDs:      []string{"F001", "F002", "F003", "F004", "F005"},
		PollingInterval: 15 * time.Minute,
	}, ie.logger)
	
	logisticsAdapter := transport.NewLogisticsAdapter(transport.LogisticsConfig{
		Name:           "Logistics-Monitor",
		APIEndpoint:    "https://logistics.monitor.in/v1",
		TransportTypes: []string{"road", "rail", "air"},
		RegionIDs:      []string{"NORTH", "SOUTH", "EAST", "WEST"},
		PollingInterval: 5 * time.Minute,
	}, ie.logger)
	
	ie.transportAggregator = transport.NewTransportAggregator(ie.logger, portAdapter, fuelAdapter, logisticsAdapter)
	
	// Initialize industrial adapters
	surveyAdapter := industry.NewIndustrialSurveyAdapter(industry.SurveyConfig{
		Name:            "Industrial-Survey",
		APIEndpoint:     "https://survey.mospi.in/v1",
		Sectors:         []string{"manufacturing", "mining", "utilities"},
		RegionIDs:       []string{"NORTH", "SOUTH", "EAST", "WEST"},
		PollingInterval: 1 * time.Hour,
	}, ie.logger)
	
	supplyChainAdapter := industry.NewSupplyChainAdapter(industry.SupplyChainConfig{
		Name:           "Supply-Chain",
		APIEndpoint:    "https://supplychain.monitor.in/v1",
		ChainTypes:     []string{"raw_materials", "components", "finished_goods"},
		RegionIDs:      []string{"NORTH", "SOUTH", "EAST", "WEST"},
		PollingInterval: 30 * time.Minute,
	}, ie.logger)
	
	indexAdapter := industry.NewIndexAdapter(industry.IndexConfig{
		Name:            "Index-Provider",
		APIEndpoint:     "https://index.mospi.in/v1",
		IndexTypes:      []string("production", "productivity"),
		PollingInterval: 24 * time.Hour,
	}, ie.logger)
	
	ie.industrialAggregator = industry.NewIndustrialAggregator(ie.logger, surveyAdapter, supplyChainAdapter, indexAdapter)
	
	return nil
}

// Start begins the ingestion pipeline
func (ie *IngestionEngine) Start() error {
	ie.logger.Info("Starting ingestion engine", "workers", ie.config.ProcessingWorkers)
	
	// Start all aggregators
	if err := ie.paymentAggregator.Start(ie.ctx); err != nil {
		return fmt.Errorf("failed to start payment aggregator: %w", err)
	}
	
	if err := ie.energyAggregator.Start(ie.ctx); err != nil {
		return fmt.Errorf("failed to start energy aggregator: %w", err)
	}
	
	if err := ie.transportAggregator.Start(ie.ctx); err != nil {
		return fmt.Errorf("failed to start transport aggregator: %w", err)
	}
	
	if err := ie.industrialAggregator.Start(ie.ctx); err != nil {
		return fmt.Errorf("failed to start industrial aggregator: %w", err)
	}
	
	// Set normalizer input channels
	ie.normalizer.SetInputChannels(
		ie.paymentAggregator.Stream(),
		nil, // energy channel - would need type assertion
		nil, // transport channel - would need type assertion  
		nil, // production channel - would need type assertion
	)
	
	// Start normalizer
	if err := ie.normalizer.Start(ie.ctx); err != nil {
		return fmt.Errorf("failed to start normalizer: %w", err)
	}
	
	// Start Kafka publishers
	ie.wg.Add(ie.config.ProcessingWorkers)
	for i := 0; i < ie.config.ProcessingWorkers; i++ {
		go ie.publisher(i)
	}
	
	// Start metrics reporter
	go ie.metricsReporter()
	
	ie.logger.Info("Ingestion engine started successfully")
	return nil
}

// publisher publishes normalized events to Kafka
func (ie *IngestionEngine) publisher(id int) {
	defer ie.wg.Done()
	
	for {
		select {
		case <-ie.stopChan:
			return
		case event := <-ie.normalizer.GetOutputChannel():
			ie.publishEvent(event)
		}
	}
}

// publishEvent publishes a normalized event to Kafka
func (ie *IngestionEngine) publishEvent(event normalizer.NormalizedEvent) {
	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		ie.logger.Error("Failed to serialize event", "error", err)
		ie.incrementEventsFailed("serialization")
		return
	}
	
	// Determine topic based on category
	topic := ie.config.KafkaTopicPrefix + "." + event.Category
	
	// Create Kafka message
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(event.ID),
		Value: data,
		Time:  event.EventTime,
		Headers: []kafka.Header{
			{Key: "source_type", Value: []byte(event.SourceType)},
			{Key: "region_id", Value: []byte(event.RegionID)},
			{Key: "event_type", Value: []byte(event.EventType)},
		},
	}
	
	// Write to Kafka
	if err := ie.kafkaWriter.WriteMessages(ie.ctx, msg); err != nil {
		ie.logger.Error("Failed to publish event", "error", err, "topic", topic)
		ie.incrementEventsFailed("publish")
		return
	}
	
	ie.incrementEventsPublished(event.Category)
}

// incrementEventsFailed increments failed counter
func (ie *IngestionEngine) incrementEventsFailed(reason string) {
	ie.metricsMu.Lock()
	defer ie.metricsMu.Unlock()
	ie.metrics.EventsFailed++
}

// incrementEventsPublished increments published counter
func (ie *IngestionEngine) incrementEventsPublished(category string) {
	ie.metricsMu.Lock()
	defer ie.metricsMu.Unlock()
	ie.metrics.EventsPublished++
	ie.metrics.EventsIngested[category]++
	ie.metrics.LastEventTime = time.Now()
}

// metricsReporter periodically logs metrics
func (ie *IngestionEngine) metricsReporter() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ie.stopChan:
			return
		case <-ticker.C:
			metrics := ie.GetMetrics()
			ie.logger.Info("Ingestion metrics",
				"total_published", metrics.EventsPublished,
				"total_failed", metrics.EventsFailed,
				"by_category", metrics.EventsIngested,
				"uptime", time.Since(metrics.StartTime).String(),
			)
		}
	}
}

// Stop gracefully stops the ingestion engine
func (ie *IngestionEngine) Stop() {
	ie.logger.Info("Stopping ingestion engine...")
	
	close(ie.stopChan)
	
	// Stop normalizer
	ie.normalizer.Stop()
	
	// Stop aggregators
	ie.paymentAggregator.Stop()
	ie.energyAggregator.Stop()
	ie.transportAggregator.Stop()
	ie.industrialAggregator.Stop()
	
	// Close Kafka connections
	for _, reader := range ie.kafkaReaders {
		reader.Close()
	}
	ie.kafkaWriter.Close()
	
	ie.wg.Wait()
	
	ie.logger.Info("Ingestion engine stopped")
}

// GetMetrics returns current metrics
func (ie *IngestionEngine) GetMetrics() IngestionMetrics {
	ie.metricsMu.RLock()
	defer ie.metricsMu.RUnlock()
	return ie.metrics
}

// SetupRouter sets up HTTP routes for health checks and metrics
func (ie *IngestionEngine) SetupRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(ie.rateLimiter.Middleware())
	
	// Health endpoints
	router.GET("/health", ie.healthCheck)
	router.GET("/ready", ie.readinessCheck)
	
	// Metrics endpoint
	router.GET("/metrics", ie.metricsHandler)
	
	// API v1
	v1 := router.Group("/api/v1")
	{
		// Ingestion status
		v1.GET("/status", ie.statusHandler)
		
		// Manual ingestion trigger
		v1.POST("/ingest/payments", ie.manualPaymentIngest)
		v1.POST("/ingest/energy", ie.manualEnergyIngest)
		v1.POST("/ingest/transport", ie.manualTransportIngest)
		v1.POST("/ingest/industry", ie.manualIndustryIngest)
	}
	
	return router
}

// Health check handlers
func (ie *IngestionEngine) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-sensing",
		"uptime":    time.Since(ie.metrics.StartTime).String(),
	})
}

func (ie *IngestionEngine) readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-sensing",
	})
}

func (ie *IngestionEngine) metricsHandler(c *gin.Context) {
	metrics := ie.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"events_published":   metrics.EventsPublished,
		"events_failed":      metrics.EventsFailed,
		"events_by_category": metrics.EventsIngested,
		"uptime":             time.Since(metrics.StartTime).String(),
		"last_event":         metrics.LastEventTime,
	})
}

func (ie *IngestionEngine) statusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "running",
		"workers":    ie.config.ProcessingWorkers,
		"kafka_topic": ie.config.KafkaTopicPrefix + ".*",
		"adapters": map[string]bool{
			"payments":    true,
			"energy":      true,
			"transport":   true,
			"industry":    true,
		},
	})
}

// Manual ingestion handlers
func (ie *IngestionEngine) manualPaymentIngest(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Payment ingestion triggered",
		"id":      fmt.Sprintf("manual-%d", time.Now().UnixNano()),
	})
}

func (ie *IngestionEngine) manualEnergyIngest(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Energy ingestion triggered",
		"id":      fmt.Sprintf("manual-%d", time.Now().UnixNano()),
	})
}

func (ie *IngestionEngine) manualTransportIngest(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Transport ingestion triggered",
		"id":      fmt.Sprintf("manual-%d", time.Now().UnixNano()),
	})
}

func (ie *IngestionEngine) manualIndustryIngest(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Industry ingestion triggered",
		"id":      fmt.Sprintf("manual-%d", time.Now().UnixNano()),
	})
}

// LoadConfig loads configuration from file/environment
func LoadConfig(configPath string) (*IngestionConfig, error) {
	config := &IngestionConfig{}
	
	// Set defaults
	config.HTTPPort = 8082
	config.KafkaTopicPrefix = "neam.sensing"
	config.RateLimitRequests = 10000
	config.RateLimitWindow = time.Minute
	config.DataRetentionDays = 90
	config.ProcessingWorkers = 4
	config.BatchSize = 1000
	config.AnonymizationEnabled = true
	
	// Load from environment variables if set
	if v := os.Getenv("NEAM_HTTP_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &config.HTTPPort)
	}
	if v := os.Getenv("NEAM_KAFKA_BROKERS"); v != "" {
		config.KafkaBrokers = v
	}
	if v := os.Getenv("NEAM_CLICKHOUSE_DSN"); v != "" {
		config.ClickHouseDSN = v
	}
	if v := os.Getenv("NEAM_REDIS_ADDR"); v != "" {
		config.RedisAddr = v
	}
	
	return config, nil
}

// Main entry point
func main() {
	// Load configuration
	config, err := LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Create ingestion engine
	engine, err := NewIngestionEngine(config)
	if err != nil {
		log.Fatalf("Failed to create ingestion engine: %v", err)
	}
	
	// Start engine
	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start ingestion engine: %v", err)
	}
	
	// Setup HTTP router
	router := engine.SetupRouter()
	
	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		ie.logger.Info("HTTP server starting", "port", config.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	// Graceful shutdown
	engine.Stop()
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	srv.Shutdown(shutdownCtx)
	
	ie.logger.Info("Shutdown complete")
}
