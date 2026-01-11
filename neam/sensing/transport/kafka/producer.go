package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/neamplatform/sensing/transport/config"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer represents a Kafka message producer
type Producer struct {
	cfg     config.KafkaConfig
	logger  *zap.Logger
	writer  *kafka.Writer
	stats   ProducerStats
	mu      sync.RWMutex
	healthy bool
}

// ProducerStats holds producer statistics
type ProducerStats struct {
	MessagesSent    int64
	MessagesFailed  int64
	BytesSent       int64
	LastMessageTime time.Time
	WriteLatency    time.Duration
	TotalLatency    time.Duration
	LatencySamples  int
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig, logger *zap.Logger) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers),
		Topic:        cfg.OutputTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		MaxAttempts:  3,
		BatchSize:    100,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &Producer{
		cfg:     cfg,
		logger:  logger,
		writer:  writer,
		healthy: true,
	}, nil
}

// Produce sends a message to Kafka
func (p *Producer) Produce(ctx context.Context, topic string, data interface{}) error {
	var messageData []byte
	var err error

	switch v := data.(type) {
	case []byte:
		messageData = v
	case string:
		messageData = []byte(v)
	default:
		messageData, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	start := time.Now()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%d", time.Now().UnixNano())),
		Value: messageData,
		Time:  time.Now(),
	})

	latency := time.Since(start)
	p.updateStats(err == nil, len(messageData), latency)

	if err != nil {
		p.mu.Lock()
		p.healthy = false
		p.mu.Unlock()
		p.logger.Error("Failed to produce message",
			zap.Error(err),
			zap.String("topic", topic))
		return fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Debug("Message produced",
		zap.String("topic", topic),
		zap.Int("bytes", len(messageData)))

	return nil
}

// ProduceWithKey sends a message with a specific key
func (p *Producer) ProduceWithKey(ctx context.Context, topic, key string, data interface{}) error {
	var messageData []byte
	var err error

	switch v := data.(type) {
	case []byte:
		messageData = v
	case string:
		messageData = []byte(v)
	default:
		messageData, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	start := time.Now()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: messageData,
		Time:  time.Now(),
	})

	latency := time.Since(start)
	p.updateStats(err == nil, len(messageData), latency)

	if err != nil {
		p.logger.Error("Failed to produce message with key",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("key", key))
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// updateStats updates producer statistics
func (p *Producer) updateStats(success bool, bytes int, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if success {
		p.stats.MessagesSent++
		p.stats.BytesSent += int64(bytes)
		p.stats.LastMessageTime = time.Now()
	} else {
		p.stats.MessagesFailed++
	}

	p.stats.TotalLatency += latency
	p.stats.LatencySamples++

	if p.stats.LatencySamples > 0 {
		p.stats.WriteLatency = p.stats.TotalLatency / time.Duration(p.stats.LatencySamples)
	}

	p.healthy = true
}

// HealthCheck verifies the producer is healthy
func (p *Producer) HealthCheck() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.healthy
}

// Stats returns producer statistics
func (p *Producer) Stats() ProducerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// Close closes the producer
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	p.logger.Info("Kafka producer closed")
	return nil
}

// GetWriter returns the underlying Kafka writer
func (p *Producer) GetWriter() *kafka.Writer {
	return p.writer
}

// Topics returns available topics
func (p *Producer) Topics(ctx context.Context) ([]string, error) {
	conn, err := kafka.DialContext(ctx, "tcp", p.cfg.Brokers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	topics := make(map[string]struct{})
	for _, partition := range partitions {
		topics[partition.Topic] = struct{}{}
	}

	result := make([]string, 0, len(topics))
	for topic := range topics {
		result = append(result, topic)
	}

	return result, nil
}

// EnsureTopic creates a topic if it doesn't exist
func (p *Producer) EnsureTopic(ctx context.Context, topic string, numPartitions int, replicationFactor int) error {
	conn, err := kafka.DialContext(ctx, "tcp", p.cfg.Brokers)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	p.logger.Info("Topic ensured",
		zap.String("topic", topic),
		zap.Int("partitions", numPartitions))

	return nil
}
