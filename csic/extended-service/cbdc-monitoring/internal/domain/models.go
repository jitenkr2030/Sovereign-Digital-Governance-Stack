package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// WalletTier represents the KYC verification level
type WalletTier int

const (
	Tier1Basic WalletTier = iota + 1
	Tier2Standard
	Tier3Premium
)

func (t WalletTier) String() string {
	switch t {
	case Tier1Basic:
		return "Basic"
	case Tier2Standard:
		return "Standard"
	case Tier3Premium:
		return "Premium"
	default:
		return "Unknown"
	}
}

// WalletStatus represents the current state of a wallet
type WalletStatus int

const (
	WalletStatusPending WalletStatus = iota
	WalletStatusActive
	WalletStatusSuspended
	WalletStatusFrozen
	WalletStatusClosed
)

func (s WalletStatus) String() string {
	switch s {
	case WalletStatusPending:
		return "Pending"
	case WalletStatusActive:
		return "Active"
	case WalletStatusSuspended:
		return "Suspended"
	case WalletStatusFrozen:
		return "Frozen"
	case WalletStatusClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// TransactionStatus represents the state of a transaction
type TransactionStatus int

const (
	TxStatusPending TransactionStatus = iota
	TxStatusProcessing
	TxStatusApproved
	TxStatusRejected
	TxStatusCompleted
	TxStatusFailed
	TxStatusCancelled
)

func (s TransactionStatus) String() string {
	switch s {
	case TxStatusPending:
		return "Pending"
	case TxStatusProcessing:
		return "Processing"
	case TxStatusApproved:
		return "Approved"
	case TxStatusRejected:
		return "Rejected"
	case TxStatusCompleted:
		return "Completed"
	case TxStatusFailed:
		return "Failed"
	case TxStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// TransactionType represents the type of CBDC transaction
type TransactionType int

const (
	TxTypeTransfer TransactionType = iota
	TxTypePayment
	TxTypeWithdrawal
	TxTypeDeposit
	TxTypeEscrow
	TxTypeAtomicSwap
)

func (t TransactionType) String() string {
	switch t {
	case TxTypeTransfer:
		return "Transfer"
	case TxTypePayment:
		return "Payment"
	case TxTypeWithdrawal:
		return "Withdrawal"
	case TxTypeDeposit:
		return "Deposit"
	case TxTypeEscrow:
		return "Escrow"
	case TxTypeAtomicSwap:
		return "AtomicSwap"
	default:
		return "Unknown"
	}
}

// Wallet represents a CBDC wallet holder
type Wallet struct {
	ID               string          `json:"id" gorm:"primaryKey;size:64"`
	OwnerID          string          `json:"owner_id" gorm:"index;size:64;not null"`
	WalletNumber     string          `json:"wallet_number" gorm:"uniqueIndex;size:32;not null"`
	Tier             WalletTier      `json:"tier" gorm:"not null;default:1"`
	Status           WalletStatus    `json:"status" gorm:"not null;default:0"`
	Balance          decimal.Decimal `json:"balance" gorm:"type:decimal(20,8);default:0"`
	FrozenBalance    decimal.Decimal `json:"frozen_balance" gorm:"type:decimal(20,8);default:0"`
	Currency         string          `json:"currency" gorm:"size:10;default:'CBDC'"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	KYCVerifiedAt    *time.Time      `json:"kyc_verified_at,omitempty"`
	SuspendedAt      *time.Time      `json:"suspended_at,omitempty"`
	Metadata         string          `json:"metadata,omitempty" gorm:"type:jsonb"`
}

// Transaction represents a CBDC transaction
type Transaction struct {
	ID              string            `json:"id" gorm:"primaryKey;size:64"`
	WalletID        string            `json:"wallet_id" gorm:"index;size:64;not null"`
	CounterpartyID  string            `json:"counterparty_id,omitempty" gorm:"index;size:64"`
	Counterparty    string            `json:"counterparty" gorm:"size:64"`
	TransactionType TransactionType   `json:"transaction_type" gorm:"not null"`
	Status          TransactionStatus `json:"status" gorm:"not null;default:0"`
	Amount          decimal.Decimal   `json:"amount" gorm:"type:decimal(20,8);not null"`
	Currency        string            `json:"currency" gorm:"size:10;default:'CBDC'"`
	Fee             decimal.Decimal   `json:"fee" gorm:"type:decimal(20,8);default:0"`
	NetAmount       decimal.Decimal   `json:"net_amount" gorm:"type:decimal(20,8)"`
	Description     string            `json:"description,omitempty" gorm:"size:500"`
	Metadata        string            `json:"metadata,omitempty" gorm:"type:jsonb"`
	Programmable    string            `json:"programmable,omitempty" gorm:"type:jsonb"`
	PolicyRuleIDs   string            `json:"policy_rule_ids,omitempty" gorm:"size:255"`
	PolicyViolated  bool              `json:"policy_violated" gorm:"default:false"`
	ViolationReason string            `json:"violation_reason,omitempty" gorm:"size:255"`
	ApprovedBy      string            `json:"approved_by,omitempty" gorm:"size:64"`
	ApprovedAt      *time.Time        `json:"approved_at,omitempty"`
	ProcessedAt     *time.Time        `json:"processed_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// PolicyRule represents a programmable money policy rule
type PolicyRule struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	RuleType    string    `json:"rule_type" gorm:"size:50;not null"` // velocity, limit, geo, category, etc.
	Condition   string    `json:"condition" gorm:"type:jsonb;not null"`
	Action      string    `json:"action" gorm:"size:50;not null"` // block, flag, limit, require_approval
	Priority    int       `json:"priority" gorm:"default:0"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	Scope       string    `json:"scope" gorm:"size:50"` // global, tier_specific, wallet_specific
	TargetIDs   string    `json:"target_ids,omitempty" gorm:"size:255"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ViolationRecord tracks policy violations
type ViolationRecord struct {
	ID             string    `json:"id" gorm:"primaryKey;size:64"`
	TransactionID  string    `json:"transaction_id" gorm:"index;size:64;not null"`
	WalletID       string    `json:"wallet_id" gorm:"index;size:64;not null"`
	PolicyRuleID   string    `json:"policy_rule_id" gorm:"size:64"`
	RuleName       string    `json:"rule_name" gorm:"size:100"`
	ViolationType  string    `json:"violation_type" gorm:"size:50"`
	Details        string    `json:"details" gorm:"type:jsonb"`
	Resolved       bool      `json:"resolved" gorm:"default:false"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy     string    `json:"resolved_by,omitempty" gorm:"size:64"`
	ResolutionNote string    `json:"resolution_note,omitempty" gorm:"size:500"`
	CreatedAt      time.Time `json:"created_at"`
}

// AuditLog records all wallet and transaction activities
type AuditLog struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	EntityType  string    `json:"entity_type" gorm:"size:50;not null"` // wallet, transaction, policy
	EntityID    string    `json:"entity_id" gorm:"size:64;not null"`
	Action      string    `json:"action" gorm:"size:50;not null"` // create, update, delete, approve, reject
	ActorID     string    `json:"actor_id" gorm:"size:64"` // user, system, admin
	ActorType   string    `json:"actor_type" gorm:"size:50"` // user, system, api
	OldValue    string    `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue    string    `json:"new_value,omitempty" gorm:"type:jsonb"`
	IPAddress   string    `json:"ip_address" gorm:"size:45"`
	UserAgent   string    `json:"user_agent" gorm:"size:255"`
	Location    string    `json:"location" gorm:"size:100"`
	Timestamp   time.Time `json:"timestamp" gorm:"index"`
}

// VelocityCheck tracks transaction velocity for fraud detection
type VelocityCheck struct {
	ID              string    `json:"id" gorm:"primaryKey;size:64"`
	WalletID        string    `json:"wallet_id" gorm:"index;size:64;not null"`
	TimeWindow      string    `json:"time_window" gorm:"size:20"` // 1h, 24h, 7d
	TransactionCount int      `json:"transaction_count" gorm:"default:0"`
	TotalAmount     decimal.Decimal `json:"total_amount" gorm:"type:decimal(20,8);default:0"`
	MaxAllowed      int      `json:"max_allowed" gorm:"default:0"`
	MaxAmount       decimal.Decimal `json:"max_amount" gorm:"type:decimal(20,8);default:0"`
	Status          string    `json:"status" gorm:"size:20"` // normal, warning, exceeded
	LastUpdated     time.Time `json:"last_updated"`
	CreatedAt       time.Time `json:"created_at"`
}

// MonitoringStats holds aggregated monitoring statistics
type MonitoringStats struct {
	TotalWallets       int64           `json:"total_wallets"`
	ActiveWallets      int64           `json:"active_wallets"`
	TotalTransactions  int64           `json:"total_transactions"`
	PendingTransactions int64          `json:"pending_transactions"`
	ApprovedToday      int64           `json:"approved_today"`
	RejectedToday      int64           `json:"rejected_today"`
	TotalVolume        decimal.Decimal `json:"total_volume"`
	VolumeToday        decimal.Decimal `json:"volume_today"`
	ViolationCount     int64           `json:"violation_count"`
	ViolationsToday    int64           `json:"violations_today"`
}
