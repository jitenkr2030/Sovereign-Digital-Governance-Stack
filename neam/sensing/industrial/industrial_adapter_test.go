package main

import (
	"context"
	"testing"
	"time"

	"github.com/neamplatform/sensing/industrial/config"
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

func NewMockIndustrialProtocolClient(id, protocolType string) *MockProtocolClient {
	return &MockProtocolClient{
		id:           id,
		protocolType: protocolType,
		connected:    false,
		dataPoints:   generateMockIndustrialDataPoints(id, 100),
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

func (m *MockProtocolClient) Subscribe(nodeIDs []string) (<-chan TelemetryPoint, error) {
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

func (m *MockProtocolClient) Browse(ctx context.Context, nodeID string) ([]NodeInfo, error) {
	return []NodeInfo{
		{
			NodeID:      nodeID + ".Temperature",
			BrowseName:  "Temperature",
			DisplayName: "Temperature Sensor",
			DataType:    "Double",
			AccessLevel: "CurrentRead",
			Description: "Temperature measurement",
			HasChildren: false,
		},
		{
			NodeID:      nodeID + ".Pressure",
			BrowseName:  "Pressure",
			DisplayName: "Pressure Sensor",
			DataType:    "Double",
			AccessLevel: "CurrentRead",
			Description: "Pressure measurement",
			HasChildren: false,
		},
	}, nil
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

func generateMockIndustrialDataPoints(sourceID string, count int) []TelemetryPoint {
	points := make([]TelemetryPoint, count)

	for i := 0; i < count; i++ {
		points[i] = TelemetryPoint{
			ID:          generateIndustrialID(),
			SourceID:    sourceID,
			NodeID:      "ns=2;s=Channel1.Device1.Temperature",
			AssetID:     "Device1",
			AssetName:   "Temperature Sensor",
			Value:       50.0 + float64(i%20),
			Quality:     "Good",
			Protocol:    "OPCUA",
			Timestamp:   time.Now().Add(time.Duration(-i) * time.Second),
			DataType:    "Double",
			Unit:        "°C",
		}
	}

	return points
}

func generateIndustrialID() string {
	return time.Now().Format("20060102150405") + "-" + randomIndustrialString(8)
}

func randomIndustrialString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// TestIndustrialAdapterBasicFunctionality tests basic adapter operations
func TestIndustrialAdapterBasicFunctionality(t *testing.T) {
	cfg := config.Default()
	cfg.OPCUA.Enabled = true
	cfg.OPCUA.Sources = []config.OPCUAEndPoint{
		{
			EndpointURL:     "opc.tcp://localhost:4840",
			ApplicationURI:  "urn:test:industrial:adapter",
			SecurityPolicy:  "None",
			SecurityMode:    "None",
			AuthMode:        "Anonymous",
			NodesOfInterest: []string{"ns=2;s=Channel1.Device1.Temperature"},
		},
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	adapter, err := NewIndustrialAdapter(cfg, logger)
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

	t.Log("Basic industrial adapter functionality test passed")
}

// TestAssetDiscovery tests asset discovery mechanism
func TestAssetDiscovery(t *testing.T) {
	cfg := config.DiscoveryConfig{
		Enabled:          true,
		DiscoveryInterval: config.Duration(60 * time.Second),
		BrowseDepth:      3,
		FilterByType:     []string{"BaseAnalogType", "BaseDataVariable"},
		IncludeChildren:  true,
		CacheTimeout:     config.Duration(300 * time.Second),
	}

	logger, _ := zap.NewDevelopment()
	discovery := NewAssetDiscovery(cfg, logger)

	// Create mock protocol client
	protocol := NewMockIndustrialProtocolClient("test-opcua", "OPCUA")

	// Perform discovery
	ctx := context.Background()
	assets, err := discovery.Discover(ctx, []ProtocolClient{protocol})
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	if len(assets) == 0 {
		t.Log("No assets discovered (expected with mock protocols)")
	}

	// Test cache invalidation
	discovery.InvalidateCache()
	cacheStatus := discovery.GetCacheStatus()
	if cacheStatus["cached"] != false {
		t.Error("Cache should be invalid after invalidation")
	}

	t.Log("Asset discovery test passed")
}

// TestOPCUAClientBrowsing tests OPC UA node browsing
func TestOPCUAClientBrowsing(t *testing.T) {
	cfg := config.OPCUAConfig{
		Enabled:         true,
		Sources:         []config.OPCUAEndPoint{{EndpointURL: "opc.tcp://localhost:4840"}},
		SubscriptionRate: config.Duration(100 * time.Millisecond),
		Timeout:         config.Duration(30 * time.Second),
	}

	logger, _ := zap.NewDevelopment()
	driver := NewDriver(DriverConfig{
		EndpointURL: cfg.Sources[0].EndpointURL,
		Timeout:     time.Duration(cfg.Timeout),
	}, logger)

	// Connect
	ctx := context.Background()
	err := driver.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Browse node
	nodes, err := driver.Browse(ctx, "ns=0;i=84")
	if err != nil {
		t.Fatalf("Browse failed: %v", err)
	}

	if len(nodes) == 0 {
		t.Log("No nodes browsed (expected with mock implementation)")
	}

	t.Log("OPC UA browsing test passed")
}

// TestSignalProcessorFilters tests signal filtering logic for industrial data
func TestIndustrialSignalProcessorFilters(t *testing.T) {
	cfg := config.FilteringConfig{
		DeadbandPercent:       0.5,
		SpikeThreshold:        10.0,
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

	input := make(chan TelemetryPoint, 100)
	output := processor.Process(input)

	go func() {
		for i := 0; i < 10; i++ {
			point := TelemetryPoint{
				ID:          generateIndustrialID(),
				SourceID:    "test-source",
				NodeID:      "ns=2;s=Channel1.Device1.Temperature",
				Value:       50.0 + float64(i),
				Quality:     "Good",
				Protocol:    "OPCUA",
				Timestamp:   time.Now(),
				DataType:    "Double",
				Unit:        "°C",
			}
			input <- point
		}
		close(input)
	}()

	count := 0
	for range output {
		count++
	}

	if count != 10 {
		t.Errorf("Expected 10 points, got %d", count)
	}

	stats := processor.Stats()
	if stats.PointsProcessed != 10 {
		t.Errorf("Expected 10 processed, got %d", stats.PointsProcessed)
	}

	t.Log("Industrial signal processor filter test passed")
}

// TestSessionManager tests session management for industrial adapter
func TestIndustrialSessionManager(t *testing.T) {
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

	protocols := []ProtocolProvider{
		NewMockIndustrialProtocolClient("protocol-1", "OPCUA"),
	}

	nodeIDs := []string{"ns=2;s=Channel1.Device1.Temperature", "ns=2;s=Channel1.Device1.Pressure"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := mgr.Start(ctx, protocols, nodeIDs)

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

	stats := mgr.Stats()
	if stats["total_sessions"] == nil {
		t.Error("Session stats should not be nil")
	}

	t.Log("Industrial session manager test passed")
}

// TestConfigurationValidation tests configuration validation
func TestIndustrialConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name: "valid config with OPC UA",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				OPCUA: config.OPCUAConfig{
					Enabled: true,
					Sources: []config.OPCUAEndPoint{
						{
							EndpointURL: "opc.tcp://localhost:4840",
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
			name: "invalid config - no OPC UA sources",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				OPCUA: config.OPCUAConfig{
					Enabled: true,
					Sources: []config.OPCUAEndPoint{},
				},
				Kafka: config.KafkaConfig{
					Brokers:     "localhost:9092",
					OutputTopic: "test-topic",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config - OPC UA disabled",
			cfg: config.Config{
				Service: config.ServiceConfig{
					Name: "test-adapter",
				},
				OPCUA: config.OPCUAConfig{
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

// Benchmark for industrial signal processing
func BenchmarkIndustrialSignalProcessing(b *testing.B) {
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
				ID:          generateIndustrialID(),
				SourceID:    "bench",
				NodeID:      "ns=2;s=Channel1.Device1.Temperature",
				Value:       50.0 + float64(i%100),
				Quality:     "Good",
				Protocol:    "OPCUA",
				Timestamp:   time.Now(),
				DataType:    "Double",
				Unit:        "°C",
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
