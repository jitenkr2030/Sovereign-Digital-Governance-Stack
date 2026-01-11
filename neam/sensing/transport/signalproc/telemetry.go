package signalproc

import (
	"time"
)

// TelemetryPoint represents a single data point from transport systems
type TelemetryPoint struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`
	PointID     string                 `json:"point_id"`
	Value       interface{}            `json:"value"`
	Quality     string                 `json:"quality"`
	Protocol    string                 `json:"protocol"`
	Timestamp   time.Time              `json:"timestamp"`
	Criticality string                 `json:"criticality,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
