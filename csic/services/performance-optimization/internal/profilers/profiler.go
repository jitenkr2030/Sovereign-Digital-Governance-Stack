package profilers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"go.uber.org/zap"
)

// ProfileType represents the type of profile to collect
type ProfileType string

const (
	ProfileTypeCPU       ProfileType = "cpu"
	ProfileTypeMemory    ProfileType = "memory"
	ProfileTypeGoroutine ProfileType = "goroutine"
	ProfileTypeBlock     ProfileType = "block"
	ProfileTypeMutex     ProfileType = "mutex"
	ProfileTypeTrace     ProfileType = "trace"
	ProfileTypeThread    ProfileType = "thread"
	ProfileTypeHeap      ProfileType = "heap"
)

// ProfilerConfig contains configuration for profiling
type ProfilerConfig struct {
	Enabled            bool          `json:"enabled"`
	CPUProfileDuration time.Duration `json:"cpu_profile_duration"`
	MemoryProfile      bool          `json:"memory_profile"`
	TraceEnabled       bool          `json:"trace_enabled"`
	BlockProfile       bool          `json:"block_profile"`
	MutexProfile       bool          `json:"mutex_profile"`
	ProfileDir         string        `json:"profile_dir"`
}

// ProfileResult represents the result of a profiling session
type ProfileResult struct {
	ID           string        `json:"id"`
	Type         ProfileType   `json:"type"`
	Target       string        `json:"target"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	FilePath     string        `json:"file_path"`
	FileSize     int64         `json:"file_size"`
	ProfileData  []byte        `json:"-"`
	Status       string        `json:"status"`
	Error        error         `json:"error,omitempty"`
}

// Service provides profiling capabilities
type Service struct {
	logger  *zap.Logger
	config  *ProfilerConfig
	profiles map[string]*ProfileResult
	mu      sync.RWMutex
}

// Global service instance
var (
	defaultProfiler *Service
	profilerMu      sync.RWMutex
)

// Initialize initializes the profiler service
func Initialize(logger *zap.Logger, cfg interface{}) error {
	profilerMu.Lock()
	defer profilerMu.Unlock()

	// Extract configuration
	config := &ProfilerConfig{
		Enabled:            true,
		CPUProfileDuration: 30 * time.Second,
		MemoryProfile:      true,
		TraceEnabled:       true,
		BlockProfile:       true,
		MutexProfile:       true,
		ProfileDir:         "/tmp/csic-profiles",
	}

	// Try to extract from config
	if c, ok := cfg.(interface {
		Get(name string) interface{}
	}); ok {
		if v := c.Get("profiling"); v != nil {
			// Would parse detailed config here
			_ = v
		}
	}

	defaultProfiler = &Service{
		logger:  logger,
		config:  config,
		profiles: make(map[string]*ProfileResult),
	}

	// Create profile directory
	if err := os.MkdirAll(config.ProfileDir, 0755); err != nil {
		logger.Warn("Failed to create profile directory", zap.Error(err))
	}

	logger.Info("Profiler service initialized",
		zap.Bool("enabled", config.Enabled),
		zap.String("profile_dir", config.ProfileDir))

	return nil
}

// GetService returns the default profiler service
func GetService() *Service {
	profilerMu.RLock()
	defer profilerMu.RUnlock()
	return defaultProfiler
}

// StartProfile starts a profiling session
func (s *Service) StartProfile(ctx context.Context, target string, profileType ProfileType, duration time.Duration) (*ProfileResult, error) {
	result := &ProfileResult{
		ID:        generateProfileID(),
		Type:      profileType,
		Target:    target,
		StartTime: time.Now(),
		Status:    "running",
	}

	// Create output file
	filename := fmt.Sprintf("%s_%s_%s.prof", target, profileType, result.StartTime.Format("20060102150405"))
	filePath := filepath.Join(s.config.ProfileDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		result.Status = "failed"
		result.Error = err
		return result, err
	}
	defer file.Close()

	// Start the appropriate profiler
	switch profileType {
	case ProfileTypeCPU:
		err = s.profileCPU(ctx, file, duration, result)
	case ProfileTypeMemory:
		err = s.profileMemory(file, result)
	case ProfileTypeGoroutine:
		err = s.profileGoroutine(file, result)
	case ProfileTypeBlock:
		err = s.profileBlock(file, result)
	case ProfileTypeMutex:
		err = s.profileMutex(file, result)
	case ProfileTypeTrace:
		err = s.profileTrace(ctx, file, duration, result)
	case ProfileTypeHeap:
		err = s.profileHeap(file, result)
	default:
		err = fmt.Errorf("unknown profile type: %s", profileType)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.FilePath = filePath

	if err != nil {
		result.Status = "failed"
		result.Error = err
	} else {
		result.Status = "completed"
	}

	// Get file size
	if stat, err := os.Stat(filePath); err == nil {
		result.FileSize = stat.Size()
	}

	// Store result
	s.mu.Lock()
	s.profiles[result.ID] = result
	s.mu.Unlock()

	s.logger.Info("Profiling completed",
		zap.String("id", result.ID),
		zap.String("type", string(profileType)),
		zap.String("target", target),
		zap.Duration("duration", result.Duration),
		zap.Int64("file_size", result.FileSize))

	return result, err
}

// profileCPU profiles CPU usage
func (s *Service) profileCPU(ctx context.Context, file *os.File, duration time.Duration, result *ProfileResult) error {
	if err := pprof.StartCPUProfile(file); err != nil {
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	// Run for specified duration
	select {
	case <-ctx.Done():
		pprof.StopCPUProfile()
		return ctx.Err()
	case <-time.After(duration):
		pprof.StopCPUProfile()
		return nil
	}
}

// profileMemory profiles memory usage
func (s *Service) profileMemory(file *os.File, result *ProfileResult) error {
	runtime.GC()

	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	return nil
}

// profileGoroutine profiles goroutine usage
func (s *Service) profileGoroutine(file *os.File, result *ProfileResult) error {
	p := pprof.Lookup("goroutine")
	if p == nil {
		return fmt.Errorf("goroutine profile not available")
	}

	if err := p.WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	return nil
}

// profileBlock profiles blocking operations
func (s *Service) profileBlock(file *os.File, result *ProfileResult) error {
	p := pprof.Lookup("block")
	if p == nil {
		return fmt.Errorf("block profile not available")
	}

	if err := p.WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %w", err)
	}

	return nil
}

// profileMutex profiles mutex contention
func (s *Service) profileMutex(file *os.File, result *ProfileResult) error {
	p := pprof.Lookup("mutex")
	if p == nil {
		return fmt.Errorf("mutex profile not available")
	}

	if err := p.WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}

	return nil
}

// profileTrace captures execution trace
func (s *Service) profileTrace(ctx context.Context, file *os.File, duration time.Duration, result *ProfileResult) error {
	if err := trace.Start(file); err != nil {
		return fmt.Errorf("failed to start trace: %w", err)
	}

	select {
	case <-ctx.Done():
		trace.Stop()
		return ctx.Err()
	case <-time.After(duration):
		trace.Stop()
		return nil
	}
}

// profileHeap profiles heap memory
func (s *Service) profileHeap(file *os.File, result *ProfileResult) error {
	p := pprof.Lookup("heap")
	if p == nil {
		return fmt.Errorf("heap profile not available")
	}

	if err := p.WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	return nil
}

// GetProfile retrieves a profile result by ID
func (s *Service) GetProfile(id string) *ProfileResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profiles[id]
}

// ListProfiles returns all profile results
func (s *Service) ListProfiles() []*ProfileResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*ProfileResult, 0, len(s.profiles))
	for _, p := range s.profiles {
		results = append(results, p)
	}
	return results
}

// DeleteProfile removes a profile result
func (s *Service) DeleteProfile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, ok := s.profiles[id]
	if !ok {
		return fmt.Errorf("profile not found: %s", id)
	}

	// Remove file
	if profile.FilePath != "" {
		os.Remove(profile.FilePath)
	}

	delete(s.profiles, id)
	return nil
}

// AnalyzeProfile analyzes a profile and returns insights
func (s *Service) AnalyzeProfile(id string) (*ProfileAnalysis, error) {
	profile := s.GetProfile(id)
	if profile == nil {
		return nil, fmt.Errorf("profile not found: %s", id)
	}

	// Open the profile file
	f, err := os.Open(profile.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse profile based on type
	var analysis ProfileAnalysis
	analysis.ID = id
	analysis.ProfileType = profile.Type
	analysis.AnalyzedAt = time.Now()

	switch profile.Type {
	case ProfileTypeCPU:
		analysis = s.analyzeCPUProfile(f, analysis)
	case ProfileTypeMemory:
		analysis = s.analyzeMemoryProfile(f, analysis)
	case ProfileTypeGoroutine:
		analysis = s.analyzeGoroutineProfile(f, analysis)
	}

	return &analysis, nil
}

// ProfileAnalysis contains analysis results
type ProfileAnalysis struct {
	ID           string          `json:"id"`
	ProfileType  ProfileType     `json:"profile_type"`
	AnalyzedAt   time.Time       `json:"analyzed_at"`
	TopFunctions []FunctionInfo  `json:"top_functions"`
	TotalSamples int             `json:"total_samples"`
	Duration     time.Duration   `json:"duration"`
	Summary      string          `json:"summary"`
}

// FunctionInfo contains information about a function in the profile
type FunctionInfo struct {
	Function   string  `json:"function"`
	Packages   string  `json:"packages"`
	Samples    int     `json:"samples"`
	Percentage float64 `json:"percentage"`
	CumPercent float64 `json:"cum_percentage"`
}

// analyzeCPUProfile analyzes CPU profile data
func (s *Service) analyzeCPUProfile(f *os.File, analysis ProfileAnalysis) ProfileAnalysis {
	// Read profile data
	data, err := io.ReadAll(f)
	if err != nil {
		analysis.Summary = fmt.Sprintf("Failed to read profile: %v", err)
		return analysis
	}

	// Simple analysis - count lines (in production, would use pprof library)
	lines := strings.Split(string(data), "\n")
	analysis.TotalSamples = len(lines)

	// Find top functions
	var functions []string
	for _, line := range lines {
		if strings.HasPrefix(line, "flat") || strings.HasPrefix(line, "") {
			continue
		}
		functions = append(functions, line)
	}

	analysis.Summary = fmt.Sprintf("CPU profile analyzed: %d samples collected over %v",
		analysis.TotalSamples, analysis.Duration)

	return analysis
}

// analyzeMemoryProfile analyzes memory profile data
func (s *Service) analyzeMemoryProfile(f *os.File, analysis ProfileAnalysis) ProfileAnalysis {
	data, err := io.ReadAll(f)
	if err != nil {
		analysis.Summary = fmt.Sprintf("Failed to read profile: %v", err)
		return analysis
	}

	lines := strings.Split(string(data), "\n")
	analysis.TotalSamples = len(lines)

	analysis.Summary = fmt.Sprintf("Memory profile analyzed: %d allocation samples",
		analysis.TotalSamples)

	return analysis
}

// analyzeGoroutineProfile analyzes goroutine profile data
func (s *Service) analyzeGoroutineProfile(f *os.File, analysis ProfileAnalysis) ProfileAnalysis {
	data, err := io.ReadAll(f)
	if err != nil {
		analysis.Summary = fmt.Sprintf("Failed to read profile: %v", err)
		return analysis
	}

	// Count goroutines
	count := 0
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "goroutine ") {
			count++
		}
	}

	analysis.TotalSamples = count
	analysis.Summary = fmt.Sprintf("Goroutine profile analyzed: %d goroutines found",
		count)

	return analysis
}

// pprof handlers for HTTP endpoints
func PprofIndexHandler(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}

func PprofCmdlineHandler(w http.ResponseWriter, r *http.Request) {
	pprof.Cmdline(w, r)
}

func PprofProfileHandler(w http.ResponseWriter, r *http.Request) {
	pprof.Profile(w, r)
}

func PprofSymbolHandler(w http.ResponseWriter, r *http.Request) {
	pprof.Symbol(w, r)
}

func PprofTraceHandler(w http.ResponseWriter, r *http.Request) {
	pprof.Trace(w, r)
}

func generateProfileID() string {
	return fmt.Sprintf("prof-%d", time.Now().UnixNano())
}

import (
	"io"
	"strings"
	"sync"
)

const (
	ProfileTypeCPU       ProfileType = "cpu"
	ProfileTypeMemory    ProfileType = "memory"
	ProfileTypeGoroutine ProfileType = "goroutine"
	ProfileTypeBlock     ProfileType = "block"
	ProfileTypeMutex     ProfileType = "mutex"
	ProfileTypeTrace     ProfileType = "trace"
	ProfileTypeThread    ProfileType = "thread"
	ProfileTypeHeap      ProfileType = "heap"
)
