package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic/wallet-governance/internal/domain/models"
	"github.com/csic/wallet-governance/internal/repository"
	"github.com/google/uuid"
)

// ComplianceService handles compliance operations
type ComplianceService struct {
	walletRepo    repository.WalletRepository
	blacklistRepo repository.BlacklistRepository
	whitelistRepo repository.WhitelistRepository
	freezeRepo    repository.WalletFreezeRepository
	auditRepo     repository.AuditRepository
}

// NewComplianceService creates a new compliance service
func NewComplianceService(
	walletRepo repository.WalletRepository,
	blacklistRepo repository.BlacklistRepository,
	whitelistRepo repository.WhitelistRepository,
	freezeRepo repository.WalletFreezeRepository,
	auditRepo repository.AuditRepository,
) *ComplianceService {
	return &ComplianceService{
		walletRepo:    walletRepo,
		blacklistRepo: blacklistRepo,
		whitelistRepo: whitelistRepo,
		freezeRepo:    freezeRepo,
		auditRepo:     auditRepo,
	}
}

// CheckWalletCompliance performs comprehensive compliance checks
func (s *ComplianceService) CheckWalletCompliance(ctx context.Context, walletID uuid.UUID) (*models.ComplianceCheck, error) {
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

	// Check 1: Blacklist
	blacklisted, err := s.blacklistRepo.IsBlacklisted(ctx, wallet.Address, wallet.Blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if blacklisted {
		entry, _ := s.blacklistRepo.GetByAddress(ctx, wallet.Address, wallet.Blockchain)
		if entry != nil {
			check.IsBlacklisted = true
			check.BlacklistReason = fmt.Sprintf("Blacklisted: %s - %s", entry.Reason, entry.ReasonDetails)
			check.RiskFactors = append(check.RiskFactors, "Address on blacklist")
		}
	}
	check.ChecksPerformed = append(check.ChecksPerformed, "blacklist")

	// Check 2: Freeze status
	freeze, err := s.freezeRepo.GetActiveByWallet(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to check freeze: %w", err)
	}
	if freeze != nil {
		check.IsFrozen = true
		check.FreezeReason = fmt.Sprintf("Frozen: %s - %s", freeze.Reason, freeze.ReasonDetails)
		check.RiskFactors = append(check.RiskFactors, "Wallet is frozen")
	}
	check.ChecksPerformed = append(check.ChecksPerformed, "freeze")

	// Check 3: Whitelist status
	whitelisted, err := s.whitelistRepo.IsWhitelisted(ctx, wallet.Address, wallet.Blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to check whitelist: %w", err)
	}
	check.IsWhitelisted = whitelisted
	check.ChecksPerformed = append(check.ChecksPerformed, "whitelist")

	// Check 4: Wallet status
	if wallet.Status == models.WalletStatusSuspended || wallet.Status == models.WalletStatusRevoked {
		check.RiskFactors = append(check.RiskFactors, fmt.Sprintf("Wallet status: %s", wallet.Status))
	}
	check.ChecksPerformed = append(check.ChecksPerformed, "status")

	// Calculate compliance status
	if check.IsBlacklisted {
		check.Status = models.ComplianceStatusNonCompliant
		check.ComplianceScore = 0
	} else if check.IsFrozen {
		check.Status = models.ComplianceStatusSuspended
		check.ComplianceScore = 0
	} else if wallet.Status == models.WalletStatusActive {
		check.Status = models.ComplianceStatusCompliant
		if whitelisted {
			check.ComplianceScore = 100
		} else {
			check.ComplianceScore = float64(wallet.ComplianceScore)
		}
	} else {
		check.Status = models.ComplianceStatusUnderReview
		check.ComplianceScore = 50
	}

	// Generate recommendations
	if !whitelisted && !check.IsBlacklisted {
		check.Recommendations = append(check.Recommendations, "Consider whitelist verification for enhanced compliance")
	}
	if len(check.RiskFactors) > 0 {
		check.Recommendations = append(check.Recommendations, "Review risk factors and take corrective action")
	}
	if wallet.ComplianceScore < 80 {
		check.Recommendations = append(check.Recommendations, "Improve compliance score through regular audits")
	}

	return check, nil
}

// CheckMultipleWallets checks compliance for multiple wallets
func (s *ComplianceService) CheckMultipleWallets(ctx context.Context, walletIDs []uuid.UUID) ([]*models.ComplianceCheck, error) {
	results := make([]*models.ComplianceCheck, 0, len(walletIDs))

	for _, walletID := range walletIDs {
		check, err := s.CheckWalletCompliance(ctx, walletID)
		if err != nil {
			continue
		}
		results = append(results, check)
	}

	return results, nil
}

// GetWalletAuditTrail retrieves the audit trail for a wallet
func (s *ComplianceService) GetWalletAuditTrail(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*models.WalletAuditLog, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	return s.auditRepo.GetByEntity(ctx, "WALLET", walletID, limit, offset)
}

// AddToBlacklist adds an address to the blacklist
func (s *ComplianceService) AddToBlacklist(ctx context.Context, entry *models.BlacklistEntry, actorID uuid.UUID, actorName string) error {
	entry.AddedBy = actorID
	entry.AddedByName = actorName
	entry.Status = models.BlacklistStatusActive

	if err := s.blacklistRepo.Create(ctx, entry); err != nil {
		return fmt.Errorf("failed to add to blacklist: %w", err)
	}

	s.logAudit(ctx, "BLACKLIST", entry.ID, "ADD", actorID, actorName, nil, entry, true, "")

	return nil
}

// RemoveFromBlacklist removes an address from the blacklist
func (s *ComplianceService) RemoveFromBlacklist(ctx context.Context, address string, blockchain models.BlockchainType, removedBy uuid.UUID, removedByName, reason string) error {
	entry, err := s.blacklistRepo.GetByAddress(ctx, address, blockchain)
	if err != nil {
		return fmt.Errorf("failed to get blacklist entry: %w", err)
	}
	if entry == nil {
		return fmt.Errorf("address not found in blacklist")
	}

	if err := s.blacklistRepo.Remove(ctx, entry.ID, removedBy, reason); err != nil {
		return fmt.Errorf("failed to remove from blacklist: %w", err)
	}

	s.logAudit(ctx, "BLACKLIST", entry.ID, "REMOVE", removedBy, removedByName, nil, map[string]interface{}{
		"address": address,
		"reason":  reason,
	}, true, "")

	return nil
}

// IsBlacklisted checks if an address is blacklisted
func (s *ComplianceService) IsBlacklisted(ctx context.Context, address string, blockchain models.BlockchainType) (bool, error) {
	return s.blacklistRepo.IsBlacklisted(ctx, address, blockchain)
}

// IsWhitelisted checks if an address is whitelisted
func (s *ComplianceService) IsWhitelisted(ctx context.Context, address string, blockchain models.BlockchainType) (bool, error) {
	return s.whitelistRepo.IsWhitelisted(ctx, address, blockchain)
}

// GetBlacklist retrieves the blacklist
func (s *ComplianceService) GetBlacklist(ctx context.Context, limit, offset int) ([]*models.BlacklistEntry, int, error) {
	filter := &models.BlacklistFilter{
		Statuses: []models.BlacklistStatus{models.BlacklistStatusActive},
	}

	entries, err := s.blacklistRepo.List(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get blacklist: %w", err)
	}

	count, err := s.blacklistRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count blacklist: %w", err)
	}

	return entries, count, nil
}

// GetWhitelist retrieves the whitelist
func (s *ComplianceService) GetWhitelist(ctx context.Context, limit, offset int) ([]*models.WhitelistEntry, int, error) {
	entries, err := s.whitelistRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get whitelist: %w", err)
	}

	count, err := s.whitelistRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count whitelist: %w", err)
	}

	return entries, count, nil
}

// logAudit logs an audit event
func (s *ComplianceService) logAudit(ctx context.Context, entityType string, entityID uuid.UUID, action string, actorID uuid.UUID, actorName string, oldValue, newValue interface{}, success bool, errorMsg string) {
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
