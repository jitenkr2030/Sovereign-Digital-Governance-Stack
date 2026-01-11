package correlation

import (
	"math"
	"testing"
	"time"

	"neam-platform/intelligence/models"
)

// MockStore implements CorrelationStore for testing
type MockStore struct {
	timeSeriesData   map[models.MetricType][]models.TimeSeriesPoint
	correlationData  []*models.CorrelationData
	regionalStress   map[string]*models.RegionalStressIndex
}

func NewMockStore() *MockStore {
	return &MockStore{
		timeSeriesData:  make(map[models.MetricType][]models.TimeSeriesPoint),
		correlationData: make([]*models.CorrelationData, 0),
		regionalStress:  make(map[string]*models.RegionalStressIndex),
	}
}

func (m *MockStore) GetTimeSeriesData(metricType models.MetricType, regionID string, lookbackDays int) ([]models.TimeSeriesPoint, error) {
	return m.timeSeriesData[metricType], nil
}

func (m *MockStore) SaveCorrelationData(data *models.CorrelationData) error {
	m.correlationData = append(m.correlationData, data)
	return nil
}

func (m *MockStore) GetRegionalStressIndex(regionID string) (*models.RegionalStressIndex, error) {
	return m.regionalStress[regionID], nil
}

func (m *MockStore) SaveRegionalStressIndex(index *models.RegionalStressIndex) error {
	m.regionalStress[index.RegionID] = index
	return nil
}

func (m *MockStore) AddTimeSeriesData(metricType models.MetricType, data []models.TimeSeriesPoint) {
	m.timeSeriesData[metricType] = data
}

func (m *MockStore) SetRegionalStress(regionID string, stress *models.RegionalStressIndex) {
	m.regionalStress[regionID] = stress
}

func TestCalculatePearsonCorrelation(t *testing.T) {
	tests := []struct {
		name     string
		seriesA  []float64
		seriesB  []float64
		expected float64
	}{
		{
			name:     "Perfect positive correlation",
			seriesA:  []float64{1, 2, 3, 4, 5},
			seriesB:  []float64{2, 4, 6, 8, 10},
			expected: 1.0,
		},
		{
			name:     "Perfect negative correlation",
			seriesA:  []float64{1, 2, 3, 4, 5},
			seriesB:  []float64{5, 4, 3, 2, 1},
			expected: -1.0,
		},
		{
			name:     "No correlation",
			seriesA:  []float64{1, 2, 3, 4, 5},
			seriesB:  []float64{1, 3, 2, 5, 4},
			expected: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePearsonCorrelation(tt.seriesA, tt.seriesB)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Expected correlation ~%f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculatePearsonCorrelation_SameSeries(t *testing.T) {
	series := []float64{1, 2, 3, 4, 5}
	
	result := calculatePearsonCorrelation(series, series)
	
	if math.Abs(result-1.0) > 0.001 {
		t.Errorf("Expected correlation of 1.0 for identical series, got %f", result)
	}
}

func TestCalculatePearsonCorrelation_DifferentLengths(t *testing.T) {
	seriesA := []float64{1, 2, 3, 4, 5}
	seriesB := []float64{2, 4, 6}
	
	result := calculatePearsonCorrelation(seriesA, seriesB)
	
	// Should only use overlapping portion
	if math.Abs(result-1.0) > 0.001 {
		t.Errorf("Expected correlation of 1.0 for matching values, got %f", result)
	}
}

func TestCalculateSpearmanCorrelation(t *testing.T) {
	tests := []struct {
		name     string
		seriesA  []float64
		seriesB  []float64
		expected float64
	}{
		{
			name:     "Perfect positive rank correlation",
			seriesA:  []float64{1, 2, 3, 4, 5},
			seriesB:  []float64{2, 4, 6, 8, 10},
			expected: 1.0,
		},
		{
			name:     "Perfect negative rank correlation",
			seriesA:  []float64{1, 2, 3, 4, 5},
			seriesB:  []float64{5, 4, 3, 2, 1},
			expected: -1.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSpearmanCorrelation(tt.seriesA, tt.seriesB)
			if math.Abs(result-tt.expected) > 0.1 {
				t.Errorf("Expected Spearman correlation ~%f, got %f", tt.expected, result)
			}
		})
	}
}

func TestAlignSeries(t *testing.T) {
	seriesA := []models.TimeSeriesPoint{
		{Value: 10, Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: 20, Timestamp: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Value: 30, Timestamp: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
	}
	
	seriesB := []models.TimeSeriesPoint{
		{Value: 100, Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: 200, Timestamp: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Value: 300, Timestamp: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{Value: 400, Timestamp: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
	}
	
	alignedA, alignedB := alignSeries(seriesA, seriesB)
	
	if len(alignedA) != 3 {
		t.Errorf("Expected 3 aligned values, got %d", len(alignedA))
	}
	
	if len(alignedB) != 3 {
		t.Errorf("Expected 3 aligned values, got %d", len(alignedB))
	}
	
	// Verify values
	if alignedA[0] != 10 || alignedB[0] != 100 {
		t.Errorf("First aligned values incorrect")
	}
}

func TestEngine_AnalyzePair(t *testing.T) {
	store := NewMockStore()
	
	// Add time series data with known correlation
	energyData := []models.TimeSeriesPoint{
		{Value: 100, Timestamp: time.Now()},
		{Value: 105, Timestamp: time.Now().Add(time.Hour)},
		{Value: 110, Timestamp: time.Now().Add(2 * time.Hour)},
		{Value: 115, Timestamp: time.Now().Add(3 * time.Hour)},
		{Value: 120, Timestamp: time.Now().Add(4 * time.Hour)},
	}
	
	paymentData := []models.TimeSeriesPoint{
		{Value: 1000, Timestamp: time.Now()},
		{Value: 1050, Timestamp: time.Now().Add(time.Hour)},
		{Value: 1100, Timestamp: time.Now().Add(2 * time.Hour)},
		{Value: 1150, Timestamp: time.Now().Add(3 * time.Hour)},
		{Value: 1200, Timestamp: time.Now().Add(4 * time.Hour)},
	}
	
	store.AddTimeSeriesData(models.MetricEnergyConsumption, energyData)
	store.AddTimeSeriesData(models.MetricPaymentVolume, paymentData)
	
	engine := NewEngine(store, DefaultConfig())
	
	corr := engine.analyzePair(models.MetricEnergyConsumption, models.MetricPaymentVolume)
	
	if corr == nil {
		t.Fatal("Expected correlation result, got nil")
	}
	
	// Should have strong positive correlation
	if corr.PearsonCoeff < 0.9 {
		t.Errorf("Expected strong positive correlation, got %f", corr.PearsonCoeff)
	}
}

func TestEngine_AnalyzePair_NoData(t *testing.T) {
	store := NewMockStore()
	
	engine := NewEngine(store, DefaultConfig())
	
	corr := engine.analyzePair(models.MetricEnergyConsumption, models.MetricPaymentVolume)
	
	if corr != nil {
		t.Error("Expected nil correlation for missing data")
	}
}

func TestEngine_GetHistoricalCorrelation(t *testing.T) {
	store := NewMockStore()
	engine := NewEngine(store, DefaultConfig())
	
	tests := []struct {
		name     string
		metricA  models.MetricType
		metricB  models.MetricType
		expected float64
	}{
		{
			name:     "Energy-Payment correlation",
			metricA:  models.MetricEnergyConsumption,
			metricB:  models.MetricPaymentVolume,
			expected: 0.75,
		},
		{
			name:     "Cash-Digital correlation",
			metricA:  models.MetricCashTransaction,
			metricB:  models.MetricDigitalTransaction,
			expected: -0.3,
		},
		{
			name:     "Retail-Employment correlation",
			metricA:  models.MetricRetailSales,
			metricB:  models.MetricEmployment,
			expected: 0.65,
		},
		{
			name:     "Unknown pair",
			metricA:  models.MetricCPI,
			metricB:  models.MetricProducerPrice,
			expected: 0.5, // Default
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.getHistoricalCorrelation(tt.metricA, tt.metricB)
			if result != tt.expected {
				t.Errorf("Expected historical correlation %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestToRanks(t *testing.T) {
	values := []float64{3, 1, 4, 1, 5, 9, 2, 6}
	
	ranks := toRanks(values)
	
	// 1s should have same rank (average of positions 1 and 2)
	rankOfFirst := ranks[0]
	rankOfSecond := ranks[1]
	rankOfThird := ranks[2]
	
	// Third element (value 4) should be rank 3 (positions 3, value is 3rd smallest)
	if ranks[2] != 3 {
		t.Errorf("Expected rank 3 for value 4, got %f", ranks[2])
	}
}

func TestHeatMapper_GenerateHeatMap(t *testing.T) {
	store := NewMockStore()
	store.SetRegionalStress("REGION_A", &models.RegionalStressIndex{
		RegionID:     "REGION_A",
		OverallScore: 75.0,
	})
	store.SetRegionalStress("REGION_B", &models.RegionalStressIndex{
		RegionID:     "REGION_B",
		OverallScore: 25.0,
	})
	
	heatMapper := NewHeatMapper(store)
	
	indices, err := heatMapper.GenerateHeatMap([]string{"REGION_A", "REGION_B"})
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(indices) != 2 {
		t.Errorf("Expected 2 regional indices, got %d", len(indices))
	}
}

func TestHeatMapper_UpdateStressFromSignals(t *testing.T) {
	store := NewMockStore()
	heatMapper := NewHeatMapper(store)
	
	stress := &models.RegionalStressIndex{
		RegionID:       "TEST",
		AnomalyScore:   0.6,
		InflationScore: 50.0,
		CorrelationBreaks: 2,
	}
	
	heatMapper.updateStressFromSignals(stress)
	
	// Overall score should be weighted combination
	expected := 0.6*100*0.4 + 50.0*0.4 + 2.0*20*0.2
	if math.Abs(stress.OverallScore-expected) > 0.01 {
		t.Errorf("Expected overall score ~%f, got %f", expected, stress.OverallScore)
	}
}

func TestRegionalAnalyzer_AnalyzeRegion(t *testing.T) {
	store := NewMockStore()
	
	// Add some test data
	paymentData := []models.TimeSeriesPoint{
		{Value: 1000, Timestamp: time.Now().AddDate(0, 0, -30)},
		{Value: 1050, Timestamp: time.Now().AddDate(0, 0, -29)},
		{Value: 1025, Timestamp: time.Now().AddDate(0, 0, -28)},
		// ... more data
	}
	store.AddTimeSeriesData(models.MetricPaymentVolume, paymentData)
	
	analyzer := NewRegionalAnalyzer(store)
	
	analysis, err := analyzer.AnalyzeRegion("TEST_REGION", 30)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if analysis.RegionID != "TEST_REGION" {
		t.Errorf("Expected region ID TEST_REGION, got %s", analysis.RegionID)
	}
}

func TestRegionalAnalyzer_DetermineTrendDirection(t *testing.T) {
	store := NewMockStore()
	analyzer := NewRegionalAnalyzer(store)
	
	tests := []struct {
		name     string
		values   []float64
		expected string
	}{
		{
			name:     "Upward trend",
			values:   []float64{100, 102, 104, 106, 108, 110},
			expected: "UP",
		},
		{
			name:     "Downward trend",
			values:   []float64{110, 108, 106, 104, 102, 100},
			expected: "DOWN",
		},
		{
			name:     "Stable",
			values:   []float64{100, 101, 100, 101, 100, 101},
			expected: "STABLE",
		},
		{
			name:     "Insufficient data",
			values:   []float64{100, 101},
			expected: "INSUFFICIENT_DATA",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.determineTrendDirection(tt.values)
			if result != tt.expected {
				t.Errorf("Expected trend %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRegionalAnalyzer_CalculateRegionalHealth(t *testing.T) {
	store := NewMockStore()
	analyzer := NewRegionalAnalyzer(store)
	
	metrics := map[models.MetricType]RegionalMetricAnalysis{
		models.MetricPaymentVolume: {
			RecentChange:   0.05,  // 5% increase
			Volatility:     0.02,  // Low volatility
		},
		models.MetricRetailSales: {
			RecentChange:   0.03,  // 3% increase
			Volatility:     0.03,  // Low volatility
		},
	}
	
	health := analyzer.calculateRegionalHealth(metrics)
	
	// Should be a weighted combination
	if health < 50 || health > 100 {
		t.Errorf("Expected health between 50 and 100, got %f", health)
	}
}

func TestRegionalAnalyzer_DetermineTrend(t *testing.T) {
	store := NewMockStore()
	analyzer := NewRegionalAnalyzer(store)
	
	tests := []struct {
		name     string
		metrics  map[models.MetricType]RegionalMetricAnalysis
		expected string
	}{
		{
			name: "Expanding",
			metrics: map[models.MetricType]RegionalMetricAnalysis{
				models.MetricPaymentVolume:  {TrendDirection: "UP"},
				models.MetricRetailSales:   {TrendDirection: "UP"},
				models.MetricEmployment:    {TrendDirection: "STABLE"},
			},
			expected: "EXPANDING",
		},
		{
			name: "Contracting",
			metrics: map[models.MetricType]RegionalMetricAnalysis{
				models.MetricPaymentVolume:  {TrendDirection: "DOWN"},
				models.MetricRetailSales:   {TrendDirection: "DOWN"},
				models.MetricEmployment:    {TrendDirection: "STABLE"},
			},
			expected: "CONTRACTING",
		},
		{
			name: "Stable",
			metrics: map[models.MetricType]RegionalMetricAnalysis{
				models.MetricPaymentVolume:  {TrendDirection: "STABLE"},
				models.MetricRetailSales:   {TrendDirection: "STABLE"},
			},
			expected: "STABLE",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.determineTrend(tt.metrics)
			if result != tt.expected {
				t.Errorf("Expected trend %s, got %s", tt.expected, result)
			}
		})
	}
}

// Benchmark tests

func BenchmarkCalculatePearsonCorrelation(b *testing.B) {
	seriesA := make([]float64, 1000)
	seriesB := make([]float64, 1000)
	for i := range seriesA {
		seriesA[i] = float64(i)
		seriesB[i] = float64(i) * 1.5
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculatePearsonCorrelation(seriesA, seriesB)
	}
}

func BenchmarkToRanks(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i % 100)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toRanks(values)
	}
}
