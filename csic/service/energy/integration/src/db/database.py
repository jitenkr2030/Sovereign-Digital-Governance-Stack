"""
Database connection and session management.

Provides async database engine and session factory.
"""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.orm import declarative_base

from src.core.config import settings

# Create async engine
engine: AsyncEngine = create_async_engine(
    settings.database.async_url,
    echo=settings.database.echo,
    pool_size=settings.database.pool_size,
    max_overflow=settings.database.max_overflow,
)

# Session factory
async_session_factory = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False,
    autocommit=False,
    autoflush=False,
)

# Base class for models
Base = declarative_base()


async def init_db() -> None:
    """Initialize database tables."""
    # Import models to ensure they are registered
    from src.models import energy_metrics, mining_sites  # noqa: F401

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

    # Enable TimescaleDB hypertables if enabled
    if settings.database.timescaledb.get("enabled", False):
        await enable_timescale_hypertables()


async def enable_timescale_hypertables() -> None:
    """Enable TimescaleDB hypertables for time-series data."""
    from src.models.energy_metrics import EnergyMetric

    async with engine.begin() as conn:
        # Convert to hypertable if not already
        await conn.execute(
            f"SELECT create_hypertable('{EnergyMetric.__tablename__}', 'time')"
        )


async def close_db() -> None:
    """Close database connections."""
    await engine.dispose()


@asynccontextmanager
async def get_session() -> AsyncGenerator[AsyncSession, None]:
    """Get database session context manager."""
    async with async_session_factory() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()


async def get_db() -> AsyncGenerator[AsyncSession, None]:
    """Dependency for getting database session."""
    async with async_session_factory() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()
