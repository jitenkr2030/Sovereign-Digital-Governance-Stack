package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/forensic-tools/internal/core/domain"
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

// PublishAnalysisJob publishes an analysis job to Kafka
func (p *KafkaProducer) PublishAnalysisJob(ctx context.Context, job *domain.AnalysisJob) error {
	event := map[string]interface{}{
		"event_type":   "ANALYSIS_JOB_SUBMITTED",
		"job_id":       job.ID,
		"evidence_id":  job.EvidenceID,
		"tool_name":    job.ToolName,
		"priority":     job.Priority,
		"requested_by": job.RequestedBy,
		"submitted_at": job.SubmittedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis job event: %w", err)
	}

	return p.Publish(ctx, "security.forensics.jobs", data)
}

// PublishAnalysisResult publishes an analysis result to Kafka
func (p *KafkaProducer) PublishAnalysisResult(ctx context.Context, result *domain.AnalysisResult) error {
	event := map[string]interface{}{
		"event_type":  "ANALYSIS_RESULT_COMPLETED",
		"job_id":      result.JobID,
		"evidence_id": result.EvidenceID,
		"tool_name":   result.ToolName,
		"result_type": result.ResultType,
		"severity":    result.Severity,
		"tags":        result.Tags,
		"created_at":  result.CreatedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis result event: %w", err)
	}

	return p.Publish(ctx, "security.forensics.results", data)
}

// PublishCustodyEvent publishes a chain of custody event to Kafka
func (p *KafkaProducer) PublishCustodyEvent(ctx context.Context, entry *domain.ChainOfCustody) error {
	event := map[string]interface{}{
		"event_type": "CUSTODY_EVENT",
		"evidence_id": entry.EvidenceID,
		"actor_id":   entry.ActorID,
		"action":     entry.Action,
		"details":    entry.Details,
		"ip_address": entry.IPAddress,
		"timestamp":  entry.Timestamp.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal custody event: %w", err)
	}

	return p.Publish(ctx, "security.forensics.custody", data)
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

// Close closes all Kafka writers
func (p *KafkaProducer) Close() error {
	for _, w := range p.writers {
		if err := w.Close(); err != nil {
			return fmt.Errorf("failed to close Kafka writer: %w", err)
		}
	}
	return nil
}
