package anomaly

import (
	"neam-platform/intelligence/models"
)

// BlackEconomyRiskDetector detects potential black economy activity
// Trigger: Cash volume increases while digital volume remains flat or decreases
type BlackEconomyRiskDetector struct {
	CashIncreaseThreshold    float64 // e.g., 0.20 (20% increase)
	DigitalStagnationThreshold float64 // e.g., 0.05 (5% change max)
	LookbackDays             int
}

// NewBlackEconomyRiskDetector creates a new black economy risk detector
func NewBlackEconomyRiskDetector(cashThreshold, digitalThreshold float64, lookbackDays int) *BlackEconomyRiskDetector {
	return &BlackEconomyRiskDetector{
		CashIncreaseThreshold:     cashThreshold,
		DigitalStagnationThreshold: digitalThreshold,
		LookbackDays:              lookbackDays,
	}
}

// GetName returns the detector name
func (b *BlackEconomyRiskDetector) GetName() string {
	return "BLACK_ECONOMY_RISK"
}

// Configure sets detector parameters
func (b *BlackEconomyRiskDetector) Configure(params map[string]interface{}) error {
	// Configuration handled in constructor for domain-specific detectors
	return nil
}

// Detect identifies potential black economy indicators
func (b *BlackEconomyRiskDetector) Detect(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	// This detector needs context from multiple metrics
	// In practice, it would be called with aggregated data
	return false, 0.0, 0.0
}

// BlackEconomyContext holds multi-metric context for black economy detection
type BlackEconomyContext struct {
	CashVolumeHistory    []models.TimeSeriesPoint
	DigitalVolumeHistory []models.TimeSeriesPoint
	RetailSalesHistory   []models.TimeSeriesPoint
	CurrentCashVolume    float64
	CurrentDigitalVolume float64
}

// DetectWithContext performs detection with full context
func (b *BlackEconomyRiskDetector) DetectWithContext(ctx BlackEconomyContext) (bool, float64, float64) {
	if len(ctx.CashVolumeHistory) < 7 || len(ctx.DigitalVolumeHistory) < 7 {
		return false, 0.0, 0.5
	}
	
	// Calculate averages for the lookback period
	cashAvg := calculateAverage(ctx.CashVolumeHistory)
	digitalAvg := calculateAverage(ctx.DigitalVolumeHistory)
	
	if digitalAvg == 0 {
		return false, 0.0, 0.6 // Cannot compute ratio with zero
	}
	
	// Calculate changes
	cashChange := (ctx.CurrentCashVolume - cashAvg) / cashAvg
	digitalChange := (ctx.CurrentDigitalVolume - digitalAvg) / digitalAvg
	
	// Black economy indicator: high cash growth with stagnant digital growth
	cashThresholdMet := cashChange > b.CashIncreaseThreshold
	digitalStagnant := mathAbs(digitalChange) < b.DigitalStagnationThreshold
	
	if cashThresholdMet && digitalStagnant {
		// Calculate severity based on magnitude
		severity := (cashChange * (1 - mathAbs(digitalChange))) / 2
		severity = min(1.0, max(0.0, severity))
		
		// Confidence based on data quality
		confidence := 0.7 + 0.2*min(1.0, float64(len(ctx.CashVolumeHistory))/14)
		
		return true, severity, confidence
	}
	
	return false, 0.0, 0.85
}

// SupplyChainShockDetector identifies supply chain disruptions
// Trigger: Inventory levels drop significantly while prices rise
type SupplyChainShockDetector struct {
	InventoryDropThreshold float64 // e.g., 0.20 (20% drop)
	PriceIncreaseThreshold float64 // e.g., 0.05 (5% increase)
	LookbackDays           int
}

// NewSupplyChainShockDetector creates a new supply chain shock detector
func NewSupplyChainShockDetector(inventoryDrop, priceIncrease float64, lookbackDays int) *SupplyChainShockDetector {
	return &SupplyChainShockDetector{
		InventoryDropThreshold: inventoryDrop,
		PriceIncreaseThreshold: priceIncrease,
		LookbackDays:           lookbackDays,
	}
}

// GetName returns the detector name
func (s *SupplyChainShockDetector) GetName() string {
	return "SUPPLY_CHAIN_SHOCK"
}

// Configure sets detector parameters
func (s *SupplyChainShockDetector) Configure(params map[string]interface{}) error {
	return nil
}

// SupplyChainContext holds multi-metric context for supply chain detection
type SupplyChainContext struct {
	InventoryHistory []models.TimeSeriesPoint
	PriceHistory     []models.TimeSeriesPoint
	CurrentInventory float64
	CurrentPrice     float64
	RegionID         string
	SectorID         string
}

// DetectWithContext performs detection with full context
func (s *SupplyChainShockDetector) DetectWithContext(ctx SupplyChainContext) (bool, float64, float64) {
	if len(ctx.InventoryHistory) < 7 || len(ctx.PriceHistory) < 7 {
		return false, 0.0, 0.5
	}
	
	inventoryAvg := calculateAverage(ctx.InventoryHistory)
	priceAvg := calculateAverage(ctx.PriceHistory)
	
	if inventoryAvg == 0 {
		return false, 0.0, 0.6
	}
	
	// Calculate percentage changes
	inventoryChange := (ctx.CurrentInventory - inventoryAvg) / inventoryAvg
	priceChange := (ctx.CurrentPrice - priceAvg) / priceAvg
	
	// Supply chain shock: inventory drops AND prices rise
	inventoryDrop := inventoryChange < -s.InventoryDropThreshold
	priceRise := priceChange > s.PriceIncreaseThreshold
	
	if inventoryDrop && priceRise {
		// Severity based on the magnitude of both factors
		severity := (-inventoryChange + priceChange) / 2
		severity = min(1.0, max(0.0, severity))
		
		confidence := 0.75 + 0.15*min(1.0, float64(len(ctx.InventoryHistory))/14)
		
		return true, severity, confidence
	}
	
	return false, 0.0, 0.8
}

// EnergyPaymentDecouplingDetector detects when energy usage doesn't correlate with payments
// Trigger: Strong historical correlation breaks (potential meter tampering or theft)
type EnergyPaymentDecouplingDetector struct {
	HistoricalCorrelationThreshold float64 // e.g., 0.8
	DecouplingThreshold           float64 // e.g., 0.3 (correlation drops below this)
	MinCorrelationHistory         int
}

// NewEnergyPaymentDecouplingDetector creates a new decoupling detector
func NewEnergyPaymentDecouplingDetector(historicalCorr, decouplingThreshold float64, minHistory int) *EnergyPaymentDecouplingDetector {
	return &EnergyPaymentDecouplingDetector{
		HistoricalCorrelationThreshold: historicalCorr,
		DecouplingThreshold:           decouplingThreshold,
		MinCorrelationHistory:         minHistory,
	}
}

// GetName returns the detector name
func (e *EnergyPaymentDecouplingDetector) GetName() string {
	return "ENERGY_PAYMENT_DECOUPLING"
}

// Configure sets detector parameters
func (e *EnergyPaymentDecouplingDetector) Configure(params map[string]interface{}) error {
	return nil
}

// DecouplingContext holds correlation data for decoupling detection
type DecouplingContext struct {
	EnergyHistory      []models.TimeSeriesPoint
	PaymentHistory     []models.TimeSeriesPoint
	HistoricalCorr     float64
	CurrentCorr        float64
	RegionID           string
}

// DetectWithContext performs detection with full context
func (e *EnergyPaymentDecouplingDetector) DetectWithContext(ctx DecouplingContext) (bool, float64, float64) {
	if len(ctx.EnergyHistory) < e.MinCorrelationHistory || len(ctx.PaymentHistory) < e.MinCorrelationHistory {
		return false, 0.0, 0.5
	}
	
	// Check if there was historically strong correlation
	if ctx.HistoricalCorrelation < e.HistoricalCorrelationThreshold {
		return false, 0.0, 0.6 // No baseline correlation to break
	}
	
	// Check if correlation has decoupled
	if ctx.CurrentCorr < e.DecouplingThreshold {
		// Calculate severity based on the drop
		correlationDrop := ctx.HistoricalCorrelation - ctx.CurrentCorr
		severity := min(1.0, correlationDrop/e.HistoricalCorrelation)
		
		confidence := 0.8
		if ctx.HistoricalCorrelation > 0.9 {
			confidence = 0.9
		}
		
		return true, severity, confidence
	}
	
	return false, 0.0, 0.85
}

// SpendingSurgeDetector detects unusual spending increases
type SpendingSurgeDetector struct {
	SurgeThreshold   float64 // e.g., 0.50 (50% increase)
	SustainedWindow  int     // How many periods to confirm
	LookbackPeriods  int
}

// NewSpendingSurgeDetector creates a new spending surge detector
func NewSpendingSurgeDetector(surgeThreshold float64, sustainedWindow, lookbackPeriods int) *SpendingSurgeDetector {
	return &SpendingSurgeDetector{
		SurgeThreshold:  surgeThreshold,
		SustainedWindow: sustainedWindow,
		LookbackPeriods: lookbackPeriods,
	}
}

// GetName returns the detector name
func (s *SpendingSurgeDetector) GetName() string {
	return "SPENDING_SURGE"
}

// DetectWithContext performs detection with full context
func (s *SpendingSurgeDetector) DetectWithContext(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	if len(history) < s.LookbackPeriods {
		return false, 0.0, 0.5
	}
	
	avg := calculateAverage(history)
	if avg == 0 {
		return false, 0.0, 0.6
	}
	
	change := (current - avg) / avg
	
	if change > s.SurgeThreshold {
		// Check if this is sustained (multiple periods of increase)
		sustainedCount := 0
		recentPeriods := history[len(history)-s.SustainedWindow:]
		for _, point := range recentPeriods {
			pointChange := (point.Value - avg) / avg
			if pointChange > s.SurgeThreshold*0.5 {
				sustainedCount++
			}
		}
		
		// Severity increases with sustained pattern
		sustainedRatio := float64(sustainedCount) / float64(s.SustainedWindow)
		severity := min(1.0, change*sustainedRatio)
		
		confidence := 0.7 + 0.2*sustainedRatio
		
		return true, severity, confidence
	}
	
	return false, 0.0, 0.85
}

// SpendingDropDetector detects unusual spending decreases (potential slowdown indicator)
type SpendingDropDetector struct {
	DropThreshold    float64 // e.g., 0.30 (30% decrease)
	LookbackPeriods  int
	SeverityWindow   int
}

// NewSpendingDropDetector creates a new spending drop detector
func NewSpendingDropDetector(dropThreshold float64, lookbackPeriods, severityWindow int) *SpendingDropDetector {
	return &SpendingDropDetector{
		DropThreshold:   dropThreshold,
		LookbackPeriods: lookbackPeriods,
		SeverityWindow:  severityWindow,
	}
}

// GetName returns the detector name
func (s *SpendingDropDetector) GetName() string {
	return "SPENDING_DROP"
}

// DetectWithContext performs detection with full context
func (s *SpendingDropDetector) DetectWithContext(history []models.TimeSeriesPoint, current float64) (bool, float64, float64) {
	if len(history) < s.LookbackPeriods {
		return false, 0.0, 0.5
	}
	
	avg := calculateAverage(history)
	if avg == 0 {
		return false, 0.0, 0.6
	}
	
	change := (current - avg) / avg // Will be negative for drops
	
	if change < -s.DropThreshold {
		// Check for sustained drop
		sustainedCount := 0
		recentPeriods := history[len(history)-s.SeverityWindow:]
		for _, point := range recentPeriods {
			pointChange := (point.Value - avg) / avg
			if pointChange < -s.DropThreshold*0.5 {
				sustainedCount++
			}
		}
		
		// Severity based on magnitude and persistence
		sustainedRatio := float64(sustainedCount) / float64(s.SeverityWindow)
		baseSeverity := -change // Convert to positive
		severity := min(1.0, baseSeverity*sustainedRatio)
		
		confidence := 0.75 + 0.15*sustainedRatio
		
		return true, severity, confidence
	}
	
	return false, 0.0, 0.85
}

// Helper function to calculate average
func calculateAverage(points []models.TimeSeriesPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, p := range points {
		sum += p.Value
	}
	return sum / float64(len(points))
}

// Helper function for absolute value
func mathAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function for min
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Helper function for max
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
