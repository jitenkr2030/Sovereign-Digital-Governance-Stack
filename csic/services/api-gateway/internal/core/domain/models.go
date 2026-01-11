package domain

import (
	"time"
)

// EntityID is a type alias for entity identifiers.
type EntityID string

// NewEntityID generates a new unique entity identifier.
func NewEntityID() EntityID {
	return EntityID(generateGatewayUUID())
}

// Route represents an API route configuration.
type Route struct {
	ID          EntityID        `json:"id"`
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	Methods     []string        `json:"methods"`
	UpstreamURL string          `json:"upstream_url"`
	UpstreamPath string         `json:"upstream_path"`
	Timeout     int             `json:"timeout"` // in seconds
	RetryCount  int             `json:"retry_count"`
	Plugins     []PluginConfig  `json:"plugins"`
	IsActive    bool            `json:"is_active"`
	RateLimit   *RateLimitConfig `json:"rate_limit,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PluginConfig represents configuration for a gateway plugin.
type PluginConfig struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
}

// RateLimitConfig represents rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerMinute int    `json:"requests_per_minute"`
	RequestsPerHour   int    `json:"requests_per_hour"`
	LimitBy           string `json:"limit_by"` // "ip", "user", "api_key"
	Policy            string `json:"policy"`   // "sliding_window", "fixed_window", "token_bucket"
}

// Consumer represents an API consumer.
type Consumer struct {
	ID        EntityID            `json:"id"`
	Username  string              `json:"username"`
	APIKeys   []APIKey            `json:"api_keys"`
	Groups    []string            `json:"groups"`
	Quota     *QuotaConfig        `json:"quota,omitempty"`
	IsActive  bool                `json:"is_active"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// APIKey represents an API key for authentication.
type APIKey struct {
	ID        EntityID    `json:"id"`
	Key       string      `json:"key"`
	Name      string      `json:"name"`
	Scopes    []string    `json:"scopes"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// QuotaConfig represents quota configuration for a consumer.
type QuotaConfig struct {
	RequestsPerDay  int    `json:"requests_per_day"`
	RequestsPerMonth int   `json:"requests_per_month"`
	BandwidthMB     int    `json:"bandwidth_mb"`
	ResetTime       string `json:"reset_time"` // UTC time when quota resets
}

// Service represents an upstream service.
type Service struct {
	ID          EntityID        `json:"id"`
	Name        string          `json:"name"`
	Host        string          `json:"host"`
	Port        int             `json:"port"`
	Protocol    string          `json:"protocol"` // "http", "https", "grpc"
	HealthCheck *HealthCheck    `json:"health_check,omitempty"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// HealthCheck represents health check configuration for a service.
type HealthCheck struct {
	Path        string `json:"path"`
	Interval    int    `json:"interval"` // in seconds
	Timeout     int    `json:"timeout"`  // in seconds
	UnhealthyThreshold int `json:"unhealthy_threshold"`
	HealthyThreshold  int `json:"healthy_threshold"`
}

// Certificate represents an SSL/TLS certificate.
type Certificate struct {
	ID          EntityID    `json:"id"`
	Domain      string      `json:"domain"`
	CertData    string      `json:"cert_data"`
	KeyData     string      `json:"key_data"`
	Issuer      string      `json:"issuer"`
	NotBefore   time.Time   `json:"not_before"`
	NotAfter    time.Time   `json:"not_after"`
	IsActive    bool        `json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
}

// GatewayMetrics represents gateway operation metrics.
type GatewayMetrics struct {
	TotalRequests      int64             `json:"total_requests"`
	TotalErrors        int64             `json:"total_errors"`
	TotalLatencyMs     int64             `json:"total_latency_ms"`
	RequestsByRoute    map[string]int64  `json:"requests_by_route"`
	RequestsByStatus   map[string]int64  `json:"requests_by_status"`
	ErrorsByType       map[string]int64  `json:"errors_by_type"`
	RateLimitedCount   int64             `json:"rate_limited_count"`
	AuthenticatedCount int64             `json:"authenticated_count"`
}

// RateLimitEntry represents a rate limit tracking entry.
type RateLimitEntry struct {
	ID          EntityID    `json:"id"`
	Identifier  string      `json:"identifier"` // IP, user ID, or API key
	RouteID     string      `json:"route_id"`
	Requests    int         `json:"requests"`
	WindowStart time.Time   `json:"window_start"`
	ExpiresAt   time.Time   `json:"expires_at"`
}

// AnalyticsEvent represents an analytics event.
type AnalyticsEvent struct {
	ID          EntityID              `json:"id"`
	Timestamp   time.Time             `json:"timestamp"`
	EventType   string                `json:"event_type"`
	RouteID     string                `json:"route_id"`
	ConsumerID  *string               `json:"consumer_id,omitempty"`
	RequestID   string                `json:"request_id"`
	IPAddress   string                `json:"ip_address"`
	UserAgent   string                `json:"user_agent"`
	Method      string                `json:"method"`
	Path        string                `json:"path"`
	StatusCode  int                   `json:"status_code"`
	LatencyMs   int64                 `json:"latency_ms"`
	BodyBytes   int64                 `json:"body_bytes"`
	ErrorMessage *string              `json:"error_message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TransformRule represents a request/response transformation rule.
type TransformRule struct {
	ID          EntityID            `json:"id"`
	Name        string              `json:"name"`
	RouteID     string              `json:"route_id"`
	Type        string              `json:"type"` // "request_header", "response_header", "request_body", "response_body"
	Action      string              `json:"action"` // "add", "remove", "replace", "rename"
	Target      string              `json:"target"` // header name or json path
	Value       string              `json:"value,omitempty"`
	Pattern     string              `json:"pattern,omitempty"`
	IsActive    bool                `json:"is_active"`
	CreatedAt   time.Time           `json:"created_at"`
}

// generateGatewayUUID generates a UUID-like identifier.
func generateGatewayUUID() string {
	// Simplified UUID generation for gateway entities
	return time.Now().UTC().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
