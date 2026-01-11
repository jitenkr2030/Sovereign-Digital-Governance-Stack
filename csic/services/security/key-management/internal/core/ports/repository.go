package ports

import (
	"context"

	"github.com/csic-platform/services/security/key-management/internal/core/domain"
)

// KeyRepository defines the interface for key data access operations
type KeyRepository interface {
	// Key lifecycle operations
	CreateKey(ctx context.Context, key *domain.Key) error
	GetKey(ctx context.Context, id string) (*domain.Key, error)
	GetKeyByAlias(ctx context.Context, alias string) (*domain.Key, error)
	ListKeys(ctx context.Context, page, pageSize int) ([]*domain.Key, int64, error)
	UpdateKey(ctx context.Context, key *domain.Key) error
	DeleteKey(ctx context.Context, id string) error

	// Key version operations
	CreateKeyVersion(ctx context.Context, version *domain.KeyVersion) error
	GetKeyVersion(ctx context.Context, keyID string, version int) (*domain.KeyVersion, error)
	GetLatestKeyVersion(ctx context.Context, keyID string) (*domain.KeyVersion, error)
	GetAllKeyVersions(ctx context.Context, keyID string) ([]*domain.KeyVersion, error)
	
	// Key policy operations
	GetKeyPolicy(ctx context.Context, keyID string) (*domain.KeyPolicy, error)
	UpdateKeyPolicy(ctx context.Context, policy *domain.KeyPolicy) error
	
	// Audit operations
	AuditKeyOperation(ctx context.Context, operation *domain.KeyOperation) error
	GetKeyAuditLog(ctx context.Context, keyID string, limit int) ([]*domain.KeyOperation, error)
}

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	// Key metadata caching
	GetKeyMetadata(ctx context.Context, keyID string) (*domain.Key, error)
	SetKeyMetadata(ctx context.Context, keyID string, key *domain.Key, ttl int) error
	DeleteKeyMetadata(ctx context.Context, keyID string) error
	
	// Public key caching (for asymmetric keys)
	GetPublicKey(ctx context.Context, keyID string) ([]byte, error)
	SetPublicKey(ctx context.Context, keyID string, publicKey []byte, ttl int) error
	
	// Invalidation
	InvalidateKeyCache(ctx context.Context, keyID string) error
}

// MessageProducer defines the interface for message publishing operations
type MessageProducer interface {
	PublishKeyOperation(ctx context.Context, operation *domain.KeyOperation) error
	PublishKeyLifecycleEvent(ctx context.Context, eventType string, key *domain.Key) error
}
