package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WalletType represents the type of wallet
type WalletType string

const (
	WalletTypeCustodial      WalletType = "CUSTODIAL"
	WalletTypeMultiSig       WalletType = "MULTI_SIG"
	WalletTypeExchangeHot    WalletType = "EXCHANGE_HOT"
	WalletTypeExchangeCold   WalletType = "EXCHANGE_COLD"
	WalletTypeTreasury       WalletType = "TREASURY"
	WalletTypeEscrow         WalletType = "ESCROW"
	WalletTypeUserControlled WalletType = "USER_CONTROLLED"
)

// WalletStatus represents the status of a wallet
type WalletStatus string

const (
	WalletStatusActive      WalletStatus = "ACTIVE"
	WalletStatusSuspended   WalletStatus = "SUSPENDED"
	WalletStatusFrozen      WalletStatus = "FROZEN"
	WalletStatusRevoked     WalletStatus = "REVOKED"
	WalletStatusPending     WalletStatus = "PENDING"
	WalletStatusUnderReview WalletStatus = "UNDER_REVIEW"
)

// BlockchainType represents the blockchain network
type BlockchainType string

const (
	BlockchainBitcoin  BlockchainType = "BITCOIN"
	BlockchainEthereum BlockchainType = "ETHEREUM"
	BlockchainERC20    BlockchainType = "ERC20"
	BlockchainTRC20    BlockchainType = "TRC20"
	BlockchainBEP20    BlockchainType = "BEP20"
	BlockchainPolygon  BlockchainType = "POLYGON"
	BlockchainSolana   BlockchainType = "SOLANA"
	BlockchainLitecoin BlockchainType = "LITECOIN"
	BlockchainDash     BlockchainType = "DASH"
	BlockchainMonero   BlockchainType = "MONERO"
)

// SignerType represents the type of wallet signer
type SignerType string

const (
	SignerTypeInstitution SignerType = "INSTITUTION"
	SignerTypeIndividual  SignerType = "INDIVIDUAL"
	SignerTypeAutomated   SignerType = "AUTOMATED"
	SignerTypeEmergency   SignerType = "EMERGENCY"
)

// SignerStatus represents the status of a signer
type SignerStatus string

const (
	SignerStatusActive   SignerStatus = "ACTIVE"
	SignerStatusInactive SignerStatus = "INACTIVE"
	SignerStatusRevoked  SignerStatus = "REVOKED"
	SignerStatusPending  SignerStatus = "PENDING"
)

// TransactionStatus represents the status of a proposed transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusApproved  TransactionStatus = "APPROVED"
	TransactionStatusRejected  TransactionStatus = "REJECTED"
	TransactionStatusExecuted  TransactionStatus = "EXECUTED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusExpired   TransactionStatus = "EXPIRED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

// SignatureStatus represents the status of a signature request
type SignatureStatus string

const (
	SignatureStatusPending    SignatureStatus = "PENDING"
	SignatureStatusInProgress SignatureStatus = "IN_PROGRESS"
	SignatureStatusCompleted  SignatureStatus = "COMPLETED"
	SignatureStatusFailed     SignatureStatus = "FAILED"
	SignatureStatusCancelled  SignatureStatus = "CANCELLED"
)

// FreezeStatus represents the status of a wallet freeze
type FreezeStatus string

const (
	FreezeStatusActive   FreezeStatus = "ACTIVE"
	FreezeStatusPartial  FreezeStatus = "PARTIAL"
	FreezeStatusReleased FreezeStatus = "RELEASED"
	FreezeStatusExpired  FreezeStatus = "EXPIRED"
)

// FreezeReason represents the reason for a wallet freeze
type FreezeReason string

const (
	FreezeReasonLegalOrder     FreezeReason = "LEGAL_ORDER"
	FreezeReasonRegulatory     FreezeReason = "REGULATORY"
	FreezeReasonSuspicious     FreezeReason = "SUSPICIOUS_ACTIVITY"
	FreezeReasonExchangeHack   FreezeReason = "EXCHANGE_HACK"
	FreezeReasonUserRequest    FreezeReason = "USER_REQUEST"
	FreezeReasonMaintenance    FreezeReason = "MAINTENANCE"
)

// BlacklistStatus represents the status of a blacklist entry
type BlacklistStatus string

const (
	BlacklistStatusActive   BlacklistStatus = "ACTIVE"
	BlacklistStatusExpired  BlacklistStatus = "EXPIRED"
	BlacklistStatusRemoved  BlacklistStatus = "REMOVED"
	BlacklistStatusUnderReview BlacklistStatus = "UNDER_REVIEW"
)

// ComplianceStatus represents the compliance status of a wallet
type ComplianceStatus string

const (
	ComplianceStatusCompliant         ComplianceStatus = "COMPLIANT"
	ComplianceStatusNonCompliant      ComplianceStatus = "NON_COMPLIANT"
	ComplianceStatusUnderReview       ComplianceStatus = "UNDER_REVIEW"
	ComplianceStatusPendingApproval   ComplianceStatus = "PENDING_APPROVAL"
	ComplianceStatusSuspended         ComplianceStatus = "SUSPENDED"
)

// Wallet represents a registered wallet
type Wallet struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	WalletID        string         `json:"wallet_id" db:"wallet_id"`
	Type            WalletType     `json:"type" db:"type"`
	Status          WalletStatus   `json:"status" db:"status"`
	Blockchain      BlockchainType `json:"blockchain" db:"blockchain"`
	Address         string         `json:"address" db:"address"`
	AddressChecksum string         `json:"address_checksum" db:"address_checksum"`
	Label           string         `json:"label" db:"label"`
	Description     string         `json:"description" db:"description"`
	ExchangeID      *uuid.UUID     `json:"exchange_id,omitempty" db:"exchange_id"`
	ExchangeName    string         `json:"exchange_name" db:"exchange_name"`
	OwnerEntityID   uuid.UUID      `json:"owner_entity_id" db:"owner_entity_id"`
	OwnerEntityName string         `json:"owner_entity_name" db:"owner_entity_name"`
	SignersRequired int            `json:"signers_required" db:"signers_required"`
	SignersTotal    int            `json:"signers_total" db:"signers_total"`
	Threshold       int            `json:"threshold" db:"threshold"`
	TotalBalance    decimal.Decimal `json:"total_balance" db:"total_balance"`
	BalanceCurrency string         `json:"balance_currency" db:"balance_currency"`
	LastActivityAt  *time.Time     `json:"last_activity_at,omitempty" db:"last_activity_at"`
	ComplianceScore decimal.Decimal `json:"compliance_score" db:"compliance_score"`
	IsWhitelisted   bool           `json:"is_whitelisted" db:"is_whitelisted"`
	IsBlacklisted   bool           `json:"is_blacklisted" db:"is_blacklisted"`
	Metadata        JSONMap        `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	RevokedAt       *time.Time     `json:"revoked_at,omitempty" db:"revoked_at"`
}

// WalletSigner represents a signer for a multi-signature wallet
type WalletSigner struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	WalletID      uuid.UUID    `json:"wallet_id" db:"wallet_id"`
	SignerID      string       `json:"signer_id" db:"signer_id"`
	SignerType    SignerType   `json:"signer_type" db:"signer_type"`
	SignerName    string       `json:"signer_name" db:"signer_name"`
	PublicKey     string       `json:"public_key" db:"public_key"`
	PublicKeyHash string       `json:"public_key_hash" db:"public_key_hash"`
	Status        SignerStatus `json:"status" db:"status"`
	Order         int          `json:"order" db:"order"`
	IsEmergency   bool         `json:"is_emergency" db:"is_emergency"`
	Weight        int          `json:"weight" db:"weight"`
	LastSignedAt  *time.Time   `json:"last_signed_at,omitempty" db:"last_signed_at"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

// TransactionProposal represents a proposed transaction for multi-sig approval
type TransactionProposal struct {
	ID                 uuid.UUID           `json:"id" db:"id"`
	WalletID           uuid.UUID           `json:"wallet_id" db:"wallet_id"`
	TransactionID      string              `json:"transaction_id" db:"transaction_id"`
	Type               string              `json:"type" db:"type"`
	Blockchain         BlockchainType      `json:"blockchain" db:"blockchain"`
	ToAddress          string              `json:"to_address" db:"to_address"`
	Amount             decimal.Decimal     `json:"amount" db:"amount"`
	AssetSymbol        string              `json:"asset_symbol" db:"asset_symbol"`
	ContractAddress    string              `json:"contract_address,omitempty" db:"contract_address"`
	GasLimit           int                 `json:"gas_limit" db:"gas_limit"`
	GasPrice           decimal.Decimal     `json:"gas_price" db:"gas_price"`
	Nonce              int                 `json:"nonce" db:"nonce"`
	RawTransaction     string              `json:"raw_transaction" db:"raw_transaction"`
	SignedTransactions []string            `json:"signed_transactions,omitempty" db:"signed_transactions"`
	Status             TransactionStatus   `json:"status" db:"status"`
	ProposerID         uuid.UUID           `json:"proposer_id" db:"proposer_id"`
	ProposerName       string              `json:"proposer_name" db:"proposer_name"`
	ApproversRequired  int                 `json:"approvers_required" db:"approvers_required"`
	ApprovalsCount     int                 `json:"approvals_count" db:"approvals_count"`
	RejectionsCount    int                 `json:"rejections_count" db:"rejections_count"`
	ExpiresAt          time.Time           `json:"expires_at" db:"expires_at"`
	ExecutedAt         *time.Time          `json:"executed_at,omitempty" db:"executed_at"`
	FailureReason      string              `json:"failure_reason,omitempty" db:"failure_reason"`
	Metadata           JSONMap             `json:"metadata,omitempty" db:"metadata"`
	CreatedAt          time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at" db:"updated_at"`
}

// TransactionApproval represents an approval or rejection for a transaction
type TransactionApproval struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	TransactionID   uuid.UUID    `json:"transaction_id" db:"transaction_id"`
	SignerID        uuid.UUID    `json:"signer_id" db:"signer_id"`
	SignerName      string       `json:"signer_name" db:"signer_name"`
	Decision        string       `json:"decision" db:"decision"` // APPROVED, REJECTED
	Reason          string       `json:"reason,omitempty" db:"reason"`
	Signature       string       `json:"signature,omitempty" db:"signature"`
	SignatureExpiry time.Time    `json:"signature_expiry" db:"signature_expiry"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
}

// SignatureRequest represents a request for a digital signature
type SignatureRequest struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	RequestID       string          `json:"request_id" db:"request_id"`
	WalletID        uuid.UUID       `json:"wallet_id" db:"wallet_id"`
	SignerID        uuid.UUID       `json:"signer_id" db:"signer_id"`
	MessageHash     string          `json:"message_hash" db:"message_hash"`
	Message         string          `json:"message" db:"message"`
	SignatureType   string          `json:"signature_type" db:"signature_type"`
	Status          SignatureStatus `json:"status" db:"status"`
	Signature       string          `json:"signature,omitempty" db:"signature"`
	PublicKey       string          `json:"public_key,omitempty" db:"public_key"`
	ExpiresAt       time.Time       `json:"expires_at" db:"expires_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	FailureReason   string          `json:"failure_reason,omitempty" db:"failure_reason"`
	RetryCount      int             `json:"retry_count" db:"retry_count"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// BlacklistEntry represents an entry in the blacklist
type BlacklistEntry struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	Address       string           `json:"address" db:"address"`
	AddressHash   string           `json:"address_hash" db:"address_hash"`
	Blockchain    BlockchainType   `json:"blockchain" db:"blockchain"`
	Reason        string           `json:"reason" db:"reason"`
	ReasonDetails string           `json:"reason_details" db:"reason_details"`
	Source        string           `json:"source" db:"source"` // "CHAINALYSIS", "LAW_ENFORCEMENT", "INTERNAL"
	RiskLevel     string           `json:"risk_level" db:"risk_level"`
	Status        BlacklistStatus  `json:"status" db:"status"`
	ExpiresAt     *time.Time       `json:"expires_at,omitempty" db:"expires_at"`
	AddedBy       uuid.UUID        `json:"added_by" db:"added_by"`
	AddedByName   string           `json:"added_by_name" db:"added_by_name"`
	ApprovedBy    *uuid.UUID       `json:"approved_by,omitempty" db:"approved_by"`
	RemovedBy     *uuid.UUID       `json:"removed_by,omitempty" db:"removed_by"`
	RemovalReason string           `json:"removal_reason,omitempty" db:"removal_reason"`
	Metadata      JSONMap          `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
}

// WhitelistEntry represents an entry in the whitelist
type WhitelistEntry struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	Address       string         `json:"address" db:"address"`
	AddressHash   string         `json:"address_hash" db:"address_hash"`
	Blockchain    BlockchainType `json:"blockchain" db:"blockchain"`
	Label         string         `json:"label" db:"label"`
	Description   string         `json:"description" db:"description"`
	EntityType    string         `json:"entity_type" db:"entity_type"` // "EXCHANGE", "INSTITUTION", "MERCHANT", "INDIVIDUAL"
	EntityID      *uuid.UUID     `json:"entity_id,omitempty" db:"entity_id"`
	EntityName    string         `json:"entity_name" db:"entity_name"`
	Status        string         `json:"status" db:"status"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty" db:"expires_at"`
	AddedBy       uuid.UUID      `json:"added_by" db:"added_by"`
	AddedByName   string         `json:"added_by_name" db:"added_by_name"`
	Metadata      JSONMap        `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

// WalletFreeze represents a wallet freeze record
type WalletFreeze struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	WalletID      uuid.UUID     `json:"wallet_id" db:"wallet_id"`
	WalletAddress string        `json:"wallet_address" db:"wallet_address"`
	Blockchain    BlockchainType `json:"blockchain" db:"blockchain"`
	Reason        FreezeReason  `json:"reason" db:"reason"`
	ReasonDetails string        `json:"reason_details" db:"reason_details"`
	Status        FreezeStatus  `json:"status" db:"status"`
	FreezeLevel   string        `json:"freeze_level" db:"freeze_level"` // "FULL", "INCOMING", "OUTGOING"
	LegalOrderID  string        `json:"legal_order_id,omitempty" db:"legal_order_id"`
	IssuedBy      uuid.UUID     `json:"issued_by" db:"issued_by"`
	IssuedByName  string        `json:"issued_by_name" db:"issued_by_name"`
	ApprovedBy    *uuid.UUID    `json:"approved_by,omitempty" db:"approved_by"`
	ExpiresAt     *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
	ReleasedAt    *time.Time    `json:"released_at,omitempty" db:"released_at"`
	ReleaseReason string        `json:"release_reason,omitempty" db:"release_reason"`
	Metadata      JSONMap       `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// AssetRecoveryRequest represents an asset recovery request
type AssetRecoveryRequest struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	RequestID       string       `json:"request_id" db:"request_id"`
	WalletID        uuid.UUID    `json:"wallet_id" db:"wallet_id"`
	SourceBlockchain BlockchainType `json:"source_blockchain" db:"source_blockchain"`
	SourceAddress   string       `json:"source_address" db:"source_address"`
	TargetAddress   string       `json:"target_address" db:"target_address"`
	AssetSymbol     string       `json:"asset_symbol" db:"asset_symbol"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	Reason          string       `json:"reason" db:"reason"`
	LegalCaseID     string       `json:"legal_case_id,omitempty" db:"legal_case_id"`
	Status          string       `json:"status" db:"status"` // "PENDING", "APPROVED", "EXECUTED", "REJECTED", "FAILED"
	RequesterID     uuid.UUID    `json:"requester_id" db:"requester_id"`
	RequesterName   string       `json:"requester_name" db:"requester_name"`
	ApproverID      *uuid.UUID   `json:"approver_id,omitempty" db:"approver_id"`
	TxHash          string       `json:"tx_hash,omitempty" db:"tx_hash"`
	FailureReason   string       `json:"failure_reason,omitempty" db:"failure_reason"`
	Metadata        JSONMap      `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
	ExecutedAt      *time.Time   `json:"executed_at,omitempty" db:"executed_at"`
}

// WalletAuditLog represents an audit log entry for wallet operations
type WalletAuditLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EntityType  string    `json:"entity_type" db:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id" db:"entity_id"`
	Action      string    `json:"action" db:"action"`
	ActorID     uuid.UUID `json:"actor_id" db:"actor_id"`
	ActorName   string    `json:"actor_name" db:"actor_name"`
	ActorType   string    `json:"actor_type" db:"actor_type"`
	OldValue    JSONMap   `json:"old_value,omitempty" db:"old_value"`
	NewValue    JSONMap   `json:"new_value,omitempty" db:"new_value"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	RequestID   string    `json:"request_id" db:"request_id"`
	Success     bool      `json:"success" db:"success"`
	ErrorMessage string   `json:"error_message,omitempty" db:"error_message"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	WalletID         uuid.UUID `json:"wallet_id"`
	IsBlacklisted    bool      `json:"is_blacklisted"`
	BlacklistReason  string    `json:"blacklist_reason,omitempty"`
	IsFrozen         bool      `json:"is_frozen"`
	FreezeReason     string    `json:"freeze_reason,omitempty"`
	IsWhitelisted    bool      `json:"is_whitelisted"`
	ComplianceScore  float64   `json:"compliance_score"`
	Status           string    `json:"status"`
	ChecksPerformed  []string  `json:"checks_performed"`
	RiskFactors      []string  `json:"risk_factors"`
	Recommendations  []string  `json:"recommendations"`
	CheckedAt        time.Time `json:"checked_at"`
}

// JSONMap represents a JSON object that can store arbitrary data
type JSONMap map[string]interface{}

// WalletSummary represents a summary of wallet statistics
type WalletSummary struct {
	TotalWallets         int            `json:"total_wallets"`
	ActiveWallets        int            `json:"active_wallets"`
	FrozenWallets        int            `json:"frozen_wallets"`
	SuspendedWallets     int            `json:"suspended_wallets"`
	ByType               map[string]int `json:"by_type"`
	ByBlockchain         map[string]int `json:"by_blockchain"`
	TotalBalance         decimal.Decimal `json:"total_balance"`
	BlacklistedAddresses int            `json:"blacklisted_addresses"`
	WhitelistedAddresses int            `json:"whitelisted_addresses"`
	ActiveFreezes        int            `json:"active_freezes"`
	PendingTransactions  int            `json:"pending_transactions"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// Filter structures
type WalletFilter struct {
	Types           []WalletType     `json:"types,omitempty"`
	Statuses        []WalletStatus   `json:"statuses,omitempty"`
	Blockchains     []BlockchainType `json:"blockchains,omitempty"`
	ExchangeIDs     []uuid.UUID      `json:"exchange_ids,omitempty"`
	OwnerEntityIDs  []uuid.UUID      `json:"owner_entity_ids,omitempty"`
	IsBlacklisted   *bool            `json:"is_blacklisted,omitempty"`
	IsWhitelisted   *bool            `json:"is_whitelisted,omitempty"`
	MinBalance      *decimal.Decimal `json:"min_balance,omitempty"`
	MaxBalance      *decimal.Decimal `json:"max_balance,omitempty"`
}

type BlacklistFilter struct {
	Blockchains   []BlockchainType `json:"blockchains,omitempty"`
	Statuses      []BlacklistStatus `json:"statuses,omitempty"`
	RiskLevels    []string         `json:"risk_levels,omitempty"`
	Sources       []string         `json:"sources,omitempty"`
	AddedAfter    *time.Time       `json:"added_after,omitempty"`
	AddedBefore   *time.Time       `json:"added_before,omitempty"`
}

type FreezeFilter struct {
	Statuses     []FreezeStatus `json:"statuses,omitempty"`
	Reasons      []FreezeReason `json:"reasons,omitempty"`
	FreezeLevels []string       `json:"freeze_levels,omitempty"`
	WalletIDs    []uuid.UUID    `json:"wallet_ids,omitempty"`
	IssuedBy     *uuid.UUID     `json:"issued_by,omitempty"`
	ActiveOnly   bool           `json:"active_only,omitempty"`
}
