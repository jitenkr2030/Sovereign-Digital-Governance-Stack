package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// Config contains Kafka configuration
type Config struct {
	Brokers          string
	ConsumerGroup    string
	Version          string
	MaxOpenReqs      int
	DialTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	SessionTimeout   time.Duration
	HeartbeatInterval time.Duration
}

// Consumer represents a Kafka consumer
type Consumer interface {
	Consume(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

// MessageHandler is a function that handles a single message
type MessageHandler func(ctx context.Context, message []byte) error

// saramaConsumer implements Consumer using Sarama library
type saramaConsumer struct {
	client    sarama.ConsumerGroup
	logger    *zap.Logger
	ready     chan bool
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg Config, logger *zap.Logger) (Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = parseKafkaVersion(cfg.Version)
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Consumer.Group.Session.Timeout = cfg.SessionTimeout
	saramaConfig.Consumer.Heartbeat.Interval = cfg.HeartbeatInterval
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Net.MaxOpenRequests = cfg.MaxOpenReqs
	saramaConfig.Net.DialTimeout = cfg.DialTimeout
	saramaConfig.Net.ReadTimeout = cfg.ReadTimeout
	saramaConfig.Net.WriteTimeout = cfg.WriteTimeout

	client, err := sarama.NewConsumerGroup([]string{cfg.Brokers}, cfg.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &saramaConsumer{
		client: client,
		logger: logger,
		ready:  make(chan bool),
	}, nil
}

// Consume starts consuming messages from a topic
func (c *saramaConsumer) Consume(ctx context.Context, topic string, handler MessageHandler) error {
	consumerHandler := &consumerGroupHandler{
		handler:  handler,
		logger:   c.logger,
		ready:    c.ready,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.client.Consume(ctx, []string{topic}, consumerHandler); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					return nil
				}
				return fmt.Errorf("error from consumer: %w", err)
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
}

// Close closes the consumer
func (c *saramaConsumer) Close() error {
	return c.client.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler MessageHandler
	logger  *zap.Logger
	ready   chan bool
}

// Setup is run at the beginning of a new session
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition claim
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			// Process message with retry logic
			if err := h.handleWithRetry(session.Context(), message, h.handler); err != nil {
				h.logger.Error("Failed to process message",
					zap.Error(err),
					zap.Int32("partition", message.Partition),
					zap.Int64("offset", message.Offset))
			}

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleWithRetry handles a message with retry logic
func (h *consumerGroupHandler) handleWithRetry(ctx context.Context, message *sarama.ConsumerMessage, handler MessageHandler) error {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<attempt)
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
			time.Sleep(delay)
		}

		err := handler(ctx, message.Value)
		if err == nil {
			return nil
		}

		h.logger.Warn("Message processing failed, retrying",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", maxRetries))
	}

	return fmt.Errorf("message processing failed after %d attempts", maxRetries+1)
}

// Producer represents a Kafka producer
type Producer interface {
	Produce(ctx context.Context, topic string, value interface{}) error
	Close() error
}

// saramaProducer implements Producer using Sarama library
type saramaProducer struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg Config, logger *zap.Logger) (Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = parseKafkaVersion(cfg.Version)
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Flush.Frequency = 500 * time.Millisecond
	saramaConfig.Net.MaxOpenRequests = cfg.MaxOpenReqs
	saramaConfig.Net.DialTimeout = cfg.DialTimeout
	saramaConfig.Net.ReadTimeout = cfg.ReadTimeout
	saramaConfig.Net.WriteTimeout = cfg.WriteTimeout

	producer, err := sarama.NewSyncProducer([]string{cfg.Brokers}, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &saramaProducer{
		producer: producer,
		logger:   logger,
	}, nil
}

// Produce sends a message to a Kafka topic
func (p *saramaProducer) Produce(ctx context.Context, topic string, value interface{}) error {
	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(generateMessageKey(data)),
		Value: sarama.ByteEncoder(data),
	}

	// Add timestamp
	msg.Timestamp = time.Now()

	// Send message
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.logger.Debug("Message produced to Kafka",
		zap.String("topic", topic),
		zap.Int("size", len(data)))

	return nil
}

// ProduceWithKey sends a message with a specific key
func (p *saramaProducer) ProduceWithKey(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Close closes the producer
func (p *saramaProducer) Close() error {
	return p.producer.Close()
}

// generateMessageKey generates a key for message partitioning
func generateMessageKey(data []byte) string {
	// Use hash of message content for consistent partitioning
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:8])
}

// parseKafkaVersion parses Kafka version string to Sarama version
func parseKafkaVersion(version string) sarama.KafkaVersion {
	switch version {
	case "0.8.2":
		return sarama.V0_8_2_0
	case "0.9.0":
		return sarama.V0_9_0_0
	case "0.10.0":
		return sarama.V0_10_0_0
	case "0.10.1":
		return sarama.V0_10_1_0
	case "0.10.2":
		return sarama.V0_10_2_0
	case "0.11.0":
		return sarama.V0_11_0_0
	case "1.0":
		return sarama.V1_0_0_0
	case "1.1":
		return sarama.V1_1_0_0
	case "2.0":
		return sarama.V2_0_0_0
	case "2.1":
		return sarama.V2_1_0_0
	case "2.2":
		return sarama.V2_2_0_0
	case "2.3":
		return sarama.V2_3_0_0
	case "2.4":
		return sarama.V2_4_0_0
	case "2.5":
		return sarama.V2_5_0_0
	case "2.6":
		return sarama.V2_6_0_0
	case "2.7":
		return sarama.V2_7_0_0
	case "2.8":
		return sarama.V2_8_0_0
	default:
		return sarama.V2_8_0_0
	}
}

// Import for additional packages
import (
	"crypto/sha256"
	"sync"
)
