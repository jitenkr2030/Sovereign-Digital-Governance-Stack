package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/api-gateway/internal/core/domain"
	"github.com/csic-platform/services/api-gateway/internal/core/ports"
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

// PublishAlert publishes an alert event to Kafka
func (p *KafkaProducer) PublishAlert(ctx context.Context, alert *domain.Alert) error {
	event := map[string]interface{}{
		"event_type":   "ALERT_CREATED",
		"alert_id":     alert.ID,
		"alert_title":  alert.Title,
		"severity":     alert.Severity,
		"category":     alert.Category,
		"source":       alert.Source,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal alert event: %w", err)
	}

	return p.Publish(ctx, "csic.alerts", data)
}

// PublishTransaction publishes a transaction event to Kafka
func (p *KafkaProducer) PublishTransaction(ctx context.Context, tx interface{}) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	return p.Publish(ctx, "csic.transactions", data)
}

// PublishAuditLog publishes an audit log event to Kafka
func (p *KafkaProducer) PublishAuditLog(ctx context.Context, log *domain.AuditLog) error {
	event := map[string]interface{}{
		"event_type":    "AUDIT_LOG_CREATED",
		"log_id":        log.ID,
		"timestamp":     log.Timestamp,
		"user_id":       log.UserID,
		"username":      log.Username,
		"action":        log.Action,
		"resource_type": log.ResourceType,
		"status":        log.Status,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log event: %w", err)
	}

	return p.Publish(ctx, "csic.audit_logs", data)
}

// PublishComplianceReport publishes a compliance report event to Kafka
func (p *KafkaProducer) PublishComplianceReport(ctx context.Context, report *domain.ComplianceReport) error {
	event := map[string]interface{}{
		"event_type":  "COMPLIANCE_REPORT_GENERATED",
		"report_id":   report.ID,
		"entity_type": report.EntityType,
		"entity_id":   report.EntityID,
		"period":      report.Period,
		"status":      report.Status,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal compliance report event: %w", err)
	}

	return p.Publish(ctx, "csic.compliance_reports", data)
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

// KafkaConsumer implements a Kafka consumer for testing and development
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer creates a new Kafka consumer instance
func NewKafkaConsumer(brokers []string, topic, consumerGroup string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  consumerGroup,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{reader: reader}
}

// Consume consumes a message from Kafka
func (c *KafkaConsumer) Consume(ctx context.Context) ([]byte, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from Kafka: %w", err)
	}

	return msg.Value, nil
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// Ensure the KafkaProducer implements the MessageProducer interface
var _ ports.MessageProducer = (*KafkaProducer)(nil)
