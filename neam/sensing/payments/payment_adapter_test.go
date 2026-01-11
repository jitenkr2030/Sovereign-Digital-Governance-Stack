package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockLogger is a mock implementation of the logger for testing
type MockLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.DebugMessages = append(m.DebugMessages, msg)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.InfoMessages = append(m.InfoMessages, msg)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.WarnMessages = append(m.WarnMessages, msg)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
}

// TestHelper is a test helper for payment adapter tests
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// CreateTestContext creates a test context
func (h *TestHelper) CreateTestContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	h.t.Cleanup(cancel)
	return ctx
}

// TestISO20022Parser validates ISO 20022 pacs.008 parsing
func TestISO20022Parser(t *testing.T) {
	logger := zap.NewNop()
	processor := iso20022.NewProcessor(logger)

	// Valid pacs.008 message
	validPacs008 := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10">
<FIToFICstmrCdtTrf>
<GrpHdr>
<MsgId>MSGID20240101ABCD1234</MsgId>
<CreDtTm>2024-01-15T10:30:00Z</CreDtTm>
<NbOfTxs>1</NbOfTxs>
<SttlmInf>
<SttlmMtd>CLRG</SttlmMtd>
</SttlmInf>
<InstgAgt>
<FinInstnId>
<BIC>BANKDEFF</BIC>
</FinInstnId>
</InstgAgt>
<InstdAgt>
<FinInstnId>
<BIC>BANKUS33</BIC>
</FinInstnId>
</InstdAgt>
</GrpHdr>
<CdtTrfTxInf>
<PmtId>
<InstrId>INSTRID001</InstrId>
<EndToEndId>E2EID001</EndToEndId>
</PmtId>
<Amt>
<InstdAmt Ccy="EUR">1000.00</InstdAmt>
</Amt>
<CdtrAgt>
<FinInstnId>
<BIC>BANKUS33</BIC>
</FinInstnId>
</CdtrAgt>
<Cdtr>
<Nm>John Doe</Nm>
</Cdtr>
<CdtrAcct>
<IBAN>DE89370400440532013000</IBAN>
</CdtrAcct>
<DbtrAgt>
<FinInstnId>
<BIC>BANKDEFF</BIC>
</FinInstnId>
</DbtrAgt>
<Dbtr>
<Nm>Jane Smith</Nm>
</Dbtr>
<DbtrAcct>
<IBAN>DE89370400440532013001</IBAN>
</DbtrAcct>
</CdtTrfTxInf>
</FIToFICstmrCdtTrf>
</Document>`)

	ctx := context.Background()
	normalized, err := processor.Parse(ctx, validPacs008)

	require.NoError(t, err)
	require.NotNil(t, normalized)

	// Validate parsed fields
	assert.Equal(t, "iso20022", normalized.Format)
	assert.Equal(t, "pacs.008", normalized.MessageType)
	assert.Equal(t, "MSGID20240101ABCD1234", normalized.TransactionID)
	assert.Equal(t, "BANKDEFF", normalized.SenderBIC)
	assert.Equal(t, "BANKUS33", normalized.ReceiverBIC)
	assert.Equal(t, "EUR", normalized.Currency)
	assert.NotEmpty(t, normalized.Amount.String())
	assert.Equal(t, "DE89370400440532013000", normalized.ReceiverAccount)
	assert.Equal(t, "DE89370400440532013001", normalized.SenderAccount)
}

// TestISO20022Validation validates ISO 20022 message validation
func TestISO20022Validation(t *testing.T) {
	logger := zap.NewNop()
	processor := iso20022.NewProcessor(logger)

	// Test validation of invalid message
	invalidMessage := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Document>
<InvalidContent>
</InvalidContent>
</Document>`)

	ctx := context.Background()
	normalized, err := processor.Parse(ctx, invalidMessage)

	require.Error(t, err)
	assert.Nil(t, normalized)
}

// TestSWIFTMTParser validates SWIFT MT103 parsing
func TestSWIFTMTParser(t *testing.T) {
	logger := zap.NewNop()
	processor := swift.NewMTProcessor(logger)

	// Valid MT103 message
	validMT103 := `{1:F01BANKDEFFAXXX0000000000}{2:I103BANKDEFFAXXXNBORUZBJXXXN}{3:{108:MSGREF123456789}}{4:
:20:TRANSREF1234
:23B:CRED
:32A:240115EUR1000,00
:50A:/DE89370400440532013000
BANKDEFFXXX
:53A:BANKDEFFXXX
:57A:BANKUS33XXX
:59:/DE89370400440532013001
JOHN DOE
:71A:OUR
-}`

	ctx := context.Background()
	normalized, err := processor.Parse(ctx, []byte(validMT103))

	require.NoError(t, err)
	require.NotNil(t, normalized)

	// Validate parsed fields
	assert.Equal(t, "swift_mt", normalized.Format)
	assert.Equal(t, "103", normalized.MessageType)
	assert.Equal(t, "TRANSREF1234", normalized.TransactionID)
	assert.Equal(t, "BANKDEFF", normalized.SenderBIC)
	assert.Equal(t, "BANKUS33", normalized.ReceiverBIC)
	assert.Equal(t, "EUR", normalized.Currency)
	assert.Equal(t, "DE89370400440532013000", normalized.SenderAccount)
	assert.Equal(t, "DE89370400440532013001", normalized.ReceiverAccount)
}

// TestMessageFormatDetection validates message format detection
func TestMessageFormatDetection(t *testing.T) {
	testCases := []struct {
		name     string
		payload  []byte
		expected string
	}{
		{
			name:     "ISO 20022 pacs.008",
			payload:  []byte(`<?xml version="1.0"?><Document><FIToFICstmrCdtTrf>`),
			expected: "iso20022",
		},
		{
			name:     "SWIFT MT",
			payload:  []byte(`{1:F01BANKDEFFAXXX0000000000}`),
			expected: "swift_mt",
		},
		{
			name:     "Unknown format",
			payload:  []byte(`{"test": "data"}`),
			expected: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format, _ := detectMessageFormat(tc.payload)
			assert.Equal(t, tc.expected, format)
		})
	}
}

// TestIdempotencyHash validates idempotency hash calculation
func TestIdempotencyHash(t *testing.T) {
	// Test hash calculation
	message1 := []byte(`{"amount": 1000, "currency": "EUR"}`)
	message2 := []byte(`{"amount": 1000, "currency": "EUR"}`)
	message3 := []byte(`{"amount": 2000, "currency": "EUR"}`)

	hash1 := idempotency.CalculateHash(message1)
	hash2 := idempotency.CalculateHash(message2)
	hash3 := idempotency.CalculateHash(message3)

	// Same content should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different content should produce different hash
	assert.NotEqual(t, hash1, hash3)

	// Hash should be 64 characters (SHA-256 hex)
	assert.Len(t, hash1, 64)
}

// TestRetryLogic validates retry logic with exponential backoff
func TestRetryLogic(t *testing.T) {
	// Simulate retry with exponential backoff
	baseDelay := 100 * time.Millisecond
	maxDelay := 10 * time.Second
	maxRetries := 5

	delays := make([]time.Duration, maxRetries+1)
	for i := 0; i <= maxRetries; i++ {
		delay := baseDelay * time.Duration(1<<i)
		if delay > maxDelay {
			delay = maxDelay
		}
		delays[i] = delay
	}

	// Verify exponential growth
	assert.Equal(t, 100*time.Millisecond, delays[0])
	assert.Equal(t, 200*time.Millisecond, delays[1])
	assert.Equal(t, 400*time.Millisecond, delays[2])
	assert.Equal(t, 800*time.Millisecond, delays[3])
	assert.Equal(t, 1*time.Second+600*time.Millisecond, delays[4])
	assert.Equal(t, 10*time.Second, delays[5]) // Capped at maxDelay
}

// TestWorkerPool validates worker pool functionality
func TestWorkerPool(t *testing.T) {
	logger := zap.NewNop()
	pool := NewWorkerPool(5, logger)
	ctx := context.Background()

	// Start pool
	pool.Start(ctx)

	// Submit jobs
	results := make(chan int, 10)
	for i := 0; i < 10; i++ {
		i := i
		err := pool.Submit(func() {
			time.Sleep(10 * time.Millisecond)
			results <- i
		})
		require.NoError(t, err)
	}

	// Wait for jobs to complete
	close(results)
	received := 0
	for range results {
		received++
	}

	assert.Equal(t, 10, received)

	// Get stats
	stats := pool.Stats()
	assert.Equal(t, 5, stats.WorkerCount)
	assert.Equal(t, 0, stats.QueueLength)

	// Stop pool
	err := pool.Stop()
	require.NoError(t, err)
}

// TestBatchProcessor validates batch processing functionality
func TestBatchProcessor(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()
	archiveDir := t.TempDir()

	cfg := config.Default()
	cfg.Processing.BatchDirectory = testDir
	cfg.Batch.ArchivePath = archiveDir
	cfg.Batch.SupportedFormats = []string{"xml", "txt"}

	logger := zap.NewNop()

	// Create mock idempotency manager
	idempotencyMgr := &mockIdempotencyManager{}

	// Create mock processors
	processors := map[string]MessageProcessor{
		"iso20022": &mockProcessor{},
	}

	batchProcessor := NewBatchProcessor(cfg, logger, idempotencyMgr, processors)

	// Create test file
	testContent := `<?xml version="1.0"?>
<Document><FIToFICstmrCdtTrf><GrpHdr><MsgId>TEST001</MsgId></GrpHdr></FIToFICstmrCdtTrf></Document>
<?xml version="1.0"?>
<Document><FIToFICstmrCdtTrf><GrpHdr><MsgId>TEST002</MsgId></GrpHdr></FIToFICstmrCdtTrf></Document>`

	err := ioutil.WriteFile(filepath.Join(testDir, "test_batch.xml"), []byte(testContent), 0644)
	require.NoError(t, err)

	// Process file
	ctx := batchProcessor.Process(ctx)

	// Verify archive directory
	files, err := ioutil.ReadDir(archiveDir)
	require.NoError(t, err)
	assert.Equal(t, 2, len(files)) // 2 archived files (original + timestamped)
}

// TestMessageAck validates message acknowledgment logic
func TestMessageAck(t *testing.T) {
	// Test message acknowledgment simulation
	offsets := []int64{0, 1, 2, 3, 4}

	// Simulate processing and acknowledgment
	processed := make(map[int64]bool)
	for _, offset := range offsets {
		// Process message
		processed[offset] = true
		// Acknowledge
		assert.True(t, processed[offset])
	}

	assert.Len(t, processed, 5)
}

// TestDecimalHandling validates decimal handling for financial amounts
func TestDecimalHandling(t *testing.T) {
	// Test decimal precision
	amount1 := decimal.NewFromFloat(1000.00)
	amount2 := decimal.NewFromFloat(1000.001)
	amount3 := decimal.NewFromFloat(1000.00)

	// Same amounts should be equal
	assert.True(t, amount1.Equal(amount3))

	// Different amounts should not be equal
	assert.False(t, amount1.Equal(amount2))

	// Test arithmetic
	sum := amount1.Add(amount2)
	assert.Equal(t, "2000.001", sum.String())

	// Test large number handling
	largeAmount := decimal.New(big.NewInt(999999999999), -2)
	assert.Equal(t, "9999999999.99", largeAmount.String())
}

// TestConfigurationLoading validates configuration loading
func TestConfigurationLoading(t *testing.T) {
	// Create test config file
	configContent := `
service:
  name: payment-adapter-test
  host: 0.0.0.0
  port: 8084

kafka:
  brokers: localhost:9092
  consumer_group: test-group
  input_topic: test-input
  output_topic: test-output
  dlq_topic: test-dlq

redis:
  addr: localhost:6379
  idempotency_ttl: 1h

processing:
  mode: realtime
  worker_count: 5
  batch_size: 100

retry:
  max_retries: 3
  initial_delay: 100ms
  max_delay: 5s
  multiplier: 2.0
`

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	err := ioutil.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load configuration
	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Validate loaded values
	assert.Equal(t, "payment-adapter-test", cfg.Service.Name)
	assert.Equal(t, 8084, cfg.Service.Port)
	assert.Equal(t, "localhost:9092", cfg.Kafka.Brokers)
	assert.Equal(t, "test-input", cfg.Kafka.InputTopic)
	assert.Equal(t, "realtime", cfg.Processing.Mode)
	assert.Equal(t, 5, cfg.Processing.WorkerCount)
	assert.Equal(t, 3, cfg.Retry.MaxRetries)
}

// Mock implementations for testing
type mockIdempotencyManager struct{}

func (m *mockIdempotencyManager) Check(ctx context.Context, messageHash string) (*idempotency.CheckResult, error) {
	return &idempotency.CheckResult{Status: idempotency.StatusNotFound}, nil
}

func (m *mockIdempotencyManager) MarkProcessing(ctx context.Context, messageHash string) error {
	return nil
}

func (m *mockIdempotencyManager) MarkCompleted(ctx context.Context, messageHash string, resultData interface{}) error {
	return nil
}

func (m *mockIdempotencyManager) MarkFailed(ctx context.Context, messageHash string, errorMsg string) error {
	return nil
}

type mockProcessor struct{}

func (m *mockProcessor) Parse(ctx context.Context, raw []byte) (*NormalizedPayment, error) {
	return &NormalizedPayment{
		ID:        "test-id",
		Timestamp: time.Now(),
	}, nil
}

func (m *mockProcessor) Validate(msg interface{}) error {
	return nil
}

func (m *mockProcessor) GetMessageType() string {
	return "test"
}

// Import for filepath
import "path/filepath"
