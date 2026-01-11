"""
Forecasting API Endpoints

Endpoints for energy load and hashrate forecasting using ML models.
"""

from datetime import datetime, timedelta
from typing import Any, Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.database import get_db
from src.models.energy_metrics import MiningSite, Forecast
from src.services.forecasting import ForecastingService


router = APIRouter(prefix="/forecast", tags=["forecasting"])


class ForecastResponse(BaseModel):
    """Forecast response model."""
    site_id: str
    forecast_type: str
    horizon_days: int
    generated_at: datetime
    valid_from: datetime
    valid_to: datetime
    data: list[dict[str, Any]]
    confidence_intervals: Optional[list[dict[str, Any]]]
    model_info: dict[str, Any]


class LoadForecastParams(BaseModel):
    """Parameters for load forecast generation."""
    horizon_days: int = Field(30, ge=1, le=365, description="Forecast horizon in days")
    include_confidence: bool = Field(True, description="Include confidence intervals")
    recalculate: bool = Field(False, description="Force recalculation")


@router.get("/load/{site_id}", response_model=ForecastResponse)
async def get_load_forecast(
    site_id: str,
    horizon_days: int = Query(30, ge=1, le=365),
    include_confidence: bool = Query(True),
    recalculate: bool = Query(False),
    db: AsyncSession = Depends(get_db),
) -> ForecastResponse:
    """
    Get energy load forecast for a mining site.

    Returns predicted energy consumption for the specified
    horizon using Prophet-based time series forecasting.
    """
    # Validate site
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Get or generate forecast
    forecast_service = ForecastingService(db)

    forecast = await forecast_service.get_or_generate_forecast(
        site_id=site_id,
        forecast_type="load",
        horizon_days=horizon_days,
        recalculate=recalculate,
    )

    if not forecast:
        raise HTTPException(status_code=500, detail="Failed to generate forecast")

    return ForecastResponse(
        site_id=str(site_id),
        forecast_type=forecast.forecast_type,
        horizon_days=forecast.horizon_days,
        generated_at=forecast.generated_at,
        valid_from=forecast.valid_from,
        valid_to=forecast.valid_to,
        data=forecast.forecast_data,
        confidence_intervals=forecast.confidence_intervals if include_confidence else None,
        model_info={
            "version": forecast.model_version,
            "training_points": forecast.training_data_points,
            "metrics": forecast.model_metrics,
        },
    )


@router.get("/hashrate/{site_id}", response_model=ForecastResponse)
async def get_hashrate_forecast(
    site_id: str,
    horizon_days: int = Query(30, ge=1, le=365),
    include_confidence: bool = Query(True),
    recalculate: bool = Query(False),
    db: AsyncSession = Depends(get_db),
) -> ForecastResponse:
    """
    Get hashrate forecast for a mining site.

    Returns predicted hashrate values for the specified horizon.
    """
    # Validate site
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Get or generate forecast
    forecast_service = ForecastingService(db)

    forecast = await forecast_service.get_or_generate_forecast(
        site_id=site_id,
        forecast_type="hashrate",
        horizon_days=horizon_days,
        recalculate=recalculate,
    )

    if not forecast:
        raise HTTPException(status_code=500, detail="Failed to generate forecast")

    return ForecastResponse(
        site_id=str(site_id),
        forecast_type=forecast.forecast_type,
        horizon_days=forecast.horizon_days,
        generated_at=forecast.generated_at,
        valid_from=forecast.valid_from,
        valid_to=forecast.valid_to,
        data=forecast.forecast_data,
        confidence_intervals=forecast.confidence_intervals if include_confidence else None,
        model_info={
            "version": forecast.model_version,
            "training_points": forecast.training_data_points,
            "metrics": forecast.model_metrics,
        },
    )


@router.get("/carbon/{site_id}", response_model=ForecastResponse)
async def get_carbon_forecast(
    site_id: str,
    horizon_days: int = Query(30, ge=1, le=365),
    include_confidence: bool = Query(True),
    recalculate: bool = Query(False),
    db: AsyncSession = Depends(get_db),
) -> ForecastResponse:
    """
    Get carbon emission forecast for a mining site.

    Returns predicted carbon emissions based on energy forecast
    and power source mix.
    """
    # Validate site
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Get or generate forecast
    forecast_service = ForecastingService(db)

    forecast = await forecast_service.get_or_generate_forecast(
        site_id=site_id,
        forecast_type="carbon",
        horizon_days=horizon_days,
        recalculate=recalculate,
    )

    if not forecast:
        raise HTTPException(status_code=500, detail="Failed to generate forecast")

    return ForecastResponse(
        site_id=str(site_id),
        forecast_type=forecast.forecast_type,
        horizon_days=forecast.horizon_days,
        generated_at=forecast.generated_at,
        valid_from=forecast.valid_from,
        valid_to=forecast.valid_to,
        data=forecast.forecast_data,
        confidence_intervals=forecast.confidence_intervals if include_confidence else None,
        model_info={
            "version": forecast.model_version,
            "training_points": forecast.training_data_points,
            "metrics": forecast.model_metrics,
        },
    )


@router.post("/custom/{site_id}")
async def generate_custom_forecast(
    site_id: str,
    params: LoadForecastParams,
    db: AsyncSession = Depends(get_db),
) -> dict:
    """
    Generate a custom forecast with specified parameters.

    Allows customization of forecast horizon and confidence intervals.
    """
    # Validate site
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Generate forecast
    forecast_service = ForecastingService(db)

    forecast = await forecast_service.generate_forecast(
        site_id=site_id,
        forecast_type="load",
        horizon_days=params.horizon_days,
    )

    if not forecast:
        raise HTTPException(status_code=500, detail="Failed to generate forecast")

    return {
        "status": "generated",
        "forecast_id": str(forecast.id),
        "site_id": str(site_id),
        "forecast_type": forecast.forecast_type,
        "horizon_days": forecast.horizon_days,
        "generated_at": forecast.generated_at.isoformat(),
    }


@router.get("/model/status/{site_id}")
async def get_model_status(
    site_id: str,
    db: AsyncSession = Depends(get_db),
) -> dict:
    """
    Get forecast model status for a site.

    Returns information about model training status,
    last training time, and next scheduled retrain.
    """
    # Validate site
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    forecast_service = ForecastingService(db)
    status = await forecast_service.get_model_status(site_id)

    return status
