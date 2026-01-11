# Compliance Module

A comprehensive regulatory compliance system for managing licensing workflows, compliance scoring, obligation tracking, and audit trails. This service is part of the CSIC (Crypto State Infrastructure Contractor) platform.

## Features

### 1. Licensing Workflow Engine
- Complete license application lifecycle (Draft → Submitted → Under Review → Approved/Rejected)
- License issuance, suspension, and revocation
- Automatic license number generation
- Expiration tracking and renewal reminders
- Support for multiple license types (Exchange, Custody, Mining, Trading, etc.)

### 2. Compliance Scoring System
- Automated compliance score calculation (0-100 scale)
- Tiered classification (Gold, Silver, Bronze, At-Risk, Critical)
- Score breakdown by category
- Historical score tracking and trends
- Deductions for overdue obligations and suspended licenses

### 3. Obligation Management
- Create and track regulatory obligations
- Deadline monitoring and overdue detection
- Evidence submission and fulfillment tracking
- Priority-based obligation management
- Automatic status updates for overdue items

### 4. Audit Support Tools
- Complete audit trail for all compliance activities
- Filter by entity, actor, resource, action type, and date range
- Immutable audit records with timestamps
- IP address and user agent tracking
- Support for regulatory audits and investigations

## Architecture

This service follows the **Hexagonal Architecture** (Ports & Adapters) pattern:

```
csic-platform/services/compliance/
├── cmd/
│   └── main.go                     # Application entry point
├── config/
│   └── config.yaml                 # Configuration file
├── internal/
│   ├── adapters/
│   │   ├── handler/http/           # HTTP API adapters
│   │   │   ├── handlers.go         # HTTP handlers
│   │   │   ├── router.go           # Gin router configuration
│   │   │   └── middleware/         # Request logging, CORS, Request-ID
│   │   └── repository/             # Database adapters
│   │       └── postgres/           # PostgreSQL implementation
│   │           ├── connection.go
│   │           ├── config.go
│   │           └── repository.go
│   └── core/
│       ├── domain/                 # Business entities
│       │   └── models.go
│       ├── ports/                  # Interface definitions
│       │   ├── repository.go       # Output ports
│       │   └── service.go          # Input ports
│       └── services/               # Business logic
│           ├── license_service.go
│           ├── compliance_service.go
│           ├── obligation_service.go
│           ├── audit_service.go
│           └── license_service_test.go
├── migrations/                     # Database migrations
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker (optional)

### Configuration

Copy the example configuration and adjust as needed:

```bash
cp config.yaml.example config.yaml
```

Key configuration options:

```yaml
app:
  host: "0.0.0.0"
  port: 8081

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "csic_platform"

licensing:
  default_validity_days: 365
  renewal_warning_days: 30

scoring:
  base_score: 100.0
  overdue_deduction: 10.0
  suspended_license_deduction: 50.0
```

### Running with Docker

```bash
docker-compose up -d
```

### Running Locally

1. Set up the database:

```bash
psql -U postgres -f migrations/001_initial_schema.up.sql
```

2. Run the application:

```bash
go run cmd/main.go
```

## API Documentation

### License Application Endpoints

#### Submit Application
```http
POST /api/v1/licenses/applications
Content-Type: application/json

{
  "entity_id": "uuid",
  "type": "NEW",
  "license_type": "EXCHANGE",
  "requested_terms": "Exchange license for trading",
  "documents": ["doc1.pdf"]
}
```

#### Review Application
```http
PATCH /api/v1/licenses/applications/:id/review
Content-Type: application/json

{
  "approved": true,
  "reviewer_id": "uuid",
  "notes": "All requirements met",
  "granted_terms": "Full exchange license",
  "conditions": "Standard conditions"
}
```

#### List Applications
```http
GET /api/v1/licenses/applications?status=SUBMITTED&page=1&page_size=20
```

### License Endpoints

#### Issue License
```http
POST /api/v1/licenses
Content-Type: application/json

{
  "entity_id": "uuid",
  "license_type": "EXCHANGE",
  "license_number": "LCC-2024-000001",
  "expiry_days": 365,
  "jurisdiction": "US",
  "issued_by": "Admin"
}
```

#### Suspend License
```http
POST /api/v1/licenses/:id/suspend
Content-Type: application/json

{
  "reason": "Compliance investigation in progress"
}
```

#### Revoke License
```http
POST /api/v1/licenses/:id/revoke
Content-Type: application/json

{
  "reason": "Regulatory violation"
}
```

#### Get Expiring Licenses
```http
GET /api/v1/licenses/expiring?days=30
```

### Entity Endpoints

#### Register Entity
```http
POST /api/v1/entities
Content-Type: application/json

{
  "name": "Crypto Exchange Pro",
  "legal_name": "Crypto Exchange Pro LLC",
  "registration_num": "CRYPTO-2024-001",
  "jurisdiction": "US",
  "entity_type": "EXCHANGE",
  "address": "123 Blockchain Ave, NY",
  "contact_email": "compliance@cryptoexchange.com",
  "risk_level": "MEDIUM"
}
```

### Compliance Scoring Endpoints

#### Get Compliance Score
```http
GET /api/v1/entities/:id/compliance/score
```

Response:
```json
{
  "id": "uuid",
  "entity_id": "uuid",
  "total_score": 85.5,
  "tier": "SILVER",
  "breakdown": "{\"base_score\": 100, \"overdue_deductions\": -10, ...}",
  "calculated_at": "2024-01-15T10:30:00Z"
}
```

#### Recalculate Score
```http
POST /api/v1/entities/:id/compliance/score/recalculate
```

#### Get Score History
```http
GET /api/v1/entities/:id/compliance/score/history?limit=10
```

#### Get Compliance Statistics
```http
GET /api/v1/compliance/stats
```

### Obligation Endpoints

#### Create Obligation
```http
POST /api/v1/obligations
Content-Type: application/json

{
  "entity_id": "uuid",
  "regulation_id": "uuid",
  "description": "Submit quarterly audit report",
  "due_date": "2024-03-31T23:59:59Z",
  "priority": 1,
  "evidence_refs": "report.pdf"
}
```

#### Get Entity Obligations
```http
GET /api/v1/entities/:id/obligations
```

#### Get Overdue Obligations
```http
GET /api/v1/obligations/overdue
```

#### Fulfill Obligation
```http
POST /api/v1/obligations/:id/fulfill
Content-Type: application/json

{
  "evidence": "Quarterly report submitted on time"
}
```

#### Check Overdue Obligations
```http
POST /api/v1/obligations/check-overdue
```

### Audit Endpoints

#### Get Audit Logs
```http
GET /api/v1/audit-logs?entity_id=uuid&action_type=CREATE&from=2024-01-01&to=2024-12-31
```

#### Get Entity Audit Trail
```http
GET /api/v1/entities/:id/audit-trail?limit=100&offset=0
```

## Domain Models

### License Status
- `ACTIVE`: License is valid and in force
- `SUSPENDED`: License temporarily suspended
- `REVOKED`: License permanently revoked
- `EXPIRED`: License has expired

### Application Status
- `DRAFT`: Application is being prepared
- `SUBMITTED`: Application submitted for review
- `UNDER_REVIEW`: Application under active review
- `APPROVED`: Application approved
- `REJECTED`: Application rejected
- `WITHDRAWN`: Application withdrawn by applicant

### Compliance Tier
- `GOLD`: Score ≥ 90
- `SILVER`: Score ≥ 75
- `BRONZE`: Score ≥ 60
- `AT_RISK`: Score ≥ 40
- `CRITICAL`: Score < 40

### Obligation Status
- `PENDING`: Obligation not yet fulfilled
- `IN_PROGRESS`: Obligation is being worked on
- `FULFILLED`: Obligation completed
- `OVERdue`: Obligation past due date
- `WAIVED`: Obligation waived by regulator

## Scoring Algorithm

The compliance score is calculated as follows:

```
Base Score: 100 points

Deductions:
- -10 points per overdue obligation
- -50 points per suspended license

Bonuses:
- +5 points for early fulfillment (optional)

Final Score = Base Score - Overdue Deductions - Suspended License Deductions + Bonuses
```

The score is clamped between 0 and 100.

## Testing

Run unit tests:

```bash
go test ./internal/core/services/... -v
```

Run integration tests (requires PostgreSQL):

```bash
go test ./internal/adapters/... -v
```

## Code Examples

The following examples demonstrate common usage patterns for the Compliance Module, including license management, compliance scoring, and obligation tracking.

### Basic Module Usage

The following example shows how to initialize the compliance module and perform basic operations.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/compliance/internal/service"
    "github.com/csic-platform/compliance/internal/repository"
)

func main() {
    // Initialize repository
    repo, err := repository.NewPostgresRepository(repository.Config{
        Host:     "localhost",
        Port:     5432,
        Username: "postgres",
        Password: "postgres",
        Database: "csic_platform",
    })
    if err != nil {
        log.Fatalf("Failed to create repository: %v", err)
    }
    defer repo.Close()
    
    // Initialize services
    licenseSvc := service.NewLicenseService(repo)
    complianceSvc := service.NewComplianceService(repo)
    obligationSvc := service.NewObligationService(repo)
    
    ctx := context.Background()
    
    // Check service health
    health, err := complianceSvc.HealthCheck(ctx)
    if err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    fmt.Printf("Service status: %s\n", health.Status)
}
```

### License Management

The following examples demonstrate complete license lifecycle management including application submission, review, issuance, and status changes.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/compliance/internal/models"
    "github.com/csic-platform/compliance/internal/service"
)

func licenseManagement(repo *repository.PostgresRepository) {
    ctx := context.Background()
    licenseSvc := service.NewLicenseService(repo)
    
    // Step 1: Register an entity first
    entity := &models.Entity{
        ID:           generateUUID(),
        Name:         "Crypto Exchange Pro",
        LegalName:    "Crypto Exchange Pro LLC",
        Registration: "CEP-2024-001",
        Jurisdiction: "US",
        EntityType:   models.EntityTypeExchange,
        Address:      "123 Blockchain Ave, NY",
        ContactEmail: "compliance@cryptoexchange.com",
    }
    
    if err := repo.CreateEntity(ctx, entity); err != nil {
        log.Fatalf("Failed to create entity: %v", err)
    }
    fmt.Printf("Created entity: %s (%s)\n", entity.Name, entity.ID)
    
    // Step 2: Submit license application
    application, err := licenseSvc.SubmitApplication(ctx, &service.SubmitApplicationRequest{
        EntityID:       entity.ID,
        LicenseType:    models.LicenseTypeExchange,
        RequestedTerms: "Full exchange license for cryptocurrency trading",
        Documents:      []string{"application.pdf", "business-plan.pdf"},
    })
    if err != nil {
        log.Fatalf("Failed to submit application: %v", err)
    }
    
    fmt.Printf("Application submitted: %s (status: %s)\n", 
        application.ID, application.Status)
    
    // Step 3: Review and approve application
    reviewed, err := licenseSvc.ReviewApplication(ctx, &service.ReviewApplicationRequest{
        ApplicationID: application.ID,
        ReviewerID:    "admin-001",
        Approved:      true,
        Notes:         "All requirements met, proceeding to license issuance",
        GrantedTerms:  "Full exchange license valid for 12 months",
        Conditions:    "Standard regulatory conditions apply",
    })
    if err != nil {
        log.Fatalf("Failed to review application: %v", err)
    }
    
    fmt.Printf("Application reviewed: %s\n", reviewed.Status)
    
    // Step 4: Issue the license
    license, err := licenseSvc.IssueLicense(ctx, &service.IssueLicenseRequest{
        EntityID:   entity.ID,
        LicenseType: models.LicenseTypeExchange,
        ExpiryDays: 365,
        Jurisdiction: "US",
        IssuedBy:   "admin-001",
        Conditions: []string{"KYC required", "Quarterly audits"},
    })
    if err != nil {
        log.Fatalf("Failed to issue license: %v", err)
    }
    
    fmt.Printf("License issued: %s (valid until: %s)\n", 
        license.LicenseNumber, license.ExpiryDate.Format(time.RFC3339))
    
    // Step 5: Manage license status changes
    // Suspend license for investigation
    suspended, err := licenseSvc.SuspendLicense(ctx, &service.SuspendLicenseRequest{
        LicenseID: license.ID,
        Reason:    "Compliance investigation in progress",
        SuspendedBy: "compliance-officer-001",
    })
    if err != nil {
        log.Fatalf("Failed to suspend license: %v", err)
    }
    fmt.Printf("License suspended: %s\n", suspended.Status)
    
    // Reactivate after investigation
    reactivated, err := licenseSvc.UpdateLicenseStatus(ctx, license.ID, 
        models.StatusActive, "Investigation closed, no violations found")
    if err != nil {
        log.Fatalf("Failed to reactivate license: %v", err)
    }
    fmt.Printf("License reactivated: %s\n", reactivated.Status)
    
    // Revoke license (permanent action)
    revoked, err := licenseSvc.RevokeLicense(ctx, &service.RevokeLicenseRequest{
        LicenseID: license.ID,
        Reason:    "Severe regulatory violations",
        RevokedBy: "senior-officer-001",
    })
    if err != nil {
        log.Fatalf("Failed to revoke license: %v", err)
    }
    fmt.Printf("License revoked: %s\n", revoked.Status)
}
```

### Compliance Scoring

The following example demonstrates how to calculate and track compliance scores for regulated entities.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/csic-platform/compliance/internal/models"
    "github.com/csic-platform/compliance/internal/service"
)

func complianceScoring(repo *repository.PostgresRepository) {
    ctx := context.Background()
    complianceSvc := service.NewComplianceService(repo)
    
    entityID := "entity-uuid-here"
    
    // Get current compliance score
    score, err := complianceSvc.GetComplianceScore(ctx, entityID)
    if err != nil {
        log.Fatalf("Failed to get compliance score: %v", err)
    }
    
    fmt.Printf("Entity: %s\n", entityID)
    fmt.Printf("Overall Score: %.1f/100\n", score.TotalScore)
    fmt.Printf("Tier: %s\n", score.Tier)
    fmt.Printf("Last Calculated: %s\n", score.CalculatedAt.Format("2006-01-02 15:04"))
    
    // Print score breakdown
    fmt.Println("\nScore Breakdown:")
    fmt.Printf("  Base Score: %.1f\n", score.BaseScore)
    fmt.Printf("  Deductions: %.1f\n", score.Deductions)
    fmt.Printf("  Bonuses: %.1f\n", score.Bonuses)
    
    if len(score.CategoryScores) > 0 {
        fmt.Println("\nCategory Scores:")
        for category, catScore := range score.CategoryScores {
            fmt.Printf("  %s: %.1f/100\n", category, catScore)
        }
    }
    
    // Recalculate score (useful after significant changes)
    newScore, err := complianceSvc.RecalculateScore(ctx, entityID)
    if err != nil {
        log.Fatalf("Failed to recalculate score: %v", err)
    }
    fmt.Printf("\nRecalculated Score: %.1f\n", newScore.TotalScore)
    
    // Get score history
    history, err := complianceSvc.GetScoreHistory(ctx, entityID, 12) // Last 12 months
    if err != nil {
        log.Fatalf("Failed to get score history: %v", err)
    }
    
    fmt.Println("\nScore History:")
    for _, entry := range history {
        fmt.Printf("  %s: %.1f (%s)\n", 
            entry.CalculatedAt.Format("2006-01-02"),
            entry.Score,
            entry.Tier)
    }
    
    // Get compliance statistics for all entities
    stats, err := complianceSvc.GetStatistics(ctx)
    if err != nil {
        log.Fatalf("Failed to get statistics: %v", err)
    }
    
    fmt.Printf("\nOverall Statistics:")
    fmt.Printf("  Total Entities: %d\n", stats.TotalEntities)
    fmt.Printf("  Average Score: %.1f\n", stats.AverageScore)
    fmt.Printf("  Gold Tier: %d\n", stats.GoldCount)
    fmt.Printf("  Silver Tier: %d\n", stats.SilverCount)
    fmt.Printf("  Bronze Tier: %d\n", stats.BronzeCount)
    fmt.Printf("  At Risk: %d\n", stats.AtRiskCount)
    fmt.Printf("  Critical: %d\n", stats.CriticalCount)
}
```

### Obligation Management

The following example shows how to create, track, and manage regulatory obligations for entities.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/compliance/internal/models"
    "github.com/csic-platform/compliance/internal/service"
)

func obligationManagement(repo *repository.PostgresRepository) {
    ctx := context.Background()
    obligationSvc := service.NewObligationService(repo)
    
    entityID := "entity-uuid-here"
    
    // Create a new obligation
    obligation, err := obligationSvc.CreateObligation(ctx, &service.CreateObligationRequest{
        EntityID:    entityID,
        Description: "Submit quarterly compliance audit report",
        Regulation:  "FIN-2024-Q1",
        DueDate:     time.Now().AddDate(0, 3, 0), // 3 months from now
        Priority:    models.PriorityHigh,
        EvidenceReq: []string{"audit-report.pdf", "financial-statement.pdf"},
    })
    if err != nil {
        log.Fatalf("Failed to create obligation: %v", err)
    }
    
    fmt.Printf("Created obligation: %s\n", obligation.ID)
    fmt.Printf("  Description: %s\n", obligation.Description)
    fmt.Printf("  Due Date: %s\n", obligation.DueDate.Format("2006-01-02"))
    fmt.Printf("  Status: %s\n", obligation.Status)
    
    // Create another obligation
    obligation2, err := obligationSvc.CreateObligation(ctx, &service.CreateObligationRequest{
        EntityID:    entityID,
        Description: "Annual KYC review completion",
        Regulation:  "KYC-2024",
        DueDate:     time.Now().AddDate(0, 6, 0),
        Priority:    models.PriorityMedium,
    })
    if err != nil {
        log.Fatalf("Failed to create obligation: %v", err)
    }
    
    // Get all obligations for entity
    obligations, err := obligationSvc.GetEntityObligations(ctx, entityID, nil)
    if err != nil {
        log.Fatalf("Failed to get obligations: %v", err)
    }
    
    fmt.Printf("\nEntity Obligations (%d):\n", len(obligations))
    for _, obs := range obligations {
        status := "PENDING"
        if obs.Status == models.ObligationOverdue {
            status = "OVERDUE!"
        }
        fmt.Printf("  - %s: %s (%s)\n", 
            obs.ID[:8], obs.Description[:30]+"...", status)
    }
    
    // Get overdue obligations across all entities
    overdue, err := obligationSvc.GetOverdueObligations(ctx)
    if err != nil {
        log.Fatalf("Failed to get overdue obligations: %v", err)
    }
    
    fmt.Printf("\nOverdue Obligations (%d):\n", len(overdue))
    for _, obs := range overdue {
        fmt.Printf("  - Entity %s: %s (due: %s)\n",
            obs.EntityID[:8], obs.Description[:30]+"...", 
            obs.DueDate.Format("2006-01-02"))
    }
    
    // Fulfill an obligation
    fulfillment, err := obligationSvc.FulfillObligation(ctx, &service.FulfillObligationRequest{
        ObligationID: obligation.ID,
        Evidence:     "Quarterly audit report submitted",
        SubmittedBy:  "compliance-officer-001",
    })
    if err != nil {
        log.Fatalf("Failed to fulfill obligation: %v", err)
    }
    fmt.Printf("\nObligation fulfilled: %s\n", fulfillment.Status)
    
    // Check for newly overdue obligations
    newlyOverdue, err := obligationSvc.CheckOverdue(ctx)
    if err != nil {
        log.Fatalf("Failed to check overdue: %v", err)
    }
    fmt.Printf("\nNewly overdue obligations: %d\n", len(newlyOverdue))
}
```

### Audit Trail Querying

The following example demonstrates how to query audit logs for compliance and forensic purposes.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/compliance/internal/service"
)

func auditTrailQuerying(repo *repository.PostgresRepository) {
    ctx := context.Background()
    auditSvc := service.NewAuditService(repo)
    
    entityID := "entity-uuid-here"
    
    // Get complete audit trail for an entity
    auditTrail, err := auditSvc.GetEntityAuditTrail(ctx, entityID, &service.AuditQuery{
        Limit:  100,
        Offset: 0,
    })
    if err != nil {
        log.Fatalf("Failed to get audit trail: %v", err)
    }
    
    fmt.Printf("Audit Trail for Entity %s:\n", entityID[:8])
    fmt.Printf("Total Entries: %d\n\n", auditTrail.TotalCount)
    
    for _, entry := range auditTrail.Entries {
        fmt.Printf("[%s] %s: %s - %s (%s)\n",
            entry.Timestamp.Format("2006-01-02 15:04:05"),
            entry.ActorID[:8],
            entry.Action,
            entry.ResourceType,
            entry.Result)
    }
    
    // Query with filters
    filteredQuery, err := auditSvc.QueryAuditLogs(ctx, &service.AuditQuery{
        StartTime:   time.Now().AddDate(0, -1, 0), // Last month
        EndTime:     time.Now(),
        ActionTypes: []string{"LICENSE_UPDATE", "STATUS_CHANGE"},
        Limit:       50,
    })
    if err != nil {
        log.Fatalf("Failed to query audit logs: %v", err)
    }
    
    fmt.Printf("\nFiltered Audit Logs (%d):\n", filteredQuery.TotalCount)
    for _, entry := range filteredQuery.Entries {
        fmt.Printf("  %s: %s - %s\n",
            entry.Timestamp.Format("2006-01-02"),
            entry.Action,
            entry.ResourceType)
    }
    
    // Export audit trail for regulatory filing
    export, err := auditSvc.ExportAuditTrail(ctx, &service.ExportRequest{
        EntityID:  entityID,
        StartTime: time.Now().AddDate(0, -3, 0),
        EndTime:   time.Now(),
        Format:    "json",
    })
    if err != nil {
        log.Fatalf("Failed to export audit trail: %v", err)
    }
    fmt.Printf("\nAudit trail exported: %d bytes\n", len(export.Data))
}
```

### Batch Operations

For managing multiple entities efficiently, the module supports batch operations.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/compliance/internal/service"
)

func batchOperations(repo *repository.PostgresRepository) {
    ctx := context.Background()
    complianceSvc := service.NewComplianceService(repo)
    
    // Calculate scores for all entities
    entityIDs := []string{
        "entity-1-uuid",
        "entity-2-uuid",
        "entity-3-uuid",
    }
    
    scores, err := complianceSvc.CalculateBatchScores(ctx, entityIDs)
    if err != nil {
        log.Fatalf("Failed to calculate batch scores: %v", err)
    }
    
    fmt.Println("Batch Score Results:")
    for i, score := range scores {
        if score.Error != nil {
            fmt.Printf("  %s: ERROR - %v\n", entityIDs[i], score.Error)
        } else {
            fmt.Printf("  %s: %.1f (%s)\n", 
                entityIDs[i][:8], score.Score, score.Tier)
        }
    }
    
    // Check expiring licenses
    expiringIn30Days := time.Now().AddDate(0, 1, 0)
    expiringLicenses, err := complianceSvc.GetExpiringLicenses(ctx, expiringIn30Days)
    if err != nil {
        log.Fatalf("Failed to get expiring licenses: %v", err)
    }
    
    fmt.Printf("\nExpiring Licenses (%d):\n", len(expiringLicenses))
    for _, lic := range expiringLicenses {
        fmt.Printf("  %s (%s) - Expires: %s\n",
            lic.EntityName, lic.LicenseNumber, 
            lic.ExpiryDate.Format("2006-01-02"))
    }
    
    // Send renewal reminders
    remindersSent, err := complianceSvc.SendRenewalReminders(ctx, 30) // 30 days notice
    if err != nil {
        log.Fatalf("Failed to send renewal reminders: %v", err)
    }
    fmt.Printf("\nRenewal reminders sent: %d\n", remindersSent)
}
```

## Monitoring

Health check endpoint:

```http
GET /health
```

Response:

```json
{
  "status": "healthy",
  "service": "compliance-module"
}
```

## License

Part of the CSIC Platform - All rights reserved.
