# NEAM Payment Adapter

## Overview

The Payment Adapter is a critical component of the NEAM Platform's Sensing Layer, responsible for ingesting, parsing, validating, and normalizing financial messages from various sources. It supports ISO 20022 and SWIFT MT/MX message formats, ensuring transactional integrity and enabling real-time and batch processing modes.

## Key Features

### Supported Message Formats

- **ISO 20022 Messages**: Supports multiple message types including:
  - `pacs.008` - Credit Transfer
  - `pacs.004` - Payment Return
  - `camt.053` - Bank-to-Customer Statement
  - `camt.052` - Bank-to-Customer Account Report
  - `pain.001` - Customer Credit Transfer Initiation
  - `pain.002` - Customer Payment Status Request

- **SWIFT MT Messages**: Supports major MT message types including:
  - `MT103` - Customer Credit Transfer
  - `MT202` - General Financial Institution Transfer
  - `MT202COV` - Cover Payment
  - `MT900` - Confirmation of Debit
  - `MT910` - Confirmation of Credit

- **SWIFT MX Messages**: XML-based messages following ISO 20022 standards

### Processing Capabilities

- **Real-time Processing**: Kafka-based streaming for immediate message handling
- **Batch Processing**: File-based processing for bulk message ingestion
- **Hybrid Processing**: Combined real-time and batch capabilities

### Reliability Features

- **Idempotency**: Prevents duplicate processing using Redis-based deduplication
- **Retry Logic**: Exponential backoff with jitter for transient failures
- **Dead Letter Queue**: Isolates failed messages for later analysis
- **Circuit Breaker**: Prevents cascading failures from downstream issues

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Payment Adapter                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐                  │
│  │  Kafka      │───▶│  Message    │───▶│  Format         │                  │
│  │  Consumer   │    │  Detector   │    │  Detection      │                  │
│  └─────────────┘    └─────────────┘    └────────┬────────┘                  │
│                                                  │                           │
│                    ┌─────────────────────────────▼──────────────────────┐   │
│                    │              Message Processor                      │   │
│                    │  ┌─────────────┐    ┌─────────────┐                 │   │
│                    │  │   ISO       │    │   SWIFT      │                 │   │
│                    │  │   20022     │    │   MT/MX      │                 │   │
│                    │  │   Parser    │    │   Parser     │                 │   │
│                    │  └─────────────┘    └─────────────┘                 │   │
│                    └─────────────────────────────┬──────────────────────┘   │
│                                                  │                           │
│                    ┌─────────────────────────────▼──────────────────────┐   │
│                    │            Idempotency Manager                      │   │
│                    │   (Redis-based deduplication with TTL)             │   │
│                    └─────────────────────────────┬──────────────────────┘   │
│                                                  │                           │
│  ┌─────────────┐    ┌─────────────────────────────▼──────────────────────┐   │
│  │  Kafka      │◀───│           Output Normalizer                        │   │
│  │  Producer   │    │   (Standardized payment representation)            │   │
│  └─────────────┘    └────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐   │
│  │                     Worker Pool (10 workers)                           │   │
│  └───────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.21 or later
- Apache Kafka 2.8+ or compatible
- Redis 6+ or compatible
- Access to NEAM shared utilities

### Building

```bash
cd sensing/payments
go mod tidy
go build -o payment-adapter main.go
```

### Configuration

Create a configuration file based on the example:

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your environment settings
```

### Running

```bash
# With default configuration
./payment-adapter

# With custom configuration path
./payment-adapter --config /path/to/config.yaml

# Using environment variables
export KAFKA_BROKERS=kafka:29092
export REDIS_ADDR=redis:6379
./payment-adapter
```

### Docker

```bash
docker build -t neam-payment-adapter .
docker run -d \
  --name payment-adapter \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/logs:/app/logs \
  neam-payment-adapter
```

## Usage

### Real-time Processing

Configure the adapter for real-time Kafka message processing:

```yaml
processing:
  mode: realtime
  worker_count: 10
```

Messages consumed from the input topic are processed and normalized to the output topic.

### Batch Processing

Configure the adapter for batch file processing:

```yaml
processing:
  mode: batch
  batch_directory: /data/batch/payments
  scan_interval: 1m
  archive_processed: true
```

Supported file formats: XML, TXT, JSON

### Hybrid Processing

Configure the adapter for combined real-time and batch processing:

```yaml
processing:
  mode: hybrid
```

## API Reference

### Health Check

```bash
curl http://localhost:8084/health
```

### Metrics

```bash
curl http://localhost:9090/metrics
```

### Input/Output Formats

#### Input Topics

| Topic | Description |
|-------|-------------|
| `payments.raw` | Raw payment messages in ISO 20022 or SWIFT format |

#### Output Topics

| Topic | Description |
|-------|-------------|
| `payments.normalized` | Normalized payment data |
| `payments.dlq` | Failed messages for manual review |
| `payments.retry` | Messages for retry processing |

#### Normalized Payment Structure

```json
{
  "id": "PAY-abc123",
  "message_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "format": "iso20022",
  "message_type": "pacs.008",
  "transaction_id": "MSGID20240101ABCD1234",
  "amount": "1000.00",
  "currency": "EUR",
  "sender_bic": "BANKDEFF",
  "receiver_bic": "BANKUS33",
  "sender_account": "DE89370400440532013001",
  "receiver_account": "DE89370400440532013000",
  "value_date": "2024-01-15T00:00:00Z",
  "processing_date": "2024-01-15T10:30:00Z",
  "timestamp": "2024-01-15T10:30:00Z",
  "raw_payload": {...},
  "metadata": {...}
}
```

## Testing

### Unit Tests

```bash
go test -v -cover ./...
go test -race -cover ./...
```

### Performance Testing

```bash
# Run benchmarks
go test -bench=. -benchmem

# Expected results:
# - Throughput: >5000 messages/second
# - P99 Latency: <50ms
```

### Integration Tests

```bash
# Requires running Kafka and Redis
go test -v ./integration/...
```

## Monitoring

### Key Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| `messages_processed_total` | Total messages processed | Increasing |
| `messages_failed_total` | Total failed messages | Minimize |
| `processing_duration_seconds` | Message processing duration | P99 < 50ms |
| `idempotency_duplicates` | Duplicate messages detected | Monitor |
| `retry_attempts_total` | Total retry attempts | Minimize |

### Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| Processing Lag | Consumer lag > 1000 messages | Warning |
| High Failure Rate | Failure rate > 5% | Critical |
| DLQ Growth | DLQ topic size increasing | Warning |

## Troubleshooting

### Common Issues

#### Messages Not Being Processed

1. Check Kafka connectivity:
   ```bash
   # Verify broker accessibility
   kafkacat -b kafka:29092 -L
   ```

2. Check Redis connectivity:
   ```bash
   redis-cli -h redis ping
   ```

3. Check adapter logs:
   ```bash
   tail -f /var/log/neam/payment-adapter.log
   ```

#### Duplicate Processing

1. Verify idempotency TTL configuration
2. Check for Redis memory pressure
3. Review idempotency metrics

#### High Latency

1. Check worker pool size
2. Verify Kafka consumer group lag
3. Review downstream service health

### Log Analysis

```bash
# Filter for errors
grep "ERROR" /var/log/neam/payment-adapter.log

# Filter for specific message type
grep "pacs.008" /var/log/neam/payment-adapter.log

# Filter by transaction ID
grep "MSGID20240101" /var/log/neam/payment-adapter.log
```

## Configuration Reference

### Complete Configuration Options

```yaml
service:
  name: string                    # Service name
  host: string                    # Bind address
  port: int                       # HTTP port
  environment: string             # Environment name
  shutdown_timeout: duration      # Graceful shutdown timeout

kafka:
  brokers: string                 # Kafka broker addresses
  consumer_group: string          # Consumer group ID
  input_topic: string             # Input topic name
  output_topic: string            # Output topic name
  dlq_topic: string               # Dead letter queue topic
  retry_topic: string             # Retry topic name
  version: string                 # Kafka version
  max_open_requests: int          # Max concurrent requests
  dial_timeout: duration          # Connection timeout
  read_timeout: duration          # Read timeout
  write_timeout: duration         # Write timeout
  session_timeout: duration       # Session timeout
  heartbeat_interval: duration    # Heartbeat interval

redis:
  addr: string                    # Redis address
  password: string                # Redis password
  db: int                         # Database number
  pool_size: int                  # Connection pool size
  min_idle_conns: int             # Min idle connections
  dial_timeout: duration          # Connection timeout
  read_timeout: duration          # Read timeout
  write_timeout: duration         # Write timeout
  idempotency_ttl: duration       # Idempotency key TTL

processing:
  mode: string                    # realtime, batch, hybrid
  worker_count: int               # Number of workers
  batch_size: int                 # Batch processing size
  batch_timeout: duration         # Batch timeout
  batch_directory: string         # Batch file directory
  parallel_files: int             # Parallel file processing

retry:
  max_retries: int                # Maximum retry attempts
  initial_delay: duration         # Initial backoff delay
  max_delay: duration             # Maximum backoff delay
  multiplier: float64             # Backoff multiplier
  jitter: bool                    # Enable jitter
  jitter_factor: float64          # Jitter factor
  dlq_max_retries: int            # DLQ max retries

logging:
  level: string                   # Log level
  format: string                  # Log format
  output_path: string             # Log output path
  json_format: bool               # JSON logging

monitoring:
  enabled: bool                   # Enable metrics
  port: int                       # Metrics port
  path: string                    # Metrics path
  health_path: string             # Health check path

batch:
  supported_formats: []string     # Supported file formats
  max_file_size: int              # Max file size in bytes
  scan_interval: duration         # Directory scan interval
  archive_processed: bool         # Archive processed files
  archive_path: string            # Archive directory path
```

## Performance Tuning

### Throughput Optimization

1. **Increase worker count** for higher parallelism
2. **Adjust batch size** for optimal throughput
3. **Enable compression** on Kafka topics
4. **Optimize Redis pool** size for workload

### Latency Optimization

1. **Reduce batch timeout** for faster processing
2. **Optimize message size** for network efficiency
3. **Enable connection pooling** for Kafka and Redis
4. **Monitor and tune** GC settings

## Security Considerations

### Data Protection

- All data is encrypted in transit using TLS
- Sensitive fields can be masked in logs
- Access control via Kafka ACLs and Redis authentication

### Audit Trail

- All processing is logged with trace IDs
- Idempotency keys provide processing evidence
- DLQ maintains original messages for compliance

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit pull request with description

## License

This component is part of the NEAM Platform and follows the same licensing terms.

## Support

For issues and questions:
- Review troubleshooting section
- Check monitoring dashboards
- Contact NEAM platform team
