package coordination

import (
	"sync"
	"time"

	"neam-platform/shared"
)

// DeadlockDetector implements deadlock detection and prevention mechanisms
type DeadlockDetector struct {
	lockGraph    *LockGraph
	provider     LockProvider
	logger       *shared.Logger
	config       *DeadlockConfig
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	detectedDeadlocks chan<- *DeadlockInfo
}

// DeadlockConfig holds configuration for deadlock detection
type DeadlockConfig struct {
	// DetectionInterval is how often to check for deadlocks
	DetectionInterval time.Duration

	// WaitDieEnabled enables wait-die prevention strategy
	WaitDieEnabled bool

	// WoundWaitEnabled enables wound-wait prevention strategy
	WoundWaitEnabled bool

	// TimeoutEnabled enables timeout-based deadlock prevention
	TimeoutEnabled bool

	// DefaultTimeout is the default lock timeout for deadlock prevention
	DefaultTimeout time.Duration

	// MaxWaitTime is the maximum time a transaction can wait for a lock
	MaxWaitTime time.Duration

	// EnableFencingToken enables fencing token for zombie prevention
	EnableFencingToken bool
}

// DefaultDeadlockConfig returns the default configuration
func DefaultDeadlockConfig() *DeadlockConfig {
	return &DeadlockConfig{
		DetectionInterval:  5 * time.Second,
		WaitDieEnabled:     true,
		WoundWaitEnabled:   false,
		TimeoutEnabled:     true,
		DefaultTimeout:     30 * time.Second,
		MaxWaitTime:        10 * time.Second,
		EnableFencingToken: true,
	}
}

// DeadlockInfo contains information about a detected deadlock
type DeadlockInfo struct {
	ID              string              `json:"id"`
	DetectedAt      time.Time           `json:"detected_at"`
	Transactions    []*TransactionInfo  `json:"transactions"`
	ResourceWaiters map[string][]string `json:"resource_waiters"`
	Resolution      DeadlockResolution  `json:"resolution"`
}

// TransactionInfo contains information about a transaction in a deadlock
type TransactionInfo struct {
	ID            string          `json:"id"`
	ResourcesHeld []string        `json:"resources_held"`
	ResourcesWait []string        `json:"resources_wait"`
	StartTime     time.Time       `json:"start_time"`
	WaitDuration  time.Duration   `json:"wait_duration"`
	Priority      int             `json:"priority"`
}

// DeadlockResolution describes how a deadlock was resolved
type DeadlockResolution string

const (
	ResolutionAborted    DeadlockResolution = "aborted"
	ResolutionRolledBack DeadlockResolution = "rolled_back"
	ResolutionTimeout    DeadlockResolution = "timeout"
	ResolutionRecovered  DeadlockResolution = "recovered"
)

// LockGraph represents the wait-for graph of locks
type LockGraph struct {
	mu             sync.RWMutex
	nodes          map[string]*LockNode
	edges          map[string]map[string]bool // from -> to
	lastCleanup    time.Time
	cleanupInterval time.Duration
}

// LockNode represents a node in the lock graph
type LockNode struct {
	TransactionID string
	ResourceKey   string
	HeldResources map[string]bool
	WaitResource  string
	WaitStart     time.Time
	Priority      int
}

// NewDeadlockDetector creates a new deadlock detector
func NewDeadlockDetector(
	provider LockProvider,
	config *DeadlockConfig,
	logger *shared.Logger,
	detectedDeadlocks chan<- *DeadlockInfo,
) *DeadlockDetector {
	if config == nil {
		config = DefaultDeadlockConfig()
	}

	return &DeadlockDetector{
		lockGraph:         NewLockGraph(),
		provider:          provider,
		logger:            logger,
		config:            config,
		stopCh:            make(chan struct{}),
		detectedDeadlocks: detectedDeadlocks,
	}
}

// NewLockGraph creates a new lock graph
func NewLockGraph() *LockGraph {
	return &LockGraph{
		nodes:          make(map[string]*LockNode),
		edges:          make(map[string]map[string]bool),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:    time.Now(),
	}
}

// Start begins the deadlock detection process
func (dd *DeadlockDetector) Start() {
	dd.wg.Add(1)
	go dd.detectionLoop()
	dd.logger.Info("Deadlock detector started",
		"detection_interval", dd.config.DetectionInterval,
		"wait_die_enabled", dd.config.WaitDieEnabled,
		"timeout_enabled", dd.config.TimeoutEnabled,
	)
}

// Stop stops the deadlock detection process
func (dd *DeadlockDetector) Stop() {
	close(dd.stopCh)
	dd.wg.Wait()
	dd.logger.Info("Deadlock detector stopped")
}

// detectionLoop runs the deadlock detection loop
func (dd *DeadlockDetector) detectionLoop() {
	defer dd.wg.Done()

	ticker := time.NewTicker(dd.config.DetectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dd.stopCh:
			return
		case <-ticker.C:
			dd.detectAndResolve()
		}
	}
}

// detectAndResolve detects and resolves deadlocks
func (dd *DeadlockDetector) detectAndResolve() {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	// Clean up old entries
	dd.lockGraph.Cleanup()

	// Detect cycles in the wait-for graph
	cycles := dd.lockGraph.DetectCycles()
	if len(cycles) == 0 {
		return
	}

	// Handle detected deadlocks
	for _, cycle := range cycles {
		deadlockInfo := dd.buildDeadlockInfo(cycle)

		// Report deadlock
		if dd.detectedDeadlocks != nil {
			select {
			case dd.detectedDeadlocks <- deadlockInfo:
			default:
				// Channel full, skip
			}
		}

		// Resolve deadlock using wait-die or timeout
		resolution := dd.resolveDeadlock(cycle)

		deadlockInfo.Resolution = resolution
		dd.logger.Warn("Deadlock detected and resolved",
			"deadlock_id", deadlockInfo.ID,
			"transactions_involved", len(cycle),
			"resolution", resolution,
		)
	}
}

// buildDeadlockInfo builds deadlock information from a cycle
func (dd *DeadlockDetector) buildDeadlockInfo(cycle []string) *DeadlockInfo {
	info := &DeadlockInfo{
		ID:              generateDeadlockID(),
		DetectedAt:      time.Now(),
		Transactions:    make([]*TransactionInfo, 0, len(cycle)),
		ResourceWaiters: make(map[string][]string),
	}

	for _, txID := range cycle {
		node := dd.lockGraph.GetNode(txID)
		if node == nil {
			continue
		}

		txInfo := &TransactionInfo{
			ID:            txID,
			ResourcesHeld: make([]string, 0),
			ResourcesWait: make([]string, 0),
			StartTime:     node.WaitStart,
			WaitDuration:  time.Since(node.WaitStart),
			Priority:      node.Priority,
		}

		// Collect held resources
		for resource := range node.HeldResources {
			txInfo.ResourcesHeld = append(txInfo.ResourcesHeld, resource)
		}

		// Collect wait resources
		if node.WaitResource != "" {
			txInfo.ResourcesWait = append(txInfo.ResourcesWait, node.WaitResource)
			info.ResourceWaiters[node.WaitResource] = append(info.ResourceWaiters[node.WaitResource], txID)
		}

		info.Transactions = append(info.Transactions, txInfo)
	}

	return info
}

// resolveDeadlock resolves a deadlock using the configured strategy
func (dd *DeadlockDetector) resolveDeadlock(cycle []string) DeadlockResolution {
	if dd.config.WaitDieEnabled {
		return dd.resolveWaitDie(cycle)
	}

	if dd.config.TimeoutEnabled {
		return dd.resolveTimeout(cycle)
	}

	// Default: abort the youngest transaction
	return dd.resolveByPriority(cycle, false)
}

// resolveWaitDie uses the wait-die strategy to resolve deadlock
// Older transactions wait, younger transactions are aborted
func (dd *DeadlockDetector) resolveWaitDie(cycle []string) DeadlockResolution {
	if len(cycle) < 2 {
		return ResolutionRecovered
	}

	// Find the youngest transaction (highest priority number = newer)
	var youngestTx string
	var maxPriority int

	for _, txID := range cycle {
		node := dd.lockGraph.GetNode(txID)
		if node != nil && node.Priority > maxPriority {
			maxPriority = node.Priority
			youngestTx = txID
		}
	}

	if youngestTx == "" {
		return ResolutionTimeout
	}

	// Abort the youngest transaction
	dd.abortTransaction(youngestTx)

	return ResolutionAborted
}

// resolveByPriority aborts transactions based on priority
// killOlder: if true, kill older transactions; if false, kill younger
func (dd *DeadlockDetector) resolveByPriority(cycle []string, killOlder bool) DeadlockResolution {
	if len(cycle) < 2 {
		return ResolutionRecovered
	}

	// Find the transaction to abort
	var targetTx string
	var targetPriority int

	for _, txID := range cycle {
		node := dd.lockGraph.GetNode(txID)
		if node == nil {
			continue
		}

		isOlder := node.Priority < targetPriority
		isYounger := node.Priority > targetPriority

		if targetTx == "" {
			targetTx = txID
			targetPriority = node.Priority
		} else if killOlder && isOlder {
			targetTx = txID
			targetPriority = node.Priority
		} else if !killOlder && isYounger {
			targetTx = txID
			targetPriority = node.Priority
		}
	}

	if targetTx == "" {
		return ResolutionTimeout
	}

	dd.abortTransaction(targetTx)

	return ResolutionAborted
}

// resolveTimeout uses timeout to resolve deadlock
func (dd *DeadlockDetector) resolveTimeout(cycle []string) DeadlockResolution {
	now := time.Now()
	var oldestTx string
	var oldestStart time.Time

	for _, txID := range cycle {
		node := dd.lockGraph.GetNode(txID)
		if node != nil {
			if oldestTx == "" || node.WaitStart.Before(oldestStart) {
				oldestTx = txID
				oldestStart = node.WaitStart
			}
		}
	}

	if oldestTx == "" {
		return ResolutionTimeout
	}

	// Check if the oldest transaction has exceeded max wait time
	if now.Sub(oldestStart) > dd.config.MaxWaitTime {
		dd.abortTransaction(oldestTx)
		return ResolutionTimeout
	}

	// No transaction exceeded timeout yet
	return ResolutionRecovered
}

// abortTransaction aborts a transaction
func (dd *DeadlockDetector) abortTransaction(txID string) {
	dd.logger.Info("Aborting transaction due to deadlock",
		"transaction_id", txID,
	)

	// Remove from lock graph
	dd.lockGraph.RemoveNode(txID)
}

// RegisterWait registers that a transaction is waiting for a lock
func (dd *DeadlockDetector) RegisterWait(txID, resourceKey string, priority int) {
	dd.lockGraph.AddWaiter(txID, resourceKey, priority)
}

// RegisterHold registers that a transaction holds a lock
func (dd *DeadlockDetector) RegisterHold(txID, resourceKey string) {
	dd.lockGraph.AddHolder(txID, resourceKey)
}

// RegisterRelease registers that a transaction released a lock
func (dd *DeadlockDetector) RegisterRelease(txID, resourceKey string) {
	dd.lockGraph.RemoveHolder(txID, resourceKey)
}

// GetWaitForGraph returns the current wait-for graph
func (dd *DeadlockDetector) GetWaitForGraph() *LockGraph {
	return dd.lockGraph
}

// AddWaiter adds a transaction waiting for a resource
func (lg *LockGraph) AddWaiter(txID, resourceKey string, priority int) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	node := lg.getOrCreateNode(txID)
	node.WaitResource = resourceKey
	node.WaitStart = time.Now()
	node.Priority = priority

	// Find transactions holding this resource and add edge
	for otherID, otherNode := range lg.nodes {
		if otherID != txID && otherNode.HeldResources[resourceKey] {
			lg.addEdge(txID, otherID)
		}
	}
}

// AddHolder adds a transaction holding a resource
func (lg *LockGraph) AddHolder(txID, resourceKey string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	node := lg.getOrCreateNode(txID)
	node.HeldResources[resourceKey] = true

	// Check if any transactions are waiting for this resource
	for otherID, otherNode := range lg.nodes {
		if otherID != txID && otherNode.WaitResource == resourceKey {
			lg.addEdge(otherID, txID)
		}
	}
}

// RemoveHolder removes a transaction's hold on a resource
func (lg *LockGraph) RemoveHolder(txID, resourceKey string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	node := lg.nodes[txID]
	if node != nil {
		delete(node.HeldResources, resourceKey)

		// Remove edges from waiters to this transaction
		for otherID := range lg.edges {
			if otherID != txID {
				delete(lg.edges[otherID], txID)
			}
		}
	}
}

// AddEdge adds an edge from one transaction to another
func (lg *LockGraph) addEdge(from, to string) {
	if lg.edges[from] == nil {
		lg.edges[from] = make(map[string]bool)
	}
	lg.edges[from][to] = true
}

// GetNode returns a node by transaction ID
func (lg *LockGraph) GetNode(txID string) *LockNode {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	return lg.nodes[txID]
}

// RemoveNode removes a node from the graph
func (lg *LockGraph) RemoveNode(txID string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	delete(lg.nodes, txID)
	delete(lg.edges, txID)

	// Remove edges pointing to this node
	for from := range lg.edges {
		delete(lg.edges[from], txID)
	}
}

// Cleanup removes stale entries from the graph
func (lg *LockGraph) Cleanup() {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if time.Since(lg.lastCleanup) < lg.cleanupInterval {
		return
	}

	lg.lastCleanup = time.Now()
	staleThreshold := time.Now().Add(-10 * time.Minute)

	// Remove nodes that have been waiting too long (stale)
	for id, node := range lg.nodes {
		if node.WaitResource != "" && node.WaitStart.Before(staleThreshold) {
			delete(lg.nodes, id)
			delete(lg.edges, id)
		}
	}

	// Clean up orphaned edges
	for from := range lg.edges {
		for to := range lg.edges[from] {
			if _, exists := lg.nodes[to]; !exists {
				delete(lg.edges[from], to)
			}
		}
	}
}

// DetectCycles detects cycles in the wait-for graph using DFS
func (lg *LockGraph) DetectCycles() [][]string {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		// Check neighbors
		for neighbor := range lg.edges[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if recStack[neighbor] {
				// Found a cycle
				cycle := make([]string, 0)
				for i := len(path) - 1; i >= 0; i-- {
					cycle = append(cycle, path[i])
					if path[i] == neighbor {
						break
					}
				}
				cycles = append(cycles, cycle)
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	// Start DFS from each unvisited node
	for node := range lg.nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// getOrCreateNode gets or creates a node
func (lg *LockGraph) getOrCreateNode(txID string) *LockNode {
	if node, exists := lg.nodes[txID]; exists {
		return node
	}

	node := &LockNode{
		TransactionID: txID,
		HeldResources: make(map[string]bool),
	}
	lg.nodes[txID] = node
	return node
}

// generateDeadlockID generates a unique deadlock ID
func generateDeadlockID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// WaitDiePolicy implements the wait-die deadlock prevention strategy
type WaitDiePolicy struct {
	mu           sync.RWMutex
	transactions map[string]*TransactionState
	logger       *shared.Logger
}

// TransactionState holds the state of a transaction
type TransactionState struct {
	ID        string
	Timestamp time.Time
	Priority  int // Higher = newer
	Resources map[string]bool
}

// NewWaitDiePolicy creates a new wait-die policy
func NewWaitDiePolicy(logger *shared.Logger) *WaitDiePolicy {
	return &WaitDiePolicy{
		transactions: make(map[string]*TransactionState),
		logger:       logger,
	}
}

// CanWait determines if a transaction can wait for a resource
// Returns true if the transaction should wait, false if it should die (abort)
func (wp *WaitDiePolicy) CanWait(requestingTx, holdingTx string) bool {
wp.mu.RLock()
	requestingState := wp.transactions[requestingTx]
	holdingState := wp.transactions[holdingTx]
wp.mu.RUnlock()

	if requestingState == nil || holdingState == nil {
		// Default: allow wait
		return true
	}

	// Wait-die: older transactions wait, newer transactions die
	// "Older" means lower priority (earlier timestamp)
	if requestingState.Timestamp.After(holdingState.Timestamp) {
		// Requesting transaction is newer, it should die
		wp.logger.Debug("Wait-die: younger transaction dies",
			"requesting_tx", requestingTx,
			"holding_tx", holdingTx,
		)
		return false
	}

	// Requesting transaction is older, it can wait
	wp.logger.Debug("Wait-die: older transaction waits",
			"requesting_tx", requestingTx,
			"holding_tx", holdingTx,
	)
	return true
}

// RegisterTransaction registers a new transaction
func (wp *WaitDiePolicy) RegisterTransaction(txID string) {
wp.mu.Lock()
	defer wp.mu.Unlock()

wp.transactions[txID] = &TransactionState{
		ID:        txID,
		Timestamp: time.Now(),
		Priority:  int(time.Now().UnixNano()),
		Resources: make(map[string]bool),
	}
}

// UnregisterTransaction unregisters a transaction
func (wp *WaitDiePolicy) UnregisterTransaction(txID string) {
wp.mu.Lock()
	defer wp.mu.Unlock()

	delete(wp.transactions, txID)
}

// AddResource records that a transaction holds a resource
func (wp *WaitDiePolicy) AddResource(txID, resource string) {
wp.mu.Lock()
	defer wp.mu.Unlock()

	if state, exists := wp.transactions[txID]; exists {
		state.Resources[resource] = true
	}
}

// RemoveResource records that a transaction released a resource
func (wp *WaitDiePolicy) RemoveResource(txID, resource string) {
wp.mu.Lock()
	defer wp.mu.Unlock()

	if state, exists := wp.transactions[txID]; exists {
		delete(state.Resources, resource)
	}
}

// FencingTokenHandler handles fencing tokens for zombie prevention
type FencingTokenHandler struct {
	mu            sync.RWMutex
	tokenStore    map[string]int64 // resource -> latest valid token
	currentTokens map[string]int64 // resource -> current token
}

// NewFencingTokenHandler creates a new fencing token handler
func NewFencingTokenHandler() *FencingTokenHandler {
	return &FencingTokenHandler{
		tokenStore:    make(map[string]int64),
		currentTokens: make(map[string]int64),
	}
}

// ValidateToken validates that a fencing token is still valid
// Returns true if the token is valid, false otherwise
func (fth *FencingTokenHandler) ValidateToken(resource string, token int64) bool {
	fth.mu.RLock()
	defer fth.mu.RUnlock()

	latestValid, exists := fth.tokenStore[resource]
	if !exists {
		// No token recorded, accept all tokens
		return true
	}

	return token >= latestValid
}

// UpdateToken updates the latest valid token for a resource
func (fth *FencingTokenHandler) UpdateToken(resource string, token int64) {
	fth.mu.Lock()
	defer fth.mu.Unlock()

	if current, exists := fth.tokenStore[resource]; !exists || token > current {
		fth.tokenStore[resource] = token
	}
}

// GetCurrentToken returns the current fencing token for a resource
func (fth *FencingTokenHandler) GetCurrentToken(resource string) int64 {
	fth.mu.RLock()
	defer fth.mu.RUnlock()

	return fth.currentTokens[resource]
}

// IncrementToken increments and returns the current fencing token
func (fth *FencingTokenHandler) IncrementToken(resource string) int64 {
	fth.mu.Lock()
	defer fth.mu.Unlock()

	fth.currentTokens[resource]++
	return fth.currentTokens[resource]
}
