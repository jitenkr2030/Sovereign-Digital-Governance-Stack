// NEAM Sensing Layer - Main Entry Point
// Economic Data Intake and Normalization

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"neam-platform/sensing/common"
	"neam-platform/sensing/energy"
	"neam-platform/sensing/industry"
	"neam-platform/sensing/payments"
	"neam-platform/sensing/transport"
)

func main() {
	// Initialize configuration
	config := common.NewConfig()
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := common.NewLogger(config)
	ctx := context.Background()

	// Initialize rate limiter
	rateLimiter := common.NewRateLimiter(config.RateLimitRequests, config.RateLimitWindow)

	// Initialize anonymizer
	anonymizer := common.NewAnonymizer(config.AnonymizationSalt)

	// Initialize validator
	validator := common.NewValidator()

	// Initialize Kafka producer for outgoing messages
	producer, err := common.NewKafkaProducer(config.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Initialize services
	paymentService := payments.NewService(payments.Config{
		Logger:       logger,
		Validator:    validator,
		Anonymizer:   anonymizer,
		RateLimiter:  rateLimiter,
		Producer:     producer,
		KafkaTopic:   config.KafkaPrefix + ".payments.raw",
		ClickHouseDSN: config.ClickHouseDSN,
	})

	energyService := energy.NewService(energy.Config{
		Logger:       logger,
		Validator:    validator,
		RateLimiter:  rateLimiter,
		Producer:     producer,
		KafkaTopic:   config.KafkaPrefix + ".energy.raw",
	})

	transportService := transport.NewService(transport.Config{
		Logger:       logger,
		Validator:    validator,
		RateLimiter:  rateLimiter,
		Producer:     producer,
		KafkaTopic:   config.KafkaPrefix + ".transport.raw",
	})

	industryService := industry.NewService(industry.Config{
		Logger:       logger,
		Validator:    validator,
		RateLimiter:  rateLimiter,
		Producer:     producer,
		KafkaTopic:   config.KafkaPrefix + ".industry.raw",
	})

	// Start data ingestion workers
	go paymentService.StartIngestion(ctx)
	go energyService.StartIngestion(ctx)
	go transportService.StartIngestion(ctx)
	go industryService.StartIngestion(ctx)

	// Setup HTTP server
	router := setupRouter(config, paymentService, energyService, transportService, industryService, rateLimiter)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Sensing layer starting", "port", config.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down sensing layer...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Sensing layer stopped")
}

func setupRouter(config *common.Config, paymentSvc, energySvc, transportSvc, industrySvc interface{}, rateLimiter *common.RateLimiter) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(rateLimiter.Middleware())

	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Payment endpoints
		payments := v1.Group("/payments")
		{
			payments.POST("/ingest", paymentHandler(paymentSvc))
			payments.GET("/status", paymentStatusHandler(paymentSvc))
		}

		// Energy endpoints
		energy := v1.Group("/energy")
		{
			energy.POST("/telemetry", energyTelemetryHandler(energySvc))
			energy.GET("/load", energyLoadHandler(energySvc))
		}

		// Transport endpoints
		transport := v1.Group("/transport")
		{
			transport.POST("/logistics", logisticsHandler(transportSvc))
			transport.POST("/fuel", fuelHandler(transportSvc))
		}

		// Industry endpoints
		industry := v1.Group("/industry")
		{
			industry.POST("/production", productionHandler(industrySvc))
			industry.GET("/output", outputHandler(industrySvc))
		}
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-sensing",
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-sensing",
	})
}

func paymentHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation placeholder
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Payment ingestion queued",
			"id":      generateID(),
		})
	}
}

func paymentStatusHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "operational",
			"ingestion_rate": "1500/min",
			"last_update": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func energyTelemetryHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Energy telemetry queued",
			"id":      generateID(),
		})
	}
}

func energyLoadHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "operational",
			"load_factor": 0.72,
			"last_update": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func logisticsHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Logistics data queued",
			"id":      generateID(),
		})
	}
}

func fuelHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Fuel data queued",
			"id":      generateID(),
		})
	}
}

func productionHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Production data queued",
			"id":      generateID(),
		})
	}
}

func outputHandler(svc interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "operational",
			"output_index": 1.05,
			"last_update": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func generateID() string {
	return fmt.Sprintf("neam-%d-%s", time.Now().UnixNano(), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
