package ports

import (
	"context"
	"time"

	"github.com/csic/monitoring/internal/core/domain"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
	GetByHash(ctx context.Context, txHash string) (*domain.Transaction, error)
	Update(ctx context.Context, tx *domain.Transaction) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *domain.TransactionFilter) ([]*domain.Transaction, int64, error)
	GetByAddress(ctx context.Context, address, chain string, limit int) ([]*domain.Transaction, error)
	GetFlagged(ctx context.Context, page, pageSize int) ([]*domain.Transaction, int64, error)
}

// SanctionsRepository defines the interface for sanctions list data access
type SanctionsRepository interface {
	Create(ctx context.Context, sanction *domain.SanctionedAddress) error
	CreateBatch(ctx context.Context, sanctions []*domain.SanctionedAddress) error
	GetByID(ctx context.Context, id string) (*domain.SanctionedAddress, error)
	GetByAddress(ctx context.Context, address, chain string) ([]*domain.SanctionedAddress, error)
	Exists(ctx context.Context, address, chain string) (bool, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int) ([]*domain.SanctionedAddress, int64, error)
	Search(ctx context.Context, query string) ([]*domain.SanctionedAddress, error)
}

// WalletProfileRepository defines the interface for wallet profile data access
type WalletProfileRepository interface {
	Create(ctx context.Context, profile *domain.WalletProfile) error
	Update(ctx context.Context, profile *domain.WalletProfile) error
	GetByID(ctx context.Context, id string) (*domain.WalletProfile, error)
	GetByAddress(ctx context.Context, address, chain string) (*domain.WalletProfile, error)
	Upsert(ctx context.Context, profile *domain.WalletProfile) error
	GetHighRisk(ctx context.Context, limit int) ([]*domain.WalletProfile, error)
	Search(ctx context.Context, query string, page, pageSize int) ([]*domain.WalletProfile, int64, error)
}

// RiskEngine defines the interface for risk calculation
type RiskEngine interface {
	CalculateRisk(ctx context.Context, tx *domain.Transaction) (*domain.RiskAssessment, error)
	CalculateWalletRisk(ctx context.Context, address, chain string) (*domain.WalletProfile, error)
}
