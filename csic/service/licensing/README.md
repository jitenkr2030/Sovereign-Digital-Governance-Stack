# CSIC Licensing Service

## Overview

The **CSIC Licensing Service** is a Go-based microservice that provides comprehensive licensing and compliance management for the Crypto State Infrastructure Contractor (CSIC) Platform. This service handles exchange licensing registries, compliance monitoring, automated regulatory reporting, and compliance dashboards.

## Features

### License Management
- **License Application Processing**: Handle new license applications with automated validation
- **License Lifecycle Management**: Track licenses from application through approval, renewal, suspension, and revocation
- **License Status Tracking**: Real-time status updates with complete audit trail
- **Expiry Monitoring**: Automated alerts for upcoming license expirations
- **License Amendments**: Process and track license modifications and updates

### Compliance Monitoring
- **Compliance Assessment**: Automated compliance checking against regulatory requirements
- **Violation Tracking**: Detect, track, and manage compliance violations
- **Risk Scoring**: Weighted risk assessment based on multiple factors
- **Investigation Management**: Handle compliance investigations with proper workflow
- **Enforcement Actions**: Track and manage enforcement actions and penalties

### Regulatory Reporting
- **Automated Report Generation**: Scheduled generation of regulatory reports
- **SAR Processing**: Suspicious Activity Report (SAR) filing management
- **Report Submission**: Track regulatory report submissions and acknowledgments
- **Template Management**: Manage report templates for different regulatory bodies

### Compliance Dashboard
- **Real-time Metrics**: Live compliance rate and risk score monitoring
- **Trend Analysis**: Historical trend visualization and forecasting
- **Geographic Distribution**: License and compliance distribution by jurisdiction
- **Executive Reporting**: High-level summaries for management review

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
cmd/api/main.go          # Application entry point
internal/
├── domain/               # Business logic and entities
│   └── models/          # Domain models
├── service/              # Business services
├── repository/           # Data access layer
├── handler/              # HTTP and event handlers
└── config/              # Configuration management
```

### API Structure

```
/api/v1/
├── licenses/             # License management endpoints
│   ├── GET    /licenses              # List licenses
│   ├── POST   /licenses              # Create license application
│   ├── GET    /licenses/{id}         # Get license details
│   ├── PUT    /licenses/{id}         # Update license
│   ├── POST   /licenses/{id}/renew   # Renew license
│   └── POST   /licenses/{id}/amend   # Amend license
│
├── compliance/           # Compliance management endpoints
│   ├── GET    /compliance/records    # List compliance records
│   ├── POST   /compliance/assessments # Create assessment
│   ├── GET    /compliance/violations # List violations
│   └── GET    /compliance/risk       # Get risk scores
│
├── reporting/            # Reporting endpoints
│   ├── GET    /reports              # List reports
│   ├── POST   /reports              # Create report
│   ├── GET    /reports/{id}         # Get report details
│   └── POST   /reports/{id}/submit  # Submit report
│
└── dashboard/            # Dashboard endpoints
    ├── GET    /dashboard/metrics    # Get metrics
    ├── GET    /dashboard/summary    # Get summary
    └── GET    /dashboard/trends     # Get trends
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
   cd /workspace/csic-platform/service/licensing
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
   go build -o csic-licensing-service ./cmd/api
   ./csic-licensing-service
   ```

### Using Docker

```bash
# Build the image
docker build -t csic-licensing-service .

# Run the service
docker run -d \
  --name csic-licensing \
  -p 8081:8081 \
  -p 9091:9091 \
  --network csic-network \
  csic-licensing-service
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
| `DB_HOST` | PostgreSQL host | `licensing-db` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `csic_admin` |
| `DB_PASSWORD` | Database password | `secure_password` |
| `DB_NAME` | Database name | `csic_licensing` |
| `REDIS_HOST` | Redis host | `licensing-cache` |
| `REDIS_PORT` | Redis port | `6379` |
| `KAFKA_BROKERS` | Kafka brokers | `kafka:9092` |
| `APP_PORT` | Application port | `8081` |
| `APP_MODE` | Gin mode | `release` |

### Configuration File

The `config.yaml` file contains all application settings:

```yaml
app:
  host: "0.0.0.0"
  port: 8081
  mode: "release"

database:
  host: "localhost"
  port: 5432
  username: "csic_admin"
  password: "secure_password"
  name: "csic_licensing"

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
curl http://localhost:8081/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### License Endpoints

#### Create License Application

```bash
curl -X POST http://localhost:8081/api/v1/licenses \
  -H "Content-Type: application/json" \
  -d '{
    "entity_name": "Example Exchange",
    "entity_type": "exchange",
    "license_type": "exchange_license",
    "jurisdiction": "Switzerland",
    "issuing_authority_id": "authority-uuid"
  }'
```

#### Get License Details

```bash
curl http://localhost:8081/api/v1/licenses/{license_id}
```

### Compliance Endpoints

#### Get Compliance Status

```bash
curl http://localhost:8081/api/v1/compliance/status/{license_id}
```

#### Report Violation

```bash
curl -X POST http://localhost:8081/api/v1/compliance/violations \
  -H "Content-Type: application/json" \
  -d '{
    "license_id": "license-uuid",
    "violation_type": "kyc_failure",
    "severity": "high",
    "description": "KYC verification not completed"
  }'
```

### Reporting Endpoints

#### Generate Report

```bash
curl -X POST http://localhost:8081/api/v1/reports \
  -H "Content-Type: application/json" \
  -d '{
    "report_type": "periodic",
    "license_id": "license-uuid",
    "period_start": "2024-01-01",
    "period_end": "2024-03-31"
  }'
```

## Monitoring

### Metrics Endpoint

Prometheus metrics are available at `http://localhost:8081/metrics`

Key metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `license_applications_total` - License application count
- `compliance_violations_total` - Violation count
- `reports_generated_total` - Report generation count

### Health Checks

| Endpoint | Description |
|----------|-------------|
| `/health` | Overall service health |
| `/health/ready` | Readiness probe |
| `/health/live` | Liveness probe |

### Grafana Dashboards

Pre-configured dashboards available:
- **Service Overview**: General service health and metrics
- **License Management**: License statistics and trends
- **Compliance Monitoring**: Compliance rates and violations
- **Infrastructure**: Database, cache, and Kafka metrics

## Database Schema

### Core Tables

| Table | Description |
|-------|-------------|
| `licensing_authorities` | Regulatory authorities registry |
| `exchange_licenses` | Main license records |
| `license_status_history` | License status change audit trail |
| `compliance_requirements` | Regulatory requirements catalog |
| `compliance_records` | Compliance assessment records |
| `compliance_violations` | Violation tracking |
| `enforcement_actions` | Enforcement action records |
| `regulatory_reports` | Generated reports |
| `suspicious_activity_reports` | SAR records |
| `dashboard_metrics` | Aggregated metrics |
| `compliance_audit_log` | Complete audit trail |

## Security

### Authentication

The service uses JWT-based authentication:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8081/api/v1/licenses
```

### Rate Limiting

Default rate limits:
- 1000 requests per minute
- Burst size: 100

### CORS Configuration

Cross-origin resource sharing is configurable in `config.yaml`:

```yaml
security:
  cors:
    allowed_origins:
      - "http://localhost:3000"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
```

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

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and updates.
