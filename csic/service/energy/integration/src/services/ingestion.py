"""
Telemetry Ingestion Service

Handles real-time and batch ingestion of energy telemetry data
from mining operations via Kafka and HTTP endpoints.
"""

import asyncio
import logging
from datetime import datetime
from typing import Any, Optional

from kafka import KafkaConsumer, KafkaProducer
from kafka.errors import KafkaError

from src.core.config import settings
from src.models.energy_metrics import EnergyMetric

logger = logging.getLogger(__name__)


class TelemetryConsumer:
    """
    Kafka consumer for real-time telemetry ingestion.

    Consumes messages from the mining.telemetry topic and
    processes energy readings asynchronously.
    """

    def __init__(self):
        self.consumer: Optional[KafkaConsumer] = None
        self.producer: Optional[KafkaProducer] = None
        self.running = False
        self._task: Optional[asyncio.Task] = None

    async def start(self):
        """Start the Kafka consumer."""
        try:
            self.consumer = KafkaConsumer(
                settings.kafka.topics.get("telemetry", "mining.telemetry"),
                bootstrap_servers=settings.kafka.brokers,
                group_id=settings.kafka.consumer_group,
                auto_offset_reset="latest",
                enable_auto_commit=True,
                value_deserializer=lambda x: x.decode("utf-8"),
            )

            self.producer = KafkaProducer(
                bootstrap_servers=settings.kafka.brokers,
                value_serializer=lambda x: x.encode("utf-8"),
            )

            self.running = True
            self._task = asyncio.create_task(self._consume_loop())

            logger.info("Telemetry consumer started")

        except KafkaError as e:
            logger.error(f"Failed to start telemetry consumer: {e}")
            raise

    async def stop(self):
        """Stop the Kafka consumer."""
        self.running = False

        if self._task:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass

        if self.consumer:
            self.consumer.close()

        if self.producer:
            self.producer.close()

        logger.info("Telemetry consumer stopped")

    async def _consume_loop(self):
        """Main consumption loop."""
        while self.running:
            try:
                # Poll for messages (non-blocking)
                messages = self.consumer.poll(timeout_ms=1000)

                for topic_partition, msgs in messages.items():
                    for message in msgs:
                        await self._process_message(message.value)

            except Exception as e:
                logger.error(f"Error in consume loop: {e}")
                await asyncio.sleep(1)

    async def _process_message(self, message: str):
        """
        Process a single telemetry message.

        Expected format:
        {
            "site_id": "uuid",
            "timestamp": "iso8601",
            "reading": {
                "consumption_kwh": float,
                "hashrate_th": float,
                "power_sources": {"hydro": 0.6, "coal": 0.4},
                ...
            }
        }
        """
        import json

        try:
            data = json.loads(message)

            # Extract reading
            reading = data.get("reading", {})
            site_id = data.get("site_id")

            if not site_id:
                logger.warning("Message missing site_id")
                return

            # Calculate derived metrics
            co2_grams = self._calculate_co2(
                reading.get("consumption_kwh", 0),
                reading.get("power_sources"),
            )

            efficiency = self._calculate_efficiency(
                reading.get("watts_per_th"),
                reading.get("hashrate_th"),
                reading.get("consumption_kwh"),
            )

            # Log processed reading
            logger.debug(
                f"Processed telemetry for site {site_id}: "
                f"consumption={reading.get('consumption_kwh')} kWh, "
                f"CO2={co2_grams} g"
            )

            # Here we would save to database
            # For now, just log the processed data
            return {
                "site_id": site_id,
                "consumption_kwh": reading.get("consumption_kwh"),
                "co2_grams": co2_grams,
                "efficiency": efficiency,
            }

        except json.JSONDecodeError as e:
            logger.error(f"Failed to parse message: {e}")
        except Exception as e:
            logger.error(f"Error processing message: {e}")

    def _calculate_co2(
        self,
        consumption_kwh: float,
        power_sources: Optional[dict[str, float]],
    ) -> float:
        """Calculate CO2 emissions from power source mix."""
        if not power_sources:
            return consumption_kwh * 350  # Default grid average

        total_emissions = 0.0
        factors = {
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

        for source, proportion in power_sources.items():
            factor = factors.get(source, 350)
            total_emissions += consumption_kwh * proportion * factor

        return total_emissions

    def _calculate_efficiency(
        self,
        watts_per_th: Optional[float],
        hashrate_th: Optional[float],
        consumption_kwh: float,
    ) -> Optional[float]:
        """Calculate mining efficiency percentage."""
        if watts_per_th and hashrate_th and hashrate_th > 0:
            theoretical_min = hashrate_th * 20
            actual = watts_per_th * hashrate_th * 0.001
            if actual > 0:
                return min(100, (theoretical_min / actual) * 100)
        return None


class TelemetryPublisher:
    """
    Kafka publisher for sending forecast and alert events.
    """

    def __init__(self):
        self.producer: Optional[KafkaProducer] = None

    async def start(self):
        """Initialize the producer."""
        try:
            self.producer = KafkaProducer(
                bootstrap_servers=settings.kafka.brokers,
                value_serializer=lambda x: x.encode("utf-8"),
            )
            logger.info("Telemetry publisher initialized")
        except KafkaError as e:
            logger.error(f"Failed to initialize publisher: {e}")
            raise

    async def stop(self):
        """Close the producer."""
        if self.producer:
            self.producer.close()

    async def publish_forecast(
        self,
        site_id: str,
        forecast_data: dict[str, Any],
    ):
        """Publish a generated forecast."""
        import json

        if not self.producer:
            await self.start()

        message = json.dumps({
            "event_type": "energy.forecast",
            "site_id": site_id,
            "data": forecast_data,
            "timestamp": datetime.utcnow().isoformat(),
        })

        self.producer.send(
            settings.kafka.topics.get("forecast", "energy.forecast"),
            value=message,
        )

        logger.debug(f"Published forecast for site {site_id}")

    async def publish_alert(
        self,
        alert_type: str,
        site_id: str,
        alert_data: dict[str, Any],
    ):
        """Publish an alert event."""
        import json

        if not self.producer:
            await self.start()

        message = json.dumps({
            "event_type": f"energy.alert.{alert_type}",
            "site_id": site_id,
            "data": alert_data,
            "timestamp": datetime.utcnow().isoformat(),
        })

        self.producer.send(
            "energy.alerts",
            value=message,
        )

        logger.info(f"Published {alert_type} alert for site {site_id}")
