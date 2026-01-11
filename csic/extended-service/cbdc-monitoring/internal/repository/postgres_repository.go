package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"csic-platform/extended-services/cbdc-monitoring/internal/domain"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrPolicyNotFound       = errors.New("policy rule not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrWalletFrozen         = errors.New("wallet is frozen")
	ErrWalletSuspended      = errors.New("wallet is suspended")
	ErrWalletInactive       = errors.New("wallet is not active")
	ErrDuplicateWallet      = errors.New("wallet number already exists")
)

// WalletRepository handles wallet data operations
type WalletRepository struct {
	db          *gorm.DB
	tableName   string
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{
		db:        db,
		tableName: "wallets",
	}
}

// Create creates a new wallet
func (r *WalletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	result := r.db.WithContext(ctx).Table("wallets").Create(wallet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrDuplicateWallet
		}
		return fmt.Errorf("failed to create wallet: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a wallet by ID
func (r *WalletRepository) GetByID(ctx context.Context, id string) (*domain.Wallet, error) {
	var wallet domain.Wallet
	result := r.db.WithContext(ctx).Table("wallets").Where("id = ?", id).First(&wallet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", result.Error)
	}
	return &wallet, nil
}

// GetByWalletNumber retrieves a wallet by wallet number
func (r *WalletRepository) GetByWalletNumber(ctx context.Context, number string) (*domain.Wallet, error) {
	var wallet domain.Wallet
	result := r.db.WithContext(ctx).Table("wallets").Where("wallet_number = ?", number).First(&wallet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", result.Error)
	}
	return &wallet, nil
}

// Update updates a wallet
func (r *WalletRepository) Update(ctx context.Context, wallet *domain.Wallet) error {
	result := r.db.WithContext(ctx).Table("wallets").Save(wallet)
	if result.Error != nil {
		return fmt.Errorf("failed to update wallet: %w", result.Error)
	}
	return nil
}

// UpdateBalance updates wallet balance atomically
func (r *WalletRepository) UpdateBalance(ctx context.Context, id string, amount decimal.Decimal, add bool) error {
	var query string
	var args []interface{}

	if add {
		query = "balance = balance + ?"
		args = []interface{}{amount}
	} else {
		query = "balance = balance - ?"
		args = []interface{}{amount}
	}

	result := r.db.WithContext(ctx).Table("wallets").
		Where("id = ?", id).
		Update(query, args...)

	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrWalletNotFound
	}
	return nil
}

// List retrieves wallets with pagination
func (r *WalletRepository) List(ctx context.Context, offset, limit int, status *domain.WalletStatus) ([]domain.Wallet, int64, error) {
	var wallets []domain.Wallet
	var total int64

	query := r.db.WithContext(ctx).Table("wallets")
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count wallets: %w", err)
	}

	// Get paginated results
	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&wallets)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list wallets: %w", result.Error)
	}

	return wallets, total, nil
}

// TransactionRepository handles transaction data operations
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	result := r.db.WithContext(ctx).Table("transactions").Create(tx)
	if result.Error != nil {
		return fmt.Errorf("failed to create transaction: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	var tx domain.Transaction
	result := r.db.WithContext(ctx).Table("transactions").Where("id = ?", id).First(&tx)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", result.Error)
	}
	return &tx, nil
}

// Update updates a transaction
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	result := r.db.WithContext(ctx).Table("transactions").Save(tx)
	if result.Error != nil {
		return fmt.Errorf("failed to update transaction: %w", result.Error)
	}
	return nil
}

// UpdateStatus updates transaction status atomically
func (r *TransactionRepository) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	result := r.db.WithContext(ctx).Table("transactions").
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

// List retrieves transactions with filters
func (r *TransactionRepository) List(ctx context.Context, filters TransactionFilters) ([]domain.Transaction, int64, error) {
	var transactions []domain.Transaction
	var total int64

	query := r.buildQuery(filters)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get paginated results
	offset := (filters.Page - 1) * filters.Limit
	result := query.Offset(offset).Limit(filters.Limit).
		Order("created_at DESC").
		Find(&transactions)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", result.Error)
	}

	return transactions, total, nil
}

// GetByWalletID retrieves transactions for a wallet
func (r *TransactionRepository) GetByWalletID(ctx context.Context, walletID string, limit int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	result := r.db.WithContext(ctx).Table("transactions").
		Where("wallet_id = ?", walletID).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get wallet transactions: %w", result.Error)
	}
	return transactions, nil
}

// GetTodayStats retrieves today's transaction statistics
func (r *TransactionRepository) GetTodayStats(ctx context.Context) (approved, rejected int64, volume decimal.Decimal, err error) {
	today := time.Now().Format("2006-01-02")

	// Approved count
	if err := r.db.WithContext(ctx).Table("transactions").
		Where("DATE(created_at) = ? AND status = ?", today, domain.TxStatusApproved).
		Count(&approved).Error; err != nil {
		return 0, 0, decimal.Zero, fmt.Errorf("failed to count approved: %w", err)
	}

	// Rejected count
	if err := r.db.WithContext(ctx).Table("transactions").
		Where("DATE(created_at) = ? AND status = ?", today, domain.TxStatusRejected).
		Count(&rejected).Error; err != nil {
		return 0, 0, decimal.Zero, fmt.Errorf("failed to count rejected: %w", err)
	}

	// Volume
	if err := r.db.WithContext(ctx).Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total").
		Where("DATE(created_at) = ? AND status = ?", today, domain.TxStatusApproved).
		Scan(&volume).Error; err != nil {
		return 0, 0, decimal.Zero, fmt.Errorf("failed to get volume: %w", err)
	}

	return approved, rejected, volume, nil
}

// PolicyRepository handles policy rule data operations
type PolicyRepository struct {
	db *gorm.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *gorm.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// Create creates a new policy rule
func (r *PolicyRepository) Create(ctx context.Context, policy *domain.PolicyRule) error {
	result := r.db.WithContext(ctx).Table("policy_rules").Create(policy)
	if result.Error != nil {
		return fmt.Errorf("failed to create policy: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a policy rule by ID
func (r *PolicyRepository) GetByID(ctx context.Context, id string) (*domain.PolicyRule, error) {
	var policy domain.PolicyRule
	result := r.db.WithContext(ctx).Table("policy_rules").Where("id = ?", id).First(&policy)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPolicyNotFound
		}
		return nil, fmt.Errorf("failed to get policy: %w", result.Error)
	}
	return &policy, nil
}

// Update updates a policy rule
func (r *PolicyRepository) Update(ctx context.Context, policy *domain.PolicyRule) error {
	result := r.db.WithContext(ctx).Table("policy_rules").Save(policy)
	if result.Error != nil {
		return fmt.Errorf("failed to update policy: %w", result.Error)
	}
	return nil
}

// Delete soft deletes a policy rule
func (r *PolicyRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Table("policy_rules").
		Where("id = ?", id).
		Update("enabled", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete policy: %w", result.Error)
	}
	return nil
}

// List retrieves all enabled policy rules
func (r *PolicyRepository) List(ctx context.Context, ruleType string) ([]domain.PolicyRule, error) {
	var policies []domain.PolicyRule

	query := r.db.WithContext(ctx).Table("policy_rules").
		Where("enabled = ?", true).
		Order("priority DESC")

	if ruleType != "" {
		query = query.Where("rule_type = ?", ruleType)
	}

	result := query.Find(&policies)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list policies: %w", result.Error)
	}

	return policies, nil
}

// ViolationRepository handles violation record operations
type ViolationRepository struct {
	db *gorm.DB
}

// NewViolationRepository creates a new violation repository
func NewViolationRepository(db *gorm.DB) *ViolationRepository {
	return &ViolationRepository{db: db}
}

// Create creates a new violation record
func (r *ViolationRepository) Create(ctx context.Context, violation *domain.ViolationRecord) error {
	result := r.db.WithContext(ctx).Table("violation_records").Create(violation)
	if result.Error != nil {
		return fmt.Errorf("failed to create violation: %w", result.Error)
	}
	return nil
}

// List retrieves violations with filters
func (r *ViolationRepository) List(ctx context.Context, walletID string, unresolvedOnly bool, limit int) ([]domain.ViolationRecord, error) {
	var violations []domain.ViolationRecord

	query := r.db.WithContext(ctx).Table("violation_records").
		Order("created_at DESC")

	if walletID != "" {
		query = query.Where("wallet_id = ?", walletID)
	}
	if unresolvedOnly {
		query = query.Where("resolved = ?", false)
	}

	result := query.Limit(limit).Find(&violations)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list violations: %w", result.Error)
	}

	return violations, nil
}

// GetTodayCount retrieves today's violation count
func (r *ViolationRepository) GetTodayCount(ctx context.Context) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")

	if err := r.db.WithContext(ctx).Table("violation_records").
		Where("DATE(created_at) = ?", today).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count violations: %w", err)
	}

	return count, nil
}

// Helper functions
func (r *TransactionRepository) buildQuery(filters TransactionFilters) *gorm.DB {
	query := r.db.WithContext(ctx).Table("transactions")

	if filters.WalletID != "" {
		query = query.Where("wallet_id = ?", filters.WalletID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.Type != nil {
		query = query.Where("transaction_type = ?", *filters.Type)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	return query
}

// TransactionFilters holds filtering options for transactions
type TransactionFilters struct {
	WalletID  string
	Status    *domain.TransactionStatus
	Type      *domain.TransactionType
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
}

// AuditRepository handles audit log operations
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	result := r.db.WithContext(ctx).Table("audit_logs").Create(log)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}
	return nil
}

// GetByEntity retrieves audit logs for a specific entity
func (r *AuditRepository) GetByEntity(ctx context.Context, entityType, entityID string, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	result := r.db.WithContext(ctx).Table("audit_logs").
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", result.Error)
	}
	return logs, nil
}
