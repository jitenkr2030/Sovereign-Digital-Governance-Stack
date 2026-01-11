package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/adapter/connector"
	"github.com/csic-platform/services/exchange-ingestion/internal/adapter/publisher"
	"github.com/csic-platform/services/exchange-ingestion/internal/adapter/repository"
	"github.com/csic-platform/services/exchange-ingestion/internal/config"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/service"
	"github.com/csic-platform/services/exchange-ingestion/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Exchange Ingestion Service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Info("Configuration loaded",
		zap.String("env", cfg.Env),
		zap.String("db_host", cfg.DBHost),
		zap.String("kafka_brokers", cfg.KafkaBrokers))

	// Connect to database
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// Initialize repositories
	dataSourceRepo := repository.NewPostgresDataSourceRepository(db)
	marketDataRepo := repository.NewPostgresMarketDataRepository(db)
	statsRepo := repository.NewPostgresIngestionStatsRepository(db)

	// Initialize Kafka publisher
	kafkaPublisher := publisher.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaTopicPrefix, logger)
	defer kafkaPublisher.Close()
	logger.Info("Kafka publisher initialized")

	// Initialize connector factory
	connectorFactory := connector.NewConnectorFactory(logger)

	// Initialize services
	dataSourceService := service.NewDataSourceService(dataSourceRepo, connectorFactory, logger)
	ingestionService := service.NewIngestionService(
		dataSourceRepo,
		marketDataRepo,
		statsRepo,
		kafkaPublisher,
		connectorFactory,
		logger,
	)

	// Initialize HTTP handler
	httpHandler := handler.NewHTTPHandler(dataSourceService, ingestionService, logger)

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(30 * time.Second))

	httpHandler.RegisterRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting",
			zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop ingestion service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ingestionService.StopIngestion(ctx); err != nil {
		logger.Error("Error stopping ingestion", zap.Error(err))
	}

	// Graceful shutdown of HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// initLogger initializes the Zap logger
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}
