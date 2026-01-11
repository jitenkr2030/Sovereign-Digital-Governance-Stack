package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"neam-platform/integration-gateway/config"
	"neam-platform/integration-gateway/models"
	"neam-platform/integration-gateway/transformers"
)

// MinistryAdapter handles integration with government ministry systems
type MinistryAdapter struct {
	*BaseAdapter
	cfg *config.Config
}

// NewMinistryAdapter creates a new MinistryAdapter
func NewMinistryAdapter(cfg *config.Config) *MinistryAdapter {
	transformer := transformers.NewDataTransformer()
	return &MinistryAdapter{
		BaseAdapter: NewBaseAdapter(cfg, "ministry", "ministry-adapter", transformer),
		cfg:         cfg,
	}
}

// Connect establishes connection to ministry systems
func (a *MinistryAdapter) Connect(ctx context.Context) error {
	// Ministry adapters typically use REST APIs, so connection is established per request
	// Perform a health check to verify connectivity
	return a.HealthCheck(ctx)
}

// Disconnect closes the connection
func (a *MinistryAdapter) Disconnect(ctx context.Context) error {
	// No persistent connection to close
	return nil
}

// Fetch retrieves data from ministry systems
func (a *MinistryAdapter) Fetch(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	// Build request URL
	url := a.getEndpoint(request.Endpoint)

	// Add query parameters if provided
	if request.Params != nil {
		url = BuildRequestURL(a.config.Adapters.Ministry.Endpoint, request.Endpoint, request.Params)
	}

	// Perform HTTP request
	resp, err := a.doRequest(ctx, request.Method, url, request.Payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	response, err := a.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// Transform data if needed
	if request.Transform {
		transformedData, err := a.transformer.TransformMinistryData(response.Body, request.DataType)
		if err != nil {
			return nil, fmt.Errorf("data transformation failed: %w", err)
		}
		response.Body = transformedData
	}

	return response, nil
}

// Push sends data to ministry systems
func (a *MinistryAdapter) Push(ctx context.Context, message *models.IntegrationMessage) error {
	// Determine endpoint based on message type
	endpoint := a.getEndpointForMessage(message)

	// Create request
	request := &models.AdapterRequest{
		Method:  http.MethodPost,
		Endpoint: endpoint,
		Payload: message.Payload,
		DataType: message.DataType,
	}

	// Send request
	resp, err := a.Fetch(ctx, request)
	if err != nil {
		return err
	}

	// Check response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ministry push failed with status: %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs health check on ministry connection
func (a *MinistryAdapter) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to reach the ministry endpoint
	healthURL := BuildRequestURL(a.config.Adapters.Ministry.Endpoint, "/health", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("health check request creation failed: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ministry system unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

// getEndpointForMessage returns the appropriate endpoint for a message type
func (a *MinistryAdapter) getEndpointForMessage(message *models.IntegrationMessage) string {
	switch message.DataType {
	case "citizen":
		return "/api/v1/citizens"
	case "economic":
		return "/api/v1/economic-data"
	case "policy":
		return "/api/v1/policies"
	case "report":
		return "/api/v1/reports"
	default:
		return "/api/v1/data"
	}
}

// SupportedDataTypes returns the data types supported by this adapter
func (a *MinistryAdapter) SupportedDataTypes() []string {
	return []string{
		"citizen",
		"demographic",
		"economic",
		"policy",
		"report",
		"notification",
	}
}

// MinistryAdapter implements IntegrationAdapter
var _ IntegrationAdapter = (*MinistryAdapter)(nil)

// CentralBankAdapter handles integration with central bank systems
type CentralBankAdapter struct {
	*BaseAdapter
	cfg *config.Config
}

// NewCentralBankAdapter creates a new CentralBankAdapter
func NewCentralBankAdapter(cfg *config.Config) *CentralBankAdapter {
	transformer := transformers.NewDataTransformer()
	return &CentralBankAdapter{
		BaseAdapter: NewBaseAdapter(cfg, "central_bank", "central-bank-adapter", transformer),
		cfg:         cfg,
	}
}

// Connect establishes connection to central bank systems
func (a *CentralBankAdapter) Connect(ctx context.Context) error {
	// Add API key header for central bank authentication
	// Connection is established per request
	return a.HealthCheck(ctx)
}

// Disconnect closes the connection
func (a *CentralBankAdapter) Disconnect(ctx context.Context) error {
	return nil
}

// Fetch retrieves financial data from central bank
func (a *CentralBankAdapter) Fetch(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	url := a.getEndpoint(request.Endpoint)

	// Add authentication header
	ctx = context.WithValue(ctx, "api_key", a.config.Adapters.CentralBank.APIKey)

	resp, err := a.doRequest(ctx, request.Method, url, request.Payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := a.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// Transform data for financial formats
	if request.Transform {
		transformedData, err := a.transformer.TransformFinancialData(response.Body, request.DataType)
		if err != nil {
			return nil, fmt.Errorf("financial data transformation failed: %w", err)
		}
		response.Body = transformedData
	}

	return response, nil
}

// Push sends data to central bank
func (a *CentralBankAdapter) Push(ctx context.Context, message *models.IntegrationMessage) error {
	endpoint := a.getEndpointForMessage(message)

	request := &models.AdapterRequest{
		Method:  http.MethodPost,
		Endpoint: endpoint,
		Payload: message.Payload,
		DataType: message.DataType,
	}

	resp, err := a.Fetch(ctx, request)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("central bank push failed with status: %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs health check on central bank connection
func (a *CentralBankAdapter) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthURL := BuildRequestURL(a.config.Adapters.CentralBank.Endpoint, "/health", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return err
	}

	// Add API key header
	req.Header.Set("X-API-Key", a.config.Adapters.CentralBank.APIKey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("central bank health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("central bank system unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

// getEndpointForMessage returns the appropriate endpoint for financial messages
func (a *CentralBankAdapter) getEndpointForMessage(message *models.IntegrationMessage) string {
	switch message.DataType {
	case "transaction":
		return "/api/v1/transactions"
	case "account":
		return "/api/v1/accounts"
	case "credit":
		return "/api/v1/credit-reports"
	case "monetary":
		return "/api/v1/monetary-data"
	case "regulatory":
		return "/api/v1/regulatory-reports"
	default:
		return "/api/v1/financial-data"
	}
}

// SupportedDataTypes returns the data types supported by this adapter
func (a *CentralBankAdapter) SupportedDataTypes() []string {
	return []string{
		"transaction",
		"account",
		"credit",
		"monetary",
		"regulatory",
		"compliance",
	}
}

// CentralBankAdapter implements IntegrationAdapter
var _ IntegrationAdapter = (*CentralBankAdapter)(nil)
