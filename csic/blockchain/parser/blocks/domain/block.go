package domain

import (
	"time"
)

// Block represents a parsed blockchain block
type Block struct {
	Hash              string         `json:"hash"`
	ParentHash        string         `json:"parent_hash"`
	Height            uint64         `json:"height"`
	Timestamp         time.Time      `json:"timestamp"`
	Transactions      []*Transaction `json:"transactions"`
	TransactionCount  int            `json:"transaction_count"`
	Size              uint64         `json:"size"`
	GasUsed           uint64         `json:"gas_used"`
	GasLimit          uint64         `json:"gas_limit"`
	Nonce             string         `json:"nonce"`
	StateRoot         string         `json:"state_root"`
	TransactionsRoot  string         `json:"transactions_root"`
	ReceiptsRoot      string         `json:"receipts_root"`
	Miner             string         `json:"miner"`
	Difficulty        string         `json:"difficulty"`
	TotalDifficulty   string         `json:"total_difficulty"`
	ExtraData         string         `json:"extra_data"`
	BaseFeePerGas     string         `json:"base_fee_per_gas"`
	WithdrawalsRoot   string         `json:"withdrawals_root"`
	BlobGasUsed       uint64         `json:"blob_gas_used"`
	ExcessBlobGas     uint64         `json:"excess_blob_gas"`
	ParsedAt          time.Time      `json:"parsed_at"`
}

// BlockHeader represents just the block header data
type BlockHeader struct {
	Hash            string    `json:"hash"`
	ParentHash      string    `json:"parent_hash"`
	StateRoot       string    `json:"state_root"`
	TransactionsRoot string   `json:"transactions_root"`
	ReceiptsRoot    string    `json:"receipts_root"`
	Number          uint64    `json:"number"`
	GasLimit        uint64    `json:"gas_limit"`
	GasUsed         uint64    `json:"gas_used"`
	Timestamp       time.Time `json:"timestamp"`
	ExtraData       string    `json:"extra_data"`
	MixHash         string    `json:"mix_hash"`
	Nonce           string    `json:"nonce"`
	Bloom           string    `json:"bloom"`
	Difficulty      string    `json:"difficulty"`
	TotalDifficulty string    `json:"total_difficulty"`
	Miner           string    `json:"miner"`
	BaseFeePerGas   string    `json:"base_fee_per_gas"`
}

// BlockFilter defines filters for querying blocks
type BlockFilter struct {
	StartHeight  uint64    `json:"start_height"`
	EndHeight    uint64    `json:"end_height"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Miner        string    `json:"miner"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
	Transactions bool      `json:"include_transactions"`
}

// BlockStats holds aggregated statistics for blocks
type BlockStats struct {
	TotalBlocks      uint64            `json:"total_blocks"`
	TotalTransactions uint64           `json:"total_transactions"`
	TotalGasUsed     uint64            `json:"total_gas_used"`
	AverageGasUsed   float64           `json:"average_gas_used"`
	TimeBetweenBlocks time.Duration    `json:"time_between_blocks"`
	BlockSizeStats   map[string]float64 `json:"block_size_stats"`
	GasUsedStats     map[string]float64 `json:"gas_used_stats"`
}
