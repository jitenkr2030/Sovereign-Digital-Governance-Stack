package testing

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"neam-platform/sensing/common/config"
	"neam-platform/sensing/common/logger"
)

// TestConfig holds common test configuration
type TestConfig struct {
	DatabaseDSN      string
	RedisAddr        string
	KafkaBrokers     []string
	SchemaRegistryURL string
	OTLPEndpoint     string
	LogLevel         string
	Environment      string
}

// GetTestConfig returns configuration for testing
func GetTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseDSN:      getEnv("TEST_DATABASE_DSN", "postgresql://neam_test:neam_test@localhost:5432/neam_test?sslmode=disable"),
		RedisAddr:        getEnv("TEST_REDIS_ADDR", "localhost:6379"),
		KafkaBrokers:     strings.Split(getEnv("TEST_KAFKA_BROKERS", "localhost:9092"), ","),
		SchemaRegistryURL: getEnv("TEST_SCHEMA_REGISTRY_URL", "http://localhost:8081"),
		OTLPEndpoint:     getEnv("TEST_OTLP_ENDPOINT", "http://localhost:4317"),
		LogLevel:         getEnv("TEST_LOG_LEVEL", "debug"),
		Environment:      getEnv("TEST_ENVIRONMENT", "test"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestDatabase wraps database testing utilities
type TestDatabase struct {
	DB     *sql.DB
	Mock   sqlmock.Sqlmock
	closed bool
}

// NewTestDatabase creates a new test database connection with mocking
func NewTestDatabase(t *testing.T) (*TestDatabase, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	td := &TestDatabase{
		DB:   db,
		Mock: mock,
	}

	cleanup := func() {
		if !td.closed {
			td.closed = true
			db.Close()
		}
	}

	return td, cleanup
}

// MustQuery simulates a database query with mock data
func (td *TestDatabase) MustQuery(t *testing.T, query string, args ...interface{}) *sqlmock.Rows {
	t.Helper()
	rows := sqlmock.NewRows([]string{})
	for i, arg := range args {
		rows = rows.AddRow(fmt.Sprintf("result_%d_%v", i, arg))
	}
	td.Mock.ExpectQuery(query).WithArgs(args...).WillReturnRows(rows)
	return rows
}

// ExpectExec simulates a database execution
func (td *TestDatabase) ExpectExec(t *testing.T, query string) *sqlmock.ExpectedExec {
	t.Helper()
	return td.Mock.ExpectExec(query)
}

// TestRedis wraps Redis testing utilities
type TestRedis struct {
	Client *redis.Client
	Mock   *redis.Client
	closed bool
}

// NewTestRedis creates a new test Redis client
func NewTestRedis(t *testing.T) (*TestRedis, func()) {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:     GetTestConfig().RedisAddr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		// If connection fails, use mock
		t.Logf("Redis connection failed, using mock: %v", err)
	}

	tr := &TestRedis{
		Client: client,
	}

	cleanup := func() {
		if !tr.closed {
			tr.closed = true
			client.Close()
		}
	}

	return tr, cleanup
}

// MockRedisClient creates a mocked Redis client
func MockRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "mock-redis:6379",
	})
}

// TestLogger provides a test logger instance
func NewTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.New(&config.LoggerConfig{
		Level:      GetTestConfig().LogLevel,
		Format:     "json",
		OutputPath: "stdout",
	})
	require.NoError(t, err)
	return log
}

// TestContext provides a test context with timeout
func NewTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestDataPath returns the path to test data directory
func TestDataPath(t *testing.T, filename string) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "testdata", filename)
}

// EnsureTestDataDir creates the test data directory if it doesn't exist
func EnsureTestDataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := filepath.Join(wd, "testdata")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	return dir
}

// CreateTestFile creates a temporary test file with the given content
func CreateTestFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := EnsureTestDataDir(t)
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// AssertPanic tests that a function panics with the expected value
func AssertPanic(t *testing.T, expectedPanic interface{}, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		assert.Equal(t, expectedPanic, r, "Expected panic with value %v, got %v", expectedPanic, r)
	}()
	f()
}

// AssertNoError is a simple error assertion helper
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NoError(t, err, msgAndArgs...)
}

// AssertError checks that an error is of the expected type
func AssertError(t *testing.T, err error, msg string, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msg, msgAndArgs...)
	assert.Contains(t, err.Error(), msg, msgAndArgs...)
}

// SetupTestModule initializes a test module with all dependencies
type SetupTestModule struct {
	t          *testing.T
	databases  []*TestDatabase
	redis      *TestRedis
	logger     *logger.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	mocks      []interface{}
}

// NewSetupTestModule creates a new test module setup
func NewSetupTestModule(t *testing.T) *SetupTestModule {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	return &SetupTestModule{
		t:      t,
		ctx:    ctx,
		cancel: cancel,
		mocks:  make([]interface{}, 0),
	}
}

// WithDatabase adds a test database to the setup
func (sm *SetupTestModule) WithDatabase() *SetupTestModule {
	sm.t.Helper()
	db, cleanup := NewTestDatabase(sm.t)
	sm.databases = append(sm.databases, db)
	sm.t.Cleanup(cleanup)
	return sm
}

// WithRedis adds a test Redis client to the setup
func (sm *SetupTestModule) WithRedis() *SetupTestModule {
	sm.t.Helper()
	redis, cleanup := NewTestRedis(sm.t)
	sm.redis = redis
	sm.t.Cleanup(cleanup)
	return sm
}

// WithLogger adds a test logger to the setup
func (sm *SetupTestModule) WithLogger() *SetupTestModule {
	sm.t.Helper()
	sm.logger = NewTestLogger(sm.t)
	return sm
}

// GetDatabase returns the first test database
func (sm *SetupTestModule) GetDatabase() *sql.DB {
	if len(sm.databases) > 0 {
		return sm.databases[0].DB
	}
	return nil
}

// GetRedis returns the test Redis client
func (sm *SetupTestModule) GetRedis() *redis.Client {
	if sm.redis != nil {
		return sm.redis.Client
	}
	return nil
}

// GetLogger returns the test logger
func (sm *SetupTestModule) GetLogger() *logger.Logger {
	return sm.logger
}

// GetContext returns the test context
func (sm *SetupTestModule) GetContext() context.Context {
	return sm.ctx
}

// AddMock adds a mock object to the setup
func (sm *SetupTestModule) AddMock(m interface{}) *SetupTestModule {
	sm.mocks = append(sm.mocks, m)
	return sm
}

// Mockery interface for mocking
type Mockery interface {
	On(methodName string, args ...interface{}) *mock.Call
}

// SetupMock creates a new mock for the given interface
func SetupMock(t *testing.T, mockery interface{}) *mock.Mock {
	t.Helper()
	if m, ok := mockery.(*mock.Mock); ok {
		return m
	}
	m := new(mock.Mock)
	return m
}

// CreateMockInterface creates a mock implementation of an interface
func CreateMockInterface[T any](t *testing.T) *mock.Mock {
	t.Helper()
	m := new(mock.Mock)
	return m
}

// MockMethod mocks a specific method on a mock object
func MockMethod(mockObj *mock.Mock, methodName string, args ...interface{}) *mock.Call {
	return mockObj.On(methodName, args...)
}

// AssertExpectations verifies all mock expectations
func AssertExpectations(t *testing.T, mocks ...interface{}) {
	t.Helper()
	for _, m := range mocks {
		if mockObj, ok := m.(*mock.Mock); ok {
			mockObj.AssertExpectations(t)
		}
	}
}

// BenchmarkResult holds benchmark testing results
type BenchmarkResult struct {
	Name           string
	Iterations     int
	NsPerOp        float64
	BytesAlloc     int64
	AllocsPerOp    int64
	TotalDuration  time.Duration
}

// RunBenchmark runs a function as a benchmark
func RunBenchmark(b *testing.B, name string, f func(b *testing.B)) BenchmarkResult {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		f(b)
	})

	return BenchmarkResult{
		Name:           name,
		Iterations:     b.N,
		NsPerOp:        float64(b.Elapsed()) / float64(b.N) * 1000000,
		BytesAlloc:     b.Bytes,
		AllocsPerOp:    int64(b.AllocedPerProcess()),
		TotalDuration:  b.Elapsed(),
	}
}

// AssertBenchmarkResult validates benchmark results against thresholds
func AssertBenchmarkResult(t *testing.T, result BenchmarkResult, maxNsPerOp float64, maxAllocsPerOp int64) {
	t.Helper()
	assert.Less(t, result.NsPerOp, maxNsPerOp, "Benchmark %s took too long: %f ns/op", result.Name, result.NsPerOp)
	assert.Less(t, result.AllocsPerOp, maxAllocsPerOp, "Benchmark %s allocated too much: %d allocs/op", result.Name, result.AllocsPerOp)
}

// TestHelpers provides common test assertion helpers
type TestHelpers struct{}

// NewTestHelpers creates a new TestHelpers instance
func NewTestHelpers() *TestHelpers {
	return &TestHelpers{}
}

// EqualFloat compares two float values with a tolerance
func (th *TestHelpers) EqualFloat(t *testing.T, expected, actual, tolerance float64, msg string) {
	t.Helper()
	diff := expected - actual
	assert.Less(t, diff, tolerance, msg)
	assert.Greater(t, -diff, tolerance, msg)
}

// ContainsAll checks that a slice contains all expected values
func (th *TestHelpers) ContainsAll(t *testing.T, slice, expected []string, msg string) {
	t.Helper()
	for _, e := range expected {
		assert.Contains(t, slice, e, msg)
	}
}

// UniqueSlice removes duplicates from a slice
func UniqueSlice[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
