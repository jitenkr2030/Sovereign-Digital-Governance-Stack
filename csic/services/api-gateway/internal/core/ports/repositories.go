package ports

import (
	"context"

	"github.com/api-gateway/gateway/internal/core/domain"
)

// RouteRepository defines the interface for route persistence.
type RouteRepository interface {
	// Create creates a new route.
	Create(ctx context.Context, route *domain.Route) error

	// GetByID retrieves a route by ID.
	GetByID(ctx context.Context, id string) (*domain.Route, error)

	// GetByPath retrieves a route by path and method.
	GetByPath(ctx context.Context, path, method string) (*domain.Route, error)

	// Update updates an existing route.
	Update(ctx context.Context, route *domain.Route) error

	// List retrieves all active routes.
	List(ctx context.Context) ([]*domain.Route, error)

	// Delete soft-deletes a route.
	Delete(ctx context.Context, id string) error
}

// ConsumerRepository defines the interface for consumer persistence.
type ConsumerRepository interface {
	// Create creates a new consumer.
	Create(ctx context.Context, consumer *domain.Consumer) error

	// GetByID retrieves a consumer by ID.
	GetByID(ctx context.Context, id string) (*domain.Consumer, error)

	// GetByUsername retrieves a consumer by username.
	GetByUsername(ctx context.Context, username string) (*domain.Consumer, error)

	// GetByAPIKey retrieves a consumer by API key.
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.Consumer, error)

	// Update updates an existing consumer.
	Update(ctx context.Context, consumer *domain.Consumer) error

	// List retrieves all active consumers.
	List(ctx context.Context) ([]*domain.Consumer, error)

	// Delete soft-deletes a consumer.
	Delete(ctx context.Context, id string) error
}

// ServiceRepository defines the interface for service persistence.
type ServiceRepository interface {
	// Create creates a new service.
	Create(ctx context.Context, service *domain.Service) error

	// GetByID retrieves a service by ID.
	GetByID(ctx context.Context, id string) (*domain.Service, error)

	// GetByName retrieves a service by name.
	GetByName(ctx context.Context, name string) (*domain.Service, error)

	// Update updates an existing service.
	Update(ctx context.Context, service *domain.Service) error

	// List retrieves all active services.
	List(ctx context.Context) ([]*domain.Service, error)

	// Delete soft-deletes a service.
	Delete(ctx context.Context, id string) error
}

// RateLimitRepository defines the interface for rate limit persistence.
type RateLimitRepository interface {
	// Increment increments the request count for an identifier.
	Increment(ctx context.Context, identifier, routeID string, window time.Duration) (int, error)

	// GetRemaining returns remaining requests for an identifier.
	GetRemaining(ctx context.Context, identifier, routeID string, window time.Duration) (int, int, error)

	// Reset resets the rate limit for an identifier.
	Reset(ctx context.Context, identifier, routeID string) error

	// GetUsage retrieves current usage for an identifier.
	GetUsage(ctx context.Context, identifier, routeID string, window time.Duration) (*domain.RateLimitEntry, error)
}

// AnalyticsRepository defines the interface for analytics data persistence.
type AnalyticsRepository interface {
	// Record records an analytics event.
	Record(ctx context.Context, event *domain.AnalyticsEvent) error

	// Query retrieves analytics events based on filter.
	Query(ctx context.Context, filter AnalyticsFilter) ([]*domain.AnalyticsEvent, error)

	// GetSummary retrieves analytics summary.
	GetSummary(ctx context.Context, startTime, endTime time.Time, groupBy string) (*AnalyticsSummary, error)

	// Cleanup removes old analytics data.
	Cleanup(ctx context.Context, olderThan time.Time) error
}

// AnalyticsFilter represents filtering criteria for analytics queries.
type AnalyticsFilter struct {
	StartTime   *time.Time
	EndTime     *time.Time
	RouteID     string
	ConsumerID  string
	EventType   string
	StatusCode  int
	IPAddress   string
	Limit       int
	Offset      int
}

// AnalyticsSummary represents aggregated analytics data.
type AnalyticsSummary struct {
	TotalRequests   int64             `json:"total_requests"`
	TotalErrors     int64             `json:"total_errors"`
	AverageLatency  float64           `json:"average_latency_ms"`
	RequestsByRoute map[string]int64  `json:"requests_by_route"`
	RequestsByDay   map[string]int64  `json:"requests_by_day"`
	TopErrors       []ErrorSummary    `json:"top_errors"`
}

// ErrorSummary represents error statistics.
type ErrorSummary struct {
	ErrorType   string `json:"error_type"`
	Count       int64  `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// CertificateRepository defines the interface for certificate persistence.
type CertificateRepository interface {
	// Create creates a new certificate.
	Create(ctx context.Context, cert *domain.Certificate) error

	// GetByID retrieves a certificate by ID.
	GetByID(ctx context.Context, id string) (*domain.Certificate, error)

	// GetByDomain retrieves a certificate by domain.
	GetByDomain(ctx context.Context, domain string) (*domain.Certificate, error)

	// Update updates an existing certificate.
	Update(ctx context.Context, cert *domain.Certificate) error

	// List retrieves all active certificates.
	List(ctx context.Context) ([]*domain.Certificate, error)

	// Delete deletes a certificate.
	Delete(ctx context.Context, id string) error
}

// TransformRuleRepository defines the interface for transform rule persistence.
type TransformRuleRepository interface {
	// Create creates a new transform rule.
	Create(ctx context.Context, rule *domain.TransformRule) error

	// GetByID retrieves a transform rule by ID.
	GetByID(ctx context.Context, id string) (*domain.TransformRule, error)

	// GetByRouteID retrieves all transform rules for a route.
	GetByRouteID(ctx context.Context, routeID string) ([]*domain.TransformRule, error)

	// Update updates an existing transform rule.
	Update(ctx context.Context, rule *domain.TransformRule) error

	// Delete deletes a transform rule.
	Delete(ctx context.Context, id string) error
}

// HealthCheckRepository defines the interface for health check results persistence.
type HealthCheckRepository interface {
	// Record records a health check result.
	Record(ctx context.Context, serviceID string, healthy bool, latency int64) error

	// GetLatest retrieves the latest health check result for a service.
	GetLatest(ctx context.Context, serviceID string) (*HealthCheckResult, error)

	// GetHistory retrieves health check history for a service.
	GetHistory(ctx context.Context, serviceID string, limit int) ([]*HealthCheckResult, error)
}

// HealthCheckResult represents a health check result.
type HealthCheckResult struct {
	ServiceID   string    `json:"service_id"`
	Healthy     bool      `json:"healthy"`
	LatencyMs   int64     `json:"latency_ms"`
	StatusCode  int       `json:"status_code,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}
