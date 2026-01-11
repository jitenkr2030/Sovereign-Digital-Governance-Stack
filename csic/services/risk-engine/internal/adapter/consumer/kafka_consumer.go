package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/risk-engine/internal/core/ports"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaMarketDataConsumer implements MarketDataConsumer using Kafka
type KafkaMarketDataConsumer struct {
	reader        *kafka.Reader
	handler       func(data *domain.MarketData) error
	logger        *zap.Logger
	topicPrefix   string
	running       bool
	stopChan      chan struct{}
}

// NewKafkaMarketDataConsumer creates a new KafkaMarketDataConsumer
func NewKafkaMarketDataConsumer(
	brokers string,
	topicPrefix string,
	logger *zap.Logger,
) *KafkaMarketDataConsumer {
	return &KafkaMarketDataConsumer{
		logger:      logger,
		topicPrefix: topicPrefix,
		stopChan:    make(chan struct{}),
	}
}

// Start begins consuming market data messages
func (c *KafkaMarketDataConsumer) Start(ctx context.Context) error {
	if c.running {
		return nil
	}

	topic := fmt.Sprintf("%s_market_data", c.topicPrefix)

	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{},
		Topic:          topic,
		GroupID:        "risk-engine-consumer",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        1 * time.Second,
		CommitInterval: 1 * time.Second,
		StartOffset:    kafka.LastOffset,
	})

	c.running = true

	go c.consume(ctx)

	c.logger.Info("Kafka market data consumer started", zap.String("topic", topic))
	return nil
}

// consume is the internal consumer loop
func (c *KafkaMarketDataConsumer) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping consumer")
			return
		case <-c.stopChan:
			c.logger.Info("Stop signal received, stopping consumer")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}

			var data domain.MarketData
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				c.logger.Error("Failed to unmarshal market data", zap.Error(err))
				c.reader.CommitMessages(ctx, msg)
				continue
			}

			// Call the handler
			if c.handler != nil {
				if err := c.handler(&data); err != nil {
					c.logger.Error("Handler error",
						zap.String("symbol", data.Symbol),
						zap.Error(err))
				}
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("Failed to commit message", zap.Error(err))
			}
		}
	}
}

// Stop stops the consumer
func (c *KafkaMarketDataConsumer) Stop() error {
	if !c.running {
		return nil
	}

	close(c.stopChan)
	c.running = false

	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			return fmt.Errorf("failed to close reader: %w", err)
		}
	}

	c.logger.Info("Kafka market data consumer stopped")
	return nil
}

// SetHandler sets the handler function for processing market data
func (c *KafkaMarketDataConsumer) SetHandler(handler func(data *domain.MarketData) error) {
	c.handler = handler
}

// Ensure KafkaMarketDataConsumer implements MarketDataConsumer
var _ ports.MarketDataConsumer = (*KafkaMarketDataConsumer)(nil)
