package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// HSMService provides Hardware Security Module operations
type HSMService interface {
	// Initialize initializes the HSM connection
	Initialize(ctx context.Context) error

	// Sign signs data using the HSM
	Sign(ctx context.Context, data []byte) ([]byte, error)

	// Verify verifies a signature using the HSM
	Verify(ctx context.Context, data, signature []byte) (bool, error)

	// GenerateKey generates a new key in the HSM
	GenerateKey(ctx context.Context, keyID string) error

	// ImportKey imports an existing key into the HSM
	ImportKey(ctx context.Context, keyID string, keyData []byte) error

	// ExportKey exports a public key from the HSM
	ExportKey(ctx context.Context, keyID string) ([]byte, error)

	// DestroyKey destroys a key in the HSM
	DestroyKey(ctx context.Context, keyID string) error

	// GetKeyInfo returns information about a key
	GetKeyInfo(ctx context.Context, keyID string) (*KeyInfo, error)

	// IsConnected checks if HSM is connected
	IsConnected() bool
}

// KeyInfo represents information about an HSM key
type KeyInfo struct {
	KeyID       string    `json:"key_id"`
	KeyType     string    `json:"key_type"`
	Algorithm   string    `json:"algorithm"`
	KeySize     int       `json:"key_size"`
	CreatedAt   string    `json:"created_at"`
	Usage       string    `json:"usage"`
	Exportable  bool      `json:"exportable"`
}

// hsmService implements HSMService using PKCS#11
type hsmService struct {
	module    string
	pin       string
	keyLabel  string
	connected bool
}

// NewHSMService creates a new HSM service
func NewHSMService(module, pin, keyLabel string) HSMService {
	return &hsmService{
		module:   module,
		pin:      pin,
		keyLabel: keyLabel,
	}
}

// Initialize initializes the HSM connection
func (h *hsmService) Initialize(ctx context.Context) error {
	// In a real implementation, this would load the PKCS#11 library
	// and initialize the HSM connection
	// For this implementation, we simulate HSM behavior

	h.connected = true
	return nil
}

// Sign signs data using the HSM
func (h *hsmService) Sign(ctx context.Context, data []byte) ([]byte, error) {
	if !h.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// In a real implementation, this would use PKCS#11 C_Sign
	// For simulation, use software signing with HSM key metadata
	hashed := sha256.Sum256(data)

	signature, err := rsa.SignPKCS1v15(rand.Reader, h.getSimulatedPrivateKey(), crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return signature, nil
}

// Verify verifies a signature using the HSM
func (h *hsmService) Verify(ctx context.Context, data, signature []byte) (bool, error) {
	if !h.connected {
		return false, fmt.Errorf("HSM not connected")
	}

	hashed := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(&h.getSimulatedPrivateKey().PublicKey, crypto.SHA256, hashed[:], signature)
	return err == nil, nil
}

// GenerateKey generates a new key in the HSM
func (h *hsmService) GenerateKey(ctx context.Context, keyID string) error {
	if !h.connected {
		return fmt.Errorf("HSM not connected")
	}

	// In a real implementation, this would use PKCS#11 C_GenerateKeyPair
	// For simulation, log the key generation
	return nil
}

// ImportKey imports an existing key into the HSM
func (h *hsmService) ImportKey(ctx context.Context, keyID string, keyData []byte) error {
	if !h.connected {
		return fmt.Errorf("HSM not connected")
	}

	// In a real implementation, this would use PKCS#11 C_UnwrapKey
	return nil
}

// ExportKey exports a public key from the HSM
func (h *hsmService) ExportKey(ctx context.Context, keyID string) ([]byte, error) {
	if !h.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// In a real implementation, this would use PKCS#11 C_ExportKey
	// For simulation, export the simulated public key
	pubKey := h.getSimulatedPrivateKey().PublicKey

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: marshalPKCS1PublicKey(&pubKey),
	})

	return pemData, nil
}

// DestroyKey destroys a key in the HSM
func (h *hsmService) DestroyKey(ctx context.Context, keyID string) error {
	if !h.connected {
		return fmt.Errorf("HSM not connected")
	}

	// In a real implementation, this would use PKCS#11 C_DestroyObject
	return nil
}

// GetKeyInfo returns information about a key
func (h *hsmService) GetKeyInfo(ctx context.Context, keyID string) (*KeyInfo, error) {
	return &KeyInfo{
		KeyID:      keyID,
		KeyType:    "RSA",
		Algorithm:  "RSA-SHA256",
		KeySize:    2048,
		CreatedAt:  "2024-01-01T00:00:00Z",
		Usage:      "digital_signature",
		Exportable: false,
	}, nil
}

// IsConnected checks if HSM is connected
func (h *hsmService) IsConnected() bool {
	return h.connected
}

// getSimulatedPrivateKey returns a simulated private key for testing
func (h *hsmService) getSimulatedPrivateKey() *rsa.PrivateKey {
	// Generate a 2048-bit RSA key for simulation
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate private key: %v", err))
	}
	return privateKey
}

// FileHSMService provides file-based HSM simulation for testing
type FileHSMService struct {
	keyPath    string
	privateKey *rsa.PrivateKey
}

// NewFileHSMService creates a new file-based HSM service
func NewFileHSMService(keyPath string) (*FileHSMService, error) {
	// Try to load existing key
	fhs := &FileHSMService{keyPath: keyPath}

	data, err := ioutil.ReadFile(filepath.Join(keyPath, "private.pem"))
	if err == nil {
		// Load existing key
		block, _ := pem.Decode(data)
		fhs.privateKey, err = x509PKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	} else {
		// Generate new key
		fhs.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key: %w", err)
		}

		// Save key
		os.MkdirAll(keyPath, 0700)
		pemData := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509EncodePKCS1PrivateKey(fhs.privateKey),
		})
		if err := ioutil.WriteFile(filepath.Join(keyPath, "private.pem"), pemData, 0600); err != nil {
			return nil, fmt.Errorf("failed to save key: %w", err)
		}
	}

	return fhs, nil
}

// Sign signs data
func (f *FileHSMService) Sign(ctx context.Context, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, f.privateKey, crypto.SHA256, hashed[:])
}

// Verify verifies a signature
func (f *FileHSMService) Verify(ctx context.Context, data, signature []byte) (bool, error) {
	hashed := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(&f.privateKey.PublicKey, crypto.SHA256, hashed[:], signature)
	return err == nil, nil
}

// GetPublicKey returns the public key
func (f *FileHSMService) GetPublicKey() []byte {
	pubKey := f.privateKey.PublicKey
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: marshalPKCS1PublicKey(&pubKey),
	})
}

// Helper functions for key encoding (simplified)

func x509EncodePKCS1PrivateKey(key *rsa.PrivateKey) []byte {
	// Simplified - in production use x509.MarshalPKCS1PrivateKey
	return key.N.Bytes()
}

func x509PKCS1PrivateKey(data []byte) (*rsa.PrivateKey, error) {
	// Simplified - in production use x509.ParsePKCS1PrivateKey
	return &rsa.PrivateKey{}, nil
}

func marshalPKCS1PublicKey(key *rsa.PublicKey) []byte {
	// Simplified - in production use x509.MarshalPKCS1PublicKey
	return key.N.Bytes()
}

// KeyRotationService handles automatic key rotation
type KeyRotationService struct {
	hsm        HSMService
	keyID      string
	interval   int // days
	lastRotated time.Time
}

// NewKeyRotationService creates a new key rotation service
func NewKeyRotationService(hsm HSMService, keyID string, intervalDays int) *KeyRotationService {
	return &KeyRotationService{
		hsm:      hsm,
		keyID:    keyID,
		interval: intervalDays,
	}
}

// ShouldRotate checks if keys should be rotated
func (k *KeyRotationService) ShouldRotate() bool {
	return int(k.lastRotated.Sub(time.Now()).Hours())/24 > k.interval
}

// Rotate rotates the key
func (k *KeyRotationService) Rotate(ctx context.Context) error {
	// Generate new key
	if err := k.hsm.GenerateKey(ctx, k.keyID); err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	k.lastRotated = time.Now()
	return nil
}

// AuditLogService provides audit logging for HSM operations
type AuditLogService struct {
	logger func(event string, data map[string]interface{})
}

// NewAuditLogService creates a new audit log service
func NewAuditLogService(logger func(event string, data map[string]interface{})) *AuditLogService {
	return &AuditLogService{logger: logger}
}

// Log logs an HSM event
func (a *AuditLogService) Log(event string, data map[string]interface{}) {
	if a.logger != nil {
		a.logger(event, data)
	}
}

// HSMEvents defines HSM audit event types
const (
	HSMEventInitialize     = "HSM_INITIALIZE"
	HSMEventSign           = "HSM_SIGN"
	HSMEventVerify         = "HSM_VERIFY"
	HSMEventGenerateKey    = "HSM_GENERATE_KEY"
	HSMEventImportKey      = "HSM_IMPORT_KEY"
	HSMEventExportKey      = "HSM_EXPORT_KEY"
	HSMEventDestroyKey     = "HSM_DESTROY_KEY"
	HSMEventKeyRotation    = "HSM_KEY_ROTATION"
	HSMEventError          = "HSM_ERROR"
)

// SignWithAudit signs data and logs the operation
func (a *AuditLogService) SignWithAudit(ctx context.Context, hsm HSMService, keyID string, data []byte) ([]byte, error) {
	signature, err := hsm.Sign(ctx, data)
	a.Log(HSMEventSign, map[string]interface{}{
		"key_id":   keyID,
		"data_len": len(data),
		"sig_len":  len(signature),
		"error":    err,
	})
	return signature, err
}

// VerifyWithAudit verifies a signature and logs the operation
func (a *AuditLogService) VerifyWithAudit(ctx context.Context, hsm HSMService, keyID string, data, signature []byte) (bool, error) {
	valid, err := hsm.Verify(ctx, data, signature)
	a.Log(HSMEventVerify, map[string]interface{}{
		"key_id":   keyID,
		"valid":    valid,
		"error":    err,
	})
	return valid, err
}
