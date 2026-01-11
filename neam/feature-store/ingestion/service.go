package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/redis/go-redis/v9"
	"neam-platform/shared"
)

// Config holds the configuration for the ingestion service
type Config struct {
	Kafka      *shared.KafkaConsumer
	ClickHouse clickhouse.Conn
	Redis      *redis.Client
	OnlineStore *OnlineStore
	Logger     *shared.Logger
}

// Service handles feature ingestion from various sources
type Service struct {
	config      Config
	logger      *shared.Logger
	kafka       *shared.KafkaConsumer
	clickhouse  clickhouse.Conn
	redis       *redis.Client
	onlineStore *OnlineStore
	transformers map[string]Transformer
	processors   map[string]Processor
	jobs         map[string]*IngestionJob
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// IngestionJob represents a feature ingestion job
type IngestionJob struct {
	ID           string                 `json:"id"`
	FeatureName  string                 `json:"feature_name"`
	Source       string                 `json:"source"`
	SourceType   string                 `json:"source_type"` // kafka, batch, stream
	Status       string                 `json:"status"` // running, paused, failed, completed
	RecordsProcessed int64              `json:"records_processed"`
	RecordsFailed int64                 `json:"records_failed"`
	LastRun      *time.Time             `json:"last_run"`
	Schedule     string                 `json:"schedule"` // cron expression
	Config       IngestionConfig        `json:"config"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// IngestionConfig holds configuration for ingestion
type IngestionConfig struct {
	KafkaTopic    string            `json:"kafka_topic"`
	KafkaGroup    string            `json:"kafka_group"`
	BatchSize     int               `json:"batch_size"`
	FlushInterval time.Duration     `json:"flush_interval"`
	Transform     string            `json:"transform"` // name of transformer
	Validate      bool              `json:"validate"`
	TargetTable   string            `json:"target_table"`
	TargetColumns []string          `json:"target_columns"`
}

// FeatureEvent represents an incoming feature event
type FeatureEvent struct {
	EntityID     string                 `json:"entity_id"`
	FeatureName  string                 `json:"feature_name"`
	Value        interface{}            `json:"value"`
	Timestamp    time.Time              `json:"timestamp"`
	Version      string                 `json:"version"`
	Source       string                 `json:"source"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Transformer transforms raw data into feature events
type Transformer interface {
	Transform(ctx context.Context, rawData map[string]interface{}) (*FeatureEvent, error)
	GetName() string
}

// Processor processes feature events
type Processor interface {
	Process(ctx context.Context, events []*FeatureEvent) (processed int64, failed int64, err error)
	GetName() string
}

// JSONTransformer transforms JSON data
type JSONTransformer struct{}

// CSVTransformer transforms CSV data
type CSVTransformer struct{}

// KafkaProcessor processes events from Kafka
type KafkaProcessor struct{}

// BatchProcessor processes events in batches
type BatchProcessor struct{}

// StreamProcessor processes events in real-time
type StreamProcessor struct{}

// NewService creates a new ingestion service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	svc := &Service{
		config:      cfg,
		logger:      cfg.Logger,
		kafka:       cfg.Kafka,
		clickhouse:  cfg.ClickHouse,
		redis:       cfg.Redis,
		onlineStore: cfg.OnlineStore,
		transformers: make(map[string]Transformer),
		processors:  make(map[string]Processor),
		jobs:        make(map[string]*IngestionJob),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Register transformers
	svc.transformers["json"] = &JSONTransformer{}
	svc.transformers["csv"] = &CSVTransformer{}

	// Register processors
	svc.processors["kafka"] = &KafkaProcessor{}
	svc.processors["batch"] = &BatchProcessor{}
	svc.processors["stream"] = &StreamProcessor{}

	return svc
}

// Start begins the ingestion service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting ingestion service")

	// Start job scheduler
	s.wg.Add(1)
	go s.jobSchedulerLoop()

	// Start running jobs
	s.wg.Add(1)
	go s.jobExecutionLoop()

	s.logger.Info("Ingestion service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping ingestion service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Ingestion service stopped")
}

// jobSchedulerLoop manages job scheduling
func (s *Service) jobSchedulerLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkScheduledJobs()
		}
	}
}

// checkScheduledJobs checks and triggers scheduled jobs
func (s *Service) checkScheduledJobs() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, job := range s.jobs {
		if job.Status == "paused" {
			continue
		}

		// Check if job should run based on schedule
		if s.shouldRunJob(job) {
			s.startJobExecution(job)
		}
	}
}

// shouldRunJob determines if a job should run
func (s *Service) shouldRunJob(job *IngestionJob) bool {
	if job.Schedule == "" {
		return false
	}

	// Parse cron expression and check if job should run
	// Simplified - in production, use a proper cron parser
	return true
}

// jobExecutionLoop executes running jobs
func (s *Service) jobExecutionLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processActiveJobs()
		}
	}
}

// processActiveJobs processes active ingestion jobs
func (s *Service) processActiveJobs() {
	s.mu.RLock()
	jobs := make([]*IngestionJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		if job.Status == "running" {
			jobs = append(jobs, job)
		}
	}
	s.mu.RUnlock()

	for _, job := range jobs {
		s.executeJob(job)
	}
}

// startJobExecution starts executing a job
func (s *Service) startJobExecution(job *IngestionJob) {
	s.mu.Lock()
	job.Status = "running"
	job.UpdatedAt = time.Now()
	s.mu.Unlock()

	go s.executeJob(job)
}

// executeJob executes an ingestion job
func (s *Service) executeJob(job *IngestionJob) {
	s.logger.Info("Executing ingestion job", "job_id", job.ID, "feature", job.FeatureName)

	processor, exists := s.processors[job.SourceType]
	if !exists {
		s.logger.Error("Processor not found", "source_type", job.SourceType)
		return
	}

	switch job.SourceType {
	case "kafka":
		s.processKafkaJob(job, processor)
	case "batch":
		s.processBatchJob(job, processor)
	case "stream":
		s.processStreamJob(job, processor)
	}
}

// processKafkaJob processes Kafka-based ingestion
func (s *Service) processKafkaJob(job *IngestionJob, processor Processor) {
	// Consume messages from Kafka topic
	// This would use the Kafka consumer to read messages

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Process messages in batches
			batchSize := job.Config.BatchSize
			events := s.consumeKafkaMessages(job.Config.KafkaTopic, batchSize)

			if len(events) == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			processed, failed, err := processor.Process(s.ctx, events)
			if err != nil {
				s.logger.Error("Processing failed", "job_id", job.ID, "error", err)
				job.RecordsFailed += failed
				continue
			}

			job.RecordsProcessed += processed
			job.RecordsFailed += failed
			job.LastRun = time.NowPtr()

			s.updateJobStats(job)
		}
	}
}

// consumeKafkaMessages consumes messages from Kafka
func (s *Service) consumeKafkaMessages(topic string, limit int) []*FeatureEvent {
	var events []*FeatureEvent

	// This would consume from Kafka
	// Simplified implementation
	return events
}

// processBatchJob processes batch ingestion
func (s *Service) processBatchJob(job *IngestionJob, processor Processor) {
	// Process batch data
	batch := s.loadBatchData(job)

	for len(batch) > 0 {
		processed, failed, err := processor.Process(s.ctx, batch)
		if err != nil {
			s.logger.Error("Batch processing failed", "job_id", job.ID, "error", err)
			job.RecordsFailed += failed
			break
		}

		job.RecordsProcessed += processed
		job.RecordsFailed += failed
		job.LastRun = time.Now()
		batch = s.loadBatchData(job)
	}

	job.Status = "completed"
	job.UpdatedAt = time.Now()
}

// processStreamJob processes stream ingestion
func (s *Service) processStreamJob(job *IngestionJob, processor Processor) {
	// Process real-time stream
	stream := s.openStream(job)

	for event := range stream {
		_, failed, err := processor.Process(s.ctx, []*FeatureEvent{event})
		if err != nil {
			s.logger.Error("Stream processing failed", "job_id", job.ID, "error", err)
			job.RecordsFailed++
		}
		job.RecordsProcessed++
		job.LastRun = time.Now()
	}
}

// loadBatchData loads batch data for processing
func (s *Service) loadBatchData(job *IngestionJob) []*FeatureEvent {
	// Load data from source based on job config
	return []*FeatureEvent{}
}

// openStream opens a stream for continuous ingestion
func (s *Service) openStream(job *IngestionJob) <-chan *FeatureEvent {
	events := make(chan *FeatureEvent)

	// This would open a stream (e.g., from Kafka, Kinesis, etc.)
	go func() {
		defer close(events)
	}()

	return events
}

// updateJobStats updates job statistics
func (s *Service) updateJobStats(job *IngestionJob) {
	s.mu.Lock()
	job.UpdatedAt = time.Now()
	s.mu.Unlock()

	// Cache stats in Redis
	statsKey := fmt.Sprintf("neam:ingestion:%s:stats", job.ID)
	s.redis.HSet(s.ctx, statsKey, map[string]interface{}{
		"records_processed": job.RecordsProcessed,
		"records_failed":    job.RecordsFailed,
		"last_run":          job.LastRun.Format(time.RFC3339),
		"updated_at":        job.UpdatedAt.Format(time.RFC3339),
	})
}

// CreateJob creates a new ingestion job
func (s *Service) CreateJob(ctx context.Context, featureName, source, sourceType string, config IngestionConfig) (*IngestionJob, error) {
	job := &IngestionJob{
		ID:          fmt.Sprintf("ingest-%d", time.Now().UnixNano()),
		FeatureName: featureName,
		Source:      source,
		SourceType:  sourceType,
		Status:      "paused",
		Config:      config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	s.logger.Info("Ingestion job created", "job_id", job.ID, "feature", featureName)

	return job, nil
}

// StartJob starts an ingestion job
func (s *Service) StartJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = "running"
	job.UpdatedAt = time.Now()

	return nil
}

// StopJob stops an ingestion job
func (s *Service) StopJob(ctx context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Status = "paused"
	job.UpdatedAt = time.Now()

	return nil
}

// GetJob retrieves an ingestion job
func (s *Service) GetJob(ctx context.Context, jobID string) (*IngestionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all ingestion jobs
func (s *Service) ListJobs(ctx context.Context, filters map[string]string) ([]*IngestionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*IngestionJob
	for _, job := range s.jobs {
		if filters["feature"] != "" && job.FeatureName != filters["feature"] {
			continue
		}
		if filters["status"] != "" && job.Status != filters["status"] {
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// IngestEvent ingests a single feature event
func (s *Service) IngestEvent(ctx context.Context, event *FeatureEvent) error {
	// Validate event
	if err := s.validateEvent(event); err != nil {
		return err
	}

	// Store in ClickHouse (offline store)
	if err := s.storeEvent(ctx, event); err != nil {
		return err
	}

	// Update online store
	if s.onlineStore != nil {
		if err := s.onlineStore.SetFeature(ctx, event.EntityID, event.FeatureName, event.Version, event.Value, 24*time.Hour); err != nil {
			s.logger.Warn("Failed to update online store", "error", err)
		}
	}

	// Publish to Kafka for downstream consumers
	s.publishEvent(event)

	return nil
}

// IngestEvents ingests multiple feature events
func (s *Service) IngestEvents(ctx context.Context, events []*FeatureEvent) (int64, int64, error) {
	var processed, failed int64

	for _, event := range events {
		if err := s.IngestEvent(ctx, event); err != nil {
			failed++
			s.logger.Warn("Event ingestion failed", "error", err)
			continue
		}
		processed++
	}

	return processed, failed, nil
}

// validateEvent validates a feature event
func (s *Service) validateEvent(event *FeatureEvent) error {
	if event.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	if event.FeatureName == "" {
		return fmt.Errorf("feature_name is required")
	}
	if event.Value == nil {
		return fmt.Errorf("value is required")
	}
	return nil
}

// storeEvent stores event in ClickHouse
func (s *Service) storeEvent(ctx context.Context, event *FeatureEvent) error {
	query := `
		INSERT INTO neam.feature_store (entity_id, feature_name, value, timestamp, version, source)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	return s.clickhouse.Exec(ctx, query,
		event.EntityID,
		event.FeatureName,
		fmt.Sprintf("%v", event.Value),
		event.Timestamp,
		event.Version,
		event.Source,
	)
}

// publishEvent publishes event to Kafka
func (s *Service) publishEvent(event *FeatureEvent) {
	// This would publish to a Kafka topic for downstream processing
}

// Transformer implementations
func (t *JSONTransformer) Transform(ctx context.Context, rawData map[string]interface{}) (*FeatureEvent, error) {
	event := &FeatureEvent{
		EntityID:    rawData["entity_id"].(string),
		FeatureName: rawData["feature_name"].(string),
		Value:       rawData["value"],
		Timestamp:   time.Now(),
		Version:     "latest",
	}
	return event, nil
}

func (t *JSONTransformer) GetName() string {
	return "json"
}

func (t *CSVTransformer) Transform(ctx context.Context, rawData map[string]interface{}) (*FeatureEvent, error) {
	return nil, nil
}

func (t *CSVTransformer) GetName() string {
	return "csv"
}

// Processor implementations
func (p *KafkaProcessor) Process(ctx context.Context, events []*FeatureEvent) (int64, int64, error) {
	return int64(len(events)), 0, nil
}

func (p *KafkaProcessor) GetName() string {
	return "kafka"
}

func (p *BatchProcessor) Process(ctx context.Context, events []*FeatureEvent) (int64, int64, error) {
	return int64(len(events)), 0, nil
}

func (p *BatchProcessor) GetName() string {
	return "batch"
}

func (p *StreamProcessor) Process(ctx context.Context, events []*FeatureEvent) (int64, int64, error) {
	return int64(len(events)), 0, nil
}

func (p *StreamProcessor) GetName() string {
	return "stream"
}
