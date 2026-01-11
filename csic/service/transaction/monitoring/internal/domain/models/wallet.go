package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Wallet represents a cryptocurrency wallet address
type Wallet struct {
	ID             string          `json:"id" db:"id"`
	Address        string          `json:"address" db:"address"`
	Network        Network         `json:"network" db:"network"`
	WalletType     string          `json:"wallet_type" db:"wallet_type"` // EOA, Contract, Multisig
	Label          string          `json:"label" db:"label"`
	Tags           []string        `json:"tags" db:"-"`
	FirstSeen      time.Time       `json:"first_seen" db:"first_seen"`
	LastSeen       time.Time       `json:"last_seen" db:"last_seen"`
	TxCount        int64           `json:"tx_count" db:"tx_count"`
	TotalReceived  decimal.Decimal `json:"total_received" db:"total_received"`
	TotalSent      decimal.Decimal `json:"total_sent" db:"total_sent"`
	CurrentBalance decimal.Decimal `json:"current_balance" db:"current_balance"`
	RiskScore      float64         `json:"risk_score" db:"risk_score"`
	RiskLevel      string          `json:"risk_level" db:"risk_level"`
	IsSanctioned   bool            `json:"is_sanctioned" db:"is_sanctioned"`
	IsBlacklisted  bool            `json:"is_blacklisted" db:"is_blacklisted"`
	IsWhitelisted  bool            `json:"is_whitelisted" db:"is_whitelisted"`
	ClusterID      *string         `json:"cluster_id" db:"cluster_id"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// WalletTag represents tags associated with a wallet
type WalletTag struct {
	ID        string    `json:"id" db:"id"`
	WalletID  string    `json:"wallet_id" db:"wallet_id"`
	Tag       string    `json:"tag" db:"tag"`
	Source    string    `json:"source" db:"source"` // On-chain, External, User
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WalletMetrics contains calculated metrics for a wallet
type WalletMetrics struct {
	WalletID             string          `json:"wallet_id"`
	Velocity24h          decimal.Decimal `json:"velocity_24h"`
	Velocity7d           decimal.Decimal `json:"velocity_7d"`
	AverageTxSize        decimal.Decimal `json:"average_tx_size"`
	MaxTxSize            decimal.Decimal `json:"max_tx_size"`
	UniqueCounterparties int             `json:"unique_counterparties"`
	Volume24h            decimal.Decimal `json:"volume_24h"`
	Volume7d             decimal.Decimal `json:"volume_7d"`
	IncomingRatio        float64         `json:"incoming_ratio"`
	OutgoingRatio        float64         `json:"outgoing_ratio"`
	ActiveDaysLast30     int             `json:"active_days_last_30"`
	FirstTxAge           int             `json:"first_tx_age_days"`
	LastTxAge            int             `json:"last_tx_age_days"`
}

// RiskFactors represents individual risk factors for a wallet
type RiskFactors struct {
	WalletID            string  `json:"wallet_id"`
	HistoricalScore     float64 `json:"historical_score"`
	VelocityScore       float64 `json:"velocity_score"`
	PatternScore        float64 `json:"pattern_score"`
	LinkScore           float64 `json:"link_score"`
	SanctionsScore      float64 `json:"sanctions_score"`
	GeolocationScore    float64 `json:"geolocation_score"`
	BehavioralScore     float64 `json:"behavioral_score"`
	TotalScore          float64 `json:"total_score"`
	TriggeredRules      []string `json:"triggered_rules"`
	LastCalculatedAt    string   `json:"last_calculated_at"`
}

// WalletHistory represents historical data for a wallet
type WalletHistory struct {
	ID            string          `json:"id" db:"id"`
	WalletID      string          `json:"wallet_id" db:"wallet_id"`
	Date          time.Time       `json:"date" db:"date"`
	TxCount       int64           `json:"tx_count" db:"tx_count"`
	Volume        decimal.Decimal `json:"volume" db:"volume"`
	RiskScore     float64         `json:"risk_score" db:"risk_score"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// WalletLabel represents user-defined labels for a wallet
type WalletLabel struct {
	ID        string    `json:"id" db:"id"`
	WalletID  string    `json:"wallet_id" db:"wallet_id"`
	Label     string    `json:"label" db:"label"`
	Category  string    `json:"category" db:"category"`
	Source    string    `json:"source" db:"source"`
	Confirmed bool      `json:"confirmed" db:"confirmed"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// WalletActivity represents wallet activity for monitoring
type WalletActivity struct {
	ID          string    `json:"id" db:"id"`
	WalletID    string    `json:"wallet_id" db:"wallet_id"`
	ActivityType string   `json:"activity_type" db:"activity_type"` // tx_sent, tx_received, label_added
	ReferenceID string    `json:"reference_id" db:"reference_id"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Network     Network   `json:"network" db:"network"`
	Details     string    `json:"details" db:"details"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
