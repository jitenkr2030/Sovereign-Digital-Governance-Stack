# Exchange Surveillance Service

A high-performance, government-grade market surveillance system for cryptocurrency exchanges. This service provides real-time market monitoring, pattern detection for market manipulation, and alert generation for regulatory compliance.

## Overview

The Exchange Surveillance Service is part of the Crypto State Infrastructure Contractor (CSIC) platform. It ingests real-time market data from licensed cryptocurrency exchanges, applies sophisticated pattern detection algorithms to identify potential market manipulation, and generates immutable audit records for regulatory oversight.

### Key Features

- **Real-time Market Data Ingestion**: WebSocket-based streaming for high-throughput market data
- **Pattern Detection**: Advanced algorithms for detecting wash trading, spoofing, layering, and price manipulation
- **Alert Management**: Comprehensive alert lifecycle management with severity levels and escalation
- **Time-Series Storage**: TimescaleDB for efficient storage and querying of market data
- **Event Streaming**: Kafka integration for scalable event processing
- **Cryptographic Auditing**: Integration with the Audit Log Service for tamper-evident record keeping

## Architecture

### Clean Architecture

The service follows Clean Architecture principles with clear separation of concerns:

```
internal/
├── domain/          # Business entities and models
│   ├── market_event.go    # Market data events, trades, orders
│   └── alert.go           # Alert models, patterns, thresholds
├── port/            # Interface definitions (ports)
│   └── ports.go             # Repository and service interfaces
├── service/         # Business logic
│   ├── ingestion_service.go     # Real-time data ingestion
│   ├── analysis_service.go      # Pattern detection algorithms
│   └── alert_service.go         # Alert lifecycle management
├── handler/         # Input/output adapters
│   ├── http_handler.go          # REST API handlers
│   └── websocket_handler.go     # WebSocket handlers
└── repository/      # Data access
    ├── postgres_repository.go    # PostgreSQL for alerts
    └── timescale_repository.go   # TimescaleDB for time-series
```

### Data Flow

```
Exchange → WebSocket → Ingestion Service → Kafka Topic
                                              ↓
                                    Analysis Service
                                              ↓
                                    Alert Service → Audit Log
                                              ↓
                                    PostgreSQL/TimescaleDB
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+ with TimescaleDB 2.11+
- Redis 7+
- Apache Kafka 3.x

### Quick Start with Docker

```bash
# Start all dependencies
docker-compose up -d

# Run migrations
go run -tags 'postgres' cmd/migrate/main.go -path migrations

# Start the service
go run main.go
```

### Configuration

The service is configured via `config.yaml`. Key configuration sections:

```yaml
# Server Configuration
server:
  host: "0.0.0.0"
  port: 8080

# Database Configuration
database:
  host: "localhost"
  port: 5432
  username: "csic_surveillance"
  password: "surveillance_password"
  name: "csic_surveillance"

# Kafka Configuration
kafka:
  brokers:
    - "localhost:9092"
  topics:
    market_ingress: "market.ingress"
    alerts: "surveillance.alerts"

# Analysis Configuration
analysis:
  wash_trade:
    enabled: true
    time_window: 60  # seconds
    price_deviation_threshold: 0.01
  spoofing:
    enabled: true
    order_lifetime_window: 300  # seconds
    cancellation_rate_threshold: 0.8
```

## API Reference

### REST Endpoints

#### Alerts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/alerts` | List alerts with filtering |
| GET | `/api/v1/alerts/:id` | Get alert by ID |
| PATCH | `/api/v1/alerts/:id/status` | Update alert status |
| POST | `/api/v1/analysis/wash-trade` | Trigger wash trade analysis |
| POST | `/api/v1/analysis/spoofing` | Trigger spoofing analysis |

#### Market Data

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/markets/:exchange_id/summary` | Get market summary |

#### Configuration

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/config/thresholds` | Get analysis thresholds |
| PUT | `/api/v1/config/thresholds` | Update analysis thresholds |

### WebSocket Endpoints

| Endpoint | Description |
|----------|-------------|
| `/ws/ingest` | Market data ingestion |
| `/ws/alerts` | Real-time alert streaming |

### WebSocket Message Format

#### Ingestion

```json
{
  "exchange_id": "uuid",
  "symbol": "BTC/USD",
  "packet_type": "trade",
  "trades": [{
    "trade_id": "uuid",
    "order_id": "uuid",
    "user_id": "account123",
    "price": "50000.00",
    "quantity": "1.5",
    "direction": "BUY",
    "timestamp": "2024-01-15T10:30:00Z"
  }],
  "timestamp": "2024-01-15T10:30:00Z",
  "sequence_num": 1000
}
```

#### Alert Stream

```json
{
  "id": "uuid",
  "type": "WASH_TRADING",
  "severity": "CRITICAL",
  "status": "OPEN",
  "exchange_id": "uuid",
  "symbol": "BTC/USD",
  "title": "Potential Wash Trading Detected",
  "description": "...",
  "risk_score": 85.5,
  "detected_at": "2024-01-15T10:30:00Z"
}
```

## Pattern Detection

### Wash Trading Detection

Detects when the same accounts buy and sell to themselves or coordinate trades to create false volume.

**Parameters:**
- Time window: 60 seconds (default)
- Price deviation threshold: 1% (default)
- Minimum trades for detection: 3 (default)

**Detection Logic:**
1. Group trades by account
2. Check for self-trading (same account buy/sell)
3. Analyze price and quantity similarity
4. Calculate pattern confidence score

### Spoofing Detection

Detects when large orders are placed to manipulate price then cancelled before execution.

**Parameters:**
- Order lifetime window: 300 seconds (default)
- Large order threshold: 10% of daily volume
- Cancellation rate threshold: 80% (default)

**Detection Logic:**
1. Track orders from creation to cancellation
2. Calculate cancellation rates
3. Detect bait-and-switch patterns
4. Identify layer adding behavior

### Price Manipulation Detection

Detects price deviations from global averages that may indicate manipulation.

**Parameters:**
- Global average window: 3600 seconds (default)
- Deviation threshold: 5% (default)
- Minimum data points: 100 (default)

### Volume Anomaly Detection

Detects unusual volume spikes or drops using statistical analysis.

**Parameters:**
- Anomaly threshold: 3 standard deviations (default)
- Rolling window: 24 hours (default)

## Database Schema

### Main Tables

- `alerts`: Regulatory alerts generated by the system
- `market_events`: Time-series market data (TimescaleDB hypertable)
- `alert_thresholds`: Configurable detection thresholds
- `alert_thresholds_audit`: Audit trail for threshold changes

### Continuous Aggregates

- `candle_1m`: 1-minute OHLCV candles for market summary queries

## Monitoring and Observability

### Metrics

The service exposes Prometheus metrics at `/metrics`:

- `csic_surveillance_alerts_total`: Total alerts generated
- `csic_surveillance_alerts_by_type`: Alerts by type
- `csic_surveillance_alerts_by_severity`: Alerts by severity
- `csic_surveillance_ingestion_rate`: Events per second
- `csic_surveillance_processing_latency`: Processing latency histogram

### Health Checks

- `GET /health`: Service health status
- `GET /ready`: Readiness probe

### Logging

Structured JSON logging with configurable levels. Log fields include:
- `service`: Service name
- `level`: Log level
- `timestamp`: ISO8601 timestamp
- `trace_id`: Request tracing ID
- `span_id`: OpenTelemetry span ID

## Security

### Authentication

- JWT-based authentication for API endpoints
- API key support for exchange data ingestion
- mTLS support for production deployments

### Authorization

- Role-based access control (RBAC)
- Exchange-level data isolation
- Audit logging for all access

### Rate Limiting

- Configurable requests per minute
- Per-exchange rate limits
- Burst allowance

## Development

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out

# Run specific test
go test ./internal/service/... -run TestWashTrading
```

### Code Style

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run ./...

# Vet code
go vet ./...
```

### Adding New Patterns

1. Define new alert type in `internal/domain/alert.go`
2. Implement detection logic in `internal/service/analysis_service.go`
3. Add threshold configuration in `internal/domain/alert.go`
4. Create unit tests
5. Update documentation

## Deployment

### Kubernetes

Helm charts are available in the `deployments/helm` directory:

```bash
# Install the chart
helm install surveillance deployments/helm/csic-surveillance

# Upgrade the chart
helm upgrade surveillance deployments/helm/csic-surveillance
```

### Scaling

- Horizontal scaling via Kubernetes HPA
- Database read replicas for query scaling
- Kafka partition scaling for event throughput

## Compliance

### Audit Trail

All alerts and threshold changes are logged to the Audit Log Service for regulatory compliance.

### Data Retention

- Raw market events: 1 year (configurable)
- Aggregates: 5 years
- Alerts: Indefinite (configurable)

## License

This project is part of the CSIC platform and is proprietary software.

## Support

For support, please contact the CSIC platform team.
