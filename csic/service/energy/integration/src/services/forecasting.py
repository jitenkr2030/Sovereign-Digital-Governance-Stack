"""
Forecasting Service

ML-based forecasting for energy consumption and hashrate
using Prophet and Scikit-learn models.
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Optional
from uuid import uuid4

import numpy as np
import pandas as pd
from prophet import Prophet
from sqlalchemy import select, func
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.config import settings
from src.models.energy_metrics import EnergyMetric, Forecast, MiningSite


logger = logging.getLogger(__name__)


class ForecastingService:
    """
    Forecasting service using Prophet for time-series prediction.

    Provides energy load, hashrate, and carbon emission forecasts
    with configurable horizons and confidence intervals.
    """

    def __init__(self, db: AsyncSession):
        self.db = db
        self.prophet_config = settings.forecasting.prophet

    async def get_or_generate_forecast(
        self,
        site_id: str,
        forecast_type: str,
        horizon_days: int = 30,
        recalculate: bool = False,
    ) -> Optional[Forecast]:
        """
        Get existing forecast or generate a new one.

        Args:
            site_id: Mining site ID
            forecast_type: Type of forecast (load, hashrate, carbon)
            horizon_days: Forecast horizon in days
            recalculate: Force recalculation even if valid forecast exists

        Returns:
            Forecast record or None if generation fails
        """
        # Check for existing valid forecast
        if not recalculate:
            existing = await self._get_existing_forecast(site_id, forecast_type, horizon_days)
            if existing:
                return existing

        # Generate new forecast
        forecast = await self.generate_forecast(site_id, forecast_type, horizon_days)
        return forecast

    async def generate_forecast(
        self,
        site_id: str,
        forecast_type: str,
        horizon_days: int,
    ) -> Optional[Forecast]:
        """
        Generate a new forecast using Prophet.

        Args:
            site_id: Mining site ID
            forecast_type: Type of forecast
            horizon_days: Forecast horizon

        Returns:
            Generated forecast record
        """
        try:
            # Get historical data
            historical_data = await self._get_historical_data(site_id, forecast_type)

            if len(historical_data) < settings.forecasting.min_training_data_days:
                logger.warning(
                    f"Insufficient data for forecast: {len(historical_data)} points, "
                    f"need {settings.forecasting.min_training_data_days}"
                )
                return None

            # Train Prophet model
            model = self._train_prophet_model(historical_data, forecast_type)

            # Generate forecast
            forecast_df = model.predict(
                model.make_future_dataframe(periods=horizon_days * 24, freq="H")
            )

            # Process forecast data
            forecast_data, confidence_intervals = self._process_forecast_output(
                forecast_df, horizon_days
            )

            # Calculate model metrics
            model_metrics = self._calculate_model_metrics(model, historical_data)

            # Create forecast record
            valid_from = datetime.utcnow()
            valid_to = valid_from + timedelta(days=horizon_days)

            forecast = Forecast(
                site_id=site_id,
                forecast_type=forecast_type,
                horizon_days=horizon_days,
                forecast_data=forecast_data,
                confidence_intervals=confidence_intervals,
                model_version="1.0.0",
                training_data_points=len(historical_data),
                model_metrics=model_metrics,
                status="completed",
                generated_at=valid_from,
                valid_from=valid_from,
                valid_to=valid_to,
                next_retrain_at=valid_from + timedelta(hours=settings.forecasting.retrain_interval_hours),
            )

            self.db.add(forecast)
            await self.db.commit()
            await self.db.refresh(forecast)

            logger.info(f"Generated {forecast_type} forecast for site {site_id}")
            return forecast

        except Exception as e:
            logger.error(f"Failed to generate forecast: {e}")
            # Create failed forecast record
            forecast = Forecast(
                site_id=site_id,
                forecast_type=forecast_type,
                horizon_days=horizon_days,
                forecast_data=[],
                status="failed",
                error_message=str(e),
                generated_at=datetime.utcnow(),
                valid_from=datetime.utcnow(),
                valid_to=datetime.utcnow() + timedelta(days=horizon_days),
            )
            self.db.add(forecast)
            await self.db.commit()
            return forecast

    async def _get_historical_data(
        self,
        site_id: str,
        forecast_type: str,
    ) -> pd.DataFrame:
        """
        Get historical data for forecast training.

        Returns DataFrame formatted for Prophet.
        """
        # Get data for last 90 days
        start_time = datetime.utcnow() - timedelta(days=90)

        query = select(EnergyMetric).where(
            EnergyMetric.site_id == site_id,
            EnergyMetric.time >= start_time,
        ).order_by(EnergyMetric.time)

        result = await self.db.execute(query)
        metrics = result.scalars().all()

        # Convert to DataFrame
        data = []
        for m in metrics:
            if forecast_type == "load":
                value = m.consumption_kwh
            elif forecast_type == "hashrate":
                value = m.hashrate_th or 0
            elif forecast_type == "carbon":
                value = m.co2_emission_grams or 0
            else:
                value = m.consumption_kwh

            data.append({
                "ds": m.time,
                "y": value,
            })

        return pd.DataFrame(data)

    def _train_prophet_model(
        self,
        historical_data: pd.DataFrame,
        forecast_type: str,
    ) -> Prophet:
        """Train Prophet model with configured parameters."""
        model = Prophet(
            yearly_seasonality=self.prophet_config.get("yearly_seasonality", True),
            weekly_seasonality=self.prophet_config.get("weekly_seasonality", True),
            daily_seasonality=self.prophet_config.get("daily_seasonality", False),
            seasonality_mode=self.prophet_config.get("seasonality_mode", "multiplicative"),
            changepoint_prior_scale=self.prophet_config.get("changepoint_prior_scale", 0.05),
        )

        # Add monthly seasonality for load patterns
        if forecast_type == "load":
            model.add_seasonality(
                name="monthly",
                period=30.5,
                fourier_order=5,
            )

        model.fit(historical_data)
        return model

    def _process_forecast_output(
        self,
        forecast_df: pd.DataFrame,
        horizon_days: int,
    ) -> tuple[list[dict], list[dict]]:
        """
        Process Prophet forecast output into API-friendly format.

        Returns forecast data and confidence intervals.
        """
        # Filter to future dates only
        now = datetime.utcnow()
        future = forecast_df[forecast_df["ds"] > now].head(horizon_days * 24)

        forecast_data = []
        confidence_intervals = []

        for _, row in future.iterrows():
            forecast_data.append({
                "timestamp": row["ds"].isoformat(),
                "value": round(row["yhat"], 2),
            })

            confidence_intervals.append({
                "timestamp": row["ds"].isoformat(),
                "lower": round(row["yhat_lower"], 2),
                "upper": round(row["yhat_upper"], 2),
            })

        return forecast_data, confidence_intervals

    def _calculate_model_metrics(
        self,
        model: Prophet,
        historical_data: pd.DataFrame,
    ) -> dict[str, float]:
        """Calculate model performance metrics."""
        predictions = model.predict(historical_data[["ds"]])
        actual = historical_data["y"].values
        predicted = predictions["yhat"].values

        # Mean Absolute Error
        mae = np.mean(np.abs(actual - predicted))

        # Root Mean Square Error
        rmse = np.sqrt(np.mean((actual - predicted) ** 2))

        # Mean Absolute Percentage Error
        mape = np.mean(np.abs((actual - predicted) / np.maximum(actual, 1))) * 100

        return {
            "mae": round(mae, 2),
            "rmse": round(rmse, 2),
            "mape": round(mape, 2),
        }

    async def _get_existing_forecast(
        self,
        site_id: str,
        forecast_type: str,
        horizon_days: int,
    ) -> Optional[Forecast]:
        """Get existing valid forecast."""
        query = select(Forecast).where(
            Forecast.site_id == site_id,
            Forecast.forecast_type == forecast_type,
            Forecast.horizon_days == horizon_days,
            Forecast.status == "completed",
            Forecast.valid_to >= datetime.utcnow(),
        ).order_by(Forecast.generated_at.desc())

        result = await self.db.execute(query)
        return result.scalar_one_or_none()

    async def get_model_status(self, site_id: str) -> dict[str, Any]:
        """Get forecast model status for a site."""
        # Get latest forecasts
        query = select(Forecast).where(
            Forecast.site_id == site_id,
        ).order_by(Forecast.generated_at.desc()).limit(3)

        result = await self.db.execute(query)
        forecasts = result.scalars().all()

        if not forecasts:
            return {
                "site_id": site_id,
                "status": "no_model",
                "message": "No forecasts have been generated",
            }

        latest = forecasts[0]
        retrain_needed = (
            latest.next_retrain_at and datetime.utcnow() >= latest.next_retrain_at
        )

        return {
            "site_id": site_id,
            "status": "active" if not retrain_needed else "needs_retrain",
            "latest_forecast": {
                "type": latest.forecast_type,
                "horizon_days": latest.horizon_days,
                "generated_at": latest.generated_at.isoformat(),
                "valid_to": latest.valid_to.isoformat(),
            },
            "model_metrics": latest.model_metrics,
            "retrain_needed": retrain_needed,
            "next_retrain_at": latest.next_retrain_at.isoformat() if latest.next_retrain_at else None,
            "forecast_history": [
                {
                    "type": f.forecast_type,
                    "generated_at": f.generated_at.isoformat(),
                    "status": f.status,
                }
                for f in forecasts
            ],
        }

    async def detect_anomalies(
        self,
        site_id: str,
        window_hours: int = 24,
    ) -> list[dict[str, Any]]:
        """
        Detect anomalies in recent energy data.

        Uses z-score based anomaly detection.
        """
        start_time = datetime.utcnow() - timedelta(hours=window_hours)

        query = select(EnergyMetric).where(
            EnergyMetric.site_id == site_id,
            EnergyMetric.time >= start_time,
        ).order_by(EnergyMetric.time)

        result = await self.db.execute(query)
        metrics = result.scalars().all()

        if len(metrics) < 3:
            return []

        # Calculate z-scores
        values = [m.consumption_kwh for m in metrics]
        mean_val = np.mean(values)
        std_val = np.std(values)

        if std_val == 0:
            return []

        z_threshold = settings.forecasting.anomaly_detection.get("z_threshold", 3.0)
        anomalies = []

        for m in metrics:
            z_score = abs((m.consumption_kwh - mean_val) / std_val)

            if z_score > z_threshold:
                anomalies.append({
                    "timestamp": m.time.isoformat(),
                    "value": m.consumption_kwh,
                    "z_score": round(z_score, 2),
                    "type": "consumption_spike" if m.consumption_kwh > mean_val else "consumption_drop",
                })

        return anomalies
