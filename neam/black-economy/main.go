// Black Economy Monitoring Service - Main Entry Point
// Detection of informal economy, tax evasion, and hidden economic networks

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
	"neam-platform/black-economy/anomalydetection"
	"neam-platform/black-economy/entitylinkage"
	"neam-platform/black-economy/forensicanalytics"
	"neam-platform/black-economy/risk_scoring"
)

func main() {
	// Initialize configuration
	config := shared.NewConfig()
	if err := config.Load("black-economy"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := shared.NewLogger("black-economy")
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

	neo4j, err := shared.NewNeo4j(config.Neo4jURI, config.Neo4jUser, config.Neo4jPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4j.Close()

	elasticsearch, err := shared.NewElasticsearch(config.ElasticsearchAddrs)
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	defer elasticsearch.Close()

	// Initialize black economy components
	var wg sync.WaitGroup

	// Anomaly Detection Service
	anomalyService := anomalydetection.NewService(anomalydetection.Config{
		ClickHouse: clickhouse,
		Redis:      redis,
		Logger:     logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		anomalyService.Start(ctx)
	}()

	// Entity Linkage Service
	entityService := entitylinkage.NewService(entitylinkage.Config{
		Neo4j:      neo4j,
		ClickHouse: clickhouse,
		Redis:      redis,
		Logger:     logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		entityService.Start(ctx)
	}()

	// Forensic Analytics Service
	forensicService := forensicanalytics.NewService(forensicanalytics.Config{
		ClickHouse:    clickhouse,
		PostgreSQL:    postgres,
		Neo4j:         neo4j,
		Elasticsearch: elasticsearch,
		Logger:        logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		forensicService.Start(ctx)
	}()

	// Risk Scoring Service
	riskService := risk_scoring.NewService(risk_scoring.Config{
		ClickHouse: clickhouse,
		PostgreSQL: postgres,
		Redis:      redis,
		Logger:     logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		riskService.Start(ctx)
	}()

	// Setup HTTP server
	router := setupRouter(config, anomalyService, entityService, forensicService, riskService)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Black economy monitoring service starting", "port", config.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down black economy monitoring service...")
	cancel()
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Black economy monitoring service stopped")
}

func setupRouter(config *shared.Config, anomalyService *anomalydetection.Service, entityService *entitylinkage.Service, forensicService *forensicanalytics.Service, riskService *risk_scoring.Service) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(forensicService.AuthMiddleware())

	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Anomaly detection endpoints
		anomalies := v1.Group("/anomalies")
		{
			anomalies.GET("/", listAnomaliesHandler(anomalyService))
			anomalies.GET("/:id", getAnomalyHandler(anomalyService))
			anomalies.POST("/cash-intensive", detectCashAnomaliesHandler(anomalyService))
			anomalies.POST("/benford", analyzeBenfordHandler(anomalyService))
			anomalies.POST("/circular-transactions", detectCircularHandler(anomalyService))
		}

		// Entity linkage endpoints
		entities := v1.Group("/entities")
		{
			entities.GET("/", searchEntitiesHandler(entityService))
			entities.GET("/:id", getEntityHandler(entityService))
			entities.GET("/:id/network", getEntityNetworkHandler(entityService))
			entities.GET("/:id/connections", getEntityConnectionsHandler(entityService))
			entities.POST("/link", linkEntitiesHandler(entityService))
			entities.GET("/clusters", getClustersHandler(entityService))
		}

		// Forensic analytics endpoints
		forensic := v1.Group("/forensic")
		{
			forensic.GET("/cases", listCasesHandler(forensicService))
			forensic.POST("/cases", createCaseHandler(forensicService))
			forensic.GET("/cases/:id", getCaseHandler(forensicService))
			forensic.POST("/cases/:id/investigate", investigateCaseHandler(forensicService))
			forensic.GET("/reports", listReportsHandler(forensicService))
			forensic.POST("/reports/generate", generateReportHandler(forensicService))
		}

		// Risk scoring endpoints
		risk := v1.Group("/risk")
		{
			risk.GET("/entities/:id", getEntityRiskHandler(riskService))
			risk.GET("/sectors/:sector", getSectorRiskHandler(riskService))
			risk.GET("/regions/:region", getRegionRiskHandler(riskService))
			risk.GET("/scores/:id/history", getRiskHistoryHandler(riskService))
			risk.GET("/leaderboard", getRiskLeaderboardHandler(riskService))
			risk.POST("/calculate", calculateRiskHandler(riskService))
		}

		// Dashboard endpoints
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/overview", getDashboardOverviewHandler(forensicService, riskService))
			dashboard.GET("/heatmaps/risk", getRiskHeatmapHandler(riskService))
			dashboard.GET("/heatmaps/sector", getSectorHeatmapHandler(riskService))
			dashboard.GET("/trends/anomalies", getAnomalyTrendsHandler(anomalyService))
		}
	}

	return router
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-black-economy",
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-black-economy",
	})
}

// Handler functions
func listAnomaliesHandler(svc *anomalydetection.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"anomalies": []interface{}{},
			"total":     0,
		})
	}
}

func getAnomalyHandler(svc *anomalydetection.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":     c.Param("id"),
			"type":   "cash_anomaly",
			"status": "detected",
		})
	}
}

func detectCashAnomaliesHandler(svc *anomalydetection.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "processing",
			"job_id": fmt.Sprintf("job-%d", time.Now().UnixNano()),
		})
	}
}

func analyzeBenfordHandler(svc *anomalydetection.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "analyzing",
			"deviation":   0.05,
			"significance": "high",
		})
	}
}

func detectCircularHandler(svc *anomalydetection.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "analyzing",
			"cycles_found": 0,
		})
	}
}

func searchEntitiesHandler(svc *entitylinkage.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entities": []interface{}{},
			"total":    0,
		})
	}
}

func getEntityHandler(svc *entitylinkage.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":           c.Param("id"),
			"name":         "Entity Name",
			"type":         "company",
			"risk_score":   0.45,
		})
	}
}

func getEntityNetworkHandler(svc *entitylinkage.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"nodes": []interface{}{},
			"edges": []interface{}{},
		})
	}
}

func getEntityConnectionsHandler(svc *entitylinkage.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"connections": []interface{}{},
			"degrees":     2,
		})
	}
}

func linkEntitiesHandler(svc *entitylinkage.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "linked",
			"link_type":  "shared_address",
		})
	}
}

func getClustersHandler(svc *entitylinkage.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"clusters": []interface{}{},
			"total":    0,
		})
	}
}

func listCasesHandler(svc *forensicanalytics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"cases": []interface{}{},
			"total": 0,
		})
	}
}

func createCaseHandler(svc *forensicanalytics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":     fmt.Sprintf("case-%d", time.Now().UnixNano()),
			"status": "open",
		})
	}
}

func getCaseHandler(svc *forensicanalytics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":     c.Param("id"),
			"status": "open",
		})
	}
}

func investigateCaseHandler(svc *forensicanalytics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "investigating",
			"findings":  []interface{}{},
		})
	}
}

func listReportsHandler(svc *forensicanalytics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"reports": []interface{}{},
			"total":   0,
		})
	}
}

func generateReportHandler(svc *forensicanalytics.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{
			"id":        fmt.Sprintf("rpt-%d", time.Now().UnixNano()),
			"status":    "generating",
			"format":    "pdf",
		})
	}
}

func getEntityRiskHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"entity_id":  c.Param("id"),
			"score":      0.65,
			"level":      "high",
			"factors":    []interface{}{},
		})
	}
}

func getSectorRiskHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"sector":    c.Param("sector"),
			"avg_score": 0.35,
			"trend":     "stable",
		})
	}
}

func getRegionRiskHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"region":   c.Param("region"),
			"score":    0.42,
			"entities": 150,
		})
	}
}

func getRiskHistoryHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"history": []interface{}{},
		})
	}
}

func getRiskLeaderboardHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"top_risks": []interface{}{},
		})
	}
}

func calculateRiskHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"score":    0.55,
			"level":    "medium",
			"confidence": 0.85,
		})
	}
}

func getDashboardOverviewHandler(forensicSvc *forensicanalytics.Service, riskSvc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"total_entities":    5000,
			"active_anomalies":  45,
			"high_risk_count":   120,
			"open_cases":        15,
			"trend":             "increasing",
		})
	}
}

func getRiskHeatmapHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []interface{}{},
		})
	}
}

func getSectorHeatmapHandler(svc *risk_scoring.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []interface{}{},
		})
	}
}

func getAnomalyTrendsHandler(svc *anomalydetection.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"trends": []interface{}{},
		})
	}
}
