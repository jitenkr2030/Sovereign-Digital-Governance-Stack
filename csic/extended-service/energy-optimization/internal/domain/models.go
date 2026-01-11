package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// EnergyReading represents a single energy consumption reading from a meter
type EnergyReading struct {
	ID              string          `json:"id" db:"id"`
	MeterID         string          `json:"meter_id" db:"meter_id"`
	FacilityID      string          `json:"facility_id" db:"facility_id"`
	EnergySource    EnergySource    `json:"energy_source" db:"energy_source"`
	Consumption     decimal.Decimal `json:"consumption" db:"consumption"` // in kWh
	Voltage         decimal.Decimal `json:"voltage" db:"voltage"`         // in Volts
	Current         decimal.Decimal `json:"current" db:"current"`         // in Amperes
	PowerFactor     decimal.Decimal `json:"power_factor" db:"power_factor"`
	Frequency       decimal.Decimal `json:"frequency" db:"frequency"`     // in Hz
	Timestamp       time.Time       `json:"timestamp" db:"timestamp"`
	Quality         ReadingQuality  `json:"quality" db:"quality"`
	Metadata        JSONMap         `json:"metadata" db:"metadata"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// EnergySource represents the type of energy source
type EnergySource string

const (
	EnergySourceGrid      EnergySource = "grid"
	EnergySourceSolar     EnergySource = "solar"
	EnergySourceWind      EnergySource = "wind"
	EnergySourceNaturalGas EnergySource = "natural_gas"
	EnergySourceCoal      EnergySource = "coal"
	EnergySourceNuclear   EnergySource = "nuclear"
	EnergySourceHydro     EnergySource = "hydro"
	EnergySourceBattery   EnergySource = "battery"
)

// ReadingQuality represents the quality of an energy reading
type ReadingQuality string

const (
	ReadingQualityGood    ReadingQuality = "good"
	ReadingQualitySuspect ReadingQuality = "suspect"
	ReadingQualityBad     ReadingQuality = "bad"
	ReadingQualityEstimated ReadingQuality = "estimated"
)

// CarbonEmission represents calculated carbon emissions for a reading
type CarbonEmission struct {
	ID                  string          `json:"id" db:"id"`
	ReadingID           string          `json:"reading_id" db:"reading_id"`
	FacilityID          string          `json:"facility_id" db:"facility_id"`
	EnergySource        EnergySource    `json:"energy_source" db:"energy_source"`
	EmissionFactor      decimal.Decimal `json:"emission_factor" db:"emission_factor"` // kg CO2 per kWh
	Consumption         decimal.Decimal `json:"consumption" db:"consumption"`         // kWh
	TotalEmission       decimal.Decimal `json:"total_emission" db:"total_emission"`   // kg CO2
	CarbonIntensity     decimal.Decimal `json:"carbon_intensity" db:"carbon_intensity"` // g CO2 per kWh
	Timestamp           time.Time       `json:"timestamp" db:"timestamp"`
	CalculatedAt        time.Time       `json:"calculated_at" db:"calculated_at"`
}

// CarbonBudget represents a facility's carbon budget and tracking
type CarbonBudget struct {
	ID                  string          `json:"id" db:"id"`
	FacilityID          string          `json:"facility_id" db:"facility_id"`
	BudgetPeriod        BudgetPeriod    `json:"budget_period" db:"budget_period"`
	BudgetAmount        decimal.Decimal `json:"budget_amount" db:"budget_amount"`         // kg CO2
	UsedAmount          decimal.Decimal `json:"used_amount" db:"used_amount"`             // kg CO2
	RemainingAmount     decimal.Decimal `json:"remaining_amount" db:"remaining_amount"`   // kg CO2
	UsagePercentage     decimal.Decimal `json:"usage_percentage" db:"usage_percentage"`
	Status              BudgetStatus    `json:"status" db:"status"`
	WarningThreshold    decimal.Decimal `json:"warning_threshold" db:"warning_threshold"`   // percentage
	CriticalThreshold   decimal.Decimal `json:"critical_threshold" db:"critical_threshold"` // percentage
	PeriodStart         time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd           time.Time       `json:"period_end" db:"period_end"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// BudgetPeriod represents the time period for a carbon budget
type BudgetPeriod string

const (
	BudgetPeriodDaily   BudgetPeriod = "daily"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
	BudgetPeriodQuarterly BudgetPeriod = "quarterly"
	BudgetPeriodYearly  BudgetPeriod = "yearly"
)

// BudgetStatus represents the status of a carbon budget
type BudgetStatus string

const (
	BudgetStatusNormal     BudgetStatus = "normal"
	BudgetStatusWarning    BudgetStatus = "warning"
	BudgetStatusCritical   BudgetStatus = "critical"
	BudgetStatusExceeded   BudgetStatus = "exceeded"
)

// LoadMetric represents aggregated load metrics for a facility or zone
type LoadMetric struct {
	ID                  string          `json:"id" db:"id"`
	FacilityID          string          `json:"facility_id" db:"facility_id"`
	ZoneID              string          `json:"zone_id" db:"zone_id"`
	Timestamp           time.Time       `json:"timestamp" db:"timestamp"`
	WindowDuration      time.Duration   `json:"window_duration" db:"window_duration"`
	AvgConsumption      decimal.Decimal `json:"avg_consumption" db:"avg_consumption"`       // kW
	PeakConsumption     decimal.Decimal `json:"peak_consumption" db:"peak_consumption"`     // kW
	MinConsumption      decimal.Decimal `json:"min_consumption" db:"min_consumption"`       // kW
	TotalConsumption    decimal.Decimal `json:"total_consumption" db:"total_consumption"`   // kWh
	AvgUtilization      decimal.Decimal `json:"avg_utilization" db:"avg_utilization"`       // percentage
	PeakUtilization     decimal.Decimal `json:"peak_utilization" db:"peak_utilization"`     // percentage
	ReactivePower       decimal.Decimal `json:"reactive_power" db:"reactive_power"`         // kVAR
	HarmonicDistortion  decimal.Decimal `json:"harmonic_distortion" db:"harmonic_distortion"`
	PowerFactor         decimal.Decimal `json:"power_factor" db:"power_factor"`
	SampleCount         int             `json:"sample_count" db:"sample_count"`
}

// LoadBalancingDecision represents an automated load balancing decision
type LoadBalancingDecision struct {
	ID                  string            `json:"id" db:"id"`
	FacilityID          string            `json:"facility_id" db:"facility_id"`
	DecisionType        DecisionType      `json:"decision_type" db:"decision_type"`
	SourceZoneID        string            `json:"source_zone_id" db:"source_zone_id"`
	TargetZoneID        string            `json:"target_zone_id" db:"target_zone_id"`
	LoadShiftAmount     decimal.Decimal   `json:"load_shift_amount" db:"load_shift_amount"`     // kW
	Reason              string            `json:"reason" db:"reason"`
	TriggeringCondition string            `json:"triggering_condition" db:"triggering_condition"`
	Priority            int               `json:"priority" db:"priority"`
	Status              DecisionStatus    `json:"status" db:"status"`
	ScheduledAt         time.Time         `json:"scheduled_at" db:"scheduled_at"`
	ExecutedAt          *time.Time        `json:"executed_at,omitempty" db:"executed_at"`
	CancelledAt         *time.Time        `json:"cancelled_at,omitempty" db:"cancelled_at"`
	Result              string            `json:"result,omitempty" db:"result"`
	CreatedAt           time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at" db:"updated_at"`
}

// DecisionType represents the type of load balancing decision
type DecisionType string

const (
	DecisionTypeShiftLoad   DecisionType = "shift_load"
	DecisionTypeScaleUp     DecisionType = "scale_up"
	DecisionTypeScaleDown   DecisionType = "scale_down"
	DecisionTypeReschedule  DecisionType = "reschedule"
	DecisionTypeEmergency   DecisionType = "emergency"
	DecisionTypeOptimize    DecisionType = "optimize"
)

// DecisionStatus represents the status of a load balancing decision
type DecisionStatus string

const (
	DecisionStatusPending   DecisionStatus = "pending"
	DecisionStatusApproved  DecisionStatus = "approved"
	DecisionStatusExecuting DecisionStatus = "executing"
	DecisionStatusCompleted DecisionStatus = "completed"
	DecisionStatusCancelled DecisionStatus = "cancelled"
	DecisionStatusFailed    DecisionStatus = "failed"
)

// EnergyFacility represents a facility being monitored for energy optimization
type EnergyFacility struct {
	ID              string          `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Location        string          `json:"location" db:"location"`
	Timezone        string          `json:"timezone" db:"timezone"`
	TotalCapacity   decimal.Decimal `json:"total_capacity" db:"total_capacity"`     // kW
	CurrentLoad     decimal.Decimal `json:"current_load" db:"current_load"`         // kW
	UtilizationRate decimal.Decimal `json:"utilization_rate" db:"utilization_rate"` // percentage
	EnergySources   []EnergySource  `json:"energy_sources" db:"-"`
	CarbonBudgetID  string          `json:"carbon_budget_id,omitempty" db:"carbon_budget_id"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// EnergyZone represents a zone within a facility for granular load management
type EnergyZone struct {
	ID              string          `json:"id" db:"id"`
	FacilityID      string          `json:"facility_id" db:"facility_id"`
	Name            string          `json:"name" db:"name"`
	ZoneType        ZoneType        `json:"zone_type" db:"zone_type"`
	Capacity        decimal.Decimal `json:"capacity" db:"capacity"`           // kW
	CurrentLoad     decimal.Decimal `json:"current_load" db:"current_load"`   // kW
	Priority        int             `json:"priority" db:"priority"`
	Can ShedLoad    bool            `json:"can_shed_load" db:"can_shed_load"`
	MinLoad         decimal.Decimal `json:"min_load" db:"min_load"`           // kW
	MaxLoad         decimal.Decimal `json:"max_load" db:"max_load"`           // kW
	ScheduledEvents []ScheduledEvent `json:"scheduled_events,omitempty" db:"-"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// ZoneType represents the type of energy zone
type ZoneType string

const (
	ZoneTypeProduction    ZoneType = "production"
	ZoneTypeOffice        ZoneType = "office"
	ZoneTypeHVAC          ZoneType = "hvac"
	ZoneTypeLighting      ZoneType = "lighting"
	ZoneTypeDataCenter    ZoneType = "data_center"
	ZoneTypeManufacturing ZoneType = "manufacturing"
	ZoneTypeStorage       ZoneType = "storage"
	ZoneTypeGeneral       ZoneType = "general"
)

// ScheduledEvent represents a scheduled event that affects energy consumption
type ScheduledEvent struct {
	ID              string          `json:"id" db:"id"`
	ZoneID          string          `json:"zone_id" db:"zone_id"`
	EventType       EventType       `json:"event_type" db:"event_type"`
	Description     string          `json:"description" db:"description"`
	ScheduledStart  time.Time       `json:"scheduled_start" db:"scheduled_start"`
	ScheduledEnd    time.Time       `json:"scheduled_end" db:"scheduled_end"`
	EstimatedImpact decimal.Decimal `json:"estimated_impact" db:"estimated_impact"` // kW
	IsRecurring     bool            `json:"is_recurring" db:"is_recurring"`
	RecurrenceRule  string          `json:"recurrence_rule,omitempty" db:"recurrence_rule"`
	Status          EventStatus     `json:"status" db:"status"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// EventType represents the type of scheduled event
type EventType string

const (
	EventTypeMaintenance       EventType = "maintenance"
	EventTypeProductionRun     EventType = "production_run"
	EventTypeEquipmentStart    EventType = "equipment_start"
	EventTypeEquipmentStop     EventType = "equipment_stop"
	EventTypeLoadShift         EventType = "load_shift"
	EventTypePeakShaving       EventType = "peak_shaving"
	EventTypeDemandResponse    EventType = "demand_response"
)

// EventStatus represents the status of a scheduled event
type EventStatus string

const (
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusActive    EventStatus = "active"
	EventStatusCompleted EventStatus = "completed"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusDelayed   EventStatus = "delayed"
)

// EnergyOptimizationReport represents a generated optimization report
type EnergyOptimizationReport struct {
	ID                string          `json:"id" db:"id"`
	FacilityID        string          `json:"facility_id" db:"facility_id"`
	ReportType        ReportType      `json:"report_type" db:"report_type"`
	PeriodStart       time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time       `json:"period_end" db:"period_end"`
	TotalConsumption  decimal.Decimal `json:"total_consumption" db:"total_consumption"`   // kWh
	PeakDemand        decimal.Decimal `json:"peak_demand" db:"peak_demand"`               // kW
	TotalCarbonEmission decimal.Decimal `json:"total_carbon_emission" db:"total_carbon_emission"` // kg CO2
	CarbonSavings     decimal.Decimal `json:"carbon_savings" db:"carbon_savings"`         // kg CO2
	EnergySavings     decimal.Decimal `json:"energy_savings" db:"energy_savings"`         // kWh
	CostSavings       decimal.Decimal `json:"cost_savings" db:"cost_savings"`
	OptimizationsApplied int          `json:"optimizations_applied" db:"optimizations_applied"`
	Recommendations   []Recommendation `json:"recommendations" db:"-"`
	GeneratedAt       time.Time       `json:"generated_at" db:"generated_at"`
}

// ReportType represents the type of optimization report
type ReportType string

const (
	ReportTypeDaily     ReportType = "daily"
	ReportTypeWeekly    ReportType = "weekly"
	ReportTypeMonthly   ReportType = "monthly"
	ReportTypeQuarterly ReportType = "quarterly"
	ReportTypeAnnual    ReportType = "annual"
	ReportTypeCustom    ReportType = "custom"
)

// Recommendation represents an optimization recommendation
type Recommendation struct {
	ID              string          `json:"id" db:"id"`
	ReportID        string          `json:"report_id" db:"report_id"`
	Category        string          `json:"category" db:"category"`
	Title           string          `json:"title" db:"title"`
	Description     string          `json:"description" db:"description"`
	EstimatedSavings decimal.Decimal `json:"estimated_savings" db:"estimated_savings"` // currency
	Priority        int             `json:"priority" db:"priority"`
	Implementation  string          `json:"implementation,omitempty" db:"implementation"`
	Status          RecStatus       `json:"status" db:"status"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// RecStatus represents the status of a recommendation
type RecStatus string

const (
	RecStatusPending    RecStatus = "pending"
	RecStatusApproved   RecStatus = "approved"
	RecStatusImplemented RecStatus = "implemented"
	RecStatusDismissed  RecStatus = "dismissed"
)

// JSONMap is a generic type for JSON metadata
type JSONMap map[string]interface{}
