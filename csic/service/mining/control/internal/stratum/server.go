// Stratum Server - Mining Pool Protocol Implementation
// Implements Stratum V1 protocol for mining pool operations

package stratum

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StratumServer manages Stratum protocol connections
type StratumServer struct {
	listener     net.Listener
	clients      map[string]*StratumClient
	clientMu     sync.RWMutex
	config       *StratumConfig
	jobManager   *JobManager
	running      bool
	wg           sync.WaitGroup
}

// StratumConfig contains Stratum server configuration
type StratumConfig struct {
	Port                     int    `yaml:"port"`
	MaxConnections           int    `yaml:"max_connections"`
	DifficultyInitial        uint64 `yaml:"difficulty_initial"`
	DifficultyAdjustmentSecs int    `yaml:"difficulty_adjustment_interval"`
	JobInterval              int    `yaml:"job_interval"`
	MaxJobDelay              int    `yaml:"max_job_delay"`
	RegisterTimeout          int    `yaml:"register_timeout"` // seconds
}

// StratumClient represents a connected mining client
type StratumClient struct {
	ID           string
	conn         net.Conn
	reader       *bufio.Reader
	writer       *json.Encoder
	subscribed   bool
	authorized   bool
	workerName   string
	workerPass   string
	difficulty   uint64
	lastShare    time.Time
	shares       uint64
	staleShares  uint64
	acceptedShares uint64
	rejectedShares uint64
	pendingJobs  map[string]*StratumJob
	currentJob   *StratumJob
	extranonce1  string
	extranonce2Size int
	mu           sync.Mutex
	connectedAt  time.Time
}

// StratumMessage represents a Stratum protocol message
type StratumMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Error  interface{}     `json:"error,omitempty"`
	Result interface{}     `json:"result,omitempty"`
}

// StratumJob represents a mining job
type StratumJob struct {
	JobID        string
	Hash1        string
	CoinBase1    string
	CoinBase2    string
	MerkleBranch []string
	Version      string
	NBits        string
	NPrevHash    string
	CleanJobs    bool
	CreatedAt    time.Time
	Difficulty   uint64
}

// SubmitShare represents a submitted share
type SubmitShare struct {
	WorkerName string
	JobID      string
	Extranonce2 string
	NTime      string
	Nonce      string
}

// NewStratumServer creates a new Stratum server
func NewStratumServer(config *StratumConfig, jobManager *JobManager) *StratumServer {
	return &StratumServer{
		clients:    make(map[string]*StratumClient),
		config:     config,
		jobManager: jobManager,
	}
}

// Start begins listening for Stratum connections
func (s *StratumServer) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start Stratum listener: %w", err)
	}

	s.listener = listener
	s.running = true

	log.Printf("Stratum server listening on port %d", s.config.Port)

	// Start accepting connections
	s.wg.Add(1)
	go s.acceptLoop()

	// Start difficulty adjustment routine
	s.wg.Add(1)
	go s.difficultyAdjustmentLoop()

	return nil
}

// Stop gracefully stops the Stratum server
func (s *StratumServer) Stop() {
	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all client connections
	s.clientMu.Lock()
	for _, client := range s.clients {
		client.conn.Close()
	}
	s.clientMu.Unlock()

	s.wg.Wait()
	log.Println("Stratum server stopped")
}

// GetStats returns server statistics
func (s *StratumServer) GetStats() StratumStats {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	var connectedClients int
	var authorizedClients int
	for _, client := range s.clients {
		if client.authorized {
			authorizedClients++
		}
		connectedClients++
	}

	return StratumStats{
		ConnectedClients: connectedClients,
		AuthorizedClients: authorizedClients,
		TotalShares:       s.getTotalShares(),
		TotalAccepted:     s.getTotalAccepted(),
		TotalRejected:     s.getTotalRejected(),
	}
}

// StratumStats contains server statistics
type StratumStats struct {
	ConnectedClients int
	AuthorizedClients int
	TotalShares       uint64
	TotalAccepted     uint64
	TotalRejected     uint64
}

func (s *StratumServer) acceptLoop() {
	defer s.wg.Done()

	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				log.Printf("Error accepting connection: %v", err)
			}
			continue
		}

		// Check max connections
		s.clientMu.Lock()
		if len(s.clients) >= s.config.MaxConnections {
			s.clientMu.Unlock()
			conn.Close()
			continue
		}
		s.clientMu.Unlock()

		client := s.newClient(conn)
		s.wg.Add(1)
		go s.handleClient(client)
	}
}

func (s *StratumServer) newClient(conn net.Conn) *StratumClient {
	return &StratumClient{
		ID:           uuid.New().String()[:8],
		conn:         conn,
		reader:       bufio.NewReader(conn),
		writer:       json.NewEncoder(conn),
		subscribed:   false,
		authorized:   false,
		pendingJobs:  make(map[string]*StratumJob),
		connectedAt:  time.Now(),
		difficulty:   s.config.DifficultyInitial,
		extranonce2Size: 4,
	}
}

func (s *StratumServer) handleClient(client *StratumClient) {
	defer s.wg.Done()
	defer s.removeClient(client.ID)
	defer client.conn.Close()

	// Set read timeout
	client.conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.RegisterTimeout) * time.Second))

	// Add to client map
	s.clientMu.Lock()
	s.clients[client.ID] = client
	s.clientMu.Unlock()

	reader := client.reader

	for s.running {
		// Read message
		var msg StratumMessage
		if err := reader.Decode(&msg); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Check if client timed out before subscribing
				if !client.subscribed {
					log.Printf("Client %s timed out before subscribing", client.ID)
					return
				}
				continue
			}
			log.Printf("Error reading from client %s: %v", client.ID, err)
			return
		}

		// Reset deadline
		client.conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// Handle message
		if err := s.handleMessage(client, &msg); err != nil {
			log.Printf("Error handling message from %s: %v", client.ID, err)
			s.sendError(client, msg.ID, err.Error())
		}
	}
}

func (s *StratumServer) handleMessage(client *StratumClient, msg *StratumMessage) error {
	switch msg.Method {
	case "mining.subscribe":
		return s.handleSubscribe(client, msg)
	case "mining.authorize":
		return s.handleAuthorize(client, msg)
	case "mining.submit":
		return s.handleSubmit(client, msg)
	case "mining.extranonce.subscribe":
		return s.handleExtranonceSubscribe(client, msg)
	case "mining.get_transactions":
		return s.handleGetTransactions(client, msg)
	default:
		return fmt.Errorf("unknown method: %s", msg.Method)
	}
}

// mining.subscribe - Client requests to subscribe to mining notifications
func (s *StratumServer) handleSubscribe(client *StratumClient, msg *StratumMessage) error {
	// Parse optional parameters
	var params []string
	if msg.Params != nil {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return err
		}
	}

	// Generate extranonce1 for this client
	extranonce1 := generateExtranonce1()
	client.extranonce1 = extranonce1

	// Send subscription response
	result := []interface{}{
		[[2]string{{"mining.set_difficulty", ""}, {"mining.notify", ""}},
		extranonce1,
		int(s.config.DifficultyAdjustmentSecs),
	}
	}

	return s.sendResponse(client, msg.ID, result)
}

// mining.authorize - Client authenticates with the pool
func (s *StratumServer) handleAuthorize(client *StratumClient, msg *StratumMessage) error {
	var params []string
	if msg.Params != nil {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return err
		}
	}

	if len(params) < 2 {
		return errors.New("invalid authorization params")
	}

	client.workerName = params[0]
	client.workerPass = params[1]
	client.authorized = true

	log.Printf("Client %s authorized as %s", client.ID, client.workerName)

	// Send success response
	return s.sendResponse(client, msg.ID, true)
}

// mining.submit - Client submits a share
func (s *StratumServer) handleSubmit(client *StratumClient, msg *StratumMessage) error {
	if !client.authorized {
		return errors.New("unauthorized")
	}

	var params []string
	if msg.Params != nil {
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return err
		}
	}

	if len(params) < 5 {
		return errors.New("invalid submit params")
	}

	workerName := params[0]
	jobID := params[1]
	extranonce2 := params[2]
	nTime := params[3]
	nonce := params[4]

	// Validate share
	share := SubmitShare{
		WorkerName:  workerName,
		JobID:       jobID,
		Extranonce2: extranonce2,
		NTime:       nTime,
		Nonce:       nonce,
	}

	valid, reason := s.validateShare(client, &share)
	if !valid {
		client.mu.Lock()
		client.rejectedShares++
		client.mu.Unlock()
		return errors.New(reason)
	}

	// Accept share
	client.mu.Lock()
	client.shares++
	client.acceptedShares++
	client.lastShare = time.Now()
	client.mu.Unlock()

	// Notify job manager
	s.jobManager.SubmitShare(client.workerName, &share)

	log.Printf("Share accepted from %s@%s (difficulty: %d)", client.workerName, client.ID, client.difficulty)

	return s.sendResponse(client, msg.ID, true)
}

func (s *StratumServer) handleExtranonceSubscribe(client *StratumClient, msg *StratumMessage) error {
	return s.sendResponse(client, msg.ID, client.extranonce1)
}

func (s *StratumServer) handleGetTransactions(client *StratumClient, msg *StratumMessage) error {
	return s.sendResponse(client, msg.ID, []string{})
}

func (s *StratumServer) validateShare(client *StratumClient, share *SubmitShare) (bool, string) {
	// Check if job exists
	client.mu.Lock()
	job, exists := client.pendingJobs[share.JobID]
	if !exists {
		job = client.currentJob
	}
	client.mu.Unlock()

	if job == nil {
		return false, "Job not found"
	}

	// Check timestamp
	nTime, err := parseUint32(share.NTime)
	if err != nil {
		return false, "Invalid nTime"
	}

	jobTime := uint32(job.CreatedAt.Unix())
	if nTime < jobTime {
		return false, "Stale share"
	}

	// Check if nTime is too far in the future
	maxTime := uint32(time.Now().Add(2 * time.Hour).Unix())
	if nTime > maxTime {
		return false, "nTime too far in future"
	}

	// Validate nonce format (must be hex)
	if _, err := hex.DecodeString(share.Nonce); err != nil {
		return false, "Invalid nonce format"
	}

	return true, ""
}

// BroadcastJob sends a new mining job to all authorized clients
func (s *StratumServer) BroadcastJob(job *StratumJob) {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	for _, client := range s.clients {
		if !client.authorized {
			continue
		}

		client.mu.Lock()
		client.pendingJobs[job.JobID] = job
		client.currentJob = job
		client.mu.Unlock()

		go s.sendJob(client, job)
	}
}

func (s *StratumServer) sendJob(client *StratumClient, job *StratumJob) {
	// Build mining.notify params
	jobParams := []interface{}{
		job.JobID,
		job.Hash1,
		job.CoinBase1,
		job.CoinBase2,
		job.MerkleBranch,
		job.Version,
		job.NBits,
		job.NPrevHash,
		job.CleanJobs,
	}

	// Send notification
	notification := StratumMessage{
		Method: "mining.notify",
		Params: mustMarshal(jobParams),
	}

	if err := client.writer.Encode(&notification); err != nil {
		log.Printf("Error sending job to %s: %v", client.ID, err)
	}
}

// SetDifficulty adjusts the difficulty for a specific client
func (s *StratumServer) SetDifficulty(clientID string, difficulty uint64) error {
	s.clientMu.RLock()
	client, exists := s.clients[clientID]
	s.clientMu.RUnlock()

	if !exists {
		return errors.New("client not found")
	}

	client.mu.Lock()
	client.difficulty = difficulty
	client.mu.Unlock()

	// Send difficulty notification
	notification := StratumMessage{
		Method: "mining.set_difficulty",
		Params: mustMarshal([]interface{}{difficulty}),
	}

	if err := client.writer.Encode(&notification); err != nil {
		return err
	}

	log.Printf("Set difficulty %d for client %s", difficulty, clientID)
	return nil
}

func (s *StratumServer) difficultyAdjustmentLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(s.config.DifficultyAdjustmentSecs) * time.Second)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ticker.C:
			s.adjustDifficulties()
		}
	}
}

func (s *StratumServer) adjustDifficulties() {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	for _, client := range s.clients {
		if !client.authorized {
			continue
		}

		client.mu.Lock()

		// Calculate hashrate from shares
		timeSinceLastShare := time.Since(client.lastShare)
		var estimatedHashrate float64

		if !client.lastShare.IsZero() && client.shares > 0 {
			// Simple estimate: shares * difficulty / time
			estimatedHashrate = float64(client.shares) * float64(client.difficulty) / timeSinceLastShare.Seconds()
		}

		client.mu.Unlock()

		// Adjust difficulty based on estimated hashrate
		var newDifficulty uint64
		if estimatedHashrate > 1000 { // Very high hashrate
			newDifficulty = client.difficulty * 2
		} else if estimatedHashrate < 10 && client.difficulty > s.config.DifficultyInitial/2 {
			newDifficulty = client.difficulty / 2
		}

		if newDifficulty > 0 && newDifficulty != client.difficulty {
			if newDifficulty < s.config.DifficultyInitial {
				newDifficulty = s.config.DifficultyInitial
			}
			if newDifficulty > 1000000 {
				newDifficulty = 1000000
			}

			client.difficulty = newDifficulty
			go s.SetDifficulty(client.ID, newDifficulty)
		}
	}
}

func (s *StratumServer) sendResponse(client *StratumClient, id json.RawMessage, result interface{}) error {
	return client.writer.Encode(&StratumMessage{
		ID:     id,
		Result: result,
	})
}

func (s *StratumServer) sendError(client *StratumClient, id json.RawMessage, errMsg string) {
	client.writer.Encode(&StratumMessage{
		ID:    id,
		Error: []interface{}{-1, errMsg, nil},
	})
}

func (s *StratumServer) removeClient(id string) {
	s.clientMu.Lock()
	delete(s.clients, id)
	s.clientMu.Unlock()
}

func (s *StratumServer) getTotalShares() uint64 {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	var total uint64
	for _, client := range s.clients {
		total += client.shares
	}
	return total
}

func (s *StratumServer) getTotalAccepted() uint64 {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	var total uint64
	for _, client := range s.clients {
		total += client.acceptedShares
	}
	return total
}

func (s *StratumServer) getTotalRejected() uint64 {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	var total uint64
	for _, client := range s.clients {
		total += client.rejectedShares
	}
	return total
}

// Helper functions

func generateExtranonce1() string {
	randBytes := make([]byte, 4)
	// In production, use crypto/rand
	hash := sha256.Sum256(randBytes)
	return hex.EncodeToString(hash[:4])
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func parseUint32(s string) (uint32, error) {
	var n uint64
	_, err := fmt.Sscanf(s, "%x", &n)
	return uint32(n), err
}
