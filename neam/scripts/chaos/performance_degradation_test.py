#!/usr/bin/env python3
"""
Performance Degradation Testing Framework for Chaos Engineering

This module provides comprehensive performance testing capabilities under
chaos conditions, measuring system behavior and validating degradation bounds.
"""

import asyncio
import json
import time
import hashlib
import uuid
import logging
import statistics
from dataclasses import dataclass, field, asdict
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional, Callable, Any, Tuple
from collections import defaultdict
import aiohttp
import locust
from locust import HttpUser, task, between, events
import redis
import psycopg2
import subprocess
import os

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class LoadPattern(Enum):
    """Types of load patterns for performance testing."""
    CONSTANT = "constant"
    RAMP_UP = "ramp_up"
    RAMP_DOWN = "ramp_down"
    STEP = "step"
    SPIKE = "spike"
    WAVE = "wave"


class MetricType(Enum):
    """Types of performance metrics."""
    LATENCY_P50 = "latency_p50"
    LATENCY_P95 = "latency_p95"
    LATENCY_P99 = "latency_p99"
    THROUGHPUT = "throughput"
    ERROR_RATE = "error_rate"
    CPU_USAGE = "cpu_usage"
    MEMORY_USAGE = "memory_usage"
    NETWORK_IO = "network_io"
    DISK_IO = "disk_io"
    ACTIVE_CONNECTIONS = "active_connections"


@dataclass
class PerformanceMetric:
    """Single performance metric measurement."""
    metric_type: MetricType
    value: float
    unit: str
    timestamp: str
    labels: Dict[str, str] = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        return {
            "metric_type": self.metric_type.value,
            "value": self.value,
            "unit": self.unit,
            "timestamp": self.timestamp,
            "labels": self.labels
        }


@dataclass
class PerformanceMeasurement:
    """Complete performance measurement under specific conditions."""
    measurement_id: str
    experiment_id: str
    chaos_condition: str
    start_time: str
    end_time: str
    duration_seconds: int
    load_pattern: LoadPattern
    user_count: int
    metrics: List[PerformanceMetric] = field(default_factory=list)
    requests_total: int = 0
    requests_success: int = 0
    requests_failed: int = 0
    
    def to_dict(self) -> dict:
        return {
            "measurement_id": self.measurement_id,
            "experiment_id": self.experiment_id,
            "chaos_condition": self.chaos_condition,
            "start_time": self.start_time,
            "end_time": self.end_time,
            "duration_seconds": self.duration_seconds,
            "load_pattern": self.load_pattern.value,
            "user_count": self.user_count,
            "requests_total": self.requests_total,
            "requests_success": self.requests_success,
            "requests_failed": self.requests_failed,
            "metrics": [m.to_dict() for m in self.metrics]
        }


@dataclass
class DegradationResult:
    """Result of comparing performance under chaos vs baseline."""
    metric_type: MetricType
    baseline_value: float
    chaos_value: float
    degradation_percentage: float
    within_bounds: bool
    threshold: float
    verdict: str  # "acceptable", "degraded", "critical"
    
    def to_dict(self) -> dict:
        return {
            "metric_type": self.metric_type.value,
            "baseline_value": self.baseline_value,
            "chaos_value": self.chaos_value,
            "degradation_percentage": self.degradation_percentage,
            "within_bounds": self.within_bounds,
            "threshold": self.threshold,
            "verdict": self.verdict
        }


@dataclass
class PerformanceBaseline:
    """Performance baseline for comparison."""
    baseline_id: str
    service_name: str
    created_at: str
    load_pattern: LoadPattern
    user_count: int
    metrics: Dict[str, float] = field(default_factory=dict)  # metric_type -> value
    metadata: Dict[str, Any] = field(default_factory=dict)


class ChaosLoadGenerator:
    """Load generator for performance testing under chaos conditions."""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.metrics_collector = MetricsCollector()
        self.results: List[PerformanceMeasurement] = []
    
    async def generate_load(
        self,
        endpoint: str,
        method: str = "GET",
        payload: Optional[dict] = None,
        headers: Optional[dict] = None,
        concurrent_users: int = 10,
        requests_per_user: int = 100,
        delay_between_requests: float = 1.0
    ) -> Tuple[int, int, List[PerformanceMetric]]:
        """Generate load against an endpoint."""
        success_count = 0
        error_count = 0
        latencies = []
        
        async with aiohttp.ClientSession() as session:
            tasks = []
            for user_id in range(concurrent_users):
                for _ in range(requests_per_user):
                    task = self._make_request(
                        session, endpoint, method, payload, headers, latencies
                    )
                    tasks.append(task)
                    
                    if delay_between_requests > 0:
                        await asyncio.sleep(delay_between_requests)
            
            # Execute all requests concurrently with rate limiting
            semaphore = asyncio.Semaphore(50)
            
            async def limited_task(task):
                async with semaphore:
                    return await task
            
            tasks = [limited_task(t) for t in tasks]
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            for result in results:
                if isinstance(result, Exception):
                    error_count += 1
                elif result:
                    success_count += 1
        
        # Calculate metrics
        metrics = self.metrics_collector.calculate_latency_metrics(latencies)
        metrics.append(self.metrics_collector.create_metric(
            MetricType.THROUGHPUT,
            success_count + error_count,
            "requests",
            f"{concurrent_users * requests_per_user} total requests"
        ))
        
        return success_count, error_count, metrics
    
    async def _make_request(
        self,
        session: aiohttp.ClientSession,
        endpoint: str,
        method: str,
        payload: Optional[dict],
        headers: Optional[dict],
        latencies: List[float]
    ) -> bool:
        """Make a single HTTP request and record latency."""
        url = f"{self.base_url}/{endpoint.lstrip('/')}"
        start_time = time.time()
        
        try:
            if method.upper() == "GET":
                async with session.get(url, headers=headers) as response:
                    await response.read()
                    elapsed = time.time() - start_time
                    latencies.append(elapsed)
                    return response.status < 400
            elif method.upper() == "POST":
                async with session.post(url, json=payload, headers=headers) as response:
                    await response.read()
                    elapsed = time.time() - start_time
                    latencies.append(elapsed)
                    return response.status < 400
            elif method.upper() == "PUT":
                async with session.put(url, json=payload, headers=headers) as response:
                    await response.read()
                    elapsed = time.time() - start_time
                    latencies.append(elapsed)
                    return response.status < 400
            elif method.upper() == "DELETE":
                async with session.delete(url, headers=headers) as response:
                    await response.read()
                    elapsed = time.time() - start_time
                    latencies.append(elapsed)
                    return response.status < 400
            else:
                logger.warning(f"Unsupported HTTP method: {method}")
                return False
        except Exception as e:
            elapsed = time.time() - start_time
            latencies.append(elapsed)
            logger.error(f"Request failed: {e}")
            return False


class MetricsCollector:
    """Collector for performance metrics."""
    
    def __init__(self):
        self.metrics: List[PerformanceMetric] = []
    
    def create_metric(
        self,
        metric_type: MetricType,
        value: float,
        unit: str,
        context: str = ""
    ) -> PerformanceMetric:
        """Create a new performance metric."""
        metric = PerformanceMetric(
            metric_type=metric_type,
            value=value,
            unit=unit,
            timestamp=datetime.utcnow().isoformat(),
            labels={"context": context} if context else {}
        )
        self.metrics.append(metric)
        return metric
    
    def calculate_latency_metrics(
        self,
        latencies: List[float],
        context: str = ""
    ) -> List[PerformanceMetric]:
        """Calculate latency metrics from raw latencies."""
        if not latencies:
            return []
        
        metrics = []
        sorted_latencies = sorted(latencies)
        n = len(sorted_latencies)
        
        # P50
        p50_index = int(n * 0.50)
        p50 = sorted_latencies[p50_index] * 1000  # Convert to ms
        metrics.append(self.create_metric(
            MetricType.LATENCY_P50, p50, "ms", context
        ))
        
        # P95
        p95_index = int(n * 0.95)
        p95 = sorted_latencies[p95_index] * 1000
        metrics.append(self.create_metric(
            MetricType.LATENCY_P95, p95, "ms", context
        ))
        
        # P99
        p99_index = int(n * 0.99)
        p99 = sorted_latencies[p99_index] * 1000
        metrics.append(self.create_metric(
            MetricType.LATENCY_P99, p99, "ms", context
        ))
        
        # Average
        avg = statistics.mean(latencies) * 1000
        metrics.append(self.create_metric(
            MetricType.LATENCY_P50, avg, "ms", f"{context}_avg"
        ))
        
        return metrics
    
    def reset(self) -> None:
        """Reset metrics collection."""
        self.metrics = []


class PerformanceDegradationTester:
    """Main class for testing performance degradation under chaos conditions."""
    
    def __init__(self, experiment_id: str):
        self.experiment_id = experiment_id
        self.baselines: Dict[str, PerformanceBaseline] = {}
        self.measurements: List[PerformanceMeasurement] = []
        self.degradation_results: List[DegradationResult] = []
        
        # Degradation thresholds (percentage)
        self.thresholds = {
            MetricType.LATENCY_P50: {
                "acceptable": 20.0,
                "degraded": 50.0,
                "critical": 100.0
            },
            MetricType.LATENCY_P95: {
                "acceptable": 30.0,
                "degraded": 75.0,
                "critical": 150.0
            },
            MetricType.LATENCY_P99: {
                "acceptable": 50.0,
                "degraded": 100.0,
                "critical": 200.0
            },
            MetricType.THROUGHPUT: {
                "acceptable": 15.0,  # Throughput drop
                "degraded": 30.0,
                "critical": 50.0
            },
            MetricType.ERROR_RATE: {
                "acceptable": 1.0,  # Error rate threshold
                "degraded": 5.0,
                "critical": 10.0
            }
        }
        
        self.load_generator = ChaosLoadGenerator()
    
    def set_baseline(
        self,
        service_name: str,
        load_pattern: LoadPattern,
        user_count: int,
        metrics: Dict[str, float]
    ) -> PerformanceBaseline:
        """Set a performance baseline for a service."""
        baseline = PerformanceBaseline(
            baseline_id=str(uuid.uuid4()),
            service_name=service_name,
            created_at=datetime.utcnow().isoformat(),
            load_pattern=load_pattern,
            user_count=user_count,
            metrics=metrics
        )
        
        self.baselines[f"{service_name}_{load_pattern.value}_{user_count}"] = baseline
        logger.info(f"Set baseline for {service_name}: {metrics}")
        
        return baseline
    
    def get_threshold(self, metric_type: MetricType, level: str) -> float:
        """Get threshold for a metric at a specific level."""
        metric_thresholds = self.thresholds.get(metric_type, {})
        return metric_thresholds.get(level, 100.0)
    
    def compare_degradation(
        self,
        metric_type: MetricType,
        baseline_value: float,
        chaos_value: float
    ) -> DegradationResult:
        """Compare performance degradation between baseline and chaos conditions."""
        if baseline_value == 0:
            # Handle division by zero
            if chaos_value == 0:
                degradation = 0.0
            else:
                degradation = 100.0
        elif metric_type in [MetricType.THROUGHPUT]:
            # For throughput, higher is better (so we look at decrease)
            degradation = ((baseline_value - chaos_value) / baseline_value) * 100
        else:
            # For latency/error, lower is better (so we look at increase)
            degradation = ((chaos_value - baseline_value) / baseline_value) * 100
        
        # Determine if within bounds
        acceptable_threshold = self.get_threshold(metric_type, "acceptable")
        degraded_threshold = self.get_threshold(metric_type, "degraded")
        critical_threshold = self.get_threshold(metric_type, "critical")
        
        within_bounds = degradation < critical_threshold
        
        # Determine verdict
        if degradation < acceptable_threshold:
            verdict = "acceptable"
        elif degradation < degraded_threshold:
            verdict = "degraded"
        else:
            verdict = "critical"
        
        result = DegradationResult(
            metric_type=metric_type,
            baseline_value=baseline_value,
            chaos_value=chaos_value,
            degradation_percentage=degradation,
            within_bounds=within_bounds,
            threshold=critical_threshold,
            verdict=verdict
        )
        
        self.degradation_results.append(result)
        
        return result
    
    async def run_performance_test(
        self,
        service_name: str,
        endpoint: str,
        chaos_condition: str,
        load_pattern: LoadPattern,
        user_count: int,
        duration_seconds: int = 60
    ) -> PerformanceMeasurement:
        """Run a complete performance test under chaos conditions."""
        measurement_id = str(uuid.uuid4())
        start_time = datetime.utcnow().isoformat()
        
        logger.info(f"Starting performance test for {service_name} under {chaos_condition}")
        
        # Generate load
        success_count, error_count, metrics = await self.load_generator.generate_load(
            endpoint=endpoint,
            concurrent_users=user_count,
            requests_per_user=duration_seconds // 2,  # Approximate
            delay_between_requests=0.5
        )
        
        end_time = datetime.utcnow().isoformat()
        
        measurement = PerformanceMeasurement(
            measurement_id=measurement_id,
            experiment_id=self.experiment_id,
            chaos_condition=chaos_condition,
            start_time=start_time,
            end_time=end_time,
            duration_seconds=duration_seconds,
            load_pattern=load_pattern,
            user_count=user_count,
            metrics=metrics,
            requests_total=success_count + error_count,
            requests_success=success_count,
            requests_failed=error_count
        )
        
        self.measurements.append(measurement)
        
        logger.info(f"Performance test complete: {success_count} success, {error_count} failed")
        
        return measurement
    
    def validate_degradation_bounds(
        self,
        baseline_key: str,
        chaos_measurement: PerformanceMeasurement
    ) -> List[DegradationResult]:
        """Validate that degradation is within acceptable bounds."""
        baseline = self.baselines.get(baseline_key)
        if not baseline:
            logger.warning(f"Baseline not found: {baseline_key}")
            return []
        
        results = []
        
        for metric in chaos_measurement.metrics:
            metric_key = metric.metric_type.value
            if metric_key in baseline.metrics:
                baseline_value = baseline.metrics[metric_key]
                result = self.compare_degradation(
                    metric.metric_type,
                    baseline_value,
                    metric.value
                )
                result.measurement_id = chaos_measurement.measurement_id
                results.append(result)
        
        return results
    
    def generate_degradation_report(self) -> dict:
        """Generate a comprehensive degradation report."""
        return {
            "experiment_id": self.experiment_id,
            "generated_at": datetime.utcnow().isoformat(),
            "total_measurements": len(self.measurements),
            "total_degradation_checks": len(self.degradation_results),
            "degradation_results": [r.to_dict() for r in self.degradation_results],
            "summary": {
                "acceptable_count": sum(
                    1 for r in self.degradation_results if r.verdict == "acceptable"
                ),
                "degraded_count": sum(
                    1 for r in self.degradation_results if r.verdict == "degraded"
                ),
                "critical_count": sum(
                    1 for r in self.degradation_results if r.verdict == "critical"
                ),
                "overall_passed": all(
                    r.within_bounds for r in self.degradation_results
                )
            }
        }


class LocustPerformanceTest(HttpUser):
    """Locust-based performance test for HTTP services."""
    
    wait_time = between(1, 3)
    host = "http://localhost:8080"
    
    def on_start(self):
        """Called when user starts."""
        self.headers = {"Content-Type": "application/json"}
    
    @task(3)
    def test_health_endpoint(self):
        """Test health endpoint."""
        self.client.get("/health", name="health")
    
    @task(2)
    def test_api_endpoint(self):
        """Test API endpoint."""
        self.client.get("/api/v1/data", name="api_read")
    
    @task(1)
    def test_write_endpoint(self):
        """Test write endpoint."""
        payload = {
            "test_id": str(uuid.uuid4()),
            "timestamp": datetime.utcnow().isoformat(),
            "data": {"key": "value"}
        }
        self.client.post("/api/v1/data", json=payload, name="api_write")


def run_locust_test(
    host: str = "http://localhost:8080",
    users: int = 10,
    spawn_rate: int = 1,
    duration: int = 60,
    output_file: str = "locust_results.json"
) -> dict:
    """Run a Locust performance test."""
    import subprocess
    import json
    
    cmd = [
        "locust",
        "-f", "-",  # Read from stdin
        "--host", host,
        "-u", str(users),
        "-r", str(spawn_rate),
        "--run-time", f"{duration}s",
        "--json",
        "--only-summary"
    ]
    
    # Locust test script
    test_script = f"""
from locust import HttpUser, task, between
import json
import uuid
from datetime import datetime

class ChaosTestUser(HttpUser):
    wait_time = between(1, 3)
    
    def on_start(self):
        self.headers = {{"Content-Type": "application/json"}}
    
    @task(3)
    def health_check(self):
        self.client.get("/health", name="health")
    
    @task(2)
    def api_read(self):
        self.client.get("/api/v1/data", name="api_read")
    
    @task(1)
    def api_write(self):
        payload = {{
            "test_id": str(uuid.uuid4()),
            "timestamp": datetime.utcnow().isoformat()
        }}
        self.client.post("/api/v1/data", json=payload, name="api_write")
"""
    
    result = subprocess.run(
        cmd,
        input=test_script,
        capture_output=True,
        text=True
    )
    
    logger.info(f"Locust test output: {result.stdout}")
    if result.stderr:
        logger.error(f"Locust test errors: {result.stderr}")
    
    return {"return_code": result.returncode, "output": result.stdout}


class ChaosPerformanceAnalyzer:
    """Analyzer for combined chaos and performance test results."""
    
    def __init__(self):
        self.chaos_results: List[dict] = []
        self.performance_results: List[dict] = []
        self.correlations: Dict[str, float] = {}
    
    def add_chaos_result(self, result: dict) -> None:
        """Add a chaos experiment result."""
        self.chaos_results.append(result)
    
    def add_performance_result(self, result: dict) -> None:
        """Add a performance test result."""
        self.performance_results.append(result)
    
    def analyze_correlation(
        self,
        chaos_metric: str,
        performance_metric: str
    ) -> float:
        """Analyze correlation between chaos conditions and performance."""
        # Simplified correlation analysis
        chaos_values = [r.get(chaos_metric, 0) for r in self.chaos_results]
        performance_values = [r.get(performance_metric, 0) for r in self.performance_results]
        
        if len(chaos_values) != len(performance_values) or len(chaos_values) == 0:
            return 0.0
        
        n = len(chaos_values)
        mean_chaos = sum(chaos_values) / n
        mean_perf = sum(performance_values) / n
        
        covariance = sum(
            (chaos_values[i] - mean_chaos) * (performance_values[i] - mean_perf)
            for i in range(n)
        ) / n
        
        std_chaos = (sum((x - mean_chaos) ** 2 for x in chaos_values) / n) ** 0.5
        std_perf = (sum((x - mean_perf) ** 2 for x in performance_values) / n) ** 0.5
        
        if std_chaos == 0 or std_perf == 0:
            return 0.0
        
        correlation = covariance / (std_chaos * std_perf)
        self.correlations[f"{chaos_metric}_{performance_metric}"] = correlation
        
        return correlation
    
    def generate_combined_report(self) -> dict:
        """Generate a combined analysis report."""
        return {
            "timestamp": datetime.utcnow().isoformat(),
            "chaos_experiments_count": len(self.chaos_results),
            "performance_tests_count": len(self.performance_results),
            "correlations": self.correlations,
            "chaos_summary": {
                "total_experiments": len(self.chaos_results),
                "successful_experiments": sum(
                    1 for r in self.chaos_results if r.get("status") == "success"
                ),
                "failed_experiments": sum(
                    1 for r in self.chaos_results if r.get("status") == "failed"
                )
            },
            "performance_summary": {
                "total_tests": len(self.performance_results),
                "passed_tests": sum(
                    1 for r in self.performance_results if r.get("passed")
                ),
                "failed_tests": sum(
                    1 for r in self.performance_results if not r.get("passed")
                )
            },
            "recommendations": self._generate_recommendations()
        }
    
    def _generate_recommendations(self) -> List[str]:
        """Generate recommendations based on analysis."""
        recommendations = []
        
        # Check for high correlations
        high_correlations = [
            k for k, v in self.correlations.items() if abs(v) > 0.7
        ]
        
        if high_correlations:
            recommendations.append(
                f"Strong correlations detected: {high_correlations}. "
                "Consider adjusting chaos experiment parameters."
            )
        
        # Check for performance degradation
        degradation_count = sum(
            1 for r in self.degradation_results if r.verdict == "critical"
        ) if hasattr(self, 'degradation_results') else 0
        
        if degradation_count > 0:
            recommendations.append(
                f"Critical degradation detected in {degradation_count} tests. "
                "Review system capacity and resilience patterns."
            )
        
        if not recommendations:
            recommendations.append("System performance is within acceptable bounds.")
        
        return recommendations


def run_combined_chaos_performance_test(
    experiment_id: str,
    endpoints: List[dict],
    chaos_conditions: List[str],
    baseline_config: dict
) -> dict:
    """Run a combined chaos and performance test."""
    tester = PerformanceDegradationTester(experiment_id)
    
    # Set up baseline
    for baseline in baseline_config.get("baselines", []):
        tester.set_baseline(
            service_name=baseline["service_name"],
            load_pattern=LoadPattern(baseline["load_pattern"]),
            user_count=baseline["user_count"],
            metrics=baseline["metrics"]
        )
    
    # Run tests for each chaos condition
    results = []
    for condition in chaos_conditions:
        for endpoint in endpoints:
            measurement = asyncio.run(tester.run_performance_test(
                service_name=endpoint["service"],
                endpoint=endpoint["path"],
                chaos_condition=condition,
                load_pattern=LoadPattern(endpoint.get("load_pattern", "constant")),
                user_count=endpoint.get("users", 10),
                duration_seconds=endpoint.get("duration", 60)
            ))
            
            # Validate against baseline
            baseline_key = f"{endpoint['service']}_{endpoint.get('load_pattern', 'constant')}_{endpoint.get('users', 10)}"
            degradation_results = tester.validate_degradation_bounds(baseline_key, measurement)
            
            results.append({
                "chaos_condition": condition,
                "endpoint": endpoint["path"],
                "measurement": measurement.to_dict(),
                "degradation_results": [r.to_dict() for r in degradation_results]
            })
    
    # Generate report
    report = tester.generate_degradation_report()
    report["detailed_results"] = results
    
    return report


if __name__ == "__main__":
    import sys
    import asyncio
    
    experiment_id = sys.argv[1] if len(sys.argv) > 1 else "chaos-perf-test"
    
    # Example baseline configuration
    baseline_config = {
        "baselines": [
            {
                "service_name": "api-gateway",
                "load_pattern": "constant",
                "user_count": 10,
                "metrics": {
                    "latency_p50": 50.0,
                    "latency_p95": 150.0,
                    "latency_p99": 300.0,
                    "throughput": 100.0,
                    "error_rate": 0.1
                }
            }
        ]
    }
    
    endpoints = [
        {"service": "api-gateway", "path": "api/v1/data", "users": 10, "duration": 30}
    ]
    
    chaos_conditions = ["pod-failure", "network-delay", "cpu-pressure"]
    
    # Run combined test
    report = run_combined_chaos_performance_test(
        experiment_id,
        endpoints,
        chaos_conditions,
        baseline_config
    )
    
    print(json.dumps(report, indent=2))
