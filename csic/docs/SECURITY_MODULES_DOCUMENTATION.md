# Security Modules - Comprehensive Documentation

This document covers the complete implementation of security modules for the CSIC Platform as part of P3 (Long-term) priority tasks.

## Overview

The CSIC Platform security architecture consists of five core modules:

1. **IAM (Identity and Access Management)** - User authentication and identity management
2. **Access Control** - Policy-based authorization and resource protection
3. **Key Management** - Cryptographic key lifecycle management
4. **Forensic Tools** - Post-incident analysis and evidence collection
5. **Incident Response** - Security incident workflow management

---

## 1. IAM Service (Identity and Access Management)

### Location
`services/security/iam/`

### Purpose
Centralized user identity management, authentication, and authorization.

### Key Features
- JWT-based authentication with access/refresh tokens
- Password management with bcrypt hashing
- Multi-factor authentication (TOTP-based)
- Session management and revocation
- Role-based access control
- API key management

### Architecture Layers

**Domain Layer**
- `User` - User entity with credentials and profile
- `Role` - Role with associated permissions
- `Session` - Authenticated session tracking
- `Permission` - Granular system permissions
- `APIKey` - Programmatic access keys

**Service Layer**
- `AuthenticationService` - Login, registration, token management
- `UserService` - User CRUD and password operations
- `RoleService` - Role and permission management
- `SessionService` - Session lifecycle management
- `APIKeyService` - API key generation and validation

**Infrastructure Layer**
- JWT token generation and validation
- Password hashing (bcrypt)
- Token blacklist for logout
- Email notification service

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User authentication |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/refresh` | Token refresh |
| POST | `/api/v1/auth/logout` | Session revocation |
| GET | `/api/v1/users/me` | Current user profile |
| PUT | `/api/v1/users/me/password` | Password change |
| POST | `/api/v1/users` | Create user (admin) |
| GET | `/api/v1/roles` | List roles |
| POST | `/api/v1/roles` | Create role |

### Default Roles

| Role | Description | Permissions |
|------|-------------|-------------|
| Admin | Full system access | All permissions |
| Analyst | Read access to compliance | View, report generation |
| Auditor | Read-only access | View audit logs only |
| User | Basic authenticated access | Limited to own data |

---

## 2. Access Control Service

### Location
`services/security/access-control/`

### Purpose
Policy-based authorization engine for resource protection.

### Key Features
- ABAC (Attribute-Based Access Control) policy engine
- Policy priority and evaluation
- Resource ownership validation
- Environment-based conditions (IP, time, device)
- Comprehensive audit logging

### Architecture Layers

**Domain Layer**
- `Policy` - Access control policy with conditions
- `AccessRequest` - Access check request
- `AccessResponse` - Access check result
- `ResourceOwnership` - Resource ownership tracking
- `AuditLog` - Access control audit trail

**Service Layer**
- `AccessControlService` - Policy evaluation and access checking
- `PolicyManagementService` - Policy CRUD operations

**Policy Conditions**
- **Subject Conditions**: Role, user, group matching
- **Resource Conditions**: Type, ID, pattern matching
- **Action Conditions**: Operation, HTTP method matching
- **Environment Conditions**: IP range, time range, device

### Access Decision Types

| Decision | Description |
|----------|-------------|
| ALLOW | Access granted |
| DENY | Access denied |
| NOT_APPLICABLE | No matching policy |
| INDETERMINATE | Evaluation error |

### Policy Evaluation Logic

1. Retrieve applicable policies
2. Sort by priority (highest first)
3. Evaluate each policy against request
4. Return first matching policy result
5. Default decision: DENY

### Example Policy Configuration

```json
{
  "name": "Admin Resource Access",
  "description": "Allow admins to access all resources",
  "effect": "ALLOW",
  "priority": 100,
  "conditions": {
    "subjects": [{
      "roles": ["Admin"]
    }],
    "resources": [{
      "types": ["*"]
    }],
    "actions": [{
      "operations": ["*"]
    }]
  }
}
```

---

## 3. Key Management Service

### Location
`services/security/key-management/`

### Purpose
Cryptographic key lifecycle management for encryption and signing.

### Key Features
- Master key generation and secure storage
- Key rotation with automated scheduling
- Data key wrapping/unwrapping
- Key versioning and archiving
- Hardware Security Module (HSM) integration support

### Key Types

| Key Type | Purpose | Rotation Period |
|----------|---------|-----------------|
| Master Key | Encrypt data keys | Annual |
| Data Key | Encrypt data | Monthly |
| Signing Key | Digital signatures | Quarterly |
| API Key | Service authentication | 90 days |

### Key Lifecycle

```
Generation -> Activation -> Rotation -> Revocation -> Destruction
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/keys/generate` | Generate new key |
| POST | `/api/v1/keys/rotate` | Rotate existing key |
| POST | `/api/v1/keys/wrap` | Wrap data key |
| POST | `/api/v1/keys/unwrap` | Unwrap data key |
| GET | `/api/v1/keys` | List active keys |
| DELETE | `/api/v1/keys/{id}` | Revoke key |

---

## 4. Forensic Tools Service

### Location
`services/security/forensic-tools/`

### Purpose
Post-incident data analysis and evidence collection utilities.

### Key Features
- Log ingestion and normalization
- Timeline reconstruction
- Artifact hashing (MD5, SHA256)
- Evidence chain of custody
- Search and correlation engine

### Components

**Log Ingestion Pipeline**
- Multi-source log collection
- Format normalization (JSON, syslog, CEF)
- Timestamp synchronization
- Field extraction and enrichment

**Timeline Analysis**
- Event correlation
- Pattern detection
- Anomaly identification
- Visualization generation

**Evidence Management**
- Cryptographic hashing
- Chain of custody tracking
- Tamper detection
- Secure storage

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/forensics/ingest` | Ingest log data |
| POST | `/api/v1/forensics/search` | Search logs |
| POST | `/api/v1/forensics/timeline` | Generate timeline |
| GET | `/api/v1/forensics/evidence/{id}` | Get evidence |
| POST | `/api/v1/forensics/hash` | Calculate hash |

---

## 5. Incident Response Service

### Location
`services/security/incident-response/`

### Purpose
Security incident workflow management and tracking.

### Key Features
- Ticket creation and management
- Automated playbook triggers
- Status state machine
- Escalation procedures
- Post-incident review

### Incident State Machine

```
┌─────────┐     ┌─────────────┐     ┌──────────────┐
│  OPEN   │ ──► │ INVESTIGATING│ ──► │  CONTAINED  │
└─────────┘     └─────────────┘     └──────────────┘
      │                                    │
      │                                    ▼
      │                              ┌──────────┐
      │                              │  ERADICATED│
      │                              └──────────┘
      │                                    │
      │                                    ▼
      │                              ┌──────────┐
      │                              │ RECOVERED│
      │                              └──────────┘
      │                                    │
      │                                    ▼
      │                              ┌──────────┐
      └─────────────────────────────►│ RESOLVED │
                                     └──────────┘
```

### Severity Levels

| Level | Response Time | Description |
|-------|--------------|-------------|
| P1 - Critical | 15 minutes | Active breach, immediate response |
| P2 - High | 1 hour | Significant impact, urgent response |
| P3 - Medium | 4 hours | Limited impact, standard response |
| P4 - Low | 24 hours | Minimal impact, routine response |

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/incidents` | Create incident |
| GET | `/api/v1/incidents` | List incidents |
| GET | `/api/v1/incidents/{id}` | Get incident details |
| PUT | `/api/v1/incidents/{id}` | Update incident |
| POST | `/api/v1/incidents/{id}/status` | Update status |
| POST | `/api/v1/incidents/{id}/escalate` | Escalate incident |
| POST | `/api/v1/incidents/{id}/playbook` | Execute playbook |

---

## Integration Architecture

### Service Communication

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                               │
└─────────────────────────────────────────────────────────────────┘
         │                  │                  │
         ▼                  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│     IAM       │  │ Access Control│  │ Incident Resp │
└───────────────┘  └───────────────┘  └───────────────┘
         │                  │                  │
         └──────────────────┼──────────────────┘
                            ▼
                  ┌───────────────────┐
                  │   PostgreSQL      │
                  │   (Shared DB)     │
                  └───────────────────┘
                            │
                            ▼
                  ┌───────────────────┐
                  │      Kafka        │
                  │ (Event Bus)       │
                  └───────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│    Logging    │  │   Reporting   │  │   Frontend    │
└───────────────┘  └───────────────┘  └───────────────┘
```

### Security Flow

1. **Authentication Flow**
   - Client → IAM Service → JWT Token Generation
   - Token stored in HTTP-only cookie or header

2. **Authorization Flow**
   - Client → Access Control → Policy Evaluation
   - Decision cached for performance

3. **Audit Flow**
   - All access decisions logged to Kafka
   - Consumer writes to audit database

---

## Deployment Configuration

### Environment Variables

| Service | Port | Key Variables |
|---------|------|---------------|
| IAM | 8083 | JWT secrets, DB credentials |
| Access Control | 8084 | Policy database, Kafka |
| Key Management | 8085 | HSM config, encryption keys |
| Forensic Tools | 8086 | Log storage, Elasticsearch |
| Incident Response | 8087 | Ticket database, notification |

### Docker Compose Setup

```yaml
version: '3.8'
services:
  iam:
    build: ./services/security/iam
    ports:
      - "8083:8083"
    environment:
      - DB_HOST=postgres
      - JWT_ACCESS_SECRET=${JWT_ACCESS_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
    depends_on:
      postgres:
        condition: service_healthy

  access-control:
    build: ./services/security/access-control
    ports:
      - "8084:8084"
    depends_on:
      postgres:
        condition: service_healthy
      kafka:
        condition: service_healthy

  key-management:
    build: ./services/security/key-management
    ports:
      - "8085:8085"
    volumes:
      - key-vault:/app/keys
    depends_on:
      postgres:
        condition: service_healthy

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: csic
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: csic_platform
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U csic -d csic_platform"]
      interval: 10s
      timeout: 5s

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
    depends_on:
      zookeeper:
        condition: service_healthy

volumes:
  postgres-data:
  key-vault:
```

---

## Security Best Practices

### Authentication Security
- Use strong password policies (min 12 chars, complexity requirements)
- Implement rate limiting on authentication endpoints
- Use secure HTTP-only cookies for tokens
- Implement account lockout after failed attempts
- Use MFA for privileged accounts

### Authorization Security
- Implement least privilege principle
- Regular access reviews and audits
- Time-bound access for sensitive resources
- Separate production and development access

### Key Management
- Use Hardware Security Modules (HSM) for master keys
- Implement automatic key rotation
- Secure key storage with encryption at rest
- Separate key usage by environment

### Audit and Monitoring
- Log all authentication and authorization events
- Implement real-time alerting for suspicious activity
- Retain audit logs according to compliance requirements
- Regular log analysis and anomaly detection

---

## Testing Strategy

### Unit Tests
- Service logic with mocked dependencies
- Policy evaluation with various conditions
- Token generation and validation

### Integration Tests
- Full authentication flow
- Database operations
- Kafka event publishing

### Security Tests
- Penetration testing
- Vulnerability scanning
- Token forgery attempts
- Privilege escalation testing

---

## Monitoring and Observability

### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| auth_login_failures | Failed login attempts | > 100 in 5 min |
| auth_session_active | Active sessions | N/A |
| policy_evaluation_time | Policy evaluation latency | > 100ms |
| key_rotation_status | Key rotation success rate | < 99% |
| incident_response_time | Time to first response | > SLA threshold |

### Health Checks
- Database connectivity
- Kafka connectivity
- Cache availability
- Service dependencies

---

## Compliance Considerations

### Data Protection
- Encrypt sensitive data at rest and in transit
- Implement data masking for non-production environments
- Secure data retention and disposal

### Access Control
- Enforce separation of duties
- Implement role-based access control
- Regular access reviews

### Audit Trail
- Immutable audit logging
- Tamper-evident storage
- Long-term retention capabilities

---

## Next Steps

### Immediate Tasks
1. Complete PostgreSQL repository implementations
2. Implement Kafka event publishing for audit logs
3. Create comprehensive unit tests (>80% coverage)
4. Set up CI/CD pipeline with security scanning

### Short-term Tasks
1. HSM integration for key management
2. MFA service implementation
3. Real-time alerting system
4. Performance optimization

### Long-term Tasks
1. Multi-factor authentication support
2. Single Sign-On (SSO) integration
3. Advanced threat detection
4. Compliance reporting automation

---

## References

- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [NIST SP 800-63 Digital Identity Guidelines](https://pages.nist.gov/800-63-3/)
- [ABAC Conceptual Framework](https://csrc.nist.gov/projects/abac)
- [GDPR Article 32 - Security of Processing](https://gdpr-info.eu/art-32-gdpr/)
