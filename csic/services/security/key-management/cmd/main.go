package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic-platform/services/security/key-management/internal/adapter/crypto"
	"github.com/csic-platform/services/security/key-management/internal/adapter/messaging"
	"github.com/csic-platform/services/security/key-management/internal/adapter/repository"
	"github.com/csic-platform/services/security/key-management/internal/config"
	"github.com/csic-platform/services/security/key-management/internal/core/ports"
	"github.com/csic-platform/services/security/key-management/internal/core/service"
	"github.com/csic-platform/services/security/key-management/internal/handler"
	"github.com/csic-platform/shared/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize configuration loader
	configLoader := config.NewConfigLoader(*configPath)
	cfg, err := configLoader.Load()
	if err != nil {
		cfg = getDefaultConfig()
		fmt.Printf("Warning: Failed to load config file, using defaults: %v\n", err)
	}

	// Initialize logger
	appLogger, err := logger.NewLogger(logger.Config{
		ServiceName: "key-management",
		LogLevel:    cfg.Logging.Level,
		OutputPath:  cfg.Logging.Output,
		Development: cfg.App.Environment == "development",
		JSONOutput:  cfg.Logging.Format == "json",
	})
	if err != nil {
		fmt.Printf("Fatal: Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Close()

	appLogger.Info("starting Key Management Service",
		logger.WithFields(
			logger.String("http_port", fmt.Sprintf("%d", cfg.Server.HTTPPort)),
			logger.String("grpc_port", fmt.Sprintf("%d", cfg.GRPC.Port)),
		),
	)

	// Initialize PostgreSQL repository
	repo, err := repository.NewPostgresRepository(&cfg.Database)
	if err != nil {
		appLogger.Fatal("failed to connect to PostgreSQL", logger.WithFields(logger.Error(err)))
	}
	defer repo.Close()

	// Initialize Kafka producer
	producer := messaging.NewKafkaProducer(cfg.Kafka.Brokers)
	defer producer.Close()

	// Initialize crypto provider
	cryptoProvider := crypto.NewCryptoProvider()

	// Initialize master key
	masterKey, err := decodeMasterKey(cfg.Security.MasterKey.Key)
	if err != nil {
		appLogger.Fatal("failed to decode master key", logger.WithFields(logger.Error(err)))
	}

	// Initialize services
	keyService := service.NewKeyService(repo, nil, producer, cryptoProvider, nil, masterKey, &cfg.Security)

	// Initialize HTTP handler
	httpHandler := handler.NewHTTPHandler(keyService, cfg)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.New()

	// Apply middleware
	ginRouter.Use(gin.Recovery())
	ginRouter.Use(handler.ErrorHandler())
	ginRouter.Use(handler.SecurityHeaders())

	// Setup routes
	setupRoutes(ginRouter, httpHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      ginRouter,
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
	}

	// Start HTTP server in goroutine
	go func() {
		appLogger.Info("HTTP server listening", logger.WithFields(logger.Int("port", cfg.Server.HTTPPort)))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("HTTP server error", logger.WithFields(logger.Error(err)))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down Key Management Service")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("server forced shutdown", logger.WithFields(logger.Error(err)))
	}

	appLogger.Info("Key Management Service exited gracefully")
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine, h *handler.HTTPHandler) {
	// Health check endpoint
	router.GET("/health", h.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Key management endpoints
		keys := v1.Group("/keys")
		{
			keys.POST("", h.CreateKey)
			keys.GET("", h.ListKeys)
			keys.GET("/:id", h.GetKey)
			keys.POST("/:id/rotate", h.RotateKey)
			keys.POST("/:id/revoke", h.RevokeKey)
			keys.GET("/:id/versions", h.GetKeyVersions)
			keys.GET("/:id/audit", h.GetKeyAuditLog)
		}

		// Data key operations
		v1.POST("/data-keys/generate", h.GenerateDataKey)

		// Encryption operations
		v1.POST("/encrypt", h.EncryptData)
		v1.POST("/decrypt", h.DecryptData)
	}
}

func decodeMasterKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}

func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "csic-key-management",
			Version:     "1.0.0",
			Environment: "development",
		},
		Server: config.ServerConfig{
			HTTPPort:    8080,
			ReadTimeout: 30,
			WriteTimeout: 30,
		},
		GRPC: config.GRPCConfig{
			Port: 50051,
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Username:        "csic_user",
			Password:        "change_in_production",
			Name:            "csic_platform",
			SSLMode:         "require",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Redis: config.RedisConfig{
			Host:        "localhost",
			Port:        6379,
			Password:    "",
			DB:          0,
			PoolSize:    10,
			KeyCacheTTL: 300,
		},
		Kafka: config.KafkaConfig{
			Brokers: []string{"localhost:9092"},
		},
		Security: config.SecurityConfig{
			MasterKey: config.MasterKeyConfig{
				Key: "ZXhhbXBsZS1tYXN0ZXIta2V5LWZvci1rZXktbWFuYWdlbWVudC1zZXJ2aWNl",
			},
			Rotation: config.RotationConfig{
				Enabled:           true,
				DefaultPeriodDays: 90,
			},
			Algorithms: []string{"AES-256-GCM", "RSA-4096", "ECDSA-P384", "Ed25519"},
		},
		Logging: config.LoggingConfig{
			Level:  "INFO",
			Format: "json",
			Output: "stdout",
		},
	}
}
