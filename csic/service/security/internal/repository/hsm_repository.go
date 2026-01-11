package repository

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"csic-platform/service/security/internal/domain/models"
)

// ErrHSMUnavailable indicates that the HSM is not available and no fallback is configured.
var ErrHSMUnavailable = errors.New("HSM is unavailable and no software fallback is configured")

// ErrKeyNotFound indicates that the requested key was not found in the HSM.
var ErrKeyNotFound = errors.New("key not found in HSM")

// ErrInvalidKeyType indicates that the key type is not supported.
var ErrInvalidKeyType = errors.New("invalid key type for operation")

// ErrSignatureInvalid indicates that the signature verification failed.
var ErrSignatureInvalid = errors.New("signature verification failed")

// SoftwareHSMRepository provides a software-based implementation of HSM operations.
// This implementation uses Go's standard crypto library and is suitable for development,
// testing, and scenarios where a physical HSM is not available.
type SoftwareHSMRepository struct {
	mu         sync.RWMutex
	keys       map[string]*storedKey
	config     *models.HSMConfig
	initialized bool
}

// storedKey represents a key stored in the software HSM.
type storedKey struct {
	KeyID       string
	KeyType     models.HSMKeyType
	Purpose     models.KeyPurpose
	KeyData     []byte
	PublicKey   []byte
	CreatedAt   time.Time
	LastUsed    time.Time
	UsageCount  int
	Metadata    map[string]string
}

// NewSoftwareHSMRepository creates a new software-based HSM repository.
func NewSoftwareHSMRepository() *SoftwareHSMRepository {
	return &SoftwareHSMRepository{
		keys: make(map[string]*storedKey),
	}
}

// Initialize prepares the software HSM with the provided configuration.
func (r *SoftwareHSMRepository) Initialize(ctx context.Context, config *models.HSMConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if config == nil {
		config = &models.HSMConfig{
			Provider:      "software",
			SoftHSMEnabled: true,
		}
	}

	r.config = config
	r.initialized = true

	// Initialize with a root key if none exists.
	if len(r.keys) == 0 {
		// Create a default root signing key for audit log hashing.
		rootKey := &storedKey{
			KeyID:     "root-signing-key",
			KeyType:   models.KeyTypeRSA2048,
			Purpose:   models.PurposeAuditSigning,
			CreatedAt: time.Now().UTC(),
			Metadata: map[string]string{
				"description": "Root signing key for audit log integrity",
				"algorithm":   "RSA-SHA256",
			},
		}

		// Generate RSA 2048 key pair.
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate root key: %w", err)
		}

		rootKey.KeyData = x509.MarshalPKCS1PrivateKey(privateKey)
		rootKey.PublicKey = x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)

		r.keys[rootKey.KeyID] = rootKey
	}

	return nil
}

// Sign performs a digital signature operation using the specified key.
func (r *SoftwareHSMRepository) Sign(ctx context.Context, keyID string, data []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	key, err := r.getKey(keyID)
	if err != nil {
		return nil, err
	}

	// Update usage statistics.
	key.UsageCount++
	key.LastUsed = time.Now().UTC()

	// Parse the private key.
	privateKey, err := parsePrivateKey(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create signature using SHA-256.
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	return signature, nil
}

// Verify verifies a digital signature against the provided data.
func (r *SoftwareHSMRepository) Verify(ctx context.Context, keyID string, data, signature []byte) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return false, ErrHSMUnavailable
	}

	key, err := r.getKey(keyID)
	if err != nil {
		return false, err
	}

	// Get the public key.
	var publicKey crypto.PublicKey
	switch key.KeyType {
	case models.KeyTypeRSA2048, models.KeyTypeRSA4096:
		pubKey, err := x509.ParsePKCS1PublicKey(key.PublicKey)
		if err != nil {
			return false, fmt.Errorf("failed to parse public key: %w", err)
		}
		publicKey = pubKey
	default:
		return false, ErrInvalidKeyType
	}

	// Verify the signature.
	hash := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		if errors.Is(err, rsa.ErrVerification) {
			return false, nil
		}
		return false, fmt.Errorf("signature verification error: %w", err)
	}

	return true, nil
}

// GenerateKey generates a new cryptographic key within the HSM.
func (r *SoftwareHSMRepository) GenerateKey(ctx context.Context, keySpec models.HSMKeySpec) (*models.HSMKeyInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	var keyID string
	var privateKey any
	var err error

	switch keySpec.Type {
	case models.KeyTypeRSA2048:
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		keyID = fmt.Sprintf("rsa-2048-%s", generateKeyID())
	case models.KeyTypeRSA4096:
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		keyID = fmt.Sprintf("rsa-4096-%s", generateKeyID())
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keySpec.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Extract key data for storage.
	var keyData, publicKey []byte
	priv := privateKey.(*rsa.PrivateKey)
	keyData = x509.MarshalPKCS1PrivateKey(priv)
	publicKey = x509.MarshalPKCS1PublicKey(&priv.PublicKey)

	// Store the key.
	storedKey := &storedKey{
		KeyID:     keyID,
		KeyType:   keySpec.Type,
		Purpose:   keySpec.Purpose,
		KeyData:   keyData,
		PublicKey: publicKey,
		CreatedAt: time.Now().UTC(),
		Metadata:  keySpec.Metadata,
	}

	r.keys[keyID] = storedKey

	return &models.HSMKeyInfo{
		KeyID:       keyID,
		KeyType:     keySpec.Type,
		Purpose:     keySpec.Purpose,
		PublicKey:   hex.EncodeToString(publicKey),
		CreatedAt:   storedKey.CreatedAt,
		UsageCount:  0,
		Algorithm:   "RSA-SHA256",
	}, nil
}

// ImportKey imports an existing key into the HSM.
func (r *SoftwareHSMRepository) ImportKey(ctx context.Context, keyID string, keyData []byte, keySpec models.HSMKeySpec) (*models.HSMKeyInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	// Parse and validate the key.
	privateKey, err := parsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrInvalidKeyType
	}

	// Store the key.
	storedKey := &storedKey{
		KeyID:     keyID,
		KeyType:   keySpec.Type,
		Purpose:   keySpec.Purpose,
		KeyData:   keyData,
		PublicKey: x509.MarshalPKCS1PublicKey(&rsaKey.PublicKey),
		CreatedAt: time.Now().UTC(),
		Metadata:  keySpec.Metadata,
	}

	r.keys[keyID] = storedKey

	return &models.HSMKeyInfo{
		KeyID:       keyID,
		KeyType:     keySpec.Type,
		Purpose:     keySpec.Purpose,
		PublicKey:   hex.EncodeToString(storedKey.PublicKey),
		CreatedAt:   storedKey.CreatedAt,
		UsageCount:  0,
		Algorithm:   "RSA-SHA256",
	}, nil
}

// ExportPublicKey exports a public key from the HSM in DER format.
func (r *SoftwareHSMRepository) ExportPublicKey(ctx context.Context, keyID string) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	key, err := r.getKey(keyID)
	if err != nil {
		return nil, err
	}

	return key.PublicKey, nil
}

// DeleteKey removes a key from the HSM.
func (r *SoftwareHSMRepository) DeleteKey(ctx context.Context, keyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.initialized {
		return ErrHSMUnavailable
	}

	// Prevent deletion of the root signing key.
	if keyID == "root-signing-key" {
		return errors.New("cannot delete root signing key")
	}

	if _, exists := r.keys[keyID]; !exists {
		return ErrKeyNotFound
	}

	delete(r.keys, keyID)
	return nil
}

// GetKeyInfo retrieves metadata about a key stored in the HSM.
func (r *SoftwareHSMRepository) GetKeyInfo(ctx context.Context, keyID string) (*models.HSMKeyInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	key, err := r.getKey(keyID)
	if err != nil {
		return nil, err
	}

	return &models.HSMKeyInfo{
		KeyID:       key.KeyID,
		KeyType:     key.KeyType,
		Purpose:     key.Purpose,
		PublicKey:   hex.EncodeToString(key.PublicKey),
		CreatedAt:   key.CreatedAt,
		LastUsed:    key.LastUsed,
		UsageCount:  key.UsageCount,
		Algorithm:   "RSA-SHA256",
	}, nil
}

// Encrypt encrypts data using the specified key (for symmetric encryption).
// This implementation uses RSA-OAEP for encryption.
func (r *SoftwareHSMRepository) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	key, err := r.getKey(keyID)
	if err != nil {
		return nil, err
	}

	// Parse the public key.
	publicKey, err := x509.ParsePKCS1PublicKey(key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Encrypt using RSA-OAEP with SHA-256.
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return ciphertext, nil
}

// Decrypt decrypts data using the specified key (for symmetric encryption).
// This implementation uses RSA-OAEP for decryption.
func (r *SoftwareHSMRepository) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, ErrHSMUnavailable
	}

	key, err := r.getKey(keyID)
	if err != nil {
		return nil, err
	}

	// Parse the private key.
	privateKey, err := parsePrivateKey(key.KeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Decrypt using RSA-OAEP with SHA-256.
	rsaKey := privateKey.(*rsa.PrivateKey)
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// HealthCheck verifies the HSM connection is healthy.
func (r *SoftwareHSMRepository) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return ErrHSMUnavailable
	}

	// Check that we have at least one key.
	if len(r.keys) == 0 {
		return errors.New("no keys found in HSM")
	}

	return nil
}

// Close releases all resources held by the HSM connection.
func (r *SoftwareHSMRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear all keys from memory.
	r.keys = make(map[string]*storedKey)
	r.initialized = false

	return nil
}

// getKey retrieves a key from storage.
func (r *SoftwareHSMRepository) getKey(keyID string) (*storedKey, error) {
	key, exists := r.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, keyID)
	}
	return key, nil
}

// parsePrivateKey parses a DER-encoded private key.
func parsePrivateKey(keyData []byte) (any, error) {
	// Try PKCS1 format first.
	key, err := x509.ParsePKCS1PrivateKey(keyData)
	if err == nil {
		return key, nil
	}

	// Try PKCS8 format.
	pkcs8Key, err := x509.ParsePKCS8PrivateKey(keyData)
	if err == nil {
		return pkcs8Key, nil
	}

	// Try PEM format.
	block, _ := pem.Decode(keyData)
	if block != nil {
		if block.Type == "RSA PRIVATE KEY" {
			return x509.ParsePKCS1PrivateKey(block.Bytes)
		}
		if block.Type == "PRIVATE KEY" {
			return x509.ParsePKCS8PrivateKey(block.Bytes)
		}
	}

	return nil, errors.New("failed to parse private key: unsupported format")
}

// generateKeyID generates a unique key identifier.
func generateKeyID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
