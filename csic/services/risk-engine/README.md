# Risk Engine Service

The Risk Engine Service is a core component of the CSIC Platform responsible for evaluating market data against risk rules, generating alerts, and performing comprehensive risk assessments.

## Architecture

This service follows Clean Architecture principles with clear separation between:

- **Domain Layer**: Core business entities (RiskRule, RiskAlert, RiskProfile, Exposure)
- **Ports Layer**: Interface definitions for repositories, services, and infrastructure
- **Service Layer**: Business logic for rule evaluation and risk assessment
- **Adapter Layer**: External system integrations (Kafka, PostgreSQL)

## Features

### Risk Rule Management
- Create, update, delete risk rules
- Support for multiple operators (GT, LT, EQ, GTE, LTE, NEQ)
- Configurable severity levels (LOW, MEDIUM, HIGH, CRITICAL)
- Symbol and source filtering
- Enable/disable rules

### Real-time Risk Evaluation
- Consumes market data from Kafka
- Evaluates data against enabled rules
- Generates alerts when thresholds are breached
- Publishes alerts to Kafka for downstream processing

### Alert Management
- View all alerts with filtering
- Acknowledge, resolve, or dismiss alerts
- Alert statistics and metrics
- Severity-based categorization

### Risk Assessment
- Comprehensive entity risk assessment
- Exposure calculation and tracking
- Risk score calculation
- Overall risk level determination

### Integration
- Kafka consumer for market data streams
- Kafka publisher for alerts
- PostgreSQL for persistent storage
- HTTP API for external access

## Project Structure

```
services/risk-engine/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── entity.go          # Domain entities (RiskRule, RiskAlert, etc.)
│   │   │   └── errors.go          # Domain error definitions
│   │   ├── ports/
│   │   │   ├── repository.go      # Repository interface definitions
│   │   │   ├── service.go         # Service interface definitions
│   │   │   └── infrastructure.go  # Infrastructure interface definitions
│   │   └── service/
│   │       ├── risk_evaluation_service.go  # Core risk evaluation logic
│   │       └── risk_rule_service.go        # Rule management
│   ├── adapter/
│   │   ├── consumer/
│   │   │   └── kafka_consumer.go  # Kafka market data consumer
│   │   └── publisher/
│   │       └── alert_publisher.go # Kafka alert publisher
│   └── handler/
│       └── http_handler.go        # HTTP API handlers
├── migrations/
│   └── 001_create_tables.sql      # Database schema
├── go.mod                         # Go module definition
└── README.md                      # This file
```

## API Endpoints

### Risk Rules

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/rules` | Create a new risk rule |
| GET | `/api/v1/rules` | List all risk rules |
| GET | `/api/v1/rules/{id}` | Get a specific rule |
| PUT | `/api/v1/rules/{id}` | Update a rule |
| DELETE | `/api/v1/rules/{id}` | Delete a rule |
| POST | `/api/v1/rules/{id}/enable` | Enable a rule |
| POST | `/api/v1/rules/{id}/disable` | Disable a rule |

### Alerts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/alerts` | List all alerts |
| GET | `/api/v1/alerts/{id}` | Get a specific alert |
| GET | `/api/v1/alerts/active` | Get all active alerts |
| POST | `/api/v1/alerts/{id}/acknowledge` | Acknowledge an alert |
| POST | `/api/v1/alerts/{id}/resolve` | Resolve an alert |
| POST | `/api/v1/alerts/{id}/dismiss` | Dismiss an alert |
| GET | `/api/v1/alerts/stats` | Get alert statistics |

### Risk Assessments

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/assessments` | Perform risk assessment |
| GET | `/api/v1/assessments/{entity_id}` | Get entity assessment |

### Thresholds

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/thresholds` | Get threshold configuration |
| PUT | `/api/v1/thresholds` | Update threshold configuration |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |

## Supported Rule Metrics

| Metric | Description |
|--------|-------------|
| `price` | Current price of the asset |
| `volume` | Trading volume |
| `change_24h` | 24-hour price change percentage |
| `volatility` | Price volatility (high-low spread) |

## Supported Rule Operators

| Operator | Description |
|----------|-------------|
| `GT` | Greater than |
| `LT` | Less than |
| `EQ` | Equal to |
| `GTE` | Greater than or equal |
| `LTE` | Less than or equal |
| `NEQ` | Not equal to |

## Alert Severity Levels

| Level | Score Range | Color |
|-------|------------|-------|
| `LOW` | 0-39 | Green |
| `MEDIUM` | 40-59 | Yellow |
| `HIGH` | 60-79 | Orange |
| `CRITICAL` | 80+ | Red |

## Kafka Topics

### Input Topics
- `{topic_prefix}_market_data`: Market data from ingestion service

### Output Topics
- `{topic_prefix}_risk_alerts`: Generated risk alerts

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_HOST` | HTTP server host | `0.0.0.0` |
| `APP_PORT` | HTTP server port | `8081` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `csic` |
| `DB_PASSWORD` | PostgreSQL password | `csic_secret` |
| `DB_NAME` | PostgreSQL database | `csic_platform` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `TOPIC_PREFIX` | Kafka topic prefix | `csic` |
| `LOG_LEVEL` | Logging level | `info` |

## Installation

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Apache Kafka 3.0+

### Building

```bash
cd services/risk-engine
go build -o bin/risk-engine ./cmd
```

### Running

```bash
./bin/risk-engine
```

### Database Migration

```bash
psql -h localhost -U csic -d csic_platform -f migrations/001_create_tables.sql
```

## Usage Examples

### Create a Risk Rule

```bash
curl -X POST http://localhost:8081/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Price Alert",
    "description": "Alert when price exceeds threshold",
    "metric": "price",
    "operator": "GT",
    "threshold": "50000",
    "severity": "HIGH",
    "symbol_filter": "BTC-USD"
  }'
```

### Get Active Alerts

```bash
curl http://localhost:8081/api/v1/alerts/active
```

### Perform Risk Assessment

```bash
curl -X POST http://localhost:8081/api/v1/assessments \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "user-123",
    "entity_type": "trader"
  }'
```

## Example Risk Rule Configuration

```json
{
  "name": "Volume Spike Alert",
  "description": "Alert when trading volume spikes 200%",
  "metric": "volume",
  "operator": "GT",
  "threshold": "1000000",
  "severity": "MEDIUM",
  "symbol_filter": "BTC-USD",
  "enabled": true
}
```

## Testing

```bash
# Unit tests
go test ./... -v

# Integration tests
go test ./... -tags=integration -v
```

## Monitoring

The service exposes the following metrics:

- `risk_alerts_total`: Total number of alerts by severity
- `risk_assessment_score`: Risk assessment scores
- `risk_evaluation_latency`: Risk evaluation processing time

## License

This project is part of the CSIC Platform and is licensed under the MIT License.
