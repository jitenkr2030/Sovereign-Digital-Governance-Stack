"""
NEAM Simulation Service - Data Models
Pydantic models for simulation requests and responses
"""

from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Union
from pydantic import BaseModel, Field, field_validator
import numpy as np


class SimulationType(str, Enum):
    """Types of economic simulations available"""
    MONTE_CARLO = "monte_carlo"
    ECONOMIC_PROJECTION = "economic_projection"
    POLICY_SCENARIO = "policy_scenario"
    SENSITIVITY_ANALYSIS = "sensitivity_analysis"
    STRESS_TEST = "stress_test"
    SCENARIO_COMPARISON = "scenario_comparison"


class RegionType(str, Enum):
    """Geographic region classifications"""
    NATIONAL = "national"
    STATE = "state"
    COUNTY = "county"
    METRO_AREA = "metro_area"
    CUSTOM = "custom"


class SectorType(str, Enum):
    """Economic sector classifications"""
    MANUFACTURING = "manufacturing"
    SERVICES = "services"
    AGRICULTURE = "agriculture"
    TECHNOLOGY = "technology"
    RETAIL = "retail"
    CONSTRUCTION = "construction"
    HEALTHCARE = "healthcare"
    FINANCE = "finance"
    ENERGY = "energy"
    TRANSPORTATION = "transportation"


class DistributionType(str, Enum):
    """Probability distribution types for Monte Carlo simulations"""
    NORMAL = "normal"
    UNIFORM = "uniform"
    LOGNORMAL = "lognormal"
    TRIANGULAR = "triangular"
    EXPONENTIAL = "exponential"
    POISSON = "poisson"
    BETA = "beta"
    GAMMA = "gamma"


class InterventionType(str, Enum):
    """Types of policy interventions"""
    FISCAL_STIMULUS = "fiscal_stimulus"
    TAX_CHANGE = "tax_change"
    INTEREST_RATE_ADJUSTMENT = "interest_rate_adjustment"
    TRADE_POLICY = "trade_policy"
    REGULATORY_CHANGE = "regulatory_change"
    SUBSIDY_PROGRAM = "subsidy_program"
    INFRASTRUCTURE_INVESTMENT = "infrastructure_investment"
    MINIMUM_WAGE_CHANGE = "minimum_wage_change"


class SimulationStatus(str, Enum):
    """Status of a simulation request"""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"
    TIMEOUT = "timeout"


# Base Models
class BaseSimulationRequest(BaseModel):
    """Base class for simulation requests"""
    simulation_id: Optional[str] = Field(None, description="Unique identifier for the simulation")
    simulation_type: SimulationType = Field(..., description="Type of simulation to run")
    description: Optional[str] = Field(None, max_length=500, description="Human-readable description")
    regions: List[str] = Field(default_factory=list, description="List of regions to simulate")
    sectors: List[SectorType] = Field(default_factory=list, description="Economic sectors to include")
    start_date: datetime = Field(..., description="Simulation start date")
    end_date: datetime = Field(..., description="Simulation end date")
    metadata: Optional[Dict[str, Any]] = Field(None, description="Additional metadata")
    
    @field_validator('end_date')
    @classmethod
    def validate_end_date(cls, v: datetime, info) -> datetime:
        if 'start_date' in info.data and v <= info.data['start_date']:
            raise ValueError('end_date must be after start_date')
        return v


class BaseSimulationResponse(BaseModel):
    """Base class for simulation responses"""
    simulation_id: str = Field(..., description="Unique identifier for the simulation")
    simulation_type: SimulationType = Field(..., description="Type of simulation that was run")
    status: SimulationStatus = Field(..., description="Final status of the simulation")
    created_at: datetime = Field(..., description="When the simulation was created")
    completed_at: Optional[datetime] = Field(None, description="When the simulation completed")
    duration_ms: Optional[float] = Field(None, description="Simulation execution time in milliseconds")
    regions: List[str] = Field(..., description="Regions that were simulated")
    sectors: List[SectorType] = Field(..., description="Sectors that were simulated")
    result: Dict[str, Any] = Field(..., description="Simulation results")
    metadata: Optional[Dict[str, Any]] = Field(None, description="Additional result metadata")


# Monte Carlo Models
class DistributionParameter(BaseModel):
    """Parameters for a probability distribution"""
    distribution_type: DistributionType = Field(..., description="Type of distribution")
    mean: Optional[float] = Field(None, description="Mean value (for normal, lognormal, etc.)")
    std_dev: Optional[float] = Field(None, description="Standard deviation (for normal, etc.)")
    min_value: Optional[float] = Field(None, description="Minimum value (for uniform, triangular)")
    max_value: Optional[float] = Field(None, description="Maximum value (for uniform, triangular)")
    mode: Optional[float] = Field(None, description="Mode value (for triangular)")
    rate: Optional[float] = Field(None, description="Rate parameter (for exponential, poisson)")
    alpha: Optional[float] = Field(None, description="Alpha parameter (for beta, gamma)")
    beta_param: Optional[float] = Field(None, description="Beta parameter (for beta, gamma)")
    
    @field_validator('mean', 'std_dev', 'min_value', 'max_value', 'mode', 'rate', 'alpha', 'beta_param')
    @classmethod
    def validate_parameters(cls, v: Optional[float]) -> Optional[float]:
        if v is not None and not np.isfinite(v):
            raise ValueError('Distribution parameter must be a finite number')
        return v


class VariableInput(BaseModel):
    """Input variable for Monte Carlo simulation"""
    name: str = Field(..., description="Variable name (e.g., 'inflation_rate', 'gdp_growth')")
    description: Optional[str] = Field(None, description="Description of the variable")
    distribution: DistributionParameter = Field(..., description="Probability distribution parameters")
    correlation_matrix: Optional[Dict[str, float]] = Field(None, description="Correlations with other variables")
    unit: Optional[str] = Field(None, description="Unit of measurement")
    base_value: Optional[float] = Field(None, description="Historical base value for reference")


class MonteCarloRequest(BaseSimulationRequest):
    """Request for Monte Carlo simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.MONTE_CARLO, 
        description="Type of simulation"
    )
    variables: List[VariableInput] = Field(
        ..., 
        min_length=1,
        max_length=50,
        description="Input variables with their probability distributions"
    )
    output_variables: List[str] = Field(
        ...,
        min_length=1,
        max_length=100,
        description="Names of output variables to calculate"
    )
    iterations: int = Field(
        default=10000,
        ge=100,
        le=100000,
        description="Number of simulation iterations"
    )
    confidence_level: float = Field(
        default=0.95,
        ge=0.5,
        le=0.999,
        description="Confidence level for confidence intervals"
    )
    random_seed: Optional[int] = Field(
        None,
        description="Random seed for reproducibility"
    )
    output_formula: Optional[str] = Field(
        None,
        description="Formula for calculating output variables (Python expression)"
    )


class MonteCarloResult(BaseModel):
    """Results from Monte Carlo simulation"""
    variable_name: str = Field(..., description="Name of the variable")
    mean: float = Field(..., description="Mean value across all iterations")
    std_dev: float = Field(..., description="Standard deviation across all iterations")
    median: float = Field(..., description="Median value")
    min_value: float = Field(..., description="Minimum value observed")
    max_value: float = Field(..., description="Maximum value observed")
    percentiles: Dict[str, float] = Field(
        default_factory=dict,
        description="Percentile values (p5, p25, p50, p75, p95, etc.)"
    )
    confidence_interval: Dict[str, float] = Field(
        default_factory=dict,
        description="Lower and upper bounds of confidence interval"
    )
    distribution_fit: Optional[Dict[str, Any]] = Field(
        None,
        description="Best-fit distribution parameters if applicable"
    )
    histogram: Optional[Dict[str, List[float]]] = Field(
        None,
        description="Histogram data for visualization"
    )


class MonteCarloResponse(BaseSimulationResponse):
    """Response for Monte Carlo simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.MONTE_CARLO,
        description="Type of simulation"
    )
    iterations_completed: int = Field(..., description="Number of iterations executed")
    input_results: Dict[str, MonteCarloResult] = Field(
        ..., 
        description="Results for each input variable"
    )
    output_results: Dict[str, MonteCarloResult] = Field(
        ..., 
        description="Results for each output variable"
    )
    correlation_matrix: Optional[List[List[float]]] = Field(
        None,
        description="Correlation matrix of all variables"
    )
    convergence_metrics: Optional[Dict[str, float]] = Field(
        None,
        description="Metrics indicating simulation convergence"
    )


# Economic Projection Models
class EconomicIndicator(BaseModel):
    """Economic indicator for projection"""
    name: str = Field(..., description="Indicator name (e.g., 'GDP', 'CPI', 'unemployment')")
    code: str = Field(..., description="Standardized indicator code")
    value: float = Field(..., description="Current/historical value")
    unit: str = Field(..., description="Unit of measurement")
    growth_rate: Optional[float] = Field(None, description="Historical growth rate")
    volatility: Optional[float] = Field(None, description="Historical volatility (std dev)")


class EconomicProjectionRequest(BaseSimulationRequest):
    """Request for economic projection simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.ECONOMIC_PROJECTION,
        description="Type of simulation"
    )
    indicators: List[EconomicIndicator] = Field(
        ...,
        min_length=1,
        max_length=50,
        description="Economic indicators to project"
    )
    projection_months: int = Field(
        default=24,
        ge=1,
        le=120,
        description="Number of months to project into the future"
    )
    model_type: str = Field(
        default="arima",
        description="Forecasting model type (arima, exponential_smoothing, prophet, linear)"
    )
    seasonality: bool = Field(
        default=True,
        description="Whether to account for seasonal patterns"
    )
    trend_type: str = Field(
        default="auto",
        description="Trend type (linear, exponential, damped, auto)"
    )
    external_factors: Optional[Dict[str, List[float]]] = Field(
        None,
        description="External factors to incorporate (e.g., policy changes, shocks)"
    )


class EconomicProjectionResult(BaseModel):
    """Result of economic projection"""
    indicator_code: str = Field(..., description="Indicator code")
    indicator_name: str = Field(..., description="Indicator name")
    projections: List[Dict[str, Any]] = Field(
        ...,
        description="Projected values over time"
    )
    confidence_intervals: List[Dict[str, Any]] = Field(
        ...,
        description="Confidence intervals for each projection point"
    )
    expected_growth: float = Field(..., description="Expected total growth over projection period")
    growth_ci_lower: float = Field(..., description="Lower bound of growth confidence interval")
    growth_ci_upper: float = Field(..., description="Upper bound of growth confidence interval")
    peak_value: float = Field(..., description="Peak projected value")
    peak_date: datetime = Field(..., description="Date of peak value")
    trough_value: float = Field(..., description="Trough projected value")
    trough_date: datetime = Field(..., description="Date of trough value")
    model_parameters: Optional[Dict[str, Any]] = Field(
        None,
        description="Fitted model parameters"
    )
    model_fit_metrics: Optional[Dict[str, float]] = Field(
        None,
        description="Metrics indicating model fit quality"
    )


class EconomicProjectionResponse(BaseSimulationResponse):
    """Response for economic projection simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.ECONOMIC_PROJECTION,
        description="Type of simulation"
    )
    projection_months: int = Field(..., description="Number of months projected")
    model_type: str = Field(..., description="Forecasting model used")
    results: Dict[str, EconomicProjectionResult] = Field(
        ..., 
        description="Projection results for each indicator"
    )
    scenario_assumptions: Dict[str, Any] = Field(
        ...,
        description="Assumptions made in the projection"
    )
    risk_factors: List[Dict[str, Any]] = Field(
        default_factory=list,
        description="Identified risk factors"
    )


# Policy Scenario Models
class PolicyIntervention(BaseModel):
    """Policy intervention to simulate"""
    intervention_type: InterventionType = Field(..., description="Type of intervention")
    name: str = Field(..., description="Name of the intervention")
    description: Optional[str] = Field(None, description="Description of the intervention")
    start_date: datetime = Field(..., description="When the intervention takes effect")
    end_date: Optional[datetime] = Field(None, description="When the intervention ends (if temporary)")
    magnitude: float = Field(..., description="Magnitude of the intervention")
    unit: str = Field(..., description="Unit of measurement (percent, dollars, basis points)")
    target_sectors: List[SectorType] = Field(
        default_factory=list,
        description="Sectors targeted by the intervention"
    )
    target_regions: List[str] = Field(
        default_factory=list,
        description="Regions targeted by the intervention"
    )
    parameters: Optional[Dict[str, Any]] = Field(
        None,
        description="Additional intervention-specific parameters"
    )
    elasticity: Optional[float] = Field(
        None,
        ge=0.0,
        description="Elasticity of the intervention effect"
    )
    lag_months: int = Field(
        default=0,
        ge=0,
        description="Number of months before the intervention takes effect"
    )
    decay_rate: Optional[float] = Field(
        None,
        ge=0.0,
        le=1.0,
        description="Rate at which intervention effect decays over time"
    )


class ScenarioInput(BaseModel):
    """Input for scenario comparison simulation"""
    name: str = Field(..., description="Scenario name")
    description: Optional[str] = Field(None, description="Scenario description")
    base_conditions: Dict[str, float] = Field(
        ...,
        description="Starting economic conditions"
    )
    interventions: List[PolicyIntervention] = Field(
        default_factory=list,
        description="Policy interventions in this scenario"
    )
    external_shocks: Optional[List[Dict[str, Any]]] = Field(
        None,
        description="External economic shocks to simulate"
    )


class PolicyScenarioRequest(BaseSimulationRequest):
    """Request for policy scenario simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.POLICY_SCENARIO,
        description="Type of simulation"
    )
    baseline: ScenarioInput = Field(..., description="Baseline scenario (no interventions)")
    scenarios: List[ScenarioInput] = Field(
        ...,
        min_length=1,
        max_length=20,
        description="Alternative scenarios to compare"
    )
    indicators: List[str] = Field(
        ...,
        min_length=1,
        max_length=50,
        description="Economic indicators to track"
    )
    projection_months: int = Field(
        default=24,
        ge=1,
        le=120,
        description="Number of months to simulate"
    )
    compare_with_baseline: bool = Field(
        default=True,
        description="Whether to compare scenarios with baseline"
    )


class ScenarioComparisonResult(BaseModel):
    """Result of comparing one scenario against baseline"""
    scenario_name: str = Field(..., description="Name of the scenario")
    vs_baseline: Dict[str, Dict[str, float]] = Field(
        ...,
        description="Difference from baseline for each indicator at each time point"
    )
    cumulative_effect: Dict[str, float] = Field(
        ...,
        description="Cumulative effect over the simulation period"
    )
    peak_effect: Dict[str, Dict[str, Any]] = Field(
        ...,
        description="Maximum effect on each indicator"
    )
    timing_of_effects: Dict[str, datetime] = Field(
        ...,
        description="When effects peak for each indicator"
    )
    cost_effectiveness: Optional[Dict[str, float]] = Field(
        None,
        description="Cost per unit of effect achieved"
    )
    side_effects: List[Dict[str, Any]] = Field(
        default_factory=list,
        description="Unintended consequences"
    )


class PolicyScenarioResponse(BaseSimulationResponse):
    """Response for policy scenario simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.POLICY_SCENARIO,
        description="Type of simulation"
    )
    baseline_results: Dict[str, List[float]] = Field(
        ...,
        description="Baseline projections for each indicator"
    )
    scenario_results: Dict[str, Dict[str, List[float]]] = Field(
        ...,
        description="Results for each scenario by indicator"
    )
    comparisons: Dict[str, ScenarioComparisonResult] = Field(
        ...,
        description="Comparison results for each scenario"
    )
    recommendations: List[Dict[str, Any]] = Field(
        default_factory=list,
        description="Policy recommendations based on results"
    )


# Sensitivity Analysis Models
class ParameterRange(BaseModel):
    """Range of values for a parameter in sensitivity analysis"""
    name: str = Field(..., description="Parameter name")
    base_value: float = Field(..., description="Baseline parameter value")
    min_value: float = Field(..., description="Minimum value to test")
    max_value: float = Field(..., description="Maximum value to test")
    step_type: str = Field(
        default="linear",
        description="Step type (linear, log, custom)"
    )
    num_steps: int = Field(
        default=10,
        ge=2,
        le=100,
        description="Number of steps between min and max"
    )


class SensitivityAnalysisRequest(BaseSimulationRequest):
    """Request for sensitivity analysis"""
    simulation_type: SimulationType = Field(
        default=SimulationType.SENSITIVITY_ANALYSIS,
        description="Type of simulation"
    )
    output_metric: str = Field(
        ...,
        description="Output metric to analyze sensitivity for"
    )
    parameters: List[ParameterRange] = Field(
        ...,
        min_length=1,
        max_length=20,
        description="Parameters to vary in the analysis"
    )
    method: str = Field(
        default="sobol",
        description="Sensitivity analysis method (sobol, morris, fast, elementary)"
    )
    sample_size: int = Field(
        default=1000,
        ge=100,
        le=10000,
        description="Number of samples for analysis"
    )
    model_script: Optional[str] = Field(
        None,
        description="Python script defining the model to analyze (if not using predefined)"
    )


class SensitivityAnalysis(BaseModel):
    """Results of sensitivity analysis"""
    output_metric: str = Field(..., description="Output metric analyzed")
    method: str = Field(..., description="Method used for analysis")
    parameter_sensitivities: Dict[str, Dict[str, float]] = Field(
        ...,
        description="Sensitivity indices for each parameter"
    )
    total_order_indices: Dict[str, float] = Field(
        ...,
        description="Total-order sensitivity indices"
    )
    first_order_indices: Dict[str, float] = Field(
        ...,
        description="First-order sensitivity indices"
    )
    interaction_indices: Dict[str, Dict[str, float]] = Field(
        default_factory=dict,
        description="Interaction effects between parameters"
    )
    parameter_rankings: List[Dict[str, Any]] = Field(
        ...,
        description="Parameters ranked by sensitivity"
    )
    confidence_intervals: Dict[str, Dict[str, float]] = Field(
        ...,
        description="Confidence intervals for sensitivity indices"
    )


class SensitivityAnalysisResponse(BaseSimulationResponse):
    """Response for sensitivity analysis"""
    simulation_type: SimulationType = Field(
        default=SimulationType.SENSITIVITY_ANALYSIS,
        description="Type of simulation"
    )
    analysis: SensitivityAnalysis = Field(..., description="Sensitivity analysis results")
    parameter_interactions: Optional[Dict[str, Any]] = Field(
        None,
        description="Visualization data for parameter interactions"
    )
    recommended_ranges: Dict[str, Dict[str, float]] = Field(
        ...,
        description="Recommended parameter ranges for desired output"
    )


# Stress Test Models
class StressScenario(BaseModel):
    """Stress test scenario definition"""
    name: str = Field(..., description="Scenario name")
    description: Optional[str] = Field(None, description="Scenario description")
    severity: str = Field(..., description="Severity level (mild, moderate, severe, extreme)")
    shock_parameters: Dict[str, float] = Field(
        ...,
        description="Shock parameters for each indicator"
    )
    duration_months: int = Field(
        default=12,
        ge=1,
        description="Duration of the stress period"
    )
    recovery_pattern: str = Field(
        default="v",
        description="Recovery pattern (v, u, l, w)"
    )


class StressTestRequest(BaseSimulationRequest):
    """Request for stress test simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.STRESS_TEST,
        description="Type of simulation"
    )
    scenarios: List[StressScenario] = Field(
        ...,
        min_length=1,
        max_length=10,
        description="Stress scenarios to test"
    )
    indicators: List[str] = Field(
        ...,
        min_length=1,
        max_length=50,
        description="Economic indicators to test"
    )
    baseline_conditions: Dict[str, float] = Field(
        ...,
        description="Baseline economic conditions"
    )
    test_resilience: bool = Field(
        default=True,
        description="Whether to test economic resilience"
    )


class StressTestResult(BaseModel):
    """Result of stress test"""
    scenario_name: str = Field(..., description="Name of the stress scenario")
    severity: str = Field(..., description="Severity level tested")
    indicator_impacts: Dict[str, Dict[str, float]] = Field(
        ...,
        description="Impact on each indicator"
    )
    time_to_recovery: Optional[Dict[str, int]] = Field(
        None,
        description="Time periods to recover for each indicator"
    )
    peak_impact: Dict[str, Dict[str, Any]] = Field(
        ...,
        description="Peak impact on each indicator"
    )
    systemic_risk_score: float = Field(
        ...,
        description="Overall systemic risk score (0-1)"
    )
    vulnerable_sectors: List[str] = Field(
        default_factory=list,
        description="Sectors most vulnerable in this scenario"
    )
    vulnerable_regions: List[str] = Field(
        default_factory=list,
        description="Regions most vulnerable in this scenario"
    )
    recommendations: List[str] = Field(
        default_factory=list,
        description="Mitigation recommendations"
    )


class StressTestResponse(BaseSimulationResponse):
    """Response for stress test simulation"""
    simulation_type: SimulationType = Field(
        default=SimulationType.STRESS_TEST,
        description="Type of simulation"
    )
    results: Dict[str, StressTestResult] = Field(
        ...,
        description="Results for each stress scenario"
    )
    comparison_summary: Dict[str, Any] = Field(
        ...,
        description="Summary comparison of all scenarios"
    )
    resilience_score: float = Field(
        ...,
        description="Overall economic resilience score (0-1)"
    )


# Scenario Comparison Models
class ScenarioComparisonRequest(BaseSimulationRequest):
    """Request for comparing multiple scenarios"""
    simulation_type: SimulationType = Field(
        default=SimulationType.SCENARIO_COMPARISON,
        description="Type of simulation"
    )
    scenarios: List[ScenarioInput] = Field(
        ...,
        min_length=2,
        max_length=10,
        description="Scenarios to compare"
    )
    metrics: List[str] = Field(
        ...,
        min_length=1,
        max_length=20,
        description="Metrics to compare across scenarios"
    )
    weights: Optional[Dict[str, float]] = Field(
        None,
        description="Weights for multi-criteria decision analysis"
    )
    projection_months: int = Field(
        default=24,
        ge=1,
        le=120,
        description="Projection period in months"
    )


# Utility Models
class SimulationJobStatus(BaseModel):
    """Status of an async simulation job"""
    simulation_id: str = Field(..., description="Simulation identifier")
    status: SimulationStatus = Field(..., description="Current status")
    progress_percent: float = Field(
        default=0.0,
        ge=0.0,
        le=100.0,
        description="Progress percentage"
    )
    message: Optional[str] = Field(None, description="Status message")
    estimated_completion: Optional[datetime] = Field(None, description="Estimated completion time")
    created_at: datetime = Field(..., description="When the job was created")
    started_at: Optional[datetime] = Field(None, description="When execution started")
    error: Optional[str] = Field(None, description="Error message if failed")


class BatchSimulationRequest(BaseModel):
    """Request to run multiple simulations in batch"""
    simulations: List[BaseSimulationRequest] = Field(
        ...,
        min_length=1,
        max_length=10,
        description="Simulations to run"
    )
    parallel: bool = Field(
        default=True,
        description="Whether to run simulations in parallel"
    )
    priority: int = Field(
        default=0,
        ge=0,
        le=10,
        description="Priority (higher = more urgent)"
    )


class BatchSimulationResponse(BaseModel):
    """Response for batch simulation request"""
    batch_id: str = Field(..., description="Batch identifier")
    total_simulations: int = Field(..., description="Total simulations in batch")
    submitted_at: datetime = Field(..., description="When the batch was submitted")
    simulation_ids: List[str] = Field(..., description="Individual simulation IDs")
    estimated_duration_ms: float = Field(
        ...,
        description="Estimated total duration"
    )
    status: SimulationStatus = Field(..., description="Overall batch status")
