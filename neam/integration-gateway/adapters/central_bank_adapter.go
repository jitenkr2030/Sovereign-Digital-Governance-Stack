package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"integration-gateway/models"
)

// CentralBankAdapter handles integration with the Central Bank of Egypt
type CentralBankAdapter struct {
	baseURL       string
	apiKey        string
	httpClient    *http.Client
	timeout       time.Duration
	certificate   []byte
	privateKey    []byte
}

// NewCentralBankAdapter creates a new Central Bank adapter
func NewCentralBankAdapter(config map[string]string) (*CentralBankAdapter, error) {
	baseURL, ok := config["base_url"]
	if !ok || baseURL == "" {
		return nil, fmt.Errorf("central bank adapter requires 'base_url' configuration")
	}

	adapter := &CentralBankAdapter{
		baseURL:     baseURL,
		apiKey:      config["api_key"],
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		timeout:     30 * time.Second,
		certificate: []byte(config["certificate"]),
		privateKey:  []byte(config["private_key"]),
	}

	return adapter, nil
}

// GetSystemName returns the system name for this adapter
func (a *CentralBankAdapter) GetSystemName() string {
	return "central-bank"
}

// FetchData fetches data from Central Bank systems
func (a *CentralBankAdapter) FetchData(ctx context.Context, req *models.IntegrationRequest) ([]interface{}, error) {
	// Build the request based on request parameters
	url := a.buildRequestURL(req)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication headers
	a.addAuthHeaders(httpReq)

	// Execute the request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("central bank API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract data array from response
	results, ok := responseData["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from Central Bank")
	}

	log.Printf("Fetched %d records from Central Bank", len(results))
	return results, nil
}

// SendData sends data to Central Bank systems
func (a *CentralBankAdapter) SendData(ctx context.Context, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/submissions", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewBuffer(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	a.addAuthHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("central bank API error: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully sent data to Central Bank")
	return nil
}

// Validate validates the configuration for this adapter
func (a *CentralBankAdapter) Validate() error {
	if a.baseURL == "" {
		return fmt.Errorf("base_url is required for Central Bank adapter")
	}
	return nil
}

// buildRequestURL constructs the appropriate URL for the request
func (a *CentralBankAdapter) buildRequestURL(req *models.IntegrationRequest) string {
	endpoint := "transactions"
	params := req.Parameters

	if endpointParam, ok := params["endpoint"].(string); ok {
		endpoint = endpointParam
	}

	url := fmt.Sprintf("%s/api/v1/%s", a.baseURL, endpoint)

	// Add query parameters
	if len(req.Parameters) > 0 {
		url += "?"
		first := true
		for key, value := range req.Parameters {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%v", key, value)
			first = false
		}
	}

	return url
}

// addAuthHeaders adds authentication headers to requests
func (a *CentralBankAdapter) addAuthHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("X-Request-ID", generateRequestID())
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
