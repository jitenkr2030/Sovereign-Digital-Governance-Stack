package signalproc

import (
	"time"
)

// TelemetryPoint represents a single data point from the grid
type TelemetryPoint struct {
	ID        string          `json:"id"`
	SourceID  string          `json:"source_id"`
	PointID   string          `json:"point_id"`
	Value     float64         `json:"value"`
	Quality   string          `json:"quality"`
	Protocol  string          `json:"protocol"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  []byte          `json:"metadata,omitempty"`
}
