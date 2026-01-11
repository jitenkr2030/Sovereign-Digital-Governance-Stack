package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/csic/wallet-governance/internal/repository"
	"github.com/google/uuid"
)

// FreezeService handles wallet freeze operations
type FreezeService struct {
	walletRepo   repository.WalletRepository
	freezeRepo   repository.WalletFreezeRepository
	signatureSvc *SignatureService
	auditRepo    repository.AuditRepository

	stopChan chan struct{}
}

// NewFreezeService creates a new freeze service
func NewFreezeService(
	walletRepo repository.WalletRepository,
	freezeRepo repository.WalletFreezeRepository,
	signatureSvc *SignatureService,
	auditRepo repository.AuditRepository,
) *FreezeService {
	return &FreezeService{
		walletRepo:   walletRepo,
		freezeRepo:   freezeRepo,
		signatureSvc: signatureSvc,
		auditRepo:    auditRepo,
		stopChan:     make(chan struct{}),
	}
}

// FreezeWallet freezes a wallet
func (s *FreezeService) FreezeWallet(ctx context.Context, freeze *models.WalletFreeze, actorID uuid.UUID, actorName string) error {
	// Check if wallet exists
	wallet, err := s.walletRepo.GetByID(ctx, freeze.WalletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	// Check if already frozen
	existing, err := s.freezeRepo.GetActiveByWallet(ctx, freeze.WalletID)
	if err != nil {
		return fmt.Errorf("failed to check existing freeze: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("wallet is already frozen")
	}

	freeze.Status = models.FreezeStatusActive
	freeze.WalletAddress = wallet.Address
	freeze.Blockchain = wallet.Blockchain

	if err := s.freezeRepo.Create(ctx, freeze); err != nil {
		return fmt.Errorf("failed to create freeze: %w", err)
	}

	// Update wallet status
	wallet.Status = models.WalletStatusFrozen
	s.walletRepo.Update(ctx, wallet)

	s.logAudit(ctx, "WALLET_FREEZE", freeze.ID, "CREATE", actorID, actorName, nil, freeze, true, "")

	return nil
}

// EmergencyFreeze performs an emergency freeze without approval
func (s *FreezeService) EmergencyFreeze(ctx context.Context, walletID uuid.UUID, reason models.FreezeReason, reasonDetails, legalOrderID string, actorID uuid.UUID, actorName string) error {
	freeze := &models.WalletFreeze{
		WalletID:      walletID,
		Reason:        reason,
		ReasonDetails: reasonDetails,
		LegalOrderID:  legalOrderID,
		FreezeLevel:   "FULL",
		IssuedBy:      actorID,
		IssuedByName:  actorName,
	}

	return s.FreezeWallet(ctx, freeze, actorID, actorName)
}

// UnfreezeWallet unfreezes a wallet
func (s *FreezeService) UnfreezeWallet(ctx context.Context, walletID uuid.UUID, reason string, actorID uuid.UUID, actorName string) error {
	freeze, err := s.freezeRepo.GetActiveByWallet(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get freeze: %w", err)
	}
	if freeze == nil {
		return fmt.Errorf("no active freeze found")
	}

	// Release freeze
	if err := s.freezeRepo.Release(ctx, freeze.ID, actorID, reason); err != nil {
		return fmt.Errorf("failed to release freeze: %w", err)
	}

	// Update wallet status
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet != nil {
		wallet.Status = models.WalletStatusActive
		s.walletRepo.Update(ctx, wallet)
	}

	s.logAudit(ctx, "WALLET_FREEZE", freeze.ID, "RELEASE", actorID, actorName, nil, map[string]interface{}{
		"reason": reason,
	}, true, "")

	return nil
}

// GetFreezeStatus retrieves the freeze status for a wallet
func (s *FreezeService) GetFreezeStatus(ctx context.Context, walletID uuid.UUID) (*models.WalletFreeze, error) {
	return s.freezeRepo.GetActiveByWallet(ctx, walletID)
}

// GetActiveFreezes retrieves all active freezes
func (s *FreezeService) GetActiveFreezes(ctx context.Context) ([]*models.WalletFreeze, error) {
	return s.freezeRepo.GetActiveFreezes(ctx)
}

// GetFreezeHistory retrieves the freeze history for a wallet
func (s *FreezeService) GetFreezeHistory(ctx context.Context, walletID uuid.UUID) ([]*models.WalletFreeze, error) {
	filter := &models.FreezeFilter{
		WalletIDs: []uuid.UUID{walletID},
	}
	return s.freezeRepo.List(ctx, filter, 100, 0)
}

// StartFreezeExpiryChecker starts the background task to check for expired freezes
func (s *FreezeService) StartFreezeExpiryChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			expired, err := s.freezeRepo.GetExpiredFreezes(ctx)
			if err != nil {
				continue
			}

			for _, freeze := range expired {
				freeze.Status = models.FreezeStatusExpired
				s.freezeRepo.Update(ctx, freeze)
			}
		case <-s.stopChan:
			return
		}
	}
}

// StopFreezeExpiryChecker stops the freeze expiry checker
func (s *FreezeService) StopFreezeExpiryChecker() {
	close(s.stopChan)
}

// logAudit logs an audit event
func (s *FreezeService) logAudit(ctx context.Context, entityType string, entityID uuid.UUID, action string, actorID uuid.UUID, actorName string, oldValue, newValue interface{}, success bool, errorMsg string) {
	log := &models.WalletAuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		ActorID:    actorID,
		ActorName:  actorName,
		ActorType:  "USER",
		NewValue:   toJSONMap(newValue),
		Success:    success,
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
