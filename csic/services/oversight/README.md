# CSIC Exchange Oversight Service

A high-performance microservices solution for monitoring cryptocurrency exchanges, detecting market abuse patterns, and calculating exchange health scores in real-time.

## Overview

The Exchange Oversight Service is a critical component of the CSIC (Crypto State Infrastructure Contractor) platform. It provides:

- **Live Trade Monitoring**: Real-time ingestion and processing of trade events from connected exchanges
- **Market Abuse Detection**: Automated detection of wash trading, price manipulation, spoofing, and other market abuse patterns
- **Exchange Health Scoring**: Continuous monitoring of exchange performance, latency, and reliability
- **Regulatory Reporting**: Generation of compliance reports for regulatory bodies

## Architecture

The service follows **Hexagonal Architecture (Ports & Adapters)** pattern, ensuring clean separation between core business logic and external dependencies.

```
┌─────────────────────────────────────────────────────────────────┐
│                      Exchange Oversight Service                   │
├─────────────────────────────────────────────────────────────────┤
│  Core (Business Logic)                                          │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────────┐│
│  │ IngestionService│ │ AbuseDetector   │ │ HealthScorerService ││
│  └────────┬────────┘ └────────┬────────┘ └──────────┬──────────┘│
│           │                   │                      │           │
│  ┌────────┴──────────────────┴──────────────────────┴──────────┐│
│  │                     Ports (Interfaces)                       ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐││
│  │  │AlertRepository│ │RuleRepository│ │ TradeAnalyticsEngine │││
│  │  └─────────────┘ └─────────────┘ └─────────────────────────┘││
├─────────────────────────────────────────────────────────────────┤
│  Adapters (External Interfaces)                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────────┐│
│  │   HTTP Server   │ │  gRPC Server    │ │   Kafka Consumer    ││
│  │   (REST API)    │ │  (Internal)     │ │   (Trade Stream)    ││
│  └────────┬────────┘ └────────┬────────┘ └──────────┬──────────┘│
│  ┌────────┴──────────────────┴──────────────────────┴──────────┐│
│  │  PostgreSQL  │  OpenSearch  │  Redis Cache  │  Kafka Bus    ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Features

### Market Abuse Detection

The service implements sophisticated algorithms to detect various forms of market abuse:

- **Wash Trading Detection**: Identifies self-trading and circular trading patterns
- **Volume Spike Detection**: Alerts on unusual volume increases compared to historical averages
- **Price Manipulation Detection**: Monitors for artificial price movements
- **Spoofing/Layering Detection**: Detects fake orders designed to manipulate markets

### Exchange Health Monitoring

Real-time monitoring of exchange performance:

- **Health Score Calculation**: Composite score based on latency, error rates, and uptime
- **Automatic Throttling**: Triggers rate limiting when health scores drop below thresholds
- **Latency Tracking**: Monitors API response times and detects connectivity issues
- **Uptime Monitoring**: Tracks exchange availability and generates alerts on outages

### API Interfaces

- **REST API** (Port 8080): External management interface
- **gRPC API** (Port 50051): High-performance internal communication

## Quick Start

### Prerequisites

- Docker and Docker Compose
- 4GB RAM minimum (8GB recommended)
- 10GB disk space

### Development Setup

1. Clone the repository and navigate to the oversight service:

```bash
cd services/oversight
```

2. Start the infrastructure services:

```bash
docker-compose up -d postgres redis kafka opensearch
```

3. Run database migrations:

```bash
docker-compose exec postgres psql -U csic_admin -d csic_oversight -f /docker-entrypoint-initdb.d/001_init_schema.sql
```

4. Build and run the service:

```bash
go build -o oversight-server ./cmd
./oversight-server
```

### Production Deployment

Deploy with all services:

```bash
docker-compose -f docker-compose.yml up -d
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CSIC_OVERSIGHT_ENVIRONMENT` | Runtime environment | `development` |
| `CSIC_OVERSIGHT_SERVER_HTTP_PORT` | REST API port | `8080` |
| `CSIC_OVERSIGHT_SERVER_GRPC_PORT` | gRPC port | `50051` |
| `CSIC_OVERSIGHT_POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `CSIC_OVERSIGHT_REDIS_ADDR` | Redis address | `localhost:6379` |
| `CSIC_OVERSIGHT_KAFKA_BROKERS` | Kafka brokers | `localhost:9092` |
| `CSIC_OVERSIGHT_OPENSEARCH_ADDRESSES` | OpenSearch URLs | `http://localhost:9200` |

### Configuration File

The service uses `config/config.yaml` for configuration:

```yaml
server:
  http_port: 8080
  grpc_port: 50051

kafka:
  trade_topic: "market.trades.raw"
  alert_topic: "risk.alerts"
  throttle_topic: "control.commands"
```

## API Documentation

### REST Endpoints

#### Alerts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/alerts` | List alerts with filtering |
| GET | `/api/v1/alerts/:id` | Get specific alert |
| PUT | `/api/v1/alerts/:id/status` | Update alert status |
| GET | `/api/v1/alerts/stats` | Get alert statistics |

#### Rules

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/rules` | List detection rules |
| POST | `/api/v1/rules` | Create new rule |
| PUT | `/api/v1/rules/:id` | Update rule |
| DELETE | `/api/v1/rules/:id` | Delete rule |

#### Exchange Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/exchanges` | List exchanges |
| GET | `/api/v1/exchanges/:id/health` | Get exchange health |
| GET | `/api/v1/exchanges/health` | Get all health scores |
| POST | `/api/v1/exchanges/:id/throttle` | Throttle exchange |

### Example Usage

```bash
# Get all active alerts
curl http://localhost:8080/api/v1/alerts?status=OPEN

# Get exchange health
curl http://localhost:8080/api/v1/exchanges/binance/health

# Throttle an exchange
curl -X POST http://localhost:8080/api/v1/exchanges/binance/throttle \
  -H "Content-Type: application/json" \
  -d '{"target_rate_percent": 50, "duration_secs": 300, "reason": "High latency detected"}'
```

## Monitoring

### Health Endpoints

- `GET /health`: Service health status
- `GET /ready`: Readiness probe
- `GET /metrics`: Prometheus metrics

### Metrics

The service exposes the following Prometheus metrics:

- `oversight_trades_processed_total`: Total trades processed
- `oversight_alerts_generated_total`: Total alerts generated by type
- `oversight_exchange_health_score`: Current health scores
- `oversight_processing_latency_seconds`: Trade processing latency

## Project Structure

```
oversight/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go           # Configuration loading
│   │   └── config.yaml         # Configuration file
│   ├── core/
│   │   ├── domain/
│   │   │   └── models.go       # Domain models
│   │   ├── ports/
│   │   │   └── interfaces.go   # Port interfaces
│   │   └── services/
│   │       ├── ingestion_service.go
│   │       ├── abuse_detector.go
│   │       └── health_scorer.go
│   └── adapters/
│       ├── handlers/
│       │   ├── http_server.go
│       │   └── grpc_server.go
│       ├── storage/
│       │   ├── alert_repository.go
│       │   └── repositories.go
│       ├── messaging/
│       │   └── kafka_adapters.go
│       └── analytics/
│           └── opensearch_adapter.go
├── migrations/
│   └── 001_init_schema.sql
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## License

Proprietary - Crypto State Infrastructure Contractor (CSIC)
