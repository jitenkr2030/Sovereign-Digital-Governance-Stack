package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/platform/blockchain/nodes/internal/config"
	"github.com/IBM/sarama"
)

// KafkaProducer interface for publishing messages
type KafkaProducer interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	Close() error
}

// kafkaProducer implements KafkaProducer using Sarama
type kafkaProducer struct {
	producer sarama.SyncProducer
	topics   config.KafkaTopicsConfig
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg config.KafkaConfig) (KafkaProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = cfg.Producer.Retries

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &kafkaProducer{
		producer: producer,
		topics:   cfg.Topics,
	}, nil
}

// Publish publishes a message to a Kafka topic
func (p *kafkaProducer) Publish(ctx context.Context, topic string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
		Key:   sarama.StringEncoder(time.Now().Format(time.RFC3339)),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka producer
func (p *kafkaProducer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// KafkaConsumer interface for consuming messages
type KafkaConsumer interface {
	Subscribe(ctx context.Context, topics []string, handler func(message interface{}) error) error
	Close() error
}

// kafkaConsumer implements KafkaConsumer using Sarama
type kafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	topics        []string
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg config.KafkaConfig) (*kafkaConsumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	topics := []string{cfg.Topics.NodeEvents, cfg.Topics.NodeMetrics}

	return &kafkaConsumer{
		consumerGroup: consumerGroup,
		topics:        topics,
	}, nil
}

// Subscribe subscribes to Kafka topics and processes messages
func (c *kafkaConsumer) Subscribe(ctx context.Context, topics []string, handler func(message interface{}) error) error {
	handlerImpl := func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
		for message := range claim.Messages() {
			var msg interface{}
			if err := json.Unmarshal(message.Value, &msg); err != nil {
				continue
			}

			if err := handler(msg); err != nil {
				continue
			}

			session.MarkMessage(message, "")
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.consumerGroup.Consume(ctx, topics, &consumerGroupHandler{handler: handlerImpl}); err != nil {
				return fmt.Errorf("error from consumer: %w", err)
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *kafkaConsumer) Close() error {
	if c.consumerGroup != nil {
		return c.consumerGroup.Close()
	}
	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	return h.handler(session, claim)
}
