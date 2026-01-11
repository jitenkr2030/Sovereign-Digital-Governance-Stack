package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/csic/wallet-governance/internal/config"
	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/google/uuid"
)

// HSMService handles HSM (Hardware Security Module) operations
type HSMService struct {
	config config.HSMConfig
	key    *ecdsa.PrivateKey
}

// NewHSMService creates a new HSM service
func NewHSMService(cfg config.HSMConfig) (*HSMService, error) {
	svc := &HSMService{
		config: cfg,
	}

	if !cfg.Enabled {
		// Generate soft key for development
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate soft key: %w", err)
		}
		svc.key = key
	}

	return svc, nil
}

// Sign signs a message hash using the HSM
func (s *HSMService) Sign(ctx context.Context, messageHash string) (*models.SignatureResult, error) {
	hash, err := hex.DecodeString(messageHash)
	if err != nil {
		return nil, fmt.Errorf("invalid message hash: %w", err)
	}

	if s.key == nil {
		return nil, fmt.Errorf("HSM not initialized")
	}

	r, s2, err := ecdsa.Sign(rand.Reader, s.key, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature as R|S format
	signature := append(r.Bytes(), s2.Bytes()...)

	return &models.SignatureResult{
		Signature:    hex.EncodeToString(signature),
		PublicKey:    s.GetPublicKey(),
		Algorithm:    "ECDSA",
		Curve:        "P-256",
		Timestamp:    time.Now(),
	}, nil
}

// Verify verifies a signature
func (s *HSMService) Verify(ctx context.Context, messageHash, signature, publicKey string) (bool, error) {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	pubBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, fmt.Errorf("invalid public key format: %w", err)
	}

	hash, err := hex.DecodeString(messageHash)
	if err != nil {
		return false, fmt.Errorf("invalid message hash: %w", err)
	}

	// Extract R and S values
	if len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature length")
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s2 := new(big.Int).SetBytes(sigBytes[32:])

	// Reconstruct public key
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	if x == nil {
		return false, fmt.Errorf("invalid public key")
	}

	publicKeyECDSA := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return ecdsa.Verify(publicKeyECDSA, hash, r, s2), nil
}

// GenerateKey generates a new key pair
func (s *HSMService) GenerateKey(ctx context.Context, label string) (*models.KeyPair, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Encode private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	publicKeyBytes := elliptic.Marshal(elliptic.P256(), key.X, key.Y)

	return &models.KeyPair{
		ID:           uuid.New().String(),
		Label:        label,
		PrivateKey:   hex.EncodeToString(privateKeyPEM),
		PublicKey:    hex.EncodeToString(publicKeyBytes),
		PublicKeyHash: s.HashPublicKey(publicKeyBytes),
		Algorithm:    "ECDSA",
		Curve:        "P-256",
		CreatedAt:    time.Now(),
	}, nil
}

// HashPublicKey hashes a public key
func (s *HSMService) HashPublicKey(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}

// GetPublicKey returns the HSM public key
func (s *HSMService) GetPublicKey() string {
	if s.key == nil {
		return ""
	}

	publicKeyBytes := elliptic.Marshal(elliptic.P256(), s.key.X, s.key.Y)
	return hex.EncodeToString(publicKeyBytes)
}

// GetPublicKeyHash returns the hash of the HSM public key
func (s *HSMService) GetPublicKeyHash() string {
	return s.HashPublicKey(elliptic.Marshal(elliptic.P256(), s.key.X, s.key.Y))
}

// KeyPair represents a generated key pair
type KeyPair struct {
	ID           string
	Label        string
	PrivateKey   string
	PublicKey    string
	PublicKeyHash string
	Algorithm    string
	Curve        string
	CreatedAt    time.Time
}

// SignatureResult represents the result of a signature operation
type SignatureResult struct {
	Signature    string
	PublicKey    string
	Algorithm    string
	Curve        string
	Timestamp    time.Time
}
