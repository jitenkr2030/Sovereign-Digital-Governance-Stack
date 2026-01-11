package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// KeyManagementService handles HSM operations for the platform
type KeyManagementService struct {
	provider CryptoProvider
	keyCache map[string]*CachedKey
}

// CachedKey holds key metadata in memory
type CachedKey struct {
	KeyInfo      *KeyInfo
	LastAccessed time.Time
	UsageCount   int
}

// NewKeyManagementService creates a new key management service
func NewKeyManagementService(provider CryptoProvider) *KeyManagementService {
	return &KeyManagementService{
		provider: provider,
		keyCache: make(map[string]*CachedKey),
	}
}

// GenerateSigningKey generates a new signing key
func (s *KeyManagementService) GenerateSigningKey(ctx context.Context, keyType KeyType, keySize int) (*KeyInfo, error) {
	return s.provider.GenerateKey(keyType, keySize)
}

// SignTransaction signs a transaction payload
func (s *KeyManagementService) SignTransaction(ctx context.Context, keyID string, transactionData []byte) (*SignatureResult, error) {
	s.updateCache(keyID)
	
	signature, err := s.provider.Sign(keyID, transactionData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}
	
	return &SignatureResult{
		Signature:  signature,
		Algorithm:  crypto.SHA256,
		KeyID:      keyID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// VerifySignature verifies a transaction signature
func (s *KeyManagementService) VerifySignature(ctx context.Context, keyID string, data, signature []byte) (bool, error) {
	return s.provider.Verify(keyID, data, signature), nil
}

// GenerateKeyPair generates a new RSA key pair for an entity
func (s *KeyManagementService) GenerateKeyPair(ctx context.Context, entityID uuid.UUID, keySize int) (*KeyInfo, error) {
	keyInfo, err := s.provider.GenerateKey(KeyTypeRSA, keySize)
	if err != nil {
		return nil, err
	}
	
	// Cache the key
	s.keyCache[keyInfo.ID] = &CachedKey{
		KeyInfo:      keyInfo,
		LastAccessed: time.Now(),
		UsageCount:   0,
	}
	
	// Store entity association
	s.storeKeyAssociation(entityID, keyInfo.ID)
	
	return keyInfo, nil
}

// GetEntityKeys retrieves all keys for an entity
func (s *KeyManagementService) GetEntityKeys(ctx context.Context, entityID uuid.UUID) ([]*KeyInfo, error) {
	// In production, this would query a database
	return []*KeyInfo{}, nil
}

// RotateKey rotates a key for an entity
func (s *KeyManagementService) RotateKey(ctx context.Context, entityID uuid.UUID, keyID string) (*KeyInfo, error) {
	// Generate new key
	newKey, err := s.GenerateKeyPair(ctx, entityID, 4096)
	if err != nil {
		return nil, err
	}
	
	// Mark old key for archival (don't delete immediately)
	s.archiveKey(keyID)
	
	return newKey, nil
}

// ExportPublicKey exports the public key for sharing
func (s *KeyManagementService) ExportPublicKey(ctx context.Context, keyID string) ([]byte, error) {
	return s.provider.ExportPublicKey(keyID)
}

// EncryptData encrypts sensitive data
func (s *KeyManagementService) EncryptData(ctx context.Context, keyID string, data []byte) (*EncryptionResult, error) {
	encrypted, err := s.provider.Encrypt(keyID, data)
	if err != nil {
		return nil, err
	}
	
	return &EncryptionResult{
		Ciphertext: encrypted,
		Algorithm:  "AES-256-GCM",
		KeyID:      keyID,
	}, nil
}

// DecryptData decrypts sensitive data
func (s *KeyManagementService) DecryptData(ctx context.Context, keyID string, encryptedData []byte) ([]byte, error) {
	return s.provider.Decrypt(keyID, encryptedData)
}

// HealthCheck checks HSM connectivity
func (s *KeyManagementService) HealthCheck(ctx context.Context) error {
	// Try to generate a test key
	_, err := s.provider.GenerateKey(KeyTypeRSA, 2048)
	if err != nil {
		return fmt.Errorf("HSM health check failed: %w", err)
	}
	return nil
}

// KeyCeremony generates keys with enhanced security controls
func (s *KeyManagementService) KeyCeremony(ctx context.Context, entityID uuid.UUID, requireQuorum bool) (*KeyInfo, error) {
	fmt.Println("Starting key ceremony...")
	
	// In production, this would implement M-of-N quorum
	// For now, generate a standard key
	keyInfo, err := s.GenerateKeyPair(ctx, entityID, 4096)
	if err != nil {
		return nil, err
	}
	
	// Generate audit record
	auditRecord := map[string]interface{}{
		"event":           "key_ceremony",
		"entity_id":       entityID.String(),
		"key_id":          keyInfo.ID,
		"key_type":        string(keyInfo.Type),
		"key_size":        keyInfo.Size,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"quorum_required": requireQuorum,
	}
	
	fmt.Printf("Key ceremony completed: %+v\n", auditRecord)
	
	return keyInfo, nil
}

// Helper functions

func (s *KeyManagementService) updateCache(keyID string) {
	if cached, exists := s.keyCache[keyID]; exists {
		cached.LastAccessed = time.Now()
		cached.UsageCount++
	}
}

func (s *KeyManagementService) storeKeyAssociation(entityID uuid.UUID, keyID string) {
	// In production, this would store to database
	_ = entityID
	_ = keyID
}

func (s *KeyManagementService) archiveKey(keyID string) {
	// In production, this would mark key as archived in database
	_ = keyID
}

// KeyInfoToPEM converts a KeyInfo to PEM format
func KeyInfoToPEM(keyInfo *KeyInfo, privateKey crypto.PrivateKey) ([]byte, error) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}
	
	return pem.EncodeToMemory(block), nil
}

// GenerateNonce generates a cryptographically secure nonce
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// Signer wraps crypto.Signer with additional metadata
type Signer struct {
	KeyID    string
	Signer   crypto.Signer
	CertPath string
}

// NewSigner creates a new Signer
func NewSigner(keyID string, signer crypto.Signer) *Signer {
	return &Signer{
		KeyID:  keyID,
		Signer: signer,
	}
}

// Sign implements crypto.Signer
func (s *Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return s.Signer.Sign(rand, digest, opts)
}

// CreateTestKeys creates test keys for development
func CreateTestKeys() error {
	// Generate test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate test key: %w", err)
	}
	
	// Save to file
	pemData := PEMEncode(privateKey)
	if err := os.WriteFile("test-key.pem", pemData, 0600); err != nil {
		return fmt.Errorf("failed to save test key: %w", err)
	}
	
	fmt.Println("Test keys created: test-key.pem")
	return nil
}
