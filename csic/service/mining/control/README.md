# Mining Control Module

A comprehensive microservices module for managing cryptocurrency mining operations with a focus on energy consumption monitoring, regulatory compliance, and operational efficiency. Built using Go and Clean Architecture principles.

## Overview

The Mining Control module provides a robust platform for monitoring and controlling mining operations, ensuring compliance with energy regulations, detecting violations, and optimizing power consumption. The system integrates with PostgreSQL (with TimescaleDB for time-series data), Redis for caching, and supports event-driven architecture via Kafka.

### Key Features

The Mining Control module delivers comprehensive capabilities for mining operation management. Pool and Machine Management enables registration, monitoring, and control of mining pools and individual machines with real-time status tracking. Energy Consumption Monitoring provides precise tracking of power usage at pool and machine levels with time-series storage and anomaly detection. Compliance Management automates regulatory compliance checks, tracks violations, and generates compliance reports. Alert System delivers real-time notifications for energy threshold breaches and policy violations. Analytics and Reporting offers historical trend analysis, efficiency metrics, and customizable reports through integrated Grafana dashboards.

## Architecture

### Technology Stack

The Mining Control module leverages modern technologies to deliver high performance and reliability. Go serves as the primary programming language, chosen for its excellent concurrency support and performance characteristics. PostgreSQL with TimescaleDB handles relational data storage and time-series energy consumption data. Redis provides high-performance caching for frequently accessed data. Kafka enables event-driven architecture for real-time data processing. Prometheus collects and stores metrics, while Grafana provides visualization and dashboards. Docker containers ensure consistent deployment across environments.

### Project Structure

```
csic-platform/mining/control/
├── main.go                      # Application entry point
├── config.yaml                  # Configuration file
├── Dockerfile                   # Docker image definition
├── docker-compose.yml           # Local development environment
├── prometheus.yml               # Prometheus monitoring configuration
├── migrations/                  # Database migration files
│   ├── V1__create_initial_schema.sql
│   └── V2__create_timescale_hypertables.sql
├── internal/
│   ├── domain/                  # Business entities and value objects
│   │   └── models.go
│   ├── service/                 # Business logic and use cases
│   │   ├── registration_service.go
│   │   ├── monitoring_service.go
│   │   ├── enforcement_service.go
│   │   └── reporting_service.go
│   ├── handler/                 # API handlers (HTTP, gRPC)
│   │   └── http_handler.go
│   └── repository/              # Data access layer
│       ├── pool_repository.go
│       ├── machine_repository.go
│       ├── energy_repository.go
│       └── violation_repository.go
├── grafana/                     # Dashboard configurations
└── README.md                    # This file
```

## Getting Started

### Prerequisites

Before you begin, ensure you have the following prerequisites installed and configured on your system. Docker and Docker Compose are required for containerized deployment. Go 1.21 or higher is needed if you plan to build from source. PostgreSQL 16 with TimescaleDB 2.13 is required for the database. Make sure your system has at least 4GB of RAM available for running the full stack.

### Environment Setup

Clone the repository and navigate to the Mining Control module directory. Create a `.env` file based on the example configuration to customize your environment settings. The application uses sensible defaults, so basic operation is possible without extensive configuration.

```bash
cd csic-platform/mining/control
cp config.example.yaml config.yaml
```

### Running with Docker Compose

The easiest way to run the Mining Control module is using Docker Compose, which starts all required services including the database, cache, and monitoring stack.

```bash
# Start all services in detached mode
docker-compose up -d

# View logs
docker-compose logs -f mining-control

# Stop all services
docker-compose down
```

After starting, the following services will be available: the Mining Control API at http://localhost:8080, Prometheus metrics at http://localhost:9090, Grafana dashboards at http://localhost:3000, PostgreSQL database at localhost:5432, and Redis cache at localhost:6379.

### Running Locally

For development purposes, you can run the application directly on your local machine after setting up the database.

```bash
# Start PostgreSQL with TimescaleDB (using Docker for the database only)
docker run -d --name mining-db \
  -e POSTGRES_USER=mining_control \
  -e POSTGRES_PASSWORD=mining_control_secret \
  -e POSTGRES_DB=mining_control \
  -p 5432:5432 \
  timescale/timescaledb:2.13.1-pg16

# Apply database migrations
export DB_PASSWORD=mining_control_secret
cat migrations/V1__create_initial_schema.sql | docker exec -i mining-db psql -U mining_control -d mining_control
cat migrations/V2__create_timescale_hypertables.sql | docker exec -i mining-db psql -U mining_control -d mining_control

# Build and run the application
go build -o mining-control main.go
./mining-control
```

## Configuration

The application is configured using a YAML configuration file. All settings can be overridden using environment variables with the appropriate prefix.

### Configuration File Structure

```yaml
# Application Configuration
app:
  host: "0.0.0.0"
  port: 8080
  environment: "development"
  log_level: "debug"

# Database Configuration
database:
  host: "localhost"
  port: 5432
  username: "mining_control"
  password: "mining_control_secret"
  name: "mining_control"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"

# Redis Configuration
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# Alert Configuration
alerts:
  check_interval: "30s"
  energy_threshold_warning: 80
  energy_threshold_critical: 95
```

### Environment Variables

All configuration options can be set via environment variables. The following variables are commonly used in production deployments. APP_ENV sets the environment (development, staging, production). DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, and DB_NAME configure the database connection. REDIS_HOST and REDIS_PORT configure the Redis connection. METRICS_ENABLED controls whether metrics are exposed. ENERGY_THRESHOLD_WARNING and ENERGY_THRESHOLD_CRITICAL set alert thresholds.

## API Documentation

The Mining Control module exposes a RESTful API for integration with other systems and front-end applications.

### Health Check

To verify the service is running correctly, send a GET request to the health endpoint. This endpoint returns the service status and checks the database connection.

```bash
curl http://localhost:8080/health
```

### Pool Management

Pool management endpoints allow you to create, read, update, and delete mining pool records. Creating a new pool requires providing the pool name, operator ID, location, and energy allocation limits. The API validates all inputs and returns appropriate error messages for invalid requests.

### Energy Monitoring

Energy monitoring endpoints provide access to real-time and historical energy consumption data. You can query current power usage, hourly and daily statistics, and detect anomalies in consumption patterns. The time-series data is optimized for efficient querying using TimescaleDB.

### Compliance Endpoints

Compliance endpoints manage violation tracking and compliance reporting. Violations can be created manually or detected automatically by the monitoring system. The API supports filtering violations by type, severity, status, and date range.

## Database Schema

### Core Tables

The database schema consists of several interconnected tables that store all operational data. Mining Pools stores pool-level information including energy allocations and compliance scores. Mining Machines tracks individual machines within pools with their hashrate and power consumption. Energy Consumption Logs maintains time-series data for power consumption with high granularity. Compliance Violations records all detected violations with severity levels and resolution status. Compliance Records stores periodic inspection results and compliance scores.

### Time-Series Configuration

Energy consumption data is stored in TimescaleDB hypertables for optimal performance. Data is automatically compressed after 7 days and retained for 2 years by default. Continuous aggregates provide pre-computed hourly and daily statistics for efficient querying.

## Monitoring and Observability

### Metrics

The application exposes metrics in Prometheus format at the /metrics endpoint. Key metrics include HTTP request latency and throughput, pool and machine counts, energy consumption statistics, violation counts by type and severity, and database query performance.

### Dashboards

Grafana dashboards are pre-configured for monitoring key metrics. The main dashboard displays real-time energy consumption across all pools, while compliance dashboards show violation trends and compliance scores.

### Alerting

Alerts are triggered when energy consumption exceeds configured thresholds. The warning threshold triggers at 80% of the energy limit, while the critical threshold triggers at 95%. Alerts can be integrated with external notification systems via webhooks.

## Development

### Running Tests

Execute the test suite to verify the application functionality. The test suite includes unit tests for business logic, integration tests for API endpoints, and repository tests for data access.

```bash
# Run all tests
go test ./... -v

# Run tests with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Adding New Features

Follow the Clean Architecture principles when adding new features. Define domain entities in the internal/domain package, implement business logic in internal/service, add API handlers in internal/handler, and create repository implementations in internal/repository. Database migrations should be added to the migrations directory with sequential version numbers.

## Security Considerations

For production deployments, implement the following security measures. Use strong, unique passwords for database and Redis connections. Enable TLS for all network communication, including database connections. Implement proper authentication and authorization for API endpoints. Regularly update dependencies to patch security vulnerabilities. Use secrets management for sensitive configuration values.

## License

This module is part of the CSIC Platform and is subject to the platform's license agreement.

## Support

For questions, issues, or feature requests, please open an issue in the project repository or contact the CSIC Platform team.
