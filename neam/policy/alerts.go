package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"neam-platform/policy/models"
)

// AlertDispatcher manages alert distribution
type AlertDispatcher struct {
	config       *AlertConfig
	channels     map[models.NotificationChannel]AlertChannel
	httpClient   *http.Client
	alertQueue   chan *Alert
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
}

// AlertConfig holds alert configuration
type AlertConfig struct {
	MaxRetries      int
	RetryDelay      time.Duration
	BatchSize       int
	BatchInterval   time.Duration
	EmailEnabled    bool
	SMSEnabled      bool
	WebhookEnabled  bool
	DashboardEnabled bool
	RateLimit       int // Alerts per minute
}

// AlertChannel handles alerts for a specific channel
type AlertChannel interface {
	Send(ctx context.Context, alert *Alert) (*AlertResult, error)
	GetChannelType() models.NotificationChannel
}

// AlertResult contains the result of alert sending
type AlertResult struct {
	Success    bool     `json:"success"`
	Recipient  string   `json:"recipient"`
	Channel    string   `json:"channel"`
	MessageID  string   `json:"message_id,omitempty"`
	Error      string   `json:"error,omitempty"`
	SentAt     time.Time `json:"sent_at"`
}

// DefaultAlertConfig returns default alert configuration
func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		MaxRetries:       3,
		RetryDelay:       5 * time.Second,
		BatchSize:        100,
		BatchInterval:    1 * time.Minute,
		EmailEnabled:     true,
		SMSEnabled:       true,
		WebhookEnabled:   true,
		DashboardEnabled: true,
		RateLimit:        1000,
	}
}

// NewAlertDispatcher creates a new alert dispatcher
func NewAlertDispatcher(config *AlertConfig) *AlertDispatcher {
	if config == nil {
		config = DefaultAlertConfig()
	}
	
	dispatcher := &AlertDispatcher{
		config:     config,
		channels:   make(map[models.NotificationChannel]AlertChannel),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		alertQueue: make(chan *Alert, 10000),
		stopChan:   make(chan struct{}),
	}
	
	// Register default channels
	dispatcher.registerDefaultChannels()
	
	return dispatcher
}

// registerDefaultChannels registers built-in alert channels
func (a *AlertDispatcher) registerDefaultChannels() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.channels[models.ChannelEmail] = &EmailChannel{}
	a.channels[models.ChannelWebhook] = &WebhookChannel{
		httpClient: a.httpClient,
	}
	a.channels[models.ChannelDashboard] = &DashboardChannel{}
}

// RegisterChannel registers a custom alert channel
func (a *AlertDispatcher) RegisterChannel(channel AlertChannel) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.channels[channel.GetChannelType()] = channel
}

// Start begins the alert dispatcher
func (a *AlertDispatcher) Start() {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.mu.Unlock()
	
	log.Println("Starting Alert Dispatcher...")
	
	ticker := time.NewTicker(a.config.BatchInterval)
	defer ticker.Stop()
	
	alertBatch := make([]*Alert, 0, a.config.BatchSize)
	
	for {
		select {
		case <-ticker.C:
			if len(alertBatch) > 0 {
				a.sendBatch(alertBatch)
				alertBatch = alertBatch[:0]
			}
		case alert := <-a.alertQueue:
			if len(alertBatch) >= a.config.BatchSize {
				a.sendBatch(alertBatch)
				alertBatch = alertBatch[:0]
			}
			alertBatch = append(alertBatch, alert)
		case <-a.stopChan:
			// Send remaining alerts
			if len(alertBatch) > 0 {
				a.sendBatch(alertBatch)
			}
			log.Println("Alert Dispatcher stopped")
			return
		}
	}
}

// Stop halts the alert dispatcher
func (a *AlertDispatcher) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.running {
		close(a.stopChan)
		a.running = false
	}
}

// SendAlert sends a single alert
func (a *AlertDispatcher) SendAlert(alert *Alert) error {
	// Validate alert
	if err := a.validateAlert(alert); err != nil {
		return fmt.Errorf("invalid alert: %w", err)
	}
	
	// Set default values
	if alert.ID == "" {
		alert.ID = generateAlertID()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}
	
	// Send through each channel
	for _, channelType := range alert.Channels {
		a.mu.RLock()
		channel, ok := a.channels[channelType]
		a.mu.RUnlock()
		
		if !ok {
			log.Printf("Unknown channel type: %s", channelType)
			continue
		}
		
		result, err := channel.Send(context.Background(), alert)
		if err != nil {
			log.Printf("Failed to send alert via %s: %v", channelType, err)
			continue
		}
		
		log.Printf("Alert %s sent via %s to %s: %v", alert.ID, channelType, result.Recipient, result.Success)
	}
	
	return nil
}

// SendBulkAlerts sends multiple alerts
func (a *AlertDispatcher) SendBulkAlerts(alerts []*Alert) error {
	for _, alert := range alerts {
		if err := a.SendAlert(alert); err != nil {
			log.Printf("Failed to send alert: %v", err)
		}
	}
	return nil
}

// sendBatch sends a batch of alerts
func (a *AlertDispatcher) sendBatch(alerts []*Alert) {
	for _, alert := range alerts {
		_ = a.SendAlert(alert)
	}
}

// validateAlert validates an alert
func (a *AlertDispatcher) validateAlert(alert *Alert) error {
	if alert.Type == "" {
		return fmt.Errorf("alert type is required")
	}
	if len(alert.Recipients) == 0 && alert.Channels[0] != models.ChannelDashboard {
		return fmt.Errorf("at least one recipient is required")
	}
	if len(alert.Channels) == 0 {
		return fmt.Errorf("at least one channel is required")
	}
	return nil
}

// GetAlertStatus gets the status of an alert
func (a *AlertDispatcher) GetAlertStatus(alertID string) (*AlertStatus, error) {
	// In a real implementation, this would query a database
	return &AlertStatus{
		AlertID: alertID,
		Status:  "DELIVERED",
	}, nil
}

// CreateAlertFromPolicy creates an alert from a policy trigger
func (a *AlertDispatcher) CreateAlertFromPolicy(
	policy *models.PolicyDefinition,
	action models.PolicyAction,
	intervention *models.Intervention,
) *Alert {
	
	return &Alert{
		ID:              generateAlertID(),
		Type:            action.Type,
		Priority:        policy.Priority,
		Title:           fmt.Sprintf("Policy Alert: %s", policy.Name),
		Message:         action.Parameters.Message,
		Recipients:      action.Parameters.Recipients,
		Channels:        action.Parameters.Channels,
		PolicyID:        policy.ID,
		InterventionID:  intervention.ID,
		RegionID:        getFirstOrDefault(policy.TargetScope.Regions, "ALL"),
		SectorID:        getFirstOrDefault(policy.TargetScope.Sectors, "ALL"),
		Metadata: map[string]interface{}{
			"policy_name":   policy.Name,
			"category":      policy.Category,
			"severity":      intervention.TriggerEvent.Condition,
			"metric_value":  intervention.TriggerEvent.MetricValue,
			"threshold":     intervention.TriggerEvent.Threshold,
		},
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
}

// Channel implementations

type EmailChannel struct{}

func (e *EmailChannel) GetChannelType() models.NotificationChannel { return models.ChannelEmail }
func (e *EmailChannel) Send(ctx context.Context, alert *Alert) (*AlertResult, error) {
	// In production, integrate with email service (SendGrid, SES, etc.)
	log.Printf("[Email] Sending alert %s to %v", alert.ID, alert.Recipients)
	
	return &AlertResult{
		Success:   true,
		Recipient: alert.Recipients[0],
		Channel:   string(models.ChannelEmail),
		SentAt:    time.Now(),
	}, nil
}

type WebhookChannel struct {
	httpClient *http.Client
}

func (w *WebhookChannel) GetChannelType() models.NotificationChannel { return models.ChannelWebhook }
func (w *WebhookChannel) Send(ctx context.Context, alert *Alert) (*AlertResult, error) {
	// Send webhook to configured endpoints
	for _, recipient := range alert.Recipients {
		data, _ := json.Marshal(map[string]interface{}{
			"alert_id":      alert.ID,
			"title":         alert.Title,
			"message":       alert.Message,
			"priority":      alert.Priority,
			"region_id":     alert.RegionID,
			"sector_id":     alert.SectorID,
			"metadata":      alert.Metadata,
			"created_at":    alert.CreatedAt,
		})
		
		// In production, make actual HTTP POST request
		log.Printf("[Webhook] Would send to %s: %s", recipient, string(data))
	}
	
	return &AlertResult{
		Success:    true,
		Recipient:  alert.Recipients[0],
		Channel:    string(models.ChannelWebhook),
		MessageID:  alert.ID,
		SentAt:     time.Now(),
	}, nil
}

type DashboardChannel struct{}

func (d *DashboardChannel) GetChannelType() models.NotificationChannel { return models.ChannelDashboard }
func (d *DashboardChannel) Send(ctx context.Context, alert *Alert) (*AlertResult, error) {
	// Push to Redis pub/sub for real-time dashboard updates
	// In production, integrate with WebSocket or SSE
	log.Printf("[Dashboard] Pushing alert %s to dashboard", alert.ID)
	
	return &AlertResult{
		Success:   true,
		Recipient: "dashboard",
		Channel:   string(models.ChannelDashboard),
		SentAt:    time.Now(),
	}, nil
}

// Helper functions

func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

func getFirstOrDefault(slice []string, defaultValue string) string {
	if len(slice) > 0 {
		return slice[0]
	}
	return defaultValue
}

// MinistryAlertConfig holds ministry-specific alert configurations
type MinistryAlertConfig struct {
	MinistryID     string   `json:"ministry_id"`
	MinistryName   string   `json:"ministry_name"`
	EmailRecipients []string `json:"email_recipients"`
	WebhookURL     string   `json:"webhook_url"`
	AlertCategories []models.PolicyCategory `json:"alert_categories"`
	PriorityLevel  int      `json:"priority_level"` // Minimum priority to alert
}

// MinistryAlertManager manages ministry-specific alerting
type MinistryAlertManager struct {
	configs map[string]*MinistryAlertConfig
	alerter *AlertDispatcher
	mu      sync.RWMutex
}

// NewMinistryAlertManager creates a new ministry alert manager
func NewMinistryAlertManager(alerter *AlertDispatcher) *MinistryAlertManager {
	return &MinistryAlertManager{
		configs: make(map[string]*MinistryAlertConfig),
		alerter: alerter,
	}
}

// RegisterMinistry registers a ministry for alerts
func (m *MinistryAlertManager) RegisterMinistry(config *MinistryAlertConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.configs[config.MinistryID] = config
}

// GetMinistryConfig gets ministry configuration
func (m *MinistryAlertManager) GetMinistryConfig(ministryID string) (*MinistryAlertConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	config, ok := m.configs[ministryID]
	if !ok {
		return nil, fmt.Errorf("ministry not found: %s", ministryID)
	}
	
	return config, nil
}

// DispatchToMinistry dispatches an alert to relevant ministries
func (m *MinistryAlertManager) DispatchToMinistry(category models.PolicyCategory, priority int, alert *Alert) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, config := range m.configs {
		// Check if ministry should receive this alert
		if !containsCategory(config.AlertCategories, category) {
			continue
		}
		
		if priority < config.PriorityLevel {
			continue
		}
		
		// Create ministry-specific alert
		ministryAlert := *alert
		ministryAlert.Recipients = config.EmailRecipients
		ministryAlert.Metadata["ministry_id"] = config.MinistryID
		ministryAlert.Metadata["ministry_name"] = config.MinistryName
		
		// Add webhook if configured
		if config.WebhookURL != "" {
			ministryAlert.Channels = append(ministryAlert.Channels, models.ChannelWebhook)
			ministryAlert.Recipients = append(ministryAlert.Recipients, config.WebhookURL)
		}
		
		_ = m.alerter.SendAlert(&ministryAlert)
	}
}

func containsCategory(categories []models.PolicyCategory, category models.PolicyCategory) bool {
	for _, c := range categories {
		if c == category {
			return true
		}
	}
	return false
}

// AlertTemplateManager manages alert message templates
type AlertTemplateManager struct {
	templates map[string]*AlertTemplate
}

// AlertTemplate defines an alert message template
type AlertTemplate struct {
	TemplateID   string                 `json:"template_id"`
	Name         string                 `json:"name"`
	Subject      string                 `json:"subject"`
	Body         string                 `json:"body"`
	Variables    []string               `json:"variables"`
	Format       string                 `json:"format"` // EMAIL, SMS, WEBHOOK
}

// NewAlertTemplateManager creates a new template manager
func NewAlertTemplateManager() *AlertTemplateManager {
	manager := &AlertTemplateManager{
		templates: make(map[string]*AlertTemplate),
	}
	
	// Register default templates
	manager.registerDefaultTemplates()
	
	return manager
}

// registerDefaultTemplates registers built-in alert templates
func (t *AlertTemplateManager) registerDefaultTemplates() {
	t.templates["inflation_alert"] = &AlertTemplate{
		TemplateID: "inflation_alert",
		Name:       "Inflation Alert",
		Subject:    "Inflation Threshold Exceeded: {{metric_value}}%",
		Body: `Dear Policy Team,

The inflation metric "{{metric_name}}" has exceeded the threshold of {{threshold}}%.

Current Value: {{metric_value}}%
Region: {{region_id}}
Sector: {{sector_id}}

Recommended Actions:
{{recommendations}}

Please review and take appropriate action.

NEAM Policy System`,
		Variables: []string{"metric_name", "metric_value", "threshold", "region_id", "sector_id", "recommendations"},
		Format:    "EMAIL",
	}
	
	t.templates["emergency_alert"] = &AlertTemplate{
		TemplateID: "emergency_alert",
		Name:       "Emergency Alert",
		Subject:    "EMERGENCY: {{policy_name}}",
		Body: `URGENT ALERT

Policy: {{policy_name}}
Category: {{category}}
Severity: {{severity}}

{{message}}

Immediate action required.

NEAM Emergency System`,
		Variables: []string{"policy_name", "category", "severity", "message"},
		Format:    "EMAIL",
	}
	
	t.templates["intervention_update"] = &AlertTemplate{
		TemplateID: "intervention_update",
		Name:       "Intervention Update",
		Subject:    "Intervention Update: {{intervention_id}}",
		Body: `Intervention Status Update

Intervention ID: {{intervention_id}}
Policy: {{policy_name}}
Status: {{status}}

{{update_message}}

NEAM Policy System`,
		Variables: []string{"intervention_id", "policy_name", "status", "update_message"},
		Format:    "EMAIL",
	}
}

// GetTemplate gets a template by ID
func (t *AlertTemplateManager) GetTemplate(templateID string) (*AlertTemplate, error) {
	template, ok := t.templates[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	return template, nil
}

// RenderTemplate renders a template with provided variables
func (t *AlertTemplateManager) RenderTemplate(templateID string, variables map[string]string) (string, error) {
	template, err := t.GetTemplate(templateID)
	if err != nil {
		return "", err
	}
	
	result := template.Body
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = replaceAll(result, placeholder, value)
	}
	
	return result, nil
}

func replaceAll(s, old, new string) string {
	result := s
	for {
		if result == "" {
			break
		}
		newResult := replaceOne(result, old, new)
		if newResult == result {
			break
		}
		result = newResult
	}
	return result
}

func replaceOne(s, old, new string) string {
	if i := find(s, old); i >= 0 {
		return s[:i] + new + s[i+len(old):]
	}
	return s
}

func find(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
