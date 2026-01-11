# NEAM Reporting & Evidence System

The Reporting & Evidence System is a critical component of the National Economic Alert Mechanism (NEAM) platform, providing comprehensive real-time monitoring, statutory reporting, forensic evidence management, and long-term data archival capabilities.

## Architecture Overview

The reporting system consists of four primary services that work together to provide a complete economic monitoring and compliance solution:

### 1. Real-Time Dashboard Service
The Real-Time Dashboard Service provides instantaneous visibility into key economic indicators and metrics. This service continuously collects data from multiple sources including payment systems, energy grids, transportation networks, and industrial output metrics. The service employs a sophisticated anomaly detection engine that identifies unusual patterns in real-time, enabling rapid response to potential economic disruptions. Dashboard widgets can be customized to display metrics relevant to different user roles and organizational needs.

### 2. Statutory Reporting Service
The Statutory Reporting Service handles the generation and management of mandatory reports required by parliamentary committees and the Ministry of Finance. This service ensures complete compliance with regulatory requirements by automatically gathering data from across the NEAM platform and formatting it according to established templates and standards. The service supports multiple report types including quarterly economic reviews, annual financial statements, and special investigation reports. Export capabilities allow reports to be generated in PDF, Excel, CSV, and JSON formats for distribution to relevant stakeholders.

### 3. Forensic Evidence Service
The Forensic Evidence Service provides tamper-proof evidence collection and verification capabilities essential for legal and regulatory proceedings. Every intervention action and economic event is cryptographically logged with a complete chain of custody, ensuring that all evidence can be verified and authenticated. The service implements a Merkle tree-based proof system that makes any modification to historical data immediately detectable. Evidence chains can be sealed to prevent further modifications, creating an immutable audit trail for regulatory compliance and dispute resolution.

### 4. Archives Service
The Archives Service manages the long-term storage and retention of economic data according to regulatory requirements. The service implements a tiered storage architecture that automatically moves data between hot, warm, and cold storage tiers based on access patterns and retention policies. Automated compression and encryption ensure that archived data remains secure while minimizing storage costs. The service supports flexible retrieval workflows for accessing historical data, with priority-based options for time-sensitive investigations.

## Directory Structure

```
reporting/
├── main.go                     # Main entry point and HTTP handlers
├── config.go                   # Configuration management
├── Dockerfile                  # Container build configuration
├── docker-compose.yml          # Local development environment
├── README.md                   # This documentation
├── go.mod                      # Go module definition
├── real-time/
│   ├── service.go              # Real-time monitoring implementation
│   └── go.mod                  # Real-time module
├── statutory/
│   ├── service.go              # Statutory reporting implementation
│   └── go.mod                  # Statutory module
├── forensic/
│   ├── service.go              # Forensic evidence implementation
│   └── go.mod                  # Forensic module
└── archives/
    ├── service.go              # Archives implementation
    └── go.mod                  # Archives module
```

## API Endpoints

### Real-Time Metrics
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/metrics/overview` | GET | Get aggregated overview metrics |
| `/api/v1/metrics/payments` | GET | Get real-time payment statistics |
| `/api/v1/metrics/energy` | GET | Get energy grid metrics |
| `/api/v1/metrics/transport` | GET | Get transport and logistics metrics |
| `/api/v1/metrics/industry` | GET | Get industrial output metrics |
| `/api/v1/metrics/inflation` | GET | Get inflation tracking metrics |
| `/api/v1/metrics/anomalies` | GET | Get active anomaly alerts |

### Dashboards
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/dashboards` | GET | List available dashboards |
| `/api/v1/dashboards/:id` | GET | Get dashboard details |
| `/api/v1/dashboards/:id/export` | GET | Export dashboard to PDF |

### Statutory Reports
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/statutory/parliament` | GET | List parliamentary reports |
| `/api/v1/statutory/parliament/:id` | GET | Get specific parliamentary report |
| `/api/v1/statutory/parliament/generate` | POST | Generate new parliamentary report |
| `/api/v1/statutory/finance-ministry` | GET | List finance ministry reports |
| `/api/v1/statutory/finance-ministry/generate` | POST | Generate finance ministry report |
| `/api/v1/statutory/exports` | GET | List report exports |
| `/api/v1/statutory/exports/:id` | GET | Get export status |
| `/api/v1/statutory/exports/:id/download` | POST | Download exported report |

### Forensic Evidence
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/forensic/evidence` | GET | List evidence records |
| `/api/v1/forensic/evidence/:id` | GET | Get specific evidence |
| `/api/v1/forensic/evidence/generate` | POST | Generate new evidence record |
| `/api/v1/forensic/tamper-check` | POST | Check evidence for tampering |
| `/api/v1/forensic/verify` | POST | Verify evidence authenticity |
| `/api/v1/forensic/chain` | GET | Get evidence chain status |

### Archives
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/archives` | GET | List archives |
| `/api/v1/archives/:id` | GET | Get archive details |
| `/api/v1/archives/retrieve` | POST | Request data retrieval |
| `/api/v1/archives/retention` | GET | Get retention policies |

## Configuration

The reporting service is configured through environment variables or a configuration file. Key configuration options include:

### Database Connections
- `CLICKHOUSE_HOST`: ClickHouse server hostname
- `CLICKHOUSE_PORT`: ClickHouse server port
- `CLICKHOUSE_DATABASE`: Target database name
- `POSTGRES_HOST`: PostgreSQL server hostname
- `POSTGRES_PORT`: PostgreSQL server port
- `POSTGRES_DATABASE`: Target database name

### Cache and Message Queue
- `REDIS_HOST`: Redis server hostname
- `REDIS_PORT`: Redis server port
- `KAFKA_BROKERS`: Kafka broker addresses

### Storage
- `MINIO_ENDPOINT`: MinIO/S3 endpoint
- `EXPORT_DIR`: Directory for report exports
- `ARCHIVE_DIR`: Directory for data archives

### Retention Settings
- `DAILY_RETENTION_DAYS`: Retention period for daily archives (default: 90)
- `MONTHLY_RETENTION_YEARS`: Retention period for monthly archives (default: 10)
- `QUARTERLY_RETENTION_YEARS`: Retention period for quarterly archives (default: 20)

## Getting Started

### Local Development

1. Start the development environment using Docker Compose:
```bash
docker-compose up -d
```

2. Access the reporting service at http://localhost:8082

3. Check service health:
```bash
curl http://localhost:8082/health
```

### Building from Source

1. Clone the repository and navigate to the reporting directory:
```bash
cd neam-platform/reporting
```

2. Download dependencies:
```bash
go mod download
```

3. Build the service:
```bash
go build -o reporting-service .
```

4. Run locally:
```bash
./reporting-service
```

### Running Tests

```bash
go test ./...
```

## Deployment

### Kubernetes Deployment

The reporting service can be deployed to Kubernetes using the provided Helm charts. Ensure you have the necessary secrets and configmaps configured:

```bash
helm install neam-reporting ./helm/reporting --values values.yaml
```

### Docker Deployment

Build and push the container image:

```bash
docker build -t neam-platform/reporting:latest .
docker push neam-platform/reporting:latest
```

## Dependencies

The reporting service depends on the following infrastructure components:

- **ClickHouse**: Analytics and time-series data storage
- **PostgreSQL**: Metadata and configuration storage
- **Redis**: Caching and real-time metrics
- **Kafka**: Event streaming and async processing
- **MinIO**: Object storage for archives and exports

## Architecture Highlights

### Real-Time Processing
The real-time service employs a sophisticated event-driven architecture that processes incoming data streams with minimal latency. Metrics are collected every 30 seconds and cached in Redis for rapid access by dashboard clients. Anomaly detection runs as a background process, continuously monitoring for unusual patterns that might indicate economic stress or disruption.

### Evidence Chain Integrity
Every piece of evidence is cryptographically linked to its predecessor through a hash chain. The Merkle proof system allows efficient verification of any evidence record without requiring access to the complete chain. When evidence chains are sealed, the root hash is recorded, creating an immutable timestamp that can be verified by third parties.

### Tiered Storage
The archives service implements a sophisticated tiered storage strategy. Recent data remains in hot storage for rapid access, while older data is automatically migrated to warm and cold storage tiers. This approach balances access performance with storage costs, ensuring that frequently accessed data remains readily available while historical archives are preserved cost-effectively.

## Contributing

When contributing to the reporting service, please follow the established coding patterns and ensure all new code includes appropriate tests. All HTTP handlers should include OpenAPI documentation, and new services should implement the standard service interface used throughout the codebase.

## License

This component is part of the NEAM platform and is subject to the same licensing terms as the overall project.
