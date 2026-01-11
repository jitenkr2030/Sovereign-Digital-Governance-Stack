"""
NEAM Simulation Service - Unit Tests
Comprehensive test suite for simulation engines and API endpoints
"""

import pytest
import numpy as np
from datetime import datetime, timedelta
from unittest.mock import Mock, patch, AsyncMock
import sys
import os

# Add parent directory to path
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


class TestMonteCarloEngine:
    """Tests for Monte Carlo simulation engine"""
    
    def test_normal_distribution_sampling(self):
        """Test normal distribution sample generation"""
        from simulation.engines.monte_carlo import MonteCarloEngine
        from simulation.models.simulation import DistributionParameter, DistributionType
        
        engine = MonteCarloEngine(random_seed=42)
        
        dist = DistributionParameter(
            distribution_type=DistributionType.NORMAL,
            mean=100.0,
            std_dev=15.0
        )
        
        samples = engine._generate_samples(dist, 1000)
        
        assert len(samples) == 1000
        assert np.mean(samples) == pytest.approx(100.0, abs=2.0)
        assert np.std(samples) == pytest.approx(15.0, abs=1.0)
    
    def test_uniform_distribution_sampling(self):
        """Test uniform distribution sample generation"""
        from simulation.engines.monte_carlo import MonteCarloEngine
        from simulation.models.simulation import DistributionParameter, DistributionType
        
        engine = MonteCarloEngine(random_seed=42)
        
        dist = DistributionParameter(
            distribution_type=DistributionType.UNIFORM,
            min_value=0.0,
            max_value=100.0
        )
        
        samples = engine._generate_samples(dist, 1000)
        
        assert len(samples) == 1000
        assert np.min(samples) >= 0.0
        assert np.max(samples) <= 100.0
        assert np.mean(samples) == pytest.approx(50.0, abs=3.0)
    
    def test_sample_analysis(self):
        """Test sample statistical analysis"""
        from simulation.engines.monte_carlo import MonteCarloEngine
        from simulation.models.simulation import DistributionParameter, DistributionType
        
        engine = MonteCarloEngine(random_seed=42)
        
        # Generate samples
        dist = DistributionParameter(
            distribution_type=DistributionType.NORMAL,
            mean=50.0,
            std_dev=10.0
        )
        
        samples = engine._generate_samples(dist, 5000)
        
        # Analyze samples
        result = engine._analyze_samples("test_var", samples, 0.95)
        
        assert result.variable_name == "test_var"
        assert result.mean == pytest.approx(50.0, abs=1.0)
        assert result.std_dev == pytest.approx(10.0, abs=0.5)
        assert result.median == pytest.approx(50.0, abs=1.0)
        assert result.min_value < result.mean
        assert result.max_value > result.mean
        assert "p5" in result.percentiles
        assert "p95" in result.percentiles
        assert "lower" in result.confidence_interval
        assert "upper" in result.confidence_interval
        assert result.histogram is not None
    
    def test_correlation_matrix_calculation(self):
        """Test correlation matrix calculation"""
        from simulation.engines.monte_carlo import MonteCarloEngine
        
        engine = MonteCarloEngine(random_seed=42)
        
        samples = {
            "var1": np.random.normal(0, 1, 1000),
            "var2": np.random.normal(0, 1, 1000),
            "var3": np.random.normal(0, 1, 1000)
        }
        
        # Add correlation
        samples["var2"] = samples["var1"] * 0.5 + np.random.normal(0, 0.5, 1000)
        
        corr_matrix = engine._calculate_correlation_matrix(samples)
        
        assert corr_matrix is not None
        assert len(corr_matrix) == 3
        assert len(corr_matrix[0]) == 3
        # Diagonal should be 1.0
        assert abs(corr_matrix[0][0] - 1.0) < 0.01
        # var1 and var2 should be correlated
        assert abs(corr_matrix[0][1]) > 0.3
    
    def test_formula_evaluation(self):
        """Test formula evaluation for output variables"""
        from simulation.engines.monte_carlo import MonteCarloEngine
        
        engine = MonteCarloEngine(random_seed=42)
        
        samples = {
            "revenue": np.array([100, 200, 300, 400, 500]),
            "costs": np.array([50, 80, 120, 150, 200])
        }
        
        # Test profit calculation
        profit = engine._evaluate_formula("revenue - costs", samples, 5)
        
        expected = np.array([50, 120, 180, 250, 300])
        assert np.allclose(profit, expected)
    
    def test_convergence_metrics(self):
        """Test convergence metrics calculation"""
        from simulation.engines.monte_carlo import MonteCarloEngine
        
        engine = MonteCarloEngine(random_seed=42)
        
        samples = {
            "var1": np.random.normal(100, 15, 10000),
            "var2": np.random.uniform(0, 100, 10000)
        }
        
        metrics = engine._calculate_convergence_metrics(samples, 10000)
        
        assert "var1_sem" in metrics
        assert "var2_sem" in metrics
        assert "overall_convergence" in metrics
        assert 0 <= metrics["overall_convergence"] <= 1


class TestEconomicModelEngine:
    """Tests for Economic Model engine"""
    
    def test_linear_extrapolation(self):
        """Test linear extrapolation forecast"""
        from simulation.engines.economic_model import EconomicModelEngine
        
        engine = EconomicModelEngine(random_seed=42)
        
        # Create simple linear series
        data = np.array([100, 102, 104, 106, 108])
        dates = [datetime(2024, i + 1, 1) for i in range(5)]
        series = pd.Series(data, index=dates)
        
        forecast = engine._linear_extrapolation(series, 6)
        
        assert len(forecast) == 6
        # Should continue the linear trend
        assert forecast.iloc[-1] > series.iloc[-1]
    
    def test_auto_arima_order(self):
        """Test automatic ARIMA order selection"""
        from simulation.engines.economic_model import EconomicModelEngine
        import pandas as pd
        
        engine = EconomicModelEngine(random_seed=42)
        
        # Generate AR(1) process
        np.random.seed(42)
        n = 100
        phi = 0.8
        data = np.zeros(n)
        data[0] = np.random.normal(0, 1)
        for t in range(1, n):
            data[t] = phi * data[t-1] + np.random.normal(0, 1)
        
        series = pd.Series(data)
        
        order = engine._auto_arima_order(series)
        
        assert len(order) == 3
        assert order[0] >= 0  # p
        assert order[1] >= 0  # d
        assert order[2] >= 0  # q
    
    def test_historical_data_generation(self):
        """Test historical data generation"""
        from simulation.engines.economic_model import EconomicModelEngine
        from simulation.models.simulation import EconomicIndicator
        
        engine = EconomicModelEngine(random_seed=42)
        
        indicator = EconomicIndicator(
            name="GDP",
            code="gdp",
            value=100.0,
            unit="billion",
            growth_rate=0.02,
            volatility=0.01
        )
        
        start_date = datetime(2023, 1, 1)
        end_date = datetime(2024, 1, 1)
        
        historical = engine._generate_historical_data(indicator, start_date, end_date)
        
        assert len(historical) == 13  # 12 months + 1
        assert historical.iloc[0] == pytest.approx(100.0, abs=1.0)
        # Should show overall growth
        assert historical.iloc[-1] > historical.iloc[0]
    
    def test_projection_analysis(self):
        """Test projection analysis and metrics"""
        from simulation.engines.economic_model import EconomicModelEngine
        from simulation.models.simulation import EconomicIndicator
        
        engine = EconomicModelEngine(random_seed=42)
        
        indicator = EconomicIndicator(
            name="CPI",
            code="cpi",
            value=100.0,
            unit="index",
            growth_rate=0.02,
            volatility=0.005
        )
        
        # Create sample projections
        projections = [
            {"date": (datetime(2024, 1, 1) + timedelta(days=30*i)).isoformat(), "value": 100.0 + i}
            for i in range(12)
        ]
        
        confidence_intervals = [
            {"lower": p["value"] - 1, "upper": p["value"] + 1}
            for p in projections
        ]
        
        model_result = {
            "parameters": {"order": (1, 0, 1)},
            "fit_metrics": {"aic": 100.5}
        }
        
        result = engine._analyze_projection(
            indicator, projections, confidence_intervals, model_result
        )
        
        assert result.indicator_code == "cpi"
        assert result.indicator_name == "CPI"
        assert len(result.projections) == 12
        assert "expected_growth" in result.__fields_set__
        assert "peak_value" in result.__fields_set__
        assert "trough_value" in result.__fields_set__


class TestScenarioEngine:
    """Tests for Scenario Engine"""
    
    def test_base_projection_generation(self):
        """Test base projection without interventions"""
        from simulation.engines.scenario import ScenarioEngine
        
        engine = ScenarioEngine(random_seed=42)
        
        projection = engine._generate_base_projection(100.0, 12)
        
        assert len(projection) == 12
        assert projection[0] > 0
        # Should show growth trend
        assert np.mean(projection) > 100.0
    
    def test_fiscal_stimulus_effect(self):
        """Test fiscal stimulus intervention effect"""
        from simulation.engines.scenario import ScenarioEngine
        from simulation.models.simulation import PolicyIntervention, InterventionType
        
        engine = ScenarioEngine(random_seed=42)
        
        intervention = PolicyIntervention(
            intervention_type=InterventionType.FISCAL_STIMULUS,
            name="Government Spending",
            magnitude=2.0,  # 2% of GDP
            unit="percent",
            start_date=datetime(2024, 1, 1),
            lag_months=0
        )
        
        effect = engine._calculate_fiscal_stimulus_effect(intervention, 12)
        
        assert len(effect) == 12
        assert effect[0] > 0  # Immediate effect
        assert effect[5] > effect[0]  # Effect builds up
    
    def test_interest_rate_effect(self):
        """Test interest rate adjustment effect"""
        from simulation.engines.scenario import ScenarioEngine
        from simulation.models.simulation import PolicyIntervention, InterventionType
        
        engine = ScenarioEngine(random_seed=42)
        
        # Rate cut
        intervention = PolicyIntervention(
            intervention_type=InterventionType.INTEREST_RATE_ADJUSTMENT,
            name="Rate Cut",
            magnitude=0.5,  # 0.5% cut
            unit="percent",
            start_date=datetime(2024, 1, 1),
            lag_months=0
        )
        
        effect = engine._calculate_interest_rate_effect(intervention, 12)
        
        assert len(effect) == 12
        assert effect[0] == 0  # No immediate effect
        assert effect[3] > 0  # Effect after lag
    
    def test_scenario_comparison(self):
        """Test scenario comparison calculation"""
        from simulation.engines.scenario import ScenarioEngine
        from simulation.models.simulation import ScenarioInput, PolicyIntervention, InterventionType
        
        engine = ScenarioEngine(random_seed=42)
        
        baseline = {
            "gdp": [100, 102, 104, 106, 108],
            "inflation": [2.0, 2.1, 2.2, 2.3, 2.4]
        }
        
        scenario = {
            "gdp": [100, 103, 106, 109, 112],  # Higher growth
            "inflation": [2.0, 2.2, 2.4, 2.6, 2.8]  # Higher inflation
        }
        
        scenario_input = ScenarioInput(
            name="Stimulus",
            base_conditions={"gdp": 100, "inflation": 2.0},
            interventions=[
                PolicyIntervention(
                    intervention_type=InterventionType.FISCAL_STIMULUS,
                    name="Stimulus",
                    magnitude=2.0,
                    unit="percent",
                    start_date=datetime(2024, 1, 1)
                )
            ]
        )
        
        comparison = engine._compare_scenarios(
            "Stimulus", baseline, scenario, scenario_input, ["gdp", "inflation"]
        )
        
        assert comparison.scenario_name == "Stimulus"
        assert "gdp" in comparison.cumulative_effect
        assert comparison.cumulative_effect["gdp"] > 0  # Positive effect


class TestSensitivityEngine:
    """Tests for Sensitivity Analysis Engine"""
    
    def test_random_sampling(self):
        """Test random parameter sampling"""
        from simulation.engines.sensitivity import SensitivityEngine
        from simulation.models.simulation import ParameterRange
        
        engine = SensitivityEngine(random_seed=42)
        
        param = ParameterRange(
            name="growth_rate",
            base_value=0.02,
            min_value=-0.05,
            max_value=0.10,
            step_type="linear"
        )
        
        samples = engine._random_sample(param, 1000)
        
        assert len(samples) == 1000
        assert np.min(samples) >= -0.05
        assert np.max(samples) <= 0.10
        assert np.mean(samples) == pytest.approx(0.025, abs=0.01)
    
    def test_default_model_evaluation(self):
        """Test default economic model evaluation"""
        from simulation.engines.sensitivity import SensitivityEngine
        
        engine = SensitivityEngine(random_seed=42)
        
        sample_matrix = np.array([
            [0.02, 0.2, 0.02, 0.05],  # Normal values
            [0.05, 0.3, 0.03, 0.03],  # High growth
            [-0.02, 0.1, 0.01, 0.10],  # Recession
        ])
        
        param_names = ["gdp_growth", "investment_rate", "productivity_growth", "unemployment_rate"]
        
        outputs = engine._evaluate_default_model(sample_matrix, param_names, 3)
        
        assert len(outputs) == 3
        # High growth scenario should have higher output
        assert outputs[1] > outputs[0]
    
    def test_parameter_ranking(self):
        """Test parameter sensitivity ranking"""
        from simulation.engines.sensitivity import SensitivityEngine
        from simulation.models.simulation import SensitivityAnalysis
        
        engine = SensitivityEngine(random_seed=42)
        
        sensitivity = SensitivityAnalysis(
            output_metric="gdp",
            method="sobol",
            parameter_sensitivities={},
            total_order_indices={
                "investment_rate": 0.35,
                "productivity_growth": 0.25,
                "unemployment_rate": 0.20,
                "inflation_rate": 0.15
            },
            first_order_indices={
                "investment_rate": 0.30,
                "productivity_growth": 0.20,
                "unemployment_rate": 0.15,
                "inflation_rate": 0.10
            },
            interaction_indices={},
            parameter_rankings=[],
            confidence_intervals={}
        )
        
        rankings = engine._rank_parameters(sensitivity)
        
        assert len(rankings) == 4
        # Should be sorted by sensitivity
        assert rankings[0]["rank"] == 1
        assert rankings[0]["parameter"] == "investment_rate"
        assert rankings[0]["interpretation"] == "high"
    
    def test_sensitivity_interpretation(self):
        """Test sensitivity index interpretation"""
        from simulation.engines.sensitivity import SensitivityEngine
        
        engine = SensitivityEngine()
        
        assert engine._interpret_sensitivity(0.02) == "negligible"
        assert engine._interpret_sensitivity(0.07) == "low"
        assert engine._interpret_sensitivity(0.15) == "moderate"
        assert engine._interpret_sensitivity(0.30) == "high"
        assert engine._interpret_sensitivity(0.50) == "very high"


class TestAPIEndpoints:
    """Tests for API endpoints"""
    
    @pytest.fixture
    def client(self):
        """Create test client"""
        from simulation.main import app
        from fastapi.testclient import TestClient
        
        return TestClient(app)
    
    def test_root_endpoint(self, client):
        """Test root endpoint"""
        response = client.get("/")
        
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "NEAM Simulation Service"
        assert data["version"] == "1.0.0"
    
    def test_health_check(self, client):
        """Test health check endpoint"""
        response = client.get("/api/v1/health")
        
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        assert "components" in data
    
    def test_liveness_probe(self, client):
        """Test liveness probe endpoint"""
        response = client.get("/api/v1/health/live")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "alive"
    
    def test_simulation_types_endpoint(self, client):
        """Test simulation types endpoint"""
        response = client.get("/api/v1/simulations/types")
        
        assert response.status_code == 200
        data = response.json()
        assert "simulation_types" in data
        assert len(data["simulation_types"]) >= 4


class TestModels:
    """Tests for Pydantic models"""
    
    def test_monte_carlo_request_validation(self):
        """Test Monte Carlo request validation"""
        from simulation.models.simulation import MonteCarloRequest, VariableInput, DistributionParameter, DistributionType
        
        request = MonteCarloRequest(
            simulation_type=SimulationType.MONTE_CARLO,
            start_date=datetime(2024, 1, 1),
            end_date=datetime(2025, 1, 1),
            variables=[
                VariableInput(
                    name="gdp_growth",
                    distribution=DistributionParameter(
                        distribution_type=DistributionType.NORMAL,
                        mean=0.02,
                        std_dev=0.01
                    )
                )
            ],
            output_variables=["projected_gdp"],
            iterations=1000
        )
        
        assert request.simulation_type == SimulationType.MONTE_CARLO
        assert request.iterations == 1000
        assert len(request.variables) == 1
        assert request.variables[0].name == "gdp_growth"
    
    def test_economic_indicator_model(self):
        """Test Economic Indicator model"""
        from simulation.models.simulation import EconomicIndicator
        
        indicator = EconomicIndicator(
            name="GDP",
            code="gdp",
            value=100.0,
            unit="billion",
            growth_rate=0.025,
            volatility=0.01
        )
        
        assert indicator.name == "GDP"
        assert indicator.value == 100.0
        assert indicator.growth_rate == 0.025
    
    def test_policy_intervention_model(self):
        """Test Policy Intervention model"""
        from simulation.models.simulation import PolicyIntervention, InterventionType
        
        intervention = PolicyIntervention(
            intervention_type=InterventionType.FISCAL_STIMULUS,
            name="Infrastructure Spending",
            description="Government infrastructure investment program",
            magnitude=2.0,
            unit="percent",
            start_date=datetime(2024, 1, 1),
            target_sectors=[SectorType.CONSTRUCTION, SectorType.MANUFACTURING],
            lag_months=6
        )
        
        assert intervention.intervention_type == InterventionType.FISCAL_STIMULUS
        assert intervention.magnitude == 2.0
        assert intervention.lag_months == 6
        assert SectorType.CONSTRUCTION in intervention.target_sectors
    
    def test_date_validation(self):
        """Test date validation in requests"""
        from simulation.models.simulation import MonteCarloRequest, SimulationStatus
        import pytest
        
        # Valid request
        valid_request = MonteCarloRequest(
            simulation_type=SimulationType.MONTE_CARLO,
            start_date=datetime(2024, 1, 1),
            end_date=datetime(2025, 1, 1),
            variables=[],
            output_variables=["output"],
            iterations=100
        )
        
        assert valid_request.end_date > valid_request.start_date
        
        # Invalid request (end before start)
        with pytest.raises(ValueError):
            MonteCarloRequest(
                simulation_type=SimulationType.MONTE_CARLO,
                start_date=datetime(2025, 1, 1),
                end_date=datetime(2024, 1, 1),
                variables=[],
                output_variables=["output"],
                iterations=100
            )


class TestProgressTracker:
    """Tests for Progress Tracker"""
    
    def test_progress_update(self):
        """Test progress tracking"""
        from simulation.engines.base import ProgressTracker
        
        updates = []
        
        def callback(progress):
            updates.append(progress)
        
        tracker = ProgressTracker(total_steps=10, callback=callback, update_frequency=2)
        
        tracker.update(2)
        assert len(updates) == 0  # Not enough progress for update
        
        tracker.update(3)
        assert len(updates) == 1
        assert updates[0] == pytest.approx(50.0, abs=1.0)
    
    def test_progress_completion(self):
        """Test progress completion"""
        from simulation.engines.base import ProgressTracker
        
        final_progress = []
        
        def callback(progress):
            final_progress.append(progress)
        
        tracker = ProgressTracker(total_steps=100, callback=callback)
        
        tracker.update(50)
        tracker.complete()
        
        assert len(final_progress) == 2
        assert final_progress[1] == 100.0
    
    def test_progress_percent_property(self):
        """Test progress percentage property"""
        from simulation.engines.base import ProgressTracker
        
        tracker = ProgressTracker(total_steps=100)
        
        assert tracker.percent == 0.0
        
        tracker.update(25)
        assert tracker.percent == pytest.approx(25.0, abs=0.1)


# Import pandas for tests
import pandas as pd


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
