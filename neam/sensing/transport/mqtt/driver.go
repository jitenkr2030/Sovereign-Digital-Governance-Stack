package mqtt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DriverConfig contains MQTT driver configuration
type DriverConfig struct {
	BrokerURL        string
	ClientID         string
	Username         string
	Password         string
	CleanStart       bool
	KeepAlive        time.Duration
	SessionExpiry    time.Duration
	QoSByCriticality map[string]int
}

// Driver implements MQTT v5 protocol client
type Driver struct {
	config        DriverConfig
	logger        *zap.Logger
	connected     bool
	client        *paho.Client
	subscriptions map[string]SubscriptionConfig
	sessionState  *SessionState
	mu            sync.RWMutex
	wg            sync.WaitGroup
}

// SessionState holds MQTT session persistence state
type SessionState struct {
	ClientID         string
	Subscriptions    map[string]int // topic -> QoS
	MessageID        uint16
	LastConnected    time.Time
	ServerKeepAlive  int
	SessionExpiry    int
	ReceivedMessages []ReceivedMessage
	mu               sync.RWMutex
}

// ReceivedMessage represents a message received during disconnection
type ReceivedMessage struct {
	Topic     string
	Payload   []byte
	QoS       int
	Timestamp time.Time
}

// SubscriptionConfig holds subscription configuration
type SubscriptionConfig struct {
	Topic      string
	QoS        int
	Criticality string
	RetainHandling int
	NoLocal    bool
}

// NewDriver creates a new MQTT driver instance
func NewDriver(config DriverConfig, logger *zap.Logger) *Driver {
	return &Driver{
		config: config,
		logger: logger,
		subscriptions: make(map[string]SubscriptionConfig),
		sessionState: &SessionState{
			ClientID:      config.ClientID,
			Subscriptions: make(map[string]int),
		},
	}
}

// ID returns the driver identifier
func (d *Driver) ID() string {
	return fmt.Sprintf("MQTT-%s", d.config.BrokerURL)
}

// GetProtocolType returns the protocol type
func (d *Driver) GetProtocolType() string {
	return "MQTTv5"
}

// Connect establishes an MQTT v5 connection
func (d *Driver) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return nil
	}

	d.logger.Info("Connecting to MQTT broker",
		zap.String("broker", d.config.BrokerURL),
		zap.String("client_id", d.config.ClientID),
		zap.Bool("clean_start", d.config.CleanStart))

	// Generate client ID if not provided
	if d.config.ClientID == "" {
		d.config.ClientID = generateClientID()
	}

	// Create MQTT client
	d.client = paho.NewClient(paho.ClientConfig{
		ClientID: d.config.ClientID,
		Username: d.config.Username,
		Password: d.config.Password,
		CleanStart: d.config.CleanStart,
		KeepAlive: uint16(d.config.KeepAlive.Seconds()),
		ConnectTimeout: d.config.Timeout,
		MaxReconnectDelay: d.config.MaxReconnectDelay,
		OnServerConnect: d.onServerConnect,
		OnConnectionLost: d.onConnectionLost,
	})

	// Parse broker URL
	protocol := "tcp"
	addr := d.config.BrokerURL
	if strings.HasPrefix(addr, "mqtt://") {
		addr = strings.TrimPrefix(addr, "mqtt://")
	} else if strings.HasPrefix(addr, "ssl://") || strings.HasPrefix(addr, "tls://") {
		protocol = "ssl"
		addr = strings.TrimPrefix(addr, "ssl://")
		addr = strings.TrimPrefix(addr, "tls://")
	}

	// Establish connection
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := paho.Connect(connCtx, &paho.Connect{
		Host:      addr,
		ClientID:  d.config.ClientID,
		Username:  d.config.Username,
		Password:  d.config.Password,
		CleanStart: d.config.CleanStart,
		KeepAlive: uint16(d.config.KeepAlive.Seconds()),
		Properties: &paho.ConnectProperties{
			SessionExpiryInterval: uint32(d.config.SessionExpiry.Seconds()),
			AuthenticationMethod:  "oauth2",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to establish MQTT connection: %w", err)
	}

	d.client.Conn = conn
	d.connected = true
	d.sessionState.LastConnected = time.Now()

	d.logger.Info("MQTT connection established",
		zap.String("broker", d.config.BrokerURL),
		zap.String("client_id", d.config.ClientID))

	return nil
}

// onServerConnect handles server connection acknowledgment
func (d *Driver) onServerConnect(connack *paho.Connack) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if connack.ReasonCode == 0 {
		d.connected = true
		d.sessionState.ServerKeepAlive = int(connack.Properties.ServerKeepAlive)
		d.logger.Info("MQTT server acknowledged connection",
			zap.Uint8("reason_code", connack.ReasonCode),
			zap.Int("server_keep_alive", d.sessionState.ServerKeepAlive))
	}
}

// onConnectionLost handles connection loss
func (d *Driver) onConnectionLost(err error) {
	d.mu.Lock()
	d.connected = false
	d.mu.Unlock()

	d.logger.Warn("MQTT connection lost",
		zap.Error(err),
		zap.String("client_id", d.config.ClientID))
}

// Disconnect closes the MQTT connection
func (d *Driver) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil
	}

	// Send disconnect packet with session expiry
	if d.client != nil && d.client.Conn != nil {
		d.client.Conn.Disconnect(&paho.Disconnect{
			ReasonCode: 0,
			Properties: &paho.DisconnectProperties{
				ReasonString: "Client disconnect",
			},
		})
	}

	d.connected = false
	d.logger.Info("MQTT connection closed",
		zap.String("broker", d.config.BrokerURL))

	return nil
}

// HealthCheck verifies connection health
func (d *Driver) HealthCheck() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.connected {
		return false
	}

	return d.client != nil && d.client.Conn != nil && d.client.Conn.IsConnected()
}

// Subscribe initiates subscription for specified topics
func (d *Driver) Subscribe(topicPatterns []string) (<-chan TelemetryPoint, error) {
	d.mu.RLock()
	if !d.connected || d.client == nil {
		d.mu.RUnlock()
		return nil, fmt.Errorf("not connected")
	}
	d.mu.RUnlock()

	// Create subscription channel
	output := make(chan TelemetryPoint, 1000)

	// Setup subscriptions
	subscriptions := make([]*paho.Subscribe, 0, len(topicPatterns))
	for _, topic := range topicPatterns {
		qos := d.getQoSForTopic(topic)
		sub := &paho.Subscribe{
			Topic: topic,
			QoS:   byte(qos),
			Options: &paho.SubscribeOptions{
				NoLocal:          false,
				RetainAsPublished: false,
				RetainHandling:   0,
			},
		}
		subscriptions = append(subscriptions, sub)

		d.subscriptions[topic] = SubscriptionConfig{
			Topic:       topic,
			QoS:         qos,
			Criticality: d.getCriticalityForTopic(topic),
		}

		d.sessionState.Subscriptions[topic] = qos

		d.logger.Info("Subscribing to topic",
			zap.String("topic", topic),
			zap.Int("qos", qos))
	}

	// Send subscribe packet
	subCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suback, err := d.client.Subscribe(subCtx, &paho.Subscribe{
		Subscriptions: subscriptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	if suback.Reasons[0] != 0x00 && suback.Reasons[0] != 0x01 && suback.Reasons[0] != 0x02 {
		return nil, fmt.Errorf("subscribe failed with reason code: %d", suback.Reasons[0])
	}

	// Register message handler
	d.client.RegisterMessageHandler(func(msg *paho.Message) {
		point := d.processMessage(msg)
		if point != nil {
			select {
			case output <- *point:
			case <-context.Background().Done():
				return
			}
		}
	})

	// Start processing messages
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.messageProcessor(output)
	}()

	return output, nil
}

// messageProcessor processes incoming messages
func (d *Driver) messageProcessor(output chan<- TelemetryPoint) {
	<-context.Background().Done()
}

// processMessage converts MQTT message to TelemetryPoint
func (d *Driver) processMessage(msg *paho.Message) *TelemetryPoint {
	// Parse payload based on content type
	var value interface{}
	contentType := msg.Properties.ContentType

	switch contentType {
	case "application/json":
		var jsonData interface{}
		if err := json.Unmarshal(msg.Payload, &jsonData); err != nil {
			value = string(msg.Payload)
		} else {
			value = jsonData
		}
	default:
		// Try to parse as number
		var num float64
		if _, err := fmt.Sscanf(string(msg.Payload), "%f", &num); err == nil {
			value = num
		} else {
			value = string(msg.Payload)
		}
	}

	// Determine criticality from topic
	criticality := d.getCriticalityForTopic(msg.Topic)

	// Extract metadata from user properties
	metadata := make(map[string]interface{})
	for k, v := range msg.Properties.User {
		metadata[k] = v
	}

	metadataBytes, _ := json.Marshal(metadata)

	point := &TelemetryPoint{
		ID:          uuid.New().String(),
		SourceID:    d.config.BrokerURL,
		PointID:     msg.Topic,
		Value:       value,
		Quality:     d.getQualityForMessage(msg),
		Protocol:    "MQTTv5",
		Timestamp:   time.Now(),
		Criticality: criticality,
		Metadata:    metadataBytes,
	}

	return point
}

// getQoSForTopic determines QoS level for a topic
func (d *Driver) getQoSForTopic(topic string) int {
	// Check for criticality indicators in topic
	topicLower := strings.ToLower(topic)
	if strings.Contains(topicLower, "critical") || strings.Contains(topicLower, "alert") || strings.Contains(topicLower, "emergency") {
		if qos, ok := d.config.QoSByCriticality["critical"]; ok {
			return qos
		}
		return 2 // Default to exactly-once for critical
	}

	if strings.Contains(topicLower, "important") || strings.Contains(topicLower, "warning") {
		if qos, ok := d.config.QoSByCriticality["important"]; ok {
			return qos
		}
		return 1 // Default to at-least-once for important
	}

	if strings.Contains(topicLower, "background") || strings.Contains(topicLower, "debug") {
		if qos, ok := d.config.QoSByCriticality["background"]; ok {
			return qos
		}
		return 0 // Default to at-most-once for background
	}

	// Default to standard QoS
	return 1
}

// getCriticalityForTopic determines criticality level for a topic
func (d *Driver) getCriticalityForTopic(topic string) string {
	topicLower := strings.ToLower(topic)
	if strings.Contains(topicLower, "critical") || strings.Contains(topicLower, "alert") || strings.Contains(topicLower, "emergency") {
		return "critical"
	}
	if strings.Contains(topicLower, "warning") || strings.Contains(topicLower, "important") {
		return "important"
	}
	if strings.Contains(topicLower, "debug") || strings.Contains(topicLower, "background") {
		return "background"
	}
	return "standard"
}

// getQualityForMessage determines data quality from message properties
func (d *Driver) getQualityForMessage(msg *paho.Message) string {
	// Check for quality indicators in user properties
	if q, ok := msg.Properties.User["quality"]; ok {
		return q
	}
	// Check for retained message flag
	if msg.Retain {
		return "Retained"
	}
	return "Good"
}

// GetSessionState returns the current session state
func (d *Driver) GetSessionState() SessionState {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return *d.sessionState
}

// generateClientID generates a unique client ID
func generateClientID() string {
	id := make([]byte, 12)
	rand.Read(id)
	return fmt.Sprintf("neam-%s", hex.EncodeToString(id))
}

// SetTimeout sets connection timeout
func (d *Driver) SetTimeout(timeout time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.Timeout = timeout
}

// SetMaxReconnectDelay sets maximum reconnect delay
func (d *Driver) SetMaxReconnectDelay(delay time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.RetryInterval = delay
}

// TelemetryPoint represents a single data point from MQTT
type TelemetryPoint struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`
	PointID     string                 `json:"point_id"`
	Value       interface{}            `json:"value"`
	Quality     string                 `json:"quality"`
	Protocol    string                 `json:"protocol"`
	Timestamp   time.Time              `json:"timestamp"`
	Criticality string                 `json:"criticality"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
