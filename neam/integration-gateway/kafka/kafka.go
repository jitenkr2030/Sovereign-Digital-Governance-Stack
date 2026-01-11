package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/Shopify/sarama"

	"neam-platform/integration-gateway/models"
)

// Producer represents a Kafka producer
type Producer interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, message *models.IntegrationMessage) error

	// PublishAsync publishes a message asynchronously
	PublishAsync(ctx context.Context, topic string, message *models.IntegrationMessage) error

	// PublishBatch publishes multiple messages
	PublishBatch(ctx context.Context, topic string, messages []*models.IntegrationMessage) error

	// Close closes the producer
	Close() error

	// Health checks producer health
	Health(ctx context.Context) error
}

// saramaProducer implements Producer using Sarama
type saramaProducer struct {
	producer sarama.SyncProducer
	topics   map[string]int32
	mu       sync.RWMutex
	closed   bool
}

// NewProducer creates a new Kafka producer
func NewProducer(ctx context.Context, brokers []string, tlsEnabled bool) (*saramaProducer, error) {
	config := sarama.NewConfig()

	// Set producer configuration
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Flush.Messages = 100
	config.Producer.Flush.Bytes = 32768

	// Configure TLS if enabled
	if tlsEnabled {
		tlsConfig, err := loadKafkaTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	// Create producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &saramaProducer{
		producer: producer,
		topics:   make(map[string]int32),
	}, nil
}

// Publish publishes a message synchronously
func (p *saramaProducer) Publish(ctx context.Context, topic string, message *models.IntegrationMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.MessageID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("data_type"), Value: []byte(message.DataType)},
			{Key: []byte("source"), Value: []byte(message.Source)},
			{Key: []byte("version"), Value: []byte(message.Version)},
		},
		Timestamp: time.Now(),
	}

	// Get partition
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)

	return nil
}

// PublishAsync publishes a message asynchronously
func (p *saramaProducer) PublishAsync(ctx context.Context, topic string, message *models.IntegrationMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	// Serialize message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Create Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.MessageID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("data_type"), Value: []byte(message.DataType)},
			{Key: []byte("source"), Value: []byte(message.Source)},
			{Key: []byte("version"), Value: []byte(message.Version)},
		},
		Timestamp: time.Now(),
	}

	// Send asynchronously
	go func() {
		_, _, err := p.producer.SendMessage(msg)
		if err != nil {
			log.Printf("Async message send failed: %v", err)
		}
	}()

	return nil
}

// PublishBatch publishes multiple messages
func (p *saramaProducer) PublishBatch(ctx context.Context, topic string, messages []*models.IntegrationMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	// Prepare batch
	messagesToSend := make([]*sarama.ProducerMessage, 0, len(messages))

	for _, message := range messages {
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(message.MessageID),
			Value: sarama.ByteEncoder(data),
			Headers: []sarama.RecordHeader{
				{Key: []byte("data_type"), Value: []byte(message.DataType)},
				{Key: []byte("source"), Value: []byte(message.Source)},
				{Key: []byte("version"), Value: []byte(message.Version)},
			},
			Timestamp: time.Now(),
		}

		messagesToSend = append(messagesToSend, msg)
	}

	// Send batch
	if err := p.producer.SendMessages(messagesToSend); err != nil {
		return fmt.Errorf("failed to send message batch: %w", err)
	}

	return nil
}

// Close closes the producer
func (p *saramaProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	return p.producer.Close()
}

// Health checks producer health
func (p *saramaProducer) Health(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("producer is closed")
	}

	// Try to get metadata to verify connectivity
	broker := p.producer.HighWaterMark()
	if broker == 0 {
		return fmt.Errorf("no brokers available")
	}

	return nil
}

// loadKafkaTLSConfig loads TLS configuration for Kafka
func loadKafkaTLSConfig() (*tls.Config, error) {
	// Load CA certificate
	caCert, err := ioutil.ReadFile("/certs/kafka-ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair("/certs/kafka-client.crt", "/certs/kafka-client.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// ConsumerGroup represents a Kafka consumer group
type ConsumerGroup interface {
	// Start starts consuming messages
	Start(ctx context.Context, topics []string, handler MessageHandler) error

	// Close closes the consumer
	Close() error
}

// MessageHandler handles incoming Kafka messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, topic string, message []byte) error
}

// saramaConsumerGroup implements ConsumerGroup using Sarama
type saramaConsumerGroup struct {
	consumerGroup sarama.ConsumerGroup
	groupID       string
	closed        bool
	mu            sync.RWMutex
}

// NewConsumerGroup creates a new Kafka consumer group
func NewConsumerGroup(ctx context.Context, brokers []string, groupID string, tlsEnabled bool) (*saramaConsumerGroup, error) {
	config := sarama.NewConfig()

	// Set consumer configuration
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	// Configure TLS if enabled
	if tlsEnabled {
		tlsConfig, err := loadKafkaTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &saramaConsumerGroup{
		consumerGroup: consumerGroup,
		groupID:       groupID,
	}, nil
}

// Start starts consuming messages
func (c *saramaConsumerGroup) Start(ctx context.Context, topics []string, handler MessageHandler) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("consumer is closed")
	}
	c.mu.RUnlock()

	// Create consumer group handler
	groupHandler := &consumerGroupHandler{
		handler:  handler,
		ready:    make(chan bool),
		ctx:      ctx,
	}

	// Start consuming
	for {
		c.mu.RLock()
		if c.closed {
			c.mu.RUnlock()
			break
		}
		c.mu.RUnlock()

		if err := c.consumerGroup.Consume(ctx, topics, groupHandler); err != nil {
			if err == sarama.ErrClosedConsumerGroup {
				return nil
			}
			return fmt.Errorf("error from consumer: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return nil
}

// Close closes the consumer
func (c *saramaConsumerGroup) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.consumerGroup.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler  MessageHandler
	ready    chan bool
	ctx      context.Context
}

// Setup is called when the consumer group session is starting
func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup is called when the consumer group session is ending
func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if err := h.handler.HandleMessage(h.ctx, message.Topic, message.Value); err != nil {
				log.Printf("Error processing message: %v", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// TopicManager provides Kafka topic management
type TopicManager struct {
	adminClient sarama.ClusterAdmin
}

// NewTopicManager creates a new topic manager
func NewTopicManager(ctx context.Context, brokers []string, tlsEnabled bool) (*TopicManager, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0

	if tlsEnabled {
		tlsConfig, err := loadKafkaTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	adminClient, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster admin: %w", err)
	}

	return &TopicManager{
		adminClient: adminClient,
	}, nil
}

// CreateTopic creates a new topic
func (t *TopicManager) CreateTopic(topic string, numPartitions int32, replicationFactor int16) error {
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}

	return t.adminClient.CreateTopic(topic, topicDetail, false)
}

// DeleteTopic deletes a topic
func (t *TopicManager) DeleteTopic(topic string) error {
	return t.adminClient.DeleteTopic(topic)
}

// ListTopics lists all topics
func (t *TopicManager) ListTopics() ([]string, error) {
	topics, err := t.adminClient.ListTopics()
	if err != nil {
		return nil, err
	}

	topicNames := make([]string, 0, len(topics))
	for name := range topics {
		topicNames = append(topicNames, name)
	}

	return topicNames, nil
}

// Close closes the topic manager
func (t *TopicManager) Close() error {
	return t.adminClient.Close()
}
