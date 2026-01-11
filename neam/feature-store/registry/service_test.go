package registry_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/feature-store/registry"
)

// MockValidator is a mock implementation of the validator interface
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(ctx context.Context, feature registry.FeatureDefinition) error {
	args := m.Called(ctx, feature)
	return args.Error(0)
}

func (m *MockValidator) GetValidatorType() string {
	args := m.Called()
	return args.String(0)
}

// TestRegistryService creation and initialization
func TestNewRegistryService(t *testing.T) {
	t.Run("creates service with valid config", func(t *testing.T) {
		config := registry.ServiceConfig{
			StorageType:     "postgresql",
			EnableVersioning: true,
			MaxVersions:      10,
			ValidationEnabled: true,
		}

		service, err := registry.NewRegistryService(config)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, "postgresql", service.GetStorageType())
		assert.True(t, service.IsVersioningEnabled())
		assert.Equal(t, 10, service.GetMaxVersions())
	})

	t.Run("creates service with default config", func(t *testing.T) {
		service, err := registry.NewRegistryService(registry.ServiceConfig{})
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("fails with invalid storage type", func(t *testing.T) {
		config := registry.ServiceConfig{
			StorageType: "invalid",
		}
		service, err := registry.NewRegistryService(config)
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

// Test feature registration
func TestServiceRegisterFeature(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{
		EnableVersioning: true,
	})
	require.NoError(t, err)

	feature := registry.FeatureDefinition{
		Name:        "transaction_amount",
		Description: "Total transaction amount",
		DataType:    "float64",
		Source:      "bank_transactions",
		Owner:       "analytics_team",
		Tags:        []string{"financial", "transaction"},
	}

	err = service.RegisterFeature(context.Background(), feature)
	require.NoError(t, err)

	// Verify feature was registered
	retrieved, err := service.GetFeature(context.Background(), "transaction_amount")
	require.NoError(t, err)
	assert.Equal(t, "transaction_amount", retrieved.Name)
	assert.Equal(t, "float64", retrieved.DataType)
}

func TestServiceRegisterDuplicateFeature(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	feature := registry.FeatureDefinition{
		Name:  "duplicate_feature",
		Owner: "team_a",
	}

	err = service.RegisterFeature(context.Background(), feature)
	require.NoError(t, err)

	// Try to register same feature with different owner
	feature.Owner = "team_b"
	err = service.RegisterFeature(context.Background(), feature)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// Test feature versioning
func TestServiceCreateFeatureVersion(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{
		EnableVersioning: true,
		MaxVersions:      3,
	})
	require.NoError(t, err)

	// Register initial version
	feature := registry.FeatureDefinition{
		Name:        "versioned_feature",
		Description: "A versioned feature",
		DataType:    "int64",
		Version:     "v1",
	}

	err = service.RegisterFeature(context.Background(), feature)
	require.NoError(t, err)

	// Create new version
	newVersion := registry.FeatureDefinition{
		Name:        "versioned_feature",
		Description: "A versioned feature - updated",
		DataType:    "int64",
		Version:     "v2",
	}

	err = service.CreateVersion(context.Background(), "versioned_feature", newVersion)
	require.NoError(t, err)

	// Get latest version
	latest, err := service.GetFeature(context.Background(), "versioned_feature")
	require.NoError(t, err)
	assert.Equal(t, "v2", latest.Version)

	// Get specific version
	v1, err := service.GetFeatureVersion(context.Background(), "versioned_feature", "v1")
	require.NoError(t, err)
	assert.Equal(t, "v1", v1.Version)
}

func TestServiceVersionLimit(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{
		EnableVersioning: true,
		MaxVersions:      2,
	})
	require.NoError(t, err)

	// Create initial version
	feature := registry.FeatureDefinition{
		Name:    "limited_feature",
		Version: "v1",
	}
	service.RegisterFeature(context.Background(), feature)

	// Create v2
	feature.Version = "v2"
	service.CreateVersion(context.Background(), "limited_feature", feature)

	// Try to create v3 - should fail due to version limit
	feature.Version = "v3"
	err = service.CreateVersion(context.Background(), "limited_feature", feature)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version limit")
}

// Test feature search
func TestServiceSearchFeatures(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	// Register some features
	features := []registry.FeatureDefinition{
		{Name: "feature_1", Owner: "team_a", Tags: []string{"financial"}},
		{Name: "feature_2", Owner: "team_a", Tags: []string{"financial", "transaction"}},
		{Name: "feature_3", Owner: "team_b", Tags: []string{"operational"}},
		{Name: "feature_4", Owner: "team_b", Tags: []string{"financial", "operational"}},
	}

	for _, f := range features {
		service.RegisterFeature(context.Background(), f)
	}

	// Search by owner
	results, err := service.SearchFeatures(context.Background(), registry.FeatureQuery{
		Owner: "team_a",
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Search by tag
	results, err = service.SearchFeatures(context.Background(), registry.FeatureQuery{
		Tags: []string{"financial"},
	})
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Search by owner and tag
	results, err = service.SearchFeatures(context.Background(), registry.FeatureQuery{
		Owner: "team_b",
		Tags:  []string{"financial"},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)

	// Search with pagination
	results, err = service.SearchFeatures(context.Background(), registry.FeatureQuery{
		Limit:  2,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

// Test feature dependencies
func TestServiceAddFeatureDependency(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	// Register features
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "feature_a"})
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "feature_b"})

	// Add dependency
	err = service.AddDependency(context.Background(), "feature_a", "feature_b")
	require.NoError(t, err)

	// Get dependencies
	deps, err := service.GetDependencies(context.Background(), "feature_a")
	require.NoError(t, err)
	assert.Contains(t, deps, "feature_b")
}

func TestServiceDetectCircularDependency(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	// Register features
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "feature_x"})
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "feature_y"})
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "feature_z"})

	// Add dependencies: x -> y -> z -> x (circular)
	service.AddDependency(context.Background(), "feature_x", "feature_y")
	service.AddDependency(context.Background(), "feature_y", "feature_z")

	// Try to add circular dependency
	err = service.AddDependency(context.Background(), "feature_z", "feature_x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

// Test feature lineage
func TestServiceGetFeatureLineage(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	// Create a dependency chain: a -> b -> c
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "lineage_a"})
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "lineage_b"})
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{Name: "lineage_c"})

	service.AddDependency(context.Background(), "lineage_a", "lineage_b")
	service.AddDependency(context.Background(), "lineage_b", "lineage_c")

	// Get lineage for a - should include b and c
	lineage, err := service.GetLineage(context.Background(), "lineage_a")
	require.NoError(t, err)
	assert.Len(t, lineage.Dependencies, 2)
	assert.Contains(t, lineage.Dependencies, "lineage_b")
	assert.Contains(t, lineage.Dependencies, "lineage_c")
}

// Test feature validation
func TestServiceValidateFeature(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{
		ValidationEnabled: true,
	})
	require.NoError(t, err)

	mockValidator := new(MockValidator)
	mockValidator.On("Validate", mock.Anything, mock.Anything).Return(nil)
	mockValidator.On("GetValidatorType").Return("schema")

	err = service.RegisterValidator(mockValidator)
	require.NoError(t, err)

	// Valid feature
	feature := registry.FeatureDefinition{
		Name:        "valid_feature",
		Description: "A valid feature",
		DataType:    "string",
	}

	err = service.ValidateFeature(context.Background(), feature)
	require.NoError(t, err)

	// Invalid feature (validator returns error)
	mockValidator2 := new(MockValidator)
	mockValidator2.On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)
	mockValidator2.On("GetValidatorType").Return("failing")
	service.RegisterValidator(mockValidator2)

	invalidFeature := registry.FeatureDefinition{
		Name: "invalid_feature",
	}

	err = service.ValidateFeature(context.Background(), invalidFeature)
	assert.Error(t, err)
}

// Test feature deprecation
func TestServiceDeprecateFeature(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	// Register a feature
	service.RegisterFeature(context.Background(), registry.FeatureDefinition{
		Name:     "deprecated_feature",
		Owner:    "team",
		Status:   registry.FeatureStatusActive,
	})

	// Deprecate the feature
	err = service.DeprecateFeature(context.Background(), "deprecated_feature", "Use new_feature instead")
	require.NoError(t, err)

	// Verify status changed
	feature, err := service.GetFeature(context.Background(), "deprecated_feature")
	require.NoError(t, err)
	assert.Equal(t, registry.FeatureStatusDeprecated, feature.Status)
	assert.Contains(t, feature.DeprecationNotice, "Use new_feature instead")
}

// Test JSON serialization
func TestFeatureDefinitionSerialization(t *testing.T) {
	feature := registry.FeatureDefinition{
		ID:          "feature_001",
		Name:        "transaction_amount",
		Description: "Total transaction amount",
		DataType:    "float64",
		Source:      "bank_transactions",
		Owner:       "analytics_team",
		Version:     "v1.0.0",
		Status:      registry.FeatureStatusActive,
		Tags:        []string{"financial", "transaction"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"retention": "1y",
		},
	}

	data, err := json.Marshal(feature)
	require.NoError(t, err)

	var decoded registry.FeatureDefinition
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, feature.ID, decoded.ID)
	assert.Equal(t, feature.Name, decoded.Name)
	assert.Equal(t, feature.DataType, decoded.DataType)
	assert.Equal(t, feature.Status, decoded.Status)
	assert.Len(t, decoded.Tags, 2)
}

// Test statistics
func TestServiceGetStatistics(t *testing.T) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	require.NoError(t, err)

	// Register some features
	for i := 0; i < 5; i++ {
		service.RegisterFeature(context.Background(), registry.FeatureDefinition{
			Name:  "stats_feature_" + string(rune('0'+i)),
			Owner: "team_" + string(rune('0'+i%2)),
		})
	}

	stats := service.GetStatistics()
	assert.Equal(t, int64(5), stats.TotalFeatures)
	assert.Equal(t, int64(0), stats.DeprecatedFeatures)
	assert.Len(stats.FeaturesByOwner, 2)
}

// Benchmark tests
func BenchmarkRegisterFeature(b *testing.B) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	if err != nil {
		b.Fatal(err)
	}

	feature := registry.FeatureDefinition{
		Name:        "bench_feature",
		Description: "A benchmark feature",
		DataType:    "float64",
		Owner:       "bench_team",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		feature.Name = "bench_feature_" + string(rune('0'+i))
		service.RegisterFeature(ctx, feature)
	}
}

func BenchmarkSearchFeatures(b *testing.B) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate with features
	for i := 0; i < 1000; i++ {
		service.RegisterFeature(context.Background(), registry.FeatureDefinition{
			Name:  "search_feature_" + string(rune('0'+i%100)),
			Owner: "team_" + string(rune('0'+i%10)),
			Tags:  []string{"tag_" + string(rune('0'+i%5))},
		})
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SearchFeatures(ctx, registry.FeatureQuery{
			Owner: "team_0",
			Tags:  []string{"tag_0"},
		})
	}
}

func BenchmarkGetLineage(b *testing.B) {
	service, err := registry.NewRegistryService(registry.ServiceConfig{})
	if err != nil {
		b.Fatal(err)
	}

	// Create a deep dependency chain
	for i := 0; i < 50; i++ {
		service.RegisterFeature(context.Background(), registry.FeatureDefinition{
			Name: "lineage_feature_" + string(rune('0'+i)),
		})
		if i > 0 {
			service.AddDependency(context.Background(), 
				"lineage_feature_"+string(rune('0'+i-1)),
				"lineage_feature_"+string(rune('0'+i)))
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetLineage(ctx, "lineage_feature_0")
	}
}
