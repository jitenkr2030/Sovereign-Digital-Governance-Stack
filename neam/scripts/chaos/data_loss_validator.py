#!/usr/bin/env python3
"""
Data Loss Validation Framework for Chaos Engineering

This module provides comprehensive data loss measurement, validation, and alerting
capabilities for testing system resilience under chaos conditions.
"""

import json
import time
import hashlib
import uuid
import asyncio
import logging
from dataclasses import dataclass, field, asdict
from datetime import datetime, timedelta
from enum import Enum
from typing import Dict, List, Optional, Callable, Any, Tuple
from collections import defaultdict
from abc import ABC, abstractmethod
import redis
import psycopg2
from psycopg2 import sql
import boto3
from botocore.exceptions import ClientError
import kafka
from kafka import KafkaProducer, KafkaConsumer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class DataLossType(Enum):
    """Types of data loss that can occur."""
    TRANSIENT = "transient"
    PERMANENT = "permanent"
    CORRUPTION = "corruption"
    DUPLICATION = "duplication"
    ORDERING_VIOLATION = "ordering_violation"


class ValidationStatus(Enum):
    """Status of data loss validation."""
    PASSED = "passed"
    FAILED = "failed"
    INCONCLUSIVE = "inconclusive"
    PENDING = "pending"
    RUNNING = "running"


@dataclass
class DataRecord:
    """Represents a data record for tracking."""
    id: str
    data: dict
    checksum: str
    created_at: str
    updated_at: str
    version: int = 1
    deleted: bool = False
    deleted_at: Optional[str] = None
    source: str = ""
    metadata: dict = field(default_factory=dict)

    @classmethod
    def create(cls, data: dict, source: str = "") -> 'DataRecord':
        """Create a new data record."""
        now = datetime.utcnow().isoformat()
        record_id = str(uuid.uuid4())
        checksum = cls._compute_checksum(data)
        
        return cls(
            id=record_id,
            data=data,
            checksum=checksum,
            created_at=now,
            updated_at=now,
            source=source
        )

    @staticmethod
    def _compute_checksum(data: dict) -> str:
        """Compute checksum for data."""
        content = json.dumps(data, sort_keys=True)
        return hashlib.sha256(content.encode()).hexdigest()

    def verify_integrity(self) -> bool:
        """Verify data integrity using checksum."""
        return self.checksum == self._compute_checksum(self.data)

    def to_dict(self) -> dict:
        return asdict(self)


@dataclass
class DataLossMeasurement:
    """Measurement of data loss during chaos experiment."""
    experiment_id: str
    measurement_id: str
    start_time: str
    end_time: str
    data_type: str
    total_records: int
    records_written: int
    records_read: int
    records_updated: int
    records_deleted: int
    records_lost: int
    records_corrupted: int
    records_duplicated: int
    ordering_violations: int
    loss_percentage: float
    corruption_percentage: float
    duplication_percentage: float
    validation_status: ValidationStatus
    error_details: List[str] = field(default_factory=list)

    def to_dict(self) -> dict:
        result = asdict(self)
        result['validation_status'] = self.validation_status.value
        return result


@dataclass
class DataLossAlert:
    """Alert for data loss detection."""
    alert_id: str
    experiment_id: str
    timestamp: str
    severity: str  # critical, high, medium, low
    data_loss_type: DataLossType
    measurement_id: str
    affected_records: int
    loss_percentage: float
    description: str
    recommendations: List[str] = field(default_factory=list)
    acknowledged: bool = False
    acknowledged_by: Optional[str] = None
    acknowledged_at: Optional[str] = None

    def to_dict(self) -> dict:
        result = asdict(self)
        result['data_loss_type'] = self.data_loss_type.value
        return result


class DataStore(ABC):
    """Abstract base class for data store operations."""
    
    @abstractmethod
    def connect(self) -> None:
        pass
    
    @abstractmethod
    def disconnect(self) -> None:
        pass
    
    @abstractmethod
    def write(self, record: DataRecord) -> bool:
        pass
    
    @abstractmethod
    def read(self, record_id: str) -> Optional[DataRecord]:
        pass
    
    @abstractmethod
    def read_all(self, limit: int = 1000) -> List[DataRecord]:
        pass
    
    @abstractmethod
    def update(self, record: DataRecord) -> bool:
        pass
    
    @abstractmethod
    def delete(self, record_id: str) -> bool:
        pass
    
    @abstractmethod
    def verify_all(self) -> Tuple[int, int]:
        """Verify all records, return (total, corrupted)."""
        pass
    
    @abstractmethod
    def count(self) -> int:
        pass


class RedisDataStore(DataStore):
    """Redis implementation of data store."""
    
    def __init__(
        self,
        host: str = "localhost",
        port: int = 6379,
        db: int = 0,
        password: Optional[str] = None,
        key_prefix: str = "chaos_test:"
    ):
        self.host = host
        self.port = port
        self.db = db
        self.password = password
        self.key_prefix = key_prefix
        self.client: Optional[redis.Redis] = None
    
    def connect(self) -> None:
        self.client = redis.Redis(
            host=self.host,
            port=self.port,
            db=self.db,
            password=self.password,
            decode_responses=True
        )
        self.client.ping()
        logger.info(f"Connected to Redis at {self.host}:{self.port}")
    
    def disconnect(self) -> None:
        if self.client:
            self.client.close()
            logger.info("Disconnected from Redis")
    
    def _key(self, record_id: str) -> str:
        return f"{self.key_prefix}{record_id}"
    
    def write(self, record: DataRecord) -> bool:
        try:
            self.client.set(
                self._key(record.id),
                json.dumps(record.to_dict()),
                ex=3600  # 1 hour TTL
            )
            return True
        except Exception as e:
            logger.error(f"Redis write error: {e}")
            return False
    
    def read(self, record_id: str) -> Optional[DataRecord]:
        try:
            data = self.client.get(self._key(record_id))
            if data:
                return DataRecord.from_dict(json.loads(data))
            return None
        except Exception as e:
            logger.error(f"Redis read error: {e}")
            return None
    
    def read_all(self, limit: int = 1000) -> List[DataRecord]:
        try:
            keys = self.client.keys(f"{self.key_prefix}*")[:limit]
            records = []
            for key in keys:
                data = self.client.get(key)
                if data:
                    records.append(DataRecord.from_dict(json.loads(data)))
            return records
        except Exception as e:
            logger.error(f"Redis read_all error: {e}")
            return []
    
    def update(self, record: DataRecord) -> bool:
        try:
            record.updated_at = datetime.utcnow().isoformat()
            record.version += 1
            self.client.set(
                self._key(record.id),
                json.dumps(record.to_dict()),
                ex=3600
            )
            return True
        except Exception as e:
            logger.error(f"Redis update error: {e}")
            return False
    
    def delete(self, record_id: str) -> bool:
        try:
            self.client.delete(self._key(record_id))
            return True
        except Exception as e:
            logger.error(f"Redis delete error: {e}")
            return False
    
    def verify_all(self) -> Tuple[int, int]:
        try:
            keys = self.client.keys(f"{self.key_prefix}*")
            total = len(keys)
            corrupted = 0
            
            for key in keys:
                data = self.client.get(key)
                if data:
                    record = DataRecord.from_dict(json.loads(data))
                    if not record.verify_integrity():
                        corrupted += 1
            
            return total, corrupted
        except Exception as e:
            logger.error(f"Redis verify error: {e}")
            return 0, 0
    
    def count(self) -> int:
        try:
            return len(self.client.keys(f"{self.key_prefix}*"))
        except Exception as e:
            logger.error(f"Redis count error: {e}")
            return 0


class PostgreSQLDataStore(DataStore):
    """PostgreSQL implementation of data store."""
    
    def __init__(
        self,
        host: str = "localhost",
        port: int = 5432,
        database: str = "chaos_test",
        user: str = "postgres",
        password: str = "",
        table_name: str = "chaos_data_records"
    ):
        self.host = host
        self.port = port
        self.database = database
        self.user = user
        self.password = password
        self.table_name = table_name
        self.conn: Optional[psycopg2.extensions.connection] = None
    
    def connect(self) -> None:
        self.conn = psycopg2.connect(
            host=self.host,
            port=self.port,
            database=self.database,
            user=self.user,
            password=self.password
        )
        self.conn.autocommit = False
        self._create_table()
        logger.info(f"Connected to PostgreSQL at {self.host}:{self.port}")
    
    def _create_table(self) -> None:
        """Create the data table if not exists."""
        query = f"""
        CREATE TABLE IF NOT EXISTS {self.table_name} (
            id VARCHAR(36) PRIMARY KEY,
            data JSONB NOT NULL,
            checksum VARCHAR(64) NOT NULL,
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            version INTEGER DEFAULT 1,
            deleted BOOLEAN DEFAULT FALSE,
            deleted_at TIMESTAMP,
            source VARCHAR(255),
            metadata JSONB
        )
        """
        with self.conn.cursor() as cur:
            cur.execute(query)
        self.conn.commit()
    
    def disconnect(self) -> None:
        if self.conn:
            self.conn.close()
            logger.info("Disconnected from PostgreSQL")
    
    def write(self, record: DataRecord) -> bool:
        try:
            query = f"""
            INSERT INTO {self.table_name}
            (id, data, checksum, created_at, updated_at, version, source, metadata)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            """
            with self.conn.cursor() as cur:
                cur.execute(query, (
                    record.id,
                    json.dumps(record.data),
                    record.checksum,
                    record.created_at,
                    record.updated_at,
                    record.version,
                    record.source,
                    json.dumps(record.metadata)
                ))
            self.conn.commit()
            return True
        except Exception as e:
            logger.error(f"PostgreSQL write error: {e}")
            self.conn.rollback()
            return False
    
    def read(self, record_id: str) -> Optional[DataRecord]:
        try:
            query = f"""
            SELECT id, data, checksum, created_at, updated_at,
                   version, deleted, deleted_at, source, metadata
            FROM {self.table_name}
            WHERE id = %s AND deleted = FALSE
            """
            with self.conn.cursor() as cur:
                cur.execute(query, (record_id,))
                row = cur.fetchone()
            
            if row:
                return DataRecord(
                    id=row[0],
                    data=row[1],
                    checksum=row[2],
                    created_at=row[3].isoformat() if isinstance(row[3], datetime) else str(row[3]),
                    updated_at=row[4].isoformat() if isinstance(row[4], datetime) else str(row[4]),
                    version=row[5],
                    deleted=row[6],
                    deleted_at=row[7].isoformat() if row[7] and isinstance(row[7], datetime) else str(row[7]) if row[7] else None,
                    source=row[8],
                    metadata=row[9] if row[9] else {}
                )
            return None
        except Exception as e:
            logger.error(f"PostgreSQL read error: {e}")
            return None
    
    def read_all(self, limit: int = 1000) -> List[DataRecord]:
        try:
            query = f"""
            SELECT id, data, checksum, created_at, updated_at,
                   version, deleted, deleted_at, source, metadata
            FROM {self.table_name}
            WHERE deleted = FALSE
            ORDER BY created_at DESC
            LIMIT %s
            """
            with self.conn.cursor() as cur:
                cur.execute(query, (limit,))
                rows = cur.fetchall()
            
            records = []
            for row in rows:
                records.append(DataRecord(
                    id=row[0],
                    data=row[1],
                    checksum=row[2],
                    created_at=row[3].isoformat() if isinstance(row[3], datetime) else str(row[3]),
                    updated_at=row[4].isoformat() if isinstance(row[4], datetime) else str(row[4]),
                    version=row[5],
                    deleted=row[6],
                    deleted_at=row[7].isoformat() if row[7] and isinstance(row[7], datetime) else str(row[7]) if row[7] else None,
                    source=row[8],
                    metadata=row[9] if row[9] else {}
                ))
            return records
        except Exception as e:
            logger.error(f"PostgreSQL read_all error: {e}")
            return []
    
    def update(self, record: DataRecord) -> bool:
        try:
            record.updated_at = datetime.utcnow().isoformat()
            record.version += 1
            query = f"""
            UPDATE {self.table_name}
            SET data = %s, checksum = %s, updated_at = %s,
                version = %s, metadata = %s
            WHERE id = %s
            """
            with self.conn.cursor() as cur:
                cur.execute(query, (
                    json.dumps(record.data),
                    record.checksum,
                    record.updated_at,
                    record.version,
                    json.dumps(record.metadata),
                    record.id
                ))
            self.conn.commit()
            return cur.rowcount > 0
        except Exception as e:
            logger.error(f"PostgreSQL update error: {e}")
            self.conn.rollback()
            return False
    
    def delete(self, record_id: str) -> bool:
        try:
            now = datetime.utcnow().isoformat()
            query = f"""
            UPDATE {self.table_name}
            SET deleted = TRUE, deleted_at = %s
            WHERE id = %s
            """
            with self.conn.cursor() as cur:
                cur.execute(query, (now, record_id))
            self.conn.commit()
            return cur.rowcount > 0
        except Exception as e:
            logger.error(f"PostgreSQL delete error: {e}")
            self.conn.rollback()
            return False
    
    def verify_all(self) -> Tuple[int, int]:
        try:
            query = f"""
            SELECT id, data, checksum FROM {self.table_name}
            WHERE deleted = FALSE
            """
            with self.conn.cursor() as cur:
                cur.execute(query)
                rows = cur.fetchall()
            
            total = len(rows)
            corrupted = 0
            
            for row in rows:
                record = DataRecord(
                    id=row[0],
                    data=row[1],
                    checksum=row[2],
                    created_at="",
                    updated_at=""
                )
                if not record.verify_integrity():
                    corrupted += 1
            
            return total, corrupted
        except Exception as e:
            logger.error(f"PostgreSQL verify error: {e}")
            return 0, 0
    
    def count(self) -> int:
        try:
            query = f"SELECT COUNT(*) FROM {self.table_name} WHERE deleted = FALSE"
            with self.conn.cursor() as cur:
                cur.execute(query)
                return cur.fetchone()[0]
        except Exception as e:
            logger.error(f"PostgreSQL count error: {e}")
            return 0


class KafkaDataStore(DataStore):
    """Kafka implementation for message-based data store."""
    
    def __init__(
        self,
        bootstrap_servers: str = "localhost:9092",
        topic: str = "chaos-test-events",
        consumer_group: str = "chaos-test-consumer"
    ):
        self.bootstrap_servers = bootstrap_servers
        self.topic = topic
        self.consumer_group = consumer_group
        self.producer: Optional[KafkaProducer] = None
        self.consumer: Optional[KafkaConsumer] = None
        self.written_records: Dict[str, DataRecord] = {}
    
    def connect(self) -> None:
        self.producer = KafkaProducer(
            bootstrap_servers=self.bootstrap_servers,
            value_serializer=lambda v: json.dumps(v).encode(),
            key_serializer=lambda k: k.encode() if k else None
        )
        logger.info(f"Connected to Kafka at {self.bootstrap_servers}")
    
    def disconnect(self) -> None:
        if self.producer:
            self.producer.close()
        if self.consumer:
            self.consumer.close()
        logger.info("Disconnected from Kafka")
    
    def write(self, record: DataRecord) -> bool:
        try:
            future = self.producer.send(
                self.topic,
                key=record.id,
                value=record.to_dict()
            )
            future.get(timeout=10)
            self.written_records[record.id] = record
            return True
        except Exception as e:
            logger.error(f"Kafka write error: {e}")
            return False
    
    def read(self, record_id: str) -> Optional[DataRecord]:
        return self.written_records.get(record_id)
    
    def read_all(self, limit: int = 1000) -> List[DataRecord]:
        return list(self.written_records.values())[:limit]
    
    def update(self, record: DataRecord) -> bool:
        return self.write(record)
    
    def delete(self, record_id: str) -> bool:
        if record_id in self.written_records:
            del self.written_records[record_id]
            return True
        return False
    
    def verify_all(self) -> Tuple[int, int]:
        total = len(self.written_records)
        corrupted = 0
        for record in self.written_records.values():
            if not record.verify_integrity():
                corrupted += 1
        return total, corrupted
    
    def count(self) -> int:
        return len(self.written_records)


class DataLossValidator:
    """Main class for validating data loss during chaos experiments."""
    
    def __init__(self, experiment_id: str):
        self.experiment_id = experiment_id
        self.datastores: Dict[str, DataStore] = {}
        self.measurements: List[DataLossMeasurement] = []
        self.alerts: List[DataLossAlert] = []
        self.start_time: Optional[str] = None
        self.end_time: Optional[str] = None
        self.callbacks: List[Callable] = []
        
        # Thresholds for alerts
        self.thresholds = {
            'critical_loss_percentage': 5.0,
            'high_loss_percentage': 1.0,
            'medium_loss_percentage': 0.5,
            'low_loss_percentage': 0.1,
            'critical_corruption_percentage': 1.0,
            'high_corruption_percentage': 0.5
        }
    
    def register_datastore(self, name: str, store: DataStore) -> None:
        """Register a data store for validation."""
        self.datastores[name] = store
        logger.info(f"Registered data store: {name}")
    
    def register_alert_callback(self, callback: Callable[[DataLossAlert], None]) -> None:
        """Register a callback for alert handling."""
        self.callbacks.append(callback)
    
    def _trigger_alert(self, alert: DataLossAlert) -> None:
        """Trigger alert and invoke callbacks."""
        self.alerts.append(alert)
        logger.warning(f"Data loss alert: {alert.description}")
        for callback in self.callbacks:
            try:
                callback(alert)
            except Exception as e:
                logger.error(f"Alert callback error: {e}")
    
    def start_measurement(self) -> None:
        """Start data loss measurement."""
        self.start_time = datetime.utcnow().isoformat()
        
        # Take initial snapshots of all data stores
        for name, store in self.datastores.items():
            try:
                store.connect()
            except Exception as e:
                logger.error(f"Failed to connect to {name}: {e}")
        
        logger.info(f"Started data loss measurement for experiment {self.experiment_id}")
    
    def write_test_data(
        self,
        store_name: str,
        count: int = 100,
        data_template: Optional[dict] = None
    ) -> List[DataRecord]:
        """Write test data to a data store."""
        store = self.datastores.get(store_name)
        if not store:
            logger.error(f"Data store {store_name} not found")
            return []
        
        records = []
        template = data_template or {
            "experiment": self.experiment_id,
            "timestamp": datetime.utcnow().isoformat()
        }
        
        for i in range(count):
            data = template.copy()
            data["sequence"] = i
            data["random_value"] = uuid.uuid4().hex[:8]
            
            record = DataRecord.create(data, source=store_name)
            if store.write(record):
                records.append(record)
        
        logger.info(f"Wrote {len(records)} records to {store_name}")
        return records
    
    def measure_data_loss(self, data_type: str = "general") -> DataLossMeasurement:
        """Measure data loss across all data stores."""
        if not self.start_time:
            raise RuntimeError("Measurement not started")
        
        self.end_time = datetime.utcnow().isoformat()
        measurement_id = str(uuid.uuid4())
        
        total_records = 0
        records_written = 0
        records_read = 0
        records_updated = 0
        records_deleted = 0
        records_lost = 0
        records_corrupted = 0
        records_duplicated = 0
        ordering_violations = 0
        error_details = []
        
        for name, store in self.datastores.items():
            try:
                # Count total records
                count = store.count()
                total_records += count
                
                # Verify integrity
                verified_total, corrupted = store.verify_all()
                records_read += verified_total
                records_corrupted += corrupted
                
                if corrupted > 0:
                    error_details.append(
                        f"Store {name}: {corrupted} corrupted records out of {verified_total}"
                    )
                
                logger.info(f"Store {name}: {count} records, {corrupted} corrupted")
                
            except Exception as e:
                error_details.append(f"Store {name}: {str(e)}")
                logger.error(f"Error measuring {name}: {e}")
        
        # Calculate percentages
        loss_percentage = (records_lost / max(total_records, 1)) * 100
        corruption_percentage = (records_corrupted / max(total_records, 1)) * 100
        duplication_percentage = (records_duplicated / max(total_records, 1)) * 100
        
        # Determine validation status
        if loss_percentage > self.thresholds['critical_loss_percentage']:
            status = ValidationStatus.FAILED
        elif loss_percentage > self.thresholds['high_loss_percentage']:
            status = ValidationStatus.FAILED
        elif corruption_percentage > self.thresholds['critical_corruption_percentage']:
            status = ValidationStatus.FAILED
        else:
            status = ValidationStatus.PASSED
        
        measurement = DataLossMeasurement(
            experiment_id=self.experiment_id,
            measurement_id=measurement_id,
            start_time=self.start_time,
            end_time=self.end_time,
            data_type=data_type,
            total_records=total_records,
            records_written=records_written,
            records_read=records_read,
            records_updated=records_updated,
            records_deleted=records_deleted,
            records_lost=records_lost,
            records_corrupted=records_corrupted,
            records_duplicated=records_duplicated,
            ordering_violations=ordering_violations,
            loss_percentage=loss_percentage,
            corruption_percentage=corruption_percentage,
            duplication_percentage=duplication_percentage,
            validation_status=status,
            error_details=error_details
        )
        
        self.measurements.append(measurement)
        
        # Generate alerts if thresholds exceeded
        self._check_thresholds(measurement)
        
        logger.info(f"Data loss measurement complete: {status.value}, "
                   f"loss={loss_percentage:.2f}%, corruption={corruption_percentage:.2f}%")
        
        return measurement
    
    def _check_thresholds(self, measurement: DataLossMeasurement) -> None:
        """Check thresholds and generate alerts."""
        # Loss percentage alerts
        if measurement.loss_percentage >= self.thresholds['critical_loss_percentage']:
            self._trigger_alert(DataLossAlert(
                alert_id=str(uuid.uuid4()),
                experiment_id=self.experiment_id,
                timestamp=datetime.utcnow().isoformat(),
                severity="critical",
                data_loss_type=DataLossType.PERMANENT,
                measurement_id=measurement.measurement_id,
                affected_records=measurement.records_lost,
                loss_percentage=measurement.loss_percentage,
                description=f"Critical data loss detected: {measurement.loss_percentage:.2f}%",
                recommendations=[
                    "Review recent chaos experiment parameters",
                    "Check storage system health",
                    "Verify backup restoration procedures",
                    "Analyze error logs for root cause"
                ]
            ))
        elif measurement.loss_percentage >= self.thresholds['high_loss_percentage']:
            self._trigger_alert(DataLossAlert(
                alert_id=str(uuid.uuid4()),
                experiment_id=self.experiment_id,
                timestamp=datetime.utcnow().isoformat(),
                severity="high",
                data_loss_type=DataLossType.TRANSIENT,
                measurement_id=measurement.measurement_id,
                affected_records=measurement.records_lost,
                loss_percentage=measurement.loss_percentage,
                description=f"High data loss detected: {measurement.loss_percentage:.2f}%"
            ))
        
        # Corruption alerts
        if measurement.corruption_percentage >= self.thresholds['critical_corruption_percentage']:
            self._trigger_alert(DataLossAlert(
                alert_id=str(uuid.uuid4()),
                experiment_id=self.experiment_id,
                timestamp=datetime.utcnow().isoformat(),
                severity="critical",
                data_loss_type=DataLossType.CORRUPTION,
                measurement_id=measurement.measurement_id,
                affected_records=measurement.records_corrupted,
                loss_percentage=measurement.corruption_percentage,
                description=f"Critical data corruption detected: {measurement.corruption_percentage:.2f}%"
            ))
    
    def cleanup(self) -> None:
        """Clean up data stores."""
        for name, store in self.datastores.items():
            try:
                store.disconnect()
            except Exception as e:
                logger.error(f"Error disconnecting {name}: {e}")
        logger.info("Cleaned up data stores")
    
    def get_measurement_report(self) -> dict:
        """Generate a comprehensive measurement report."""
        return {
            "experiment_id": self.experiment_id,
            "start_time": self.start_time,
            "end_time": self.end_time,
            "measurements": [m.to_dict() for m in self.measurements],
            "alerts": [a.to_dict() for a in self.alerts],
            "summary": {
                "total_measurements": len(self.measurements),
                "passed_measurements": sum(
                    1 for m in self.measurements
                    if m.validation_status == ValidationStatus.PASSED
                ),
                "failed_measurements": sum(
                    1 for m in self.measurements
                    if m.validation_status == ValidationStatus.FAILED
                ),
                "total_alerts": len(self.alerts),
                "critical_alerts": sum(
                    1 for a in self.alerts if a.severity == "critical"
                )
            }
        }


def create_data_loss_validator(experiment_id: str) -> DataLossValidator:
    """Factory function to create a configured data loss validator."""
    validator = DataLossValidator(experiment_id)
    
    # Configure with environment variables or defaults
    validator.register_datastore("redis", RedisDataStore(
        host=os.getenv("REDIS_HOST", "localhost"),
        port=int(os.getenv("REDIS_PORT", 6379))
    ))
    
    validator.register_datastore("postgres", PostgreSQLDataStore(
        host=os.getenv("POSTGRES_HOST", "localhost"),
        port=int(os.getenv("POSTGRES_PORT", 5432)),
        database=os.getenv("POSTGRES_DB", "chaos_test"),
        user=os.getenv("POSTGRES_USER", "postgres"),
        password=os.getenv("POSTGRES_PASSWORD", "")
    ))
    
    validator.register_datastore("kafka", KafkaDataStore(
        bootstrap_servers=os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
        topic=os.getenv("KAFKA_TOPIC", "chaos-test-events")
    ))
    
    return validator


if __name__ == "__main__":
    import os
    import sys
    
    # Example usage
    experiment_id = sys.argv[1] if len(sys.argv) > 1 else "test-experiment"
    
    validator = create_data_loss_validator(experiment_id)
    
    # Start measurement
    validator.start_measurement()
    
    # Write test data
    validator.write_test_data("redis", count=100)
    validator.write_test_data("postgres", count=100)
    
    # Measure data loss
    measurement = validator.measure_data_loss()
    
    # Print report
    print(json.dumps(validator.get_measurement_report(), indent=2))
    
    # Cleanup
    validator.cleanup()
