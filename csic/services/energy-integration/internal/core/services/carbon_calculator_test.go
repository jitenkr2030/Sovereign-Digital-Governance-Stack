package services

import (
	"context"
	"testing"

	"github.com/energy-integration/energy/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEnergyProfileRepository for CarbonCalculator tests.
type MockEnergyProfileRepository struct {
	mock.Mock
}

func (m *MockEnergyProfileRepository) Create(ctx context.Context, profile *domain.EnergyProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEnergyProfileRepository) GetByID(ctx context.Context, id string) (*domain.EnergyProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EnergyProfile), args.Error(1)
}

func (m *MockEnergyProfileRepository) GetByChainID(ctx context.Context, chainID string) (*domain.EnergyProfile, error) {
	args := m.Called(ctx, chainID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EnergyProfile), args.Error(1)
}

func (m *MockEnergyProfileRepository) Update(ctx context.Context, profile *domain.EnergyProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockEnergyProfileRepository) List(ctx context.Context) ([]*domain.EnergyProfile, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.EnergyProfile), args.Error(1)
}

func (m *MockEnergyProfileRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockExternalOracleRepository for CarbonCalculator tests.
type MockExternalOracleRepository struct {
	mock.Mock
}

func (m *MockExternalOracleRepository) GetCarbonIntensity(ctx context.Context, regionCode string) (float64, error) {
	args := m.Called(ctx, regionCode)
	return float64(args.Float(0)), args.Error(1)
}

func (m *MockExternalOracleRepository) VerifyCertificate(ctx context.Context, certificateNumber string) (*domain.OffsetCertificate, error) {
	args := m.Called(ctx, certificateNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OffsetCertificate), args.Error(1)
}

func (m *MockExternalOracleRepository) GetEnergyMix(ctx context.Context, regionCode string) (map[domain.EnergySource]float64, error) {
	args := m.Called(ctx, regionCode)
	return args.Get(0).(map[domain.EnergySource]float64), args.Error(1)
}

func TestCarbonCalculatorService_CalculateFootprint_WithProfile(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	profile := &domain.EnergyProfile{
		ID:              domain.NewEntityID(),
		ChainID:         "ethereum-pos",
		ChainName:       "Ethereum PoS",
		ConsensusType:   domain.ConsensusProofOfStake,
		AvgKWhPerTx:     0.0006,
		BaseCarbonGrams: 0.3,
		EnergySource:    domain.EnergySourceMixed,
		CarbonIntensity: 0.5,
		IsActive:        true,
	}

	profileRepo.On("GetByChainID", ctx, "ethereum-pos").Return(profile, nil)

	req := domain.EnergyCalculationRequest{
		ChainID: "ethereum-pos",
		TransactionData: map[string]interface{}{
			"gas_used": 50000.0,
		},
	}

	result, err := calculator.CalculateFootprint(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.EnergyUnitKilowattHours, result.EnergyUsed.Unit)
	assert.Equal(t, domain.CarbonUnitGrams, result.CarbonEmission.Unit)
	assert.NotNil(t, result.ComparisonData)
	assert.NotEmpty(t, result.Recommendations)
	profileRepo.AssertExpectations(t)
}

func TestCarbonCalculatorService_CalculateFootprint_UnknownChain(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	profileRepo.On("GetByChainID", ctx, "unknown-chain").Return(nil, nil)

	req := domain.EnergyCalculationRequest{
		ChainID: "unknown-chain",
		TransactionData: map[string]interface{}{},
	}

	result, err := calculator.CalculateFootprint(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Default profile should be used
	assert.Equal(t, domain.ConsensusProofOfStake, result.EnergyUsed.Source)
	profileRepo.AssertExpectations(t)
}

func TestCarbonCalculatorService_CalculateFootprint_WithOracle(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	profile := &domain.EnergyProfile{
		ID:              domain.NewEntityID(),
		ChainID:         "ethereum",
		ChainName:       "Ethereum",
		ConsensusType:   domain.ConsensusProofOfWork,
		AvgKWhPerTx:     0.038,
		BaseCarbonGrams: 19.0,
		EnergySource:    domain.EnergySourceMixed,
		CarbonIntensity: 0.5, // Will be overridden by oracle
		IsActive:        true,
	}

	profileRepo.On("GetByChainID", ctx, "ethereum").Return(profile, nil)
	oracleRepo.On("GetCarbonIntensity", ctx, "EU").Return(0.3, nil)

	req := domain.EnergyCalculationRequest{
		ChainID:      "ethereum",
		NodeLocation: "EU",
		TransactionData: map[string]interface{}{
			"gas_used": 100000.0,
		},
	}

	result, err := calculator.CalculateFootprint(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Oracle intensity (0.3) should be used instead of profile intensity (0.5)
	oracleRepo.AssertExpectations(t)
}

func TestCarbonCalculatorService_CalculateFootprint_PoWRecommendation(t *testing.T) {
	ctx := context.Background()
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	// Create a PoW profile (high energy consumption)
	powProfile := &domain.EnergyProfile{
		ID:              domain.NewEntityID(),
		ChainID:         "bitcoin",
		ChainName:       "Bitcoin",
		ConsensusType:   domain.ConsensusProofOfWork,
		AvgKWhPerTx:     850.0,
		BaseCarbonGrams: 425000.0,
		EnergySource:    domain.EnergySourceFossilFuel,
		CarbonIntensity: 0.9,
		IsActive:        true,
	}

	profileRepo.On("GetByChainID", ctx, "bitcoin").Return(powProfile, nil)

	req := domain.EnergyCalculationRequest{
		ChainID: "bitcoin",
		TransactionData: map[string]interface{}{},
	}

	result, err := calculator.CalculateFootprint(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should have recommendation for PoW chains
	assert.Contains(t, result.Recommendations, "Consider using proof-of-stake alternatives")
	profileRepo.AssertExpectations(t)
}

func TestCarbonCalculatorService_GenerateComparisonData(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	// Test with a known emission value
	emission := domain.CarbonEmission{
		Value: 12000.0, // 12kg of CO2
		Unit:  domain.CarbonUnitGrams,
	}

	comparison := calculator.generateComparisonData(emission)

	assert.NotNil(t, comparison)
	assert.Greater(t, comparison.EquivalentKmByCar, 0.0)
	assert.Greater(t, comparison.EquivalentTreeDays, 0.0)
	assert.Greater(t, comparison.EquivalentSmartphoneCharges, int64(0))
	assert.Greater(t, comparison.EquivalentLEDBulbHours, 0.0)
}

func TestCarbonCalculatorService_GenerateRecommendations_PositiveOffset(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	// Create a renewable energy profile
	profile := &domain.EnergyProfile{
		ConsensusType: domain.ConsensusProofOfStake,
		EnergySource:  domain.EnergySourceRenewable,
	}

	// Low emission - should not trigger offset recommendation
	emission := domain.CarbonEmission{
		Value: 100.0, // 100g
		Unit:  domain.CarbonUnitGrams,
	}

	recommendations := calculator.generateRecommendations(profile, emission)

	// Should not suggest offsetting for low emissions
	assert.NotContains(t, recommendations, "Consider purchasing carbon offsets")
	assert.Contains(t, recommendations, "relatively low carbon footprint")
}

func TestCarbonCalculatorService_GenerateRecommendations_FossilFuel(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	// Create a fossil fuel profile
	profile := &domain.EnergyProfile{
		ConsensusType: domain.ConsensusProofOfWork,
		EnergySource:  domain.EnergySourceFossilFuel,
	}

	// Low emission - should still suggest renewable alternatives
	emission := domain.CarbonEmission{
		Value: 100.0,
		Unit:  domain.CarbonUnitGrams,
	}

	recommendations := calculator.generateRecommendations(profile, emission)

	// Should suggest looking for renewable-powered chains
	assert.Contains(t, recommendations, "Look for blockchain networks powered by renewable energy")
}

func TestCarbonCalculatorService_GetDefaultProfile(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	profile := calculator.getDefaultProfile("test-chain")

	assert.NotNil(t, profile)
	assert.Equal(t, "test-chain", profile.ChainID)
	assert.Equal(t, domain.ConsensusProofOfStake, profile.ConsensusType)
	assert.Equal(t, domain.EnergySourceMixed, profile.EnergySource)
	assert.True(t, profile.IsActive)
}

func TestCarbonCalculatorService_CalculateEnergyUsage_WithGasAdjustment(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	profile := &domain.EnergyProfile{
		AvgKWhPerTx: 0.001,
		EnergySource: domain.EnergySourceMixed,
	}

	req := domain.EnergyCalculationRequest{
		TransactionData: map[string]interface{}{
			"gas_used": 200000.0, // Double the average gas
		},
	}

	energy := calculator.calculateEnergyUsage(req, profile)

	assert.Equal(t, 0.002, energy.Value) // Should be doubled due to gas adjustment
	assert.Equal(t, domain.EnergyUnitKilowattHours, energy.Unit)
}

func TestCarbonCalculatorService_CalculateEnergyUsage_WithNetworkCongestion(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	profile := &domain.EnergyProfile{
		AvgKWhPerTx: 0.001,
		EnergySource: domain.EnergySourceMixed,
	}

	req := domain.EnergyCalculationRequest{
		TransactionData: map[string]interface{}{
			"network_congestion": 0.5, // 50% congestion
		},
	}

	energy := calculator.calculateEnergyUsage(req, profile)

	// Should be increased by (1 + 0.5 * 0.5) = 1.25x
	assert.Equal(t, 0.00125, energy.Value)
}

func TestCarbonCalculatorService_CalculateCarbonEmission_RenewableSource(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	energy := domain.EnergyConsumption{
		Value:  1.0, // 1 kWh
		Unit:   domain.EnergyUnitKilowattHours,
		Source: domain.EnergySourceSolar,
	}

	emission := calculator.calculateCarbonEmission(energy, domain.EnergySourceSolar, 0.5)

	// Solar has 0.05 multiplier, so 1 kWh * 0.5 * 0.05 = 0.025 kg = 25g
	assert.Equal(t, 25.0, emission.Value)
	assert.Equal(t, domain.CarbonUnitGrams, emission.Unit)
}

func TestCarbonCalculatorService_CalculateCarbonEmission_FossilFuelSource(t *testing.T) {
	profileRepo := new(MockEnergyProfileRepository)
	oracleRepo := new(MockExternalOracleRepository)

	calculator := NewCarbonCalculatorService(profileRepo, oracleRepo)

	energy := domain.EnergyConsumption{
		Value:  1.0, // 1 kWh
		Unit:   domain.EnergyUnitKilowattHours,
		Source: domain.EnergySourceFossilFuel,
	}

	emission := calculator.calculateCarbonEmission(energy, domain.EnergySourceFossilFuel, 0.5)

	// Fossil fuel has 1.0 multiplier, so 1 kWh * 0.5 * 1.0 = 0.5 kg = 500g
	assert.Equal(t, 500.0, emission.Value)
	assert.Equal(t, domain.CarbonUnitGrams, emission.Unit)
}

func TestGetSourceMultiplier(t *testing.T) {
	tests := []struct {
		source   domain.EnergySource
		expected float64
	}{
		{domain.EnergySourceRenewable, 0.05},
		{domain.EnergySourceSolar, 0.05},
		{domain.EnergySourceWind, 0.05},
		{domain.EnergySourceHydro, 0.05},
		{domain.EnergySourceNuclear, 0.02},
		{domain.EnergySourceMixed, 0.5},
		{domain.EnergySourceFossilFuel, 1.0},
		{"unknown", 0.5}, // Default case
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			result := getSourceMultiplier(tt.source)
			assert.Equal(t, tt.expected, result)
		})
	}
}
