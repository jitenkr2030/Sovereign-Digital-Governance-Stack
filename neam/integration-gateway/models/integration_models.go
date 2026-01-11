package models

import (
	"time"
)

// IntegrationMessage represents a message for data exchange
type IntegrationMessage struct {
	MessageID   string                 `json:"message_id"`
	Source      string                 `json:"source"`
	Destination string                 `json:"destination,omitempty"`
	DataType    string                 `json:"data_type"`
	Version     string                 `json:"version"`
	Payload     interface{}            `json:"payload"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
}

// AdapterRequest represents a request to an external adapter
type AdapterRequest struct {
	Method      string                 `json:"method"`
	Endpoint    string                 `json:"endpoint"`
	Payload     interface{}            `json:"payload,omitempty"`
	Params      map[string]string      `json:"params,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	DataType    string                 `json:"data_type"`
	Transform   bool                   `json:"transform"`
	SourceType  string                 `json:"source_type,omitempty"` // "rest", "file", "ftp"
	Timeout     time.Duration          `json:"timeout,omitempty"`
}

// AdapterResponse represents a response from an external adapter
type AdapterResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// IntegrationRequest represents a request to the integration gateway
type IntegrationRequest struct {
	RequestID    string                 `json:"request_id"`
	Source       string                 `json:"source"`
	DataType     string                 `json:"data_type"`
	Operation    string                 `json:"operation"` // "fetch", "push", "query"
	Payload      interface{}            `json:"payload,omitempty"`
	Params       map[string]string      `json:"params,omitempty"`
	Transform    bool                   `json:"transform"`
	Federated    bool                   `json:"federated"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
	Priority     int                    `json:"priority,omitempty"` // 0-9, higher is more urgent
	CallbackURL  string                 `json:"callback_url,omitempty"`
}

// IntegrationResponse represents a response from the integration gateway
type IntegrationResponse struct {
	RequestID    string                 `json:"request_id"`
	Success      bool                   `json:"success"`
	StatusCode   int                    `json:"status_code"`
	Message      string                 `json:"message,omitempty"`
	Data         interface{}            `json:"data,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	ProcessingTime time.Duration        `json:"processing_time"`
}

// FederationRequest represents a federated query request
type FederationRequest struct {
	RequestID    string                 `json:"request_id"`
	Query        FederationQuery         `json:"query"`
	Sources      []FederationSource      `json:"sources"`
	MergeStrategy string                `json:"merge_strategy"` // "union", "intersection", "join"
	Timeout      time.Duration          `json:"timeout"`
	Priority     int                    `json:"priority"`
}

// FederationQuery represents a query for federated search
type FederationQuery struct {
	Terms       []string               `json:"terms"`
	Filters     map[string]interface{} `json:"filters"`
	DateRange   *DateRange             `json:"date_range,omitempty"`
	Fields      []string               `json:"fields,omitempty"`
	Sort        *SortSpec              `json:"sort,omitempty"`
	Pagination  *PaginationSpec        `json:"pagination,omitempty"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SortSpec represents sorting specification
type SortSpec struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// PaginationSpec represents pagination specification
type PaginationSpec struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// FederationSource represents a data source for federation
type FederationSource struct {
	Type        string                 `json:"type"` // "ministry", "central_bank", "state", "district"
	Endpoint    string                 `json:"endpoint"`
	Credentials map[string]string      `json:"credentials,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
}

// FederationResponse represents a federated query response
type FederationResponse struct {
	RequestID    string                 `json:"request_id"`
	Success      bool                   `json:"success"`
	Results      []FederationResult     `json:"results"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	TotalCount   int                    `json:"total_count"`
	Timestamp    time.Time              `json:"timestamp"`
}

// FederationResult represents a result from a federated source
type FederationResult struct {
	Source     string                 `json:"source"`
	SourceType string                 `json:"source_type"`
	Data       interface{}            `json:"data"`
	Count      int                    `json:"count"`
	TimeTaken  time.Duration          `json:"time_taken"`
	Error      string                 `json:"error,omitempty"`
}

// LegacyFile represents a legacy file to process
type LegacyFile struct {
	ID           string                 `json:"id"`
	Filename     string                 `json:"filename"`
	FilePath     string                 `json:"file_path"`
	Format       string                 `json:"format"` // "csv", "xml", "fixed-width"
	Encoding     string                 `json:"encoding"`
	DataType     string                 `json:"data_type"`
	Schema       string                 `json:"schema"`
	ProcessedAt  time.Time              `json:"processed_at,omitempty"`
	Status       string                 `json:"status"` // "pending", "processing", "completed", "failed"
	Error        string                 `json:"error,omitempty"`
	RecordCount  int                    `json:"record_count"`
}

// LegacyMapping represents a field mapping for legacy data
type LegacyMapping struct {
	ID           string   `json:"id"`
	SourceFormat string   `json:"source_format"` // "csv", "xml", etc.
	TargetFormat string   `json:"target_format"` // "neam-canonical"
	SourceFields []string `json:"source_fields"`
	TargetField  string   `json:"target_field"`
	Transform    string   `json:"transform,omitempty"` // "direct", "date", "number", "custom"
	DefaultValue string   `json:"default_value,omitempty"`
}

// DataMapping represents a data mapping configuration
type DataMapping struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	SourceType   string              `json:"source_type"`
	TargetType   string              `json:"target_type"`
	Version      string              `json:"version"`
	Mappings     []FieldMapping      `json:"mappings"`
	Transforms   []TransformRule     `json:"transforms"`
	ValidFrom    time.Time           `json:"valid_from"`
	ValidUntil   time.Time           `json:"valid_until,omitempty"`
	Active       bool                `json:"active"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// FieldMapping represents a field-level mapping
type FieldMapping struct {
	SourceField    string      `json:"source_field"`
	TargetField    string      `json:"target_field"`
	SourceType     string      `json:"source_type"`
	TargetType     string      `json:"target_type"`
	Required       bool        `json:"required"`
	Transform      string      `json:"transform,omitempty"`
	TransformFunc  string      `json:"transform_func,omitempty"`
	DefaultValue   interface{} `json:"default_value,omitempty"`
	Validation     *Validation  `json:"validation,omitempty"`
}

// TransformRule represents a transformation rule
type TransformRule struct {
	Field        string                 `json:"field"`
	Type         string                 `json:"type"` // "date", "number", "string", "custom"
	Format       string                 `json:"format"`
	TargetFormat string                 `json:"target_format"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// Validation represents field validation rules
type Validation struct {
	Type       string      `json:"type"` // "required", "pattern", "range", "enum"
	Value      interface{} `json:"value"`
	Message    string      `json:"message"`
}

// IntegrationConfig represents integration configuration
type IntegrationConfig struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"` // "ministry", "central_bank", "state", "district", "legacy"
	Endpoint       string                 `json:"endpoint"`
	Authentication *AuthConfig            `json:"authentication"`
	RateLimit      *RateLimitConfig       `json:"rate_limit"`
	Timeout        time.Duration          `json:"timeout"`
	RetryPolicy    *RetryPolicy           `json:"retry_policy"`
	Enabled        bool                   `json:"enabled"`
	HealthCheck    *HealthCheckConfig     `json:"health_check"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type        string            `json:"type"` // "basic", "oauth2", "api_key", "mtls"
	Credentials map[string]string `json:"credentials"`
	TokenURL    string            `json:"token_url,omitempty"`
	Scopes      []string          `json:"scopes,omitempty"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
}

// RetryPolicy represents retry policy configuration
type RetryPolicy struct {
	MaxAttempts  int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Endpoint string        `json:"endpoint"`
}

// DataExchangeEvent represents an event for data exchange
type DataExchangeEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"` // "request", "response", "error", "timeout"
	RequestID   string                 `json:"request_id"`
	Source     string                 `json:"source"`
	Destination string                 `json:"destination"`
	DataType   string                 `json:"data_type"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}
