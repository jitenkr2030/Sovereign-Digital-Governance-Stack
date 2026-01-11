# CSIC Reporting Service

## Overview

The **CSIC Reporting Service** is a Go-based microservice that provides comprehensive regulatory reporting capabilities for the Crypto State Infrastructure Contractor (CSIC) Platform. This service handles automated report generation, custom regulator views, and secure exports.

## Features

### Automated Report Generation
- **Template-based Reports**: Define reusable report templates with configurable schemas, data sources, and output formats
- **Scheduled Reports**: Automated report generation based on configurable cron schedules (daily, weekly, monthly, quarterly, annually)
- **Bulk Generation**: Generate multiple reports concurrently for efficient processing
- **Report History**: Complete audit trail of all report generations, modifications, and submissions

### Custom Regulator Views
- **Regulator Configuration**: Configure multiple regulatory bodies with their specific requirements
- **Custom Dashboards**: Create regulator-specific dashboard views with customizable widgets
- **View Preferences**: Save and apply view preferences per regulator (filters, columns, sorting)
- **Multi-language Support**: Generate reports in multiple languages for different regulators

### Secure Exports
- **Encrypted Exports**: AES-256 encryption for all exported reports
- **Audit Trail**: Complete logging of all export activities including user, IP, and timestamp
- **Access Control**: Configurable download limits, expiration times, and access restrictions
- **Watermarking**: Optional watermarking for exported documents
- **Multiple Formats**: Export reports in PDF, CSV, XLSX, JSON, XBRL, and XML formats

## Architecture

### Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Framework | Gin |
| Database | PostgreSQL 15 |
| Cache | Redis 7 |
| Message Queue | Apache Kafka |
| Containerization | Docker |
| Monitoring | Prometheus + Grafana |

### Architecture Pattern

The service follows **Hexagonal Architecture** (Ports & Adapters):

```
cmd/api/main.go              # Application entry point
internal/
├── domain/                   # Business logic and entities
│   └── models/              # Domain models (ReportTemplate, GeneratedReport, etc.)
├── service/                  # Business services
│   ├── reporting_service.go  # Report generation logic
│   ├── export_service.go     # Secure export handling
│   └── scheduler_service.go  # Automated scheduling
├── repository/               # Data access layer
│   ├── postgres_repository.go # PostgreSQL implementations
│   └── cache_repository.go   # Redis caching
├── handler/                  # API and event handlers
│   └── http/handler.go       # HTTP API handlers
│   └── kafka/kafka.go        # Kafka producer/consumer
└── config/                   # Configuration
```

### API Structure

```
/api/v1/
├── reports/                   # Report management
│   ├── GET    /reports                    # List reports
│   ├── POST   /reports                    # Generate new report
│   ├── GET    /reports/{id}               # Get report details
│   ├── POST   /reports/{id}/regenerate    # Regenerate report
│   ├── POST   /reports/{id}/approve       # Approve report
│   ├── POST   /reports/{id}/validate      # Validate report
│   ├── POST   /reports/{id}/archive       # Archive report
│   ├── POST   /reports/{id}/submit        # Submit to regulator
│   └── GET    /reports/{id}/download      # Download report
│
├── templates/                # Report templates
│   ├── GET    /templates                 # List templates
│   ├── POST   /templates                 # Create template
│   ├── GET    /templates/{id}            # Get template
│   ├── PUT    /templates/{id}            # Update template
│   └── DELETE /templates/{id}            # Delete template
│
├── schedules/               # Report schedules
│   ├── GET    /schedules                 # List schedules
│   ├── POST   /schedules                 # Create schedule
│   ├── GET    /schedules/{id}            # Get schedule
│   ├── PUT    /schedules/{id}            # Update schedule
│   ├── DELETE /schedules/{id}            # Delete schedule
│   ├── POST   /schedules/{id}/trigger    # Trigger immediately
│   ├── POST   /schedules/{id}/activate   # Activate schedule
│   └── POST   /schedules/{id}/deactivate # Deactivate schedule
│
├── exports/                 # Export management
│   ├── GET    /exports                    # List exports
│   ├── GET    /exports/{id}               # Get export details
│   └── POST   /exports/{id}/download      # Download export
│
└── stats/                   # Statistics
    ├── GET    /stats/reports              # Report statistics
    ├── GET    /stats/exports              # Export statistics
    └── GET    /stats/scheduler            # Scheduler statistics
```

## Getting Started

### Prerequisites

- Go 1.21 or later
- PostgreSQL 15
- Redis 7
- Apache Kafka 7.5
- Docker and Docker Compose

### Local Development

1. **Clone the repository**
   ```bash
   cd /workspace/csic-platform/service/reporting
   ```

2. **Start infrastructure services**
   ```bash
   docker-compose up -d
   ```

3. **Run database migrations**
   ```bash
   # Connect to the database and run migrations from:
   # internal/db/migrations/001_init_schema.sql
   ```

4. **Configure the application**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your settings
   ```

5. **Build and run**
   ```bash
   go build -o csic-reporting-service ./cmd/api
   ./csic-reporting-service
   ```

### Using Docker

```bash
# Build the image
docker build -t csic-reporting-service .

# Run the service
docker run -d \
  --name csic-reporting \
  -p 8080:8080 \
  -p 9090:9090 \
  --network csic-network \
  csic-reporting-service
```

### Docker Compose (Full Stack)

```bash
# Start all services including monitoring
docker-compose --profile monitoring up -d

# Start only core services
docker-compose up -d
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `reporting-db` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `csic_admin` |
| `DB_PASSWORD` | Database password | `secure_password` |
| `DB_NAME` | Database name | `csic_reporting` |
| `REDIS_HOST` | Redis host | `reporting-cache` |
| `REDIS_PORT` | Redis port | `6379` |
| `KAFKA_BROKERS` | Kafka brokers | `kafka:9092` |
| `APP_PORT` | Application port | `8080` |
| `APP_MODE` | Gin mode | `release` |

### Configuration File

The `config.yaml` file contains all application settings:

```yaml
app:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  host: "localhost"
  port: 5432
  username: "csic_admin"
  password: "secure_password"
  name: "csic_reporting"

cache:
  host: "localhost"
  port: 6379
  ttl: 3600

kafka:
  enabled: true
  brokers:
    - "localhost:9092"
```

## API Documentation

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "service": "csic-reporting-service"
}
```

### Generate Report

```bash
curl -X POST http://localhost:8080/api/v1/reports \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "template-uuid",
    "period_start": "2024-01-01T00:00:00Z",
    "period_end": "2024-03-31T23:59:59Z",
    "output_formats": ["pdf", "csv"]
  }'
```

### Download Report

```bash
curl -O -J http://localhost:8080/api/v1/reports/{report_id}/download?format=pdf
```

### Create Schedule

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "template-uuid",
    "name": "Weekly Compliance Report",
    "frequency": "weekly",
    "cron_expression": "0 6 * * 1",
    "output_formats": ["pdf"]
  }'
```

## Monitoring

### Metrics Endpoint

Prometheus metrics are available at `http://localhost:8080/metrics`

Key metrics:
- `report_generation_total` - Total reports generated
- `report_generation_duration_seconds` - Report generation duration
- `report_generation_failed_total` - Failed report generations
- `export_downloads_total` - Total exports downloaded
- `export_failures_total` - Failed exports
- `scheduler_running` - Scheduler status (0/1)

### Health Checks

| Endpoint | Description |
|----------|-------------|
| `/health` | Overall service health |
| `/ready` | Readiness probe |

### Grafana Dashboards

Pre-configured dashboards available:
- **Service Overview**: General service health and metrics
- **Report Generation**: Report generation statistics
- **Export Monitoring**: Export activity and failures

## Database Schema

### Core Tables

| Table | Description |
|-------|-------------|
| `regulators` | Regulatory authority registry |
| `report_templates` | Report template definitions |
| `generated_reports` | Generated report instances |
| `report_schedules` | Automated schedule definitions |
| `report_queue` | Report generation queue |
| `report_history` | Report change history |
| `export_logs` | Export audit trail |
| `regulator_users` | Regulator user accounts |
| `regulator_api_keys` | API keys for regulator access |

## Security

### Authentication

The service uses JWT-based authentication:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/reports
```

### Rate Limiting

Default rate limits:
- 1000 requests per minute
- Burst size: 100

### Export Security

- All exports are logged with user, IP, and timestamp
- Configurable download limits per export
- Optional encryption for exported files
- Watermarking for sensitive documents
- Automatic expiration of download links

## Testing

### Unit Tests

```bash
go test ./... -v
```

### Integration Tests

```bash
go test ./... -tags=integration -v
```

### Test Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Deployment

### Kubernetes

Example deployment manifest available in `deploy/kubernetes/`

```bash
kubectl apply -f deploy/kubernetes/
```

### Production Considerations

1. **Database**: Use PostgreSQL with high availability configuration
2. **Cache**: Configure Redis cluster for caching
3. **Message Queue**: Use Kafka with appropriate replication
4. **Monitoring**: Enable all monitoring and alerting
5. **Security**: Enable TLS and proper authentication
6. **Scaling**: Configure horizontal pod autoscaling

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is proprietary software of the Crypto State Infrastructure Contractor (CSIC) Platform.

## Support

For support and questions:
- Create an issue in the repository
- Contact the platform team
- Review the documentation
