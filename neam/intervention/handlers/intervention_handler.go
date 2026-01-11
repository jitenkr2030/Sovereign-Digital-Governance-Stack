package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"intervention-service/models"
	"intervention-service/services"
	"intervention-service/workflow"
)

// InterventionHandler handles HTTP requests for interventions
type InterventionHandler struct {
	policyService    *services.PolicyService
	interventionService *services.InterventionService
	scenarioService  *services.ScenarioService
	emergencyService *services.EmergencyService
	workflowEngine   *workflow.TemporalEngine
}

// NewInterventionHandler creates a new intervention handler
func NewInterventionHandler(
	policyService *services.PolicyService,
	interventionService *services.InterventionService,
	scenarioService *services.ScenarioService,
	emergencyService *services.EmergencyService,
	workflowEngine *workflow.TemporalEngine,
) *InterventionHandler {
	return &InterventionHandler{
		policyService:     policyService,
		interventionService: interventionService,
		scenarioService:   scenarioService,
		emergencyService:  emergencyService,
		workflowEngine:    workflowEngine,
	}
}

// HealthCheck handles health check requests
func (h *InterventionHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-intervention",
	})
}

// Policy handlers

func (h *InterventionHandler) CreatePolicy(c *gin.Context) {
	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.policyService.CreatePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      policy.ID,
		"status":  policy.Status,
		"message": "Policy created successfully",
	})
}

func (h *InterventionHandler) ListPolicies(c *gin.Context) {
	filters := make(map[string]interface{})
	
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if policyType := c.Query("type"); policyType != "" {
		filters["type"] = policyType
	}

	policies, err := h.policyService.ListPolicies(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"total":    len(policies),
	})
}

func (h *InterventionHandler) GetPolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return
	}

	policy, err := h.policyService.GetPolicy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

func (h *InterventionHandler) UpdatePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return
	}

	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy.ID = id
	if err := h.policyService.UpdatePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy updated successfully"})
}

func (h *InterventionHandler) DeletePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy ID"})
		return
	}

	if err := h.policyService.DeletePolicy(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy deleted successfully"})
}

// Intervention handlers

func (h *InterventionHandler) CreateIntervention(c *gin.Context) {
	var intervention models.Intervention
	if err := c.ShouldBindJSON(&intervention); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.interventionService.CreateIntervention(c.Request.Context(), &intervention); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      intervention.ID,
		"status":  intervention.Status,
		"message": "Intervention created successfully",
	})
}

func (h *InterventionHandler) ListInterventions(c *gin.Context) {
	filters := make(map[string]interface{})
	
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if policyID := c.Query("policy_id"); policyID != "" {
		filters["policy_id"] = policyID
	}

	interventions, err := h.interventionService.ListInterventions(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interventions": interventions,
		"total":         len(interventions),
	})
}

func (h *InterventionHandler) GetIntervention(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intervention ID"})
		return
	}

	intervention, err := h.interventionService.GetIntervention(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "intervention not found"})
		return
	}

	c.JSON(http.StatusOK, intervention)
}

func (h *InterventionHandler) ApproveIntervention(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intervention ID"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.interventionService.ApproveIntervention(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"status":  "approved",
		"message": "Intervention approved successfully",
	})
}

func (h *InterventionHandler) RejectIntervention(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intervention ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.interventionService.RejectIntervention(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"status":  "rejected",
		"message": "Intervention rejected",
	})
}

func (h *InterventionHandler) ExecuteIntervention(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intervention ID"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.interventionService.ExecuteIntervention(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"status":  "executing",
		"message": "Intervention execution started",
	})
}

func (h *InterventionHandler) RollbackIntervention(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intervention ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.interventionService.RollbackIntervention(c.Request.Context(), id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"status":  "rolling_back",
		"message": "Intervention rollback initiated",
	})
}

func (h *InterventionHandler) GetInterventionStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid intervention ID"})
		return
	}

	status, err := h.interventionService.GetInterventionStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Emergency handlers

func (h *InterventionHandler) CreateEmergencyLock(c *gin.Context) {
	var lock models.EmergencyLock
	if err := c.ShouldBindJSON(&lock); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.emergencyService.CreateEmergencyLock(c.Request.Context(), &lock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        lock.ID,
		"status":    lock.Status,
		"message":   "Emergency lock created",
		"timestamp": lock.CreatedAt.Format(time.RFC3339),
	})
}

func (h *InterventionHandler) ListEmergencyLocks(c *gin.Context) {
	filters := make(map[string]interface{})
	
	if lockType := c.Query("type"); lockType != "" {
		filters["lock_type"] = lockType
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["resource_type"] = resourceType
	}

	locks, err := h.emergencyService.ListEmergencyLocks(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locks": locks,
		"total": len(locks),
	})
}

func (h *InterventionHandler) ReleaseEmergencyLock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lock ID"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))

	if err := h.emergencyService.ReleaseEmergencyLock(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Emergency lock released"})
}

func (h *InterventionHandler) CreateOverride(c *gin.Context) {
	var override models.Override
	if err := c.ShouldBindJSON(&override); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.emergencyService.CreateOverride(c.Request.Context(), &override); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      override.ID,
		"status":  override.Status,
		"message": "Override activated",
		"expires": override.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *InterventionHandler) GetEmergencyStatus(c *gin.Context) {
	status, err := h.emergencyService.GetEmergencyStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Scenario handlers

func (h *InterventionHandler) SimulateScenario(c *gin.Context) {
	var scenario models.Scenario
	if err := c.ShouldBindJSON(&scenario); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.scenarioService.SimulateScenario(c.Request.Context(), &scenario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"run_id":   results.RunID,
		"status":   results.Status,
		"summary":  results.Summary,
		"duration": results.Duration.String(),
	})
}

func (h *InterventionHandler) ListScenarioTemplates(c *gin.Context) {
	templates, err := h.scenarioService.ListScenarioTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"total":     len(templates),
	})
}

func (h *InterventionHandler) CreateScenarioTemplate(c *gin.Context) {
	var template models.ScenarioTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.scenarioService.CreateScenarioTemplate(c.Request.Context(), &template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      template.ID,
		"message": "Scenario template created",
	})
}

// Workflow handlers

func (h *InterventionHandler) GetPendingApprovals(c *gin.Context) {
	approvals, err := h.workflowEngine.GetPendingApprovals(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"approvals": approvals,
		"total":     len(approvals),
	})
}

func (h *InterventionHandler) SignWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req struct {
		Signature string `json:"signature"`
		SignerID  string `json:"signer_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workflowEngine.SignWorkflow(c.Request.Context(), workflowID, req.SignerID, req.Signature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow_id": workflowID,
		"status":      "signed",
		"message":     "Workflow signed successfully",
	})
}
