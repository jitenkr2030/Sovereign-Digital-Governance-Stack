# Mining Control Platform

A regulatory compliance system for registering mining operations, monitoring real-time hashrate, enforcing quotas, and issuing remote shutdown commands. This service is part of the CSIC (Crypto State Infrastructure Contractor) platform.

## Features

- **National Mining Registry**: Register and manage mining operations with wallet address tracking
- **Hashrate Monitoring**: Real-time telemetry ingestion with automatic unit normalization (TH/s, PH/s, GH/s)
- **Quota Enforcement**: Dynamic and fixed quotas with automatic violation detection
- **Remote Shutdown**: Graceful, immediate, and force-kill shutdown command capabilities
- **Compliance Reporting**: Track violations, pending commands, and registry statistics

## Architecture

This service follows the **Hexagonal Architecture** (Ports & Adapters) pattern:

```
csic-platform/services/mining/
├── cmd/
│   └── main.go                 # Application entry point
├── config/
│   └── config.yaml             # Configuration file
├── internal/
│   ├── adapters/
│   │   ├── handler/http/       # HTTP API adapters
│   │   │   ├── handlers.go     # HTTP handlers
│   │   │   ├── router.go       # Gin router configuration
│   │   │   └── middleware/     # Request logging, CORS, etc.
│   │   └── repository/         # Database adapters
│   │       └── postgres/       # PostgreSQL implementation
│   │           ├── connection.go
│   │           ├── config.go
│   │           └── repository.go
│   └── core/
│       ├── domain/             # Business entities
│       │   └── models.go
│       ├── ports/              # Interface definitions
│       │   ├── repository.go   # Output ports
│       │   └── service.go      # Input ports
│       └── services/           # Business logic
│           ├── mining_service.go
│           └── mining_service_test.go
├── migrations/                 # Database migrations
├── Dockerfile
├── docker-compose.yml
└── go.mod
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
  port: 8080

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "csic_platform"

compliance:
  violation_threshold: 3  # Consecutive violations before auto-shutdown
  grace_period_seconds: 300
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

### Registry Management

#### Register Operation
```http
POST /api/v1/operations
Content-Type: application/json

{
  "operator_name": "Mining Corp Inc",
  "wallet_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
  "location": "Data Center A, Nevada",
  "region": "US-WEST",
  "machine_type": "Antminer S19 Pro",
  "initial_hashrate": 100.0
}
```

Response:
```json
{
  "message": "Operation registered successfully",
  "operation": {
    "id": "uuid",
    "status": "ACTIVE",
    ...
  }
}
```

#### Get Operation
```http
GET /api/v1/operations/:id
```

#### List Operations
```http
GET /api/v1/operations?status=ACTIVE&page=1&page_size=20
```

### Hashrate Monitoring

#### Report Telemetry
```http
POST /api/v1/operations/:id/telemetry
Content-Type: application/json

{
  "hashrate": 150.5,
  "unit": "TH/s",
  "block_height": 1000000,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Get Metrics History
```http
GET /api/v1/operations/:id/metrics?limit=1000
```

### Quota Management

#### Assign Quota
```http
POST /api/v1/quotas
Content-Type: application/json

{
  "operation_id": "uuid",
  "max_hashrate": 200.0,
  "quota_type": "FIXED",
  "valid_from": "2024-01-15T00:00:00Z",
  "valid_to": "2024-12-31T23:59:59Z",
  "region": "US-WEST",
  "priority": 1
}
```

#### Get Current Quota
```http
GET /api/v1/operations/:id/quota
```

### Remote Shutdown

#### Issue Shutdown Command
```http
POST /api/v1/operations/:id/shutdown
Content-Type: application/json

{
  "command_type": "GRACEFUL",
  "reason": "Grid emergency - load balancing",
  "issued_by": "Grid Operator",
  "expires_in": 300
}
```

#### Acknowledge Command (for mining client)
```http
POST /api/v1/commands/:id/ack
```

#### Confirm Shutdown Execution
```http
POST /api/v1/commands/:id/confirm
Content-Type: application/json

{
  "success": true,
  "result": "Graceful shutdown completed successfully"
}
```

#### Get Pending Commands (for mining client polling)
```http
GET /api/v1/operations/:id/commands
```

### Statistics

#### Get Registry Statistics
```http
GET /api/v1/stats
```

Response:
```json
{
  "total_operations": 150,
  "active_operations": 120,
  "suspended_operations": 20,
  "shutdown_operations": 10,
  "total_hashrate": 15000.5,
  "quota_violations": 5,
  "pending_commands": 2
}
```

## Domain Models

### Operation Status
- `ACTIVE`: Operation is running normally
- `SUSPENDED`: Operation is temporarily suspended
- `SHUTDOWN_ORDERED`: Shutdown command issued
- `NON_COMPLIANT`: Quota violations detected
- `SHUTDOWN_EXECUTED`: Shutdown completed
- `PENDING_REGISTRATION`: Awaiting registration approval

### Command Types
- `GRACEFUL`: Graceful shutdown with timeout
- `IMMEDIATE`: Immediate shutdown
- `FORCE_KILL`: Force kill without cleanup

### Quota Types
- `FIXED`: Fixed maximum hashrate
- `DYNAMIC_GRID`: Grid-dependent dynamic quotas

## Compliance Rules

### Quota Enforcement
When a mining operation reports hashrate exceeding their quota:

1. Log the violation
2. Increment violation counter on the operation
3. After 3 consecutive violations (configurable):
   - Create automatic shutdown command
   - Mark operation as `NON_COMPLIANT`
   - Issue `GRACEFUL` shutdown command

### Shutdown Workflow
1. **ISSUED**: Command created, operation notified
2. **ACKNOWLEDGED**: Mining client acknowledged the command
3. **EXECUTED**: Mining client confirmed shutdown completion
4. **FAILED**: Shutdown failed (operation remains in violation)

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

The following examples demonstrate common usage patterns for the Mining Control Service, including operation registration, telemetry monitoring, quota management, and remote shutdown commands.

### Basic Service Usage

The following example shows how to initialize the mining control service and perform basic operations.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/mining/internal/service"
    "github.com/csic-platform/mining/internal/repository"
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
    miningSvc := service.NewMiningService(repo)
    quotaSvc := service.NewQuotaService(repo)
    controlSvc := service.NewControlService(repo)
    
    ctx := context.Background()
    
    // Check service health
    health, err := miningSvc.HealthCheck(ctx)
    if err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    fmt.Printf("Service status: %s\n", health.Status)
}
```

### Operation Registration

The following examples demonstrate registering mining operations and managing their lifecycle.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/csic-platform/mining/internal/models"
    "github.com/csic-platform/mining/internal/service"
)

func operationManagement(repo *repository.PostgresRepository) {
    ctx := context.Background()
    miningSvc := service.NewMiningService(repo)
    
    // Register a new mining operation
    operation, err := miningSvc.RegisterOperation(ctx, &service.RegisterOperationRequest{
        OperatorName:   "Northern Mining Corp",
        WalletAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b",
        Location:       "Data Center A, Nevada",
        Region:         "US-WEST",
        MachineType:    "Antminer S19 Pro",
        MachineCount:   100,
        InitialHashrate: 100.0, // TH/s per machine
        EnergySource:   "Hydroelectric",
    })
    if err != nil {
        log.Fatalf("Failed to register operation: %v", err)
    }
    
    fmt.Printf("Operation registered: %s\n", operation.ID)
    fmt.Printf("  Name: %s\n", operation.OperatorName)
    fmt.Printf("  Location: %s\n", operation.Location)
    fmt.Printf("  Initial Hashrate: %.1f TH/s\n", operation.CurrentHashrate)
    fmt.Printf("  Status: %s\n", operation.Status)
    
    // Get operation details
    opDetails, err := miningSvc.GetOperation(ctx, operation.ID)
    if err != nil {
        log.Fatalf("Failed to get operation: %v", err)
    }
    
    fmt.Printf("\nOperation Details:\n")
    fmt.Printf("  Wallet: %s\n", opDetails.WalletAddress)
    fmt.Printf("  Machines: %d x %s\n", opDetails.MachineCount, opDetails.MachineType)
    fmt.Printf("  Energy Source: %s\n", opDetails.EnergySource)
    
    // List all operations with filtering
    operations, err := miningSvc.ListOperations(ctx, &service.OperationFilter{
        Status:   models.StatusActive,
        Region:   "US-WEST",
        Page:     1,
        PageSize: 20,
    })
    if err != nil {
        log.Fatalf("Failed to list operations: %v", err)
    }
    
    fmt.Printf("\nActive Operations in US-WEST (%d):\n", len(operations))
    for _, op := range operations {
        fmt.Printf("  - %s: %.1f TH/s\n", op.OperatorName, op.CurrentHashrate)
    }
    
    // Update operation status
    updated, err := miningSvc.UpdateOperation(ctx, operation.ID, &service.UpdateOperationRequest{
        Status: models.StatusSuspended,
        Reason: "Scheduled maintenance",
    })
    if err != nil {
        log.Fatalf("Failed to update operation: %v", err)
    }
    fmt.Printf("\nOperation updated: %s\n", updated.Status)
}
```

### Telemetry Monitoring

The following examples demonstrate how to report and query mining telemetry data.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/mining/internal/service"
)

func telemetryMonitoring(repo *repository.PostgresRepository) {
    ctx := context.Background()
    miningSvc := service.NewMiningService(repo)
    
    operationID := "operation-uuid-here"
    
    // Report telemetry from mining hardware
    telemetry, err := miningSvc.ReportTelemetry(ctx, &service.ReportTelemetryRequest{
        OperationID: operationID,
        Hashrate:    150.5,
        Unit:        "TH/s",
        BlockHeight: 820000,
        Shares:      10000,
        RejectedShares: 5,
        Temperature: 65.0,
        PowerConsumption: 3200.0, // Watts
    })
    if err != nil {
        log.Fatalf("Failed to report telemetry: %v", err)
    }
    
    fmt.Printf("Telemetry recorded: %s\n", telemetry.ID)
    fmt.Printf("  Hashrate: %.1f %s\n", telemetry.Hashrate, telemetry.Unit)
    fmt.Printf("  Rejection Rate: %.2f%%\n", telemetry.RejectionRate)
    fmt.Printf("  Temperature: %.1f°C\n", telemetry.Temperature)
    fmt.Printf("  Power: %.0fW\n", telemetry.PowerConsumption)
    
    // Get current operation metrics
    metrics, err := miningSvc.GetCurrentMetrics(ctx, operationID)
    if err != nil {
        log.Fatalf("Failed to get current metrics: %v", err)
    }
    
    fmt.Printf("\nCurrent Metrics:\n")
    fmt.Printf("  Hashrate: %.1f TH/s\n", metrics.CurrentHashrate)
    fmt.Printf("  24h Average: %.1f TH/s\n", metrics.AverageHashrate24h)
    fmt.Printf("  Uptime: %.2f%%\n", metrics.UptimePercentage)
    fmt.Printf("  Total Mined: %.4f BTC\n", metrics.TotalMined)
    
    // Get historical metrics
    startTime := time.Now().AddDate(0, 0, -7) // Last 7 days
    history, err := miningSvc.GetMetricsHistory(ctx, operationID, startTime, time.Now(), 1000)
    if err != nil {
        log.Fatalf("Failed to get metrics history: %v", err)
    }
    
    fmt.Printf("\nMetrics History (%d points):\n", len(history))
    for _, m := range history {
        fmt.Printf("  %s: %.1f TH/s\n", 
            m.Timestamp.Format("2006-01-02 15:04"), m.Hashrate)
    }
    
    // Aggregate statistics
    stats, err := miningSvc.GetAggregatedStats(ctx, "US-WEST", "24h")
    if err != nil {
        log.Fatalf("Failed to get aggregated stats: %v", err)
    }
    
    fmt.Printf("\nRegional Statistics (US-WEST, 24h):\n")
    fmt.Printf("  Total Operations: %d\n", stats.TotalOperations)
    fmt.Printf("  Active Operations: %d\n", stats.ActiveOperations)
    fmt.Printf("  Total Hashrate: %.1f PH/s\n", stats.TotalHashrate/1000)
    fmt.Printf("  Total Power: %.2f MW\n", stats.TotalPower/1000000)
    fmt.Printf("  Average Efficiency: %.1f J/TH\n", stats.AverageEfficiency)
}
```

### Quota Management

The following examples demonstrate how to assign and manage energy quotas for mining operations.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/mining/internal/models"
    "github.com/csic-platform/mining/internal/service"
)

func quotaManagement(repo *repository.PostgresRepository) {
    ctx := context.Background()
    quotaSvc := service.NewQuotaService(repo)
    
    operationID := "operation-uuid-here"
    
    // Assign a fixed quota
    quota, err := quotaSvc.AssignQuota(ctx, &service.AssignQuotaRequest{
        OperationID: operationID,
        QuotaType:   models.QuotaTypeFixed,
        MaxHashrate: 200.0, // TH/s
        ValidFrom:   time.Now(),
        ValidTo:     time.Now().AddDate(1, 0, 0), // 1 year
        Region:      "US-WEST",
        Priority:    1,
    })
    if err != nil {
        log.Fatalf("Failed to assign quota: %v", err)
    }
    
    fmt.Printf("Quota assigned: %s\n", quota.ID)
    fmt.Printf("  Type: %s\n", quota.QuotaType)
    fmt.Printf("  Max Hashrate: %.1f TH/s\n", quota.MaxHashrate)
    fmt.Printf("  Valid: %s to %s\n", 
        quota.ValidFrom.Format("2006-01-02"),
        quota.ValidTo.Format("2006-01-02"))
    
    // Assign a dynamic grid-dependent quota
    dynamicQuota, err := quotaSvc.AssignQuota(ctx, &service.AssignQuotaRequest{
        OperationID: operationID,
        QuotaType:   models.QuotaTypeDynamicGrid,
        MaxHashrate: 150.0,
        GridCapacity: 5000.0, // Total grid capacity in TH/s
        ValidFrom:   time.Now(),
        ValidTo:     time.Now().AddDate(0, 1, 0),
        Region:      "US-WEST",
    })
    if err != nil {
        log.Fatalf("Failed to assign dynamic quota: %v", err)
    }
    
    fmt.Printf("\nDynamic Quota assigned: %s\n", dynamicQuota.ID)
    
    // Get current quota for operation
    currentQuota, err := quotaSvc.GetCurrentQuota(ctx, operationID)
    if err != nil {
        log.Fatalf("Failed to get current quota: %v", err)
    }
    
    fmt.Printf("\nCurrent Quota:\n")
    fmt.Printf("  Type: %s\n", currentQuota.QuotaType)
    fmt.Printf("  Max Hashrate: %.1f TH/s\n", currentQuota.MaxHashrate)
    fmt.Printf("  Allocated: %.1f TH/s\n", currentQuota.AllocatedHashrate)
    fmt.Printf("  Available: %.1f TH/s\n", currentQuota.AvailableHashrate)
    
    // Check if operation is within quota
    check, err := quotaSvc.CheckQuotaCompliance(ctx, operationID, 180.0) // 180 TH/s requested
    if err != nil {
        log.Fatalf("Failed to check quota compliance: %v", err)
    }
    
    fmt.Printf("\nQuota Compliance Check:\n")
    fmt.Printf("  Requested: %.1f TH/s\n", check.Requested)
    fmt.Printf("  Allowed: %.1f TH/s\n", check.Allowed)
    fmt.Printf("  Within Quota: %v\n", check.WithinQuota)
    if !check.WithinQuota {
        fmt.Printf("  Over by: %.1f TH/s\n", check.OverAmount)
    }
    
    // List all quotas for region
    regionQuotas, err := quotaSvc.ListRegionQuotas(ctx, "US-WEST")
    if err != nil {
        log.Fatalf("Failed to list region quotas: %v", err)
    }
    
    fmt.Printf("\nRegion Quotas (US-WEST):\n")
    for _, q := range regionQuotas {
        fmt.Printf("  %s: %.1f/%.1f TH/s (%s)\n",
            q.OperationID[:8], q.AllocatedHashrate, q.MaxHashrate, q.Status)
    }
}
```

### Remote Shutdown Commands

The following examples demonstrate how to issue and manage remote shutdown commands for mining operations.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/mining/internal/models"
    "github.com/csic-platform/mining/internal/service"
)

func shutdownCommands(repo *repository.PostgresRepository) {
    ctx := context.Background()
    controlSvc := service.NewControlService(repo)
    
    operationID := "operation-uuid-here"
    
    // Issue a graceful shutdown command
    shutdown, err := controlSvc.IssueShutdownCommand(ctx, &service.IssueShutdownRequest{
        OperationID: operationID,
        CommandType: models.CommandTypeGraceful,
        Reason:      "Grid emergency - load balancing required",
        IssuedBy:    "grid-operator-001",
        ExpiresIn:   300, // 5 minutes to comply
    })
    if err != nil {
        log.Fatalf("Failed to issue shutdown command: %v", err)
    }
    
    fmt.Printf("Shutdown command issued: %s\n", shutdown.ID)
    fmt.Printf("  Type: %s\n", shutdown.CommandType)
    fmt.Printf("  Reason: %s\n", shutdown.Reason)
    fmt.Printf("  Expires: %s\n", shutdown.ExpiresAt.Format("2006-01-02 15:04:05"))
    fmt.Printf("  Status: %s\n", shutdown.Status)
    
    // Poll for pending commands (for mining client)
    pendingCommands, err := controlSvc.GetPendingCommands(ctx, operationID)
    if err != nil {
        log.Fatalf("Failed to get pending commands: %v", err)
    }
    
    fmt.Printf("\nPending Commands (%d):\n", len(pendingCommands))
    for _, cmd := range pendingCommands {
        fmt.Printf("  - %s: %s (%s)\n", 
            cmd.ID, cmd.CommandType, cmd.Status)
    }
    
    // Simulate mining client acknowledging command
    acknowledged, err := controlSvc.AcknowledgeCommand(ctx, shutdown.ID, &service.AcknowledgeRequest{
        OperationID: operationID,
        AcknowledgedBy: "mining-client-001",
    })
    if err != nil {
        log.Fatalf("Failed to acknowledge command: %v", err)
    }
    fmt.Printf("\nCommand acknowledged: %s\n", acknowledged.Status)
    
    // Simulate mining client confirming shutdown completion
    confirmed, err := controlSvc.ConfirmShutdown(ctx, shutdown.ID, &service.ConfirmRequest{
        OperationID: operationID,
        Success:     true,
        Result:      "Graceful shutdown completed successfully",
    })
    if err != nil {
        log.Fatalf("Failed to confirm shutdown: %v", err)
    }
    fmt.Printf("Shutdown confirmed: %s\n", confirmed.Status)
    
    // Issue immediate shutdown (no grace period)
    immediate, err := controlSvc.IssueShutdownCommand(ctx, &service.IssueShutdownRequest{
        OperationID: operationID,
        CommandType: models.CommandTypeImmediate,
        Reason:      "Regulatory order - immediate cessation required",
        IssuedBy:    "regulator-001",
    })
    if err != nil {
        log.Fatalf("Failed to issue immediate shutdown: %v", err)
    }
    fmt.Printf("\nImmediate shutdown issued: %s\n", immediate.ID)
    
    // List all commands for an operation
    allCommands, err := controlSvc.ListCommands(ctx, operationID, &service.CommandFilter{
        Limit: 50,
    })
    if err != nil {
        log.Fatalf("Failed to list commands: %v", err)
    }
    
    fmt.Printf("\nCommand History (%d):\n", len(allCommands))
    for _, cmd := range allCommands {
        fmt.Printf("  [%s] %s: %s (%s)\n",
            cmd.CreatedAt.Format("2006-01-02 15:04"),
            cmd.CommandType,
            cmd.Reason[:30]+"...",
            cmd.Status)
    }
}
```

### Compliance and Statistics

The following examples demonstrate how to track compliance and generate statistics.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/mining/internal/models"
    "github.com/csic-platform/mining/internal/service"
)

func complianceAndStats(repo *repository.PostgresRepository) {
    ctx := context.Background()
    miningSvc := service.NewMiningService(repo)
    
    // Check compliance for an operation
    compliance, err := miningSvc.CheckCompliance(ctx, "operation-uuid-here")
    if err != nil {
        log.Fatalf("Failed to check compliance: %v", err)
    }
    
    fmt.Printf("Compliance Check:\n")
    fmt.Printf("  Operation: %s\n", compliance.OperationID[:8])
    fmt.Printf("  Status: %s\n", compliance.Status)
    fmt.Printf("  Violations: %d\n", compliance.ViolationCount)
    fmt.Printf("  Quota Compliance: %v\n", compliance.WithinQuota)
    fmt.Printf("  Last Check: %s\n", compliance.LastCheck.Format("2006-01-02 15:04"))
    
    // Get registry statistics
    stats, err := miningSvc.GetRegistryStats(ctx)
    if err != nil {
        log.Fatalf("Failed to get registry stats: %v", err)
    }
    
    fmt.Printf("\nRegistry Statistics:\n")
    fmt.Printf("  Total Operations: %d\n", stats.TotalOperations)
    fmt.Printf("  Active Operations: %d\n", stats.ActiveOperations)
    fmt.Printf("  Suspended Operations: %d\n", stats.SuspendedOperations)
    fmt.Printf("  Shutdown Operations: %d\n", stats.ShutdownOperations)
    fmt.Printf("  Total Hashrate: %.1f PH/s\n", stats.TotalHashrate/1000)
    fmt.Printf("  Total Power: %.2f MW\n", stats.TotalPower/1000000)
    fmt.Printf("  Quota Violations (24h): %d\n", stats.QuotaViolations24h)
    fmt.Printf("  Pending Commands: %d\n", stats.PendingCommands)
    
    // Get violation history
    violations, err := miningSvc.GetViolationHistory(ctx, &service.ViolationQuery{
        StartTime: time.Now().AddDate(0, -1, 0),
        EndTime:   time.Now(),
        Limit:     100,
    })
    if err != nil {
        log.Fatalf("Failed to get violations: %v", err)
    }
    
    fmt.Printf("\nRecent Violations (%d):\n", len(violations))
    for _, v := range violations {
        fmt.Printf("  [%s] %s: %s (severity: %s)\n",
            v.Timestamp.Format("2006-01-02 15:04"),
            v.OperationID[:8],
            v.Type,
            v.Severity)
    }
    
    // Get regional breakdown
    regional, err := miningSvc.GetRegionalBreakdown(ctx)
    if err != nil {
        log.Fatalf("Failed to get regional breakdown: %v", err)
    }
    
    fmt.Printf("\nRegional Breakdown:\n")
    for region, data := range regional {
        fmt.Printf("  %s:\n", region)
        fmt.Printf("    Operations: %d\n", data.OperationCount)
        fmt.Printf("    Hashrate: %.1f PH/s\n", data.TotalHashrate/1000)
        fmt.Printf("    Power: %.2f MW\n", data.TotalPower/1000000)
    }
}
```

## Monitoring

Health check endpoint:

```http
GET /health
```

## License

Part of the CSIC Platform - All rights reserved.
