package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/google/uuid"
	"csic-platform/service/reporting/internal/domain"
)

// KafkaConfig holds the Kafka configuration
type KafkaConfig struct {
	Brokers       []string
	ConsumerGroup string
	TopicPrefix   string
}

// Producer implements the KafkaProducer interface
type Producer struct {
	writer *kafka.Writer
	config KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(config KafkaConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{
		writer: writer,
		config: config,
	}
}

// PublishReportGenerated publishes a report generated event
func (p *Producer) PublishReportGenerated(ctx context.Context, report *domain.GeneratedReport) error {
	topic := fmt.Sprintf("%s_report_generated", p.config.TopicPrefix)

	event := ReportGeneratedEvent{
		EventID:       uuid.New().String(),
		EventType:     "report.generated",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		ReportID:      report.ID.String(),
		TemplateID:    report.TemplateID.String(),
		TemplateName:  report.TemplateName,
		ReportType:    string(report.ReportType),
		RegulatorID:   report.RegulatorID.String(),
		Status:        string(report.Status),
		Title:         report.Title,
		PeriodStart:   report.PeriodStart.Format(time.RFC3339),
		PeriodEnd:     report.PeriodEnd.Format(time.RFC3339),
		GeneratedBy:   report.GeneratedBy,
		FileCount:     len(report.Files),
		TotalSize:     report.TotalSize,
	}

	return p.publish(ctx, topic, report.ID.String(), event)
}

// PublishReportFailed publishes a report failed event
func (p *Producer) PublishReportFailed(ctx context.Context, reportID uuid.UUID, reason string) error {
	topic := fmt.Sprintf("%s_report_failed", p.config.TopicPrefix)

	event := ReportFailedEvent{
		EventID:     uuid.New().String(),
		EventType:   "report.failed",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		ReportID:    reportID.String(),
		Reason:      reason,
		RetryPolicy: "exponential_backoff",
	}

	return p.publish(ctx, topic, reportID.String(), event)
}

// PublishReportScheduled publishes a report scheduled event
func (p *Producer) PublishReportScheduled(ctx context.Context, schedule *domain.ReportSchedule) error {
	topic := fmt.Sprintf("%s_report_scheduled", p.config.TopicPrefix)

	event := ReportScheduledEvent{
		EventID:        uuid.New().String(),
		EventType:      "report.scheduled",
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		ScheduleID:     schedule.ID.String(),
		TemplateID:     schedule.TemplateID.String(),
		ScheduleName:   schedule.Name,
		Frequency:      string(schedule.Frequency),
		CronExpression: schedule.CronExpression,
		NextRunAt:      schedule.NextRunAt.Format(time.RFC3339),
	}

	return p.publish(ctx, topic, schedule.ID.String(), event)
}

// PublishExportAudit publishes an export audit event
func (p *Producer) PublishExportAudit(ctx context.Context, exportLog *domain.ExportLog) error {
	topic := fmt.Sprintf("%s_export_audit", p.config.TopicPrefix)

	event := ExportAuditEvent{
		EventID:        uuid.New().String(),
		EventType:      "export.audit",
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		ExportLogID:    exportLog.ID.String(),
		ReportID:       exportLog.ReportID.String(),
		UserID:         exportLog.UserID,
		UserName:       exportLog.UserName,
		Format:         string(exportLog.ExportFormat),
		FileName:       exportLog.FileName,
		FileSize:       exportLog.FileSize,
		EncryptionType: string(exportLog.EncryptionType),
		IPAddress:      exportLog.IPAddress,
		Status:         string(exportLog.Status),
	}

	return p.publish(ctx, topic, exportLog.ID.String(), event)
}

// publish publishes a message to Kafka
func (p *Producer) publish(ctx context.Context, topic, key string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Failed to publish Kafka event to topic %s: %v", topic, err)
		return err
	}

	log.Printf("Published Kafka event to topic %s: %s", topic, key)
	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer implements the KafkaConsumer interface
type Consumer struct {
	reader *kafka.Reader
	config KafkaConfig
}

// ConsumerConfig holds the consumer configuration
type ConsumerConfig struct {
	KafkaConfig
	Topics       []string
	MinBytes     int
	MaxBytes     int
	MaxWait      time.Duration
	StartOffset  int64
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Brokers,
		GroupID:        config.ConsumerGroup,
		Topic:          config.Topics[0],
		MinBytes:       config.MinBytes,
		MaxBytes:       config.MaxBytes,
		MaxWait:        config.MaxWait,
		StartOffset:    config.StartOffset,
		CommitInterval: time.Second,
	})

	return &Consumer{
		reader: reader,
		config: config.KafkaConfig,
	}
}

// Consume starts consuming messages
func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			if err := handler(ctx, msg.Value); err != nil {
				log.Printf("Error handling message: %v", err)
				// Continue processing other messages
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// Event types

// ReportGeneratedEvent represents a report generated event
type ReportGeneratedEvent struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	Timestamp    string `json:"timestamp"`
	ReportID     string `json:"report_id"`
	TemplateID   string `json:"template_id"`
	TemplateName string `json:"template_name"`
	ReportType   string `json:"report_type"`
	RegulatorID  string `json:"regulator_id"`
	Status       string `json:"status"`
	Title        string `json:"title"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	GeneratedBy  string `json:"generated_by"`
	FileCount    int    `json:"file_count"`
	TotalSize    int64  `json:"total_size"`
}

// ReportFailedEvent represents a report failed event
type ReportFailedEvent struct {
	EventID     string `json:"event_id"`
	EventType   string `json:"event_type"`
	Timestamp   string `json:"timestamp"`
	ReportID    string `json:"report_id"`
	Reason      string `json:"reason"`
	RetryPolicy string `json:"retry_policy"`
}

// ReportScheduledEvent represents a report scheduled event
type ReportScheduledEvent struct {
	EventID        string `json:"event_id"`
	EventType      string `json:"event_type"`
	Timestamp      string `json:"timestamp"`
	ScheduleID     string `json:"schedule_id"`
	TemplateID     string `json:"template_id"`
	ScheduleName   string `json:"schedule_name"`
	Frequency      string `json:"frequency"`
	CronExpression string `json:"cron_expression"`
	NextRunAt      string `json:"next_run_at"`
}

// ExportAuditEvent represents an export audit event
type ExportAuditEvent struct {
	EventID        string `json:"event_id"`
	EventType      string `json:"event_type"`
	Timestamp      string `json:"timestamp"`
	ExportLogID    string `json:"export_log_id"`
	ReportID       string `json:"report_id"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	Format         string `json:"format"`
	FileName       string `json:"file_name"`
	FileSize       int64  `json:"file_size"`
	EncryptionType string `json:"encryption_type"`
	IPAddress      string `json:"ip_address"`
	Status         string `json:"status"`
}

// DefaultKafkaConfig returns the default Kafka configuration
func DefaultKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers:       []string{"localhost:9092"},
		ConsumerGroup: "csic-reporting-consumer",
		TopicPrefix:   "csic",
	}
}
