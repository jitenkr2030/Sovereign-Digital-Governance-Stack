# Access Control Service

A comprehensive Attribute-Based Access Control (ABAC) service for the CSIC Platform. This service provides fine-grained access control through policy-based authorization, supporting complex access control scenarios with temporal, environmental, and contextual conditions.

## Features

- **Attribute-Based Access Control (ABAC)**: Fine-grained authorization based on subject, resource, action, and environmental attributes
- **Policy Management**: Create, update, delete, and manage access control policies with priority-based evaluation
- **Role-Based Access Control (RBAC)**: Support for role-based assignments with scope-based permissions
- **Real-time Policy Updates**: Kafka-based event streaming for dynamic policy updates
- **Audit Logging**: Comprehensive audit trail of all access decisions
- **Caching**: Redis-based policy caching for high-performance access checks
- **RESTful API**: Clean HTTP API for integration with other services

## Architecture

The service follows the Clean Architecture pattern with clear separation of concerns:

```
cmd/
  └── main.go                 # Application entrypoint
internal/
  ├── adapter/
  │   ├── consumer/
  │   │   └── kafka_consumer.go       # Kafka event consumer
  │   └── repository/
  │       └── postgres_repository.go  # PostgreSQL data access
  ├── config/
  │   └── config.go           # Configuration management
  ├── core/
  │   ├── domain/
  │   │   └── entity.go       # Domain models
  │   ├── ports/
  │   │   ├── repository.go   # Repository interfaces
  │   │   └── service.go      # Service interfaces
  │   └── service/
  │       └── access_control_service.go  # Business logic
  └── handler/
      └── http_handler.go     # HTTP request handlers
```

## Prerequisites

- Go 1.21 or later
- PostgreSQL 15+
- Redis 7+
- Apache Kafka 3.0+

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f access-control

# Stop services
docker-compose down
```

### Manual Setup

1. **Initialize the database**:

```bash
# Create database schema
psql -U postgres -d csic_platform -f migrations/001_create_tables.sql
```

2. **Configure the service**:

```bash
# Copy example configuration
cp config.example.yaml config.yaml

# Edit configuration
vim config.yaml
```

3. **Build and run**:

```bash
# Build the service
go build -o access-control-service ./cmd/main.go

# Run the service
./access-control-service
```

## Configuration

The service is configured through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Application name | access-control-service |
| `HOST` | Server host | 0.0.0.0 |
| `PORT` | Server port | 8080 |
| `ENV` | Environment | development |
| `POSTGRES_HOST` | PostgreSQL host | localhost |
| `POSTGRES_PORT` | PostgreSQL port | 5432 |
| `POSTGRES_USER` | PostgreSQL user | postgres |
| `POSTGRES_PASSWORD` | PostgreSQL password | postgres |
| `POSTGRES_DB` | PostgreSQL database | csic_platform |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `KAFKA_BROKERS` | Kafka brokers | localhost:9092 |
| `KAFKA_TOPIC` | Kafka topic | access-control-events |
| `LOG_LEVEL` | Logging level | info |
| `POLICY_CACHE_TTL` | Cache TTL (seconds) | 300 |

## API Endpoints

### Access Check

```bash
# Single access check
POST /api/v1/check
Content-Type: application/json

{
  "subject": {
    "id": "user-123",
    "type": "user",
    "roles": ["admin"],
    "groups": ["engineering"],
    "attributes": {"department": "tech"}
  },
  "resource": {
    "type": "document",
    "id": "doc-456",
    "path": "/documents/reports",
    "attributes": {"classification": "confidential"}
  },
  "action": {
    "name": "read",
    "method": "GET"
  },
  "context": {
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0",
    "time": "2024-01-15T10:30:00Z"
  }
}
```

### Policy Management

```bash
# List policies
GET /api/v1/policies?enabled_only=false&page=1&page_size=20

# Create policy
POST /api/v1/policies
Content-Type: application/json

{
  "name": "Admin Full Access",
  "description": "Administrators have full access to all resources",
  "effect": "ALLOW",
  "priority": 100,
  "conditions": {
    "subjects": [{"roles": ["admin"]}],
    "resources": [{"types": ["*"]}],
    "actions": [{"operations": ["*"]}]
  }
}

# Get policy
GET /api/v1/policies/{id}

# Update policy
PUT /api/v1/policies/{id}

# Delete policy
DELETE /api/v1/policies/{id}

# Enable/Disable policy
POST /api/v1/policies/{id}/enable
POST /api/v1/policies/{id}/disable
```

### Health Check

```bash
GET /health
```

## Policy Structure

Policies are defined with the following structure:

```yaml
Policy:
  id: uuid
  name: string
  description: string
  effect: ALLOW | DENY
  priority: int (higher = evaluated first)
  enabled: boolean
  conditions:
    subjects:
      - roles: [string]
        users: [string]
        groups: [string]
        attributes: map[string]string
    resources:
      - types: [string]
        ids: [string]
        patterns: [string]
        attributes: map[string]string
    actions:
      - operations: [string]
        methods: [string]
    environment:
      - ip_ranges: [string]
        time_ranges:
          - start_time: string
            end_time: string
            days_of_week: [int]
        devices: [string]
        locations: [string]
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/core/service/...
```

### Code Generation

```bash
# Generate mocks
go generate ./...

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq create_tables

# Run migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/csic_platform?sslmode=disable" up

# Rollback migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/csic_platform?sslmode=disable" down
```

## Performance Considerations

- **Policy Caching**: Policies are cached in Redis with configurable TTL
- **Connection Pooling**: PostgreSQL connection pool is configured with optimal settings
- **Kafka Consumer**: Batch message processing for efficient event handling
- **HTTP Server**: Configurable timeouts and concurrency limits

## Monitoring

### Metrics

The service exposes Prometheus metrics at `/metrics`:

- `access_check_total`: Total number of access checks
- `access_check_allow_total`: Total number of allow decisions
- `access_check_deny_total`: Total number of deny decisions
- `policy_cache_hits_total`: Policy cache hit count
- `policy_cache_misses_total`: Policy cache miss count

### Health Endpoints

- `/health`: Basic health check
- `/ready`: Readiness check (includes dependency checks)

## Security

- **Authentication**: JWT-based authentication for API access
- **Authorization**: Fine-grained ABAC authorization
- **Audit Logging**: All access decisions are logged
- **Data Protection**: Encrypted connections to all dependencies

## License

This project is part of the CSIC Platform and is licensed under the MIT License.
