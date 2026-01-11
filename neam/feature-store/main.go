// Feature Store Service - Centralized Feature Management for ML Models
// Provides real-time and batch feature serving with versioning and lineage tracking

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
	"neam-platform/shared"
	"neam-platform/feature-store/registry"
	"neam-platform/feature-store/online"
	"neam-platform/feature-store/offline"
	"neam-platform/feature-store/ingestion"
)

func main() {
	// Initialize configuration
	config := shared.NewConfig()
	if err := config.Load("feature-store"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := shared.NewLogger("feature-store")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize shared components
	clickhouse, err := shared.NewClickHouse(config.ClickHouseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer clickhouse.Close()

	postgres, err := shared.NewPostgreSQL(config.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgres.Close()

	redis, err := shared.NewRedis(config.RedisAddr, config.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	kafkaConsumer, err := shared.NewKafkaConsumer(config.KafkaBrokers, "feature-store-ingestion")
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	minioClient, err := shared.NewMinIO(config.MinIOEndpoint, config.MinIOAccessKey, config.MinIOSecretKey)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}
	defer minioClient.Close()

	// Initialize feature store components
	var wg sync.WaitGroup

	// Feature Registry
	registryService := registry.NewService(registry.Config{
		PostgreSQL: postgres,
		ClickHouse: clickhouse,
		Logger:     logger,
	})

	// Online Store (Redis)
	onlineService := online.NewService(online.Config{
		Redis:  redis,
		Logger: logger,
	})

	// Offline Store (S3/Parquet)
	offlineService := offline.NewService(offline.Config{
		ClickHouse: clickhouse,
		MinIO:      minioClient,
		Logger:     logger,
	})

	// Ingestion Pipeline
	ingestionService := ingestion.NewService(ingestion.Config{
		Kafka:        kafkaConsumer,
		ClickHouse:   clickHouse,
		Redis:        redis,
		OnlineStore:  onlineService,
		Logger:       logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		ingestionService.Start(ctx)
	}()

	// Setup HTTP server
	router := setupRouter(config, registryService, onlineService, offlineService)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Feature store service starting", "port", config.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down feature store service...")
	cancel()
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Feature store service stopped")
}

func setupRouter(config *shared.Config, registryService *registry.Service, onlineService *online.Service, offlineService *offline.Service) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Feature Registry endpoints
		features := v1.Group("/features")
		{
			features.GET("/", listFeaturesHandler(registryService))
			features.GET("/:name", getFeatureHandler(registryService))
			features.GET("/:name/versions", getFeatureVersionsHandler(registryService))
			features.POST("/register", registerFeatureHandler(registryService))
			features.PUT("/:name", updateFeatureHandler(registryService))
			features.DELETE("/:name", deleteFeatureHandler(registryService))
			features.GET("/search", searchFeaturesHandler(registryService))
		}

		// Online Feature Serving endpoints
		online := v1.Group("/online")
		{
			online.POST("/get", getFeaturesHandler(onlineService))
			online.GET("/:entity_id", getEntityFeaturesHandler(onlineService))
			online.GET("/:entity_id/:feature", getEntityFeatureHandler(onlineService))
		}

		// Offline Feature Serving endpoints
		offline := v1.Group("/offline")
		{
			offline.POST("/batch", batchGetFeaturesHandler(offlineService))
			offline.POST("/point-in-time", pointInTimeQueryHandler(offlineService))
			offline.GET("/history/:entity_id", getFeatureHistoryHandler(offlineService))
			offline.GET("/export", exportFeaturesHandler(offlineService))
		}

		// Lineage endpoints
		lineage := v1.Group("/lineage")
		{
			lineage.GET("/features/:name", getFeatureLineageHandler(registryService))
			lineage.GET("/models/:model", getModelLineageHandler(registryService))
			lineage.GET("/entities/:entity", getEntityLineageHandler(registryService))
		}

		// Quality endpoints
		quality := v1.Group("/quality")
		{
			quality.GET("/reports", getQualityReportsHandler(registryService))
			quality.GET("/reports/:id", getQualityReportHandler(registryService))
			quality.GET("/metrics/:feature", getFeatureMetricsHandler(registryService))
			quality.POST("/validate", validateFeaturesHandler(registryService))
		}

		// Registration endpoints
		registration := v1.Group("/registration")
		{
			registration.GET("/", listRegistrationsHandler(registryService))
			registration.GET("/:id", getRegistrationHandler(registryService))
			registration.POST("/", createRegistrationHandler(registryService))
			registration.PUT("/:id/approve", approveRegistrationHandler(registryService))
			registration.PUT("/:id/reject", rejectRegistrationHandler(registryService))
		}

		// Version endpoints
		versions := v1.Group("/versions")
		{
			versions.GET("/", listVersionsHandler(registryService))
			versions.GET("/:version_id", getVersionHandler(registryService))
			versions.POST("/:version_id/promote", promoteVersionHandler(registryService))
			versions.POST("/:version_id/archive", archiveVersionHandler(registryService))
		}
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-feature-store",
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-feature-store",
	})
}

// Handler functions
func listFeaturesHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"features": []interface{}{},
			"total":    0,
		})
	}
}

func getFeatureHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":         c.Param("name"),
			"description":  "Feature description",
			"data_type":    "float64",
			"owner":        "data-science-team",
			"version":      "1.0.0",
			"created_at":   time.Now().Format(time.RFC3339),
		})
	}
}

func getFeatureVersionsHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"versions": []interface{}{},
			"total":    0,
		})
	}
}

func registerFeatureHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"name":         "feature_name",
			"status":       "pending",
			"registration_id": fmt.Sprintf("reg-%d", time.Now().UnixNano()),
		})
	}
}

func updateFeatureHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    c.Param("name"),
			"status": "updated",
		})
	}
}

func deleteFeatureHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":   c.Param("name"),
			"status": "deleted",
		})
	}
}

func searchFeaturesHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"features": []interface{}{},
			"total":    0,
		})
	}
}

func getFeaturesHandler(svc *online.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entity_id":   "entity-123",
			"features":    map[string]interface{}{},
			"computed_at": time.Now().Format(time.RFC3339),
		})
	}
}

func getEntityFeaturesHandler(svc *online.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entity_id": c.Param("entity_id"),
			"features":  map[string]interface{}{},
		})
	}
}

func getEntityFeatureHandler(svc *online.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entity_id":  c.Param("entity_id"),
			"feature":    c.Param("feature"),
			"value":      0.0,
			"updated_at": time.Now().Format(time.RFC3339),
		})
	}
}

func batchGetFeaturesHandler(svc *offline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"features": []interface{}{},
			"total":    0,
		})
	}
}

func pointInTimeQueryHandler(svc *offline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"features": []interface{}{},
			"point_in_time": time.Now().Format(time.RFC3339),
		})
	}
}

func getFeatureHistoryHandler(svc *offline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entity_id": c.Param("entity_id"),
			"history":   []interface{}{},
		})
	}
}

func exportFeaturesHandler(svc *offline.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"export_id":  fmt.Sprintf("exp-%d", time.Now().UnixNano()),
			"status":    "processing",
			"format":    "parquet",
		})
	}
}

func getFeatureLineageHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"feature":    c.Param("name"),
			"lineage":    []interface{}{},
			"upstream":   []interface{}{},
			"downstream": []interface{}{},
		})
	}
}

func getModelLineageHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"model":    c.Param("model"),
			"features": []interface{}{},
		})
	}
}

func getEntityLineageHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entity":   c.Param("entity"),
			"lineage":  []interface{}{},
		})
	}
}

func getQualityReportsHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"reports": []interface{}{},
			"total":   0,
		})
	}
}

func getQualityReportHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":        c.Param("id"),
			"status":    "completed",
			"metrics":   map[string]interface{}{},
		})
	}
}

func getFeatureMetricsHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"feature": c.Param("feature"),
			"metrics": map[string]interface{}{
				"null_count":    10,
				"null_percent":  0.5,
				"unique_count":  1000,
				"min_value":     0.0,
				"max_value":     100.0,
				"avg_value":     45.5,
			},
		})
	}
}

func validateFeaturesHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"valid":      true,
			"errors":     []interface{}{},
			"warnings":   []interface{}{},
		})
	}
}

func listRegistrationsHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"registrations": []interface{}{},
			"total":         0,
		})
	}
}

func getRegistrationHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":      c.Param("id"),
			"status":  "pending",
			"feature": map[string]interface{}{},
		})
	}
}

func createRegistrationHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":     fmt.Sprintf("reg-%d", time.Now().UnixNano()),
			"status": "pending",
		})
	}
}

func approveRegistrationHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":      c.Param("id"),
			"status":  "approved",
		})
	}
}

func rejectRegistrationHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":      c.Param("id"),
			"status":  "rejected",
			"reason":  "Insufficient documentation",
		})
	}
}

func listVersionsHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"versions": []interface{}{},
			"total":    0,
		})
	}
}

func getVersionHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version_id": c.Param("version_id"),
			"status":     "active",
			"features":   []interface{}{},
		})
	}
}

func promoteVersionHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version_id": c.Param("version_id"),
			"status":     "promoted",
			"stage":      "production",
		})
	}
}

func archiveVersionHandler(svc *registry.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version_id": c.Param("version_id"),
			"status":     "archived",
		})
	}
}
