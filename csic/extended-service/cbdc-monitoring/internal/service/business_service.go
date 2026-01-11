package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"csic-platform/extended-services/cbdc-monitoring/internal/domain"
	"csic-platform/extended-services/cbdc-monitoring/internal/repository"
	"csic-platform/shared/config"
	"csic-platform/shared/logger"
	"csic-platform/shared/queue"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrInvalidAmount          = errors.New("invalid transaction amount")
	ErrExceedsLimit           = errors.New("transaction exceeds allowed limit")
	ErrVelocityExceeded       = errors.New("transaction velocity exceeded")
	ErrRestrictedCategory     = errors.New("transaction category is restricted")
	ErrGeographicRestriction  = errors.New("transaction from restricted location")
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrWalletNotActive        = errors.New("wallet is not active")
	ErrPolicyValidationFailed = errors.New("policy validation failed")
)

// CBDCMonitoringService handles business logic for CBDC operations
type CBDCMonitoringService struct {
	walletRepo  *repository.WalletRepository
	txRepo      *repository.TransactionRepository
	policyRepo  *repository.PolicyRepository
	violationRepo *repository.ViolationRepository
	auditRepo   *repository.AuditRepository
	redis       *queue.RedisClient
	kafka       *queue.KafkaManager
	cfg         *config.Config
	log         *logger.Logger
}

// NewCBDCMonitoringService creates a new CBDC monitoring service
func NewCBDCMonitoringService(db *gorm.DB, redis *queue.RedisClient, kafka *queue.KafkaManager, cfg *config.Config, log *logger.Logger) *CBDCMonitoringService {
	return &CBDCMonitoringService{
		walletRepo:    repository.NewWalletRepository(db),
		txRepo:        repository.NewTransactionRepository(db),
		policyRepo:    repository.NewPolicyRepository(db),
		violationRepo: repository.NewViolationRepository(db),
		auditRepo:     repository.NewAuditRepository(db),
		redis:         redis,
		kafka:         kafka,
		cfg:           cfg,
		log:           log,
	}
}

// CreateWallet creates a new CBDC wallet
func (s *CBDCMonitoringService) CreateWallet(ctx context.Context, req CreateWalletRequest) (*domain.Wallet, error) {
	// Validate request
	if req.OwnerID == "" {
		return nil, errors.New("owner_id is required")
	}

	// Generate wallet ID and number
	wallet := &domain.Wallet{
		ID:           uuid.New().String(),
		OwnerID:      req.OwnerID,
		WalletNumber: generateWalletNumber(),
		Tier:         domain.WalletTier(req.Tier),
		Status:       domain.WalletStatusActive,
		Balance:      decimal.Zero,
		FrozenBalance: decimal.Zero,
		Currency:     "CBDC",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     marshalJSON(req.Metadata),
	}

	// Set KYC verification time if tier requires it
	if wallet.Tier >= domain.Tier2Standard {
		now := time.Now()
		wallet.KYCVerifiedAt = &now
	}

	// Create wallet in database
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	// Create audit log
	s.createAuditLog(ctx, "wallet", wallet.ID, "create", "system", "system", nil, wallet)

	// Publish wallet created event
	s.publishWalletEvent(ctx, "wallet.created", wallet)

	s.log.Infof("Created wallet %s for owner %s", wallet.ID, wallet.OwnerID)
	return wallet, nil
}

// GetWallet retrieves a wallet by ID
func (s *CBDCMonitoringService) GetWallet(ctx context.Context, id string) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// UpdateWallet updates wallet information
func (s *CBDCMonitoringService) UpdateWallet(ctx context.Context, id string, req UpdateWalletRequest) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldWallet := cloneWallet(wallet)

	// Update fields
	if req.Tier > 0 {
		wallet.Tier = domain.WalletTier(req.Tier)
	}
	if req.Status > 0 {
		wallet.Status = domain.WalletStatus(req.Status)
	}
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return nil, err
	}

	// Create audit log
	s.createAuditLog(ctx, "wallet", wallet.ID, "update", "system", "system", oldWallet, wallet)

	s.log.Infof("Updated wallet %s", wallet.ID)
	return wallet, nil
}

// GetWalletBalance retrieves wallet balance with caching
func (s *CBDCMonitoringService) GetWalletBalance(ctx context.Context, id string) (decimal.Decimal, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("balance:%s", id)
	if cached, err := s.redis.Get(ctx, cacheKey); err == nil {
		if balance, err := decimal.NewFromString(cached); err == nil {
			return balance, nil
		}
	}

	wallet, err := s.walletRepo.GetByID(ctx, id)
	if err != nil {
		return decimal.Zero, err
	}

	// Cache the balance
	s.redis.Set(ctx, cacheKey, wallet.Balance.String(), time.Minute*5)

	return wallet.Balance, nil
}

// CreateTransaction creates and processes a new transaction
func (s *CBDCMonitoringService) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*domain.Transaction, error) {
	// Validate request
	if req.WalletID == "" {
		return nil, errors.New("wallet_id is required")
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Get wallet
	wallet, err := s.walletRepo.GetByID(ctx, req.WalletID)
	if err != nil {
		return nil, err
	}

	// Check wallet status
	if wallet.Status != domain.WalletStatusActive {
		return nil, ErrWalletNotActive
	}

	// Check sufficient balance
	if wallet.Balance.LessThan(req.Amount) {
		return nil, ErrInsufficientFunds
	}

	// Check tier limits
	tierLimits := s.getTierLimits(wallet.Tier)
	if req.Amount.GreaterThan(tierLimits.MaxTransaction) {
		return nil, fmt.Errorf("%w: exceeds tier %d limit of %s", ErrExceedsLimit, wallet.Tier, tierLimits.MaxTransaction)
	}

	// Create transaction
	tx := &domain.Transaction{
		ID:             uuid.New().String(),
		WalletID:       req.WalletID,
		CounterpartyID: req.CounterpartyID,
		Counterparty:   req.Counterparty,
		TransactionType: domain.TransactionType(req.Type),
		Status:         domain.TxStatusPending,
		Amount:         req.Amount,
		Currency:       "CBDC",
		Fee:            s.calculateFee(req.Amount, wallet.Tier),
		NetAmount:      req.Amount.Sub(s.calculateFee(req.Amount, wallet.Tier)),
		Description:    req.Description,
		Metadata:       marshalJSON(req.Metadata),
		Programmable:   marshalJSON(req.Programmable),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Validate against policies
	if err := s.validateTransaction(ctx, tx, wallet); err != nil {
		// Mark transaction as rejected due to policy violation
		tx.Status = domain.TxStatusRejected
		tx.PolicyViolated = true
		tx.ViolationReason = err.Error()
		tx.CreatedAt = time.Now()
		tx.UpdatedAt = time.Now()

		if createErr := s.txRepo.Create(ctx, tx); createErr != nil {
			return nil, createErr
		}

		// Create violation record
		violation := &domain.ViolationRecord{
			ID:            uuid.New().String(),
			TransactionID: tx.ID,
			WalletID:      wallet.ID,
			ViolationType: "policy_violation",
			Details:       marshalJSON(map[string]interface{}{"error": err.Error(), "tx": tx}),
			CreatedAt:     time.Now(),
		}
		s.violationRepo.Create(ctx, violation)

		// Publish violation event
		s.publishTransactionEvent(ctx, "transaction.violation", tx)

		return nil, fmt.Errorf("%w: %s", ErrPolicyValidationFailed, err.Error())
	}

	// Save transaction
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// Update wallet balance (hold funds)
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, req.Amount, false); err != nil {
		return nil, err
	}

	// Update velocity check
	s.updateVelocityCheck(ctx, wallet.ID, req.Amount)

	// Create audit log
	s.createAuditLog(ctx, "transaction", tx.ID, "create", "system", "system", nil, tx)

	// Publish transaction created event
	s.publishTransactionEvent(ctx, "transaction.proposed", tx)

	s.log.Infof("Created transaction %s for wallet %s, amount: %s", tx.ID, wallet.ID, tx.Amount.String())
	return tx, nil
}

// ApproveTransaction approves a pending transaction
func (s *CBDCMonitoringService) ApproveTransaction(ctx context.Context, txID string, approverID string) (*domain.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	if tx.Status != domain.TxStatusPending {
		return nil, errors.New("transaction is not pending")
	}

	now := time.Now()
	tx.Status = domain.TxStatusApproved
	tx.ApprovedBy = approverID
	tx.ApprovedAt = &now
	tx.ProcessedAt = &now
	tx.UpdatedAt = now

	if err := s.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}

	// Create audit log
	s.createAuditLog(ctx, "transaction", tx.ID, "approve", approverID, "user", nil, tx)

	// Publish approval event
	s.publishTransactionEvent(ctx, "transaction.approved", tx)

	s.log.Infof("Approved transaction %s by %s", txID, approverID)
	return tx, nil
}

// RejectTransaction rejects a pending transaction
func (s *CBDCMonitoringService) RejectTransaction(ctx context.Context, txID string, reason string) (*domain.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	if tx.Status != domain.TxStatusPending {
		return nil, errors.New("transaction is not pending")
	}

	// Refund the held amount
	wallet, err := s.walletRepo.GetByID(ctx, tx.WalletID)
	if err == nil {
		s.walletRepo.UpdateBalance(ctx, wallet.ID, tx.Amount, true)
	}

	now := time.Now()
	tx.Status = domain.TxStatusRejected
	tx.RejectedReason = reason
	tx.ProcessedAt = &now
	tx.UpdatedAt = now

	if err := s.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}

	// Create audit log
	s.createAuditLog(ctx, "transaction", tx.ID, "reject", "system", "system", nil, tx)

	// Publish rejection event
	s.publishTransactionEvent(ctx, "transaction.rejected", tx)

	s.log.Infof("Rejected transaction %s: %s", txID, reason)
	return tx, nil
}

// GetTransaction retrieves a transaction by ID
func (s *CBDCMonitoringService) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}

// ListTransactions retrieves transactions with filters
func (s *CBDCMonitoringService) ListTransactions(ctx context.Context, filters TransactionListFilters) ([]domain.Transaction, int64, error) {
	repoFilters := repository.TransactionFilters{
		WalletID:  filters.WalletID,
		Status:    (*domain.TransactionStatus)(filters.Status),
		Type:      (*domain.TransactionType)(filters.Type),
		StartDate: filters.StartDate,
		EndDate:   filters.EndDate,
		Page:      filters.Page,
		Limit:     filters.Limit,
	}
	return s.txRepo.List(ctx, repoFilters)
}

// CreatePolicy creates a new policy rule
func (s *CBDCMonitoringService) CreatePolicy(ctx context.Context, req CreatePolicyRequest) (*domain.PolicyRule, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.RuleType == "" {
		return nil, errors.New("rule_type is required")
	}
	if req.Action == "" {
		return nil, errors.New("action is required")
	}

	policy := &domain.PolicyRule{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		RuleType:    req.RuleType,
		Condition:   marshalJSON(req.Condition),
		Action:      req.Action,
		Priority:    req.Priority,
		Enabled:     true,
		Scope:       req.Scope,
		TargetIDs:   req.TargetIDs,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.policyRepo.Create(ctx, policy); err != nil {
		return nil, err
	}

	s.log.Infof("Created policy rule %s: %s", policy.ID, policy.Name)
	return policy, nil
}

// GetPolicy retrieves a policy rule by ID
func (s *CBDCMonitoringService) GetPolicy(ctx context.Context, id string) (*domain.PolicyRule, error) {
	return s.policyRepo.GetByID(ctx, id)
}

// ListPolicies retrieves all enabled policy rules
func (s *CBDCMonitoringService) ListPolicies(ctx context.Context, ruleType string) ([]domain.PolicyRule, error) {
	return s.policyRepo.List(ctx, ruleType)
}

// GetMonitoringStats retrieves aggregated monitoring statistics
func (s *CBDCMonitoringService) GetMonitoringStats(ctx context.Context) (*domain.MonitoringStats, error) {
	stats := &domain.MonitoringStats{}

	// Count wallets
	if err := s.walletRepo.db.WithContext(ctx).Table("wallets").Count(&stats.TotalWallets).Error; err != nil {
		return nil, err
	}

	// Count active wallets
	if err := s.walletRepo.db.WithContext(ctx).Table("wallets").
		Where("status = ?", domain.WalletStatusActive).
		Count(&stats.ActiveWallets).Error; err != nil {
		return nil, err
	}

	// Count transactions
	if err := s.txRepo.db.WithContext(ctx).Table("transactions").Count(&stats.TotalTransactions).Error; err != nil {
		return nil, err
	}

	// Count pending transactions
	if err := s.txRepo.db.WithContext(ctx).Table("transactions").
		Where("status = ?", domain.TxStatusPending).
		Count(&stats.PendingTransactions).Error; err != nil {
		return nil, err
	}

	// Get today's stats
	approved, rejected, volume, err := s.txRepo.GetTodayStats(ctx)
	if err != nil {
		return nil, err
	}
	stats.ApprovedToday = approved
	stats.RejectedToday = rejected
	stats.VolumeToday = volume

	// Get violation count
	violations, err := s.violationRepo.GetTodayCount(ctx)
	if err != nil {
		return nil, err
	}
	stats.ViolationsToday = violations

	return stats, nil
}

// ComplianceCheck performs a compliance check on a transaction
func (s *CBDCMonitoringService) ComplianceCheck(ctx context.Context, req ComplianceCheckRequest) (*ComplianceCheckResult, error) {
	result := &ComplianceCheckResult{
		Passed: true,
		Checks: []CheckResult{},
	}

	// Check wallet status
	wallet, err := s.walletRepo.GetByID(ctx, req.WalletID)
	if err != nil {
		result.Passed = false
		result.Checks = append(result.Checks, CheckResult{
			CheckName: "wallet_exists",
			Passed:    false,
			Message:   "Wallet not found",
		})
	} else {
		result.Checks = append(result.Checks, CheckResult{
			CheckName: "wallet_status",
			Passed:    wallet.Status == domain.WalletStatusActive,
			Message:   fmt.Sprintf("Wallet status: %s", wallet.Status),
		})

		// Check tier limits
		limits := s.getTierLimits(wallet.Tier)
		if req.Amount.GreaterThan(limits.MaxTransaction) {
			result.Passed = false
			result.Checks = append(result.Checks, CheckResult{
				CheckName: "tier_limit",
				Passed:    false,
				Message:   fmt.Sprintf("Amount %s exceeds tier limit %s", req.Amount, limits.MaxTransaction),
			})
		} else {
			result.Checks = append(result.Checks, CheckResult{
				CheckName: "tier_limit",
				Passed:    true,
				Message:   "Within tier limits",
			})
		}
	}

	// Check restricted categories
	if s.isRestrictedCategory(req.Category) {
		result.Passed = false
		result.Checks = append(result.Checks, CheckResult{
			CheckName: "category_restriction",
			Passed:    false,
			Message:   "Transaction category is restricted",
		})
	}

	// Check velocity
	velocityOK := s.checkVelocity(ctx, req.WalletID, req.Amount)
	result.Checks = append(result.Checks, CheckResult{
		CheckName: "velocity_check",
		Passed:    velocityOK,
		Message:   "Velocity check passed",
	})

	if !velocityOK {
		result.Passed = false
	}

	return result, nil
}

// Helper functions

func (s *CBDCMonitoringService) validateTransaction(ctx context.Context, tx *domain.Transaction, wallet *domain.Wallet) error {
	// Get applicable policies
	policies, err := s.policyRepo.List(ctx, "")
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if s.evaluatePolicy(policy, tx, wallet) {
			if policy.Action == "block" {
				return fmt.Errorf("violated policy: %s", policy.Name)
			}
		}
	}

	return nil
}

func (s *CBDCMonitoringService) evaluatePolicy(policy domain.PolicyRule, tx *domain.Transaction, wallet *domain.Wallet) bool {
	// Simple policy evaluation - in production this would be more sophisticated
	switch policy.RuleType {
	case "limit":
		// Check amount limits
		var condition map[string]interface{}
		if err := json.Unmarshal([]byte(policy.Condition), &condition); err != nil {
			return false
		}
		if maxAmount, ok := condition["max_amount"].(float64); ok {
			if tx.Amount.GreaterThan(decimal.NewFromFloat(maxAmount)) {
				return true
			}
		}
	case "velocity":
		// Check velocity
		if !s.checkVelocity(ctx, tx.WalletID, tx.Amount) {
			return true
		}
	}

	return false
}

func (s *CBDCMonitoringService) checkVelocity(ctx context.Context, walletID string, amount decimal.Decimal) bool {
	cacheKey := fmt.Sprintf("velocity:%s", walletID)
	velocityData, _ := s.redis.HGetAll(ctx, cacheKey)

	var currentTotal decimal.Decimal
	if totalStr, ok := velocityData["total"]; ok {
		currentTotal, _ = decimal.NewFromString(totalStr)
	}

	newTotal := currentTotal.Add(amount)
	limit := decimal.NewFromFloat(s.cfg.Monitoring.Thresholds.MaxTransactionThreshold)

	return newTotal.LessThanOrEqual(limit)
}

func (s *CBDCMonitoringService) updateVelocityCheck(ctx context.Context, walletID string, amount decimal.Decimal) {
	cacheKey := fmt.Sprintf("velocity:%s", walletID)
	s.redis.HIncrBy(ctx, cacheKey, "count", 1)
	s.redis.HIncrByFloat(ctx, cacheKey, "total", amount.String())
	s.redis.Expire(ctx, cacheKey, time.Hour)
}

func (s *CBDCMonitoringService) getTierLimits(tier domain.WalletTier) TierLimits {
	switch tier {
	case domain.Tier1Basic:
		return TierLimits{
			MaxTransaction: decimal.NewFromFloat(1000),
			DailyLimit:     decimal.NewFromFloat(5000),
		}
	case domain.Tier2Standard:
		return TierLimits{
			MaxTransaction: decimal.NewFromFloat(10000),
			DailyLimit:     decimal.NewFromFloat(100000),
		}
	case domain.Tier3Premium:
		return TierLimits{
			MaxTransaction: decimal.NewFromFloat(50000),
			DailyLimit:     decimal.NewFromFloat(500000),
		}
	default:
		return TierLimits{
			MaxTransaction: decimal.NewFromFloat(1000),
			DailyLimit:     decimal.NewFromFloat(5000),
		}
	}
}

func (s *CBDCMonitoringService) calculateFee(amount decimal.Decimal, tier domain.WalletTier) decimal.Decimal {
	// Fee schedule based on tier
	var rate decimal.Decimal
	switch tier {
	case domain.Tier1Basic:
		rate = decimal.NewFromFloat(0.01) // 1%
	case domain.Tier2Standard:
		rate = decimal.NewFromFloat(0.005) // 0.5%
	case domain.Tier3Premium:
		rate = decimal.NewFromFloat(0.002) // 0.2%
	default:
		rate = decimal.NewFromFloat(0.01)
	}
	return amount.Mul(rate).Round(8)
}

func (s *CBDCMonitoringService) isRestrictedCategory(category string) bool {
	for _, restricted := range s.cfg.Monitoring.ProgrammableMoney.RestrictedCategories {
		if category == restricted {
			return true
		}
	}
	return false
}

func (s *CBDCMonitoringService) createAuditLog(ctx context.Context, entityType, entityID, action, actorID, actorType string, oldValue, newValue interface{}) {
	log := &domain.AuditLog{
		ID:        uuid.New().String(),
		EntityType: entityType,
		EntityID:  entityID,
		Action:    action,
		ActorID:   actorID,
		ActorType: actorType,
		OldValue:  marshalJSON(oldValue),
		NewValue:  marshalJSON(newValue),
		Timestamp: time.Now(),
	}
	s.auditRepo.Create(ctx, log)
}

func (s *CBDCMonitoringService) publishWalletEvent(ctx context.Context, eventType string, wallet *domain.Wallet) {
	data, _ := json.Marshal(wallet)
	s.kafka.Publish(ctx, s.cfg.Kafka.Topics.Wallets, eventType, data)
}

func (s *CBDCMonitoringService) publishTransactionEvent(ctx context.Context, eventType string, tx *domain.Transaction) {
	data, _ := json.Marshal(tx)
	s.kafka.Publish(ctx, s.cfg.Kafka.Topics.Transactions, eventType, data)
}

// Request/Response types

type CreateWalletRequest struct {
	OwnerID  string                 `json:"owner_id"`
	Tier     int                    `json:"tier"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateWalletRequest struct {
	Tier   int    `json:"tier,omitempty"`
	Status int    `json:"status,omitempty"`
	Note   string `json:"note,omitempty"`
}

type CreateTransactionRequest struct {
	WalletID       string                    `json:"wallet_id"`
	CounterpartyID string                    `json:"counterparty_id,omitempty"`
	Counterparty   string                    `json:"counterparty"`
	Type           int                       `json:"type"`
	Amount         decimal.Decimal           `json:"amount"`
	Description    string                    `json:"description,omitempty"`
	Category       string                    `json:"category,omitempty"`
	Metadata       map[string]interface{}    `json:"metadata,omitempty"`
	Programmable   map[string]interface{}    `json:"programmable,omitempty"`
}

type CreatePolicyRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"`
	Condition   map[string]interface{} `json:"condition"`
	Action      string                 `json:"action"`
	Priority    int                    `json:"priority"`
	Scope       string                 `json:"scope,omitempty"`
	TargetIDs   string                 `json:"target_ids,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

type TransactionListFilters struct {
	WalletID  string     `json:"wallet_id,omitempty"`
	Status    *int       `json:"status,omitempty"`
	Type      *int       `json:"type,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

type ComplianceCheckRequest struct {
	WalletID string          `json:"wallet_id"`
	Amount   decimal.Decimal `json:"amount"`
	Category string          `json:"category,omitempty"`
}

type ComplianceCheckResult struct {
	Passed bool          `json:"passed"`
	Checks []CheckResult `json:"checks"`
}

type CheckResult struct {
	CheckName string `json:"check_name"`
	Passed    bool   `json:"passed"`
	Message   string `json:"message"`
}

type TierLimits struct {
	MaxTransaction decimal.Decimal
	DailyLimit     decimal.Decimal
}

// Helper functions

func generateWalletNumber() string {
	return fmt.Sprintf("CBDC-%s", uuid.New().String()[:12])
}

func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	data, _ := json.Marshal(v)
	return string(data)
}

func cloneWallet(w *domain.Wallet) *domain.Wallet {
	clone := *w
	return &clone
}
