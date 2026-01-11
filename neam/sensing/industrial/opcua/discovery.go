package opcua

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/neamplatform/sensing/industrial/config"
	"go.uber.org/zap"
)

// AssetDiscovery performs industrial asset discovery
type AssetDiscovery struct {
	cfg    config.DiscoveryConfig
	logger *zap.Logger
	cache  map[string]*CachedAsset
	mu     sync.RWMutex
}

// CachedAsset represents a cached asset discovery result
type CachedAsset struct {
	Assets     []AssetInfo
	DiscoveryTime time.Time
	ExpiresAt   time.Time
}

// AssetInfo represents information about a discovered industrial asset
type AssetInfo struct {
	AssetID      string            `json:"asset_id"`
	AssetName    string            `json:"asset_name"`
	AssetType    string            `json:"asset_type"`
	Manufacturer string            `json:"manufacturer"`
	Model        string            `json:"model"`
	SerialNumber string            `json:"serial_number"`
	Nodes        []NodeInfo        `json:"nodes"`
	Metadata     map[string]string `json:"metadata"`
}

// NewAssetDiscovery creates a new asset discovery instance
func NewAssetDiscovery(cfg config.DiscoveryConfig, logger *zap.Logger) *AssetDiscovery {
	return &AssetDiscovery{
		cfg:    cfg,
		logger: logger,
		cache:  make(map[string]*CachedAsset),
	}
}

// Discover performs asset discovery using the provided protocol drivers
func (d *AssetDiscovery) Discover(ctx context.Context, protocols []ProtocolClient) ([]AssetInfo, error) {
	d.mu.RLock()
	cached := d.getCached()
	d.mu.RUnlock()

	// Return cached result if still valid
	if cached != nil {
		d.logger.Debug("Using cached discovery results")
		return cached.Assets, nil
	}

	// Perform discovery
	assets, err := d.performDiscovery(ctx, protocols)
	if err != nil {
		return nil, err
	}

	// Cache results
	d.mu.Lock()
	d.cache["default"] = &CachedAsset{
		Assets:        assets,
		DiscoveryTime: time.Now(),
		ExpiresAt:     time.Now().Add(time.Duration(d.cfg.CacheTimeout)),
	}
	d.mu.Unlock()

	return assets, nil
}

// performDiscovery performs actual asset discovery
func (d *AssetDiscovery) performDiscovery(ctx context.Context, protocols []ProtocolClient) ([]AssetInfo, error) {
	var allAssets []AssetInfo

	for _, protocol := range protocols {
		driverAssets, err := d.discoverProtocolAssets(ctx, protocol)
		if err != nil {
			d.logger.Warn("Asset discovery failed for protocol",
				zap.Error(err),
				zap.String("protocol_id", protocol.ID()))
			continue
		}
		allAssets = append(allAssets, driverAssets...)
	}

	d.logger.Info("Asset discovery completed",
		zap.Int("total_assets", len(allAssets)))

	return allAssets, nil
}

// discoverProtocolAssets discovers assets from a single protocol driver
func (d *AssetDiscovery) discoverProtocolAssets(ctx context.Context, protocol ProtocolClient) ([]AssetInfo, error) {
	var assets []AssetInfo

	// Start from root nodes
	rootNodes := []string{
		"ns=0;i=84",        // Objects folder
		"ns=0;i=85",        // Types folder
	}

	for _, rootNode := range rootNodes {
		nodeAssets, err := d.browseNodeTree(ctx, protocol, rootNode, 0)
		if err != nil {
			d.logger.Debug("Failed to browse root node",
				zap.Error(err),
				zap.String("node_id", rootNode))
			continue
		}
		assets = append(assets, nodeAssets...)
	}

	return assets, nil
}

// browseNodeTree recursively browses node tree
func (d *AssetDiscovery) browseNodeTree(ctx context.Context, protocol ProtocolClient, nodeID string, depth int) ([]AssetInfo, error) {
	if depth >= d.cfg.BrowseDepth {
		return nil, nil
	}

	var assets []AssetInfo

	// Browse current node
	nodes, err := protocol.Browse(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		// Check if node matches filter criteria
		if d.matchesFilter(node) {
			asset := d.nodeToAsset(node, protocol.ID())
			assets = append(assets, asset)
		}

		// Recursively browse children if enabled
		if d.cfg.IncludeChildren && node.HasChildren {
			childAssets, err := d.browseNodeTree(ctx, protocol, node.NodeID, depth+1)
			if err != nil {
				d.logger.Warn("Failed to browse child node",
					zap.Error(err),
					zap.String("node_id", node.NodeID))
				continue
			}
			assets = append(assets, childAssets...)
		}
	}

	return assets, nil
}

// matchesFilter checks if node matches discovery filters
func (d *AssetDiscovery) matchesFilter(node NodeInfo) bool {
	// If no filters specified, match all
	if len(d.cfg.FilterByType) == 0 {
		return true
	}

	// Check if node data type matches filter
	for _, filterType := range d.cfg.FilterByType {
		if strings.Contains(strings.ToLower(node.DataType), strings.ToLower(filterType)) {
			return true
		}
	}

	return false
}

// nodeToAsset converts a NodeInfo to an AssetInfo
func (d *AssetDiscovery) nodeToAsset(node NodeInfo, protocolID string) AssetInfo {
	// Extract asset information from node
	assetID := extractAssetID(node.NodeID)
	assetName := node.DisplayName
	assetType := inferAssetType(node.NodeID)

	return AssetInfo{
		AssetID:      assetID,
		AssetName:    assetName,
		AssetType:    assetType,
		Manufacturer: inferManufacturer(node.NodeID),
		Model:        inferModel(node.NodeID),
		SerialNumber: "",
		Nodes:        []NodeInfo{node},
		Metadata: map[string]string{
			"protocol_id":  protocolID,
			"node_id":      node.NodeID,
			"browse_name":  node.BrowseName,
			"data_type":    node.DataType,
			"access_level": node.AccessLevel,
		},
	}
}

// getCached returns cached discovery results if still valid
func (d *AssetDiscovery) getCached() *CachedAsset {
	cached, exists := d.cache["default"]
	if !exists {
		return nil
	}

	if time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached
}

// InvalidateCache clears the discovery cache
func (d *AssetDiscovery) InvalidateCache() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[string]*CachedAsset)
	d.logger.Info("Discovery cache invalidated")
}

// GetCacheStatus returns cache status information
func (d *AssetDiscovery) GetCacheStatus() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	status := make(map[string]interface{})
	cached := d.cache["default"]
	if cached != nil {
		status["cached"] = true
		status["discovery_time"] = cached.DiscoveryTime
		status["expires_at"] = cached.ExpiresAt
		status["assets_count"] = len(cached.Assets)
	} else {
		status["cached"] = false
	}

	return status
}

// Helper functions for asset extraction

func extractAssetID(nodeID string) string {
	// Format: ns=2;s=Channel1.Device1.AssetName.SubNode
	parts := strings.Split(nodeID, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if strings.HasPrefix(part, "Device") || strings.HasPrefix(part, "PLC") ||
		   strings.HasPrefix(part, "Robot") || strings.HasPrefix(part, "CNC") {
			return part
		}
	}
	// Fallback to node ID
	return strings.ReplaceAll(nodeID, ":", "_")
}

func inferAssetType(nodeID string) string {
	nodeIDLower := strings.ToLower(nodeID)
	if strings.Contains(nodeIDLower, "robot") {
		return "robot"
	}
	if strings.Contains(nodeIDLower, "cnc") || strings.Contains(nodeIDLower, "machine") {
		return "cnc_machine"
	}
	if strings.Contains(nodeIDLower, "plc") {
		return "plc"
	}
	if strings.Contains(nodeIDLower, "conveyor") {
		return "conveyor"
	}
	if strings.Contains(nodeIDLower, "sensor") {
		return "sensor"
	}
	return "unknown"
}

func inferManufacturer(nodeID string) string {
	nodeIDLower := strings.ToLower(nodeID)
	if strings.Contains(nodeIDLower, "siemens") {
		return "Siemens"
	}
	if strings.Contains(nodeIDLower, "allen") || strings.Contains(nodeIDLower, "bradley") {
		return "Allen-Bradley"
	}
	if strings.Contains(nodeIDLower, "schneider") {
		return "Schneider Electric"
	}
	if strings.Contains(nodeIDLower, "abb") {
		return "ABB"
	}
	return "Unknown"
}

func inferModel(nodeID string) string {
	parts := strings.Split(nodeID, ".")
	for _, part := range parts {
		if strings.HasPrefix(part, "S7-") {
			return part
		}
		if strings.HasPrefix(part, "Logix") {
			return part
		}
	}
	return ""
}

// ProtocolClient interface for asset discovery
type ProtocolClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Subscribe(nodeIDs []string) (<-chan TelemetryPoint, error)
	Browse(ctx context.Context, nodeID string) ([]NodeInfo, error)
	HealthCheck() bool
	ID() string
	GetProtocolType() string
}

// TelemetryPoint placeholder for compatibility
type TelemetryPoint struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`
	NodeID      string                 `json:"node_id"`
	Value       interface{}            `json:"value"`
	Quality     string                 `json:"quality"`
	Protocol    string                 `json:"protocol"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
