package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"csic-platform/extended-services/cbdc-monitoring/internal/handlers"
	"csic-platform/extended-services/cbdc-monitoring/internal/repository"
	"csic-platform/extended-services/cbdc-monitoring/internal/service"
	"csic-platform/shared/config"
	"csic-platform/shared/logger"
	"csic-platform/shared/queue"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger("cbdc-monitoring", cfg.Observability.Logging.Format)
	log.Info("Starting CBDC Monitoring Service")

	// Initialize database connection
	db, err := initDatabase(cfg.Database, log)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis client
	redisClient, err := initRedis(cfg.Redis, log)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize Kafka producer/consumer
	kafkaManager, err := queue.NewKafkaManager(cfg.Kafka, log)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka: %v", err)
	}
	defer kafkaManager.Close()

	// Initialize services
	cbdcService := service.NewCBDCMonitoringService(db, redisClient, kafkaManager, cfg, log)

	// Initialize handlers
	httpHandler := handlers.NewCBdCHandler(cbdcService, log)

	// Setup Gin router
	router := setupRouter(httpHandler, cfg)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info("Shutting down CBDC Monitoring Service...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("Server shutdown error: %v", err)
		}
	}()

	log.Infof("CBDC Monitoring Service listening on %s:%d", cfg.App.Host, cfg.App.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Info("CBDC Monitoring Service stopped")
}

func initDatabase(cfg config.DatabaseConfig, log *logger.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name)

	gormLogger := logger.New(
		logger.WithSkip(1),
		logger.SetLogLevel(logger.Info),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Auto-migrate schema
	if err := db.AutoMigrate(&repository.Wallet{}, &repository.Transaction{}, &repository.PolicyRule{}); err != nil {
		log.Warnf("Auto-migration failed: %v", err)
	}

	return db, nil
}

func initRedis(cfg config.RedisConfig, log *logger.Logger) (*queue.RedisClient, error) {
	client := queue.NewRedisClient(cfg)
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return client, nil
}

func setupRouter(handler *handlers.CBdcHandler, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.GinLogger())
	router.Use(CORSMiddleware())

	// Health endpoints
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Readiness)
	router.GET("/live", handler.Liveness)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// Wallet management
		api.POST("/wallets", handler.CreateWallet)
		api.GET("/wallets/:id", handler.GetWallet)
		api.PUT("/wallets/:id", handler.UpdateWallet)
		api.DELETE("/wallets/:id", handler.DeleteWallet)
		api.GET("/wallets/:id/balance", handler.GetWalletBalance)

		// Transaction processing
		api.POST("/transactions", handler.CreateTransaction)
		api.GET("/transactions/:id", handler.GetTransaction)
		api.GET("/transactions", handler.ListTransactions)
		api.POST("/transactions/:id/approve", handler.ApproveTransaction)
		api.POST("/transactions/:id/reject", handler.RejectTransaction)

		// Policy management
		api.POST("/policies", handler.CreatePolicy)
		api.GET("/policies/:id", handler.GetPolicy)
		api.GET("/policies", handler.ListPolicies)
		api.PUT("/policies/:id", handler.UpdatePolicy)
		api.DELETE("/policies/:id", handler.DeletePolicy)

		// Monitoring and analytics
		api.GET("/monitoring/stats", handler.GetMonitoringStats)
		api.GET("/monitoring/velocity/:walletId", handler.GetTransactionVelocity)
		api.GET("/monitoring/violations", handler.GetViolations)

		// Compliance endpoints
		api.POST("/compliance/check", handler.ComplianceCheck)
		api.GET("/compliance/audit/:walletId", handler.GetAuditTrail)
	}

	return router
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
