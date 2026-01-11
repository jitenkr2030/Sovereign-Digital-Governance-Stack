"""
NEAM Simulation Service - Health Routes
Health check and status endpoints
"""

import logging
from datetime import datetime
from typing import Any, Dict
from fastapi import APIRouter
from pydantic import BaseModel


logger = logging.getLogger(__name__)

router = APIRouter(tags=["Health"])


class HealthStatus(BaseModel):
    """Health status response model"""
    status: str
    service: str
    version: str
    timestamp: str
    uptime_seconds: float
    components: Dict[str, Any]


class ComponentHealth(BaseModel):
    """Individual component health status"""
    status: str
    latency_ms: float = 0.0
    message: str = "healthy"


# Service start time (set on module load)
service_start_time = datetime.utcnow()


@router.get("/health", response_model=HealthStatus)
async def health_check():
    """
    Comprehensive health check endpoint.
    
    Returns the health status of all service components
    including dependencies and internal state.
    """
    uptime = (datetime.utcnow() - service_start_time).total_seconds()
    
    # Check all components
    components = await _check_components()
    
    # Determine overall status
    all_healthy = all(
        c.get("status") == "healthy" 
        for c in components.values()
    )
    
    overall_status = "healthy" if all_healthy else "degraded"
    
    return HealthStatus(
        status=overall_status,
        service="neam-simulation-service",
        version="1.0.0",
        timestamp=datetime.utcnow().isoformat(),
        uptime_seconds=uptime,
        components=components
    )


@router.get("/health/live")
async def liveness_check():
    """
    Kubernetes liveness probe endpoint.
    
    Returns 200 if the service is alive and responding.
    """
    return {"status": "alive", "timestamp": datetime.utcnow().isoformat()}


@router.get("/health/ready")
async def readiness_check():
    """
    Kubernetes readiness probe endpoint.
    
    Returns 200 if the service is ready to accept traffic.
    Checks dependencies like Redis and databases.
    """
    ready, details = await _check_readiness()
    
    if ready:
        return {"status": "ready", "timestamp": datetime.utcnow().isoformat()}
    else:
        return {
            "status": "not_ready",
            "timestamp": datetime.utcnow().isoformat(),
            "details": details
        }


@router.get("/health/version")
async def version_check():
    """
    Version information endpoint.
    
    Returns service version and build information.
    """
    return {
        "service": "neam-simulation-service",
        "version": "1.0.0",
        "build": "2024.01.01",
        "api_version": "v1",
        "timestamp": datetime.utcnow().isoformat()
    }


@router.get("/metrics")
async def metrics_endpoint():
    """
    Basic metrics endpoint.
    
    Returns service metrics for monitoring.
    """
    uptime = (datetime.utcnow() - service_start_time).total_seconds()
    
    return {
        "uptime_seconds": uptime,
        "uptime_hours": uptime / 3600,
        "timestamp": datetime.utcnow().isoformat(),
        "service": "neam-simulation-service"
    }


async def _check_components() -> Dict[str, Any]:
    """Check the health of all service components"""
    components = {}
    
    # Check memory usage
    components["memory"] = await _check_memory()
    
    # Check CPU
    components["cpu"] = await _check_cpu()
    
    # Check dependencies
    components["redis"] = await _check_redis()
    components["database"] = await _check_database()
    
    return components


async def _check_memory() -> Dict[str, Any]:
    """Check memory usage"""
    try:
        import psutil
        process = psutil.Process()
        memory_info = process.memory_info()
        
        memory_mb = memory_info.rss / (1024 * 1024)
        
        if memory_mb < 500:
            status = "healthy"
        elif memory_mb < 1000:
            status = "degraded"
        else:
            status = "unhealthy"
        
        return {
            "status": status,
            "memory_mb": round(memory_mb, 2),
            "vms_mb": round(memory_info.vms / (1024 * 1024), 2),
            "percent": process.memory_percent()
        }
    except Exception as e:
        return {"status": "unknown", "error": str(e)}


async def _check_cpu() -> Dict[str, Any]:
    """Check CPU usage"""
    try:
        import psutil
        cpu_percent = psutil.cpu_percent(interval=1)
        
        status = "healthy" if cpu_percent < 80 else "degraded"
        
        return {
            "status": status,
            "percent": cpu_percent,
            "count": psutil.cpu_count(),
            "frequency_mhz": psutil.cpu_freq().current if psutil.cpu_freq() else None
        }
    except Exception as e:
        return {"status": "unknown", "error": str(e)}


async def _check_redis() -> Dict[str, Any]:
    """Check Redis connectivity"""
    try:
        import redis
        import os
        
        redis_host = os.getenv("REDIS_HOST", "10.112.2.4")
        redis_port = int(os.getenv("REDIS_PORT", "6379"))
        
        client = redis.Redis(
            host=redis_host,
            port=redis_port,
            socket_timeout=2,
            socket_connect_timeout=2
        )
        
        start_time = datetime.utcnow()
        client.ping()
        latency_ms = (datetime.utcnow() - start_time).total_seconds() * 1000
        
        return {
            "status": "healthy",
            "latency_ms": round(latency_ms, 2),
            "host": redis_host,
            "port": redis_port
        }
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e),
            "latency_ms": 0
        }


async def _check_database() -> Dict[str, Any]:
    """Check database connectivity"""
    try:
        import asyncio
        import asyncpg
        
        db_host = "10.112.2.4"
        db_port = 5432
        db_user = "neam_user"
        db_password = "neam_password"
        db_name = "neam_db"
        
        start_time = datetime.utcnow()
        
        conn = await asyncpg.connect(
            host=db_host,
            port=db_port,
            user=db_user,
            password=db_password,
            database=db_name,
            timeout=5
        )
        
        # Simple query to verify connection
        await conn.fetchval("SELECT 1")
        await conn.close()
        
        latency_ms = (datetime.utcnow() - start_time).total_seconds() * 1000
        
        return {
            "status": "healthy",
            "latency_ms": round(latency_ms, 2),
            "host": db_host,
            "port": db_port,
            "database": db_name
        }
    except Exception as e:
        return {
            "status": "unhealthy",
            "error": str(e)[:100],
            "latency_ms": 0
        }


async def _check_readiness() -> tuple:
    """
    Check if service is ready to accept traffic.
    
    Returns (is_ready, details)
    """
    details = {}
    all_ready = True
    
    # Check Redis
    redis_check = await _check_redis()
    details["redis"] = redis_check
    if redis_check.get("status") != "healthy":
        all_ready = False
    
    # Check database
    db_check = await _check_database()
    details["database"] = db_check
    if db_check.get("status") != "healthy":
        all_ready = False
    
    return all_ready, details
