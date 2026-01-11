"""
NEAM Simulation Service - Base Engine Classes
Abstract base classes for simulation engines
"""

from abc import ABC, abstractmethod
from datetime import datetime
from typing import Any, Dict, Generic, List, Optional, TypeVar
import logging
import time
import asyncio
from concurrent.futures import ThreadPoolExecutor
import numpy as np

from ..models.simulation import (
    BaseSimulationRequest,
    BaseSimulationResponse,
    SimulationStatus,
)


logger = logging.getLogger(__name__)


R = TypeVar('R', bound=BaseSimulationResponse)
T = TypeVar('T', bound=BaseSimulationRequest)


class BaseSimulationEngine(ABC):
    """Abstract base class for all simulation engines"""
    
    def __init__(
        self,
        max_workers: int = 4,
        default_timeout: float = 300.0,
        random_seed: Optional[int] = None
    ):
        """
        Initialize the simulation engine
        
        Args:
            max_workers: Maximum number of worker threads for parallel execution
            default_timeout: Default timeout for simulations in seconds
            random_seed: Global random seed for reproducibility
        """
        self.max_workers = max_workers
        self.default_timeout = default_timeout
        self.random_seed = random_seed
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self._simulation_cache: Dict[str, Dict[str, Any]] = {}
        
        if random_seed is not None:
            np.random.seed(random_seed)
    
    def __del__(self):
        """Cleanup executor on deletion"""
        if hasattr(self, 'executor'):
            self.executor.shutdown(wait=False)
    
    def set_random_seed(self, seed: int) -> None:
        """Set random seed for reproducibility"""
        self.random_seed = seed
        np.random.seed(seed)
    
    async def run_simulation(
        self,
        request: T,
        progress_callback: Optional[callable] = None
    ) -> R:
        """
        Run a simulation with optional progress tracking
        
        Args:
            request: Simulation request parameters
            progress_callback: Optional callback for progress updates
            
        Returns:
            Simulation results
        """
        start_time = time.time()
        status = SimulationStatus.RUNNING
        
        try:
            # Generate simulation ID if not provided
            if not request.simulation_id:
                request.simulation_id = self._generate_simulation_id()
            
            logger.info(f"Starting simulation {request.simulation_id} of type {request.simulation_type}")
            
            # Execute the simulation
            result = await self._execute_simulation(request, progress_callback)
            
            # Calculate duration
            duration_ms = (time.time() - start_time) * 1000
            result.duration_ms = duration_ms
            result.status = SimulationStatus.COMPLETED
            result.completed_at = datetime.utcnow()
            
            logger.info(
                f"Simulation {request.simulation_id} completed in {duration_ms:.2f}ms"
            )
            
            # Cache the result
            self._cache_result(request.simulation_id, result)
            
            return result
            
        except asyncio.CancelledError:
            duration_ms = (time.time() - start_time) * 1000
            status = SimulationStatus.CANCELLED
            logger.warning(f"Simulation {request.simulation_id} was cancelled")
            raise
            
        except Exception as e:
            duration_ms = (time.time() - start_time) * 1000
            status = SimulationStatus.FAILED
            logger.error(f"Simulation {request.simulation_id} failed: {str(e)}")
            
            # Create a failed response
            result = self._create_failed_response(
                request,
                str(e),
                duration_ms
            )
            raise
            
        finally:
            # Cleanup resources if needed
            await self._cleanup()
    
    def run_simulation_sync(
        self,
        request: T,
        progress_callback: Optional[callable] = None
    ) -> R:
        """
        Synchronous version of run_simulation
        
        Args:
            request: Simulation request parameters
            progress_callback: Optional callback for progress updates
            
        Returns:
            Simulation results
        """
        loop = asyncio.new_event_loop()
        try:
            return loop.run_until_complete(
                self.run_simulation(request, progress_callback)
            )
        finally:
            loop.close()
    
    @abstractmethod
    async def _execute_simulation(
        self,
        request: T,
        progress_callback: Optional[callable] = None
    ) -> R:
        """
        Execute the specific simulation logic
        
        Args:
            request: Simulation request parameters
            progress_callback: Optional callback for progress updates
            
        Returns:
            Simulation results
        """
        pass
    
    def _generate_simulation_id(self) -> str:
        """Generate a unique simulation ID"""
        import uuid
        return f"sim_{uuid.uuid4().hex[:12]}"
    
    def _cache_result(self, simulation_id: str, result: R) -> None:
        """Cache simulation result for later retrieval"""
        self._simulation_cache[simulation_id] = {
            'result': result,
            'cached_at': datetime.utcnow()
        }
    
    def get_cached_result(self, simulation_id: str) -> Optional[R]:
        """Retrieve a cached simulation result"""
        cache_entry = self._simulation_cache.get(simulation_id)
        if cache_entry:
            return cache_entry['result']
        return None
    
    def clear_cache(self, simulation_id: Optional[str] = None) -> None:
        """Clear simulation cache"""
        if simulation_id:
            self._simulation_cache.pop(simulation_id, None)
        else:
            self._simulation_cache.clear()
    
    @abstractmethod
    def _create_failed_response(
        self,
        request: T,
        error_message: str,
        duration_ms: float
    ) -> R:
        """Create a response for a failed simulation"""
        pass
    
    async def _cleanup(self) -> None:
        """Cleanup resources after simulation"""
        pass
    
    def validate_request(self, request: T) -> List[str]:
        """
        Validate a simulation request
        
        Args:
            request: Simulation request to validate
            
        Returns:
            List of validation error messages (empty if valid)
        """
        errors = []
        
        # Check date range
        if hasattr(request, 'start_date') and hasattr(request, 'end_date'):
            if request.end_date <= request.start_date:
                errors.append("end_date must be after start_date")
        
        # Check region/sector constraints
        if hasattr(request, 'regions') and len(request.regions) > 100:
            errors.append("Maximum 100 regions allowed per simulation")
        
        if hasattr(request, 'sectors') and len(request.sectors) > 20:
            errors.append("Maximum 20 sectors allowed per simulation")
        
        return errors
    
    async def run_batch(
        self,
        requests: List[T],
        parallel: bool = True,
        progress_callback: Optional[callable] = None
    ) -> List[R]:
        """
        Run multiple simulations in batch
        
        Args:
            requests: List of simulation requests
            parallel: Whether to run simulations in parallel
            progress_callback: Optional callback for progress updates
            
        Returns:
            List of simulation results
        """
        results = []
        
        if parallel:
            tasks = [
                self.run_simulation(req, progress_callback) 
                for req in requests
            ]
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Handle exceptions
            for i, result in enumerate(results):
                if isinstance(result, Exception):
                    logger.error(f"Batch simulation {i} failed: {result}")
                    results[i] = self._convert_exception_to_response(
                        requests[i], result
                    )
        else:
            for i, request in enumerate(requests):
                try:
                    result = await self.run_simulation(request)
                    results.append(result)
                except Exception as e:
                    logger.error(f"Batch simulation {i} failed: {e}")
                    results.append(self._convert_exception_to_response(request, e))
                
                if progress_callback:
                    progress_callback((i + 1) / len(requests) * 100)
        
        return results
    
    def _convert_exception_to_response(
        self,
        request: T,
        exception: Exception
    ) -> R:
        """Convert an exception to a failed response"""
        return self._create_failed_response(
            request,
            str(exception),
            0.0
        )


class ProgressTracker:
    """Helper class for tracking simulation progress"""
    
    def __init__(
        self,
        total_steps: int,
        callback: Optional[callable] = None,
        update_frequency: int = 10
    ):
        """
        Initialize progress tracker
        
        Args:
            total_steps: Total number of steps in the simulation
            callback: Callback function for progress updates
            update_frequency: How often to call the callback (every N steps)
        """
        self.total_steps = total_steps
        self.callback = callback
        self.update_frequency = update_frequency
        self.current_step = 0
        self.last_update = 0
    
    def update(self, steps: int = 1) -> float:
        """
        Update progress
        
        Args:
            steps: Number of steps completed
            
        Returns:
            Current progress percentage
        """
        self.current_step += steps
        progress = (self.current_step / self.total_steps) * 100
        
        if self.callback and (self.current_step - self.last_update) >= self.update_frequency:
            self.callback(progress)
            self.last_update = self.current_step
        
        return progress
    
    def complete(self) -> float:
        """Mark simulation as complete"""
        self.current_step = self.total_steps
        progress = 100.0
        
        if self.callback:
            self.callback(progress)
        
        return progress
    
    @property
    def percent(self) -> float:
        """Get current progress percentage"""
        return (self.current_step / self.total_steps) * 100 if self.total_steps > 0 else 0.0
