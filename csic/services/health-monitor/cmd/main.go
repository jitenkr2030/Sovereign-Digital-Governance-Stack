package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"csic-platform/health-monitor/internal/adapters/handlers"
	"csic-platform/health-monitor/internal/adapters/messaging"
	"csic-platform/health-monitor/internal/adapters/storage"
	"csic-platform/health-monitor/internal/config"
	"csic-platform/health-monitor/internal/core/ports"
	"csic-platform/health-monitor/internal/core/services"
	"csic-platform/health-monitor/pkg/logger"
	"csic-platform/health-monitor/pkg/metrics"

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

	zapLogger.Info("Starting Health Monitor Service",
		logger.String("service", "health-monitor"),
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
	)

	// Initialize metrics
	metricsCollector := metrics.NewPrometheusCollector("health_monitor")

	// Initialize PostgreSQL repositories
	serviceStatusRepo, err := storage.NewPostgresServiceStatusRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL for service status", logger.Error(err))
	}
	defer serviceStatusRepo.Close()

	alertRuleRepo, err := storage.NewPostgresAlertRuleRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL for alert rules", logger.Error(err))
	}
	defer alertRuleRepo.Close()

	outageRepo, err := storage.NewPostgresOutageRepository(cfg.DatabaseURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to PostgreSQL for outages", logger.Error(err))
	}
	defer outageRepo.Close()

	// Initialize Redis client
	redisClient, err := storage.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Redis", logger.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka consumer for heartbeats
	kafkaConsumer, err := messaging.NewKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaConsumerGroup)
	if err != nil {
		zapLogger.Fatal("Failed to create Kafka consumer", logger.Error(err))
	}
	defer kafkaConsumer.Close()

	// Initialize Kafka producer for alerts
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.KafkaBrokers)
	if err != nil {
		zapLogger.Fatal("Failed to create Kafka producer", logger.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize repositories ports
	repositories := ports.Repositories{
		ServiceStatusRepository: serviceStatusRepo,
		AlertRuleRepository:     alertRuleRepo,
		OutageRepository:        outageRepo,
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
	monitorService := services.NewMonitorService(repositories, cachePort, messagingPort, zapLogger, metricsCollector)
	alertService := services.NewAlertService(repositories, cachePort, messagingPort, zapLogger, metricsCollector)

	// Initialize HTTP handler
	httpHandler := handlers.NewHTTPHandler(
		monitorService,
		alertService,
		metricsCollector,
		zapLogger,
	)

	// Start heartbeat consumer
	go monitorService.StartHeartbeatConsumer(zapLogger)

	// Start health check monitor
	go monitorService.StartHealthCheckMonitor(zapLogger)

	// Start alert evaluator
	go alertService.StartAlertEvaluator(zapLogger)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server
	go func() {
		if err := httpHandler.Start(cfg.HTTPPort, zapLogger); err != nil {
			zapLogger.Error("HTTP server error", logger.Error(err))
		}
	}()

	zapLogger.Info("Health Monitor Service started successfully",
		logger.Int("http_port", cfg.HTTPPort),
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

	zapLogger.Info("Health Monitor Service shutdown complete")
}
