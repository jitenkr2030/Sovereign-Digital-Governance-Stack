package ingest

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/csic/transaction-monitoring/internal/repository"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// IngestionService handles blockchain data ingestion
type IngestionService struct {
	cfg          *config.Config
	repo         *repository.Repository
	ethClient    *ethclient.Client
	btcClient    interface{} // BTC RPC client
	kafkaProducer interface{}
	logger       *zap.Logger
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex

	// State tracking
	lastBlockHeight map[string]int64
	isRunning       bool
}

// NewIngestionService creates a new ingestion service
func NewIngestionService(cfg *config.Config, repo *repository.Repository, logger *zap.Logger) (*IngestionService, error) {
	svc := &IngestionService{
		cfg:             cfg,
		repo:            repo,
		logger:          logger,
		stopChan:        make(chan struct{}),
		lastBlockHeight: make(map[string]int64),
	}

	// Initialize Ethereum client
	if cfg.Blockchain.Ethereum.Enabled {
		client, err := ethclient.Dial(cfg.Blockchain.Ethereum.RPCEndpoint)
		if err != nil {
			return nil, err
		}
		svc.ethClient = client
	}

	return svc, nil
}

// Start begins the ingestion process
func (s *IngestionService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return nil
	}
	s.isRunning = true
	s.mu.Unlock()

	s.logger.Info("Starting blockchain ingestion service")

	// Start Ethereum ingestion
	if s.cfg.Blockchain.Ethereum.Enabled {
		s.wg.Add(1)
		go s.ingestEthereum(ctx))
	}

	// Start Bitcoin ingestion
	if s.cfg.Blockchain.Bitcoin.Enabled {
		s.wg.Add(1)
		go s.ingestBitcoin(ctx))
	}

	return nil
}

// Stop gracefully stops the ingestion service
func (s *IngestionService) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	s.logger.Info("Stopping blockchain ingestion service")
	close(s.stopChan)
	s.wg.Wait()
}

// ingestEthereum continuously ingests Ethereum blocks
func (s *IngestionService) ingestEthereum(ctx context.Context) {
	defer s.wg.Done()

	s.logger.Info("Starting Ethereum block ingestion",
		zap.String("endpoint", s.cfg.Blockchain.Ethereum.RPCEndpoint))

	// Get starting block
	startBlock := s.cfg.Blockchain.Chains[0].StartBlock
	if startBlock == 0 {
		header, err := s.ethClient.HeaderByNumber(ctx, nil)
		if err == nil {
			startBlock = int64(header.Number.Int64()) - 10 // Start 10 blocks back
		}
	}

	s.mu.Lock()
	s.lastBlockHeight["ethereum"] = startBlock
	s.mu.Unlock()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			s.processEthereumBlocks(ctx, startBlock)
			time.Sleep(time.Duration(s.cfg.Blockchain.Chains[0].Confirmations) * time.Second)
		}
	}
}

// processEthereumBlocks processes blocks from the given height
func (s *IngestionService) processEthereumBlocks(ctx context.Context, startBlock int64) {
	s.mu.RLock()
	currentHeight := s.lastBlockHeight["ethereum"]
	s.mu.RUnlock()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			header, err := s.ethClient.HeaderByNumber(ctx, big.NewInt(currentHeight+1))
			if err != nil {
				s.logger.Warn("Failed to get header", zap.Int64("block", currentHeight+1), zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			// Parse block
			block, err := s.ethClient.BlockByNumber(ctx, big.NewInt(currentHeight+1))
			if err != nil {
				s.logger.Warn("Failed to get block", zap.Int64("block", currentHeight+1), zap.Error(err))
				continue
			}

			// Process transactions
			txCount := s.processEthereumBlock(block)

			s.logger.Info("Processed Ethereum block",
				zap.Int64("block", block.Number().Int64()),
				zap.Int("transactions", txCount))

			s.mu.Lock()
			s.lastBlockHeight["ethereum"] = block.Number().Int64()
			s.mu.Unlock()

			currentHeight = block.Number().Int64()
		}
	}
}

// processEthereumBlock processes transactions in an Ethereum block
func (s *IngestionService) processEthereumBlock(block *types.Block) int {
	ctx := context.Background()
	txCount := 0

	for _, tx := range block.Transactions() {
		normalizedTx := s.normalizeEthereumTransaction(tx, block)

		// Save to database
		if err := s.repo.SaveTransaction(ctx, normalizedTx); err != nil {
			s.logger.Error("Failed to save transaction", zap.String("hash", tx.Hash().Hex()), zap.Error(err))
			continue
		}

		// Publish to Kafka for downstream processing
		s.publishNormalizedTransaction(ctx, normalizedTx)

		txCount++
	}

	return txCount
}

// normalizeEthereumTransaction converts an Ethereum transaction to normalized format
func (s *IngestionService) normalizeEthereumTransaction(tx *types.Block, ethBlock *types.Block) *models.NormalizedTransaction {
	var inputs []models.NormalizedInput
	var outputs []models.NormalizedOutput

	sender, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err == nil {
		inputs = append(inputs, models.NormalizedInput{
			Address: sender.Hex(),
			Amount:  s.getTxValue(tx),
		})
	}

	receipt, err := s.ethClient.TransactionReceipt(context.Background(), tx.Hash())
	if err == nil {
		for _, log := range receipt.Logs {
			// Parse token transfers from logs
			if len(log.Topics) >= 3 {
				// ERC-20 Transfer event
				if log.Topics[0].Hex() == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
					from := common.HexToAddress(log.Topics[1].Hex()).Hex()
					to := common.HexToAddress(log.Topics[2].Hex()).Hex()
					outputs = append(outputs, models.NormalizedOutput{
						Address: to,
						Amount:  s.parseLogData(log.Data),
					})
				}
			}
		}
	}

	// If no token transfers, add recipient
	if len(outputs) == 0 && tx.To() != nil {
		outputs = append(outputs, models.NormalizedOutput{
			Address: tx.To().Hex(),
			Amount:  s.getTxValue(tx),
		})
	}

	return &models.NormalizedTransaction{
		TxHash:        tx.Hash().Hex(),
		Network:       models.NetworkEthereum,
		BlockNumber:   ethBlock.Number().Int64(),
		BlockHash:     ethBlock.Hash().Hex(),
		Timestamp:     time.Unix(int64(ethBlock.Time()), 0),
		Inputs:        inputs,
		Outputs:       outputs,
		TotalValue:    s.getTxValue(tx),
		Asset:         "ETH",
		Fee:           s.calculateEthereumFee(tx, receipt),
		InputCount:    len(inputs),
		OutputCount:   len(outputs),
	}
}

// getTxValue returns the transaction value in Ether
func (s *IngestionService) getTxValue(tx *types.Transaction) decimal.Decimal {
	if tx.Value() == nil {
		return decimal.Zero
	}
	// Convert from Wei to Ether (divide by 10^18)
	wei := tx.Value()
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(big.NewInt(1e18)))
	f, _ := ether.Float64()
	return decimal.NewFromFloat(f)
}

// calculateEthereumFee calculates the transaction fee
func (s *IngestionService) calculateEthereumFee(tx *types.Transaction, receipt *types.Receipt) decimal.Decimal {
	if receipt == nil {
		return decimal.Zero
	}
	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	fee := new(big.Int).Mul(gasUsed, tx.GasPrice())
	ether := new(big.Float).Quo(new(big.Float).SetInt(fee), new(big.Float).SetInt(big.NewInt(1e18)))
	f, _ := ether.Float64()
	return decimal.NewFromFloat(f)
}

// parseLogData parses log data as token transfer amount
func (s *IngestionService) parseLogData(data []byte) decimal.Decimal {
	if len(data) >= 32 {
		amount := new(big.Int).SetBytes(data[:32])
		ether := new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(big.NewInt(1e18)))
		f, _ := ether.Float64()
		return decimal.NewFromFloat(f)
	}
	return decimal.Zero
}

// ingestBitcoin handles Bitcoin block ingestion (placeholder)
func (s *IngestionService) ingestBitcoin(ctx context.Context) {
	defer s.wg.Done()
	s.logger.Info("Bitcoin ingestion started (placeholder - requires BTC RPC configuration)")
}

// publishNormalizedTransaction publishes a normalized transaction to Kafka
func (s *IngestionService) publishNormalizedTransaction(ctx context.Context, tx *models.NormalizedTransaction) {
	// Implementation would publish to Kafka topic
	s.logger.Debug("Publishing normalized transaction",
		zap.String("tx_hash", tx.TxHash),
		zap.String("network", string(tx.Network)))
}

// IngestTransaction handles manual transaction ingestion
func (s *IngestionService) IngestTransaction(ctx context.Context, txHash string, network models.Network) error {
	s.logger.Info("Manual transaction ingestion requested",
		zap.String("tx_hash", txHash),
		zap.String("network", string(network)))

	// Implementation would fetch transaction from blockchain and process
	return nil
}

// GetIngestionStatus returns the current ingestion status
func (s *IngestionService) GetIngestionStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]interface{})
	for network, height := range s.lastBlockHeight {
		status[network] = map[string]interface{}{
			"last_block_height": height,
			"is_running":        s.isRunning,
		}
	}
	return status
}

// HealthCheck verifies the service is healthy
func (s *IngestionService) HealthCheck(ctx context.Context) error {
	if s.ethClient != nil {
		_, err := s.ethClient.HeaderByNumber(ctx, nil)
		return err
	}
	return nil
}
