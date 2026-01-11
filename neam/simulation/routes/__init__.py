"""
NEAM Simulation Service - Routes Package
API route modules
"""

from .health import router as health_router
from .simulation import router as simulation_router

__all__ = ["health_router", "simulation_router"]
