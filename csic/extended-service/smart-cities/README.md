# Smart Cities Service

A comprehensive microservices platform for managing urban infrastructure, including traffic management, emergency response coordination, waste management optimization, and smart lighting control.

## Features

### Traffic Management
- Real-time traffic flow monitoring and analytics
- Smart traffic light synchronization
- Traffic pattern prediction using historical data
- Automatic incident detection and response coordination
- Congestion level calculations and optimal signal timing

### Emergency Response
- Real-time emergency incident tracking and prioritization
- Intelligent unit dispatching based on proximity and availability
- Response time estimation and performance metrics
- Integration with multiple emergency service types (police, fire, medical, hazmat)
- Automatic traffic signal adjustment for emergency routes

### Waste Management
- Smart waste bin monitoring with fill level sensors
- Predictive collection scheduling
- Optimal route planning for collection vehicles
- Real-time tracking of waste collection operations
- Environmental impact reporting and analytics

### Smart Lighting
- Adaptive lighting based on ambient conditions and time of day
- Motion-activated brightness adjustment
- Zone-based lighting control and scheduling
- Energy consumption monitoring and optimization
- Predictive maintenance scheduling

## Architecture

The service follows hexagonal (ports and adapters) architecture with the following components:

```
smart-cities/
├── cmd/server/              # Application entry point
├── internal/
│   ├── domain/              # Domain models and business logic
│   │   └── models.go        # Data models for all entities
│   ├── repository/          # Data access layer
│   │   ├── postgres_repository.go
│   │   └── redis_repository.go
│   ├── service/             # Business logic services
│   │   ├── traffic_service.go
│   │   ├── emergency_service.go
│   │   ├── waste_management_service.go
│   │   └── lighting_service.go
│   └── handlers/            # HTTP API handlers
│       └── http_handler.go
├── config.yaml              # Configuration file
├── Dockerfile               # Docker container definition
├── docker-compose.yml       # Multi-container orchestration
├── prometheus.yml           # Prometheus metrics configuration
└── README.md               # This file
```

## API Endpoints

### Traffic Management
- `GET /api/v1/traffic/nodes` - List all traffic monitoring nodes
- `GET /api/v1/traffic/nodes/:id` - Get specific traffic node details
- `GET /api/v1/traffic/nodes/:id/analytics` - Get traffic analytics
- `GET /api/v1/traffic/incidents` - List traffic incidents
- `POST /api/v1/traffic/incidents` - Report traffic incident
- `GET /api/v1/traffic/predict/:node_id` - Predict traffic patterns
- `POST /api/v1/traffic/sync-lights` - Synchronize traffic lights

### Emergency Response
- `GET /api/v1/emergency/units` - List all emergency units
- `GET /api/v1/emergency/units/:id` - Get specific unit details
- `GET /api/v1/emergency/incidents` - List active incidents
- `POST /api/v1/emergency/incidents` - Report new incident
- `POST /api/v1/emergency/dispatch` - Dispatch unit to incident
- `POST /api/v1/emergency/resolve` - Mark incident as resolved
- `GET /api/v1/emergency/metrics` - Get response metrics

### Waste Management
- `GET /api/v1/waste/bins` - List all waste bins
- `GET /api/v1/waste/bins/:id` - Get specific bin details
- `PUT /api/v1/waste/bins/:id/fill` - Update bin fill level
- `POST /api/v1/waste/collect` - Record waste collection
- `GET /api/v1/waste/route/:vehicle_id` - Get optimal collection route
- `GET /api/v1/waste/analytics` - Get waste management analytics

### Smart Lighting
- `GET /api/v1/lighting/fixtures` - List all lighting fixtures
- `GET /api/v1/lighting/fixtures/:id` - Get specific fixture details
- `PUT /api/v1/lighting/fixtures/:id/brightness` - Set fixture brightness
- `POST /api/v1/lighting/zones/:id/adjust` - Adjust zone lighting
- `POST /api/v1/lighting/fixtures/:id/motion` - Handle motion event
- `GET /api/v1/lighting/metrics` - Get lighting system metrics
- `GET /api/v1/lighting/savings` - Get energy savings report

### Health Check
- `GET /health` - Service health status

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (if running without Docker)
- Redis 7+ (if running without Docker)

### Running with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd extended-services/smart-cities

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f smart-cities

# Stop all services
docker-compose down
```

### Running Locally

```bash
# Clone the repository
git clone <repository-url>
cd extended-services/smart-cities

# Install dependencies
go mod download

# Run the service
go run cmd/server/main.go --config config.yaml
```

### Running Tests

```bash
# Run unit tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Configuration

The service uses a YAML configuration file (`config.yaml`) for all settings:

```yaml
app:
  host: "0.0.0.0"
  port: 8090
  environment: "development"
  log_level: "info"

database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  name: "smart_cities"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

metrics:
  enabled: true
  path: "/metrics"
  port: 9090
```

## Monitoring

### Prometheus Metrics

The service exposes Prometheus metrics at `/metrics`:

- `http_requests_total` - Total HTTP requests by method and path
- `http_request_duration_seconds` - HTTP request latency
- `traffic_congestion_level` - Current traffic congestion levels
- `emergency_response_time_seconds` - Emergency response times
- `waste_bin_fill_level` - Waste bin fill percentages
- `lighting_energy_consumption_watts` - Energy consumption metrics

### Grafana Dashboards

Pre-configured dashboards are available in the `grafana/dashboards` directory:

- Traffic Management Overview
- Emergency Response Performance
- Waste Collection Efficiency
- Lighting System Analytics

Access Grafana at `http://localhost:3000` (default credentials: admin/smartcities123)

## Database Schema

The service uses PostgreSQL with the following main tables:

- `traffic_nodes` - Traffic monitoring and signal control points
- `traffic_incidents` - Reported traffic incidents
- `emergency_units` - Available emergency response units
- `emergency_incidents` - Emergency incident reports
- `waste_bins` - Smart waste collection bins
- `waste_collections` - Collection event history
- `lighting_fixtures` - Smart lighting fixtures
- `lighting_zones` - Lighting control zones

## License

This project is proprietary software developed for smart city infrastructure management.
