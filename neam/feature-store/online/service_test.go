package online_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/feature-store/online"
)

// MockCacheClient is a mock implementation of the cache client interface
type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheClient) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestOnlineService creation and initialization
func TestNewOnlineService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		config := online.ServiceConfig{
			CacheType:        "redis",
			CacheSize:        10000,
			TTL:              time.Hour,
			EnableCompression: true,
			MaxLatency:       time.Millisecond * 100,
		}

		service, err := online.NewOnlineService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, "redis", service.GetCacheType())
		assert.Equal(t, time.Hour, service.GetTTL())
		assert.True(t, service.IsCompressionEnabled())
	})

	t.Run("creates service with default config", func(t *testing.T) {
		service, err := online.NewOnlineService(online.ServiceConfig{})
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("fails with invalid cache size", func(t *testing.T) {
		config := online.ServiceConfig{
			CacheSize: 0,
		}
		service, err := online.NewOnlineService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

// Test feature retrieval
func TestServiceGetFeature(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
		TTL:       time.Hour,
	})
	require.NoError(t, err)

	// Set a feature first
	featureData := map[string]interface{}{
		"value":     15000.0,
		"timestamp": time.Now(),
		"version":   "v1",
	}
	err = service.SetFeature(context.Background(), "entity_123", "transaction_amount", featureData)
	require.NoError(t, err)

	// Get the feature
	result, err := service.GetFeature(context.Background(), "entity_123", "transaction_amount")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 15000.0, result["value"])
	assert.Equal(t, "v1", result["version"])
}

func TestServiceGetNonExistentFeature(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{})
	require.NoError(t, err)

	result, err := service.GetFeature(context.Background(), "non_existent", "non_existent_feature")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

// Test feature retrieval for multiple entities
func TestServiceGetFeaturesForEntities(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
	})
	require.NoError(t, err)

	// Set features for multiple entities
	for i := 0; i < 3; i++ {
		err = service.SetFeature(context.Background(), "entity_"+string(rune('0'+i)), "feature_1", map[string]interface{}{
			"value": float64(i * 100),
		})
		require.NoError(t, err)
	}

	// Get features for all entities
	entities := []string{"entity_0", "entity_1", "entity_2"}
	features, err := service.GetFeaturesForEntities(context.Background(), entities, "feature_1")
	require.NoError(t, err)
	assert.Len(t, features, 3)

	for i, entity := range entities {
		assert.Contains(t, features, entity)
		assert.Equal(t, float64(i*100), features[entity]["value"])
	}
}

// Test feature retrieval with wildcards
func TestServiceGetFeaturesWithWildcard(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
	})
	require.NoError(t, err)

	// Set multiple features for an entity
	features := map[string]interface{}{
		"feature_1": map[string]interface{}{"value": 100},
		"feature_2": map[string]interface{}{"value": 200},
		"feature_3": map[string]interface{}{"value": 300},
	}
	err = service.SetAllFeatures(context.Background(), "entity_wildcard", features)
	require.NoError(t, err)

	// Get all features using wildcard
	result, err := service.GetFeaturesWithWildcard(context.Background(), "entity_wildcard", "*")
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "feature_1")
	assert.Contains(t, result, "feature_2")
	assert.Contains(t, result, "feature_3")
}

// Test feature versioning
func TestServiceFeatureVersioning(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
	})
	require.NoError(t, err)

	// Set initial version
	err = service.SetFeature(context.Background(), "entity_version", "feature", map[string]interface{}{
		"value":    100,
		"version":  "v1",
	})
	require.NoError(t, err)

	// Update to new version
	err = service.SetFeature(context.Background(), "entity_version", "feature", map[string]interface{}{
		"value":    150,
		"version":  "v2",
	})
	require.NoError(t, err)

	// Get should return latest version
	result, err := service.GetFeature(context.Background(), "entity_version", "feature")
	require.NoError(t, err)
	assert.Equal(t, 150.0, result["value"])
	assert.Equal(t, "v2", result["version"])

	// Get specific version
	specificResult, err := service.GetFeatureVersion(context.Background(), "entity_version", "feature", "v1")
	require.NoError(t, err)
	assert.Equal(t, 100.0, specificResult["value"])
}

// Test TTL functionality
func TestServiceFeatureTTL(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
		TTL:       time.Second * 1, // Very short TTL
	})
	require.NoError(t, err)

	// Set a feature
	err = service.SetFeature(context.Background(), "entity_ttl", "feature", map[string]interface{}{
		"value": 100,
	})
	require.NoError(t, err)

	// Should be retrievable immediately
	_, err = service.GetFeature(context.Background(), "entity_ttl", "feature")
	require.NoError(t, err)

	// Wait for TTL to expire
	time.Sleep(time.Second * 2)

	// Should not be retrievable
	_, err = service.GetFeature(context.Background(), "entity_ttl", "feature")
	assert.Error(t, err)
}

// Test cache eviction
func TestServiceCacheEviction(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 3, // Very small cache
		EvictionPolicy: "lru",
	})
	require.NoError(t, err)

	// Fill the cache
	for i := 0; i < 3; i++ {
		err = service.SetFeature(context.Background(), "entity_"+string(rune('0'+i)), "feature", map[string]interface{}{
			"value": float64(i),
		})
		require.NoError(t, err)
	}

	// Cache should be full
	status := service.GetCacheStatus()
	assert.Equal(t, 3, status.CurrentSize)

	// Add one more - should evict oldest (LRU)
	err = service.SetFeature(context.Background(), "entity_3", "feature", map[string]interface{}{
		"value": 3.0,
	})
	require.NoError(t, err)

	// Access entity_0 to make it recently used
	_, err = service.GetFeature(context.Background(), "entity_0", "feature")
	require.NoError(t, err)

	// Add another - should evict entity_1 (now oldest LRU)
	err = service.SetFeature(context.Background(), "entity_4", "feature", map[string]interface{}{
		"value": 4.0,
	})
	require.NoError(t, err)

	// entity_0 should still be accessible (was accessed recently)
	_, err = service.GetFeature(context.Background(), "entity_0", "feature")
	require.NoError(t, err)

	// entity_1 should have been evicted
	_, err = service.GetFeature(context.Background(), "entity_1", "feature")
	assert.Error(t, err)
}

// Test feature deletion
func TestServiceDeleteFeature(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{})
	require.NoError(t, err)

	// Set a feature
	err = service.SetFeature(context.Background(), "entity_delete", "feature", map[string]interface{}{
		"value": 100,
	})
	require.NoError(t, err)

	// Verify it exists
	_, err = service.GetFeature(context.Background(), "entity_delete", "feature")
	require.NoError(t, err)

	// Delete the feature
	err = service.DeleteFeature(context.Background(), "entity_delete", "feature")
	require.NoError(t, err)

	// Should not exist anymore
	_, err = service.GetFeature(context.Background(), "entity_delete", "feature")
	assert.Error(t, err)
}

// Test bulk operations
func TestServiceBulkSetFeatures(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
	})
	require.NoError(t, err)

	features := map[string]map[string]interface{}{
		"entity_1": {"feature_1": map[string]interface{}{"value": 100}},
		"entity_2": {"feature_1": map[string]interface{}{"value": 200}},
		"entity_3": {"feature_1": map[string]interface{}{"value": 300}},
	}

	err = service.BulkSetFeatures(context.Background(), features)
	require.NoError(t, err)

	// Verify all were set
	for i := 1; i <= 3; i++ {
		result, err := service.GetFeature(context.Background(), "entity_"+string(rune('0'+i)), "feature_1")
		require.NoError(t, err)
		assert.Equal(t, float64(i*100), result["value"])
	}
}

// Test JSON serialization
func TestFeatureRecordSerialization(t *testing.T) {
	record := online.FeatureRecord{
		EntityID:    "entity_001",
		FeatureName: "transaction_amount",
		Value:       15000.0,
		Timestamp:   time.Now(),
		Version:     "v1",
		Metadata: map[string]interface{}{
			"source": "bank",
		},
	}

	data, err := json.Marshal(record)
	require.NoError(t, err)

	var decoded online.FeatureRecord
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, record.EntityID, decoded.EntityID)
	assert.Equal(t, record.FeatureName, decoded.FeatureName)
	assert.Equal(t, record.Value, decoded.Value)
	assert.Equal(t, record.Version, decoded.Version)
}

// Test statistics
func TestServiceGetStatistics(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
	})
	require.NoError(t, err)

	// Perform some operations
	for i := 0; i < 10; i++ {
		service.SetFeature(context.Background(), "entity_"+string(rune('0'+i%3)), "feature", map[string]interface{}{
			"value": float64(i),
		})
		service.GetFeature(context.Background(), "entity_0", "feature")
	}

	stats := service.GetStatistics()
	assert.Equal(t, int64(10), stats.TotalSets)
	assert.GreaterOrEqual(t, stats.TotalGets, int64(1))
	assert.Equal(t, int64(0), stats.TotalErrors)
}

// Test service health check
func TestServiceHealthCheck(t *testing.T) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 100,
	})
	require.NoError(t, err)

	health := service.HealthCheck()
	assert.True(t, health.Healthy)
	assert.Equal(t, "operational", health.Status)
}

// Benchmark tests
func BenchmarkGetFeature(b *testing.B) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 10000,
	})
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		service.SetFeature(context.Background(), "entity_"+string(rune('0'+i%10)), "feature", map[string]interface{}{
			"value": float64(i),
		})
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetFeature(ctx, "entity_0", "feature")
	}
}

func BenchmarkSetFeature(b *testing.B) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 10000,
	})
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	data := map[string]interface{}{
		"value":     float64(b.N),
		"timestamp": time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SetFeature(ctx, "entity_bench", "feature_bench", data)
	}
}

func BenchmarkBulkSetFeatures(b *testing.B) {
	service, err := online.NewOnlineService(online.ServiceConfig{
		CacheType: "memory",
		CacheSize: 10000,
	})
	if err != nil {
		b.Fatal(err)
	}

	features := make(map[string]map[string]interface{})
	for i := 0; i < 100; i++ {
		features["entity_"+string(rune('0'+i))] = map[string]interface{}{
			"feature": map[string]interface{}{"value": float64(i)},
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.BulkSetFeatures(ctx, features)
	}
}
