package coordination

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"neam-platform/shared"
)

// TestLockGraph tests the lock graph data structure
func TestLockGraph(t *testing.T) {
	t.Run("add waiter creates node", func(t *testing.T) {
		graph := NewLockGraph()

		graph.AddWaiter("tx1", "resource1", 1)

		node := graph.GetNode("tx1")
		assert.NotNil(t, node)
		assert.Equal(t, "tx1", node.TransactionID)
		assert.Equal(t, "resource1", node.WaitResource)
		assert.Equal(t, 1, node.Priority)
	})

	t.Run("add holder creates node and edges", func(t *testing.T) {
		graph := NewLockGraph()

		graph.AddHolder("tx1", "resource1")

		node := graph.GetNode("tx1")
		assert.NotNil(t, node)
		assert.True(t, node.HeldResources["resource1"])
	})

	t.Run("remove holder", func(t *testing.T) {
		graph := NewLockGraph()

		graph.AddHolder("tx1", "resource1")
		graph.RemoveHolder("tx1", "resource1")

		node := graph.GetNode("tx1")
		assert.NotNil(t, node)
		assert.False(t, node.HeldResources["resource1"])
	})

	t.Run("remove node", func(t *testing.T) {
		graph := NewLockGraph()

		graph.AddWaiter("tx1", "resource1", 1)
		graph.RemoveNode("tx1")

		node := graph.GetNode("tx1")
		assert.Nil(t, node)
	})
}

// TestLockGraphAddWaiterEdge tests edge creation when adding waiters
func TestLockGraphAddWaiterEdge(t *testing.T) {
	t.Run("edge to holder", func(t *testing.T) {
		graph := NewLockGraph()

		// tx1 holds resource1
		graph.AddHolder("tx1", "resource1")

		// tx2 waits for resource1
		graph.AddWaiter("tx2", "resource1", 2)

		// Should have edge from tx2 to tx1
		node := graph.GetNode("tx2")
		assert.NotNil(t, node)
		assert.True(t, node.HeldResources["resource1"])
	})
}

// TestLockGraphDetectCycles tests cycle detection
func TestLockGraphDetectCycles(t *testing.T) {
	t.Run("no cycles", func(t *testing.T) {
		graph := NewLockGraph()

		graph.AddHolder("tx1", "resource1")
		graph.AddHolder("tx2", "resource2")

		cycles := graph.DetectCycles()
		assert.Empty(t, cycles)
	})

	t.Run("simple cycle", func(t *testing.T) {
		// Create a deadlock scenario:
		// tx1 holds resource1 and waits for resource2
		// tx2 holds resource2 and waits for resource1
		graph := NewLockGraph()

		graph.AddHolder("tx1", "resource1")
		graph.AddWaiter("tx1", "resource2", 1)

		graph.AddHolder("tx2", "resource2")
		graph.AddWaiter("tx2", "resource1", 2)

		cycles := graph.DetectCycles()
		assert.NotEmpty(t, cycles)
		assert.Equal(t, 2, len(cycles[0])) // tx1 -> tx2 -> tx1 or vice versa
	})

	t.Run("cycle with three transactions", func(t *testing.T) {
		// tx1 holds resource1, waits for resource2
		// tx2 holds resource2, waits for resource3
		// tx3 holds resource3, waits for resource1
		graph := NewLockGraph()

		graph.AddHolder("tx1", "resource1")
		graph.AddWaiter("tx1", "resource2", 1)

		graph.AddHolder("tx2", "resource2")
		graph.AddWaiter("tx2", "resource3", 2)

		graph.AddHolder("tx3", "resource3")
		graph.AddWaiter("tx3", "resource1", 3)

		cycles := graph.DetectCycles()
		assert.NotEmpty(t, cycles)
	})

	t.Run("independent waiters", func(t *testing.T) {
		// Two transactions waiting for the same resource
		graph := NewLockGraph()

		graph.AddHolder("tx1", "resource1")
		graph.AddWaiter("tx2", "resource1", 2)
		graph.AddWaiter("tx3", "resource1", 3)

		cycles := graph.DetectCycles()
		// No cycle because tx1 doesn't wait for anything
		assert.Empty(t, cycles)
	})
}

// TestLockGraphCleanup tests stale entry cleanup
func TestLockGraphCleanup(t *testing.T) {
	t.Run("cleanup removes old entries", func(t *testing.T) {
		graph := NewLockGraph()
		graph.cleanupInterval = 1 * time.Second

		// Add a waiter with old timestamp
		graph.AddWaiter("tx1", "resource1", 1)
		graph.lastCleanup = time.Now().Add(-10 * time.Second)

		// Manually set wait start to old time
		node := graph.GetNode("tx1")
		if node != nil {
			node.WaitStart = time.Now().Add(-15 * time.Minute)
		}

		graph.Cleanup()

		// The node should be removed
		assert.Nil(t, graph.GetNode("tx1"))
	})
}

// TestDeadlockDetectorConfig tests deadlock configuration
func TestDeadlockDetectorConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultDeadlockConfig()

		assert.Equal(t, 5*time.Second, config.DetectionInterval)
		assert.True(t, config.WaitDieEnabled)
		assert.False(t, config.WoundWaitEnabled)
		assert.True(t, config.TimeoutEnabled)
		assert.Equal(t, 30*time.Second, config.DefaultTimeout)
		assert.Equal(t, 10*time.Second, config.MaxWaitTime)
		assert.True(t, config.EnableFencingToken)
	})
}

// TestDeadlockDetectorStartStop tests detector lifecycle
func TestDeadlockDetector(t *testing.T) {
	detectedDeadlocks := make(chan *DeadlockInfo, 10)

	detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)

	t.Run("start and stop", func(t *testing.T) {
		detector.Start()

		// Wait briefly
		time.Sleep(100 * time.Millisecond)

		detector.Stop()
	})

	t.Run("no deadlock channel", func(t *testing.T) {
		detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), nil)
		detector.Start()

		time.Sleep(100 * time.Millisecond)

		detector.Stop()
	})
}

// TestDeadlockDetectorResolveWaitDie tests wait-die resolution
func TestDeadlockDetectorResolveWaitDie(t *testing.T) {
	t.Run("wait-die kills younger transaction", func(t *testing.T) {
		detectedDeadlocks := make(chan *DeadlockInfo, 10)
		detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)
		detector.config.WaitDieEnabled = true
		detector.config.TimeoutEnabled = false

		// Simulate a cycle: tx1 (priority 1) -> tx2 (priority 2)
		cycle := []string{"tx1", "tx2"}

		// Add nodes with priorities
		detector.lockGraph.AddWaiter("tx1", "resource1", 1)
		detector.lockGraph.AddWaiter("tx2", "resource2", 2)

		// Manually add edge for test
		detector.lockGraph.addEdge("tx1", "tx2")
		detector.lockGraph.addEdge("tx2", "tx1")

		resolution := detector.resolveDeadlock(cycle)
		assert.Equal(t, ResolutionAborted, resolution)

		// tx2 should be removed (higher priority = younger = killed)
		assert.Nil(t, detector.lockGraph.GetNode("tx2"))
	})

	t.Run("single transaction cycle", func(t *testing.T) {
		detectedDeadlocks := make(chan *DeadlockInfo, 10)
		detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)

		cycle := []string{"tx1"}

		resolution := detector.resolveWaitDie(cycle)
		assert.Equal(t, ResolutionRecovered, resolution)
	})
}

// TestDeadlockDetectorResolveTimeout tests timeout resolution
func TestDeadlockDetectorResolveTimeout(t *testing.T) {
	t.Run("timeout kills oldest waiter", func(t *testing.T) {
		detectedDeadlocks := make(chan *DeadlockInfo, 10)
		detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)
		detector.config.WaitDieEnabled = false
		detector.config.TimeoutEnabled = true
		detector.config.MaxWaitTime = 1 * time.Second

		cycle := []string{"tx1", "tx2"}

		// Add nodes with different wait times
		detector.lockGraph.AddWaiter("tx1", "resource1", 1)
		detector.lockGraph.AddWaiter("tx2", "resource2", 2)

		// Set tx1's wait start to be older
		node1 := detector.lockGraph.GetNode("tx1")
		if node1 != nil {
			node1.WaitStart = time.Now().Add(-5 * time.Second)
		}

		// Set tx2's wait start to be recent
		node2 := detector.lockGraph.GetNode("tx2")
		if node2 != nil {
			node2.WaitStart = time.Now().Add(-500 * time.Millisecond)
		}

		// Add edges
		detector.lockGraph.addEdge("tx1", "tx2")
		detector.lockGraph.addEdge("tx2", "tx1")

		resolution := detector.resolveTimeout(cycle)
		assert.Equal(t, ResolutionTimeout, resolution)

		// tx1 should be killed (oldest)
		assert.Nil(t, detector.lockGraph.GetNode("tx1"))
	})

	t.Run("no timeout yet", func(t *testing.T) {
		detectedDeadlocks := make(chan *DeadlockInfo, 10)
		detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)
		detector.config.MaxWaitTime = 10 * time.Second

		cycle := []string{"tx1", "tx2"}

		// Add recent waiters
		detector.lockGraph.AddWaiter("tx1", "resource1", 1)
		detector.lockGraph.AddWaiter("tx2", "resource2", 2)

		resolution := detector.resolveTimeout(cycle)
		assert.Equal(t, ResolutionRecovered, resolution)
	})
}

// TestDeadlockDetectorBuildInfo tests deadlock info building
func TestDeadlockDetectorBuildInfo(t *testing.T) {
	detectedDeadlocks := make(chan *DeadlockInfo, 10)
	detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)

	// Setup a simple cycle
	detector.lockGraph.AddHolder("tx1", "resource1")
	detector.lockGraph.AddWaiter("tx2", "resource1", 2)

	cycle := []string{"tx2"}

	info := detector.buildDeadlockInfo(cycle)

	assert.NotEmpty(t, info.ID)
	assert.NotZero(t, info.DetectedAt)
	assert.Len(t, info.Transactions, 1)
	assert.Equal(t, "tx2", info.Transactions[0].ID)
	assert.Contains(t, info.Transactions[0].ResourcesHeld, "resource1")
}

// TestDeadlockDetectorRegister tests registration of lock operations
func TestDeadlockDetectorRegister(t *testing.T) {
	detectedDeadlocks := make(chan *DeadlockInfo, 10)
	detector := NewDeadlockDetector(nil, nil, shared.NewLogger(nil), detectedDeadlocks)

	t.Run("register wait", func(t *testing.T) {
		detector.RegisterWait("tx1", "resource1", 1)

		node := detector.lockGraph.GetNode("tx1")
		assert.NotNil(t, node)
		assert.Equal(t, "resource1", node.WaitResource)
	})

	t.Run("register hold", func(t *testing.T) {
		detector.RegisterHold("tx1", "resource1")

		node := detector.lockGraph.GetNode("tx1")
		assert.NotNil(t, node)
		assert.True(t, node.HeldResources["resource1"])
	})

	t.Run("register release", func(t *testing.T) {
		detector.RegisterHold("tx1", "resource1")
		detector.RegisterRelease("tx1", "resource1")

		node := detector.lockGraph.GetNode("tx1")
		assert.NotNil(t, node)
		assert.False(t, node.HeldResources["resource1"])
	})
}

// TestWaitDiePolicy tests the wait-die policy
func TestWaitDiePolicy(t *testing.T) {
	logger := shared.NewLogger(nil)
	policy := NewWaitDiePolicy(logger)

	t.Run("older transaction waits", func(t *testing.T) {
		policy.RegisterTransaction("older")
		policy.RegisterTransaction("newer")

		// Older transaction requests resource held by newer
		canWait := policy.CanWait("older", "newer")
		assert.True(t, canWait)
	})

	t.Run("newer transaction dies", func(t *testing.T) {
		policy.RegisterTransaction("older")
		policy.RegisterTransaction("newer")

		// Newer transaction requests resource held by older
		canWait := policy.CanWait("newer", "older")
		assert.False(t, canWait)
	})

	t.Run("missing transaction defaults to wait", func(t *testing.T) {
		canWait := policy.CanWait("unknown1", "unknown2")
		assert.True(t, canWait)
	})

	t.Run("transaction lifecycle", func(t *testing.T) {
		policy.RegisterTransaction("tx1")
		policy.AddResource("tx1", "resource1")

		state := getTransactionState(policy, "tx1")
		assert.NotNil(t, state)
		assert.True(t, state.Resources["resource1"])

		policy.RemoveResource("tx1", "resource1")
		state = getTransactionState(policy, "tx1")
		assert.False(t, state.Resources["resource1"])

		policy.UnregisterTransaction("tx1")
		state = getTransactionState(policy, "tx1")
		assert.Nil(t, state)
	})
}

// TestFencingTokenHandler tests fencing token handling
func TestFencingTokenHandler(t *testing.T) {
	handler := NewFencingTokenHandler()

	t.Run("initial token acceptance", func(t *testing.T) {
		// No tokens recorded, should accept any
		valid := handler.ValidateToken("resource1", 1)
		assert.True(t, valid)
	})

	t.Run("update and validate", func(t *testing.T) {
		handler.UpdateToken("resource1", 5)

		// Valid tokens
		assert.True(t, handler.ValidateToken("resource1", 5))
		assert.True(t, handler.ValidateToken("resource1", 10))

		// Invalid token
		assert.False(t, handler.ValidateToken("resource1", 3))
	})

	t.Run("increment token", func(t *testing.T) {
		token1 := handler.IncrementToken("resource1")
		token2 := handler.IncrementToken("resource1")

		assert.Equal(t, int64(1), token1)
		assert.Equal(t, int64(2), token2)
	})

	t.Run("get current token", func(t *testing.T) {
		handler.IncrementToken("resource1")
		handler.IncrementToken("resource2")

		assert.Equal(t, int64(2), handler.GetCurrentToken("resource1"))
		assert.Equal(t, int64(1), handler.GetCurrentToken("resource2"))
		assert.Equal(t, int64(0), handler.GetCurrentToken("resource3"))
	})
}

// TestDeadlockInfo tests deadlock info structure
func TestDeadlockInfo(t *testing.T) {
	t.Run("resolution types", func(t *testing.T) {
		assert.Equal(t, DeadlockResolution("aborted"), ResolutionAborted)
		assert.Equal(t, DeadlockResolution("rolled_back"), ResolutionRolledBack)
		assert.Equal(t, DeadlockResolution("timeout"), ResolutionTimeout)
		assert.Equal(t, DeadlockResolution("recovered"), ResolutionRecovered)
	})
}

// TestTransactionInfo tests transaction info structure
func TestTransactionInfo(t *testing.T) {
	info := &TransactionInfo{
		ID:            "tx1",
		ResourcesHeld: []string{"resource1", "resource2"},
		ResourcesWait: []string{"resource3"},
		StartTime:     time.Now(),
		WaitDuration:  5 * time.Second,
		Priority:      1,
	}

	assert.Equal(t, "tx1", info.ID)
	assert.Len(t, info.ResourcesHeld, 2)
	assert.Len(t, info.ResourcesWait, 1)
}

// Helper function to get transaction state for testing
func getTransactionState(policy *WaitDiePolicy, txID string) *TransactionState {
	policy.mu.RLock()
	defer policy.mu.RUnlock()
	return policy.transactions[txID]
}
