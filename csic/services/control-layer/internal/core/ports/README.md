# Control Layer Service - Ports Layer Documentation

## Overview

This document provides a comprehensive overview of the ports layer in the Control Layer Service, following Clean Architecture principles. The ports layer defines interfaces (contracts) that separate the core domain logic from external concerns such as databases, APIs, message queues, and infrastructure services.

## Architecture Principles

The ports layer is organized following these key principles:

1. **Interface Segregation**: Each port defines a focused set of related operations
2. **Dependency Inversion**: Inner layers (domain) define interfaces, outer layers (infrastructure) implement them
3. **Single Responsibility**: Each port has a clear, specific purpose
4. **Testability**: Ports enable easy mocking for unit testing

## Port Categories

### Domain-Specific Ports

#### 1. Policy Ports (`policy.go`)

The policy ports define interfaces for regulatory policy management operations.

**PolicyRepository**
- CRUD operations for policy entities
- Advanced queries (by category, scope, effective date)
- Version control and history management
- Search capabilities

**PolicyService**
- Business logic for policy lifecycle
- Compliance checking and validation
- Policy import/export functionality
- Policy cloning and rollback

**PolicyEventPublisher/Subscriber**
- Event-driven communication for policy changes
- Async processing of policy updates

**PolicyCacheRepository**
- Performance optimization through caching
- Cache invalidation strategies

#### 2. Enforcement Ports (`enforcement.go`)

The enforcement ports define interfaces for regulatory enforcement action management.

**EnforcementRepository**
- Enforcement action data persistence
- Timeline and document management
- History tracking
- Bulk operations

**EnforcementService**
- Enforcement action lifecycle management
- Evidence and hearing management
- Workflow transitions
- Escalation handling

**EnforcementNotificationService**
- Stakeholder notification management
- Template-based notifications

**EnforcementSearchRepository**
- Full-text search capabilities
- Suggestion and reindexing

#### 3. State Ports (`state.go`)

The state ports define interfaces for regulatory state management and tracking.

**StateRepository**
- State data persistence
- State history and comparison
- Alert management
- Snapshot capabilities

**StateService**
- State transition logic
- Batch operations
- Scheduled transitions
- Trend analysis

**StateCacheRepository**
- Current state caching
- Cache invalidation by type/entity

**StateAuditRepository**
- Audit trail persistence
- Compliance-ready export

#### 4. Intervention Ports (`intervention.go`)

The intervention ports define interfaces for regulatory intervention operations.

**InterventionRepository**
- Intervention data persistence
- Timeline and document management
- Action and outcome tracking
- Effectiveness metrics

**InterventionService**
- Intervention lifecycle management
- Simulation and recommendation engine
- Workflow transitions
- Clone and archive operations

**InterventionTemplateRepository**
- Reusable intervention templates
- Template versioning

#### 5. Compliance Ports (`compliance.go`)

The compliance ports define interfaces for comprehensive compliance management.

**ComplianceRepository**
- Compliance record persistence
- Violation and exemption tracking
- Historical data and snapshots
- Metrics aggregation

**ComplianceService**
- Integrated compliance checking
- Self-assessment workflows
- Scheduled checks
- Trend analysis

**ComplianceExternalService**
- External regulatory body integration
- Certification management
- External audit support

## Infrastructure Ports

### Communication Ports

#### 6. gRPC Ports (`grpc.go`)

**GRPCServer/Client**
- Server lifecycle management
- Client connection handling
- Service registration

**Domain-specific gRPC Services**
- GRPCAuthService
- GRPCPolicyService
- GRPCEnforcementService
- GRPCStateService
- GRPCInterventionService
- GRPCComplianceService

**gRPC Middleware**
- Authentication interceptor
- Rate limiting
- Logging
- Circuit breaking
- Recovery
- Metrics collection

#### 7. Kafka Ports (`kafka.go`)

**KafkaProducer/Consumer**
- Message production and consumption
- Batch operations
- Async processing

**KafkaAdminClient**
- Topic management
- Partition configuration
- Schema registry integration

**KafkaEventBus**
- Domain event publishing
- Event consumption
- Handler registration

### Data Access Ports

#### 8. Repository Ports (`repository.go`)

**Generic Repository Patterns**
- CRUD operations
- Soft delete
- Versioning
- Auditing
- Multi-tenancy
- Bulk operations
- CQRS

**Event Sourcing**
- Event persistence
- Snapshot management
- Aggregate reconstruction

#### 9. Database Ports (`database.go`)

**Database Operations**
- Connection management
- Transaction handling
- Query building

**Schema Management**
- Migration support
- Table/column management
- Index and constraint management

**Database Features**
- Replication
- Backup/Restore
- Encryption
- Connection pooling
- Query caching

### API Ports

#### 10. HTTP/REST Ports (`http.go`)

**HTTPServer**
- Route registration
- Middleware support
- Server lifecycle

**HTTP Handlers**
- Request/Response handling
- Validation
- Error handling

**HTTP Middleware**
- Authentication
- Rate limiting
- CORS
- Recovery
- Security headers

**Domain-specific Handlers**
- PolicyHTTPHandler
- EnforcementHTTPHandler
- StateHTTPHandler
- InterventionHTTPHandler
- ComplianceHTTPHandler

### Observability Ports

#### 11. Metrics Ports (`metrics.go`)

**Metrics Collection**
- Counters, Gauges, Histograms
- Summaries, Timers
- Tag management

**Alert Management**
- Alert rules and thresholds
- Notification channels
- Escalation handling

**Monitoring**
- Dashboard management
- Service monitoring
- Health checks
- Distributed tracing

### Security Ports

#### 12. Security Ports (`security.go`)

**Authentication**
- AuthenticationService
- TokenService
- SessionManager
- MFA support

**Authorization**
- Authorizer
- AuthorizationService
- Role/Permission management

**Cryptography**
- EncryptionService
- Hasher
- KeyManager
- CertificateManager

**Security Features**
- SSO/OAuth management
- Audit logging
- Intrusion detection
- Rate limiting
- WAF management

### Common Ports

#### 13. Service Ports (`service.go`)

**Service Patterns**
- GenericService
- CRUDService
- QueryService
- CommandService
- DomainService
- ApplicationService

**Advanced Patterns**
- WorkflowService
- SagaService
- EventDrivenService
- AggregateService

**Service Management**
- Factory patterns
- Middleware
- Health checking
- Metrics and logging

#### 14. Common Ports (`common.go`)

**Infrastructure Abstractions**
- UnitOfWork
- CacheRepository
- SearchRepository
- Logger
- ConfigurationLoader

**Integration Patterns**
- EventPublisher/Subscriber
- NotificationService
- FileStorage
- EmailService
- Scheduler

**Resilience Patterns**
- CircuitBreaker
- RateLimiter
- RetryService
- TimeoutService

## Usage Guidelines

### Implementing Ports

1. **Create implementations in infrastructure layer**
   ```go
   // internal/adapters/repository/policy_repository.go
   type policyRepository struct {
       db *sql.DB
   }

   func (r *policyRepository) Create(ctx context.Context, policy *domain.Policy) error {
       // Implementation
   }
   ```

2. **Register implementations in dependency injection**
   ```go
   // internal/adapters/wiring.go
   func WireControllers(container *di.Container) {
       container.Bind(ports.PolicyRepositoryType, func() interface{} {
           return repository.NewPolicyRepository(database)
       })
   }
   ```

### Testing with Ports

1. **Use mock implementations**
   ```go
   // internal/core/ports/testing.go
   type mockPolicyRepository struct {
       policies []*domain.Policy
       err error
   }

   func (m *mockPolicyRepository) Create(ctx context.Context, policy *domain.Policy) error {
       if m.err != nil {
           return m.err
       }
       m.policies = append(m.policies, policy)
       return nil
   }
   ```

2. **Inject dependencies in tests**
   ```go
   func TestPolicyService_CreatePolicy(t *testing.T) {
       repo := &mockPolicyRepository{}
       service := NewPolicyService(repo)

       policy, err := service.CreatePolicy(ctx, input)

       assert.NoError(t, err)
       assert.NotNil(t, policy)
   }
   ```

## Best Practices

1. **Keep interfaces focused**: Each port should have a single, clear purpose
2. **Favor composition**: Split large interfaces into smaller, focused ones
3. **Use consistent naming**: Follow Go conventions (e.g., `Repository`, `Service`)
4. **Document expectations**: Add comments for error conditions and side effects
5. **Version APIs**: Consider versioning for externally exposed interfaces
6. **Consider backward compatibility**: Minimize breaking changes to ports

## File Structure

```
internal/core/ports/
├── README.md                    # This file
├── policy.go                    # Policy domain ports
├── enforcement.go               # Enforcement domain ports
├── state.go                     # State domain ports
├── intervention.go              # Intervention domain ports
├── compliance.go                # Compliance domain ports
├── common.go                    # Common infrastructure ports
├── grpc.go                      # gRPC communication ports
├── kafka.go                     # Kafka event bus ports
├── repository.go                # Repository patterns
├── service.go                   # Service patterns
├── http.go                      # HTTP/REST API ports
├── database.go                  # Database operation ports
├── metrics.go                   # Metrics and monitoring ports
└── security.go                  # Security and auth ports
```

## Summary

The ports layer provides a clean separation between the core domain logic of the Control Layer Service and external systems. By defining clear interfaces for all external interactions, the architecture ensures:

- **Testability**: Easy to mock and test core logic
- **Maintainability**: Clear boundaries between concerns
- **Flexibility**: Easy to swap implementations
- **Scalability**: Clean patterns for adding new features
- **Reliability**: Well-defined contracts for error handling

This architecture follows industry best practices and enables the Control Layer Service to be robust, maintainable, and adaptable to changing requirements.
