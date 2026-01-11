package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/csic/platform/blockchain/nodes/internal/handler"
	"github.com/csic/platform/blockchain/nodes/internal/messaging"
	"github.com/csic/platform/blockchain/nodes/internal/repository"
	"github.com/csic/platform/blockchain/nodes/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize database connection
	db, err := repository.NewDatabase(cfg.Database)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := repository.RunMigrations(db); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		os.Exit(1)
	}

	// Initialize Kafka producer and consumer
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.Kafka)
	if err != nil {
		fmt.Printf("Failed to create Kafka producer: %v\n", err)
		os.Exit(1)
	}
	defer kafkaProducer.Close()

	kafkaConsumer, err := messaging.NewKafkaConsumer(cfg.Kafka)
	if err != nil {
		fmt.Printf("Failed to create Kafka consumer: %v\n", err)
		os.Exit(1)
	}
	defer kafkaConsumer.Close()

	// Initialize repositories
	nodeRepo := repository.NewNodeRepository(db, cfg.Redis.KeyPrefix)
	metricsRepo := repository.NewMetricsRepository(db, redisClient, cfg.Redis.KeyPrefix)

	// Initialize services
	nodeService := service.NewNodeService(cfg, nodeRepo, kafkaProducer)
	metricsService := service.NewMetricsService(cfg, metricsRepo, kafkaProducer)
	healthService := service.NewHealthService(cfg, nodeRepo, kafkaProducer)

	// Initialize handlers
	nodeHandler := handler.NewNodeHandler(nodeService)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	healthHandler := handler.NewHealthHandler(healthService)

	// Start background services
	go healthService.StartHealthCheckLoop(ctx)

	// Setup Gin router
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup routes
	setupRoutes(router, nodeHandler, metricsHandler, healthHandler, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting Blockchain Node Manager on %s:%d\n", cfg.App.Host, cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited")
}

func setupRoutes(
	router *gin.Engine,
	nodeHandler *handler.NodeHandler,
	metricsHandler *handler.MetricsHandler,
	healthHandler *handler.HealthHandler,
	cfg *config.Config,
) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Node management endpoints
		nodes := v1.Group("/nodes")
		{
			nodes.POST("", nodeHandler.CreateNode)
			nodes.GET("", nodeHandler.ListNodes)
			nodes.GET("/:id", nodeHandler.GetNode)
			nodes.PUT("/:id", nodeHandler.UpdateNode)
			nodes.DELETE("/:id", nodeHandler.DeleteNode)
			nodes.POST("/:id/restart", nodeHandler.RestartNode)
			nodes.POST("/:id/sync", nodeHandler.ForceSync)
		}

		// Blockchain network endpoints
		networks := v1.Group("/networks")
		{
			networks.GET("", nodeHandler.ListNetworks)
			networks.GET("/:network/nodes", nodeHandler.GetNetworkNodes)
		}

		// Metrics endpoints
		metrics := v1.Group("/metrics")
		{
			metrics.GET("/node/:id", metricsHandler.GetNodeMetrics)
			metrics.GET("/node/:id/history", metricsHandler.GetNodeMetricsHistory)
			metrics.GET("/network/:network", metricsHandler.GetNetworkMetrics)
			metrics.GET("/summary", metricsHandler.GetSystemSummary)
		}

		// Health endpoints
		health := v1.Group("/health")
		{
			health.GET("", healthHandler.GetHealth)
			health.GET("/node/:id", healthHandler.GetNodeHealth)
			health.GET("/live", healthHandler.LivenessCheck)
			health.GET("/ready", healthHandler.ReadinessCheck)
		}
	}

	// Metrics endpoint for Prometheus
	router.GET("/internal/metrics", metricsHandler.PrometheusMetrics)
}
