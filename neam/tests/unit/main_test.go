// NEAM Unit Tests - Main Test Suite
// Comprehensive unit tests for NEAM platform components

package tests

import (
	"encoding/json"
	"testing"
	"time"

	"neam-platform/shared"
)

// Test Configuration
func TestConfigLoad(t *testing.T) {
	config := &shared.Config{}

	// Test default values
	err := config.Load("sensing")
	if err != nil {
		t.Errorf("Expected no error loading config, got: %v", err)
	}

	if config.HTTPPort != 8082 {
		t.Errorf("Expected sensing port 8082, got: %d", config.HTTPPort)
	}

	if config.KafkaPrefix != "neam" {
		t.Errorf("Expected kafka prefix 'neam', got: %s", config.KafkaPrefix)
	}
}

func TestConfigLoadIntelligence(t *testing.T) {
	config := &shared.Config{}
	err := config.Load("intelligence")
	if err != nil {
		t.Errorf("Expected no error loading config, got: %v", err)
	}

	if config.HTTPPort != 8083 {
		t.Errorf("Expected intelligence port 8083, got: %d", config.HTTPPort)
	}
}

func TestConfigLoadIntervention(t *testing.T) {
	config := &shared.Config{}
	err := config.Load("intervention")
	if err != nil {
		t.Errorf("Expected no error loading config, got: %v", err)
	}

	if config.HTTPPort != 8084 {
		t.Errorf("Expected intervention port 8084, got: %d", config.HTTPPort)
	}
}

// Test Data Structures
func TestEconomicIndicator(t *testing.T) {
	indicator := shared.EconomicIndicator{
		ID:        "test-indicator-001",
		Type:      "payment_volume",
		Source:    "upi",
		Region:    "REGION-NORTH",
		Value:     1.5e9,
		Unit:      "INR",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"transaction_count": 4500000,
		},
	}

	if indicator.ID != "test-indicator-001" {
		t.Errorf("Expected ID 'test-indicator-001', got: %s", indicator.ID)
	}

	if indicator.Type != "payment_volume" {
		t.Errorf("Expected type 'payment_volume', got: %s", indicator.Type)
	}

	if indicator.Value != 1.5e9 {
		t.Errorf("Expected value 1.5e9, got: %f", indicator.Value)
	}

	// Test JSON serialization
	data, err := json.Marshal(indicator)
	if err != nil {
		t.Errorf("Failed to serialize indicator: %v", err)
	}

	var unmarshaled shared.EconomicIndicator
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to deserialize indicator: %v", err)
	}

	if unmarshaled.ID != indicator.ID {
		t.Errorf("Expected same ID after serialization")
	}
}

func TestAnomaly(t *testing.T) {
	anomaly := shared.Anomaly{
		ID:          "anom-test-001",
		Type:        "payment_surge",
		Severity:    "high",
		Region:      "REGION-WEST",
		Description: "Unusual payment volume increase detected",
		Timestamp:   time.Now(),
		Data: map[string]interface{}{
			"volume_increase_pct": 45.2,
		},
		Resolved: false,
	}

	if anomaly.Severity != "high" {
		t.Errorf("Expected severity 'high', got: %s", anomaly.Severity)
	}

	if anomaly.Resolved {
		t.Errorf("Expected anomaly to be unresolved")
	}

	// Mark as resolved
	anomaly.Resolved = true
	if !anomaly.Resolved {
		t.Errorf("Expected anomaly to be resolved after update")
	}
}

func TestPolicy(t *testing.T) {
	policy := shared.Policy{
		ID:   "pol-test-001",
		Name: "Regional Inflation Control",
		Type: "inflation_control",
		Rules: []shared.PolicyRule{
			{
				ID:        "rule-001",
				Condition: "inflation_rate > 6.0",
				Action:    "alert_ministry",
				Priority:  1,
				Enabled:   true,
			},
			{
				ID:        "rule-002",
				Condition: "inflation_rate > 8.0",
				Action:    "intervene",
				Parameters: map[string]interface{}{
					"intervention_type": "price_controls",
				},
				Priority: 2,
				Enabled:  true,
			},
		},
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if len(policy.Rules) != 2 {
		t.Errorf("Expected 2 rules, got: %d", len(policy.Rules))
	}

	if policy.Rules[0].Priority != 1 {
		t.Errorf("Expected first rule priority 1, got: %d", policy.Rules[0].Priority)
	}

	if policy.Status != "active" {
		t.Errorf("Expected status 'active', got: %s", policy.Status)
	}
}

func TestIntervention(t *testing.T) {
	now := time.Now()
	intervention := shared.Intervention{
		ID:         "int-test-001",
		PolicyID:   "pol-test-001",
		Type:       "price_controls",
		Status:     "executed",
		Region:     "REGION-NORTH",
		Parameters: map[string]interface{}{},
		ExecutedAt: now,
	}

	if intervention.Status != "executed" {
		t.Errorf("Expected status 'executed', got: %s", intervention.Status)
	}

	if intervention.CompletedAt != nil {
		t.Errorf("Expected completed_at to be nil")
	}

	// Complete the intervention
	completed := now.Add(1 * time.Hour)
	intervention.CompletedAt = &completed

	if intervention.CompletedAt == nil {
		t.Errorf("Expected completed_at to be set")
	}
}

func TestTimeSeriesData(t *testing.T) {
	now := time.Now()
	series := shared.TimeSeriesData{
		Metric:    "payment_volume",
		Region:    "REGION-NORTH",
		Frequency: "hourly",
		DataPoints: []shared.DataPoint{
			{Timestamp: now.Add(-3 * time.Hour), Value: 1.4e9},
			{Timestamp: now.Add(-2 * time.Hour), Value: 1.5e9},
			{Timestamp: now.Add(-1 * time.Hour), Value: 1.6e9},
			{Timestamp: now, Value: 1.5e9},
		},
	}

	if len(series.DataPoints) != 4 {
		t.Errorf("Expected 4 data points, got: %d", len(series.DataPoints))
	}

	// Calculate average
	total := 0.0
	for _, dp := range series.DataPoints {
		total += dp.Value
	}
	average := total / float64(len(series.DataPoints))

	if average < 1.4e9 || average > 1.6e9 {
		t.Errorf("Expected average around 1.5e9, got: %f", average)
	}
}

func TestExportRequest(t *testing.T) {
	now := time.Now()
	request := shared.ExportRequest{
		Type:    "parliament",
		Format:  "pdf",
		Period: shared.TimePeriod{
			Start: now.AddDate(0, -3, 0),
			End:   now,
		},
		Regions: []string{"REGION-NORTH", "REGION-SOUTH"},
		Metrics: []string{"payment_volume", "inflation_rate"},
		Options: map[string]bool{
			"include_charts":   true,
			"include_tables":   true,
			"include_summary":  true,
		},
	}

	if request.Type != "parliament" {
		t.Errorf("Expected type 'parliament', got: %s", request.Type)
	}

	if request.Format != "pdf" {
		t.Errorf("Expected format 'pdf', got: %s", request.Format)
	}

	if len(request.Regions) != 2 {
		t.Errorf("Expected 2 regions, got: %d", len(request.Regions))
	}

	if len(request.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got: %d", len(request.Metrics))
	}

	if !request.Options["include_charts"] {
		t.Errorf("Expected include_charts to be true")
	}
}

func TestAuditRecord(t *testing.T) {
	now := time.Now()
	record := shared.AuditRecord{
		ID:         "audit-test-001",
		Actor:      "usr-001",
		Action:     "login",
		Resource:   "auth",
		ResourceID: "auth-001",
		Details: map[string]interface{}{
			"method": "password",
			"mfa":    true,
		},
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0",
		Timestamp: now,
		Success:   true,
	}

	if record.Actor != "usr-001" {
		t.Errorf("Expected actor 'usr-001', got: %s", record.Actor)
	}

	if !record.Success {
		t.Errorf("Expected success to be true")
	}

	if record.Error != "" {
		t.Errorf("Expected error to be empty")
	}

	// Test failed action
	record.Success = false
	record.Error = "invalid_credentials"

	if record.Success {
		t.Errorf("Expected success to be false after error")
	}

	if record.Error != "invalid_credentials" {
		t.Errorf("Expected error message 'invalid_credentials', got: %s", record.Error)
	}
}

func TestRegion(t *testing.T) {
	region := shared.Region{
		ID:        "REGION-NORTH",
		Name:      "Northern Region",
		Code:      "NR",
		ParentID:  "NATION-001",
		Coordinates: []float64{28.6139, 77.2090}, // Delhi coordinates
	}

	if region.ID != "REGION-NORTH" {
		t.Errorf("Expected ID 'REGION-NORTH', got: %s", region.ID)
	}

	if region.Name != "Northern Region" {
		t.Errorf("Expected name 'Northern Region', got: %s", region.Name)
	}

	if len(region.Coordinates) != 2 {
		t.Errorf("Expected 2 coordinates, got: %d", len(region.Coordinates))
	}

	if region.Coordinates[0] != 28.6139 {
		t.Errorf("Expected first coordinate 28.6139, got: %f", region.Coordinates[0])
	}
}

// Test validation logic
func TestPolicyValidation(t *testing.T) {
	tests := []struct {
		name    string
		policy  shared.Policy
		valid   bool
	}{
		{
			name: "valid policy",
			policy: shared.Policy{
				ID:     "pol-001",
				Name:   "Test Policy",
				Type:   "inflation_control",
				Status: "active",
				Rules: []shared.PolicyRule{
					{ID: "r1", Condition: "x > 0", Action: "alert", Priority: 1, Enabled: true},
				},
			},
			valid: true,
		},
		{
			name: "policy without rules",
			policy: shared.Policy{
				ID:     "pol-002",
				Name:   "Empty Policy",
				Type:   "test",
				Status: "draft",
				Rules:  []shared.PolicyRule{},
			},
			valid: true, // Empty policies are valid (for drafting)
		},
		{
			name: "policy without name",
			policy: shared.Policy{
				ID:     "pol-003",
				Name:   "",
				Type:   "test",
				Status: "draft",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.policy.Name != ""
			if isValid != tt.valid {
				t.Errorf("Expected validity %v, got %v", tt.valid, isValid)
			}
		})
	}
}

// Test data aggregation
func TestDataAggregation(t *testing.T) {
	// Simulate aggregating data from multiple regions
	regionData := map[string]float64{
		"REGION-NORTH": 1.5e9,
		"REGION-SOUTH": 1.3e9,
		"REGION-EAST":  0.9e9,
		"REGION-WEST":  1.7e9,
	}

	total := 0.0
	count := 0
	for _, value := range regionData {
		total += value
		count++
	}

	average := total / float64(count)

	if average < 1.3e9 || average > 1.4e9 {
		t.Errorf("Expected average around 1.35e9, got: %f", average)
	}

	if total < 5.3e9 || total > 5.5e9 {
		t.Errorf("Expected total around 5.4e9, got: %f", total)
	}
}

// Test time period calculations
func TestTimePeriod(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		period shared.TimePeriod
		days   int
	}{
		{
			name: "monthly period",
			period: shared.TimePeriod{
				Start: now.AddDate(0, -1, 0),
				End:   now,
			},
			days: 30,
		},
		{
			name: "quarterly period",
			period: shared.TimePeriod{
				Start: now.AddDate(0, -3, 0),
				End:   now,
			},
			days: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tt.period.End.Sub(tt.period.Start)
			days := int(duration.Hours() / 24)

			// Allow for some variance due to month lengths
			if days < tt.days-2 || days > tt.days+2 {
				t.Errorf("Expected around %d days, got: %d", tt.days, days)
			}
		})
	}
}

// Test JSON serialization edge cases
func TestJSONSerialization(t *testing.T) {
	// Test with special characters
	indicator := shared.EconomicIndicator{
		ID:          "test-001",
		Type:        "test_type",
		Description: "Test with special chars: <>&\"",
		Value:       100.0,
		Timestamp:   time.Now(),
	}

	data, err := json.Marshal(indicator)
	if err != nil {
		t.Errorf("Failed to serialize: %v", err)
	}

	var unmarshaled shared.EconomicIndicator
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to deserialize: %v", err)
	}

	if unmarshaled.ID != indicator.ID {
		t.Errorf("ID mismatch after serialization")
	}
}
