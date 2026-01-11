package services

import (
	"context"
	"fmt"
	"math"

	"github.com/energy-integration/energy/internal/core/domain"
	"github.com/energy-integration/energy/internal/core/ports"
)

// CarbonCalculatorService provides carbon emission calculation logic.
type CarbonCalculatorService struct {
	profileRepo ports.EnergyProfileRepository
	oracleRepo  ports.ExternalOracleRepository
}

// NewCarbonCalculatorService creates a new carbon calculator service.
func NewCarbonCalculatorService(
	profileRepo ports.EnergyProfileRepository,
	oracleRepo ports.ExternalOracleRepository,
) *CarbonCalculatorService {
	return &CarbonCalculatorService{
		profileRepo: profileRepo,
		oracleRepo:  oracleRepo,
	}
}

// CalculateFootprint calculates the carbon footprint for a transaction.
func (s *CarbonCalculatorService) CalculateFootprint(
	ctx context.Context,
	req domain.EnergyCalculationRequest,
) (*domain.EnergyCalculationResult, error) {
	// Get energy profile for the chain
	profile, err := s.profileRepo.GetByChainID(ctx, req.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get energy profile: %w", err)
	}

	if profile == nil {
		// Use default values for unknown chains
		profile = s.getDefaultProfile(req.ChainID)
	}

	// Calculate base energy consumption
	energyUsed := s.calculateEnergyUsage(req, profile)

	// Get carbon intensity (from profile or external oracle)
	carbonIntensity := profile.CarbonIntensity
	if req.NodeLocation != "" && s.oracleRepo != nil {
		if oracleIntensity, err := s.oracleRepo.GetCarbonIntensity(ctx, req.NodeLocation); err == nil {
			carbonIntensity = oracleIntensity
		}
	}

	// Calculate carbon emission
	carbonEmission := s.calculateCarbonEmission(energyUsed, profile.EnergySource, carbonIntensity)

	// Generate comparison data for context
	comparisonData := s.generateComparisonData(carbonEmission)

	// Generate recommendations
	recommendations := s.generateRecommendations(profile, carbonEmission)

	return &domain.EnergyCalculationResult{
		EnergyUsed:      energyUsed,
		CarbonEmission:  carbonEmission,
		ComparisonData:  comparisonData,
		Recommendations: recommendations,
	}, nil
}

// calculateEnergyUsage calculates energy consumption for a transaction.
func (s *CarbonCalculatorService) calculateEnergyUsage(
	req domain.EnergyCalculationRequest,
	profile *domain.EnergyProfile,
) domain.EnergyConsumption {
	// Base calculation using average kWh per transaction
	energyKWh := profile.AvgKWhPerTx

	// Adjust based on transaction complexity (if data available)
	if txData, ok := req.TransactionData["gas_used"]; ok {
		if gasUsed, ok := txData.(float64); ok {
			// For EVM chains, adjust based on actual gas used
			// Assuming 100,000 gas = 1 tx average
			gasMultiplier := gasUsed / 100000.0
			energyKWh *= gasMultiplier
		}
	}

	// Adjust based on network load
	if txData, ok := req.TransactionData["network_congestion"]; ok {
		if congestion, ok := txData.(float64); ok {
			// Higher congestion means more energy per transaction
			energyKWh *= (1 + congestion*0.5)
		}
	}

	return domain.EnergyConsumption{
		Value:  energyKWh,
		Unit:  domain.EnergyUnitKilowattHours,
		Source: profile.EnergySource,
	}
}

// calculateCarbonEmission calculates carbon emission from energy consumption.
func (s *CarbonCalculatorService) calculateCarbonEmission(
	energy domain.EnergyConsumption,
	source domain.EnergySource,
	intensity float64,
) domain.CarbonEmission {
	// Convert energy to kWh if needed
	energyKWh := energy.Value
	if energy.Unit == domain.EnergyUnitWattHours {
		energyKWh = energy.Value / 1000
	} else if energy.Unit == domain.EnergyUnitMegawattHours {
		energyKWh = energy.Value * 1000
	}

	// Calculate carbon emission
	emissionValue := energyKWh * intensity * getSourceMultiplier(source)

	return domain.CarbonEmission{
		Value:    emissionValue * 1000, // Convert kg to grams
		Unit:     domain.CarbonUnitGrams,
		FactorID: fmt.Sprintf("intensity_%s_%.2f", source, intensity),
	}
}

// generateComparisonData creates relatable comparisons for the carbon emission.
func (s *CarbonCalculatorService) generateComparisonData(emission domain.CarbonEmission) *domain.ComparisonData {
	// Convert to kg for easier comparison
	emissionKg := emission.Value / 1000

	// Average car emits ~120g CO2 per km
	equivalentKm := emissionKg / 0.12

	// A mature tree absorbs ~22kg CO2 per year
	equivalentTreeDays := (emissionKg / 22.0) * 365

	// Smartphone battery ~15Wh per full charge, emits ~7.5g CO2
	equivalentCharges := int64(emissionKg / 0.0075)

	// LED bulb ~10W, emits ~5g CO2 per hour
	equivalentBulbHours := emissionKg / 0.005

	return &domain.ComparisonData{
		EquivalentKmByCar:          math.Round(equivalentKm*100) / 100,
		EquivalentTreeDays:         math.Round(equivalentTreeDays*100) / 100,
		EquivalentSmartphoneCharges: equivalentCharges,
		EquivalentLEDBulbHours:     math.Round(equivalentBulbHours*100) / 100,
	}
}

// generateRecommendations generates suggestions to reduce carbon footprint.
func (s *CarbonCalculatorService) generateRecommendations(
	profile *domain.EnergyProfile,
	emission domain.CarbonEmission,
) []string {
	var recommendations []string

	// PoW chains have higher footprint
	if profile.ConsensusType == domain.ConsensusProofOfWork {
		recommendations = append(recommendations,
			"Consider using proof-of-stake alternatives which typically have 99% lower energy consumption.",
		)
	}

	// Suggest offsetting for significant emissions
	if emission.Value > 10000 { // More than 10kg
		recommendations = append(recommendations,
			"Consider purchasing carbon offsets to neutralize this transaction's environmental impact.",
			"Batch multiple transactions to reduce per-transaction footprint.",
		)
	}

	// Renewable energy suggestions
	if profile.EnergySource == domain.EnergySourceFossilFuel || profile.EnergySource == domain.EnergySourceMixed {
		recommendations = append(recommendations,
			"Look for blockchain networks powered by renewable energy sources.",
		)
	}

	// Add general recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"This transaction has a relatively low carbon footprint.",
		)
	}

	return recommendations
}

// getDefaultProfile returns a default energy profile for unknown chains.
func (s *CarbonCalculatorService) getDefaultProfile(chainID string) *domain.EnergyProfile {
	return &domain.EnergyProfile{
		ChainID:         chainID,
		ChainName:       chainID,
		ConsensusType:   domain.ConsensusProofOfStake,
		AvgKWhPerTx:     0.001, // Very low default (like PoS)
		BaseCarbonGrams: 0.5,   // 0.5kg CO2 default
		EnergySource:    domain.EnergySourceMixed,
		CarbonIntensity: 0.5,   // Global average
		IsActive:        true,
	}
}

// getSourceMultiplier returns carbon intensity multiplier based on energy source.
func getSourceMultiplier(source domain.EnergySource) float64 {
	switch source {
	case domain.EnergySourceRenewable, domain.EnergySourceSolar, domain.EnergySourceWind, domain.EnergySourceHydro:
		return 0.05
	case domain.EnergySourceNuclear:
		return 0.02
	case domain.EnergySourceMixed:
		return 0.5
	case domain.EnergySourceFossilFuel:
		return 1.0
	default:
		return 0.5
	}
}
