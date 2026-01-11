# National Identity Service

A comprehensive national identity management service built with Go, providing citizen identity verification, authentication, behavioral biometrics, fraud detection, and federated learning capabilities.

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

### Core Identity Management
- **Citizen Registration**: Create and manage national identity records with comprehensive personal information
- **Identity Verification**: Multiple verification levels (basic, standard, enhanced, biometric)
- **Document Verification**: Automated document validation and data extraction
- **Biometric Integration**: Support for fingerprint, iris, and facial recognition templates
- **Identity History**: Complete audit trail of all identity changes

### Authentication & Security
- **Multi-Factor Authentication**: TOTP, SMS, and email-based MFA support
- **Behavioral Biometrics**: Keystroke dynamics, mouse patterns, and touch analysis
- **Risk-Based Authentication**: Dynamic risk scoring based on behavioral patterns
- **Session Management**: Secure session handling with automatic expiration
- **Password Security**: Secure password hashing and reset capabilities

### Fraud Detection
- **Real-Time Risk Scoring**: Continuous fraud score calculation
- **Behavioral Anomaly Detection**: Machine learning-based anomaly detection
- **Alert Management**: Automated fraud alerts with severity levels
- **Velocity Checking**: Detection of rapid successive actions

### Advanced Features
- **Federated Learning**: Privacy-preserving model training across distributed clients
- **Privacy Controls**: GDPR-compliant data handling with consent management
- **Data Anonymization**: Automatic PII masking and anonymization
- **Consent Management**: Granular consent tracking and enforcement

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      National Identity Service                   │
├─────────────────────────────────────────────────────────────────┤
│  API Gateway                                                      │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                      HTTP Handlers                           ││
│  │  Identity | Verification | Auth | Behavioral | Fraud | FL   ││
│  └─────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                        Service Layer                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │              Identity Service (Business Logic)               ││
│  │  - Identity Management  - Verification Logic                 ││
│  │  - Authentication       - Fraud Detection                    ││
│  │  - Behavioral Analysis  - Federated Learning                 ││
│  └─────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│                      Domain & Repository                         │
│  ┌──────────────────┐  ┌──────────────────────────────────────┐ ││
│  │   Domain Models  │  │         Repository Layer             │ ││
│  │  - Identity      │  │  - PostgreSQL (Identity, Sessions)   │ ││
│  │  - Verification  │  │  - Redis (Cache, Sessions)           │ ││
│  │  - Behavioral    │  │  - Kafka (Events)                    │ ││
│  │  - Fraud         │  │                                      │ ││
│  └──────────────────┘  └──────────────────────────────────────┘ │└─────────────────────────────────────────────────────────────┘
├─────────────────────────────────────────────────────────────────┤
│                     Infrastructure                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │  PostgreSQL │  │    Redis    │  │        Kafka             │  │
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
cd csic-platform/extended-services/national-identity

# Start all services
docker-compose up -d

# Check service health
curl http://localhost:8086/health
```

### Manual Setup

```bash
# Clone the repository
git clone https://github.com/your-org/csic-platform.git
cd csic-platform/extended-services/national-identity

# Initialize PostgreSQL database
createdb csic_platform
psql -U postgres -d csic_platform -f init-scripts/schema.sql

# Build the application
go build -o national-identity-service ./cmd/server

# Run the service
./national-identity-service -config config.yaml
```

## Configuration

The service is configured via `config.yaml`:

```yaml
# Application Settings
app:
  name: "national-identity-service"
  host: "0.0.0.0"
  port: 8086
  environment: "production"
  log_level: "info"

# Database Configuration
database:
  host: "10.112.2.4"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "csic_platform"
  schema: "identity"
  max_open_conns: 100

# Redis Configuration
redis:
  host: "10.112.2.4"
  port: 6379
  key_prefix: "csic:identity:"
  pool_size: 50

# Kafka Configuration
kafka:
  brokers:
    - "10.112.2.4:9092"
  topics:
    events: "identity.events"
    verifications: "identity.verifications"

# Identity Service Settings
identity:
  fraud_score_threshold: 85
  max_login_attempts: 5
  behavioral:
    baseline_window: "30d"
    anomaly_detection_sensitivity: "high"
  biometric:
    required_factors: 2
    liveness_check_enabled: true

# Privacy Settings
privacy:
  data_retention: "7y"
  pii_encryption: true
  gdpr_compliance: true
```

## API Documentation

### Identity Management

#### Create Identity
```bash
POST /api/v1/identities
Content-Type: application/json

{
  "national_id": "123456789",
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-15",
  "place_of_birth": "Capital City",
  "gender": "male",
  "nationality": "Country",
  "current_address": {
    "street1": "123 Main St",
    "city": "Capital City",
    "state": "Central Province",
    "postal_code": "12345",
    "country": "Country",
    "address_type": "residential"
  }
}
```

#### Get Identity
```bash
GET /api/v1/identities/{id}
```

#### Update Identity
```bash
PUT /api/v1/identities/{id}
Content-Type: application/json

{
  "email": "john.doe@example.com",
  "phone": "+1234567890"
}
```

### Verification

#### Verify Identity
```bash
POST /api/v1/verify
Content-Type: application/json

{
  "identity_id": "uuid-here",
  "verification_type": "document",
  "document_data": {
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1990-01-15"
  }
}
```

#### Biometric Verification
```bash
POST /api/v1/verify/biometric
Content-Type: application/json

{
  "identity_id": "uuid-here",
  "biometric_type": "fingerprint",
  "template": "base64-encoded-template",
  "liveness_token": "liveness-session-token"
}
```

#### Liveness Check
```bash
POST /api/v1/verify/liveness
Content-Type: application/json

{
  "identity_id": "uuid-here",
  "challenge": "blink",
  "response": "base64-encoded-video-frame"
}
```

### Authentication

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "national_id": "123456789",
  "password": "secure-password",
  "device_info": {
    "device_type": "desktop",
    "os": "Windows 11",
    "browser": "Chrome",
    "ip_address": "192.168.1.100"
  }
}
```

#### Setup MFA
```bash
POST /api/v1/auth/mfa
Content-Type: application/json

{
  "identity_id": "uuid-here",
  "token_type": "totp"
}
```

### Behavioral Biometrics

#### Submit Keystroke Data
```bash
POST /api/v1/behavioral/keystroke
Content-Type: application/json

{
  "identity_id": "uuid-here",
  "data": {
    "key_presses": [
      {
        "key": "a",
        "press_time": 0.1,
        "release_time": 0.15
      }
    ],
    "total_time": 5.2,
    "text": "Hello World"
  }
}
```

#### Analyze Behavioral Pattern
```bash
POST /api/v1/behavioral/analyze
Content-Type: application/json

{
  "identity_id": "uuid-here"
}
```

### Fraud Detection

#### Get Fraud Score
```bash
GET /api/v1/fraud/score/{id}
```

#### Report Fraud
```bash
POST /api/v1/fraud/report
Content-Type: application/json

{
  "identity_id": "uuid-here",
  "alert_type": "suspicious_activity",
  "severity": "high",
  "score": 85,
  "indicator": "multiple_failed_logins",
  "description": "Multiple failed login attempts detected from different IPs"
}
```

### Federated Learning

#### Get Model Status
```bash
GET /api/v1/federated/model/status
```

#### Submit Model Update
```bash
POST /api/v1/federated/model/update
Content-Type: application/json

{
  "client_id": "client-123",
  "model_version": "1.0.0",
  "update_data": "base64-encoded-weights",
  "sample_count": 1000,
  "accuracy": 0.92
}
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
  name: national-identity-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: national-identity-service
  template:
    metadata:
      labels:
        app: national-identity-service
    spec:
      containers:
      - name: national-identity-service
        image: national-identity-service:latest
        ports:
        - containerPort: 8086
        env:
        - name: CONFIG_PATH
          value: "/app/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: identity-config
---
apiVersion: v1
kind: Service
metadata:
  name: national-identity-service
spec:
  selector:
    app: national-identity-service
  ports:
  - port: 8086
    targetPort: 8086
  type: LoadBalancer
```

### Docker Build and Push

```bash
# Build the image
docker build -t national-identity-service:latest .

# Tag for registry
docker tag national-identity-service:latest registry.example.com/national-identity-service:v1.0.0

# Push to registry
docker push registry.example.com/national-identity-service:v1.0.0
```

## Security Considerations

### Data Protection
- All PII is encrypted at rest using AES-256
- TLS 1.3 is required for all connections
- Sensitive data is masked in logs

### Authentication
- Multi-factor authentication is enforced for sensitive operations
- Sessions are short-lived (24 hours) with automatic rotation
- Failed authentication attempts are rate-limited

### Audit Logging
- All identity changes are logged with actor information
- Access to sensitive endpoints is tracked
- Logs are retained for compliance requirements

### Privacy Compliance
- GDPR-compliant data handling
- Consent is required before data processing
- Data retention policies are enforced
- Right to erasure is supported

## Monitoring

### Prometheus Metrics

The service exposes the following metrics:

- `identity_requests_total` - Total number of identity requests
- `identity_verification_duration_seconds` - Verification operation duration
- `identity_fraud_score` - Current fraud score for identities
- `identity_behavioral_anomaly_count` - Number of behavioral anomalies detected
- `identity_sessions_active` - Number of active sessions
- `identity_federated_model_updates_total` - Total federated model updates

### Grafana Dashboards

Pre-configured dashboards are available for:
- Identity Operations Overview
- Verification Success Rates
- Fraud Detection Metrics
- Behavioral Analysis Dashboard
- System Health and Performance

### Health Endpoints

- `GET /health` - Service health check
- `GET /ready` - Readiness probe
- `GET /live` - Liveness probe
- `GET /metrics` - Prometheus metrics

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue in the repository or contact the development team.
