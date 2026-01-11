package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/config"
	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/csic/wallet-governance/internal/repository"
	"github.com/google/uuid"
)

// WalletService handles wallet management operations
type WalletService struct {
	walletRepo   repository.WalletRepository
	blacklistRepo repository.BlacklistRepository
	whitelistRepo repository.WhitelistRepository
	freezeRepo   repository.WalletFreezeRepository
	auditRepo    repository.AuditRepository
}

// NewWalletService creates a new wallet service
func NewWalletService(
	walletRepo repository.WalletRepository,
	blacklistRepo repository.BlacklistRepository,
	whitelistRepo repository.WhitelistRepository,
	freezeRepo repository.WalletFreezeRepository,
	auditRepo repository.AuditRepository,
) *WalletService {
	return &WalletService{
		walletRepo:   walletRepo,
		blacklistRepo: blacklistRepo,
		whitelistRepo: whitelistRepo,
		freezeRepo:   freezeRepo,
		auditRepo:    auditRepo,
	}
}

// RegisterWallet registers a new wallet
func (s *WalletService) RegisterWallet(ctx context.Context, wallet *models.Wallet, actorID uuid.UUID, actorName string) error {
	// Check if address is already registered
	existing, err := s.walletRepo.GetByAddress(ctx, wallet.Address, wallet.Blockchain)
	if err != nil {
		return fmt.Errorf("failed to check existing wallet: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("wallet address already registered")
	}

	// Check if address is blacklisted
	blacklisted, err := s.blacklistRepo.IsBlacklisted(ctx, wallet.Address, wallet.Blockchain)
	if err != nil {
		return fmt.Errorf("failed to check blacklist: %w", err)
	}
	if blacklisted {
		return fmt.Errorf("wallet address is blacklisted")
	}

	// Generate wallet ID
	wallet.WalletID = generateWalletID(wallet.Blockchain)
	wallet.Status = models.WalletStatusActive
	wallet.IsBlacklisted = false
	wallet.ComplianceScore = 100.0

	// Calculate address checksum
	wallet.AddressChecksum = calculateAddressHash(wallet.Address)

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	// Log audit event
	s.logAudit(ctx, "WALLET", wallet.ID, "CREATE", actorID, actorName, nil, wallet, true, "")

	return nil
}

// GetWallet retrieves a wallet by ID
func (s *WalletService) GetWallet(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	return s.walletRepo.GetByID(ctx, id)
}

// GetWalletByAddress retrieves a wallet by address
func (s *WalletService) GetWalletByAddress(ctx context.Context, address string, blockchain models.BlockchainType) (*models.Wallet, error) {
	return s.walletRepo.GetByAddress(ctx, address, blockchain)
}

// UpdateWallet updates a wallet
func (s *WalletService) UpdateWallet(ctx context.Context, wallet *models.Wallet, actorID uuid.UUID, actorName string) error {
	// Get existing wallet for audit
	existing, err := s.walletRepo.GetByID(ctx, wallet.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing wallet: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("wallet not found")
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	s.logAudit(ctx, "WALLET", wallet.ID, "UPDATE", actorID, actorName, existing, wallet, true, "")

	return nil
}

// RevokeWallet revokes a wallet
func (s *WalletService) RevokeWallet(ctx context.Context, id uuid.UUID, reason string, actorID uuid.UUID, actorName string) error {
	if err := s.walletRepo.Revoke(ctx, id, reason); err != nil {
		return fmt.Errorf("failed to revoke wallet: %w", err)
	}

	s.logAudit(ctx, "WALLET", id, "REVOKE", actorID, actorName, nil, map[string]interface{}{
		"reason": reason,
	}, true, "")

	return nil
}

// ListWallets lists wallets with filters
func (s *WalletService) ListWallets(ctx context.Context, filter *models.WalletFilter, limit, offset int) ([]*models.Wallet, int, error) {
	wallets, err := s.walletRepo.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list wallets: %w", err)
	}

	count, err := s.walletRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count wallets: %w", err)
	}

	return wallets, count, nil
}

// GetWalletSummary retrieves wallet statistics
func (s *WalletService) GetWalletSummary(ctx context.Context) (*models.WalletSummary, error) {
	return s.walletRepo.GetSummary(ctx)
}

// RegisterCustodialWallet registers a custodial wallet
func (s *WalletService) RegisterCustodialWallet(ctx context.Context, wallet *models.Wallet, actorID uuid.UUID, actorName string) error {
	wallet.Type = models.WalletTypeCustodial
	return s.RegisterWallet(ctx, wallet, actorID, actorName)
}

// RegisterMultiSigWallet registers a multi-signature wallet
func (s *WalletService) RegisterMultiSigWallet(ctx context.Context, wallet *models.Wallet, signers []*models.WalletSigner, actorID uuid.UUID, actorName string) error {
	wallet.Type = models.WalletTypeMultiSig
	wallet.SignersTotal = len(signers)

	if err := s.RegisterWallet(ctx, wallet, actorID, actorName); err != nil {
		return err
	}

	return nil
}

// RegisterExchangeHotWallet registers an exchange hot wallet
func (s *WalletService) RegisterExchangeHotWallet(ctx context.Context, wallet *models.Wallet, actorID uuid.UUID, actorName string) error {
	wallet.Type = models.WalletTypeExchangeHot
	return s.RegisterWallet(ctx, wallet, actorID, actorName)
}

// RegisterExchangeColdWallet registers an exchange cold wallet
func (s *WalletService) RegisterExchangeColdWallet(ctx context.Context, wallet *models.Wallet, actorID uuid.UUID, actorName string) error {
	wallet.Type = models.WalletTypeExchangeCold
	return s.RegisterWallet(ctx, wallet, actorID, actorName)
}

// CheckWalletCompliance performs compliance checks on a wallet
func (s *WalletService) CheckWalletCompliance(ctx context.Context, walletID uuid.UUID) (*models.ComplianceCheck, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	check := &models.ComplianceCheck{
		WalletID:        walletID,
		ChecksPerformed: []string{},
		RiskFactors:     []string{},
		Recommendations: []string{},
		CheckedAt:       time.Now(),
	}

	// Check blacklist
	blacklisted, err := s.blacklistRepo.IsBlacklisted(ctx, wallet.Address, wallet.Blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if blacklisted {
		entry, _ := s.blacklistRepo.GetByAddress(ctx, wallet.Address, wallet.Blockchain)
		if entry != nil {
			check.IsBlacklisted = true
			check.BlacklistReason = entry.Reason
			check.RiskFactors = append(check.RiskFactors, "Address is blacklisted")
			check.Status = models.ComplianceStatusNonCompliant
		}
	}
	check.ChecksPerformed = append(check.ChecksPerformed, "blacklist_check")

	// Check freeze status
	freeze, err := s.freezeRepo.GetActiveByWallet(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to check freeze: %w", err)
	}
	if freeze != nil {
		check.IsFrozen = true
		check.FreezeReason = string(freeze.Reason)
		check.RiskFactors = append(check.RiskFactors, "Wallet is frozen")
		if check.Status == "" {
			check.Status = models.ComplianceStatusSuspended
		}
	}
	check.ChecksPerformed = append(check.ChecksPerformed, "freeze_check")

	// Check whitelist status
	whitelisted, err := s.whitelistRepo.IsWhitelisted(ctx, wallet.Address, wallet.Blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to check whitelist: %w", err)
	}
	check.IsWhitelisted = whitelisted
	check.ChecksPerformed = append(check.ChecksPerformed, "whitelist_check")

	// Set compliance score
	if !check.IsBlacklisted && !check.IsFrozen {
		check.ComplianceScore = wallet.ComplianceScore.Float64()
		check.Status = models.ComplianceStatusCompliant
	} else {
		check.ComplianceScore = 0
	}

	// Add recommendations
	if !whitelisted {
		check.Recommendations = append(check.Recommendations, "Consider whitelisting trusted addresses")
	}
	if len(check.RiskFactors) > 0 {
		check.Recommendations = append(check.Recommendations, "Review and address risk factors")
	}

	return check, nil
}

// logAudit logs an audit event
func (s *WalletService) logAudit(ctx context.Context, entityType string, entityID uuid.UUID, action string, actorID uuid.UUID, actorName string, oldValue, newValue interface{}, success bool, errorMsg string) {
	log := &models.WalletAuditLog{
		EntityType:   entityType,
		EntityID:     entityID,
		Action:       action,
		ActorID:      actorID,
		ActorName:    actorName,
		ActorType:    "USER",
		NewValue:     toJSONMap(newValue),
		Success:      success,
		ErrorMessage: errorMsg,
	}

	if oldValue != nil {
		log.OldValue = toJSONMap(oldValue)
	}

	s.auditRepo.Create(ctx, log)
}

// Helper functions
func generateWalletID(blockchain models.BlockchainType) string {
	prefix := map[models.BlockchainType]string{
		models.BlockchainBitcoin:   "BTC",
		models.BlockchainEthereum:  "ETH",
		models.BlockchainERC20:     "ERC",
		models.BlockchainTRC20:     "TRC",
		models.BlockchainBEP20:     "BEP",
		models.BlockchainPolygon:   "POL",
		models.BlockchainSolana:    "SOL",
		models.BlockchainLitecoin:  "LTC",
		models.BlockchainDash:      "DASH",
		models.BlockchainMonero:    "XMR",
	}[blockchain]

	return fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8])
}

func calculateAddressHash(address string) string {
	hash := sha256.Sum256([]byte(address))
	return hex.EncodeToString(hash[:])[:16]
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
