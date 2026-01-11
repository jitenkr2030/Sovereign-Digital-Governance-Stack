# CSIC Platform - Forensic Tools Service
# Digital forensics analysis, evidence management, and chain of custody

## Architecture Overview

The Forensic Tools Service follows Clean Architecture with the following layers:

### Directory Structure

```
cmd/
  └── main.go                      # Application entry point

internal/
  ├── config/
  │   └── config.go                # Configuration loading and management
  
  ├── core/
  │   ├── domain/
  │   │   ├── entity.go            # Domain entities (Evidence, AnalysisJob, ChainOfCustody)
  │   │   └── exceptions.go        # Domain exceptions
  │   ├── ports/
  │   │   ├── repository.go        # Repository interfaces
  │   │   └── service.go           # Service interfaces
  │   └── service/
  │       └── forensic_service.go  # Business logic and use cases
  │
  ├── adapter/
  │   ├── repository/
  │   │   └── postgres_repository.go     # PostgreSQL implementations
  │   ├── messaging/
  │   │   └── kafka_producer.go          # Kafka event publishing
  │   └── storage/
  │       └── blob_storage.go      # Evidence file storage (S3/MinIO/Local)

config.yaml                       # Service configuration
Dockerfile                        # Docker containerization
docker-compose.yml                # Local development setup
README.md                         # Documentation
migrations/                       # Database migrations
```

### Key Features

- **Evidence Management**: Upload, store, and manage digital evidence
- **Chain of Custody**: Complete audit trail of evidence access and handling
- **Hash Integrity**: SHA-256 file hashing for evidence integrity verification
- **Analysis Pipeline**: Async job processing for forensic analysis tools
- **Multi-format Support**: Support for various evidence types (files, disk images, memory dumps)
- **Metadata Extraction**: Extract and store file metadata
- **Event Streaming**: Kafka integration for job queuing and notifications

### Security Standards

- Evidence immutability: Once uploaded, evidence cannot be modified
- Chain of custody logging for all evidence access
- Access control for evidence operations
- Secure storage with encryption at rest

### Configuration

All configuration is managed through `config.yaml`. Key sections:

- **app**: Application metadata
- **server**: HTTP server settings
- **database**: PostgreSQL connection settings
- **storage**: Blob storage configuration (S3/MinIO)
- **kafka**: Kafka broker settings for job events
- **logging**: Logging preferences

### Running the Service

```bash
# Local development
go run cmd/main.go -config=config.yaml

# Docker
docker build -t csic-forensic-tools .
docker run -p 8080:8080 csic-forensic-tools

# Docker Compose
docker-compose up -d
```

### API Endpoints

#### REST API

- `POST /api/v1/evidence/upload` - Upload new evidence (multipart/form-data)
- `GET /api/v1/evidence` - List evidence with pagination
- `GET /api/v1/evidence/:id` - Get evidence details
- `GET /api/v1/evidence/:id/download` - Download evidence file
- `GET /api/v1/evidence/:id/custody` - Get chain of custody
- `POST /api/v1/analysis` - Request analysis of evidence
- `GET /api/v1/analysis/:id` - Get analysis job status
- `GET /api/v1/analysis/:id/results` - Get analysis results

### Database Schema

The service uses PostgreSQL with the following main tables:

- **evidence**: Stores evidence metadata and references to stored files
- **chain_of_custody**: Complete audit trail of evidence handling
- **analysis_jobs**: Tracks forensic analysis jobs
- **analysis_results**: Stores results from analysis tools

### Dependencies

- **Gin**: HTTP web framework
- **Viper**: Configuration management
- **Zap**: Structured logging
- **MinIO/S3**: Blob storage for evidence files
- **PostgreSQL**: Metadata storage
- **Kafka**: Event streaming for analysis jobs

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
