# Regulatory Reporting Service

## Overview

The Regulatory Reporting Service is a Node.js-based microservice that handles the generation, scheduling, and storage of regulatory reports for the CSIC platform. It provides automated report generation with PDF and CSV exports, scheduled reporting via cron jobs, and immutable WORM storage for compliance with regulatory audit requirements.

## Features

### Core Functionality
- **Report Generation**: PDF and CSV report generation with customizable templates
- **Automated Scheduling**: Cron-based scheduled report generation
- **WORM Storage**: Immutable storage with S3 Object Lock compliance
- **Report Versioning**: Full audit trail of all generated reports
- **Template Management**: Dynamic report template system
- **Chain of Custody**: Complete audit trail for report access and download

### Technical Features
- **Hybrid Architecture**: Node.js for API orchestration, Python workers for PDF generation
- **Template Engine**: Handlebars and ReportLab integration
- **Scheduled Tasks**: BullMQ with cron patterns for recurring reports
- **Immutable Storage**: S3 Object Lock with configurable retention periods
- **Data Consistency**: PostgreSQL transaction snapshots for report data

## Architecture

### Technology Stack
- **Runtime**: Node.js 18+ (API), Python 3.10+ (PDF Workers)
- **Framework**: Express.js (Node), ReportLab (Python PDF)
- **Database**: PostgreSQL
- **Queue**: BullMQ (Redis-based)
- **Storage**: AWS S3 with Object Lock (WORM compliant)
- **Scheduling**: BullMQ with cron support

### Directory Structure
```
service/reporting/regulatory/
├── src/
│   ├── api/
│   │   └── endpoints/       # API route handlers
│   ├── core/                # Configuration, services
│   ├── models/              # Database models
│   ├── services/            # Business logic
│   ├── workers/             # Python PDF workers
│   └── main.js              # Application entry point
├── internal/
│   ├── db/
│   │   └── migrations/      # Database migrations
│   ├── deploy/
│   │   ├── docker-compose.yml
│   │   ├── Dockerfile.node
│   │   └── Dockerfile.python
│   └── monitoring/
│       ├── prometheus.yml
│       └── grafana/
├── templates/               # Report templates
├── tests/                   # Unit tests
├── package.json
└── README.md
```

## API Endpoints

### Reports
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/reports/generate` | Trigger ad-hoc report generation |
| GET | `/api/v1/reports/history` | Get report history with filtering |
| GET | `/api/v1/reports/:id` | Get report details |
| GET | `/api/v1/reports/:id/download` | Get signed download URL |
| GET | `/api/v1/reports/:id/verify` | Verify report integrity |

### Schedules
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/schedules` | Create new report schedule |
| GET | `/api/v1/schedules` | List all schedules |
| GET | `/api/v1/schedules/:id` | Get schedule details |
| PATCH | `/api/v1/schedules/:id` | Update schedule |
| DELETE | `/api/v1/schedules/:id` | Delete schedule |

### Templates
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/templates` | Create new template |
| GET | `/api/v1/templates` | List templates |
| GET | `/api/v1/templates/:id` | Get template |
| PUT | `/api/v1/templates/:id` | Update template |

## Configuration

### Environment Variables
```env
# Application
NODE_ENV=development
PORT=3001
REDIS_HOST=localhost
REDIS_PORT=6379

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=csic_reporting
DATABASE_PASSWORD=your_password
DATABASE_NAME=csic_reporting

# S3/WORM Storage
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=csic-reports
S3_REGION=us-east-1
WORM_RETENTION_DAYS=2555  # 7 years for regulatory compliance

# Python Worker
PYTHON_WORKER_ENABLED=true
PYTHON_WORKER_CONCURRENCY=4

# JWT Auth
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h
```

## Getting Started

### Prerequisites
- Node.js 18+
- Python 3.10+
- PostgreSQL 14+
- Redis 6+
- S3-compatible storage (MinIO/AWS S3)

### Installation
```bash
# Install Node.js dependencies
npm install

# Install Python dependencies
pip install -r workers/requirements.txt

# Run database migrations
npm run migration:run

# Start the service
npm run start

# Start Python workers
python workers/pdf_worker.py
```

### Docker Deployment
```bash
docker-compose up -d
```

## Integration Points

### Kafka Events
- `report.generated` - Report generation completed
- `report.failed` - Report generation failed
- `report.scheduled` - Scheduled report triggered

### Neo4j Relationships
- `(Report)-[:GENERATED_FOR]->(Entity)`
- `(Report)-[:USES_TEMPLATE]->(Template)`
- `(Schedule)-[:GENERATES]->(Report)`

## WORM Compliance

### Object Lock Configuration
The service uses S3 Object Lock with:
- **Retention Mode**: COMPLIANCE (immutable)
- **Retention Period**: 7 years (configurable)
- **Legal Hold**: Enabled for critical reports

### Integrity Verification
- SHA-256 hash of report stored in database
- Hash verification on every download
- Tamper detection alerts

## License

Internal CSIC Platform Component
