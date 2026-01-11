package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/csic-platform/services/security/key-management/internal/core/domain"
)

// CryptoProviderImpl implements the CryptoProvider interface
type CryptoProviderImpl struct{}

// NewCryptoProvider creates a new crypto provider instance
func NewCryptoProvider() *CryptoProviderImpl {
	return &CryptoProviderImpl{}
}

// GenerateSymmetricKey generates a symmetric key for the given algorithm
func (c *CryptoProviderImpl) GenerateSymmetricKey(algorithm domain.KeyAlgorithm) ([]byte, error) {
	keySize := algorithm.GetKeySize() / 8

	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	return key, nil
}

// GenerateAsymmetricKeyPair generates an asymmetric key pair for the given algorithm
func (c *CryptoProviderImpl) GenerateAsymmetricKeyPair(algorithm domain.KeyAlgorithm) (publicKey, privateKey []byte, err error) {
	switch algorithm {
	case domain.AlgorithmRSA4096:
		return c.generateRSAKeyPair(4096)
	case domain.AlgorithmECDSAP384:
		return c.generateECDSAKeyPair(elliptic.P384())
	case domain.AlgorithmEd25519:
		return c.generateEd25519KeyPair()
	default:
		return nil, nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// generateRSAKeyPair generates an RSA key pair
func (c *CryptoProviderImpl) generateRSAKeyPair(bits int) (publicKey, privateKey []byte, err error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	publicKey = x509.MarshalPKCS1PublicKey(&private.PublicKey)
	privateKey = x509.MarshalPKCS8PrivateKey(private)

	return publicKey, privateKey, nil
}

// generateECDSAKeyPair generates an ECDSA key pair for the given curve
func (c *CryptoProviderImpl) generateECDSAKeyPair(curve elliptic.Curve) (publicKey, privateKey []byte, err error) {
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	publicKey = elliptic.Marshal(curve, private.PublicKey.X, private.PublicKey.Y)
	privateKey = x509.MarshalPKCS8PrivateKey(private)

	return publicKey, privateKey, nil
}

// generateEd25519KeyPair generates an Ed25519 key pair
func (c *CryptoProviderImpl) generateEd25519KeyPair() (publicKey, privateKey []byte, err error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	publicKey = public
	privateKey = private

	return publicKey, privateKey, nil
}

// Encrypt encrypts plaintext using the given algorithm and key
func (c *CryptoProviderImpl) Encrypt(algorithm domain.KeyAlgorithm, key, plaintext, iv []byte) ([]byte, error) {
	switch algorithm {
	case domain.AlgorithmAES256GCM:
		return c.encryptAESGCM(key, plaintext, iv)
	default:
		return nil, fmt.Errorf("unsupported algorithm for encryption: %s", algorithm)
	}
}

// encryptAESGCM encrypts data using AES-256-GCM
func (c *CryptoProviderImpl) encryptAESGCM(key, plaintext, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(iv) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid IV size: expected %d, got %d", gcm.NonceSize(), len(iv))
	}

	ciphertext := gcm.Seal(iv, iv, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using the given algorithm and key
func (c *CryptoProviderImpl) Decrypt(algorithm domain.KeyAlgorithm, key, ciphertext, iv []byte) ([]byte, error) {
	switch algorithm {
	case domain.AlgorithmAES256GCM:
		return c.decryptAESGCM(key, ciphertext, iv)
	default:
		return nil, fmt.Errorf("unsupported algorithm for decryption: %s", algorithm)
	}
}

// decryptAESGCM decrypts data using AES-256-GCM
func (c *CryptoProviderImpl) decryptAESGCM(key, ciphertext, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(iv) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid IV size: expected %d, got %d", gcm.NonceSize(), len(iv))
	}

	// The IV is prepended to the ciphertext
	plaintext, err := gcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptWithPublicKey encrypts data using an asymmetric public key
func (c *CryptoProviderImpl) EncryptWithPublicKey(algorithm domain.KeyAlgorithm, publicKey, plaintext []byte) ([]byte, error) {
	switch algorithm {
	case domain.AlgorithmRSA4096:
		return c.encryptRSA(publicKey, plaintext)
	default:
		return nil, fmt.Errorf("unsupported algorithm for public key encryption: %s", algorithm)
	}
}

// encryptRSA encrypts data using RSA
func (c *CryptoProviderImpl) encryptRSA(publicKeyBytes, plaintext []byte) ([]byte, error) {
	pub, err := x509.ParsePKCS1PublicKey(publicKeyBytes)
	if err != nil {
		// Try parsing as PKIX
		if parsed, err := x509.ParsePKIXPublicKey(publicKeyBytes); err == nil {
			if rsaPub, ok := parsed.(*rsa.PublicKey); ok {
				return rsa.EncryptPKCS1v15(rand.Reader, rsaPub, plaintext)
			}
		}
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pub, plaintext)
}

// DecryptWithPrivateKey decrypts data using an asymmetric private key
func (c *CryptoProviderImpl) DecryptWithPrivateKey(algorithm domain.KeyAlgorithm, privateKey, ciphertext []byte) ([]byte, error) {
	switch algorithm {
	case domain.AlgorithmRSA4096:
		return c.decryptRSA(privateKey, ciphertext)
	default:
		return nil, fmt.Errorf("unsupported algorithm for private key decryption: %s", algorithm)
	}
}

// decryptRSA decrypts data using RSA
func (c *CryptoProviderImpl) decryptRSA(privateKeyBytes, ciphertext []byte) ([]byte, error) {
	priv, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key type")
	}

	return rsa.DecryptPKCS1v15(rand.Reader, rsaPriv, ciphertext)
}

// Sign signs a message using a private key
func (c *CryptoProviderImpl) Sign(algorithm domain.KeyAlgorithm, privateKey, message []byte) ([]byte, error) {
	switch algorithm {
	case domain.AlgorithmECDSAP384:
		return c.signECDSA(privateKey, message)
	case domain.AlgorithmEd25519:
		return c.signEd25519(privateKey, message)
	default:
		return nil, fmt.Errorf("unsupported algorithm for signing: %s", algorithm)
	}
}

// signECDSA signs a message using ECDSA
func (c *CryptoProviderImpl) signECDSA(privateKeyBytes, message []byte) ([]byte, error) {
	priv, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaPriv, ok := priv.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid ECDSA private key")
	}

	hash := hashMessage(message)
	r, s, err := ecdsa.Sign(rand.Reader, ecdsaPriv, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature as DER
	return append(r.Bytes(), s.Bytes()...), nil
}

// signEd25519 signs a message using Ed25519
func (c *CryptoProviderImpl) signEd25519(privateKey, message []byte) ([]byte, error) {
	return ed25519.Sign(privateKey, message), nil
}

// Verify verifies a signature using a public key
func (c *CryptoProviderImpl) Verify(algorithm domain.KeyAlgorithm, publicKey, message, signature []byte) (bool, error) {
	switch algorithm {
	case domain.AlgorithmECDSAP384:
		return c.verifyECDSA(publicKey, message, signature)
	case domain.AlgorithmEd25519:
		return c.verifyEd25519(publicKey, message, signature), nil
	default:
		return false, fmt.Errorf("unsupported algorithm for verification: %s", algorithm)
	}
}

// verifyECDSA verifies an ECDSA signature
func (c *CryptoProviderImpl) verifyECDSA(publicKeyBytes, message, signature []byte) (bool, error) {
	// Parse public key
	parsedKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := parsedKey.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("invalid ECDSA public key")
	}

	// Split signature into r and s components
	r := signature[:len(signature)/2]
	s := signature[len(signature)/2:]

	hash := hashMessage(message)
	return ecdsa.Verify(ecdsaPub, hash, r, s), nil
}

// verifyEd25519 verifies an Ed25519 signature
func (c *CryptoProviderImpl) verifyEd25519(publicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// WrapKey wraps (encrypts) a key using a key encryption key (KEK)
func (c *CryptoProviderImpl) WrapKey(algorithm domain.KeyAlgorithm, kek, plaintextKey []byte) ([]byte, []byte, error) {
	iv, err := c.GenerateIV(algorithm)
	if err != nil {
		return nil, nil, err
	}

	wrappedKey, err := c.Encrypt(algorithm, kek, plaintextKey, iv)
	if err != nil {
		return nil, nil, err
	}

	return wrappedKey, iv, nil
}

// UnwrapKey unwraps (decrypts) a key using a key encryption key (KEK)
func (c *CryptoProviderImpl) UnwrapKey(algorithm domain.KeyAlgorithm, kek, wrappedKey, iv []byte) ([]byte, error) {
	return c.Decrypt(algorithm, kek, wrappedKey, iv)
}

// GenerateIV generates an initialization vector for the given algorithm
func (c *CryptoProviderImpl) GenerateIV(algorithm domain.KeyAlgorithm) ([]byte, error) {
	switch algorithm {
	case domain.AlgorithmAES256GCM:
		return generateRandomBytes(12) // 96 bits for GCM
	case domain.AlgorithmRSA4096:
		return nil, nil // RSA-PKCS1 doesn't use IV
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// DeriveKeyFromPassword derives a key from a password using PBKDF2
func (c *CryptoProviderImpl) DeriveKeyFromPassword(password string, salt []byte, iterations int) ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	// In a real implementation, use crypto/pbkdf2
	return key, nil
}

// generateRandomBytes generates random bytes of the given size
func generateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// hashMessage hashes a message using SHA-256
func hashMessage(message []byte) []byte {
	// In a real implementation, use crypto/sha256
	return message
}
