"""
Configuration management for Energy Integration Service.

Loads configuration from config.yaml with environment variable override support.
"""

import os
from pathlib import Path
from typing import Any, Dict, List, Optional

import yaml
from pydantic import BaseModel, Field
from pydantic_settings import BaseSettings


class AppConfig(BaseModel):
    """Application configuration."""
    name: str = "energy-integration-service"
    environment: str = "development"
    version: str = "1.0.0"
    debug: bool = False


class ServerConfig(BaseModel):
    """Server configuration."""
    host: str = "0.0.0.0"
    port: int = 8000
    reload: bool = True


class DatabaseConfig(BaseModel):
    """Database configuration."""
    host: str = "localhost"
    port: int = 5432
    username: str = "csic_energy"
    password: str = "password"
    name: str = "csic_energy"
    pool_size: int = 20
    max_overflow: int = 10
    echo: bool = False
    timescaledb: Dict[str, Any] = {"enabled": True}

    @property
    def async_url(self) -> str:
        """Generate async database URL."""
        return f"postgresql+asyncpg://{self.username}:{self.password}@{self.host}:{self.port}/{self.name}"

    @property
    def sync_url(self) -> str:
        """Generate sync database URL."""
        return f"postgresql://{self.username}:{self.password}@{self.host}:{self.port}/{self.name}"


class RedisConfig(BaseModel):
    """Redis configuration."""
    host: str = "localhost"
    port: int = 6379
    password: str = ""
    db: int = 0
    key_prefix: str = "csic:energy:"


class KafkaConfig(BaseModel):
    """Kafka configuration."""
    brokers: List[str] = ["localhost:9092"]
    consumer_group: str = "energy-consumer"
    topics: Dict[str, str] = {
        "telemetry": "mining.telemetry",
        "forecast": "energy.forecast",
    }


class PowerSourceConfig(BaseModel):
    """Power source configuration."""
    id: str
    name: str
    carbon_factor_grams_per_kwh: float
    renewable: Optional[bool] = None


class EnergyConfig(BaseModel):
    """Energy-specific configuration."""
    power_sources: List[PowerSourceConfig] = []
    grid_impact: Dict[str, Any] = {
        "high_threshold_percent": 10,
        "medium_threshold_percent": 5,
        "peak_hour_detection": True,
    }


class ForecastingConfig(BaseModel):
    """Forecasting configuration."""
    default_horizon_days: int = 30
    max_horizon_days: int = 365
    retrain_interval_hours: int = 168
    min_training_data_days: int = 30
    confidence_level: float = 0.95
    prophet: Dict[str, Any] = {
        "yearly_seasonality": True,
        "weekly_seasonality": True,
        "daily_seasonality": False,
        "seasonality_mode": "multiplicative",
        "changepoint_prior_scale": 0.05,
    }
    anomaly_detection: Dict[str, Any] = {
        "enabled": True,
        "z_threshold": 3.0,
        "window_size_hours": 24,
    }


class SustainabilityConfig(BaseModel):
    """Sustainability scoring configuration."""
    weights: Dict[str, float] = {
        "renewable_percentage": 0.4,
        "carbon_intensity": 0.3,
        "efficiency": 0.2,
        "grid_strain": 0.1,
    }
    thresholds: Dict[str, int] = {
        "excellent": 80,
        "good": 60,
        "moderate": 40,
        "poor": 20,
    }


class CarbonConfig(BaseModel):
    """Carbon calculation configuration."""
    include_transmission_loss: bool = True
    transmission_loss_factor: float = 0.07


class AlertingConfig(BaseModel):
    """Alerting configuration."""
    enabled: bool = True
    thresholds: Dict[str, float] = {
        "consumption_spike_percent": 50,
        "carbon_increase_percent": 25,
        "efficiency_drop_percent": 20,
    }


class LoggingConfig(BaseModel):
    """Logging configuration."""
    level: str = "INFO"
    format: str = "json"
    include_caller: bool = True


class MetricsConfig(BaseModel):
    """Metrics configuration."""
    enabled: bool = True
    endpoint: str = "/metrics"
    port: int = 9090
    prefix: str = "csic_energy"


class Settings(BaseModel):
    """Main settings class."""
    app: AppConfig = Field(default_factory=AppConfig)
    server: ServerConfig = Field(default_factory=ServerConfig)
    database: DatabaseConfig = Field(default_factory=DatabaseConfig)
    redis: RedisConfig = Field(default_factory=RedisConfig)
    kafka: KafkaConfig = Field(default_factory=KafkaConfig)
    energy: EnergyConfig = Field(default_factory=EnergyConfig)
    forecasting: ForecastingConfig = Field(default_factory=ForecastingConfig)
    sustainability: SustainabilityConfig = Field(default_factory=SustainabilityConfig)
    carbon: CarbonConfig = Field(default_factory=CarbonConfig)
    alerting: AlertingConfig = Field(default_factory=AlertingConfig)
    logging: LoggingConfig = Field(default_factory=LoggingConfig)
    metrics: MetricsConfig = Field(default_factory=MetricsConfig)

    # Computed properties
    @property
    def CORS_ORIGINS(self) -> List[str]:
        """Get CORS origins from app environment."""
        env_origins = os.getenv("CORS_ORIGINS", "")
        if env_origins:
            return [o.strip() for o in env_origins.split(",")]
        return ["http://localhost:3000", "http://localhost:5173"]

    @property
    def LOG_LEVEL(self) -> str:
        """Get log level."""
        return os.getenv("LOG_LEVEL", self.logging.level)

    @property
    def LOG_FORMAT(self) -> str:
        """Get log format."""
        return "%(asctime)s - %(name)s - %(levelname)s - %(message)s" if not self.logging.include_caller else "%(asctime)s - %(name)s - %(levelname)s - %(funcName)s:%(lineno)d - %(message)s"

    @property
    def SERVER_HOST(self) -> str:
        """Get server host."""
        return os.getenv("SERVER_HOST", self.server.host)

    @property
    def SERVER_PORT(self) -> int:
        """Get server port."""
        env_port = os.getenv("SERVER_PORT")
        if env_port:
            return int(env_port)
        return self.server.port

    @property
    def DEBUG(self) -> bool:
        """Get debug mode."""
        env_debug = os.getenv("DEBUG")
        if env_debug:
            return env_debug.lower() in ("true", "1", "yes")
        return self.app.debug


def load_config(config_path: str | None = None) -> Settings:
    """
    Load configuration from YAML file.

    Args:
        config_path: Path to config.yaml file. If None, uses default path.

    Returns:
        Settings object with loaded configuration.
    """
    if config_path is None:
        # Look for config.yaml in current directory or parent directories
        current_path = Path(__file__).parent
        config_path = current_path / "config.yaml"
        if not config_path.exists():
            config_path = current_path.parent / "config.yaml"
        if not config_path.exists():
            # Return default settings if no config file found
            return Settings()

    with open(config_path, "r") as f:
        config_data = yaml.safe_load(f)

    return Settings(**config_data)


# Global settings instance
settings = load_config()
