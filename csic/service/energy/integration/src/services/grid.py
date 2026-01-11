"""
Grid Analysis Service

Analyzes mining site impact on local power grid including
strain metrics, peak usage patterns, and grid correlation.
"""

import logging
from datetime import datetime, timedelta
from typing import Any, Optional

from sqlalchemy import select, func, and_
from sqlalchemy.ext.asyncio import AsyncSession

from src.models.energy_metrics import EnergyMetric


logger = logging.getLogger(__name__)


class GridAnalyzer:
    """
    Grid impact analyzer for mining operations.

    Provides analysis of how mining operations impact local
    power grid conditions including strain metrics and peak usage.
    """

    # Strain thresholds (percentage of local grid capacity)
    HIGH_STRAIN_THRESHOLD = 10.0
    MEDIUM_STRAIN_THRESHOLD = 5.0

    def __init__(self, db: AsyncSession):
        self.db = db

    async def analyze_grid_impact(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
    ) -> dict[str, Any]:
        """
        Analyze grid impact for a mining site.

        Args:
            site_id: Mining site ID
            start_time: Analysis period start
            end_time: Analysis period end

        Returns:
            Dictionary with grid impact analysis
        """
        # Get grid metrics
        query = select(
            func.avg(EnergyMetric.grid_strain_percent).label("avg_strain"),
            func.max(EnergyMetric.grid_strain_percent).label("peak_strain"),
            func.min(EnergyMetric.grid_strain_percent).label("min_strain"),
            func.stddev(EnergyMetric.grid_strain_percent).label("std_strain"),
            func.count(EnergyMetric.id).label("data_points"),
        ).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
                EnergyMetric.grid_strain_percent.isnot(None),
            )
        )

        result = await self.db.execute(query)
        row = result.one()

        # Calculate strain hours
        high_strain_hours = await self._count_strain_hours(
            site_id, start_time, end_time, self.HIGH_STRAIN_THRESHOLD, 100
        )
        medium_strain_hours = await self._count_strain_hours(
            site_id, start_time, end_time, self.MEDIUM_STRAIN_THRESHOLD, self.HIGH_STRAIN_THRESHOLD
        )

        # Get peak strain time
        peak_time = await self._get_peak_strain_time(site_id, start_time, end_time)

        # Analyze peak usage hours
        peak_usage_hours = await self._analyze_peak_hours(site_id, start_time, end_time)

        # Calculate grid correlation with hashrate
        correlation = await self._calculate_grid_hashrate_correlation(
            site_id, start_time, end_time
        )

        return {
            "site_id": site_id,
            "analysis_period": {
                "start": start_time.isoformat(),
                "end": end_time.isoformat(),
            },
            "strain_metrics": {
                "average_percent": round(row.avg_strain or 0, 2),
                "peak_percent": round(row.peak_strain or 0, 2),
                "minimum_percent": round(row.min_strain or 0, 2),
                "std_deviation": round(row.std_strain or 0, 2),
                "data_points": row.data_points or 0,
            },
            "strain_hours": {
                "high_strain": high_strain_hours,
                "medium_strain": medium_strain_hours,
                "total_hours": (end_time - start_time).total_seconds() / 3600,
            },
            "peak_strain_time": peak_time.isoformat() if peak_time else None,
            "peak_usage_hours": peak_usage_hours,
            "grid_correlation": correlation,
            "risk_assessment": self._assess_risk(
                row.avg_strain or 0,
                high_strain_hours,
                medium_strain_hours,
            ),
        }

    async def _count_strain_hours(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
        min_threshold: float,
        max_threshold: float,
    ) -> int:
        """Count hours with strain above threshold."""
        query = select(
            func.count(func.distinct(func.date_trunc("hour", EnergyMetric.time)))
        ).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
                EnergyMetric.grid_strain_percent >= min_threshold,
                EnergyMetric.grid_strain_percent < max_threshold,
            )
        )

        result = await self.db.execute(query)
        return result.scalar() or 0

    async def _get_peak_strain_time(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
    ) -> Optional[datetime]:
        """Get timestamp of peak grid strain."""
        query = select(
            EnergyMetric.time,
            EnergyMetric.grid_strain_percent,
        ).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
                EnergyMetric.grid_strain_percent.isnot(None),
            )
        ).order_by(EnergyMetric.grid_strain_percent.desc()).limit(1)

        result = await self.db.execute(query)
        row = result.one_or_none()

        return row[0] if row else None

    async def _analyze_peak_hours(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
    ) -> dict[str, Any]:
        """
        Analyze which hours of the day have peak usage.

        Returns distribution of consumption by hour of day.
        """
        query = select(
            func.date_part("hour", EnergyMetric.time).label("hour"),
            func.avg(EnergyMetric.consumption_kwh).label("avg_consumption"),
        ).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
            )
        ).group_by(
            func.date_part("hour", EnergyMetric.time)
        ).order_by(
            func.avg(EnergyMetric.consumption_kwh).desc()
        )

        result = await self.db.execute(query)
        rows = result.all()

        # Get peak hours (top 25%)
        all_consumption = [r[1] for r in rows if r[1]]
        if not all_consumption:
            return {"peak_hours": [], "distribution": {}}

        threshold = sorted(all_consumption, reverse=True)[
            int(len(all_consumption) * 0.25)
        ]

        peak_hours = [int(r[0]) for r in rows if r[1] and r[1] >= threshold]

        distribution = {
            f"{int(r[0]):02d}:00": round(r[1] or 0, 2)
            for r in rows
        }

        return {
            "peak_hours": peak_hours,
            "distribution": distribution,
        }

    async def _calculate_grid_hashrate_correlation(
        self,
        site_id: str,
        start_time: datetime,
        end_time: datetime,
    ) -> dict[str, float]:
        """
        Calculate correlation between grid strain and hashrate.

        Returns Pearson correlation coefficient.
        """
        query = select(
            EnergyMetric.grid_strain_percent,
            EnergyMetric.hashrate_th,
        ).where(
            and_(
                EnergyMetric.site_id == site_id,
                EnergyMetric.time >= start_time,
                EnergyMetric.time <= end_time,
                EnergyMetric.grid_strain_percent.isnot(None),
                EnergyMetric.hashrate_th.isnot(None),
            )
        )

        result = await self.db.execute(query)
        rows = result.all()

        if len(rows) < 2:
            return {"correlation": 0.0, "interpretation": "insufficient_data"}

        strains = [r[0] for r in rows]
        hashrates = [r[1] for r in rows]

        # Calculate Pearson correlation
        n = len(rows)
        sum_strain = sum(strains)
        sum_hashrate = sum(hashrates)
        sum_strain_hashrate = sum(s * h for s, h in zip(strains, hashrates))
        sum_strain_sq = sum(s**2 for s in strains)
        sum_hashrate_sq = sum(h**2 for h in hashrates)

        numerator = n * sum_strain_hashrate - sum_strain * sum_hashrate
        denominator = (
            (n * sum_strain_sq - sum_strain**2) ** 0.5 *
            (n * sum_hashrate_sq - sum_hashrate**2) ** 0.5
        )

        if denominator == 0:
            return {"correlation": 0.0, "interpretation": "no_variance"}

        correlation = numerator / denominator

        # Interpret correlation
        if abs(correlation) < 0.3:
            interpretation = "weak"
        elif abs(correlation) < 0.7:
            interpretation = "moderate"
        else:
            interpretation = "strong"

        direction = "positive" if correlation > 0 else "negative"

        return {
            "correlation": round(correlation, 3),
            "interpretation": f"{direction}_{interpretation}",
        }

    def _assess_risk(
        self,
        avg_strain: float,
        high_strain_hours: int,
        medium_strain_hours: int,
    ) -> dict[str, Any]:
        """Assess grid risk level based on metrics."""
        # Calculate risk score
        risk_score = 0

        if avg_strain > self.HIGH_STRAIN_THRESHOLD:
            risk_score += 40
        elif avg_strain > self.MEDIUM_STRAIN_THRESHOLD:
            risk_score += 20

        risk_score += min(30, high_strain_hours * 2)
        risk_score += min(20, medium_strain_hours * 0.5)

        # Determine risk level
        if risk_score >= 60:
            level = "high"
            recommendation = "Implement demand response or capacity expansion"
        elif risk_score >= 30:
            level = "medium"
            recommendation = "Monitor closely and consider load shifting"
        else:
            level = "low"
            recommendation = "Continue monitoring, current operations acceptable"

        return {
            "level": level,
            "score": min(100, risk_score),
            "recommendation": recommendation,
        }

    async def get_grid_recommendations(
        self,
        site_id: str,
        days: int = 30,
    ) -> list[dict[str, Any]]:
        """
        Get recommendations for reducing grid impact.

        Args:
            site_id: Mining site ID
            days: Analysis period in days

        Returns:
            List of recommendations with priority
        """
        end_time = datetime.utcnow()
        start_time = end_time - timedelta(days=days)

        analysis = await self.analyze_grid_impact(site_id, start_time, end_time)

        recommendations = []

        # Check strain levels
        if analysis["strain_metrics"]["average_percent"] > self.HIGH_STRAIN_THRESHOLD:
            recommendations.append({
                "id": "reduce_load",
                "priority": "high",
                "category": "load_management",
                "title": "Reduce Peak Load",
                "description": f"Average grid strain of {analysis['strain_metrics']['average_percent']}% exceeds recommended threshold. Consider reducing operations during peak hours.",
                "potential_impact": "15-25% reduction in grid strain",
            })

        # Check for peak hour usage
        peak_hours = analysis["peak_usage_hours"].get("peak_hours", [])
        peak_grid_hours = [14, 15, 16, 17, 18, 19]  # Typical peak hours

        overlapping = set(peak_hours) & set(peak_grid_hours)
        if overlapping:
            recommendations.append({
                "id": "shift_load",
                "priority": "medium",
                "category": "load_shifting",
                "title": "Shift Operations to Off-Peak Hours",
                "description": f"Operations detected during grid peak hours ({sorted(overlapping)}). Shifting to off-peak hours can reduce strain and energy costs.",
                "potential_impact": "10-20% reduction in energy costs",
            })

        # Check correlation
        correlation = analysis["grid_correlation"]["correlation"]
        if correlation > 0.7:
            recommendations.append({
                "id": "decouple_hashrate",
                "priority": "medium",
                "category": "efficiency",
                "title": "Decouple Hashrate from Grid Strain",
                "description": "Strong correlation between hashrate and grid strain. Consider using more efficient mining equipment or renewable energy sources.",
                "potential_impact": "Improved efficiency and reduced carbon footprint",
            })

        # General recommendations
        if not recommendations:
            recommendations.append({
                "id": "maintain",
                "priority": "low",
                "category": "monitoring",
                "title": "Continue Current Operations",
                "description": "Grid impact metrics are within acceptable ranges. Continue monitoring and periodic review.",
                "potential_impact": "Maintain current performance",
            })

        # Sort by priority
        priority_order = {"high": 0, "medium": 1, "low": 2}
        recommendations.sort(key=lambda x: priority_order.get(x["priority"], 3))

        return recommendations
