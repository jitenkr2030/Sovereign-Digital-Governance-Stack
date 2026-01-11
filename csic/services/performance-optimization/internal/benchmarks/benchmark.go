package benchmarks

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/csic-platform/services/performance-optimization/internal/metrics"
	"go.uber.org/zap"
)

// BenchmarkType represents the type of benchmark
type BenchmarkType string

const (
	BenchmarkTypeAll     BenchmarkType = "all"
	BenchmarkTypeCPU     BenchmarkType = "cpu"
	BenchmarkTypeMemory  BenchmarkType = "memory"
	BenchmarkTypeNetwork BenchmarkType = "network"
	BenchmarkTypeDisk    BenchmarkType = "disk"
	BenchmarkTypeCustom  BenchmarkType = "custom"
)

// BenchmarkResult represents the result of a benchmark
type BenchmarkResult struct {
	ID           string            `json:"id"`
	Type         BenchmarkType     `json:"type"`
	Target       string            `json:"target"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Duration     time.Duration     `json:"duration"`
	Iterations   int               `json:"iterations"`
	Concurrency  int               `json:"concurrency"`
	Metrics      BenchmarkMetrics  `json:"metrics"`
	Statistics   BenchmarkStats    `json:"statistics"`
	Errors       []BenchmarkError  `json:"errors,omitempty"`
	Status       string            `json:"status"`
	Config       BenchmarkConfig   `json:"config"`
}

// BenchmarkMetrics contains the measured metrics
type BenchmarkMetrics struct {
	TotalOps        int64           `json:"total_ops"`
	SuccessfulOps   int64           `json:"successful_ops"`
	FailedOps       int64           `json:"failed_ops"`
	TotalDuration   time.Duration   `json:"total_duration"`
	MinDuration     time.Duration   `json:"min_duration"`
	MaxDuration     time.Duration   `json:"max_duration"`
	AvgDuration     time.Duration   `json:"avg_duration"`
	Throughput      float64         `json:"throughput"`
	BytesProcessed  int64           `json:"bytes_processed"`
	BytesPerSec     float64         `json:"bytes_per_sec"`
	LatencyP50      time.Duration   `json:"latency_p50"`
	LatencyP95      time.Duration   `json:"latency_p95"`
	LatencyP99      time.Duration   `json:"latency_p99"`
	ErrorRate       float64         `json:"error_rate"`
	CPUUsage        float64         `json:"cpu_usage_percent"`
	MemoryUsage     uint64          `json:"memory_usage_bytes"`
	GCOverhead      float64         `json:"gc_overhead_percent"`
}

// BenchmarkStatistics contains statistical analysis
type BenchmarkStats struct {
	StdDev          time.Duration   `json:"std_dev"`
	Variance        float64         `json:"variance"`
	Skewness        float64         `json:"skewness"`
	Kurtosis        float64         `json:"kurtosis"`
	ConfidenceLevel float64         `json:"confidence_level"`
	SampleSize      int             `json:"sample_size"`
	Outliers        int             `json:"outliers"`
	Anomalies       int             `json:"anomalies"`
}

// BenchmarkError represents an error during benchmarking
type BenchmarkError struct {
	Timestamp   time.Time `json:"timestamp"`
	Operation   string    `json:"operation"`
	Error       string    `json:"error"`
	Count       int       `json:"count"`
}

// BenchmarkConfig contains benchmark configuration
type BenchmarkConfig struct {
	Type        BenchmarkType     `json:"type"`
	Duration    time.Duration     `json:"duration"`
	Concurrency int               `json:"concurrency"`
	Iterations  int               `json:"iterations"`
	Warmup      time.Duration     `json:"warmup"`
	Timeout     time.Duration     `json:"timeout"`
	OutputPath  string            `json:"output_path"`
	CustomWork  WorkloadFunction  `json:"-"`
}

// WorkloadFunction is a custom workload function
type WorkloadFunction func(ctx context.Context, iter int) error

// Service provides benchmarking capabilities
type Service struct {
	logger     *zap.Logger
	config     *BenchmarkConfig
	collector  *metrics.Collector
	results    map[string]*BenchmarkResult
	mu         sync.RWMutex
}

// Global service instance
var (
	defaultService *Service
	serviceMu      sync.RWMutex
)

// Initialize initializes the benchmark service
func Initialize(logger *zap.Logger, cfg interface{}) error {
	serviceMu.Lock()
	defer serviceMu.Unlock()

	// Extract config
	config := extractConfig(cfg)

	defaultService = &Service{
		logger:    logger,
		config:    config,
		collector: metrics.GetCollector(),
		results:   make(map[string]*BenchmarkResult),
	}

	logger.Info("Benchmark service initialized",
		zap.String("default_type", string(config.Type)),
		zap.Duration("default_duration", config.Duration),
		zap.Int("default_concurrency", config.Concurrency))

	return nil
}

// GetService returns the default benchmark service
func GetService() *Service {
	serviceMu.RLock()
	defer serviceMu.RUnlock()
	return defaultService
}

// NewBenchmark creates a new benchmark configuration
func (s *Service) NewBenchmark(target string, benchType BenchmarkType) *BenchmarkBuilder {
	return &BenchmarkBuilder{
		service:  s,
		target:   target,
		benchType: benchType,
		config: BenchmarkConfig{
			Type:        benchType,
			Duration:    60 * time.Second,
			Concurrency: 1,
			Iterations:  1000,
			Warmup:      10 * time.Second,
			Timeout:     5 * time.Minute,
		},
	}
}

// Run executes a benchmark and returns results
func (s *Service) Run(ctx context.Context, target string, config BenchmarkConfig) (*BenchmarkResult, error) {
	result := &BenchmarkResult{
		ID:          generateID(),
		Type:        config.Type,
		Target:      target,
		StartTime:   time.Now(),
		Config:      config,
		Status:      "running",
		Metrics:     BenchmarkMetrics{MinDuration: math.MaxDuration},
		Statistics:  BenchmarkStats{SampleSize: 0},
	}

	s.mu.Lock()
	s.results[result.ID] = result
	s.mu.Unlock()

	// Start metrics collection
	if s.collector != nil {
		s.collector.StartCollection(ctx, "benchmark", result.ID)
	}

	// Run the appropriate benchmark type
	var err error
	switch config.Type {
	case BenchmarkTypeAll:
		err = s.runAllBenchmarks(ctx, result, config)
	case BenchmarkTypeCPU:
		err = s.runCPUBenchmark(ctx, result, config)
	case BenchmarkTypeMemory:
		err = s.runMemoryBenchmark(ctx, result, config)
	case BenchmarkTypeNetwork:
		err = s.runNetworkBenchmark(ctx, result, config)
	case BenchmarkTypeDisk:
		err = s.runDiskBenchmark(ctx, result, config)
	case BenchmarkTypeCustom:
		err = s.runCustomBenchmark(ctx, result, config)
	default:
		err = fmt.Errorf("unknown benchmark type: %s", config.Type)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = "completed"

	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, BenchmarkError{
			Timestamp: time.Now(),
			Operation: "benchmark",
			Error:     err.Error(),
			Count:     1,
		})
	}

	// Stop metrics collection
	if s.collector != nil {
		s.collector.StopCollection("benchmark", result.ID)
	}

	// Calculate statistics
	s.calculateStats(result)

	s.logger.Info("Benchmark completed",
		zap.String("id", result.ID),
		zap.String("type", string(config.Type)),
		zap.Duration("duration", result.Duration),
		zap.Float64("throughput", result.Metrics.Throughput),
		zap.Float64("error_rate", result.Metrics.ErrorRate))

	return result, err
}

// runAllBenchmarks runs all benchmark types sequentially
func (s *Service) runAllBenchmarks(ctx context.Context, result *BenchmarkResult, config BenchmarkConfig) error {
	benchmarks := []BenchmarkType{
		BenchmarkTypeCPU,
		BenchmarkTypeMemory,
		BenchmarkTypeNetwork,
		BenchmarkTypeDisk,
	}

	for _, bt := range benchmarks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			subConfig := config
			subConfig.Type = bt
			subConfig.Duration = config.Duration / time.Duration(len(benchmarks))

			_, err := s.Run(ctx, result.Target, subConfig)
			if err != nil {
				s.logger.Warn("Sub-benchmark failed",
					zap.String("type", string(bt)),
					zap.Error(err))
			}
		}
	}

	return nil
}

// runCPUBenchmark runs CPU-bound benchmarks
func (s *Service) runCPUBenchmark(ctx context.Context, result *BenchmarkResult, config BenchmarkConfig) error {
	s.logger.Info("Running CPU benchmark",
		zap.String("target", result.Target),
		zap.Duration("duration", config.Duration))

	var wg sync.WaitGroup
	ops := atomic.NewInt64(0)
	errors := atomic.NewInt64(0)
	latencies := make([]time.Duration, 0)
	latencyMu := sync.Mutex{}

	// Create worker pool
	workerCount := config.Concurrency
	if workerCount <= 0 {
		workerCount = 1
	}

	// Calculate work duration per worker
	workDuration := config.Duration / time.Duration(workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			endTime := time.Now().Add(workDuration)
			workerOps := int64(0)
			workerErrors := int64(0)
			workerLatencies := make([]time.Duration, 0)

			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					// CPU-intensive work (compute Pi)
					iterations := 1000
					pi := computePI(iterations)

					_ = pi // Use the result

					latency := time.Since(start)
					workerLatencies = append(workerLatencies, latency)
					workerOps++

					// Progress reporting
					if workerOps%10000 == 0 {
						s.logger.Debug("CPU benchmark progress",
							zap.Int("worker", workerID),
							zap.Int64("ops", workerOps))
					}
				}
			}

			ops.Add(workerOps)
			errors.Add(workerErrors)

			latencyMu.Lock()
			latencies = append(latencies, workerLatencies...)
			latencyMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Update result metrics
	result.Metrics.TotalOps = ops.Load()
	result.Metrics.FailedOps = errors.Load()
	result.Metrics.SuccessfulOps = result.Metrics.TotalOps - result.Metrics.FailedOps
	result.Metrics.TotalDuration = config.Duration
	result.Metrics.Throughput = float64(result.Metrics.TotalOps) / config.Duration.Seconds()

	// Calculate latency percentiles
	if len(latencies) > 0 {
		result.Metrics.AvgDuration = calculateAverage(latencies)
		result.Metrics.MinDuration = calculateMin(latencies)
		result.Metrics.MaxDuration = calculateMax(latencies)
		result.Metrics.LatencyP50 = calculatePercentile(latencies, 50)
		result.Metrics.LatencyP95 = calculatePercentile(latencies, 95)
		result.Metrics.LatencyP99 = calculatePercentile(latencies, 99)
	}

	if result.Metrics.TotalOps > 0 {
		result.Metrics.ErrorRate = float64(result.Metrics.FailedOps) / float64(result.Metrics.TotalOps) * 100
	}

	return nil
}

// runMemoryBenchmark runs memory-intensive benchmarks
func (s *Service) runMemoryBenchmark(ctx context.Context, result *BenchmarkResult, config BenchmarkConfig) error {
	s.logger.Info("Running Memory benchmark",
		zap.String("target", result.Target),
		zap.Duration("duration", config.Duration))

	var wg sync.WaitGroup
	ops := atomic.NewInt64(0)
	errors := atomic.NewInt64(0)
	latencies := make([]time.Duration, 0)
	latencyMu := sync.Mutex{}

	workerCount := config.Concurrency
	if workerCount <= 0 {
		workerCount = 1
	}

	workDuration := config.Duration / time.Duration(workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			endTime := time.Now().Add(workDuration)
			workerOps := int64(0)
			workerErrors := int64(0)
			workerLatencies := make([]time.Duration, 0)

			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					// Memory-intensive work (allocate and process)
					size := 1024 * 1024 // 1MB chunks
					data := make([]byte, size)
					for j := 0; j < size; j += 64 {
						data[j] = byte(j % 256)
					}

					// Process the data
					sum := 0
					for _, b := range data {
						sum += int(b)
					}
					_ = sum

					_ = data // Keep reference to prevent GC

					latency := time.Since(start)
					workerLatencies = append(workerLatencies, latency)
					workerOps++

					if workerOps%100 == 0 {
						s.logger.Debug("Memory benchmark progress",
							zap.Int("worker", workerID),
							zap.Int64("ops", workerOps))
					}
				}
			}

			ops.Add(workerOps)
			errors.Add(workerErrors)

			latencyMu.Lock()
			latencies = append(latencies, workerLatencies...)
			latencyMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Update result metrics
	result.Metrics.TotalOps = ops.Load()
	result.Metrics.FailedOps = errors.Load()
	result.Metrics.SuccessfulOps = result.Metrics.TotalOps - result.Metrics.FailedOps
	result.Metrics.TotalDuration = config.Duration
	result.Metrics.Throughput = float64(result.Metrics.TotalOps) / config.Duration.Seconds()

	if len(latencies) > 0 {
		result.Metrics.AvgDuration = calculateAverage(latencies)
		result.Metrics.MinDuration = calculateMin(latencies)
		result.Metrics.MaxDuration = calculateMax(latencies)
		result.Metrics.LatencyP50 = calculatePercentile(latencies, 50)
		result.Metrics.LatencyP95 = calculatePercentile(latencies, 95)
		result.Metrics.LatencyP99 = calculatePercentile(latencies, 99)
	}

	if result.Metrics.TotalOps > 0 {
		result.Metrics.ErrorRate = float64(result.Metrics.FailedOps) / float64(result.Metrics.TotalOps) * 100
	}

	return nil
}

// runNetworkBenchmark runs network-related benchmarks
func (s *Service) runNetworkBenchmark(ctx context.Context, result *BenchmarkResult, config BenchmarkConfig) error {
	s.logger.Info("Running Network benchmark",
		zap.String("target", result.Target),
		zap.Duration("duration", config.Duration))

	// Network benchmark simulates network operations
	// In production, this would connect to actual network endpoints
	var wg sync.WaitGroup
	ops := atomic.NewInt64(0)
	errors := atomic.NewInt64(0)
	bytes := atomic.NewInt64(0)
	latencies := make([]time.Duration, 0)
	latencyMu := sync.Mutex{}

	workerCount := config.Concurrency
	if workerCount <= 0 {
		workerCount = 1
	}

	workDuration := config.Duration / time.Duration(workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			endTime := time.Now().Add(workDuration)
			workerOps := int64(0)
			workerErrors := int64(0)
			workerBytes := int64(0)
			workerLatencies := make([]time.Duration, 0)

			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					// Simulate network operation
					packetSize := 1024
					processingTime := time.Duration(packetSize) * time.Nanosecond * 100

					time.Sleep(processingTime)

					latency := time.Since(start)
					workerLatencies = append(workerLatencies, latency)
					workerOps++
					workerBytes += int64(packetSize)
				}
			}

			ops.Add(workerOps)
			errors.Add(workerErrors)
			bytes.Add(workerBytes)

			latencyMu.Lock()
			latencies = append(latencies, workerLatencies...)
			latencyMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Update result metrics
	result.Metrics.TotalOps = ops.Load()
	result.Metrics.FailedOps = errors.Load()
	result.Metrics.SuccessfulOps = result.Metrics.TotalOps - result.Metrics.FailedOps
	result.Metrics.TotalDuration = config.Duration
	result.Metrics.Throughput = float64(result.Metrics.TotalOps) / config.Duration.Seconds()
	result.Metrics.BytesProcessed = bytes.Load()
	result.Metrics.BytesPerSec = float64(bytes.Load()) / config.Duration.Seconds()

	if len(latencies) > 0 {
		result.Metrics.AvgDuration = calculateAverage(latencies)
		result.Metrics.MinDuration = calculateMin(latencies)
		result.Metrics.MaxDuration = calculateMax(latencies)
		result.Metrics.LatencyP50 = calculatePercentile(latencies, 50)
		result.Metrics.LatencyP95 = calculatePercentile(latencies, 95)
		result.Metrics.LatencyP99 = calculatePercentile(latencies, 99)
	}

	if result.Metrics.TotalOps > 0 {
		result.Metrics.ErrorRate = float64(result.Metrics.FailedOps) / float64(result.Metrics.TotalOps) * 100
	}

	return nil
}

// runDiskBenchmark runs disk I/O benchmarks
func (s *Service) runDiskBenchmark(ctx context.Context, result *BenchmarkResult, config BenchmarkConfig) error {
	s.logger.Info("Running Disk benchmark",
		zap.String("target", result.Target),
		zap.Duration("duration", config.Duration))

	// Disk benchmark simulates disk operations
	var wg sync.WaitGroup
	ops := atomic.NewInt64(0)
	errors := atomic.NewInt64(0)
	bytes := atomic.NewInt64(0)
	latencies := make([]time.Duration, 0)
	latencyMu := sync.Mutex{}

	workerCount := config.Concurrency
	if workerCount <= 0 {
		workerCount = 1
	}

	workDuration := config.Duration / time.Duration(workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			endTime := time.Now().Add(workDuration)
			workerOps := int64(0)
			workerErrors := int64(0)
			workerBytes := int64(0)
			workerLatencies := make([]time.Duration, 0)

			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					// Simulate disk operation
					blockSize := 4096
					ioTime := time.Duration(blockSize) * time.Nanosecond * 50

					time.Sleep(ioTime)

					latency := time.Since(start)
					workerLatencies = append(workerLatencies, latency)
					workerOps++
					workerBytes += int64(blockSize)
				}
			}

			ops.Add(workerOps)
			errors.Add(workerErrors)
			bytes.Add(workerBytes)

			latencyMu.Lock()
			latencies = append(latencies, workerLatencies...)
			latencyMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Update result metrics
	result.Metrics.TotalOps = ops.Load()
	result.Metrics.FailedOps = errors.Load()
	result.Metrics.SuccessfulOps = result.Metrics.TotalOps - result.Metrics.FailedOps
	result.Metrics.TotalDuration = config.Duration
	result.Metrics.Throughput = float64(result.Metrics.TotalOps) / config.Duration.Seconds()
	result.Metrics.BytesProcessed = bytes.Load()
	result.Metrics.BytesPerSec = float64(bytes.Load()) / config.Duration.Seconds()

	if len(latencies) > 0 {
		result.Metrics.AvgDuration = calculateAverage(latencies)
		result.Metrics.MinDuration = calculateMin(latencies)
		result.Metrics.MaxDuration = calculateMax(latencies)
		result.Metrics.LatencyP50 = calculatePercentile(latencies, 50)
		result.Metrics.LatencyP95 = calculatePercentile(latencies, 95)
		result.Metrics.LatencyP99 = calculatePercentile(latencies, 99)
	}

	if result.Metrics.TotalOps > 0 {
		result.Metrics.ErrorRate = float64(result.Metrics.FailedOps) / float64(result.Metrics.TotalOps) * 100
	}

	return nil
}

// runCustomBenchmark runs a custom benchmark using provided workload function
func (s *Service) runCustomBenchmark(ctx context.Context, result *BenchmarkResult, config BenchmarkConfig) error {
	s.logger.Info("Running Custom benchmark",
		zap.String("target", result.Target),
		zap.Duration("duration", config.Duration))

	if config.CustomWork == nil {
		return fmt.Errorf("custom workload function not provided")
	}

	var wg sync.WaitGroup
	ops := atomic.NewInt64(0)
	errors := atomic.NewInt64(0)
	latencies := make([]time.Duration, 0)
	latencyMu := sync.Mutex{}

	workerCount := config.Concurrency
	if workerCount <= 0 {
		workerCount = 1
	}

	workDuration := config.Duration / time.Duration(workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			endTime := time.Now().Add(workDuration)
			workerOps := int64(0)
			workerErrors := int64(0)
			workerLatencies := make([]time.Duration, 0)

			iter := 0
			for time.Now().Before(endTime) {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					err := config.CustomWork(ctx, iter)
					if err != nil {
						workerErrors++
					} else {
						workerOps++
					}

					latency := time.Since(start)
					workerLatencies = append(workerLatencies, latency)
					iter++
				}
			}

			ops.Add(workerOps)
			errors.Add(workerErrors)

			latencyMu.Lock()
			latencies = append(latencies, workerLatencies...)
			latencyMu.Unlock()
		}(i)
	}

	wg.Wait()

	// Update result metrics
	result.Metrics.TotalOps = ops.Load() + errors.Load()
	result.Metrics.FailedOps = errors.Load()
	result.Metrics.SuccessfulOps = ops.Load()
	result.Metrics.TotalDuration = config.Duration
	result.Metrics.Throughput = float64(result.Metrics.TotalOps) / config.Duration.Seconds()

	if len(latencies) > 0 {
		result.Metrics.AvgDuration = calculateAverage(latencies)
		result.Metrics.MinDuration = calculateMin(latencies)
		result.Metrics.MaxDuration = calculateMax(latencies)
		result.Metrics.LatencyP50 = calculatePercentile(latencies, 50)
		result.Metrics.LatencyP95 = calculatePercentile(latencies, 95)
		result.Metrics.LatencyP99 = calculatePercentile(latencies, 99)
	}

	if result.Metrics.TotalOps > 0 {
		result.Metrics.ErrorRate = float64(result.Metrics.FailedOps) / float64(result.Metrics.TotalOps) * 100
	}

	return nil
}

// calculateStats calculates statistical measures for the benchmark results
func (s *Service) calculateStats(result *BenchmarkResult) {
	// Sample size is the number of operations
	result.Statistics.SampleSize = int(result.Metrics.TotalOps)

	// Calculate standard deviation
	if len(result.Metrics.LatencyP50.String()) > 0 {
		// Simplified calculation
		result.Statistics.StdDev = result.Metrics.MaxDuration - result.Metrics.MinDuration
		result.Statistics.Variance = float64(result.Statistics.StdDev)
	}

	// Outlier detection (simple IQR method)
	iqr := result.Metrics.LatencyP95 - result.Metrics.LatencyP50
	upperBound := result.Metrics.LatencyP95 + 1.5*iqr
	if result.Metrics.MaxDuration > upperBound {
		result.Statistics.Outliers = int(float64(result.Statistics.SampleSize) * 0.01) // Rough estimate
	}

	// Anomaly detection
	if result.Metrics.ErrorRate > 1.0 {
		result.Statistics.Anomalies = 1
	}
}

// Helper functions
func computePI(iterations int) float64 {
	inside := 0
	for i := 0; i < iterations; i++ {
		x := float64(i%10000) / 10000 * 2 - 1
		y := float64((i/10000)%10000) / 10000 * 2 - 1
		if x*x+y*y <= 1 {
			inside++
		}
	}
	return float64(inside) / float64(iterations) * 4
}

func calculateAverage(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var sum int64
	for _, l := range latencies {
		sum += l.Nanoseconds()
	}
	return time.Duration(sum / int64(len(latencies)))
}

func calculateMin(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	min := latencies[0]
	for _, l := range latencies {
		if l < min {
			min = l
		}
	}
	return min
}

func calculateMax(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	max := latencies[0]
	for _, l := range latencies {
		if l > max {
			max = l
		}
	}
	return max
}

func calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile
	index := int(float64(len(sorted)) * float64(percentile) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func generateID() string {
	return fmt.Sprintf("bench-%d", time.Now().UnixNano())
}

func extractConfig(cfg interface{}) *BenchmarkConfig {
	// Default configuration
	return &BenchmarkConfig{
		Type:        BenchmarkTypeAll,
		Duration:    60 * time.Second,
		Concurrency: 1,
		Iterations:  1000,
		Warmup:      10 * time.Second,
		Timeout:     5 * time.Minute,
	}
}

// BenchmarkBuilder provides a fluent API for building benchmarks
type BenchmarkBuilder struct {
	service   *Service
	target    string
	benchType BenchmarkType
	config    BenchmarkConfig
}

// WithType sets the benchmark type
func (b *BenchmarkBuilder) WithType(t BenchmarkType) *BenchmarkBuilder {
	b.benchType = t
	b.config.Type = t
	return b
}

// WithDuration sets the benchmark duration
func (b *BenchmarkBuilder) WithDuration(d time.Duration) *BenchmarkBuilder {
	b.config.Duration = d
	return b
}

// WithConcurrency sets the number of concurrent workers
func (b *BenchmarkBuilder) WithConcurrency(c int) *BenchmarkBuilder {
	b.config.Concurrency = c
	return b
}

// WithIterations sets the number of iterations
func (b *BenchmarkBuilder) WithIterations(i int) *BenchmarkBuilder {
	b.config.Iterations = i
	return b
}

// WithWarmup sets the warmup duration
func (b *BenchmarkBuilder) WithWarmup(w time.Duration) *BenchmarkBuilder {
	b.config.Warmup = w
	return b
}

// WithCustomWorkload sets a custom workload function
func (b *BenchmarkBuilder) WithCustomWorkload(w WorkloadFunction) *BenchmarkBuilder {
	b.config.CustomWork = w
	b.benchType = BenchmarkTypeCustom
	b.config.Type = BenchmarkTypeCustom
	return b
}

// Build builds the benchmark configuration
func (b *BenchmarkBuilder) Build() BenchmarkConfig {
	return b.config
}

// Run executes the benchmark
func (b *BenchmarkBuilder) Run(ctx context.Context) (*BenchmarkResult, error) {
	return b.service.Run(ctx, b.target, b.config)
}

// GetResult retrieves a benchmark result by ID
func (s *Service) GetResult(id string) *BenchmarkResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results[id]
}

// ListResults returns all benchmark results
func (s *Service) ListResults() []*BenchmarkResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*BenchmarkResult, 0, len(s.results))
	for _, r := range s.results {
		results = append(results, r)
	}
	return results
}

// DeleteResult removes a benchmark result
func (s *Service) DeleteResult(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.results, id)
}
