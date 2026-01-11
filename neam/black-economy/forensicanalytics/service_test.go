package forensicanalytics_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/black-economy/forensicanalytics"
)

// MockAnalyzer is a mock implementation of the analyzer interface
type MockAnalyzer struct {
	mock.Mock
}

func (m *MockAnalyzer) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAnalyzer) Analyze(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAnalyzer) GetAnalyzerType() string {
	args := m.Called()
	return args.String(0)
}

// TestForensicAnalyticsService creation
func TestNewForensicAnalyticsService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		config := forensicanalytics.ServiceConfig{
			AnalysisTimeout:   time.Minute * 5,
			MaxConcurrent:     10,
			EnableParallelism: true,
			ResultTTL:         time.Hour * 24,
		}

		service, err := forensicanalytics.NewForensicAnalyticsService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, time.Minute*5, service.GetAnalysisTimeout())
		assert.True(t, service.IsParallelismEnabled())
	})

	t.Run("creates service with default config", func(t *testing.T) {
		service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("fails with invalid timeout", func(t *testing.T) {
		config := forensicanalytics.ServiceConfig{
			AnalysisTimeout: 0,
		}
		service, err := forensicanalytics.NewForensicAnalyticsService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

// Test analyzer registration
func TestServiceRegisterAnalyzer(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	mockAnalyzer := new(MockAnalyzer)
	mockAnalyzer.On("Name").Return("pattern_analyzer")
	mockAnalyzer.On("GetAnalyzerType").Return("pattern")

	err = service.RegisterAnalyzer(mockAnalyzer)
	require.NoError(t, err)

	analyzers := service.GetRegisteredAnalyzers()
	assert.Contains(t, analyzers, "pattern_analyzer")
}

func TestServiceRegisterDuplicateAnalyzer(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	mockAnalyzer1 := new(MockAnalyzer)
	mockAnalyzer1.On("Name").Return("duplicate_analyzer")
	mockAnalyzer1.On("GetAnalyzerType").Return("pattern")

	mockAnalyzer2 := new(MockAnalyzer)
	mockAnalyzer2.On("Name").Return("duplicate_analyzer")
	mockAnalyzer2.On("GetAnalyzerType").Return("statistical")

	err = service.RegisterAnalyzer(mockAnalyzer1)
	require.NoError(t, err)

	err = service.RegisterAnalyzer(mockAnalyzer2)
	assert.Error(t, err)
}

// Test case creation
func TestServiceCreateCase(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	caseInput := forensicanalytics.CaseInput{
		CaseType:    "fraud_investigation",
		SubjectID:   "subject_123",
		SubjectType: "entity",
		Description: "Investigation of suspicious transactions",
		Priority:    forensicanalytics.PriorityHigh,
		Metadata: map[string]interface{}{
			"reported_by": "analyst_1",
			"region":      "EMEA",
		},
	}

	caseResult, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)
	require.NotNil(t, caseResult)
	assert.NotEmpty(t, caseResult.CaseID)
	assert.Equal(t, "fraud_investigation", caseResult.CaseType)
	assert.Equal(t, forensicanalytics.StatusOpen, caseResult.Status)
	assert.Equal(t, forensicanalytics.PriorityHigh, caseResult.Priority)
}

func TestServiceCreateCaseWithInvalidInput(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	testCases := []struct {
		name  string
		input forensicanalytics.CaseInput
	}{
		{
			name: "empty case type",
			input: forensicanalytics.CaseInput{
				CaseType: "",
				SubjectID: "subject_123",
			},
		},
		{
			name: "empty subject ID",
			input: forensicanalytics.CaseInput{
				CaseType: "fraud_investigation",
				SubjectID: "",
			},
		},
		{
			name: "invalid priority",
			input: forensicanalytics.CaseInput{
				CaseType:    "fraud_investigation",
				SubjectID:   "subject_123",
				Priority:    forensicanalytics.Priority(-1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.CreateCase(context.Background(), tc.input)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

// Test evidence management
func TestServiceAddEvidence(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case first
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "fraud_investigation",
		SubjectID: "subject_123",
	}
	caseResult, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Add evidence
	evidence := forensicanalytics.EvidenceInput{
		CaseID:     caseResult.CaseID,
		EvidenceType: forensicanalytics.EvidenceTypeTransaction,
		Content:     map[string]interface{}{"transaction_id": "tx_456"},
		Source:      "bank_records",
		Timestamp:   time.Now(),
	}

	evidenceResult, err := service.AddEvidence(context.Background(), evidence)
	require.NoError(t, err)
	require.NotNil(t, evidenceResult)
	assert.NotEmpty(t, evidenceResult.EvidenceID)
	assert.Equal(t, caseResult.CaseID, evidenceResult.CaseID)
	assert.Equal(t, forensicanalytics.EvidenceTypeTransaction, evidenceResult.EvidenceType)
}

func TestServiceAddEvidenceToNonExistentCase(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	evidence := forensicanalytics.EvidenceInput{
		CaseID:       "non_existent_case",
		EvidenceType: forensicanalytics.EvidenceTypeDocument,
		Content:      map[string]interface{}{"doc_id": "doc_789"},
	}

	result, err := service.AddEvidence(context.Background(), evidence)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "case not found")
}

// Test case analysis
func TestServiceAnalyzeCase(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{
		EnableParallelism: true,
	})
	require.NoError(t, err)

	// Register mock analyzer
	mockAnalyzer := new(MockAnalyzer)
	mockAnalyzer.On("Name").Return("test_analyzer")
	mockAnalyzer.On("GetAnalyzerType").Return("pattern")
	mockAnalyzer.On("Analyze", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"pattern_detected": true,
		"risk_score":       0.75,
	}, nil)

	err = service.RegisterAnalyzer(mockAnalyzer)
	require.NoError(t, err)

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "pattern_analysis",
		SubjectID: "subject_456",
	}
	caseResult, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Add some evidence
	evidence := forensicanalytics.EvidenceInput{
		CaseID:       caseResult.CaseID,
		EvidenceType: forensicanalytics.EvidenceTypeTransaction,
		Content:      map[string]interface{}{"transactions": []map[string]interface{}{}},
	}
	_, err = service.AddEvidence(context.Background(), evidence)
	require.NoError(t, err)

	// Analyze the case
	analysisResult, err := service.AnalyzeCase(context.Background(), caseResult.CaseID)
	require.NoError(t, err)
	require.NotNil(t, analysisResult)
	assert.Equal(t, caseResult.CaseID, analysisResult.CaseID)
	assert.NotEmpty(t, analysisResult.AnalysisID)
	assert.Equal(t, forensicanalytics.StatusAnalyzing, analysisResult.Status)
}

func TestServiceAnalyzeCaseWithNoAnalyzers(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	caseInput := forensicanalytics.CaseInput{
		CaseType:  "analysis",
		SubjectID: "subject_789",
	}
	caseResult, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	result, err := service.AnalyzeCase(context.Background(), caseResult.CaseID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no analyzers registered")
}

// Test case retrieval
func TestServiceGetCase(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:    "test_case",
		SubjectID:   "subject_001",
		Description: "Test case for retrieval",
	}
	createdCase, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Get the case
	retrievedCase, err := service.GetCase(context.Background(), createdCase.CaseID)
	require.NoError(t, err)
	assert.Equal(t, createdCase.CaseID, retrievedCase.CaseID)
	assert.Equal(t, "test_case", retrievedCase.CaseType)
	assert.Equal(t, "subject_001", retrievedCase.SubjectID)
}

func TestServiceGetNonExistentCase(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	result, err := service.GetCase(context.Background(), "non_existent_case")
	assert.Error(t, err)
	assert.Nil(t, result)
}

// Test case status updates
func TestServiceUpdateCaseStatus(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "status_test",
		SubjectID: "subject_002",
	}
	createdCase, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Update status
	updatedCase, err := service.UpdateCaseStatus(context.Background(), createdCase.CaseID, forensicanalytics.StatusInProgress)
	require.NoError(t, err)
	assert.Equal(t, forensicanalytics.StatusInProgress, updatedCase.Status)

	// Update again to closed
	closedCase, err := service.UpdateCaseStatus(context.Background(), createdCase.CaseID, forensicanalytics.StatusClosed)
	require.NoError(t, err)
	assert.Equal(t, forensicanalytics.StatusClosed, closedCase.Status)
}

func TestServiceUpdateCaseStatusInvalidTransition(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "transition_test",
		SubjectID: "subject_003",
	}
	createdCase, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Try to transition from open to closed directly (should fail)
	_, err = service.UpdateCaseStatus(context.Background(), createdCase.CaseID, forensicanalytics.StatusClosed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

// Test case search
func TestServiceSearchCases(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create multiple cases
	for i := 0; i < 5; i++ {
		_, err := service.CreateCase(context.Background(), forensicanalytics.CaseInput{
			CaseType:  "search_test",
			SubjectID: "subject_search_" + string(rune('0'+i)),
		})
		require.NoError(t, err)
	}

	// Search by case type
	results, err := service.SearchCases(context.Background(), forensicanalytics.CaseSearchQuery{
		CaseType: "search_test",
	})
	require.NoError(t, err)
	assert.Len(t, results, 5)

	// Search by status
	results, err = service.SearchCases(context.Background(), forensicanalytics.CaseSearchQuery{
		Status: forensicanalytics.StatusOpen,
	})
	require.NoError(t, err)
	assert.Len(t, results, 5)

	// Search with pagination
	results, err = service.SearchCases(context.Background(), forensicanalytics.CaseSearchQuery{
		CaseType: "search_test",
		Limit:    2,
		Offset:   1,
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

// Test analysis report generation
func TestServiceGenerateReport(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case with analysis
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "report_test",
		SubjectID: "subject_report",
	}
	caseResult, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Add evidence
	evidence := forensicanalytics.EvidenceInput{
		CaseID:       caseResult.CaseID,
		EvidenceType: forensicanalytics.EvidenceTypeDocument,
		Content:      map[string]interface{}{"document": "evidence_doc"},
	}
	_, err = service.AddEvidence(context.Background(), evidence)
	require.NoError(t, err)

	// Generate report
	report, err := service.GenerateReport(context.Background(), caseResult.CaseID, forensicanalytics.ReportFormatPDF)
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.NotEmpty(t, report.ReportID)
	assert.Equal(t, caseResult.CaseID, report.CaseID)
	assert.Equal(t, forensicanalytics.ReportFormatPDF, report.Format)
}

// Test timeline generation
func TestServiceGenerateTimeline(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "timeline_test",
		SubjectID: "subject_timeline",
	}
	caseResult, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Add evidence at different times
	baseTime := time.Now().Add(-time.Hour)
	for i := 0; i < 5; i++ {
		evidence := forensicanalytics.EvidenceInput{
			CaseID:       caseResult.CaseID,
			EvidenceType: forensicanalytics.EvidenceTypeTransaction,
			Content:      map[string]interface{}{"event": i},
			Timestamp:    baseTime.Add(time.Duration(i) * time.Minute),
		}
		_, err = service.AddEvidence(context.Background(), evidence)
		require.NoError(t, err)
	}

	// Generate timeline
	timeline, err := service.GenerateTimeline(context.Background(), caseResult.CaseID)
	require.NoError(t, err)
	require.NotNil(t, timeline)
	assert.Len(t, timeline.Events, 5)

	// Events should be sorted by timestamp
	for i := 1; i < len(timeline.Events); i++ {
		assert.True(t, timeline.Events[i-1].Timestamp.Before(timeline.Events[i].Timestamp))
	}
}

// Test JSON serialization
func TestForensicCaseSerialization(t *testing.T) {
	forensicCase := forensicanalytics.ForensicCase{
		CaseID:      "case_001",
		CaseType:    "fraud_investigation",
		SubjectID:   "subject_123",
		SubjectType: "entity",
		Status:      forensicanalytics.StatusOpen,
		Priority:    forensicanalytics.PriorityHigh,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"analyst": "analyst_1",
		},
	}

	data, err := json.Marshal(forensicCase)
	require.NoError(t, err)

	var decoded forensicanalytics.ForensicCase
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, forensicCase.CaseID, decoded.CaseID)
	assert.Equal(t, forensicCase.CaseType, decoded.CaseType)
	assert.Equal(t, forensicCase.Status, decoded.Status)
}

// Test statistics
func TestServiceGetStatistics(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create some cases
	for i := 0; i < 3; i++ {
		_, err := service.CreateCase(context.Background(), forensicanalytics.CaseInput{
			CaseType:  "stats_test",
			SubjectID: "subject_stats_" + string(rune('0'+i)),
		})
		require.NoError(t, err)
	}

	stats := service.GetStatistics()
	assert.Equal(t, int64(3), stats.TotalCases)
	assert.Equal(t, int64(3), stats.OpenCases)
	assert.Equal(t, int64(0), stats.ClosedCases)
}

// Test case deletion (soft delete)
func TestServiceDeleteCase(t *testing.T) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	require.NoError(t, err)

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "delete_test",
		SubjectID: "subject_delete",
	}
	createdCase, err := service.CreateCase(context.Background(), caseInput)
	require.NoError(t, err)

	// Delete the case
	err = service.DeleteCase(context.Background(), createdCase.CaseID)
	require.NoError(t, err)

	// Case should not be found with normal retrieval
	_, err = service.GetCase(context.Background(), createdCase.CaseID)
	assert.Error(t, err)

	// But should exist in deleted cases
	deletedCase, err := service.GetDeletedCase(context.Background(), createdCase.CaseID)
	require.NoError(t, err)
	assert.Equal(t, createdCase.CaseID, deletedCase.CaseID)
}

// Benchmark tests
func BenchmarkAnalyzeCase(b *testing.B) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{
		EnableParallelism: true,
	})
	if err != nil {
		b.Fatal(err)
	}

	// Register analyzers
	for i := 0; i < 5; i++ {
		analyzer := forensicanalytics.NewBasicAnalyzer("analyzer_" + string(rune('0'+i)))
		service.RegisterAnalyzer(analyzer)
	}

	// Create a case
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "benchmark",
		SubjectID: "subject_benchmark",
	}
	caseResult, service.CreateCase(context.Background(), caseInput)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.AnalyzeCase(ctx, caseResult.CaseID)
	}
}

func BenchmarkGenerateTimeline(b *testing.B) {
	service, err := forensicanalytics.NewForensicAnalyticsService(forensicanalytics.ServiceConfig{})
	if err != nil {
		b.Fatal(err)
	}

	// Create a case with evidence
	caseInput := forensicanalytics.CaseInput{
		CaseType:  "timeline_benchmark",
		SubjectID: "subject_timeline_bm",
	}
	caseResult, err := service.CreateCase(context.Background(), caseInput)
	if err != nil {
		b.Fatal(err)
	}

	baseTime := time.Now()
	for i := 0; i < 100; i++ {
		evidence := forensicanalytics.EvidenceInput{
			CaseID:       caseResult.CaseID,
			EvidenceType: forensicanalytics.EvidenceTypeTransaction,
			Content:      map[string]interface{}{"event": i},
			Timestamp:    baseTime.Add(time.Duration(i) * time.Minute),
		}
		service.AddEvidence(context.Background(), evidence)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GenerateTimeline(ctx, caseResult.CaseID)
	}
}
