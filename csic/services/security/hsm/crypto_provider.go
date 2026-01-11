package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// CryptoProvider defines the interface for cryptographic operations
type CryptoProvider interface {
	// Key Management
	GenerateKey(keyType KeyType, keySize int) (*KeyInfo, error)
	ImportKey(keyType KeyType, keyData []byte, password string) (*KeyInfo, error)
	ExportPublicKey(keyID string) ([]byte, error)
	DeleteKey(keyID string) error
	
	// Signing Operations
	Sign(keyID string, data []byte) ([]byte, error)
	Verify(keyID string, data, signature []byte) bool
	
	// Encryption Operations
	Encrypt(keyID string, data []byte) ([]byte, error)
	Decrypt(keyID string, encryptedData []byte) ([]byte, error)
	
	// Session Management
	Close() error
}

// KeyType defines supported key types
type KeyType string

const (
	KeyTypeRSA     KeyType = "RSA"
	KeyTypeECDSA   KeyType = "ECDSA"
	KeyTypeAES     KeyType = "AES"
	KeyTypeHMAC    KeyType = "HMAC"
)

// KeyInfo contains key metadata
type KeyInfo struct {
	ID        string    `json:"id"`
	Type      KeyType   `json:"type"`
	Algorithm string    `json:"algorithm"`
	Size      int       `json:"size"`
	CreatedAt string    `json:"created_at"`
	PublicKey []byte    `json:"public_key,omitempty"`
}

// KeyGenerationParams contains parameters for key generation
type KeyGenerationParams struct {
	KeyType     KeyType
	KeySize     int
	Extractable bool
	Label       string
}

// SignatureResult contains signature and metadata
type SignatureResult struct {
	Signature   []byte
	Algorithm   crypto.SignerOpts
	KeyID       string
	Timestamp   string
}

// EncryptionResult contains encrypted data and metadata
type EncryptionResult struct {
	Ciphertext  []byte
	IV          []byte
	Algorithm   string
	KeyID       string
}

// SoftHSMProvider implements CryptoProvider using software-based cryptography
type SoftHSMProvider struct {
	keys map[string]*softKey
}

// NewSoftHSMProvider creates a new software-based HSM provider
func NewSoftHSMProvider() *SoftHSMProvider {
	return &SoftHSMProvider{
		keys: make(map[string]*softKey),
	}
}

// softKey represents a software-managed key
type softKey struct {
	ID        string
	Type      KeyType
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey
	CreatedAt  string
}

// GenerateKey generates a new key pair
func (p *SoftHSMProvider) GenerateKey(keyType KeyType, keySize int) (*KeyInfo, error) {
	keyID := fmt.Sprintf("key-%d", len(p.keys)+1)
	
	var privKey crypto.PrivateKey
	var pubKey crypto.PublicKey
	
	switch keyType {
	case KeyTypeRSA:
		rsaKey, err := rsa.GenerateKey(crypto.rand.Reader, keySize)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		privKey = rsaKey
		pubKey = &rsaKey.PublicKey
		
	case KeyTypeECDSA:
		// For ECDSA, we use P-256 by default
		// In production, this would use a proper curve
		return nil, fmt.Errorf("ECDSA not implemented in soft provider")
		
	case KeyTypeAES:
		// AES is symmetric, no public/private key pair
		key := make([]byte, keySize/8)
		crypto.rand.Read(key)
		privKey = key
		pubKey = nil
		
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
	
	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(pubKey)
	
	p.keys[keyID] = &softKey{
		ID:        keyID,
		Type:      keyType,
		PrivateKey: privKey,
		PublicKey:  pubKey,
		CreatedAt:  "now",
	}
	
	return &KeyInfo{
		ID:        keyID,
		Type:      keyType,
		Algorithm: string(keyType),
		Size:      keySize,
		CreatedAt: "now",
		PublicKey: pubKeyBytes,
	}, nil
}

// ImportKey imports an existing key
func (p *SoftHSMProvider) ImportKey(keyType KeyType, keyData []byte, password string) (*KeyInfo, error) {
	keyID := fmt.Sprintf("imported-%d", len(p.keys)+1)
	
	var privKey crypto.PrivateKey
	var err error
	
	if password != "" {
		// Decrypt key data
		keyData, err = decryptData(keyData, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key: %w", err)
		}
	}
	
	// Parse key data
	privKey, err = x509.ParsePKCS8PrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key data: %w", err)
	}
	
	p.keys[keyID] = &softKey{
		ID:        keyID,
		Type:      keyType,
		PrivateKey: privKey,
		PublicKey:  nil,
		CreatedAt:  "now",
	}
	
	return &KeyInfo{
		ID:        keyID,
		Type:      keyType,
		Algorithm: string(keyType),
		CreatedAt: "now",
	}, nil
}

// ExportPublicKey exports the public key
func (p *SoftHSMProvider) ExportPublicKey(keyID string) ([]byte, error) {
	key, exists := p.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	
	if key.PublicKey == nil {
		return nil, fmt.Errorf("no public key for key type: %s", key.Type)
	}
	
	return x509.MarshalPKIXPublicKey(key.PublicKey)
}

// DeleteKey removes a key
func (p *SoftHSMProvider) DeleteKey(keyID string) error {
	if _, exists := p.keys[keyID]; !exists {
		return fmt.Errorf("key not found: %s", keyID)
	}
	delete(p.keys, keyID)
	return nil
}

// Sign signs data using the specified key
func (p *SoftHSMProvider) Sign(keyID string, data []byte) ([]byte, error) {
	key, exists := p.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	
	signer, ok := key.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("key does not support signing: %s", keyID)
	}
	
	return signer.Sign(crypto.rand.Reader, data, crypto.SHA256)
}

// Verify verifies a signature
func (p *SoftHSMProvider) Verify(keyID string, data, signature []byte) bool {
	key, exists := p.keys[keyID]
	if !exists {
		return false
	}
	
	pubKey, ok := key.PublicKey.(*rsa.PublicKey)
	if !ok {
		return false
	}
	
	err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, data, signature)
	return err == nil
}

// Encrypt encrypts data
func (p *SoftHSMProvider) Encrypt(keyID string, data []byte) ([]byte, error) {
	// Simplified encryption for demonstration
	// In production, this would use proper AES-GCM
	key, exists := p.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %keyID")
	}
	
	aesKey, ok := key.PrivateKey.([]byte)
	if !ok {
		return nil, fmt.Errorf("key does not support encryption: %s", keyID)
	}
	
	// Return encrypted data (simplified)
	encrypted := make([]byte, len(data))
	copy(encrypted, data)
	return encrypted, nil
}

// Decrypt decrypts data
func (p *SoftHSMProvider) Decrypt(keyID string, encryptedData []byte) ([]byte, error) {
	key, exists := p.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	
	aesKey, ok := key.PrivateKey.([]byte)
	if !ok {
		return nil, fmt.Errorf("key does not support decryption: %s", keyID)
	}
	
	// Return decrypted data (simplified)
	decrypted := make([]byte, len(encryptedData))
	copy(decrypted, encryptedData)
	return decrypted, nil
}

// Close cleans up the provider
func (p *SoftHSMProvider) Close() error {
	p.keys = make(map[string]*softKey)
	return nil
}

// Helper function for decrypting key data
func decryptData(data []byte, password string) ([]byte, error) {
	// Simplified - in production this would use proper decryption
	return data, nil
}

// PEMEncode encodes a private key to PEM format
func PEMEncode(key crypto.PrivateKey) ([]byte, error) {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}
	
	return pem.EncodeToMemory(block), nil
}
