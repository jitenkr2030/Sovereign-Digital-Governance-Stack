package keymgmt

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// KeyType represents the type of encryption key
type KeyType string

const (
	KeyTypeSymmetric   KeyType = "symmetric"
	KeyTypeAsymmetric  KeyType = "asymmetric"
	KeyTypeSession     KeyType = "session"
	KeyTypeMaster      KeyType = "master"
)

// KeyAlgorithm represents the encryption algorithm
type KeyAlgorithm string

const (
	AlgorithmAES256GCM        KeyAlgorithm = "AES-256-GCM"
	AlgorithmChaCha20Poly1305 KeyAlgorithm = "ChaCha20-Poly1305"
	AlgorithmRSA2048          KeyAlgorithm = "RSA-2048"
	AlgorithmRSA4096          KeyAlgorithm = "RSA-4096"
	AlgorithmECDSAP256        KeyAlgorithm = "ECDSA-P256"
	AlgorithmECDSAP384        KeyAlgorithm = "ECDSA-P384"
	AlgorithmECDSAP521        KeyAlgorithm = "ECDSA-P521"
)

// Key represents an encryption key
type Key struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Type          KeyType       `json:"type"`
	Algorithm     KeyAlgorithm  `json:"algorithm"`
	Material      []byte        `json:"material,omitempty"`
	PublicKey     []byte        `json:"public_key,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	ExpiresAt     time.Time     `json:"expires_at,omitempty"`
	RotatedAt     time.Time     `json:"rotated_at,omitempty"`
	Version       int           `json:"version"`
	Status        KeyStatus     `json:"status"`
	Usage         KeyUsage      `json:"usage"`
	Metadata      KeyMetadata   `json:"metadata"`
	PreviousKeyID string        `json:"previous_key_id,omitempty"`
}

// KeyStatus represents the status of a key
type KeyStatus string

const (
	KeyStatusActive   KeyStatus = "active"
	KeyStatusRotated  KeyStatus = "rotated"
	KeyStatusExpired  KeyStatus = "expired"
	KeyStatusRevoked  KeyStatus = "revoked"
	KeyStatusPending  KeyStatus = "pending"
	KeyStatusDestroyed KeyStatus = "destroyed"
)

// KeyUsage represents the allowed usage of a key
type KeyUsage string

const (
	KeyUsageEncrypt    KeyUsage = "encrypt"
	KeyUsageDecrypt    KeyUsage = "decrypt"
	KeyUsageSign       KeyUsage = "sign"
	KeyUsageVerify     KeyUsage = "verify"
	KeyUsageKeyWrap    KeyUsage = "key_wrap"
	KeyUsageKeyUnwrap  KeyUsage = "key_unwrap"
	KeyUsageDerive     KeyUsage = "derive"
)

// KeyMetadata contains additional key information
type KeyMetadata struct {
	Description   string            `json:"description,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	ACL           []string          `json:"acl,omitempty"`
	RotationCount int               `json:"rotation_count"`
	ParentKeyID   string            `json:"parent_key_id,omitempty"`
	Custodian     string            `json:"custodian,omitempty"`
}

// KeyGenerationConfig contains configuration for key generation
type KeyGenerationConfig struct {
	Type       KeyType
	Algorithm  KeyAlgorithm
	KeySize    int
	Expiration time.Duration
	Usage      []KeyUsage
	Metadata   KeyMetadata
}

// KeyRotationConfig contains configuration for key rotation
type KeyRotationConfig struct {
	PreserveOldKey    bool
	ArchiveOldKey     bool
	MigrationPeriod   time.Duration
	NotifyRecipients  []string
}

// Service provides key management capabilities
type Service struct {
	logger       *zap.Logger
	config       *Config
	keys         map[string]*Key
	keyVersions  map[string][]string
	mu           sync.RWMutex
	storage      KeyStorage
	auditLogger  *AuditLogger
}

// Global service instance
var (
	defaultKeyService *Service
	keyServiceMu      sync.RWMutex
)

// Config for key management service
type Config struct {
	StorageBackend    string
	MasterKeyID       string
	RotationPeriod    int
	MaxKeyVersions    int
	HSMEnabled        bool
	VaultEnabled      bool
}

// KeyStorage interface for key persistence
type KeyStorage interface {
	StoreKey(ctx context.Context, key *Key) error
	GetKey(ctx context.Context, keyID string) (*Key, error)
	DeleteKey(ctx context.Context, keyID string) error
	ListKeys(ctx context.Context) ([]*Key, error)
	UpdateKey(ctx context.Context, key *Key) error
}

// AuditLogger logs key operations
type AuditLogger struct {
	logger *zap.Logger
}

// Initialize initializes the key management service
func Initialize(logger *zap.Logger, cfg interface{}) error {
	keyServiceMu.Lock()
	defer keyServiceMu.Unlock()

	// Extract configuration
	config := &Config{
		StorageBackend: "database",
		MasterKeyID:    "master-key-v1",
		RotationPeriod: 90,
		MaxKeyVersions: 10,
	}

	defaultKeyService = &Service{
		logger:      logger,
		config:      config,
		keys:        make(map[string]*Key),
		keyVersions: make(map[string][]string),
		storage:     NewInMemoryStorage(),
		auditLogger: &AuditLogger{logger: logger},
	}

	logger.Info("Key management service initialized",
		zap.String("storage_backend", config.StorageBackend),
		zap.Int("rotation_period_days", config.RotationPeriod))

	return nil
}

// GetService returns the default key management service
func GetService() *Service {
	keyServiceMu.RLock()
	defer keyServiceMu.RUnlock()
	return defaultKeyService
}

// GenerateKey generates a new encryption key
func (s *Service) GenerateKey(ctx context.Context, config KeyGenerationConfig) (*Key, error) {
	s.logger.Info("Generating key",
		zap.String("type", string(config.Type)),
		zap.String("algorithm", string(config.Algorithm)))

	var keyMaterial []byte
	var publicKey []byte
	var err error

	switch config.Type {
	case KeyTypeSymmetric:
		keyMaterial, err = s.generateSymmetricKey(config)
	case KeyTypeAsymmetric:
		keyMaterial, publicKey, err = s.generateAsymmetricKey(config)
	case KeyTypeMaster:
		keyMaterial, err = s.generateMasterKey(config)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}

	// Create key object
	key := &Key{
		ID:        uuid.New().String(),
		Name:      config.Metadata.Description,
		Type:      config.Type,
		Algorithm: config.Algorithm,
		Material:  keyMaterial,
		PublicKey: publicKey,
		CreatedAt: time.Now(),
		Version:   1,
		Status:    KeyStatusActive,
		Usage:     config.Usage[0], // Primary usage
		Metadata:  config.Metadata,
	}

	// Set expiration
	if config.Expiration > 0 {
		key.ExpiresAt = time.Now().Add(config.Expiration)
	}

	// Store key
	if err := s.storage.StoreKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to store key: %w", err)
	}

	// Cache key
	s.mu.Lock()
	s.keys[key.ID] = key
	s.keyVersions[key.ID] = []string{key.ID}
	s.mu.Unlock()

	// Audit log
	s.auditLogger.LogKeyGeneration(key)

	s.logger.Info("Key generated",
		zap.String("key_id", key.ID),
		zap.String("algorithm", string(key.Algorithm)))

	return key, nil
}

// GetKey retrieves a key by ID
func (s *Service) GetKey(ctx context.Context, keyID string) (*Key, error) {
	// Check cache first
	s.mu.RLock()
	key, ok := s.keys[keyID]
	s.mu.RUnlock()

	if ok {
		s.auditLogger.LogKeyAccess(key, "get")
		return key, nil
	}

	// Fetch from storage
	key, err := s.storage.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	// Cache key
	s.mu.Lock()
	s.keys[key.ID] = key
	s.mu.Unlock()

	s.auditLogger.LogKeyAccess(key, "get")
	return key, nil
}

// ListKeys lists all keys
func (s *Service) ListKeys(ctx context.Context) ([]*Key, error) {
	return s.storage.ListKeys(ctx)
}

// ListActiveKeys lists all active keys
func (s *Service) ListActiveKeys(ctx context.Context) ([]*Key, error) {
	keys, err := s.storage.ListKeys(ctx)
	if err != nil {
		return nil, err
	}

	activeKeys := make([]*Key, 0)
	for _, key := range keys {
		if key.Status == KeyStatusActive {
			activeKeys = append(activeKeys, key)
		}
	}

	return activeKeys, nil
}

// RotateKey rotates a key, creating a new version
func (s *Service) RotateKey(ctx context.Context, keyID string, config KeyRotationConfig) (*Key, error) {
	s.logger.Info("Rotating key",
		zap.String("key_id", keyID))

	// Get existing key
	oldKey, err := s.GetKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	if oldKey.Status != KeyStatusActive {
		return nil, errors.New("can only rotate active keys")
	}

	// Generate new key
	newConfig := KeyGenerationConfig{
		Type:       oldKey.Type,
		Algorithm:  oldKey.Algorithm,
		Expiration: time.Duration(s.config.RotationPeriod) * 24 * time.Hour,
		Usage:      []KeyUsage{oldKey.Usage},
		Metadata: KeyMetadata{
			Description:   fmt.Sprintf("Rotated from %s", oldKey.ID),
			Labels:        oldKey.Metadata.Labels,
			RotationCount: oldKey.Metadata.RotationCount + 1,
		},
	}

	newKey, err := s.GenerateKey(ctx, newConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key: %w", err)
	}

	// Link keys
	newKey.PreviousKeyID = oldKey.ID
	oldKey.RotatedAt = time.Now()
	oldKey.Status = KeyStatusRotated

	// Update old key
	if err := s.storage.UpdateKey(ctx, oldKey); err != nil {
		return nil, fmt.Errorf("failed to update old key: %w", err)
	}

	// Update new key
	if err := s.storage.UpdateKey(ctx, newKey); err != nil {
		return nil, fmt.Errorf("failed to update new key: %w", err)
	}

	// Update cache
	s.mu.Lock()
	s.keys[newKey.ID] = newKey
	s.keys[oldKey.ID] = oldKey
	s.keyVersions[keyID] = append(s.keyVersions[keyID], newKey.ID)
	s.mu.Unlock()

	// Audit log
	s.auditLogger.LogKeyRotation(oldKey, newKey)

	s.logger.Info("Key rotated",
		zap.String("old_key_id", oldKey.ID),
		zap.String("new_key_id", newKey.ID))

	return newKey, nil
}

// RevokeKey revokes a key
func (s *Service) RevokeKey(ctx context.Context, keyID string, reason string) error {
	s.logger.Info("Revoking key",
		zap.String("key_id", keyID),
		zap.String("reason", reason))

	key, err := s.GetKey(ctx, keyID)
	if err != nil {
		return err
	}

	key.Status = KeyStatusRevoked
	key.Metadata.Description = reason

	if err := s.storage.UpdateKey(ctx, key); err != nil {
		return fmt.Errorf("failed to update key: %w", err)
	}

	// Update cache
	s.mu.Lock()
	s.keys[keyID] = key
	s.mu.Unlock()

	s.auditLogger.LogKeyRevocation(key, reason)

	return nil
}

// DestroyKey permanently destroys a key
func (s *Service) DestroyKey(ctx context.Context, keyID string) error {
	s.logger.Info("Destroying key",
		zap.String("key_id", keyID))

	// Get key
	key, err := s.GetKey(ctx, keyID)
	if err != nil {
		return err
	}

	// Verify key is revoked first
	if key.Status != KeyStatusRevoked {
		return errors.New("key must be revoked before destruction")
	}

	// Delete from storage
	if err := s.storage.DeleteKey(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	// Remove from cache
	s.mu.Lock()
	delete(s.keys, keyID)
	delete(s.keyVersions, keyID)
	s.mu.Unlock()

	s.auditLogger.LogKeyDestruction(key)

	return nil
}

// ImportKey imports an existing key
func (s *Service) ImportKey(ctx context.Context, keyMaterial []byte, config KeyGenerationConfig) (*Key, error) {
	s.logger.Info("Importing key")

	key := &Key{
		ID:        uuid.New().String(),
		Name:      config.Metadata.Description,
		Type:      config.Type,
		Algorithm: config.Algorithm,
		Material:  keyMaterial,
		CreatedAt: time.Now(),
		Version:   1,
		Status:    KeyStatusActive,
		Usage:     config.Usage[0],
		Metadata:  config.Metadata,
	}

	if config.Expiration > 0 {
		key.ExpiresAt = time.Now().Add(config.Expiration)
	}

	if err := s.storage.StoreKey(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to store key: %w", err)
	}

	s.mu.Lock()
	s.keys[key.ID] = key
	s.mu.Unlock()

	s.auditLogger.LogKeyImport(key)

	return key, nil
}

// ExportKey exports a key (in encrypted form)
func (s *Service) ExportKey(ctx context.Context, keyID string, wrappingKeyID string) ([]byte, error) {
	s.logger.Info("Exporting key",
		zap.String("key_id", keyID),
		zap.String("wrapping_key_id", wrappingKeyID))

	key, err := s.GetKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// In production, would use proper key wrapping (e.g., AES Key Wrap)
	wrappingKey, err := s.GetKey(ctx, wrappingKeyID)
	if err != nil {
		return nil, fmt.Errorf("wrapping key not found: %s", wrappingKeyID)
	}

	// Simple XOR for demonstration - in production use proper key wrapping
	wrappedKey := make([]byte, len(key.Material))
	for i := range key.Material {
		wrappedKey[i] = key.Material[i] ^ wrappingKey.Material[i%len(wrappingKey.Material)]
	}

	// Encode as PEM
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "ENCRYPTED KEY",
		Bytes: wrappedKey,
		Headers: map[string]string{
			"Key-ID":     keyID,
			"Algorithm":  string(key.Algorithm),
			"Wrapped-By": wrappingKeyID,
		},
	})

	s.auditLogger.LogKeyExport(key)

	return pemData, nil
}

// GetKeyVersions returns all versions of a key
func (s *Service) GetKeyVersions(keyID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keyVersions[keyID]
}

// generateSymmetricKey generates a symmetric key
func (s *Service) generateSymmetricKey(config KeyGenerationConfig) ([]byte, error) {
	keySize := 32 // 256 bits for AES-256

	switch config.Algorithm {
	case AlgorithmAES256GCM:
		keySize = 32
	case AlgorithmChaCha20Poly1305:
		keySize = 32
	default:
		keySize = config.KeySize
	}

	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return key, nil
}

// generateAsymmetricKey generates an asymmetric key pair
func (s *Service) generateAsymmetricKey(config KeyGenerationConfig) ([]byte, []byte, error) {
	var bits int

	switch config.Algorithm {
	case AlgorithmRSA2048:
		bits = 2048
	case AlgorithmRSA4096:
		bits = 4096
	default:
		bits = 2048
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	return privateKeyBytes, publicKeyBytes, nil
}

// generateMasterKey generates a master key
func (s *Service) generateMasterKey(config KeyGenerationConfig) ([]byte, error) {
	// Master keys are 256 bits
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	return key, nil
}

// InMemoryStorage implements KeyStorage with in-memory storage
type InMemoryStorage struct {
	mu   sync.RWMutex
	keys map[string]*Key
}

// NewInMemoryStorage creates a new in-memory key storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		keys: make(map[string]*Key),
	}
}

// StoreKey stores a key
func (s *InMemoryStorage) StoreKey(ctx context.Context, key *Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[key.ID] = key
	return nil
}

// GetKey retrieves a key
func (s *InMemoryStorage) GetKey(ctx context.Context, keyID string) (*Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key, ok := s.keys[keyID]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	return key, nil
}

// DeleteKey deletes a key
func (s *InMemoryStorage) DeleteKey(ctx context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keys, keyID)
	return nil
}

// ListKeys lists all keys
func (s *InMemoryStorage) ListKeys(ctx context.Context) ([]*Key, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*Key, 0, len(s.keys))
	for _, key := range s.keys {
		keys = append(keys, key)
	}
	return keys, nil
}

// UpdateKey updates a key
func (s *InMemoryStorage) UpdateKey(ctx context.Context, key *Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[key.ID] = key
	return nil
}

// AuditLogger methods

// LogKeyGeneration logs key generation
func (a *AuditLogger) LogKeyGeneration(key *Key) {
	a.logger.Info("Key generated",
		zap.String("key_id", key.ID),
		zap.String("key_type", string(key.Type)),
		zap.String("algorithm", string(key.Algorithm)),
		zap.Time("created_at", key.CreatedAt))
}

// LogKeyAccess logs key access
func (a *AuditLogger) LogKeyAccess(key *Key, operation string) {
	a.logger.Debug("Key accessed",
		zap.String("key_id", key.ID),
		zap.String("operation", operation))
}

// LogKeyRotation logs key rotation
func (a *AuditLogger) LogKeyRotation(oldKey, newKey *Key) {
	a.logger.Info("Key rotated",
		zap.String("old_key_id", oldKey.ID),
		zap.String("new_key_id", newKey.ID),
		zap.Int("old_version", oldKey.Version),
		zap.Int("new_version", newKey.Version))
}

// LogKeyRevocation logs key revocation
func (a *AuditLogger) LogKeyRevocation(key *Key, reason string) {
	a.logger.Info("Key revoked",
		zap.String("key_id", key.ID),
		zap.String("reason", reason))
}

// LogKeyDestruction logs key destruction
func (a *AuditLogger) LogKeyDestruction(key *Key) {
	a.logger.Info("Key destroyed",
		zap.String("key_id", key.ID))
}

// LogKeyImport logs key import
func (a *AuditLogger) LogKeyImport(key *Key) {
	a.logger.Info("Key imported",
		zap.String("key_id", key.ID),
		zap.String("key_type", string(key.Type)),
		zap.String("algorithm", string(key.Algorithm)))
}

// LogKeyExport logs key export
func (a *AuditLogger) LogKeyExport(key *Key) {
	a.logger.Info("Key exported",
		zap.String("key_id", key.ID))
}

// Helper functions

// KeyToPEM converts a key to PEM format
func KeyToPEM(key *Key) ([]byte, error) {
	var blockType string
	var data []byte

	switch key.Type {
	case KeyTypeSymmetric:
		blockType = "SYMMETRIC KEY"
		data = key.Material
	case KeyTypeAsymmetric:
		blockType = "PRIVATE KEY"
		data = key.Material
	case KeyTypeMaster:
		blockType = "MASTER KEY"
		data = key.Material
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.Type)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  blockType,
		Bytes: data,
		Headers: map[string]string{
			"Key-ID":    key.ID,
			"Algorithm": string(key.Algorithm),
			"Created":   key.CreatedAt.Format(time.RFC3339),
		},
	}), nil
}

// PEMToKey converts PEM to Key
func PEMToKey(pemData []byte) (*Key, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM")
	}

	return &Key{
		Material: block.Bytes,
		Metadata: KeyMetadata{
			Description: block.Headers["Key-Description"],
			Labels:      make(map[string]string),
		},
	}, nil
}

// DeriveKey derives a new key from an existing key using HKDF
func (s *Service) DeriveKey(ctx context.Context, keyID string, salt []byte, info []byte, length int) ([]byte, error) {
	key, err := s.GetKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// Use crypto.HKDF
	hkdf := crypto.HKDF

	derivedKey := make([]byte, length)
	n, err := hkdf(crypto.SHA256, key.Material, salt, info, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	if n != length {
		return nil, errors.New("derived key length mismatch")
	}

	return derivedKey, nil
}

// HashKeyID creates a hashed version of a key ID
func HashKeyID(keyID string) string {
	hash := crypto.SHA256.New()
	hash.Write([]byte(keyID))
	return hex.EncodeToString(hash.Sum(nil))
}

// EncodeKeyID encodes a key ID to base64
func EncodeKeyID(keyID string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(keyID))
}

// DecodeKeyID decodes a base64 key ID
func DecodeKeyID(encoded string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode key ID: %w", err)
	}
	return string(data), nil
}
