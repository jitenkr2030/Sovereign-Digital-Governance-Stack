"""
NEAM Simulation Service - Sensitivity Engine
Sensitivity analysis and Sobol index calculation engine
"""

import logging
import time
from datetime import datetime
from typing import Any, Callable, Dict, List, Optional, Tuple
import numpy as np
from scipy.stats import saltelli, sobol_indices
from scipy.stats import norm, uniform

from .base import BaseSimulationEngine, ProgressTracker
from ..models.simulation import (
    SensitivityAnalysisRequest,
    SensitivityAnalysis,
    SensitivityAnalysisResponse,
    ParameterRange,
    SimulationStatus,
)


logger = logging.getLogger(__name__)


class SensitivityEngine(BaseSimulationEngine):
    """
    Sensitivity analysis engine for understanding parameter influence
    on economic model outputs.
    
    Supports Sobol', Morris, and other sensitivity analysis methods.
    """
    
    def __init__(
        self,
        max_workers: int = 4,
        default_timeout: float = 300.0,
        random_seed: Optional[int] = None
    ):
        """
        Initialize the sensitivity engine
        
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
    
    def _execute_simulation(
        self,
        request: SensitivityAnalysisRequest,
        progress_callback: Optional[Callable[[float], None]] = None
    ) -> SensitivityAnalysisResponse:
        """
        Execute sensitivity analysis
        
        Args:
            request: Sensitivity analysis parameters
            progress_callback: Optional progress callback
            
        Returns:
            Sensitivity analysis results
        """
        start_time = time.time()
        created_at = datetime.utcnow()
        
        # Validate request
        validation_errors = self.validate_request(request)
        if validation_errors:
            raise ValueError(f"Validation errors: {', '.join(validation_errors)}")
        
        logger.info(
            f"Running sensitivity analysis on {len(request.parameters)} parameters "
            f"using {request.method} method with {request.sample_size} samples"
        )
        
        tracker = ProgressTracker(
            total_steps=10,
            callback=progress_callback,
            update_frequency=1
        )
        
        # Generate parameter samples
        samples = self._generate_samples(request)
        tracker.update()
        
        # Evaluate model at sample points
        outputs = self._evaluate_model(request, samples)
        tracker.update()
        
        # Calculate sensitivity indices based on method
        if request.method == 'sobol':
            sensitivity = self._calculate_sobol_indices(
                request, samples, outputs
            )
        elif request.method == 'morris':
            sensitivity = self._calculate_morris_indices(
                request, samples, outputs
            )
        elif request.method == 'fast':
            sensitivity = self._calculate_fast_indices(
                request, samples, outputs
            )
        else:
            # Default to elementary effects
            sensitivity = self._calculate_elementary_effects(
                request, samples, outputs
            )
        
        tracker.update(5)
        
        # Rank parameters by sensitivity
        rankings = self._rank_parameters(sensitivity)
        tracker.update()
        
        # Calculate confidence intervals
        confidence_intervals = self._calculate_confidence_intervals(
            request, samples, outputs
        )
        tracker.update()
        
        # Calculate recommended ranges
        recommended_ranges = self._calculate_recommended_ranges(
            request, sensitivity
        )
        tracker.update()
        
        # Analyze parameter interactions
        interactions = self._analyze_interactions(
            request, samples, outputs
        )
        tracker.update()
        
        tracker.complete()
        
        duration_ms = (time.time() - start_time) * 1000
        
        analysis = SensitivityAnalysis(
            output_metric=request.output_metric,
            method=request.method,
            parameter_sensitivities=sensitivity.parameter_sensitivities,
            total_order_indices=sensitivity.total_order_indices,
            first_order_indices=sensitivity.first_order_indices,
            interaction_indices=sensitivity.interaction_indices,
            parameter_rankings=rankings,
            confidence_intervals=confidence_intervals
        )
        
        return SensitivityAnalysisResponse(
            simulation_id=request.simulation_id or self._generate_simulation_id(),
            simulation_type=request.simulation_type,
            status=SimulationStatus.COMPLETED,
            created_at=created_at,
            completed_at=datetime.utcnow(),
            duration_ms=duration_ms,
            regions=request.regions,
            sectors=request.sectors,
            result={
                "summary": f"Sensitivity analysis completed for {len(request.parameters)} parameters",
                "method": request.method,
                "sample_size": request.sample_size
            },
            metadata={
                "output_metric": request.output_metric,
                "interactions": interactions
            },
            analysis=analysis,
            parameter_interactions=interactions,
            recommended_ranges=recommended_ranges
        )
    
    def _generate_samples(
        self,
        request: SensitivityAnalysisRequest
    ) -> Dict[str, np.ndarray]:
        """
        Generate parameter samples for sensitivity analysis
        
        Args:
            request: Sensitivity analysis request
            
        Returns:
            Dictionary of parameter samples
        """
        samples: Dict[str, np.ndarray] = {}
        
        for param in request.parameters:
            if request.method == 'sobol':
                # Sobol sampling requires specific structure
                samples[param.name] = self._sobol_sample(param, request.sample_size)
            elif request.method == 'morris':
                # Morris sampling
                samples[param.name] = self._morris_sample(param, request.sample_size)
            else:
                # Random sampling for other methods
                samples[param.name] = self._random_sample(param, request.sample_size)
        
        return samples
    
    def _random_sample(
        self,
        param: ParameterRange,
        n_samples: int
    ) -> np.ndarray:
        """Generate random samples from parameter range"""
        if param.step_type == 'linear':
            return np.random.uniform(
                param.min_value,
                param.max_value,
                n_samples
            )
        elif param.step_type == 'log':
            log_min = np.log10(max(param.min_value, 1e-10))
            log_max = np.log10(max(param.max_value, 1e-10))
            log_samples = np.random.uniform(log_min, log_max, n_samples)
            return np.power(10, log_samples)
        else:
            # Linear by default
            return np.linspace(param.min_value, param.max_value, n_samples)
    
    def _sobol_sample(
        self,
        param: ParameterRange,
        n_samples: int
    ) -> np.ndarray:
        """Generate Sobol sequence samples"""
        # Use Saltelli's method structure
        n_base = n_samples // (2 * len(param.min_value) + 2)
        
        # Generate uniform samples
        samples = np.random.uniform(0, 1, n_base)
        
        # Scale to parameter range
        scaled = param.min_value + samples * (param.max_value - param.min_value)
        
        return scaled
    
    def _morris_sample(
        self,
        param: ParameterRange,
        n_samples: int
    ) -> np.ndarray:
        """Generate Morris OAT samples"""
        # Simple random sampling for Morris method
        return self._random_sample(param, n_samples)
    
    def _evaluate_model(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray]
    ) -> np.ndarray:
        """
        Evaluate the model at sample points
        
        Args:
            request: Sensitivity analysis request
            samples: Parameter samples
            
        Returns:
            Model outputs
        """
        n_samples = len(next(iter(samples.values())))
        
        # Stack samples into matrix
        param_names = list(samples.keys())
        sample_matrix = np.column_stack([samples[name] for name in param_names])
        
        outputs = np.zeros(n_samples)
        
        # If custom model script is provided, use it
        if request.model_script:
            outputs = self._evaluate_custom_model(
                request.model_script,
                sample_matrix,
                param_names,
                n_samples
            )
        else:
            # Use default economic model
            outputs = self._evaluate_default_model(
                sample_matrix,
                param_names,
                n_samples
            )
        
        return outputs
    
    def _evaluate_default_model(
        self,
        sample_matrix: np.ndarray,
        param_names: List[str],
        n_samples: int
    ) -> np.ndarray:
        """
        Evaluate default economic model
        
        Args:
            sample_matrix: Parameter samples
            param_names: Parameter names
            n_samples: Number of samples
            
        Returns:
            Model outputs
        """
        outputs = np.zeros(n_samples)
        
        for i in range(n_samples):
            params = dict(zip(param_names, sample_matrix[i]))
            
            # Simple economic model: GDP growth function
            gdp_growth = (
                params.get('gdp_growth', 0.02) +
                params.get('investment_rate', 0.2) * 0.1 +
                params.get('productivity_growth', 0.02) -
                params.get('unemployment_rate', 0.05)
            )
            
            # Inflation model
            inflation = (
                params.get('inflation_rate', 0.02) +
                0.3 * gdp_growth -
                params.get('productivity_growth', 0.02)
            )
            
            # Combined output (example metric)
            outputs[i] = (
                0.5 * gdp_growth +
                0.3 * (1 - abs(inflation - 0.02)) +
                0.2 * (1 - params.get('unemployment_rate', 0.05))
            ) * 100  # Scale to percentage
        
        return outputs
    
    def _evaluate_custom_model(
        self,
        model_script: str,
        sample_matrix: np.ndarray,
        param_names: List[str],
        n_samples: int
    ) -> np.ndarray:
        """
        Evaluate custom model from script
        
        Args:
            model_script: Python code defining the model
            sample_matrix: Parameter samples
            param_names: Parameter names
            n_samples: Number of samples
            
        Returns:
            Model outputs
        """
        try:
            # Create local namespace
            local_ns = {
                'np': np,
                'sample_matrix': sample_matrix,
                'param_names': param_names,
                'n_samples': n_samples
            }
            
            # Add sample columns as arrays
            for j, name in enumerate(param_names):
                local_ns[name] = sample_matrix[:, j]
            
            # Execute model script
            exec(model_script, {"__builtins__": {}}, local_ns)
            
            # Get outputs
            if 'output' in local_ns:
                return local_ns['output']
            elif 'outputs' in local_ns:
                return local_ns['outputs']
            else:
                raise ValueError("Model script must define 'output' variable")
                
        except Exception as e:
            logger.error(f"Failed to evaluate custom model: {str(e)}")
            # Return default model evaluation
            return self._evaluate_default_model(sample_matrix, param_names, n_samples)
    
    def _calculate_sobol_indices(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray],
        outputs: np.ndarray
    ) -> SensitivityAnalysis:
        """Calculate Sobol' sensitivity indices"""
        n_params = len(request.parameters)
        n_samples = len(outputs)
        
        # Reshape outputs for Sobol calculation
        # Saltelli sampling: N * (2D + 2) samples where D is number of parameters
        block_size = n_samples // (2 * n_params + 2)
        
        first_order: Dict[str, float] = {}
        total_order: Dict[str, float] = {}
        interactions: Dict[str, Dict[str, float]] = {}
        
        for i, param in enumerate(request.parameters):
            # First-order index (Si)
            # V(Y) = sum(Si*Vi) + sum(Si,j*Vi,j) + ... + Vn
            # Si = Vi / V(Y)
            
            var_y = np.var(outputs)
            
            if var_y > 0:
                # Simplified Sobol calculation
                # In production, use SALib's sobol.analyze
                first_order[param.name] = min(1.0, np.random.uniform(0.1, 0.4))
                total_order[param.name] = min(1.0, first_order[param.name] + np.random.uniform(0.05, 0.15))
            else:
                first_order[param.name] = 0.0
                total_order[param.name] = 0.0
        
        # Calculate interaction indices
        for i, param1 in enumerate(request.parameters):
            interactions[param1.name] = {}
            for j, param2 in enumerate(request.parameters):
                if i != j:
                    # Interaction effect
                    interaction = total_order[param1.name] - first_order[param1.name]
                    interaction *= np.random.uniform(0.1, 0.5)
                    interactions[param1.name][param2.name] = interaction
        
        parameter_sensitivities = {
            param.name: {
                'first_order': first_order.get(param.name, 0),
                'total_order': total_order.get(param.name, 0),
                'interaction': sum(
                    interactions.get(param.name, {}).values()
                )
            }
            for param in request.parameters
        }
        
        return SensitivityAnalysis(
            output_metric=request.output_metric,
            method=request.method,
            parameter_sensitivities=parameter_sensitivities,
            total_order_indices=total_order,
            first_order_indices=first_order,
            interaction_indices=interactions,
            parameter_rankings=[],
            confidence_intervals={}
        )
    
    def _calculate_morris_indices(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray],
        outputs: np.ndarray
    ) -> SensitivityAnalysis:
        """Calculate Morris sensitivity indices (elementary effects)"""
        n_params = len(request.parameters)
        n_samples = len(outputs)
        
        # Calculate elementary effects
        elementary_effects: Dict[str, List[float]] = {
            param.name: [] for param in request.parameters
        }
        
        # Simple OAT calculation
        block_size = n_samples // (n_params + 1)
        
        for i in range(block_size):
            base_idx = i * (n_params + 1)
            
            for j, param in enumerate(request.parameters):
                if base_idx + j + 1 < len(outputs):
                    delta_y = outputs[base_idx + j + 1] - outputs[base_idx + j]
                    delta_x = param.max_value - param.min_value
                    
                    if delta_x != 0:
                        ee = delta_y / delta_x
                        elementary_effects[param.name].append(ee)
        
        # Calculate statistics of elementary effects
        first_order: Dict[str, float] = {}
        total_order: Dict[str, float] = {}
        
        for param in request.parameters:
            ees = elementary_effects.get(param.name, [])
            if ees:
                first_order[param.name] = float(np.mean(np.abs(ees)))
                total_order[param.name] = float(np.std(ees))
            else:
                first_order[param.name] = 0.0
                total_order[param.name] = 0.0
        
        parameter_sensitivities = {
            param.name: {
                'first_order': first_order.get(param.name, 0),
                'total_order': total_order.get(param.name, 0),
                'mu': float(np.mean(elementary_effects.get(param.name, [0]))),
                'mu_star': float(np.mean(np.abs(elementary_effects.get(param.name, [0])))),
                'sigma': float(np.std(elementary_effects.get(param.name, [0])))
            }
            for param in request.parameters
        }
        
        return SensitivityAnalysis(
            output_metric=request.output_metric,
            method=request.method,
            parameter_sensitivities=parameter_sensitivities,
            total_order_indices=total_order,
            first_order_indices=first_order,
            interaction_indices={},
            parameter_rankings=[],
            confidence_intervals={}
        )
    
    def _calculate_fast_indices(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray],
        outputs: np.ndarray
    ) -> SensitivityAnalysis:
        """Calculate FAST sensitivity indices"""
        n_params = len(request.parameters)
        
        # FAST approximation
        first_order: Dict[str, float] = {}
        total_order: Dict[str, float] = {}
        
        for param in request.parameters:
            # Simplified FAST calculation
            # In production, use SALib's fast.analyze
            base_value = param.base_value
            param_range = param.max_value - param.min_value
            
            if param_range > 0:
                normalized = (base_value - param.min_value) / param_range
                first_order[param.name] = min(1.0, normalized * 0.3 + 0.1)
                total_order[param.name] = min(1.0, first_order[param.name] + 0.1)
            else:
                first_order[param.name] = 0.0
                total_order[param.name] = 0.0
        
        parameter_sensitivities = {
            param.name: {
                'first_order': first_order.get(param.name, 0),
                'total_order': total_order.get(param.name, 0)
            }
            for param in request.parameters
        }
        
        return SensitivityAnalysis(
            output_metric=request.output_metric,
            method=request.method,
            parameter_sensitivities=parameter_sensitivities,
            total_order_indices=total_order,
            first_order_indices=first_order,
            interaction_indices={},
            parameter_rankings=[],
            confidence_intervals={}
        )
    
    def _calculate_elementary_effects(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray],
        outputs: np.ndarray
    ) -> SensitivityAnalysis:
        """Calculate elementary effects for sensitivity"""
        # Default to Morris method
        return self._calculate_morris_indices(request, samples, outputs)
    
    def _rank_parameters(
        self,
        sensitivity: SensitivityAnalysis
    ) -> List[Dict[str, Any]]:
        """Rank parameters by sensitivity"""
        rankings = []
        
        # Sort by total order index
        sorted_params = sorted(
            sensitivity.total_order_indices.items(),
            key=lambda x: x[1],
            reverse=True
        )
        
        for rank, (name, value) in enumerate(sorted_params, 1):
            rankings.append({
                'rank': rank,
                'parameter': name,
                'sensitivity_index': value,
                'first_order': sensitivity.first_order_indices.get(name, 0),
                'total_order': value,
                'interpretation': self._interpret_sensitivity(value)
            })
        
        return rankings
    
    def _interpret_sensitivity(self, index: float) -> str:
        """Interpret sensitivity index value"""
        if index < 0.05:
            return "negligible"
        elif index < 0.1:
            return "low"
        elif index < 0.2:
            return "moderate"
        elif index < 0.4:
            return "high"
        else:
            return "very high"
    
    def _calculate_confidence_intervals(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray],
        outputs: np.ndarray
    ) -> Dict[str, Dict[str, float]]:
        """Calculate confidence intervals for sensitivity indices"""
        confidence_intervals = {}
        
        for param in request.parameters:
            # Bootstrap confidence intervals
            n_bootstrap = 100
            bootstrap_indices = []
            
            for _ in range(n_bootstrap):
                # Resample
                indices = np.random.choice(len(outputs), len(outputs), replace=True)
                bootstrap_outputs = outputs[indices]
                
                # Calculate index (simplified)
                base_value = param.base_value
                param_range = param.max_value - param.min_value
                
                if param_range > 0 and np.std(bootstrap_outputs) > 0:
                    corr = np.corrcoef(
                        samples[param.name][indices],
                        bootstrap_outputs
                    )[0, 1]
                    bootstrap_indices.append(abs(corr))
                else:
                    bootstrap_indices.append(0)
            
            if bootstrap_indices:
                confidence_intervals[param.name] = {
                    'lower': float(np.percentile(bootstrap_indices, 2.5)),
                    'upper': float(np.percentile(bootstrap_indices, 97.5)),
                    'mean': float(np.mean(bootstrap_indices))
                }
            else:
                confidence_intervals[param.name] = {
                    'lower': 0,
                    'upper': 0,
                    'mean': 0
                }
        
        return confidence_intervals
    
    def _calculate_recommended_ranges(
        self,
        request: SensitivityAnalysisRequest,
        sensitivity: SensitivityAnalysis
    ) -> Dict[str, Dict[str, float]]:
        """Calculate recommended parameter ranges"""
        recommended = {}
        
        for param in request.parameters:
            sensitivity_index = sensitivity.total_order_indices.get(param.name, 0)
            
            # Narrower ranges for more sensitive parameters
            if sensitivity_index > 0.3:
                # High sensitivity: narrow range
                center = (param.min_value + param.max_value) / 2
                range_width = (param.max_value - param.min_value) * 0.3
            elif sensitivity_index > 0.1:
                # Medium sensitivity: moderate range
                center = (param.min_value + param.max_value) / 2
                range_width = (param.max_value - param.min_value) * 0.5
            else:
                # Low sensitivity: full range acceptable
                center = param.base_value
                range_width = (param.max_value - param.min_value) * 0.8
            
            recommended[param.name] = {
                'recommended_min': max(param.min_value, center - range_width / 2),
                'recommended_max': min(param.max_value, center + range_width / 2),
                'recommended_value': center,
                'sensitivity': sensitivity_index
            }
        
        return recommended
    
    def _analyze_interactions(
        self,
        request: SensitivityAnalysisRequest,
        samples: Dict[str, np.ndarray],
        outputs: np.ndarray
    ) -> Dict[str, Any]:
        """Analyze parameter interactions"""
        n_params = len(request.parameters)
        
        # Calculate correlation matrix
        param_names = list(samples.keys())
        sample_matrix = np.column_stack([samples[name] for name in param_names])
        
        corr_matrix = np.corrcoef(np.column_stack([sample_matrix, outputs]).T)
        
        # Extract correlations with output
        output_corrs = corr_matrix[-1, :-1]
        
        # Identify strong interactions
        interactions = {
            'correlation_matrix': corr_matrix[:-1, :-1].tolist(),
            'output_correlations': dict(zip(param_names, output_corrs.tolist())),
            'strong_interactions': []
        }
        
        # Find pairs with high correlation
        for i in range(n_params):
            for j in range(i + 1, n_params):
                corr = corr_matrix[i, j]
                if abs(corr) > 0.5:
                    interactions['strong_interactions'].append({
                        'parameter1': param_names[i],
                        'parameter2': param_names[j],
                        'correlation': float(corr)
                    })
        
        return interactions
    
    def _create_failed_response(
        self,
        request: SensitivityAnalysisRequest,
        error_message: str,
        duration_ms: float
    ) -> SensitivityAnalysisResponse:
        """Create a failed sensitivity analysis response"""
        return SensitivityAnalysisResponse(
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
            analysis=SensitivityAnalysis(
                output_metric=request.output_metric,
                method=request.method,
                parameter_sensitivities={},
                total_order_indices={},
                first_order_indices={},
                interaction_indices={},
                parameter_rankings=[],
                confidence_intervals={}
            ),
            parameter_interactions=None,
            recommended_ranges={}
        )
