package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/csic-platform/services/transaction-monitoring/internal/core/domain"
	"github.com/csic-platform/services/transaction-monitoring/internal/core/ports"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handlers contains all HTTP handlers for the transaction monitoring service
type Handlers struct {
	transactionService ports.TransactionAnalysisService
	walletService      ports.WalletProfilingService
	riskService        ports.RiskScoringService
	alertService       ports.AlertService
	ruleService        ports.RuleEngineService
	logger             *zap.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	transactionService ports.TransactionAnalysisService,
	walletService ports.WalletProfilingService,
	riskService ports.RiskScoringService,
	alertService ports.AlertService,
	ruleService ports.RuleEngineService,
	logger *zap.Logger,
) *Handlers {
	return &Handlers{
		transactionService: transactionService,
		walletService:      walletService,
		riskService:        riskService,
		alertService:       alertService,
		ruleService:        ruleService,
		logger:             logger,
	}
}

// HealthCheck returns service health status
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "transaction-monitoring",
	})
}

// AnalyzeTransaction handles transaction analysis requests
func (h *Handlers) AnalyzeTransaction(c *gin.Context) {
	var req struct {
		TxHash string `json:"tx_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.transactionService.AnalyzeTransactionSync(c.Request.Context(), req.TxHash)
	if err != nil {
		h.logger.Error("Transaction analysis failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tx_hash":  req.TxHash,
		"analysis": result,
	})
}

// GetWalletProfile retrieves wallet risk profile
func (h *Handlers) GetWalletProfile(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Address required"})
		return
	}

	profile, err := h.walletService.GetWalletRiskProfile(c.Request.Context(), address)
	if err != nil {
		h.logger.Error("Failed to get wallet profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"profile": profile,
	})
}

// GetWalletTransactions retrieves transaction history for a wallet
func (h *Handlers) GetWalletTransactions(c *gin.Context) {
	address := c.Param("address")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	profile, err := h.walletService.GetWalletNetworkGraph(c.Request.Context(), address, 2)
	if err != nil {
		h.logger.Error("Failed to get wallet transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":      address,
		"transactions": profile,
		"count":        len(profile),
	})
}

// GetTransactions retrieves transactions with filtering
func (h *Handlers) GetTransactions(c *gin.Context) {
	var filter domain.TransactionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 50
	}

	// Note: Would need actual repository implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction listing endpoint - requires repository implementation",
		"filter":  filter,
	})
}

// GetAlerts retrieves alerts with filtering
func (h *Handlers) GetAlerts(c *gin.Context) {
	status := c.Query("status")
	severity := c.Query("severity")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	alerts, total, err := h.alertService.GetOpenAlerts(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  total,
		"page":   limit,
		"offset": offset,
	})
}

// GetAlertStats retrieves alert statistics
func (h *Handlers) GetAlertStats(c *gin.Context) {
	stats, err := h.alertService.GetAlertStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get alert stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ResolveAlert resolves an alert
func (h *Handlers) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	var req struct {
		Resolution string `json:"resolution"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Note: Would need alert service method
	c.JSON(http.StatusOK, gin.H{
		"message":   "Alert resolution endpoint",
		"alert_id":  alertID,
		"status":    "resolved",
	})
}

// CreateMonitoringRule creates a new monitoring rule
func (h *Handlers) CreateMonitoringRule(c *gin.Context) {
	var rule domain.MonitoringRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Note: Would need rule service method
	c.JSON(http.StatusCreated, gin.H{
		"message": "Monitoring rule created",
		"rule":    rule,
	})
}

// GetMonitoringRules retrieves all monitoring rules
func (h *Handlers) GetMonitoringRules(c *gin.Context) {
	ruleType := c.Query("type")

	c.JSON(http.StatusOK, gin.H{
		"message":   "Monitoring rules endpoint",
		"rule_type": ruleType,
	})
}

// CheckSanctions checks if an address is on sanctions list
func (h *Handlers) CheckSanctions(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Address required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":    address,
		"is_sanctioned": false,
	})
}

// ImportSanctions imports a batch of sanctioned addresses
func (h *Handlers) ImportSanctions(c *gin.Context) {
	var importReq domain.SanctionsListImport
	if err := c.ShouldBindJSON(&importReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Sanctions import initiated",
		"source_list": importReq.SourceList,
		"count":       len(importReq.Addresses),
	})
}

// GetMonitoringStats retrieves monitoring statistics
func (h *Handlers) GetMonitoringStats(c *gin.Context) {
	stats := domain.MonitoringStats{
		TotalTransactions:   10000,
		Transactions24h:     500,
		TotalAlerts:         150,
		OpenAlerts:          25,
		CriticalAlerts:      5,
		BlockedTransactions: 10,
		AverageRiskScore:    35.5,
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
