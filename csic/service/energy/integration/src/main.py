"""
Energy Integration Service - Main Entry Point

A FastAPI-based microservice for mining energy analytics,
load forecasting, and carbon footprint tracking.
"""

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import make_asgi_app

from src.core.config import settings
from src.api.endpoints import energy, forecasting, analytics, sites, health
from src.db.database import init_db, close_db
from src.services.ingestion import TelemetryConsumer

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL),
    format=settings.LOG_FORMAT,
)
logger = logging.getLogger(__name__)

# Global telemetry consumer
telemetry_consumer: TelemetryConsumer | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager for startup and shutdown events."""
    global telemetry_consumer

    # Startup
    logger.info("Starting Energy Integration Service...")
    
    # Initialize database
    await init_db()
    logger.info("Database initialized")

    # Start telemetry consumer
    telemetry_consumer = TelemetryConsumer()
    await telemetry_consumer.start()
    logger.info("Telemetry consumer started")

    yield

    # Shutdown
    logger.info("Shutting down Energy Integration Service...")
    
    if telemetry_consumer:
        await telemetry_consumer.stop()
    await close_db()
    logger.info("Shutdown complete")


# Create FastAPI application
app = FastAPI(
    title="CSIC Energy Integration Service",
    description="""
    API for mining energy analytics, load forecasting, and carbon footprint tracking.
    
    ## Features
    - Real-time energy consumption monitoring
    - ML-powered load forecasting
    - Carbon footprint calculation
    - Grid integration analysis
    - Sustainability scoring
    
    ## API Version
    v1
    """,
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routers
app.include_router(health.router, prefix="/api/v1")
app.include_router(energy.router, prefix="/api/v1")
app.include_router(forecasting.router, prefix="/api/v1")
app.include_router(analytics.router, prefix="/api/v1")
app.include_router(sites.router, prefix="/api/v1")

# Prometheus metrics endpoint
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)


@app.get("/")
async def root():
    """Root endpoint with service information."""
    return {
        "service": "CSIC Energy Integration Service",
        "version": "1.0.0",
        "status": "running",
        "docs": "/docs",
        "metrics": "/metrics",
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "src.main:app",
        host=settings.SERVER_HOST,
        port=settings.SERVER_PORT,
        reload=settings.DEBUG,
    )
