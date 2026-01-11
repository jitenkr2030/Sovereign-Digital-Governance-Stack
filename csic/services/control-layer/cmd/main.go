package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"csic-platform/control-layer/internal/adapters/handlers"
	"csic-platform/control-layer/internal/adapters/messaging"
	"csic-platform/control-layer/internal/adapters/storage"
	"csic-platform/control-layer/internal/config"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/internal/core/services"
	"csic-platform/control-layer/pkg/logger"
	"csic-platform/control-layer/pkg/metrics"

	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "internal/config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	zapLogger, err := logger.NewZapLogger("info")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting Control Layer Service",
		logger.String("service", "control-layer"),
		logger.String("config_path", *configPath),
	)

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		zapLogger.Fatal("Failed to load configuration", logger.Error(err))
	}

	// Update log level from config
	zapLogger, err = logger.NewZapLogger(cfg.LogLevel)
	if err != nil {
		zapLogger.Fatal("Failed to create logger with configured level", logger.Error(err))
	}
	defer zapLogger.Sync()

	zapLogger.Info("Configuration loaded successfully",
		logger.String("environment", cfg.Environment),
		logger.Int("http_port", cfg.HTTPPort),
		logger.Int("grpc_port", cfg.GRPCPort),
	)

	// Initialize metrics
	metricsCollector := metrics.NewPrometheusCollector("control_layer")

	// Initialize PostgreSQL repositories
	policyRepo, err := storage.NewPostgresPolicyRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL for policies", logger.Error(err))
	}
	defer policyRepo.Close()

	enforcementRepo, err := storage.NewPostgresEnforcementRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL for enforcement", logger.Error(err))
	}
	defer enforcementRepo.Close()

	interventionRepo, err := storage.NewPostgresInterventionRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL for interventions", logger.Error(err))
	}
	defer interventionRepo.Close()

	// Initialize Redis client
	redisClient, err := storage.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Redis", logger.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.KafkaBrokers)
	if err != nil {
		zapLogger.Fatal("Failed to create Kafka producer", logger.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize Kafka consumer for policy updates
	kafkaConsumer, err := messaging.NewKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaConsumerGroup)
	if err != nil {
		zapLogger.Fatal("Failed to create Kafka consumer", logger.Error(err))
	}
	defer kafkaConsumer.Close()

	// Initialize repositories ports
	repositories := ports.Repositories{
		PolicyRepository:       policyRepo,
		EnforcementRepository:  enforcementRepo,
		InterventionRepository: interventionRepo,
	}

	// Initialize cache port
	cachePort := ports.CachePort{
		Client: redisClient,
	}

	// Initialize messaging port
	messagingPort := ports.MessagingPort{
		Producer: kafkaProducer,
		Consumer: kafkaConsumer,
	}

	// Initialize services
	policyEngine := services.NewPolicyEngine(repositories, cachePort, messagingPort, zapLogger, metricsCollector)
	enforcementHandler := services.NewEnforcementHandler(repositories, messagingPort, zapLogger, metricsCollector)
	stateRegistry := services.NewStateRegistry(repositories, cachePort, zapLogger, metricsCollector)
	interventionService := services.NewInterventionService(repositories, messagingPort, zapLogger, metricsCollector, policyEngine)

	// Initialize HTTP handler
	httpHandler := handlers.NewHTTPHandler(
		policyEngine,
		enforcementHandler,
		stateRegistry,
		interventionService,
		metricsCollector,
		zapLogger,
	)

	// Initialize gRPC handler
	grpcHandler := handlers.NewGRPCHandler(
		policyEngine,
		enforcementHandler,
		stateRegistry,
		interventionService,
		metricsCollector,
		zapLogger,
	)

	// Start policy update consumer in background
	go policyEngine.StartPolicyUpdateConsumer(zapLogger)

	// Start intervention monitor in background
	go interventionService.StartInterventionMonitor(zapLogger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server
	go func() {
		if err := httpHandler.Start(cfg.HTTPPort, zapLogger); err != nil {
			zapLogger.Error("HTTP server error", logger.Error(err))
		}
	}()

	// Start gRPC server
	go func() {
		if err := grpcHandler.Start(cfg.GRPCPort, zapLogger); err != nil {
			zapLogger.Error("gRPC server error", logger.Error(err))
		}
	}()

	zapLogger.Info("Control Layer Service started successfully",
		logger.Int("http_port", cfg.HTTPPort),
		logger.Int("grpc_port", cfg.GRPCPort),
	)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	zapLogger.Info("Received shutdown signal", logger.String("signal", sig.String()))

	// Cancel context to stop background workers
	cancel()

	// Give background workers time to gracefully shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	shutdownErr := httpHandler.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		zapLogger.Error("Error shutting down HTTP server", logger.Error(shutdownErr))
	}

	// Shutdown gRPC server
	grpcHandler.Shutdown()

	zapLogger.Info("Control Layer Service shutdown complete")
}

// WaitForShutdown waits for shutdown signal
func WaitForShutdown(zapLogger *zap.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	zapLogger.Info("Received shutdown signal", logger.String("signal", sig.String()))
}

// HealthCheckHandler returns a health check handler
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
