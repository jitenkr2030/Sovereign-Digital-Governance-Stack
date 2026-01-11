package repository

import (
	"context"
	"time"

	"csic-platform/service/security/internal/domain/models"
)

// AuditRepository defines the interface for audit log persistence operations.
// It supports tamper-proof logging through hash chain verification.
type AuditRepository interface {
	// Create creates a new audit log entry with hash chain linkage.
	// The entry's PreviousHash and CurrentHash are computed during creation.
	Create(ctx context.Context, log *models.AuditLog) error

	// GetByID retrieves an audit log entry by its unique identifier.
	GetByID(ctx context.Context, id string) (*models.AuditLog, error)

	// GetByTimestampRange retrieves audit logs within a specified time range.
	GetByTimestampRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*models.AuditLog, error)

	// GetByEntity retrieves audit logs for a specific entity (e.g., "user:123", "license:456").
	GetByEntity(ctx context.Context, entityID string, limit, offset int) ([]*models.AuditLog, error)

	// VerifyChain performs cryptographic verification of the hash chain.
	// Returns true if the chain is valid, false if tampering is detected.
	VerifyChain(ctx context.Context, startID, endID string) (bool, error)

	// GetChainHead retrieves the most recent audit log entry (the head of the chain).
	GetChainHead(ctx context.Context) (*models.AuditLog, error)

	// Count returns the total number of audit log entries.
	Count(ctx context.Context) (int64, error)

	// Search performs a full-text search on audit log metadata and details.
	Search(ctx context.Context, query string, limit, offset int) ([]*models.AuditLog, error)
}

// KeyManagementRepository defines the interface for cryptographic key management operations.
type KeyManagementRepository interface {
	// StoreKey securely stores a cryptographic key with its metadata.
	StoreKey(ctx context.Context, key *models.CryptoKey) error

	// RetrieveKey retrieves a cryptographic key by its identifier.
	RetrieveKey(ctx context.Context, keyID string) (*models.CryptoKey, error)

	// GetKeyVersion retrieves a specific version of a key.
	GetKeyVersion(ctx context.Context, keyID string, version int) (*models.CryptoKey, error)

	// ListKeys lists all active keys for a given purpose.
	ListKeys(ctx context.Context, purpose models.KeyPurpose) ([]*models.CryptoKey, error)

	// MarkKeyRevoked marks a key as revoked and optionally archives it.
	MarkKeyRevoked(ctx context.Context, keyID string, reason string) error

	// RotateKey creates a new version of an existing key for rotation purposes.
	RotateKey(ctx context.Context, keyID string) (*models.CryptoKey, error)

	// GetKeyByPurpose retrieves the active key for a specific purpose.
	GetActiveKey(ctx context.Context, purpose models.KeyPurpose) (*models.CryptoKey, error)
}

// SIEMEventRepository defines the interface for SIEM integration events.
type SIEMEventRepository interface {
	// StoreEvent stores a SIEM event for forwarding to external systems.
	StoreEvent(ctx context.Context, event *models.SIEMEvent) error

	// GetPendingEvents retrieves events that haven't been forwarded yet.
	GetPendingEvents(ctx context.Context, limit int) ([]*models.SIEMEvent, error)

	// MarkEventForwarded marks an event as successfully forwarded.
	MarkEventForwarded(ctx context.Context, eventID string, remoteID string) error

	// MarkEventFailed marks an event as failed after maximum retries.
	MarkEventFailed(ctx context.Context, eventID string, errorMsg string) error

	// GetEventStats retrieves statistics about SIEM event forwarding.
	GetEventStats(ctx context.Context) (*models.SIEMEventStats, error)
}

// HSMRepository defines the interface for Hardware Security Module operations.
// This abstraction allows for different HSM implementations or software fallbacks.
type HSMRepository interface {
	// Initialize prepares the HSM connection with the provided configuration.
	Initialize(ctx context.Context, config *models.HSMConfig) error

	// Sign performs a digital signature operation using the specified key.
	Sign(ctx context.Context, keyID string, data []byte) ([]byte, error)

	// Verify verifies a digital signature against the provided data.
	Verify(ctx context.Context, keyID string, data, signature []byte) (bool, error)

	// GenerateKey generates a new cryptographic key within the HSM.
	GenerateKey(ctx context.Context, keySpec models.HSMKeySpec) (*models.HSMKeyInfo, error)

	// ImportKey imports an existing key into the HSM.
	ImportKey(ctx context.Context, keyID string, keyData []byte, keySpec models.HSMKeySpec) (*models.HSMKeyInfo, error)

	// ExportKey exports a public key from the HSM in DER format.
	ExportPublicKey(ctx context.Context, keyID string) ([]byte, error)

	// DeleteKey removes a key from the HSM.
	DeleteKey(ctx context.Context, keyID string) error

	// GetKeyInfo retrieves metadata about a key stored in the HSM.
	GetKeyInfo(ctx context.Context, keyID string) (*models.HSMKeyInfo, error)

	// Encrypt encrypts data using the specified key (for symmetric encryption).
	Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error)

	// Decrypt decrypts data using the specified key (for symmetric encryption).
	Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error)

	// HealthCheck verifies the HSM connection is healthy.
	HealthCheck(ctx context.Context) error

	// Close releases all resources held by the HSM connection.
	Close() error
}
