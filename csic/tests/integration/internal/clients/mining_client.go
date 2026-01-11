package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// MiningClient for mining control service
type MiningClient struct {
	baseURL string
	client  *http.Client
}

// NewMiningClient creates a new mining service client
func NewMiningClient(ctx context.Context, baseURL string) (*MiningClient, error) {
	return &MiningClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// RegisterOperationRequest for registering mining operations
type RegisterOperationRequest struct {
	OperatorName   string  `json:"operator_name"`
	WalletAddress  string  `json:"wallet_address"`
	Location       string  `json:"location"`
	Region         string  `json:"region"`
	MachineType    string  `json:"machine_type"`
	InitialHashrate float64 `json:"initial_hashrate,omitempty"`
}

// MiningOperation response
type MiningOperation struct {
	ID              uuid.UUID `json:"id"`
	OperatorName    string    `json:"operator_name"`
	WalletAddress   string    `json:"wallet_address"`
	CurrentHashrate float64   `json:"current_hashrate"`
	Status          string    `json:"status"`
	Location        string    `json:"location"`
	Region          string    `json:"region"`
	MachineType     string    `json:"machine_type"`
	RegisteredAt    time.Time `json:"registered_at"`
}

// RegisterOperation registers a new mining operation
func (c *MiningClient) RegisterOperation(ctx context.Context, req *RegisterOperationRequest) (*MiningOperation, error) {
	url := fmt.Sprintf("%s/api/v1/operations", c.baseURL)
	
	body, _ := json.Marshal(req)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to register operation: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to register operation: %s", string(respBody))
	}
	
	var result struct {
		Operation *MiningOperation `json:"operation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Operation, nil
}

// GetOperation retrieves a mining operation
func (c *MiningClient) GetOperation(ctx context.Context, id uuid.UUID) (*MiningOperation, error) {
	url := fmt.Sprintf("%s/api/v1/operations/%s", c.baseURL, id.String())
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("operation not found")
	}
	
	var op MiningOperation
	if err := json.NewDecoder(resp.Body).Decode(&op); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &op, nil
}

// ReportHashrate reports hashrate telemetry
func (c *MiningClient) ReportHashrate(ctx context.Context, opID uuid.UUID, hashrate float64, unit string) error {
	url := fmt.Sprintf("%s/api/v1/operations/%s/telemetry", c.baseURL, opID.String())
	
	reqBody := map[string]interface{}{
		"operation_id": opID.String(),
		"hashrate":     hashrate,
		"unit":         unit,
	}
	body, _ := json.Marshal(reqBody)
	
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to report hashrate: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to report hashrate: %s", string(respBody))
	}
	
	return nil
}

// ListOperations lists all operations
func (c *MiningClient) ListOperations(ctx context.Context, status string, page, pageSize int) ([]MiningOperation, error) {
	url := fmt.Sprintf("%s/api/v1/operations?page=%d&page_size=%d", c.baseURL, page, pageSize)
	if status != "" {
		url += fmt.Sprintf("&status=%s", status)
	}
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}
	defer resp.Body.Close()
	
	var result struct {
		Operations []MiningOperation `json:"operations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Operations, nil
}
