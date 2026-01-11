"""
Mining Sites API Endpoints

Endpoints for managing mining site registration and configuration.
"""

from datetime import datetime
from typing import Any, Optional
from uuid import uuid4

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel, Field, HttpUrl
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from src.db.database import get_db
from src.models.energy_metrics import MiningSite


router = APIRouter(prefix="/sites", tags=["sites"])


# Pydantic models
class LocationData(BaseModel):
    """Location data for a mining site."""
    latitude: float = Field(..., ge=-90, le=90)
    longitude: float = Field(..., ge=-180, le=180)
    address: Optional[str] = None
    region: Optional[str] = None
    country: Optional[str] = None


class PowerSourceConfig(BaseModel):
    """Power source configuration."""
    source: str = Field(..., description="Power source code (e.g., hydro, solar)")
    proportion: float = Field(..., ge=0, le=1, description="Proportion of total power")


class EquipmentConfig(BaseModel):
    """Mining equipment configuration."""
    type: str = Field(..., description="Equipment type")
    model: Optional[str] = None
    count: int = Field(..., ge=1)
    hashrate_th_per_unit: float
    power_consumption_watts: float


class CreateSiteRequest(BaseModel):
    """Request model for creating a mining site."""
    operator_id: str = Field(..., description="Operating entity ID")
    name: str = Field(..., min_length=1, max_length=255)
    code: str = Field(..., min_length=1, max_length=50, description="Unique site code")
    location: Optional[LocationData] = None
    grid_connection_type: Optional[str] = Field(None, description="Grid connection type")
    max_power_capacity_kw: Optional[float] = Field(None, ge=0)
    primary_power_source: Optional[str] = None
    backup_power_sources: Optional[list[str]] = None
    mining_equipment: Optional[list[EquipmentConfig]] = None
    total_hashrate_th: Optional[float] = Field(None, ge=0)
    energy_permit_id: Optional[str] = None
    green_energy_certified: bool = False
    certification_expiry: Optional[datetime] = None


class UpdateSiteRequest(BaseModel):
    """Request model for updating a mining site."""
    name: Optional[str] = Field(None, min_length=1, max_length=255)
    location: Optional[LocationData] = None
    grid_connection_type: Optional[str] = None
    max_power_capacity_kw: Optional[float] = Field(None, ge=0)
    primary_power_source: Optional[str] = None
    backup_power_sources: Optional[list[str]] = None
    mining_equipment: Optional[list[EquipmentConfig]] = None
    total_hashrate_th: Optional[float] = Field(None, ge=0)
    status: Optional[str] = Field(None, regex="^(active|inactive|suspended|decommissioned)$")
    green_energy_certified: Optional[bool] = None
    certification_expiry: Optional[datetime] = None


class SiteResponse(BaseModel):
    """Response model for site data."""
    id: str
    operator_id: str
    name: str
    code: str
    location: Optional[dict[str, Any]]
    grid_connection_type: Optional[str]
    max_power_capacity_kw: Optional[float]
    primary_power_source: Optional[str]
    backup_power_sources: Optional[list[str]]
    total_hashrate_th: Optional[float]
    status: str
    energy_permit_id: Optional[str]
    green_energy_certified: bool
    certification_expiry: Optional[datetime]
    created_at: datetime
    updated_at: datetime


class SiteListResponse(BaseModel):
    """Response model for site list."""
    sites: list[SiteResponse]
    total: int
    page: int
    limit: int


@router.get("", response_model=SiteListResponse)
async def list_sites(
    operator_id: Optional[str] = None,
    status: Optional[str] = None,
    page: int = 1,
    limit: int = 20,
    db: AsyncSession = Depends(get_db),
) -> SiteListResponse:
    """
    List all mining sites with optional filtering.

    Returns paginated list of mining sites.
    """
    query = select(MiningSite)

    if operator_id:
        query = query.where(MiningSite.operator_id == operator_id)
    if status:
        query = query.where(MiningSite.status == status)

    # Get total count
    count_query = select(func.count()).select_from(query.subquery())
    count_result = await db.execute(count_query)
    total = count_result.scalar()

    # Apply pagination
    query = query.order_by(MiningSite.created_at.desc())
    query = query.offset((page - 1) * limit).limit(limit)

    result = await db.execute(query)
    sites = result.scalars().all()

    return SiteListResponse(
        sites=[
            SiteResponse(
                id=str(s.id),
                operator_id=str(s.operator_id),
                name=s.name,
                code=s.code,
                location=s.location,
                grid_connection_type=s.grid_connection_type,
                max_power_capacity_kw=s.max_power_capacity_kw,
                primary_power_source=s.primary_power_source,
                backup_power_sources=s.backup_power_sources,
                total_hashrate_th=s.total_hashrate_th,
                status=s.status,
                energy_permit_id=s.energy_permit_id,
                green_energy_certified=s.green_energy_certified,
                certification_expiry=s.certification_expiry,
                created_at=s.created_at,
                updated_at=s.updated_at,
            )
            for s in sites
        ],
        total=total,
        page=page,
        limit=limit,
    )


@router.get("/{site_id}", response_model=SiteResponse)
async def get_site(site_id: str, db: AsyncSession = Depends(get_db)) -> SiteResponse:
    """
    Get details of a specific mining site.
    """
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    return SiteResponse(
        id=str(site.id),
        operator_id=str(site.operator_id),
        name=site.name,
        code=site.code,
        location=site.location,
        grid_connection_type=site.grid_connection_type,
        max_power_capacity_kw=site.max_power_capacity_kw,
        primary_power_source=site.primary_power_source,
        backup_power_sources=site.backup_power_sources,
        total_hashrate_th=site.total_hashrate_th,
        status=site.status,
        energy_permit_id=site.energy_permit_id,
        green_energy_certified=site.green_energy_certified,
        certification_expiry=site.certification_expiry,
        created_at=site.created_at,
        updated_at=site.updated_at,
    )


@router.post("", response_model=SiteResponse, status_code=201)
async def create_site(
    request: CreateSiteRequest,
    user_id: Optional[str] = None,
    db: AsyncSession = Depends(get_db),
) -> SiteResponse:
    """
    Register a new mining site.

    Creates a new mining site record with the provided configuration.
    """
    # Check for duplicate code
    existing = await db.execute(
        select(MiningSite).where(MiningSite.code == request.code)
    )
    if existing.scalar_one_or_none():
        raise HTTPException(status_code=409, detail="Site code already exists")

    # Create site
    site = MiningSite(
        operator_id=request.operator_id,
        name=request.name,
        code=request.code,
        location=request.location.model_dump() if request.location else None,
        grid_connection_type=request.grid_connection_type,
        max_power_capacity_kw=request.max_power_capacity_kw,
        primary_power_source=request.primary_power_source,
        backup_power_sources=request.backup_power_sources,
        mining_equipment=[e.model_dump() for e in request.mining_equipment] if request.mining_equipment else None,
        total_hashrate_th=request.total_hashrate_th,
        energy_permit_id=request.energy_permit_id,
        green_energy_certified=request.green_energy_certified,
        certification_expiry=request.certification_expiry,
        status="active",
        created_by=user_id,
    )

    db.add(site)
    await db.commit()
    await db.refresh(site)

    return SiteResponse(
        id=str(site.id),
        operator_id=str(site.operator_id),
        name=site.name,
        code=site.code,
        location=site.location,
        grid_connection_type=site.grid_connection_type,
        max_power_capacity_kw=site.max_power_capacity_kw,
        primary_power_source=site.primary_power_source,
        backup_power_sources=site.backup_power_sources,
        total_hashrate_th=site.total_hashrate_th,
        status=site.status,
        energy_permit_id=site.energy_permit_id,
        green_energy_certified=site.green_energy_certified,
        certification_expiry=site.certification_expiry,
        created_at=site.created_at,
        updated_at=site.updated_at,
    )


@router.patch("/{site_id}", response_model=SiteResponse)
async def update_site(
    site_id: str,
    request: UpdateSiteRequest,
    user_id: Optional[str] = None,
    db: AsyncSession = Depends(get_db),
) -> SiteResponse:
    """
    Update a mining site configuration.
    """
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    # Update fields
    update_data = request.model_dump(exclude_unset=True)

    for field, value in update_data.items():
        if field == "location":
            value = value.model_dump() if value else None
        elif field == "mining_equipment":
            value = [e.model_dump() for e in value] if value else None
        setattr(site, field, value)

    site.updated_by = user_id
    await db.commit()
    await db.refresh(site)

    return SiteResponse(
        id=str(site.id),
        operator_id=str(site.operator_id),
        name=site.name,
        code=site.code,
        location=site.location,
        grid_connection_type=site.grid_connection_type,
        max_power_capacity_kw=site.max_power_capacity_kw,
        primary_power_source=site.primary_power_source,
        backup_power_sources=site.backup_power_sources,
        total_hashrate_th=site.total_hashrate_th,
        status=site.status,
        energy_permit_id=site.energy_permit_id,
        green_energy_certified=site.green_energy_certified,
        certification_expiry=site.certification_expiry,
        created_at=site.created_at,
        updated_at=site.updated_at,
    )


@router.delete("/{site_id}", status_code=204)
async def delete_site(site_id: str, db: AsyncSession = Depends(get_db)) -> None:
    """
    Decommission a mining site.

    Sets site status to 'decommissioned' instead of hard delete.
    """
    site = await db.get(MiningSite, site_id)
    if not site:
        raise HTTPException(status_code=404, detail="Site not found")

    site.status = "decommissioned"
    await db.commit()
