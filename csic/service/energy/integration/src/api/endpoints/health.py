"""
Health check endpoints for monitoring service status.
"""

from datetime import datetime
from typing import Any

from fastapi import APIRouter, Depends
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.database import get_db
from src.core.config import settings


router = APIRouter(tags=["health"])


@router.get("/health")
async def health_check() -> dict[str, Any]:
    """
    Basic health check endpoint.

    Returns service status and basic information.
    """
    return {
        "status": "healthy",
        "service": settings.app.name,
        "version": settings.app.version,
        "timestamp": datetime.utcnow().isoformat(),
    }


@router.get("/health/detailed")
async def detailed_health_check(db: AsyncSession = Depends(get_db)) -> dict[str, Any]:
    """
    Detailed health check including database connectivity.

    Returns comprehensive health status of all service dependencies.
    """
    health_status = {
        "status": "healthy",
        "service": settings.app.name,
        "version": settings.app.version,
        "timestamp": datetime.utcnow().isoformat(),
        "checks": {},
    }

    # Check database
    try:
        from sqlalchemy import text
        await db.execute(text("SELECT 1"))
        health_status["checks"]["database"] = {
            "status": "healthy",
            "type": "postgresql",
        }
    except Exception as e:
        health_status["status"] = "degraded"
        health_status["checks"]["database"] = {
            "status": "unhealthy",
            "type": "postgresql",
            "error": str(e),
        }

    # Check Redis
    try:
        import redis.asyncio as redis
        client = redis.Redis(
            host=settings.redis.host,
            port=settings.redis.port,
            password=settings.redis.password or None,
            db=settings.redis.db,
        )
        await client.ping()
        await client.close()
        health_status["checks"]["redis"] = {"status": "healthy"}
    except Exception as e:
        health_status["status"] = "degraded"
        health_status["checks"]["redis"] = {
            "status": "unhealthy",
            "error": str(e),
        }

    # Check Kafka
    try:
        from kafka import KafkaProducer
        producer = KafkaProducer(
            bootstrap_servers=settings.kafka.brokers,
            api_version_auto_timeout_ms=5000,
        )
        producer.close()
        health_status["checks"]["kafka"] = {"status": "healthy"}
    except Exception as e:
        health_status["status"] = "degraded"
        health_status["checks"]["kafka"] = {
            "status": "unhealthy",
            "error": str(e),
        }

    return health_status


@router.get("/ready")
async def readiness_check() -> dict[str, str]:
    """
    Kubernetes readiness probe endpoint.

    Returns 200 if service is ready to accept traffic.
    """
    return {"status": "ready"}


@router.get("/live")
async def liveness_check() -> dict[str, str]:
    """
    Kubernetes liveness probe endpoint.

    Returns 200 if service is alive.
    """
    return {"status": "alive"}
