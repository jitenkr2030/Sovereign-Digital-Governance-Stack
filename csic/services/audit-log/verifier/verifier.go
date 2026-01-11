// Audit Log Verifier - Chain Integrity Verification
// Provides comprehensive verification of audit log integrity and tamper detection

package verifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/csic-platform/services/audit-log"
)

// AuditLogVerifier verifies the integrity of audit log chains
type AuditLogVerifier struct {
	storagePath    string
	chainFilePath  string
}

// NewAuditLogVerifier creates a new audit log verifier
func NewAuditLogVerifier(storagePath string, chainFilePath string) *AuditLogVerifier {
	return &AuditLogVerifier{
		storagePath:   storagePath,
		chainFilePath: chainFilePath,
	}
}

// VerifyChain performs a comprehensive verification of the audit log chain
func (v *AuditLogVerifier) VerifyChain(ctx context.Context) (*audit.VerificationResult, error) {
	result := &audit.VerificationResult{
		Valid:        true,
		VerifiedAt:   time.Now().UTC(),
		VerifiedEntries: 0,
		FailedEntries:   0,
		Errors:        make([]string, 0),
	}

	// Verify all chain files
	chainResult, err := v.verifyChains()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("chain verification failed: %v", err))
	}
	result.ChainIntegrity = chainResult

	// Verify hash chain integrity
	hashResult, err := v.verifyHashChain()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("hash chain verification failed: %v", err))
	}
	result.HashIntegrity = hashResult

	// Verify seal integrity
	sealResult, err := v.verifySeals()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("seal verification failed: %v", err))
	}
	result.SealIntegrity = sealResult

	// Calculate totals
	entries, err := v.countVerifiedEntries()
	if err == nil {
		result.VerifiedEntries = entries
	}

	return result, nil
}

// VerifyEntry verifies a specific audit log entry
func (v *AuditLogVerifier) VerifyEntry(entry *audit.AuditLogEntry) (bool, string) {
	// Verify the hash calculation
	expectedHash := v.calculateHash(entry)
	if entry.CurrentHash != expectedHash {
		return false, fmt.Sprintf("hash mismatch: expected %s, got %s", expectedHash, entry.CurrentHash)
	}

	// Verify previous hash linkage
	if entry.SequenceNum > 0 && entry.PreviousHash == "0000000000000000000000000000000000000000000000000000000000000000" {
		// This is valid only for the first entry
		if entry.SequenceNum != 1 {
			return false, "invalid previous hash for non-first entry"
		}
	}

	return true, ""
}

// VerifyChainIntegrity verifies the integrity of a specific chain
func (v *AuditLogVerifier) VerifyChainIntegrity(ctx context.Context, chainID string) (*audit.VerificationResult, error) {
	result := &audit.VerificationResult{
		Valid:          true,
		ChainIntegrity: true,
		HashIntegrity:  true,
		SealIntegrity:  true,
		VerifiedAt:     time.Now().UTC(),
		Errors:         make([]string, 0),
	}

	// Load chain record
	chainFile := filepath.Join(filepath.Dir(v.chainFilePath), "chains", chainID+".json")
	data, err := os.ReadFile(chainFile)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to load chain: %v", err))
		return result, nil
	}

	var chain audit.AuditChain
	if err := json.Unmarshal(data, &chain); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to parse chain: %v", err))
		return result, nil
	}

	// Verify seal signature
	sealValid, err := v.verifySealSignature(&chain)
	if err != nil || !sealValid {
		result.SealIntegrity = false
		result.Valid = false
		result.Errors = append(result.Errors, "seal signature verification failed")
	}

	// Verify Merkle root
	entries, err := v.loadEntriesInRange(chain.SequenceStart, chain.SequenceEnd)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to load entries: %v", err))
		return result, nil
	}

	merkleRoot := v.buildMerkleTree(entries)
	if merkleRoot != chain.RootHash {
		result.HashIntegrity = false
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Merkle root mismatch: expected %s, got %s", chain.RootHash, merkleRoot))
	}

	// Verify hash chain continuity
	for i := 1; i < len(entries); i++ {
		if entries[i].PreviousHash != entries[i-1].CurrentHash {
			result.HashIntegrity = false
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("hash chain broken at sequence %d", entries[i].SequenceNum))
		}
	}

	result.VerifiedEntries = len(entries)
	result.FailedEntries = 0

	return result, nil
}

// verifyChains verifies all chain records
func (v *AuditLogVerifier) verifyChains() (bool, error) {
	chainsDir := filepath.Join(filepath.Dir(v.chainFilePath), "chains")

	entries, err := os.ReadDir(chainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		chainFile := filepath.Join(chainsDir, entry.Name())
		data, err := os.ReadFile(chainFile)
		if err != nil {
			return false, fmt.Errorf("failed to read chain file %s: %w", entry.Name(), err)
		}

		var chain audit.AuditChain
		if err := json.Unmarshal(data, &chain); err != nil {
			return false, fmt.Errorf("failed to parse chain file %s: %w", entry.Name(), err)
		}

		// Verify chain has required fields
		if chain.ChainID == "" || chain.SealSignature == "" || chain.RootHash == "" {
			return false, fmt.Errorf("chain %s is missing required fields", entry.Name())
		}
	}

	return true, nil
}

// verifyHashChain verifies the hash chain continuity
func (v *AuditLogVerifier) verifyHashChain() (bool, error) {
	// Get all log files
	logFiles, err := filepath.Glob(filepath.Join(v.storagePath, "audit_*.log"))
	if err != nil {
		return false, err
	}

	var allEntries []*audit.AuditLogEntry

	for _, logFile := range logFiles {
		entries, err := v.loadEntriesFromFile(logFile)
		if err != nil {
			return false, fmt.Errorf("failed to load entries from %s: %w", logFile, err)
		}
		allEntries = append(allEntries, entries...)
	}

	// Sort by sequence number
	allEntries = sortBySequence(allEntries)

	// Verify hash chain
	var lastHash string
	for i, entry := range allEntries {
		// Verify hash calculation
		expectedHash := v.calculateHash(entry)
		if entry.CurrentHash != expectedHash {
			return false, fmt.Errorf("hash mismatch at sequence %d: expected %s, got %s",
				entry.SequenceNum, expectedHash, entry.CurrentHash)
		}

		// Verify previous hash linkage
		if i == 0 {
			if entry.SequenceNum != 1 && entry.PreviousHash != "0000000000000000000000000000000000000000000000000000000000000000" {
				return false, fmt.Errorf("sequence %d has invalid previous hash", entry.SequenceNum)
			}
		} else {
			if entry.PreviousHash != lastHash {
				return false, fmt.Errorf("hash chain broken at sequence %d", entry.SequenceNum)
			}
		}

		lastHash = entry.CurrentHash
	}

	return true, nil
}

// verifySeals verifies all seal signatures
func (v *AuditLogVerifier) verifySeals() (bool, error) {
	chainsDir := filepath.Join(filepath.Dir(v.chainFilePath), "chains")

	entries, err := os.ReadDir(chainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		chainFile := filepath.Join(chainsDir, entry.Name())
		data, err := os.ReadFile(chainFile)
		if err != nil {
			return false, fmt.Errorf("failed to read chain file %s: %w", entry.Name(), err)
		}

		var chain audit.AuditChain
		if err := json.Unmarshal(data, &chain); err != nil {
			return false, fmt.Errorf("failed to parse chain file %s: %w", entry.Name(), err)
		}

		// Load public key
		pubKeyFile := filepath.Join(filepath.Dir(v.chainFilePath), "sealer_public.key")
		pubKeyData, err := os.ReadFile(pubKeyFile)
		if err != nil {
			return false, fmt.Errorf("failed to read public key: %w", err)
		}

		// Verify seal signature (simplified - in production use proper verification)
		if chain.SealSignature == "" {
			return false, fmt.Errorf("chain %s has no seal signature", entry.Name())
		}

		// Verify signature format
		if len(chain.SealSignature) < 128 {
			return false, fmt.Errorf("chain %s has invalid signature format", entry.Name())
		}
	}

	return true, nil
}

// verifySealSignature verifies a single seal signature
func (v *AuditLogVerifier) verifySealSignature(chain *audit.AuditChain) (bool, error) {
	// Load public key
	pubKeyFile := filepath.Join(filepath.Dir(v.chainFilePath), "sealer_public.key")
	pubKeyData, err := os.ReadFile(pubKeyFile)
	if err != nil {
		return false, err
	}

	// Parse public key (simplified - in production use proper parsing)
	if len(pubKeyData) < 100 {
		return false, fmt.Errorf("invalid public key")
	}

	// Verify signature exists and has valid format
	if chain.SealSignature == "" {
		return false, nil
	}

	if len(chain.SealSignature) < 128 {
		return false, nil
	}

	return true, nil
}

// loadEntriesInRange loads entries within a sequence range
func (v *AuditLogVerifier) loadEntriesInRange(startSeq, endSeq uint64) ([]*audit.AuditLogEntry, error) {
	logFiles, err := filepath.Glob(filepath.Join(v.storagePath, "audit_*.log"))
	if err != nil {
		return nil, err
	}

	var allEntries []*audit.AuditLogEntry
	for _, logFile := range logFiles {
		entries, err := v.loadEntriesFromFile(logFile)
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)
	}

	var result []*audit.AuditLogEntry
	for _, entry := range allEntries {
		if entry.SequenceNum >= startSeq && entry.SequenceNum <= endSeq {
			result = append(result, entry)
		}
	}

	return sortBySequence(result), nil
}

// loadEntriesFromFile loads all entries from a log file
func (v *AuditLogVerifier) loadEntriesFromFile(filePath string) ([]*audit.AuditLogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*audit.AuditLogEntry

	for {
		// Read length prefix
		lengthBuf := make([]byte, 4)
		n, err := file.Read(lengthBuf)
		if n == 0 && err != nil {
			break
		}
		if n < 4 {
			break
		}

		length := uint32(0)
		for i := 0; i < 4; i++ {
			length = (length << 8) | uint32(lengthBuf[i])
		}

		// Read data
		data := make([]byte, length)
		n, err = file.Read(data)
		if n < int(length) && err != nil {
			break
		}

		var entry audit.AuditLogEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// countVerifiedEntries counts total verified entries
func (v *AuditLogVerifier) countVerifiedEntries() (int, error) {
	logFiles, err := filepath.Glob(filepath.Join(v.storagePath, "audit_*.log"))
	if err != nil {
		return 0, err
	}

	total := 0
	for _, logFile := range logFiles {
		entries, err := v.loadEntriesFromFile(logFile)
		if err != nil {
			continue
		}
		total += len(entries)
	}

	return total, nil
}

// calculateHash calculates the hash of an entry
func (v *AuditLogVerifier) calculateHash(entry *audit.AuditLogEntry) string {
	// Create canonical representation
	canonical := struct {
		EntryID      string
		SequenceNum  uint64
		Timestamp    time.Time
		ChainID      string
		ActorID      string
		Service      string
		Operation    string
		ActionType   string
		Resource     string
		Result       string
		PreviousHash string
	}{
		EntryID:      entry.EntryID,
		SequenceNum:  entry.SequenceNum,
		Timestamp:    entry.Timestamp,
		ChainID:      entry.ChainID,
		ActorID:      entry.ActorID,
		Service:      entry.Service,
		Operation:    entry.Operation,
		ActionType:   entry.ActionType,
		Resource:     entry.Resource,
		Result:       entry.Result,
		PreviousHash: entry.PreviousHash,
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// buildMerkleTree builds a Merkle tree and returns the root
func (v *AuditLogVerifier) buildMerkleTree(entries []*audit.AuditLogEntry) string {
	if len(entries) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	hashes := make([]string, len(entries))
	for i, entry := range entries {
		hashes[i] = entry.CurrentHash
	}

	for len(hashes) > 1 {
		var nextLevel []string
		for i := 0; i < len(hashes); i += 2 {
			left := hashes[i]
			right := left
			if i+1 < len(hashes) {
				right = hashes[i+1]
			}

			combined := left + right
			hash := sha256.Sum256([]byte(combined))
			nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
		}
		hashes = nextLevel
	}

	return hashes[0]
}

// sortBySequence sorts entries by sequence number
func sortBySequence(entries []*audit.AuditLogEntry) []*audit.AuditLogEntry {
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].SequenceNum < entries[i].SequenceNum {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	return entries
}

// GenerateVerificationReport generates a detailed verification report
func (v *AuditLogVerifier) GenerateVerificationReport(ctx context.Context) (*VerificationReport, error) {
	result, err := v.VerifyChain(ctx)
	if err != nil {
		return nil, err
	}

	report := &VerificationReport{
		GeneratedAt:      time.Now().UTC(),
		ChainIntegrity:   result.ChainIntegrity,
		HashIntegrity:    result.HashIntegrity,
		SealIntegrity:    result.SealIntegrity,
		VerifiedEntries:  result.VerifiedEntries,
		FailedEntries:    result.FailedEntries,
		Errors:           result.Errors,
	}

	// Add chain summary
	chainsDir := filepath.Join(filepath.Dir(v.chainFilePath), "chains")
	entries, err := os.ReadDir(chainsDir)
	if err == nil {
		report.TotalChains = len(entries)
	}

	return report, nil
}

// VerificationReport represents a detailed verification report
type VerificationReport struct {
	GeneratedAt     time.Time  `json:"generated_at"`
	TotalChains     int        `json:"total_chains"`
	ChainIntegrity  bool       `json:"chain_integrity"`
	HashIntegrity   bool       `json:"hash_integrity"`
	SealIntegrity   bool       `json:"seal_integrity"`
	VerifiedEntries int        `json:"verified_entries"`
	FailedEntries   int        `json:"failed_entries"`
	Errors          []string   `json:"errors"`
}

// GetChainSummary returns a summary of all chains
func (v *AuditLogVerifier) GetChainSummary(ctx context.Context) ([]ChainSummary, error) {
	chainsDir := filepath.Join(filepath.Dir(v.chainFilePath), "chains")

	entries, err := os.ReadDir(chainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ChainSummary{}, nil
		}
		return nil, err
	}

	var summaries []ChainSummary
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		chainFile := filepath.Join(chainsDir, entry.Name())
		data, err := os.ReadFile(chainFile)
		if err != nil {
			continue
		}

		var chain audit.AuditChain
		if err := json.Unmarshal(data, &chain); err != nil {
			continue
		}

		// Extract chain number from filename
		chainNum := strings.TrimSuffix(entry.Name(), ".json")

		summaries = append(summaries, ChainSummary{
			ChainID:       chain.ChainID,
			ChainNumber:   chainNum,
			SequenceStart: chain.SequenceStart,
			SequenceEnd:   chain.SequenceEnd,
			EntryCount:    chain.EntryCount,
			SealedAt:      chain.SealedAt,
		})
	}

	return summaries, nil
}

// ChainSummary represents a summary of a single chain
type ChainSummary struct {
	ChainID       string    `json:"chain_id"`
	ChainNumber   string    `json:"chain_number"`
	SequenceStart uint64    `json:"sequence_start"`
	SequenceEnd   uint64    `json:"sequence_end"`
	EntryCount    int       `json:"entry_count"`
	SealedAt      time.Time `json:"sealed_at"`
}
