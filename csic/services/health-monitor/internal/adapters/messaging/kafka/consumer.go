package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
)

// KafkaConsumer receives messages from Kafka
type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	handler       ports.MessageHandler
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers, groupID string) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup([]string{brokers}, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &KafkaConsumer{
		consumerGroup: consumerGroup,
	}, nil
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	return c.consumerGroup.Close()
}

// SetMessageHandler sets the message handler
func (c *KafkaConsumer) SetMessageHandler(handler ports.MessageHandler) error {
	c.handler = handler
	return nil
}

// StartConsuming starts consuming from specified topics
func (c *KafkaConsumer) StartConsuming(ctx context.Context, topics []string) error {
	ctx, c.cancel = context.WithCancel(ctx)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}

			// Create consumer group handler
			handler := &consumerGroupHandler{
				handler:   c.handler,
				ctx:       ctx,
				ready:     make(chan bool),
			}

			// Start consuming
			if err := c.consumerGroup.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			// Check if context is cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	// Wait for handler to be ready
	<-handler.ready

	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler ports.MessageHandler
	ctx     context.Context
	ready   chan bool
}

// Setup is run at the beginning of a new session
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-h.ctx.Done():
			return nil
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if err := h.processMessage(h.ctx, message); err != nil {
				log.Printf("Error processing message: %v", err)
			}

			session.MarkMessage(message, "")
		}
	}
}

// processMessage processes a single message
func (h *consumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	topic := message.Topic

	switch topic {
	case "system.heartbeat":
		return h.handleHeartbeat(ctx, message)
	case "health-monitor.events":
		return h.handleHealthEvent(ctx, message)
	default:
		log.Printf("Received message from topic %s: %s", topic, string(message.Value))
		return nil
	}
}

// handleHeartbeat handles heartbeat messages
func (h *consumerGroupHandler) handleHeartbeat(ctx context.Context, message *sarama.ConsumerMessage) error {
	if h.handler == nil {
		return nil
	}

	var heartbeat domain.Heartbeat
	if err := json.Unmarshal(message.Value, &heartbeat); err != nil {
		return fmt.Errorf("failed to unmarshal heartbeat: %w", err)
	}

	return h.handler.HandleHeartbeat(ctx, &heartbeat)
}

// handleHealthEvent handles health check event messages
func (h *consumerGroupHandler) handleHealthEvent(ctx context.Context, message *sarama.ConsumerMessage) error {
	if h.handler == nil {
		return nil
	}

	var check domain.HealthCheck
	if err := json.Unmarshal(message.Value, &check); err != nil {
		return fmt.Errorf("failed to unmarshal health check: %w", err)
	}

	return h.handler.HandleHealthCheck(ctx, &check)
}
