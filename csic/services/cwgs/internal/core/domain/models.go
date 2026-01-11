package domain

import (
	"time"
)

// Custodian represents a registered cryptocurrency custodian
type Custodian struct {
	ID             string    `json:"id" db:"id"`
	CustodianID    string    `json:"custodian_id" db:"custodian_id"`
	LegalName      string    `json:"legal_name" db:"legal_name"`
	LicenseNumber  string    `json:"license_number" db:"license_number"`
	Jurisdiction   string    `json:"jurisdiction" db:"jurisdiction"`
	RegulatoryBody string    `json:"regulatory_body" db:"regulatory_body"`
	KYCLevel       string    `json:"kyc_level" db:"kyc_level"`
	Status         string    `json:"status" db:"status"`
	LicenseExpires *time.Time `json:"license_expires,omitempty" db:"license_expires"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// CustodianStatus represents the status of a custodian
type CustodianStatus string

const (
	CustodianStatusActive     CustodianStatus = "ACTIVE"
	CustodianStatusSuspended  CustodianStatus = "SUSPENDED"
	CustodianStatusRevoked    CustodianStatus = "REVOKED"
	CustodianStatusPending    CustodianStatus = "PENDING"
)

// Wallet represents a custodial wallet
type Wallet struct {
	ID            string    `json:"id" db:"id"`
	Address       string    `json:"address" db:"address"`
	Chain         string    `json:"chain" db:"chain"`
	CustodianID   *string   `json:"custodian_id,omitempty" db:"custodian_id"`
	WalletType    string    `json:"wallet_type" db:"wallet_type"`
	AssetType     string    `json:"asset_type" db:"asset_type"`
	Status        string    `json:"status" db:"status"`
	KYCLevel      string    `json:"kyc_level" db:"kyc_level"`
	RiskLevel     string    `json:"risk_level" db:"risk_level"`
	DailyLimit    float64   `json:"daily_limit" db:"daily_limit"`
	Balance       float64   `json:"balance" db:"balance"`
	LastActivity  *time.Time `json:"last_activity,omitempty" db:"last_activity"`
	BlacklistedAt *time.Time `json:"blacklisted_at,omitempty" db:"blacklisted_at"`
	BlacklistReason string   `json:"blacklist_reason,omitempty" db:"blacklist_reason"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// WalletStatus represents the status of a wallet
type WalletStatus string

const (
	WalletStatusPending    WalletStatus = "PENDING"
	WalletStatusActive     WalletStatus = "ACTIVE"
	WalletStatusFrozen     WalletStatus = "FROZEN"
	WalletStatusBlacklisted WalletStatus = "BLACKLISTED"
	WalletStatusRevoked    WalletStatus = "REVOKED"
)

// WhitelistEntry represents a whitelisted wallet
type WhitelistEntry struct {
	ID          string     `json:"id" db:"id"`
	WalletID    string     `json:"wallet_id" db:"wallet_id"`
	WhitelistType string   `json:"whitelist_type" db:"whitelist_type"`
	Reason      string     `json:"reason" db:"reason"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	ApprovedBy  string     `json:"approved_by" db:"approved_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// WhitelistType represents the type of whitelist
type WhitelistType string

const (
	WhitelistTypeTrusted    WhitelistType = "TRUSTED"
	WhitelistTypeInstitutional WhitelistType = "INSTITUTIONAL"
	WhitelistTypePartner   WhitelistType = "PARTNER"
	WhitelistTypeRegulated WhitelistType = "REGULATED"
)

// BlacklistEntry represents a blacklisted wallet
type BlacklistEntry struct {
	ID          string     `json:"id" db:"id"`
	WalletID    string     `json:"wallet_id" db:"wallet_id"`
	Address     string     `json:"address" db:"address"`
	Chain       string     `json:"chain" db:"chain"`
	Reason      string     `json:"reason" db:"reason"`
	Severity    string     `json:"severity" db:"severity"`
	Source      string     `json:"source" db:"source"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	BlockedAt   time.Time  `json:"blocked_at" db:"blocked_at"`
}

// BlacklistSeverity represents the severity of a blacklist entry
type BlacklistSeverity string

const (
	BlacklistSeverityLow      BlacklistSeverity = "LOW"
	BlacklistSeverityMedium   BlacklistSeverity = "MEDIUM"
	BlacklistSeverityHigh     BlacklistSeverity = "HIGH"
	BlacklistSeverityCritical BlacklistSeverity = "CRITICAL"
)

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	Address       string   `json:"address"`
	Chain         string   `json:"chain"`
	IsWhitelisted bool     `json:"is_whitelisted"`
	IsBlacklisted bool     `json:"is_blacklisted"`
	RiskLevel     string   `json:"risk_level"`
	KYCStatus     string   `json:"kyc_status"`
	Flags         []string `json:"flags"`
	CheckedAt     time.Time `json:"checked_at"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ReportID        string            `json:"report_id"`
	PeriodStart     time.Time         `json:"period_start"`
	PeriodEnd       time.Time         `json:"period_end"`
	TotalWallets    int64             `json:"total_wallets"`
	ActiveWallets   int64             `json:"active_wallets"`
	FrozenWallets   int64             `json:"frozen_wallets"`
	BlacklistedWallets int64          `json:"blacklisted_wallets"`
	WhitelistedWallets int64         `json:"whitelisted_wallets"`
	ComplianceRate  float64           `json:"compliance_rate"`
	ByChain         map[string]int64  `json:"by_chain"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

// WalletRegistrationRequest represents a request to register a new wallet
type WalletRegistrationRequest struct {
	Address    string  `json:"address"`
	Chain      string  `json:"chain"`
	CustodianID string `json:"custodian_id,omitempty"`
	WalletType string  `json:"wallet_type"`
	AssetType  string  `json:"asset_type"`
	KYCLevel   string  `json:"kyc_level,omitempty"`
	DailyLimit float64 `json:"daily_limit,omitempty"`
}

// WalletStatusUpdateRequest represents a request to update wallet status
type WalletStatusUpdateRequest struct {
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
	UpdatedBy string `json:"updated_by"`
}

// WhitelistRequest represents a request to add a wallet to whitelist
type WhitelistRequest struct {
	Address       string `json:"address"`
	Chain         string `json:"chain"`
	WhitelistType string `json:"whitelist_type"`
	Reason        string `json:"reason"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	ApprovedBy    string `json:"approved_by"`
}

// BlacklistRequest represents a request to blacklist a wallet
type BlacklistRequest struct {
	Address  string `json:"address"`
	Chain    string `json:"chain"`
	Reason   string `json:"reason"`
	Severity string `json:"severity"`
	Source   string `json:"source,omitempty"`
}
