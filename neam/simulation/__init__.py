"""
NEAM Platform - Python Simulation Service
What-If Analysis and Scenario Simulation Engine

This service provides economic simulation capabilities including:
- Monte Carlo simulations for risk assessment
- Economic model projections
- Policy intervention scenarios
- Sensitivity analysis
"""

from .main import app
from .models import (
    SimulationRequest,
    SimulationResponse,
    ScenarioInput,
    MonteCarloResult,
    SensitivityAnalysis,
    EconomicProjection,
)
from .engines import (
    MonteCarloEngine,
    EconomicModelEngine,
    ScenarioEngine,
    SensitivityEngine,
)

__version__ = "1.0.0"
__all__ = [
    "app",
    "SimulationRequest",
    "SimulationResponse",
    "ScenarioInput",
    "MonteCarloResult",
    "SensitivityAnalysis",
    "EconomicProjection",
    "MonteCarloEngine",
    "EconomicModelEngine",
    "ScenarioEngine",
    "SensitivityEngine",
]
