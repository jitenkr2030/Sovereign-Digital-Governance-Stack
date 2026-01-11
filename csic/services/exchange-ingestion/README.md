# Exchange Ingestion Service

The Exchange Ingestion Service is a core component of the CSIC Platform responsible for collecting, normalizing, and distributing market data from various cryptocurrency and traditional exchanges.

## Architecture

This service follows Clean Architecture principles with clear separation between:

- **Domain Layer**: Core business entities and domain logic
- **Ports Layer**: Interface definitions (input/output boundaries)
- **Service Layer**: Business logic and use cases
- **Adapter Layer**: External system integrations (HTTP connectors, Kafka, PostgreSQL)

## Features

### Data Source Management
- Register and configure exchange connections
- Support for multiple exchange types (Binance, Coinbase, Kraken, CME, Generic)
- WebSocket and REST API polling modes
- Connection health monitoring

### Data Ingestion
- Real-time market data streaming
- Trade and order book data collection
- Automatic data normalization
- Configurable polling rates and timeouts

### Event Publishing
- Publishes market data to Kafka topics
- Batch publishing for high throughput
- Guaranteed delivery with acknowledgments

### Monitoring
- Ingestion statistics tracking
- Connection status monitoring
- Latency metrics collection

## Project Structure

```
services/exchange-ingestion/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── core/
│   │   ├── domain/
│   │   │   ├── entity.go          # Domain entities (MarketData, Trade, etc.)
│   │   │   └── errors.go          # Domain error definitions
│   │   ├── ports/
│   │   │   ├── repository.go      # Repository interface definitions
│   │   │   ├── service.go         # Service interface definitions
│   │   │   └── infrastructure.go  # Infrastructure interface definitions
│   │   └── service/
│   │       ├── data_source_service.go  # Data source management
│   │       └── ingestion_service.go    # Data ingestion logic
│   ├── adapter/
│   │   ├── connector/
│   │   │   ├── mock_connector.go  # Mock connector for testing
│   │   │   ├── http_connector.go  # REST API connector
│   │   │   └── factory.go         # Connector factory
│   │   ├── publisher/
│   │   │   └── kafka_publisher.go # Kafka event publisher
│   │   └── repository/
│   │       └── postgres_repository.go # PostgreSQL repository
│   └── handler/
│       └── http_handler.go        # HTTP API handlers
├── migrations/
│   └── 001_create_tables.sql      # Database schema
├── go.mod                         # Go module definition
└── README.md                      # This file
```

## API Endpoints

### Data Source Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/data-sources` | Register a new data source |
| GET | `/api/v1/data-sources` | List all data sources |
| GET | `/api/v1/data-sources/{id}` | Get a specific data source |
| PUT | `/api/v1/data-sources/{id}` | Update a data source |
| DELETE | `/api/v1/data-sources/{id}` | Delete a data source |
| POST | `/api/v1/data-sources/{id}/enable` | Enable a data source |
| POST | `/api/v1/data-sources/{id}/disable` | Disable a data source |
| POST | `/api/v1/data-sources/{id}/test-connection` | Test connection |

### Ingestion Control

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/ingestion/start` | Start all ingestion processes |
| POST | `/api/v1/ingestion/stop` | Stop all ingestion processes |
| GET | `/api/v1/ingestion/status` | Get overall ingestion status |
| POST | `/api/v1/ingestion/source/{id}/start` | Start ingestion for a source |
| POST | `/api/v1/ingestion/source/{id}/stop` | Stop ingestion for a source |
| GET | `/api/v1/ingestion/source/{id}/status` | Get source status |
| POST | `/api/v1/ingestion/source/{id}/sync` | Force sync data |
| GET | `/api/v1/ingestion/source/{id}/stats` | Get source statistics |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_HOST` | HTTP server host | `0.0.0.0` |
| `APP_PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `csic` |
| `DB_PASSWORD` | PostgreSQL password | `csic_secret` |
| `DB_NAME` | PostgreSQL database | `csic_platform` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `KAFKA_TOPIC_PREFIX` | Kafka topic prefix | `csic` |
| `DEFAULT_POLLING_RATE` | Default polling interval | `5s` |
| `DEFAULT_TIMEOUT` | Default API timeout | `30s` |
| `LOG_LEVEL` | Logging level | `info` |

## Supported Exchanges

| Exchange | REST API | WebSocket |
|----------|----------|-----------|
| Binance | Supported | Not implemented |
| Coinbase | Supported | Not implemented |
| Kraken | Supported | Not implemented |
| CME | Supported | Not implemented |
| Generic | Supported | Not implemented |
| Mock | Supported | Supported |

## Installation

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Apache Kafka 3.0+

### Building

```bash
cd services/exchange-ingestion
go build -o bin/exchange-ingestion ./cmd
```

### Running

```bash
# With environment variables
export DB_HOST=localhost
export KAFKA_BROKERS=localhost:9092
./bin/exchange-ingestion

# Or using docker-compose
docker-compose up -d
```

### Database Migration

```bash
psql -h localhost -U csic -d csic_platform -f migrations/001_create_tables.sql
```

## Usage Examples

### Register a Mock Data Source

```bash
curl -X POST http://localhost:8080/api/v1/data-sources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-mock-exchange",
    "exchange_type": "MOCK",
    "endpoint": "http://localhost:8081",
    "symbols": ["BTC-USD", "ETH-USD"],
    "polling_rate": "1s",
    "ws_enabled": false,
    "auth_type": "NONE"
  }'
```

### Start Ingestion

```bash
curl -X POST http://localhost:8080/api/v1/ingestion/start
```

### Check Status

```bash
curl http://localhost:8080/api/v1/ingestion/status
```

## Kafka Topics

The service publishes to the following Kafka topics:

- `{topic_prefix}_market_data`: Normalized market data
- `{topic_prefix}_trades`: Trade executions
- `{topic_prefix}_orderbook`: Order book snapshots

## Testing

```bash
# Unit tests
go test ./... -v

# Integration tests
go test ./... -tags=integration -v
```

## License

This project is part of the CSIC Platform and is licensed under the MIT License.
