// Job Manager - Block Template Management
// Fetches block templates from blockchain node and creates mining jobs

package stratum

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/csic/mining-control/internal/domain"
	"github.com/google/uuid"
)

// JobManager manages mining jobs and block templates
type JobManager struct {
	rpcClient   BlockchainRPCClient
	pendingJobs map[string]*StratumJob
	currentJob  *StratumJob
	config      *JobManagerConfig
	mu          sync.RWMutex
	wg          sync.WaitGroup
	running     bool
}

// JobManagerConfig contains job manager configuration
type JobManagerConfig struct {
	JobInterval     int    // seconds
	MaxJobsAge      int    // seconds before job is stale
	RPCURL          string // blockchain node RPC URL
	RPCUser         string
	RPCPassword     string
	CoinbasePayout  string // payout address
}

// BlockchainRPCClient interface for blockchain node communication
type BlockchainRPCClient interface {
	GetBlockTemplate() (*BlockTemplate, error)
	SubmitBlock(block []byte) error
	GetNetworkInfo() (*NetworkInfo, error)
}

// BlockTemplate represents a blockchain block template
type BlockTemplate struct {
	Version       string   `json:"version"`
	PreviousHash  string   `json:"previoushash"`
	CoinBaseAux   map[string]string `json:"coinbaseaux"`
	CoinBaseValue int64    `json:"coinbasevalue"`
	Target        string   `json:"target"`
	CurTime       int64    `json:"curtime"`
	Bits          string   `json:"bits"`
	Height        int64    `json:"height"`
	MerkleBranch  []string `json:"merkleroot"`
	Transactions  []Transaction `json:"transactions"`
	TransactionFees []int64 `json:"transactionfees"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	Data     string `json:"data"`
	Hash     string `json:"hash"`
	Fee      int64  `json:"fee"`
	Required bool   `json:"required"`
}

// NetworkInfo contains blockchain network information
type NetworkInfo struct {
	Version         int64   `json:"version"`
	ProtocolVersion int64   `json:"protocolversion"`
	Blocks          int64   `json:"blocks"`
	Connections     int64   `json:"connections"`
	Difficulty      float64 `json:"difficulty"`
	NetworkHashRate float64 `json:"networkhashps"`
}

// NewJobManager creates a new job manager
func NewJobManager(client BlockchainRPCClient, config *JobManagerConfig) *JobManager {
	return &JobManager{
		rpcClient:   client,
		pendingJobs: make(map[string]*StratumJob),
		config:      config,
	}
}

// Start begins the job manager
func (jm *JobManager) Start(ctx context.Context) error {
	jm.running = true

	// Fetch initial block template
	if err := jm.fetchAndCreateJob(); err != nil {
		log.Printf("Warning: Failed to fetch initial block template: %v", err)
	}

	// Start job refresh loop
	jm.wg.Add(1)
	go jm.jobRefreshLoop(ctx)

	return nil
}

// Stop gracefully stops the job manager
func (jm *JobManager) Stop() {
	jm.running = false
	jm.wg.Wait()
}

// GetCurrentJob returns the current mining job
func (jm *JobManager) GetCurrentJob() *StratumJob {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.currentJob
}

// SubmitShare processes a submitted share
func (jm *JobManager) SubmitShare(worker string, share *SubmitShare) {
	// In production, this would:
	// 1. Calculate the share difficulty
	// 2. Credit the worker
	// 3. Check if the share meets the network target
	// 4. If so, submit a full block to the network

	log.Printf("Share submitted by %s: job=%s nonce=%s", worker, share.JobID, share.Nonce)

	// Calculate share hash and check if it meets difficulty
	// This is a simplified version - real implementation would:
	// 1. Reconstruct the coinbase transaction
	// 2. Calculate the merkle root
	// 3. Build the block header
	// 4. Hash and check against target
}

// SubmitBlock submits a solved block to the network
func (jm *JobManager) SubmitBlock(blockData []byte) error {
	return jm.rpcClient.SubmitBlock(blockData)
}

// GetStats returns job manager statistics
func (jm *JobManager) GetStats() JobManagerStats {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	return JobManagerStats{
		CurrentHeight: jm.currentJob.Height,
		PendingJobs:   len(jm.pendingJobs),
		LastJobTime:   jm.currentJob.CreatedAt,
	}
}

// JobManagerStats contains job manager statistics
type JobManagerStats struct {
	CurrentHeight int64
	PendingJobs   int
	LastJobTime   time.Time
}

func (jm *JobManager) jobRefreshLoop(ctx context.Context) {
	defer jm.wg.Done()

	ticker := time.NewTicker(time.Duration(jm.config.JobInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := jm.fetchAndCreateJob(); err != nil {
				log.Printf("Error refreshing block template: %v", err)
			}
		}
	}
}

func (jm *JobManager) fetchAndCreateJob() error {
	// Fetch block template from blockchain node
	template, err := jm.rpcClient.GetBlockTemplate()
	if err != nil {
		return fmt.Errorf("failed to get block template: %w", err)
	}

	// Create Stratum job from template
	job := jm.createStratumJob(template)

	jm.mu.Lock()
	jm.currentJob = job
	jm.pendingJobs[job.JobID] = job
	jm.mu.Unlock()

	log.Printf("New job created: height=%d job_id=%s", template.Height, job.JobID)

	return nil
}

func (jm *JobManager) createStratumJob(template *BlockTemplate) *StratumJob {
	jobID := uuid.New().String()[:8]

	// Build coinbase transaction
	coinbase1, coinbase2 := jm.buildCoinbase(template.Height, template.CoinBaseValue)

	// Calculate merkle root
	merkleRoot := jm.calculateMerkleRoot(coinbase1+coinbase2, template.MerkleBranch)

	// Create job
	return &StratumJob{
		JobID:        jobID,
		Hash1:        "0000000000000000000000000000000000000000000000000000000000000000",
		CoinBase1:    coinbase1,
		CoinBase2:    coinbase2,
		MerkleBranch: template.MerkleBranch,
		Version:      template.Version,
		NBits:        template.Bits,
		NPrevHash:    template.PreviousHash,
		CleanJobs:    true,
		CreatedAt:    time.Now(),
		Difficulty:   1, // Will be set per-client
		Height:       template.Height,
	}
}

func (jm *JobManager) buildCoinbase(height int64, value int64) (string, string) {
	// Coinbase typically contains:
	// - Height (little-endian, varint)
	// - Extra nonce (random data)
	// - Timestamp
	// - Pool information

	heightBytes := []byte{
		byte(height),
		byte(height >> 8),
		byte(height >> 16),
	}

	coinbase1 := hex.EncodeToString(heightBytes)
	coinbase2 := jm.config.CoinbasePayout

	return coinbase1, coinbase2
}

func (jm *JobManager) calculateMerkleRoot(coinbaseTx string, merkleBranch []string) string {
	// First hash the coinbase transaction
	hash := sha256.Sum256([]byte(coinbaseTx))
	hash = sha256.Sum256(hash[:])

	// Hash with each level of the merkle tree
	for _, branch := range merkleBranch {
		combined := hex.EncodeToString(hash[:]) + branch
		hash = sha256.Sum256([]byte(combined))
		hash = sha256.Sum256(hash[:])
	}

	return hex.EncodeToString(hash[:])
}

// SimpleRPCClient implements BlockchainRPCClient for Bitcoin-like nodes
type SimpleRPCClient struct {
	url      string
	user     string
	password string
}

// NewSimpleRPCClient creates a simple RPC client
func NewSimpleRPCClient(url, user, password string) *SimpleRPCClient {
	return &SimpleRPCClient{
		url:      url,
		user:     user,
		password: password,
	}
}

// GetBlockTemplate fetches the current block template
func (c *SimpleRPCClient) GetBlockTemplate() (*BlockTemplate, error) {
	// In production, this would make an HTTP POST to the blockchain node
	// For now, return a mock template
	return &BlockTemplate{
		Version:      "20000000",
		PreviousHash: "0000000000000000000a2b5c5d6e7f8g9h0i1j2k3l4m5n6o7p8q9r0s1t2u3v4",
		CoinBaseValue: 625000000,
		CurTime:       time.Now().Unix(),
		Bits:          "170c8f6f",
		Height:        800000,
		MerkleBranch:  []string{},
		Transactions:  []Transaction{},
	}, nil
}

// SubmitBlock submits a solved block
func (c *SimpleRPCClient) SubmitBlock(block []byte) error {
	return nil
}

// GetNetworkInfo fetches network information
func (c *SimpleRPCClient) GetNetworkInfo() (*NetworkInfo, error) {
	return &NetworkInfo{
		Version:         70015,
		ProtocolVersion: 70015,
		Blocks:          800000,
		Connections:     50,
		Difficulty:      50000000000,
		NetworkHashRate: 300000000000000,
	}, nil
}

// CalculateShareDifficulty calculates if a share meets the target
func CalculateShareDifficulty(shareHash string, target string) bool {
	// Convert hex hashes to byte arrays
	shareBytes, _ := hex.DecodeString(shareHash)
	targetBytes, _ := hex.DecodeString(target)

	// Compare as big-endian integers
	for i := 0; i < len(shareBytes); i++ {
		if shareBytes[i] < targetBytes[i] {
			return true
		} else if shareBytes[i] > targetBytes[i] {
			return false
		}
	}
	return true
}

// ShareSubmission represents a share submission for tracking
type ShareSubmission struct {
	ID            string    `json:"id"`
	WorkerID      string    `json:"worker_id"`
	JobID         string    `json:"job_id"`
	Timestamp     time.Time `json:"timestamp"`
	Difficulty    uint64    `json:"difficulty"`
	IsValid       bool      `json:"is_valid"`
	IsStale       bool      `json:"is_stale"`
	RejectReason  string    `json:"reject_reason,omitempty"`
	BlockHeight   int64     `json:"block_height,omitempty"`
	RewardCredits float64   `json:"reward_credits"`
}

// PoolStats represents pool statistics
type PoolStats struct {
	TotalHashRate      float64 `json:"total_hash_rate"`
	TotalMiners        int     `json:"total_miners"`
	ActiveMiners       int     `json:"active_miners"`
	TotalShares        uint64  `json:"total_shares"`
	AcceptedShares     uint64  `json:"accepted_shares"`
	RejectedShares     uint64  `json:"rejected_shares"`
	StaleShares        uint64  `json:"stale_shares"`
	CurrentDifficulty  float64 `json:"current_difficulty"`
	CurrentHeight      int64   `json:"current_height"`
	NetworkDifficulty  float64 `json:"network_difficulty"`
	EstimatedBlockTime int     `json:"estimated_block_time_seconds"`
}

// WorkerStats represents statistics for a single worker
type WorkerStats struct {
	WorkerName       string    `json:"worker_name"`
	HashRate         float64   `json:"hash_rate"`
	SharesAccepted   uint64    `json:"shares_accepted"`
	SharesRejected   uint64    `json:"shares_rejected"`
	SharesStale      uint64    `json:"shares_stale"`
	LastShareTime    time.Time `json:"last_share_time"`
	CurrentDifficulty uint64   `json:"current_difficulty"`
}

// RegisterSubmissionHandler handles share submissions
type SubmissionHandler struct {
	mu        sync.RWMutex
	submissions map[string]*ShareSubmission
	workerStats  map[string]*WorkerStats
	poolStats    PoolStats
}

// NewSubmissionHandler creates a new submission handler
func NewSubmissionHandler() *SubmissionHandler {
	return &SubmissionHandler{
		submissions: make(map[string]*ShareSubmission),
		workerStats:  make(map[string]*WorkerStats),
	}
}

// HandleSubmission processes a share submission
func (h *SubmissionHandler) HandleSubmission(submission *domain.HashRateTelemetryPacket) ShareSubmission {
	h.mu.Lock()
	defer h.mu.Unlock()

	workerID := submission.PoolID.String()

	stats, exists := h.workerStats[workerID]
	if !exists {
		stats = &WorkerStats{
			WorkerName: workerID,
		}
		h.workerStats[workerID] = stats
	}

	// Update worker stats
	stats.SharesAccepted++
	stats.LastShareTime = time.Now()

	// Update pool stats
	h.poolStats.TotalShares++
	h.poolStats.AcceptedShares++

	return ShareSubmission{
		ID:            uuid.New().String(),
		WorkerID:      workerID,
		Timestamp:     time.Now(),
		IsValid:       true,
		RewardCredits: 1.0 / float64(submission.Difficulty),
	}
}
