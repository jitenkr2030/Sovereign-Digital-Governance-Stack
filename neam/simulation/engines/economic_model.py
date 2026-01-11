"""
NEAM Simulation Service - Economic Model Engine
Economic projection and forecasting engine using time-series methods
"""

import logging
import time
from datetime import datetime, timedelta
from typing import Any, Callable, Dict, List, Optional, Tuple
import numpy as np
import pandas as pd
from scipy import stats
from scipy.signal import detrend
from statsmodels.tsa.arima.model import ARIMA
from statsmodels.tsa.holtwinters import ExponentialSmoothing
from statsmodels.tsa.stattools import adfuller, acf, pacf
from statsmodels.tsa.seasonal import seasonal_decompose
from statsmodels.nonparametric.smoothers_lowess import lowess
import warnings

from .base import BaseSimulationEngine, ProgressTracker
from ..models.simulation import (
    EconomicProjectionRequest,
    EconomicProjectionResult,
    EconomicProjectionResponse,
    EconomicIndicator,
    SimulationStatus,
)


logger = logging.getLogger(__name__)


class EconomicModelEngine(BaseSimulationEngine):
    """
    Economic projection engine using various time-series forecasting methods.
    
    Supports ARIMA, Exponential Smoothing, and other econometric models
    for projecting economic indicators into the future.
    """
    
    def __init__(
        self,
        max_workers: int = 4,
        default_timeout: float = 300.0,
        random_seed: Optional[int] = None
    ):
        """
        Initialize the economic model engine
        
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
        request: EconomicProjectionRequest,
        progress_callback: Optional[Callable[[float], None]] = None
    ) -> EconomicProjectionResponse:
        """
        Execute economic projection simulation
        
        Args:
            request: Economic projection parameters
            progress_callback: Optional progress callback
            
        Returns:
            Economic projection results
        """
        start_time = time.time()
        created_at = datetime.utcnow()
        
        # Validate request
        validation_errors = self.validate_request(request)
        if validation_errors:
            raise ValueError(f"Validation errors: {', '.join(validation_errors)}")
        
        # Create time series for each indicator
        logger.info(f"Processing {len(request.indicators)} indicators over {request.projection_months} months")
        
        tracker = ProgressTracker(
            total_steps=len(request.indicators),
            callback=progress_callback,
            update_frequency=1
        )
        
        results: Dict[str, EconomicProjectionResult] = {}
        scenario_assumptions = self._get_default_assumptions()
        
        # Process each indicator
        for indicator in request.indicators:
            try:
                # Generate historical data points (simulated for now)
                historical_data = self._generate_historical_data(
                    indicator,
                    request.start_date,
                    request.end_date
                )
                
                # Fit forecasting model
                model_result = self._fit_forecast_model(
                    historical_data,
                    request.projection_months,
                    request.model_type,
                    request.seasonality,
                    request.trend_type
                )
                
                # Generate projections
                projection_result = self._generate_projections(
                    indicator,
                    model_result,
                    request.projection_months,
                    request.start_date
                )
                
                # Calculate confidence intervals
                confidence_intervals = self._calculate_confidence_intervals(
                    model_result,
                    request.projection_months,
                    request.confidence_level if hasattr(request, 'confidence_level') else 0.95
                )
                
                # Analyze results
                result = self._analyze_projection(
                    indicator,
                    projection_result,
                    confidence_intervals,
                    model_result
                )
                
                results[indicator.code] = result
                
            except Exception as e:
                logger.error(f"Failed to project indicator {indicator.code}: {str(e)}")
                results[indicator.code] = self._create_error_result(indicator)
            
            tracker.update()
        
        tracker.complete()
        
        # Identify risk factors
        risk_factors = self._identify_risk_factors(results)
        
        duration_ms = (time.time() - start_time) * 1000
        
        return EconomicProjectionResponse(
            simulation_id=request.simulation_id or self._generate_simulation_id(),
            simulation_type=request.simulation_type,
            status=SimulationStatus.COMPLETED,
            created_at=created_at,
            completed_at=datetime.utcnow(),
            duration_ms=duration_ms,
            regions=request.regions,
            sectors=request.sectors,
            result={
                "summary": f"Economic projection completed for {len(results)} indicators",
                "projection_months": request.projection_months,
                "model_type": request.model_type
            },
            metadata={
                "seasonality": request.seasonality,
                "trend_type": request.trend_type,
                "external_factors": request.external_factors
            },
            projection_months=request.projection_months,
            model_type=request.model_type,
            results=results,
            scenario_assumptions=scenario_assumptions,
            risk_factors=risk_factors
        )
    
    def _generate_historical_data(
        self,
        indicator: EconomicIndicator,
        start_date: datetime,
        end_date: datetime
    ) -> pd.Series:
        """
        Generate historical data for an indicator
        
        In production, this would fetch from a database or API.
        For simulation, we generate realistic synthetic data.
        
        Args:
            indicator: Economic indicator definition
            start_date: Start of historical period
            end_date: End of historical period
            
        Returns:
            Time series of historical values
        """
        # Calculate number of months
        months = (end_date.year - start_date.year) * 12 + (end_date.month - start_date.month)
        
        # Generate time index
        dates = pd.date_range(start=start_date, periods=months + 1, freq='MS')
        
        # Generate synthetic data with trend and seasonality
        n_points = len(dates)
        
        # Base value with some noise
        base_value = indicator.value
        volatility = indicator.volatility or 0.02
        growth_rate = indicator.growth_rate or 0.005
        
        # Generate values with trend
        trend = np.linspace(0, growth_rate * n_points, n_points)
        noise = np.random.normal(0, volatility, n_points)
        
        # Add seasonal component
        seasonal = 0.03 * np.sin(2 * np.pi * np.arange(n_points) / 12)
        
        values = base_value * (1 + trend + noise + seasonal)
        
        # Ensure non-negative for certain indicators
        if indicator.code in ['unemployment_rate', 'inflation_rate']:
            values = np.maximum(values, 0)
        
        return pd.Series(values, index=dates)
    
    def _fit_forecast_model(
        self,
        data: pd.Series,
        forecast_periods: int,
        model_type: str,
        seasonality: bool,
        trend_type: str
    ) -> Dict[str, Any]:
        """
        Fit a forecasting model to the data
        
        Args:
            data: Historical time series data
            forecast_periods: Number of periods to forecast
            model_type: Type of model to fit
            seasonality: Whether to include seasonal component
            trend_type: Type of trend to model
            
        Returns:
            Fitted model and parameters
        """
        model_result = {
            'model_type': model_type,
            'fitted': False,
            'parameters': {},
            'forecast': None,
            'residuals': None,
            'fit_metrics': {}
        }
        
        try:
            with warnings.catch_warnings():
                warnings.simplefilter("ignore")
                
                if model_type == 'arima':
                    # Fit ARIMA model
                    # Auto-select order if not specified
                    order = self._auto_arima_order(data)
                    
                    model = ARIMA(data, order=order)
                    fitted = model.fit()
                    
                    # Generate forecast
                    forecast = fitted.forecast(steps=forecast_periods)
                    residuals = fitted.resid
                    
                    model_result['parameters'] = {
                        'order': order,
                        'aic': fitted.aic,
                        'bic': fitted.bic
                    }
                    model_result['forecast'] = forecast
                    model_result['residuals'] = residuals
                    model_result['fit_metrics'] = {
                        'aic': fitted.aic,
                        'bic': fitted.bic,
                        'mse': float(np.mean(residuals ** 2))
                    }
                    model_result['fitted'] = True
                    
                elif model_type == 'exponential_smoothing':
                    # Fit exponential smoothing model
                    trend = 'add' if trend_type in ['linear', 'damped'] else None
                    damped_trend = trend_type == 'damped'
                    
                    model = ExponentialSmoothing(
                        data,
                        trend=trend,
                        damped_trend=damped_trend,
                        seasonal='add' if seasonality else None,
                        seasonal_periods=12 if seasonality else None
                    )
                    fitted = model.fit(optimized=True)
                    
                    forecast = fitted.forecast(steps=forecast_periods)
                    residuals = data - fitted.fittedvalues
                    
                    model_result['parameters'] = {
                        'smoothing_level': fitted.params.get('smoothing_level'),
                        'smoothing_trend': fitted.params.get('smoothing_trend'),
                        'damping_trend': fitted.params.get('damping_trend')
                    }
                    model_result['forecast'] = forecast
                    model_result['residuals'] = residuals
                    model_result['fit_metrics'] = {
                        'mse': float(np.mean(residuals ** 2)),
                        'mae': float(np.mean(np.abs(residuals)))
                    }
                    model_result['fitted'] = True
                    
                else:
                    # Default to simple linear extrapolation
                    forecast = self._linear_extrapolation(data, forecast_periods)
                    residuals = np.zeros(len(data))
                    
                    model_result['forecast'] = forecast
                    model_result['residuals'] = residuals
                    model_result['fit_metrics'] = {
                        'mse': 0,
                        'method': 'linear_extrapolation'
                    }
                    model_result['fitted'] = True
                    
        except Exception as e:
            logger.warning(f"Model fitting failed: {str(e)}. Using fallback method.")
            # Fallback to linear extrapolation
            model_result['forecast'] = self._linear_extrapolation(data, forecast_periods)
            model_result['fallback'] = True
        
        return model_result
    
    def _auto_arima_order(self, data: pd.Series) -> Tuple[int, int, int]:
        """
        Automatically determine optimal ARIMA order using AIC
        
        Args:
            data: Time series data
            
        Returns:
            (p, d, q) order for ARIMA model
        """
        # Simple heuristic for ARIMA order selection
        # In production, use pmdarima.auto_arima
        
        # Check for stationarity
        adf_result = adfuller(data.dropna())
        d = 0 if adf_result[1] < 0.05 else 1
        
        # Simple bounds for p and q
        max_p = min(5, len(data) // 10)
        max_q = min(5, len(data) // 10)
        
        best_aic = np.inf
        best_order = (1, d, 1)
        
        for p in range(max_p + 1):
            for q in range(max_q + 1):
                try:
                    model = ARIMA(data, order=(p, d, q))
                    fitted = model.fit()
                    
                    if fitted.aic < best_aic:
                        best_aic = fitted.aic
                        best_order = (p, d, q)
                except Exception:
                    continue
        
        return best_order
    
    def _linear_extrapolation(
        self,
        data: pd.Series,
        forecast_periods: int
    ) -> pd.Series:
        """
        Simple linear extrapolation forecast
        
        Args:
            data: Historical time series
            forecast_periods: Number of periods to forecast
            
        Returns:
            Forecasted values
        """
        # Fit linear trend
        x = np.arange(len(data))
        slope, intercept, _, _, _ = stats.linregress(x, data.values)
        
        # Generate forecast
        future_x = np.arange(len(data), len(data) + forecast_periods)
        forecast_values = intercept + slope * future_x
        
        # Create date index
        last_date = data.index[-1]
        future_dates = pd.date_range(
            start=last_date + pd.DateOffset(months=1),
            periods=forecast_periods,
            freq='MS'
        )
        
        return pd.Series(forecast_values, index=future_dates)
    
    def _generate_projections(
        self,
        indicator: EconomicIndicator,
        model_result: Dict[str, Any],
        projection_months: int,
        start_date: datetime
    ) -> List[Dict[str, Any]]:
        """
        Generate detailed projection data
        
        Args:
            indicator: Economic indicator
            model_result: Fitted model result
            projection_months: Number of months to project
            start_date: Projection start date
            
        Returns:
            List of projection data points
        """
        forecast = model_result.get('forecast', pd.Series())
        
        if forecast is None or len(forecast) == 0:
            # Generate placeholder projections
            forecast = self._linear_extrapolation(
                pd.Series([indicator.value] * 12),
                projection_months
            )
        
        projections = []
        
        for i, (date, value) in enumerate(forecast.items()):
            projections.append({
                'date': date.isoformat(),
                'month': i + 1,
                'value': float(value),
                'yoy_change': float((value / indicator.value - 1) * 100) if i >= 11 else None,
                'mom_change': float((value / forecast.iloc[i-1] - 1) * 100) if i > 0 else None
            })
        
        return projections
    
    def _calculate_confidence_intervals(
        self,
        model_result: Dict[str, Any],
        projection_months: int,
        confidence_level: float
    ) -> List[Dict[str, Any]]:
        """
        Calculate confidence intervals for projections
        
        Args:
            model_result: Fitted model result
            projection_months: Number of months projected
            confidence_level: Confidence level (e.g., 0.95)
            
        Returns:
            Confidence interval data for each projection point
        """
        residuals = model_result.get('residuals', np.array([]))
        forecast = model_result.get('forecast', pd.Series())
        
        if len(residuals) == 0:
            # Return placeholder intervals
            alpha = 1 - confidence_level
            return [
                {
                    'lower': float(forecast.iloc[i]) * 0.95,
                    'upper': float(forecast.iloc[i]) * 1.05,
                    'level': confidence_level
                }
                for i in range(len(forecast))
            ]
        
        # Estimate standard error from residuals
        std_error = np.std(residuals)
        
        # Calculate confidence intervals
        z_score = stats.norm.ppf(1 - alpha / 2)
        
        intervals = []
        
        for i, value in enumerate(forecast):
            # Confidence interval widens over time
            time_factor = np.sqrt(i + 1)
            margin = z_score * std_error * time_factor
            
            intervals.append({
                'lower': float(value - margin),
                'upper': float(value + margin),
                'level': confidence_level
            })
        
        return intervals
    
    def _analyze_projection(
        self,
        indicator: EconomicIndicator,
        projections: List[Dict[str, Any]],
        confidence_intervals: List[Dict[str, Any]],
        model_result: Dict[str, Any]
    ) -> EconomicProjectionResult:
        """
        Analyze projection results
        
        Args:
            indicator: Economic indicator
            projections: Projection data
            confidence_intervals: Confidence interval data
            model_result: Fitted model result
            
        Returns:
            Analyzed projection result
        """
        values = [p['value'] for p in projections]
        
        # Calculate expected growth
        start_value = indicator.value
        end_value = values[-1]
        expected_growth = (end_value / start_value - 1) * 100
        
        # Growth confidence interval
        start_ci_lower = confidence_intervals[0]['lower']
        start_ci_upper = confidence_intervals[0]['upper']
        end_ci_lower = confidence_intervals[-1]['lower']
        end_ci_upper = confidence_intervals[-1]['upper']
        
        growth_ci_lower = (end_ci_lower / start_value - 1) * 100
        growth_ci_upper = (end_ci_upper / start_value - 1) * 100
        
        # Find peak and trough
        peak_idx = np.argmax(values)
        trough_idx = np.argmin(values)
        
        peak_value = values[peak_idx]
        peak_date = projections[peak_idx]['date']
        trough_value = values[trough_idx]
        trough_date = projections[trough_idx]['date']
        
        return EconomicProjectionResult(
            indicator_code=indicator.code,
            indicator_name=indicator.name,
            projections=projections,
            confidence_intervals=confidence_intervals,
            expected_growth=expected_growth,
            growth_ci_lower=growth_ci_lower,
            growth_ci_upper=growth_ci_upper,
            peak_value=peak_value,
            peak_date=datetime.fromisoformat(peak_date),
            trough_value=trough_value,
            trough_date=datetime.fromisoformat(trough_date),
            model_parameters=model_result.get('parameters', {}),
            model_fit_metrics=model_result.get('fit_metrics', {})
        )
    
    def _identify_risk_factors(
        self,
        results: Dict[str, EconomicProjectionResult]
    ) -> List[Dict[str, Any]]:
        """
        Identify risk factors from projection results
        
        Args:
            results: Projection results for all indicators
            
        Returns:
            List of identified risk factors
        """
        risk_factors = []
        
        for code, result in results.items():
            # Check for high volatility
            values = [p['value'] for p in result.projections]
            volatility = np.std(values) / np.mean(values) if np.mean(values) != 0 else 0
            
            if volatility > 0.1:
                risk_factors.append({
                    'indicator': code,
                    'type': 'high_volatility',
                    'description': f"High volatility projected for {result.indicator_name}",
                    'severity': 'medium' if volatility < 0.2 else 'high',
                    'metric': volatility
                })
            
            # Check for declining trends
            if result.expected_growth < -5:
                risk_factors.append({
                    'indicator': code,
                    'type': 'declining_trend',
                    'description': f"Significant decline projected for {result.indicator_name}",
                    'severity': 'high' if result.expected_growth < -10 else 'medium',
                    'metric': result.expected_growth
                })
            
            # Check for widening uncertainty
            if len(result.confidence_intervals) > 1:
                first_ci = result.confidence_intervals[0]
                last_ci = result.confidence_intervals[-1]
                first_width = first_ci['upper'] - first_ci['lower']
                last_width = last_ci['upper'] - last_ci['lower']
                
                if last_width > 2 * first_width:
                    risk_factors.append({
                        'indicator': code,
                        'type': 'increasing_uncertainty',
                        'description': f"Uncertainty increasing for {result.indicator_name}",
                        'severity': 'low',
                        'metric': last_width / first_width
                    })
        
        return risk_factors
    
    def _get_default_assumptions(self) -> Dict[str, Any]:
        """Get default scenario assumptions"""
        return {
            'productivity_growth': 0.02,
            'population_growth': 0.005,
            'technology_progress': 0.03,
            'policy_stance': 'neutral',
            'global_growth': 0.025,
            'trade_conditions': 'stable'
        }
    
    def _create_error_result(
        self,
        indicator: EconomicIndicator
    ) -> EconomicProjectionResult:
        """Create an error result for a failed projection"""
        return EconomicProjectionResult(
            indicator_code=indicator.code,
            indicator_name=indicator.name,
            projections=[],
            confidence_intervals=[],
            expected_growth=0,
            growth_ci_lower=0,
            growth_ci_upper=0,
            peak_value=indicator.value,
            peak_date=datetime.utcnow(),
            trough_value=indicator.value,
            trough_date=datetime.utcnow(),
            model_parameters={},
            model_fit_metrics={}
        )
    
    def _create_failed_response(
        self,
        request: EconomicProjectionRequest,
        error_message: str,
        duration_ms: float
    ) -> EconomicProjectionResponse:
        """Create a failed economic projection response"""
        return EconomicProjectionResponse(
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
            projection_months=request.projection_months,
            model_type=request.model_type,
            results={},
            scenario_assumptions={},
            risk_factors=[]
        )
