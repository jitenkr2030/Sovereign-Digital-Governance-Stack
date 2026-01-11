package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/neamplatform/sensing/industrial/config"
	"go.uber.org/zap"
)

// Manager handles persistent sessions for protocol connections
type Manager struct {
	cfg       config.SessionConfig
	logger    *zap.Logger
	sessions  map[string]*Session
	stateStore StateStore
	mu        sync.RWMutex
}

// Session holds session state and configuration
type Session struct {
	ID          string
	ProtocolID  string
	State       SessionState
	RetryCount  int
	LastActivity time.Time
	CreatedAt   time.Time
	Config      config.SessionConfig
	connected   bool
	nodeIDs     []string
	subscriptionID int
	monitoredItems map[string]int
}

// SessionState represents the state of a session
type SessionState int

const (
	SessionDisconnected SessionState = iota
	SessionConnecting
	SessionConnected
	SessionReconnecting
	SessionFailed
)

// StateStore defines the interface for session state persistence
type StateStore interface {
	SaveSession(session *Session) error
	LoadSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	GetAllSessions() ([]*Session, error)
}

// MemoryStateStore implements StateStore in-memory
type MemoryStateStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewMemoryStateStore creates a memory-based state store
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		sessions: make(map[string]*Session),
	}
}

// RedisStateStore implements StateStore with Redis (placeholder)
type RedisStateStore struct {
	client    interface{}
	keyPrefix string
	mu        sync.RWMutex
}

// NewManager creates a new session manager
func NewManager(cfg config.SessionConfig, logger *zap.Logger) *Manager {
	var stateStore StateStore

	switch cfg.StateStoreType {
	case "redis":
		stateStore = NewRedisStateStore(cfg.StateStoreURL, "industrial-adapter")
	default:
		stateStore = NewMemoryStateStore()
	}

	return &Manager{
		cfg:       cfg,
		logger:    logger,
		sessions:  make(map[string]*Session),
		stateStore: stateStore,
	}
}

// Start initiates sessions for all protocols and node IDs
func (m *Manager) Start(ctx context.Context, protocols []ProtocolProvider, nodeIDs []string) <-chan TelemetryPoint {
	output := make(chan TelemetryPoint, 1000)

	// Create sessions for each protocol
	for _, protocol := range protocols {
		session := &Session{
			ID:             fmt.Sprintf("session-%s", protocol.ID()),
			ProtocolID:     protocol.ID(),
			State:          SessionDisconnected,
			Config:         m.cfg,
			CreatedAt:      time.Now(),
			nodeIDs:        nodeIDs,
			monitoredItems: make(map[string]int),
		}

		// Try to restore session from state store
		if restored, err := m.stateStore.LoadSession(session.ID); err == nil {
			session = restored
			m.logger.Info("Restored session from state store",
				zap.String("session_id", session.ID),
				zap.Int("subscription_id", session.subscriptionID))
		}

		m.mu.Lock()
		m.sessions[session.ID] = session
		m.mu.Unlock()

		m.logger.Info("Session created",
			zap.String("session_id", session.ID),
			zap.String("protocol_id", protocol.ID()))
	}

	// Start connection and data collection
	go m.runSessions(ctx, protocols, nodeIDs, output)

	return output
}

// runSessions manages all sessions
func (m *Manager) runSessions(ctx context.Context, protocols []ProtocolProvider, nodeIDs []string, output chan<- TelemetryPoint) {
	var wg sync.WaitGroup

	for _, protocol := range protocols {
		wg.Add(1)
		go func(p ProtocolProvider) {
			defer wg.Done()
			m.runSession(ctx, p, nodeIDs, output)
		}(protocol)
	}

	wg.Wait()
	close(output)
}

// runSession manages a single protocol session
func (m *Manager) runSession(ctx context.Context, protocol ProtocolProvider, nodeIDs []string, output chan<- TelemetryPoint) {
	session := m.getSession(protocol.ID())
	if session == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !session.connected {
			if err := m.connectSession(ctx, protocol, session); err != nil {
				m.logger.Error("Failed to connect session",
					zap.Error(err),
					zap.String("session_id", session.ID))
				m.waitBackoff(session)
				continue
			}
		}

		// Subscribe to node IDs
		dataChan, err := protocol.Subscribe(nodeIDs)
		if err != nil {
			m.logger.Error("Failed to subscribe",
				zap.Error(err),
				zap.String("session_id", session.ID))
			session.connected = false
			m.waitBackoff(session)
			continue
		}

		// Forward data
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-dataChan:
				if !ok {
					m.logger.Warn("Data channel closed",
						zap.String("session_id", session.ID))
					session.connected = false
					break
				}
				session.LastActivity = time.Now()
				
				// Update monitored items
				session.monitoredItems[data.PointID] = session.monitoredItems[data.PointID] + 1
				
				// Save session state periodically
				if session.monitoredItems[data.PointID]%100 == 0 {
					m.stateStore.SaveSession(session)
				}
				
				select {
				case output <- data:
				default:
					m.logger.Warn("Output channel full, dropping data",
						zap.String("session_id", session.ID))
				}
			}
		}
	}
}

// connectSession establishes a connection
func (m *Manager) connectSession(ctx context.Context, protocol ProtocolProvider, session *Session) error {
	m.mu.Lock()
	session.State = SessionConnecting
	m.mu.Unlock()

	if err := protocol.Connect(ctx); err != nil {
		m.mu.Lock()
		session.State = SessionFailed
		session.RetryCount++
		m.mu.Unlock()
		return fmt.Errorf("connection failed: %w", err)
	}

	m.mu.Lock()
	session.State = SessionConnected
	session.connected = true
	session.RetryCount = 0
	m.mu.Unlock()

	m.logger.Info("Session connected",
		zap.String("session_id", session.ID))

	return nil
}

// waitBackoff implements exponential backoff
func (m *Manager) waitBackoff(session *Session) {
	m.mu.Lock()
	delay := time.Duration(session.RetryCount) * time.Second
	if delay > time.Duration(m.cfg.MaxDelay) {
		delay = time.Duration(m.cfg.MaxDelay)
	}
	if delay < time.Duration(m.cfg.InitialDelay) {
		delay = time.Duration(m.cfg.InitialDelay)
	}
	m.mu.Unlock()

	time.Sleep(delay)
}

// getSession retrieves a session by protocol ID
func (m *Manager) getSession(protocolID string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		if session.ProtocolID == protocolID {
			return session
		}
	}
	return nil
}

// GetSession returns a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// GetAllSessions returns all sessions
func (m *Manager) GetAllSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Stats returns session statistics
func (m *Manager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_sessions"] = len(m.sessions)
	stats["connected_sessions"] = 0
	stats["disconnected_sessions"] = 0
	stats["failed_sessions"] = 0
	stats["monitored_items_total"] = 0

	for _, session := range m.sessions {
		switch session.State {
		case SessionConnected:
			stats["connected_sessions"] = stats["connected_sessions"].(int) + 1
		case SessionDisconnected:
			stats["disconnected_sessions"] = stats["disconnected_sessions"].(int) + 1
		case SessionFailed:
			stats["failed_sessions"] = stats["failed_sessions"].(int) + 1
		}
		stats["monitored_items_total"] = stats["monitored_items_total"].(int) + len(session.monitoredItems)
	}

	return stats
}

// Close closes all sessions
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, session := range m.sessions {
		session.connected = false
		m.stateStore.SaveSession(session)
	}

	m.sessions = make(map[string]*Session)

	m.logger.Info("Session manager closed")
	return nil
}

// MemoryStateStore methods
func (s *MemoryStateStore) SaveSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

func (s *MemoryStateStore) LoadSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

func (s *MemoryStateStore) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}

func (s *MemoryStateStore) GetAllSessions() ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// RedisStateStore methods (placeholder implementations)
func (s *RedisStateStore) SaveSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

func (s *RedisStateStore) LoadSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return nil, fmt.Errorf("Redis state store not configured")
}

func (s *RedisStateStore) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

func (s *RedisStateStore) GetAllSessions() ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return nil, nil
}

// NewRedisStateStore creates a Redis state store
func NewRedisStateStore(url, keyPrefix string) *RedisStateStore {
	return &RedisStateStore{
		keyPrefix: keyPrefix,
	}
}

// ProtocolProvider interface for protocol operations
type ProtocolProvider interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Subscribe(nodeIDs []string) (<-chan TelemetryPoint, error)
	Browse(ctx context.Context, nodeID string) ([]NodeInfo, error)
	HealthCheck() bool
	ID() string
}

// TelemetryPoint represents a telemetry data point
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

// NodeInfo placeholder for interface compatibility
type NodeInfo struct {
	NodeID      string
	BrowseName  string
	DisplayName string
	DataType    string
	AccessLevel string
	Description string
	ParentID    string
	HasChildren bool
}
