"""
NEAM Simulation Service - Monte Carlo Engine
Monte Carlo simulation for risk assessment and probabilistic analysis
"""

import logging
import time
from datetime import datetime
from typing import Any, Callable, Dict, List, Optional, Tuple
import numpy as np
from scipy import stats
from scipy.stats import norm, uniform, lognorm, triang, expon, poisson, beta, gamma
import scipy.cluster.hierarchy as sch

from .base import BaseSimulationEngine, ProgressTracker
from ..models.simulation import (
    MonteCarloRequest,
    MonteCarloResult,
    MonteCarloResponse,
    SimulationStatus,
    VariableInput,
    DistributionParameter,
    DistributionType,
)


logger = logging.getLogger(__name__)


class MonteCarloEngine(BaseSimulationEngine):
    """
    Monte Carlo simulation engine for economic risk assessment.
    
    Supports various probability distributions and provides comprehensive
    statistical analysis of simulation results including confidence intervals,
    sensitivity analysis, and convergence diagnostics.
    """
    
    def __init__(
        self,
        max_workers: int = 4,
        default_timeout: float = 300.0,
        random_seed: Optional[int] = None
    ):
        """
        Initialize the Monte Carlo engine
        
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
        
        self._variable_cache: Dict[str, np.ndarray] = {}
    
    def _execute_simulation(
        self,
        request: MonteCarloRequest,
        progress_callback: Optional[Callable[[float], None]] = None
    ) -> MonteCarloResponse:
        """
        Execute Monte Carlo simulation
        
        Args:
            request: Monte Carlo simulation parameters
            progress_callback: Optional progress callback
            
        Returns:
            Monte Carlo simulation results
        """
        start_time = time.time()
        created_at = datetime.utcnow()
        
        # Set random seed if provided
        if request.random_seed is not None:
            self.set_random_seed(request.random_seed)
        
        # Validate request
        validation_errors = self.validate_request(request)
        if validation_errors:
            raise ValueError(f"Validation errors: {', '.join(validation_errors)}")
        
        # Generate samples for each variable
        logger.info(f"Generating {request.iterations} samples for {len(request.variables)} variables")
        
        variable_results: Dict[str, MonteCarloResult] = {}
        samples: Dict[str, np.ndarray] = {}
        
        # Create progress tracker
        tracker = ProgressTracker(
            total_steps=len(request.variables) + len(request.output_variables),
            callback=progress_callback,
            update_frequency=1
        )
        
        # Generate samples for input variables
        for var in request.variables:
            samples[var.name] = self._generate_samples(
                var.distribution,
                request.iterations
            )
            
            variable_results[var.name] = self._analyze_samples(
                var.name,
                samples[var.name],
                request.confidence_level
            )
            
            tracker.update()
        
        # Calculate output variables
        output_results: Dict[str, MonteCarloResult] = {}
        
        for output_name in request.output_variables:
            if request.output_formula:
                # Use the provided formula
                output_samples = self._evaluate_formula(
                    request.output_formula,
                    samples,
                    request.iterations
                )
            else:
                # Default: sum of all input variables
                output_samples = sum(samples.values())
            
            output_results[output_name] = self._analyze_samples(
                output_name,
                output_samples,
                request.confidence_level
            )
            
            tracker.update()
        
        # Calculate correlation matrix
        all_samples = {**samples}
        correlation_matrix = self._calculate_correlation_matrix(all_samples)
        
        # Calculate convergence metrics
        convergence_metrics = self._calculate_convergence_metrics(
            samples,
            request.iterations
        )
        
        # Generate histogram data for visualization
        histogram_data = self._generate_histogram_data(samples, output_samples)
        
        tracker.complete()
        
        duration_ms = (time.time() - start_time) * 1000
        
        return MonteCarloResponse(
            simulation_id=request.simulation_id or self._generate_simulation_id(),
            simulation_type=request.simulation_type,
            status=SimulationStatus.COMPLETED,
            created_at=created_at,
            completed_at=datetime.utcnow(),
            duration_ms=duration_ms,
            regions=request.regions,
            sectors=request.sectors,
            result={
                "summary": "Monte Carlo simulation completed successfully",
                "iterations": request.iterations,
                "confidence_level": request.confidence_level
            },
            metadata={
                "random_seed": request.random_seed,
                "histogram_data": histogram_data
            },
            iterations_completed=request.iterations,
            input_results=variable_results,
            output_results=output_results,
            correlation_matrix=correlation_matrix,
            convergence_metrics=convergence_metrics
        )
    
    def _generate_samples(
        self,
        distribution: DistributionParameter,
        n_samples: int
    ) -> np.ndarray:
        """
        Generate random samples from the specified distribution
        
        Args:
            distribution: Distribution parameters
            n_samples: Number of samples to generate
            
        Returns:
            Array of samples
        """
        dist_type = distribution.distribution_type
        
        if dist_type == DistributionType.NORMAL:
            if distribution.mean is None or distribution.std_dev is None:
                raise ValueError("Normal distribution requires mean and std_dev")
            samples = np.random.normal(
                distribution.mean,
                distribution.std_dev,
                n_samples
            )
            
        elif dist_type == DistributionType.UNIFORM:
            if distribution.min_value is None or distribution.max_value is None:
                raise ValueError("Uniform distribution requires min_value and max_value")
            samples = np.random.uniform(
                distribution.min_value,
                distribution.max_value,
                n_samples
            )
            
        elif dist_type == DistributionType.LOGNORMAL:
            if distribution.mean is None or distribution.std_dev is None:
                raise ValueError("Lognormal distribution requires mean and std_dev")
            # Convert to lognormal parameters
            sigma = np.sqrt(np.log(1 + (distribution.std_dev / distribution.mean) ** 2))
            mu = np.log(distribution.mean) - 0.5 * sigma ** 2
            samples = np.random.lognormal(mu, sigma, n_samples)
            
        elif dist_type == DistributionType.TRIANGULAR:
            if distribution.min_value is None or distribution.max_value is None or distribution.mode is None:
                raise ValueError("Triangular distribution requires min_value, max_value, and mode")
            # Convert to scipy triangular parameters
            c = (distribution.mode - distribution.min_value) / (distribution.max_value - distribution.min_value)
            samples = np.random.triangular(
                distribution.min_value,
                distribution.mode,
                distribution.max_value,
                n_samples
            )
            
        elif dist_type == DistributionType.EXPONENTIAL:
            if distribution.rate is None:
                raise ValueError("Exponential distribution requires rate parameter")
            samples = np.random.exponential(1.0 / distribution.rate, n_samples)
            
        elif dist_type == DistributionType.POISSON:
            if distribution.rate is None:
                raise ValueError("Poisson distribution requires rate parameter")
            samples = np.random.poisson(distribution.rate, n_samples)
            
        elif dist_type == DistributionType.BETA:
            if distribution.alpha is None or distribution.beta_param is None:
                raise ValueError("Beta distribution requires alpha and beta parameters")
            samples = np.random.beta(distribution.alpha, distribution.beta_param, n_samples)
            
        elif dist_type == DistributionType.GAMMA:
            if distribution.alpha is None or distribution.beta_param is None:
                raise ValueError("Gamma distribution requires alpha and beta parameters")
            samples = np.random.gamma(distribution.alpha, distribution.beta_param, n_samples)
            
        else:
            raise ValueError(f"Unsupported distribution type: {dist_type}")
        
        # Ensure all values are finite
        samples = samples[np.isfinite(samples)]
        
        return samples
    
    def _analyze_samples(
        self,
        name: str,
        samples: np.ndarray,
        confidence_level: float
    ) -> MonteCarloResult:
        """
        Analyze samples and calculate statistics
        
        Args:
            name: Variable name
            samples: Array of samples
            confidence_level: Confidence level for intervals
            
        Returns:
            Monte Carlo result for the variable
        """
        # Basic statistics
        mean = float(np.mean(samples))
        std_dev = float(np.std(samples))
        median = float(np.median(samples))
        min_val = float(np.min(samples))
        max_val = float(np.max(samples))
        
        # Percentiles
        percentiles = {
            'p5': float(np.percentile(samples, 5)),
            'p10': float(np.percentile(samples, 10)),
            'p25': float(np.percentile(samples, 25)),
            'p50': float(np.percentile(samples, 50)),
            'p75': float(np.percentile(samples, 75)),
            'p90': float(np.percentile(samples, 90)),
            'p95': float(np.percentile(samples, 95))
        }
        
        # Confidence interval
        alpha = 1 - confidence_level
        ci_lower = float(np.percentile(samples, (alpha / 2) * 100))
        ci_upper = float(np.percentile(samples, (1 - alpha / 2) * 100))
        
        # Try to fit a distribution
        distribution_fit = self._fit_distribution(samples)
        
        # Generate histogram
        hist, bin_edges = np.histogram(samples, bins=50, density=True)
        
        histogram = {
            'bins': bin_edges.tolist(),
            'counts': hist.tolist()
        }
        
        return MonteCarloResult(
            variable_name=name,
            mean=mean,
            std_dev=std_dev,
            median=median,
            min_value=min_val,
            max_value=max_val,
            percentiles=percentiles,
            confidence_interval={
                'lower': ci_lower,
                'upper': ci_upper,
                'level': confidence_level
            },
            distribution_fit=distribution_fit,
            histogram=histogram
        )
    
    def _fit_distribution(self, samples: np.ndarray) -> Optional[Dict[str, Any]]:
        """
        Fit the best distribution to the samples
        
        Args:
            samples: Array of samples
            
        Returns:
            Best-fit distribution parameters or None
        """
        try:
            # Try fitting a normal distribution first
            mu, sigma = norm.fit(samples)
            
            # Calculate goodness of fit
            _, p_value = stats.kstest(samples, 'norm', args=(mu, sigma))
            
            return {
                'distribution': 'normal',
                'parameters': {'mean': mu, 'std_dev': sigma},
                'ks_p_value': p_value
            }
        except Exception:
            return None
    
    def _evaluate_formula(
        self,
        formula: str,
        samples: Dict[str, np.ndarray],
        n_samples: int
    ) -> np.ndarray:
        """
        Evaluate a formula using the generated samples
        
        Args:
            formula: Python expression for output calculation
            samples: Dictionary of variable samples
            n_samples: Number of samples
            
        Returns:
            Calculated output samples
        """
        # Create local namespace with all variables
        local_ns = {**samples}
        
        # Add numpy functions
        local_ns.update({
            'np': np,
            'sum': np.sum,
            'mean': np.mean,
            'std': np.std,
            'exp': np.exp,
            'log': np.log,
            'sqrt': np.sqrt,
            'abs': np.abs,
            'maximum': np.maximum,
            'minimum': np.minimum
        })
        
        try:
            result = eval(formula, {"__builtins__": {}}, local_ns)
            
            if isinstance(result, (int, float)):
                # Broadcast scalar to array
                return np.full(n_samples, result)
            elif isinstance(result, np.ndarray):
                return result
            else:
                raise ValueError(f"Formula returned unexpected type: {type(result)}")
                
        except Exception as e:
            raise ValueError(f"Failed to evaluate formula '{formula}': {str(e)}")
    
    def _calculate_correlation_matrix(
        self,
        samples: Dict[str, np.ndarray]
    ) -> Optional[List[List[float]]]:
        """
        Calculate correlation matrix for all variables
        
        Args:
            samples: Dictionary of variable samples
            
        Returns:
            Correlation matrix or None if insufficient data
        """
        if len(samples) < 2:
            return None
        
        var_names = list(samples.keys())
        n_vars = len(var_names)
        
        # Stack samples into matrix
        data_matrix = np.column_stack([samples[name] for name in var_names])
        
        # Calculate correlation matrix
        corr_matrix = np.corrcoef(data_matrix.T)
        
        return corr_matrix.tolist()
    
    def _calculate_convergence_metrics(
        self,
        samples: Dict[str, np.ndarray],
        total_iterations: int
    ) -> Dict[str, float]:
        """
        Calculate convergence diagnostics
        
        Args:
            samples: Dictionary of variable samples
            total_iterations: Total number of iterations
            
        Returns:
            Convergence metrics
        """
        # Calculate mean estimate stability
        metrics = {}
        
        for name, var_samples in samples.items():
            # Split into batches and check convergence
            batch_size = max(total_iterations // 10, 100)
            batch_means = [
                np.mean(var_samples[i:i + batch_size])
                for i in range(0, len(var_samples), batch_size)
            ]
            
            # Standard error of batch means
            sem = np.std(batch_means) / np.sqrt(len(batch_means))
            
            # Relative efficiency
            rel_eff = np.var(batch_means) / np.var(var_samples)
            
            metrics[f"{name}_sem"] = sem
            metrics[f"{name}_relative_efficiency"] = rel_eff
        
        # Overall convergence score
        convergence_scores = [
            1 / (1 + m) for m in metrics.values() if m > 0
        ]
        metrics['overall_convergence'] = np.mean(convergence_scores) if convergence_scores else 0.0
        
        return metrics
    
    def _generate_histogram_data(
        self,
        input_samples: Dict[str, np.ndarray],
        output_samples: Dict[str, np.ndarray]
    ) -> Dict[str, Any]:
        """
        Generate histogram data for visualization
        
        Args:
            input_samples: Input variable samples
            output_samples: Output variable samples
            
        Returns:
            Histogram data for visualization
        """
        histograms = {}
        
        for name, samples in {**input_samples, **output_samples}.items():
            hist, bin_edges = np.histogram(samples, bins=50, density=True)
            
            histograms[name] = {
                'bins': bin_edges.tolist(),
                'density': hist.tolist(),
                'count': len(samples)
            }
        
        return histograms
    
    def _create_failed_response(
        self,
        request: MonteCarloRequest,
        error_message: str,
        duration_ms: float
    ) -> MonteCarloResponse:
        """Create a failed Monte Carlo simulation response"""
        return MonteCarloResponse(
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
            iterations_completed=0,
            input_results={},
            output_results={},
            correlation_matrix=None,
            convergence_metrics=None
        )
