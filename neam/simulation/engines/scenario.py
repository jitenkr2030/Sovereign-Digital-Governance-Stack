"""
NEAM Simulation Service - Scenario Engine
Policy scenario simulation and comparison engine
"""

import logging
import time
from datetime import datetime, timedelta
from typing import Any, Callable, Dict, List, Optional, Tuple
import numpy as np
import pandas as pd

from .base import BaseSimulationEngine, ProgressTracker
from .economic_model import EconomicModelEngine
from ..models.simulation import (
    PolicyScenarioRequest,
    PolicyScenarioResponse,
    ScenarioInput,
    PolicyIntervention,
    ScenarioComparisonResult,
    SimulationStatus,
    InterventionType,
    SectorType,
)


logger = logging.getLogger(__name__)


class ScenarioEngine(BaseSimulationEngine):
    """
    Policy scenario simulation engine for comparing economic outcomes
    under different policy interventions and external conditions.
    """
    
    def __init__(
        self,
        max_workers: int = 4,
        default_timeout: float = 600.0,
        random_seed: Optional[int] = None
    ):
        """
        Initialize the scenario engine
        
        Args:
            max_workers: Maximum number of worker threads
            default_timeout: Default timeout for simulations
            random_seed: Random seed for reproducibility
        """
        super().__init__(
            max_workers=max_workers,
            default_timeout=default_timeout,
            random_seed=random_seed
        )
        
        self.economic_engine = EconomicModelEngine(
            max_workers=2,
            random_seed=random_seed
        )
    
    def _execute_simulation(
        self,
        request: PolicyScenarioRequest,
        progress_callback: Optional[Callable[[float], None]] = None
    ) -> PolicyScenarioResponse:
        """
        Execute policy scenario simulation
        
        Args:
            request: Policy scenario parameters
            progress_callback: Optional progress callback
            
        Returns:
            Policy scenario simulation results
        """
        start_time = time.time()
        created_at = datetime.utcnow()
        
        # Validate request
        validation_errors = self.validate_request(request)
        if validation_errors:
            raise ValueError(f"Validation errors: {', '.join(validation_errors)}")
        
        logger.info(
            f"Running scenario simulation: baseline + {len(request.scenarios)} scenarios"
        )
        
        # Create progress tracker
        total_steps = len(request.scenarios) + 1  # +1 for baseline
        tracker = ProgressTracker(
            total_steps=total_steps,
            callback=progress_callback,
            update_frequency=1
        )
        
        # Simulate baseline
        logger.info("Simulating baseline scenario")
        baseline_results = self._simulate_scenario(
            request.baseline,
            request.indicators,
            request.projection_months
        )
        tracker.update()
        
        # Simulate each alternative scenario
        scenario_results: Dict[str, Dict[str, List[float]]] = {}
        comparisons: Dict[str, ScenarioComparisonResult] = {}
        
        for scenario in request.scenarios:
            logger.info(f"Simulating scenario: {scenario.name}")
            
            scenario_results[scenario.name] = self._simulate_scenario(
                scenario,
                request.indicators,
                request.projection_months
            )
            
            # Compare with baseline if requested
            if request.compare_with_baseline:
                comparisons[scenario.name] = self._compare_scenarios(
                    scenario.name,
                    baseline_results,
                    scenario_results[scenario.name],
                    scenario,
                    request.indicators
                )
            
            tracker.update()
        
        tracker.complete()
        
        # Generate recommendations
        recommendations = self._generate_recommendations(
            baseline_results,
            scenario_results,
            comparisons,
            request
        )
        
        duration_ms = (time.time() - start_time) * 1000
        
        return PolicyScenarioResponse(
            simulation_id=request.simulation_id or self._generate_simulation_id(),
            simulation_type=request.simulation_type,
            status=SimulationStatus.COMPLETED,
            created_at=created_at,
            completed_at=datetime.utcnow(),
            duration_ms=duration_ms,
            regions=request.regions,
            sectors=request.sectors,
            result={
                "summary": f"Scenario comparison completed: baseline + {len(request.scenarios)} scenarios",
                "scenarios_compared": len(request.scenarios) + 1
            },
            metadata={
                "projection_months": request.projection_months,
                "compare_with_baseline": request.compare_with_baseline
            },
            baseline_results=baseline_results,
            scenario_results=scenario_results,
            comparisons=comparisons,
            recommendations=recommendations
        )
    
    def _simulate_scenario(
        self,
        scenario: ScenarioInput,
        indicators: List[str],
        projection_months: int
    ) -> Dict[str, List[float]]:
        """
        Simulate a single scenario
        
        Args:
            scenario: Scenario definition
            indicators: Indicators to track
            projection_months: Projection period
            
        Returns:
            Projected values for each indicator
        """
        results: Dict[str, List[float]] = {}
        
        for indicator_code in indicators:
            base_value = scenario.base_conditions.get(indicator_code, 100.0)
            
            # Generate base projection
            projection = self._generate_base_projection(
                base_value,
                projection_months
            )
            
            # Apply intervention effects
            for intervention in scenario.interventions:
                effect = self._calculate_intervention_effect(
                    intervention,
                    indicator_code,
                    projection_months
                )
                projection = self._apply_effect(projection, effect)
            
            # Apply external shocks
            if scenario.external_shocks:
                for shock in scenario.external_shocks:
                    shock_effect = self._calculate_shock_effect(
                        shock,
                        indicator_code,
                        projection_months
                    )
                    projection = self._apply_effect(projection, shock_effect)
            
            results[indicator_code] = projection
        
        return results
    
    def _generate_base_projection(
        self,
        base_value: float,
        projection_months: int
    ) -> List[float]:
        """
        Generate a base projection without interventions
        
        Args:
            base_value: Starting value
            projection_months: Number of months to project
            
        Returns:
            List of projected values
        """
        # Assume moderate growth with some volatility
        monthly_growth = 0.005  # ~6% annual
        volatility = 0.01
        
        values = [base_value]
        
        for month in range(1, projection_months + 1):
            growth = monthly_growth + np.random.normal(0, volatility)
            new_value = values[-1] * (1 + growth)
            values.append(new_value)
        
        return values[1:]  # Exclude initial value
    
    def _calculate_intervention_effect(
        self,
        intervention: PolicyIntervention,
        indicator_code: str,
        projection_months: int
    ) -> List[float]:
        """
        Calculate the effect of a policy intervention
        
        Args:
            intervention: Policy intervention definition
            indicator_code: Target indicator
            projection_months: Projection period
            
        Returns:
            Effect on the indicator over time
        """
        effect = np.zeros(projection_months)
        
        # Check if intervention targets this indicator
        if not self._intervention_targets_indicator(intervention, indicator_code):
            return effect.tolist()
        
        # Calculate effect based on intervention type
        if intervention.intervention_type == InterventionType.FISCAL_STIMULUS:
            # Direct impact on GDP and related indicators
            effect = self._calculate_fiscal_stimulus_effect(
                intervention, projection_months
            )
            
        elif intervention.intervention_type == InterventionType.INTEREST_RATE_ADJUSTMENT:
            # Impact on inflation, investment, etc.
            effect = self._calculate_interest_rate_effect(
                intervention, projection_months
            )
            
        elif intervention.intervention_type == InterventionType.TAX_CHANGE:
            # Impact on consumption, investment
            effect = self._calculate_tax_change_effect(
                intervention, projection_months
            )
            
        elif intervention.intervention_type == InterventionType.INFRASTRUCTURE_INVESTMENT:
            # Long-term impact on GDP and employment
            effect = self._calculate_infrastructure_effect(
                intervention, projection_months
            )
            
        elif intervention.intervention_type == InterventionType.MINIMUM_WAGE_CHANGE:
            # Impact on employment, prices
            effect = self._calculate_minimum_wage_effect(
                intervention, projection_months
            )
            
        else:
            # Generic effect
            effect = self._calculate_generic_effect(
                intervention, projection_months
            )
        
        # Apply elasticity
        if intervention.elasticity is not None:
            effect = effect * intervention.elasticity
        
        # Apply decay if specified
        if intervention.decay_rate is not None:
            effect = self._apply_decay(effect, intervention.decay_rate)
        
        return effect.tolist()
    
    def _intervention_targets_indicator(
        self,
        intervention: PolicyIntervention,
        indicator_code: str
    ) -> bool:
        """Check if intervention targets a specific indicator"""
        # Mapping of intervention types to typical target indicators
        target_mapping = {
            InterventionType.FISCAL_STIMULUS: ['gdp', 'gdp_growth', 'employment'],
            InterventionType.INTEREST_RATE_ADJUSTMENT: ['inflation', 'gdp_growth', 'exchange_rate'],
            InterventionType.TAX_CHANGE: ['gdp', 'consumption', 'investment'],
            InterventionType.TRADE_POLICY: ['trade_balance', 'gdp', 'exchange_rate'],
            InterventionType.REGULATORY_CHANGE: ['gdp', 'employment', 'productivity'],
            InterventionType.SUBSIDY_PROGRAM: ['gdp', 'employment', 'specific_sector'],
            InterventionType.INFRASTRUCTURE_INVESTMENT: ['gdp', 'employment', 'productivity'],
            InterventionType.MINIMUM_WAGE_CHANGE: ['wages', 'employment', 'inflation'],
        }
        
        targets = target_mapping.get(intervention.intervention_type, [])
        return indicator_code in targets
    
    def _calculate_fiscal_stimulus_effect(
        self,
        intervention: PolicyIntervention,
        projection_months: int
    ) -> np.ndarray:
        """Calculate fiscal stimulus effect"""
        effect = np.zeros(projection_months)
        
        # Effect builds up over time with lag
        lag_months = intervention.lag_months
        magnitude = intervention.magnitude / 100  # Convert percentage to decimal
        
        # GDP multiplier effect
        for month in range(lag_months, projection_months):
            months_active = month - lag_months + 1
            # Effect increases then stabilizes
            if months_active <= 6:
                effect[month] = magnitude * (months_active / 6)
            else:
                effect[month] = magnitude * 0.9  # Slight decay
        
        return effect
    
    def _calculate_interest_rate_effect(
        self,
        intervention: PolicyIntervention,
        projection_months: int
    ) -> np.ndarray:
        """Calculate interest rate adjustment effect"""
        effect = np.zeros(projection_months)
        
        # Interest rate changes have delayed effects
        rate_change = -intervention.magnitude / 100  # Negative = rate cut
        
        # Effect on GDP growth (approximate relationship)
        for month in range(3, projection_months):  # 3-month lag
            months_since = month - 3
            # Effect builds up then stabilizes
            effect[month] = rate_change * 0.5 * min(months_since / 12, 1.0)
        
        return effect
    
    def _calculate_tax_change_effect(
        self,
        intervention: PolicyIntervention,
        projection_months: int
    ) -> np.ndarray:
        """Calculate tax change effect"""
        effect = np.zeros(projection_months)
        
        tax_change = -intervention.magnitude / 100  # Negative = tax cut
        lag_months = intervention.lag_months
        
        # Tax changes affect consumption with some lag
        for month in range(lag_months, projection_months):
            months_active = month - lag_months + 1
            # Marginal propensity to consume effect
            effect[month] = tax_change * 0.7 * min(months_active / 6, 1.0)
        
        return effect
    
    def _calculate_infrastructure_effect(
        self,
        intervention: PolicyIntervention,
        projection_months: int
    ) -> np.ndarray:
        """Calculate infrastructure investment effect"""
        effect = np.zeros(projection_months)
        
        magnitude = intervention.magnitude / 100
        lag_months = intervention.lag_months + 6  # Long lead time
        
        # Infrastructure effects are long-term
        for month in range(lag_months, projection_months):
            months_active = month - lag_months + 1
            # Effect grows over time as projects complete
            effect[month] = magnitude * 0.1 * min(months_active / 24, 1.0)
        
        return effect
    
    def _calculate_minimum_wage_effect(
        self,
        intervention: PolicyIntervention,
        projection_months: int
    ) -> np.ndarray:
        """Calculate minimum wage change effect"""
        effect = np.zeros(projection_months)
        
        wage_change = intervention.magnitude / 100
        
        # Minimum wage increases affect employment negatively
        # but wages positively
        for month in range(3, projection_months):
            effect[month] = -wage_change * 0.15  # Small negative employment effect
        
        return effect
    
    def _calculate_generic_effect(
        self,
        intervention: PolicyIntervention,
        projection_months: int
    ) -> np.ndarray:
        """Calculate generic intervention effect"""
        effect = np.zeros(projection_months)
        
        magnitude = intervention.magnitude / 100
        lag_months = intervention.lag_months
        
        for month in range(lag_months, projection_months):
            effect[month] = magnitude
        
        return effect
    
    def _calculate_shock_effect(
        self,
        shock: Dict[str, Any],
        indicator_code: str,
        projection_months: int
    ) -> np.ndarray:
        """Calculate external shock effect"""
        effect = np.zeros(projection_months)
        
        shock_magnitude = shock.get('magnitude', 0) / 100
        shock_timing = shock.get('timing_month', 0)
        shock_duration = shock.get('duration_months', 3)
        
        # Apply shock at specified timing
        for month in range(shock_timing, min(shock_timing + shock_duration, projection_months)):
            effect[month] = shock_magnitude
        
        # Recovery
        recovery_rate = shock.get('recovery_rate', 0.5)
        for month in range(shock_timing + shock_duration, projection_months):
            remaining_effect = effect[month - 1] * (1 - recovery_rate)
            effect[month] = remaining_effect
        
        return effect
    
    def _apply_effect(
        self,
        base_projection: List[float],
        effect: List[float]
    ) -> List[float]:
        """
        Apply effect to base projection
        
        Args:
            base_projection: Base projection values
            effect: Effect to apply (as percentage changes)
            
        Returns:
            Modified projection
        """
        if len(effect) == 0:
            return base_projection
        
        # Extend effect if needed
        if len(effect) < len(base_projection):
            effect = effect + [effect[-1]] * (len(base_projection) - len(effect))
        
        # Apply effect multiplicatively
        modified = [
            base * (1 + e) 
            for base, e in zip(base_projection, effect[:len(base_projection)])
        ]
        
        return modified
    
    def _apply_decay(
        self,
        effect: np.ndarray,
        decay_rate: float
    ) -> np.ndarray:
        """Apply decay to effect over time"""
        for month in range(1, len(effect)):
            effect[month] = effect[month - 1] * (1 - decay_rate) + effect[month] * decay_rate
        return effect
    
    def _compare_scenarios(
        self,
        scenario_name: str,
        baseline: Dict[str, List[float]],
        scenario: Dict[str, List[float]],
        scenario_input: ScenarioInput,
        indicators: List[str]
    ) -> ScenarioComparisonResult:
        """
        Compare a scenario against the baseline
        
        Args:
            scenario_name: Name of the scenario
            baseline: Baseline projections
            scenario: Scenario projections
            scenario_input: Scenario input definition
            indicators: List of indicators
            
        Returns:
            Comparison results
        """
        vs_baseline: Dict[str, Dict[str, float]] = {}
        cumulative_effect: Dict[str, float] = {}
        peak_effect: Dict[str, Dict[str, Any]] = {}
        timing_of_effects: Dict[str, datetime] = {}
        
        for indicator in indicators:
            baseline_values = baseline.get(indicator, [])
            scenario_values = scenario.get(indicator, [])
            
            if not baseline_values or not scenario_values:
                continue
            
            # Calculate difference at each time point
            differences = [
                s - b for s, b in zip(scenario_values, baseline_values)
            ]
            
            vs_baseline[indicator] = {
                'month_1': differences[0] if len(differences) > 0 else 0,
                'month_6': differences[5] if len(differences) > 5 else differences[-1],
                'month_12': differences[11] if len(differences) > 11 else differences[-1],
                'final': differences[-1]
            }
            
            # Cumulative effect (sum of differences)
            cumulative_effect[indicator] = sum(differences)
            
            # Peak effect
            if differences:
                max_idx = np.argmax(differences)
                min_idx = np.argmin(differences)
                
                peak_effect[indicator] = {
                    'max_change': max(differences),
                    'min_change': min(differences),
                    'max_month': max_idx + 1,
                    'min_month': min_idx + 1
                }
                
                timing_of_effects[indicator] = datetime.utcnow() + timedelta(days=30 * max_idx)
        
        # Calculate cost effectiveness
        total_cost = sum(
            abs(i.magnitude) 
            for i in scenario_input.interventions 
            if hasattr(i, 'magnitude')
        )
        
        cost_effectiveness = {}
        if total_cost > 0:
            for indicator, cumulative in cumulative_effect.items():
                cost_effectiveness[indicator] = cumulative / total_cost
        
        return ScenarioComparisonResult(
            scenario_name=scenario_name,
            vs_baseline=vs_baseline,
            cumulative_effect=cumulative_effect,
            peak_effect=peak_effect,
            timing_of_effects=timing_of_effects,
            cost_effectiveness=cost_effectiveness if cost_effectiveness else None,
            side_effects=self._identify_side_effects(scenario_input, indicators)
        )
    
    def _identify_side_effects(
        self,
        scenario: ScenarioInput,
        indicators: List[str]
    ) -> List[Dict[str, Any]]:
        """Identify potential side effects of interventions"""
        side_effects = []
        
        for intervention in scenario.interventions:
            if intervention.intervention_type == InterventionType.FISCAL_STIMULUS:
                side_effects.append({
                    'type': 'debt_increase',
                    'description': 'Increased government debt',
                    'severity': 'medium',
                    'affected_indicators': ['debt_to_gdp', 'deficit']
                })
            
            elif intervention.intervention_type == InterventionType.INTEREST_RATE_ADJUSTMENT:
                if intervention.magnitude < 0:
                    side_effects.append({
                        'type': 'inflation_risk',
                        'description': 'Potential for increased inflation',
                        'severity': 'medium',
                        'affected_indicators': ['inflation', 'currency_value']
                    })
        
        return side_effects
    
    def _generate_recommendations(
        self,
        baseline_results: Dict[str, List[float]],
        scenario_results: Dict[str, Dict[str, List[float]]],
        comparisons: Dict[str, ScenarioComparisonResult],
        request: PolicyScenarioRequest
    ) -> List[Dict[str, Any]]:
        """Generate policy recommendations based on simulation results"""
        recommendations = []
        
        for scenario_name, comparison in comparisons.items():
            scenario = next(
                (s for s in request.scenarios if s.name == scenario_name),
                None
            )
            
            if not scenario:
                continue
            
            # Find indicators with best outcomes
            best_indicators = sorted(
                comparison.cumulative_effect.items(),
                key=lambda x: x[1],
                reverse=True
            )[:3]
            
            for indicator, effect in best_indicators:
                if effect > 0:
                    recommendations.append({
                        'scenario': scenario_name,
                        'indicator': indicator,
                        'effect': effect,
                        'recommendation': f"Scenario '{scenario_name}' shows {effect:.2f} improvement in {indicator}",
                        'confidence': 'medium',
                        'interventions': [i.name for i in scenario.interventions]
                    })
        
        # Add general recommendations
        recommendations.append({
            'scenario': 'general',
            'indicator': 'overall',
            'effect': 0,
            'recommendation': 'Consider running Monte Carlo simulations to assess uncertainty',
            'confidence': 'high',
            'interventions': []
        })
        
        return recommendations
    
    def _create_failed_response(
        self,
        request: PolicyScenarioRequest,
        error_message: str,
        duration_ms: float
    ) -> PolicyScenarioResponse:
        """Create a failed policy scenario response"""
        return PolicyScenarioResponse(
            simulation_id=request.simulation_id or "failed_simulation",
            simulation_type=request.simulation_type,
            status=SimulationStatus.FAILED,
            created_at=datetime.utcnow(),
            completed_at=datetime.utcnow(),
            duration_ms=duration_ms,
            regions=request.regions,
            sectors=request.sectors,
            result={"error": error_message},
            metadata={"error_type": "execution_failure"},
            baseline_results={},
            scenario_results={},
            comparisons={},
            recommendations=[]
        )
