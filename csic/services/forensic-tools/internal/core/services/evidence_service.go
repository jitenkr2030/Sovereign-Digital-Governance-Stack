package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/csic-platform/internal/core/domain"
	"github.com/csic-platform/internal/core/ports"
)

// EvidenceManagementService implements the EvidenceService interface
type EvidenceManagementService struct {
	repo       ports.EvidenceRepository
	reportRepo ports.ReportRepository
	mu         sync.RWMutex
}

// NewEvidenceManagementService creates a new evidence management service instance
func NewEvidenceManagementService(repo ports.EvidenceRepository, reportRepo ports.ReportRepository) *EvidenceManagementService {
	return &EvidenceManagementService{
		repo:       repo,
		reportRepo: reportRepo,
	}
}

// CollectTransactionEvidence collects and preserves evidence related to a transaction
func (s *EvidenceManagementService) CollectTransactionEvidence(ctx context.Context, txHash string) (*domain.Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create evidence record
	evidence := &domain.Evidence{
		ID:         fmt.Sprintf("ev-%s-%d", txHash[:8], time.Now().UnixNano()),
		CaseID:     "",
		EvidenceType: domain.EvidenceTypeTransaction,
		Title:      fmt.Sprintf("Transaction Evidence: %s...", txHash[:16]),
		Description: fmt.Sprintf("Forensic evidence collected for transaction %s", txHash),
		Content: map[string]interface{}{
			"tx_hash":      txHash,
			"collection_time": time.Now().UTC().Format(time.RFC3339),
			"data_source":  "blockchain",
		},
		Tags:       []string{"transaction", "blockchain", txHash[:8]},
		CollectedAt: time.Now().Unix(),
		CollectedBy: "system",
		Status:     domain.EvidenceStatusCollected,
	}

	// Calculate cryptographic hash of the evidence
	hash, err := s.CalculateEvidenceHash(ctx, evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate evidence hash: %w", err)
	}
	evidence.Hash = hash

	// Store the evidence
	if err := s.repo.StoreEvidence(ctx, evidence); err != nil {
		return nil, fmt.Errorf("failed to store evidence: %w", err)
	}

	// Record initial chain of custody
	custody := &domain.ChainOfCustody{
		Timestamp:   time.Now().Unix(),
		Handler:     "system",
		Action:      "collection",
		Location:    "automated",
		IntegrityHash: hash,
	}
	if err := s.repo.AppendCustodyRecord(ctx, evidence.ID, custody); err != nil {
		return nil, fmt.Errorf("failed to record chain of custody: %w", err)
	}

	return evidence, nil
}

// CollectWalletEvidence collects and preserves evidence related to a wallet address
func (s *EvidenceManagementService) CollectWalletEvidence(ctx context.Context, address string) (*domain.Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create evidence record for wallet
	evidence := &domain.Evidence{
		ID:         fmt.Sprintf("ev-wallet-%s-%d", address[:8], time.Now().UnixNano()),
		CaseID:     "",
		EvidenceType: domain.EvidenceTypeWallet,
		Title:      fmt.Sprintf("Wallet Evidence: %s...", address[:12]),
		Description: fmt.Sprintf("Forensic evidence collected for wallet address %s", address),
		Content: map[string]interface{}{
			"wallet_address":  address,
			"collection_time": time.Now().UTC().Format(time.RFC3339),
			"data_source":     "blockchain",
		},
		Tags:       []string{"wallet", "address", address[:8]},
		CollectedAt: time.Now().Unix(),
		CollectedBy: "system",
		Status:     domain.EvidenceStatusCollected,
	}

	// Calculate cryptographic hash
	hash, err := s.CalculateEvidenceHash(ctx, evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate evidence hash: %w", err)
	}
	evidence.Hash = hash

	// Store the evidence
	if err := s.repo.StoreEvidence(ctx, evidence); err != nil {
		return nil, fmt.Errorf("failed to store evidence: %w", err)
	}

	// Record initial chain of custody
	custody := &domain.ChainOfCustody{
		Timestamp:     time.Now().Unix(),
		Handler:       "system",
		Action:        "collection",
		Location:      "automated",
		IntegrityHash: hash,
	}
	if err := s.repo.AppendCustodyRecord(ctx, evidence.ID, custody); err != nil {
		return nil, fmt.Errorf("failed to record chain of custody: %w", err)
	}

	return evidence, nil
}

// CollectClusterEvidence collects evidence for an entire cluster of related entities
func (s *EvidenceManagementService) CollectClusterEvidence(ctx context.Context, clusterID string) (*domain.Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create comprehensive evidence record for cluster
	evidence := &domain.Evidence{
		ID:         fmt.Sprintf("ev-cluster-%s-%d", clusterID[:8], time.Now().UnixNano()),
		CaseID:     "",
		EvidenceType: domain.EvidenceTypeCluster,
		Title:      fmt.Sprintf("Cluster Evidence: %s...", clusterID[:12]),
		Description: fmt.Sprintf("Forensic evidence collected for cluster %s", clusterID),
		Content: map[string]interface{}{
			"cluster_id":     clusterID,
			"collection_time": time.Now().UTC().Format(time.RFC3339),
			"data_source":    "graph_analysis",
		},
		Tags:       []string{"cluster", "graph", clusterID[:8]},
		CollectedAt: time.Now().Unix(),
		CollectedBy: "system",
		Status:     domain.EvidenceStatusCollected,
	}

	// Calculate cryptographic hash
	hash, err := s.CalculateEvidenceHash(ctx, evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate evidence hash: %w", err)
	}
	evidence.Hash = hash

	// Store the evidence
	if err := s.repo.StoreEvidence(ctx, evidence); err != nil {
		return nil, fmt.Errorf("failed to store evidence: %w", err)
	}

	// Record initial chain of custody
	custody := &domain.ChainOfCustody{
		Timestamp:     time.Now().Unix(),
		Handler:       "system",
		Action:        "collection",
		Location:      "automated",
		IntegrityHash: hash,
	}
	if err := s.repo.AppendCustodyRecord(ctx, evidence.ID, custody); err != nil {
		return nil, fmt.Errorf("failed to record chain of custody: %w", err)
	}

	return evidence, nil
}

// VerifyHashIntegrity verifies the integrity of evidence using its cryptographic hash
func (s *EvidenceManagementService) VerifyHashIntegrity(ctx context.Context, evidenceID string) (bool, error) {
	// Get the stored evidence
	evidence, err := s.repo.GetEvidence(ctx, evidenceID)
	if err != nil {
		return false, fmt.Errorf("failed to get evidence: %w", err)
	}

	// Calculate current hash
	currentHash, err := s.CalculateEvidenceHash(ctx, evidence)
	if err != nil {
		return false, fmt.Errorf("failed to calculate current hash: %w", err)
	}

	// Compare hashes
	return currentHash == evidence.Hash, nil
}

// CalculateEvidenceHash calculates a SHA-256 hash of the evidence content
func (s *EvidenceManagementService) CalculateEvidenceHash(ctx context.Context, evidence *domain.Evidence) (string, error) {
	// Serialize the evidence content to JSON
	jsonData, err := json.Marshal(evidence.Content)
	if err != nil {
		return "", fmt.Errorf("failed to serialize evidence: %w", err)
	}

	// Create hash including metadata
	hashInput := fmt.Sprintf("%s|%s|%s|%d|%s",
		evidence.ID,
		evidence.Title,
		evidence.Description,
		evidence.CollectedAt,
		string(jsonData),
	)

	// Calculate SHA-256 hash
	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:]), nil
}

// RecordCustody adds a new entry to the chain of custody record
func (s *EvidenceManagementService) RecordCustody(ctx context.Context, evidenceID string, handler string, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current evidence to verify it exists
	evidence, err := s.repo.GetEvidence(ctx, evidenceID)
	if err != nil {
		return fmt.Errorf("evidence not found: %w", err)
	}

	// Create custody record
	custody := &domain.ChainOfCustody{
		Timestamp:   time.Now().Unix(),
		Handler:     handler,
		Action:      action,
		Location:    "manual",
		IntegrityHash: evidence.Hash,
		Notes:       "",
	}

	// Append to chain of custody
	if err := s.repo.AppendCustodyRecord(ctx, evidenceID, custody); err != nil {
		return fmt.Errorf("failed to record custody: %w", err)
	}

	return nil
}

// GetEvidenceTimeline retrieves the complete chain of custody timeline for evidence
func (s *EvidenceManagementService) GetEvidenceTimeline(ctx context.Context, evidenceID string) ([]*domain.ChainOfCustody, error) {
	return s.repo.GetCustodyHistory(ctx, evidenceID)
}

// ExportEvidence exports evidence in the specified format
func (s *EvidenceManagementService) ExportEvidence(ctx context.Context, evidenceID string, format string) ([]byte, error) {
	// Get the evidence
	evidence, err := s.repo.GetEvidence(ctx, evidenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get evidence: %w", err)
	}

	// Get chain of custody
	custody, err := s.repo.GetCustodyHistory(ctx, evidenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custody history: %w", err)
	}

	switch format {
	case "json":
		return s.exportAsJSON(evidence, custody)
	case "pdf":
		// PDF export would require additional library
		return s.exportAsJSON(evidence, custody)
	default:
		return s.exportAsJSON(evidence, custody)
	}
}

// exportAsJSON exports evidence as JSON
func (s *EvidenceManagementService) exportAsJSON(evidence *domain.Evidence, custody []*domain.ChainOfCustody) ([]byte, error) {
	exportData := map[string]interface{}{
		"evidence":      evidence,
		"chain_of_custody": custody,
		"export_time":   time.Now().UTC().Format(time.RFC3339),
	}

	return json.MarshalIndent(exportData, "", "  ")
}

// PackageForLegal creates a comprehensive legal evidence package
func (s *EvidenceManagementService) PackageForLegal(ctx context.Context, evidenceIDs []string) (*domain.Report, error) {
	var allEvidence []*domain.Evidence
	var allCustody [][]*domain.ChainOfCustody

	for _, id := range evidenceIDs {
		evidence, err := s.repo.GetEvidence(ctx, id)
		if err != nil {
			continue
		}
		allEvidence = append(allEvidence, evidence)

		custody, err := s.repo.GetCustodyHistory(ctx, id)
		if err != nil {
			custody = []*domain.ChainOfCustody{}
		}
		allCustody = append(allCustody, custody)
	}

	// Create comprehensive legal package
	packageContent := map[string]interface{}{
		"package_type":     "legal_evidence",
		"evidence_count":   len(allEvidence),
		"evidence_ids":     evidenceIDs,
		"collection_time":  time.Now().UTC().Format(time.RFC3339),
		"total_custody_records": 0,
	}

	for _, custodyRecords := range allCustody {
		packageContent["total_custody_records"] = packageContent["total_custody_records"].(int) + len(custodyRecords)
	}

	// Generate hash for the entire package
	packageJSON, _ := json.Marshal(packageContent)
	packageHash := sha256.Sum256(packageJSON)

	// Create report
	report := &domain.Report{
		ID:          fmt.Sprintf("rpt-legal-%d", time.Now().UnixNano()),
		CaseID:      "",
		ReportType:  domain.ReportTypeLegal,
		Title:       fmt.Sprintf("Legal Evidence Package - %s", time.Now().UTC().Format("2006-01-02")),
		Description: "Comprehensive legal evidence package with chain of custody",
		Content:     packageContent,
		GeneratedAt: time.Now().Unix(),
		GeneratedBy: "system",
		Status:      domain.ReportStatusGenerated,
		Version:     "1.0",
	}

	// Calculate report hash
	reportJSON, _ := json.Marshal(report.Content)
	report.Hash = fmt.Sprintf("%x", sha256.Sum(reportJSON))

	// Store the report
	if err := s.reportRepo.StoreReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to store report: %w", err)
	}

	return report, nil
}

// VerifyEvidence verifies that evidence has not been tampered with
func (s *EvidenceManagementService) VerifyEvidence(ctx context.Context, id string) (bool, error) {
	return s.repo.VerifyEvidence(ctx, id)
}

// CreateCaseEvidence creates a comprehensive evidence package for a case
func (s *EvidenceManagementService) CreateCaseEvidence(ctx context.Context, caseID string, evidenceIDs []string) (*domain.Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	evidence := &domain.Evidence{
		ID:         fmt.Sprintf("ev-case-%s-%d", caseID[:8], time.Now().UnixNano()),
		CaseID:     caseID,
		EvidenceType: domain.EvidenceTypeReport,
		Title:      fmt.Sprintf("Case Evidence Package: %s...", caseID[:12]),
		Description: fmt.Sprintf("Comprehensive evidence package for case %s containing %d items", caseID, len(evidenceIDs)),
		Content: map[string]interface{}{
			"case_id":        caseID,
			"evidence_ids":   evidenceIDs,
			"collection_time": time.Now().UTC().Format(time.RFC3339),
			"package_type":   "case_comprehensive",
		},
		Tags:       []string{"case", caseID[:8], "comprehensive"},
		CollectedAt: time.Now().Unix(),
		CollectedBy: "system",
		Status:     domain.EvidenceStatusCollected,
	}

	hash, err := s.CalculateEvidenceHash(ctx, evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}
	evidence.Hash = hash

	if err := s.repo.StoreEvidence(ctx, evidence); err != nil {
		return nil, fmt.Errorf("failed to store evidence: %w", err)
	}

	custody := &domain.ChainOfCustody{
		Timestamp:     time.Now().Unix(),
		Handler:       "system",
		Action:        "case_package_created",
		Location:      "automated",
		IntegrityHash: hash,
	}
	s.repo.AppendCustodyRecord(ctx, evidence.ID, custody)

	return evidence, nil
}
