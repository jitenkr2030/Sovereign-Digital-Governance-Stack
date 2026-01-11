# NEAM Platform Test Configuration
# Pytest configuration and fixtures

import os
import sys
import pytest
from pathlib import Path

# Add the project root to the Python path
project_root = Path(__file__).parent.parent.parent
sys.path.insert(0, str(project_root))

# Set environment variables for testing
os.environ["NEAM_ENVIRONMENT"] = "test"
os.environ["NEAM_DEBUG"] = "true"
os.environ["CLICKHOUSE_HOST"] = "localhost"
os.environ["POSTGRES_HOST"] = "localhost"
os.environ["REDIS_HOST"] = "localhost:6379"
os.environ["KAFKA_BROKERS"] = "localhost:9092"


@pytest.fixture(scope="session")
def project_root():
    """Return the project root directory."""
    return Path(__file__).parent.parent.parent


@pytest.fixture(scope="session")
def test_data_dir(project_root):
    """Return the path to test data fixtures."""
    return project_root / "tests" / "fixtures"


@pytest.fixture(scope="session")
def config():
    """Create a test configuration."""
    from neam.shared import Config
    config = Config()
    config.Environment = "test"
    config.Debug = True
    config.HTTPPort = 8080
    return config


@pytest.fixture
def mock_logger():
    """Create a mock logger."""
    from neam.shared import Logger
    return Logger("test")


@pytest.fixture
def mock_postgres():
    """Create a mock PostgreSQL connection."""
    from unittest.mock import Mock
    mock = Mock()
    mock.ping.return_value = True
    mock.execute.return_value = None
    mock.query.return_value = []
    return mock


@pytest.fixture
def mock_clickhouse():
    """Create a mock ClickHouse connection."""
    from unittest.mock import Mock
    mock = Mock()
    mock.ping.return_value = True
    mock.query.return_value = []
    mock.exec.return_value = None
    return mock


@pytest.fixture
def mock_redis():
    """Create a mock Redis connection."""
    from unittest.mock import Mock
    mock = Mock()
    mock.ping.return_value = True
    mock.get.return_value = None
    mock.set.return_value = True
    return mock


@pytest.fixture
def sample_entities(test_data_dir):
    """Load sample entity data for testing."""
    import json
    with open(test_data_dir / "sample_entities.json") as f:
        return json.load(f)


@pytest.fixture
def sample_features(test_data_dir):
    """Load sample feature data for testing."""
    import json
    with open(test_data_dir / "sample_features.json") as f:
        return json.load(f)


@pytest.fixture
def sample_transactions(test_data_dir):
    """Load sample transaction data for testing."""
    import json
    with open(test_data_dir / "sample_transactions.json") as f:
        return json.load(f)


@pytest.fixture
def temp_dir(tmp_path):
    """Create a temporary directory for test artifacts."""
    return tmp_path
