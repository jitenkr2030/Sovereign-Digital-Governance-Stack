package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SignatureType represents the type of cryptographic signature
type SignatureType string

const (
	SignatureHMAC     SignatureType = "HMAC"
	SignatureRSA2048  SignatureType = "RSA2048"
	SignatureRSA4096  SignatureType = "RSA4096"
)

// SignatureRecord represents a cryptographic signature
type SignatureRecord struct {
	ID             string       `json:"id"`
	RequestID      string       `json:"request_id"`
	ActorID        string       `json:"actor_id"`
	ActionType     string       `json:"action_type"`
	Signature      string       `json:"signature"`
	PublicKey      string       `json:"public_key,omitempty"`
	SignatureType  SignatureType `json:"signature_type"`
	Timestamp      time.Time    `json:"timestamp"`
	Nonce          string       `json:"nonce"`
	HashAlgorithm  string       `json:"hash_algorithm"`
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Algorithm  string `json:"algorithm"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// LedgerEntry represents an entry in the immutable ledger
type LedgerEntry struct {
	EntryID      string    `json:"entry_id"`
	PreviousHash string    `json:"previous_hash"`
	CurrentHash  string    `json:"current_hash"`
	Data         string    `json:"data"`
	Timestamp    time.Time `json:"timestamp"`
	Signature    string    `json:"signature"`
	SequenceNum  int       `json:"sequence_num"`
}

// Signer provides cryptographic signing capabilities
type Signer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
}

// NewSigner creates a new signer with the specified key pair
func NewSigner(keyPair KeyPair) (*Signer, error) {
	block, _ := pem.Decode([]byte(keyPair.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Signer{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      keyPair.PublicKey,
	}, nil
}

// NewSignerFromPEM creates a new signer from PEM strings
func NewSignerFromPEM(privateKeyPEM, publicKeyPEM string) (*Signer, error) {
	privateBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if privateBlock == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Signer{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyID:      publicKeyPEM,
	}, nil
}

// Sign signs data using RSA-PSS with SHA-256
func (s *Signer) Sign(data string) (string, error) {
	hashed := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPSS(rand.Reader, s.privateKey, crypto.SHA256, hashed[:], nil)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify verifies a signature
func (s *Signer) Verify(data, signature string) bool {
	hashed := sha256.Sum256([]byte(data))
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	err = rsa.VerifyPSS(s.publicKey, crypto.SHA256, hashed[:], sigBytes, nil)
	return err == nil
}

// GenerateKeyPair generates a new RSA key pair
func GenerateKeyPair(bits int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	return &KeyPair{
		PublicKey:  string(publicKeyPEM),
		PrivateKey: string(privateKeyPEM),
		Algorithm:  fmt.Sprintf("RSA-%d", bits),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(365 * 24 * time.Hour),
	}, nil
}

// ApprovalSigner provides signing capabilities for approval records
type ApprovalSigner struct {
	signer *Signer
}

// NewApprovalSigner creates a new approval signer
func NewApprovalSigner(signer *Signer) *ApprovalSigner {
	return &ApprovalSigner{signer: signer}
}

// SignApprovalAction signs an approval action record
func (as *ApprovalSigner) SignApprovalAction(
	requestID uuid.UUID,
	actorID uuid.UUID,
	actionType string,
	timestamp time.Time,
	previousHash string,
	nonce string,
) (*SignatureRecord, error) {
	// Create the data to sign
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		requestID.String(),
		actorID.String(),
		actionType,
		timestamp.Format(time.RFC3339),
		previousHash,
		nonce,
	)

	// Generate the hash first
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	// Sign the hash
	signature, err := as.signer.Sign(hashStr)
	if err != nil {
		return nil, err
	}

	return &SignatureRecord{
		ID:            uuid.New().String(),
		RequestID:     requestID.String(),
		ActorID:       actorID.String(),
		ActionType:    actionType,
		Signature:     signature,
		PublicKey:     as.signer.keyID,
		SignatureType: SignatureRSA2048,
		Timestamp:     timestamp,
		Nonce:         nonce,
		HashAlgorithm: "SHA-256",
	}, nil
}

// VerifyApprovalAction verifies an approval action signature
func (as *ApprovalSigner) VerifyApprovalAction(
	record *SignatureRecord,
	previousHash string,
) bool {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		record.RequestID,
		record.ActorID,
		record.ActionType,
		record.Timestamp.Format(time.RFC3339),
		previousHash,
		record.Nonce,
	)

	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return as.signer.Verify(hashStr, record.Signature)
}

// ImmutableLedger provides an immutable ledger for approval records
type ImmutableLedger struct {
	entries    map[string]*LedgerEntry
	headHash   string
	sequence   int
	signer     *ApprovalSigner
}

// NewImmutableLedger creates a new immutable ledger
func NewImmutableLedger(signer *ApprovalSigner) *ImmutableLedger {
	return &ImmutableLedger{
		entries:  make(map[string]*LedgerEntry),
		sequence: 0,
		signer:   signer,
	}
}

// Append adds a new entry to the ledger
func (l *ImmutableLedger) Append(data string) (*LedgerEntry, error) {
	l.sequence++

	// Generate nonce
	nonce := generateSecureNonce()

	// Get the last entry's hash if exists
	previousHash := l.headHash
	if l.sequence > 1 {
		previousHash = l.entries[previousHash].CurrentHash
	}

	// Create entry data
	entryData := fmt.Sprintf("%s|%s|%s|%d",
		data,
		previousHash,
		nonce,
		l.sequence,
	)

	// Generate hash for the entry
	hash := sha256.Sum256([]byte(entryData))
	currentHash := hex.EncodeToString(hash[:])

	// Create the entry
	entry := &LedgerEntry{
		EntryID:      uuid.New().String(),
		PreviousHash: previousHash,
		CurrentHash:  currentHash,
		Data:         data,
		Timestamp:    time.Now(),
		SequenceNum:  l.sequence,
	}

	// Sign the entry
	signature, err := l.signer.signer.Sign(currentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign ledger entry: %w", err)
	}
	entry.Signature = signature

	// Store the entry
	l.entries[entry.EntryID] = entry
	l.headHash = entry.EntryID

	return entry, nil
}

// VerifyChain verifies the integrity of the ledger chain
func (l *ImmutableLedger) VerifyChain() (bool, int, []string) {
	var brokenAt []string
	validCount := 0

	for i := 1; i <= l.sequence; i++ {
		var entry *LedgerEntry
		for _, e := range l.entries {
			if e.SequenceNum == i {
				entry = e
				break
			}
		}

		if entry == nil {
			break
		}

		// Verify hash
		entryData := fmt.Sprintf("%s|%s|%s|%d",
			entry.Data,
			entry.PreviousHash,
			generateNonceFromEntry(entry),
			entry.SequenceNum,
		)
		hash := sha256.Sum256([]byte(entryData))
		expectedHash := hex.EncodeToString(hash[:])

		if entry.CurrentHash != expectedHash {
			brokenAt = append(brokenAt, fmt.Sprintf("Sequence %d: hash mismatch", i))
		} else {
			validCount++
		}
	}

	return len(brokenAt) == 0, validCount, brokenAt
}

// GetEntry retrieves an entry by ID
func (l *ImmutableLedger) GetEntry(entryID string) (*LedgerEntry, bool) {
	entry, ok := l.entries[entryID]
	return entry, ok
}

// GetHead returns the current head of the ledger
func (l *ImmutableLedger) GetHead() *LedgerEntry {
	if l.headHash == "" {
		return nil
	}
	return l.entries[l.headHash]
}

// AuditReport represents an audit report for approval history
type AuditReport struct {
	ReportID        string           `json:"report_id"`
	RequestID       string           `json:"request_id"`
	GeneratedAt     time.Time        `json:"generated_at"`
	GeneratedBy     string           `json:"generated_by"`
	ChainStatus     ChainStatus      `json:"chain_status"`
	Actions         []AuditAction    `json:"actions"`
	ComplianceNotes []string         `json:"compliance_notes"`
	Summary         AuditSummary     `json:"summary"`
}

// ChainStatus represents the status of the hash chain
type ChainStatus struct {
	IsValid        bool     `json:"is_valid"`
	TotalEntries   int      `json:"total_entries"`
	ValidEntries   int      `json:"valid_entries"`
	BrokenAt       []string `json:"broken_at,omitempty"`
	FirstEntryTime time.Time `json:"first_entry_time"`
	LastEntryTime  time.Time `json:"last_entry_time"`
}

// AuditAction represents a single action in the audit report
type AuditAction struct {
	SequenceNum    int       `json:"sequence_num"`
	Timestamp      time.Time `json:"timestamp"`
	ActorID        string    `json:"actor_id"`
	ActionType     string    `json:"action_type"`
	Comments       string    `json:"comments"`
	SignatureValid bool      `json:"signature_valid"`
	IPAddress      string    `json:"ip_address"`
	DelegatedFrom  *string   `json:"delegated_from,omitempty"`
}

// AuditSummary represents a summary of the audit
type AuditSummary struct {
	TotalActions   int       `json:"total_actions"`
	Approvals      int       `json:"approvals"`
	Rejections     int       `json:"rejections"`
	Abstentions    int       `json:"abstentions"`
	DelegatedCount int       `json:"delegated_count"`
	Duration       string    `json:"duration"`
}

// GenerateAuditReport generates an audit report for a request
func GenerateAuditReport(
	requestID string,
	actions []AuditAction,
	chainValid bool,
	firstEntry, lastEntry time.Time,
) *AuditReport {
	summary := AuditSummary{
		TotalActions:   len(actions),
		Duration:       fmt.Sprintf("%v", lastEntry.Sub(firstEntry)),
	}

	complianceNotes := []string{}

	for _, action := range actions {
		switch action.ActionType {
		case "APPROVE":
			summary.Approvals++
		case "REJECT":
			summary.Rejections++
		case "ABSTAIN":
			summary.Abstentions++
		}

		if action.DelegatedFrom != nil {
			summary.DelegatedCount++
			complianceNotes = append(complianceNotes, fmt.Sprintf(
				"Action %d was delegated from user %s", action.SequenceNum, *action.DelegatedFrom,
			))
		}

		if !action.SignatureValid {
			complianceNotes = append(complianceNotes, fmt.Sprintf(
				"WARNING: Signature verification failed for action %d", action.SequenceNum,
			))
		}
	}

	if !chainValid {
		complianceNotes = append(complianceNotes, "CRITICAL: Hash chain integrity compromised")
	}

	return &AuditReport{
		ReportID:    uuid.New().String(),
		RequestID:   requestID,
		GeneratedAt: time.Now(),
		ChainStatus: ChainStatus{
			IsValid:        chainValid,
			TotalEntries:   len(actions),
			ValidEntries:   len(actions),
			FirstEntryTime: firstEntry,
			LastEntryTime:  lastEntry,
		},
		Actions:         actions,
		ComplianceNotes: complianceNotes,
		Summary:         summary,
	}
}

// Helper functions

func generateSecureNonce() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to time-based nonce
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func generateNonceFromEntry(entry *LedgerEntry) string {
	// Extract nonce from entry data for verification
	// This is a simplified version
	return entry.Timestamp.Format(time.RFC3339Nano)
}
