package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/csic/platform/blockchain/indexer/internal/config"
	"github.com/csic/platform/blockchain/indexer/internal/domain"
	"github.com/csic/platform/blockchain/indexer/internal/indexer"
	"github.com/csic/platform/blockchain/indexer/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// QueryService handles blockchain data queries
type QueryService struct {
	repo   *repository.PostgresRepository
	redis  *repository.RedisClient
	config *config.Config
	logger *zap.Logger
}

// NewQueryService creates a new query service
func NewQueryService(
	repo *repository.PostgresRepository,
	redis *repository.RedisClient,
	cfg *config.Config,
	logger *zap.Logger,
) *QueryService {
	return &QueryService{
		repo:   repo,
		redis:  redis,
		config: cfg,
		logger: logger,
	}
}

// HTTPHandler handles HTTP requests
type HTTPHandler struct {
	queryService *QueryService
	logger       *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(queryService *QueryService, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		queryService: queryService,
		logger:       logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.HealthCheck)
	router.GET("/ready", h.ReadinessCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Block endpoints
		blocks := v1.Group("/blocks")
		{
			blocks.GET("", h.ListBlocks)
			blocks.GET("/:hash_or_number", h.GetBlock)
			blocks.GET("/latest", h.GetLatestBlock)
		}

		// Transaction endpoints
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", h.ListTransactions)
			transactions.GET("/:hash", h.GetTransaction)
			transactions.GET("/address/:address", h.GetTransactionsByAddress)
		}

		// Address endpoints
		addresses := v1.Group("/addresses")
		{
			addresses.GET("/:address", h.GetAddress)
			addresses.GET("/:address/logs", h.GetAddressLogs)
		}

		// Token endpoints
		tokens := v1.Group("/tokens")
		{
			tokens.GET("/transfers", h.GetTokenTransfers)
			tokens.GET("/address/:address", h.GetTokenBalance)
		}

		// Indexer status
		v1.GET("/status", h.GetIndexerStatus)
		v1.POST("/reindex", h.TriggerReindex)
	}
}

// HealthCheck handles health check requests
func (h *HTTPHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "blockchain-indexer",
		"timestamp": time.Now().UTC(),
	})
}

// ReadinessCheck handles readiness check requests
func (h *HTTPHandler) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// ListBlocks handles block listing requests
func (h *HTTPHandler) ListBlocks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()
	blocks, err := h.queryService.ListBlocks(ctx, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list blocks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list blocks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blocks": blocks,
		"count":  len(blocks),
		"limit":  limit,
		"offset": offset,
	})
}

// GetBlock handles block retrieval requests
func (h *HTTPHandler) GetBlock(c *gin.Context) {
	hashOrNumber := c.Param("hash_or_number")
	ctx := c.Request.Context()

	var block *domain.Block
	var err error

	// Try to parse as number first
	if number, parseErr := strconv.ParseInt(hashOrNumber, 10, 64); parseErr == nil {
		block, err = h.queryService.GetBlockByNumber(ctx, number)
	} else {
		block, err = h.queryService.GetBlockByHash(ctx, hashOrNumber)
	}

	if err != nil {
		h.logger.Error("Failed to get block", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get block"})
		return
	}

	if block == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
		return
	}

	c.JSON(http.StatusOK, block)
}

// GetLatestBlock handles latest block retrieval
func (h *HTTPHandler) GetLatestBlock(c *gin.Context) {
	ctx := c.Request.Context()
	block, err := h.queryService.GetLatestBlock(ctx)
	if err != nil {
		h.logger.Error("Failed to get latest block", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get latest block"})
		return
	}

	if block == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no blocks found"})
		return
	}

	c.JSON(http.StatusOK, block)
}

// ListTransactions handles transaction listing requests
func (h *HTTPHandler) ListTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()
	transactions, err := h.queryService.ListTransactions(ctx, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list transactions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"count":        len(transactions),
		"limit":        limit,
		"offset":       offset,
	})
}

// GetTransaction handles transaction retrieval requests
func (h *HTTPHandler) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")
	ctx := c.Request.Context()

	tx, err := h.queryService.GetTransactionByHash(ctx, hash)
	if err != nil {
		h.logger.Error("Failed to get transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transaction"})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GetTransactionsByAddress handles address transaction listing
func (h *HTTPHandler) GetTransactionsByAddress(c *gin.Context) {
	address := c.Param("address")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()
	transactions, err := h.queryService.GetTransactionsByAddress(ctx, address, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get transactions by address", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":     address,
		"transactions": transactions,
		"count":       len(transactions),
		"limit":       limit,
		"offset":      offset,
	})
}

// GetAddress handles address metadata retrieval
func (h *HTTPHandler) GetAddress(c *gin.Context) {
	address := c.Param("address")
	ctx := c.Request.Context()

	meta, err := h.queryService.GetAddressMetadata(ctx, address)
	if err != nil {
		h.logger.Error("Failed to get address", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get address"})
		return
	}

	if meta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}

	c.JSON(http.StatusOK, meta)
}

// GetAddressLogs handles address logs retrieval
func (h *HTTPHandler) GetAddressLogs(c *gin.Context) {
	address := c.Param("address")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()
	logs, err := h.queryService.GetLogsByAddress(ctx, address, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get address logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"logs":    logs,
		"count":   len(logs),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetTokenTransfers handles token transfer listing
func (h *HTTPHandler) GetTokenTransfers(c *gin.Context) {
	tokenAddress := c.Query("token_address")
	holderAddress := c.Query("holder_address")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()
	transfers, err := h.queryService.GetTokenTransfers(ctx, tokenAddress, holderAddress, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get token transfers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get token transfers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_address":  tokenAddress,
		"holder_address": holderAddress,
		"transfers":      transfers,
		"count":          len(transfers),
		"limit":          limit,
		"offset":         offset,
	})
}

// GetTokenBalance handles token balance queries
func (h *HTTPHandler) GetTokenBalance(c *gin.Context) {
	address := c.Param("address")
	ctx := c.Request.Context()

	balance, err := h.queryService.GetTokenBalances(ctx, address)
	if err != nil {
		h.logger.Error("Failed to get token balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get token balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"tokens":  balance,
	})
}

// GetIndexerStatus handles indexer status requests
func (h *HTTPHandler) GetIndexerStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status, err := h.queryService.GetSyncStatus(ctx)
	if err != nil {
		h.logger.Error("Failed to get indexer status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// TriggerReindex handles reindex trigger requests
func (h *HTTPHandler) TriggerReindex(c *gin.Context) {
	// This would trigger a background reindex operation
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Reindex operation triggered",
	})
}

// QueryService methods

// ListBlocks lists blocks with pagination
func (s *QueryService) ListBlocks(ctx context.Context, limit, offset int) ([]domain.Block, error) {
	// For now, return mock data
	return []domain.Block{}, nil
}

// GetBlockByHash gets a block by hash
func (s *QueryService) GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error) {
	// Try cache first
	data, err := s.redis.GetCachedBlock(ctx, 0)
	if err == nil && data != nil {
		var block domain.Block
		if err := json.Unmarshal(data, &block); err == nil {
			return &block, nil
		}
	}

	return s.repo.GetBlockByHash(ctx, hash)
}

// GetBlockByNumber gets a block by number
func (s *QueryService) GetBlockByNumber(ctx context.Context, number int64) (*domain.Block, error) {
	// Try cache first
	data, err := s.redis.GetCachedBlock(ctx, number)
	if err == nil && data != nil {
		var block domain.Block
		if err := json.Unmarshal(data, &block); err == nil {
			return &block, nil
		}
	}

	block, err := s.repo.GetBlockByNumber(ctx, number)
	if err != nil || block == nil {
		return block, err
	}

	// Cache the result
	if data, err := json.Marshal(block); err == nil {
		s.redis.CacheBlock(ctx, number, data, time.Hour)
	}

	return block, nil
}

// GetLatestBlock gets the latest block
func (s *QueryService) GetLatestBlock(ctx context.Context) (*domain.Block, error) {
	number, err := s.repo.GetLatestBlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetBlockByNumber(ctx, number)
}

// ListTransactions lists transactions with pagination
func (s *QueryService) ListTransactions(ctx context.Context, limit, offset int) ([]domain.Transaction, error) {
	// For now, return mock data
	return []domain.Transaction{}, nil
}

// GetTransactionByHash gets a transaction by hash
func (s *QueryService) GetTransactionByHash(ctx context.Context, hash string) (*domain.Transaction, error) {
	// Try cache first
	data, err := s.redis.GetCachedTransaction(ctx, hash)
	if err == nil && data != nil {
		var tx domain.Transaction
		if err := json.Unmarshal(data, &tx); err == nil {
			return &tx, nil
		}
	}

	tx, err := s.repo.GetTransactionByHash(ctx, hash)
	if err != nil || tx == nil {
		return tx, err
	}

	// Cache the result
	if data, err := json.Marshal(tx); err == nil {
		s.redis.CacheTransaction(ctx, hash, data, time.Hour)
	}

	return tx, nil
}

// GetTransactionsByAddress gets transactions for an address
func (s *QueryService) GetTransactionsByAddress(ctx context.Context, address string, limit, offset int) ([]domain.Transaction, error) {
	return s.repo.GetTransactionsByAddress(ctx, address, limit, offset)
}

// GetAddressMetadata gets address metadata
func (s *QueryService) GetAddressMetadata(ctx context.Context, address string) (*domain.AddressMeta, error) {
	// Try cache first
	data, err := s.redis.GetCachedAddress(ctx, address)
	if err == nil && data != nil {
		var meta domain.AddressMeta
		if err := json.Unmarshal(data, &meta); err == nil {
			return &meta, nil
		}
	}

	meta, err := s.repo.GetAddressMeta(ctx, address)
	if err != nil || meta == nil {
		return meta, err
	}

	// Cache the result
	if data, err := json.Marshal(meta); err == nil {
		s.redis.CacheAddress(ctx, address, data, time.Hour)
	}

	return meta, nil
}

// GetLogsByAddress gets logs for an address
func (s *QueryService) GetLogsByAddress(ctx context.Context, address string, limit, offset int) ([]domain.Log, error) {
	return s.repo.GetLogsByAddress(ctx, address, limit, offset)
}

// GetTokenTransfers gets token transfers
func (s *QueryService) GetTokenTransfers(ctx context.Context, tokenAddress, holderAddress string, limit, offset int) ([]domain.TokenTransfer, error) {
	// For now, return mock data
	return []domain.TokenTransfer{}, nil
}

// GetTokenBalances gets token balances for an address
func (s *QueryService) GetTokenBalances(ctx context.Context, address string) ([]domain.TokenTransfer, error) {
	// For now, return mock data
	return []domain.TokenTransfer{}, nil
}

// GetSyncStatus gets the sync status
func (s *QueryService) GetSyncStatus(ctx context.Context) (*domain.SyncStatus, error) {
	// For now, return mock data
	return &domain.SyncStatus{
		Network:          s.config.Blockchain.Network,
		CurrentBlock:     0,
		IndexedBlock:     0,
		IndexedPercent:   0,
		PendingBlocks:    0,
		LastSyncTime:     time.Now(),
		SyncStatus:       "syncing",
	}, nil
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
