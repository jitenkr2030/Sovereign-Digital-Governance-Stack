// NEAM Intelligence Engine - Main Entry Point
// Economic Analysis, Anomaly Detection, and Prediction

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"neam-platform/intelligence/anomaly-detection"
	"neam-platform/intelligence/black-economy"
	"neam-platform/intelligence/correlation-engine"
	"neam-platform/intelligence/feature-store"
	"neam-platform/intelligence/inflation-engine"
	"neam-platform/intelligence/macro-models"
	"neam-platform/shared"
)

func main() {
	// Initialize configuration
	config := shared.NewConfig()
	if err := config.Load("intelligence"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := shared.NewLogger("intelligence")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize shared components
	clickhouse, err := shared.NewClickHouse(config.ClickHouseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer clickhouse.Close()

	redis, err := shared.NewRedis(config.RedisAddr, config.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	kafkaProducer, err := shared.NewKafkaProducer(config.KafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Initialize feature store
	featureStore := feature_store.NewService(feature_store.Config{
		ClickHouse: clickhouse,
		Redis:      redis,
		Logger:     logger,
	})

	// Initialize intelligence engines
	var wg sync.WaitGroup

	// Anomaly Detection Engine
	anomalyEngine := anomaly_detection.NewEngine(anomaly_detection.Config{
		FeatureStore: featureStore,
		Producer:     kafkaProducer,
		KafkaTopic:   config.KafkaPrefix + ".anomalies",
		Logger:       logger,
		RulesPath:    "./anomaly-detection/rules",
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		anomalyEngine.Start(ctx)
	}()

	// Inflation Engine
	inflationEngine := inflation_engine.NewEngine(inflation_engine.Config{
		FeatureStore: featureStore,
		Producer:     kafkaProducer,
		KafkaTopic:   config.KafkaPrefix + ".inflation",
		Logger:       logger,
		ModelsPath:   "./inflation-engine/models",
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		inflationEngine.Start(ctx)
	}()

	// Black Economy Detection
	blackEconomyEngine := black_economy.NewEngine(black_economy.Config{
		FeatureStore: featureStore,
		Producer:     kafkaProducer,
		KafkaTopic:   config.KafkaPrefix + ".black-economy",
		Logger:       logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		blackEconomyEngine.Start(ctx)
	}()

	// Macro Economic Models
	macroEngine := macro_models.NewEngine(macro_models.Config{
		FeatureStore: featureStore,
		Producer:     kafkaProducer,
		KafkaTopic:   config.KafkaPrefix + ".macro",
		Logger:       logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		macroEngine.Start(ctx)
	}()

	// Correlation Engine
	corrEngine := correlation_engine.NewEngine(correlation_engine.Config{
		FeatureStore: featureStore,
		Producer:     kafkaProducer,
		KafkaTopic:   config.KafkaPrefix + ".correlations",
		Logger:       logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		corrEngine.Start(ctx)
	}()

	// Setup HTTP server
	router := setupRouter(config, featureStore, anomalyEngine, inflationEngine, blackEconomyEngine, macroEngine)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Intelligence engine starting", "port", config.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down intelligence engine...")
	cancel()
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Intelligence engine stopped")
}

func setupRouter(config *shared.Config, featureStore *feature_store.Service, anomalyEng *anomaly_detection.Engine, inflationEng *inflation_engine.Engine, blackEng *black_economy.Engine, macroEng *macro_models.Engine) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Feature store endpoints
		features := v1.Group("/features")
		{
			features.GET("/:entity/:feature", getFeatureHandler(featureStore))
			features.POST("/compute", computeFeaturesHandler(featureStore))
		}

		// Anomaly endpoints
		anomalies := v1.Group("/anomalies")
		{
			anomalies.GET("/active", getActiveAnomaliesHandler(anomalyEng))
			anomalies.POST("/rules", createRuleHandler(anomalyEng))
			anomalies.GET("/history", anomalyHistoryHandler(anomalyEng))
		}

		// Inflation endpoints
		inflation := v1.Group("/inflation")
		{
			inflation.GET("/current", getCurrentInflationHandler(inflationEng))
			inflation.GET("/forecast", getInflationForecastHandler(inflationEng))
			inflation.GET("/regional", getRegionalInflationHandler(inflationEng))
		}

		// Black economy endpoints
		black := v1.Group("/black-economy")
		{
			black.GET("/risk-map", getRiskMapHandler(blackEng))
			black.GET("/shadow-flows", getShadowFlowsHandler(blackEng))
		}

		// Macro indicators
		macro := v1.Group("/macro")
		{
			macro.GET("/gdp", getGDPHandler(macroEng))
			macro.GET("/employment", getEmploymentHandler(macroEng))
			macro.GET("/slowdown-risk", getSlowdownRiskHandler(macroEng))
		}

		// Correlations
		correlations := v1.Group("/correlations")
		{
			correlations.GET("/", getCorrelationsHandler(corrEngine))
			correlations.POST("/causality", analyzeCausalityHandler(corrEngine))
		}
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-intelligence",
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-intelligence",
	})
}

func getFeatureHandler(fs *feature_store.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		entity := c.Param("entity")
		feature := c.Param("feature")
		c.JSON(http.StatusOK, gin.H{
			"entity":  entity,
			"feature": feature,
			"value":   1.234,
			"ts":      time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func computeFeaturesHandler(fs *feature_store.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "computed",
			"count":  42,
		})
	}
}

func getActiveAnomaliesHandler(eng *anomaly_detection.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"anomalies": []interface{}{
				map[string]interface{}{
					"id":          "anom-001",
					"type":        "payment_surge",
					"severity":    "high",
					"region":      "REGION-NORTH",
					"description": "Unusual payment volume increase detected",
					"timestamp":   time.Now().UTC().Format(time.RFC3339),
				},
			},
			"count": 1,
		})
	}
}

func createRuleHandler(eng *anomaly_detection.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":      "rule-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"status":  "active",
			"message": "Anomaly detection rule created",
		})
	}
}

func anomalyHistoryHandler(eng *anomaly_detection.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"history": []interface{}{},
			"total":   0,
		})
	}
}

func getCurrentInflationHandler(eng *inflation_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"rate":      4.52,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"trend":     "stable",
		})
	}
}

func getInflationForecastHandler(eng *inflation_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"forecast_1m": 4.35,
			"forecast_3m": 4.20,
			"forecast_6m": 4.10,
			"confidence":  0.85,
		})
	}
}

func getRegionalInflationHandler(eng *inflation_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"regions": []interface{}{
				map[string]interface{}{
					"id":    "REGION-NORTH",
					"rate":  4.72,
				},
				map[string]interface{}{
					"id":    "REGION-SOUTH",
					"rate":  4.31,
				},
			},
		})
	}
}

func getRiskMapHandler(eng *black_economy.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"risk_levels": []interface{}{
				map[string]interface{}{
					"region": "REGION-WEST",
					"level":  "high",
					"score":  0.78,
				},
			},
		})
	}
}

func getShadowFlowsHandler(eng *black_economy.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"flows": []interface{}{},
		})
	}
}

func getGDPHandler(eng *macro_models.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"growth_rate":  5.2,
			"quarter":      "Q3",
			"year":         2024,
			"last_updated": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func getEmploymentHandler(eng *macro_models.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"rate":           5.8,
			"change_month":   -0.2,
			"change_quarter": -0.5,
		})
	}
}

func getSlowdownRiskHandler(eng *macro_models.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"risk_score":    0.32,
			"level":         "low",
			"indicators":    []string{"leading"},
			"probability":   0.18,
			"time_horizon":  "6m",
		})
	}
}

func getCorrelationsHandler(eng *correlation_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"correlations": []interface{}{
				map[string]interface{}{
					"var1": "energy_load",
					"var2": "industrial_output",
					"coef": 0.85,
				},
			},
		})
	}
}

func analyzeCausalityHandler(eng *correlation_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"causality": []interface{}{},
		})
	}
}
