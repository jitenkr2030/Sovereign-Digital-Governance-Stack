package domain

import (
	"encoding/base64"
	"time"
)

// KeyStatus represents the status of a key
type KeyStatus string

const (
	KeyStatusActive   KeyStatus = "ACTIVE"
	KeyStatusRotated  KeyStatus = "ROTATED"
	KeyStatusRevoked  KeyStatus = "REVOKED"
	KeyStatusArchived KeyStatus = "ARCHIVED"
)

// KeyAlgorithm represents the cryptographic algorithm
type KeyAlgorithm string

const (
	AlgorithmAES256GCM  KeyAlgorithm = "AES-256-GCM"
	AlgorithmRSA4096    KeyAlgorithm = "RSA-4096"
	AlgorithmECDSAP384  KeyAlgorithm = "ECDSA-P384"
	AlgorithmEd25519    KeyAlgorithm = "Ed25519"
)

// KeyUsage represents how a key can be used
type KeyUsage string

const (
	KeyUsageEncrypt    KeyUsage = "encrypt"
	KeyUsageDecrypt    KeyUsage = "decrypt"
	KeyUsageSign       KeyUsage = "sign"
	KeyUsageVerify     KeyUsage = "verify"
	KeyUsageWrap       KeyUsage = "wrap"
	KeyUsageUnwrap     KeyUsage = "unwrap"
)

// Key represents a cryptographic key in the vault
type Key struct {
	ID            string      `json:"id"`
	Alias         string      `json:"alias"`
	Algorithm     KeyAlgorithm `json:"algorithm"`
	CurrentVersion int        `json:"current_version"`
	Status        KeyStatus   `json:"status"`
	Usage         []KeyUsage  `json:"usage"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	RotatedAt     *time.Time  `json:"rotated_at,omitempty"`
	ExpiresAt     *time.Time  `json:"expires_at,omitempty"`
	Description   string      `json:"description,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// KeyVersion represents a specific version of a key
type KeyVersion struct {
	ID              string    `json:"id"`
	KeyID           string    `json:"key_id"`
	Version         int       `json:"version"`
	EncryptedMaterial string  `json:"encrypted_material"` // Base64 encoded
	IV              string    `json:"iv"`                 // Base64 encoded
	CreatedAt       time.Time `json:"created_at"`
}

// KeyPolicy represents access policies for a key
type KeyPolicy struct {
	ID              string   `json:"id"`
	KeyID           string   `json:"key_id"`
	AllowedRoles    []string `json:"allowed_roles"`
	AllowedServices []string `json:"allowed_services"`
	MaxOperations   int      `json:"max_operations"`
	RateLimit       int      `json:"rate_limit"` // requests per minute
}

// KeyOperation represents a key operation event
type KeyOperation struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	Operation string    `json:"operation"` // CREATE, ROTATE, REVOKE, ENCRYPT, DECRYPT
	ActorID   string    `json:"actor_id"`
	ServiceID string    `json:"service_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CreateKeyRequest represents a request to create a new key
type CreateKeyRequest struct {
	Alias       string        `json:"alias" binding:"required"`
	Algorithm   KeyAlgorithm  `json:"algorithm" binding:"required"`
	Usage       []KeyUsage    `json:"usage" binding:"required"`
	Description string        `json:"description"`
	Metadata    map[string]string `json:"metadata"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty"`
}

// RotateKeyRequest represents a request to rotate a key
type RotateKeyRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RevokeKeyRequest represents a request to revoke a key
type RevokeKeyRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// GenerateDataKeyRequest represents a request to generate a data key
type GenerateDataKeyRequest struct {
	KeyAlias     string `json:"key_alias" binding:"required"`
	KeySize      int    `json:"key_size"` // in bits, e.g., 256 for AES-256
}

// DataKeyResponse represents the response from generating a data key
type DataKeyResponse struct {
	KeyID           string `json:"key_id"`
	PlaintextKey    string `json:"plaintext_key"` // Base64 encoded - for immediate use
	EncryptedKey    string `json:"encrypted_key"` // Base64 encoded - for storage
	Version         int    `json:"version"`
}

// EncryptRequest represents an encryption request
type EncryptRequest struct {
	KeyID   string `json:"key_id" binding:"required"`
	Plaintext string `json:"plaintext" binding:"required"` // Base64 encoded
}

// DecryptRequest represents a decryption request
type DecryptRequest struct {
	KeyID     string `json:"key_id" binding:"required"`
	Ciphertext string `json:"ciphertext" binding:"required"` // Base64 encoded
	IV        string `json:"iv" binding:"required"` // Base64 encoded
}

// EncryptDecryptResponse represents the result of encryption/decryption
type EncryptDecryptResponse struct {
	Output    string `json:"output"` // Base64 encoded
	Version   int    `json:"version"`
}

// KeyListResponse represents a paginated list of keys
type KeyListResponse struct {
	Keys      []*Key  `json:"keys"`
	Total     int64   `json:"total"`
	Page      int     `json:"page"`
	PageSize  int     `json:"page_size"`
}

// NewKeyListResponse creates a new paginated key list response
func NewKeyListResponse(keys []*Key, total int64, page, pageSize int) *KeyListResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &KeyListResponse{
		Keys:      keys,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}
}

// Validate validates the key algorithm
func (a KeyAlgorithm) Validate() bool {
	switch a {
	case AlgorithmAES256GCM, AlgorithmRSA4096, AlgorithmECDSAP384, AlgorithmEd25519:
		return true
	default:
		return false
	}
}

// GetKeySize returns the key size in bits for symmetric algorithms
func (a KeyAlgorithm) GetKeySize() int {
	switch a {
	case AlgorithmAES256GCM:
		return 256
	case AlgorithmRSA4096:
		return 4096
	case AlgorithmECDSAP384:
		return 384
	case AlgorithmEd25519:
		return 256
	default:
		return 256
	}
}

// DecryptMaterial decrypts the base64-encoded key material
func (v *KeyVersion) DecryptMaterial() ([]byte, error) {
	material, err := base64.StdEncoding.DecodeString(v.EncryptedMaterial)
	if err != nil {
		return nil, err
	}
	return material, nil
}

// EncryptMaterial encrypts key material and stores it as base64
func (v *KeyVersion) EncryptMaterial(material []byte) {
	v.EncryptedMaterial = base64.StdEncoding.EncodeToString(material)
}

// SetIV sets the initialization vector as base64
func (v *KeyVersion) SetIV(iv []byte) {
	v.IV = base64.StdEncoding.EncodeToString(iv)
}

// GetIV returns the initialization vector bytes
func (v *KeyVersion) GetIV() ([]byte, error) {
	return base64.StdEncoding.DecodeString(v.IV)
}
