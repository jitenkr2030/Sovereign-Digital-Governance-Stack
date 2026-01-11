// Kafka Queue Package - Message queue utilities for CSIC Platform
// Kafka producer and consumer implementations

package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// Config holds Kafka configuration
type Config struct {
	Brokers        []string
	ConsumerGroup  string
	TopicPrefix    string
	ClientID       string
	RetryMax       int
	RetryBackoff   time.Duration
	RequiredAcks   sarama.RequiredAcks
	ProducerFlush  time.Duration
	ConsumerRetries int
}

// Message represents a Kafka message
type Message struct {
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Headers   map[string]string      `json:"headers,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Topic     string                 `json:"topic"`
	Partition int32                  `json:"partition"`
	Offset    int64                  `json:"offset"`
}

// Handler is a function that processes a message
type Handler func(ctx context.Context, msg *Message) error

// Producer wraps a Kafka sync producer
type Producer struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
	config   Config
	mu       sync.Mutex
	closed   bool
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg Config, logger *zap.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.ClientID = cfg.ClientID
	config.Retry.Max = cfg.RetryMax
	config.Retry.Backoff = cfg.RetryBackoff
	config.Producer.RequiredAcks = cfg.RequiredAcks
	config.Producer.Flush.Frequency = cfg.ProducerFlush
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		logger:   logger,
		config:   cfg,
	}, nil
}

// Send sends a message to Kafka
func (p *Producer) Send(ctx context.Context, topic string, key string, value interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("failed to send message",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.logger.Debug("message sent to Kafka",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// SendWithHeaders sends a message with custom headers
func (p *Producer) SendWithHeaders(ctx context.Context, topic string, key string, value interface{}, headers map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message value: %w", err)
	}

	saramaHeaders := make([]sarama.RecordHeader, 0, len(headers))
	for k, v := range headers {
		saramaHeaders = append(saramaHeaders, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(data),
		Headers: saramaHeaders,
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.logger.Debug("message sent with headers",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	return p.producer.Close()
}

// Consumer wraps a Kafka consumer group
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	logger        *zap.Logger
	config        Config
	handlers      map[string]Handler
	ready         chan bool
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg Config, logger *zap.Logger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.ClientID = cfg.ClientID
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		logger:        logger,
		config:        cfg,
		handlers:      make(map[string]Handler),
		ready:         make(chan bool),
	}, nil
}

// RegisterHandler registers a handler for a topic
func (c *Consumer) RegisterHandler(topic string, handler Handler) {
	c.handlers[topic] = handler
}

// Start starts consuming messages from registered topics
func (c *Consumer) Start(ctx context.Context) error {
	ctx, c.cancel = context.WithCancel(ctx)

	topics := make([]string, 0, len(c.handlers))
	for topic := range c.handlers {
		topics = append(topics, topic)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}

			// Start consuming
			handler := &consumerGroupHandler{
				consumer: c,
				ctx:      ctx,
			}

			if err := c.consumerGroup.Consume(ctx, topics, handler); err != nil {
				c.logger.Error("consumer error",
					zap.Error(err))
			}

			// Check if we're still ready
			select {
			case <-c.ready:
			default:
				return
			}
		}
	}()

	// Wait for consumer to be ready
	<-c.ready
	c.logger.Info("consumer started",
		zap.Strings("topics", topics))

	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	return c.consumerGroup.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
	ctx      context.Context
}

// Setup is run at the beginning of a new session
func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	close(h.consumer.ready)
	return nil
}

// Cleanup is run at the end of a session
func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			// Deserialize message
			var value interface{}
			if err := json.Unmarshal(msg.Value, &value); err != nil {
				h.consumer.logger.Error("failed to unmarshal message",
					zap.Error(err))
				continue
			}

			// Create message object
			kafkaMsg := &Message{
				Key:       string(msg.Key),
				Value:     value,
				Timestamp: msg.Timestamp,
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
			}

			// Extract headers
			headers := make(map[string]string)
			for _, header := range msg.Headers {
				if header != nil {
					headers[string(header.Key)] = string(header.Value)
				}
			}
			kafkaMsg.Headers = headers

			// Get handler and process message
			handler := h.consumer.handlers[msg.Topic]
			if handler == nil {
				h.consumer.logger.Warn("no handler for topic",
					zap.String("topic", msg.Topic))
				continue
			}

			if err := handler(h.ctx, kafkaMsg); err != nil {
				h.consumer.logger.Error("failed to process message",
					zap.String("topic", msg.Topic),
					zap.Error(err))
				continue
			}

			// Mark message as processed
			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// TopicManager manages Kafka topics
type TopicManager struct {
	admin sarama.ClusterAdmin
	logger *zap.Logger
}

// NewTopicManager creates a new topic manager
func NewTopicManager(cfg Config, logger *zap.Logger) (*TopicManager, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0

	admin, err := sarama.NewClusterAdmin(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster admin: %w", err)
	}

	return &TopicManager{
		admin:  admin,
		logger: logger,
	}, nil
}

// CreateTopic creates a new topic with the specified configuration
func (tm *TopicManager) CreateTopic(topic string, partitions int32, replicationFactor int16) error {
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	err := tm.admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topic, err)
	}

	tm.logger.Info("topic created",
		zap.String("topic", topic),
		zap.Int32("partitions", partitions),
		zap.Int16("replication_factor", replicationFactor))

	return nil
}

// ListTopics lists all topics
func (tm *TopicManager) ListTopics() ([]string, error) {
	topics, err := tm.admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	topicNames := make([]string, 0, len(topics))
	for topic := range topics {
		topicNames = append(topicNames, topic)
	}

	return topicNames, nil
}

// Close closes the topic manager
func (tm *TopicManager) Close() error {
	return tm.admin.Close()
}

// Import required packages
import "sync"
