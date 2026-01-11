package dataquality

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Monitor provides centralized data quality monitoring
type Monitor struct {
	mu              sync.RWMutex
	scorer          *Scorer
	detector        *Detector
	registry        *SchemaRegistry
	metrics         *DataQualityMetrics
	config          MonitorConfig
	alertManager    *AlertManager
	logger          *zap.Logger
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// MonitorConfig contains configuration for the monitor
type MonitorConfig struct {
	ScorerConfig      QualityConfig
	AnomalyConfig     AnomalyConfig
	MetricsEnabled    bool
	AlertEnabled      bool
	CheckInterval     time.Duration
	ReportInterval    time.Duration
}

// DefaultMonitorConfig returns the default monitor configuration
func DefaultMonitorConfig() MonitorConfig {
	return MonitorConfig{
		ScorerConfig:  DefaultConfig(),
		AnomalyConfig: DefaultAnomalyConfig(),
		MetricsEnabled: true,
		AlertEnabled:   true,
		CheckInterval:  time.Minute,
		ReportInterval: 5 * time.Minute,
	}
}

// NewMonitor creates a new data quality monitor
func NewMonitor(config MonitorConfig, registry *SchemaRegistry, logger *zap.Logger) (*Monitor, error) {
	scorer, err := NewScorer(config.ScorerConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create scorer: %w", err)
	}

	detector := NewDetector(config.AnomalyConfig, logger)

	var metrics *DataQualityMetrics
	if config.MetricsEnabled {
		metrics = NewDataQualityMetrics("neam")
	}

	alertManager := NewAlertManager(logger)

	monitor := &Monitor{
		scorer:       scorer,
		detector:     detector,
		registry:     registry,
		metrics:      metrics,
		config:       config,
		alertManager: alertManager,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}

	// Set up anomaly alert handler
	detector.SetAlertFunc(func(anomaly *Anomaly) {
		monitor.handleAnomaly(anomaly)
	})

	return monitor, nil
}

// Start starts the monitoring background tasks
func (m *Monitor) Start(ctx context.Context) {
	m.wg.Add(2)

	// Start periodic health check
	go m.runHealthCheck(ctx)

	// Start periodic report generation
	go m.runReportGeneration(ctx)

	m.logger.Info("Data quality monitor started")
}

// Stop stops the monitoring background tasks
func (m *Monitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	m.logger.Info("Data quality monitor stopped")
}

// Observe processes a new record and updates quality metrics
func (m *Monitor) Observe(sourceID string, record map[string]interface{}, metadata RecordMetadata) *RecordAnalysis {
	// Analyze the record
	analysis := m.scorer.AnalyzeRecord(sourceID, record, metadata)

	// Update anomaly detection
	if analysis.OverallScore > 0 {
		m.detector.Observe(sourceID, "quality_score", analysis.OverallScore)
	}

	// Record metrics
	if m.metrics != nil && m.config.MetricsEnabled {
		m.recordMetrics(sourceID, analysis)
	}

	return analysis
}

// ValidateAndScore performs schema validation and quality scoring
func (m *Monitor) ValidateAndScore(ctx context.Context, sourceID string, data []byte, metadata RecordMetadata) (*ValidationResult, *RecordAnalysis) {
	// Parse the data
	var record map[string]interface{}
	if err := json.Unmarshal(data, &record); err != nil {
		m.logger.Error("Failed to parse record data",
			zap.Error(err),
			zap.String("source_id", sourceID))

		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{{
				Field:     "root",
				Message:   "Invalid JSON format",
				ErrorType: "parse_error",
			}},
		}, nil
	}

	// Validate against schema if available
	var validationResult *ValidationResult
	if m.registry != nil {
		validator, err := m.registry.GetLatestValidator(sourceID)
		if err == nil {
			result, err := validator.Validate(ctx, data)
			if err == nil {
				validationResult = result
			}
		}
	}

	// Calculate quality score
	analysis := m.Observe(sourceID, record, metadata)

	return validationResult, analysis
}

// GetScore returns the current quality score for a source
func (m *Monitor) GetScore(sourceID string) *QualityScore {
	return m.scorer.GetScore(sourceID)
}

// GetAllScores returns quality scores for all monitored sources
func (m *Monitor) GetAllScores() map[string]*QualityScore {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scores := make(map[string]*QualityScore)
	for sourceID := range m.scorer.sourceAnalyses {
		scores[sourceID] = m.scorer.GetScore(sourceID)
	}
	return scores
}

// GetAnomalies returns recent anomalies for a source
func (m *Monitor) GetAnomalies(sourceID string, since time.Duration) []*Anomaly {
	return m.alertManager.GetAnomalies(sourceID, since)
}

// GetAllAnomalies returns all recent anomalies
func (m *Monitor) GetAllAnomalies(since time.Duration) []*Anomaly {
	return m.alertManager.GetAllAnomalies(since)
}

// recordMetrics records Prometheus metrics for the analysis
func (m *Monitor) recordMetrics(sourceID string, analysis *RecordAnalysis) {
	// Overall score
	m.metrics.RecordOverallQuality(sourceID, analysis.OverallScore)

	// Dimension scores
	m.metrics.RecordCompleteness(sourceID, "all", analysis.Completeness, analysis.Completeness < 0.8)
	m.metrics.RecordValidity(sourceID, "all", analysis.Validity, analysis.Validity < 0.9)
	m.metrics.RecordConsistency(sourceID, "all", analysis.Consistency, analysis.Consistency < 0.9)
	m.metrics.RecordTimeliness(sourceID, analysis.Timeliness, 0, analysis.Timeliness < 0.8)
}

// runHealthCheck performs periodic health checks
func (m *Monitor) runHealthCheck(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck performs a single health check
func (m *Monitor) performHealthCheck() {
	scores := m.GetAllScores()

	for sourceID, score := range scores {
		// Check for low scores
		if score.OverallScore < m.config.ScorerConfig.OverallThreshold {
			m.alertManager.AddAlert(&Alert{
				AlertID:   fmt.Sprintf("alert_%d", time.Now().UnixNano()),
				Timestamp: time.Now(),
				Severity:  "warning",
				Category:  "quality_score",
				SourceID:  sourceID,
				Message:   fmt.Sprintf("Quality score %f below threshold %f",
					score.OverallScore, m.config.ScorerConfig.OverallThreshold),
			})
		}

		// Update Prometheus gauges
		if m.metrics != nil {
			for dimension, dimScore := range score.DimensionScores {
				m.metrics.QualityScore.WithLabelValues(sourceID, string(dimension)).Set(dimScore)
			}
		}
	}
}

// runReportGeneration generates periodic reports
func (m *Monitor) runReportGeneration(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.generateReport()
		}
	}
}

// generateReport generates a quality report
func (m *Monitor) generateReport() {
	report := &QualityReport{
		ReportID:   fmt.Sprintf("report_%d", time.Now().UnixNano()),
		Timestamp:  time.Now(),
		Period: ReportPeriod{
			Start: time.Now().Add(-m.config.ReportInterval),
			End:   time.Now(),
		},
	}

	scores := m.GetAllScores()
	for sourceID, score := range scores {
		report.SourceScores = append(report.SourceScores, *score)
	}

	anomalies := m.GetAllAnomalies(m.config.ReportInterval)
	for _, anomaly := range anomalies {
		report.Anomalies = append(report.Anomalies, *anomaly)
	}

	// Calculate overall statistics
	var totalScore float64
	for _, score := range report.SourceScores {
		totalScore += score.OverallScore
	}
	if len(report.SourceScores) > 0 {
		report.AverageScore = totalScore / float64(len(report.SourceScores))
	}

	// Log the report
	m.logger.Info("Quality report generated",
		zap.String("report_id", report.ReportID),
		zap.Float64("average_score", report.AverageScore),
		zap.Int("source_count", len(report.SourceScores)),
		zap.Int("anomaly_count", len(report.Anomalies)))

	// Store the report
	m.mu.Lock()
	m.lastReport = report
	m.mu.Unlock()
}

// GetLastReport returns the last generated report
func (m *Monitor) GetLastReport() *QualityReport {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastReport
}

var lastReport *QualityReport

// handleAnomaly handles detected anomalies
func (m *Monitor) handleAnomaly(anomaly *Anomaly) {
	if m.config.AlertEnabled {
		m.alertManager.AddAnomaly(anomaly)
	}
}

// AddConsistencyCheck adds a consistency check result
func (m *Monitor) AddConsistencyCheck(sourceID, checkName string, passed bool) {
	m.scorer.AddConsistencyCheck(sourceID, checkName, passed)
}

// UpdateLatency updates latency statistics
func (m *Monitor) UpdateLatency(sourceID string, latency time.Duration) {
	m.scorer.UpdateLatency(sourceID, latency)

	// Update anomaly detection
	m.detector.Observe(sourceID, "latency", float64(latency.Seconds()))
}

// QualityReport represents a quality monitoring report
type QualityReport struct {
	ReportID     string         `json:"report_id"`
	Timestamp    time.Time      `json:"timestamp"`
	Period       ReportPeriod   `json:"period"`
	SourceScores []QualityScore `json:"source_scores"`
	Anomalies    []Anomaly      `json:"anomalies"`
	AverageScore float64        `json:"average_score"`
}

// ReportPeriod represents the time period of a report
type ReportPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Alert represents a quality alert
type Alert struct {
	AlertID   string                 `json:"alert_id"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  string                 `json:"severity"`
	Category  string                 `json:"category"`
	SourceID  string                 `json:"source_id"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Resolved  bool                   `json:"resolved"`
}

// AlertManager manages quality alerts
type AlertManager struct {
	mu        sync.RWMutex
	alerts    []*Alert
	anomalies []*Anomaly
	logger    *zap.Logger
	maxAlerts int
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		logger:    logger,
		maxAlerts: 1000,
	}
}

// AddAlert adds a new alert
func (am *AlertManager) AddAlert(alert *Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.alerts = append(am.alerts, alert)

	// Maintain max size
	if len(am.alerts) > am.maxAlerts {
		am.alerts = am.alerts[len(am.alerts)-am.maxAlerts:]
	}

	am.logger.Info("Alert added",
		zap.String("alert_id", alert.AlertID),
		zap.String("severity", alert.Severity),
		zap.String("source_id", alert.SourceID),
		zap.String("message", alert.Message))
}

// AddAnomaly adds a new anomaly
func (am *AlertManager) AddAnomaly(anomaly *Anomaly) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.anomalies = append(am.anomalies, anomaly)

	// Maintain max size
	if len(am.anomalies) > am.maxAlerts {
		am.anomalies = am.anomalies[len(am.anomalies)-am.maxAlerts:]
	}
}

// GetAlerts returns alerts for a source
func (am *AlertManager) GetAlerts(sourceID string, since time.Duration) []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	cutoff := time.Now().Add(-since)
	var result []*Alert
	for _, alert := range am.alerts {
		if alert.SourceID == sourceID && alert.Timestamp.After(cutoff) {
			result = append(result, alert)
		}
	}
	return result
}

// GetAnomalies returns anomalies for a source
func (am *AlertManager) GetAnomalies(sourceID string, since time.Duration) []*Anomaly {
	am.mu.RLock()
	defer am.mu.RUnlock()

	cutoff := time.Now().Add(-since)
	var result []*Anomaly
	for _, anomaly := range am.anomalies {
		if anomaly.SourceID == sourceID && anomaly.Timestamp.After(cutoff) {
			result = append(result, anomaly)
		}
	}
	return result
}

// GetAllAlerts returns all recent alerts
func (am *AlertManager) GetAllAlerts(since time.Duration) []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	cutoff := time.Now().Add(-since)
	var result []*Alert
	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoff) {
			result = append(result, alert)
		}
	}
	return result
}

// GetAllAnomalies returns all recent anomalies
func (am *AlertManager) GetAllAnomalies(since time.Duration) []*Anomaly {
	am.mu.RLock()
	defer am.mu.RUnlock()

	cutoff := time.Now().Add(-since)
	var result []*Anomaly
	for _, anomaly := range am.anomalies {
		if anomaly.Timestamp.After(cutoff) {
			result = append(result, anomaly)
		}
	}
	return result
}

// SchemaRegistry provides schema validation capabilities
type SchemaRegistry struct {
	mu        sync.RWMutex
	schemas   map[string]*JSONSchema
	validators map[string]Validator
	logger    *zap.Logger
}

// JSONSchema represents a JSON schema
type JSONSchema struct {
	ID          string          `json:"id"`
	Version     string          `json:"version"`
	Schema      json.RawMessage `json:"schema"`
	Description string          `json:"description"`
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry(logger *zap.Logger) *SchemaRegistry {
	return &SchemaRegistry{
		schemas:   make(map[string]*JSONSchema),
		validators: make(map[string]Validator),
		logger:    logger,
	}
}

// Register registers a new schema
func (r *SchemaRegistry) Register(id, version string, schemaData json.RawMessage, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", id, version)

	// Create validator
	validator, err := NewJSONValidator(string(schemaData))
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// Store schema and validator
	r.schemas[key] = &JSONSchema{
		ID:          id,
		Version:     version,
		Schema:      schemaData,
		Description: description,
	}
	r.validators[key] = validator

	r.logger.Info("Schema registered",
		zap.String("id", id),
		zap.String("version", version))

	return nil
}

// GetValidator returns a validator for a schema
func (r *SchemaRegistry) GetValidator(id, version string) (Validator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", id, version)
	if validator, exists := r.validators[key]; exists {
		return validator, nil
	}

	// Try latest version
	if latest, err := r.GetLatestValidator(id); err == nil {
		return latest, nil
	}

	return nil, fmt.Errorf("schema not found: %s:%s", id, version)
}

// GetLatestValidator returns the latest validator for a schema
func (r *SchemaRegistry) GetLatestValidator(id string) (Validator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latestVersion string
	var latestValidator Validator

	for key, validator := range r.validators {
		if len(key) > len(id)+1 && key[:len(id)] == id {
			version := key[len(id)+1:]
			if latestVersion == "" || version > latestVersion {
				latestVersion = version
				latestValidator = validator
			}
		}
	}

	if latestValidator == nil {
		return nil, fmt.Errorf("no versions found for schema: %s", id)
	}

	return latestValidator, nil
}
