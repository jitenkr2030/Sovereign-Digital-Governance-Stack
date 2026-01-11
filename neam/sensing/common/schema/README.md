# Schema Validation Framework

This module provides comprehensive schema validation capabilities for the NEAM platform, supporting both JSON Schema and Protocol Buffer validation with versioning, metrics, and centralized registry management.

## Architecture

```
neam-platform/sensing/common/schema/
├── registry.go       # Central schema registry with versioning
├── validator.go      # Validator interfaces and common types
├── json_validator.go # JSON Schema validation implementation
├── proto_validator.go# Protobuf validation implementation
├── errors.go         # Structured error handling
├── metrics.go        # Prometheus metrics
└── README.md         # This file
```

## Key Features

- **JSON Schema Validation**: Full Draft-07+ support with custom rules
- **Protocol Buffer Validation**: Dynamic message validation
- **Schema Versioning**: Semantic versioning with migration support
- **Centralized Registry**: Thread-safe schema storage and retrieval
- **Prometheus Metrics**: Built-in observability
- **Structured Errors**: Detailed validation error reporting

## Usage

### Basic JSON Schema Validation

```go
import "neam-platform/sensing/common/schema"

logger := zap.NewNop()

// Create validator
validator, err := schema.NewJSONValidator(`{
    "type": "object",
    "properties": {
        "id": {"type": "string"},
        "value": {"type": "number"},
        "enabled": {"type": "boolean"}
    },
    "required": ["id"]
}`)
if err != nil {
    log.Fatalf("Failed to create validator: %v", err)
}

ctx := context.Background()

// Validate data
data := []byte(`{"id": "test-123", "value": 42.5, "enabled": true}`)
result, err := validator.Validate(ctx, data)
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}

if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Field: %s, Error: %s\n", err.Field, err.Message)
    }
}
```

### Schema Registry

```go
registry := schema.NewRegistry(logger)

// Register a schema
schemaJSON := `{
    "type": "object",
    "properties": {
        "vehicle_id": {"type": "string"},
        "speed": {"type": "number", "minimum": 0, "maximum": 200},
        "fuel_level": {"type": "number", "minimum": 0, "maximum": 1}
    },
    "required": ["vehicle_id"]
}`

err = registry.Register(
    "transport-telemetry",
    "1.0.0",
    schema.SchemaTypeJSON,
    []byte(schemaJSON),
    "Transport vehicle telemetry schema",
)
if err != nil {
    log.Fatalf("Failed to register schema: %v", err)
}

// Get validator for specific version
validator, err = registry.GetValidator("transport-telemetry", "1.0.0")
if err != nil {
    log.Fatalf("Failed to get validator: %v", err)
}

// Get latest version
latestValidator, err := registry.GetLatestValidator("transport-telemetry")
```

### Schema Versioning and Migration

```go
// Register v1.0.0
registry.Register("telemetry", "1.0.0", schema.SchemaTypeJSON, []byte(schemaV1), "")

// Register v2.0.0 with new fields
registry.Register("telemetry", "2.0.0", schema.SchemaTypeJSON, []byte(schemaV2), "")

// Migrate data from v1 to v2
migratedData, err := registry.Migrate(data, "1.0.0", "2.0.0")
```

### Protobuf Validation

```go
schemaData := []byte(`{
    "message_name": "VehicleTelemetry",
    "fields": [
        {"name": "vehicle_id", "number": 1, "type": "string"},
        {"name": "speed", "number": 2, "type": "number"},
        {"name": "timestamp", "number": 3, "type": "int64"}
    ]
}`)

validator, err := schema.NewProtobufValidator(schemaData)
if err != nil {
    log.Fatalf("Failed to create protobuf validator: %v", err)
}
```

### Custom Validation Rules

```go
config := schema.ValidatorConfig{
    CustomRules: []schema.CustomRule{
        {
            Name:       "speed_positive",
            Field:      "speed",
            Predicate:  "greater_than",
            Message:    "Speed must be positive",
            Parameters: map[string]string{"min": "0"},
        },
        {
            Name:       "speed_max",
            Field:      "speed",
            Predicate:  "less_than",
            Message:    "Speed cannot exceed maximum",
            Parameters: map[string]string{"max": "200"},
        },
    },
    RequiredFields: []string{"vehicle_id", "speed"},
}

validator, err := schema.NewJSONValidatorWithConfig(
    "transport",
    "1.0.0",
    schemaJSON,
    config,
    logger,
)
```

## Prometheus Metrics

### Validation Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `neam_schema_validation_total` | Counter | source, schema_id, schema_version, status | Total validation attempts |
| `neam_schema_validation_success_total` | Counter | - | Successful validations |
| `neam_schema_validation_failure_total` | Counter | source, schema_id, schema_version, error_type | Failed validations |
| `neam_schema_validation_duration_seconds` | Histogram | source, schema_id, schema_type | Validation duration |

### Error Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `neam_schema_validation_errors_by_field_total` | Counter | source, schema_id, field, error_type | Errors by field |
| `neam_schema_validation_errors_by_type_total` | Counter | source, schema_id, error_type | Errors by type |
| `neam_schema_validation_errors_by_category_total` | Counter | source, schema_id, category | Errors by category |

### Schema Registry Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `neam_schema_registered_total` | Counter | Total schemas registered |
| `neam_schema_active` | Gauge | Number of active schemas |

## Alert Rules

```yaml
groups:
  - name: schema-validation-alerts
    rules:
      # High validation failure rate
      - alert: SchemaValidationHighFailureRate
        expr: |
          rate(neam_schema_validation_failure_total[5m]) /
          rate(neam_schema_validation_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Schema validation failure rate > 5% for {{ $labels.source }}"

      # Critical validation failures
      - alert: SchemaValidationCriticalFailures
        expr: |
          increase(neam_schema_validation_failure_total{error_type="required"}[5m]) > 100
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Critical validation failures for {{ $labels.source }}"

      # Missing required fields spike
      - alert: SchemaMissingRequiredFields
        expr: |
          rate(neam_schema_validation_errors_by_category_total{category="required"}[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Spike in missing required fields for {{ $labels.source }}"
```

## Error Categories

| Category | Description | Example |
|----------|-------------|---------|
| `required` | Missing required field | `{"id": "123"}` when `name` is required |
| `type` | Wrong data type | `"abc"` when number expected |
| `format` | Invalid format | `"not-an-email"` for email format |
| `range` | Value out of range | `201` when max is 200 |
| `pattern` | Pattern mismatch | `"abc"` doesn't match `^[0-9]+$` |
| `enum` | Invalid enum value | `"invalid"` for allowed values |
| `custom` | Custom rule violation | Speed below minimum |

## Error Response Format

```json
{
  "valid": false,
  "errors": [
    {
      "field": "vehicle_id",
      "message": "field is required",
      "error_type": "required",
      "constraint": "required"
    },
    {
      "field": "speed",
      "message": "value 250 exceeds maximum 200",
      "expected": "200",
      "actual": "250",
      "error_type": "maximum",
      "constraint": "maximum"
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": "0.00123s"
}
```

## Integration with Adapters

### Transport Adapter

```go
// In transport/adapter.go
type TransportAdapter struct {
    // ... existing fields
    schemaRegistry *schema.Registry
    metrics        *schema.SchemaMetrics
}

// Initialize schema registry
t.schemaRegistry = schema.NewRegistry(logger)

// Register transport telemetry schema
t.schemaRegistry.Register(
    "transport-telemetry",
    "1.0.0",
    schema.SchemaTypeJSON,
    []byte(transportTelemetrySchema),
    "Transport vehicle telemetry schema",
)

// Validate incoming data
validator, _ := t.schemaRegistry.GetLatestValidator("transport-telemetry")
result, err := validator.Validate(ctx, data)

if !result.Valid {
    t.metrics.RecordValidation("transport", "transport-telemetry", "1.0.0", false, 0.001, "validation_error")
    // Handle invalid data
}
```

### Industrial Adapter

```go
// In industrial/adapter.go
type IndustrialAdapter struct {
    // ... existing fields
    schemaRegistry *schema.Registry
}

// Register industrial telemetry schema
i.schemaRegistry.Register(
    "industrial-sensor",
    "1.0.0",
    schema.SchemaTypeJSON,
    []byte(sensorSchema),
    "Industrial sensor data schema",
)

// Validate OPC UA data
validator, _ := i.schemaRegistry.GetLatestValidator("industrial-sensor")
result, _ := validator.Validate(ctx, data)
```

### Energy Adapter

```go
// In energy/adapter.go
type EnergyAdapter struct {
    // ... existing fields
    schemaRegistry *schema.Registry
}

// Register grid telemetry schema
e.schemaRegistry.Register(
    "energy-grid",
    "1.0.0",
    schema.SchemaTypeJSON,
    []byte(gridTelemetrySchema),
    "Energy grid telemetry schema",
)

// Validate ICCP/IEC 61850 data
validator, _ := e.schemaRegistry.GetLatestValidator("energy-grid")
result, _ := validator.Validate(ctx, data)
```

## Schema Definition Examples

### Transport Telemetry Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Transport Vehicle Telemetry",
  "description": "Schema for transport vehicle telemetry data",
  "type": "object",
  "properties": {
    "vehicle_id": {
      "type": "string",
      "minLength": 1,
      "maxLength": 50,
      "description": "Unique vehicle identifier"
    },
    "timestamp": {
      "type": "string",
      "format": "date-time",
      "description": "Event timestamp in ISO 8601 format"
    },
    "speed": {
      "type": "number",
      "minimum": 0,
      "maximum": 200,
      "description": "Vehicle speed in km/h"
    },
    "fuel_level": {
      "type": "number",
      "minimum": 0,
      "maximum": 1,
      "description": "Fuel level as fraction (0-1)"
    },
    "location": {
      "type": "object",
      "properties": {
        "latitude": {
          "type": "number",
          "minimum": -90,
          "maximum": 90
        },
        "longitude": {
          "type": "number",
          "minimum": -180,
          "maximum": 180
        }
      },
      "required": ["latitude", "longitude"]
    },
    "engine_temperature": {
      "type": "number",
      "minimum": -40,
      "maximum": 150
    }
  },
  "required": ["vehicle_id", "timestamp", "speed"]
}
```

### Industrial Sensor Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Industrial Sensor Data",
  "type": "object",
  "properties": {
    "node_id": {
      "type": "string",
      "pattern": "^ns=2;s=Channel\\d+\\.Device\\d+\\..+$"
    },
    "asset_id": {
      "type": "string",
      "minLength": 1
    },
    "value": {
      "type": "number"
    },
    "quality": {
      "type": "string",
      "enum": ["Good", "Uncertain", "Bad"]
    },
    "timestamp": {
      "type": "string",
      "format": "date-time"
    },
    "data_type": {
      "type": "string",
      "enum": ["Double", "Int", "Float", "Boolean", "String"]
    },
    "unit": {
      "type": "string"
    }
  },
  "required": ["node_id", "value", "quality", "timestamp"]
}
```

### Energy Grid Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Energy Grid Telemetry",
  "type": "object",
  "properties": {
    "point_id": {
      "type": "string",
      "description": "Grid measurement point identifier"
    },
    "frequency": {
      "type": "number",
      "minimum": 49.5,
      "maximum": 50.5
    },
    "voltage": {
      "type": "number",
      "minimum": 0,
      "maximum": 500
    },
    "current": {
      "type": "number",
      "minimum": 0
    },
    "power": {
      "type": "number"
    },
    "timestamp": {
      "type": "string",
      "format": "date-time"
    },
    "quality": {
      "type": "string",
      "enum": ["Good", "Uncertain", "Bad"]
    }
  },
  "required": ["point_id", "frequency", "voltage", "timestamp"]
}
```

## Best Practices

1. **Schema Evolution**: Use backward-compatible changes when possible. Add new fields as optional.

2. **Versioning Strategy**: Use semantic versioning (MAJOR.MINOR.PATCH):
   - MAJOR: Breaking changes
   - MINOR: New optional fields
   - PATCH: Bug fixes

3. **Performance**: For high-throughput scenarios, cache compiled validators.

4. **Monitoring**: Always track validation metrics to detect data quality issues early.

5. **Error Handling**: Log validation errors with context but don't expose schema internals to clients.

6. **Testing**: Write unit tests for schema validation, especially for edge cases.

## Performance Considerations

- JSON Schema compilation is expensive: compile once, reuse validator
- Use validation mode `permissive` for data exploration, `strict` for production
- Consider schema caching for frequently used schemas
- Monitor validation latency in production
