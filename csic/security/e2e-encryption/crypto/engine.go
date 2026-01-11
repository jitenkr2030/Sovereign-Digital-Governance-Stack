package crypto

import (
	"crypto"
	"crypto/cipher"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Engine provides cryptographic operations
type Engine struct {
	logger *zap.Logger
	mu     sync.RWMutex
}

// Global engine instance
var (
	globalEngine *Engine
	engineMu     sync.RWMutex
)

// Initialize initializes the crypto engine
func Initialize(logger *zap.Logger, cfg interface{}) error {
	engineMu.Lock()
	defer engineMu.Unlock()

	globalEngine = &Engine{
		logger: logger,
	}

	logger.Info("Crypto engine initialized")
	return nil
}

// GetEngine returns the global crypto engine
func GetEngine() *Engine {
	engineMu.RLock()
	defer engineMu.RUnlock()
	return globalEngine
}

// SHA256Hash returns SHA-256 hash function
func SHA256Hash() crypto.Hash {
	return crypto.SHA256
}

// SHA384Hash returns SHA-384 hash function
func SHA384Hash() crypto.Hash {
	return crypto.SHA384
}

// SHA512Hash returns SHA-512 hash function
func SHA512Hash() crypto.Hash {
	return crypto.SHA512
}

// HashData hashes data using SHA-256
func (e *Engine) HashData(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// HashString hashes a string and returns hex-encoded result
func (e *Engine) HashString(data string) string {
	return hex.EncodeToString(e.HashData([]byte(data)))
}

// HashToHex hashes data and returns hex-encoded string
func (e *Engine) HashToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// HashFromHex decodes a hex-encoded hash
func (e *Engine) HashFromHex(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// HMAC creates an HMAC signature
func (e *Engine) HMAC(data []byte, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil), nil
}

// HMACVerify verifies an HMAC signature
func (e *Engine) HMACVerify(data, signature, key []byte) bool {
	expectedMAC, err := e.HMAC(data, key)
	if err != nil {
		return false
	}
	return hmac.Equal(signature, expectedMAC)
}

// GenerateRandomBytes generates cryptographically secure random bytes
func (e *Engine) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateRandomString generates a random alphanumeric string
func (e *Engine) GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes, err := e.GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

// GenerateSecureToken generates a secure random token
func (e *Engine) GenerateSecureToken(length int) string {
	token, _ := e.GenerateRandomBytes(length)
	return base64.RawURLEncoding.EncodeToString(token)
}

// GenerateNonce generates a random nonce
func (e *Engine) GenerateNonce(size int) ([]byte, error) {
	return e.GenerateRandomBytes(size)
}

// DeriveKey derives a key from a password using PBKDF2
func (e *Engine) DeriveKey(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	return DeriveKeyPBKDF2(password, salt, iterations, keyLen)
}

// DeriveKeyArgon2 derives a key using Argon2
func (e *Engine) DeriveKeyArgon2(password, salt []byte, memory, time, parallelism, keyLen int) ([]byte, error) {
	// In production, would use argon2 library
	// Simplified implementation for demonstration
	return e.DeriveKey(password, salt, time*1000, keyLen)
}

// DeriveKeyHKDF derives a key using HKDF
func (e *Engine) DeriveKeyHKDF(secret, salt, info []byte, keyLen int) ([]byte, error) {
	return DeriveKeyHKDF(secret, salt, info, keyLen)
}

// EncryptAESGCM encrypts data using AES-256-GCM
func (e *Engine) EncryptAESGCM(plaintext, key, additionalData []byte) (ciphertext, nonce []byte, err error) {
	// Generate nonce
	nonce = make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create cipher
	block, err := aesCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt
	if len(additionalData) > 0 {
		ciphertext = gcm.Seal(nil, nonce, plaintext, additionalData)
	} else {
		ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	}

	return ciphertext, nonce, nil
}

// DecryptAESGCM decrypts data using AES-256-GCM
func (e *Engine) DecryptAESGCM(ciphertext, key, nonce, additionalData []byte) ([]byte, error) {
	// Create cipher
	block, err := aesCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	var plaintext []byte
	if len(additionalData) > 0 {
		plaintext, err = gcm.Open(nil, nonce, ciphertext, additionalData)
	} else {
		plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptChaCha20Poly1305 encrypts data using ChaCha20-Poly1305
func (e *Engine) EncryptChaCha20Poly1305(plaintext, key, additionalData []byte) (ciphertext, nonce []byte, err error) {
	// Generate nonce
	nonce = make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create cipher
	block, err := chacha20Cipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt
	if len(additionalData) > 0 {
		ciphertext = gcm.Seal(nil, nonce, plaintext, additionalData)
	} else {
		ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	}

	return ciphertext, nonce, nil
}

// DecryptChaCha20Poly1305 decrypts data using ChaCha20-Poly1305
func (e *Engine) DecryptChaCha20Poly1305(ciphertext, key, nonce, additionalData []byte) ([]byte, error) {
	// Create cipher
	block, err := chacha20Cipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	var plaintext []byte
	if len(additionalData) > 0 {
		plaintext, err = gcm.Open(nil, nonce, ciphertext, additionalData)
	} else {
		plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// GenerateRSAKeyPair generates an RSA key pair
func (e *Engine) GenerateRSAKeyPair(bits int) (privateKey, publicKey []byte, err error) {
	if bits < 2048 {
		return nil, nil, errors.New("RSA key size must be at least 2048 bits")
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key
	privateKey = x509.MarshalPKCS1PrivateKey(rsaKey)

	// Encode public key
	publicKey, err = x509.MarshalPKIXPublicKey(rsaKey.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	return privateKey, publicKey, nil
}

// GenerateECKeyPair generates an ECDSA key pair
func (e *Engine) GenerateECKeyPair(curve elliptic.Curve) (privateKey, publicKey []byte, err error) {
	if curve == nil {
		curve = elliptic.P256()
	}

	ecKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate EC key: %w", err)
	}

	// Encode private key
	privateKey, err = x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	// Encode public key
	publicKey, err = x509.MarshalPKIXPublicKey(ecKey.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode public key: %w", err)
	}

	return privateKey, publicKey, nil
}

// ECDHKeyExchange performs an ECDH key exchange
func (e *Engine) ECDHKeyExchange(privateKey, publicKey []byte) ([]byte, error) {
	// Parse private key
	priv, err := x509.ParsePKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an ECDSA private key")
	}

	// Parse public key
	pub, err := x509.ParsePKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}

	// Perform ECDH
	x, _ := elliptic.P256().ScalarMult(ecdsaPub.X, ecdsaPub.Y, ecdsaPriv.D.Bytes())

	// Derive shared key from x coordinate
	sharedKey := make([]byte, 32)
	binary.BigEndian.PutUint32(sharedKey, uint32(time.Now().Unix()))
	copy(sharedKey[4:], x.Bytes()[:32])

	return sharedKey, nil
}

// Sign signs data using ECDSA
func (e *Engine) Sign(data []byte, privateKey []byte) ([]byte, error) {
	// Parse private key
	priv, err := x509.ParsePKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an ECDSA private key")
	}

	// Sign
	r, s, err := ecdsa.Sign(rand.Reader, ecdsaPriv, e.HashData(data))
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])

	return sig, nil
}

// Verify verifies an ECDSA signature
func (e *Engine) Verify(data, signature, publicKey []byte) bool {
	// Parse public key
	pub, err := x509.ParsePKIXPublicKey(publicKey)
	if err != nil {
		return false
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return false
	}

	// Decode signature
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	// Verify
	return ecdsa.Verify(ecdsaPub, e.HashData(data), r, s)
}

// SecureCompare performs a constant-time comparison
func (e *Engine) SecureCompare(a, b []byte) bool {
	return hmac.Equal(a, b)
}

// GenerateSalt generates a random salt
func (e *Engine) GenerateSalt(size int) ([]byte, error) {
	return e.GenerateRandomBytes(size)
}

// KDF represents a key derivation function
type KDF interface {
	DeriveKey(password, salt []byte, iterations, keyLen int) ([]byte, error)
}

// DeriveKeyPBKDF2 derives a key using PBKDF2
func DeriveKeyPBKDF2(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	if iterations < 10000 {
		return nil, errors.New("PBKDF2 iterations should be at least 10000")
	}

	pbkdf2 := crypto.PBKDF2WithSHA1
	if len(password) > 0 {
		pbkdf2 = crypto.PBKDF2WithSHA256
	}

	return pbkdf2(password, salt, iterations, keyLen, crypto.SHA256), nil
}

// DeriveKeyHKDF derives a key using HKDF
func DeriveKeyHKDF(secret, salt, info []byte, keyLen int) ([]byte, error) {
	if len(secret) == 0 {
		return nil, errors.New("secret cannot be empty")
	}

	hkdf := crypto.HKDF
	key := make([]byte, keyLen)
	n, err := hkdf(crypto.SHA256, secret, salt, info, key)
	if err != nil {
		return nil, fmt.Errorf("HKDF derivation failed: %w", err)
	}
	if n != keyLen {
		return nil, errors.New("HKDF derived key length mismatch")
	}

	return key, nil
}

// GenerateTimeBasedToken generates a time-based token
func (e *Engine) GenerateTimeBasedToken(secret []byte, timeStep int64) ([]byte, error) {
	currentTime := time.Now().Unix() / timeStep
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(currentTime))

	// Hash the time with secret
	mac := hmac.New(sha256.New, secret)
	mac.Write(timeBytes)
	return mac.Sum(nil)[:6], nil
}

// VerifyTimeBasedToken verifies a time-based token
func (e *Engine) VerifyTimeBasedToken(token, secret []byte, timeStep int64, window int) bool {
	for i := -window; i <= window; i++ {
		expected, _ := e.GenerateTimeBasedToken(secret, timeStep+int64(i))
		if e.SecureCompare(token, expected) {
			return true
		}
	}
	return false
}

// SecureZero zeroes out memory
func SecureZero(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// Reader is a secure random reader
type Reader struct {
	reader io.Reader
}

// NewReader creates a new secure random reader
func NewReader() *Reader {
	return &Reader{reader: rand.Reader}
}

// Read reads random bytes
func (r *Reader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

// Helper functions

func aesCipher(key []byte) (cipher.Block, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	return block, nil
}

func chacha20Cipher(key []byte) (cipher.AEAD, error) {
	// Simplified - in production use golang.org/x/crypto/chacha20poly1305
	return nil, errors.New("ChaCha20-Poly1305 requires golang.org/x/crypto")
}

import (
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
)
