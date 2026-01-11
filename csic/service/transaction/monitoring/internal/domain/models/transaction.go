package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Network represents blockchain network types
type Network string

const (
	NetworkBitcoin  Network = "bitcoin"
	NetworkEthereum Network = "ethereum"
	NetworkPolygon  Network = "polygon"
	NetworkBSC      Network = "bsc"
)

// TransactionStatus represents transaction confirmation status
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusConfirmed TransactionStatus = "confirmed"
	TxStatusFailed    TransactionStatus = "failed"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID          string             `json:"id" db:"id"`
	TxHash      string             `json:"tx_hash" db:"tx_hash"`
	Network     Network            `json:"network" db:"network"`
	BlockNumber int64              `json:"block_number" db:"block_number"`
	BlockHash   string             `json:"block_hash" db:"block_hash"`
	Timestamp   time.Time          `json:"timestamp" db:"timestamp"`
	Sender      string             `json:"sender" db:"sender"`
	Receiver    string             `json:"receiver" db:"receiver"`
	Amount      decimal.Decimal    `json:"amount" db:"amount"`
	Asset       string             `json:"asset" db:"asset"`
	GasUsed     decimal.Decimal    `json:"gas_used" db:"gas_used"`
	GasPrice    decimal.Decimal    `json:"gas_price" db:"gas_price"`
	Fee         decimal.Decimal    `json:"fee" db:"fee"`
	Status      TransactionStatus  `json:"status" db:"status"`
	Metadata    map[string]string  `json:"metadata,omitempty" db:"-"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
}

// TransactionInput represents an input in a transaction
type TransactionInput struct {
	ID            string          `json:"id" db:"id"`
	TransactionID string          `json:"transaction_id" db:"transaction_id"`
	Index         int             `json:"index" db:"index"`
	Address       string          `json:"address" db:"address"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	ScriptPubKey  string          `json:"script_pub_key" db:"script_pub_key"`
}

// TransactionOutput represents an output in a transaction
type TransactionOutput struct {
	ID            string          `json:"id" db:"id"`
	TransactionID string          `json:"transaction_id" db:"transaction_id"`
	Index         int             `json:"index" db:"index"`
	Address       string          `json:"address" db:"address"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	ScriptPubKey  string          `json:"script_pub_key" db:"script_pub_key"`
	IsChange      bool            `json:"is_change" db:"is_change"`
}

// NormalizedTransaction represents a normalized transaction for processing
type NormalizedTransaction struct {
	TxHash        string             `json:"tx_hash"`
	Network       Network            `json:"network"`
	BlockNumber   int64              `json:"block_number"`
	BlockHash     string             `json:"block_hash"`
	Timestamp     time.Time          `json:"timestamp"`
	Inputs        []NormalizedInput  `json:"inputs"`
	Outputs       []NormalizedOutput `json:"outputs"`
	TotalValue    decimal.Decimal    `json:"total_value"`
	Asset         string             `json:"asset"`
	Fee           decimal.Decimal    `json:"fee"`
	InputCount    int                `json:"input_count"`
	OutputCount   int                `json:"output_count"`
}

// NormalizedInput represents a normalized transaction input
type NormalizedInput struct {
	Address  string          `json:"address"`
	Amount   decimal.Decimal `json:"amount"`
	Tag      string          `json:"tag,omitempty"`
}

// NormalizedOutput represents a normalized transaction output
type NormalizedOutput struct {
	Address string          `json:"address"`
	Amount  decimal.Decimal `json:"amount"`
	IsChange bool           `json:"is_change"`
	Tag     string          `json:"tag,omitempty"`
}
