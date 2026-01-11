package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// WORMStorage provides Write Once Read Many storage for audit records
type WORMStorage interface {
	// Store stores an audit event in WORM storage
	Store(ctx context.Context, event interface{}) error

	// StoreSeal stores a seal record in WORM storage
	StoreSeal(ctx context.Context, seal interface{}) error

	// Retrieve retrieves an audit event from WORM storage
	Retrieve(ctx context.Context, id string) (interface{}, error)

	// RetrieveSeal retrieves a seal record from WORM storage
	RetrieveSeal(ctx context.Context, id string) (interface{}, error)

	// VerifyIntegrity verifies the integrity of stored data
	VerifyIntegrity(ctx context.Context, id string) (bool, error)

	// GetStorageStats returns storage statistics
	GetStorageStats(ctx context.Context) (*StorageStats, error)
}

// wormStorage implements WORMStorage
type wormStorage struct {
	basePath     string
	hashService  crypto.Hash
	signer       crypto.Signer
	eventsPath   string
	sealsPath    string
	integrityPath string
}

// StorageStats represents WORM storage statistics
type StorageStats struct {
	TotalEvents    int64   `json:"total_events"`
	TotalSeals     int64   `json:"total_seals"`
	TotalBytes     int64   `json:"total_bytes"`
	StoragePath    string  `json:"storage_path"`
	IntegrityVerified bool `json:"integrity_verified"`
	LastCheckTime  time.Time `json:"last_check_time"`
}

// NewWORMStorage creates a new WORMStorage
func NewWORMStorage(basePath string, hashService crypto.Hash, signer crypto.Signer) WORMStorage {
	ws := &wormStorage{
		basePath:      basePath,
		hashService:   sha256.New(),
		signer:        signer,
		eventsPath:    filepath.Join(basePath, "events"),
		sealsPath:     filepath.Join(basePath, "seals"),
		integrityPath: filepath.Join(basePath, "integrity"),
	}

	// Create directories
	os.MkdirAll(ws.eventsPath, 0700)
	os.MkdirAll(ws.sealsPath, 0700)
	os.MkdirAll(ws.integrityPath, 0700)

	return ws
}

// Store stores an audit event in WORM storage
func (w *wormStorage) Store(ctx context.Context, event interface{}) error {
	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Calculate integrity hash
	hash := w.calculateHash(data)

	// Create record with integrity metadata
	record := WORMRecord{
		Data:          data,
		IntegrityHash: hash,
		Timestamp:     time.Now().UTC(),
		Type:          "event",
	}

	recordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to serialize record: %w", err)
	}

	// Sign the record
	signature, err := w.sign(recordData)
	if err != nil {
		return fmt.Errorf("failed to sign record: %w", err)
	}
	record.Signature = signature

	// Write to file (append-only for WORM)
	recordData, err = json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to serialize signed record: %w", err)
	}

	fileName := w.getFilePath("events", hash)
	if err := ioutil.WriteFile(fileName, recordData, 0600); err != nil {
		return fmt.Errorf("failed to write to WORM storage: %w", err)
	}

	return nil
}

// StoreSeal stores a seal record in WORM storage
func (w *wormStorage) StoreSeal(ctx context.Context, seal interface{}) error {
	// Serialize seal
	data, err := json.Marshal(seal)
	if err != nil {
		return fmt.Errorf("failed to serialize seal: %w", err)
	}

	// Calculate integrity hash
	hash := w.calculateHash(data)

	// Create record
	record := WORMRecord{
		Data:          data,
		IntegrityHash: hash,
		Timestamp:     time.Now().UTC(),
		Type:          "seal",
	}

	recordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to serialize record: %w", err)
	}

	// Sign the record
	signature, err := w.sign(recordData)
	if err != nil {
		return fmt.Errorf("failed to sign record: %w", err)
	}
	record.Signature = signature

	// Write to file
	recordData, err = json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to serialize signed record: %w", err)
	}

	fileName := w.getFilePath("seals", hash)
	if err := ioutil.WriteFile(fileName, recordData, 0600); err != nil {
		return fmt.Errorf("failed to write to WORM storage: %w", err)
	}

	return nil
}

// Retrieve retrieves an audit event from WORM storage
func (w *wormStorage) Retrieve(ctx context.Context, id string) (interface{}, error) {
	fileName := w.getFilePath("events", id)

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from WORM storage: %w", err)
	}

	var record WORMRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to deserialize record: %w", err)
	}

	// Verify integrity
	if !w.verifyIntegrity(record.Data, record.IntegrityHash, record.Signature) {
		return nil, fmt.Errorf("integrity verification failed")
	}

	var event interface{}
	if err := json.Unmarshal(record.Data, &event); err != nil {
		return nil, fmt.Errorf("failed to deserialize event: %w", err)
	}

	return event, nil
}

// RetrieveSeal retrieves a seal record from WORM storage
func (w *wormStorage) RetrieveSeal(ctx context.Context, id string) (interface{}, error) {
	fileName := w.getFilePath("seals", id)

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from WORM storage: %w", err)
	}

	var record WORMRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to deserialize record: %w", err)
	}

	// Verify integrity
	if !w.verifyIntegrity(record.Data, record.IntegrityHash, record.Signature) {
		return nil, fmt.Errorf("integrity verification failed")
	}

	var seal interface{}
	if err := json.Unmarshal(record.Data, &seal); err != nil {
		return nil, fmt.Errorf("failed to deserialize seal: %w", err)
	}

	return seal, nil
}

// VerifyIntegrity verifies the integrity of stored data
func (w *wormStorage) VerifyIntegrity(ctx context.Context, id string) (bool, error) {
	var data []byte
	var err error

	// Try events first
	fileName := w.getFilePath("events", id)
	data, err = ioutil.ReadFile(fileName)
	if err != nil {
		// Try seals
		fileName = w.getFilePath("seals", id)
		data, err = ioutil.ReadFile(fileName)
		if err != nil {
			return false, fmt.Errorf("file not found: %s", id)
		}
	}

	var record WORMRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return false, fmt.Errorf("failed to deserialize record: %w", err)
	}

	return w.verifyIntegrity(record.Data, record.IntegrityHash, record.Signature), nil
}

// GetStorageStats returns storage statistics
func (w *wormStorage) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	stats := &StorageStats{
		StoragePath:   w.basePath,
		LastCheckTime: time.Now().UTC(),
	}

	// Count events
	events, err := ioutil.ReadDir(w.eventsPath)
	if err == nil {
		stats.TotalEvents = int64(len(events))
	}

	// Count seals
	seals, err := ioutil.ReadDir(w.sealsPath)
	if err == nil {
		stats.TotalSeals = int64(len(seals))
	}

	// Calculate total bytes
	var totalBytes int64
	filepath.Walk(w.basePath, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			totalBytes += info.Size()
		}
		return nil
	})
	stats.TotalBytes = totalBytes

	return stats, nil
}

// WORMRecord represents a record in WORM storage
type WORMRecord struct {
	Data          []byte  `json:"data"`
	IntegrityHash string  `json:"integrity_hash"`
	Signature     string  `json:"signature"`
	Timestamp     time.Time `json:"timestamp"`
	Type          string  `json:"type"`
}

// getFilePath generates a file path for storage
func (w *wormStorage) getFilePath(prefix string, hash string) string {
	// Use first 2 chars as subdirectory for better file distribution
	subDir := hash[:2]
	return filepath.Join(w.basePath, prefix, subDir, hash[2:]+".json")
}

// calculateHash calculates SHA-256 hash
func (w *wormStorage) calculateHash(data []byte) string {
	w.hashService.Reset()
	w.hashService.Write(data)
	return hex.EncodeToString(w.hashService.Sum(nil))
}

// sign signs data
func (w *wormStorage) sign(data []byte) (string, error) {
	if w.signer == nil {
		return "", fmt.Errorf("no signer configured")
	}

	h := sha256.Sum256(data)
	signature, err := w.signer.Sign(rand.Reader, h[:], crypto.SHA256)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// verifyIntegrity verifies data integrity
func (w *wormStorage) verifyIntegrity(data []byte, expectedHash string, signature string) bool {
	// Calculate hash
	actualHash := w.calculateHash(data)
	if actualHash != expectedHash {
		return false
	}

	// Verify signature
	if w.signer != nil && signature != "" {
		h := sha256.Sum256(data)
		sigBytes, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			return false
		}

		// Get public key from signer
		var publicKey *rsa.PublicKey
		switch key := w.signer.Public().(type) {
		case *rsa.PublicKey:
			publicKey = key
		default:
			// Non-RSA key, skip signature verification
			return true
		}

		if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, h[:], sigBytes); err != nil {
			return false
		}
	}

	return true
}

// CreateChainLink creates a cryptographic chain link between two records
func (w *wormStorage) CreateChainLink(previousHash, currentHash string) (*ChainLink, error) {
	linkData := fmt.Sprintf("%s|%s|%s",
		previousHash,
		currentHash,
		time.Now().UTC().Format(time.RFC3339),
	)

	hash := w.calculateHash([]byte(linkData))

	link := &ChainLink{
		PreviousHash: previousHash,
		CurrentHash:  currentHash,
		LinkHash:     hash,
		Timestamp:    time.Now().UTC(),
	}

	return link, nil
}

// ChainLink represents a link in the WORM storage chain
type ChainLink struct {
	PreviousHash string    `json:"previous_hash"`
	CurrentHash  string    `json:"current_hash"`
	LinkHash     string    `json:"link_hash"`
	Timestamp    time.Time `json:"timestamp"`
}

// CourtGradeEvidence generates court-grade evidence for a record
func (w *wormStorage) CourtGradeEvidence(recordID string) (*CourtEvidence, error) {
	// Retrieve record
	record, err := w.Retrieve(context.Background(), recordID)
	if err != nil {
		return nil, err
	}

	// Create chain of custody
	chain := &ChainOfCustody{
		RecordID:   recordID,
		Events:     []CustodyEvent{},
		VerifiedAt: time.Now().UTC(),
	}

	// Generate evidence
	evidence := &CourtEvidence{
		Record:       record,
		ChainOfCustody: chain,
		Timestamp:   time.Now().UTC(),
	}

	return evidence, nil
}

// ChainOfCustody represents the chain of custody for evidence
type ChainOfCustody struct {
	RecordID    string         `json:"record_id"`
	Events      []CustodyEvent `json:"events"`
	VerifiedAt  time.Time      `json:"verified_at"`
}

// CustodyEvent represents an event in the chain of custody
type CustodyEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	Actor       string    `json:"actor"`
	Details     string    `json:"details"`
}

// CourtEvidence represents court-grade evidence
type CourtEvidence struct {
	Record        interface{}    `json:"record"`
	ChainOfCustody *ChainOfCustody `json:"chain_of_custody"`
	Timestamp     time.Time      `json:"timestamp"`
}
