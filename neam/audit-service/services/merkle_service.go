package services

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
)

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash  []byte
	Left  *MerkleNode
	Right *MerkleNode
	Leaf  bool
}

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	Root     *MerkleNode
	Leaves   []*MerkleNode
	TreeSize int
}

// MerkleService provides Merkle tree operations
type MerkleService struct {
	hashFunc crypto.Hash
}

// NewMerkleService creates a new MerkleService
func NewMerkleService(hashService crypto.HashService) *MerkleService {
	return &MerkleService{
		hashFunc: sha256.New(),
	}
}

// BuildTree builds a Merkle tree from leaf hashes
func (m *MerkleService) BuildTree(leaves [][]byte) (*MerkleTree, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("no leaves provided")
	}

	// Create leaf nodes
	tree := &MerkleTree{
		Leaves:   make([]*MerkleNode, len(leaves)),
		TreeSize: len(leaves),
	}

	for i, leaf := range leaves {
		tree.Leaves[i] = &MerkleNode{
			Hash: m.hashLeaf(leaf),
			Leaf: true,
		}
	}

	// Build tree bottom-up
	tree.Root = m.buildTreeRecursive(tree.Leaves)

	return tree, nil
}

// buildTreeRecursive recursively builds the tree
func (m *MerkleService) buildTreeRecursive(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 1 {
		return nodes[0]
	}

	// Ensure even number of nodes
	if len(nodes)%2 == 1 {
		// Duplicate the last node
		nodes = append(nodes, nodes[len(nodes)-1])
	}

	// Process pairs
	var nextLevel []*MerkleNode
	for i := 0; i < len(nodes); i += 2 {
		hash := m.hashPair(nodes[i].Hash, nodes[i+1].Hash)
		nextLevel = append(nextLevel, &MerkleNode{
			Hash:  hash,
			Left:  nodes[i],
			Right: nodes[i+1],
		})
	}

	return m.buildTreeRecursive(nextLevel)
}

// hashLeaf hashes a leaf value
func (m *MerkleService) hashLeaf(data []byte) []byte {
	m.hashFunc.Reset()
	m.hashFunc.Write([]byte("LEAF:"))
	m.hashFunc.Write(data)
	return m.hashFunc.Sum(nil)
}

// hashPair hashes a pair of child hashes
func (m *MerkleService) hashPair(left, right []byte) []byte {
	m.hashFunc.Reset()
	m.hashFunc.Write([]byte("NODE:"))
	m.hashFunc.Write(left)
	m.hashFunc.Write(right)
	return m.hashFunc.Sum(nil)
}

// GetProof generates a Merkle proof for a leaf at index
func (m *MerkleService) GetProof(tree *MerkleTree, index int) ([][]byte, error) {
	if index < 0 || index >= len(tree.Leaves) {
		return nil, fmt.Errorf("invalid leaf index")
	}

	var proof [][]byte
	m.generateProof(tree.Root, 0, len(tree.Leaves), index, &proof)

	return proof, nil
}

// generateProof recursively generates proof
func (m *MerkleService) generateProof(node *MerkleNode, start, end, target int, proof *[][]byte) {
	if node == nil || node.Leaf {
		return
	}

	leftSize := (end - start) / 2
	leftStart := start
	leftEnd := start + leftSize
	rightStart := leftEnd
	rightEnd := end

	// Determine which branch contains the target
	if target < leftSize {
		// Target is in left subtree, add right sibling to proof
		if node.Right != nil {
			*proof = append(*proof, node.Right.Hash)
		}
		m.generateProof(node.Left, leftStart, leftEnd, target, proof)
	} else {
		// Target is in right subtree, add left sibling to proof
		if node.Left != nil {
			*proof = append(*proof, node.Left.Hash)
		}
		m.generateProof(node.Right, rightStart, rightEnd, target-leftSize, proof)
	}
}

// VerifyProof verifies a Merkle proof
func (m *MerkleService) VerifyProof(rootHash []byte, leafHash []byte, proof [][]byte, leafIndex int) bool {
	// Recompute the root from the leaf and proof
	computedHash := leafHash

	// Sort proof by level (not strictly necessary for binary trees but helps)
	// For simplicity, we process proof in order

	for i, siblingHash := range proof {
		// Determine if this node is left or right sibling
		// This depends on the index; for binary trees, we alternate
		if (leafIndex+i)%2 == 0 {
			// Current hash is left sibling
			computedHash = m.hashPair(computedHash, siblingHash)
		} else {
			// Current hash is right sibling
			computedHash = m.hashPair(siblingHash, computedHash)
		}
	}

	// Compare with root
	return string(computedHash) == string(rootHash)
}

// GetRootHash returns the root hash of the tree
func (m *MerkleService) GetRootHash(tree *MerkleTree) []byte {
	if tree == nil || tree.Root == nil {
		return nil
	}
	return tree.Root.Hash
}

// GetLeafHash returns the hash of a leaf at index
func (m *MerkleService) GetLeafHash(tree *MerkleTree, index int) ([]byte, error) {
	if index < 0 || index >= len(tree.Leaves) {
		return nil, fmt.Errorf("invalid leaf index")
	}
	return tree.Leaves[index].Hash, nil
}

// GetTreeDepth returns the depth of the tree
func (m *MerkleService) GetTreeDepth(tree *MerkleTree) int {
	if tree == nil || tree.Root == nil {
		return 0
	}

	depth := 0
	node := tree.Root
	for node != nil && !node.Leaf {
		depth++
		node = node.Left
	}
	return depth
}

// SerializeTree serializes the tree to a compact format
func (m *MerkleService) SerializeTree(tree *MerkleTree) ([]byte, error) {
	if tree == nil {
		return nil, fmt.Errorf("tree is nil")
	}

	var hashes []string
	var indices []int64

	// Collect all hashes with their indices
	m.collectHashes(tree.Root, 0, &hashes, &indices)

	serialized := struct {
		RootHash  string   `json:"root_hash"`
		TreeSize  int      `json:"tree_size"`
		Depth     int      `json:"depth"`
		Hashes    []string `json:"hashes"`
		Indices   []int64  `json:"indices"`
	}{
		RootHash:  hex.EncodeToString(tree.Root.Hash),
		TreeSize:  tree.TreeSize,
		Depth:     m.GetTreeDepth(tree),
		Hashes:    hashes,
		Indices:   indices,
	}

	return json.Marshal(serialized)
}

// collectHashes recursively collects hashes
func (m *MerkleService) collectHashes(node *MerkleNode, index int64, hashes *[]string, indices *[]int64) {
	if node == nil {
		return
	}

	*hashes = append(*hashes, hex.EncodeToString(node.Hash))
	*indices = append(*indices, index)

	if node.Left != nil {
		m.collectHashes(node.Left, index*2, hashes, indices)
	}
	if node.Right != nil {
		m.collectHashes(node.Right, index*2+1, hashes, indices)
	}
}

// CreateAuditChain creates a chain of events using Merkle tree
func (m *MerkleService) CreateAuditChain(events []EventWithHash) (*AuditChain, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no events provided")
	}

	// Sort events by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// Build Merkle tree from event hashes
	var leaves [][]byte
	for _, event := range events {
		leaves = append(leaves, []byte(event.IntegrityHash))
	}

	tree, err := m.BuildTree(leaves)
	if err != nil {
		return nil, err
	}

	chain := &AuditChain{
		FirstEventID:    events[0].ID,
		LastEventID:     events[len(events)-1].ID,
		CurrentRootHash: hex.EncodeToString(tree.Root.Hash),
		Length:          int64(len(events)),
	}

	return chain, nil
}

// EventWithHash represents an event with its hash
type EventWithHash struct {
	ID             string
	Timestamp      interface{}
	IntegrityHash  string
}

// AuditChain represents a chain of audit events
type AuditChain struct {
	FirstEventID    string `json:"first_event_id"`
	LastEventID     string `json:"last_event_id"`
	CurrentRootHash string `json:"current_root_hash"`
	Length          int64  `json:"length"`
}

// CompactProof creates a compact proof for court-grade verification
func (m *MerkleService) CompactProof(proof [][]byte) []byte {
	var compact []byte

	for _, hash := range proof {
		// Encode each hash as length-prefixed
		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(len(hash)))
		compact = append(compact, length...)
		compact = append(compact, hash...)
	}

	return compact
}

// ExpandProof expands a compact proof
func (m *MerkleService) ExpandProof(compact []byte) ([][]byte, error) {
	var proof [][]byte
	offset := 0

	for offset < len(compact) {
		if offset+2 > len(compact) {
			return nil, fmt.Errorf("invalid compact proof format")
		}

		length := binary.BigEndian.Uint16(compact[offset : offset+2])
		offset += 2

		if offset+int(length) > len(compact) {
			return nil, fmt.Errorf("invalid compact proof format")
		}

		hash := make([]byte, length)
		copy(hash, compact[offset:offset+int(length)])
		proof = append(proof, hash)
		offset += int(length)
	}

	return proof, nil
}
