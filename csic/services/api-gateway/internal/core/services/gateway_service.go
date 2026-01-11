package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/api-gateway/gateway/internal/core/domain"
	"github.com/api-gateway/gateway/internal/core/ports"
)

// Common errors for gateway operations.
var (
	ErrRouteNotFound       = errors.New("route not found")
	ErrRouteAlreadyExists  = errors.New("route already exists")
	ErrConsumerNotFound    = errors.New("consumer not found")
	ErrConsumerAlreadyExists = errors.New("consumer already exists")
	ErrInvalidAPIKey       = errors.New("invalid API key")
	ErrAPIKeyExpired       = errors.New("API key has expired")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrQuotaExceeded       = errors.New("quota exceeded")
	ErrServiceNotFound     = errors.New("service not found")
	ErrInvalidRoutePath    = errors.New("invalid route path")
	ErrRouteNotActive      = errors.New("route is not active")
)

// GatewayService provides the core business logic for API gateway operations.
type GatewayService struct {
	routeRepo      ports.RouteRepository
	consumerRepo   ports.ConsumerRepository
	serviceRepo    ports.ServiceRepository
	rateLimitRepo  ports.RateLimitRepository
	analyticsRepo  ports.AnalyticsRepository
	certRepo       ports.CertificateRepository
	transformRepo  ports.TransformRuleRepository
	healthRepo     ports.HealthCheckRepository

	// In-memory cache for routes (for performance)
	routeCache      map[string]*domain.Route
	routeCacheMutex sync.RWMutex
}

// NewGatewayService creates a new GatewayService with the required dependencies.
func NewGatewayService(
	routeRepo ports.RouteRepository,
	consumerRepo ports.ConsumerRepository,
	serviceRepo ports.ServiceRepository,
	rateLimitRepo ports.RateLimitRepository,
	analyticsRepo ports.AnalyticsRepository,
	certRepo ports.CertificateRepository,
	transformRepo ports.TransformRuleRepository,
	healthRepo ports.HealthCheckRepository,
) *GatewayService {
	return &GatewayService{
		routeRepo:     routeRepo,
		consumerRepo:  consumerRepo,
		serviceRepo:   serviceRepo,
		rateLimitRepo: rateLimitRepo,
		analyticsRepo: analyticsRepo,
		certRepo:      certRepo,
		transformRepo: transformRepo,
		healthRepo:    healthRepo,
		routeCache:    make(map[string]*domain.Route),
	}
}

// ==================== Route Management ====================

// CreateRoute creates a new API route.
func (s *GatewayService) CreateRoute(
	ctx context.Context,
	name, path string,
	methods []string,
	upstreamURL, upstreamPath string,
	timeout, retryCount int,
) (*domain.Route, error) {
	// Validate path
	if err := s.validateRoutePath(path); err != nil {
		return nil, err
	}

	// Check for existing route
	existing, err := s.routeRepo.GetByPath(ctx, path, methods[0])
	if err != nil {
		return nil, fmt.Errorf("failed to check existing route: %w", err)
	}
	if existing != nil {
		return nil, ErrRouteAlreadyExists
	}

	route := &domain.Route{
		ID:           domain.NewEntityID(),
		Name:         name,
		Path:         path,
		Methods:      methods,
		UpstreamURL:  upstreamURL,
		UpstreamPath: upstreamPath,
		Timeout:      timeout,
		RetryCount:   retryCount,
		IsActive:     true,
		Plugins:      []domain.PluginConfig{},
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.routeRepo.Create(ctx, route); err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}

	// Update cache
	s.updateRouteCache(route)

	return route, nil
}

// GetRoute retrieves a route by path and method.
func (s *GatewayService) GetRoute(ctx context.Context, path, method string) (*domain.Route, error) {
	// Check cache first
	s.routeCacheMutex.RLock()
	if route, ok := s.routeCache[s.buildCacheKey(path, method)]; ok && route.IsActive {
		s.routeCacheMutex.RUnlock()
		return route, nil
	}
	s.routeCacheMutex.RUnlock()

	// Fetch from database
	route, err := s.routeRepo.GetByPath(ctx, path, method)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	if route == nil || !route.IsActive {
		return nil, ErrRouteNotFound
	}

	return route, nil
}

// ListRoutes retrieves all active routes.
func (s *GatewayService) ListRoutes(ctx context.Context) ([]*domain.Route, error) {
	return s.routeRepo.List(ctx)
}

// UpdateRoute updates an existing route.
func (s *GatewayService) UpdateRoute(ctx context.Context, route *domain.Route) error {
	route.UpdatedAt = time.Now().UTC()

	if err := s.routeRepo.Update(ctx, route); err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	// Update cache
	s.updateRouteCache(route)

	return nil
}

// DeleteRoute soft-deletes a route.
func (s *GatewayService) DeleteRoute(ctx context.Context, id string) error {
	if err := s.routeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	// Remove from cache
	s.removeFromCache(id)

	return nil
}

// InvalidateCache invalidates the route cache.
func (s *GatewayService) InvalidateCache() {
	s.routeCacheMutex.Lock()
	s.routeCache = make(map[string]*domain.Route)
	s.routeCacheMutex.Unlock()
}

// ==================== Consumer Management ====================

// CreateConsumer creates a new API consumer.
func (s *GatewayService) CreateConsumer(
	ctx context.Context,
	username string,
	groups []string,
	quota *domain.QuotaConfig,
) (*domain.Consumer, error) {
	// Check for existing consumer
	existing, err := s.consumerRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing consumer: %w", err)
	}
	if existing != nil {
		return nil, ErrConsumerAlreadyExists
	}

	consumer := &domain.Consumer{
		ID:        domain.NewEntityID(),
		Username:  username,
		Groups:    groups,
		APIKeys:   []domain.APIKey{},
		Quota:     quota,
		IsActive:  true,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.consumerRepo.Create(ctx, consumer); err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return consumer, nil
}

// GetConsumerByAPIKey retrieves a consumer by API key.
func (s *GatewayService) GetConsumerByAPIKey(ctx context.Context, apiKey string) (*domain.Consumer, error) {
	consumer, err := s.consumerRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}
	if consumer == nil {
		return nil, ErrInvalidAPIKey
	}
	if !consumer.IsActive {
		return nil, errors.New("consumer is not active")
	}

	return consumer, nil
}

// ValidateAPIKey validates an API key and returns consumer info.
func (s *GatewayService) ValidateAPIKey(ctx context.Context, apiKey string) (*domain.Consumer, error) {
	consumer, err := s.GetConsumerByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// Find the specific API key
	var validKey *domain.APIKey
	for i := range consumer.APIKeys {
		if consumer.APIKeys[i].Key == apiKey {
			validKey = &consumer.APIKeys[i]
			break
		}
	}

	if validKey == nil {
		return nil, ErrInvalidAPIKey
	}

	// Check expiration
	if validKey.ExpiresAt != nil && validKey.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrAPIKeyExpired
	}

	if !validKey.IsActive {
		return nil, errors.New("API key is not active")
	}

	return consumer, nil
}

// GenerateAPIKey generates a new API key for a consumer.
func (s *GatewayService) GenerateAPIKey(ctx context.Context, consumerID, keyName string, scopes []string, expiresAt *time.Time) (*domain.APIKey, error) {
	consumer, err := s.consumerRepo.GetByID(ctx, consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}
	if consumer == nil {
		return nil, ErrConsumerNotFound
	}

	key := &domain.APIKey{
		ID:        domain.NewEntityID(),
		Key:       s.generateSecureAPIKey(),
		Name:      keyName,
		Scopes:    scopes,
		ExpiresAt: expiresAt,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}

	consumer.APIKeys = append(consumer.APIKeys, *key)
	consumer.UpdatedAt = time.Now().UTC()

	if err := s.consumerRepo.Update(ctx, consumer); err != nil {
		return nil, fmt.Errorf("failed to update consumer: %w", err)
	}

	return key, nil
}

// ==================== Rate Limiting ====================

// CheckRateLimit checks if a request exceeds rate limits.
func (s *GatewayService) CheckRateLimit(
	ctx context.Context,
	identifier, routeID string,
	route *domain.Route,
) (bool, int, int, error) {
	// If no rate limit configured, allow request
	if route.RateLimit == nil {
		return true, 0, 0, nil
	}

	window := time.Minute // Use 1-minute window for per-minute limits
	limit := route.RateLimit.RequestsPerMinute

	requests, err := s.rateLimitRepo.Increment(ctx, identifier, routeID, window)
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to check rate limit: %w", err)
	}

	remaining := limit - requests
	if remaining < 0 {
		remaining = 0
	}

	if requests > limit {
		return false, limit, remaining, ErrRateLimitExceeded
	}

	return true, limit, remaining, nil
}

// ==================== Analytics ====================

// RecordRequest records an analytics event for a request.
func (s *GatewayService) RecordRequest(ctx context.Context, event *domain.AnalyticsEvent) error {
	return s.analyticsRepo.Record(ctx, event)
}

// GetAnalyticsSummary retrieves analytics summary.
func (s *GatewayService) GetAnalyticsSummary(
	ctx context.Context,
	startTime, endTime time.Time,
	groupBy string,
) (*ports.AnalyticsSummary, error) {
	return s.analyticsRepo.GetSummary(ctx, startTime, endTime, groupBy)
}

// ==================== Service Management ====================

// CreateService creates a new upstream service.
func (s *GatewayService) CreateService(
	ctx context.Context,
	name, host string,
	port int,
	protocol string,
	healthCheck *domain.HealthCheck,
) (*domain.Service, error) {
	service := &domain.Service{
		ID:          domain.NewEntityID(),
		Name:        name,
		Host:        host,
		Port:        port,
		Protocol:    protocol,
		HealthCheck: healthCheck,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.serviceRepo.Create(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return service, nil
}

// GetService retrieves a service by name.
func (s *GatewayService) GetService(ctx context.Context, name string) (*domain.Service, error) {
	service, err := s.serviceRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	if service == nil {
		return nil, ErrServiceNotFound
	}

	return service, nil
}

// ListServices retrieves all active services.
func (s *GatewayService) ListServices(ctx context.Context) ([]*domain.Service, error) {
	return s.serviceRepo.List(ctx)
}

// ==================== Transform Rules ====================

// CreateTransformRule creates a new transform rule.
func (s *GatewayService) CreateTransformRule(
	ctx context.Context,
	name, routeID, transformType, action, target, value string,
) (*domain.TransformRule, error) {
	rule := &domain.TransformRule{
		ID:       domain.NewEntityID(),
		Name:     name,
		RouteID:  routeID,
		Type:     transformType,
		Action:   action,
		Target:   target,
		Value:    value,
		IsActive: true,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.transformRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create transform rule: %w", err)
	}

	return rule, nil
}

// GetTransformRules retrieves transform rules for a route.
func (s *GatewayService) GetTransformRules(ctx context.Context, routeID string) ([]*domain.TransformRule, error) {
	return s.transformRepo.GetByRouteID(ctx, routeID)
}

// ==================== Helper Methods ====================

// validateRoutePath validates a route path.
func (s *GatewayService) validateRoutePath(path string) error {
	if len(path) == 0 || path[0] != '/' {
		return ErrInvalidRoutePath
	}

	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[?#\[\]{}<>\\^|]`)
	if invalidChars.MatchString(path) {
		return ErrInvalidRoutePath
	}

	return nil
}

// buildCacheKey builds a cache key for a route.
func (s *GatewayService) buildCacheKey(path, method string) string {
	return strings.ToLower(method) + ":" + path
}

// updateRouteCache updates the route cache.
func (s *GatewayService) updateRouteCache(route *domain.Route) {
	s.routeCacheMutex.Lock()
	defer s.routeCacheMutex.Unlock()

	for _, method := range route.Methods {
		s.routeCache[s.buildCacheKey(route.Path, method)] = route
	}
}

// removeFromCache removes a route from the cache.
func (s *GatewayService) removeFromCache(routeID string) {
	s.routeCacheMutex.Lock()
	defer s.routeCacheMutex.Unlock()

	for key, route := range s.routeCache {
		if string(route.ID) == routeID {
			delete(s.routeCache, key)
		}
	}
}

// generateSecureAPIKey generates a secure random API key.
func (s *GatewayService) generateSecureAPIKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 32

	var result strings.Builder
	for i := 0; i < keyLength; i++ {
		result.WriteByte(charset[time.Now().UnixNano()%int64(len(charset))])
		time.Sleep(time.Nanosecond)
	}

	return "gw_" + result.String()
}
