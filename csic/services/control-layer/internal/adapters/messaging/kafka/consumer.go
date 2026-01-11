package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"csic-platform/control-layer/internal/core/domain"
)

// KafkaConsumer receives messages from Kafka
type KafkaConsumer struct {
	consumerGroup sarama.ConsumerGroup
	handler       MessageHandler
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// MessageHandler handles messages from Kafka
type MessageHandler interface {
	HandlePolicyUpdate(ctx context.Context, event map[string]interface{}) error
	HandleTelemetry(ctx context.Context, event domain.TelemetryEvent) error
	HandleStateUpdate(ctx context.Context, event domain.StateUpdate) error
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

// SetHandler sets the message handler
func (c *KafkaConsumer) SetHandler(handler MessageHandler) {
	c.handler = handler
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
	handler MessageHandler
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
	// Determine message type based on topic
	topic := message.Topic

	switch {
	case topic == "control-layer.policy.updates":
		return h.handlePolicyUpdate(ctx, message)
	case topic == "system.telemetry":
		return h.handleTelemetry(ctx, message)
	case topic == "control-layer.state.updates":
		return h.handleStateUpdate(ctx, message)
	default:
		// Try to handle as generic event
		return h.handleGenericEvent(ctx, message)
	}
}

// handlePolicyUpdate handles policy update messages
func (h *consumerGroupHandler) handlePolicyUpdate(ctx context.Context, message *sarama.ConsumerMessage) error {
	if h.handler == nil {
		return nil
	}

	var event map[string]interface{}
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal policy update: %w", err)
	}

	return h.handler.HandlePolicyUpdate(ctx, event)
}

// handleTelemetry handles telemetry messages
func (h *consumerGroupHandler) handleTelemetry(ctx context.Context, message *sarama.ConsumerMessage) error {
	if h.handler == nil {
		return nil
	}

	var event domain.TelemetryEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal telemetry event: %w", err)
	}

	return h.handler.HandleTelemetry(ctx, event)
}

// handleStateUpdate handles state update messages
func (h *consumerGroupHandler) handleStateUpdate(ctx context.Context, message *sarama.ConsumerMessage) error {
	if h.handler == nil {
		return nil
	}

	var event domain.StateUpdate
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal state update: %w", err)
	}

	return h.handler.HandleStateUpdate(ctx, event)
}

// handleGenericEvent handles generic events
func (h *consumerGroupHandler) handleGenericEvent(ctx context.Context, message *sarama.ConsumerMessage) error {
	log.Printf("Received message from topic %s: %s", message.Topic, string(message.Value))
	return nil
}

// KafkaConsumerGroup returns the underlying consumer group
func (c *KafkaConsumer) KafkaConsumerGroup() sarama.ConsumerGroup {
	return c.consumerGroup
}
