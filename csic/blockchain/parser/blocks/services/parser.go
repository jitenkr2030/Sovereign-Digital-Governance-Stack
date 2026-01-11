package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/csic-platform/blockchain/parser/blocks/domain"
	"github.com/csic-platform/blockchain/parser/blocks/ports"
	"github.com/csic-platform/blockchain/parser/transactions/domain"
)

// BlockParserService implements the BlockParser interface
type BlockParserService struct {
	repo      ports.BlockRepository
	publisher ports.BlockEventPublisher
	mu        sync.RWMutex
}

// NewBlockParserService creates a new block parser service
func NewBlockParserService(repo ports.BlockRepository, publisher ports.BlockEventPublisher) *BlockParserService {
	return &BlockParserService{
		repo:      repo,
		publisher: publisher,
	}
}

// ParseBlock parses raw block data into a Block structure
func (s *BlockParserService) ParseBlock(ctx context.Context, rawData []byte) (*domain.Block, error) {
	var rawBlock map[string]interface{}
	if err := json.Unmarshal(rawData, &rawBlock); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block data: %w", err)
	}

	block, err := s.parseBlockFromMap(rawBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}

	return block, nil
}

// parseBlockFromMap converts a map to a Block structure
func (s *BlockParserService) parseBlockFromMap(data map[string]interface{}) (*domain.Block, error) {
	block := &domain.Block{
		ParsedAt: time.Now(),
	}

	// Parse header fields
	if hash, ok := data["hash"].(string); ok {
		block.Hash = hash
	}
	if parentHash, ok := data["parentHash"].(string); ok {
		block.ParentHash = parentHash
	}
	if stateRoot, ok := data["stateRoot"].(string); ok {
		block.StateRoot = stateRoot
	}
	if transactionsRoot, ok := data["transactionsRoot"].(string); ok {
		block.TransactionsRoot = transactionsRoot
	}
	if receiptsRoot, ok := data["receiptsRoot"].(string); ok {
		block.ReceiptsRoot = receiptsRoot
	}
	if miner, ok := data["miner"].(string); ok {
		block.Miner = miner
	}
	if extraData, ok := data["extraData"].(string); ok {
		block.ExtraData = extraData
	}
	if mixHash, ok := data["mixHash"].(string); ok {
		block.ExtraData = mixHash
	}
	if nonce, ok := data["nonce"].(string); ok {
		block.Nonce = nonce
	}
	if difficulty, ok := data["difficulty"].(string); ok {
		block.Difficulty = difficulty
	}
	if totalDifficulty, ok := data["totalDifficulty"].(string); ok {
		block.TotalDifficulty = totalDifficulty
	}
	if baseFeePerGas, ok := data["baseFeePerGas"].(string); ok {
		block.BaseFeePerGas = baseFeePerGas
	}
	if withdrawalsRoot, ok := data["withdrawalsRoot"].(string); ok {
		block.WithdrawalsRoot = withdrawalsRoot
	}

	// Parse numeric fields
	if number, ok := data["number"].(string); ok {
		fmt.Sscanf(number, "0x%x", &block.Height)
	}
	if gasUsed, ok := data["gasUsed"].(string); ok {
		fmt.Sscanf(gasUsed, "0x%x", &block.GasUsed)
	}
	if gasLimit, ok := data["gasLimit"].(string); ok {
		fmt.Sscanf(gasLimit, "0x%x", &block.GasLimit)
	}
	if size, ok := data["size"].(string); ok {
		fmt.Sscanf(size, "0x%x", &block.Size)
	}
	if blobGasUsed, ok := data["blobGasUsed"].(string); ok {
		fmt.Sscanf(blobGasUsed, "0x%x", &block.BlobGasUsed)
	}
	if excessBlobGas, ok := data["excessBlobGas"].(string); ok {
		fmt.Sscanf(excessBlobGas, "0x%x", &block.ExcessBlobGas)
	}

	// Parse timestamp
	if timestamp, ok := data["timestamp"].(string); ok {
		var ts uint64
		fmt.Sscanf(timestamp, "0x%x", &ts)
		block.Timestamp = time.Unix(int64(ts), 0)
	}

	// Parse transactions
	if transactions, ok := data["transactions"].([]interface{}); ok {
		block.TransactionCount = len(transactions)
		for _, txRaw := range transactions {
			txData, ok := txRaw.(map[string]interface{})
			if !ok {
				continue
			}
			tx := s.parseTransactionFromMap(txData)
			if tx != nil {
				block.Transactions = append(block.Transactions, tx)
			}
		}
	}

	return block, nil
}

// parseTransactionFromMap converts a map to a Transaction structure
func (s *BlockParserService) parseTransactionFromMap(data map[string]interface{}) *domain.Transaction {
	tx := &domain.Transaction{}

	if hash, ok := data["hash"].(string); ok {
		tx.Hash = hash
	}
	if nonce, ok := data["nonce"].(string); ok {
		fmt.Sscanf(nonce, "0x%x", &tx.Nonce)
	}
	if blockHash, ok := data["blockHash"].(string); ok {
		tx.BlockHash = blockHash
	}
	if from, ok := data["from"].(string); ok {
		tx.From = from
	}
	if to, ok := data["to"].(string); ok {
		tx.To = to
	}
	if input, ok := data["input"].(string); ok {
		tx.Input = input
	}
	if value, ok := data["value"].(string); ok {
		fmt.Sscanf(value, "0x%x", &tx.Value)
	}
	if gas, ok := data["gas"].(string); ok {
		fmt.Sscanf(gas, "0x%x", &tx.Gas)
	}
	if gasPrice, ok := data["gasPrice"].(string); ok {
		fmt.Sscanf(gasPrice, "0x%x", &tx.GasPrice)
	}
	if maxFeePerGas, ok := data["maxFeePerGas"].(string); ok {
		fmt.Sscanf(maxFeePerGas, "0x%x", &tx.MaxFeePerGas)
	}
	if maxPriorityFeePerGas, ok := data["maxPriorityFeePerGas"].(string); ok {
		fmt.Sscanf(maxPriorityFeePerGas, "0x%x", &tx.MaxPriorityFeePerGas)
	}
	if transactionIndex, ok := data["transactionIndex"].(string); ok {
		fmt.Sscanf(transactionIndex, "0x%x", &tx.TransactionIndex)
	}
	if chainId, ok := data["chainId"].(string); ok {
		fmt.Sscanf(chainId, "0x%x", &tx.ChainID)
	}

	// Parse boolean fields
	if v, ok := data["type"].(string); ok {
		tx.Type = v
	}
	if v, ok := data["status"].(string); ok {
		tx.Status = v
	}
	if v, ok := data["v"].(string); ok {
		tx.V = v
	}
	if r, ok := data["r"].(string); ok {
		tx.R = r
	}
	if s, ok := data["s"].(string); ok {
		tx.S = s
	}

	// Parse block number and transaction index
	if blockNumber, ok := data["blockNumber"].(string); ok {
		fmt.Sscanf(blockNumber, "0x%x", &tx.BlockNumber)
	}

	return tx
}

// ParseBlockHeader parses raw block header data
func (s *BlockParserService) ParseBlockHeader(ctx context.Context, rawData []byte) (*domain.BlockHeader, error) {
	var rawHeader map[string]interface{}
	if err := json.Unmarshal(rawData, &rawHeader); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	header := &domain.BlockHeader{}

	if hash, ok := rawHeader["hash"].(string); ok {
		header.Hash = hash
	}
	if parentHash, ok := rawHeader["parentHash"].(string); ok {
		header.ParentHash = parentHash
	}
	if stateRoot, ok := rawHeader["stateRoot"].(string); ok {
		header.StateRoot = stateRoot
	}
	if transactionsRoot, ok := rawHeader["transactionsRoot"].(string); ok {
		header.TransactionsRoot = transactionsRoot
	}
	if receiptsRoot, ok := rawHeader["receiptsRoot"].(string); ok {
		header.ReceiptsRoot = receiptsRoot
	}
	if miner, ok := rawHeader["miner"].(string); ok {
		header.Miner = miner
	}
	if extraData, ok := rawHeader["extraData"].(string); ok {
		header.ExtraData = extraData
	}
	if mixHash, ok := rawHeader["mixHash"].(string); ok {
		header.MixHash = mixHash
	}
	if nonce, ok := rawHeader["nonce"].(string); ok {
		header.Nonce = nonce
	}
	if difficulty, ok := rawHeader["difficulty"].(string); ok {
		header.Difficulty = difficulty
	}
	if totalDifficulty, ok := rawHeader["totalDifficulty"].(string); ok {
		header.TotalDifficulty = totalDifficulty
	}
	if baseFeePerGas, ok := rawHeader["baseFeePerGas"].(string); ok {
		header.BaseFeePerGas = baseFeePerGas
	}
	if bloom, ok := rawHeader["logsBloom"].(string); ok {
		header.Bloom = bloom
	}

	// Parse numeric fields
	if number, ok := rawHeader["number"].(string); ok {
		fmt.Sscanf(number, "0x%x", &header.Number)
	}
	if gasLimit, ok := rawHeader["gasLimit"].(string); ok {
		fmt.Sscanf(gasLimit, "0x%x", &header.GasLimit)
	}
	if gasUsed, ok := rawHeader["gasUsed"].(string); ok {
		fmt.Sscanf(gasUsed, "0x%x", &header.GasUsed)
	}

	if timestamp, ok := rawHeader["timestamp"].(string); ok {
		var ts uint64
		fmt.Sscanf(timestamp, "0x%x", &ts)
		header.Timestamp = time.Unix(int64(ts), 0)
	}

	return header, nil
}

// ValidateBlock performs validation on a parsed block
func (s *BlockParserService) ValidateBlock(ctx context.Context, block *domain.Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	if block.Hash == "" {
		return fmt.Errorf("block hash is empty")
	}
	if block.Height == 0 {
		return fmt.Errorf("block height is zero")
	}
	if len(block.Transactions) != block.TransactionCount {
		return fmt.Errorf("transaction count mismatch: expected %d, got %d",
			block.TransactionCount, len(block.Transactions))
	}
	return nil
}

// ValidateBlockHeader performs validation on a parsed block header
func (s *BlockParserService) ValidateBlockHeader(ctx context.Context, header *domain.BlockHeader) error {
	if header == nil {
		return fmt.Errorf("header is nil")
	}
	if header.Hash == "" {
		return fmt.Errorf("header hash is empty")
	}
	if header.Number == 0 {
		return fmt.Errorf("header number is zero")
	}
	return nil
}

// SerializeBlock serializes a block to JSON
func (s *BlockParserService) SerializeBlock(ctx context.Context, block *domain.Block) ([]byte, error) {
	return json.MarshalIndent(block, "", "  ")
}

// DeserializeBlock deserializes a block from JSON
func (s *BlockParserService) DeserializeBlock(ctx context.Context, data []byte) (*domain.Block, error) {
	var block domain.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}
	return &block, nil
}
