package service

import (
	"context"
	"fmt"
	"time"

	"github.com/csic-platform/services/security/access-control/internal/core/domain"
	"github.com/csic-platform/services/security/access-control/internal/core/ports"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AccessControlServiceImpl implements the AccessControlService interface
type AccessControlServiceImpl struct {
	policyRepo     ports.PolicyRepository
	ownershipRepo  ports.ResourceOwnershipRepository
	auditRepo      ports.AuditLogRepository
	logger         *zap.Logger
}

// NewAccessControlService creates a new AccessControlServiceImpl
func NewAccessControlService(
	policyRepo ports.PolicyRepository,
	ownershipRepo ports.ResourceOwnershipRepository,
	auditRepo ports.AuditLogRepository,
	logger *zap.Logger,
) *AccessControlServiceImpl {
	return &AccessControlServiceImpl{
		policyRepo:    policyRepo,
		ownershipRepo: ownershipRepo,
		auditRepo:     auditRepo,
		logger:        logger,
	}
}

// CheckAccess performs access control check
func (s *AccessControlServiceImpl) CheckAccess(
	ctx context.Context,
	req *domain.AccessRequest,
) (*domain.AccessResponse, error) {
	s.logger.Debug("Checking access",
		zap.String("subject", req.Subject.ID),
		zap.String("resource", req.Resource.Type),
		zap.String("action", req.Action.Name))

	// Get applicable policies
	policies, err := s.policyRepo.FindApplicable(ctx, req)
	if err != nil {
		s.logger.Error("Failed to get applicable policies", zap.Error(err))
		return &domain.AccessResponse{
			Decision:  domain.AccessDecisionIndeterminate,
			Reason:    "Failed to evaluate policies",
			Timestamp: time.Now(),
		}, nil
	}

	// Sort policies by priority (higher priority first)
	sortPolicies(policies)

	// Evaluate policies
	var matchingPolicy *domain.Policy
	for _, policy := range policies {
		if s.evaluatePolicy(policy, req) {
			matchingPolicy = policy
			break
		}
	}

	// Determine decision
	response := &domain.AccessResponse{
		Timestamp: time.Now(),
	}

	if matchingPolicy == nil {
		// No matching policy - default deny
		response.Decision = domain.AccessDecisionNotApplicable
		response.Reason = "No applicable policy found"
	} else {
		response.PolicyID = &matchingPolicy.ID
		if matchingPolicy.Effect == domain.PolicyEffectAllow {
			response.Decision = domain.AccessDecisionAllow
			response.Reason = fmt.Sprintf("Allowed by policy: %s", matchingPolicy.Name)
		} else {
			response.Decision = domain.AccessDecisionDeny
			response.Reason = fmt.Sprintf("Denied by policy: %s", matchingPolicy.Name)
		}
	}

	// Log the access check
	auditLog := &domain.AuditLog{
		ID:         uuid.New(),
		Request:    *req,
		Response:   *response,
		Decision:   response.Decision,
		UserID:     req.Subject.ID,
		ResourceID: req.Resource.ID,
		Action:     req.Action.Name,
		IPAddress:  req.Context.IPAddress,
		UserAgent:  req.Context.UserAgent,
		Timestamp:  time.Now(),
	}
	s.auditRepo.Create(ctx, auditLog)

	s.logger.Info("Access check completed",
		zap.String("decision", string(response.Decision)),
		zap.String("subject", req.Subject.ID),
		zap.String("resource", req.Resource.Type),
		zap.String("action", req.Action.Name))

	return response, nil
}

// evaluatePolicy evaluates if a policy matches the request
func (s *AccessControlServiceImpl) evaluatePolicy(
	policy *domain.Policy,
	req *domain.AccessRequest,
) bool {
	// Evaluate subject conditions
	if !s.evaluateSubjectConditions(policy.Conditions.Subjects, req.Subject) {
		return false
	}

	// Evaluate resource conditions
	if !s.evaluateResourceConditions(policy.Conditions.Resources, req.Resource) {
		return false
	}

	// Evaluate action conditions
	if !s.evaluateActionConditions(policy.Conditions.Actions, req.Action) {
		return false
	}

	// Evaluate environment conditions
	if !s.evaluateEnvironmentConditions(policy.Conditions.Environment, req.Context) {
		return false
	}

	return true
}

// evaluateSubjectConditions evaluates subject conditions
func (s *AccessControlServiceImpl) evaluateSubjectConditions(
	conditions []domain.SubjectCondition,
	subject domain.Subject,
) bool {
	if len(conditions) == 0 {
		return true // No subject conditions means match all
	}

	for _, cond := range conditions {
		// Check roles
		if len(cond.Roles) > 0 {
			hasRole := false
			for _, role := range subject.Roles {
				for _, allowedRole := range cond.Roles {
					if role == allowedRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			if !hasRole {
				continue
			}
		}

		// Check users
		if len(cond.Users) > 0 {
			found := false
			for _, user := range cond.Users {
				if user == subject.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// All conditions matched
		return true
	}

	return false
}

// evaluateResourceConditions evaluates resource conditions
func (s *AccessControlServiceImpl) evaluateResourceConditions(
	conditions []domain.ResourceCondition,
	resource domain.Resource,
) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, cond := range conditions {
		// Check resource types
		if len(cond.Types) > 0 {
			found := false
			for _, rt := range cond.Types {
				if rt == resource.Type {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check resource IDs
		if len(cond.IDs) > 0 {
			found := false
			for _, id := range cond.IDs {
				if id == resource.ID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		return true
	}

	return false
}

// evaluateActionConditions evaluates action conditions
func (s *AccessControlServiceImpl) evaluateActionConditions(
	conditions []domain.ActionCondition,
	action domain.Action,
) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, cond := range conditions {
		// Check operations
		if len(cond.Operations) > 0 {
			found := false
			for _, op := range cond.Operations {
				if op == action.Name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check HTTP methods
		if len(cond.Methods) > 0 {
			found := false
			for _, method := range cond.Methods {
				if method == action.Method {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		return true
	}

	return false
}

// evaluateEnvironmentConditions evaluates environment conditions
func (s *AccessControlServiceImpl) evaluateEnvironmentConditions(
	conditions []domain.EnvironmentCondition,
	context domain.Context,
) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, cond := range conditions {
		// Check IP ranges (simplified - would need proper IP range checking)
		if len(cond.IPRanges) > 0 {
			ipMatch := false
			for _, ipRange := range cond.IPRanges {
				if context.IPAddress == ipRange {
					ipMatch = true
					break
				}
			}
			if !ipMatch {
				continue
			}
		}

		// Check time ranges (simplified)
		if len(cond.TimeRanges) > 0 {
			currentTime := context.Time
			currentWeekday := int(currentTime.Weekday())
			for _, tr := range cond.TimeRanges {
				for _, day := range tr.DaysOfWeek {
					if day == currentWeekday {
						return true
					}
				}
			}
			continue
		}

		return true
	}

	return false
}

// sortPolicies sorts policies by priority (higher priority first)
func sortPolicies(policies []*domain.Policy) {
	// Simple bubble sort for small policy lists
	for i := 0; i < len(policies); i++ {
		for j := i + 1; j < len(policies); j++ {
			if policies[j].Priority > policies[i].Priority {
				policies[i], policies[j] = policies[j], policies[i]
			}
		}
	}
}

// IsOwner checks if a subject is the owner of a resource
func (s *AccessControlServiceImpl) IsOwner(
	ctx context.Context,
	resourceID string,
	subjectID string,
) (bool, error) {
	ownership, err := s.ownershipRepo.FindByResourceID(ctx, resourceID)
	if err != nil {
		return false, err
	}

	if ownership == nil {
		return false, nil
	}

	return ownership.OwnerID == subjectID, nil
}

// CanAccessResource checks if a subject can access a resource with a specific action
func (s *AccessControlServiceImpl) CanAccessResource(
	ctx context.Context,
	subjectID string,
	resourceID string,
	actionName string,
) (bool, error) {
	// Check ownership first
	isOwner, err := s.IsOwner(ctx, resourceID, subjectID)
	if err != nil {
		return false, err
	}

	if isOwner {
		return true, nil
	}

	// Build access request
	req := &domain.AccessRequest{
		Subject: domain.Subject{
			ID:    subjectID,
			Type:  "user",
			Roles: []string{},
		},
		Resource: domain.Resource{
			ID:   resourceID,
			Type: "resource",
		},
		Action: domain.Action{
			Name: actionName,
		},
		Context: domain.Context{
			Time: time.Now(),
		},
	}

	response, err := s.CheckAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Decision == domain.AccessDecisionAllow, nil
}

// Ensure AccessControlServiceImpl implements AccessControlService
var _ ports.AccessControlService = (*AccessControlServiceImpl)(nil)
