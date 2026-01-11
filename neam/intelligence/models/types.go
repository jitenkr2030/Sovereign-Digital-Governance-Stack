package models

import (
	"encoding/json"
	"time"
)

// EconomicEvent represents a standardized input event from the Sensing Layer
type EconomicEvent struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	SourceID   string            `json:"source_id"`
	RegionID   string            `json:"region_id"`
	SectorID   string            `json:"sector_id"`
	MetricType MetricType        `json:"metric_type"`
	Value      float64           `json:"value"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// MetricType categorizes the type of economic measurement
type MetricType string

const (
	MetricPaymentVolume      MetricType = "PAYMENT_VOL"
	MetricPaymentValue       MetricType = "PAYMENT_VALUE"
	MetricEnergyConsumption  MetricType = "ENERGY_KWH"
	MetricEnergyPrice        MetricType = "ENERGY_PRICE"
	MetricCPI                MetricType = "CPI"
	MetricProducerPrice      MetricType = "PPI"
	MetricCommodityPrice     MetricType = "COMMODITY_PRICE"
	MetricInventoryLevel     MetricType = "INVENTORY_LEVEL"
	MetricEmployment         MetricType = "EMPLOYMENT"
	MetricManufacturingOutput MetricType = "MANUFACTURING_OUTPUT"
	MetricRetailSales        MetricType = "RETAIL_SALES"
	MetricCashTransaction    MetricType = "CASH_TRANSACTION"
	MetricDigitalTransaction MetricType = "DIGITAL_TRANSACTION"
)

// Signal represents an output from the intelligence engines
type Signal struct {
	ID              string        `json:"id"`
	Timestamp       time.Time     `json:"timestamp"`
	SignalType      SignalType    `json:"signal_type"`
	Severity        float64       `json:"severity"` // 0.0 to 1.0
	Confidence      float64       `json:"confidence"` // 0.0 to 1.0
	RegionID        string        `json:"region_id,omitempty"`
	SectorID        string        `json:"sector_id,omitempty"`
	TriggeringEvent *EconomicEvent `json:"triggering_event,omitempty"`
	Payload         SignalPayload `json:"payload"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SignalType categorizes the type of intelligence signal
type SignalType string

const (
	SignalAnomaly           SignalType = "ANOMALY"
	SignalInflationWarning  SignalType = "INFLATION_WARNING"
	SignalSlowdownWarning   SignalType = "SLOWDOWN_WARNING"
	SignalCorrelationBreak  SignalType = "CORRELATION_BREAK"
	SignalSupplyChainShock  SignalType = "SUPPLY_CHAIN_SHOCK"
	SignalBlackEconomyRisk  SignalType = "BLACK_ECONOMY_RISK"
	SignalRegionalStress    SignalType = "REGIONAL_STRESS"
	SignalDecouplingAlert   SignalType = "DECOUPLING_ALERT"
)

// SignalPayload contains detailed information about the signal
type SignalPayload struct {
	AnomalyScore      float64  `json:"anomaly_score,omitempty"`
	ExpectedValue     float64  `json:"expected_value,omitempty"`
	ActualValue       float64  `json:"actual_value,omitempty"`
	Threshold         float64  `json:"threshold,omitempty"`
	Prediction        float64  `json:"prediction,omitempty"`
	PredictionWindow  string   `json:"prediction_window,omitempty"`
	CorrelationCoeff  float64  `json:"correlation_coefficient,omitempty"`
	LeadingIndicators []string `json:"leading_indicators,omitempty"`
	Recommendations   []string `json:"recommendations,omitempty"`
}

// AnalysisContext carries analysis state through the pipeline
type AnalysisContext struct {
	WindowDuration    time.Duration
	BaselineStats     *BaselineStatistics
	RelatedMetrics    map[MetricType][]float64
	RegionID          string
	SectorID          string
}

// BaselineStatistics holds statistical baselines for comparison
type BaselineStatistics struct {
	Mean       float64
	StdDev     float64
	Median     float64
	Min        float64
	Max        float64
	Percentile map[float64]float64 // e.g., 0.95 -> value at 95th percentile
	Count      int
}

// DetectionRule defines a configurable detection rule stored in PostgreSQL
type DetectionRule struct {
	ID               string     `json:"id"`
	MetricType       MetricType `json:"metric_type"`
	Algorithm        string     `json:"algorithm"`
	ThresholdValue   float64    `json:"threshold_value"`
	LookbackMinutes  int        `json:"lookback_minutes"`
	SeverityWeight   float64    `json:"severity_weight"`
	RegionID         string     `json:"region_id,omitempty"`
	SectorID         string     `json:"sector_id,omitempty"`
	IsActive         bool       `json:"is_active"`
	CooldownSeconds  int        `json:"cooldown_seconds"`
}

// InflationData holds inflation-related data points
type InflationData struct {
	Timestamp        time.Time `json:"timestamp"`
	RegionID         string    `json:"region_id"`
	CPI              float64   `json:"cpi"`
	CPIYoY           float64   `json:"cpi_yoy"`
	CPIMoM           float64   `json:"cpi_mom"`
	PPI              float64   `json:"ppi"`
	CoreCPI          float64   `json:"core_cpi"`
	CommodityIndex   float64   `json:"commodity_index"`
	ProducerPrice    float64   `json:"producer_price"`
}

// InflationPrediction holds the output of the inflation engine
type InflationPrediction struct {
	Timestamp           time.Time  `json:"timestamp"`
	RegionID            string     `json:"region_id"`
	CurrentInflation    float64    `json:"current_inflation"`
	Predicted30Day      float64    `json:"predicted_30_day"`
	Predicted90Day      float64    `json:"predicted_90_day"`
	Predicted180Day     float64    `json:"predicted_180_day"`
	Confidence          float64    `json:"confidence"`
	RiskLevel           string     `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	LeadingIndicators   []string   `json:"leading_indicators"`
	ContributingFactors []string   `json:"contributing_factors"`
}

// CorrelationData holds cross-sector correlation metrics
type CorrelationData struct {
	Timestamp       time.Time  `json:"timestamp"`
	MetricA         MetricType `json:"metric_a"`
	MetricB         MetricType `json:"metric_b"`
	RegionID        string     `json:"region_id"`
	PearsonCoeff    float64    `json:"pearson_coefficient"`
	SpearmanCoeff   float64    `json:"spearman_coefficient"`
	WindowDays      int        `json:"window_days"`
	HistoricalAvg   float64    `json:"historical_average"`
	Deviation       float64    `json:"deviation_from_historical"`
}

// RegionalStressIndex represents economic stress by region
type RegionalStressIndex struct {
	Timestamp     time.Time `json:"timestamp"`
	RegionID      string    `json:"region_id"`
	OverallScore  float64   `json:"overall_score"` // 0-100 normalized
	AnomalyCount  int       `json:"anomaly_count"`
	AnomalyScore  float64   `json:"anomaly_score"`
	InflationScore float64  `json:"inflation_score"`
	CorrelationBreaks int   `json:"correlation_breaks"`
}

// TimeSeriesPoint represents a single point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// ToJSON converts a Signal to JSON bytes
func (s *Signal) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON parses JSON bytes into a Signal
func (s *Signal) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// AnomalySeverity classifies the severity of an anomaly
type AnomalySeverity string

const (
	SeverityLow      AnomalySeverity = "LOW"
	SeverityMedium   AnomalySeverity = "MEDIUM"
	SeverityHigh     AnomalySeverity = "HIGH"
	SeverityCritical AnomalySeverity = "CRITICAL"
)

// CalculateSeverity converts a numeric score to severity level
func CalculateSeverity(score float64) AnomalySeverity {
	switch {
	case score < 0.25:
		return SeverityLow
	case score < 0.5:
		return SeverityMedium
	case score < 0.75:
		return SeverityHigh
	default:
		return SeverityCritical
	}
}

// ConfidenceLevel represents the confidence in a detection
type ConfidenceLevel float64

const (
	ConfidenceLow    ConfidenceLevel = 0.5
	ConfidenceMedium ConfidenceLevel = 0.7
	ConfidenceHigh   ConfidenceLevel = 0.9
)
