package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"csic-platform/control-layer/internal/core/domain"
)

// KafkaProducer sends messages to Kafka
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
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
		topic:    "control-layer",
	}, nil
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// PublishEnforcement publishes an enforcement action to Kafka
func (p *KafkaProducer) PublishEnforcement(enforcement *domain.Enforcement) error {
	enforcementEvent := domain.EnforcementEvent{
		Type:        "enforcement",
		Enforcement: enforcement,
		Timestamp:   time.Now(),
	}

	data, err := json.Marshal(enforcementEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal enforcement event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.enforcement.actions", p.topic),
		Key:   sarama.StringEncoder(enforcement.ID.String()),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send enforcement message: %w", err)
	}

	return nil
}

// PublishIntervention publishes an intervention to Kafka
func (p *KafkaProducer) PublishIntervention(intervention *domain.Intervention) error {
	interventionEvent := domain.InterventionEvent{
		Type:         "intervention",
		Intervention: intervention,
		Timestamp:    time.Now(),
	}

	data, err := json.Marshal(interventionEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal intervention event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.interventions", p.topic),
		Key:   sarama.StringEncoder(intervention.ID.String()),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send intervention message: %w", err)
	}

	return nil
}

// PublishAlert publishes an alert to Kafka
func (p *KafkaProducer) PublishAlert(alert *domain.ControlAlert) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.alerts", p.topic),
		Key:   sarama.StringEncoder(alert.ID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send alert message: %w", err)
	}

	return nil
}

// PublishStateUpdate publishes a state update to Kafka
func (p *KafkaProducer) PublishStateUpdate(stateUpdate *domain.StateUpdate) error {
	data, err := json.Marshal(stateUpdate)
	if err != nil {
		return fmt.Errorf("failed to marshal state update: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.state.updates", p.topic),
		Key:   sarama.StringEncoder(stateUpdate.ServiceName),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send state update message: %w", err)
	}

	return nil
}

// PublishPolicyUpdate publishes a policy update notification
func (p *KafkaProducer) PublishPolicyUpdate(policyID string) error {
	event := map[string]string{
		"type":      "policy_update",
		"policy_id": policyID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal policy update event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s.policy.updates", p.topic),
		Key:   sarama.StringEncoder(policyID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send policy update message: %w", err)
	}

	return nil
}

// GetTopic returns the base topic name
func (p *KafkaProducer) GetTopic() string {
	return p.topic
}
