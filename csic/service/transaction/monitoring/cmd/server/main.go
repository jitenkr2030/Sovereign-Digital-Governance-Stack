package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	httpHandler "github.com/csic/transaction-monitoring/internal/handler/http"
	kafkaConsumer "github.com/csic/transaction-monitoring/internal/handler/kafka"
	"github.com/csic/transaction-monitoring/internal/repository"
	graphSvc "github.com/csic/transaction-monitoring/internal/service/graph"
	ingestSvc "github.com/csic/transaction-monitoring/internal/service/ingest"
	riskSvc "github.com/csic/transaction-monitoring/internal/service/risk"
	sanctionsSvc "github.com/csic/transaction-monitoring/internal/service/sanctions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Transaction Monitoring Service",
		zap.String("version", "1.0.0"))

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize repository
	repo, err := repository.NewRepository(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize repository", zap.Error(err))
	}
	defer repo.Close()

	// Initialize cache repository
	cacheRepo, err := repository.NewCacheRepository(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize cache repository", zap.Error(err))
	}
	defer cacheRepo.Close()

	// Initialize services
	ingestionService, err := ingestSvc.NewIngestionService(cfg, repo, logger)
	if err != nil {
		logger.Fatal("Failed to initialize ingestion service", zap.Error(err))
	}

	riskService := riskSvc.NewRiskScoringService(cfg, repo, cacheRepo, logger)

	clusteringService := graphSvc.NewClusteringService(cfg, repo, nil, logger)

	sanctionsService := sanctionsSvc.NewSanctionsService(cfg, repo, cacheRepo, logger)

	// Start services
	if err := sanctionsService.Start(ctx); err != nil {
		logger.Fatal("Failed to start sanctions service", zap.Error(err))
	}
	defer sanctionsService.Stop()

	if err := ingestionService.Start(ctx); err != nil {
		logger.Fatal("Failed to start ingestion service", zap.Error(err))
	}
	defer ingestionService.Stop()

	if err := clusteringService.Start(ctx); err != nil {
		logger.Fatal("Failed to start clustering service", zap.Error(err))
	}
	defer clusteringService.Stop()

	// Initialize Kafka consumer
	consumer := kafkaConsumer.NewConsumer(
		cfg, repo, cacheRepo, riskService, clusteringService, sanctionsService, logger)
	if err := consumer.Start(ctx); err != nil {
		logger.Fatal("Failed to start Kafka consumer", zap.Error(err))
	}
	defer consumer.Stop()

	// Initialize HTTP handler
	handler := httpHandler.NewHandler(
		cfg, repo, cacheRepo, ingestionService, riskService, clusteringService, sanctionsService, logger)

	// Setup router
	router := handler.SetupRouter()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting",
			zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Cancel main context to stop services
	cancel()

	logger.Info("Server stopped")
}

func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	return logger
}
