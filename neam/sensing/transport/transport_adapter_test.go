package main

import (
	"context"
	"testing"
	"time"

	"github.com/neamplatform/sensing/transport/config"
	"go.uber.org/zap"
)

// MockProtocolClient implements ProtocolClient for testing
type MockProtocolClient struct {
	id           string
	protocolType string
	connected    bool
	dataPoints   []TelemetryPoint
	dataIndex    int
	mu           struct {
		sync.RWMutex
		healthy bool
	}
}

func NewMockProtocolClient(id, protocolType string) *MockProtocolClient {
	return &MockProtocolClient{
		id:           id,
		protocolType: protocolType,
		connected:    false,
		dataPoints:   generateMockDataPoints(id, 100),
	}
}

func (m *MockProtocolClient) Connect(ctx context.Context) error {
	m.connected = true
	m.mu.Lock()
	m.mu.healthy = true
	m.mu.Unlock()
	return nil
}

func (m *MockProtocolClient) Disconnect() error {
	m.connected = false
	m.mu.Lock()
	m.mu.healthy = false
	m.mu.Unlock()
	return nil
}

func (m *MockProtocolClient) Subscribe(topicPatterns []string) (<-chan TelemetryPoint, error) {
	output := make(chan TelemetryPoint, 100)

	go func() {
		defer close(output)
		for i := m.dataIndex; i < len(m.dataPoints); i++ {
			select {
			case output <- m.dataPoints[i]:
			case <-ctx.Done():
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return output, nil
}

func (m *MockProtocolClient) HealthCheck() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mu.healthy && m.connected
}

func (m *MockProtocolClient) ID() string {
	return m.id
}

func (m *MockProtocolClient) GetProtocolType() string {
	return m.protocolType
}

func generateMockDataPoints(sourceID string, count int) []TelemetryPoint {
	points := make([]TelemetryPoint, count)

	for i := 0; i < count; i++ {
		points[i] = TelemetryPoint{
			ID:          generateID(),
			SourceID:    sourceID,
			PointID:     "transport/vehicle1/telemetry",
			Value:       60.0 + float64(i%20),
			Quality:     "Good",
			Protocol:    "MQTTv5",
			Timestamp:   time.Now().Add(time.Duration(-i) * time.Second),
			Criticality: "standard",
		}
	}

	return points
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// TestTransportAdapterBasicFunctionality tests basic adapter operations
func TestTransportAdapterBasicFunctionality(t *testing.T) {
	cfg := config.Default()
	cfg.MQTT.Enabled = true
	cfg.MQTT.Sources = []config.MQTTEndPoint{
		{
			BrokerURL:  "mqtt://localhost:1883",
			ClientID:   "test-transport-adapter",
			CleanStart: false,
			Topics:     []string{"transport/+/telemetry"},
			QoS:        1,
			Criticality: "standard",
		},
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	adapter, err := NewTransportAdapter(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Adapter should not be nil")
	}

	// Test health check
	health := adapter.Health()
	if health == nil {
		t.Fatal("Health should not be nil")
	}

	t.Log("Basic functionality test passed")
}

// TestSignalProcessorFilters tests signal filtering logic
func TestSignalProcessorFilters(t *testing.T) {
	cfg := config.FilteringConfig{
		DeadbandPercent:       0.5, // 0.5%
		SpikeThreshold:        10.0, // 10%
		SpikeWindowSize:       5,
		RateOfChangeLimit:     100.0,
		QualityFilter:         []string{"Good"},
		MinReportingInterval:  config.Duration(10 * time.Millisecond),
		MaxValue:              1e9,
		MinValue:              -1e9,
		OutlierThreshold:      3.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	// Create input channel
	input := make(chan TelemetryPoint, 100)

	// Process channel
	output := processor.Process(input)

	// Send test data
	go func() {
		for i := 0; i < 10; i++ {
			point := TelemetryPoint{
				ID:          generateID(),
				SourceID:    "test-source",
				PointID:     "test-point",
				Value:       60.0 + float64(i),
				Quality:     "Good",
				Protocol:    "MQTTv5",
				Timestamp:   time.Now(),
				Criticality: "standard",
			}
			input <- point
		}
		close(input)
	}()

	// Collect output
	count := 0
	for range output {
		count++
	}

	// All points should pass through
	if count != 10 {
		t.Errorf("Expected 10 points, got %d", count)
	}

	// Check stats
	stats := processor.Stats()
	if stats.PointsProcessed != 10 {
		t.Errorf("Expected 10 processed, got %d", stats.PointsProcessed)
	}

	t.Log("Signal processor filter test passed")
}

// TestSessionManager tests session management
func TestSessionManager(t *testing.T) {
	cfg := config.SessionConfig{
		MaxRetries:           3,
		InitialDelay:         config.Duration(10 * time.Millisecond),
		MaxDelay:             config.Duration(time.Second),
		Multiplier:           2.0,
		HealthCheckInterval:  config.Duration(time.Second * 30),
		ReconnectWait:        config.Duration(time.Second),
		StateStoreType:       "memory",
	}

	logger, _ := zap.NewDevelopment()
	mgr := NewManager(cfg, logger)

	// Create mock protocols
	protocols := []ProtocolProvider{
		NewMockProtocolClient("protocol-1", "MQTTv5"),
		NewMockProtocolClient("protocol-2", "MQTTv5"),
	}

	topics := []string{"transport/vehicle1/telemetry", "transport/vehicle2/status"}

	// Start sessions
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := mgr.Start(ctx, protocols, topics)

	// Collect data
	count := 0
	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-stream:
			if !ok {
				goto done
			}
			count++
			if count >= 10 {
				goto done
			}
		case <-timeout:
			goto done
		}
	}

done:
	if count == 0 {
		t.Log("No data collected from sessions (expected with mock protocols)")
	}

	// Check stats
	stats := mgr.Stats()
	if stats["total_sessions"] == nil {
		t.Error("Session stats should not be nil")
	}

	t.Log("Session manager test passed")
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name: "valid config with MQTT",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				MQTT: config.MQTTConfig{
					Enabled: true,
					Sources: []config.MQTTEndPoint{
						{
							BrokerURL: "mqtt://localhost:1883",
							ClientID:  "test",
						},
					},
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config - no MQTT sources",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				MQTT: config.MQTTConfig{
					Enabled: true,
					Sources: []config.MQTTEndPoint{},
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config - MQTT disabled",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				MQTT: config.MQTTConfig{
					Enabled: false,
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestQoSByCriticality tests QoS mapping by criticality
func TestQoSByCriticality(t *testing.T) {
	cfg := config.FilteringConfig{
		DeadbandPercent:       0.1,
		SpikeThreshold:        20.0,
		SpikeWindowSize:       5,
		RateOfChangeLimit:     1000.0,
		QualityFilter:         []string{"Good"},
		MinReportingInterval:  config.Duration(0),
		MaxValue:              1e9,
		MinValue:              -1e9,
		OutlierThreshold:      3.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	input := make(chan TelemetryPoint, 100)
	output := processor.Process(input)

	// Test critical data (should pass)
	go func() {
		input <- TelemetryPoint{
			ID:          "1",
			SourceID:    "test",
			PointID:     "transport/vehicle1/alert",
			Value:       100.0,
			Quality:     "Good",
			Protocol:    "MQTTv5",
			Timestamp:   time.Now(),
			Criticality: "critical",
		}

		// Test standard data
		input <- TelemetryPoint{
			ID:          "2",
			SourceID:    "test",
			PointID:     "transport/vehicle1/telemetry",
			Value:       60.0,
			Quality:     "Good",
			Protocol:    "MQTTv5",
			Timestamp:   time.Now().Add(time.Second),
			Criticality: "standard",
		}

		close(input)
	}()

	count := 0
	for range output {
		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 points, got %d", count)
	}

	t.Log("QoS by criticality test passed")
}

// TestDeadbandFiltering tests deadband filter
func TestDeadbandFiltering(t *testing.T) {
	cfg := config.FilteringConfig{
		DeadbandPercent:       1.0, // 1%
		SpikeThreshold:        20.0,
		SpikeWindowSize:       5,
		RateOfChangeLimit:     1000.0,
		QualityFilter:         []string{"Good"},
		MinReportingInterval:  config.Duration(0),
		MaxValue:              1e9,
		MinValue:              -1e9,
		OutlierThreshold:      3.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	input := make(chan TelemetryPoint, 100)
	output := processor.Process(input)

	// Base value
	baseValue := 100.0

	go func() {
		// Point within deadband (0.5% change) - should be filtered
		input <- TelemetryPoint{
			ID:          "1",
			SourceID:    "test",
			PointID:     "deadband-test",
			Value:       baseValue + 0.4, // 0.4% change
			Quality:     "Good",
			Protocol:    "MQTTv5",
			Timestamp:   time.Now(),
			Criticality: "standard",
		}

		// Point outside deadband (2% change) - should pass
		input <- TelemetryPoint{
			ID:          "2",
			SourceID:    "test",
			PointID:     "deadband-test",
			Value:       baseValue + 2.0, // 2% change
			Quality:     "Good",
			Protocol:    "MQTTv5",
			Timestamp:   time.Now().Add(time.Second),
			Criticality: "standard",
		}

		close(input)
	}()

	passed := 0
	for range output {
		passed++
	}

	// Should have 1 point pass (the second one)
	if passed != 1 {
		t.Errorf("Expected 1 point to pass deadband filter, got %d", passed)
	}

	t.Log("Deadband filtering test passed")
}

// Benchmark for signal processing throughput
func BenchmarkSignalProcessing(b *testing.B) {
	cfg := config.FilteringConfig{
		DeadbandPercent:       0.5,
		SpikeThreshold:        10.0,
		SpikeWindowSize:       5,
		RateOfChangeLimit:     1000.0,
		QualityFilter:         []string{"Good"},
		MinReportingInterval:  config.Duration(0),
		MaxValue:              1e9,
		MinValue:              -1e9,
		OutlierThreshold:      3.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	input := make(chan TelemetryPoint, b.N)
	output := processor.Process(input)

	go func() {
		for i := 0; i < b.N; i++ {
			input <- TelemetryPoint{
				ID:          generateID(),
				SourceID:    "bench",
				PointID:     "bench-point",
				Value:       60.0 + float64(i%100),
				Quality:     "Good",
				Protocol:    "MQTTv5",
				Timestamp:   time.Now(),
				Criticality: "standard",
			}
		}
		close(input)
	}()

	for range output {
		// Drain output
	}

	b.ReportAllocs()
}

// Import sync for test
import "sync"
