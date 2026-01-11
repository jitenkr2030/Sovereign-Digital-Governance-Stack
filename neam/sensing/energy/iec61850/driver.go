package iec61850

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DriverConfig contains IEC 61850/MMS driver configuration
type DriverConfig struct {
	EndPoint     string
	Port         int
	IEDName      string
	APTitle      string
	AEQualifier  int
	Timeout      time.Duration
	RetryInterval time.Duration
}

// Driver implements IEC 61850/MMS protocol client
type Driver struct {
	config    DriverConfig
	logger    *zap.Logger
	connected bool
	connection *MMSConnection
	reportCB  *ReportControlBlock
	mu        sync.RWMutex
}

// NewDriver creates a new IEC 61850 driver instance
func NewDriver(config DriverConfig, logger *zap.Logger) *Driver {
	return &Driver{
		config: config,
		logger: logger,
	}
}

// ID returns the driver identifier
func (d *Driver) ID() string {
	return fmt.Sprintf("IEC61850-%s-%s", d.config.IEDName, d.config.EndPoint)
}

// GetProtocolType returns the protocol type
func (d *Driver) GetProtocolType() string {
	return "IEC61850"
}

// Connect establishes an IEC 61850/MMS connection
func (d *Driver) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return nil
	}

	d.logger.Info("Connecting to IEC 61850 endpoint",
		zap.String("endpoint", d.config.EndPoint),
		zap.Int("port", d.config.Port),
		zap.String("ied_name", d.config.IEDName))

	// Create MMS connection
	conn := NewMMSConnection(MMSConnectionConfig{
		EndPoint:    d.config.EndPoint,
		Port:        d.config.Port,
		APTitle:     d.config.APTitle,
		AEQualifier: d.config.AEQualifier,
		Timeout:     d.config.Timeout,
	}, d.logger)

	// Establish MMS connection
	if err := conn.Connect(ctx); err != nil {
		return fmt.Errorf("failed to establish MMS connection: %w", err)
	}

	// Associate with IED
	if err := conn.Associate(ctx, d.config.IEDName); err != nil {
		conn.Disconnect()
		return fmt.Errorf("failed to associate with IED: %w", err)
	}

	// Setup report control block if configured
	if d.config.ReportCBName != "" {
		rcb, err := conn.SetupReportControlBlock(ctx, d.config.ReportCBName)
		if err != nil {
			d.logger.Warn("Failed to setup report control block",
				zap.Error(err),
				zap.String("rcb_name", d.config.ReportCBName))
		} else {
			d.reportCB = rcb
		}
	}

	d.connection = conn
	d.connected = true

	d.logger.Info("IEC 61850 connection established",
		zap.String("ied_name", d.config.IEDName))

	return nil
}

// Disconnect closes the IEC 61850 connection
func (d *Driver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil
	}

	if d.reportCB != nil && d.reportCB.Enabled {
		if d.connection != nil {
			d.connection.DisableReportControlBlock(context.Background(), d.config.ReportCBName)
		}
	}

	if d.connection != nil {
		d.connection.Disconnect()
	}

	d.connected = false
	d.logger.Info("IEC 61850 connection closed",
		zap.String("ied_name", d.config.IEDName))

	return nil
}

// HealthCheck verifies connection health
func (d *Driver) HealthCheck() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.connected {
		return false
	}

	return d.connection != nil && d.connection.Healthy()
}

// Subscribe initiates subscription for specified data points
func (d *Driver) Subscribe(pointIDs []string) (<-chan TelemetryPoint, error) {
	d.mu.RLock()
	if !d.connected || d.connection == nil {
		d.mu.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	d.mu.RUnlock()

	output := make(chan TelemetryPoint, 100)

	// Start reading from data model
	dataChan, err := d.connection.ReadDataPoints(context.Background(), pointIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to start data reading: %w", err)
	}

	// Process incoming data
	go func() {
		defer close(output)

		for data := range dataChan {
			point := TelemetryPoint{
				ID:        uuid.New().String(),
				SourceID:  fmt.Sprintf("%s/%s", d.config.IEDName, d.config.EndPoint),
				PointID:   data.Path,
				Value:     data.Value,
				Quality:   data.Quality,
				Protocol:  "IEC61850",
				Timestamp: data.Timestamp,
				Metadata:  data.Metadata,
			}
			select {
			case output <- point:
			case <-context.Background().Done():
				return
			}
		}
	}()

	return output, nil
}

// MMSConnection represents an MMS protocol connection
type MMSConnection struct {
	config     MMSConnectionConfig
	logger     *zap.Logger
	connected  bool
	association *Association
}

// MMSConnectionConfig contains MMS connection configuration
type MMSConnectionConfig struct {
	EndPoint    string
	Port        int
	APTitle     string
	AEQualifier int
	Timeout     time.Duration
}

// NewMMSConnection creates a new MMS connection instance
func NewMMSConnection(config MMSConnectionConfig, logger *zap.Logger) *MMSConnection {
	return &MMSConnection{
		config: config,
		logger: logger,
	}
}

// Connect establishes the MMS connection
func (m *MMSConnection) Connect(ctx context.Context) error {
	m.logger.Info("Establishing MMS connection",
		zap.String("endpoint", m.config.EndPoint),
		zap.Int("port", m.config.Port))

	// Simulate connection establishment
	m.connected = true
	m.association = &Association{
		Initiated: true,
		IedName:   "IED",
	}

	return nil
}

// Associate associates with an IED
func (m *MMSConnection) Associate(ctx context.Context, iedName string) error {
	m.logger.Info("Associating with IED",
		zap.String("ied_name", iedName))

	if m.association != nil {
		m.association.IedName = iedName
	}

	return nil
}

// Disconnect closes the MMS connection
func (m *MMSConnection) Disconnect() {
	if m.connected {
		m.logger.Info("Closing MMS connection")
		m.connected = false
		m.association = nil
	}
}

// Healthy checks connection health
func (m *MMSConnection) Healthy() bool {
	return m.connected
}

// ReadDataPoints reads specified data points from the data model
func (m *MMSConnection) ReadDataPoints(ctx context.Context, pointIDs []string) (<-chan DataPoint, error) {
	output := make(chan DataPoint, 100)

	go func() {
		defer close(output)
		ticker := time.NewTicker(100 * time.Millisecond) // IEC 61850 typically 10Hz
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, pointID := range pointIDs {
					// Parse the IEC 61850 object reference
					parsed := parseObjectReference(pointID)

					dataPoint := DataPoint{
						Path:      pointID,
						Timestamp: time.Now(),
						Quality:   "Good",
						Metadata:  nil,
					}

					// Generate realistic values based on object type
					switch parsed.ObjectType {
					case "MMXU": // Measurement unit
						dataPoint.Value = generateMMXUValue(parsed)
					case "XCBR": // Circuit breaker
						dataPoint.Value = generateXCBRValue(parsed)
					case "CSWI": // Switch controller
						dataPoint.Value = generateCSWIValue(parsed)
					case "PTOC": // Protection
						dataPoint.Value = generatePTOCValue(parsed)
					default:
						dataPoint.Value = 0.0
					}

					select {
					case output <- dataPoint:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return output, nil
}

// ObjectReference represents a parsed IEC 61850 object reference
type ObjectReference struct {
	IEDName    string
	LNPrefix   string
	LNClass    string
	LNInstance int
	ObjectType string
	FC         string // Functional Constraint
}

// parseObjectReference parses an IEC 61850 object reference
func parseObjectReference(ref string) ObjectReference {
	parts := strings.Split(ref, ".")

	result := ObjectReference{}

	if len(parts) >= 3 {
		// Extract IED name
		if !strings.Contains(parts[0], "/") {
			result.IEDName = parts[0]
			parts = parts[1:]
		}

		// Extract logical node (LN)
		ln := parts[0]
		if len(ln) >= 4 {
			result.LNPrefix = ln[0:1]
			result.LNClass = ln[1:4]
			result.LNInstance = int(ln[3] - '0')
		}

		// Extract object type and FC
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "$") {
				result.FC = strings.TrimPrefix(part, "$")
			} else if len(part) >= 4 {
				result.ObjectType = part
			}
		}
	}

	return result
}

// SetupReportControlBlock configures a report control block
func (m *MMSConnection) SetupReportControlBlock(ctx context.Context, rcbName string) (*ReportControlBlock, error) {
	m.logger.Info("Setting up report control block",
		zap.String("rcb_name", rcbName))

	return &ReportControlBlock{
		Name:    rcbName,
		Enabled: true,
		Period:  1000, // 1 second reporting
	}, nil
}

// DisableReportControlBlock disables a report control block
func (m *MMSConnection) DisableReportControlBlock(ctx context.Context, rcbName string) error {
	m.logger.Info("Disabling report control block",
		zap.String("rcb_name", rcbName))
	return nil
}

// Association represents an MMS association
type Association struct {
	Initiated bool
	IedName   string
}

// ReportControlBlock represents an IEC 61850 report control block
type ReportControlBlock struct {
	Name    string
	Enabled bool
	Period  int // milliseconds
}

// DataPoint represents a single data point from IEC 61850
type DataPoint struct {
	Path      string
	Value     float64
	Timestamp time.Time
	Quality   string
	Metadata  []byte
}

// generateMMXUValue generates realistic measurement values
func generateMMXUValue(parsed ObjectReference) float64 {
	baseValues := map[string]float64{
		"Vol": 230.0,  // Phase voltage
		"A":   100.0,  // Current
		"W":   23000.0, // Active power
		"VA":  25000.0, // Apparent power
		"VAr": 5000.0,  // Reactive power
		"Hz":  50.0,   // Frequency
		"PF":  0.95,   // Power factor
	}

	base := baseValues[parsed.ObjectType]
	if base == 0 {
		base = 100.0
	}

	// Add small random variation for realism
	variation := (rand.Float64() - 0.5) * 0.02 // ±1%
	return base * (1 + variation)
}

// generateXCBRValue generates circuit breaker status
func generateXCBRValue(parsed ObjectReference) float64 {
	// Circuit breaker position: 0 = intermediate, 1 = off, 2 = on
	return 2.0 // Typically on
}

// generateCSWIValue generates switch controller status
func generateCSWIValue(parsed ObjectReference) float64 {
	// Switch controller position
	return 1.0 // On
}

// generatePTOCValue generates protection relay values
func generatePTOCValue(parsed ObjectReference) float64 {
	// Protection operational values (currents, times)
	if strings.Contains(parsed.ObjectType, "Str") {
		return 0.0 // Pickup setting, typically 0 until triggered
	}
	return 0.0
}
