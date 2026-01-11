# Energy Adapter

The Energy Adapter is a component of the NEAM (National Energy and Analytics Marketplace) Platform's Sensing Layer. It provides real-time telemetry ingestion from electrical grid infrastructure using industrial communication protocols.

## Overview

This adapter connects to various energy grid data sources including control centers, substations, and renewable energy monitoring systems. It processes and filters telemetry data before publishing it to downstream systems via Kafka.

## Features

- **Multi-Protocol Support**: Connect to grid infrastructure using ICCP (TASE.2) and IEC 61850 (MMS) protocols
- **Signal Processing**: Advanced filtering including deadband, spike detection, and rate-of-change validation
- **Connection Pooling**: Efficient management of multiple concurrent protocol connections
- **Session Management**: Persistent session handling with automatic reconnection
- **Real-Time Processing**: Low-latency processing of high-frequency grid measurements
- **Kubernetes Ready**: Production-ready deployment manifests with monitoring

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Energy Adapter Service                        │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ ICCP Driver │  │ IEC 61850   │  │ Signal Processing       │  │
│  │ (TASE.2)    │  │ Driver      │  │ - Deadband Filter       │  │
│  └─────────────┘  └─────────────┘  │ - Spike Detection       │  │
│                                    │ - Rate of Change        │  │
│  ┌─────────────┐  ┌─────────────┐  └─────────────────────────┘  │
│  │ Connection  │  │ Session     │                               │
│  │ Pool        │  │ Manager     │                               │
│  └─────────────┘  └─────────────┘                               │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Kafka Producer                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.21 or later
- Access to Kafka cluster
- Configuration file (config.yaml)

### Building

```bash
# Clone the repository
git clone https://github.com/neamplatform/neam-platform.git

# Navigate to energy adapter directory
cd neam-platform/sensing/energy

# Build the binary
go build -o energy-adapter .

```

### Configuration

Create a `config.yaml` file with the following structure:

```yaml
service:
  name: "energy-adapter"
  host: "0.0.0.0"
  port: 8085
  environment: "production"

# ICCP/TASE.2 Protocol Configuration
iccp:
  enabled: true
  sources:
    - endpoint: "192.168.1.100:102"
      local_tsap: "1.0.1"
      remote_tsap: "1.0.2"
      bilateral_table: "grid-exchange"
  points_of_interest:
    - "Frequency"
    - "Voltage"
    - "Current"
    - "Power"
  timeout: 30s
  retry_interval: 5s
  keep_alive: 10s

# IEC 61850/MMS Protocol Configuration
iec61850:
  enabled: true
  sources:
    - endpoint: "192.168.1.101"
      port: 102
      ied_name: "Substation-IED-001"
      ap_title: "1.3.9999.13"
      ae_qualifier: 12
  points_of_interest:
    - "MMXU1.Vol"
    - "MMXU1.A"
    - "MMXU1.W"
    - "XCBR1.Pos"
  timeout: 30s
  retry_interval: 5s
  keep_alive: 10s

# Kafka Output Configuration
kafka:
  brokers: "kafka-cluster:9092"
  output_topic: "energy.telemetry"
  dlq_topic: "energy.dlq"
  version: "2.8.0"
  max_open_requests: 100
  dial_timeout: 10s
  write_timeout: 30s

# Session Management
session:
  max_retries: 10
  initial_delay: 1s
  max_delay: 60s
  multiplier: 2.0
  health_check_interval: 30s
  reconnect_wait: 5s

# Signal Processing Configuration
filtering:
  deadband_percent: 0.1       # 0.1% minimum change to report
  spike_threshold: 20.0       # 20% deviation triggers spike detection
  spike_window_size: 5        # Number of samples for spike detection
  rate_of_change_limit: 5.0   # Maximum % change per second
  quality_filter:
    - "Good"
  min_reporting_interval: 100ms
  max_value: 1000000.0
  min_value: -1000000.0

# Logging Configuration
logging:
  level: "info"
  format: "json"
  output_path: "/var/log/neam"
  json_format: true

# Monitoring Configuration
monitoring:
  enabled: true
  port: 9091
  path: "/metrics"
  health_path: "/health"
```

### Running

```bash
# Run with default configuration
./energy-adapter

# Run with custom configuration
./energy-adapter --config=/path/to/config.yaml

# Set configuration via environment variables
export KAFKA_BROKERS="kafka:9092"
export ICCP_ENDPOINT="192.168.1.100:102"
./energy-adapter
```

## Signal Processing

### Deadband Filtering

The deadband filter suppresses small value changes that fall within a specified percentage threshold. This reduces data volume and noise for stable measurements.

```
Deadband Formula:
|change%| = |(new_value - last_reported) / last_reported| × 100
If |change%| ≤ deadband_percent → Filter out
```

### Spike Detection

Spike detection identifies sudden large deviations from recent values. It uses a sliding window average to establish a baseline and flags values exceeding the threshold.

```
Spike Detection:
deviation% = |(current_value - window_average) / window_average| × 100
If deviation% ≥ spike_threshold → Mark as spike
```

### Anomaly Detection

When multiple spikes are detected in succession, the system marks the data stream as potentially anomalous and logs a warning for operator review.

### Rate of Change Validation

Rate of change limiting prevents unrealistic data by checking that value changes do not exceed a maximum percentage per second. This helps identify sensor failures or data corruption.

```
Rate of Change:
rate_of_change%/sec = |(value_change / last_value) × 100| / time_delta_seconds
If rate_of_change%/sec > rate_of_change_limit → Filter out
```

## Testing

### Unit Tests

```bash
# Run all unit tests
go test -v ./...

# Run specific test
go test -v -run TestSignalProcessorFilters

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests

```bash
# Run integration tests (requires Kafka)
go test -v -tags=integration ./...

# Skip short tests
go test -v -short=false ./...
```

### Benchmark Tests

```bash
# Run performance benchmarks
go test -bench=. -benchmem ./...
```

## Docker Deployment

```bash
# Build Docker image
docker build -t neam-energy-adapter:latest .

# Run container
docker run -d \
  --name energy-adapter \
  -v /path/to/config.yaml:/etc/neam/config.yaml \
  -p 9091:9091 \
  neam-energy-adapter:latest
```

## Kubernetes Deployment

The adapter includes Kubernetes manifests for production deployment:

```bash
# Deploy to Kubernetes
kubectl apply -f deployments/k8s/

# Check deployment status
kubectl get pods -l app=energy-adapter

# View logs
kubectl logs -l app=energy-adapter -f
```

### Kubernetes Resources

- **Deployment**: Manages adapter replicas
- **Service**: Provides network access
- **ConfigMap**: Stores configuration
- **Secret**: Manages credentials
- **PrometheusRule**: Configures alerting

## Monitoring

### Metrics

The adapter exposes Prometheus metrics at `/metrics`:

- `energy_adapter_points_processed_total`: Total telemetry points processed
- `energy_adapter_points_filtered_total`: Total points filtered out
- `energy_adapter_spikes_detected_total`: Total spikes detected
- `energy_adapter_connection_health`: Connection health status
- `energy_adapter_processing_latency_seconds`: Processing latency

### Health Endpoints

- `/health`: Basic health check
- `/ready`: Readiness probe
- `/metrics`: Prometheus metrics

### Grafana Dashboards

Pre-configured Grafana dashboards are available in `deployments/grafana/` for visualizing adapter performance.

## Protocol Details

### ICCP (TASE.2)

The Inter-Control Center Protocol (ICCP), also known as TASE.2, is an international standard for data exchange between utility control centers. The adapter implements:

- OSI transport connection management
- Application association establishment
- Bilateral table-based data access
- Transfer set creation for real-time data

### IEC 61850 (MMS)

IEC 61850 is a standard for communication networks and systems in substations. The adapter implements:

- MMS (Manufacturing Message Specification) protocol
- IED (Intelligent Electronic Device) association
- Report control block configuration
- Data model access (MMXU, XCBR, CSWI, PTOC)

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Processing Latency | < 1ms per point |
| Throughput | 10,000+ points/sec |
| Connection Overhead | < 100ms reconnect |
| Memory Usage | ~50MB base |
| CPU Usage | ~1 core at full load |

## Troubleshooting

### Connection Issues

```bash
# Check protocol driver logs
kubectl logs -l app=energy-adapter -c energy-adapter | grep -i "connection"

# Verify endpoint connectivity
nc -zv <endpoint> <port>
```

### Data Not Flowing

1. Verify protocol configuration is correct
2. Check point IDs match data source definitions
3. Review filtering configuration
4. Confirm Kafka connectivity

### High Filter Rate

If many points are being filtered:
1. Review deadband percentage setting
2. Check for appropriate spike threshold
3. Verify rate of change limits

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For questions or issues:
- Open a GitHub issue
- Contact: neam-platform@energy.example.com
- Documentation: [docs/neam-platform](https://docs.neam-platform.example.com)
