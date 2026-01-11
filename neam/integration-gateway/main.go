package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"neam-platform/integration-gateway/adapters"
	"neam-platform/integration-gateway/api"
	"neam-platform/integration-gateway/config"
	"neam-platform/integration-gateway/kafka"
	"neam-platform/integration-gateway/middleware"
	"neam-platform/integration-gateway/models"
	"neam-platform/integration-gateway/monitoring"
	"neam-platform/integration-gateway/repository"
	"neam-platform/integration-gateway/security"
	"neam-platform/integration-gateway/transformers"
)

// @title NEAM Integration & Interoperability Gateway
// @description Secure data exchange layer for government system integration
// @version 1.0.0
// @host 0.0.0.0:8080
// @BasePath /api/v1

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	logger := initLogger(cfg)
	logger.Info("Starting NEAM Integration Gateway", map[string]interface{}{
		"environment": cfg.Environment,
		"version":     "1.0.0",
	})

	// Initialize TLS configuration
	tlsConfig, err := security.LoadTLSConfig(cfg)
	if err != nil {
		logger.Fatal("Failed to load TLS configuration", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(ctx, cfg.KafkaBrokers, cfg.KafkaTLSEnabled)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka producer", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer kafkaProducer.Close()

	// Initialize Kafka consumer group
	kafkaConsumer, err := kafka.NewConsumerGroup(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup, cfg.KafkaTLSEnabled)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka consumer", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize data transformer
	transformer := transformers.NewDataTransformer()

	// Initialize adapters
	adapterRegistry := adapters.NewAdapterRegistry()
	adapterRegistry.Register("ministry", adapters.NewMinistryAdapter(cfg))
	adapterRegistry.Register("central_bank", adapters.NewCentralBankAdapter(cfg))
	adapterRegistry.Register("state", adapters.NewStateAdapter(cfg))
	adapterRegistry.Register("district", adapters.NewDistrictAdapter(cfg))
	adapterRegistry.Register("legacy", adapters.NewLegacyAdapter(cfg, transformer))

	// Initialize repository
	repo := repository.NewIntegrationRepository(cfg)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimitRequests,
		cfg.RateLimitWindow,
	)

	// Initialize health checker
	healthChecker := monitoring.NewHealthChecker(kafkaProducer, repo)

	// Initialize metrics server
	metricsServer := monitoring.NewMetricsServer(cfg.MetricsPort)

	// Start metrics server
	go func() {
		if err := metricsServer.Start(); err != nil {
			logger.Error("Metrics server failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Initialize REST handler
	restHandler := api.NewRESTHandler(
		adapterRegistry,
		kafkaProducer,
		transformer,
		repo,
		rateLimiter,
		logger,
	)

	// Initialize gRPC server
	grpcHandler := api.NewGRPCHandler(
		adapterRegistry,
		kafkaProducer,
		transformer,
		repo,
		logger,
	)

	// Initialize federation handler
	federationHandler := api.NewFederationHandler(
		adapterRegistry,
		kafkaProducer,
		transformer,
		logger,
	)

	// Set up HTTP server with TLS
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      restHandler.SetupRouter(tlsConfig),
		TLSConfig:    tlsConfig,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Set up gRPC server
	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	// Register gRPC services
	api.RegisterIntegrationServiceServer(grpcServer, grpcHandler)
	api.RegisterFederationServiceServer(grpcServer, federationHandler)

	// Start Kafka consumer
	go func() {
		topics := []string{
			cfg.KafkaIngestionTopic,
			cfg.KafkaBroadcastTopic,
		}
		if err := kafkaConsumer.Start(ctx, topics, &kafkaMessageHandler{
			adapterRegistry: adapterRegistry,
			transformer:     transformer,
			logger:          logger,
		}); err != nil {
			logger.Error("Kafka consumer error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Starting HTTPS server", map[string]interface{}{
			"port": cfg.HTTPPort,
		})
		if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Start gRPC server in goroutine
	go func() {
		logger.Info("Starting gRPC server", map[string]interface{}{
			"port": cfg.GRPCPort,
		})
		listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort), tlsConfig)
		if err != nil {
			logger.Fatal("Failed to create gRPC listener", map[string]interface{}{
				"error": err.Error(),
			})
		}
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC server error", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down gracefully...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop servers
	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Stop Kafka consumer
	kafkaConsumer.Close()

	cancel()
	logger.Info("Integration Gateway shutdown complete", nil)
}

// kafkaMessageHandler handles incoming Kafka messages
type kafkaMessageHandler struct {
	adapterRegistry *adapters.AdapterRegistry
	transformer     *transformers.DataTransformer
	logger          *Logger
}

// HandleMessage processes a Kafka message
func (h *kafkaMessageHandler) HandleMessage(ctx context.Context, topic string, message []byte) error {
	var payload models.IntegrationMessage
	if err := json.Unmarshal(message, &payload); err != nil {
		h.logger.Error("Failed to unmarshal Kafka message", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Route to appropriate adapter based on topic
	switch topic {
	case "neam.broadcast.v1":
		return h.handleBroadcastMessage(ctx, &payload)
	default:
		return h.handleIngestionMessage(ctx, &payload)
	}
}

// handleBroadcastMessage handles broadcast messages from NEAM core
func (h *kafkaMessageHandler) handleBroadcastMessage(ctx context.Context, msg *models.IntegrationMessage) error {
	h.logger.Info("Processing broadcast message", map[string]interface{}{
		"message_id": msg.MessageID,
		"source":     msg.Source,
	})

	// Push to configured adapters
	adapterNames := h.adapterRegistry.ListAdapters()
	for _, name := range adapterNames {
		adapter, err := h.adapterRegistry.GetAdapter(name)
		if err != nil {
			h.logger.Error("Failed to get adapter", map[string]interface{}{
				"adapter": name,
				"error":   err.Error(),
			})
			continue
		}

		if err := adapter.Push(ctx, msg); err != nil {
			h.logger.Error("Failed to push to adapter", map[string]interface{}{
				"adapter": name,
				"error":   err.Error(),
			})
		}
	}

	return nil
}

// handleIngestionMessage handles ingestion messages
func (h *kafkaMessageHandler) handleIngestionMessage(ctx context.Context, msg *models.IntegrationMessage) error {
	h.logger.Info("Processing ingestion message", map[string]interface{}{
		"message_id": msg.MessageID,
		"data_type":  msg.DataType,
	})

	// Process the message (would be handled by specific processors)
	return nil
}

// Logger provides structured logging
type Logger struct {
	service string
}

// initLogger initializes the logger
func initLogger(cfg *config.Config) *Logger {
	return &Logger{
		service: "integration-gateway",
	}
}

// Info logs info level message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	log.Printf("[INFO] %s: %v", message, fields)
}

// Error logs error level message
func (l *Logger) Error(message string, fields map[string]interface{}) {
	log.Printf("[ERROR] %s: %v", message, fields)
}

// Fatal logs fatal message and exits
func (l *Logger) Fatal(message string, fields map[string]interface{}) {
	log.Fatalf("[FATAL] %s: %v", message, fields)
}

// loadClientCA loads client CA certificate for mTLS
func loadClientCA(caPath string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return certPool, nil
}
