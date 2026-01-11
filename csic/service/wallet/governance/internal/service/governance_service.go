package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/config"
	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/csic/wallet-governance/internal/repository"
	"github.com/google/uuid"
)

// GovernanceService handles multi-signature governance operations
type GovernanceService struct {
	walletRepo  repository.WalletRepository
	signatureSvc *SignatureService
	hsmService  *HSMService
	auditRepo   repository.AuditRepository
	config      config.GovernanceConfig

	stopChan chan struct{}
}

// NewGovernanceService creates a new governance service
func NewGovernanceService(
	walletRepo repository.WalletRepository,
	signatureSvc *SignatureService,
	hsmService *HSMService,
	auditRepo repository.AuditRepository,
) *GovernanceService {
	return &GovernanceService{
		walletRepo:   walletRepo,
		signatureSvc: signatureSvc,
		hsmService:   hsmService,
		auditRepo:    auditRepo,
		stopChan:     make(chan struct{}),
	}
}

// Configure configures the governance service
func (s *GovernanceService) Configure(cfg config.GovernanceConfig) {
	s.config = cfg
}

// ProposeTransaction proposes a new transaction for multi-sig approval
func (s *GovernanceService) ProposeTransaction(ctx context.Context, proposal *models.TransactionProposal, proposerID uuid.UUID, proposerName string) (*models.TransactionProposal, error) {
	// Verify wallet exists and is multi-sig
	wallet, err := s.walletRepo.GetByID(ctx, proposal.WalletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	if wallet.Type != models.WalletTypeMultiSig {
		return nil, fmt.Errorf("wallet is not a multi-signature wallet")
	}

	// Set proposal fields
	proposal.TransactionID = fmt.Sprintf("TX-%s", uuid.New().String()[:8])
	proposal.Status = models.TransactionStatusPending
	proposal.ProposerID = proposerID
	proposal.ProposerName = proposerName
	proposal.ApproversRequired = wallet.Threshold
	proposal.ExpiresAt = time.Now().Add(time.Duration(s.config.TransactionExpiryHours) * time.Hour)

	// This would be stored in a transaction proposal repository
	// For now, we return the proposal

	s.logAudit(ctx, "TRANSACTION_PROPOSAL", proposal.ID, "CREATE", proposerID, proposerName, nil, proposal, true, "")

	return proposal, nil
}

// ApproveTransaction approves a proposed transaction
func (s *GovernanceService) ApproveTransaction(ctx context.Context, proposalID, signerID uuid.UUID, signerName, signature string) (*models.TransactionProposal, error) {
	// Get proposal (would come from repository)
	proposal := &models.TransactionProposal{
		ID:             proposalID,
		Status:         models.TransactionStatusApproved,
		ApprovalsCount: 1,
	}

	// Verify signature
	valid, err := s.hsmService.Verify(ctx, proposal.MessageHash, signature, "")
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid signature")
	}

	s.logAudit(ctx, "TRANSACTION_PROPOSAL", proposalID, "APPROVE", signerID, signerName, nil, map[string]interface{}{
		"signer": signerName,
	}, true, "")

	return proposal, nil
}

// RejectTransaction rejects a proposed transaction
func (s *GovernanceService) RejectTransaction(ctx context.Context, proposalID, signerID uuid.UUID, signerName, reason string) error {
	s.logAudit(ctx, "TRANSACTION_PROPOSAL", proposalID, "REJECT", signerID, signerName, nil, map[string]interface{}{
		"signer": signerName,
		"reason": reason,
	}, true, "")

	return nil
}

// GetPendingTransactions retrieves pending transactions for a wallet
func (s *GovernanceService) GetPendingTransactions(ctx context.Context, walletID uuid.UUID) ([]*models.TransactionProposal, error) {
	// Would query from transaction proposal repository
	return []*models.TransactionProposal{}, nil
}

// StartTransactionExpiryChecker starts the background task to check for expired transactions
func (s *GovernanceService) StartTransactionExpiryChecker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check for expired transactions
			// This would update transaction status to EXPIRED
		case <-s.stopChan:
			return
		}
	}
}

// StopTransactionExpiryChecker stops the transaction expiry checker
func (s *GovernanceService) StopTransactionExpiryChecker() {
	close(s.stopChan)
}

// AddSigner adds a signer to a multi-sig wallet
func (s *GovernanceService) AddSigner(ctx context.Context, walletID uuid.UUID, signer *models.WalletSigner, actorID uuid.UUID, actorName string) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	if wallet.Type != models.WalletTypeMultiSig {
		return fmt.Errorf("wallet is not a multi-signature wallet")
	}

	signer.WalletID = walletID
	signer.Status = models.SignerStatusActive

	// Update wallet signer count
	wallet.SignersTotal++

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	s.logAudit(ctx, "WALLET_SIGNER", signer.ID, "ADD", actorID, actorName, nil, signer, true, "")

	return nil
}

// RemoveSigner removes a signer from a multi-sig wallet
func (s *GovernanceService) RemoveSigner(ctx context.Context, walletID, signerID uuid.UUID, actorID uuid.UUID, actorName string) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	// Update wallet signer count
	wallet.SignersTotal--
	s.walletRepo.Update(ctx, wallet)

	s.logAudit(ctx, "WALLET_SIGNER", signerID, "REMOVE", actorID, actorName, nil, map[string]interface{}{
		"wallet_id": walletID,
	}, true, "")

	return nil
}

// UpdateSignerThreshold updates the threshold for a multi-sig wallet
func (s *GovernanceService) UpdateSignerThreshold(ctx context.Context, walletID uuid.UUID, threshold int, actorID uuid.UUID, actorName string) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	if threshold > wallet.SignersTotal {
		return fmt.Errorf("threshold cannot exceed number of signers")
	}

	wallet.Threshold = threshold
	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	s.logAudit(ctx, "WALLET", walletID, "UPDATE_THRESHOLD", actorID, actorName, map[string]interface{}{
		"old_threshold": wallet.Threshold,
	}, map[string]interface{}{
		"new_threshold": threshold,
	}, true, "")

	return nil
}

// GetWalletSigners retrieves signers for a wallet
func (s *GovernanceService) GetWalletSigners(ctx context.Context, walletID uuid.UUID) ([]*models.WalletSigner, error) {
	// Would query from wallet signer repository
	return []*models.WalletSigner{}, nil
}

// logAudit logs an audit event
func (s *GovernanceService) logAudit(ctx context.Context, entityType string, entityID uuid.UUID, action string, actorID uuid.UUID, actorName string, oldValue, newValue interface{}, success bool, errorMsg string) {
	log := &models.WalletAuditLog{
		EntityType:   entityType,
		EntityID:     entityID,
		Action:       action,
		ActorID:      actorID,
		ActorName:    actorName,
		ActorType:    "USER",
		OldValue:     toJSONMap(oldValue),
		NewValue:     toJSONMap(newValue),
		Success:      success,
		ErrorMessage: errorMsg,
	}

	s.auditRepo.Create(ctx, log)
}

func toJSONMap(v interface{}) models.JSONMap {
	if v == nil {
		return nil
	}
	data, _ := json.Marshal(v)
	var result models.JSONMap
	json.Unmarshal(data, &result)
	return result
}
