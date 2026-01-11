package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/csic-platform/blockchain/parser/blocks/domain"
	"github.com/csic-platform/blockchain/parser/blocks/ports"
)

// BlockService implements the BlockService input port
type BlockService struct {
	parser    ports.BlockParser
	repo      ports.BlockRepository
	publisher ports.BlockEventPublisher
	mu        sync.RWMutex
}

// NewBlockService creates a new block service
func NewBlockService(parser ports.BlockParser, repo ports.BlockRepository, publisher ports.BlockEventPublisher) *BlockService {
	return &BlockService{
		parser:    parser,
		repo:      repo,
		publisher: publisher,
	}
}

// ParseRawBlock parses raw block data
func (s *BlockService) ParseRawBlock(ctx context.Context, rawData []byte) (*domain.Block, error) {
	return s.parser.ParseBlock(ctx, rawData)
}

// FetchAndParseBlock fetches and parses a block by hash or height
func (s *BlockService) FetchAndParseBlock(ctx context.Context, hashOrHeight string) (*domain.Block, error) {
	// For now, assume hashOrHeight is a raw block data reference
	// In a real implementation, this would fetch from RPC or other source
	return s.repo.GetBlock(ctx, hashOrHeight)
}

// ParseBlockRange parses a range of blocks
func (s *BlockService) ParseBlockRange(ctx context.Context, startHeight, endHeight uint64) ([]*domain.Block, error) {
	var blocks []*domain.Block

	for height := startHeight; height <= endHeight; height++ {
		block, err := s.repo.GetBlockByHeight(ctx, height)
		if err != nil {
			return nil, fmt.Errorf("failed to get block at height %d: %w", height, err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetBlock retrieves a block by hash or height
func (s *BlockService) GetBlock(ctx context.Context, hashOrHeight string) (*domain.Block, error) {
	return s.repo.GetBlock(ctx, hashOrHeight)
}

// GetLatestBlock retrieves the most recent block
func (s *BlockService) GetLatestBlock(ctx context.Context) (*domain.Block, error) {
	return s.repo.GetLatestBlock(ctx)
}

// GetBlocks retrieves blocks matching the filter criteria
func (s *BlockService) GetBlocks(ctx context.Context, filter domain.BlockFilter) ([]*domain.Block, error) {
	return s.repo.ListBlocks(ctx, filter)
}

// ProcessBlock processes a single block
func (s *BlockService) ProcessBlock(ctx context.Context, block *domain.Block) error {
	// Validate the block
	if err := s.parser.ValidateBlock(ctx, block); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	// Store the block
	if err := s.repo.StoreBlock(ctx, block); err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}

	// Publish new block event
	if s.publisher != nil {
		if err := s.publisher.PublishNewBlock(ctx, block); err != nil {
			// Log error but don't fail the processing
			fmt.Printf("Failed to publish new block event: %v\n", err)
		}
	}

	return nil
}

// ProcessBlockRange processes a range of blocks
func (s *BlockService) ProcessBlockRange(ctx context.Context, startHeight, endHeight uint64) error {
	for height := startHeight; height <= endHeight; height++ {
		block, err := s.repo.GetBlockByHeight(ctx, height)
		if err != nil {
			return fmt.Errorf("failed to get block at height %d: %w", height, err)
		}

		if err := s.ProcessBlock(ctx, block); err != nil {
			return fmt.Errorf("failed to process block at height %d: %w", height, err)
		}
	}

	return nil
}

// SyncBlocks syncs blocks from a starting height
func (s *BlockService) SyncBlocks(ctx context.Context, fromHeight uint64) error {
	latestBlock, err := s.repo.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	if latestBlock == nil {
		return fmt.Errorf("no latest block found")
	}

	return s.ProcessBlockRange(ctx, fromHeight, latestBlock.Height)
}

// GetBlockStatistics calculates statistics for a range of blocks
func (s *BlockService) GetBlockStatistics(ctx context.Context, startHeight, endHeight uint64) (*domain.BlockStats, error) {
	return s.repo.GetBlockStats(ctx, startHeight, endHeight)
}

// BlockRepositoryAdapter adapts BlockRepository to support BlockService operations
type BlockRepositoryAdapter struct {
	repo ports.BlockRepository
}

// NewBlockRepositoryAdapter creates a new repository adapter
func NewBlockRepositoryAdapter(repo ports.BlockRepository) *BlockRepositoryAdapter {
	return &BlockRepositoryAdapter{repo: repo}
}

// StoreBlock stores a single block
func (a *BlockRepositoryAdapter) StoreBlock(ctx context.Context, block *domain.Block) error {
	return a.repo.StoreBlock(ctx, block)
}

// StoreBlocks stores multiple blocks
func (a *BlockRepositoryAdapter) StoreBlocks(ctx context.Context, blocks []*domain.Block) error {
	return a.repo.StoreBlocks(ctx, blocks)
}

// GetBlock retrieves a block by hash or height
func (a *BlockRepositoryAdapter) GetBlock(ctx context.Context, hashOrHeight string) (*domain.Block, error) {
	// Try hash first
	block, err := a.repo.GetBlockByHash(ctx, hashOrHeight)
	if err == nil {
		return block, nil
	}

	// Try height
	var height uint64
	if _, err := fmt.Sscanf(hashOrHeight, "%d", &height); err == nil {
		return a.repo.GetBlockByHeight(ctx, height)
	}

	return nil, fmt.Errorf("block not found: %s", hashOrHeight)
}

// GetBlockByHeight retrieves a block by height
func (a *BlockRepositoryAdapter) GetBlockByHeight(ctx context.Context, height uint64) (*domain.Block, error) {
	return a.repo.GetBlockByHeight(ctx, height)
}

// GetBlockByHash retrieves a block by hash
func (a *BlockRepositoryAdapter) GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error) {
	return a.repo.GetBlockByHash(ctx, hash)
}

// GetLatestBlock retrieves the most recent block
func (a *BlockRepositoryAdapter) GetLatestBlock(ctx context.Context) (*domain.Block, error) {
	return a.repo.GetLatestBlock(ctx)
}

// ListBlocks retrieves blocks matching filter criteria
func (a *BlockRepositoryAdapter) ListBlocks(ctx context.Context, filter domain.BlockFilter) ([]*domain.Block, error) {
	return a.repo.ListBlocks(ctx, filter)
}

// GetBlocksByTimeRange retrieves blocks within a time range
func (a *BlockRepositoryAdapter) GetBlocksByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*domain.Block, error) {
	return a.repo.GetBlocksByTimeRange(ctx, start, end, limit)
}

// GetBlocksByMiner retrieves blocks mined by a specific address
func (a *BlockRepositoryAdapter) GetBlocksByMiner(ctx context.Context, miner string, limit, offset int) ([]*domain.Block, error) {
	return a.repo.GetBlocksByMiner(ctx, miner, limit, offset)
}

// GetBlockStats retrieves statistics for a range of blocks
func (a *BlockRepositoryAdapter) GetBlockStats(ctx context.Context, startHeight, endHeight uint64) (*domain.BlockStats, error) {
	return a.repo.GetBlockStats(ctx, startHeight, endHeight)
}

// CountBlocks counts blocks matching filter criteria
func (a *BlockRepositoryAdapter) CountBlocks(ctx context.Context, filter domain.BlockFilter) (int64, error) {
	return a.repo.CountBlocks(ctx, filter)
}

// DeleteBlock removes a block from storage
func (a *BlockRepositoryAdapter) DeleteBlock(ctx context.Context, hash string) error {
	return a.repo.DeleteBlock(ctx, hash)
}

// PruneBlocks removes old blocks while keeping recent ones
func (a *BlockRepositoryAdapter) PruneBlocks(ctx context.Context, beforeHeight uint64, keepCount uint64) error {
	return a.repo.PruneBlocks(ctx, beforeHeight, keepCount)
}
