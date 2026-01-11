package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/key-management/internal/core/domain"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer implements the MessageProducer interface using Kafka
type KafkaProducer struct {
	writers map[string]*kafka.Writer
	brokers []string
}

// NewKafkaProducer creates a new Kafka producer instance
func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
	}
}

// getWriter gets or creates a Kafka writer for a topic
func (p *KafkaProducer) getWriter(topic string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	p.writers[topic] = w
	return w
}

// PublishKeyOperation publishes a key operation event to Kafka
func (p *KafkaProducer) PublishKeyOperation(ctx context.Context, operation *domain.KeyOperation) error {
	event := map[string]interface{}{
		"event_type":  "KEY_OPERATION",
		"key_id":      operation.KeyID,
		"operation":   operation.Operation,
		"actor_id":    operation.ActorID,
		"service_id":  operation.ServiceID,
		"success":     operation.Success,
		"error":       operation.Error,
		"metadata":    operation.Metadata,
		"timestamp":   operation.Timestamp.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal key operation event: %w", err)
	}

	return p.Publish(ctx, "security.kms.audit", data)
}

// PublishKeyLifecycleEvent publishes a key lifecycle event to Kafka
func (p *KafkaProducer) PublishKeyLifecycleEvent(ctx context.Context, eventType string, key *domain.Key) error {
	event := map[string]interface{}{
		"event_type":       eventType,
		"key_id":           key.ID,
		"key_alias":        key.Alias,
		"algorithm":        key.Algorithm,
		"current_version":  key.CurrentVersion,
		"status":           key.Status,
		"created_at":       key.CreatedAt.Format(time.RFC3339),
	}

	if key.RotatedAt != nil {
		event["rotated_at"] = key.RotatedAt.Format(time.RFC3339)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal key lifecycle event: %w", err)
	}

	return p.Publish(ctx, "security.kms.audit", data)
}

// Publish publishes a message to a Kafka topic
func (p *KafkaProducer) Publish(ctx context.Context, topic string, message []byte) error {
	writer := p.getWriter(topic)

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(time.Now().Format(time.RFC3339)),
		Value: message,
		Time:  time.Now(),
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish message to Kafka: %w", err)
	}

	return nil
}

// Close closes all Kafka writers
func (p *KafkaProducer) Close() error {
	for _, w := range p.writers {
		if err := w.Close(); err != nil {
			return fmt.Errorf("failed to close Kafka writer: %w", err)
		}
	}
	return nil
}
