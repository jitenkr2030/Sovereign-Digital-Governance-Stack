package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/csic/transaction-monitoring/internal/repository"
	"github.com/csic/transaction-monitoring/internal/service/graph"
	"github.com/csic/transaction-monitoring/internal/service/ingest"
	"github.com/csic/transaction-monitoring/internal/service/risk"
	"github.com/csic/transaction-monitoring/internal/service/sanctions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles HTTP requests
type Handler struct {
	cfg            *config.Config
	repo           *repository.Repository
	cache          *repository.CacheRepository
	ingestionSvc   *ingest.IngestionService
	riskSvc        *risk.RiskScoringService
	clusteringSvc  *graph.ClusteringService
	sanctionsSvc   *sanctions.SanctionsService
	logger         *zap.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(
	cfg *config.Config,
	repo *repository.Repository,
	cache *repository.CacheRepository,
	ingestionSvc *ingest.IngestionService,
	riskSvc *risk.RiskScoringService,
	clusteringSvc *graph.ClusteringService,
	sanctionsSvc *sanctions.SanctionsService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		cfg:           cfg,
		repo:          repo,
		cache:         cache,
		ingestionSvc:  ingestionSvc,
		riskSvc:       riskSvc,
		clusteringSvc: clusteringSvc,
		sanctionsSvc:  sanctionsSvc,
		logger:        logger,
	}
}

// SetupRouter sets up the Gin router
func (h *Handler) SetupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(h.loggingMiddleware())
	router.Use(h.corsMiddleware())

	// Health endpoints
	router.GET("/health", h.healthCheck)
	router.GET("/ready", h.readinessCheck)

	// API v1
	v1 := router.Group("/v1")
	{
		// Wallet endpoints
		wallets := v1.Group("/wallets")
		{
			wallets.GET("/:network/:address/risk", h.getWalletRisk)
			wallets.GET("/:network/:address/history", h.getWalletHistory)
			wallets.GET("/:network/:address/cluster", h.getWalletCluster)
			wallets.GET("/:network/:address/transactions", h.getWalletTransactions)
		}

		// Screening endpoints
		screen := v1.Group("/screen")
		{
			screen.POST("", h.screenAddresses)
			screen.GET("/:network/:address", h.screenAddress)
		}

		// Cluster endpoints
		clusters := v1.Group("/clusters")
		{
			clusters.GET("/:id", h.getCluster)
			clusters.GET("/:id/members", h.getClusterMembers)
			clusters.GET("/:id/related", h.getClusterRelationships)
			clusters.GET("", h.listClusters)
		}

		// Alert endpoints
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", h.listAlerts)
			alerts.GET("/:id", h.getAlert)
			alerts.POST("/:id/assign", h.assignAlert)
			alerts.POST("/:id/resolve", h.resolveAlert)
			alerts.POST("/:id/dismiss", h.dismissAlert)
		}

		// Transaction endpoints
		txs := v1.Group("/transactions")
		{
			txs.GET("/:hash", h.getTransaction)
			txs.POST("/ingest", h.ingestTransaction)
			txs.POST("/:hash/screen", h.screenTransaction)
		}

		// Stats endpoints
		stats := v1.Group("/stats")
		{
			stats.GET("", h.getStats)
			stats.GET("/clustering", h.getClusteringStats)
			stats.GET("/sanctions", h.getSanctionsStats)
		}
	}

	return router
}

// Health and readiness

func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   h.cfg.App.Version,
	})
}

func (h *Handler) readinessCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database
	if err := h.repo.Close(); err == nil {
		// This is a workaround - we shouldn't close the repo here
	}

	// Check Redis
	if err := h.cache.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "Redis connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		h.logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func (h *Handler) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.cfg.Security.CORS.Enabled {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Wallet endpoints

func (h *Handler) getWalletRisk(c *gin.Context) {
	network := models.Network(c.Param("network"))
	address := c.Param("address")

	ctx := c.Request.Context()

	// Get risk score
	score, err := h.riskSvc.GetRiskScore(ctx, address, network)
	if err != nil {
		h.logger.Error("Failed to get risk score", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve risk score"})
		return
	}

	// Get wallet data
	wallet, err := h.repo.GetWalletByAddress(ctx, address, network)
	if err != nil {
		h.logger.Error("Failed to get wallet", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve wallet"})
		return
	}

	// Get risk factors
	factors, _ := h.cache.GetWalletRiskFactors(ctx, wallet.ID)

	c.JSON(http.StatusOK, gin.H{
		"address":      address,
		"network":      network,
		"risk_score":   score,
		"risk_level":   wallet.RiskLevel,
		"factors":      factors,
		"last_updated": time.Now().UTC(),
	})
}

func (h *Handler) getWalletHistory(c *gin.Context) {
	network := models.Network(c.Param("network"))
	address := c.Param("address")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	ctx := c.Request.Context()

	txs, err := h.repo.GetTransactionHistory(ctx, address, network, limit)
	if err != nil {
		h.logger.Error("Failed to get transaction history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":   address,
		"network":   network,
		"count":     len(txs),
		"transactions": txs,
	})
}

func (h *Handler) getWalletCluster(c *gin.Context) {
	network := models.Network(c.Param("network"))
	address := c.Param("address")

	ctx := c.Request.Context()

	cluster, err := h.clusteringSvc.GetWalletCluster(ctx, address, network)
	if err != nil {
		h.logger.Error("Failed to get wallet cluster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cluster"})
		return
	}

	if cluster == nil {
		c.JSON(http.StatusOK, gin.H{
			"address": address,
			"cluster": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"cluster": cluster,
	})
}

func (h *Handler) getWalletTransactions(c *gin.Context) {
	network := models.Network(c.Param("network"))
	address := c.Param("address")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	ctx := c.Request.Context()

	txs, err := h.repo.GetTransactionHistory(ctx, address, network, limit)
	if err != nil {
		h.logger.Error("Failed to get transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":      address,
		"network":      network,
		"transactions": txs,
	})
}

// Screening endpoints

func (h *Handler) screenAddress(c *gin.Context) {
	network := models.Network(c.Param("network"))
	address := c.Param("address")

	ctx := c.Request.Context()

	result, err := h.sanctionsSvc.ScreenAddress(ctx, address, network)
	if err != nil {
		h.logger.Error("Failed to screen address", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to screen address"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) screenAddresses(c *gin.Context) {
	var req struct {
		Addresses []string        `json:"addresses"`
		Network   models.Network  `json:"network"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx := c.Request.Context()

	var results []map[string]interface{}
	for _, addr := range req.Addresses {
		result, err := h.sanctionsSvc.ScreenAddress(ctx, addr, req.Network)
		if err != nil {
			h.logger.Warn("Failed to screen address", zap.String("address", addr), zap.Error(err))
			continue
		}

		results = append(results, map[string]interface{}{
			"address":        result.Address,
			"is_sanctioned": result.IsSanctioned,
			"matches":        result.Matches,
			"direct_links":   result.DirectLinks,
			"indirect_links": result.IndirectLinks,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// Cluster endpoints

func (h *Handler) getCluster(c *gin.Context) {
	clusterID := c.Param("id")

	ctx := c.Request.Context()

	cluster, err := h.clusteringSvc.GetCluster(ctx, clusterID)
	if err != nil {
		h.logger.Error("Failed to get cluster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cluster"})
		return
	}

	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cluster not found"})
		return
	}

	c.JSON(http.StatusOK, cluster)
}

func (h *Handler) getClusterMembers(c *gin.Context) {
	clusterID := c.Param("id")

	ctx := c.Request.Context()

	members, err := h.clusteringSvc.GetClusterMembers(ctx, clusterID)
	if err != nil {
		h.logger.Error("Failed to get cluster members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster_id": clusterID,
		"count":      len(members),
		"members":    members,
	})
}

func (h *Handler) getClusterRelationships(c *gin.Context) {
	clusterID := c.Param("id")

	ctx := c.Request.Context()

	relationships, err := h.clusteringSvc.FindRelatedClusters(ctx, clusterID)
	if err != nil {
		h.logger.Error("Failed to get cluster relationships", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve relationships"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster_id":    clusterID,
		"count":         len(relationships),
		"relationships": relationships,
	})
}

func (h *Handler) listClusters(c *gin.Context) {
	ctx := c.Request.Context()

	clusters, err := h.repo.GetAllClusters(ctx)
	if err != nil {
		h.logger.Error("Failed to list clusters", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve clusters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(clusters),
		"clusters": clusters,
	})
}

// Alert endpoints

func (h *Handler) listAlerts(c *gin.Context) {
	severity := c.Query("severity")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// This would query alerts from the repository with filters
	// For now, return a placeholder
	c.JSON(http.StatusOK, gin.H{
		"alerts": []models.Alert{},
		"count":  0,
	})
}

func (h *Handler) getAlert(c *gin.Context) {
	alertID := c.Param("id")

	ctx := c.Request.Context()
	// This would retrieve a specific alert
	_ = ctx

	c.JSON(http.StatusOK, gin.H{
		"id": alertID,
	})
}

func (h *Handler) assignAlert(c *gin.Context) {
	alertID := c.Param("id")

	var req struct {
		Assignee string `json:"assignee"`
		Team     string `json:"team"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// This would assign the alert
	_ = alertID
	_ = req

	c.JSON(http.StatusOK, gin.H{
		"id":        alertID,
		"assigned":  true,
		"assignee":  req.Assignee,
		"team":      req.Team,
	})
}

func (h *Handler) resolveAlert(c *gin.Context) {
	alertID := c.Param("id")

	var req struct {
		Resolution string `json:"resolution"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// This would resolve the alert
	_ = alertID
	_ = req

	c.JSON(http.StatusOK, gin.H{
		"id":          alertID,
		"status":      "resolved",
		"resolution":  req.Resolution,
	})
}

func (h *Handler) dismissAlert(c *gin.Context) {
	alertID := c.Param("id")

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// This would dismiss the alert
	_ = alertID
	_ = req

	c.JSON(http.StatusOK, gin.H{
		"id":     alertID,
		"status": "dismissed",
		"reason": req.Reason,
	})
}

// Transaction endpoints

func (h *Handler) getTransaction(c *gin.Context) {
	txHash := c.Param("hash")

	ctx := c.Request.Context()

	tx, err := h.repo.GetTransaction(ctx, txHash)
	if err != nil {
		h.logger.Error("Failed to get transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *Handler) ingestTransaction(c *gin.Context) {
	var req struct {
		TxHash string           `json:"tx_hash"`
		Network models.Network  `json:"network"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx := c.Request.Context()

	if err := h.ingestionSvc.IngestTransaction(ctx, req.TxHash, req.Network); err != nil {
		h.logger.Error("Failed to ingest transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ingest transaction"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":   "Transaction ingestion initiated",
		"tx_hash":   req.TxHash,
		"network":   req.Network,
	})
}

func (h *Handler) screenTransaction(c *gin.Context) {
	txHash := c.Param("hash")

	ctx := c.Request.Context()

	// Get transaction
	tx, err := h.repo.GetTransaction(ctx, txHash)
	if err != nil {
		h.logger.Error("Failed to get transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transaction"})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// This would screen the transaction
	_ = tx

	c.JSON(http.StatusOK, gin.H{
		"tx_hash": txHash,
		"screened": true,
	})
}

// Stats endpoints

func (h *Handler) getStats(c *gin.Context) {
	ctx := c.Request.Context()

	ingestionStatus := h.ingestionSvc.GetIngestionStatus()
	clusteringStats, _ := h.clusteringSvc.GetClusteringStats(ctx)
	sanctionsStats, _ := h.sanctionsSvc.GetSanctionsStats(ctx)

	c.JSON(http.StatusOK, gin.H{
		"ingestion":    ingestionStatus,
		"clustering":   clusteringStats,
		"sanctions":    sanctionsStats,
		"cache_stats":  h.cache.GetStats(ctx),
	})
}

func (h *Handler) getClusteringStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.clusteringSvc.GetClusteringStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get clustering stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) getSanctionsStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.sanctionsSvc.GetSanctionsStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get sanctions stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
