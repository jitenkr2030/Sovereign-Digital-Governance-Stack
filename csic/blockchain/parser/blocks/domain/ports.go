package ports

import (
	"context"

	"github.com/csic-platform/blockchain/parser/blocks/domain"
	"github.com/csic-platform/blockchain/parser/transactions/domain"
)

// BlockRepository defines the interface for block storage operations
type BlockRepository interface {
	// Store operations
	StoreBlock(ctx context.Context, block *domain.Block) error
	StoreBlocks(ctx context.Context, blocks []*domain.Block) error

	// Retrieval operations
	GetBlock(ctx context.Context, hashOrHeight string) (*domain.Block, error)
	GetBlockByHash(ctx context.Context, hash string) (*domain.Block, error)
	GetBlockByHeight(ctx context.Context, height uint64) (*domain.Block, error)
	GetLatestBlock(ctx context.Context) (*domain.Block, error)
	GetBlockHeader(ctx context.Context, hashOrHeight string) (*domain.BlockHeader, error)

	// Query operations
	ListBlocks(ctx context.Context, filter domain.BlockFilter) ([]*domain.Block, error)
	GetBlocksByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*domain.Block, error)
	GetBlocksByMiner(ctx context.Context, miner string, limit, offset int) ([]*domain.Block, error)

	// Statistics
	GetBlockStats(ctx context.Context, startHeight, endHeight uint64) (*domain.BlockStats, error)
	CountBlocks(ctx context.Context, filter domain.BlockFilter) (int64, error)

	// Cleanup
	DeleteBlock(ctx context.Context, hash string) error
	PruneBlocks(ctx context.Context, beforeHeight uint64, keepCount uint64) error
}

// BlockParser defines the interface for parsing raw block data
type BlockParser interface {
	// Parse methods
	ParseBlock(ctx context.Context, rawData []byte) (*domain.Block, error)
	ParseBlockHeader(ctx context.Context, rawData []byte) (*domain.BlockHeader, error)
	ParseBlockFromRPC(ctx context.Context, rpcResponse map[string]interface{}) (*domain.Block, error)

	// Validation
	ValidateBlock(ctx context.Context, block *domain.Block) error
	ValidateBlockHeader(ctx context.Context, header *domain.BlockHeader) error

	// Serialization
	SerializeBlock(ctx context.Context, block *domain.Block) ([]byte, error)
	DeserializeBlock(ctx context.Context, data []byte) (*domain.Block, error)
}

// BlockEventPublisher defines the interface for publishing block events
type BlockEventPublisher interface {
	PublishNewBlock(ctx context.Context, block *domain.Block) error
	PublishBlockFinalized(ctx context.Context, block *domain.Block) error
	PublishReorgDetected(ctx context.Context, oldTip, newTip *domain.Block) error
}

// BlockService defines the input port for block operations
type BlockService interface {
	// Parsing
	ParseRawBlock(ctx context.Context, rawData []byte) (*domain.Block, error)
	FetchAndParseBlock(ctx context.Context, hashOrHeight string) (*domain.Block, error)
	ParseBlockRange(ctx context.Context, startHeight, endHeight uint64) ([]*domain.Block, error)

	// Retrieval
	GetBlock(ctx context.Context, hashOrHeight string) (*domain.Block, error)
	GetLatestBlock(ctx context.Context) (*domain.Block, error)
	GetBlocks(ctx context.Context, filter domain.BlockFilter) ([]*domain.Block, error)

	// Processing
	ProcessBlock(ctx context.Context, block *domain.Block) error
	ProcessBlockRange(ctx context.Context, startHeight, endHeight uint64) error
	SyncBlocks(ctx context.Context, fromHeight uint64) error

	// Statistics
	GetBlockStatistics(ctx context.Context, startHeight, endHeight uint64) (*domain.BlockStats, error)
}
