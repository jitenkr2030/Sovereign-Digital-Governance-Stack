package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/services"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaEventBus implements the EventBus interface for Kafka
type KafkaEventBus struct {
	writer  *kafka.Writer
	alertTopic string
	logger  *zap.Logger
}

// NewKafkaEventBus creates a new KafkaEventBus
func NewKafkaEventBus(writer *kafka.Writer, alertTopic string, logger *zap.Logger) *KafkaEventBus {
	return &KafkaEventBus{
		writer:    writer,
		alertTopic: alertTopic,
		logger:    logger,
	}
}

// PublishAlert publishes a market alert to the alert topic
func (b *KafkaEventBus) PublishAlert(ctx context.Context, alert domain.MarketAlert) error {
	value, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	msg := kafka.Message{
		Topic: b.alertTopic,
		Key:   []byte(alert.ID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "alert_type", Value: []byte(alert.AlertType)},
			{Key: "severity", Value: []byte(alert.Severity)},
			{Key: "exchange_id", Value: []byte(alert.ExchangeID)},
			{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		},
	}

	if err := b.writer.WriteMessages(ctx, msg); err != nil {
		b.logger.Error("Failed to publish alert to Kafka",
			zap.String("alert_id", alert.ID),
			zap.String("topic", b.alertTopic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish alert: %w", err)
	}

	b.logger.Debug("Alert published to Kafka",
		zap.String("alert_id", alert.ID),
		zap.String("topic", b.alertTopic),
	)

	return nil
}

// PublishBatch publishes multiple alerts
func (b *KafkaEventBus) PublishBatch(ctx context.Context, alerts []domain.MarketAlert) error {
	if len(alerts) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(alerts))

	for _, alert := range alerts {
		value, err := json.Marshal(alert)
		if err != nil {
			b.logger.Warn("Failed to marshal alert in batch",
				zap.String("alert_id", alert.ID),
				zap.Error(err),
			)
			continue
		}

		msg := kafka.Message{
			Topic: b.alertTopic,
			Key:   []byte(alert.ID),
			Value: value,
			Headers: []kafka.Header{
				{Key: "alert_type", Value: []byte(alert.AlertType)},
				{Key: "severity", Value: []byte(alert.Severity)},
				{Key: "exchange_id", Value: []byte(alert.ExchangeID)},
				{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
			},
		}

		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return nil
	}

	if err := b.writer.WriteMessages(ctx, messages...); err != nil {
		b.logger.Error("Failed to publish alert batch to Kafka",
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish alert batch: %w", err)
	}

	b.logger.Info("Alert batch published to Kafka",
		zap.Int("count", len(messages)),
		zap.String("topic", b.alertTopic),
	)

	return nil
}

// StreamToRegulator streams regulatory reports to the reporting topic
func (b *KafkaEventBus) StreamToRegulator(ctx context.Context, report domain.RegulatoryReport) error {
	// This would publish to a different topic for regulatory reporting
	b.logger.Info("Regulatory report generated",
		zap.String("report_id", report.ID),
		zap.String("report_type", report.ReportType),
		zap.Time("period_start", report.PeriodStart),
		zap.Time("period_end", report.PeriodEnd),
	)

	// In a real implementation, this would publish to a reporting topic
	return nil
}

// PublishAlertEvent implements OversightEventPort
func (b *KafkaEventBus) PublishAlertEvent(ctx context.Context, alertType string, payload map[string]interface{}) error {
	// Convert the payload to a MarketAlert for publishing
	alert := domain.MarketAlert{
		ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		AlertType: domain.AlertType(alertType),
		Details: domain.AlertDetails{
			Description: payload["description"].(string),
			Metadata:    payload,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	return b.PublishAlert(ctx, alert)
}

// PublishAnomalyDetected implements OversightEventPort
func (b *KafkaEventBus) PublishAnomalyDetected(ctx context.Context, anomaly *domain.TradeAnomaly) error {
	b.logger.Info("Anomaly detected",
		zap.String("anomaly_id", anomaly.ID),
		zap.String("exchange_id", anomaly.ExchangeID),
		zap.String("abuse_type", string(anomaly.AbuseType)),
	)
	return nil
}

// PublishHealthAlert implements OversightEventPort
func (b *KafkaEventBus) PublishHealthAlert(ctx context.Context, exchangeID string, healthScore float64) error {
	b.logger.Info("Health alert",
		zap.String("exchange_id", exchangeID),
		zap.Float64("health_score", healthScore),
	)
	return nil
}

// KafkaThrottleCommandPublisher implements ThrottleCommandPublisher for Kafka
type KafkaThrottleCommandPublisher struct {
	writer       *kafka.Writer
	throttleTopic string
	logger       *zap.Logger
}

// NewKafkaThrottleCommandPublisher creates a new KafkaThrottleCommandPublisher
func NewKafkaThrottleCommandPublisher(writer *kafka.Writer, throttleTopic string, logger *zap.Logger) *KafkaThrottleCommandPublisher {
	return &KafkaThrottleCommandPublisher{
		writer:        writer,
		throttleTopic: throttleTopic,
		logger:        logger,
	}
}

// PublishThrottleCommand publishes a throttle command to the control topic
func (p *KafkaThrottleCommandPublisher) PublishThrottleCommand(ctx context.Context, cmd domain.ThrottleCommand) error {
	value, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal throttle command: %w", err)
	}

	msg := kafka.Message{
		Topic: p.throttleTopic,
		Key:   []byte(cmd.ID),
		Value: value,
		Headers: []kafka.Header{
			{Key: "command", Value: []byte(cmd.Command)},
			{Key: "exchange_id", Value: []byte(cmd.ExchangeID)},
			{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("Failed to publish throttle command to Kafka",
			zap.String("command_id", cmd.ID),
			zap.String("exchange_id", cmd.ExchangeID),
			zap.String("topic", p.throttleTopic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish throttle command: %w", err)
	}

	p.logger.Info("Throttle command published to Kafka",
		zap.String("command_id", cmd.ID),
		zap.String("exchange_id", cmd.ExchangeID),
		zap.String("command", string(cmd.Command)),
		zap.String("topic", p.throttleTopic),
	)

	return nil
}

// PublishBatch publishes multiple throttle commands
func (p *KafkaThrottleCommandPublisher) PublishBatch(ctx context.Context, cmds []domain.ThrottleCommand) error {
	if len(cmds) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(cmds))

	for _, cmd := range cmds {
		value, err := json.Marshal(cmd)
		if err != nil {
			p.logger.Warn("Failed to marshal throttle command in batch",
				zap.String("command_id", cmd.ID),
				zap.Error(err),
			)
			continue
		}

		msg := kafka.Message{
			Topic: p.throttleTopic,
			Key:   []byte(cmd.ID),
			Value: value,
			Headers: []kafka.Header{
				{Key: "command", Value: []byte(cmd.Command)},
				{Key: "exchange_id", Value: []byte(cmd.ExchangeID)},
				{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
			},
		}

		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return nil
	}

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		p.logger.Error("Failed to publish throttle command batch to Kafka",
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish throttle command batch: %w", err)
	}

	p.logger.Info("Throttle command batch published to Kafka",
		zap.Int("count", len(messages)),
		zap.String("topic", p.throttleTopic),
	)

	return nil
}

// KafkaConsumer reads messages from Kafka topics
type KafkaConsumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	handler MessageHandler
}

// MessageHandler defines the interface for processing consumed messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, topic string, key []byte, value []byte, headers []kafka.Header) error
}

// NewKafkaConsumer creates a new KafkaConsumer
func NewKafkaConsumer(reader *kafka.Reader, handler MessageHandler, logger *zap.Logger) *KafkaConsumer {
	return &KafkaConsumer{
		reader:  reader,
		logger:  logger,
		handler: handler,
	}
}

// Start begins consuming messages from the configured topic
func (c *KafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Kafka consumer shutting down")
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				c.logger.Error("Failed to read message from Kafka",
					zap.Error(err),
					zap.String("topic", c.reader.Config().Topic),
				)
				continue
			}

			if err := c.handler.HandleMessage(
				ctx,
				msg.Topic,
				msg.Key,
				msg.Value,
				msg.Headers,
			); err != nil {
				c.logger.Error("Failed to handle message",
					zap.Error(err),
					zap.ByteString("key", msg.Key),
				)
			}
		}
	}
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// TradeEventConsumer is a specific message handler for trade events
type TradeEventConsumer struct {
	Logger  *zap.Logger
	Service *services.IngestionService
}

// NewTradeEventConsumer creates a new TradeEventConsumer
func NewTradeEventConsumer(logger *zap.Logger, service *services.IngestionService) *TradeEventConsumer {
	return &TradeEventConsumer{
		Logger:  logger,
		Service: service,
	}
}

// HandleMessage processes a trade event message
func (c *TradeEventConsumer) HandleMessage(ctx context.Context, topic string, key []byte, value []byte, headers []kafka.Header) error {
	var trade domain.TradeEvent
	if err := json.Unmarshal(value, &trade); err != nil {
		c.Logger.Error("Failed to unmarshal trade event",
			zap.Error(err),
			zap.ByteString("key", key),
		)
		return err
	}

	c.Logger.Debug("Trade event consumed from Kafka",
		zap.String("trade_id", trade.TradeID),
		zap.String("exchange_id", trade.ExchangeID),
		zap.String("trading_pair", trade.TradingPair),
		zap.Float64("price", trade.Price),
		zap.Float64("volume", trade.Volume),
	)

	// Process the trade event through the ingestion service
	if c.Service != nil {
		if err := c.Service.ProcessTradeStream(ctx, trade); err != nil {
			c.Logger.Error("Failed to process trade event",
				zap.Error(err),
				zap.String("trade_id", trade.TradeID),
			)
			return err
		}
	}

	return nil
}

// AlertEventConsumer is a specific message handler for alert events
type AlertEventConsumer struct {
	logger *zap.Logger
}

// NewAlertEventConsumer creates a new AlertEventConsumer
func NewAlertEventConsumer(logger *zap.Logger) *AlertEventConsumer {
	return &AlertEventConsumer{
		logger: logger,
	}
}

// HandleMessage processes an alert event message
func (c *AlertEventConsumer) HandleMessage(ctx context.Context, topic string, key []byte, value []byte, headers []kafka.Header) error {
	var alert domain.MarketAlert
	if err := json.Unmarshal(value, &alert); err != nil {
		c.logger.Error("Failed to unmarshal alert event",
			zap.Error(err),
			zap.ByteString("key", key),
		)
		return err
	}

	c.logger.Info("Alert event consumed from Kafka",
		zap.String("alert_id", alert.ID),
		zap.String("alert_type", string(alert.AlertType)),
		zap.String("severity", string(alert.Severity)),
		zap.String("exchange_id", alert.ExchangeID),
	)

	// Process the alert event - this would trigger notifications, etc.
	return nil
}
