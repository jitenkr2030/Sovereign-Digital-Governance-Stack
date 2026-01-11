package domain

import (
	"time"
)

// Transaction represents an on-chain transaction
type Transaction struct {
	ID             string                 `json:"id" db:"id"`
	TxHash         string                 `json:"tx_hash" db:"tx_hash"`
	Chain          string                 `json:"chain" db:"chain"`
	BlockNumber    *int64                 `json:"block_number,omitempty" db:"block_number"`
	FromAddress    string                 `json:"from_address" db:"from_address"`
	ToAddress      *string                `json:"to_address,omitempty" db:"to_address"`
	TokenAddress   *string                `json:"token_address,omitempty" db:"token_address"`
	Amount         float64                `json:"amount" db:"amount"`
	AmountUSD      float64                `json:"amount_usd" db:"amount_usd"`
	GasUsed        *int64                 `json:"gas_used,omitempty" db:"gas_used"`
	GasPrice       *float64               `json:"gas_price,omitempty" db:"gas_price"`
	GasFeeUSD      *float64               `json:"gas_fee_usd,omitempty" db:"gas_fee_usd"`
	Nonce          *int                   `json:"nonce,omitempty" db:"nonce"`
	TxTimestamp    time.Time              `json:"tx_timestamp" db:"tx_timestamp"`
	RiskScore      int                    `json:"risk_score" db:"risk_score"`
	RiskFactors    []RiskFactor           `json:"risk_factors"`
	Flagged        bool                   `json:"flagged" db:"flagged"`
	FlagReason     *string                `json:"flag_reason,omitempty" db:"flag_reason"`
	ReviewedAt     *time.Time             `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewedBy     *string                `json:"reviewed_by,omitempty" db:"reviewed_by"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// RiskFactor represents a specific risk indicator
type RiskFactor struct {
	Type       string  `json:"type"`
	Score      int     `json:"score"`
	Threshold  float64 `json:"threshold"`
	Observed   float64 `json:"observed"`
	Description string `json:"description"`
}

// SanctionedAddress represents a blacklisted address
type SanctionedAddress struct {
	ID          string     `json:"id" db:"id"`
	Address     string     `json:"address" db:"address"`
	Chain       string     `json:"chain" db:"chain"`
	SourceList  string     `json:"source_list" db:"source_list"`
	Reason      string     `json:"reason" db:"reason"`
	EntityName  string     `json:"entity_name" db:"entity_name"`
	EntityType  string     `json:"entity_type" db:"entity_type"`
	Program     string     `json:"program" db:"program"`
	AddedAt     time.Time  `json:"added_at" db:"added_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// WalletProfile represents a wallet's risk profile
type WalletProfile struct {
	ID              string                 `json:"id" db:"id"`
	Address         string                 `json:"address" db:"address"`
	Chain           string                 `json:"chain" db:"chain"`
	FirstSeen       *time.Time             `json:"first_seen,omitempty" db:"first_seen"`
	LastSeen        *time.Time             `json:"last_seen,omitempty" db:"last_seen"`
	TxCount         int                    `json:"tx_count" db:"tx_count"`
	TotalVolumeUSD  float64                `json:"total_volume_usd" db:"total_volume_usd"`
	AvgTxValueUSD   float64                `json:"avg_tx_value_usd" db:"avg_tx_value_usd"`
	ConnectedTags   []string               `json:"connected_tags" db:"connected_tags"`
	RiskIndicators  []RiskIndicator        `json:"risk_indicators"`
	WalletAgeHours  int                    `json:"wallet_age_hours" db:"wallet_age_hours"`
	IsContract      bool                   `json:"is_contract" db:"is_contract"`
	ContractType    *string                `json:"contract_type,omitempty" db:"contract_type"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// RiskIndicator represents a specific risk flag for a wallet
type RiskIndicator struct {
	Indicator    string  `json:"indicator"`
	Severity     string  `json:"severity"`
	Description  string  `json:"description"`
	FirstObserved time.Time `json:"first_observed"`
	LastObserved  time.Time `json:"last_observed"`
	Count        int     `json:"count"`
}

// RiskAssessment contains the complete risk analysis for a transaction
type RiskAssessment struct {
	OverallScore      int            `json:"overall_score"`
	RiskLevel         string         `json:"risk_level"`
	Factors           []RiskFactor   `json:"factors"`
	TotalRiskFromFactors int        `json:"total_risk_from_factors"`
	Recommendations   []string       `json:"recommendations"`
	RequiresReview    bool           `json:"requires_review"`
	Flagged           bool           `json:"flagged"`
}

// Alert represents a generated alert for suspicious activity
type Alert struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Severity    string          `json:"severity"`
	Transaction *Transaction    `json:"transaction,omitempty"`
	Wallet      *WalletProfile  `json:"wallet,omitempty"`
	Reason      string          `json:"reason"`
	Evidence    []EvidenceItem  `json:"evidence"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty"`
}

// EvidenceItem contains supporting evidence for an alert
type EvidenceItem struct {
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Threshold   interface{} `json:"threshold,omitempty"`
	Description string      `json:"description"`
}

// TransactionFilter represents filter criteria for querying transactions
type TransactionFilter struct {
	Chain        string     `form:"chain"`
	FromAddress  string     `form:"from_address"`
	ToAddress    string     `form:"to_address"`
	MinAmountUSD float64    `form:"min_amount_usd"`
	MaxAmountUSD float64    `form:"max_amount_usd"`
	Flagged      *bool      `form:"flagged"`
	MinRiskScore int        `form:"min_risk_score"`
	StartTime    *time.Time `form:"start_time"`
	EndTime      *time.Time `form:"end_time"`
	Page         int        `form:"page"`
	PageSize     int        `form:"page_size"`
}

// PaginationResult represents paginated query results
type PaginationResult struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// SanctionsListImport represents a batch import of sanctioned addresses
type SanctionsListImport struct {
	SourceList string               `json:"source_list"`
	Addresses  []SanctionedAddress  `json:"addresses"`
	ImportedBy string               `json:"imported_by"`
}

// RiskReport represents a risk summary report
type RiskReport struct {
	PeriodStart     time.Time              `json:"period_start"`
	PeriodEnd       time.Time              `json:"period_end"`
	TotalTransactions int64                `json:"total_transactions"`
	TotalVolumeUSD  float64                `json:"total_volume_usd"`
	FlaggedCount    int64                  `json:"flagged_count"`
	HighRiskCount   int64                  `json:"high_risk_count"`
	AverageRiskScore float64               `json:"average_risk_score"`
	TopRiskFactors  []RiskFactorSummary    `json:"top_risk_factors"`
	ByChain         map[string]ChainStats  `json:"by_chain"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// RiskFactorSummary summarizes risk factor occurrences
type RiskFactorSummary struct {
	Type       string `json:"type"`
	Count      int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ChainStats contains statistics for a specific chain
type ChainStats struct {
	Transactions   int64   `json:"transactions"`
	VolumeUSD      float64 `json:"volume_usd"`
	Flagged        int64   `json:"flagged"`
	AverageRisk    float64 `json:"average_risk"`
}
