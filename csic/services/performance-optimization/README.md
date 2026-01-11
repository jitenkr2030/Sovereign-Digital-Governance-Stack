# Performance Optimization Service

A comprehensive performance benchmarking, load testing, and profiling service for the CSIC Platform.

## Features

- **Benchmarking**: Run CPU, memory, network, and disk benchmarks
- **Load Testing**: Various scenarios (steady, ramp-up, spike, soak, stress, wave)
- **Profiling**: CPU, memory, goroutine, block, mutex, and trace profiling
- **Metrics Collection**: Real-time system metrics collection
- **Reporting**: Generate HTML and JSON reports

## Quick Start

### Using Docker

```bash
# Build and start the service
docker-compose up -d

# Access the API
curl http://localhost:8085/health
```

### From Source

```bash
# Build the service
cd services/performance-optimization
go build -o perf-optimize ./cmd/server

# Start the server
./perf-optimize serve

# Run a benchmark
./perf-optimize benchmark http://api.example.com -t all -d 60

# Run a load test
./perf-optimize loadtest http://api.example.com -s 300 -p 1000

# Run profiling
./perf-optimize profile my-service -t cpu -d 30
```

## Configuration

All configuration is managed via `config.yaml`:

```yaml
# Application settings
app:
  host: "0.0.0.0"
  port: 8085

# Benchmark settings
benchmark:
  default_duration: 60
  max_concurrency: 1000
  warmup_duration: 10

# Metrics settings
metrics:
  collection_interval: 5
  profiling_enabled: true
  profiling_port: 6060

# Load test settings
loadtest:
  rampup_time: 30
  steady_state_time: 60
  target_rps: 1000
  max_error_rate: 5.0
```

## API Endpoints

### Health Check
```
GET /health
```

### Benchmarks
```
GET  /api/v1/benchmarks       - List all benchmarks
POST /api/v1/benchmarks       - Create a new benchmark
GET  /api/v1/benchmarks/{id}  - Get benchmark details
DELETE /api/v1/benchmarks/{id} - Delete a benchmark
```

### Load Tests
```
GET  /api/v1/loadtests        - List all load tests
POST /api/v1/loadtests        - Create a new load test
GET  /api/v1/loadtests/{id}   - Get load test details
DELETE /api/v1/loadtests/{id} - Delete a load test
```

### Profiling
```
GET  /api/v1/profiles         - List all profiles
POST /api/v1/profiles         - Start a profiling session
GET  /api/v1/profiles/{id}    - Get profile details
DELETE /api/v1/profiles/{id}  - Delete a profile
```

### Metrics
```
GET /api/v1/metrics           - Get current metrics
GET /api/v1/metrics/history   - Get metrics history
```

### Reports
```
GET  /api/v1/reports          - List all reports
POST /api/v1/reports          - Generate a new report
GET  /api/v1/reports/{id}     - Get report details
```

## Profiling Endpoints

When profiling is enabled, standard Go pprof endpoints are available:

```
/debug/pprof/           - Index
/debug/pprof/cmdline    - Command line
/debug/pprof/profile    - CPU profile
/debug/pprof/symbol     - Symbol
/debug/pprof/trace      - Trace
/debug/pprof/goroutine  - Goroutine profile
/debug/pprof/heap       - Heap profile
/debug/pprof/block      - Block profile
/debug/pprof/mutex      - Mutex profile
```

## Benchmark Types

- `cpu` - CPU-intensive benchmarks
- `memory` - Memory allocation benchmarks
- `network` - Network operation benchmarks
- `disk` - Disk I/O benchmarks
- `all` - Run all benchmark types

## Load Test Scenarios

- `steady` - Constant load scenario
- `rampup` - Gradually increasing load
- `spike` - Sudden spike in load
- `soak` - Extended endurance test
- `stress` - Increasing load until failure
- `wave` - Wave pattern load

## Metrics Collected

- CPU usage percentage
- Memory usage (used, total, available)
- Goroutine count
- GC statistics (pause time, count)
- Custom metrics from registered providers

## Architecture

```
services/performance-optimization/
├── cmd/
│   └── server/           # Main entry point
├── internal/
│   ├── benchmarks/       # Benchmarking logic
│   ├── loadtest/         # Load testing logic
│   ├── metrics/          # Metrics collection
│   └── profilers/        # Profiling logic
├── config.yaml           # Configuration
├── Dockerfile            # Docker build
└── docker-compose.yml    # Docker orchestration
```

## Integration with Other Services

The Performance Optimization service integrates with:

- **Forensic Tools**: Can be used to profile suspicious activity
- **Oversight**: Performance data for anomaly detection
- **Incident Response**: Performance incidents trigger investigations

## License

Part of the CSIC Platform.
