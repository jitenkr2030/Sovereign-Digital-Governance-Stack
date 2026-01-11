package offline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"neam-platform/shared"
)

// Config holds the configuration for the offline feature store
type Config struct {
	ClickHouse clickhouse.Conn
	MinIO      *shared.MinIO
	Logger     *shared.Logger
}

// Service handles batch feature storage and retrieval
type Service struct {
	config     Config
	logger     *shared.Logger
	clickhouse clickhouse.Conn
	minio      *shared.MinIO
	exports    map[string]*ExportJob
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// ExportJob represents a feature export job
type ExportJob struct {
	ID           string                 `json:"id"`
	FeatureNames []string               `json:"feature_names"`
	EntityIDs    []string               `json:"entity_ids"`
	Version      string                 `json:"version"`
	StartTime    *time.Time             `json:"start_time"`
	EndTime      *time.Time             `json:"end_time"`
	AsOfTime     *time.Time             `json:"as_of_time"`
	Format       string                 `json:"format"` // parquet, csv, json
	Status       string                 `json:"status"` // pending, processing, completed, failed
	OutputPath   string                 `json:"output_path"`
	FileSize     int64                  `json:"file_size"`
	RowCount     int64                  `json:"row_count"`
	Error        string                 `json:"error"`
	RequestedBy  string                 `json:"requested_by"`
	RequestedAt  time.Time              `json:"requested_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
}

// PointInTimeQuery represents a point-in-time correctness query
type PointInTimeQuery struct {
	EntityIDs     []string    `json:"entity_ids"`
	Features      []string    `json:"features"`
	AsOfTime      time.Time   `json:"as_of_time"`
	Version       string      `json:"version"`
	LookbackDays  int         `json:"lookback_days"`
}

// FeatureSnapshot represents a feature snapshot at a point in time
type FeatureSnapshot struct {
	EntityID     string                 `json:"entity_id"`
	FeatureName  string                 `json:"feature_name"`
	Value        interface{}            `json:"value"`
	AsOfTime     time.Time              `json:"as_of_time"`
	SourceTime   time.Time              `json:"source_time"`
	Version      string                 `json:"version"`
}

// BatchResult represents the result of a batch feature retrieval
type BatchResult struct {
	Rows         []map[string]interface{} `json:"rows"`
	TotalRows    int64                    `json:"total_rows"`
	ColumnNames  []string                 `json:"column_names"`
	FeatureNames []string                 `json:"feature_names"`
	AsOfTime     time.Time                `json:"as_of_time"`
	Version      string                   `json:"version"`
	QueryTime    time.Duration            `json:"query_time"`
}

// FeatureHistoryEntry represents a historical feature value
type FeatureHistoryEntry struct {
	EntityID    string      `json:"entity_id"`
	FeatureName string      `json:"feature_name"`
	Value       interface{} `json:"value"`
	Timestamp   time.Time   `json:"timestamp"`
	Version     string      `json:"version"`
}

// NewService creates a new offline feature store service
func NewService(cfg Config) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:    cfg,
		logger:    cfg.Logger,
		clickhouse: cfg.ClickHouse,
		minio:     cfg.MinIO,
		exports:   make(map[string]*ExportJob),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins the offline feature store service
func (s *Service) Start(ctx context.Context) {
	s.logger.Info("Starting offline feature store service")

	// Start export processing
	s.wg.Add(1)
	go s.exportProcessingLoop()

	s.logger.Info("Offline feature store service started")
}

// Stop gracefully stops the service
func (s *Service) Stop() {
	s.logger.Info("Stopping offline feature store service")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("Offline feature store service stopped")
}

// exportProcessingLoop processes pending export jobs
func (s *Service) exportProcessingLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processPendingExports()
		}
	}
}

// processPendingExports processes export jobs that are pending
func (s *Service) processPendingExports() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, job := range s.exports {
		if job.Status == "pending" {
			s.processExport(job)
		}
	}
}

// processExport executes an export job
func (s *Service) processExport(job *ExportJob) {
	job.Status = "processing"
	now := time.Now()
	job.StartTime = &now

	// Build query for feature data
	query := s.buildExportQuery(job)

	// Execute query
	rows, err := s.clickhouse.Query(context.Background(), query)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		return
	}
	defer rows.Close()

	// Write to output format
	outputPath, rowCount, err := s.writeExportData(job, rows)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		return
	}

	// Upload to MinIO
	if s.minio != nil {
		if err := s.uploadToMinIO(job, outputPath); err != nil {
			job.Status = "failed"
			job.Error = err.Error()
			return
		}
	}

	job.Status = "completed"
	job.OutputPath = outputPath
	job.RowCount = rowCount
	completedAt := time.Now()
	job.CompletedAt = &completedAt

	s.logger.Info("Export completed", "job_id", job.ID, "rows", rowCount)
}

// buildExportQuery constructs the SQL query for export
func (s *Service) buildExportQuery(job *ExportJob) string {
	// Construct ClickHouse query for feature export
	featureColumns := ""
	for i, name := range job.FeatureNames {
		if i > 0 {
			featureColumns += ", "
		}
		featureColumns += fmt.Sprintf("maxBy(value, timestamp) as %s", name)
	}

	query := fmt.Sprintf(`
		SELECT 
			entity_id,
			%s
		FROM neam.feature_store
		WHERE entity_id IN ('%s')
		AND feature_name IN ('%s')
		AND timestamp <= '%s'
		GROUP BY entity_id
		ORDER BY entity_id
	`,
		featureColumns,
		joinStrings(job.EntityIDs, "','"),
		joinStrings(job.FeatureNames, "','"),
		job.AsOfTime.Format(time.RFC3339),
	)

	return query
}

// writeExportData writes the exported data to a file
func (s *Service) writeExportData(job *ExportJob, rows *sql.Rows) (string, int64, error) {
	// Write data to Parquet format
	outputPath := fmt.Sprintf("/tmp/neam/exports/%s.%s", job.ID, job.Format)

	switch job.Format {
	case "parquet":
		return s.writeParquet(job, rows)
	case "csv":
		return s.writeCSV(job, rows)
	case "json":
		return s.writeJSON(job, rows)
	default:
		return s.writeParquet(job, rows)
	}
}

// writeParquet writes data in Parquet format
func (s *Service) writeParquet(job *ExportJob, rows *sql.Rows) (string, int64, error) {
	// Implementation using parquet-go or similar library
	outputPath := fmt.Sprintf("/tmp/neam/exports/%s.parquet", job.ID)
	return outputPath, 0, nil
}

// writeCSV writes data in CSV format
func (s *Service) writeCSV(job *ExportJob, rows *sql.Rows) (string, int64, error) {
	outputPath := fmt.Sprintf("/tmp/neam/exports/%s.csv", job.ID)
	return outputPath, 0, nil
}

// writeJSON writes data in JSON format
func (s *Service) writeJSON(job *ExportJob, rows *sql.Rows) (string, int64, error) {
	outputPath := fmt.Sprintf("/tmp/neam/exports/%s.json", job.ID)
	return outputPath, 0, nil
}

// uploadToMinIO uploads the exported file to MinIO
func (s *Service) uploadToMinIO(job *ExportJob, localPath string) error {
	bucket := "neam-exports"
	objectName := fmt.Sprintf("exports/%s/%s", job.ID, fmt.Sprintf("%s.%s", job.ID, job.Format))

	return s.minio.UploadFile(bucket, objectName, localPath)
}

// PointInTimeQuery executes a point-in-time correct query
func (s *Service) PointInTimeQuery(ctx context.Context, query PointInTimeQuery) (*BatchResult, error) {
	startTime := time.Now()

	// Build query with AS OF clause for point-in-time correctness
	sqlQuery := s.buildPITQuery(query)

	// Execute query
	rows, err := s.clickhouse.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("point-in-time query failed: %w", err)
	}
	defer rows.Close()

	// Process results
	results, err := s.processQueryResults(rows)
	if err != nil {
		return nil, err
	}

	queryTime := time.Since(startTime)
	results.QueryTime = queryTime
	results.AsOfTime = query.AsOfTime
	results.Version = query.Version

	return results, nil
}

// buildPITQuery constructs a point-in-time correct query
func (s *Service) buildPITQuery(query PointInTimeQuery) string {
	lookbackTime := query.AsOfTime.AddDate(0, 0, -query.LookbackDays)

	return fmt.Sprintf(`
		SELECT 
			entity_id,
			%s,
			timestamp as source_time
		FROM neam.feature_store FINAL
		WHERE entity_id IN ('%s')
		AND feature_name IN ('%s')
		AND timestamp <= '%s'
		AND timestamp >= '%s'
		ORDER BY entity_id, feature_name
	`,
		joinStrings(query.Features, ",\n\t\t"),
		joinStrings(query.EntityIDs, "','"),
		joinStrings(query.Features, "','"),
		query.AsOfTime.Format(time.RFC3339),
		lookbackTime.Format(time.RFC3339),
	)
}

// processQueryResults processes query results into BatchResult
func (s *Service) processQueryResults(rows *sql.Rows) (*BatchResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Get column types
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var results BatchResult
	results.ColumnNames = columns
	results.FeatureNames = columns[1:] // Skip entity_id

	// Process rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var rowCount int64
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results.Rows = append(results.Rows, row)
		rowCount++
	}

	results.TotalRows = rowCount

	return &results, nil
}

// BatchGetFeatures retrieves historical features for multiple entities
func (s *Service) BatchGetFeatures(ctx context.Context, entityIDs, featureNames []string, startTime, endTime time.Time, version string) (*BatchResult, error) {
	query := fmt.Sprintf(`
		SELECT 
			entity_id,
			%s,
			timestamp
		FROM neam.feature_store
		WHERE entity_id IN ('%s')
		AND feature_name IN ('%s')
		AND timestamp >= '%s'
		AND timestamp <= '%s'
		ORDER BY entity_id, timestamp
	`,
		joinStrings(featureNames, ",\n\t\t"),
		joinStrings(entityIDs, "','"),
		joinStrings(featureNames, "','"),
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)

	rows, err := s.clickhouse.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.processQueryResults(rows)
}

// GetFeatureHistory retrieves the history of a feature for an entity
func (s *Service) GetFeatureHistory(ctx context.Context, entityID, featureName string, startTime, endTime time.Time, version string) ([]FeatureHistoryEntry, error) {
	query := fmt.Sprintf(`
		SELECT 
			entity_id,
			feature_name,
			value,
			timestamp,
			version
		FROM neam.feature_store
		WHERE entity_id = '%s'
		AND feature_name = '%s'
		AND timestamp >= '%s'
		AND timestamp <= '%s'
		ORDER BY timestamp DESC
	`,
		entityID,
		featureName,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)

	rows, err := s.clickhouse.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []FeatureHistoryEntry
	for rows.Next() {
		var entry FeatureHistoryEntry
		var value string
		var timestamp time.Time

		err := rows.Scan(&entry.EntityID, &entry.FeatureName, &value, &timestamp, &entry.Version)
		if err != nil {
			continue
		}

		entry.Value = value
		entry.Timestamp = timestamp
		history = append(history, entry)
	}

	return history, nil
}

// CreateExportJob creates a new export job
func (s *Service) CreateExportJob(ctx context.Context, entityIDs, featureNames []string, asOfTime time.Time, format, requestedBy string) (*ExportJob, error) {
	job := &ExportJob{
		ID:           fmt.Sprintf("exp-%d", time.Now().UnixNano()),
		FeatureNames: featureNames,
		EntityIDs:    entityIDs,
		Version:      "latest",
		AsOfTime:     &asOfTime,
		Format:       format,
		Status:       "pending",
		RequestedBy:  requestedBy,
		RequestedAt:  time.Now(),
	}

	s.mu.Lock()
	s.exports[job.ID] = job
	s.mu.Unlock()

	return job, nil
}

// GetExportJob retrieves an export job
func (s *Service) GetExportJob(ctx context.Context, jobID string) (*ExportJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.exports[jobID]
	if !exists {
		return nil, fmt.Errorf("export job not found: %s", jobID)
	}

	return job, nil
}

// GetExportDownloadURL generates a presigned URL for download
func (s *Service) GetExportDownloadURL(ctx context.Context, jobID string) (string, error) {
	s.mu.RLock()
	job, exists := s.exports[jobID]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("export job not found: %s", jobID)
	}

	if job.Status != "completed" {
		return "", fmt.Errorf("export not ready: %s", job.Status)
	}

	if s.minio == nil {
		return "", nil
	}

	bucket := "neam-exports"
	objectName := fmt.Sprintf("exports/%s/%s", job.ID, fmt.Sprintf("%s.%s", job.ID, job.Format))

	return s.minio.GetPresignedURL(bucket, objectName, 24*time.Hour)
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
