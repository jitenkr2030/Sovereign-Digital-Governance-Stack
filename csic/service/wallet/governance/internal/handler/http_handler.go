package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/csic/wallet-governance/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HTTPHandler handles HTTP requests for wallet governance
type HTTPHandler struct {
	walletSvc      *service.WalletService
	signatureSvc   *service.SignatureService
	governanceSvc  *service.GovernanceService
	freezeSvc      *service.FreezeService
	complianceSvc  *service.ComplianceService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(
	walletSvc *service.WalletService,
	signatureSvc *service.SignatureService,
	governanceSvc *service.GovernanceService,
	freezeSvc *service.FreezeService,
	complianceSvc *service.ComplianceService,
) *HTTPHandler {
	return &HTTPHandler{
		walletSvc:      walletSvc,
		signatureSvc:   signatureSvc,
		governanceSvc:  governanceSvc,
		freezeSvc:      freezeSvc,
		complianceSvc:  complianceSvc,
	}
}

// Wallet handlers

// RegisterWallet registers a new wallet
func (h *HTTPHandler) RegisterWallet(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.RegisterWallet(c.Request.Context(), &wallet, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// GetWallet retrieves a wallet by ID
func (h *HTTPHandler) GetWallet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	wallet, err := h.walletSvc.GetWallet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// UpdateWallet updates a wallet
func (h *HTTPHandler) UpdateWallet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wallet.ID = id

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.UpdateWallet(c.Request.Context(), &wallet, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// RevokeWallet revokes a wallet
func (h *HTTPHandler) RevokeWallet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.RevokeWallet(c.Request.Context(), id, req.Reason, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wallet revoked"})
}

// ListWallets lists wallets with filters
func (h *HTTPHandler) ListWallets(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := &models.WalletFilter{}
	c.ShouldBindQuery(filter)

	wallets, count, err := h.walletSvc.ListWallets(c.Request.Context(), filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallets": wallets,
		"total":   count,
		"limit":   limit,
		"offset":  offset,
	})
}

// RegisterCustodialWallet registers a custodial wallet
func (h *HTTPHandler) RegisterCustodialWallet(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.RegisterCustodialWallet(c.Request.Context(), &wallet, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// RegisterMultiSigWallet registers a multi-signature wallet
func (h *HTTPHandler) RegisterMultiSigWallet(c *gin.Context) {
	var req struct {
		Wallet  models.Wallet        `json:"wallet"`
		Signers []*models.WalletSigner `json:"signers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.RegisterMultiSigWallet(c.Request.Context(), &req.Wallet, req.Signers, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req.Wallet)
}

// RegisterExchangeHotWallet registers an exchange hot wallet
func (h *HTTPHandler) RegisterExchangeHotWallet(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.RegisterExchangeHotWallet(c.Request.Context(), &wallet, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// RegisterExchangeColdWallet registers an exchange cold wallet
func (h *HTTPHandler) RegisterExchangeColdWallet(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.walletSvc.RegisterExchangeColdWallet(c.Request.Context(), &wallet, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// Compliance handlers

// GetWalletComplianceStatus retrieves compliance status for a wallet
func (h *HTTPHandler) GetWalletComplianceStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	check, err := h.walletSvc.CheckWalletCompliance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, check)
}

// CheckWalletCompliance performs compliance check for a wallet
func (h *HTTPHandler) CheckWalletCompliance(c *gin.Context) {
	var req struct {
		WalletID uuid.UUID `json:"wallet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	check, err := h.complianceSvc.CheckWalletCompliance(c.Request.Context(), req.WalletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, check)
}

// GetWalletAuditTrail retrieves audit trail for a wallet
func (h *HTTPHandler) GetWalletAuditTrail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.complianceSvc.GetWalletAuditTrail(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// Freeze handlers

// FreezeWallet freezes a wallet
func (h *HTTPHandler) FreezeWallet(c *gin.Context) {
	var freeze models.WalletFreeze
	if err := c.ShouldBindJSON(&freeze); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.freezeSvc.FreezeWallet(c.Request.Context(), &freeze, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, freeze)
}

// EmergencyFreeze performs emergency freeze
func (h *HTTPHandler) EmergencyFreeze(c *gin.Context) {
	var req struct {
		WalletID      uuid.UUID        `json:"wallet_id" binding:"required"`
		Reason        models.FreezeReason `json:"reason" binding:"required"`
		ReasonDetails string           `json:"reason_details"`
		LegalOrderID  string           `json:"legal_order_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.freezeSvc.EmergencyFreeze(c.Request.Context(), req.WalletID, req.Reason, req.ReasonDetails, req.LegalOrderID, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "emergency freeze executed"})
}

// UnfreezeWallet unfreezes a wallet
func (h *HTTPHandler) UnfreezeWallet(c *gin.Context) {
	var req struct {
		WalletID uuid.UUID `json:"wallet_id" binding:"required"`
		Reason   string    `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.freezeSvc.UnfreezeWallet(c.Request.Context(), req.WalletID, req.Reason, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wallet unfrozen"})
}

// GetFreezeStatus retrieves freeze status for a wallet
func (h *HTTPHandler) GetFreezeStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	freeze, err := h.freezeSvc.GetFreezeStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if freeze == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active freeze found"})
		return
	}

	c.JSON(http.StatusOK, freeze)
}

// GetActiveFreezes retrieves all active freezes
func (h *HTTPHandler) GetActiveFreezes(c *gin.Context) {
	freezes, err := h.freezeSvc.GetActiveFreezes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, freezes)
}

// GetFreezeHistory retrieves freeze history for a wallet
func (h *HTTPHandler) GetFreezeHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("wallet_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	freezes, err := h.freezeSvc.GetFreezeHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, freezes)
}

// Blacklist handlers

// AddToBlacklist adds an address to the blacklist
func (h *HTTPHandler) AddToBlacklist(c *gin.Context) {
	var entry models.BlacklistEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.complianceSvc.AddToBlacklist(c.Request.Context(), &entry, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// RemoveFromBlacklist removes an address from the blacklist
func (h *HTTPHandler) RemoveFromBlacklist(c *gin.Context) {
	address := c.Param("address")
	blockchainStr := c.Query("blockchain")
	blockchain := models.BlockchainType(blockchainStr)

	var req struct {
		RemovedByName string `json:"removed_by_name"`
		Reason        string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	removedBy := getUserID(c)

	if err := h.complianceSvc.RemoveFromBlacklist(c.Request.Context(), address, blockchain, removedBy, req.RemovedByName, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "address removed from blacklist"})
}

// GetBlacklist retrieves the blacklist
func (h *HTTPHandler) GetBlacklist(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entries, count, err := h.complianceSvc.GetBlacklist(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"total":   count,
	})
}

// CheckBlacklist checks if an address is blacklisted
func (h *HTTPHandler) CheckBlacklist(c *gin.Context) {
	address := c.Param("address")
	blockchainStr := c.Query("blockchain")
	blockchain := models.BlockchainType(blockchainStr)

	blacklisted, err := h.complianceSvc.IsBlacklisted(c.Request.Context(), address, blockchain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":     address,
		"blockchain":  blockchain,
		"blacklisted": blacklisted,
	})
}

// BatchAddToBlacklist adds multiple addresses to the blacklist
func (h *HTTPHandler) BatchAddToBlacklist(c *gin.Context) {
	var entries []*models.BlacklistEntry
	if err := c.ShouldBindJSON(&entries); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, entry := range entries {
		entry.Status = models.BlacklistStatusActive
	}

	// Would implement batch add in service
	c.JSON(http.StatusCreated, gin.H{"message": "batch add initiated"})
}

// Whitelist handlers (similar to blacklist)

// GetWhitelist retrieves the whitelist
func (h *HTTPHandler) GetWhitelist(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entries, count, err := h.complianceSvc.GetWhitelist(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"total":   count,
	})
}

// AddToWhitelist adds an address to the whitelist
func (h *HTTPHandler) AddToWhitelist(c *gin.Context) {
	var entry models.WhitelistEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.Status = "ACTIVE"
	entry.AddedBy = getUserID(c)
	entry.AddedByName = getUserName(c)

	// Would implement in service
	c.JSON(http.StatusCreated, entry)
}

// RemoveFromWhitelist removes an address from the whitelist
func (h *HTTPHandler) RemoveFromWhitelist(c *gin.Context) {
	address := c.Param("address")
	c.JSON(http.StatusOK, gin.H{"message": "whitelist removal not implemented"})
}

// CheckWhitelist checks if an address is whitelisted
func (h *HTTPHandler) CheckWhitelist(c *gin.Context) {
	address := c.Param("address")
	blockchainStr := c.Query("blockchain")
	blockchain := models.BlockchainType(blockchainStr)

	whitelisted, err := h.complianceSvc.IsWhitelisted(c.Request.Context(), address, blockchain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":     address,
		"blockchain":  blockchain,
		"whitelisted": whitelisted,
	})
}

// BatchAddToWhitelist adds multiple addresses to the whitelist
func (h *HTTPHandler) BatchAddToWhitelist(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "batch whitelist add not implemented"})
}

// Governance handlers

// GetWalletSigners retrieves signers for a wallet
func (h *HTTPHandler) GetWalletSigners(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	signers, err := h.governanceSvc.GetWalletSigners(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, signers)
}

// AddSigner adds a signer to a wallet
func (h *HTTPHandler) AddSigner(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	var signer models.WalletSigner
	if err := c.ShouldBindJSON(&signer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.governanceSvc.AddSigner(c.Request.Context(), id, &signer, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, signer)
}

// RemoveSigner removes a signer from a wallet
func (h *HTTPHandler) RemoveSigner(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	signerID, err := uuid.Parse(c.Param("signer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signer ID"})
		return
	}

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.governanceSvc.RemoveSigner(c.Request.Context(), walletID, signerID, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "signer removed"})
}

// UpdateSignerThreshold updates the threshold for a wallet
func (h *HTTPHandler) UpdateSignerThreshold(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	signerID, err := uuid.Parse(c.Param("signer_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signer ID"})
		return
	}

	var req struct {
		Threshold int `json:"threshold" binding:"required,min=1"`
	}
	c.ShouldBindJSON(&req)

	actorID := getUserID(c)
	actorName := getUserName(c)

	if err := h.governanceSvc.UpdateSignerThreshold(c.Request.Context(), walletID, req.Threshold, actorID, actorName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "threshold updated"})
}

// ProposeTransaction proposes a transaction
func (h *HTTPHandler) ProposeTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	var proposal models.TransactionProposal
	if err := c.ShouldBindJSON(&proposal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	proposal.WalletID = id

	proposerID := getUserID(c)
	proposerName := getUserName(c)

	result, err := h.governanceSvc.ProposeTransaction(c.Request.Context(), &proposal, proposerID, proposerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPendingTransactions retrieves pending transactions for a wallet
func (h *HTTPHandler) GetPendingTransactions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	transactions, err := h.governanceSvc.GetPendingTransactions(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// ApproveTransaction approves a transaction
func (h *HTTPHandler) ApproveTransaction(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	txID, err := uuid.Parse(c.Param("tx_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var req struct {
		SignerID  uuid.UUID `json:"signer_id" binding:"required"`
		SignerName string   `json:"signer_name" binding:"required"`
		Signature string   `json:"signature" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	result, err := h.governanceSvc.ApproveTransaction(c.Request.Context(), txID, req.SignerID, req.SignerName, req.Signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RejectTransaction rejects a transaction
func (h *HTTPHandler) RejectTransaction(c *gin.Context) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet ID"})
		return
	}

	txID, err := uuid.Parse(c.Param("tx_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	var req struct {
		SignerID   uuid.UUID `json:"signer_id" binding:"required"`
		SignerName string    `json:"signer_name" binding:"required"`
		Reason     string    `json:"reason" binding:"required"`
	}
	c.ShouldBindJSON(&req)

	if err := h.governanceSvc.RejectTransaction(c.Request.Context(), txID, req.SignerID, req.SignerName, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transaction rejected"})
}

// Signature handlers

// RequestSignature requests a signature
func (h *HTTPHandler) RequestSignature(c *gin.Context) {
	var req struct {
		WalletID      uuid.UUID `json:"wallet_id" binding:"required"`
		SignerID      uuid.UUID `json:"signer_id" binding:"required"`
		Message       string    `json:"message" binding:"required"`
		SignatureType string    `json:"signature_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.signatureSvc.RequestSignature(c.Request.Context(), req.WalletID, req.SignerID, req.Message, req.SignatureType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetSignatureStatus retrieves signature status
func (h *HTTPHandler) GetSignatureStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature ID"})
		return
	}

	req, err := h.signatureSvc.GetSignatureStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, req)
}

// VerifySignature verifies a signature
func (h *HTTPHandler) VerifySignature(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature ID"})
		return
	}

	req, err := h.signatureSvc.GetSignatureStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.signatureSvc.VerifySignature(c.Request.Context(), req.MessageHash, req.Signature, req.PublicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      valid,
		"message":    req.Message,
		"public_key": req.PublicKey,
	})
}

// Recovery handlers (stubs)

// RequestAssetRecovery requests asset recovery
func (h *HTTPHandler) RequestAssetRecovery(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "asset recovery not implemented"})
}

// GetRecoveryRequest retrieves recovery request
func (h *HTTPHandler) GetRecoveryRequest(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "recovery not implemented"})
}

// ApproveRecovery approves a recovery request
func (h *HTTPHandler) ApproveRecovery(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "recovery not implemented"})
}

// ExecuteRecovery executes a recovery request
func (h *HTTPHandler) ExecuteRecovery(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "recovery not implemented"})
}

// Reporting handlers

// GetWalletSummary retrieves wallet statistics summary
func (h *HTTPHandler) GetWalletSummary(c *gin.Context) {
	summary, err := h.walletSvc.GetWalletSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetWalletsByExchange retrieves wallets grouped by exchange
func (h *HTTPHandler) GetWalletsByExchange(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "exchange grouping not implemented"})
}

// GetWalletsByType retrieves wallets grouped by type
func (h *HTTPHandler) GetWalletsByType(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "type grouping not implemented"})
}

// GetFreezeSummary retrieves freeze statistics
func (h *HTTPHandler) GetFreezeSummary(c *gin.Context) {
	freezes, err := h.freezeSvc.GetActiveFreezes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_freezes": len(freezes),
		"freezes":        freezes,
	})
}

// Helper functions

func getUserID(c *gin.Context) uuid.UUID {
	// In production, this would extract the user ID from JWT or session
	return uuid.Nil
}

func getUserName(c *gin.Context) string {
	// In production, this would extract the user name from JWT or session
	return "system"
}
