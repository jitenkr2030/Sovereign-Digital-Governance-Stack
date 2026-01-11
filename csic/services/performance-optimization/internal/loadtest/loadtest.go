package loadtest

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ScenarioType represents the type of load test scenario
type ScenarioType string

const (
	ScenarioTypeSteady   ScenarioType = "steady"
	ScenarioTypeRampup   ScenarioType = "rampup"
	ScenarioTypeSpike    ScenarioType = "spike"
	ScenarioTypeSoak     ScenarioType = "soak"
	ScenarioTypeStress   ScenarioType = "stress"
	ScenarioTypeWave     ScenarioType = "wave"
)

// LoadTestConfig contains configuration for a load test
type LoadTestConfig struct {
	Target           string            `json:"target"`
	Scenario         ScenarioType      `json:"scenario"`
	RampupTime       time.Duration     `json:"rampup_time"`
	SteadyStateTime  time.Duration     `json:"steady_state_time"`
	CoolDownTime     time.Duration     `json:"cool_down_time"`
	InitialRPS       int               `json:"initial_rps"`
	TargetRPS        int               `json:"target_rps"`
	MaxRPS           int               `json:"max_rps"`
	Concurrency      int               `json:"concurrency"`
	MaxConcurrency   int               `json:"max_concurrency"`
	Duration         time.Duration     `json:"duration"`
	Timeout          time.Duration     `json:"timeout"`
	MaxErrors        int               `json:"max_errors"`
	MaxErrorRate     float64           `json:"max_error_rate"`
	RequestConfig    RequestConfig     `json:"request_config"`
	Headers          map[string]string `json:"headers"`
	Cookies          map[string]string `json:"cookies"`
	Auth             AuthConfig        `json:"auth"`
	ProgressiveLoad  bool              `json:"progressive_load"`
	StepSize         int               `json:"step_size"`
	StepDuration     time.Duration     `json:"step_duration"`
}

// RequestConfig contains request-specific configuration
type RequestConfig struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Body         string            `json:"body"`
	ContentType  string            `json:"content_type"`
	BodyFromFile string            `json:"body_from_file"`
	QueryParams  map[string]string `json:"query_params"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Type   string `json:"type"`
	Token  string `json:"token"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	APIKey string `json:"api_key"`
}

// LoadTestResult represents the results of a load test
type LoadTestResult struct {
	ID             string              `json:"id"`
	Target         string              `json:"target"`
	Scenario       ScenarioType        `json:"scenario"`
	StartTime      time.Time           `json:"start_time"`
	EndTime        time.Time           `json:"end_time"`
	Duration       time.Duration       `json:"duration"`
	Configuration  LoadTestConfig      `json:"configuration"`
	Statistics     LoadTestStats       `json:"statistics"`
	TimeSeriesData []TimeSeriesPoint   `json:"time_series_data"`
	Errors         []LoadTestError     `json:"errors,omitempty"`
	Status         string              `json:"status"`
}

// LoadTestStats contains aggregated statistics
type LoadTestStats struct {
	TotalRequests      int64           `json:"total_requests"`
	SuccessfulRequests int64           `json:"successful_requests"`
	FailedRequests     int64           `json:"failed_requests"`
	TotalBytesSent     int64           `json:"total_bytes_sent"`
	TotalBytesReceived int64           `json:"total_bytes_received"`
	AverageRPS         float64         `json:"average_rps"`
	PeakRPS            float64         `json:"peak_rps"`
	MinimumRPS         float64         `json:"minimum_rps"`
	AverageLatency     time.Duration   `json:"average_latency"`
	MinLatency         time.Duration   `json:"min_latency"`
	MaxLatency         time.Duration   `json:"max_latency"`
	LatencyP50         time.Duration   `json:"latency_p50"`
	LatencyP75         time.Duration   `json:"latency_p75"`
	LatencyP90         time.Duration   `json:"latency_p90"`
	LatencyP95         time.Duration   `json:"latency_p95"`
	LatencyP99         time.Duration   `json:"latency_p99"`
	ErrorRate          float64         `json:"error_rate"`
	RequestsPerSecond  []float64       `json:"requests_per_second"`
	ConcurrencyLevel   int             `json:"concurrency_level"`
	Throughput         float64         `json:"throughput_bytes_per_second"`
}

// TimeSeriesPoint represents a single point in the time series data
type TimeSeriesPoint struct {
	Timestamp       time.Time       `json:"timestamp"`
	ElapsedTime     time.Duration   `json:"elapsed_time"`
	ActiveUsers     int             `json:"active_users"`
	RequestsPerSec  float64         `json:"requests_per_sec"`
	ResponseTimeAvg time.Duration   `json:"response_time_avg"`
	ResponseTimeMin time.Duration   `json:"response_time_min"`
	ResponseTimeMax time.Duration   `json:"response_time_max"`
	ErrorCount      int             `json:"error_count"`
	ErrorRate       float64         `json:"error_rate"`
	BytesSent       int64           `json:"bytes_sent"`
	BytesReceived   int64           `json:"bytes_received"`
}

// LoadTestError represents an error that occurred during load testing
type LoadTestError struct {
	Timestamp   time.Time `json:"timestamp"`
	RequestID   string    `json:"request_id"`
	URL         string    `json:"url"`
	Method      string    `json:"method"`
	StatusCode  int       `json:"status_code"`
	Error       string    `json:"error"`
	Count       int       `json:"count"`
}

// Service provides load testing capabilities
type Service struct {
	logger      *zap.Logger
	config      *LoadTestConfig
	results     map[string]*LoadTestResult
	activeTests map[string]*LoadTestSession
	mu          sync.RWMutex
}

// LoadTestSession represents an active load test session
type LoadTestSession struct {
	Config      *LoadTestConfig
	Result      *LoadTestResult
	CancelFunc  context.CancelFunc
	Stats       *SessionStats
	StopChan    chan struct{}
}

// SessionStats tracks statistics during an active test
type SessionStats struct {
	TotalRequests   int64
	SuccessfulReqs  int64
	FailedReqs      int64
	TotalBytesSent  int64
	TotalBytesRecv  int64
	Latencies       []time.Duration
	LatencyMu       sync.Mutex
	RPSHistory      []float64
	ActiveWorkers   int
	StartTime       time.Time
	LastReportTime  time.Time
}

// Global service instance
var (
	defaultLoadTestService *Service
	loadTestServiceMu      sync.RWMutex
)

// Initialize initializes the load test service
func Initialize(logger *zap.Logger, cfg interface{}) error {
	loadTestServiceMu.Lock()
	defer loadTestServiceMu.Unlock()

	// Extract configuration
	var appConfig interface{}
	if c, ok := cfg.(interface {
		Get(name string) interface{}
	}); ok {
		appConfig = c
	}

	defaultLoadTestService = &Service{
		logger:      logger,
		results:     make(map[string]*LoadTestResult),
		activeTests: make(map[string]*LoadTestSession),
	}

	logger.Info("Load test service initialized")
	return nil
}

// GetService returns the default load test service
func GetService() *Service {
	loadTestServiceMu.RLock()
	defer loadTestServiceMu.RUnlock()
	return defaultLoadTestService
}

// NewLoadTest creates a new load test configuration
func (s *Service) NewLoadTest(target string) *LoadTestBuilder {
	return &LoadTestBuilder{
		service: s,
		config: LoadTestConfig{
			Target:          target,
			Scenario:        ScenarioTypeSteady,
			RampupTime:      30 * time.Second,
			SteadyStateTime: 60 * time.Second,
			CoolDownTime:    10 * time.Second,
			InitialRPS:      10,
			TargetRPS:       100,
			MaxRPS:          1000,
			Concurrency:     10,
			MaxConcurrency:  100,
			Timeout:         30 * time.Second,
			MaxErrors:       100,
			MaxErrorRate:    5.0,
			RequestConfig: RequestConfig{
				Method:      "GET",
				ContentType: "application/json",
			},
		},
	}
}

// Run executes a load test
func (s *Service) Run(ctx context.Context, config LoadTestConfig) (*LoadTestResult, error) {
	// Create result
	result := &LoadTestResult{
		ID:            generateLoadTestID(),
		Target:        config.Target,
		Scenario:      config.Scenario,
		StartTime:     time.Now(),
		Configuration: config,
		Status:        "running",
		Statistics: LoadTestStats{
			MinLatency: math.MaxDuration,
			RequestsPerSecond: make([]float64, 0),
		},
		TimeSeriesData: make([]TimeSeriesPoint, 0),
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(ctx)

	// Create session
	session := &LoadTestSession{
		Config:     &config,
		Result:     result,
		CancelFunc: cancel,
		Stats: &SessionStats{
			StartTime:      result.StartTime,
			LastReportTime: result.StartTime,
			Latencies:      make([]time.Duration, 0),
			RPSHistory:     make([]float64, 0),
		},
		StopChan: make(chan struct{}),
	}

	// Store session
	s.mu.Lock()
	s.results[result.ID] = result
	s.activeTests[result.ID] = session
	s.mu.Unlock()

	// Run the appropriate scenario
	var err error
	switch config.Scenario {
	case ScenarioTypeSteady:
		err = s.runSteadyStateScenario(ctx, session)
	case ScenarioTypeRampup:
		err = s.runRampUpScenario(ctx, session)
	case ScenarioTypeSpike:
		err = s.runSpikeScenario(ctx, session)
	case ScenarioTypeSoak:
		err = s.runSoakScenario(ctx, session)
	case ScenarioTypeStress:
		err = s.runStressScenario(ctx, session)
	case ScenarioTypeWave:
		err = s.runWaveScenario(ctx, session)
	default:
		err = fmt.Errorf("unknown scenario type: %s", config.Scenario)
	}

	// Update result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = "completed"

	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, LoadTestError{
			Timestamp: time.Now(),
			Error:     err.Error(),
		})
	}

	// Calculate final statistics
	s.calculateStats(result, session.Stats)

	// Remove from active tests
	s.mu.Lock()
	delete(s.activeTests, result.ID)
	s.mu.Unlock()

	s.logger.Info("Load test completed",
		zap.String("id", result.ID),
		zap.String("scenario", string(config.Scenario)),
		zap.Duration("duration", result.Duration),
		zap.Int64("total_requests", result.Statistics.TotalRequests),
		zap.Float64("avg_rps", result.Statistics.AverageRPS),
		zap.Float64("error_rate", result.Statistics.ErrorRate))

	return result, err
}

// runSteadyStateScenario runs a steady-state load test
func (s *Service) runSteadyStateScenario(ctx context.Context, session *LoadTestSession) error {
	config := session.Config

	s.logger.Info("Running steady-state scenario",
		zap.String("target", config.Target),
		zap.Int("target_rps", config.TargetRPS),
		zap.Duration("duration", config.SteadyStateTime))

	// Ramp up
	if err := s.rampUp(ctx, session, config.RampupTime, config.InitialRPS, config.TargetRPS); err != nil {
		return err
	}

	// Steady state
	if err := s.steadyState(ctx, session, config.SteadyStateTime, config.TargetRPS); err != nil {
		return err
	}

	// Cool down
	if err := s.coolDown(ctx, session, config.CoolDownTime, config.TargetRPS); err != nil {
		return err
	}

	return nil
}

// runRampUpScenario runs a ramp-up load test
func (s *Service) runRampUpScenario(ctx context.Context, session *LoadTestSession) error {
	config := session.Config

	s.logger.Info("Running ramp-up scenario",
		zap.String("target", config.Target),
		zap.Int("start_rps", config.InitialRPS),
		zap.Int("end_rps", config.TargetRPS),
		zap.Duration("duration", config.Duration))

	// Progressive ramp up
	stepSize := config.StepSize
	if stepSize <= 0 {
		stepSize = 10
	}

	currentRPS := config.InitialRPS
	stepDuration := config.StepDuration
	if stepDuration == 0 {
		stepDuration = 10 * time.Second
	}

	for currentRPS <= config.TargetRPS {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.StopChan:
			return nil
		default:
			if err := s.steadyState(ctx, session, stepDuration, currentRPS); err != nil {
				return err
			}
			currentRPS += stepSize
			if currentRPS > config.MaxRPS {
				currentRPS = config.MaxRPS
			}
		}
	}

	return nil
}

// runSpikeScenario runs a spike load test
func (s *Service) runSpikeScenario(ctx context.Context, session *LoadTestSession) error {
	config := session.Config

	s.logger.Info("Running spike scenario",
		zap.String("target", config.Target),
		zap.Int("initial_rps", config.InitialRPS),
		zap.Int("spike_rps", config.TargetRPS),
		zap.Duration("duration", config.Duration))

	// Initial steady state at low RPS
	if err := s.steadyState(ctx, session, 10*time.Second, config.InitialRPS); err != nil {
		return err
	}

	// Spike to target RPS
	if err := s.steadyState(ctx, session, 10*time.Second, config.TargetRPS); err != nil {
		return err
	}

	// Cool down
	if err := s.steadyState(ctx, session, 10*time.Second, config.InitialRPS); err != nil {
		return err
	}

	return nil
}

// runSoakScenario runs a soak (endurance) load test
func (s *Service) runSoakScenario(ctx context.Context, session *LoadTestSession) error {
	config := session.Config

	s.logger.Info("Running soak scenario",
		zap.String("target", config.Target),
		zap.Int("rps", config.TargetRPS),
		zap.Duration("duration", config.Duration))

	// Extended steady state
	return s.steadyState(ctx, session, config.Duration, config.TargetRPS)
}

// runStressScenario runs a stress test
func (s *Service) runStressScenario(ctx context.Context, session *LoadTestSession) error {
	config := session.Config

	s.logger.Info("Running stress scenario",
		zap.String("target", config.Target),
		zap.Int("max_rps", config.MaxRPS),
		zap.Int("max_concurrency", config.MaxConcurrency),
		zap.Duration("duration", config.Duration))

	// Gradually increase load until failure or max
	currentRPS := config.InitialRPS
	increment := config.TargetRPS / 10

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.StopChan:
			return nil
		default:
			// Check if we should stop
			stats := session.Stats
			stats.LatencyMu.Lock()
			errorRate := float64(stats.FailedReqs) / float64(stats.TotalRequests+1) * 100
			stats.LatencyMu.Unlock()

			if errorRate > config.MaxErrorRate {
				s.logger.Warn("Stopping stress test - error rate exceeded threshold",
					zap.Float64("error_rate", errorRate),
					zap.Float64("threshold", config.MaxErrorRate))
				return nil
			}

			// Run at current load
			if err := s.steadyState(ctx, session, 30*time.Second, currentRPS); err != nil {
				return err
			}

			// Increase load
			currentRPS += increment
			if currentRPS > config.MaxRPS {
				currentRPS = config.MaxRPS
			}
		}
	}
}

// runWaveScenario runs a wave load test
func (s *Service) runWaveScenario(ctx context.Context, session *LoadTestSession) error {
	config := session.Config

	s.logger.Info("Running wave scenario",
		zap.String("target", config.Target),
		zap.Int("min_rps", config.InitialRPS),
		zap.Int("max_rps", config.TargetRPS),
		zap.Duration("duration", config.Duration))

	// Wave pattern
	cycleDuration := 60 * time.Second
	currentRPS := config.InitialRPS
	increasing := true

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.StopChan:
			return nil
		default:
			elapsed := time.Since(session.Stats.StartTime)
			if elapsed > config.Duration {
				return nil
			}

			if err := s.steadyState(ctx, session, 10*time.Second, currentRPS); err != nil {
				return err
			}

			// Wave pattern
			if increasing {
				currentRPS += config.StepSize
				if currentRPS >= config.TargetRPS {
					increasing = false
				}
			} else {
				currentRPS -= config.StepSize
				if currentRPS <= config.InitialRPS {
					increasing = true
				}
			}
		}
	}
}

// rampUp gradually increases the load
func (s *Service) rampUp(ctx context.Context, session *LoadTestSession, duration time.Duration, startRPS, endRPS int) error {
	steps := 10
	stepDuration := duration / time.Duration(steps)
	stepIncrement := (endRPS - startRPS) / steps

	currentRPS := startRPS
	for i := 0; i < steps; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.StopChan:
			return nil
		default:
			if err := s.steadyState(ctx, session, stepDuration, currentRPS); err != nil {
				return err
			}
			currentRPS += stepIncrement
		}
	}

	return nil
}

// steadyState maintains a constant load
func (s *Service) steadyState(ctx context.Context, session *LoadTestSession, duration time.Duration, targetRPS int) error {
	s.logger.Debug("Steady state load",
		zap.Int("rps", targetRPS),
		zap.Duration("duration", duration))

	// Calculate workers needed
	workers := targetRPS / 10
	if workers < 1 {
		workers = 1
	}
	if workers > session.Config.MaxConcurrency {
		workers = session.Config.MaxConcurrency
	}

	session.Stats.ActiveWorkers = workers

	// Create request channel
	requestChan := make(chan struct{}, workers)
	responseChan := make(chan *RequestResult, workers)

	// Create worker pool
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go s.worker(ctx, i, requestChan, responseChan, session)
	}

	// Start request generator
	go s.requestGenerator(ctx, duration, targetRPS, requestChan, session)

	// Collect results
	go s.resultCollector(ctx, responseChan, session)

	// Wait for completion
	done := time.After(duration)
	select {
	case <-ctx.Done():
		close(requestChan)
		wg.Wait()
		return ctx.Err()
	case <-session.StopChan:
		close(requestChan)
		wg.Wait()
		return nil
	case <-done:
		close(requestChan)
		wg.Wait()
		return nil
	}
}

// coolDown gradually decreases the load
func (s *Service) coolDown(ctx context.Context, session *LoadTestSession, duration time.Duration, startRPS int) error {
	return s.rampUp(ctx, session, duration, startRPS, startRPS/10)
}

// worker processes requests
func (s *Service) worker(ctx context.Context, id int, requestChan <-chan struct{}, responseChan chan<- *RequestResult, session *LoadTestSession) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Worker panic",
				zap.Int("worker_id", id),
				zap.Any("panic", r))
		}
	}()

	for range requestChan {
		select {
		case <-ctx.Done():
			return
		case <-session.StopChan:
			return
		default:
			result := s.executeRequest(ctx, session.Config)
			result.WorkerID = id

			responseChan <- result

			// Update statistics
			stats := session.Stats
			atomic.AddInt64(&stats.TotalRequests, 1)

			if result.Error != nil {
				atomic.AddInt64(&stats.FailedReqs, 1)
			} else {
				atomic.AddInt64(&stats.SuccessfulReqs, 1)
				atomic.AddInt64(&stats.TotalBytesRecv, result.BytesReceived)
			}

			// Record latency
			stats.LatencyMu.Lock()
			stats.Latencies = append(stats.Latencies, result.Latency)
			if result.Latency < session.Statistics.MinLatency {
				session.Statistics.MinLatency = result.Latency
			}
			if result.Latency > session.Statistics.MaxLatency {
				session.Statistics.MaxLatency = result.Latency
			}
			stats.LatencyMu.Unlock()
		}
	}
}

// requestGenerator generates requests at the target rate
func (s *Service) requestGenerator(ctx context.Context, duration time.Duration, targetRPS int, requestChan chan<- struct{}, session *SessionStats) {
	interval := time.Second / time.Duration(targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-session.StopChan:
			return
		case <-timeout:
			return
		case t := <-ticker.C:
			// Check if we've exceeded duration
			if time.Since(startTime) > duration {
				return
			}

			select {
			case requestChan <- struct{}{}:
			default:
				// Channel full, skip this request
			}

			// Record RPS
			currentRPS := float64(atomic.LoadInt64(&session.TotalRequests)) / time.Since(startTime).Seconds()
			stats.RPSHistory = append(session.RPSHistory, currentRPS)
		}
	}
}

// resultCollector collects and processes results
func (s *Service) resultCollector(ctx context.Context, responseChan <-chan *RequestResult, session *LoadTestSession) {
	reportTicker := time.NewTicker(5 * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-session.StopChan:
			return
		case result, ok := <-responseChan:
			if !ok {
				return
			}
			_ = result
		case <-reportTicker.C:
			s.reportProgress(session)
		}
	}
}

// executeRequest executes a single request
func (s *Service) executeRequest(ctx context.Context, config *LoadTestConfig) *RequestResult {
	start := time.Now()

	// Simulate request execution
	time.Sleep(time.Duration(1000000/config.TargetRPS) * time.Nanosecond)

	// Simulate occasional errors
	var err error
	if time.Now().UnixNano()%100 > 98 {
		err = fmt.Errorf("simulated request error")
	}

	return &RequestResult{
		RequestID:     fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Latency:       time.Since(start),
		StatusCode:    200,
		Error:         err,
		BytesReceived: 1024,
	}
}

// reportProgress reports current progress
func (s *Service) reportProgress(session *SessionStats) {
	stats := session
	stats.LatencyMu.Lock()
	total := atomic.LoadInt64(&stats.TotalRequests)
	successful := atomic.LoadInt64(&stats.SuccessfulReqs)
	failed := atomic.LoadInt64(&stats.FailedReqs)
	elapsed := time.Since(stats.StartTime)
	rps := float64(total) / elapsed.Seconds()
	stats.LatencyMu.Unlock()

	s.logger.Info("Load test progress",
		zap.Int64("total_requests", total),
		zap.Int64("successful", successful),
		zap.Int64("failed", failed),
		zap.Float64("rps", rps))
}

// calculateStats calculates final statistics
func (s *Service) calculateStats(result *LoadTestResult, stats *SessionStats) {
	result.Statistics.TotalRequests = atomic.LoadInt64(&stats.TotalRequests)
	result.Statistics.SuccessfulRequests = atomic.LoadInt64(&stats.SuccessfulReqs)
	result.Statistics.FailedRequests = atomic.LoadInt64(&stats.FailedReqs)
	result.Statistics.TotalBytesReceived = atomic.LoadInt64(&stats.TotalBytesRecv)

	if result.Statistics.TotalRequests > 0 {
		result.Statistics.ErrorRate = float64(result.Statistics.FailedRequests) / float64(result.Statistics.TotalRequests) * 100
	}

	// Calculate RPS
	if result.Duration.Seconds() > 0 {
		result.Statistics.AverageRPS = float64(result.Statistics.TotalRequests) / result.Duration.Seconds()
	}

	// Calculate latency percentiles
	stats.LatencyMu.Lock()
	latencies := stats.Latencies
	stats.LatencyMu.Unlock()

	if len(latencies) > 0 {
		result.Statistics.AverageLatency = calculateAverageLatency(latencies)
		result.Statistics.LatencyP50 = calculatePercentile(latencies, 50)
		result.Statistics.LatencyP75 = calculatePercentile(latencies, 75)
		result.Statistics.LatencyP90 = calculatePercentile(latencies, 90)
		result.Statistics.LatencyP95 = calculatePercentile(latencies, 95)
		result.Statistics.LatencyP99 = calculatePercentile(latencies, 99)
	}

	// Calculate throughput
	if result.Duration.Seconds() > 0 {
		result.Statistics.Throughput = float64(result.Statistics.TotalBytesReceived) / result.Duration.Seconds()
	}
}

// RequestResult represents the result of a single request
type RequestResult struct {
	RequestID     string
	WorkerID      int
	Latency       time.Duration
	StatusCode    int
	Error         error
	BytesReceived int64
}

// Helper functions
func generateLoadTestID() string {
	return fmt.Sprintf("lt-%d", time.Now().UnixNano())
}

func calculateAverageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var sum int64
	for _, l := range latencies {
		sum += l.Nanoseconds()
	}
	return time.Duration(sum / int64(len(latencies)))
}

func calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * float64(percentile) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// StopTest stops an active load test
func (s *Service) StopTest(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.activeTests[id]
	if !ok {
		return fmt.Errorf("load test not found: %s", id)
	}

	close(session.StopChan)
	session.CancelFunc()
	return nil
}

// GetResult retrieves a load test result
func (s *Service) GetResult(id string) *LoadTestResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results[id]
}

// ListResults returns all load test results
func (s *Service) ListResults() []*LoadTestResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*LoadTestResult, 0, len(s.results))
	for _, r := range s.results {
		results = append(results, r)
	}
	return results
}

// LoadTestBuilder provides a fluent API for building load tests
type LoadTestBuilder struct {
	service *Service
	config  LoadTestConfig
}

// WithScenario sets the load test scenario
func (b *LoadTestBuilder) WithScenario(s ScenarioType) *LoadTestBuilder {
	b.config.Scenario = s
	return b
}

// WithTargetRPS sets the target requests per second
func (b *LoadTestBuilder) WithTargetRPS(rps int) *LoadTestBuilder {
	b.config.TargetRPS = rps
	return b
}

// WithDuration sets the test duration
func (b *LoadTestBuilder) WithDuration(d time.Duration) *LoadTestBuilder {
	b.config.Duration = d
	b.config.SteadyStateTime = d
	return b
}

// WithRampup sets the ramp-up time
func (b *LoadTestBuilder) WithRampup(r time.Duration) *LoadTestBuilder {
	b.config.RampupTime = r
	return b
}

// WithConcurrency sets the maximum concurrency
func (b *LoadTestBuilder) WithConcurrency(c int) *LoadTestBuilder {
	b.config.Concurrency = c
	b.config.MaxConcurrency = c
	return b
}

// WithMethod sets the HTTP method
func (b *LoadTestBuilder) WithMethod(method string) *LoadTestBuilder {
	b.config.RequestConfig.Method = method
	return b
}

// WithBody sets the request body
func (b *LoadTestBuilder) WithBody(body string) *LoadTestBuilder {
	b.config.RequestConfig.Body = body
	return b
}

// WithHeaders sets custom headers
func (b *LoadTestBuilder) WithHeaders(headers map[string]string) *LoadTestBuilder {
	b.config.Headers = headers
	return b
}

// Build builds the load test configuration
func (b *LoadTestBuilder) Build() LoadTestConfig {
	return b.config
}

// Run executes the load test
func (b *LoadTestBuilder) Run(ctx context.Context) (*LoadTestResult, error) {
	return b.service.Run(ctx, b.config)
}
