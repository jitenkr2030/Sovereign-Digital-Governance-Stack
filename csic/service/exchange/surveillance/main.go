package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csic/surveillance/internal/handler"
	"github.com/csic/surveillance/internal/repository"
	"github.com/csic/surveillance/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize repository layer
	alertRepo, err := repository.NewPostgresAlertRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize alert repository: %v", err)
	}
	defer alertRepo.Close()

	marketRepo, err := repository.NewTimescaleMarketRepository(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize market repository: %v", err)
	}
	defer marketRepo.Close()

	// Initialize service layer
	analysisSvc := service.NewAnalysisService(marketRepo, alertRepo)
	ingestionSvc := service.NewIngestionService(marketRepo, analysisSvc, cfg.Ingestion)
	alertSvc := service.NewAlertService(alertRepo)

	// Initialize handlers
	httpHandler := handler.NewHTTPHandler(alertSvc, analysisSvc)
	wsHandler := handler.NewWebSocketHandler(ingestionSvc)

	// Setup Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Alert management endpoints
		api.GET("/alerts", httpHandler.ListAlerts)
		api.GET("/alerts/:id", httpHandler.GetAlert)
		api.PATCH("/alerts/:id/status", httpHandler.UpdateAlertStatus)

		// Market data endpoints
		api.GET("/markets/:exchange_id/summary", httpHandler.GetMarketSummary)

		// Analysis endpoints
		api.POST("/analysis/wash-trade", httpHandler.AnalyzeWashTrading)
		api.POST("/analysis/spoofing", httpHandler.AnalyzeSpoofing)

		// Configuration endpoints
		api.GET("/config/thresholds", httpHandler.GetThresholds)
		api.PUT("/config/thresholds", httpHandler.UpdateThresholds)
	}

	// WebSocket endpoint for real-time market data ingestion
	router.GET("/ws/ingest", wsHandler.HandleIngestion)
	router.GET("/ws/alerts", wsHandler.HandleAlertStream)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Exchange Surveillance Service on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
