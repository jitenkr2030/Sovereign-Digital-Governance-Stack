package ports

import (
	"context"

	"github.com/csic-platform/services/security/key-management/internal/core/domain"
)

// KeyService defines the interface for key management business logic
type KeyService interface {
	// Key lifecycle management
	CreateKey(ctx context.Context, req *domain.CreateKeyRequest, actorID string) (*domain.Key, error)
	GetKey(ctx context.Context, id string, actorID string) (*domain.Key, error)
	ListKeys(ctx context.Context, page, pageSize int, actorID string) (*domain.KeyListResponse, error)
	RotateKey(ctx context.Context, id string, req *domain.RotateKeyRequest, actorID string) (*domain.Key, error)
	RevokeKey(ctx context.Context, id string, req *domain.RevokeKeyRequest, actorID string) error
	
	// Data key operations (envelope encryption)
	GenerateDataKey(ctx context.Context, req *domain.GenerateDataKeyRequest, actorID string) (*domain.DataKeyResponse, error)
	EncryptData(ctx context.Context, req *domain.EncryptRequest, actorID string) (*domain.EncryptDecryptResponse, error)
	DecryptData(ctx context.Context, req *domain.DecryptRequest, actorID string) (*domain.EncryptDecryptResponse, error)
	
	// Key metadata operations
	UpdateKeyMetadata(ctx context.Context, id string, metadata map[string]string, actorID string) error
	
	// Audit and querying
	GetKeyAuditLog(ctx context.Context, keyID string, limit int, actorID string) ([]*domain.KeyOperation, error)
	GetKeyVersions(ctx context.Context, keyID string, actorID string) ([]*domain.KeyVersion, error)
	
	// Health check
	HealthCheck(ctx context.Context) error
}

// CryptoProvider defines the interface for cryptographic operations
type CryptoProvider interface {
	// Key generation
	GenerateSymmetricKey(algorithm domain.KeyAlgorithm) ([]byte, error)
	GenerateAsymmetricKeyPair(algorithm domain.KeyAlgorithm) (publicKey, privateKey []byte, err error)
	
	// Encryption/Decryption (symmetric)
	Encrypt(algorithm domain.KeyAlgorithm, key, plaintext, iv []byte) ([]byte, error)
	Decrypt(algorithm domain.KeyAlgorithm, key, ciphertext, iv []byte) ([]byte, error)
	
	// Encryption/Decryption (asymmetric)
	EncryptWithPublicKey(algorithm domain.KeyAlgorithm, publicKey, plaintext []byte) ([]byte, error)
	DecryptWithPrivateKey(algorithm domain.KeyAlgorithm, privateKey, ciphertext []byte) ([]byte, error)
	
	// Signing/Verification
	Sign(algorithm domain.KeyAlgorithm, privateKey, message []byte) ([]byte, error)
	Verify(algorithm domain.KeyAlgorithm, publicKey, message, signature []byte) (bool, error)
	
	// Key wrapping (for envelope encryption)
	WrapKey(algorithm domain.KeyAlgorithm, kek, plaintextKey []byte) ([]byte, []byte, error)
	UnwrapKey(algorithm domain.KeyAlgorithm, kek, wrappedKey, iv []byte) ([]byte, error)
	
	// Utility
	GenerateIV(algorithm domain.KeyAlgorithm) ([]byte, error)
	DeriveKeyFromPassword(password string, salt []byte, iterations int) ([]byte, error)
}

// AuditService defines the interface for audit logging
type AuditService interface {
	LogKeyOperation(ctx context.Context, operation *domain.KeyOperation) error
	LogKeyEvent(ctx context.Context, eventType string, key *domain.Key, actorID string, metadata map[string]interface{}) error
}
