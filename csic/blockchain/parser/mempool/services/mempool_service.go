package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/csic-platform/blockchain/parser/mempool/domain"
	"github.com/csic-platform/blockchain/parser/mempool/ports"
)

// MempoolParserService implements the MempoolParser interface
type MempoolParserService struct{}

// NewMempoolParserService creates a new mempool parser service
func NewMempoolParserService() *MempoolParserService {
	return &MempoolParserService{}
}

// ParseTransaction parses raw transaction data into a MempoolTransaction
func (s *MempoolParserService) ParseTransaction(ctx context.Context, rawData []byte) (*domain.MempoolTransaction, error) {
	var rawTx map[string]interface{}
	if err := json.Unmarshal(rawData, &rawTx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	tx := &domain.MempoolTransaction{
		ReceivedAt: time.Now(),
	}

	if hash, ok := rawTx["hash"].(string); ok {
		tx.Hash = hash
	}
	if nonce, ok := rawTx["nonce"].(string); ok {
		fmt.Sscanf(nonce, "0x%x", &tx.Nonce)
	}
	if from, ok := rawTx["from"].(string); ok {
		tx.From = from
	}
	if to, ok := rawTx["to"].(string); ok {
		tx.To = to
	}
	if input, ok := rawTx["input"].(string); ok {
		tx.Input = input
	}
	if value, ok := rawTx["value"].(string); ok {
		fmt.Sscanf(value, "0x%x", &tx.Value)
	}
	if gas, ok := rawTx["gas"].(string); ok {
		fmt.Sscanf(gas, "0x%x", &tx.Gas)
	}
	if gasPrice, ok := rawTx["gasPrice"].(string); ok {
		fmt.Sscanf(gasPrice, "0x%x", &tx.GasPrice)
	}
	if maxFeePerGas, ok := rawTx["maxFeePerGas"].(string); ok {
		fmt.Sscanf(maxFeePerGas, "0x%x", &tx.MaxFeePerGas)
	}
	if maxPriorityFeePerGas, ok := rawTx["maxPriorityFeePerGas"].(string); ok {
		fmt.Sscanf(maxPriorityFeePerGas, "0x%x", &tx.MaxPriorityFeePerGas)
	}
	if txType, ok := rawTx["type"].(string); ok {
		tx.Type = txType
	}
	if chainId, ok := rawTx["chainId"].(string); ok {
		fmt.Sscanf(chainId, "0x%x", &tx.ChainID)
	}

	// Parse signature
	if v, ok := rawTx["v"].(string); ok {
		tx.V = v
	}
	if r, ok := rawTx["r"].(string); ok {
		tx.R = r
	}
	if s, ok := rawTx["s"].(string); ok {
		tx.S = s
	}

	// Calculate derived fields
	tx.GasPriceGwei = float64(tx.GasPrice) / 1e9
	tx.EstimatedConfirmTime = s.EstimateConfirmTime(ctx, tx)

	return tx, nil
}

// ParseTransactions parses multiple raw transactions
func (s *MempoolParserService) ParseTransactions(ctx context.Context, rawData [][]byte) ([]*domain.MempoolTransaction, error) {
	transactions := make([]*domain.MempoolTransaction, 0, len(rawData))

	for _, data := range rawData {
		tx, err := s.ParseTransaction(ctx, data)
		if err != nil {
			continue // Skip invalid transactions
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// ValidateTransaction validates a mempool transaction
func (s *MempoolParserService) ValidateTransaction(ctx context.Context, tx *domain.MempoolTransaction) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if tx.Hash == "" {
		return fmt.Errorf("transaction hash is empty")
	}
	if tx.From == "" {
		return fmt.Errorf("from address is empty")
	}
	if tx.Gas == 0 {
		return fmt.Errorf("gas limit is zero")
	}
	if tx.GasPrice == 0 && tx.MaxFeePerGas == 0 {
		return fmt.Errorf("gas price is zero")
	}
	return nil
}

// EstimateConfirmTime estimates the confirmation time based on gas price
func (s *MempoolParserService) EstimateConfirmTime(ctx context.Context, tx *domain.MempoolTransaction) float64 {
	// Simple estimation based on gas price tiers
	// In a real implementation, this would use historical data
	gasPriceGwei := float64(tx.GasPrice) / 1e9

	switch {
	case gasPriceGwei > 100:
		return 5.0 // seconds, very fast
	case gasPriceGwei > 50:
		return 15.0
	case gasPriceGwei > 20:
		return 30.0
	case gasPriceGwei > 10:
		return 60.0
	case gasPriceGwei > 5:
		return 120.0
	case gasPriceGwei > 1:
		return 300.0
	default:
		return 600.0 // 10 minutes, slow
	}
}

// MempoolService implements the MempoolService input port
type MempoolService struct {
	parser  ports.MempoolParser
	repo    ports.MempoolRepository
	mu      chan struct{} // simple mutex alternative
}

// NewMempoolService creates a new mempool service
func NewMempoolService(parser ports.MempoolParser, repo ports.MempoolRepository) *MempoolService {
	return &MempoolService{
		parser: parser,
		repo:   repo,
		mu:     make(chan struct{}, 1),
	}
}

// AddTransaction adds a transaction to the mempool
func (s *MempoolService) AddTransaction(ctx context.Context, rawData []byte) (*domain.MempoolTransaction, error) {
	// Parse the transaction
	tx, err := s.parser.ParseTransaction(ctx, rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	// Validate
	if err := s.parser.ValidateTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("transaction validation failed: %w", err)
	}

	// Add to repository
	if err := s.repo.AddTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to add transaction: %w", err)
	}

	return tx, nil
}

// RemoveTransaction removes a transaction from the mempool
func (s *MempoolService) RemoveTransaction(ctx context.Context, txHash string) error {
	return s.repo.RemoveTransaction(ctx, txHash)
}

// GetTransaction retrieves a transaction from the mempool
func (s *MempoolService) GetTransaction(ctx context.Context, txHash string) (*domain.MempoolTransaction, error) {
	return s.repo.GetTransaction(ctx, txHash)
}

// GetTransactions retrieves transactions matching the filter
func (s *MempoolService) GetTransactions(ctx context.Context, filter domain.MempoolFilter) ([]*domain.MempoolTransaction, error) {
	return s.repo.GetAllTransactions(ctx, filter)
}

// GetMempoolStats retrieves mempool statistics
func (s *MempoolService) GetMempoolStats(ctx context.Context) (*domain.MempoolStats, error) {
	return s.repo.GetMempoolStats(ctx)
}

// GetGasPricePercentiles calculates gas price percentiles
func (s *MempoolService) GetGasPricePercentiles(ctx context.Context, percentiles []float64) ([]float64, error) {
	// Get all transactions
	filter := domain.MempoolFilter{Limit: 10000}
	txs, err := s.repo.GetAllTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(txs) == 0 {
		return make([]float64, len(percentiles)), nil
	}

	// Extract gas prices
	gasPrices := make([]uint64, len(txs))
	for i, tx := range txs {
		gasPrices[i] = tx.GasPrice
	}
	sort.Slice(gasPrices, func(i, j int) bool { return gasPrices[i] < gasPrices[j] })

	// Calculate percentiles
	results := make([]float64, len(percentiles))
	for i, p := range percentiles {
		idx := int(float64(len(gasPrices)-1) * p / 100.0)
		if idx >= len(gasPrices) {
			idx = len(gasPrices) - 1
		}
		results[i] = float64(gasPrices[idx]) / 1e9 // Convert to Gwei
	}

	return results, nil
}

// BuildBlock builds a candidate block from the mempool
func (s *MempoolService) BuildBlock(ctx context.Context, gasLimit uint64) (*domain.PendingBlock, error) {
	filter := domain.MempoolFilter{
		SortByGasPrice: true,
		Limit:          int(gasLimit / 21000), // Rough estimate
	}

	txs, err := s.repo.GetAllTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	pending := &domain.PendingBlock{
		Transactions:  make([]*domain.MempoolTransaction, 0),
		BlockGasLimit: gasLimit,
		ProposedAt:    time.Now(),
	}

	var totalGas uint64
	for _, tx := range txs {
		if totalGas+tx.Gas > gasLimit {
			break
		}
		pending.Transactions = append(pending.Transactions, tx)
		totalGas += tx.Gas
	}

	pending.TotalGasUsed = totalGas
	for _, tx := range pending.Transactions {
		pending.TotalValue += tx.Value
	}

	return pending, nil
}

// MaintainMempool performs maintenance on the mempool
func (s *MempoolService) MaintainMempool(ctx context.Context, maxSize int, maxAge time.Duration) error {
	// Remove stale transactions
	cleaned, err := s.repo.CleanupStaleTransactions(ctx, maxAge)
	if err != nil {
		return err
	}

	// If still over max size, prune by gas price
	count, _ := s.repo.GetTransactionCount(ctx)
	if count > int64(maxSize) {
		// Get median gas price
		percentiles, _ := s.GetGasPricePercentiles(ctx, []float64{50})
		medianGasPrice := uint64(percentiles[0] * 1e9)

		pruned, err := s.repo.PruneByGasPrice(ctx, medianGasPrice)
		if err != nil {
			return err
		}

		_ = pruned
	}

	_ = cleaned
	return nil
}
