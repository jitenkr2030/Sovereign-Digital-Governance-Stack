package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/pkg/metrics"
)

// PolicyEngine evaluates and manages policies
type PolicyEngine interface {
	// Policy CRUD operations
	ListPolicies(ctx context.Context) ([]*domain.Policy, error)
	GetPolicy(ctx context.Context, id string) (*domain.Policy, error)
	CreatePolicy(ctx context.Context, req *domain.CreatePolicyRequest) (*domain.Policy, error)
	UpdatePolicy(ctx context.Context, id string, req *domain.UpdatePolicyRequest) (*domain.Policy, error)
	DeletePolicy(ctx context.Context, id string) error

	// Policy evaluation
	EvaluatePolicy(ctx context.Context, policyID string, data map[string]interface{}) (*domain.PolicyResult, error)
	EvaluateAllPolicies(ctx context.Context, data map[string]interface{}) ([]*domain.PolicyResult, error)

	// Lifecycle
	StartPolicyUpdateConsumer(logger *zap.Logger)
	IsReady() bool
}

// PolicyEngineService implements the PolicyEngine interface
type PolicyEngineService struct {
	repositories  ports.Repositories
	cachePort     ports.CachePort
	messagingPort ports.MessagingPort
	logger        *zap.Logger
	metrics       *metrics.MetricsCollector

	// Internal state
	mu           sync.RWMutex
	policyCache  map[string]*domain.Policy
	cacheExpiry  time.Time
	ready        bool
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(
	repositories ports.Repositories,
	cachePort ports.CachePort,
	messagingPort ports.MessagingPort,
	logger *zap.Logger,
	metricsCollector *metrics.MetricsCollector,
) PolicyEngine {
	return &PolicyEngineService{
		repositories:  repositories,
		cachePort:     cachePort,
		messagingPort: messagingPort,
		logger:        logger,
		metrics:       metricsCollector,
		policyCache:   make(map[string]*domain.Policy),
		cacheExpiry:   time.Now(),
	}
}

// ListPolicies lists all policies
func (e *PolicyEngineService) ListPolicies(ctx context.Context) ([]*domain.Policy, error) {
	return e.repositories.PolicyRepository.GetAllPolicies(ctx)
}

// GetPolicy gets a policy by ID
func (e *PolicyEngineService) GetPolicy(ctx context.Context, id string) (*domain.Policy, error) {
	policyUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	return e.repositories.PolicyRepository.GetPolicyByID(ctx, policyUUID)
}

// CreatePolicy creates a new policy
func (e *PolicyEngineService) CreatePolicy(ctx context.Context, req *domain.CreatePolicyRequest) (*domain.Policy, error) {
	now := time.Now()
	policy := &domain.Policy{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Rule:        *req.Rule,
		Priority:    req.Priority,
		IsActive:    req.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := e.repositories.PolicyRepository.CreatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// Invalidate cache
	e.invalidateCache()

	// Publish policy update event
	e.messagingPort.Producer.PublishPolicyUpdate(policy.ID.String())

	e.logger.Info("Created policy",
		logger.String("policy_id", policy.ID.String()),
		logger.String("name", policy.Name),
	)

	return policy, nil
}

// UpdatePolicy updates an existing policy
func (e *PolicyEngineService) UpdatePolicy(ctx context.Context, id string, req *domain.UpdatePolicyRequest) (*domain.Policy, error) {
	policyUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID: %w", err)
	}

	// Get existing policy
	policy, err := e.repositories.PolicyRepository.GetPolicyByID(ctx, policyUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if policy == nil {
		return nil, fmt.Errorf("policy not found: %s", id)
	}

	// Update fields
	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.Rule != nil {
		policy.Rule = *req.Rule
	}
	if req.Priority != nil {
		policy.Priority = *req.Priority
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	policy.UpdatedAt = time.Now()

	if err := e.repositories.PolicyRepository.UpdatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// Invalidate cache
	e.invalidateCache()

	// Publish policy update event
	e.messagingPort.Producer.PublishPolicyUpdate(policy.ID.String())

	e.logger.Info("Updated policy",
		logger.String("policy_id", policy.ID.String()),
		logger.String("name", policy.Name),
	)

	return policy, nil
}

// DeletePolicy deletes a policy
func (e *PolicyEngineService) DeletePolicy(ctx context.Context, id string) error {
	policyUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid policy ID: %w", err)
	}

	if err := e.repositories.PolicyRepository.DeletePolicy(ctx, policyUUID); err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	// Invalidate cache
	e.invalidateCache()

	e.logger.Info("Deleted policy", logger.String("policy_id", id))

	return nil
}

// EvaluatePolicy evaluates a single policy against the provided data
func (e *PolicyEngineService) EvaluatePolicy(ctx context.Context, policyID string, data map[string]interface{}) (*domain.PolicyResult, error) {
	start := time.Now()

	policy, err := e.getCachedPolicy(ctx, policyID)
	if err != nil {
		e.metrics.RecordPolicyEvaluation(policyID, "error", float64(time.Since(start).Milliseconds()))
		return nil, err
	}

	if policy == nil {
		e.metrics.RecordPolicyEvaluation(policyID, "not_found", float64(time.Since(start).Milliseconds()))
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}

	if !policy.IsActive {
		e.metrics.RecordPolicyEvaluation(policyID, "skipped", float64(time.Since(start).Milliseconds()))
		return &domain.PolicyResult{
			PolicyID:  policy.ID.String(),
			Compliant: true,
			Details:   "Policy is inactive",
		}, nil
	}

	// Evaluate the policy rule
	result := e.evaluateRule(&policy.Rule, data)

	e.metrics.RecordPolicyEvaluation(policyID, result.Status, float64(time.Since(start).Milliseconds()))

	e.logger.Debug("Evaluated policy",
		logger.String("policy_id", policy.ID.String()),
		logger.String("result", result.Status),
	)

	return result, nil
}

// EvaluateAllPolicies evaluates all active policies against the provided data
func (e *PolicyEngineService) EvaluateAllPolicies(ctx context.Context, data map[string]interface{}) ([]*domain.PolicyResult, error) {
	policies, err := e.repositories.PolicyRepository.GetActivePolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active policies: %w", err)
	}

	var results []*domain.PolicyResult
	for _, policy := range policies {
		result := e.evaluateRule(&policy.Rule, data)
		result.PolicyID = policy.ID.String()

		results = append(results, result)
	}

	return results, nil
}

// evaluateRule evaluates a policy rule against the provided data
func (e *PolicyEngineService) evaluateRule(rule *domain.PolicyRule, data map[string]interface{}) *domain.PolicyResult {
	result := &domain.PolicyResult{
		PolicyID:  "",
		Compliant: true,
		Details:   make(map[string]interface{}),
		Status:    "compliant",
	}

	// Get the target value from data
	targetValue, exists := e.getTargetValue(rule.Target, data)
	if !exists {
		result.Compliant = true
		result.Details["message"] = "Target not found in data"
		result.Status = "target_not_found"
		return result
	}

	// Compare values based on operator
	compliant := e.compareValues(targetValue, rule.Threshold, rule.Operator)

	result.Compliant = compliant
	result.Details["target"] = rule.Target
	result.Details["actual_value"] = targetValue
	result.Details["threshold"] = rule.Threshold
	result.Details["operator"] = rule.Operator

	if compliant {
		result.Details["message"] = "Policy condition met"
		result.Status = "compliant"
	} else {
		result.Details["message"] = fmt.Sprintf("Policy violation: %s %s %v", rule.Target, rule.Operator, rule.Threshold)
		result.Status = "violation"
	}

	return result
}

// getTargetValue extracts the target value from data
func (e *PolicyEngineService) getTargetValue(target string, data map[string]interface{}) (interface{}, bool) {
	// Handle nested keys with dot notation
	keys := splitKey(target)
	value := data

	for _, key := range keys {
		if val, ok := value[key]; ok {
			if nested, ok := val.(map[string]interface{}); ok {
				value = nested
			} else {
				return val, true
			}
		} else {
			return nil, false
		}
	}

	return nil, false
}

// compareValues compares a value against a threshold using the operator
func (e *PolicyEngineService) compareValues(value interface{}, threshold interface{}, operator string) bool {
	switch operator {
	case "gt", ">":
		return e.greaterThan(value, threshold)
	case "gte", ">=":
		return e.greaterThanOrEqual(value, threshold)
	case "lt", "<":
		return e.lessThan(value, threshold)
	case "lte", "<=":
		return e.lessThanOrEqual(value, threshold)
	case "eq", "==":
		return e.equal(value, threshold)
	case "neq", "!=":
		return !e.equal(value, threshold)
	case "contains":
		return e.contains(value, threshold)
	case "in":
		return e.inList(value, threshold)
	default:
		return true
	}
}

func (e *PolicyEngineService) greaterThan(value, threshold interface{}) bool {
	switch v := value.(type) {
	case float64:
		t, ok := threshold.(float64)
		if !ok {
			t = parseFloat(threshold)
		}
		return v > t
	case int:
		t := parseFloat(threshold)
		return float64(v) > t
	default:
		return false
	}
}

func (e *PolicyEngineService) greaterThanOrEqual(value, threshold interface{}) bool {
	return e.greaterThan(value, threshold) || e.equal(value, threshold)
}

func (e *PolicyEngineService) lessThan(value, threshold interface{}) bool {
	switch v := value.(type) {
	case float64:
		t, ok := threshold.(float64)
		if !ok {
			t = parseFloat(threshold)
		}
		return v < t
	case int:
		t := parseFloat(threshold)
		return float64(v) < t
	default:
		return false
	}
}

func (e *PolicyEngineService) lessThanOrEqual(value, threshold interface{}) bool {
	return e.lessThan(value, threshold) || e.equal(value, threshold)
}

func (e *PolicyEngineService) equal(value, threshold interface{}) bool {
	// Convert both to float64 for comparison if possible
	v := parseFloat(value)
	t := parseFloat(threshold)
	return v == t
}

func (e *PolicyEngineService) contains(value, threshold interface{}) bool {
	v := fmt.Sprintf("%v", value)
	t := fmt.Sprintf("%v", threshold)
	return len(v) > 0 && len(t) > 0 && (containsString(v, t) || containsString(t, v))
}

func (e *PolicyEngineService) inList(value, threshold interface{}) bool {
	// Threshold should be a list
	list, ok := threshold.([]interface{})
	if !ok {
		return false
	}

	v := fmt.Sprintf("%v", value)
	for _, item := range list {
		if fmt.Sprintf("%v", item) == v {
			return true
		}
	}
	return false
}

func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitKey(key string) []string {
	// Simple split by dot
	var keys []string
	var current string
	for _, c := range key {
		if c == '.' {
			if current != "" {
				keys = append(keys, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		keys = append(keys, current)
	}
	return keys
}

// getCachedPolicy gets a policy from cache or fetches it
func (e *PolicyEngineService) getCachedPolicy(ctx context.Context, policyID string) (*domain.Policy, error) {
	e.mu.RLock()
	if time.Now().Before(e.cacheExpiry) {
		if policy, ok := e.policyCache[policyID]; ok {
			e.mu.RUnlock()
			e.metrics.UpdateStateCacheHit()
			return policy, nil
		}
	}
	e.mu.RUnlock()

	e.metrics.UpdateStateCacheMiss()

	// Fetch from database
	policyUUID, err := uuid.Parse(policyID)
	if err != nil {
		return nil, err
	}

	policy, err := e.repositories.PolicyRepository.GetPolicyByID(ctx, policyUUID)
	if err != nil {
		return nil, err
	}

	if policy != nil {
		// Update cache
		e.mu.Lock()
		e.policyCache[policyID] = policy
		e.cacheExpiry = time.Now().Add(5 * time.Minute)
		e.mu.Unlock()
	}

	return policy, nil
}

// invalidateCache clears the policy cache
func (e *PolicyEngineService) invalidateCache() {
	e.mu.Lock()
	e.policyCache = make(map[string]*domain.Policy)
	e.cacheExpiry = time.Time{}
	e.mu.Unlock()
}

// StartPolicyUpdateConsumer starts consuming policy update events
func (e *PolicyEngineService) StartPolicyUpdateConsumer(logger *zap.Logger) {
	// This would consume from Kafka in production
	e.mu.Lock()
	e.ready = true
	e.mu.Unlock()

	logger.Info("Policy update consumer started")
}

// IsReady returns true if the service is ready
func (e *PolicyEngineService) IsReady() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ready
}
