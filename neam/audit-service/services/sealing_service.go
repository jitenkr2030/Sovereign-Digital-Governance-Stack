package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"neam-platform/audit-service/crypto"
	"neam-platform/audit-service/models"
	"neam-platform/audit-service/repository"
)

// SealingService handles cryptographic sealing of audit records
type SealingService interface {
	// SealEvents seals a batch of events
	SealEvents(ctx context.Context, eventIDs []string) (*models.SealRecord, error)

	// SealMerkleTree seals the current Merkle tree
	SealMerkleTree(ctx context.Context) error

	// VerifySeal verifies a seal record
	VerifySeal(ctx context.Context, seal *models.SealRecord) (bool, error)

	// CreateWitness creates an external witness for a seal
	CreateWitness(ctx context.Context, sealID string, witnessType string) (*models.WitnessRecord, error)
}

// sealingService implements SealingService
type sealingService struct {
	hashService   crypto.HashService
	signatureService crypto.SignatureService
	merkleService *MerkleService
	wormStorage   WORMStorage
	repo          repository.AuditRepository
}

// NewSealingService creates a new SealingService
func NewSealingService(
	hashService crypto.HashService,
	signatureService crypto.SignatureService,
	merkleService *MerkleService,
	wormStorage WORMStorage,
) SealingService {
	return &sealingService{
		hashService:     hashService,
		signatureService: signatureService,
		merkleService:   merkleService,
		wormStorage:     wormStorage,
	}
}

// SetRepository sets the repository for the sealing service
func (s *sealingService) SetRepository(repo repository.AuditRepository) {
	s.repo = repo
}

// SealEvents seals a batch of events
func (s *sealingService) SealEvents(ctx context.Context, eventIDs []string) (*models.SealRecord, error) {
	if len(eventIDs) == 0 {
		return nil, fmt.Errorf("no events to seal")
	}

	// Get events to seal
	events, err := s.repo.GetByIDs(ctx, eventIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	// Build Merkle tree from events
	var leaves [][]byte
	for _, event := range events {
		leaves = append(leaves, []byte(event.IntegrityHash))
	}

	// Create Merkle tree
	tree, err := s.merkleService.BuildTree(leaves)
	if err != nil {
		return nil, fmt.Errorf("failed to build Merkle tree: %w", err)
	}

	rootHash := hex.EncodeToString(tree.Root)

	// Get previous seal for chain continuity
	previousSeal, _ := s.repo.GetLatestSeal(ctx)
	previousRootHash := ""
	previousSealID := ""
	if previousSeal != nil {
		previousRootHash = previousSeal.RootHash
		previousSealID = previousSeal.ID
	}

	// Create seal record
	seal := &models.SealRecord{
		ID:              generateID(),
		Timestamp:       time.Now().UTC(),
		RootHash:        rootHash,
		FirstEventID:    events[0].ID,
		LastEventID:     events[len(events)-1].ID,
		EventCount:      int64(len(events)),
		SignerID:        "system",
		SignerType:      "SYSTEM",
		Algorithm:       "SHA256",
		PreviousSealID:  previousSealID,
		PreviousRootHash: previousRootHash,
		Metadata:        models.JSONMap{"tree_size": len(events)},
		CreatedAt:       time.Now().UTC(),
	}

	// Sign the seal
	signature, err := s.signatureService.Sign(ctx, []byte(rootHash))
	if err != nil {
		return nil, fmt.Errorf("failed to sign seal: %w", err)
	}
	seal.Signature = signature

	// Store seal
	if err := s.repo.CreateSeal(ctx, seal); err != nil {
		return nil, fmt.Errorf("failed to store seal: %w", err)
	}

	log.Printf("Sealed %d events with root hash %s", len(events), rootHash)
	return seal, nil
}

// SealMerkleTree seals the current Merkle tree
func (s *sealingService) SealMerkleTree(ctx context.Context) error {
	// Get unsealed events
	events, err := s.repo.GetUnsealedEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get unsealed events: %w", err)
	}

	if len(events) == 0 {
		log.Println("No unsealed events to seal")
		return nil
	}

	// Build Merkle tree
	var leaves [][]byte
	for _, event := range events {
		leaves = append(leaves, []byte(event.IntegrityHash))
	}

	tree, err := s.merkleService.BuildTree(leaves)
	if err != nil {
		return fmt.Errorf("failed to build Merkle tree: %w", err)
	}

	// Seal events
	eventIDs := make([]string, len(events))
	for i, event := range events {
		eventIDs[i] = event.ID
	}

	_, err = s.SealEvents(ctx, eventIDs)
	return err
}

// VerifySeal verifies a seal record
func (s *sealingService) VerifySeal(ctx context.Context, seal *models.SealRecord) (bool, error) {
	// Verify signature
	sigBytes, err := base64.StdEncoding.DecodeString(seal.Signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	// Verify using the signature service
	valid, err := s.signatureService.Verify(ctx, []byte(seal.RootHash), sigBytes)
	if err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}

	if !valid {
		return false, nil
	}

	// Verify chain continuity
	if seal.PreviousRootHash != "" {
		prevSeal, err := s.repo.GetSealByID(ctx, seal.PreviousSealID)
		if err != nil {
			return false, fmt.Errorf("failed to get previous seal: %w", err)
		}
		if prevSeal.RootHash != seal.PreviousRootHash {
			return false, fmt.Errorf("chain discontinuity")
		}
	}

	return true, nil
}

// CreateWitness creates an external witness for a seal
func (s *sealingService) CreateWitness(ctx context.Context, sealID string, witnessType string) (*models.WitnessRecord, error) {
	seal, err := s.repo.GetSealByID(ctx, sealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seal: %w", err)
	}

	// Create witness data
	witnessData := fmt.Sprintf("%s|%s|%d|%s",
		seal.ID,
		seal.RootHash,
		seal.EventCount,
		seal.Timestamp.Format(time.RFC3339),
	)

	// Sign witness data
	signature, err := s.signatureService.Sign(ctx, []byte(witnessData))
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	witness := &models.WitnessRecord{
		WitnessType: witnessType,
		WitnessID:   generateID(),
		Timestamp:   time.Now().UTC(),
		Signature:   signature,
	}

	// Update seal with witness
	seal.WitnessType = witnessType
	seal.WitnessID = witness.WitnessID
	seal.WitnessSignature = signature
	if err := s.repo.UpdateSeal(ctx, seal); err != nil {
		return nil, fmt.Errorf("failed to update seal with witness: %w", err)
	}

	return witness, nil
}

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
