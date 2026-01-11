# CSIC Platform - API Gateway Service
# Clean Architecture Implementation following Domain-Driven Design principles

## Architecture Overview

The API Gateway service follows Clean Architecture with the following layers:

### Directory Structure

```
cmd/
  └── main.go                    # Application entry point

internal/
  ├── config/
  │   └── config.go              # Configuration loading and management
  
  ├── core/
  │   ├── domain/
  │   │   ├── entity.go          # Domain entities (Exchange, Wallet, Miner, Alert)
  │   │   └── value_objects.go   # Value objects and types
  │   ├── ports/
  │   │   ├── repository.go      # Repository interfaces
  │   │   └── service.go         # Service interfaces
  │   └── service/
  │       └── gateway_service.go # Business logic and use cases
  │
  ├── adapter/
  │   ├── repository/
  │   │   └── postgres_repository.go  # PostgreSQL implementations
  │   └── messaging/
  │       └── kafka_producer.go       # Kafka producer implementations
  │
  └── handler/
      └── http_handler.go        # HTTP handlers (Controllers)

config.yaml                       # Service configuration
go.mod                           # Go module definition
Dockerfile                       # Docker containerization
docker-compose.yml               # Local development setup
README.md                        # Documentation
```

### Layer Responsibilities

1. **Domain Layer** (`internal/core/domain/`)
   - Contains business entities (Exchange, Wallet, Miner, Alert, User)
   - Defines value objects and domain exceptions
   - Pure Go code with no external dependencies

2. **Ports Layer** (`internal/core/ports/`)
   - Repository interfaces (data access contracts)
   - Service interfaces (business logic contracts)
   - Defines contracts that adapters must implement

3. **Service Layer** (`internal/core/service/`)
   - Contains business logic and use cases
   - Orchestrates domain entities to fulfill business requirements
   - Depends on ports (interfaces), not concrete implementations

4. **Adapter Layer** (`internal/adapter/`)
   - Repository implementations (PostgreSQL, Redis, etc.)
   - Message producers (Kafka, etc.)
   - External service clients
   - Implements ports defined in the core layer

5. **Handler Layer** (`internal/handler/`)
   - HTTP handlers (Controllers/Controllers)
   - Receives HTTP requests
   - Calls service layer
   - Returns HTTP responses

6. **Config Layer** (`internal/config/`)
   - Configuration loading from YAML files
   - Environment variable overrides
   - Configuration validation

### Key Features

- **Request Routing**: Intelligent routing to backend services
- **Authentication & Authorization**: JWT-based auth with role-based access control
- **Rate Limiting**: Protection against DDoS and abuse
- **Request/Response Transformation**: Data format conversion
- **Circuit Breaker**: Resilience against downstream failures
- **Load Balancing**: Distribution across service instances
- **Caching**: Redis-based caching for performance
- **Metrics & Monitoring**: Prometheus metrics and health checks
- **Distributed Tracing**: OpenTelemetry integration

### Configuration

All configuration is managed through `config.yaml`. Key sections:

- **app**: Application metadata
- **server**: HTTP server settings
- **database**: PostgreSQL connection settings
- **redis**: Redis connection settings
- **kafka**: Kafka broker settings
- **security**: JWT, password, and session configuration
- **logging**: Logging preferences
- **monitoring**: Metrics and health check settings

### Running the Service

```bash
# Local development
go run cmd/main.go -config=config.yaml

# Docker
docker build -t csic-api-gateway .
docker run -p 8080:8080 csic-api-gateway

# Docker Compose
docker-compose up -d
```

### API Endpoints

- `GET /health` - Health check endpoint
- `GET /api/v1/dashboard/stats` - Dashboard statistics
- `GET /api/v1/alerts` - List alerts
- `GET /api/v1/alerts/:id` - Get alert by ID
- `POST /api/v1/alerts/:id/acknowledge` - Acknowledge alert
- `GET /api/v1/exchanges` - List exchanges
- `GET /api/v1/exchanges/:id` - Get exchange by ID
- `POST /api/v1/exchanges/:id/suspend` - Suspend exchange
- `GET /api/v1/wallets` - List wallets
- `POST /api/v1/wallets/:id/freeze` - Freeze wallet
- `GET /api/v1/miners` - List miners
- `GET /api/v1/compliance/reports` - List compliance reports
- `POST /api/v1/compliance/reports` - Generate compliance report
- `GET /api/v1/audit/logs` - List audit logs
- `GET /api/v1/blockchain/status` - Blockchain node status
- `GET /api/v1/users/me` - Get current user
- `GET /api/v1/users` - List users

### Dependencies

- **Gin**: HTTP web framework
- **Viper**: Configuration management
- **Zap**: Structured logging
- **Redis**: Caching and rate limiting
- **PostgreSQL**: Persistent storage
- **Kafka**: Event streaming
- **JWT**: Token-based authentication

### Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### License

This project is part of the CSIC Platform and is proprietary software.
