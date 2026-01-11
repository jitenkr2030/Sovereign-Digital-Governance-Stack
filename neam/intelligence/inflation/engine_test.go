package inflation

import (
	"math"
	"testing"
	"time"

	"neam-platform/intelligence/models"
)

// MockStore implements InflationStore for testing
type MockStore struct {
	cpiData       []models.InflationData
	commodityData []models.TimeSeriesPoint
	producerData  []models.TimeSeriesPoint
}

func NewMockStore() *MockStore {
	return &MockStore{
		cpiData:       make([]models.InflationData, 0),
		commodityData: make([]models.TimeSeriesPoint, 0),
		producerData:  make([]models.TimeSeriesPoint, 0),
	}
}

func (m *MockStore) GetCPIHistory(regionID string, lookbackDays int) ([]models.InflationData, error) {
	return m.cpiData, nil
}

func (m *MockStore) GetCommodityPrices(lookbackDays int) ([]models.TimeSeriesPoint, error) {
	return m.commodityData, nil
}

func (m *MockStore) GetProducerPrices(regionID string, lookbackDays int) ([]models.TimeSeriesPoint, error) {
	return m.producerData, nil
}

func (m *MockStore) SavePrediction(prediction *models.InflationPrediction) error {
	return nil
}

func (m *MockStore) AddCPIHistory(data []models.InflationData) {
	m.cpiData = data
}

func (m *MockStore) AddCommodityData(data []models.TimeSeriesPoint) {
	m.commodityData = data
}

func (m *MockStore) AddProducerData(data []models.TimeSeriesPoint) {
	m.producerData = data
}

func TestEngine_CalculateCurrentInflation(t *testing.T) {
	store := NewMockStore()
	
	// Create 365 days of CPI data with steady 5% inflation
	data := make([]models.InflationData, 365)
	baseCPI := 100.0
	for i := 0; i < 365; i++ {
		cpi := baseCPI * math.Pow(1.05, float64(i)/365)
		data[i] = models.InflationData{
			Timestamp: time.Now().AddDate(0, 0, -365+i),
			CPI:       cpi,
		}
	}
	store.AddCPIHistory(data)
	
	engine := NewEngine(store, DefaultConfig())
	
	inflation := engine.calculateCurrentInflation(data)
	
	// Should be approximately 5% (0.05)
	if math.Abs(inflation-0.05) > 0.01 {
		t.Errorf("Expected inflation ~0.05, got %f", inflation)
	}
}

func TestEngine_GeneratePredictions(t *testing.T) {
	store := NewMockStore()
	
	// Create CPI history with upward trend
	data := make([]models.InflationData, 90)
	baseCPI := 100.0
	for i := 0; i < 90; i++ {
		cpi := baseCPI + float64(i)*0.02 // Gradual increase
		data[i] = models.InflationData{
			Timestamp: time.Now().AddDate(0, 0, -90+i),
			CPI:       cpi,
		}
	}
	store.AddCPIHistory(data)
	
	// Add some leading indicator data
	commodityData := make([]models.TimeSeriesPoint, 30)
	for i := 0; i < 30; i++ {
		commodityData[i] = models.TimeSeriesPoint{
			Timestamp: time.Now().AddDate(0, 0, -30+i),
			Value:     100 + float64(i)*0.1,
		}
	}
	store.AddCommodityData(commodityData)
	
	engine := NewEngine(store, DefaultConfig())
	
	predictions := engine.generatePredictions(data, commodityData, nil)
	
	// Check that predictions exist for all horizons
	horizons := []int{30, 90, 180}
	for _, h := range horizons {
		if _, ok := predictions[h]; !ok {
			t.Errorf("Missing prediction for horizon %d", h)
		}
	}
}

func TestEngine_AssessRiskLevel(t *testing.T) {
	store := NewMockStore()
	engine := NewEngine(store, DefaultConfig())
	
	tests := []struct {
		name           string
		predictions    map[int]float64
		currentInflation float64
		expectedRisk   string
	}{
		{
			name: "Critical inflation",
			predictions: map[int]float64{
				30:  0.12,
				90:  0.15,
				180: 0.18,
			},
			currentInflation: 0.12,
			expectedRisk:     "CRITICAL",
		},
		{
			name: "High inflation",
			predictions: map[int]float64{
				30:  0.08,
				90:  0.09,
				180: 0.10,
			},
			currentInflation: 0.08,
			expectedRisk:     "HIGH",
		},
		{
			name: "Medium inflation",
			predictions: map[int]float64{
				30:  0.05,
				90:  0.06,
				180: 0.07,
			},
			currentInflation: 0.05,
			expectedRisk:     "MEDIUM",
		},
		{
			name: "Low inflation",
			predictions: map[int]float64{
				30:  0.02,
				90:  0.03,
				180: 0.02,
			},
			currentInflation: 0.02,
			expectedRisk:     "LOW",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			riskLevel := engine.assessRiskLevel(tt.predictions, tt.currentInflation)
			if riskLevel != tt.expectedRisk {
				t.Errorf("Expected risk level %s, got %s", tt.expectedRisk, riskLevel)
			}
		})
	}
}

func TestEngine_IdentifyContributingFactors(t *testing.T) {
	store := NewMockStore()
	engine := NewEngine(store, DefaultConfig())
	
	// Create data with rising commodity prices
	data := []models.InflationData{
		{
			Timestamp:       time.Now(),
			CPI:             105,
			CoreCPI:         106,
			PPI:             110,
			CommodityIndex:  120,
		},
	}
	
	commodityData := []models.TimeSeriesPoint{
		{Value: 100},
		{Value: 105},
		{Value: 110},
		{Value: 115},
		{Value: 120},
	}
	
	factors := engine.identifyContributingFactors(data, commodityData, nil)
	
	// Should identify rising commodity prices
	found := false
	for _, factor := range factors {
		if factor == "Rising commodity prices" {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Expected 'Rising commodity prices' in factors, got: %v", factors)
	}
}

func TestSimplePredictionModel(t *testing.T) {
	model := NewSimplePredictionModel()
	
	// Create linear trend data
	series := make([]float64, 100)
	for i := range series {
		series[i] = 100 + float64(i)*0.5
	}
	
	prediction, confidence := model.Predict(series, 30)
	
	// Should predict future growth
	if prediction <= 0 {
		t.Errorf("Expected positive prediction for upward trend, got %f", prediction)
	}
	
	// Confidence should be reasonable for stable trend
	if confidence < 0.5 {
		t.Errorf("Expected confidence > 0.5 for stable trend, got %f", confidence)
	}
}

func TestSimplePredictionModel_Volatile(t *testing.T) {
	model := NewSimplePredictionModel()
	
	// Create volatile data
	series := make([]float64, 100)
	for i := range series {
		series[i] = 100 + float64(i%10)*10
	}
	
	prediction, confidence := model.Predict(series, 30)
	
	// Confidence should be lower for volatile data
	if confidence > 0.6 {
		t.Errorf("Expected lower confidence for volatile data, got %f", confidence)
	}
}

func TestCalculateTrend(t *testing.T) {
	engine := NewEngine(NewMockStore(), DefaultConfig())
	
	tests := []struct {
		name     string
		series   []float64
		expected float64
	}{
		{
			name:     "Upward trend",
			series:   []float64{100, 101, 102, 103, 104},
			expected: 0.01, // ~1% per step
		},
		{
			name:     "Downward trend",
			series:   []float64{104, 103, 102, 101, 100},
			expected: -0.01, // ~-1% per step
		},
		{
			name:     "Flat trend",
			series:   []float64{100, 100, 100, 100, 100},
			expected: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := engine.calculateTrend(tt.series)
			if math.Abs(trend-tt.expected) > 0.005 {
				t.Errorf("Expected trend ~%f, got %f", tt.expected, trend)
			}
		})
	}
}

func TestCalculatePredictionConfidence(t *testing.T) {
	store := NewMockStore()
	engine := NewEngine(store, DefaultConfig())
	
	tests := []struct {
		name     string
		data     []models.InflationData
		expected float64
	}{
		{
			name: "Low volatility",
			data: generateStableCPI(90),
			expected: 0.7, // Should be high
		},
		{
			name:     "Insufficient data",
			data:     []models.InflationData{},
			expected: 0.5,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := engine.calculatePredictionConfidence(tt.data)
			if math.Abs(confidence-tt.expected) > 0.2 {
				t.Errorf("Expected confidence ~%f, got %f", tt.expected, confidence)
			}
		})
	}
}

func TestCalculateLeadingIndicatorInfluence(t *testing.T) {
	engine := NewEngine(NewMockStore(), DefaultConfig())
	
	tests := []struct {
		name     string
		data     []models.TimeSeriesPoint
		expected float64
	}{
		{
			name: "Rising prices",
			data: []models.TimeSeriesPoint{
				{Value: 100},
				{Value: 105},
			},
			expected: 0.05, // 5% increase
		},
		{
			name:     "Falling prices",
			data:     []models.TimeSeriesPoint{{Value: 100}, {Value: 95}},
			expected: -0.05,
		},
		{
			name:     "No data",
			data:     []models.TimeSeriesPoint{},
			expected: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			influence := engine.calculateLeadingIndicatorInfluence(tt.data)
			if math.Abs(influence-tt.expected) > 0.001 {
				t.Errorf("Expected influence ~%f, got %f", tt.expected, influence)
			}
		})
	}
}

// Helper functions

func generateStableCPI(days int) []models.InflationData {
	data := make([]models.InflationData, days)
	for i := range data {
		data[i] = models.InflationData{
			Timestamp: time.Now().AddDate(0, 0, -days+i),
			CPI:       100 + float64(i)*0.001, // Very stable
		}
	}
	return data
}

// Benchmark tests

func BenchmarkEngine_CalculateCurrentInflation(b *testing.B) {
	store := NewMockStore()
	store.AddCPIHistory(generateStableCPI(365))
	engine := NewEngine(store, DefaultConfig())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.calculateCurrentInflation(store.cpiData)
	}
}

func BenchmarkEngine_GeneratePredictions(b *testing.B) {
	store := NewMockStore()
	store.AddCPIHistory(generateStableCPI(90))
	store.AddCommodityData([]models.TimeSeriesPoint{{Value: 100}})
	engine := NewEngine(store, DefaultConfig())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.generatePredictions(store.cpiData, store.commodityData, nil)
	}
}

func BenchmarkSimplePredictionModel(b *testing.B) {
	model := NewSimplePredictionModel()
	series := make([]float64, 365)
	for i := range series {
		series[i] = 100 + float64(i)*0.02
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Predict(series, 30)
	}
}
