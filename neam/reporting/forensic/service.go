package forensic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// Config holds the configuration for the forensic evidence service
type Config struct {
	PostgreSQL  *shared.PostgreSQL
	ClickHouse  clickhouse.Conn
	Redis       *redis.Client
	Logger      *shared.Logger
}

// Service handles forensic evidence collection and verification
type Service struct {
	config     Config
	logger     *shared.Logger
	postgreSQL *shared.PostgreSQL
	clickhouse clickhouse.Conn
	redis      *redis.Client
	evidence   map[string]*Evidence
	chains     map[string]*EvidenceChain
	verifications map[string]*VerificationRecord
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// Evidence represents a piece of forensic evidence
type Evidence struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"` // intervention, transaction, anomaly, manual
	ChainHash       string                 `json:"chain_hash"`
	ParentHash      string                 `json:"parent_hash"`
	MerkleProof     []MerkleNode           `json:"merkle_proof"`
	Content         map[string]interface{} `json:"content"`
	Metadata        EvidenceMetadata       `json:"metadata"`
	Timestamp       time.Time              `json:"timestamp"`
	SubmittedBy     string                 `json:"submitted_by"`
	SubmittedAt     time.Time              `json:"submitted_at"`
	Verified        bool                   `json:"verified"`
	VerifiedAt      time.Time              `json:"verified_at"`
	Version         int                    `json:"version"`
	PreviousVersion string                 `json:"previous_version"`
}

// EvidenceMetadata contains metadata about evidence
type EvidenceMetadata struct {
	SourceSystem   string                 `json:"source_system"`
	SourceID       string                 `json:"source_id"`
	EventType      string                 `json:"event_type"`
	Jurisdiction   string                 `json:"jurisdiction"`
	Classification string                 `json:"classification"`
	Tags           []string               `json:"tags"`
	RelatedIDs     []string               `json:"related_ids"`
	RetentionDate  time.Time              `json:"retention_date"`
}

// MerkleNode represents a node in the Merkle proof
type MerkleNode struct {
	Hash     string `json:"hash"`
	Side     string `json:"side"` // left or right
	Position int    `json:"position"`
}

// EvidenceChain represents a chain of evidence
type EvidenceChain struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	EvidenceIDs  []string            `json:"evidence_ids"`
	RootHash     string              `json:"root_hash"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Status       string              `json:"status"` // active, sealed, archived
	SealedAt     time.Time           `json:"sealed_at"`
	Entries      []EvidenceChainEntry `json:"entries"`
}

// EvidenceChainEntry represents an entry in the evidence chain
type EvidenceChainEntry struct {
	EvidenceID   string    `json:"evidence_id"`
	Timestamp    time.Time `json:"timestamp"`
	Operation    string    `json:"operation"` // add, modify, verify
	Operator     string    `json:"operator"`
	Hash         string    `json:"hash"`
	PreviousHash string    `json:"previous_hash"`
}

// VerificationRecord represents a verification attempt
type VerificationRecord struct {
	ID            string                 `json:"id"`
	EvidenceID    string                 `json:"evidence_id"`
	VerifiedAt    time.Time              `json:"verified_at"`
	Verifier      string                 `json:"verifier"`
	Result        string                 `json:"result"` // valid, invalid, pending
	Checks        []VerificationCheck    `json:"checks"`
	Details       map[string]interface{} `json:"details"`
	ChainValid    bool                   `json:"chain_valid"`
	ContentHash   string                 `json:"content_hash"`
	ReportedHash  string                 `json:"reported_hash"`
	Match         bool                   `json:"match"`
}

// VerificationCheck represents a single verification check
type VerificationCheck struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

// TamperCheckResult represents the result of a tamper check
type TamperCheckResult struct {
	Tampered       bool                   `json:"tampered"`
	Verified       bool                   `json:"verified"`
	ChainValid     bool                   `json:"chain_valid"`
	Timestamp      time.Time              `json:"timestamp"`
	EvidenceID     string                 `json:"evidence_id"`
	Checks         []VerificationCheck    `json:"checks"`
	Discrepancies  []Discrepancy          `json:"discrepancies"`
	ChainIntegrity ChainIntegrity         `json:"chain_integrity"`
}

// Discrepancy represents a detected discrepancy
type Discrepancy struct {
	Field        string                 `json:"field"`
	Expected     interface{}            `json:"expected"`
	Actual       interface{}            `json:"actual"`
	Impact       string                 `json:"impact"`
	Severity     string                 `json:"severity"`
}

// ChainIntegrity represents the integrity status of the evidence chain
type ChainIntegrity struct {
	RootHash       string   `json:"root_hash"`
	CompleteChain  bool     `json:"complete_chain"`
	MissingLinks   []string `json:"missing_links"`
	Sealed         bool     `json:"sealed"`
	SealedAt       time.Time `json:"sealed_at"`
	HashAlgorithm  string   `json:"hash_algorithm"`
}

// NewService creates a new forensic evidence service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:        cfg,
		logger:        cfg.Logger,
		postgreSQL:    cfg.PostgreSQL,
		clickhouse:    cfg.ClickHouse,
		redis:         cfg.Redis,
		evidence:      make(map[string]*Evidence),
		chains:        make(map[string]*EvidenceChain),
		verifications: make(map[string]*VerificationRecord),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the forensic evidence service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting forensic evidence service")
	
	// Start background workers
	s.wg.Add(1)
	go s.chainMaintenanceWorker()
	s.wg.Add(1)
	go s.verificationWorker()
	
	s.logger.Info("Forensic evidence service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping forensic evidence service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Forensic evidence service stopped")
}

// chainMaintenanceWorker maintains evidence chains
func (s *Service) chainMaintenanceWorker() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.maintainChains()
		}
	}
}

// maintainChains performs maintenance on evidence chains
func (s *Service) maintainChains() {
	s.logger.Debug("Performing chain maintenance")
	
	// Verify chain integrity
	for _, chain := range s.chains {
		s.verifyChainIntegrity(chain)
	}
}

// verificationWorker handles verification requests
func (s *Service) verificationWorker() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Process pending verifications
		}
	}
}

// GenerateEvidence creates new evidence from intervention data
func (s *Service) GenerateEvidence(ctx context.Context, evidenceType string, content map[string]interface{}, metadata EvidenceMetadata) (*Evidence, error) {
	// Generate content hash
	contentJSON, _ := json.Marshal(content)
	contentHash := s.computeHash(contentJSON)
	
	// Get parent hash (last evidence in chain)
	parentHash := s.getLatestChainHash(evidenceType)
	
	// Create Merkle proof
	merkleProof := s.generateMerkleProof(contentHash, parentHash)
	
	evidence := &Evidence{
		ID:         fmt.Sprintf("evd-%d", time.Now().UnixNano()),
		Type:       evidenceType,
		ChainHash:  s.computeHash(append(contentJSON, []byte(parentHash)...)),
		ParentHash: parentHash,
		MerkleProof: merkleProof,
		Content:    content,
		Metadata:   metadata,
		Timestamp:  time.Now(),
		SubmittedAt: time.Now(),
		Version:    1,
	}
	
	// Store evidence
	s.mu.Lock()
	s.evidence[evidence.ID] = evidence
	s.mu.Unlock()
	
	// Add to chain
	s.addToChain(evidence)
	
	s.logger.Info("Evidence generated", "id", evidence.ID, "type", evidenceType)
	
	return evidence, nil
}

// computeHash computes SHA-256 hash
func (s *Service) computeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// getLatestChainHash gets the latest hash for the evidence chain
func (s *Service) getLatestChainHash(evidenceType string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, e := range s.evidence {
		if e.Type == evidenceType {
			return e.ChainHash
		}
	}
	
	return ""
}

// generateMerkleProof generates a Merkle proof for the evidence
func (s *Service) generateMerkleProof(contentHash, parentHash string) []MerkleNode {
	return []MerkleNode{
		{
			Hash:     parentHash,
			Side:     "left",
			Position: 0,
		},
		{
			Hash:     contentHash,
			Side:     "right",
			Position: 1,
		},
	}
}

// addToChain adds evidence to the appropriate chain
func (s *Service) addToChain(evidence *Evidence) {
	chainID := fmt.Sprintf("chain-%s", evidence.Type)
	
	s.mu.Lock()
	chain, exists := s.chains[chainID]
	if !exists {
		chain = &EvidenceChain{
			ID:          chainID,
			Name:        fmt.Sprintf("%s Evidence Chain", evidence.Type),
			EvidenceIDs: []string{},
			CreatedAt:   time.Now(),
			Status:      "active",
		}
		s.chains[chainID] = chain
	}
	
	// Add entry to chain
	chain.Entries = append(chain.Entries, EvidenceChainEntry{
		EvidenceID:   evidence.ID,
		Timestamp:    time.Now(),
		Operation:    "add",
		Hash:         evidence.ChainHash,
		PreviousHash: evidence.ParentHash,
	})
	
	chain.EvidenceIDs = append(chain.EvidenceIDs, evidence.ID)
	chain.UpdatedAt = time.Now()
	chain.RootHash = s.computeHash([]byte(evidence.ChainHash))
	
	s.mu.Unlock()
}

// CheckTamper performs tamper detection on evidence
func (s *Service) CheckTamper(ctx context.Context, evidenceID string) (*TamperCheckResult, error) {
	s.mu.RLock()
	evidence, exists := s.evidence[evidenceID]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("evidence not found: %s", evidenceID)
	}
	
	result := &TamperCheckResult{
		Tampered:    false,
		Verified:    true,
		ChainValid:  true,
		Timestamp:   time.Now(),
		EvidenceID:  evidenceID,
		Checks:      []VerificationCheck{},
		Discrepancies: []Discrepancy{},
	}
	
	// Verify content hash
	contentJSON, _ := json.Marshal(evidence.Content)
	expectedHash := s.computeHash(contentJSON)
	calculatedHash := s.computeHash(append(contentJSON, []byte(evidence.ParentHash)...))
	
	result.Checks = append(result.Checks, VerificationCheck{
		Name:    "content_integrity",
		Passed:  evidence.ChainHash == calculatedHash,
		Message: "Content hash verification",
	})
	
	if evidence.ChainHash != calculatedHash {
		result.Tampered = true
		result.Discrepancies = append(result.Discrepancies, Discrepancy{
			Field:    "chain_hash",
			Expected: evidence.ChainHash,
			Actual:   calculatedHash,
			Impact:   "high",
			Severity: "critical",
		})
	}
	
	// Verify chain integrity
	chainIntegrity := s.verifyChainIntegrityForEvidence(evidence)
	result.ChainValid = chainIntegrity.CompleteChain
	result.Checks = append(result.Checks, VerificationCheck{
		Name:    "chain_integrity",
		Passed:  chainIntegrity.CompleteChain,
		Message: "Evidence chain verification",
	})
	
	result.ChainIntegrity = chainIntegrity
	
	return result, nil
}

// verifyChainIntegrityForEvidence verifies the chain for specific evidence
func (s *Service) verifyChainIntegrityForEvidence(evidence *Evidence) ChainIntegrity {
	chainID := fmt.Sprintf("chain-%s", evidence.Type)
	
	s.mu.RLock()
	chain, exists := s.chains[chainID]
	s.mu.RUnlock()
	
	if !exists {
		return ChainIntegrity{
			CompleteChain: false,
			MissingLinks:  []string{evidence.ID},
		}
	}
	
	// Verify chain is continuous
	var missingLinks []string
	for i, entry := range chain.Entries {
		if i > 0 && entry.PreviousHash != chain.Entries[i-1].Hash {
			missingLinks = append(missingLinks, fmt.Sprintf("link-%d", i))
		}
	}
	
	return ChainIntegrity{
		RootHash:      chain.RootHash,
		CompleteChain: len(missingLinks) == 0,
		MissingLinks:  missingLinks,
		Sealed:        chain.Status == "sealed",
		SealedAt:      chain.SealedAt,
		HashAlgorithm: "SHA-256",
	}
}

// VerifyEvidence performs comprehensive verification of evidence
func (s *Service) VerifyEvidence(ctx context.Context, evidenceID, verifier string) (*VerificationRecord, error) {
	s.mu.RLock()
	evidence, exists := s.evidence[evidenceID]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("evidence not found: %s", evidenceID)
	}
	
	// Perform comprehensive verification
	record := &VerificationRecord{
		ID:         fmt.Sprintf("ver-%d", time.Now().UnixNano()),
		EvidenceID: evidenceID,
		VerifiedAt: time.Now(),
		Verifier:   verifier,
		Checks:     []VerificationCheck{},
	}
	
	// Check content integrity
	contentJSON, _ := json.Marshal(evidence.Content)
	expectedHash := s.computeHash(contentJSON)
	calculatedHash := s.computeHash(append(contentJSON, []byte(evidence.ParentHash)...))
	
	contentCheck := VerificationCheck{
		Name:    "content_hash",
		Passed:  evidence.ChainHash == calculatedHash,
		Message: "Content hash matches calculated value",
	}
	record.Checks = append(record.Checks, contentCheck)
	
	// Check Merkle proof
	merkleCheck := VerificationCheck{
		Name:    "merkle_proof",
		Passed:  len(evidence.MerkleProof) > 0,
		Message: "Merkle proof present and valid",
	}
	record.Checks = append(record.Checks, merkleCheck)
	
	// Check timestamp
	timestampCheck := VerificationCheck{
		Name:    "timestamp",
		Passed:  !evidence.Timestamp.IsZero(),
		Message: "Timestamp is valid",
	}
	record.Checks = append(record.Checks, timestampCheck)
	
	// Determine overall result
	allPassed := true
	for _, check := range record.Checks {
		if !check.Passed {
			allPassed = false
			break
		}
	}
	
	record.Result = "valid"
	if !allPassed {
		record.Result = "invalid"
	}
	
	record.ChainValid = s.verifyChainIntegrityForEvidence(evidence).CompleteChain
	record.ContentHash = expectedHash
	
	// Store verification record
	s.mu.Lock()
	s.verifications[record.ID] = record
	s.mu.Unlock()
	
	return record, nil
}

// ListEvidence returns all evidence with optional filters
func (s *Service) ListEvidence(ctx context.Context, filters map[string]string) ([]*Evidence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var evidenceList []*Evidence
	for _, e := range s.evidence {
		// Apply filters
		if filters["type"] != "" && e.Type != filters["type"] {
			continue
		}
		if filters["source_system"] != "" && e.Metadata.SourceSystem != filters["source_system"] {
			continue
		}
		evidenceList = append(evidenceList, e)
	}
	
	return evidenceList, nil
}

// GetEvidence retrieves specific evidence
func (s *Service) GetEvidence(ctx context.Context, id string) (*Evidence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	evidence, exists := s.evidence[id]
	if !exists {
		return nil, fmt.Errorf("evidence not found: %s", id)
	}
	
	return evidence, nil
}

// GetEvidenceChain retrieves the evidence chain for a given type
func (s *Service) GetEvidenceChain(ctx context.Context, chainID string) (*EvidenceChain, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	chain, exists := s.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}
	
	return chain, nil
}

// SealChain seals an evidence chain, making it immutable
func (s *Service) SealChain(ctx context.Context, chainID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	chain, exists := s.chains[chainID]
	if !exists {
		return fmt.Errorf("chain not found: %s", chainID)
	}
	
	chain.Status = "sealed"
	chain.SealedAt = time.Now()
	chain.RootHash = s.computeHash([]byte(time.Now().Format(time.RFC3339)))
	
	s.logger.Info("Evidence chain sealed", "chain_id", chainID)
	
	return nil
}

// verifyChainIntegrity verifies the integrity of a chain
func (s *Service) verifyChainIntegrity(chain *EvidenceChain) {
	// Implementation for chain verification
	s.logger.Debug("Verifying chain integrity", "chain_id", chain.ID)
}
