package signalproc

import (
	"time"
)

// TelemetryPoint represents a single data point from industrial systems
type TelemetryPoint struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`
	NodeID      string                 `json:"node_id"`
	AssetID     string                 `json:"asset_id,omitempty"`
	AssetName   string                 `json:"asset_name,omitempty"`
	Value       interface{}            `json:"value"`
	Quality     string                 `json:"quality"`
	Protocol    string                 `json:"protocol"`
	Timestamp   time.Time              `json:"timestamp"`
	DataType    string                 `json:"data_type,omitempty"`
	Unit        string                 `json:"unit,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
