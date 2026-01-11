# Reporting Infrastructure Service

The Reporting Infrastructure Service is a core component of the CSIC Platform responsible for generating, managing, and distributing comprehensive reports for compliance, risk, audit, and exposure analysis.

## Architecture

This service follows Clean Architecture principles with clear separation between:

- **Domain Layer**: Core business entities (Report, ReportTemplate, ScheduledReport)
- **Ports Layer**: Interface definitions for repositories, services, and infrastructure
- **Service Layer**: Business logic for report generation and template management
- **Adapter Layer**: External system integrations (PostgreSQL, storage, formatters)

## Features

### Report Generation
- Support for multiple report types (Compliance, Risk, Audit, Exposure)
- Multiple output formats (PDF, CSV, JSON, HTML, XLSX)
- Async report generation with status tracking
- Configurable date ranges and filters
- Real-time progress tracking

### Report Templates
- Create and manage reusable report templates
- Parameterized templates for flexibility
- Template versioning
- Default template support per report type

### Scheduled Reports
- Cron-based scheduling for automated reports
- Email distribution on completion
- Retain scheduled reports configuration
- Manual execution option

### Report Management
- View all generated reports
- Filter by type, status, date range
- Download reports in original format
- Archive old reports
- Statistics and analytics

### Integration
- PostgreSQL for persistent storage
- Local filesystem or object storage for report files
- Kafka for event publishing
- HTTP API for external access

## Project Structure

```
services/reporting/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── entity.go          # Domain entities (Report, Template, etc.)
│   │   │   └── errors.go          # Domain error definitions
│   │   ├── ports/
│   │   │   ├── repository.go      # Repository interface definitions
│   │   │   ├── service.go         # Service interface definitions
│   │   │   └── infrastructure.go  # Infrastructure interface definitions
│   │   └── service/
│   │       ├── report_service.go  # Report generation and management
│   │       └── template_service.go # Template management
│   ├── adapter/
│   │   ├── formatter/             # Report formatters (PDF, CSV, etc.)
│   │   ├── generator/             # Report content generators
│   │   ├── repository/            # PostgreSQL repositories
│   │   └── storage/               # File storage adapters
│   └── handler/
│       └── http_handler.go        # HTTP API handlers
├── migrations/
│   └── 001_create_tables.sql      # Database schema
├── go.mod                         # Go module definition
└── README.md                      # This file
```

## API Endpoints

### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/reports` | Generate a new report |
| GET | `/api/v1/reports` | List all reports |
| GET | `/api/v1/reports/{id}` | Get report details |
| GET | `/api/v1/reports/{id}/download` | Download report file |
| DELETE | `/api/v1/reports/{id}` | Delete a report |
| POST | `/api/v1/reports/{id}/archive` | Archive a report |
| GET | `/api/v1/reports/statistics` | Get report statistics |

### Templates

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/templates` | Create a new template |
| GET | `/api/v1/templates` | List all templates |
| GET | `/api/v1/templates/{id}` | Get template details |
| PUT | `/api/v1/templates/{id}` | Update a template |
| DELETE | `/api/v1/templates/{id}` | Delete a template |

### Scheduled Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/scheduled` | Create a scheduled report |
| GET | `/api/v1/scheduled` | List all scheduled reports |
| GET | `/api/v1/scheduled/{id}` | Get scheduled report details |
| PUT | `/api/v1/scheduled/{id}` | Update a scheduled report |
| DELETE | `/api/v1/scheduled/{id}` | Delete a scheduled report |
| POST | `/api/v1/scheduled/{id}/enable` | Enable a scheduled report |
| POST | `/api/v1/scheduled/{id}/disable` | Disable a scheduled report |
| POST | `/api/v1/scheduled/{id}/execute` | Execute immediately |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check endpoint |

## Report Types

| Type | Description |
|------|-------------|
| `COMPLIANCE` | Compliance status and violation reports |
| `RISK` | Risk assessment and alert summaries |
| `AUDIT` | Audit log and activity reports |
| `EXPOSURE` | Position and exposure reports |
| `TRANSACTION` | Transaction history reports |
| `PERFORMANCE` | Performance and analytics reports |

## Output Formats

| Format | MIME Type | Description |
|--------|-----------|-------------|
| `PDF` | application/pdf | Adobe PDF format |
| `CSV` | text/csv | Comma-separated values |
| `JSON` | application/json | JavaScript Object Notation |
| `XLSX` | application/vnd.openxmlformats-officedocument.spreadsheetml.spreadsheet | Excel format |
| `HTML` | text/html | HyperText Markup Language |

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_HOST` | HTTP server host | `0.0.0.0` |
| `APP_PORT` | HTTP server port | `8082` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `csic` |
| `DB_PASSWORD` | PostgreSQL password | `csic_secret` |
| `DB_NAME` | PostgreSQL database | `csic_platform` |
| `STORAGE_PATH` | Report file storage path | `/app/reports` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `TOPIC_PREFIX` | Kafka topic prefix | `csic` |
| `LOG_LEVEL` | Logging level | `info` |

## Installation

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- LibreOffice (for PDF conversion, optional)

### Building

```bash
cd services/reporting
go build -o bin/reporting ./cmd
```

### Running

```bash
./bin/reporting
```

### Database Migration

```bash
psql -h localhost -U csic -d csic_platform -f migrations/001_create_tables.sql
```

## Usage Examples

### Generate a Compliance Report

```bash
curl -X POST http://localhost:8082/api/v1/reports \
  -H "Content-Type: application/json" \
  -d '{
    "type": "COMPLIANCE",
    "format": "PDF",
    "name": "Monthly Compliance Report",
    "description": "Monthly compliance status report",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "filters": {
      "severities": ["HIGH", "CRITICAL"]
    }
  }'
```

### List All Reports

```bash
curl http://localhost:8082/api/v1/reports
```

### Download a Report

```bash
curl -O http://localhost:8082/api/v1/reports/{report_id}/download
```

### Create a Report Template

```bash
curl -X POST http://localhost:8082/api/v1/templates \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekly Risk Summary",
    "description": "Weekly risk assessment summary template",
    "type": "RISK",
    "format": "PDF",
    "content": "...",
    "parameters": [
      {
        "name": "date_range",
        "type": "string",
        "required": true,
        "description": "Report date range"
      }
    ]
  }'
```

### Schedule a Daily Report

```bash
curl -X POST http://localhost:8082/api/v1/scheduled \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Risk Report",
    "description": "Daily risk assessment report",
    "template_id": "template-uuid",
    "type": "RISK",
    "format": "PDF",
    "cron_schedule": "0 8 * * *",
    "time_zone": "UTC",
    "recipients": ["risk-team@example.com"]
  }'
```

## Report Generation Flow

1. **Request Submission**: Client submits report request with type, format, and date range
2. **Validation**: Request is validated for completeness and permissions
3. **Record Creation**: Report record is created in database with PENDING status
4. **Data Fetching**: Report generator fetches required data from data providers
5. **Content Generation**: Report content is generated based on type
6. **Formatting**: Content is formatted to the requested output format
7. **Storage**: Report file is stored in the configured storage backend
8. **Completion**: Report status is updated to COMPLETED
9. **Notification**: If scheduled, recipients are notified of completion

## Monitoring

The service exposes the following metrics:

- `reports_generated_total`: Total reports generated by type and format
- `reports_downloaded_total`: Total report downloads by type
- `report_generation_time`: Report generation duration
- `storage_used_bytes`: Total storage used by reports

## Testing

```bash
# Unit tests
go test ./... -v

# Integration tests
go test ./... -tags=integration -v
```

## License

This project is part of the CSIC Platform and is licensed under the MIT License.
