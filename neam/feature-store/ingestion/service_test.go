package ingestion_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/feature-store/ingestion"
)

// MockProcessor is a mock implementation of the processor interface
type MockProcessor struct {
	mock.Mock
}

func (m *MockProcessor) Process(ctx context.Context, data map[string]interface{}) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockProcessor) GetProcessorType() string {
	args := m.Called()
	return args.String(0)
}

// TestIngestionService creation and initialization
func TestNewIngestionService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		config := ingestion.ServiceConfig{
			BatchSize:          100,
			FlushInterval:      time.Second * 30,
			MaxConcurrent:      10,
			EnableDeduplication: true,
			BufferSize:         1000,
		}

		service, err := ingestion.NewIngestionService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, 100, service.GetBatchSize())
		assert.Equal(t, time.Second*30, service.GetFlushInterval())
		assert.True(t, service.IsDeduplicationEnabled())
	})

	t.Run("creates service with default config", func(t *testing.T) {
		service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{})
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("fails with invalid batch size", func(t *testing.T) {
		config := ingestion.ServiceConfig{
			BatchSize: 0,
		}
		service, err := ingestion.NewIngestionService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

// Test processor registration
func TestServiceRegisterProcessor(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{})
	require.NoError(t, err)

	mockProcessor := new(MockProcessor)
	mockProcessor.On("GetProcessorType").Return("batch")

	err = service.RegisterProcessor(mockProcessor)
	require.NoError(t, err)

	processors := service.GetRegisteredProcessors()
	assert.Contains(t, processors, "batch")
}

func TestServiceRegisterDuplicateProcessor(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{})
	require.NoError(t, err)

	mockProcessor1 := new(MockProcessor)
	mockProcessor1.On("GetProcessorType").Return("duplicate")

	mockProcessor2 := new(MockProcessor)
	mockProcessor2.On("GetProcessorType").Return("duplicate")

	err = service.RegisterProcessor(mockProcessor1)
	require.NoError(t, err)

	err = service.RegisterProcessor(mockProcessor2)
	assert.Error(t, err)
}

// Test data ingestion
func TestServiceIngestData(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		BatchSize: 10,
	})
	require.NoError(t, err)

	mockProcessor := new(MockProcessor)
	mockProcessor.On("GetProcessorType").Return("test_processor")
	mockProcessor.On("Process", mock.Anything, mock.Anything).Return(nil)

	err = service.RegisterProcessor(mockProcessor)
	require.NoError(t, err)

	testData := map[string]interface{}{
		"feature_name":  "transaction_amount",
		"feature_value": 15000.0,
		"entity_id":     "entity_123",
	}

	result, err := service.IngestData(context.Background(), testData)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.RecordID)
	assert.Equal(t, "entity_123", result.EntityID)
	assert.Equal(t, "test_processor", result.ProcessorType)
}

func TestServiceIngestBatchData(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		BatchSize: 10,
	})
	require.NoError(t, err)

	mockProcessor := new(MockProcessor)
	mockProcessor.On("GetProcessorType").Return("batch_processor")
	mockProcessor.On("Process", mock.Anything, mock.Anything).Return(nil)

	err = service.RegisterProcessor(mockProcessor)
	require.NoError(t, err)

	batchData := []map[string]interface{}{
		{"feature_name": "feature_1", "entity_id": "entity_1"},
		{"feature_name": "feature_2", "entity_id": "entity_2"},
		{"feature_name": "feature_3", "entity_id": "entity_3"},
	}

	results, err := service.IngestBatchData(context.Background(), batchData)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	for _, result := range results {
		assert.NotEmpty(t, result.RecordID)
	}
}

// Test deduplication
func TestServiceDeduplication(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		EnableDeduplication: true,
		DeduplicationKeys:   []string{"entity_id", "feature_name"},
	})
	require.NoError(t, err)

	testData := map[string]interface{}{
		"feature_name":  "transaction_amount",
		"feature_value": 15000.0,
		"entity_id":     "entity_123",
	}

	// First ingestion
	result1, err := service.IngestData(context.Background(), testData)
	require.NoError(t, err)

	// Duplicate ingestion
	result2, err := service.IngestData(context.Background(), testData)
	require.NoError(t, err)

	// Should have same record ID (deduplicated)
	assert.Equal(t, result1.RecordID, result2.RecordID)
}

// Test buffer management
func TestServiceBufferStatus(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		BufferSize: 100,
		BatchSize:  10,
	})
	require.NoError(t, err)

	// Initially empty
	status := service.GetBufferStatus()
	assert.Equal(t, 0, status.CurrentSize)
	assert.Equal(t, 100, status.Capacity)

	// Add some data
	for i := 0; i < 50; i++ {
		service.IngestData(context.Background(), map[string]interface{}{
			"entity_id": "entity_" + string(rune('0'+i%10)),
			"feature":   i,
		})
	}

	status = service.GetBufferStatus()
	assert.Equal(t, 50, status.CurrentSize)
}

// Test flush trigger
func TestServiceFlush(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		BatchSize:     5,
		FlushInterval: time.Hour, // Long interval so we can manually flush
	})
	require.NoError(t, err)

	mockProcessor := new(MockProcessor)
	mockProcessor.On("GetProcessorType").Return("flush_processor")
	mockProcessor.On("Process", mock.Anything, mock.Anything).Return(nil)

	err = service.RegisterProcessor(mockProcessor)
	require.NoError(t, err)

	// Add data
	for i := 0; i < 3; i++ {
		service.IngestData(context.Background(), map[string]interface{}{
			"entity_id": "entity_flush",
			"feature":   i,
		})
	}

	// Flush manually
	flushed, err := service.Flush(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, flushed)

	// Buffer should be empty
	status := service.GetBufferStatus()
	assert.Equal(t, 0, status.CurrentSize)
}

// Test schema validation
func TestServiceValidateSchema(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		EnableSchemaValidation: true,
		RequiredFields:         []string{"entity_id", "feature_name"},
	})
	require.NoError(t, err)

	testCases := []struct {
		name      string
		data      map[string]interface{}
		expectErr bool
	}{
		{
			name: "valid data",
			data: map[string]interface{}{
				"entity_id":    "entity_123",
				"feature_name": "feature_1",
				"value":        100.0,
			},
			expectErr: false,
		},
		{
			name: "missing entity_id",
			data: map[string]interface{}{
				"feature_name": "feature_1",
				"value":        100.0,
			},
			expectErr: true,
		},
		{
			name: "missing feature_name",
			data: map[string]interface{}{
				"entity_id": "entity_123",
				"value":     100.0,
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateSchema(tc.data)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test statistics
func TestServiceGetStatistics(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{})
	require.NoError(t, err)

	// Perform some ingestions
	for i := 0; i < 10; i++ {
		service.IngestData(context.Background(), map[string]interface{}{
			"entity_id": "entity_stats",
			"feature":   i,
		})
	}

	stats := service.GetStatistics()
	assert.Equal(t, int64(10), stats.TotalIngested)
	assert.Equal(t, int64(0), stats.TotalErrors)
	assert.Equal(t, int64(0), stats.TotalFlushed)
}

// Test JSON serialization
func TestIngestionRecordSerialization(t *testing.T) {
	record := ingestion.IngestionRecord{
		RecordID:     "record_001",
		EntityID:     "entity_123",
		FeatureName:  "transaction_amount",
		FeatureValue: 15000.0,
		Timestamp:    time.Now(),
		ProcessorType: "test",
		Metadata: map[string]interface{}{
			"source": "bank",
		},
	}

	data, err := json.Marshal(record)
	require.NoError(t, err)

	var decoded ingestion.IngestionRecord
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, record.RecordID, decoded.RecordID)
	assert.Equal(t, record.EntityID, decoded.EntityID)
	assert.Equal(t, record.FeatureName, decoded.FeatureName)
	assert.Equal(t, record.FeatureValue, decoded.FeatureValue)
}

// Test graceful shutdown
func TestServiceClose(t *testing.T) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{})
	require.NoError(t, err)

	err = service.Close()
	require.NoError(t, err)

	// Operations after close should fail
	_, err = service.IngestData(context.Background(), map[string]interface{}{"entity_id": "test"})
	assert.Error(t, err)
}

// Test with real processor implementations
func TestWithRealProcessorImplementations(t *testing.T) {
	t.Run("batch processor", func(t *testing.T) {
		processor := ingestion.NewBatchProcessor(ingestion.BatchConfig{
			BatchSize:     100,
			FlushInterval: time.Second,
		})
		assert.Equal(t, "batch", processor.GetProcessorType())

		ctx := context.Background()
		err := processor.Process(ctx, map[string]interface{}{"value": 100})
		require.NoError(t, err)
	})

	t.Run("streaming processor", func(t *testing.T) {
		processor := ingestion.NewStreamingProcessor(ingestion.StreamingConfig{
			WindowSize:  60,
			Aggregation: "avg",
		})
		assert.Equal(t, "streaming", processor.GetProcessorType())

		ctx := context.Background()
		err := processor.Process(ctx, map[string]interface{}{"value": 100})
		require.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkIngestData(b *testing.B) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		BatchSize: 1000,
	})
	if err != nil {
		b.Fatal(err)
	}

	processor := ingestion.NewBatchProcessor(ingestion.BatchConfig{BatchSize: 1000})
	service.RegisterProcessor(processor)

	ctx := context.Background()
	data := map[string]interface{}{
		"entity_id":    "benchmark_entity",
		"feature_name": "benchmark_feature",
		"feature_value": 100.0,
		"timestamp":    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.IngestData(ctx, data)
	}
}

func BenchmarkIngestBatchData(b *testing.B) {
	service, err := ingestion.NewIngestionService(ingestion.ServiceConfig{
		BatchSize: 100,
	})
	if err != nil {
		b.Fatal(err)
	}

	processor := ingestion.NewBatchProcessor(ingestion.BatchConfig{BatchSize: 100})
	service.RegisterProcessor(processor)

	ctx := context.Background()
	batch := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		batch[i] = map[string]interface{}{
			"entity_id":    "entity_" + string(rune('0'+i%10)),
			"feature_name": "feature_" + string(rune('0'+i%10)),
			"value":        float64(i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.IngestBatchData(ctx, batch)
	}
}
