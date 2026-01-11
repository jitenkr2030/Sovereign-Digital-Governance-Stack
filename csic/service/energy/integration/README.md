# Energy Integration Service

## Overview

The Energy Integration Service is a Python-based microservice built with FastAPI that provides comprehensive energy analytics, load forecasting, and carbon footprint tracking for cryptocurrency mining operations. This service enables regulators to monitor mining energy consumption, enforce green energy compliance, and predict grid impact.

## Features

### Core Functionality
- **Energy Metrics Ingestion**: Real-time and batch ingestion of energy consumption data from mining sites
- **Load Forecasting**: ML-based prediction of energy demand using Prophet and Scikit-learn
- **Grid Integration Analysis**: Correlation of mining operations with local grid conditions
- **Carbon Footprint Tracking**: Calculation and reporting of CO2 emissions based on power sources
- **Sustainability Scoring**: Composite scoring of mining operations based on energy sources

### Technical Features
- **Async Architecture**: Built on FastAPI with native AsyncSQLAlchemy support
- **Time-Series Optimization**: TimescaleDB extension for high-volume energy data
- **ML Pipeline**: Prophet for time-series forecasting, Scikit-learn for pattern detection
- **Real-Time Streaming**: Kafka consumer for live telemetry ingestion
- **D3.js Visualizations**: RESTful API serving data for D3.js-based dashboards

## Architecture

### Technology Stack
- **Runtime**: Python 3.10+
- **Framework**: FastAPI
- **Database**: PostgreSQL with TimescaleDB extension
- **ML Libraries**: Prophet, Scikit-learn, Pandas, NumPy
- **Messaging**: Apache Kafka
- **Visualization**: D3.js-compatible REST API

### Directory Structure
```
service/energy/integration/
├── src/
│   ├── api/
│   │   └── endpoints/       # Route definitions
│   ├── core/                # Config, security
│   ├── models/              # SQLAlchemy models
│   ├── services/
│   │   ├── ingestion.py     # Data ingestion logic
│   │   ├── forecasting.py   # ML forecasting models
│   │   ├── carbon.py        # Carbon calculation
│   │   └── grid.py          # Grid integration analysis
│   ├── worker/              # Celery/Arq tasks
│   └── main.py              # Application entry point
├── internal/
│   ├── db/
│   │   └── migrations/      # Database migrations
│   ├── deploy/
│   │   ├── docker-compose.yml
│   │   └── Dockerfile
│   └── monitoring/
│       ├── prometheus.yml
│       └── grafana/
├── tests/                   # Unit tests
├── requirements.txt
└── README.md
```

## API Endpoints

### Energy Metrics
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/ingest/telemetry` | Ingest batch energy telemetry data |
| POST | `/api/v1/ingest/realtime` | Ingest single real-time reading |
| GET | `/api/v1/metrics/site/{id}` | Get energy metrics for a site |
| GET | `/api/v1/metrics/aggregate` | Get aggregated network metrics |

### Forecasting
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/forecast/load` | Get load forecast for specified period |
| GET | `/api/v1/forecast/hashrate` | Get hashrate forecast |
| POST | `/api/v1/forecast/custom` | Generate custom forecast |

### Analytics
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/analytics/sustainability/{entity_id}` | Get sustainability score |
| GET | `/api/v1/analytics/carbon/{entity_id}` | Get carbon footprint report |
| GET | `/api/v1/analytics/grid-impact` | Get grid impact analysis |

### Sites
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sites` | List all mining sites |
| GET | `/api/v1/sites/{id}` | Get site details |
| POST | `/api/v1/sites` | Register new mining site |
| PATCH | `/api/v1/sites/{id}` | Update site configuration |

## Configuration

### Environment Variables
```env
# Application
ENVIRONMENT=development
DEBUG=false
API_HOST=0.0.0.0
API_PORT=8000

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=csic_energy
DATABASE_PASSWORD=your_password
DATABASE_NAME=csic_energy
TIMESCALEDB_ENABLED=true

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_ENERGY=mining.telemetry

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Forecasting
FORECAST_DAYS=30
MODEL_RETRAIN_INTERVAL=168  # Hours (weekly)
```

## Getting Started

### Prerequisites
- Python 3.10+
- PostgreSQL 14+ with TimescaleDB extension
- Kafka 3.x
- Redis 6+

### Installation
```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # Linux/Mac
# or
.\venv\Scripts\activate  # Windows

# Install dependencies
pip install -r requirements.txt

# Run database migrations
alembic upgrade head

# Start development server
uvicorn src.main:app --reload

# Run background workers
celery -A worker.celery worker --loglevel=info
```

### Docker Deployment
```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f energy-service
```

## Integration Points

### Kafka Topics
- `mining.telemetry` - Raw energy telemetry from mining operations
- `energy.forecast` - Generated forecasts for downstream services

### Neo4j Relationships
- `(MiningSite)-[:CONSUMES_ENERGY]->(Grid)`
- `(MiningSite)-[:USES_SOURCE]->(PowerSource)`

### Prometheus Metrics
- `mining_energy_consumption_kwh` - Total energy consumed
- `mining_hashrate_th` - Network hashrate
- `carbon_emission_grams` - CO2 emissions
- `sustainability_score` - Green energy score

## License

Internal CSIC Platform Component
