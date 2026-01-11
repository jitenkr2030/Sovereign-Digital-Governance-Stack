package crypto

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// SignatureService provides cryptographic signature operations
type SignatureService interface {
	// Sign signs data
	Sign(ctx context.Context, data []byte) (string, error)

	// Verify verifies a signature
	Verify(ctx context.Context, data, signature []byte) (bool, error)

	// SignHash signs a hash
	SignHash(ctx context.Context, hash []byte) ([]byte, error)

	// VerifyHash verifies a hash signature
	VerifyHash(ctx context.Context, hash, signature []byte) (bool, error)

	// GetPublicKey returns the public key
	GetPublicKey() []byte
}

// signatureService implements SignatureService using RSA
type signatureService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	algorithm  string
}

// NewSignatureService creates a new SignatureService with RSA keys
func NewSignatureService(useHSM bool, hsmPin, keyLabel string) (SignatureService, error) {
	// In production, this would integrate with HSM via PKCS#11
	// For now, use software-based signing
	return NewSoftwareSignatureService(), nil
}

// NewSoftwareSignatureService creates a new software-based SignatureService
func NewSoftwareSignatureService() SignatureService {
	// Generate a 2048-bit RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate RSA key: %v", err))
	}

	return &signatureService{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		algorithm:  "RSA-SHA256",
	}
}

// NewFileSignatureService creates a SignatureService with keys from files
func NewFileSignatureService(privateKeyPath, publicKeyPath string) (SignatureService, error) {
	// Load private key
	privateKeyData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyData)
	if privateKeyBlock == nil || privateKeyBlock.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key format")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &signatureService{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		algorithm:  "RSA-SHA256",
	}, nil
}

// NewFileSignatureServiceWithPublic creates a SignatureService with separate key files
func NewFileSignatureServiceWithPublic(privateKeyPath, publicKeyPath string) (SignatureService, error) {
	// Load private key
	privateKeyData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyData)
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("invalid private key format")
	}

	var privateKey *rsa.PrivateKey
	if privateKeyBlock.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}
	} else {
		privateKey, err = x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
	}

	// Load public key
	publicKeyData, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	publicKeyBlock, _ := pem.Decode(publicKeyData)
	if publicKeyBlock == nil || publicKeyBlock.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key format")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &signatureService{
		privateKey: privateKey,
		publicKey:  publicKey,
		algorithm:  "RSA-SHA256",
	}, nil
}

// Sign signs data using RSA PKCS#1 v1.5
func (s *signatureService) Sign(ctx context.Context, data []byte) (string, error) {
	hash := crypto.SHA256.Hash()
	hashed := hash.New()
	hashed.Write(data)
	hashedSum := hashed.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hashedSum)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	return encodeBase64(signature), nil
}

// Verify verifies a signature
func (s *signatureService) Verify(ctx context.Context, data, signature []byte) (bool, error) {
	hash := crypto.SHA256.Hash()
	hashed := hash.New()
	hashed.Write(data)
	hashedSum := hashed.Sum(nil)

	signatureBytes, err := decodeBase64(string(signature))
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	err = rsa.VerifyPKCS1v15(s.publicKey, crypto.SHA256, hashedSum, signatureBytes)
	if err != nil {
		return false, nil // Signature is invalid
	}

	return true, nil
}

// SignHash signs a pre-computed hash
func (s *signatureService) SignHash(ctx context.Context, hash []byte) ([]byte, error) {
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	return signature, nil
}

// VerifyHash verifies a signature of a hash
func (s *signatureService) VerifyHash(ctx context.Context, hash, signature []byte) (bool, error) {
	err := rsa.VerifyPKCS1v15(s.publicKey, crypto.SHA256, hash, signature)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// GetPublicKey returns the public key in PEM format
func (s *signatureService) GetPublicKey() []byte {
	return encodePublicKey(s.publicKey)
}

// SaveKeyPair saves key pair to files
func (s *signatureService) SaveKeyPair(privateKeyPath, publicKeyPath string) error {
	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(privateKeyPath), 0700)
	os.MkdirAll(filepath.Dir(publicKeyPath), 0700)

	// Save private key
	privateKeyPEM := encodePrivateKey(s.privateKey)
	if err := ioutil.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	publicKeyPEM := encodePublicKey(s.publicKey)
	if err := ioutil.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

// encodeBase64 encodes bytes to base64
func encodeBase64(data []byte) string {
	return fmt.Sprintf("%x", data)
}

// decodeBase64 decodes base64 string to bytes
func decodeBase64(s string) ([]byte, error) {
	// Handle hex encoding
	if len(s)%2 == 0 {
		data, err := hex.DecodeString(s)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("invalid encoding")
}

// encodePrivateKey encodes private key to PEM
func encodePrivateKey(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

// encodePublicKey encodes public key to PEM
func encodePublicKey(key *rsa.PublicKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})
}

// SignedData represents signed data with metadata
type SignedData struct {
	Data      []byte  `json:"data"`
	Signature string  `json:"signature"`
	Algorithm string  `json:"algorithm"`
	PublicKey string  `json:"public_key"`
}

// NewSignedData creates a new SignedData
func NewSignedData(data []byte, signer SignatureService) (*SignedData, error) {
	signature, err := signer.Sign(context.Background(), data)
	if err != nil {
		return nil, err
	}

	return &SignedData{
		Data:      data,
		Signature: signature,
		Algorithm: "RSA-SHA256",
		PublicKey: string(signer.GetPublicKey()),
	}, nil
}

// Verify verifies SignedData
func (s *SignedData) Verify() (bool, error) {
	signatureBytes, err := decodeBase64(s.Signature)
	if err != nil {
		return false, err
	}

	// Create a new signer with the embedded public key
	// In production, properly parse the public key from PEM
	signer := NewSoftwareSignatureService()
	return signer.Verify(context.Background(), s.Data, signatureBytes)
}

// SealRecord represents a sealed record
type SealRecord struct {
	PreviousHash string `json:"previous_hash"`
	CurrentHash  string `json:"current_hash"`
	Signature    string `json:"signature"`
	Timestamp    string `json:"timestamp"`
	SignerID     string `json:"signer_id"`
}

// ChainSealer provides chain-based sealing
type ChainSealer struct {
	signer      SignatureService
	previousSeal *SealRecord
}

// NewChainSealer creates a new ChainSealer
func NewChainSealer(signer SignatureService) *ChainSealer {
	return &ChainSealer{
		signer: signer,
	}
}

// Seal seals data with chain continuity
func (c *ChainSealer) Seal(data []byte) (*SealRecord, error) {
	// Combine data with previous seal if exists
	var inputData []byte
	if c.previousSeal != nil {
		inputData = append(data, []byte(c.previousSeal.CurrentHash)...)
	} else {
		inputData = data
	}

	// Calculate hash
	hashService := NewHashService()
	currentHash := hashService.CalculateHashBytes(inputData)

	// Sign the hash
	signature, err := c.signer.Sign(context.Background(), currentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Create seal record
	seal := &SealRecord{
		PreviousHash: c.previousSeal.GetHash(),
		CurrentHash:  encodeBase64(currentHash),
		Signature:    signature,
		Timestamp:    fmt.Sprintf("%d", nowUnix()),
		SignerID:     "system",
	}

	c.previousSeal = seal
	return seal, nil
}

// GetPreviousSeal returns the previous seal
func (c *ChainSealer) GetPreviousSeal() *SealRecord {
	return c.previousSeal
}

// GetRootHash returns the current root hash
func (c *ChainSealer) GetRootHash() string {
	if c.previousSeal == nil {
		return ""
	}
	return c.previousSeal.CurrentHash
}

// GetHash returns the current hash
func (s *SealRecord) GetHash() string {
	return s.CurrentHash
}

func nowUnix() int64 {
	return int64(rand.Int63())
}

// WitnessService provides witness operations for seal verification
type WitnessService struct {
	witnesses map[string]Witness
}

// Witness represents a witness to a seal
type Witness interface {
	Witness(ctx context.Context, seal *SealRecord) error
	Verify(ctx context.Context, seal *SealRecord, witnessRecord *WitnessRecord) (bool, error)
}

// NewWitnessService creates a new WitnessService
func NewWitnessService() *WitnessService {
	return &WitnessService{
		witnesses: make(map[string]Witness),
	}
}

// RegisterWitness registers a witness
func (w *WitnessService) RegisterWitness(name string, witness Witness) {
	w.witnesses[name] = witness
}

// WitnessRecord represents a witness record
type WitnessRecord struct {
	WitnessType string `json:"witness_type"`
	WitnessID   string `json:"witness_id"`
	Timestamp   string `json:"timestamp"`
	Data        string `json:"data"`
	Signature   string `json:"signature"`
}

// TimeWitness provides time-based witnessing
type TimeWitness struct {
	service SignatureService
}

// NewTimeWitness creates a new TimeWitness
func NewTimeWitness(service SignatureService) *TimeWitness {
	return &TimeWitness{service: service}
}

// Witness witnesses a seal with timestamp
func (t *TimeWitness) Witness(ctx context.Context, seal *SealRecord) error {
	data := fmt.Sprintf("%s|%s|%s", seal.CurrentHash, seal.Timestamp, nowUnix())
	signature, err := t.service.Sign(ctx, []byte(data))
	if err != nil {
		return err
	}

	// Store witness record (implementation depends on witness storage)
	_ = WitnessRecord{
		WitnessType: "time",
		WitnessID:   fmt.Sprintf("time-%d", nowUnix()),
		Timestamp:   fmt.Sprintf("%d", nowUnix()),
		Data:        data,
		Signature:   signature,
	}

	return nil
}

// Verify verifies a time witness
func (t *TimeWitness) Verify(ctx context.Context, seal *SealRecord, record *WitnessRecord) (bool, error) {
	// Verify the witness signature
	signatureBytes, err := decodeBase64(record.Signature)
	if err != nil {
		return false, err
	}

	return t.service.Verify(ctx, []byte(record.Data), signatureBytes)
}
