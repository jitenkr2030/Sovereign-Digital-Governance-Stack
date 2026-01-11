package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// OpenSearchClient provides OpenSearch integration for SIEM
type OpenSearchClient interface {
	// IndexEvent indexes an audit event in OpenSearch
	IndexEvent(ctx context.Context, event interface{}) error

	// SearchEvents searches for audit events
	SearchEvents(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, error)

	// GetEvent retrieves a specific event
	GetEvent(ctx context.Context, index, id string) (map[string]interface{}, error)

	// DeleteEvent deletes an event
	DeleteEvent(ctx context.Context, index, id string) error

	// HealthCheck checks OpenSearch health
	HealthCheck(ctx context.Context) error
}

// opensearchClient implements OpenSearchClient
type opensearchClient struct {
	baseURL    string
	index      string
	httpClient *http.Client
	username   string
	password   string
}

// NewOpenSearchClient creates a new OpenSearchClient
func NewOpenSearchClient(baseURL, index string) (*opensearchClient, error) {
	client := &opensearchClient{
		baseURL:    baseURL,
		index:      index,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}

	return client, nil
}

// WithCredentials creates an OpenSearchClient with credentials
func NewOpenSearchClientWithCredentials(baseURL, index, username, password string) (*opensearchClient, error) {
	client := &opensearchClient{
		baseURL:    baseURL,
		index:      index,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		username:   username,
		password:   password,
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}

	return client, nil
}

// IndexEvent indexes an audit event in OpenSearch
func (c *opensearchClient) IndexEvent(ctx context.Context, event interface{}) error {
	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create index if it doesn't exist
	if err := c.ensureIndex(ctx); err != nil {
		return fmt.Errorf("failed to ensure index: %w", err)
	}

	// Index the event
	url := fmt.Sprintf("%s/%s/_doc", c.baseURL, c.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to index event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to index event: %s", string(body))
	}

	return nil
}

// SearchEvents searches for audit events
func (c *opensearchClient) SearchEvents(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, error) {
	// Build search query
	searchQuery := map[string]interface{}{
		"query": query,
		"size":  100,
		"sort": []map[string]interface{}{
			{"@timestamp": map[string]string{"order": "desc"}},
		},
	}

	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	url := fmt.Sprintf("%s/%s/_search", c.baseURL, c.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(body))
	}

	// Parse response
	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	events := make([]map[string]interface{}, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		events[i] = hit.Source
	}

	return events, nil
}

// GetEvent retrieves a specific event
func (c *opensearchClient) GetEvent(ctx context.Context, index, id string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/_doc/%s", c.baseURL, index, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("event not found")
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get event: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// DeleteEvent deletes an event
func (c *opensearchClient) DeleteEvent(ctx context.Context, index, id string) error {
	url := fmt.Sprintf("%s/%s/_doc/%s", c.baseURL, index, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusNotFound {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete event: %s", string(body))
	}

	return nil
}

// HealthCheck checks OpenSearch health
func (c *opensearchClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/_cluster/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("OpenSearch not healthy: status %d", resp.StatusCode)
	}

	return nil
}

// ensureIndex creates the index if it doesn't exist
func (c *opensearchClient) ensureIndex(ctx context.Context) error {
	// Check if index exists
	checkURL := fmt.Sprintf("%s/%s", c.baseURL, c.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, checkURL, nil)
	if err != nil {
		return err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // Index exists
	}

	// Create index with mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"@timestamp":    map[string]string{"type": "date"},
				"event_type":   map[string]string{"type": "keyword"},
				"actor_id":     map[string]string{"type": "keyword"},
				"actor_type":   map[string]string{"type": "keyword"},
				"action":       map[string]string{"type": "keyword"},
				"resource":     map[string]string{"type": "keyword"},
				"resource_id":  map[string]string{"type": "keyword"},
				"outcome":      map[string]string{"type": "keyword"},
				"ip_address":   map[string]string{"type": "ip"},
				"location":     map[string]string{"type": "geo_point"},
				"session_id":   map[string]string{"type": "keyword"},
				"integrity_hash": map[string]string{"type": "keyword"},
				"sealed":       map[string]string{"type": "boolean"},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPut, checkURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	c.setHeaders(createReq)
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := c.httpClient.Do(createReq)
	if err != nil {
		return err
	}
	defer createResp.Body.Close()

	if createResp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(createResp.Body)
		return fmt.Errorf("failed to create index: %s", string(body))
	}

	log.Printf("Created OpenSearch index: %s", c.index)
	return nil
}

// setHeaders sets authentication and content headers
func (c *opensearchClient) setHeaders(req *http.Request) {
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}

// SearchResponse represents OpenSearch search response
type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string                 `json:"_id"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// CreateIndexForAudit creates an index for audit logs with proper settings
func CreateIndexForAudit(client *opensearchClient, indexName string) error {
	ctx := context.Background()

	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"@timestamp":       map[string]string{"type": "date"},
				"event_type":      map[string]string{"type": "keyword"},
				"actor_id":        map[string]string{"type": "keyword"},
				"actor_type":      map[string]string{"type": "keyword"},
				"actor_name":      map[string]string{"type": "text"},
				"action":          map[string]string{"type": "keyword"},
				"resource":        map[string]string{"type": "keyword"},
				"resource_id":     map[string]string{"type": "keyword"},
				"resource_type":   map[string]string{"type": "keyword"},
				"outcome":         map[string]string{"type": "keyword"},
				"details":         map[string]string{"type": "object", "enabled": true},
				"ip_address":      map[string]string{"type": "ip"},
				"user_agent":      map[string]string{"type": "text"},
				"location":        map[string]string{"type": "geo_point"},
				"session_id":      map[string]string{"type": "keyword"},
				"request_id":      map[string]string{"type": "keyword"},
				"trace_id":        map[string]string{"type": "keyword"},
				"integrity_hash":  map[string]string{"type": "keyword"},
				"previous_hash":   map[string]string{"type": "keyword"},
				"chain_index":     map[string]string{"type": "long"},
				"sealed":          map[string]string{"type": "boolean"},
				"compliance_tags": map[string]string{"type": "keyword"},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   3,
			"number_of_replicas": 1,
			"refresh_interval":  "1s",
		},
	}

	data, _ := json.Marshal(mapping)

	url := fmt.Sprintf("%s/%s", client.baseURL, indexName)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Warning: Failed to create index: %s", string(body))
	}

	return nil
}

// RetentionPolicy defines a retention policy for audit logs
type RetentionPolicy struct {
	Name        string        `json:"name"`
	IndexPattern string       `json:"index_pattern"`
	Retention   time.Duration `json:"retention"`
	Enabled     bool          `json:"enabled"`
}

// ApplyRetentionPolicy applies a retention policy to audit logs
func ApplyRetentionPolicy(client *opensearchClient, policy RetentionPolicy) error {
	ctx := context.Background()

	// Calculate cutoff date
	cutoff := time.Now().Add(-policy.Retention)

	// Build query to find indices older than cutoff
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"lt": cutoff.Format(time.RFC3339),
				},
			},
		},
	}

	data, _ := json.Marshal(query)

	url := fmt.Sprintf("%s/%s/_delete_by_query", client.baseURL, policy.IndexPattern)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	client.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to apply retention policy: %s", string(body))
	}

	log.Printf("Applied retention policy: %s", policy.Name)
	return nil
}
