# CSIC Platform - Key Management Service (KMS)
# Production-grade cryptographic key management following NIST SP 800-57

## Architecture Overview

The Key Management Service follows Clean Architecture with the following layers:

### Directory Structure

```
cmd/
  └── main.go                      # Application entry point (HTTP + gRPC)

internal/
  ├── config/
  │   └── config.go                # Configuration loading and management
  
  ├── core/
  │   ├── domain/
  │   │   ├── entity.go            # Domain entities (Key, KeyVersion, KeyPolicy)
  │   │   └── exceptions.go        # Domain exceptions
  │   ├── ports/
  │   │   ├── repository.go        # Repository interfaces (data access contracts)
  │   │   └── service.go           # Service interfaces (business logic contracts)
  │   └── service/
  │       └── key_service.go       # Business logic and use cases
  │
  ├── adapter/
  │   ├── repository/
  │   │   └── postgres_repository.go     # PostgreSQL implementations
  │   ├── messaging/
  │   │   └── kafka_producer.go          # Kafka event publishing
  │   └── crypto/
  │       └── crypto_provider.go         # Cryptographic operations

config.yaml                       # Service configuration
Dockerfile                        # Docker containerization
docker-compose.yml                # Local development setup
README.md                         # Documentation
migrations/                       # Database migrations
```

### Key Features

- **Centralized Key Storage**: Secure storage of cryptographic keys with encryption at rest
- **Key Lifecycle Management**: Key generation, rotation, revocation, and archival
- **Envelope Encryption**: Data keys are encrypted by master keys (key encryption key pattern)
- **Multi-Algorithm Support**: AES-256-GCM, RSA-4096, ECDSA P-384, Ed25519
- **Key Versioning**: Full support for key rotation with version tracking
- **Access Auditing**: All key operations are logged to Kafka for audit purposes
- **gRPC API**: High-performance internal API for service-to-service communication
- **REST API**: Management API for administrative operations

### Security Standards

This implementation follows NIST SP 800-57 recommendations:
- Keys are never stored in plaintext
- Master key is kept separate from data keys
- Automatic key rotation with configurable periods
- Complete audit trail of all key operations
- Memory hygiene to prevent key material leakage

### Configuration

All configuration is managed through `config.yaml`. Key sections:

- **app**: Application metadata
- **server**: HTTP and gRPC server settings
- **database**: PostgreSQL connection settings
- **redis**: Redis connection for caching
- **kafka**: Kafka broker settings for audit events
- **security**: Master key configuration and crypto settings
- **logging**: Logging preferences

### Running the Service

```bash
# Local development
go run cmd/main.go -config=config.yaml

# Docker
docker build -t csic-key-management .
docker run -p 8080:8080 -p 50051:50051 csic-key-management

# Docker Compose
docker-compose up -d
```

### API Endpoints

#### REST API (Management)

- `POST /api/v1/keys` - Create a new key
- `GET /api/v1/keys` - List all keys
- `GET /api/v1/keys/:id` - Get key details
- `POST /api/v1/keys/:id/rotate` - Rotate a key
- `POST /api/v1/keys/:id/revoke` - Revoke a key
- `POST /api/v1/keys/generate-data-key` - Generate a data key (envelope encryption)
- `POST /api/v1/keys/decrypt` - Decrypt data using a key

#### gRPC API (Service-to-Service)

- `GenerateKey`: Generate a new cryptographic key
- `GenerateDataKey`: Generate a data key for envelope encryption
- `Encrypt`: Encrypt data using a key
- `Decrypt`: Decrypt data using a key
- `RotateKey`: Rotate a key to a new version
- `RevokeKey`: Revoke a key
- `GetKeyMetadata`: Get metadata for a key

### Database Schema

The service uses PostgreSQL with the following main tables:

- **key_vault**: Stores key metadata and current state
- **key_versions**: Stores encrypted key material for each version
- **key_audit_log**: Complete audit trail of all key operations

### Dependencies

- **Gin**: HTTP web framework
- **gRPC**: High-performance RPC framework
- **Viper**: Configuration management
- **Zap**: Structured logging
- **Redis**: Caching for public keys and metadata
- **PostgreSQL**: Persistent storage
- **Kafka**: Event streaming for audit logs

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
