package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/csic/platform/blockchain/indexer/internal/config"
	"github.com/csic/platform/blockchain/indexer/internal/domain"
	"github.com/csic/platform/blockchain/indexer/internal/repository"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// IndexerService handles block synchronization and event processing
type IndexerService struct {
	repo             *repository.PostgresRepository
	redis            *repository.RedisClient
	blockchainClient *BlockchainClient
	config           *config.Config
	logger           *zap.Logger
	mu               sync.RWMutex
	running          bool
}

// NewIndexerService creates a new indexer service
func NewIndexerService(
	repo *repository.PostgresRepository,
	redis *repository.RedisClient,
	blockchainClient *BlockchainClient,
	cfg *config.Config,
	logger *zap.Logger,
) *IndexingService {
	return &IndexerService{
		repo:             repo,
		redis:            redis,
		blockchainClient: blockchainClient,
		config:           cfg,
		logger:           logger,
	}
}

// SyncBlocks synchronizes blocks from the blockchain
func (s *IndexerService) SyncBlocks(ctx context.Context, batchSize int) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("indexer is already running")
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	// Get last indexed block
	lastIndexed, err := s.redis.GetLastIndexedBlock(ctx)
	if err != nil {
		lastIndexed = s.config.Blockchain.StartBlock - 1
	}

	s.logger.Info("Starting block synchronization",
		zap.Int64("last_indexed", lastIndexed))

	// Main sync loop
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Block synchronization stopped due to context cancellation")
			return ctx.Err()
		default:
			// Get current block number
			currentBlock, err := s.blockchainClient.GetCurrentBlockNumber(ctx)
			if err != nil {
				s.logger.Error("Failed to get current block number", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			// Calculate safe block (with confirmation blocks)
			safeBlock := currentBlock - int64(s.config.Blockchain.ConfirmationBlocks)

			// If no blocks indexed yet, start from configured start block
			if lastIndexed < s.config.Blockchain.StartBlock-1 {
				lastIndexed = s.config.Blockchain.StartBlock - 1
			}

			// Sync next batch
			if lastIndexed < safeBlock {
				startBlock := lastIndexed + 1
				endBlock := min(startBlock+int64(batchSize)-1, safeBlock)

				s.logger.Debug("Syncing block range",
					zap.Int64("start", startBlock),
					zap.Int64("end", endBlock))

				if err := s.syncBlockRange(ctx, startBlock, endBlock); err != nil {
					s.logger.Error("Failed to sync block range",
						zap.Int64("start", startBlock),
						zap.Int64("end", endBlock),
						zap.Error(err))
					time.Sleep(5 * time.Second)
					continue
				}

				lastIndexed = endBlock
				s.redis.SetLastIndexedBlock(ctx, lastIndexed)
				s.redis.IncrementIndexedBlocks(ctx)

				s.logger.Info("Block range synced",
					zap.Int64("start", startBlock),
					zap.Int64("end", endBlock),
					zap.Int64("last_indexed", lastIndexed))
			} else {
				// Wait for new blocks
				time.Sleep(time.Duration(s.config.Indexer.BatchInterval) * time.Millisecond)
			}
		}
	}
}

// syncBlockRange syncs a range of blocks
func (s *IndexerService) syncBlockRange(ctx context.Context, startBlock, endBlock int64) error {
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		// Check if already indexed
		indexed, err := s.redis.IsBlockIndexed(ctx, blockNum)
		if err != nil {
			return fmt.Errorf("failed to check if block is indexed: %w", err)
		}
		if indexed {
			continue
		}

		// Fetch block
		block, err := s.blockchainClient.GetBlockByNumber(ctx, blockNum, false)
		if err != nil {
			return fmt.Errorf("failed to fetch block %d: %w", blockNum, err)
		}

		// Insert block
		if err := s.repo.InsertBlock(ctx, block); err != nil {
			return fmt.Errorf("failed to insert block %d: %w", blockNum, err)
		}

		// Fetch and insert transactions
		blockWithTx, err := s.blockchainClient.GetBlockWithTransactions(ctx, blockNum)
		if err != nil {
			return fmt.Errorf("failed to fetch block with transactions %d: %w", blockNum, err)
		}

		for _, tx := range blockWithTx.Transactions {
			if err := s.repo.InsertTransaction(ctx, &tx); err != nil {
				s.logger.Warn("Failed to insert transaction",
					zap.String("tx_hash", tx.Hash),
					zap.Error(err))
			}

			// Update address metadata
			s.updateAddressMetadata(ctx, &tx)

			// Fetch and insert logs
			if err := s.processTransactionLogs(ctx, &tx); err != nil {
				s.logger.Warn("Failed to process transaction logs",
					zap.String("tx_hash", tx.Hash),
					zap.Error(err))
			}
		}

		// Mark as indexed
		s.redis.SetIndexedBlock(ctx, blockNum)

		s.logger.Debug("Block synced",
			zap.Int64("block_number", blockNum),
			zap.String("block_hash", block.Hash),
			zap.Int("transaction_count", block.TransactionCount))
	}

	return nil
}

// updateAddressMetadata updates address metadata for a transaction
func (s *IndexerService) updateAddressMetadata(ctx context.Context, tx *domain.Transaction) {
	now := time.Now()

	// Update from address
	fromMeta := &domain.AddressMeta{
		Address:   tx.From,
		Type:      domain.AddressTypeEOA,
		FirstSeen: now,
		LastSeen:  now,
		TxCount:   1,
	}

	// Check if it's a contract
	if tx.To == "" {
		fromMeta.Type = domain.AddressTypeContract
		fromMeta.IsContract = true
	}

	s.repo.UpsertAddressMeta(ctx, fromMeta)

	// Update to address if present
	if tx.To != "" {
		toMeta := &domain.AddressMeta{
			Address:   tx.To,
			Type:      domain.AddressTypeEOA,
			FirstSeen: now,
			LastSeen:  now,
			TxCount:   1,
		}
		s.repo.UpsertAddressMeta(ctx, toMeta)
	}
}

// processTransactionLogs processes and stores logs for a transaction
func (s *IndexerService) processTransactionLogs(ctx context.Context, tx *domain.Transaction) error {
	// Get transaction receipt
	receipt, err := s.blockchainClient.GetTransactionReceipt(ctx, common.HexToHash(tx.Hash))
	if err != nil {
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	// Process each log
	for _, ethLog := range receipt.Logs {
		log := &domain.Log{
			TransactionHash: tx.Hash,
			BlockHash:       tx.BlockHash,
			BlockNumber:     tx.BlockNumber,
			LogIndex:        ethLog.Index,
			Address:         ethLog.Address.Hex(),
			Topics:          make([]string, len(ethLog.Topics)),
			Data:            fmt.Sprintf("0x%x", ethLog.Data),
			Removed:         ethLog.Removed,
			CreatedAt:       time.Now(),
		}

		for i, topic := range ethLog.Topics {
			log.Topics[i] = topic.Hex()
		}

		if err := s.repo.InsertLog(ctx, log); err != nil {
			return fmt.Errorf("failed to insert log: %w", err)
		}

		// Check for token transfers
		if len(log.Topics) >= 3 && log.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			transfer, err := ParseERC20Transfer(*ethLog)
			if err == nil {
				transfer.Timestamp = tx.Timestamp
				if err := s.repo.InsertTokenTransfer(ctx, transfer); err != nil {
					s.logger.Warn("Failed to insert token transfer", zap.Error(err))
				}
			}
		}
	}

	return nil
}

// ProcessEvents processes blockchain events
func (s *IndexerService) ProcessEvents(ctx context.Context) error {
	s.logger.Info("Starting event processing")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Event processing stopped")
			return ctx.Err()
		default:
			// Get pending events from queue
			blockNum, err := s.redis.GetBlockFromQueue(ctx)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			// Process block events
			if err := s.processBlockEvents(ctx, blockNum); err != nil {
				s.logger.Error("Failed to process block events",
					zap.Int64("block_number", blockNum),
					zap.Error(err))
				// Requeue for retry
				s.redis.AddBlockToQueue(ctx, blockNum)
			}
		}
	}
}

// processBlockEvents processes events for a specific block
func (s *IndexerService) processBlockEvents(ctx context.Context, blockNum int64) error {
	block, err := s.blockchainClient.GetBlockByNumber(ctx, blockNum, true)
	if err != nil {
		return fmt.Errorf("failed to fetch block: %w", err)
	}

	// Create indexer event
	event := &domain.IndexerEvent{
		Type:      "BLOCK_MINED",
		Block:     block,
		Timestamp: time.Now(),
	}

	// Marshal event for Kafka
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// TODO: Publish to Kafka topic
	s.logger.Debug("Block event processed",
		zap.Int64("block_number", blockNum),
		zap.String("block_hash", block.Hash))

	return nil
}

// GetSyncStatus gets the current sync status
func (s *IndexerService) GetSyncStatus(ctx context.Context) (*domain.SyncStatus, error) {
	currentBlock, err := s.blockchainClient.GetCurrentBlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current block: %w", err)
	}

	lastIndexed, err := s.redis.GetLastIndexedBlock(ctx)
	if err != nil {
		lastIndexed = 0
	}

	totalIndexed, err := s.redis.GetTotalIndexedBlocks(ctx)
	if err != nil {
		totalIndexed = 0
	}

	status := &domain.SyncStatus{
		Network:        s.config.Blockchain.Network,
		CurrentBlock:   currentBlock,
		IndexedBlock:   lastIndexed,
		IndexedPercent: 0,
		PendingBlocks:  currentBlock - lastIndexed,
		LastSyncTime:   time.Now(),
		SyncStatus:     "syncing",
	}

	if currentBlock > 0 {
		status.IndexedPercent = float64(lastIndexed) / float64(currentBlock) * 100
	}

	if lastIndexed >= currentBlock {
		status.SyncStatus = "synced"
	}

	status.TotalIndexedBlocks = totalIndexed

	return status, nil
}

// Reindex performs a full reindex from the start block
func (s *IndexerService) Reindex(ctx context.Context) error {
	if !s.config.Indexer.ReindexEnabled {
		return fmt.Errorf("reindexing is disabled")
	}

	s.logger.Info("Starting full reindex",
		zap.Int64("start_block", s.config.Blockchain.StartBlock))

	// Clear existing index
	// TODO: Clear Redis cache

	// Reset last indexed block
	s.redis.SetLastIndexedBlock(ctx, s.config.Blockchain.StartBlock-1)

	// Restart sync
	return s.SyncBlocks(ctx, s.config.Indexer.ReindexBatchSize)
}

// Stop stops the indexer
func (s *IndexerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}

// IsRunning returns whether the indexer is running
func (s *IndexerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
