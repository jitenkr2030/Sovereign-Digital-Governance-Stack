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

	"github.com/csic-platform/services/api-gateway/internal/adapter/auth"
	"github.com/csic-platform/services/api-gateway/internal/adapter/messaging"
	"github.com/csic-platform/services/api-gateway/internal/adapter/repository"
	"github.com/csic-platform/services/api-gateway/internal/config"
	"github.com/csic-platform/services/api-gateway/internal/core/ports"
	"github.com/csic-platform/services/api-gateway/internal/core/service"
	"github.com/csic-platform/services/api-gateway/internal/handler"
	"github.com/csic-platform/services/api-gateway/internal/middleware"
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
		// Fallback to default configuration if file loading fails
		cfg = getDefaultConfig()
		fmt.Printf("Warning: Failed to load config file, using defaults: %v\n", err)
	}

	// Initialize logger using shared logger package
	appLogger, err := logger.NewLogger(logger.Config{
		ServiceName:  "api-gateway",
		LogLevel:     cfg.Logging.Level,
		OutputPath:   cfg.Logging.Output,
		AuditLogPath: cfg.Logging.Path,
		Development:  cfg.App.Environment == "development",
		JSONOutput:   cfg.Logging.Format == "json",
	})
	if err != nil {
		fmt.Printf("Fatal: Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Close()

	// Log startup
	appLogger.Info("starting API gateway",
		logger.WithFields(
			logger.String("port", fmt.Sprintf("%d", cfg.Server.Port)),
			logger.String("environment", cfg.App.Environment),
		),
	)

	// Initialize repositories
	var repo ports.Repository
	var cache ports.CacheRepository

	// Initialize PostgreSQL repository
	repo, err = repository.NewPostgresRepository(&cfg.Database)
	if err != nil {
		appLogger.Warn("Failed to connect to PostgreSQL, using in-memory storage",
			logger.WithFields(logger.Error(err)))
		// In production, this should fail hard
	}

	// Initialize Redis cache (optional, for rate limiting and sessions)
	// cache = redis.NewRedisCache(&cfg.Redis)

	// Initialize Kafka producer
	producer := messaging.NewKafkaProducer(cfg.Kafka.Brokers)
	defer producer.Close()

	// Initialize services
	authService := auth.NewAuthService(cfg.Security.JWT.Secret)
	gatewayService := service.NewGatewayService(repo, cache, producer, authService)

	// Initialize HTTP handler
	httpHandler := handler.NewHTTPHandler(gatewayService, cfg)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(appLogger, cfg.Security.JWT.Secret)
	rateLimiter := middleware.NewRateLimitMiddleware(appLogger)
	loggingMiddleware := middleware.NewLoggingMiddleware(appLogger)
	securityHeaders := middleware.NewSecurityHeadersMiddleware()
	corsMiddleware := middleware.NewCORSMiddleware()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.New()

	// Apply global middleware
	ginRouter.Use(gin.Recovery())
	ginRouter.Use(handler.ErrorHandler())
	ginRouter.Use(loggingMiddleware.Middleware())
	ginRouter.Use(securityHeaders.Headers())
	ginRouter.Use(corsMiddleware.Middleware())
	ginRouter.Use(rateLimiter.Middleware())

	// Apply authentication middleware to API routes
	authRequired := v1.Group("")
	authRequired.Use(authMiddleware.Authenticate())
	{
		// Dashboard
		authRequired.GET("/dashboard/stats", h.GetDashboardStats)

		// Alerts
		authRequired.GET("/alerts", h.GetAlerts)
		authRequired.GET("/alerts/:id", h.GetAlertByID)
		authRequired.POST("/alerts/:id/acknowledge", h.AcknowledgeAlert)

		// Exchanges
		authRequired.GET("/exchanges", h.GetExchanges)
		authRequired.GET("/exchanges/:id", h.GetExchangeByID)
		authRequired.POST("/exchanges/:id/suspend", h.SuspendExchange)

		// Wallets
		authRequired.GET("/wallets", h.GetWallets)
		authRequired.GET("/wallets/:id", h.GetWalletByID)
		authRequired.POST("/wallets/:id/freeze", h.FreezeWallet)

		// Miners
		authRequired.GET("/miners", h.GetMiners)
		authRequired.GET("/miners/:id", h.GetMinerByID)

		// Compliance
		authRequired.GET("/compliance/reports", h.GetComplianceReports)
		authRequired.POST("/compliance/reports", h.GenerateComplianceReport)

		// Audit Logs
		authRequired.GET("/audit/logs", h.GetAuditLogs)

		// Blockchain
		authRequired.GET("/blockchain/status", h.GetBlockchainStatus)

		// Users
		authRequired.GET("/users/me", h.GetCurrentUser)
		authRequired.GET("/users", h.GetUsers)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      ginRouter,
		ReadTimeout:  cfg.Server.GetReadTimeout(),
		WriteTimeout: cfg.Server.GetWriteTimeout(),
		IdleTimeout:  cfg.Server.GetIdleTimeout(),
	}

	// Start server in goroutine
	go func() {
		appLogger.Info("API gateway listening",
			logger.WithFields(
				logger.Int("port", cfg.Server.Port),
			),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("server error",
				logger.WithFields(
					logger.Error(err),
				),
			)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("server forced shutdown",
			logger.WithFields(
				logger.Error(err),
			),
		)
	}

	appLogger.Info("server exited gracefully")
}

// setupRoutes configures all API routes (public endpoints)
func setupRoutes(router *gin.Engine, h *handler.HTTPHandler) {
	// Health check endpoint (public)
	router.GET("/health", h.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public endpoints (authentication not required)
		// Add any public endpoints here if needed
	}
}

func getDefaultConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "csic-api-gateway",
			Version:     "1.0.0",
			Environment: "development",
		},
		Server: config.ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
			MaxBodySize:  10485760,
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
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		Kafka: config.KafkaConfig{
			Brokers: []string{"localhost:9092"},
		},
		Security: config.SecurityConfig{
			JWT: config.JWTConfig{
				Secret:      "default-secret-change-in-production",
				ExpiryHours: 8,
				Algorithm:   "HS256",
			},
		},
		Logging: config.LoggingConfig{
			Level:   "INFO",
			Format:  "json",
			Output:  "stdout",
		},
	}
}
