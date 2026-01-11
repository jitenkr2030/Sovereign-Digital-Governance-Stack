"""
Energy Integration Service API Endpoints Package
"""

from src.api.endpoints.health import router as health
from src.api.endpoints.energy import router as energy
from src.api.endpoints.forecasting import router as forecasting
from src.api.endpoints.analytics import router as analytics
from src.api.endpoints.sites import router as sites

__all__ = ["health", "energy", "forecasting", "analytics", "sites"]
