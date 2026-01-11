package dataquality

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWeightConfig_Validate(t *testing.T) {
	tests := []struct {
		name   string
		config WeightConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: WeightConfig{
				Completeness: 0.25,
				Validity:     0.35,
				Consistency:  0.20,
				Timeliness:  0.20,
			},
			valid: true,
		},
		{
			name: "invalid - doesn't sum to 1",
			config: WeightConfig{
				Completeness: 0.50,
				Validity:     0.50,
				Consistency:  0.50,
				Timeliness:  0.50,
			},
			valid: false,
		},
		{
			name: "invalid - all zeros",
			config: WeightConfig{
				Completeness: 0,
				Validity:     0,
				Consistency:  0,
				Timeliness:  0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected valid config, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected invalid config, got no error")
			}
		})
	}
}

func TestWeightConfig_Normalize(t *testing.T) {
	config := WeightConfig{
		Completeness: 1,
		Validity:     1,
		Consistency:  1,
		Timeliness:  1,
	}

	normalized := config.Normalize()

	// Each should be 0.25
	expected := 0.25
	if normalized.Completeness != expected {
		t.Errorf("Expected completeness %f, got %f", expected, normalized.Completeness)
	}
	if normalized.Validity != expected {
		t.Errorf("Expected validity %f, got %f", expected, normalized.Validity)
	}
	if normalized.Consistency != expected {
		t.Errorf("Expected consistency %f, got %f", expected, normalized.Consistency)
	}
	if normalized.Timeliness != expected {
		t.Errorf("Expected timeliness %f, got %f", expected, normalized.Timeliness)
	}
}

func TestScorer_AnalyzeRecord(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	record := map[string]interface{}{
		"id":    "test-123",
		"name":  "Test Item",
		"value": 42.5,
		"tags":  []interface{}{"a", "b"},
	}

	metadata := RecordMetadata{
		RecordID:  "rec-001",
		EventTime: time.Now(),
	}

	analysis := scorer.AnalyzeRecord("test-source", record, metadata)

	if analysis.RecordID != "rec-001" {
		t.Errorf("Expected record ID 'rec-001', got '%s'", analysis.RecordID)
	}

	if analysis.Completeness < 0.9 {
		t.Errorf("Expected high completeness, got %f", analysis.Completeness)
	}

	if analysis.Validity < 0.9 {
		t.Errorf("Expected high validity, got %f", analysis.Validity)
	}

	if analysis.OverallScore < 0.8 {
		t.Errorf("Expected high overall score, got %f", analysis.OverallScore)
	}
}

func TestScorer_AnalyzeRecordWithNulls(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	record := map[string]interface{}{
		"id":    "test-123",
		"name":  nil, // Null field
		"value": 42.5,
	}

	metadata := RecordMetadata{
		RecordID:  "rec-002",
		EventTime: time.Now(),
	}

	analysis := scorer.AnalyzeRecord("test-source", record, metadata)

	if analysis.Completeness > 0.8 {
		t.Errorf("Expected lower completeness due to null field, got %f", analysis.Completeness)
	}
}

func TestScorer_GetScore(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	// Analyze some records
	for i := 0; i < 10; i++ {
		record := map[string]interface{}{
			"id":    "test-123",
			"value": float64(i),
		}
		metadata := RecordMetadata{
			RecordID:  "rec-001",
			EventTime: time.Now(),
		}
		scorer.AnalyzeRecord("test-source", record, metadata)
	}

	score := scorer.GetScore("test-source")

	if score.SourceID != "test-source" {
		t.Errorf("Expected source ID 'test-source', got '%s'", score.SourceID)
	}

	if score.OverallScore < 0 {
		t.Errorf("Expected non-negative score, got %f", score.OverallScore)
	}

	if score.OverallScore > 100 {
		t.Errorf("Expected score <= 100, got %f", score.OverallScore)
	}

	if len(score.DimensionScores) != 4 {
		t.Errorf("Expected 4 dimension scores, got %d", len(score.DimensionScores))
	}
}

func TestScorer_GetScore_NonExistentSource(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	score := scorer.GetScore("non-existent")

	if score.OverallScore != 0 {
		t.Errorf("Expected zero score for non-existent source, got %f", score.OverallScore)
	}
}

func TestScorer_AddConsistencyCheck(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	// Add consistency checks
	scorer.AddConsistencyCheck("test-source", "check-1", true)
	scorer.AddConsistencyCheck("test-source", "check-1", true)
	scorer.AddConsistencyCheck("test-source", "check-1", false) // One failure

	score := scorer.GetScore("test-source")

	if score.DimensionScores[DimensionConsistency] < 0.6 {
		t.Errorf("Expected consistency around 0.67 due to 2/3 passing, got %f", score.DimensionScores[DimensionConsistency])
	}
}

func TestScorer_UpdateLatency(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	// Update latency
	scorer.UpdateLatency("test-source", 100*time.Millisecond)
	scorer.UpdateLatency("test-source", 200*time.Millisecond)
	scorer.UpdateLatency("test-source", 300*time.Millisecond)

	score := scorer.GetScore("test-source")

	if score.DimensionScores[DimensionTimeliness] < 0.8 {
		t.Errorf("Expected high timeliness score, got %f", score.DimensionScores[DimensionTimeliness])
	}
}

func TestScorer_DetectIssues(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultConfig()
	config.ValidityThreshold = 0.99 // Set high threshold
	scorer, _ := NewScorer(config, logger)

	// Analyze a record with missing field (will lower validity)
	record := map[string]interface{}{
		"value": 42.5,
	}
	metadata := RecordMetadata{
		RecordID:  "rec-001",
		EventTime: time.Now(),
	}

	analysis := scorer.AnalyzeRecord("test-source", record, metadata)

	if len(analysis.Issues) == 0 {
		t.Error("Expected quality issues to be detected")
	}

	// Check if validity issue was detected
	foundValidityIssue := false
	for _, issue := range analysis.Issues {
		if issue.Category == "validity" {
			foundValidityIssue = true
			break
		}
	}

	if !foundValidityIssue {
		t.Error("Expected validity category issue")
	}
}

func TestScorer_GetAllScores(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	// Analyze records from multiple sources
	for i := 0; i < 5; i++ {
		record := map[string]interface{}{
			"id": "test",
		}
		metadata := RecordMetadata{EventTime: time.Now()}
		scorer.AnalyzeRecord("source-1", record, metadata)
		scorer.AnalyzeRecord("source-2", record, metadata)
	}

	scores := scorer.GetAllScores()

	if len(scores) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(scores))
	}

	if _, exists := scores["source-1"]; !exists {
		t.Error("Expected source-1 in scores")
	}

	if _, exists := scores["source-2"]; !exists {
		t.Error("Expected source-2 in scores")
	}
}

func TestFieldAnalysis_Statistics(t *testing.T) {
	logger := zap.NewNop()
	scorer := NewDefaultScorer(logger)

	// Analyze multiple records for the same field
	for i := 0; i < 100; i++ {
		record := map[string]interface{}{
			"value": float64(i),
		}
		metadata := RecordMetadata{EventTime: time.Now()}
		scorer.AnalyzeRecord("test-source", record, metadata)
	}

	score := scorer.GetScore("test-source")
	componentScore := score.ComponentScores["value"]

	if componentScore < 0.99 {
		t.Errorf("Expected high component score for consistent field, got %f", componentScore)
	}
}
