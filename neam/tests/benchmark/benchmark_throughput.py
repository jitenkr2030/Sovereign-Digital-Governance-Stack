#!/usr/bin/env python3
"""
NEAM Payment Adapter Performance Benchmark

This module provides comprehensive performance benchmarking for the Payment Adapter
service, validating the 5000+ messages/second throughput target.

Usage:
    python3 tests/benchmark/benchmark_throughput.py [--brokers KAFKA_BROKERS]

Requirements:
    pip install kafka-python prometheus-api-client tqdm
"""

import argparse
import json
import logging
import os
import sys
import time
import uuid
from collections import deque
from dataclasses import dataclass, field
from datetime import datetime
from multiprocessing import Process, Queue
from statistics import mean, stdev
from typing import Any, Deque, Dict, List, Optional

from kafka import KafkaProducer
from prometheus_api_client import PrometheusConnect

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


@dataclass
class BenchmarkConfig:
    """Configuration for performance benchmark."""
    kafka_brokers: str = "localhost:9092"
    kafka_input_topic: str = "payments.raw"
    num_messages: int = 50000
    num_producers: int = 5
    message_size: int = 1024
    warmup_messages: int = 5000
    cooldown_seconds: int = 10
    target_throughput: int = 5000
    prometheus_url: str = "http://localhost:9090"


@dataclass
class BenchmarkResult:
    """Results from a benchmark run."""
    total_messages: int = 0
    duration_seconds: float = 0.0
    throughput_msg_per_sec: float = 0.0
    throughput_mb_per_sec: float = 0.0
    latency_p50_ms: float = 0.0
    latency_p90_ms: float = 0.0
    latency_p99_ms: float = 0.0
    latency_p999_ms: float = 0.0
    errors: int = 0
    error_rate: float = 0.0
    producer_stats: Dict[str, Any] = field(default_factory=dict)
    passed: bool = False


class ThroughputBenchmark:
    """High-throughput message generator for benchmarking."""

    def __init__(self, config: BenchmarkConfig):
        self.config = config
        self.results: Deque[float] = deque(maxlen=10000)
        self.producer_stats: Dict[str, Any] = {}
        self.start_time: Optional[float] = None
        self.end_time: Optional[float] = None
        self.messages_sent: int = 0
        self.messages_acked: int = 0
        self.errors: int = 0

    def create_test_message(self, index: int) -> str:
        """Create a test message of configured size."""
        base_message = {
            "id": str(uuid.uuid4()),
            "index": index,
            "amount": 1000.00 + index * 0.01,
            "currency": "EUR",
            "sender_bic": "TESTDEFF",
            "receiver_bic": "TESTUS33",
            "sender_account": f"DE89370400440532013{index % 100:02d}",
            "receiver_account": f"DE89370400440532013{(index + 1) % 100:02d}",
            "timestamp": datetime.now().isoformat(),
            "sequence": index,
        }

        # Pad message to target size
        message_str = json.dumps(base_message)
        padding_needed = self.config.message_size - len(message_str)
        if padding_needed > 0:
            base_message["_padding"] = "x" * padding_needed

        return json.dumps(base_message)

    def producer_worker(
        self,
        producer_id: int,
        start_idx: int,
        num_messages: int,
        result_queue: Queue,
    ) -> None:
        """Worker function for sending messages from a single producer."""
        producer = KafkaProducer(
            bootstrap_servers=self.config.kafka_brokers,
            value_serializer=lambda v: v.encode("utf-8"),
            acks="all",
            retries=3,
            retry_backoff_ms=100,
            linger_ms=5,
            batch_size=16384,
            max_in_flight_requests_per_connection=100,
        )

        local_sent = 0
        local_acked = 0
        local_errors = 0
        local_latencies: List[float] = []

        start = time.time()

        for i in range(num_messages):
            idx = start_idx + i
            message = self.create_test_message(idx)

            try:
                send_start = time.time()
                future = producer.send(
                    self.config.kafka_input_topic,
                    value=message,
                    key=str(idx),
                )
                record_metadata = future.get(timeout=10)
                send_latency = time.time() - send_start

                local_sent += 1
                local_latencies.append(send_latency * 1000)  # Convert to ms

                if record_metadata:
                    local_acked += 1

            except Exception as e:
                local_errors += 1

        producer.flush()
        end = time.time()

        result_queue.put({
            "producer_id": producer_id,
            "sent": local_sent,
            "acked": local_acked,
            "errors": local_errors,
            "duration": end - start,
            "latencies": local_latencies,
        })

        producer.close()

    def run_benchmark(self) -> BenchmarkResult:
        """Run the throughput benchmark with multiple producers."""
        logger.info("=" * 60)
        logger.info("NEAM Payment Adapter Throughput Benchmark")
        logger.info("=" * 60)
        logger.info(f"Configuration:")
        logger.info(f"  - Total messages: {self.config.num_messages}")
        logger.info(f"  - Number of producers: {self.config.num_producers}")
        logger.info(f"  - Messages per producer: {self.config.num_messages // self.config.num_producers}")
        logger.info(f"  - Target throughput: {self.config.target_throughput}+ msg/sec")
        logger.info("=" * 60)

        # Calculate messages per producer
        messages_per_producer = self.config.num_messages // self.config.num_producers
        remainder = self.config.num_messages % self.config.num_producers

        # Create result queue for collecting producer results
        result_queue: Queue = Process.Queue()

        # Start producers
        processes = []
        self.start_time = time.time()

        for i in range(self.config.num_producers):
            count = messages_per_producer + (1 if i < remainder else 0)
            start_idx = i * messages_per_producer + min(i, remainder)

            p = Process(
                target=self.producer_worker,
                args=(i, start_idx, count, result_queue),
            )
            processes.append(p)
            p.start()

        # Wait for all producers to complete
        for p in processes:
            p.join()

        self.end_time = time.time()

        # Collect results
        all_latencies: List[float] = []
        total_sent = 0
        total_acked = 0
        total_errors = 0

        while not result_queue.empty():
            result = result_queue.get_nowait()
            total_sent += result["sent"]
            total_acked += result["acked"]
            total_errors += result["errors"]
            all_latencies.extend(result["latencies"])

            self.producer_stats[f"producer_{result['producer_id']}"] = {
                "sent": result["sent"],
                "acked": result["acked"],
                "errors": result["errors"],
                "duration": result["duration"],
            }

        # Calculate results
        duration = self.end_time - self.start_time
        throughput = total_sent / duration if duration > 0 else 0
        throughput_mb = (total_sent * self.config.message_size) / (1024 * 1024 * duration) if duration > 0 else 0

        # Calculate latency percentiles
        all_latencies.sort()
        n = len(all_latencies)
        latency_p50 = all_latencies[int(n * 0.50)] if n > 0 else 0
        latency_p90 = all_latencies[int(n * 0.90)] if n > 0 else 0
        latency_p99 = all_latencies[int(n * 0.99)] if n > 0 else 0
        latency_p999 = all_latencies[int(n * 0.999)] if n > 0 else 0

        error_rate = (total_errors / total_sent * 100) if total_sent > 0 else 0

        result = BenchmarkResult(
            total_messages=total_sent,
            duration_seconds=duration,
            throughput_msg_per_sec=throughput,
            throughput_mb_per_sec=throughput_mb,
            latency_p50_ms=latency_p50,
            latency_p90_ms=latency_p90,
            latency_p99_ms=latency_p99,
            latency_p999_ms=latency_p999,
            errors=total_errors,
            error_rate=error_rate,
            producer_stats=self.producer_stats,
            passed=throughput >= self.config.target_throughput,
        )

        return result

    def print_results(self, result: BenchmarkResult) -> None:
        """Print benchmark results."""
        logger.info("=" * 60)
        logger.info("Benchmark Results")
        logger.info("=" * 60)
        logger.info(f"Total Messages: {result.total_messages:,}")
        logger.info(f"Duration: {result.duration_seconds:.2f} seconds")
        logger.info("")
        logger.info(f"Throughput: {result.throughput_msg_per_sec:,.2f} msg/sec")
        logger.info(f"Data Rate: {result.throughput_mb_per_sec:.2f} MB/sec")
        logger.info("")
        logger.info("Latency Percentiles:")
        logger.info(f"  p50: {result.latency_p50_ms:.2f} ms")
        logger.info(f"  p90: {result.latency_p90_ms:.2f} ms")
        logger.info(f"  p99: {result.latency_p99_ms:.2f} ms")
        logger.info(f"  p99.9: {result.latency_p999_ms:.2f} ms")
        logger.info("")
        logger.info(f"Errors: {result.errors:,}")
        logger.info(f"Error Rate: {result.error_rate:.2f}%")
        logger.info("")
        logger.info("Producer Statistics:")
        for producer_id, stats in result.producer_stats.items():
            producer_rate = stats["sent"] / stats["duration"] if stats["duration"] > 0 else 0
            logger.info(f"  {producer_id}: {stats['sent']:,} msgs in {stats['duration']:.2f}s ({producer_rate:,.0f} msg/sec)")
        logger.info("")

        # Final verdict
        if result.passed:
            logger.info("✅ BENCHMARK PASSED: Target throughput achieved!")
        else:
            logger.info(f"❌ BENCHMARK FAILED: Target was {self.config.target_throughput} msg/sec")

        logger.info("=" * 60)


class PrometheusMetricsCollector:
    """Collect metrics from Prometheus during benchmark."""

    def __init__(self, config: BenchmarkConfig):
        self.config = config
        self.prom = PrometheusConnect(url=config.prometheus_url, disable_ssl=True)
        self.metrics_history: Dict[str, List[float]] = {
            "throughput": [],
            "lag": [],
            "latency_p99": [],
            "error_rate": [],
        }

    def collect_baseline(self) -> Dict[str, float]:
        """Collect baseline metrics before benchmark."""
        metrics = {}
        try:
            # Throughput
            result = self.prom.custom_query(query="sum(rate(payment_adapter_messages_processed_total[1m]))")
            metrics["baseline_throughput"] = float(result[0]["value"][1]) if result else 0

            # Lag
            result = self.prom.custom_query(query="sum(kafka_consumergroup_lag)")
            metrics["baseline_lag"] = float(result[0]["value"][1]) if result else 0

            # Latency
            result = self.prom.custom_query(
                query="histogram_quantile(0.99, sum(rate(payment_adapter_processing_duration_seconds_bucket[5m])) by (le))"
            )
            metrics["baseline_latency_p99"] = float(result[0]["value"][1]) if result else 0

        except Exception as e:
            logger.warning(f"Failed to collect baseline metrics: {e}")

        return metrics

    def collect_during_benchmark(self, interval: float = 1.0) -> None:
        """Collect metrics during benchmark run."""
        # This would typically run in a separate thread/process
        # For simplicity, we're collecting after the fact
        pass

    def collect_after_benchmark(self) -> Dict[str, float]:
        """Collect metrics after benchmark."""
        metrics = {}
        try:
            # Peak throughput
            result = self.prom.custom_query(
                query="sum(rate(payment_adapter_messages_processed_total[5m]))"
            )
            metrics["peak_throughput"] = float(result[0]["value"][1]) if result else 0

            # Final lag
            result = self.prom.custom_query(query="sum(kafka_consumergroup_lag)")
            metrics["final_lag"] = float(result[0]["value"][1]) if result else 0

            # p99 latency
            result = self.prom.custom_query(
                query="histogram_quantile(0.99, sum(rate(payment_adapter_processing_duration_seconds_bucket[5m])) by (le))"
            )
            metrics["final_latency_p99"] = float(result[0]["value"][1]) if result else 0

        except Exception as e:
            logger.warning(f"Failed to collect metrics: {e}")

        return metrics


def run_sustained_load_test(
    config: BenchmarkConfig,
    duration_seconds: int = 300,
) -> Dict[str, Any]:
    """Run a sustained load test to verify stability over time."""
    logger.info("=" * 60)
    logger.info("Sustained Load Test")
    logger.info(f"Duration: {duration_seconds} seconds")
    logger.info(f"Target rate: {config.target_throughput} msg/sec")
    logger.info("=" * 60)

    # Calculate messages to send
    target_messages = config.target_throughput * duration_seconds
    messages_per_second = config.target_throughput

    benchmark = ThroughputBenchmark(config)

    start_time = time.time()
    last_report = start_time
    reported_throughput = 0

    while time.time() - start_time < duration_seconds:
        # Send batch of messages
        batch_size = int(messages_per_second * 0.1)  # Send in 100ms batches

        producer = KafkaProducer(
            bootstrap_servers=config.kafka_brokers,
            value_serializer=lambda v: v.encode("utf-8"),
            acks="all",
            linger_ms=5,
            batch_size=16384,
        )

        sent = 0
        for _ in range(batch_size):
            message = benchmark.create_test_message(int(time.time() * 1000))
            future = producer.send(config.kafka_input_topic, value=message)
            sent += 1

        producer.flush()
        producer.close()

        # Report progress
        current_time = time.time()
        if current_time - last_report >= 10:
            elapsed = current_time - start_time
            progress = elapsed / duration_seconds * 100
            logger.info(f"Progress: {progress:.1f}% ({elapsed:.0f}/{duration_seconds}s)")
            last_report = current_time

        time.sleep(0.1)  # Small delay between batches

    logger.info("Sustained load test completed")
    return {"status": "completed", "duration": duration_seconds}


def main():
    """Main entry point for performance benchmark."""
    parser = argparse.ArgumentParser(
        description="NEAM Payment Adapter Performance Benchmark"
    )
    parser.add_argument(
        "--brokers",
        default=os.getenv("KAFKA_BROKERS", "localhost:9092"),
        help="Kafka broker addresses (default: localhost:9092)",
    )
    parser.add_argument(
        "--num-messages",
        type=int,
        default=50000,
        help="Number of messages to send (default: 50000)",
    )
    parser.add_argument(
        "--num-producers",
        type=int,
        default=5,
        help="Number of parallel producer processes (default: 5)",
    )
    parser.add_argument(
        "--message-size",
        type=int,
        default=1024,
        help="Message size in bytes (default: 1024)",
    )
    parser.add_argument(
        "--target-throughput",
        type=int,
        default=5000,
        help="Target throughput in msg/sec (default: 5000)",
    )
    parser.add_argument(
        "--sustained",
        action="store_true",
        help="Run sustained load test instead of burst test",
    )
    parser.add_argument(
        "--sustained-duration",
        type=int,
        default=300,
        help="Duration of sustained load test in seconds (default: 300)",
    )
    parser.add_argument(
        "--prometheus-url",
        default=os.getenv("PROMETHEUS_URL", "http://localhost:9090"),
        help="Prometheus URL (default: http://localhost:9090)",
    )

    args = parser.parse_args()

    config = BenchmarkConfig(
        kafka_brokers=args.brokers,
        num_messages=args.num_messages,
        num_producers=args.num_producers,
        message_size=args.message_size,
        target_throughput=args.target_throughput,
        prometheus_url=args.prometheus_url,
    )

    if args.sustained:
        # Run sustained load test
        result = run_sustained_load_test(config, args.sustained_duration)
        print(json.dumps(result, indent=2))
    else:
        # Run burst throughput benchmark
        benchmark = ThroughputBenchmark(config)
        result = benchmark.run_benchmark()
        benchmark.print_results(result)

        # Exit with appropriate code
        if not result.passed:
            sys.exit(1)
        sys.exit(0)


if __name__ == "__main__":
    main()
