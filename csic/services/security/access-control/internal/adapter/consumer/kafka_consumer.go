package kafka_consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/csic-platform/services/security/access-control/internal/core/domain"
	"github.com/csic-platform/services/security/access-control/internal/core/ports"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaConsumer handles consuming policy and role update events from Kafka
type KafkaConsumer struct {
	reader       *kafka.Reader
	policyHandler ports.PolicyEventHandler
	roleHandler   ports.RoleEventHandler
	logger       *zap.Logger
	stopCh       chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
}

// Config holds Kafka consumer configuration
type Config struct {
	Brokers        string
	Topic          string
	ConsumerGroup  string
	MinBytes       int
	MaxBytes       int
	MaxWait        time.Duration
	StartOffset    int64
}

// NewKafkaConsumer creates a new Kafka consumer for policy and role updates
func NewKafkaConsumer(cfg Config, policyHandler ports.PolicyEventHandler, roleHandler ports.RoleEventHandler, logger *zap.Logger) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Brokers},
		Topic:          cfg.Topic,
		GroupID:        cfg.ConsumerGroup,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		StartOffset:    cfg.StartOffset,
		CommitInterval: time.Second,
	})

	return &KafkaConsumer{
		reader:       reader,
		policyHandler: policyHandler,
		roleHandler:   roleHandler,
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
}

// Start begins consuming messages from Kafka
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("consumer is already running")
	}
	c.running = true
	c.mu.Unlock()

	c.wg.Add(1)
	go c.consume(ctx)

	c.logger.Info("Kafka consumer started", zap.String("topic", c.reader.Config().Topic))
	return nil
}

// Stop stops the Kafka consumer gracefully
func (c *KafkaConsumer) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	close(c.stopCh)
	c.wg.Wait()

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close kafka reader: %w", err)
	}

	c.logger.Info("Kafka consumer stopped")
	return nil
}

// consume is the main consumption loop
func (c *KafkaConsumer) consume(ctx context.Context) {
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("Context cancelled, stopping consumer")
			return
		case <-c.stopCh:
			c.logger.Debug("Stop signal received, stopping consumer")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Error("Failed to process message", zap.Error(err))
				// Continue processing other messages
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("Failed to commit message", zap.Error(err))
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	c.logger.Debug("Processing message",
		zap.String("topic", msg.Topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
	)

	var event domain.PolicyUpdate
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		// Try parsing as RoleUpdate
		var roleEvent domain.RoleUpdate
		if err := json.Unmarshal(msg.Value, &roleEvent); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}

		// Handle role update
		if c.roleHandler != nil {
			if err := c.roleHandler.HandleRoleUpdate(ctx, &roleEvent); err != nil {
				return fmt.Errorf("failed to handle role update: %w", err)
			}
		}
		return nil
	}

	// Handle policy update
	if c.policyHandler != nil {
		if err := c.policyHandler.HandlePolicyUpdate(ctx, &event); err != nil {
			return fmt.Errorf("failed to handle policy update: %w", err)
		}
	}

	return nil
}

// IsRunning returns whether the consumer is currently running
func (c *KafkaConsumer) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// KafkaPolicyEventHandler implements PolicyEventHandler for handling policy events
type KafkaPolicyEventHandler struct {
	policyRepo ports.PolicyRepository
	cache      PolicyCache
	logger     *zap.Logger
}

// PolicyCache defines the interface for caching policies
type PolicyCache interface {
	Set(policies []domain.Policy) error
	Remove(policyID string) error
	Clear() error
}

// NewKafkaPolicyEventHandler creates a new policy event handler
func NewKafkaPolicyEventHandler(policyRepo ports.PolicyRepository, cache PolicyCache, logger *zap.Logger) *KafkaPolicyEventHandler {
	return &KafkaPolicyEventHandler{
		policyRepo: policyRepo,
		cache:      cache,
		logger:     logger,
	}
}

// HandlePolicyUpdate handles policy create, update, and delete events
func (h *KafkaPolicyEventHandler) HandlePolicyUpdate(ctx context.Context, update *domain.PolicyUpdate) error {
	h.logger.Info("Handling policy update",
		zap.String("type", update.Type),
		zap.String("policy_id", update.PolicyID),
	)

	switch update.Type {
	case "create":
		if update.Policy == nil {
			return fmt.Errorf("policy is required for create event")
		}
		if err := h.policyRepo.Create(ctx, update.Policy); err != nil {
			return fmt.Errorf("failed to create policy: %w", err)
		}
		// Refresh cache
		policies, err := h.policyRepo.FindEnabled(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh cache: %w", err)
		}
		return h.cache.Set(policies)

	case "update":
		if update.Policy == nil {
			return fmt.Errorf("policy is required for update event")
		}
		if err := h.policyRepo.Update(ctx, update.Policy); err != nil {
			return fmt.Errorf("failed to update policy: %w", err)
		}
		// Refresh cache
		policies, err := h.policyRepo.FindEnabled(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh cache: %w", err)
		}
		return h.cache.Set(policies)

	case "delete":
		if err := h.policyRepo.Delete(ctx, update.PolicyID); err != nil {
			return fmt.Errorf("failed to delete policy: %w", err)
		}
		return h.cache.Remove(update.PolicyID)

	default:
		return fmt.Errorf("unknown policy update type: %s", update.Type)
	}
}

// Compile-time check that KafkaPolicyEventHandler implements PolicyEventHandler
var _ ports.PolicyEventHandler = (*KafkaPolicyEventHandler)(nil)

// KafkaRoleEventHandler implements RoleEventHandler for handling role events
type KafkaRoleEventHandler struct {
	roleRepo ports.RoleRepository
	logger   *zap.Logger
}

// NewKafkaRoleEventHandler creates a new role event handler
func NewKafkaRoleEventHandler(roleRepo ports.RoleRepository, logger *zap.Logger) *KafkaRoleEventHandler {
	return &KafkaRoleEventHandler{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

// HandleRoleUpdate handles role create, update, and delete events
func (h *KafkaRoleEventHandler) HandleRoleUpdate(ctx context.Context, update *domain.RoleUpdate) error {
	h.logger.Info("Handling role update",
		zap.String("type", update.Type),
		zap.String("role_id", update.RoleID),
	)

	switch update.Type {
	case "create":
		if update.Role == nil {
			return fmt.Errorf("role is required for create event")
		}
		if err := h.roleRepo.Create(ctx, update.Role); err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}

	case "update":
		if update.Role == nil {
			return fmt.Errorf("role is required for update event")
		}
		if err := h.roleRepo.Update(ctx, update.Role); err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}

	case "delete":
		if err := h.roleRepo.Delete(ctx, update.RoleID); err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}

	default:
		return fmt.Errorf("unknown role update type: %s", update.Type)
	}

	return nil
}

// Compile-time check that KafkaRoleEventHandler implements RoleEventHandler
var _ ports.RoleEventHandler = (*KafkaRoleEventHandler)(nil)
