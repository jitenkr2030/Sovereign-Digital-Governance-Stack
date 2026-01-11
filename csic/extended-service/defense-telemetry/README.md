# Defense Telemetry Service

A microservices-based defense telemetry service that provides real-time monitoring, threat detection, and strategic deployment management for military units.

## Features

- **Real-time Telemetry Monitoring**: Ingest and process telemetry data from air, land, sea, and cyber units
- **Threat Detection**: Automated threat detection with severity-based alerting
- **Equipment Health Tracking**: Predictive maintenance based on component health analysis
- **Strategic Deployment**: Optimized unit positioning and deployment management
- **Geo-Spatial Tracking**: Real-time unit location tracking with proximity alerts
- **Multi-domain Support**: Air, land, sea, naval, cyber, and space domain units

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

### Telemetry
- `POST /api/v1/telemetry/ingest` - Ingest single telemetry data point
- `POST /api/v1/telemetry/batch` - Ingest batch telemetry data
- `GET /api/v1/telemetry/unit/:unit_id` - Get latest telemetry for a unit
- `GET /api/v1/telemetry/unit/:unit_id/status` - Get comprehensive unit status
- `GET /api/v1/telemetry/units` - List all units
- `POST /api/v1/telemetry/units` - Create a new unit
- `GET /api/v1/telemetry/units/nearby` - Find nearby units

### Threats
- `POST /api/v1/threats` - Report a new threat
- `GET /api/v1/threats/active` - Get active threats
- `GET /api/v1/threats/stats` - Get threat statistics
- `GET /api/v1/threats/:id` - Get threat details
- `POST /api/v1/threats/:id/resolve` - Resolve a threat
- `POST /api/v1/threats/:id/escalate` - Escalate a threat

### Maintenance
- `POST /api/v1/maintenance/logs` - Create maintenance log
- `GET /api/v1/maintenance/unit/:unit_id/predictions` - Get maintenance predictions
- `GET /api/v1/maintenance/unit/:unit_id/health` - Analyze equipment health
- `GET /api/v1/maintenance/upcoming` - Get upcoming maintenance

### Deployment
- `POST /api/v1/deployment` - Create a deployment
- `GET /api/v1/deployment/active` - Get active deployments
- `GET /api/v1/deployment/stats` - Get deployment statistics
- `GET /api/v1/deployment/optimize/:unit_id` - Get optimization recommendation
- `GET /api/v1/deployment/zones` - List strategic zones
- `POST /api/v1/deployment/zones` - Create strategic zone

## Configuration

Configuration is managed through `config.yaml`:

```yaml
app:
  host: "0.0.0.0"
  port: 8089
  name: "defense-telemetry-service"

database:
  host: "10.112.2.4"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "backend_shared"
  table_prefix: "sess_8ab84f7d_defense_"

redis:
  host: "10.112.2.4"
  port: 6379
  key_prefix: "sess:8ab84f7d:defense:"

equipment:
  hull_integrity_warning: 70.0
  hull_integrity_critical: 50.0
  fuel_level_warning: 20.0
  fuel_level_critical: 10.0

threats:
  high_severity_threshold: 3
  critical_severity_threshold: 4
  auto_escalate: true
```

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f defense-telemetry

# Stop services
docker-compose down
```

### Manual Setup

```bash
# Install dependencies
go mod download

# Build the service
go build -o defense-telemetry ./cmd/server

# Run the service
./defense-telemetry
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
| `APP_PORT` | Application port | 8089 |

## Data Isolation

This service uses the following isolation patterns:

- **PostgreSQL**: Table prefix `sess_8ab84f7d_defense_` for session isolation
- **Redis**: Key prefix `sess:8ab84f7d:defense:` for session isolation
- **Kafka**: Topic prefix `sess_8ab84f7d_defense_` for session isolation

## Unit Types

| Type | Description |
|------|-------------|
| `air` | Fixed-wing and rotary aircraft |
| `land` | Ground vehicles and infantry |
| `sea` | Surface vessels |
| `naval` | Naval submarines and ships |
| `cyber` | Cyber warfare units |
| `space` | Space-based assets |

## Threat Types

| Type | Description |
|------|-------------|
| `missile` | Ballistic or cruise missile |
| `drone` | Unmanned aerial vehicle |
| `cyber` | Cyber attack |
| `electronic_jamming` | Electronic warfare |
| `maritime_vessel` | Hostile surface vessel |
| `aerial` | Manned aircraft |
| `ground_assault` | Ground force assault |
| `submarine` | Submarine threat |
| `ied` | Improvised explosive device |
| `chemical` | Chemical threat |
| `biological` | Biological threat |
| `nuclear` | Nuclear threat |

## Severity Levels

| Level | Description | Response |
|-------|-------------|----------|
| 1 | Minor | Monitor |
| 2 | Low | Investigate |
| 3 | Medium | Alert commander |
| 4 | High | Immediate response |
| 5 | Critical | Emergency protocols |

## Example Usage

### Ingest Telemetry

```bash
curl -X POST http://localhost:8089/api/v1/telemetry/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "unit_id": "unit-001",
    "coordinates": {
      "latitude": 40.7128,
      "longitude": -74.0060,
      "altitude": 1000
    },
    "velocity": 150.5,
    "fuel_level": 75.0,
    "ammunition_count": 200,
    "hull_integrity": 95.0,
    "engine_temp": 75.0
  }'
```

### Report Threat

```bash
curl -X POST http://localhost:8089/api/v1/threats \
  -H "Content-Type: application/json" \
  -d '{
    "unit_id": "unit-001",
    "threat_type": "missile",
    "threat_source": "Unknown",
    "severity": 4,
    "source_location": {
      "latitude": 40.7500,
      "longitude": -73.9800
    }
  }'
```

### Get Unit Status

```bash
curl http://localhost:8089/api/v1/telemetry/unit/unit-001/status
```

## Monitoring

### Prometheus Metrics

The service exposes metrics on port 9090:

- `telemetry_ingested_total` - Total telemetry readings ingested
- `threats_detected_total` - Total threats detected
- `active_threats_count` - Current active threats
- `unit_health_score` - Average unit health score

### Health Check

```bash
curl http://localhost:8089/health
```

## Grafana Dashboards

Pre-configured dashboards are available in `grafana/dashboards/`:

1. **Defense Overview** - Main dashboard with key metrics
2. **Threat Analysis** - Threat visualization and trends
3. **Unit Status** - Individual unit health and status

## Project Structure

```
defense-telemetry/
├── cmd/server/
│   └── main.go
├── internal/
│   ├── domain/
│   │   └── models.go
│   ├── repository/
│   │   ├── postgres_repository.go
│   │   └── redis_repository.go
│   ├── service/
│   │   ├── telemetry_service.go
│   │   ├── threat_service.go
│   │   ├── maintenance_service.go
│   │   └── deployment_service.go
│   └── handlers/
│       └── http_handler.go
├── deploy/
├── grafana/
│   ├── dashboards/
│   └── provisioning/
│       ├── datasources/
│       └── dashboards/
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
