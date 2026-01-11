package domain

import (
	"time"
)

// Block represents a blockchain block
type Block struct {
	ID              string                 `json:"id"`
	Hash            string                 `json:"hash"`
	ParentHash      string                 `json:"parent_hash"`
	Number          int64                  `json:"number"`
	Timestamp       time.Time              `json:"timestamp"`
	Miner           string                 `json:"miner"`
	Difficulty      string                 `json:"difficulty"`
	GasLimit        uint64                 `json:"gas_limit"`
	GasUsed         uint64                 `json:"gas_used"`
	TransactionCount int                   `json:"transaction_count"`
	BaseFeePerGas   string                 `json:"base_fee_per_gas"`
	ExtraData       string                 `json:"extra_data"`
	Transactions    []Transaction          `json:"transactions,omitempty"`
	Size            uint64                 `json:"size"`
	UncleCount      int                    `json:"uncle_count"`
	CreatedAt       time.Time              `json:"created_at"`
	IndexedAt       time.Time              `json:"indexed_at"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ID              string                 `json:"id"`
	Hash            string                 `json:"hash"`
	BlockHash       string                 `json:"block_hash"`
	BlockNumber     int64                  `json:"block_number"`
	TransactionIndex int                   `json:"transaction_index"`
	From            string                 `json:"from"`
	To              string                 `json:"to"`
	Value           string                 `json:"value"`
	GasPrice        string                 `json:"gas_price"`
	Gas             uint64                 `json:"gas"`
	InputData       string                 `json:"input_data"`
	Nonce           uint64                 `json:"nonce"`
	V               string                 `json:"v"`
	R               string                 `json:"r"`
	S               string                 `json:"s"`
	Status          TransactionStatus      `json:"status"`
	GasUsed         uint64                 `json:"gas_used"`
	ContractAddress string                 `json:"contract_address,omitempty"`
	CumulativeGasUsed uint64               `json:"cumulative_gas_used"`
	Timestamp       time.Time              `json:"timestamp"`
	CreatedAt       time.Time              `json:"created_at"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "PENDING"
	TxStatusConfirmed TransactionStatus = "CONFIRMED"
	TxStatusFailed    TransactionStatus = "FAILED"
	TxStatusReverted  TransactionStatus = "REVERTED"
)

// Log represents a smart contract event log
type Log struct {
	ID              string                 `json:"id"`
	TransactionHash string                 `json:"transaction_hash"`
	BlockHash       string                 `json:"block_hash"`
	BlockNumber     int64                  `json:"block_number"`
	LogIndex        int                    `json:"log_index"`
	Address         string                 `json:"address"`
	Topics          []string               `json:"topics"`
	Data            string                 `json:"data"`
	Removed         bool                   `json:"removed"`
	CreatedAt       time.Time              `json:"created_at"`
}

// TokenTransfer represents an ERC20/ERC721 token transfer
type TokenTransfer struct {
	ID              string                 `json:"id"`
	TransactionHash string                 `json:"transaction_hash"`
	BlockNumber     int64                  `json:"block_number"`
	TokenAddress    string                 `json:"token_address"`
	From            string                 `json:"from"`
	To              string                 `json:"to"`
	Value           string                 `json:"value"`
	TokenID         string                 `json:"token_id,omitempty"`
	TokenType       TokenType              `json:"token_type"`
	Timestamp       time.Time              `json:"timestamp"`
	CreatedAt       time.Time              `json:"created_at"`
}

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeERC20  TokenType = "ERC20"
	TokenTypeERC721 TokenType = "ERC721"
	TokenTypeERC1155 TokenType = "ERC1155"
)

// AddressMeta stores metadata for addresses
type AddressMeta struct {
	Address        string                 `json:"address"`
	Label          string                 `json:"label,omitempty"`
	Type           AddressType            `json:"type"`
	FirstSeen      time.Time              `json:"first_seen"`
	LastSeen       time.Time              `json:"last_seen"`
	TxCount        int                    `json:"tx_count"`
	IsContract     bool                   `json:"is_contract"`
	ContractName   string                 `json:"contract_name,omitempty"`
	ContractSymbol string                 `json:"contract_symbol,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// AddressType represents the type of address
type AddressType string

const (
	AddressTypeEOA        AddressType = "EOA"
	AddressTypeContract   AddressType = "CONTRACT"
	AddressTypeExchange   AddressType = "EXCHANGE"
	AddressTypeMiner      AddressType = "MINER"
	AddressTypeProtocol   AddressType = "PROTOCOL"
)

// Contract represents a smart contract
type Contract struct {
	Address        string                 `json:"address"`
	Name           string                 `json:"name"`
	Symbol         string                 `json:"symbol,omitempty"`
	Decimals       int                    `json:"decimals,omitempty"`
	TotalSupply    string                 `json:"total_supply,omitempty"`
	CompilerVersion string                `json:"compiler_version,omitempty"`
	ABI            string                 `json:"abi"`
	Bytecode       string                 `json:"bytecode"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// BlockQuery represents a query for blocks
type BlockQuery struct {
	StartBlock   int64     `json:"start_block"`
	EndBlock     int64     `json:"end_block"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Miner        string    `json:"miner"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
	OrderBy      string    `json:"order_by"`
	Order        string    `json:"order"`
}

// TransactionQuery represents a query for transactions
type TransactionQuery struct {
	BlockNumber     int64    `json:"block_number"`
	FromAddress     string   `json:"from_address"`
	ToAddress       string   `json:"to_address"`
	ContractAddress string   `json:"contract_address"`
	Status         []TransactionStatus `json:"status"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Limit           int      `json:"limit"`
	Offset          int      `json:"offset"`
	OrderBy         string   `json:"order_by"`
	Order           string   `json:"order"`
}

// AddressQuery represents a query for addresses
type AddressQuery struct {
	Address     string      `json:"address"`
	Type        AddressType `json:"type"`
	Label       string      `json:"label"`
	HasLabel    bool        `json:"has_label"`
	MinTxCount  int         `json:"min_tx_count"`
	Limit       int         `json:"limit"`
	Offset      int         `json:"offset"`
}

// TokenQuery represents a query for tokens
type TokenQuery struct {
	TokenAddress  string    `json:"token_address"`
	TokenType     TokenType `json:"token_type"`
	HolderAddress string    `json:"holder_address"`
	MinValue      string    `json:"min_value"`
	Limit         int       `json:"limit"`
	Offset        int       `json:"offset"`
}

// SyncStatus represents the synchronization status
type SyncStatus struct {
	Network          string    `json:"network"`
	CurrentBlock     int64     `json:"current_block"`
	IndexedBlock     int64     `json:"indexed_block"`
	IndexedPercent   float64   `json:"indexed_percent"`
	PendingBlocks    int64     `json:"pending_blocks"`
	LastSyncTime     time.Time `json:"last_sync_time"`
	SyncStatus       string    `json:"sync_status"`
}

// IndexerEvent represents an event from the indexer
type IndexerEvent struct {
	Type      string                 `json:"type"`
	Block     *Block                 `json:"block,omitempty"`
	Transaction *Transaction         `json:"transaction,omitempty"`
	Log       *Log                   `json:"log,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
