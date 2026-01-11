package ports

import (
	"context"

	"github.com/csic-platform/blockchain/parser/mempool/domain"
)

// MempoolRepository defines the interface for mempool storage
type MempoolRepository interface {
	// Transaction operations
	AddTransaction(ctx context.Context, tx *domain.MempoolTransaction) error
	RemoveTransaction(ctx context.Context, txHash string) error
	GetTransaction(ctx context.Context, txHash string) (*domain.MempoolTransaction, error)
	GetAllTransactions(ctx context.Context, filter domain.MempoolFilter) ([]*domain.MempoolTransaction, error)
	GetTransactionCount(ctx context.Context) (int64, error)

	// Statistics
	GetMempoolStats(ctx context.Context) (*domain.MempoolStats, error)
	GetGasPriceDistribution(ctx context.Context) (map[string]int, error)

	// Cleanup
	CleanupStaleTransactions(ctx context.Context, maxAge time.Duration) (int64, error)
	PruneByGasPrice(ctx context.Context, minGasPrice uint64) (int64, error)
}

// MempoolParser defines the interface for parsing raw mempool transactions
type MempoolParser interface {
	ParseTransaction(ctx context.Context, rawData []byte) (*domain.MempoolTransaction, error)
	ParseTransactions(ctx context.Context, rawData [][]byte) ([]*domain.MempoolTransaction, error)
	ValidateTransaction(ctx context.Context, tx *domain.MempoolTransaction) error
	EstimateConfirmTime(ctx context.Context, tx *domain.MempoolTransaction) float64
}

// MempoolService defines the input port for mempool operations
type MempoolService interface {
	// Transaction management
	AddTransaction(ctx context.Context, rawData []byte) (*domain.MempoolTransaction, error)
	RemoveTransaction(ctx context.Context, txHash string) error
	GetTransaction(ctx context.Context, txHash string) (*domain.MempoolTransaction, error)
	GetTransactions(ctx context.Context, filter domain.MempoolFilter) ([]*domain.MempoolTransaction, error)

	// Statistics and analysis
	GetMempoolStats(ctx context.Context) (*domain.MempoolStats, error)
	GetGasPricePercentiles(ctx context.Context, percentiles []float64) ([]float64, error)

	// Block building
	BuildBlock(ctx context.Context, gasLimit uint64) (*domain.PendingBlock, error)

	// Maintenance
	MaintainMempool(ctx context.Context, maxSize int, maxAge time.Duration) error
}
