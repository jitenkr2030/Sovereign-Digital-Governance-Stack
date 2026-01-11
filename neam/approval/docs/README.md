# Multi-Level Approval System (MLAS)

## Overview

The Multi-Level Approval System (MLAS) is a comprehensive approval workflow engine designed for secure, audit-compliant decision-making processes. It implements quorum-based approval mechanisms with cryptographic signing, immutable audit trails, and full delegation support.

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────────┐
│                     Multi-Level Approval System                  │
├─────────────────────────────────────────────────────────────────┤
│  Frontend Layer                                                 │
│  ├── ApprovalInbox     - Pending approvals list view             │
│  ├── ApprovalDetail    - Detailed request view with context      │
│  └── ConfirmationDialog - Secure action confirmation             │
├─────────────────────────────────────────────────────────────────┤
│  API Layer                                                      │
│  ├── Approval Handler  - REST API endpoints                      │
│  ├── Approval API      - Frontend service layer                  │
│  └── Auth Middleware   - Authentication & authorization          │
├─────────────────────────────────────────────────────────────────┤
│  Workflow Layer (Temporal.io)                                   │
│  ├── ExecuteApprovalWorkflow    - Main approval workflow         │
│  ├── HumanInTheLoopWorkflow     - Human-in-the-loop approval     │
│  └── BatchApprovalWorkflow      - Batch processing               │
├─────────────────────────────────────────────────────────────────┤
│  Activity Layer                                                 │
│  ├── InitializeApprovalRequest    - Request initialization       │
│  ├── RecordApprovalAction         - Action recording             │
│  ├── SendNotifications            - Notification delivery        │
│  └── ExecuteApprovedAction        - Post-approval execution      │
├─────────────────────────────────────────────────────────────────┤
│  Security Layer                                                 │
│  ├── Immutable Ledger    - Hash-chained audit trail              │
│  ├── Crypto Signer       - RSA-PSS signatures                    │
│  └── Non-Repudiation     - Tamper-proof records                  │
├─────────────────────────────────────────────────────────────────┤
│  Data Layer                                                     │
│  ├── PostgreSQL         - Relational data                        │
│  └── Redis              - Caching & sessions                     │
└─────────────────────────────────────────────────────────────────┘
```

## Features

### 1. Quorum-Based Approvals
- **M of N Signatures**: Require minimum approvals before execution
- Configurable thresholds per approval definition
- Automatic progress tracking

### 2. Time-Bounded Windows
- Configurable approval timeouts
- Automatic expiration handling
- Reminder notifications

### 3. Delegation Chains
- Temporary delegation to other users
- Delegation scope restrictions
- Circular delegation prevention

### 4. Immutable Audit Trail
- SHA-256 cryptographic hashing
- Hash chain verification
- Non-repudiation guarantees

### 5. Notification System
- Multi-channel notifications (email, Slack, in-app)
- Configurable triggers
- Delivery tracking

## Installation

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- Temporal.io Server 1.22+

### Database Setup

```bash
# Run migration
psql -d neam_platform -f backend/models/schema.sql
```

### Environment Configuration

```yaml
# config/approval.yaml
database:
  host: localhost
  port: 5432
  name: neam_platform
  user: postgres
  password: ${DB_PASSWORD}

redis:
  host: localhost
  port: 6379

temporal:
  host: localhost
  port: 7233
  namespace: neam-production
  task_queue: approval-task-queue

security:
  private_key_path: /etc/mlas/private.key
  public_key_path: /etc/mlas/public.key
```

## Usage

### Creating an Approval Definition

```bash
curl -X POST /api/v1/approval-definitions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budget Approval",
    "description": "Approval required for budget allocations over $10,000",
    "category": "financial",
    "policy_config": {
      "quorum_required": 2,
      "total_approvers": 3,
      "timeout_hours": 48,
      "delegate_allowed": true,
      "priority": "high"
    },
    "created_by": "user-admin-123"
  }'
```

### Initiating an Approval Request

```bash
curl -X POST /api/v1/approval-requests \
  -H "Content-Type: application/json" \
  -d '{
    "definition_id": "def-uuid-here",
    "requester_id": "user-123",
    "resource_type": "budget",
    "resource_id": "budget-456",
    "context_data": {
      "amount": 50000,
      "currency": "USD",
      "description": "Q4 Marketing Budget"
    },
    "approver_ids": ["user-1", "user-2", "user-3"],
    "priority": "high"
  }'
```

### Submitting a Decision

```bash
curl -X POST /api/v1/approval-requests/{request_id}/decide \
  -H "Content-Type: application/json" \
  -d '{
    "approver_id": "user-1",
    "decision": "APPROVE",
    "comments": "Approved after budget review"
  }'
```

### Creating a Delegation

```bash
curl -X POST /api/v1/delegations \
  -H "Content-Type: application/json" \
  -d '{
    "delegator_id": "user-1",
    "delegatee_id": "user-2",
    "valid_from": "2024-01-15T00:00:00Z",
    "valid_until": "2024-01-20T00:00:00Z",
    "reason": "Out of office"
  }'
```

## API Reference

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/approval-definitions` | Create approval definition |
| GET | `/api/v1/approval-definitions` | List definitions |
| GET | `/api/v1/approval-definitions/:id` | Get definition |
| POST | `/api/v1/approval-requests` | Create approval request |
| GET | `/api/v1/approval-requests/:id` | Get request details |
| GET | `/api/v1/approvals/pending` | Get pending approvals |
| POST | `/api/v1/approval-requests/:id/decide` | Submit decision |
| GET | `/api/v1/approval-requests/:id/audit-trail` | Get audit trail |
| POST | `/api/v1/delegations` | Create delegation |
| GET | `/api/v1/delegations` | Get delegations |

## Policy Configuration

### Example Policy Configurations

#### Simple Majority (2 of 3)
```json
{
  "quorum_required": 2,
  "total_approvers": 3,
  "timeout_hours": 48,
  "delegate_allowed": true,
  "priority": "normal"
}
```

#### Unanimous Approval (3 of 3)
```json
{
  "quorum_required": 3,
  "total_approvers": 3,
  "timeout_hours": 72,
  "delegate_allowed": false,
  "priority": "critical"
}
```

#### Single Approver with Backup
```json
{
  "quorum_required": 1,
  "total_approvers": 2,
  "timeout_hours": 24,
  "delegate_allowed": true,
  "priority": "high"
}
```

## Security

### Cryptographic Signing

The system uses RSA-PSS with SHA-256 for cryptographic signing:

1. **Key Generation**
```go
keyPair, err := crypto.GenerateKeyPair(2048)
```

2. **Signing Actions**
```go
signer, _ := crypto.NewSigner(keyPair)
signature, err := signer.Sign(data)
```

3. **Verification**
```go
valid := signer.Verify(data, signature)
```

### Hash Chain Verification

Each approval action is linked to the previous action via hash chaining:

```
Action 1: Hash(Action1Data + "GENESIS")
Action 2: Hash(Action2Data + Hash(Action1Data))
Action 3: Hash(Action3Data + Hash(Action2Data))
...
```

This ensures any tampering with historical records can be detected.

## Audit & Compliance

### Generating Audit Reports

```bash
curl -X POST /api/v1/approval-requests/{id}/audit-report \
  -H "Content-Type: application/json" \
  -d '{
    "generated_by": "compliance-officer"
  }'
```

### Audit Report Contents

- Request metadata
- Complete action history
- Hash chain verification status
- Cryptographic signatures
- Compliance notes
- Summary statistics

## Testing

### Unit Tests

```bash
# Run backend tests
cd backend && go test ./... -v

# Run frontend tests
cd frontend && npm test
```

### Integration Tests

```bash
# Run with Temporal test suite
cd backend && go test ./workflows/... -tags=integration
```

## Monitoring

### Key Metrics

- Approval request throughput
- Average approval time
- Quorum achievement rate
- Expiration rate
- Delegation usage

### Health Checks

```bash
curl /health
```

## Troubleshooting

### Common Issues

1. **Workflow Not Starting**
   - Check Temporal server connectivity
   - Verify task queue configuration
   - Check definition exists and is active

2. **Actions Not Recording**
   - Verify database connectivity
   - Check for constraint violations
   - Review activity logs

3. **Hash Chain Verification Fails**
   - Check for database tampering
   - Verify key pairs match
   - Review audit logs

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests
4. Submit pull request

## License

Proprietary - NEAM Platform
