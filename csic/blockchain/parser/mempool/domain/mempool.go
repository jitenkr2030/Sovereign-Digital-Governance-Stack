package domain

import (
	"time"
)

// MempoolTransaction represents a transaction in the mempool
type MempoolTransaction struct {
	Hash              string    `json:"hash"`
	Nonce             uint64    `json:"nonce"`
	From              string    `json:"from"`
	To                string    `json:"to"`
	Value             uint64    `json:"value"`
	Gas               uint64    `json:"gas"`
	GasPrice          uint64    `json:"gas_price"`
	MaxFeePerGas       uint64    `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas uint64  `json:"max_priority_fee_per_gas"`
	Input             string    `json:"input"`
	Type              string    `json:"type"`
	ChainID           uint64    `json:"chain_id"`
	V                 string    `json:"v"`
	R                 string    `json:"r"`
	S                 string    `json:"s"`
	ReceivedAt        time.Time `json:"received_at"`
	GasPriceGwei      float64   `json:"gas_price_gwei"`
	EstimatedConfirmTime float64 `json:"estimated_confirm_time"`
}

// MempoolStats holds mempool statistics
type MempoolStats struct {
	TransactionCount    int               `json:"transaction_count"`
	TotalValue          uint64            `json:"total_value"`
	AverageGasPrice     uint64            `json:"average_gas_price"`
	MedianGasPrice      uint64            `json:"median_gas_price"`
	GasPriceDistribution map[string]int   `json:"gas_price_distribution"`
	TopSenders          []AddressCount    `json:"top_senders"`
	TopReceivers        []AddressCount    `json:"top_receivers"`
	PendingValue        uint64            `json:"pending_value"`
	ByType              map[string]int    `json:"by_type"`
}

// MempoolFilter defines filters for mempool queries
type MempoolFilter struct {
	MinGasPrice        uint64   `json:"min_gas_price"`
	MaxGasPrice        uint64   `json:"max_gas_price"`
	FromAddress        string   `json:"from_address"`
	ToAddress          string   `json:"to_address"`
	ValueMin           uint64   `json:"value_min"`
	ValueMax           uint64   `json:"value_max"`
	ContractAddress    string   `json:"contract_address"`
	MethodSignature    string   `json:"method_signature"`
	Limit              int      `json:"limit"`
	SortByGasPrice     bool     `json:"sort_by_gas_price"`
	SortByValue        bool     `json:"sort_by_value"`
}

// PendingBlock represents a proposed block from the mempool
type PendingBlock struct {
	Transactions   []*MempoolTransaction `json:"transactions"`
	TotalGasUsed   uint64                 `json:"total_gas_used"`
	TotalValue     uint64                 `json:"total_value"`
	BlockGasLimit  uint64                 `json:"block_gas_limit"`
	BaseFeePerGas  uint64                 `json:"base_fee_per_gas"`
	ProposedAt     time.Time              `json:"proposed_at"`
}
