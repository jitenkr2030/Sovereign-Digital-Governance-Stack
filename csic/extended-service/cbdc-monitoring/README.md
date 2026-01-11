# CBDC Monitoring Service

Centralized monitoring and control service for Central Bank Digital Currency (CBDC) operations.

## Overview

The CBDC Monitoring Service provides real-time transaction monitoring, programmable money controls, tiered privacy levels, and policy enforcement for CBDC operations. It integrates with the existing CSIC platform infrastructure to provide comprehensive oversight of digital currency flows.

## Features

### Core Functionality
- **Real-time Transaction Monitoring**: Monitor and control CBDC transactions in real-time
- **Programmable Money Controls**: Enforce programmable money constraints and smart contract controls
- **Tiered Privacy Levels**: Implement KYC-based privacy tiers with configurable limits
- **Policy Enforcement**: Apply and enforce regulatory policies automatically
- **Velocity Checks**: Detect and prevent suspicious transaction patterns

### Integration Capabilities
- **Kafka Integration**: Event-driven architecture for real-time processing
- **PostgreSQL Storage**: Persistent storage for transactions and wallet data
- **Redis Caching**: High-performance caching for balance and velocity checks
- **Prometheus Metrics**: Built-in metrics for monitoring and alerting
- **Grafana Dashboards**: Pre-configured dashboards for visualization

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     CBDC Monitoring Service                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   HTTP API  │  │  WebSocket  │  │      Kafka Consumer     │  │
│  │  (RESTful)  │  │  (Real-time)│  │    (Event Streams)      │  │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘  │
│         │                │                      │                │
│         └────────────────┼──────────────────────┘                │
│                          ▼                                       │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Service Layer                         │    │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────────────┐   │    │
│  │  │   Wallet   │ │ Transaction│ │   Policy Engine    │   │    │
│  │  │  Service   │ │  Service   │ │                    │   │    │
│  │  └────────────┘ └────────────┘ └────────────────────┘   │    │
│  └────────────────────────┬───────────────────────────────┘    │
│                           ▼                                      │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Repository Layer                       │    │
│  │         PostgreSQL │ Redis │ Kafka │ External APIs      │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | 10.112.2.4 |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | PostgreSQL user | postgres |
| `DB_PASSWORD` | PostgreSQL password | postgres |
| `DB_NAME` | PostgreSQL database | csic_platform |
| `REDIS_HOST` | Redis host | 10.112.2.4 |
| `REDIS_PORT` | Redis port | 6379 |
| `KAFKA_BROKERS` | Kafka broker list | 10.112.2.4:9092 |
| `SERVICE_PORT` | HTTP server port | 8085 |

### Configuration File

See `config.yaml` for detailed configuration options including:
- Monitoring thresholds
- Policy enforcement rules
- Privacy tier settings
- Rate limiting
- Security settings

## API Endpoints

### Wallet Management

```
POST   /api/v1/wallets              - Create a new wallet
GET    /api/v1/wallets/:id          - Get wallet details
PUT    /api/v1/wallets/:id          - Update wallet
DELETE /api/v1/wallets/:id          - Close wallet
GET    /api/v1/wallets/:id/balance  - Get wallet balance
```

### Transaction Processing

```
POST   /api/v1/transactions              - Create transaction
GET    /api/v1/transactions/:id          - Get transaction details
GET    /api/v1/transactions              - List transactions
POST   /api/v1/transactions/:id/approve  - Approve transaction
POST   /api/v1/transactions/:id/reject   - Reject transaction
```

### Policy Management

```
POST   /api/v1/policies              - Create policy rule
GET    /api/v1/policies/:id          - Get policy rule
GET    /api/v1/policies              - List policies
PUT    /api/v1/policies/:id          - Update policy
DELETE /api/v1/policies/:id          - Disable policy
```

### Monitoring & Compliance

```
GET    /api/v1/monitoring/stats           - Get monitoring statistics
GET    /api/v1/monitoring/velocity/:id    - Get transaction velocity
GET    /api/v1/monitoring/violations      - Get violations
POST   /api/v1/compliance/check           - Compliance check
GET    /api/v1/compliance/audit/:walletId - Get audit trail
```

## Wallet Tiers

| Tier | Name | Transaction Limit | Daily Limit | KYC Required |
|------|------|-------------------|-------------|--------------|
| 1 | Basic | 1,000 CBDC | 5,000 CBDC | No |
| 2 | Standard | 10,000 CBDC | 100,000 CBDC | Yes |
| 3 | Premium | 50,000 CBDC | 500,000 CBDC | Enhanced |

## Policy Rules

The service supports various policy rule types:

- **Limit Rules**: Enforce transaction and daily limits
- **Velocity Rules**: Detect and prevent suspicious transaction patterns
- **Geographic Rules**: Restrict transactions from certain regions
- **Category Rules**: Block restricted transaction categories
- **AML Rules**: Anti-money laundering compliance

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f cbdc-monitoring

# Stop services
docker-compose down
```

### Manual Deployment

```bash
# Build the service
docker build -t csic-cbdc-monitoring .

# Run the service
docker run -d \
  --name csic-cbdc-monitoring \
  -p 8085:8085 \
  -v $(pwd)/config.yaml:/home/appuser/config.yaml:ro \
  csic-cbdc-monitoring
```

## Kafka Topics

### Consumed Topics
- `cbdc.wallets` - Wallet lifecycle events
- `cbdc.transactions` - Transaction events

### Produced Topics
- `cbdc.approvals` - Transaction approval events
- `cbdc.violations` - Policy violation events

## Monitoring

### Prometheus Metrics

The service exposes metrics at `/metrics`:

- `cbdc_wallet_total` - Total number of wallets
- `cbdc_transaction_total` - Total number of transactions
- `cbdc_transaction_amount` - Transaction amount histogram
- `cbdc_policy_violations_total` - Total policy violations
- `cbdc_latency_seconds` - Request latency histogram

### Grafana Dashboard

Pre-configured dashboard available at `grafana/provisioning/dashboards/cbdc-overview.json`

## Testing

```bash
# Run unit tests
go test ./... -v

# Run integration tests
go test ./tests/... -v -tags=integration

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Security Considerations

- All data encrypted at rest using AES-256-GCM
- mTLS required for service-to-service communication
- Rate limiting enabled by default
- Audit logging for all operations
- Compliance with financial regulations

## License

This service is part of the CSIC platform and is licensed under the same terms as the main project.
