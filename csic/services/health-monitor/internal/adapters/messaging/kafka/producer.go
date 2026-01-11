package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"

	"csic-platform/health-monitor/internal/core/domain"
)

// KafkaProducer sends messages to Kafka
type KafkaProducer struct {
	producer sarama.SyncProducer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
	}, nil
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// PublishHeartbeat publishes a heartbeat to Kafka
func (p *KafkaProducer) PublishHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error {
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "system.heartbeat",
		Key:   sarama.StringEncoder(heartbeat.ServiceName),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	return nil
}

// PublishAlert publishes an alert to Kafka
func (p *KafkaProducer) PublishAlert(ctx context.Context, alert *domain.Alert) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "health-monitor.alerts",
		Key:   sarama.StringEncoder(alert.ID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send alert: %w", err)
	}

	return nil
}

// PublishOutage publishes an outage to Kafka
func (p *KafkaProducer) PublishOutage(ctx context.Context, outage *domain.Outage) error {
	data, err := json.Marshal(outage)
	if err != nil {
		return fmt.Errorf("failed to marshal outage: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "health-monitor.outages",
		Key:   sarama.StringEncoder(outage.ID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send outage: %w", err)
	}

	return nil
}

// PublishHealthEvent publishes a health check event to Kafka
func (p *KafkaProducer) PublishHealthEvent(ctx context.Context, event *domain.HealthCheck) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal health event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "health-monitor.events",
		Key:   sarama.StringEncoder(event.ServiceName),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send health event: %w", err)
	}

	return nil
}
