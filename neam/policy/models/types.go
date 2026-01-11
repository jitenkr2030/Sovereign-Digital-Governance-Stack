package models

import (
	"encoding/json"
	"time"
)

// PolicyDefinition represents a configurable policy rule
type PolicyDefinition struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Category        PolicyCategory  `json:"category"`
	Priority        int             `json:"priority"` // 1-5, 5 being highest
	Rules           PolicyRules     `json:"rules"`
	TargetScope     TargetScope     `json:"target_scope"`
	Actions         []PolicyAction  `json:"actions"`
	CooldownMinutes int             `json:"cooldown_minutes"`
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CreatedBy       string          `json:"created_by"`
	Version         int             `json:"version"`
}

// PolicyCategory categorizes policies by domain
type PolicyCategory string

const (
	CategoryInflation       PolicyCategory = "INFLATION"
	CategoryEmployment      PolicyCategory = "EMPLOYMENT"
	CategoryAgriculture     PolicyCategory = "AGRICULTURE"
	CategoryManufacturing   PolicyCategory = "MANUFACTURING"
	CategoryTrade           PolicyCategory = "TRADE"
	CategoryEnergy          PolicyCategory = "ENERGY"
	CategoryFinancial       PolicyCategory = "FINANCIAL"
	CategoryEmergency       PolicyCategory = "EMERGENCY"
)

// TargetScope defines the applicability of a policy
type TargetScope struct {
	Regions    []string `json:"regions"`    // Empty means all regions
	Sectors    []string `json:"sectors"`    // Empty means all sectors
	Entities   []string `json:"entities"`   // Specific districts/organizations
	Conditions []Condition `json:"conditions"` // Additional filters
}

// PolicyRules contains the rule logic for policy evaluation
type PolicyRules struct {
	LogicExpression string         `json:"logic_expression"` // e.g., "IF stress > 80 AND inflation > 5%"
	Conditions      []RuleCondition `json:"conditions"`
	TriggerType     TriggerType    `json:"trigger_type"`
	ThresholdValues map[string]float64 `json:"threshold_values"`
	TimeWindow      int            `json:"time_window_minutes"`
	SustainedPeriod int            `json:"sustained_period_minutes"` // How long condition must hold
}

// RuleCondition defines a single condition for policy evaluation
type RuleCondition struct {
	Metric      string      `json:"metric"`
	Operator    Operator    `json:"operator"`
	Value       interface{} `json:"value"`
	Aggregation Aggregation `json:"aggregation"` // SUM, AVG, MAX, MIN
}

// Operator defines comparison operators
type Operator string

const (
	OpGreaterThan    Operator = "GT"
	OpLessThan       Operator = "LT"
	OpEqual          Operator = "EQ"
	OpGreaterOrEqual Operator = "GTE"
	OpLessOrEqual    Operator = "LTE"
	OpIn             Operator = "IN"
	OpBetween        Operator = "BETWEEN"
)

// Aggregation defines how to aggregate metrics
type Aggregation string

const (
	AggSum   Aggregation = "SUM"
	AggAvg   Aggregation = "AVG"
	AggMax   Aggregation = "MAX"
	AggMin   Aggregation = "MIN"
	AggCount Aggregation = "COUNT"
	AggLast  Aggregation = "LAST"
)

// TriggerType defines how the policy is triggered
type TriggerType string

const (
	TriggerImmediate   TriggerType = "IMMEDIATE"
	TriggerSustained   TriggerType = "SUSTAINED" // Condition must hold for period
	TriggerScheduled   TriggerType = "SCHEDULED"
	TriggerManual      TriggerType = "MANUAL"
)

// PolicyAction defines actions to take when policy triggers
type PolicyAction struct {
	Type        ActionType    `json:"type"`
	Parameters  ActionParams  `json:"parameters"`
	Priority    int           `json:"priority"` // Order of execution
	TimeoutSecs int           `json:"timeout_secs"`
}

// ActionType defines types of policy actions
type ActionType string

const (
	ActionNotify           ActionType = "NOTIFY"
	ActionAlert            ActionType = "ALERT"
	ActionWorkflow         ActionType = "WORKFLOW"
	ActionStateChange      ActionType = "STATE_CHANGE"
	ActionRestriction      ActionType = "RESTRICTION"
	ActionSubsidy          ActionType = "SUBSIDY"
	ActionExport           ActionType = "EXPORT"
	ActionSimulation       ActionType = "SIMULATION"
)

// ActionParams contains parameters for policy actions
type ActionParams struct {
	Recipients    []string          `json:"recipients"` // Ministry IDs, email addresses
	Channels      []NotificationChannel `json:"channels"` // EMAIL, SMS, WEBHOOK, DASHBOARD
	Message       string            `json:"message"`
	WorkflowID    string            `json:"workflow_id"`
	TargetState   string            `json:"target_state"`
	Restrictions  []Restriction     `json:"restrictions"`
	SubsidyAmount float64           `json:"subsidy_amount"`
	SubsidyUnit   string            `json:"subsidy_unit"`
	TemplateID    string            `json:"template_id"`
}

// NotificationChannel defines how to send notifications
type NotificationChannel string

const (
	ChannelEmail     NotificationChannel = "EMAIL"
	ChannelSMS       NotificationChannel = "SMS"
	ChannelWebhook   NotificationChannel = "WEBHOOK"
	ChannelDashboard NotificationChannel = "DASHBOARD"
	ChannelTelegram  NotificationChannel = "TELEGRAM"
)

// Restriction defines a temporary control/restriction
type Restriction struct {
	Type        RestrictionType `json:"type"`
	Description string          `json:"description"`
	DurationMin int             `json:"duration_minutes"`
	Scope       string          `json:"scope"`
	MaxValue    float64         `json:"max_value,omitempty"`
}

// RestrictionType defines types of restrictions
type RestrictionType string

const (
	RestrictionPriceCap    RestrictionType = "PRICE_CAP"
	RestrictionExportLimit RestrictionType = "EXPORT_LIMIT"
	RestrictionImportQuota RestrictionType = "IMPORT_QUOTA"
	RestrictionMovement    RestrictionType = "MOVEMENT"
	RestrictionFinancial   RestrictionType = "FINANCIAL"
)

// Intervention represents an active or historical intervention
type Intervention struct {
	ID            string            `json:"id"`
	PolicyID      string            `json:"policy_id"`
	PolicyName    string            `json:"policy_name"`
	WorkflowID    string            `json:"workflow_id"`
	Status        InterventionStatus `json:"status"`
	Scope         TargetScope       `json:"scope"`
	TriggerEvent  TriggerEvent      `json:"trigger_event"`
	ActionsTaken  []ActionLog       `json:"actions_taken"`
	Outcomes      InterventionOutcome `json:"outcomes"`
	Timeline      []TimelineEvent   `json:"timeline"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	CompletedAt   *time.Time        `json:"completed_at,omitempty"`
	CreatedBy     string            `json:"created_by"`
	ApprovalNotes string            `json:"approval_notes,omitempty"`
}

// InterventionStatus represents the status of an intervention
type InterventionStatus string

const (
	InterventionPending   InterventionStatus = "PENDING"
	InterventionApproved  InterventionStatus = "APPROVED"
	InterventionExecuting InterventionStatus = "EXECUTING"
	InterventionCompleted InterventionStatus = "COMPLETED"
	InterventionCancelled InterventionStatus = "CANCELLED"
	InterventionFailed    InterventionStatus = "FAILED"
)

// TriggerEvent describes what triggered the intervention
type TriggerEvent struct {
	Condition    string                 `json:"condition"`
	MetricValue  float64                `json:"metric_value"`
	Threshold    float64                `json:"threshold"`
	SustainedFor int                    `json:"sustained_for_minutes"`
	Evidence     map[string]interface{} `json:"evidence"`
}

// ActionLog records actions taken during intervention
type ActionLog struct {
	ActionType   ActionType    `json:"action_type"`
	Timestamp    time.Time     `json:"timestamp"`
	Status       string        `json:"status"`
	Details      string        `json:"details"`
	Result       interface{}   `json:"result,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// InterventionOutcome tracks the results of an intervention
type InterventionOutcome struct {
	PreInterventionMetrics map[string]float64 `json:"pre_intervention_metrics"`
	PostInterventionMetrics map[string]float64 `json:"post_intervention_metrics"`
	TargetAchieved         bool               `json:"target_achieved"`
	ImpactScore            float64            `json:"impact_score"` // 0-1
	SideEffects            []string           `json:"side_effects"`
	CostIncurred           float64            `json:"cost_incurred"`
	Beneficiaries          int                `json:"beneficiaries"`
}

// TimelineEvent represents an event in the intervention timeline
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"event_type"`
	Actor     string    `json:"actor"`
	Details   string    `json:"details"`
}

// Simulation represents an economic simulation
type Simulation struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	CreatedBy     string           `json:"created_by"`
	BaseScenario  BaseScenario     `json:"base_scenario"`
	Interventions []SimulationInput `json:"interventions"`
	Results       *SimulationResult `json:"results,omitempty"`
	Status        SimulationStatus `json:"status"`
	CreatedAt     time.Time        `json:"created_at"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
}

// SimulationStatus represents the status of a simulation
type SimulationStatus string

const (
	SimulationPending    SimulationStatus = "PENDING"
	SimulationRunning    SimulationStatus = "RUNNING"
	SimulationCompleted  SimulationStatus = "COMPLETED"
	SimulationFailed     SimulationStatus = "FAILED"
)

// BaseScenario represents the baseline economic scenario
type BaseScenario struct {
	RegionID      string                  `json:"region_id"`
	SectorID      string                  `json:"sector_id"`
	StartDate     time.Time               `json:"start_date"`
	EndDate       time.Time               `json:"end_date"`
	InitialState  map[string]float64      `json:"initial_state"`
	Assumptions   map[string]interface{}  `json:"assumptions"`
}

// SimulationInput represents an intervention to simulate
type SimulationInput struct {
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	StartDate   time.Time              `json:"start_date"`
	Duration    int                    `json:"duration_days"`
}

// SimulationResult contains simulation output
type SimulationResult struct {
	ProjectedMetrics  map[string][]TimeSeriesPoint `json:"projected_metrics"`
	BaselineMetrics   map[string][]TimeSeriesPoint `json:"baseline_metrics"`
	ComparisonMetrics map[string]ComparisonMetric  `json:"comparison_metrics"`
	Summary           SimulationSummary            `json:"summary"`
	ConfidenceLevel   float64                      `json:"confidence_level"`
	RiskFactors       []RiskFactor                 `json:"risk_factors"`
}

// TimeSeriesPoint represents a data point in time
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// ComparisonMetric compares simulated vs baseline values
type ComparisonMetric struct {
	BaselineValue  float64 `json:"baseline_value"`
	SimulatedValue float64 `json:"simulated_value"`
	Delta          float64 `json:"delta"`
	DeltaPercent   float64 `json:"delta_percent"`
	Impact         string  `json:"impact"` // POSITIVE, NEGATIVE, NEUTRAL
}

// SimulationSummary provides a high-level summary
type SimulationSummary struct {
	OverallImpact  string  `json:"overall_impact"` // POSITIVE, NEGATIVE, MIXED
	PrimaryBenefit string  `json:"primary_benefit"`
	PrimaryCost    string  `json:"primary_cost"`
	NetPresentValue float64 `json:"net_present_value"`
	ROI            float64 `json:"roi_percent"`
}

// RiskFactor represents a risk identified in simulation
type RiskFactor struct {
	Description   string  `json:"description"`
	Probability   float64 `json:"probability"` // 0-1
	Severity      string  `json:"severity"`    // LOW, MEDIUM, HIGH
	Mitigation    string  `json:"mitigation"`
}

// Condition represents a filter condition
type Condition struct {
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// ToJSON converts a policy to JSON bytes
func (p *PolicyDefinition) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON parses JSON bytes into a policy
func (p *PolicyDefinition) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// ToJSON converts an intervention to JSON bytes
func (i *Intervention) ToJSON() ([]byte, error) {
	return json.Marshal(i)
}

// FromJSON parses JSON bytes into an intervention
func (i *Intervention) FromJSON(data []byte) error {
	return json.Unmarshal(data, i)
}

// ToJSON converts a simulation to JSON bytes
func (s *Simulation) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON parses JSON bytes into a simulation
func (s *Simulation) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
