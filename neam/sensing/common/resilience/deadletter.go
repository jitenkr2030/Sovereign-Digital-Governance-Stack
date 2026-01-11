package resilience

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// DLQConfig holds configuration for dead letter queue
type DLQConfig struct {
	// Enabled controls whether DLQ is active
	Enabled bool `yaml:"enabled"`

	// TopicSuffix is the suffix added to create DLQ topic names
	TopicSuffix string `yaml:"topic_suffix"`

	// RetryCount is the number of retry attempts before sending to DLQ
	RetryCount int `yaml:"retry_count"`

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration `yaml:"retry_delay"`

	// MaxMessageSize is the maximum size of messages in DLQ
	MaxMessageSize int `yaml:"max_message_size"`

	// EnableSchemaValidation enables schema validation at ingress
	EnableSchemaValidation bool `yaml:"enable_schema_validation"`

	// KafkaBrokers is the list of Kafka brokers
	KafkaBrokers []string `yaml:"brokers"`

	// ConsumerGroup is the consumer group for DLQ processing
	ConsumerGroup string `yaml:"consumer_group"`
}

// DLQMessage represents a message in the dead letter queue
type DLQMessage struct {
	// ID is a unique identifier for the message
	ID string `json:"id"`

	// CorrelationID links this message to the original request
	CorrelationID string `json:"correlation_id"`

	// Timestamp when the message was created
	Timestamp time.Time `json:"timestamp"`

	// SourceAdapter indicates which adapter produced this message
	SourceAdapter string `json:"source_adapter"`

	// SourceTopic is the original Kafka topic
	SourceTopic string `json:"source_topic"`

	// ErrorReason categorizes why the message failed
	ErrorReason ErrorReason `json:"error_reason"`

	// ErrorCode is a specific error code
	ErrorCode string `json:"error_code"`

	// ErrorMessage is a human-readable error description
	ErrorMessage string `json:"error_message"`

	// OriginalPayload is the raw message that failed
	OriginalPayload json.RawMessage `json:"original_payload"`

	// Metadata contains additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// RetryCount is the number of retry attempts made
	RetryCount int `json:"retry_count"`

	// FirstFailureTime when the first failure occurred
	FirstFailureTime time.Time `json:"first_failure_time"`

	// LastFailureTime when the most recent failure occurred
	LastFailureTime time.Time `json:"last_failure_time"`
}

// ErrorReason categorizes DLQ entries
type ErrorReason string

const (
	// ErrorReasonSchemaValidation indicates JSON schema validation failure
	ErrorReasonSchemaValidation ErrorReason = "SCHEMA_VALIDATION"

	// ErrorReasonProcessing indicates processing/fatal error
	ErrorReasonProcessing ErrorReason = "PROCESSING"

	// ErrorReasonTimeout indicates timeout during processing
	ErrorReasonTimeout ErrorReason = "TIMEOUT"

	// ErrorReasonCircuitBreaker indicates circuit breaker was open
	ErrorReasonCircuitBreaker ErrorReason = "CIRCUIT_BREAKER"

	// ErrorReasonNetwork indicates network connectivity issues
	ErrorReasonNetwork ErrorReason = "NETWORK"

	// ErrorReasonValidation indicates business logic validation failure
	ErrorReasonValidation ErrorReason = "VALIDATION"

	// ErrorReasonUnknown indicates unknown error
	ErrorReasonUnknown ErrorReason = "UNKNOWN"
)

// DLQProducer produces messages to the dead letter queue
type DLQProducer struct {
	writers     map[string]*kafka.Writer
	config      DLQConfig
	metrics     *MetricsCollector
	logger      *zap.Logger
	serializer  *MessageSerializer
}

// NewDLQProducer creates a new DLQ producer
func NewDLQProducer(
	config DLQConfig,
	metrics *MetricsCollector,
	logger *zap.Logger,
) *DLQProducer {
	return &DLQProducer{
		writers:    make(map[string]*kafka.Writer),
		config:     config,
		metrics:    metrics,
		logger:     logger,
		serializer: NewMessageSerializer(logger),
	}
}

// getTopicName returns the full DLQ topic name for an adapter
func (p *DLQProducer) getTopicName(adapterName string) string {
	return fmt.Sprintf("%s%s", adapterName, p.config.TopicSuffix)
}

// getWriter returns or creates a Kafka writer for a DLQ topic
func (p *DLQProducer) getWriter(adapterName string) *kafka.Writer {
	p.logger.Debug("Getting DLQ writer",
		zap.String("adapter", adapterName),
		zap.String("topic", p.getTopicName(adapterName)))

	topic := p.getTopicName(adapterName)

	if writer, exists := p.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.config.KafkaBrokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  3,
		BatchTimeout: 10 * time.Millisecond,
		Async:        false, // Synchronous for reliability
	}

	p.writers[topic] = writer
	return writer
}

// Publish sends a message to the dead letter queue
func (p *DLQProducer) Publish(
	ctx context.Context,
	adapterName string,
	originalTopic string,
	payload []byte,
	reason ErrorReason,
	err error,
	metadata map[string]interface{},
) (*DLQMessage, error) {
	if !p.config.Enabled {
		p.logger.Warn("DLQ is disabled, message will be dropped",
			zap.String("adapter", adapterName),
			zap.String("reason", string(reason)))
		return nil, nil
	}

	now := time.Now()
	dlqMsg := &DLQMessage{
		ID:                uuid.New().String(),
		CorrelationID:     uuid.New().String(),
		Timestamp:         now,
		SourceAdapter:     adapterName,
		SourceTopic:       originalTopic,
		ErrorReason:       reason,
		ErrorCode:         getErrorCode(err),
		ErrorMessage:      err.Error(),
		OriginalPayload:   payload,
		Metadata:          metadata,
		RetryCount:        0,
		FirstFailureTime:  now,
		LastFailureTime:   now,
	}

	// Serialize the message
	value, serErr := p.serializer.Serialize(dlqMsg)
	if serErr != nil {
		p.logger.Error("Failed to serialize DLQ message",
			zap.Error(serErr),
			zap.String("adapter", adapterName))
		return nil, serErr
	}

	// Write to Kafka
	writer := p.getWriter(adapterName)
	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(dlqMsg.CorrelationID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "dlq_id", Value: []byte(dlqMsg.ID)},
			{Key: "correlation_id", Value: []byte(dlqMsg.CorrelationID)},
			{Key: "adapter", Value: []byte(adapterName)},
			{Key: "error_reason", Value: []byte(string(reason))},
			{Key: "error_code", Value: []byte(dlqMsg.ErrorCode)},
			{Key: "timestamp", Value: []byte(now.Format(time.RFC3339))},
		},
		Time: now,
	})

	if err != nil {
		p.logger.Error("Failed to write DLQ message",
			zap.Error(err),
			zap.String("adapter", adapterName),
			zap.String("topic", p.getTopicName(adapterName)))
		return nil, err
	}

	p.logger.Info("Message sent to DLQ",
		zap.String("id", dlqMsg.ID),
		zap.String("correlation_id", dlqMsg.CorrelationID),
		zap.String("adapter", adapterName),
		zap.String("reason", string(reason)),
		zap.String("topic", p.getTopicName(adapterName)))

	if p.metrics != nil {
		p.metrics.RecordDLQMessage(adapterName, string(reason))
	}

	return dlqMsg, nil
}

// Close closes all DLQ writers
func (p *DLQProducer) Close() error {
	p.logger.Info("Closing DLQ producers")

	var lastErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Error closing DLQ writer",
				zap.Error(err),
				zap.String("topic", topic))
			lastErr = err
		}
		delete(p.writers, topic)
	}

	return lastErr
}

// DLQConsumer consumes messages from the dead letter queue
type DLQConsumer struct {
	readers      map[string]*kafka.Reader
	config       DLQConfig
	metrics      *MetricsCollector
	logger       *zap.Logger
	deserializer *MessageSerializer
}

// NewDLQConsumer creates a new DLQ consumer
func NewDLQConsumer(
	config DLQConfig,
	metrics *MetricsCollector,
	logger *zap.Logger,
) *DLQConsumer {
	return &DLQConsumer{
		readers:     make(map[string]*kafka.Reader),
		config:      config,
		metrics:     metrics,
		logger:      logger,
		deserializer: NewMessageSerializer(logger),
	}
}

// getTopicName returns the full DLQ topic name for an adapter
func (c *DLQConsumer) getTopicName(adapterName string) string {
	return fmt.Sprintf("%s%s", adapterName, c.config.TopicSuffix)
}

// getReader returns or creates a Kafka reader for a DLQ topic
func (c *DLQConsumer) getReader(adapterName string) *kafka.Reader {
	topic := c.getTopicName(adapterName)

	if reader, exists := c.readers[topic]; exists {
		return reader
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.config.KafkaBrokers,
		Topic:          topic,
		GroupID:        c.config.ConsumerGroup,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	c.readers[topic] = reader
	return reader
}

// Consume reads a single message from the DLQ
func (c *DLQConsumer) Consume(ctx context.Context, adapterName string) (*DLQMessage, error) {
	reader := c.getReader(adapterName)

	msg, err := reader.FetchMessage(ctx)
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return nil, err
		}
		c.logger.Error("Error fetching DLQ message",
			zap.Error(err),
			zap.String("adapter", adapterName))
		return nil, err
	}

	// Deserialize the message
	dlqMsg, err := c.deserializer.DeserializeDLQ(msg.Value)
	if err != nil {
		c.logger.Error("Failed to deserialize DLQ message",
			zap.Error(err),
			zap.String("adapter", adapterName))
		return nil, err
	}

	// Parse headers
	for _, h := range msg.Headers {
		switch h.Key {
		case "retry_count":
			dlqMsg.RetryCount = int(parseInt64(h.Value))
		}
	}

	return dlqMsg, nil
}

// Commit commits the offset of a consumed message
func (c *DLQConsumer) Commit(ctx context.Context, adapterName string, msg kafka.Message) error {
	reader := c.getReader(adapterName)
	return reader.CommitMessages(ctx, msg)
}

// Close closes all DLQ readers
func (c *DLQConsumer) Close() error {
	c.logger.Info("Closing DLQ consumers")

	var lastErr error
	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.Error("Error closing DLQ reader",
				zap.Error(err),
				zap.String("topic", topic))
			lastErr = err
		}
		delete(c.readers, topic)
	}

	return lastErr
}

// DLQAdmin provides administrative operations for DLQ
type DLQAdmin struct {
	producer    *DLQProducer
	consumer    *DLQConsumer
	metrics     *MetricsCollector
	logger      *zap.Logger
}

// NewDLQAdmin creates a new DLQ admin interface
func NewDLQAdmin(
	producer *DLQProducer,
	consumer *DLQConsumer,
	metrics *MetricsCollector,
	logger *zap.Logger,
) *DLQAdmin {
	return &DLQAdmin{
		producer: producer,
		consumer: consumer,
		metrics:  metrics,
		logger:   logger,
	}
}

// ListMessages lists messages in the DLQ
func (a *DLQAdmin) ListMessages(ctx context.Context, adapterName string, limit, offset int) ([]*DLQMessage, error) {
	messages := make([]*DLQMessage, 0, limit)

	consumer := a.consumer.getReader(adapterName)

	for i := 0; i < limit; i++ {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			a.logger.Error("Error fetching DLQ message", zap.Error(err))
			continue
		}

		dlqMsg, err := a.consumer.deserializer.DeserializeDLQ(msg.Value)
		if err != nil {
			a.logger.Error("Error deserializing DLQ message", zap.Error(err))
			continue
		}

		messages = append(messages, dlqMsg)

		// Commit the message as we read it
		if err := consumer.CommitMessages(ctx, msg); err != nil {
			a.logger.Error("Error committing DLQ message", zap.Error(err))
		}
	}

	return messages, nil
}

// GetMessage retrieves a specific message from the DLQ
func (a *DLQAdmin) GetMessage(ctx context.Context, adapterName, messageID string) (*DLQMessage, error) {
	consumer := a.consumer.getReader(adapterName)

	for {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			return nil, err
		}

		dlqMsg, err := a.consumer.deserializer.DeserializeDLQ(msg.Value)
		if err != nil {
			continue
		}

		if dlqMsg.ID == messageID {
			// Commit after reading
			consumer.CommitMessages(ctx, msg)
			return dlqMsg, nil
		}

		// Commit processed messages
		consumer.CommitMessages(ctx, msg)
	}
}

// ReprocessMessage moves a message from DLQ back to the main topic
func (a *DLQAdmin) ReprocessMessage(ctx context.Context, adapterName string, messageID string, originalTopic string) error {
	a.logger.Info("Reprocessing DLQ message",
		zap.String("adapter", adapterName),
		zap.String("message_id", messageID),
		zap.String("original_topic", originalTopic))

	// Get the message from DLQ
	msg, err := a.GetMessage(ctx, adapterName, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Create a new producer for the original topic
	producer := &kafka.Writer{
		Addr:     kafka.TCP(a.producer.config.KafkaBrokers...),
		Topic:    originalTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer producer.Close()

	// Send to original topic
	err = producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.CorrelationID),
		Value: msg.OriginalPayload,
		Headers: []kafka.Header{
			{Key: "reprocessed", Value: []byte("true")},
			{Key: "original_dlq_id", Value: []byte(msg.ID)},
			{Key: "reprocess_time", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	})

	if err != nil {
		a.logger.Error("Failed to reprocess message",
			zap.Error(err),
			zap.String("adapter", adapterName),
			zap.String("message_id", messageID))
		return err
	}

	a.logger.Info("Message successfully reprocessed",
		zap.String("adapter", adapterName),
		zap.String("message_id", messageID),
		zap.String("topic", originalTopic))

	if a.metrics != nil {
		a.metrics.RecordDLQReprocess(adapterName, "success")
	}

	return nil
}

// DiscardMessage removes a message from the DLQ
func (a *DLQAdmin) DiscardMessage(ctx context.Context, adapterName, messageID string) error {
	a.logger.Info("Discarding DLQ message",
		zap.String("adapter", adapterName),
		zap.String("message_id", messageID))

	consumer := a.consumer.getReader(adapterName)

	for {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch message: %w", err)
		}

		dlqMsg, err := a.consumer.deserializer.DeserializeDLQ(msg.Value)
		if err != nil {
			consumer.CommitMessages(ctx, msg)
			continue
		}

		if dlqMsg.ID == messageID {
			// Commit to mark as processed (discarded)
			consumer.CommitMessages(ctx, msg)
			a.logger.Info("Message discarded",
				zap.String("adapter", adapterName),
				zap.String("message_id", messageID))
			return nil
		}

		consumer.CommitMessages(ctx, msg)
	}
}

// CountMessages returns the count of messages in the DLQ
func (a *DLQAdmin) CountMessages(ctx context.Context, adapterName string) (int64, error) {
	// This is a simple implementation - in production, you'd use Kafka admin API
	// or a separate count tracking mechanism
	count := int64(0)
	consumer := a.consumer.getReader(adapterName)

	for {
		_, err := consumer.FetchMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded || err == context.Canceled {
				break
			}
			a.logger.Error("Error counting DLQ messages", zap.Error(err))
			continue
		}
		count++
	}

	return count, nil
}

// Helper functions

func getErrorCode(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	switch {
	case contains(errStr, "timeout"):
		return "TIMEOUT"
	case contains(errStr, "connection"):
		return "CONNECTION_ERROR"
	case contains(errStr, "schema"):
		return "SCHEMA_ERROR"
	case contains(errStr, "validation"):
		return "VALIDATION_ERROR"
	case contains(errStr, "circuit"):
		return "CIRCUIT_BREAKER"
	default:
		return "PROCESSING_ERROR"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func parseInt64(b []byte) int64 {
	var result int64
	for _, c := range b {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}
