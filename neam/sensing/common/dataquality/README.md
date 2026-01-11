# Data Quality Monitoring System

This module provides comprehensive data quality monitoring capabilities for the NEAM platform, including quality scoring, anomaly detection, and alerting.

## Architecture

```
neam-platform/sensing/common/dataquality/
├── scorer.go       # Quality score calculation
├── anomaly.go      # Anomaly detection
├── monitor.go      # Central monitoring service
└── dashboard.go    # Grafana dashboard
```

## Key Concepts

### Quality Dimensions

The system monitors four key quality dimensions:

| Dimension | Weight | Description | Threshold |
|-----------|--------|-------------|-----------|
| Completeness | 25% | Ratio of non-null required fields | 80% |
| Validity | 35% | Percentage of valid records | 95% |
| Consistency | 20% | Self-consistency of data | 90% |
| Timeliness | 20% | Latency of data arrival | 85% |

### Overall Score Calculation

```
OverallScore = w1*Completeness + w2*Validity + w3*Consistency + w4*Timeliness
```

Where weights default to: `{0.25, 0.35, 0.20, 0.20}`

## Usage

### Basic Scoring

```go
import "neam-platform/sensing/common/dataquality"

logger := zap.NewNop()
scorer := dataquality.NewDefaultScorer(logger)

record := map[string]interface{}{
    "id":    "vehicle-123",
    "speed": 65.5,
    "fuel":  0.75,
}

metadata := dataquality.RecordMetadata{
    RecordID:  "rec-001",
    EventTime: time.Now(),
}

analysis := scorer.AnalyzeRecord("transport-adapter", record, metadata)

fmt.Printf("Quality Score: %.2f%%\n", analysis.OverallScore)
fmt.Printf("Completeness: %.2f%%\n", analysis.Completeness)
fmt.Printf("Validity: %.2f%%\n", analysis.Validity)
```

### Anomaly Detection

```go
config := dataquality.DefaultAnomalyConfig()
detector := dataquality.NewDetector(config, logger)

// Observe metric values
for i := 0; i < 100; i++ {
    detector.Observe("transport-adapter", "quality_score", 90.0)
}

// Detect anomalies
anomaly := detector.Observe("transport-adapter", "quality_score", 50.0)
if anomaly != nil {
    fmt.Printf("Anomaly detected: %s\n", anomaly.Description)
}
```

### Monitoring Service

```go
monitor, err := dataquality.NewMonitor(
    dataquality.DefaultMonitorConfig(),
    nil, // Optional schema registry
    logger,
)
defer monitor.Stop()

ctx := context.Background()
monitor.Start(ctx)

// Process records
monitor.Observe("transport-adapter", record, metadata)

// Get current scores
scores := monitor.GetAllScores()
for source, score := range scores {
    fmt.Printf("%s: %.2f%%\n", source, score.OverallScore)
}
```

## Prometheus Metrics

### Quality Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `neam_data_quality_overall_score` | Gauge | source | Overall quality score (0-100) |
| `neam_data_quality_completeness_score` | Gauge | source, field | Completeness score |
| `neam_data_quality_validity_score` | Gauge | source, rule | Validity score |
| `neam_data_quality_consistency_score` | Gauge | source, check | Consistency score |
| `neam_data_quality_timeliness_score` | Gauge | source | Timeliness score |

### Anomaly Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `neam_data_anomaly_detected_total` | Counter | source, type, severity | Anomalies detected |
| `neam_data_anomaly_score` | Gauge | source | Current anomaly indicator |

### Histogram Metrics

| Metric | Description |
|--------|-------------|
| `neam_data_quality_score_histogram` | Distribution of quality scores |
| `neam_data_quality_overall_histogram` | Distribution of overall scores |

## Alert Rules

### Critical Alerts

```yaml
groups:
  - name: data-quality-alerts
    rules:
      # Critical: Quality score too low
      - alert: DataQualityCritical
        expr: neam_data_quality_overall_score < 60
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Data quality critical for {{ $labels.source }}"

      # Warning: Quality degradation
      - alert: DataQualityDegraded
        expr: neam_data_quality_overall_score < 80
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Data quality degraded for {{ $labels.source }}"

      # Critical: Validity too low
      - alert: DataValidityCritical
        expr: neam_data_quality_validity_score{source=~".*"} < 70
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Data validity critical for {{ $labels.source }}"
```

## Grafana Dashboard

The module provides a pre-configured Grafana dashboard (`dashboard.go`) with panels for:

1. Overall Quality Score Gauge
2. Quality by Dimension Bar Gauge
3. Quality Score Trend Over Time
4. Schema Validation Success Rate
5. Validation Errors by Type
6. Anomalies Detected
7. Data Completeness Heatmap
8. Quality Score Distribution
9. Data Timeliness (Latency)
10. Active Alerts Table

### Import Dashboard

```bash
# Export dashboard JSON
dashboard := dataquality.GenerateDataQualityDashboard()
jsonBytes, _ := dashboard.ToJSON()

# Save to file
os.WriteFile("neam-data-quality-dashboard.json", jsonBytes, 0644)
```

Then import `neam-data-quality-dashboard.json` in Grafana.

## Configuration

### Quality Config

```go
config := dataquality.QualityConfig{
    Weights: dataquality.WeightConfig{
        Completeness: 0.25,
        Validity:     0.35,
        Consistency:  0.20,
        Timeliness:  0.20,
    },
    CompletenessThreshold: 0.80,
    ValidityThreshold:     0.95,
    ConsistencyThreshold:  0.90,
    TimelinessThreshold:   0.85,
    OverallThreshold:      0.85,
    WindowSize:            100,
}
```

### Anomaly Config

```go
config := dataquality.AnomalyConfig{
    ZScoreThreshold:    3.0,
    StdDevThreshold:    2.0,
    WindowSize:         100,
    MinWindowSize:      10,
    SilenceDuration:    5 * time.Minute,
    DropThreshold:      0.2,
    SpikeMultiplier:    3.0,
    AlertCooldown:      5 * time.Minute,
    EnableAlerts:       true,
}
```

## Thresholds Reference

| Metric | Warning | Critical | Unit |
|--------|---------|----------|------|
| Overall Quality Score | < 80 | < 60 | % |
| Completeness | < 80 | < 60 | % |
| Validity | < 90 | < 70 | % |
| Consistency | < 85 | < 70 | % |
| Timeliness | < 80 | < 60 | % |
| Data Latency | > 60s | > 300s | seconds |
| Validation Success Rate | < 95% | < 90% | % |

## Best Practices

1. **Initial Baseline**: Run the system for 24-48 hours to establish baseline quality scores before setting strict thresholds.

2. **Dimension Weighting**: Adjust weights based on your domain:
   - High-stakes: Increase Validity weight
   - Real-time systems: Increase Timeliness weight
   - Data warehousing: Increase Completeness weight

3. **Anomaly Tuning**: Start with conservative Z-score thresholds (3.0+) and adjust based on false positive rates.

4. **Alert Cooldown**: Set alert cooldowns to prevent alert fatigue (5-15 minutes recommended).

5. **Dashboard Review**: Review the Grafana dashboard at least daily during initial deployment.

## Integration with Adapters

### Transport Adapter

```go
// In transport/adapter.go
type TransportAdapter struct {
    // ... existing fields
    qualityMonitor *dataquality.Monitor
}

// In Start()
for point := range stream {
    record := map[string]interface{}{
        "id":        point.ID,
        "source_id": point.SourceID,
        "speed":     point.Value["speed"],
        "fuel":      point.Value["fuel_level"],
    }

    metadata := dataquality.RecordMetadata{
        RecordID:  point.ID,
        EventTime: point.Timestamp,
    }

    analysis := t.qualityMonitor.Observe("transport", record, metadata)
    if analysis.OverallScore < 80 {
        t.logger.Warn("Low quality score detected",
            zap.Float64("score", analysis.OverallScore))
    }
}
```

### Industrial Adapter

```go
// In industrial/adapter.go
type IndustrialAdapter struct {
    // ... existing fields
    qualityMonitor *dataquality.Monitor
}

// Process telemetry
record := map[string]interface{}{
    "node_id":   point.NodeID,
    "asset_id":  point.AssetID,
    "value":     point.Value,
    "quality":   point.Quality,
}

metadata := dataquality.RecordMetadata{
    RecordID:  point.ID,
    EventTime: point.Timestamp,
}

analysis := i.qualityMonitor.Observe("industrial", record, metadata)
```

### Energy Adapter

```go
// In energy/adapter.go
type EnergyAdapter struct {
    // ... existing fields
    qualityMonitor *dataquality.Monitor
}

// Process grid telemetry
record := map[string]interface{}{
    "point_id": point.PointID,
    "value":    point.Value,
    "quality":  point.Quality,
}

metadata := dataquality.RecordMetadata{
    RecordID:  point.ID,
    EventTime: point.Timestamp,
}

analysis := e.qualityMonitor.Observe("energy", record, metadata)
```
