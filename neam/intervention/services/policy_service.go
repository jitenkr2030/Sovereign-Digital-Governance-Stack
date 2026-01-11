package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"intervention-service/models"
	"intervention-service/repository"
)

// PolicyService handles policy business logic
type PolicyService struct {
	repo *repository.PostgresRepository
}

// NewPolicyService creates a new policy service
func NewPolicyService(repo *repository.PostgresRepository) *PolicyService {
	return &PolicyService{
		repo: repo,
	}
}

// CreatePolicy creates a new policy
func (s *PolicyService) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	policy.ID = uuid.New()
	policy.Status = models.PolicyStatusDraft
	policy.Version = "1.0.0"
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	if err := s.validatePolicy(policy); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.repo.CreatePolicy(ctx, policy)
}

// GetPolicy retrieves a policy by ID
func (s *PolicyService) GetPolicy(ctx context.Context, id uuid.UUID) (*models.Policy, error) {
	return s.repo.GetPolicy(ctx, id)
}

// ListPolicies retrieves policies with filters
func (s *PolicyService) ListPolicies(ctx context.Context, filters map[string]interface{}) ([]*models.Policy, error) {
	return s.repo.ListPolicies(ctx, filters)
}

// UpdatePolicy updates an existing policy
func (s *PolicyService) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	policy.UpdatedAt = time.Now()
	policy.Version = incrementVersion(policy.Version)

	return s.repo.UpdatePolicy(ctx, policy)
}

// DeletePolicy deletes a policy (soft delete by archiving)
func (s *PolicyService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	policy.Status = models.PolicyStatusArchived
	policy.UpdatedAt = time.Now()

	return s.repo.UpdatePolicy(ctx, policy)
}

// ActivatePolicy activates a policy
func (s *PolicyService) ActivatePolicy(ctx context.Context, id uuid.UUID) error {
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	if policy.Status == models.PolicyStatusActive {
		return fmt.Errorf("policy is already active")
	}

	now := time.Now()
	policy.Status = models.PolicyStatusActive
	policy.ActivatedAt = &now
	policy.UpdatedAt = now

	return s.repo.UpdatePolicy(ctx, policy)
}

// DeactivatePolicy deactivates a policy
func (s *PolicyService) DeactivatePolicy(ctx context.Context, id uuid.UUID) error {
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	policy.Status = models.PolicyStatusSuspended
	policy.UpdatedAt = time.Now()
	policy.ActivatedAt = nil

	return s.repo.UpdatePolicy(ctx, policy)
}

// ValidatePolicy validates a policy's rules and conditions
func (s *PolicyService) ValidatePolicy(ctx context.Context, policy *models.Policy) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"valid":    true,
		"errors":   []string{},
		"warnings": []string{},
	}

	// Validate policy structure
	if err := s.validatePolicy(policy); err != nil {
		result["valid"] = false
		result["errors"] = append(result["errors"].([]string), err.Error())
	}

	// Validate OPA conditions if present
	if policy.Conditions != "" {
		// OPA validation would go here
		result["warnings"] = append(result["warnings"].([]string), "OPA policy validation not implemented")
	}

	return result, nil
}

func (s *PolicyService) validatePolicy(policy *models.Policy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}
	if policy.Type == "" {
		return fmt.Errorf("policy type is required")
	}
	if policy.EnforcementLevel == "" {
		return fmt.Errorf("enforcement level is required")
	}
	return nil
}

func incrementVersion(version string) string {
	// Simple version increment - in production, use proper semver handling
	return version
}
