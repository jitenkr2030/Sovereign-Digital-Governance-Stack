// Crypto Utilities Package - Cryptographic operations for CSIC Platform
// Hashing, signature verification, and data encryption utilities

package cryptoutils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"hash"
	"strings"
)

// HashAlgorithm represents the supported hashing algorithms
type HashAlgorithm string

const (
	HashSHA256   HashAlgorithm = "SHA256"
	HashSHA512   HashAlgorithm = "SHA512"
	HashSHA3_256 HashAlgorithm = "SHA3_256"
	HashSHA3_512 HashAlgorithm = "SHA3_512"
)

// HashResult represents the result of a hashing operation
type HashResult struct {
	Algorithm HashAlgorithm
	Value     string
	Length    int
}

// GenerateHash creates a hash of the input data using the specified algorithm
func GenerateHash(data []byte, algorithm HashAlgorithm) (*HashResult, error) {
	var h hash.Hash
	var result []byte

	switch algorithm {
	case HashSHA256:
		h = sha256.New()
	case HashSHA512:
		h = sha512.New()
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	h.Write(data)
	result = h.Sum(nil)

	return &HashResult{
		Algorithm: algorithm,
		Value:     hex.EncodeToString(result),
		Length:    len(result) * 8,
	}, nil
}

// GenerateSHA256 creates a SHA-256 hash of the input data
func GenerateSHA256(data []byte) (*HashResult, error) {
	return GenerateHash(data, HashSHA256)
}

// GenerateSHA512 creates a SHA-512 hash of the input data
func GenerateSHA512(data []byte) (*HashResult, error) {
	return GenerateHash(data, HashSHA512)
}

// HashString creates a SHA-256 hash of a string and returns hex-encoded result
func HashString(input string) string {
	hash, _ := GenerateSHA256([]byte(input))
	return hash.Value
}

// VerifyHash compares the input data against an expected hash
func VerifyHash(data []byte, expectedHash string, algorithm HashAlgorithm) (bool, error) {
	computed, err := GenerateHash(data, algorithm)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(computed.Value, expectedHash), nil
}

// DigitalSignature represents a digital signature
type DigitalSignature struct {
	Algorithm string
	Signature []byte
	PublicKey []byte
}

// SignData creates a digital signature of the data using RSA-SHA256
func SignData(privateKey *rsa.PrivateKey, data []byte) (*DigitalSignature, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is required")
	}

	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return &DigitalSignature{
		Algorithm: "RSA-SHA256",
		Signature: signature,
		PublicKey: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	}, nil
}

// VerifySignature verifies a digital signature
func VerifySignature(publicKey *rsa.PublicKey, data []byte, signature []byte) (bool, error) {
	if publicKey == nil || signature == nil {
		return false, fmt.Errorf("public key and signature are required")
	}

	hash := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GenerateRSAKeyPair generates a new RSA key pair
func GenerateRSAKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if keySize < 2048 {
		return nil, nil, fmt.Errorf("key size must be at least 2048 bits for security")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	return privateKey, &privateKey.PublicKey, nil
}

// ExportPrivateKey exports a private key to PEM format
func ExportPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	bytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: bytes,
	}
	return pem.EncodeToMemory(block), nil
}

// ExportPublicKey exports a public key to PEM format
func ExportPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	bytes := x509.MarshalPKCS1PublicKey(publicKey)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes,
	}
	return pem.EncodeToMemory(block), nil
}

// ImportPrivateKey imports a private key from PEM format
func ImportPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

// ImportPublicKey imports a public key from PEM format
func ImportPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// PBKDF2Hash applies PBKDF2 key derivation to a password
func PBKDF2Hash(password, salt []byte, iterations, keyLength int) ([]byte, error) {
	if iterations < 10000 {
		return nil, fmt.Errorf("iterations must be at least 10000 for security")
	}

	hash := crypto.PBKDF2Key(
		crypto.SHA512,
		password,
		salt,
		iterations,
		keyLength,
	)

	return hash, nil
}

// GenerateSalt generates a cryptographically secure salt
func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// MaskSensitiveData masks sensitive data for logging purposes
func MaskSensitiveData(data string, visiblePrefix, visibleSuffix int) string {
	if len(data) <= visiblePrefix+visibleSuffix {
		return strings.Repeat("*", len(data))
	}
	return data[:visiblePrefix] + strings.Repeat("*", len(data)-visiblePrefix-visibleSuffix) + data[len(data)-visibleSuffix:]
}

// MaskAddress masks a cryptocurrency address for display
func MaskAddress(address string) string {
	if len(address) <= 8 {
		return strings.Repeat("*", len(address))
	}
	return address[:4] + strings.Repeat("*", len(address)-8) + address[len(address)-4:]
}

// ValidateAddressFormat validates the format of a cryptocurrency address
func ValidateAddressFormat(address string, prefix string) bool {
	if !strings.HasPrefix(address, prefix) {
		return false
	}
	if len(address) < 26 || len(address) > 35 {
		return false
	}
	return true
}

// Checksum validates the checksum of an address
func ValidateAddressChecksum(address string) bool {
	// Simplified checksum validation
	// In production, use proper base58 or bech32 validation
	return len(address) >= 26 && len(address) <= 35
}

// MerkleProof represents a Merkle proof for transaction verification
type MerkleProof struct {
	RootHash     string
	LeafHash     string
	Proof        []string
	ProofFlags   []bool
	TargetHash   string
	TargetIndex  int
}

// VerifyMerkleProof verifies a Merkle proof
func VerifyMerkleProof(proof MerkleProof) bool {
	if len(proof.Proof) == 0 {
		return proof.RootHash == proof.LeafHash
	}

	hash := proof.LeafHash
	for i, sibling := range proof.Proof {
		if hash < sibling {
			hash = HashString(hash + sibling)
		} else {
			hash = HashString(sibling + hash)
		}
		if i < len(proof.ProofFlags) && !proof.ProofFlags[i] {
			hash = proof.LeafHash
		}
	}

	return hash == proof.RootHash
}
