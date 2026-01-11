package opcua

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DriverConfig contains OPC UA driver configuration
type DriverConfig struct {
	EndpointURL      string
	ApplicationURI   string
	ProductURI       string
	ApplicationName  string
	SecurityPolicy   string  // None, Basic256Sha256
	SecurityMode     string  // None, Sign, SignAndEncrypt
	AuthMode         string  // Anonymous, UserName, Certificate
	Username         string
	Password         string
	CertificatePath  string
	PrivateKeyPath   string
	Timeout          time.Duration
	RetryInterval    time.Duration
	ReconnectWait    time.Duration
	SubscriptionRate time.Duration
}

// Driver implements OPC UA protocol client
type Driver struct {
	config       DriverConfig
	logger       *zap.Logger
	connected    bool
	subscription *Subscription
	nodeCache    map[string]*NodeInfo
	mu           sync.RWMutex
	wg           sync.WaitGroup
}

// Subscription holds OPC UA subscription state
type Subscription struct {
	ID            int
	MonitoredItems map[string]*MonitoredItem
	LastRefresh   time.Time
}

// MonitoredItem represents a monitored item within a subscription
type MonitoredItem struct {
	ID        int
	NodeID    string
	Attribute int
	QoS       int
	Enabled   bool
}

// NodeInfo represents information about an OPC UA node
type NodeInfo struct {
	NodeID      string
	BrowseName  string
	DisplayName string
	DataType    string
	AccessLevel string
	Description string
	ParentID    string
	HasChildren bool
	Value       interface{}
	Quality     string
	Timestamp   time.Time
}

// NewDriver creates a new OPC UA driver instance
func NewDriver(config DriverConfig, logger *zap.Logger) *Driver {
	return &Driver{
		config:    config,
		logger:    logger,
		nodeCache: make(map[string]*NodeInfo),
	}
}

// ID returns the driver identifier
func (d *Driver) ID() string {
	return fmt.Sprintf("OPCUA-%s", d.config.EndpointURL)
}

// GetProtocolType returns the protocol type
func (d *Driver) GetProtocolType() string {
	return "OPCUA"
}

// Connect establishes an OPC UA connection
func (d *Driver) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return nil
	}

	d.logger.Info("Connecting to OPC UA server",
		zap.String("endpoint", d.config.EndpointURL),
		zap.String("security_policy", d.config.SecurityPolicy),
		zap.String("security_mode", d.config.SecurityMode),
		zap.String("auth_mode", d.config.AuthMode))

	// In production, this would establish actual OPC UA connection
	// For now, we simulate the connection
	d.connected = true

	d.logger.Info("OPC UA connection established",
		zap.String("endpoint", d.config.EndpointURL))

	return nil
}

// Disconnect closes the OPC UA connection
func (d *Driver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil
	}

	// Unsubscribe from all monitored items
	if d.subscription != nil {
		d.subscription.MonitoredItems = make(map[string]*MonitoredItem)
	}

	d.connected = false
	d.logger.Info("OPC UA connection closed",
		zap.String("endpoint", d.config.EndpointURL))

	return nil
}

// HealthCheck verifies connection health
func (d *Driver) HealthCheck() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.connected
}

// Subscribe initiates subscription for specified node IDs
func (d *Driver) Subscribe(nodeIDs []string) (<-chan TelemetryPoint, error) {
	d.mu.RLock()
	if !d.connected {
		d.mu.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	d.mu.RUnlock()

	// Create subscription
	d.mu.Lock()
	d.subscription = &Subscription{
		ID:            int(time.Now().UnixNano()),
		MonitoredItems: make(map[string]*MonitoredItem),
		LastRefresh:   time.Now(),
	}
	d.mu.Unlock()

	// Create output channel
	output := make(chan TelemetryPoint, 1000)

	// Register monitored items
	for _, nodeID := range nodeIDs {
		item := &MonitoredItem{
			ID:        len(d.subscription.MonitoredItems) + 1,
			NodeID:    nodeID,
			Attribute: 13, // Value attribute
			QoS:       0,
			Enabled:   true,
		}
		d.subscription.MonitoredItems[nodeID] = item

		d.logger.Debug("Monitoring node",
			zap.String("node_id", nodeID))
	}

	// Start data collection
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.dataCollector(output, nodeIDs)
	}()

	return output, nil
}

// dataCollector collects data from monitored items
func (d *Driver) dataCollector(output chan<- TelemetryPoint, nodeIDs []string) {
	ticker := time.NewTicker(d.config.SubscriptionRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, nodeID := range nodeIDs {
				point := d.readNode(nodeID)
				if point != nil {
					select {
					case output <- *point:
					case <-context.Background().Done():
						return
					}
				}
			}
		case <-context.Background().Done():
			return
		}
	}
}

// readNode reads a single node value
func (d *Driver) readNode(nodeID string) *TelemetryPoint {
	// Get node info from cache or create placeholder
	d.mu.RLock()
	nodeInfo, exists := d.nodeCache[nodeID]
	d.mu.RUnlock()

	if !exists {
		nodeInfo = &NodeInfo{
			NodeID:      nodeID,
			BrowseName:  getBrowseName(nodeID),
			DisplayName: getDisplayName(nodeID),
			DataType:    "Double",
			AccessLevel: "CurrentRead",
		}
		d.mu.Lock()
		d.nodeCache[nodeID] = nodeInfo
		d.mu.Unlock()
	}

	// Generate realistic value based on node type
	value := generateNodeValue(nodeID)
	quality := "Good"

	point := &TelemetryPoint{
		ID:          uuid.New().String(),
		SourceID:    d.config.EndpointURL,
		NodeID:      nodeID,
		AssetID:     getAssetID(nodeID),
		AssetName:   nodeInfo.DisplayName,
		Value:       value,
		Quality:     quality,
		Protocol:    "OPCUA",
		Timestamp:   time.Now(),
		DataType:    nodeInfo.DataType,
		Unit:        getUnit(nodeID),
	}

	return point
}

// Browse performs node browsing on the OPC UA server
func (d *Driver) Browse(ctx context.Context, nodeID string) ([]NodeInfo, error) {
	d.mu.RLock()
	if !d.connected {
		d.mu.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	d.mu.RUnlock()

	d.logger.Info("Browsing OPC UA node",
		zap.String("node_id", nodeID))

	// Simulate node browsing
	// In production, this would call Browse() on the OPC UA client
	nodes := []NodeInfo{
		{
			NodeID:      fmt.Sprintf("%s.Temperature", nodeID),
			BrowseName:  "Temperature",
			DisplayName: "Temperature Sensor",
			DataType:    "Double",
			AccessLevel: "CurrentRead",
			Description: "Temperature measurement",
			ParentID:    nodeID,
			HasChildren: false,
		},
		{
			NodeID:      fmt.Sprintf("%s.Pressure", nodeID),
			BrowseName:  "Pressure",
			DisplayName: "Pressure Sensor",
			DataType:    "Double",
			AccessLevel: "CurrentRead",
			Description: "Pressure measurement",
			ParentID:    nodeID,
			HasChildren: false,
		},
		{
			NodeID:      fmt.Sprintf("%s.Flow", nodeID),
			BrowseName:  "Flow",
			DisplayName: "Flow Sensor",
			DataType:    "Double",
			AccessLevel: "CurrentRead",
			Description: "Flow measurement",
			ParentID:    nodeID,
			HasChildren: false,
		},
	}

	// Update cache
	for _, node := range nodes {
		d.mu.Lock()
		d.nodeCache[node.NodeID] = &node
		d.mu.Unlock()
	}

	return nodes, nil
}

// GetSubscriptionState returns the current subscription state
func (d *Driver) GetSubscriptionState() *Subscription {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.subscription
}

// GetNodeInfo returns cached node information
func (d *Driver) GetNodeInfo(nodeID string) *NodeInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.nodeCache[nodeID]
}

// GetAllNodeInfo returns all cached node information
func (d *Driver) GetAllNodeInfo() []NodeInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	nodes := make([]NodeInfo, 0, len(d.nodeCache))
	for _, node := range d.nodeCache {
		nodes = append(nodes, *node)
	}
	return nodes
}

// Helper functions

func getBrowseName(nodeID string) string {
	parts := strings.Split(nodeID, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return nodeID
}

func getDisplayName(nodeID string) string {
	parts := strings.Split(nodeID, ".")
	if len(parts) > 1 {
		return strings.Title(strings.ToLower(parts[len(parts)-1]))
	}
	return nodeID
}

func getAssetID(nodeID string) string {
	// Extract asset ID from node path
	// Format: ns=2;s=Channel1.Device1.AssetName.Node
	parts := strings.Split(nodeID, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.HasPrefix(parts[i], "Device") {
			return parts[i]
		}
	}
	return "Unknown"
}

func getUnit(nodeID string) string {
	nodeIDLower := strings.ToLower(nodeID)
	if strings.Contains(nodeIDLower, "temperature") || strings.Contains(nodeIDLower, "temp") {
		return "°C"
	}
	if strings.Contains(nodeIDLower, "pressure") || strings.Contains(nodeIDLower, "pres") {
		return "bar"
	}
	if strings.Contains(nodeIDLower, "flow") {
		return "m³/h"
	}
	if strings.Contains(nodeIDLower, "current") || strings.Contains(nodeIDLower, "amp") {
		return "A"
	}
	if strings.Contains(nodeIDLower, "voltage") || strings.Contains(nodeIDLower, "volt") {
		return "V"
	}
	if strings.Contains(nodeIDLower, "power") {
		return "kW"
	}
	return ""
}

func generateNodeValue(nodeID string) float64 {
	nodeIDLower := strings.ToLower(nodeID)
	seed := float64(time.Now().UnixNano() % 10000)

	if strings.Contains(nodeIDLower, "temperature") || strings.Contains(nodeIDLower, "temp") {
		// Temperature: typically 20-100°C for industrial processes
		return 20 + (seed % 80)
	}
	if strings.Contains(nodeIDLower, "pressure") || strings.Contains(nodeIDLower, "pres") {
		// Pressure: typically 0-10 bar
		return seed % 100 / 10
	}
	if strings.Contains(nodeIDLower, "flow") {
		// Flow: typically 0-100 m³/h
		return seed % 1000 / 10
	}
	if strings.Contains(nodeIDLower, "current") || strings.Contains(nodeIDLower, "amp") {
		// Current: typically 0-100 A
		return seed % 1000 / 10
	}
	if strings.Contains(nodeIDLower, "voltage") || strings.Contains(nodeIDLower, "volt") {
		// Voltage: typically 380-415V for industrial
		return 380 + (seed % 350 / 10)
	}
	if strings.Contains(nodeIDLower, "power") {
		// Power: typically 0-100 kW
		return seed % 1000 / 10
	}
	// Default value
	return seed % 100
}

// TelemetryPoint represents a single data point from OPC UA
type TelemetryPoint struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`
	NodeID      string                 `json:"node_id"`
	AssetID     string                 `json:"asset_id"`
	AssetName   string                 `json:"asset_name"`
	Value       float64                `json:"value"`
	Quality     string                 `json:"quality"`
	Protocol    string                 `json:"protocol"`
	Timestamp   time.Time              `json:"timestamp"`
	DataType    string                 `json:"data_type"`
	Unit        string                 `json:"unit"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// loadCertificate loads a certificate from file
func loadCertificate(path string) (*x509.Certificate, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	return x509.ParseCertificate(block.Bytes)
}

// loadPrivateKey loads a private key from file
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS8 first
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}

	// Try PKCS1
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// generateCertificate generates a self-signed certificate
func generateCertificate() (tls.Certificate, error) {
	// Generate RSA key
	privateKey, err := rsa.GenerateKey(ioutil.Discard, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Generate serial number
	serialNumber, err := rand.Int(ioutil.Discard, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Create template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"NEAM"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(ioutil.Discard, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to create certificate: %w", err)
	}

	return tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
	}, nil
}

func randInt(min, max int) int {
	return min + int(time.Now().UnixNano()%int64(max-min))
}

// Helper for rand.Int (placeholder)
func rand() *big.Int {
	n, _ := big.NewInt(0).SetString(hex.EncodeToString([]byte{
		byte(time.Now().UnixNano() & 0xFF),
		byte(time.Now().UnixNano() >> 8 & 0xFF),
		byte(time.Now().UnixNano() >> 16 & 0xFF),
		byte(time.Now().UnixNano() >> 24 & 0xFF),
	}), 16)
	return n
}
