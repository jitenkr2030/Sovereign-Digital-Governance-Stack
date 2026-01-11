# NEAM Intervention Workflow Worker

This service implements the Temporal workflow infrastructure for the NEAM platform's intervention layer.

## Overview

The Intervention Workflow Worker is responsible for executing economic intervention workflows using Temporal's durable execution model. It implements the Saga pattern to ensure reliable execution of interventions with compensation transactions for rollback scenarios.

## Architecture

### Components

1. **Main Worker Service** (`main.go`)
   - Initializes Temporal client connection
   - Manages worker lifecycle
   - Handles graceful shutdown

2. **Workflows** (`workflows/`)
   - `ExecuteInterventionWorkflow`: Main saga for intervention execution
   - `InterventionApprovalWorkflow`: Human-in-the-loop approval workflow
   - `BatchInterventionWorkflow`: Batch processing of multiple interventions

3. **Activities** (`activities/`)
   - Policy validation and retrieval
   - Budget reservation/release
   - Intervention execution
   - Feature store updates
   - Stakeholder notifications
   - Outcome recording

### Configuration

Configuration is loaded from environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `TEMPORAL_HOST` | Temporal server address | `temporal-frontend.temporal.svc.cluster.local:7233` |
| `TEMPORAL_NAMESPACE` | Temporal namespace | `neam-dev` |
| `TEMPORAL_TASK_QUEUE` | Task queue for interventions | `intervention-critical` |
| `APP_ENVIRONMENT` | Environment (dev/staging/prod) | `development` |
| `APP_LOG_LEVEL` | Logging level | `info` |

## Workflow Details

### ExecuteInterventionWorkflow

The main intervention workflow implements a Saga pattern with the following steps:

1. **Validate Policy** - Verify the policy is valid and applicable
2. **Get Policy Details** - Retrieve policy configuration and limits
3. **Reserve Budget** - Reserve budget for the intervention (with compensation)
4. **Execute Intervention** - Execute the actual intervention
5. **Update Feature Store** - Record the intervention result
6. **Notify Stakeholders** - Send notifications
7. **Record Outcome** - Store outcome for analytics

### Compensation Transactions

If any step fails after budget reservation:
- Release Budget - Rollback the budget reservation
- Rollback Intervention - Attempt to undo the intervention if possible

### Timeout Configuration

- Total Workflow Timeout: 30 minutes
- Activity Timeout: 5 minutes
- Decision Timeout: 10 seconds
- Heartbeat Timeout: 30 seconds

## Building and Running

### Prerequisites

- Go 1.21+
- Access to Temporal cluster
- PostgreSQL, Redis, and Kafka connections (for activities)

### Build

```bash
cd intervention
go build -o bin/workflow-worker .
```

### Run

```bash
./bin/workflow-worker
```

### Run with Custom Configuration

```bash
TEMPORAL_HOST=localhost:7233 TEMPORAL_NAMESPACE=neam-dev \
  TEMPORAL_TASK_QUEUE=intervention-critical \
  ./bin/workflow-worker
```

## Testing

### Unit Tests

```bash
go test -v ./...
```

### Integration Tests

```bash
# Requires Temporal cluster running
go test -v ./tests/integration/...
```

## Monitoring

### Metrics

The worker exposes Prometheus metrics on port 9090:
- `workflow_completions_total`
- `workflow_failures_total`
- `activity_execution_total`
- `activity_latency_seconds`

### Health Check

```bash
curl http://localhost:9090/metrics
```

## Deployment

### Kubernetes

Deploy using Helm:

```bash
helm install neam-workflow ./deployments/helm/neam-workflow \
  --set temporal.host=temporal-frontend.temporal.svc.cluster.local \
  --set temporal.namespace=neam-prod \
  --set image.tag=v2.0.0
```

### Temporal UI

Access the Temporal Web UI at `http://temporal-web.temporal.svc.cluster.local:8088` to:
- View running workflows
- Inspect workflow history
- Debug failed executions
- Manually signal workflows
