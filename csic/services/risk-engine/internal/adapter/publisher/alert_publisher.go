package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/risk-engine/internal/core/domain"
	"github.com/csic-platform/services/risk-engine/internal/core/ports"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaAlertPublisher implements AlertPublisher using Kafka
type KafkaAlertPublisher struct {
	writers     map[string]*kafka.Writer
	brokers     string
	topicPrefix string
	logger      *zap.Logger
}

// NewKafkaAlertPublisher creates a new KafkaAlertPublisher
func NewKafkaAlertPublisher(brokers, topicPrefix string, logger *zap.Logger) *KafkaAlertPublisher {
	return &KafkaAlertPublisher{
		writers:     make(map[string]*kafka.Writer),
		brokers:     brokers,
		topicPrefix: topicPrefix,
		logger:      logger,
	}
}

// getTopicName returns the full topic name for alerts
func (p *KafkaAlertPublisher) getTopicName() string {
	return fmt.Sprintf("%s_risk_alerts", p.topicPrefix)
}

// getWriter gets or creates a Kafka writer for alerts
func (p *KafkaAlertPublisher) getWriter() *kafka.Writer {
	topic := p.getTopicName()
	if w, exists := p.writers[topic]; exists {
		return w
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    10,
		BatchTimeout: 5 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	p.writers[topic] = w
	return w
}

// PublishAlert publishes a risk alert to the message broker
func (p *KafkaAlertPublisher) PublishAlert(ctx context.Context, alert *domain.RiskAlert) error {
	topic := p.getTopicName()

	value, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(alert.ID.String()),
		Value: value,
		Time:  time.Now(),
	}

	if err := p.getWriter().WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish alert",
			zap.String("topic", topic),
			zap.String("alert_id", alert.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to publish alert: %w", err)
	}

	p.logger.Debug("Published risk alert",
		zap.String("topic", topic),
		zap.String("alert_id", alert.ID.String()),
		zap.String("severity", string(alert.Severity)))

	return nil
}

// PublishBatch publishes multiple alerts in a batch
func (p *KafkaAlertPublisher) PublishBatch(ctx context.Context, alerts []*domain.RiskAlert) error {
	if len(alerts) == 0 {
		return nil
	}

	topic := p.getTopicName()
	messages := make([]kafka.Message, 0, len(alerts))

	for _, alert := range alerts {
		value, err := json.Marshal(alert)
		if err != nil {
			p.logger.Error("Failed to marshal alert in batch",
				zap.Error(err))
			continue
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(alert.ID.String()),
			Value: value,
			Time:  time.Now(),
		})
	}

	if err := p.getWriter().WriteMessages(ctx, messages...); err != nil {
		p.logger.Error("Failed to publish alert batch",
			zap.String("topic", topic),
			zap.Int("count", len(messages)),
			zap.Error(err))
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	p.logger.Info("Published risk alert batch",
		zap.String("topic", topic),
		zap.Int("count", len(messages)))

	return nil
}

// Close closes all Kafka writers
func (p *KafkaAlertPublisher) Close() error {
	p.logger.Info("Closing Kafka alert publishers")

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

// Ensure KafkaAlertPublisher implements AlertPublisher
var _ ports.AlertPublisher = (*KafkaAlertPublisher)(nil)
