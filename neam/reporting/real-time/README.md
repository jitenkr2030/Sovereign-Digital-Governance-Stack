# Real-Time Dashboard Service

This package provides real-time economic monitoring and dashboard capabilities.

## Overview

The Real-Time Dashboard Service continuously monitors key economic indicators and provides instantaneous visibility into the national economic landscape. The service collects data from multiple sources including payment processing systems, energy grid operators, transportation networks, and industrial production facilities.

## Key Features

### Economic Metrics Collection
The service aggregates metrics across six primary economic domains:

- **Payment Metrics**: Tracks transaction volumes, payment method distribution, and regional payment patterns
- **Energy Metrics**: Monitors electricity consumption, generation by source, and grid stability
- **Transport Metrics**: Measures logistics throughput, fuel consumption, and port activity
- **Industry Metrics**: Tracks industrial output indices and sector-specific performance
- **Inflation Metrics**: Monitors price indices across categories and regions
- **Anomaly Detection**: Identifies unusual patterns that may indicate economic stress

### Dashboard Management
Pre-configured dashboards provide focused views into specific economic domains:
- National Overview: Aggregate view of all economic indicators
- Regional Dashboards: Detailed metrics for specific geographic regions
- Sector Dashboards: Focused analysis of individual economic sectors

### Anomaly Detection
The service implements a rule-based anomaly detection system that monitors for:
- Unusual payment volume spikes or drops
- Significant energy consumption changes
- Transportation disruptions
- Inflation rate anomalies
- Supply chain disruptions

## Architecture

### Data Flow
1. Metrics are collected from ClickHouse every 30 seconds
2. Data is cached in Redis for rapid dashboard access
3. Anomaly detection runs every minute against cached data
4. Detected anomalies are published to Kafka for distribution

### Service Components
- `MetricsCollector`: Gathers metrics from data sources
- `DashboardManager`: Manages dashboard configurations
- `AnomalyDetector`: Identifies unusual patterns
- `CacheManager`: Handles Redis caching

## Usage

### Initialize Service
```go
config := real_time.Config{
    ClickHouse: clickhouseConn,
    Redis:      redisClient,
    Logger:     logger,
}
service := real_time.NewService(config)
```

### Get Current Metrics
```go
metrics, err := service.GetCurrentMetrics(ctx)
```

### List Dashboards
```go
dashboards, err := service.ListDashboards(ctx)
```

## Dependencies
- ClickHouse: Primary data source for metrics
- Redis: Caching layer for rapid access
- Kafka: Event distribution for anomalies
