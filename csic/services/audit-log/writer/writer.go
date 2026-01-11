// Audit Log Writer - Immutable Audit Entry Storage
// Provides WORM (Write Once Read Many) storage for audit log entries

package writer

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/csic-platform/services/audit-log"
)

// AuditLogWriter handles immutable writing of audit log entries
type AuditLogWriter struct {
	storagePath    string
	enableWORM     bool
	currentFile    *os.File
	currentFileNum uint32
	fileMu         sync.Mutex
	mu             sync.RWMutex
	sequenceNum    uint64
	writtenEntries map[string]*writerEntry
}

// writerEntry represents an entry in the writer's memory
type writerEntry struct {
	entry     *audit.AuditLogEntry
	fileNum   uint32
	fileOffset int64
	writtenAt time.Time
}

// NewAuditLogWriter creates a new audit log writer
func NewAuditLogWriter(storagePath string, enableWORM bool) *AuditLogWriter {
	// Ensure storage directory exists
	if err := os.MkdirAll(storagePath, 0700); err != nil {
		panic(fmt.Sprintf("failed to create storage directory: %v", err))
	}

	// Ensure sequences directory exists
	seqDir := filepath.Join(storagePath, "sequences")
	if err := os.MkdirAll(seqDir, 0700); err != nil {
		panic(fmt.Sprintf("failed to create sequences directory: %v", err))
	}

	writer := &AuditLogWriter{
		storagePath:    storagePath,
		enableWORM:     enableWORM,
		writtenEntries: make(map[string]*writerEntry),
	}

	// Initialize sequence number from disk
	writer.loadSequenceNumber()

	return writer
}

// Write writes a new audit log entry
func (w *AuditLogWriter) Write(ctx context.Context, entry *audit.AuditLogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Generate entry ID if not provided
	if entry.EntryID == "" {
		entry.EntryID = generateEntryID()
	}

	// Assign sequence number
	entry.SequenceNum = w.nextSequenceNumber()
	entry.Timestamp = time.Now().UTC()

	// Generate chain ID (could be configurable per environment)
	if entry.ChainID == "" {
		entry.ChainID = "csic-main-chain"
	}

	// Calculate and set the hash
	w.mu.RLock()
	prevHash := w.getLastHash()
	w.mu.RUnlock()

	entry.PreviousHash = prevHash
	entry.CurrentHash = w.calculateHash(entry)

	// Write to storage
	if err := w.writeToStorage(entry); err != nil {
		return fmt.Errorf("failed to write entry to storage: %w", err)
	}

	// Track in memory
	w.writtenEntries[entry.EntryID] = &writerEntry{
		entry:     entry,
		fileNum:   w.currentFileNum,
		writtenAt: entry.Timestamp,
	}

	return nil
}

// WriteBatch writes multiple audit log entries
func (w *AuditLogWriter) WriteBatch(ctx context.Context, entries []*audit.AuditLogEntry) error {
	for _, entry := range entries {
		if err := w.Write(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

// Read reads a specific audit log entry
func (w *AuditLogWriter) Read(ctx context.Context, entryID string) (*audit.AuditLogEntry, error) {
	w.mu.RLock()
	we, exists := w.writtenEntries[entryID]
	w.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("entry not found: %s", entryID)
	}

	// Read from file
	return w.readFromStorage(we.fileNum, we.fileOffset)
}

// Query queries audit log entries with filters
func (w *AuditLogWriter) Query(ctx context.Context, query *audit.AuditQuery) ([]*audit.AuditLogEntry, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var results []*audit.AuditLogEntry

	for _, we := range w.writtenEntries {
		if w.matchesQuery(we.entry, query) {
			results = append(results, we.entry)
		}
	}

	// Apply pagination
	if query.Limit > 0 {
		start := query.Offset
		if start >= len(results) {
			return []*audit.AuditLogEntry{}, nil
		}
		end := start + query.Limit
		if end > len(results) {
			end = len(results)
		}
		results = results[start:end]
	}

	return results, nil
}

// GetSequenceNumber returns the current sequence number
func (w *AuditLogWriter) GetSequenceNumber() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.sequenceNum
}

// GetLastHash returns the hash of the last written entry
func (w *AuditLogWriter) GetLastHash() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.getLastHash()
}

// getLastHash returns the hash without locking
func (w *AuditLogWriter) getLastHash() string {
	if len(w.writtenEntries) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}

	// Find the entry with the highest sequence number
	var lastEntry *audit.AuditLogEntry
	for _, we := range w.writtenEntries {
		if lastEntry == nil || we.entry.SequenceNum > lastEntry.SequenceNum {
			lastEntry = we.entry
		}
	}

	return lastEntry.CurrentHash
}

// nextSequenceNumber increments and returns the next sequence number
func (w *AuditLogWriter) nextSequenceNumber() uint64 {
	w.sequenceNum++
	w.saveSequenceNumber()
	return w.sequenceNum
}

// loadSequenceNumber loads the sequence number from disk
func (w *AuditLogWriter) loadSequenceNumber() {
	seqFile := filepath.Join(w.storagePath, "sequences", "sequence.dat")

	data, err := os.ReadFile(seqFile)
	if err != nil {
		if os.IsNotExist(err) {
			w.sequenceNum = 0
			return
		}
		panic(fmt.Sprintf("failed to read sequence file: %v", err))
	}

	if len(data) < 8 {
		w.sequenceNum = 0
		return
	}

	w.sequenceNum = binary.BigEndian.Uint64(data)
}

// saveSequenceNumber saves the sequence number to disk
func (w *AuditLogWriter) saveSequenceNumber() {
	seqFile := filepath.Join(w.storagePath, "sequences", "sequence.dat")

	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, w.sequenceNum)

	if err := os.WriteFile(seqFile, data, 0600); err != nil {
		// Log error but don't fail
		fmt.Printf("failed to save sequence number: %v\n", err)
	}
}

// writeToStorage writes an entry to the storage file
func (w *AuditLogWriter) writeToStorage(entry *audit.AuditLogEntry) error {
	w.fileMu.Lock()
	defer w.fileMu.Unlock()

	// Rotate file if needed (every 10000 entries or 100MB)
	if w.shouldRotateFile() {
		if err := w.rotateFile(); err != nil {
			return fmt.Errorf("failed to rotate file: %w", err)
		}
	}

	// Open file for appending
	filePath := w.getCurrentFilePath()
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get offset before writing
	offset, err := file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return fmt.Errorf("failed to get file offset: %w", err)
	}

	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Write length prefix for record boundaries
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := file.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if w.enableWORM {
		// Make file read-only after write
		if err := os.Chmod(filePath, 0400); err != nil {
			return fmt.Errorf("failed to set WORM permissions: %w", err)
		}
	}

	return nil
}

// readFromStorage reads an entry from the storage file
func (w *AuditLogWriter) readFromStorage(fileNum uint32, offset int64) (*audit.AuditLogEntry, error) {
	w.fileMu.Lock()
	defer w.fileMu.Unlock()

	filePath := w.getFilePath(fileNum)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Seek to offset
	if _, err := file.Seek(offset+4, os.SEEK_SET); err != nil {
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Read length
	lengthBuf := make([]byte, 4)
	if _, err := file.Read(lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// Read data
	data := make([]byte, length)
	if _, err := file.Read(data); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	var entry audit.AuditLogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	return &entry, nil
}

// shouldRotateFile determines if the current file should be rotated
func (w *AuditLogWriter) shouldRotateFile() bool {
	if w.currentFileNum == 0 {
		return true
	}

	filePath := w.getCurrentFilePath()
	info, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	// Rotate if file is larger than 100MB
	return info.Size() > 100*1024*1024
}

// rotateFile rotates to a new storage file
func (w *AuditLogWriter) rotateFile() error {
	if w.currentFile != nil {
		if err := w.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close current file: %w", err)
		}
	}

	w.currentFileNum++
	w.currentFile = nil

	return nil
}

// getCurrentFilePath returns the path to the current storage file
func (w *AuditLogWriter) getCurrentFilePath() string {
	return w.getFilePath(w.currentFileNum)
}

// getFilePath returns the path to a specific storage file
func (w *AuditLogWriter) getFilePath(fileNum uint32) string {
	return filepath.Join(w.storagePath, fmt.Sprintf("audit_%08d.log", fileNum))
}

// calculateHash calculates the SHA-256 hash of an entry
func (w *AuditLogWriter) calculateHash(entry *audit.AuditLogEntry) string {
	// Create canonical representation for hashing
	canonical := struct {
		EntryID     string
		SequenceNum uint64
		Timestamp   time.Time
		ChainID     string
		ActorID     string
		Service     string
		Operation   string
		ActionType  string
		Resource    string
		Result      string
		PreviousHash string
	}{
		EntryID:      entry.EntryID,
		SequenceNum:  entry.SequenceNum,
		Timestamp:    entry.Timestamp,
		ChainID:      entry.ChainID,
		ActorID:      entry.ActorID,
		Service:      entry.Service,
		Operation:    entry.Operation,
		ActionType:   entry.ActionType,
		Resource:     entry.Resource,
		Result:       entry.Result,
		PreviousHash: entry.PreviousHash,
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// matchesQuery checks if an entry matches the query filters
func (w *AuditLogWriter) matchesQuery(entry *audit.AuditLogEntry, query *audit.AuditQuery) bool {
	if query.StartTime.IsZero() && query.EndTime.IsZero() &&
		query.ActorID == "" && query.Service == "" &&
		query.Operation == "" && query.ActionType == "" &&
		query.Resource == "" && query.Result == "" &&
		query.RiskLevel == "" && query.SequenceFrom == 0 &&
		query.SequenceTo == 0 {
		return true
	}

	// Time range check
	if !query.StartTime.IsZero() && entry.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && entry.Timestamp.After(query.EndTime) {
		return false
	}

	// Sequence range check
	if query.SequenceFrom > 0 && entry.SequenceNum < query.SequenceFrom {
		return false
	}
	if query.SequenceTo > 0 && entry.SequenceNum > query.SequenceTo {
		return false
	}

	// String filters
	if query.ActorID != "" && entry.ActorID != query.ActorID {
		return false
	}
	if query.Service != "" && entry.Service != query.Service {
		return false
	}
	if query.Operation != "" && entry.Operation != query.Operation {
		return false
	}
	if query.ActionType != "" && entry.ActionType != query.ActionType {
		return false
	}
	if query.Resource != "" && entry.Resource != query.Resource {
		return false
	}
	if query.Result != "" && entry.Result != query.Result {
		return false
	}
	if query.RiskLevel != "" && entry.RiskLevel != query.RiskLevel {
		return false
	}

	return true
}

// generateEntryID generates a unique entry identifier
func generateEntryID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)

	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[:8], uint64(timestamp))
	copy(data[8:], randomBytes[:8])

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:32])
}
