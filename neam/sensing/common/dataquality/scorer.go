package dataquality

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Dimension represents a data quality dimension
type Dimension string

const (
	DimensionCompleteness  Dimension = "completeness"
	DimensionValidity      Dimension = "validity"
	DimensionConsistency   Dimension = "consistency"
	DimensionTimeliness    Dimension = "timeliness"
)

// QualityScore represents a calculated quality score
type QualityScore struct {
	SourceID        string             `json:"source_id"`
	Timestamp       time.Time          `json:"timestamp"`
	OverallScore    float64            `json:"overall_score"`
	DimensionScores map[Dimension]float64 `json:"dimension_scores"`
	ComponentScores map[string]float64 `json:"component_scores"`
	WeightConfig    WeightConfig       `json:"weight_config"`
}

// WeightConfig defines weights for each dimension
type WeightConfig struct {
	Completeness float64 `json:"completeness"`
	Validity     float64 `json:"validity"`
	Consistency  float64 `json:"consistency"`
	Timeliness  float64 `json:"timeliness"`
}

// DefaultWeights returns the default weight configuration
func DefaultWeights() WeightConfig {
	return WeightConfig{
		Completeness: 0.25,
		Validity:     0.35,
		Consistency:  0.20,
		Timeliness:  0.20,
	}
}

// Validate validates the weight configuration
func (w WeightConfig) Validate() error {
	total := w.Completeness + w.Validity + w.Consistency + w.Timeliness
	if math.Abs(total-1.0) > 0.001 {
		return fmt.Errorf("weights must sum to 1.0, got %f", total)
	}
	return nil
}

// Normalize normalizes weights to sum to 1.0
func (w WeightConfig) Normalize() WeightConfig {
	total := w.Completeness + w.Validity + w.Consistency + w.Timeliness
	if total == 0 {
		return DefaultWeights()
	}
	return WeightConfig{
		Completeness: w.Completeness / total,
		Validity:     w.Validity / total,
		Consistency:  w.Consistency / total,
		Timeliness:  w.Timeliness / total,
	}
}

// QualityConfig contains configuration for quality scoring
type QualityConfig struct {
	Weights               WeightConfig
	CompletenessThreshold float64
	ValidityThreshold     float64
	ConsistencyThreshold  float64
	TimelinessThreshold   float64
	OverallThreshold      float64
	WindowSize            int
}

// DefaultConfig returns the default quality configuration
func DefaultConfig() QualityConfig {
	return QualityConfig{
		Weights:               DefaultWeights(),
		CompletenessThreshold: 0.80,
		ValidityThreshold:     0.95,
		ConsistencyThreshold:  0.90,
		TimelinessThreshold:   0.85,
		OverallThreshold:      0.85,
		WindowSize:            100,
	}
}

// FieldAnalysis contains analysis of a single field
type FieldAnalysis struct {
	FieldName     string    `json:"field_name"`
	TotalCount    int       `json:"total_count"`
	NullCount     int       `json:"null_count"`
	EmptyCount    int       `json:"empty_count"`
	ValidCount    int       `json:"valid_count"`
	InvalidCount  int       `json:"invalid_count"`
	Completeness  float64   `json:"completeness"`
	Validity      float64   `json:"validity"`
	LastUpdated   time.Time `json:"last_updated"`
}

// RecordAnalysis contains analysis results for a record
type RecordAnalysis struct {
	RecordID       string                    `json:"record_id"`
	Timestamp      time.Time                 `json:"timestamp"`
	FieldAnalyses  map[string]*FieldAnalysis `json:"field_analyses"`
	Completeness   float64                   `json:"completeness"`
	Validity       float64                   `json:"validity"`
	Consistency    float64                   `json:"consistency"`
	Timeliness     float64                   `json:"timeliness"`
	OverallScore   float64                   `json:"overall_score"`
	Issues         []QualityIssue            `json:"issues"`
}

// QualityIssue represents a detected quality issue
type QualityIssue struct {
	IssueID     string        `json:"issue_id"`
	Timestamp   time.Time     `json:"timestamp"`
	Severity    string        `json:"severity"`
	Category    string        `json:"category"`
	Field       string        `json:"field"`
	Description string        `json:"description"`
	Details     json.RawMessage `json:"details,omitempty"`
}

// Scorer calculates data quality scores
type Scorer struct {
	mu              sync.RWMutex
	sourceAnalyses  map[string]*SourceAnalysis
	config          QualityConfig
	logger          *zap.Logger
}

// SourceAnalysis maintains running analysis for a data source
type SourceAnalysis struct {
	SourceID          string
	WindowStart       time.Time
	FieldStats        map[string]*FieldStats
	ConsistencyChecks map[string]*ConsistencyCheck
	TotalRecords      int64
	ValidRecords      int64
	InvalidRecords    int64
	TotalLatency      time.Duration
	LastUpdate        time.Time
}

// FieldStats maintains statistics for a field
type FieldStats struct {
	FieldName     string
	TotalCount    int64
	NullCount     int64
	EmptyCount    int64
	ValidCount    int64
	InvalidCount  int64
	SumValue      float64
	SumSquares    float64
	MinValue      float64
	MaxValue      float64
	LastValue     interface{}
	LastUpdated   time.Time
}

// ConsistencyCheck maintains statistics for a consistency check
type ConsistencyCheck struct {
	CheckName       string
	TotalChecks     int64
	PassedChecks    int64
	FailedChecks    int64
	LastCheckResult bool
	LastUpdated     time.Time
}

// NewScorer creates a new quality scorer
func NewScorer(config QualityConfig, logger *zap.Logger) (*Scorer, error) {
	if err := config.Weights.Validate(); err != nil {
		return nil, fmt.Errorf("invalid weight configuration: %w", err)
	}

	return &Scorer{
		sourceAnalyses: make(map[string]*SourceAnalysis),
		config:         config,
		logger:         logger,
	}, nil
}

// NewDefaultScorer creates a scorer with default configuration
func NewDefaultScorer(logger *zap.Logger) *Scorer {
	scorer, _ := NewScorer(DefaultConfig(), logger)
	return scorer
}

// AnalyzeRecord analyzes a single record and updates scores
func (s *Scorer) AnalyzeRecord(sourceID string, record map[string]interface{}, metadata RecordMetadata) *RecordAnalysis {
	s.mu.Lock()
	defer s.mu.Unlock()

	analysis := &RecordAnalysis{
		RecordID:      metadata.RecordID,
		Timestamp:     time.Now(),
		FieldAnalyses: make(map[string]*FieldAnalysis),
		Issues:        make([]QualityIssue, 0),
	}

	// Get or create source analysis
	sourceAnalysis, exists := s.sourceAnalyses[sourceID]
	if !exists {
		sourceAnalysis = &SourceAnalysis{
			SourceID:          sourceID,
			WindowStart:       time.Now(),
			FieldStats:        make(map[string]*FieldStats),
			ConsistencyChecks: make(map[string]*ConsistencyCheck),
		}
		s.sourceAnalyses[sourceID] = sourceAnalysis
	}

	sourceAnalysis.TotalRecords++
	sourceAnalysis.LastUpdate = time.Now()

	// Analyze each field
	var totalFields, validFields int
	for fieldName, value := range record {
		fieldAnalysis := s.analyzeField(sourceAnalysis, fieldName, value)
		analysis.FieldAnalyses[fieldName] = fieldAnalysis

		totalFields++
		if fieldAnalysis.Validity > 0 {
			validFields++
		}
	}

	// Calculate completeness
	analysis.Completeness = s.calculateCompleteness(analysis.FieldAnalyses)
	analysis.Validity = s.calculateValidity(totalFields, validFields)
	analysis.Consistency = s.calculateConsistency(sourceAnalysis)
	analysis.Timeliness = s.calculateTimeliness(metadata)

	// Calculate overall score
	analysis.OverallScore = s.calculateOverallScore(analysis.Completeness, analysis.Validity, analysis.Consistency, analysis.Timeliness)

	// Detect issues
	analysis.Issues = s.detectIssues(sourceID, analysis)

	// Update metrics
	if analysis.Validity >= s.config.ValidityThreshold {
		sourceAnalysis.ValidRecords++
	} else {
		sourceAnalysis.InvalidRecords++
	}

	return analysis
}

// analyzeField analyzes a single field
func (s *Scorer) analyzeField(sourceAnalysis *SourceAnalysis, fieldName string, value interface{}) *FieldAnalysis {
	// Get or create field stats
	fieldStats, exists := sourceAnalysis.FieldStats[fieldName]
	if !exists {
		fieldStats = &FieldStats{
			FieldName:   fieldName,
			MinValue:    math.MaxFloat64,
			MaxValue:    -math.MaxFloat64,
		}
		sourceAnalysis.FieldStats[fieldName] = fieldStats
	}

	fieldStats.TotalCount++
	fieldStats.LastUpdated = time.Now()

	// Check for null/empty
	isNull := value == nil
	isEmpty := false

	switch v := value.(type) {
	case string:
		isEmpty = len(v) == 0
	case []interface{}:
		isEmpty = len(v) == 0
	case map[string]interface{}:
		isEmpty = len(v) == 0
	}

	if isNull {
		fieldStats.NullCount++
	} else if isEmpty {
		fieldStats.EmptyCount++
	} else {
		fieldStats.ValidCount++
		
		// Update numeric statistics
		if num, ok := value.(float64); ok {
			fieldStats.SumValue += num
			fieldStats.SumSquares += num * num
			if num < fieldStats.MinValue {
				fieldStats.MinValue = num
			}
			if num > fieldStats.MaxValue {
				fieldStats.MaxValue = num
			}
		}
	}

	fieldStats.LastValue = value

	// Calculate field scores
	completeness := float64(fieldStats.TotalCount-fieldStats.NullCount-fieldStats.EmptyCount) / float64(fieldStats.TotalCount)
	validity := float64(fieldStats.ValidCount) / float64(fieldStats.TotalCount)

	return &FieldAnalysis{
		FieldName:    fieldName,
		TotalCount:   int(fieldStats.TotalCount),
		NullCount:    int(fieldStats.NullCount),
		EmptyCount:   int(fieldStats.EmptyCount),
		ValidCount:   int(fieldStats.ValidCount),
		InvalidCount: int(fieldStats.NullCount + fieldStats.EmptyCount),
		Completeness: completeness,
		Validity:     validity,
		LastUpdated:  fieldStats.LastUpdated,
	}
}

// calculateCompleteness calculates the completeness score
func (s *Scorer) calculateCompleteness(fieldAnalyses map[string]*FieldAnalysis) float64 {
	if len(fieldAnalyses) == 0 {
		return 1.0
	}

	var totalCompleteness float64
	for _, fa := range fieldAnalyses {
		totalCompleteness += fa.Completeness
	}

	return totalCompleteness / float64(len(fieldAnalyses))
}

// calculateValidity calculates the validity score
func (s *Scorer) calculateValidity(totalFields, validFields int) float64 {
	if totalFields == 0 {
		return 1.0
	}
	return float64(validFields) / float64(totalFields)
}

// calculateConsistency calculates the consistency score
func (s *Scorer) calculateConsistency(sourceAnalysis *SourceAnalysis) float64 {
	if len(sourceAnalysis.ConsistencyChecks) == 0 {
		return 1.0
	}

	var totalPassRate float64
	for _, check := range sourceAnalysis.ConsistencyChecks {
		if check.TotalChecks > 0 {
			passRate := float64(check.PassedChecks) / float64(check.TotalChecks)
			totalPassRate += passRate
		}
	}

	return totalPassRate / float64(len(sourceAnalysis.ConsistencyChecks))
}

// calculateTimeliness calculates the timeliness score
func (s *Scorer) calculateTimeliness(metadata RecordMetadata) float64 {
	if metadata.EventTime.IsZero() {
		return 0.5 // Default to 50% if no event time provided
	}

	latency := time.Since(metadata.EventTime)
	maxAcceptableLatency := 5 * time.Minute

	if latency <= maxAcceptableLatency {
		return 1.0 - (float64(latency) / float64(maxAcceptableLatency) * 0.2)
	}

	return 0.0
}

// calculateOverallScore calculates the weighted overall score
func (s *Scorer) calculateOverallScore(completeness, validity, consistency, timeliness float64) float64 {
	weights := s.config.Weights.Normalize()
	return completeness*weights.Completeness +
		validity*weights.Validity +
		consistency*weights.Consistency +
		timeliness*weights.Timeliness
}

// detectIssues detects quality issues based on scores
func (s *Scorer) detectIssues(sourceID string, analysis *RecordAnalysis) []QualityIssue {
	var issues []QualityIssue

	// Check completeness threshold
	if analysis.Completeness < s.config.CompletenessThreshold {
		issues = append(issues, QualityIssue{
			IssueID:     fmt.Sprintf("issue_%d", time.Now().UnixNano()),
			Timestamp:   time.Now(),
			Severity:    "warning",
			Category:    "completeness",
			Field:       "all",
			Description: fmt.Sprintf("Completeness score %f below threshold %f", analysis.Completeness, s.config.CompletenessThreshold),
		})
	}

	// Check validity threshold
	if analysis.Validity < s.config.ValidityThreshold {
		issues = append(issues, QualityIssue{
			IssueID:     fmt.Sprintf("issue_%d", time.Now().UnixNano()),
			Timestamp:   time.Now(),
			Severity:    "critical",
			Category:    "validity",
			Field:       "all",
			Description: fmt.Sprintf("Validity score %f below threshold %f", analysis.Validity, s.config.ValidityThreshold),
		})
	}

	// Check overall threshold
	if analysis.OverallScore < s.config.OverallThreshold {
		issues = append(issues, QualityIssue{
			IssueID:     fmt.Sprintf("issue_%d", time.Now().UnixNano()),
			Timestamp:   time.Now(),
			Severity:    "warning",
			Category:    "overall",
			Field:       "all",
			Description: fmt.Sprintf("Overall quality score %f below threshold %f", analysis.OverallScore, s.config.OverallThreshold),
		})
	}

	return issues
}

// GetScore returns the current quality score for a source
func (s *Scorer) GetScore(sourceID string) *QualityScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sourceAnalysis, exists := s.sourceAnalyses[sourceID]
	if !exists {
		return &QualityScore{
			SourceID:        sourceID,
			Timestamp:       time.Now(),
			OverallScore:    0.0,
			DimensionScores: make(map[Dimension]float64),
			ComponentScores: make(map[string]float64),
			WeightConfig:    s.config.Weights,
		}
	}

	dimensionScores := map[Dimension]float64{
		DimensionCompleteness: s.calculateCompletenessFromStats(sourceAnalysis),
		DimensionValidity:     s.calculateValidityFromStats(sourceAnalysis),
		DimensionConsistency:  s.calculateConsistency(sourceAnalysis),
		DimensionTimeliness:  s.calculateTimelinessFromStats(sourceAnalysis),
	}

	componentScores := make(map[string]float64)
	for fieldName, fieldStats := range sourceAnalysis.FieldStats {
		componentScores[fieldName] = (fieldStats.ValidCount + fieldStats.TotalCount - fieldStats.NullCount - fieldStats.EmptyCount) / float64(fieldStats.TotalCount)
	}

	return &QualityScore{
		SourceID:        sourceID,
		Timestamp:       time.Now(),
		OverallScore:    dimensionScores[DimensionCompleteness]*s.config.Weights.Completeness +
			dimensionScores[DimensionValidity]*s.config.Weights.Validity +
			dimensionScores[DimensionConsistency]*s.config.Weights.Consistency +
			dimensionScores[DimensionTimeliness]*s.config.Weights.Timeliness,
		DimensionScores: dimensionScores,
		ComponentScores: componentScores,
		WeightConfig:    s.config.Weights,
	}
}

// Helper methods for calculating scores from statistics
func (s *Scorer) calculateCompletenessFromStats(sourceAnalysis *SourceAnalysis) float64 {
	if len(sourceAnalysis.FieldStats) == 0 {
		return 1.0
	}

	var totalCompleteness float64
	for _, stats := range sourceAnalysis.FieldStats {
		completeness := float64(stats.TotalCount-stats.NullCount-stats.EmptyCount) / float64(stats.TotalCount)
		totalCompleteness += completeness
	}

	return totalCompleteness / float64(len(sourceAnalysis.FieldStats))
}

func (s *Scorer) calculateValidityFromStats(sourceAnalysis *SourceAnalysis) float64 {
	if sourceAnalysis.TotalRecords == 0 {
		return 1.0
	}
	return float64(sourceAnalysis.ValidRecords) / float64(sourceAnalysis.TotalRecords)
}

func (s *Scorer) calculateTimelinessFromStats(sourceAnalysis *SourceAnalysis) float64 {
	if sourceAnalysis.TotalRecords == 0 {
		return 1.0
	}
	avgLatency := sourceAnalysis.TotalLatency / time.Duration(sourceAnalysis.TotalRecords)
	maxAcceptableLatency := 5 * time.Minute

	if avgLatency <= maxAcceptableLatency {
		return 1.0 - (float64(avgLatency) / float64(maxAcceptableLatency) * 0.2)
	}

	return 0.0
}

// RecordMetadata contains metadata for a record
type RecordMetadata struct {
	RecordID    string
	EventTime   time.Time
	IngestTime  time.Time
	SourceType  string
	SessionID   string
	RequestID   string
}

// AddConsistencyCheck adds a consistency check result
func (s *Scorer) AddConsistencyCheck(sourceID, checkName string, passed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceAnalysis, exists := s.sourceAnalyses[sourceID]
	if !exists {
		sourceAnalysis = &SourceAnalysis{
			SourceID:          sourceID,
			FieldStats:        make(map[string]*FieldStats),
			ConsistencyChecks: make(map[string]*ConsistencyCheck),
		}
		s.sourceAnalyses[sourceID] = sourceAnalysis
	}

	check, exists := sourceAnalysis.ConsistencyChecks[checkName]
	if !exists {
		check = &ConsistencyCheck{
			CheckName: checkName,
		}
		sourceAnalysis.ConsistencyChecks[checkName] = check
	}

	check.TotalChecks++
	check.LastUpdated = time.Now()
	if passed {
		check.PassedChecks++
	} else {
		check.FailedChecks++
	}
	check.LastCheckResult = passed
}

// UpdateLatency updates the latency statistics
func (s *Scorer) UpdateLatency(sourceID string, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sourceAnalysis, exists := s.sourceAnalyses[sourceID]
	if !exists {
		sourceAnalysis = &SourceAnalysis{
			SourceID: sourceID,
		}
		s.sourceAnalyses[sourceID] = sourceAnalysis
	}

	sourceAnalysis.TotalLatency += latency
}
