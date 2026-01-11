# Transport Adapter

The Transport Adapter is a component of the NEAM (National Energy and Analytics Marketplace) Platform's Sensing Layer. It provides real-time telemetry ingestion from transportation systems using MQTT v5 protocol.

## Overview

This adapter connects to various transportation data sources including vehicle telemetry, fleet management systems, and IoT sensors. It processes and filters telemetry data before publishing it to downstream systems via Kafka.

## Features

- **MQTT v5 Support**: Full MQTT v5 protocol implementation with enhanced features
- **Quality of Service (QoS)**: Configurable QoS levels based on data criticality
- **Session Persistence**: Automatic session resumption after connection loss
- **Signal Processing**: Advanced filtering including deadband, spike detection, and outlier detection
- **Connection Pooling**: Efficient management of multiple concurrent connections
- **Real-Time Processing**: Low-latency processing of high-frequency transport data
- **Kubernetes Ready**: Production-ready deployment manifests with monitoring

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Transport Adapter Service                       │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ MQTT v5     │  │ Signal      │  │ Session                │  │
│  │ Driver      │  │ Processing  │  │ Management             │  │
│  │ - QoS 0/1/2 │  │ - Deadband  │  │ - State Persistence    │  │
│  │ - CleanStart│  │ - Outlier   │  │ - Reconnection         │  │
│  │ - User Props│  │ - Rate Limit│  │                        │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Kafka Producer                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## MQTT v5 QoS Configuration

The adapter supports QoS levels based on data criticality:

| Criticality | QoS Level | Delivery Guarantee | Use Case |
|-------------|-----------|-------------------|----------|
| critical    | 2         | Exactly-once      | Alerts, emergency signals |
| important   | 1         | At-least-once     | Status updates, warnings |
| standard    | 1         | At-least-once     | Regular telemetry |
| background  | 0         | At-most-once      | Debug logs, bulk data |

### Example Configuration

```yaml
mqtt:
  enabled: true
  sources:
    - broker_url: "mqtt://broker.example.com:1883"
      client_id: "neam-transport-adapter"
      clean_start: false
      topics:
        - "transport/+/telemetry"     # Standard QoS
        - "transport/+/alerts"        # Critical QoS
        - "transport/+/status"        # Important QoS
  qos_by_criticality:
    critical: 2
    important: 1
    standard: 1
    background: 0
  session_expiry: 3600s  # 1 hour session persistence
  keep_alive: 30s
```

### Session Persistence

The adapter maintains MQTT session state across restarts:

- **CleanStart=false**: Resume previous session
- **Session Expiry Interval**: Configurable session timeout
- **Client ID Stability**: Persistent client identifier

## Installation

### Prerequisites

- Go 1.21 or later
- Access to MQTT broker
- Access to Kafka cluster
- Configuration file (config.yaml)

### Building

```bash
# Clone the repository
git clone https://github.com/neamplatform/neam-platform.git

# Navigate to transport adapter directory
cd neam-platform/sensing/transport

# Build the binary
go build -o transport-adapter .
```

### Configuration

Create a `config.yaml` file:

```yaml
service:
  name: "transport-adapter"
  host: "0.0.0.0"
  port: 8086
  environment: "production"

mqtt:
  enabled: true
  sources:
    - broker_url: "mqtt://mqtt-broker:1883"
      client_id: "neam-transport"
      username: "adapter"
      password: "secure-password"
      clean_start: false
      topics:
        - "fleet/+/vehicle/+/telemetry"
        - "fleet/+/vehicle/+/gps"
        - "fleet/+/vehicle/+/alerts"
        - "fleet/+/vehicle/+/status"
  qos_by_criticality:
    critical: 2
    important: 1
    standard: 1
    background: 0
  keep_alive: 30s
  session_expiry: 3600s
  timeout: 10s
  retry_interval: 5s
  reconnect_wait: 1s

kafka:
  brokers: "kafka-cluster:9092"
  output_topic: "transport.telemetry"
  dlq_topic: "transport.dlq"

session:
  max_retries: 10
  initial_delay: 1s
  max_delay: 60s
  multiplier: 2.0
  health_check_interval: 30s
  state_store_type: "redis"
  state_store_url: "redis://redis:6379"

filtering:
  deadband_percent: 0.1
  spike_threshold: 20.0
  spike_window_size: 5
  rate_of_change_limit: 10.0
  quality_filter:
    - "Good"
  min_reporting_interval: 100ms
  max_value: 1e9
  min_value: -1e9
  outlier_threshold: 3.0

logging:
  level: "info"
  format: "json"
  output_path: "/var/log/neam"

monitoring:
  enabled: true
  port: 9092
  path: "/metrics"
  health_path: "/health"
```

### Running

```bash
# Run with default configuration
./transport-adapter

# Run with custom configuration
./transport-adapter --config=/path/to/config.yaml

# Set via environment variables
export MQTT_BROKER_URL="mqtt://broker:1883"
export KAFKA_BROKERS="kafka:9092"
./transport-adapter
```

## Signal Processing

### Deadband Filtering

Suppresses small value changes within a specified threshold:

```
|change%| = |(new_value - last_reported) / last_reported| × 100
If |change%| ≤ deadband_percent → Filter out
```

### Outlier Detection

Statistical outlier detection using z-score:

```
z_score = |value - mean| / std_dev
If z_score ≥ threshold → Mark as outlier
```

### Spike Detection

Identifies sudden large deviations:

```
deviation% = |(current - window_avg) / window_avg| × 100
If deviation% ≥ spike_threshold → Flag as spike
```

### Rate of Change Validation

Prevents unrealistic data changes:

```
rate_of_change%/sec = |(value_change / last_value) × 100| / time_delta
If rate_of_change%/sec > limit → Filter out
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
# Run integration tests (requires MQTT and Kafka)
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
docker build -t neam-transport-adapter:latest .

# Run container
docker run -d \
  --name transport-adapter \
  -v /path/to/config.yaml:/etc/neam/config.yaml \
  -p 9092:9092 \
  neam-transport-adapter:latest
```

## Kubernetes Deployment

```bash
# Deploy to Kubernetes
kubectl apply -f deployments/k8s/

# Check deployment status
kubectl get pods -l app=transport-adapter

# View logs
kubectl logs -l app=transport-adapter -f
```

### Kubernetes Resources

- **Deployment**: Manages adapter replicas
- **Service**: Provides network access
- **ConfigMap**: Stores configuration
- **Secret**: Manages credentials
- **PrometheusRule**: Configures alerting
- **HorizontalPodAutoscaler**: Auto-scaling

## Monitoring

### Metrics

The adapter exposes Prometheus metrics at `/metrics`:

- `transport_adapter_points_processed_total`: Total telemetry points processed
- `transport_adapter_points_filtered_total`: Total points filtered out
- `transport_adapter_spikes_detected_total`: Total spikes detected
- `transport_adapter_outliers_detected_total`: Total outliers detected
- `transport_adapter_connection_health`: Connection health status
- `transport_adapter_processing_latency_seconds`: Processing latency

### Health Endpoints

- `/health`: Basic health check
- `/ready`: Readiness probe
- `/metrics`: Prometheus metrics

### Grafana Dashboards

Pre-configured Grafana dashboards are available in `deployments/grafana/`.

## MQTT Topic Structure

Recommended topic hierarchy:

```
transport/
  {region}/
    {fleet_id}/
      vehicle/{vehicle_id}/
        telemetry        # Standard telemetry (QoS 1)
        gps              # GPS coordinates (QoS 1)
        alerts           # Critical alerts (QoS 2)
        status           # Vehicle status (QoS 1)
        diagnostics      # Debug data (QoS 0)
```

### Example MQTT Message

```json
{
  "vehicle_id": "VH001",
  "timestamp": "2024-01-15T10:30:00Z",
  "location": {
    "latitude": 37.7749,
    "longitude": -122.4194
  },
  "speed": 65.5,
  "heading": 270,
  "fuel_level": 75.2,
  "engine_temp": 92.5,
  "alerts": []
}
```

## Troubleshooting

### Connection Issues

```bash
# Check protocol driver logs
kubectl logs -l app=transport-adapter -c transport-adapter | grep -i "connection"

# Verify broker connectivity
nc -zv <broker_host> <port>
```

### Data Not Flowing

1. Verify MQTT configuration (broker URL, topics)
2. Check topic subscriptions match published topics
3. Review filtering configuration
4. Confirm Kafka connectivity

### High Filter Rate

If many points are filtered:
1. Review deadband percentage setting
2. Check for appropriate spike threshold
3. Verify rate of change limits

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Processing Latency | < 1ms per point |
| Throughput | 50,000+ points/sec |
| Connection Overhead | < 50ms reconnect |
| Memory Usage | ~100MB base |
| CPU Usage | ~1 core at full load |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
