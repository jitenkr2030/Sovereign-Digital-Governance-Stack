package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/csic/platform/blockchain/indexer/internal/domain"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

// BlockchainConfig holds blockchain node configuration
type BlockchainConfig struct {
	Network            string
	NodeURL            string
	WSURL              string
	ChainID            int
	StartBlock         int64
	BatchSize          int
	ConfirmationBlocks int
}

// BlockchainClient handles blockchain node communication
type BlockchainClient struct {
	httpClient *ethclient.Client
	wsClient   *ethclient.Client
	rpcClient  *rpc.Client
	config     BlockchainConfig
	logger     *zap.Logger
}

// NewBlockchainClient creates a new blockchain client
func NewBlockchainClient(config BlockchainConfig, logger *zap.Logger) (*BlockchainClient, error) {
	// Create HTTP client
	httpClient, err := ethclient.Dial(config.NodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create WebSocket client
	wsClient, err := ethclient.Dial(config.WSURL)
	if err != nil {
		httpClient.Close()
		return nil, fmt.Errorf("failed to create WebSocket client: %w", err)
	}

	// Create RPC client for subscriptions
	rpcClient, err := rpc.Dial(config.WSURL)
	if err != nil {
		httpClient.Close()
		wsClient.Close()
		return nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	logger.Info("Connected to blockchain node",
		zap.String("network", config.Network),
		zap.String("node_url", config.NodeURL),
		zap.String("chain_id", fmt.Sprintf("%d", config.ChainID)))

	return &BlockchainClient{
		httpClient: httpClient,
		wsClient:   wsClient,
		rpcClient:  rpcClient,
		config:     config,
		logger:     logger,
	}, nil
}

// Close closes all connections
func (c *BlockchainClient) Close() {
	if c.httpClient != nil {
		c.httpClient.Close()
	}
	if c.wsClient != nil {
		c.wsClient.Close()
	}
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}

// GetCurrentBlockNumber gets the current block number
func (c *BlockchainClient) GetCurrentBlockNumber(ctx context.Context) (int64, error) {
	blockNumber, err := c.httpClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %w", err)
	}
	return int64(blockNumber), nil
}

// GetBlockByNumber gets a block by number
func (c *BlockchainClient) GetBlockByNumber(ctx context.Context, number int64, includeTransactions bool) (*domain.Block, error) {
	block, err := c.httpClient.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number: %w", err)
	}

	return c.convertBlock(block, includeTransactions), nil
}

// GetBlockByHash gets a block by hash
func (c *BlockchainClient) GetBlockByHash(ctx context.Context, hash common.Hash, includeTransactions bool) (*domain.Block, error) {
	block, err := c.httpClient.BlockByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}

	return c.convertBlock(block, includeTransactions), nil
}

// GetTransactionByHash gets a transaction by hash
func (c *BlockchainClient) GetTransactionByHash(ctx context.Context, hash common.Hash) (*domain.Transaction, error) {
	tx, isPending, err := c.httpClient.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Get receipt for status
	receipt, err := c.httpClient.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return c.convertTransaction(tx, receipt, isPending), nil
}

// GetTransactionReceipt gets a transaction receipt
func (c *BlockchainClient) GetTransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	receipt, err := c.httpClient.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}
	return receipt, nil
}

// GetBlockWithTransactions gets a block with all transactions
func (c *BlockchainClient) GetBlockWithTransactions(ctx context.Context, number int64) (*domain.Block, error) {
	return c.GetBlockByNumber(ctx, number, true)
}

// BatchGetBlocks gets multiple blocks in batch
func (c *BlockchainClient) BatchGetBlocks(ctx context.Context, startBlock, endBlock int64) ([]*domain.Block, error) {
	blocks := make([]*domain.Block, 0, endBlock-startBlock + 1)

	for i := startBlock; i <= endBlock; i++ {
		block, err := c.GetBlockByNumber(ctx, i, false)
		if err != nil {
			c.logger.Warn("Failed to get block",
				zap.Int64("block_number", i),
				zap.Error(err))
			continue
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// FilterLogs gets logs for a block range
func (c *BlockchainClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	logs, err := c.httpClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}
	return logs, nil
}

// GetLogsForBlock gets logs for a specific block
func (c *BlockchainClient) GetLogsForBlock(ctx context.Context, blockNumber int64) ([]types.Log, endBlock *big.Int, err error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNumber),
		ToBlock:   big.NewInt(blockNumber),
	}

	logs, err := c.FilterLogs(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	return logs, big.NewInt(blockNumber), nil
}

// GetBalance gets the balance of an address
func (c *BlockchainClient) GetBalance(ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error) {
	balance, err := c.httpClient.BalanceAt(ctx, address, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// GetCode gets the bytecode of an address
func (c *BlockchainClient) GetCode(ctx context.Context, address common.Address, blockNumber *big.Int) ([]byte, error) {
	code, err := c.httpClient.CodeAt(ctx, address, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get code: %w", err)
	}
	return code, nil
}

// IsContract checks if an address is a contract
func (c *BlockchainClient) IsContract(ctx context.Context, address common.Address) (bool, error) {
	code, err := c.GetCode(ctx, address, nil)
	if err != nil {
		return false, err
	}
	return len(code) > 0, nil
}

// GetChainID gets the chain ID
func (c *BlockchainClient) GetChainID(ctx context.Context) (int, error) {
	chainID, err := c.httpClient.ChainID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get chain ID: %w", err)
	}
	return int(chainID.Int64()), nil
}

// GetNetworkID gets the network ID
func (c *BlockchainClient) GetNetworkID(ctx context.Context) (uint64, error) {
	networkID, err := c.httpClient.NetworkID(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get network ID: %w", err)
	}
	return networkID.Uint64(), nil
}

// SyncProgress gets the sync progress
func (c *BlockchainClient) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	progress, err := c.httpClient.SyncProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync progress: %w", err)
	}
	return progress, nil
}

// convertBlock converts an Ethereum block to our domain model
func (c *BlockchainClient) convertBlock(block *types.Block, includeTransactions bool) *domain.Block {
	domainBlock := &domain.Block{
		ID:              block.Hash().Hex(),
		Hash:            block.Hash().Hex(),
		ParentHash:      block.ParentHash().Hex(),
		Number:          block.Number().Int64(),
		Timestamp:       time.Unix(int64(block.Time()), 0),
		Miner:           block.Coinbase().Hex(),
		Difficulty:      block.Difficulty().String(),
		GasLimit:        block.GasLimit(),
		GasUsed:         block.GasUsed(),
		TransactionCount: len(block.Transactions()),
		BaseFeePerGas:   block.BaseFee().String(),
		ExtraData:       strings.ToHex(block.Extra()),
		Size:            block.Size().Int64(),
		UncleCount:      len(block.Uncles()),
		CreatedAt:       time.Now(),
		IndexedAt:       time.Now(),
	}

	if includeTransactions {
		domainBlock.Transactions = make([]domain.Transaction, len(block.Transactions()))
		for i, tx := range block.Transactions() {
			domainBlock.Transactions[i] = *c.convertTransaction(tx, nil, false)
		}
	}

	return domainBlock
}

// convertTransaction converts an Ethereum transaction to our domain model
func (c *BlockchainClient) convertTransaction(tx *types.Transaction, receipt *types.Receipt, isPending bool) *domain.Transaction {
	domainTx := &domain.Transaction{
		ID:              tx.Hash().Hex(),
		Hash:            tx.Hash().Hex(),
		BlockHash:       tx.BlockHash().Hex(),
		BlockNumber:     tx.BlockNumber().Int64(),
		TransactionIndex: tx.Index(),
		From:            c.getFromAddress(tx),
		To:              c.getToAddress(tx),
		Value:           tx.Value().String(),
		GasPrice:        tx.GasPrice().String(),
		Gas:             tx.Gas(),
		InputData:       strings.ToHex(tx.Data()),
		Nonce:           tx.Nonce(),
		V:               fmt.Sprintf("0x%x", tx.V()),
		R:               tx.R().String(),
		S:               tx.S().String(),
		Timestamp:       time.Now(),
		CreatedAt:       time.Now(),
	}

	// Set status
	if isPending {
		domainTx.Status = domain.TxStatusPending
	} else if receipt != nil {
		if receipt.Status == types.ReceiptStatusSuccessful {
			domainTx.Status = domain.TxStatusConfirmed
		} else {
			domainTx.Status = domain.TxStatusFailed
		}
		domainTx.GasUsed = receipt.GasUsed
		domainTx.CumulativeGasUsed = receipt.CumulativeGasUsed
		if len(receipt.Logs) > 0 {
			domainTx.ContractAddress = receipt.Logs[0].Address.Hex()
		}
	}

	return domainTx
}

// getFromAddress gets the from address of a transaction
func (c *BlockchainClient) getFromAddress(tx *types.Transaction) string {
	if from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx); err == nil {
		return from.Hex()
	}
	return ""
}

// getToAddress gets the to address of a transaction
func (c *BlockchainClient) getToAddress(tx *types.Transaction) string {
	if tx.To() != nil {
		return tx.To().Hex()
	}
	return ""
}

// ParseERC20Transfer parses an ERC20 transfer event
func ParseERC20Transfer(log types.Log) (*domain.TokenTransfer, error) {
	// Standard ERC20 Transfer event signature
	eventSignature := []byte{0xdd, 0xf2, 0x52, 0xad, 0x1b, 0xe2, 0xc8, 0x32, 0x4a, 0xbf, 0xb1, 0xf4, 0x5f, 0x8a, 0x2c, 0x1b, 0x02, 0x68, 0x7e, 0x6a, 0x83, 0x14, 0x6c, 0x09, 0x62, 0x91, 0xc3, 0x37, 0x23, 0x51, 0x74, 0x59}

	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("invalid ERC20 Transfer log")
	}

	// Verify event signature
	if len(log.Topics) < 1 || log.Topics[0] != common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") {
		return nil, fmt.Errorf("not an ERC20 Transfer event")
	}

	transfer := &domain.TokenTransfer{
		TransactionHash: log.TxHash.Hex(),
		BlockNumber:     log.BlockNumber,
		TokenAddress:    log.Address.Hex(),
		From:            common.HexToAddress(log.Topics[1].Hex()).Hex(),
		To:              common.HexToAddress(log.Topics[2].Hex()).Hex(),
		TokenType:       domain.TokenTypeERC20,
		Timestamp:       time.Now(),
	}

	// Parse value from data
	if len(log.Data) > 0 {
		value := new(big.Int)
		value.SetBytes(log.Data)
		transfer.Value = value.String()
	}

	return transfer, nil
}

// DecodeMethodCall decodes a method call input data
func DecodeMethodCall(inputData string, abiJSON string) (string, map[string]interface{}, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	data := common.FromHex(inputData)
	if len(data) < 4 {
		return "", nil, fmt.Errorf("invalid input data")
	}

	methodID := common.BytesToHash(data[:4])
	method, err := parsedABI.MethodById(data)
	if err != nil {
		return methodID.Hex(), nil, nil
	}

	inputs, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		return method.Sig, nil, fmt.Errorf("failed to unpack inputs: %w", err)
	}

	result := make(map[string]interface{})
	for i, input := range method.Inputs {
		result[input.Name] = inputs[i]
	}

	return method.Sig, result, nil
}

// MarshalBlockForJSON marshals a block for JSON serialization
func MarshalBlockForJSON(block *domain.Block) ([]byte, error) {
	return json.MarshalIndent(block, "", "  ")
}

// MarshalTransactionForJSON marshals a transaction for JSON serialization
func MarshalTransactionForJSON(tx *domain.Transaction) ([]byte, error) {
	return json.MarshalIndent(tx, "", "  ")
}
