# Financial Crime Unit Service

A comprehensive financial crime detection and investigation service built with Go, providing transaction monitoring, sanctions screening, case management, SAR generation, fraud detection, and network analysis capabilities.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Running Tests](#running-tests)
- [Deployment](#deployment)
- [Security Considerations](#security-considerations)
- [Monitoring](#monitoring)

## Features

### Transaction Monitoring
- **Real-Time Screening**: Synchronous fraud check for critical transfers with sub-200ms latency
- **Velocity Detection**: Monitors transaction frequency and volume within configurable time windows
- **Structuring Detection**: Identifies transactions just below reporting thresholds ($9,999)
- **Geographic Risk Assessment**: Flags transactions involving high-risk or sanctioned jurisdictions
- **Behavioral Analysis**: Detects unusual transaction patterns and anomalies

### Sanctions & PEP Screening
- **Real-Time Screening**: KYC/Sanctions screening for individuals and companies
- **Fuzzy Matching**: Advanced name matching against sanction lists
- **Wallet Address Screening**: Checks blockchain addresses against watchlists
- **PEP Detection**: Identifies Politically Exposed Persons
- **Automatic List Updates**: Scheduled synchronization with official sanctions sources

### Case Management
- **Automated Case Creation**: Auto-generates cases when risk score exceeds threshold
- **Workflow Management**: Structured workflow states (New → Assigned → Investigating → SAR Filed → Closed)
- **Evidence Management**: Attach and manage evidence linked to cases
- **Audit Trail**: Complete audit logging of all case actions and state transitions
- **Assignment & Escalation**: Assign cases to investigators with escalation paths

### Regulatory Reporting (SAR)
- **Automated SAR Generation**: Generate Suspicious Activity Reports from case data
- **Template Engine**: XML/JSON reports compliant with regulatory standards
- **Validation**: Mandatory field validation before submission
- **Filing Tracking**: Track submission status and regulatory references
- **Compliance Deadlines**: Monitor and alert on filing deadlines

### Network Analysis
- **Transaction Graph Analysis**: Visualize transaction flows and relationships
- **Pattern Detection**: Identify money laundering patterns (fan-in, fan-out, layering)
- **Transaction Tracing**: Trace funds upstream and downstream
- **Connection Mapping**: Map entity relationships and interaction patterns

### Risk Scoring
- **Multi-Factor Scoring**: Comprehensive risk assessment across multiple dimensions
- **Real-Time Updates**: Dynamic risk score updates based on behavior
- **Historical Tracking**: Track risk score changes over time
- **Threshold Alerts**: Automated alerts when risk exceeds defined thresholds

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Financial Crime Unit Service                    │
├─────────────────────────────────────────────────────────────────┤
│  API Gateway                                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                      HTTP Handlers                           ││
│  │  Screening | Sanctions | Cases | SAR | Risk | Network | FL  ││
│  └─────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                        Service Layer                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │              FCU Service (Business Logic)                    ││
│  │  - Transaction Monitoring  - Sanctions Screening            ││
│  │  - Case Management         - SAR Generation                  ││
│  │  - Risk Scoring            - Network Analysis                ││
│  └─────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                      Domain & Repository                         │
│  ┌──────────────────┐  ┌──────────────────────────────────────┐ ││
│  │   Domain Models  │  │         Repository Layer             │ ││
│  │  - Alert         │  │  - PostgreSQL (Cases, SARs, Alerts)  │ ││
│  │  - Case          │  │  - Redis (Velocity, Risk Scores)     │ ││
│  │  - SAR           │  │  - Kafka (Events, Reports)           │ ││
│  │  - Transaction   │  │                                      │ ││
│  │  - Entity        │  │                                      │ ││
│  └──────────────────┘  └──────────────────────────────────────┘ │└─────────────────────────────────────────────────────────────┘
├─────────────────────────────────────────────────────────────────┤
│                     Infrastructure                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  PostgreSQL │  │    Redis    │  │        Kafka             │  │
│  │  (FCU DB)   │  │  (Cache)    │  │  (Alerts, SARs, Reports)│  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (for containerized deployment)
- Make (optional, for build automation)

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/csic-platform.git
cd csic-platform/extended-services/financial-crime-unit

# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8087/health
```

### Manual Setup

```bash
# Clone the repository
git clone https://github.com/your-org/csic-platform.git
cd csic-platform/extended-services/financial-crime-unit

# Initialize PostgreSQL database
createdb csic_platform
psql -U postgres -d csic_platform -f init-scripts/schema.sql

# Build the application
go build -o financial-crime-unit-service ./cmd/server

# Run the service
./financial-crime-unit-service -config config.yaml
```

## Configuration

The service is configured via `config.yaml`:

```yaml
# Application Settings
app:
  name: "financial-crime-unit-service"
  host: "0.0.0.0"
  port: 8087
  environment: "production"
  log_level: "info"

# Database Configuration
database:
  host: "10.112.2.4"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "csic_platform"
  schema: "financial_crime"
  max_open_conns: 100

# Redis Configuration
redis:
  host: "10.112.2.4"
  port: 6379
  key_prefix: "csic:fc:"
  pool_size: 50

# Kafka Configuration
kafka:
  brokers:
    - "10.112.2.4:9092"
  topics:
    alerts: "csic.fc.alerts"
    sars: "csic.fc.sars"

# Financial Crime Detection Settings
detection:
  thresholds:
    high_risk: 85
    medium_risk: 60
    critical_alert_score: 90
  transaction:
    velocity_window: "1h"
    max_transactions_per_window: 100
    structuring_threshold: 9999
  geographic:
    high_risk_countries: ["XX", "YY"]
    sanctioned_countries: ["ZZ"]

# Case Management Settings
case_management:
  auto_create_cases_above_score: 75
  workflow:
    - state: "NEW"
      next_states: ["ASSIGNED", "DISMISSED"]

# SAR Settings
sar:
  enabled: true
  regulatory_body: "FINCEN"
  auto_generate_threshold: 95
  filing_deadline_hours: 30
```

## API Documentation

### Transaction Screening

#### Screen Transaction
```bash
POST /api/v1/screen/transaction
Content-Type: application/json

{
  "transaction_id": "tx-123",
  "transaction_hash": "0xabc123...",
  "type": "transfer",
  "asset": "USDT",
  "amount": 50000,
  "currency": "USD",
  "sender_id": "user-456",
  "receiver_id": "user-789",
  "timestamp": "2024-01-15T10:30:00Z",
  "metadata": {
    "geo": {"country": "US"}
  }
}
```

#### Screen Entity
```bash
POST /api/v1/screen/entity
Content-Type: application/json

{
  "entity_id": "user-456",
  "entity_type": "individual",
  "name": "John Doe",
  "wallets": ["0xabc123..."],
  "national_id": "123456789"
}
```

### Sanctions Screening

#### Check Sanctions
```bash
POST /api/v1/sanctions/check
Content-Type: application/json

{
  "entity_id": "user-456",
  "entity_type": "individual",
  "name": "John Doe",
  "wallets": ["0xabc123..."]
}
```

### Case Management

#### Create Case
```bash
POST /api/v1/cases
Content-Type: application/json

{
  "title": "Suspicious Transaction Pattern",
  "description": "Multiple high-value transactions detected",
  "subject_id": "user-456",
  "subject_type": "individual",
  "priority": "high",
  "risk_score": 85,
  "created_by": "investigator-1"
}
```

#### Transition Workflow
```bash
POST /api/v1/cases/{id}/workflow
Content-Type: application/json

{
  "new_status": "INVESTIGATING",
  "user_id": "investigator-1"
}
```

### SAR Management

#### Create SAR
```bash
POST /api/v1/sar
Content-Type: application/json

{
  "case_id": "case-uuid",
  "regulatory_body": "FINCEN",
  "report_type": "SAR",
  "subject_id": "user-456",
  "subject_name": "John Doe",
  "subject_type": "individual",
  "total_amount": 150000,
  "currency": "USD",
  "narrative": "Detailed description of suspicious activity...",
  "suspicious_activity": "Pattern of structuring detected...",
  "filing_institution": "CSIC Bank",
  "filing_user_id": "compliance-officer-1"
}
```

#### Submit SAR
```bash
POST /api/v1/sar/{id}/submit
Content-Type: application/json

{
  "submitted_by": "compliance-officer-1"
}
```

### Risk Scoring

#### Get Risk Score
```bash
GET /api/v1/risk/score/{entity_id}
```

#### Calculate Risk
```bash
POST /api/v1/risk/calculate
Content-Type: application/json

{
  "entity_id": "user-456",
  "transaction": {
    "transaction_id": "tx-123",
    "amount": 50000,
    "currency": "USD"
  }
}
```

### Network Analysis

#### Analyze Network
```bash
POST /api/v1/network/analyze
Content-Type: application/json

{
  "entity_id": "user-456",
  "max_hops": 3
}
```

#### Trace Transaction Flow
```bash
POST /api/v1/network/trace
Content-Type: application/json

{
  "transaction_id": "tx-123",
  "direction": "both",
  "max_hops": 5
}
```

### Watchlist Management

#### Add to Watchlist
```bash
POST /api/v1/entities/{id}/watchlist-add
Content-Type: application/json

{
  "entity_id": "user-456",
  "entity_type": "individual",
  "reason": "Suspicious activity detected",
  "risk_level": "high",
  "added_by": "investigator-1",
  "tags": ["high-risk", "monitoring"]
}
```

### Reports

#### Get Report Summary
```bash
GET /api/v1/reports/summary?date_from=2024-01-01&date_to=2024-01-31
```

#### Get Compliance Report
```bash
GET /api/v1/reports/compliance?regulatory_body=FINCEN&period=monthly
```

## Running Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test categories
go test ./internal/domain/... -v
go test ./internal/service/... -v
go test ./internal/handlers/... -v
```

## Deployment

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: financial-crime-unit-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: financial-crime-unit-service
  template:
    metadata:
      labels:
        app: financial-crime-unit-service
    spec:
      containers:
      - name: financial-crime-unit-service
        image: financial-crime-unit-service:latest
        ports:
        - containerPort: 8087
        env:
        - name: CONFIG_PATH
          value: "/app/config.yaml"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: financial-crime-unit-service
spec:
  selector:
    app: financial-crime-unit-service
  ports:
  - port: 8087
    targetPort: 8087
  type: LoadBalancer
```

### Docker Build and Push

```bash
# Build the image
docker build -t financial-crime-unit-service:latest .

# Tag for registry
docker tag financial-crime-unit-service:latest registry.example.com/financial-crime-unit-service:v1.0.0

# Push to registry
docker push registry.example.com/financial-crime-unit-service:v1.0.0
```

## Security Considerations

### Data Protection
- All sensitive data is encrypted at rest using AES-256
- TLS 1.3 is required for all connections
- PII is masked in logs
- Audit logs are immutable and retained per compliance requirements

### Access Control
- Role-based access control (RBAC) for investigators
- Multi-level approval for SAR submission
- Session management with automatic timeout
- Failed attempt rate limiting

### Compliance
- GDPR-compliant data handling
- AML/KYC regulatory compliance
- SAR filing within regulatory deadlines
- Complete audit trail for regulatory audits

### Alert Handling
- High-severity alerts trigger immediate notification
- Case escalation paths defined for critical alerts
- Automated case creation for high-risk scenarios
- SLA tracking for case resolution

## Monitoring

### Prometheus Metrics

The service exposes the following metrics:

- `fcu_transactions_screened_total` - Total transactions screened
- `fcu_alerts_generated_total` - Total alerts generated
- `fcu_cases_created_total` - Total cases created
- `fcu_sars_generated_total` - Total SARs generated
- `fcu_sanctions_hits_total` - Total sanctions matches
- `fcu_screening_duration_seconds` - Screening operation latency
- `fcu_risk_score_distribution` - Risk score distribution

### Grafana Dashboards

Pre-configured dashboards are available for:
- Transaction Monitoring Overview
- Alert Analysis Dashboard
- Case Management Metrics
- SAR Filing Compliance
- Network Analysis Visualization
- Risk Score Trends
- System Performance

### Health Endpoints

- `GET /health` - Service health check
- `GET /ready` - Readiness probe
- `GET /live` - Liveness probe
- `GET /metrics` - Prometheus metrics

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue in the repository or contact the development team.
