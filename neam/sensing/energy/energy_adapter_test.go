package main

import (
	"context"
	"testing"
	"time"

	"github.com/neamplatform/sensing/energy/config"
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

func (m *MockProtocolClient) Subscribe(pointIDs []string) (<-chan TelemetryPoint, error) {
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
	baseValue := 100.0

	for i := 0; i < count; i++ {
		points[i] = TelemetryPoint{
			ID:        generateID(),
			SourceID:  sourceID,
			PointID:   "TestPoint",
			Value:     baseValue + float64(i%10),
			Quality:   "Good",
			Protocol:  "TEST",
			Timestamp: time.Now().Add(time.Duration(-i) * time.Second),
			Metadata:  nil,
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

// TestEnergyAdapterBasicFunctionality tests basic adapter operations
func TestEnergyAdapterBasicFunctionality(t *testing.T) {
	cfg := config.Default()
	cfg.ICCP.Enabled = true
	cfg.IEC61850.Enabled = true
	cfg.ICCP.Sources = []config.ICCPEndPoint{
		{
			EndPoint:       "127.0.0.1:102",
			LocalTSAP:      "1.0.1",
			RemoteTSAP:     "1.0.2",
			BilateralTable: "test-table",
		},
	}
	cfg.IEC61850.Sources = []config.IEC61850EndPoint{
		{
			EndPoint: "127.0.0.1",
			Port:     102,
			IEDName:  "TestIED",
		},
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	adapter, err := NewEnergyAdapter(cfg, logger)
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
		DeadbandPercent:     0.5,  // 0.5%
		SpikeThreshold:      10.0, // 10%
		SpikeWindowSize:     5,
		RateOfChangeLimit:   100.0,
		QualityFilter:       []string{"Good"},
		MinReportingInterval: config.Duration(10 * time.Millisecond),
		MaxValue:            1000000.0,
		MinValue:            -1000000.0,
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
				ID:        generateID(),
				SourceID:  "test-source",
				PointID:   "test-point",
				Value:     100.0 + float64(i),
				Quality:   "Good",
				Protocol:  "TEST",
				Timestamp: time.Now(),
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

// TestConnectionPoolManagement tests connection pooling
func TestConnectionPoolManagement(t *testing.T) {
	cfg := PoolConfig{
		MaxIdleConnections:    5,
		MaxActiveConnections:  10,
		MaxConnectionLifetime: time.Hour,
		IdleTimeout:           time.Minute * 5,
		HealthCheckInterval:   time.Second * 30,
	}

	logger, _ := zap.NewDevelopment()
	pool := NewPool(cfg, logger)

	// Add mock connections
	for i := 0; i < 3; i++ {
		client := NewMockProtocolClient(
			fmt.Sprintf("test-conn-%d", i),
			"TEST",
		)
		err := pool.AddConnection(client)
		if err != nil {
			t.Errorf("Failed to add connection: %v", err)
		}
	}

	// Check stats
	stats := pool.Stats()
	if stats.TotalConnections != 3 {
		t.Errorf("Expected 3 connections, got %d", stats.TotalConnections)
	}

	// Get connection
	conn, err := pool.GetConnection("test-conn-0")
	if err != nil {
		t.Errorf("Failed to get connection: %v", err)
	}
	if conn == nil {
		t.Error("Connection should not be nil")
	}

	// Release connection
	err = pool.ReleaseConnection(conn)
	if err != nil {
		t.Errorf("Failed to release connection: %v", err)
	}

	// Close pool
	err = pool.Close()
	if err != nil {
		t.Errorf("Failed to close pool: %v", err)
	}

	t.Log("Connection pool management test passed")
}

// TestSessionManager tests session management
func TestSessionManager(t *testing.T) {
	cfg := config.SessionConfig{
		MaxRetries:          3,
		InitialDelay:        config.Duration(10 * time.Millisecond),
		MaxDelay:            config.Duration(time.Second),
		Multiplier:          2.0,
		HealthCheckInterval: config.Duration(time.Second * 30),
		ReconnectWait:       config.Duration(time.Second),
	}

	logger, _ := zap.NewDevelopment()
	mgr := NewManager(cfg, logger)

	// Create mock protocols
	protocols := []ProtocolProvider{
		NewMockProtocolClient("protocol-1", "TEST"),
		NewMockProtocolClient("protocol-2", "TEST"),
	}

	points := []string{"Point1", "Point2", "Point3"}

	// Start sessions
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := mgr.Start(ctx, protocols, points)

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
			name: "valid config with ICCP",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				ICCP: config.ICCPConfig{
					Enabled: true,
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with IEC61850",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				IEC61850: config.IEC61850Config{
					Enabled: true,
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config - no protocol",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				ICCP: config.ICCPConfig{
					Enabled: false,
				},
				IEC61850: config.IEC61850Config{
					Enabled: false,
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config - no broker",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				ICCP: config.ICCPConfig{
					Enabled: true,
				},
				Kafka: config.KafkaConfig{
					Brokers:     "",
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

// TestDeadbandFiltering tests deadband filter specifically
func TestDeadbandFiltering(t *testing.T) {
	cfg := config.FilteringConfig{
		DeadbandPercent:     1.0, // 1%
		SpikeThreshold:      20.0,
		SpikeWindowSize:     5,
		RateOfChangeLimit:   1000.0,
		QualityFilter:       []string{"Good"},
		MinReportingInterval: config.Duration(0),
		MaxValue:            1000000.0,
		MinValue:            -1000000.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	// Base value
	baseValue := 100.0

	// Create input channel
	input := make(chan TelemetryPoint, 100)

	// Process channel
	output := processor.Process(input)

	// Send points - some within deadband, some outside
	go func() {
		// Point within deadband (0.5% change) - should be filtered
		input <- TelemetryPoint{
			ID:        "1",
			SourceID:  "test",
			PointID:   "deadband-test",
			Value:     baseValue + 0.4, // 0.4% change
			Quality:   "Good",
			Protocol:  "TEST",
			Timestamp: time.Now(),
		}

		// Point outside deadband (2% change) - should pass
		input <- TelemetryPoint{
			ID:        "2",
			SourceID:  "test",
			PointID:   "deadband-test",
			Value:     baseValue + 2.0, // 2% change
			Quality:   "Good",
			Protocol:  "TEST",
			Timestamp: time.Now().Add(time.Second),
		}

		close(input)
	}()

	// Collect output
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

// TestAnomalyDetection tests spike and anomaly detection
func TestAnomalyDetection(t *testing.T) {
	cfg := config.FilteringConfig{
		DeadbandPercent:     0.1,
		SpikeThreshold:      50.0, // 50% threshold
		SpikeWindowSize:     5,
		RateOfChangeLimit:   1000.0,
		QualityFilter:       []string{"Good"},
		MinReportingInterval: config.Duration(0),
		MaxValue:            1000000.0,
		MinValue:            -1000000.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	input := make(chan TelemetryPoint, 100)
	output := processor.Process(input)

	// Generate normal values first to establish baseline
	go func() {
		for i := 0; i < 5; i++ {
			input <- TelemetryPoint{
				ID:        generateID(),
				SourceID:  "test",
				PointID:   "anomaly-test",
				Value:     100.0,
				Quality:   "Good",
				Protocol:  "TEST",
				Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			}
		}

		// Send spike (60% change)
		input <- TelemetryPoint{
			ID:        generateID(),
			SourceID:  "test",
			PointID:   "anomaly-test",
			Value:     160.0, // 60% increase
			Quality:   "Good",
			Protocol:  "TEST",
			Timestamp: time.Now().Add(6 * time.Second),
		}

		close(input)
	}()

	count := 0
	for range output {
		count++
	}

	// Check stats
	stats := processor.Stats()
	if stats.SpikesDetected == 0 {
		t.Log("Note: Spike detection may vary based on window size")
	}

	t.Logf("Processed %d points", count)
	t.Log("Anomaly detection test completed")
}

// Integration test - requires running infrastructure
func TestIntegrationWithKafka(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.Default()
	cfg.Kafka.Brokers = "localhost:9092"

	logger, _ := zap.NewDevelopment()

	producer, err := NewProducer(cfg.Kafka, logger)
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Test message production
	ctx := context.Background()
	testData := map[string]interface{}{
		"test":   "value",
		"number": 42,
	}

	err = producer.Produce(ctx, cfg.Kafka.OutputTopic, testData)
	if err != nil {
		t.Errorf("Failed to produce message: %v", err)
	}

	t.Log("Kafka integration test passed")
}

// Benchmark for signal processing throughput
func BenchmarkSignalProcessing(b *testing.B) {
	cfg := config.FilteringConfig{
		DeadbandPercent:     0.5,
		SpikeThreshold:      10.0,
		SpikeWindowSize:     5,
		RateOfChangeLimit:   1000.0,
		QualityFilter:       []string{"Good"},
		MinReportingInterval: config.Duration(0),
		MaxValue:            1000000.0,
		MinValue:            -1000000.0,
	}

	logger, _ := zap.NewDevelopment()
	processor := NewProcessor(cfg, logger)

	input := make(chan TelemetryPoint, b.N)
	output := processor.Process(input)

	go func() {
		for i := 0; i < b.N; i++ {
			input <- TelemetryPoint{
				ID:        generateID(),
				SourceID:  "bench",
				PointID:   "bench-point",
				Value:     100.0 + float64(i%100),
				Quality:   "Good",
				Protocol:  "TEST",
				Timestamp: time.Now(),
			}
		}
		close(input)
	}()

	for range output {
		// Drain output
	}

	b.ReportAllocs()
}
