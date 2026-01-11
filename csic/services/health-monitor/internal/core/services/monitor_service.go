package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"csic-platform/health-monitor/internal/core/domain"
	"csic-platform/health-monitor/internal/core/ports"
	"csic-platform/health-monitor/pkg/metrics"
)

// MonitorService monitors service health
type MonitorService interface {
	ProcessHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error
	GetServiceStatus(ctx context.Context, serviceName string) (*domain.ServiceStatus, error)
	GetAllServiceStatuses(ctx context.Context) ([]*domain.ServiceStatus, error)
	GetHealthSummary(ctx context.Context) (*domain.HealthSummary, error)
	PerformHealthCheck(ctx context.Context, serviceName string, endpoint string, timeout time.Duration) (*domain.HealthCheck, error)
	StartHealthCheckMonitor(logger *zap.Logger)
	StartHeartbeatConsumer(logger *zap.Logger)
	RegisterService(ctx context.Context, req *domain.RegisterServiceRequest) (*domain.RegisterServiceResponse, error)
	GetRegisteredServices(ctx context.Context) ([]string, error)
}

// MonitorServiceService implements MonitorService
type MonitorServiceService struct {
	repositories  ports.Repositories
	cachePort     ports.CachePort
	messagingPort ports.MessagingPort
	logger        *zap.Logger
	metrics       *metrics.MetricsCollector

	// Internal state
	mu          sync.RWMutex
	services    map[string]*domain.ServiceStatus
	lastSummary time.Time
}

// NewMonitorService creates a new monitor service
func NewMonitorService(
	repositories ports.Repositories,
	cachePort ports.CachePort,
	messagingPort ports.MessagingPort,
	logger *zap.Logger,
	metricsCollector *metrics.MetricsCollector,
) MonitorService {
	return &MonitorServiceService{
		repositories:  repositories,
		cachePort:     cachePort,
		messagingPort: messagingPort,
		logger:        logger,
		metrics:       metricsCollector,
		services:      make(map[string]*domain.ServiceStatus),
	}
}

// ProcessHeartbeat processes a heartbeat from a service
func (s *MonitorServiceService) ProcessHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error {
	start := time.Now()

	// Update last heartbeat in cache
	ttl := 2 * time.Minute
	if err := s.cachePort.SetHeartbeat(ctx, heartbeat.ServiceName, heartbeat, ttl); err != nil {
		return fmt.Errorf("failed to set heartbeat: %w", err)
	}

	// Get or create service status
	status, err := s.getOrCreateServiceStatus(ctx, heartbeat.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	// Update status based on heartbeat
	previousStatus := status.Status
	status.LastHeartbeat = heartbeat.Timestamp
	status.Status = heartbeat.Status
	status.PreviousStatus = previousStatus
	status.LastCheck = time.Now()
	status.UpdatedAt = time.Now()

	// Update metrics from heartbeat
	if heartbeat.Metrics != nil {
		status.CPUUsage = heartbeat.Metrics["cpu_percent"]
		status.MemoryUsage = heartbeat.Metrics["memory_percent"]
		status.DiskUsage = heartbeat.Metrics["disk_percent"]
		status.ResponseTime = heartbeat.Metrics["latency_ms"]
		status.ErrorRate = heartbeat.Metrics["error_rate_percent"]
	}

	// Calculate uptime
	uptime, err := s.repositories.ServiceStatusRepository.GetServiceUptime(ctx, heartbeat.ServiceName, time.Now().Add(-24*time.Hour))
	if err == nil {
		status.Uptime = uptime
	}

	// Save to database
	if err := s.repositories.ServiceStatusRepository.UpsertServiceStatus(ctx, status); err != nil {
		return fmt.Errorf("failed to upsert service status: %w", err)
	}

	// Update local cache
	s.mu.Lock()
	s.services[heartbeat.ServiceName] = status
	s.mu.Unlock()

	// Check for status change
	if previousStatus != "" && previousStatus != status.Status {
		s.logger.Info("Service status changed",
			logger.String("service", heartbeat.ServiceName),
			logger.String("from", previousStatus),
			logger.String("to", status.Status),
		)

		// Create outage if service became unhealthy
		if status.Status == domain.HealthStatusUnhealthy && previousStatus == domain.HealthStatusHealthy {
			s.handleServiceDown(ctx, heartbeat.ServiceName, status)
		}

		// Create alert if service recovered
		if status.Status == domain.HealthStatusHealthy && previousStatus == domain.HealthStatusUnhealthy {
			s.handleServiceUp(ctx, heartbeat.ServiceName, status)
		}
	}

	s.metrics.RecordHeartbeatProcessed(heartbeat.ServiceName, float64(time.Since(start).Milliseconds()))

	return nil
}

// GetServiceStatus gets the status of a specific service
func (s *MonitorServiceService) GetServiceStatus(ctx context.Context, serviceName string) (*domain.ServiceStatus, error) {
	// Try local cache first
	s.mu.RLock()
	if status, ok := s.services[serviceName]; ok {
		s.mu.RUnlock()
		return status, nil
	}
	s.mu.RUnlock()

	// Fetch from database
	status, err := s.repositories.ServiceStatusRepository.GetServiceStatus(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get service status: %w", err)
	}

	if status == nil {
		return nil, nil
	}

	// Update local cache
	s.mu.Lock()
	s.services[serviceName] = status
	s.mu.Unlock()

	return status, nil
}

// GetAllServiceStatuses gets the status of all services
func (s *MonitorServiceService) GetAllServiceStatuses(ctx context.Context) ([]*domain.ServiceStatus, error) {
	// Fetch from database
	statuses, err := s.repositories.ServiceStatusRepository.GetAllServiceStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all service statuses: %w", err)
	}

	// Update local cache
	s.mu.Lock()
	for _, status := range statuses {
		s.services[status.Name] = status
	}
	s.mu.Unlock()

	return statuses, nil
}

// GetHealthSummary generates a health summary
func (s *MonitorServiceService) GetHealthSummary(ctx context.Context) (*domain.HealthSummary, error) {
	statuses, err := s.GetAllServiceStatuses(ctx)
	if err != nil {
		return nil, err
	}

	summary := &domain.HealthSummary{
		TotalServices:   len(statuses),
		LastUpdated:     time.Now(),
		Services:        statuses,
		RecentAlerts:    []*domain.Alert{},
		RecentOutages:   []*domain.Outage{},
	}

	for _, status := range statuses {
		switch status.Status {
		case domain.HealthStatusHealthy:
			summary.HealthyServices++
		case domain.HealthStatusDegraded:
			summary.DegradedServices++
		case domain.HealthStatusUnhealthy:
			summary.UnhealthyServices++
		default:
			summary.UnknownServices++
		}
	}

	// Get recent alerts
	alerts, err := s.getRecentAlerts(ctx)
	if err == nil {
		summary.RecentAlerts = alerts
	}

	// Get recent outages
	outages, err := s.getRecentOutages(ctx)
	if err == nil {
		summary.RecentOutages = outages
	}

	return summary, nil
}

// PerformHealthCheck performs an HTTP health check on a service
func (s *MonitorServiceService) PerformHealthCheck(ctx context.Context, serviceName string, endpoint string, timeout time.Duration) (*domain.HealthCheck, error) {
	start := time.Now()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Perform health check
	check := &domain.HealthCheck{
		ServiceName: serviceName,
		Endpoint:    endpoint,
		CheckedAt:   time.Now(),
	}

	resp, err := client.Get(endpoint)
	duration := time.Since(start)

	check.ResponseTime = float64(duration.Milliseconds())

	if err != nil {
		check.Status = domain.HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Health check failed: %v", err)
		return check, nil
	}
	defer resp.Body.Close()

	check.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		check.Status = domain.HealthStatusHealthy
		check.Message = "Health check passed"
	} else if resp.StatusCode >= 500 {
		check.Status = domain.HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Health check failed with status: %d", resp.StatusCode)
	} else {
		check.Status = domain.HealthStatusDegraded
		check.Message = fmt.Sprintf("Health check returned warning status: %d", resp.StatusCode)
	}

	// Update service status
	status, err := s.GetServiceStatus(ctx, serviceName)
	if err == nil && status != nil {
		status.LastCheck = check.CheckedAt
		status.Status = check.Status
		status.ResponseTime = check.ResponseTime
		status.UpdatedAt = time.Now()

		if err := s.repositories.ServiceStatusRepository.UpsertServiceStatus(ctx, status); err != nil {
			s.logger.Warn("Failed to update service status", logger.Error(err))
		}
	}

	s.metrics.RecordHealthCheck(serviceName, check.Status, check.ResponseTime)

	return check, nil
}

// StartHealthCheckMonitor starts the background health check monitor
func (s *MonitorServiceService) StartHealthCheckMonitor(logger *zap.Logger) {
	go func() {
		// Run health checks every 30 seconds
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.runHealthChecks()
			}
		}
	}()

	logger.Info("Health check monitor started")
}

// StartHeartbeatConsumer starts consuming heartbeat messages
func (s *MonitorServiceService) StartHeartbeatConsumer(logger *zap.Logger) {
	// This would consume from Kafka in production
	logger.Info("Heartbeat consumer started")
}

// RegisterService registers a new service
func (s *MonitorServiceService) RegisterService(ctx context.Context, req *domain.RegisterServiceRequest) (*domain.RegisterServiceResponse, error) {
	status := &domain.ServiceStatus{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Status:    domain.HealthStatusHealthy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repositories.ServiceStatusRepository.UpsertServiceStatus(ctx, status); err != nil {
		return nil, fmt.Errorf("failed to register service: %w", err)
	}

	s.mu.Lock()
	s.services[req.Name] = status
	s.mu.Unlock()

	logger.Info("Service registered",
		logger.String("service_id", status.ID),
		logger.String("name", req.Name),
	)

	return &domain.RegisterServiceResponse{
		ServiceID: status.ID,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}, nil
}

// GetRegisteredServices returns all registered services
func (s *MonitorServiceService) GetRegisteredServices(ctx context.Context) ([]string, error) {
	statuses, err := s.GetAllServiceStatuses(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(statuses))
	for i, status := range statuses {
		names[i] = status.Name
	}

	return names, nil
}

// getOrCreateServiceStatus gets or creates a service status
func (s *MonitorServiceService) getOrCreateServiceStatus(ctx context.Context, serviceName string) (*domain.ServiceStatus, error) {
	// Check local cache first
	s.mu.RLock()
	if status, ok := s.services[serviceName]; ok {
		s.mu.RUnlock()
		return status, nil
	}
	s.mu.RUnlock()

	// Fetch from database
	status, err := s.repositories.ServiceStatusRepository.GetServiceStatus(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	if status != nil {
		return status, nil
	}

	// Create new status
	status = &domain.ServiceStatus{
		ID:        uuid.New().String(),
		Name:      serviceName,
		Status:    domain.HealthStatusUnknown,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return status, nil
}

// handleServiceDown handles a service going down
func (s *MonitorServiceService) handleServiceDown(ctx context.Context, serviceName string, status *domain.ServiceStatus) {
	outage := &domain.Outage{
		ID:         uuid.New().String(),
		ServiceName: serviceName,
		Severity:   domain.AlertSeverityCritical,
		Status:     domain.OutageStatusDetected,
		StartedAt:  time.Now(),
		DetectedAt: time.Now(),
		Impact:     "Service is not responding",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repositories.OutageRepository.CreateOutage(ctx, outage); err != nil {
		s.logger.Error("Failed to create outage record", logger.Error(err))
	}
}

// handleServiceUp handles a service coming back up
func (s *MonitorServiceService) handleServiceUp(ctx context.Context, serviceName string, status *domain.ServiceStatus) {
	outages, err := s.repositories.OutageRepository.GetActiveOutages(ctx)
	if err != nil {
		s.logger.Error("Failed to get active outages", logger.Error(err))
		return
	}

	for _, outage := range outages {
		if outage.ServiceName == serviceName && outage.Status != domain.OutageStatusResolved {
			if err := s.repositories.OutageRepository.ResolveOutage(ctx, outage.ID, "Service recovered"); err != nil {
				s.logger.Error("Failed to resolve outage", logger.Error(err))
			}
		}
	}
}

// getRecentAlerts gets recent alerts
func (s *MonitorServiceService) getRecentAlerts(ctx context.Context) ([]*domain.Alert, error) {
	// This would be implemented in a real service
	return []*domain.Alert{}, nil
}

// getRecentOutages gets recent outages
func (s *MonitorServiceService) getRecentOutages(ctx context.Context) ([]*domain.Outage, error) {
	outages, err := s.repositories.OutageRepository.GetRecentOutages(ctx, time.Now().Add(-24*time.Hour))
	if err != nil {
		return nil, err
	}
	return outages, nil
}

// runHealthChecks runs health checks on all registered services
func (s *MonitorServiceService) runHealthChecks() {
	ctx := context.Background()

	statuses, err := s.GetAllServiceStatuses(ctx)
	if err != nil {
		s.logger.Error("Failed to get service statuses for health checks", logger.Error(err))
		return
	}

	for _, status := range statuses {
		// Skip services that have recently sent heartbeats
		if time.Since(status.LastHeartbeat) < 2*time.Minute {
			continue
		}

		// Perform health check
		check, err := s.PerformHealthCheck(ctx, status.Name, "http://"+status.Name+":8080/health", 5*time.Second)
		if err != nil {
			s.logger.Warn("Health check failed",
				logger.String("service", status.Name),
				logger.Error(err),
			)
			continue
		}

		s.logger.Debug("Health check completed",
			logger.String("service", status.Name),
			logger.String("status", check.Status),
		)
	}
}

// GetServiceMetrics gets metrics for a service
func (s *MonitorServiceService) GetServiceMetrics(ctx context.Context, serviceName string, since time.Time) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get uptime
	uptime, err := s.repositories.ServiceStatusRepository.GetServiceUptime(ctx, serviceName, since)
	if err == nil {
		metrics["uptime_percent"] = uptime
	}

	// Get current status
	status, err := s.GetServiceStatus(ctx, serviceName)
	if err == nil && status != nil {
		metrics["current_status"] = status.Status
		metrics["cpu_usage"] = status.CPUUsage
		metrics["memory_usage"] = status.MemoryUsage
		metrics["disk_usage"] = status.DiskUsage
		metrics["response_time_ms"] = status.ResponseTime
		metrics["error_rate"] = status.ErrorRate
	}

	return metrics, nil
}

// HeartbeatMessageHandler handles heartbeat messages
type HeartbeatMessageHandler struct {
	monitorService MonitorService
}

// NewHeartbeatMessageHandler creates a new heartbeat message handler
func NewHeartbeatMessageHandler(monitorService MonitorService) *HeartbeatMessageHandler {
	return &HeartbeatMessageHandler{
		monitorService: monitorService,
	}
}

// HandleHeartbeat processes a heartbeat message
func (h *HeartbeatMessageHandler) HandleHeartbeat(ctx context.Context, heartbeat *domain.Heartbeat) error {
	return h.monitorService.ProcessHeartbeat(ctx, heartbeat)
}

// HandleHealthCheck processes a health check result
func (h *HeartbeatMessageHandler) HandleHealthCheck(ctx context.Context, check *domain.HealthCheck) error {
	// Update service status based on health check
	_ = check
	return nil
}

// HandleStateUpdate processes a state update
func (h *HeartbeatMessageHandler) HandleStateUpdate(ctx context.Context, update map[string]interface{}) error {
	data, _ := json.Marshal(update)
	var heartbeat domain.Heartbeat
	if err := json.Unmarshal(data, &heartbeat); err != nil {
		return err
	}
	return h.monitorService.ProcessHeartbeat(ctx, &heartbeat)
}
