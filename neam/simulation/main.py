"""
NEAM Simulation Service - Main Application
FastAPI application for What-If Analysis and Scenario Simulation
"""

import logging
import sys
from contextlib import asynccontextmanager
from datetime import datetime

import uvicorn
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

from .routes import health_router, simulation_router
from .models import SimulationStatus


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler("/var/log/neam/simulation.log")
    ]
)

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler"""
    # Startup
    logger.info("Starting NEAM Simulation Service...")
    logger.info("Loading configuration...")
    logger.info("Initializing simulation engines...")
    
    yield
    
    # Shutdown
    logger.info("Shutting down NEAM Simulation Service...")
    logger.info("Cleaning up resources...")


# Create FastAPI application
app = FastAPI(
    title="NEAM Economic Simulation Service",
    description="""
    ## National Economic Activity Monitor - Simulation Engine
    
    This service provides advanced economic simulation capabilities for 
    government decision support and policy analysis.
    
    ### Available Simulations:
    
    * **Monte Carlo Simulation** - Probabilistic risk assessment using random sampling
    * **Economic Projection** - Time-series forecasting of economic indicators
    * **Policy Scenario Analysis** - Compare outcomes under different policy interventions
    * **Sensitivity Analysis** - Identify influential parameters using sensitivity indices
    * **Stress Testing** - Evaluate economic resilience under adverse scenarios
    
    ### Features:
    
    * Multiple probability distributions for Monte Carlo simulations
    * ARIMA and Exponential Smoothing forecasting models
    * Comprehensive policy intervention modeling
    * Sobol', Morris, and FAST sensitivity analysis methods
    * Real-time progress tracking
    * Batch simulation support
    """,
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json",
    lifespan=lifespan
)

# Configure CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure based on environment
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Exception handlers
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """Global exception handler"""
    logger.error(f"Unhandled exception: {str(exc)}", exc_info=True)
    
    return JSONResponse(
        status_code=500,
        content={
            "error": "Internal server error",
            "message": str(exc) if logger.isEnabledFor(logging.DEBUG) else "An unexpected error occurred",
            "timestamp": datetime.utcnow().isoformat()
        }
    )


# Request logging middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    """Log all incoming requests"""
    start_time = datetime.utcnow()
    
    try:
        response = await call_next(request)
        
        # Log request details
        process_time = (datetime.utcnow() - start_time).total_seconds()
        
        logger.info(
            f"{request.method} {request.url.path} - "
            f"Status: {response.status_code} - "
            f"Time: {process_time:.3f}s"
        )
        
        return response
        
    except Exception as e:
        logger.error(f"Request failed: {str(e)}")
        raise


# Include routers
app.include_router(health_router, prefix="/api/v1")
app.include_router(simulation_router, prefix="/api/v1")


# Root endpoint
@app.get("/")
async def root():
    """Root endpoint with service information"""
    return {
        "service": "NEAM Simulation Service",
        "version": "1.0.0",
        "status": "operational",
        "timestamp": datetime.utcnow().isoformat(),
        "endpoints": {
            "docs": "/docs",
            "health": "/api/v1/health",
            "simulations": "/api/v1/simulations"
        },
        "description": "Economic simulation and scenario analysis engine"
    }


# Legacy endpoints for backward compatibility
@app.get("/api")
async def api_root():
    """API root with version information"""
    return {
        "api_version": "v1",
        "service": "neam-simulation-service",
        "timestamp": datetime.utcnow().isoformat()
    }


@app.get("/api/v1")
async def api_v1_root():
    """API v1 root"""
    return {
        "version": "v1",
        "name": "NEAM Simulation API",
        "description": "Economic simulation and scenario analysis endpoints",
        "endpoints": {
            "health": "/api/v1/health",
            "simulations": "/api/v1/simulations",
            "monte-carlo": "/api/v1/simulations/monte-carlo",
            "economic-projection": "/api/v1/simulations/economic-projection",
            "policy-scenario": "/api/v1/simulations/policy-scenario",
            "sensitivity-analysis": "/api/v1/simulations/sensitivity-analysis",
            "stress-test": "/api/v1/simulations/stress-test"
        }
    }


if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8083,
        reload=False,
        log_level="info"
    )
