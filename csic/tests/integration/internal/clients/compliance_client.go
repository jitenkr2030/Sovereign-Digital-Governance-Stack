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

// ComplianceClient for compliance service
type ComplianceClient struct {
	baseURL string
	client  *http.Client
}

// NewComplianceClient creates a new compliance service client
func NewComplianceClient(ctx context.Context, baseURL string) (*ComplianceClient, error) {
	return &ComplianceClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// RegisterEntityRequest for registering entities
type RegisterEntityRequest struct {
	Name            string `json:"name"`
	LegalName       string `json:"legal_name"`
	RegistrationNum string `json:"registration_num"`
	Jurisdiction    string `json:"jurisdiction"`
	EntityType      string `json:"entity_type"`
	Address         string `json:"address"`
	ContactEmail    string `json:"contact_email"`
	RiskLevel       string `json:"risk_level"`
}

// Entity response
type Entity struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	LegalName       string    `json:"legal_name"`
	RegistrationNum string    `json:"registration_num"`
	Jurisdiction    string    `json:"jurisdiction"`
	EntityType      string    `json:"entity_type"`
	Status          string    `json:"status"`
	RiskLevel       string    `json:"risk_level"`
	CreatedAt       time.Time `json:"created_at"`
}

// IssueLicenseRequest for issuing licenses
type IssueLicenseRequest struct {
	EntityID      uuid.UUID `json:"entity_id"`
	LicenseType   string    `json:"license_type"`
	LicenseNumber string    `json:"license_number"`
	ExpiryDays    int       `json:"expiry_days,omitempty"`
	Conditions    string    `json:"conditions,omitempty"`
	Jurisdiction  string    `json:"jurisdiction"`
	IssuedBy      string    `json:"issued_by"`
}

// License response
type License struct {
	ID            uuid.UUID `json:"id"`
	EntityID      uuid.UUID `json:"entity_id"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	LicenseNumber string    `json:"license_number"`
	IssuedDate    time.Time `json:"issued_date"`
	ExpiryDate    time.Time `json:"expiry_date"`
}

// ComplianceScore response
type ComplianceScore struct {
	ID            uuid.UUID `json:"id"`
	EntityID      uuid.UUID `json:"entity_id"`
	TotalScore    float64   `json:"total_score"`
	Tier          string    `json:"tier"`
	Breakdown     string    `json:"breakdown"`
	CalculatedAt  time.Time `json:"calculated_at"`
}

// RegisterEntity registers a new entity
func (c *ComplianceClient) RegisterEntity(ctx context.Context, req *RegisterEntityRequest) (*Entity, error) {
	url := fmt.Sprintf("%s/api/v1/entities", c.baseURL)
	
	body, _ := json.Marshal(req)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to register entity: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to register entity: %s", string(respBody))
	}
	
	var result struct {
		Entity *Entity `json:"entity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Entity, nil
}

// IssueLicense issues a new license
func (c *ComplianceClient) IssueLicense(ctx context.Context, req *IssueLicenseRequest) (*License, error) {
	url := fmt.Sprintf("%s/api/v1/licenses", c.baseURL)
	
	body, _ := json.Marshal(req)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to issue license: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to issue license: %s", string(respBody))
	}
	
	var result struct {
		License *License `json:"license"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.License, nil
}

// CalculateScore calculates compliance score for an entity
func (c *ComplianceClient) CalculateScore(ctx context.Context, entityID uuid.UUID) (*ComplianceScore, error) {
	url := fmt.Sprintf("%s/api/v1/entities/%s/compliance/score/recalculate", c.baseURL, entityID.String())
	
	resp, err := c.client.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate score: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to calculate score: %s", string(respBody))
	}
	
	var result struct {
		Score *ComplianceScore `json:"score"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.Score, nil
}

// GetComplianceScore retrieves current compliance score
func (c *ComplianceClient) GetComplianceScore(ctx context.Context, entityID uuid.UUID) (*ComplianceScore, error) {
	url := fmt.Sprintf("%s/api/v1/entities/%s/compliance/score", c.baseURL, entityID.String())
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get score: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("score not found")
	}
	
	var score ComplianceScore
	if err := json.NewDecoder(resp.Body).Decode(&score); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &score, nil
}

// GetEntity retrieves an entity
func (c *ComplianceClient) GetEntity(ctx context.Context, entityID uuid.UUID) (*Entity, error) {
	url := fmt.Sprintf("%s/api/v1/entities/%s", c.baseURL, entityID.String())
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("entity not found")
	}
	
	var entity Entity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &entity, nil
}

// GetComplianceStats retrieves overall compliance statistics
func (c *ComplianceClient) GetComplianceStats(ctx context.Context) (*ComplianceStats, error) {
	url := fmt.Sprintf("%s/api/v1/compliance/stats", c.baseURL)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()
	
	var stats ComplianceStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &stats, nil
}

// ComplianceStats response
type ComplianceStats struct {
	TotalEntities       int64   `json:"total_entities"`
	ActiveLicenses      int64   `json:"active_licenses"`
	PendingApplications int64   `json:"pending_applications"`
	OverdueObligations  int64   `json:"overdue_obligations"`
	AverageScore        float64 `json:"average_score"`
	GoldTierCount       int64   `json:"gold_tier_count"`
	AtRiskCount         int64   `json:"at_risk_count"`
}
