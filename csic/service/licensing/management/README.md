# Licensing Management Service

## Overview

The Licensing Management Service is a Node.js-based microservice built with NestJS that handles the complete lifecycle of cryptocurrency exchange licensing (VASP/CASP). This service provides workflow-driven license application processing, document management, compliance tracking, and renewal alerts.

## Features

### Core Functionality
- **License Application Workflow**: State machine-driven application processing from draft to approval/rejection
- **Document Management**: Secure document upload, verification, and integrity checking via checksums
- **License Lifecycle Management**: Issue, suspend, renew, and revoke licenses with full audit trails
- **Renewal Alerts**: Automated alerting system for upcoming license expirations (30, 60, 90 days)
- **Compliance Dashboards**: Real-time metrics on application queue, approval rates, and compliance status

### Technical Features
- **State Machine Architecture**: Hard-coded workflow transitions ensuring regulatory compliance
- **Queue-Based Processing**: BullMQ integration for asynchronous document processing
- **Event-Driven Architecture**: Kafka integration for publishing license events
- **API Gateway Ready**: RESTful API designed for Kong/Nginx integration
- **Full Observability**: Prometheus metrics, structured logging, and Grafana dashboards

## Architecture

### Technology Stack
- **Runtime**: Node.js 18+ (LTS)
- **Framework**: NestJS 10.x
- **Database**: PostgreSQL with TypeORM
- **Queue**: BullMQ (Redis-based)
- **File Storage**: AWS S3 compatible storage
- **Messaging**: Apache Kafka
- **Validation**: class-validator + class-transformer

### Directory Structure
```
service/licensing/management/
├── src/
│   ├── modules/
│   │   ├── applications/     # License application logic
│   │   ├── licenses/         # License lifecycle management
│   │   ├── documents/        # Document handling
│   │   ├── workflow/         # State machine logic
│   │   └── renewals/         # Renewal alert system
│   ├── common/
│   │   ├── guards/           # Auth guards
│   │   ├── interceptors/     # Response interceptors
│   │   ├── filters/          # Exception filters
│   │   └── decorators/       # Custom decorators
│   ├── config/               # Configuration
│   └── main.ts               # Application entry point
├── internal/
│   ├── db/
│   │   └── migrations/       # Database migrations
│   ├── deploy/
│   │   ├── docker-compose.yml
│   │   └── Dockerfile
│   └── monitoring/
│       ├── prometheus.yml
│       └── grafana/
├── test/                      # E2E tests
├── package.json
├── tsconfig.json
└── README.md
```

## API Endpoints

### Applications
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/applications` | Submit new license application |
| GET | `/api/v1/applications` | List all applications with filtering |
| GET | `/api/v1/applications/:id` | Get application details |
| PATCH | `/api/v1/applications/:id` | Update application (draft state) |
| GET | `/api/v1/applications/:id/workflow` | Get workflow status and next steps |
| POST | `/api/v1/applications/:id/submit` | Submit application for review |

### Licenses
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/licenses` | List all active licenses |
| GET | `/api/v1/licenses/:id` | Get license details |
| POST | `/api/v1/licenses/:id/suspend` | Suspend license |
| POST | `/api/v1/licenses/:id/revoke` | Revoke license |
| POST | `/api/v1/licenses/:id/renew` | Initiate renewal process |
| GET | `/api/v1/licenses/expiring` | Get licenses approaching expiry |

### Documents
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/documents/upload` | Upload document (multipart) |
| GET | `/api/v1/documents/:id` | Get document metadata |
| GET | `/api/v1/documents/:id/download` | Get signed download URL |
| POST | `/api/v1/documents/:id/verify` | Verify document integrity |

### Dashboard
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/dashboard/stats` | Get compliance dashboard statistics |
| GET | `/api/v1/dashboard/queue` | Get application queue status |
| GET | `/api/v1/dashboard/approvals` | Get approval rate metrics |

## Workflow States

```
DRAFT → SUBMITTED → UNDER_REVIEW → (APPROVED | REJECTED)
                                    ↓
                              (RENEWAL_WINDOW)
                                    ↓
                              RENEWAL_IN_PROGRESS → ACTIVE
                                    ↓
                              (SUSPENSION/REVOCATION)
                                    ↓
                              (SUSPENDED | EXPIRED | REVOKED)
```

## Configuration

### Environment Variables
```env
# Application
NODE_ENV=development
PORT=3000

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=csic_user
DATABASE_PASSWORD=your_password
DATABASE_NAME=csic_licensing

# Redis (for BullMQ)
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_CLIENT_ID=licensing-service

# S3 Storage
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=your_access_key
S3_SECRET_KEY=your_secret_key
S3_BUCKET=csic-documents
S3_REGION=us-east-1

# JWT Auth
JWT_SECRET=your_jwt_secret
JWT_EXPIRY=24h
```

## Getting Started

### Prerequisites
- Node.js 18+
- PostgreSQL 14+
- Redis 6+
- Kafka 3.x
- S3-compatible storage (MinIO/AWS S3)

### Installation
```bash
# Install dependencies
npm install

# Run database migrations
npm run migration:run

# Start development server
npm run start:dev

# Build for production
npm run build

# Run production build
npm run start:prod
```

### Docker Deployment
```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f licensing-service
```

## Integration Points

### Kafka Topics
- `license.application.submitted` - New application submitted
- `license.application.approved` - Application approved
- `license.application.rejected` - Application rejected
- `license.issued` - New license issued
- `license.expiring` - License expiring soon
- `license.expired` - License expired

### Neo4j Relationships
- `(Entity)-[:HAS_LICENSE]->(License)`
- `(License)-[:ISSUED_BY]->(Regulator)`
- `(Application)-[:SUBMITTED_BY]->(Entity)`

## License

Internal CSIC Platform Component
