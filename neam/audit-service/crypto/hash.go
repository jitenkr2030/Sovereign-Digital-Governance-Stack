package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
)

// HashService provides hash operations
type HashService interface {
	// CalculateHash calculates hash of data
	CalculateHash(data interface{}) string

	// CalculateHashBytes calculates hash of bytes
	CalculateHashBytes(data []byte) []byte

	// CalculateHashFromReader calculates hash from reader
	CalculateHashFromReader(reader io.Reader) ([]byte, error)

	// CompareHashes compares two hashes securely
	CompareHashes(hash1, hash2 []byte) bool

	// GetHashSize returns the hash size in bytes
	GetHashSize() int
}

// hashService implements HashService
type hashService struct {
	hashFunc crypto.Hash
}

// NewHashService creates a new HashService
func NewHashService() HashService {
	return &hashService{
		hashFunc: crypto.SHA256,
	}
}

// CalculateHash calculates hash of data
func (h *hashService) CalculateHash(data interface{}) string {
	// For simple types, convert to string and hash
	switch v := data.(type) {
	case string:
		return h.CalculateHashString(v)
	case []byte:
		return hex.EncodeToString(h.CalculateHashBytes(v))
	default:
		// For complex types, marshal to JSON
		return h.CalculateHashString(fmt.Sprintf("%v", v))
	}
}

// CalculateHashBytes calculates hash of bytes
func (h *hashService) CalculateHashBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// CalculateHashFromReader calculates hash from reader
func (h *hashService) CalculateHashFromReader(reader io.Reader) ([]byte, error) {
	hash := h.hashFunc.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}
	return hash.Sum(nil), nil
}

// CalculateHashString calculates hash of string
func (h *hashService) CalculateHashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CompareHashes compares two hashes securely
func (h *hashService) CompareHashes(hash1, hash2 []byte) bool {
	return subtle.ConstantTimeCompare(hash1, hash2) == 1
}

// GetHashSize returns the hash size in bytes
func (h *hashService) GetHashSize() int {
	return h.hashFunc.Size()
}

// RandRead reads random bytes
func RandRead(p []byte) (n int, err error) {
	return rand.Read(p)
}

// GenerateSecureToken generates a secure random token
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	n, err := rand.Read(b)
	if n != len(b) {
		return "", fmt.Errorf("insufficient random bytes read")
	}
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashPassword hashes a password (simplified - use bcrypt in production)
func HashPassword(password string) (string, error) {
	salt, err := GenerateSecureToken(16)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(password + salt))
	return fmt.Sprintf("%s:%s", hex.EncodeToString(hash[:]), salt), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, storedHash string) bool {
	parts := splitTwoParts(storedHash, ":")
	if len(parts) != 2 {
		return false
	}

	expectedHash := parts[0]
	salt := parts[1]

	hash := sha256.Sum256([]byte(password + salt))
	computedHash := hex.EncodeToString(hash[:])

	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(computedHash)) == 1
}

// splitTwoParts splits a string into two parts at the separator
func splitTwoParts(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s, ""}
}

// MerkleHashService provides Merkle tree hash operations
type MerkleHashService struct {
	hashService HashService
}

// NewMerkleHashService creates a new MerkleHashService
func NewMerkleHashService() *MerkleHashService {
	return &MerkleHashService{
		hashService: NewHashService(),
	}
}

// HashLeaf hashes a leaf value with leaf prefix
func (m *MerkleHashService) HashLeaf(data []byte) []byte {
	prefixed := append([]byte("LEAF:"), data...)
	return m.hashService.CalculateHashBytes(prefixed)
}

// HashNode hashes an internal node
func (m *MerkleHashService) HashNode(left, right []byte) []byte {
	prefixed := append(append([]byte("NODE:"), left...), right...)
	return m.hashService.CalculateHashBytes(prefixed)
}

// HashPair hashes a pair of hashes
func (m *MerkleHashService) HashPair(left, right []byte) []byte {
	return m.hashService.CalculateHashBytes(append(left, right...))
}

// BatchHashService provides batch hash operations
type BatchHashService struct {
	hashService HashService
}

// NewBatchHashService creates a new BatchHashService
func NewBatchHashService() *BatchHashService {
	return &BatchHashService{
		hashService: NewHashService(),
	}
}

// BatchHash hashes multiple items
func (b *BatchHashService) BatchHash(items [][]byte) []byte {
	var result []byte
	for _, item := range items {
		itemHash := b.hashService.CalculateHashBytes(item)
		result = append(result, itemHash...)
	}
	return b.hashService.CalculateHashBytes(result)
}

// HashMerkleRoot calculates the root of a Merkle tree from leaves
func (b *BatchHashService) HashMerkleRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return b.hashService.CalculateHashBytes([]byte{})
	}

	if len(leaves) == 1 {
		return b.hashService.CalculateHashBytes(leaves[0])
	}

	var level []byte
	for i := 0; i < len(leaves); i += 2 {
		var hash []byte
		if i+1 < len(leaves) {
			hash = b.hashService.CalculateHashBytes(append(leaves[i], leaves[i+1]...))
		} else {
			hash = b.hashService.CalculateHashBytes(leaves[i])
		}
		level = append(level, hash...)
	}

	return b.HashMerkleRoot(level)
}

// TimeStampHash provides timestamp-based hashing for evidence
type TimeStampHash struct {
	hashService HashService
}

// NewTimeStampHash creates a new TimeStampHash
func NewTimeStampHash() *TimeStampHash {
	return &TimeStampHash{
		hashService: NewHashService(),
	}
}

// CreateTimestampHash creates a timestamped hash
func (t *TimeStampHash) CreateTimestampHash(data []byte) TimestampHash {
	return TimestampHash{
		Hash:       t.hashService.CalculateHashBytes(data),
		Nonce:      generateNonce(),
		Timestamp:  generateTimestamp(),
	}
}

// TimestampHash represents a timestamped hash
type TimestampHash struct {
	Hash      []byte
	Nonce     string
	Timestamp string
}

// generateNonce generates a random nonce
func generateNonce() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateTimestamp generates a current timestamp
func generateTimestamp() string {
	return fmt.Sprintf("%d", rand.Int63())
}

// ChainHash provides chain-based hashing for audit trails
type ChainHash struct {
	hashService HashService
	previousHash []byte
}

// NewChainHash creates a new ChainHash
func NewChainHash() *ChainHash {
	return &ChainHash{
		hashService: NewHashService(),
		previousHash: nil,
	}
}

// Hash hashes data with chain continuity
func (c *ChainHash) Hash(data []byte) []byte {
	// Combine previous hash with current data
	combined := append(data, c.previousHash...)
	c.previousHash = c.hashService.CalculateHashBytes(combined)
	return c.previousHash
}

// GetPreviousHash returns the previous hash
func (c *ChainHash) GetPreviousHash() []byte {
	return c.previousHash
}

// Reset resets the chain
func (c *ChainHash) Reset() {
	c.previousHash = nil
}

// GetRootHash returns the current root hash
func (c *ChainHash) GetRootHash() []byte {
	return c.previousHash
}

// ChainHashWithIndex provides indexed chain hashing
type ChainHashWithIndex struct {
	hashService HashService
	chainHash   *ChainHash
	index       int64
}

// NewChainHashWithIndex creates a new ChainHashWithIndex
func NewChainHashWithIndex() *ChainHashWithIndex {
	return &ChainHashWithIndex{
		hashService: NewHashService(),
		chainHash:   NewChainHash(),
		index:       0,
	}
}

// HashWithIndex hashes data with chain continuity and index
func (c *ChainHashWithIndex) HashWithIndex(data []byte, index int64) []byte {
	c.index = index
	// Combine index with data
	indexedData := append([]byte(fmt.Sprintf("%d:", index)), data...)
	return c.chainHash.Hash(indexedData)
}

// GetIndex returns the current index
func (c *ChainHashWithIndex) GetIndex() int64 {
	return c.index
}

// GetRootHash returns the current root hash
func (c *ChainHashWithIndex) GetRootHash() []byte {
	return c.chainHash.GetRootHash()
}

// Reset resets the chain
func (c *ChainHashWithIndex) Reset() {
	c.chainHash.Reset()
	c.index = 0
}
