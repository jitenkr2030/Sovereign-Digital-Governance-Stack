"""
Energy Metrics API Endpoints

Endpoints for energy consumption monitoring and telemetry ingestion.
"""

from datetime import datetime
from typing import Any, List, Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from pydantic import BaseModel, Field
from sqlalchemy import select, func, and_, or_
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.database import get_db
from src.models.energy_metrics import EnergyMetric, MiningSite
from src.core.config import settings


router = APIRouter(prefix="/ingest", tags=["energy"])


# Pydantic models for request/response
class EnergyReadingBase(BaseModel):
    """Base model for energy readings."""
    consumption_kwh: float = Field(..., ge=0, description="Energy consumed in kWh")
    hashrate_th: Optional[float] = Field(None, ge=0, description="Hashrate in TH/s")
    power_sources: Optional[dict[str, float]] = Field(
        None,
        description="Power source breakdown (source: proportion)",
    )
    grid_import_kwh: float = Field(0, ge=0, description="Energy imported from grid")
    grid_export_kwh: float = Field(0, ge=0, description="Energy exported to grid")
    watts_per_th: Optional[float] = Field(None, ge=0, description="Watts per Terahash")
    device_count: Optional[int] = Field(None, ge=0, description="Number of devices")
    uptime_percent: Optional[float] = Field(None, ge=0, le=100, description="Uptime percentage")
    metadata: Optional[dict[str, Any]] = Field(None, description="Additional metadata")


class TelemetryPayload(BaseModel):
    """Single telemetry reading payload."""
    site_id: str = Field(..., description="Mining site ID")
    timestamp: Optional[datetime] = Field(None, description="Reading timestamp")
    reading: EnergyReadingBase


class BatchTelemetryPayload(BaseModel):
    """Batch telemetry reading payload."""
    readings: List[TelemetryPayload] = Field(..., max_length=1000)


class EnergyReadingResponse(BaseModel):
    """Response model for energy reading."""
    id: str
    site_id: str
    time: datetime
    consumption_kwh: float
    co2_emission_grams: Optional[float]
    watts_per_th: Optional[float]
    efficiency_percent: Optional[float]


class AggregatedMetrics(BaseModel):
    """Aggregated energy metrics response."""
    period_start: datetime
    period_end: datetime
    total_consumption_kwh: float
    average_consumption_kwh: float
    peak_consumption_kwh: float
    total_co2_grams: float
    average_carbon_intensity: float
    average_watts_per_th: float
    total_hashrate_th: float
    data_points: int


@router.post("/realtime", response_model=dict)
async def ingest_realtime_telemetry(
    payload: TelemetryPayload,
    db: AsyncSession = Depends(get_db),
) -> dict:
    """
    Ingest a single real-time energy telemetry reading.

    Processes the reading and calculates derived metrics
    including CO2 emissions and efficiency scores.
    """
    # Validate site exists
    site = await db.get(MiningSite, payload.site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Calculate CO2 emissions
    co2_grams = calculate_co2_emissions(
        payload.reading.consumption_kwh,
        payload.reading.power_sources,
    )

    # Calculate efficiency
    efficiency = calculate_efficiency(
        payload.reading.watts_per_th,
        payload.reading.hashrate_th,
        payload.reading.consumption_kwh,
    )

    # Create metric record
    metric = EnergyMetric(
        site_id=payload.site_id,
        time=payload.timestamp or datetime.utcnow(),
        consumption_kwh=payload.reading.consumption_kwh,
        hashrate_th=payload.reading.hashrate_th,
        power_sources=payload.reading.power_sources,
        co2_emission_grams=co2_grams,
        grid_import_kwh=payload.reading.grid_import_kwh,
        grid_export_kwh=payload.reading.grid_export_kwh,
        watts_per_th=payload.reading.watts_per_th,
        efficiency_percent=efficiency,
        device_count=payload.reading.device_count,
        uptime_percent=payload.reading.uptime_percent,
        metadata=payload.reading.metadata,
    )

    db.add(metric)
    await db.commit()
    await db.refresh(metric)

    return {
        "id": str(metric.id),
        "status": "ingested",
        "site_id": str(payload.site_id),
        "consumption_kwh": metric.consumption_kwh,
        "co2_emission_grams": metric.co2_emission_grams,
    }


@router.post("/telemetry", response_model=dict)
async def ingest_batch_telemetry(
    payload: BatchTelemetryPayload,
    db: AsyncSession = Depends(get_db),
) -> dict:
    """
    Ingest batch energy telemetry data.

    Accepts up to 1000 readings per batch for efficient
    bulk ingestion of historical or aggregated data.
    """
    metrics = []
    errors = []

    for i, reading in enumerate(payload.readings):
        try:
            # Validate site
            site = await db.get(MiningSite, reading.site_id)
            if not site:
                errors.append({"index": i, "error": "Site not found"})
                continue

            # Calculate derived metrics
            co2_grams = calculate_co2_emissions(
                reading.reading.consumption_kwh,
                reading.reading.power_sources,
            )

            efficiency = calculate_efficiency(
                reading.reading.watts_per_th,
                reading.reading.hashrate_th,
                reading.reading.consumption_kwh,
            )

            metric = EnergyMetric(
                site_id=reading.site_id,
                time=reading.timestamp or datetime.utcnow(),
                consumption_kwh=reading.reading.consumption_kwh,
                hashrate_th=reading.reading.hashrate_th,
                power_sources=reading.reading.power_sources,
                co2_emission_grams=co2_grams,
                grid_import_kwh=reading.reading.grid_import_kwh,
                grid_export_kwh=reading.reading.grid_export_kwh,
                watts_per_th=reading.reading.watts_per_th,
                efficiency_percent=efficiency,
                device_count=reading.reading.device_count,
                uptime_percent=reading.reading.uptime_percent,
                metadata=reading.reading.metadata,
            )
            metrics.append(metric)

        except Exception as e:
            errors.append({"index": i, "error": str(e)})

    # Bulk insert
    if metrics:
        db.add_all(metrics)
        await db.commit()

    return {
        "status": "completed",
        "ingested": len(metrics),
        "failed": len(errors),
        "errors": errors[:10],  # Limit error details
    }


@router.get("/site/{site_id}", response_model=list[EnergyReadingResponse])
async def get_site_metrics(
    site_id: str,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    limit: int = Query(100, ge=1, le=10000),
    db: AsyncSession = Depends(get_db),
) -> list:
    """
    Get energy metrics for a specific site.

    Returns time-series data for the specified time range,
    ordered by timestamp descending.
    """
    # Validate site exists
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Build query
    query = select(EnergyMetric).where(EnergyMetric.site_id == site_id)

    if start_time:
        query = query.where(EnergyMetric.time >= start_time)
    if end_time:
        query = query.where(EnergyMetric.time <= end_time)

    query = query.order_by(EnergyMetric.time.desc()).limit(limit)

    result = await db.execute(query)
    metrics = result.scalars().all()

    return [
        EnergyReadingResponse(
            id=str(m.id),
            site_id=str(m.site_id),
            time=m.time,
            consumption_kwh=m.consumption_kwh,
            co2_emission_grams=m.co2_emission_grams,
            watts_per_th=m.watts_per_th,
            efficiency_percent=m.efficiency_percent,
        )
        for m in metrics
    ]


@router.get("/aggregate", response_model=AggregatedMetrics)
async def get_aggregated_metrics(
    site_id: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    interval: str = Query("1h", regex="^(1h|6h|24h|7d|30d)$"),
    db: AsyncSession = Depends(get_db),
) -> AggregatedMetrics:
    """
    Get aggregated energy metrics across sites.

    Returns summarized metrics for analysis and reporting.
    """
    # Default time range: last 24 hours
    if not end_time:
        end_time = datetime.utcnow()
    if not start_time:
        start_time = end_time - timedelta(hours=24)

    # Build query
    conditions = []
    if site_id:
        conditions.append(EnergyMetric.site_id == site_id)
    if start_time:
        conditions.append(EnergyMetric.time >= start_time)
    if end_time:
        conditions.append(EnergyMetric.time <= end_time)

    query = select(
        func.min(EnergyMetric.time).label("period_start"),
        func.max(EnergyMetric.time).label("period_end"),
        func.sum(EnergyMetric.consumption_kwh).label("total_consumption"),
        func.avg(EnergyMetric.consumption_kwh).label("avg_consumption"),
        func.max(EnergyMetric.consumption_kwh).label("peak_consumption"),
        func.sum(EnergyMetric.co2_emission_grams).label("total_co2"),
        func.avg(EnergyMetric.carbon_intensity_grams_per_kwh).label("avg_carbon_intensity"),
        func.avg(EnergyMetric.watts_per_th).label("avg_watts_per_th"),
        func.sum(EnergyMetric.hashrate_th).label("total_hashrate"),
        func.count().label("data_points"),
    )

    if conditions:
        query = query.where(and_(*conditions))

    result = await db.execute(query)
    row = result.one()

    return AggregatedMetrics(
        period_start=row.period_start,
        period_end=row.period_end,
        total_consumption_kwh=row.total_consumption or 0,
        average_consumption_kwh=row.avg_consumption or 0,
        peak_consumption_kwh=row.peak_consumption or 0,
        total_co2_grams=row.total_co2 or 0,
        average_carbon_intensity=row.avg_carbon_intensity or 0,
        average_watts_per_th=row.avg_watts_per_th or 0,
        total_hashrate_th=row.total_hashrate or 0,
        data_points=row.data_points or 0,
    )


# Helper functions
def calculate_co2_emissions(
    consumption_kwh: float,
    power_sources: Optional[dict[str, float]],
) -> float:
    """Calculate CO2 emissions based on power source mix."""
    if not power_sources:
        # Default to mixed grid average
        return consumption_kwh * 350  # Approximate global average

    total_emissions = 0.0
    for source, proportion in power_sources.items():
        carbon_factor = get_carbon_factor(source)
        total_emissions += consumption_kwh * proportion * carbon_factor

    return total_emissions


def get_carbon_factor(source_code: str) -> float:
    """Get carbon factor for a power source."""
    factors = {
        "hydro": 24,
        "solar": 41,
        "wind": 11,
        "nuclear": 12,
        "geothermal": 38,
        "natural_gas": 450,
        "coal": 820,
        "oil": 650,
        "mixed": 350,
    }
    return factors.get(source_code, 350)


def calculate_efficiency(
    watts_per_th: Optional[float],
    hashrate_th: Optional[float],
    consumption_kwh: float,
) -> Optional[float]:
    """Calculate mining efficiency percentage."""
    if watts_per_th and hashrate_th and hashrate_th > 0:
        # Theoretical minimum: 0 J/TH (ideal), typical: 20-40 J/TH
        # Efficiency as percentage of ideal baseline
        theoretical_min = hashrate_th * 20  # 20 J/TH as baseline
        actual = watts_per_th * hashrate_th * 0.001  # Convert W to kW
        if actual > 0:
            return min(100, (theoretical_min / actual) * 100)
    return None


# Import required modules
from datetime import timedelta
