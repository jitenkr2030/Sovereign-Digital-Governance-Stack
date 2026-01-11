package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/domain"
	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaPublisher implements EventPublisher using Kafka
type KafkaPublisher struct {
	writers     map[string]*kafka.Writer
	brokers     string
	topicPrefix string
	logger      *zap.Logger
}

// NewKafkaPublisher creates a new KafkaPublisher
func NewKafkaPublisher(brokers, topicPrefix string, logger *zap.Logger) *KafkaPublisher {
	return &KafkaPublisher{
		writers:     make(map[string]*kafka.Writer),
		brokers:     brokers,
		topicPrefix: topicPrefix,
		logger:      logger,
	}
}

// getTopicName returns the full topic name for a given data type
func (p *KafkaPublisher) getTopicName(dataType string) string {
	return fmt.Sprintf("%s_%s", p.topicPrefix, dataType)
}

// getWriter gets or creates a Kafka writer for a specific topic
func (p *KafkaPublisher) getWriter(topic string) *kafka.Writer {
	if w, exists := p.writers[topic]; exists {
		return w
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	p.writers[topic] = w
	return w
}

// PublishMarketData publishes market data to the message broker
func (p *KafkaPublisher) PublishMarketData(ctx context.Context, data *domain.MarketData) error {
	topic := p.getTopicName("market_data")

	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal market data: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(data.Symbol),
		Value: value,
		Time:  time.Now(),
	}

	if err := p.getWriter(topic).WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish market data",
			zap.String("topic", topic),
			zap.String("symbol", data.Symbol),
			zap.Error(err))
		return fmt.Errorf("failed to publish market data: %w", err)
	}

	p.logger.Debug("Published market data",
		zap.String("topic", topic),
		zap.String("symbol", data.Symbol),
		zap.String("price", data.Price.String()))

	return nil
}

// PublishTrade publishes trade data to the message broker
func (p *KafkaPublisher) PublishTrade(ctx context.Context, trade *domain.Trade) error {
	topic := p.getTopicName("trades")

	value, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("failed to marshal trade: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(trade.TradeID),
		Value: value,
		Time:  time.Now(),
	}

	if err := p.getWriter(topic).WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish trade",
			zap.String("topic", topic),
			zap.String("trade_id", trade.TradeID),
			zap.Error(err))
		return fmt.Errorf("failed to publish trade: %w", err)
	}

	p.logger.Debug("Published trade",
		zap.String("topic", topic),
		zap.String("symbol", trade.Symbol),
		zap.String("trade_id", trade.TradeID))

	return nil
}

// PublishOrderBook publishes order book data to the message broker
func (p *KafkaPublisher) PublishOrderBook(ctx context.Context, orderbook *domain.OrderBook) error {
	topic := p.getTopicName("orderbook")

	value, err := json.Marshal(orderbook)
	if err != nil {
		return fmt.Errorf("failed to marshal orderbook: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(orderbook.Symbol),
		Value: value,
		Time:  time.Now(),
	}

	if err := p.getWriter(topic).WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish orderbook",
			zap.String("topic", topic),
			zap.String("symbol", orderbook.Symbol),
			zap.Error(err))
		return fmt.Errorf("failed to publish orderbook: %w", err)
	}

	p.logger.Debug("Published orderbook",
		zap.String("topic", topic),
		zap.String("symbol", orderbook.Symbol))

	return nil
}

// PublishBatch publishes multiple market data records in a batch
func (p *KafkaPublisher) PublishBatch(ctx context.Context, data []*domain.MarketData) error {
	if len(data) == 0 {
		return nil
	}

	topic := p.getTopicName("market_data")
	messages := make([]kafka.Message, 0, len(data))

	for _, d := range data {
		value, err := json.Marshal(d)
		if err != nil {
			p.logger.Error("Failed to marshal market data in batch",
				zap.Error(err))
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(d.Symbol),
			Value: value,
			Time:  time.Now(),
		})
	}

	if err := p.getWriter(topic).WriteMessages(ctx, messages...); err != nil {
		p.logger.Error("Failed to publish batch",
			zap.String("topic", topic),
			zap.Int("count", len(messages)),
			zap.Error(err))
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	p.logger.Info("Published market data batch",
		zap.String("topic", topic),
		zap.Int("count", len(messages)))

	return nil
}

// Close closes all Kafka writers
func (p *KafkaPublisher) Close() error {
	p.logger.Info("Closing Kafka publishers")

	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Error closing writer",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}

	p.writers = make(map[string]*kafka.Writer)
	return nil
}

// Ensure KafkaPublisher implements EventPublisher
var _ ports.EventPublisher = (*KafkaPublisher)(nil)
