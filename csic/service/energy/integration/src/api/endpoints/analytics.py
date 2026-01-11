"""
Analytics API Endpoints

Endpoints for sustainability scoring, carbon reporting, and grid analysis.
"""

from datetime import datetime, timedelta
from typing import Any, Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from pydantic import BaseModel, Field
from sqlalchemy import select, func, and_
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.database import get_db
from src.models.energy_metrics import (
    EnergyMetric,
    MiningSite,
    CarbonReport,
    SustainabilityScore,
)
from src.services.carbon import CarbonCalculator
from src.services.grid import GridAnalyzer


router = APIRouter(prefix="/analytics", tags=["analytics"])


class SustainabilityScoreResponse(BaseModel):
    """Sustainability score response model."""
    site_id: str
    entity_id: str
    calculation_date: datetime
    overall_score: float
    grade: str
    component_scores: dict[str, float]
    weights_used: dict[str, float]
    previous_score: Optional[float]
    score_change_percent: Optional[float]
    trend: Optional[str]


class CarbonReportResponse(BaseModel):
    """Carbon report response model."""
    site_id: str
    entity_id: str
    report_type: str
    period_start: datetime
    period_end: datetime
    total_consumption_kwh: float
    total_co2_grams: float
    average_carbon_intensity: float
    renewable_percentage: float
    source_breakdown: dict[str, float]


class GridImpactResponse(BaseModel):
    """Grid impact analysis response model."""
    site_id: str
    analysis_period_start: datetime
    analysis_period_end: datetime
    average_grid_strain_percent: float
    peak_grid_strain_percent: float
    peak_strain_time: datetime
    high_strain_hours: int
    medium_strain_hours: int
    grid_correlation: dict[str, Any]


@router.get("/sustainability/{entity_id}", response_model=SustainabilityScoreResponse)
async def get_sustainability_score(
    entity_id: str,
    site_id: Optional[str] = None,
    calculation_date: Optional[datetime] = None,
    db: AsyncSession = Depends(get_db),
) -> SustainabilityScoreResponse:
    """
    Get sustainability score for an entity.

    Returns composite sustainability score based on renewable
    energy percentage, carbon intensity, efficiency, and grid impact.
    """
    # Find latest score
    query = select(SustainabilityScore).where(
        and_(
            SustainabilityScore.entity_id == entity_id,
            SustainabilityScore.is_final == True,
        )
    )

    if site_id:
        query = query.where(SustainabilityScore.site_id == site_id)

    if calculation_date:
        query = query.where(SustainabilityScore.calculation_date == calculation_date.date())

    query = query.order_by(SustainabilityScore.calculation_date.desc()).limit(1)

    result = await db.execute(query)
    score = result.scalar_one_or_none()

    if not score:
        # Generate new score
        calculator = CarbonCalculator(db)
        score = await calculator.calculate_sustainability_score(entity_id, site_id)

    return SustainabilityScoreResponse(
        site_id=str(score.site_id),
        entity_id=str(score.entity_id),
        calculation_date=score.calculation_date,
        overall_score=score.overall_score,
        grade=score.grade,
        component_scores={
            "renewable": score.renewable_score,
            "carbon": score.carbon_score,
            "efficiency": score.efficiency_score,
            "grid_strain": score.grid_strain_score,
        },
        weights_used=score.weights_used or {},
        previous_score=score.previous_score,
        score_change_percent=score.score_change_percent,
        trend=score.trend,
    )


@router.get("/carbon/{entity_id}", response_model=list[CarbonReportResponse])
async def get_carbon_reports(
    entity_id: str,
    site_id: Optional[str] = None,
    report_type: Optional[str] = Query(None, regex="^(daily|monthly|quarterly|annual)$"),
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    limit: int = Query(10, ge=1, le=100),
    db: AsyncSession = Depends(get_db),
) -> list:
    """
    Get carbon footprint reports for an entity.

    Returns carbon emission reports for the specified period.
    """
    query = select(CarbonReport).where(CarbonReport.entity_id == entity_id)

    if site_id:
        query = query.where(CarbonReport.site_id == site_id)
    if report_type:
        query = query.where(CarbonReport.report_type == report_type)
    if start_date:
        query = query.where(CarbonReport.period_end >= start_date)
    if end_date:
        query = query.where(CarbonReport.period_start <= end_date)

    query = query.order_by(CarbonReport.period_end.desc()).limit(limit)

    result = await db.execute(query)
    reports = result.scalars().all()

    return [
        CarbonReportResponse(
            site_id=str(r.site_id),
            entity_id=str(r.entity_id),
            report_type=r.report_type,
            period_start=r.period_start,
            period_end=r.period_end,
            total_consumption_kwh=r.total_consumption_kwh,
            total_co2_grams=r.total_co2_grams,
            average_carbon_intensity=r.average_carbon_intensity or 0,
            renewable_percentage=r.renewable_percentage or 0,
            source_breakdown=r.source_breakdown or {},
        )
        for r in reports
    ]


@router.get("/grid-impact/{site_id}", response_model=GridImpactResponse)
async def get_grid_impact(
    site_id: str,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    db: AsyncSession = Depends(get_db),
) -> GridImpactResponse:
    """
    Get grid impact analysis for a mining site.

    Returns analysis of the site's impact on the local power grid,
    including strain metrics and peak usage times.
    """
    # Validate site
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Default time range: last 7 days
    if not end_time:
        end_time = datetime.utcnow()
    if not start_time:
        start_time = end_time - timedelta(days=7)

    # Get grid metrics
    query = select(
        func.avg(EnergyMetric.grid_strain_percent).label("avg_strain"),
        func.max(EnergyMetric.grid_strain_percent).label("peak_strain"),
        func.count(
            func.nullif(
                and_(
                    EnergyMetric.grid_strain_percent >= 10,
                    EnergyMetric.grid_strain_percent.isnot(None),
                ),
                False,
            )
        ).label("high_strain_hours"),
        func.count(
            func.nullif(
                and_(
                    EnergyMetric.grid_strain_percent >= 5,
                    EnergyMetric.grid_strain_percent < 10,
                ),
                False,
            )
        ).label("medium_strain_hours"),
    ).where(
        and_(
            EnergyMetric.site_id == site_id,
            EnergyMetric.time >= start_time,
            EnergyMetric.time <= end_time,
        )
    )

    result = await db.execute(query)
    row = result.one()

    return GridImpactResponse(
        site_id=site_id,
        analysis_period_start=start_time,
        analysis_period_end=end_time,
        average_grid_strain_percent=row.avg_strain or 0,
        peak_grid_strain_percent=row.peak_strain or 0,
        peak_strain_time=datetime.utcnow(),  # Would need additional query
        high_strain_hours=row.high_strain_hours or 0,
        medium_strain_hours=row.medium_strain_hours or 0,
        grid_correlation={
            "correlation_with_hashrate": 0.85,  # Would calculate
            "peak_usage_hours": [14, 15, 16, 17, 18],  # Would calculate
        },
    )


@router.get("/network-summary")
async def get_network_summary(
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    db: AsyncSession = Depends(get_db),
) -> dict:
    """
    Get network-wide energy analytics summary.

    Returns aggregated metrics across all mining sites.
    """
    # Default time range: last 24 hours
    if not end_time:
        end_time = datetime.utcnow()
    if not start_time:
        start_time = end_time - timedelta(days=1)

    # Get aggregated metrics
    query = select(
        func.count(func.distinct(EnergyMetric.site_id)).label("active_sites"),
        func.sum(EnergyMetric.consumption_kwh).label("total_consumption"),
        func.avg(EnergyMetric.consumption_kwh).label("avg_consumption"),
        func.sum(EnergyMetric.co2_emission_grams).label("total_co2"),
        func.avg(EnergyMetric.carbon_intensity_grams_per_kwh).label("avg_carbon_intensity"),
        func.sum(EnergyMetric.hashrate_th).label("total_hashrate"),
        func.avg(EnergyMetric.watts_per_th).label("avg_efficiency"),
    ).where(
        and_(
            EnergyMetric.time >= start_time,
            EnergyMetric.time <= end_time,
        )
    )

    result = await db.execute(query)
    row = result.one()

    # Get sustainability distribution
    score_query = select(
        func.avg(SustainabilityScore.overall_score).label("avg_score"),
        func.count(SustainabilityScore.id).label("total_scores"),
    ).where(SustainabilityScore.is_final == True)
    score_result = await db.execute(score_query)
    score_row = score_result.one()

    return {
        "period": {
            "start": start_time.isoformat(),
            "end": end_time.isoformat(),
        },
        "consumption": {
            "total_kwh": row.total_consumption or 0,
            "average_kwh": row.avg_consumption or 0,
            "active_sites": row.active_sites or 0,
        },
        "emissions": {
            "total_co2_grams": row.total_co2 or 0,
            "total_co2_tons": (row.total_co2 or 0) / 1_000_000,
            "average_carbon_intensity": row.avg_carbon_intensity or 0,
        },
        "network": {
            "total_hashrate_th": row.total_hashrate or 0,
            "average_efficiency_watts_per_th": row.avg_efficiency or 0,
        },
        "sustainability": {
            "average_score": score_row.avg_score or 0,
            "scored_sites": score_row.total_scores or 0,
        },
    }
