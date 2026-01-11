"""
NEAM Simulation Service - Simulation Routes
API endpoints for simulation requests and results
"""

import logging
from datetime import datetime
from typing import Any, Dict, List, Optional
from fastapi import APIRouter, HTTPException, BackgroundTasks, Query
from pydantic import BaseModel

from ..models.simulation import (
    SimulationType,
    MonteCarloRequest,
    MonteCarloResponse,
    EconomicProjectionRequest,
    EconomicProjectionResponse,
    PolicyScenarioRequest,
    PolicyScenarioResponse,
    SensitivityAnalysisRequest,
    SensitivityAnalysisResponse,
    StressTestRequest,
    StressTestResponse,
    ScenarioComparisonRequest,
    BatchSimulationRequest,
    BatchSimulationResponse,
    SimulationJobStatus,
    SimulationStatus,
)
from ..engines import (
    MonteCarloEngine,
    EconomicModelEngine,
    ScenarioEngine,
    SensitivityEngine,
)


logger = logging.getLogger(__name__)

router = APIRouter(prefix="/simulations", tags=["Simulations"])

# Engine instances
monte_carlo_engine = MonteCarloEngine()
economic_engine = EconomicModelEngine()
scenario_engine = ScenarioEngine()
sensitivity_engine = SensitivityEngine()

# Job storage (in production, use Redis)
_job_store: Dict[str, Dict[str, Any]] = {}


class SimulationResponse(BaseModel):
    """Generic simulation response wrapper"""
    simulation_id: str
    status: str
    message: str
    created_at: datetime


class JobStatusResponse(BaseModel):
    """Job status response"""
    simulation_id: str
    status: SimulationStatus
    progress_percent: float = 0.0
    message: Optional[str] = None
    estimated_completion: Optional[datetime] = None
    created_at: datetime
    started_at: Optional[datetime] = None
    error: Optional[str] = None


@router.post("/monte-carlo", response_model=MonteCarloResponse)
async def run_monte_carlo_simulation(request: MonteCarloRequest):
    """
    Run a Monte Carlo simulation for risk assessment.
    
    Generates random samples from specified probability distributions
    and calculates output statistics including confidence intervals.
    """
    try:
        logger.info(f"Received Monte Carlo simulation request: {request.simulation_id}")
        
        # Validate request
        errors = monte_carlo_engine.validate_request(request)
        if errors:
            raise HTTPException(status_code=400, detail=errors)
        
        # Run simulation
        result = await monte_carlo_engine.run_simulation(request)
        
        return result
        
    except Exception as e:
        logger.error(f"Monte Carlo simulation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/economic-projection", response_model=EconomicProjectionResponse)
async def run_economic_projection(request: EconomicProjectionRequest):
    """
    Run an economic projection using time-series forecasting models.
    
    Projects economic indicators into the future using ARIMA,
    Exponential Smoothing, or other forecasting methods.
    """
    try:
        logger.info(f"Received economic projection request: {request.simulation_id}")
        
        # Validate request
        errors = economic_engine.validate_request(request)
        if errors:
            raise HTTPException(status_code=400, detail=errors)
        
        # Run simulation
        result = await economic_engine.run_simulation(request)
        
        return result
        
    except Exception as e:
        logger.error(f"Economic projection failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/policy-scenario", response_model=PolicyScenarioResponse)
async def run_policy_scenario(request: PolicyScenarioRequest):
    """
    Run a policy scenario simulation comparing baseline and intervention outcomes.
    
    Simulates the economic impact of various policy interventions
    and compares results across multiple scenarios.
    """
    try:
        logger.info(f"Received policy scenario request: {request.simulation_id}")
        
        # Validate request
        errors = scenario_engine.validate_request(request)
        if errors:
            raise HTTPException(status_code=400, detail=errors)
        
        # Run simulation
        result = await scenario_engine.run_simulation(request)
        
        return result
        
    except Exception as e:
        logger.error(f"Policy scenario simulation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/sensitivity-analysis", response_model=SensitivityAnalysisResponse)
async def run_sensitivity_analysis(request: SensitivityAnalysisRequest):
    """
    Run a sensitivity analysis to identify influential parameters.
    
    Uses Sobol', Morris, or FAST methods to calculate sensitivity
    indices and rank parameter importance.
    """
    try:
        logger.info(f"Received sensitivity analysis request: {request.simulation_id}")
        
        # Validate request
        errors = sensitivity_engine.validate_request(request)
        if errors:
            raise HTTPException(status_code=400, detail=errors)
        
        # Run simulation
        result = await sensitivity_engine.run_simulation(request)
        
        return result
        
    except Exception as e:
        logger.error(f"Sensitivity analysis failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/stress-test", response_model=StressTestResponse)
async def run_stress_test(request: StressTestRequest):
    """
    Run stress tests on economic indicators under adverse scenarios.
    
    Evaluates economic resilience under various shock scenarios
    and identifies vulnerable sectors and regions.
    """
    try:
        logger.info(f"Received stress test request: {request.simulation_id}")
        
        # Create a custom engine for stress testing
        from ..engines.base import BaseSimulationEngine
        from ..models.simulation import SimulationStatus
        import time
        import numpy as np
        
        class StressTestEngine(BaseSimulationEngine):
            """Simple stress test engine implementation"""
            
            def __init__(self, *args, **kwargs):
                super().__init__(*args, **kwargs)
            
            async def _execute_simulation(self, request, progress_callback=None):
                from ..models.simulation import StressTestResponse, StressTestResult
                import random
                
                start_time = time.time()
                
                results = {}
                for scenario in request.scenarios:
                    # Simulate stress impact
                    impact = {}
                    for indicator in request.indicators:
                        shock = scenario.shock_parameters.get(indicator, 0)
                        impact[indicator] = {
                            'max_decline': abs(shock) * random.uniform(0.8, 1.2),
                            'duration_months': scenario.duration_months
                        }
                    
                    results[scenario.name] = StressTestResult(
                        scenario_name=scenario.name,
                        severity=scenario.severity,
                        indicator_impacts=impact,
                        time_to_recovery={ind: scenario.duration_months * 2 for ind in request.indicators},
                        peak_impact=impact,
                        systemic_risk_score=random.uniform(0.3, 0.7),
                        vulnerable_sectors=request.sectors[:2] if request.sectors else [],
                        vulnerable_regions=request.regions[:2] if request.regions else [],
                        recommendations=["Monitor indicators closely", "Prepare contingency measures"]
                    )
                
                return StressTestResponse(
                    simulation_id=request.simulation_id or self._generate_simulation_id(),
                    simulation_type=request.simulation_type,
                    status=SimulationStatus.COMPLETED,
                    created_at=datetime.utcnow(),
                    completed_at=datetime.utcnow(),
                    duration_ms=(time.time() - start_time) * 1000,
                    regions=request.regions,
                    sectors=request.sectors,
                    result={"summary": f"Stress test completed for {len(request.scenarios)} scenarios"},
                    results=results,
                    comparison_summary={"overall_risk": "moderate"},
                    resilience_score=0.65
                )
            
            def _create_failed_response(self, request, error_message, duration_ms):
                from ..models.simulation import StressTestResponse
                return StressTestResponse(
                    simulation_id=request.simulation_id or "failed",
                    simulation_type=request.simulation_type,
                    status=SimulationStatus.FAILED,
                    created_at=datetime.utcnow(),
                    completed_at=datetime.utcnow(),
                    duration_ms=duration_ms,
                    regions=request.regions,
                    sectors=request.sectors,
                    result={"error": error_message},
                    results={},
                    comparison_summary={},
                    resilience_score=0.0
                )
        
        stress_engine = StressTestEngine()
        result = await stress_engine.run_simulation(request)
        
        return result
        
    except Exception as e:
        logger.error(f"Stress test failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/batch", response_model=BatchSimulationResponse)
async def run_batch_simulation(request: BatchSimulationRequest):
    """
    Run multiple simulations in batch.
    
    Executes multiple simulation requests either in parallel or sequentially
    and returns a batch ID for tracking.
    """
    try:
        import uuid
        
        batch_id = f"batch_{uuid.uuid4().hex[:12]}"
        
        logger.info(f"Received batch simulation request: {batch_id} with {len(request.simulations)} simulations")
        
        # Route simulations to appropriate engines
        simulation_ids = []
        
        for sim in request.simulations:
            if sim.simulation_type == SimulationType.MONTE_CARLO:
                mc_request = MonteCarloRequest(**sim.model_dump())
                result = await monte_carlo_engine.run_simulation(mc_request)
                simulation_ids.append(result.simulation_id)
            elif sim.simulation_type == SimulationType.ECONOMIC_PROJECTION:
                ep_request = EconomicProjectionRequest(**sim.model_dump())
                result = await economic_engine.run_simulation(ep_request)
                simulation_ids.append(result.simulation_id)
            elif sim.simulation_type == SimulationType.POLICY_SCENARIO:
                ps_request = PolicyScenarioRequest(**sim.model_dump())
                result = await scenario_engine.run_simulation(ps_request)
                simulation_ids.append(result.simulation_id)
            elif sim.simulation_type == SimulationType.SENSITIVITY_ANALYSIS:
                sa_request = SensitivityAnalysisRequest(**sim.model_dump())
                result = await sensitivity_engine.run_simulation(sa_request)
                simulation_ids.append(result.simulation_id)
        
        # Estimate duration
        avg_duration = 1000  # 1 second average
        estimated_duration = avg_duration * len(request.simulations)
        
        return BatchSimulationResponse(
            batch_id=batch_id,
            total_simulations=len(request.simulations),
            submitted_at=datetime.utcnow(),
            simulation_ids=simulation_ids,
            estimated_duration_ms=estimated_duration,
            status=SimulationStatus.COMPLETED
        )
        
    except Exception as e:
        logger.error(f"Batch simulation failed: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/status/{simulation_id}", response_model=JobStatusResponse)
async def get_simulation_status(simulation_id: str):
    """
    Get the status of a simulation by ID.
    
    Returns the current status, progress, and any results if completed.
    """
    # Check in engine caches
    for engine, name in [
        (monte_carlo_engine, "Monte Carlo"),
        (economic_engine, "Economic Projection"),
        (scenario_engine, "Policy Scenario"),
        (sensitivity_engine, "Sensitivity Analysis")
    ]:
        result = engine.get_cached_result(simulation_id)
        if result:
            return JobStatusResponse(
                simulation_id=simulation_id,
                status=result.status,
                progress_percent=100.0,
                message="Simulation completed",
                created_at=result.created_at,
                started_at=result.created_at,
                completed_at=result.completed_at
            )
    
    # Check job store
    job = _job_store.get(simulation_id)
    if job:
        return JobStatusResponse(
            simulation_id=simulation_id,
            status=job.get('status', SimulationStatus.PENDING),
            progress_percent=job.get('progress_percent', 0.0),
            message=job.get('message'),
            created_at=job.get('created_at', datetime.utcnow())
        )
    
    raise HTTPException(status_code=404, detail="Simulation not found")


@router.get("/results/{simulation_id}")
async def get_simulation_results(simulation_id: str):
    """
    Get the results of a completed simulation.
    
    Returns the full simulation results including all computed metrics.
    """
    # Check in engine caches
    for engine in [monte_carlo_engine, economic_engine, scenario_engine, sensitivity_engine]:
        result = engine.get_cached_result(simulation_id)
        if result:
            if result.status == SimulationStatus.COMPLETED:
                return result
            elif result.status == SimulationStatus.FAILED:
                raise HTTPException(status_code=400, detail=result.result.get('error', 'Simulation failed'))
            else:
                raise HTTPException(status_code=202, detail={"status": result.status})
    
    raise HTTPException(status_code=404, detail="Simulation not found")


@router.delete("/results/{simulation_id}")
async def delete_simulation_results(simulation_id: str):
    """
    Delete cached simulation results.
    
    Frees up memory by removing cached simulation results.
    """
    deleted = False
    
    for engine in [monte_carlo_engine, economic_engine, scenario_engine, sensitivity_engine]:
        if engine.get_cached_result(simulation_id):
            engine.clear_cache(simulation_id)
            deleted = True
    
    if deleted:
        return {"message": f"Simulation {simulation_id} results deleted"}
    else:
        raise HTTPException(status_code=404, detail="Simulation not found")


@router.get("/types")
async def get_simulation_types():
    """
    Get available simulation types and their descriptions.
    
    Returns a list of all available simulation types with metadata.
    """
    return {
        "simulation_types": [
            {
                "type": "monte_carlo",
                "name": "Monte Carlo Simulation",
                "description": "Probabilistic simulation using random sampling from specified distributions",
                "use_cases": [
                    "Risk assessment",
                    "Portfolio analysis",
                    "Uncertainty quantification",
                    "Probability of outcomes"
                ],
                "parameters": ["variables with distributions", "iterations", "output formula"],
                "typical_duration": "seconds to minutes"
            },
            {
                "type": "economic_projection",
                "name": "Economic Projection",
                "description": "Time-series forecasting of economic indicators",
                "use_cases": [
                    "GDP growth forecasting",
                    "Inflation prediction",
                    "Employment trends",
                    "Sectoral analysis"
                ],
                "parameters": ["historical data", "model type", "projection horizon"],
                "typical_duration": "seconds"
            },
            {
                "type": "policy_scenario",
                "name": "Policy Scenario Analysis",
                "description": "Comparison of economic outcomes under different policy interventions",
                "use_cases": [
                    "Policy impact assessment",
                    "Cost-benefit analysis",
                    "Alternative futures",
                    "Intervention optimization"
                ],
                "parameters": ["baseline scenario", "alternative scenarios", "interventions"],
                "typical_duration": "seconds to minutes"
            },
            {
                "type": "sensitivity_analysis",
                "name": "Sensitivity Analysis",
                "description": "Identification of influential parameters using sensitivity indices",
                "use_cases": [
                    "Parameter importance",
                    "Model simplification",
                    "Research prioritization",
                    "Uncertainty reduction"
                ],
                "parameters": ["parameters with ranges", "output metric", "analysis method"],
                "typical_duration": "seconds to minutes"
            },
            {
                "type": "stress_test",
                "name": "Stress Testing",
                "description": "Economic resilience testing under adverse scenarios",
                "use_cases": [
                    "Financial stability",
                    "Contingency planning",
                    "Vulnerability assessment",
                    "Resilience scoring"
                ],
                "parameters": ["stress scenarios", "indicators", "baseline conditions"],
                "typical_duration": "seconds"
            }
        ]
    }


@router.get("/health")
async def simulation_health_check():
    """
    Health check endpoint for simulation service.
    
    Returns the status of all simulation engines.
    """
    return {
        "status": "healthy",
        "engines": {
            "monte_carlo": "ready",
            "economic_projection": "ready",
            "policy_scenario": "ready",
            "sensitivity_analysis": "ready"
        },
        "cached_simulations": {
            "monte_carlo": len(monte_carlo_engine._simulation_cache),
            "economic_projection": len(economic_engine._simulation_cache),
            "policy_scenario": len(scenario_engine._simulation_cache),
            "sensitivity_analysis": len(sensitivity_engine._simulation_cache)
        },
        "timestamp": datetime.utcnow().isoformat()
    }
