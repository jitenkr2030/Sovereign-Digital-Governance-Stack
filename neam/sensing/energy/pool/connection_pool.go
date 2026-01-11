package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PoolConfig contains connection pool configuration
type PoolConfig struct {
	MaxIdleConnections    int
	MaxActiveConnections  int
	MaxConnectionLifetime time.Duration
	IdleTimeout           time.Duration
	HealthCheckInterval   time.Duration
}

// Connection represents a pooled connection
type Connection interface {
	ID() string
	HealthCheck() bool
	Connect(ctx context.Context) error
	Disconnect() error
	GetProtocolType() string
}

// Pool manages a pool of connections
type Pool struct {
	config     PoolConfig
	logger     *zap.Logger
	conns      []Connection
	activeMap  map[string]Connection
	idleChan   chan Connection
	mu         sync.RWMutex
	closed     bool
	wg         sync.WaitGroup
}

// PoolStats holds pool statistics
type PoolStats struct {
	TotalConnections  int
	ActiveConnections int
	IdleConnections   int
	HealthChecks      int64
	ConnectionErrors  int64
	LastHealthCheck   time.Time
}

// NewPool creates a new connection pool
func NewPool(config PoolConfig, logger *zap.Logger) *Pool {
	pool := &Pool{
		config:    config,
		logger:    logger,
		conns:     make([]Connection, 0),
		activeMap: make(map[string]Connection),
		idleChan:  make(chan Connection, config.MaxIdleConnections),
	}

	if config.HealthCheckInterval > 0 {
		go pool.healthCheckLoop()
	}

	return pool
}

// AddConnection adds a new connection to the pool
func (p *Pool) AddConnection(conn Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("pool is closed")
	}

	if err := conn.Connect(context.Background()); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	p.conns = append(p.conns, conn)
	p.activeMap[conn.ID()] = conn

	p.logger.Info("Connection added to pool",
		zap.String("connection_id", conn.ID()),
		zap.String("protocol", conn.GetProtocolType()))

	return nil
}

// GetConnection retrieves an active connection
func (p *Pool) GetConnection(id string) (Connection, error) {
	p.mu.RLock()
	conn, exists := p.activeMap[id]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	if !conn.HealthCheck() {
		return nil, fmt.Errorf("connection unhealthy: %s", id)
	}

	return conn, nil
}

// GetConnectionByProtocol retrieves an active connection by protocol type
func (p *Pool) GetConnectionByProtocol(protocol string) (Connection, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, conn := range p.activeMap {
		if conn.GetProtocolType() == protocol && conn.HealthCheck() {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("no healthy connection found for protocol: %s", protocol)
}

// ReleaseConnection returns a connection to the idle pool
func (p *Pool) ReleaseConnection(conn Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return conn.Disconnect()
	}

	if !conn.HealthCheck() {
		p.logger.Warn("Releasing unhealthy connection, disconnecting",
			zap.String("connection_id", conn.ID()))
		conn.Disconnect()
		return nil
	}

	select {
	case p.idleChan <- conn:
		p.logger.Debug("Connection released to idle pool",
			zap.String("connection_id", conn.ID()))
	default:
		p.logger.Debug("Idle pool full, disconnecting connection",
			zap.String("connection_id", conn.ID()))
		conn.Disconnect()
	}

	return nil
}

// RemoveConnection removes a connection from the pool
func (p *Pool) RemoveConnection(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, exists := p.activeMap[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	delete(p.activeMap, id)

	for i, c := range p.conns {
		if c.ID() == id {
			p.conns = append(p.conns[:i], p.conns[i+1:]...)
			break
		}
	}

	conn.Disconnect()

	p.logger.Info("Connection removed from pool",
		zap.String("connection_id", id))

	return nil
}

// GetAllConnections returns all connections in the pool
func (p *Pool) GetAllConnections() []Connection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	connections := make([]Connection, len(p.conns))
	copy(connections, p.conns)
	return connections
}

// GetActiveConnections returns all active connections
func (p *Pool) GetActiveConnections() []Connection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	connections := make([]Connection, 0, len(p.activeMap))
	for _, conn := range p.activeMap {
		connections = append(connections, conn)
	}
	return connections
}

// GetIdleConnection retrieves an idle connection if available
func (p *Pool) GetIdleConnection() (Connection, error) {
	select {
	case conn := <-p.idleChan:
		return conn, nil
	default:
		return nil, fmt.Errorf("no idle connections available")
	}
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		TotalConnections:  len(p.conns),
		ActiveConnections: len(p.activeMap),
		IdleConnections:   len(p.idleChan),
	}

	return stats
}

// healthCheckLoop periodically checks connection health
func (p *Pool) healthCheckLoop() {
	p.wg.Add(1)
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performHealthCheck()
		case <-p.closedChan():
			return
		}
	}
}

// performHealthCheck checks health of all connections
func (p *Pool) performHealthCheck() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, conn := range p.activeMap {
		if !conn.HealthCheck() {
			p.logger.Warn("Connection failed health check",
				zap.String("connection_id", conn.ID()))
		}
	}
}

// closedChan returns a channel that closes when pool is closed
func (p *Pool) closedChan() <-chan struct{} {
	closed := make(chan struct{})
	go func() {
		p.mu.RLock()
		defer p.mu.RUnlock()
		for !p.closed {
			p.mu.RUnlock()
			time.Sleep(100 * time.Millisecond)
			p.mu.RLock()
		}
		close(closed)
	}()
	return closed
}

// Close closes the pool and all connections
func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true

	// Close idle channel
	close(p.idleChan)

	// Disconnect all connections
	for _, conn := range p.conns {
		conn.Disconnect()
	}

	p.conns = make([]Connection, 0)
	p.activeMap = make(map[string]Connection)

	p.mu.Unlock()

	// Wait for health check loop
	p.wg.Wait()

	p.logger.Info("Connection pool closed")
	return nil
}

// Size returns the total number of connections
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.conns)
}

// Capacity returns the pool capacity
func (p *Pool) Capacity() int {
	return p.config.MaxIdleConnections + p.config.MaxActiveConnections
}
