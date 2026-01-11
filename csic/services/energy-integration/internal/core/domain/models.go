package domain

import (
	"time"
)

// EnergySource represents the type of energy source used for mining/validation.
type EnergySource string

const (
	EnergySourceRenewable    EnergySource = "renewable"
	EnergySourceMixed        EnergySource = "mixed"
	EnergySourceFossilFuel   EnergySource = "fossil_fuel"
	EnergySourceNuclear      EnergySource = "nuclear"
	EnergySourceHydro        EnergySource = "hydro"
	EnergySourceSolar        EnergySource = "solar"
	EnergySourceWind         EnergySource = "wind"
)

// ConsensusType represents the blockchain consensus mechanism.
type ConsensusType string

const (
	ConsensusProofOfWork      ConsensusType = "proof_of_work"
	ConsensusProofOfStake     ConsensusType = "proof_of_stake"
	ConsensusDelegatedPoS     ConsensusType = "delegated_proof_of_stake"
	ConsensusProofOfAuthority ConsensusType = "proof_of_authority"
	ConsensusProofOfHistory   ConsensusType = "proof_of_history"
)

// OffsetStatus represents the carbon offset status of a transaction.
type OffsetStatus string

const (
	OffsetStatusNone     OffsetStatus = "none"
	OffsetStatusPartial  OffsetStatus = "partial"
	OffsetStatusFull     OffsetStatus = "full"
	OffsetStatusPending  OffsetStatus = "pending"
)

// CarbonUnit represents the unit of carbon measurement.
type CarbonUnit string

const (
	CarbonUnitGrams    CarbonUnit = "g"
	CarbonUnitKilograms CarbonUnit = "kg"
	CarbonUnitTonnes   CarbonUnit = "t"
)

// EnergyUnit represents the unit of energy measurement.
type EnergyUnit string

const (
	EnergyUnitWattHours     EnergyUnit = "Wh"
	EnergyUnitKilowattHours EnergyUnit = "kWh"
	EnergyUnitMegawattHours EnergyUnit = "MWh"
)

// EntityID is a type alias for entity identifiers.
type EntityID string

// NewEntityID generates a new unique entity identifier.
func NewEntityID() EntityID {
	return EntityID(generateUUID())
}

// CarbonEmission represents a calculated carbon emission value.
type CarbonEmission struct {
	Value    float64     `json:"value"`
	Unit     CarbonUnit  `json:"unit"`
	FactorID string      `json:"factor_id,omitempty"`
}

// EnergyConsumption represents energy usage data.
type EnergyConsumption struct {
	Value    float64     `json:"value"`
	Unit     EnergyUnit  `json:"unit"`
	Source   EnergySource `json:"source"`
}

// EnergyProfile represents the energy characteristics of a blockchain network.
type EnergyProfile struct {
	ID               EntityID        `json:"id"`
	ChainID          string          `json:"chain_id"`
	ChainName        string          `json:"chain_name"`
	ConsensusType    ConsensusType   `json:"consensus_type"`
	AvgKWhPerTx      float64         `json:"avg_kwh_per_tx"`
	BaseCarbonGrams  float64         `json:"base_carbon_grams"`
	EnergySource     EnergySource    `json:"energy_source"`
	CarbonIntensity  float64         `json:"carbon_intensity"` // kg CO2 per kWh
	NodeLocations    []string        `json:"node_locations"`
	NetworkHashRate  float64         `json:"network_hash_rate"`
	TransactionsPerSecond float64    `json:"transactions_per_second"`
	IsActive         bool            `json:"is_active"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// TransactionFootprint represents the environmental impact of a single transaction.
type TransactionFootprint struct {
	ID                EntityID            `json:"id"`
	TransactionHash   string              `json:"transaction_hash"`
	ChainID           string              `json:"chain_id"`
	BlockNumber       int64               `json:"block_number"`
	EnergyUsed        EnergyConsumption   `json:"energy_used"`
	CarbonEmission    CarbonEmission      `json:"carbon_emission"`
	OffsetStatus      OffsetStatus        `json:"offset_status"`
	OffsetAmount      *CarbonEmission     `json:"offset_amount,omitempty"`
	CertificateID     string              `json:"certificate_id,omitempty"`
	Timestamp         time.Time           `json:"timestamp"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
}

// OffsetCertificate represents a renewable energy certificate.
type OffsetCertificate struct {
	ID                EntityID            `json:"id"`
	CertificateNumber string              `json:"certificate_number"`
	EnergySource      EnergySource        `json:"energy_source"`
	EnergyAmount      float64             `json:"energy_amount"`
	EnergyUnit        EnergyUnit          `json:"energy_unit"`
	CarbonOffset      CarbonEmission      `json:"carbon_offset"`
	Registry          string              `json:"registry"` // e.g., "I-REC", "GSE"
	IssueDate         time.Time           `json:"issue_date"`
	ExpiryDate        time.Time           `json:"expiry_date"`
	Status            string              `json:"status"` // "active", "expired", "redeemed"
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
}

// CarbonFootprintSummary represents aggregated footprint data.
type CarbonFootprintSummary struct {
	PeriodStart       time.Time           `json:"period_start"`
	PeriodEnd         time.Time           `json:"period_end"`
	ChainID           string              `json:"chain_id,omitempty"`
	TotalTransactions int64               `json:"total_transactions"`
	TotalEnergyKWh    float64             `json:"total_energy_kwh"`
	TotalCarbonGrams  float64             `json:"total_carbon_grams"`
	OffsetCarbonGrams float64             `json:"offset_carbon_grams"`
	NetCarbonGrams    float64             `json:"net_carbon_grams"`
	AverageCarbonPerTx float64            `json:"average_carbon_per_tx"`
	BreakdownByChain  map[string]float64  `json:"breakdown_by_chain,omitempty"`
}

// EnergyCalculationRequest represents input data for footprint calculation.
type EnergyCalculationRequest struct {
	ChainID        string                  `json:"chain_id"`
	TransactionData map[string]interface{} `json:"transaction_data"`
	NodeLocation   string                  `json:"node_location,omitempty"`
	Timestamp      time.Time               `json:"timestamp"`
}

// EnergyCalculationResult represents the calculated footprint.
type EnergyCalculationResult struct {
	EnergyUsed       EnergyConsumption  `json:"energy_used"`
	CarbonEmission   CarbonEmission     `json:"carbon_emission"`
	ComparisonData   *ComparisonData    `json:"comparison_data,omitempty"`
	Recommendations  []string           `json:"recommendations,omitempty"`
}

// ComparisonData provides context for the calculated footprint.
type ComparisonData struct {
	EquivalentKmByCar   float64 `json:"equivalent_km_by_car"`
	EquivalentTreeDays  float64 `json:"equivalent_tree_days"`
	EquivalentSmartphoneCharges int64 `json:"equivalent_smartphone_charges"`
	EquivalentLEDBulbHours float64 `json:"equivalent_led_bulb_hours"`
}

// calculateDefaultCarbonEmission calculates carbon emission based on average values.
func CalculateDefaultCarbonEmission(energyKWh float64, source EnergySource, intensity float64) CarbonEmission {
	// Default carbon intensity: 0.5 kg CO2 per kWh (global average)
	if intensity == 0 {
		intensity = 0.5
	}

	// Adjust based on energy source
	sourceMultiplier := getSourceMultiplier(source)
	adjustedIntensity := intensity * sourceMultiplier

	return CarbonEmission{
		Value: energyKWh * adjustedIntensity * 1000, // Convert to grams
		Unit:  CarbonUnitGrams,
	}
}

// getSourceMultiplier returns carbon intensity multiplier based on energy source.
func getSourceMultiplier(source EnergySource) float64 {
	switch source {
	case EnergySourceRenewable, EnergySourceSolar, EnergySourceWind, EnergySourceHydro:
		return 0.05 // 95% cleaner than fossil fuels
	case EnergySourceNuclear:
		return 0.02 // Very low carbon
	case EnergySourceMixed:
		return 0.5 // Average grid mix
	case EnergySourceFossilFuel:
		return 1.0 // Baseline
	default:
		return 0.5 // Default assumption
	}
}

// IsValidEnergySource checks if the energy source is valid.
func IsValidEnergySource(source EnergySource) bool {
	switch source {
	case EnergySourceRenewable, EnergySourceMixed, EnergySourceFossilFuel,
		EnergySourceNuclear, EnergySourceHydro, EnergySourceSolar, EnergySourceWind:
		return true
	default:
		return false
	}
}

// IsValidConsensusType checks if the consensus type is valid.
func IsValidConsensusType(consensus ConsensusType) bool {
	switch consensus {
	case ConsensusProofOfWork, ConsensusProofOfStake, ConsensusDelegatedPoS,
		ConsensusProofOfAuthority, ConsensusProofOfHistory:
		return true
	default:
		return false
	}
}
