"""
Energy Metrics Database Models

SQLAlchemy models for energy consumption, mining site data,
and related metrics with TimescaleDB support.
"""

from datetime import datetime
from decimal import Decimal
from typing import Optional
from uuid import uuid4

from sqlalchemy import (
    BigInteger,
    Column,
    DateTime,
    Float,
    ForeignKey,
    Integer,
    String,
    Text,
    UUID,
    Boolean,
    Date,
    func,
)
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import relationship

from src.db.database import Base


class EnergyMetric(Base):
    """
    Energy consumption metrics for mining operations.

    This table is designed to be a TimescaleDB hypertable for
    efficient time-series queries.
    """

    __tablename__ = "energy_metrics"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)

    # Time dimension (TimescaleDB will use this as the primary dimension)
    time = Column(DateTime, primary_key=True, default=func.now())

    # Mining site reference
    site_id = Column(UUID(as_uuid=True), ForeignKey("mining_sites.id"), nullable=False)

    # Energy consumption
    consumption_kwh = Column(Float, nullable=False, comment="Energy consumed in kWh")
    hashrate_th = Column(Float, nullable=True, comment="Hashrate in TH/s")

    # Power source breakdown (JSON for flexibility)
    power_sources = Column(
        JSONB,
        nullable=True,
        comment="Breakdown of power sources used (e.g., {'hydro': 0.6, 'coal': 0.4})",
    )

    # Carbon and emissions
    co2_emission_grams = Column(
        Float,
        nullable=True,
        comment="CO2 emissions in grams calculated from power sources",
    )
    carbon_intensity_grams_per_kwh = Column(
        Float,
        nullable=True,
        comment="Carbon intensity of the energy mix",
    )

    # Grid interaction
    grid_import_kwh = Column(
        Float,
        default=0,
        comment="Energy imported from the grid",
    )
    grid_export_kwh = Column(
        Float,
        default=0,
        comment="Energy exported back to the grid (e.g., from renewable surplus)",
    )
    grid_strain_percent = Column(
        Float,
        nullable=True,
        comment="Percentage of local grid capacity being used",
    )

    # Efficiency metrics
    watts_per_th = Column(
        Float,
        nullable=True,
        comment="Power efficiency: Watts per Terahash",
    )
    efficiency_percent = Column(
        Float,
        nullable=True,
        comment="Overall mining efficiency as a percentage",
    )

    # Metadata
    device_count = Column(
        Integer,
        nullable=True,
        comment="Number of mining devices reporting",
    )
    uptime_percent = Column(
        Float,
        nullable=True,
        comment="Site uptime as a percentage",
    )
    metadata = Column(JSONB, default={}, comment="Additional metadata")

    # Timestamps
    recorded_at = Column(DateTime, default=func.now())
    created_at = Column(DateTime, default=func.now())

    # Relationships
    site = relationship("MiningSite", back_populates="energy_metrics")


class MiningSite(Base):
    """
    Mining site registration and configuration.
    """

    __tablename__ = "mining_sites"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)

    # Basic information
    operator_id = Column(
        UUID(as_uuid=True),
        ForeignKey("entities.id"),
        nullable=False,
        comment="Reference to the operating entity",
    )
    name = Column(String(255), nullable=False, comment="Site name")
    code = Column(String(50), unique=True, nullable=False, comment="Site code")

    # Location
    location = Column(
        JSONB,
        nullable=True,
        comment="Location data: {lat, lng, address, region, country}",
    )
    grid_connection_type = Column(
        String(50),
        nullable=True,
        comment="Type of grid connection (e.g., industrial, residential)",
    )

    # Power configuration
    max_power_capacity_kw = Column(
        Float,
        nullable=True,
        comment="Maximum power capacity in kW",
    )
    primary_power_source = Column(
        String(50),
        nullable=True,
        comment="Primary power source (e.g., hydro, solar, coal)",
    )
    backup_power_sources = Column(
        JSONB,
        nullable=True,
        comment="List of backup power sources",
    )

    # Equipment
    mining_equipment = Column(
        JSONB,
        nullable=True,
        comment="Mining equipment details",
    )
    total_hashrate_th = Column(
        Float,
        nullable=True,
        comment="Total installed hashrate capacity in TH/s",
    )

    # Compliance
    energy_permit_id = Column(
        String(100),
        nullable=True,
        comment="Energy regulatory permit ID",
    )
    green_energy_certified = Column(
        Boolean,
        default=False,
        comment="Whether site has green energy certification",
    )
    certification_expiry = Column(
        Date,
        nullable=True,
        comment="Green energy certification expiry date",
    )

    # Status
    status = Column(
        String(20),
        default="active",
        comment="Site status: active, inactive, suspended, decommissioned",
    )

    # Audit fields
    created_at = Column(DateTime, default=func.now())
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now())
    created_by = Column(UUID(as_uuid=True), nullable=True)
    updated_by = Column(UUID(as_uuid=True), nullable=True)

    # Relationships
    energy_metrics = relationship("EnergyMetric", back_populates="site", order_by=EnergyMetric.time)


class PowerSource(Base):
    """
    Power source configuration and carbon factors.
    """

    __tablename__ = "power_sources"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)

    code = Column(String(50), unique=True, nullable=False, comment="Power source code")
    name = Column(String(100), nullable=False, comment="Power source name")

    # Carbon factors
    carbon_factor_grams_per_kwh = Column(
        Float,
        nullable=False,
        comment="CO2 emissions per kWh in grams",
    )
    is_renewable = Column(Boolean, nullable=True, comment="Whether power source is renewable")

    # Metadata
    description = Column(Text, nullable=True)
    metadata = Column(JSONB, default={})

    # Status
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=func.now())


class Forecast(Base):
    """
    Generated forecasts for energy consumption and hashrate.
    """

    __tablename__ = "forecasts"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)

    # Forecast metadata
    site_id = Column(
        UUID(as_uuid=True),
        ForeignKey("mining_sites.id"),
        nullable=False,
    )
    forecast_type = Column(
        String(50),
        nullable=False,
        comment="Type of forecast: load, hashrate, carbon",
    )
    horizon_days = Column(Integer, nullable=False, comment="Forecast horizon in days")

    # Data
    forecast_data = Column(
        JSONB,
        nullable=False,
        comment="Forecast data with timestamps and predictions",
    )
    confidence_intervals = Column(
        JSONB,
        nullable=True,
        comment="Confidence intervals for predictions",
    )

    # Model metadata
    model_version = Column(String(20), nullable=True)
    training_data_points = Column(Integer, nullable=True)
    model_metrics = Column(
        JSONB,
        nullable=True,
        comment="Model performance metrics (MAE, RMSE, etc.)",
    )

    # Status
    status = Column(
        String(20),
        default="completed",
        comment="Status: pending, processing, completed, failed",
    )
    error_message = Column(Text, nullable=True)

    # Timestamps
    generated_at = Column(DateTime, default=func.now())
    valid_from = Column(DateTime, nullable=False)
    valid_to = Column(DateTime, nullable=False)
    next_retrain_at = Column(DateTime, nullable=True)

    # Relationships
    site = relationship("MiningSite")


class CarbonReport(Base):
    """
    Carbon footprint reports for mining operations.
    """

    __tablename__ = "carbon_reports"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)

    # Report metadata
    site_id = Column(
        UUID(as_uuid=True),
        ForeignKey("mining_sites.id"),
        nullable=False,
    )
    entity_id = Column(
        UUID(as_uuid=True),
        nullable=False,
        comment="Reference to the entity",
    )
    report_type = Column(
        String(50),
        nullable=False,
        comment="Report type: daily, monthly, quarterly, annual",
    )

    # Period
    period_start = Column(Date, nullable=False)
    period_end = Column(Date, nullable=False)

    # Carbon data
    total_consumption_kwh = Column(Float, nullable=False)
    total_co2_grams = Column(Float, nullable=False)
    average_carbon_intensity = Column(Float, nullable=True)
    renewable_percentage = Column(Float, nullable=True)

    # Breakdown by source
    source_breakdown = Column(
        JSONB,
        nullable=True,
        comment="Carbon breakdown by power source",
    )

    # Comparison with baseline
    baseline_comparison = Column(
        JSONB,
        nullable=True,
        comment="Comparison with historical baseline",
    )

    # Status
    status = Column(
        String(20),
        default="generated",
        comment="Status: draft, verified, finalized",
    )

    # Timestamps
    generated_at = Column(DateTime, default=func.now())
    verified_at = Column(DateTime, nullable=True)
    verified_by = Column(UUID(as_uuid=True), nullable=True)

    # Relationships
    site = relationship("MiningSite")


class SustainabilityScore(Base):
    """
    Sustainability scores for mining operations.
    """

    __tablename__ = "sustainability_scores"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid4)

    # Score metadata
    site_id = Column(
        UUID(as_uuid=True),
        ForeignKey("mining_sites.id"),
        nullable=False,
    )
    entity_id = Column(UUID(as_uuid=True), nullable=False)
    calculation_date = Column(Date, nullable=False)

    # Overall score
    overall_score = Column(Float, nullable=False, comment="Overall sustainability score (0-100)")
    grade = Column(String(5), nullable=True, comment="Letter grade: A+, A, B, C, D, F")

    # Component scores
    renewable_score = Column(Float, nullable=True)
    carbon_score = Column(Float, nullable=True)
    efficiency_score = Column(Float, nullable=True)
    grid_strain_score = Column(Float, nullable=True)

    # Component weights used
    weights_used = Column(
        JSONB,
        nullable=True,
        comment="Weights used in calculation",
    )

    # Comparison
    previous_score = Column(Float, nullable=True)
    score_change_percent = Column(Float, nullable=True)
    trend = Column(
        String(20),
        nullable=True,
        comment="Trend: improving, stable, declining",
    )

    # Status
    is_final = Column(Boolean, default=True)
    created_at = Column(DateTime, default=func.now())

    # Relationships
    site = relationship("MiningSite")
