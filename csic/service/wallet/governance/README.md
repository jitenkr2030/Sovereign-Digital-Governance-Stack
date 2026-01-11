# Wallet Governance Service

A comprehensive service for managing custodial wallets, multi-signature governance, blacklists/whitelists, and wallet freeze operations for the CSIC Platform.

## Features

- **Wallet Management**: Register and manage custodial, multi-signature, exchange hot/cold wallets
- **Multi-Signature Governance**: Propose, approve, and execute transactions with multiple signers
- **Blacklist/Whitelist**: Manage sanctioned and trusted wallet addresses
- **Wallet Freeze**: Freeze and unfreeze wallets for legal or regulatory reasons
- **Compliance Checks**: Verify wallet compliance status and audit trails
- **Digital Signatures**: HSM-backed signature generation and verification
- **Asset Recovery**: Request and execute asset recovery operations

## Quick Start

```bash
# Start all services
docker-compose up -d

# Access the service
curl http://localhost:8084/health

# View metrics
curl http://localhost:9094/metrics
```

## API Endpoints

### Wallet Management
- `POST /api/v1/wallet/wallets` - Register a new wallet
- `GET /api/v1/wallet/wallets/:id` - Get wallet details
- `PUT /api/v1/wallet/wallets/:id` - Update wallet
- `DELETE /api/v1/wallet/wallets/:id` - Revoke wallet

### Multi-Signature
- `POST /api/v1/wallet/wallets/:id/signers` - Add signer
- `DELETE /api/v1/wallet/wallets/:id/signers/:signer_id` - Remove signer
- `POST /api/v1/wallet/wallets/:id/propose-transaction` - Propose transaction
- `POST /api/v1/wallet/wallets/:id/transactions/:tx_id/approve` - Approve transaction

### Compliance
- `GET /api/v1/wallet/compliance/wallets/:id` - Get compliance status
- `POST /api/v1/wallet/blacklist/addresses` - Add to blacklist
- `GET /api/v1/wallet/blacklist/addresses` - Get blacklist
- `POST /api/v1/wallet/freeze` - Freeze wallet
- `POST /api/v1/wallet/unfreeze` - Unfreeze wallet

## Configuration

Configuration is managed through `internal/config/config.yaml`:

```yaml
server:
  port: 8084

database:
  host: "localhost"
  port: 5432
  username: "csic_wallet"
  password: "wallet_password"
  name: "csic_wallet"

hsm:
  enabled: true
  provider: "soft"

governance:
  min_signers: 2
  default_threshold: 2
  transaction_expiry_hours: 24
```

## Architecture

```
internal/
├── cmd/server/           # Application entry point
├── config/               # Configuration
├── domain/models/        # Business entities
├── handler/              # HTTP handlers
├── repository/           # Data access layer
├── service/              # Business logic
│   ├── wallet_service.go
│   ├── freeze_service.go
│   ├── compliance_service.go
│   ├── governance_service.go
│   └── signature_service.go
├── hsm/                  # HSM integration
├── db/migrations/        # Database migrations
├── monitoring/           # Prometheus & Grafana
└── deploy/               # Docker configuration
```

## Database Schema

Key tables:
- `wallets` - Registered wallet records
- `wallet_signers` - Multi-signer configurations
- `transaction_proposals` - Pending transactions
- `blacklist` - Sanctioned addresses
- `whitelist` - Trusted addresses
- `wallet_freezes` - Freeze records
- `wallet_audit_logs` - Audit trail

## License

Part of CSIC Platform - All rights reserved.
