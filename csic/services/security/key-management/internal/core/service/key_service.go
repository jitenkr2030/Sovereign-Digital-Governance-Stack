package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/key-management/internal/config"
	"github.com/csic-platform/services/security/key-management/internal/core/domain"
	"github.com/csic-platform/services/security/key-management/internal/core/ports"
	"github.com/google/uuid"
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrKeyAlreadyExists = errors.New("key with this alias already exists")
	ErrKeyRevoked       = errors.New("key has been revoked")
	ErrInvalidAlgorithm = errors.New("invalid key algorithm")
	ErrEncryptionFailed = errors.New("encryption operation failed")
	ErrDecryptionFailed = errors.New("decryption operation failed")
)

// KeyServiceImpl implements the KeyService interface
type KeyServiceImpl struct {
	repo      ports.KeyRepository
	cache     ports.CacheRepository
	producer  ports.MessageProducer
	crypto    ports.CryptoProvider
	audit     ports.AuditService
	masterKey []byte
	cfg       *config.SecurityConfig
}

// NewKeyService creates a new key service instance
func NewKeyService(
	repo ports.KeyRepository,
	cache ports.CacheRepository,
	producer ports.MessageProducer,
	crypto ports.CryptoProvider,
	audit ports.AuditService,
	masterKey []byte,
	cfg *config.SecurityConfig,
) *KeyServiceImpl {
	return &KeyServiceImpl{
		repo:      repo,
		cache:     cache,
		producer:  producer,
		crypto:    crypto,
		audit:     audit,
		masterKey: masterKey,
		cfg:       cfg,
	}
}

// CreateKey creates a new cryptographic key
func (s *KeyServiceImpl) CreateKey(ctx context.Context, req *domain.CreateKeyRequest, actorID string) (*domain.Key, error) {
	// Validate algorithm
	if !req.Algorithm.Validate() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAlgorithm, req.Algorithm)
	}

	// Check if key with alias already exists
	existing, err := s.repo.GetKeyByAlias(ctx, req.Alias)
	if err == nil && existing != nil {
		return nil, ErrKeyAlreadyExists
	}

	// Generate key material
	keyMaterial, err := s.crypto.GenerateSymmetricKey(req.Algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}

	// Create key entity
	key := &domain.Key{
		ID:             uuid.New().String(),
		Alias:          req.Alias,
		Algorithm:      req.Algorithm,
		CurrentVersion: 1,
		Status:         domain.KeyStatusActive,
		Usage:          req.Usage,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Description:    req.Description,
		Metadata:       req.Metadata,
		ExpiresAt:      req.ExpiresAt,
	}

	// Create first key version with encrypted material
	iv, _ := s.crypto.GenerateIV(req.Algorithm)
	encryptedMaterial, _ := s.crypto.Encrypt(req.Algorithm, s.masterKey, keyMaterial, iv)

	version := &domain.KeyVersion{
		ID:               uuid.New().String(),
		KeyID:            key.ID,
		Version:          1,
		EncryptedMaterial: base64.StdEncoding.EncodeToString(encryptedMaterial),
		IV:               base64.StdEncoding.EncodeToString(iv),
		CreatedAt:        time.Now(),
	}

	// Persist to database
	if err := s.repo.CreateKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	if err := s.repo.CreateKeyVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create key version: %w", err)
	}

	// Audit log
	s.logOperation(ctx, key.ID, "CREATE", actorID, "", true, nil)

	return key, nil
}

// GetKey retrieves a key by ID
func (s *KeyServiceImpl) GetKey(ctx context.Context, id string, actorID string) (*domain.Key, error) {
	// Try cache first
	if cached, err := s.cache.GetKeyMetadata(ctx, id); err == nil {
		s.logOperation(ctx, id, "ACCESS", actorID, "", true, nil)
		return cached, nil
	}

	// Fetch from database
	key, err := s.repo.GetKey(ctx, id)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Cache the result
	s.cache.SetKeyMetadata(ctx, id, key, int(s.cfg.Rotation.DefaultPeriodDays))

	s.logOperation(ctx, id, "ACCESS", actorID, "", true, nil)

	return key, nil
}

// ListKeys retrieves a paginated list of keys
func (s *KeyServiceImpl) ListKeys(ctx context.Context, page, pageSize int, actorID string) (*domain.KeyListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	keys, total, err := s.repo.ListKeys(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	return domain.NewKeyListResponse(keys, total, page, pageSize), nil
}

// RotateKey rotates a key to a new version
func (s *KeyServiceImpl) RotateKey(ctx context.Context, id string, req *domain.RotateKeyRequest, actorID string) (*domain.Key, error) {
	key, err := s.GetKey(ctx, id, actorID)
	if err != nil {
		return nil, err
	}

	if key.Status == domain.KeyStatusRevoked {
		return nil, ErrKeyRevoked
	}

	// Generate new key material
	newKeyMaterial, err := s.crypto.GenerateSymmetricKey(key.Algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key material: %w", err)
	}

	// Create new version
	iv, _ := s.crypto.GenerateIV(key.Algorithm)
	encryptedMaterial, _ := s.crypto.Encrypt(key.Algorithm, s.masterKey, newKeyMaterial, iv)

	newVersion := &domain.KeyVersion{
		ID:               uuid.New().String(),
		KeyID:            key.ID,
		Version:          key.CurrentVersion + 1,
		EncryptedMaterial: base64.StdEncoding.EncodeToString(encryptedMaterial),
		IV:               base64.StdEncoding.EncodeToString(iv),
		CreatedAt:        time.Now(),
	}

	// Update key status
	now := time.Now()
	key.CurrentVersion = newVersion.Version
	key.Status = domain.KeyStatusActive
	key.RotatedAt = &now
	key.UpdatedAt = now

	// Persist changes
	if err := s.repo.CreateKeyVersion(ctx, newVersion); err != nil {
		return nil, fmt.Errorf("failed to create new key version: %w", err)
	}

	if err := s.repo.UpdateKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to update key: %w", err)
	}

	// Invalidate cache
	s.cache.InvalidateKeyCache(ctx, id)

	// Audit log
	s.logOperation(ctx, id, "ROTATE", actorID, "", true, map[string]interface{}{
		"reason":   req.Reason,
		"new_version": newVersion.Version,
	})

	return key, nil
}

// RevokeKey revokes a key
func (s *KeyServiceImpl) RevokeKey(ctx context.Context, id string, req *domain.RevokeKeyRequest, actorID string) error {
	key, err := s.GetKey(ctx, id, actorID)
	if err != nil {
		return err
	}

	if key.Status == domain.KeyStatusRevoked {
		return errors.New("key is already revoked")
	}

	now := time.Now()
	key.Status = domain.KeyStatusRevoked
	key.UpdatedAt = now

	if err := s.repo.UpdateKey(ctx, key); err != nil {
		return fmt.Errorf("failed to revoke key: %w", err)
	}

	// Invalidate cache
	s.cache.InvalidateKeyCache(ctx, id)

	// Audit log
	s.logOperation(ctx, id, "REVOKE", actorID, "", true, map[string]interface{}{
		"reason": req.Reason,
	})

	return nil
}

// GenerateDataKey generates a data key for envelope encryption
func (s *KeyServiceImpl) GenerateDataKey(ctx context.Context, req *domain.GenerateDataKeyRequest, actorID string) (*domain.DataKeyResponse, error) {
	key, err := s.GetKey(ctx, req.KeyAlias, actorID)
	if err != nil {
		return nil, err
	}

	if key.Status != domain.KeyStatusActive {
		return nil, fmt.Errorf("key is not active: %s", key.Status)
	}

	// Get master key for key encryption
	masterKeyMaterial := s.masterKey

	// Determine key size
	keySize := req.KeySize
	if keySize == 0 {
		keySize = key.Algorithm.GetKeySize()
	}

	// Generate data key
	dataKey, err := s.crypto.GenerateSymmetricKey(domain.AlgorithmAES256GCM)
	if err != nil {
		return nil, fmt.Errorf("failed to generate data key: %w", err)
	}

	// Wrap (encrypt) the data key with the master key
	wrappedKey, iv, err := s.crypto.WrapKey(domain.AlgorithmAES256GCM, masterKeyMaterial, dataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap data key: %w", err)
	}

	// Audit log
	s.logOperation(ctx, key.ID, "GENERATE_DATA_KEY", actorID, "", true, map[string]interface{}{
		"key_size": keySize,
	})

	return &domain.DataKeyResponse{
		KeyID:        key.ID,
		PlaintextKey: base64.StdEncoding.EncodeToString(dataKey),
		EncryptedKey: base64.StdEncoding.EncodeToString(wrappedKey),
		Version:      key.CurrentVersion,
	}, nil
}

// EncryptData encrypts data using a key
func (s *KeyServiceImpl) EncryptData(ctx context.Context, req *domain.EncryptRequest, actorID string) (*domain.EncryptDecryptResponse, error) {
	key, err := s.GetKey(ctx, req.KeyID, actorID)
	if err != nil {
		return nil, err
	}

	if key.Status != domain.KeyStatusActive {
		return nil, fmt.Errorf("key is not active: %s", key.Status)
	}

	// Get the latest key version
	version, err := s.repo.GetLatestKeyVersion(ctx, req.KeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version: %w", err)
	}

	// Decrypt key material
	keyMaterial, err := s.crypto.Decrypt(
		key.Algorithm,
		s.masterKey,
		[]byte(version.EncryptedMaterial),
		[]byte(version.IV),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key material: %w", err)
	}

	// Generate IV if not provided
	iv, _ := s.crypto.GenerateIV(key.Algorithm)

	// Decrypt the plaintext if it's base64 encoded
	plaintext, err := base64.StdEncoding.DecodeString(req.Plaintext)
	if err != nil {
		plaintext = []byte(req.Plaintext)
	}

	// Encrypt the data
	ciphertext, err := s.crypto.Encrypt(key.Algorithm, keyMaterial, plaintext, iv)
	if err != nil {
		s.logOperation(ctx, req.KeyID, "ENCRYPT", actorID, "", false, map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	s.logOperation(ctx, req.KeyID, "ENCRYPT", actorID, "", true, nil)

	return &domain.EncryptDecryptResponse{
		Output:  base64.StdEncoding.EncodeToString(ciphertext),
		Version: key.CurrentVersion,
	}, nil
}

// DecryptData decrypts data using a key
func (s *KeyServiceImpl) DecryptData(ctx context.Context, req *domain.DecryptRequest, actorID string) (*domain.EncryptDecryptResponse, error) {
	key, err := s.GetKey(ctx, req.KeyID, actorID)
	if err != nil {
		return nil, err
	}

	// Get the appropriate key version
	version, err := s.repo.GetKeyVersion(ctx, req.KeyID, key.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version: %w", err)
	}

	// Decrypt key material
	keyMaterial, err := s.crypto.Decrypt(
		key.Algorithm,
		s.masterKey,
		[]byte(version.EncryptedMaterial),
		[]byte(version.IV),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key material: %w", err)
	}

	// Decrypt the ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(req.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext encoding: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(req.IV)
	if err != nil {
		return nil, fmt.Errorf("invalid IV encoding: %w", err)
	}

	plaintext, err := s.crypto.Decrypt(key.Algorithm, keyMaterial, ciphertext, iv)
	if err != nil {
		s.logOperation(ctx, req.KeyID, "DECRYPT", actorID, "", false, map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	s.logOperation(ctx, req.KeyID, "DECRYPT", actorID, "", true, nil)

	return &domain.EncryptDecryptResponse{
		Output:  base64.StdEncoding.EncodeToString(plaintext),
		Version: key.CurrentVersion,
	}, nil
}

// UpdateKeyMetadata updates the metadata for a key
func (s *KeyServiceImpl) UpdateKeyMetadata(ctx context.Context, id string, metadata map[string]string, actorID string) error {
	key, err := s.GetKey(ctx, id, actorID)
	if err != nil {
		return err
	}

	key.Metadata = metadata
	key.UpdatedAt = time.Now()

	if err := s.repo.UpdateKey(ctx, key); err != nil {
		return fmt.Errorf("failed to update key metadata: %w", err)
	}

	s.cache.InvalidateKeyCache(ctx, id)

	return nil
}

// GetKeyAuditLog retrieves the audit log for a key
func (s *KeyServiceImpl) GetKeyAuditLog(ctx context.Context, keyID string, limit int, actorID string) ([]*domain.KeyOperation, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.GetKeyAuditLog(ctx, keyID, limit)
}

// GetKeyVersions retrieves all versions of a key
func (s *KeyServiceImpl) GetKeyVersions(ctx context.Context, keyID string, actorID string) ([]*domain.KeyVersion, error) {
	_, err := s.GetKey(ctx, keyID, actorID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetAllKeyVersions(ctx, keyID)
}

// HealthCheck performs a health check
func (s *KeyServiceImpl) HealthCheck(ctx context.Context) error {
	// Verify master key is available
	if len(s.masterKey) == 0 {
		return errors.New("master key not configured")
	}

	// Verify database connectivity
	_, err := s.repo.ListKeys(ctx, 1, 1)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// logOperation logs a key operation to the audit trail
func (s *KeyServiceImpl) logOperation(ctx context.Context, keyID, operation, actorID, serviceID string, success bool, metadata map[string]interface{}) {
	operationLog := &domain.KeyOperation{
		ID:        uuid.New().String(),
		KeyID:     keyID,
		Operation: operation,
		ActorID:   actorID,
		ServiceID: serviceID,
		Success:   success,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	s.repo.AuditKeyOperation(ctx, operationLog)

	if s.producer != nil {
		s.producer.PublishKeyOperation(ctx, operationLog)
	}
}
