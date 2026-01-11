"""
Carbon Calculation Service

Calculates carbon emissions and sustainability scores
for mining operations based on power source mix.
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Optional

from sqlalchemy import select, func, and_
from sqlalchemy.ext.asyncio import AsyncSession

from src.core.config import settings
from src.models.energy_metrics import (
    EnergyMetric,
    MiningSite,
    CarbonReport,
    SustainabilityScore,
)


logger = logging.getLogger(__name__)


class CarbonCalculator:
    """
    Carbon emission calculator and sustainability scorer.

    Calculates CO2 emissions based on power source mix and
    generates sustainability scores for mining operations.
    """

    # Carbon factors (grams CO2 per kWh)
    CARBON_FACTORS = {
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

    def __init__(self, db: AsyncSession):
        self.db = db
        self.weights = settings.sustainability.weights
        self.thresholds = settings.sustainability.thresholds

    async def calculate_emissions(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
    ) -> dict[str, float]:
        """
        Calculate total carbon emissions for a site.

        Args:
            site_id: Mining site ID
            start_time: Start of period
            end_time: End of period

        Returns:
            Dictionary with emission metrics
        """
        query = select(
            func.sum(EnergyMetric.consumption_kwh).label("total_consumption"),
            func.sum(EnergyMetric.co2_emission_grams).label("total_co2"),
            func.avg(EnergyMetric.carbon_intensity_grams_per_kwh).label("avg_intensity"),
        ).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
            )
        )

        result = await self.db.execute(query)
        row = result.one()

        total_consumption = row.total_consumption or 0
        total_co2 = row.total_co2 or 0
        avg_intensity = row.avg_intensity or 0

        return {
            "total_consumption_kwh": total_consumption,
            "total_co2_grams": total_co2,
            "total_co2_tons": total_co2 / 1_000_000,
            "average_carbon_intensity": avg_intensity,
        }

    async def calculate_sustainability_score(
        self,
        entity_id: str,
        site_id: Optional[str] = None,
    ) -> SustainabilityScore:
        """
        Calculate overall sustainability score for an entity.

        Args:
            entity_id: Entity ID
            site_id: Optional specific site ID

        Returns:
            SustainabilityScore record
        """
        # Get recent data (last 30 days)
        end_time = datetime.utcnow()
        start_time = end_time - timedelta(days=30)

        # Build query
        conditions = [
            EnergyMetric.time >= start_time,
            EnergyMetric.time <= end_time,
        ]

        if site_id:
            conditions.append(EnergyMetric.site_id == site_id)

        query = select(
            func.avg(
                func.coalesce(
                    func.jsonb_extract_path_text(
                        EnergyMetric.power_sources::jsonb, "hydro"
                    )::float * 0.6 +
                    func.jsonb_extract_path_text(
                        EnergyMetric.power_sources::jsonb, "solar"
                    )::float * 0.3 +
                    func.jsonb_extract_path_text(
                        EnergyMetric.power_sources::jsonb, "wind"
                    )::float * 0.1,
                    0
                )
            ).label("renewable_score"),
            func.avg(EnergyMetric.carbon_intensity_grams_per_kwh).label("carbon_score"),
            func.avg(EnergyMetric.efficiency_percent).label("efficiency_score"),
            func.avg(EnergyMetric.grid_strain_percent).label("grid_strain_score"),
        ).where(and_(*conditions))

        result = await self.db.execute(query)
        row = result.one()

        # Calculate component scores (normalized 0-100)
        renewable_score = self._normalize_score(
            getattr(row, "renewable_score", 0) or 0,
            100, 0, True
        )
        carbon_score = self._normalize_score(
            getattr(row, "carbon_score", 0) or 0,
            100, 500, True  # Lower is better
        )
        efficiency_score = self._normalize_score(
            getattr(row, "efficiency_score", 0) or 0,
            100, 20, True  # Higher is better
        )
        grid_strain_score = self._normalize_score(
            getattr(row, "grid_strain_score", 0) or 0,
            100, 20, True  # Lower is better
        )

        # Calculate weighted overall score
        overall_score = (
            renewable_score * self.weights.get("renewable_percentage", 0.4) +
            carbon_score * self.weights.get("carbon_intensity", 0.3) +
            efficiency_score * self.weights.get("efficiency", 0.2) +
            grid_strain_score * self.weights.get("grid_strain", 0.1)
        )

        # Get previous score for trend
        prev_score = await self._get_previous_score(entity_id, site_id)

        score_change = None
        trend = None
        if prev_score:
            score_change = ((overall_score - prev_score) / prev_score) * 100
            if score_change > 5:
                trend = "improving"
            elif score_change < -5:
                trend = "declining"
            else:
                trend = "stable"

        # Determine grade
        grade = self._calculate_grade(overall_score)

        # Create record
        sustainability_score = SustainabilityScore(
            site_id=site_id or entity_id,
            entity_id=entity_id,
            calculation_date=end_time.date(),
            overall_score=overall_score,
            grade=grade,
            renewable_score=renewable_score,
            carbon_score=carbon_score,
            efficiency_score=efficiency_score,
            grid_strain_score=grid_strain_score,
            weights_used=self.weights,
            previous_score=prev_score,
            score_change_percent=score_change,
            trend=trend,
            is_final=True,
        )

        self.db.add(sustainability_score)
        await self.db.commit()
        await self.db.refresh(sustainability_score)

        logger.info(
            f"Calculated sustainability score for entity {entity_id}: "
            f"{overall_score:.1f} ({grade})"
        )

        return sustainability_score

    async def generate_carbon_report(
        self,
        entity_id: str,
        site_id: str,
        report_type: str,
        period_start: datetime,
        period_end: datetime,
    ) -> CarbonReport:
        """
        Generate a carbon footprint report.

        Args:
            entity_id: Entity ID
            site_id: Site ID
            report_type: Report type (daily, monthly, quarterly, annual)
            period_start: Report period start
            period_end: Report period end

        Returns:
            CarbonReport record
        """
        # Calculate emissions
        emissions = await self.calculate_emissions(site_id, period_start, period_end)

        # Get power source breakdown
        source_breakdown = await self._get_source_breakdown(site_id, period_start, period_end)

        # Calculate renewable percentage
        renewable_pct = self._calculate_renewable_percentage(source_breakdown)

        # Create report
        report = CarbonReport(
            site_id=site_id,
            entity_id=entity_id,
            report_type=report_type,
            period_start=period_start.date(),
            period_end=period_end.date(),
            total_consumption_kwh=emissions["total_consumption_kwh"],
            total_co2_grams=emissions["total_co2_grams"],
            average_carbon_intensity=emissions["average_carbon_intensity"],
            renewable_percentage=renewable_pct,
            source_breakdown=source_breakdown,
            status="generated",
        )

        self.db.add(report)
        await self.db.commit()
        await self.db.refresh(report)

        logger.info(
            f"Generated {report_type} carbon report for site {site_id}: "
            f"{emissions['total_co2_tons']:.2f} tons CO2"
        )

        return report

    async def _get_source_breakdown(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
    ) -> dict[str, float]:
        """Get power source breakdown for a period."""
        query = select(EnergyMetric.power_sources).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
                EnergyMetric.power_sources.isnot(None),
            )
        )

        result = await self.db.execute(query)
        sources = result.scalars().all()

        if not sources:
            return {}

        # Aggregate source proportions
        totals = {}
        count = 0

        for source in sources:
            if source:
                for src, prop in source.items():
                    totals[src] = totals.get(src, 0) + prop
                count += 1

        if count == 0:
            return {}

        # Average proportions
        return {src: prop / count for src, prop in totals.items()}

    def _calculate_renewable_percentage(self, source_breakdown: dict[str, float]) -> float:
        """Calculate renewable energy percentage from source breakdown."""
        renewable_sources = {"hydro", "solar", "wind", "nuclear", "geothermal"}

        if not source_breakdown:
            return 0

        total = sum(source_breakdown.values())
        if total == 0:
            return 0

        renewable = sum(
            prop for src, prop in source_breakdown.items()
            if src in renewable_sources
        )

        return (renewable / total) * 100

    async def _get_previous_score(
        self,
        entity_id: str,
        site_id: Optional[str],
    ) -> Optional[float]:
        """Get the most recent sustainability score."""
        query = select(SustainabilityScore.overall_score).where(
            SustainabilityScore.entity_id == entity_id,
            SustainabilityScore.is_final == True,
        )

        if site_id:
            query = query.where(SustainabilityScore.site_id == site_id)

        query = query.order_by(SustainabilityScore.calculation_date.desc()).limit(1)

        result = await self.db.execute(query)
        row = result.scalar_one_or_none()

        return row

    def _normalize_score(
        self,
        value: float,
        max_val: float,
        min_val: float,
        higher_is_better: bool = True,
    ) -> float:
        """Normalize a score to 0-100 range."""
        # Handle edge cases
        if value <= min_val:
            return 100 if higher_is_better else 0
        if value >= max_val:
            return 0 if higher_is_better else 100

        # Linear interpolation
        normalized = (value - min_val) / (max_val - min_val) * 100

        if not higher_is_better:
            normalized = 100 - normalized

        return max(0, min(100, normalized))

    def _calculate_grade(self, score: float) -> str:
        """Calculate letter grade from score."""
        if score >= self.thresholds.get("excellent", 80):
            return "A"
        elif score >= self.thresholds.get("good", 60):
            return "B"
        elif score >= self.thresholds.get("moderate", 40):
            return "C"
        elif score >= self.thresholds.get("poor", 20):
            return "D"
        else:
            return "F"
