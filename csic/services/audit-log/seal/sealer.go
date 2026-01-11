// Audit Log Sealer - Cryptographic Chain Sealing
// Provides Merkle tree-based chain sealing for tamper-evident audit logs

package seal

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/csic-platform/services/audit-log"
)

// AuditLogSealer handles cryptographic sealing of audit log chains
type AuditLogSealer struct {
	chainFilePath  string
	sealInterval   int
	privateKey     *rsa.PrivateKey
	publicKeyPEM   []byte
	mu             sync.RWMutex
	pendingEntries []*audit.AuditLogEntry
	lastSealTime   time.Time
	sealMu         sync.Mutex
}

// NewAuditLogSealer creates a new audit log sealer
func NewAuditLogSealer(chainFilePath string, sealInterval int) *AuditLogSealer {
	sealer := &AuditLogSealer{
		chainFilePath:  chainFilePath,
		sealInterval:   sealInterval,
		pendingEntries: make([]*audit.AuditLogEntry, 0),
	}

	// Ensure chain directory exists
	if err := os.MkdirAll(filepath.Dir(chainFilePath), 0700); err != nil {
		panic(fmt.Sprintf("failed to create chain directory: %v", err))
	}

	// Generate or load signing key
	if err := sealer.loadOrGenerateKey(); err != nil {
		panic(fmt.Sprintf("failed to load or generate key: %v", err))
	}

	return sealer
}

// SealPending seals all pending entries in the chain
func (s *AuditLogSealer) SealPending(currentSequence uint64) error {
	s.sealMu.Lock()
	defer s.sealMu.Unlock()

	if len(s.pendingEntries) == 0 {
		return nil
	}

	// Get the chain information
	chain, err := s.loadLatestChain()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to load latest chain: %w", err)
	}

	// Build Merkle tree from pending entries
	merkleRoot := s.buildMerkleTree(s.pendingEntries)

	// Create chain entry
	now := time.Now().UTC()
	sealInput := SealInput{
		ChainID:       getChainID(),
		SequenceStart: s.pendingEntries[0].SequenceNum,
		SequenceEnd:   s.pendingEntries[len(s.pendingEntries)-1].SequenceNum,
		EntryCount:    len(s.pendingEntries),
		RootHash:      merkleRoot,
		PreviousChain: "",
		Timestamp:     now,
	}

	if chain != nil {
		sealInput.PreviousChain = chain.ChainID
	}

	// Create seal signature
	signature, err := s.createSealSignature(sealInput)
	if err != nil {
		return fmt.Errorf("failed to create seal signature: %w", err)
	}

	// Create chain record
	newChain := &audit.AuditChain{
		ChainID:        sealInput.ChainID,
		CreatedAt:      now,
		SealedAt:       now,
		SequenceStart:  sealInput.SequenceStart,
		SequenceEnd:    sealInput.SequenceEnd,
		EntryCount:     sealInput.EntryCount,
		RootHash:       merkleRoot,
		SealSignature:  signature,
		PreviousChain:  sealInput.PreviousChain,
	}

	// Save chain record
	if err := s.saveChain(newChain); err != nil {
		return fmt.Errorf("failed to save chain: %w", err)
	}

	// Clear pending entries
	s.pendingEntries = make([]*audit.AuditLogEntry, 0)
	s.lastSealTime = now

	return nil
}

// AddEntry adds an entry to the pending seal list
func (s *AuditLogSealer) AddEntry(entry *audit.AuditLogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingEntries = append(s.pendingEntries, entry)
}

// GetChain retrieves a specific chain record
func (s *AuditLogSealer) GetChain(chainID string) (*audit.AuditChain, error) {
	chain, err := s.loadChain(chainID)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

// ExportChain exports a sealed chain for legal discovery
func (s *AuditLogSealer) ExportChain(ctx context.Context, chainID string) ([]byte, error) {
	chain, err := s.loadChain(chainID)
	if err != nil {
		return nil, err
	}

	export := ChainExport{
		Chain:        chain,
		ExportedAt:   time.Now().UTC(),
		PublicKey:    string(s.publicKeyPEM),
	}

	return json.MarshalIndent(export, "", "  ")
}

// GetPublicKey returns the public key for verification
func (s *AuditLogSealer) GetPublicKey() []byte {
	return s.publicKeyPEM
}

// loadOrGenerateKey loads or generates the signing key
func (s *AuditLogSealer) loadOrGenerateKey() error {
	keyFile := filepath.Join(filepath.Dir(s.chainFilePath), "sealer_private.key")

	// Try to load existing key
	data, err := os.ReadFile(keyFile)
	if err == nil {
		key, err := x509.ParsePKCS1PrivateKey(data)
		if err == nil {
			s.privateKey = key
			s.publicKeyPEM = s.encodePublicKey(&key.PublicKey)
			return nil
		}
	}

	// Generate new key
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	s.privateKey = key
	s.publicKeyPEM = s.encodePublicKey(&key.PublicKey)

	// Save private key
	if err := os.WriteFile(keyFile, x509.MarshalPKCS1PrivateKey(key), 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	return nil
}

// encodePublicKey encodes a public key to PEM format
func (s *AuditLogSealer) encodePublicKey(key *rsa.PublicKey) []byte {
	bytes := x509.MarshalPKCS1PublicKey(key)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: bytes,
	})
}

// buildMerkleTree builds a Merkle tree from entries and returns the root hash
func (s *AuditLogSealer) buildMerkleTree(entries []*audit.AuditLogEntry) string {
	if len(entries) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	// Create leaf hashes
	hashes := make([]string, len(entries))
	for i, entry := range entries {
		hashes[i] = entry.CurrentHash
	}

	// Build Merkle tree
	for len(hashes) > 1 {
		var nextLevel []string
		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i]
			right := left
			if i+1 < len(hashes) {
				right = hashes[i+1]
			}

			// Combine and hash
			combined := left + right
			hash := sha256.Sum256([]byte(combined))
			nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
		}
		hashes = nextLevel
	}

	return hashes[0]
}

// createSealSignature creates a cryptographic signature for the seal
func (s *AuditLogSealer) createSealSignature(input SealInput) (string, error) {
	// Create canonical representation
	data, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	// Hash the input
	hash := sha256.Sum256(data)

	// Sign with private key
	signature, err := s.privateKey.Sign(rand.Reader, hash[:], crypto.SHA256)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signature), nil
}

// verifySealSignature verifies a seal signature
func (s *AuditLogSealer) verifySealSignature(chain *audit.AuditChain) (bool, error) {
	// Decode public key
	block, _ := pem.Decode(s.publicKeyPEM)
	if block == nil {
		return false, errors.New("failed to decode public key")
	}

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return false, err
	}

	// Verify signature
	signature, err := hex.DecodeString(chain.SealSignature)
	if err != nil {
		return false, err
	}

	input := SealInput{
		ChainID:       chain.ChainID,
		SequenceStart: chain.SequenceStart,
		SequenceEnd:   chain.SequenceEnd,
		EntryCount:    chain.EntryCount,
		RootHash:      chain.RootHash,
		PreviousChain: chain.PreviousChain,
	}

	data, err := json.Marshal(input)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature)
	return err == nil, err
}

// loadLatestChain loads the most recent chain record
func (s *AuditLogSealer) loadLatestChain() (*audit.AuditChain, error) {
	chainsDir := filepath.Join(filepath.Dir(s.chainFilePath), "chains")
	entries, err := os.ReadDir(chainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var latestChain *audit.AuditChain
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		chain, err := s.loadChain(entry.Name())
		if err != nil {
			continue
		}

		if chain.SealedAt.After(latestTime) {
			latestTime = chain.SealedAt
			latestChain = chain
		}
	}

	return latestChain, nil
}

// loadChain loads a specific chain record
func (s *AuditLogSealer) loadChain(chainID string) (*audit.AuditChain, error) {
	chainFile := filepath.Join(filepath.Dir(s.chainFilePath), "chains", chainID+".json")

	data, err := os.ReadFile(chainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain file: %w", err)
	}

	var chain audit.AuditChain
	if err := json.Unmarshal(data, &chain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chain: %w", err)
	}

	return &chain, nil
}

// saveChain saves a chain record
func (s *AuditLogSealer) saveChain(chain *audit.AuditChain) error {
	chainsDir := filepath.Join(filepath.Dir(s.chainFilePath), "chains")
	if err := os.MkdirAll(chainsDir, 0700); err != nil {
		return fmt.Errorf("failed to create chains directory: %w", err)
	}

	data, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chain: %w", err)
	}

	chainFile := filepath.Join(chainsDir, chain.ChainID+".json")
	if err := os.WriteFile(chainFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write chain file: %w", err)
	}

	return nil
}

// SealInput represents the input data for creating a seal
type SealInput struct {
	ChainID        string
	SequenceStart  uint64
	SequenceEnd    uint64
	EntryCount     int
	RootHash       string
	PreviousChain  string
	Timestamp      time.Time
}

// ChainExport represents an exported chain for legal discovery
type ChainExport struct {
	Chain      *audit.AuditChain `json:"chain"`
	ExportedAt time.Time         `json:"exported_at"`
	PublicKey  string            `json:"public_key"`
}

// getChainID generates a unique chain identifier
func getChainID() string {
	return fmt.Sprintf("chain_%d", time.Now().UnixNano())
}
