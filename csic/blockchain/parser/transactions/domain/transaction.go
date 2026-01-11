package domain

import (
	"time"
)

// Transaction represents a parsed blockchain transaction
type Transaction struct {
	Hash              string    `json:"hash"`
	Nonce             uint64    `json:"nonce"`
	BlockHash         string    `json:"block_hash"`
	BlockNumber       uint64    `json:"block_number"`
	TransactionIndex  uint64    `json:"transaction_index"`
	From              string    `json:"from"`
	To                string    `json:"to"`
	Value             uint64    `json:"value"`
	Gas               uint64    `json:"gas"`
	GasPrice          uint64    `json:"gas_price"`
	MaxFeePerGas      uint64    `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas uint64  `json:"max_priority_fee_per_gas"`
	Input             string    `json:"input"`
	Type              string    `json:"type"`
	ChainID           uint64    `json:"chain_id"`
	Status            string    `json:"status"`
	V                 string    `json:"v"`
	R                 string    `json:"r"`
	S                 string    `json:"s"`
	Timestamp         time.Time `json:"timestamp"`
	Fee               uint64    `json:"fee"`
	GasUsed           uint64    `json:"gas_used"`
	CumulativeGasUsed uint64    `json:"cumulative_gas_used"`
	EffectiveGasPrice uint64    `json:"effective_gas_price"`
	ContractAddress   string    `json:"contract_address"`
	Logs              []*Log    `json:"logs"`
	Root              string    `json:"root"`
}

// Log represents a transaction event log
type Log struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      uint64   `json:"block_number"`
	TransactionHash  string   `json:"transaction_hash"`
	TransactionIndex uint64   `json:"transaction_index"`
	BlockHash        string   `json:"block_hash"`
	LogIndex         uint64   `json:"log_index"`
	Removed          bool     `json:"removed"`
}

// TransactionReceipt represents the receipt of a transaction
type TransactionReceipt struct {
	TransactionHash   string `json:"transaction_hash"`
	TransactionIndex  uint64 `json:"transaction_index"`
	BlockHash         string `json:"block_hash"`
	BlockNumber       uint64 `json:"block_number"`
	From              string `json:"from"`
	To                string `json:"to"`
	CumulativeGasUsed uint64 `json:"cumulative_gas_used"`
	GasUsed           uint64 `json:"gas_used"`
	EffectiveGasPrice uint64 `json:"effective_gas_price"`
	ContractAddress   string `json:"contract_address"`
	Root              string `json:"root"`
	Status            string `json:"status"`
	Logs              []*Log `json:"logs"`
	LogsBloom         string `json:"logs_bloom"`
}

// TransactionFilter defines filters for querying transactions
type TransactionFilter struct {
	FromBlock        uint64   `json:"from_block"`
	ToBlock          uint64   `json:"to_block"`
	FromAddress      string   `json:"from_address"`
	ToAddress        string   `json:"to_address"`
	ContractAddress  string   `json:"contract_address"`
	Hash             string   `json:"hash"`
	BlockHash        string   `json:"block_hash"`
	ValueMin         uint64   `json:"value_min"`
	ValueMax         uint64   `json:"value_max"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	MethodSignature  string   `json:"method_signature"`
	IncludeReceipt   bool     `json:"include_receipt"`
	IncludeLogs      bool     `json:"include_logs"`
	Limit            int      `json:"limit"`
	Offset           int      `json:"offset"`
}

// TransactionStats holds aggregated transaction statistics
type TransactionStats struct {
	TotalTransactions   uint64            `json:"total_transactions"`
	TotalValue          uint64            `json:"total_value"`
	AverageValue        float64           `json:"average_value"`
	TotalFees           uint64            `json:"total_fees"`
	AverageGasPrice     float64           `json:"average_gas_price"`
	GasUsedStats        map[string]float64 `json:"gas_used_stats"`
	ValueDistribution   map[string]int64  `json:"value_distribution"`
	TopSenders          []AddressCount    `json:"top_senders"`
	TopReceivers        []AddressCount    `json:"top_receivers"`
	TransactionTypes    map[string]int64  `json:"transaction_types"`
}

// AddressCount holds count for an address
type AddressCount struct {
	Address string `json:"address"`
	Count   int64  `json:"count"`
}

// InternalTransaction represents an internal transaction (call trace)
type InternalTransaction struct {
	TransactionHash   string `json:"transaction_hash"`
	BlockNumber       uint64 `json:"block_number"`
	Position          int    `json:"position"`
	CallType          string `json:"call_type"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             uint64 `json:"value"`
	Input             string `json:"input"`
	Output            string `json:"output"`
	Gas               uint64 `json:"gas"`
	GasUsed           uint64 `json:"gas_used"`
	TraceAddress      []int  `json:"trace_address"`
	Error             string `json:"error"`
}
