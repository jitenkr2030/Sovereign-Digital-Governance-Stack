package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"neam-platform/policy/models"
)

// WorkflowEngine manages intervention workflows
type WorkflowEngine struct {
	store        WorkflowStore
	state        StateManager
	config       *WorkflowConfig
	stepHandlers map[string]WorkflowStepHandler
	activeFlows  map[string]*ActiveWorkflowContext
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
}

// WorkflowStore persists workflow state
type WorkflowStore interface {
	SaveWorkflowExecution(exec *WorkflowExecution) error
	GetWorkflowExecution(id string) (*WorkflowExecution, error)
	UpdateWorkflowExecution(exec *WorkflowExecution) error
	ListWorkflowExecutions(filter WorkflowFilter) ([]*WorkflowExecution, error)
}

// WorkflowConfig holds workflow configuration
type WorkflowConfig struct {
	DefaultTimeout      time.Duration
	MaxRetries          int
	ApprovalRequired    bool
	ApprovalTimeout     time.Duration
	NotificationEnabled bool
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID             string          `json:"id"`
	WorkflowID     string          `json:"workflow_id"`
	InterventionID string          `json:"intervention_id"`
	PolicyID       string          `json:"policy_id"`
	CurrentStep    string          `json:"current_step"`
	Steps          []WorkflowStep  `json:"steps"`
	Status         WorkflowStatus  `json:"status"`
	Context        map[string]interface{} `json:"context"`
	Errors         []WorkflowError `json:"errors"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CreatedBy      string          `json:"created_by"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name         string          `json:"name"`
	Type         StepType        `json:"type"`
	Status       StepStatus      `json:"status"`
	Handler      string          `json:"handler"`
	Parameters   json.RawMessage `json:"parameters"`
	Result       json.RawMessage `json:"result,omitempty"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	Error        string          `json:"error,omitempty"`
	RetryCount   int             `json:"retry_count"`
}

// StepType defines the type of workflow step
type StepType string

const (
	StepNotification   StepType = "NOTIFICATION"
	StepApproval       StepType = "APPROVAL"
	StepAction         StepType = "ACTION"
	StepCondition      StepType = "CONDITION"
	StepParallel       StepType = "PARALLEL"
	StepDelay          StepType = "DELAY"
	StepStateChange    StepType = "STATE_CHANGE"
	StepAssessment     StepType = "ASSESSMENT"
)

// StepStatus represents the status of a workflow step
type StepStatus string

const (
	StepPending    StepStatus = "PENDING"
	StepRunning    StepStatus = "RUNNING"
	StepCompleted  StepStatus = "COMPLETED"
	StepFailed     StepStatus = "FAILED"
	StepSkipped    StepStatus = "SKIPPED"
	StepWaiting    StepStatus = "WAITING"
)

// WorkflowStatus represents overall workflow status
type WorkflowStatus string

const (
	WorkflowRunning   WorkflowStatus = "RUNNING"
	WorkflowCompleted WorkflowStatus = "COMPLETED"
	WorkflowFailed    WorkflowStatus = "FAILED"
	WorkflowCancelled WorkflowStatus = "CANCELLED"
	WorkflowPending   WorkflowStatus = "PENDING"
)

// WorkflowError represents an error in workflow execution
type WorkflowError struct {
	Step    string    `json:"step"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// WorkflowFilter filters workflow queries
type WorkflowFilter struct {
	PolicyID       string
	InterventionID string
	Status         WorkflowStatus
	StartTime      *time.Time
	EndTime        *time.Time
	Limit          int
}

// WorkflowStepHandler handles execution of a specific step type
type WorkflowStepHandler interface {
	Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error)
	GetStepType() StepType
}

// StepResult contains the result of step execution
type StepResult struct {
	Success    bool                   `json:"success"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
	NextStep   string                 `json:"next_step,omitempty"`
	ShouldSkip bool                   `json:"should_skip"`
}

// ActiveWorkflowContext tracks an active workflow
type ActiveWorkflowContext struct {
	Execution   *WorkflowExecution
	StartedAt   time.Time
	LastUpdated time.Time
	TimeoutAt   time.Time
}

// DefaultWorkflowConfig returns default workflow configuration
func DefaultWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		DefaultTimeout:      30 * time.Minute,
		MaxRetries:          3,
		ApprovalRequired:    true,
		ApprovalTimeout:     4 * time.Hour,
		NotificationEnabled: true,
	}
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(store WorkflowStore, state StateManager, config *WorkflowConfig) *WorkflowEngine {
	if config == nil {
		config = DefaultWorkflowConfig()
	}
	
	engine := &WorkflowEngine{
		store:       store,
		state:       state,
		config:      config,
		stepHandlers: make(map[string]WorkflowStepHandler),
		activeFlows:  make(map[string]*ActiveWorkflowContext),
		stopChan:    make(chan struct{}),
	}
	
	// Register default step handlers
	engine.registerDefaultHandlers()
	
	return engine
}

// registerDefaultHandlers registers built-in step handlers
func (w *WorkflowEngine) registerDefaultHandlers() {
	w.stepHandlers[string(StepNotification)] = &NotificationStepHandler{}
	w.stepHandlers[string(StepApproval)] = &ApprovalStepHandler{}
	w.stepHandlers[string(StepAction)] = &ActionStepHandler{}
	w.stepHandlers[string(StepCondition)] = &ConditionStepHandler{}
	w.stepHandlers[string(StepDelay)] = &DelayStepHandler{}
	w.stepHandlers[string(StepStateChange)] = &StateChangeStepHandler{}
	w.stepHandlers[string(StepAssessment)] = &AssessmentStepHandler{}
}

// RegisterStepHandler registers a custom step handler
func (w *WorkflowEngine) RegisterStepHandler(handler WorkflowStepHandler) {
	w.stepHandlers[string(handler.GetStepType())] = handler
}

// StartWorkflow initiates a new workflow
func (w *WorkflowEngine) StartWorkflow(workflowID string, intervention *models.Intervention) error {
	execution := &WorkflowExecution{
		ID:             generateWorkflowID(),
		WorkflowID:     workflowID,
		InterventionID: intervention.ID,
		PolicyID:       intervention.PolicyID,
		CurrentStep:    "start",
		Steps:          w.createWorkflowSteps(workflowID),
		Status:         WorkflowRunning,
		Context:        make(map[string]interface{}),
		StartedAt:      time.Now(),
		CreatedBy:      "system",
	}
	
	// Set initial context
	execution.Context["intervention"] = intervention
	execution.Context["policy_id"] = intervention.PolicyID
	execution.Context["region_id"] = intervention.Scope.Regions
	execution.Context["sector_id"] = intervention.Scope.Sectors
	
	// Save execution
	if err := w.store.SaveWorkflowExecution(execution); err != nil {
		return fmt.Errorf("failed to save workflow execution: %w", err)
	}
	
	// Track active workflow
	w.mu.Lock()
	w.activeFlows[execution.ID] = &ActiveWorkflowContext{
		Execution:   execution,
		StartedAt:   time.Now(),
		LastUpdated: time.Now(),
		TimeoutAt:   time.Now().Add(w.config.DefaultTimeout),
	}
	w.mu.Unlock()
	
	// Start execution goroutine
	go w.executeWorkflow(execution)
	
	return nil
}

// createWorkflowSteps creates the steps for a workflow
func (w *WorkflowEngine) createWorkflowSteps(workflowID string) []WorkflowStep {
	switch workflowID {
	case "emergency_relief":
		return []WorkflowStep{
			{Name: "start", Type: StepNotification, Handler: "notify_stakeholders", Status: StepPending},
			{Name: "approval", Type: StepApproval, Handler: "minister_approval", Status: StepPending},
			{Name: "funds", Type: StepAction, Handler: "disburse_funds", Status: StepPending},
			{Name: "distribution", Type: StepAction, Handler: "coordinate_distribution", Status: StepPending},
			{Name: "assessment", Type: StepAssessment, Handler: "impact_assessment", Status: StepPending},
		}
	case "price_stabilization":
		return []WorkflowStep{
			{Name: "start", Type: StepNotification, Handler: "notify_traders", Status: StepPending},
			{Name: "monitor", Type: StepDelay, Handler: "price_monitoring", Status: StepPending},
			{Name: "intervention", Type: StepAction, Handler: "release_stocks", Status: StepPending},
			{Name: "assessment", Type: StepAssessment, Handler: "price_impact", Status: StepPending},
		}
	case "employment_stimulus":
		return []WorkflowStep{
			{Name: "start", Type: StepNotification, Handler: "notify_agencies", Status: StepPending},
			{Name: "approval", Type: StepApproval, Handler: "finance_approval", Status: StepPending},
			{Name: "funds", Type: StepAction, Handler: "release_funds", Status: StepPending},
			{Name: "tracking", Type: StepStateChange, Handler: "setup_tracking", Status: StepPending},
			{Name: "assessment", Type: StepAssessment, Handler: "employment_impact", Status: StepPending},
		}
	default:
		return []WorkflowStep{
			{Name: "start", Type: StepNotification, Handler: "default_notification", Status: StepPending},
			{Name: "process", Type: StepAction, Handler: "default_action", Status: StepPending},
			{Name: "complete", Type: StepAssessment, Handler: "final_assessment", Status: StepPending},
		}
	}
}

// executeWorkflow runs the workflow steps
func (w *WorkflowEngine) executeWorkflow(execution *WorkflowExecution) {
	defer w.cleanupWorkflow(execution)
	
	for i := range execution.Steps {
		step := &execution.Steps[i]
		
		// Check if workflow was cancelled
		if w.isWorkflowCancelled(execution.ID) {
			return
		}
		
		// Execute step
		result, err := w.executeStep(step, execution)
		if err != nil {
			log.Printf("Workflow %s step %s failed: %v", execution.ID, step.Name, err)
			step.Error = err.Error()
			step.Status = StepFailed
			
			if step.RetryCount < w.config.MaxRetries {
				step.RetryCount++
				step.Status = StepPending
				continue
			}
			
			execution.Status = WorkflowFailed
			execution.Errors = append(execution.Errors, WorkflowError{
				Step:    step.Name,
				Message: err.Error(),
				Time:    time.Now(),
			})
			return
		}
		
		// Update step status
		now := time.Now()
		step.Status = StepCompleted
		step.CompletedAt = &now
		step.Result = mustMarshal(result.Data)
		
		execution.CurrentStep = step.Name
		
		// Update store
		w.store.UpdateWorkflowExecution(execution)
		
		// Check if we should skip remaining steps
		if result.ShouldSkip {
			for j := i + 1; j < len(execution.Steps); j++ {
				execution.Steps[j].Status = StepSkipped
			}
			break
		}
		
		// Check for branching
		if result.NextStep != "" {
			// Skip to specific step
			for j := i + 1; j < len(execution.Steps); j++ {
				if execution.Steps[j].Name == result.NextStep {
					i = j - 1 // -1 because loop will increment
					break
				}
			}
		}
	}
	
	// Mark as completed
	execution.Status = WorkflowCompleted
	now := time.Now()
	execution.CompletedAt = &now
	w.store.UpdateWorkflowExecution(execution)
	
	log.Printf("Workflow %s completed successfully", execution.ID)
}

// executeStep executes a single workflow step
func (w *WorkflowEngine) executeStep(step *WorkflowStep, execution *WorkflowExecution) (*StepResult, error) {
	// Get handler
	handler, ok := w.stepHandlers[step.Handler]
	if !ok {
		return nil, fmt.Errorf("unknown handler: %s", step.Handler)
	}
	
	// Update step status
	now := time.Now()
	step.Status = StepRunning
	step.StartedAt = &now
	
	// Execute handler
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	result, err := handler.Execute(ctx, *step, execution)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// cleanupWorkflow cleans up after workflow completion
func (w *WorkflowEngine) cleanupWorkflow(execution *WorkflowExecution) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	delete(w.activeFlows, execution.ID)
}

// isWorkflowCancelled checks if workflow was cancelled
func (w *WorkflowEngine) isWorkflowCancelled(executionID string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	ctx, ok := w.activeFlows[executionID]
	if !ok {
		return true
	}
	
	exec, err := w.store.GetWorkflowExecution(executionID)
	if err != nil {
		return false
	}
	
	return exec.Status == WorkflowCancelled
}

// GetWorkflowStatus returns the status of a workflow
func (w *WorkflowEngine) GetWorkflowStatus(executionID string) (*WorkflowStatusResponse, error) {
	exec, err := w.store.GetWorkflowExecution(executionID)
	if err != nil {
		return nil, err
	}
	
	response := &WorkflowStatusResponse{
		ID:             exec.ID,
		WorkflowID:     exec.WorkflowID,
		InterventionID: exec.InterventionID,
		Status:         string(exec.Status),
		CurrentStep:    exec.CurrentStep,
		Steps:          make([]StepStatusResponse, len(exec.Steps)),
		StartedAt:      exec.StartedAt,
		CompletedAt:    exec.CompletedAt,
	}
	
	for i, step := range exec.Steps {
		response.Steps[i] = StepStatusResponse{
			Name:      step.Name,
			Type:      string(step.Type),
			Status:    string(step.Status),
			StartedAt: step.StartedAt,
			CompletedAt: step.CompletedAt,
			Error:     step.Error,
		}
	}
	
	return response, nil
}

// ApproveIntervention approves a pending intervention
func (w *WorkflowEngine) ApproveIntervention(interventionID, approver, notes string) error {
	exec, err := w.findExecutionByIntervention(interventionID)
	if err != nil {
		return err
	}
	
	// Find approval step and mark as complete
	for i := range exec.Steps {
		if exec.Steps[i].Type == StepApproval {
			now := time.Now()
			exec.Steps[i].Status = StepCompleted
			exec.Steps[i].CompletedAt = &now
			exec.Steps[i].Result = mustMarshal(map[string]interface{}{
				"approved_by":  approver,
				"notes":        notes,
				"approved_at":  now,
			})
			break
		}
	}
	
	exec.Context["approved_by"] = approver
	exec.Context["approval_notes"] = notes
	exec.Context["approved_at"] = time.Now()
	
	return w.store.UpdateWorkflowExecution(exec)
}

// CancelIntervention cancels an intervention workflow
func (w *WorkflowEngine) CancelIntervention(interventionID, reason string) error {
	exec, err := w.findExecutionByIntervention(interventionID)
	if err != nil {
		return err
	}
	
	exec.Status = WorkflowCancelled
	now := time.Now()
	exec.CompletedAt = &now
	
	exec.Errors = append(exec.Errors, WorkflowError{
		Step:    "cancellation",
		Message: reason,
		Time:    now,
	})
	
	return w.store.UpdateWorkflowExecution(exec)
}

// GetActiveWorkflows returns all active workflows
func (w *WorkflowEngine) GetActiveWorkflows() ([]*ActiveWorkflowInfo, error) {
	w.mu.RLock()
	flows := make([]*ActiveWorkflowInfo, 0, len(w.activeFlows))
	w.mu.RUnlock()
	
	execs, err := w.store.ListWorkflowExecutions(WorkflowFilter{
		Status: WorkflowRunning,
	})
	if err != nil {
		return nil, err
	}
	
	for _, exec := range execs {
		flows = append(flows, &ActiveWorkflowInfo{
			ID:            exec.ID,
			WorkflowID:    exec.WorkflowID,
			InterventionID: exec.InterventionID,
			Status:        string(exec.Status),
			CurrentStep:   exec.CurrentStep,
			Progress:      w.calculateProgress(exec),
			StartedAt:     exec.StartedAt,
		})
	}
	
	return flows, nil
}

// findExecutionByIntervention finds workflow execution by intervention ID
func (w *WorkflowEngine) findExecutionByIntervention(interventionID string) (*WorkflowExecution, error) {
	execs, err := w.store.ListWorkflowExecutions(WorkflowFilter{
		InterventionID: interventionID,
	})
	if err != nil {
		return nil, err
	}
	
	if len(execs) == 0 {
		return nil, fmt.Errorf("workflow execution not found for intervention: %s", interventionID)
	}
	
	return execs[0], nil
}

// calculateProgress calculates workflow progress percentage
func (w *WorkflowEngine) calculateProgress(exec *WorkflowExecution) float64 {
	total := len(exec.Steps)
	completed := 0
	
	for _, step := range exec.Steps {
		if step.Status == StepCompleted {
			completed++
		}
	}
	
	return float64(completed) / float64(total) * 100
}

// WorkflowStatusResponse represents workflow status for API
type WorkflowStatusResponse struct {
	ID             string              `json:"id"`
	WorkflowID     string              `json:"workflow_id"`
	InterventionID string              `json:"intervention_id"`
	Status         string              `json:"status"`
	CurrentStep    string              `json:"current_step"`
	Steps          []StepStatusResponse `json:"steps"`
	StartedAt      time.Time           `json:"started_at"`
	CompletedAt    *time.Time          `json:"completed_at,omitempty"`
}

// StepStatusResponse represents step status for API
type StepStatusResponse struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// ActiveWorkflowInfo represents brief info about active workflow
type ActiveWorkflowInfo struct {
	ID            string    `json:"id"`
	WorkflowID    string    `json:"workflow_id"`
	InterventionID string   `json:"intervention_id"`
	Status        string    `json:"status"`
	CurrentStep   string    `json:"current_step"`
	Progress      float64   `json:"progress"`
	StartedAt     time.Time `json:"started_at"`
}

// Step handlers

type NotificationStepHandler struct{}

func (n *NotificationStepHandler) GetStepType() StepType { return StepNotification }
func (n *NotificationStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	// Send notifications
	log.Printf("Sending notifications for workflow %s, step %s", exec.ID, step.Name)
	return &StepResult{Success: true, Data: map[string]interface{}{
		"notifications_sent": true,
	}}, nil
}

type ApprovalStepHandler struct{}

func (a *ApprovalStepHandler) GetStepType() StepType { return StepApproval }
func (a *ApprovalStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	// Wait for approval - in production this would integrate with approval system
	log.Printf("Approval required for workflow %s", exec.ID)
	
	// Check if auto-approved
	if exec.Context["auto_approve"] == true {
		return &StepResult{Success: true, Data: map[string]interface{}{
			"auto_approved": true,
		}}, nil
	}
	
	// Return waiting status
	return &StepResult{Success: true, ShouldSkip: false, Data: map[string]interface{}{
		"waiting_approval": true,
	}}, nil
}

type ActionStepHandler struct{}

func (a *ActionStepHandler) GetStepType() StepType { return StepAction }
func (a *ActionStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	log.Printf("Executing action %s for workflow %s", step.Handler, exec.ID)
	return &StepResult{Success: true, Data: map[string]interface{}{
		"action_completed": step.Handler,
	}}, nil
}

type ConditionStepHandler struct{}

func (c *ConditionStepHandler) GetStepType() StepType { return StepCondition }
func (c *ConditionStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	return &StepResult{Success: true, ShouldSkip: false}, nil
}

type DelayStepHandler struct{}

func (d *DelayStepHandler) GetStepType() StepType { return StepDelay }
func (d *DelayStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	// Simulate delay
	time.Sleep(100 * time.Millisecond)
	return &StepResult{Success: true}, nil
}

type StateChangeStepHandler struct{}

func (s *StateChangeStepHandler) GetStepType() StepType { return StepStateChange }
func (s *StateChangeStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	log.Printf("Executing state change for workflow %s", exec.ID)
	return &StepResult{Success: true, Data: map[string]interface{}{
		"state_changed": true,
	}}, nil
}

type AssessmentStepHandler struct{}

func (a *AssessmentStepHandler) GetStepType() StepType { return StepAssessment }
func (a *AssessmentStepHandler) Execute(ctx context.Context, step WorkflowStep, exec *WorkflowExecution) (*StepResult, error) {
	log.Printf("Performing assessment for workflow %s", exec.ID)
	return &StepResult{Success: true, Data: map[string]interface{}{
		"assessment_complete": true,
	}}, nil
}

// Helper function
func generateWorkflowID() string {
	return fmt.Sprintf("wf-%d", time.Now().UnixNano())
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
