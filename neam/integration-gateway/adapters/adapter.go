package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"neam-platform/integration-gateway/config"
	"neam-platform/integration-gateway/models"
	"neam-platform/integration-gateway/transformers"
)

// IntegrationAdapter defines the interface for all adapters
type IntegrationAdapter interface {
	// Connect establishes connection to the external system
	Connect(ctx context.Context) error

	// Disconnect closes the connection
	Disconnect(ctx context.Context) error

	// Fetch retrieves data from the external system
	Fetch(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error)

	// Push sends data to the external system
	Push(ctx context.Context, message *models.IntegrationMessage) error

	// HealthCheck checks if the adapter is healthy
	HealthCheck(ctx context.Context) error

	// GetType returns the adapter type
	GetType() string

	// GetName returns the adapter name
	GetName() string
}

// AdapterRegistry manages all registered adapters
type AdapterRegistry struct {
	adapters map[string]IntegrationAdapter
}

// NewAdapterRegistry creates a new adapter registry
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]IntegrationAdapter),
	}
}

// Register registers an adapter
func (r *AdapterRegistry) Register(name string, adapter IntegrationAdapter) {
	r.adapters[name] = adapter
}

// GetAdapter retrieves an adapter by name
func (r *AdapterRegistry) GetAdapter(name string) (IntegrationAdapter, error) {
	adapter, exists := r.adapters[name]
	if !exists {
		return nil, fmt.Errorf("adapter not found: %s", name)
	}
	return adapter, nil
}

// ListAdapters lists all registered adapter names
func (r *AdapterRegistry) ListAdapters() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// HealthCheckAll checks health of all adapters
func (r *AdapterRegistry) HealthCheckAll(ctx context.Context) map[string]error {
	results := make(map[string]error)
	for name, adapter := range r.adapters {
		results[name] = adapter.HealthCheck(ctx)
	}
	return results
}

// BaseAdapter provides common functionality for all adapters
type BaseAdapter struct {
	config     *config.Config
	adapterType string
	adapterName string
	httpClient  *http.Client
	transformer *transformers.DataTransformer
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(cfg *config.Config, adapterType, adapterName string, transformer *transformers.DataTransformer) *BaseAdapter {
	return &BaseAdapter{
		config:      cfg,
		adapterType: adapterType,
		adapterName: adapterName,
		transformer: transformer,
		httpClient: &http.Client{
			Timeout: cfg.GetAdapterTimeout(adapterType),
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost:  10,
				IdleConnTimeout:      90 * time.Second,
				TLSHandshakeTimeout:  10 * time.Second,
			},
		},
	}
}

// GetType returns the adapter type
func (a *BaseAdapter) GetType() string {
	return a.adapterType
}

// GetName returns the adapter name
func (a *BaseAdapter) GetName() string {
	return a.adapterName
}

// HealthCheck performs basic health check
func (a *BaseAdapter) HealthCheck(ctx context.Context) error {
	// Base implementation - can be overridden
	return nil
}

// doRequest performs an HTTP request with retry logic
func (a *BaseAdapter) doRequest(ctx context.Context, method, endpoint string, payload interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	if payload != nil {
		reqBody, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, ioutil.NopCloser(nil))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if reqBody != nil {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "NEAM-Integration-Gateway/1.0")

	// Perform request with retry
	maxRetries := a.getRetryAttempts()
	retryDelay := time.Duration(a.getRetryDelay()) * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := a.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
		}
		return resp, nil
	}

	return nil, fmt.Errorf("request failed: unexpected error")
}

// getRetryAttempts returns retry attempts for the adapter
func (a *BaseAdapter) getRetryAttempts() int {
	switch a.adapterType {
	case "ministry":
		return a.config.Adapters.Ministry.RetryAttempts
	case "central_bank":
		return a.config.Adapters.CentralBank.RetryAttempts
	case "state":
		return a.config.Adapters.State.RetryAttempts
	case "district":
		return a.config.Adapters.District.RetryAttempts
	case "legacy":
		return a.config.Adapters.Legacy.RetryAttempts
	default:
		return 3
	}
}

// getRetryDelay returns retry delay in seconds
func (a *BaseAdapter) getRetryDelay() int {
	switch a.adapterType {
	case "ministry":
		return a.config.Adapters.Ministry.RetryDelay
	case "central_bank":
		return a.config.Adapters.CentralBank.RetryDelay
	case "state":
		return a.config.Adapters.State.RetryDelay
	case "district":
		return a.config.Adapters.District.RetryDelay
	case "legacy":
		return a.config.Adapters.Legacy.RetryDelay
	default:
		return 1
	}
}

// getEndpoint returns the endpoint URL for the adapter
func (a *BaseAdapter) getEndpoint(path string) string {
	var baseURL string
	switch a.adapterType {
	case "ministry":
		baseURL = a.config.Adapters.Ministry.Endpoint
	case "central_bank":
		baseURL = a.config.Adapters.CentralBank.Endpoint
	case "state":
		baseURL = a.config.Adapters.State.Endpoint
	case "district":
		baseURL = a.config.Adapters.District.Endpoint
	case "legacy":
		baseURL = a.config.Adapters.Legacy.Endpoint
	default:
		baseURL = ""
	}

	if baseURL == "" {
		return path
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return path
	}

	u.Path = u.Path + path
	return u.String()
}

// parseResponse parses HTTP response into adapter response
func (a *BaseAdapter) parseResponse(resp *http.Response) (*models.AdapterResponse, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &models.AdapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       body,
		Timestamp:  time.Now(),
	}

	// Copy response headers
	for k, v := range resp.Header {
		if len(v) > 0 {
			response.Headers[k] = v[0]
		}
	}

	return response, nil
}

// BuildRequestURL builds a complete request URL with query parameters
func BuildRequestURL(baseURL string, path string, params map[string]string) string {
	u, _ := url.Parse(baseURL)
	u.Path = path

	if params != nil {
		query := u.Query()
		for k, v := range params {
			query.Set(k, v)
		}
		u.RawQuery = query.Encode()
	}

	return u.String()
}
