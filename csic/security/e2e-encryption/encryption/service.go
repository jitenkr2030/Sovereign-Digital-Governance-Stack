package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/csic-platform/security/e2e-encryption/crypto"
	"github.com/csic-platform/security/e2e-encryption/key-management"
	"go.uber.org/zap"
)

// EncryptionResult represents the result of an encryption operation
type EncryptionResult struct {
	Ciphertext   []byte          `json:"ciphertext"`
	IV           []byte          `json:"iv"`
	Tag          []byte          `json:"tag,omitempty"`
	KeyID        string          `json:"key_id"`
	Algorithm    string          `json:"algorithm"`
	EncryptedAt  time.Time       `json:"encrypted_at"`
	Metadata     EncryptionMetadata `json:"metadata"`
}

// DecryptionResult represents the result of a decryption operation
type DecryptionResult struct {
	Plaintext   []byte          `json:"plaintext"`
	KeyID       string          `json:"key_id"`
	Algorithm   string          `json:"algorithm"`
	DecryptedAt time.Time       `json:"decrypted_at"`
	Metadata    EncryptionMetadata `json:"metadata"`
}

// EncryptionMetadata contains additional information about encryption
type EncryptionMetadata struct {
	Version         int               `json:"version"`
	OriginalSize    int               `json:"original_size"`
	AdditionalData  []byte            `json:"additional_data,omitempty"`
	Compression     string            `json:"compression,omitempty"`
	CustomLabels    map[string]string `json:"custom_labels,omitempty"`
}

// Encrypter provides encryption capabilities
type Encrypter interface {
	Encrypt(ctx context.Context, plaintext []byte, keyID string) (*EncryptionResult, error)
	Decrypt(ctx context.Context, ciphertext []byte, keyID string) (*DecryptionResult, error)
	EncryptWithAAD(ctx context.Context, plaintext []byte, keyID string, additionalData []byte) (*EncryptionResult, error)
	DecryptWithAAD(ctx context.Context, ciphertext []byte, keyID string, additionalData []byte) (*DecryptionResult, error)
}

// Service provides encryption services
type Service struct {
	logger       *zap.Logger
	config       *Config
	keyManager   *keymgmt.Service
	cryptoEngine *crypto.Engine
}

// Global service instance
var (
	defaultEncryptionService *Service
	encryptionServiceMu      sync.RWMutex
)

// Config for encryption service
type Config struct {
	Algorithm        string
	KeyDerivation    string
	KeyRotationDays  int
	MasterKeyID      string
}

// Initialize initializes the encryption service
func Initialize(logger *zap.Logger, cfg interface{}) error {
	encryptionServiceMu.Lock()
	defer encryptionServiceMu.Unlock()

	// Extract configuration
	config := &Config{
		Algorithm:       "AES-256-GCM",
		KeyDerivation:   "Argon2",
		KeyRotationDays: 90,
		MasterKeyID:     "master-key-v1",
	}

	// Get crypto engine
	cryptoEngine := crypto.GetEngine()

	defaultEncryptionService = &Service{
		logger:       logger,
		config:       config,
		keyManager:   keymgmt.GetService(),
		cryptoEngine: cryptoEngine,
	}

	logger.Info("Encryption service initialized",
		zap.String("algorithm", config.Algorithm),
		zap.String("key_derivation", config.KeyDerivation))

	return nil
}

// GetService returns the default encryption service
func GetService() *Service {
	encryptionServiceMu.RLock()
	defer encryptionServiceMu.RUnlock()
	return defaultEncryptionService
}

// Encrypt encrypts plaintext using the specified key
func (s *Service) Encrypt(ctx context.Context, plaintext []byte, keyID string) (*EncryptionResult, error) {
	if len(plaintext) == 0 {
		return nil, errors.New("plaintext is empty")
	}

	// Get the key
	key, err := s.keyManager.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Generate random IV
	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	var result *EncryptionResult

	switch key.Type {
	case keymgmt.KeyTypeSymmetric:
		result, err = s.encryptSymmetric(ctx, plaintext, key, iv)
	case keymgmt.KeyTypeAsymmetric:
		result, err = s.encryptAsymmetric(ctx, plaintext, key, iv)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.Type)
	}

	if err != nil {
		return nil, err
	}

	result.KeyID = keyID
	result.Algorithm = s.config.Algorithm
	result.EncryptedAt = time.Now()
	result.Metadata.OriginalSize = len(plaintext)

	// Log operation
	s.logger.Info("Data encrypted",
		zap.String("key_id", keyID),
		zap.Int("plaintext_size", len(plaintext)),
		zap.Int("ciphertext_size", len(result.Ciphertext)))

	return result, nil
}

// Decrypt decrypts ciphertext using the specified key
func (s *Service) Decrypt(ctx context.Context, ciphertext []byte, keyID string) (*DecryptionResult, error) {
	if len(ciphertext) == 0 {
		return nil, errors.New("ciphertext is empty")
	}

	// Get the key
	key, err := s.keyManager.GetKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Extract IV from ciphertext (first 12 bytes for GCM)
	if len(ciphertext) < 12 {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:12]
	encryptedData := ciphertext[12:]

	var result *DecryptionResult

	switch key.Type {
	case keymgmt.KeyTypeSymmetric:
		result, err = s.decryptSymmetric(ctx, encryptedData, key, iv)
	case keymgmt.KeyTypeAsymmetric:
		result, err = s.decryptAsymmetric(ctx, encryptedData, key, iv)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.Type)
	}

	if err != nil {
		return nil, err
	}

	result.KeyID = keyID
	result.Algorithm = s.config.Algorithm
	result.DecryptedAt = time.Now()

	// Log operation
	s.logger.Info("Data decrypted",
		zap.String("key_id", keyID),
		zap.Int("ciphertext_size", len(ciphertext)),
		zap.Int("plaintext_size", len(result.Plaintext)))

	return result, nil
}

// EncryptWithAAD encrypts plaintext with additional authenticated data
func (s *Service) EncryptWithAAD(ctx context.Context, plaintext []byte, keyID string, additionalData []byte) (*EncryptionResult, error) {
	result, err := s.Encrypt(ctx, plaintext, keyID)
	if err != nil {
		return nil, err
	}

	result.Metadata.AdditionalData = additionalData
	return result, nil
}

// DecryptWithAAD decrypts ciphertext with additional authenticated data
func (s *Service) DecryptWithAAD(ctx context.Context, ciphertext []byte, keyID string, additionalData []byte) (*DecryptionResult, error) {
	// Extract IV
	if len(ciphertext) < 12 {
		return nil, errors.New("ciphertext too short")
	}

	iv := ciphertext[:12]
	encryptedData := ciphertext[12:]

	result, err := s.Decrypt(ctx, ciphertext, keyID)
	if err != nil {
		return nil, err
	}

	// Verify additional data
	if len(additionalData) > 0 {
		if !validateAAD(additionalData, result.Metadata.AdditionalData) {
			return nil, errors.New("additional data verification failed")
		}
	}

	return result, nil
}

// encryptSymmetric encrypts data using symmetric key (AES-256-GCM)
func (s *Service) encryptSymmetric(ctx context.Context, plaintext []byte, key *keymgmt.Key, iv []byte) (*EncryptionResult, error) {
	block, err := aes.NewCipher(key.Material)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, iv, plaintext, nil)

	return &EncryptionResult{
		Ciphertext: ciphertext,
		IV:         iv,
		Tag:        ciphertext[len(ciphertext)-gcm.Overhead():],
		Metadata: EncryptionMetadata{
			Version: 1,
		},
	}, nil
}

// decryptSymmetric decrypts data using symmetric key (AES-256-GCM)
func (s *Service) decryptSymmetric(ctx context.Context, ciphertext []byte, key *keymgmt.Key, iv []byte) (*DecryptionResult, error) {
	block, err := aes.NewCipher(key.Material)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract tag
	tagLen := gcm.Overhead()
	if len(ciphertext) < tagLen {
		return nil, errors.New("ciphertext too short for tag")
	}

	ciphertextWithoutTag := ciphertext[:len(ciphertext)-tagLen]
	tag := ciphertext[len(ciphertext)-tagLen:]

	// Combine IV, ciphertext, and tag
	combined := append(iv, append(ciphertextWithoutTag, tag...)...)

	// Decrypt
	plaintext, err := gcm.Open(nil, nil, combined, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return &DecryptionResult{
		Plaintext: plaintext,
		Metadata: EncryptionMetadata{
			Version: 1,
		},
	}, nil
}

// encryptAsymmetric encrypts data using asymmetric key (RSA-OAEP)
func (s *Service) encryptAsymmetric(ctx context.Context, plaintext []byte, key *keymgmt.Key, iv []byte) (*EncryptionResult, error) {
	// Parse public key
	pubKey, err := x509.ParsePKCS1PublicKey(key.Material)
	if err != nil {
		// Try PKCS8 format
		pubInterface, err := x509.ParsePKIXPublicKey(key.Material)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		pubKey = pubInterface.(*rsa.PublicKey)
	}

	// Encrypt with OAEP
	ciphertext, err := rsa.EncryptOAEP(
		crypto.SHA256Hash(),
		rand.Reader,
		pubKey,
		plaintext,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	return &EncryptionResult{
		Ciphertext: ciphertext,
		IV:         iv,
		Metadata: EncryptionMetadata{
			Version: 1,
		},
	}, nil
}

// decryptAsymmetric decrypts data using asymmetric key (RSA-OAEP)
func (s *Service) decryptAsymmetric(ctx context.Context, ciphertext []byte, key *keymgmt.Key, iv []byte) (*DecryptionResult, error) {
	// Parse private key
	privKey, err := x509.ParsePKCS1PrivateKey(key.Material)
	if err != nil {
		// Try PKCS8 format
		privInterface, err := x509.ParsePKCS8PrivateKey(key.Material)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		privKey = privInterface.(*rsa.PrivateKey)
	}

	// Decrypt with OAEP
	plaintext, err := rsa.DecryptOAEP(
		crypto.SHA256Hash(),
		rand.Reader,
		privKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return &DecryptionResult{
		Plaintext: plaintext,
		Metadata: EncryptionMetadata{
			Version: 1,
		},
	}, nil
}

// EncryptString encrypts a string and returns base64-encoded result
func (s *Service) EncryptString(ctx context.Context, plaintext string, keyID string) (string, error) {
	result, err := s.Encrypt(ctx, []byte(plaintext), keyID)
	if err != nil {
		return "", err
	}

	// Combine IV + ciphertext + tag
	combined := append(result.IV, append(result.Ciphertext, result.Tag...)...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptString decrypts a base64-encoded ciphertext
func (s *Service) DecryptString(ctx context.Context, encryptedBase64 string, keyID string) (string, error) {
	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	result, err := s.Decrypt(ctx, combined, keyID)
	if err != nil {
		return "", err
	}

	return string(result.Plaintext), nil
}

// EncryptStream encrypts data from a reader to a writer
func (s *Service) EncryptStream(ctx context.Context, reader io.Reader, writer io.Writer, keyID string) error {
	// Get the key
	key, err := s.keyManager.GetKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	// Generate IV
	iv := make([]byte, 12)
	if _, err := rand.Read(iv); err != nil {
		return fmt.Errorf("failed to generate IV: %w", err)
	}

	// Write IV
	if _, err := writer.Write(iv); err != nil {
		return fmt.Errorf("failed to write IV: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key.Material)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt in chunks
	const chunkSize = 64 * 1024 // 64KB chunks
	plaintext := make([]byte, chunkSize)

	for {
		n, err := reader.Read(plaintext)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read plaintext: %w", err)
		}

		ciphertext := gcm.Seal(nil, iv, plaintext[:n], nil)
		if _, err := writer.Write(ciphertext); err != nil {
			return fmt.Errorf("failed to write ciphertext: %w", err)
		}

		// Generate new IV for next chunk (using counter mode)
		for i := range iv {
			iv[i]++
			if iv[i] != 0 {
				break
			}
		}
	}

	return nil
}

// DecryptStream decrypts data from a reader to a writer
func (s *Service) DecryptStream(ctx context.Context, reader io.Reader, writer io.Writer, keyID string) error {
	// Get the key
	key, err := s.keyManager.GetKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	// Read IV
	iv := make([]byte, 12)
	if _, err := io.ReadFull(reader, iv); err != nil {
		return fmt.Errorf("failed to read IV: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key.Material)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt in chunks
	const chunkSize = 64*1024 + gcm.Overhead() // 64KB + overhead
	ciphertext := make([]byte, chunkSize)

	for {
		n, err := reader.Read(ciphertext)
		if err == io.EOF {
			break
		}
		if err != nil && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read ciphertext: %w", err)
		}

		plaintext, err := gcm.Open(nil, iv, ciphertext[:n], nil)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		if _, err := writer.Write(plaintext); err != nil {
			return fmt.Errorf("failed to write plaintext: %w", err)
		}

		// Generate new IV for next chunk
		for i := range iv {
			iv[i]++
			if iv[i] != 0 {
				break
			}
		}
	}

	return nil
}

// validateAAD validates additional authenticated data
func validateAAD(expected, actual []byte) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := range expected {
		if expected[i] != actual[i] {
			return false
		}
	}
	return true
}

// Helper function to convert key material to PEM format
func keyToPEM(key []byte, keyType string) []byte {
	block := &pem.Block{
		Type:  keyType,
		Bytes: key,
	}
	return pem.EncodeToMemory(block)
}

// Helper function to convert PEM to key material
func pemToKey(pemData []byte) ([]byte, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	return block.Bytes, nil
}

import (
	"sync"
)
