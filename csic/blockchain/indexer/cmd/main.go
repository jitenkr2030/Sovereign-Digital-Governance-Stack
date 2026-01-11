package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/platform/blockchain/indexer/internal/config"
	"github.com/csic/platform/blockchain/indexer/internal/indexer"
	"github.com/csic/platform/blockchain/indexer/internal/repository"
	"github.com/csic/platform/blockchain/indexer/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Update logger level based on config
	if cfg.App.Debug {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	logger.Info("Starting Blockchain Indexer Service",
		zap.String("name", cfg.App.Name),
		zap.String("environment", cfg.App.Environment),
		zap.String("network", cfg.Blockchain.Network))

	// Initialize database connection
	dbConfig := repository.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	repo, err := repository.NewPostgresRepository(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer repo.Close()

	// Initialize Redis connection
	redisConfig := repository.RedisConfig{
		Host:      cfg.Redis.Host,
		Port:      cfg.Redis.Port,
		Password:  cfg.Redis.Password,
		DB:        cfg.Redis.DB,
		KeyPrefix: cfg.Redis.KeyPrefix,
		PoolSize:  cfg.Redis.PoolSize,
	}

	redisClient, err := repository.NewRedisClient(redisConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize blockchain client
	chainConfig := indexer.BlockchainConfig{
		Network:           cfg.Blockchain.Network,
		NodeURL:           cfg.Blockchain.NodeURL,
		WSURL:             cfg.Blockchain.WSURL,
		ChainID:           cfg.Blockchain.ChainID,
		StartBlock:        cfg.Blockchain.StartBlock,
		ConfirmationBlocks: cfg.Blockchain.ConfirmationBlocks,
	}

	blockchainClient, err := indexer.NewBlockchainClient(chainConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize blockchain client", zap.Error(err))
	}
	defer blockchainClient.Close()

	// Initialize indexer service
	indexerService := indexer.NewIndexerService(
		repo,
		redisClient,
		blockchainClient,
		cfg,
		logger,
	)

	// Initialize query service
	queryService := service.NewQueryService(repo, redisClient, cfg, logger)

	// Initialize HTTP handlers
	httpHandler := service.NewHTTPHandler(queryService, logger)

	// Set up Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(service.CORSMiddleware())

	// Register routes
	httpHandler.RegisterRoutes(router)

	// Start indexer workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start block syncing
	go func() {
		batchSize := cfg.Blockchain.BatchSize
		if err := indexerService.SyncBlocks(ctx, batchSize); err != nil {
			logger.Error("Block syncing error", zap.Error(err))
		}
	}()

	// Start event processing
	go func() {
		if err := indexerService.ProcessEvents(ctx); err != nil {
			logger.Error("Event processing error", zap.Error(err))
		}
	}()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Blockchain Indexer Service listening",
			zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Blockchain Indexer Service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Blockchain Indexer Service stopped")
}
