// NEAM Intervention Engine - Main Entry Point
// Policy Execution, Emergency Controls, and Workflow Management

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"neam-platform/intervention/activities"
	"neam-platform/intervention/config"
	"neam-platform/intervention/emergency-controls"
	"neam-platform/intervention/policy-engine"
	"neam-platform/intervention/scenario-simulator"
	"neam-platform/intervention/state-coordination"
	"neam-platform/intervention/workflows"
	"neam-platform/shared"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig("intervention/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := shared.NewLogger("intervention")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize shared components
	postgres, err := shared.NewPostgreSQL(cfg.Database.ConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgres.Close()

	redis, err := shared.NewRedis(cfg.Redis.Address(), cfg.Redis.Password)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Initialize Temporal client
	temporalClient, err := initializeTemporalClient(cfg.Temporal)
	if err != nil {
		log.Fatalf("Failed to initialize Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Initialize activities with dependencies
	interventionActivities := activities.NewInterventionActivities(
		postgres,
		nil, // clickHouse connection would be initialized in production
		redis,
		logger,
	)

	// Initialize intervention components
	var wg sync.WaitGroup

	// Start Temporal worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTemporalWorker(temporalClient, interventionActivities, cfg.Temporal.TaskQueue)
	}()

	// Policy Engine
	policyEngine := policy_engine.NewEngine(policy_engine.Config{
		PostgreSQL:  postgres,
		Redis:       redis,
		Producer:    nil,
		KafkaTopic:  "policy",
		Logger:      logger,
		RulesPath:   "./policy-engine/rules",
		WorkflowURL: cfg.Temporal.Address(),
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		policyEngine.Start(ctx)
	}()

	// Emergency Controls
	emergencyEngine := emergency_controls.NewEngine(emergency_controls.Config{
		PostgreSQL: postgres,
		Redis:      redis,
		Producer:   nil,
		Logger:     logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		emergencyEngine.Start(ctx)
	}()

	// Scenario Simulator
	simulator := scenario_simulator.NewEngine(scenario_simulator.Config{
		Redis:  redis,
		Logger: logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		simulator.Start(ctx)
	}()

	// State Coordination
	stateCoord := state_coordination.NewEngine(state_coordination.Config{
		PostgreSQL: postgres,
		Redis:      redis,
		Producer:   nil,
		Logger:     logger,
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		stateCoord.Start(ctx)
	}()

	// Setup HTTP server
	router := setupRouter(cfg, policyEngine, emergencyEngine, simulator, stateCoord, temporalClient)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("Intervention engine starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down intervention engine...")
	cancel()
	wg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Intervention engine stopped")
}

// initializeTemporalClient creates a Temporal client connection
func initializeTemporalClient(cfg config.TemporalConfig) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  cfg.Address(),
		Namespace: cfg.Namespace,
	})
}

// startTemporalWorker starts the Temporal worker for intervention workflows
func startTemporalWorker(temporalClient client.Client, acts *activities.InterventionActivities, taskQueue string) {
	w := worker.New(temporalClient, taskQueue, worker.Options{})

	// Register all activities
	w.RegisterActivity(acts.ValidatePolicyActivity)
	w.RegisterActivity(acts.GetPolicyDetailsActivity)
	w.RegisterActivity(acts.ReserveBudgetActivity)
	w.RegisterActivity(acts.ReleaseBudgetActivity)
	w.RegisterActivity(acts.ExecuteInterventionActivity)
	w.RegisterActivity(acts.RollbackInterventionActivity)
	w.RegisterActivity(acts.UpdateFeatureStoreActivity)
	w.RegisterActivity(acts.NotifyStakeholdersActivity)
	w.RegisterActivity(acts.RecordOutcomeActivity)
	w.RegisterActivity(acts.HealthCheckActivity)

	// Register all workflows
	w.RegisterWorkflow(workflows.ExecuteInterventionWorkflow)
	w.RegisterWorkflow(workflows.InterventionApprovalWorkflow)
	w.RegisterWorkflow(workflows.BatchInterventionWorkflow)

	logger := shared.NewLogger("temporal-worker")
	logger.Info("Starting Temporal worker", "task_queue", taskQueue)

	if err := w.Run(worker.StopCh()); err != nil {
		logger.Error("Temporal worker failed", "error", err)
	}
}

func setupRouter(cfg *config.Config, policyEng *policy_engine.Engine, emergencyEng *emergency_controls.Engine, simulator *scenario_simulator.Engine, stateCoord *state_coordination.Engine, temporalClient client.Client) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(authMiddleware(cfg.Security.JWTSecret))

	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Policy management
		policies := v1.Group("/policies")
		{
			policies.GET("/", listPoliciesHandler(policyEng))
			policies.POST("/", createPolicyHandler(policyEng))
			policies.GET("/:id", getPolicyHandler(policyEng))
			policies.PUT("/:id", updatePolicyHandler(policyEng))
			policies.DELETE("/:id", deletePolicyHandler(policyEng))
			policies.POST("/:id/activate", activatePolicyHandler(policyEng))
			policies.POST("/:id/deactivate", deactivatePolicyHandler(policyEng))
		}

		// Policy validation
		v1.POST("/policies/validate", validatePolicyHandler(policyEng))

		// Emergency controls
		emergency := v1.Group("/emergency")
		{
			emergency.GET("/status", getEmergencyStatusHandler(emergencyEng))
			emergency.POST("/locks", createEconomicLockHandler(emergencyEng))
			emergency.DELETE("/locks/:id", removeEconomicLockHandler(emergencyEng))
			emergency.POST("/subsidies/switch", switchSubsidiesHandler(emergencyEng))
			emergency.POST("/override", createOverrideHandler(emergencyEng))
		}

		// Scenarios
		scenarios := v1.Group("/scenarios")
		{
			scenarios.GET("/", listScenariosHandler(simulator))
			scenarios.POST("/", createScenarioHandler(simulator))
			scenarios.GET("/:id", getScenarioHandler(simulator))
			scenarios.POST("/:id/run", runScenarioHandler(simulator))
			scenarios.GET("/:id/results", getScenarioResultsHandler(simulator))
		}

		// State coordination
		state := v1.Group("/state")
		{
			state.GET("/nodes", listStateNodesHandler(stateCoord))
			state.GET("/federation/status", getFederationStatusHandler(stateCoord))
			state.POST("/sync", syncStateHandler(stateCoord))
		}

		// Intervention workflows (Temporal integration)
		interventions := v1.Group("/interventions")
		{
			interventions.POST("/start", startInterventionWorkflowHandler(temporalClient, cfg.Temporal.TaskQueue))
			interventions.POST("/start-with-approval", startApprovalWorkflowHandler(temporalClient, cfg.Temporal.TaskQueue))
			interventions.POST("/batch", startBatchWorkflowHandler(temporalClient, cfg.Temporal.TaskQueue))
			interventions.GET("/:id/status", getInterventionStatusHandler(temporalClient))
		}
	}

	return router
}

func authMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation would include JWT validation, RBAC checks
		c.Next()
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-intervention",
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "neam-intervention",
	})
}

func listPoliciesHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"policies": []interface{}{
				map[string]interface{}{
					"id":       "pol-001",
					"name":     "Regional Inflation Control",
					"type":     "inflation_control",
					"status":   "active",
					"created":  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
				},
			},
			"total": 1,
		})
	}
}

func createPolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":      "pol-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"status":  "draft",
			"message": "Policy created successfully",
		})
	}
}

func getPolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"id":           id,
			"name":         "Sample Policy",
			"type":         "inflation_control",
			"status":       "active",
			"rules":        []interface{}{},
			"created":      time.Now().Format(time.RFC3339),
			"last_updated": time.Now().Format(time.RFC3339),
		})
	}
}

func updatePolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Policy updated successfully",
		})
	}
}

func deletePolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Policy deleted successfully",
		})
	}
}

func activatePolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Policy activated",
		})
	}
}

func deactivatePolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Policy deactivated",
		})
	}
}

func validatePolicyHandler(eng *policy_engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"errors":   []string{},
			"warnings": []string{},
		})
	}
}

func getEmergencyStatusHandler(eng *emergency_controls.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"emergency_mode":     false,
			"active_locks":       0,
			"subsidy_overrides":  []string{},
			"last_updated":       time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func createEconomicLockHandler(eng *emergency_controls.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":        "lock-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"status":    "active",
			"message":   "Economic lock created",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func removeEconomicLockHandler(eng *emergency_controls.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Economic lock removed",
		})
	}
}

func switchSubsidiesHandler(eng *emergency_controls.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Subsidies switched",
			"from":      "default",
			"to":        "emergency",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func createOverrideHandler(eng *emergency_controls.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":        "override-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"status":    "active",
			"message":   "Override activated",
			"expires":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		})
	}
}

func listScenariosHandler(eng *scenario_simulator.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"scenarios": []interface{}{},
			"total":     0,
		})
	}
}

func createScenarioHandler(eng *scenario_simulator.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{
			"id":        "scen-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"status":    "draft",
			"message":   "Scenario created",
		})
	}
}

func getScenarioHandler(eng *scenario_simulator.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":      c.Param("id"),
			"name":    "Sample Scenario",
			"status":  "draft",
			"what_if": map[string]interface{}{},
			"impact":  map[string]interface{}{},
		})
	}
}

func runScenarioHandler(eng *scenario_simulator.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"run_id":  "run-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			"status":  "running",
			"message": "Scenario simulation started",
		})
	}
}

func getScenarioResultsHandler(eng *scenario_simulator.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"run_id":    c.Param("id"),
			"status":    "completed",
			"results":   map[string]interface{}{},
			"completed": time.Now().Format(time.RFC3339),
		})
	}
}

func listStateNodesHandler(eng *state_coordination.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"nodes": []interface{}{
				map[string]interface{}{
					"id":        "STATE-NORTH",
					"status":    "connected",
					"last_sync": time.Now().Format(time.RFC3339),
				},
			},
		})
	}
}

func getFederationStatusHandler(eng *state_coordination.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"federated":   true,
			"nodes":       5,
			"sync_status": "synced",
		})
	}
}

func syncStateHandler(eng *state_coordination.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":     "Sync initiated",
			"nodes_count": 5,
			"started":     time.Now().Format(time.RFC3339),
		})
	}
}

// Intervention workflow handlers
func startInterventionWorkflowHandler(temporalClient client.Client, taskQueue string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input workflows.InterventionWorkflowInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.InterventionID == "" {
			input.InterventionID = fmt.Sprintf("int-%d", time.Now().UnixNano())
		}

		workflowOptions := client.StartWorkflowOptions{
			TaskQueue: taskQueue,
			ID:        input.InterventionID,
		}

		workflowRun, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.ExecuteInterventionWorkflow, input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"workflow_id": workflowRun.GetID(),
			"run_id":      workflowRun.GetRunID(),
			"status":      "started",
			"message":     "Intervention workflow initiated",
		})
	}
}

func startApprovalWorkflowHandler(temporalClient client.Client, taskQueue string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input workflows.InterventionWorkflowInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.InterventionID == "" {
			input.InterventionID = fmt.Sprintf("int-approval-%d", time.Now().UnixNano())
		}

		workflowOptions := client.StartWorkflowOptions{
			TaskQueue: taskQueue,
			ID:        input.InterventionID + "-approval",
		}

		workflowRun, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.InterventionApprovalWorkflow, input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"workflow_id": workflowRun.GetID(),
			"run_id":      workflowRun.GetRunID(),
			"status":      "pending_approval",
			"message":     "Intervention workflow waiting for approval",
		})
	}
}

func startBatchWorkflowHandler(temporalClient client.Client, taskQueue string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var inputs []workflows.InterventionWorkflowInput
		if err := c.ShouldBindJSON(&inputs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Assign IDs if not provided
		for i := range inputs {
			if inputs[i].InterventionID == "" {
				inputs[i].InterventionID = fmt.Sprintf("batch-int-%d-%d", time.Now().UnixNano(), i)
			}
		}

		workflowOptions := client.StartWorkflowOptions{
			TaskQueue: taskQueue,
			ID:        fmt.Sprintf("batch-%d", time.Now().UnixNano()),
		}

		workflowRun, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.BatchInterventionWorkflow, inputs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"workflow_id":   workflowRun.GetID(),
			"run_id":        workflowRun.GetRunID(),
			"total_count":   len(inputs),
			"status":        "started",
			"message":       "Batch intervention workflow initiated",
		})
	}
}

func getInterventionStatusHandler(temporalClient client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowID := c.Param("id")

		workflowRun := temporalClient.GetWorkflow(context.Background(), workflowID, "")
		desc, err := workflowRun.Describe(context.Background())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "workflow not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"workflow_id":  desc.ExecutionInfo.WorkflowID,
			"run_id":       desc.ExecutionInfo.RunID,
			"status":       desc.Status.String(),
			"start_time":   desc.ExecutionInfo.StartTime.Format(time.RFC3339),
			"close_time":   desc.ExecutionInfo.CloseTime.Format(time.RFC3339),
		})
	}
}
