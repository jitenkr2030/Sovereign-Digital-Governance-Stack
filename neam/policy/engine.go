package policy

import (
	"fmt"
	"log"
	"sync"
	"time"

	"neam-platform/policy/models"
)

// Engine evaluates policies and triggers interventions
type Engine struct {
	store      PolicyStore
	workflow   WorkflowManager
	state      StateManager
	alerter    AlertDispatcher
	evaluators map[models.PolicyCategory]RuleEvaluator
	mu         sync.RWMutex
	running    bool
	stopChan   chan struct{}
}

// PolicyStore interface for data persistence
type PolicyStore interface {
	GetActivePolicies() ([]models.PolicyDefinition, error)
	GetPolicyByID(id string) (*models.PolicyDefinition, error)
	GetPoliciesByCategory(category models.PolicyCategory) ([]models.PolicyDefinition, error)
	SavePolicy(policy *models.PolicyDefinition) error
	GetPolicyHistory(policyID string) ([]models.PolicyDefinition, error)
}

// WorkflowManager handles intervention workflows
type WorkflowManager interface {
	StartWorkflow(workflowID string, intervention *models.Intervention) error
	GetWorkflowStatus(workflowID string) (*WorkflowStatus, error)
	ApproveIntervention(interventionID string, approver string, notes string) error
	CancelIntervention(interventionID string, reason string) error
	GetActiveWorkflows() ([]*ActiveWorkflow, error)
}

// StateManager manages policy state with Redis
type StateManager interface {
	SetState(key string, value interface{}, ttl time.Duration) error
	GetState(key string) (interface{}, error)
	DeleteState(key string) error
	GetMetricValue(metric string, regionID string, sectorID string) (float64, error)
	IncrementCounter(key string) (int64, error)
	GetCounter(key string) (int64, error)
}

// AlertDispatcher sends alerts to ministries
type AlertDispatcher interface {
	SendAlert(alert *Alert) error
	SendBulkAlerts(alerts []*Alert) error
	GetAlertStatus(alertID string) (*AlertStatus, error)
}

// WorkflowStatus represents workflow execution status
type WorkflowStatus struct {
	WorkflowID   string    `json:"workflow_id"`
	InterventionID string  `json:"intervention_id"`
	Status       string    `json:"status"`
	CurrentStep  string    `json:"current_step"`
	StartedAt    time.Time `json:"started_at"`
	LastUpdated  time.Time `json:"last_updated"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// Alert represents an alert to be sent
type Alert struct {
	ID          string                `json:"id"`
	Type        models.ActionType     `json:"type"`
	Priority    int                   `json:"priority"`
	Title       string                `json:"title"`
	Message     string                `json:"message"`
	Recipients  []string              `json:"recipients"`
	Channels    []models.NotificationChannel `json:"channels"`
	PolicyID    string                `json:"policy_id"`
	InterventionID string             `json:"intervention_id,omitempty"`
	RegionID    string                `json:"region_id,omitempty"`
	SectorID    string                `json:"sector_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	ExpiresAt   time.Time             `json:"expires_at,omitempty"`
}

// AlertStatus tracks alert delivery status
type AlertStatus struct {
	AlertID      string    `json:"alert_id"`
	SentTo       []string  `json:"sent_to"`
	FailedTo     []string  `json:"failed_to"`
	Status       string    `json:"status"`
	DeliveredAt  time.Time `json:"delivered_at"`
	ReadAt       time.Time `json:"read_at,omitempty"`
}

// ActiveWorkflow represents an ongoing workflow
type ActiveWorkflow struct {
	WorkflowID    string    `json:"workflow_id"`
	InterventionID string   `json:"intervention_id"`
	PolicyName    string    `json:"policy_name"`
	RegionID      string    `json:"region_id"`
	Status        string    `json:"status"`
	CurrentStep   string    `json:"current_step"`
	Progress      float64   `json:"progress"` // 0-100
	StartedAt     time.Time `json:"started_at"`
	EstimatedCompletion time.Time `json:"estimated_completion,omitempty"`
}

// RuleEvaluator evaluates rules for a specific category
type RuleEvaluator interface {
	// Evaluate evaluates a policy against current metrics
	Evaluate(policy *models.PolicyDefinition, currentMetrics map[string]float64) (*EvaluationResult, error)
	
	// CanEvaluate checks if this evaluator can handle the policy
	CanEvaluate(policy *models.PolicyDefinition) bool
}

// EvaluationResult contains the result of policy evaluation
type EvaluationResult struct {
	PolicyID     string                 `json:"policy_id"`
	Triggered    bool                   `json:"triggered"`
	Condition    string                 `json:"condition"`
	MetricValues map[string]float64     `json:"metric_values"`
	Threshold    float64                `json:"threshold"`
	SustainedFor int                    `json:"sustained_for_minutes"`
	Confidence   float64                `json:"confidence"`
	Severity     string                 `json:"severity"`
	Recommendation string               `json:"recommendation"`
}

// NewEngine creates a new policy engine
func NewEngine(
	store PolicyStore,
	workflow WorkflowManager,
	state StateManager,
	alerter AlertDispatcher,
) *Engine {
	
	engine := &Engine{
		store:      store,
		workflow:   workflow,
		state:      state,
		alerter:    alerter,
		evaluators: make(map[models.PolicyCategory]RuleEvaluator),
		stopChan:   make(chan struct{}),
	}
	
	// Register default evaluators
	engine.registerDefaultEvaluators()
	
	return engine
}

// registerDefaultEvaluators registers built-in rule evaluators
func (e *Engine) registerDefaultEvaluators() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.evaluators[models.CategoryInflation] = &InflationEvaluator{}
	e.evaluators[models.CategoryEmployment] = &EmploymentEvaluator{}
	e.evaluators[models.CategoryAgriculture] = &AgricultureEvaluator{}
	e.evaluators[models.CategoryEmergency] = &EmergencyEvaluator{}
}

// RegisterEvaluator registers a custom rule evaluator
func (e *Engine) RegisterEvaluator(category models.PolicyCategory, evaluator RuleEvaluator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.evaluators[category] = evaluator
}

// Start begins the policy engine
func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()
	
	log.Println("Starting Policy Engine...")
	
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			e.runEvaluationCycle()
		case <-e.stopChan:
			log.Println("Policy Engine stopped")
			return
		}
	}
}

// Stop halts the policy engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if e.running {
		close(e.stopChan)
		e.running = false
	}
}

// runEvaluationCycle evaluates all active policies
func (e *Engine) runEvaluationCycle() {
	policies, err := e.store.GetActivePolicies()
	if err != nil {
		log.Printf("Error fetching active policies: %v", err)
		return
	}
	
	for _, policy := range policies {
		result, err := e.EvaluatePolicy(policy)
		if err != nil {
			log.Printf("Error evaluating policy %s: %v", policy.ID, err)
			continue
		}
		
		if result.Triggered {
			e.handleTriggeredPolicy(policy, result)
		}
	}
}

// EvaluatePolicy evaluates a single policy
func (e *Engine) EvaluatePolicy(policy *models.PolicyDefinition) (*EvaluationResult, error) {
	e.mu.RLock()
	evaluator, ok := e.evaluators[policy.Category]
	e.mu.RUnlock()
	
	if !ok {
		// Use default evaluator
		evaluator = &DefaultEvaluator{}
	}
	
	// Get current metrics
	metrics, err := e.gatherMetrics(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to gather metrics: %w", err)
	}
	
	// Evaluate with appropriate evaluator
	result, err := evaluator.Evaluate(policy, metrics)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}
	
	result.PolicyID = policy.ID
	return result, nil
}

// gatherMetrics collects current metric values for policy evaluation
func (e *Engine) gatherMetrics(policy *models.PolicyDefinition) (map[string]float64, error) {
	metrics := make(map[string]float64)
	
	// Get metrics from state manager
	for _, condition := range policy.Rules.Conditions {
		metricKey := condition.Metric
		regionID := ""
		sectorID := ""
		
		if len(policy.TargetScope.Regions) > 0 {
			regionID = policy.TargetScope.Regions[0]
		}
		if len(policy.TargetScope.Sectors) > 0 {
			sectorID = policy.TargetScope.Sectors[0]
		}
		
		value, err := e.state.GetMetricValue(metricKey, regionID, sectorID)
		if err != nil {
			// Metric not available, use default or skip
			continue
		}
		
		metrics[metricKey] = value
	}
	
	return metrics, nil
}

// handleTriggeredPolicy processes a triggered policy
func (e *Engine) handleTriggeredPolicy(policy *models.PolicyDefinition, result *EvaluationResult) {
	// Check cooldown
	cooldownKey := fmt.Sprintf("policy_cooldown:%s", policy.ID)
	count, _ := e.state.GetCounter(cooldownKey)
	if count > 0 && time.Now().Add(-time.Duration(policy.CooldownMinutes)*time.Minute).Unix() < count {
		log.Printf("Policy %s in cooldown, skipping", policy.ID)
		return
	}
	
	// Create intervention
	intervention := &models.Intervention{
		ID:           generateUUID(),
		PolicyID:     policy.ID,
		PolicyName:   policy.Name,
		WorkflowID:   generateUUID(),
		Status:       models.InterventionPending,
		Scope:        policy.TargetScope,
		TriggerEvent: models.TriggerEvent{
			Condition:    result.Condition,
			MetricValue:  result.MetricValues[result.Condition],
			Threshold:    result.Threshold,
			SustainedFor: result.SustainedFor,
			Evidence:     result.MetricValues,
		},
		CreatedAt: time.Now(),
		CreatedBy: "system",
	}
	
	// Execute policy actions
	for _, action := range policy.Actions {
		switch action.Type {
		case models.ActionAlert, models.ActionNotify:
			alert := e.createAlert(policy, action, intervention)
			_ = e.alerter.SendAlert(alert)
			
		case models.ActionWorkflow:
			_ = e.workflow.StartWorkflow(action.Parameters.WorkflowID, intervention)
			
		case models.ActionStateChange:
			_ = e.applyStateChange(action.Parameters)
		}
		
		// Log action
		intervention.ActionsTaken = append(intervention.ActionsTaken, models.ActionLog{
			ActionType: action.Type,
			Timestamp:  time.Now(),
			Status:     "EXECUTED",
			Details:    fmt.Sprintf("Executed action: %s", action.Type),
		})
	}
	
	// Set cooldown
	e.state.IncrementCounter(cooldownKey)
	
	log.Printf("Policy %s triggered intervention %s", policy.ID, intervention.ID)
}

// createAlert creates an alert from policy action
func (e *Engine) createAlert(policy *models.PolicyDefinition, action models.PolicyAction, intervention *models.Intervention) *Alert {
	return &Alert{
		ID:              generateUUID(),
		Type:            action.Type,
		Priority:        policy.Priority,
		Title:           fmt.Sprintf("Policy Alert: %s", policy.Name),
		Message:         action.Parameters.Message,
		Recipients:      action.Parameters.Recipients,
		Channels:        action.Parameters.Channels,
		PolicyID:        policy.ID,
		InterventionID:  intervention.ID,
		RegionID:        policy.TargetScope.Regions[0],
		SectorID:        policy.TargetScope.Sectors[0],
		Metadata: map[string]interface{}{
			"policy_name":   policy.Name,
			"category":      policy.Category,
			"severity":      intervention.TriggerEvent.Condition,
		},
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
}

// applyStateChange applies a state change
func (e *Engine) applyStateChange(params models.ActionParams) error {
	if params.TargetState != "" {
		stateKey := fmt.Sprintf("alert_level:%s", params.TargetState)
		return e.state.SetState(stateKey, "ACTIVE", 24*time.Hour)
	}
	return nil
}

// GetActiveInterventions returns currently active interventions
func (e *Engine) GetActiveInterventions() ([]*ActiveWorkflow, error) {
	return e.workflow.GetActiveWorkflows()
}

// ApproveIntervention approves a pending intervention
func (e *Engine) ApproveIntervention(interventionID, approver, notes string) error {
	return e.workflow.ApproveIntervention(interventionID, approver, notes)
}

// CancelIntervention cancels an intervention
func (e *Engine) CancelIntervention(interventionID, reason string) error {
	return e.workflow.CancelIntervention(interventionID, reason)
}

// GetPolicy returns a policy by ID
func (e *Engine) GetPolicy(id string) (*models.PolicyDefinition, error) {
	return e.store.GetPolicyByID(id)
}

// CreatePolicy creates a new policy
func (e *Engine) CreatePolicy(policy *models.PolicyDefinition) error {
	policy.ID = generateUUID()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.Version = 1
	policy.IsActive = true
	return e.store.SavePolicy(policy)
}

// UpdatePolicy updates an existing policy
func (e *Engine) UpdatePolicy(policy *models.PolicyDefinition) error {
	policy.UpdatedAt = time.Now()
	policy.Version++
	return e.store.SavePolicy(policy)
}

// ListPolicies lists all policies
func (e *Engine) ListPolicies() ([]models.PolicyDefinition, error) {
	return e.store.GetActivePolicies()
}

// ListPoliciesByCategory lists policies by category
func (e *Engine) ListPoliciesByCategory(category models.PolicyCategory) ([]models.PolicyDefinition, error) {
	return e.store.GetPoliciesByCategory(category)
}

// DefaultEvaluator is the default rule evaluator
type DefaultEvaluator struct{}

// Evaluate implements default policy evaluation
func (d *DefaultEvaluator) Evaluate(policy *models.PolicyDefinition, currentMetrics map[string]float64) (*EvaluationResult, error) {
	result := &EvaluationResult{
		PolicyID:     policy.ID,
		Triggered:    false,
		MetricValues: currentMetrics,
		Confidence:   0.9,
	}
	
	allConditionsMet := true
	
	for _, condition := range policy.Rules.Conditions {
		currentValue, ok := currentMetrics[condition.Metric]
		if !ok {
			allConditionsMet = false
			continue
		}
		
		threshold := policy.Rules.ThresholdValues[condition.Metric]
		result.Threshold = threshold
		
		met := d.evaluateCondition(condition, currentValue, threshold)
		if !met {
			allConditionsMet = false
			break
		}
		
		result.Condition = condition.Metric
	}
	
	// Check sustained period if required
	if policy.Rules.TriggerType == models.TriggerSustained && allConditionsMet {
		// In a real implementation, check sustained period
		result.SustainedFor = policy.Rules.SustainedPeriod
	}
	
	result.Triggered = allConditionsMet
	result.Severity = d.calculateSeverity(policy, currentMetrics)
	result.Recommendation = d.generateRecommendation(policy, result)
	
	return result, nil
}

// CanEvaluate checks if evaluator can handle the policy
func (d *DefaultEvaluator) CanEvaluate(policy *models.PolicyDefinition) bool {
	return true
}

// evaluateCondition evaluates a single condition
func (d *DefaultEvaluator) evaluateCondition(condition models.RuleCondition, currentValue, threshold float64) bool {
	switch condition.Operator {
	case models.OpGreaterThan:
		return currentValue > threshold
	case models.OpLessThan:
		return currentValue < threshold
	case models.OpGreaterOrEqual:
		return currentValue >= threshold
	case models.OpLessOrEqual:
		return currentValue <= threshold
	case models.OpEqual:
		return currentValue == threshold
	default:
		return false
	}
}

// calculateSeverity calculates severity level
func (d *DefaultEvaluator) calculateSeverity(policy *models.PolicyDefinition, metrics map[string]float64) string {
	maxRatio := 0.0
	for _, condition := range policy.Rules.Conditions {
		current := metrics[condition.Metric]
		threshold := policy.Rules.ThresholdValues[condition.Metric]
		if threshold != 0 {
			ratio := current / threshold
			if ratio > maxRatio {
				maxRatio = ratio
			}
		}
	}
	
	switch {
	case maxRatio > 1.5:
		return "CRITICAL"
	case maxRatio > 1.2:
		return "HIGH"
	case maxRatio > 1.0:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// generateRecommendation generates a recommendation message
func (d *DefaultEvaluator) generateRecommendation(policy *models.PolicyDefinition, result *EvaluationResult) string {
	return fmt.Sprintf("Policy %s triggered. Immediate attention to %s required.", policy.Name, result.Condition)
}

// InflationEvaluator evaluates inflation-related policies
type InflationEvaluator struct{}

// CanEvaluate checks if this evaluator handles the policy
func (i *InflationEvaluator) CanEvaluate(policy *models.PolicyDefinition) bool {
	return policy.Category == models.CategoryInflation
}

// Evaluate implements inflation-specific evaluation
func (i *InflationEvaluator) Evaluate(policy *models.PolicyDefinition, currentMetrics map[string]float64) (*EvaluationResult, error) {
	result := &EvaluationResult{
		PolicyID:     policy.ID,
		Triggered:    false,
		MetricValues: currentMetrics,
		Confidence:   0.95,
	}
	
	// Inflation-specific evaluation logic
	cpiValue := currentMetrics["CPI"]
	foodInflation := currentMetrics["FOOD_INFLATION"]
	coreInflation := currentMetrics["CORE_INFLATION"]
	
	threshold := policy.Rules.ThresholdValues["CPI"]
	
	if cpiValue > threshold {
		result.Triggered = true
		result.Condition = "CPI"
		result.Threshold = threshold
		
		// Determine severity based on inflation components
		if foodInflation > coreInflation*1.5 {
			result.Severity = "HIGH"
			result.Recommendation = "High food inflation detected. Consider supply-side interventions."
		} else {
			result.Severity = "MEDIUM"
			result.Recommendation = "General inflation rising. Monitor monetary policy indicators."
		}
	}
	
	return result, nil
}

// EmploymentEvaluator evaluates employment-related policies
type EmploymentEvaluator struct{}

// CanEvaluate checks if this evaluator handles the policy
func (e *EmploymentEvaluator) CanEvaluate(policy *models.PolicyDefinition) bool {
	return policy.Category == models.CategoryEmployment
}

// Evaluate implements employment-specific evaluation
func (e *EmploymentEvaluator) Evaluate(policy *models.PolicyDefinition, currentMetrics map[string]float64) (*EvaluationResult, error) {
	result := &EvaluationResult{
		PolicyID:     policy.ID,
		Triggered:    false,
		MetricValues: currentMetrics,
		Confidence:   0.9,
	}
	
	unemploymentRate := currentMetrics["UNEMPLOYMENT_RATE"]
	threshold := policy.Rules.ThresholdValues["UNEMPLOYMENT_RATE"]
	
	if unemploymentRate > threshold {
		result.Triggered = true
		result.Condition = "UNEMPLOYMENT_RATE"
		result.Threshold = threshold
		result.Severity = "HIGH"
		result.Recommendation = "Unemployment threshold breached. Consider stimulus measures."
	}
	
	return result, nil
}

// AgricultureEvaluator evaluates agriculture-related policies
type AgricultureEvaluator struct{}

// CanEvaluate checks if this evaluator handles the policy
func (a *AgricultureEvaluator) CanEvaluate(policy *models.PolicyDefinition) bool {
	return policy.Category == models.CategoryAgriculture
}

// Evaluate implements agriculture-specific evaluation
func (a *AgricultureEvaluator) Evaluate(policy *models.PolicyDefinition, currentMetrics map[string]float64) (*EvaluationResult, error) {
	result := &EvaluationResult{
		PolicyID:     policy.ID,
		Triggered:    false,
		MetricValues: currentMetrics,
		Confidence:   0.85,
	}
	
	cropYield := currentMetrics["CROP_YIELD"]
	rainfallIndex := currentMetrics["RAINFALL_INDEX"]
	threshold := policy.Rules.ThresholdValues["CROP_YIELD"]
	
	if cropYield < threshold {
		result.Triggered = true
		result.Condition = "CROP_YIELD"
		result.Threshold = threshold
		result.Severity = "CRITICAL"
		result.Recommendation = "Crop failure detected. Emergency relief protocols should be activated."
	}
	
	return result, nil
}

// EmergencyEvaluator evaluates emergency policies
type EmergencyEvaluator struct{}

// CanEvaluate checks if this evaluator handles the policy
func (e *EmergencyEvaluator) CanEvaluate(policy *models.PolicyDefinition) bool {
	return policy.Category == models.CategoryEmergency
}

// Evaluate implements emergency-specific evaluation
func (e *EmergencyEvaluator) Evaluate(policy *models.PolicyDefinition, currentMetrics map[string]float64) (*EvaluationResult, error) {
	result := &EvaluationResult{
		PolicyID:     policy.ID,
		Triggered:    false,
		MetricValues: currentMetrics,
		Confidence:   1.0, // Emergency evaluations have highest confidence
	}
	
	// Emergency policies trigger immediately
	for _, condition := range policy.Rules.Conditions {
		currentValue := currentMetrics[condition.Metric]
		threshold := policy.Rules.ThresholdValues[condition.Metric]
		
		if currentValue > threshold || currentValue < (threshold * 0.5) {
			result.Triggered = true
			result.Condition = condition.Metric
			result.Threshold = threshold
			result.Severity = "CRITICAL"
			result.Recommendation = "EMERGENCY: Immediate intervention required."
			break
		}
	}
	
	return result, nil
}

// Helper function to generate UUID
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
