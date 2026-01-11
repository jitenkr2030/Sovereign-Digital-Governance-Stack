# Energy Optimization Service

A microservices-based energy optimization service that provides real-time energy monitoring, carbon emission tracking, and smart load balancing capabilities.

## Features

- **Real-time Energy Monitoring**: Ingest and process energy readings from multiple sources
- **Carbon Emission Tracking**: Calculate and track carbon emissions based on energy source
- **Carbon Budget Management**: Set and monitor carbon budgets with automatic alerts
- **Smart Load Balancing**: Automated load balancing decisions based on real-time metrics
- **Optimization Recommendations**: AI-driven recommendations for energy optimization
- **Multi-source Support**: Grid, solar, wind, natural gas, coal, nuclear, and hydro energy sources

## Architecture

The service follows Hexagonal (Ports & Adapters) architecture:

```
cmd/server/           # Application entry point
internal/
  domain/            # Core business entities and models
  repository/        # Database and cache implementations
  service/           # Business logic layer
  handlers/          # HTTP/gRPC handlers
deploy/              # Deployment configurations
```

## API Endpoints

### Energy Readings
- `POST /api/v1/readings` - Ingest new energy reading
- `GET /api/v1/readings/:id` - Get reading by ID
- `GET /api/v1/readings/facility/:facility_id` - Get readings by facility

### Carbon Tracking
- `POST /api/v1/carbon/process` - Process energy reading and calculate emissions
- `GET /api/v1/carbon/emissions/:facility_id` - Get emissions for facility
- `GET /api/v1/carbon/budget/:facility_id` - Get carbon budget
- `POST /api/v1/carbon/budget` - Create carbon budget
- `GET /api/v1/carbon/status/:facility_id` - Get budget status

### Load Balancing
- `POST /api/v1/load/metrics` - Record load metric
- `GET /api/v1/load/evaluate/:facility_id` - Evaluate current load
- `GET /api/v1/load/zones/:facility_id` - Get zone loads
- `POST /api/v1/balancing/decisions` - Create load balancing decision
- `GET /api/v1/balancing/decisions/:facility_id` - Get pending decisions
- `POST /api/v1/balancing/decisions/:id/execute` - Execute decision
- `POST /api/v1/balancing/auto/:facility_id` - Auto-balance loads

### Optimization
- `GET /api/v1/optimization/recommend/:facility_id` - Get recommendations
- `GET /api/v1/optimization/report/:facility_id` - Get optimization report

## Configuration

Configuration is managed through `config.yaml`:

```yaml
app:
  host: "0.0.0.0"
  port: 8088
  name: "energy-optimization-service"

database:
  host: "10.112.2.4"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "backend_shared"
  table_prefix: "sess_8ab84f7d_energy_"

redis:
  host: "10.112.2.4"
  port: 6379
  key_prefix: "sess:8ab84f7d:energy:"

carbon:
  emission_factors:
    grid: 0.42        # kg CO2 per kWh
    solar: 0.041
    wind: 0.011
    natural_gas: 0.45
    coal: 0.95
    nuclear: 0.012
    hydro: 0.024
  budget_warning_threshold: 80.0
  budget_critical_threshold: 95.0

load_balancing:
  target_utilization_min: 60.0
  target_utilization_max: 80.0
  scale_up_threshold: 85.0
  scale_down_threshold: 50.0
```

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f energy-optimization

# Stop services
docker-compose down
```

### Manual Setup

```bash
# Install dependencies
go mod download

# Build the service
go build -o energy-optimization ./cmd/server

# Run the service
./energy-optimization
```

## Environment Variables

Override configuration using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | 10.112.2.4 |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | PostgreSQL user | postgres |
| `DB_PASSWORD` | PostgreSQL password | postgres |
| `DB_NAME` | Database name | backend_shared |
| `REDIS_HOST` | Redis host | 10.112.2.4 |
| `REDIS_PORT` | Redis port | 6379 |
| `APP_PORT` | Application port | 8088 |

## Data Isolation

This service uses the following isolation patterns:

- **PostgreSQL**: Table prefix `sess_8ab84f7d_energy_` for session isolation
- **Redis**: Key prefix `sess:8ab84f7d:energy:` for session isolation
- **Kafka**: Topic prefix `sess_8ab84f7d_` for session isolation

## Carbon Emission Factors

| Energy Source | kg CO2/kWh | Category |
|--------------|------------|----------|
| Coal | 0.95 | High |
| Natural Gas | 0.45 | Medium |
| Grid Average | 0.42 | Medium |
| Nuclear | 0.012 | Low |
| Solar | 0.041 | Low |
| Wind | 0.011 | Low |
| Hydro | 0.024 | Low |
| Battery | 0.00 | Zero |

## Example Usage

### Ingest Energy Reading

```bash
curl -X POST http://localhost:8088/api/v1/readings \
  -H "Content-Type: application/json" \
  -d '{
    "meter_id": "meter-001",
    "facility_id": "facility-001",
    "energy_source": "grid",
    "consumption": 150.5,
    "voltage": 400.0,
    "current": 62.5,
    "power_factor": 0.95,
    "timestamp": "2024-01-15T10:30:00Z"
  }'
```

### Get Carbon Budget Status

```bash
curl http://localhost:8088/api/v1/carbon/status/facility-001
```

### Get Load Evaluation

```bash
curl http://localhost:8088/api/v1/load/evaluate/facility-001
```

## Monitoring

### Prometheus Metrics

The service exposes metrics on port 9090:

- `energy_consumption_total` - Total energy consumed
- `energy_utilization_rate` - Current utilization percentage
- `carbon_emission_total` - Total carbon emissions
- `carbon_budget_usage_percentage` - Budget usage percentage
- `load_balancing_decisions_total` - Number of load balancing decisions

### Health Check

```bash
curl http://localhost:8088/health
```

## Grafana Dashboards

Pre-configured dashboards are available in `grafana/dashboards/`:

1. **Energy Overview** - Main dashboard with key metrics
2. **Carbon Tracking** - Carbon emission visualization
3. **Load Balancing** - Load metrics and decisions

## Project Structure

```
energy-optimization/
├── cmd/server/
│   └── main.go
├── internal/
│   ├── domain/
│   │   └── models.go
│   ├── repository/
│   │   ├── postgres_repository.go
│   │   └── redis_repository.go
│   ├── service/
│   │   ├── energy_service.go
│   │   ├── carbon_service.go
│   │   └── load_balancing_service.go
│   └── handlers/
│       └── http_handler.go
├── deploy/
├── grafana/
│   ├── dashboards/
│   │   └── energy-overview.json
│   └── provisioning/
│       ├── datasources/
│       │   └── datasources.yml
│       └── dashboards/
│           └── dashboards.yml
├── Dockerfile
├── docker-compose.yml
├── prometheus.yml
├── config.yaml
├── go.mod
└── README.md
```

## Dependencies

- **Go 1.21+**
- **PostgreSQL 14+**
- **Redis 6+**
- **Prometheus 2.48+**
- **Grafana 10+**

## License

This project is part of the CSIC (Crypto Security Integration Core) platform.
