package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/csic/transaction-monitoring/internal/repository"
	"github.com/csic/transaction-monitoring/internal/service/graph"
	"github.com/csic/transaction-monitoring/internal/service/risk"
	"github.com/csic/transaction-monitoring/internal/service/sanctions"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	cfg           *config.Config
	repo          *repository.Repository
	cache         *repository.CacheRepository
	riskSvc       *risk.RiskScoringService
	clusteringSvc *graph.ClusteringService
	sanctionsSvc  *sanctions.SanctionsService
	logger        *zap.Logger
	stopChan      chan struct{}
	wg            sync.WaitGroup
	readers       []*kafka.Reader
	isRunning     bool
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(
	cfg *config.Config,
	repo *repository.Repository,
	cache *repository.CacheRepository,
	riskSvc *risk.RiskScoringService,
	clusteringSvc *graph.ClusteringService,
	sanctionsSvc *sanctions.SanctionsService,
	logger *zap.Logger,
) *Consumer {
	return &Consumer{
		cfg:           cfg,
		repo:          repo,
		cache:         cache,
		riskSvc:       riskSvc,
		clusteringSvc: clusteringSvc,
		sanctionsSvc:  sanctionsSvc,
		logger:        logger,
		stopChan:      make(chan struct{}),
	}
}

// Start begins consuming messages from Kafka topics
func (c *Consumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.isRunning {
		c.mu.Unlock()
		return nil
	}
	c.isRunning = true
	c.mu.Unlock()

	c.logger.Info("Starting Kafka consumers")

	// Consumer for normalized transactions
	c.wg.Add(1)
	go c.consumeNormalizedTransactions(ctx)

	// Consumer for blockchain events
	c.wg.Add(1)
	go c.consumeBlockchainEvents(ctx)

	return nil
}

// Stop gracefully stops the consumer
func (c *Consumer) Stop() {
	c.mu.Lock()
	if !c.isRunning {
		c.mu.Unlock()
		return
	}
	c.isRunning = false
	c.mu.Unlock()

	c.logger.Info("Stopping Kafka consumers")
	close(c.stopChan)

	// Close all readers
	for _, reader := range c.readers {
		reader.Close()
	}

	c.wg.Wait()
}

func (c *Consumer) consumeNormalizedTransactions(ctx context.Context) {
	defer c.wg.Done()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.cfg.Kafka.Brokers,
		Topic:          "csic.tx.normalized",
		GroupID:        "tx-monitor-risk-processor",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 1 * time.Second,
	})
	c.readers = append(c.readers, reader)

	c.logger.Info("Consuming normalized transactions", zap.String("topic", "csic.tx.normalized"))

	for {
		select {
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Warn("Failed to fetch message", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			if err := c.processNormalizedTransaction(ctx, msg); err != nil {
				c.logger.Error("Failed to process transaction", zap.Error(err))
				// Continue processing even on error
			}

			// Commit message
			if err := reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Warn("Failed to commit message", zap.Error(err))
			}
		}
	}
}

func (c *Consumer) processNormalizedTransaction(ctx context.Context, msg kafka.Message) error {
	var tx models.NormalizedTransaction
	if err := json.Unmarshal(msg.Value, &tx); err != nil {
		return fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	c.logger.Debug("Processing transaction",
		zap.String("tx_hash", tx.TxHash),
		zap.String("network", string(tx.Network)),
		zap.Int("inputs", len(tx.Inputs)),
		zap.Int("outputs", len(tx.Outputs)))

	// 1. Update wallet metrics for all addresses
	for _, input := range tx.Inputs {
		if err := c.updateWalletMetrics(ctx, input.Address, tx.Network, input.Amount, "outgoing"); err != nil {
			c.logger.Warn("Failed to update input wallet metrics",
				zap.String("address", input.Address),
				zap.Error(err))
		}

		// Screen for sanctions
		screenResult, err := c.sanctionsSvc.ScreenAddress(ctx, input.Address, tx.Network)
		if err != nil {
			c.logger.Warn("Failed to screen address",
				zap.String("address", input.Address),
				zap.Error(err))
		} else if screenResult.IsSanctioned || screenResult.DirectLinks > 0 {
			// Create alert for sanctions match
			c.sanctionsSvc.CreateSanctionsAlert(ctx, screenResult)
		}
	}

	for _, output := range tx.Outputs {
		if err := c.updateWalletMetrics(ctx, output.Address, tx.Network, output.Amount, "incoming"); err != nil {
			c.logger.Warn("Failed to update output wallet metrics",
				zap.String("address", output.Address),
				zap.Error(err))
		}
	}

	// 2. Calculate risk scores for involved wallets
	for _, input := range tx.Inputs {
		if err := c.riskSvc.UpdateWalletRiskScore(ctx, input.Address, tx.Network); err != nil {
			c.logger.Warn("Failed to update risk score for input",
				zap.String("address", input.Address),
				zap.Error(err))
		}
	}

	for _, output := range tx.Outputs {
		if err := c.riskSvc.UpdateWalletRiskScore(ctx, output.Address, tx.Network); err != nil {
			c.logger.Warn("Failed to update risk score for output",
				zap.String("address", output.Address),
				zap.Error(err))
		}
	}

	// 3. Evaluate transaction risk
	riskScore, reasons := c.riskSvc.EvaluateTransactionRisk(ctx, &tx)
	if riskScore >= float64(c.cfg.RiskScoring.HighThreshold) {
		c.createTransactionRiskAlert(ctx, &tx, riskScore, reasons)
	}

	return nil
}

func (c *Consumer) updateWalletMetrics(ctx context.Context, address string, network models.Network, amount interface{}, txType string) error {
	// Get or create wallet
	wallet, err := c.repo.GetWalletByAddress(ctx, address, network)
	if err != nil {
		return err
	}

	if wallet == nil {
		wallet = &models.Wallet{
			ID:        generateID(),
			Address:   address,
			Network:   network,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			TxCount:   0,
		}
		if err := c.repo.CreateWallet(ctx, wallet); err != nil {
			return err
		}
	}

	// Update wallet
	wallet.LastSeen = time.Now()
	wallet.TxCount++
	// Update total received/sent based on txType

	return nil
}

func (c *Consumer) consumeBlockchainEvents(ctx context.Context) {
	defer c.wg.Done()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.cfg.Kafka.Brokers,
		Topic:          "csic.blockchain.raw",
		GroupID:        "tx-monitor-event-processor",
		MinBytes:       10e3,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 1 * time.Second,
	})
	c.readers = append(c.readers, reader)

	c.logger.Info("Consuming blockchain events", zap.String("topic", "csic.blockchain.raw"))

	for {
		select {
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Warn("Failed to fetch blockchain event", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			c.processBlockchainEvent(ctx, msg)

			if err := reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Warn("Failed to commit message", zap.Error(err))
			}
		}
	}
}

func (c *Consumer) processBlockchainEvent(ctx context.Context, msg kafka.Message) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Warn("Failed to unmarshal blockchain event", zap.Error(err))
		return
	}

	eventType, ok := event["type"].(string)
	if !ok {
		return
	}

	switch eventType {
	case "new_block":
		c.handleNewBlock(ctx, event)
	case "reorg":
		c.handleReorg(ctx, event)
	case "mempool_tx":
		c.handleMempoolTx(ctx, event)
	}
}

func (c *Consumer) handleNewBlock(ctx context.Context, event map[string]interface{}) {
	c.logger.Info("New block detected",
		zap.String("network", fmt.Sprintf("%v", event["network"])),
		zap.Int64("height", int64(event["height"].(float64))))
}

func (c *Consumer) handleReorg(ctx context.Context, event map[string]interface{}) {
	c.logger.Warn("Blockchain reorganization detected",
		zap.String("network", fmt.Sprintf("%v", event["network"])),
		zap.Int64("height", int64(event["height"].(float64))))
}

func (c *Consumer) handleMempoolTx(ctx context.Context, event map[string]interface{}) {
	c.logger.Debug("Mempool transaction detected")
}

func (c *Consumer) createTransactionRiskAlert(ctx context.Context, tx *models.NormalizedTransaction, score float64, reasons []string) {
	alert := models.NewAlert(
		models.AlertTypeTransactionAnomaly,
		mapScoreToSeverity(score),
		"transaction",
		tx.TxHash,
		fmt.Sprintf("High risk transaction detected: %s", tx.TxHash),
	)

	alert.RuleName = "transaction_risk_evaluation"
	alert.Score = score
	alert.Evidence = reasons

	if err := c.repo.CreateAlert(ctx, alert); err != nil {
		c.logger.Error("Failed to create transaction risk alert", zap.Error(err))
	}
}

func mapScoreToSeverity(score float64) models.AlertSeverity {
	switch {
	case score >= 80:
		return models.SeverityCritical
	case score >= 60:
		return models.SeverityHigh
	case score >= 40:
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
