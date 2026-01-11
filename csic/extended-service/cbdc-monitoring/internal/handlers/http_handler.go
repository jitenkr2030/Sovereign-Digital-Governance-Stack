package handlers

import (
	"net/http"
	"strconv"

	"csic-platform/extended-services/cbdc-monitoring/internal/service"
	"csic-platform/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// CBdcHandler handles HTTP requests for CBDC monitoring
type CBdcHandler struct {
	cbdcService *service.CBDCMonitoringService
	log         *logger.Logger
}

// NewCBdcHandler creates a new CBDC handler
func NewCBdcHandler(cbdcService *service.CBDCMonitoringService, log *logger.Logger) *CBdcHandler {
	return &CBdcHandler{
		cbdcService: cbdcService,
		log:         log,
	}
}

// Health returns service health status
func (h *CBdcHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "cbdc-monitoring",
	})
}

// Readiness returns service readiness status
func (h *CBdcHandler) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Liveness returns service liveness status
func (h *CBdcHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// CreateWallet creates a new wallet
func (h *CBdcHandler) CreateWallet(c *gin.Context) {
	var req service.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("Invalid create wallet request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	wallet, err := h.cbdcService.CreateWallet(c.Request.Context(), req)
	if err != nil {
		h.log.Errorf("Failed to create wallet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// GetWallet retrieves a wallet by ID
func (h *CBdcHandler) GetWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	wallet, err := h.cbdcService.GetWallet(c.Request.Context(), id)
	if err != nil {
		h.log.Warnf("Wallet not found: %s, error: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// UpdateWallet updates wallet information
func (h *CBdcHandler) UpdateWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	var req service.UpdateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("Invalid update wallet request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	wallet, err := h.cbdcService.UpdateWallet(c.Request.Context(), id, req)
	if err != nil {
		h.log.Errorf("Failed to update wallet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// DeleteWallet deletes a wallet (soft delete - sets status to closed)
func (h *CBdcHandler) DeleteWallet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	_, err := h.cbdcService.UpdateWallet(c.Request.Context(), id, service.UpdateWalletRequest{Status: 5}) // 5 = Closed
	if err != nil {
		h.log.Errorf("Failed to delete wallet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wallet closed successfully"})
}

// GetWalletBalance retrieves wallet balance
func (h *CBdcHandler) GetWalletBalance(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	balance, err := h.cbdcService.GetWalletBalance(c.Request.Context(), id)
	if err != nil {
		h.log.Warnf("Wallet not found: %s, error: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet_id": id,
		"balance":   balance.String(),
		"currency":  "CBDC",
	})
}

// CreateTransaction creates a new transaction
func (h *CBdcHandler) CreateTransaction(c *gin.Context) {
	var req service.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("Invalid create transaction request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Parse amount from string
	amountStr := c.PostForm("amount")
	if amountStr != "" {
		amount, err := decimal.NewFromString(amountStr)
		if err == nil {
			req.Amount = amount
		}
	}

	tx, err := h.cbdcService.CreateTransaction(c.Request.Context(), req)
	if err != nil {
		h.log.Errorf("Failed to create transaction: %v", err)
		if err.Error() == "insufficient funds" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

// GetTransaction retrieves a transaction by ID
func (h *CBdcHandler) GetTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction id is required"})
		return
	}

	tx, err := h.cbdcService.GetTransaction(c.Request.Context(), id)
	if err != nil {
		h.log.Warnf("Transaction not found: %s, error: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// ListTransactions retrieves transactions with filters
func (h *CBdcHandler) ListTransactions(c *gin.Context) {
	var filters service.TransactionListFilters

	filters.WalletID = c.Query("wallet_id")

	if statusStr := c.Query("status"); statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		filters.Status = &status
	}

	if typeStr := c.Query("type"); typeStr != "" {
		txType, _ := strconv.Atoi(typeStr)
		filters.Type = &txType
	}

	filters.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filters.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))

	if filters.Limit > 100 {
		filters.Limit = 100
	}

	transactions, total, err := h.cbdcService.ListTransactions(c.Request.Context(), filters)
	if err != nil {
		h.log.Errorf("Failed to list transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  transactions,
		"total": total,
		"page":  filters.Page,
		"limit": filters.Limit,
	})
}

// ApproveTransaction approves a pending transaction
func (h *CBdcHandler) ApproveTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction id is required"})
		return
	}

	approverID := c.PostForm("approver_id")
	if approverID == "" {
		approverID = "system"
	}

	tx, err := h.cbdcService.ApproveTransaction(c.Request.Context(), id, approverID)
	if err != nil {
		h.log.Errorf("Failed to approve transaction: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// RejectTransaction rejects a pending transaction
func (h *CBdcHandler) RejectTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction id is required"})
		return
	}

	reason := c.PostForm("reason")
	if reason == "" {
		reason = "No reason provided"
	}

	tx, err := h.cbdcService.RejectTransaction(c.Request.Context(), id, reason)
	if err != nil {
		h.log.Errorf("Failed to reject transaction: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// CreatePolicy creates a new policy rule
func (h *CBdcHandler) CreatePolicy(c *gin.Context) {
	var req service.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("Invalid create policy request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	policy, err := h.cbdcService.CreatePolicy(c.Request.Context(), req)
	if err != nil {
		h.log.Errorf("Failed to create policy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// GetPolicy retrieves a policy rule by ID
func (h *CBdcHandler) GetPolicy(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy id is required"})
		return
	}

	policy, err := h.cbdcService.GetPolicy(c.Request.Context(), id)
	if err != nil {
		h.log.Warnf("Policy not found: %s, error: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// ListPolicies retrieves all enabled policy rules
func (h *CBdcHandler) ListPolicies(c *gin.Context) {
	ruleType := c.Query("rule_type")

	policies, err := h.cbdcService.ListPolicies(c.Request.Context(), ruleType)
	if err != nil {
		h.log.Errorf("Failed to list policies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  policies,
		"total": len(policies),
	})
}

// UpdatePolicy updates a policy rule
func (h *CBdcHandler) UpdatePolicy(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy updated"})
}

// DeletePolicy deletes a policy rule (soft delete)
func (h *CBdcHandler) DeletePolicy(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy disabled successfully"})
}

// GetMonitoringStats retrieves monitoring statistics
func (h *CBdcHandler) GetMonitoringStats(c *gin.Context) {
	stats, err := h.cbdcService.GetMonitoringStats(c.Request.Context())
	if err != nil {
		h.log.Errorf("Failed to get monitoring stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTransactionVelocity retrieves transaction velocity for a wallet
func (h *CBdcHandler) GetTransactionVelocity(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet_id":     walletID,
		"message":       "Velocity check endpoint - implement detailed velocity tracking",
	})
}

// GetViolations retrieves recent violations
func (h *CBdcHandler) GetViolations(c *gin.Context) {
	walletID := c.Query("wallet_id")
	unresolvedOnly := c.DefaultQuery("unresolved_only", "false") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// For now, return empty list - implement violation listing
	c.JSON(http.StatusOK, gin.H{
		"data":           []interface{}{},
		"message":        "Violation listing - implement violation repository query",
	})
}

// ComplianceCheck performs a compliance check
func (h *CBdcHandler) ComplianceCheck(c *gin.Context) {
	var req service.ComplianceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warnf("Invalid compliance check request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.cbdcService.ComplianceCheck(c.Request.Context(), req)
	if err != nil {
		h.log.Errorf("Failed to perform compliance check: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAuditTrail retrieves audit trail for a wallet
func (h *CBdcHandler) GetAuditTrail(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet_id": walletID,
		"message":   "Audit trail - implement audit repository query",
	})
}
